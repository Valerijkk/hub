package notification_test

// Тесты NATS consumer уведомлений (волна 99): интеграционные и unit-сценарии testframework.

import (
	"encoding/json"
	"testing"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	entnotif "github.com/valerii/hub/backend/ent/notification"
)

// TestConsumer_handleBlockerAdded_SkipsWhenAssigneeIsActor проверяет сценарий: consumer handle blocker added skips when assignee is actor.
func TestConsumer_handleBlockerAdded_SkipsWhenAssigneeIsActor(t *testing.T) {
	c, repo, ctx, assignee, actor, issueEnt, projectID := newConsumerIssueEnv(t)

	// Назначаем задачу актору, чтобы путь assignee == actor завершался досрочно.
	_, err := issueEnt.Update().SetAssigneeID(actor.ID).Save(ctx)
	require.NoError(t, err)

	evt := struct {
		SourceID  string `json:"source_id"`
		TargetID  string `json:"target_id"`
		LinkType  string `json:"link_type"`
		ProjectID string `json:"project_id"`
		CreatedBy string `json:"created_by"`
	}{
		SourceID:  uuid.New().String(),
		TargetID:  issueEnt.ID.String(),
		LinkType:  "blocks",
		ProjectID: projectID.String(),
		CreatedBy: actor.ID.String(),
	}
	data, err := json.Marshal(evt)
	require.NoError(t, err)

	notification.ExportHandleBlockerAdded(c, ctx, data, "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
}

// TestConsumer_handleBlockerAdded_InvalidJSON проверяет отклонение при consumer handle blocker added некорректный json.
func TestConsumer_handleBlockerAdded_InvalidJSON(t *testing.T) {
	c, repo, ctx, assignee, _, _, _ := newConsumerIssueEnv(t)
	notification.ExportHandleBlockerAdded(c, ctx, []byte("{"), "")

	count, err := notification.ExportRepositoryDB(repo).Notification.Query().
		Where(entnotif.UserID(assignee.ID)).
		Count(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
}
