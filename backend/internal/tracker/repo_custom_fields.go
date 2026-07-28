package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/customfielddefinition"
	"github.com/valerii/hub/backend/ent/project"
)

// ---------------------------------------------------------------------------
// Кастомные поля
// ---------------------------------------------------------------------------

// CreateCustomField создаёт новое определение кастомного поля в проекте.
func (r *EntRepository) CreateCustomField(ctx context.Context, projectID uuid.UUID, input CreateCustomFieldInput) (*ent.CustomFieldDefinition, error) {
	create := r.client.CustomFieldDefinition.Create().
		SetProjectID(projectID).
		SetName(input.Name).
		SetFieldType(customfielddefinition.FieldType(input.FieldType)).
		SetRequired(input.Required)

	if len(input.Options) > 0 {
		create = create.SetOptions(input.Options)
	}

	return create.Save(ctx)
}

// ListCustomFields возвращает все определения кастомных полей проекта.
func (r *EntRepository) ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]*ent.CustomFieldDefinition, error) {
	return r.client.CustomFieldDefinition.Query().
		Where(customfielddefinition.HasProjectWith(project.IDEQ(projectID))).
		Order(ent.Asc(customfielddefinition.FieldSortOrder)).
		All(ctx)
}

// DeleteCustomField удаляет определение кастомного поля только внутри projectID (P1-BE-02).
func (r *EntRepository) DeleteCustomField(ctx context.Context, projectID, fieldID uuid.UUID) error {
	affected, err := r.client.CustomFieldDefinition.Delete().
		Where(
			customfielddefinition.IDEQ(fieldID),
			customfielddefinition.HasProjectWith(project.IDEQ(projectID)),
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
