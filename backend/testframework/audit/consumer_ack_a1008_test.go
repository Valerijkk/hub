package audit_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/audit"
)

// countingAuditJSMsg считает Ack/Nak для mutation-proof A1-008.
type countingAuditJSMsg struct {
	subject string
	data    []byte
	acks    int
	naks    int
}

func (m *countingAuditJSMsg) Subject() string                           { return m.subject }
func (m *countingAuditJSMsg) Data() []byte                              { return m.data }
func (m *countingAuditJSMsg) Headers() nats.Header                      { return nil }
func (m *countingAuditJSMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *countingAuditJSMsg) Ack() error                                { m.acks++; return nil }
func (m *countingAuditJSMsg) DoubleAck(context.Context) error           { return nil }
func (m *countingAuditJSMsg) Nak() error                                { m.naks++; return nil }
func (m *countingAuditJSMsg) NakWithDelay(time.Duration) error {
	m.naks++
	return nil
}
func (m *countingAuditJSMsg) InProgress() error           { return nil }
func (m *countingAuditJSMsg) Term() error                 { return nil }
func (m *countingAuditJSMsg) TermWithReason(string) error { return nil }
func (m *countingAuditJSMsg) Reply() string               { return "" }

// TestAuditConsumer_NakOnCreateError_A1_008 — CreateAuditLog err → Nak, не Ack.
// Mutation: всегда Ack (игнорировать err) → naks==0 FAIL.
func TestAuditConsumer_NakOnCreateError_A1_008(t *testing.T) {
	repo := &consumerRepoStub{createErr: errors.New("db down")}
	consumer := audit.ExportNewAuditConsumerForTest(repo)

	msg := &countingAuditJSMsg{
		subject: "project.created",
		data:    []byte(`{"id":"proj-1"}`),
	}
	audit.ExportProcessEvent(consumer, context.Background(), msg)
	assert.Equal(t, 1, msg.naks, "CreateAuditLog failure must Nak")
	assert.Equal(t, 0, msg.acks, "CreateAuditLog failure must not Ack")
	require.Len(t, repo.createCalls, 1)
}

// TestAuditConsumer_AckOnSuccess_A1_008 — успешная запись → Ack.
func TestAuditConsumer_AckOnSuccess_A1_008(t *testing.T) {
	repo := &consumerRepoStub{}
	consumer := audit.ExportNewAuditConsumerForTest(repo)

	payload, err := json.Marshal(map[string]interface{}{
		"id":      uuid.New().String(),
		"user_id": uuid.New().String(),
	})
	require.NoError(t, err)

	msg := &countingAuditJSMsg{
		subject: "project.created",
		data:    payload,
	}
	audit.ExportProcessEvent(consumer, context.Background(), msg)
	assert.Equal(t, 1, msg.acks)
	assert.Equal(t, 0, msg.naks)
}

// TestAuditConsumer_NakOnBadJSON_A1_008 — битый JSON → Nak.
func TestAuditConsumer_NakOnBadJSON_A1_008(t *testing.T) {
	repo := &consumerRepoStub{}
	consumer := audit.ExportNewAuditConsumerForTest(repo)

	msg := &countingAuditJSMsg{
		subject: "project.created",
		data:    []byte(`{not-json`),
	}
	audit.ExportProcessEvent(consumer, context.Background(), msg)
	assert.Equal(t, 1, msg.naks)
	assert.Equal(t, 0, msg.acks)
	assert.Empty(t, repo.createCalls)
}
