package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	_maxIdleConns    = 1
	_maxOpenConns    = 10
	_connMaxLifetime = time.Second
	_DSN             = "host=127.0.0.1 user=energy password=%s dbname=energy " +
		"port=5432 TimeZone=Europe/Moscow sslmode=disable"
)

// Writer log writer interface
type Writer interface {
	Printf(string, ...interface{})
}

type Conn struct {
	*gorm.DB
  	Writer
	db              *sql.DB
	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration
	dsn             string
}

// Connect func.
func Connect(opts ...Option) (*Conn, error) {
	conn := &Conn{
		maxIdleConns:    _maxIdleConns,
		maxOpenConns:    _maxOpenConns,
		connMaxLifetime: _connMaxLifetime,
		dsn:             _DSN,
    Writer:          log.New(os.Stdout, "\r\n", log.LstdFlags),
	}
	// Custom options
	for _, opt := range opts {
		opt(conn)
	}

	sqlDB, err := sql.Open("pgx", fmt.Sprintf(conn.dsn, os.Getenv("ENERGY_DB_PASS")))
	if err != nil {
		return nil, err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(conn.maxIdleConns)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(conn.maxOpenConns)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(conn.connMaxLifetime)

	conn.db = sqlDB

  newLogger := logger.New(
  conn.Writer,
  logger.Config{
    SlowThreshold:              500*time.Millisecond,   // Slow SQL threshold
    LogLevel:                   logger.Error, // Log level
    IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
    Colorful:                  false,          // Disable color
  },
)

	conn.DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{CreateBatchSize: 100, Logger: newLogger })
	if err != nil {
		return nil, err
	}

	conn.Exec("set timezone to 'Europe/Moscow'")

	return conn, nil
}

// SetNullFieldDB func.
func SetNullFieldDB(db *gorm.DB, table, field string, id int) (err error) {
	err = db.
		// Debug().
		Table(table).
		Where("id=?", id).
		Updates(map[string]interface{}{field: nil}).
		Error
	if err != nil {
		return err
	}

	return nil
}

func (c *Conn) Ping(ctx context.Context) error {
	err := c.db.Ping()
	if err != nil {
		return err
	}

	c.DB.Logger.Info(ctx, "%#v\n", c.db.Stats())

	return nil
}

func (c *Conn) Close() error {
	// Close
	return c.db.Close()
}
