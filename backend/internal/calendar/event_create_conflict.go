package calendar

// CalendarEventCreateConflictKind classifies safe create-time conflicts.
type CalendarEventCreateConflictKind string

const (
	CalendarEventCreateConflictIdempotencyKeyReused       CalendarEventCreateConflictKind = "idempotency_key_reused"
	CalendarEventCreateConflictScheduledCallLinkMismatch  CalendarEventCreateConflictKind = "scheduled_call_link_mismatch"
	CalendarEventCreateConflictScheduledCallNotAttachable CalendarEventCreateConflictKind = "scheduled_call_not_attachable"
)

// CalendarEventCreateConflictError carries only a safe machine-readable kind.
type CalendarEventCreateConflictError struct {
	kind CalendarEventCreateConflictKind
}

func (err *CalendarEventCreateConflictError) Error() string {
	return "calendar event create conflict"
}

func (err *CalendarEventCreateConflictError) Unwrap() error {
	return ErrEventConflict
}

func (err *CalendarEventCreateConflictError) Kind() CalendarEventCreateConflictKind {
	if err == nil {
		return ""
	}
	return err.kind
}

func newCalendarEventCreateConflict(kind CalendarEventCreateConflictKind) error {
	if !validCalendarEventCreateConflictKind(kind) {
		return ErrEventConflict
	}
	return &CalendarEventCreateConflictError{kind: kind}
}

func validCalendarEventCreateConflictKind(kind CalendarEventCreateConflictKind) bool {
	switch kind {
	case CalendarEventCreateConflictIdempotencyKeyReused,
		CalendarEventCreateConflictScheduledCallLinkMismatch,
		CalendarEventCreateConflictScheduledCallNotAttachable:
		return true
	default:
		return false
	}
}
