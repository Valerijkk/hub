package project_test

// Тесты модуля проектов: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/projecttemplate"
)

// TestTemplateService_ExportImport_EntRepo проверяет слой репозитория: шаблон сервис экспорт импорт ent repo.
func TestTemplateService_ExportImport_EntRepo(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewTemplateService(env.repo)

	tpl, err := env.repo.CreateTemplate(env.ctx, "EntExport", "roundtrip", projecttemplate.TemplateTypeQa,
		map[string]interface{}{"statuses": []string{"Todo", "Done"}}, &env.owner.ID, false)
	require.NoError(t, err)

	data, err := svc.ExportTemplate(env.ctx, tpl.ID)
	require.NoError(t, err)

	var export project.TemplateExport
	require.NoError(t, json.Unmarshal(data, &export))
	assert.Equal(t, "EntExport", export.Name)
	assert.Equal(t, project.ExportTemplateExportVersion(), export.Version)

	imported, err := svc.ImportTemplate(env.ctx, env.owner.ID, data)
	require.NoError(t, err)
	assert.Equal(t, "EntExport", imported.Name)
	assert.Equal(t, string(projecttemplate.TemplateTypeCustom), imported.TemplateType)
}

// TestTemplateService_ImportTemplate_InvalidJSON проверяет отклонение при шаблон сервис импорт шаблон некорректный json.
func TestTemplateService_ImportTemplate_InvalidJSON(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewTemplateService(env.repo)

	_, err := svc.ImportTemplate(env.ctx, env.owner.ID, []byte("{bad json"))
	require.Error(t, err)

	_, err = svc.ImportTemplate(env.ctx, env.owner.ID, []byte(`{"config":{}}`))
	require.Error(t, err)
}

// TestTemplateService_ExportTemplate_NotFound проверяет бизнес-логику: шаблон сервис экспорт шаблон not found.
func TestTemplateService_ExportTemplate_NotFound(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewTemplateService(env.repo)

	_, err := svc.ExportTemplate(env.ctx, uuid.New())
	require.Error(t, err)
}
