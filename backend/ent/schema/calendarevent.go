package schema

import (
	"crypto/sha256"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CalendarEvent хранит каноническое состояние календарного события.
type CalendarEvent struct {
	ent.Schema
}

// Fields описывает поля канонического календарного события.
func (CalendarEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("calendar_owner_id", uuid.UUID{}),
		field.String("uid").
			NotEmpty().
			MaxLen(255).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(500),
		field.String("description").
			Optional().
			MaxLen(200000),
		field.String("location").
			Optional().
			MaxLen(500),
		field.Time("starts_at").
			Optional().
			Nillable(),
		field.Time("ends_at").
			Optional().
			Nillable(),
		field.Time("start_date").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Time("end_date_exclusive").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Bool("all_day").
			Default(false),
		field.String("timezone").
			NotEmpty().
			MaxLen(64),
		field.String("rrule").
			Optional().
			Nillable().
			MaxLen(1000),
		field.JSON("exdates", []string{}).
			Default([]string{}).
			Validate(func(exdates []string) error {
				if len(exdates) > 1000 {
					return errors.New("число исключённых дат превышает 1000")
				}
				return nil
			}),
		field.Int("sequence").
			Default(0).
			NonNegative(),
		field.Enum("source").
			Values("hub", "mail_import", "ical_feed").
			Default("hub"),
		field.Enum("status").
			Values("confirmed", "tentative", "cancelled").
			Default("confirmed"),
		field.UUID("organizer_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("organizer_email").
			NotEmpty().
			MaxLen(320),
		field.UUID("scheduled_call_id", uuid.UUID{}).
			Optional().
			Nillable().
			Unique().
			Immutable(),
		field.Int("reminder_minutes").
			Optional().
			Nillable().
			Min(0).
			Max(40320),
		field.Uint8("payload_hash_version").
			Optional().
			Nillable().
			Min(1),
		field.Bytes("payload_hash").
			Optional().
			Nillable().
			MaxLen(sha256.Size).
			Validate(func(hash []byte) error {
				if len(hash) != sha256.Size {
					return errors.New("хеш полезной нагрузки должен иметь длину SHA-256")
				}
				return nil
			}),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Mixin подключает optimistic-lock версию календарного события.
func (CalendarEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

// Annotations задаёт ограничения целостности канонического календарного события.
func (CalendarEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_events_payload_hash_pair": "(payload_hash_version IS NULL) = (payload_hash IS NULL)",
		}),
	}
}

// Edges описывает связи календарного события.
func (CalendarEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Unique().
			Required(),
		edge.To("calendar_owner", User.Type).
			Field("calendar_owner_id").
			Unique().
			Required(),
		edge.To("organizer", User.Type).
			Field("organizer_id").
			Unique(),
		edge.To("scheduled_call", ScheduledCall.Type).
			Field("scheduled_call_id").
			Unique().
			Immutable(),
		edge.To("attendees", CalendarAttendee.Type),
		edge.To("exceptions", CalendarEventException.Type),
		edge.To("mutation_receipts", CalendarMutationReceipt.Type),
	}
}

// Indexes описывает индексы календарного события.
func (CalendarEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id"),
		index.Fields("calendar_owner_id"),
		index.Fields("calendar_owner_id", "uid").
			Unique(),
		index.Fields("deleted_at"),
	}
}
