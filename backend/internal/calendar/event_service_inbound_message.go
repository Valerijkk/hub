package calendar

import (
	"reflect"

	"github.com/google/uuid"
)

// CalendarInboundMessageClaim identifies one processing lease that was
// acquired from a durable Calendar inbound-message receipt.
type CalendarInboundMessageClaim struct {
	ReceiptID       uuid.UUID
	LeaseToken      uuid.UUID
	LeaseGeneration int64
}

// CalendarInboundMessageScope is an opaque capability bound to one live event
// transaction and one exact receipt claim.
type CalendarInboundMessageScope struct {
	claim       CalendarInboundMessageClaim
	transaction *EventTransaction
}

// CalendarInboundObservedState is the only non-mutating terminal state that
// may be recorded after a Calendar decision has already been observed.
type CalendarInboundObservedState string

const (
	CalendarInboundObservedIgnoredStale CalendarInboundObservedState = "ignored_stale"
	CalendarInboundObservedDuplicate    CalendarInboundObservedState = "duplicate"
)

// CalendarInboundMessageErrorCode is the closed, content-free receipt error
// vocabulary. It must never hold a provider, MIME, iTIP, SQL, or free-form
// error value.
type CalendarInboundMessageErrorCode string

const (
	CalendarInboundMessageErrorMessageUnavailable  CalendarInboundMessageErrorCode = "message_unavailable"
	CalendarInboundMessageErrorPartUnavailable     CalendarInboundMessageErrorCode = "part_unavailable"
	CalendarInboundMessageErrorMalformedITIP       CalendarInboundMessageErrorCode = "malformed_itip"
	CalendarInboundMessageErrorUnsupportedMethod   CalendarInboundMessageErrorCode = "unsupported_method"
	CalendarInboundMessageErrorUnsupportedVersion  CalendarInboundMessageErrorCode = "unsupported_version"
	CalendarInboundMessageErrorUntrustedSender     CalendarInboundMessageErrorCode = "untrusted_sender"
	CalendarInboundMessageErrorOrganizerMismatch   CalendarInboundMessageErrorCode = "organizer_mismatch"
	CalendarInboundMessageErrorEqualSequenceConflict CalendarInboundMessageErrorCode = "equal_sequence_conflict"
	CalendarInboundMessageErrorMissingMasterExpired  CalendarInboundMessageErrorCode = "missing_master_expired"
	CalendarInboundMessageErrorResourceLimit         CalendarInboundMessageErrorCode = "resource_limit"
	CalendarInboundMessageErrorRetryExhausted        CalendarInboundMessageErrorCode = "retry_exhausted"
	CalendarInboundMessageErrorEngineUnavailable     CalendarInboundMessageErrorCode = "engine_unavailable"
	CalendarInboundMessageErrorDatabaseUnavailable   CalendarInboundMessageErrorCode = "database_unavailable"
	CalendarInboundMessageErrorInternal              CalendarInboundMessageErrorCode = "internal_error"
)

type calendarInboundMessageCompletionKind string

const (
	calendarInboundMessageCompletionApplied     calendarInboundMessageCompletionKind = "applied"
	calendarInboundMessageCompletionTombstoned  calendarInboundMessageCompletionKind = "tombstoned"
	calendarInboundMessageCompletionObserved    calendarInboundMessageCompletionKind = "observed"
	calendarInboundMessageCompletionQuarantined calendarInboundMessageCompletionKind = "quarantined"
	calendarInboundMessageCompletionDeadLetter  calendarInboundMessageCompletionKind = "dead_letter"
	calendarInboundMessageCompletionRetry       calendarInboundMessageCompletionKind = "retry"
	calendarInboundMessageCompletionDeferred    calendarInboundMessageCompletionKind = "deferred_missing_master"
)

// calendarInboundMessageReceiptSnapshot carries the canonical component
// identity that may be written when a receipt is completed. Its zero value is
// unknown and preserves any existing durable snapshot.
type calendarInboundMessageReceiptSnapshot struct {
	calendarComponentScope string
	recurrenceID           *string
}

const (
	calendarInboundMessageReceiptSnapshotScopeSeries     = "series"
	calendarInboundMessageReceiptSnapshotScopeOccurrence = "occurrence"
)

// CalendarInboundMessageCompletion is an opaque, constructor-only receipt
// disposition. Its optional event identifier is copied at every boundary.
type CalendarInboundMessageCompletion struct {
	kind                  calendarInboundMessageCompletionKind
	observedState         CalendarInboundObservedState
	errorCode             CalendarInboundMessageErrorCode
	calendarEventID       *uuid.UUID
	deferredObjectStateID uuid.UUID
	receiptSnapshot       calendarInboundMessageReceiptSnapshot
}

// NewCalendarInboundAppliedCompletion constructs an applied receipt outcome
// for one resolved Calendar event.
func NewCalendarInboundAppliedCompletion(calendarEventID uuid.UUID) (CalendarInboundMessageCompletion, error) {
	if calendarEventID == uuid.Nil {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return CalendarInboundMessageCompletion{
		kind:            calendarInboundMessageCompletionApplied,
		calendarEventID: cloneCalendarInboundMessageUUIDPointer(&calendarEventID),
	}, nil
}

// NewCalendarInboundTombstonedCompletion constructs a terminal CANCEL
// disposition whose resolved Calendar event identity remains optional.
func NewCalendarInboundTombstonedCompletion(calendarEventID *uuid.UUID) (CalendarInboundMessageCompletion, error) {
	if !validCalendarInboundMessageUUIDPointer(calendarEventID) {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return CalendarInboundMessageCompletion{
		kind:            calendarInboundMessageCompletionTombstoned,
		calendarEventID: cloneCalendarInboundMessageUUIDPointer(calendarEventID),
	}, nil
}

// NewCalendarInboundObservedCompletion constructs a duplicate or stale
// disposition without an error code.
func NewCalendarInboundObservedCompletion(state CalendarInboundObservedState, calendarEventID *uuid.UUID) (CalendarInboundMessageCompletion, error) {
	if !validCalendarInboundObservedState(state) || !validCalendarInboundMessageUUIDPointer(calendarEventID) {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return CalendarInboundMessageCompletion{
		kind:            calendarInboundMessageCompletionObserved,
		observedState:   state,
		calendarEventID: cloneCalendarInboundMessageUUIDPointer(calendarEventID),
	}, nil
}

// NewCalendarInboundQuarantinedCompletion constructs a definitive safe
// quarantine disposition.
func NewCalendarInboundQuarantinedCompletion(code CalendarInboundMessageErrorCode) (CalendarInboundMessageCompletion, error) {
	if !validCalendarInboundQuarantinedErrorCode(code) {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return CalendarInboundMessageCompletion{kind: calendarInboundMessageCompletionQuarantined, errorCode: code}, nil
}

// NewCalendarInboundDeadLetterCompletion constructs a terminal retry or
// unavailable disposition after the receipt policy has exhausted it.
func NewCalendarInboundDeadLetterCompletion(code CalendarInboundMessageErrorCode) (CalendarInboundMessageCompletion, error) {
	if !validCalendarInboundDeadLetterErrorCode(code) {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return CalendarInboundMessageCompletion{kind: calendarInboundMessageCompletionDeadLetter, errorCode: code}, nil
}

// NewCalendarInboundRetryCompletion constructs a bounded retriable receipt
// disposition. The repository derives its due time from the durable attempt.
func NewCalendarInboundRetryCompletion(code CalendarInboundMessageErrorCode) (CalendarInboundMessageCompletion, error) {
	if !validCalendarInboundRetryErrorCode(code) {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return CalendarInboundMessageCompletion{kind: calendarInboundMessageCompletionRetry, errorCode: code}, nil
}

// NewCalendarInboundDeferredCompletion constructs a missing-master deferral
// bound to the persisted inbound object state, with no event identity or error.
func NewCalendarInboundDeferredCompletion(deferredObjectStateID uuid.UUID) (CalendarInboundMessageCompletion, error) {
	if deferredObjectStateID == uuid.Nil {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return CalendarInboundMessageCompletion{
		kind:                  calendarInboundMessageCompletionDeferred,
		deferredObjectStateID: deferredObjectStateID,
	}, nil
}

func (claim CalendarInboundMessageClaim) valid() bool {
	return claim.ReceiptID != uuid.Nil && claim.LeaseToken != uuid.Nil && claim.LeaseGeneration >= 1
}

func (scope CalendarInboundMessageScope) zero() bool {
	return scope.claim == (CalendarInboundMessageClaim{}) && scope.transaction == nil
}

func (scope CalendarInboundMessageScope) isBackendEchoFor(claim CalendarInboundMessageClaim) bool {
	return scope.transaction == nil && scope.claim == claim
}

func (scope CalendarInboundMessageScope) isBoundTo(transaction *EventTransaction) bool {
	return transaction != nil && scope.transaction == transaction && scope.claim.valid()
}

func (completion CalendarInboundMessageCompletion) valid() bool {
	if !validCalendarInboundMessageUUIDPointer(completion.calendarEventID) {
		return false
	}
	if !completion.receiptSnapshot.valid() {
		return false
	}
	switch completion.kind {
	case calendarInboundMessageCompletionApplied:
		return completion.observedState == "" && completion.errorCode == "" && completion.calendarEventID != nil && completion.deferredObjectStateID == uuid.Nil
	case calendarInboundMessageCompletionTombstoned:
		return completion.observedState == "" && completion.errorCode == "" && completion.deferredObjectStateID == uuid.Nil
	case calendarInboundMessageCompletionObserved:
		return validCalendarInboundObservedState(completion.observedState) && completion.errorCode == "" && completion.deferredObjectStateID == uuid.Nil
	case calendarInboundMessageCompletionQuarantined:
		return completion.observedState == "" && completion.calendarEventID == nil && validCalendarInboundQuarantinedErrorCode(completion.errorCode) && completion.deferredObjectStateID == uuid.Nil
	case calendarInboundMessageCompletionDeadLetter:
		return completion.observedState == "" && completion.calendarEventID == nil && validCalendarInboundDeadLetterErrorCode(completion.errorCode) && completion.deferredObjectStateID == uuid.Nil
	case calendarInboundMessageCompletionRetry:
		return completion.observedState == "" && completion.calendarEventID == nil && validCalendarInboundRetryErrorCode(completion.errorCode) && completion.deferredObjectStateID == uuid.Nil
	case calendarInboundMessageCompletionDeferred:
		return completion.observedState == "" &&
			completion.errorCode == "" &&
			completion.calendarEventID == nil &&
			completion.deferredObjectStateID != uuid.Nil &&
			(completion.receiptSnapshot.calendarComponentScope == "" ||
				completion.receiptSnapshot.calendarComponentScope == calendarInboundMessageReceiptSnapshotScopeOccurrence)
	default:
		return false
	}
}

func cloneCalendarInboundMessageCompletion(completion CalendarInboundMessageCompletion) CalendarInboundMessageCompletion {
	if !completion.valid() {
		return CalendarInboundMessageCompletion{}
	}
	completion.calendarEventID = cloneCalendarInboundMessageUUIDPointer(completion.calendarEventID)
	completion.receiptSnapshot = cloneCalendarInboundMessageReceiptSnapshot(completion.receiptSnapshot)
	return completion
}

func newCalendarInboundMessageReceiptSnapshot(
	hasRecurrenceID bool,
	recurrenceID string,
) (calendarInboundMessageReceiptSnapshot, error) {
	if !hasRecurrenceID {
		if recurrenceID != "" {
			return calendarInboundMessageReceiptSnapshot{}, ErrEventInvalid
		}
		return calendarInboundMessageReceiptSnapshot{
			calendarComponentScope: calendarInboundMessageReceiptSnapshotScopeSeries,
		}, nil
	}
	if !validCalendarInboundRecurrenceID(recurrenceID) {
		return calendarInboundMessageReceiptSnapshot{}, ErrEventInvalid
	}
	return calendarInboundMessageReceiptSnapshot{
		calendarComponentScope: calendarInboundMessageReceiptSnapshotScopeOccurrence,
		recurrenceID:           cloneCalendarInboundMessageStringPointer(&recurrenceID),
	}, nil
}

func (snapshot calendarInboundMessageReceiptSnapshot) valid() bool {
	switch snapshot.calendarComponentScope {
	case "":
		return snapshot.recurrenceID == nil
	case calendarInboundMessageReceiptSnapshotScopeSeries:
		return snapshot.recurrenceID == nil
	case calendarInboundMessageReceiptSnapshotScopeOccurrence:
		return snapshot.recurrenceID != nil && validCalendarInboundRecurrenceID(*snapshot.recurrenceID)
	default:
		return false
	}
}

func cloneCalendarInboundMessageReceiptSnapshot(
	snapshot calendarInboundMessageReceiptSnapshot,
) calendarInboundMessageReceiptSnapshot {
	if !snapshot.valid() {
		return calendarInboundMessageReceiptSnapshot{}
	}
	snapshot.recurrenceID = cloneCalendarInboundMessageStringPointer(snapshot.recurrenceID)
	return snapshot
}

func calendarInboundMessageCompletionWithReceiptSnapshot(
	completion CalendarInboundMessageCompletion,
	snapshot calendarInboundMessageReceiptSnapshot,
) (CalendarInboundMessageCompletion, error) {
	if !completion.valid() || !snapshot.valid() {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	if completion.kind == calendarInboundMessageCompletionDeferred &&
		snapshot.calendarComponentScope != calendarInboundMessageReceiptSnapshotScopeOccurrence {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	completion.receiptSnapshot = cloneCalendarInboundMessageReceiptSnapshot(snapshot)
	if !completion.valid() {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	return completion, nil
}

func validCalendarInboundObservedState(state CalendarInboundObservedState) bool {
	return state == CalendarInboundObservedIgnoredStale || state == CalendarInboundObservedDuplicate
}

func validCalendarInboundMessageErrorCode(code CalendarInboundMessageErrorCode) bool {
	switch code {
	case CalendarInboundMessageErrorMessageUnavailable,
		CalendarInboundMessageErrorPartUnavailable,
		CalendarInboundMessageErrorMalformedITIP,
		CalendarInboundMessageErrorUnsupportedMethod,
		CalendarInboundMessageErrorUnsupportedVersion,
		CalendarInboundMessageErrorUntrustedSender,
		CalendarInboundMessageErrorOrganizerMismatch,
		CalendarInboundMessageErrorEqualSequenceConflict,
		CalendarInboundMessageErrorMissingMasterExpired,
		CalendarInboundMessageErrorResourceLimit,
		CalendarInboundMessageErrorRetryExhausted,
		CalendarInboundMessageErrorEngineUnavailable,
		CalendarInboundMessageErrorDatabaseUnavailable,
		CalendarInboundMessageErrorInternal:
		return true
	default:
		return false
	}
}

func validCalendarInboundQuarantinedErrorCode(code CalendarInboundMessageErrorCode) bool {
	if !validCalendarInboundMessageErrorCode(code) {
		return false
	}
	switch code {
	case CalendarInboundMessageErrorMalformedITIP,
		CalendarInboundMessageErrorUnsupportedMethod,
		CalendarInboundMessageErrorUnsupportedVersion,
		CalendarInboundMessageErrorUntrustedSender,
		CalendarInboundMessageErrorOrganizerMismatch,
		CalendarInboundMessageErrorEqualSequenceConflict,
		CalendarInboundMessageErrorResourceLimit,
		CalendarInboundMessageErrorInternal:
		return true
	default:
		return false
	}
}

func validCalendarInboundDeadLetterErrorCode(code CalendarInboundMessageErrorCode) bool {
	if !validCalendarInboundMessageErrorCode(code) {
		return false
	}
	switch code {
	case CalendarInboundMessageErrorMessageUnavailable,
		CalendarInboundMessageErrorPartUnavailable,
		CalendarInboundMessageErrorMissingMasterExpired,
		CalendarInboundMessageErrorResourceLimit,
		CalendarInboundMessageErrorRetryExhausted,
		CalendarInboundMessageErrorInternal:
		return true
	default:
		return false
	}
}

func validCalendarInboundRetryErrorCode(code CalendarInboundMessageErrorCode) bool {
	if !validCalendarInboundMessageErrorCode(code) {
		return false
	}
	return code == CalendarInboundMessageErrorEngineUnavailable || code == CalendarInboundMessageErrorDatabaseUnavailable || code == CalendarInboundMessageErrorInternal
}

func cloneCalendarInboundMessageUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneCalendarInboundMessageStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func validCalendarInboundMessageUUIDPointer(value *uuid.UUID) bool {
	return value == nil || *value != uuid.Nil
}

func nilCalendarInboundMessageTransactionBackend(backend CalendarInboundMessageTransactionBackend) bool {
	if backend == nil {
		return true
	}
	value := reflect.ValueOf(backend)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
