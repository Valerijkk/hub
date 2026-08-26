// Package search реализует глобальный полнотекстовый поиск по PostgreSQL (tsvector) и MongoDB ($text).
package search

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
"github.com/valerii/hub/backend/internal/messenger"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// defaultSearchLimit — потолок на количество результатов на каждый search-
// шарду (issues, projects, messages и т.д.) перед общим объединением и
// финальным `limit` в Search(). Значение подобрано так, чтобы суммарный
// результат хорошо отображался в UI (типичная команда — 5-10 проектов,
// несколько десятков задач) и при этом не нагружал tsvector indexes.
const defaultSearchLimit = 10

// ErrFilterForbidden — A4-003/A4-004 / A2W5-A4-001/004: 403 без утечки причины
// (нет membership или не владелец фильтра).
var ErrFilterForbidden = errors.New("forbidden")

// ErrFilterNotFound — A2W5-A4-001: отсутствующий фильтр → 404, не 500.
var ErrFilterNotFound = errors.New("фильтр не найден")

// ErrSearchUnavailable — A2W5-A4-003: все шарды упали (total outage → hard error).
var ErrSearchUnavailable = errors.New("search unavailable: all shards failed")
// AuthorizedMessageSearchScopeResolver supplies fresh PostgreSQL authority
// before the global-search Mongo shard can be scheduled.
type AuthorizedMessageSearchScopeResolver interface {
	ResolveAuthorizedMessageSearchScope(context.Context, uuid.UUID) (messenger.AuthorizedMessageSearchScope, error)
}

// AuthorizedMessageSearchExecutor is the only Mongo message-search boundary.
type AuthorizedMessageSearchExecutor interface {
	SearchAuthorizedMessages(context.Context, messenger.AuthorizedMessageSearchScope, string, int) ([]SearchResult, error)
}

// AuthorizedMessageSearchTaskStarter makes scheduling observable while keeping
// the production path concurrent with the other global-search shards.
type AuthorizedMessageSearchTaskStarter interface {
	StartAuthorizedMessageSearch(func())
}

type goAuthorizedMessageSearchTaskStarter struct{}

func (goAuthorizedMessageSearchTaskStarter) StartAuthorizedMessageSearch(task func()) {
	go task()
}
// SearchService реализует глобальный поиск по всем сущностям платформы.
// Выполняет параллельные запросы к PostgreSQL (tsvector) и MongoDB ($text),
// затем объединяет результаты по релевантности.
//
// SC-14-3: typed Ent-операции (SystemRole lookup, SavedFilter CRUD) вынесены
// в Repository. Сам сервис держит только sqlDB (raw tsvector) и mongoDB ($text) —
// эти БД-уровневые драйверы концептуально ниже ORM и для них репозиторий-
// абстракция избыточна.
type SearchService struct {
	repo                 Repository
	sqlDB                *sql.DB
	mongoDB              *mongo.Database
	messageScopeResolver AuthorizedMessageSearchScopeResolver
	messageExecutor      AuthorizedMessageSearchExecutor
	messageTaskStarter   AuthorizedMessageSearchTaskStarter
}

// NewSearchService создаёт новый сервис поиска.
// sqlDB — прямое подключение к PostgreSQL для raw-запросов с tsvector.
func NewSearchService(repo Repository, sqlDB *sql.DB, mongoDB *mongo.Database) *SearchService {
	return &SearchService{repo: repo, sqlDB: sqlDB, mongoDB: mongoDB}
}

// NewSearchServiceWithAuthorizedMessageSearchScope creates the production
// service. The resolver is mandatory; Mongo execution and scheduling use the
// same dependencies exercised by the Phase 151 contract tests.
func NewSearchServiceWithAuthorizedMessageSearchScope(
	repo Repository,
	sqlDB *sql.DB,
	mongoDB *mongo.Database,
	resolver AuthorizedMessageSearchScopeResolver,
) *SearchService {
	if resolver == nil {
		panic("search authorized message scope resolver is required")
	}
	service := newSearchServiceWithAuthorizedMessageSearchDependencies(
		repo, sqlDB, mongoDB, resolver, nil, goAuthorizedMessageSearchTaskStarter{},
	)
	service.messageExecutor = service
	return service
}

func newSearchServiceWithAuthorizedMessageSearchDependencies(
	repo Repository,
	sqlDB *sql.DB,
	mongoDB *mongo.Database,
	resolver AuthorizedMessageSearchScopeResolver,
	executor AuthorizedMessageSearchExecutor,
	starter AuthorizedMessageSearchTaskStarter,
) *SearchService {
	return &SearchService{
		repo: repo, sqlDB: sqlDB, mongoDB: mongoDB,
		messageScopeResolver: resolver,
		messageExecutor:      executor,
		messageTaskStarter:   starter,
	}
}
const maxAuthorizedMessageSearchTargets = 500

func (s *SearchService) resolveAuthorizedMessageSearchScope(
	ctx context.Context,
	userID uuid.UUID,
) (messenger.AuthorizedMessageSearchScope, bool, error) {
	if s == nil || s.messageScopeResolver == nil {
		return messenger.AuthorizedMessageSearchScope{}, false, nil
	}
	scope, err := s.messageScopeResolver.ResolveAuthorizedMessageSearchScope(ctx, userID)
	if err != nil {
		if errors.Is(err, messenger.ErrChannelAccessDenied) {
			return messenger.AuthorizedMessageSearchScope{}, false, nil
		}
		return messenger.AuthorizedMessageSearchScope{}, false, fmt.Errorf("resolve authorized message search scope: %w", err)
	}
	if scope.TenantID == uuid.Nil || scope.UserID != userID || len(scope.Targets) == 0 || len(scope.Targets) > maxAuthorizedMessageSearchTargets {
		return messenger.AuthorizedMessageSearchScope{}, false, nil
	}
	seen := make(map[uuid.UUID]struct{}, len(scope.Targets))
	targets := make([]messenger.AuthorizedMessageSearchTarget, 0, len(scope.Targets))
	for _, target := range scope.Targets {
		if target.ChannelID == uuid.Nil || (target.Kind != "channel" && target.Kind != "dm") {
			return messenger.AuthorizedMessageSearchScope{}, false, nil
		}
		if _, duplicate := seen[target.ChannelID]; duplicate {
			continue
		}
		seen[target.ChannelID] = struct{}{}
		targets = append(targets, target)
	}
	if len(targets) == 0 {
		return messenger.AuthorizedMessageSearchScope{}, false, nil
	}
	scope.Targets = targets
	return scope, true, nil
}

func (s *SearchService) SearchAuthorizedMessages(
	ctx context.Context,
	scope messenger.AuthorizedMessageSearchScope,
	query string,
	limit int,
) ([]SearchResult, error) {
	if s == nil || s.mongoDB == nil || scope.TenantID == uuid.Nil || scope.UserID == uuid.Nil || len(scope.Targets) == 0 || len(scope.Targets) > maxAuthorizedMessageSearchTargets {
		return nil, nil
	}
	channelIDs := make([]uuid.UUID, 0, len(scope.Targets))
	seen := make(map[uuid.UUID]struct{}, len(scope.Targets))
	for _, target := range scope.Targets {
		if target.ChannelID == uuid.Nil || (target.Kind != "channel" && target.Kind != "dm") {
			return nil, nil
		}
		if _, duplicate := seen[target.ChannelID]; duplicate {
			continue
		}
		seen[target.ChannelID] = struct{}{}
		channelIDs = append(channelIDs, target.ChannelID)
	}
	return s.searchMessages(ctx, query, channelIDs, limit)
}

func (s *SearchService) searchAuthorizedMessagesForUser(
	ctx context.Context,
	query string,
	userID uuid.UUID,
) ([]SearchResult, error) {
	scope, allowed, err := s.resolveAuthorizedMessageSearchScope(ctx, userID)
	if err != nil || !allowed {
		return nil, err
	}
	if s.messageExecutor == nil || s.messageTaskStarter == nil {
		return nil, errors.New("authorized message search runtime is not configured")
	}
	var results []SearchResult
	var searchErr error
	done := make(chan struct{})
	s.messageTaskStarter.StartAuthorizedMessageSearch(func() {
		defer close(done)
		results, searchErr = s.messageExecutor.SearchAuthorizedMessages(ctx, scope, query, defaultSearchLimit)
	})
	select {
	case <-done:
		return results, searchErr
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
// Search выполняет глобальный поиск по всем сущностям с фильтрацией по правам доступа.
// Параллельно запрашивает PostgreSQL (tsvector) и MongoDB ($text), объединяет по score.
// partial=true — часть шардов упала, но есть результаты (W5-A4-007 / P2-BE-18).
func (s *SearchService) Search(ctx context.Context, query string, userID uuid.UUID, types []string, limit int) ([]SearchResult, bool, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	projectIDs, err := s.getUserProjectIDs(ctx, userID)
	if err != nil {
		return nil, false, fmt.Errorf("get user projects: %w", err)
	}

	typeSet := make(map[string]bool, len(types))
	for _, resultType := range types {
		typeSet[resultType] = true
	}
	searchAll := len(typeSet) == 0
	messageRequested := searchAll || typeSet["message"]
	messageScope := messenger.AuthorizedMessageSearchScope{}
	if messageRequested {
		resolvedScope, allowed, scopeErr := s.resolveAuthorizedMessageSearchScope(ctx, userID)
		if scopeErr != nil {
			return nil, false, scopeErr
		}
		if allowed && (s.messageExecutor == nil || s.messageTaskStarter == nil) {
			return nil, false, errors.New("authorized message search runtime is not configured")
		}
		messageRequested = allowed
		messageScope = resolvedScope
	}

	var (
		mu       sync.Mutex
		results  []SearchResult
		wg       sync.WaitGroup
		errs     []error
		shardOK  int
		shardErr int
	)
	appendResults := func(items []SearchResult, searchErr error) {
		mu.Lock()
		defer mu.Unlock()
		if searchErr != nil {
			errs = append(errs, searchErr)
			shardErr++
			return
		}
		shardOK++
		results = append(results, items...)
	}

	if searchAll || typeSet["issue"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := s.searchIssues(ctx, query, projectIDs)
			appendResults(items, searchErr)
		}()
	}
	if searchAll || typeSet["test_case"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := s.searchTestCases(ctx, query, projectIDs)
			appendResults(items, searchErr)
		}()
	}
	if searchAll || typeSet["user"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := s.searchUsers(ctx, query, userID)
			appendResults(items, searchErr)
		}()
	}
	if searchAll || typeSet["project"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := s.searchProjects(ctx, query, projectIDs)
			appendResults(items, searchErr)
		}()
	}
	if searchAll || typeSet["file"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := s.searchFiles(ctx, query, projectIDs)
			appendResults(items, searchErr)
		}()
	}
	if searchAll || typeSet["channel"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := s.searchChannels(ctx, query, projectIDs, userID)
			appendResults(items, searchErr)
		}()
	}
	if messageRequested {
		wg.Add(1)
		s.messageTaskStarter.StartAuthorizedMessageSearch(func() {
			defer wg.Done()
			items, searchErr := s.messageExecutor.SearchAuthorizedMessages(
				ctx, messageScope, query, defaultSearchLimit,
			)
			appendResults(items, searchErr)
		})
	}

	wg.Wait()
	joined := errors.Join(errs...)
	if joined != nil {
		slog.Warn("search shard errors", "count", len(errs), "errors", joined)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	if shardErr > 0 && shardOK == 0 {
		if joined != nil {
			return nil, false, fmt.Errorf("%w: %w", ErrSearchUnavailable, joined)
		}
		return nil, false, ErrSearchUnavailable
	}
	return results, shardErr > 0, nil
}
// getUserChannelIDs возвращает ID активных каналов мессенджера, в которых
// состоит пользователь. Используется SEC-H2 для скоупа Mongo $text-поиска.
// JOIN на channels + deleted_at IS NULL (W5-A4-004): после soft-delete канала
// строки membership остаются, но корзина не должна индексироваться в Ctrl+K.
func (s *SearchService) getUserChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.sqlDB.QueryContext(ctx,
		`SELECT cm.channel_id
		FROM channel_members cm
		INNER JOIN channels c ON c.id = cm.channel_id
		WHERE cm.user_id = $1 AND c.deleted_at IS NULL`,
		userID.String())
	if err != nil {
		return nil, fmt.Errorf("getUserChannelIDs: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var channelID uuid.UUID
		if err := rows.Scan(&channelID); err != nil {
			continue
		}
		ids = append(ids, channelID)
	}
	return ids, rows.Err()
}

// getUserProjectIDs возвращает ID проектов, в которых состоит пользователь.
// Использует raw SQL, т.к. ProjectMember хранит FK через Ent-edges.
// Имена колонок Ent: project_members — FK проекта, user_project_members — FK пользователя.
func (s *SearchService) getUserProjectIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.sqlDB.QueryContext(ctx,
		`SELECT project_members FROM project_members WHERE user_project_members = $1`,
		userID.String())
	if err != nil {
		return nil, fmt.Errorf("getUserProjectIDs: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var projectID uuid.UUID
		if err := rows.Scan(&projectID); err != nil {
			continue
		}
		ids = append(ids, projectID)
	}
	return ids, rows.Err()
}

// searchIssues ищет задачи по tsvector с фильтрацией по проектам пользователя.
func (s *SearchService) searchIssues(ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	rows, err := s.sqlDB.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, issue_key, title, project_issues,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM issues
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND project_issues = ANY($2)
			AND deleted_at IS NULL
		ORDER BY rank DESC LIMIT %d`, defaultSearchLimit),
		query, uuidSliceToPgArray(projectIDs))
	if err != nil {
		return nil, fmt.Errorf("поиск задач: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, issueKey, title, projID string
		var rank float64
		if err := rows.Scan(&id, &issueKey, &title, &projID, &rank); err != nil {
			continue
		}
		results = append(results, SearchResult{
			ID:        id,
			Type:      "issue",
			Title:     fmt.Sprintf("%s: %s", issueKey, title),
			Subtitle:  issueKey,
			ProjectID: projID,
			URL:       fmt.Sprintf("/projects/%s/tracker/issues/%s", projID, id),
			Score:     rank,
		})
	}
	return results, rows.Err()
}

// searchTestCases ищет тест-кейсы по tsvector с фильтрацией по проектам.
func (s *SearchService) searchTestCases(ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	rows, err := s.sqlDB.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, case_key, title, project_test_cases,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM test_cases
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND project_test_cases = ANY($2)
			AND deleted_at IS NULL
		ORDER BY rank DESC LIMIT %d`, defaultSearchLimit),
		query, uuidSliceToPgArray(projectIDs))
	if err != nil {
		return nil, fmt.Errorf("поиск тест-кейсов: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, caseKey, title, projID string
		var rank float64
		if err := rows.Scan(&id, &caseKey, &title, &projID, &rank); err != nil {
			continue
		}
		results = append(results, SearchResult{
			ID:        id,
			Type:      "test_case",
			Title:     fmt.Sprintf("%s: %s", caseKey, title),
			Subtitle:  caseKey,
			ProjectID: projID,
			URL:       fmt.Sprintf("/projects/%s/tms/cases/%s", projID, id),
			Score:     rank,
		})
	}
	return results, rows.Err()
}

// isSystemAdmin возвращает true, если у пользователя установлена системная
// роль admin. Используется searchUsers для обхода project-scoping
// (HI-12 follow-through: системные администраторы должны видеть всех
// пользователей платформы, например для управления учётками). Один lookup
// на запрос — допустимо, админы редки.
func (s *SearchService) isSystemAdmin(ctx context.Context, userID uuid.UUID) bool {
	if s.repo == nil {
		return false
	}
	ok, err := s.repo.IsSystemAdmin(ctx, userID)
	if err != nil {
		return false
	}
	return ok
}

// searchUsers ищет пользователей по tsvector, ограничивая выдачу теми,
// кто состоит хотя бы в одном общем с callerID проекте (HI-12). Кроме того,
// сам callerID всегда возвращается, если попадает в поисковую выдачу — это
// корректно, потому что пользователь имеет право увидеть «себя».
//
// HI-12 follow-through: системные администраторы (SystemRole=admin) обходят
// project-scoping и видят всех пользователей — нужно для управления
// учётками. Проверка через ent.User.Get — тот же паттерн, что в
// rbac.RequireAdmin (rbac/middleware.go:113-116) и
// rbac.GetEffectivePermissions (rbac/service.go:46) — единый источник прав.
//
// Колонки project_members: project_members = FK на project,
// user_project_members = FK на user. JOIN самой таблицы с собой даёт
// все user-id, делящие хотя бы одну запись project_members с callerID.
func (s *SearchService) searchUsers(ctx context.Context, query string, callerID uuid.UUID) ([]SearchResult, error) {
	// Админский обход project-scoping.
	if s.isSystemAdmin(ctx, callerID) {
		return s.searchUsersUnscoped(ctx, query)
	}

	rows, err := s.sqlDB.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, name, email, display_name,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM users
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		  AND (
		        id = $2
		        OR id IN (
		            SELECT DISTINCT pm2.user_project_members
		            FROM project_members pm1
		            JOIN project_members pm2
		              ON pm1.project_members = pm2.project_members
		            WHERE pm1.user_project_members = $2
		        )
		      )
		ORDER BY rank DESC LIMIT %d`, defaultSearchLimit),
		query, callerID.String())
	if err != nil {
		return nil, fmt.Errorf("поиск пользователей: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name, email string
		var displayName *string
		var rank float64
		if err := rows.Scan(&id, &name, &email, &displayName, &rank); err != nil {
			continue
		}
		title := name
		if title == "" && displayName != nil {
			title = *displayName
		}
		if title == "" {
			title = email
		}
		results = append(results, SearchResult{
			ID:       id,
			Type:     "user",
			Title:    title,
			Subtitle: email,
			URL:      fmt.Sprintf("/intranet/directory?user=%s", id),
			Score:    rank,
		})
	}
	return results, rows.Err()
}

// searchUsersUnscoped — поиск пользователей без project-scoping.
// Вызывается только для системных администраторов (HI-12 follow-through).
func (s *SearchService) searchUsersUnscoped(ctx context.Context, query string) ([]SearchResult, error) {
	rows, err := s.sqlDB.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, name, email, display_name,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM users
		WHERE search_vector @@ plainto_tsquery('simple', $1)
		ORDER BY rank DESC LIMIT %d`, defaultSearchLimit),
		query)
	if err != nil {
		return nil, fmt.Errorf("поиск пользователей (admin): %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name, email string
		var displayName *string
		var rank float64
		if err := rows.Scan(&id, &name, &email, &displayName, &rank); err != nil {
			continue
		}
		title := name
		if title == "" && displayName != nil {
			title = *displayName
		}
		if title == "" {
			title = email
		}
		results = append(results, SearchResult{
			ID:       id,
			Type:     "user",
			Title:    title,
			Subtitle: email,
			URL:      fmt.Sprintf("/intranet/directory?user=%s", id),
			Score:    rank,
		})
	}
	return results, rows.Err()
}

// searchProjects ищет проекты по tsvector с фильтрацией по членству.
func (s *SearchService) searchProjects(ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	rows, err := s.sqlDB.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, "key", name,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM projects
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND id = ANY($2)
			AND deleted_at IS NULL
		ORDER BY rank DESC LIMIT %d`, defaultSearchLimit),
		query, uuidSliceToPgArray(projectIDs))
	if err != nil {
		return nil, fmt.Errorf("поиск проектов: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, key, name string
		var rank float64
		if err := rows.Scan(&id, &key, &name, &rank); err != nil {
			continue
		}
		results = append(results, SearchResult{
			ID:       id,
			Type:     "project",
			Title:    name,
			Subtitle: key,
			URL:      fmt.Sprintf("/projects/%s", id),
			Score:    rank,
		})
	}
	return results, rows.Err()
}

// searchFiles ищет файлы по tsvector с фильтрацией по проектам.
func (s *SearchService) searchFiles(ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	rows, err := s.sqlDB.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, filename, project_files,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM file_metadata
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND project_files = ANY($2)
			AND deleted_at IS NULL
			AND status = 'confirmed'
		ORDER BY rank DESC LIMIT %d`, defaultSearchLimit),
		query, uuidSliceToPgArray(projectIDs))
	if err != nil {
		return nil, fmt.Errorf("поиск файлов: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, filename, projID string
		var rank float64
		if err := rows.Scan(&id, &filename, &projID, &rank); err != nil {
			continue
		}
		results = append(results, SearchResult{
			ID:        id,
			Type:      "file",
			Title:     filename,
			ProjectID: projID,
			URL:       fmt.Sprintf("/api/files/%s", id),
			Score:     rank,
		})
	}
	return results, rows.Err()
}

// searchChannels ищет каналы по tsvector с фильтрацией по проектам.
// Fallback на ILIKE если tsvector-колонка отсутствует.
//
// P1-BE-32 (W5-A4-001/002): soft-deleted исключены; public виден участникам
// проекта; private/group — только при membership в channel_members (NF-03/NFR-19).
func (s *SearchService) searchChannels(ctx context.Context, query string, projectIDs []uuid.UUID, userID uuid.UUID) ([]SearchResult, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	// Пробуем tsvector поиск (если search_vector создан через indexer)
	rows, err := s.sqlDB.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, name, channel_type, project_id,
			ts_rank(to_tsvector('simple', name), plainto_tsquery('simple', $1)) AS rank
		FROM channels
		WHERE to_tsvector('simple', name) @@ plainto_tsquery('simple', $1)
			AND project_id = ANY($2)
			AND deleted_at IS NULL
			AND (
				channel_type = 'public'
				OR id IN (SELECT channel_id FROM channel_members WHERE user_id = $3)
			)
		ORDER BY rank DESC LIMIT %d`, defaultSearchLimit),
		query, uuidSliceToPgArray(projectIDs), userID.String())
	if err != nil {
		// Fallback на ILIKE если tsvector не работает
		rows, err = s.sqlDB.QueryContext(ctx,
			fmt.Sprintf(`SELECT id, name, channel_type, project_id, 0.5::float8 AS rank
			FROM channels
			WHERE name ILIKE '%%' || $1 || '%%'
				AND project_id = ANY($2)
				AND deleted_at IS NULL
				AND (
					channel_type = 'public'
					OR id IN (SELECT channel_id FROM channel_members WHERE user_id = $3)
				)
			ORDER BY name ASC LIMIT %d`, defaultSearchLimit),
			query, uuidSliceToPgArray(projectIDs), userID.String())
		if err != nil {
			return nil, fmt.Errorf("поиск каналов: %w", err)
		}
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, name, channelType string
		var projID *string
		var rank float64
		if err := rows.Scan(&id, &name, &channelType, &projID, &rank); err != nil {
			continue
		}
		pid := ""
		if projID != nil {
			pid = *projID
		}
		results = append(results, SearchResult{
			ID:        id,
			Type:      "channel",
			Title:     "#" + name,
			Subtitle:  channelType,
			ProjectID: pid,
			URL:       fmt.Sprintf("/messenger/channels/%s", id),
			Score:     rank,
		})
	}
	return results, rows.Err()
}

// searchMessages ищет сообщения в MongoDB по $text-индексу.
//
// SEC-H2: $text-запрос ограничен каналами, в которых состоит вызывающий
// пользователь. Раньше /api/search?type=message возвращал сообщения из
// любого канала (HI-01 закрыл только messenger-handler-путь, но search
// шёл напрямую в Mongo без scope). Без права хотя бы в одном канале —
// ранний выход с пустым результатом.
func (s *SearchService) searchMessages(ctx context.Context, query string, channelIDs []uuid.UUID, limit int) ([]SearchResult, error) {
	if s.mongoDB == nil || len(channelIDs) == 0 {
		return nil, nil
	}
	if limit <= 0 || limit > defaultSearchLimit {
		limit = defaultSearchLimit
	}
	coll := s.mongoDB.Collection("messages")
	// P1-BE-32 / W5-A4-003: soft-deleted сообщения (messenger SoftDeleteMessage)
	// не должны попадать в Ctrl+K — тот же фильтр, что ListMessages.
	filter := bson.M{
		"channel_id": bson.M{"$in": channelIDs},
		"deleted_at": nil,
		"$text":      bson.M{"$search": query},
	}
	opts := options.Find().
		SetProjection(bson.M{
			"score":        bson.M{"$meta": "textScore"},
			"_id":          1,
			"content_text": 1,
			"channel_id":   1,
			"sender_id":    1,
		}).
		SetSort(bson.M{"score": bson.M{"$meta": "textScore"}}).
		SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("поиск сообщений: %w", err)
	}
	defer cursor.Close(ctx)

	var results []SearchResult
	for cursor.Next(ctx) {
		var doc struct {
			ID          string  `bson:"_id"`
			ContentText string  `bson:"content_text"`
			ChannelID   string  `bson:"channel_id"`
			SenderID    string  `bson:"sender_id"`
			Score       float64 `bson:"score"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		title := truncateRunes(doc.ContentText, 100)
		results = append(results, SearchResult{
			ID:       doc.ID,
			Type:     "message",
			Title:    title,
			Subtitle: doc.ChannelID,
			URL:      fmt.Sprintf("/messenger/channels/%s", doc.ChannelID),
			Score:    doc.Score * 0.8,
		})
	}
	return results, cursor.Err()
}

// CreateSavedFilter создаёт новый сохранённый фильтр.
// A4-003: IsProjectMember before create; else ErrFilterForbidden.
func (s *SearchService) CreateSavedFilter(ctx context.Context, userID uuid.UUID, req SavedFilterCreateRequest) (*SavedFilterDTO, error) {
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("некорректный project_id: %w", err)
	}

	if err := s.requireProjectMembership(ctx, userID, projectID); err != nil {
		return nil, err
	}

	filter, err := s.repo.CreateSavedFilter(ctx, CreateSavedFilterInput{
		Name:         req.Name,
		FilterConfig: req.FilterConfig,
		OwnerID:      userID,
		ProjectID:    projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось создать фильтр: %w", err)
	}

	return savedFilterToDTO(filter), nil
}

// requireProjectMembership — fail-closed: non-member → ErrFilterForbidden.
func (s *SearchService) requireProjectMembership(ctx context.Context, userID, projectID uuid.UUID) error {
	isMember, err := s.repo.IsProjectMember(ctx, userID, projectID)
	if err != nil {
		return fmt.Errorf("project membership check: %w", err)
	}
	if !isMember {
		return ErrFilterForbidden
	}
	return nil
}

// ListSavedFilters возвращает фильтры пользователя и расшаренные фильтры проекта.
func (s *SearchService) ListSavedFilters(ctx context.Context, userID, projectID uuid.UUID) ([]SavedFilterDTO, error) {
	if err := s.requireProjectMembership(ctx, userID, projectID); err != nil {
		return nil, err
	}

	filters, err := s.repo.ListSavedFilters(ctx, userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить фильтры: %w", err)
	}

	dtos := make([]SavedFilterDTO, len(filters))
	for i, f := range filters {
		dtos[i] = *savedFilterToDTO(f)
	}
	return dtos, nil
}

// DeleteSavedFilter удаляет сохранённый фильтр (только владелец).
func (s *SearchService) DeleteSavedFilter(ctx context.Context, userID uuid.UUID, filterID uuid.UUID) error {
	filter, err := s.repo.GetSavedFilter(ctx, filterID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrFilterNotFound
		}
		return fmt.Errorf("фильтр не найден: %w", err)
	}
	if filter.OwnerID != userID {
		return ErrFilterForbidden
	}
	return s.repo.DeleteSavedFilter(ctx, filterID)
}

// ShareFilter переключает флаг расшаривания фильтра (только владелец).
func (s *SearchService) ShareFilter(ctx context.Context, userID uuid.UUID, filterID uuid.UUID, shared bool) (*SavedFilterDTO, error) {
	filter, err := s.repo.GetSavedFilter(ctx, filterID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrFilterNotFound
		}
		return nil, fmt.Errorf("фильтр не найден: %w", err)
	}
	if filter.OwnerID != userID {
		return nil, ErrFilterForbidden
	}

	updated, err := s.repo.UpdateSavedFilterShared(ctx, filterID, shared)
	if err != nil {
		return nil, fmt.Errorf("не удалось обновить фильтр: %w", err)
	}
	return savedFilterToDTO(updated), nil
}

// savedFilterToDTO конвертирует Ent-сущность в DTO.
func savedFilterToDTO(f *ent.SavedFilter) *SavedFilterDTO {
	return &SavedFilterDTO{
		ID:           f.ID.String(),
		Name:         f.Name,
		FilterConfig: f.FilterConfig,
		Shared:       f.Shared,
		OwnerID:      f.OwnerID.String(),
		ProjectID:    f.ProjectID.String(),
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}

// uuidSliceToPgArray конвертирует срез UUID в строку PostgreSQL-массива.
func uuidSliceToPgArray(ids []uuid.UUID) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.String()
	}
	return "{" + strings.Join(parts, ",") + "}"
}
