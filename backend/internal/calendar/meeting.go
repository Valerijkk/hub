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
	meetingCreateMaxAttendees       = 50
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
