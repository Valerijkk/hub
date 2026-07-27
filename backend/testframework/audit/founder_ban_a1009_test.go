package audit_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
)

// founderBanRepo — stub роли target для A1-009 (ban/deactivate founder).
type founderBanRepo struct {
	gdprStubRepo
	targetRole    user.SystemRole
	deactivateIDs []uuid.UUID
}

func (r *founderBanRepo) GetUserSystemRole(_ context.Context, _ uuid.UUID) (user.SystemRole, error) {
	return r.targetRole, nil
}

func (r *founderBanRepo) DeactivateUser(_ context.Context, id uuid.UUID) error {
	r.deactivateIDs = append(r.deactivateIDs, id)
	return nil
}

// TestDeactivateUser_RejectsFounder_A1_009 — admin не может деактивировать founder.
func TestDeactivateUser_RejectsFounder_A1_009(t *testing.T) {
	repo := &founderBanRepo{targetRole: user.SystemRoleFounder}
	svc := audit.NewAuditService(repo, nil)

	err := svc.DeactivateUser(context.Background(), uuid.New())
	require.ErrorIs(t, err, audit.ErrCannotModifyFounder)
	assert.Empty(t, repo.deactivateIDs)
}

// TestBanUser_RejectsFounder_A1_009 — admin не может забанить founder.
func TestBanUser_RejectsFounder_A1_009(t *testing.T) {
	repo := &founderBanRepo{targetRole: user.SystemRoleFounder}
	svc := audit.NewAuditService(repo, nil)
	adminID := uuid.New()
	founderID := uuid.New()

	err := svc.BanUser(context.Background(), founderID, adminID, "abuse", "127.0.0.1", "test")
	require.ErrorIs(t, err, audit.ErrCannotModifyFounder)
	assert.Empty(t, repo.deactivateIDs)
}

// TestDeactivateUser_AllowsMember_A1_009 — member по-прежнему деактивируется.
func TestDeactivateUser_AllowsMember_A1_009(t *testing.T) {
	repo := &founderBanRepo{targetRole: user.SystemRoleMember}
	svc := audit.NewAuditService(repo, nil)
	target := uuid.New()

	require.NoError(t, svc.DeactivateUser(context.Background(), target))
	require.Contains(t, repo.deactivateIDs, target)
}
