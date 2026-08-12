package crm_test
import (
	"github.com/valerii/hub/backend/internal/crm"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
)

func TestService_DeleteContact(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_del_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("del-" + uuid.New().String()[:8] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	repo := crm.NewEntRepository(client)
	svc := crm.NewService(repo)

	created, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "To Delete", Email: "del@test.com"})
	require.NoError(t, err)

	list, err := svc.ListContacts(ctx, owner.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, svc.DeleteContact(ctx, owner.ID, created.ID))
	list, err = svc.ListContacts(ctx, owner.ID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestService_UpdateContact_NotFound(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_upd_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	svc := crm.NewService(crm.NewEntRepository(client))
	name := "X"
	_, err := svc.UpdateContact(context.Background(), uuid.New(), uuid.New(), crm.UpdateContactRequest{Name: &name})
	require.ErrorIs(t, err, crm.ErrNotFound)
}

func TestService_CreateDeal_InvalidProject(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_deal_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().SetEmail("deal@test.com").Save(ctx)
	require.NoError(t, err)
	repo := crm.NewEntRepository(client)
	svc := crm.NewService(repo)

	contact, err := svc.CreateContact(ctx, owner.ID, crm.CreateContactRequest{Name: "Buyer", Email: "b@test.com"})
	require.NoError(t, err)

	contactID := contact.ID
	_, err = svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title:       "Deal",
		Stage:       crm.StageLead,
		AmountCents: ptrInt64(100),
		ContactID:   &contactID,
		ProjectID:   ptrUUID(uuid.New()),
	})
	require.Error(t, err)

	proj, err := client.Project.Create().SetName("CRM").SetKey("CRM").SetStatus(project.StatusActive).Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetProjectID(proj.ID).SetUserID(owner.ID).Save(ctx)
	require.NoError(t, err)

	projID := proj.ID
	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title:       "Valid deal",
		Stage:       crm.StageLead,
		AmountCents: ptrInt64(500),
		ContactID:   &contactID,
		ProjectID:   &projID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Valid deal", deal.Title)
}

func ptrInt64(v int64) *int64 { return &v }

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
