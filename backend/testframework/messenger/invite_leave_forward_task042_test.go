package messenger_test

// TASK-042 / P2-BE-13: last owner leave, forward unread, mention filter, AcceptInvite ok.
import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func TestAcceptInvite_PayloadViaInviteBool_P2_BE_13(t *testing.T) {
	// Регрессия формы payload: via_invite — bool true, не raw token (W3-A2-010).
	// PublishJSON(nil) no-op; проверяем канон сериализации ожидаемого события.
	token := "secret-invite-token-must-not-appear"
	inviteID := uuid.New()
	payload := map[string]interface{}{
		"channel_id": uuid.New(),
		"user_id":    uuid.New(),
		"via_invite": true,
		"invite_id":  inviteID,
	}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"via_invite":true`)
	assert.NotContains(t, string(raw), token)

	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	inv := &ent.ChannelInvite{
		ID: inviteID, ChannelID: channelID, Token: token,
		MaxUses: 5, CreatedAt: time.Now().UTC(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	ch, err := svc.AcceptInvite(ctx, token, joinerID)
	require.NoError(t, err)
	require.NotNil(t, ch)
}

// TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13 — mutation: via_invite false/omit → RED.
// Ловит реальный PublishJSON из AcceptInvite (не handmade map).
func TestAcceptInvite_NATSPayloadViaInvite_P2_BE_13(t *testing.T) {
	opts := &server.Options{Port: -1}
	ns := natsserver.RunServer(opts)
	t.Cleanup(ns.Shutdown)

	nc, err := natsclient.NewClient(context.Background(), ns.ClientURL())
	require.NoError(t, err)
	t.Cleanup(func() { nc.Close() })

	repo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceWithNATS(repo, nc)
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()
	token := "secret-invite-token-must-not-appear"
	inviteID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID)
	inv := &ent.ChannelInvite{
		ID: inviteID, ChannelID: channelID, Token: token,
		MaxUses: 5, CreatedAt: time.Now().UTC(),
	}
	repo.invites[inv.ID] = inv
	repo.invitesByToken[inv.Token] = inv

	subject := fmt.Sprintf(messenger.SubjectMemberJoined, channelID.String())
	got := make(chan []byte, 1)
	sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		got <- append([]byte(nil), msg.Data...)
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sub.Unsubscribe() })
	require.NoError(t, nc.FlushTimeout(time.Second))

	_, err = svc.AcceptInvite(ctx, token, joinerID)
	require.NoError(t, err)

	select {
	case raw := <-got:
		assert.Contains(t, string(raw), `"via_invite":true`)
		assert.NotContains(t, string(raw), token)
		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &payload))
		assert.Equal(t, true, payload["via_invite"])
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for member_joined NATS event")
	}
}

func TestLeaveChannel_LastOwnerRejected_P2_BE_13(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	err := svc.LeaveChannel(ctx, channelID, ownerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner")
}

func TestForwardMessage_IncrementsUnread_P2_BE_13(t *testing.T) {
	repo := newMockRepository()
	_, rdb := newTestRedis(t)
	unread := messenger.NewUnreadService(rdb)
	svc := messenger.NewMessengerService(repo, nil, nil, unread)
	ctx := context.Background()

	userID := uuid.New()
	peerID := uuid.New()
	sourceID := setupChannelWithMembers(repo, "public", false, userID)
	targetID := setupChannelWithMembers(repo, "public", false, userID, peerID)

	doc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   sourceID,
		SenderID:    userID,
		ContentText: "fwd me",
		CreatedAt:   time.Now().UTC(),
	}
	repo.messages[doc.ID] = doc

	_, err := svc.ForwardMessage(ctx, sourceID, userID, doc.ID, targetID.String())
	require.NoError(t, err)

	counts, err := unread.GetAll(ctx, peerID.String())
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[targetID.String()])
}

func TestResolveMentions_DropsNonMember_P2_BE_13(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	mentions := messenger.ExportMessengerServiceResolveGroupMentions(
		svc, ctx, channelID, nil, nil,
		[]string{memberID.String(), outsiderID.String()},
		ownerID,
	)
	assert.Contains(t, mentions, memberID.String())
	assert.NotContains(t, mentions, outsiderID.String())
}
