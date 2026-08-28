package dev

import (
	"context"

	"github.com/google/uuid"
)

// MailCalendarDemoSeedInput identifies the baseline dev data available to the demo seed.
type MailCalendarDemoSeedInput struct {
	ProjectID uuid.UUID
	UserIDs   map[string]uuid.UUID
}

// MailCalendarDemoSeeder extends the baseline dev seed with Mail and Calendar demo data.
type MailCalendarDemoSeeder interface {
	SeedMailCalendar(context.Context, MailCalendarDemoSeedInput) (map[string]any, error)
}
