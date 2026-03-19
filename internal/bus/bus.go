package bus

import (
	"sync"
)

// EventType represents the type of an event
type EventType string

const (
	EventInit             EventType = "init"
	EventUpdate           EventType = "update"
	EventDone             EventType = "done"
	EventError            EventType = "error"
	EventRetry            EventType = "retry"
	EventAllDone          EventType = "all_done"
	EventDryRunResult     EventType = "dry_run_result"
	EventDDLProgress      EventType = "ddl_progress"
	EventWarning          EventType = "warning"
	EventValidationStart  EventType = "validation_start"
	EventValidationResult EventType = "validation_result"
	EventDiscoverySummary EventType = "discovery_summary"
	EventMetrics          EventType = "metrics" // v11 Phase 1
)

// Event payload
type Event struct {
	Type          EventType
	Table         string
	Count         int
	Total         int
	Error         error
	Message       string
	ZipFileID     string
	ConnectionOk  bool
	Object        string
	ObjectName    string
	Status        string
	ObjectGroup   string
	Tables        []string
	Sequences     []string
	ReportSummary interface{} // using interface{} to decouple
	Attempt       int
	MaxAttempts   int
	WaitSeconds   int
}

// Handler is the callback function when an event is published
type Handler func(Event)

// EventBus provides publish/subscribe capabilities
type EventBus interface {
	Subscribe(eventType EventType, handler Handler)
	Publish(event Event)
}

type memoryEventBus struct {
	handlers map[EventType][]Handler
	mu       sync.RWMutex
}

// NewEventBus creates a new in-memory EventBus
func NewEventBus() EventBus {
	return &memoryEventBus{
		handlers: make(map[EventType][]Handler),
	}
}

func (b *memoryEventBus) Subscribe(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *memoryEventBus) Publish(event Event) {
	b.mu.RLock()
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()

	for _, handler := range handlers {
		// Call handler synchronously to preserve event ordering.
		// If asynchronous processing is needed, it should be handled
		// by the subscriber, e.g. using channels, to ensure ordering.
		handler(event)
	}
}
