package schema

import (
	"crypto/sha256"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CalendarInboundObjectState хранит текущий порядок входящего iTIP-объекта.
type CalendarInboundObjectState struct {
	ent.Schema
}

// Fields описывает идентичность, порядок и срок хранения входящего объекта.
func (CalendarInboundObjectState) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}).
			Immutable(),
		field.UUID("calendar_owner_id", uuid.UUID{}).
			Immutable(),
		field.String("uid").
			NotEmpty().
			MaxLen(255).
			Immutable(),
		field.String("recurrence_id").
			Optional().
			Nillable().
			MaxLen(255).
			Immutable(),
		field.String("organizer_email").
			NotEmpty().
			MaxLen(320).
			Immutable(),
		field.Int("sequence").
			NonNegative(),
		field.Enum("method").
			Values("request", "cancel"),
		field.Uint8("payload_hash_version").
			Min(1),
		field.Bytes("payload_hash").
			MinLen(sha256.Size).
			MaxLen(sha256.Size).
			Validate(func(hash []byte) error {
				if len(hash) != sha256.Size {
					return errors.New("хеш полезной нагрузки должен иметь длину SHA-256")
				}
				return nil
			}),
		field.Enum("state").
			Values("applied", "tombstone", "deferred_missing_master"),
		field.UUID("calendar_event_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("expires_at").
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

// Annotations задаёт переносимые PostgreSQL-проверки состояния входящего объекта.
func (CalendarInboundObjectState) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_inbound_object_states_payload_hash_length": "octet_length(payload_hash) = 32",
			"calendar_inbound_object_states_state_expiry":       "(state = 'deferred_missing_master') = (expires_at IS NOT NULL)",
		}),
	}
}

// Edges связывает объект только с принадлежащими Calendar сущностями.
func (CalendarInboundObjectState) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("calendar_owner", User.Type).
			Field("calendar_owner_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("event", CalendarEvent.Type).
			Field("calendar_event_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.From("deferred_message_receipts", CalendarInboundMessageReceipt.Type).
			Ref("deferred_object_state"),
	}
}

// Indexes задаёт отдельные области уникальности для series и occurrence объектов.
func (CalendarInboundObjectState) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("calendar_owner_id", "uid").
			Unique().
			Annotations(entsql.IndexWhere("recurrence_id IS NULL")),
		index.Fields("calendar_owner_id", "uid", "recurrence_id").
			Unique().
			Annotations(entsql.IndexWhere("recurrence_id IS NOT NULL")),
	}
}
