package audit_test

// HTTP regression A1-013: PUT /admin/limits/{key} rejects keys outside the 3-key allowlist (400).

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
)

// TestHandler_UpdateLimit_RejectsUnknownKey_A1_013 — PUT /admin/limits/{key} → 400 outside 3-key allowlist.
func TestHandler_UpdateLimit_RejectsUnknownKey_A1_013(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	rejected := []string{
		"retention.chat_messages_days",
		"retention.allure_results_days",
		"limit.unknown.fake",
		"anything.else",
	}
	for _, key := range rejected {
		t.Run(key, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPut,
				"/api/admin/limits/"+key,
				strings.NewReader(`{"value":"1"}`),
			)
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req, 5*time.Second)
			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		})
	}
}

// TestHandler_UpdateLimit_AllowlistKeysOK_A1_013 — только 3 known keys принимают PUT.
func TestHandler_UpdateLimit_AllowlistKeysOK_A1_013(t *testing.T) {
	adminID := uuid.New()
	repo := &adminMockRepo{adminID: adminID}
	app := setupAuditHandlerTest(t, repo, adminID)

	allowed := []string{
		audit.LimitFileMaxSizeMB,
		audit.LimitUserMaxProjects,
		audit.LimitProjectMaxStorageMB,
	}
	for _, key := range allowed {
		t.Run(key, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPut,
				"/api/admin/limits/"+key,
				strings.NewReader(`{"value":"42"}`),
			)
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req, 5*time.Second)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
