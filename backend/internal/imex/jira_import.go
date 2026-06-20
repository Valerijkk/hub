package imex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
)

// ---------------------------------------------------------------------------
// Типы для импорта Jira JSON
// ---------------------------------------------------------------------------

// JiraIssue представляет задачу Jira из JSON-экспорта.
type JiraIssue struct {
	Key    string     `json:"key"`    // например, "PROJ-123"
	Fields JiraFields `json:"fields"`
}

// JiraFields содержит все поля задачи Jira.
type JiraFields struct {
	Summary     string          `json:"summary"`
	Description interface{}     `json:"description"` // ADF or string
	IssueType   JiraNamedField  `json:"issuetype"`
	Priority    JiraNamedField  `json:"priority"`
	Status      JiraNamedField  `json:"status"`
	Assignee    *JiraUser       `json:"assignee"`
	Reporter    *JiraUser       `json:"reporter"`
	Created     string          `json:"created"`
	Updated     string          `json:"updated"`
	Labels      []string        `json:"labels"`
	Parent      *JiraIssue      `json:"parent"`
	Subtasks    []JiraIssue     `json:"subtasks"`
	Comment     *JiraComments   `json:"comment"`
}

// JiraNamedField — общий объект Jira со свойством name.
type JiraNamedField struct {
	Name string `json:"name"`
}

// JiraUser представляет ссылку на пользователя Jira.
type JiraUser struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// JiraComments оборачивает список комментариев Jira.
type JiraComments struct {
	Comments []JiraComment `json:"comments"`
}

// JiraComment представляет один комментарий Jira.
type JiraComment struct {
	Body    interface{} `json:"body"` // ADF or string
	Author  JiraUser    `json:"author"`
	Created string      `json:"created"`
}

// JiraSearchResponse — ответ верхнего уровня от Jira search API.
type JiraSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
}

// ---------------------------------------------------------------------------
// Вспомогательные функции маппинга
// ---------------------------------------------------------------------------

// mapJiraIssueType преобразует название типа задачи Jira в тип Hub IssueType.
func mapJiraIssueType(jiraType string) string {
	switch strings.ToLower(jiraType) {
	case "bug":
		return string(issue.IssueTypeBug)
	case "story":
		return string(issue.IssueTypeStory)
	case "task":
		return string(issue.IssueTypeTask)
	case "epic":
		return string(issue.IssueTypeEpic)
	case "sub-task", "subtask", "sub_task":
		return string(issue.IssueTypeSubTask)
	case "improvement":
		return string(issue.IssueTypeImprovement)
	case "spike":
		return string(issue.IssueTypeSpike)
	case "tech debt", "technical debt":
		return string(issue.IssueTypeTechDebt)
	default:
		return string(issue.IssueTypeTask)
	}
}

// mapJiraPriority преобразует название приоритета Jira в приоритет Hub.
func mapJiraPriority(jiraPriority string) string {
	switch strings.ToLower(jiraPriority) {
	case "highest", "blocker":
		return string(issue.PriorityCritical)
	case "high":
		return string(issue.PriorityHigh)
	case "medium", "normal":
		return string(issue.PriorityMedium)
	case "low":
		return string(issue.PriorityLow)
	case "lowest", "trivial":
		return string(issue.PriorityTrivial)
	default:
		return string(issue.PriorityMedium)
	}
}

// jiraDescriptionToText извлекает обычный текст из поля описания Jira.
// Описание может быть строкой или объектом Atlassian Document Format (ADF).
func jiraDescriptionToText(desc interface{}) string {
	if desc == nil {
		return ""
	}
	switch v := desc.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Simplified ADF extraction: walk content nodes and collect text.
		return extractADFText(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// extractADFText рекурсивно извлекает текст из узла Atlassian Document Format.
func extractADFText(node map[string]interface{}) string {
	var sb strings.Builder

	if text, ok := node["text"].(string); ok {
		sb.WriteString(text)
	}

	if content, ok := node["content"].([]interface{}); ok {
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				sb.WriteString(extractADFText(childMap))
			}
		}
		// Add newline after paragraph-level nodes.
		if nodeType, _ := node["type"].(string); nodeType == "paragraph" || nodeType == "heading" {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ---------------------------------------------------------------------------
// Логика импорта
// ---------------------------------------------------------------------------

// ImportJiraJSON импортирует задачи из JSON-экспорта Jira.
// Принимает массив JiraIssue или ответ Jira search {"issues": [...]}.
func (s *ImExService) ImportJiraJSON(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, data []byte) (*ImportResultResponse, error) {
	issues, err := parseJiraJSON(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing Jira JSON: %w", err)
	}

	if len(issues) == 0 {
		return &ImportResultResponse{Imported: 0, Skipped: 0}, nil
	}

	imported := 0
	skipped := 0
	var errs []string

	// Первый проход: создаём все задачи и строим карту key->ID для связывания родителей.
	keyToID := make(map[string]uuid.UUID)

	for i, ji := range issues {
		title := ji.Fields.Summary
		if title == "" {
			skipped++
			errs = append(errs, fmt.Sprintf("issue %d (%s): missing summary", i, ji.Key))
			continue
		}

		issueType := mapJiraIssueType(ji.Fields.IssueType.Name)
		priority := mapJiraPriority(ji.Fields.Priority.Name)
		descText := jiraDescriptionToText(ji.Fields.Description)

		proj, err := s.db.Project.Get(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("project not found: %w", err)
		}
		nextNum := proj.TaskCounter + 1

		_, err = s.db.Project.UpdateOneID(projectID).
			SetTaskCounter(nextNum).
			Save(ctx)
		if err != nil {
			skipped++
			continue
		}

		issueKey := fmt.Sprintf("%s-%d", proj.Key, nextNum)

		statuses, err := s.db.Project.Query().
			Where(project.ID(projectID)).
			QueryWorkflowStatuses().
			First(ctx)
		if err != nil {
			skipped++
			slog.Error("failed to get default status for Jira import", "error", err)
			continue
		}

		creator := s.db.Issue.Create().
			SetTitle(title).
			SetIssueType(issue.IssueType(issueType)).
			SetPriority(issue.Priority(priority)).
			SetIssueNumber(nextNum).
			SetIssueKey(issueKey).
			SetProjectID(projectID).
			SetStatusID(statuses.ID).
			SetReporterID(userID)

		if descText != "" {
			creator = creator.SetDescriptionText(descText)
		}

		created, err := creator.Save(ctx)
		if err != nil {
			skipped++
			errs = append(errs, fmt.Sprintf("issue %d (%s): %s", i, ji.Key, err.Error()))
			continue
		}

		keyToID[ji.Key] = created.ID

		// Import comments for this issue.
		if ji.Fields.Comment != nil {
			for _, jc := range ji.Fields.Comment.Comments {
				bodyText := jiraDescriptionToText(jc.Body)
				if bodyText == "" {
					continue
				}
				_, cerr := s.db.Comment.Create().
					SetBodyText(bodyText).
					SetIssueID(created.ID).
					SetAuthorID(userID).
					Save(ctx)
				if cerr != nil {
					slog.Error("failed to create comment during Jira import",
						"issue_key", ji.Key, "error", cerr)
				}
			}
		}

		imported++
	}

	// Второй проход: связываем родительские/дочерние задачи.
	for _, ji := range issues {
		myID, ok := keyToID[ji.Key]
		if !ok {
			continue
		}

		// Link via parent field.
		if ji.Fields.Parent != nil {
			if parentID, ok := keyToID[ji.Fields.Parent.Key]; ok {
				_, err := s.db.Issue.UpdateOneID(myID).
					SetParentID(parentID).
					Save(ctx)
				if err != nil {
					slog.Error("failed to set parent during Jira import",
						"child", ji.Key, "parent", ji.Fields.Parent.Key, "error", err)
				}
			}
		}

		// Link subtasks (reverse direction: subtask.parent = current issue).
		for _, sub := range ji.Fields.Subtasks {
			if subID, ok := keyToID[sub.Key]; ok {
				_, err := s.db.Issue.UpdateOneID(subID).
					SetParentID(myID).
					Save(ctx)
				if err != nil {
					slog.Error("failed to set subtask parent during Jira import",
						"subtask", sub.Key, "parent", ji.Key, "error", err)
				}
			}
		}
	}

	return &ImportResultResponse{
		Imported: imported,
		Skipped:  skipped,
		Errors:   errs,
	}, nil
}

// parseJiraJSON парсит JSON-данные Jira, принимая массив задач
// или объект ответа поиска {"issues": [...]}.
func parseJiraJSON(data []byte) ([]JiraIssue, error) {
	// Try array first.
	var issues []JiraIssue
	if err := json.Unmarshal(data, &issues); err == nil && len(issues) > 0 {
		return issues, nil
	}

	// Try search response format.
	var searchResp JiraSearchResponse
	if err := json.Unmarshal(data, &searchResp); err == nil && len(searchResp.Issues) > 0 {
		return searchResp.Issues, nil
	}

	// Try single issue.
	var single JiraIssue
	if err := json.Unmarshal(data, &single); err == nil && single.Key != "" {
		return []JiraIssue{single}, nil
	}

	return nil, fmt.Errorf("unrecognized Jira JSON format")
}

// ---------------------------------------------------------------------------
// HTTP-обработчик
// ---------------------------------------------------------------------------

// ImportJira обрабатывает POST /api/projects/:projectId/import/jira (multipart).
// ImportJira godoc
// @Summary Импорт из Jira
// @Description Импортирует задачи из JSON-экспорта Jira с маппингом статусов и приоритетов.
// @Tags Импорт/Экспорт
// @Accept multipart/form-data
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param file formData file true "JSON-файл экспорта Jira"
// @Success 200 {object} JiraImportResponse "Результат импорта"
// @Failure 400 {object} fiber.Error "Ошибка файла или данных"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/import/jira [post]
func (h *Handler) ImportJira(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user ID")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	f, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "error opening file")
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxImportFileBytes+1))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "error reading file")
	}
	if len(data) > maxImportFileBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "file too large")
	}

	result, err := h.svc.ImportJiraJSON(c.Context(), projectID, userID, data)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(result)
}
