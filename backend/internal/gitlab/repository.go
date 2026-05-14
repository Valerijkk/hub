package gitlab

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/gitlabintegration"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
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
//
// Чувствительные поля (AccessToken, WebhookSecret) шифруются на запись и
// расшифровываются на чтение через cryptopkg.Cipher (HI-05). Если
// cipher == nil — fallback на plaintext-поведение (для тестов).
type EntRepository struct {
	client *ent.Client
	cipher *cryptopkg.Cipher
}

// NewEntRepository создаёт новый репозиторий GitLab-интеграции.
// cipher может быть nil — тогда токены/секреты будут писаться и читаться
// в открытом виде. В production cipher обязателен.
func NewEntRepository(client *ent.Client, cipher *cryptopkg.Cipher) *EntRepository {
	return &EntRepository{client: client, cipher: cipher}
}

func (r *EntRepository) encryptField(v string) (string, error) {
	if r.cipher == nil || v == "" {
		return v, nil
	}
	return r.cipher.EncryptString(v)
}

func (r *EntRepository) decryptField(v string) string {
	if r.cipher == nil || v == "" {
		return v
	}
	pt, _ := r.cipher.MaybeDecryptString(v)
	return pt
}

func (r *EntRepository) decryptIntegration(ig *ent.GitLabIntegration) {
	if ig == nil || r.cipher == nil {
		return
	}
	ig.AccessToken = r.decryptField(ig.AccessToken)
	ig.WebhookSecret = r.decryptField(ig.WebhookSecret)
}

// CreateIntegration создаёт интеграцию с GitLab-репозиторием.
func (r *EntRepository) CreateIntegration(ctx context.Context, projectID uuid.UUID, input CreateIntegrationInput) (*ent.GitLabIntegration, error) {
	gitlabURL := input.GitLabURL
	if gitlabURL == "" {
		gitlabURL = "https://gitlab.com"
	}

	encToken, err := r.encryptField(input.AccessToken)
	if err != nil {
		return nil, err
	}

	q := r.client.GitLabIntegration.Create().
		SetGitlabURL(gitlabURL).
		SetProjectPath(input.ProjectPath).
		SetAccessToken(encToken).
		SetProjectID(projectID)

	if input.WebhookSecret != "" {
		encSecret, err := r.encryptField(input.WebhookSecret)
		if err != nil {
			return nil, err
		}
		q.SetWebhookSecret(encSecret)
	}

	created, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}
	r.decryptIntegration(created)
	return created, nil
}

// ListIntegrations возвращает все GitLab-интеграции проекта.
func (r *EntRepository) ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]*ent.GitLabIntegration, error) {
	list, err := r.client.GitLabIntegration.Query().
		Where(gitlabintegration.HasProjectWith(project.ID(projectID))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, ig := range list {
		r.decryptIntegration(ig)
	}
	return list, nil
}

// GetIntegration возвращает интеграцию по ID.
func (r *EntRepository) GetIntegration(ctx context.Context, id uuid.UUID) (*ent.GitLabIntegration, error) {
	ig, err := r.client.GitLabIntegration.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	r.decryptIntegration(ig)
	return ig, nil
}

// DeleteIntegration удаляет интеграцию по ID.
func (r *EntRepository) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return r.client.GitLabIntegration.DeleteOneID(id).Exec(ctx)
}

// FindIntegrationByProjectPath ищет интеграцию по пути проекта GitLab.
func (r *EntRepository) FindIntegrationByProjectPath(ctx context.Context, projectPath string) (*ent.GitLabIntegration, error) {
	ig, err := r.client.GitLabIntegration.Query().
		Where(gitlabintegration.ProjectPath(projectPath)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	r.decryptIntegration(ig)
	return ig, nil
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

// ListAllWebhookSecrets возвращает все webhook-секреты зарегистрированных
// GitLab-интеграций. Расшифровывает каждое значение через MaybeDecrypt,
// чтобы поддерживать как новые (encrypted) так и legacy (plaintext) записи.
// Caller использует значения только для HMAC-сверки и не сохраняет.
func (r *EntRepository) ListAllWebhookSecrets(ctx context.Context) ([]string, error) {
	integrations, err := r.client.GitLabIntegration.Query().
		Select(gitlabintegration.FieldWebhookSecret).
		All(ctx)
	if err != nil {
		return nil, err
	}

	secrets := make([]string, 0, len(integrations))
	for _, ig := range integrations {
		if ig.WebhookSecret == "" {
			continue
		}
		pt := r.decryptField(ig.WebhookSecret)
		if pt == "" {
			slog.Warn("gitlab webhook secret decrypted to empty value — skipping")
			continue
		}
		secrets = append(secrets, pt)
	}
	return secrets, nil
}
