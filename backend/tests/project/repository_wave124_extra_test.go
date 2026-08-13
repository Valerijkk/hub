package project_test

// Тесты репозитория проектов (волна 124): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

type failingProjectOutbox struct{}

func (f *failingProjectOutbox) PublishInTx(_ context.Context, _ *ent.Tx, _ string, _ any) error {
	return errors.New("outbox unavailable")
}

// Фаза 124 — repository rollback when outbox publish fails..

func TestEntRepository_Wave124_Create_RollbackOnOutboxError(t *testing.T) {
	env := newProjectRepoEnv(t)
	repo := project.NewEntRepository(env.client, &failingProjectOutbox{})

	_, err := repo.Create(env.ctx, "FailOutbox", "FOB", "", env.owner.ID)
	require.Error(t, err)

	_, err = env.repo.FindByKey(env.ctx, "FOB")
	require.Error(t, err)
}
