package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/ai"
	githubmod "github.com/valerii/hub/backend/internal/github"
)

// TestAdminUpdate_UserScopeProvider_404_TASK066 — W5-A1-010: admin PATCH на user BYOK → 404.
func TestAdminUpdate_UserScopeProvider_404_TASK066(t *testing.T) {
	userID := uuid.New()
	owner := uuid.New()
	byok := &ent.AIProvider{
		ID:           uuid.New(),
		Name:         "user-byok",
		ProviderType: aiprovider.ProviderTypeOpenaiCompatible,
		BaseURL:      "https://api.openai.com/v1",
		Models:       []string{"gpt-4o-mini"},
		Enabled:      true,
		Scope:        aiprovider.ScopeUser,
	}
	repo := &mockRepo{
		providers:   []*ent.AIProvider{byok},
		ownerByProv: map[uuid.UUID]uuid.UUID{byok.ID: owner},
	}
	svc := ai.NewService(repo, &mockAdapter{}, nil)
	h := ai.NewAdminHandler(svc, func(_ fiber.Ctx, id uuid.UUID) (bool, error) {
		return id == userID, nil
	})

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	name := "hacked"
	body, _ := json.Marshal(ai.UpdateProviderRequest{Name: &name})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/ai/providers/"+byok.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/admin/ai/providers/"+byok.ID.String()+"/test", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestFetchReview_GitLabURL_NilGitLab_TASK066 — W5-A3-013: GitLab URL + nil gitlab → явная ошибка.
func TestFetchReview_GitLabURL_NilGitLab_TASK066(t *testing.T) {
	p := ai.NewUnifiedCodeReviewProvider(&githubmod.GitHubService{}, nil)
	p.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{}})
	_, err := p.FetchReview(context.Background(), uuid.New(), "https://gitlab.com/g/p/-/merge_requests/5")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GitLab not configured")
}

// TestEnsureBotUser_Parallel_OneUser_TASK066 — W5-A3-015: parallel EnsureBot → 1 user.
func TestEnsureBotUser_Parallel_OneUser_TASK066(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:aibot_race_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	const n = 16
	ids := make([]uuid.UUID, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ids[i], errs[i] = ai.EnsureBotUser(ctx, ai.NewEntBotUserRepository(client))
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		require.NoError(t, errs[i], "goroutine %d", i)
		require.NotEqual(t, uuid.Nil, ids[i])
		assert.Equal(t, ids[0], ids[i])
	}
	count, err := client.User.Query().Where(user.EmailEQ(ai.BotEmail)).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
