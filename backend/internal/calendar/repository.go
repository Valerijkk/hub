package calendar

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/user"
)

// MaxCalendarList — W7-A2-010: hard cap на списки календаря (events/calls/issues/sprints/leave).
// Mutation: снять Limit → unbounded All при N>cap.
const MaxCalendarList = 100

// Repository — доступ к данным календаря.
type Repository interface {
	ListUserProjectIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	ListCalendarEventsForActor(ctx context.Context, actorID uuid.UUID, from, to time.Time) ([]*ent.CalendarEvent, error)
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

// ListCalendarEventsForActor возвращает видимые активному актору канонические
// события и загружает их исключения экземпляров в одном снимке.
func (r *EntRepository) ListCalendarEventsForActor(ctx context.Context, actorID uuid.UUID, from, to time.Time) ([]*ent.CalendarEvent, error) {
	if r == nil || r.client == nil || ctx == nil || actorID == uuid.Nil {
		return nil, ErrEventInvalid
	}
	from = from.UTC()
	to = to.UTC()
	if err := validateRange(from, to); err != nil {
		return nil, err
	}
	lowerDate, upperDate := availabilitySourceDateBounds(from, to)
	transaction, err := r.client.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, ErrEventInternal
	}
	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback()
			panic(recovered)
		}
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	organizations, err := transaction.Organization.Query().
		Where(organization.HasUsersWith(user.IDEQ(actorID), user.StatusEQ(user.StatusActive))).
		Limit(2).
		All(ctx)
	if err != nil {
		return nil, ErrEventInternal
	}
	if len(organizations) == 0 {
		return []*ent.CalendarEvent{}, nil
	}
	if len(organizations) != 1 || organizations[0] == nil || organizations[0].ID == uuid.Nil {
		return nil, ErrEventInternal
	}

	events, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.OrganizationIDEQ(organizations[0].ID),
			calendarevent.DeletedAtIsNil(),
			calendarevent.StatusIn(calendarevent.StatusConfirmed, calendarevent.StatusTentative),
			calendarevent.Or(
				calendarevent.CalendarOwnerIDEQ(actorID),
				calendarevent.OrganizerIDEQ(actorID),
				calendarevent.HasAttendeesWith(
					calendarattendee.UserIDEQ(actorID),
					calendarattendee.RemovedAtIsNil(),
					calendarattendee.RecurrenceIDIsNil(),
				),
			),
			calendarevent.Or(
				calendarevent.And(
					calendarevent.AllDay(false),
					calendarevent.StartsAtLT(to),
					calendarevent.Or(
						calendarevent.RruleNotNil(),
						calendarevent.EndsAtGT(from),
					),
				),
				calendarevent.And(
					calendarevent.AllDay(true),
					calendarevent.StartDateLT(upperDate),
					calendarevent.Or(
						calendarevent.RruleNotNil(),
						calendarevent.EndDateExclusiveGT(lowerDate),
					),
				),
			),
		).
		WithExceptions().
		Order(ent.Asc(calendarevent.FieldID)).
		Limit(MaxCalendarList + 1).
		All(ctx)
	if err != nil {
		return nil, ErrEventInternal
	}
	if len(events) > MaxCalendarList {
		return nil, ErrAvailabilityLimitExceeded
	}
	if err := transaction.Commit(); err != nil {
		return nil, ErrEventInternal
	}
	committed = true
	return events, nil
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
			scheduledcall.Not(scheduledcall.HasCalendarEventLink()),
		).
		WithProject().
		Limit(MaxCalendarList).
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
			// BE-V4-002 / TASK-189: корзина не в /events и .ics.
			sprintEnt.DeletedAtIsNil(),
		).
		WithProject().
		Limit(MaxCalendarList).
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
		Limit(MaxCalendarList).
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
			// BE-V4-001 / TASK-189: как hr/repository — GDPR soft-delete не в календаре.
			leaverequest.DeletedAtIsNil(),
		).
		WithRequester().
		Limit(MaxCalendarList).
		All(ctx)
}
