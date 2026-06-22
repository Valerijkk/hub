package accounting_test
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func TestValidateLines_NegativeAmounts(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:acct_neg?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-neg@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo)

	accounts, err := svc.ListAccounts(ctx, admin.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(accounts), 2)

	a, b := accounts[0].ID, accounts[1].ID
	_, err = svc.CreateJournal(ctx, admin.ID, CreateJournalRequest{
		EntryDate:   "2026-06-26",
		Description: "negative debit",
		Lines: []JournalLineInput{
			{AccountID: a, DebitCents: -100, CreditCents: 0},
			{AccountID: b, DebitCents: 0, CreditCents: 100},
		},
	})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestValidateLines_UnknownAccount(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:acct_unk?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-unk@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo)

	accounts, err := svc.ListAccounts(ctx, admin.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	unknown := uuid.New()
	_, err = svc.CreateJournal(ctx, admin.ID, CreateJournalRequest{
		EntryDate:   "2026-06-26",
		Description: "unknown account",
		Lines: []JournalLineInput{
			{AccountID: unknown, DebitCents: 100, CreditCents: 0},
			{AccountID: accounts[0].ID, DebitCents: 0, CreditCents: 100},
		},
	})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestCreateInvoiceFromWonDeal_SkipsEmptyTitle(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:acct_skip?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-skip@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo)

	err = svc.CreateInvoiceFromWonDeal(ctx, admin.ID, uuid.New(), "Client", "   ", 1000, "RUB")
	require.NoError(t, err)

	invoices, err := svc.ListInvoices(ctx, admin.ID)
	require.NoError(t, err)
	assert.Empty(t, invoices)
}
