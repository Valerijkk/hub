package hr

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/leaverequest"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository — доступ к данным HR.
type Repository interface {
	CreateLeaveRequest(ctx context.Context, requesterID uuid.UUID, leaveType leaverequest.LeaveType, start, end time.Time, reason string) (*ent.LeaveRequest, error)
	ListLeaveByRequester(ctx context.Context, requesterID uuid.UUID) ([]*ent.LeaveRequest, error)
	ListPendingLeave(ctx context.Context) ([]*ent.LeaveRequest, error)
	ListApprovedLeaveInRange(ctx context.Context, from, to time.Time) ([]*ent.LeaveRequest, error)
	GetLeaveRequest(ctx context.Context, id uuid.UUID) (*ent.LeaveRequest, error)
	HasOverlappingLeave(ctx context.Context, requesterID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) (bool, error)
	ApproveLeave(ctx context.Context, id, reviewerID uuid.UUID, note string, version int) (*ent.LeaveRequest, error)
	RejectLeave(ctx context.Context, id, reviewerID uuid.UUID, note string, version int) (*ent.LeaveRequest, error)
	CancelLeave(ctx context.Context, id, requesterID uuid.UUID, version int) (*ent.LeaveRequest, error)
	// SoftDeleteLeave — GDPR/trash soft-delete (A2W8-3-001 / ADR-ENT-SOFT-DELETE-GAPS-2026).
	SoftDeleteLeave(ctx context.Context, id uuid.UUID) error
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	ListSystemAdminIDs(ctx context.Context) ([]uuid.UUID, error)
}

// EntRepository реализует Repository через Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий HR.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreateLeaveRequest(ctx context.Context, requesterID uuid.UUID, leaveType leaverequest.LeaveType, start, end time.Time, reason string) (*ent.LeaveRequest, error) {
	create := r.client.LeaveRequest.Create().
		SetRequesterID(requesterID).
		SetLeaveType(leaveType).
		SetStartDate(start).
		SetEndDate(end)
	if reason != "" {
		create.SetReason(reason)
	}
	return create.Save(ctx)
}

func (r *EntRepository) ListLeaveByRequester(ctx context.Context, requesterID uuid.UUID) ([]*ent.LeaveRequest, error) {
	return r.client.LeaveRequest.Query().
		Where(
			leaverequest.HasRequesterWith(user.IDEQ(requesterID)),
			leaverequest.DeletedAtIsNil(),
		).
		WithRequester().
		WithReviewer().
		Order(ent.Desc(leaverequest.FieldCreatedAt)).
		All(ctx)
}

func (r *EntRepository) ListPendingLeave(ctx context.Context) ([]*ent.LeaveRequest, error) {
	return r.client.LeaveRequest.Query().
		Where(
			leaverequest.StatusEQ(leaverequest.StatusPending),
			leaverequest.DeletedAtIsNil(),
		).
		WithRequester().
		WithReviewer().
		Order(ent.Asc(leaverequest.FieldStartDate)).
		All(ctx)
}

func (r *EntRepository) ListApprovedLeaveInRange(ctx context.Context, from, to time.Time) ([]*ent.LeaveRequest, error) {
	return r.client.LeaveRequest.Query().
		Where(
			leaverequest.StatusEQ(leaverequest.StatusApproved),
			leaverequest.DeletedAtIsNil(),
			leaverequest.StartDateLTE(to),
			leaverequest.EndDateGTE(from),
		).
		WithRequester().
		All(ctx)
}

func (r *EntRepository) GetLeaveRequest(ctx context.Context, id uuid.UUID) (*ent.LeaveRequest, error) {
	return r.client.LeaveRequest.Query().
		Where(leaverequest.ID(id), leaverequest.DeletedAtIsNil()).
		WithRequester().
		WithReviewer().
		Only(ctx)
}

func (r *EntRepository) HasOverlappingLeave(ctx context.Context, requesterID uuid.UUID, start, end time.Time, excludeID *uuid.UUID) (bool, error) {
	q := r.client.LeaveRequest.Query().
		Where(
			leaverequest.HasRequesterWith(user.IDEQ(requesterID)),
			leaverequest.DeletedAtIsNil(),
			leaverequest.StatusIn(leaverequest.StatusPending, leaverequest.StatusApproved),
			leaverequest.StartDateLTE(end),
			leaverequest.EndDateGTE(start),
		)
	if excludeID != nil {
		q = q.Where(leaverequest.IDNEQ(*excludeID))
	}
	return q.Exist(ctx)
}

func (r *EntRepository) ApproveLeave(ctx context.Context, id, reviewerID uuid.UUID, note string, version int) (*ent.LeaveRequest, error) {
	now := time.Now().UTC()
	upd := r.client.LeaveRequest.UpdateOneID(id).
		Where(
			leaverequest.StatusEQ(leaverequest.StatusPending),
			leaverequest.DeletedAtIsNil(),
			leaverequest.Version(version),
		).
		SetStatus(leaverequest.StatusApproved).
		SetReviewerID(reviewerID).
		SetReviewedAt(now).
		AddVersion(1)
	if note != "" {
		upd.SetReviewNote(note)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) RejectLeave(ctx context.Context, id, reviewerID uuid.UUID, note string, version int) (*ent.LeaveRequest, error) {
	now := time.Now().UTC()
	upd := r.client.LeaveRequest.UpdateOneID(id).
		Where(
			leaverequest.StatusEQ(leaverequest.StatusPending),
			leaverequest.DeletedAtIsNil(),
			leaverequest.Version(version),
		).
		SetStatus(leaverequest.StatusRejected).
		SetReviewerID(reviewerID).
		SetReviewedAt(now).
		AddVersion(1)
	if note != "" {
		upd.SetReviewNote(note)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) CancelLeave(ctx context.Context, id, requesterID uuid.UUID, version int) (*ent.LeaveRequest, error) {
	return r.client.LeaveRequest.UpdateOneID(id).
		Where(
			leaverequest.HasRequesterWith(user.IDEQ(requesterID)),
			leaverequest.DeletedAtIsNil(),
			leaverequest.StatusEQ(leaverequest.StatusPending),
			leaverequest.Version(version),
		).
		SetStatus(leaverequest.StatusCancelled).
		AddVersion(1).
		Save(ctx)
}

// SoftDeleteLeave — A2W8-3-001 / ADR-ENT-SOFT-DELETE-GAPS-2026: GDPR soft-path для HR-PII.
func (r *EntRepository) SoftDeleteLeave(ctx context.Context, id uuid.UUID) error {
	affected, err := r.client.LeaveRequest.Update().
		Where(leaverequest.IDEQ(id), leaverequest.DeletedAtIsNil()).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

func (r *EntRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.SystemRole == user.SystemRoleAdmin || u.SystemRole == user.SystemRoleFounder, nil
}

func (r *EntRepository) ListSystemAdminIDs(ctx context.Context) ([]uuid.UUID, error) {
	users, err := r.client.User.Query().
		Where(
			user.StatusEQ(user.StatusActive),
			user.Or(
				user.SystemRoleEQ(user.SystemRoleAdmin),
				user.SystemRoleEQ(user.SystemRoleFounder),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids, nil
}
