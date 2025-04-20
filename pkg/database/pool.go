package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/COTBU/sotbi.lib/pkg/log/slog"
)

type traceLogger struct {
	*slog.Logger
}

func (l *traceLogger) Log(
	_ context.Context,
	level tracelog.LogLevel,
	msg string,
	data map[string]interface{},
) {
	var (
		i      int
		values = make([]any, len(data)*2)
	)

	for k, v := range data {
		values[i] = k
		i++

		values[i] = v
		i++
	}

	switch level {
	case tracelog.LogLevelError:
		l.Error(msg, values...)
	case tracelog.LogLevelWarn:
		l.Warn(msg, values...)
	case tracelog.LogLevelInfo:
		l.Info(msg, values...)
	case tracelog.LogLevelDebug:
		l.Debug(msg, values...)
	}
}

type DataTypes []string

func (d DataTypes) RegisterDataTypes(ctx context.Context, conn *pgx.Conn) error {
	for _, typeName := range d {
		dataType, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return err
		}

		conn.TypeMap().RegisterType(dataType)
	}

	return nil
}

type PoolConfig struct {
	DSN             string
	MaxOpenConns    int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	MaxIdleConns    int32
}

func NewConnectionPool(
	ctx context.Context,
	cfg *PoolConfig,
	sLogger *slog.Logger,
	logLevel string,
	dataTypeNames []string,
) (*pgxpool.Pool, error) {
	conf, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	dtn := DataTypes(dataTypeNames)

	traceLogLevel, err := tracelog.LogLevelFromString(logLevel)
	if err != nil {
		traceLogLevel = tracelog.LogLevelInfo
	}

	conf.MaxConns = cfg.MaxOpenConns
	conf.MaxConnLifetime = cfg.MaxConnLifetime
	conf.MaxConnIdleTime = cfg.MaxConnIdleTime
	conf.AfterConnect = dtn.RegisterDataTypes
	conf.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &traceLogger{sLogger},
		LogLevel: traceLogLevel,
	}

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create pg pool: %w", err)
	}

	return pool, nil
}
