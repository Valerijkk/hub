//go:build hubtest

package accounting_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/invoice"
	"github.com/valerii/hub/backend/ent/user"
)

// TestJournalEntry_InvoiceO2O_TASK025 — вторая Invoice на тот же JournalEntry → constraint.
// Mutation: убрать .Unique() у JournalEntry.To("invoice") и index.Edges("journal_entry").Unique()
// → второй Create проходит → FAIL.
func TestJournalEntry_InvoiceO2O_TASK025(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:journal_inv_o2o_t025?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-o2o@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	entry, err := client.JournalEntry.Create().
		SetEntryDate(time.Now().UTC()).
		SetDescription("payment posting").
		SetCreatedByID(admin.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Invoice.Create().
		SetNumber("INV-O2O-1").
		SetCustomerName("Acme").
		SetAmountCents(1000).
		SetStatus(invoice.StatusIssued).
		SetCreatedByID(admin.ID).
		SetJournalEntryID(entry.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Invoice.Create().
		SetNumber("INV-O2O-2").
		SetCustomerName("Acme").
		SetAmountCents(2000).
		SetStatus(invoice.StatusIssued).
		SetCreatedByID(admin.ID).
		SetJournalEntryID(entry.ID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err), "second invoice on same journal must fail: %v", err)
}
