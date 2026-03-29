// Package calls реализует голосовые созвоны через LiveKit: создание, присоединение, запись, планирование.
package calls

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/config"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// CallService реализует бизнес-логику модуля голосовых звонков.
// Управляет жизненным циклом комнат LiveKit, генерирует токены доступа,
// обрабатывает webhook-события, публикует NATS-уведомления и управляет эфемерным чатом.
type CallService struct {
	roomClient *lksdk.RoomServiceClient
	cfg        config.LiveKitConfig
	repo       *EntRepository
	chatRepo   *ChatRepository
	nats       *natsclient.Client
}

// NewCallService создаёт новый сервис звонков.
// Инициализирует клиент LiveKit RoomService для управления комнатами.
// chatRepo может быть nil если MongoDB не сконфигурирован.
func NewCallService(cfg config.LiveKitConfig, repo *EntRepository, nc *natsclient.Client, chatRepo *ChatRepository) *CallService {
	roomClient := lksdk.NewRoomServiceClient(cfg.URL, cfg.APIKey, cfg.APISecret)
	return &CallService{
		roomClient: roomClient,
		cfg:        cfg,
		repo:       repo,
		chatRepo:   chatRepo,
		nats:       nc,
	}
}

// GenerateToken генерирует JWT-токен для подключения пользователя к комнате LiveKit.
// Токен содержит VideoGrant с правом присоединения к указанной комнате.
func (s *CallService) GenerateToken(roomName, userID, userName string) (string, error) {
	at := auth.NewAccessToken(s.cfg.APIKey, s.cfg.APISecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.SetVideoGrant(grant).
		SetIdentity(userID).
		SetName(userName).
		SetValidFor(24 * time.Hour)

	return at.ToJWT()
}

// CreateCall создаёт новый звонок в канале.
// Создаёт комнату в LiveKit, сохраняет метаданные в БД и возвращает токен.
func (s *CallService) CreateCall(ctx context.Context, channelID, userID uuid.UUID, userName string) (*CallResponse, error) {
	roomName := "call-" + uuid.New().String()

	_, err := s.roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		MaxParticipants: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать комнату LiveKit: %w", err)
	}

	callEnt, err := s.repo.CreateCall(ctx, roomName, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить звонок: %w", err)
	}

	token, err := s.GenerateToken(roomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}

	// Добавляем создателя как участника
	if err := s.repo.AddParticipant(ctx, callEnt.ID, userID); err != nil {
		slog.Error("не удалось добавить создателя как участника", "error", err)
	}

	if channelID != uuid.Nil {
		s.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.started", channelID.String()), map[string]interface{}{
			"call_id":    callEnt.ID,
			"room_name":  roomName,
			"created_by": userID,
			"started_by": userID,
		})
	}

	return toCallResponse(callEnt, token), nil
}

// CreateInstantCall создаёт быстрый звонок без привязки к каналу.
// Генерирует invite_code для шаринга ссылки.
func (s *CallService) CreateInstantCall(ctx context.Context, userID uuid.UUID, userName, title string) (*CallResponse, error) {
	roomName := "call-" + uuid.New().String()

	_, err := s.roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		MaxParticipants: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать комнату LiveKit: %w", err)
	}

	// Генерируем короткий invite code
	inviteCode := uuid.New().String()[:8]

	if title == "" {
		title = userName + " — быстрый звонок"
	}

	callEnt, err := s.repo.CreateInstantCall(ctx, roomName, userID, inviteCode, title)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить звонок: %w", err)
	}

	token, err := s.GenerateToken(roomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}

	if err := s.repo.AddParticipant(ctx, callEnt.ID, userID); err != nil {
		slog.Error("не удалось добавить создателя как участника", "error", err)
	}

	resp := toCallResponse(callEnt, token)
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

	token, err := s.GenerateToken(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}

	resp := toCallResponse(callEnt, token)
	resp.InviteCode = code
	return resp, nil
}

// JoinCall присоединяет пользователя к существующему звонку.
// Проверяет что звонок активен, добавляет участника и возвращает токен.
func (s *CallService) JoinCall(ctx context.Context, callID, userID uuid.UUID, userName string) (*CallResponse, error) {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}

	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}

	if err := s.repo.AddParticipant(ctx, callID, userID); err != nil {
		return nil, fmt.Errorf("не удалось добавить участника: %w", err)
	}

	token, err := s.GenerateToken(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}

	s.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.joined", channelIDStr(callEnt.ChannelID)), map[string]interface{}{
		"call_id":   callID,
		"user_id":   userID,
		"user_name": userName,
	})

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

	s.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.left", channelIDStr(callEnt.ChannelID)), map[string]interface{}{
		"call_id": callID,
		"user_id": userID,
	})

	return nil
}

// EndCall завершает звонок: удаляет комнату в LiveKit и обновляет статус.
func (s *CallService) EndCall(ctx context.Context, callID uuid.UUID) error {
	callEnt, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("звонок не найден: %w", err)
	}

	// Удаляем комнату в LiveKit
	_, delErr := s.roomClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: callEnt.RoomName,
	})
	if delErr != nil {
		slog.Error("не удалось удалить комнату LiveKit", "room", callEnt.RoomName, "error", delErr)
	}

	// Обновляем статус в БД
	if err := s.repo.UpdateCallEnded(ctx, callID, time.Now(), callEnt.ParticipantCount); err != nil {
		return fmt.Errorf("не удалось обновить статус звонка: %w", err)
	}

	s.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.ended", channelIDStr(callEnt.ChannelID)), map[string]interface{}{
		"call_id":           callID,
		"participant_count": callEnt.ParticipantCount,
	})

	return nil
}

// GetActiveCallByChannel возвращает активный звонок в канале.
// Возвращает nil если активного звонка нет.
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

// SetRaiseHand обновляет метаданные участника в LiveKit для функции "поднять руку".
func (s *CallService) SetRaiseHand(ctx context.Context, roomName, identity string, raised bool) error {
	meta := fmt.Sprintf(`{"hand_raised":%v,"raised_at":"%s"}`, raised, time.Now().Format(time.RFC3339))
	_, err := s.roomClient.UpdateParticipant(ctx, &livekit.UpdateParticipantRequest{
		Room:     roomName,
		Identity: identity,
		Metadata: meta,
	})
	return err
}

// OnRoomStarted обрабатывает webhook room_started от LiveKit.
// Устанавливает время начала звонка.
func (s *CallService) OnRoomStarted(ctx context.Context, roomName string) error {
	return s.repo.UpdateCallStarted(ctx, roomName, time.Now())
}

// OnRoomFinished обрабатывает webhook room_finished от LiveKit.
// Завершает звонок, обновляет статус и удаляет эфемерный чат (CALL-06).
func (s *CallService) OnRoomFinished(ctx context.Context, roomName string) error {
	callEnt, err := s.repo.GetCallByRoomName(ctx, roomName)
	if err != nil {
		return fmt.Errorf("звонок не найден по имени комнаты: %w", err)
	}

	if err := s.repo.UpdateCallEnded(ctx, callEnt.ID, time.Now(), callEnt.ParticipantCount); err != nil {
		return err
	}

	// Удаляем эфемерный чат звонка (CALL-06)
	if s.chatRepo != nil {
		if delErr := s.chatRepo.DeleteCallChat(ctx, callEnt.ID.String()); delErr != nil {
			slog.Error("не удалось удалить чат звонка", "call_id", callEnt.ID, "error", delErr)
		}
	}

	s.nats.PublishJSON(fmt.Sprintf("calls.channel.%s.ended", channelIDStr(callEnt.ChannelID)), map[string]interface{}{
		"call_id":           callEnt.ID,
		"participant_count": callEnt.ParticipantCount,
	})

	return nil
}

// SendChatMessage отправляет сообщение в эфемерный чат звонка.
// Проверяет что звонок активен, сохраняет сообщение в MongoDB и публикует событие в NATS.
func (s *CallService) SendChatMessage(ctx context.Context, callID string, userID, userName, text string) (*CallChatMessage, error) {
	// Проверяем что звонок существует и активен
	callUUID, err := uuid.Parse(callID)
	if err != nil {
		return nil, fmt.Errorf("некорректный call ID: %w", err)
	}
	callEnt, err := s.repo.GetCall(ctx, callUUID)
	if err != nil {
		return nil, fmt.Errorf("звонок не найден: %w", err)
	}
	if callEnt.Status != "active" {
		return nil, fmt.Errorf("звонок уже завершён")
	}

	if s.chatRepo == nil {
		return nil, fmt.Errorf("чат звонков не настроен")
	}

	msg, err := s.chatRepo.SendMessage(ctx, callID, userID, userName, text)
	if err != nil {
		return nil, err
	}

	// Публикуем событие для real-time доставки через NATS -> WS bridge
	s.nats.PublishJSON("calls.chat."+callID, map[string]interface{}{
		"id":         msg.ID.Hex(),
		"call_id":    msg.CallID,
		"user_id":    msg.UserID,
		"user_name":  msg.UserName,
		"text":       msg.Text,
		"created_at": msg.CreatedAt,
	})

	return msg, nil
}

// ListChatMessages возвращает последние N сообщений эфемерного чата звонка.
func (s *CallService) ListChatMessages(ctx context.Context, callID string, limit int) ([]CallChatMessage, error) {
	if s.chatRepo == nil {
		return nil, fmt.Errorf("чат звонков не настроен")
	}
	return s.chatRepo.ListMessages(ctx, callID, limit)
}

// OnParticipantJoined обрабатывает webhook participant_joined от LiveKit.
func (s *CallService) OnParticipantJoined(ctx context.Context, roomName, identity string) error {
	callEnt, err := s.repo.GetCallByRoomName(ctx, roomName)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(identity)
	if err != nil {
		slog.Error("некорректный identity участника", "identity", identity)
		return nil
	}

	return s.repo.AddParticipant(ctx, callEnt.ID, userID)
}

// OnParticipantLeft обрабатывает webhook participant_left от LiveKit.
func (s *CallService) OnParticipantLeft(ctx context.Context, roomName, identity string) error {
	callEnt, err := s.repo.GetCallByRoomName(ctx, roomName)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(identity)
	if err != nil {
		slog.Error("некорректный identity участника", "identity", identity)
		return nil
	}

	return s.repo.RemoveParticipant(ctx, callEnt.ID, userID, time.Now())
}

// channelIDStr безопасно возвращает строку из *uuid.UUID (nil → "instant").
func channelIDStr(id *uuid.UUID) string {
	if id == nil {
		return "instant"
	}
	return id.String()
}

// toCallResponse конвертирует Ent Call в DTO ответа.
func toCallResponse(c *ent.Call, token string) *CallResponse {
	resp := &CallResponse{
		ID:               c.ID,
		RoomName:         c.RoomName,
		Status:           string(c.Status),
		CreatedBy:        c.CreatedBy,
		ParticipantCount: c.ParticipantCount,
		Token:            token,
	}
	if c.ChannelID != nil {
		resp.ChannelID = c.ChannelID
	}
	if c.InviteCode != "" {
		resp.InviteCode = c.InviteCode
	}
	if c.Title != "" {
		resp.Title = c.Title
	}
	if c.StartedAt != nil {
		resp.StartedAt = c.StartedAt
	}
	if c.EndedAt != nil {
		resp.EndedAt = c.EndedAt
	}
	return resp
}
