package tracker_test

// Тесты NATS consumer трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

type stubAutomationJSMsg struct {
	subject string
	data    []byte
}

func (m *stubAutomationJSMsg) Subject() string                         { return m.subject }
func (m *stubAutomationJSMsg) Data() []byte                            { return m.data }
func (m *stubAutomationJSMsg) Headers() nats.Header                    { return nil }
func (m *stubAutomationJSMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *stubAutomationJSMsg) Ack() error                              { return nil }
func (m *stubAutomationJSMsg) DoubleAck(context.Context) error         { return nil }
func (m *stubAutomationJSMsg) Nak() error                              { return nil }
func (m *stubAutomationJSMsg) NakWithDelay(time.Duration) error        { return nil }
func (m *stubAutomationJSMsg) InProgress() error                       { return nil }
func (m *stubAutomationJSMsg) Term() error                             { return nil }
func (m *stubAutomationJSMsg) TermWithReason(string) error             { return nil }
func (m *stubAutomationJSMsg) Reply() string                           { return "" }

func startAutomationJetStream(t *testing.T) *natsclient.Client {
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
	require.NoError(t, client.EnsureStreams(context.Background()))
	return client
}

// TestParseIssueEventKind проверяет сценарий: разбор issue event kind.
func TestParseIssueEventKind(t *testing.T) {
	pid := uuid.New().String()
	assert.Equal(t, "moved", tracker.ExportParseIssueEventKind("project."+pid+".issue.moved"))
	assert.Equal(t, "created", tracker.ExportParseIssueEventKind("project."+pid+".issue.created"))
	assert.Equal(t, "assigned", tracker.ExportParseIssueEventKind("project."+pid+".issue.assigned"))
	assert.Equal(t, "", tracker.ExportParseIssueEventKind("other.subject"))
	assert.Equal(t, "", tracker.ExportParseIssueEventKind("project."+pid+".task.moved"))
}

// TestAutomationConsumer_StartStop_NilDeps проверяет сценарий: автоматизация consumer start stop nil deps.
func TestAutomationConsumer_StartStop_NilDeps(t *testing.T) {
	c := tracker.NewAutomationConsumer(nil, nil)
	require.NoError(t, c.Start(context.Background()))
	c.Stop()
}

// TestAutomationConsumer_StartStop_WithJetStream проверяет сценарий: автоматизация consumer start stop with jet stream.
func TestAutomationConsumer_StartStop_WithJetStream(t *testing.T) {
	nc := startAutomationJetStream(t)
	repo := newMockRepo()
	engine := tracker.NewAutomationEngine(repo, nil, nil)
	c := tracker.NewAutomationConsumer(nc, engine)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, c.Start(ctx))
	time.Sleep(100 * time.Millisecond)
	c.Stop()
}

// TestAutomationConsumer_processMessage_Events проверяет сценарий: автоматизация consumer process message events.
func TestAutomationConsumer_processMessage_Events(t *testing.T) {
	repo := newMockRepo()
	engine := tracker.NewAutomationEngine(repo, nil, nil)
	c := tracker.NewAutomationConsumer(nil, engine)
	ctx := context.Background()
	pid := uuid.New().String()

	moved, err := json.Marshal(tracker.IssueMovedEvent{
		IssueID: uuid.New().String(), IssueKey: "HUB-1", ProjectID: pid,
	})
	require.NoError(t, err)
	tracker.ExportAutomationConsumerProcessMessage(c, ctx, &stubAutomationJSMsg{
		subject: "project." + pid + ".issue.moved",
		data:    moved,
	})

	created, err := json.Marshal(tracker.IssueCreatedEvent{
		IssueID: uuid.New().String(), IssueKey: "HUB-2", ProjectID: pid,
	})
	require.NoError(t, err)
	tracker.ExportAutomationConsumerProcessMessage(c, ctx, &stubAutomationJSMsg{
		subject: "project." + pid + ".issue.created",
		data:    created,
	})

	assigned, err := json.Marshal(tracker.IssueAssignedEvent{
		IssueID: uuid.New().String(), IssueKey: "HUB-3", ProjectID: pid,
		AssigneeID: uuid.New().String(),
	})
	require.NoError(t, err)
	tracker.ExportAutomationConsumerProcessMessage(c, ctx, &stubAutomationJSMsg{
		subject: "project." + pid + ".issue.assigned",
		data:    assigned,
	})

	tracker.ExportAutomationConsumerProcessMessage(c, ctx, &stubAutomationJSMsg{
		subject: "project." + pid + ".issue.unknown",
		data:    []byte(`{}`),
	})

	tracker.ExportAutomationConsumerProcessMessage(c, ctx, &stubAutomationJSMsg{
		subject: "project." + pid + ".issue.moved",
		data:    []byte(`not-json`),
	})
}
