package dev

import (
	"context"

	"github.com/google/uuid"
)

// MailCalendarDemoSeedInput identifies the baseline dev data available to the demo seed.
type MailCalendarDemoSeedInput struct {
	ProjectID uuid.UUID
	// UserIDs is a fresh caller-owned snapshot; seeders must treat it as read-only and not retain it after the call.
	UserIDs   map[string]uuid.UUID
}

// MailCalendarDemoSeeder extends the baseline dev seed with Mail and Calendar demo data.
type MailCalendarDemoSeeder interface {
	// SeedMailCalendar returns a fresh summary map for the immediate dev HTTP response; it must not be retained as domain state.
	SeedMailCalendar(context.Context, MailCalendarDemoSeedInput) (map[string]any, error)
}
