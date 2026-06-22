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

type mockInvoiceCreator struct {
	called     bool
	dealID     uuid.UUID
	amount     int64
	customer   string
}

func (m *mockInvoiceCreator) CreateInvoiceFromWonDeal(
	_ context.Context,
	_ uuid.UUID,
	dealID uuid.UUID,
	customerName, _ string,
	amountCents int64,
	_ string,
) error {
	m.called = true
	m.dealID = dealID
	m.amount = amountCents
	m.customer = customerName
	return nil
}

func TestService_UpdateDealWonTriggersInvoice(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_inv?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("sales2@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Beta").
		SetKey("BETA").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	mock := &mockInvoiceCreator{}
	svc := crm.NewService(crm.NewEntRepository(client))
	svc.SetDealInvoiceCreator(mock)

	amount := int64(500000)
	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title:       "Enterprise license",
		Stage:       crm.StageProposal,
		AmountCents: &amount,
		ProjectID:   &proj.ID,
	})
	require.NoError(t, err)
	assert.False(t, mock.called)

	won := crm.StageWon
	updated, err := svc.UpdateDeal(ctx, owner.ID, deal.ID, crm.UpdateDealRequest{
		Stage:   &won,
		Version: deal.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, crm.StageWon, updated.Stage)
	assert.True(t, mock.called)
	assert.Equal(t, deal.ID, mock.dealID)
	assert.Equal(t, amount, mock.amount)
}

func TestService_UpdateDealWonSkipsZeroAmount(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:crm_inv2?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("sales3@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	mock := &mockInvoiceCreator{}
	svc := crm.NewService(crm.NewEntRepository(client))
	svc.SetDealInvoiceCreator(mock)

	deal, err := svc.CreateDeal(ctx, owner.ID, crm.CreateDealRequest{
		Title: "No amount deal",
		Stage: crm.StageLead,
	})
	require.NoError(t, err)

	won := crm.StageWon
	_, err = svc.UpdateDeal(ctx, owner.ID, deal.ID, crm.UpdateDealRequest{
		Stage:   &won,
		Version: deal.Version,
	})
	require.NoError(t, err)
	assert.False(t, mock.called)
}
