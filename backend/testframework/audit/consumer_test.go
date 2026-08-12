package audit_test

// Тесты NATS consumer аудита: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/audit"
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

	"github.com/valerii/hub/backend/ent"
)

// stubJetStreamMsg — минимальная реализация jetstream.Msg для unit-тестов processEvent.
type stubJetStreamMsg struct {
	subject string
	data    []byte
}

func (m *stubJetStreamMsg) Subject() string { return m.subject }
func (m *stubJetStreamMsg) Data() []byte    { return m.data }

func (m *stubJetStreamMsg) Headers() nats.Header { return nil }
func (m *stubJetStreamMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *stubJetStreamMsg) Ack() error { return nil }
func (m *stubJetStreamMsg) DoubleAck(context.Context) error { return nil }
func (m *stubJetStreamMsg) Nak() error { return nil }
func (m *stubJetStreamMsg) NakWithDelay(time.Duration) error { return nil }
func (m *stubJetStreamMsg) InProgress() error { return nil }
func (m *stubJetStreamMsg) Term() error { return nil }
func (m *stubJetStreamMsg) TermWithReason(string) error { return nil }
func (m *stubJetStreamMsg) Reply() string { return "" }

type consumerRepoStub struct {
	mockAuditRepo
	createCalls []createAuditLogCall
	createErr   error
}

func (s *consumerRepoStub) CreateAuditLog(_ context.Context, action, entityType, entityID string, userID, _ *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) (*ent.AuditLog, error) {
	s.createCalls = append(s.createCalls, createAuditLogCall{
		action: action, entityType: entityType, entityID: entityID,
		userID: userID, details: details, ipAddress: ipAddress, userAgent: userAgent,
	})
	if s.createErr != nil {
		return nil, s.createErr
	}
	return &ent.AuditLog{Action: action, EntityType: entityType, EntityID: entityID}, nil
}

// TestProcessEvent_CreatesAuditLog проверяет сценарий: process event creates audit log.
func TestProcessEvent_CreatesAuditLog(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	issueID := uuid.New().String()

	payload := map[string]interface{}{
		"id":         issueID,
		"user_id":    userID.String(),
		"project_id": projectID.String(),
		"title":      "Demo issue",
	}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	repo := &consumerRepoStub{}
	consumer := audit.ExportNewAuditConsumerForTest(repo)
	audit.ExportProcessEvent(consumer, context.Background(), &stubJetStreamMsg{
		subject: "tracker." + projectID.String() + ".issue.created",
		data:    raw,
	})

	require.Len(t, repo.createCalls, 1)
	call := repo.createCalls[0]
	assert.Equal(t, "create", call.action)
	assert.Equal(t, "issue", call.entityType)
	assert.Equal(t, issueID, call.entityID)
	require.NotNil(t, call.userID)
	assert.Equal(t, userID, *call.userID)
}

// TestProcessEvent_ExtractsAlternateIDFields проверяет сценарий: process event extracts alternate id fields.
func TestProcessEvent_ExtractsAlternateIDFields(t *testing.T) {
	channelID := uuid.New().String()
	senderID := uuid.New()

	payload := map[string]interface{}{
		"channel_id": channelID,
		"sender_id":  senderID.String(),
	}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	repo := &consumerRepoStub{}
	consumer := audit.ExportNewAuditConsumerForTest(repo)
	audit.ExportProcessEvent(consumer, context.Background(), &stubJetStreamMsg{
		subject: "messenger.channel.ch-1.message.sent",
		data:    raw,
	})

	require.Len(t, repo.createCalls, 1)
	assert.Equal(t, channelID, repo.createCalls[0].entityID)
	require.NotNil(t, repo.createCalls[0].userID)
	assert.Equal(t, senderID, *repo.createCalls[0].userID)
}

// TestProcessEvent_SkipsUnknownSubject проверяет сценарий: process event skips unknown subject.
func TestProcessEvent_SkipsUnknownSubject(t *testing.T) {
	repo := &consumerRepoStub{}
	consumer := audit.ExportNewAuditConsumerForTest(repo)
	audit.ExportProcessEvent(consumer, context.Background(), &stubJetStreamMsg{
		subject: "metrics.cpu.high",
		data:    []byte(`{"id":"x"}`),
	})
	assert.Empty(t, repo.createCalls)
}

// TestProcessEvent_InvalidJSON_NaksWithoutCreate — W6-A4-014 / P3-BE-02:
// прежнее имя SkipsInvalidJSON устарело; assert Ack/Nak через counting msg.
func TestProcessEvent_InvalidJSON_NaksWithoutCreate(t *testing.T) {
	repo := &consumerRepoStub{}
	consumer := audit.ExportNewAuditConsumerForTest(repo)
	msg := &countingAuditJSMsg{
		subject: "project.created",
		data:    []byte(`{not-json`),
	}
	audit.ExportProcessEvent(consumer, context.Background(), msg)
	assert.Empty(t, repo.createCalls)
	assert.Equal(t, 1, msg.naks, "invalid JSON must Nak")
	assert.Equal(t, 0, msg.acks)
}

// TestProcessEvent_RepoErrorDoesNotPanic гарантирует отсутствие паники/гонок при process event repo ошибка does not паника.
func TestProcessEvent_RepoErrorDoesNotPanic(t *testing.T) {
	repo := &consumerRepoStub{createErr: errors.New("db down")}
	consumer := audit.ExportNewAuditConsumerForTest(repo)

	assert.NotPanics(t, func() {
		audit.ExportProcessEvent(consumer, context.Background(), &stubJetStreamMsg{
			subject: "project.created",
			data:    []byte(`{"id":"proj-1"}`),
		})
	})
	require.Len(t, repo.createCalls, 1)
}
