// Package postgres implementa los detalles de infraestructura PostgreSQL.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool crea un pool y verifica que PostgreSQL acepta conexiones antes de
// permitir que la API empiece a atender peticiones.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("configurar el pool PostgreSQL: %w", err)
	}

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
