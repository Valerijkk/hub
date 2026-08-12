package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	searchmod "github.com/valerii/hub/backend/internal/search"
)

// DocSearchProvider — RAG-lite context из Hub search (M3-021).
type DocSearchProvider interface {
	SearchContext(ctx context.Context, userID uuid.UUID, query string) (string, error)
}

// hubSearchBackend — subset SearchService для doc context (stub в tests).
type hubSearchBackend interface {
	Search(ctx context.Context, query string, userID uuid.UUID, types []string, limit int) ([]searchmod.SearchResult, bool, error)
}

// SearchDocProvider адаптирует SearchService для @hub-ai /ai docs.
type SearchDocProvider struct {
	search hubSearchBackend
}

// NewSearchDocProvider — doc context из search.
func NewSearchDocProvider(search hubSearchBackend) *SearchDocProvider {
	return &SearchDocProvider{search: search}
}

// SearchContext — snippets из issues, test cases, messages.
func (p *SearchDocProvider) SearchContext(ctx context.Context, userID uuid.UUID, query string) (string, error) {
	if p == nil || p.search == nil {
		return "", fmt.Errorf("search not configured")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	results, _, err := p.search.Search(ctx, query, userID, []string{"issue", "test_case", "message"}, 8)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "Релевантных результатов в Hub не найдено.", nil
	}
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "[%d] (%s) %s\n", i+1, r.Type, r.Title)
		if r.Subtitle != "" {
			fmt.Fprintf(&b, "    %s\n", r.Subtitle)
		}
		if r.URL != "" {
			fmt.Fprintf(&b, "    %s\n", r.URL)
		}
	}
	return strings.TrimSpace(b.String()), nil
}
