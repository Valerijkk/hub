package hr

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/leaverequest"
)

var (
	ErrNotFound        = errors.New("leave request not found")
	ErrForbidden       = errors.New("admin access required")
	ErrInvalidInput    = errors.New("invalid leave request input")
	ErrVersionConflict = errors.New("leave request version conflict")
	ErrOverlap         = errors.New("leave dates overlap existing request")
)

// Notifier — уведомления approver/requester (PROD-01-style notify chain).
type Notifier interface {
	NotifyLeaveSubmitted(ctx context.Context, adminIDs []uuid.UUID, req LeaveRequestResponse, requesterID uuid.UUID) error
	NotifyLeaveDecision(ctx context.Context, requesterID uuid.UUID, approved bool, req LeaveRequestResponse, reviewerID uuid.UUID) error
}

// LeaveExporter — export в external HR/1C (MOD-07).
type LeaveExporter interface {
	ExportApprovedLeave(ctx context.Context, req LeaveRequestResponse, reviewerID uuid.UUID) error
}

// Service — бизнес-логика HR-lite.
type Service struct {
	repo     Repository
	notifier Notifier
	exporter LeaveExporter
}

// NewService создаёт сервис HR.
func NewService(repo Repository, notifier Notifier, exporter LeaveExporter) *Service {
	return &Service{repo: repo, notifier: notifier, exporter: exporter}
}

func (s *Service) ListMyLeave(ctx context.Context, userID uuid.UUID) ([]LeaveRequestResponse, error) {
	items, err := s.repo.ListLeaveByRequester(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list my leave: %w", err)
	}
	out := make([]LeaveRequestResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toLeaveResponse(item))
	}
	return out, nil
}

func (s *Service) ListPending(ctx context.Context, userID uuid.UUID) ([]LeaveRequestResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListPendingLeave(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending leave: %w", err)
	}
	out := make([]LeaveRequestResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toLeaveResponse(item))
	}
	return out, nil
}

func (s *Service) CreateLeave(ctx context.Context, userID uuid.UUID, req CreateLeaveRequest) (*LeaveRequestResponse, error) {
	start, end, err := parseDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	if !isValidLeaveType(req.LeaveType) {
		return nil, ErrInvalidInput
	}
	overlap, err := s.repo.HasOverlappingLeave(ctx, userID, start, end, nil)
	if err != nil {
		return nil, fmt.Errorf("check overlap: %w", err)
	}
	if overlap {
		return nil, ErrOverlap
	}
	l, err := s.repo.CreateLeaveRequest(ctx, userID, leaverequest.LeaveType(req.LeaveType), start, end, strings.TrimSpace(req.Reason))
	if err != nil {
		return nil, fmt.Errorf("create leave: %w", err)
	}
	l, err = s.repo.GetLeaveRequest(ctx, l.ID)
	if err != nil {
		return nil, fmt.Errorf("reload leave: %w", err)
	}
	resp := toLeaveResponse(l)
	if s.notifier != nil {
		adminIDs, err := s.repo.ListSystemAdminIDs(ctx)
		if err == nil && len(adminIDs) > 0 {
			_ = s.notifier.NotifyLeaveSubmitted(ctx, adminIDs, resp, userID)
		}
	}
	return &resp, nil
}

func (s *Service) ApproveLeave(ctx context.Context, reviewerID, leaveID uuid.UUID, req ReviewLeaveRequest) (*LeaveRequestResponse, error) {
	if err := s.requireAdmin(ctx, reviewerID); err != nil {
		return nil, err
	}
	version := req.Version
	if version == 0 {
		current, err := s.repo.GetLeaveRequest(ctx, leaveID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("get leave: %w", err)
		}
		version = current.Version
	}
	l, err := s.repo.ApproveLeave(ctx, leaveID, reviewerID, strings.TrimSpace(req.ReviewNote), version)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("approve leave: %w", err)
	}
	l, err = s.repo.GetLeaveRequest(ctx, l.ID)
	if err != nil {
		return nil, fmt.Errorf("reload leave: %w", err)
	}
	resp := toLeaveResponse(l)
	if s.notifier != nil {
		_ = s.notifier.NotifyLeaveDecision(ctx, resp.RequesterID, true, resp, reviewerID)
	}
	if s.exporter != nil {
		if err := s.exporter.ExportApprovedLeave(ctx, resp, reviewerID); err != nil {
			// Export failure не откатывает approval — лог exporter.
			_ = err
		}
	}
	return &resp, nil
}

func (s *Service) RejectLeave(ctx context.Context, reviewerID, leaveID uuid.UUID, req ReviewLeaveRequest) (*LeaveRequestResponse, error) {
	if err := s.requireAdmin(ctx, reviewerID); err != nil {
		return nil, err
	}
	version := req.Version
	if version == 0 {
		current, err := s.repo.GetLeaveRequest(ctx, leaveID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("get leave: %w", err)
		}
		version = current.Version
	}
	l, err := s.repo.RejectLeave(ctx, leaveID, reviewerID, strings.TrimSpace(req.ReviewNote), version)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("reject leave: %w", err)
	}
	l, err = s.repo.GetLeaveRequest(ctx, l.ID)
	if err != nil {
		return nil, fmt.Errorf("reload leave: %w", err)
	}
	resp := toLeaveResponse(l)
	if s.notifier != nil {
		_ = s.notifier.NotifyLeaveDecision(ctx, resp.RequesterID, false, resp, reviewerID)
	}
	return &resp, nil
}

func (s *Service) CancelLeave(ctx context.Context, userID, leaveID uuid.UUID, version int) (*LeaveRequestResponse, error) {
	if version == 0 {
		current, err := s.repo.GetLeaveRequest(ctx, leaveID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, ErrNotFound
			}
			return nil, fmt.Errorf("get leave: %w", err)
		}
		if current.Edges.Requester != nil && current.Edges.Requester.ID != userID {
			return nil, ErrForbidden
		}
		version = current.Version
	}
	l, err := s.repo.CancelLeave(ctx, leaveID, userID, version)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("cancel leave: %w", err)
	}
	l, err = s.repo.GetLeaveRequest(ctx, l.ID)
	if err != nil {
		return nil, fmt.Errorf("reload leave: %w", err)
	}
	resp := toLeaveResponse(l)
	return &resp, nil
}

func (s *Service) requireAdmin(ctx context.Context, userID uuid.UUID) error {
	ok, err := s.repo.IsSystemAdmin(ctx, userID)
	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func parseDateRange(startRaw, endRaw string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	end, err := time.Parse("2006-01-02", endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	start = start.UTC()
	end = end.UTC()
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	return start, end, nil
}

func isValidLeaveType(t LeaveType) bool {
	switch t {
	case LeaveVacation, LeaveSick, LeavePersonal, LeaveOther:
		return true
	default:
		return t == ""
	}
}
