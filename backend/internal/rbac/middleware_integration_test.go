package rbac

import (
	"context"
	"crypto/rand"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

// randomProjectKey генерирует валидный по schema-валидатору ключ проекта:
// [A-Z][A-Z0-9]{1..9}. Использует криптогенератор для уникальности — иначе
// несколько тестов в одной shared-cache sqlite-сессии получили бы дубликат
// уникального ключа.
func randomProjectKey() string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const alnum = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	key := []byte{alphabet[b[0]%26]}
	for i := 1; i < 8; i++ {
		key = append(key, alnum[b[i]%36])
	}
	return string(key)
}

// rbacFixture хранит общую тестовую обвязку: ent-клиент (sqlite in-memory),
// miniredis-кэш разрешений и сервис. Заполняется одним owner, одним viewer и
// одним admin-пользователем в одном проекте.
type rbacFixture struct {
	client    *ent.Client
	rdb       *redis.Client
	svc       *RBACService
	ownerID   uuid.UUID
	viewerID  uuid.UUID
	adminID   uuid.UUID
	strangerID uuid.UUID // существует в users, но не в project_members
	projectID uuid.UUID
}

// newRBACFixture поднимает sqlite-ent, miniredis-кэш и наполняет минимальный
// набор сущностей. Все cleanup'ы регистрируются через t.Cleanup.
func newRBACFixture(t *testing.T) *rbacFixture {
	t.Helper()

	// sqlite in-memory с FK-проверками — близко к поведению postgres.
	// Каждый тест получает свою отдельную in-memory БД, чтобы исключить
	// взаимовлияние при параллельном/последовательном прогоне.
	dsn := "file:rbactest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	cache := NewPermissionCache(rdb)
	svc := NewRBACService(client, cache)

	ctx := context.Background()

	// Project. Ключ должен матчить ^[A-Z][A-Z0-9]{1,9}$ (см. ent/schema/helpers.go),
	// поэтому используем псевдослучайную uppercase-строку.
	proj, err := client.Project.Create().
		SetName("Test Project").
		SetKey(randomProjectKey()).
		Save(ctx)
	require.NoError(t, err)

	// owner (member), viewer (member), admin (system role), stranger (not a member).
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	viewer, err := client.User.Create().
		SetEmail("viewer-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	admin, err := client.User.Create().
		SetEmail("admin-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	stranger, err := client.User.Create().
		SetEmail("stranger-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	// ProjectMember: owner has role "owner" (all perms), viewer has role "viewer" (read-only).
	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(viewer.ID).
		SetProjectID(proj.ID).
		SetProjectRole("viewer").
		Save(ctx)
	require.NoError(t, err)

	return &rbacFixture{
		client:     client,
		rdb:        rdb,
		svc:        svc,
		ownerID:    owner.ID,
		viewerID:   viewer.ID,
		adminID:    admin.ID,
		strangerID: stranger.ID,
		projectID:  proj.ID,
	}
}

// newAppWithUserID собирает мини-fiber-приложение, идентичное production-
// паттерну (см. backend/internal/apigateway/handler.go):
//   api.Group("/projects/:projectId/...", authMw)
//   group.Get(path, mainHandler, middlewares...)
//
// ВАЖНО: в Fiber v3 сигнатура Get(path, handler, middleware...) — это
// **handler идёт последним** в фактической цепочке исполнения. То есть
// `Get(path, h, mw)` собирает [mw, h] и middleware mw запускается ПЕРВЫМ.
// См. router.go:291 в gofiber/fiber/v3@v3.0.0-beta.3: `handlers = middleware
// + [handler]`. Если хочется поставить guard ПЕРЕД terminal, нужно передать
// guard как middleware (последний variadic-аргумент), а terminal как handler.
func newAppWithUserID(userIDForLocals string, guard fiber.Handler, route string) *fiber.App {
	app := fiber.New()
	localsMw := func(c fiber.Ctx) error {
		if userIDForLocals != "" {
			c.Locals("userID", userIDForLocals)
		}
		return c.Next()
	}
	// Группа: localsMw как mw (variadic), terminal-handler в самой Get-цепочке.
	// guard также передаётся как middleware — он запустится ДО terminal'а.
	grp := app.Group("", localsMw)
	terminal := func(c fiber.Ctx) error {
		return c.SendString("ok")
	}
	grp.Get(route, terminal, guard)
	return app
}

// doGet выполняет httptest-style вызов против Fiber-приложения и возвращает
// статус и тело. Использует app.Test (стандартный для Fiber путь).
func doGet(t *testing.T, app *fiber.App, target string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(fiber.MethodGet, target, nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

// TestRequirePermission_Anonymous_401 — если в Locals нет userID, гард должен
// вернуть 401 "user not authenticated".
func TestRequirePermission_Anonymous_401(t *testing.T) {
	f := newRBACFixture(t)

	guard := RequirePermission(PermProjectView, f.svc)
	app := newAppWithUserID("", guard, "/api/projects/:projectId/test")

	status, body := doGet(t, app, "/api/projects/"+f.projectID.String()+"/test")
	assert.Equal(t, fiber.StatusUnauthorized, status)
	assert.Contains(t, body, "user not authenticated")
}

// TestRequirePermission_NonMember_403 — пользователь есть в users, но не
// зачислен в project_members. Должен вернуть 403 "not a project member".
func TestRequirePermission_NonMember_403(t *testing.T) {
	f := newRBACFixture(t)

	guard := RequirePermission(PermProjectView, f.svc)
	app := newAppWithUserID(f.strangerID.String(), guard, "/api/projects/:projectId/test")

	status, body := doGet(t, app, "/api/projects/"+f.projectID.String()+"/test")
	assert.Equal(t, fiber.StatusForbidden, status)
	assert.Contains(t, body, "not a project member")
}

// TestRequirePermission_MemberWithPermission_Next — viewer имеет
// PermProjectView, гард пропускает запрос дальше и terminal handler пишет
// "ok" с 200.
func TestRequirePermission_MemberWithPermission_Next(t *testing.T) {
	f := newRBACFixture(t)

	guard := RequirePermission(PermProjectView, f.svc)
	app := newAppWithUserID(f.viewerID.String(), guard, "/api/projects/:projectId/test")

	status, body := doGet(t, app, "/api/projects/"+f.projectID.String()+"/test")
	assert.Equal(t, fiber.StatusOK, status)
	assert.Equal(t, "ok", body)
}

// TestRequirePermission_MemberWithoutPermission_403 — viewer не имеет
// PermProjectEdit; гард возвращает 403 "insufficient permissions".
func TestRequirePermission_MemberWithoutPermission_403(t *testing.T) {
	f := newRBACFixture(t)

	guard := RequirePermission(PermProjectEdit, f.svc)
	app := newAppWithUserID(f.viewerID.String(), guard, "/api/projects/:projectId/test")

	status, body := doGet(t, app, "/api/projects/"+f.projectID.String()+"/test")
	assert.Equal(t, fiber.StatusForbidden, status)
	assert.Contains(t, body, "insufficient permissions")
}

// TestRequireSystemRole_AdminOnly — endpoint защищён RequireAdmin: admin
// (system_role=admin) проходит, viewer (system_role=member) — 403.
func TestRequireSystemRole_AdminOnly(t *testing.T) {
	f := newRBACFixture(t)

	guard := RequireAdmin(f.client)

	t.Run("admin allowed", func(t *testing.T) {
		app := newAppWithUserID(f.adminID.String(), guard, "/admin/test")
		status, body := doGet(t, app, "/admin/test")
		assert.Equal(t, fiber.StatusOK, status)
		assert.Equal(t, "ok", body)
	})

	t.Run("member forbidden", func(t *testing.T) {
		app := newAppWithUserID(f.viewerID.String(), guard, "/admin/test")
		status, body := doGet(t, app, "/admin/test")
		assert.Equal(t, fiber.StatusForbidden, status)
		assert.Contains(t, body, "insufficient system role")
	})

	t.Run("anonymous unauthorized", func(t *testing.T) {
		app := newAppWithUserID("", guard, "/admin/test")
		status, body := doGet(t, app, "/admin/test")
		assert.Equal(t, fiber.StatusUnauthorized, status)
		assert.Contains(t, body, "user not authenticated")
	})
}

// TestRequirePermission_InvalidUserID_401 — если в Locals значение не-UUID,
// гард падает с 401 "invalid user ID".
func TestRequirePermission_InvalidUserID_401(t *testing.T) {
	f := newRBACFixture(t)

	guard := RequirePermission(PermProjectView, f.svc)
	app := newAppWithUserID("not-a-uuid", guard, "/api/projects/:projectId/test")

	status, body := doGet(t, app, "/api/projects/"+f.projectID.String()+"/test")
	assert.Equal(t, fiber.StatusUnauthorized, status)
	assert.Contains(t, body, "invalid user ID")
}

// TestRequirePermission_MissingProjectID_400 — если в пути нет projectId,
// гард падает с 400 "project ID required". Защищает контракт middleware:
// он всегда должен видеть параметр пути.
func TestRequirePermission_MissingProjectID_400(t *testing.T) {
	f := newRBACFixture(t)

	guard := RequirePermission(PermProjectView, f.svc)
	// Маршрут БЕЗ :projectId — Fiber передаст пустую строку.
	app := newAppWithUserID(f.viewerID.String(), guard, "/api/no-project/test")

	status, body := doGet(t, app, "/api/no-project/test")
	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Contains(t, body, "project ID required")
}
