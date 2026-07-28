package audit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/audit"
)

var errGDPRCascadeStep = errors.New("gdpr cascade step failed")

// stubMessageEraser — MessageEraser для P1-BE-34 (Mongo cascade).
type stubMessageEraser struct {
	calls   []uuid.UUID
	failErr error
}

func (e *stubMessageEraser) DeleteMessagesBySender(_ context.Context, senderID uuid.UUID) (int64, error) {
	e.calls = append(e.calls, senderID)
	if e.failErr != nil {
		return 0, e.failErr
	}
	return 3, nil
}

type gdprFullCascadeRepo struct {
	gdprStubRepo
	hardDeleted []uuid.UUID
	files       int
	projects    int
	channels    int
	teams       int
	scrubbed    int
	failFiles   bool
}

func (r *gdprFullCascadeRepo) DeleteUserFiles(context.Context, uuid.UUID) (int, error) {
	if r.failFiles {
		return 0, errGDPRCascadeStep
	}
	r.files++
	return 1, nil
}
func (r *gdprFullCascadeRepo) DeleteUserProjectMemberships(context.Context, uuid.UUID) (int, error) {
	r.projects++
	return 1, nil
}
func (r *gdprFullCascadeRepo) DeleteUserChannelMemberships(context.Context, uuid.UUID) (int, error) {
	r.channels++
	return 1, nil
}
func (r *gdprFullCascadeRepo) DeleteUserTeamMemberships(context.Context, uuid.UUID) (int, error) {
	r.teams++
	return 1, nil
}
func (r *gdprFullCascadeRepo) ScrubUserAuditPII(context.Context, uuid.UUID) (int, error) {
	r.scrubbed++
	return 2, nil
}
func (r *gdprFullCascadeRepo) HardDeleteUser(_ context.Context, id uuid.UUID) error {
	r.hardDeleted = append(r.hardDeleted, id)
	return nil
}

func (r *gdprFullCascadeRepo) WithTx(_ context.Context, fn func(audit.Repository) error) error {
	return fn(r)
}

// TestDeleteUserGDPR_CascadeIncludesFilesMembershipsScrub_P1BE34 —
// после успешного erase вызваны расширенные шаги; HardDelete один раз.
func TestDeleteUserGDPR_CascadeIncludesFilesMembershipsScrub_P1BE34(t *testing.T) {
	repo := &gdprFullCascadeRepo{}
	eraser := &stubMessageEraser{}
	svc := audit.NewAuditService(repo, nil)
	svc.SetMessageEraser(eraser)

	target := uuid.New()
	actor := uuid.New()
	require.NoError(t, svc.DeleteUserGDPR(context.Background(), actor, target, nil))

	assert.Equal(t, 1, repo.files)
	assert.Equal(t, 1, repo.projects)
	assert.Equal(t, 1, repo.channels)
	assert.Equal(t, 1, repo.teams)
	assert.Equal(t, 1, repo.scrubbed)
	require.Len(t, eraser.calls, 1)
	assert.Equal(t, target, eraser.calls[0])
	require.Len(t, repo.hardDeleted, 1)
	assert.Equal(t, target, repo.hardDeleted[0])
}

// TestDeleteUserGDPR_MessageEraserFailAbortsHardDelete_P1BE34 —
// мутация: fail Mongo erase → HardDelete не вызван.
func TestDeleteUserGDPR_MessageEraserFailAbortsHardDelete_P1BE34(t *testing.T) {
	repo := &gdprFullCascadeRepo{}
	eraser := &stubMessageEraser{failErr: errGDPRCascadeStep}
	svc := audit.NewAuditService(repo, nil)
	svc.SetMessageEraser(eraser)

	err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.Empty(t, repo.hardDeleted)
}

// TestDeleteUserGDPR_FilesFailAbortsHardDelete_P1BE34 —
// мутация: fail DeleteUserFiles → HardDelete не вызван.
func TestDeleteUserGDPR_FilesFailAbortsHardDelete_P1BE34(t *testing.T) {
	repo := &gdprFullCascadeRepo{failFiles: true}
	svc := audit.NewAuditService(repo, nil)

	err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.Empty(t, repo.hardDeleted)
}
