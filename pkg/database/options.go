package database

import "time"

type Option func(*Conn)

// MaxOpenConns -.
func MaxOpenConns(size int) Option {
	return func(c *Conn) {
		c.maxOpenConns = size
	}
}

// MaxIdleConns -.
func MaxIdleConns(size int) Option {
	return func(c *Conn) {
		c.maxIdleConns = size
	}
}

// ConnMaxLifetime -.
func ConnMaxLifetime(timeout time.Duration) Option {
	return func(c *Conn) {
		c.connMaxLifetime = timeout
	}
}

func ConnDSN(dsn string) Option {
	return func(c *Conn) {
		c.dsn = dsn
	}
}
