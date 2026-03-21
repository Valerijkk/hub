// Команда auth сохраняет API-ключ в ~/.hub/config.json
// для последующего использования без флага --api-key.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// hubConfig — структура конфигурационного файла ~/.hub/config.json.
type hubConfig struct {
	APIKey string `json:"api_key"`
	APIURL string `json:"api_url,omitempty"`
}

// authCmd сохраняет API-ключ в конфигурационный файл.
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Сохранить API-ключ для авторизации",
	Long:  `Сохраняет API-ключ в ~/.hub/config.json для использования без флага --api-key.`,
	RunE:  runAuth,
}

var authToken string

func init() {
	authCmd.Flags().StringVar(&authToken, "token", "", "API-ключ для сохранения")
	_ = authCmd.MarkFlagRequired("token")
	rootCmd.AddCommand(authCmd)
}

func runAuth(_ *cobra.Command, _ []string) error {
	cfgPath, err := configPath()
	if err != nil {
		printError(fmt.Sprintf("не удалось определить путь конфига: %s", err))
		return err
	}

	// Читаем существующий конфиг или создаём новый.
	cfg := hubConfig{}
	if data, readErr := os.ReadFile(cfgPath); readErr == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	cfg.APIKey = authToken
	if apiURL != "" && apiURL != "http://localhost:3000" {
		cfg.APIURL = apiURL
	}

	// Создаём директорию ~/.hub если не существует.
	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		printError(fmt.Sprintf("не удалось создать директорию %s: %s", dir, err))
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации конфига: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		printError(fmt.Sprintf("не удалось записать конфиг: %s", err))
		return err
	}

	printResult(fmt.Sprintf("API-ключ сохранён в %s", cfgPath), map[string]string{"config_path": cfgPath})
	return nil
}

// configPath возвращает путь к конфигурационному файлу ~/.hub/config.json.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hub", "config.json"), nil
}

// loadConfig загружает конфигурацию из ~/.hub/config.json.
// Возвращает пустую конфигурацию если файл не найден.
func loadConfig() hubConfig {
	cfg := hubConfig{}
	p, err := configPath()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}
