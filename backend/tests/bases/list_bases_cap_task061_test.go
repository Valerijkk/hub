//go:build hubtest

package bases_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/bases"
)

// TestEntRepository_ListBases_Caps_P3BE06 — N>MaxOwnerBases → Limit.
// Mutation: снять Limit → FAIL (len > MaxOwnerBases).
func TestEntRepository_ListBases_Caps_P3BE06(t *testing.T) {
	repo, ctx, ownerID := setupBasesRepo(t)

	capN := bases.MaxOwnerBases
	for i := 0; i < capN+3; i++ {
		_, err := repo.CreateBase(ctx, ownerID, CreateBaseRequest{Name: fmt.Sprintf("Base %d", i)})
		require.NoError(t, err)
	}

	list, err := repo.ListBases(ctx, ownerID)
	require.NoError(t, err)
	assert.Equal(t, capN, len(list))
}
