// Package database управляет подключениями к PostgreSQL и MongoDB.
package database

import (
	"context"
	dbsql "database/sql"
	"log/slog"
	"os"
	"strconv"
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
// W4-A4-003: Schema.Create (auto-migrate) только при DEV_MODE=true;
// в production схема применяется через Atlas migrate apply.
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

	if envBool("DEV_MODE", false) {
		if err := client.Schema.Create(ctx); err != nil {
			slog.Error("failed to create schema", "error", err)
			return nil, nil, err
		}
	}

	slog.Info("connected to PostgreSQL")
	return client, rawDB, nil
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
