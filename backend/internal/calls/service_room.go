package calls

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// provisionRoom creates a provider room with a unique call identifier.
func provisionRoom(ctx context.Context, provider ConferenceProvider) (string, error) {
	roomName := "call-" + uuid.New().String()
	if err := provider.Provision(ctx, roomName, defaultMaxCallParticipants); err != nil {
		return "", err
	}
	return roomName, nil
}

// finalizeCreate issues the creator credential and persists the first participant.
func finalizeCreate(ctx context.Context, provider ConferenceProvider, repo *EntRepository, callEnt *ent.Call, userID uuid.UUID, userName string) (*CallResponse, error) {
	credential, err := provider.IssueJoinCredential(callEnt.RoomName, userID.String(), userName)
	if err != nil {
		return nil, fmt.Errorf("не удалось сгенерировать токен: %w", err)
	}
	if err := repo.AddParticipant(ctx, callEnt.ID, userID); err != nil {
		rollbackCreate(ctx, provider, repo, callEnt)
		return nil, fmt.Errorf("не удалось добавить создателя как участника: %w", err)
	}
	return toCallResponse(callEnt, credential.Token), nil
}

// rollbackCreate removes the provider room and call record after finalizeCreate fails.
func rollbackCreate(ctx context.Context, provider ConferenceProvider, repo *EntRepository, callEnt *ent.Call) {
	if delErr := provider.Delete(ctx, callEnt.RoomName); delErr != nil {
		slog.Error("finalizeCreate rollback: не удалось удалить комнату", "room", callEnt.RoomName, "error", delErr)
	}
	if delErr := repo.HardDeleteCall(ctx, callEnt.ID); delErr != nil {
		slog.Error("finalizeCreate rollback: не удалось удалить звонок", "call_id", callEnt.ID, "error", delErr)
	}
}

// resolveParticipant finds the call by room and parses a webhook participant identity.
// It returns (nil, uuid.Nil, nil) for an invalid identity so callers can skip it.
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

// toCallResponse converts an Ent Call and compatibility credential token to a DTO response.
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
