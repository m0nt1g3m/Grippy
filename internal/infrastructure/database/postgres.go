package database

import (
	"context"
	"fmt"
	"time"

	"Grippy/internal/infrastructure/logging"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool   *pgxpool.Pool
	logger *logging.Logger
}

var ErrNoRows = pgx.ErrNoRows

func NewPostgresClient(connectLink string, logger *logging.Logger) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(connectLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	Pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := Pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping pool: %w", err)
	}

	logger.Info("Database initialized successfully")
	return &Postgres{Pool: Pool, logger: logger}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
		p.logger.Info("Database closed successfully")
	}
}
