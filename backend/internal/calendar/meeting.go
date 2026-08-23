package calendar

import (
	"crypto/sha256"
	"encoding/json"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	meetingCreateFingerprintVersion = 1
	meetingUpdateFingerprintVersion = 1
	meetingCancelFingerprintVersion = 1
	meetingCreateMaxAttendees       = 50
	meetingMutationMaxStoredAttendees = 1000
	meetingCreateMaxTitleBytes      = 255
	meetingCreateMaxDescription     = 2000
	meetingCreateMaxTimezoneBytes   = 64
	meetingCreateMaxDuration        = 24 * time.Hour
)

// MeetingCreateInput contains only client-controlled meeting values. Actor and
// organization identities are always resolved from the authenticated tx scope.
type MeetingCreateInput struct {
	IdempotencyKey  string
	Title           string
	Description     string
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	StartsAt        time.Time
	EndsAt          time.Time
	Timezone        string
	RRULE           string
	AttendeeUserIDs []uuid.UUID
}

// MeetingCreateCommand is the immutable canonical form used for receipt
// hashing, authorization and persistence.
type MeetingCreateCommand struct {
	input       MeetingCreateInput
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

func NewMeetingCreateCommand(input MeetingCreateInput) (MeetingCreateCommand, error) {
	startsAt, startsOK := canonicalAvailabilityTime(input.StartsAt)
	endsAt, endsOK := canonicalAvailabilityTime(input.EndsAt)
	if !validCalendarCreateIdempotencyKey(input.IdempotencyKey) ||
		!validCalendarCreateText(input.Title, meetingCreateMaxTitleBytes, true) ||
		!validCalendarCreateText(input.Description, meetingCreateMaxDescription, false) ||
		input.ProjectID == uuid.Nil || input.ChannelID == uuid.Nil ||
		!startsOK || !endsOK || !endsAt.After(startsAt) ||
		endsAt.Sub(startsAt) > meetingCreateMaxDuration ||
		endsAt.Sub(startsAt)%time.Minute != 0 ||
		!utf8.ValidString(input.Timezone) || len(input.Timezone) > meetingCreateMaxTimezoneBytes ||
		len(input.AttendeeUserIDs) == 0 || len(input.AttendeeUserIDs) > meetingCreateMaxAttendees {
		return MeetingCreateCommand{}, ErrEventInvalid
	}
	temporal := CalendarEvent{StartsAt: startsAt, EndsAt: endsAt, Timezone: input.Timezone}
	if temporal.Validate() != nil {
		return MeetingCreateCommand{}, ErrEventInvalid
	}
	if input.RRULE != "" {
		normalized, err := normalizeCalendarCreateRRULE(input.RRULE, temporal)
		if err != nil {
			return MeetingCreateCommand{}, ErrEventInvalid
		}
		input.RRULE = normalized
	}
	input.StartsAt = startsAt
	input.EndsAt = endsAt
	input.AttendeeUserIDs = append([]uuid.UUID(nil), input.AttendeeUserIDs...)
	sort.Slice(input.AttendeeUserIDs, func(left, right int) bool {
		return input.AttendeeUserIDs[left].String() < input.AttendeeUserIDs[right].String()
	})
	for index, attendeeID := range input.AttendeeUserIDs {
		if attendeeID == uuid.Nil || index > 0 && attendeeID == input.AttendeeUserIDs[index-1] {
			return MeetingCreateCommand{}, ErrEventInvalid
		}
	}
	requestHash, err := meetingCreateRequestHash(input)
	if err != nil {
		return MeetingCreateCommand{}, ErrEventInvalid
	}
	return MeetingCreateCommand{
		input:       input,
		keyHash:     MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))),
		requestHash: requestHash,
	}, nil
}

func (command MeetingCreateCommand) valid() bool {
	rebuilt, err := NewMeetingCreateCommand(command.input)
	return err == nil && rebuilt.keyHash == command.keyHash && rebuilt.requestHash == command.requestHash
}

func (command MeetingCreateCommand) allowsActor(actorID uuid.UUID) bool {
	if actorID == uuid.Nil {
		return false
	}
	for _, attendeeID := range command.input.AttendeeUserIDs {
		if attendeeID == actorID {
			return false
		}
	}
	return true
}

func meetingCreateRequestHash(input MeetingCreateInput) (MutationReceiptHash, error) {
	encoded, err := json.Marshal(meetingCreateFingerprint{
		Version:           meetingCreateFingerprintVersion,
		Title:             input.Title,
		Description:       input.Description,
		ProjectID:         input.ProjectID,
		ChannelID:         input.ChannelID,
		StartsAt:          input.StartsAt,
		EndsAt:            input.EndsAt,
		Timezone:          input.Timezone,
		RRULE:             input.RRULE,
		AttendeeUserIDs:   input.AttendeeUserIDs,
	})
	if err != nil {
		return MutationReceiptHash{}, err
	}
	return MutationReceiptHash(sha256.Sum256(encoded)), nil
}

type meetingCreateFingerprint struct {
	Version           int         `json:"version"`
	Title             string      `json:"title"`
	Description       string      `json:"description"`
	ProjectID         uuid.UUID   `json:"project_id"`
	ChannelID         uuid.UUID   `json:"channel_id"`
	StartsAt          time.Time   `json:"starts_at"`
	EndsAt            time.Time   `json:"ends_at"`
	Timezone          string      `json:"timezone"`
	RRULE             string      `json:"rrule"`
	AttendeeUserIDs   []uuid.UUID `json:"attendee_user_ids"`
}

// MeetingUpdateInput is a full desired meeting replacement guarded by the
// canonical event version from If-Match.
type MeetingUpdateInput struct {
	EventID         uuid.UUID
	ExpectedVersion uint
	Desired         MeetingCreateInput
}

type MeetingUpdateCommand struct {
	input       MeetingUpdateInput
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

func NewMeetingUpdateCommand(input MeetingUpdateInput) (MeetingUpdateCommand, error) {
	canonical, err := NewMeetingCreateCommand(input.Desired)
	if err != nil || input.EventID == uuid.Nil || input.ExpectedVersion == 0 || input.ExpectedVersion == ^uint(0) {
		return MeetingUpdateCommand{}, ErrEventInvalid
	}
	input.Desired = canonical.input
	encoded, err := json.Marshal(struct {
		Version         int                `json:"version"`
		EventID         uuid.UUID          `json:"event_id"`
		ExpectedVersion uint               `json:"expected_version"`
		Desired         meetingCreateFingerprint `json:"desired"`
	}{
		Version: meetingUpdateFingerprintVersion, EventID: input.EventID, ExpectedVersion: input.ExpectedVersion,
		Desired: meetingCreateFingerprint{
			Version: meetingCreateFingerprintVersion, Title: input.Desired.Title, Description: input.Desired.Description,
			ProjectID: input.Desired.ProjectID, ChannelID: input.Desired.ChannelID, StartsAt: input.Desired.StartsAt,
			EndsAt: input.Desired.EndsAt, Timezone: input.Desired.Timezone, RRULE: input.Desired.RRULE,
			AttendeeUserIDs: append([]uuid.UUID(nil), input.Desired.AttendeeUserIDs...),
		},
	})
	if err != nil {
		return MeetingUpdateCommand{}, ErrEventInvalid
	}
	return MeetingUpdateCommand{
		input: input,
		keyHash: MutationReceiptHash(sha256.Sum256([]byte(input.Desired.IdempotencyKey))),
		requestHash: MutationReceiptHash(sha256.Sum256(encoded)),
	}, nil
}

func (command MeetingUpdateCommand) valid() bool {
	rebuilt, err := NewMeetingUpdateCommand(command.input)
	return err == nil && rebuilt.keyHash == command.keyHash && rebuilt.requestHash == command.requestHash
}

func (command MeetingUpdateCommand) allowsActor(actorID uuid.UUID) bool {
	canonical, err := NewMeetingCreateCommand(command.input.Desired)
	return err == nil && canonical.allowsActor(actorID)
}

type MeetingCancelInput struct {
	EventID         uuid.UUID
	ExpectedVersion uint
	IdempotencyKey  string
}

type MeetingCancelCommand struct {
	input       MeetingCancelInput
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

func NewMeetingCancelCommand(input MeetingCancelInput) (MeetingCancelCommand, error) {
	if input.EventID == uuid.Nil || input.ExpectedVersion == 0 || input.ExpectedVersion == ^uint(0) || !validCalendarCreateIdempotencyKey(input.IdempotencyKey) {
		return MeetingCancelCommand{}, ErrEventInvalid
	}
	encoded, err := json.Marshal(struct {
		Version         int       `json:"version"`
		EventID         uuid.UUID `json:"event_id"`
		ExpectedVersion uint      `json:"expected_version"`
	}{meetingCancelFingerprintVersion, input.EventID, input.ExpectedVersion})
	if err != nil {
		return MeetingCancelCommand{}, ErrEventInvalid
	}
	return MeetingCancelCommand{
		input: input,
		keyHash: MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))),
		requestHash: MutationReceiptHash(sha256.Sum256(encoded)),
	}, nil
}

func (command MeetingCancelCommand) valid() bool {
	rebuilt, err := NewMeetingCancelCommand(command.input)
	return err == nil && rebuilt.keyHash == command.keyHash && rebuilt.requestHash == command.requestHash
}

// MeetingCreateAuthorizationRequest is the bounded candidate identity set for
// the in-transaction project, channel and member resolver.
type MeetingCreateAuthorizationRequest struct {
	ProjectID         uuid.UUID
	ChannelID         uuid.UUID
	AttendeeUserIDs   []uuid.UUID
}

func (request MeetingCreateAuthorizationRequest) valid() bool {
	if request.ProjectID == uuid.Nil || request.ChannelID == uuid.Nil || len(request.AttendeeUserIDs) == 0 || len(request.AttendeeUserIDs) > meetingCreateMaxAttendees {
		return false
	}
	previous := ""
	for _, attendeeID := range request.AttendeeUserIDs {
		current := attendeeID.String()
		if attendeeID == uuid.Nil || current <= previous {
			return false
		}
		previous = current
	}
	return true
}

// MeetingCreateMember is an authoritative active same-organization channel
// member resolved under row locks.
type MeetingCreateMember struct {
	UserID      uuid.UUID
	Email       string
	DisplayName *string
}

func (member MeetingCreateMember) valid() bool {
	return member.UserID != uuid.Nil && validCalendarCreateEmail(member.Email) && validCalendarCreateDisplayName(member.DisplayName)
}

// MeetingCreateAuthorization carries the exact identities authorized by the
// transaction backend. It never carries caller-provided actor or organization.
type MeetingCreateAuthorization struct {
	ProjectID uuid.UUID
	ChannelID uuid.UUID
	Organizer MeetingCreateMember
	Attendees []MeetingCreateMember
}

func (authorization MeetingCreateAuthorization) valid() bool {
	if authorization.ProjectID == uuid.Nil || authorization.ChannelID == uuid.Nil || !authorization.Organizer.valid() || len(authorization.Attendees) == 0 || len(authorization.Attendees) > meetingCreateMaxAttendees {
		return false
	}
	previous := ""
	for _, attendee := range authorization.Attendees {
		current := attendee.UserID.String()
		if !attendee.valid() || attendee.UserID == authorization.Organizer.UserID || current <= previous {
			return false
		}
		previous = current
	}
	return true
}

// MeetingScheduledCallCommand is the complete atomic call/link write issued
// only after the canonical event and availability reservations exist in tx.
type MeetingScheduledCallCommand struct {
	callID            uuid.UUID
	eventID           uuid.UUID
	organizationID    uuid.UUID
	projectID         uuid.UUID
	channelID         uuid.UUID
	creatorID         uuid.UUID
	title             string
	description       string
	startsAt          time.Time
	endsAt            time.Time
	timezone          string
	rrule             string
	attendeeUserIDs   []uuid.UUID
}

func NewMeetingScheduledCallCommand(callID, eventID, organizationID, projectID, channelID, creatorID uuid.UUID, input MeetingCreateInput) (MeetingScheduledCallCommand, error) {
	rebuilt, err := NewMeetingCreateCommand(input)
	if err != nil || callID == uuid.Nil || eventID == uuid.Nil || organizationID == uuid.Nil || creatorID == uuid.Nil ||
		projectID != rebuilt.input.ProjectID || channelID != rebuilt.input.ChannelID {
		return MeetingScheduledCallCommand{}, ErrEventInvalid
	}
	for _, attendeeID := range rebuilt.input.AttendeeUserIDs {
		if attendeeID == creatorID {
			return MeetingScheduledCallCommand{}, ErrEventInvalid
		}
	}
	return MeetingScheduledCallCommand{
		callID:          callID,
		eventID:         eventID,
		organizationID:  organizationID,
		projectID:       projectID,
		channelID:       channelID,
		creatorID:       creatorID,
		title:           rebuilt.input.Title,
		description:     rebuilt.input.Description,
		startsAt:        rebuilt.input.StartsAt,
		endsAt:          rebuilt.input.EndsAt,
		timezone:        rebuilt.input.Timezone,
		rrule:           rebuilt.input.RRULE,
		attendeeUserIDs: append([]uuid.UUID(nil), rebuilt.input.AttendeeUserIDs...),
	}, nil
}

func (command MeetingScheduledCallCommand) valid() bool {
	input := MeetingCreateInput{
		IdempotencyKey:  "meeting-command-validation",
		Title:           command.title,
		Description:     command.description,
		ProjectID:       command.projectID,
		ChannelID:       command.channelID,
		StartsAt:        command.startsAt,
		EndsAt:          command.endsAt,
		Timezone:        command.timezone,
		RRULE:           command.rrule,
		AttendeeUserIDs: command.attendeeUserIDs,
	}
	_, err := NewMeetingScheduledCallCommand(command.callID, command.eventID, command.organizationID, command.projectID, command.channelID, command.creatorID, input)
	return err == nil
}

func (command MeetingScheduledCallCommand) DurationMinutes() int {
	return int(command.endsAt.Sub(command.startsAt) / time.Minute)
}

// MeetingScheduledCallPersistenceResult proves the call and immutable event
// link were both written before the transaction can commit.
type MeetingScheduledCallPersistenceResult struct {
	callID  uuid.UUID
	version uint
}

func NewMeetingScheduledCallPersistenceResult(callID uuid.UUID, version uint) (MeetingScheduledCallPersistenceResult, error) {
	if callID == uuid.Nil || version == 0 {
		return MeetingScheduledCallPersistenceResult{}, ErrEventInvalid
	}
	return MeetingScheduledCallPersistenceResult{callID: callID, version: version}, nil
}

func (result MeetingScheduledCallPersistenceResult) CallID() uuid.UUID { return result.callID }
func (result MeetingScheduledCallPersistenceResult) Version() uint     { return result.version }

type MeetingCreateResult struct {
	response MutationReceiptResponse
	replayed bool
}

func (result MeetingCreateResult) Response() MutationReceiptResponse { return result.response }
func (result MeetingCreateResult) Replayed() bool                     { return result.replayed }
func (result MeetingCreateResult) valid() bool                        { return result.response.valid() }

// MeetingStoredAttendee is the bounded authoritative master-row state used by
// update/cancel reconciliation. Removed rows are retained only for safe reuse.
type MeetingStoredAttendee struct {
	AttendeeID  uuid.UUID
	UserID      uuid.UUID
	Email       string
	DisplayName string
	Role        CalendarEventAttendeeRole
	Status      RSVPStatus
	Version     uint
	RemovedAt   *time.Time
}

type MeetingMutationSnapshot struct {
	EventID             uuid.UUID
	ScheduledCallID     uuid.UUID
	OrganizationID      uuid.UUID
	OwnerID             uuid.UUID
	UID                 string
	Title               string
	Description         string
	ProjectID           uuid.UUID
	ChannelID           uuid.UUID
	StartsAt            time.Time
	EndsAt              time.Time
	Timezone            string
	RRULE               string
	Sequence            uint
	EventVersion        uint
	ScheduledCallVersion uint
	OrganizerEmail      string
	Attendees           []MeetingStoredAttendee
}

func (snapshot MeetingMutationSnapshot) valid() bool {
	if snapshot.EventID == uuid.Nil || snapshot.ScheduledCallID == uuid.Nil || snapshot.OrganizationID == uuid.Nil || snapshot.OwnerID == uuid.Nil ||
		snapshot.ProjectID == uuid.Nil || snapshot.ChannelID == uuid.Nil || snapshot.EventVersion == 0 || snapshot.ScheduledCallVersion == 0 ||
		snapshot.UID != "hub-"+snapshot.EventID.String() || !validCalendarCreateEmail(snapshot.OrganizerEmail) ||
		len(snapshot.Attendees) < 2 || len(snapshot.Attendees) > meetingMutationMaxStoredAttendees {
		return false
	}
	canonical, err := NewMeetingCreateCommand(MeetingCreateInput{
		IdempotencyKey: "meeting-snapshot-validation", Title: snapshot.Title, Description: snapshot.Description,
		ProjectID: snapshot.ProjectID, ChannelID: snapshot.ChannelID, StartsAt: snapshot.StartsAt, EndsAt: snapshot.EndsAt,
		Timezone: snapshot.Timezone, RRULE: snapshot.RRULE, AttendeeUserIDs: snapshot.activeInviteeUserIDs(),
	})
	if err != nil || !canonical.allowsActor(snapshot.OwnerID) {
		return false
	}
	seenIDs := make(map[uuid.UUID]struct{}, len(snapshot.Attendees))
	seenUsers := make(map[uuid.UUID]struct{}, len(snapshot.Attendees))
	organizers := 0
	activeInvitees := 0
	for _, attendee := range snapshot.Attendees {
		if attendee.AttendeeID == uuid.Nil || attendee.UserID == uuid.Nil || attendee.Version == 0 || !validCalendarCreateEmail(attendee.Email) ||
			!validCalendarEventStoredAttendeeRoleState(attendee.Role, attendee.Status, "") {
			return false
		}
		if _, duplicate := seenIDs[attendee.AttendeeID]; duplicate {
			return false
		}
		seenIDs[attendee.AttendeeID] = struct{}{}
		if _, duplicate := seenUsers[attendee.UserID]; duplicate {
			return false
		}
		seenUsers[attendee.UserID] = struct{}{}
		if attendee.RemovedAt != nil {
			if attendee.RemovedAt.IsZero() || attendee.RemovedAt.Location() != time.UTC || attendee.Role == CalendarEventAttendeeRoleOrganizer || attendee.UserID == snapshot.OwnerID {
				return false
			}
			continue
		}
		userID := attendee.UserID
		if attendee.Role == CalendarEventAttendeeRoleOrganizer || attendee.UserID == snapshot.OwnerID {
			if attendee.Email != snapshot.OrganizerEmail || !validCalendarEventStoredOrganizerAttendeeState(attendee.Role, attendee.Status, "", &userID, snapshot.OwnerID, snapshot.OwnerID) {
				return false
			}
			organizers++
			continue
		}
		activeInvitees++
	}
	return organizers == 1 && activeInvitees >= 1 && activeInvitees <= meetingCreateMaxAttendees
}

func (snapshot MeetingMutationSnapshot) activeInviteeUserIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(snapshot.Attendees))
	for _, attendee := range snapshot.Attendees {
		if attendee.RemovedAt == nil && attendee.UserID != snapshot.OwnerID {
			ids = append(ids, attendee.UserID)
		}
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left].String() < ids[right].String() })
	return ids
}

func (snapshot MeetingMutationSnapshot) activeInviteeRecordIDs() []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(snapshot.Attendees))
	for _, attendee := range snapshot.Attendees {
		if attendee.RemovedAt == nil && attendee.UserID != snapshot.OwnerID {
			ids = append(ids, attendee.AttendeeID)
		}
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left].String() < ids[right].String() })
	return ids
}

type MeetingUpdateAuthorizationRequest struct {
	EventID uuid.UUID
	Desired MeetingCreateAuthorizationRequest
}

func (request MeetingUpdateAuthorizationRequest) valid() bool {
	return request.EventID != uuid.Nil && request.Desired.valid()
}

type MeetingUpdateAuthorization struct {
	Current MeetingMutationSnapshot
	Desired MeetingCreateAuthorization
}

func (authorization MeetingUpdateAuthorization) valid() bool {
	return authorization.Current.valid() && authorization.Desired.valid() && authorization.Desired.Organizer.UserID == authorization.Current.OwnerID
}

type MeetingCancelAuthorization struct {
	Current MeetingMutationSnapshot
}

func (authorization MeetingCancelAuthorization) valid() bool { return authorization.Current.valid() }

type MeetingUpdatePersistenceCommand struct {
	current         MeetingMutationSnapshot
	desired         MeetingCreateInput
	authorization   MeetingCreateAuthorization
	reservations    CalendarAvailabilityReservationSet
	expectedVersion uint
}

func NewMeetingUpdatePersistenceCommand(current MeetingMutationSnapshot, desired MeetingCreateInput, authorization MeetingCreateAuthorization, reservations CalendarAvailabilityReservationSet, expectedVersion uint) (MeetingUpdatePersistenceCommand, error) {
	canonical, err := NewMeetingCreateCommand(desired)
	if err != nil || !current.valid() || !authorization.valid() || !reservations.valid() || expectedVersion == 0 ||
		!meetingMutationSnapshotsEqual(current, authorization.Current) ||
		current.EventVersion != expectedVersion || authorization.Organizer.UserID != current.OwnerID ||
		authorization.ProjectID != canonical.input.ProjectID || authorization.ChannelID != canonical.input.ChannelID ||
		reservations.EventID() != current.EventID || reservations.OrganizationID() != current.OrganizationID || reservations.OrganizerID() != current.OwnerID ||
		len(authorization.Attendees) != len(canonical.input.AttendeeUserIDs) {
		return MeetingUpdatePersistenceCommand{}, ErrEventInvalid
	}
	for index, attendee := range authorization.Attendees {
		if attendee.UserID != canonical.input.AttendeeUserIDs[index] {
			return MeetingUpdatePersistenceCommand{}, ErrEventInvalid
		}
	}
	expectedParticipants := make([]uuid.UUID, 0, len(canonical.input.AttendeeUserIDs)+1)
	expectedParticipants = append(expectedParticipants, current.OwnerID)
	expectedParticipants = append(expectedParticipants, canonical.input.AttendeeUserIDs...)
	sort.Slice(expectedParticipants, func(left, right int) bool { return expectedParticipants[left].String() < expectedParticipants[right].String() })
	if !meetingMutationUUIDsEqual(expectedParticipants, reservations.Participants()) {
		return MeetingUpdatePersistenceCommand{}, ErrEventInvalid
	}
	return MeetingUpdatePersistenceCommand{
		current: cloneMeetingMutationSnapshot(current), desired: canonical.input, authorization: cloneMeetingCreateAuthorization(authorization),
		reservations: cloneCalendarAvailabilityReservationSet(reservations), expectedVersion: expectedVersion,
	}, nil
}

func (command MeetingUpdatePersistenceCommand) valid() bool {
	_, err := NewMeetingUpdatePersistenceCommand(command.current, command.desired, command.authorization, command.reservations, command.expectedVersion)
	return err == nil
}

type MeetingCancelPersistenceCommand struct {
	current         MeetingMutationSnapshot
	expectedVersion uint
}

func NewMeetingCancelPersistenceCommand(current MeetingMutationSnapshot, expectedVersion uint) (MeetingCancelPersistenceCommand, error) {
	if !current.valid() || expectedVersion == 0 || current.EventVersion != expectedVersion {
		return MeetingCancelPersistenceCommand{}, ErrEventInvalid
	}
	return MeetingCancelPersistenceCommand{current: cloneMeetingMutationSnapshot(current), expectedVersion: expectedVersion}, nil
}

func (command MeetingCancelPersistenceCommand) valid() bool {
	_, err := NewMeetingCancelPersistenceCommand(command.current, command.expectedVersion)
	return err == nil
}

type MeetingMutationPersistenceResult struct {
	eventVersion      uint
	callVersion       uint
	sequence          uint
	activeAttendeeIDs []uuid.UUID
	removedAttendeeIDs []uuid.UUID
}

func newMeetingMutationPersistenceResult(eventVersion, callVersion, sequence uint, activeAttendeeIDs, removedAttendeeIDs []uuid.UUID) (MeetingMutationPersistenceResult, error) {
	if eventVersion == 0 || callVersion == 0 || sequence == 0 ||
		len(activeAttendeeIDs) == 0 || len(activeAttendeeIDs) > meetingCreateMaxAttendees ||
		len(removedAttendeeIDs) > meetingCreateMaxAttendees {
		return MeetingMutationPersistenceResult{}, ErrEventInvalid
	}
	active := append([]uuid.UUID(nil), activeAttendeeIDs...)
	removed := append([]uuid.UUID(nil), removedAttendeeIDs...)
	sort.Slice(active, func(left, right int) bool { return active[left].String() < active[right].String() })
	sort.Slice(removed, func(left, right int) bool { return removed[left].String() < removed[right].String() })
	seen := make(map[uuid.UUID]struct{}, len(active)+len(removed))
	for _, identifier := range append(append([]uuid.UUID(nil), active...), removed...) {
		if identifier == uuid.Nil {
			return MeetingMutationPersistenceResult{}, ErrEventInvalid
		}
		if _, duplicate := seen[identifier]; duplicate {
			return MeetingMutationPersistenceResult{}, ErrEventInvalid
		}
		seen[identifier] = struct{}{}
	}
	return MeetingMutationPersistenceResult{
		eventVersion: eventVersion, callVersion: callVersion, sequence: sequence,
		activeAttendeeIDs: active, removedAttendeeIDs: removed,
	}, nil
}

func (result MeetingMutationPersistenceResult) EventVersion() uint { return result.eventVersion }
func (result MeetingMutationPersistenceResult) CallVersion() uint { return result.callVersion }
func (result MeetingMutationPersistenceResult) Sequence() uint { return result.sequence }
func (result MeetingMutationPersistenceResult) ActiveAttendeeIDs() []uuid.UUID { return append([]uuid.UUID(nil), result.activeAttendeeIDs...) }
func (result MeetingMutationPersistenceResult) RemovedAttendeeIDs() []uuid.UUID { return append([]uuid.UUID(nil), result.removedAttendeeIDs...) }
func (result MeetingMutationPersistenceResult) valid() bool {
	_, err := newMeetingMutationPersistenceResult(result.eventVersion, result.callVersion, result.sequence, result.activeAttendeeIDs, result.removedAttendeeIDs)
	return err == nil
}

type MeetingMutationResult struct {
	response MutationReceiptResponse
	replayed bool
}

func (result MeetingMutationResult) Response() MutationReceiptResponse { return result.response }
func (result MeetingMutationResult) Replayed() bool { return result.replayed }
func (result MeetingMutationResult) valid() bool { return result.response.valid() }

func cloneMeetingCreateAuthorizationRequest(request MeetingCreateAuthorizationRequest) MeetingCreateAuthorizationRequest {
	request.AttendeeUserIDs = append([]uuid.UUID(nil), request.AttendeeUserIDs...)
	return request
}

func cloneMeetingCreateAuthorization(authorization MeetingCreateAuthorization) MeetingCreateAuthorization {
	authorization.Organizer.DisplayName = cloneCalendarEventCreateString(authorization.Organizer.DisplayName)
	authorization.Attendees = append([]MeetingCreateMember(nil), authorization.Attendees...)
	for index := range authorization.Attendees {
		authorization.Attendees[index].DisplayName = cloneCalendarEventCreateString(authorization.Attendees[index].DisplayName)
	}
	return authorization
}

func cloneMeetingScheduledCallCommand(command MeetingScheduledCallCommand) MeetingScheduledCallCommand {
	command.attendeeUserIDs = append([]uuid.UUID(nil), command.attendeeUserIDs...)
	return command
}

func cloneMeetingMutationSnapshot(snapshot MeetingMutationSnapshot) MeetingMutationSnapshot {
	snapshot.Attendees = append([]MeetingStoredAttendee(nil), snapshot.Attendees...)
	for index := range snapshot.Attendees {
		if snapshot.Attendees[index].RemovedAt != nil {
			value := *snapshot.Attendees[index].RemovedAt
			snapshot.Attendees[index].RemovedAt = &value
		}
	}
	return snapshot
}

func cloneMeetingUpdateAuthorizationRequest(request MeetingUpdateAuthorizationRequest) MeetingUpdateAuthorizationRequest {
	request.Desired = cloneMeetingCreateAuthorizationRequest(request.Desired)
	return request
}

func cloneMeetingUpdateAuthorization(authorization MeetingUpdateAuthorization) MeetingUpdateAuthorization {
	authorization.Current = cloneMeetingMutationSnapshot(authorization.Current)
	authorization.Desired = cloneMeetingCreateAuthorization(authorization.Desired)
	return authorization
}

func cloneMeetingCancelAuthorization(authorization MeetingCancelAuthorization) MeetingCancelAuthorization {
	authorization.Current = cloneMeetingMutationSnapshot(authorization.Current)
	return authorization
}

func cloneMeetingUpdatePersistenceCommand(command MeetingUpdatePersistenceCommand) MeetingUpdatePersistenceCommand {
	command.current = cloneMeetingMutationSnapshot(command.current)
	command.desired.AttendeeUserIDs = append([]uuid.UUID(nil), command.desired.AttendeeUserIDs...)
	command.authorization = cloneMeetingCreateAuthorization(command.authorization)
	command.reservations = cloneCalendarAvailabilityReservationSet(command.reservations)
	return command
}

func cloneMeetingCancelPersistenceCommand(command MeetingCancelPersistenceCommand) MeetingCancelPersistenceCommand {
	command.current = cloneMeetingMutationSnapshot(command.current)
	return command
}

func meetingMutationSnapshotsEqual(left, right MeetingMutationSnapshot) bool {
	if left.EventID != right.EventID || left.ScheduledCallID != right.ScheduledCallID || left.OrganizationID != right.OrganizationID ||
		left.OwnerID != right.OwnerID || left.UID != right.UID || left.Title != right.Title || left.Description != right.Description ||
		left.ProjectID != right.ProjectID || left.ChannelID != right.ChannelID || !left.StartsAt.Equal(right.StartsAt) ||
		!left.EndsAt.Equal(right.EndsAt) || left.Timezone != right.Timezone || left.RRULE != right.RRULE ||
		left.Sequence != right.Sequence || left.EventVersion != right.EventVersion ||
		left.ScheduledCallVersion != right.ScheduledCallVersion || left.OrganizerEmail != right.OrganizerEmail ||
		len(left.Attendees) != len(right.Attendees) {
		return false
	}
	for index := range left.Attendees {
		leftAttendee, rightAttendee := left.Attendees[index], right.Attendees[index]
		if leftAttendee.AttendeeID != rightAttendee.AttendeeID || leftAttendee.UserID != rightAttendee.UserID ||
			leftAttendee.Email != rightAttendee.Email || leftAttendee.DisplayName != rightAttendee.DisplayName ||
			leftAttendee.Role != rightAttendee.Role || leftAttendee.Status != rightAttendee.Status ||
			leftAttendee.Version != rightAttendee.Version || !meetingMutationTimesEqual(leftAttendee.RemovedAt, rightAttendee.RemovedAt) {
			return false
		}
	}
	return true
}

func meetingMutationTimesEqual(left, right *time.Time) bool {
	return left == nil && right == nil || left != nil && right != nil && left.Equal(*right)
}

func meetingMutationUUIDsEqual(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
