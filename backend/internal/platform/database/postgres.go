// Package database управляет подключениями к PostgreSQL и MongoDB.
package database

import (
	"context"
	dbsql "database/sql"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"

	"github.com/valerii/hub/backend/ent"
)

// P1-BE-25: лимиты пула. MaxOpenConns=0 в database/sql = безлимит →
// исчерпание max_connections Postgres под нагрузкой.
const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 10
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

// applyPoolLimits задаёт потолок open/idle и lifetime для *sql.DB.
func applyPoolLimits(db *dbsql.DB) {
	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetMaxIdleConns(defaultMaxIdleConns)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)
}

// NewPostgresClient создаёт Ent-клиент, подключённый к PostgreSQL.
// Выполняет автомиграцию в режиме разработки; для продакшена используйте Atlas.
//
// Возвращает также raw `*sql.DB`, разделяющий пул соединений с Ent —
// его можно переиспользовать в health-probe (BE-H5) и других местах,
// где нужен прямой Ping, чтобы не открывать каждый раз новое подключение.
func NewPostgresClient(ctx context.Context, url string) (*ent.Client, *dbsql.DB, error) {
	drv, err := sql.Open("postgres", url)
	if err != nil {
		return nil, nil, err
	}
	rawDB := drv.DB()
	applyPoolLimits(rawDB)
	client := ent.NewClient(ent.Driver(drv))

	// Auto-migrate в dev; в production — Atlas migrations.
	if err := client.Schema.Create(ctx); err != nil {
		slog.Error("failed to create schema", "error", err)
		return nil, nil, err
	}

	slog.Info("connected to PostgreSQL")
	return client, rawDB, nil
}

// OpenRawDB открывает прямое подключение к PostgreSQL для raw SQL-запросов.
// Используется модулем поиска для tsvector-запросов, которые не поддерживает Ent ORM.
func OpenRawDB(url string) (*dbsql.DB, error) {
	db, err := dbsql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	applyPoolLimits(db)
	return db, nil
}
