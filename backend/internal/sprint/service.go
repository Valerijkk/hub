// Package sprint реализует управление спринтами: создание, бэклог, burndown, velocity, Planning Poker.
package sprint

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/redis/go-redis/v9"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// SprintService реализует бизнес-логику модуля спринтов.
// Управляет жизненным циклом спринтов, Planning Poker и WIP-лимитами.
type SprintService struct {
	repo  Repository
	redis *redis.Client
	nats  *natsclient.Client
}

// NewSprintService создаёт новый сервис спринтов.
func NewSprintService(repo Repository, rdb *redis.Client, nc *natsclient.Client) *SprintService {
	return &SprintService{repo: repo, redis: rdb, nats: nc}
}

// CreateSprint создаёт спринт в статусе planning.
func (s *SprintService) CreateSprint(ctx context.Context, projectID uuid.UUID, req CreateSprintRequest) (*SprintResponse, error) {
	if req.EndDate.Before(req.StartDate) {
		return nil, apperrors.NewBadRequest("дата окончания должна быть после даты начала")
	}

	sp, err := s.repo.CreateSprint(ctx, projectID, req)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания спринта")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// GetSprint возвращает спринт по ID.
func (s *SprintService) GetSprint(ctx context.Context, sprintID uuid.UUID) (*SprintResponse, error) {
	sp, err := s.repo.GetSprint(ctx, sprintID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения спринта")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// ListSprints возвращает все спринты проекта.
func (s *SprintService) ListSprints(ctx context.Context, projectID uuid.UUID) ([]SprintResponse, error) {
	sprints, err := s.repo.ListSprints(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения спринтов")
	}

	result := make([]SprintResponse, 0, len(sprints))
	for _, sp := range sprints {
		result = append(result, *s.toSprintResponse(ctx, sp))
	}
	return result, nil
}

// UpdateSprint обновляет спринт.
func (s *SprintService) UpdateSprint(ctx context.Context, sprintID uuid.UUID, req UpdateSprintRequest) (*SprintResponse, error) {
	sp, err := s.repo.UpdateSprint(ctx, sprintID, req)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления спринта")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// StartSprint переводит спринт в статус active.
// Возвращает ошибку если в проекте уже есть активный спринт.
func (s *SprintService) StartSprint(ctx context.Context, projectID, sprintID uuid.UUID) (*SprintResponse, error) {
	// A3-016 / RT-005 sibling: project-scope before status mutation (same as CompleteSprint).
	if _, err := s.repo.GetSprintInProject(ctx, projectID, sprintID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения спринта")
	}

	// A3-016: DB error from FindActiveSprint must fail closed (not skip conflict check).
	existing, err := s.repo.FindActiveSprint(ctx, projectID)
	if err != nil && !ent.IsNotFound(err) {
		return nil, apperrors.NewInternal("ошибка поиска активного спринта")
	}
	if err == nil && existing.ID != sprintID {
		return nil, apperrors.NewConflict("в проекте уже есть активный спринт: " + existing.Name)
	}

	sp, err := s.repo.UpdateSprintStatus(ctx, sprintID, sprintEnt.StatusActive)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка запуска спринта")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// CompleteSprint завершает спринт и переносит незавершённые задачи (SCRM-04).
func (s *SprintService) CompleteSprint(ctx context.Context, projectID, sprintID uuid.UUID) (*SprintResponse, error) {
	sp, err := s.repo.GetSprintInProject(ctx, projectID, sprintID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения спринта")
	}

	if sp.Status != sprintEnt.StatusActive {
		return nil, apperrors.NewBadRequest("можно завершить только активный спринт")
	}

	// Находим незавершённые задачи
	incomplete, err := s.repo.FindIncompleteIssues(ctx, sprintID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка поиска незавершённых задач")
	}

	// Переносим незавершённые задачи в следующий спринт
	if len(incomplete) > 0 {
		nextSprint, err := s.repo.FindOrCreateNextPlanningSprint(ctx, projectID, sp)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка создания следующего спринта")
		}

		issueIDs := make([]uuid.UUID, len(incomplete))
		for i, iss := range incomplete {
			issueIDs[i] = iss.ID
		}
		if err := s.repo.MoveIssuesToSprint(ctx, issueIDs, nextSprint.ID); err != nil {
			return nil, apperrors.NewInternal("ошибка переноса задач")
		}

		slog.Info("перенесены незавершённые задачи",
			"count", len(issueIDs),
			"from_sprint", sprintID,
			"to_sprint", nextSprint.ID,
		)
	}

	// Обновляем статус
	updated, err := s.repo.UpdateSprintStatus(ctx, sprintID, sprintEnt.StatusCompleted)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка завершения спринта")
	}

	return s.toSprintResponse(ctx, updated), nil
}

// UpdateRetroNotes сохраняет заметки ретроспективы (SCRM-03).
func (s *SprintService) UpdateRetroNotes(ctx context.Context, sprintID uuid.UUID, req RetroNotesRequest) (*SprintResponse, error) {
	notes := map[string]interface{}{
		"went_well":  req.WentWell,
		"to_improve": req.ToImprove,
		"actions":    req.Actions,
	}

	sp, err := s.repo.UpdateRetroNotes(ctx, sprintID, notes)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка сохранения заметок ретроспективы")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// GetBurndown возвращает данные для графика burndown (SCRM-02).
func (s *SprintService) GetBurndown(ctx context.Context, sprintID uuid.UUID) ([]BurndownPoint, error) {
	sp, err := s.repo.GetSprint(ctx, sprintID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения спринта")
	}

	total, done, err := s.repo.GetBurndownData(ctx, sprintID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения данных burndown")
	}

	remaining := total - done
	days := int(math.Ceil(sp.EndDate.Sub(sp.StartDate).Hours() / 24))
	if days < 1 {
		days = 1
	}

	now := time.Now().UTC()
	points := make([]BurndownPoint, 0, days+1)
	for d := 0; d <= days; d++ {
		day := sp.StartDate.AddDate(0, 0, d)
		dayStr := day.Format("2006-01-02")

		idealRemaining := total - int(math.Round(float64(total)*float64(d)/float64(days)))

		actualRemaining := total
		if day.After(now) {
			// Будущие дни показывают текущий remaining (горизонтальная линия).
			actualRemaining = remaining
		} else if day.Before(now) || day.Equal(now) {
			// Линейная интерполяция фактического прогресса.
			//
			// Логика: мы знаем сколько задач сделано с начала спринта до now
			// (completedSoFar = total - remaining). Для каждого прошедшего дня d
			// масштабируем прогресс пропорционально:
			//   ratio = d / elapsed — какую долю прошедших дней покрывает день d
			//   completedByDay = completedSoFar * ratio — выполнено к этому дню
			//   actualRemaining = total - completedByDay — остаток к этому дню
			//
			// Это даёт линейную аппроксимацию от total (день 0) до remaining (сегодня).
			elapsed := now.Sub(sp.StartDate).Hours() / 24
			if elapsed > 0 && d > 0 {
				ratio := float64(d) / elapsed
				actualRemaining = total - int(math.Round(float64(total-remaining)*ratio))
				if actualRemaining < 0 {
					actualRemaining = 0
				}
			}
		}

		points = append(points, BurndownPoint{
			Date:      dayStr,
			Remaining: actualRemaining,
			Ideal:     idealRemaining,
		})
	}

	return points, nil
}

// GetVelocity возвращает данные velocity по проекту (SCRM-02).
func (s *SprintService) GetVelocity(ctx context.Context, projectID uuid.UUID) ([]VelocityPoint, error) {
	return s.repo.GetVelocityData(ctx, projectID)
}

// AddIssuesToSprint добавляет задачи в спринт.
func (s *SprintService) AddIssuesToSprint(ctx context.Context, sprintID uuid.UUID, issueIDs []uuid.UUID) error {
	return s.repo.SetIssuesSprint(ctx, issueIDs, sprintID)
}

// RemoveIssuesFromSprint убирает задачи из спринта (возвращает в бэклог).
func (s *SprintService) RemoveIssuesFromSprint(ctx context.Context, issueIDs []uuid.UUID) error {
	return s.repo.ClearIssuesSprint(ctx, issueIDs)
}

// CheckWIPLimit проверяет WIP-лимит для целевого статуса (KANB-02).
// Возвращает ошибку 409 если лимит достигнут.
// A3-015: status not found / count error → fail closed (не silent skip).
func (s *SprintService) CheckWIPLimit(ctx context.Context, statusID uuid.UUID) error {
	ws, err := s.repo.GetWorkflowStatus(ctx, statusID)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("статус не найден")
		}
		return apperrors.NewInternal("ошибка проверки WIP-лимита")
	}

	if ws.WipLimit <= 0 {
		return nil // Лимит не установлен
	}

	count, err := s.repo.CountIssuesInStatus(ctx, statusID)
	if err != nil {
		return apperrors.NewInternal("ошибка подсчёта задач для WIP-лимита")
	}

	if count >= ws.WipLimit {
		return apperrors.NewConflict(
			fmt.Sprintf("WIP-лимит достигнут: %d/%d задач в статусе %s", count, ws.WipLimit, ws.Name),
		)
	}

	return nil
}

// toSprintResponse конвертирует Ent Sprint в DTO.
func (s *SprintService) toSprintResponse(ctx context.Context, sp *ent.Sprint) *SprintResponse {
	issueCount := 0
	if sp.Edges.Issues != nil {
		issueCount = len(sp.Edges.Issues)
	} else {
		cnt, err := s.repo.GetSprintIssueCount(ctx, sp.ID)
		if err == nil {
			issueCount = cnt
		}
	}

	return &SprintResponse{
		ID:         sp.ID.String(),
		Name:       sp.Name,
		Goal:       sp.Goal,
		Status:     string(sp.Status),
		StartDate:  sp.StartDate,
		EndDate:    sp.EndDate,
		RetroNotes: sp.RetroNotes,
		IssueCount: issueCount,
		CreatedAt:  sp.CreatedAt,
		UpdatedAt:  sp.UpdatedAt,
	}
}
