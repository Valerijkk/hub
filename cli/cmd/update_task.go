// Команда update-task обновляет задачу в трекере Hub по ключу.
//
// Использование:
//
//	hub-cli update-task HUB-123 --status done
//	hub-cli update-task HUB-123 --status in_progress --assignee user@example.com
//	hub-cli update-task HUB-123 --priority critical
//
// Аргументы:
//   - ISSUE_KEY (обязательный) — ключ задачи (например, HUB-123)
//
// Флаги:
//   - --status   (необязательный) — новый статус задачи (done, in_progress, и т.д.)
//   - --assignee (необязательный) — email нового исполнителя
//   - --priority (необязательный) — новый приоритет (critical, high, medium, low, trivial)
//   - --comment  (необязательный) — комментарий к обновлению
//
// Отправляет PUT-запрос в /api/projects/_/issues/:key.
// Проект определяется автоматически из ключа задачи (HUB-123 -> HUB).
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// updateTaskCmd — команда обновления задачи в трекере Hub.
var updateTaskCmd = &cobra.Command{
	Use:   "update-task [ISSUE_KEY]",
	Short: "Обновить статус задачи в трекере Hub",
	Long: `Обновляет задачу по ключу (например, HUB-123) через Hub API.

Примеры:
  hub-cli update-task HUB-123 --status done
  hub-cli update-task HUB-123 --status in_progress --priority high
  hub-cli update-task HUB-123 --assignee user@example.com --comment "Готово"`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdateTask,
}

var (
	// taskStatus — новый статус задачи.
	taskStatus string
	// taskAssignee — email нового исполнителя.
	taskAssignee string
	// taskPriority — новый приоритет задачи.
	taskPriority string
	// taskComment — комментарий к обновлению.
	taskComment string
)

func init() {
	updateTaskCmd.Flags().StringVar(&taskStatus, "status", "", "Новый статус задачи")
	updateTaskCmd.Flags().StringVar(&taskAssignee, "assignee", "", "Email нового исполнителя")
	updateTaskCmd.Flags().StringVar(&taskPriority, "priority", "", "Новый приоритет")
	updateTaskCmd.Flags().StringVar(&taskComment, "comment", "", "Комментарий к обновлению")
	rootCmd.AddCommand(updateTaskCmd)
}

// runUpdateTask — основная логика команды update-task.
// Собирает только указанные флаги в тело запроса и отправляет PUT.
func runUpdateTask(_ *cobra.Command, args []string) error {
	issueKey := args[0]

	body := map[string]interface{}{
		"version": 0, // Серверная версия — обновляется автоматически
	}

	if taskStatus != "" {
		body["status_id"] = taskStatus
	}
	if taskAssignee != "" {
		body["assignee_id"] = taskAssignee
	}
	if taskPriority != "" {
		body["priority"] = taskPriority
	}
	if taskComment != "" {
		body["comment"] = taskComment
	}

	// Проект определяется из ключа задачи (HUB-123 -> HUB)
	path := fmt.Sprintf("/projects/_/issues/%s", issueKey)
	respBody, statusCode, err := apiRequest("PUT", path, body)
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

	printResult(fmt.Sprintf("Задача %s обновлена", issueKey), resp)
	return nil
}
