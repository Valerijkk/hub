package tracker

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Учёт времени
// ---------------------------------------------------------------------------

// LogTime записывает вручную введённое время на задачу.
func (r *EntRepository) LogTime(ctx context.Context, issueID, userID uuid.UUID, minutes int, description string) (*ent.TimeLog, error) {
	now := time.Now()
	create := r.client.TimeLog.Create().
		SetIssueID(issueID).
		SetUserID(userID).
		SetMinutes(minutes).
		SetStartedAt(now).
		SetEndedAt(now)

	if description != "" {
		create = create.SetDescription(description)
	}

	tl, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("запись времени: %w", err)
	}

	// Обновляем суммарное время на задаче
	err = r.client.Issue.UpdateOneID(issueID).
		AddTimeSpentMinutes(minutes).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление суммарного времени: %w", err)
	}

	return tl, nil
}

// StartTimer запускает таймер на задаче для пользователя.
// A3-005: отклоняет повторный старт, если у пользователя уже есть открытый
// таймер на этой задаче (EndedAtIsNil + StartedAtNotNil) — иначе StopTimer.Only
// становится неоднозначным.
// Exist→Create сериализованы в tx с row-lock parent issue (UPDATE … WHERE id) —
// Ent/SQLite отвергает SELECT FOR UPDATE; UPDATE даёт эквивалентный lock на PG
// и сериализует writers на SQLite. Без lock два concurrent StartTimer оба видят
// «нет открытого» и создают два.
func (r *EntRepository) StartTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin start timer tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Lock parent issue row so concurrent open-timer creates serialize.
	err = tx.Issue.UpdateOneID(issueID).
		SetUpdatedAt(time.Now().UTC()).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("задача не найдена")
		}
		return nil, fmt.Errorf("lock issue for timer: %w", err)
	}

	exists, err := tx.TimeLog.Query().
		Where(
			timelog.HasIssueWith(issue.IDEQ(issueID)),
			timelog.HasUserWith(user.IDEQ(userID)),
			timelog.EndedAtIsNil(),
			timelog.StartedAtNotNil(),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("проверка открытого таймера: %w", err)
	}
	if exists {
		return nil, apperrors.NewConflict("таймер уже запущен")
	}

	now := time.Now()
	// Создаём TimeLog с started_at, но без ended_at и минимальным minutes=1
	// (minutes будет пересчитан при StopTimer)
	tl, err := tx.TimeLog.Create().
		SetIssueID(issueID).
		SetUserID(userID).
		SetMinutes(1). // placeholder, будет пересчитано при остановке
		SetStartedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("запуск таймера: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit start timer: %w", err)
	}
	return tl, nil
}

// StopTimer останавливает запущенный таймер и записывает прошедшее время.
func (r *EntRepository) StopTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error) {
	// Ищем открытый таймер (ended_at is nil, started_at задан — не manual LogTime).
	openTimer, err := r.client.TimeLog.Query().
		Where(
			timelog.HasIssueWith(issue.IDEQ(issueID)),
			timelog.HasUserWith(user.IDEQ(userID)),
			timelog.EndedAtIsNil(),
			timelog.StartedAtNotNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("активный таймер не найден")
		}
		return nil, fmt.Errorf("поиск открытого таймера: %w", err)
	}

	now := time.Now()
	elapsed := int(math.Ceil(now.Sub(*openTimer.StartedAt).Minutes()))
	if elapsed < 1 {
		elapsed = 1
	}

	// Обновляем таймер
	tl, err := r.client.TimeLog.UpdateOneID(openTimer.ID).
		SetEndedAt(now).
		SetMinutes(elapsed).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("остановка таймера: %w", err)
	}

	// Обновляем суммарное время на задаче
	err = r.client.Issue.UpdateOneID(issueID).
		AddTimeSpentMinutes(elapsed).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление суммарного времени: %w", err)
	}

	return tl, nil
}

// ListTimeLogs возвращает все записи времени для задачи.
func (r *EntRepository) ListTimeLogs(ctx context.Context, issueID uuid.UUID) ([]*ent.TimeLog, error) {
	return r.client.TimeLog.Query().
		Where(timelog.HasIssueWith(issue.IDEQ(issueID))).
		WithUser().
		Order(ent.Desc(timelog.FieldCreatedAt)).
		All(ctx)
}
