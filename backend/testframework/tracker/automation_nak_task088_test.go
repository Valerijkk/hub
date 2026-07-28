package tracker_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

type failListRulesRepo struct {
	mockRepo
}

func (r *failListRulesRepo) ListEnabledAutomationRules(context.Context, uuid.UUID, string) ([]*ent.AutomationRule, error) {
	return nil, errors.New("db unavailable")
}

// TestAutomationConsumer_NakOnEngineError_A2W1_A3_001 — transient ListRules → Nak.
// Mutation: Handle* глотает err / consumer всегда Ack → naks==0 FAIL.
func TestAutomationConsumer_NakOnEngineError_A2W1_A3_001(t *testing.T) {
	repo := &failListRulesRepo{mockRepo: *newMockRepo()}
	engine := tracker.NewAutomationEngine(repo, nil, nil)
	c := tracker.NewAutomationConsumer(nil, engine)
	pid := uuid.New().String()

	payload, err := json.Marshal(tracker.IssueMovedEvent{
		IssueID: uuid.New().String(), IssueKey: "HUB-1", ProjectID: pid,
		NewStatusID: uuid.New().String(),
	})
	require.NoError(t, err)

	msg := &countingAutomationJSMsg{
		subject: "project." + pid + ".issue.moved",
		data:    payload,
	}
	tracker.ExportAutomationConsumerProcessMessage(c, context.Background(), msg)
	assert.Equal(t, 1, msg.naks, "engine list-rules error must Nak")
	assert.Equal(t, 0, msg.acks)
}

// TestAutomationConsumer_NakOnUnknownSubject_A2W1_A3_002 — unknown subject → Nak + metric.
// Mutation: default → Ack без счётчика → naks==0 / metric не растёт FAIL.
func TestAutomationConsumer_NakOnUnknownSubject_A2W1_A3_002(t *testing.T) {
	repo := newMockRepo()
	engine := tracker.NewAutomationEngine(repo, nil, nil)
	c := tracker.NewAutomationConsumer(nil, engine)

	before := tracker.AutomationUnknownSubjectTotal()
	msg := &countingAutomationJSMsg{
		subject: "project." + uuid.New().String() + ".issue.deleted",
		data:    []byte(`{}`),
	}
	tracker.ExportAutomationConsumerProcessMessage(c, context.Background(), msg)
	assert.Equal(t, 1, msg.naks, "unknown subject must Nak")
	assert.Equal(t, 0, msg.acks)
	assert.Equal(t, before+1, tracker.AutomationUnknownSubjectTotal())

	bad := &countingAutomationJSMsg{
		subject: "not.a.valid.subject",
		data:    []byte(`{}`),
	}
	tracker.ExportAutomationConsumerProcessMessage(c, context.Background(), bad)
	assert.Equal(t, 1, bad.naks)
	assert.Equal(t, 0, bad.acks)
}
