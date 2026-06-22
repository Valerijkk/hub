package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

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
		input.ProjectID = &pid
	}

	ch, err := s.repo.CreateChannel(ctx, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания канала")
	}

	// Создатель всегда owner — иначе канал останется без управляющего.
	if err := s.repo.AddMembers(ctx, ch.ID, []uuid.UUID{userID}, "owner"); err != nil {
		slog.Error("не удалось добавить создателя в канал", "error", err)
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
				slog.Error("не удалось добавить участников в канал", "error", err)
			}
		}
	}

	return toChannelResponse(ch), nil
}

// GetChannel возвращает канал по ID.
// HI-01: данные канала (включая DM/private) читают только участники.
func (s *MessengerService) GetChannel(ctx context.Context, channelID, userID uuid.UUID) (*ChannelResponse, error) {
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
	for _, ch := range channels {
		resp := toChannelResponse(ch)
		if m, ok := membershipMap[ch.ID.String()]; ok {
			resp.IsStarred = m.IsStarred
			if m.MutedUntil != nil && m.MutedUntil.After(time.Now().UTC()) {
				resp.MutedUntil = m.MutedUntil
			}
		}
		result = append(result, *resp)
	}
	return result, nil
}

// UpdateChannel обновляет канал (только owner/admin).
func (s *MessengerService) UpdateChannel(ctx context.Context, channelID, userID uuid.UUID, req CreateChannelRequest) (*ChannelResponse, error) {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return nil, apperrors.NewForbidden("только owner или admin может редактировать канал")
	}

	input := UpdateChannelInput{}
	if req.Name != "" {
		input.Name = &req.Name
	}
	if req.Description != "" {
		input.Description = &req.Description
	}
	input.IsReadonly = &req.IsReadonly

	ch, err := s.repo.UpdateChannel(ctx, channelID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка обновления канала")
	}
	return toChannelResponse(ch), nil
}

// DeleteChannel удаляет канал (только owner).
func (s *MessengerService) DeleteChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" {
		return apperrors.NewForbidden("только owner может удалить канал")
	}
	return s.repo.DeleteChannel(ctx, channelID)
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

	// Уже member — идempotent OK
	isMember, _ := s.repo.IsMember(ctx, channelID, userID)
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

	return nil
}

// LeaveChannel позволяет покинуть канал.
func (s *MessengerService) LeaveChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	if err := s.repo.RemoveMember(ctx, channelID, userID); err != nil {
		return apperrors.NewInternal("ошибка выхода из канала")
	}

	s.nats.PublishJSON(fmt.Sprintf(SubjectMemberLeft, channelID.String()), map[string]interface{}{
		"channel_id": channelID,
		"user_id":    userID,
	})

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

	uuids := make([]uuid.UUID, 0, len(memberIDs))
	for _, mid := range memberIDs {
		uid, err := uuid.Parse(mid)
		if err != nil {
			return apperrors.NewBadRequest("некорректный member ID: " + mid)
		}
		uuids = append(uuids, uid)
	}

	return s.repo.AddMembers(ctx, channelID, uuids, "member")
}

// AddMembersInternal добавляет участников без проверки прав.
// Используется внутренними сервисами (например AutoChannelService).
func (s *MessengerService) AddMembersInternal(ctx context.Context, channelID uuid.UUID, memberIDs []uuid.UUID) error {
	return s.repo.AddMembers(ctx, channelID, memberIDs, "member")
}

// ToggleStar переключает флаг is_starred для пользователя в канале.
func (s *MessengerService) ToggleStar(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	starred, err := s.repo.ToggleStar(ctx, channelID, userID)
	if err != nil {
		return false, apperrors.NewInternal("ошибка переключения закрепления")
	}
	return starred, nil
}

// MuteChannel устанавливает muted_until для пользователя в канале.
// duration: "1h", "2h", "8h", "1d", "forever" (значения API).
func (s *MessengerService) MuteChannel(ctx context.Context, channelID, userID uuid.UUID, duration string) (string, error) {
	mutedUntil, err := s.repo.MuteChannel(ctx, channelID, userID, duration)
	if err != nil {
		return "", apperrors.NewInternal("ошибка мьюта канала")
	}
	return mutedUntil.Format("2006-01-02T15:04:05Z07:00"), nil
}

// UnmuteChannel сбрасывает muted_until для пользователя в канале.
func (s *MessengerService) UnmuteChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	if err := s.repo.UnmuteChannel(ctx, channelID, userID); err != nil {
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
