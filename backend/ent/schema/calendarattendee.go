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

// CalendarAttendee хранит состояние приглашённого участника календарного события.
type CalendarAttendee struct {
	ent.Schema
}

// Fields описывает поля участника календарного события.
func (CalendarAttendee) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("calendar_event_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("normalized_email").
			NotEmpty().
			MaxLen(320),
		field.String("display_name").
			Optional().
			Nillable().
			MaxLen(160),
		field.Enum("role").
			Values("organizer", "required", "optional").
			Default("required"),
		field.Enum("rsvp_status").
			Values("needs_action", "accepted", "declined", "tentative").
			Default("needs_action"),
		field.String("recurrence_id").
			Optional().
			Nillable().
			MaxLen(255),
		field.Time("responded_at").
			Optional().
			Nillable(),
		field.Time("removed_at").
			Optional().
			Nillable(),
		field.Enum("reply_delivery_status").
			Values("not_required", "pending", "sent", "failed").
			Default("not_required"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Annotations keeps the internal organizer row canonical. Invitee roles retain
// their existing required/optional shapes and are unaffected by this check.
func (CalendarAttendee) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_attendees_organizer_shape": "role <> 'organizer' OR (user_id IS NOT NULL AND rsvp_status = 'accepted' AND recurrence_id IS NULL AND removed_at IS NULL)",
		}),
	}
}

// Mixin подключает optimistic-lock версию участника календарного события.
func (CalendarAttendee) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

// Edges описывает связи участника календарного события.
func (CalendarAttendee) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", CalendarEvent.Type).
			Ref("attendees").
			Field("calendar_event_id").
			Unique().
			Required(),
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("inbound_reply_receipts", CalendarInboundReplyReceipt.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("reply_deliveries", CalendarReplyDelivery.Type).
			Ref("attendee"),
	}
}

// Indexes описывает частичные уникальные индексы участника календарного события.
func (CalendarAttendee) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("calendar_event_id", "normalized_email").
			Unique().
			Annotations(entsql.IndexWhere("recurrence_id IS NULL AND removed_at IS NULL")),
		index.Fields("calendar_event_id", "role").
			Unique().
			Annotations(entsql.IndexWhere("role = 'organizer' AND recurrence_id IS NULL AND removed_at IS NULL")),
		index.Fields("calendar_event_id", "normalized_email", "recurrence_id").
			Unique().
			Annotations(entsql.IndexWhere("recurrence_id IS NOT NULL AND removed_at IS NULL")),
	}
}
