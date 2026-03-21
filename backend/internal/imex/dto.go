package imex

import "time"

// FieldMapping описывает маппинг полей при импорте из CSV.
type FieldMapping struct {
	SourceColumn string `json:"source_column"`
	TargetField  string `json:"target_field"`
}

// ImportPreviewRow — одна строка предпросмотра импорта.
type ImportPreviewRow struct {
	Fields map[string]string `json:"fields"`
}

// ImportPreviewResponse — ответ с предпросмотром (первые 5 строк).
type ImportPreviewResponse struct {
	Headers []string           `json:"headers"`
	Rows    []ImportPreviewRow `json:"rows"`
	Total   int                `json:"total"`
}

// ImportResultResponse — результат импорта.
type ImportResultResponse struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

// ExportFilter — фильтры для экспорта.
type ExportFilter struct {
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
	Type     string `json:"type,omitempty"`
}

// IssueExportRow — строка экспорта задачи.
type IssueExportRow struct {
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Priority  string    `json:"priority"`
	Status    string    `json:"status"`
	Assignee  string    `json:"assignee"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
