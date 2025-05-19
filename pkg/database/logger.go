package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type SlogLogger struct {
	logLevel logger.LogLevel
	slog     *slog.Logger
}

func New(level string) *SlogLogger {
	var l slog.Level
	var lvl logger.LogLevel

	switch strings.ToLower(level) {
	case "error":
		l = slog.LevelError
		lvl = logger.Error
	case "warn":
		l = slog.LevelWarn
		lvl = logger.Warn
	case "info":
		l = slog.LevelInfo
		lvl = logger.Info
	case "debug":
		l = slog.LevelDebug
		lvl = logger.Info
	default:
		l = slog.LevelInfo
		lvl = logger.Info
	}

	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     l,
	}

	return &SlogLogger{
		logLevel: lvl,
		slog:     slog.New(slog.NewJSONHandler(os.Stdout, opts)),
	}
}

func (l *SlogLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

func (l *SlogLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= logger.Info {
		l.slog.InfoContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= logger.Warn {
		l.slog.WarnContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *SlogLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.logLevel >= logger.Error {
		l.slog.ErrorContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *SlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := []slog.Attr{
		slog.String("elapsed", elapsed.String()),
		slog.Int64("rows", rows),
		slog.String("sql", sql),
		slog.String("file", utils.FileWithLineNum()),
	}

	switch {
	case err != nil && l.logLevel >= logger.Error:
		attrs = append(attrs, slog.String("error", err.Error()))
		l.slog.LogAttrs(ctx, slog.LevelError, "query error", attrs...)
	case elapsed > 200*time.Millisecond && l.logLevel >= logger.Warn:
		l.slog.LogAttrs(ctx, slog.LevelWarn, "slow query", attrs...)
	case l.logLevel >= logger.Info:
		l.slog.LogAttrs(ctx, slog.LevelInfo, "query", attrs...)
	}
}
