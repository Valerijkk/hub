//go:build hubtest

package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/database"
)

// TestNewPostgresClient_UnreachableDSN_FailClosed_A3W44001 — sql.Open ok, Ping fail → err, nil client/DB.
// Mutation: убрать PingContext → тест краснеет (ожидал error).
func TestNewPostgresClient_UnreachableDSN_FailClosed_A3W44001(t *testing.T) {
	t.Setenv("DEV_MODE", "false")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, rawDB, err := database.NewPostgresClient(ctx, "postgres://hub:hub@127.0.0.1:1/hub?sslmode=disable")
	require.Error(t, err, "мёртвый DSN обязан fail-closed через PingContext")
	require.Nil(t, client)
	require.Nil(t, rawDB)
}

// TestNewPostgresClient_LiveDSN_PingOK_A3W44001 — живой DSN (compose) → client+rawDB, Ping ok.
func TestNewPostgresClient_LiveDSN_PingOK_A3W44001(t *testing.T) {
	uri := os.Getenv("POSTGRES_TEST_URL")
	if uri == "" {
		uri = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
	}

	t.Setenv("DEV_MODE", "false")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, rawDB, err := database.NewPostgresClient(ctx, uri)
	if err != nil {
		t.Skipf("PostgreSQL недоступен для happy-path Ping (POSTGRES_TEST_URL=%q): %v", uri, err)
	}
	require.NotNil(t, client)
	require.NotNil(t, rawDB)
	t.Cleanup(func() { _ = client.Close() })

	require.NoError(t, rawDB.PingContext(ctx))
}
