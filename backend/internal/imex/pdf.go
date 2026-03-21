package imex

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/valerii/hub/backend/ent"
)

// ExportRunPDF генерирует PDF-отчёт по тестовому прогону.
// Содержит заголовок, метаданные прогона и таблицу результатов.
func (s *ImExService) ExportRunPDF(ctx context.Context, projectID, runID uuid.UUID) ([]byte, error) {
	run, err := s.db.TestRun.Get(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("прогон не найден")
		}
		return nil, fmt.Errorf("ошибка получения прогона: %w", err)
	}

	results, err := s.db.TestRun.QueryResults(run).
		WithTestCase().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения результатов: %w", err)
	}

	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	m := maroto.New(cfg)

	// Заголовок отчёта.
	m.AddRows(buildHeaderRow(fmt.Sprintf("Отчёт по прогону: %s", run.Title)))
	m.AddRows(buildInfoRow("Статус", string(run.Status)))
	m.AddRows(buildInfoRow("Дата создания", run.CreatedAt.Format("02.01.2006 15:04")))
	m.AddRows(buildInfoRow("Всего результатов", fmt.Sprintf("%d", len(results))))

	// Разделитель.
	m.AddRows(row.New(5))

	// Таблица результатов.
	m.AddRows(buildTableHeader([]string{"Тест-кейс", "Статус", "Комментарий"}))
	for _, r := range results {
		caseName := ""
		if r.Edges.TestCase != nil {
			caseName = r.Edges.TestCase.Title
		}
		m.AddRows(buildTableRow([]string{caseName, string(r.Status), r.Comment}))
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации PDF: %w", err)
	}

	return doc.GetBytes(), nil
}

// ExportSprintPDF генерирует PDF-отчёт по спринту.
// Содержит заголовок, метаданные спринта и таблицу задач.
func (s *ImExService) ExportSprintPDF(ctx context.Context, projectID, sprintID uuid.UUID) ([]byte, error) {
	sp, err := s.db.Sprint.Get(ctx, sprintID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("спринт не найден")
		}
		return nil, fmt.Errorf("ошибка получения спринта: %w", err)
	}

	issues, err := s.db.Sprint.QueryIssues(sp).
		WithStatus().
		WithAssignee().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач спринта: %w", err)
	}

	cfg := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		Build()

	m := maroto.New(cfg)

	// Заголовок.
	m.AddRows(buildHeaderRow(fmt.Sprintf("Отчёт по спринту: %s", sp.Name)))
	m.AddRows(buildInfoRow("Статус", string(sp.Status)))
	if !sp.StartDate.IsZero() {
		m.AddRows(buildInfoRow("Начало", sp.StartDate.Format("02.01.2006")))
	}
	if !sp.EndDate.IsZero() {
		m.AddRows(buildInfoRow("Окончание", sp.EndDate.Format("02.01.2006")))
	}
	m.AddRows(buildInfoRow("Задач в спринте", fmt.Sprintf("%d", len(issues))))

	// Подсчёт завершённых задач.
	completed := 0
	for _, iss := range issues {
		if iss.Edges.Status != nil && iss.Edges.Status.Category == "done" {
			completed++
		}
	}
	m.AddRows(buildInfoRow("Завершено", fmt.Sprintf("%d из %d", completed, len(issues))))

	m.AddRows(row.New(5))

	// Таблица задач.
	m.AddRows(buildTableHeader([]string{"Ключ", "Заголовок", "Статус", "Исполнитель"}))
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
		m.AddRows(buildTableRow([]string{iss.IssueKey, iss.Title, status, assignee}))
	}

	// Дата генерации.
	m.AddRows(row.New(10))
	m.AddRows(buildInfoRow("Сгенерировано", time.Now().UTC().Format("02.01.2006 15:04 UTC")))

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации PDF: %w", err)
	}

	return doc.GetBytes(), nil
}

// ---------------------------------------------------------------------------
// Вспомогательные функции для построения PDF-компонентов (maroto v2).
// ---------------------------------------------------------------------------

// buildHeaderRow создаёт строку заголовка отчёта.
func buildHeaderRow(title string) core.Row {
	return row.New(12).Add(
		col.New(12).Add(
			text.New(title, props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	)
}

// buildInfoRow создаёт строку метаданных (ключ: значение).
func buildInfoRow(label, value string) core.Row {
	return row.New(7).Add(
		col.New(4).Add(
			text.New(label+":", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(8).Add(
			text.New(value, props.Text{
				Size: 10,
			}),
		),
	)
}

// buildTableHeader создаёт строку заголовков таблицы.
func buildTableHeader(headers []string) core.Row {
	colSize := 12 / len(headers)
	r := row.New(8)
	for _, h := range headers {
		r.Add(col.New(colSize).Add(
			text.New(h, props.Text{
				Size:  9,
				Style: fontstyle.Bold,
			}),
		))
	}
	return r
}

// buildTableRow создаёт строку данных таблицы.
func buildTableRow(values []string) core.Row {
	colSize := 12 / len(values)
	r := row.New(7)
	for _, v := range values {
		cellVal := v
		if len(cellVal) > 50 {
			cellVal = cellVal[:47] + "..."
		}
		r.Add(col.New(colSize).Add(
			text.New(cellVal, props.Text{
				Size: 8,
			}),
		))
	}
	return r
}
