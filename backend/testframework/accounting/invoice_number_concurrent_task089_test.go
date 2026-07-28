package accounting_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/invoice"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/accounting"
)

// TestCreateInvoice_ConcurrentUniqueNumbers_W7A2007 — два параллельных create без deal:
// оба успеха, номера различны (unique index + MAX+1 retry).
// Mutation: вернуть Count+1 без retry → duplicate / 1 FAIL → FAIL.
func TestCreateInvoice_ConcurrentUniqueNumbers_W7A2007(t *testing.T) {
	client := openAccountingPostgresEnt(t)
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-num-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	svc := accounting.NewService(accounting.NewEntRepository(client))

	var start sync.WaitGroup
	start.Add(2)
	type result struct {
		inv *accounting.InvoiceResponse
		err error
	}
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			start.Done()
			start.Wait()
			inv, err := svc.CreateInvoice(ctx, admin.ID, accounting.CreateInvoiceRequest{
				CustomerName: "Acme",
				Description:  "Race number",
				AmountCents:  int64(10000 + n),
				Currency:     "RUB",
			})
			results <- result{inv: inv, err: err}
		}(i)
	}
	wg.Wait()
	close(results)

	numbers := make(map[string]struct{}, 2)
	for r := range results {
		require.NoError(t, r.err)
		require.NotNil(t, r.inv)
		numbers[r.inv.Number] = struct{}{}
	}
	assert.Len(t, numbers, 2, "concurrent creates must allocate distinct invoice numbers")

	n, err := client.Invoice.Query().
		Where(invoice.DeletedAtIsNil()).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}
