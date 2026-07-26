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

// TestUpdateDelete_RejectsNonOwner_A6_001 — IDOR: чужой member не может Update/Delete contact/deal.
func TestUpdateDelete_RejectsNonOwner_A6_001(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_a6001_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	contact, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "Secret", Email: "s@test.com"})
	require.NoError(t, err)
	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{Title: "Secret deal", Stage: crm.StageLead})
	require.NoError(t, err)

	name := "Hacked"
	_, err = svc.UpdateContact(ctx, stranger.ID, contact.ID, crm.UpdateContactRequest{Name: &name, Version: contact.Version})
	require.ErrorIs(t, err, crm.ErrForbidden)

	err = svc.DeleteContact(ctx, stranger.ID, contact.ID)
	require.ErrorIs(t, err, crm.ErrForbidden)

	won := crm.StageWon
	_, err = svc.UpdateDeal(ctx, stranger.ID, deal.ID, crm.UpdateDealRequest{Stage: &won, Version: deal.Version})
	require.ErrorIs(t, err, crm.ErrForbidden)

	err = svc.DeleteDeal(ctx, stranger.ID, deal.ID)
	require.ErrorIs(t, err, crm.ErrForbidden)

	// owner still can
	_, err = svc.UpdateContact(ctx, owner.ID, contact.ID, crm.UpdateContactRequest{Name: &name, Version: contact.Version})
	require.NoError(t, err)
	assert.Equal(t, name, name)
}

// TestUpdateContact_AdminAllowed_A6_001 — admin/founder могут править чужой контакт.
func TestUpdateContact_AdminAllowed_A6_001(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_a6001adm_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("o-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	admin, err := client.User.Create().
		SetEmail("a-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	svc := crm.NewService(crm.NewEntRepository(client))
	contact, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "Lead"})
	require.NoError(t, err)

	name := "Admin edit"
	updated, err := svc.UpdateContact(ctx, admin.ID, contact.ID, crm.UpdateContactRequest{Name: &name, Version: contact.Version})
	require.NoError(t, err)
	assert.Equal(t, name, updated.Name)
}
