package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRedactSensitiveAttr_RedactsKnownKeys(t *testing.T) {
	attr := redactSensitiveAttr(nil, slog.String("password", "plain"))
	if attr.Value.String() != redactedValue {
		t.Fatalf("expected redacted value, got %q", attr.Value.String())
	}

	attr = redactSensitiveAttr(nil, slog.String("DBM_MASTER_KEY", "secret-key"))
	if attr.Value.String() != redactedValue {
		t.Fatalf("expected redacted value for master key, got %q", attr.Value.String())
	}
}

func TestRedactSensitiveAttr_LeavesNonSensitiveKey(t *testing.T) {
	attr := redactSensitiveAttr(nil, slog.String("table", "USERS"))
	if attr.Value.String() != "USERS" {
		t.Fatalf("expected original value, got %q", attr.Value.String())
	}
}

func TestNewHandler_JSONRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(newHandler(&buf, true))
	logger.Info("login", "user", "alice", "password", "plain", "token", "abc")

	out := buf.String()
	if strings.Contains(out, "plain") || strings.Contains(out, "abc") {
		t.Fatalf("sensitive values should be redacted: %s", out)
	}
	if !strings.Contains(out, redactedValue) {
		t.Fatalf("expected redacted marker in output: %s", out)
	}
}

func TestNewHandler_TextRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(newHandler(&buf, false))
	logger.Info("credential saved", "password_enc", "ciphertext")

	out := buf.String()
	if strings.Contains(out, "ciphertext") {
		t.Fatalf("encrypted secret should be redacted: %s", out)
	}
	if !strings.Contains(out, redactedValue) {
		t.Fatalf("expected redacted marker in output: %s", out)
	}
}

func TestSetJSONMode_TogglesWithoutPanic(t *testing.T) {
	for i := 0; i < 5; i++ {
		SetJSONMode(i%2 == 0)
	}
	SetJSONMode(false)
}

func TestSetup_DoesNotPanic(t *testing.T) {
	Setup(true)
	Setup(false)
}
