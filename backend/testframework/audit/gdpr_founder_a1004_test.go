package audit_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
)

// founderGDPRRepo — stub с настраиваемой ролью target для A1-004.
type founderGDPRRepo struct {
	gdprStubRepo
	targetRole user.SystemRole
	hardDeleted []uuid.UUID
}

func (r *founderGDPRRepo) GetUserSystemRole(_ context.Context, _ uuid.UUID) (user.SystemRole, error) {
	return r.targetRole, nil
}

func (r *founderGDPRRepo) HardDeleteUser(_ context.Context, id uuid.UUID) error {
	r.hardDeleted = append(r.hardDeleted, id)
	return nil
}

// TestDeleteUserGDPR_RejectsFounderByAdmin_A1_004 — admin не может стереть founder.
func TestDeleteUserGDPR_RejectsFounderByAdmin_A1_004(t *testing.T) {
	repo := &founderGDPRRepo{targetRole: user.SystemRoleFounder}
	svc := audit.NewAuditService(repo, nil)
	adminID := uuid.New()
	founderID := uuid.New()

	err := svc.DeleteUserGDPR(context.Background(), adminID, founderID, nil)
	require.ErrorIs(t, err, audit.ErrCannotDeleteFounder)
	assert.Empty(t, repo.hardDeleted)
}

// TestDeleteUserGDPR_AllowsFounderSelfErase_A1_004 — founder может удалить себя (GDPR).
func TestDeleteUserGDPR_AllowsFounderSelfErase_A1_004(t *testing.T) {
	repo := &founderGDPRRepo{targetRole: user.SystemRoleFounder}
	svc := audit.NewAuditService(repo, nil)
	founderID := uuid.New()

	err := svc.DeleteUserGDPR(context.Background(), founderID, founderID, nil)
	require.NoError(t, err)
	require.Len(t, repo.hardDeleted, 1)
	assert.Equal(t, founderID, repo.hardDeleted[0])
}

// compile-time: ent import used by gdprStubRepo embedding
var _ = ent.AuditLog{}
