//go:build hubtest

package config_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/platform/crypto"
)

func setProductionBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DEV_MODE", "false")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("JWT_SECRET", "production-jwt-secret-at-least-32-bytes")
	t.Setenv("S3_ACCESS_KEY", "production-access-key")
	t.Setenv("S3_SECRET_KEY", "production-secret-key")
	t.Setenv("LIVEKIT_API_KEY", "production-livekit-key")
	t.Setenv("LIVEKIT_API_SECRET", "production-livekit-secret")
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", strings.Repeat("01", 32))
	t.Setenv("METRICS_ALLOWED_CIDRS", "127.0.0.1/32")
	t.Setenv("NATS_URL", "nats://produser:prodpass@nats:4222")
	t.Setenv("DATABASE_URL", "postgresql://produser:prodpass@db:5432/hub?sslmode=require")
	t.Setenv("MONGODB_URL", "mongodb://produser:prodpass@mongo:27017/hub?authSource=admin")
	t.Setenv("REDIS_URL", "redis://:prodpass@redis:6379/0")
	t.Setenv("ALLOW_INSECURE_DB_SSL", "false")
}

// TestLoad_productionRejectsDeriveDevKey_W4A4009 — prod+DeriveDevKey → validate fail.
func TestLoad_productionRejectsDeriveDevKey_W4A4009(t *testing.T) {
	setProductionBaseEnv(t)
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", hex.EncodeToString(crypto.DeriveDevKey()))

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for DeriveDevKey in production")
	}
	if !strings.Contains(err.Error(), "DeriveDevKey") && !strings.Contains(err.Error(), "INTEGRATION_ENCRYPTION_KEY") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestLoad_productionRejectsInsecureDBSSL_W4A4004 — sslmode=disable без override → fail.
func TestLoad_productionRejectsInsecureDBSSL_W4A4004(t *testing.T) {
	setProductionBaseEnv(t)
	t.Setenv("DATABASE_URL", "postgresql://produser:prodpass@db:5432/hub?sslmode=disable")
	t.Setenv("ALLOW_INSECURE_DB_SSL", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for sslmode=disable in production")
	}
	if !strings.Contains(err.Error(), "sslmode") && !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestLoad_productionAllowsInsecureDBSSLOverride_W4A4004 — явный override разрешает disable.
func TestLoad_productionAllowsInsecureDBSSLOverride_W4A4004(t *testing.T) {
	setProductionBaseEnv(t)
	t.Setenv("DATABASE_URL", "postgresql://produser:prodpass@db:5432/hub?sslmode=disable")
	t.Setenv("ALLOW_INSECURE_DB_SSL", "true")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if !cfg.Postgres.AllowInsecureDBSSL {
		t.Error("expected AllowInsecureDBSSL=true")
	}
}
