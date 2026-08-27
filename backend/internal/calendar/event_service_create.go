package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	ical "github.com/valerii/hub/backend/internal/ical"
)

const (
	calendarCreateFingerprintVersion = 1
	calendarCreateMaxAttendees       = 200
	calendarCreateMaxTitleBytes      = 500
	calendarCreateMaxDescription     = 200000
	calendarCreateMaxLocationBytes   = 500
	calendarCreateMaxNameRunes       = 160
	calendarCreateMaxNameBytes       = 640
	calendarCreateMaxTimezoneBytes   = 64
	calendarCreateMaxRRuleBytes      = ical.CalendarRecurrenceMaxRRuleBytes
	calendarCreateMaxExdates         = 1000
	calendarCreateMaxEmailBytes      = 320
	calendarCreateMaxIdempotencyKey  = 128
	calendarCreateMinIdempotencyKey  = 16
	calendarCreateOccurrenceCap      = ical.CalendarRecurrenceOccurrenceCap
	calendarCreateCandidateWorkGuard = ical.CalendarRecurrenceCandidateWorkGuard
)

// CalendarEventCreateInput contains untrusted content and identity candidates.
// The in-transaction resolver is the only organizer/attendee authority.
type CalendarEventCreateInput struct {
	IdempotencyKey  string
	Title           string
	Description     string
	Location        string
	Temporal        CalendarEvent
	RRULE           string
	Exdates         []string
	ScheduledCallID *uuid.UUID
	ReminderMinutes *int
	Attendees       []CalendarEventCreateAttendeeCandidate
}

// CalendarEventCreateCommand is an immutable normalized create request.
type CalendarEventCreateCommand struct {
	input       CalendarEventCreateInput
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

func NewCalendarEventCreateCommand(input CalendarEventCreateInput) (CalendarEventCreateCommand, error) {
	if !validCalendarCreateIdempotencyKey(input.IdempotencyKey) ||
		!validCalendarCreateText(input.Title, calendarCreateMaxTitleBytes, true) ||
		!validCalendarCreateText(input.Description, calendarCreateMaxDescription, false) ||
		!validCalendarCreateText(input.Location, calendarCreateMaxLocationBytes, false) ||
		!utf8.ValidString(input.Temporal.Timezone) ||
		len(input.Temporal.Timezone) > calendarCreateMaxTimezoneBytes ||
		input.Temporal.Validate() != nil ||
		len(input.Attendees) > calendarCreateMaxAttendees {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	if input.ScheduledCallID != nil && *input.ScheduledCallID == uuid.Nil {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	if input.ReminderMinutes != nil && (*input.ReminderMinutes < 0 || *input.ReminderMinutes > 40320) {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	if input.RRULE != "" {
		if len(input.RRULE) > calendarCreateMaxRRuleBytes || !utf8.ValidString(input.RRULE) {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		normalized, err := normalizeCalendarCreateRRULE(input.RRULE, input.Temporal)
		if err != nil {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		input.RRULE = normalized
	}
	var err error
	input.Exdates, err = canonicalizeCalendarCreateExdates(input.Temporal, input.Exdates)
	if err != nil {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	input.ScheduledCallID = copyUUID(input.ScheduledCallID)
	input.ReminderMinutes = copyInt(input.ReminderMinutes)
	attendees := make([]CalendarEventCreateAttendeeCandidate, len(input.Attendees))
	seenEmails := make(map[string]struct{}, len(input.Attendees))
	seenUsers := make(map[uuid.UUID]struct{}, len(input.Attendees))
	for index, candidate := range input.Attendees {
		if !validCalendarCreateEmail(candidate.Email) || !validCalendarCreateDisplayName(candidate.DisplayName) || !validCalendarEventAttendeeRole(candidate.Role) {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		candidate.Email = normalizeCalendarCreateEmail(candidate.Email)
		if _, exists := seenEmails[candidate.Email]; exists {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		seenEmails[candidate.Email] = struct{}{}
		if candidate.UserID != nil {
			if *candidate.UserID == uuid.Nil {
				return CalendarEventCreateCommand{}, ErrEventInvalid
			}
			if _, exists := seenUsers[*candidate.UserID]; exists {
				return CalendarEventCreateCommand{}, ErrEventInvalid
			}
			seenUsers[*candidate.UserID] = struct{}{}
		}
		attendees[index] = cloneCalendarEventCreateAttendeeCandidate(candidate)
	}
	input.Attendees = attendees
	requestHash, err := calendarCreateRequestHash(input)
	if err != nil {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	return CalendarEventCreateCommand{
		input: input, keyHash: MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))), requestHash: requestHash,
	}, nil
}

func (command CalendarEventCreateCommand) valid() bool {
	rebuilt, err := NewCalendarEventCreateCommand(command.input)
	return err == nil && rebuilt.keyHash == command.keyHash && rebuilt.requestHash == command.requestHash
}

type CalendarEventCreateResult struct {
	response MutationReceiptResponse
	replayed bool
}

func (result CalendarEventCreateResult) Response() MutationReceiptResponse { return result.response }
func (result CalendarEventCreateResult) Replayed() bool { return result.replayed }
func (result CalendarEventCreateResult) valid() bool { return result.response.valid() }

// CreateCalendarEvent orders fresh flow as scope, receipt reservation/replay,
// authorization and writes. Replay intentionally precedes authorization.
func (service *EventService) CreateCalendarEvent(ctx context.Context, actorID uuid.UUID, command CalendarEventCreateCommand) (CalendarEventCreateResult, error) {
	if service == nil || service.repository == nil || service.clock == nil || service.ids == nil || actorID == uuid.Nil || !command.valid() {
		return CalendarEventCreateResult{}, ErrEventInvalid
	}
	var result CalendarEventCreateResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveMutationScope(ctx, actorID, MutationReceiptEventCreate)
		if err != nil {
			return err
		}
		reservation, err := NewMutationReceiptReservation(scope, command.keyHash, command.requestHash)
		if err != nil {
			return err
		}
		reserved, err := transaction.ReserveMutationReceipt(ctx, scope, reservation)
		if err != nil {
			return err
		}
		switch reserved.State() {
		case MutationReceiptReplayed:
			response, found := reserved.ReplayedResponse()
			if !found {
				return ErrEventInvalid
			}
			result = CalendarEventCreateResult{response: response, replayed: true}
			return nil
		case MutationReceiptReserved:
		default:
			return ErrEventInvalid
		}
		authorization, err := transaction.ResolveCalendarEventCreateAuthorization(ctx, scope, calendarCreateAuthorizationRequest(scope, command.input))
		if err != nil {
			return err
		}
		if err := validateCalendarCreateAuthorization(scope, command.input, authorization); err != nil {
			return err
		}
		scopeActorID, scopeOrganizationID := scope.ActorID(), scope.OrganizationID()
		eventID := calendarCreateUUID("hub.calendar.event.id.v1", scopeActorID[:], scopeOrganizationID[:], command.keyHash[:], command.requestHash[:])
		event, organizer, attendees, attendeeIDs, err := calendarCreatePersistenceCommands(scope, eventID, command.input, authorization)
		if err != nil {
			return err
		}
		persistedEvent, err := transaction.PersistCalendarEvent(ctx, scope, event)
		if err != nil {
			return err
		}
		if persistedEvent.EventID() != eventID || persistedEvent.Version() != 1 {
			return ErrEventInvalid
		}
		if err := calendarCreatePersistAttendee(ctx, transaction, scope, organizer); err != nil {
			return err
		}
		for _, attendee := range attendees {
			if err := calendarCreatePersistAttendee(ctx, transaction, scope, attendee); err != nil {
				return err
			}
		}
		sourceEnvelopeID := service.ids.NewEventID()
		if sourceEnvelopeID == uuid.Nil {
			return ErrEventInvalid
		}
		eventSnapshot := event.Snapshot()
		if eventSnapshot.EventID != eventID || eventSnapshot.OrganizerID == nil || *eventSnapshot.OrganizerID == uuid.Nil {
			return ErrEventInvalid
		}
		actor := scope.ActorID()
		envelope, err := NewCalendarInviteDeliveryEnvelopeV2(
			CalendarInviteCreatedSubjectV2, sourceEnvelopeID, service.clock.Now().UTC(), scope.OrganizationID(), &actor, eventID.String(),
			CalendarInviteDeliveryEnvelopeV2Input{
				EventID: eventID, EventSequence: eventSnapshot.Sequence, OrganizerID: *eventSnapshot.OrganizerID,
				RecurrenceID: nil, AttendeeIDs: attendeeIDs,
			},
		)
		if err != nil {
			return ErrEventInvalid
		}
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil {
			return err
		}
		if err := transaction.AppendOutbox(ctx, scope, descriptor); err != nil {
			return err
		}
		response, err := calendarCreateResponse(eventID, command.input, authorization, attendees, attendeeIDs)
		if err != nil {
			return err
		}
		completion, err := NewMutationReceiptCompletion(reserved.ReceiptID(), &eventID, response)
		if err != nil {
			return err
		}
		if err := transaction.CompleteMutationReceipt(ctx, scope, completion); err != nil {
			return err
		}
		result = CalendarEventCreateResult{response: response}
		return nil
	})
	if err != nil {
		return CalendarEventCreateResult{}, calendarCreateSafeError(err)
	}
	if !result.valid() {
		return CalendarEventCreateResult{}, ErrEventInvalid
	}
	return result, nil
}

func calendarCreateAuthorizationRequest(scope EventMutationScope, input CalendarEventCreateInput) CalendarEventCreateAuthorizationRequest {
	organizerID := scope.ActorID()
	attendees := make([]CalendarEventCreateAttendeeCandidate, len(input.Attendees))
	for index, candidate := range input.Attendees {
		attendees[index] = cloneCalendarEventCreateAttendeeCandidate(candidate)
	}
	return CalendarEventCreateAuthorizationRequest{
		Organizer: CalendarEventCreateAttendeeCandidate{UserID: &organizerID, Role: CalendarEventAttendeeRoleRequired}, Attendees: attendees,
		Title: input.Title, Description: input.Description, Location: input.Location, Temporal: input.Temporal, RRULE: input.RRULE,
		Exdates: append([]string(nil), input.Exdates...), ReminderMinutes: copyInt(input.ReminderMinutes), ScheduledCallID: copyUUID(input.ScheduledCallID),
	}
}

func validateCalendarCreateAuthorization(scope EventMutationScope, input CalendarEventCreateInput, authorization CalendarEventCreateAuthorization) error {
	if !sameCalendarCreateUUID(input.ScheduledCallID, authorization.ScheduledCallID) {
		return newCalendarEventCreateConflict(CalendarEventCreateConflictScheduledCallLinkMismatch)
	}
	if authorization.Organizer.UserID != scope.ActorID() || !validCalendarCreateEmail(authorization.Organizer.Email) ||
		!validCalendarCreateDisplayName(authorization.Organizer.DisplayName) || len(authorization.Attendees) != len(input.Attendees) {
		return ErrEventConflict
	}
	if authorization.ScheduledCallID != nil {
		for _, resolved := range authorization.Attendees {
			if resolved.UserID == nil || *resolved.UserID == uuid.Nil {
				return newCalendarEventCreateConflict(CalendarEventCreateConflictScheduledCallNotAttachable)
			}
		}
	}
	seenEmails := map[string]struct{}{normalizeCalendarCreateEmail(authorization.Organizer.Email): {}}
	seenUsers := map[uuid.UUID]struct{}{authorization.Organizer.UserID: {}}
	for index, resolved := range authorization.Attendees {
		candidate := input.Attendees[index]
		if !validCalendarCreateEmail(resolved.Email) || !validCalendarCreateDisplayName(resolved.DisplayName) ||
			!validCalendarEventAttendeeRole(resolved.Role) || resolved.Role != candidate.Role ||
			normalizeCalendarCreateEmail(resolved.Email) != candidate.Email {
			return ErrEventConflict
		}
		if candidate.UserID != nil && (resolved.UserID == nil || *resolved.UserID != *candidate.UserID) {
			return ErrEventConflict
		}
		if resolved.UserID != nil {
			if *resolved.UserID == uuid.Nil {
				return ErrEventConflict
			}
			if _, exists := seenUsers[*resolved.UserID]; exists {
				return ErrEventConflict
			}
			seenUsers[*resolved.UserID] = struct{}{}
		}
		email := normalizeCalendarCreateEmail(resolved.Email)
		if _, exists := seenEmails[email]; exists {
			return ErrEventConflict
		}
		seenEmails[email] = struct{}{}
	}
	return nil
}

func sameCalendarCreateUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func calendarCreatePersistAttendee(ctx context.Context, transaction *EventTransaction, scope EventMutationScope, attendee CalendarEventAttendeeCommand) error {
	persisted, err := transaction.PersistCalendarEventAttendee(ctx, scope, attendee)
	if err != nil {
		return err
	}
	if persisted.AttendeeID() != attendee.AttendeeID() || persisted.Version() != 1 {
		return ErrEventInvalid
	}
	return nil
}

func calendarCreatePersistenceCommands(scope EventMutationScope, eventID uuid.UUID, input CalendarEventCreateInput, authorization CalendarEventCreateAuthorization) (CalendarEventCommand, CalendarEventAttendeeCommand, []CalendarEventAttendeeCommand, []uuid.UUID, error) {
	organizerID := authorization.Organizer.UserID
	event, err := NewCalendarEventCommand(CalendarEventCommandInput{
		EventID: eventID, OrganizationID: scope.OrganizationID(), OwnerID: scope.ActorID(), UID: "hub-" + eventID.String(),
		Title: input.Title, Description: input.Description, Location: input.Location, Temporal: input.Temporal, RRULE: input.RRULE, Exdates: input.Exdates,
		Sequence: 0, Version: 1, Source: CalendarEventSourceHub, Status: CalendarEventStatusConfirmed, OrganizerID: &organizerID,
		OrganizerEmail: normalizeCalendarCreateEmail(authorization.Organizer.Email), ScheduledCallID: copyUUID(authorization.ScheduledCallID), ReminderMinutes: copyInt(input.ReminderMinutes),
	})
	if err != nil {
		return CalendarEventCommand{}, CalendarEventAttendeeCommand{}, nil, nil, err
	}
	organizerAttendeeID := calendarCreateUUID("hub.calendar.attendee.id.v1", eventID[:], []byte("organizer"))
	organizer, err := newCalendarEventOrganizerAttendeeCommand(CalendarEventAttendeeCommandInput{
		AttendeeID: organizerAttendeeID, EventID: eventID, UserID: &organizerID, Email: normalizeCalendarCreateEmail(authorization.Organizer.Email),
		DisplayName: calendarCreateDisplayName(authorization.Organizer.DisplayName), Role: CalendarEventAttendeeRoleOrganizer, Status: RSVPAccepted, Version: 1,
	})
	if err != nil {
		return CalendarEventCommand{}, CalendarEventAttendeeCommand{}, nil, nil, err
	}
	attendees := make([]CalendarEventAttendeeCommand, 0, len(authorization.Attendees))
	attendeeIDs := make([]uuid.UUID, 0, len(authorization.Attendees))
	for index, resolved := range authorization.Attendees {
		attendeeID := calendarCreateUUID("hub.calendar.attendee.id.v1", eventID[:], calendarCreateIndexBytes(index))
		attendee, err := NewCalendarEventAttendeeCommand(CalendarEventAttendeeCommandInput{
			AttendeeID: attendeeID, EventID: eventID, UserID: copyUUID(resolved.UserID), Email: normalizeCalendarCreateEmail(resolved.Email),
			DisplayName: calendarCreateDisplayName(resolved.DisplayName), Role: resolved.Role, Status: RSVPNeedsAction, Version: 1,
		})
		if err != nil {
			return CalendarEventCommand{}, CalendarEventAttendeeCommand{}, nil, nil, err
		}
		attendees = append(attendees, attendee)
		attendeeIDs = append(attendeeIDs, attendeeID)
	}
	return event, organizer, attendees, attendeeIDs, nil
}

func calendarCreateDisplayName(value *string) string {
	if value == nil { return "" }
	return *value
}

func calendarCreateResponse(eventID uuid.UUID, input CalendarEventCreateInput, authorization CalendarEventCreateAuthorization, attendees []CalendarEventAttendeeCommand, attendeeIDs []uuid.UUID) (MutationReceiptResponse, error) {
	responseAttendees := make([]calendarCreateResponseAttendee, 0, len(authorization.Attendees))
	for index, resolved := range authorization.Attendees {
		responseAttendees = append(responseAttendees, calendarCreateResponseAttendee{
			ID: attendeeIDs[index], UserID: copyUUID(resolved.UserID), Email: normalizeCalendarCreateEmail(resolved.Email), DisplayName: cloneCalendarEventCreateString(resolved.DisplayName),
			Role: resolved.Role, RSVPStatus: RSVPNeedsAction, ReplyDeliveryStatus: "not_required", Version: attendees[index].Version(),
		})
	}
	startsAt, endsAt, startDate, endDateExclusive := calendarCreateTemporalResponse(input.Temporal)
	organizerID := authorization.Organizer.UserID
	encoded, err := json.Marshal(calendarCreateResponseBody{
		ID: eventID, UID: "hub-" + eventID.String(), Title: input.Title, Description: input.Description, Location: input.Location,
		StartsAt: startsAt, EndsAt: endsAt, StartDate: startDate, EndDateExclusive: endDateExclusive, AllDay: input.Temporal.AllDay,
		Timezone: input.Temporal.Timezone, RRULE: input.RRULE, Exdates: append([]string(nil), input.Exdates...), Sequence: 0,
		Source: CalendarEventSourceHub, Status: CalendarEventStatusConfirmed, OrganizerID: &organizerID, OrganizerEmail: normalizeCalendarCreateEmail(authorization.Organizer.Email),
		ScheduledCallID: copyUUID(authorization.ScheduledCallID), ReminderMinutes: copyInt(input.ReminderMinutes), Attendees: responseAttendees, Version: 1,
	})
	if err != nil { return MutationReceiptResponse{}, ErrEventInvalid }
	return NewMutationReceiptResponse(201, encoded)
}

type calendarCreateResponseBody struct {
	ID uuid.UUID `json:"id"`; UID string `json:"uid"`; Title string `json:"title"`; Description string `json:"description,omitempty"`; Location string `json:"location,omitempty"`
	StartsAt *time.Time `json:"starts_at"`; EndsAt *time.Time `json:"ends_at"`; StartDate *string `json:"start_date"`; EndDateExclusive *string `json:"end_date_exclusive"`
	AllDay bool `json:"all_day"`; Timezone string `json:"timezone"`; RRULE string `json:"rrule,omitempty"`; Exdates []string `json:"exdates,omitempty"`; Sequence uint `json:"sequence"`
	Source CalendarEventSource `json:"source"`; Status CalendarEventStatus `json:"status"`; OrganizerID *uuid.UUID `json:"organizer_id"`; OrganizerEmail string `json:"organizer_email"`
	ScheduledCallID *uuid.UUID `json:"scheduled_call_id"`; ReminderMinutes *int `json:"reminder_minutes"`; Attendees []calendarCreateResponseAttendee `json:"attendees"`; Version uint `json:"version"`
}

type calendarCreateResponseAttendee struct {
	ID uuid.UUID `json:"id"`; UserID *uuid.UUID `json:"user_id"`; Email string `json:"email"`; DisplayName *string `json:"display_name"`; Role CalendarEventAttendeeRole `json:"role"`
	RSVPStatus RSVPStatus `json:"rsvp_status"`; RecurrenceID *string `json:"recurrence_id"`; RespondedAt *time.Time `json:"responded_at"`; Version uint `json:"version"`; ReplyDeliveryStatus string `json:"reply_delivery_status"`
}

func calendarCreateTemporalResponse(temporal CalendarEvent) (*time.Time, *time.Time, *string, *string) {
	if temporal.AllDay { startDate, endDateExclusive := temporal.StartDate, temporal.EndDateExclusive; return nil, nil, &startDate, &endDateExclusive }
	startsAt, endsAt := temporal.StartsAt.UTC(), temporal.EndsAt.UTC()
	return &startsAt, &endsAt, nil, nil
}

// normalizeCalendarCreateRRULE implements the closed section 3.11 create grammar.
// It deliberately does not share the inbound iTIP parser or its repair rules.
func normalizeCalendarCreateRRULE(value string, temporal CalendarEvent) (string, error) {
	normalized, ok := ical.NormalizeCalendarRRULE(temporal, value)
	if !ok {
		return "", ErrEventInvalid
	}
	return normalized, nil
}

func parseCalendarCreatePositiveInteger(value string, maximum int) (int, error) {
	parsed, ok := ical.ParseCalendarRecurrencePositiveInteger(value, maximum)
	if !ok {
		return 0, ErrEventInvalid
	}
	return parsed, nil
}

type calendarCreateCivilTime = ical.CalendarRecurrenceCivilTime

func calendarCreateCivilTimeFromTime(value time.Time) calendarCreateCivilTime {
	return ical.CalendarRecurrenceCivilTimeFromTime(value)
}

func calendarCreateCompareCivilTime(left, right calendarCreateCivilTime) int {
	return ical.CompareCalendarRecurrenceCivilTime(left, right)
}

func calendarCreateResolveCivilTime(location *time.Location, civil calendarCreateCivilTime) (time.Time, bool) {
	return ical.ResolveCalendarRecurrenceCivilTime(location, civil)
}

func calendarCreateWeekdayOffsets(value string) ([]int, error) {
	offsets, ok := ical.CalendarRecurrenceWeekdayOffsets(value)
	if !ok {
		return nil, ErrEventInvalid
	}
	return offsets, nil
}

func calendarCreateMonday(value time.Time) time.Time {
	return ical.CalendarRecurrenceMonday(value)
}

func calendarCreateDaysInMonth(year int, month time.Month) int {
	return ical.CalendarRecurrenceDaysInMonth(year, month)
}

func canonicalizeCalendarCreateExdates(temporal CalendarEvent, exdates []string) ([]string, error) {
	if len(exdates) > calendarCreateMaxExdates { return nil, ErrEventInvalid }
	if len(exdates) == 0 { return nil, nil }
	canonical := make([]string, len(exdates))
	for index, value := range exdates {
		if temporal.AllDay {
			if _, err := parseCalendarDate(value); err != nil { return nil, ErrEventInvalid }
		} else {
			parsed, err := time.Parse("2006-01-02T15:04:05Z", value)
			if err != nil || parsed.Location() != time.UTC || parsed.Format("2006-01-02T15:04:05Z") != value { return nil, ErrEventInvalid }
		}
		canonical[index] = value
	}
	sort.Strings(canonical)
	for index := 1; index < len(canonical); index++ { if canonical[index-1] == canonical[index] { return nil, ErrEventInvalid } }
	return canonical, nil
}

func calendarCreateRequestHash(input CalendarEventCreateInput) (MutationReceiptHash, error) {
	encoded, err := json.Marshal(calendarCreateFingerprint{Version: calendarCreateFingerprintVersion, Title: input.Title, Description: input.Description, Location: input.Location,
		Temporal: input.Temporal, RRULE: input.RRULE, Exdates: input.Exdates, ScheduledCallID: input.ScheduledCallID, ReminderMinutes: input.ReminderMinutes, Attendees: input.Attendees})
	if err != nil { return MutationReceiptHash{}, err }
	return MutationReceiptHash(sha256.Sum256(encoded)), nil
}

type calendarCreateFingerprint struct {
	Version int `json:"version"`; Title string `json:"title"`; Description string `json:"description"`; Location string `json:"location"`; Temporal CalendarEvent `json:"temporal"`
	RRULE string `json:"rrule"`; Exdates []string `json:"exdates"`; ScheduledCallID *uuid.UUID `json:"scheduled_call_id"`; ReminderMinutes *int `json:"reminder_minutes"`; Attendees []CalendarEventCreateAttendeeCandidate `json:"attendees"`
}

func calendarCreateUUID(domain string, components ...[]byte) uuid.UUID {
	hash := sha256.New(); _, _ = hash.Write([]byte(domain)); _, _ = hash.Write([]byte{0})
	for _, component := range components { var length [4]byte; binary.BigEndian.PutUint32(length[:], uint32(len(component))); _, _ = hash.Write(length[:]); _, _ = hash.Write(component) }
	sum := hash.Sum(nil); var identifier uuid.UUID; copy(identifier[:], sum); identifier[6] = (identifier[6] & 0x0f) | 0x50; identifier[8] = (identifier[8] & 0x3f) | 0x80
	return identifier
}

func calendarCreateIndexBytes(index int) []byte { var value [4]byte; binary.BigEndian.PutUint32(value[:], uint32(index)); return value[:] }

func calendarCreateSafeError(err error) error {
	var createConflict *CalendarEventCreateConflictError
	if errors.As(err, &createConflict) && createConflict != nil && validCalendarEventCreateConflictKind(createConflict.Kind()) {
		return newCalendarEventCreateConflict(createConflict.Kind())
	}
	switch { case errors.Is(err, ErrEventNotFound): return ErrEventNotFound; case errors.Is(err, ErrEventForbidden): return ErrEventForbidden; case errors.Is(err, ErrEventConflict): return ErrEventConflict; case errors.Is(err, ErrEventTransactionClosed): return ErrEventTransactionClosed; default: return ErrEventInvalid }
}

func validCalendarCreateIdempotencyKey(value string) bool { return utf8.ValidString(value) && len(value) >= calendarCreateMinIdempotencyKey && len(value) <= calendarCreateMaxIdempotencyKey && strings.TrimSpace(value) == value }

func validCalendarCreateText(value string, maximum int, required bool) bool {
	if !utf8.ValidString(value) || len(value) > maximum || (required && strings.TrimSpace(value) == "") { return false }
	for _, runeValue := range value { if unicode.IsControl(runeValue) { return false } }
	return true
}

func validCalendarCreateDisplayName(value *string) bool {
	return value == nil || (utf8.ValidString(*value) && len(*value) <= calendarCreateMaxNameBytes && utf8.RuneCountInString(*value) <= calendarCreateMaxNameRunes)
}

func validCalendarCreateEmail(value string) bool {
	value = normalizeCalendarCreateEmail(value)
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > calendarCreateMaxEmailBytes || strings.Count(value, "@") != 1 { return false }
	for _, runeValue := range value { if runeValue > unicode.MaxASCII || unicode.IsSpace(runeValue) || unicode.IsControl(runeValue) { return false } }
	local, domain, found := strings.Cut(value, "@")
	return found && local != "" && domain != "" && !strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".") && !strings.Contains(domain, "..")
}

func normalizeCalendarCreateEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
