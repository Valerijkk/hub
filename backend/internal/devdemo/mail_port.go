package devdemo

import (
	"context"

	"github.com/google/uuid"

	devmod "github.com/valerii/hub/backend/internal/dev"
)

// MailInvitationEventKey identifies the one Calendar invitation handed to a
// configured Mail demo adapter after the Calendar seed has reconciled it.
const MailInvitationEventKey = "mail-invitation"

// MailDemoCalendarEvent identifies exactly one seeded Calendar event without
// exposing Calendar repository, provider, or server details to the Mail port.
type MailDemoCalendarEvent struct {
	Key     string
	EventID uuid.UUID
}

// MailDemoInput is a caller-owned, safe demo scope and one keyed Calendar
// event. Implementations must not retain Scope.UserIDs after the call.
type MailDemoInput struct {
	Scope         devmod.MailCalendarDemoSeedInput
	CalendarEvent MailDemoCalendarEvent
}

// MailDemoResult reports only aggregate reconciliation counts and no raw Mail
// identifiers, provider state, message content, or credentials.
type MailDemoResult struct {
	DatasetVersion       int
	AccountsTotal        int
	FoldersTotal         int
	ThreadsTotal         int
	MessagesTotal        int
	AttachmentsTotal     int
	InvitationsTotal    int
	InvitationsReplayed int
}

// MailDemoPort keeps Mail demo reconciliation provider-neutral. The server
// adapter owns every concrete engine, Ent, and worker dependency.
type MailDemoPort interface {
	ReconcileMailDemo(context.Context, MailDemoInput) (MailDemoResult, error)
}
