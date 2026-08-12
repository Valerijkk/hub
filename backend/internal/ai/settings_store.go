package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/systemsetting"
)

const settingBotSystemPrompt = "ai.bot_system_prompt"

// SettingsStore — key/value админ-настройки (SystemSetting).
// Значения non-secret (system prompt и т.п.) — не для API keys / credentials.
type SettingsStore interface {
	Get(ctx context.Context, key string) (string, error)
	Upsert(ctx context.Context, key, value, description string) error
}

// EntSettingsStore — SettingsStore через Ent.
type EntSettingsStore struct {
	db *ent.Client
}

// NewEntSettingsStore — хранилище AI-настроек.
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

func isSettingsBusy(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database table is locked") ||
		strings.Contains(msg, "sqlite_busy")
}

// Upsert создаёт или обновляет non-secret AI-настройку.
// A3W5-3-003: Create→conflict→Update (не check-then-act race); retry на SQLite busy / stale version.
// A3W5-3-006: Update + VersionEQ + AddVersion(1).
func (s *EntSettingsStore) Upsert(ctx context.Context, key, value, description string) error {
	const maxAttempts = 12
	var last error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 5 * time.Millisecond):
			}
		}
		err := s.upsertOnce(ctx, key, value, description)
		if err == nil {
			return nil
		}
		last = err
		if errors.Is(err, ErrVersionConflict) || isSettingsBusy(err) || ent.IsConstraintError(err) {
			continue
		}
		// wrapped busy / constraint
		if isSettingsBusy(err) {
			continue
		}
		msg := err.Error()
		if strings.Contains(msg, "unique") || strings.Contains(msg, "constraint") {
			continue
		}
		return err
	}
	return last
}

func (s *EntSettingsStore) upsertOnce(ctx context.Context, key, value, description string) error {
	row, err := s.db.SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("load ai setting %s: %w", key, err)
	}
	if err == nil {
		n, uerr := s.db.SystemSetting.Update().
			Where(
				systemsetting.IDEQ(row.ID),
				systemsetting.VersionEQ(row.Version),
			).
			SetValue(value).
			SetDescription(description).
			AddVersion(1).
			Save(ctx)
		if uerr != nil {
			return fmt.Errorf("update ai setting %s: %w", key, uerr)
		}
		if n == 0 {
			return ErrVersionConflict
		}
		return nil
	}

	_, err = s.db.SystemSetting.Create().
		SetKey(key).
		SetValue(value).
		SetDescription(description).
		Save(ctx)
	if err == nil {
		return nil
	}
	// SQLite busy → retry на уровне Upsert; unique conflict → Update ниже.
	if isSettingsBusy(err) {
		return err
	}
	if !ent.IsConstraintError(err) {
		return fmt.Errorf("create ai setting %s: %w", key, err)
	}
	// A3W5-3-003: concurrent Create → Update по key.
	row, rerr := s.db.SystemSetting.Query().Where(systemsetting.KeyEQ(key)).Only(ctx)
	if rerr != nil {
		if isSettingsBusy(rerr) || ent.IsNotFound(rerr) {
			return rerr
		}
		return fmt.Errorf("create ai setting %s (after conflict): %w", key, err)
	}
	n, uerr := s.db.SystemSetting.Update().
		Where(
			systemsetting.IDEQ(row.ID),
			systemsetting.VersionEQ(row.Version),
		).
		SetValue(value).
		SetDescription(description).
		AddVersion(1).
		Save(ctx)
	if uerr != nil {
		return fmt.Errorf("update ai setting %s (after conflict): %w", key, uerr)
	}
	if n == 0 {
		return ErrVersionConflict
	}
	return nil
}
