package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/gitlab"
)

// TestRedactPipelineVariables_P1_BE_36 — secret var → payload без plaintext values.
// Mutation: PublishJSON с req.Variables as-is → assert FAIL.
func TestRedactPipelineVariables_P1_BE_36(t *testing.T) {
	secret := "super-secret-token-value"
	redacted := gitlab.ExportRedactPipelineVariables(map[string]string{
		"CI_DEPLOY_TOKEN": secret,
		"ENV":             "prod",
	})
	require.NotNil(t, redacted)
	assert.Equal(t, "***", redacted["CI_DEPLOY_TOKEN"])
	assert.Equal(t, "***", redacted["ENV"])
	assert.NotContains(t, redacted["CI_DEPLOY_TOKEN"], secret)
	for _, v := range redacted {
		assert.NotEqual(t, secret, v)
	}
	assert.Nil(t, gitlab.ExportRedactPipelineVariables(nil))
	assert.Nil(t, gitlab.ExportRedactPipelineVariables(map[string]string{}))
}
