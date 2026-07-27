package calls

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// CallChatService инкапсулирует эфемерный чат звонков.
// Делегирует персистенцию ChatRepository (MongoDB), а публикацию событий —
// CallEventPublisher. Оркестратор (CallService) обращается только к нему,
// не зная про MongoDB и subject-строки (SC-14-2).
type CallChatService struct {
	repo *EntRepository      // проверка статуса звонка перед отправкой
	chat *ChatRepository     // MongoDB-персистенция; может быть nil
	pub  *CallEventPublisher // публикация события в NATS
}

// NewCallChatService создаёт сервис чата звонков.
// chat может быть nil — в этом случае SendMessage(ephemeral=false) вернёт ошибку,
// а ListMessages/DeleteCallChat станут no-op.
func NewCallChatService(repo *EntRepository, chat *ChatRepository, pub *CallEventPublisher) *CallChatService {
	return &CallChatService{repo: repo, chat: chat, pub: pub}
}

// authorizeCallChatAccess — A2-003: R/W чата только для допущенных.
// Channel-call: active participant OR IsChannelMember.
// Instant: active participant OR created_by.
func (s *CallChatService) authorizeCallChatAccess(ctx context.Context, callEnt *ent.Call, userID uuid.UUID) error {
	isParticipant, err := s.repo.IsActiveParticipant(ctx, callEnt.ID, userID)
	if err != nil {
		return fmt.Errorf("ошибка проверки участника звонка: %w", err)
	}
	if isParticipant {
		return nil
	}
	if callEnt.ChannelID != nil {
		isMember, mErr := s.repo.IsChannelMember(ctx, *callEnt.ChannelID, userID)
		if mErr != nil {
			return fmt.Errorf("ошибка проверки членства в канале: %w", mErr)
		}
		if isMember {
			return nil
		}
		return ErrCallForbidden
	}
	if callEnt.CreatedBy == userID {
		return nil
	}
	return ErrCallForbidden
}

// SendMessage отправляет сообщение в эфемерный чат звонка.
// Проверяет что звонок активен, опционально сохраняет в MongoDB (ephemeral=false)
// и публикует событие в NATS для real-time доставки.
func (s *CallChatService) SendMessage(ctx context.Context, callID, userID, userName, text string, ephemeral bool) (*CallChatMessage, error) {
	callUUID, err := uuid.Parse(callID)
	if err != nil {
		return nil, fmt.Errorf("некорректный call ID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("некорректный user ID: %w", err)
	}
	callEnt, err := s.repo.GetCall(ctx, callUUID)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}
	if err := s.authorizeCallChatAccess(ctx, callEnt, userUUID); err != nil {
		return nil, err
	}

	var msg *CallChatMessage

	if ephemeral {
		// Эфемерный режим: не сохраняем в БД, только broadcast.
		msg = &CallChatMessage{
			ID:        bson.NewObjectID(),
			CallID:    callID,
			UserID:    userID,
			UserName:  userName,
			Text:      text,
			CreatedAt: time.Now(),
		}
	} else {
		// Обычный режим: сохраняем в MongoDB.
		if s.chat == nil {
			return nil, fmt.Errorf("чат звонков не настроен")
		}
		msg, err = s.chat.SendMessage(ctx, callID, userID, userName, text)
		if err != nil {
			return nil, err
		}
	}

	s.pub.ChatMessage(callID, msg)
	return msg, nil
}

// ListMessages возвращает последние N сообщений эфемерного чата звонка.
// Если MongoDB-репо не сконфигурирован — возвращает ошибку.
// A2-003: userID обязателен — без participant/channel/created_by доступ запрещён.
func (s *CallChatService) ListMessages(ctx context.Context, callID, userID string, limit int) ([]CallChatMessage, error) {
	callUUID, err := uuid.Parse(callID)
	if err != nil {
		return nil, fmt.Errorf("некорректный call ID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("некорректный user ID: %w", err)
	}
	callEnt, err := s.repo.GetCall(ctx, callUUID)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if err := s.authorizeCallChatAccess(ctx, callEnt, userUUID); err != nil {
		return nil, err
	}
	if s.chat == nil {
		return nil, fmt.Errorf("чат звонков не настроен")
	}
	return s.chat.ListMessages(ctx, callID, limit)
}

// DeleteCallChat удаляет все сообщения чата указанного звонка (CALL-06).
// No-op если MongoDB-репо не сконфигурирован — инкапсулирует nil-check,
// который раньше жил в CallService.OnRoomFinished.
func (s *CallChatService) DeleteCallChat(ctx context.Context, callID string) {
	if s.chat == nil {
		return
	}
	if err := s.chat.DeleteCallChat(ctx, callID); err != nil {
		slog.Error("не удалось удалить чат звонка", "call_id", callID, "error", err)
	}
}
