package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/ent/systemsetting"
)

// RetentionStore — слой доступа к Postgres для RetentionService (A2W6-A4-004).
// *ent.Client живёт только здесь; сервис зависит от интерфейса.
type RetentionStore interface {
	GetSystemSetting(ctx context.Context, key string) (string, error)
	ListExpiredRecordings(ctx context.Context, cutoff time.Time) ([]*ent.Recording, error)
	DeleteRecordingsByIDs(ctx context.Context, ids []uuid.UUID) (int, error)
	DeleteRecordingsBefore(ctx context.Context, cutoff time.Time) (int, error)
}

type entRetentionStore struct {
	client *ent.Client
}

// NewEntRetentionStore создаёт RetentionStore поверх Ent.
func NewEntRetentionStore(client *ent.Client) RetentionStore {
	return &entRetentionStore{client: client}
}

func (s *entRetentionStore) GetSystemSetting(ctx context.Context, key string) (string, error) {
	setting, err := s.client.SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("get retention setting %s: %w", key, err)
	}
	return setting.Value, nil
}

func (s *entRetentionStore) ListExpiredRecordings(ctx context.Context, cutoff time.Time) ([]*ent.Recording, error) {
	return s.client.Recording.Query().
		Where(recording.CreatedAtLT(cutoff)).
		All(ctx)
}

func (s *entRetentionStore) DeleteRecordingsByIDs(ctx context.Context, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return s.client.Recording.Delete().
		Where(recording.IDIn(ids...)).
		Exec(ctx)
}

func (s *entRetentionStore) DeleteRecordingsBefore(ctx context.Context, cutoff time.Time) (int, error) {
	return s.client.Recording.Delete().
		Where(recording.CreatedAtLT(cutoff)).
		Exec(ctx)
}
