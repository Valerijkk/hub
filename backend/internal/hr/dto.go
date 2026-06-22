package hr

import (
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// LeaveType — тип отсутствия.
type LeaveType string

const (
	LeaveVacation LeaveType = "vacation"
	LeaveSick     LeaveType = "sick"
	LeavePersonal LeaveType = "personal"
	LeaveOther    LeaveType = "other"
)

// LeaveStatus — статус заявки.
type LeaveStatus string

const (
	StatusPending   LeaveStatus = "pending"
	StatusApproved  LeaveStatus = "approved"
	StatusRejected  LeaveStatus = "rejected"
	StatusCancelled LeaveStatus = "cancelled"
)

// CreateLeaveRequest — подача заявки.
type CreateLeaveRequest struct {
	LeaveType LeaveType `json:"leave_type"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Reason    string    `json:"reason,omitempty"`
}

// ReviewLeaveRequest — решение approver (PROD-01 chain → admin queue).
type ReviewLeaveRequest struct {
	ReviewNote string `json:"review_note,omitempty"`
	Version    int    `json:"version"`
}

// LeaveRequestResponse — DTO заявки.
type LeaveRequestResponse struct {
	ID              uuid.UUID   `json:"id"`
	LeaveType       LeaveType   `json:"leave_type"`
	Status          LeaveStatus `json:"status"`
	StartDate       string      `json:"start_date"`
	EndDate         string      `json:"end_date"`
	Reason          string      `json:"reason,omitempty"`
	ReviewNote      string      `json:"review_note,omitempty"`
	RequesterID     uuid.UUID   `json:"requester_id"`
	RequesterName   string      `json:"requester_name,omitempty"`
	ReviewerID      *uuid.UUID  `json:"reviewer_id,omitempty"`
	ReviewerName    string      `json:"reviewer_name,omitempty"`
	ReviewedAt      *time.Time  `json:"reviewed_at,omitempty"`
	Version         int         `json:"version"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func formatDateUTC(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func toLeaveResponse(l *ent.LeaveRequest) LeaveRequestResponse {
	resp := LeaveRequestResponse{
		ID:          l.ID,
		LeaveType:   LeaveType(l.LeaveType),
		Status:      LeaveStatus(l.Status),
		StartDate:   formatDateUTC(l.StartDate),
		EndDate:     formatDateUTC(l.EndDate),
		Reason:      l.Reason,
		ReviewNote:  l.ReviewNote,
		Version:     l.Version,
		ReviewedAt:  l.ReviewedAt,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
	if l.Edges.Requester != nil {
		resp.RequesterID = l.Edges.Requester.ID
		resp.RequesterName = displayName(l.Edges.Requester)
	}
	if l.Edges.Reviewer != nil {
		resp.ReviewerID = &l.Edges.Reviewer.ID
		resp.ReviewerName = displayName(l.Edges.Reviewer)
	}
	return resp
}

func displayName(u *ent.User) string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}
