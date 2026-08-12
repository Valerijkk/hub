package messenger_test

// Тесты сервиса мессенджера: интеграционные и unit-сценарии testframework.

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/internal/messenger"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func startEmbeddedNATSServerForAuto(t *testing.T) string {
	t.Helper()
	opts := &server.Options{Port: -1}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)
	return s.ClientURL()
}

// TestAutoChannelService_HandleProjectCreated_AddsEntMembers проверяет слой репозитория: auto channel сервис handle project created adds ent members.
func TestAutoChannelService_HandleProjectCreated_AddsEntMembers(t *testing.T) {
	dsn := "file:autochan_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().
		SetEmail("mem-" + uuid.New().String()[:8] + "@test.com").
		SetName("Member").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Auto Channel Proj").
		SetKey(testMessengerProjectKey()).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(owner.ID).
		SetProjectRole(projectmember.ProjectRoleOwner).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(member.ID).
		SetProjectRole(projectmember.ProjectRoleDeveloper).
		Save(ctx)
	require.NoError(t, err)

	mockRepo := newMockRepository()
	// AUDIT-TM-004: CreateChannel на mock проверяет IsProjectMember (в prod — Ent).
	mockRepo.projectMembers[proj.ID] = map[uuid.UUID]bool{owner.ID: true}
	svc := messenger.ExportNewMessengerServiceTest(mockRepo)
	entRepo := messenger.NewEntRepository(client)
	auto := messenger.NewAutoChannelService(svc, entRepo)
	messenger.ExportAutoChannelServiceSetLifecycleCtx(auto, ctx)

	payload, err := json.Marshal(messenger.ProjectCreatedPayload{
		ProjectID:   proj.ID.String(),
		Name:        proj.Name,
		Key:         proj.Key,
		CreatedByID: owner.ID.String(),
	})
	require.NoError(t, err)

	messenger.ExportAutoChannelServiceHandleProjectCreated(auto, payload)

	var channelID uuid.UUID
	expectedName := "#project-" + strings.ToLower(proj.Key)
	for _, ch := range mockRepo.channels {
		if ch.Name == expectedName {
			channelID = ch.ID
			break
		}
	}
	require.NotEqual(t, uuid.Nil, channelID)

	isMember, err := mockRepo.IsMember(ctx, channelID, member.ID)
	require.NoError(t, err)
	assert.True(t, isMember, "project member must be added to auto-channel")
}

// TestAutoChannelService_StartStop_WithNATS проверяет бизнес-логику: auto channel сервис start stop with nats.
func TestAutoChannelService_StartStop_WithNATS(t *testing.T) {
	url := startEmbeddedNATSServerForAuto(t)
	nc, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(nc.Close)

	dsn := "file:autochan_nats_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mockRepo := newMockRepository()
	svc := messenger.ExportNewMessengerServiceTest(mockRepo)
	auto := messenger.NewAutoChannelService(svc, messenger.NewEntRepository(client))

	require.NoError(t, auto.Start(context.Background(), nc))
	assert.NotNil(t, messenger.ExportAutoChannelServiceSubscription(auto))

	auto.Stop()
	assert.Nil(t, messenger.ExportAutoChannelServiceSubscription(auto))
}

// TestAutoChannelService_HandleProjectCreated_InvalidPayload проверяет отклонение при auto channel сервис handle project created некорректный payload.
func TestAutoChannelService_HandleProjectCreated_InvalidPayload(t *testing.T) {
	auto := messenger.NewAutoChannelService(messenger.ExportNewMessengerServiceTest(newMockRepository()), nil)
	messenger.ExportAutoChannelServiceHandleProjectCreated(auto, []byte("not-json"))
	messenger.ExportAutoChannelServiceHandleProjectCreated(auto, []byte(`{"project_id":"bad","created_by_id":"`+uuid.New().String()+`"}`))
	messenger.ExportAutoChannelServiceHandleProjectCreated(auto, []byte(`{"project_id":"`+uuid.New().String()+`","created_by_id":"bad"}`))
}

// TestAutoChannelService_StartNilNATS проверяет бизнес-логику: auto channel сервис start nil nats.
func TestAutoChannelService_StartNilNATS(t *testing.T) {
	dsn := "file:autochan_nil_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	auto := messenger.NewAutoChannelService(messenger.ExportNewMessengerServiceEmpty(), messenger.NewEntRepository(client))
	require.NotNil(t, auto)
	require.NoError(t, auto.Start(context.Background(), nil))
	auto.Stop()
}

// TestNewCombinedRepository проверяет слой репозитория: new combined репозиторий.
func TestNewCombinedRepository(t *testing.T) {
	dsn := "file:combined_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	entRepo := messenger.NewEntRepository(client)
	combined := messenger.NewCombinedRepository(entRepo, nil)
	require.NotNil(t, combined)
}
