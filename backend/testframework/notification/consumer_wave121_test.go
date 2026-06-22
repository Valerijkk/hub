package notification_test

// Тесты NATS consumer уведомлений (волна 121): интеграционные и unit-сценарии testframework.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entnotif "github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/internal/notification"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/tracker"
)

// Фаза 121 — testframework external tests (export_test.go hooks)..

func TestExportHandleStatusChanged_NotifiesAssignee(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.IssueMovedEvent{
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		MovedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleStatusChanged(c, ctx, data, "")

	db := notification.ExportRepositoryDB(repo)
	notifs, err := db.Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeStatusChanged, notifs[0].Type)
}

// TestExportShouldDispatch_Debounce проверяет сценарий: экспорт should dispatch дебаунс.
func TestExportShouldDispatch_Debounce(t *testing.T) {
	c := notification.NewConsumer(nil, nil, nil)
	key := uuid.New().String()

	assert.True(t, notification.ExportShouldDispatch(c, key, notification.TypeAssigned, "issue-1"))
	assert.False(t, notification.ExportShouldDispatch(c, key, notification.TypeAssigned, "issue-1"))
}

// TestExportBuildPushURL проверяет сценарий: экспорт build push url.
func TestExportBuildPushURL(t *testing.T) {
	assert.Equal(t, "/tracker/issues/abc", notification.ExportBuildPushURL("issue", "abc"))
	assert.Equal(t, "/", notification.ExportBuildPushURL("unknown", "x"))
}

// TestExportProcessEvent_UnknownSubjectAcks проверяет сценарий: экспорт process event unknown subject acks.
func TestExportProcessEvent_UnknownSubjectAcks(t *testing.T) {
	c, _, ctx, _, _, _, _ := newConsumerIssueEnv(t)
	msg := &stubJSMsg{subject: "unknown.event", data: []byte(`{}`)}
	notification.ExportProcessEvent(c, ctx, msg)
	assert.True(t, msg.acked)
}

// TestExportHandleAssigned_CreatesNotification проверяет сценарий: экспорт handle assigned creates notification.
func TestExportHandleAssigned_CreatesNotification(t *testing.T) {
	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	db := notification.ExportRepositoryDB(repo)
	actor, err := db.User.Create().
		SetEmail("actor-" + uuid.New().String()[:8] + "@test.com").
		SetName("Actor").
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	evt := tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-99",
		ProjectID:    uuid.New().String(),
		AssigneeID:   assigneeID.String(),
		AssignedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleAssigned(c, ctx, data, "project.x.issue.assigned")

	notifs, err := db.Notification.Query().
		Where(entnotif.UserID(assigneeID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeAssigned, notifs[0].Type)
}

// TestExportConsumerStart_NilNATS проверяет сценарий: экспорт consumer start nil nats.
func TestExportConsumerStart_NilNATS(t *testing.T) {
	c := notification.NewConsumer(nil, nil, nil)
	require.NoError(t, notification.ExportConsumerStart(c, context.Background()))
}

// TestExportTelegramHandleLinkCommand_GeneratesCode проверяет сценарий: экспорт Telegram handle link command generates code.
func TestExportTelegramHandleLinkCommand_GeneratesCode(t *testing.T) {
	svc, rdb, ctx, _ := setupTelegramBindForHandleMessage(t)
	notification.ExportTelegramHandleLinkCommand(svc, ctx, 7001, "888777666")
	keys, err := rdb.Keys(ctx, notification.ExportTGLinkCodePrefix()+"*").Result()
	require.NoError(t, err)
	assert.NotEmpty(t, keys)
}

// TestExportHandleNewLogin_CreatesNotification проверяет сценарий: экспорт handle new вход creates notification.
func TestExportHandleNewLogin_CreatesNotification(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)
	evt := auth.SessionCreatedEvent{
		UserID:    userID.String(),
		SessionID: uuid.New().String(),
		Device:    "TestAgent/1.0",
		IPAddress: "10.0.0.1",
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)
	notification.ExportHandleNewLogin(c, ctx, data)
	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(userID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeNewLogin, notifs[0].Type)
}

// TestExportHandleMessageReceived_MentionInPublicChannel проверяет сценарий: экспорт handle message received mention in public channel.
func TestExportHandleMessageReceived_MentionInPublicChannel(t *testing.T) {
	svc, repo, ctx, senderID := newNotificationTestEnv(t)
	db := notification.ExportRepositoryDB(repo)

	mentioned, err := db.User.Create().
		SetEmail("mention-" + uuid.New().String()[:8] + "@test.com").
		SetName("Mentioned").
		Save(ctx)
	require.NoError(t, err)

	ch, err := db.Channel.Create().
		SetChannelType(channel.ChannelTypePublic).
		SetName("general").
		SetCreatedBy(senderID).
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	subject := "messenger.channel." + ch.ID.String() + ".message_sent"
	longText := strings.Repeat("x", 120)
	evt := struct {
		MessageID        string   `json:"message_id"`
		ChannelID        string   `json:"channel_id"`
		SenderID         string   `json:"sender_id"`
		ContentText      string   `json:"content_text"`
		MentionedUserIDs []string `json:"mentioned_user_ids"`
	}{
		MessageID:        uuid.New().String(),
		ChannelID:        ch.ID.String(),
		SenderID:         senderID.String(),
		ContentText:      longText,
		MentionedUserIDs: []string{mentioned.ID.String()},
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleMessageReceived(c, ctx, data, subject)

	notifs, err := db.Notification.Query().
		Where(entnotif.UserID(mentioned.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeMentioned, notifs[0].Type)
	assert.Contains(t, notifs[0].Body, "...")
}

// TestExportConsumerStart_WithJetStream проверяет сценарий: экспорт consumer start with jet stream.
func TestExportConsumerStart_WithJetStream(t *testing.T) {
	nc := startWave79NATSServer(t)
	require.NoError(t, nc.EnsureStreams(context.Background()))
	svc, repo, _, _ := newNotificationTestEnv(t)
	c := notification.NewConsumer(nc, repo, svc)
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, notification.ExportConsumerStart(c, runCtx))
	time.Sleep(80 * time.Millisecond)
	c.Stop()
}
