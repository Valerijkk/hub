package sprint_test
import (
	"github.com/valerii/hub/backend/internal/sprint"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockRepository struct {
	sprints          map[uuid.UUID]*ent.Sprint
	sprintProject    map[uuid.UUID]uuid.UUID
	issues           map[uuid.UUID]*ent.Issue
	statuses         map[uuid.UUID]*ent.WorkflowStatus
	movedToSprint    uuid.UUID
	movedIssueIDs    []uuid.UUID
	clearedIssueIDs  []uuid.UUID
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		sprints:       make(map[uuid.UUID]*ent.Sprint),
		sprintProject: make(map[uuid.UUID]uuid.UUID),
		issues:        make(map[uuid.UUID]*ent.Issue),
		statuses:      make(map[uuid.UUID]*ent.WorkflowStatus),
	}
}

func (m *mockRepository) CreateSprint(_ context.Context, projectID uuid.UUID, req sprint.CreateSprintRequest) (*ent.Sprint, error) {
	id := uuid.New()
	s := &ent.Sprint{
		ID:        id,
		Name:      req.Name,
		Goal:      req.Goal,
		Status:    sprintEnt.StatusPlanning,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.sprints[id] = s
	m.sprintProject[id] = projectID
	return s, nil
}

func (m *mockRepository) GetSprint(_ context.Context, sprintID uuid.UUID) (*ent.Sprint, error) {
	s, ok := m.sprints[sprintID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return s, nil
}

func (m *mockRepository) GetSprintInProject(_ context.Context, projectID, sprintID uuid.UUID) (*ent.Sprint, error) {
	s, ok := m.sprints[sprintID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if m.sprintProject[sprintID] != projectID {
		return nil, &ent.NotFoundError{}
	}
	return s, nil
}

func (m *mockRepository) ListSprints(_ context.Context, _ uuid.UUID) ([]*ent.Sprint, error) {
	result := make([]*ent.Sprint, 0, len(m.sprints))
	for _, s := range m.sprints {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockRepository) UpdateSprint(_ context.Context, sprintID uuid.UUID, req sprint.UpdateSprintRequest) (*ent.Sprint, error) {
	s, ok := m.sprints[sprintID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if req.Name != nil {
		s.Name = *req.Name
	}
	return s, nil
}

func (m *mockRepository) UpdateSprintStatus(_ context.Context, sprintID uuid.UUID, status sprintEnt.Status) (*ent.Sprint, error) {
	s, ok := m.sprints[sprintID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	s.Status = status
	return s, nil
}

func (m *mockRepository) UpdateRetroNotes(_ context.Context, sprintID uuid.UUID, notes map[string]interface{}) (*ent.Sprint, error) {
	s, ok := m.sprints[sprintID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	s.RetroNotes = notes
	return s, nil
}

func (m *mockRepository) FindActiveSprint(_ context.Context, _ uuid.UUID) (*ent.Sprint, error) {
	for _, s := range m.sprints {
		if s.Status == sprintEnt.StatusActive {
			return s, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) FindOrCreateNextPlanningSprint(_ context.Context, _ uuid.UUID, afterSprint *ent.Sprint) (*ent.Sprint, error) {
	// Ищем planning спринт
	for _, s := range m.sprints {
		if s.Status == sprintEnt.StatusPlanning {
			return s, nil
		}
	}
	// Создаём новый
	id := uuid.New()
	s := &ent.Sprint{
		ID:        id,
		Name:      afterSprint.Name + " (следующий)",
		Status:    sprintEnt.StatusPlanning,
		StartDate: afterSprint.EndDate,
		EndDate:   afterSprint.EndDate.AddDate(0, 0, 14),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.sprints[id] = s
	return s, nil
}

func (m *mockRepository) FindIncompleteIssues(_ context.Context, sprintID uuid.UUID) ([]*ent.Issue, error) {
	var result []*ent.Issue
	for _, iss := range m.issues {
		if iss.Edges.Sprint != nil && iss.Edges.Sprint.ID == sprintID {
			result = append(result, iss)
		}
	}
	return result, nil
}

func (m *mockRepository) MoveIssuesToSprint(_ context.Context, issueIDs []uuid.UUID, targetSprintID uuid.UUID) error {
	m.movedIssueIDs = issueIDs
	m.movedToSprint = targetSprintID
	return nil
}

func (m *mockRepository) SetIssuesSprint(_ context.Context, issueIDs []uuid.UUID, sprintID uuid.UUID) error {
	m.movedIssueIDs = issueIDs
	m.movedToSprint = sprintID
	return nil
}

func (m *mockRepository) ClearIssuesSprint(_ context.Context, issueIDs []uuid.UUID) error {
	m.clearedIssueIDs = issueIDs
	return nil
}

func (m *mockRepository) GetSprintIssueCount(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockRepository) GetBurndownData(_ context.Context, _ uuid.UUID) (int, int, error) {
	return 10, 3, nil
}

func (m *mockRepository) GetVelocityData(_ context.Context, _ uuid.UUID) ([]sprint.VelocityPoint, error) {
	return []sprint.VelocityPoint{
		{SprintName: "Sprint 1", Committed: 10, Completed: 8},
	}, nil
}

func (m *mockRepository) CountIssuesInStatus(_ context.Context, statusID uuid.UUID) (int, error) {
	ws, ok := m.statuses[statusID]
	if !ok {
		return 0, nil
	}
	// Простой подсчёт по тегу Count в WipLimit для тестов
	_ = ws
	return 5, nil
}

func (m *mockRepository) GetWorkflowStatus(_ context.Context, statusID uuid.UUID) (*ent.WorkflowStatus, error) {
	ws, ok := m.statuses[statusID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return ws, nil
}

// ---------------------------------------------------------------------------
// Тесты
// ---------------------------------------------------------------------------

func TestCreateSprint(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	resp, err := svc.CreateSprint(ctx, uuid.New(), sprint.CreateSprintRequest{
		Name:      "Sprint 1",
		Goal:      "Цель спринта",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 14),
	})
	require.NoError(t, err)
	assert.Equal(t, "Sprint 1", resp.Name)
	assert.Equal(t, "planning", resp.Status)
}

func TestCreateSprint_EndDateBeforeStartDate(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	_, err := svc.CreateSprint(ctx, uuid.New(), sprint.CreateSprintRequest{
		Name:      "Sprint 1",
		StartDate: time.Now().AddDate(0, 0, 14),
		EndDate:   time.Now(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "дата окончания")
}

func TestStartSprint_Success(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	// Создаём спринт
	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID:        sprintID,
		Name:      "Sprint 1",
		Status:    sprintEnt.StatusPlanning,
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	resp, err := svc.StartSprint(ctx, uuid.New(), sprintID)
	require.NoError(t, err)
	assert.Equal(t, "active", resp.Status)
}

func TestStartSprint_RejectsWhenActiveExists(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	// Создаём активный спринт
	activeID := uuid.New()
	repo.sprints[activeID] = &ent.Sprint{
		ID:        activeID,
		Name:      "Active Sprint",
		Status:    sprintEnt.StatusActive,
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 14),
	}

	// Создаём planning спринт
	newID := uuid.New()
	repo.sprints[newID] = &ent.Sprint{
		ID:        newID,
		Name:      "New Sprint",
		Status:    sprintEnt.StatusPlanning,
		StartDate: time.Now().AddDate(0, 0, 15),
		EndDate:   time.Now().AddDate(0, 0, 28),
	}

	_, err := svc.StartSprint(ctx, projectID, newID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "активный спринт")
}

func TestCompleteSprint_MovesIncompleteIssues(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID:        sprintID,
		Name:      "Sprint 1",
		Status:    sprintEnt.StatusActive,
		StartDate: time.Now().AddDate(0, 0, -14),
		EndDate:   time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = projectID

	// Незавершённая задача
	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{
		ID: issueID,
		Edges: ent.IssueEdges{
			Sprint: repo.sprints[sprintID],
		},
	}

	resp, err := svc.CompleteSprint(ctx, projectID, sprintID)
	require.NoError(t, err)
	assert.Equal(t, "completed", resp.Status)
	assert.Len(t, repo.movedIssueIDs, 1)
	assert.Equal(t, issueID, repo.movedIssueIDs[0])
}

// TestCompleteSprint_RejectsForeignProject — A3-001: нельзя завершить чужой спринт и утащить issues.
func TestCompleteSprint_RejectsForeignProject(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	attackerProject := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID:        sprintID,
		Name:      "Victim Sprint",
		Status:    sprintEnt.StatusActive,
		StartDate: time.Now().AddDate(0, 0, -14),
		EndDate:   time.Now(),
	}
	repo.sprintProject[sprintID] = ownerProject

	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{
		ID: issueID,
		Edges: ent.IssueEdges{Sprint: repo.sprints[sprintID]},
	}

	_, err := svc.CompleteSprint(ctx, attackerProject, sprintID)
	require.Error(t, err)
	assert.Empty(t, repo.movedIssueIDs, "foreign CompleteSprint must not move issues")
	assert.Equal(t, sprintEnt.StatusActive, repo.sprints[sprintID].Status)
}

func TestWIPLimit_BlocksWhenLimitReached(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	statusID := uuid.New()
	repo.statuses[statusID] = &ent.WorkflowStatus{
		ID:       statusID,
		Name:     "In Progress",
		Category: workflowstatus.CategoryInProgress,
		WipLimit: 5,
	}

	err := svc.CheckWIPLimit(ctx, statusID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WIP-лимит достигнут")
}

func TestWIPLimit_AllowsWhenNoLimit(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	statusID := uuid.New()
	repo.statuses[statusID] = &ent.WorkflowStatus{
		ID:       statusID,
		Name:     "In Progress",
		Category: workflowstatus.CategoryInProgress,
		WipLimit: 0,
	}

	err := svc.CheckWIPLimit(ctx, statusID)
	require.NoError(t, err)
}

func TestWIPLimit_AllowsWhenBelowLimit(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	statusID := uuid.New()
	repo.statuses[statusID] = &ent.WorkflowStatus{
		ID:       statusID,
		Name:     "In Progress",
		Category: workflowstatus.CategoryInProgress,
		WipLimit: 10,
	}

	err := svc.CheckWIPLimit(ctx, statusID)
	require.NoError(t, err)
}

func TestGetBurndown(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID:        sprintID,
		Name:      "Sprint 1",
		Status:    sprintEnt.StatusActive,
		StartDate: time.Now().AddDate(0, 0, -7),
		EndDate:   time.Now().AddDate(0, 0, 7),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	points, err := svc.GetBurndown(ctx, sprintID)
	require.NoError(t, err)
	assert.NotEmpty(t, points)
	// Первая точка — полное количество задач
	assert.Equal(t, 10, points[0].Remaining)
	assert.Equal(t, 10, points[0].Ideal)
	// Последняя ideal точка — 0
	assert.Equal(t, 0, points[len(points)-1].Ideal)
}

func TestGetVelocity(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	points, err := svc.GetVelocity(ctx, uuid.New())
	require.NoError(t, err)
	assert.Len(t, points, 1)
	assert.Equal(t, "Sprint 1", points[0].SprintName)
	assert.Equal(t, 10, points[0].Committed)
	assert.Equal(t, 8, points[0].Completed)
}
