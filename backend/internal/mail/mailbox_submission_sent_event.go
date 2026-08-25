package mail

import (
	"crypto/sha256"
	"encoding/binary"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const mailboxSubmissionSentEventDomain = "hub.mail.submission.sent.event.v1"

// MailboxSubmissionSentRecipientResolution is one account-scoped, canonical recipient route
// resolved by the persistence adapter for a confirmed submission.
type MailboxSubmissionSentRecipientResolution struct {
	Email         string
	DeliveryClass MailDeliveryClass
}

// MailboxSubmissionSentEventInput is the complete bounded, confirmed source for mail.sent.v1.
// RecipientResolutions must exactly cover Projection.To, Projection.CC, and Projection.BCC.
type MailboxSubmissionSentEventInput struct {
	AccountID            uuid.UUID
	OrganizationID       uuid.UUID
	ActorID              uuid.UUID
	SubmissionID         uuid.UUID
	MessageID            uuid.UUID
	ThreadID             uuid.UUID
	DispatchAuthorizedAt time.Time
	Projection           EngineMessageProjection
	RecipientResolutions []MailboxSubmissionSentRecipientResolution
}

// NewMailboxSubmissionSentEnvelopeV1 deterministically maps one confirmed owner submission to
// its sender-owned mail.sent.v1 envelope. It has no I/O and retains no caller-owned input.
func NewMailboxSubmissionSentEnvelopeV1(input MailboxSubmissionSentEventInput) (events.EnvelopeV1, error) {
	if input.AccountID == uuid.Nil || input.OrganizationID == uuid.Nil || input.ActorID == uuid.Nil ||
		input.SubmissionID == uuid.Nil || input.MessageID == uuid.Nil || input.ThreadID == uuid.Nil ||
		!mailboxSubmissionSentEventAuthorizedAtValid(input.DispatchAuthorizedAt) {
		return events.EnvelopeV1{}, mailboxSubmissionSentEventInvalid()
	}
	if err := syncValidateProjection(input.Projection); err != nil ||
		input.Projection.Origin != EngineProjectionOriginConfirmedOwnerSubmission ||
		input.Projection.Flags.Draft || input.Projection.SentAt == nil || !syncValidTime(*input.Projection.SentAt) {
		return events.EnvelopeV1{}, mailboxSubmissionSentEventInvalid()
	}
	deliveryClass, valid := mailboxSubmissionSentEventDeliveryClass(input.Projection, input.RecipientResolutions)
	if !valid {
		return events.EnvelopeV1{}, mailboxSubmissionSentEventInvalid()
	}
	from, err := syncEventAddress(input.Projection.From)
	if err != nil {
		return events.EnvelopeV1{}, mailboxSubmissionSentEventInvalid()
	}
	to, err := syncEventAddresses(input.Projection.To)
	if err != nil {
		return events.EnvelopeV1{}, mailboxSubmissionSentEventInvalid()
	}
	actorID := input.ActorID
	event, err := NewMailSentEnvelopeV1(
		mailboxSubmissionSentEventID(input.SubmissionID),
		input.DispatchAuthorizedAt,
		input.OrganizationID,
		&actorID,
		input.SubmissionID.String(),
		MailSentEventPayload{
			Mailbox:       MailboxEventID{AccountID: input.AccountID, ID: input.AccountID},
			Submission:    MailboxEventID{AccountID: input.AccountID, ID: input.SubmissionID},
			Message:       MailboxEventID{AccountID: input.AccountID, ID: input.MessageID},
			Thread:        MailboxEventID{AccountID: input.AccountID, ID: input.ThreadID},
			From:          from,
			To:            to,
			SentAt:        *input.Projection.SentAt,
			DeliveryClass: deliveryClass,
		},
	)
	if err != nil {
		return events.EnvelopeV1{}, mailboxSubmissionSentEventInvalid()
	}
	return event, nil
}

func mailboxSubmissionSentEventAuthorizedAtValid(value time.Time) bool {
	if !syncValidTime(value) || value.Nanosecond() != 0 || value != value.Round(0) {
		return false
	}
	return !value.Before(time.Unix(0, 0).UTC())
}

func mailboxSubmissionSentEventDeliveryClass(projection EngineMessageProjection, resolutions []MailboxSubmissionSentRecipientResolution) (MailDeliveryClass, bool) {
	recipientCount := len(projection.To) + len(projection.CC) + len(projection.BCC)
	if recipientCount < 1 || recipientCount > MaxEngineDraftTotalRecipientItems || resolutions == nil || len(resolutions) != recipientCount {
		return "", false
	}
	recipients := make(map[string]struct{}, recipientCount)
	for _, group := range [][]EngineAddress{projection.To, projection.CC, projection.BCC} {
		for _, recipient := range group {
			if _, duplicate := recipients[recipient.Email]; duplicate {
				return "", false
			}
			recipients[recipient.Email] = struct{}{}
		}
	}
	routes := make(map[string]MailDeliveryClass, len(resolutions))
	for _, resolution := range resolutions {
		normalized, err := NormalizeAddress(EngineAddress{Email: resolution.Email})
		if err != nil || normalized.Email != resolution.Email || !validMailDeliveryClass(resolution.DeliveryClass) {
			return "", false
		}
		if _, duplicate := routes[resolution.Email]; duplicate {
			return "", false
		}
		routes[resolution.Email] = resolution.DeliveryClass
	}
	if len(routes) != len(recipients) {
		return "", false
	}
	deliveryClass := MailDeliveryClassInternal
	for recipient := range recipients {
		route, found := routes[recipient]
		if !found {
			return "", false
		}
		if mailboxSubmissionSentEventDeliveryPriority(route) > mailboxSubmissionSentEventDeliveryPriority(deliveryClass) {
			deliveryClass = route
		}
	}
	return deliveryClass, true
}

func mailboxSubmissionSentEventDeliveryPriority(value MailDeliveryClass) int {
	switch value {
	case MailDeliveryClassExternal:
		return 3
	case MailDeliveryClassInstanceDomain:
		return 2
	case MailDeliveryClassInternal:
		return 1
	default:
		return 0
	}
}

func mailboxSubmissionSentEventID(submissionID uuid.UUID) uuid.UUID {
	frame := make([]byte, 0, 4+len(mailboxSubmissionSentEventDomain)+len(submissionID))
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(mailboxSubmissionSentEventDomain)))
	frame = append(frame, length[:]...)
	frame = append(frame, mailboxSubmissionSentEventDomain...)
	frame = append(frame, submissionID[:]...)
	digest := sha256.Sum256(frame)
	var result uuid.UUID
	copy(result[:], digest[:16])
	result[6] = result[6]&0x0f | 0x80
	result[8] = result[8]&0x3f | 0x80
	return result
}

func mailboxSubmissionSentEventInvalid() error {
	return mailEventPolicyError(MailEventErrorInvalid)
}
