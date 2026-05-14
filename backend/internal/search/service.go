// Package search реализует глобальный полнотекстовый поиск по PostgreSQL (tsvector) и MongoDB ($text).
package search

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/savedfilter"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SearchService реализует глобальный поиск по всем сущностям платформы.
// Выполняет параллельные запросы к PostgreSQL (tsvector) и MongoDB ($text),
// затем объединяет результаты по релевантности.
type SearchService struct {
	db      *ent.Client
	sqlDB   *sql.DB
	mongoDB *mongo.Database
}

// NewSearchService создаёт новый сервис поиска.
// sqlDB — прямое подключение к PostgreSQL для raw-запросов с tsvector.
func NewSearchService(db *ent.Client, sqlDB *sql.DB, mongoDB *mongo.Database) *SearchService {
	return &SearchService{db: db, sqlDB: sqlDB, mongoDB: mongoDB}
}

// Search выполняет глобальный поиск по всем сущностям с фильтрацией по правам доступа.
// Параллельно запрашивает PostgreSQL (tsvector) и MongoDB ($text), объединяет по score.
func (s *SearchService) Search(ctx context.Context, query string, userID uuid.UUID, types []string, limit int) ([]SearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Получаем ID проектов пользователя для фильтрации по правам
	projectIDs, err := s.getUserProjectIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить проекты пользователя: %w", err)
	}

	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	searchAll := len(typeSet) == 0

	var (
		mu      sync.Mutex
		results []SearchResult
		wg      sync.WaitGroup
		errs    []error
	)

	// Вспомогательная функция для добавления результатов потокобезопасно
	appendResults := func(items []SearchResult, searchErr error) {
		mu.Lock()
		defer mu.Unlock()
		if searchErr != nil {
			errs = append(errs, searchErr)
			return
		}
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
			items, searchErr := s.searchChannels(ctx, query, projectIDs)
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

	if len(errs) > 0 {
		slog.Warn("ошибки при поиске", "count", len(errs), "first", errs[0])
	}

	// Сортировка по убыванию score и ограничение по limit
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// getUserProjectIDs возвращает ID проектов, в которых состоит пользователь.
// Использует raw SQL, т.к. ProjectMember хранит FK через Ent-edges.
// Ent column names: project_members = project FK, user_project_members = user FK.
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
		`SELECT id, issue_key, title, project_issues,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM issues
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND project_issues = ANY($2)
		ORDER BY rank DESC LIMIT 10`,
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
		`SELECT id, case_key, title, project_test_cases,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM test_cases
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND project_test_cases = ANY($2)
		ORDER BY rank DESC LIMIT 10`,
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

// searchUsers ищет пользователей по tsvector, ограничивая выдачу теми,
// кто состоит хотя бы в одном общем с callerID проекте (HI-12). Кроме того,
// сам callerID всегда возвращается, если попадает в поисковую выдачу — это
// корректно, потому что пользователь имеет право увидеть «себя».
//
// Колонки project_members: project_members = FK на project,
// user_project_members = FK на user. JOIN самой таблицы с собой даёт
// все user-id, делящие хотя бы одну запись project_members с callerID.
func (s *SearchService) searchUsers(ctx context.Context, query string, callerID uuid.UUID) ([]SearchResult, error) {
	rows, err := s.sqlDB.QueryContext(ctx,
		`SELECT id, name, email, display_name,
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
		ORDER BY rank DESC LIMIT 10`,
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
			URL:      fmt.Sprintf("/settings/users/%s", id),
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
		`SELECT id, "key", name,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM projects
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND id = ANY($2)
		ORDER BY rank DESC LIMIT 10`,
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
		`SELECT id, filename, project_files,
			ts_rank(search_vector, plainto_tsquery('simple', $1)) AS rank
		FROM file_metadata
		WHERE search_vector @@ plainto_tsquery('simple', $1)
			AND project_files = ANY($2)
		ORDER BY rank DESC LIMIT 10`,
		query, uuidSliceToPgArray(projectIDs))
	if err != nil {
		return nil, fmt.Errorf("поиск файлов: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var id, filename string
		var projID *string
		var rank float64
		if err := rows.Scan(&id, &filename, &projID, &rank); err != nil {
			continue
		}
		pid := ""
		if projID != nil {
			pid = *projID
		}
		results = append(results, SearchResult{
			ID:        id,
			Type:      "file",
			Title:     filename,
			ProjectID: pid,
			URL:       fmt.Sprintf("/api/files/%s", id),
			Score:     rank,
		})
	}
	return results, rows.Err()
}

// searchChannels ищет каналы по tsvector с фильтрацией по проектам.
// Fallback на ILIKE если tsvector-колонка отсутствует.
func (s *SearchService) searchChannels(ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}

	// Пробуем tsvector поиск (если search_vector создан через indexer)
	rows, err := s.sqlDB.QueryContext(ctx,
		`SELECT id, name, channel_type, project_id,
			ts_rank(to_tsvector('simple', name), plainto_tsquery('simple', $1)) AS rank
		FROM channels
		WHERE to_tsvector('simple', name) @@ plainto_tsquery('simple', $1)
			AND project_id = ANY($2)
		ORDER BY rank DESC LIMIT 10`,
		query, uuidSliceToPgArray(projectIDs))
	if err != nil {
		// Fallback на ILIKE если tsvector не работает
		rows, err = s.sqlDB.QueryContext(ctx,
			`SELECT id, name, channel_type, project_id, 0.5::float8 AS rank
			FROM channels
			WHERE name ILIKE '%' || $1 || '%'
				AND project_id = ANY($2)
			ORDER BY name ASC LIMIT 10`,
			query, uuidSliceToPgArray(projectIDs))
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
func (s *SearchService) searchMessages(ctx context.Context, query string, _ uuid.UUID) ([]SearchResult, error) {
	if s.mongoDB == nil {
		return nil, nil
	}

	coll := s.mongoDB.Collection("messages")
	filter := bson.M{
		"$text": bson.M{"$search": query},
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
		SetLimit(10)

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
		title := doc.ContentText
		if len(title) > 100 {
			title = title[:100] + "..."
		}
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
func (s *SearchService) CreateSavedFilter(ctx context.Context, userID uuid.UUID, req SavedFilterCreateRequest) (*SavedFilterDTO, error) {
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("некорректный project_id: %w", err)
	}

	filter, err := s.db.SavedFilter.Create().
		SetName(req.Name).
		SetFilterConfig(req.FilterConfig).
		SetOwnerID(userID).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать фильтр: %w", err)
	}

	return savedFilterToDTO(filter), nil
}

// ListSavedFilters возвращает фильтры пользователя и расшаренные фильтры проекта.
func (s *SearchService) ListSavedFilters(ctx context.Context, userID, projectID uuid.UUID) ([]SavedFilterDTO, error) {
	filters, err := s.db.SavedFilter.Query().
		Where(
			savedfilter.ProjectID(projectID),
			savedfilter.Or(
				savedfilter.OwnerID(userID),
				savedfilter.Shared(true),
			),
		).
		All(ctx)
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
	filter, err := s.db.SavedFilter.Get(ctx, filterID)
	if err != nil {
		return fmt.Errorf("фильтр не найден: %w", err)
	}
	if filter.OwnerID != userID {
		return fmt.Errorf("доступ запрещён: только владелец может удалить фильтр")
	}
	return s.db.SavedFilter.DeleteOneID(filterID).Exec(ctx)
}

// ShareFilter переключает флаг расшаривания фильтра (только владелец).
func (s *SearchService) ShareFilter(ctx context.Context, userID uuid.UUID, filterID uuid.UUID, shared bool) (*SavedFilterDTO, error) {
	filter, err := s.db.SavedFilter.Get(ctx, filterID)
	if err != nil {
		return nil, fmt.Errorf("фильтр не найден: %w", err)
	}
	if filter.OwnerID != userID {
		return nil, fmt.Errorf("доступ запрещён: только владелец может изменить расшаривание")
	}

	updated, err := filter.Update().SetShared(shared).Save(ctx)
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
