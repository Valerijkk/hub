package devdemo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/valerii/hub/backend/ent"
	devmod "github.com/valerii/hub/backend/internal/dev"
)

var _ devmod.MailCalendarDemoSeeder = (*MailCalendarDemoSeeder)(nil)

// MailCalendarDemoSeeder provides the dev seed integration boundary for Mail and Calendar.
type MailCalendarDemoSeeder struct {
	organization *organizationSeeder
}

// NewMailCalendarDemoSeeder constructs the Mail and Calendar demo seeder.
func NewMailCalendarDemoSeeder(client *ent.Client, rawDB *sql.DB, postgresAdvisoryLocks bool) (*MailCalendarDemoSeeder, error) {
	transactions, err := newTransactionRunner(client, rawDB, postgresAdvisoryLocks)
	if err != nil {
		return nil, fmt.Errorf("create mail calendar demo transaction runner: %w", err)
	}
	organization, err := newOrganizationSeeder(transactions)
	if err != nil {
		return nil, fmt.Errorf("create mail calendar demo organization seeder: %w", err)
	}
	return &MailCalendarDemoSeeder{organization: organization}, nil
}

// SeedMailCalendar reconciles the currently implemented shared demo prerequisites.
func (seeder *MailCalendarDemoSeeder) SeedMailCalendar(ctx context.Context, input devmod.MailCalendarDemoSeedInput) (map[string]any, error) {
	if seeder == nil || seeder.organization == nil {
		return nil, errors.New("mail calendar demo seeder is unavailable")
	}
	organization, err := seeder.organization.seed(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("seed mail calendar demo: %w", err)
	}
	return map[string]any{
		"organization": map[string]any{
			"id":               organization.organizationID.String(),
			"created":          organization.created,
			"users_attached":   organization.usersAttached,
			"project_attached": organization.projectAttached,
		},
	}, nil
}
