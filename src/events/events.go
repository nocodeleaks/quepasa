package events

import (
	"sync"
	"sync/atomic"
	"time"
)

// Event represents an internal operational fact emitted by the system.
// Producers should publish stable domain-oriented names and statuses.
type Event struct {
	Name       string
	Source     string
	Status     string
	Timestamp  time.Time
	Duration   time.Duration
	Attributes map[string]string
}

// Handler consumes one event asynchronously.
type Handler func(Event)

// Filter optionally decides whether a subscriber should receive an event.
type Filter func(Event) bool

// SubscribeOptions configures one asynchronous subscriber.
type SubscribeOptions struct {
	Name       string
	BufferSize int
	Filter     Filter
	Handler    Handler
}

type subscriber struct {
	filter  Filter
	handler Handler
	events  chan Event
	done    chan struct{}
}

// Bus is an in-process non-blocking event bus.
type Bus struct {
	mu          sync.RWMutex
	nextID      uint64
	subscribers map[uint64]*subscriber
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{subscribers: make(map[uint64]*subscriber)}
}

var defaultBus = NewBus()

// Publish sends the event to all matching subscribers without blocking.
func Publish(event Event) {
	defaultBus.Publish(event)
}

// Subscribe registers a new subscriber on the default bus.
func Subscribe(options SubscribeOptions) func() {
	return defaultBus.Subscribe(options)
}

// Publish sends the event to all matching subscribers without blocking.
func (bus *Bus) Publish(event Event) {
	if bus == nil {
		return
	}

	normalized := normalizeEvent(event)

	bus.mu.RLock()
	snapshot := make([]*subscriber, 0, len(bus.subscribers))
	for _, current := range bus.subscribers {
		snapshot = append(snapshot, current)
	}
	bus.mu.RUnlock()

	for _, current := range snapshot {
		if current == nil {
			continue
		}
		if current.filter != nil && !current.filter(normalized) {
			continue
		}
		current.enqueue(normalized)
	}
}

// Subscribe registers a new subscriber and returns an unsubscribe callback.
func (bus *Bus) Subscribe(options SubscribeOptions) func() {
	if bus == nil || options.Handler == nil {
		return func() {}
	}

	bufferSize := options.BufferSize
	if bufferSize <= 0 {
		bufferSize = 64
	}

	identifier := atomic.AddUint64(&bus.nextID, 1)
	current := &subscriber{
		filter:  options.Filter,
		handler: options.Handler,
		events:  make(chan Event, bufferSize),
		done:    make(chan struct{}),
	}

	bus.mu.Lock()
	bus.subscribers[identifier] = current
	bus.mu.Unlock()

	go current.run()

	return func() {
		bus.mu.Lock()
		delete(bus.subscribers, identifier)
		bus.mu.Unlock()

		select {
		case <-current.done:
		default:
			close(current.done)
		}
	}
}

func normalizeEvent(event Event) Event {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Attributes != nil {
		attributes := make(map[string]string, len(event.Attributes))
		for key, value := range event.Attributes {
			attributes[key] = value
		}
		event.Attributes = attributes
	}
	return event
}

func (source *subscriber) enqueue(event Event) {
	if source == nil {
		return
	}

	select {
	case <-source.done:
		return
	case source.events <- event:
	default:
		return
	}
}

func (source *subscriber) run() {
	if source == nil {
		return
	}

	for {
		select {
		case <-source.done:
			return
		case event := <-source.events:
			source.handler(event)
		}
	}
}
