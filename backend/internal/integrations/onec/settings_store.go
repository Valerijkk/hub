package onec

import (
	"context"
	"fmt"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/systemsetting"
)

// SettingsStore — хранилище конфигурации коннектора.
type SettingsStore interface {
	Get(ctx context.Context, key string) (string, error)
	Upsert(ctx context.Context, key, value, description string) error
}

// EntSettingsStore реализует SettingsStore через SystemSetting.
type EntSettingsStore struct {
	db *ent.Client
}

// NewEntSettingsStore создаёт хранилище настроек.
func NewEntSettingsStore(db *ent.Client) *EntSettingsStore {
	return &EntSettingsStore{db: db}
}

func (s *EntSettingsStore) Get(ctx context.Context, key string) (string, error) {
	row, err := s.db.SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return row.Value, nil
}

func (s *EntSettingsStore) Upsert(ctx context.Context, key, value, description string) error {
	n, err := s.db.SystemSetting.Update().
		Where(systemsetting.KeyEQ(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update setting %s: %w", key, err)
	}
	if n == 0 {
		_, err = s.db.SystemSetting.Create().
			SetKey(key).
			SetValue(value).
			SetDescription(description).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create setting %s: %w", key, err)
		}
	}
	return nil
}
