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

func TestHandler_Wave91_NewsAdminUpdateGet(t *testing.T) {
	app, adminID, memberID := setupIntranetTest(t)

	body := `{"title":"Draft","body":"Hidden","published":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/intranet/news", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created intranet.NewsPostResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/news/admin?page=1&per_page=10", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/news/admin?page=1&per_page=10", nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var adminList intranet.NewsListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&adminList))
	require.Len(t, adminList.Items, 1)
	assert.Equal(t, 1, adminList.Page)
	assert.Equal(t, 10, adminList.PerPage)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/news/"+created.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	updateBody, _ := json.Marshal(map[string]any{
		"title":     "Published draft",
		"published": true,
		"version":   created.Version,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/intranet/news/"+created.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated intranet.NewsPostResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "Published draft", updated.Title)
	assert.True(t, updated.Published)

	req = httptest.NewRequest(http.MethodPut, "/api/intranet/news/"+created.ID.String(), bytes.NewReader([]byte(`{"title":"nope"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestHandler_Wave91_CoffeeReadEndpoints(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee91?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().SetEmail("admin91@hub.local").SetSystemRole(user.SystemRoleAdmin).Save(ctx)
	require.NoError(t, err)
	u1, err := client.User.Create().SetEmail("u1@hub.local").SetSystemRole(user.SystemRoleMember).Save(ctx)
	require.NoError(t, err)
	u2, err := client.User.Create().SetEmail("u2@hub.local").SetSystemRole(user.SystemRoleMember).Save(ctx)
	require.NoError(t, err)

	repo := intranet.NewEntRepository(client)
	h := intranet.NewHandler(intranet.NewService(repo, nil, nil))
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User-ID"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	req := httptest.NewRequest(http.MethodGet, "/api/intranet/coffee/enrollment", nil)
	req.Header.Set("X-Test-User-ID", u1.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var en intranet.CoffeeEnrollmentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&en))
	assert.False(t, en.Enrolled)

	body, _ := json.Marshal(map[string]any{"active": true})
	for _, uid := range []uuid.UUID{u1.ID, u2.ID} {
		req = httptest.NewRequest(http.MethodPut, "/api/intranet/coffee/enrollment", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", uid.String())
		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/intranet/coffee/run", nil)
	req.Header.Set("X-Test-User-ID", admin.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/coffee/match", nil)
	req.Header.Set("X-Test-User-ID", u1.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/coffee/stats", nil)
	req.Header.Set("X-Test-User-ID", admin.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var stats intranet.CoffeeStatsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stats))
	assert.Equal(t, 2, stats.ActiveEnrollments)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/coffee/rounds", nil)
	req.Header.Set("X-Test-User-ID", admin.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var rounds struct {
		Items []intranet.CoffeeRoundResponse `json:"items"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rounds))
	assert.Len(t, rounds.Items, 1)
}

func TestHandler_Wave91_KudoAndCoffeeErrors(t *testing.T) {
	app, adminID, memberID := setupIntranetTest(t)

	body, _ := json.Marshal(map[string]any{
		"recipient_id": memberID.String(),
		"message":      "self",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/intranet/kudos/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, _ = json.Marshal(map[string]any{
		"recipient_id": uuid.New().String(),
		"message":      "thanks",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/intranet/kudos/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/intranet/coffee/run", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/coffee/stats", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/news/not-a-uuid", nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/intranet/directory/users/bad-id", nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_Wave91_GetDirectoryUserSuccess(t *testing.T) {
	app, _, memberID := setupIntranetTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/intranet/directory/users/"+memberID.String(), nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var u intranet.DirectoryUserResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&u))
	assert.Equal(t, memberID, u.ID)
}
