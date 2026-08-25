package calendar

import (
	"crypto/sha256"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/ical"
)

const calendarInboundMaximumSequence uint32 = uint32(ical.MaxITIPSequence)

// CalendarInboundAuthenticationClass is the already-resolved authentication
// admission class. It intentionally carries no sender evidence or MIME data.
type CalendarInboundAuthenticationClass string

const (
	CalendarInboundAuthenticationTrustedInternal CalendarInboundAuthenticationClass = "trusted_internal"
	CalendarInboundAuthenticationDMARCAligned     CalendarInboundAuthenticationClass = "dmarc_aligned"
	CalendarInboundAuthenticationUntrusted        CalendarInboundAuthenticationClass = "untrusted"
)

// CalendarInboundObjectKey is the exact ordering identity for one series or
// one occurrence. A nil recurrence ID denotes the series component.
type CalendarInboundObjectKey struct {
	ownerID      uuid.UUID
	uid          string
	recurrenceID *string
}

// NewCalendarInboundObjectKey snapshots one already-canonical ordering key.
// It performs no normalization so a caller cannot accidentally change the
// parser identity while constructing a decision input.
func NewCalendarInboundObjectKey(ownerID uuid.UUID, uid string, recurrenceID *string) (CalendarInboundObjectKey, error) {
	key := CalendarInboundObjectKey{ownerID: ownerID, uid: uid}
	if recurrenceID != nil {
		value := *recurrenceID
		key.recurrenceID = &value
	}
	if !key.valid() {
		return CalendarInboundObjectKey{}, ErrEventInvalid
	}
	return key, nil
}

func (key CalendarInboundObjectKey) OwnerID() uuid.UUID { return key.ownerID }
func (key CalendarInboundObjectKey) UID() string         { return key.uid }

// RecurrenceID returns an independent optional occurrence key.
func (key CalendarInboundObjectKey) RecurrenceID() *string {
	if key.recurrenceID == nil {
		return nil
	}
	value := *key.recurrenceID
	return &value
}

func (key CalendarInboundObjectKey) IsSeries() bool { return key.recurrenceID == nil }

func (key CalendarInboundObjectKey) valid() bool {
	if key.ownerID == uuid.Nil || ical.ValidateITIPUID(key.uid) != nil {
		return false
	}
	return key.recurrenceID == nil || validCalendarInboundRecurrenceID(*key.recurrenceID)
}

func (key CalendarInboundObjectKey) equal(other CalendarInboundObjectKey) bool {
	if key.ownerID != other.ownerID || key.uid != other.uid || (key.recurrenceID == nil) != (other.recurrenceID == nil) {
		return false
	}
	return key.recurrenceID == nil || *key.recurrenceID == *other.recurrenceID
}

func (key CalendarInboundObjectKey) sameSeries(other CalendarInboundObjectKey) bool {
	return key.ownerID == other.ownerID && key.uid == other.uid
}

func (key CalendarInboundObjectKey) clone() CalendarInboundObjectKey {
	clone := key
	clone.recurrenceID = key.RecurrenceID()
	return clone
}

// CalendarInboundCanonicalPayload keeps the parser's versioned canonical
// SHA-256 proof without retaining iTIP bytes.
type CalendarInboundCanonicalPayload struct {
	canonicalVersion uint8
	hash             [sha256.Size]byte
}

// NewCalendarInboundCanonicalPayload snapshots a stored or parser-provided
// canonical proof. A non-v1 stored version remains representable so equal
// sequence comparison can quarantine when this policy lacks its encoder.
func NewCalendarInboundCanonicalPayload(version uint8, hash [sha256.Size]byte) (CalendarInboundCanonicalPayload, error) {
	if version == 0 {
		return CalendarInboundCanonicalPayload{}, ErrEventInvalid
	}
	return CalendarInboundCanonicalPayload{canonicalVersion: version, hash: hash}, nil
}

func (payload CalendarInboundCanonicalPayload) CanonicalVersion() uint8 { return payload.canonicalVersion }
func (payload CalendarInboundCanonicalPayload) Hash() [sha256.Size]byte { return payload.hash }
func (payload CalendarInboundCanonicalPayload) valid() bool             { return payload.canonicalVersion != 0 }

// CalendarInboundITIP is the bounded, already-canonical REQUEST or CANCEL
// input accepted by the pure ordering policy.
type CalendarInboundITIP struct {
	component      CalendarInboundObjectKey
	organizerEmail string
	authentication CalendarInboundAuthenticationClass
	method         ITIPMethod
	sequence       uint32
	payload        CalendarInboundCanonicalPayload
}

// NewCalendarInboundITIP validates the exact parser identity without parsing
// or normalizing anything. Only the currently supported parser version may be
// admitted as a new inbound part.
func NewCalendarInboundITIP(component CalendarInboundObjectKey, organizerEmail string, authentication CalendarInboundAuthenticationClass, method ITIPMethod, sequence uint32, payload CalendarInboundCanonicalPayload) (CalendarInboundITIP, error) {
	incoming := CalendarInboundITIP{
		component:      component.clone(),
		organizerEmail: organizerEmail,
		authentication: authentication,
		method:         method,
		sequence:       sequence,
		payload:        payload,
	}
	if !incoming.valid() {
		return CalendarInboundITIP{}, ErrEventInvalid
	}
	return incoming, nil
}

func (incoming CalendarInboundITIP) Component() CalendarInboundObjectKey { return incoming.component.clone() }
func (incoming CalendarInboundITIP) OrganizerEmail() string               { return incoming.organizerEmail }
func (incoming CalendarInboundITIP) Authentication() CalendarInboundAuthenticationClass {
	return incoming.authentication
}
func (incoming CalendarInboundITIP) Method() ITIPMethod { return incoming.method }
func (incoming CalendarInboundITIP) Sequence() uint32   { return incoming.sequence }
func (incoming CalendarInboundITIP) Payload() CalendarInboundCanonicalPayload {
	return incoming.payload
}

func (incoming CalendarInboundITIP) valid() bool {
	return incoming.component.valid() &&
		validCalendarInboundOrganizer(incoming.organizerEmail) &&
		validCalendarInboundAuthentication(incoming.authentication) &&
		validCalendarInboundMethod(incoming.method) &&
		incoming.sequence <= calendarInboundMaximumSequence &&
		incoming.payload.valid() &&
		incoming.payload.canonicalVersion == ical.CanonicalITIPVersion
}

// CalendarInboundObjectStatus is the durable state classification selected by
// a transaction after an ordering decision.
type CalendarInboundObjectStatus string

const (
	CalendarInboundObjectApplied                CalendarInboundObjectStatus = "applied"
	CalendarInboundObjectTombstone              CalendarInboundObjectStatus = "tombstone"
	CalendarInboundObjectDeferredMissingMaster  CalendarInboundObjectStatus = "deferred_missing_master"
)

// CalendarInboundObjectState is an immutable current ordering snapshot. Its
// optional event ID is copied at every boundary and never exposes message data.
type CalendarInboundObjectState struct {
	component      CalendarInboundObjectKey
	organizerEmail string
	sequence       uint32
	method         ITIPMethod
	payload        CalendarInboundCanonicalPayload
	status         CalendarInboundObjectStatus
	eventID        *uuid.UUID
}

// NewCalendarInboundObjectState snapshots one authoritative ordering row.
func NewCalendarInboundObjectState(component CalendarInboundObjectKey, organizerEmail string, sequence uint32, method ITIPMethod, payload CalendarInboundCanonicalPayload, status CalendarInboundObjectStatus, eventID *uuid.UUID) (CalendarInboundObjectState, error) {
	state := CalendarInboundObjectState{
		component:      component.clone(),
		organizerEmail: organizerEmail,
		sequence:       sequence,
		method:         method,
		payload:        payload,
		status:         status,
	}
	if eventID != nil {
		value := *eventID
		state.eventID = &value
	}
	if !state.valid() {
		return CalendarInboundObjectState{}, ErrEventInvalid
	}
	return state, nil
}

func (state CalendarInboundObjectState) Component() CalendarInboundObjectKey { return state.component.clone() }
func (state CalendarInboundObjectState) OrganizerEmail() string               { return state.organizerEmail }
func (state CalendarInboundObjectState) Sequence() uint32                     { return state.sequence }
func (state CalendarInboundObjectState) Method() ITIPMethod                   { return state.method }
func (state CalendarInboundObjectState) Payload() CalendarInboundCanonicalPayload {
	return state.payload
}
func (state CalendarInboundObjectState) Status() CalendarInboundObjectStatus { return state.status }

// EventID returns an independent optional bound Calendar event identifier.
func (state CalendarInboundObjectState) EventID() *uuid.UUID {
	if state.eventID == nil {
		return nil
	}
	value := *state.eventID
	return &value
}

func (state CalendarInboundObjectState) valid() bool {
	if !state.component.valid() || !validCalendarInboundOrganizer(state.organizerEmail) ||
		!validCalendarInboundMethod(state.method) || state.sequence > calendarInboundMaximumSequence || !state.payload.valid() {
		return false
	}
	if state.eventID != nil && *state.eventID == uuid.Nil {
		return false
	}
	switch state.status {
	case CalendarInboundObjectApplied:
		return state.method == ITIPMethodRequest && state.eventID != nil
	case CalendarInboundObjectTombstone:
		return state.method == ITIPMethodCancel
	case CalendarInboundObjectDeferredMissingMaster:
		return !state.component.IsSeries() && state.method == ITIPMethodRequest && state.eventID == nil
	default:
		return false
	}
}

func (state CalendarInboundObjectState) clone() CalendarInboundObjectState {
	clone := state
	clone.component = state.component.clone()
	clone.eventID = state.EventID()
	return clone
}

// CalendarInboundIngestInput contains one exact component state and, for an
// occurrence, its distinct series-admission state. The policy never compares
// series and occurrence sequence numbers.
type CalendarInboundIngestInput struct {
	incoming CalendarInboundITIP
	current  *CalendarInboundObjectState
	series   *CalendarInboundObjectState
}

// NewCalendarInboundIngestInput snapshots the complete state required for a
// pure decision. Series input is forbidden for a series component and must be
// the matching master component for an occurrence.
func NewCalendarInboundIngestInput(incoming CalendarInboundITIP, current *CalendarInboundObjectState, series *CalendarInboundObjectState) (CalendarInboundIngestInput, error) {
	input := CalendarInboundIngestInput{incoming: incoming}
	if current != nil {
		value := current.clone()
		input.current = &value
	}
	if series != nil {
		value := series.clone()
		input.series = &value
	}
	if !input.valid() {
		return CalendarInboundIngestInput{}, ErrEventInvalid
	}
	return input, nil
}

func (input CalendarInboundIngestInput) valid() bool {
	if !input.incoming.valid() {
		return false
	}
	if input.current != nil && (!input.current.valid() || !input.current.component.equal(input.incoming.component)) {
		return false
	}
	if input.incoming.component.IsSeries() {
		return input.series == nil
	}
	return input.series == nil || input.series.valid() && input.series.component.IsSeries() && input.series.component.sameSeries(input.incoming.component)
}

// CalendarInboundIngestOutcome is a zero-safe request for the transaction
// layer. The empty value is invalid and can never authorize an apply.
type CalendarInboundIngestOutcome string

const (
	CalendarInboundIngestApplyRequest        CalendarInboundIngestOutcome = "apply_request"
	CalendarInboundIngestTombstoneOccurrence CalendarInboundIngestOutcome = "tombstone_occurrence"
	CalendarInboundIngestTombstoneSeries     CalendarInboundIngestOutcome = "tombstone_series"
	CalendarInboundIngestDeferred            CalendarInboundIngestOutcome = "deferred_missing_master"
	CalendarInboundIngestIgnoredStale        CalendarInboundIngestOutcome = "ignored_stale"
	CalendarInboundIngestDuplicate           CalendarInboundIngestOutcome = "duplicate"
	CalendarInboundIngestQuarantined         CalendarInboundIngestOutcome = "quarantined"
)

// CalendarInboundIngestResult is an opaque decision result. A state target is
// present only when the transaction must advance the current ordering record.
type CalendarInboundIngestResult struct {
	outcome   CalendarInboundIngestOutcome
	nextState CalendarInboundObjectStatus
}

func (result CalendarInboundIngestResult) Outcome() CalendarInboundIngestOutcome { return result.outcome }

// NextState returns the durable target selected by a mutating outcome.
func (result CalendarInboundIngestResult) NextState() (CalendarInboundObjectStatus, bool) {
	if result.nextState == "" {
		return "", false
	}
	return result.nextState, true
}

// CancelsWholeSeries identifies the only decision that must tombstone the
// master and cancel every materialized exception atomically.
func (result CalendarInboundIngestResult) CancelsWholeSeries() bool {
	return result.outcome == CalendarInboundIngestTombstoneSeries
}

func (result CalendarInboundIngestResult) valid() bool {
	switch result.outcome {
	case CalendarInboundIngestApplyRequest:
		return result.nextState == CalendarInboundObjectApplied
	case CalendarInboundIngestTombstoneOccurrence, CalendarInboundIngestTombstoneSeries:
		return result.nextState == CalendarInboundObjectTombstone
	case CalendarInboundIngestDeferred:
		return result.nextState == CalendarInboundObjectDeferredMissingMaster
	case CalendarInboundIngestIgnoredStale, CalendarInboundIngestDuplicate, CalendarInboundIngestQuarantined:
		return result.nextState == ""
	default:
		return false
	}
}

// DecideCalendarInboundITIP implements the REQUEST/CANCEL ordering table
// without I/O, clocks, repositories, event mutation, or payload parsing.
func DecideCalendarInboundITIP(input CalendarInboundIngestInput) (CalendarInboundIngestResult, error) {
	if !input.valid() {
		return CalendarInboundIngestResult{}, ErrEventInvalid
	}
	incoming := input.incoming
	if !calendarInboundAuthenticationAdmitted(incoming.authentication) || calendarInboundOrganizerMismatch(input, incoming) {
		return calendarInboundIngestResult(CalendarInboundIngestQuarantined, ""), nil
	}
	if input.current != nil {
		switch {
		case incoming.sequence < input.current.sequence:
			return calendarInboundIngestResult(CalendarInboundIngestIgnoredStale, ""), nil
		case incoming.sequence == input.current.sequence:
			return decideCalendarInboundEqualSequence(input, incoming)
		}
	}
	return decideCalendarInboundAdvance(input, incoming), nil
}

func decideCalendarInboundEqualSequence(input CalendarInboundIngestInput, incoming CalendarInboundITIP) (CalendarInboundIngestResult, error) {
	current := input.current
	if current == nil || current.payload.canonicalVersion != ical.CanonicalITIPVersion || current.payload.canonicalVersion != incoming.payload.canonicalVersion {
		return calendarInboundIngestResult(CalendarInboundIngestQuarantined, ""), nil
	}
	if current.method != incoming.method || current.payload.hash != incoming.payload.hash {
		return calendarInboundIngestResult(CalendarInboundIngestQuarantined, ""), nil
	}
	if current.status == CalendarInboundObjectDeferredMissingMaster && calendarInboundSeriesIsActive(input.series) {
		return calendarInboundIngestResult(CalendarInboundIngestApplyRequest, CalendarInboundObjectApplied), nil
	}
	return calendarInboundIngestResult(CalendarInboundIngestDuplicate, ""), nil
}

func decideCalendarInboundAdvance(input CalendarInboundIngestInput, incoming CalendarInboundITIP) CalendarInboundIngestResult {
	switch incoming.method {
	case ITIPMethodCancel:
		if incoming.component.IsSeries() {
			return calendarInboundIngestResult(CalendarInboundIngestTombstoneSeries, CalendarInboundObjectTombstone)
		}
		return calendarInboundIngestResult(CalendarInboundIngestTombstoneOccurrence, CalendarInboundObjectTombstone)
	case ITIPMethodRequest:
		if incoming.component.IsSeries() || calendarInboundSeriesIsActive(input.series) {
			return calendarInboundIngestResult(CalendarInboundIngestApplyRequest, CalendarInboundObjectApplied)
		}
		return calendarInboundIngestResult(CalendarInboundIngestDeferred, CalendarInboundObjectDeferredMissingMaster)
	default:
		return CalendarInboundIngestResult{}
	}
}

func calendarInboundIngestResult(outcome CalendarInboundIngestOutcome, nextState CalendarInboundObjectStatus) CalendarInboundIngestResult {
	result := CalendarInboundIngestResult{outcome: outcome, nextState: nextState}
	if !result.valid() {
		return CalendarInboundIngestResult{}
	}
	return result
}

func calendarInboundOrganizerMismatch(input CalendarInboundIngestInput, incoming CalendarInboundITIP) bool {
	return input.current != nil && input.current.organizerEmail != incoming.organizerEmail ||
		input.series != nil && input.series.organizerEmail != incoming.organizerEmail
}

func calendarInboundSeriesIsActive(series *CalendarInboundObjectState) bool {
	return series != nil && series.status == CalendarInboundObjectApplied
}

func calendarInboundAuthenticationAdmitted(authentication CalendarInboundAuthenticationClass) bool {
	return authentication == CalendarInboundAuthenticationTrustedInternal || authentication == CalendarInboundAuthenticationDMARCAligned
}

func validCalendarInboundAuthentication(authentication CalendarInboundAuthenticationClass) bool {
	return calendarInboundAuthenticationAdmitted(authentication) || authentication == CalendarInboundAuthenticationUntrusted
}

func validCalendarInboundMethod(method ITIPMethod) bool {
	return method == ITIPMethodRequest || method == ITIPMethodCancel
}

func validCalendarInboundOrganizer(value string) bool {
	if len(value) > ical.MaxITIPEmailBytes || !utf8.ValidString(value) {
		return false
	}
	normalized, err := ical.NormalizeITIPEmail(value)
	return err == nil && normalized == value
}

func validCalendarInboundRecurrenceID(value string) bool {
	if value == "" || len(value) > len(time.RFC3339) || !utf8.ValidString(value) || ical.HasITIPControl(value) || ical.HasLeadingOrTrailingASCIIWhitespace(value) {
		return false
	}
	if len(value) == len("2006-01-02") {
		parsed, err := time.Parse("2006-01-02", value)
		return err == nil && parsed.Format("2006-01-02") == value
	}
	parsed, err := time.Parse(time.RFC3339, value)
	return err == nil && parsed.UTC().Format(time.RFC3339) == value
}
