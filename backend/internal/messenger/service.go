// Package messenger реализует мессенджер: каналы, сообщения, треды, реакции, присутствие, поиск.
package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// MessengerService реализует бизнес-логику модуля мессенджера.
// Оркестрирует операции с каналами, сообщениями, реакциями,
// публикует NATS-события для real-time доставки.
//
// SC-14-3: сервис не хранит *ent.Client напрямую — все БД-запросы идут через
// Repository интерфейс. Это устраняет утечку Ent ORM в бизнес-логику и
// позволяет подменять backend в тестах через mock-репозиторий.
type MessengerService struct {
	repo              Repository
	nats              EventPublisher
	presence          *PresenceService
	durableUnread     Phase151DurableUnreadRuntime
	channelAccess     ChannelAccessResolver
	unreadScope       AuthorizedUnreadScopeResolver
	messageSearchScope AuthorizedMessageSearchScopeResolver
	promotionConfigMu sync.RWMutex
	issuePromoter     IssuePromoter            // опционально: для /promote-to-issue (см. service_promote.go)
	promoteCheck      PromotePermissionChecker // опционально: RBAC-проверка целевого проекта при /promote-to-issue
	issueThreadLinker IssueThreadLinker        // опционально: auto-link после promote (PROD-03)
	messageSentHook   MessageSentHook          // опционально: @hub-ai bot (Phase 17 W2)
	systemBotID       uuid.UUID                // системный бот @hub-ai
}

// NewMessengerService создаёт новый сервис мессенджера.
// Repository должен реализовывать в том числе MentionsRepository, ChannelInviteRepository
// и PollRepository — иначе вызовы соответствующих методов вернут ошибку реализации.
func NewMessengerService(repo Repository, nats *natsclient.Client, presence *PresenceService, unread *UnreadService) *MessengerService {
	return newMessengerServiceWithEventPublisher(repo, nats, presence, unread)
}

// NewMessengerServiceWithAuthority is the production constructor. Every
// Mongo-touching consumer receives explicit PostgreSQL authority dependencies
// before handlers, Search or Dashboard can observe the service.
func NewMessengerServiceWithAuthority(
	repo Repository,
	nats *natsclient.Client,
	presence *PresenceService,
	channelAccess ChannelAccessResolver,
	unreadScope AuthorizedUnreadScopeResolver,
	messageSearchScope AuthorizedMessageSearchScopeResolver,
) *MessengerService {
	if repo == nil || channelAccess == nil || unreadScope == nil || messageSearchScope == nil {
		panic("messenger production authority dependencies are required")
	}
	service := newMessengerServiceWithEventPublisher(repo, nats, presence, nil)
	service.channelAccess = channelAccess
	service.unreadScope = unreadScope
	service.messageSearchScope = messageSearchScope
	return service
}
type noopEventPublisher struct{}

func (noopEventPublisher) PublishJSON(string, any) {}

func newMessengerServiceWithEventPublisher(repo Repository, publisher EventPublisher, presence *PresenceService, _ *UnreadService) *MessengerService {
	service := &MessengerService{
		repo:     repo,
		nats:     normalizeEventPublisher(publisher),
		presence: presence,
	}
	if durableUnread, ok := repo.(Phase151DurableUnreadRuntime); ok {
		service.durableUnread = durableUnread
	}
	return service
}

func normalizeEventPublisher(publisher EventPublisher) EventPublisher {
	if publisher == nil {
		return noopEventPublisher{}
	}
	if natsPublisher, ok := publisher.(*natsclient.Client); ok && natsPublisher == nil {
		return noopEventPublisher{}
	}
	return publisher
}

// ---------------------------------------------------------------------------
// Присутствие
// ---------------------------------------------------------------------------

// Heartbeat обновляет время активности пользователя.
func (s *MessengerService) Heartbeat(ctx context.Context, userID string) error {
	if s.presence == nil {
		return nil
	}
	return s.presence.Heartbeat(ctx, userID)
}

// GetPresence возвращает статусы присутствия для списка пользователей.
// W3-A1-007 / P2-BE-12: cap длины списка (DoS/enumeration).
// A2W3-1-001: только пользователи, с которыми caller делит ≥1 активный канал.
func (s *MessengerService) GetPresence(ctx context.Context, callerID uuid.UUID, userIDs []string) (map[string]string, error) {
	if s.presence == nil {
		return map[string]string{}, nil
	}
	const maxPresenceQueryIDs = 50
	if len(userIDs) > maxPresenceQueryIDs {
		return nil, apperrors.NewBadRequest("слишком много user_ids (макс. 50)")
	}
	candidates := make([]uuid.UUID, 0, len(userIDs))
	seen := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, raw := range userIDs {
		id, err := uuid.Parse(raw)
		if err != nil || id == uuid.Nil || id == callerID {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return map[string]string{}, nil
	}
	allowed, err := s.repo.FilterUsersSharingChannel(ctx, callerID, candidates)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки общих каналов")
	}
	if len(allowed) == 0 {
		return map[string]string{}, nil
	}
	allowedStrs := make([]string, len(allowed))
	for i, id := range allowed {
		allowedStrs[i] = id.String()
	}
	return s.presence.GetBulkStatus(ctx, allowedStrs)
}

// SetCustomStatus устанавливает кастомный текст статуса пользователя.
// duration — строковое значение длительности ("30m", "1h", "4h", "today", "1d", "1w", "0").
func (s *MessengerService) SetCustomStatus(ctx context.Context, userID, text, duration string) error {
	if s.presence == nil {
		return nil
	}
	return s.presence.SetCustomStatus(ctx, userID, text, duration)
}

// GetCustomStatus возвращает кастомный текст статуса пользователя.
func (s *MessengerService) GetCustomStatus(ctx context.Context, userID string) (string, error) {
	if s.presence == nil {
		return "", nil
	}
	return s.presence.GetCustomStatus(ctx, userID)
}

// ---------------------------------------------------------------------------
// Набор текста
// ---------------------------------------------------------------------------

// PublishTyping публикует событие набора текста в NATS (ephemeral, без записи в БД).
// A2-009: только участник канала может инжектить typing в NATS room канала.
func (s *MessengerService) PublishTyping(ctx context.Context, channelID string, userID uuid.UUID) error {
	chID, err := uuid.Parse(channelID)
	if err != nil {
		return apperrors.NewBadRequest("некорректный ID канала")
	}
	_, scoped, accessErr := s.resolveChannelAction(ctx, userID, chID, ChannelActionTyping)
	if accessErr != nil {
		return accessErr
	}
	if !scoped && s.repo != nil {
		ok, membershipErr := s.repo.IsMember(ctx, chID, userID)
		if membershipErr != nil {
			return apperrors.NewInternal("ошибка проверки участия")
		}
		if !ok {
			return apperrors.NewForbidden("вы не являетесь участником канала")
		}
	}

	// Определяем отображаемое имя пользователя через репозиторий.
	userName := "user"
	if s.repo != nil {
		if name, err := s.repo.GetUserDisplayName(ctx, userID); err == nil && name != "" {
			userName = name
		}
	}

	subject := fmt.Sprintf(SubjectTyping, channelID)
	s.nats.PublishJSON(subject, TypingEvent{
		ChannelID: channelID,
		UserID:    userID.String(),
		UserName:  userName,
	})
	return nil
}

// ---------------------------------------------------------------------------
// Непрочитанные
// ---------------------------------------------------------------------------

// GetUnreadCounts возвращает счётчики непрочитанных для пользователя.
func (s *MessengerService) GetUnreadCounts(ctx context.Context, userID string) (map[string]int64, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil || parsedUserID == uuid.Nil {
		return nil, apperrors.NewBadRequest("некорректный ID пользователя")
	}
	if s.durableUnread == nil {
		return map[string]int64{}, nil
	}
	return s.durableUnread.GetDurableUnreadCounts(ctx, parsedUserID)
}

// MarkChannelRead сбрасывает счётчик непрочитанных для канала.
func (s *MessengerService) MarkChannelRead(ctx context.Context, userID, channelID string) error {
	parsedUserID, userErr := uuid.Parse(userID)
	parsedChannelID, channelErr := uuid.Parse(channelID)
	if userErr != nil || channelErr != nil || parsedUserID == uuid.Nil || parsedChannelID == uuid.Nil {
		return apperrors.NewBadRequest("invalid user or channel ID")
	}
	if s.unreadScope == nil {
		return apperrors.NewForbidden("authorized unread scope is required")
	}
	scope, err := s.unreadScope.ResolveAuthorizedUnreadScope(ctx, parsedUserID)
	if err != nil || scope.TenantID == uuid.Nil || scope.UserID != parsedUserID {
		return apperrors.NewForbidden("authorized unread scope is unavailable")
	}
	var target *AuthorizedUnreadTarget
	for index := range scope.Targets {
		candidate := &scope.Targets[index]
		if candidate.ChannelID == parsedChannelID {
			target = candidate
			break
		}
	}
	if target == nil {
		return apperrors.NewForbidden("unread target is unavailable")
	}
	category := UnreadCategoryChannel
	if target.Kind == "dm" {
		category = UnreadCategoryDM
	} else if target.Kind != "channel" {
		return apperrors.NewForbidden("unread target kind is unavailable")
	}
	runtime, ok := s.repo.(phase151UnreadReceiptRuntime)
	if !ok {
		return apperrors.NewInternal("durable unread receipt runtime is not configured")
	}
	position, err := runtime.CurrentAuthorizedTargetHead(ctx, scope, *target, category)
	if err != nil {
		return apperrors.NewInternal("failed to resolve unread target head")
	}
	if position.CreatedAt.IsZero() && position.MessageID == "" {
		if err := runtime.PersistUnreadCutoverBaseline(ctx, scope, category, parsedChannelID.String(), position); err != nil {
			return apperrors.NewInternal("failed to persist empty unread baseline")
		}
		return nil
	}
	if _, err := runtime.AdvanceAuthorizedUnreadCursor(ctx, scope, category, parsedChannelID.String(), position); err != nil {
		return apperrors.NewInternal("failed to advance unread cursor")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------------------------

// resolveGroupMentions разрешает групповые упоминания (@all, @online, @role)
// в конкретные user ID. Объединяет с индивидуальными упоминаниями из поля Mentions.
// Исключает отправителя из результата.
// W3-A2-013 / P2-BE-13: individual Mentions тоже пересекаются с memberSet.
func (s *MessengerService) resolveGroupMentions(ctx context.Context, channelID uuid.UUID, projectID *uuid.UUID, groups []string, individualMentions []string, senderID uuid.UUID) []string {
	seen := make(map[string]struct{})
	senderStr := senderID.String()

	// Участники канала нужны и для individual mentions, и для group.
	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		slog.Error("не удалось получить участников канала для упоминаний", "error", err)
		return nil
	}
	memberSet := make(map[string]struct{}, len(memberIDs))
	for _, m := range memberIDs {
		memberSet[m.String()] = struct{}{}
	}

	// Индивидуальные упоминания — только члены канала (не spam UUID вне канала).
	for _, uid := range individualMentions {
		if uid == senderStr {
			continue
		}
		if _, ok := memberSet[uid]; ok {
			seen[uid] = struct{}{}
		}
	}

	if len(groups) == 0 {
		return mapKeys(seen)
	}

	for _, g := range groups {
		switch g {
		case "all":
			// @all — все участники канала.
			for _, m := range memberIDs {
				uid := m.String()
				if uid != senderStr {
					seen[uid] = struct{}{}
				}
			}

		case "online":
			// @online — онлайн-пользователи, пересечённые с участниками канала.
			if s.presence == nil {
				continue
			}
			onlineIDs, pErr := s.presence.GetOnlineUserIDs(ctx)
			if pErr != nil {
				slog.Error("не удалось получить онлайн-пользователей", "error", pErr)
				continue
			}
			for _, oid := range onlineIDs {
				if _, ok := memberSet[oid]; ok && oid != senderStr {
					seen[oid] = struct{}{}
				}
			}

		default:
			// Предполагаем имя роли проекта (owner, developer, tester, viewer).
			if projectID == nil {
				continue
			}
			roleMemberIDs, qErr := s.repo.ListProjectMembersByRole(ctx, *projectID, g)
			if qErr != nil {
				slog.Error("не удалось получить участников по роли", "role", g, "error", qErr)
				continue
			}
			for _, mid := range roleMemberIDs {
				uid := mid.String()
				// Учитываем только тех, кто является участником канала.
				if _, ok := memberSet[uid]; ok && uid != senderStr {
					seen[uid] = struct{}{}
				}
			}
		}
	}

	return mapKeys(seen)
}

// mapKeys возвращает ключи map[string]struct{} как срез строк.
func mapKeys(m map[string]struct{}) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// toChannelResponse конвертирует Ent Channel в DTO ответа.
func toChannelResponse(ch *ent.Channel) *ChannelResponse {
	resp := &ChannelResponse{
		ID:          ch.ID.String(),
		ChannelType: ch.ChannelType.String(),
		Name:        ch.Name,
		Description: ch.Description,
		IsReadonly:  ch.IsReadonly,
		CreatedBy:   ch.CreatedBy.String(),
		Version:     ch.Version,
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
	}
	if ch.ProjectID != nil {
		resp.ProjectID = ch.ProjectID.String()
	}
	return resp
}

// toRecipientChannelResponse maps one fresh authorization proof without role inference.
func toRecipientChannelResponse(access ChannelAccess) (*ChannelResponse, error) {
	if access.Channel == nil {
		return nil, ErrChannelAccessDenied
	}

	projection := ExportPhase151RecipientChannelProjection(access)
	response := toChannelResponse(access.Channel)
	response.Capabilities = &ChannelCapabilitiesResponse{
		IsMember:                   projection.IsMember,
		HasActiveProjectMembership: projection.HasActiveProjectMembership,
		CanRead:                    projection.CanRead,
		CanPost:                    projection.CanPost,
		CanReply:                   projection.CanReply,
		CanJoin:                    projection.CanJoin,
		CanFollow:                  projection.CanFollow,
		CanUnfollow:                projection.CanUnfollow,
		CanViewUnread:              projection.CanViewUnread,
		CanType:                    projection.CanType,
		CanEdit:                    projection.CanEdit,
		CanForward:                 projection.CanForward,
		CanReact:                   projection.CanReact,
		CanDeleteMessage:           projection.CanDeleteMessage,
		CanPin:                     projection.CanPin,
		CanListPins:                projection.CanListPins,
		CanBookmark:                projection.CanBookmark,
		CanPromote:                 projection.CanPromote,
	}
	return response, nil
}
