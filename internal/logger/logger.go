package logger

import (
	"log/slog"
	"os"
)

func Setup(json bool) {
	var handler slog.Handler
	if json {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func SetJSONMode(enabled bool) {
	var handler slog.Handler
	if enabled {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	slog.SetDefault(slog.New(handler))
}
