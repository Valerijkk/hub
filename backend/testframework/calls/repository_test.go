package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/callparticipant"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/recording"
)

func TestRepository_CreateCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	callEnt, err := f.repo.CreateCall(ctx, "room-create-1", f.channel.ID, f.owner.ID)
	require.NoError(t, err)
	assert.Equal(t, f.channel.ID, *callEnt.ChannelID)
	assert.Equal(t, f.owner.ID, callEnt.CreatedBy)
}

func TestRepository_CreateInstantCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	callEnt, err := f.repo.CreateInstantCall(ctx, "room-instant-1", f.owner.ID, "abc12345", "Quick")
	require.NoError(t, err)
	assert.Equal(t, "abc12345", callEnt.InviteCode)
	assert.Nil(t, callEnt.ChannelID)
}

func TestRepository_ListParticipants(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	parts, err := f.repo.ListParticipants(ctx, callEnt.ID)
	require.NoError(t, err)
	assert.Len(t, parts, 2)
}

func TestRepository_AddParticipant_IdempotentForActive(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	count, err := f.client.CallParticipant.Query().
		Where(callparticipant.CallID(callEnt.ID), callparticipant.UserID(f.member.ID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRepository_AddParticipant_RejoinAfterLeave(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))
	require.NoError(t, f.repo.RemoveParticipant(ctx, callEnt.ID, f.member.ID, time.Now().UTC()))
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	p, err := f.client.CallParticipant.Query().
		Where(callparticipant.CallID(callEnt.ID), callparticipant.UserID(f.member.ID)).
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, p.LeftAt)
}

func TestRepository_RemoveParticipant_AlreadyLeft(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	require.NoError(t, f.repo.RemoveParticipant(ctx, callEnt.ID, f.stranger.ID, time.Now().UTC()))
}

func TestRepository_GetChannelMemberRole(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	role, err := f.repo.GetChannelMemberRole(ctx, f.channel.ID, f.owner.ID)
	require.NoError(t, err)
	assert.Equal(t, string(channelmember.RoleOwner), role)

	role, err = f.repo.GetChannelMemberRole(ctx, f.channel.ID, f.stranger.ID)
	require.NoError(t, err)
	assert.Equal(t, "", role)
}

func TestRepository_GetProjectIDByChannel(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	proj, err := f.client.Project.Create().SetName("Voice Proj").SetKey("VPR").Save(ctx)
	require.NoError(t, err)
	_, err = f.client.Channel.UpdateOne(f.channel).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)

	got, err := f.repo.GetProjectIDByChannel(ctx, f.channel.ID)
	require.NoError(t, err)
	assert.Equal(t, proj.ID, got)
}

func TestRepository_RecordingCRUD(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	proj, err := f.client.Project.Create().SetName("Rec").SetKey("REC").Save(ctx)
	require.NoError(t, err)

	rec, err := f.repo.CreateRecording(ctx, "egress-1", callEnt.ID, proj.ID, f.owner.ID, "/files/rec.mp4")
	require.NoError(t, err)

	active, err := f.repo.CountActiveRecordings(ctx, callEnt.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, active)

	got, err := f.repo.GetRecordingByEgressID(ctx, "egress-1")
	require.NoError(t, err)
	assert.Equal(t, rec.ID, got.ID)

	require.NoError(t, f.repo.UpdateRecordingStatus(ctx, rec.ID, recording.StatusProcessing))
	require.NoError(t, f.repo.UpdateRecordingCompleted(ctx, rec.ID, 1024, 60, time.Now().UTC()))

	items, total, err := f.repo.ListRecordingsByProject(ctx, proj.ID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
}

func TestRepository_UpdateScheduledCallTime(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("Sched").SetKey("SCH").Save(ctx)
	require.NoError(t, err)

	at := time.Now().UTC().Add(2 * time.Hour)
	scheduled, err := f.repo.CreateScheduledCall(ctx, "Daily", "", at, "", f.channel.ID, proj.ID, f.owner.ID, 0, nil, nil)
	require.NoError(t, err)

	next := at.Add(24 * time.Hour)
	require.NoError(t, f.repo.UpdateScheduledCallTime(ctx, scheduled.ID, next))

	updated, err := f.client.ScheduledCall.Get(ctx, scheduled.ID)
	require.NoError(t, err)
	assert.WithinDuration(t, next, updated.ScheduledAt, time.Second)
}

func TestCallChatService_DeleteCallChat_NilMongo(t *testing.T) {
	f := newCallsTestFixture(t)
	assert.NotPanics(t, func() {
		calls.ExportCallServiceChat(f.svc).DeleteCallChat(context.Background(), uuid.New().String())
	})
}

func TestSendChatMessage_EndedCallRejected(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	_, err := f.client.Call.UpdateOne(callEnt).SetStatus(call.StatusEnded).Save(ctx)
	require.NoError(t, err)

	_, err = f.svc.SendChatMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "hi", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "завершён")
}

func TestNewCallService_Constructor(t *testing.T) {
	f := newCallsTestFixture(t)
	svc := calls.NewCallService(
		calls.ExportCallRoomServiceCfg(calls.ExportCallServiceRoom(f.svc)),
		f.repo,
		nil,
		nil,
	)
	require.NotNil(t, svc)
	token, err := svc.GenerateToken("room-x", f.owner.ID.String(), "Owner")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestScheduledCallService_DeleteScheduledCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("Del").SetKey("DEL").Save(ctx)
	require.NoError(t, err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	f.linkChannelToProject(t, proj.ID)
	created, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Remove me", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	require.NoError(t, scheduled.DeleteScheduledCall(ctx, created.ID, f.owner.ID))
	list, err := scheduled.ListScheduledCalls(ctx, f.channel.ID, f.owner.ID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestCallChatService_SendMessage_InvalidCallID(t *testing.T) {
	f := newCallsTestFixture(t)
	_, err := calls.ExportCallServiceChat(f.svc).SendMessage(context.Background(), "not-uuid", f.owner.ID.String(), "Owner", "hi", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "некорректный call ID")
}
