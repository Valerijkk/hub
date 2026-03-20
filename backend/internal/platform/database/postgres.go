package database

import (
	"context"
	"log/slog"

	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/valerii/hub/backend/ent"
)

// NewPostgresClient creates an Ent client connected to PostgreSQL.
// It runs auto-migration in development; use Atlas for production migrations.
func NewPostgresClient(ctx context.Context, url string) (*ent.Client, error) {
	drv, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	client := ent.NewClient(ent.Driver(drv))

	// Auto-migrate in development. Use Atlas for production.
	if err := client.Schema.Create(ctx); err != nil {
		slog.Error("failed to create schema", "error", err)
		return nil, err
	}

	slog.Info("connected to PostgreSQL")
	return client, nil
}
