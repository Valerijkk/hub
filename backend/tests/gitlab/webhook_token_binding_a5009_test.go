//go:build hubtest

package gitlab_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/gitlab"
)

// TestHandler_HandleWebhook_RejectsForeignToken_A5_009 proves X-Gitlab-Token is bound
// to the integration matched by project_path — not any instance secret (A5-009).
// Mutation: restore ListAllWebhookSecrets any-of loop in VerifyWebhookToken → this FAIL.
func TestHandler_HandleWebhook_RejectsForeignToken_A5_009(t *testing.T) {
	secretA := "token-integration-a"
	secretB := "token-integration-b"
	repo := &gitlabMockRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"group/app-a": {
				ID:            uuid.New(),
				ProjectPath:   "group/app-a",
				WebhookSecret: secretA,
			},
			"group/app-b": {
				ID:            uuid.New(),
				ProjectPath:   "group/app-b",
				WebhookSecret: secretB,
			},
		},
		webhookSecrets: []string{secretA, secretB},
	}
	app := testGitLabHandlerApp(t, repo)

	body, err := json.Marshal(gitlab.WebhookPushEvent{
		ObjectKind: "push",
		Project:    gitlab.WebhookProject{PathWithNamespace: "group/app-b"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/gitlab/webhook", bytes.NewReader(body))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", secretA) // foreign token for app-b
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"X-Gitlab-Token from foreign integration must be rejected (A5-009)")
}

func TestHandler_HandleWebhook_MatchedTokenOK_A5_009(t *testing.T) {
	secret := "secret-token-matched"
	repo := &gitlabMockRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"hub/backend": {
				ID:            uuid.New(),
				ProjectPath:   "hub/backend",
				WebhookSecret: secret,
			},
		},
	}
	app := testGitLabHandlerApp(t, repo)

	body, err := json.Marshal(gitlab.WebhookPushEvent{
		ObjectKind: "push",
		Project:    gitlab.WebhookProject{PathWithNamespace: "hub/backend"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/gitlab/webhook", bytes.NewReader(body))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", secret)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
