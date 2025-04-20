package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"

	"github.com/COTBU/sotbi.lib/pkg/log"
)

type Migrator struct {
	logger log.Logger
	mg     *migrate.Migrate
}

func NewMigrator(migrationsPath string, logger log.Logger, dsn string) (*Migrator, error) {
	fileSource, err := (&file.File{}).Open(fmt.Sprintf("file://%s", migrationsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open file migration source: %w", err)
	}

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	sqlConf := &SQLConfig{
		DSN: dsn,
	}

	conn, err := NewConnection(sqlConf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	conf := &postgres.Config{
		DatabaseName: config.Database,
	}
	if searchPath, ok := config.RuntimeParams["search_path"]; ok {
		conf.SchemaName = searchPath
	}

	driver, err := postgres.WithInstance(conn.DB, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"file",
		fileSource,
		config.Database,
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{
		logger: logger,
		mg:     m,
	}, nil
}

func (m *Migrator) Up() error {
	defer func() {
		sourceErr, dbErr := m.mg.Close()
		if sourceErr != nil {
			m.logger.Error("error closing source after migrate", "up", sourceErr.Error())
		}

		if dbErr != nil {
			m.logger.Error("error closing database connection after migrate", "up", dbErr.Error())
		}
	}()

	if err := m.mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply up migrations: %w", err)
	}

	m.logger.Info("up migrations applied successfully")

	return nil
}

func (m *Migrator) Down() error {
	defer func() {
		sourceErr, dbErr := m.mg.Close()
		if sourceErr != nil {
			m.logger.Error("error closing source after migrate", "down", sourceErr.Error())
		}

		if dbErr != nil {
			m.logger.Error("error closing database connection after migrate", "down", dbErr.Error())
		}
	}()

	if err := m.mg.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply down migrations: %w", err)
	}

	m.logger.Info("down migrations applied successfully")

	return nil
}
