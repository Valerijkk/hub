// Package tms реализует систему управления тестированием: кейсы, прогоны, Allure-интеграция, отчёты.
package tms

import (
	"github.com/google/uuid"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// TMSService implements the business logic for the test management module.
// Orchestrates repository calls, versioning, run state machine,
// Allure result imports and NATS event publishing.
type TMSService struct {
	repo    Repository
	nats    *natsclient.Client
	matcher *Matcher
}

// NewTMSService creates a new TMS service.
func NewTMSService(repo Repository, nats *natsclient.Client) *TMSService {
	return &TMSService{
		repo:    repo,
		nats:    nats,
		matcher: NewMatcher(repo),
	}
}

// ---------------------------------------------------------------------------
// Run state machine
// ---------------------------------------------------------------------------

// isValidRunTransition checks if a run status transition is allowed.
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
// Shared helpers
// ---------------------------------------------------------------------------

// parseUUIDs converts a string slice to a UUID slice.
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
