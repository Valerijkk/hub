package imex

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/testrun"
)

// CreateIssueInput — импортированная строка issue.
type CreateIssueInput struct {
	Title           string
	IssueType       issue.IssueType
	Priority        issue.Priority
	IssueNumber     int
	IssueKey        string
	ProjectID       uuid.UUID
	StatusID        uuid.UUID
	ReporterID      *uuid.UUID
	DescriptionText *string
}

// Repository — персистентность import/export (ARCH-01).
type Repository interface {
	AllocateIssueKey(ctx context.Context, projectID uuid.UUID) (issueNumber int, issueKey string, statusID uuid.UUID, err error)
	CreateIssue(ctx context.Context, input CreateIssueInput) (*ent.Issue, error)
	CreateIssueComment(ctx context.Context, issueID, authorID uuid.UUID, body string) error
	SetIssueParent(ctx context.Context, issueID, parentID uuid.UUID) error
	ListProjectIssuesForExport(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error)
	ListProjectIssuesPlain(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error)
	// GetTestRunInProject — прогон только внутри projectID (A5-010 IDOR guard).
	GetTestRunInProject(ctx context.Context, projectID, runID uuid.UUID) (*ent.TestRun, error)
	ListTestRunResults(ctx context.Context, run *ent.TestRun) ([]*ent.TestResult, error)
	// GetSprintInProject — спринт только внутри projectID (A5-010 IDOR guard).
	GetSprintInProject(ctx context.Context, projectID, sprintID uuid.UUID) (*ent.Sprint, error)
	ListSprintIssues(ctx context.Context, sp *ent.Sprint) ([]*ent.Issue, error)
}

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий imex.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) AllocateIssueKey(ctx context.Context, projectID uuid.UUID) (int, string, uuid.UUID, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, "", uuid.Nil, fmt.Errorf("начало транзакции: %w", err)
	}

	p, err := tx.Project.UpdateOneID(projectID).
		AddTaskCounter(1).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, "", uuid.Nil, fmt.Errorf("инкремент счётчика задач: %w", err)
	}

	status, err := tx.Project.Query().
		Where(project.ID(projectID)).
		QueryWorkflowStatuses().
		First(ctx)
	if err != nil {
		_ = tx.Rollback()
		return 0, "", uuid.Nil, fmt.Errorf("получение дефолтного статуса: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, "", uuid.Nil, fmt.Errorf("commit tx: %w", err)
	}

	return p.TaskCounter, fmt.Sprintf("%s-%d", p.Key, p.TaskCounter), status.ID, nil
}

func (r *EntRepository) CreateIssue(ctx context.Context, input CreateIssueInput) (*ent.Issue, error) {
	b := r.client.Issue.Create().
		SetTitle(input.Title).
		SetIssueType(input.IssueType).
		SetPriority(input.Priority).
		SetIssueNumber(input.IssueNumber).
		SetIssueKey(input.IssueKey).
		SetProjectID(input.ProjectID).
		SetStatusID(input.StatusID)
	if input.ReporterID != nil {
		b.SetReporterID(*input.ReporterID)
	}
	if input.DescriptionText != nil && *input.DescriptionText != "" {
		b.SetDescriptionText(*input.DescriptionText)
	}
	return b.Save(ctx)
}

func (r *EntRepository) CreateIssueComment(ctx context.Context, issueID, authorID uuid.UUID, body string) error {
	_, err := r.client.Comment.Create().
		SetBodyText(body).
		SetBodyJSON(map[string]any{
			"type": "doc",
			"content": []any{
				map[string]any{
					"type": "paragraph",
					"content": []any{
						map[string]any{"type": "text", "text": body},
					},
				},
			},
		}).
		SetIssueID(issueID).
		SetAuthorID(authorID).
		Save(ctx)
	return err
}

func (r *EntRepository) SetIssueParent(ctx context.Context, issueID, parentID uuid.UUID) error {
	return r.client.Issue.UpdateOneID(issueID).
		SetParentID(parentID).
		Exec(ctx)
}

func (r *EntRepository) ListProjectIssuesForExport(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error) {
	return r.client.Project.Query().
		Where(project.ID(projectID)).
		QueryIssues().
		Where(issue.DeletedAtIsNil()).
		WithStatus().
		WithAssignee().
		All(ctx)
}

func (r *EntRepository) ListProjectIssuesPlain(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error) {
	return r.client.Project.Query().
		Where(project.ID(projectID)).
		QueryIssues().
		All(ctx)
}

// GetTestRunInProject возвращает прогон по ID только внутри projectID.
// Mutation: снять testrun.HasProjectWith → foreign run находится (A5-010).
func (r *EntRepository) GetTestRunInProject(ctx context.Context, projectID, runID uuid.UUID) (*ent.TestRun, error) {
	return r.client.TestRun.Query().
		Where(
			testrun.ID(runID),
			testrun.HasProjectWith(project.ID(projectID)),
		).
		Only(ctx)
}

func (r *EntRepository) ListTestRunResults(ctx context.Context, run *ent.TestRun) ([]*ent.TestResult, error) {
	return r.client.TestRun.QueryResults(run).
		WithTestCase().
		All(ctx)
}

// GetSprintInProject возвращает спринт по ID только внутри projectID.
// Mutation: снять sprint.HasProjectWith → foreign sprint находится (A5-010).
func (r *EntRepository) GetSprintInProject(ctx context.Context, projectID, sprintID uuid.UUID) (*ent.Sprint, error) {
	return r.client.Sprint.Query().
		Where(
			sprint.ID(sprintID),
			sprint.HasProjectWith(project.ID(projectID)),
		).
		Only(ctx)
}

func (r *EntRepository) ListSprintIssues(ctx context.Context, sp *ent.Sprint) ([]*ent.Issue, error) {
	return r.client.Sprint.QueryIssues(sp).
		WithStatus().
		WithAssignee().
		All(ctx)
}

// IsNotFound — сообщает об ошибках ent not-found.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
