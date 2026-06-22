package project_test

// Тесты HTTP-обработчика проектов (волна 101): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"bytes"
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
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestTemplateHandler_SaveAsTemplate проверяет HTTP-слой: шаблон HTTP-обработчик save as шаблон.
func TestTemplateHandler_SaveAsTemplate(t *testing.T) {
	dsn := "file:save_tpl_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()

	user, err := client.User.Create().
		SetEmail("save-tpl-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Save Template Source").
		SetKey("STS").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(user.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.WorkflowStatus.Create().
		SetName("Open").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(proj.ID).
		SetIsDefault(true).
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

	body := fmt.Sprintf(`{"project_id":"%s","name":"From Project","description":"snapshot"}`, proj.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/project-templates/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestTemplateHandler_SaveAsTemplate_BadProjectID проверяет HTTP-слой: шаблон HTTP-обработчик save as шаблон bad project id.
func TestTemplateHandler_SaveAsTemplate_BadProjectID(t *testing.T) {
	env := newTemplateRoutesEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/project-templates/", bytes.NewReader([]byte(`{"project_id":"bad","name":"X"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
