package calls

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/ent/callparticipant"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/recording"
	"github.com/valerii/hub/backend/ent/scheduledcall"
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

// CreateInstantCall создаёт звонок без привязки к каналу с invite-кодом.
func (r *EntRepository) CreateInstantCall(ctx context.Context, roomName string, createdBy uuid.UUID, inviteCode, title string) (*ent.Call, error) {
	return r.db.Call.Create().
		SetRoomName(roomName).
		SetCreatedBy(createdBy).
		SetInviteCode(inviteCode).
		SetTitle(title).
		Save(ctx)
}

// GetCallByInviteCode возвращает активный звонок по invite-коду.
func (r *EntRepository) GetCallByInviteCode(ctx context.Context, code string) (*ent.Call, error) {
	return r.db.Call.Query().
		Where(
			call.InviteCode(code),
			call.StatusEQ(call.StatusActive),
		).
		Only(ctx)
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
// Если участник уже был (left_at != nil), обнуляет left_at для повторного входа.
func (r *EntRepository) AddParticipant(ctx context.Context, callID, userID uuid.UUID) error {
	// Проверяем, нет ли уже активного участника (без left_at)
	activeExists, err := r.db.CallParticipant.Query().
		Where(
			callparticipant.CallID(callID),
			callparticipant.UserID(userID),
			callparticipant.LeftAtIsNil(),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if activeExists {
		return nil // Уже активный участник
	}

	// Проверяем, есть ли бывший участник (с left_at) — повторный вход
	former, err := r.db.CallParticipant.Query().
		Where(
			callparticipant.CallID(callID),
			callparticipant.UserID(userID),
		).
		Only(ctx)
	if err == nil {
		// Обнуляем left_at для повторного входа
		if err := r.db.CallParticipant.UpdateOne(former).ClearLeftAt().Exec(ctx); err != nil {
			return err
		}
		return r.db.Call.UpdateOneID(callID).
			AddParticipantCount(1).
			Exec(ctx)
	}
	if !ent.IsNotFound(err) {
		return err
	}

	// Новый участник
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

// IsActiveParticipant — true, если user сейчас в звонке (left_at IS NULL).
// A2-003 / A2-006 / A2-007: chat, RaiseHand, WS call:{id}.
func (r *EntRepository) IsActiveParticipant(ctx context.Context, callID, userID uuid.UUID) (bool, error) {
	return r.db.CallParticipant.Query().
		Where(
			callparticipant.CallID(callID),
			callparticipant.UserID(userID),
			callparticipant.LeftAtIsNil(),
		).
		Exist(ctx)
}


// IsCallParticipantOrChannelMember — A2-007: доступ к WS-комнате call:{id}.
// Канальный звонок: член канала ИЛИ активный участник.
// Instant (ChannelID == nil): только активный участник (раньше всегда deny).
func (r *EntRepository) IsCallParticipantOrChannelMember(ctx context.Context, callID, userID uuid.UUID) (bool, error) {
	call, err := r.GetCall(ctx, callID)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if call.ChannelID != nil {
		isMember, mErr := r.IsChannelMember(ctx, *call.ChannelID, userID)
		if mErr != nil {
			return false, mErr
		}
		if isMember {
			return true, nil
		}
	}
	return r.IsActiveParticipant(ctx, callID, userID)
}

// IsChannelMember проверяет, состоит ли пользователь в активном канале мессенджера.
// HI-03: используется для авторизации JoinCall/EndCall — у звонка есть
// channel_id (для канальных звонков), членство в этом канале — необходимое
// условие для участия/завершения. Soft-deleted каналы не считаются (A8-005).
func (r *EntRepository) IsChannelMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.db.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
			channelmember.HasChannelWith(channel.DeletedAtIsNil()),
		).
		Exist(ctx)
}

// GetChannelMemberRole возвращает роль пользователя в канале ("", если не состоит).
// HI-03: EndCall разрешён владельцу/админу канала; этот метод инкапсулирует
// чтение роли без раскрытия Ent наружу.
func (r *EntRepository) GetChannelMemberRole(ctx context.Context, channelID, userID uuid.UUID) (string, error) {
	m, err := r.db.ChannelMember.Query().
		Where(
			channelmember.ChannelID(channelID),
			channelmember.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return string(m.Role), nil
}

// ---------------------------------------------------------------------------
// Recording методы (CALL-04, CALL-05)
// ---------------------------------------------------------------------------

// GetProjectIDByChannel возвращает project_id канала.
// Необходим для привязки записей к проекту.
func (r *EntRepository) GetProjectIDByChannel(ctx context.Context, channelID uuid.UUID) (uuid.UUID, error) {
	ch, err := r.db.Channel.Query().
		Where(
			channel.ID(channelID),
			channel.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	if ch.ProjectID == nil {
		return uuid.Nil, nil
	}
	return *ch.ProjectID, nil
}

// CreateRecording создаёт запись о новой записи звонка в БД.
func (r *EntRepository) CreateRecording(ctx context.Context, egressID string, callID, projectID, userID uuid.UUID, filePath string) (*ent.Recording, error) {
	return r.db.Recording.Create().
		SetEgressID(egressID).
		SetCallID(callID).
		SetProjectID(projectID).
		SetStartedBy(userID).
		SetFilePath(filePath).
		Save(ctx)
}

// GetRecording возвращает запись по ID.
func (r *EntRepository) GetRecording(ctx context.Context, id uuid.UUID) (*ent.Recording, error) {
	return r.db.Recording.Get(ctx, id)
}

// GetRecordingByEgressID возвращает запись по ID egress-процесса LiveKit.
func (r *EntRepository) GetRecordingByEgressID(ctx context.Context, egressID string) (*ent.Recording, error) {
	return r.db.Recording.Query().
		Where(recording.EgressID(egressID)).
		Only(ctx)
}

// CountActiveRecordings подсчитывает активные записи (status=recording) для звонка.
func (r *EntRepository) CountActiveRecordings(ctx context.Context, callID uuid.UUID) (int, error) {
	return r.db.Recording.Query().
		Where(
			recording.CallID(callID),
			recording.StatusEQ(recording.StatusRecording),
		).
		Count(ctx)
}

// UpdateRecordingStatus обновляет статус записи.
func (r *EntRepository) UpdateRecordingStatus(ctx context.Context, id uuid.UUID, status recording.Status) error {
	return r.db.Recording.UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
}

// UpdateRecordingCompleted обновляет запись как завершённую (статус ready, размер, длительность).
func (r *EntRepository) UpdateRecordingCompleted(ctx context.Context, id uuid.UUID, fileSize int64, duration int, endedAt time.Time) error {
	return r.db.Recording.UpdateOneID(id).
		SetStatus(recording.StatusReady).
		SetFileSize(fileSize).
		SetDurationSeconds(duration).
		SetEndedAt(endedAt).
		Exec(ctx)
}

// ListRecordingsByProject возвращает записи проекта с пагинацией.
func (r *EntRepository) ListRecordingsByProject(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*ent.Recording, int, error) {
	query := r.db.Recording.Query().
		Where(recording.ProjectID(projectID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	recordings, err := query.
		Order(ent.Desc(recording.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return recordings, total, nil
}

// ---------------------------------------------------------------------------
// ScheduledCall методы (CALL-02)
// ---------------------------------------------------------------------------

// CreateScheduledCall создаёт запланированный звонок.
func (r *EntRepository) CreateScheduledCall(ctx context.Context, title, description string, scheduledAt time.Time, rruleStr string, channelID, projectID, userID uuid.UUID) (*ent.ScheduledCall, error) {
	q := r.db.ScheduledCall.Create().
		SetTitle(title).
		SetScheduledAt(scheduledAt).
		SetChannelID(channelID).
		SetProjectID(projectID).
		SetCreatedBy(userID)

	if description != "" {
		q = q.SetDescription(description)
	}
	if rruleStr != "" {
		q = q.SetRrule(rruleStr)
	}

	return q.Save(ctx)
}

// ListScheduledCalls возвращает активные запланированные звонки канала.
func (r *EntRepository) ListScheduledCalls(ctx context.Context, channelID uuid.UUID) ([]*ent.ScheduledCall, error) {
	return r.db.ScheduledCall.Query().
		Where(
			scheduledcall.ChannelID(channelID),
			scheduledcall.IsActive(true),
		).
		Order(ent.Asc(scheduledcall.FieldScheduledAt)).
		All(ctx)
}

// UpdateScheduledCall обновляет запланированный звонок.
func (r *EntRepository) UpdateScheduledCall(ctx context.Context, id uuid.UUID, req UpdateScheduledCallRequest) (*ent.ScheduledCall, error) {
	q := r.db.ScheduledCall.UpdateOneID(id)

	if req.Title != nil {
		q = q.SetTitle(*req.Title)
	}
	if req.Description != nil {
		q = q.SetDescription(*req.Description)
	}
	if req.ScheduledAt != nil {
		q = q.SetScheduledAt(*req.ScheduledAt)
	}
	if req.RRule != nil {
		q = q.SetRrule(*req.RRule)
	}
	if req.IsActive != nil {
		q = q.SetIsActive(*req.IsActive)
	}

	return q.Save(ctx)
}

// DeleteScheduledCall деактивирует запланированный звонок.
func (r *EntRepository) DeleteScheduledCall(ctx context.Context, id uuid.UUID) error {
	return r.db.ScheduledCall.UpdateOneID(id).
		SetIsActive(false).
		Exec(ctx)
}

// GetDueScheduledCalls возвращает запланированные звонки, время которых наступило.
func (r *EntRepository) GetDueScheduledCalls(ctx context.Context, now time.Time) ([]*ent.ScheduledCall, error) {
	return r.db.ScheduledCall.Query().
		Where(
			scheduledcall.IsActive(true),
			scheduledcall.ScheduledAtLTE(now),
		).
		All(ctx)
}

// DeactivateScheduledCall деактивирует запланированный звонок.
func (r *EntRepository) DeactivateScheduledCall(ctx context.Context, id uuid.UUID) error {
	return r.db.ScheduledCall.UpdateOneID(id).
		SetIsActive(false).
		Exec(ctx)
}

// FindConflictingCalls ищет запланированные звонки, которые пересекаются с заданным временным окном
// для каналов, в которых состоят указанные участники.
func (r *EntRepository) FindConflictingCalls(ctx context.Context, participantIDs []uuid.UUID, windowStart, windowEnd time.Time) ([]*ent.ScheduledCall, error) {
	// Сначала находим все каналы, в которых состоят указанные участники
	members, err := r.db.ChannelMember.Query().
		Where(
			channelmember.UserIDIn(participantIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, nil
	}

	channelIDs := make([]uuid.UUID, 0, len(members))
	for _, m := range members {
		channelIDs = append(channelIDs, m.ChannelID)
	}

	// Ищем активные запланированные звонки в этих каналах, пересекающиеся с окном
	// Пересечение: call_start < windowEnd AND call_start + duration > windowStart
	// Поскольку у нас нет поля длительности в БД, считаем DEFAULT_DURATION = 60 мин
	return r.db.ScheduledCall.Query().
		Where(
			scheduledcall.IsActive(true),
			scheduledcall.ChannelIDIn(channelIDs...),
			scheduledcall.ScheduledAtLT(windowEnd),
			scheduledcall.ScheduledAtGTE(windowStart.Add(-60*time.Minute)),
		).
		All(ctx)
}

// UpdateScheduledCallTime обновляет время следующего запланированного звонка.
func (r *EntRepository) UpdateScheduledCallTime(ctx context.Context, id uuid.UUID, nextAt time.Time) error {
	return r.db.ScheduledCall.UpdateOneID(id).
		SetScheduledAt(nextAt).
		Exec(ctx)
}
