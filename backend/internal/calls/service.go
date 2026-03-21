package calls

import (
	"context"
	"encoding/json"
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
// обрабатывает webhook-события и публикует NATS-уведомления.
type CallService struct {
	roomClient *lksdk.RoomServiceClient
	cfg        config.LiveKitConfig
	repo       *EntRepository
	nats       *natsclient.Client
}

// NewCallService создаёт новый сервис звонков.
// Инициализирует клиент LiveKit RoomService для управления комнатами.
func NewCallService(cfg config.LiveKitConfig, repo *EntRepository, nc *natsclient.Client) *CallService {
	roomClient := lksdk.NewRoomServiceClient(cfg.URL, cfg.APIKey, cfg.APISecret)
	return &CallService{
		roomClient: roomClient,
		cfg:        cfg,
		repo:       repo,
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

	s.publishEvent(fmt.Sprintf("calls.channel.%s.started", channelID.String()), map[string]interface{}{
		"call_id":    callEnt.ID,
		"room_name":  roomName,
		"created_by": userID,
	})

	return toCallResponse(callEnt, token), nil
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

	s.publishEvent(fmt.Sprintf("calls.channel.%s.joined", callEnt.ChannelID.String()), map[string]interface{}{
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

	s.publishEvent(fmt.Sprintf("calls.channel.%s.left", callEnt.ChannelID.String()), map[string]interface{}{
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

	s.publishEvent(fmt.Sprintf("calls.channel.%s.ended", callEnt.ChannelID.String()), map[string]interface{}{
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
// Завершает звонок и обновляет статус.
func (s *CallService) OnRoomFinished(ctx context.Context, roomName string) error {
	callEnt, err := s.repo.GetCallByRoomName(ctx, roomName)
	if err != nil {
		return fmt.Errorf("звонок не найден по имени комнаты: %w", err)
	}

	if err := s.repo.UpdateCallEnded(ctx, callEnt.ID, time.Now(), callEnt.ParticipantCount); err != nil {
		return err
	}

	s.publishEvent(fmt.Sprintf("calls.channel.%s.ended", callEnt.ChannelID.String()), map[string]interface{}{
		"call_id":           callEnt.ID,
		"participant_count": callEnt.ParticipantCount,
	})

	return nil
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

// publishEvent публикует событие в NATS (fire-and-forget).
// Безопасен при nil NATS-клиенте (пропускает публикацию с логированием).
func (s *CallService) publishEvent(subject string, payload interface{}) {
	if s.nats == nil {
		slog.Debug("NATS не подключён, пропускаем публикацию", "subject", subject)
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать событие", "subject", subject, "error", err)
		return
	}
	if err := s.nats.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать событие", "subject", subject, "error", err)
	}
}

// toCallResponse конвертирует Ent Call в DTO ответа.
func toCallResponse(c *ent.Call, token string) *CallResponse {
	resp := &CallResponse{
		ID:               c.ID,
		RoomName:         c.RoomName,
		ChannelID:        c.ChannelID,
		Status:           string(c.Status),
		CreatedBy:        c.CreatedBy,
		ParticipantCount: c.ParticipantCount,
		Token:            token,
	}
	if c.StartedAt != nil {
		resp.StartedAt = c.StartedAt
	}
	if c.EndedAt != nil {
		resp.EndedAt = c.EndedAt
	}
	return resp
}
