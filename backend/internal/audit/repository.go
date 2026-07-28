package audit

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/activitylog"
	"github.com/valerii/hub/backend/ent/auditlog"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/newspost"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/systemsetting"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository — контракт доступа к данным журнала аудита, лимитов, корзины и
// управления пользователями. Конкретная реализация поверх Ent ORM (entRepository)
// удовлетворяет этот интерфейс; тесты могут подставлять in-memory mock без
// поднятия Ent.
type Repository interface {
	// Журнал аудита
	CreateAuditLog(ctx context.Context, action, entityType, entityID string, userID, projectID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) (*ent.AuditLog, error)
	ListAuditLogs(ctx context.Context, filters AuditLogFilters) ([]*ent.AuditLog, int, error)

	// Системные настройки
	GetSystemSetting(ctx context.Context, key string) (string, error)
	GetSystemSettingsBatch(ctx context.Context, keys []string) (map[string]string, error)
	UpsertSystemSetting(ctx context.Context, key, value, description string) error

	// Корзина — листинг
	ListDeletedIssues(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)
	ListDeletedTestCases(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)
	ListDeletedFiles(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)
	ListDeletedProjects(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)
	ListDeletedNewsPosts(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)

	// Корзина — восстановление
	RestoreIssue(ctx context.Context, id uuid.UUID) error
	RestoreTestCase(ctx context.Context, id uuid.UUID) error
	RestoreFile(ctx context.Context, id uuid.UUID) error
	RestoreProject(ctx context.Context, id uuid.UUID) error
	RestoreNewsPost(ctx context.Context, id uuid.UUID) error

	// Корзина — мягкое удаление
	SoftDeleteIssue(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteTestCase(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteFile(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteProject(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteNewsPost(ctx context.Context, id uuid.UUID, when time.Time) error

	// Корзина — окончательное удаление
	PurgeExpiredIssues(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredTestCases(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredFiles(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredProjects(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredNewsPosts(ctx context.Context, cutoff time.Time) (int, error)

	// Администрирование пользователей
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) (int, error)
	ReassignAssignedIssues(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error)
	ReassignReportedIssues(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error)
	ClearUserAssignments(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserComments(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserReactions(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserActivityLogs(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserTimeLogs(ctx context.Context, userID uuid.UUID) (int, error)
	HardDeleteUser(ctx context.Context, userID uuid.UUID) error
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	// GetUserSystemRole — для GDPR founder-guard (A1-004).
	GetUserSystemRole(ctx context.Context, userID uuid.UUID) (user.SystemRole, error)
}

const slowRequestThresholdMS = 500

// entRepository — Ent-реализация Repository.
type entRepository struct {
	client *ent.Client
}

// NewRepository создаёт новый репозиторий журнала аудита поверх Ent.
func NewRepository(client *ent.Client) Repository {
	return &entRepository{client: client}
}

// CreateAuditLog записывает новую запись в журнал аудита.
func (r *entRepository) CreateAuditLog(ctx context.Context, action, entityType, entityID string, userID, projectID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) (*ent.AuditLog, error) {
	create := r.client.AuditLog.Create().
		SetAction(action).
		SetEntityType(entityType)

	if entityID != "" {
		create = create.SetEntityID(entityID)
	}
	if userID != nil {
		create = create.SetUserID(*userID)
	}
	if projectID != nil {
		create = create.SetProjectID(*projectID)
	}
	if details != nil {
		create = create.SetDetails(details)
	}
	if ipAddress != "" {
		create = create.SetIPAddress(ipAddress)
	}
	if userAgent != "" {
		create = create.SetUserAgent(userAgent)
	}

	return create.Save(ctx)
}

// ListAuditLogs возвращает пагинированный список записей аудита с фильтрами.
func (r *entRepository) ListAuditLogs(ctx context.Context, filters AuditLogFilters) ([]*ent.AuditLog, int, error) {
	query := r.client.AuditLog.Query()

	if filters.Action != "" {
		query = query.Where(auditlog.ActionEQ(filters.Action))
	}
	if filters.EntityType != "" {
		query = query.Where(auditlog.EntityTypeEQ(filters.EntityType))
	}
	if filters.UserID != nil {
		query = query.Where(auditlog.UserIDEQ(*filters.UserID))
	}
	if filters.ProjectID != nil {
		query = query.Where(auditlog.ProjectIDEQ(*filters.ProjectID))
	}
	if filters.EntityID != "" {
		query = query.Where(auditlog.EntityIDEQ(filters.EntityID))
	}
	if filters.IPAddress != "" {
		query = query.Where(auditlog.IPAddressContains(filters.IPAddress))
	}
	if filters.DateFrom != nil {
		query = query.Where(auditlog.CreatedAtGTE(*filters.DateFrom))
	}
	if filters.DateTo != nil {
		query = query.Where(auditlog.CreatedAtLTE(*filters.DateTo))
	}
	if filters.SlowOnly {
		query = query.Where(auditlog.EntityTypeEQ("http"))
		query = query.Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"details IS NOT NULL AND (details->>'duration_ms')::float > $1",
				slowRequestThresholdMS,
			))
		})
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт записей аудита: %w", err)
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	perPage := filters.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage

	items, err := query.
		Order(ent.Desc(auditlog.FieldCreatedAt)).
		Offset(offset).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список записей аудита: %w", err)
	}

	return items, total, nil
}

// ---------------------------------------------------------------------------
// SystemSetting — лимиты ресурсов
// ---------------------------------------------------------------------------

// GetSystemSetting возвращает значение настройки по ключу.
// Возвращает ("", nil) если запись отсутствует — вызывающий код сам выбирает дефолт.
func (r *entRepository) GetSystemSetting(ctx context.Context, key string) (string, error) {
	setting, err := r.client.SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return setting.Value, nil
}

// GetSystemSettingsBatch возвращает значения нескольких настроек одним запросом.
// Возвращает map[key]value; отсутствующие ключи не включаются в результат.
func (r *entRepository) GetSystemSettingsBatch(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}
	settings, err := r.client.SystemSetting.Query().
		Where(systemsetting.KeyIn(keys...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

// UpsertSystemSetting создаёт или обновляет настройку.
func (r *entRepository) UpsertSystemSetting(ctx context.Context, key, value, description string) error {
	n, err := r.client.SystemSetting.Update().
		Where(systemsetting.KeyEQ(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("обновление настройки %s: %w", key, err)
	}
	if n == 0 {
		_, err = r.client.SystemSetting.Create().
			SetKey(key).
			SetValue(value).
			SetDescription(description).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("создание настройки %s: %w", key, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Trash — мягкое удаление и восстановление
// ---------------------------------------------------------------------------

// TrashItem — нормализованное представление элемента в корзине.
type TrashItem struct {
	ID         uuid.UUID
	Name       string
	EntityType string
	DeletedAt  time.Time
}

// ListDeletedIssues возвращает количество и срез удалённых задач (после cutoff).
func (r *entRepository) ListDeletedIssues(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.Issue.Query().
		Where(issue.DeletedAtNotNil(), issue.DeletedAtGT(cutoff))
	if projectID != nil {
		query = query.Where(issue.HasProjectWith(project.IDEQ(*projectID)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых задач: %w", err)
	}
	items, err := query.
		Order(ent.Desc(issue.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых задач: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Title,
			EntityType: "issue",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// ListDeletedTestCases возвращает количество и срез удалённых тест-кейсов.
func (r *entRepository) ListDeletedTestCases(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.TestCase.Query().
		Where(testcase.DeletedAtNotNil(), testcase.DeletedAtGT(cutoff))
	if projectID != nil {
		query = query.Where(testcase.HasProjectWith(project.IDEQ(*projectID)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых тест-кейсов: %w", err)
	}
	items, err := query.
		Order(ent.Desc(testcase.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых тест-кейсов: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Title,
			EntityType: "test_case",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// ListDeletedFiles возвращает количество и срез удалённых файлов.
func (r *entRepository) ListDeletedFiles(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.FileMetadata.Query().
		Where(filemetadata.DeletedAtNotNil(), filemetadata.DeletedAtGT(cutoff))
	if projectID != nil {
		query = query.Where(filemetadata.HasProjectWith(project.IDEQ(*projectID)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых файлов: %w", err)
	}
	items, err := query.
		Order(ent.Desc(filemetadata.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых файлов: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Filename,
			EntityType: "file",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// ListDeletedProjects возвращает количество и срез удалённых проектов.
func (r *entRepository) ListDeletedProjects(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.Project.Query().
		Where(project.DeletedAtNotNil(), project.DeletedAtGT(cutoff))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых проектов: %w", err)
	}
	items, err := query.
		Order(ent.Desc(project.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых проектов: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Name,
			EntityType: "project",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// RestoreIssue сбрасывает deleted_at у задачи.
func (r *entRepository) RestoreIssue(ctx context.Context, id uuid.UUID) error {
	return r.client.Issue.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
}

// RestoreTestCase сбрасывает deleted_at у тест-кейса.
func (r *entRepository) RestoreTestCase(ctx context.Context, id uuid.UUID) error {
	return r.client.TestCase.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
}

// RestoreFile сбрасывает deleted_at у файла.
func (r *entRepository) RestoreFile(ctx context.Context, id uuid.UUID) error {
	return r.client.FileMetadata.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
}

// RestoreProject сбрасывает deleted_at у проекта и переводит status=active.
func (r *entRepository) RestoreProject(ctx context.Context, id uuid.UUID) error {
	return r.client.Project.UpdateOneID(id).ClearDeletedAt().SetStatus(project.StatusActive).Exec(ctx)
}

// SoftDeleteIssue ставит deleted_at=now у задачи.
func (r *entRepository) SoftDeleteIssue(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.Issue.UpdateOneID(id).SetDeletedAt(when).Exec(ctx)
}

// SoftDeleteTestCase ставит deleted_at=now у тест-кейса.
func (r *entRepository) SoftDeleteTestCase(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.TestCase.UpdateOneID(id).SetDeletedAt(when).Exec(ctx)
}

// SoftDeleteFile ставит deleted_at=now у файла.
func (r *entRepository) SoftDeleteFile(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.FileMetadata.UpdateOneID(id).SetDeletedAt(when).Exec(ctx)
}

// SoftDeleteProject ставит deleted_at=now и status=deleted у проекта.
func (r *entRepository) SoftDeleteProject(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.Project.UpdateOneID(id).SetDeletedAt(when).SetStatus(project.StatusDeleted).Exec(ctx)
}

// PurgeExpiredIssues окончательно удаляет задачи, которые были помечены deleted_at <= cutoff.
func (r *entRepository) PurgeExpiredIssues(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.Issue.Delete().Where(issue.DeletedAtNotNil(), issue.DeletedAtLT(cutoff)).Exec(ctx)
}

// PurgeExpiredTestCases окончательно удаляет просроченные тест-кейсы.
func (r *entRepository) PurgeExpiredTestCases(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.TestCase.Delete().Where(testcase.DeletedAtNotNil(), testcase.DeletedAtLT(cutoff)).Exec(ctx)
}

// PurgeExpiredFiles окончательно удаляет просроченные файлы.
func (r *entRepository) PurgeExpiredFiles(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.FileMetadata.Delete().Where(filemetadata.DeletedAtNotNil(), filemetadata.DeletedAtLT(cutoff)).Exec(ctx)
}

// PurgeExpiredProjects окончательно удаляет просроченные проекты.
func (r *entRepository) PurgeExpiredProjects(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.Project.Delete().Where(project.DeletedAtNotNil(), project.DeletedAtLT(cutoff)).Exec(ctx)
}

// ListDeletedNewsPosts возвращает мягко удалённые новости в окне retention (NFR-11).
// Org-level: project_id не фильтруется (как у project).
func (r *entRepository) ListDeletedNewsPosts(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.NewsPost.Query().
		Where(newspost.DeletedAtNotNil(), newspost.DeletedAtGT(cutoff))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых новостей: %w", err)
	}
	items, err := query.
		Order(ent.Desc(newspost.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых новостей: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Title,
			EntityType: "newspost",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// RestoreNewsPost сбрасывает deleted_at у новости (FR-INTR-08).
// Связанный опрос/голоса (когда появятся news_poll*) скрываются через soft-delete поста
// и снова видны после ClearDeletedAt без отдельной операции.
func (r *entRepository) RestoreNewsPost(ctx context.Context, id uuid.UUID) error {
	return r.client.NewsPost.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
}

// SoftDeleteNewsPost ставит deleted_at у новости (админ-путь корзины).
func (r *entRepository) SoftDeleteNewsPost(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.NewsPost.UpdateOneID(id).SetDeletedAt(when).Exec(ctx)
}

// PurgeExpiredNewsPosts окончательно удаляет новости старше retention.
func (r *entRepository) PurgeExpiredNewsPosts(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.NewsPost.Delete().Where(newspost.DeletedAtNotNil(), newspost.DeletedAtLT(cutoff)).Exec(ctx)
}

// ---------------------------------------------------------------------------
// User admin — деактивация, GDPR-удаление, переназначение задач
// ---------------------------------------------------------------------------

// DeactivateUser помечает пользователя статусом deactivated.
func (r *entRepository) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	return r.client.User.UpdateOneID(userID).
		SetStatus(user.StatusDeactivated).
		Exec(ctx)
}

// DeleteUserSessions удаляет все сессии пользователя; возвращает количество удалённых.
func (r *entRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.Session.Delete().
		Where(session.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
}

// ReassignAssignedIssues переназначает все задачи, где fromUserID = assignee, на toUserID.
func (r *entRepository) ReassignAssignedIssues(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error) {
	return r.client.Issue.Update().
		Where(issue.HasAssigneeWith(user.IDEQ(fromUserID))).
		SetAssigneeID(toUserID).
		Save(ctx)
}

// ReassignReportedIssues переназначает все задачи, где fromUserID = reporter, на toUserID.
func (r *entRepository) ReassignReportedIssues(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error) {
	return r.client.Issue.Update().
		Where(issue.HasReporterWith(user.IDEQ(fromUserID))).
		SetReporterID(toUserID).
		Save(ctx)
}

// ClearUserAssignments очищает поле assignee во всех задачах, где user = assignee.
func (r *entRepository) ClearUserAssignments(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.Issue.Update().
		Where(issue.HasAssigneeWith(user.IDEQ(userID))).
		ClearAssignee().
		Save(ctx)
}

// DeleteUserComments удаляет все комментарии пользователя.
func (r *entRepository) DeleteUserComments(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.Comment.Delete().
		Where(comment.HasAuthorWith(user.IDEQ(userID))).
		Exec(ctx)
}

// DeleteUserReactions удаляет все реакции пользователя.
func (r *entRepository) DeleteUserReactions(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.Reaction.Delete().
		Where(reaction.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
}

// DeleteUserActivityLogs удаляет все логи активности пользователя.
func (r *entRepository) DeleteUserActivityLogs(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.ActivityLog.Delete().
		Where(activitylog.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
}

// DeleteUserTimeLogs удаляет все записи времени пользователя.
func (r *entRepository) DeleteUserTimeLogs(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.TimeLog.Delete().
		Where(timelog.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
}

// HardDeleteUser окончательно удаляет запись пользователя.
func (r *entRepository) HardDeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.client.User.DeleteOneID(userID).Exec(ctx)
}

// IsSystemAdmin возвращает true, если у пользователя установлена системная роль admin.
// Используется audit handler.requireAdmin для проверки доступа к admin-эндпоинтам.
func (r *entRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.SystemRole == user.SystemRoleFounder || u.SystemRole == user.SystemRoleAdmin, nil
}

// GetUserSystemRole возвращает system_role пользователя.
func (r *entRepository) GetUserSystemRole(ctx context.Context, userID uuid.UUID) (user.SystemRole, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return "", err
	}
	return u.SystemRole, nil
}
