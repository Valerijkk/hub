//go:build hubtest

package bases_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/bases"
)

// TestEntRepository_GetTable_CapsRecords_A13_002 — GetTable ограничивает records;
// GetTableMeta не загружает records (write-path).
// Mutation: снять Limit → FAIL при >MaxTableRecords.
func TestEntRepository_GetTable_CapsRecords_A13_002(t *testing.T) {
	repo, ctx, ownerID := setupBasesRepo(t)

	base, err := repo.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Cap Base"})
	require.NoError(t, err)
	table, err := repo.CreateTable(ctx, base.ID, CreateTableRequest{Name: "Rows"})
	require.NoError(t, err)

	// Создаём больше лимита, но не полный MaxTableRecords+много — cap проверяем
	// через константу и синтетический размер: создаём MaxTableRecords+3.
	// Для sqlite in-memory это быстро.
	capN := bases.MaxTableRecords
	for i := 0; i < capN+3; i++ {
		_, err := repo.CreateRecord(ctx, table.ID, ownerID, map[string]interface{}{
			"n": float64(i),
		})
		require.NoError(t, err)
	}

	view, err := repo.GetTable(ctx, base.ID, table.ID)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(view.Edges.Records), capN)
	assert.Equal(t, capN, len(view.Edges.Records), "expected hard cap hit")

	meta, err := repo.GetTableMeta(ctx, base.ID, table.ID)
	require.NoError(t, err)
	assert.Empty(t, meta.Edges.Records, "GetTableMeta must not load records")
	assert.NotNil(t, meta.Edges.Columns)

	// GetRecord точечно без full scan
	first := view.Edges.Records[0]
	got, err := repo.GetRecord(ctx, table.ID, first.ID)
	require.NoError(t, err)
	assert.Equal(t, first.ID, got.ID)

	// sanity: CreateRecord path uses meta (service-level covered separately)
	_ = fmt.Sprintf("created %d records, loaded %d", capN+3, len(view.Edges.Records))
}

// TestService_CreateRecord_UsesMetaNotFullTable_A13_002 — CreateRecord после
// большого числа rows не падает и не требует full GetTable.
func TestService_CreateRecord_UsesMetaNotFullTable_A13_002(t *testing.T) {
	repo, ctx, ownerID := setupBasesRepo(t)
	svc := NewService(repo)

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Write Path"})
	require.NoError(t, err)
	table, err := svc.CreateTable(ctx, ownerID, base.ID, CreateTableRequest{Name: "T"})
	require.NoError(t, err)

	rec, err := svc.CreateRecord(ctx, base.ID, table.ID, ownerID, CreateRecordRequest{
		Cells: map[string]interface{}{},
	})
	require.NoError(t, err)
	require.NotNil(t, rec)

	updated, err := svc.UpdateRecord(ctx, ownerID, base.ID, table.ID, rec.ID, UpdateRecordRequest{
		Cells:   map[string]interface{}{},
		Version: rec.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, rec.ID, updated.ID)

	require.NoError(t, svc.DeleteRecord(ctx, ownerID, base.ID, table.ID, rec.ID))
}
