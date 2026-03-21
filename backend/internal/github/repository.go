package github

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/githubintegration"
	"github.com/valerii/hub/backend/ent/githublink"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// Repository определяет интерфейс доступа к данным для модуля GitHub-интеграции.
type Repository interface {
	// Интеграции
	CreateIntegration(ctx context.Context, projectID uuid.UUID, input CreateIntegrationInput) (*ent.GitHubIntegration, error)
	ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]*ent.GitHubIntegration, error)
	GetIntegration(ctx context.Context, id uuid.UUID) (*ent.GitHubIntegration, error)
	DeleteIntegration(ctx context.Context, id uuid.UUID) error
	FindIntegrationByRepo(ctx context.Context, owner, repo string) (*ent.GitHubIntegration, error)

	// GitHub-ссылки
	CreateGitHubLink(ctx context.Context, issueID, integrationID uuid.UUID, input CreateGitHubLinkInput) (*ent.GitHubLink, error)
	FindGitHubLink(ctx context.Context, externalID, linkType string, integrationID uuid.UUID) (*ent.GitHubLink, error)
	UpdateGitHubLinkStatus(ctx context.Context, linkID uuid.UUID, status string) error
	ListGitHubLinks(ctx context.Context, issueID uuid.UUID) ([]*ent.GitHubLink, error)

	// Задачи (используется для поиска по ключу)
	FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error)
	UpdateIssueStatus(ctx context.Context, issueID, statusID uuid.UUID) error

	// FindDoneStatus возвращает первый статус категории "done" для проекта.
	FindDoneStatus(ctx context.Context, projectID uuid.UUID) (*ent.WorkflowStatus, error)
	// LogActivity записывает событие в лог активности задачи.
	LogActivity(ctx context.Context, issueID uuid.UUID, action string) error

	// ListAllWebhookSecrets возвращает все webhook-секреты зарегистрированных интеграций.
	ListAllWebhookSecrets(ctx context.Context) ([]string, error)
}

// CreateIntegrationInput -- входные данные для создания интеграции.
type CreateIntegrationInput struct {
	RepoOwner     string
	RepoName      string
	AccessToken   string
	WebhookSecret string
	MergeStatusID *uuid.UUID
}

// CreateGitHubLinkInput -- входные данные для создания связи.
type CreateGitHubLinkInput struct {
	LinkType   string
	ExternalID string
	Title      string
	URL        string
	Status     string
	Author     string
	Branch     string
}

// ---------------------------------------------------------------------------
// EntRepository реализует Repository через Ent ORM.
// ---------------------------------------------------------------------------

// EntRepository реализует Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт новый репозиторий GitHub-интеграции.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// CreateIntegration создаёт интеграцию с GitHub-репозиторием.
func (r *EntRepository) CreateIntegration(ctx context.Context, projectID uuid.UUID, input CreateIntegrationInput) (*ent.GitHubIntegration, error) {
	q := r.client.GitHubIntegration.Create().
		SetRepoOwner(input.RepoOwner).
		SetRepoName(input.RepoName).
		SetAccessToken(input.AccessToken).
		SetWebhookSecret(input.WebhookSecret).
		SetProjectID(projectID)

	if input.MergeStatusID != nil {
		q.SetMergeStatusID(*input.MergeStatusID)
	}

	return q.Save(ctx)
}

// ListIntegrations возвращает все интеграции проекта.
func (r *EntRepository) ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]*ent.GitHubIntegration, error) {
	return r.client.GitHubIntegration.Query().
		Where(githubintegration.HasProjectWith(project.ID(projectID))).
		All(ctx)
}

// GetIntegration возвращает интеграцию по ID.
func (r *EntRepository) GetIntegration(ctx context.Context, id uuid.UUID) (*ent.GitHubIntegration, error) {
	return r.client.GitHubIntegration.Get(ctx, id)
}

// DeleteIntegration удаляет интеграцию по ID.
func (r *EntRepository) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return r.client.GitHubIntegration.DeleteOneID(id).Exec(ctx)
}

// FindIntegrationByRepo ищет интеграцию по owner и repo name.
func (r *EntRepository) FindIntegrationByRepo(ctx context.Context, owner, repo string) (*ent.GitHubIntegration, error) {
	return r.client.GitHubIntegration.Query().
		Where(
			githubintegration.RepoOwner(owner),
			githubintegration.RepoName(repo),
		).
		Only(ctx)
}

// CreateGitHubLink создаёт связь PR/коммита с задачей.
func (r *EntRepository) CreateGitHubLink(ctx context.Context, issueID, integrationID uuid.UUID, input CreateGitHubLinkInput) (*ent.GitHubLink, error) {
	q := r.client.GitHubLink.Create().
		SetLinkType(githublink.LinkType(input.LinkType)).
		SetExternalID(input.ExternalID).
		SetTitle(input.Title).
		SetURL(input.URL).
		SetStatus(githublink.Status(input.Status)).
		SetIssueID(issueID).
		SetIntegrationID(integrationID)

	if input.Author != "" {
		q.SetAuthor(input.Author)
	}
	if input.Branch != "" {
		q.SetBranch(input.Branch)
	}

	return q.Save(ctx)
}

// FindGitHubLink ищет связь по внешнему ID, типу и интеграции.
func (r *EntRepository) FindGitHubLink(ctx context.Context, externalID, linkType string, integrationID uuid.UUID) (*ent.GitHubLink, error) {
	return r.client.GitHubLink.Query().
		Where(
			githublink.ExternalID(externalID),
			githublink.LinkTypeEQ(githublink.LinkType(linkType)),
			githublink.HasIntegrationWith(githubintegration.ID(integrationID)),
		).
		Only(ctx)
}

// UpdateGitHubLinkStatus обновляет статус связи (open/merged/closed).
func (r *EntRepository) UpdateGitHubLinkStatus(ctx context.Context, linkID uuid.UUID, status string) error {
	return r.client.GitHubLink.UpdateOneID(linkID).
		SetStatus(githublink.Status(status)).
		Exec(ctx)
}

// ListGitHubLinks возвращает все GitHub-ссылки для задачи.
func (r *EntRepository) ListGitHubLinks(ctx context.Context, issueID uuid.UUID) ([]*ent.GitHubLink, error) {
	return r.client.GitHubLink.Query().
		Where(githublink.HasIssueWith(issue.ID(issueID))).
		Order(ent.Desc(githublink.FieldCreatedAt)).
		All(ctx)
}

// FindIssueByKey ищет задачу по ключу (например "HUB-1").
// Загружает связи status, assignee, project для использования в webhook-обработчиках.
func (r *EntRepository) FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(issue.IssueKey(issueKey)).
		WithStatus().
		WithAssignee().
		WithProject().
		Only(ctx)
}

// UpdateIssueStatus обновляет статус задачи.
func (r *EntRepository) UpdateIssueStatus(ctx context.Context, issueID, statusID uuid.UUID) error {
	return r.client.Issue.UpdateOneID(issueID).
		SetStatusID(statusID).
		Exec(ctx)
}

// FindDoneStatus возвращает первый статус категории "done" для проекта.
func (r *EntRepository) FindDoneStatus(ctx context.Context, projectID uuid.UUID) (*ent.WorkflowStatus, error) {
	return r.client.WorkflowStatus.Query().
		Where(
			workflowstatus.HasProjectWith(project.IDEQ(projectID)),
			workflowstatus.CategoryEQ(workflowstatus.CategoryDone),
		).
		First(ctx)
}

// LogActivity записывает событие в лог активности задачи.
func (r *EntRepository) LogActivity(ctx context.Context, issueID uuid.UUID, action string) error {
	_, err := r.client.ActivityLog.Create().
		SetIssueID(issueID).
		SetAction(action).
		Save(ctx)
	return err
}

// ListAllWebhookSecrets возвращает все webhook-секреты зарегистрированных интеграций.
// Используется для валидации подписи входящих webhook-запросов.
func (r *EntRepository) ListAllWebhookSecrets(ctx context.Context) ([]string, error) {
	integrations, err := r.client.GitHubIntegration.Query().
		Select(githubintegration.FieldWebhookSecret).
		All(ctx)
	if err != nil {
		return nil, err
	}

	secrets := make([]string, 0, len(integrations))
	for _, ig := range integrations {
		if ig.WebhookSecret != "" {
			secrets = append(secrets, ig.WebhookSecret)
		}
	}
	return secrets, nil
}
