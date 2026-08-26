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

// TestDeleteUserGDPR_MessageEraserFailAfterCommit_A3W64_001 —
// A3W6-4-001: Mongo после Commit — fail eraser не откатывает HardDelete;
// Ent cascade уже закоммичен (orphan messages — trade-off vs partial erasure).
func TestDeleteUserGDPR_MessageEraserFailAfterCommit_A3W64_001(t *testing.T) {
	repo := &gdprFullCascadeRepo{}
	eraser := &stubMessageEraser{failErr: errGDPRCascadeStep}
	svc := audit.NewAuditService(repo, nil)
	svc.SetMessageEraser(eraser)

	err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	require.Len(t, repo.hardDeleted, 1, "HardDelete already committed before mongo erase")
	require.Len(t, eraser.calls, 1)
}

// TestDeleteUserGDPR_HardDeleteFailKeepsMongoIntact_A3W64_001 —
// мутация: HardDelete fail → eraser не вызван (messages intact).
// Mutation: вернуть DeleteMessagesBySender внутрь WithTx до Commit → eraser.calls непустой / тест краснеет.
func TestDeleteUserGDPR_HardDeleteFailKeepsMongoIntact_A3W64_001(t *testing.T) {
	repo := &gdprHardDeleteFailRepo{}
	eraser := &stubMessageEraser{}
	svc := audit.NewAuditService(repo, nil)
	svc.SetMessageEraser(eraser)

	err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)
	require.Error(t, err)
	assert.Empty(t, eraser.calls, "mongo erase must not run before successful Ent Commit")
	assert.Empty(t, repo.hardDeleted)
}

type gdprHardDeleteFailRepo struct {
	gdprFullCascadeRepo
}

func (r *gdprHardDeleteFailRepo) HardDeleteUser(context.Context, uuid.UUID) error {
	return errGDPRCascadeStep
}

// WithTx обязан пробрасывать *gdprHardDeleteFailRepo, иначе embedded WithTx
// отдаёт *gdprFullCascadeRepo и override HardDeleteUser не вызывается.
func (r *gdprHardDeleteFailRepo) WithTx(_ context.Context, fn func(audit.Repository) error) error {
	return fn(r)
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

type phase151GDPRCommitOrderRepo struct {
	gdprFullCascadeRepo
	steps *[]string
}

func (r *phase151GDPRCommitOrderRepo) HardDeleteUser(_ context.Context, id uuid.UUID) error {
	r.hardDeleted = append(r.hardDeleted, id)
	*r.steps = append(*r.steps, "ent_commit:"+id.String())
	return nil
}

func (r *phase151GDPRCommitOrderRepo) WithTx(_ context.Context, fn func(audit.Repository) error) error {
	return fn(r)
}

type phase151GDPRMessengerDataEraser struct {
	steps *[]string
	calls []uuid.UUID
	err   error
}

func (e *phase151GDPRMessengerDataEraser) EraseMessengerUserData(
	_ context.Context, userID uuid.UUID,
) (audit.MessengerDataEraseResult, error) {
	e.calls = append(e.calls, userID)
	*e.steps = append(*e.steps, "messenger_erase:"+userID.String())
	return audit.MessengerDataEraseResult{
		DeletedMessages: 3, DeletedReadStates: 4, DeletedThreadFollows: 5, PulledResolvedMentions: 6,
	}, e.err
}

func TestPhase151DeleteUserGDPR_InvokesOneFullMessengerEraseAfterEntCommit(t *testing.T) {
	steps := []string{}
	repo := &phase151GDPRCommitOrderRepo{steps: &steps}
	eraser := &phase151GDPRMessengerDataEraser{steps: &steps}
	service := audit.NewAuditService(repo, nil)
	service.SetMessengerDataEraser(eraser)
	targetID := uuid.New()

	err := service.DeleteUserGDPR(context.Background(), uuid.New(), targetID, nil)

	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{targetID}, eraser.calls)
	assert.Equal(t, []string{"ent_commit:" + targetID.String(), "messenger_erase:" + targetID.String()}, steps)
}

func TestPhase151DeleteUserGDPR_MessengerEraseFailureRemainsAfterCommittedEntDelete(t *testing.T) {
	steps := []string{}
	repo := &phase151GDPRCommitOrderRepo{steps: &steps}
	eraser := &phase151GDPRMessengerDataEraser{steps: &steps, err: errors.New("partial messenger erase")}
	service := audit.NewAuditService(repo, nil)
	service.SetMessengerDataEraser(eraser)
	targetID := uuid.New()

	err := service.DeleteUserGDPR(context.Background(), uuid.New(), targetID, nil)

	require.Error(t, err)
	assert.Equal(t, []string{"ent_commit:" + targetID.String(), "messenger_erase:" + targetID.String()}, steps)
	assert.Equal(t, []uuid.UUID{targetID}, eraser.calls, "same full erase operation is the retry boundary")
}
