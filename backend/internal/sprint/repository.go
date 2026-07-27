package sprint

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// Repository определяет интерфейс доступа к данным для модуля спринтов.
type Repository interface {
	// CreateSprint создаёт новый спринт в проекте.
	CreateSprint(ctx context.Context, projectID uuid.UUID, req CreateSprintRequest) (*ent.Sprint, error)
	// GetSprint возвращает спринт по ID.
	GetSprint(ctx context.Context, sprintID uuid.UUID) (*ent.Sprint, error)
	// GetSprintInProject возвращает спринт только если он принадлежит projectID (IDOR guard).
	GetSprintInProject(ctx context.Context, projectID, sprintID uuid.UUID) (*ent.Sprint, error)
	// ListSprints возвращает все спринты проекта.
	ListSprints(ctx context.Context, projectID uuid.UUID) ([]*ent.Sprint, error)
	// UpdateSprint обновляет спринт.
	UpdateSprint(ctx context.Context, sprintID uuid.UUID, req UpdateSprintRequest) (*ent.Sprint, error)
	// UpdateSprintStatus изменяет статус спринта.
	UpdateSprintStatus(ctx context.Context, sprintID uuid.UUID, status sprintEnt.Status) (*ent.Sprint, error)
	// UpdateRetroNotes сохраняет заметки ретроспективы.
	UpdateRetroNotes(ctx context.Context, sprintID uuid.UUID, notes map[string]interface{}) (*ent.Sprint, error)
	// FindActiveSprint находит активный спринт в проекте (если есть).
	FindActiveSprint(ctx context.Context, projectID uuid.UUID) (*ent.Sprint, error)
	// FindOrCreateNextPlanningSprint находит или создаёт следующий спринт в статусе planning.
	FindOrCreateNextPlanningSprint(ctx context.Context, projectID uuid.UUID, afterSprint *ent.Sprint) (*ent.Sprint, error)

	// FindIncompleteIssues находит незавершённые задачи спринта (status.category != done).
	FindIncompleteIssues(ctx context.Context, sprintID uuid.UUID) ([]*ent.Issue, error)
	// MoveIssuesToSprint переносит задачи в целевой спринт.
	MoveIssuesToSprint(ctx context.Context, issueIDs []uuid.UUID, targetSprintID uuid.UUID) error
	// SetIssuesSprint устанавливает sprint_id для задач.
	SetIssuesSprint(ctx context.Context, issueIDs []uuid.UUID, sprintID uuid.UUID) error
	// ClearIssuesSprint убирает sprint_id у задач (возврат в бэклог).
	ClearIssuesSprint(ctx context.Context, issueIDs []uuid.UUID) error

	// GetSprintIssueCount возвращает количество задач в спринте.
	GetSprintIssueCount(ctx context.Context, sprintID uuid.UUID) (int, error)
	// GetSprintIssueCounts возвращает количество задач по списку спринтов одним GROUP BY (A13-006 N+1).
	GetSprintIssueCounts(ctx context.Context, sprintIDs []uuid.UUID) (map[uuid.UUID]int, error)
	// GetBurndownData возвращает данные для графика burndown.
	GetBurndownData(ctx context.Context, sprintID uuid.UUID) (total int, doneCount int, err error)
	// GetVelocityData возвращает данные velocity по завершённым спринтам проекта.
	GetVelocityData(ctx context.Context, projectID uuid.UUID) ([]VelocityPoint, error)

	// CountIssuesInStatus возвращает количество задач в данном статусе.
	CountIssuesInStatus(ctx context.Context, statusID uuid.UUID) (int, error)
	// GetWorkflowStatus возвращает статус workflow по ID.
	GetWorkflowStatus(ctx context.Context, statusID uuid.UUID) (*ent.WorkflowStatus, error)
}

// EntRepository реализует Repository на основе Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт новый Ent-репозиторий спринтов.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// CreateSprint создаёт новый спринт в проекте.
func (r *EntRepository) CreateSprint(ctx context.Context, projectID uuid.UUID, req CreateSprintRequest) (*ent.Sprint, error) {
	return r.client.Sprint.
		Create().
		SetName(req.Name).
		SetNillableGoal(nilIfEmpty(req.Goal)).
		SetStartDate(req.StartDate).
		SetEndDate(req.EndDate).
		SetProjectID(projectID).
		Save(ctx)
}

// MaxSprintIssues — A13-006: hard cap issues при GetSprint (detail).
const MaxSprintIssues = 500

// GetSprint возвращает спринт по ID с загруженными issues (capped).
func (r *EntRepository) GetSprint(ctx context.Context, sprintID uuid.UUID) (*ent.Sprint, error) {
	return r.client.Sprint.
		Query().
		Where(sprintEnt.ID(sprintID)).
		WithIssues(func(q *ent.IssueQuery) {
			q.Limit(MaxSprintIssues)
		}).
		Only(ctx)
}

// GetSprintInProject возвращает спринт по ID в рамках проекта.
func (r *EntRepository) GetSprintInProject(ctx context.Context, projectID, sprintID uuid.UUID) (*ent.Sprint, error) {
	return r.client.Sprint.
		Query().
		Where(
			sprintEnt.ID(sprintID),
			sprintEnt.HasProjectWith(project.ID(projectID)),
		).
		WithIssues(func(q *ent.IssueQuery) {
			q.Limit(MaxSprintIssues)
		}).
		Only(ctx)
}

// ListSprints возвращает все спринты проекта отсортированные по дате начала.
// A13-006: без WithIssues() — issue_count через GetSprintIssueCounts в ListSprints (service).
func (r *EntRepository) ListSprints(ctx context.Context, projectID uuid.UUID) ([]*ent.Sprint, error) {
	return r.client.Sprint.
		Query().
		Where(sprintEnt.HasProjectWith(project.ID(projectID))).
		Order(ent.Desc(sprintEnt.FieldStartDate)).
		All(ctx)
}

// UpdateSprint обновляет спринт.
func (r *EntRepository) UpdateSprint(ctx context.Context, sprintID uuid.UUID, req UpdateSprintRequest) (*ent.Sprint, error) {
	upd := r.client.Sprint.UpdateOneID(sprintID)
	if req.Name != nil {
		upd.SetName(*req.Name)
	}
	if req.Goal != nil {
		upd.SetGoal(*req.Goal)
	}
	if req.StartDate != nil {
		upd.SetStartDate(*req.StartDate)
	}
	if req.EndDate != nil {
		upd.SetEndDate(*req.EndDate)
	}
	return upd.Save(ctx)
}

// UpdateSprintStatus изменяет статус спринта.
func (r *EntRepository) UpdateSprintStatus(ctx context.Context, sprintID uuid.UUID, status sprintEnt.Status) (*ent.Sprint, error) {
	return r.client.Sprint.
		UpdateOneID(sprintID).
		SetStatus(status).
		Save(ctx)
}

// UpdateRetroNotes сохраняет заметки ретроспективы.
func (r *EntRepository) UpdateRetroNotes(ctx context.Context, sprintID uuid.UUID, notes map[string]interface{}) (*ent.Sprint, error) {
	return r.client.Sprint.
		UpdateOneID(sprintID).
		SetRetroNotes(notes).
		Save(ctx)
}

// FindActiveSprint находит активный спринт в проекте.
func (r *EntRepository) FindActiveSprint(ctx context.Context, projectID uuid.UUID) (*ent.Sprint, error) {
	return r.client.Sprint.
		Query().
		Where(
			sprintEnt.StatusEQ(sprintEnt.StatusActive),
			sprintEnt.HasProjectWith(project.ID(projectID)),
		).
		Only(ctx)
}

// FindOrCreateNextPlanningSprint находит или создаёт следующий спринт в статусе planning.
func (r *EntRepository) FindOrCreateNextPlanningSprint(ctx context.Context, projectID uuid.UUID, afterSprint *ent.Sprint) (*ent.Sprint, error) {
	s, err := r.client.Sprint.
		Query().
		Where(
			sprintEnt.StatusEQ(sprintEnt.StatusPlanning),
			sprintEnt.HasProjectWith(project.ID(projectID)),
		).
		Order(ent.Asc(sprintEnt.FieldStartDate)).
		First(ctx)
	if err == nil {
		return s, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	// Создаём новый спринт
	nextStart := afterSprint.EndDate
	nextEnd := nextStart.AddDate(0, 0, 14)
	return r.client.Sprint.
		Create().
		SetName(afterSprint.Name + " (следующий)").
		SetStartDate(nextStart).
		SetEndDate(nextEnd).
		SetProjectID(projectID).
		Save(ctx)
}

// FindIncompleteIssues находит незавершённые задачи спринта.
func (r *EntRepository) FindIncompleteIssues(ctx context.Context, sprintID uuid.UUID) ([]*ent.Issue, error) {
	return r.client.Issue.
		Query().
		Where(
			issue.HasSprintWith(sprintEnt.ID(sprintID)),
			issue.HasStatusWith(
				workflowstatus.CategoryNEQ(workflowstatus.CategoryDone),
			),
		).
		All(ctx)
}

// MoveIssuesToSprint переносит задачи в целевой спринт.
func (r *EntRepository) MoveIssuesToSprint(ctx context.Context, issueIDs []uuid.UUID, targetSprintID uuid.UUID) error {
	return r.client.Issue.
		Update().
		Where(issue.IDIn(issueIDs...)).
		SetSprintID(targetSprintID).
		Exec(ctx)
}

// SetIssuesSprint устанавливает sprint_id для задач.
func (r *EntRepository) SetIssuesSprint(ctx context.Context, issueIDs []uuid.UUID, sprintID uuid.UUID) error {
	return r.client.Issue.
		Update().
		Where(issue.IDIn(issueIDs...)).
		SetSprintID(sprintID).
		Exec(ctx)
}

// ClearIssuesSprint убирает sprint_id у задач.
func (r *EntRepository) ClearIssuesSprint(ctx context.Context, issueIDs []uuid.UUID) error {
	return r.client.Issue.
		Update().
		Where(issue.IDIn(issueIDs...)).
		ClearSprint().
		Exec(ctx)
}

// GetSprintIssueCount возвращает количество задач в спринте.
func (r *EntRepository) GetSprintIssueCount(ctx context.Context, sprintID uuid.UUID) (int, error) {
	return r.client.Issue.
		Query().
		Where(issue.HasSprintWith(sprintEnt.ID(sprintID))).
		Count(ctx)
}

// GetSprintIssueCounts возвращает количество задач для нескольких спринтов одним запросом.
// Паттерн как project.MemberCountBatch — GROUP BY FK, без N× Count.
func (r *EntRepository) GetSprintIssueCounts(ctx context.Context, sprintIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(sprintIDs) == 0 {
		return make(map[uuid.UUID]int), nil
	}

	var rows []struct {
		SprintID uuid.UUID `json:"sprint_issues"`
		Count    int       `json:"count"`
	}
	err := r.client.Issue.Query().
		Where(issue.HasSprintWith(sprintEnt.IDIn(sprintIDs...))).
		GroupBy(issue.SprintColumn).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	counts := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		counts[row.SprintID] = row.Count
	}
	return counts, nil
}

// GetBurndownData возвращает общее количество задач и количество завершённых.
func (r *EntRepository) GetBurndownData(ctx context.Context, sprintID uuid.UUID) (int, int, error) {
	total, err := r.client.Issue.
		Query().
		Where(issue.HasSprintWith(sprintEnt.ID(sprintID))).
		Count(ctx)
	if err != nil {
		return 0, 0, err
	}

	done, err := r.client.Issue.
		Query().
		Where(
			issue.HasSprintWith(sprintEnt.ID(sprintID)),
			issue.HasStatusWith(workflowstatus.CategoryEQ(workflowstatus.CategoryDone)),
		).
		Count(ctx)
	if err != nil {
		return 0, 0, err
	}

	return total, done, nil
}

// GetVelocityData возвращает данные velocity по завершённым спринтам проекта.
func (r *EntRepository) GetVelocityData(ctx context.Context, projectID uuid.UUID) ([]VelocityPoint, error) {
	sprints, err := r.client.Sprint.
		Query().
		Where(
			sprintEnt.StatusEQ(sprintEnt.StatusCompleted),
			sprintEnt.HasProjectWith(project.ID(projectID)),
		).
		Order(ent.Asc(sprintEnt.FieldStartDate)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	points := make([]VelocityPoint, 0, len(sprints))
	for _, s := range sprints {
		committed, _ := r.client.Issue.
			Query().
			Where(issue.HasSprintWith(sprintEnt.ID(s.ID))).
			Count(ctx)

		completed, _ := r.client.Issue.
			Query().
			Where(
				issue.HasSprintWith(sprintEnt.ID(s.ID)),
				issue.HasStatusWith(workflowstatus.CategoryEQ(workflowstatus.CategoryDone)),
			).
			Count(ctx)

		points = append(points, VelocityPoint{
			SprintName: s.Name,
			Committed:  committed,
			Completed:  completed,
		})
	}

	return points, nil
}

// CountIssuesInStatus возвращает количество задач в данном статусе.
func (r *EntRepository) CountIssuesInStatus(ctx context.Context, statusID uuid.UUID) (int, error) {
	return r.client.Issue.
		Query().
		Where(issue.HasStatusWith(workflowstatus.ID(statusID))).
		Count(ctx)
}

// GetWorkflowStatus возвращает статус workflow по ID.
func (r *EntRepository) GetWorkflowStatus(ctx context.Context, statusID uuid.UUID) (*ent.WorkflowStatus, error) {
	return r.client.WorkflowStatus.Get(ctx, statusID)
}

// nilIfEmpty возвращает nil для пустой строки.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
