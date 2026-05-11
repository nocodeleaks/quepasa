package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	cache "github.com/nocodeleaks/quepasa/cache"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient encapsulates the RabbitMQ connection and channel with reconnection logic and a message cache.
type RabbitMQClient struct {
	connURI string
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex // Mutex to protect access to the channel during reconnection

	notify chan *amqp.Error // Channel for AMQP connection/channel close notifications
	closed chan struct{}    // Signals that the client should stop
	wg     sync.WaitGroup   // To wait for goroutines to finish

	messageCache cache.BytesQueueBackend
	maxCacheSize int

	// Flag to track if QuePasa Exchange and Queues have been set up
	quepasaSetupDone bool
	setupMutex       sync.Mutex // Protects quepasaSetupDone
}

// NewRabbitMQClient creates and initializes a new RabbitMQClient instance.
// The cache backend must be set separately using SetCacheBackend() before the client is used.
// This constructor does not initialize the cache - that is handled by the centralized CacheService.
func NewRabbitMQClient(connURI string, maxCacheSize uint64) *RabbitMQClient {
	actualCacheSize := int(maxCacheSize)

	const DefaultUnlimitedCacheCapacity = 100000

	if maxCacheSize == 0 {
		actualCacheSize = DefaultUnlimitedCacheCapacity
		log.Printf("maxCacheSize is 0, setting cache capacity to effectively unlimited (default: %d).", actualCacheSize)
	}

	client := &RabbitMQClient{
		connURI:      connURI,
		mu:           sync.RWMutex{},
		closed:       make(chan struct{}),
		messageCache: nil, // Backend must be injected via SetCacheBackend()
		maxCacheSize: actualCacheSize,
	}
	client.wg.Add(1)
	go client.monitorConnection()

	return client
}

// connect establishes a new connection and channel with RabbitMQ.
func (r *RabbitMQClient) connect() error {
	var err error

	// Connect to RabbitMQ
	r.conn, err = amqp.Dial(r.connURI)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Open a channel
	r.channel, err = r.conn.Channel()
	if err != nil {
		r.conn.Close() // Close the connection if the channel cannot be opened
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Configure the channel for close notifications
	r.notify = make(chan *amqp.Error)
	r.channel.NotifyClose(r.notify)

	// Reset setup flag since we have a new connection
	r.setupMutex.Lock()
	r.quepasaSetupDone = false
	r.setupMutex.Unlock()

	log.Println("RabbitMQ connection and channel established successfully.")

	ConnectionsEstablished.Inc()

	return nil
}

// monitorConnection monitors the connection state and attempts to reconnect on failure.
func (r *RabbitMQClient) monitorConnection() {
	defer r.wg.Done()

	for {
		// Try to connect if there's no active connection
		r.mu.RLock()
		channelAvailable := r.channel != nil
		r.mu.RUnlock()

		if !channelAvailable {
			log.Println("Attempting to connect to RabbitMQ...")
			err := r.connect()
			if err != nil {
				log.Printf("Error connecting: %v. Retrying in 5 seconds...", err)
				ReconnectionAttempts.Inc()
				select {
				case <-time.After(5 * time.Second):
					continue
				case <-r.closed: // Exit if the client is closed
					return
				}
			}
		}

		select {
		case err := <-r.notify:
			if err != nil {
				log.Printf("RabbitMQ connection closed unexpectedly: %v. Attempting to reconnect...", err)
				ConnectionsLost.Inc()
				ReconnectionAttempts.Inc()
			} else {
				log.Println("RabbitMQ connection closed (clean disconnect). Attempting to reconnect...")
			}
			r.mu.Lock()     // Lock the mutex to prevent using the invalid channel
			r.channel = nil // Mark the channel as invalid
			r.conn = nil    // Mark the connection as invalid
			r.mu.Unlock()   // Release the mutex
			// Loop to attempt immediate reconnection
			continue
		case <-r.closed:
			log.Println("RabbitMQ connection monitor shutting down.")
			return
		}
	}
}

// GetChannel returns the active channel for publishing. It blocks until a channel is available.
func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	for {
		r.mu.RLock() // Read lock
		channel := r.channel
		r.mu.RUnlock()

		if channel != nil {
			return channel
		}
		// If the channel is nil, wait a bit and try again
		time.Sleep(100 * time.Millisecond) // Small delay to avoid consuming CPU in a tight loop
	}
}

// SetCacheBackend sets the cache backend for this RabbitMQ client.
// This is called by the centralized CacheService during application initialization.
// After setting the backend, the client will start processing the cache.
func (r *RabbitMQClient) SetCacheBackend(backend cache.BytesQueueBackend) {
	r.messageCache = backend
	// Start cache processing after backend is set
	r.wg.Add(1)
	go r.processCache()
}

// AddToCache adds a message to the configured retry cache backend.
// Returns true if added to cache, false if cache is full.
func (r *RabbitMQClient) AddToCache(msg RabbitMQMessage) bool {
	payload, err := json.Marshal(msg)
	if r.messageCache == nil {
		return false
	}

	if err != nil {
		log.Printf("Failed to marshal message ID %s for retry cache: %v", msg.ID, err)
		return false
	}

	added, err := r.messageCache.Enqueue(payload)
	if err != nil {
		log.Printf("Failed to enqueue retry cache message ID %s: %v", msg.ID, err)
		return false
	}
	if !added {
		log.Printf("Cache is full, dropping message: ID=%s. Max cache size: %d", msg.ID, r.maxCacheSize)
		return false
	}

	payloadStr := fmt.Sprintf("%v", msg.Payload)
	if len(payloadStr) > 50 {
		payloadStr = payloadStr[:47] + "..."
	}
	cacheLength, err := r.messageCache.Len()
	if err != nil {
		cacheLength = -1
	}
	log.Printf("Message cached: ID=%s, Payload=%s (Cache size: %d/%d)", msg.ID, payloadStr, cacheLength, r.maxCacheSize)
	return true
}

// processCache attempts to publish messages from the cache when a connection is available.
func (r *RabbitMQClient) processCache() {
	defer r.wg.Done()
	log.Println("RabbitMQ cache processor started.")

	for {
		select {
		case <-r.closed:
			log.Println("RabbitMQ cache processor shutting down.")
			return
		case <-time.After(500 * time.Millisecond):
			ch := r.channel
			if ch == nil {
				continue
			}

			for {
				payload, found, err := r.messageCache.Dequeue()
				if err != nil {
					log.Printf("Failed to read retry cache entry: %v", err)
					goto CACHE_LOOP_END
				}
				if !found {
					goto CACHE_LOOP_END
				}

				var msg RabbitMQMessage
				if err := json.Unmarshal(payload, &msg); err != nil {
					log.Printf("Error unmarshaling cached message payload: %v", err)
					continue
				}

				log.Printf("Attempting to publish cached message: %s", msg.ID)
				currentChannel := r.GetChannel()
				if currentChannel == nil {
					log.Printf("Channel became unavailable while processing cache. Putting message %s back to cache.", msg.ID)
					added, enqueueErr := r.messageCache.Enqueue(payload)
					if enqueueErr != nil || !added {
						log.Printf("Cache is full (failed to put back), dropping cached message: %s.", msg.ID)
					}
					break
				}

				err = currentChannel.Publish(
					msg.Exchange,
					msg.RoutingKey,
					false,
					false,
					amqp.Publishing{
						ContentType:  "application/json",
						Body:         payload,
						DeliveryMode: amqp.Persistent,
					})

				if err != nil {
					log.Printf("Failed to publish cached message ID %s to exchange '%s' with routing key '%s': %v. Putting message back to cache.", msg.ID, msg.Exchange, msg.RoutingKey, err)
					added, enqueueErr := r.messageCache.Enqueue(payload)
					if enqueueErr != nil || !added {
						log.Printf("Cache is full (failed to put back), dropping cached message: %s.", msg.ID)
					}
					goto CACHE_LOOP_END
				}

				cacheLength, lenErr := r.messageCache.Len()
				if lenErr != nil {
					cacheLength = -1
				}
				log.Printf("Cached message ID %s published successfully to exchange '%s' with routing key '%s'. Cache size: %d/%d", msg.ID, msg.Exchange, msg.RoutingKey, cacheLength, r.maxCacheSize)
				MessagesCacheProcessed.Inc()

				if cacheLength == 0 {
					goto CACHE_LOOP_END
				}
			}
		CACHE_LOOP_END:
		}
	}
}

// Close closes the RabbitMQ connection and channel and stops the reconnection monitor and cache processor.
func (r *RabbitMQClient) Close() {
	log.Println("Closing RabbitMQ client...")
	close(r.closed)
	r.wg.Wait()

	if r.messageCache != nil {
		if err := r.messageCache.Close(); err != nil {
			log.Printf("Error closing RabbitMQ retry cache backend: %v", err)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		} else {
			log.Println("RabbitMQ channel closed.")
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		} else {
			log.Println("RabbitMQ connection closed.")
		}
	}
	log.Println("RabbitMQ client fully closed.")
}

// PublishMessageToExchange publishes a JSON message to a RabbitMQ exchange with routing key.
// It accepts any Go type as messageContent, which will be marshaled into the 'payload' field of RabbitMQMessage.
// If the connection is unavailable, it caches the message. This method provides exchange-based routing.
func (r *RabbitMQClient) PublishMessageToExchange(exchangeName, routingKey string, messageContent any) {
	msg := RabbitMQMessage{
		ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Payload:    messageContent,
		Timestamp:  time.Now(),
		Exchange:   exchangeName,
		RoutingKey: routingKey,
	}

	ch := r.channel
	if ch == nil {
		log.Printf("Connection is down. Attempting to cache message ID %s for exchange '%s' with routing key '%s'", msg.ID, exchangeName, routingKey)
		if r.AddToCache(msg) {
			payloadStr := fmt.Sprintf("%v", msg.Payload)
			if len(payloadStr) > 50 {
				payloadStr = payloadStr[:47] + "..."
			}
			log.Printf("Message ID %s with payload '%s' successfully added to cache for exchange '%s' with routing key '%s'.", msg.ID, payloadStr, exchangeName, routingKey)
			MessagesCached.Inc()
		}
		return
	}

	log.Printf("Connection is active. Attempting to publish message ID %s to exchange '%s' with routing key '%s'", msg.ID, exchangeName, routingKey)

	// Exchange should already be declared via EnsureExchangeAndQueues()
	// No need to declare it again here

	body, err := json.Marshal(msg)
	handleError(err, fmt.Sprintf("Failed to convert RabbitMQMessage ID %s to JSON for exchange '%s'", msg.ID, exchangeName))
	if err != nil {
		log.Printf("Failed to marshal message ID %s (payload type %T). Not caching (invalid format or unmarshalable payload). Error: %v", msg.ID, msg.Payload, err)
		return
	}

	log.Printf("JSON message created for exchange '%s' with routing key '%s': %s\n", exchangeName, routingKey, string(body))

	err = ch.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		log.Printf("Failed to publish message ID %s to exchange '%s' with routing key '%s': %v. Attempting to cache.", msg.ID, exchangeName, routingKey, err)
		MessagePublishErrors.Inc()
		if r.AddToCache(msg) {
			payloadStr := fmt.Sprintf("%v", msg.Payload)
			if len(payloadStr) > 50 {
				payloadStr = payloadStr[:47] + "..."
			}
			log.Printf("Message ID %s with payload '%s' successfully added to cache after publish failure for exchange '%s' with routing key '%s'.", msg.ID, payloadStr, exchangeName, routingKey)
			MessagesCached.Inc()
		}
		return
	}
	log.Printf("JSON message ID %s published successfully to exchange '%s' with routing key '%s'!", msg.ID, exchangeName, routingKey)
	MessagesPublished.Inc()
}

// EnsureExchangeAndQueues ensures that the QuePasa standard exchange and queues exist
// All bots use the same fixed Exchange and Queue names
// This method only runs once per connection to avoid repeated declarations
func (r *RabbitMQClient) EnsureExchangeAndQueues() error {
	// Check if already set up for this connection
	r.setupMutex.Lock()
	if r.quepasaSetupDone {
		r.setupMutex.Unlock()
		return nil // Already set up, no need to do it again
	}
	defer r.setupMutex.Unlock()

	ch := r.channel
	if ch == nil {
		return fmt.Errorf("rabbitmq channel not available")
	}

	// Declare the QuePasa standard exchange
	err := ch.ExchangeDeclare(
		QuePasaExchangeName,
		"direct", // exchange type - direct for routing keys
		true,     // durable
		false,    // auto-delete
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange '%s': %v", QuePasaExchangeName, err)
	}
	log.Printf("Exchange '%s' declared successfully", QuePasaExchangeName)

	// Define the standard QuePasa queues
	queues := map[string]string{
		QuePasaQueueProd:    QuePasaRoutingKeyProd,    // Production messages
		QuePasaQueueHistory: QuePasaRoutingKeyHistory, // History sync messages
		QuePasaQueueEvents:  QuePasaRoutingKeyEvents,  // Debug, contacts, read receipts, etc.
	}

	// Declare each queue and bind to exchange
	for queueName, routingKey := range queues {
		// Declare queue
		q, err := ch.QueueDeclare(
			queueName,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue '%s': %v", queueName, err)
		}
		log.Printf("Queue '%s' declared successfully. Consumers: %d, Messages: %d", q.Name, q.Consumers, q.Messages)

		// Bind queue to exchange with routing key
		err = ch.QueueBind(
			queueName,           // queue name
			routingKey,          // routing key
			QuePasaExchangeName, // exchange
			false,               // no-wait
			nil,                 // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue '%s' to exchange '%s' with routing key '%s': %v", queueName, QuePasaExchangeName, routingKey, err)
		}
		log.Printf("Queue '%s' bound to exchange '%s' with routing key '%s'", queueName, QuePasaExchangeName, routingKey)
	}

	// Mark as set up
	r.quepasaSetupDone = true
	log.Printf("QuePasa Exchange and Queues setup completed successfully for this connection")

	return nil
}

// EnsureExchangeAndQueuesWithRetry tries to ensure Exchange and Queues with retry logic
func (r *RabbitMQClient) EnsureExchangeAndQueuesWithRetry() error {
	// Try immediate setup first
	err := r.EnsureExchangeAndQueues()
	if err == nil {
		return nil
	}

	// If failed, wait a bit for connection to be ready
	if r.WaitForConnection(5 * time.Second) {
		return r.EnsureExchangeAndQueues()
	}

	return fmt.Errorf("connection not ready after timeout")
}

// PublishQuePasaMessage publishes a message to the standard QuePasa Exchange and routes it to the appropriate queue
// based on the routing key. This method uses the fixed QuePasa Exchange name and routing keys.
// All bots use this method to ensure messages go to the same standard queues.
func (r *RabbitMQClient) PublishQuePasaMessage(routingKey string, messageContent any) {
	// Always use the fixed QuePasa Exchange
	r.PublishMessageToExchange(QuePasaExchangeName, routingKey, messageContent)
}

// IsConnectionReady checks if the RabbitMQ connection and channel are ready
func (r *RabbitMQClient) IsConnectionReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel != nil
}

// WaitForConnection waits for the RabbitMQ connection to be ready with timeout
func (r *RabbitMQClient) WaitForConnection(timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if r.IsConnectionReady() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// handleError is a helper function to log errors.
// It's defined within the rabbitmq package, but not as a method of RabbitMQClient,
// as it's a generic logging helper.
func handleError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}
