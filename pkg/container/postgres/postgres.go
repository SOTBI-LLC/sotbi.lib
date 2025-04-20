package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/COTBU/sotbi.lib/pkg/container"
	"github.com/COTBU/sotbi.lib/pkg/database"
	"github.com/COTBU/sotbi.lib/pkg/log"
	"github.com/COTBU/sotbi.lib/pkg/log/slog"
)

const (
	dbName                 = "energy"
	dbUser                 = "energy"
	numLogOccurrence       = 2
	postgresStartupTimeout = 5 * time.Second
)

type Service interface {
	container.Docker
	ApplyMigrations(context.Context, string) error
	Pool() *pgxpool.Pool
	DB() *sqlx.DB
	LoadFixture(context.Context, string) error
	DSN(context.Context) (string, error)
}

type Postgres struct {
	db       *sqlx.DB
	pool     *pgxpool.Pool
	log      log.Logger
	request  testcontainers.GenericContainerRequest
	postgres *postgres.PostgresContainer
}

func New(options ...testcontainers.CustomizeRequestOption) *Postgres {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"POSTGRES_DB": dbName,
			},
			ExposedPorts: []string{"5432/tcp"},
		},
	}

	for _, option := range options {
		_ = option(&req)
	}

	logger := slog.New("debug")

	return &Postgres{
		request: req,
		log:     logger,
	}
}

func (p *Postgres) Port(ctx context.Context, port nat.Port) (nat.Port, error) {
	return p.postgres.MappedPort(ctx, port)
}

func (p *Postgres) Start(ctx context.Context, datTypeNames []string) error {
	pgContainer, err := postgres.Run(
		ctx,
		"postgres:14-alpine",
		testcontainers.CustomizeRequest(p.request),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(numLogOccurrence).
				WithStartupTimeout(postgresStartupTimeout)))
	if err != nil {
		return fmt.Errorf("failed to start postgres container")
	}

	p.postgres = pgContainer

	connString, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	conf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return err
	}

	dtn := database.DataTypes(datTypeNames)

	conf.AfterConnect = dtn.RegisterDataTypes

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to create pg pool: %w", err)
	}

	p.pool = pool

	p.db = sqlx.NewDb(stdlib.OpenDBFromPool(pool), "pgx")

	return nil
}

func (p *Postgres) Stop(ctx context.Context) error {
	return p.postgres.Terminate(ctx)
}

func (p *Postgres) Status(ctx context.Context) string {
	if p.postgres == nil {
		return ""
	}

	state, err := p.postgres.State(ctx)
	if err != nil {
		panic(err)
	}

	return state.Status
}

func (p *Postgres) Name(ctx context.Context) string {
	if p.postgres == nil {
		return p.request.Name
	}

	name, err := p.postgres.Name(ctx)
	if err != nil {
		panic(err)
	}

	return name
}

func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

func (p *Postgres) DB() *sqlx.DB {
	return p.db
}

func (p *Postgres) DSN(ctx context.Context) (string, error) {
	connString, err := p.postgres.ConnectionString(ctx)
	if err != nil {
		return "", err
	}

	return connString, nil
}

func (p *Postgres) ApplyMigrations(ctx context.Context, path string) error {
	connString, err := p.postgres.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	m, err := database.NewMigrator(path, p.log, connString)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	err = m.Up()
	if err != nil {
		return fmt.Errorf("failed to apply up migrations: %w", err)
	}

	return nil
}

func (p *Postgres) LoadFixture(ctx context.Context, path string) error {
	connString, err := p.postgres.ConnectionString(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	fixture, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("pgx"),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.ResetSequencesTo(1),
		testfixtures.Directory(path),
	)
	if err != nil {
		return fmt.Errorf("failed to create fixture loader: %w", err)
	}

	err = fixture.Load()
	if err != nil {
		return fmt.Errorf("failed to load fixtures: %w", err)
	}

	p.log.Info("fixtures loaded")

	return nil
}
