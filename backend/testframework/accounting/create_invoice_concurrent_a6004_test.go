package accounting_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/ent/invoice"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/accounting"
)

// TestCreateInvoice_ConcurrentSinglePerDeal_A6_004 — два параллельных create на одну сделку:
// ровно один invoice. Mutation: убрать deal-lock / Exist в CreateInvoice tx → 2 invoices → FAIL.
func TestCreateInvoice_ConcurrentSinglePerDeal_A6_004(t *testing.T) {
	client := openAccountingPostgresEnt(t)
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-deal-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	owner, err := client.User.Create().
		SetEmail("sales-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	deal, err := client.CrmDeal.Create().
		SetTitle("Race deal").
		SetStage(crmdeal.StageWon).
		SetAmountCents(50000).
		SetCurrency("RUB").
		SetOwnerID(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	svc := accounting.NewService(accounting.NewEntRepository(client))
	dealID := deal.ID

	var start sync.WaitGroup
	start.Add(2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Done()
			start.Wait()
			_, err := svc.CreateInvoice(ctx, admin.ID, accounting.CreateInvoiceRequest{
				CustomerName: "Acme",
				Description:  "Race invoice",
				AmountCents:  50000,
				Currency:     "RUB",
				CrmDealID:    &dealID,
			})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	okCount, conflictCount := 0, 0
	for err := range errs {
		if err == nil {
			okCount++
			continue
		}
		if assert.ErrorIs(t, err, accounting.ErrInvoiceExists) {
			conflictCount++
		}
	}
	assert.Equal(t, 1, okCount, "exactly one create must succeed")
	assert.Equal(t, 1, conflictCount, "exactly one create must hit ErrInvoiceExists")

	n, err := client.Invoice.Query().
		Where(
			invoice.HasCrmDealWith(crmdeal.ID(dealID)),
			invoice.DeletedAtIsNil(),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "exactly one invoice per deal")
}
