package schema

import (
	"crypto/sha256"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CalendarEventException хранит материализованное изменение отдельного экземпляра календарного события.
type CalendarEventException struct {
	ent.Schema
}

// Fields описывает поля отмены или переопределения экземпляра календарного события.
func (CalendarEventException) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("calendar_event_id", uuid.UUID{}),
		field.String("recurrence_id").
			NotEmpty().
			MaxLen(255),
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
		field.String("timezone").
			Optional().
			Nillable().
			MaxLen(64),
		field.String("location").
			Optional().
			Nillable().
			MaxLen(500),
		field.Enum("status").
			Values("confirmed", "tentative", "cancelled").
			Optional().
			Nillable(),
		field.Int("sequence").
			Default(0).
			NonNegative(),
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
	}
}

// Edges описывает обязательную обратную связь с каноническим календарным событием.
func (CalendarEventException) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", CalendarEvent.Type).
			Ref("exceptions").
			Field("calendar_event_id").
			Unique().
			Required(),
	}
}

// Indexes описывает уникальность нормализованного идентификатора экземпляра внутри события.
func (CalendarEventException) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("calendar_event_id", "recurrence_id").
			Unique(),
	}
}
