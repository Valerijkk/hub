package messenger

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// base62Chars — алфавит для генерации токенов пригласительных ссылок.
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// inviteTokenLength — длина токена пригласительной ссылки (22 символа).
const inviteTokenLength = 22

// generateInviteToken генерирует криптографически стойкий base62-токен.
func generateInviteToken() (string, error) {
	b := make([]byte, inviteTokenLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", fmt.Errorf("ошибка генерации токена: %w", err)
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b), nil
}

// parseExpiresIn преобразует строку срока действия в time.Duration.
// Поддерживаемые значения: "1h", "1d", "7d", "30d", "" (бессрочно).
func parseExpiresIn(expiresIn string) (*time.Time, error) {
	if expiresIn == "" {
		return nil, nil
	}

	var d time.Duration
	switch expiresIn {
	case "1h":
		d = time.Hour
	case "1d":
		d = 24 * time.Hour
	case "7d":
		d = 7 * 24 * time.Hour
	case "30d":
		d = 30 * 24 * time.Hour
	default:
		return nil, fmt.Errorf("некорректный срок действия: %s", expiresIn)
	}

	t := time.Now().UTC().Add(d)
	return &t, nil
}

// CreateInviteLink создаёт пригласительную ссылку для канала.
// Требуется роль owner или admin.
func (s *MessengerService) CreateInviteLink(ctx context.Context, channelID, userID uuid.UUID, req CreateInviteLinkRequest) (*ChannelInviteResponse, error) {
	// Проверяем права: только owner/admin может создавать ссылки.
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return nil, apperrors.NewForbidden("только owner или admin может создавать пригласительные ссылки")
	}

	// Генерируем уникальный токен.
	token, err := generateInviteToken()
	if err != nil {
		return nil, apperrors.NewInternal("ошибка генерации токена приглашения")
	}

	// Определяем срок действия.
	expiresAt, err := parseExpiresIn(req.ExpiresIn)
	if err != nil {
		return nil, apperrors.NewBadRequest(err.Error())
	}

	// Создаём запись в БД через репозиторий.
	invite, err := s.repo.CreateChannelInvite(ctx, CreateChannelInviteInput{
		ChannelID: channelID,
		CreatorID: userID,
		Token:     token,
		MaxUses:   req.MaxUses,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания пригласительной ссылки")
	}

	return toChannelInviteResponse(invite), nil
}

// ValidateInvite проверяет валидность пригласительной ссылки по токену.
// Публичный метод — не требует аутентификации.
func (s *MessengerService) ValidateInvite(ctx context.Context, token string) (*InviteValidationResponse, error) {
	// Ищем приглашение по токену через репозиторий.
	found, err := s.repo.FindChannelInviteByToken(ctx, token)
	if err != nil {
		if ent.IsNotFound(err) {
			return &InviteValidationResponse{Valid: false, Message: "приглашение не найдено"}, nil
		}
		return &InviteValidationResponse{Valid: false, Message: "ошибка проверки приглашения"}, nil
	}

	// Проверяем, не отозвано ли приглашение.
	if found.IsRevoked {
		return &InviteValidationResponse{Valid: false, Message: "приглашение отозвано"}, nil
	}

	// Проверяем срок действия.
	if found.ExpiresAt != nil && found.ExpiresAt.Before(time.Now().UTC()) {
		return &InviteValidationResponse{Valid: false, Message: "срок действия приглашения истёк"}, nil
	}

	// Проверяем лимит использований.
	if found.MaxUses > 0 && found.UseCount >= found.MaxUses {
		return &InviteValidationResponse{Valid: false, Message: "исчерпан лимит использований"}, nil
	}

	// Получаем информацию о канале.
	ch, err := s.repo.GetChannel(ctx, found.ChannelID)
	if err != nil {
		return &InviteValidationResponse{Valid: false, Message: "канал не найден"}, nil
	}

	// Считаем участников.
	members, _ := s.repo.GetChannelMemberIDs(ctx, found.ChannelID)
	memberCount := len(members)

	return &InviteValidationResponse{
		Valid:       true,
		ChannelName: ch.Name,
		ChannelType: ch.ChannelType.String(),
		ChannelID:   ch.ID.String(),
		MemberCount: memberCount,
	}, nil
}

// AcceptInvite принимает пригласительную ссылку — добавляет пользователя в канал.
// Атомарный инкремент use_count с проверкой лимита.
func (s *MessengerService) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) (*ChannelResponse, error) {
	// Находим приглашение по токену через репозиторий.
	found, err := s.repo.FindChannelInviteByToken(ctx, token)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("приглашение не найдено")
		}
		return nil, apperrors.NewInternal("ошибка поиска приглашения")
	}

	// Проверяем, не является ли пользователь уже участником (A2W3-2-004: err не глотать).
	isMember, err := s.repo.IsMember(ctx, found.ChannelID, userID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка проверки участия")
	}
	if isMember {
		// Уже участник — просто возвращаем канал.
		ch, err := s.repo.GetChannel(ctx, found.ChannelID)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка получения канала")
		}
		return toChannelResponse(ch), nil
	}

	// Валидация приглашения.
	if found.IsRevoked {
		return nil, apperrors.NewBadRequest("приглашение отозвано")
	}
	if found.ExpiresAt != nil && found.ExpiresAt.Before(time.Now().UTC()) {
		return nil, apperrors.NewBadRequest("срок действия приглашения истёк")
	}
	if found.MaxUses > 0 && found.UseCount >= found.MaxUses {
		return nil, apperrors.NewBadRequest("исчерпан лимит использований")
	}

	// A2W3-2-001 / A2W3-3-001: атомарный инкремент с UseCountLT.
	if err := s.repo.IncrementInviteUses(ctx, found.ID, found.MaxUses); err != nil {
		if errors.Is(err, ErrInviteUsesExhausted) {
			return nil, apperrors.NewBadRequest("исчерпан лимит использований")
		}
		return nil, apperrors.NewInternal("ошибка обновления счётчика использований")
	}

	// Добавляем пользователя в канал; при fail — откат слота (A2W3-2-001).
	if err := s.repo.AddMembers(ctx, found.ChannelID, []uuid.UUID{userID}, "member"); err != nil {
		if decErr := s.repo.DecrementInviteUses(ctx, found.ID); decErr != nil {
			slog.Error("не удалось откатить use_count invite", "error", decErr, "invite_id", found.ID)
		}
		return nil, apperrors.NewInternal("ошибка вступления в канал")
	}

	// Публикуем NATS-событие без raw invite token (W3-A2-010 / P2-BE-13).
	s.nats.PublishJSON(fmt.Sprintf(SubjectMemberJoined, found.ChannelID.String()), map[string]interface{}{
		"channel_id": found.ChannelID,
		"user_id":    userID,
		"via_invite": true,
		"invite_id":  found.ID,
	})
	s.publishUserChannelLifecycleEvent(SubjectUserChannelCreated, found.ChannelID, []uuid.UUID{userID})

	// Возвращаем канал.
	ch, err := s.repo.GetChannel(ctx, found.ChannelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения канала")
	}
	return toChannelResponse(ch), nil
}

// RevokeInvite отзывает пригласительную ссылку.
// Требуется роль owner или admin.
func (s *MessengerService) RevokeInvite(ctx context.Context, channelID, inviteID, userID uuid.UUID) error {
	// Проверяем права.
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return apperrors.NewForbidden("только owner или admin может отзывать приглашения")
	}

	// Отзыв только invite этого канала (P1-BE-19); чужой inviteID → 404.
	if err := s.repo.RevokeChannelInvite(ctx, channelID, inviteID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("приглашение не найдено")
		}
		return apperrors.NewInternal("ошибка отзыва приглашения")
	}

	return nil
}

// ListChannelInvites возвращает список пригласительных ссылок канала.
// Требуется роль owner или admin.
func (s *MessengerService) ListChannelInvites(ctx context.Context, channelID, userID uuid.UUID) ([]ChannelInviteResponse, error) {
	// Проверяем права.
	role, err := s.repo.GetMemberRole(ctx, channelID, userID)
	if err != nil {
		return nil, apperrors.NewForbidden("нет доступа к каналу")
	}
	if role != "owner" && role != "admin" {
		return nil, apperrors.NewForbidden("только owner или admin может просматривать приглашения")
	}

	// Получаем приглашения канала через репозиторий.
	invites, err := s.repo.ListChannelInvites(ctx, channelID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения приглашений")
	}

	result := make([]ChannelInviteResponse, 0, len(invites))
	for _, inv := range invites {
		result = append(result, *toChannelInviteResponse(inv))
	}

	return result, nil
}

// toChannelInviteResponse конвертирует Ent ChannelInvite в DTO ответа.
func toChannelInviteResponse(inv *ent.ChannelInvite) *ChannelInviteResponse {
	resp := &ChannelInviteResponse{
		ID:        inv.ID,
		ChannelID: inv.ChannelID,
		CreatorID: inv.CreatorID,
		Token:     inv.Token,
		InviteURL: "/invite/" + inv.Token,
		MaxUses:   inv.MaxUses,
		UseCount:  inv.UseCount,
		ExpiresAt: inv.ExpiresAt,
		IsRevoked: inv.IsRevoked,
		CreatedAt: inv.CreatedAt,
	}
	return resp
}
