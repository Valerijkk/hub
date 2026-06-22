package project_test

// Тесты HTTP-обработчика проектов: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/projecttemplate"
)

type templateRoutesEnv struct {
	app    *fiber.App
	userID uuid.UUID
	repo   *project.EntRepository
	ctx    context.Context
}

func newTemplateRoutesEnv(t *testing.T) templateRoutesEnv {
	t.Helper()
	dsn := "file:tplroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tpl-" + uuid.New().String()[:8] + "@test.com").
		SetName("Template User").
		Save(ctx)
	require.NoError(t, err)

	repo := project.NewEntRepository(client, nil)
	svc := project.NewTemplateService(repo)
	h := project.NewTemplateHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", user.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	return templateRoutesEnv{app: app, userID: user.ID, repo: repo, ctx: ctx}
}

// TestTemplateHandler_RegisterRoutes_ListAndDelete проверяет HTTP-слой: шаблон HTTP-обработчик register маршруты список and удаление.
func TestTemplateHandler_RegisterRoutes_ListAndDelete(t *testing.T) {
	env := newTemplateRoutesEnv(t)

	tpl, err := env.repo.CreateTemplate(env.ctx, "RouteTpl", "via routes", projecttemplate.TemplateTypeCustom,
		map[string]interface{}{"statuses": []string{"Open"}}, &env.userID, false)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/project-templates/", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var list []project.TemplateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	require.NotEmpty(t, list)

	req = httptest.NewRequest(http.MethodDelete, "/api/project-templates/"+tpl.ID.String(), nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestTemplateHandler_RegisterRoutes_ExportImport проверяет HTTP-слой: шаблон HTTP-обработчик register маршруты экспорт импорт.
func TestTemplateHandler_RegisterRoutes_ExportImport(t *testing.T) {
	env := newTemplateRoutesEnv(t)

	tpl, err := env.repo.CreateTemplate(env.ctx, "ExportMe", "export test", projecttemplate.TemplateTypeQa,
		map[string]interface{}{"labels": []string{"qa"}}, &env.userID, false)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/project-templates/"+tpl.ID.String()+"/export", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var export project.TemplateExport
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&export))
	assert.Equal(t, "ExportMe", export.Name)

	importJSON := fmt.Sprintf(`{"name":"Imported","template_type":"custom","config":{"statuses":["A"]},"version":1}`)
	body := &bytes.Buffer{}
	writer := multipartWriter(t, body, "file", "template.json", []byte(importJSON))
	require.NoError(t, writer.Close())

	req = httptest.NewRequest(http.MethodPost, "/api/project-templates/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var imported project.TemplateResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&imported))
	assert.Equal(t, "Imported", imported.Name)
}
