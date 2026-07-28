package tracker_test

// P2-BE-01: StructValidator на реальном CreateIssue / CreateComment (мутация).
import (
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
	"github.com/valerii/hub/backend/internal/platform/middleware"
	"github.com/valerii/hub/backend/internal/rbac"
	"github.com/valerii/hub/backend/internal/tracker"
)

func newTrackerEnv(t *testing.T, withValidator bool) (*fiber.App, uuid.UUID) {
	t.Helper()
	dsn := "file:trval_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
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
		SetEmail("trv-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Tracker Validator").
		SetKey("TV").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, tracker.NewEntRepository(client, nil).SeedDefaultWorkflow(ctx, proj.ID))

	repo := tracker.NewEntRepository(client, nil)
	svc := tracker.NewTrackerService(repo, nil)
	h := tracker.NewHandler(svc, tracker.NewAutomationService(repo))

	cfg := fiber.Config{ErrorHandler: middleware.ErrorHandler}
	if withValidator {
		cfg.StructValidator = middleware.NewStructValidator()
	}
	app := fiber.New(cfg)
	app.Use(middleware.RequestID())
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return app, proj.ID
}

// TestCreateIssue_EmptyTitle_StructValidatorMutation:
// с validator → 400; без → ошибка/не 201 от Ent NotEmpty, либо 201 если пустой title
// проскочил — канон мутации для title: Bind-reject только с validator.
func TestCreateIssue_EmptyTitle_StructValidatorMutation(t *testing.T) {
	payload := `{"title":"","issue_type":"task","priority":"medium"}`

	t.Run("with_validator_400", func(t *testing.T) {
		app, projectID := newTrackerEnv(t, true)
		base := "/api/projects/" + projectID.String()
		req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		assert.Contains(t, body, "error")
		assert.Contains(t, body, "request_id")
		_, hasCode := body["code"]
		_, hasMessage := body["message"]
		assert.False(t, hasCode)
		assert.False(t, hasMessage)
	})

	t.Run("without_validator_not_400_from_bind", func(t *testing.T) {
		app, projectID := newTrackerEnv(t, false)
		base := "/api/projects/" + projectID.String()
		req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		// Без StructValidator Bind пропускает; Ent NotEmpty обычно даёт 500 через ErrorHandler,
		// но не 400 от validate-тегов.
		assert.NotEqual(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestCreateComment_MissingRequired_StructValidatorMutation:
// без body_text/body_json: с validator → 400; без → 201 (как wave119).
func TestCreateComment_MissingRequired_StructValidatorMutation(t *testing.T) {
	createIssue := func(t *testing.T, app *fiber.App, projectID uuid.UUID) tracker.IssueResponse {
		t.Helper()
		base := "/api/projects/" + projectID.String()
		body := `{"title":"Validator issue","issue_type":"task","priority":"medium"}`
		req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var created tracker.IssueResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
		return created
	}

	commentBody := `{"body":"legacy field without body_text"}`

	t.Run("with_validator_400", func(t *testing.T) {
		app, projectID := newTrackerEnv(t, true)
		issue := createIssue(t, app, projectID)
		base := "/api/projects/" + projectID.String() + "/issues/" + issue.IssueKey
		req := httptest.NewRequest(http.MethodPost, base+"/comments", bytes.NewReader([]byte(commentBody)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("without_validator_201", func(t *testing.T) {
		app, projectID := newTrackerEnv(t, false)
		issue := createIssue(t, app, projectID)
		base := "/api/projects/" + projectID.String() + "/issues/" + issue.IssueKey
		req := httptest.NewRequest(http.MethodPost, base+"/comments", bytes.NewReader([]byte(commentBody)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})
}
