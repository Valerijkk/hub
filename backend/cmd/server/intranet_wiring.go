package main

import (
	"context"

	"github.com/google/uuid"
	intranetmod "github.com/valerii/hub/backend/internal/intranet"
	notifmod "github.com/valerii/hub/backend/internal/notification"
)

// intranetNotifier — kudos и Random Coffee уведомления.
type intranetNotifier struct {
	svc *notifmod.NotificationService
}

func (a *intranetNotifier) NotifyKudosReceived(ctx context.Context, recipientID uuid.UUID, kudo intranetmod.KudoResponse, senderID uuid.UUID) error {
	if a.svc == nil {
		return nil
	}
	title := "Вам сказали спасибо!"
	body := kudo.SenderName + ": " + kudo.Message
	return a.svc.Dispatch(ctx, notifmod.DispatchEvent{
		UserID:     recipientID,
		Type:       notifmod.TypeMessageReceived,
		Title:      title,
		Body:       body,
		EntityType: "kudo",
		EntityID:   kudo.ID.String(),
		ActorID:    &senderID,
	})
}

func (a *intranetNotifier) NotifyCoffeeMatch(ctx context.Context, userID uuid.UUID, partner intranetmod.CoffeePartnerResponse) error {
	if a.svc == nil {
		return nil
	}
	title := "Random Coffee — ваш напарник"
	body := "Встретьтесь с " + partner.DisplayName + " на кофе на этой неделе"
	return a.svc.Dispatch(ctx, notifmod.DispatchEvent{
		UserID:     userID,
		Type:       notifmod.TypeMessageReceived,
		Title:      title,
		Body:       body,
		EntityType: "coffee_match",
		EntityID:   partner.UserID.String(),
	})
}
