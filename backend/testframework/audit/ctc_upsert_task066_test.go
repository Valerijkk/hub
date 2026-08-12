package audit_test

import (
	"context"
	"crypto/sha256"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/systemsetting"
	"github.com/valerii/hub/backend/internal/audit"
)

// TestAlertWebhook_CTCFixedDigest_TASK066 — W6-A4-010: fixed-size digest CTC (не raw-string).
func TestAlertWebhook_CTCFixedDigest_TASK066(t *testing.T) {
	assert.True(t, audit.ExportConstantTimeEqualDigest("secret", "secret"))
	assert.False(t, audit.ExportConstantTimeEqualDigest("secret", "secre"))
	assert.False(t, audit.ExportConstantTimeEqualDigest("short", "much-longer-secret-value"))

	// Digest-sensitive: CTC operands всегда sha256.Size — raw CTC kill даёт len(a) ≠ Size.
	assert.Equal(t, sha256.Size, audit.ExportConstantTimeEqualDigestOperandLen("short", "much-longer-secret-value"))
	assert.Equal(t, sha256.Size, audit.ExportConstantTimeEqualDigestOperandLen("x", "test-alert-secret"))
	assert.Equal(t, sha256.Size, audit.ExportConstantTimeEqualDigestOperandLen("secret", "secret"))

	repo := &mockAuditRepo{}
	h := audit.NewHandler(audit.NewAuditService(repo, nil))
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterAlertWebhook(api, "test-alert-secret")

	req := httptest.NewRequest(fiber.MethodPost, "/api/webhooks/alertmanager", nil)
	req.Header.Set("Authorization", "Bearer x") // другая длина
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	assert.Empty(t, repo.logs)
}

// TestUpsertSystemSetting_Concurrent_OneRow_TASK066 — W6-A4-012: concurrent Upsert → 1 row.
func TestUpsertSystemSetting_Concurrent_OneRow_TASK066(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()
	key := "limit.concurrent.task066"

	const n = 12
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = repo.UpsertSystemSetting(ctx, key, "v", "concurrent")
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		require.NoError(t, err, "goroutine %d", i)
	}

	val, err := repo.GetSystemSetting(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "v", val)

	count, err := audit.ExportRepositoryEntClient(repo).SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
