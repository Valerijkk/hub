package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"bytes"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/recording"
)

func TestHandler_ListProjectRecordings_Success(t *testing.T) {
	svc, projectID, client := newRecordingListFixture(t)
	ctx := context.Background()
	owner, err := client.User.Query().First(ctx)
	require.NoError(t, err)

	h := calls.ExportNewHandlerWithRecSvc(svc)
	app := newHandlerValidationApp(h)

	req := httptest.NewRequest(fiber.MethodGet, "/projects/"+projectID.String()+"/recordings?offset=0&limit=10", nil)
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestScheduledCallService_CheckConflicts_WithOverlap(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("Conflict").SetKey("CNF").Save(ctx)
	require.NoError(t, err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	at := time.Now().UTC().Add(2 * time.Hour)
	_, err = scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Overlap", ChannelID: f.channel.ID, ScheduledAt: at,
	})
	require.NoError(t, err)

	resp, err := scheduled.CheckConflicts(ctx, at.Add(15*time.Minute), 60, []uuid.UUID{f.owner.ID})
	require.NoError(t, err)
	assert.True(t, resp.HasConflict)
	assert.NotEmpty(t, resp.Conflicts)
}

func TestScheduledCallService_ProcessDueScheduledCalls(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("Due").SetKey("DUE").Save(ctx)
	require.NoError(t, err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	past := time.Now().UTC().Add(-time.Minute)
	_, err = scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Past call", ChannelID: f.channel.ID, ScheduledAt: past,
	})
	require.NoError(t, err)

	require.NoError(t, scheduled.ProcessDueScheduledCalls(ctx))
}

func TestScheduledCallService_UpdateScheduledCall_InvalidRRule(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("BadRR").SetKey("BRR").Save(ctx)
	require.NoError(t, err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "RR", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	bad := "NOT_VALID"
	_, err = scheduled.UpdateScheduledCall(ctx, created.ID, calls.UpdateScheduledCallRequest{RRule: &bad})
	require.Error(t, err)
}

func TestScheduledCallService_ProcessDueScheduledCalls_Recurring(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("Recur").SetKey("RCR").Save(ctx)
	require.NoError(t, err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	past := time.Now().UTC().Add(-time.Minute)
	rruleStr := "FREQ=DAILY;BYHOUR=10;BYMINUTE=0;BYSECOND=0"
	created, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Daily", ChannelID: f.channel.ID, ScheduledAt: past, RRule: rruleStr,
	})
	require.NoError(t, err)

	require.NoError(t, scheduled.ProcessDueScheduledCalls(ctx))

	updated, err := f.client.ScheduledCall.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, updated.IsActive)
	assert.True(t, updated.ScheduledAt.After(past))
}

func TestRecordingService_OnEgressEnded(t *testing.T) {
	svc, _, client := newRecordingListFixture(t)
	ctx := context.Background()

	rec, err := client.Recording.Query().First(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.OnEgressEnded(ctx, rec.EgressID, 2048, 120))
	updated, err := client.Recording.Get(ctx, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, recording.StatusReady, updated.Status)
	assert.Equal(t, int64(2048), updated.FileSize)
}

func TestRepository_FindConflictingCalls(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("Conflict2").SetKey("CF2").Save(ctx)
	require.NoError(t, err)

	at := time.Now().UTC().Add(2 * time.Hour)
	_, err = f.repo.CreateScheduledCall(ctx, "Standup", "", at, "", f.channel.ID, proj.ID, f.owner.ID)
	require.NoError(t, err)

	windowStart := at.Add(-30 * time.Minute)
	windowEnd := at.Add(30 * time.Minute)
	conflicts, err := f.repo.FindConflictingCalls(ctx, []uuid.UUID{f.owner.ID}, windowStart, windowEnd)
	require.NoError(t, err)
	assert.NotEmpty(t, conflicts)
}

func TestCallService_SetRaiseHand_LiveKitError(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	err := f.svc.SetRaiseHand(ctx, callEnt.RoomName, f.owner.ID.String(), true)
	require.Error(t, err)
}

func TestCallChatService_SendMessage_PersistentWithoutMongo(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	_, err := calls.ExportCallServiceChat(f.svc).SendMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "persist", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не настроен")
}

func TestCallService_CreateCall_LiveKitUnavailable(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	_, err := f.svc.CreateCall(ctx, f.channel.ID, f.owner.ID, "Owner")
	require.Error(t, err)
}

func TestCallService_CreateInstantCall_LiveKitUnavailable(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	_, err := f.svc.CreateInstantCall(ctx, f.owner.ID, "Owner", "Quick")
	require.Error(t, err)
}

func TestHandler_RaiseHand_LiveKitUnavailable(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))
	app := newCallsHandlerApp(f)

	req := httptest.NewRequest(fiber.MethodPost, "/calls/"+callEnt.ID.String()+"/raise-hand", bytes.NewReader([]byte(`{"raised":true}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.owner.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
