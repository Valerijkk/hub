package calendar

import "github.com/valerii/hub/backend/internal/ical"

var (
	ErrInvalidITIPMessage = ical.ErrInvalidITIPMessage
	ErrInvalidITIPMethod = ical.ErrInvalidITIPMethod
	ErrInvalidITIPStatus = ical.ErrInvalidITIPStatus
	ErrInvalidITIPParticipant = ical.ErrInvalidITIPParticipant
	ErrInvalidITIPTemporal = ical.ErrInvalidITIPTemporal
	ErrInvalidITIPRecurrence = ical.ErrInvalidITIPRecurrence
	ErrInvalidITIPStamp = ical.ErrInvalidITIPStamp
	ErrInvalidITIPSequence = ical.ErrInvalidITIPSequence
	ErrInvalidITIPEmail = ical.ErrInvalidITIPEmail
	ErrInvalidITIPLimits = ical.ErrInvalidITIPLimits
	ErrITIPLimitExceeded = ical.ErrITIPLimitExceeded
	ErrInvalidITIPSyntax = ical.ErrInvalidITIPSyntax
	ErrUnsupportedITIPContent = ical.ErrUnsupportedITIPContent
)

type ITIPParseLimits = ical.ITIPParseLimits
type ITIPMethod = ical.ITIPMethod
type ITIPStatus = ical.ITIPStatus
type ITIPParticipantRole = ical.ITIPParticipantRole
type ITIPOrganizer = ical.ITIPOrganizer
type ITIPParticipant = ical.ITIPParticipant
type ITIPTimed = ical.ITIPTimed
type ITIPAllDay = ical.ITIPAllDay
type ITIPTemporal = ical.ITIPTemporal
type ITIPRecurrenceID = ical.ITIPRecurrenceID
type ITIPMessage = ical.ITIPMessage
type ITIPIdentity = ical.ITIPIdentity

const (
	ITIPMethodRequest = ical.ITIPMethodRequest
	ITIPMethodCancel = ical.ITIPMethodCancel
	ITIPMethodReply = ical.ITIPMethodReply
	ITIPStatusConfirmed = ical.ITIPStatusConfirmed
	ITIPStatusTentative = ical.ITIPStatusTentative
	ITIPStatusCancelled = ical.ITIPStatusCancelled
	ITIPParticipantRequired = ical.ITIPParticipantRequired
	ITIPParticipantOptional = ical.ITIPParticipantOptional
)

func DefaultITIPParseLimits() ITIPParseLimits {
	return ical.DefaultITIPParseLimits()
}

func ParseITIP(payload []byte, limits ITIPParseLimits) (ITIPMessage, error) {
	return ical.ParseITIP(payload, limits)
}

func SerializeITIP(message ITIPMessage) ([]byte, error) {
	return ical.SerializeITIP(message)
}

func CanonicalITIPIdentity(message ITIPMessage) (ITIPIdentity, error) {
	return ical.CanonicalITIPIdentity(message)
}

func validateITIPUID(value string) error {
	return ical.ValidateITIPUID(value)
}
