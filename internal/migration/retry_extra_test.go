package migration

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithRetry_Defaults(t *testing.T) {
	// Tests default config values
	err := WithRetry(context.Background(), RetryConfig{}, "T", nil, func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Error("expected error")
	}
}

func TestWithRetry_ContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Done immediately
	
	err := WithRetry(ctx, RetryConfig{MaxAttempts: 2}, "T", nil, func() error {
		return &MigrationError{Recoverable: true}
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context canceled, got %v", err)
	}
}

func TestWithRetry_MaxWait(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts: 3,
		InitialWait: time.Millisecond,
		Multiplier:  10.0,
		MaxWait:     2 * time.Millisecond,
	}
	
	count := 0
	_ = WithRetry(context.Background(), cfg, "T", nil, func() error {
		count++
		return &MigrationError{Recoverable: true}
	})
	if count != 3 {
		t.Errorf("expected 3 attempts, got %d", count)
	}
}
