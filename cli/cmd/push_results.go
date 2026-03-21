// Команда push-results загружает результаты тестов (Allure JSON) в Hub TMS.
//
// Использование:
//
//	hub-cli push-results --project HUB --file ./allure-results/
//	hub-cli push-results --project HUB --run-id "Build #42" --file ./allure-results/
//
// Флаги:
//   - --project  (обязательный) — ключ проекта (например, HUB)
//   - --run-id   (необязательный) — ID прогона для добавления к существующему
//   - --file     (обязательный) — путь к директории с JSON-файлами результатов
//
// Команда рекурсивно читает все *.json файлы из указанной директории,
// парсит их и отправляет POST-запросом в /api/projects/:key/tms/allure/import.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// pushResultsCmd — команда загрузки результатов тестов (Allure) в Hub TMS.
var pushResultsCmd = &cobra.Command{
	Use:   "push-results",
	Short: "Загрузить результаты тестов (Allure) в Hub",
	Long: `Читает JSON-файлы из директории allure-results и отправляет в Hub TMS API.

Примеры:
  hub-cli push-results --project HUB --file ./allure-results/
  hub-cli push-results --project HUB --run-id run-42 --file ./allure-results/`,
	RunE: runPushResults,
}

var (
	// pushProject — ключ проекта для загрузки результатов.
	pushProject string
	// pushRunID — ID прогона (если добавляем к существующему).
	pushRunID string
	// pushFile — путь к директории с allure-результатами.
	pushFile string
)

func init() {
	pushResultsCmd.Flags().StringVar(&pushProject, "project", "", "Ключ проекта (например, HUB)")
	pushResultsCmd.Flags().StringVar(&pushRunID, "run-id", "", "ID прогона (для добавления к существующему)")
	pushResultsCmd.Flags().StringVar(&pushRunID, "run", "", "Алиас для --run-id")
	pushResultsCmd.Flags().StringVar(&pushFile, "file", "", "Путь к директории allure-results")
	_ = pushResultsCmd.MarkFlagRequired("project")
	_ = pushResultsCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(pushResultsCmd)
}

// runPushResults — основная логика команды push-results.
// Собирает JSON-файлы из директории, формирует тело запроса и отправляет в Hub API.
func runPushResults(_ *cobra.Command, args []string) error {
	// Поддержка позиционного аргумента (путь к директории) как альтернатива --file
	dir := pushFile
	if dir == "" && len(args) > 0 {
		dir = args[0]
	}

	// Собираем JSON-файлы результатов
	results, err := readAllureResults(dir)
	if err != nil {
		printError(fmt.Sprintf("ошибка чтения результатов: %s", err))
		return err
	}

	if len(results) == 0 {
		printError("не найдено JSON-файлов результатов в директории " + dir)
		return fmt.Errorf("пустая директория результатов")
	}

	// Определяем endpoint в зависимости от наличия run-id
	var path string
	if pushRunID != "" {
		// Заменяем пробелы и спецсимволы для безопасности URL
		safeRunID := strings.ReplaceAll(pushRunID, " ", "-")
		path = fmt.Sprintf("/projects/%s/tms/runs/%s/allure", pushProject, safeRunID)
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

// readAllureResults читает все JSON-файлы из указанной директории.
// Файлы с расширением .json парсятся как map[string]interface{}.
// Файлы с ошибками парсинга пропускаются без прерывания работы.
func readAllureResults(dir string) ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать директорию %s: %w", dir, err)
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
