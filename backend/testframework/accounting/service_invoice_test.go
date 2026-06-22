package accounting_test
import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/ent/user"
)

func TestService_CreateInvoiceFromWonDeal(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:acct_deal?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo2@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	owner, err := client.User.Create().
		SetEmail("sales@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	deal, err := client.CrmDeal.Create().
		SetTitle("SaaS subscription").
		SetStage(crmdeal.StageProposal).
		SetAmountCents(990000).
		SetCurrency("RUB").
		SetOwnerID(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo)

	err = svc.CreateInvoiceFromWonDeal(ctx, admin.ID, deal.ID, "Acme", deal.Title, 990000, "RUB")
	require.NoError(t, err)

	invoices, err := svc.ListInvoices(ctx, admin.ID)
	require.NoError(t, err)
	require.Len(t, invoices, 1)
	assert.Equal(t, InvoiceDraft, invoices[0].Status)
	assert.Equal(t, int64(990000), invoices[0].AmountCents)
	require.NotNil(t, invoices[0].CrmDealID)
	assert.Equal(t, deal.ID, *invoices[0].CrmDealID)

	err = svc.CreateInvoiceFromWonDeal(ctx, admin.ID, deal.ID, "Acme", deal.Title, 990000, "RUB")
	require.NoError(t, err)
	invoices, err = svc.ListInvoices(ctx, admin.ID)
	require.NoError(t, err)
	assert.Len(t, invoices, 1)
}
