package devdemo

import (
	"github.com/google/uuid"

	devmod "github.com/valerii/hub/backend/internal/dev"
)

const canonicalNewbieEmail = "newbie@hub.local"

// scopeMailCalendarDemoSeedInput keeps one scope shared by organization, Calendar,
// and future Mail seeding so the newcomer remains unbound and unprovisioned.
func scopeMailCalendarDemoSeedInput(input devmod.MailCalendarDemoSeedInput) devmod.MailCalendarDemoSeedInput {
	scoped := devmod.MailCalendarDemoSeedInput{
		ProjectID: input.ProjectID,
		UserIDs:   make(map[string]uuid.UUID, len(input.UserIDs)),
	}

	for email, userID := range input.UserIDs {
		if email == canonicalNewbieEmail {
			continue
		}
		scoped.UserIDs[email] = userID
	}

	return scoped
}
