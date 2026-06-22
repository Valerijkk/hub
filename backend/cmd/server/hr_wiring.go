package main

import (
	"context"

	"github.com/google/uuid"
	hrmod "github.com/valerii/hub/backend/internal/hr"
	notifmod "github.com/valerii/hub/backend/internal/notification"
)

// hrNotifierAdapter — уведомления админов и заявителей в workflow отпусков (MOD-06).
type hrNotifierAdapter struct {
	svc *notifmod.NotificationService
}

func (a *hrNotifierAdapter) NotifyLeaveSubmitted(ctx context.Context, adminIDs []uuid.UUID, req hrmod.LeaveRequestResponse, requesterID uuid.UUID) error {
	if a.svc == nil {
		return nil
	}
	title := "Заявка на отпуск"
	body := req.RequesterName + ": " + req.StartDate + " — " + req.EndDate
	reqID := req.ID.String()
	actor := requesterID
	for _, adminID := range adminIDs {
		if adminID == requesterID {
			continue
		}
		if err := a.svc.Dispatch(ctx, notifmod.DispatchEvent{
			UserID:     adminID,
			Type:       notifmod.TypeAssigned,
			Title:      title,
			Body:       body,
			EntityType: "leave_request",
			EntityID:   reqID,
			ActorID:    &actor,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (a *hrNotifierAdapter) NotifyLeaveDecision(ctx context.Context, requesterID uuid.UUID, approved bool, req hrmod.LeaveRequestResponse, reviewerID uuid.UUID) error {
	if a.svc == nil {
		return nil
	}
	title := "Заявка отклонена"
	if approved {
		title = "Отпуск одобрен"
	}
	body := req.StartDate + " — " + req.EndDate
	if req.ReviewNote != "" {
		body += ": " + req.ReviewNote
	}
	return a.svc.Dispatch(ctx, notifmod.DispatchEvent{
		UserID:     requesterID,
		Type:       notifmod.TypeAssigned,
		Title:      title,
		Body:       body,
		EntityType: "leave_request",
		EntityID:   req.ID.String(),
		ActorID:    &reviewerID,
	})
}
