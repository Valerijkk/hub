package devdemo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	devmod "github.com/valerii/hub/backend/internal/dev"
)

var _ devmod.MailCalendarDemoSeeder = (*MailCalendarDemoSeeder)(nil)

// MailCalendarDemoSeeder provides the dev seed integration boundary for Mail and Calendar.
type MailCalendarDemoSeeder struct {
	organization *organizationSeeder
	calendar     *calendarSeeder
	mail         MailDemoPort
}

// NewMailCalendarDemoSeeder constructs the Mail and Calendar demo seeder.
func NewMailCalendarDemoSeeder(client *ent.Client, rawDB *sql.DB, postgresAdvisoryLocks bool, calendarPort CalendarDemoPort) (*MailCalendarDemoSeeder, error) {
	return newMailCalendarDemoSeeder(client, rawDB, postgresAdvisoryLocks, calendarPort, nil)
}

// NewMailCalendarDemoSeederWithMail adds one provider-neutral Mail reconciliation
// after the Mail-enabled Calendar dataset has created its canonical invitation.
func NewMailCalendarDemoSeederWithMail(client *ent.Client, rawDB *sql.DB, postgresAdvisoryLocks bool, calendarPort CalendarDemoPort, mailPort MailDemoPort) (*MailCalendarDemoSeeder, error) {
	if isNilCalendarDemoDependency(mailPort) {
		return nil, errors.New("mail calendar demo seeder: mail port is required")
	}
	return newMailCalendarDemoSeeder(client, rawDB, postgresAdvisoryLocks, calendarPort, mailPort)
}

func newMailCalendarDemoSeeder(client *ent.Client, rawDB *sql.DB, postgresAdvisoryLocks bool, calendarPort CalendarDemoPort, mailPort MailDemoPort) (*MailCalendarDemoSeeder, error) {
	transactions, err := newTransactionRunner(client, rawDB, postgresAdvisoryLocks)
	if err != nil {
		return nil, fmt.Errorf("create mail calendar demo transaction runner: %w", err)
	}
	organization, err := newOrganizationSeeder(transactions)
	if err != nil {
		return nil, fmt.Errorf("create mail calendar demo organization seeder: %w", err)
	}
	calendar, err := newCalendarSeeder(calendarPort, mailPort != nil)
	if err != nil {
		return nil, fmt.Errorf("create mail calendar demo calendar seeder: %w", err)
	}

	return &MailCalendarDemoSeeder{organization: organization, calendar: calendar, mail: mailPort}, nil
}

// SeedMailCalendar reconciles the currently implemented shared demo prerequisites.
func (seeder *MailCalendarDemoSeeder) SeedMailCalendar(ctx context.Context, input devmod.MailCalendarDemoSeedInput) (map[string]any, error) {
	if seeder == nil || seeder.organization == nil || seeder.calendar == nil ||
		(seeder.mail != nil && isNilCalendarDemoDependency(seeder.mail)) {
		return nil, errors.New("mail calendar demo seeder is unavailable")
	}
	scopedInput := scopeMailCalendarDemoSeedInput(input)
	organization, err := seeder.organization.seed(ctx, scopedInput)
	if err != nil {
		return nil, fmt.Errorf("seed mail calendar demo: %w", err)
	}
	if seeder.mail != nil {
		if organization.organizationID == uuid.Nil {
			return nil, errors.New("seed mail calendar demo: organization result is invalid")
		}
		if err := seeder.mail.PrepareMailDemo(ctx, MailDemoPrepareInput{
			Scope:          scopeMailCalendarDemoSeedInput(scopedInput),
			OrganizationID: organization.organizationID,
		}); err != nil {
			return nil, fmt.Errorf("seed mail calendar demo: prepare mail demo: %w", err)
		}
	}
	calendar, err := seeder.calendar.seed(ctx, scopedInput.UserIDs)
	if err != nil {
		return nil, fmt.Errorf("seed mail calendar demo: %w", err)
	}

	summary := map[string]any{
		"organization": map[string]any{
			"id":               organization.organizationID.String(),
			"created":          organization.created,
			"users_attached":   organization.usersAttached,
			"project_attached": organization.projectAttached,
		},
		"calendar": map[string]int{
			"dataset_version":       calendar.datasetVersion,
			"events_total":          calendar.eventsTotal,
			"events_replayed":       calendar.eventsReplayed,
			"rsvps_total":           calendar.rsvpsTotal,
			"rsvps_replayed":        calendar.rsvpsReplayed,
			"occurrences_cancelled": calendar.occurrencesCancelled,
		},
	}
	if seeder.mail == nil {
		return summary, nil
	}

	invitationEventID, err := calendar.eventID(MailInvitationEventKey)
	if err != nil {
		return nil, fmt.Errorf("seed mail calendar demo: resolve mail invitation event: %w", err)
	}
	mail, err := seeder.mail.ReconcileMailDemo(ctx, MailDemoInput{
		Scope: scopeMailCalendarDemoSeedInput(scopedInput),
		CalendarEvent: MailDemoCalendarEvent{
			Key:     MailInvitationEventKey,
			EventID: invitationEventID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("seed mail calendar demo: reconcile mail demo: %w", err)
	}
	if mail.DatasetVersion < 0 ||
		mail.AccountsTotal < 0 ||
		mail.FoldersTotal < 0 ||
		mail.ThreadsTotal < 0 ||
		mail.MessagesTotal < 0 ||
		mail.AttachmentsTotal < 0 ||
		mail.InvitationsTotal < 0 ||
		mail.InvitationsReplayed < 0 ||
		mail.InvitationsReplayed > mail.InvitationsTotal {
		return nil, errors.New("seed mail calendar demo: mail result is invalid")
	}
	summary["mail"] = map[string]int{
		"dataset_version":       mail.DatasetVersion,
		"accounts_total":        mail.AccountsTotal,
		"folders_total":         mail.FoldersTotal,
		"threads_total":         mail.ThreadsTotal,
		"messages_total":        mail.MessagesTotal,
		"attachments_total":     mail.AttachmentsTotal,
		"invitations_total":    mail.InvitationsTotal,
		"invitations_replayed": mail.InvitationsReplayed,
	}
	return summary, nil
}
