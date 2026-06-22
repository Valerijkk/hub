package imex

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/issue"
)

// ImExService реализует импорт и экспорт задач, прогонов и проектов.
// Поддерживает CSV, JSON, Excel, PDF и ZIP форматы.
type ImExService struct {
	repo Repository
}

// NewImExService создаёт новый сервис импорта/экспорта.
func NewImExService(repo Repository) *ImExService {
	return &ImExService{repo: repo}
}

// allocateIssueKey — атомарный счётчик в repository (HI-06).
func (s *ImExService) allocateIssueKey(ctx context.Context, projectID uuid.UUID) (int, string, uuid.UUID, error) {
	return s.repo.AllocateIssueKey(ctx, projectID)
}

// ImportIssuesCSV импортирует задачи из CSV-файла с маппингом полей.
// В режиме preview возвращает первые 5 строк для подтверждения.
func (s *ImExService) ImportIssuesCSV(ctx context.Context, projectID uuid.UUID, data []byte, mappings []FieldMapping, preview bool) (*ImportPreviewResponse, *ImportResultResponse, error) {
	reader := csv.NewReader(bytes.NewReader(data))

	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка чтения заголовков CSV: %w", err)
	}

	if preview {
		var rows []ImportPreviewRow
		for i := 0; i < 5; i++ {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			row := ImportPreviewRow{Fields: make(map[string]string)}
			for j, h := range headers {
				if j < len(record) {
					row.Fields[h] = record[j]
				}
			}
			rows = append(rows, row)
		}

		total := len(rows)
		for {
			if _, err := reader.Read(); err != nil {
				break
			}
			total++
		}

		return &ImportPreviewResponse{
			Headers: headers,
			Rows:    rows,
			Total:   total,
		}, nil, nil
	}

	fieldMap := buildFieldMap(mappings, headers)
	imported := 0
	skipped := 0
	var errs []string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skipped++
			continue
		}

		title := getField(record, fieldMap, "title")
		if title == "" {
			skipped++
			errs = append(errs, fmt.Sprintf("строка %d: отсутствует title", imported+skipped+1))
			continue
		}

		issueType := getField(record, fieldMap, "type")
		if issueType == "" {
			issueType = "task"
		}

		priority := getField(record, fieldMap, "priority")
		if priority == "" {
			priority = "medium"
		}

		nextNum, issueKey, statusID, err := s.allocateIssueKey(ctx, projectID)
		if err != nil {
			skipped++
			slog.Error("не удалось зарезервировать issue key", "error", err)
			continue
		}

		_, err = s.repo.CreateIssue(ctx, CreateIssueInput{
			Title:       title,
			IssueType:   issue.IssueType(issueType),
			Priority:    issue.Priority(priority),
			IssueNumber: nextNum,
			IssueKey:    issueKey,
			ProjectID:   projectID,
			StatusID:    statusID,
		})
		if err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("строка %d: %s", imported+skipped+1, err.Error()))
			continue
		}

		imported++
	}

	return nil, &ImportResultResponse{
		Imported: imported,
		Skipped:  skipped,
		Errors:   errs,
	}, nil
}

// ImportIssuesJSON импортирует задачи из JSON-массива.
func (s *ImExService) ImportIssuesJSON(ctx context.Context, projectID uuid.UUID, data []byte) (*ImportResultResponse, error) {
	var items []map[string]interface{}
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	imported := 0
	skipped := 0
	var errs []string

	for i, item := range items {
		title, _ := item["title"].(string)
		if title == "" {
			skipped++
			errs = append(errs, fmt.Sprintf("элемент %d: отсутствует title", i))
			continue
		}

		issueType, _ := item["type"].(string)
		if issueType == "" {
			issueType = "task"
		}

		priority, _ := item["priority"].(string)
		if priority == "" {
			priority = "medium"
		}

		nextNum, issueKey, statusID, err := s.allocateIssueKey(ctx, projectID)
		if err != nil {
			skipped++
			slog.Error("не удалось зарезервировать issue key", "error", err)
			continue
		}

		_, err = s.repo.CreateIssue(ctx, CreateIssueInput{
			Title:       title,
			IssueType:   issue.IssueType(issueType),
			Priority:    issue.Priority(priority),
			IssueNumber: nextNum,
			IssueKey:    issueKey,
			ProjectID:   projectID,
			StatusID:    statusID,
		})
		if err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("элемент %d: %s", i, err.Error()))
			continue
		}

		imported++
	}

	return &ImportResultResponse{
		Imported: imported,
		Skipped:  skipped,
		Errors:   errs,
	}, nil
}

// ExportIssuesCSV экспортирует задачи проекта в CSV.
func (s *ImExService) ExportIssuesCSV(ctx context.Context, projectID uuid.UUID) ([]byte, error) {
	issues, err := s.repo.ListProjectIssuesForExport(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	_ = w.Write([]string{"Key", "Title", "Type", "Priority", "Status", "Assignee", "Created", "Updated"})

	for _, iss := range issues {
		status := ""
		if iss.Edges.Status != nil {
			status = iss.Edges.Status.Name
		}
		assignee := ""
		if iss.Edges.Assignee != nil {
			if iss.Edges.Assignee.Name != "" {
				assignee = iss.Edges.Assignee.Name
			} else {
				assignee = iss.Edges.Assignee.Email
			}
		}

		_ = w.Write([]string{
			iss.IssueKey,
			iss.Title,
			string(iss.IssueType),
			string(iss.Priority),
			status,
			assignee,
			iss.CreatedAt.Format(time.RFC3339),
			iss.UpdatedAt.Format(time.RFC3339),
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}

// ExportIssuesExcel экспортирует задачи проекта в XLSX (упрощённая версия — CSV).
func (s *ImExService) ExportIssuesExcel(ctx context.Context, projectID uuid.UUID) ([]byte, error) {
	return s.ExportIssuesCSV(ctx, projectID)
}

// ExportProjectZIP создаёт полный бэкап проекта: задачи JSON + metadata.
func (s *ImExService) ExportProjectZIP(ctx context.Context, projectID uuid.UUID) ([]byte, error) {
	issues, err := s.repo.ListProjectIssuesPlain(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач: %w", err)
	}

	issuesJSON, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ошибка сериализации задач: %w", err)
	}

	metadata := map[string]interface{}{
		"project_id":  projectID.String(),
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"issue_count": len(issues),
	}
	metaJSON, _ := json.MarshalIndent(metadata, "", "  ")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	issuesFile, _ := zw.Create("issues.json")
	_, _ = issuesFile.Write(issuesJSON)

	metaFile, _ := zw.Create("metadata.json")
	_, _ = metaFile.Write(metaJSON)

	_ = zw.Close()

	return buf.Bytes(), nil
}

// --- Вспомогательные функции ---

// buildFieldMap создаёт маппинг индексов колонок CSV к целевым полям.
func buildFieldMap(mappings []FieldMapping, headers []string) map[string]int {
	headerIdx := make(map[string]int)
	for i, h := range headers {
		headerIdx[h] = i
	}

	fieldMap := make(map[string]int)
	for _, m := range mappings {
		if idx, ok := headerIdx[m.SourceColumn]; ok {
			fieldMap[m.TargetField] = idx
		}
	}
	return fieldMap
}

// getField извлекает значение поля из строки CSV по маппингу.
func getField(record []string, fieldMap map[string]int, field string) string {
	if idx, ok := fieldMap[field]; ok && idx < len(record) {
		return record[idx]
	}
	return ""
}
