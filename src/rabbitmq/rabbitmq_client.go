package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient encapsulates the RabbitMQ connection and channel with reconnection logic and a message cache.
type RabbitMQClient struct {
	connURI string
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex // Protects conn and channel

	publishMu sync.Mutex // Serializes all channel.Publish calls (amqp.Channel is not goroutine-safe)

	notify chan *amqp.Error // Channel for AMQP channel close notifications
	closed chan struct{}    // Signals that the client should stop
	wg     sync.WaitGroup   // To wait for goroutines to finish

	messageCache    chan RabbitMQMessage // In-memory queue for messages when disconnected
	cacheProcessing sync.Once            // Ensures cache processor starts only once
	maxCacheSize    int                  // Maximum number of messages to hold in cache

	// Flag to track if QuePasa Exchange and Queues have been set up
	quepasaSetupDone bool
	setupMutex       sync.Mutex // Protects quepasaSetupDone
}

// NewRabbitMQClient creates and initializes a new RabbitMQClient instance.
// maxCacheSize defines the maximum number of messages that can be cached in memory.
// If maxCacheSize is 0, the cache capacity is set to a large, but manageable, default.
func NewRabbitMQClient(connURI string, maxCacheSize uint64) *RabbitMQClient {
	actualCacheSize := int(maxCacheSize)

	const DefaultUnlimitedCacheCapacity = 100000

	if maxCacheSize == 0 {
		actualCacheSize = DefaultUnlimitedCacheCapacity
		log.Printf("maxCacheSize is 0, setting cache capacity to effectively unlimited (default: %d).", actualCacheSize)
	}

	client := &RabbitMQClient{
		connURI:      connURI,
		closed:       make(chan struct{}),
		messageCache: make(chan RabbitMQMessage, actualCacheSize),
		maxCacheSize: actualCacheSize,
	}
	client.wg.Add(1)
	go client.monitorConnection()

	return client
}

const (
	dialTimeout    = 10 * time.Second // TCP connection timeout per attempt
	amqpHeartbeat  = 10 * time.Second // AMQP heartbeat — detects dead connections within ~20s
	initialBackoff = 5 * time.Second
	maxBackoff     = 60 * time.Second
)

// connect establishes a new connection and channel with RabbitMQ.
// Uses an explicit dial timeout and AMQP heartbeat for fast failure detection.
// Slow I/O (Dial, Channel) happens outside the mutex; fields are set under lock.
func (r *RabbitMQClient) connect() error {
	conn, err := amqp.DialConfig(r.connURI, amqp.Config{
		Heartbeat: amqpHeartbeat,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, dialTimeout)
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Prepare notification channel before exposing the new channel to other goroutines.
	notify := make(chan *amqp.Error, 1)
	ch.NotifyClose(notify)

	// Hold lock only while updating shared fields.
	r.mu.Lock()
	r.conn = conn
	r.channel = ch
	r.mu.Unlock()

	// notify is only read by monitorConnection (same goroutine), so no lock needed.
	r.notify = notify

	// Reset setup flag for the new connection.
	r.setupMutex.Lock()
	r.quepasaSetupDone = false
	r.setupMutex.Unlock()

	log.Println("RabbitMQ connection and channel established successfully.")

	// Start cache processing only once across the client lifetime.
	r.cacheProcessing.Do(func() {
		r.wg.Add(1)
		go r.processCache()
	})

	ConnectionsEstablished.Inc()
	return nil
}

// monitorConnection monitors the connection state and attempts to reconnect on failure.
// Uses exponential backoff (5s → 10s → 20s … cap 60s) to avoid overwhelming the broker.
// Backoff resets to the initial value after every successful connection.
func (r *RabbitMQClient) monitorConnection() {
	defer r.wg.Done()

	backoff := initialBackoff

	for {
		// Attempt connection if no active channel.
		r.mu.RLock()
		channelAvailable := r.channel != nil
		r.mu.RUnlock()

		if !channelAvailable {
			log.Printf("Attempting to connect to RabbitMQ (backoff: %v)...", backoff)
			err := r.connect()
			if err != nil {
				log.Printf("Error connecting: %v. Retrying in %v...", err, backoff)
				ReconnectionAttempts.Inc()
				ReconnectionBackoff.Set(backoff.Seconds())
				select {
				case <-time.After(backoff):
				case <-r.closed:
					return
				}
				// Double the backoff for next failure, capped at maxBackoff.
				if backoff *= 2; backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}
			// Successful connection — reset backoff.
			backoff = initialBackoff
			ReconnectionBackoff.Set(0)
		}

		// Wait for channel close notification or shutdown signal.
		select {
		case err := <-r.notify:
			if err != nil {
				log.Printf("RabbitMQ connection closed unexpectedly: %v. Attempting to reconnect...", err)
			} else {
				log.Println("RabbitMQ connection closed (clean disconnect). Attempting to reconnect...")
			}
			// Count every disconnect and the upcoming reconnect attempt.
			ConnectionsLost.Inc()

			r.mu.Lock()
			r.channel = nil
			r.conn = nil
			r.mu.Unlock()
			continue

		case <-r.closed:
			log.Println("RabbitMQ connection monitor shutting down.")
			return
		}
	}
}

// GetChannel returns the active channel for publishing. It blocks until a channel is available.
// For internal use where a timeout is not required. External callers should prefer
// IsConnectionReady() + a retry loop with a deadline.
func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	for {
		r.mu.RLock()
		ch := r.channel
		r.mu.RUnlock()

		if ch != nil {
			return ch
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// AddToCache adds a message to the in-memory cache if the cache is not full.
// Returns true if added, false if the cache is full.
func (r *RabbitMQClient) AddToCache(msg RabbitMQMessage) bool {
	select {
	case r.messageCache <- msg:
		size := len(r.messageCache)
		payloadStr := fmt.Sprintf("%v", msg.Payload)
		if len(payloadStr) > 50 {
			payloadStr = payloadStr[:47] + "..."
		}
		log.Printf("Message cached: ID=%s, Payload=%s (Cache size: %d/%d)", msg.ID, payloadStr, size, r.maxCacheSize)
		CacheSizeCurrent.Set(float64(size))
		return true
	default:
		log.Printf("Cache is full, dropping message: ID=%s. Max cache size: %d", msg.ID, r.maxCacheSize)
		MessagesCacheDropped.Inc()
		return false
	}
}

// processCache drains the in-memory cache by publishing messages whenever a connection is available.
func (r *RabbitMQClient) processCache() {
	defer r.wg.Done()
	log.Println("RabbitMQ cache processor started.")

	for {
		select {
		case <-r.closed:
			log.Println("RabbitMQ cache processor shutting down.")
			return

		case <-time.After(500 * time.Millisecond):
			r.mu.RLock()
			ch := r.channel
			r.mu.RUnlock()

			if ch == nil {
				continue
			}

			// Ensure exchange and queues exist before draining the cache.
			if err := r.EnsureExchangeAndQueues(); err != nil {
				log.Printf("processCache: exchange/queue setup not ready: %v", err)
				continue
			}

			for {
				select {
				case msg := <-r.messageCache:
					log.Printf("Attempting to publish cached message: %s", msg.ID)

					body, err := json.Marshal(msg)
					if err != nil {
						log.Printf("Error marshaling cached message ID %s: %v. Discarding.", msg.ID, err)
						continue
					}

					// Re-read channel under lock; it may have changed since the outer check.
					r.mu.RLock()
					currentChannel := r.channel
					r.mu.RUnlock()

					if currentChannel == nil {
						log.Printf("Channel became unavailable while processing cache. Putting message %s back.", msg.ID)
						select {
						case r.messageCache <- msg:
						default:
							log.Printf("Cache is full (failed to put back), dropping cached message: %s.", msg.ID)
						}
						goto CACHE_LOOP_END
					}

					r.publishMu.Lock()
					err = currentChannel.Publish(
						msg.Exchange,
						msg.RoutingKey,
						false,
						false,
						amqp.Publishing{
							ContentType:  "application/json",
							Body:         body,
							DeliveryMode: amqp.Persistent,
						})
					r.publishMu.Unlock()

					if err != nil {
						log.Printf("Failed to publish cached message ID %s to exchange '%s' routing key '%s': %v. Putting back.", msg.ID, msg.Exchange, msg.RoutingKey, err)
						select {
						case r.messageCache <- msg:
						default:
							log.Printf("Cache is full (failed to put back), dropping cached message: %s.", msg.ID)
						}
						goto CACHE_LOOP_END
					}

					size := len(r.messageCache)
					log.Printf("Cached message ID %s published successfully (exchange '%s', routing key '%s'). Cache size: %d/%d", msg.ID, msg.Exchange, msg.RoutingKey, size, r.maxCacheSize)
					MessagesCacheProcessed.Inc()
					CacheSizeCurrent.Set(float64(size))

				default:
					goto CACHE_LOOP_END
				}
			}
		CACHE_LOOP_END:
		}
	}
}

// Close closes the RabbitMQ connection and channel and stops all goroutines.
func (r *RabbitMQClient) Close() {
	log.Println("Closing RabbitMQ client...")
	close(r.closed)
	r.wg.Wait()

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

// PublishMessageToExchange publishes a JSON message to a RabbitMQ exchange with a routing key.
// Returns true if the message was published directly, false if it was added to the cache.
// The channel mutex ensures concurrent callers do not corrupt the AMQP protocol framing.
func (r *RabbitMQClient) PublishMessageToExchange(exchangeName, routingKey string, messageContent any) bool {
	msg := RabbitMQMessage{
		ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Payload:    messageContent,
		Timestamp:  time.Now(),
		Exchange:   exchangeName,
		RoutingKey: routingKey,
	}

	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		log.Printf("Connection is down. Caching message ID %s for exchange '%s' routing key '%s'", msg.ID, exchangeName, routingKey)
		if r.AddToCache(msg) {
			MessagesCached.Inc()
		}
		return false
	}

	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message ID %s (payload type %T): %v. Not caching (invalid format).", msg.ID, msg.Payload, err)
		return false
	}

	log.Printf("Publishing message ID %s to exchange '%s' routing key '%s'", msg.ID, exchangeName, routingKey)

	r.publishMu.Lock()
	err = ch.Publish(
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	r.publishMu.Unlock()

	if err != nil {
		log.Printf("Failed to publish message ID %s to exchange '%s' routing key '%s': %v. Caching.", msg.ID, exchangeName, routingKey, err)
		MessagePublishErrors.Inc()
		if r.AddToCache(msg) {
			MessagesCached.Inc()
		}
		return false
	}

	log.Printf("Message ID %s published successfully to exchange '%s' routing key '%s'.", msg.ID, exchangeName, routingKey)
	MessagesPublished.Inc()
	return true
}

// EnsureExchangeAndQueues ensures the QuePasa standard exchange and queues exist.
// Runs at most once per connection.
func (r *RabbitMQClient) EnsureExchangeAndQueues() error {
	r.setupMutex.Lock()
	if r.quepasaSetupDone {
		r.setupMutex.Unlock()
		return nil
	}
	defer r.setupMutex.Unlock()

	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("rabbitmq channel not available")
	}

	err := ch.ExchangeDeclare(
		QuePasaExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange '%s': %v", QuePasaExchangeName, err)
	}
	log.Printf("Exchange '%s' declared successfully", QuePasaExchangeName)

	queues := map[string]string{
		QuePasaQueueProd:    QuePasaRoutingKeyProd,
		QuePasaQueueHistory: QuePasaRoutingKeyHistory,
		QuePasaQueueEvents:  QuePasaRoutingKeyEvents,
	}

	for queueName, routingKey := range queues {
		q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("failed to declare queue '%s': %v", queueName, err)
		}
		log.Printf("Queue '%s' declared (consumers: %d, messages: %d)", q.Name, q.Consumers, q.Messages)

		err = ch.QueueBind(queueName, routingKey, QuePasaExchangeName, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind queue '%s' to exchange '%s' routing key '%s': %v", queueName, QuePasaExchangeName, routingKey, err)
		}
		log.Printf("Queue '%s' bound to exchange '%s' routing key '%s'", queueName, QuePasaExchangeName, routingKey)
	}

	r.quepasaSetupDone = true
	log.Printf("QuePasa Exchange and Queues setup completed successfully")
	return nil
}

// EnsureExchangeAndQueuesWithRetry tries to ensure Exchange and Queues with a short retry.
func (r *RabbitMQClient) EnsureExchangeAndQueuesWithRetry() error {
	if err := r.EnsureExchangeAndQueues(); err == nil {
		return nil
	}
	if r.WaitForConnection(5 * time.Second) {
		return r.EnsureExchangeAndQueues()
	}
	return fmt.Errorf("connection not ready after timeout")
}

// PublishQuePasaMessage publishes a message to the standard QuePasa Exchange.
// Returns true if published directly, false if cached.
func (r *RabbitMQClient) PublishQuePasaMessage(routingKey string, messageContent any) bool {
	return r.PublishMessageToExchange(QuePasaExchangeName, routingKey, messageContent)
}

// IsConnectionReady checks if the RabbitMQ connection and channel are ready.
func (r *RabbitMQClient) IsConnectionReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel != nil
}

// WaitForConnection waits for the RabbitMQ connection to be ready with a timeout.
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
func handleError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}
