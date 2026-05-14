package github

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/githubintegration"
	"github.com/valerii/hub/backend/ent/githublink"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
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
//
// Чувствительные поля (AccessToken, WebhookSecret) шифруются на запись
// AES-256-GCM и расшифровываются на чтение (HI-05). Если cipher == nil
// (например, в тестах), репозиторий ведёт себя как до фикса — это нужно
// чтобы не ломать существующие test-fixture'ы.
type EntRepository struct {
	client *ent.Client
	cipher *cryptopkg.Cipher
}

// NewEntRepository создаёт новый репозиторий GitHub-интеграции.
// cipher может быть nil — тогда токены/секреты будут писаться и читаться
// в открытом виде (как до HI-05). В production cipher обязателен.
func NewEntRepository(client *ent.Client, cipher *cryptopkg.Cipher) *EntRepository {
	return &EntRepository{client: client, cipher: cipher}
}

// encryptField шифрует значение, если cipher настроен. Иначе возвращает as-is.
func (r *EntRepository) encryptField(v string) (string, error) {
	if r.cipher == nil || v == "" {
		return v, nil
	}
	return r.cipher.EncryptString(v)
}

// decryptField расшифровывает значение, если cipher настроен. Если значение
// уже в открытом виде (legacy plaintext до HI-05), возвращает его как есть —
// см. MaybeDecryptString.
func (r *EntRepository) decryptField(v string) string {
	if r.cipher == nil || v == "" {
		return v
	}
	pt, _ := r.cipher.MaybeDecryptString(v)
	return pt
}

// decryptIntegration расшифровывает чувствительные поля для одной интеграции.
// На месте мутирует *ent.GitHubIntegration (Ent struct не immutable). Если
// поле не было зашифровано (legacy plaintext), оно остаётся без изменений.
// Re-encrypt-on-next-save миграция произойдёт лениво.
func (r *EntRepository) decryptIntegration(ig *ent.GitHubIntegration) {
	if ig == nil || r.cipher == nil {
		return
	}
	ig.AccessToken = r.decryptField(ig.AccessToken)
	ig.WebhookSecret = r.decryptField(ig.WebhookSecret)
}

// CreateIntegration создаёт интеграцию с GitHub-репозиторием.
func (r *EntRepository) CreateIntegration(ctx context.Context, projectID uuid.UUID, input CreateIntegrationInput) (*ent.GitHubIntegration, error) {
	encToken, err := r.encryptField(input.AccessToken)
	if err != nil {
		return nil, err
	}
	encSecret, err := r.encryptField(input.WebhookSecret)
	if err != nil {
		return nil, err
	}

	q := r.client.GitHubIntegration.Create().
		SetRepoOwner(input.RepoOwner).
		SetRepoName(input.RepoName).
		SetAccessToken(encToken).
		SetWebhookSecret(encSecret).
		SetProjectID(projectID)

	if input.MergeStatusID != nil {
		q.SetMergeStatusID(*input.MergeStatusID)
	}

	created, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}
	// Return decrypted values to caller so they don't see ciphertext.
	r.decryptIntegration(created)
	return created, nil
}

// ListIntegrations возвращает все интеграции проекта.
func (r *EntRepository) ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]*ent.GitHubIntegration, error) {
	list, err := r.client.GitHubIntegration.Query().
		Where(githubintegration.HasProjectWith(project.ID(projectID))).
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
func (r *EntRepository) GetIntegration(ctx context.Context, id uuid.UUID) (*ent.GitHubIntegration, error) {
	ig, err := r.client.GitHubIntegration.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	r.decryptIntegration(ig)
	return ig, nil
}

// DeleteIntegration удаляет интеграцию по ID.
func (r *EntRepository) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return r.client.GitHubIntegration.DeleteOneID(id).Exec(ctx)
}

// FindIntegrationByRepo ищет интеграцию по owner и repo name.
func (r *EntRepository) FindIntegrationByRepo(ctx context.Context, owner, repo string) (*ent.GitHubIntegration, error) {
	ig, err := r.client.GitHubIntegration.Query().
		Where(
			githubintegration.RepoOwner(owner),
			githubintegration.RepoName(repo),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	r.decryptIntegration(ig)
	return ig, nil
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

// ListAllWebhookSecrets возвращает все webhook-секреты зарегистрированных
// интеграций. Используется только для HMAC-сверки входящей подписи. Возврат
// расшифрованных значений — необходимое зло (нужно сравнивать HMAC), но
// дальше функции эти строки не утекают: caller сравнивает HMAC и сразу
// забывает их. См. HI-05.
func (r *EntRepository) ListAllWebhookSecrets(ctx context.Context) ([]string, error) {
	integrations, err := r.client.GitHubIntegration.Query().
		Select(githubintegration.FieldWebhookSecret).
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
			slog.Warn("github webhook secret decrypted to empty value — skipping")
			continue
		}
		secrets = append(secrets, pt)
	}
	return secrets, nil
}
