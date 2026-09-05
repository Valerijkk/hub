package mail

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/ical"
)

// MailboxCalendarCorrelationRequest identifies one mailbox message inside the
// actor's authenticated organization and mailbox-account scope.
type MailboxCalendarCorrelationRequest struct {
	actorID        uuid.UUID
	organizationID uuid.UUID
	accountID      uuid.UUID
	messageID      uuid.UUID
}

// NewMailboxCalendarCorrelationRequest creates a validated immutable resolver
// request.
func NewMailboxCalendarCorrelationRequest(
	actorID uuid.UUID,
	organizationID uuid.UUID,
	accountID uuid.UUID,
	messageID uuid.UUID,
) (MailboxCalendarCorrelationRequest, error) {
	request := MailboxCalendarCorrelationRequest{
		actorID:        actorID,
		organizationID: organizationID,
		accountID:      accountID,
		messageID:      messageID,
	}
	if !request.valid() {
		return MailboxCalendarCorrelationRequest{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	return request, nil
}

// ActorID returns the authenticated actor scope.
func (request MailboxCalendarCorrelationRequest) ActorID() uuid.UUID {
	return request.actorID
}

// OrganizationID returns the organization scope.
func (request MailboxCalendarCorrelationRequest) OrganizationID() uuid.UUID {
	return request.organizationID
}

// AccountID returns the mailbox account scope.
func (request MailboxCalendarCorrelationRequest) AccountID() uuid.UUID {
	return request.accountID
}

// MessageID returns the mailbox message scope.
func (request MailboxCalendarCorrelationRequest) MessageID() uuid.UUID {
	return request.messageID
}

func (request MailboxCalendarCorrelationRequest) valid() bool {
	return request.actorID != uuid.Nil &&
		request.organizationID != uuid.Nil &&
		request.accountID != uuid.Nil &&
		request.messageID != uuid.Nil
}

// MailboxCalendarCorrelationPath identifies the authoritative projection path
// that linked a mailbox message to a calendar event.
type MailboxCalendarCorrelationPath string

const (
	MailboxCalendarCorrelationPathInboundReceipt MailboxCalendarCorrelationPath = "inbound_receipt"
	MailboxCalendarCorrelationPathOutboundInvite MailboxCalendarCorrelationPath = "outbound_invite"
)

func (path MailboxCalendarCorrelationPath) valid() bool {
	switch path {
	case MailboxCalendarCorrelationPathInboundReceipt, MailboxCalendarCorrelationPathOutboundInvite:
		return true
	default:
		return false
	}
}

// MailboxCalendarRSVPTargetScope identifies the sole Calendar RSVP endpoint
// that an actionable inbound receipt may use.
type MailboxCalendarRSVPTargetScope string

const (
	MailboxCalendarRSVPTargetScopeSeries     MailboxCalendarRSVPTargetScope = "series"
	MailboxCalendarRSVPTargetScopeOccurrence MailboxCalendarRSVPTargetScope = "occurrence"
)

func (scope MailboxCalendarRSVPTargetScope) valid() bool {
	switch scope {
	case MailboxCalendarRSVPTargetScopeSeries, MailboxCalendarRSVPTargetScopeOccurrence:
		return true
	default:
		return false
	}
}

// MailboxCalendarRSVPStatus is the closed effective attendee RSVP state that
// Calendar will compare-and-swap for an actionable inbound Mail target.
type MailboxCalendarRSVPStatus string

const (
	MailboxCalendarRSVPStatusNeedsAction MailboxCalendarRSVPStatus = "needs_action"
	MailboxCalendarRSVPStatusAccepted    MailboxCalendarRSVPStatus = "accepted"
	MailboxCalendarRSVPStatusTentative   MailboxCalendarRSVPStatus = "tentative"
	MailboxCalendarRSVPStatusDeclined    MailboxCalendarRSVPStatus = "declined"
)

func (status MailboxCalendarRSVPStatus) valid() bool {
	switch status {
	case MailboxCalendarRSVPStatusNeedsAction,
		MailboxCalendarRSVPStatusAccepted,
		MailboxCalendarRSVPStatusTentative,
		MailboxCalendarRSVPStatusDeclined:
		return true
	default:
		return false
	}
}

// MailboxCalendarRSVPTarget is an immutable, inbound-only Calendar RSVP
// authority. It is deliberately distinct from event navigation.
type MailboxCalendarRSVPTarget struct {
	calendarEventID uuid.UUID
	scope           MailboxCalendarRSVPTargetScope
	recurrenceID    *string
	attendeeVersion uint
	rsvpStatus      MailboxCalendarRSVPStatus
}

// NewMailboxCalendarRSVPTarget creates a validated target from a durable,
// normalized Calendar receipt snapshot. The recurrence pointer is copied.
func NewMailboxCalendarRSVPTarget(
	calendarEventID uuid.UUID,
	scope MailboxCalendarRSVPTargetScope,
	recurrenceID *string,
	attendeeVersion uint,
	rsvpStatus MailboxCalendarRSVPStatus,
) (MailboxCalendarRSVPTarget, error) {
	target := MailboxCalendarRSVPTarget{
		calendarEventID: calendarEventID,
		scope:           scope,
		recurrenceID:    cloneMailboxCalendarRSVPTargetRecurrenceID(recurrenceID),
		attendeeVersion: attendeeVersion,
		rsvpStatus:      rsvpStatus,
	}
	if !target.valid() {
		return MailboxCalendarRSVPTarget{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	return target, nil
}

// CalendarEventID returns the target's master Calendar event identifier.
func (target MailboxCalendarRSVPTarget) CalendarEventID() uuid.UUID {
	return target.calendarEventID
}

// Scope returns the target's authoritative RSVP scope.
func (target MailboxCalendarRSVPTarget) Scope() MailboxCalendarRSVPTargetScope {
	return target.scope
}

// RecurrenceID returns a defensive copy of the canonical occurrence identity,
// or nil for a series target.
func (target MailboxCalendarRSVPTarget) RecurrenceID() *string {
	return cloneMailboxCalendarRSVPTargetRecurrenceID(target.recurrenceID)
}

// AttendeeVersion returns Calendar's positive effective attendee CAS version.
func (target MailboxCalendarRSVPTarget) AttendeeVersion() uint {
	return target.attendeeVersion
}

// RSVPStatus returns Calendar's closed effective attendee RSVP state.
func (target MailboxCalendarRSVPTarget) RSVPStatus() MailboxCalendarRSVPStatus {
	return target.rsvpStatus
}

func (target MailboxCalendarRSVPTarget) valid() bool {
	if target.calendarEventID == uuid.Nil || !target.scope.valid() || target.attendeeVersion == 0 || !target.rsvpStatus.valid() {
		return false
	}
	switch target.scope {
	case MailboxCalendarRSVPTargetScopeSeries:
		return target.recurrenceID == nil
	case MailboxCalendarRSVPTargetScopeOccurrence:
		return target.recurrenceID != nil && validMailboxCalendarRSVPTargetRecurrenceID(*target.recurrenceID)
	default:
		return false
	}
}

func cloneMailboxCalendarRSVPTarget(target *MailboxCalendarRSVPTarget) *MailboxCalendarRSVPTarget {
	if target == nil || !target.valid() {
		return nil
	}
	result := *target
	result.recurrenceID = cloneMailboxCalendarRSVPTargetRecurrenceID(target.recurrenceID)
	return &result
}

func cloneMailboxCalendarRSVPTargetRecurrenceID(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := strings.Clone(*value)
	return &cloned
}

// validMailboxCalendarRSVPTargetRecurrenceID mirrors Calendar's canonical
// inbound recurrence contract. Calendar keeps that validator package-private,
// so Mail validates only its persisted, normalized output without parsing text.
func validMailboxCalendarRSVPTargetRecurrenceID(value string) bool {
	if value == "" || len(value) > len(time.RFC3339) || !utf8.ValidString(value) || ical.HasITIPControl(value) || ical.HasLeadingOrTrailingASCIIWhitespace(value) {
		return false
	}
	if len(value) == len(time.DateOnly) {
		parsed, err := time.Parse(time.DateOnly, value)
		return err == nil && parsed.Format(time.DateOnly) == value
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return err == nil && parsed.UTC().Format(time.RFC3339) == value
}

// MailboxCalendarCorrelationResult is a validated, immutable association
// between a fully scoped mailbox message and a calendar event.
type MailboxCalendarCorrelationResult struct {
	actorID            uuid.UUID
	organizationID     uuid.UUID
	accountID          uuid.UUID
	messageID          uuid.UUID
	calendarEventID    uuid.UUID
	path               MailboxCalendarCorrelationPath
	calendarRSVPTarget *MailboxCalendarRSVPTarget
}

// NewMailboxCalendarCorrelationResult creates a result bound to request's
// exact actor, organization, mailbox account, and message scopes.
func NewMailboxCalendarCorrelationResult(
	request MailboxCalendarCorrelationRequest,
	calendarEventID uuid.UUID,
	path MailboxCalendarCorrelationPath,
) (MailboxCalendarCorrelationResult, error) {
	return NewMailboxCalendarCorrelationResultWithRSVPTarget(request, calendarEventID, path, nil)
}

// NewMailboxCalendarCorrelationResultWithRSVPTarget creates a result bound to
// request's exact scope and, optionally, its sole actionable inbound RSVP
// target. Navigation-only and outbound results must not carry a target.
func NewMailboxCalendarCorrelationResultWithRSVPTarget(
	request MailboxCalendarCorrelationRequest,
	calendarEventID uuid.UUID,
	path MailboxCalendarCorrelationPath,
	calendarRSVPTarget *MailboxCalendarRSVPTarget,
) (MailboxCalendarCorrelationResult, error) {
	if calendarRSVPTarget != nil && !calendarRSVPTarget.valid() {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	result := MailboxCalendarCorrelationResult{
		actorID:            request.actorID,
		organizationID:     request.organizationID,
		accountID:          request.accountID,
		messageID:          request.messageID,
		calendarEventID:    calendarEventID,
		path:               path,
		calendarRSVPTarget: cloneMailboxCalendarRSVPTarget(calendarRSVPTarget),
	}
	if !request.valid() || !result.valid() {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	return result, nil
}

// ActorID returns the authenticated actor scope.
func (result MailboxCalendarCorrelationResult) ActorID() uuid.UUID {
	return result.actorID
}

// OrganizationID returns the organization scope.
func (result MailboxCalendarCorrelationResult) OrganizationID() uuid.UUID {
	return result.organizationID
}

// AccountID returns the mailbox account scope.
func (result MailboxCalendarCorrelationResult) AccountID() uuid.UUID {
	return result.accountID
}

// MessageID returns the mailbox message scope.
func (result MailboxCalendarCorrelationResult) MessageID() uuid.UUID {
	return result.messageID
}

// CalendarEventID returns the resolved Hub calendar event identifier.
func (result MailboxCalendarCorrelationResult) CalendarEventID() uuid.UUID {
	return result.calendarEventID
}

// Path returns the authoritative projection path that produced the result.
func (result MailboxCalendarCorrelationResult) Path() MailboxCalendarCorrelationPath {
	return result.path
}

// CalendarRSVPTarget returns a defensive copy of the optional actionable
// inbound RSVP target. A nil value means navigation is non-actionable.
func (result MailboxCalendarCorrelationResult) CalendarRSVPTarget() *MailboxCalendarRSVPTarget {
	return cloneMailboxCalendarRSVPTarget(result.calendarRSVPTarget)
}

// Matches reports whether result is valid and was resolved for the exact
// validated mailbox scope in request.
func (result MailboxCalendarCorrelationResult) Matches(request MailboxCalendarCorrelationRequest) bool {
	return result.valid() && request.valid() &&
		result.actorID == request.actorID &&
		result.organizationID == request.organizationID &&
		result.accountID == request.accountID &&
		result.messageID == request.messageID
}

func (result MailboxCalendarCorrelationResult) valid() bool {
	return result.actorID != uuid.Nil &&
		result.organizationID != uuid.Nil &&
		result.accountID != uuid.Nil &&
		result.messageID != uuid.Nil &&
		result.calendarEventID != uuid.Nil &&
		result.path.valid() &&
		(result.calendarRSVPTarget == nil ||
			(result.path == MailboxCalendarCorrelationPathInboundReceipt &&
				result.calendarRSVPTarget.valid() &&
				result.calendarRSVPTarget.calendarEventID == result.calendarEventID))
}

// MailboxCalendarCorrelationResolver resolves the authoritative Hub calendar
// event for one scoped mailbox message. Consumers MUST call
// ResolveMailboxCalendarCorrelation rather than invoking implementations
// directly. Concrete adapters MUST use only an exact, fully fenced inbound
// CalendarInboundMessageReceipt or an exact, fully fenced outbound
// CalendarInviteDelivery followed by a sent MailSubmission and its result
// MailMessage. They own tenant, ownership, and live-deletion query fences.
// Conflicting event IDs are ambiguous; nonterminal or reconciling rows are
// stale. Implementations MUST NOT use provider IDs, Internet Message-ID,
// folder heuristics, or MailMessage.calendar_event_id as correlation evidence.
type MailboxCalendarCorrelationResolver interface {
	ResolveCalendarEventForMessage(context.Context, MailboxCalendarCorrelationRequest) (MailboxCalendarCorrelationResult, error)
}

// ResolveMailboxCalendarCorrelation is the mandatory safe consumer entry for
// MailboxCalendarCorrelationResolver. It validates the boundary and normalizes
// adapter failures without replacing the concrete adapter's authoritative
// query fences.
func ResolveMailboxCalendarCorrelation(
	ctx context.Context,
	resolver MailboxCalendarCorrelationResolver,
	request MailboxCalendarCorrelationRequest,
) (MailboxCalendarCorrelationResult, error) {
	if ctx == nil || !request.valid() || mailboxCalendarCorrelationResolverNil(resolver) {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	if err := ctx.Err(); err != nil {
		return MailboxCalendarCorrelationResult{}, err
	}

	result, err := resolver.ResolveCalendarEventForMessage(ctx, request)
	if contextErr := ctx.Err(); contextErr != nil {
		return MailboxCalendarCorrelationResult{}, contextErr
	}
	if err != nil {
		if kind, ok := MailboxCalendarCorrelationErrorKindOf(err); ok {
			return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(kind)
		}
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if !result.valid() || !result.Matches(request) {
		return MailboxCalendarCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return result, nil
}

// CalendarMailboxCorrelationRequest identifies one visible calendar event
// inside the actor's authenticated organization scope.
type CalendarMailboxCorrelationRequest struct {
	actorID         uuid.UUID
	organizationID  uuid.UUID
	calendarEventID uuid.UUID
}

// NewCalendarMailboxCorrelationRequest creates a validated immutable reverse
// resolver request.
func NewCalendarMailboxCorrelationRequest(
	actorID uuid.UUID,
	organizationID uuid.UUID,
	calendarEventID uuid.UUID,
) (CalendarMailboxCorrelationRequest, error) {
	request := CalendarMailboxCorrelationRequest{
		actorID:         actorID,
		organizationID:  organizationID,
		calendarEventID: calendarEventID,
	}
	if !request.valid() {
		return CalendarMailboxCorrelationRequest{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	return request, nil
}

// ActorID returns the authenticated actor scope.
func (request CalendarMailboxCorrelationRequest) ActorID() uuid.UUID {
	return request.actorID
}

// OrganizationID returns the organization scope.
func (request CalendarMailboxCorrelationRequest) OrganizationID() uuid.UUID {
	return request.organizationID
}

// CalendarEventID returns the calendar event scope.
func (request CalendarMailboxCorrelationRequest) CalendarEventID() uuid.UUID {
	return request.calendarEventID
}

func (request CalendarMailboxCorrelationRequest) valid() bool {
	return request.actorID != uuid.Nil &&
		request.organizationID != uuid.Nil &&
		request.calendarEventID != uuid.Nil
}

// CalendarMailboxCorrelationResult is a validated, immutable association
// between a fully scoped calendar event and one live mailbox message.
type CalendarMailboxCorrelationResult struct {
	actorID         uuid.UUID
	organizationID  uuid.UUID
	calendarEventID uuid.UUID
	accountID       uuid.UUID
	messageID       uuid.UUID
	path            MailboxCalendarCorrelationPath
}

// NewCalendarMailboxCorrelationResult creates a result bound to request's
// exact actor, organization, and calendar event scopes.
func NewCalendarMailboxCorrelationResult(
	request CalendarMailboxCorrelationRequest,
	accountID uuid.UUID,
	messageID uuid.UUID,
	path MailboxCalendarCorrelationPath,
) (CalendarMailboxCorrelationResult, error) {
	result := CalendarMailboxCorrelationResult{
		actorID:         request.actorID,
		organizationID:  request.organizationID,
		calendarEventID: request.calendarEventID,
		accountID:       accountID,
		messageID:       messageID,
		path:            path,
	}
	if !request.valid() || !result.valid() {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	return result, nil
}

// ActorID returns the authenticated actor scope.
func (result CalendarMailboxCorrelationResult) ActorID() uuid.UUID {
	return result.actorID
}

// OrganizationID returns the organization scope.
func (result CalendarMailboxCorrelationResult) OrganizationID() uuid.UUID {
	return result.organizationID
}

// CalendarEventID returns the resolved calendar event identifier.
func (result CalendarMailboxCorrelationResult) CalendarEventID() uuid.UUID {
	return result.calendarEventID
}

// AccountID returns the owner mailbox account identifier.
func (result CalendarMailboxCorrelationResult) AccountID() uuid.UUID {
	return result.accountID
}

// MessageID returns the live mailbox message identifier.
func (result CalendarMailboxCorrelationResult) MessageID() uuid.UUID {
	return result.messageID
}

// Path returns the authoritative projection path that produced the result.
func (result CalendarMailboxCorrelationResult) Path() MailboxCalendarCorrelationPath {
	return result.path
}

// Matches reports whether result is valid and was resolved for the exact
// validated calendar event scope in request.
func (result CalendarMailboxCorrelationResult) Matches(request CalendarMailboxCorrelationRequest) bool {
	return result.valid() && request.valid() &&
		result.actorID == request.actorID &&
		result.organizationID == request.organizationID &&
		result.calendarEventID == request.calendarEventID
}

func (result CalendarMailboxCorrelationResult) valid() bool {
	return result.actorID != uuid.Nil &&
		result.organizationID != uuid.Nil &&
		result.calendarEventID != uuid.Nil &&
		result.accountID != uuid.Nil &&
		result.messageID != uuid.Nil &&
		result.path.valid()
}

// CalendarMailboxCorrelationResolver resolves the authoritative live mailbox
// message for one scoped calendar event. Consumers MUST call
// ResolveCalendarMailboxCorrelation rather than invoking implementations
// directly. Concrete reverse adapters MAY use only exact terminal inbound
// CalendarInboundMessageReceipt rows or exact sent CalendarInviteDelivery rows
// followed by sent MailSubmission result messages. They must find the actor's
// exact active, ready, enabled mailbox and recheck calendar event visibility.
// Implementations MUST NOT use provider IDs, Internet Message-ID, folders, or
// raw MailMessage.calendar_event_id as correlation evidence. Conflicting
// messages are ambiguous; nonterminal rows are stale.
type CalendarMailboxCorrelationResolver interface {
	ResolveMailMessageForCalendarEvent(context.Context, CalendarMailboxCorrelationRequest) (CalendarMailboxCorrelationResult, error)
}

// ResolveCalendarMailboxCorrelation is the mandatory safe consumer entry for
// CalendarMailboxCorrelationResolver. It validates the boundary and
// normalizes adapter failures without replacing authoritative adapter fences.
func ResolveCalendarMailboxCorrelation(
	ctx context.Context,
	resolver CalendarMailboxCorrelationResolver,
	request CalendarMailboxCorrelationRequest,
) (CalendarMailboxCorrelationResult, error) {
	if ctx == nil || !request.valid() || calendarMailboxCorrelationResolverNil(resolver) {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindInvalid)
	}
	if err := ctx.Err(); err != nil {
		return CalendarMailboxCorrelationResult{}, err
	}

	result, err := resolver.ResolveMailMessageForCalendarEvent(ctx, request)
	if contextErr := ctx.Err(); contextErr != nil {
		return CalendarMailboxCorrelationResult{}, contextErr
	}
	if err != nil {
		if kind, ok := MailboxCalendarCorrelationErrorKindOf(err); ok {
			return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(kind)
		}
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	if !result.valid() || !result.Matches(request) {
		return CalendarMailboxCorrelationResult{}, newMailboxCalendarCorrelationError(MailboxCalendarCorrelationErrorKindUnavailable)
	}
	return result, nil
}

// MailboxCalendarCorrelationErrorKind classifies safe resolver outcomes.
type MailboxCalendarCorrelationErrorKind string

const (
	MailboxCalendarCorrelationErrorKindInvalid     MailboxCalendarCorrelationErrorKind = "invalid"
	MailboxCalendarCorrelationErrorKindNotFound    MailboxCalendarCorrelationErrorKind = "not_found"
	MailboxCalendarCorrelationErrorKindAmbiguous   MailboxCalendarCorrelationErrorKind = "ambiguous"
	MailboxCalendarCorrelationErrorKindStale       MailboxCalendarCorrelationErrorKind = "stale"
	MailboxCalendarCorrelationErrorKindUnavailable MailboxCalendarCorrelationErrorKind = "unavailable"
)

func (kind MailboxCalendarCorrelationErrorKind) valid() bool {
	switch kind {
	case MailboxCalendarCorrelationErrorKindInvalid,
		MailboxCalendarCorrelationErrorKindNotFound,
		MailboxCalendarCorrelationErrorKindAmbiguous,
		MailboxCalendarCorrelationErrorKindStale,
		MailboxCalendarCorrelationErrorKindUnavailable:
		return true
	default:
		return false
	}
}

// MailboxCalendarCorrelationError is a content-free resolver error.
type MailboxCalendarCorrelationError struct {
	kind MailboxCalendarCorrelationErrorKind
}

func (errorValue *MailboxCalendarCorrelationError) Error() string {
	return "mailbox calendar correlation failed"
}

// Kind returns a closed safe classification. Nil, malformed, or unknown values
// fail closed as invalid.
func (errorValue *MailboxCalendarCorrelationError) Kind() MailboxCalendarCorrelationErrorKind {
	if errorValue == nil || !errorValue.kind.valid() {
		return MailboxCalendarCorrelationErrorKindInvalid
	}
	return errorValue.kind
}

// MailboxCalendarCorrelationErrorKindOf extracts a safe resolver error kind
// and whether err is a recognized resolver error. Unknown errors fail closed
// as invalid and are not recognized; malformed recognized values remain
// recognized but their Kind fails closed as invalid.
func MailboxCalendarCorrelationErrorKindOf(err error) (MailboxCalendarCorrelationErrorKind, bool) {
	var correlationError *MailboxCalendarCorrelationError
	if !errors.As(err, &correlationError) {
		return MailboxCalendarCorrelationErrorKindInvalid, false
	}
	return correlationError.Kind(), true
}

func mailboxCalendarCorrelationResolverNil(resolver MailboxCalendarCorrelationResolver) bool {
	if resolver == nil {
		return true
	}

	value := reflect.ValueOf(resolver)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func calendarMailboxCorrelationResolverNil(resolver CalendarMailboxCorrelationResolver) bool {
	if resolver == nil {
		return true
	}

	value := reflect.ValueOf(resolver)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func newMailboxCalendarCorrelationError(kind MailboxCalendarCorrelationErrorKind) error {
	if !kind.valid() {
		kind = MailboxCalendarCorrelationErrorKindInvalid
	}
	return &MailboxCalendarCorrelationError{kind: kind}
}

// NewMailboxCalendarCorrelationError returns a closed correlation error for external resolvers.
func NewMailboxCalendarCorrelationError(kind MailboxCalendarCorrelationErrorKind) error {
	return newMailboxCalendarCorrelationError(kind)
}
