package main

import (
	"context"

	"github.com/google/uuid"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
	notifmod "github.com/valerii/hub/backend/internal/notification"
)

// messengerAutomationAdapter — публикация сообщений в канал для правил автоматизации.
type messengerAutomationAdapter struct {
	svc *messengermod.MessengerService
}

func (a *messengerAutomationAdapter) PostAutomationMessage(ctx context.Context, channelID, senderID uuid.UUID, text string) error {
	if a.svc == nil {
		return nil
	}
	_, err := a.svc.SendMessage(ctx, channelID, senderID, messengermod.SendMessageRequest{
		ContentText: text,
		ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{
			map[string]interface{}{"type": "p", "children": []interface{}{
				map[string]interface{}{"text": text},
			}},
		}},
	})
	return err
}

// notificationAutomationAdapter — in-app уведомления для правил автоматизации.
type notificationAutomationAdapter struct {
	svc *notifmod.NotificationService
}

func (a *notificationAutomationAdapter) NotifyAutomation(ctx context.Context, userID uuid.UUID, title, body, entityType, entityID string, projectID, actorID *uuid.UUID) error {
	if a.svc == nil {
		return nil
	}
	return a.svc.Dispatch(ctx, notifmod.DispatchEvent{
		UserID:     userID,
		Type:       notifmod.TypeAssigned,
		Title:      title,
		Body:       body,
		EntityType: entityType,
		EntityID:   entityID,
		ProjectID:  projectID,
		ActorID:    actorID,
	})
}
