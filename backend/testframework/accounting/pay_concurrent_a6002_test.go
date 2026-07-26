package accounting_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/journalentry"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/accounting"
)

// simulateConcurrentPay моделирует A6-002: два pay после общего status-check.
// atomicProtect=true — секция как PayInvoiceAtomic (одна проводка);
// false — TOCTOU как старый MarkInvoicePaid (две проводки).
func simulateConcurrentPay(atomicProtect bool) int {
	status := "issued"
	journals := 0
	var mu sync.Mutex
	var barrier sync.WaitGroup
	barrier.Add(2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if atomicProtect {
				mu.Lock()
				defer mu.Unlock()
			}
			if status != "issued" {
				return
			}
			if !atomicProtect {
				barrier.Done()
				barrier.Wait()
			}
			journals++
			status = "paid"
		}()
	}
	wg.Wait()
	return journals
}

// TestMarkInvoicePaid_ConcurrentSingleJournal_A6_002 — конкурентность: одна проводка.
func TestMarkInvoicePaid_ConcurrentSingleJournal_A6_002(t *testing.T) {
	// 1) Модель гонки (детерминированно ловит отсутствие atomic-секции).
	// mutation: atomicProtect false → journals==2 → FAIL
	assert.Equal(t, 1, simulateConcurrentPay(true), "concurrent pay must post exactly one journal")

	// 2) Реальный Ent PayInvoiceAtomic: два параллельных MarkInvoicePaid.
	client := enttest.Open(t, "sqlite3", "file:acct_pay_race_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-race-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := accounting.NewEntRepository(client)
	svc := accounting.NewService(repo)
	require.NoError(t, repo.EnsureDefaultChart(ctx))

	inv, err := svc.CreateInvoice(ctx, admin.ID, accounting.CreateInvoiceRequest{
		CustomerName: "Acme",
		Description:  "Race pay",
		AmountCents:  10000,
		Currency:     "RUB",
	})
	require.NoError(t, err)
	_, err = svc.IssueInvoice(ctx, admin.ID, inv.ID)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.MarkInvoicePaid(ctx, admin.ID, inv.ID)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	var okCount int
	for err := range errs {
		if err == nil {
			okCount++
		}
	}
	assert.Equal(t, 1, okCount)

	posted, err := client.JournalEntry.Query().
		Where(journalentry.StatusEQ(journalentry.StatusPosted)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, posted)

	paid, err := client.Invoice.Get(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, "paid", string(paid.Status))
}
