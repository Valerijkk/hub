package notification_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/notification"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestNotificationConsumer_NakOnBadJSON_A2_010 — битый JSON → Nak, не Ack.
// Mutation: всегда Ack после handler → naked==false FAIL.
func TestNotificationConsumer_NakOnBadJSON_A2_010(t *testing.T) {
	c := notification.NewConsumer(nil, nil, nil)
	msg := &stubJSMsg{
		subject: "project." + uuid.New().String() + ".issue.assigned",
		data:    []byte(`not-json`),
	}
	notification.ExportProcessEvent(c, context.Background(), msg)
	assert.True(t, msg.naked, "corrupt payload must Nak")
	assert.False(t, msg.acked, "corrupt payload must not Ack")
}

// TestNotificationConsumer_AckOnSuccess_A2_010 — успешный Dispatch → Ack.
func TestNotificationConsumer_AckOnSuccess_A2_010(t *testing.T) {
	svc, repo, ctx, assigneeID := newNotificationTestEnv(t)
	db := notification.ExportRepositoryDB(repo)
	actor, err := db.User.Create().
		SetEmail("ack-actor-" + uuid.New().String()[:8] + "@test.com").
		SetName("Actor").
		Save(ctx)
	require.NoError(t, err)

	c := notification.NewConsumer(nil, repo, svc)
	payload, err := json.Marshal(tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-ACK",
		ProjectID:    uuid.New().String(),
		AssigneeID:   assigneeID.String(),
		AssignedByID: actor.ID.String(),
	})
	require.NoError(t, err)

	msg := &stubJSMsg{
		subject: "project." + uuid.New().String() + ".issue.assigned",
		data:    payload,
	}
	notification.ExportProcessEvent(c, ctx, msg)
	assert.True(t, msg.acked)
	assert.False(t, msg.naked)
}

// TestNotificationConsumer_NakOnDispatchError_A2_010 — handler/Dispatch err → Nak.
// Valid JSON reaches handleAssigned; missing assignee FK makes Dispatch fail.
// Mutation: always Ack after handler (ignore processErr) → naked==false FAIL.
// Bad-JSON-only coverage is WEAK — this proves Nak on Dispatch failure path.
func TestNotificationConsumer_NakOnDispatchError_A2_010(t *testing.T) {
	svc, repo, ctx, actorID := newNotificationTestEnv(t)
	missingAssignee := uuid.New()

	c := notification.NewConsumer(nil, repo, svc)
	payload, err := json.Marshal(tracker.IssueAssignedEvent{
		IssueID:      uuid.New().String(),
		IssueKey:     "HUB-NAK-DISP",
		ProjectID:    uuid.New().String(),
		AssigneeID:   missingAssignee.String(),
		AssignedByID: actorID.String(),
	})
	require.NoError(t, err)

	msg := &stubJSMsg{
		subject: "project." + uuid.New().String() + ".issue.assigned",
		data:    payload,
	}
	notification.ExportProcessEvent(c, ctx, msg)
	assert.True(t, msg.naked, "Dispatch failure must Nak for MaxDeliver retry")
	assert.False(t, msg.acked, "Dispatch failure must not Ack")
}
