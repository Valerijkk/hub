// Package calls реализует голосовые созвоны через LiveKit: создание, присоединение, запись, планирование.
package calls

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/platform/config"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ErrCallForbidden — HI-03: 403 для JoinCall/EndCall без утечки причины.
var ErrCallForbidden = errors.New("forbidden")

// defaultMaxCallParticipants — потолок участников комнаты LiveKit.
const defaultMaxCallParticipants = 50

// CallService — тонкий orchestrator модуля звонков (SC-14-2).
// Делегирует LiveKit → CallRoomService, NATS → CallEventPublisher,
// чат → CallChatService, БД → EntRepository.
type CallService struct {
	repo *EntRepository
	room *CallRoomService
	pub  *CallEventPublisher
	chat *CallChatService
}

// NewCallService собирает orchestrator. Сигнатура зафиксирована (cmd/server/main.go).
func NewCallService(cfg config.LiveKitConfig, repo *EntRepository, nc *natsclient.Client, chatRepo *ChatRepository) *CallService {
	pub := NewCallEventPublisher(nc)
	return &CallService{
		repo: repo,
		room: NewCallRoomService(cfg),
		pub:  pub,
		chat: NewCallChatService(repo, chatRepo, pub),
	}
}

// GenerateToken — JWT для подключения к комнате LiveKit.
func (s *CallService) GenerateToken(roomName, userID, userName string) (string, error) {
	return s.room.GenerateToken(roomName, userID, userName)
}

// CreateCall создаёт звонок в канале (room + БД + участник + событие старта).
func (s *CallService) CreateCall(ctx context.Context, channelID, userID uuid.UUID, userName string) (*CallResponse, error) {
	roomName, err := provisionRoom(ctx, s.room)
	if err != nil {
		return nil, err
	}
	callEnt, err := s.repo.CreateCall(ctx, roomName, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить звонок: %w", err)
	}
	resp, err := finalizeCreate(ctx, s.room, s.repo, callEnt, userID, userName)
	if err != nil {
		return nil, err
	}
	if channelID != uuid.Nil {
		s.pub.Started(callEnt.ChannelID, callEnt.ID, roomName, userID)
	}
	return resp, nil
}

// CreateInstantCall создаёт быстрый звонок без канала с invite-кодом.
func (s *CallService) CreateInstantCall(ctx context.Context, userID uuid.UUID, userName, title string) (*CallResponse, error) {
	roomName, err := provisionRoom(ctx, s.room)
	if err != nil {
		return nil, err
	}
	inviteCode := uuid.New().String()[:8]
	if title == "" {
		title = userName + " — быстрый звонок"
	}
	callEnt, err := s.repo.CreateInstantCall(ctx, roomName, userID, inviteCode, title)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить звонок: %w", err)
	}
	resp, err := finalizeCreate(ctx, s.room, s.repo, callEnt, userID, userName)
	if err != nil {
		return nil, err
	}
	resp.InviteCode = inviteCode
	return resp, nil
}

// JoinByInviteCode присоединяет пользователя к звонку по invite-коду.
func (s *CallService) JoinByInviteCode(ctx context.Context, code string, userID uuid.UUID, userName string) (*CallResponse, error) {
	callEnt, err := s.repo.GetCallByInviteCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}
	if err := s.repo.AddParticipant(ctx, callEnt.ID, userID); err != nil {
		return nil, fmt.Errorf("не удалось добавить участника: %w", err)
	}
	token, err := s.room.GenerateToken(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}
	resp := toCallResponse(callEnt, token)
	resp.InviteCode = code
	return resp, nil
}

// JoinCall присоединяет к существующему звонку. HI-03: канальные — членство в канале;
// instant — только создатель (гости приходят через invite-код).
func (s *CallService) JoinCall(ctx context.Context, callID, userID uuid.UUID, userName string) (*CallResponse, error) {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}
	if callEnt.ChannelID != nil {
		isMember, mErr := s.repo.IsChannelMember(ctx, *callEnt.ChannelID, userID)
		if mErr != nil {
			return nil, fmt.Errorf("ошибка проверки членства в канале: %w", mErr)
		}
		if !isMember {
			return nil, ErrCallForbidden
		}
	} else if callEnt.CreatedBy != userID {
		return nil, ErrCallForbidden
	}
	if err := s.repo.AddParticipant(ctx, callID, userID); err != nil {
		return nil, fmt.Errorf("не удалось добавить участника: %w", err)
	}
	token, err := s.room.GenerateToken(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}
	s.pub.Joined(callEnt.ChannelID, callID, userID, userName)
	return toCallResponse(callEnt, token), nil
}

// LeaveCall обрабатывает выход пользователя из звонка.
func (s *CallService) LeaveCall(ctx context.Context, callID, userID uuid.UUID) error {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("звонок не найден: %w", err)
	}
	if err := s.repo.RemoveParticipant(ctx, callID, userID, time.Now()); err != nil {
		return fmt.Errorf("не удалось удалить участника: %w", err)
	}
	s.pub.Left(callEnt.ChannelID, callID, userID)
	return nil
}

// EndCall завершает звонок. HI-03: только создатель или владелец/админ канала.
func (s *CallService) EndCall(ctx context.Context, callID, userID uuid.UUID) error {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.CreatedBy != userID {
		if callEnt.ChannelID == nil {
			return ErrCallForbidden
		}
		role, roleErr := s.repo.GetChannelMemberRole(ctx, *callEnt.ChannelID, userID)
		if roleErr != nil {
			return fmt.Errorf("ошибка проверки роли в канале: %w", roleErr)
		}
		if role != "owner" && role != "admin" {
			return ErrCallForbidden
		}
	}
	if delErr := s.room.DeleteRoom(ctx, callEnt.RoomName); delErr != nil {
		slog.Error("не удалось удалить комнату LiveKit", "room", callEnt.RoomName, "error", delErr)
	}
	if err := s.repo.UpdateCallEnded(ctx, callID, time.Now(), callEnt.ParticipantCount); err != nil {
		return fmt.Errorf("не удалось обновить статус звонка: %w", err)
	}
	s.pub.Ended(callEnt.ChannelID, callID, callEnt.ParticipantCount)
	return nil
}

// GetActiveCallByChannel — активный звонок в канале или nil.
func (s *CallService) GetActiveCallByChannel(ctx context.Context, channelID uuid.UUID) (*CallResponse, error) {
	callEnt, err := s.repo.GetActiveCallByChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска активного звонка: %w", err)
	}
	if callEnt == nil {
		return nil, nil
	}
	return toCallResponse(callEnt, ""), nil
}

// SetRaiseHand обновляет LiveKit-метаданные участника ("поднять руку").
func (s *CallService) SetRaiseHand(ctx context.Context, roomName, identity string, raised bool) error {
	meta := fmt.Sprintf(`{"hand_raised":%v,"raised_at":"%s"}`, raised, time.Now().Format(time.RFC3339))
	return s.room.UpdateParticipantMetadata(ctx, roomName, identity, meta)
}

// OnRoomStarted — webhook room_started от LiveKit.
func (s *CallService) OnRoomStarted(ctx context.Context, roomName string) error {
	return s.repo.UpdateCallStarted(ctx, roomName, time.Now())
}

// OnRoomFinished — webhook room_finished: завершает звонок и чистит чат (CALL-06).
func (s *CallService) OnRoomFinished(ctx context.Context, roomName string) error {
	callEnt, err := s.repo.GetCallByRoomName(ctx, roomName)
	if err != nil {
		return fmt.Errorf("звонок не найден по имени комнаты: %w", err)
	}
	if err := s.repo.UpdateCallEnded(ctx, callEnt.ID, time.Now(), callEnt.ParticipantCount); err != nil {
		return err
	}
	s.chat.DeleteCallChat(ctx, callEnt.ID.String())
	s.pub.Ended(callEnt.ChannelID, callEnt.ID, callEnt.ParticipantCount)
	return nil
}

// SendChatMessage делегирует в CallChatService.
func (s *CallService) SendChatMessage(ctx context.Context, callID string, userID, userName, text string, ephemeral bool) (*CallChatMessage, error) {
	return s.chat.SendMessage(ctx, callID, userID, userName, text, ephemeral)
}

// ListChatMessages делегирует в CallChatService (A2-003: userID для authz).
func (s *CallService) ListChatMessages(ctx context.Context, callID, userID string, limit int) ([]CallChatMessage, error) {
	return s.chat.ListMessages(ctx, callID, userID, limit)
}

// OnParticipantJoined — webhook participant_joined от LiveKit.
func (s *CallService) OnParticipantJoined(ctx context.Context, roomName, identity string) error {
	callEnt, userID, err := resolveParticipant(ctx, s.repo, roomName, identity)
	if err != nil || callEnt == nil {
		return err
	}
	return s.repo.AddParticipant(ctx, callEnt.ID, userID)
}

// OnParticipantLeft — webhook participant_left от LiveKit.
func (s *CallService) OnParticipantLeft(ctx context.Context, roomName, identity string) error {
	callEnt, userID, err := resolveParticipant(ctx, s.repo, roomName, identity)
	if err != nil || callEnt == nil {
		return err
	}
	return s.repo.RemoveParticipant(ctx, callEnt.ID, userID, time.Now())
}
