// Package postgres implements PostgreSQL infrastructure details.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a pool and verifies that PostgreSQL accepts connections before
// allowing the API to start serving requests.
func NewPool(ctx context.Context, databaseURL string, tracer pgx.QueryTracer) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("configurar el pool PostgreSQL: %w", err)
	}
	poolConfig.ConnConfig.Tracer = tracer

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("crear el pool PostgreSQL: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("comprobar la conexión con PostgreSQL: %w", err)
	}

	return pool, nil
}
