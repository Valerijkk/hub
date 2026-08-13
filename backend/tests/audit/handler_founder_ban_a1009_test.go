package audit_test

// HTTP regression A1-009: ban/deactivate founder → 403 (service coverage already in founder_ban_a1009_test.go).

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
)

// founderHTTPRepo — adminMockRepo с фиксированной system_role target (A1-009 HTTP).
type founderHTTPRepo struct {
	adminMockRepo
	targetRole user.SystemRole
}

func (m *founderHTTPRepo) GetUserSystemRole(_ context.Context, _ uuid.UUID) (user.SystemRole, error) {
	return m.targetRole, nil
}

// TestHandler_BanUser_FounderForbidden_A1_009 — POST /admin/users/:id/ban → 403 for founder.
func TestHandler_BanUser_FounderForbidden_A1_009(t *testing.T) {
	adminID := uuid.New()
	founderID := uuid.New()
	repo := &founderHTTPRepo{
		adminMockRepo: adminMockRepo{adminID: adminID},
		targetRole:    user.SystemRoleFounder,
	}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users/"+founderID.String()+"/ban",
		strings.NewReader(`{"reason":"abuse"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.Empty(t, repo.deactivateIDs, "founder must not be deactivated")
}

// TestHandler_DeactivateUser_FounderForbidden_A1_009 — POST /admin/users/:id/deactivate → 403 for founder.
func TestHandler_DeactivateUser_FounderForbidden_A1_009(t *testing.T) {
	adminID := uuid.New()
	founderID := uuid.New()
	repo := &founderHTTPRepo{
		adminMockRepo: adminMockRepo{adminID: adminID},
		targetRole:    user.SystemRoleFounder,
	}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+founderID.String()+"/deactivate", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.Empty(t, repo.deactivateIDs, "founder must not be deactivated")
}
