package messenger_test

// Тесты модуля мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

// ---------------------------------------------------------------------------
// Мок провайдера участников проекта
// ---------------------------------------------------------------------------

// mockProjectMemberProvider — заглушка провайдера участников проекта.
type mockProjectMemberProvider struct {
	members map[uuid.UUID][]uuid.UUID
}

func (m *mockProjectMemberProvider) listProjectMembers(projectID uuid.UUID) ([]uuid.UUID, error) {
	members, ok := m.members[projectID]
	if !ok {
		return nil, nil
	}
	return members, nil
}

// simulateProjectCreated эмулирует handleProjectCreated messenger.AutoChannelService с entRepo.
func simulateProjectCreated(
	svc *messenger.MessengerService,
	memberProvider *mockProjectMemberProvider,
	data []byte,
) {
	var payload messenger.ProjectCreatedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}

	projectID, err := uuid.Parse(payload.ProjectID)
	if err != nil {
		return
	}

	createdBy, err := uuid.Parse(payload.CreatedByID)
	if err != nil {
		return
	}

	ctx := context.Background()
	channelName := "#project-" + strings.ToLower(payload.Key)

	ch, err := svc.CreateChannel(ctx, createdBy, messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        channelName,
		ProjectID:   payload.ProjectID,
	})
	if err != nil {
		return
	}

	if memberProvider == nil {
		return
	}

	memberIDs, err := memberProvider.listProjectMembers(projectID)
	if err != nil {
		return
	}

	channelID, err := uuid.Parse(ch.ID)
	if err != nil {
		return
	}

	membersToAdd := make([]uuid.UUID, 0, len(memberIDs))
	for _, mid := range memberIDs {
		if mid != createdBy {
			membersToAdd = append(membersToAdd, mid)
		}
	}

	if len(membersToAdd) > 0 {
		_ = svc.AddMembersInternal(ctx, channelID, membersToAdd)
	}
}

// TestHandleProjectCreated_CreatesChannel проверяет сценарий: handle project created creates channel.
func TestHandleProjectCreated_CreatesChannel(t *testing.T) {
	repo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceTest(repo)

	createdBy := uuid.New()
	projectID := uuid.New()
	payload := messenger.ProjectCreatedPayload{
		ProjectID:   projectID.String(),
		Name:        "Hub",
		Key:         "HUB",
		CreatedByID: createdBy.String(),
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	// Без memberProvider событие обрабатывается, участники не добавляются.
	simulateProjectCreated(svc, nil, data)

	// Проверяем, что канал #project-hub создан.
	found := false
	var createdChannel *messenger.ChannelResponse
	for _, ch := range repo.channels {
		if ch.Name == "#project-hub" {
			found = true
			createdChannel = messenger.ExportToChannelResponse(ch)
			break
		}
	}
	require.True(t, found, "канал #project-hub должен быть создан")
	assert.Equal(t, "public", createdChannel.ChannelType)
	assert.Equal(t, projectID.String(), createdChannel.ProjectID)
}

// TestHandleProjectCreated_AddsMembersToChannel проверяет добавление участников проекта в авто-канал.
func TestHandleProjectCreated_AddsMembersToChannel(t *testing.T) {
	repo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceTest(repo)

	createdBy := uuid.New()
	member1 := uuid.New()
	member2 := uuid.New()
	projectID := uuid.New()

	memberProvider := &mockProjectMemberProvider{
		members: map[uuid.UUID][]uuid.UUID{
			projectID: {createdBy, member1, member2},
		},
	}

	payload := messenger.ProjectCreatedPayload{
		ProjectID:   projectID.String(),
		Name:        "Hub",
		Key:         "HUB",
		CreatedByID: createdBy.String(),
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	simulateProjectCreated(svc, memberProvider, data)

	// Проверяем, что все участники проекта добавлены в канал.
	var channelID uuid.UUID
	for _, ch := range repo.channels {
		if ch.Name == "#project-hub" {
			channelID = ch.ID
			break
		}
	}
	require.NotEqual(t, uuid.Nil, channelID, "канал должен быть создан")

	// Проверяем membership всех участников проекта.
	members := repo.members[channelID]
	userIDs := memberUserIDs(members)
	assert.Contains(t, userIDs, createdBy, "создатель должен быть в канале (как owner)")
	assert.Contains(t, userIDs, member1, "member1 должен быть в канале")
	assert.Contains(t, userIDs, member2, "member2 должен быть в канале")
	assert.Len(t, members, 3, "должно быть 3 участника (owner + 2 members)")
}

// memberUserIDs извлекает user ID из списка участников канала.
func memberUserIDs(members map[uuid.UUID]*ent.ChannelMember) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	return ids
}
