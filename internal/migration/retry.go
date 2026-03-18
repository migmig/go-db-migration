package migration

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// RetryConfig는 재시도 정책을 정의한다.
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	Multiplier  float64
	MaxWait     time.Duration
}

// DefaultRetryConfig는 기본 재시도 설정을 반환한다.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: time.Second,
		Multiplier:  2.0,
		MaxWait:     30 * time.Second,
	}
}

// WithRetry는 fn을 recoverable 에러에 한해 지수 백오프로 재시도한다.
func WithRetry(
	ctx context.Context,
	cfg RetryConfig,
	tableName string,
	eventFn func(RetryEvent),
	fn func() error,
) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.InitialWait <= 0 {
		cfg.InitialWait = time.Second
	}
	if cfg.Multiplier < 1 {
		cfg.Multiplier = 1
	}
	if cfg.MaxWait <= 0 {
		cfg.MaxWait = cfg.InitialWait
	}

	wait := cfg.InitialWait
	for attempt := 1; ; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		var migErr *MigrationError
		if !errors.As(err, &migErr) || !migErr.Recoverable {
			return err
		}
		if attempt >= cfg.MaxAttempts {
			return err
		}

		if eventFn != nil {
			eventFn(RetryEvent{
				TableName:   tableName,
				Attempt:     attempt,
				MaxAttempts: cfg.MaxAttempts,
				ErrorMsg:    err.Error(),
				WaitSeconds: int(wait.Seconds()),
			})
		}

		slog.Warn("migration retry",
			"table", tableName,
			"attempt", attempt,
			"wait_s", wait.Seconds(),
			"error", err,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}

		nextWait := time.Duration(float64(wait) * cfg.Multiplier)
		if nextWait > cfg.MaxWait {
			wait = cfg.MaxWait
		} else {
			wait = nextWait
		}
	}
}
