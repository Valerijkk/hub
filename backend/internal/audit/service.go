// Package audit реализует журнал аудита и политики хранения данных.
package audit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/valerii/hub/backend/internal/auth"
)

// Настройки аудита, лимитов и корзины.
const (
	// Ключи лимитов в SystemSetting.
	LimitFileMaxSizeMB       = "limit.file.max_size_mb"
	LimitUserMaxProjects     = "limit.user.max_projects"
	LimitProjectMaxStorageMB = "limit.project.max_storage_mb"

	// TrashRetentionDays — период хранения в корзине (дней).
	TrashRetentionDays = 30
)

// AuditService содержит бизнес-логику для аудита, лимитов, корзины и управления пользователями.
//
// SC-14-3: сервис не хранит *ent.Client напрямую — все БД-запросы идут через
// единый *Repository. Это устраняет дубль db+repo и облегчает unit-тестирование
// (репозиторий моделируется в памяти без поднятия Ent).
type AuditService struct {
	repo  *Repository
	redis *redis.Client // optional — используется для инвалидации auth-кэшей (ME-13).
}

// NewAuditService создаёт новый сервис аудита.
// rdb — опциональный Redis-клиент; если задан, AuditService инвалидирует
// auth.UserStatusCacheKey при деактивации пользователя (ME-13). Может быть
// nil в тестах — деактивация всё равно отработает, только статус будет
// применён по истечении TTL (~30s).
func NewAuditService(repo *Repository, rdb *redis.Client) *AuditService {
	return &AuditService{repo: repo, redis: rdb}
}

// IsSystemAdmin возвращает true, если пользователь — системный администратор.
// Используется HTTP-обработчиками для проверки доступа к admin-эндпоинтам.
func (s *AuditService) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return s.repo.IsSystemAdmin(ctx, userID)
}

// ---- Журнал аудита ----

// ListAuditLogs возвращает пагинированный список записей аудита с фильтрами (только для администратора).
func (s *AuditService) ListAuditLogs(ctx context.Context, filters AuditLogFilters) (*AuditLogListResponse, error) {
	items, total, err := s.repo.ListAuditLogs(ctx, filters)
	if err != nil {
		return nil, err
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	perPage := filters.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 50
	}

	resp := &AuditLogListResponse{
		Items:   make([]AuditLogResponse, 0, len(items)),
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}

	for _, item := range items {
		r := AuditLogResponse{
			ID:         item.ID.String(),
			Action:     item.Action,
			EntityType: item.EntityType,
			EntityID:   item.EntityID,
			Details:    item.Details,
			IPAddress:  item.IPAddress,
			UserAgent:  item.UserAgent,
			CreatedAt:  item.CreatedAt,
		}
		if item.UserID != nil {
			uid := item.UserID.String()
			r.UserID = &uid
		}
		if item.ProjectID != nil {
			pid := item.ProjectID.String()
			r.ProjectID = &pid
		}
		resp.Items = append(resp.Items, r)
	}

	return resp, nil
}

// ---- Лимиты ресурсов ----

// GetResourceLimits возвращает текущие лимиты ресурсов из SystemSetting.
func (s *AuditService) GetResourceLimits(ctx context.Context) ([]ResourceLimitResponse, error) {
	keys := []string{LimitFileMaxSizeMB, LimitUserMaxProjects, LimitProjectMaxStorageMB}
	defaults := map[string]string{
		LimitFileMaxSizeMB:       "50",
		LimitUserMaxProjects:     "20",
		LimitProjectMaxStorageMB: "5120",
	}

	result := make([]ResourceLimitResponse, 0, len(keys))
	for _, key := range keys {
		value, err := s.repo.GetSystemSetting(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("получение лимита %s: %w", key, err)
		}
		if value == "" {
			value = defaults[key]
		}
		result = append(result, ResourceLimitResponse{Key: key, Value: value})
	}

	return result, nil
}

// UpdateResourceLimit обновляет лимит ресурса в SystemSetting (только для администратора).
func (s *AuditService) UpdateResourceLimit(ctx context.Context, key, value string) error {
	return s.repo.UpsertSystemSetting(ctx, key, value, fmt.Sprintf("Лимит ресурса: %s", key))
}

// CheckLimit проверяет, не превышен ли лимит ресурса.
func (s *AuditService) CheckLimit(ctx context.Context, limitKey string, currentValue int) (bool, error) {
	value, err := s.repo.GetSystemSetting(ctx, limitKey)
	if err != nil {
		return false, err
	}
	if value == "" {
		return true, nil // Лимит не задан — разрешаем
	}

	limitVal, err := strconv.Atoi(value)
	if err != nil {
		return true, nil // Невалидное значение — разрешаем
	}

	return currentValue < limitVal, nil
}

// ---- Корзина (мягкое удаление) ----
// Каждый тип сущности (issue, test_case, file, project) обрабатывается
// отдельным switch-case из-за type-safe API Ent ORM — у каждого типа
// свои предикаты, поля и query builder. Обобщение через interface{}
// или reflect усложнило бы код без выигрыша в поддерживаемости.

// ListTrash возвращает пагинированный список элементов в корзине.
func (s *AuditService) ListTrash(ctx context.Context, entityType string, projectID *uuid.UUID, page, perPage int) (*TrashListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage
	cutoff := time.Now().Add(-TrashRetentionDays * 24 * time.Hour)

	resp := &TrashListResponse{
		Items:   []TrashItemResponse{},
		Page:    page,
		PerPage: perPage,
	}

	var (
		items []TrashItem
		total int
		err   error
	)

	switch entityType {
	case "issue":
		items, total, err = s.repo.ListDeletedIssues(ctx, projectID, cutoff, offset, perPage)
	case "test_case":
		items, total, err = s.repo.ListDeletedTestCases(ctx, projectID, cutoff, offset, perPage)
	case "file":
		items, total, err = s.repo.ListDeletedFiles(ctx, projectID, cutoff, offset, perPage)
	case "project":
		items, total, err = s.repo.ListDeletedProjects(ctx, cutoff, offset, perPage)
	default:
		return resp, nil
	}
	if err != nil {
		return nil, err
	}

	resp.Total = total
	for _, it := range items {
		resp.Items = append(resp.Items, TrashItemResponse{
			ID:            it.ID.String(),
			Name:          it.Name,
			EntityType:    it.EntityType,
			DeletedAt:     it.DeletedAt,
			DaysRemaining: daysRemaining(it.DeletedAt),
		})
	}
	return resp, nil
}

// RestoreItem восстанавливает элемент из корзины (сбрасывает deleted_at).
func (s *AuditService) RestoreItem(ctx context.Context, entityType string, entityID uuid.UUID) error {
	switch entityType {
	case "issue":
		return s.repo.RestoreIssue(ctx, entityID)
	case "test_case":
		return s.repo.RestoreTestCase(ctx, entityID)
	case "file":
		return s.repo.RestoreFile(ctx, entityID)
	case "project":
		return s.repo.RestoreProject(ctx, entityID)
	default:
		return fmt.Errorf("неизвестный тип сущности: %s", entityType)
	}
}

// SoftDelete устанавливает deleted_at = now для указанной сущности.
func (s *AuditService) SoftDelete(ctx context.Context, entityType string, entityID uuid.UUID) error {
	now := time.Now()
	switch entityType {
	case "issue":
		return s.repo.SoftDeleteIssue(ctx, entityID, now)
	case "test_case":
		return s.repo.SoftDeleteTestCase(ctx, entityID, now)
	case "file":
		return s.repo.SoftDeleteFile(ctx, entityID, now)
	case "project":
		return s.repo.SoftDeleteProject(ctx, entityID, now)
	default:
		return fmt.Errorf("неизвестный тип сущности: %s", entityType)
	}
}

// PurgeExpired окончательно удаляет элементы, находящиеся в корзине дольше 30 дней.
// Запускается по cron (горутина с 24-часовым тикером).
func (s *AuditService) PurgeExpired(ctx context.Context) error {
	cutoff := time.Now().Add(-TrashRetentionDays * 24 * time.Hour)

	// purge вызывает deleteFunc и логирует результат.
	purge := func(entity string, deleteFunc func() (int, error)) {
		n, err := deleteFunc()
		if err != nil {
			slog.Error("purge expired "+entity, "error", err)
		} else if n > 0 {
			slog.Info("purged expired "+entity, "count", n)
		}
	}

	purge("issues", func() (int, error) { return s.repo.PurgeExpiredIssues(ctx, cutoff) })
	purge("test_cases", func() (int, error) { return s.repo.PurgeExpiredTestCases(ctx, cutoff) })
	purge("files", func() (int, error) { return s.repo.PurgeExpiredFiles(ctx, cutoff) })
	purge("projects", func() (int, error) { return s.repo.PurgeExpiredProjects(ctx, cutoff) })

	return nil
}

// ---- Управление пользователями ----

// DeactivateUser деактивирует пользователя и отзывает все его сессии.
func (s *AuditService) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	// Обновляем статус пользователя
	if err := s.repo.DeactivateUser(ctx, userID); err != nil {
		return fmt.Errorf("деактивация пользователя: %w", err)
	}

	// Отзываем все сессии
	if _, err := s.repo.DeleteUserSessions(ctx, userID); err != nil {
		slog.Error("revoke sessions on deactivate", "error", err, "user_id", userID)
	}

	// ME-13: инвалидируем закэшированный статус, чтобы запретить доступ
	// немедленно, а не через TTL окна (~30s).
	auth.InvalidateUserStatusCache(ctx, s.redis, userID)

	return nil
}

// ReassignTasks переназначает все задачи одного пользователя другому.
func (s *AuditService) ReassignTasks(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error) {
	n, err := s.repo.ReassignAssignedIssues(ctx, fromUserID, toUserID)
	if err != nil {
		return 0, fmt.Errorf("переназначение задач: %w", err)
	}
	return n, nil
}

// DeleteUserGDPR выполняет каскадное удаление пользователя в соответствии с GDPR:
// 1. Переназначает задачи на указанного пользователя (или очищает assignee)
// 2. Удаляет комментарии пользователя
// 3. Удаляет все сессии
// 4. Hard-delete записи пользователя
// 5. Записывает аудит-лог "GDPR deletion"
//
// BE-H6: ошибки каскадного удаления (reactions, sessions, time_logs,
// activity_logs, comments) накапливаются через errors.Join и возвращаются
// после Hard-delete пользователя — раньше они молча проглатывались
// (`_, _ = ...`), и handler возвращал 200 даже при частично проваленной
// GDPR-операции, что нарушает требование "right to erasure".
func (s *AuditService) DeleteUserGDPR(ctx context.Context, userID uuid.UUID, reassignTo *uuid.UUID) error {
	// 1. Переназначаем или очищаем задачи
	if reassignTo != nil {
		if _, err := s.repo.ReassignAssignedIssues(ctx, userID, *reassignTo); err != nil {
			return fmt.Errorf("переназначение задач (GDPR): %w", err)
		}
		// Также переназначаем задачи где пользователь reporter
		if _, err := s.repo.ReassignReportedIssues(ctx, userID, *reassignTo); err != nil {
			return fmt.Errorf("переназначение reporter-задач (GDPR): %w", err)
		}
	} else {
		if _, err := s.repo.ClearUserAssignments(ctx, userID); err != nil {
			return fmt.Errorf("очистка задач пользователя (GDPR): %w", err)
		}
	}

	// Каскадные удаления связанных сущностей. Ошибки накапливаем —
	// часть данных пользователя может остаться, и админ должен узнать.
	var cascadeErrs []error

	// 2. Удаляем комментарии пользователя
	if n, err := s.repo.DeleteUserComments(ctx, userID); err != nil {
		slog.Error("delete comments (GDPR)", "error", err, "user_id", userID)
		cascadeErrs = append(cascadeErrs, fmt.Errorf("delete comments: %w", err))
	} else if n > 0 {
		slog.Info("deleted user comments (GDPR)", "count", n, "user_id", userID)
	}

	// 3. Удаляем реакции пользователя
	if _, err := s.repo.DeleteUserReactions(ctx, userID); err != nil {
		slog.Error("delete reactions (GDPR)", "error", err, "user_id", userID)
		cascadeErrs = append(cascadeErrs, fmt.Errorf("delete reactions: %w", err))
	}

	// 4. Удаляем все сессии
	if _, err := s.repo.DeleteUserSessions(ctx, userID); err != nil {
		slog.Error("delete sessions (GDPR)", "error", err, "user_id", userID)
		cascadeErrs = append(cascadeErrs, fmt.Errorf("delete sessions: %w", err))
	}

	// 5. Удаляем логи активности пользователя
	if _, err := s.repo.DeleteUserActivityLogs(ctx, userID); err != nil {
		slog.Error("delete activity_logs (GDPR)", "error", err, "user_id", userID)
		cascadeErrs = append(cascadeErrs, fmt.Errorf("delete activity_logs: %w", err))
	}

	// 6. Удаляем записи времени пользователя
	if _, err := s.repo.DeleteUserTimeLogs(ctx, userID); err != nil {
		slog.Error("delete time_logs (GDPR)", "error", err, "user_id", userID)
		cascadeErrs = append(cascadeErrs, fmt.Errorf("delete time_logs: %w", err))
	}

	// 7. Записываем аудит-лог перед удалением пользователя
	if s.repo != nil {
		if _, err := s.repo.CreateAuditLog(ctx, "gdpr_deletion", "user", userID.String(),
			&userID, nil, map[string]interface{}{"deleted_user_id": userID.String()}, "", ""); err != nil {
			slog.Error("create gdpr audit_log", "error", err, "user_id", userID)
			cascadeErrs = append(cascadeErrs, fmt.Errorf("create audit_log: %w", err))
		}
	}

	// 8. Hard-delete пользователя
	if err := s.repo.HardDeleteUser(ctx, userID); err != nil {
		cascadeErrs = append(cascadeErrs, fmt.Errorf("удаление пользователя (GDPR): %w", err))
		return errors.Join(cascadeErrs...)
	}

	if len(cascadeErrs) > 0 {
		// Пользователь удалён, но часть связанных данных не очистилась —
		// возвращаем combined error, чтобы handler вернул 500 и админ
		// видел, что именно не удалось зачистить.
		slog.Warn("user deleted (GDPR) with cascade errors", "user_id", userID, "cascade_error_count", len(cascadeErrs))
		return errors.Join(cascadeErrs...)
	}

	slog.Info("user deleted (GDPR)", "user_id", userID)
	return nil
}

// daysRemaining вычисляет количество оставшихся дней до окончательного удаления.
func daysRemaining(deletedAt time.Time) int {
	elapsed := time.Since(deletedAt)
	remaining := TrashRetentionDays - int(math.Ceil(elapsed.Hours()/24))
	if remaining < 0 {
		return 0
	}
	return remaining
}
