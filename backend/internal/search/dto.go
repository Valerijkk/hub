package search

import "time"

// SearchResult представляет один результат глобального поиска.
type SearchResult struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"` // issue, test_case, user, project, channel, file, message
	Title     string  `json:"title"`
	Subtitle  string  `json:"subtitle,omitempty"`
	ProjectID string  `json:"project_id,omitempty"`
	URL       string  `json:"url"`
	Score     float64 `json:"score"`
}

// SearchRequest описывает параметры запроса глобального поиска.
type SearchRequest struct {
	Query string   `json:"query" query:"q" validate:"required,min=1,max=500"`
	Types []string `json:"types,omitempty" query:"type"`
	Limit int      `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=50"`
}

// SearchResponse содержит результаты глобального поиска.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	// Partial — true, если хотя бы один шард упал, но есть результаты (W5-A4-007).
	Partial bool `json:"partial,omitempty"`
}

// SavedFilterDTO представляет сохранённый фильтр в ответе API.
type SavedFilterDTO struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	FilterConfig map[string]interface{} `json:"filter_config"`
	Shared       bool                   `json:"shared"`
	OwnerID      string                 `json:"owner_id"`
	ProjectID    string                 `json:"project_id"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// SavedFilterCreateRequest описывает запрос на создание сохранённого фильтра.
type SavedFilterCreateRequest struct {
	Name         string                 `json:"name" validate:"required,min=1,max=200"`
	FilterConfig map[string]interface{} `json:"filter_config" validate:"required"`
	ProjectID    string                 `json:"project_id" validate:"required"`
}

// SavedFilterShareRequest описывает запрос на переключение расшаривания фильтра.
type SavedFilterShareRequest struct {
	Shared bool `json:"shared"`
}

// SavedFilterListResponse содержит список сохранённых фильтров.
type SavedFilterListResponse struct {
	Filters []SavedFilterDTO `json:"filters"`
}
