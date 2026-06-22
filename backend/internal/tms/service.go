// Package tms реализует систему управления тестированием: кейсы, прогоны, Allure-интеграция, отчёты.
package tms

import (
	"github.com/google/uuid"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// TMSService реализует бизнес-логику модуля управления тестированием.
// Оркестрирует вызовы репозитория, версионирование, конечный автомат прогонов,
// импорт Allure и публикацию NATS-событий.
type TMSService struct {
	repo    Repository
	nats    *natsclient.Client
	matcher *Matcher
}

// NewTMSService создаёт новый сервис TMS.
func NewTMSService(repo Repository, nats *natsclient.Client) *TMSService {
	return &TMSService{
		repo:    repo,
		nats:    nats,
		matcher: NewMatcher(repo),
	}
}

// ---------------------------------------------------------------------------
// Конечный автомат прогонов
// ---------------------------------------------------------------------------

// isValidRunTransition проверяет допустимость перехода статуса прогона.
func isValidRunTransition(from, to string) bool {
	validTransitions := map[string][]string{
		"created":     {"in_progress"},
		"in_progress": {"completed", "aborted"},
	}
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Общие вспомогательные функции
// ---------------------------------------------------------------------------

// parseUUIDs конвертирует срез строк в срез UUID.
func parseUUIDs(strs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
