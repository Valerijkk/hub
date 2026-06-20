package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
)

func TestAuditService_DeactivateUser_InvalidatesCache(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	target := uuid.New()
	repo := &mockAuditRepo{}
	svc := NewAuditService(repo, rdb)

	require.NoError(t, svc.DeactivateUser(context.Background(), target))
	require.Contains(t, repo.deactivateIDs, target)
	require.Contains(t, repo.deleteSessIDs, target)
}

type trashHandlerMockRepo struct {
	adminMockRepo
	items []TrashItem
	total int
}

func (m *trashHandlerMockRepo) ListDeletedIssues(_ context.Context, _ *uuid.UUID, _ time.Time, _, _ int) ([]TrashItem, int, error) {
	return m.items, m.total, nil
}

func TestHandler_ListTrash_WithProjectFilter(t *testing.T) {
	adminID := uuid.New()
	projectID := uuid.New()
	repo := &trashHandlerMockRepo{
		adminMockRepo: adminMockRepo{adminID: adminID},
		items:         []TrashItem{{ID: uuid.New(), Name: "Issue", EntityType: "issue", DeletedAt: time.Now().UTC()}},
		total:         1,
	}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/trash?type=issue&project_id="+projectID.String(), nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuditService_RestoreItem_Issue(t *testing.T) {
	repo := &userMgmtMockRepo{}
	svc := NewAuditService(repo, nil)
	id := uuid.New()

	require.NoError(t, svc.RestoreItem(context.Background(), "issue", id))
	assert.Equal(t, id, repo.restoreID)
}

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
}

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
