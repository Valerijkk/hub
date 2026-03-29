// Package database управляет подключениями к PostgreSQL и MongoDB.
package database

import (
	"context"
	dbsql "database/sql"
	"log/slog"

	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/valerii/hub/backend/ent"
)

// NewPostgresClient создаёт Ent-клиент, подключённый к PostgreSQL.
// Выполняет автомиграцию в режиме разработки; для продакшена используйте Atlas.
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

// OpenRawDB открывает прямое подключение к PostgreSQL для raw SQL-запросов.
// Используется модулем поиска для tsvector-запросов, которые не поддерживает Ent ORM.
func OpenRawDB(url string) (*dbsql.DB, error) {
	return dbsql.Open("postgres", url)
}
