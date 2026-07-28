package hr_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/internal/hr"
)

// TestEntRepository_LeaveSoftDelete_A2W8_3_001 — SoftDeleteLeave → deleted_at; lists hide HR-PII.
// Mutation: drop DeletedAtIsNil from ListLeaveByRequester → soft-deleted listed → FAIL.
func TestEntRepository_LeaveSoftDelete_A2W8_3_001(t *testing.T) {
	dsn := "file:hrsoft097_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	member, err := client.User.Create().
		SetEmail("leave-" + uuid.New().String()[:8] + "@test.com").
		SetName("Leave Member").
		Save(ctx)
	require.NoError(t, err)

	repo := hr.NewEntRepository(client)
	start := time.Now().UTC().Truncate(24 * time.Hour)
	end := start.AddDate(0, 0, 5)
	created, err := repo.CreateLeaveRequest(ctx, member.ID, leaverequest.LeaveTypeSick, start, end, "PII reason")
	require.NoError(t, err)

	require.NoError(t, repo.SoftDeleteLeave(ctx, created.ID))

	list, err := repo.ListLeaveByRequester(ctx, member.ID)
	require.NoError(t, err)
	for _, item := range list {
		assert.NotEqual(t, created.ID, item.ID)
	}

	_, err = repo.GetLeaveRequest(ctx, created.ID)
	require.True(t, ent.IsNotFound(err))

	raw, err := client.LeaveRequest.Get(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	src, err := os.ReadFile(filepath.Join(root, "backend", "internal", "hr", "repository.go"))
	require.NoError(t, err)
	assert.Contains(t, string(src), "SetDeletedAt")
	assert.Contains(t, string(src), "DeletedAtIsNil")
}
