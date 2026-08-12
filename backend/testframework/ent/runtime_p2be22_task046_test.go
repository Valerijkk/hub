//go:build hubtest

package ent_test

// TASK-046 / P2-BE-22: Ent-runtime unique + XOR CHECK (не только schema grep).
import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/ledgeraccount"
	"github.com/valerii/hub/backend/ent/user"
)

// TestComponent_UniqueNamePerProject_P2BE22 — duplicate (project,name) → constraint.
func TestComponent_UniqueNamePerProject_P2BE22(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:comp_unique_p2be22_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Proj").SetKey("CMP1").Save(ctx)
	require.NoError(t, err)

	_, err = client.Component.Create().
		SetName("API").
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Component.Create().
		SetName("API").
		SetProjectID(proj.ID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err), "unique (project,name): %v", err)
}

// TestJournalLine_DebitXorCredit_P2BE22 — debit>0 AND credit>0 → CHECK fail.
func TestJournalLine_DebitXorCredit_P2BE22(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:jl_xor_p2be22_"+t.Name()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("jl-xor@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	entry, err := client.JournalEntry.Create().
		SetEntryDate(time.Now().UTC()).
		SetDescription("xor test").
		SetCreatedByID(admin.ID).
		Save(ctx)
	require.NoError(t, err)

	acct, err := client.LedgerAccount.Create().
		SetCode("1000").
		SetName("Cash").
		SetAccountType(ledgeraccount.AccountTypeAsset).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.JournalLine.Create().
		SetDebitCents(100).
		SetCreditCents(100).
		SetEntryID(entry.ID).
		SetAccountID(acct.ID).
		Save(ctx)
	require.Error(t, err, "XOR CHECK must reject debit>0 AND credit>0")
}
