package accounting_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/accounting"
)

// TestEntRepository_JournalLineJoinFilter_A2W8_3_004 — soft-deleted entry hides lines in TrialBalance.
// ADR-ENT-SOFT-DELETE-GAPS-2026 §B: JournalLine без own deleted_at; join-filter через entry.
// Mutation: drop DeletedAtIsNil from TrialBalance HasEntryWith → soft-deleted lines counted → FAIL.
func TestEntRepository_JournalLineJoinFilter_A2W8_3_004(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:jlsoft097_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("jl-" + uuid.New().String()[:8] + "@t.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	repo := accounting.NewEntRepository(client)
	require.NoError(t, repo.EnsureDefaultChart(ctx))

	accounts, err := repo.ListAccounts(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(accounts), 2)

	entry, err := repo.CreateJournalEntry(ctx, admin.ID, time.Now().UTC(), accounting.CreateJournalRequest{
		Description: "soft-delete parent",
		Lines: []accounting.JournalLineInput{
			{AccountID: accounts[0].ID, DebitCents: 1000},
			{AccountID: accounts[1].ID, CreditCents: 1000},
		},
	})
	require.NoError(t, err)
	_, err = repo.PostJournalEntry(ctx, entry.ID)
	require.NoError(t, err)

	before, err := repo.TrialBalance(ctx)
	require.NoError(t, err)
	var debitBefore int64
	for _, row := range before {
		if row.AccountID == accounts[0].ID {
			debitBefore = row.DebitCents
		}
	}
	assert.GreaterOrEqual(t, debitBefore, int64(1000))

	require.NoError(t, repo.SoftDeleteJournalEntry(ctx, entry.ID))

	after, err := repo.TrialBalance(ctx)
	require.NoError(t, err)
	var debitAfter int64
	for _, row := range after {
		if row.AccountID == accounts[0].ID {
			debitAfter = row.DebitCents
		}
	}
	assert.Equal(t, debitBefore-1000, debitAfter, "A2W8-3-004: soft-deleted entry lines must leave TrialBalance")

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	src, err := os.ReadFile(filepath.Join(root, "backend", "internal", "accounting", "repository.go"))
	require.NoError(t, err)
	assert.Contains(t, string(src), "DeletedAtIsNil()")
	assert.Contains(t, string(src), "HasEntryWith")
}
