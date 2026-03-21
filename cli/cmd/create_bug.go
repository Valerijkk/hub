// Команда create-bug создаёт баг-репорт в трекере Hub.
//
// Использование:
//
//	hub-cli create-bug --project HUB --title "Кнопка не работает" --priority high
//	hub-cli create-bug --project HUB --title "Баг" --description "Подробное описание"
//
// Флаги:
//   - --project     (обязательный) — ключ проекта (например, HUB)
//   - --title       (обязательный) — заголовок бага
//   - --priority    (необязательный) — приоритет: critical, high, medium, low, trivial (по умолчанию medium)
//   - --description (необязательный) — описание бага
//   - --assignee    (необязательный) — email исполнителя
//   - --labels      (необязательный) — метки через запятую
//
// Создаёт задачу типа 'bug' через POST /api/projects/:key/issues.
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// createBugCmd — команда создания баг-репорта в трекере Hub.
var createBugCmd = &cobra.Command{
	Use:   "create-bug",
	Short: "Создать баг-репорт в трекере Hub",
	Long: `Создаёт задачу типа 'bug' в указанном проекте через Hub API.

Примеры:
  hub-cli create-bug --project HUB --title "Кнопка не работает" --priority high
  hub-cli create-bug --project HUB --title "Баг" --description "Описание" --assignee user@example.com`,
	RunE: runCreateBug,
}

var (
	// bugProject — ключ проекта для создания бага.
	bugProject string
	// bugTitle — заголовок бага.
	bugTitle string
	// bugPriority — приоритет бага (critical, high, medium, low, trivial).
	bugPriority string
	// bugDesc — описание бага.
	bugDesc string
	// bugAssignee — email исполнителя.
	bugAssignee string
	// bugLabels — метки через запятую.
	bugLabels string
)

func init() {
	createBugCmd.Flags().StringVar(&bugProject, "project", "", "Ключ проекта (например, HUB)")
	createBugCmd.Flags().StringVar(&bugTitle, "title", "", "Заголовок бага")
	createBugCmd.Flags().StringVar(&bugPriority, "priority", "medium", "Приоритет: critical, high, medium, low, trivial")
	createBugCmd.Flags().StringVar(&bugDesc, "description", "", "Описание бага")
	createBugCmd.Flags().StringVar(&bugAssignee, "assignee", "", "Email исполнителя")
	createBugCmd.Flags().StringVar(&bugLabels, "labels", "", "Метки через запятую")
	_ = createBugCmd.MarkFlagRequired("project")
	_ = createBugCmd.MarkFlagRequired("title")
	rootCmd.AddCommand(createBugCmd)
}

// runCreateBug — основная логика команды create-bug.
// Формирует тело запроса с типом 'bug' и отправляет POST в Hub API.
func runCreateBug(_ *cobra.Command, _ []string) error {
	body := map[string]interface{}{
		"title":      bugTitle,
		"issue_type": "bug",
		"priority":   bugPriority,
	}

	// Добавляем необязательные поля, если указаны
	if bugDesc != "" {
		body["description"] = bugDesc
	}
	if bugAssignee != "" {
		body["assignee_email"] = bugAssignee
	}
	if bugLabels != "" {
		labels := strings.Split(bugLabels, ",")
		trimmed := make([]string, 0, len(labels))
		for _, l := range labels {
			l = strings.TrimSpace(l)
			if l != "" {
				trimmed = append(trimmed, l)
			}
		}
		if len(trimmed) > 0 {
			body["labels"] = trimmed
		}
	}

	path := fmt.Sprintf("/projects/%s/issues", bugProject)
	respBody, statusCode, err := apiRequest("POST", path, body)
	if err != nil {
		printError(err.Error())
		return err
	}

	if statusCode >= 400 {
		printError(fmt.Sprintf("сервер вернул %d: %s", statusCode, string(respBody)))
		return fmt.Errorf("HTTP %d", statusCode)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(respBody, &resp)

	key, _ := resp["issue_key"].(string)
	if key == "" {
		key = "created"
	}

	printResult(fmt.Sprintf("Bug создан: %s — %s", key, bugTitle), resp)
	return nil
}
