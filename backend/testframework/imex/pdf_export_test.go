package imex_test
import (
	"github.com/valerii/hub/backend/internal/imex"
	"context"
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

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func seedImexPDFData(t *testing.T) (*imex.ImExService, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	svc, client, ctx, projectID, userID := newImexTestEnv(t)

	testCase, err := client.TestCase.Create().
		SetCaseNumber(1).
		SetCaseKey("IMP-TC-1").
		SetTitle("Login flow").
		SetProjectID(projectID).
		SetCreatedByID(userID).
		Save(ctx)
	require.NoError(t, err)

	run, err := client.TestRun.Create().
		SetTitle("Regression run").
		SetStatus(testrun.StatusCompleted).
		SetProjectID(projectID).
		SetCreatedByID(userID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.TestResult.Create().
		SetRunID(run.ID).
		SetTestCaseID(testCase.ID).
		SetStatus(testresult.StatusPassed).
		SetComment("all good").
		Save(ctx)
	require.NoError(t, err)

	done, err := client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	start := time.Now().UTC()
	sprint, err := client.Sprint.Create().
		SetName("Sprint PDF").
		SetGoal("coverage").
		SetStatus("active").
		SetStartDate(start).
		SetEndDate(start.AddDate(0, 0, 14)).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("IMP-1").
		SetTitle("Done task").
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetAssigneeID(userID).
		SetStatusID(done.ID).
		SetSprintID(sprint.ID).
		Save(ctx)
	require.NoError(t, err)

	return svc, projectID, run.ID, sprint.ID
}

func TestExportRunPDF_GeneratesBytes(t *testing.T) {
	svc, projectID, runID, _ := seedImexPDFData(t)
	ctx := contextBackground(t)

	data, err := svc.ExportRunPDF(ctx, projectID, runID)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data[:20]), "%PDF")
}

func TestExportRunPDF_NotFound(t *testing.T) {
	svc, _, _, _ := seedImexPDFData(t)
	ctx := contextBackground(t)

	_, err := svc.ExportRunPDF(ctx, uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestExportSprintPDF_GeneratesBytes(t *testing.T) {
	svc, projectID, _, sprintID := seedImexPDFData(t)
	ctx := contextBackground(t)

	data, err := svc.ExportSprintPDF(ctx, projectID, sprintID)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestExportSprintPDF_NotFound(t *testing.T) {
	svc, projectID, _, _ := seedImexPDFData(t)
	ctx := contextBackground(t)

	_, err := svc.ExportSprintPDF(ctx, projectID, uuid.New())
	require.Error(t, err)
}

func TestPDFHelpers_BuildRows(t *testing.T) {
	assert.NotNil(t, imex.ExportBuildHeaderRow("Title"))
	assert.NotNil(t, imex.ExportBuildInfoRow("Label", "Value"))
	assert.NotNil(t, imex.ExportBuildTableHeader([]string{"A", "B", "C"}))
	long := strings.Repeat("x", 55)
	assert.NotNil(t, imex.ExportBuildTableRow([]string{long, "ok", "x"}))
}

func setupImexPDFApp(h *imex.Handler, userID uuid.UUID) *fiber.App {
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app
}

func TestHandler_ExportRunAndSprintPDF(t *testing.T) {
	svc, projectID, runID, sprintID := seedImexPDFData(t)
	h := imex.NewHandler(svc)
	app := setupImexPDFApp(h, uuid.New())

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/test-runs/"+runID.String()+"/export/pdf", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "pdf")

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/sprints/"+sprintID.String()+"/export/pdf", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func contextBackground(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}
