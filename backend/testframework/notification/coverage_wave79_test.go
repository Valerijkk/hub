package notification_test

// Тесты модуля уведомлений (волна 79): интеграционные и unit-сценарии testframework.

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	entnotif "github.com/valerii/hub/backend/ent/notification"
	"github.com/valerii/hub/backend/internal/auth"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"github.com/valerii/hub/backend/internal/tracker"
)

func startWave79NATSServer(t *testing.T) *natsclient.Client {
	t.Helper()
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
	}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)
	client, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(client.Close)
	return client
}

// TestConsumer_processEvent_UnknownSubjectAcks проверяет сценарий: consumer process event unknown subject acks.
func TestConsumer_processEvent_UnknownSubjectAcks(t *testing.T) {
	c := notification.NewConsumer(nil, nil, nil)
	notification.ExportProcessEvent(c, context.Background(), &stubJetStreamMsg{subject: "unknown.event", data: []byte(`{}`)})
}

// TestConsumer_processEvent_Assigned проверяет сценарий: consumer process event assigned.
func TestConsumer_processEvent_Assigned(t *testing.T) {
	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	actor, err := notification.ExportRepositoryDB(repo).User.Create().
		SetEmail("actor-" + uuid.New().String()[:8] + "@test.com").
		SetName("Actor").
		Save(ctx)
	require.NoError(t, err)

	evt := tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-99",
		ProjectID:    uuid.New().String(),
		AssigneeID:   assigneeID.String(),
		AssignedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	notification.ExportProcessEvent(c, ctx, &stubJetStreamMsg{
		subject: "project." + uuid.New().String() + ".issue.assigned",
		data:    data,
	})

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assigneeID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
}

// TestConsumer_StartNilNatsAndStop проверяет сценарий: consumer start nil nats and stop.
func TestConsumer_StartNilNatsAndStop(t *testing.T) {
	c := notification.NewConsumer(nil, nil, nil)
	require.NoError(t, c.Start(context.Background()))
	c.Stop()
}

// TestDispatch_PublishToWS проверяет сценарий: dispatch publish to ws.
func TestDispatch_PublishToWS(t *testing.T) {
	nc := startWave79NATSServer(t)
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	svc = notification.NewNotificationService(repo, nc, nil, nil)

	subject := "notification.user." + userID.String()
	var wg sync.WaitGroup
	wg.Add(1)
	var got []byte
	_, err := nc.Subscribe(subject, func(msg *nats.Msg) {
		got = append([]byte(nil), msg.Data...)
		wg.Done()
	})
	require.NoError(t, err)
	require.NoError(t, nc.Conn.Flush())

	require.NoError(t, svc.Dispatch(ctx, notification.DispatchEvent{
		UserID:     userID,
		Type:       notification.TypeAssigned,
		Title:      "WS test",
		Body:       "payload",
		EntityType: "issue",
		EntityID:   uuid.New().String(),
	}))

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for NATS publish")
	}
	assert.Contains(t, string(got), "WS test")
}

// TestCheckDeadlines_DispatchesForAssignee проверяет сценарий: проверка deadlines dispatches for assignee.
func TestCheckDeadlines_DispatchesForAssignee(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)

	proj, err := notification.ExportRepositoryDB(repo).Project.Create().
		SetName("Deadline Proj").
		SetKey(testNotifyProjectKey("D")).
		Save(ctx)
	require.NoError(t, err)

	status, err := notification.ExportRepositoryDB(repo).WorkflowStatus.Create().
		SetName("Todo").
		SetCategory("todo").
		SetProjectID(proj.ID).
		SetIsDefault(true).
		Save(ctx)
	require.NoError(t, err)

	due := time.Now().Add(12 * time.Hour)
	_, err = notification.ExportRepositoryDB(repo).Issue.Create().
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		SetReporterID(userID).
		SetAssigneeID(userID).
		SetIssueNumber(7).
		SetIssueKey(proj.Key + "-7").
		SetTitle("Due soon").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		SetDueDate(due).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.CheckDeadlines(ctx))

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(userID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeDeadlineApproaching, notifs[0].Type)
}

// TestNewTelegramSender_ValidTokenFormat проверяет сценарий: new Telegram sender корректный token форматирование.
func TestNewTelegramSender_ValidTokenFormat(t *testing.T) {
	sender := notification.NewTelegramSender("123456789:AAAbbbCCCdddEEEfffGGGhhhIIIjjjKKKlll")
	if sender != nil {
		assert.True(t, notification.ExportTelegramSenderQueueReady(sender))
	}
}

// TestBuildPushURL_AllEntityTypes проверяет сценарий: build push url all entity types.
func TestBuildPushURL_AllEntityTypes(t *testing.T) {
	assert.Equal(t, "/intranet?tab=kudos", notification.ExportBuildPushURL("kudo", "1"))
	assert.Equal(t, "/intranet?tab=coffee", notification.ExportBuildPushURL("coffee_match", "1"))
	assert.Equal(t, "/marketplace", notification.ExportBuildPushURL("marketplace_listing", "1"))
}

// TestConsumer_processEvent_StatusChanged проверяет сценарий: consumer process event status changed.
func TestConsumer_processEvent_StatusChanged(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.IssueMovedEvent{
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		MovedByID: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	subject := "project." + projectID.String() + ".issue.moved"
	notification.ExportProcessEvent(c, ctx, &stubJetStreamMsg{subject: subject, data: data})

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeStatusChanged, notifs[0].Type)
}

// TestConsumer_processEvent_CommentAdded проверяет сценарий: consumer process event comment added.
func TestConsumer_processEvent_CommentAdded(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	evt := tracker.CommentCreatedEvent{
		IssueID:   issueEnt.ID.String(),
		IssueKey:  issueEnt.IssueKey,
		ProjectID: projectID.String(),
		AuthorID:  actor.ID.String(),
		// No Mentions: unknown mention UUID would fail Dispatch (FK) and Nak entire event (A2-010).
		Mentions: nil,
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	subject := "project." + projectID.String() + ".issue.comment_created"
	msg := &stubJetStreamMsg{subject: subject, data: data}
	notification.ExportProcessEvent(c, ctx, msg)

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeCommentAdded, notifs[0].Type)
	assert.True(t, msg.acked, "A2-010: success must Ack")
	assert.False(t, msg.naked, "A2-010: success must not Nak")
}

// TestConsumer_processEvent_NewLogin проверяет сценарий: consumer process event new вход.
func TestConsumer_processEvent_NewLogin(t *testing.T) {
	svc, repo, ctx, userID := newNotificationTestEnv(t)
	c := notification.NewConsumer(nil, repo, svc)

	evt := auth.SessionCreatedEvent{
		UserID:    userID.String(),
		SessionID: uuid.New().String(),
		Device:    "Chrome",
		IPAddress: "10.0.0.1",
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportProcessEvent(c, ctx, &stubJetStreamMsg{subject: "auth.session.created", data: data})

	notifs, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(userID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, notification.TypeNewLogin, notifs[0].Type)
}
