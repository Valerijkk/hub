package tracker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// ---------------------------------------------------------------------------
// Узкие доменные интерфейсы (ISP — Interface Segregation Principle)
// ---------------------------------------------------------------------------

// IssueRepository управляет CRUD задач и перемещением на доске.
type IssueRepository interface {
	// CreateIssue создаёт новую задачу в указанном проекте.
	CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error)
	// FindIssueByKey возвращает задачу по ключу только внутри projectID (A3-002).
	FindIssueByKey(ctx context.Context, projectID uuid.UUID, issueKey string) (*ent.Issue, error)
	// FindIssueByID возвращает задачу по UUID.
	FindIssueByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error)
	// FindIssueInProject возвращает задачу по UUID только внутри projectID (A3-011-target).
	FindIssueInProject(ctx context.Context, projectID, id uuid.UUID) (*ent.Issue, error)
	// ListIssues возвращает пагинированный список задач проекта с общим количеством.
	ListIssues(ctx context.Context, projectID uuid.UUID, params ListIssuesParams) ([]*ent.Issue, int, error)
	// UpdateIssue применяет частичное обновление задачи.
	UpdateIssue(ctx context.Context, issueID uuid.UUID, input UpdateIssueInput) (*ent.Issue, error)
	// DeleteIssue удаляет задачу по ID.
	DeleteIssue(ctx context.Context, issueID uuid.UUID) error
	// MoveIssue изменяет статус и ранг задачи (перемещение на Kanban-доске).
	MoveIssue(ctx context.Context, issueID uuid.UUID, statusID uuid.UUID, rank string) (*ent.Issue, error)
}

// IssueLinkRepository управляет направленными связями между задачами.
type IssueLinkRepository interface {
	// CreateIssueLink создаёт направленную связь между двумя задачами.
	CreateIssueLink(ctx context.Context, sourceID, targetID uuid.UUID, linkType string, customType string) (*ent.IssueLink, error)
	// ListIssueLinks возвращает все связи (входящие и исходящие) для задачи.
	ListIssueLinks(ctx context.Context, issueID uuid.UUID) ([]*ent.IssueLink, error)
	// ListProjectIssueLinks — A13-004: все связи проекта одним запросом (Gantt batch).
	ListProjectIssueLinks(ctx context.Context, projectID uuid.UUID) ([]*ent.IssueLink, error)
	// DeleteIssueLink удаляет связь только если она принадлежит sourceIssueID (A3-011-link).
	DeleteIssueLink(ctx context.Context, sourceIssueID, linkID uuid.UUID) error
}

// WorkflowRepository управляет статусами и переходами workflow проекта.
type WorkflowRepository interface {
	// ListWorkflowStatuses возвращает все статусы workflow проекта, отсортированные по sort_order.
	ListWorkflowStatuses(ctx context.Context, projectID uuid.UUID) ([]*ent.WorkflowStatus, error)
	// CreateWorkflowStatus создаёт новый статус workflow в проекте.
	CreateWorkflowStatus(ctx context.Context, projectID uuid.UUID, name, category, color string, sortOrder int) (*ent.WorkflowStatus, error)
	// UpdateWorkflowStatus применяет частичное обновление статуса workflow.
	UpdateWorkflowStatus(ctx context.Context, statusID uuid.UUID, name, color *string, sortOrder *int) (*ent.WorkflowStatus, error)
	// DeleteWorkflowStatus удаляет статус workflow по ID.
	DeleteWorkflowStatus(ctx context.Context, statusID uuid.UUID) error
	// FindDefaultStatus возвращает статус workflow по умолчанию для проекта.
	FindDefaultStatus(ctx context.Context, projectID uuid.UUID) (*ent.WorkflowStatus, error)
	// ListWorkflowTransitions возвращает все разрешённые переходы workflow проекта.
	ListWorkflowTransitions(ctx context.Context, projectID uuid.UUID) ([]*ent.WorkflowTransition, error)
	// CreateWorkflowTransition создаёт новый разрешённый переход между статусами.
	CreateWorkflowTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error)
	// DeleteWorkflowTransition удаляет переход workflow по ID.
	DeleteWorkflowTransition(ctx context.Context, transitionID uuid.UUID) error
	// FindTransition проверяет, разрешён ли переход между двумя статусами.
	FindTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error)
	// SeedDefaultWorkflow создаёт стандартные статусы и переходы workflow для нового проекта.
	SeedDefaultWorkflow(ctx context.Context, projectID uuid.UUID) error
}

// CommentRepository управляет комментариями и реакциями к задачам.
type CommentRepository interface {
	// CreateComment добавляет новый комментарий к задаче.
	CreateComment(ctx context.Context, issueID, authorID uuid.UUID, input CreateCommentInput) (*ent.Comment, error)
	// ListComments возвращает все комментарии задачи.
	ListComments(ctx context.Context, issueID uuid.UUID) ([]*ent.Comment, error)
	// UpdateComment обновляет комментарий автора на задаче.
	UpdateComment(ctx context.Context, issueID, authorID, commentID uuid.UUID, input UpdateCommentInput) (*ent.Comment, error)
	// DeleteComment удаляет комментарий автора на задаче.
	DeleteComment(ctx context.Context, issueID, authorID, commentID uuid.UUID) error
	// ToggleReaction добавляет или удаляет реакцию на комментарий задачи (A3-011-react).
	// Возвращает true если добавлена, false если удалена.
	ToggleReaction(ctx context.Context, issueID, commentID, userID uuid.UUID, emoji string) (bool, error)
}

// ActivityRepository управляет логом активности задач.
type ActivityRepository interface {
	// LogActivity записывает событие активности для задачи.
	LogActivity(ctx context.Context, issueID, userID uuid.UUID, action string, changes []FieldChange) (*ent.ActivityLog, error)
	// ListActivities возвращает все записи лога активности задачи.
	ListActivities(ctx context.Context, issueID uuid.UUID) ([]*ent.ActivityLog, error)
}

// TimeTrackingRepository управляет учётом времени на задачах.
type TimeTrackingRepository interface {
	// LogTime записывает вручную введённое время на задачу.
	LogTime(ctx context.Context, issueID, userID uuid.UUID, minutes int, description string) (*ent.TimeLog, error)
	// StartTimer запускает таймер на задаче для пользователя.
	StartTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error)
	// StopTimer останавливает таймер и фиксирует затраченное время.
	StopTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error)
	// ListTimeLogs возвращает все записи учёта времени для задачи.
	ListTimeLogs(ctx context.Context, issueID uuid.UUID) ([]*ent.TimeLog, error)
}

// LabelRepository управляет метками и группами меток проекта.
type LabelRepository interface {
	// CreateLabel создаёт новую метку в проекте, опционально в группе.
	CreateLabel(ctx context.Context, projectID uuid.UUID, name, color string, groupID *uuid.UUID) (*ent.Label, error)
	// ListLabels возвращает все метки проекта.
	ListLabels(ctx context.Context, projectID uuid.UUID) ([]*ent.Label, error)
	// DeleteLabel удаляет метку по ID.
	DeleteLabel(ctx context.Context, labelID uuid.UUID) error
	// CreateLabelGroup создаёт новую группу меток в проекте.
	CreateLabelGroup(ctx context.Context, projectID uuid.UUID, name string) (*ent.LabelGroup, error)
	// ListLabelGroups возвращает все группы меток проекта с вложенными метками.
	ListLabelGroups(ctx context.Context, projectID uuid.UUID) ([]*ent.LabelGroup, error)
}

// BoardRepository управляет досками проекта.
type BoardRepository interface {
	// CreateBoard создаёт новую доску в проекте с опциональными сохранёнными фильтрами.
	CreateBoard(ctx context.Context, projectID uuid.UUID, name string, filters map[string]interface{}) (*ent.Board, error)
	// ListBoards возвращает все доски проекта.
	ListBoards(ctx context.Context, projectID uuid.UUID) ([]*ent.Board, error)
	// GetBoardData возвращает доску по ID с данными для рендеринга.
	GetBoardData(ctx context.Context, boardID uuid.UUID) (*ent.Board, error)
}

// CustomFieldRepository управляет определениями кастомных полей проекта.
type CustomFieldRepository interface {
	// CreateCustomField создаёт новое определение кастомного поля в проекте.
	CreateCustomField(ctx context.Context, projectID uuid.UUID, input CreateCustomFieldInput) (*ent.CustomFieldDefinition, error)
	// ListCustomFields возвращает все определения кастомных полей проекта.
	ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]*ent.CustomFieldDefinition, error)
	// DeleteCustomField удаляет определение кастомного поля по ID.
	DeleteCustomField(ctx context.Context, fieldID uuid.UUID) error
}

// IssueMessageLinkRepository управляет связями задача↔тред мессенджера (PROD-03).
type IssueMessageLinkRepository interface {
	CreateIssueMessageLink(ctx context.Context, issueID, channelID, createdBy uuid.UUID, messageID string, mirrorToThread bool) (*ent.IssueMessageLink, error)
	ListIssueMessageLinksByIssue(ctx context.Context, issueID uuid.UUID) ([]*ent.IssueMessageLink, error)
	ListIssueMessageLinksByThread(ctx context.Context, channelID uuid.UUID, messageID string) ([]*ent.IssueMessageLink, error)
	FindIssueMessageLink(ctx context.Context, linkID uuid.UUID) (*ent.IssueMessageLink, error)
	DeleteIssueMessageLink(ctx context.Context, linkID, issueID uuid.UUID) error
}

// AutomationRepository управляет правилами workflow-автоматизации (PROD-01).
type AutomationRepository interface {
	ListAutomationRules(ctx context.Context, projectID uuid.UUID) ([]*ent.AutomationRule, error)
	CountAutomationRules(ctx context.Context, projectID uuid.UUID) (int, error)
	CreateAutomationRule(ctx context.Context, projectID, createdBy uuid.UUID, input CreateAutomationRuleInput) (*ent.AutomationRule, error)
	FindAutomationRule(ctx context.Context, projectID, ruleID uuid.UUID) (*ent.AutomationRule, error)
	UpdateAutomationRule(ctx context.Context, ruleID uuid.UUID, input UpdateAutomationRuleInput) (*ent.AutomationRule, error)
	DeleteAutomationRule(ctx context.Context, projectID, ruleID uuid.UUID) error
	ListEnabledAutomationRules(ctx context.Context, projectID uuid.UUID, triggerType string) ([]*ent.AutomationRule, error)
}

// ---------------------------------------------------------------------------
// Композитный интерфейс (обратная совместимость)
// ---------------------------------------------------------------------------

// Repository — композитный интерфейс для обратной совместимости.
// Позволяет тестировать сервисный слой с помощью моков.
type Repository interface {
	IssueRepository
	IssueLinkRepository
	WorkflowRepository
	CommentRepository
	ActivityRepository
	TimeTrackingRepository
	LabelRepository
	BoardRepository
	CustomFieldRepository
	AutomationRepository
	IssueMessageLinkRepository

	// GetDB возвращает Ent-клиент для прямых запросов (JQL и др.).
	GetDB() *ent.Client
}

// ---------------------------------------------------------------------------
// Входные структуры репозитория (uuid.UUID, не string)
// ---------------------------------------------------------------------------

// CreateIssueInput — распарсенные и валидированные данные для создания задачи.
type CreateIssueInput struct {
	Title           string
	IssueType       string
	DescriptionJSON map[string]interface{}
	DescriptionText string
	Priority        string
	AssigneeID      *uuid.UUID
	ParentID        *uuid.UUID
	StatusID        *uuid.UUID
	ReporterID      uuid.UUID
	Labels          []uuid.UUID
	CustomFields    map[string]interface{}
	EstimateMinutes *int
	StoryPoints     *int
	Checklist       []map[string]interface{}
	DueDate         *time.Time
}

// UpdateIssueInput — распарсенные и валидированные данные для обновления задачи.
type UpdateIssueInput struct {
	Title           *string
	DescriptionJSON *map[string]interface{}
	DescriptionText *string
	Priority        *string
	AssigneeID      *uuid.UUID
	StatusID        *uuid.UUID
	Labels          *[]uuid.UUID
	CustomFields    *map[string]interface{}
	EstimateMinutes *int
	StoryPoints     *int
	Checklist       *[]map[string]interface{}
	DueDate         *time.Time
	Version         int
}

// ListIssuesParams — распарсенные параметры фильтрации и пагинации для списка задач.
type ListIssuesParams struct {
	Cursor         string
	Limit          int
	Status         []string
	Priority       []string
	Assignee       *uuid.UUID
	UnassignedOnly bool
	IssueType      string
	Search         string
	OrderBy        string
}

// CreateCommentInput — распарсенные данные для создания комментария.
type CreateCommentInput struct {
	BodyJSON   map[string]interface{}
	BodyText   string
	IsInternal bool
	Mentions   []string
}

// UpdateCommentInput — распарсенные данные для обновления комментария.
type UpdateCommentInput struct {
	BodyJSON map[string]interface{}
	BodyText string
}

// CreateCustomFieldInput — распарсенные данные для создания определения кастомного поля.
type CreateCustomFieldInput struct {
	Name      string
	FieldType string
	Options   []string
	Required  bool
}

// ---------------------------------------------------------------------------
// OutboxPublisher и каркас EntRepository
// ---------------------------------------------------------------------------

// OutboxPublisher абстрагирует публикацию событий через outbox-таблицу для репозитория.
type OutboxPublisher interface {
	// PublishInTx записывает событие в outbox-таблицу внутри транзакции.
	PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error
	// Publish записывает событие в outbox-таблицу без внешней транзакции.
	Publish(ctx context.Context, client *ent.Client, eventType string, payload any) error
}

// EntRepository реализует Repository на основе Ent ORM с публикацией событий через outbox.
type EntRepository struct {
	client    *ent.Client
	publisher OutboxPublisher
}

// NewEntRepository создаёт новый Ent-репозиторий трекера задач.
func NewEntRepository(client *ent.Client, publisher OutboxPublisher) *EntRepository {
	return &EntRepository{client: client, publisher: publisher}
}

// GetDB возвращает Ent-клиент для прямых запросов (JQL и др.).
func (r *EntRepository) GetDB() *ent.Client {
	return r.client
}
