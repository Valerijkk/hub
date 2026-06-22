package auth_test

// Тесты сервиса аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
)

// TestOrganizationService_ValidateInviteToken проверяет бизнес-логику: организация сервис validate приглашение token.
func TestOrganizationService_ValidateInviteToken(t *testing.T) {
	repo, _, ctx := newAuthRepoEnv(t)
	svc := auth.NewOrganizationService(repo)

	require.ErrorIs(t, svc.ValidateInviteToken(ctx, ""), auth.ErrInvalidInviteToken)
	require.ErrorIs(t, svc.ValidateInviteToken(ctx, "bad"), auth.ErrInvalidInviteToken)

	_, err := repo.CreateOrganizationWithFounder(ctx, "founder-"+uuid.New().String()[:6]+"@test.com", "")
	require.NoError(t, err)
	org, err := repo.GetOrganization(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.ValidateInviteToken(ctx, org.InviteToken))
}

// TestOrganizationService_HasAndGetOrganization проверяет бизнес-логику: организация сервис has and получение организация.
func TestOrganizationService_HasAndGetOrganization(t *testing.T) {
	repo, _, ctx := newAuthRepoEnv(t)
	svc := auth.NewOrganizationService(repo)

	has, err := svc.HasOrganization(ctx)
	require.NoError(t, err)
	assert.False(t, has)

	_, err = repo.CreateOrganizationWithFounder(ctx, "founder@test.com", "")
	require.NoError(t, err)

	has, err = svc.HasOrganization(ctx)
	require.NoError(t, err)
	assert.True(t, has)

	org, err := svc.GetOrganization(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hub", org.Name)
}

// TestOrganizationService_PendingUsersCRUD проверяет бизнес-логику: организация сервис pending users crud.
func TestOrganizationService_PendingUsersCRUD(t *testing.T) {
	repo, _, ctx := newAuthRepoEnv(t)
	svc := auth.NewOrganizationService(repo)

	_, err := repo.CreateOrganizationWithFounder(ctx, "founder@test.com", "")
	require.NoError(t, err)
	org, err := repo.GetOrganization(ctx)
	require.NoError(t, err)

	pending, err := repo.CreatePendingMember(ctx, "pending@test.com", org.ID)
	require.NoError(t, err)

	list, err := svc.ListPendingUsers(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, svc.ApprovePendingUser(ctx, pending.ID))

	toReject, err := repo.CreatePendingMember(ctx, "reject@test.com", org.ID)
	require.NoError(t, err)
	require.NoError(t, svc.RejectPendingUser(ctx, toReject.ID))
}

// TestOrganizationService_PurgeExpiredPending проверяет отклонение при организация сервис очистка просроченный pending.
func TestOrganizationService_PurgeExpiredPending(t *testing.T) {
	repo, client, ctx := newAuthRepoEnv(t)
	svc := auth.NewOrganizationService(repo)

	org, err := client.Organization.Create().SetName("Hub").SetSlug("hub").SetInviteToken("purge-tok").Save(ctx)
	require.NoError(t, err)

	_, err = repo.CreatePendingMember(ctx, "fresh@test.com", org.ID)
	require.NoError(t, err)

	n, err := svc.PurgeExpiredPending(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

// TestIsSystemAdmin проверяет сценарий: is system админ.
func TestIsSystemAdmin(t *testing.T) {
	assert.True(t, auth.IsSystemAdmin(user.SystemRoleFounder))
	assert.True(t, auth.IsSystemAdmin(user.SystemRoleAdmin))
	assert.False(t, auth.IsSystemAdmin(user.SystemRoleMember))
}
