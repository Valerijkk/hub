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
	"github.com/valerii/hub/backend/ent/callparticipantconnection"
	"github.com/valerii/hub/backend/ent/callparticipantpresence"
	"github.com/valerii/hub/backend/ent/callproviderdelivery"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/crmcontact"
	"github.com/valerii/hub/backend/ent/crmdeal"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/hubbase"
	"github.com/valerii/hub/backend/ent/hubcolumn"
	"github.com/valerii/hub/backend/ent/hubrecord"
	"github.com/valerii/hub/backend/ent/hubtable"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/newspost"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/systemsetting"
	"github.com/valerii/hub/backend/ent/teammember"
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
	ListDeletedScheduledCalls(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)
	// W7-A2-006: CRM soft-delete → единая admin trash.
	ListDeletedCRMContacts(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)
	ListDeletedCRMDeals(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)
	// A3W2-A5-001: HubBase soft-delete → единая admin trash (type=hubbase).
	ListDeletedHubBases(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error)

	// Корзина — восстановление
	RestoreIssue(ctx context.Context, id uuid.UUID) error
	RestoreTestCase(ctx context.Context, id uuid.UUID) error
	RestoreFile(ctx context.Context, id uuid.UUID) error
	RestoreProject(ctx context.Context, id uuid.UUID) error
	RestoreNewsPost(ctx context.Context, id uuid.UUID) error
	RestoreScheduledCall(ctx context.Context, id uuid.UUID) error
	RestoreCRMContact(ctx context.Context, id uuid.UUID) error
	RestoreCRMDeal(ctx context.Context, id uuid.UUID) error
	RestoreHubBase(ctx context.Context, id uuid.UUID) error

	// Корзина — мягкое удаление
	SoftDeleteIssue(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteTestCase(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteFile(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteProject(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteNewsPost(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteScheduledCall(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteCRMContact(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteCRMDeal(ctx context.Context, id uuid.UUID, when time.Time) error
	SoftDeleteHubBase(ctx context.Context, id uuid.UUID, when time.Time) error

	// Корзина — окончательное удаление
	PurgeExpiredIssues(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredTestCases(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredFiles(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredProjects(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredNewsPosts(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredScheduledCalls(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredCRMContacts(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredCRMDeals(ctx context.Context, cutoff time.Time) (int, error)
	PurgeExpiredHubBases(ctx context.Context, cutoff time.Time) (int, error)

	// Администрирование пользователей
	ListAdminUsers(ctx context.Context) ([]*ent.User, error)
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) (int, error)
	ReassignAssignedIssues(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error)
	ReassignReportedIssues(ctx context.Context, fromUserID, toUserID uuid.UUID) (int, error)
	ClearUserAssignments(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserComments(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserReactions(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserActivityLogs(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserTimeLogs(ctx context.Context, userID uuid.UUID) (int, error)
	// P1-BE-34 / W6-A4-001: расширенный GDPR-каскад.
	DeleteUserFiles(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserProjectMemberships(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserChannelMemberships(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteUserTeamMemberships(ctx context.Context, userID uuid.UUID) (int, error)
	ScrubUserAuditPII(ctx context.Context, userID uuid.UUID) (int, error)
	HardDeleteUser(ctx context.Context, userID uuid.UUID) error
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	// GetUserSystemRole — для GDPR founder-guard (A1-004).
	GetUserSystemRole(ctx context.Context, userID uuid.UUID) (user.SystemRole, error)
	LockUserForGDPR(ctx context.Context, userID uuid.UUID) (user.SystemRole, error)
	EraseUserDurableCallData(ctx context.Context, userID uuid.UUID, terminalAt time.Time) error

	// WithTx выполняет fn в одной Ent-транзакции (A2W6-A4-003 GDPR cascade).
	// txRepo разделяет те же методы Repository, но пишет через tx.Client().
	WithTx(ctx context.Context, fn func(txRepo Repository) error) error
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

// WithTx выполняет fn внутри Ent-транзакции. Ошибка fn → Rollback.
func (r *entRepository) WithTx(ctx context.Context, fn func(txRepo Repository) error) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin audit tx: %w", err)
	}
	txRepo := &entRepository{client: tx.Client()}
	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w (rollback: %v)", err, rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit audit tx: %w", err)
	}
	return nil
}

// ListAdminUsers возвращает только поля, необходимые панели администратора.
// Select не позволяет случайно протащить в HTTP-слой дополнительные поля профиля.
func (r *entRepository) ListAdminUsers(ctx context.Context) ([]*ent.User, error) {
	return r.client.User.Query().
		Select(
			user.FieldID,
			user.FieldEmail,
			user.FieldName,
			user.FieldDisplayName,
			user.FieldSystemRole,
			user.FieldStatus,
			user.FieldCreatedAt,
		).
		All(ctx)
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
// W6-A4-012: при concurrent Create после Update n==0 — retry Update на unique conflict.
func (r *entRepository) UpsertSystemSetting(ctx context.Context, key, value, description string) error {
	n, err := r.client.SystemSetting.Update().
		Where(systemsetting.KeyEQ(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("обновление настройки %s: %w", key, err)
	}
	if n > 0 {
		return nil
	}
	_, err = r.client.SystemSetting.Create().
		SetKey(key).
		SetValue(value).
		SetDescription(description).
		Save(ctx)
	if err == nil {
		return nil
	}
	if ent.IsConstraintError(err) {
		if _, retryErr := r.client.SystemSetting.Update().
			Where(systemsetting.KeyEQ(key)).
			SetValue(value).
			Save(ctx); retryErr != nil {
			return fmt.Errorf("обновление настройки %s (retry): %w", key, retryErr)
		}
		return nil
	}
	return fmt.Errorf("создание настройки %s: %w", key, err)
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
// Опрос/голоса (news_poll*) скрыты через soft-delete поста (ADR-INTRANET-U03)
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

// ListDeletedScheduledCalls возвращает мягко удалённые серии/звонки в окне retention.
func (r *entRepository) ListDeletedScheduledCalls(ctx context.Context, projectID *uuid.UUID, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.ScheduledCall.Query().
		Where(scheduledcall.DeletedAtNotNil(), scheduledcall.DeletedAtGT(cutoff))
	if projectID != nil {
		query = query.Where(scheduledcall.ProjectIDEQ(*projectID))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых scheduled_call: %w", err)
	}
	items, err := query.
		Order(ent.Desc(scheduledcall.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых scheduled_call: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Title,
			EntityType: "scheduled_call",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// RestoreScheduledCall сбрасывает deleted_at и возвращает серию в активные (FR-CAL-05).
func (r *entRepository) RestoreScheduledCall(ctx context.Context, id uuid.UUID) error {
	return r.client.ScheduledCall.UpdateOneID(id).
		ClearDeletedAt().
		SetIsActive(true).
		Exec(ctx)
}

// SoftDeleteScheduledCall ставит deleted_at и деактивирует серию.
func (r *entRepository) SoftDeleteScheduledCall(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.ScheduledCall.UpdateOneID(id).
		SetDeletedAt(when).
		SetIsActive(false).
		Exec(ctx)
}

// PurgeExpiredScheduledCalls окончательно удаляет серии старше retention.
func (r *entRepository) PurgeExpiredScheduledCalls(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.ScheduledCall.Delete().
		Where(scheduledcall.DeletedAtNotNil(), scheduledcall.DeletedAtLT(cutoff)).
		Exec(ctx)
}

// ListDeletedCRMContacts — мягко удалённые CRM-контакты в окне retention (W7-A2-006).
func (r *entRepository) ListDeletedCRMContacts(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.CrmContact.Query().
		Where(crmcontact.DeletedAtNotNil(), crmcontact.DeletedAtGT(cutoff))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых crm_contact: %w", err)
	}
	items, err := query.
		Order(ent.Desc(crmcontact.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых crm_contact: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Name,
			EntityType: "crm_contact",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// ListDeletedCRMDeals — мягко удалённые CRM-сделки в окне retention (W7-A2-006).
func (r *entRepository) ListDeletedCRMDeals(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.CrmDeal.Query().
		Where(crmdeal.DeletedAtNotNil(), crmdeal.DeletedAtGT(cutoff))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых crm_deal: %w", err)
	}
	items, err := query.
		Order(ent.Desc(crmdeal.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых crm_deal: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Title,
			EntityType: "crm_deal",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

func (r *entRepository) RestoreCRMContact(ctx context.Context, id uuid.UUID) error {
	return r.client.CrmContact.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
}

func (r *entRepository) RestoreCRMDeal(ctx context.Context, id uuid.UUID) error {
	return r.client.CrmDeal.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
}

func (r *entRepository) SoftDeleteCRMContact(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.CrmContact.UpdateOneID(id).SetDeletedAt(when).Exec(ctx)
}

func (r *entRepository) SoftDeleteCRMDeal(ctx context.Context, id uuid.UUID, when time.Time) error {
	return r.client.CrmDeal.UpdateOneID(id).SetDeletedAt(when).Exec(ctx)
}

func (r *entRepository) PurgeExpiredCRMContacts(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.CrmContact.Delete().
		Where(crmcontact.DeletedAtNotNil(), crmcontact.DeletedAtLT(cutoff)).
		Exec(ctx)
}

func (r *entRepository) PurgeExpiredCRMDeals(ctx context.Context, cutoff time.Time) (int, error) {
	return r.client.CrmDeal.Delete().
		Where(crmdeal.DeletedAtNotNil(), crmdeal.DeletedAtLT(cutoff)).
		Exec(ctx)
}

// ListDeletedHubBases — мягко удалённые HubBase в окне retention (A3W2-A5-001).
func (r *entRepository) ListDeletedHubBases(ctx context.Context, cutoff time.Time, offset, limit int) ([]TrashItem, int, error) {
	query := r.client.HubBase.Query().
		Where(hubbase.DeletedAtNotNil(), hubbase.DeletedAtGT(cutoff))
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт удалённых hubbase: %w", err)
	}
	items, err := query.
		Order(ent.Desc(hubbase.FieldDeletedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список удалённых hubbase: %w", err)
	}
	result := make([]TrashItem, 0, len(items))
	for _, item := range items {
		result = append(result, TrashItem{
			ID:         item.ID,
			Name:       item.Name,
			EntityType: "hubbase",
			DeletedAt:  *item.DeletedAt,
		})
	}
	return result, total, nil
}

// RestoreHubBase сбрасывает deleted_at по дереву base→tables→columns→records (A3W2-A5-001).
func (r *entRepository) RestoreHubBase(ctx context.Context, id uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin restore hubbase tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	tableIDs, err := tx.HubTable.Query().
		Where(hubtable.HasBaseWith(hubbase.ID(id))).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("list tables for hubbase restore: %w", err)
	}
	if len(tableIDs) > 0 {
		if _, err := tx.HubRecord.Update().
			Where(hubrecord.HasTableWith(hubtable.IDIn(tableIDs...))).
			ClearDeletedAt().
			Save(ctx); err != nil {
			return fmt.Errorf("restore hubbase records: %w", err)
		}
		if _, err := tx.HubColumn.Update().
			Where(hubcolumn.HasTableWith(hubtable.IDIn(tableIDs...))).
			ClearDeletedAt().
			Save(ctx); err != nil {
			return fmt.Errorf("restore hubbase columns: %w", err)
		}
		if _, err := tx.HubTable.Update().
			Where(hubtable.IDIn(tableIDs...)).
			ClearDeletedAt().
			Save(ctx); err != nil {
			return fmt.Errorf("restore hubbase tables: %w", err)
		}
	}
	if err := tx.HubBase.UpdateOneID(id).ClearDeletedAt().Exec(ctx); err != nil {
		return fmt.Errorf("restore hubbase: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit restore hubbase: %w", err)
	}
	return nil
}

// SoftDeleteHubBase — cascade soft-delete дерева (симметрия bases.SoftDeleteBase).
func (r *entRepository) SoftDeleteHubBase(ctx context.Context, id uuid.UUID, when time.Time) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin soft-delete hubbase tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	tableIDs, err := tx.HubTable.Query().
		Where(hubtable.HasBaseWith(hubbase.ID(id))).
		IDs(ctx)
	if err != nil {
		return fmt.Errorf("list tables for hubbase soft-delete: %w", err)
	}
	if len(tableIDs) > 0 {
		if _, err := tx.HubRecord.Update().
			Where(
				hubrecord.HasTableWith(hubtable.IDIn(tableIDs...)),
				hubrecord.DeletedAtIsNil(),
			).
			SetDeletedAt(when).
			Save(ctx); err != nil {
			return fmt.Errorf("cascade soft-delete hubbase records: %w", err)
		}
		if _, err := tx.HubColumn.Update().
			Where(
				hubcolumn.HasTableWith(hubtable.IDIn(tableIDs...)),
				hubcolumn.DeletedAtIsNil(),
			).
			SetDeletedAt(when).
			Save(ctx); err != nil {
			return fmt.Errorf("cascade soft-delete hubbase columns: %w", err)
		}
		if _, err := tx.HubTable.Update().
			Where(
				hubtable.IDIn(tableIDs...),
				hubtable.DeletedAtIsNil(),
			).
			SetDeletedAt(when).
			Save(ctx); err != nil {
			return fmt.Errorf("cascade soft-delete hubbase tables: %w", err)
		}
	}
	affected, err := tx.HubBase.Update().
		Where(hubbase.IDEQ(id), hubbase.DeletedAtIsNil()).
		SetDeletedAt(when).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit soft-delete hubbase: %w", err)
	}
	return nil
}

// PurgeExpiredHubBases hard-delete дерева HubBase старше retention (A3W2-A5-001).
// Порядок: records→columns→tables→bases (FK; HubBase→tables без OnDelete Cascade).
func (r *entRepository) PurgeExpiredHubBases(ctx context.Context, cutoff time.Time) (int, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin purge hubbase tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	baseIDs, err := tx.HubBase.Query().
		Where(hubbase.DeletedAtNotNil(), hubbase.DeletedAtLT(cutoff)).
		IDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("list expired hubbases: %w", err)
	}
	if len(baseIDs) == 0 {
		return 0, nil
	}
	tableIDs, err := tx.HubTable.Query().
		Where(hubtable.HasBaseWith(hubbase.IDIn(baseIDs...))).
		IDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("list tables for hubbase purge: %w", err)
	}
	if len(tableIDs) > 0 {
		if _, err := tx.HubRecord.Delete().
			Where(hubrecord.HasTableWith(hubtable.IDIn(tableIDs...))).
			Exec(ctx); err != nil {
			return 0, fmt.Errorf("purge hubbase records: %w", err)
		}
		if _, err := tx.HubColumn.Delete().
			Where(hubcolumn.HasTableWith(hubtable.IDIn(tableIDs...))).
			Exec(ctx); err != nil {
			return 0, fmt.Errorf("purge hubbase columns: %w", err)
		}
		if _, err := tx.HubTable.Delete().
			Where(hubtable.IDIn(tableIDs...)).
			Exec(ctx); err != nil {
			return 0, fmt.Errorf("purge hubbase tables: %w", err)
		}
	}
	n, err := tx.HubBase.Delete().
		Where(hubbase.IDIn(baseIDs...)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("purge hubbases: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit purge hubbase: %w", err)
	}
	return n, nil
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

// DeleteUserFiles удаляет FileMetadata, загруженные пользователем (GDPR / W6-A4-001).
func (r *entRepository) DeleteUserFiles(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.FileMetadata.Delete().
		Where(filemetadata.HasUploaderWith(user.IDEQ(userID))).
		Exec(ctx)
}

// DeleteUserProjectMemberships удаляет членства пользователя в проектах.
func (r *entRepository) DeleteUserProjectMemberships(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.ProjectMember.Delete().
		Where(projectmember.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
}

// DeleteUserChannelMemberships удаляет членства пользователя в каналах.
func (r *entRepository) DeleteUserChannelMemberships(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.ChannelMember.Delete().
		Where(channelmember.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
}

// DeleteUserTeamMemberships удаляет членства пользователя в командах.
func (r *entRepository) DeleteUserTeamMemberships(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.TeamMember.Delete().
		Where(teammember.HasUserWith(user.IDEQ(userID))).
		Exec(ctx)
}

// ScrubUserAuditPII очищает IP/UA и отвязывает user_id в audit_log (оставляя
// факт действия для accountability без персональных идентификаторов).
func (r *entRepository) ScrubUserAuditPII(ctx context.Context, userID uuid.UUID) (int, error) {
	return r.client.AuditLog.Update().
		Where(auditlog.UserIDEQ(userID)).
		ClearIPAddress().
		ClearUserAgent().
		ClearUserID().
		Save(ctx)
}

// HardDeleteUser окончательно удаляет запись пользователя.
func (r *entRepository) HardDeleteUser(ctx context.Context, userID uuid.UUID) error {
	return r.client.User.DeleteOneID(userID).Exec(ctx)
}

// IsSystemAdmin возвращает true, если у пользователя установлена системная роль admin.
// Используется audit handler.requireAdmin для проверки доступа к admin-эндпоинтам.
// A3W6-4-003: NotFound → ErrUserNotFound (handler без ent.IsNotFound).
func (r *entRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, ErrUserNotFound
		}
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

func (r *entRepository) LockUserForGDPR(ctx context.Context, userID uuid.UUID) (user.SystemRole, error) {
	roles, err := r.client.User.Query().
		Where(
			user.IDEQ(userID),
			predicate.User(func(selector *sql.Selector) {
				selector.ForUpdate()
			}),
		).
		Select(user.FieldSystemRole).
		Strings(ctx)
	if err != nil {
		return "", err
	}
	if len(roles) != 1 {
		return "", fmt.Errorf("GDPR user lock role cardinality: expected 1, got %d", len(roles))
	}
	return user.SystemRole(roles[0]), nil
}

func (r *entRepository) EraseUserDurableCallData(
	ctx context.Context,
	userID uuid.UUID,
	terminalAt time.Time,
) error {
	if _, err := r.client.CallParticipantConnection.Delete().
		Where(callparticipantconnection.UserIDEQ(userID)).
		Exec(ctx); err != nil {
		return err
	}
	if _, err := r.client.CallParticipantPresence.Delete().
		Where(callparticipantpresence.UserIDEQ(userID)).
		Exec(ctx); err != nil {
		return err
	}
	if _, err := r.client.CallProviderDelivery.Update().
		Where(
			callproviderdelivery.UserIDEQ(userID),
			callproviderdelivery.ResultIn(
				callproviderdelivery.ResultPending,
				callproviderdelivery.ResultReconcile,
			),
		).
		SetResult(callproviderdelivery.ResultIgnored).
		SetTerminalAt(terminalAt).
		ClearUserID().
		Save(ctx); err != nil {
		return err
	}
	_, err := r.client.CallProviderDelivery.Update().
		Where(
			callproviderdelivery.UserIDEQ(userID),
			callproviderdelivery.ResultIn(
				callproviderdelivery.ResultApplied,
				callproviderdelivery.ResultIgnored,
			),
		).
		ClearUserID().
		Save(ctx)
	return err
}
