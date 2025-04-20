package slog

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New(level string) *Logger {
	var l slog.Level

	switch strings.ToLower(level) {
	case "error":
		l = slog.LevelError
	case "warn":
		l = slog.LevelWarn
	case "info":
		l = slog.LevelInfo
	case "debug":
		l = slog.LevelDebug
	default:
		l = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     l,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	return &Logger{
		logger,
	}
}

func (l Logger) Fatal(message string, args ...any) {
	l.Log(context.Background(), slog.LevelError, message, args...)
	os.Exit(1)
}

func (l Logger) Printf(message string, args ...any) {
	l.Log(context.Background(), slog.LevelError, message, args...)
}
