//go:build hubtest

package audit_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/audit"
)

type adminErrRepo struct {
	adminMockRepo
	err error
}

func (m *adminErrRepo) IsSystemAdmin(context.Context, uuid.UUID) (bool, error) {
	return false, m.err
}

// Mutation: снова мапить любой IsSystemAdmin err → 401 → assert StatusInternalServerError краснеет.
func TestHandler_requireAdmin_RepoError_500_P3BE08(t *testing.T) {
	adminID := uuid.New()
	repo := &adminErrRepo{adminMockRepo: adminMockRepo{adminID: adminID}, err: errors.New("db down")}
	svc := audit.NewAuditService(repo, nil)
	h := audit.NewHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: func(c fiber.Ctx, err error) error {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).SendString(e.Message)
		}
		return c.Status(500).SendString(err.Error())
	}})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", adminID.String())
		return c.Next()
	}
	h.RegisterRoutesSimple(api, authMw)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "администратора")
	assert.NotContains(t, string(body), "пользователь не найден")
}

func TestHandler_requireAdmin_UserNotFound_401_P3BE08(t *testing.T) {
	adminID := uuid.New()
	repo := &adminErrRepo{adminMockRepo: adminMockRepo{adminID: adminID}, err: &ent.NotFoundError{}}
	svc := audit.NewAuditService(repo, nil)
	h := audit.NewHandler(svc)

	app := fiber.New(fiber.Config{ErrorHandler: func(c fiber.Ctx, err error) error {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).SendString(e.Message)
		}
		return c.Status(500).SendString(err.Error())
	}})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", adminID.String())
		return c.Next()
	}
	h.RegisterRoutesSimple(api, authMw)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
