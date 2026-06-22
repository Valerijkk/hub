package accounting_test
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func TestParseEntryDate_EmptyAndValid(t *testing.T) {
	today, err := parseEntryDate("")
	require.NoError(t, err)
	assert.Equal(t, time.Now().UTC().Truncate(24*time.Hour), today)

	parsed, err := parseEntryDate("2026-06-20")
	require.NoError(t, err)
	assert.Equal(t, 2026, parsed.Year())
	assert.Equal(t, time.June, parsed.Month())

	_, err = parseEntryDate("not-a-date")
	require.Error(t, err)
}

func TestValidateLines_UnbalancedViaService(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:acct_val?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("cfo-val@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo)

	accounts, err := svc.ListAccounts(ctx, admin.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(accounts), 2)

	var a, b uuid.UUID
	a, b = accounts[0].ID, accounts[1].ID
	for _, acct := range accounts {
		if acct.Code == "5010" {
			a = acct.ID
		}
		if acct.Code == "1010" {
			b = acct.ID
		}
	}

	_, err = svc.CreateJournal(ctx, admin.ID, CreateJournalRequest{
		EntryDate:   "2026-06-26",
		Description: "unbalanced",
		Lines: []JournalLineInput{
			{AccountID: a, DebitCents: 100, CreditCents: 0},
			{AccountID: b, DebitCents: 0, CreditCents: 50},
		},
	})
	assert.ErrorIs(t, err, ErrUnbalanced)
}
