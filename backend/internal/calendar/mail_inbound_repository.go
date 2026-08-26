package calendar

import (
	"context"
	"crypto/sha256"
	"reflect"

	"github.com/google/uuid"
)

const (
	calendarInboundMessageDueClaimLimitMinimum = 1
	calendarInboundMessageDueClaimLimitMaximum = 50

	calendarInboundMessageFinalizedCleanupLimitMinimum = 1
	calendarInboundMessageFinalizedCleanupLimitMaximum = 500
)

// CalendarInboundMessageRegistration is the content-free durable identity of
// one calendar part projected from an inbound Mail message.
type CalendarInboundMessageRegistration struct {
	envelopeID      uuid.UUID
	organizationID  uuid.UUID
	mailboxID       uuid.UUID
	mailMessageID   uuid.UUID
	partLocatorHash [sha256.Size]byte
}

// NewCalendarInboundMessageRegistration validates the immutable receipt key
// before it crosses the Calendar repository boundary.
func NewCalendarInboundMessageRegistration(
	envelopeID uuid.UUID,
	organizationID uuid.UUID,
	mailboxID uuid.UUID,
	mailMessageID uuid.UUID,
	partLocatorHash [sha256.Size]byte,
) (CalendarInboundMessageRegistration, error) {
	if envelopeID == uuid.Nil ||
		organizationID == uuid.Nil ||
		mailboxID == uuid.Nil ||
		mailMessageID == uuid.Nil ||
		partLocatorHash == [sha256.Size]byte{} {
		return CalendarInboundMessageRegistration{}, ErrEventInvalid
	}
	return CalendarInboundMessageRegistration{
		envelopeID:      envelopeID,
		organizationID:  organizationID,
		mailboxID:       mailboxID,
		mailMessageID:   mailMessageID,
		partLocatorHash: partLocatorHash,
	}, nil
}

// EnvelopeID returns the original mail.received.v1 event identity.
func (registration CalendarInboundMessageRegistration) EnvelopeID() uuid.UUID {
	if !registration.valid() {
		return uuid.Nil
	}
	return registration.envelopeID
}

// OrganizationID returns the immutable Calendar organization boundary.
func (registration CalendarInboundMessageRegistration) OrganizationID() uuid.UUID {
	if !registration.valid() {
		return uuid.Nil
	}
	return registration.organizationID
}

// MailboxID returns the immutable receiving Mail account identity.
func (registration CalendarInboundMessageRegistration) MailboxID() uuid.UUID {
	if !registration.valid() {
		return uuid.Nil
	}
	return registration.mailboxID
}

// MailMessageID returns the immutable projected Mail message identity.
func (registration CalendarInboundMessageRegistration) MailMessageID() uuid.UUID {
	if !registration.valid() {
		return uuid.Nil
	}
	return registration.mailMessageID
}

// LocatorHash returns a by-value stable part locator hash.
func (registration CalendarInboundMessageRegistration) LocatorHash() [sha256.Size]byte {
	if !registration.valid() {
		return [sha256.Size]byte{}
	}
	return registration.partLocatorHash
}

func (registration CalendarInboundMessageRegistration) valid() bool {
	return registration.envelopeID != uuid.Nil &&
		registration.organizationID != uuid.Nil &&
		registration.mailboxID != uuid.Nil &&
		registration.mailMessageID != uuid.Nil &&
		registration.partLocatorHash != [sha256.Size]byte{}
}

// CalendarInboundMessageWork is the sealed capability granted only after a
// durable receipt has been claimed with one live processing lease.
type CalendarInboundMessageWork struct {
	registration CalendarInboundMessageRegistration
	claim        CalendarInboundMessageClaim
}

// NewCalendarInboundMessageWork binds a validated durable receipt identity to
// its exact processing claim without exposing receipt state or content.
func NewCalendarInboundMessageWork(
	registration CalendarInboundMessageRegistration,
	claim CalendarInboundMessageClaim,
) (CalendarInboundMessageWork, error) {
	if !registration.valid() || !claim.valid() {
		return CalendarInboundMessageWork{}, ErrEventInvalid
	}
	return CalendarInboundMessageWork{
		registration: registration,
		claim:        claim,
	}, nil
}

// Registration returns a by-value copy of the receipt registration.
func (work CalendarInboundMessageWork) Registration() CalendarInboundMessageRegistration {
	if !work.valid() {
		return CalendarInboundMessageRegistration{}
	}
	return work.registration
}

// Claim returns a by-value copy of the exact processing lease capability.
func (work CalendarInboundMessageWork) Claim() CalendarInboundMessageClaim {
	if !work.valid() {
		return CalendarInboundMessageClaim{}
	}
	return work.claim
}

func (work CalendarInboundMessageWork) valid() bool {
	return work.registration.valid() && work.claim.valid()
}

// CalendarInboundMessageClaimOutcome is the closed outcome of a repository
// operation that always completed a durable receipt decision on success.
type CalendarInboundMessageClaimOutcome string

const (
	CalendarInboundMessageClaimOutcomeClaimed     CalendarInboundMessageClaimOutcome = "claimed"
	CalendarInboundMessageClaimOutcomeDurableNoop CalendarInboundMessageClaimOutcome = "durable_noop"
)

// CalendarInboundMessageClaimResult contains either one sealed work
// capability or a durable no-op result. Repository errors are the only
// no-durable outcome.
type CalendarInboundMessageClaimResult struct {
	outcome CalendarInboundMessageClaimOutcome
	work    CalendarInboundMessageWork
}

// NewCalendarInboundMessageClaimedResult constructs the only result that
// grants a caller one claimed receipt capability.
func NewCalendarInboundMessageClaimedResult(
	work CalendarInboundMessageWork,
) (CalendarInboundMessageClaimResult, error) {
	if !work.valid() {
		return CalendarInboundMessageClaimResult{}, ErrEventInvalid
	}
	return CalendarInboundMessageClaimResult{
		outcome: CalendarInboundMessageClaimOutcomeClaimed,
		work:    work,
	}, nil
}

// NewCalendarInboundMessageDurableNoopResult constructs a successful durable
// result that deliberately grants no work capability.
func NewCalendarInboundMessageDurableNoopResult() CalendarInboundMessageClaimResult {
	return CalendarInboundMessageClaimResult{
		outcome: CalendarInboundMessageClaimOutcomeDurableNoop,
	}
}

// Outcome returns the closed durable result or its zero value when malformed.
func (result CalendarInboundMessageClaimResult) Outcome() CalendarInboundMessageClaimOutcome {
	if !result.valid() {
		return ""
	}
	return result.outcome
}

// Work returns a defensive sealed capability only for a valid claimed result.
func (result CalendarInboundMessageClaimResult) Work() (CalendarInboundMessageWork, bool) {
	if !result.valid() || result.outcome != CalendarInboundMessageClaimOutcomeClaimed {
		return CalendarInboundMessageWork{}, false
	}
	return CalendarInboundMessageWork{
		registration: result.work.registration,
		claim:        result.work.claim,
	}, true
}

func (result CalendarInboundMessageClaimResult) valid() bool {
	switch result.outcome {
	case CalendarInboundMessageClaimOutcomeClaimed:
		return result.work.valid()
	case CalendarInboundMessageClaimOutcomeDurableNoop:
		return result.work == (CalendarInboundMessageWork{})
	default:
		return false
	}
}

// CalendarInboundMessageRepository owns durable receipt registration and
// database-clock claim policy. Callers never provide a timestamp or lease.
type CalendarInboundMessageRepository interface {
	RegisterAndClaimCalendarInboundMessage(
		context.Context,
		CalendarInboundMessageRegistration,
	) (CalendarInboundMessageClaimResult, error)
	ClaimDueCalendarInboundMessages(context.Context, int) ([]CalendarInboundMessageWork, error)
	CleanupFinalizedCalendarInboundMessages(context.Context, int) (int, error)
}

func validCalendarInboundMessageDueClaimLimit(limit int) bool {
	return limit >= calendarInboundMessageDueClaimLimitMinimum &&
		limit <= calendarInboundMessageDueClaimLimitMaximum
}

func validCalendarInboundMessageFinalizedCleanupLimit(limit int) bool {
	return limit >= calendarInboundMessageFinalizedCleanupLimitMinimum &&
		limit <= calendarInboundMessageFinalizedCleanupLimitMaximum
}

func nilCalendarInboundMessageRepository(repository CalendarInboundMessageRepository) bool {
	if repository == nil {
		return true
	}
	value := reflect.ValueOf(repository)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
