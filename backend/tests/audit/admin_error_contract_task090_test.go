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

// listFailRepo — ListAuditLogs всегда возвращает infra-ошибку с секретом в тексте.
type listFailRepo struct {
	adminMockRepo
	leak error
}

func (m *listFailRepo) ListAuditLogs(context.Context, audit.AuditLogFilters) ([]*ent.AuditLog, int, error) {
	return nil, 0, m.leak
}

// TestHandler_ListAuditLogs_Fixed500Message_A2W6A4007 — A2W6-A4-007:
// admin 500 не отдаёт err.Error() (утечка internals).
// Мутация: снова fiber.NewError(500, err.Error()) → assert.NotContains(leak) краснеет.
func TestHandler_ListAuditLogs_Fixed500Message_A2W6A4007(t *testing.T) {
	const leak = "sql: connection refused secret=hunter2"
	adminID := uuid.New()
	repo := &listFailRepo{
		adminMockRepo: adminMockRepo{adminID: adminID},
		leak:          errors.New(leak),
	}
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
	assert.Contains(t, string(body), "не удалось загрузить журнал аудита")
	assert.NotContains(t, string(body), leak)
	assert.NotContains(t, string(body), "hunter2")
}

// restoreFailRepo — RestoreIssue падает с leak-сообщением.
type restoreFailRepo struct {
	adminMockRepo
	leak error
}

func (m *restoreFailRepo) RestoreIssue(context.Context, uuid.UUID) error {
	return m.leak
}

func TestHandler_RestoreItem_Fixed500Message_A2W6A4007(t *testing.T) {
	const leak = "pq: relation issues does not exist token=abc"
	adminID := uuid.New()
	repo := &restoreFailRepo{
		adminMockRepo: adminMockRepo{adminID: adminID},
		leak:          errors.New(leak),
	}
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

	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/trash/issue/"+id.String()+"/restore", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "не удалось восстановить элемент")
	assert.NotContains(t, string(body), leak)
	assert.NotContains(t, string(body), "token=abc")
}
