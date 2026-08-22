package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CalendarReplyDelivery хранит неизменяемое намерение внешней доставки RSVP.
type CalendarReplyDelivery struct {
	ent.Schema
}

// Fields описывает ревизию RSVP, состояние доставки и безопасную retry-метаинформацию.
func (CalendarReplyDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("calendar_event_id", uuid.UUID{}).
			Immutable(),
		field.UUID("calendar_attendee_id", uuid.UUID{}).
			Immutable(),
		field.Int("event_sequence").
			NonNegative().
			Immutable(),
		field.Int("attendee_version").
			Min(1).
			Immutable(),
		field.String("recurrence_id").
			Optional().
			Nillable().
			MaxLen(255).
			Immutable(),
		field.Enum("rsvp_status").
			Values("accepted", "declined", "tentative").
			Immutable(),
		field.Enum("state").
			Values("pending", "submitting", "unknown", "sent", "failed_permanent").
			Default("pending"),
		field.UUID("mail_submission_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("attempts").
			Default(0).
			NonNegative(),
		field.Time("next_attempt_at").
			Optional().
			Nillable(),
		field.Enum("last_error_code").
			Values("mailbox_unavailable", "engine_unavailable", "submission_rejected", "submission_outcome_unknown", "invalid_recipient", "policy_denied", "retry_exhausted", "internal_error").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges связывает intent с Calendar и запрещает каскадное удаление audit-истории.
func (CalendarReplyDelivery) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", CalendarEvent.Type).
			Field("calendar_event_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("attendee", CalendarAttendee.Type).
			Field("calendar_attendee_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.NoAction)),
	}
}

// Indexes задаёт одно неизменяемое намерение на ревизию RSVP участника.
func (CalendarReplyDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("calendar_attendee_id", "event_sequence", "attendee_version").
			Unique(),
	}
}
