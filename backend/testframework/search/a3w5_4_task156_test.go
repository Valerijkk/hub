package search_test

// TASK-156 / A3W5-4-001..003: operator Meili dial, JQL fail-closed, IndexAllIssues pagination.
import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/search"
)

// TestNewMeiliIndexerFromEnv_OperatorLoopbackDial_A3W54_001 — mutation: вернуть
// только NewSafeHTTPClient без operator dial → DialContext(127.0.0.1) краснеет.
func TestNewMeiliIndexerFromEnv_OperatorLoopbackDial_A3W54_001(t *testing.T) {
	t.Setenv("MEILI_HOST", "http://127.0.0.1:7700")
	t.Setenv("MEILI_MASTER_KEY", "x")
	idx := search.NewMeiliIndexerFromEnv()
	require.NotNil(t, idx, "operator loopback host должен создать indexer")

	client := search.ExportMeiliIndexerClient(idx)
	require.NotNil(t, client)
	rt, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, rt.DialContext)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Dial к loopback не должен дать ErrBlockedAddress (connection refused — OK).
	_, err := rt.DialContext(ctx, "tcp", "127.0.0.1:1")
	if err != nil {
		assert.False(t, errors.Is(err, safehttp.ErrBlockedAddress),
			"operator Meili dial must allow loopback, got %v", err)
	}

	// RFC1918 IP-литерал после operator allow — тоже OK на dial allowlist.
	t.Setenv("MEILI_HOST", "http://10.0.0.5:7700")
	idxRFC := search.NewMeiliIndexerFromEnv()
	require.NotNil(t, idxRFC)
	rtRFC := search.ExportMeiliIndexerClient(idxRFC).Transport.(*http.Transport)
	_, err = rtRFC.DialContext(ctx, "tcp", "10.0.0.5:1")
	if err != nil {
		assert.False(t, errors.Is(err, safehttp.ErrBlockedAddress),
			"operator Meili dial must allow RFC1918, got %v", err)
	}
	// err==nil (unreachable host всё же ответил) тоже доказывает, что dial не blocked.

	// Metadata по-прежнему FAIL / nil indexer.
	t.Setenv("MEILI_HOST", "http://169.254.169.254/")
	assert.Nil(t, search.NewMeiliIndexerFromEnv())

	// Публичный host остаётся на SafeDial (metadata dial blocked).
	t.Setenv("MEILI_HOST", "https://meili.example.com")
	idxPub := search.NewMeiliIndexerFromEnv()
	require.NotNil(t, idxPub)
	rtPub := search.ExportMeiliIndexerClient(idxPub).Transport.(*http.Transport)
	_, err = rtPub.DialContext(ctx, "tcp", "169.254.169.254:80")
	require.ErrorIs(t, err, safehttp.ErrBlockedAddress)

	// Operator dial не открывает metadata даже на operator client.
	dialOp := safehttp.SafeDialContextAllowOperatorPrivate()
	_, err = dialOp(ctx, "tcp", "169.254.169.254:80")
	require.ErrorIs(t, err, safehttp.ErrBlockedAddress)
}

// TestParseJQL_UnsupportedComparison_A3W54_002 — mutation: снова return nil
// в conditionToPredicate для > на priority → silent-nil / unfiltered → RED.
func TestParseJQL_UnsupportedComparison_A3W54_002(t *testing.T) {
	cases := []string{
		`priority > medium`,
		`priority < high`,
		`priority >= high`,
		`priority <= low`,
		`type > bug`,
		`status >= done`,
	}
	for _, jql := range cases {
		_, err := search.ParseJQL(jql)
		require.Error(t, err, "jql=%s", jql)
		assert.ErrorIs(t, err, search.ErrUnsupportedJQLOperator, "jql=%s", jql)
	}

	// Date comparison по-прежнему OK.
	q, err := search.ParseJQL(`created > 2024-01-01`)
	require.NoError(t, err)
	preds := search.ToEntPredicates(q, uuid.New())
	require.Len(t, preds, 1)

	// Bypass Parse: hand-built condition → fail-closed never-match, не nil preds.
	hand := &search.JQLQuery{
		Conditions: []search.JQLCondition{{
			Field: "priority", Operator: ">", Value: "medium",
		}},
	}
	predsHand := search.ToEntPredicates(hand, uuid.New())
	require.NotEmpty(t, predsHand,
		"silent nil predicates = unfiltered list; mutation return-nil in conditionToPredicate")
}

// TestIndexAllIssues_PaginatesBeyondBatch_A3W54_003 — mutation: голый Limit(5000)
// без batch → при batchSize=2 и 3 issues count < 3 (RED).
func TestIndexAllIssues_PaginatesBeyondBatch_A3W54_003(t *testing.T) {
	prev := search.ExportSetMeiliIndexBatchSize(2)
	t.Cleanup(func() { search.ExportSetMeiliIndexBatchSize(prev) })

	var posts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts.Add(1)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)

	idx := search.ExportNewMeiliIndexerTest(srv.URL, "", srv.Client())

	dsn := "file:meili_page_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Page Proj").SetKey("PAG").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("page-" + uuid.New().String()[:8] + "@test.com").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().
		SetName("Open").
		SetCategory("todo").
		SetProjectID(proj.ID).
		SetIsDefault(true).
		Save(ctx)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		_, err = client.Issue.Create().
			SetTitle("Issue").
			SetIssueKey(fmt.Sprintf("PAG-%d", i)).
			SetIssueNumber(i).
			SetProjectID(proj.ID).
			SetStatusID(status.ID).
			SetReporterID(user.ID).
			SetIssueType(issue.IssueTypeTask).
			SetPriority(issue.PriorityMedium).
			Save(ctx)
		require.NoError(t, err)
	}

	count, err := idx.IndexAllIssues(ctx, client)
	require.NoError(t, err)
	assert.Equal(t, 3, count, "pagination must index all issues beyond one batch")
	assert.GreaterOrEqual(t, posts.Load(), int32(2), "ожидается ≥2 upsert-страницы при batch=2")
}

func TestSafeDialContextAllowOperatorPrivate_BlocksPublic_A3W54_001(t *testing.T) {
	dial := safehttp.SafeDialContextAllowOperatorPrivate()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Публичный IP (TEST-NET / example) — не operator private.
	_, err := dial(ctx, "tcp", net.JoinHostPort("8.8.8.8", "53"))
	require.Error(t, err)
	assert.ErrorIs(t, err, safehttp.ErrBlockedAddress)
}
