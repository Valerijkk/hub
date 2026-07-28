package calls_test

// TASK-042 / P2-BE-15: guest entropy, conflicts IDOR, ephemeral limit, DurationMinutes.
import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/calls"
)

func TestCreateInstantCall_InviteCodeFullUUID_P2_BE_15(t *testing.T) {
	f := newCallsTestFixture(t)
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))

	resp, err := f.svc.CreateInstantCall(context.Background(), f.owner.ID, "Owner", "Quick")
	require.NoError(t, err)
	require.NotEmpty(t, resp.InviteCode)
	// W3-A5-008: 12 hex (~48 bit), не усечённые 8.
	assert.Len(t, resp.InviteCode, 12)
	assert.Regexp(t, `^[0-9a-f]{12}$`, resp.InviteCode)
}

func TestCheckConflicts_StrangerForbidden_P2_BE_15(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	scheduled := calls.NewScheduledCallService(f.repo, nil)

	_, err := scheduled.CheckConflicts(ctx, f.owner.ID, time.Now().UTC().Add(time.Hour), 60, []uuid.UUID{f.stranger.ID})
	require.ErrorIs(t, err, calls.ErrCallForbidden)
}

func TestSendChat_EphemeralOverLimit_P2_BE_15(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt, err := f.repo.CreateInstantCall(ctx, "room-eph", f.owner.ID, "abc123def456", "Eph")
	require.NoError(t, err)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	big := strings.Repeat("я", 2001)
	require.Equal(t, 2001, utf8.RuneCountInString(big))

	_, err = f.svc.SendChatMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", big, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "2000")
}

func TestCreateScheduledCall_PersistsDuration_P2_BE_15(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("Dur").SetKey("DUR").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	reminder := 15
	scheduled := calls.NewScheduledCallService(f.repo, nil)
	resp, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title:           "With duration",
		ChannelID:       f.channel.ID,
		ScheduledAt:     time.Now().UTC().Add(2 * time.Hour),
		DurationMinutes: 90,
		ReminderMinutes: &reminder,
		ParticipantIDs:  []uuid.UUID{f.owner.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, 90, resp.DurationMinutes)
	require.NotNil(t, resp.ReminderMinutes)
	assert.Equal(t, 15, *resp.ReminderMinutes)

	got, err := f.client.ScheduledCall.Get(ctx, resp.ID)
	require.NoError(t, err)
	assert.Equal(t, 90, got.DurationMinutes)
}
