package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type SQLConfig struct {
	DSN             string
	MaxOpenConns    *int
	MaxIdleConns    *int
	ConnMaxLifetime *time.Duration
}

func NewConnection(cfg *SQLConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if cfg.MaxOpenConns != nil {
		db.SetMaxOpenConns(*cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns != nil {
		db.SetMaxIdleConns(*cfg.MaxIdleConns)
	}

	if cfg.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(*cfg.ConnMaxLifetime)
	}

	return db, nil
}
