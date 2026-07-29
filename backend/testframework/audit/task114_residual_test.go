package audit_test

// TASK-114 / A2W6-A4/A5: residual limits, GDPR tx, purge/trash mutation tests.
import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
)

// --- A2W6-A4-002: limits fail-closed ---

func TestAuditService_UpdateResourceLimit_RejectsNonPositive_A2W6A4002(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	for _, bad := range []string{"abc", "0", "-5", ""} {
		err := svc.UpdateResourceLimit(ctx, audit.LimitFileMaxSizeMB, bad)
		require.ErrorIs(t, err, audit.ErrInvalidResourceLimitValue, "value=%q", bad)
	}
}

func TestAuditService_CheckLimit_InvalidValueDenies_A2W6A4002(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	require.NoError(t, repo.UpsertSystemSetting(ctx, audit.LimitUserMaxProjects, "not-a-number", "test"))
	ok, err := svc.CheckLimit(ctx, audit.LimitUserMaxProjects, 1)
	require.NoError(t, err)
	assert.False(t, ok, "invalid Atoi must deny (fail-closed)")
}

func TestAuditService_CheckLimit_UnknownEmptyDenies_A2W6A4002(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)

	ok, err := svc.CheckLimit(context.Background(), "limit.missing.unknown", 0)
	require.NoError(t, err)
	assert.False(t, ok, "unknown key without value must deny")
}

func TestAuditService_CheckLimit_KnownEmptyUsesDefault_A2W6A4002(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	svc := audit.NewAuditService(repo, nil)
	ctx := context.Background()

	// default max_projects = 20
	ok, err := svc.CheckLimit(ctx, audit.LimitUserMaxProjects, 19)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = svc.CheckLimit(ctx, audit.LimitUserMaxProjects, 20)
	require.NoError(t, err)
	assert.False(t, ok)
}

// --- A2W6-A4-005: middleware recover ---

func TestHTTPAuditMiddleware_RecorderPanicRecovered_A2W6A4005(t *testing.T) {
	var hit atomic.Bool
	app := fiber.New()
	app.Use(audit.HTTPAuditMiddleware(func(_ context.Context, _ audit.HTTPAuditEntry) {
		hit.Store(true)
		panic("recorder boom")
	}))
	app.Get("/api/test-panic", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/test-panic", nil)
	resp, err := app.Test(req, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if hit.Load() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.True(t, hit.Load(), "recorder must have been invoked")
	// process still alive — Test completed without crash
}

// --- A2W6-A4-006: HTTP GDPR founder → 403 ---

func TestHandler_DeleteUserGDPR_FounderForbidden_A2W6A4006(t *testing.T) {
	adminID := uuid.New()
	founderID := uuid.New()
	repo := &founderHTTPRepo{
		adminMockRepo: adminMockRepo{adminID: adminID},
		targetRole:    user.SystemRoleFounder,
	}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+founderID.String()+"/gdpr", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// --- A2W6-A4-003: Ent tx rollback mid-cascade ---

// hardDeleteFailWrap — fail HardDelete внутри WithTx (Ent rollback proof без Mongo-in-tx).
type hardDeleteFailWrap struct {
	audit.Repository
}

func (w *hardDeleteFailWrap) HardDeleteUser(context.Context, uuid.UUID) error {
	return errors.New("forced hard delete fail")
}

func (w *hardDeleteFailWrap) WithTx(ctx context.Context, fn func(audit.Repository) error) error {
	return w.Repository.WithTx(ctx, func(txRepo audit.Repository) error {
		return fn(&hardDeleteFailWrap{Repository: txRepo})
	})
}

func TestDeleteUserGDPR_CascadeRollback_A2W6A4003(t *testing.T) {
	env := newAuditTrashEnv(t)
	iss := env.createIssue(t, nil)

	_, err := env.client.Comment.Create().
		SetIssueID(iss.ID).
		SetAuthorID(env.userA).
		SetBodyText("keep-me").
		SetBodyJSON(map[string]any{"type": "doc"}).
		Save(env.ctx)
	require.NoError(t, err)

	eraser := &stubMessageEraser{}
	svc := audit.NewAuditService(&hardDeleteFailWrap{Repository: env.repo}, nil)
	svc.SetMessageEraser(eraser)

	err = svc.DeleteUserGDPR(env.ctx, env.userB, env.userA, nil)
	require.Error(t, err)

	// Comments must be rolled back (A2W6-A4-003).
	n, err := env.client.Comment.Query().Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "comment must survive rollback after mid-cascade fail")

	_, err = env.client.User.Get(env.ctx, env.userA)
	require.NoError(t, err, "user must not be hard-deleted")

	// A3W6-4-001: Mongo erase не вызывается при Ent rollback.
	assert.Empty(t, eraser.calls, "mongo erase must stay intact when Ent tx rolls back")
}

// --- A2W6-A5-001: PurgeExpired call counts + cutoffs ---

// A3W6-5-001 / A3W2-A5-001: wantKeys×9 включая crm_* и hubbases.
// Мутация: убрать PurgeExpiredHubBases из PurgeExpired → callCounts краснеет.
func TestAuditService_PurgeExpired_CapturesCutoffs_A2W6A5001(t *testing.T) {
	repo := &purgeMockRepo{
		purgeCounts: map[string]int{
			"issues": 1, "test_cases": 1, "files": 1,
			"projects": 1, "newsposts": 1, "scheduled_calls": 1,
			"crm_contacts": 1, "crm_deals": 1, "hubbases": 1,
		},
		cutoffs: make(map[string]time.Time),
	}
	svc := audit.NewAuditService(repo, nil)

	before := time.Now().Add(-audit.TrashRetentionDays*24*time.Hour - time.Minute)
	require.NoError(t, svc.PurgeExpired(context.Background()))
	after := time.Now().Add(-audit.TrashRetentionDays*24*time.Hour + time.Minute)

	wantKeys := []string{
		"issues", "test_cases", "files", "projects", "newsposts", "scheduled_calls",
		"crm_contacts", "crm_deals", "hubbases",
	}
	for _, k := range wantKeys {
		assert.Equal(t, 1, repo.callCounts[k], "purge %s must be called once", k)
		c, ok := repo.cutoffs[k]
		require.True(t, ok, "cutoff for %s", k)
		assert.True(t, !c.Before(before) && !c.After(after), "cutoff %s=%v", k, c)
	}
}

// --- A2W6-A5-004 / A3W6-5-002 / A3W2-A5-001: SoftDelete/Restore call map ×9 (hubbase) ---
// Мутация: убрать case "hubbase" → assert.Equal краснеет.

func TestAuditService_SoftDelete_RecordsAllTypes_A2W6A5004(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	id := uuid.New()
	types := []string{"issue", "test_case", "file", "project", "newspost", "scheduled_call", "crm_contact", "crm_deal", "hubbase"}
	for _, et := range types {
		require.NoError(t, svc.SoftDelete(context.Background(), et, id), et)
	}
	assert.Equal(t, types, repo.softDeleteCalls)
}

func TestAuditService_RestoreItem_RecordsAllTypes_A2W6A5004(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	id := uuid.New()
	types := []string{"issue", "test_case", "file", "project", "newspost", "scheduled_call", "crm_contact", "crm_deal", "hubbase"}
	for _, et := range types {
		require.NoError(t, svc.RestoreItem(context.Background(), et, id), et)
	}
	assert.Equal(t, types, repo.restoreCalls)
}

// --- A2W6-A5-006: project_id filter capture ---

type trashProjectFilterMock struct {
	adminMockRepo
	gotProjectID *uuid.UUID
	items        []audit.TrashItem
	total        int
}

func (m *trashProjectFilterMock) ListDeletedIssues(_ context.Context, projectID *uuid.UUID, _ time.Time, _, _ int) ([]audit.TrashItem, int, error) {
	m.gotProjectID = projectID
	return m.items, m.total, nil
}

func TestHandler_ListTrash_CapturesProjectID_A2W6A5006(t *testing.T) {
	adminID := uuid.New()
	projectID := uuid.New()
	repo := &trashProjectFilterMock{
		adminMockRepo: adminMockRepo{adminID: adminID},
		items:         []audit.TrashItem{{ID: uuid.New(), Name: "I", EntityType: "issue", DeletedAt: time.Now().UTC()}},
		total:         1,
	}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/trash?type=issue&project_id="+projectID.String(), nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, repo.gotProjectID)
	assert.Equal(t, projectID, *repo.gotProjectID)
}

// --- A2W6-A5-009: GDPR Success asserts cascade steps ---

func TestAuditService_DeleteUserGDPR_SuccessCascadeCalls_A2W6A5009(t *testing.T) {
	repo := &gdprFullCascadeRepo{}
	svc := audit.NewAuditService(repo, nil)
	target := uuid.New()

	require.NoError(t, svc.DeleteUserGDPR(context.Background(), uuid.New(), target, nil))
	assert.Equal(t, 1, repo.files)
	assert.Equal(t, 1, repo.projects)
	assert.Equal(t, 1, repo.channels)
	assert.Equal(t, 1, repo.teams)
	assert.Equal(t, 1, repo.scrubbed)
	require.Len(t, repo.hardDeleted, 1)
	assert.Equal(t, target, repo.hardDeleted[0])
}
