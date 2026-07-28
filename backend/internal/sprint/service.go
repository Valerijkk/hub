// Package sprint реализует управление спринтами: создание, бэклог, burndown, velocity, Planning Poker.
package sprint

import (
	"context"
	"errors"
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

// GetSprint возвращает спринт по ID в рамках проекта (RT-005 / A3-001 sibling).
func (s *SprintService) GetSprint(ctx context.Context, projectID, sprintID uuid.UUID) (*SprintResponse, error) {
	sp, err := s.repo.GetSprintInProject(ctx, projectID, sprintID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения спринта")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// ListSprints возвращает все спринты проекта.
// A13-006: issue_count через один GetSprintIssueCounts, не N× GetSprintIssueCount.
func (s *SprintService) ListSprints(ctx context.Context, projectID uuid.UUID) ([]SprintResponse, error) {
	sprints, err := s.repo.ListSprints(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения спринтов")
	}

	ids := make([]uuid.UUID, len(sprints))
	for i, sp := range sprints {
		ids[i] = sp.ID
	}
	counts, err := s.repo.GetSprintIssueCounts(ctx, ids)
	if err != nil {
		// W1-A4-012 / P3-BE-01: fail-closed — не отдавать 200 с ложным issue_count=0.
		return nil, apperrors.NewInternal("ошибка получения счётчиков задач спринтов")
	}

	result := make([]SprintResponse, 0, len(sprints))
	for _, sp := range sprints {
		result = append(result, *s.sprintToResponse(sp, counts[sp.ID]))
	}
	return result, nil
}

// UpdateSprint обновляет спринт в рамках проекта (RT-005 / A3-001 sibling).
// W1-A4-007: после merge дат — end не раньше start (как в CreateSprint).
// TASK-095: version обязателен; VersionEQ в repo → 409 при conflict.
func (s *SprintService) UpdateSprint(ctx context.Context, projectID, sprintID uuid.UUID, req UpdateSprintRequest) (*SprintResponse, error) {
	if req.Version < 1 {
		return nil, apperrors.NewBadRequest("version обязателен и должен быть ≥ 1")
	}
	current, err := s.repo.GetSprintInProject(ctx, projectID, sprintID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения спринта")
	}

	start, end := current.StartDate, current.EndDate
	if req.StartDate != nil {
		start = *req.StartDate
	}
	if req.EndDate != nil {
		end = *req.EndDate
	}
	if end.Before(start) {
		return nil, apperrors.NewBadRequest("дата окончания должна быть после даты начала")
	}

	sp, err := s.repo.UpdateSprint(ctx, sprintID, req)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewConflict("version mismatch: спринт был изменён другим пользователем")
		}
		return nil, apperrors.NewInternal("ошибка обновления спринта")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// StartSprint переводит спринт в статус active.
// W1-A4-009: check+activate атомарны в StartSprintAtomic (row-lock проекта).
// A3-016: DB error при поиске active → fail closed.
func (s *SprintService) StartSprint(ctx context.Context, projectID, sprintID uuid.UUID) (*SprintResponse, error) {
	sp, err := s.repo.StartSprintAtomic(ctx, projectID, sprintID)
	if err != nil {
		var conflict *ErrActiveSprintConflict
		if errors.As(err, &conflict) {
			return nil, apperrors.NewConflict("в проекте уже есть активный спринт: " + conflict.Name)
		}
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка запуска спринта")
	}

	return s.toSprintResponse(ctx, sp), nil
}

// EnsureIssueInProject проверяет принадлежность issue проекту (W1-A4-006, poker start).
func (s *SprintService) EnsureIssueInProject(ctx context.Context, projectID, issueID uuid.UUID) error {
	if err := s.repo.IssueInProject(ctx, projectID, issueID); err != nil {
		if errors.Is(err, ErrIssueNotInProject) {
			return apperrors.NewBadRequest("задача не принадлежит проекту")
		}
		return apperrors.NewInternal("ошибка проверки задачи")
	}
	return nil
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

	issueIDs := make([]uuid.UUID, len(incomplete))
	for i, iss := range incomplete {
		issueIDs[i] = iss.ID
	}

	// P1-BE-05: create-next + move + status=completed в одной tx (иначе рассинхрон при mid-fail).
	updated, err := s.repo.CompleteSprintAtomic(ctx, projectID, sprintID, sp, issueIDs)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка завершения спринта")
	}

	if len(issueIDs) > 0 {
		slog.Info("перенесены незавершённые задачи",
			"count", len(issueIDs),
			"from_sprint", sprintID,
		)
	}

	return s.toSprintResponse(ctx, updated), nil
}

// UpdateRetroNotes сохраняет заметки ретроспективы (SCRM-03; RT-005 / A3-001 sibling).
func (s *SprintService) UpdateRetroNotes(ctx context.Context, projectID, sprintID uuid.UUID, req RetroNotesRequest) (*SprintResponse, error) {
	if _, err := s.repo.GetSprintInProject(ctx, projectID, sprintID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("спринт не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения спринта")
	}

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

// GetBurndown возвращает данные для графика burndown (SCRM-02; RT-005 / A3-001 sibling).
func (s *SprintService) GetBurndown(ctx context.Context, projectID, sprintID uuid.UUID) ([]BurndownPoint, error) {
	sp, err := s.repo.GetSprintInProject(ctx, projectID, sprintID)
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

// AddIssuesToSprint добавляет задачи в спринт в рамках проекта (RT-005).
// После GetSprintInProject проверяет, что все issueIDs принадлежат тому же project.
func (s *SprintService) AddIssuesToSprint(ctx context.Context, projectID, sprintID uuid.UUID, issueIDs []uuid.UUID) error {
	if _, err := s.repo.GetSprintInProject(ctx, projectID, sprintID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("спринт не найден")
		}
		return apperrors.NewInternal("ошибка получения спринта")
	}
	if err := s.repo.SetIssuesSprint(ctx, projectID, issueIDs, sprintID); err != nil {
		if errors.Is(err, ErrIssueNotInProject) {
			return apperrors.NewBadRequest("задачи не принадлежат проекту")
		}
		return apperrors.NewInternal("ошибка добавления задач в спринт")
	}
	return nil
}

// RemoveIssuesFromSprint убирает задачи из спринта (возвращает в бэклог).
// projectID обязателен — иначе cross-project clear (P3-BE-05 / W1-A4-013).
func (s *SprintService) RemoveIssuesFromSprint(ctx context.Context, projectID uuid.UUID, issueIDs []uuid.UUID) error {
	if err := s.repo.ClearIssuesSprint(ctx, projectID, issueIDs); err != nil {
		if errors.Is(err, ErrIssueNotInProject) {
			return apperrors.NewBadRequest("задачи не принадлежат проекту")
		}
		return apperrors.NewInternal("ошибка удаления задач из спринта")
	}
	return nil
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

// toSprintResponse конвертирует Ent Sprint в DTO (single-item: edges или 1× Count).
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
	return s.sprintToResponse(sp, issueCount)
}

// sprintToResponse собирает DTO с уже известным issue_count (list path preload).
func (s *SprintService) sprintToResponse(sp *ent.Sprint, issueCount int) *SprintResponse {
	return &SprintResponse{
		ID:         sp.ID.String(),
		Name:       sp.Name,
		Goal:       sp.Goal,
		Status:     string(sp.Status),
		StartDate:  sp.StartDate,
		EndDate:    sp.EndDate,
		RetroNotes: sp.RetroNotes,
		IssueCount: issueCount,
		Version:    sp.Version,
		CreatedAt:  sp.CreatedAt,
		UpdatedAt:  sp.UpdatedAt,
	}
}
