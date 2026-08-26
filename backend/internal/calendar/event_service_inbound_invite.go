package calendar

import (
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/ical"
)

// CalendarInboundInviteRequest is an opaque normalized inbound iTIP request.
type CalendarInboundInviteRequest struct {
	message        ITIPMessage
	identity       ITIPIdentity
	authentication CalendarInboundAuthenticationClass
}

func NewCalendarInboundInviteRequest(
	message ITIPMessage,
	authentication CalendarInboundAuthenticationClass,
) (CalendarInboundInviteRequest, error) {
	normalized, err := ical.NormalizeITIPMessage(message)
	if err != nil {
		return CalendarInboundInviteRequest{}, ErrEventInvalid
	}

	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil {
		return CalendarInboundInviteRequest{}, ErrEventInvalid
	}
	if identity.CanonicalVersion != ical.CanonicalITIPVersion ||
		(identity.Method != ITIPMethodRequest && identity.Method != ITIPMethodCancel) ||
		identity.Sequence < 0 ||
		identity.Sequence > ical.MaxITIPSequence ||
		!validCalendarInboundAuthentication(authentication) {
		return CalendarInboundInviteRequest{}, ErrEventInvalid
	}

	return CalendarInboundInviteRequest{
		message:        normalized,
		identity:       identity,
		authentication: authentication,
	}, nil
}

func (request CalendarInboundInviteRequest) valid() bool {
	normalized, err := ical.NormalizeITIPMessage(request.message)
	if err != nil {
		return false
	}

	identity, err := ical.CanonicalITIPIdentity(normalized)
	if err != nil {
		return false
	}

	return identity == request.identity &&
		identity.CanonicalVersion == ical.CanonicalITIPVersion &&
		(identity.Method == ITIPMethodRequest || identity.Method == ITIPMethodCancel) &&
		identity.Sequence >= 0 &&
		identity.Sequence <= ical.MaxITIPSequence &&
		validCalendarInboundAuthentication(request.authentication)
}

func (request CalendarInboundInviteRequest) clone() (CalendarInboundInviteRequest, error) {
	return NewCalendarInboundInviteRequest(
		request.message,
		request.authentication,
	)
}

// CalendarInboundInviteAuthorization is opaque server-resolved authorization
// and must only be created by its constructor.
type CalendarInboundInviteAuthorization struct {
	ownerID         uuid.UUID
	current         *CalendarInboundObjectState
	series          *CalendarInboundObjectState
	resourceLimited bool
}

func NewCalendarInboundInviteAuthorization(
	ownerID uuid.UUID,
	current *CalendarInboundObjectState,
	series *CalendarInboundObjectState,
	resourceLimited bool,
) (CalendarInboundInviteAuthorization, error) {
	if ownerID == uuid.Nil || (resourceLimited && (current != nil || series != nil)) {
		return CalendarInboundInviteAuthorization{}, ErrEventInvalid
	}

	if current != nil {
		value := current.clone()
		current = &value
		if !current.valid() || current.Component().OwnerID() != ownerID {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
	}

	if series != nil {
		value := series.clone()
		series = &value
		if !series.valid() ||
			series.Component().OwnerID() != ownerID ||
			!series.Component().IsSeries() {
			return CalendarInboundInviteAuthorization{}, ErrEventInvalid
		}
	}

	return CalendarInboundInviteAuthorization{
		ownerID:         ownerID,
		current:         current,
		series:          series,
		resourceLimited: resourceLimited,
	}, nil
}

func (authorization CalendarInboundInviteAuthorization) valid() bool {
	_, err := NewCalendarInboundInviteAuthorization(
		authorization.ownerID,
		authorization.current,
		authorization.series,
		authorization.resourceLimited,
	)
	return err == nil
}

func (authorization CalendarInboundInviteAuthorization) clone() (
	CalendarInboundInviteAuthorization,
	error,
) {
	return NewCalendarInboundInviteAuthorization(
		authorization.ownerID,
		authorization.current,
		authorization.series,
		authorization.resourceLimited,
	)
}
