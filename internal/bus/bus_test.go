package bus

import (
	"sync"
	"testing"
)

func TestEventBusPublishDeliversSubscribedHandlersInOrder(t *testing.T) {
	b := NewEventBus()

	var (
		mu       sync.Mutex
		received []string
	)

	b.Subscribe(EventUpdate, func(e Event) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, "first:"+e.Table)
	})
	b.Subscribe(EventUpdate, func(e Event) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, "second:"+e.Table)
	})

	b.Publish(Event{Type: EventUpdate, Table: "users"})

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
	if received[0] != "first:users" {
		t.Fatalf("unexpected first handler value: %s", received[0])
	}
	if received[1] != "second:users" {
		t.Fatalf("unexpected second handler value: %s", received[1])
	}
}

func TestEventBusPublishIgnoresUnsubscribedEventType(t *testing.T) {
	b := NewEventBus()
	called := false

	b.Subscribe(EventDone, func(e Event) {
		called = true
	})

	b.Publish(Event{Type: EventError})

	if called {
		t.Fatal("handler for EventDone should not be called for EventError")
	}
}
