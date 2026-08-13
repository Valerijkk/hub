//go:build hubtest

package config_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/platform/crypto"
)

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
