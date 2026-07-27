package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/activitylog"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/callparticipant"
	"github.com/valerii/hub/backend/ent/dashboardlayout"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// Repository — персистентность dashboard (ARCH-01).
type Repository interface {
	ListMyOpenTasks(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.Issue, error)
	ListUserActiveCalls(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.Call, error)
	ListUserRecentTestRuns(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.TestRun, error)
	ListUserDeadlineIssues(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]*ent.Issue, error)
	ListUserTimeLogsBetween(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*ent.TimeLog, error)
	ListProjectIssuesWithStatusAssignee(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error)
	ListProjectCompletedIssuesSince(ctx context.Context, projectID uuid.UUID, since time.Time) ([]*ent.Issue, error)
	CountProjectCompletedIssuesInRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) (int, error)
	ListProjectRecentTestRuns(ctx context.Context, projectID uuid.UUID, limit int) ([]*ent.TestRun, error)
	ListProjectActivityLogs(ctx context.Context, projectID uuid.UUID, limit int) ([]*ent.ActivityLog, error)
	CountUsers(ctx context.Context) (int, error)
	CountDistinctUsersFromSessionsSince(ctx context.Context, since time.Time) (int, error)
	CountProjects(ctx context.Context) (int, error)
	CountIssues(ctx context.Context) (int, error)
	ListProjectsWithIssues(ctx context.Context) ([]*ent.Project, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*ent.User, error)
	// IsProjectMember — A7-001: GetProjectStats требует membership до cache/DB.
	IsProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error)
	GetDashboardLayout(ctx context.Context, userID uuid.UUID, dashboardType string, projectID *uuid.UUID) (*ent.DashboardLayout, error)
	SaveDashboardLayout(ctx context.Context, userID uuid.UUID, req SaveLayoutRequest) error
}

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий dashboard.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) ListMyOpenTasks(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(
			issue.HasAssigneeWith(user.ID(userID)),
			issue.HasStatusWith(workflowstatus.CategoryNEQ(workflowstatus.CategoryDone)),
			issue.DeletedAtIsNil(),
		).
		WithStatus().
		Order(ent.Desc(issue.FieldPriority)).
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) ListUserActiveCalls(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.Call, error) {
	return r.client.Call.Query().
		Where(
			call.Or(
				call.CreatedBy(userID),
				call.HasParticipantsWith(callparticipant.UserIDEQ(userID)),
			),
		).
		WithChannel().
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) ListUserRecentTestRuns(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.TestRun, error) {
	return r.client.TestRun.Query().
		Where(testrun.HasCreatedByWith(user.ID(userID))).
		Order(ent.Desc(testrun.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) ListUserDeadlineIssues(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(
			issue.HasAssigneeWith(user.ID(userID)),
			issue.DueDateGTE(from),
			issue.DueDateLTE(to),
			issue.DeletedAtIsNil(),
		).
		WithStatus().
		Order(ent.Asc(issue.FieldDueDate)).
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) ListUserTimeLogsBetween(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*ent.TimeLog, error) {
	return r.client.TimeLog.Query().
		Where(
			timelog.HasUserWith(user.ID(userID)),
			timelog.CreatedAtGTE(from),
			timelog.CreatedAtLTE(to),
		).
		All(ctx)
}

func (r *EntRepository) ListProjectIssuesWithStatusAssignee(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(
			issue.HasProjectWith(project.ID(projectID)),
			issue.DeletedAtIsNil(),
		).
		WithStatus().
		WithAssignee().
		All(ctx)
}

func (r *EntRepository) ListProjectCompletedIssuesSince(ctx context.Context, projectID uuid.UUID, since time.Time) ([]*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(
			issue.HasProjectWith(project.ID(projectID)),
			issue.HasStatusWith(workflowstatus.CategoryEQ(workflowstatus.CategoryDone)),
			issue.UpdatedAtGTE(since),
			issue.DeletedAtIsNil(),
		).
		All(ctx)
}

func (r *EntRepository) CountProjectCompletedIssuesInRange(ctx context.Context, projectID uuid.UUID, start, end time.Time) (int, error) {
	return r.client.Issue.Query().
		Where(
			issue.HasProjectWith(project.ID(projectID)),
			issue.HasStatusWith(workflowstatus.CategoryEQ(workflowstatus.CategoryDone)),
			issue.UpdatedAtGTE(start),
			issue.UpdatedAtLT(end),
			issue.DeletedAtIsNil(),
		).
		Count(ctx)
}

func (r *EntRepository) ListProjectRecentTestRuns(ctx context.Context, projectID uuid.UUID, limit int) ([]*ent.TestRun, error) {
	return r.client.TestRun.Query().
		Where(testrun.HasProjectWith(project.ID(projectID))).
		Order(ent.Desc(testrun.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) ListProjectActivityLogs(ctx context.Context, projectID uuid.UUID, limit int) ([]*ent.ActivityLog, error) {
	return r.client.ActivityLog.Query().
		Where(activitylog.HasIssueWith(issue.HasProjectWith(project.ID(projectID)))).
		WithUser().
		Order(ent.Desc(activitylog.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

func (r *EntRepository) CountUsers(ctx context.Context) (int, error) {
	return r.client.User.Query().Count(ctx)
}

func (r *EntRepository) CountDistinctUsersFromSessionsSince(ctx context.Context, since time.Time) (int, error) {
	ids, err := r.client.Session.Query().
		Where(session.CreatedAtGTE(since)).
		QueryUser().
		GroupBy(user.FieldID).
		Strings(ctx)
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

func (r *EntRepository) CountProjects(ctx context.Context) (int, error) {
	return r.client.Project.Query().Count(ctx)
}

func (r *EntRepository) CountIssues(ctx context.Context) (int, error) {
	return r.client.Issue.Query().
		Where(issue.DeletedAtIsNil()).
		Count(ctx)
}

func (r *EntRepository) ListProjectsWithIssues(ctx context.Context) ([]*ent.Project, error) {
	return r.client.Project.Query().
		WithIssues().
		All(ctx)
}

func (r *EntRepository) GetUser(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return r.client.User.Get(ctx, userID)
}

func (r *EntRepository) GetDashboardLayout(ctx context.Context, userID uuid.UUID, dashboardType string, projectID *uuid.UUID) (*ent.DashboardLayout, error) {
	q := r.client.DashboardLayout.Query().
		Where(
			dashboardlayout.HasUserWith(user.ID(userID)),
			dashboardlayout.DashboardTypeEQ(dashboardlayout.DashboardType(dashboardType)),
		)

	if projectID != nil {
		q = q.Where(dashboardlayout.ProjectID(*projectID))
	} else {
		q = q.Where(dashboardlayout.ProjectIDIsNil())
	}

	return q.Only(ctx)
}

func (r *EntRepository) SaveDashboardLayout(ctx context.Context, userID uuid.UUID, req SaveLayoutRequest) error {
	rawItems := make([]map[string]interface{}, 0, len(req.Items))
	for _, item := range req.Items {
		data, _ := json.Marshal(item)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		rawItems = append(rawItems, m)
	}

	q := r.client.DashboardLayout.Query().
		Where(
			dashboardlayout.HasUserWith(user.ID(userID)),
			dashboardlayout.DashboardTypeEQ(dashboardlayout.DashboardType(req.DashboardType)),
		)

	if req.ProjectID != nil {
		q = q.Where(dashboardlayout.ProjectID(*req.ProjectID))
	} else {
		q = q.Where(dashboardlayout.ProjectIDIsNil())
	}

	existing, err := q.Only(ctx)
	if err == nil {
		_, err = existing.Update().
			SetItems(rawItems).
			Save(ctx)
		return err
	}

	if !ent.IsNotFound(err) {
		return fmt.Errorf("ошибка поиска раскладки: %w", err)
	}

	creator := r.client.DashboardLayout.Create().
		SetUserID(userID).
		SetDashboardType(dashboardlayout.DashboardType(req.DashboardType)).
		SetItems(rawItems)

	if req.ProjectID != nil {
		creator = creator.SetProjectID(*req.ProjectID)
	}

	_, err = creator.Save(ctx)
	return err
}

// IsProjectMember реализует Repository.
func (r *EntRepository) IsProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	exists, err := r.client.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(project.IDEQ(projectID)),
			projectmember.HasUserWith(user.IDEQ(userID)),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("project membership: %w", err)
	}
	return exists, nil
}

// IsNotFound — сообщает об ошибках ent not-found.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
