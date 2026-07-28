//go:build hubtest

package bases_test

// TASK-059 / P3-BE-03 / W2-A5-009: cross-base tableId must 404 via HasBaseWith.
// Mutation: drop hubtable.HasBaseWith(baseID) in GetTable/GetTableMeta → PASS→FAIL.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetTableView_CrossBaseTable_NotFound — owner baseA + tableId из baseB → 404.
func TestGetTableView_CrossBaseTable_NotFound(t *testing.T) {
	ownerApp, _, svc, ownerID, _ := setupBasesRBAC(t)
	ctx := context.Background()

	baseA := createOwnerBase(t, svc, ownerID, "Base A")
	baseB := createOwnerBase(t, svc, ownerID, "Base B")
	tableFromB := baseB.Tables[0].ID

	// Service-layer: authorizeBase(baseA) OK, GetTable(baseA, tableB) → NotFound.
	_, err := svc.GetTableView(ctx, ownerID, baseA.ID, tableFromB)
	require.ErrorIs(t, err, ErrNotFound)

	// HTTP: same path → 404 (не 200 с чужой таблицей).
	path := "/api/bases/" + baseA.ID.String() + "/tables/" + tableFromB.String()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	resp, err := ownerApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
