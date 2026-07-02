package bases_test
import (
	"github.com/valerii/hub/backend/internal/bases"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateTable_InvalidName(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Tables"})
	require.NoError(t, err)

	_, err = svc.CreateTable(ctx, ownerID, base.ID, CreateTableRequest{Name: "  "})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestService_CreateTable_BaseNotFound(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	_, err := svc.CreateTable(context.Background(), ownerID, uuid.New(), CreateTableRequest{Name: "X"})
	assert.ErrorIs(t, err, bases.ErrNotFound)
}

func TestService_CreateColumn_InvalidName(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Cols"})
	require.NoError(t, err)
	tableID := base.Tables[0].ID

	_, err = svc.CreateColumn(ctx, ownerID, base.ID, tableID, CreateColumnRequest{Name: ""})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestService_UpdateRecord_NotFound(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Rec"})
	require.NoError(t, err)
	tableID := base.Tables[0].ID

	_, err = svc.UpdateRecord(ctx, ownerID, base.ID, tableID, uuid.New(), UpdateRecordRequest{
		Cells: map[string]interface{}{},
	})
	assert.ErrorIs(t, err, bases.ErrNotFound)
}

func TestService_DeleteRecord_NotFound(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "Del"})
	require.NoError(t, err)
	tableID := base.Tables[0].ID

	err = svc.DeleteRecord(ctx, ownerID, base.ID, tableID, uuid.New())
	assert.ErrorIs(t, err, bases.ErrNotFound)
}

func TestService_GetTableView_NotFound(t *testing.T) {
	svc, ownerID := setupBasesService(t)
	ctx := context.Background()

	base, err := svc.CreateBase(ctx, ownerID, CreateBaseRequest{Name: "View"})
	require.NoError(t, err)

	_, err = svc.GetTableView(ctx, ownerID, base.ID, uuid.New())
	assert.ErrorIs(t, err, bases.ErrNotFound)
}
