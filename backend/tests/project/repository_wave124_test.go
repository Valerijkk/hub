package project_test

// Тесты репозитория проектов (волна 124): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

type projectOutboxStub struct {
	events []string
}

func (s *projectOutboxStub) PublishInTx(_ context.Context, _ *ent.Tx, eventType string, _ any) error {
	s.events = append(s.events, eventType)
	return nil
}

// Фаза 124 — repository.Create org + outbox branches..

func TestEntRepository_Wave124_Create_WithOrganization(t *testing.T) {
	env := newProjectRepoEnv(t)

	org, err := env.client.Organization.Create().
		SetName("Hub Org").
		SetSlug("hub-org").
		SetInviteToken("tok-" + uuid.New().String()[:8]).
		Save(env.ctx)
	require.NoError(t, err)

	owner, err := env.client.User.Create().
		SetEmail("org-owner-" + uuid.New().String()[:8] + "@test.com").
		SetName("Org Owner").
		SetOrganization(org).
		Save(env.ctx)
	require.NoError(t, err)

	repo := project.NewEntRepository(env.client, nil)
	created, err := repo.Create(env.ctx, "OrgProject", "ORP", "with org", owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "OrgProject", created.Name)
}

// TestEntRepository_Wave124_Create_PublishesOutbox проверяет слой репозитория: ent репозиторий wave 124 создание publishes outbox.
func TestEntRepository_Wave124_Create_PublishesOutbox(t *testing.T) {
	env := newProjectRepoEnv(t)
	pub := &projectOutboxStub{}
	repo := project.NewEntRepository(env.client, pub)

	created, err := repo.Create(env.ctx, "OutboxProj", "OBX", "", env.owner.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Contains(t, pub.events, "project.created")
}
