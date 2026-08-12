package project_test

// Тесты модуля проектов: интеграционные и unit-сценарии testframework.
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

type templateMockRepo struct {
	mockRepo
	templates map[uuid.UUID]*ent.ProjectTemplate
}

func newTemplateMockRepo() *templateMockRepo {
	return &templateMockRepo{
		mockRepo:  *newMockRepo(),
		templates: make(map[uuid.UUID]*ent.ProjectTemplate),
	}
}

func (m *templateMockRepo) ListTemplates(_ context.Context) ([]*ent.ProjectTemplate, error) {
	out := make([]*ent.ProjectTemplate, 0, len(m.templates))
	for _, t := range m.templates {
		out = append(out, t)
	}
	return out, nil
}

func (m *templateMockRepo) GetTemplate(_ context.Context, id uuid.UUID) (*ent.ProjectTemplate, error) {
	t, ok := m.templates[id]
	if !ok {
		return nil, notFoundErr()
	}
	return t, nil
}

func (m *templateMockRepo) GetProjectWithWorkflowStatusesAndLabels(_ context.Context, _ uuid.UUID) (*ent.Project, []string, []string, error) {
	return &ent.Project{ID: uuid.New(), Name: "P"}, []string{"Open", "Done"}, []string{"bug"}, nil
}

func (m *templateMockRepo) CreateTemplate(_ context.Context, name, description string, templateType projecttemplate.TemplateType, config map[string]interface{}, createdBy *uuid.UUID, isSystem bool) (*ent.ProjectTemplate, error) {
	id := uuid.New()
	t := &ent.ProjectTemplate{
		ID:           id,
		Name:         name,
		Description:  description,
		TemplateType: templateType,
		Config:       config,
		IsSystem:     isSystem,
		CreatedAt:    time.Now(),
	}
	if createdBy != nil {
		t.Edges.CreatedBy = &ent.User{ID: *createdBy}
	}
	m.templates[id] = t
	return t, nil
}

func (m *templateMockRepo) DeleteTemplate(_ context.Context, id uuid.UUID) error {
	delete(m.templates, id)
	return nil
}

func (m *templateMockRepo) CountSystemTemplates(_ context.Context) (int, error) {
	count := 0
	for _, t := range m.templates {
		if t.IsSystem {
			count++
		}
	}
	return count, nil
}

// TestTemplateService_ListAndGet проверяет бизнес-логику: шаблон сервис список and получение.
func TestTemplateService_ListAndGet(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	systemID := uuid.New()
	repo.templates[systemID] = &ent.ProjectTemplate{
		ID: systemID, Name: "Empty", TemplateType: projecttemplate.TemplateTypeEmpty,
		Config: map[string]interface{}{"statuses": []string{"Open"}}, IsSystem: true, CreatedAt: time.Now(),
	}

	list, err := svc.ListTemplates(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.True(t, list[0].IsSystem)

	got, err := svc.GetTemplate(ctx, systemID)
	require.NoError(t, err)
	assert.Equal(t, "Empty", got.Name)
}

// TestTemplateService_SaveProjectAsTemplate проверяет бизнес-логику: шаблон сервис save project as шаблон.
func TestTemplateService_SaveProjectAsTemplate(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	projectID := uuid.New()
	userID := uuid.New()
	repo.projects[projectID] = &ent.Project{ID: projectID, Name: "Src", Key: "SRC", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo.members[projectID] = []*ent.ProjectMember{
		{Edges: ent.ProjectMemberEdges{User: &ent.User{ID: userID}}},
	}

	resp, err := svc.SaveProjectAsTemplate(ctx, projectID, userID, project.CreateTemplateRequest{
		Name:        "My Template",
		Description: "from project",
	})
	require.NoError(t, err)
	assert.Equal(t, "My Template", resp.Name)
	assert.Equal(t, string(projecttemplate.TemplateTypeCustom), resp.TemplateType)
}

// TestTemplateService_SaveProjectAsTemplate_Forbidden проверяет отклонение при шаблон сервис save project as шаблон запрещённый.
func TestTemplateService_SaveProjectAsTemplate_Forbidden(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	_, err := svc.SaveProjectAsTemplate(ctx, uuid.New(), uuid.New(), project.CreateTemplateRequest{Name: "X"})
	require.Error(t, err)
}

// TestTemplateService_DeleteTemplate_SystemForbidden проверяет отклонение при шаблон сервис удаление шаблон system запрещённый.
func TestTemplateService_DeleteTemplate_SystemForbidden(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	systemID := uuid.New()
	repo.templates[systemID] = &ent.ProjectTemplate{
		ID: systemID, Name: "System", IsSystem: true, CreatedAt: time.Now(),
	}

	err := svc.DeleteTemplate(ctx, systemID, uuid.New())
	require.Error(t, err)
}

// TestTemplateService_SeedSystemTemplates проверяет бизнес-логику: шаблон сервис seed system templates.
func TestTemplateService_SeedSystemTemplates(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	require.NoError(t, svc.SeedSystemTemplates(ctx))
	assert.Len(t, repo.templates, 4)

	// Повторный вызов идемпотентен.
	require.NoError(t, svc.SeedSystemTemplates(ctx))
	assert.Len(t, repo.templates, 4)
}

// TestToTemplateResponse проверяет сценарий: to шаблон response.
func TestToTemplateResponse(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	resp := project.ExportToTemplateResponse(&ent.ProjectTemplate{
		ID: id, Name: "QA", Description: "qa template",
		TemplateType: projecttemplate.TemplateTypeQa,
		Config:       map[string]interface{}{"statuses": []string{"Open"}},
		IsSystem:     true, CreatedAt: now,
	})
	assert.Equal(t, id.String(), resp.ID)
	assert.Equal(t, "qa", resp.TemplateType)
}
