package intranet_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestEntRepository_KudoSoftDelete_A2W8_3_004 — SoftDeleteKudo; feed hides.
// Mutation: drop DeletedAtIsNil from ListPublicKudos → soft-deleted listed → FAIL.
func TestEntRepository_KudoSoftDelete_A2W8_3_004(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:kudosoft097_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	sender, err := client.User.Create().SetEmail("s-" + uuid.New().String()[:8] + "@t.com").SetName("S").Save(ctx)
	require.NoError(t, err)
	recip, err := client.User.Create().SetEmail("r-" + uuid.New().String()[:8] + "@t.com").SetName("R").Save(ctx)
	require.NoError(t, err)

	repo := intranet.NewEntRepository(client)
	k, err := repo.CreateKudo(ctx, sender.ID, recip.ID, "thanks", true)
	require.NoError(t, err)

	require.NoError(t, repo.SoftDeleteKudo(ctx, k.ID))

	items, total, err := repo.ListPublicKudos(ctx, 1, 25)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, items)

	_, err = repo.GetKudo(ctx, k.ID)
	require.True(t, ent.IsNotFound(err))

	raw, err := client.Kudo.Get(ctx, k.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	src, err := os.ReadFile(filepath.Join(backendRoot, "internal", "intranet", "repository.go"))
	require.NoError(t, err)
	assert.Contains(t, string(src), "SetDeletedAt")
	assert.Contains(t, string(src), "DeletedAtIsNil")
}
