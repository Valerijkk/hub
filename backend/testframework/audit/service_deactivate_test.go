package audit_test

// Тесты сервиса аудита: интеграционные и unit-сценарии testframework.
import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/internal/auth"
)

// TestAuditService_DeactivateUser_InvalidatesCache — A3W6-5-004 / ME-13:
// сеет user:status key и assert'ит Del после InvalidateUserStatusCache.
// Мутация: убрать InvalidateUserStatusCache → mr.Exists(key) остаётся true.
func TestAuditService_DeactivateUser_InvalidatesCache(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	target := uuid.New()
	cacheKey := auth.UserStatusCacheKey(target)
	require.NoError(t, rdb.Set(context.Background(), cacheKey, auth.ExportUserStatusCacheDeactivated(), auth.ExportUserStatusCacheTTL()).Err())
	require.True(t, mr.Exists(cacheKey), "seeded cache key must exist before deactivate")

	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, rdb)

	require.NoError(t, svc.DeactivateUser(context.Background(), target))
	require.Contains(t, repo.deactivateIDs, target)
	require.Contains(t, repo.deleteSessIDs, target)
	assert.False(t, mr.Exists(cacheKey), "InvalidateUserStatusCache must Del user status key")
}

// TestAuditService_RestoreItem_Issue проверяет бизнес-логику: audit сервис restore item issue.
func TestAuditService_RestoreItem_Issue(t *testing.T) {
	repo := &userMgmtMockRepo{}
	svc := audit.NewAuditService(repo, nil)
	id := uuid.New()

	require.NoError(t, svc.RestoreItem(context.Background(), "issue", id))
	assert.Equal(t, id, repo.restoreID)
}

// TestRepository_SoftDeleteAndRestoreFile проверяет слой репозитория: репозиторий soft удаление and restore file.
func TestRepository_SoftDeleteAndRestoreFile(t *testing.T) {
	env := newAuditTrashEnv(t)
	when := time.Now().UTC()

	file, err := env.client.FileMetadata.Create().
		SetFilename("doc.pdf").
		SetContentType("application/pdf").
		SetSize(100).
		SetS3Key("k1").
		SetS3Bucket("b1").
		SetProjectID(env.projectID).
		SetUploaderID(env.userA).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.SoftDeleteFile(env.ctx, file.ID, when))

	items, total, err := env.repo.ListDeletedFiles(env.ctx, &env.projectID, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)

	require.NoError(t, env.repo.RestoreFile(env.ctx, file.ID))

	got, err := env.client.FileMetadata.Get(env.ctx, file.ID)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt, "A2W6-A5-007: RestoreFile must ClearDeletedAt")
}

// TestRepository_ReassignReportedIssues проверяет слой репозитория: репозиторий reassign reported issues.
func TestRepository_ReassignReportedIssues(t *testing.T) {
	env := newAuditTrashEnv(t)
	iss := env.createIssue(t, nil)

	n, err := env.repo.ReassignReportedIssues(env.ctx, env.userA, env.userB)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	updated, err := env.client.Issue.Query().
		Where(issue.IDEQ(iss.ID)).
		WithReporter().
		Only(env.ctx)
	require.NoError(t, err)
	require.NotNil(t, updated.Edges.Reporter)
	assert.Equal(t, env.userB, updated.Edges.Reporter.ID)
}
