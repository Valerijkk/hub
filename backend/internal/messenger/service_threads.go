package messenger

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

type phase151ThreadRuntime interface {
	FindThreadFollow(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string) (string, bool, error)
	AdvanceThreadReadCursor(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, string) (bool, error)
}

type ThreadResponse struct {
	Parent     MessageResponse   `json:"parent"`
	Replies    []MessageResponse `json:"replies"`
	FollowMode string            `json:"follow_mode,omitempty"`
}

type MarkThreadReadRequest struct {
	Category  string `json:"category"`
	MessageID string `json:"message_id"`
}

func (s *MessengerService) GetThread(
	ctx context.Context,
	channelID uuid.UUID,
	rootMessageID string,
	userID uuid.UUID,
	before *time.Time,
	limit int,
) (*ThreadResponse, error) {
	access, scoped, accessErr := s.resolveChannelAction(ctx, userID, channelID, ChannelActionReadThread)
	if accessErr != nil {
		return nil, accessErr
	}
	if !scoped {
		isMember, membershipErr := s.repo.IsMember(ctx, channelID, userID)
		if membershipErr != nil {
			return nil, apperrors.NewInternal("ошибка проверки участия")
		}
		if !isMember {
			return nil, apperrors.NewForbidden("вы не являетесь участником канала")
		}
	}

	root, err := s.repo.FindMessageByID(ctx, rootMessageID)
	if err != nil || root == nil || root.DeletedAt != nil || root.ChannelID != channelID || root.ThreadID != "" {
		return nil, apperrors.NewNotFound("тред не найден")
	}
	replyDocs, err := s.repo.ListThreadReplies(ctx, rootMessageID, before, limit)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения треда")
	}
	replies := make([]MessageResponse, 0, len(replyDocs))
	for _, reply := range replyDocs {
		if reply == nil || reply.DeletedAt != nil || reply.ChannelID != channelID || reply.ThreadID != rootMessageID {
			continue
		}
		replies = append(replies, *toMessageResponse(reply))
	}
	response := &ThreadResponse{Parent: *toMessageResponse(root), Replies: replies}
	if scoped && access.IsMember && access.CanFollow {
		if runtime, ok := s.repo.(phase151ThreadRuntime); ok {
			mode, found, followErr := runtime.FindThreadFollow(ctx, access.TenantID, userID, channelID, rootMessageID)
			if followErr != nil {
				return nil, apperrors.NewInternal("ошибка получения режима треда")
			}
			if found {
				response.FollowMode = mode
			}
		}
	}
	return response, nil
}

func (s *MessengerService) MarkThreadRead(
	ctx context.Context,
	channelID uuid.UUID,
	rootMessageID string,
	userID uuid.UUID,
	request MarkThreadReadRequest,
) (bool, error) {
	access, scoped, accessErr := s.resolveChannelAction(ctx, userID, channelID, ChannelActionReadThread)
	if accessErr != nil {
		return false, accessErr
	}
	if !scoped || !access.IsMember || !access.CanViewUnread {
		return false, apperrors.NewForbidden("нельзя изменять состояние чтения треда")
	}
	if request.Category != string(UnreadCategoryFollowedThread) || request.MessageID == "" {
		return false, apperrors.NewBadRequest("некорректная квитанция чтения треда")
	}

	root, err := s.repo.FindMessageByID(ctx, rootMessageID)
	if err != nil || root == nil || root.DeletedAt != nil || root.ChannelID != channelID || root.ThreadID != "" {
		return false, apperrors.NewNotFound("тред не найден")
	}
	candidate, err := s.repo.FindMessageByID(ctx, request.MessageID)
	if err != nil || candidate == nil || candidate.DeletedAt != nil || candidate.ChannelID != channelID {
		return false, apperrors.NewNotFound("сообщение треда не найдено")
	}
	if candidate.ID != rootMessageID && candidate.ThreadID != rootMessageID {
		return false, apperrors.NewNotFound("сообщение треда не найдено")
	}
	runtime, ok := s.repo.(phase151ThreadRuntime)
	if !ok {
		return false, apperrors.NewInternal("durable thread runtime не настроен")
	}
	advanced, err := runtime.AdvanceThreadReadCursor(
		ctx,
		access.TenantID,
		userID,
		channelID,
		rootMessageID,
		request.MessageID,
	)
	if err != nil {
		return false, apperrors.NewInternal("ошибка сохранения квитанции чтения треда")
	}
	return advanced, nil
}

type phase151ThreadFollowMutationRuntime interface {
	CurrentAuthorizedFollowedRootHead(context.Context, AuthorizedUnreadScope, AuthorizedUnreadTarget, string) (MessagePosition, error)
	ApplyAuthorizedThreadFollowTransition(context.Context, AuthorizedUnreadScope, uuid.UUID, string, ThreadFollowTransition) (bool, error)
}

func (s *MessengerService) FollowThread(
	ctx context.Context,
	channelID uuid.UUID,
	rootMessageID string,
	userID uuid.UUID,
) (bool, error) {
	return s.setThreadFollowMode(ctx, channelID, rootMessageID, userID, ThreadFollowModeManualFollow)
}

func (s *MessengerService) UnfollowThread(
	ctx context.Context,
	channelID uuid.UUID,
	rootMessageID string,
	userID uuid.UUID,
) (bool, error) {
	return s.setThreadFollowMode(ctx, channelID, rootMessageID, userID, ThreadFollowModeManualUnfollow)
}

func (s *MessengerService) setThreadFollowMode(
	ctx context.Context,
	channelID uuid.UUID,
	rootMessageID string,
	userID uuid.UUID,
	mode ThreadFollowMode,
) (bool, error) {
	access, scoped, err := s.resolveChannelAction(ctx, userID, channelID, ChannelActionReadThread)
	if err != nil {
		return false, err
	}
	if !scoped || !access.IsMember ||
		(mode == ThreadFollowModeManualFollow && !access.CanFollow) ||
		(mode == ThreadFollowModeManualUnfollow && !access.CanUnfollow) {
		return false, apperrors.NewForbidden("thread follow state is unavailable")
	}
	if rootMessageID == "" || s.unreadScope == nil {
		return false, apperrors.NewForbidden("authorized unread scope is required")
	}
	scope, err := s.unreadScope.ResolveAuthorizedUnreadScope(ctx, userID)
	if err != nil || scope.TenantID != access.TenantID || scope.UserID != userID {
		return false, apperrors.NewForbidden("authorized unread scope is unavailable")
	}
	var target *AuthorizedUnreadTarget
	for index := range scope.Targets {
		candidate := &scope.Targets[index]
		if candidate.ChannelID == channelID && candidate.Kind == "channel" {
			target = candidate
			break
		}
	}
	if target == nil {
		return false, apperrors.NewForbidden("thread follow target is unavailable")
	}
	runtime, ok := s.repo.(phase151ThreadFollowMutationRuntime)
	if !ok {
		return false, apperrors.NewInternal("durable thread follow runtime is not configured")
	}
	position, err := runtime.CurrentAuthorizedFollowedRootHead(ctx, scope, *target, rootMessageID)
	if err != nil {
		return false, apperrors.NewInternal("failed to resolve thread head")
	}
	if position.CreatedAt.IsZero() || position.MessageID == "" {
		return false, apperrors.NewNotFound("thread not found")
	}
	changed, err := runtime.ApplyAuthorizedThreadFollowTransition(ctx, scope, channelID, rootMessageID, ThreadFollowTransition{
		Mode: mode, Position: position,
	})
	if err != nil {
		return false, apperrors.NewInternal("failed to persist thread follow state")
	}
	return changed, nil
}