//go:build hubtest

package accounting_test

// TASK-169 / A3W8-3 — HubTable→columns/records + JournalEntry→lines OnDelete Cascade.
// Живёт в accounting (SCOPE: accounting IT под PermanentDelete/purge), чтобы не
// блокироваться незавершённым WIP TASK-171 в testframework/ent.
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

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/hubcolumn"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/hubtable"
	"github.com/valerii/hub/backend/ent/journalentry"
	"github.com/valerii/hub/backend/ent/journalline"
	"github.com/valerii/hub/backend/ent/ledgeraccount"
	"github.com/valerii/hub/backend/ent/user"
)

func task169SchemaDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")
}

// TestSchema_TASK169_OnDeleteCascades — A3W8-3-001 / 002 schema markers.
// Mutation: убрать entsql.OnDelete(Cascade) с HubTable columns/records или
// JournalEntry lines → FAIL.
func TestSchema_TASK169_OnDeleteCascades(t *testing.T) {
	dir := task169SchemaDir(t)
	mustContain := map[string][]string{
		"hubtable.go": {
			`edge.To("columns"`,
			`edge.To("records"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"journalentry.go": {
			`edge.To("lines"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
	}
	for file, needles := range mustContain {
		src, err := os.ReadFile(filepath.Join(dir, file))
		require.NoError(t, err, file)
		body := string(src)
		for _, n := range needles {
			assert.Contains(t, body, n, "%s missing %q", file, n)
		}
	}
}

// TestSchema_TASK169_MigrateSQLExists — migrate marker for live Postgres.
func TestSchema_TASK169_MigrateSQLExists(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260729_task169_hubtable_journalentry_cascade.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "hub_columns_hub_tables_columns")
	assert.Contains(t, body, "hub_records_hub_tables_records")
	assert.Contains(t, body, "journal_lines_journal_entries_lines")
	assert.Contains(t, body, "ON DELETE CASCADE")
}

// TestCascade_TASK169_HubTableHardDeleteRemovesColumnsRecords — A3W8-3-001.
// Mutation: убрать Cascade с HubTable→columns/records → DeleteOneID FK trap / orphans.
func TestCascade_TASK169_HubTableHardDeleteRemovesColumnsRecords(t *testing.T) {
	dsn := "file:task169_ht_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("ht169-" + uuid.New().String()[:8] + "@t.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)

	base, err := client.HubBase.Create().
		SetName("Base169").
		SetOwnerID(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	table, err := client.HubTable.Create().
		SetName("T169").
		SetBaseID(base.ID).
		Save(ctx)
	require.NoError(t, err)

	col, err := client.HubColumn.Create().
		SetName("col").
		SetColType(hubcolumn.ColTypeText).
		SetTableID(table.ID).
		Save(ctx)
	require.NoError(t, err)

	rec, err := client.HubRecord.Create().
		SetTableID(table.ID).
		SetCells(map[string]interface{}{col.ID.String(): "v"}).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.HubTable.DeleteOneID(table.ID).Exec(ctx))

	_, err = client.HubTable.Get(ctx, table.ID)
	require.True(t, ent.IsNotFound(err))

	nCol, err := client.HubColumn.Query().Where(hubcolumn.IDEQ(col.ID)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, nCol, "columns must CASCADE with hub_table hard-delete")

	nRec, err := client.HubRecord.Query().Where(hubrecord.IDEQ(rec.ID)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, nRec, "records must CASCADE with hub_table hard-delete")
}

// TestCascade_TASK169_HubTableSoftDeleteKeepsChildren — soft path не трогает OnDelete.
func TestCascade_TASK169_HubTableSoftDeleteKeepsChildren(t *testing.T) {
	dsn := "file:task169_htsoft_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("hts169-" + uuid.New().String()[:8] + "@t.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)

	base, err := client.HubBase.Create().
		SetName("BaseSoft").
		SetOwnerID(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	table, err := client.HubTable.Create().
		SetName("SoftT").
		SetBaseID(base.ID).
		Save(ctx)
	require.NoError(t, err)

	col, err := client.HubColumn.Create().
		SetName("c").
		SetColType(hubcolumn.ColTypeText).
		SetTableID(table.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.HubRecord.Create().
		SetTableID(table.ID).
		SetCells(map[string]interface{}{}).
		Save(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = client.HubTable.UpdateOneID(table.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)

	alive, err := client.HubTable.Query().Where(hubtable.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, alive)

	nCol, err := client.HubColumn.Query().Where(hubcolumn.IDEQ(col.ID)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, nCol, "soft-delete must not CASCADE-delete columns")

	nRec, err := client.HubRecord.Query().Where(hubrecord.HasTableWith(hubtable.IDEQ(table.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, nRec, "soft-delete must not CASCADE-delete records")
}

// TestCascade_TASK169_JournalEntryHardDeleteRemovesLines — A3W8-3-002.
// Mutation: убрать Cascade с JournalEntry→lines → DeleteOneID FK trap.
func TestCascade_TASK169_JournalEntryHardDeleteRemovesLines(t *testing.T) {
	dsn := "file:task169_je_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("je169-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	accDebit, err := client.LedgerAccount.Create().
		SetCode("1000").
		SetName("Cash").
		SetAccountType(ledgeraccount.AccountTypeAsset).
		Save(ctx)
	require.NoError(t, err)
	accCredit, err := client.LedgerAccount.Create().
		SetCode("4000").
		SetName("Revenue").
		SetAccountType(ledgeraccount.AccountTypeRevenue).
		Save(ctx)
	require.NoError(t, err)

	entry, err := client.JournalEntry.Create().
		SetEntryDate(time.Now().UTC()).
		SetDescription("cascade entry").
		SetCreatedByID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	line1, err := client.JournalLine.Create().
		SetEntryID(entry.ID).
		SetAccountID(accDebit.ID).
		SetDebitCents(500).
		Save(ctx)
	require.NoError(t, err)
	line2, err := client.JournalLine.Create().
		SetEntryID(entry.ID).
		SetAccountID(accCredit.ID).
		SetCreditCents(500).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.JournalEntry.DeleteOneID(entry.ID).Exec(ctx))

	_, err = client.JournalEntry.Get(ctx, entry.ID)
	require.True(t, ent.IsNotFound(err))

	n, err := client.JournalLine.Query().Where(
		journalline.IDIn(line1.ID, line2.ID),
	).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "lines must CASCADE with journal_entry hard-delete")
}

// TestCascade_TASK169_JournalEntrySoftDeleteKeepsLines — soft path не трогает OnDelete.
func TestCascade_TASK169_JournalEntrySoftDeleteKeepsLines(t *testing.T) {
	dsn := "file:task169_jesoft_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("jes169-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	acc, err := client.LedgerAccount.Create().
		SetCode("1100").
		SetName("Bank").
		SetAccountType(ledgeraccount.AccountTypeAsset).
		Save(ctx)
	require.NoError(t, err)
	acc2, err := client.LedgerAccount.Create().
		SetCode("4100").
		SetName("Sales").
		SetAccountType(ledgeraccount.AccountTypeRevenue).
		Save(ctx)
	require.NoError(t, err)

	entry, err := client.JournalEntry.Create().
		SetEntryDate(time.Now().UTC()).
		SetDescription("soft entry").
		SetCreatedByID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.JournalLine.Create().
		SetEntryID(entry.ID).
		SetAccountID(acc.ID).
		SetDebitCents(100).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.JournalLine.Create().
		SetEntryID(entry.ID).
		SetAccountID(acc2.ID).
		SetCreditCents(100).
		Save(ctx)
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = client.JournalEntry.UpdateOneID(entry.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)

	alive, err := client.JournalEntry.Query().Where(journalentry.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, alive)

	n, err := client.JournalLine.Query().Where(journalline.HasEntryWith(journalentry.IDEQ(entry.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, n, "soft-delete must not CASCADE-delete journal lines")
}
