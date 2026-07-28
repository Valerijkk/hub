//go:build hubtest

package database_test

import (
	"testing"

	"github.com/valerii/hub/backend/internal/platform/database"
)

// TestEnvBool_DevModeGate_W4A4003 — Schema.Create gated by DEV_MODE env.
func TestEnvBool_DevModeGate_W4A4003(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	if !database.ExportEnvBool("DEV_MODE", false) {
		t.Fatal("DEV_MODE=true must enable auto-migrate gate")
	}
	t.Setenv("DEV_MODE", "false")
	if database.ExportEnvBool("DEV_MODE", false) {
		t.Fatal("DEV_MODE=false must disable auto-migrate")
	}
	t.Setenv("DEV_MODE", "")
	if database.ExportEnvBool("DEV_MODE", false) {
		t.Fatal("empty DEV_MODE must fall back to false")
	}
}
