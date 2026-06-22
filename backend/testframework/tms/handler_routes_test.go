package tms_test

// Тесты HTTP-обработчика TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/rbac"
)

func newTMSRoutesEnv(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	dsn := "file:tmsroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("tms-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("TMS Routes").
		SetKey("TM").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	repo := tms.NewEntRepository(client)
	svc := tms.NewTMSService(tms.NewCombinedRepository(repo, &tms.MongoRepository{}), nil)
	h := tms.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return app, proj.ID
}

// TestHandler_RegisterRoutes_CreateCaseAndSuite проверяет HTTP-слой: HTTP-обработчик register маршруты создание case and suite.
func TestHandler_RegisterRoutes_CreateCaseAndSuite(t *testing.T) {
	app, projectID := newTMSRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	suiteBody := `{"name":"Regression"}`
	req := httptest.NewRequest(http.MethodPost, base+"/suites", bytes.NewReader([]byte(suiteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	caseBody := `{"title":"Login test","case_type":"classic","priority":"high"}`
	req = httptest.NewRequest(http.MethodPost, base+"/cases", bytes.NewReader([]byte(caseBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/cases", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RegisterRoutes_Environments проверяет HTTP-слой: HTTP-обработчик register маршруты environments.
func TestHandler_RegisterRoutes_Environments(t *testing.T) {
	app, projectID := newTMSRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/environments"

	body := `{"name":"Staging","label":"staging"}`
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tms.EnvironmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	patch := `{"name":"Staging v2"}`
	req = httptest.NewRequest(http.MethodPut, base+"/"+created.ID, bytes.NewReader([]byte(patch)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/"+created.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_RegisterRoutes_DefectCategories проверяет HTTP-слой: HTTP-обработчик register маршруты defect categories.
func TestHandler_RegisterRoutes_DefectCategories(t *testing.T) {
	app, projectID := newTMSRoutesEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/tms/reports/defect-categories", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestNewHandler проверяет HTTP-слой: new HTTP-обработчик.
func TestNewHandler(t *testing.T) {
	h := tms.NewHandler(tms.NewTMSService(newMockRepo(), nil))
	require.NotNil(t, h)
}

// TestHandler_RegisterRoutes_Wave122_RunLifecycle проверяет HTTP-слой: HTTP-обработчик register маршруты wave 122 run жизненный цикл.
func TestHandler_RegisterRoutes_Wave122_RunLifecycle(t *testing.T) {
	app, projectID := newTMSRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	runBody := `{"title":"Routes lifecycle"}`
	req := httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var run tms.TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/start", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/complete", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, base+"/runs/"+run.ID+"/reopen", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RegisterRoutes_Wave122_ExportAndImport проверяет HTTP-слой: HTTP-обработчик register маршруты wave 122 экспорт and импорт.
func TestHandler_RegisterRoutes_Wave122_ExportAndImport(t *testing.T) {
	app, projectID := newTMSRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	req := httptest.NewRequest(http.MethodGet, base+"/export?format=csv", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	runBody := `{"title":"Export run"}`
	req = httptest.NewRequest(http.MethodPost, base+"/runs", bytes.NewReader([]byte(runBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var run tms.TestRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&run))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodGet, base+"/runs/"+run.ID+"/export?format=csv", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RegisterRoutes_Wave122_ListEnvironmentsEmpty проверяет HTTP-слой: HTTP-обработчик register маршруты wave 122 список environments пустой.
func TestHandler_RegisterRoutes_Wave122_ListEnvironmentsEmpty(t *testing.T) {
	app, projectID := newTMSRoutesEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/environments/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
