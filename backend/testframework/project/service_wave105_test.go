package project_test

// Тесты сервиса проектов (волна 105): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entproject "github.com/valerii/hub/backend/ent/project"
)

type countErrRepo struct {
	mockRepo
}

func (r *countErrRepo) MemberCount(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, fmt.Errorf("member count unavailable")
}

type statusErrRepo struct {
	mockRepo
	statusErr error
}

func (r *statusErrRepo) UpdateStatus(_ context.Context, id uuid.UUID, status entproject.Status, setDeletedAt bool) (*ent.Project, error) {
	if r.statusErr != nil {
		return nil, r.statusErr
	}
	return r.mockRepo.UpdateStatus(context.Background(), id, status, setDeletedAt)
}

type teamConstraintRepo struct {
	mockRepo
}

func (r *teamConstraintRepo) CreateTeamWithCreator(_ context.Context, _ uuid.UUID, _, _ string) (*ent.Team, error) {
	return nil, constraintErr()
}

// TestUnarchiveProject_NotFound проверяет сценарий: unarchive project not found.
func TestUnarchiveProject_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.UnarchiveProject(context.Background(), uuid.New())
	require.Error(t, err)
}

// TestUnarchiveProject_UpdateStatusInternalError проверяет сценарий: unarchive project обновление status internal ошибка.
func TestUnarchiveProject_UpdateStatusInternalError(t *testing.T) {
	repo := &statusErrRepo{
		mockRepo:  *newMockRepo(),
		statusErr: fmt.Errorf("db timeout"),
	}
	projectID := uuid.New()
	repo.projects[projectID] = &ent.Project{
		ID: projectID, Name: "Archived", Key: "ARC", Status: entproject.StatusArchived,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	svc := project.NewProjectService(repo, &mockCache{}, nil)

	_, err := svc.UnarchiveProject(context.Background(), projectID)
	require.Error(t, err)
}

// TestDeleteProject_NotFound проверяет сценарий: удаление project not found.
func TestDeleteProject_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	err := svc.DeleteProject(context.Background(), uuid.New())
	require.Error(t, err)
}

// TestRestoreProject_NotFound проверяет сценарий: restore project not found.
func TestRestoreProject_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.RestoreProject(context.Background(), uuid.New())
	require.Error(t, err)
}

// TestArchiveProject_MemberCountFallback проверяет сценарий: archive project участник count fallback.
func TestArchiveProject_MemberCountFallback(t *testing.T) {
	repo := &countErrRepo{mockRepo: *newMockRepo()}
	svc := project.NewProjectService(repo, &mockCache{}, nil)
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Count Fail", "CNT", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	archived, err := svc.ArchiveProject(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 0, archived.MemberCount)
	assert.Equal(t, "archived", archived.Status)
}

// TestCreateTeam_ConstraintConflict проверяет сценарий: создание team constraint conflict.
func TestCreateTeam_ConstraintConflict(t *testing.T) {
	repo := &teamConstraintRepo{mockRepo: *newMockRepo()}
	svc := project.NewProjectService(repo, &mockCache{}, nil)

	_, err := svc.CreateTeam(context.Background(), uuid.New(), "Duplicate", "team")
	require.Error(t, err)
}

type teamFailRepo struct {
	mockRepo
	err error
}

func (r *teamFailRepo) CreateTeamWithCreator(_ context.Context, _ uuid.UUID, _, _ string) (*ent.Team, error) {
	return nil, r.err
}

// TestCreateTeam_InternalError проверяет сценарий: создание team internal ошибка.
func TestCreateTeam_InternalError(t *testing.T) {
	repo := &teamFailRepo{mockRepo: *newMockRepo(), err: fmt.Errorf("db down")}
	svc := project.NewProjectService(repo, &mockCache{}, nil)

	_, err := svc.CreateTeam(context.Background(), uuid.New(), "Squad", "desc")
	require.Error(t, err)
}
