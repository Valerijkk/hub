package messenger

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type MessengerDataEraseRepository interface {
	DeleteMessagesBySender(context.Context, uuid.UUID, uuid.UUID) (int64, error)
	DeleteReadStatesByUser(context.Context, uuid.UUID, uuid.UUID) (int64, error)
	DeleteThreadFollowsByUser(context.Context, uuid.UUID, uuid.UUID) (int64, error)
	PullResolvedMentionUserID(context.Context, uuid.UUID, uuid.UUID) (int64, error)
}

type MessengerDataEraseResult struct {
	DeletedMessages        int64
	DeletedReadStates      int64
	DeletedThreadFollows   int64
	PulledResolvedMentions int64
}

func ExportPhase151EraseMessengerUserData(
	ctx context.Context,
	tenantID uuid.UUID,
	userID uuid.UUID,
	repository MessengerDataEraseRepository,
) (MessengerDataEraseResult, error) {
	if repository == nil || tenantID == uuid.Nil || userID == uuid.Nil {
		return MessengerDataEraseResult{}, errors.New("messenger erase scope is required")
	}

	result := MessengerDataEraseResult{}
	var err error
	result.DeletedMessages, err = repository.DeleteMessagesBySender(ctx, tenantID, userID)
	if err != nil {
		return result, fmt.Errorf("delete messenger messages: %w", err)
	}
	result.DeletedReadStates, err = repository.DeleteReadStatesByUser(ctx, tenantID, userID)
	if err != nil {
		return result, fmt.Errorf("delete messenger read states: %w", err)
	}
	result.DeletedThreadFollows, err = repository.DeleteThreadFollowsByUser(ctx, tenantID, userID)
	if err != nil {
		return result, fmt.Errorf("delete messenger thread follows: %w", err)
	}
	result.PulledResolvedMentions, err = repository.PullResolvedMentionUserID(ctx, tenantID, userID)
	if err != nil {
		return result, fmt.Errorf("pull messenger resolved mentions: %w", err)
	}
	return result, nil
}
