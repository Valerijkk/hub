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

// TestNotificationConsumer_NakOnRepoErr_P1_BE_35 — transient repo/DB err → Nak, не Ack.
// Mutation: soft `return nil` на repo err → naked==false FAIL.
func TestNotificationConsumer_NakOnRepoErr_P1_BE_35(t *testing.T) {
	svc, repo, ctx, actorID := newNotificationTestEnv(t)
	db := notification.ExportRepositoryDB(repo)

	// Закрываем DB — последующий GetIssueAssignee вернёт transient error.
	require.NoError(t, db.Close())

	c := notification.NewConsumer(nil, repo, svc)
	payload, err := json.Marshal(tracker.IssueMovedEvent{
		IssueID:   uuid.New().String(),
		IssueKey:  "HUB-NAK-REPO",
		ProjectID: uuid.New().String(),
		MovedByID: actorID.String(),
	})
	require.NoError(t, err)

	msg := &stubJSMsg{
		subject: "project." + uuid.New().String() + ".issue.moved",
		data:    payload,
	}
	notification.ExportProcessEvent(c, ctx, msg)
	assert.True(t, msg.naked, "repo/DB error must Nak for MaxDeliver retry")
	assert.False(t, msg.acked, "repo/DB error must not Ack")
}

// TestNotificationConsumer_AckOnMissingIssue_P1_BE_35 — not-found issue → Ack (не retry).
func TestNotificationConsumer_AckOnMissingIssue_P1_BE_35(t *testing.T) {
	svc, repo, _, actorID := newNotificationTestEnv(t)

	c := notification.NewConsumer(nil, repo, svc)
	payload, err := json.Marshal(tracker.IssueMovedEvent{
		IssueID:   uuid.New().String(), // не существует
		IssueKey:  "HUB-GONE",
		ProjectID: uuid.New().String(),
		MovedByID: actorID.String(),
	})
	require.NoError(t, err)

	msg := &stubJSMsg{
		subject: "project." + uuid.New().String() + ".issue.moved",
		data:    payload,
	}
	notification.ExportProcessEvent(c, context.Background(), msg)
	assert.True(t, msg.acked, "not-found must Ack")
	assert.False(t, msg.naked)
}
