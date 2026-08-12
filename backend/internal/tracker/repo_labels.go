package tracker

import (
	"context"
	"time"

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

// ListLabels возвращает все метки проекта (без корзины).
// A2W8-3-005: DeletedAtIsNil. Mutation: снять фильтр → soft-deleted в списке → FAIL.
func (r *EntRepository) ListLabels(ctx context.Context, projectID uuid.UUID) ([]*ent.Label, error) {
	return r.client.Label.Query().
		Where(
			label.HasProjectWith(project.IDEQ(projectID)),
			label.DeletedAtIsNil(),
		).
		WithGroup().
		All(ctx)
}

// DeleteLabel мягко удаляет метку по ID только если она принадлежит проекту (Block F IDOR).
// A2W8-3-005 / ADR-ENT-SOFT-DELETE-GAPS-2026: SetDeletedAt, не hard Delete.
func (r *EntRepository) DeleteLabel(ctx context.Context, projectID, labelID uuid.UUID) error {
	affected, err := r.client.Label.Update().
		Where(
			label.IDEQ(labelID),
			label.HasProjectWith(project.IDEQ(projectID)),
			label.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
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
			label.DeletedAtIsNil(),
		).
		Count(ctx)
}

// FindLabelGroupInProject возвращает группу меток только внутри projectID (A3W1-2-002).
func (r *EntRepository) FindLabelGroupInProject(ctx context.Context, projectID, groupID uuid.UUID) (*ent.LabelGroup, error) {
	return r.client.LabelGroup.Query().
		Where(
			labelgroup.IDEQ(groupID),
			labelgroup.HasProjectWith(project.IDEQ(projectID)),
			labelgroup.DeletedAtIsNil(),
		).
		Only(ctx)
}

// CreateLabelGroup создаёт новую группу меток в проекте.
func (r *EntRepository) CreateLabelGroup(ctx context.Context, projectID uuid.UUID, name string) (*ent.LabelGroup, error) {
	return r.client.LabelGroup.Create().
		SetProjectID(projectID).
		SetName(name).
		Save(ctx)
}

// ListLabelGroups возвращает все группы меток проекта с их метками (без корзины).
func (r *EntRepository) ListLabelGroups(ctx context.Context, projectID uuid.UUID) ([]*ent.LabelGroup, error) {
	return r.client.LabelGroup.Query().
		Where(
			labelgroup.HasProjectWith(project.IDEQ(projectID)),
			labelgroup.DeletedAtIsNil(),
		).
		WithLabels(func(q *ent.LabelQuery) {
			q.Where(label.DeletedAtIsNil())
		}).
		All(ctx)
}
