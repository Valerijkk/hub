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

// ErrFilterForbidden — A4-003/A4-004: 403 без утечки причины (нет membership).
var ErrFilterForbidden = errors.New("forbidden")

// SearchService реализует глобальный поиск по всем сущностям платформы.
// Выполняет параллельные запросы к PostgreSQL (tsvector) и MongoDB ($text),
// затем объединяет результаты по релевантности.
//
// SC-14-3: typed Ent-операции (SystemRole lookup, SavedFilter CRUD) вынесены
// в Repository. Сам сервис держит только sqlDB (raw tsvector) и mongoDB ($text) —
// эти БД-уровневые драйверы концептуально ниже ORM и для них репозиторий-
// абстракция избыточна.
type SearchService struct {
	repo    Repository
	sqlDB   *sql.DB
	mongoDB *mongo.Database
}

// NewSearchService создаёт новый сервис поиска.
// sqlDB — прямое подключение к PostgreSQL для raw-запросов с tsvector.
func NewSearchService(repo Repository, sqlDB *sql.DB, mongoDB *mongo.Database) *SearchService {
	return &SearchService{repo: repo, sqlDB: sqlDB, mongoDB: mongoDB}
}

// Search выполняет глобальный поиск по всем сущностям с фильтрацией по правам доступа.
// Параллельно запрашивает PostgreSQL (tsvector) и MongoDB ($text), объединяет по score.
// partial=true — часть шардов упала, но есть результаты (W5-A4-007 / P2-BE-18).
func (s *SearchService) Search(ctx context.Context, query string, userID uuid.UUID, types []string, limit int) ([]SearchResult, bool, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Получаем ID проектов пользователя для фильтрации по правам
	projectIDs, err := s.getUserProjectIDs(ctx, userID)
	if err != nil {
		return nil, false, fmt.Errorf("не удалось получить проекты пользователя: %w", err)
	}

	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	searchAll := len(typeSet) == 0

	var (
		mu       sync.Mutex
		results  []SearchResult
		wg       sync.WaitGroup
		errs     []error
		shardOK  int
		shardErr int
	)

	// Вспомогательная функция для добавления результатов потокобезопасно
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

	// Параллельный поиск по PostgreSQL-сущностям
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
			// HI-12: ограничиваем выдачу пользователями, разделяющими хотя
			// бы один проект с вызывающим — иначе search превращается в
			// глобальный email/name enumerator.
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
			// P1-BE-32 / NF-03: private/group требуют membership; soft-delete исключаем.
			items, searchErr := s.searchChannels(ctx, query, projectIDs, userID)
			appendResults(items, searchErr)
		}()
	}

	// Поиск в MongoDB (сообщения)
	if searchAll || typeSet["message"] {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, searchErr := s.searchMessages(ctx, query, userID)
			appendResults(items, searchErr)
		}()
	}

	wg.Wait()

	// W5-A4-007 / P2-BE-18: не глотать ошибки шардов — partial flag или hard error.
	joined := errors.Join(errs...)
	if joined != nil {
		slog.Warn("ошибки при поиске", "count", len(errs), "errors", joined)
	}

	// Сортировка по убыванию score и ограничение по limit
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}

	if shardErr > 0 && shardOK == 0 {
		// Все шарды упали: Ctrl+K получает partial+пусто, не 500 (деградация).
		// Мутация «partial flag» закрыта; hard-error при total outage — follow-up.
		return results, true, nil
	}
	partial := shardErr > 0
	return results, partial, nil
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
func (s *SearchService) searchMessages(ctx context.Context, query string, userID uuid.UUID) ([]SearchResult, error) {
	if s.mongoDB == nil {
		return nil, nil
	}

	channelIDs, err := s.getUserChannelIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("получение каналов пользователя: %w", err)
	}
	if len(channelIDs) == 0 {
		return nil, nil
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
		SetLimit(int64(defaultSearchLimit))

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
		return fmt.Errorf("фильтр не найден: %w", err)
	}
	if filter.OwnerID != userID {
		return fmt.Errorf("доступ запрещён: только владелец может удалить фильтр")
	}
	return s.repo.DeleteSavedFilter(ctx, filterID)
}

// ShareFilter переключает флаг расшаривания фильтра (только владелец).
func (s *SearchService) ShareFilter(ctx context.Context, userID uuid.UUID, filterID uuid.UUID, shared bool) (*SavedFilterDTO, error) {
	filter, err := s.repo.GetSavedFilter(ctx, filterID)
	if err != nil {
		return nil, fmt.Errorf("фильтр не найден: %w", err)
	}
	if filter.OwnerID != userID {
		return nil, fmt.Errorf("доступ запрещён: только владелец может изменить расшаривание")
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
