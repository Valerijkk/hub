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

// DeleteLabel удаляет метку по ID только если она принадлежит проекту (Block F IDOR).
func (r *EntRepository) DeleteLabel(ctx context.Context, projectID, labelID uuid.UUID) error {
	affected, err := r.client.Label.Delete().
		Where(
			label.IDEQ(labelID),
			label.HasProjectWith(project.IDEQ(projectID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// CountLabelsInProject считает, сколько из labelIDs принадлежат проекту (P1-BE-03).
func (r *EntRepository) CountLabelsInProject(ctx context.Context, projectID uuid.UUID, labelIDs []uuid.UUID) (int, error) {
	if len(labelIDs) == 0 {
		return 0, nil
	}
	return r.client.Label.Query().
		Where(
			label.IDIn(labelIDs...),
			label.HasProjectWith(project.IDEQ(projectID)),
		).
		Count(ctx)
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
