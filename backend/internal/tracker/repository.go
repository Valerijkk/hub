package tracker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// Repository определяет интерфейс доступа к данным для модуля трекера задач.
// Позволяет тестировать сервисный слой с помощью моков.
type Repository interface {
	// Issue CRUD

	// CreateIssue создаёт новую задачу в указанном проекте.
	CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error)
	// FindIssueByKey возвращает задачу по проектному ключу (например "HUB-1").
	FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error)
	// FindIssueByID возвращает задачу по UUID.
	FindIssueByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error)
	// ListIssues возвращает пагинированный список задач проекта с общим количеством.
	ListIssues(ctx context.Context, projectID uuid.UUID, params ListIssuesParams) ([]*ent.Issue, int, error)
	// UpdateIssue применяет частичное обновление задачи.
	UpdateIssue(ctx context.Context, issueID uuid.UUID, input UpdateIssueInput) (*ent.Issue, error)
	// DeleteIssue удаляет задачу по ID.
	DeleteIssue(ctx context.Context, issueID uuid.UUID) error
	// MoveIssue изменяет статус и ранг задачи (перемещение на Kanban-доске).
	MoveIssue(ctx context.Context, issueID uuid.UUID, statusID uuid.UUID, rank string) (*ent.Issue, error)

	// Issue links

	// CreateIssueLink создаёт направленную связь между двумя задачами.
	CreateIssueLink(ctx context.Context, sourceID, targetID uuid.UUID, linkType string, customType string) (*ent.IssueLink, error)
	// ListIssueLinks возвращает все связи (входящие и исходящие) для задачи.
	ListIssueLinks(ctx context.Context, issueID uuid.UUID) ([]*ent.IssueLink, error)
	// DeleteIssueLink удаляет связь между задачами.
	DeleteIssueLink(ctx context.Context, linkID uuid.UUID) error

	// Workflow

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

	// Comments

	// CreateComment добавляет новый комментарий к задаче.
	CreateComment(ctx context.Context, issueID, authorID uuid.UUID, input CreateCommentInput) (*ent.Comment, error)
	// ListComments возвращает все комментарии задачи.
	ListComments(ctx context.Context, issueID uuid.UUID) ([]*ent.Comment, error)
	// UpdateComment обновляет содержимое комментария.
	UpdateComment(ctx context.Context, commentID uuid.UUID, input UpdateCommentInput) (*ent.Comment, error)
	// DeleteComment удаляет комментарий по ID.
	DeleteComment(ctx context.Context, commentID uuid.UUID) error

	// Reactions

	// ToggleReaction добавляет или удаляет реакцию. Возвращает true если добавлена, false если удалена.
	ToggleReaction(ctx context.Context, commentID, userID uuid.UUID, emoji string) (bool, error)

	// Activity log

	// LogActivity записывает событие активности для задачи.
	LogActivity(ctx context.Context, issueID, userID uuid.UUID, action string, changes []FieldChange) (*ent.ActivityLog, error)
	// ListActivities возвращает все записи лога активности задачи.
	ListActivities(ctx context.Context, issueID uuid.UUID) ([]*ent.ActivityLog, error)

	// Time tracking

	// LogTime записывает вручную введённое время на задачу.
	LogTime(ctx context.Context, issueID, userID uuid.UUID, minutes int, description string) (*ent.TimeLog, error)
	// StartTimer запускает таймер на задаче для пользователя.
	StartTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error)
	// StopTimer останавливает таймер и фиксирует затраченное время.
	StopTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error)
	// ListTimeLogs возвращает все записи учёта времени для задачи.
	ListTimeLogs(ctx context.Context, issueID uuid.UUID) ([]*ent.TimeLog, error)

	// Labels

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

	// Boards

	// CreateBoard создаёт новую доску в проекте с опциональными сохранёнными фильтрами.
	CreateBoard(ctx context.Context, projectID uuid.UUID, name string, filters map[string]interface{}) (*ent.Board, error)
	// ListBoards возвращает все доски проекта.
	ListBoards(ctx context.Context, projectID uuid.UUID) ([]*ent.Board, error)
	// GetBoardData возвращает доску по ID с данными для рендеринга.
	GetBoardData(ctx context.Context, boardID uuid.UUID) (*ent.Board, error)

	// Custom fields

	// CreateCustomField создаёт новое определение кастомного поля в проекте.
	CreateCustomField(ctx context.Context, projectID uuid.UUID, input CreateCustomFieldInput) (*ent.CustomFieldDefinition, error)
	// ListCustomFields возвращает все определения кастомных полей проекта.
	ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]*ent.CustomFieldDefinition, error)
	// DeleteCustomField удаляет определение кастомного поля по ID.
	DeleteCustomField(ctx context.Context, fieldID uuid.UUID) error

	// GetDB возвращает Ent-клиент для прямых запросов (JQL и др.).
	GetDB() *ent.Client
}

// ---------------------------------------------------------------------------
// Repository input structs (use uuid.UUID, not string)
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
	Cursor    string
	Limit     int
	Status    []string
	Priority  []string
	Assignee  *uuid.UUID
	IssueType string
	Search    string
	OrderBy   string
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
// OutboxPublisher + EntRepository skeleton
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
