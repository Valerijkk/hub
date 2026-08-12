//go:build hubtest

package bases_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/hubcolumn"
)

// TestCreateTable_DefaultNameColumnInTx_Task062 — W2-A5-007 / P3-BE-07:
// CreateTable атомарно создаёт колонку Name в той же tx.
// Мутация: CreateColumn вне tx / без Name → GetTableMeta без Name → assert краснеет.
func TestCreateTable_DefaultNameColumnInTx_Task062(t *testing.T) {
	repo, ctx, ownerID := setupBasesRepo(t)
	base, err := repo.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "TX Base"})
	require.NoError(t, err)

	table, err := repo.CreateTable(ctx, base.ID, CreateTableRequest{Name: "Atomic"})
	require.NoError(t, err)

	meta, err := repo.GetTableMeta(ctx, base.ID, table.ID)
	require.NoError(t, err)
	require.Len(t, meta.Edges.Columns, 1)
	assert.Equal(t, "Name", meta.Edges.Columns[0].Name)
	assert.Equal(t, hubcolumn.ColTypeText, meta.Edges.Columns[0].ColType)
}
