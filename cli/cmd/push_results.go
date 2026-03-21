package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// pushResultsCmd загружает результаты тестов (Allure) в Hub TMS.
var pushResultsCmd = &cobra.Command{
	Use:   "push-results",
	Short: "Загрузить результаты тестов (Allure) в Hub",
	Long:  `Читает JSON-файлы из директории allure-results и отправляет в Hub TMS API.`,
	RunE:  runPushResults,
}

var (
	pushProject string
	pushRunID   string
	pushFile    string
)

func init() {
	pushResultsCmd.Flags().StringVar(&pushProject, "project", "", "Ключ проекта (например, HUB)")
	pushResultsCmd.Flags().StringVar(&pushRunID, "run-id", "", "ID прогона (для добавления к существующему)")
	pushResultsCmd.Flags().StringVar(&pushFile, "file", "", "Путь к директории allure-results")
	_ = pushResultsCmd.MarkFlagRequired("project")
	_ = pushResultsCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(pushResultsCmd)
}

func runPushResults(_ *cobra.Command, _ []string) error {
	// Собираем JSON-файлы результатов
	results, err := readAllureResults(pushFile)
	if err != nil {
		printError(fmt.Sprintf("ошибка чтения результатов: %s", err))
		return err
	}

	if len(results) == 0 {
		printError("не найдено JSON-файлов результатов")
		return fmt.Errorf("пустая директория результатов")
	}

	// Определяем endpoint
	var path string
	if pushRunID != "" {
		path = fmt.Sprintf("/projects/%s/tms/runs/%s/allure", pushProject, pushRunID)
	} else {
		path = fmt.Sprintf("/projects/%s/tms/allure/import", pushProject)
	}

	body := map[string]interface{}{
		"results": results,
	}

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

	printResult(fmt.Sprintf("Загружено %d результатов в проект %s", len(results), pushProject), resp)
	return nil
}

// readAllureResults читает JSON-файлы из директории allure-results.
func readAllureResults(dir string) ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать директорию: %w", err)
	}

	var results []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}

		results = append(results, result)
	}

	return results, nil
}
