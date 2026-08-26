package messenger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

const phase151ChannelDiscoveryLimit = 100

type phase151ChannelDiscoveryRepository interface {
	ListDiscoverableChannels(context.Context, uuid.UUID, int) ([]*ent.Channel, error)
}

// ListDiscoverableChannels returns a bounded PostgreSQL-only candidate list.
// The service applies a second fresh ChannelAccess proof before projection.
func (r *EntRepository) ListDiscoverableChannels(
	ctx context.Context,
	actorID uuid.UUID,
	limit int,
) ([]*ent.Channel, error) {
	if r == nil || r.client == nil || actorID == uuid.Nil || limit < 1 || limit > phase151ChannelDiscoveryLimit {
		return nil, ErrChannelAccessDenied
	}

	tenantID, err := ExportPhase151ResolveInstanceOrganizationID(ctx, r)
	if err != nil {
		return nil, err
	}
	actorActive, err := r.client.User.Query().Where(
		user.IDEQ(actorID),
		user.StatusEQ(user.StatusActive),
		user.HasOrganizationWith(organization.IDEQ(tenantID)),
	).Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve active discovery actor: %w", err)
	}
	if !actorActive {
		return nil, ErrChannelAccessDenied
	}

	return r.client.Channel.Query().Where(
		channel.DeletedAtIsNil(),
		channel.ChannelTypeEQ(channel.ChannelTypePublic),
		channel.Or(
			channel.And(
				channel.ProjectIDIsNil(),
				channel.HasCreatorWith(user.HasOrganizationWith(organization.IDEQ(tenantID))),
			),
			channel.And(
				channel.ProjectIDNotNil(),
				channel.HasProjectWith(
					project.StatusEQ(project.StatusActive),
					project.DeletedAtIsNil(),
					project.HasOrganizationWith(organization.IDEQ(tenantID)),
					project.HasMembersWith(projectmember.HasUserWith(user.IDEQ(actorID))),
				),
			),
		),
	).Order(
		ent.Desc(channel.FieldCreatedAt),
		ent.Desc(channel.FieldID),
	).Limit(limit).All(ctx)
}

// CreateChannel создаёт канал и добавляет создателя как owner.
func (s *MessengerService) CreateChannel(ctx context.Context, userID uuid.UUID, req CreateChannelRequest) (*ChannelResponse, error) {
	input := CreateChannelInput{
		ChannelType: req.ChannelType,
		Name:        req.Name,
		Description: req.Description,
		IsReadonly:  req.IsReadonly,
		CreatedBy:   userID,
	}

	if req.ProjectID != "" {
		pid, err := uuid.Parse(req.ProjectID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный project_id")
		}
		// AUDIT-TM-004: канал с project_id только для участников проекта (fail-closed IDOR).
		ok, mErr := s.repo.IsProjectMember(ctx, pid, userID)
		if mErr != nil {
			return nil, apperrors.NewInternal("ошибка проверки участия в проекте")
		}
		if !ok {
			return nil, apperrors.NewForbidden("нет доступа к проекту")
		}
		input.ProjectID = &pid
	}

	ch, err := s.repo.CreateChannel(ctx, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания канала")
	}

	// Создатель всегда owner — иначе канал останется без управляющего (A2W3-2-002).
	if err := s.repo.AddMembers(ctx, ch.ID, []uuid.UUID{userID}, "owner"); err != nil {
		slog.Error("не удалось добавить создателя в канал", "error", err, "channel_id", ch.ID)
		if delErr := s.repo.DeleteChannel(ctx, ch.ID); delErr != nil {
			slog.Error("не удалось удалить orphan-канал", "error", delErr, "channel_id", ch.ID)
		}
		return nil, apperrors.NewInternal("ошибка создания канала")
	}

	// Добавление дополнительных участников
	if len(req.MemberIDs) > 0 {
		memberUUIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
		for _, mid := range req.MemberIDs {
			uid, err := uuid.Parse(mid)
			if err != nil {
				continue
			}
			if uid != userID {
				memberUUIDs = append(memberUUIDs, uid)
			}
		}
		if len(memberUUIDs) > 0 {
			if err := s.repo.AddMembers(ctx, ch.ID, memberUUIDs, "member"); err != nil {
				// A3W3-2-002: после успешного owner partial fail → rollback канала (симметрия orphan-owner).
				slog.Error("не удалось добавить участников в канал", "error", err, "channel_id", ch.ID)
				if delErr := s.repo.DeleteChannel(ctx, ch.ID); delErr != nil {
					slog.Error("не удалось удалить канал после fail доп. участников", "error", delErr, "channel_id", ch.ID)
				}
				return nil, apperrors.NewInternal("ошибка добавления участников канала")
			}
		}
	}

	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, ch.ID)
	if err != nil {
		slog.Error("channel lifecycle member lookup failed after create", "error", err, "channel_id", ch.ID)
	} else {
		s.publishUserChannelLifecycleEvent(SubjectUserChannelCreated, ch.ID, memberIDs)
	}

	return s.recipientChannelResponse(ctx, userID, ch)
}

// GetChannel возвращает канал по ID.
// HI-01: данные канала (включая DM/private) читают только участники.
func (s *MessengerService) GetChannel(ctx context.Context, channelID, userID uuid.UUID) (*ChannelResponse, error) {
	access, scoped, authErr := s.resolveChannelAction(
		ctx,
		userID,
		channelID,
		ChannelActionReadTimeline,
	)
	if authErr != nil {
		return nil, authErr
	}
	if scoped {
		return toRecipientChannelResponse(access)
	}

	isMember, err := s.repo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("канал не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения канала")
	}
	return toChannelResponse(ch), nil
}

// ListMyChannels возвращает все каналы пользователя.
// Обогащает ответ данными membership: is_starred, muted_until.
func (s *MessengerService) ListMyChannels(ctx context.Context, userID uuid.UUID) ([]ChannelResponse, error) {
	channels, err := s.repo.ListUserChannels(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения каналов")
	}

	// Загружаем membership-данные для обогащения ответа.
	memberships, mErr := s.repo.GetUserChannelMemberships(ctx, userID)
	membershipMap := make(map[string]*ent.ChannelMember)
	if mErr == nil {
		for _, m := range memberships {
			membershipMap[m.ChannelID.String()] = m
		}
	}

	result := make([]ChannelResponse, 0, len(channels))
	for _, candidate := range channels {
		access, scoped, accessErr := s.resolveChannelAction(
			ctx,
			userID,
			candidate.ID,
			ChannelActionReadTimeline,
		)
		if accessErr != nil {
			continue
		}

		resp := toChannelResponse(candidate)
		if scoped {
			resp, accessErr = toRecipientChannelResponse(access)
			if accessErr != nil {
				continue
			}
		}
		if m, ok := membershipMap[candidate.ID.String()]; ok {
			resp.IsStarred = m.IsStarred
			if m.MutedUntil != nil && m.MutedUntil.After(time.Now().UTC()) {
				resp.MutedUntil = m.MutedUntil
			}
		}
		result = append(result, *resp)
	}
	return result, nil
}

// DiscoverChannels returns only fresh recipient projections for public channels.
func (s *MessengerService) DiscoverChannels(
	ctx context.Context,
	actorID uuid.UUID,
) ([]ChannelResponse, error) {
	if s == nil || s.repo == nil || actorID == uuid.Nil || s.channelAccess == nil {
		return nil, apperrors.NewForbidden("messenger channel discovery denied")
	}
	discovery, ok := s.repo.(phase151ChannelDiscoveryRepository)
	if !ok {
		return nil, apperrors.NewInternal("messenger PostgreSQL discovery dependency is required")
	}

	candidates, err := discovery.ListDiscoverableChannels(ctx, actorID, phase151ChannelDiscoveryLimit)
	if err != nil {
		if errors.Is(err, ErrChannelAccessDenied) {
			return nil, apperrors.NewForbidden("messenger channel discovery denied")
		}
		return nil, apperrors.NewInternal("channel discovery failed")
	}

	result := make([]ChannelResponse, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil || candidate.ID == uuid.Nil || string(candidate.ChannelType) != "public" {
			continue
		}
		access, accessErr := s.channelAccess.AuthorizeChannelAccess(
			ctx,
			actorID,
			candidate.ID,
			ChannelActionReadTimeline,
		)
		if accessErr != nil {
			if errors.Is(accessErr, ErrChannelAccessDenied) {
				continue
			}
			return nil, apperrors.NewInternal("channel discovery authorization failed")
		}
		if access.TenantID == uuid.Nil ||
			access.Channel == nil ||
			access.Channel.ID != candidate.ID ||
			string(access.Channel.ChannelType) != "public" ||
			!access.CanRead ||
			(access.Channel.ProjectID != nil && !access.HasActiveProjectMembership) {
			continue
		}
		response, responseErr := toRecipientChannelResponse(access)
		if responseErr != nil {
			return nil, apperrors.NewInternal("channel discovery projection failed")
		}
		result = append(result, *response)
	}
	return result, nil
}

// UpdateChannel обновляет канал (только owner/admin).
// TASK-095: version обязателен; VersionEQ в repo → 409 при conflict.
func (s *MessengerService) UpdateChannel(ctx context.Context, channelID, userID uuid.UUID, req CreateChannelRequest) (*ChannelResponse, error) {
	if req.Version < 1 {
		return nil, apperrors.NewBadRequest("version обязателен и должен быть ≥ 1")
	}
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return nil, apperrors.NewForbidden("только owner или admin может редактировать канал")
	}

	input := UpdateChannelInput{Version: req.Version}
	if req.Name != "" {
		input.Name = &req.Name
	}
	if req.Description != "" {
		input.Description = &req.Description
	}
	input.IsReadonly = &req.IsReadonly

	ch, err := s.repo.UpdateChannel(ctx, channelID, input)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewConflict("version mismatch: канал был изменён другим пользователем")
		}
		return nil, apperrors.NewInternal("ошибка обновления канала")
	}
	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		slog.Error("channel lifecycle member lookup failed after update", "error", err, "channel_id", channelID)
	} else {
		s.publishUserChannelLifecycleEvent(SubjectUserChannelUpdated, channelID, memberIDs)
	}
	return s.recipientChannelResponse(ctx, userID, ch)
}

// DeleteChannel удаляет канал (только owner).
// A2W3-3-003: чистит unread hash fields у всех участников до soft-delete.
func (s *MessengerService) DeleteChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" {
		return apperrors.NewForbidden("только owner может удалить канал")
	}
	memberIDs, err := s.repo.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		return apperrors.NewInternal("ошибка получения участников канала")
	}
	// A3W3-2-006: сырой repo/Ent err не утекает — только AppError.
	if err := s.repo.DeleteChannel(ctx, channelID); err != nil {
		var ae *apperrors.AppError
		if errors.As(err, &ae) {
			return ae
		}
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("канал не найден")
		}
		return apperrors.NewInternal("ошибка удаления канала")
	}
	s.publishUserChannelLifecycleEvent(SubjectUserChannelRemoved, channelID, memberIDs)
	return nil
}

// JoinChannel позволяет вступить в публичный канал.
func (s *MessengerService) JoinChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	ch, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("канал не найден")
		}
		return apperrors.NewInternal("ошибка получения канала")
	}
	switch ch.ChannelType.String() {
	case "public":
		// Публичный канал — свободное вступление
	case "private":
		return apperrors.NewForbidden("канал приватный, требуется приглашение")
	case "dm":
		return apperrors.NewForbidden("нельзя присоединиться к DM, используйте CreateOrGetDM")
	default:
		return apperrors.NewForbidden("можно присоединиться только к публичному каналу")
	}

	// Уже member — идемpotent OK (A3W3-2-001: err IsMember не глотать — зеркало AcceptInvite).
	isMember, err := s.repo.IsMember(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewInternal("ошибка проверки участия")
	}
	if isMember {
		return nil
	}

	if err := s.repo.AddMembers(ctx, channelID, []uuid.UUID{userID}, "member"); err != nil {
		return apperrors.NewInternal("ошибка вступления в канал")
	}

	// Публикация NATS-события
	s.nats.PublishJSON(fmt.Sprintf(SubjectMemberJoined, channelID.String()), map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})
	s.publishUserChannelLifecycleEvent(SubjectUserChannelCreated, channelID, []uuid.UUID{userID})

	return nil
}

// LeaveChannel позволяет покинуть канал.
// W3-A2-012 / P2-BE-13: единственный owner не может уйти без transfer/delete.
func (s *MessengerService) LeaveChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewForbidden("нет доступа к каналу")
	}
	if role == "owner" {
		members, listErr := s.repo.ListMembers(ctx, channelID)
		if listErr != nil {
			return apperrors.NewInternal("ошибка проверки участников канала")
		}
		ownerCount := 0
		for _, m := range members {
			if string(m.Role) == "owner" {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return apperrors.NewBadRequest("единственный owner не может покинуть канал; передайте роль или удалите канал")
		}
	}

	if err := s.repo.RemoveMember(ctx, channelID, userID); err != nil {
		return apperrors.NewInternal("ошибка выхода из канала")
	}


	s.nats.PublishJSON(fmt.Sprintf(SubjectMemberLeft, channelID.String()), map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})
	s.publishUserChannelLifecycleEvent(SubjectUserChannelRemoved, channelID, []uuid.UUID{userID})

	return nil
}

// AddMembers добавляет участников в канал (только admin/owner).
func (s *MessengerService) AddMembers(ctx context.Context, channelID, userID uuid.UUID, memberIDs []string) error {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return apperrors.NewForbidden("только owner или admin может добавлять участников")
	}

	candidates := make([]uuid.UUID, 0, len(memberIDs))
	seenCandidates := make(map[uuid.UUID]struct{}, len(memberIDs))
	for _, mid := range memberIDs {
		uid, err := uuid.Parse(mid)
		if err != nil {
			return apperrors.NewBadRequest("некорректный member ID: " + mid)
		}
		if uid == uuid.Nil {
			return apperrors.NewBadRequest("некорректный member ID: " + mid)
		}
		if _, duplicate := seenCandidates[uid]; duplicate {
			continue
		}
		seenCandidates[uid] = struct{}{}
		candidates = append(candidates, uid)
	}

	existingMemberIDs, err := s.repo.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		return apperrors.NewInternal("ошибка получения участников")
	}
	existingMembers := make(map[uuid.UUID]struct{}, len(existingMemberIDs))
	for _, existingID := range existingMemberIDs {
		existingMembers[existingID] = struct{}{}
	}
	newMemberIDs := make([]uuid.UUID, 0, len(candidates))
	for _, candidateID := range candidates {
		if _, exists := existingMembers[candidateID]; exists {
			continue
		}
		newMemberIDs = append(newMemberIDs, candidateID)
	}
	if len(newMemberIDs) == 0 {
		return nil
	}

	// A3W3-2-006: wrap repo err → AppError.
	if err := s.repo.AddMembers(ctx, channelID, newMemberIDs, "member"); err != nil {
		var ae *apperrors.AppError
		if errors.As(err, &ae) {
			return ae
		}
		return apperrors.NewInternal("ошибка добавления участников")
	}
	s.publishUserChannelLifecycleEvent(SubjectUserChannelCreated, channelID, newMemberIDs)
	return nil
}

// AddMembersInternal добавляет участников без проверки прав.
// Используется внутренними сервисами (например AutoChannelService).
func (s *MessengerService) AddMembersInternal(ctx context.Context, channelID uuid.UUID, memberIDs []uuid.UUID) error {
	if channelID == uuid.Nil {
		return apperrors.NewBadRequest("некорректный channel ID")
	}
	if len(memberIDs) == 0 {
		return nil
	}

	candidates := make([]uuid.UUID, 0, len(memberIDs))
	seenCandidates := make(map[uuid.UUID]struct{}, len(memberIDs))
	for _, memberID := range memberIDs {
		if memberID == uuid.Nil {
			return apperrors.NewBadRequest("некорректный member ID")
		}
		if _, duplicate := seenCandidates[memberID]; duplicate {
			continue
		}
		seenCandidates[memberID] = struct{}{}
		candidates = append(candidates, memberID)
	}

	existingMemberIDs, err := s.repo.GetChannelMemberIDs(ctx, channelID)
	if err != nil {
		return apperrors.NewInternal("ошибка получения участников")
	}
	existingMembers := make(map[uuid.UUID]struct{}, len(existingMemberIDs))
	for _, existingMemberID := range existingMemberIDs {
		existingMembers[existingMemberID] = struct{}{}
	}

	newMemberIDs := make([]uuid.UUID, 0, len(candidates))
	for _, candidate := range candidates {
		if _, exists := existingMembers[candidate]; !exists {
			newMemberIDs = append(newMemberIDs, candidate)
		}
	}
	if len(newMemberIDs) == 0 {
		return nil
	}

	if err := s.repo.AddMembers(ctx, channelID, newMemberIDs, "member"); err != nil {
		var ae *apperrors.AppError
		if errors.As(err, &ae) {
			return ae
		}
		return apperrors.NewInternal("ошибка добавления участников")
	}

	s.publishUserChannelLifecycleEvent(SubjectUserChannelCreated, channelID, newMemberIDs)
	return nil
}

// ToggleStar переключает флаг is_starred для пользователя в канале.
// A3W3-2-004: non-member (repo NotFound) → 403, не 500 Internal.
func (s *MessengerService) ToggleStar(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	starred, err := s.repo.ToggleStar(ctx, channelID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, apperrors.NewForbidden("вы не являетесь участником канала")
		}
		return false, apperrors.NewInternal("ошибка переключения закрепления")
	}
	return starred, nil
}

// MuteChannel устанавливает muted_until для пользователя в канале.
// duration: "1h", "2h", "8h", "1d", "forever" (значения API).
// A3W3-2-004: non-member (repo NotFound) → 403, не 500 Internal.
func (s *MessengerService) MuteChannel(ctx context.Context, channelID, userID uuid.UUID, duration string) (string, error) {
	mutedUntil, err := s.repo.MuteChannel(ctx, channelID, userID, duration)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", apperrors.NewForbidden("вы не являетесь участником канала")
		}
		return "", apperrors.NewInternal("ошибка мьюта канала")
	}
	return mutedUntil.Format("2006-01-02T15:04:05Z07:00"), nil
}

// UnmuteChannel сбрасывает muted_until для пользователя в канале.
// A3W3-2-004: non-member (repo NotFound) → 403, не 500 Internal.
func (s *MessengerService) UnmuteChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	if err := s.repo.UnmuteChannel(ctx, channelID, userID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewForbidden("вы не являетесь участником канала")
		}
		return apperrors.NewInternal("ошибка снятия мьюта канала")
	}
	return nil
}

// IsChannelMuted проверяет, замьючен ли канал для пользователя.
func (s *MessengerService) IsChannelMuted(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return s.repo.IsChannelMuted(ctx, channelID, userID)
}

// ListMembers возвращает всех участников канала.
// HI-01: список участников видят только members канала.
func (s *MessengerService) ListMembers(ctx context.Context, channelID, userID uuid.UUID) ([]ChannelMemberResponse, error) {
	isMember, err := s.repo.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if !isMember {
		return nil, apperrors.NewForbidden("вы не являетесь участником канала")
	}

	members, err := s.repo.ListMembers(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения участников")
	}
	result := make([]ChannelMemberResponse, 0, len(members))
	for _, m := range members {
		resp := ChannelMemberResponse{
			ID:       m.ID.String(),
			UserID:   m.UserID.String(),
			Role:     string(m.Role),
			JoinedAt: m.JoinedAt,
		}
		if m.Edges.User != nil {
			resp.UserName = m.Edges.User.Name
		}
		result = append(result, resp)
	}
	return result, nil
}
