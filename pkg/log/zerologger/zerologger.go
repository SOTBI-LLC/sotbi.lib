package zerologger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type Logger struct {
	*zerolog.Logger
}

// New -.
func New(level string) Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		l = zerolog.ErrorLevel
	case "warn":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	zl := zerolog.
		New(os.Stdout).
		Level(l).
		With().
		Timestamp().
		CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).
		Logger()

	return Logger{
		&zl,
	}
}

func (l Logger) GetLevel() zerolog.Level {
	return l.Logger.GetLevel()
}

// Debug -.
func (l Logger) Debug(message string, args ...interface{}) {
	if l.Logger.GetLevel() <= zerolog.DebugLevel {
		l.msg("debug", message, args...)
	}
}

// Info -.
func (l Logger) Info(message string, args ...interface{}) {
	if l.Logger.GetLevel() <= zerolog.InfoLevel {
		l.log(message, args...)
	}
}

// Warn -.
func (l Logger) Warn(message string, args ...interface{}) {
	if l.Logger.GetLevel() <= zerolog.WarnLevel {
		l.log(message, args...)
	}
}

// Error -.
func (l Logger) Error(message string, args ...interface{}) {
	if l.Logger.GetLevel() == zerolog.DebugLevel {
		l.Debug(message, args...)
	}

	l.msg("error", message, args...)
}

// Fatal -.
func (l Logger) Fatal(message string, args ...interface{}) {
	l.msg("fatal", message, args...)

	os.Exit(1)
}

func (l Logger) log(message string, args ...interface{}) {
	if len(args) == 0 {
		l.Logger.Info().Msg(message)
	} else {
		l.Logger.Info().Msgf(message, args...)
	}
}

func (l Logger) msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(msg.Error(), args...)
	case string:
		l.log(msg, args...)
	default:
		l.log(fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}
