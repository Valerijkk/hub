package calendar

// EventErrorKind определяет безопасную категорию ошибки канонического календарного события.
type EventErrorKind string

const (
	// EventErrorNotFound означает отсутствие события в разрешённой области доступа.
	EventErrorNotFound EventErrorKind = "not_found"
	// EventErrorForbidden означает запрет операции без раскрытия состояния другого события.
	EventErrorForbidden EventErrorKind = "forbidden"
	// EventErrorConflict означает конфликт версии, состояния или идемпотентной квитанции.
	EventErrorConflict EventErrorKind = "conflict"
	// EventErrorInvalid означает некорректную команду или значение доменного контракта.
	EventErrorInvalid EventErrorKind = "invalid"
	// EventErrorTransactionClosed означает попытку использовать завершённую транзакционную область.
	EventErrorTransactionClosed EventErrorKind = "transaction_closed"
	// EventErrorInternal is a fail-closed capability, persistence or invariant
	// failure that must never be presented as client-invalid input.
	EventErrorInternal EventErrorKind = "internal"
)

var (
	// ErrEventNotFound используется для несуществующего или недоступного календарного события.
	ErrEventNotFound = EventError{kind: EventErrorNotFound}
	// ErrEventForbidden используется для запрещённой календарной операции.
	ErrEventForbidden = EventError{kind: EventErrorForbidden}
	// ErrEventConflict используется для конфликтующей календарной операции.
	ErrEventConflict = EventError{kind: EventErrorConflict}
	// ErrEventInvalid используется для некорректной календарной команды.
	ErrEventInvalid = EventError{kind: EventErrorInvalid}
	// ErrEventTransactionClosed используется после возврата callback транзакции.
	ErrEventTransactionClosed = EventError{kind: EventErrorTransactionClosed}
	// ErrEventInternal preserves a safe 500 boundary without leaking a storage,
	// capability, source or invariant cause.
	ErrEventInternal = EventError{kind: EventErrorInternal}
)

// EventError представляет типизированную, не содержащую входных данных ошибку календарного домена.
type EventError struct {
	kind EventErrorKind
}

// Error возвращает стабильное безопасное описание без идентификаторов, ключей и пользовательских данных.
func (err EventError) Error() string {
	switch err.kind {
	case EventErrorNotFound:
		return "calendar event not found"
	case EventErrorForbidden:
		return "calendar event access denied"
	case EventErrorConflict:
		return "calendar event conflict"
	case EventErrorInvalid:
		return "calendar event invalid"
	case EventErrorTransactionClosed:
		return "calendar event transaction closed"
	case EventErrorInternal:
		return "calendar event internal failure"
	default:
		return "calendar event error"
	}
}

// Kind возвращает безопасную категорию ошибки.
func (err EventError) Kind() EventErrorKind {
	return err.kind
}

// Is сопоставляет ошибки календарного домена по категории для errors.Is.
func (err EventError) Is(target error) bool {
	switch value := target.(type) {
	case EventError:
		return err.kind == value.kind
	case *EventError:
		return value != nil && err.kind == value.kind
	default:
		return false
	}
}

// MeetingCreateConflictKind is the closed safe 409 result set for the meeting
// create vertical. It intentionally carries no participant, time or source.
type MeetingCreateConflictKind string

const (
	MeetingCreateConflictIdempotencyKeyReused MeetingCreateConflictKind = "idempotency_key_reused"
	MeetingCreateConflictSlotUnavailable      MeetingCreateConflictKind = "slot_unavailable"
)

// MeetingCreateConflictError exposes only one stable machine-readable kind.
type MeetingCreateConflictError struct {
	kind MeetingCreateConflictKind
}

func (err MeetingCreateConflictError) Error() string {
	switch err.kind {
	case MeetingCreateConflictIdempotencyKeyReused:
		return "meeting idempotency key reused"
	case MeetingCreateConflictSlotUnavailable:
		return "meeting slot unavailable"
	default:
		return "meeting create conflict"
	}
}

func (err MeetingCreateConflictError) Unwrap() error { return ErrEventConflict }
func (err MeetingCreateConflictError) Kind() MeetingCreateConflictKind { return err.kind }

func (err MeetingCreateConflictError) Is(target error) bool {
	switch value := target.(type) {
	case MeetingCreateConflictError:
		return err.kind == value.kind
	case *MeetingCreateConflictError:
		return value != nil && err.kind == value.kind
	default:
		return false
	}
}

func newMeetingCreateConflict(kind MeetingCreateConflictKind) error {
	switch kind {
	case MeetingCreateConflictIdempotencyKeyReused, MeetingCreateConflictSlotUnavailable:
		return MeetingCreateConflictError{kind: kind}
	default:
		return ErrEventInternal
	}
}
