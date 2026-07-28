//go:build hubtest

package marketplace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/marketplace"
)

// TestEntRepository_SoftDeleteListing_RequiresVersion_P3BE05 — SoftDelete без VersionEQ
// игнорировал optimistic lock. Mutation: снять VersionEQ → stale version delete succeeds → FAIL.
func TestEntRepository_SoftDeleteListing_RequiresVersion_P3BE05(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:mp_softdel_v?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()
	repo := marketplace.NewEntRepository(client)

	author, err := client.User.Create().
		SetEmail("mp-softdel@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	listing, err := repo.CreateListing(ctx, author.ID, marketplace.CreateListingRequest{
		Title: "Versioned delete", Description: "d", Category: marketplace.CategoryEquipment,
	})
	require.NoError(t, err)

	err = repo.SoftDeleteListing(ctx, listing.ID, listing.Version+99)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	still, err := repo.GetListing(ctx, listing.ID)
	require.NoError(t, err)
	assert.Nil(t, still.DeletedAt)

	require.NoError(t, repo.SoftDeleteListing(ctx, listing.ID, listing.Version))
	_, err = repo.GetListing(ctx, listing.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	err = repo.SoftDeleteListing(ctx, listing.ID, listing.Version)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}
