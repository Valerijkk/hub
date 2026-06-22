package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/label"
	"github.com/valerii/hub/backend/ent/labelgroup"
	"github.com/valerii/hub/backend/ent/project"
)

// ---------------------------------------------------------------------------
// Метки
// ---------------------------------------------------------------------------

// CreateLabel создаёт новую метку в проекте, опционально привязывая к группе.
func (r *EntRepository) CreateLabel(ctx context.Context, projectID uuid.UUID, name, color string, groupID *uuid.UUID) (*ent.Label, error) {
	create := r.client.Label.Create().
		SetProjectID(projectID).
		SetName(name).
		SetColor(color)

	if groupID != nil {
		create = create.SetGroupID(*groupID)
	}

	return create.Save(ctx)
}

// ListLabels возвращает все метки проекта.
func (r *EntRepository) ListLabels(ctx context.Context, projectID uuid.UUID) ([]*ent.Label, error) {
	return r.client.Label.Query().
		Where(label.HasProjectWith(project.IDEQ(projectID))).
		WithGroup().
		All(ctx)
}

// DeleteLabel удаляет метку по ID.
func (r *EntRepository) DeleteLabel(ctx context.Context, labelID uuid.UUID) error {
	return r.client.Label.DeleteOneID(labelID).Exec(ctx)
}

// CreateLabelGroup создаёт новую группу меток в проекте.
func (r *EntRepository) CreateLabelGroup(ctx context.Context, projectID uuid.UUID, name string) (*ent.LabelGroup, error) {
	return r.client.LabelGroup.Create().
		SetProjectID(projectID).
		SetName(name).
		Save(ctx)
}

// ListLabelGroups возвращает все группы меток проекта с их метками.
func (r *EntRepository) ListLabelGroups(ctx context.Context, projectID uuid.UUID) ([]*ent.LabelGroup, error) {
	return r.client.LabelGroup.Query().
		Where(labelgroup.HasProjectWith(project.IDEQ(projectID))).
		WithLabels().
		All(ctx)
}
