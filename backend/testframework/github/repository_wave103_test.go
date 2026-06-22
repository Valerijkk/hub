package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

func TestEntRepository_EncryptDecryptIntegrationFields(t *testing.T) {
	env := newGitHubRepoEnv(t)
	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	env.repo = github.NewEntRepository(env.client, cipher)

	plainToken := "ghp_wave103_secret_token"
	plainSecret := "whsec_wave103"

	created, err := env.repo.CreateIntegration(env.ctx, env.projectID, github.CreateIntegrationInput{
		RepoOwner:     "acme",
		RepoName:      "hub",
		AccessToken:   plainToken,
		WebhookSecret: plainSecret,
	})
	require.NoError(t, err)
	assert.Equal(t, plainToken, created.AccessToken)

	raw, err := env.client.GitHubIntegration.Get(env.ctx, created.ID)
	require.NoError(t, err)
	assert.NotEqual(t, plainToken, raw.AccessToken)
	assert.NotEqual(t, plainSecret, raw.WebhookSecret)

	loaded, err := env.repo.GetIntegration(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, plainToken, loaded.AccessToken)
	assert.Equal(t, plainSecret, loaded.WebhookSecret)
}

func TestEntRepository_encryptFieldRoundTrip(t *testing.T) {
	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	repo := github.ExportNewEntRepositoryForEncryptTest(cipher)

	enc, err := github.ExportEntRepositoryEncryptField(repo, "ghp_secret")
	require.NoError(t, err)
	assert.NotEqual(t, "ghp_secret", enc)
	assert.Equal(t, "ghp_secret", github.ExportEntRepositoryDecryptField(repo, enc))
}

func TestEntRepository_decryptField_PlaintextLegacy(t *testing.T) {
	env := newGitHubRepoEnv(t)
	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	repo := github.NewEntRepository(env.client, cipher)

	assert.Equal(t, "legacy-plain", github.ExportEntRepositoryDecryptField(repo, "legacy-plain"))
	assert.Equal(t, "", github.ExportEntRepositoryDecryptField(repo, ""))
}

func TestEntRepository_ListAllWebhookSecrets(t *testing.T) {
	env := newGitHubRepoEnv(t)
	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	env.repo = github.NewEntRepository(env.client, cipher)

	_, err = env.repo.CreateIntegration(env.ctx, env.projectID, github.CreateIntegrationInput{
		RepoOwner: "o1", RepoName: "r1", AccessToken: "tok1", WebhookSecret: "sec-a",
	})
	require.NoError(t, err)
	_, err = env.repo.CreateIntegration(env.ctx, env.projectID, github.CreateIntegrationInput{
		RepoOwner: "o2", RepoName: "r2", AccessToken: "tok2", WebhookSecret: "sec-b",
	})
	require.NoError(t, err)

	secrets, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Contains(t, secrets, "sec-a")
	assert.Contains(t, secrets, "sec-b")

	secrets2, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, secrets, secrets2)
}

func TestGetWorkflowRunDetail_SuccessWithJobs(t *testing.T) {
	now := time.Now().UTC()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/actions/runs/42") && !strings.Contains(r.URL.Path, "/jobs"):
			_, _ = w.Write([]byte(`{
				"id":42,"name":"CI","status":"completed","conclusion":"success",
				"head_branch":"main","event":"push",
				"html_url":"https://github.com/acme/hub/actions/runs/42",
				"created_at":"` + now.Format(time.RFC3339) + `",
				"updated_at":"` + now.Format(time.RFC3339) + `",
				"head_sha":"abc","run_number":7
			}`))
		case strings.Contains(r.URL.Path, "/actions/runs/42/jobs"):
			_, _ = w.Write([]byte(`{"jobs":[
				{"id":1,"name":"build","status":"completed","conclusion":"success",
				 "started_at":"` + now.Format(time.RFC3339) + `",
				 "completed_at":"` + now.Format(time.RFC3339) + `"}
			]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	repo := &githubMockRepo{integrations: map[string]*ent.GitHubIntegration{}}
	_, err := repo.CreateIntegration(context.Background(), projectID, github.CreateIntegrationInput{
		RepoOwner: "acme", RepoName: "hub", AccessToken: "ghp_test",
	})
	require.NoError(t, err)

	svc := github.NewGitHubServiceWithHTTPClient(repo, nil, srv.Client(), srv.URL)
	detail, err := svc.GetWorkflowRunDetail(context.Background(), projectID, 42)
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, int64(42), detail.ID)
	require.Len(t, detail.Jobs, 1)
	assert.Equal(t, "build", detail.Jobs[0].Name)
}
