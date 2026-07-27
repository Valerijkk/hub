package gitlab

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/gitlabintegration"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// ME-08: TTL для кеша webhook-секретов GitLab-интеграций.
const webhookSecretsCacheTTL = 60 * time.Second

// Repository определяет интерфейс доступа к данным для модуля GitLab-интеграции.
type Repository interface {
	// Интеграции
	CreateIntegration(ctx context.Context, projectID uuid.UUID, input CreateIntegrationInput) (*ent.GitLabIntegration, error)
	ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]*ent.GitLabIntegration, error)
	GetIntegration(ctx context.Context, id uuid.UUID) (*ent.GitLabIntegration, error)
	// DeleteIntegration удаляет интеграцию только если она принадлежит projectID (A5-004).
	DeleteIntegration(ctx context.Context, projectID, id uuid.UUID) error
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

	// ME-08: in-memory кеш webhook-секретов с TTL.
	secretsMu     sync.RWMutex
	secretsCache  []string
	secretsExpire time.Time
}

// NewEntRepository создаёт новый репозиторий GitLab-интеграции.
// cipher может быть nil — тогда токены/секреты будут писаться и читаться
// в открытом виде. В production cipher обязателен.
func NewEntRepository(client *ent.Client, cipher *cryptopkg.Cipher) *EntRepository {
	return &EntRepository{client: client, cipher: cipher}
}

// invalidateSecretsCache сбрасывает кеш webhook-секретов.
func (r *EntRepository) invalidateSecretsCache() {
	r.secretsMu.Lock()
	r.secretsCache = nil
	r.secretsExpire = time.Time{}
	r.secretsMu.Unlock()
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
	// ME-08: сбрасываем кеш — новый secret должен немедленно учитываться.
	r.invalidateSecretsCache()
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

// DeleteIntegration удаляет интеграцию по ID только в рамках projectID (A5-004).
// Чужой projectId в URL при UUID интеграции другого проекта → NotFound (0 rows).
func (r *EntRepository) DeleteIntegration(ctx context.Context, projectID, id uuid.UUID) error {
	affected, err := r.client.GitLabIntegration.Delete().
		Where(
			gitlabintegration.IDEQ(id),
			gitlabintegration.HasProjectWith(project.ID(projectID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	// ME-08: удалённый секрет должен немедленно исчезнуть из HMAC-сверки.
	r.invalidateSecretsCache()
	return nil
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
//
// ME-08: результат кешируется на webhookSecretsCacheTTL (60s) ровно тем
// же способом, что и в github/repository.go. Инвалидируется явно после
// Create/Delete интеграции; новые секреты подхватываются с задержкой не
// более 60 секунд.
func (r *EntRepository) ListAllWebhookSecrets(ctx context.Context) ([]string, error) {
	r.secretsMu.RLock()
	if !r.secretsExpire.IsZero() && time.Now().Before(r.secretsExpire) {
		out := make([]string, len(r.secretsCache))
		copy(out, r.secretsCache)
		r.secretsMu.RUnlock()
		return out, nil
	}
	r.secretsMu.RUnlock()

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

	r.secretsMu.Lock()
	r.secretsCache = secrets
	r.secretsExpire = time.Now().Add(webhookSecretsCacheTTL)
	r.secretsMu.Unlock()

	out := make([]string, len(secrets))
	copy(out, secrets)
	return out, nil
}
