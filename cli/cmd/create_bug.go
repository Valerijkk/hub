package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// createBugCmd создаёт баг-репорт в трекере Hub.
var createBugCmd = &cobra.Command{
	Use:   "create-bug",
	Short: "Создать баг-репорт в трекере Hub",
	Long:  `Создаёт задачу типа 'bug' в указанном проекте через Hub API.`,
	RunE:  runCreateBug,
}

var (
	bugProject  string
	bugTitle    string
	bugPriority string
	bugDesc     string
)

func init() {
	createBugCmd.Flags().StringVar(&bugProject, "project", "", "Ключ проекта (например, HUB)")
	createBugCmd.Flags().StringVar(&bugTitle, "title", "", "Заголовок бага")
	createBugCmd.Flags().StringVar(&bugPriority, "priority", "medium", "Приоритет: critical, high, medium, low, trivial")
	createBugCmd.Flags().StringVar(&bugDesc, "description", "", "Описание бага")
	_ = createBugCmd.MarkFlagRequired("project")
	_ = createBugCmd.MarkFlagRequired("title")
	rootCmd.AddCommand(createBugCmd)
}

func runCreateBug(_ *cobra.Command, _ []string) error {
	body := map[string]interface{}{
		"title":      bugTitle,
		"issue_type": "bug",
		"priority":   bugPriority,
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
