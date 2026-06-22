package auth_test

// Тесты репозитория аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
)

// TestEntRepository_OrganizationPendingFlow проверяет слой репозитория: ent репозиторий организация pending поток.
func TestEntRepository_OrganizationPendingFlow(t *testing.T) {
	repo, client, ctx := newAuthRepoEnv(t)

	founder, err := repo.CreateOrganizationWithFounder(ctx, "founder-"+uuid.New().String()[:6]+"@test.com", "")
	require.NoError(t, err)
	assert.Equal(t, user.SystemRoleFounder, founder.SystemRole)

	org, err := repo.GetOrganization(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, org.InviteToken)

	ok, err := repo.OrganizationExistsByInviteToken(ctx, org.InviteToken)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.OrganizationExistsByInviteToken(ctx, "invalid-token")
	require.NoError(t, err)
	assert.False(t, ok)

	pending, err := repo.CreatePendingMember(ctx, "pending-"+uuid.New().String()[:6]+"@test.com", org.ID)
	require.NoError(t, err)
	assert.Equal(t, user.StatusPending, pending.Status)

	active, err := repo.CreateActiveMember(ctx, "active-"+uuid.New().String()[:6]+"@test.com", org.ID)
	require.NoError(t, err)
	assert.Equal(t, user.StatusActive, active.Status)

	list, err := repo.ListPendingUsers(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, pending.ID, list[0].ID)

	require.NoError(t, repo.ApprovePendingUser(ctx, pending.ID))
	approved, err := client.User.Get(ctx, pending.ID)
	require.NoError(t, err)
	assert.Equal(t, user.StatusActive, approved.Status)

	toReject, err := repo.CreatePendingMember(ctx, "reject-"+uuid.New().String()[:6]+"@test.com", org.ID)
	require.NoError(t, err)
	require.NoError(t, repo.RejectPendingUser(ctx, toReject.ID))
	_, err = client.User.Get(ctx, toReject.ID)
	assert.True(t, auth.IsNotFound(err))
}

// TestEntRepository_ApproveReject_NotFound проверяет слой репозитория: ent репозиторий approve reject not found.
func TestEntRepository_ApproveReject_NotFound(t *testing.T) {
	repo, client, ctx := newAuthRepoEnv(t)

	active, err := client.User.Create().
		SetEmail("active-" + uuid.New().String()[:6] + "@test.com").
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	err = repo.ApprovePendingUser(ctx, active.ID)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)

	err = repo.RejectPendingUser(ctx, uuid.New())
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}

// TestEntRepository_PurgeExpiredPending проверяет отклонение при ent репозиторий очистка просроченный pending.
func TestEntRepository_PurgeExpiredPending(t *testing.T) {
	repo, client, ctx := newAuthRepoEnv(t)

	org, err := client.Organization.Create().SetName("Hub").SetSlug("hub").SetInviteToken("tok123").Save(ctx)
	require.NoError(t, err)

	oldTime := time.Now().UTC().Add(-8 * 24 * time.Hour)
	_, err = client.User.Create().
		SetEmail("expired-" + uuid.New().String()[:6] + "@test.com").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusPending).
		SetOrganizationID(org.ID).
		SetPendingSince(oldTime).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.CreatePendingMember(ctx, "fresh-"+uuid.New().String()[:6]+"@test.com", org.ID)
	require.NoError(t, err)

	n, err := repo.PurgeExpiredPending(ctx, time.Now().UTC().Add(-7*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	remaining, err := repo.ListPendingUsers(ctx)
	require.NoError(t, err)
	assert.Len(t, remaining, 1)
}

// TestEntRepository_HasOrganization проверяет слой репозитория: ent репозиторий has организация.
func TestEntRepository_HasOrganization(t *testing.T) {
	repo, _, ctx := newAuthRepoEnv(t)

	n, err := repo.CountOrganizations(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	_, err = repo.CreateOrganizationWithFounder(ctx, "founder@test.com", "")
	require.NoError(t, err)

	n, err = repo.CountOrganizations(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}
