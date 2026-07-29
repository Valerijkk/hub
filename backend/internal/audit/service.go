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
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/auth"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// ErrCannotDeleteFounder — GDPR-erase founder запрещён всегда (ORG-04 / P2-BE-19),
// включая self-erase: иначе org остаётся без founder (break-glass только через ADR).
var ErrCannotDeleteFounder = errors.New("cannot delete founder account")

// ErrCannotModifyFounder — A1-009: ban/deactivate founder запрещены (ORG-04).
var ErrCannotModifyFounder = errors.New("cannot ban or deactivate founder account")

// ErrBanReasonRequired — A3W6-4-002: BanUser без причины.
var ErrBanReasonRequired = errors.New("укажите причину блокировки")

// ErrCannotBanSelf — A3W6-4-002: admin не может забанить сам себя.
var ErrCannotBanSelf = errors.New("нельзя заблокировать самого себя")

// ErrUserNotFound — A3W6-4-003: пользователь отсутствует (IsSystemAdmin / admin gate).
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidResourceLimitKey — A1-013: PUT /admin/limits/{key} только allowlist.
var ErrInvalidResourceLimitKey = errors.New("invalid resource limit key")

// ErrInvalidResourceLimitValue — A2W6-A4-002: значение лимита — целое > 0.
var ErrInvalidResourceLimitValue = errors.New("invalid resource limit value")

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
// единый Repository-интерфейс. Это устраняет дубль db+repo и облегчает
// unit-тестирование (репозиторий моделируется в памяти без поднятия Ent).
//
// Phase-14: Repository теперь интерфейс (а не *struct), что позволяет
// подменять реализацию в тестах без cgo-зависимостей Ent.
type AuditService struct {
	repo       Repository
	redis      *redis.Client // optional — используется для инвалидации auth-кэшей (ME-13).
	msgEraser  MessageEraser // optional — Mongo messages by sender_id (P1-BE-34).
	sessions   SessionRevoker // optional — C-W19-001: Redis+NATS session kill on deactivate/ban.
	eraseRetry MongoEraseRetryEnqueuer // optional — C-W19-002: JetStream retry after post-Commit fail.
}

// MessageEraser удаляет персональные сообщения пользователя (Mongo messenger).
// Опциональная зависимость: если nil, шаг пропускается (тесты без Mongo).
type MessageEraser interface {
	DeleteMessagesBySender(ctx context.Context, senderID uuid.UUID) (int64, error)
}

// SessionRevoker — C-W19-001: DB+Redis invalidate + NATS revoked/logout → WS Bridge kill.
// Реализация: auth.SessionService.RevokeAllUserSessions (не дублировать pipeline).
type SessionRevoker interface {
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

// NewAuditService создаёт новый сервис аудита.
// rdb — опциональный Redis-клиент; если задан, AuditService инвалидирует
// auth.UserStatusCacheKey при деактивации пользователя (ME-13). Может быть
// nil в тестах — деактивация всё равно отработает, только статус будет
// применён по истечении TTL (~30s).
func NewAuditService(repo Repository, rdb *redis.Client) *AuditService {
	return &AuditService{repo: repo, redis: rdb}
}

// SetMessageEraser подключает стирание Mongo-сообщений при GDPR erase (P1-BE-34).
func (s *AuditService) SetMessageEraser(e MessageEraser) {
	s.msgEraser = e
}

// SetSessionRevoker подключает session-kill path (Redis + NATS revoked) для Deactivate/Ban.
// C-W19-001: без этого живой WS жертвы не рвётся. Wiring: wireAuth → authSvc.Sessions.
func (s *AuditService) SetSessionRevoker(r SessionRevoker) {
	s.sessions = r
}

// SetMongoEraseRetry подключает JetStream compensating retry для Mongo erase после GDPR Commit.
// C-W19-002: sync eraser fail → user уже HardDelete; retry по sender_id (идемпотентный DeleteMany).
func (s *AuditService) SetMongoEraseRetry(e MongoEraseRetryEnqueuer) {
	s.eraseRetry = e
}

// IsSystemAdmin возвращает true, если пользователь — системный администратор.
// Используется HTTP-обработчиками для проверки доступа к admin-эндпоинтам.
func (s *AuditService) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return s.repo.IsSystemAdmin(ctx, userID)
}

// ---- Журнал аудита ----

// ListAuditLogs возвращает пагинированный список записей аудита с фильтрами (только для администратора).
func (s *AuditService) ListAuditLogs(ctx context.Context, filters AuditLogFilters) (*AuditLogListResponse, error) {
	return observability.Timed(ctx, "audit.ListAuditLogs", func(ctx context.Context) (*AuditLogListResponse, error) {
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
	})
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

	// Один batch-запрос вместо N последовательных GetSystemSetting (fix N+1).
	values, err := s.repo.GetSystemSettingsBatch(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("получение лимитов ресурсов: %w", err)
	}

	result := make([]ResourceLimitResponse, 0, len(keys))
	for _, key := range keys {
		value := values[key]
		if value == "" {
			value = defaults[key]
		}
		result = append(result, ResourceLimitResponse{Key: key, Value: value})
	}

	return result, nil
}

// UpdateResourceLimit обновляет лимит ресурса в SystemSetting (только для администратора).
// A1-013: ключ обязан быть из allowlist — иначе можно перезаписать retention.* и др.
// A2W6-A4-002: value — положительное целое (Atoi + >0); иначе fail-closed.
func (s *AuditService) UpdateResourceLimit(ctx context.Context, key, value string) error {
	switch key {
	case LimitFileMaxSizeMB, LimitUserMaxProjects, LimitProjectMaxStorageMB:
		// ok
	default:
		return ErrInvalidResourceLimitKey
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return ErrInvalidResourceLimitValue
	}
	return s.repo.UpsertSystemSetting(ctx, key, value, fmt.Sprintf("Лимит ресурса: %s", key))
}

// defaultResourceLimits — канонические дефолты (совпадают с GetResourceLimits).
var defaultResourceLimits = map[string]int{
	LimitFileMaxSizeMB:       50,
	LimitUserMaxProjects:     20,
	LimitProjectMaxStorageMB: 5120,
}

// CheckLimit проверяет, не превышен ли лимит ресурса.
// A2W6-A4-002: пустой ключ без дефолта / невалидный Atoi / non-positive → deny (fail-closed).
// Известные ключи без записи в БД используют defaultResourceLimits.
func (s *AuditService) CheckLimit(ctx context.Context, limitKey string, currentValue int) (bool, error) {
	value, err := s.repo.GetSystemSetting(ctx, limitKey)
	if err != nil {
		return false, err
	}
	var limitVal int
	if value == "" {
		def, ok := defaultResourceLimits[limitKey]
		if !ok {
			return false, nil // неизвестный ключ без значения — deny
		}
		limitVal = def
	} else {
		limitVal, err = strconv.Atoi(value)
		if err != nil || limitVal <= 0 {
			return false, nil // мусор в БД — deny, не fail-open
		}
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
	case "newspost":
		items, total, err = s.repo.ListDeletedNewsPosts(ctx, cutoff, offset, perPage)
	case "scheduled_call":
		items, total, err = s.repo.ListDeletedScheduledCalls(ctx, projectID, cutoff, offset, perPage)
	case "crm_contact":
		items, total, err = s.repo.ListDeletedCRMContacts(ctx, cutoff, offset, perPage)
	case "crm_deal":
		items, total, err = s.repo.ListDeletedCRMDeals(ctx, cutoff, offset, perPage)
	case "hubbase":
		items, total, err = s.repo.ListDeletedHubBases(ctx, cutoff, offset, perPage)
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
	case "newspost":
		return s.repo.RestoreNewsPost(ctx, entityID)
	case "scheduled_call":
		return s.repo.RestoreScheduledCall(ctx, entityID)
	case "crm_contact":
		return s.repo.RestoreCRMContact(ctx, entityID)
	case "crm_deal":
		return s.repo.RestoreCRMDeal(ctx, entityID)
	case "hubbase":
		return s.repo.RestoreHubBase(ctx, entityID)
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
	case "newspost":
		return s.repo.SoftDeleteNewsPost(ctx, entityID, now)
	case "scheduled_call":
		return s.repo.SoftDeleteScheduledCall(ctx, entityID, now)
	case "crm_contact":
		return s.repo.SoftDeleteCRMContact(ctx, entityID, now)
	case "crm_deal":
		return s.repo.SoftDeleteCRMDeal(ctx, entityID, now)
	case "hubbase":
		return s.repo.SoftDeleteHubBase(ctx, entityID, now)
	default:
		return fmt.Errorf("неизвестный тип сущности: %s", entityType)
	}
}

// PurgeExpired окончательно удаляет элементы, находящиеся в корзине дольше 30 дней.
// Запускается по cron (горутина с 24-часовым тикером).
// A2W6-A4-001: ошибка любого purge_* пробрасывается наружу (fail-closed для cron),
// иначе tick trash-purge-expired считается успешным при полном отказе БД.
func (s *AuditService) PurgeExpired(ctx context.Context) error {
	cutoff := time.Now().Add(-TrashRetentionDays * 24 * time.Hour)

	var errs []error
	purge := func(entity string, deleteFunc func() (int, error)) {
		n, err := deleteFunc()
		if err != nil {
			slog.Error("purge expired "+entity, "error", err)
			errs = append(errs, fmt.Errorf("purge expired %s: %w", entity, err))
			return
		}
		if n > 0 {
			slog.Info("purged expired "+entity, "count", n)
		}
	}

	purge("issues", func() (int, error) { return s.repo.PurgeExpiredIssues(ctx, cutoff) })
	purge("test_cases", func() (int, error) { return s.repo.PurgeExpiredTestCases(ctx, cutoff) })
	purge("files", func() (int, error) { return s.repo.PurgeExpiredFiles(ctx, cutoff) })
	purge("projects", func() (int, error) { return s.repo.PurgeExpiredProjects(ctx, cutoff) })
	purge("newsposts", func() (int, error) { return s.repo.PurgeExpiredNewsPosts(ctx, cutoff) })
	purge("scheduled_calls", func() (int, error) { return s.repo.PurgeExpiredScheduledCalls(ctx, cutoff) })
	purge("crm_contacts", func() (int, error) { return s.repo.PurgeExpiredCRMContacts(ctx, cutoff) })
	purge("crm_deals", func() (int, error) { return s.repo.PurgeExpiredCRMDeals(ctx, cutoff) })
	purge("hubbases", func() (int, error) { return s.repo.PurgeExpiredHubBases(ctx, cutoff) })

	return errors.Join(errs...)
}

// ---- Управление пользователями ----

// DeactivateUser деактивирует пользователя и отзывает все его сессии.
func (s *AuditService) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	return s.deactivateUser(ctx, userID)
}

func (s *AuditService) deactivateUser(ctx context.Context, userID uuid.UUID) error {
	// A1-009 / ORG-04: founder нельзя деактивировать (в т.ч. через BanUser).
	role, err := s.repo.GetUserSystemRole(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user role (deactivate): %w", err)
	}
	if role == user.SystemRoleFounder {
		return ErrCannotModifyFounder
	}

	// Обновляем статус пользователя
	if err := s.repo.DeactivateUser(ctx, userID); err != nil {
		return fmt.Errorf("деактивация пользователя: %w", err)
	}

	// C-W19-001 / C-W19-003: session-kill fail-closed — Redis Del + NATS revoked
	// (тот же path, что Logout/Revoke). Без SessionRevoker — DB-only fallback для тестов.
	if s.sessions != nil {
		if err := s.sessions.RevokeAllUserSessions(ctx, userID); err != nil {
			slog.Error("revoke sessions on deactivate", "error", err, "user_id", userID)
			return fmt.Errorf("revoke sessions on deactivate: %w", err)
		}
	} else if _, err := s.repo.DeleteUserSessions(ctx, userID); err != nil {
		slog.Error("revoke sessions on deactivate", "error", err, "user_id", userID)
		return fmt.Errorf("revoke sessions on deactivate: %w", err)
	}

	// ME-13: инвалидируем закэшированный статус, чтобы запретить доступ
	// немедленно, а не через TTL окна (~30s).
	auth.InvalidateUserStatusCache(ctx, s.redis, userID)

	return nil
}

// BanUser блокирует пользователя с указанием причины и записывает событие в аудит.
func (s *AuditService) BanUser(ctx context.Context, targetUserID, adminUserID uuid.UUID, reason, ipAddress, userAgent string) error {
	return observability.TimedVoid(ctx, "audit.BanUser", func(ctx context.Context) error {
		if reason == "" {
			return ErrBanReasonRequired
		}
		if targetUserID == adminUserID {
			return ErrCannotBanSelf
		}

		if err := s.deactivateUser(ctx, targetUserID); err != nil {
			return err
		}

		details := map[string]interface{}{
			"reason":    reason,
			"banned_by": adminUserID.String(),
		}
		if _, err := s.repo.CreateAuditLog(ctx, "ban", "user", targetUserID.String(),
			&adminUserID, nil, details, ipAddress, userAgent); err != nil {
			// W6-A4-015 / P3-BE-01: ban без audit trail — fail-closed для caller.
			return fmt.Errorf("audit ban log: %w", err)
		}

		return nil
	})
}

// RecordHTTPRequest сохраняет HTTP-запрос в журнал аудита.
// В DEV/Docker логируются все API-запросы; ошибки и медленные запросы
// всегда попадают в журнал независимо от окружения.
func (s *AuditService) RecordHTTPRequest(ctx context.Context, entry HTTPAuditEntry) {
	_ = observability.TimedVoid(ctx, "audit.RecordHTTPRequest", func(ctx context.Context) error {
		action := "request"
		if entry.Status >= 400 {
			action = "request.error"
		}

		details := map[string]interface{}{
			"method":      entry.Method,
			"path":        entry.Path,
			"status":      entry.Status,
			"duration_ms": entry.DurationMS,
		}
		if entry.RequestID != "" {
			details["request_id"] = entry.RequestID
		}

		_, err := s.repo.CreateAuditLog(ctx, action, "http", entry.Path,
			entry.UserID, nil, details, entry.IPAddress, entry.UserAgent)
		if err != nil {
			slog.Error("audit http request", "error", err, "path", entry.Path)
			return err
		}
		return nil
	})
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
// 2. Удаляет комментарии/реакции/сессии/activity/time_logs
// 3. Удаляет файлы, memberships; scrub IP/UA в audit_log
// 4. (опционально) Mongo messages по sender_id
// 5. Hard-delete пользователя — только если весь каскад успешен
// 6. Записывает аудит-лог "GDPR deletion"
//
// A1-004 / ORG-04 / P2-BE-19: target=founder запрещён всегда (включая self-erase).
// Self-erase founder оставляет org без founder — break-glass только через ADR.
//
// A2W6-A4-003 / A3W6-4-001: все Ent-шаги (включая HardDelete) — одна транзакция;
// mid-cascade fail → Rollback. Mongo eraser — строго ПОСЛЕ Commit: иначе Mongo ok +
// Ent Rollback → user жив, chat-история уже стёрта (cross-store partial erasure).
// C-W19-002 / TASK-178: Mongo fail после Commit → enqueue JetStream retry (sender_id);
// orphan window до MaxDeliver, dual erase идемпотентен. Mongo НЕ внутрь WithTx.
func (s *AuditService) DeleteUserGDPR(ctx context.Context, actorID, userID uuid.UUID, reassignTo *uuid.UUID) error {
	role, err := s.repo.GetUserSystemRole(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user role (GDPR): %w", err)
	}
	if role == user.SystemRoleFounder {
		return ErrCannotDeleteFounder
	}

	if err := s.repo.WithTx(ctx, func(txRepo Repository) error {
		// 1. Переназначаем или очищаем задачи
		if reassignTo != nil {
			if _, err := txRepo.ReassignAssignedIssues(ctx, userID, *reassignTo); err != nil {
				return fmt.Errorf("переназначение задач (GDPR): %w", err)
			}
			if _, err := txRepo.ReassignReportedIssues(ctx, userID, *reassignTo); err != nil {
				return fmt.Errorf("переназначение reporter-задач (GDPR): %w", err)
			}
		} else {
			if _, err := txRepo.ClearUserAssignments(ctx, userID); err != nil {
				return fmt.Errorf("очистка задач пользователя (GDPR): %w", err)
			}
		}

		run := func(step string, fn func() error) error {
			if err := fn(); err != nil {
				slog.Error("GDPR cascade step failed", "step", step, "error", err, "user_id", userID)
				return fmt.Errorf("%s: %w", step, err)
			}
			return nil
		}

		if err := run("delete comments", func() error {
			_, e := txRepo.DeleteUserComments(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete reactions", func() error {
			_, e := txRepo.DeleteUserReactions(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete sessions", func() error {
			_, e := txRepo.DeleteUserSessions(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete activity_logs", func() error {
			_, e := txRepo.DeleteUserActivityLogs(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete time_logs", func() error {
			_, e := txRepo.DeleteUserTimeLogs(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete files", func() error {
			_, e := txRepo.DeleteUserFiles(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete project memberships", func() error {
			_, e := txRepo.DeleteUserProjectMemberships(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete channel memberships", func() error {
			_, e := txRepo.DeleteUserChannelMemberships(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("delete team memberships", func() error {
			_, e := txRepo.DeleteUserTeamMemberships(ctx, userID)
			return e
		}); err != nil {
			return err
		}
		if err := run("scrub audit PII", func() error {
			_, e := txRepo.ScrubUserAuditPII(ctx, userID)
			return e
		}); err != nil {
			return err
		}

		// Аудит до HardDelete — actor в details (accountability).
		if _, err := txRepo.CreateAuditLog(ctx, "gdpr_deletion", "user", userID.String(),
			&actorID, nil, map[string]interface{}{
				"deleted_user_id": userID.String(),
				"deleted_by":      actorID.String(),
			}, "", ""); err != nil {
			slog.Error("create gdpr audit_log", "error", err, "user_id", userID)
			return fmt.Errorf("create audit_log: %w", err)
		}

		if err := txRepo.HardDeleteUser(ctx, userID); err != nil {
			return fmt.Errorf("удаление пользователя (GDPR): %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	// A3W6-4-001: Mongo только после успешного Ent Commit.
	if s.msgEraser != nil {
		if _, err := s.msgEraser.DeleteMessagesBySender(ctx, userID); err != nil {
			slog.Error("GDPR mongo erase after commit failed",
				"error", err, "user_id", userID, "actor_id", actorID)
			return fmt.Errorf("delete mongo messages: %w", err)
		}
	}

	slog.Info("user deleted (GDPR)", "user_id", userID, "actor_id", actorID)
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
