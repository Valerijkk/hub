package calls

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/teambition/rrule-go"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// ScheduledCallService реализует бизнес-логику запланированных и повторяющихся звонков.
// Поддерживает RFC 5545 RRULE для задания расписания повторяющихся звонков
// (ежедневный стендап, еженедельное демо и т.д.).
type ScheduledCallService struct {
	repo *EntRepository
	nats *natsclient.Client
}

// NewScheduledCallService создаёт новый сервис запланированных звонков.
func NewScheduledCallService(repo *EntRepository, nc *natsclient.Client) *ScheduledCallService {
	return &ScheduledCallService{
		repo: repo,
		nats: nc,
	}
}

// requireChannelMember — W3-A5-003/004: Create/List только для члена канала.
func (s *ScheduledCallService) requireChannelMember(ctx context.Context, channelID, userID uuid.UUID) error {
	ok, err := s.repo.IsChannelMember(ctx, channelID, userID)
	if err != nil {
		return fmt.Errorf("проверка членства в канале: %w", err)
	}
	if !ok {
		return ErrCallForbidden
	}
	return nil
}

// authorizeScheduledCallMutate — W3-A5-001/002: Update/Delete — creator или owner/admin канала.
func (s *ScheduledCallService) authorizeScheduledCallMutate(ctx context.Context, sc *ent.ScheduledCall, userID uuid.UUID) error {
	if sc.CreatedBy == userID {
		return nil
	}
	role, err := s.repo.GetChannelMemberRole(ctx, sc.ChannelID, userID)
	if err != nil {
		return fmt.Errorf("проверка роли в канале: %w", err)
	}
	if role == "owner" || role == "admin" {
		return nil
	}
	return ErrCallForbidden
}

// CreateScheduledCall создаёт запланированный звонок.
// Если задан RRULE, парсит его и вычисляет ближайшее время запланированного звонка.
// W3-A5-003: требует членство в канале и совпадение channel.project_id с path projectID.
func (s *ScheduledCallService) CreateScheduledCall(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, req CreateScheduledCallRequest) (*ScheduledCallResponse, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("название обязательно")
	}
	if req.ChannelID == uuid.Nil {
		return nil, fmt.Errorf("channel_id обязателен")
	}

	if err := s.requireChannelMember(ctx, req.ChannelID, userID); err != nil {
		return nil, err
	}
	chProjectID, err := s.repo.GetProjectIDByChannel(ctx, req.ChannelID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCallForbidden
		}
		return nil, fmt.Errorf("проверка проекта канала: %w", err)
	}
	if chProjectID != projectID {
		return nil, ErrCallForbidden
	}

	scheduledAt := req.ScheduledAt

	// Если задан RRULE — валидируем и вычисляем ближайшее время
	if req.RRule != "" {
		rule, err := rrule.StrToRRule(req.RRule)
		if err != nil {
			return nil, fmt.Errorf("некорректный RRULE: %w", err)
		}
		next := rule.After(time.Now().UTC(), false)
		if !next.IsZero() {
			scheduledAt = next
		}
	}

	sc, err := s.repo.CreateScheduledCall(ctx, req.Title, req.Description, scheduledAt, req.RRule, req.ChannelID, projectID, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запланированный звонок: %w", err)
	}

	return toScheduledCallResponse(sc), nil
}

// ListScheduledCalls возвращает список запланированных звонков канала.
// Возвращает только активные звонки, отсортированные по времени.
// W3-A5-004: только члены канала видят расписание.
func (s *ScheduledCallService) ListScheduledCalls(ctx context.Context, channelID, userID uuid.UUID) ([]*ScheduledCallResponse, error) {
	if err := s.requireChannelMember(ctx, channelID, userID); err != nil {
		return nil, err
	}

	calls, err := s.repo.ListScheduledCalls(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить запланированные звонки: %w", err)
	}

	result := make([]*ScheduledCallResponse, len(calls))
	for i, c := range calls {
		result[i] = toScheduledCallResponse(c)
	}

	return result, nil
}

// UpdateScheduledCall обновляет запланированный звонок.
// W3-A5-001: только creator или owner/admin канала.
func (s *ScheduledCallService) UpdateScheduledCall(ctx context.Context, id, userID uuid.UUID, req UpdateScheduledCallRequest) (*ScheduledCallResponse, error) {
	sc, err := s.repo.GetScheduledCall(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCallNotFound
		}
		return nil, fmt.Errorf("получить запланированный звонок: %w", err)
	}
	if err := s.authorizeScheduledCallMutate(ctx, sc, userID); err != nil {
		return nil, err
	}

	// Валидируем RRULE если обновляется
	if req.RRule != nil && *req.RRule != "" {
		if _, err := rrule.StrToRRule(*req.RRule); err != nil {
			return nil, fmt.Errorf("некорректный RRULE: %w", err)
		}
	}

	updated, err := s.repo.UpdateScheduledCall(ctx, id, req)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCallNotFound
		}
		return nil, fmt.Errorf("не удалось обновить запланированный звонок: %w", err)
	}

	return toScheduledCallResponse(updated), nil
}

// DeleteScheduledCall удаляет запланированный звонок (деактивирует).
// W3-A5-002: только creator или owner/admin канала.
func (s *ScheduledCallService) DeleteScheduledCall(ctx context.Context, id, userID uuid.UUID) error {
	sc, err := s.repo.GetScheduledCall(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCallNotFound
		}
		return fmt.Errorf("получить запланированный звонок: %w", err)
	}
	if err := s.authorizeScheduledCallMutate(ctx, sc, userID); err != nil {
		return err
	}

	if err := s.repo.DeleteScheduledCall(ctx, id); err != nil {
		return fmt.Errorf("не удалось удалить запланированный звонок: %w", err)
	}
	return nil
}

// ProcessDueScheduledCalls обрабатывает запланированные звонки, время которых наступило.
// Вызывается cron-задачей каждую минуту. Для повторяющихся звонков вычисляет
// следующее вхождение по RRULE и обновляет scheduled_at.
func (s *ScheduledCallService) ProcessDueScheduledCalls(ctx context.Context) error {
	dueCalls, err := s.repo.GetDueScheduledCalls(ctx, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("не удалось получить наступившие звонки: %w", err)
	}

	for _, sc := range dueCalls {
		// Публикуем NATS событие для уведомления через WS
		s.nats.PublishJSON("calls.scheduled.due", map[string]interface{}{
			"scheduled_call_id": sc.ID,
			"title":             sc.Title,
			"channel_id":        sc.ChannelID,
			"project_id":        sc.ProjectID,
		})

		if sc.Rrule != "" {
			// Повторяющийся — вычисляем следующее вхождение
			rule, err := rrule.StrToRRule(sc.Rrule)
			if err != nil {
				slog.Error("не удалось распарсить RRULE", "id", sc.ID, "rrule", sc.Rrule, "error", err)
				continue
			}

			next := rule.After(time.Now().UTC(), false)
			if next.IsZero() {
				// Больше вхождений нет — деактивируем
				if err := s.repo.DeactivateScheduledCall(ctx, sc.ID); err != nil {
					slog.Error("не удалось деактивировать звонок", "id", sc.ID, "error", err)
				}
			} else {
				// Обновляем на следующее время
				if err := s.repo.UpdateScheduledCallTime(ctx, sc.ID, next); err != nil {
					slog.Error("не удалось обновить время звонка", "id", sc.ID, "error", err)
				}
			}
		} else {
			// Одноразовый — деактивируем
			if err := s.repo.DeactivateScheduledCall(ctx, sc.ID); err != nil {
				slog.Error("не удалось деактивировать звонок", "id", sc.ID, "error", err)
			}
		}
	}

	return nil
}

// toScheduledCallResponse конвертирует Ent ScheduledCall в DTO ответа.
func toScheduledCallResponse(sc *ent.ScheduledCall) *ScheduledCallResponse {
	return &ScheduledCallResponse{
		ID:          sc.ID,
		Title:       sc.Title,
		Description: sc.Description,
		ScheduledAt: sc.ScheduledAt,
		RRule:       sc.Rrule,
		ChannelID:   sc.ChannelID,
		ProjectID:   sc.ProjectID,
		CreatedBy:   sc.CreatedBy,
		IsActive:    sc.IsActive,
		CreatedAt:   sc.CreatedAt,
		UpdatedAt:   sc.UpdatedAt,
	}
}

// CheckConflicts проверяет конфликты расписания для заданного времени и участников.
func (s *ScheduledCallService) CheckConflicts(ctx context.Context, requestedTime time.Time, durationMinutes int, participantIDs []uuid.UUID) (*ConflictResponse, error) {
	if len(participantIDs) == 0 {
		return &ConflictResponse{HasConflict: false, Conflicts: []ConflictDetail{}}, nil
	}

	windowStart := requestedTime
	windowEnd := requestedTime.Add(time.Duration(durationMinutes) * time.Minute)

	calls, err := s.repo.FindConflictingCalls(ctx, participantIDs, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("не удалось проверить конфликты: %w", err)
	}

	// Находим каналы участников для маппинга
	conflicts := make([]ConflictDetail, 0)

	// Для каждого найденного звонка проверяем реальное пересечение
	for _, sc := range calls {
		callStart := sc.ScheduledAt
		callEnd := callStart.Add(60 * time.Minute) // Дефолтная длительность

		// Проверяем пересечение
		if callStart.Before(windowEnd) && callEnd.After(windowStart) {
			// Определяем, какие из запрошенных участников в этом канале
			for _, pid := range participantIDs {
				conflicts = append(conflicts, ConflictDetail{
					ParticipantID: pid,
					CallTitle:     sc.Title,
					ScheduledAt:   sc.ScheduledAt,
					ChannelID:     sc.ChannelID,
				})
			}
		}
	}

	return &ConflictResponse{
		HasConflict: len(conflicts) > 0,
		Conflicts:   conflicts,
	}, nil
}

// Подавление unused import — scheduledcall используется в repository
var _ = scheduledcall.Table
