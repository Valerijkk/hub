package tms_test

// FA-02-02: CI/CD-вебхуки TMS (/webhook, /webhook/junit) создают прогоны в
// проекте из пути запроса. Раньше они стояли только за authMw (JWT) без
// rbac.RequirePermission, из-за чего любой аутентифицированный пользователь мог
// импортировать результаты в чужой проект, зная его UUID (IDOR / OWASP A01).
// Эти тесты проходят через реальный RegisterRoutes с rbac-сервисом и проверяют,
// что участник проекта проходит, а посторонний получает 403.
import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/rbac"
	"github.com/valerii/hub/backend/internal/tms"
)

// newTMSWebhookRBACEnv поднимает реальные маршруты TMS через RegisterRoutes.
// authMw берёт userID из заголовка X-Test-User, что позволяет в одном приложении
// действовать от лица разных пользователей (участник vs посторонний). RBAC
// проверяется на реальном ent-клиенте; TMS-сервис на in-memory mock-репозитории,
// поэтому импорт результатов не требует MongoDB.
func newTMSWebhookRBACEnv(t *testing.T) (app *fiber.App, projectID, memberID, outsiderID uuid.UUID) {
	t.Helper()
	dsn := "file:tmswebhookrbac_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	member, err := client.User.Create().
		SetEmail("tms-member-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	outsider, err := client.User.Create().
		SetEmail("tms-outsider-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("TMS Webhook RBAC").
		SetKey("TW").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(member.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	svc := tms.NewTMSService(newMockRepo(), nil)
	h := tms.NewHandler(svc)

	app = fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return app, proj.ID, member.ID, outsider.ID
}

// TestWebhookAllure_RBAC_MemberAllowed проверяет, что участник проекта с
// правом PermTestRunCreate может импортировать результаты через вебхук.
func TestWebhookAllure_RBAC_MemberAllowed(t *testing.T) {
	app, projectID, memberID, _ := newTMSWebhookRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook"

	req := httptest.NewRequest(http.MethodPost, base+"?title=CI+Run", bytes.NewReader([]byte(validAllureJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestWebhookAllure_RBAC_OutsiderForbidden проверяет, что посторонний
// пользователь не может импортировать результаты в чужой проект (IDOR закрыт).
func TestWebhookAllure_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newTMSWebhookRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook"

	req := httptest.NewRequest(http.MethodPost, base+"?title=CI+Run", bytes.NewReader([]byte(validAllureJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", outsiderID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestWebhookJUnit_RBAC_MemberAllowed проверяет доступ участника к JUnit-вебхуку.
func TestWebhookJUnit_RBAC_MemberAllowed(t *testing.T) {
	app, projectID, memberID, _ := newTMSWebhookRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook/junit"

	req := httptest.NewRequest(http.MethodPost, base+"?title=JUnit+CI", bytes.NewReader([]byte(junitTestSuitesXML)))
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("X-Test-User", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestWebhookJUnit_RBAC_OutsiderForbidden проверяет, что посторонний не может
// загрузить JUnit-результаты в чужой проект.
func TestWebhookJUnit_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newTMSWebhookRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms/webhook/junit"

	req := httptest.NewRequest(http.MethodPost, base+"?title=JUnit+CI", bytes.NewReader([]byte(junitTestSuitesXML)))
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("X-Test-User", outsiderID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestTMSRoutes_RBAC_OutsiderForbidden фиксирует, что и обычные TMS-маршруты
// (не только вебхуки) теперь действительно применяют RBAC: посторонний не
// может создать тест-кейс или прогон в чужом проекте. Регрессионный тест на
// порядок аргументов Fiber v3 в RegisterRoutes.
func TestTMSRoutes_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newTMSWebhookRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/tms"

	cases := []struct {
		name string
		path string
		body string
	}{
		{"create case", base + "/cases", `{"title":"x","case_type":"classic","priority":"high"}`},
		{"create run", base + "/runs", `{"title":"x"}`},
		{"allure import", base + "/allure/import", `{"results":[],"title":"x"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User", outsiderID.String())
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		})
	}
}
