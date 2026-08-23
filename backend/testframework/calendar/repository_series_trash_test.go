package calendar_test

// FR-CAL-05: удалённая серия не попадает в календарную выборку (фильтр deleted_at).
// Мутация: снять DeletedAtIsNil в ListScheduledCalls → тест краснеет.
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/calendar"
)

func TestEntRepository_ListScheduledCalls_ExcludesSoftDeleted(t *testing.T) {
	dsn := "file:caltrash_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	user, err := client.User.Create().
		SetEmail("cal-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Cal Trash").
		SetKey("CTR").
		Save(ctx)
	require.NoError(t, err)

	ch, err := client.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("voice").
		SetCreatedBy(user.ID).
		Save(ctx)
	require.NoError(t, err)

	at := time.Now().UTC().Add(24 * time.Hour)
	active, err := client.ScheduledCall.Create().
		SetTitle("Active series").
		SetScheduledAt(at).
		SetRrule("FREQ=DAILY;COUNT=3").
		SetChannelID(ch.ID).
		SetProjectID(proj.ID).
		SetCreatedBy(user.ID).
		SetIsActive(true).
		Save(ctx)
	require.NoError(t, err)

	// Soft-deleted с is_active=true: именно DeletedAtIsNil — защита, которую бьёт мутация TASK-002.
	deleted, err := client.ScheduledCall.Create().
		SetTitle("Deleted series").
		SetScheduledAt(at).
		SetRrule("FREQ=DAILY;COUNT=5").
		SetChannelID(ch.ID).
		SetProjectID(proj.ID).
		SetCreatedBy(user.ID).
		SetIsActive(true).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	repo := calendar.NewEntRepository(client)
	list, err := repo.ListScheduledCalls(ctx, []uuid.UUID{proj.ID})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, active.ID, list[0].ID)

	for _, item := range list {
		assert.NotEqual(t, deleted.ID, item.ID, "soft-deleted series must not appear in calendar list")
	}

	// После restore (ClearDeletedAt + IsActive) серия снова видна — число вхождений RRULE то же.
	require.NoError(t, client.ScheduledCall.UpdateOneID(deleted.ID).
		ClearDeletedAt().
		SetIsActive(true).
		Exec(ctx))

	listAfter, err := repo.ListScheduledCalls(ctx, []uuid.UUID{proj.ID})
	require.NoError(t, err)
	require.Len(t, listAfter, 2)

	from := at.Add(-time.Hour)
	to := at.Add(10 * 24 * time.Hour)
	var deletedRow *ent.ScheduledCall
	for _, item := range listAfter {
		if item.ID == deleted.ID {
			deletedRow = item
			break
		}
	}
	require.NotNil(t, deletedRow)
	events := calendar.ExportExpandCallEvents(deletedRow, from, to)
	assert.Equal(t, 5, len(events), "restore must return same RRULE occurrence count")
}
