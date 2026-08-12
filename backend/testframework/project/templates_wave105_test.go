package project_test

// Тесты модуля проектов (волна 105): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/projecttemplate"
)

// TestTemplateService_GetTemplate_NotFound проверяет бизнес-логику: шаблон сервис получение шаблон not found.
func TestTemplateService_GetTemplate_NotFound(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)

	_, err := svc.GetTemplate(context.Background(), uuid.New())
	require.Error(t, err)
}

// TestTemplateService_DeleteTemplate_CustomSuccess проверяет бизнес-логику: шаблон сервис удаление шаблон custom успех.
func TestTemplateService_DeleteTemplate_CustomSuccess(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	authorID := uuid.New()
	customID := uuid.New()
	repo.templates[customID] = &ent.ProjectTemplate{
		ID: customID, Name: "Custom", IsSystem: false, CreatedAt: time.Now(),
		Edges: ent.ProjectTemplateEdges{CreatedBy: &ent.User{ID: authorID}},
	}

	require.NoError(t, svc.DeleteTemplate(ctx, customID, authorID))
	_, ok := repo.templates[customID]
	assert.False(t, ok)
}

// TestTemplateService_DeleteTemplate_NotFound проверяет бизнес-логику: шаблон сервис удаление шаблон not found.
func TestTemplateService_DeleteTemplate_NotFound(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)

	err := svc.DeleteTemplate(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
}

// TestTemplateService_GetTemplate_CustomType проверяет бизнес-логику: шаблон сервис получение шаблон custom type.
func TestTemplateService_GetTemplate_CustomType(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	customID := uuid.New()
	repo.templates[customID] = &ent.ProjectTemplate{
		ID: customID, Name: "Scrum", TemplateType: projecttemplate.TemplateTypeCustom,
		Config: map[string]interface{}{"statuses": []string{"Todo"}}, IsSystem: false, CreatedAt: time.Now(),
	}

	got, err := svc.GetTemplate(ctx, customID)
	require.NoError(t, err)
	assert.Equal(t, "custom", got.TemplateType)
}
