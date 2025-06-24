package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

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

	messageCache    chan RabbitMQMessage // In-memory queue for messages when disconnected
	cacheProcessing sync.Once            // Ensures cache processor starts only once
	maxCacheSize    int                  // Maximum number of messages to hold in cache
}

// NewRabbitMQClient creates and initializes a new RabbitMQClient instance.
// maxCacheSize defines the maximum number of messages that can be cached in memory.
// If maxCacheSize is 0, the cache capacity is set to a large, but manageable, default.
func NewRabbitMQClient(connURI string, maxCacheSize uint64) *RabbitMQClient {
	actualCacheSize := int(maxCacheSize) // Converte uint64 para int

	// Define uma capacidade padrão para o "cache ilimitado"
	const DefaultUnlimitedCacheCapacity = 100000 // Exemplo: 100.000 mensagens

	// Se maxCacheSize for 0 (indicando cache "ilimitado"), usa a capacidade padrão.
	if maxCacheSize == 0 {
		actualCacheSize = DefaultUnlimitedCacheCapacity
		log.Printf("maxCacheSize is 0, setting cache capacity to effectively unlimited (default: %d).", actualCacheSize)
	}
	// A condição 'else if actualCacheSize <= 0' foi removida porque uint64 já garante valores não-negativos.
	// Se maxCacheSize (uint64) for um número muito grande que excede math.MaxInt, Go causará um panic
	// na conversão para int, mas isso é um caso extremo que não está ligado ao zero ou negativo.

	client := &RabbitMQClient{
		connURI:      connURI,
		mu:           sync.RWMutex{},
		closed:       make(chan struct{}),
		messageCache: make(chan RabbitMQMessage, actualCacheSize), // Usa o tamanho determinado
		maxCacheSize: actualCacheSize,                             // Armazena o tamanho real configurado
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

	log.Println("RabbitMQ connection and channel established successfully.")

	// Start cache processing only once after a successful connection
	r.cacheProcessing.Do(func() {
		r.wg.Add(1) // Add one goroutine for cache processing
		go r.processCache()
	})

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

// AddToCache adds a message to the in-memory cache if the cache is not full.
// Returns true if added to cache, false if cache is full.
func (r *RabbitMQClient) AddToCache(msg RabbitMQMessage) bool {
	select {
	case r.messageCache <- msg:
		// Convert payload to string for logging if it's a simple type, otherwise log type
		payloadStr := fmt.Sprintf("%v", msg.Payload)
		if len(payloadStr) > 50 { // Truncate for cleaner logs
			payloadStr = payloadStr[:47] + "..."
		}
		log.Printf("Message cached: ID=%s, Payload=%s (Cache size: %d/%d)", msg.ID, payloadStr, len(r.messageCache), r.maxCacheSize)
		return true
	default:
		// Se o cache for "ilimitado" (capacidade math.MaxInt), esta condição 'default:'
		// só será atingida em situações de extrema escassez de memória, ou se o canal
		// for fechado indevidamente por alguma razão externa (o que não deveria acontecer
		// se a lógica de reconnection estiver correta).
		// Para o propósito de 0 = ilimitado, a ideia é que ele *quase* nunca falhe aqui por "cache cheio".
		log.Printf("Cache is full, dropping message: ID=%s. Max cache size: %d", msg.ID, r.maxCacheSize)
		return false
	}
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
		case <-time.After(500 * time.Millisecond): // Periodically check for messages in cache
			ch := r.channel // Read directly as this goroutine needs to react fast
			if ch == nil {
				continue
			}

			for {
				select {
				case msg := <-r.messageCache: // Get a message from cache
					log.Printf("Attempting to publish cached message: %s", msg.ID)
					body, err := json.Marshal(msg)
					if err != nil {
						log.Printf("Error marshaling cached message ID %s to JSON: %v", msg.ID, err)
						continue
					}

					currentChannel := r.GetChannel()
					if currentChannel == nil {
						log.Printf("Channel became unavailable while processing cache. Putting message %s back to cache.", msg.ID)
						select {
						case r.messageCache <- msg:
							break // Break inner loop to re-evaluate channel
						default:
							log.Printf("Cache is full (failed to put back), dropping cached message: %s.", msg.ID)
						}
						break
					}

					q, err := currentChannel.QueueDeclare(
						msg.TargetQueue,
						true,
						false,
						false,
						false,
						nil,
					)
					if err != nil {
						log.Printf("Failed to declare queue '%s' for cached message %s: %v. Putting message back to cache.", q.Name, msg.ID, err)
						select {
						case r.messageCache <- msg:
						default:
							log.Printf("Cache is full (failed to put back), dropping cached message: %s.", msg.ID)
						}
						break
					}

					err = currentChannel.Publish(
						"",
						q.Name,
						false,
						false,
						amqp.Publishing{
							ContentType: "application/json",
							Body:        body,
						})

					if err != nil {
						log.Printf("Failed to publish cached message ID %s: %v. Putting message back to cache.", msg.ID, err)
						select {
						case r.messageCache <- msg:
						default:
							log.Printf("Cache is full (failed to put back), dropping cached message: %s.", msg.ID)
						}
						break
					}
					log.Printf("Cached message ID %s published successfully. Cache size: %d/%d", msg.ID, len(r.messageCache), r.maxCacheSize)

				default:
					break
				}
				if len(r.messageCache) == 0 {
					break
				}
			}
		}
	}
}

// Close closes the RabbitMQ connection and channel and stops the reconnection monitor and cache processor.
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

// PublishMessageOnQueue publishes a JSON message to a specific RabbitMQ queue using this RabbitMQClient instance.
// It accepts any Go type as messageContent, which will be marshaled into the 'payload' field of RabbitMQMessage.
// If the connection is unavailable, it caches the message. This method provides explicit queue control.
func (r *RabbitMQClient) PublishMessageOnQueue(queueName string, messageContent any) {
	msg := RabbitMQMessage{
		ID:          fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Payload:     messageContent,
		Timestamp:   time.Now(),
		TargetQueue: queueName,
	}

	ch := r.channel
	if ch == nil {
		log.Printf("Connection is down. Attempting to cache message ID %s for queue '%s'", msg.ID, queueName)
		if r.AddToCache(msg) {
			payloadStr := fmt.Sprintf("%v", msg.Payload)
			if len(payloadStr) > 50 {
				payloadStr = payloadStr[:47] + "..."
			}
			log.Printf("Message ID %s with payload '%s' successfully added to cache for queue '%s'.", msg.ID, payloadStr, queueName)
		}
		return
	}

	log.Printf("Connection is active. Attempting to publish message ID %s directly to queue '%s'", msg.ID, queueName)

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	handleError(err, fmt.Sprintf("Failed to declare queue '%s'", queueName))
	if err != nil {
		log.Printf("Queue declaration failed for message ID %s on queue '%s': %v. Attempting to cache.", msg.ID, queueName, err)
		if r.AddToCache(msg) {
			payloadStr := fmt.Sprintf("%v", msg.Payload)
			if len(payloadStr) > 50 {
				payloadStr = payloadStr[:47] + "..."
			}
			log.Printf("Message ID %s with payload '%s' successfully added to cache after queue declaration failure for queue '%s'.", msg.ID, payloadStr, queueName)
		}
		return
	}

	log.Printf("Queue '%s' declared. Consumers: %d, Messages: %d\n", q.Name, q.Consumers, q.Messages)

	body, err := json.Marshal(msg)
	handleError(err, fmt.Sprintf("Failed to convert RabbitMQMessage ID %s to JSON for queue '%s'", msg.ID, queueName))
	if err != nil {
		log.Printf("Failed to marshal message ID %s (payload type %T). Not caching (invalid format or unmarshalable payload). Error: %v", msg.ID, msg.Payload, err)
		return
	}

	log.Printf("JSON message created for queue '%s': %s\n", queueName, string(body))

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Printf("Failed to publish message ID %s directly to queue '%s': %v. Attempting to cache.", msg.ID, queueName, err)
		if r.AddToCache(msg) {
			payloadStr := fmt.Sprintf("%v", msg.Payload)
			if len(payloadStr) > 50 {
				payloadStr = payloadStr[:47] + "..."
			}
			log.Printf("Message ID %s with payload '%s' successfully added to cache after publish failure for queue '%s'.", msg.ID, payloadStr, queueName)
		}
		return
	}
	log.Printf("JSON message ID %s published successfully to queue '%s'!", msg.ID, queueName)
}

// PublishMessage publishes a JSON message to the default RabbitMQ queue (RabbitMQQueueDefault).
// It accepts any Go type as messageContent, which will be marshaled into the 'payload' field of RabbitMQMessage.
// This is a convenience method that wraps PublishMessageOnQueue.
func (r *RabbitMQClient) PublishMessage(messageContent any) { // Alterado para 'any'
	log.Printf("Publishing message to default queue '%s' with payload type %T.", RabbitMQQueueDefault, messageContent)
	r.PublishMessageOnQueue(RabbitMQQueueDefault, messageContent)
}

// handleError is a helper function to log errors.
// It's defined within the rabbitmq package, but not as a method of RabbitMQClient,
// as it's a generic logging helper.
func handleError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}
