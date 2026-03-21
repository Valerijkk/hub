package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// updateTaskCmd обновляет задачу в трекере Hub.
var updateTaskCmd = &cobra.Command{
	Use:   "update-task [ISSUE_KEY]",
	Short: "Обновить статус задачи в трекере Hub",
	Long:  `Обновляет задачу по ключу (например, HUB-123) через Hub API.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUpdateTask,
}

var (
	taskStatus   string
	taskAssignee string
	taskPriority string
)

func init() {
	updateTaskCmd.Flags().StringVar(&taskStatus, "status", "", "Новый статус задачи")
	updateTaskCmd.Flags().StringVar(&taskAssignee, "assignee", "", "Email нового исполнителя")
	updateTaskCmd.Flags().StringVar(&taskPriority, "priority", "", "Новый приоритет")
	rootCmd.AddCommand(updateTaskCmd)
}

func runUpdateTask(_ *cobra.Command, args []string) error {
	issueKey := args[0]

	body := map[string]interface{}{
		"version": 0, // Будет обновлён сервером
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

	// Используем project key из issue key (HUB-123 -> HUB)
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
