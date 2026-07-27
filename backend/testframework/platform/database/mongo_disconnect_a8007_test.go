package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/database"
)

// TestMongoDisconnectCleanup_A8_007 — cleanup Disconnect закрывает pool;
// повторный Ping после cleanup обязан упасть.
func TestMongoDisconnectCleanup_A8_007(t *testing.T) {
	uri := os.Getenv("MONGODB_TEST_URL")
	if uri == "" {
		uri = "mongodb://hub:hub@localhost:27017/?authSource=admin"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.NewMongoClient(ctx, uri, "hub_a8007_"+t.Name())
	if err != nil {
		t.Skip("MongoDB недоступен:", err)
	}

	cleanup := database.MongoDisconnectCleanup(db)
	require.NotNil(t, cleanup)

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer pingCancel()
	require.NoError(t, db.Client().Ping(pingCtx, nil))

	cleanup()

	afterCtx, afterCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer afterCancel()
	err = db.Client().Ping(afterCtx, nil)
	require.Error(t, err, "A8-007: after Disconnect Ping must fail")
}

// TestMongoDisconnectCleanup_NilSafe_A8_007 — nil db не паникует.
func TestMongoDisconnectCleanup_NilSafe_A8_007(t *testing.T) {
	require.NotPanics(t, database.MongoDisconnectCleanup(nil))
}
