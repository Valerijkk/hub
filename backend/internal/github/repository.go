package github

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/githubintegration"
	"github.com/valerii/hub/backend/ent/githublink"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// ME-08: TTL для in-memory кеша webhook-секретов. 60s — короткий
// горизонт, достаточно для гашения хвостовой нагрузки GitHub push-spam
// и при этом не требует явной инвалидации после CreateIntegration:
// новые секреты попадают в HMAC-сверку с задержкой максимум 60 секунд,
// что для регистрации webhook'а допустимо.
const webhookSecretsCacheTTL = 60 * time.Second

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

	// MarkTestCasesOutdatedByFiles находит test cases со статусом "synced",
	// у которых ClassName/FullName совпадает с одним из переданных путей файлов
	// (с нормализацией / и \ → .), помечает их sync_status = "outdated"
	// и возвращает case_keys обновлённых записей.
	//
	// LO-3: метод инкапсулирует прямой доступ к ent.TestCase, чтобы
	// GitHubService мог обойтись без *ent.Client как поля.
	MarkTestCasesOutdatedByFiles(ctx context.Context, changedFiles []string) ([]string, error)
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

	// ME-08: кеш расшифрованных webhook-секретов для быстрой HMAC-сверки
	// при высокой частоте webhook'ов от GitHub (push-spam на больших
	// репозиториях). TTL=60s — короткий, чтобы свежие интеграции
	// подхватывались автоматически без явной инвалидации.
	secretsMu     sync.RWMutex
	secretsCache  []string
	secretsExpire time.Time
}

// NewEntRepository создаёт новый репозиторий GitHub-интеграции.
// cipher может быть nil — тогда токены/секреты будут писаться и читаться
// в открытом виде (как до HI-05). В production cipher обязателен.
func NewEntRepository(client *ent.Client, cipher *cryptopkg.Cipher) *EntRepository {
	return &EntRepository{client: client, cipher: cipher}
}

// invalidateSecretsCache сбрасывает кеш webhook-секретов.
// Вызывается после Create/Delete интеграции, чтобы новая интеграция
// сразу подхватывалась HMAC-сверкой webhook'ов (без ожидания TTL).
func (r *EntRepository) invalidateSecretsCache() {
	r.secretsMu.Lock()
	r.secretsCache = nil
	r.secretsExpire = time.Time{}
	r.secretsMu.Unlock()
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
	// ME-08: новая интеграция — сбрасываем кеш секретов, чтобы её secret
	// сразу учитывался в HMAC-сверке (а не через 60 секунд).
	r.invalidateSecretsCache()
	// Возвращаем decrypt caller — не показываем ciphertext в API.
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
	if err := r.client.GitHubIntegration.DeleteOneID(id).Exec(ctx); err != nil {
		return err
	}
	// ME-08: удалённый секрет должен немедленно исчезнуть из HMAC-сверки.
	r.invalidateSecretsCache()
	return nil
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
//
// ME-08: результат кешируется в памяти на webhookSecretsCacheTTL (60s),
// чтобы спам webhook'ов от GitHub (push-events на больших репозиториях)
// не приводил к запросу к Postgres на каждом hit. Кеш инвалидируется
// явно после CreateIntegration / DeleteIntegration; в худшем случае
// (если invalidate не вызывался) новый секрет подхватится с задержкой
// до 60 секунд.
func (r *EntRepository) ListAllWebhookSecrets(ctx context.Context) ([]string, error) {
	// Fast path: cache hit под RLock — без лишних decrypt.
	r.secretsMu.RLock()
	if !r.secretsExpire.IsZero() && time.Now().Before(r.secretsExpire) {
		// Возвращаем копию, чтобы caller не мог мутировать слайс кеша.
		out := make([]string, len(r.secretsCache))
		copy(out, r.secretsCache)
		r.secretsMu.RUnlock()
		return out, nil
	}
	r.secretsMu.RUnlock()

	// Slow path: rebuild cache. Допускаем гонку — два параллельных
	// промаха выполнят по одному запросу к БД, но это много лучше,
	// чем sync.Mutex на весь хот-путь HMAC-сверки.
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

	r.secretsMu.Lock()
	r.secretsCache = secrets
	r.secretsExpire = time.Now().Add(webhookSecretsCacheTTL)
	r.secretsMu.Unlock()

	// Возвращаем копию, чтобы caller не мог мутировать слайс кеша.
	out := make([]string, len(secrets))
	copy(out, secrets)
	return out, nil
}

// MarkTestCasesOutdatedByFiles находит test cases со статусом "synced",
// у которых ClassName/FullName совпадает с одним из переданных путей файлов,
// помечает их sync_status = "outdated" и возвращает case_keys обновлённых.
//
// LO-3: до рефакторинга эта логика была inline в GitHubService.markTestCasesOutdated
// и требовала *ent.Client как поле сервиса. Перенос в Repository убирает
// последнее место в сервисе, где требовался прямой доступ к ent.
func (r *EntRepository) MarkTestCasesOutdatedByFiles(ctx context.Context, changedFiles []string) ([]string, error) {
	if len(changedFiles) == 0 {
		return nil, nil
	}

	cases, err := r.client.TestCase.Query().
		Where(
			testcase.Or(
				testcase.ClassNameNotNil(),
				testcase.FullNameNotNil(),
			),
			testcase.SyncStatusEQ(testcase.SyncStatusSynced),
		).
		All(ctx)
	if err != nil {
		slog.Error("ошибка получения тест-кейсов для sync_status", "error", err)
		return nil, err
	}

	updated := make([]string, 0, len(cases))
	for _, tc := range cases {
		matched := false
		for _, file := range changedFiles {
			if matchTestCaseFileByPath(tc.ClassName, tc.FullName, file) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		if err := r.client.TestCase.UpdateOneID(tc.ID).
			SetSyncStatus(testcase.SyncStatusOutdated).
			Exec(ctx); err != nil {
			slog.Error("ошибка обновления sync_status", "case_key", tc.CaseKey, "error", err)
			continue
		}
		updated = append(updated, tc.CaseKey)
	}
	return updated, nil
}

// matchTestCaseFileByPath проверяет, соответствует ли изменённый файл
// тест-кейсу по className или fullName (с учётом паттернов Java/Python-пакетов).
//
// Например: src/test/java/com/example/MyTest.java нормализуется в
// com.example.MyTest и сравнивается через strings.Contains. Это даёт
// разумное приближение — false-positive возможен (если className —
// слишком короткая подстрока), но достаточно для пометки sync_status.
func matchTestCaseFileByPath(className, fullName, filePath string) bool {
	normalized := strings.ReplaceAll(filePath, "/", ".")
	normalized = strings.ReplaceAll(normalized, "\\", ".")
	if className != "" && strings.Contains(normalized, className) {
		return true
	}
	if fullName != "" && strings.Contains(normalized, fullName) {
		return true
	}
	return false
}
