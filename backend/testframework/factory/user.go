package factory

// Тесты модуля factory: интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

// UserFactory создаёт пользователей в тестах (паттерн Factory, SRP).
type UserFactory struct {
	client *ent.Client
	ctx    context.Context
	t      *testing.T
}

// NewUserFactory привязывает фабрику к тестовому контексту.
func NewUserFactory(t *testing.T, client *ent.Client, ctx context.Context) *UserFactory {
	t.Helper()
	return &UserFactory{t: t, client: client, ctx: ctx}
}

// Create вставляет пользователя с уникальным email.
func (f *UserFactory) Create(namePrefix string) *ent.User {
	f.t.Helper()
	u, err := f.client.User.Create().
		SetEmail(namePrefix + "-" + uuid.New().String()[:8] + "@test.com").
		SetName(namePrefix).
		Save(f.ctx)
	require.NoError(f.t, err)
	return u
}
