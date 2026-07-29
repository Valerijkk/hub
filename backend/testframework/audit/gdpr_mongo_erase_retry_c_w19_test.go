//go:build hubtest

package audit_test

// TASK-178 / C-W19-002: post-Commit Mongo erase fail → JetStream compensating retry.
import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
)

type stubEraseRetry struct {
	mu    sync.Mutex
	calls []uuid.UUID
	err   error
}

func (s *stubEraseRetry) EnqueueMongoErase(_ context.Context, senderID, _ uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, senderID)
	return s.err
}

// countingEraseJSMsg — jetstream.Msg stub с подсчётом Ack/Nak/Term.
type countingEraseJSMsg struct {
	subject string
	data    []byte
	acked   bool
	naked   bool
	termed  bool
}

func (m *countingEraseJSMsg) Subject() string                           { return m.subject }
func (m *countingEraseJSMsg) Data() []byte                              { return m.data }
func (m *countingEraseJSMsg) Headers() nats.Header                      { return nil }
func (m *countingEraseJSMsg) Metadata() (*jetstream.MsgMetadata, error) { return &jetstream.MsgMetadata{}, nil }
func (m *countingEraseJSMsg) Reply() string                             { return "" }
func (m *countingEraseJSMsg) Ack() error                                { m.acked = true; return nil }
func (m *countingEraseJSMsg) DoubleAck(context.Context) error           { return m.Ack() }
func (m *countingEraseJSMsg) Nak() error                                { m.naked = true; return nil }
func (m *countingEraseJSMsg) NakWithDelay(time.Duration) error          { return m.Nak() }
func (m *countingEraseJSMsg) InProgress() error                         { return nil }
func (m *countingEraseJSMsg) Term() error                               { m.termed = true; return nil }
func (m *countingEraseJSMsg) TermWithReason(string) error               { return m.Term() }

// TestDeleteUserGDPR_MongoEraseFail_EnqueuesRetry_CW19002 —
// C-W19-002: HardDelete уже Commit; fail eraser → EnqueueMongoErase(sender_id);
// admin всё ещё получает error.
// Mutation: убрать вызов eraseRetry.EnqueueMongoErase в DeleteUserGDPR → тест краснеет.
func TestDeleteUserGDPR_MongoEraseFail_EnqueuesRetry_CW19002(t *testing.T) {
	repo := &gdprFullCascadeRepo{}
	eraser := &stubMessageEraser{failErr: errGDPRCascadeStep}
	retry := &stubEraseRetry{}
	svc := audit.NewAuditService(repo, nil)
	svc.SetMessageEraser(eraser)
	svc.SetMongoEraseRetry(retry)

	target := uuid.New()
	actor := uuid.New()
	err := svc.DeleteUserGDPR(context.Background(), actor, target, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "delete mongo messages")

	require.Len(t, repo.hardDeleted, 1, "Ent HardDelete already committed")
	assert.Equal(t, target, repo.hardDeleted[0])
	require.Len(t, eraser.calls, 1)
	require.Len(t, retry.calls, 1, "compensating retry must enqueue on eraser fail")
	assert.Equal(t, target, retry.calls[0])
}

// TestDeleteUserGDPR_MongoEraseOK_NoRetry_CW19002 — happy path не ставит retry.
func TestDeleteUserGDPR_MongoEraseOK_NoRetry_CW19002(t *testing.T) {
	repo := &gdprFullCascadeRepo{}
	eraser := &stubMessageEraser{}
	retry := &stubEraseRetry{}
	svc := audit.NewAuditService(repo, nil)
	svc.SetMessageEraser(eraser)
	svc.SetMongoEraseRetry(retry)

	require.NoError(t, svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil))
	assert.Empty(t, retry.calls)
}

// TestDeleteUserGDPR_MongoEraseFail_EnqueueAlsoFails_CW19002 —
// оба fail → error содержит enqueue; HardDelete всё равно committed.
func TestDeleteUserGDPR_MongoEraseFail_EnqueueAlsoFails_CW19002(t *testing.T) {
	repo := &gdprFullCascadeRepo{}
	eraser := &stubMessageEraser{failErr: errGDPRCascadeStep}
	retry := &stubEraseRetry{err: errors.New("nats down")}
	svc := audit.NewAuditService(repo, nil)
	svc.SetMessageEraser(eraser)
	svc.SetMongoEraseRetry(retry)

	err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "enqueue retry")
	require.Len(t, repo.hardDeleted, 1)
}

// flakyEraser — первый вызов fail, второй ok (идемпотентный dual erase / retry).
type flakyEraser struct {
	calls int
	ids   []uuid.UUID
}

func (e *flakyEraser) DeleteMessagesBySender(_ context.Context, senderID uuid.UUID) (int64, error) {
	e.calls++
	e.ids = append(e.ids, senderID)
	if e.calls == 1 {
		return 0, errGDPRCascadeStep
	}
	return 2, nil
}

// TestMongoEraseRetryConsumer_ProcessMsg_AckNak_CW19002 —
// consumer: fail → Nak; success → Ack; dual call идемпотентен по sender_id.
func TestMongoEraseRetryConsumer_ProcessMsg_AckNak_CW19002(t *testing.T) {
	sender := uuid.New()
	payload, err := json.Marshal(audit.MongoEraseRetryEvent{SenderID: sender})
	require.NoError(t, err)

	eraser := &flakyEraser{}
	cons := audit.NewMongoEraseRetryConsumer(nil, eraser)

	failMsg := &countingEraseJSMsg{subject: audit.SubjectGDPRMongoErase, data: payload}
	audit.ExportMongoEraseRetryProcessMsg(cons, context.Background(), failMsg)
	assert.True(t, failMsg.naked)
	assert.False(t, failMsg.acked)
	assert.Equal(t, 1, eraser.calls)

	okMsg := &countingEraseJSMsg{subject: audit.SubjectGDPRMongoErase, data: payload}
	audit.ExportMongoEraseRetryProcessMsg(cons, context.Background(), okMsg)
	assert.True(t, okMsg.acked)
	assert.False(t, okMsg.naked)
	assert.Equal(t, 2, eraser.calls)
	assert.Equal(t, []uuid.UUID{sender, sender}, eraser.ids)
}

// TestMongoEraseRetryConsumer_BadPayload_Term_CW19002 — poison → Term, не Nak loop.
func TestMongoEraseRetryConsumer_BadPayload_Term_CW19002(t *testing.T) {
	eraser := &stubMessageEraser{}
	cons := audit.NewMongoEraseRetryConsumer(nil, eraser)
	msg := &countingEraseJSMsg{subject: audit.SubjectGDPRMongoErase, data: []byte(`{not-json`)}
	audit.ExportMongoEraseRetryProcessMsg(cons, context.Background(), msg)
	assert.True(t, msg.termed)
	assert.Empty(t, eraser.calls)
}

// TestSubjectGDPRMongoErase_Canonical — mutation: смена subject без stream/wiring → drift.
func TestSubjectGDPRMongoErase_Canonical(t *testing.T) {
	assert.Equal(t, "audit.gdpr.mongo_erase", audit.SubjectGDPRMongoErase)
}
