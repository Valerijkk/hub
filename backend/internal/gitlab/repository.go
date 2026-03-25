package gitlab

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/gitlabintegration"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
)

// Repository определяет интерфейс доступа к данным для модуля GitLab-интеграции.
type Repository interface {
	// Интеграции
	CreateIntegration(ctx context.Context, projectID uuid.UUID, input CreateIntegrationInput) (*ent.GitLabIntegration, error)
	ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]*ent.GitLabIntegration, error)
	GetIntegration(ctx context.Context, id uuid.UUID) (*ent.GitLabIntegration, error)
	DeleteIntegration(ctx context.Context, id uuid.UUID) error
	FindIntegrationByProjectPath(ctx context.Context, projectPath string) (*ent.GitLabIntegration, error)

	// Задачи (используется для поиска по ключу)
	FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error)

	// ListAllWebhookSecrets возвращает все webhook-секреты зарегистрированных интеграций.
	ListAllWebhookSecrets(ctx context.Context) ([]string, error)
}

// CreateIntegrationInput -- входные данные для создания интеграции.
type CreateIntegrationInput struct {
	GitLabURL     string
	ProjectPath   string
	AccessToken   string
	WebhookSecret string
}

// ---------------------------------------------------------------------------
// EntRepository реализует Repository через Ent ORM.
// ---------------------------------------------------------------------------

// EntRepository реализует Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт новый репозиторий GitLab-интеграции.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// CreateIntegration создаёт интеграцию с GitLab-репозиторием.
func (r *EntRepository) CreateIntegration(ctx context.Context, projectID uuid.UUID, input CreateIntegrationInput) (*ent.GitLabIntegration, error) {
	gitlabURL := input.GitLabURL
	if gitlabURL == "" {
		gitlabURL = "https://gitlab.com"
	}

	q := r.client.GitLabIntegration.Create().
		SetGitlabURL(gitlabURL).
		SetProjectPath(input.ProjectPath).
		SetAccessToken(input.AccessToken).
		SetProjectID(projectID)

	if input.WebhookSecret != "" {
		q.SetWebhookSecret(input.WebhookSecret)
	}

	return q.Save(ctx)
}

// ListIntegrations возвращает все GitLab-интеграции проекта.
func (r *EntRepository) ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]*ent.GitLabIntegration, error) {
	return r.client.GitLabIntegration.Query().
		Where(gitlabintegration.HasProjectWith(project.ID(projectID))).
		All(ctx)
}

// GetIntegration возвращает интеграцию по ID.
func (r *EntRepository) GetIntegration(ctx context.Context, id uuid.UUID) (*ent.GitLabIntegration, error) {
	return r.client.GitLabIntegration.Get(ctx, id)
}

// DeleteIntegration удаляет интеграцию по ID.
func (r *EntRepository) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return r.client.GitLabIntegration.DeleteOneID(id).Exec(ctx)
}

// FindIntegrationByProjectPath ищет интеграцию по пути проекта GitLab.
func (r *EntRepository) FindIntegrationByProjectPath(ctx context.Context, projectPath string) (*ent.GitLabIntegration, error) {
	return r.client.GitLabIntegration.Query().
		Where(gitlabintegration.ProjectPath(projectPath)).
		Only(ctx)
}

// FindIssueByKey ищет задачу по ключу (например "HUB-1").
func (r *EntRepository) FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(issue.IssueKey(issueKey)).
		WithStatus().
		WithAssignee().
		WithProject().
		Only(ctx)
}

// ListAllWebhookSecrets возвращает все webhook-секреты зарегистрированных GitLab-интеграций.
func (r *EntRepository) ListAllWebhookSecrets(ctx context.Context) ([]string, error) {
	integrations, err := r.client.GitLabIntegration.Query().
		Select(gitlabintegration.FieldWebhookSecret).
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
