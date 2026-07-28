package crm_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/crm"
)

// TestCreateDeal_RejectsStrangerContact_W7A2005 — member знает UUID чужого контакта,
// но не может привязать его к своей сделке (owner-scope в validateDealRefs).
// Mutation: вернуть GetContact без authorizeContactWrite → SUCCESS → FAIL.
func TestCreateDeal_RejectsStrangerContact_W7A2005(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_w7a2005_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
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
	contact, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "Secret lead"})
	require.NoError(t, err)

	_, err = svc.CreateDeal(ctx, stranger.ID, crm.CreateDealRequest{
		Title:     "Stolen ref",
		Stage:     crm.StageLead,
		ContactID: &contact.ID,
	})
	require.ErrorIs(t, err, crm.ErrForbidden)

	// owner — ok
	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title:     "Own deal",
		Stage:     crm.StageLead,
		ContactID: &contact.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, deal.ContactID)
	require.Equal(t, contact.ID, *deal.ContactID)
}

// TestCreateDeal_AdminMayLinkForeignContact_W7A2005 — admin/founder обходят owner-scope.
func TestCreateDeal_AdminMayLinkForeignContact_W7A2005(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_w7a2005adm_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
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

	deal, err := svc.CreateDeal(ctx, admin.ID, crm.CreateDealRequest{
		Title:     "Admin deal",
		Stage:     crm.StageLead,
		ContactID: &contact.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, deal.ContactID)
	require.Equal(t, contact.ID, *deal.ContactID)
}
