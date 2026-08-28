package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"reflect"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/ical"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const calendarInboundReplyCanonicalVersion uint8 = 1

// CalendarInboundReplyInput is the complete authenticated-envelope input for
// a parsed inbound iTIP REPLY. OwnerID is a mailbox ownership selector, never
// an authenticated Hub mutation actor.
type CalendarInboundReplyInput struct {
	OwnerID uuid.UUID
	Sender  string
	Message ITIPMessage
}

type CalendarInboundReplyOutcome string

const (
	CalendarInboundReplyApplied           CalendarInboundReplyOutcome = "applied"
	CalendarInboundReplyObservedNoChange  CalendarInboundReplyOutcome = "observed_no_change"
	CalendarInboundReplyDuplicate         CalendarInboundReplyOutcome = "duplicate"
	CalendarInboundReplyStale             CalendarInboundReplyOutcome = "stale"
	CalendarInboundReplyQuarantined       CalendarInboundReplyOutcome = "quarantined"
)

// CalendarInboundReplyResult intentionally exposes only an outcome; inbound
// mail processing must not disclose event, attendee, UID or hash state.
type CalendarInboundReplyResult struct {
	outcome         CalendarInboundReplyOutcome
	calendarEventID *uuid.UUID
}

func (result CalendarInboundReplyResult) Outcome() CalendarInboundReplyOutcome { return result.outcome }

func (result CalendarInboundReplyResult) valid() bool {
	if !validCalendarInboundMessageUUIDPointer(result.calendarEventID) {
		return false
	}
	switch result.outcome {
	case CalendarInboundReplyApplied, CalendarInboundReplyObservedNoChange, CalendarInboundReplyDuplicate, CalendarInboundReplyStale, CalendarInboundReplyQuarantined:
		return true
	default:
		return false
	}
}

func newCalendarInboundReplyResult(outcome CalendarInboundReplyOutcome, calendarEventID *uuid.UUID) (CalendarInboundReplyResult, error) {
	result := CalendarInboundReplyResult{
		outcome:         outcome,
		calendarEventID: cloneCalendarInboundMessageUUIDPointer(calendarEventID),
	}
	if !result.valid() {
		return CalendarInboundReplyResult{}, ErrEventInvalid
	}
	return result, nil
}

func (result CalendarInboundReplyResult) calendarInboundMessageCompletion() (CalendarInboundMessageCompletion, error) {
	if !result.valid() {
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
	switch result.outcome {
	case CalendarInboundReplyApplied, CalendarInboundReplyObservedNoChange:
		if result.calendarEventID == nil {
			return CalendarInboundMessageCompletion{}, ErrEventInvalid
		}
		return NewCalendarInboundAppliedCompletion(*result.calendarEventID)
	case CalendarInboundReplyDuplicate:
		return NewCalendarInboundObservedCompletion(CalendarInboundObservedDuplicate, result.calendarEventID)
	case CalendarInboundReplyStale:
		return NewCalendarInboundObservedCompletion(CalendarInboundObservedIgnoredStale, result.calendarEventID)
	case CalendarInboundReplyQuarantined:
		return NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorOrganizerMismatch)
	default:
		return CalendarInboundMessageCompletion{}, ErrEventInvalid
	}
}

type CalendarInboundReplyReceiptOutcome string

const (
	CalendarInboundReplyReceiptApplied           CalendarInboundReplyReceiptOutcome = "applied"
	CalendarInboundReplyReceiptObservedNoChange  CalendarInboundReplyReceiptOutcome = "observed_no_change"
)

// CalendarInboundReplyReceiptKey keeps a versioned sender-bound duplicate key
// opaque. It never retains raw iTIP input or the parser-only payload hash.
type CalendarInboundReplyReceiptKey struct {
	canonicalVersion uint8
	payloadHash      [sha256.Size]byte
}

func NewCalendarInboundReplyReceiptKey(version uint8, sender string, parserPayloadHash [sha256.Size]byte) (CalendarInboundReplyReceiptKey, error) {
	if version != calendarInboundReplyCanonicalVersion || len(sender) > ical.MaxITIPEmailBytes || !utf8.ValidString(sender) {
		return CalendarInboundReplyReceiptKey{}, ErrEventInvalid
	}
	normalizedSender, err := ical.NormalizeITIPEmail(sender)
	if err != nil || normalizedSender != sender {
		return CalendarInboundReplyReceiptKey{}, ErrEventInvalid
	}
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(sender)))
	hash := sha256.New()
	_, _ = hash.Write([]byte("hub.calendar.itip.reply"))
	_, _ = hash.Write([]byte{0, version})
	_, _ = hash.Write(length[:])
	_, _ = hash.Write([]byte(sender))
	_, _ = hash.Write(parserPayloadHash[:])
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return CalendarInboundReplyReceiptKey{canonicalVersion: version, payloadHash: result}, nil
}

func (key CalendarInboundReplyReceiptKey) CanonicalVersion() uint8 { return key.canonicalVersion }
func (key CalendarInboundReplyReceiptKey) PayloadHash() [sha256.Size]byte { return key.payloadHash }
func (key CalendarInboundReplyReceiptKey) valid() bool { return key.canonicalVersion == calendarInboundReplyCanonicalVersion }

// ApplyInboundITIPReply applies one exact, owner-bound inbound REPLY without
// entering actor-bound mutation receipt or outbound transport paths.
func (service *EventService) ApplyInboundITIPReply(ctx context.Context, input CalendarInboundReplyInput) (result CalendarInboundReplyResult, err error) {
	defer func() {
		if recover() != nil {
			result, err = CalendarInboundReplyResult{}, ErrEventInvalid
		}
	}()
	if nilCalendarInboundReplyDependency(ctx) || service == nil || nilCalendarInboundReplyDependency(service.repository) || nilCalendarInboundReplyDependency(service.clock) || nilCalendarInboundReplyDependency(service.ids) || input.OwnerID == uuid.Nil {
		return CalendarInboundReplyResult{}, ErrEventInvalid
	}
	err = service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		var applyErr error
		result, applyErr = service.applyInboundITIPReplyInTransaction(ctx, transaction, input)
		return applyErr
	})
	if err != nil {
		return CalendarInboundReplyResult{}, calendarInboundReplySafeError(err)
	}
	if !result.valid() {
		return CalendarInboundReplyResult{}, ErrEventInvalid
	}
	return result, nil
}

func (service *EventService) applyInboundITIPReplyInTransaction(ctx context.Context, transaction *EventTransaction, input CalendarInboundReplyInput) (result CalendarInboundReplyResult, err error) {
	defer func() {
		if recover() != nil {
			result, err = CalendarInboundReplyResult{}, ErrEventInvalid
		}
	}()
	if nilCalendarInboundReplyDependency(ctx) || service == nil || nilCalendarInboundReplyDependency(service.repository) || nilCalendarInboundReplyDependency(service.clock) || nilCalendarInboundReplyDependency(service.ids) || input.OwnerID == uuid.Nil {
		return CalendarInboundReplyResult{}, ErrEventInvalid
	}
	canonical, normalized, sender, accepted := calendarInboundReplyPreflight(input.Sender, input.Message)
	if !accepted {
		return newCalendarInboundReplyResult(CalendarInboundReplyQuarantined, nil)
	}
	receipt, receiptErr := NewCalendarInboundReplyReceiptKey(canonical.CanonicalVersion, sender, canonical.PayloadHash)
	if receiptErr != nil {
		return CalendarInboundReplyResult{}, ErrEventInvalid
	}

	err = func() error {
		if transaction == nil {
			return ErrEventInvalid
		}
		setResult := func(outcome CalendarInboundReplyOutcome, calendarEventID *uuid.UUID) error {
			candidate, resultErr := newCalendarInboundReplyResult(outcome, calendarEventID)
			if resultErr != nil {
				return ErrEventInvalid
			}
			result = candidate
			return nil
		}
		scope, found, scopeErr := transaction.ResolveCalendarInboundReplyScope(ctx, input.OwnerID)
		if scopeErr != nil {
			return scopeErr
		}
		if !found {
			return setResult(CalendarInboundReplyQuarantined, nil)
		}
		scope = scope.withOrganizerEmail(normalized.Organizer.Email)
		authorization, resolveErr := transaction.ResolveCalendarInboundReply(ctx, scope, CalendarInboundReplyAuthorizationRequest{
			UID: canonical.UID, HasRecurrenceID: canonical.HasRecurrenceID, RecurrenceID: canonical.RecurrenceID,
			Sender: sender, EventSequence: uint(canonical.Sequence), Receipt: receipt,
		})
		if resolveErr != nil {
			return resolveErr
		}
		if !authorization.Matched {
			return setResult(CalendarInboundReplyQuarantined, nil)
		}
		// Existing receipt has strict precedence over all mutable comparisons so
		// replaying an older A after a later B cannot roll the attendee back.
		if authorization.ExistingReceipt {
			return setResult(CalendarInboundReplyDuplicate, &authorization.Current.EventID)
		}
		effectiveSequence, policyOutcome, policyErr := calendarInboundReplyAuthorize(scope, authorization, normalized, sender)
		if policyErr != nil {
			return policyErr
		}
		if policyOutcome != "" {
			return setResult(policyOutcome, &authorization.Current.EventID)
		}
		incomingSequence := uint(canonical.Sequence)
		if incomingSequence < effectiveSequence {
			return setResult(CalendarInboundReplyStale, &authorization.Current.EventID)
		}
		if incomingSequence > effectiveSequence || authorization.DistinctReceiptCount == calendarInboundReplyMaxDistinctReceipts {
			return setResult(CalendarInboundReplyQuarantined, &authorization.Current.EventID)
		}

		prior := authorization.Target
		if prior.Status == normalized.Attendees[0].Status {
			command, commandErr := NewCalendarInboundReplyPersistenceCommand(prior, nil, effectiveSequence, receipt, CalendarInboundReplyReceiptObservedNoChange, nil)
			if commandErr != nil {
				return ErrEventInvalid
			}
			if _, persistErr := transaction.PersistCalendarInboundReply(ctx, scope, command); persistErr != nil {
				return persistErr
			}
			return setResult(CalendarInboundReplyObservedNoChange, &authorization.Current.EventID)
		}
		if prior.Version == ^uint(0) {
			return ErrEventInvalid
		}
		occurredAt := service.clock.Now().UTC()
		if !validCalendarInboundReplyUTC(occurredAt) {
			return ErrEventInvalid
		}
		applicationEventID := service.ids.NewEventID()
		if applicationEventID == uuid.Nil || applicationEventID == authorization.Current.EventID || applicationEventID == prior.AttendeeID || authorization.Exception != nil && applicationEventID == authorization.Exception.ExceptionID {
			return ErrEventInvalid
		}
		updated := cloneCalendarInboundReplyAttendeeSnapshot(prior)
		previousStatus := updated.Status
		updated.Status = normalized.Attendees[0].Status
		updated.Version++
		updated.RespondedAt = &occurredAt
		var recurrenceID *string
		if updated.RecurrenceID != nil {
			value := *updated.RecurrenceID
			recurrenceID = &value
		}
		envelope, envelopeErr := events.NewEnvelopeV1(
			"calendar.rsvp.changed.v1", applicationEventID, occurredAt, scope.OrganizationID(), nil, authorization.Current.EventID.String(),
			calendarEventSeriesRSVPChangedPayload{
				EventID: authorization.Current.EventID, AttendeeID: updated.AttendeeID, AttendeeEmail: updated.Email,
				PreviousStatus: previousStatus, NewStatus: updated.Status, Version: updated.Version,
				RecurrenceID: recurrenceID, ReplyDeliveryStatus: "not_required",
			},
		)
		if envelopeErr != nil {
			return ErrEventInvalid
		}
		descriptor, descriptorErr := NewEventOutboxDescriptor(envelope)
		if descriptorErr != nil {
			return ErrEventInvalid
		}
		command, commandErr := NewCalendarInboundReplyPersistenceCommand(prior, &updated, effectiveSequence, receipt, CalendarInboundReplyReceiptApplied, &descriptor)
		if commandErr != nil {
			return ErrEventInvalid
		}
		if _, persistErr := transaction.PersistCalendarInboundReply(ctx, scope, command); persistErr != nil {
			return persistErr
		}
		return setResult(CalendarInboundReplyApplied, &authorization.Current.EventID)
	}()
	if err != nil {
		return CalendarInboundReplyResult{}, calendarInboundReplySafeError(err)
	}
	if !result.valid() {
		return CalendarInboundReplyResult{}, ErrEventInvalid
	}
	return result, nil
}

func calendarInboundReplyPreflight(rawSender string, message ITIPMessage) (identity ITIPIdentity, normalized ITIPMessage, sender string, accepted bool) {
	defer func() {
		if recover() != nil {
			identity, normalized, sender, accepted = ITIPIdentity{}, ITIPMessage{}, "", false
		}
	}()
	if len(rawSender) > ical.MaxITIPEmailBytes {
		return ITIPIdentity{}, ITIPMessage{}, "", false
	}
	normalizedSender, err := ical.NormalizeITIPEmail(rawSender)
	if err != nil {
		return ITIPIdentity{}, ITIPMessage{}, "", false
	}
	identity, err = CanonicalITIPIdentity(message)
	if err != nil {
		return ITIPIdentity{}, ITIPMessage{}, "", false
	}
	normalized, err = ical.NormalizeITIPMessage(message)
	if err != nil || identity.CanonicalVersion != calendarInboundReplyCanonicalVersion || identity.Method != ITIPMethodReply || len(normalized.Attendees) != 1 || !validCalendarInboundReplyTargetStatus(normalized.Attendees[0].Status) || normalized.Attendees[0].Email != normalizedSender {
		return ITIPIdentity{}, ITIPMessage{}, "", false
	}
	return identity, normalized, normalizedSender, true
}

func calendarInboundReplyAuthorize(scope CalendarInboundReplyScope, authorization CalendarInboundReplyAuthorization, parsed ITIPMessage, sender string) (uint, CalendarInboundReplyOutcome, error) {
	current := authorization.Current
	target := authorization.Target
	if _, err := NewCalendarEventCommand(current); err != nil || !validCalendarEventUpdateAuthoritativeCurrent(current) {
		return 0, "", ErrEventInvalid
	}
	if current.Source != CalendarEventSourceHub || current.Status != CalendarEventStatusConfirmed && current.Status != CalendarEventStatusTentative || current.OrganizerID == nil || *current.OrganizerID != scope.OwnerID() {
		return 0, CalendarInboundReplyQuarantined, nil
	}
	organizer, organizerErr := ical.NormalizeITIPEmail(parsed.Organizer.Email)
	if organizerErr != nil || organizer != current.OrganizerEmail {
		return 0, CalendarInboundReplyQuarantined, nil
	}
	if target.RemovedAt != nil {
		return 0, CalendarInboundReplyQuarantined, nil
	}
	if !validCalendarInboundReplyAttendee(target) {
		return 0, "", ErrEventInvalid
	}
	if target.Email != sender || target.EventID != current.EventID {
		return 0, CalendarInboundReplyQuarantined, nil
	}
	if target.RecurrenceID == nil {
		if authorization.Exception != nil {
			return 0, CalendarInboundReplyQuarantined, nil
		}
		return current.Sequence, "", nil
	}
	recurrenceID := *target.RecurrenceID
	if current.RRULE == "" || !calendarEventOccurrenceCancelMember(current, recurrenceID) || containsCalendarEventOccurrenceCancelExdate(current.Exdates, recurrenceID) {
		return 0, CalendarInboundReplyQuarantined, nil
	}
	if authorization.Exception == nil {
		return current.Sequence, "", nil
	}
	exception := *authorization.Exception
	if exception.ExceptionID == uuid.Nil || exception.EventID != current.EventID || exception.RecurrenceID != recurrenceID || exception.Sequence > current.Sequence {
		return 0, "", ErrEventInvalid
	}
	if exception.Cancelled {
		if exception.Status != CalendarEventStatusCancelled {
			return 0, "", ErrEventInvalid
		}
		return 0, CalendarInboundReplyQuarantined, nil
	}
	if exception.Status != CalendarEventStatusConfirmed && exception.Status != CalendarEventStatusTentative || exception.Temporal != nil && exception.Temporal.Validate() != nil {
		return 0, "", ErrEventInvalid
	}
	return exception.Sequence, "", nil
}

func validCalendarInboundReplyTargetStatus(status RSVPStatus) bool {
	return status == RSVPAccepted || status == RSVPDeclined || status == RSVPTentative
}

func nilCalendarInboundReplyDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func calendarInboundReplySafeError(error) error { return ErrEventInvalid }
