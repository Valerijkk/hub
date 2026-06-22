package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
)

func TestNewMeiliIndexerFromEnv_Unset(t *testing.T) {
	t.Setenv("MEILI_HOST", "")
	assert.Nil(t, search.NewMeiliIndexerFromEnv())
}

func TestNewMeiliIndexerFromEnv_Configured(t *testing.T) {
	t.Setenv("MEILI_HOST", "http://127.0.0.1:7700")
	t.Setenv("MEILI_MASTER_KEY", "test-key")
	idx := search.NewMeiliIndexerFromEnv()
	require.NotNil(t, idx)
	assert.Equal(t, "http://127.0.0.1:7700", search.ExportMeiliIndexerHost(idx))
}

func TestMeiliIndexer_IndexAllIssues_NilReceiver(t *testing.T) {
	var idx *search.MeiliIndexer
	_, err := idx.IndexAllIssues(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestMeiliIndexer_IndexAllIssues_UpsertsViaHTTP(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)

	idx := search.ExportNewMeiliIndexerTest(srv.URL, "secret", srv.Client())

	dsn := "file:meili_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Meili Proj").SetKey("MLI").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("meili-" + uuid.New().String()[:8] + "@test.com").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().
		SetName("Open").
		SetCategory("todo").
		SetProjectID(proj.ID).
		SetIsDefault(true).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.Issue.Create().
		SetTitle("Search me").
		SetIssueKey("MLI-1").
		SetIssueNumber(1).
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(user.ID).
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(ctx)
	require.NoError(t, err)

	count, err := idx.IndexAllIssues(ctx, client)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestMeiliIndexer_upsertDocuments_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	idx := search.ExportNewMeiliIndexerTest(srv.URL, "", srv.Client())
	err := search.ExportMeiliUpsertDocuments(idx, context.Background(), search.ExportIssuesIndex(), []search.ExportMeiliIssueDoc{{ID: "1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "meilisearch 400")
}

func TestMeiliIndexer_IndexAllIssues_EntError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)

	idx := search.ExportNewMeiliIndexerTest(srv.URL, "", srv.Client())
	client := enttest.Open(t, "sqlite3", "file:meili_err_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	client.Close()

	_, err := idx.IndexAllIssues(context.Background(), client)
	require.Error(t, err)
}
