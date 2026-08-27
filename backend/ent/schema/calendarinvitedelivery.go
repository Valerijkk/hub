package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CalendarInviteDelivery stores one immutable Calendar-to-Mail delivery intent
// for a source envelope and attendee, plus only retry-safe execution metadata.
type CalendarInviteDelivery struct {
	ent.Schema
}

func (CalendarInviteDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}).
			Immutable(),
		field.UUID("calendar_event_id", uuid.UUID{}).
			Immutable(),
		field.UUID("calendar_attendee_id", uuid.UUID{}).
			Immutable(),
		field.UUID("source_envelope_id", uuid.UUID{}).
			Immutable(),
		field.Enum("kind").
			Values("created", "updated", "cancelled").
			Immutable(),
		field.Int("event_sequence").
			NonNegative().
			Immutable(),
		field.String("recurrence_id").
			Optional().
			Nillable().
			MaxLen(255).
			Immutable(),
		field.Enum("state").
			Values("pending", "submitting", "unknown", "sent", "superseded", "failed_permanent").
			Default("pending"),
		field.UUID("mail_submission_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("attempts").
			Default(0).
			Min(0).
			Max(8),
		field.UUID("lease_token", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int64("claim_generation").
			Default(0).
			Min(0).
			Max(2147483647),
		field.Time("lease_until").
			Optional().
			Nillable(),
		field.Time("next_attempt_at").
			Optional().
			Nillable(),
		field.Enum("last_error_code").
			Values("mailbox_unavailable", "engine_unavailable", "submission_rejected", "submission_outcome_unknown", "invalid_recipient", "materialization_invalid", "policy_denied", "retry_exhausted", "internal_error").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("terminal_at").
			Optional().
			Nillable(),
	}
}

func (CalendarInviteDelivery) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_invite_deliveries_event_sequence_non_negative": "event_sequence >= 0",
			"calendar_invite_deliveries_attempts_bounds":        "attempts >= 0 AND attempts <= 8",
			"calendar_invite_deliveries_lease_pair":              "COALESCE((lease_token IS NULL) = (lease_until IS NULL), FALSE)",
			"calendar_invite_deliveries_claim_generation_bounds": "claim_generation >= 0 AND claim_generation <= 2147483647",
			"calendar_invite_deliveries_state_shape": "COALESCE((state = 'pending' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND terminal_at IS NULL AND next_attempt_at IS NOT NULL AND ((attempts = 0 AND last_error_code IS NULL) OR (attempts BETWEEN 1 AND 8 AND last_error_code IN ('mailbox_unavailable','engine_unavailable','internal_error')))) OR (state = 'submitting' AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NULL) OR (state = 'unknown' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND terminal_at IS NULL AND last_error_code IN ('mailbox_unavailable','engine_unavailable','submission_outcome_unknown')) OR (state = 'sent' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL) OR (state = 'superseded' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL) OR (state = 'failed_permanent' AND lease_token IS NULL AND lease_until IS NULL AND next_attempt_at IS NULL AND last_error_code IN ('submission_rejected','invalid_recipient','materialization_invalid','policy_denied','retry_exhausted') AND terminal_at IS NOT NULL), FALSE)",
		}),
	}
}

func (CalendarInviteDelivery) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("event", CalendarEvent.Type).
			Field("calendar_event_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("attendee", CalendarAttendee.Type).
			Field("calendar_attendee_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("mail_submission", MailSubmission.Type).
			Field("mail_submission_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.NoAction)),
	}
}

func (CalendarInviteDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_envelope_id", "calendar_attendee_id").
			Unique().
			StorageKey("calendar_invite_deliveries_source_attendee_uidx"),
		index.Fields("calendar_event_id", "calendar_attendee_id", "event_sequence", "kind").
			Unique().
			StorageKey("calendar_invite_deliveries_event_attendee_sequence_kind_master_uidx").
			Annotations(entsql.IndexWhere("recurrence_id IS NULL")),
		index.Fields("calendar_event_id", "calendar_attendee_id", "event_sequence", "kind", "recurrence_id").
			Unique().
			StorageKey("calendar_invite_deliveries_event_attendee_sequence_kind_recurrence_uidx").
			Annotations(entsql.IndexWhere("recurrence_id IS NOT NULL")),
		index.Fields("next_attempt_at", "id").
			StorageKey("calendar_invite_deliveries_due_idx").
			Annotations(entsql.IndexWhere("state IN ('pending','unknown') AND next_attempt_at IS NOT NULL")),
		index.Fields("lease_until", "id").
			StorageKey("calendar_invite_deliveries_stale_lease_idx").
			Annotations(entsql.IndexWhere("state = 'submitting' AND lease_until IS NOT NULL")),
	}
}
