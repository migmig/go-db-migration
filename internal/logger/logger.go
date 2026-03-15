package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

const redactedValue = "[REDACTED]"

var sensitiveLogKeys = map[string]struct{}{
	"password":       {},
	"pass":           {},
	"dbpass":         {},
	"db_pass":        {},
	"password_hash":  {},
	"master_key":     {},
	"dbm_master_key": {},
	"token":          {},
	"secret":         {},
	"api_key":        {},
	"apikey":         {},
	"authorization":  {},
	"password_enc":   {},
}

func Setup(json bool) {
	handler := newHandler(os.Stdout, json)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func SetJSONMode(enabled bool) {
	handler := newHandler(os.Stdout, enabled)
	slog.SetDefault(slog.New(handler))
}

func newHandler(w io.Writer, json bool) slog.Handler {
	opts := &slog.HandlerOptions{ReplaceAttr: redactSensitiveAttr}
	if json {
		return slog.NewJSONHandler(w, opts)
	}
	return slog.NewTextHandler(w, opts)
}

func redactSensitiveAttr(_ []string, attr slog.Attr) slog.Attr {
	key := strings.ToLower(attr.Key)
	if _, ok := sensitiveLogKeys[key]; ok {
		attr.Value = slog.StringValue(redactedValue)
	}
	return attr
}
