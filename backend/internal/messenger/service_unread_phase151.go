package messenger

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

type phase151UnreadReceiptRuntime interface {
	CurrentAuthorizedTargetHead(context.Context, AuthorizedUnreadScope, AuthorizedUnreadTarget, UnreadCategory) (MessagePosition, error)
	AdvanceAuthorizedUnreadCursor(context.Context, AuthorizedUnreadScope, UnreadCategory, string, MessagePosition) (bool, error)
	PersistUnreadCutoverBaseline(context.Context, AuthorizedUnreadScope, UnreadCategory, string, MessagePosition) error
}
type phase151EligibleThreadFollowLookup interface {
	ListEligibleThreadFollowRoots(context.Context, uuid.UUID, uuid.UUID) ([]string, error)
}

func (s *MessengerService) GetUnreadResponse(ctx context.Context, userID string) (UnreadResponse, error) {
	response := newPhase151UnreadResponse()
	parsedUserID, err := uuid.Parse(userID)
	if err != nil || parsedUserID == uuid.Nil {
		return response, apperrors.NewBadRequest("некорректный ID пользователя")
	}

	if s.unreadScope != nil {
		scope, scopeErr := s.unreadScope.ResolveAuthorizedUnreadScope(ctx, parsedUserID)
		if scopeErr != nil || scope.TenantID == uuid.Nil || scope.UserID != parsedUserID {
			return response, apperrors.NewForbidden("область непрочитанных недоступна")
		}
		if len(scope.Targets) == 0 {
			return response, nil
		}
		if coordinator, ok := s.repo.(Phase151UnreadCoordinatorRepository); ok {
			return ExportPhase151BuildUnreadResponse(ctx, scope, coordinator)
		}
		if follows, ok := s.repo.(phase151EligibleThreadFollowLookup); ok {
			rootIDs, listErr := follows.ListEligibleThreadFollowRoots(ctx, scope.TenantID, parsedUserID)
			if listErr != nil {
				return response, apperrors.NewInternal("ошибка получения подписок на треды")
			}
			authorizedChannels := make(map[uuid.UUID]struct{}, len(scope.Targets))
			for _, target := range scope.Targets {
				if target.ChannelID != uuid.Nil {
					authorizedChannels[target.ChannelID] = struct{}{}
				}
			}
			for _, rootID := range rootIDs {
				if rootID == "" {
					continue
				}
				root, findErr := s.repo.FindMessageByID(ctx, rootID)
				if findErr != nil || root == nil || root.DeletedAt != nil || root.ThreadID != "" {
					continue
				}
				if _, allowed := authorizedChannels[root.ChannelID]; !allowed {
					continue
				}
				// A live root alone is not unread evidence. Counts are added only by
				// Phase151UnreadCoordinatorRepository candidates after the cursor.
			}
		}
		return response, nil
	}

	if s.durableUnread != nil {
		counts, countsErr := s.durableUnread.GetDurableUnreadCounts(ctx, parsedUserID)
		if countsErr != nil {
			return response, countsErr
		}
		for targetID, count := range counts {
			response.Channels[targetID] = count
			if count > 0 {
				response.AnyUnread = true
			}
		}
	}
	return response, nil
}
