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
)

// provisionRoom создаёт LiveKit-комнату с уникальным именем "call-<uuid>".
// Package-level helper для orchestrator-flow (CreateCall / CreateInstantCall).
func provisionRoom(ctx context.Context, room *CallRoomService) (string, error) {
	roomName := "call-" + uuid.New().String()
	if err := room.CreateRoom(ctx, roomName, defaultMaxCallParticipants); err != nil {
		return "", err
	}
	return roomName, nil
}

// finalizeCreate выдаёт токен создателю и регистрирует его первым участником.
// Package-level helper для orchestrator-flow (CreateCall / CreateInstantCall).
// A2W3-5-002: ошибка AddParticipant → rollback LiveKit room + hard-delete call, не 201+JWT без participant.
func finalizeCreate(ctx context.Context, room *CallRoomService, repo *EntRepository, callEnt *ent.Call, userID uuid.UUID, userName string) (*CallResponse, error) {
	token, err := room.GenerateToken(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}
	if err := repo.AddParticipant(ctx, callEnt.ID, userID); err != nil {
		rollbackCreate(ctx, room, repo, callEnt)
		return nil, fmt.Errorf("не удалось добавить создателя как участника: %w", err)
	}
	return toCallResponse(callEnt, token), nil
}

// rollbackCreate удаляет LiveKit-комнату и запись звонка после провала finalizeCreate.
func rollbackCreate(ctx context.Context, room *CallRoomService, repo *EntRepository, callEnt *ent.Call) {
	if delErr := room.DeleteRoom(ctx, callEnt.RoomName); delErr != nil {
		slog.Error("finalizeCreate rollback: не удалось удалить комнату LiveKit",
			"room", callEnt.RoomName, "error", delErr)
	}
	if delErr := repo.HardDeleteCall(ctx, callEnt.ID); delErr != nil {
		slog.Error("finalizeCreate rollback: не удалось удалить звонок",
			"call_id", callEnt.ID, "error", delErr)
	}
}

// resolveParticipant находит звонок по комнате и парсит identity участника webhook'а.
// (nil, uuid.Nil, nil) при невалидном identity — webhook логируется и пропускается.
func resolveParticipant(ctx context.Context, repo *EntRepository, roomName, identity string) (*ent.Call, uuid.UUID, error) {
	callEnt, err := repo.GetCallByRoomName(ctx, roomName)
	if err != nil {
		return nil, uuid.Nil, err
	}
	userID, err := uuid.Parse(identity)
	if err != nil {
		slog.Error("некорректный identity участника", "identity", identity)
		return nil, uuid.Nil, nil
	}
	return callEnt, userID, nil
}

// toCallResponse конвертирует Ent Call + LiveKit-токен в DTO ответа.
// Живёт рядом с CallRoomService, потому что LiveKit-комната — источник
// и room_name, и токена, образующих транспортную часть CallResponse.
func toCallResponse(c *ent.Call, token string) *CallResponse {
	return &CallResponse{
		ID:               c.ID,
		RoomName:         c.RoomName,
		Status:           string(c.Status),
		CreatedBy:        c.CreatedBy,
		ParticipantCount: c.ParticipantCount,
		Token:            token,
		ChannelID:        c.ChannelID,
		InviteCode:       c.InviteCode,
		Title:            c.Title,
		StartedAt:        c.StartedAt,
		EndedAt:          c.EndedAt,
	}
}

// CallRoomService инкапсулирует все вызовы LiveKit RoomServiceClient.
// Отделён от CallService, чтобы оркестратор не знал деталей LiveKit SDK
// и был тестируем без сетевой зависимости (SC-14-2).
type CallRoomService struct {
	client *lksdk.RoomServiceClient
	cfg    config.LiveKitConfig
}

// NewCallRoomService создаёт wrapper над LiveKit RoomServiceClient.
func NewCallRoomService(cfg config.LiveKitConfig) *CallRoomService {
	return &CallRoomService{
		client: lksdk.NewRoomServiceClient(cfg.URL, cfg.APIKey, cfg.APISecret),
		cfg:    cfg,
	}
}

// CreateRoom создаёт комнату в LiveKit с заданным потолком участников.
func (r *CallRoomService) CreateRoom(ctx context.Context, name string, maxParticipants uint32) error {
	_, err := r.client.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            name,
		MaxParticipants: maxParticipants,
	})
	if err != nil {
		return fmt.Errorf("не удалось создать комнату LiveKit: %w", err)
	}
	return nil
}

// DeleteRoom удаляет комнату в LiveKit. Безопасен к повторным вызовам
// — LiveKit вернёт ошибку, которую вызывающий обработает как warning.
func (r *CallRoomService) DeleteRoom(ctx context.Context, name string) error {
	_, err := r.client.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: name,
	})
	return err
}

// UpdateParticipantMetadata обновляет метаданные участника в LiveKit-комнате.
// Используется для функции "поднять руку" и аналогичных эфемерных состояний.
func (r *CallRoomService) UpdateParticipantMetadata(ctx context.Context, room, identity, metaJSON string) error {
	_, err := r.client.UpdateParticipant(ctx, &livekit.UpdateParticipantRequest{
		Room:     room,
		Identity: identity,
		Metadata: metaJSON,
	})
	return err
}

// GenerateToken создаёт JWT-токен для подключения пользователя к комнате LiveKit.
// Токен содержит VideoGrant с правом присоединения к указанной комнате
// и валиден 24 часа.
func (r *CallRoomService) GenerateToken(roomName, userID, userName string) (string, error) {
	at := auth.NewAccessToken(r.cfg.APIKey, r.cfg.APISecret)
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

// GenerateGuestToken выдаёт JWT гостю. Identity — полный UUID (W3-A5-008 entropy).
func (r *CallRoomService) GenerateGuestToken(roomName, guestName string) (identity, token string, err error) {
	identity = "guest-" + uuid.New().String()
	token, err = r.GenerateToken(roomName, identity, guestName)
	return identity, token, err
}
