package calls_test
import (
	"github.com/valerii/hub/backend/internal/calls"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/callparticipant"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/platform/config"
)

type callsTestFixture struct {
	client   *ent.Client
	svc      *calls.CallService
	repo     *calls.EntRepository
	owner    *ent.User
	member   *ent.User
	stranger *ent.User
	channel  *ent.Channel
}

func newCallsTestFixture(t *testing.T) *callsTestFixture {
	t.Helper()
	dsn := "file:callstest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	cfg := config.LiveKitConfig{
		URL:       "http://localhost:7880",
		APIKey:    "testkey",
		APISecret: "testsecretthatshouldbelongenough",
	}
	repo := calls.NewEntRepository(client)
	pub := calls.NewCallEventPublisher(nil)
	svc := calls.ExportNewCallServiceForTest(
		repo,
		calls.NewCallRoomService(cfg),
		pub,
		calls.NewCallChatService(repo, nil, pub),
	)

	ctx := context.Background()
	owner, err := client.User.Create().SetEmail("owner-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().SetEmail("member-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().SetEmail("stranger-" + uuid.New().String()[:6] + "@example.com").Save(ctx)
	require.NoError(t, err)

	ch, err := client.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("voice").
		SetCreatedBy(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ChannelMember.Create().
		SetChannelID(ch.ID).SetUserID(owner.ID).SetRole(channelmember.RoleOwner).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ChannelMember.Create().
		SetChannelID(ch.ID).SetUserID(member.ID).SetRole(channelmember.RoleMember).
		Save(ctx)
	require.NoError(t, err)

	return &callsTestFixture{
		client: client, svc: svc, repo: repo,
		owner: owner, member: member, stranger: stranger, channel: ch,
	}
}

func (f *callsTestFixture) createChannelCall(t *testing.T, creator *ent.User) *ent.Call {
	t.Helper()
	ctx := context.Background()
	c, err := f.client.Call.Create().
		SetRoomName("call-" + uuid.New().String()).
		SetCreatedBy(creator.ID).
		SetStatus(call.StatusActive).
		SetNillableChannelID(&f.channel.ID).
		Save(ctx)
	require.NoError(t, err)
	return c
}

func (f *callsTestFixture) createInstantCall(t *testing.T, creator *ent.User, code string) *ent.Call {
	t.Helper()
	ctx := context.Background()
	c, err := f.client.Call.Create().
		SetRoomName("call-instant-" + uuid.New().String()).
		SetCreatedBy(creator.ID).
		SetInviteCode(code).
		SetTitle("Quick call").
		SetStatus(call.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	return c
}

func TestJoinByInviteCode_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	code := "inv12345"
	callEnt := f.createInstantCall(t, f.owner, code)

	resp, err := f.svc.JoinByInviteCode(ctx, code, f.member.ID, "Member")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, callEnt.ID, resp.ID)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, code, resp.InviteCode)
}

func TestJoinByInviteCode_EndedCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	code := "ended123"
	callEnt := f.createInstantCall(t, f.owner, code)
	_, err := f.client.Call.UpdateOne(callEnt).SetStatus(call.StatusEnded).Save(ctx)
	require.NoError(t, err)

	_, err = f.svc.JoinByInviteCode(ctx, code, f.member.ID, "Member")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestJoinCall_ChannelMemberAllowed(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	resp, err := f.svc.JoinCall(ctx, callEnt.ID, f.member.ID, "Member")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
}

func TestJoinCall_NonMemberForbidden(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	_, err := f.svc.JoinCall(ctx, callEnt.ID, f.stranger.ID, "Stranger")
	require.Error(t, err)
	assert.ErrorIs(t, err, calls.ErrCallForbidden)
}

// TestCreateCall_NonMemberForbidden — A2-002: CreateCall требует IsChannelMember
// до provisionRoom. Mutation guard: при снятии проверки stranger создаст call-row
// (count>0) и тест упадёт; при живой проверке ErrCallForbidden и 0 call rows.
func TestCreateCall_NonMemberForbidden(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	cfg := liveKitTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	calls.ExportCallServiceSetRoom(f.svc, calls.NewCallRoomService(cfg))

	_, err := f.svc.CreateCall(ctx, f.channel.ID, f.stranger.ID, "Stranger")
	require.Error(t, err)
	assert.ErrorIs(t, err, calls.ErrCallForbidden)

	count, err := f.client.Call.Query().
		Where(call.HasChannelWith(channel.IDEQ(f.channel.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "non-member must not create a call row (A2-002)")
}

func TestJoinCall_InstantOnlyCreator(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createInstantCall(t, f.owner, "onlyownr")

	_, err := f.svc.JoinCall(ctx, callEnt.ID, f.member.ID, "Member")
	require.Error(t, err)
	assert.ErrorIs(t, err, calls.ErrCallForbidden)
}

func TestLeaveCall_RemovesParticipant(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	_, err := f.svc.JoinCall(ctx, callEnt.ID, f.member.ID, "Member")
	require.NoError(t, err)
	require.NoError(t, f.svc.LeaveCall(ctx, callEnt.ID, f.member.ID))

	parts, err := f.client.CallParticipant.Query().All(ctx)
	require.NoError(t, err)
	for _, p := range parts {
		if p.UserID == f.member.ID {
			assert.NotNil(t, p.LeftAt)
		}
	}
}

func TestGetActiveCallByChannel(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()

	none, err := f.svc.GetActiveCallByChannel(ctx, f.channel.ID)
	require.NoError(t, err)
	assert.Nil(t, none)

	callEnt := f.createChannelCall(t, f.owner)
	active, err := f.svc.GetActiveCallByChannel(ctx, f.channel.ID)
	require.NoError(t, err)
	require.NotNil(t, active)
	assert.Equal(t, callEnt.ID, active.ID)
}

func TestOnRoomStarted_UpdatesCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	require.NoError(t, f.svc.OnRoomStarted(ctx, callEnt.RoomName))

	updated, err := f.client.Call.Get(ctx, callEnt.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.StartedAt)
}

func TestOnRoomFinished_EndsCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	require.NoError(t, f.svc.OnRoomFinished(ctx, callEnt.RoomName))

	updated, err := f.client.Call.Get(ctx, callEnt.ID)
	require.NoError(t, err)
	assert.Equal(t, call.StatusEnded, updated.Status)
	require.NotNil(t, updated.EndedAt)
}

func TestOnParticipantJoined_AddsParticipant(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	require.NoError(t, f.svc.OnParticipantJoined(ctx, callEnt.RoomName, f.member.ID.String()))

	count, err := f.client.CallParticipant.Query().
		Where(
			callparticipant.CallID(callEnt.ID),
			callparticipant.UserID(f.member.ID),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestOnParticipantLeft_RemovesParticipant(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))

	require.NoError(t, f.svc.OnParticipantLeft(ctx, callEnt.RoomName, f.member.ID.String()))

	p, err := f.client.CallParticipant.Query().All(ctx)
	require.NoError(t, err)
	for _, part := range p {
		if part.UserID == f.member.ID {
			assert.NotNil(t, part.LeftAt)
			assert.True(t, part.LeftAt.Before(time.Now().UTC().Add(time.Second)))
		}
	}
}

func TestChannelIDStr(t *testing.T) {
	id := uuid.New()
	assert.Equal(t, id.String(), calls.ExportChannelIDStr(&id))
	assert.Equal(t, "instant", calls.ExportChannelIDStr(nil))
}

func TestCallEventPublisher_NilSafe(t *testing.T) {
	pub := calls.NewCallEventPublisher(nil)
	id := uuid.New()
	callID := uuid.New()

	assert.NotPanics(t, func() {
		pub.Started(&id, callID, "room-1", uuid.New())
		pub.Joined(&id, callID, uuid.New(), "User")
		pub.Left(&id, callID, uuid.New())
		pub.Ended(&id, callID, 3)
		pub.ChatMessage(callID.String(), &calls.CallChatMessage{CallID: callID.String(), Text: "hi"})
	})
}

func TestSendChatMessage_Ephemeral(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	msg, err := f.svc.SendChatMessage(ctx, callEnt.ID.String(), f.owner.ID.String(), "Owner", "hello", true)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "hello", msg.Text)
}

func TestEndCall_CreatorEndsCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.owner.ID))

	err := f.svc.EndCall(ctx, callEnt.ID, f.owner.ID)
	require.NoError(t, err)

	updated, err := f.client.Call.Get(ctx, callEnt.ID)
	require.NoError(t, err)
	assert.Equal(t, call.StatusEnded, updated.Status)
}

func TestEndCall_ChannelAdminEndsCall(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.member)

	_, err := f.client.ChannelMember.Update().
		Where(channelmember.ChannelID(f.channel.ID), channelmember.UserID(f.owner.ID)).
		SetRole(channelmember.RoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, f.repo.AddParticipant(ctx, callEnt.ID, f.member.ID))
	err = f.svc.EndCall(ctx, callEnt.ID, f.owner.ID)
	require.NoError(t, err)

	updated, err := f.client.Call.Get(ctx, callEnt.ID)
	require.NoError(t, err)
	assert.Equal(t, call.StatusEnded, updated.Status)
}

func TestScheduledCallService_UpdateScheduledCall_Success(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("Upd").SetKey("UPD").Save(ctx)
	require.NoError(t, err)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Old", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	newTitle := "Updated"
	updated, err := scheduled.UpdateScheduledCall(ctx, created.ID, calls.UpdateScheduledCallRequest{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
}

func TestJoinCall_EndedCallRejected(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)
	_, err := f.client.Call.UpdateOne(callEnt).SetStatus(call.StatusEnded).Save(ctx)
	require.NoError(t, err)

	_, err = f.svc.JoinCall(ctx, callEnt.ID, f.member.ID, "Member")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "завершён")
}

func TestListChatMessages_NotConfigured(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	_, err := f.svc.ListChatMessages(ctx, callEnt.ID.String(), f.owner.ID.String(), 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не настроен")
}

func TestResolveParticipant_InvalidIdentity(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	callResult, uid, err := calls.ExportResolveParticipant(ctx, f.repo, callEnt.RoomName, "not-a-uuid")
	require.NoError(t, err)
	assert.Nil(t, callResult)
	assert.Equal(t, uuid.Nil, uid)
}

func TestOnParticipantJoined_InvalidIdentity(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	callEnt := f.createChannelCall(t, f.owner)

	require.NoError(t, f.svc.OnParticipantJoined(ctx, callEnt.RoomName, "bad-identity"))
}
