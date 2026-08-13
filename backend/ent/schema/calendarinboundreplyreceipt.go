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

// CalendarInboundReplyReceipt хранит неизменяемую запись обработки входящего iTIP REPLY.
type CalendarInboundReplyReceipt struct {
	ent.Schema
}

// Fields описывает данные дедупликации входящего iTIP REPLY.
func (CalendarInboundReplyReceipt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("calendar_attendee_id", uuid.UUID{}).
			Immutable(),
		field.Int("event_sequence").
			NonNegative().
			Immutable(),
		field.Uint8("payload_hash_version").
			Min(1).
			Immutable(),
		field.Bytes("payload_hash").
			MinLen(sha256.Size).
			MaxLen(sha256.Size).
			Validate(func(hash []byte) error {
				if len(hash) != sha256.Size {
					return errors.New("хеш полезной нагрузки должен иметь длину SHA-256")
				}
				return nil
			}).
			Immutable(),
		field.Enum("outcome").
			Values("applied", "observed_no_change").
			Immutable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Annotations задаёт ограничения точной длины хеша входящего REPLY.
func (CalendarInboundReplyReceipt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_inbound_reply_receipts_payload_hash_length": "octet_length(payload_hash) = 32",
		}),
	}
}

// Edges описывает обязательную связь записи с участником календарного события.
func (CalendarInboundReplyReceipt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("attendee", CalendarAttendee.Type).
			Ref("inbound_reply_receipts").
			Field("calendar_attendee_id").
			Unique().
			Required(),
	}
}

// Indexes описывает уникальность записи дедупликации для участника и последовательности события.
func (CalendarInboundReplyReceipt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("calendar_attendee_id", "event_sequence", "payload_hash_version", "payload_hash").
			Unique(),
	}
}
