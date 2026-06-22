package main

import (
	"context"

	"github.com/google/uuid"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
)

// threadMirrorAdapter дублирует комментарии issue в связанные треды messenger (PROD-03).
type threadMirrorAdapter struct {
	svc *messengermod.MessengerService
}

func (a *threadMirrorAdapter) MirrorIssueComment(
	ctx context.Context,
	channelID, senderID uuid.UUID,
	threadRootMessageID, text string,
) error {
	if a.svc == nil {
		return nil
	}
	_, err := a.svc.SendMessage(ctx, channelID, senderID, messengermod.SendMessageRequest{
		ContentText: text,
		ThreadID:    threadRootMessageID,
		ContentJSON: map[string]interface{}{"type": "doc", "children": []interface{}{
			map[string]interface{}{"type": "p", "children": []interface{}{
				map[string]interface{}{"text": text},
			}},
		}},
	})
	return err
}
