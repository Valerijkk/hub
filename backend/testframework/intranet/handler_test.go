package intranet_test
import (
	"github.com/valerii/hub/backend/internal/intranet"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func setupIntranetTest(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:intranet_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	admin, err := client.User.Create().
		SetEmail("admin@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(context.Background())
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("member@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(context.Background())
	require.NoError(t, err)

	repo := intranet.NewEntRepository(client)
	svc := intranet.NewService(repo, nil, nil)
	h := intranet.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		uid := c.Get("X-Test-User-ID")
		if uid != "" {
			c.Locals("userID", uid)
		}
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, admin.ID, member.ID
}

func TestHandler_CreateListNews_AdminOnly(t *testing.T) {
	app, adminID, memberID := setupIntranetTest(t)

	body := `{"title":"Welcome","body":"Hello team","pinned":true,"published":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/intranet/news", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/intranet/news", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created intranet.NewsPostResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "Welcome", created.Title)
	assert.True(t, created.Pinned)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/news", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var list intranet.NewsListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	require.Len(t, list.Items, 1)
	assert.Equal(t, created.ID, list.Items[0].ID)
}

func TestHandler_DeleteNews(t *testing.T) {
	app, adminID, _ := setupIntranetTest(t)

	body := `{"title":"Temp","body":"Remove me","published":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/intranet/news", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created intranet.NewsPostResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	req = httptest.NewRequest(http.MethodDelete, "/api/intranet/news/"+created.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/news/"+created.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_ListDirectory_GroupsByDepartment(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:dir_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	member, err := client.User.Create().
		SetEmail("member@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.User.Create().
		SetEmail("eng@hub.local").
		SetDisplayName("Alice Eng").
		SetDepartment("Engineering").
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.User.Create().
		SetEmail("hr@hub.local").
		SetDisplayName("Bob HR").
		SetDepartment("HR").
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.User.Create().
		SetEmail("pending@hub.local").
		SetStatus(user.StatusPending).
		Save(ctx)
	require.NoError(t, err)

	repo := intranet.NewEntRepository(client)
	svc := intranet.NewService(repo, nil, nil)
	h := intranet.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", member.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodGet, "/api/intranet/directory", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var dir intranet.DirectoryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&dir))
	assert.Equal(t, 3, dir.Total)
	require.Len(t, dir.Departments, 3)

	deptMap := make(map[string]int)
	for _, g := range dir.Departments {
		deptMap[g.Department] = len(g.Users)
	}
	assert.Equal(t, 1, deptMap["Engineering"])
	assert.Equal(t, 1, deptMap["HR"])
	assert.Equal(t, 1, deptMap["Без отдела"])
}

func TestHandler_GetDirectoryUser_NotFoundForPending(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:pend_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	member, err := client.User.Create().
		SetEmail("viewer@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	pending, err := client.User.Create().
		SetEmail("pending2@hub.local").
		SetStatus(user.StatusPending).
		Save(ctx)
	require.NoError(t, err)

	repo := intranet.NewEntRepository(client)
	h := intranet.NewHandler(intranet.NewService(repo, nil, nil))
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", member.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodGet, "/api/intranet/directory/users/"+pending.ID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_CreateKudos_FeedAndReceived(t *testing.T) {
	app, adminID, memberID := setupIntranetTest(t)

	body, _ := json.Marshal(map[string]any{
		"recipient_id": adminID.String(),
		"message":      "Спасибо за помощь с релизом!",
		"is_public":    true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/intranet/kudos/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created intranet.KudoResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "Спасибо за помощь с релизом!", created.Message)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/kudos/feed", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	var feed intranet.KudoListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&feed))
	require.Len(t, feed.Items, 1)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/kudos/received", nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	var received intranet.KudoListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&received))
	require.Len(t, received.Items, 1)
}

func TestHandler_RandomCoffee_EnrollAndRun(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().SetEmail("admin@hub.local").SetSystemRole(user.SystemRoleAdmin).Save(ctx)
	require.NoError(t, err)
	for i, email := range []string{"u1@hub.local", "u2@hub.local", "u3@hub.local"} {
		_, err := client.User.Create().SetEmail(email).SetSystemRole(user.SystemRoleMember).Save(ctx)
		require.NoError(t, err)
		_ = i
	}
	users, err := client.User.Query().Where(user.SystemRoleEQ(user.SystemRoleMember)).All(ctx)
	require.NoError(t, err)
	require.Len(t, users, 3)

	repo := intranet.NewEntRepository(client)
	h := intranet.NewHandler(intranet.NewService(repo, nil, nil))
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User-ID"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body, _ := json.Marshal(map[string]any{"active": true})
	for _, u := range users {
		req := httptest.NewRequest(http.MethodPut, "/api/intranet/coffee/enrollment", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", u.ID.String())
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/intranet/coffee/run", nil)
	req.Header.Set("X-Test-User-ID", admin.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var round intranet.CoffeeRoundResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&round))
	// FR-INTR-10 / P1-BE-43: нечётное N → одна тройка, не unpaired.
	assert.Equal(t, 0, round.PairCount)
	assert.Equal(t, 1, round.TripleCount)
	assert.Equal(t, 0, round.UnpairedCount)
}
