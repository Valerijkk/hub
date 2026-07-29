package ai

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// ErrExternalAIConsentRequired — user отключил external LLM (M3-030).
var ErrExternalAIConsentRequired = fmt.Errorf("external AI consent required — enable in Settings → AI")

// UserPreferenceRecord — сохранённые AI-настройки пользователя.
type UserPreferenceRecord struct {
	DefaultModel      string
	DefaultProviderID *uuid.UUID
	AllowExternalAI   bool
	Version           int
}

// externalAIAllowed — согласие пользователя на external LLM.
func (s *Service) externalAIAllowed(ctx context.Context, userID uuid.UUID) (bool, error) {
	pref, err := s.repo.GetUserPreferenceRecord(ctx, userID)
	if err != nil {
		return false, err
	}
	if pref == nil {
		// W5-A3-012 / P2-BE-17: opt-in, не opt-out — без записи external AI запрещён.
		return false, nil
	}
	return pref.AllowExternalAI, nil
}

func filterProvidersByConsent(providers []*ent.AIProvider, allowExternal bool) []*ent.AIProvider {
	if allowExternal {
		return providers
	}
	out := make([]*ent.AIProvider, 0, len(providers))
	for _, p := range providers {
		if isOfflineProvider(p) {
			out = append(out, p)
		}
	}
	return out
}

func consentErrorForModel(model string, allowExternal bool, providers []*ent.AIProvider) error {
	if allowExternal || len(providers) > 0 {
		return nil
	}
	if model != "" {
		return fmt.Errorf("%w (model %q requires cloud provider)", ErrExternalAIConsentRequired, model)
	}
	return ErrExternalAIConsentRequired
}
