package audit_test

// C-W19-001 / C-W19-003 / TASK-177: Deactivate/Ban → SessionRevoker (Redis+NATS revoked).
// Mutation: убрать SetSessionRevoker call / RevokeAllUserSessions из deactivateUser →
// mock.calls пуст / fail-closed тест зелёный при deleteSessErr (false-green regression).

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
)

type stubSessionRevoker struct {
	calls []uuid.UUID
	err   error
}

func (s *stubSessionRevoker) RevokeAllUserSessions(_ context.Context, userID uuid.UUID) error {
	s.calls = append(s.calls, userID)
	return s.err
}

// TestDeactivateUser_CallsSessionRevoker_C_W19_001 — Deactivate → RevokeAllUserSessions.
func TestDeactivateUser_CallsSessionRevoker_C_W19_001(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	revoker := &stubSessionRevoker{}
	svc.SetSessionRevoker(revoker)

	target := uuid.New()
	require.NoError(t, svc.DeactivateUser(context.Background(), target))

	require.Contains(t, repo.deactivateIDs, target)
	require.Equal(t, []uuid.UUID{target}, revoker.calls)
	assert.Empty(t, repo.deleteSessIDs, "with SessionRevoker must not use audit.DeleteUserSessions fallback")
}

// TestBanUser_CallsSessionRevoker_C_W19_001 — Ban → тот же kill path через deactivateUser.
func TestBanUser_CallsSessionRevoker_C_W19_001(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	revoker := &stubSessionRevoker{}
	svc.SetSessionRevoker(revoker)

	target := uuid.New()
	admin := uuid.New()
	require.NoError(t, svc.BanUser(context.Background(), target, admin, "abuse", "10.0.0.1", "ua"))

	require.Contains(t, repo.deactivateIDs, target)
	require.Equal(t, []uuid.UUID{target}, revoker.calls)
	assert.Empty(t, repo.deleteSessIDs)
}

// TestDeactivateUser_SessionRevokerFail_FailClosed_C_W19_003 — kill fail → error admin.
func TestDeactivateUser_SessionRevokerFail_FailClosed_C_W19_003(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)
	svc.SetSessionRevoker(&stubSessionRevoker{err: errors.New("redis pipeline down")})

	target := uuid.New()
	err := svc.DeactivateUser(context.Background(), target)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "revoke sessions on deactivate")
	require.Contains(t, repo.deactivateIDs, target, "status уже deactivated; kill fail → admin retry")
}

// TestDeactivateUser_DeleteSessionsFail_FailClosed_C_W19_003 — fallback без revoker тоже fail-closed.
func TestDeactivateUser_DeleteSessionsFail_FailClosed_C_W19_003(t *testing.T) {
	repo := &mockAuditRepo{deleteSessErr: errors.New("db sessions delete failed")}
	svc := audit.NewAuditService(repo, nil) // sessions == nil → DB fallback

	err := svc.DeactivateUser(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "revoke sessions on deactivate")
	require.Len(t, repo.deleteSessIDs, 1)
}
