package imex

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func newImexTestEnv(t *testing.T) (*ImExService, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := "file:imextest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("imex-" + uuid.New().String()[:8] + "@test.com").
		SetName("Importer").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Import Project").
		SetKey("IMP").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	return NewImExService(client), ctx, proj.ID, user.ID
}

func TestImportIssuesCSV_Preview(t *testing.T) {
	svc, ctx, projectID, _ := newImexTestEnv(t)
	csvData := []byte("Title,Type,Priority\nFix login,bug,high\nAdd API,task,medium\n")

	preview, result, err := svc.ImportIssuesCSV(ctx, projectID, csvData, nil, true)
	require.NoError(t, err)
	require.Nil(t, result)
	require.NotNil(t, preview)
	assert.Equal(t, 2, preview.Total)
	require.Len(t, preview.Rows, 2)
	assert.Equal(t, "Fix login", preview.Rows[0].Fields["Title"])
}

func TestImportIssuesCSV_ImportsRows(t *testing.T) {
	svc, ctx, projectID, _ := newImexTestEnv(t)
	csvData := []byte("Summary,Type\nFirst task,task\nSecond task,bug\n")

	_, result, err := svc.ImportIssuesCSV(ctx, projectID, csvData, []FieldMapping{
		{SourceColumn: "Summary", TargetField: "title"},
		{SourceColumn: "Type", TargetField: "type"},
	}, false)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Imported)
	assert.Zero(t, result.Skipped)

	count, err := svc.db.Issue.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestImportIssuesCSV_SkipsMissingTitle(t *testing.T) {
	svc, ctx, projectID, _ := newImexTestEnv(t)
	csvData := []byte("Summary,Type\n,task\n")

	_, result, err := svc.ImportIssuesCSV(ctx, projectID, csvData, []FieldMapping{
		{SourceColumn: "Summary", TargetField: "title"},
	}, false)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Equal(t, 1, result.Skipped)
}

func TestImportIssuesJSON_Success(t *testing.T) {
	svc, ctx, projectID, _ := newImexTestEnv(t)
	data := []byte(`[
		{"title":"JSON task","type":"task","priority":"low"},
		{"title":"JSON bug","type":"bug","priority":"critical"}
	]`)

	result, err := svc.ImportIssuesJSON(ctx, projectID, data)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Imported)
}

func TestImportIssuesJSON_InvalidJSON(t *testing.T) {
	svc, ctx, projectID, _ := newImexTestEnv(t)
	_, err := svc.ImportIssuesJSON(ctx, projectID, []byte("{"))
	require.Error(t, err)
}

func TestExportIssuesCSV_IncludesIssues(t *testing.T) {
	svc, ctx, projectID, _ := newImexTestEnv(t)

	status, err := svc.db.WorkflowStatus.Query().First(ctx)
	require.NoError(t, err)

	_, err = svc.db.Issue.Create().
		SetTitle("Export me").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		SetIssueNumber(1).
		SetIssueKey("IMP-1").
		SetProjectID(projectID).
		SetStatusID(status.ID).
		Save(ctx)
	require.NoError(t, err)

	out, err := svc.ExportIssuesCSV(ctx, projectID)
	require.NoError(t, err)

	rows, err := csv.NewReader(bytes.NewReader(out)).ReadAll()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rows), 2)
	assert.Equal(t, "IMP-1", rows[1][0])
	assert.Equal(t, "Export me", rows[1][1])
}

func TestExportProjectZIP_Structure(t *testing.T) {
	svc, ctx, projectID, _ := newImexTestEnv(t)

	out, err := svc.ExportProjectZIP(ctx, projectID)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	require.NoError(t, err)

	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	assert.Contains(t, names, "issues.json")
	assert.Contains(t, names, "metadata.json")

	metaFile, err := zr.Open("metadata.json")
	require.NoError(t, err)
	defer metaFile.Close()

	var meta map[string]interface{}
	require.NoError(t, json.NewDecoder(metaFile).Decode(&meta))
	assert.Equal(t, projectID.String(), meta["project_id"])
}

func TestImportJiraJSON_SingleIssue(t *testing.T) {
	svc, ctx, projectID, userID := newImexTestEnv(t)
	data := []byte(`[{"key":"JIRA-1","fields":{"summary":"Imported from Jira","issuetype":{"name":"Bug"},"priority":{"name":"High"}}}]`)

	result, err := svc.ImportJiraJSON(ctx, projectID, userID, data)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
}
