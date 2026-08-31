//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/outboxevent"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/internal/platform/events"
)

const (
	inboundRequestFixturePublishedProjectOutboxMailReceivedSubject          = "mail.received.v1"
	inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedSubject = "calendar.invite.updated.v1"
	inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedSubject   = "calendar.rsvp.changed.v1"
)

// inboundRequestFixturePublishedProjectOutboxProof retains the physical
// OutboxEvent identities proved inside the caller-owned transaction.
type inboundRequestFixturePublishedProjectOutboxProof struct {
	roots          inboundRequestFixtureRootProof
	calendar       inboundRequestFixtureCalendarProof
	engineAuthority inboundRequestFixturePublishedEngineAuthority
	p3OutboxID     uuid.UUID
	p4OutboxID     uuid.UUID
	p5OutboxID     uuid.UUID
}

type inboundRequestFixturePublishedProjectOutboxEnvelopeJSON struct {
	Subject        string          `json:"subject"`
	SchemaVersion  int             `json:"schema_version"`
	EventID        uuid.UUID       `json:"event_id"`
	OccurredAt     time.Time       `json:"occurred_at"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ActorID        *uuid.UUID      `json:"actor_id"`
	CorrelationID  string          `json:"correlation_id"`
	Payload        json.RawMessage `json:"payload"`
}

type inboundRequestFixturePublishedProjectOutboxAddressJSON struct {
	Email string  `json:"email"`
	Name  *string `json:"name,omitempty"`
}

type inboundRequestFixturePublishedProjectOutboxMailReceivedPayloadJSON struct {
	MailboxID               uuid.UUID                                          `json:"mailbox_id"`
	MessageID               uuid.UUID                                          `json:"message_id"`
	EngineMessageID         string                                             `json:"engine_message_id"`
	ThreadID                uuid.UUID                                          `json:"thread_id"`
	From                    inboundRequestFixturePublishedProjectOutboxAddressJSON   `json:"from"`
	To                      []inboundRequestFixturePublishedProjectOutboxAddressJSON `json:"to"`
	ReceivedAt              time.Time                                          `json:"received_at"`
	FolderIDs               []uuid.UUID                                        `json:"folder_ids"`
	HasAttachments          bool                                               `json:"has_attachments"`
	CalendarMethod          *string                                            `json:"calendar_method"`
	AuthenticationClass     string                                             `json:"authentication_class"`
	CalendarPartLocatorHash *string                                            `json:"calendar_part_locator_sha256"`
}

type inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedPayloadJSON struct {
	EventID       uuid.UUID `json:"event_id"`
	UID           string    `json:"uid"`
	Sequence      uint      `json:"sequence"`
	AttendeeIDs   []uuid.UUID `json:"attendee_ids"`
	ResetRSVP     bool      `json:"reset_rsvp"`
	ChangedFields []string  `json:"changed_fields"`
}

type inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedPayloadJSON struct {
	EventID             uuid.UUID `json:"event_id"`
	AttendeeID          uuid.UUID `json:"attendee_id"`
	AttendeeEmail       string    `json:"attendee_email"`
	PreviousStatus      string    `json:"previous_status"`
	NewStatus           string    `json:"new_status"`
	Version             uint      `json:"version"`
	RecurrenceID        *string   `json:"recurrence_id"`
	ReplyDeliveryStatus string    `json:"reply_delivery_status"`
}

func observeInboundRequestFixturePublishedProjectOutbox(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
	engineAuthority inboundRequestFixturePublishedEngineAuthority,
) (inboundRequestFixturePublishedProjectOutboxProof, bool) {
	if ctx == nil ||
		tx == nil ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!inboundRequestFixturePersistedMatchesPlan(persisted, plan) ||
		!inboundRequestFixtureRootProofMatchesPlan(roots, plan, persisted) ||
		!calendar.valid() ||
		!engineAuthority.valid() ||
		engineAuthority.nonce != plan.nonce {
		return inboundRequestFixturePublishedProjectOutboxProof{}, false
	}

	p3Rows, err := tx.OutboxEvent.Query().
		Where(
			outboxevent.OriginKindEQ("mail_message"),
			outboxevent.OriginIDEQ(roots.mailMessageID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxMailReceivedSubject),
			inboundRequestFixturePublishedProjectOutboxNoWait(),
		).
		Order(ent.Asc(outboxevent.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(p3Rows) != 1 ||
		!inboundRequestFixturePublishedProjectOutboxRowMatches(
			p3Rows[0],
			"mail_message",
			roots.mailMessageID,
			inboundRequestFixturePublishedProjectOutboxMailReceivedSubject,
		) ||
		!inboundRequestFixturePublishedProjectOutboxP3Valid(p3Rows[0].Payload, plan, roots, engineAuthority) {
		return inboundRequestFixturePublishedProjectOutboxProof{}, false
	}
	p3OutboxID := p3Rows[0].ID

	p4Rows, err := tx.OutboxEvent.Query().
		Where(
			outboxevent.OriginKindEQ("calendar_event"),
			outboxevent.OriginIDEQ(calendar.eventID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedSubject),
			inboundRequestFixturePublishedProjectOutboxNoWait(),
		).
		Order(ent.Asc(outboxevent.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(p4Rows) != 1 ||
		!inboundRequestFixturePublishedProjectOutboxRowMatches(
			p4Rows[0],
			"calendar_event",
			calendar.eventID,
			inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedSubject,
		) ||
		!inboundRequestFixturePublishedProjectOutboxP4Valid(p4Rows[0].Payload, roots, calendar, plan) {
		return inboundRequestFixturePublishedProjectOutboxProof{}, false
	}
	p4OutboxID := p4Rows[0].ID

	p5Rows, err := tx.OutboxEvent.Query().
		Where(
			outboxevent.OriginKindEQ("calendar_attendee"),
			outboxevent.OriginIDEQ(calendar.attendeeID),
			outboxevent.EventTypeEQ(inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedSubject),
			inboundRequestFixturePublishedProjectOutboxNoWait(),
		).
		Order(ent.Asc(outboxevent.FieldID)).
		Limit(2).
		All(ctx)
	if err != nil || len(p5Rows) != 1 ||
		!inboundRequestFixturePublishedProjectOutboxRowMatches(
			p5Rows[0],
			"calendar_attendee",
			calendar.attendeeID,
			inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedSubject,
		) ||
		!inboundRequestFixturePublishedProjectOutboxP5Valid(p5Rows[0].Payload, roots, calendar, plan) {
		return inboundRequestFixturePublishedProjectOutboxProof{}, false
	}
	p5OutboxID := p5Rows[0].ID

	proof := inboundRequestFixturePublishedProjectOutboxProof{
		roots:           roots,
		calendar:        calendar,
		engineAuthority: engineAuthority,
		p3OutboxID:      p3OutboxID,
		p4OutboxID:      p4OutboxID,
		p5OutboxID:      p5OutboxID,
	}
	if !proof.valid(roots, calendar, engineAuthority) {
		return inboundRequestFixturePublishedProjectOutboxProof{}, false
	}
	return proof, true
}

func (proof inboundRequestFixturePublishedProjectOutboxProof) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
	engineAuthority inboundRequestFixturePublishedEngineAuthority,
) bool {
	return roots.valid() &&
		calendar.valid() &&
		engineAuthority.valid() &&
		proof.roots == roots &&
		proof.calendar == calendar &&
		proof.engineAuthority == engineAuthority &&
		proof.p3OutboxID != uuid.Nil &&
		proof.p4OutboxID != uuid.Nil &&
		proof.p5OutboxID != uuid.Nil &&
		proof.p3OutboxID != proof.p4OutboxID &&
		proof.p3OutboxID != proof.p5OutboxID &&
		proof.p4OutboxID != proof.p5OutboxID
}

func inboundRequestFixturePublishedProjectOutboxRowMatches(
	row *ent.OutboxEvent,
	originKind string,
	originID uuid.UUID,
	eventType string,
) bool {
	return row != nil &&
		row.ID != uuid.Nil &&
		row.OriginKind != nil &&
		row.OriginID != nil &&
		originID != uuid.Nil &&
		*row.OriginKind == originKind &&
		*row.OriginID == originID &&
		row.EventType == eventType
}

func inboundRequestFixturePublishedProjectOutboxP3Valid(
	encoded []byte,
	plan inboundRequestFixturePlan,
	roots inboundRequestFixtureRootProof,
	engineAuthority inboundRequestFixturePublishedEngineAuthority,
) bool {
	envelope, valid := inboundRequestFixturePublishedProjectOutboxDecodeEnvelope(encoded)
	if !valid ||
		envelope.Subject != inboundRequestFixturePublishedProjectOutboxMailReceivedSubject ||
		envelope.EventID != roots.envelopeEventID ||
		!envelope.OccurredAt.Equal(plan.fixedAt) ||
		envelope.OrganizationID != roots.organizationID ||
		envelope.ActorID != nil ||
		envelope.CorrelationID != plan.correlationID {
		return false
	}

	var payload inboundRequestFixturePublishedProjectOutboxMailReceivedPayloadJSON
	if !inboundRequestFixturePublishedProjectOutboxDecodeJSON(envelope.Payload, &payload) ||
		!inboundRequestFixturePublishedProjectOutboxCanonicalJSONEquals(envelope.Payload, payload) ||
		payload.MailboxID != roots.mailAccountID ||
		payload.MessageID != roots.mailMessageID ||
		payload.EngineMessageID != engineAuthority.messageID ||
		payload.ThreadID != roots.mailThreadID ||
		payload.From.Email != plan.organizerAddress ||
		payload.From.Name != nil ||
		len(payload.To) != 1 ||
		payload.To[0].Email != plan.ownerAddress ||
		payload.To[0].Name != nil ||
		!payload.ReceivedAt.Equal(plan.fixedAt) ||
		len(payload.FolderIDs) != 1 ||
		payload.FolderIDs[0] != roots.mailFolderID ||
		payload.HasAttachments ||
		payload.CalendarMethod == nil ||
		*payload.CalendarMethod != "request" ||
		payload.AuthenticationClass != "trusted_internal" ||
		payload.CalendarPartLocatorHash == nil ||
		*payload.CalendarPartLocatorHash != hex.EncodeToString(plan.calendarLocatorHash[:]) {
		return false
	}
	return true
}

func inboundRequestFixturePublishedProjectOutboxP4Valid(
	encoded []byte,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
	plan inboundRequestFixturePlan,
) bool {
	envelope, valid := inboundRequestFixturePublishedProjectOutboxDecodeEnvelope(encoded)
	if !valid ||
		envelope.Subject != inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedSubject ||
		envelope.EventID == uuid.Nil ||
		envelope.OrganizationID != roots.organizationID ||
		envelope.ActorID != nil ||
		envelope.CorrelationID != calendar.eventID.String() {
		return false
	}

	var payload inboundRequestFixturePublishedProjectOutboxCalendarInviteUpdatedPayloadJSON
	if !inboundRequestFixturePublishedProjectOutboxDecodeJSON(envelope.Payload, &payload) ||
		!inboundRequestFixturePublishedProjectOutboxCanonicalJSONEquals(envelope.Payload, payload) ||
		payload.EventID != calendar.eventID ||
		payload.UID != plan.itipUID ||
		payload.Sequence != 0 ||
		len(payload.AttendeeIDs) != 1 ||
		payload.AttendeeIDs[0] != calendar.attendeeID ||
		payload.ResetRSVP ||
		!inboundRequestFixturePublishedProjectOutboxP4ChangedFieldsValid(payload.ChangedFields) {
		return false
	}
	return true
}

func inboundRequestFixturePublishedProjectOutboxP5Valid(
	encoded []byte,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
	plan inboundRequestFixturePlan,
) bool {
	envelope, valid := inboundRequestFixturePublishedProjectOutboxDecodeEnvelope(encoded)
	if !valid ||
		envelope.Subject != inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedSubject ||
		envelope.EventID == uuid.Nil ||
		envelope.OrganizationID != roots.organizationID ||
		envelope.ActorID == nil ||
		*envelope.ActorID != roots.ownerUserID ||
		envelope.CorrelationID != calendar.eventID.String() {
		return false
	}

	var payload inboundRequestFixturePublishedProjectOutboxCalendarRSVPChangedPayloadJSON
	if !inboundRequestFixturePublishedProjectOutboxDecodeJSON(envelope.Payload, &payload) ||
		!inboundRequestFixturePublishedProjectOutboxCanonicalJSONEquals(envelope.Payload, payload) ||
		payload.EventID != calendar.eventID ||
		payload.AttendeeID != calendar.attendeeID ||
		payload.AttendeeEmail != plan.ownerAddress ||
		payload.PreviousStatus != "needs_action" ||
		payload.NewStatus != "accepted" ||
		payload.Version != 2 ||
		payload.RecurrenceID != nil ||
		payload.ReplyDeliveryStatus != "pending" {
		return false
	}
	return true
}

func inboundRequestFixturePublishedProjectOutboxDecodeEnvelope(
	encoded []byte,
) (inboundRequestFixturePublishedProjectOutboxEnvelopeJSON, bool) {
	decoded, err := events.DecodeEnvelopeV1(encoded)
	if err != nil {
		return inboundRequestFixturePublishedProjectOutboxEnvelopeJSON{}, false
	}

	canonical, err := decoded.MarshalJSON()
	if err != nil || !bytes.Equal(canonical, encoded) {
		return inboundRequestFixturePublishedProjectOutboxEnvelopeJSON{}, false
	}

	var envelope inboundRequestFixturePublishedProjectOutboxEnvelopeJSON
	if !inboundRequestFixturePublishedProjectOutboxDecodeJSON(canonical, &envelope) ||
		!inboundRequestFixturePublishedProjectOutboxCanonicalJSONEquals(canonical, envelope) ||
		envelope.Subject != decoded.Subject() ||
		envelope.EventID != decoded.EventID() {
		return inboundRequestFixturePublishedProjectOutboxEnvelopeJSON{}, false
	}
	return envelope, true
}

func inboundRequestFixturePublishedProjectOutboxDecodeJSON[T any](encoded []byte, target *T) bool {
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	var trailing json.RawMessage
	return decoder.Decode(&trailing) == io.EOF
}

func inboundRequestFixturePublishedProjectOutboxCanonicalJSONEquals[T any](encoded []byte, value T) bool {
	canonical, err := json.Marshal(value)
	return err == nil && bytes.Equal(canonical, encoded)
}

func inboundRequestFixturePublishedProjectOutboxP4ChangedFieldsValid(values []string) bool {
	expected := [...]string{
		"title",
		"description",
		"location",
		"starts_at",
		"ends_at",
		"start_date",
		"end_date_exclusive",
		"all_day",
		"timezone",
		"rrule",
		"exdates",
		"attendees",
	}
	if len(values) != len(expected) {
		return false
	}
	for index := range expected {
		if values[index] != expected[index] {
			return false
		}
	}
	return true
}

func inboundRequestFixturePublishedProjectOutboxNoWait() predicate.OutboxEvent {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
