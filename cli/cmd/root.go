// Package cmd содержит все команды hub-cli.
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	apiURL string
	apiKey string
	jsonOutput bool
)

// rootCmd — корневая команда hub-cli.
var rootCmd = &cobra.Command{
	Use:   "hub-cli",
	Short: "Hub CLI — утилита для интеграции Hub с CI/CD",
	Long: `hub-cli позволяет из CI/CD пайплайнов:
  - push-results — загрузить результаты тестов (Allure)
  - create-bug   — создать баг-репорт
  - update-task   — обновить статус задачи`,
}

// Execute запускает корневую команду.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", getEnv("HUB_API_URL", "http://localhost:3000"), "URL сервера Hub API")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", os.Getenv("HUB_API_KEY"), "API-ключ для авторизации")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Вывод в JSON формате")
}

// getEnv возвращает значение переменной окружения или значение по умолчанию.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// apiRequest выполняет HTTP-запрос к Hub API с авторизацией через X-API-Key.
func apiRequest(method, path string, body interface{}) ([]byte, int, error) {
	if apiKey == "" {
		return nil, 0, fmt.Errorf("API-ключ не задан (--api-key или HUB_API_KEY)")
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("ошибка сериализации: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, apiURL+"/api"+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "hub-cli/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// printResult выводит результат в зависимости от формата (text/JSON).
func printResult(message string, data interface{}) {
	if jsonOutput {
		out, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println(message)
	}
}

// printError выводит ошибку.
func printError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}
