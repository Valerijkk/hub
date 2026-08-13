package tracker_test

// TASK-104: HTTP-контракт list-issues — cursor pagination (NFR-14).
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestHandler_ListIssues_CursorPagination_Task104 проверяет wire-контракт:
// items/next_cursor/has_more/total; вторая страница по cursor без пересечения ключей.
// Mutation: ответ с offset/page вместо next_cursor → decode/assert FAIL; игнор cursor → пересечение ключей.
func TestHandler_ListIssues_CursorPagination_Task104(t *testing.T) {
	app, projectID, _, _ := setupTrackerHandlerApp(t)

	for i := 0; i < 5; i++ {
		createIssueViaHandler(t, app, projectID, fmt.Sprintf("Cursor HTTP %d", i+1))
	}

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/issues?limit=2", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var page1 tracker.PaginatedResponse[tracker.IssueResponse]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page1))
	require.Len(t, page1.Items, 2)
	assert.True(t, page1.HasMore)
	assert.Equal(t, 5, page1.Total)
	require.NotEmpty(t, page1.NextCursor)
	assert.Equal(t, page1.Items[len(page1.Items)-1].IssueKey, page1.NextCursor)

	req2 := httptest.NewRequest(
		http.MethodGet,
		"/api/projects/"+projectID.String()+"/issues?limit=2&cursor="+page1.NextCursor,
		nil,
	)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	var page2 tracker.PaginatedResponse[tracker.IssueResponse]
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&page2))
	require.Len(t, page2.Items, 2)

	seen := map[string]struct{}{}
	for _, item := range page1.Items {
		seen[item.IssueKey] = struct{}{}
	}
	for _, item := range page2.Items {
		assert.NotContains(t, seen, item.IssueKey)
	}
}
