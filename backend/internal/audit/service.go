// Package audit реализует журнал аудита и политики хранения данных.
package audit

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/activitylog"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/systemsetting"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
)

// Настройки аудита, лимитов и корзины.
const (
	// Ключи лимитов в SystemSetting.
	LimitFileMaxSizeMB      = "limit.file.max_size_mb"
	LimitUserMaxProjects    = "limit.user.max_projects"
	LimitProjectMaxStorageMB = "limit.project.max_storage_mb"

	// TrashRetentionDays — период хранения в корзине (дней).
	TrashRetentionDays = 30
)

// AuditService содержит бизнес-логику для аудита, лимитов, корзины и управления пользователями.
type AuditService struct {
	db   *ent.Client
	repo *Repository
}

// NewAuditService создаёт новый сервис аудита.
func NewAuditService(db *ent.Client, repo *Repository) *AuditService {
	return &AuditService{db: db, repo: repo}
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
		setting, err := s.db.SystemSetting.Query().
			Where(systemsetting.KeyEQ(key)).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				result = append(result, ResourceLimitResponse{Key: key, Value: defaults[key]})
				continue
			}
			return nil, fmt.Errorf("получение лимита %s: %w", key, err)
		}
		result = append(result, ResourceLimitResponse{Key: key, Value: setting.Value})
	}

	return result, nil
}

// UpdateResourceLimit обновляет лимит ресурса в SystemSetting (только для администратора).
func (s *AuditService) UpdateResourceLimit(ctx context.Context, key, value string) error {
	n, err := s.db.SystemSetting.Update().
		Where(systemsetting.KeyEQ(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("обновление лимита %s: %w", key, err)
	}
	if n == 0 {
		_, err = s.db.SystemSetting.Create().
			SetKey(key).
			SetValue(value).
			SetDescription(fmt.Sprintf("Лимит ресурса: %s", key)).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("создание лимита %s: %w", key, err)
		}
	}
	return nil
}

// CheckLimit проверяет, не превышен ли лимит ресурса.
func (s *AuditService) CheckLimit(ctx context.Context, limitKey string, currentValue int) (bool, error) {
	setting, err := s.db.SystemSetting.Query().
		Where(systemsetting.KeyEQ(limitKey)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return true, nil // Лимит не задан — разрешаем
		}
		return false, err
	}

	limitVal, err := strconv.Atoi(setting.Value)
	if err != nil {
		return true, nil // Невалидное значение — разрешаем
	}

	return currentValue < limitVal, nil
}

// ---- Корзина (мягкое удаление) ----

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

	switch entityType {
	case "issue":
		query := s.db.Issue.Query().
			Where(issue.DeletedAtNotNil(), issue.DeletedAtGT(cutoff))
		if projectID != nil {
			query = query.Where(issue.HasProjectWith(project.IDEQ(*projectID)))
		}
		total, err := query.Clone().Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("подсчёт удалённых задач: %w", err)
		}
		resp.Total = total

		items, err := query.
			Order(ent.Desc(issue.FieldDeletedAt)).
			Offset(offset).
			Limit(perPage).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("список удалённых задач: %w", err)
		}
		for _, item := range items {
			resp.Items = append(resp.Items, TrashItemResponse{
				ID:            item.ID.String(),
				Name:          item.Title,
				EntityType:    "issue",
				DeletedAt:     *item.DeletedAt,
				DaysRemaining: daysRemaining(*item.DeletedAt),
			})
		}

	case "test_case":
		query := s.db.TestCase.Query().
			Where(testcase.DeletedAtNotNil(), testcase.DeletedAtGT(cutoff))
		if projectID != nil {
			query = query.Where(testcase.HasProjectWith(project.IDEQ(*projectID)))
		}
		total, err := query.Clone().Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("подсчёт удалённых тест-кейсов: %w", err)
		}
		resp.Total = total

		items, err := query.
			Order(ent.Desc(testcase.FieldDeletedAt)).
			Offset(offset).
			Limit(perPage).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("список удалённых тест-кейсов: %w", err)
		}
		for _, item := range items {
			resp.Items = append(resp.Items, TrashItemResponse{
				ID:            item.ID.String(),
				Name:          item.Title,
				EntityType:    "test_case",
				DeletedAt:     *item.DeletedAt,
				DaysRemaining: daysRemaining(*item.DeletedAt),
			})
		}

	case "file":
		query := s.db.FileMetadata.Query().
			Where(filemetadata.DeletedAtNotNil(), filemetadata.DeletedAtGT(cutoff))
		if projectID != nil {
			query = query.Where(filemetadata.HasProjectWith(project.IDEQ(*projectID)))
		}
		total, err := query.Clone().Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("подсчёт удалённых файлов: %w", err)
		}
		resp.Total = total

		items, err := query.
			Order(ent.Desc(filemetadata.FieldDeletedAt)).
			Offset(offset).
			Limit(perPage).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("список удалённых файлов: %w", err)
		}
		for _, item := range items {
			resp.Items = append(resp.Items, TrashItemResponse{
				ID:            item.ID.String(),
				Name:          item.Filename,
				EntityType:    "file",
				DeletedAt:     *item.DeletedAt,
				DaysRemaining: daysRemaining(*item.DeletedAt),
			})
		}

	case "project":
		query := s.db.Project.Query().
			Where(project.DeletedAtNotNil(), project.DeletedAtGT(cutoff))
		total, err := query.Clone().Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("подсчёт удалённых проектов: %w", err)
		}
		resp.Total = total

		items, err := query.
			Order(ent.Desc(project.FieldDeletedAt)).
			Offset(offset).
			Limit(perPage).
			All(ctx)
		if err != nil {
			return nil, fmt.Errorf("список удалённых проектов: %w", err)
		}
		for _, item := range items {
			resp.Items = append(resp.Items, TrashItemResponse{
				ID:            item.ID.String(),
				Name:          item.Name,
				EntityType:    "project",
				DeletedAt:     *item.DeletedAt,
				DaysRemaining: daysRemaining(*item.DeletedAt),
			})
		}
	}

	return resp, nil
}

// RestoreItem восстанавливает элемент из корзины (сбрасывает deleted_at).
func (s *AuditService) RestoreItem(ctx context.Context, entityType string, entityID uuid.UUID) error {
	switch entityType {
	case "issue":
		return s.db.Issue.UpdateOneID(entityID).ClearDeletedAt().Exec(ctx)
	case "test_case":
		return s.db.TestCase.UpdateOneID(entityID).ClearDeletedAt().Exec(ctx)
	case "file":
		return s.db.FileMetadata.UpdateOneID(entityID).ClearDeletedAt().Exec(ctx)
	case "project":
		return s.db.Project.UpdateOneID(entityID).ClearDeletedAt().SetStatus(project.StatusActive).Exec(ctx)
	default:
		return fmt.Errorf("неизвестный тип сущности: %s", entityType)
	}
}

// SoftDelete устанавливает deleted_at = now для указанной сущности.
func (s *AuditService) SoftDelete(ctx context.Context, entityType string, entityID uuid.UUID) error {
	now := time.Now()
	switch entityType {
	case "issue":
		return s.db.Issue.UpdateOneID(entityID).SetDeletedAt(now).Exec(ctx)
	case "test_case":
		return s.db.TestCase.UpdateOneID(entityID).SetDeletedAt(now).Exec(ctx)
	case "file":
		return s.db.FileMetadata.UpdateOneID(entityID).SetDeletedAt(now).Exec(ctx)
	case "project":
		return s.db.Project.UpdateOneID(entityID).SetDeletedAt(now).SetStatus(project.StatusDeleted).Exec(ctx)
	default:
		return fmt.Errorf("неизвестный тип сущности: %s", entityType)
	}
}

// PurgeExpired окончательно удаляет элементы, находящиеся в корзине дольше 30 дней.
// Запускается по cron (горутина с 24-часовым тикером).
func (s *AuditService) PurgeExpired(ctx context.Context) error {
	cutoff := time.Now().Add(-TrashRetentionDays * 24 * time.Hour)

	n, err := s.db.Issue.Delete().Where(issue.DeletedAtNotNil(), issue.DeletedAtLT(cutoff)).Exec(ctx)
	if err != nil {
		slog.Error("purge expired issues", "error", err)
	} else if n > 0 {
		slog.Info("purged expired issues", "count", n)
	}

	n, err = s.db.TestCase.Delete().Where(testcase.DeletedAtNotNil(), testcase.DeletedAtLT(cutoff)).Exec(ctx)
	if err != nil {
		slog.Error("purge expired test cases", "error", err)
	} else if n > 0 {
		slog.Info("purged expired test cases", "count", n)
	}

	n, err = s.db.FileMetadata.Delete().Where(filemetadata.DeletedAtNotNil(), filemetadata.DeletedAtLT(cutoff)).Exec(ctx)
	if err != nil {
		slog.Error("purge expired files", "error", err)
	} else if n > 0 {
		slog.Info("purged expired files", "count", n)
	}

	n, err = s.db.Project.Delete().Where(project.DeletedAtNotNil(), project.DeletedAtLT(cutoff)).Exec(ctx)
	if err != nil {
		slog.Error("purge expired projects", "error", err)
	} else if n > 0 {
		slog.Info("purged expired projects", "count", n)
	}

	return nil
}

// ---- Управление пользователями ----

// DeactivateUser деактивирует пользователя и отзывает все его сессии.
func (s *AuditService) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	// Обновляем статус пользователя
	err := s.db.User.UpdateOneID(userID).
		SetStatus(user.StatusDeactivated).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("деактивация пользователя: %w", err)
	}

	// Отзываем все сессии
	_, err = s.db.Session.Delete().
		Where(session.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
	if err != nil {
		slog.Error("revoke sessions on deactivate", "error", err, "user_id", userID)
	}

	return nil
}

// ReassignTasks переназначает все задачи одного пользователя другому.
func (s *AuditService) ReassignTasks(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error) {
	n, err := s.db.Issue.Update().
		Where(issue.HasAssigneeWith(user.IDEQ(fromUserID))).
		SetAssigneeID(toUserID).
		Save(ctx)
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
func (s *AuditService) DeleteUserGDPR(ctx context.Context, userID uuid.UUID, reassignTo *uuid.UUID) error {
	// 1. Переназначаем или очищаем задачи
	if reassignTo != nil {
		_, err := s.db.Issue.Update().
			Where(issue.HasAssigneeWith(user.IDEQ(userID))).
			SetAssigneeID(*reassignTo).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("переназначение задач (GDPR): %w", err)
		}
		// Также переназначаем задачи где пользователь reporter
		_, _ = s.db.Issue.Update().
			Where(issue.HasReporterWith(user.IDEQ(userID))).
			SetReporterID(*reassignTo).
			Save(ctx)
	} else {
		_, err := s.db.Issue.Update().
			Where(issue.HasAssigneeWith(user.IDEQ(userID))).
			ClearAssignee().
			Save(ctx)
		if err != nil {
			return fmt.Errorf("очистка задач пользователя (GDPR): %w", err)
		}
	}

	// 2. Удаляем комментарии пользователя
	n, err := s.db.Comment.Delete().
		Where(comment.HasAuthorWith(user.IDEQ(userID))).
		Exec(ctx)
	if err != nil {
		slog.Error("delete comments (GDPR)", "error", err, "user_id", userID)
	} else if n > 0 {
		slog.Info("deleted user comments (GDPR)", "count", n, "user_id", userID)
	}

	// 3. Удаляем реакции пользователя
	_, _ = s.db.Reaction.Delete().
		Where(reaction.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)

	// 4. Удаляем все сессии
	_, err = s.db.Session.Delete().
		Where(session.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
	if err != nil {
		slog.Error("delete sessions (GDPR)", "error", err, "user_id", userID)
	}

	// 5. Удаляем логи активности пользователя
	_, _ = s.db.ActivityLog.Delete().
		Where(activitylog.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)

	// 6. Удаляем записи времени пользователя
	_, _ = s.db.TimeLog.Delete().
		Where(timelog.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)

	// 7. Записываем аудит-лог перед удалением пользователя
	if s.repo != nil {
		_, _ = s.repo.CreateAuditLog(ctx, "gdpr_deletion", "user", userID.String(),
			&userID, nil, map[string]interface{}{"deleted_user_id": userID.String()}, "", "")
	}

	// 8. Hard-delete пользователя
	err = s.db.User.DeleteOneID(userID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("удаление пользователя (GDPR): %w", err)
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
