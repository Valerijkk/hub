package notification_test

// Тесты NATS consumer уведомлений: интеграционные и unit-сценарии testframework.

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	entnotif "github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/internal/notification"
)

// TestConsumer_handleBlockerAdded_NotifiesAssignee проверяет сценарий: consumer handle blocker added notifies assignee.
func TestConsumer_handleBlockerAdded_NotifiesAssignee(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)
	db := notification.ExportRepositoryDB(repo)

	evt := struct {
		SourceID  string `json:"source_id"`
		TargetID  string `json:"target_id"`
		LinkType  string `json:"link_type"`
		ProjectID string `json:"project_id"`
		CreatedBy string `json:"created_by"`
	}{
		SourceID:  uuid.New().String(),
		TargetID:  issueEnt.ID.String(),
		LinkType:  "blocks",
		ProjectID: projectID.String(),
		CreatedBy: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleBlockerAdded(c, ctx, data, "")

	notifs, err := db.Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeBlockerAdded, notifs[0].Type)
	assert.True(t, notifs[0].IsUrgent)
}

// TestConsumer_handleMessageReceived_DMNotifiesRecipient проверяет сценарий: consumer handle message received dm notifies recipient.
func TestConsumer_handleMessageReceived_DMNotifiesRecipient(t *testing.T) {
	svc, repo, ctx, senderID := newNotificationTestEnv(t)
	db := notification.ExportRepositoryDB(repo)

	recipient, err := db.User.Create().
		SetEmail("recv-" + uuid.New().String()[:8] + "@test.com").
		SetName("Recipient").
		Save(ctx)
	require.NoError(t, err)

	ch, err := db.Channel.Create().
		SetChannelType(channel.ChannelTypeDm).
		SetCreatedBy(senderID).
		Save(ctx)
	require.NoError(t, err)

	for _, uid := range []uuid.UUID{senderID, recipient.ID} {
		role := channelmember.RoleMember
		if uid == senderID {
			role = channelmember.RoleOwner
		}
		_, err = db.ChannelMember.Create().
			SetChannelID(ch.ID).
			SetUserID(uid).
			SetRole(role).
			Save(ctx)
		require.NoError(t, err)
	}

	c := notification.NewConsumer(nil, repo, svc)
	subject := "messenger.channel." + ch.ID.String() + ".message_sent"
	evt := struct {
		MessageID   string `json:"message_id"`
		ChannelID   string `json:"channel_id"`
		SenderID    string `json:"sender_id"`
		ContentText string `json:"content_text"`
	}{
		MessageID:   uuid.New().String(),
		ChannelID:   ch.ID.String(),
		SenderID:    senderID.String(),
		ContentText: "Hello DM",
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleMessageReceived(c, ctx, data, subject)

	notifs, err := db.Notification.Query().
		Where(entnotif.UserID(recipient.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeMessageReceived, notifs[0].Type)
}

// TestConsumer_handleCallStarted_NotifiesChannelMembers проверяет сценарий: consumer handle call started notifies channel members.
func TestConsumer_handleCallStarted_NotifiesChannelMembers(t *testing.T) {
	svc, repo, ctx, starterID := newNotificationTestEnv(t)
	db := notification.ExportRepositoryDB(repo)

	member, err := db.User.Create().
		SetEmail("call-" + uuid.New().String()[:8] + "@test.com").
		SetName("Callee").
		Save(ctx)
	require.NoError(t, err)

	ch, err := db.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("voice-room").
		SetCreatedBy(starterID).
		Save(ctx)
	require.NoError(t, err)

	for _, uid := range []uuid.UUID{starterID, member.ID} {
		role := channelmember.RoleMember
		if uid == starterID {
			role = channelmember.RoleOwner
		}
		_, err = db.ChannelMember.Create().
			SetChannelID(ch.ID).
			SetUserID(uid).
			SetRole(role).
			Save(ctx)
		require.NoError(t, err)
	}

	c := notification.NewConsumer(nil, repo, svc)
	callID := uuid.New().String()
	subject := "calls.channel." + ch.ID.String() + ".started"
	evt := struct {
		CallID    string `json:"call_id"`
		ChannelID string `json:"channel_id"`
		StartedBy string `json:"started_by"`
	}{
		CallID:    callID,
		ChannelID: ch.ID.String(),
		StartedBy: starterID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleCallStarted(c, ctx, data, subject)

	notifs, err := db.Notification.Query().
		Where(entnotif.UserID(member.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeCallStarted, notifs[0].Type)
}

// TestIsRunFailedRecipient проверяет сценарий: is run failed recipient.
func TestIsRunFailedRecipient(t *testing.T) {
	completer := uuid.New()
	tester := uuid.New()
	dev := uuid.New()

	assert.True(t, notification.ExportIsRunFailedRecipient(projectmember.ProjectRoleDeveloper, completer, completer))
	assert.True(t, notification.ExportIsRunFailedRecipient(projectmember.ProjectRoleTester, tester, completer))
	assert.False(t, notification.ExportIsRunFailedRecipient(projectmember.ProjectRoleDeveloper, dev, completer))
}
