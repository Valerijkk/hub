package audit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func newAuditRepoTestEnv(t *testing.T) Repository {
	t.Helper()
	dsn := "file:audittest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return NewRepository(client)
}

func TestRepository_ListAuditLogs_Pagination(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := repo.CreateAuditLog(ctx, "request", "http", "/api/"+uuid.New().String()[:4], nil, nil, nil, "", "")
		require.NoError(t, err)
	}

	items, total, err := repo.ListAuditLogs(ctx, AuditLogFilters{Page: 2, PerPage: 2})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, items, 2)

	items, total, err = repo.ListAuditLogs(ctx, AuditLogFilters{Page: 0, PerPage: 0})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, items, 5)
}

func TestRepository_ListAuditLogs_Filters(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()
	entityID := uuid.New().String()

	_, err := repo.CreateAuditLog(ctx, "login", "session", entityID, &userID, &projectID, nil, "10.1.2.3", "agent-a")
	require.NoError(t, err)
	_, err = repo.CreateAuditLog(ctx, "logout", "session", uuid.New().String(), nil, nil, nil, "10.9.9.9", "")
	require.NoError(t, err)

	items, total, err := repo.ListAuditLogs(ctx, AuditLogFilters{
		Action: "login", EntityType: "session", UserID: &userID,
		ProjectID: &projectID, EntityID: entityID, IPAddress: "10.1.2",
		Page: 1, PerPage: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, entityID, items[0].EntityID)
}

func TestRepository_ListAuditLogs_SlowOnly(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()

	slowDetails := map[string]interface{}{"duration_ms": int64(750), "method": "GET"}
	fastDetails := map[string]interface{}{"duration_ms": int64(120), "method": "GET"}
	_, err := repo.CreateAuditLog(ctx, "request", "http", "/api/slow", nil, nil, slowDetails, "", "")
	require.NoError(t, err)
	_, err = repo.CreateAuditLog(ctx, "request", "http", "/api/fast", nil, nil, fastDetails, "", "")
	require.NoError(t, err)
	_, err = repo.CreateAuditLog(ctx, "login", "session", uuid.New().String(), nil, nil, slowDetails, "", "")
	require.NoError(t, err)

	items, total, err := repo.ListAuditLogs(ctx, AuditLogFilters{SlowOnly: true, Page: 1, PerPage: 50})
	if err != nil {
		if strings.Contains(err.Error(), "unrecognized token") || strings.Contains(err.Error(), "syntax error") {
			t.Skip("SlowOnly filter uses PostgreSQL JSON operators; skipped on sqlite")
		}
	}
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "/api/slow", items[0].EntityID)
}

func TestRepository_ListAuditLogs_DateRange(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()
	now := time.Now().UTC()

	_, err := repo.CreateAuditLog(ctx, "tick", "http", "/recent", nil, nil, nil, "", "")
	require.NoError(t, err)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)
	items, total, err := repo.ListAuditLogs(ctx, AuditLogFilters{
		DateFrom: &from, DateTo: &to, Page: 1, PerPage: 50,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
}
