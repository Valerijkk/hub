package calendar

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository — доступ к данным календаря.
type Repository interface {
	ListUserProjectIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	ListScheduledCalls(ctx context.Context, projectIDs []uuid.UUID) ([]*ent.ScheduledCall, error)
	ListSprintsInRange(ctx context.Context, projectIDs []uuid.UUID, from, to time.Time) ([]*ent.Sprint, error)
	ListIssuesDueInRange(ctx context.Context, projectIDs []uuid.UUID, from, to time.Time) ([]*ent.Issue, error)
	// ListApprovedLeaveInRange — только leave зрителя и участников общих проектов (A6-005).
	ListApprovedLeaveInRange(ctx context.Context, viewerID uuid.UUID, projectIDs []uuid.UUID, from, to time.Time) ([]*ent.LeaveRequest, error)
}

// EntRepository реализует Repository через Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий календаря.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) ListUserProjectIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	projects, err := r.client.Project.Query().
		Where(
			project.HasMembersWith(projectmember.HasUserWith(user.IDEQ(userID))),
			project.StatusNEQ(project.StatusDeleted),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	return ids, nil
}

func (r *EntRepository) ListScheduledCalls(ctx context.Context, projectIDs []uuid.UUID) ([]*ent.ScheduledCall, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	return r.client.ScheduledCall.Query().
		Where(
			scheduledcall.ProjectIDIn(projectIDs...),
			scheduledcall.IsActive(true),
			scheduledcall.DeletedAtIsNil(),
		).
		All(ctx)
}

func (r *EntRepository) ListSprintsInRange(ctx context.Context, projectIDs []uuid.UUID, from, to time.Time) ([]*ent.Sprint, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	return r.client.Sprint.Query().
		Where(
			sprintEnt.HasProjectWith(project.IDIn(projectIDs...)),
			sprintEnt.StartDateLTE(to),
			sprintEnt.EndDateGTE(from),
		).
		WithProject().
		All(ctx)
}

func (r *EntRepository) ListIssuesDueInRange(ctx context.Context, projectIDs []uuid.UUID, from, to time.Time) ([]*ent.Issue, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	return r.client.Issue.Query().
		Where(
			issue.HasProjectWith(project.IDIn(projectIDs...)),
			issue.DueDateNotNil(),
			issue.DueDateGTE(from),
			issue.DueDateLTE(to),
			issue.DeletedAtIsNil(),
		).
		WithProject().
		All(ctx)
}

func (r *EntRepository) ListApprovedLeaveInRange(ctx context.Context, viewerID uuid.UUID, projectIDs []uuid.UUID, from, to time.Time) ([]*ent.LeaveRequest, error) {
	// A6-005: не отдавать org-wide approved leave любому JWT.
	// Видно: своё + leave участников проектов, в которых состоит viewer.
	requesterPred := user.IDEQ(viewerID)
	if len(projectIDs) > 0 {
		requesterPred = user.Or(
			user.IDEQ(viewerID),
			user.HasProjectMembersWith(projectmember.HasProjectWith(project.IDIn(projectIDs...))),
		)
	}
	return r.client.LeaveRequest.Query().
		Where(
			leaverequest.StatusEQ(leaverequest.StatusApproved),
			leaverequest.StartDateLTE(to),
			leaverequest.EndDateGTE(from),
			leaverequest.HasRequesterWith(requesterPred),
		).
		WithRequester().
		All(ctx)
}
