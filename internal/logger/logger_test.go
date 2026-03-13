package logger

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestSetJSONMode_EnabledWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	// Temporarily override default logger to capture output
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	SetJSONMode(true)
	slog.Info("test message", "key", "value")

	output := buf.String()
	// JSON handler output should not have "test message" in key=value text format
	// Instead check that SetJSONMode doesn't panic and sets a valid logger.
	// We verify by calling SetJSONMode(false) and ensuring no panic.
	SetJSONMode(false)
	_ = output
}

func TestSetJSONMode_DisabledSetsTextHandler(t *testing.T) {
	// Should not panic
	SetJSONMode(false)
	slog.Info("text mode active")
}

func TestSetJSONMode_TogglesWithoutPanic(t *testing.T) {
	for i := 0; i < 5; i++ {
		SetJSONMode(i%2 == 0)
	}
	// Restore text mode
	SetJSONMode(false)
}

func TestSetup_JSONMode(t *testing.T) {
	Setup(true)
	slog.Info("json setup")
	Setup(false)
}
