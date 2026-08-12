package tracker_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
)

type countingAutomationJSMsg struct {
	subject string
	data    []byte
	acks    int
	naks    int
}

func (m *countingAutomationJSMsg) Subject() string                           { return m.subject }
func (m *countingAutomationJSMsg) Data() []byte                              { return m.data }
func (m *countingAutomationJSMsg) Headers() nats.Header                      { return nil }
func (m *countingAutomationJSMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *countingAutomationJSMsg) Ack() error                                { m.acks++; return nil }
func (m *countingAutomationJSMsg) DoubleAck(context.Context) error           { return nil }
func (m *countingAutomationJSMsg) Nak() error                                { m.naks++; return nil }
func (m *countingAutomationJSMsg) NakWithDelay(time.Duration) error {
	m.naks++
	return nil
}
func (m *countingAutomationJSMsg) InProgress() error             { return nil }
func (m *countingAutomationJSMsg) Term() error                   { return nil }
func (m *countingAutomationJSMsg) TermWithReason(string) error   { return nil }
func (m *countingAutomationJSMsg) Reply() string                 { return "" }

// TestAutomationConsumer_NakOnBadJSON_A3_006 — битый JSON → Nak, не Ack.
// Mutation: всегда Ack → naks==0 FAIL.
func TestAutomationConsumer_NakOnBadJSON_A3_006(t *testing.T) {
	repo := newMockRepo()
	engine := tracker.NewAutomationEngine(repo, nil, nil)
	c := tracker.NewAutomationConsumer(nil, engine)
	pid := uuid.New().String()

	bad := &countingAutomationJSMsg{
		subject: "project." + pid + ".issue.moved",
		data:    []byte(`not-json`),
	}
	tracker.ExportAutomationConsumerProcessMessage(c, context.Background(), bad)
	assert.Equal(t, 1, bad.naks, "corrupt payload must Nak")
	assert.Equal(t, 0, bad.acks, "corrupt payload must not Ack")

	okPayload, err := json.Marshal(tracker.IssueMovedEvent{
		IssueID: uuid.New().String(), IssueKey: "HUB-1", ProjectID: pid,
	})
	require.NoError(t, err)
	good := &countingAutomationJSMsg{
		subject: "project." + pid + ".issue.moved",
		data:    okPayload,
	}
	tracker.ExportAutomationConsumerProcessMessage(c, context.Background(), good)
	assert.Equal(t, 1, good.acks)
	assert.Equal(t, 0, good.naks)
}
