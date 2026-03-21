package tms

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

// ---------------------------------------------------------------------------
// Парсер Allure JSON
// ---------------------------------------------------------------------------
// Защитный парсер для allure-results JSON. Каждое поле опционально,
// парсер никогда не паникует на неполных данных.
// ---------------------------------------------------------------------------

// ParseAllureResults парсит JSON-массив результатов Allure.
// Возвращает ошибку при невалидном JSON, но никогда не паникует.
func ParseAllureResults(data []byte) ([]AllureResult, error) {
	var results []AllureResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("ошибка парсинга allure-results JSON: %w", err)
	}
	if results == nil {
		results = make([]AllureResult, 0)
	}
	for i, r := range results {
		if r.Status == "" {
			slog.Warn("allure результат без статуса", "index", i, "uuid", r.UUID, "name", r.Name)
		}
	}
	return results, nil
}

// ParseAllureResult парсит одиночный результат Allure из JSON.
func ParseAllureResult(data []byte) (*AllureResult, error) {
	var result AllureResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("ошибка парсинга allure-result JSON: %w", err)
	}
	return &result, nil
}

// extractLabelValue безопасно извлекает значение метки по имени.
// Возвращает пустую строку, если метка не найдена или labels == nil.
func extractLabelValue(labels []AllureLabel, name string) string {
	for _, l := range labels {
		if l.Name == name {
			return l.Value
		}
	}
	return ""
}

// allureStatusToTMS преобразует статус Allure в статус TMS.
// Допустимые статусы: passed, failed, broken, skipped.
// Всё остальное → unknown.
func allureStatusToTMS(status string) string {
	switch status {
	case "passed", "failed", "broken", "skipped":
		return status
	default:
		return "unknown"
	}
}
