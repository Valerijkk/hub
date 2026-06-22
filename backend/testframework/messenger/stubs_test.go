package messenger_test

// Тесты модуля мессенджера: интеграционные и unit-сценарии testframework.

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
)

func testMessengerProjectKey() string {
	suffix := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:5]
	return "MSG" + suffix
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
