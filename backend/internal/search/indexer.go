package search

import (
	"context"
	"database/sql"
	"log/slog"
)

// EnsureSearchIndexes создаёт tsvector-колонки и GIN-индексы для полнотекстового поиска.
// Использует ADD COLUMN IF NOT EXISTS для идемпотентности. Вызывается при старте сервера.
// Конфигурация 'simple' подходит для русского и английского текста без стемминга.
func EnsureSearchIndexes(ctx context.Context, sqlDB *sql.DB) error {
	statements := []string{
		// issues: search_vector на основе issue_key, title, description_text
		`ALTER TABLE issues ADD COLUMN IF NOT EXISTS search_vector tsvector
			GENERATED ALWAYS AS (
				setweight(to_tsvector('simple', coalesce(issue_key, '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(description_text, '')), 'B')
			) STORED`,
		`CREATE INDEX IF NOT EXISTS idx_issues_search_vector ON issues USING GIN (search_vector)`,

		// test_cases: search_vector на основе case_key, title, description
		`ALTER TABLE test_cases ADD COLUMN IF NOT EXISTS search_vector tsvector
			GENERATED ALWAYS AS (
				setweight(to_tsvector('simple', coalesce(case_key, '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(description, '')), 'B')
			) STORED`,
		`CREATE INDEX IF NOT EXISTS idx_test_cases_search_vector ON test_cases USING GIN (search_vector)`,

		// users: search_vector на основе name, email, display_name
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS search_vector tsvector
			GENERATED ALWAYS AS (
				setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(email, '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(display_name, '')), 'B')
			) STORED`,
		`CREATE INDEX IF NOT EXISTS idx_users_search_vector ON users USING GIN (search_vector)`,

		// projects: search_vector на основе key, name, description
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS search_vector tsvector
			GENERATED ALWAYS AS (
				setweight(to_tsvector('simple', coalesce("key", '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(name, '')), 'A') ||
				setweight(to_tsvector('simple', coalesce(description, '')), 'B')
			) STORED`,
		`CREATE INDEX IF NOT EXISTS idx_projects_search_vector ON projects USING GIN (search_vector)`,

		// file_metadata: search_vector на основе filename
		`ALTER TABLE file_metadata ADD COLUMN IF NOT EXISTS search_vector tsvector
			GENERATED ALWAYS AS (
				to_tsvector('simple', coalesce(filename, ''))
			) STORED`,
		`CREATE INDEX IF NOT EXISTS idx_file_metadata_search_vector ON file_metadata USING GIN (search_vector)`,
	}

	for _, stmt := range statements {
		if _, err := sqlDB.ExecContext(ctx, stmt); err != nil {
			slog.Error("не удалось выполнить SQL для поискового индекса", "error", err, "sql", stmt[:80])
			return err
		}
	}

	slog.Info("поисковые tsvector-индексы созданы/проверены")
	return nil
}
