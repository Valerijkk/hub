package crm_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/crm"
)

// TestUpdateContact_OptimisticLock_A6_003 — stale/zero version → conflict; VersionEQ в Ent WHERE.
// Mutation: снять crmcontact.VersionEQ в UpdateContact → stale write проходит → FAIL.
func TestUpdateContact_OptimisticLock_A6_003(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_a6003_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	created, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "Acme"})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	name := "Acme v2"
	updated, err := svc.UpdateContact(ctx, owner.ID, created.ID, crm.UpdateContactRequest{
		Name: &name, Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Acme v2", updated.Name)
	assert.Equal(t, 2, updated.Version)

	stale := "should-not-apply"
	_, err = svc.UpdateContact(ctx, owner.ID, created.ID, crm.UpdateContactRequest{
		Name: &stale, Version: created.Version, // stale: still 1
	})
	require.ErrorIs(t, err, crm.ErrVersionConflict)

	_, err = svc.UpdateContact(ctx, owner.ID, created.ID, crm.UpdateContactRequest{
		Name: &stale, Version: 0,
	})
	require.ErrorIs(t, err, crm.ErrVersionConflict)

	got, err := client.CrmContact.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Acme v2", got.Name)
	assert.Equal(t, 2, got.Version)
}

// TestUpdateDeal_OptimisticLock_A6_003 — stale version → conflict на deal Update.
func TestUpdateDeal_OptimisticLock_A6_003(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_a6003d_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{Title: "Deal", Stage: crm.StageLead})
	require.NoError(t, err)

	title := "Deal v2"
	updated, err := svc.UpdateDeal(ctx, owner.ID, deal.ID, crm.UpdateDealRequest{
		Title: &title, Version: deal.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "Deal v2", updated.Title)
	assert.Equal(t, 2, updated.Version)

	stale := "nope"
	_, err = svc.UpdateDeal(ctx, owner.ID, deal.ID, crm.UpdateDealRequest{
		Title: &stale, Version: deal.Version,
	})
	require.ErrorIs(t, err, crm.ErrVersionConflict)

	got, err := client.CrmDeal.Get(ctx, deal.ID)
	require.NoError(t, err)
	assert.Equal(t, "Deal v2", got.Title)
	assert.Equal(t, 2, got.Version)
}
