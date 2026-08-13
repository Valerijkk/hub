package sprint_test
import (
	"github.com/valerii/hub/backend/internal/sprint"
	"context"
	"sync"
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
	mu               sync.Mutex
	sprints          map[uuid.UUID]*ent.Sprint
	sprintProject    map[uuid.UUID]uuid.UUID
	issues           map[uuid.UUID]*ent.Issue
	issueProject     map[uuid.UUID]uuid.UUID
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
		issueProject:  make(map[uuid.UUID]uuid.UUID),
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
		Version:   1,
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

func (m *mockRepository) SoftDeleteSprint(_ context.Context, projectID, sprintID uuid.UUID) error {
	s, ok := m.sprints[sprintID]
	if !ok || m.sprintProject[sprintID] != projectID {
		return &ent.NotFoundError{}
	}
	now := time.Now().UTC()
	s.DeletedAt = &now
	return nil
}

func (m *mockRepository) UpdateSprint(_ context.Context, sprintID uuid.UUID, req sprint.UpdateSprintRequest) (*ent.Sprint, error) {
	s, ok := m.sprints[sprintID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if s.Version == 0 {
		s.Version = 1
	}
	if s.Version != req.Version {
		return nil, &ent.NotFoundError{}
	}
	if req.Name != nil {
		s.Name = *req.Name
	}
	if req.Goal != nil {
		s.Goal = *req.Goal
	}
	if req.StartDate != nil {
		s.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		s.EndDate = *req.EndDate
	}
	s.Version++
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

func (m *mockRepository) UpdateRetroNotes(_ context.Context, sprintID uuid.UUID, notes map[string]interface{}, version int) (*ent.Sprint, error) {
	s, ok := m.sprints[sprintID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if version < 1 || s.Version != version {
		return nil, &ent.NotFoundError{}
	}
	s.RetroNotes = notes
	s.Version++
	return s, nil
}

func (m *mockRepository) FindActiveSprint(_ context.Context, projectID uuid.UUID) (*ent.Sprint, error) {
	for id, s := range m.sprints {
		if s.Status == sprintEnt.StatusActive && m.sprintProject[id] == projectID {
			return s, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

// StartSprintAtomic — mock с mutex: два concurrent Start → один active (W1-A4-009).
func (m *mockRepository) StartSprintAtomic(ctx context.Context, projectID, sprintID uuid.UUID) (*ent.Sprint, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := m.GetSprintInProject(ctx, projectID, sprintID); err != nil {
		return nil, err
	}
	existing, err := m.FindActiveSprint(ctx, projectID)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
	} else if existing.ID != sprintID {
		return nil, &sprint.ErrActiveSprintConflict{Name: existing.Name}
	}
	return m.UpdateSprintStatus(ctx, sprintID, sprintEnt.StatusActive)
}

func (m *mockRepository) IssueInProject(_ context.Context, projectID, issueID uuid.UUID) error {
	pid, ok := m.issueProject[issueID]
	if !ok || pid != projectID {
		return sprint.ErrIssueNotInProject
	}
	return nil
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

func (m *mockRepository) CompleteSprintAtomic(ctx context.Context, projectID, sprintID uuid.UUID, afterSprint *ent.Sprint, incompleteIDs []uuid.UUID) (*ent.Sprint, error) {
	if len(incompleteIDs) > 0 {
		next, err := m.FindOrCreateNextPlanningSprint(ctx, projectID, afterSprint)
		if err != nil {
			return nil, err
		}
		if err := m.MoveIssuesToSprint(ctx, incompleteIDs, next.ID); err != nil {
			return nil, err
		}
	}
	return m.UpdateSprintStatus(ctx, sprintID, sprintEnt.StatusCompleted)
}

func (m *mockRepository) SetIssuesSprint(_ context.Context, projectID uuid.UUID, issueIDs []uuid.UUID, sprintID uuid.UUID) error {
	// W1-A4-015 / P3-BE-03: отсутствие в issueProject = как Ent n!=len → ErrIssueNotInProject.
	for _, id := range issueIDs {
		pid, ok := m.issueProject[id]
		if !ok || pid != projectID {
			return sprint.ErrIssueNotInProject
		}
	}
	m.movedIssueIDs = issueIDs
	m.movedToSprint = sprintID
	return nil
}

func (m *mockRepository) ClearIssuesSprint(_ context.Context, _ uuid.UUID, issueIDs []uuid.UUID) error {
	m.clearedIssueIDs = issueIDs
	return nil
}

func (m *mockRepository) GetSprintIssueCount(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockRepository) GetSprintIssueCounts(_ context.Context, sprintIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	counts := make(map[uuid.UUID]int, len(sprintIDs))
	for _, id := range sprintIDs {
		counts[id] = 0
	}
	return counts, nil
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
	projectID := uuid.New()

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
	repo.sprintProject[sprintID] = projectID

	resp, err := svc.StartSprint(ctx, projectID, sprintID)
	require.NoError(t, err)
	assert.Equal(t, "active", resp.Status)
}

// TestStartSprint_RejectsForeignProject_A3_016 — RT-005 sibling: foreign projectID
// + real sprintID must NotFound; same project starts OK.
func TestStartSprint_RejectsForeignProject_A3_016(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	attackerProject := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID:        sprintID,
		Name:      "Victim Sprint",
		Status:    sprintEnt.StatusPlanning,
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = ownerProject

	_, err := svc.StartSprint(ctx, attackerProject, sprintID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
	assert.Equal(t, sprintEnt.StatusPlanning, repo.sprints[sprintID].Status,
		"foreign StartSprint must not activate sprint")

	resp, err := svc.StartSprint(ctx, ownerProject, sprintID)
	require.NoError(t, err)
	assert.Equal(t, "active", resp.Status)
}

// TestGetSprint_RejectsForeignProject_RT005 — RT-005 sibling: foreign projectID + real sprintID → NotFound.
func TestGetSprint_RejectsForeignProject_RT005(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	attackerProject := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Secret Sprint", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = ownerProject

	_, err := svc.GetSprint(ctx, attackerProject, sprintID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")

	resp, err := svc.GetSprint(ctx, ownerProject, sprintID)
	require.NoError(t, err)
	assert.Equal(t, "Secret Sprint", resp.Name)
}

// TestUpdateSprint_RejectsForeignProject_RT005 — RT-005 sibling: foreign Update must not mutate.
func TestUpdateSprint_RejectsForeignProject_RT005(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	attackerProject := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Original", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = ownerProject

	stolen := "Stolen"
	_, err := svc.UpdateSprint(ctx, attackerProject, sprintID, sprint.UpdateSprintRequest{Name: &stolen, Version: 1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
	assert.Equal(t, "Original", repo.sprints[sprintID].Name, "foreign UpdateSprint must not rename")

	okName := "Owned update"
	resp, err := svc.UpdateSprint(ctx, ownerProject, sprintID, sprint.UpdateSprintRequest{Name: &okName, Version: 1})
	require.NoError(t, err)
	assert.Equal(t, "Owned update", resp.Name)
}

// TestUpdateRetroNotes_RejectsForeignProject_RT005 — RT-005 sibling: foreign Retro must not write notes.
func TestUpdateRetroNotes_RejectsForeignProject_RT005(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	attackerProject := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Retro victim", Status: sprintEnt.StatusCompleted,
		StartDate: time.Now().AddDate(0, 0, -14), EndDate: time.Now(),
		Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = ownerProject

	_, err := svc.UpdateRetroNotes(ctx, attackerProject, sprintID, sprint.RetroNotesRequest{
		WentWell: []string{"leak"},
		Version:  1,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
	assert.Nil(t, repo.sprints[sprintID].RetroNotes, "foreign UpdateRetroNotes must not write")

	resp, err := svc.UpdateRetroNotes(ctx, ownerProject, sprintID, sprint.RetroNotesRequest{
		WentWell: []string{"ok"},
		Version:  1,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.RetroNotes["went_well"])
}

// TestGetBurndown_RejectsForeignProject_RT005 — RT-005 sibling: foreign burndown read denied.
func TestGetBurndown_RejectsForeignProject_RT005(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	attackerProject := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Burn", Status: sprintEnt.StatusActive,
		StartDate: time.Now().AddDate(0, 0, -7), EndDate: time.Now().AddDate(0, 0, 7),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = ownerProject

	_, err := svc.GetBurndown(ctx, attackerProject, sprintID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")

	points, err := svc.GetBurndown(ctx, ownerProject, sprintID)
	require.NoError(t, err)
	assert.NotEmpty(t, points)
}

// TestAddIssuesToSprint_RejectsForeignProject_RT005 — RT-005 sibling: foreign AddIssues must not bind.
func TestAddIssuesToSprint_RejectsForeignProject_RT005(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	attackerProject := uuid.New()
	sprintID := uuid.New()
	issueID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Victim", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
	}
	repo.sprintProject[sprintID] = ownerProject
	repo.issueProject[issueID] = ownerProject

	err := svc.AddIssuesToSprint(ctx, attackerProject, sprintID, []uuid.UUID{issueID})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
	assert.Empty(t, repo.movedIssueIDs, "foreign AddIssuesToSprint must not move issues")

	require.NoError(t, svc.AddIssuesToSprint(ctx, ownerProject, sprintID, []uuid.UUID{issueID}))
	assert.Equal(t, []uuid.UUID{issueID}, repo.movedIssueIDs)
}

// TestAddIssuesToSprint_RejectsForeignIssue_RT005 — RT-005: issue UUID из чужого проекта отклоняется.
func TestAddIssuesToSprint_RejectsForeignIssue_RT005(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	ownerProject := uuid.New()
	foreignProject := uuid.New()
	sprintID := uuid.New()
	ownIssue := uuid.New()
	foreignIssue := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Sprint", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
	}
	repo.sprintProject[sprintID] = ownerProject
	repo.issueProject[ownIssue] = ownerProject
	repo.issueProject[foreignIssue] = foreignProject

	err := svc.AddIssuesToSprint(ctx, ownerProject, sprintID, []uuid.UUID{ownIssue, foreignIssue})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не принадлежат проекту")
	assert.Empty(t, repo.movedIssueIDs, "foreign issue UUID must not bind to sprint")

	require.NoError(t, svc.AddIssuesToSprint(ctx, ownerProject, sprintID, []uuid.UUID{ownIssue}))
	assert.Equal(t, []uuid.UUID{ownIssue}, repo.movedIssueIDs)
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
	repo.sprintProject[activeID] = projectID

	// Создаём planning спринт
	newID := uuid.New()
	repo.sprints[newID] = &ent.Sprint{
		ID:        newID,
		Name:      "New Sprint",
		Status:    sprintEnt.StatusPlanning,
		StartDate: time.Now().AddDate(0, 0, 15),
		EndDate:   time.Now().AddDate(0, 0, 28),
	}
	repo.sprintProject[newID] = projectID

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
	projectID := uuid.New()

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
	repo.sprintProject[sprintID] = projectID

	points, err := svc.GetBurndown(ctx, projectID, sprintID)
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
