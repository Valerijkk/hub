package tms_test

// TASK-111 — A2W2 residual mutation/regression coverage.
import (
	"bytes"
	"context"
	"encoding/json"
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
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/hubcolumn"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/hubtable"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/bases"
	"github.com/valerii/hub/backend/internal/platform/middleware"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestTask111_UpdateTestCase_RequiresVersion_StructValidator — A2W2-1-003.
func TestTask111_UpdateTestCase_RequiresVersion_StructValidator(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	h := tms.NewHandler(svc)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New(fiber.Config{StructValidator: middleware.NewStructValidator()})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	grp := api.Group("/projects/:projectId/tms", authMw)
	grp.Post("/cases", h.CreateTestCase)
	grp.Put("/cases/:caseKey", h.UpdateTestCase)

	base := "/api/projects/" + projectID.String() + "/tms/cases"
	createBody := `{"title":"v-req","case_type":"classic","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created tms.TestCaseResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	req = httptest.NewRequest(http.MethodPut, base+"/"+created.CaseKey, bytes.NewReader([]byte(`{"title":"no-ver"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestTask111_WebhookRawBodyCap — A2W2-1-005.
func TestTask111_WebhookRawBodyCap(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	h := tms.NewHandler(svc)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New(fiber.Config{BodyLimit: 40 * 1024 * 1024})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	grp := api.Group("/projects/:projectId/tms", authMw)
	grp.Post("/webhook", h.WebhookAllure)

	const maxBytes = 5 * 1024 * 1024
	body := bytes.Repeat([]byte("x"), maxBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/tms/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	resp, err := app.Test(req, 30*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

// TestTask111_ParseAllure_DepthCap — A2W2-4-001.
func TestTask111_ParseAllure_DepthCap(t *testing.T) {
	nested := strings.Repeat(`[`, 40) + `]` + strings.Repeat(`]`, 39)
	_, err := tms.ParseAllureResults([]byte(nested))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "глубин")
}

// TestTask111_ParseJUnit_DepthCap — A2W2-4-001.
func TestTask111_ParseJUnit_DepthCap(t *testing.T) {
	nested := strings.Repeat(`<a>`, 40) + strings.Repeat(`</a>`, 40)
	_, err := tms.ParseJUnitXML([]byte(nested))
	require.Error(t, err)
}

// TestTask111_ImportAllure_FixedErrorMessage — A2W2-2-011.
func TestTask111_ImportAllure_FixedErrorMessage(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.ImportAllureResults(context.Background(), uuid.New(), uuid.New(), tms.ImportAllureRequest{
		Results: json.RawMessage(`{not-json`),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный allure JSON")
	assert.NotContains(t, err.Error(), "парсинг allure-results")
}

// TestTask111_ImportAllureToExisting_EpochGuard — A2W2-2-005.
func TestTask111_ImportAllureToExisting_EpochGuard(t *testing.T) {
	repo := newMockRepo()
	projectID := uuid.New()
	userID := uuid.New()
	svc := tms.NewTMSService(repo, nil)

	run, err := svc.CreateTestRun(context.Background(), projectID, userID, tms.CreateTestRunRequest{Title: "r"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	badStart := int64(-1)
	badStop := int64(9_999_999_999_999)
	payload, err := json.Marshal([]tms.AllureResult{{
		UUID:   "u1",
		Name:   "t",
		Status: "passed",
		Start:  &badStart,
		Stop:   &badStop,
	}})
	require.NoError(t, err)

	_, err = svc.ImportAllureToExistingRun(context.Background(), projectID, runID, userID, tms.ImportAllureRequest{Results: payload})
	require.NoError(t, err)
	for _, r := range repo.results {
		assert.Nil(t, r.DurationMs)
	}
}

// TestTask111_PDFTruncateRuneSafe — A2W2-4-002.
func TestTask111_PDFTruncateRuneSafe(t *testing.T) {
	data := tms.TestRunExportData{
		Title: "PDF",
		Results: []tms.ResultExportData{
			{CaseKey: "A-TC-1", Title: "t", Status: "failed", Comment: "ошибка断言中文🔥длинный"},
		},
	}
	var buf bytes.Buffer
	err := tms.ExportTestRunToPDF(&buf, data)
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}

// TestTask111_MatcherAutomationID_DBError — A2W2-4-003.
func TestTask111_MatcherAutomationID_DBError(t *testing.T) {
	repo := &errMatcherRepo{err: fmt.Errorf("db timeout")}
	matcher := tms.NewMatcher(repo)
	_, _, err := matcher.MatchTestCase(context.Background(), uuid.New(), &tms.AllureResult{
		UUID:   "1",
		Name:   "n",
		Status: "passed",
		Labels: []tms.AllureLabel{{Name: "automation_id", Value: "auto-1"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db timeout")
}

// TestTask111_MatcherCrossProjectDenial — A2W2-4-006.
func TestTask111_MatcherCrossProjectDenial(t *testing.T) {
	repo := newMockMatcherRepo()
	owner := uuid.New()
	foreign := uuid.New()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-99")
	tc.Edges.Project = &ent.Project{ID: owner}
	repo.automationIDMap["auto-login"] = tc

	matcher := tms.NewMatcher(repo)
	matched, _, err := matcher.MatchTestCase(context.Background(), foreign, &tms.AllureResult{
		UUID:   "auto-1",
		Name:   "loginTest",
		Status: "passed",
		Labels: []tms.AllureLabel{{Name: "automation_id", Value: "auto-login"}},
	})
	require.NoError(t, err)
	assert.Nil(t, matched)
}

// TestTask111_ListTestCases_EntPagination — A2W2-2-009 Ent Limit=1.
func TestTask111_ListTestCases_EntPagination(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	for i := 0; i < 3; i++ {
		_, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
			Title:       fmt.Sprintf("case-%d", i),
			CreatedByID: userID,
		})
		require.NoError(t, err)
	}
	items, total, err := repo.ListTestCases(ctx, projectID, tms.ListTestCasesParams{Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, items, 1)
}

// TestTask111_DeleteTestCase_ProjectScoped — A2W2-3-003.
func TestTask111_DeleteTestCase_ProjectScoped(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "del-me", CreatedByID: userID,
	})
	require.NoError(t, err)

	err = repo.DeleteTestCase(ctx, uuid.New(), tc.ID)
	require.Error(t, err)

	err = repo.DeleteTestCase(ctx, projectID, tc.ID)
	require.NoError(t, err)
}

// TestTask111_UpdateTestCase_SnapshotTx — A2W2-2-002/010.
func TestTask111_UpdateTestCase_SnapshotTx(t *testing.T) {
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)
	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "snap", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	title := "snap-v2"
	updated, err := svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{
		Title:   &title,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "snap-v2", updated.Title)

	versions, err := svc.ListTestCaseVersions(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	require.NotEmpty(t, versions)
}

// TestTask111_SoftDeleteBase_Cascade — A2W2-5-001.
func TestTask111_SoftDeleteBase_Cascade(t *testing.T) {
	dsn := "file:bases_cascade_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("c-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	repo := bases.NewEntRepository(client)
	svc := bases.NewService(repo)

	base, err := svc.CreateBase(ctx, owner.ID, bases.CreateBaseRequest{Name: "B"})
	require.NoError(t, err)

	table, err := svc.CreateTable(ctx, owner.ID, base.ID, bases.CreateTableRequest{Name: "T"})
	require.NoError(t, err)

	_, err = svc.CreateColumn(ctx, owner.ID, base.ID, table.ID, bases.CreateColumnRequest{
		Name: "Extra", ColType: bases.ColumnText,
	})
	require.NoError(t, err)

	_, err = svc.CreateRecord(ctx, base.ID, table.ID, owner.ID, bases.CreateRecordRequest{
		Cells: map[string]interface{}{},
	})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteBase(ctx, owner.ID, base.ID))

	aliveTables, err := client.HubTable.Query().Where(hubtable.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveTables)
	aliveCols, err := client.HubColumn.Query().Where(hubcolumn.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveCols)
	aliveRecs, err := client.HubRecord.Query().Where(hubrecord.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, aliveRecs)
}

// TestTask111_EventSubjectCanon — A2W2-2-007.
func TestTask111_EventSubjectCanon(t *testing.T) {
	assert.Equal(t, "tms.testrun.completed", tms.EventTestRunCompleted)
}

// TestTask111_ExportTruncateHelper — A2W2-4-002 rune-safe helper.
func TestTask111_ExportTruncateHelper(t *testing.T) {
	got := tms.ExportTruncate("абвгдеёжзий", 3)
	assert.Equal(t, "абв...", got)
}

// errMatcherRepo — Find* возвращает infra error (не NotFound).
type errMatcherRepo struct {
	tms.Repository
	err error
}

func (r *errMatcherRepo) FindTestCaseByExternalID(context.Context, uuid.UUID, string) (*ent.TestCase, error) {
	return nil, r.err
}
func (r *errMatcherRepo) FindTestCaseByAutomationID(context.Context, uuid.UUID, string) (*ent.TestCase, error) {
	return nil, r.err
}
func (r *errMatcherRepo) FindTestCaseByKey(context.Context, uuid.UUID, string) (*ent.TestCase, error) {
	return nil, r.err
}
func (r *errMatcherRepo) FindTestCaseByFullName(context.Context, uuid.UUID, string) (*ent.TestCase, error) {
	return nil, r.err
}
func (r *errMatcherRepo) FindTestCaseByClassMethod(context.Context, uuid.UUID, string, string) (*ent.TestCase, error) {
	return nil, r.err
}
