package onec_test

import (
	"context"

	"github.com/google/uuid"
)

type memStore struct {
	data map[string]string
}

func (m *memStore) Get(_ context.Context, key string) (string, error) {
	return m.data[key], nil
}

func (m *memStore) Upsert(_ context.Context, key, value, _ string) error {
	m.data[key] = value
	return nil
}

type stubUsers struct{}

func (stubUsers) GetUserEmailPhone(_ context.Context, _ uuid.UUID) (string, string, string, error) {
	return "dev@hub.local", "+7999", "Dev", nil
}
