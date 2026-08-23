//go:build hubtest

package github_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

// TestHandler_HandleWebhook_RejectsForeignSecret_A5_008 proves HMAC is bound to the
// integration matched by payload repository — not any instance secret (A5-008).
// Mutation: restore ListAllWebhookSecrets any-of loop in VerifyWebhookSignature → this FAIL.
func TestHandler_HandleWebhook_RejectsForeignSecret_A5_008(t *testing.T) {
	secretA := "secret-integration-a"
	secretB := "secret-integration-b"
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"acme/app-a": {
				ID:            uuid.New(),
				RepoOwner:     "acme",
				RepoName:      "app-a",
				WebhookSecret: secretA,
			},
			"acme/app-b": {
				ID:            uuid.New(),
				RepoOwner:     "acme",
				RepoName:      "app-b",
				WebhookSecret: secretB,
			},
		},
		// Trap: any-of ListAllWebhookSecrets must NOT accept cross-secret signatures.
		webhookSecrets: []string{secretA, secretB},
	}
	app := testGitHubHandlerApp(t, repo)

	body := []byte(`{"repository":{"name":"app-b","owner":{"login":"acme"}},"zen":"cross-secret"}`)
	mac := hmac.New(sha256.New, []byte(secretA)) // foreign secret for app-b
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", sig)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode,
		"HMAC signed with foreign integration secret must be rejected (A5-008)")
}

func TestHandler_HandleWebhook_MatchedSecretOK_A5_008(t *testing.T) {
	secret := "webhook-secret-matched"
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"acme/hub": {
				ID:            uuid.New(),
				RepoOwner:     "acme",
				RepoName:      "hub",
				WebhookSecret: secret,
			},
		},
	}
	app := testGitHubHandlerApp(t, repo)

	body := []byte(`{"zen":"test","repository":{"name":"hub","owner":{"login":"acme"}}}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", sig)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
