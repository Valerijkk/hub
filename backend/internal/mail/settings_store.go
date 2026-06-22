package mail

import (
	"context"
	"fmt"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/systemsetting"
)

// EntSettingsStore — SystemSetting через Ent (Phase 15.2 skeleton).
type EntSettingsStore struct {
	db *ent.Client
}

// NewEntSettingsStore создаёт store.
func NewEntSettingsStore(db *ent.Client) *EntSettingsStore {
	return &EntSettingsStore{db: db}
}

// Get возвращает значение настройки.
func (s *EntSettingsStore) Get(ctx context.Context, key string) (string, error) {
	setting, err := s.db.SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return setting.Value, nil
}

// Upsert создаёт или обновляет настройку.
func (s *EntSettingsStore) Upsert(ctx context.Context, key, value, description string) error {
	n, err := s.db.SystemSetting.Update().
		Where(systemsetting.KeyEQ(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update mail setting %s: %w", key, err)
	}
	if n == 0 {
		_, err = s.db.SystemSetting.Create().
			SetKey(key).
			SetValue(value).
			SetDescription(description).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create mail setting %s: %w", key, err)
		}
	}
	return nil
}
