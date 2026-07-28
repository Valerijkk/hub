package database_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/database"
)

// P1-BE-25: SetMaxOpenConns (и idle/lifetime) применяются к пулу.
func TestApplyPoolLimits_MaxOpenConns_P1_BE_25(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:pool_limits_"+t.Name()+"?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	// До лимитов MaxOpenConnections == 0 (безлимит в database/sql).
	assert.Equal(t, 0, db.Stats().MaxOpenConnections)

	database.ExportApplyPoolLimits(db)

	want := database.ExportDefaultMaxOpenConns()
	assert.Equal(t, want, db.Stats().MaxOpenConnections)
	assert.Greater(t, want, 0)
	assert.LessOrEqual(t, db.Stats().OpenConnections, want)
}

func TestOpenRawDB_AppliesPoolLimits_P1_BE_25(t *testing.T) {
	// Open не коннектится сразу — URL может быть недостижимым.
	db, err := database.OpenRawDB("postgresql://hub:hub@127.0.0.1:1/hub?sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	assert.Equal(t, database.ExportDefaultMaxOpenConns(), db.Stats().MaxOpenConnections)
}
