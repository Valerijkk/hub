package calls

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/callparticipant"
)

// EntRepository реализует доступ к данным звонков через Ent ORM.
type EntRepository struct {
	db *ent.Client
}

// NewEntRepository создаёт новый репозиторий звонков.
func NewEntRepository(db *ent.Client) *EntRepository {
	return &EntRepository{db: db}
}

// CreateCall создаёт запись о новом звонке в базе данных.
func (r *EntRepository) CreateCall(ctx context.Context, roomName string, channelID, createdBy uuid.UUID) (*ent.Call, error) {
	return r.db.Call.Create().
		SetRoomName(roomName).
		SetChannelID(channelID).
		SetCreatedBy(createdBy).
		Save(ctx)
}

// GetCall возвращает звонок по идентификатору.
func (r *EntRepository) GetCall(ctx context.Context, id uuid.UUID) (*ent.Call, error) {
	return r.db.Call.Get(ctx, id)
}

// GetActiveCallByChannel возвращает активный звонок в канале.
// Возвращает nil, nil если активного звонка нет.
func (r *EntRepository) GetActiveCallByChannel(ctx context.Context, channelID uuid.UUID) (*ent.Call, error) {
	c, err := r.db.Call.Query().
		Where(
			call.ChannelID(channelID),
			call.StatusEQ(call.StatusActive),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

// UpdateCallEnded обновляет звонок как завершённый.
func (r *EntRepository) UpdateCallEnded(ctx context.Context, callID uuid.UUID, endedAt time.Time, participantCount int) error {
	return r.db.Call.UpdateOneID(callID).
		SetStatus(call.StatusEnded).
		SetEndedAt(endedAt).
		SetParticipantCount(participantCount).
		Exec(ctx)
}

// UpdateCallStarted устанавливает время начала звонка.
func (r *EntRepository) UpdateCallStarted(ctx context.Context, roomName string, startedAt time.Time) error {
	_, err := r.db.Call.Update().
		Where(call.RoomName(roomName)).
		SetStartedAt(startedAt).
		Save(ctx)
	return err
}

// GetCallByRoomName возвращает звонок по имени комнаты.
func (r *EntRepository) GetCallByRoomName(ctx context.Context, roomName string) (*ent.Call, error) {
	return r.db.Call.Query().
		Where(call.RoomName(roomName)).
		Only(ctx)
}

// AddParticipant добавляет участника в звонок.
func (r *EntRepository) AddParticipant(ctx context.Context, callID, userID uuid.UUID) error {
	// Проверяем, нет ли уже активного участника (без left_at)
	exists, err := r.db.CallParticipant.Query().
		Where(
			callparticipant.CallID(callID),
			callparticipant.UserID(userID),
			callparticipant.LeftAtIsNil(),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil // Уже участник
	}

	_, err = r.db.CallParticipant.Create().
		SetCallID(callID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return err
	}

	// Инкрементируем счётчик участников
	return r.db.Call.UpdateOneID(callID).
		AddParticipantCount(1).
		Exec(ctx)
}

// RemoveParticipant помечает участника как покинувшего звонок.
func (r *EntRepository) RemoveParticipant(ctx context.Context, callID, userID uuid.UUID, leftAt time.Time) error {
	// Находим активного участника
	p, err := r.db.CallParticipant.Query().
		Where(
			callparticipant.CallID(callID),
			callparticipant.UserID(userID),
			callparticipant.LeftAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // Уже покинул
		}
		return err
	}

	// Устанавливаем время ухода
	if err := r.db.CallParticipant.UpdateOne(p).
		SetLeftAt(leftAt).
		Exec(ctx); err != nil {
		return err
	}

	// Декрементируем счётчик участников
	return r.db.Call.UpdateOneID(callID).
		AddParticipantCount(-1).
		Exec(ctx)
}

// ListParticipants возвращает всех участников звонка.
func (r *EntRepository) ListParticipants(ctx context.Context, callID uuid.UUID) ([]*ent.CallParticipant, error) {
	return r.db.CallParticipant.Query().
		Where(callparticipant.CallID(callID)).
		WithUser().
		All(ctx)
}
