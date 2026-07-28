package schema_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/enttest"
	"context"
)

// TestSensitiveFields_HiddenInString — W8-A2-012/013, W8-A4-006/007 / P2-BE-09:
// Sensitive() скрывает секреты из Ent String().
func TestSensitiveFields_HiddenInString(t *testing.T) {
	dsn := "file:sensitive_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	org, err := client.Organization.Create().
		SetName("Hub Org").
		SetSlug("hub-org-" + uuid.New().String()[:8]).
		SetInviteToken("invite-token-secret-value").
		Save(ctx)
	require.NoError(t, err)
	assert.NotContains(t, org.String(), "invite-token-secret-value")

	u, err := client.User.Create().
		SetEmail("push-" + uuid.New().String()[:8] + "@example.com").
		SetName("Push").
		Save(ctx)
	require.NoError(t, err)

	sub, err := client.PushSubscription.Create().
		SetUserID(u.ID).
		SetEndpoint("https://push.example.com/sub").
		SetP256dhKey("p256dh-secret-key").
		SetAuthKey("auth-secret-key").
		Save(ctx)
	require.NoError(t, err)
	assert.NotContains(t, sub.String(), "p256dh-secret-key")
	assert.NotContains(t, sub.String(), "auth-secret-key")

	proj, err := client.Project.Create().SetName("GH").SetKey("GH1").Save(ctx)
	require.NoError(t, err)

	gh, err := client.GitHubIntegration.Create().
		SetRepoOwner("acme").
		SetRepoName("hub").
		SetAccessToken("ghp_token").
		SetWebhookSecret("gh-webhook-secret").
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	assert.NotContains(t, gh.String(), "gh-webhook-secret")
	assert.NotContains(t, gh.String(), "ghp_token")

	gl, err := client.GitLabIntegration.Create().
		SetProjectPath("group/repo").
		SetAccessToken("glpat_token").
		SetWebhookSecret("gl-webhook-secret").
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	assert.NotContains(t, gl.String(), "gl-webhook-secret")
	assert.NotContains(t, gl.String(), "glpat_token")
}
