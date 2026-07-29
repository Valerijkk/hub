package project_test

// TASK-137 / A3W1-A5-001…006 — archive resurrect, permanent-delete admin gate,
// CreateTeam tx, invite MemberCount, template error contract, import 413.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/team"
	"github.com/valerii/hub/backend/internal/platform/middleware"
	"github.com/valerii/hub/backend/internal/project"
)

// TestArchiveUnarchive_SoftDeleted_NoResurrect_A3W1_A5_001 —
// archive/unarchive на trash → 404; deleted_at остаётся; Restore очищает.
// Mutation: снова ClearDeletedAt в UpdateStatus(..., false) / убрать guard → assert DeletedAt краснеет.
func TestArchiveUnarchive_SoftDeleted_NoResurrect_A3W1_A5_001(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "TrashGuard", "TRG", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteProject(ctx, projectID))
	require.NotNil(t, repo.projects[projectID].DeletedAt)

	_, err = svc.ArchiveProject(ctx, projectID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NotNil(t, repo.projects[projectID].DeletedAt, "archive must not clear deleted_at")
	assert.Equal(t, entproject.StatusDeleted, repo.projects[projectID].Status)

	_, err = svc.UnarchiveProject(ctx, projectID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NotNil(t, repo.projects[projectID].DeletedAt, "unarchive must not clear deleted_at")
	assert.Equal(t, entproject.StatusDeleted, repo.projects[projectID].Status)

	restored, err := svc.RestoreProject(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, "active", restored.Status)
	assert.Nil(t, repo.projects[projectID].DeletedAt)
}

// TestPermanentDelete_ActiveAndNonAdmin_A3W1_A5_002 —
// active → 400; deleted+non-admin → 403; deleted+admin → OK.
// Mutation: убрать status=deleted / admin gate → assert Error краснеет.
func TestPermanentDelete_ActiveAndNonAdmin_A3W1_A5_002(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	adminID := uuid.New()
	repo.admins[adminID] = true

	created, err := svc.CreateProject(ctx, ownerID, "HardDel", "HDL", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	err = svc.PermanentDelete(ctx, projectID, adminID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "trash")

	require.NoError(t, svc.DeleteProject(ctx, projectID))

	err = svc.PermanentDelete(ctx, projectID, ownerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")

	require.NoError(t, svc.PermanentDelete(ctx, projectID, adminID))
	_, ok := repo.projects[projectID]
	assert.False(t, ok)
}

// TestCreateTeam_TxRollbackOnMidFail_A3W1_A5_003 —
// mid-fail после CreateTeam → orphan team нет; happy path — creator member.
// Mutation: CreateTeam+AddMember без общей Ent-tx → после inject-fail команда остаётся.
func TestCreateTeam_TxRollbackOnMidFail_A3W1_A5_003(t *testing.T) {
	env := newProjectRepoEnv(t)

	project.ExportSetAfterCreateTeamHook(func(context.Context) error {
		return errors.New("injected mid-fail")
	})
	t.Cleanup(func() { project.ExportSetAfterCreateTeamHook(nil) })

	svc := project.NewProjectService(env.repo, nil, nil)
	_, err := svc.CreateTeam(env.ctx, env.owner.ID, "OrphanRisk", "should roll back")
	require.Error(t, err)

	count, err := env.client.Team.Query().Where(team.NameEQ("OrphanRisk")).Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "mid-fail must roll back team create")
}

// TestCreateTeam_AtomicSuccess_A3W1_A5_003 — creator сразу member.
func TestCreateTeam_AtomicSuccess_A3W1_A5_003(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	resp, err := svc.CreateTeam(env.ctx, env.owner.ID, "AtomicOK", "ok")
	require.NoError(t, err)
	teamID, err := uuid.Parse(resp.ID)
	require.NoError(t, err)

	mem, err := env.repo.FindTeamMember(env.ctx, teamID, env.owner.ID)
	require.NoError(t, err)
	assert.NotNil(t, mem)
	assert.Equal(t, 1, resp.MemberCount)
}

// TestJoinByInvite_MemberCountError_A3W1_A5_004 —
// ошибка MemberCount после join → internal, не member_count:0 success.
// Mutation: снова count, _ := MemberCount → assert Error краснеет (ложный 0).
func TestJoinByInvite_MemberCountError_A3W1_A5_004(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "InviteCnt", "IVC", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(ctx, projectID, ownerID, "developer", 5, nil)
	require.NoError(t, err)

	repo.memberCountErr = fmt.Errorf("count unavailable")
	_, err = svc.JoinByInvite(ctx, inv.Code, joinerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "member count")

	// Join всё же добавил member (count падает после AddMember) — проверяем, что не отдали ложный success.
	_, err = repo.FindMember(ctx, projectID, joinerID)
	require.NoError(t, err)
}

// listFailTemplateRepo — ListTemplates/GetTemplate возвращают «секретную» ошибку для error-contract.
type listFailTemplateRepo struct {
	mockRepo
	listErr   error
	exportErr error
}

func (r *listFailTemplateRepo) ListTemplates(context.Context) ([]*ent.ProjectTemplate, error) {
	return nil, r.listErr
}

func (r *listFailTemplateRepo) GetTemplate(context.Context, uuid.UUID) (*ent.ProjectTemplate, error) {
	return nil, r.exportErr
}

// TestTemplateListExport_NoRawErrorLeak_A3W1_A5_005 —
// List/Export при inject err → body без raw internal string.
// Mutation: вернуть fiber.NewError(..., err.Error()) → assert NotContains краснеет.
func TestTemplateListExport_NoRawErrorLeak_A3W1_A5_005(t *testing.T) {
	secret := "SECRET_DB_DSN=postgres://internal"
	repo := &listFailTemplateRepo{
		mockRepo:  *newMockRepo(),
		listErr:   errors.New(secret),
		exportErr: errors.New(secret),
	}
	svc := project.NewTemplateService(repo)
	h := project.NewTemplateHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		c.Locals("requestID", "req-a5-005")
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodGet, "/api/project-templates/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotContains(t, body["error"], secret)
	assert.NotContains(t, body["error"], "SECRET_DB_DSN")
	_, hasErr := body["error"]
	_, hasRID := body["request_id"]
	assert.True(t, hasErr)
	assert.True(t, hasRID)

	req = httptest.NewRequest(http.MethodGet, "/api/project-templates/"+uuid.New().String()+"/export", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	body = map[string]string{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.NotContains(t, body["error"], secret)
}

// TestTemplateImport_Oversize_413_A3W1_A5_006 —
// multipart > maxTemplateImportBytes → 413.
// Mutation: убрать LimitReader/size check → assert StatusRequestEntityTooLarge краснеет.
func TestTemplateImport_Oversize_413_A3W1_A5_006(t *testing.T) {
	env := newTemplateRoutesEnv(t)
	max := project.ExportMaxTemplateImportBytes()
	payload := []byte(strings.Repeat("x", max+1))

	body := &bytes.Buffer{}
	writer := multipartWriter(t, body, "file", "huge.json", payload)
	require.NoError(t, writer.Close())

	// BodyLimit выше лимита импорта, чтобы дойти до handler size-check.
	app := fiber.New(fiber.Config{
		BodyLimit:     max + 2*1024*1024,
		ErrorHandler:  middleware.ErrorHandler,
	})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", env.userID.String())
		return c.Next()
	}
	h := project.NewTemplateHandler(project.NewTemplateService(env.repo))
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodPost, "/api/project-templates/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req, 60*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}
