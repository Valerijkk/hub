package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/activitylog"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/schema"
)

// ---------------------------------------------------------------------------
// Activity log
// ---------------------------------------------------------------------------

// LogActivity записывает запись в лог активности задачи.
func (r *EntRepository) LogActivity(ctx context.Context, issueID, userID uuid.UUID, action string, changes []FieldChange) (*ent.ActivityLog, error) {
	// Конвертируем FieldChange из пакета tracker в schema.FieldChange
	schemaChanges := make([]schema.FieldChange, len(changes))
	for i, c := range changes {
		schemaChanges[i] = schema.FieldChange{
			Field:    c.Field,
			OldValue: c.OldValue,
			NewValue: c.NewValue,
		}
	}

	create := r.client.ActivityLog.Create().
		SetIssueID(issueID).
		SetUserID(userID).
		SetAction(action)

	if len(schemaChanges) > 0 {
		create = create.SetChanges(schemaChanges)
	}

	return create.Save(ctx)
}

// ListActivities возвращает все записи лога активности задачи.
func (r *EntRepository) ListActivities(ctx context.Context, issueID uuid.UUID) ([]*ent.ActivityLog, error) {
	return r.client.ActivityLog.Query().
		Where(activitylog.HasIssueWith(issue.IDEQ(issueID))).
		WithUser().
		Order(ent.Desc(activitylog.FieldCreatedAt)).
		All(ctx)
}
