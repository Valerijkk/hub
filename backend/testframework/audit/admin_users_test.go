package audit_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
)

type adminUsersRepo struct {
	adminMockRepo
	users []*ent.User
	err   error
	calls int
}

func (r *adminUsersRepo) ListAdminUsers(context.Context) ([]*ent.User, error) {
	r.calls++
	return r.users, r.err
}

func newManagedUserFixture() *ent.User {
	return &ent.User{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Email:       "admin-user@example.com",
		Name:        "Admin User",
		DisplayName: "Администратор",
		SystemRole:  user.SystemRoleAdmin,
		Status:      user.StatusActive,
		CreatedAt:   time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC),
		AvatarURL:   "https://private.example/avatar.png",
		Phone:       "+7-900-000-00-00",
		TelegramID:  stringPtr("private-telegram-id"),
	}
}

func stringPtr(value string) *string { return &value }

// TestAuditService_ListAdminUsers_SafeDTO проверяет маппинг Ent-модели в
// allowlist DTO без дополнительных полей профиля.
func TestAuditService_ListAdminUsers_SafeDTO(t *testing.T) {
	repo := &adminUsersRepo{users: []*ent.User{newManagedUserFixture()}}
	svc := audit.NewAuditService(repo, nil)

	result, err := svc.ListAdminUsers(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, audit.AdminUserResponse{
		ID:          "11111111-1111-1111-1111-111111111111",
		Email:       "admin-user@example.com",
		Name:        "Admin User",
		DisplayName: "Администратор",
		SystemRole:  "admin",
		Status:      "active",
		CreatedAt:   time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC),
	}, result[0])
	assert.Equal(t, 1, repo.calls)
}

// TestAuditService_ListAdminUsers_RepositoryError проверяет передачу ошибки
// репозитория без формирования частичного ответа.
func TestAuditService_ListAdminUsers_RepositoryError(t *testing.T) {
	repoErr := errors.New("db unavailable")
	repo := &adminUsersRepo{err: repoErr}
	svc := audit.NewAuditService(repo, nil)

	result, err := svc.ListAdminUsers(context.Background())
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, result)
}

// TestHandler_ListAdminUsers_AdminOK_NoSecretFields проверяет рабочий GET и
// точный JSON allowlist для frontend-контракта.
func TestHandler_ListAdminUsers_AdminOK_NoSecretFields(t *testing.T) {
	adminID := uuid.New()
	repo := &adminUsersRepo{users: []*ent.User{newManagedUserFixture()}}
	repo.adminID = adminID
	app := setupAuditHandlerTest(t, repo, adminID)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/admin/users", nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var payload []map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Len(t, payload, 1)

	keys := make([]string, 0, len(payload[0]))
	for key := range payload[0] {
		keys = append(keys, key)
	}
	assert.ElementsMatch(t, []string{
		"id", "email", "name", "display_name", "system_role", "status", "created_at",
	}, keys)
	assert.Equal(t, "admin-user@example.com", payload[0]["email"])
	assert.NotContains(t, string(body), "private-telegram-id")
	assert.NotContains(t, string(body), "+7-900-000-00-00")
	assert.NotContains(t, string(body), "private.example/avatar")
}

// TestHandler_ListAdminUsers_AdminGuard проверяет, что обычный пользователь
// не может прочитать список и репозиторий при этом не вызывается.
func TestHandler_ListAdminUsers_AdminGuard(t *testing.T) {
	repo := &adminUsersRepo{}
	repo.adminID = uuid.New()
	app := setupAuditHandlerTest(t, repo, uuid.New())

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/admin/users", nil))
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.Zero(t, repo.calls)
}

// TestRepository_ListAdminUsers_SelectsFrontendFields проверяет чтение всех
// пользователей и DB-level allowlist: прочая PII остаётся незагруженной.
func TestRepository_ListAdminUsers_SelectsFrontendFields(t *testing.T) {
	repo := newAuditRepoTestEnv(t)
	ctx := context.Background()
	client := audit.ExportRepositoryEntClient(repo)
	createdAt := time.Date(2026, 8, 7, 13, 0, 0, 0, time.UTC)

	created, err := client.User.Create().
		SetEmail("repo-admin@example.com").
		SetName("Repo Admin").
		SetDisplayName("Repo Display").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusDeactivated).
		SetCreatedAt(createdAt).
		SetAvatarURL("https://private.example/repo-avatar.png").
		SetPhone("+7-911-111-11-11").
		Save(ctx)
	require.NoError(t, err)

	items, err := repo.ListAdminUsers(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, created.ID, items[0].ID)
	assert.Equal(t, "repo-admin@example.com", items[0].Email)
	assert.Equal(t, "Repo Admin", items[0].Name)
	assert.Equal(t, "Repo Display", items[0].DisplayName)
	assert.Equal(t, user.SystemRoleAdmin, items[0].SystemRole)
	assert.Equal(t, user.StatusDeactivated, items[0].Status)
	assert.True(t, createdAt.Equal(items[0].CreatedAt))
	assert.Empty(t, items[0].AvatarURL)
	assert.Empty(t, items[0].Phone)
}
