//go:build hubtest

package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	redisclient "github.com/valerii/hub/backend/internal/platform/redis"
)

// TestNewClient_PingFailCloses_W4A2004 — Ping fail → Close, нет висящего клиента.
func TestNewClient_PingFailCloses_W4A2004(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := redisclient.NewClient(ctx, "redis://127.0.0.1:1/0")
	require.Error(t, err)
	require.Nil(t, client, "failed Ping must return nil client after Close")
}
