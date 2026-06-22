package messenger

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
)

// HookTestRepo — in-memory репозиторий мессенджера для интеграционных тестов @hub-ai hook.
type HookTestRepo struct {
	r *mockRepository
}

// NewHookTestRepo создаёт репозиторий на базе mock store unit-тестов мессенджера.
func NewHookTestRepo() *HookTestRepo {
	return &HookTestRepo{r: newMockRepository()}
}

// Repository возвращает интерфейс Repository для NewMessengerService.
func (h *HookTestRepo) Repository() Repository {
	return h.r
}

// SetupDM создаёт DM-канал между userID и botID.
func (h *HookTestRepo) SetupDM(userID, botID uuid.UUID) uuid.UUID {
	return setupChannelWithMembers(h.r, string(channel.ChannelTypeDm), false, userID, botID)
}

// SetupPublicChannel создаёт публичный канал с указанными участниками.
func (h *HookTestRepo) SetupPublicChannel(ownerID uuid.UUID, memberIDs ...uuid.UUID) uuid.UUID {
	return setupChannelWithMembers(h.r, string(channel.ChannelTypePublic), false, ownerID, memberIDs...)
}

// InsertMessage сохраняет сообщение user/bot и возвращает его ID.
func (h *HookTestRepo) InsertMessage(channelID, senderID uuid.UUID, text, threadID string) (string, error) {
	doc := &MessageDocument{
		ChannelID:   channelID,
		SenderID:    senderID,
		ContentText: text,
		ThreadID:    threadID,
		CreatedAt:   time.Now().UTC(),
	}
	id, err := h.r.InsertMessage(context.Background(), doc)
	if err != nil {
		return "", fmt.Errorf("hook fixture insert message: %w", err)
	}
	return id, nil
}

// Channel возвращает сущность канала.
func (h *HookTestRepo) Channel(channelID uuid.UUID) *ent.Channel {
	return h.r.channels[channelID]
}

func setupChannelWithMembers(repo *mockRepository, channelType string, isReadonly bool, ownerID uuid.UUID, memberIDs ...uuid.UUID) uuid.UUID {
	ch := &ent.Channel{
		ID:          uuid.New(),
		ChannelType: channel.ChannelType(channelType),
		IsReadonly:  isReadonly,
		CreatedBy:   ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.channels[ch.ID] = ch
	repo.members[ch.ID] = make(map[uuid.UUID]*ent.ChannelMember)
	repo.members[ch.ID][ownerID] = &ent.ChannelMember{
		ID: uuid.New(), ChannelID: ch.ID, UserID: ownerID,
		Role: channelmember.RoleOwner, JoinedAt: time.Now(),
	}
	for _, uid := range memberIDs {
		repo.members[ch.ID][uid] = &ent.ChannelMember{
			ID: uuid.New(), ChannelID: ch.ID, UserID: uid,
			Role: channelmember.RoleMember, JoinedAt: time.Now(),
		}
	}
	return ch.ID
}
