package migration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Fatalf("MaxAttempts=%d, want 3", cfg.MaxAttempts)
	}
	if cfg.InitialWait != time.Second {
		t.Fatalf("InitialWait=%s, want 1s", cfg.InitialWait)
	}
	if cfg.Multiplier != 2.0 {
		t.Fatalf("Multiplier=%f, want 2.0", cfg.Multiplier)
	}
	if cfg.MaxWait != 30*time.Second {
		t.Fatalf("MaxWait=%s, want 30s", cfg.MaxWait)
	}
}

func TestWithRetry_SucceedsAfterOneRetry(t *testing.T) {
	attempts := 0
	events := make([]RetryEvent, 0, 1)

	cfg := RetryConfig{MaxAttempts: 3, InitialWait: time.Millisecond, Multiplier: 2, MaxWait: 5 * time.Millisecond}
	err := WithRetry(context.Background(), cfg, "USERS", func(ev RetryEvent) {
		events = append(events, ev)
	}, func() error {
		attempts++
		if attempts == 1 {
			return &MigrationError{Table: "USERS", Phase: "data", Category: ErrTimeout, Recoverable: true, RootCause: errors.New("timeout")}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("WithRetry() error = %v, want nil", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts=%d, want 2", attempts)
	}
	if len(events) != 1 {
		t.Fatalf("events=%d, want 1", len(events))
	}
	if events[0].Attempt != 1 || events[0].MaxAttempts != 3 || events[0].TableName != "USERS" {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestWithRetry_MaxAttemptsExhausted(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{MaxAttempts: 3, InitialWait: time.Millisecond, Multiplier: 2, MaxWait: 2 * time.Millisecond}

	err := WithRetry(context.Background(), cfg, "ORDERS", nil, func() error {
		attempts++
		return &MigrationError{Table: "ORDERS", Phase: "data", Category: ErrConnectionLost, Recoverable: true, RootCause: errors.New("connection reset")}
	})

	if err == nil {
		t.Fatalf("WithRetry() error=nil, want error")
	}
	if attempts != 3 {
		t.Fatalf("attempts=%d, want 3", attempts)
	}
}

func TestWithRetry_NonRecoverableReturnsImmediately(t *testing.T) {
	attempts := 0
	cfg := RetryConfig{MaxAttempts: 5, InitialWait: time.Millisecond, Multiplier: 2, MaxWait: 3 * time.Millisecond}

	err := WithRetry(context.Background(), cfg, "PAYMENTS", nil, func() error {
		attempts++
		return &MigrationError{Table: "PAYMENTS", Phase: "data", Category: ErrUniqueViolation, Recoverable: false, RootCause: errors.New("duplicate key")}
	})

	if err == nil {
		t.Fatalf("WithRetry() error=nil, want error")
	}
	if attempts != 1 {
		t.Fatalf("attempts=%d, want 1", attempts)
	}
}

func TestWithRetry_ContextCancelledDuringBackoff(t *testing.T) {
	attempts := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := RetryConfig{MaxAttempts: 5, InitialWait: 30 * time.Millisecond, Multiplier: 2, MaxWait: 50 * time.Millisecond}

	err := WithRetry(ctx, cfg, "ITEMS", nil, func() error {
		attempts++
		if attempts == 1 {
			go func() {
				time.Sleep(5 * time.Millisecond)
				cancel()
			}()
		}
		return &MigrationError{Table: "ITEMS", Phase: "data", Category: ErrTimeout, Recoverable: true, RootCause: fmt.Errorf("deadline exceeded")}
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithRetry() error=%v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts=%d, want 1", attempts)
	}
}
