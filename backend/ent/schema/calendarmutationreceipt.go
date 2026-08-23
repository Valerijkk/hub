package schema

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

const calendarEventResponseMaxSerializedUTF8Bytes = 1_048_576

// CalendarMutationReceipt хранит ограниченный replay-ответ календарной идемпотентной мутации.
type CalendarMutationReceipt struct {
	ent.Schema
}

// Fields описывает идентичность запроса, безопасный ответ и срок хранения receipt.
func (CalendarMutationReceipt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("actor_user_id", uuid.UUID{}),
		field.Enum("operation").
			Values("event_create", "legacy_call_create", "rsvp_series", "rsvp_occurrence", "meeting_create", "meeting_update", "meeting_cancel").
			Immutable(),
		field.Bytes("idempotency_key_hash").
			MinLen(sha256.Size).
			MaxLen(sha256.Size).
			Validate(func(hash []byte) error {
				if len(hash) != sha256.Size {
					return errors.New("хеш ключа идемпотентности должен иметь длину SHA-256")
				}
				return nil
			}).
			Immutable(),
		field.Bytes("request_hash").
			MinLen(sha256.Size).
			MaxLen(sha256.Size).
			Validate(func(hash []byte) error {
				if len(hash) != sha256.Size {
					return errors.New("хеш запроса должен иметь длину SHA-256")
				}
				return nil
			}).
			Immutable(),
		field.Enum("state").
			Values("in_progress", "completed").
			Default("in_progress"),
		field.Int("response_status").
			Optional().
			Nillable().
			Min(100).
			Max(599),
		field.Bytes("response_json").
			Optional().
			Nillable().
			MaxLen(calendarEventResponseMaxSerializedUTF8Bytes).
			Validate(func(response []byte) error {
				if !utf8.Valid(response) || !json.Valid(response) {
					return errors.New("ответ календарной мутации должен содержать корректный JSON")
				}
				return nil
			}),
		field.UUID("calendar_event_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("expires_at"),
	}
}

// Annotations задаёт переносимые проверки размера хешей и replay-ответа.
func (CalendarMutationReceipt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_mutation_receipts_idempotency_key_hash_length": "octet_length(idempotency_key_hash) = 32",
			"calendar_mutation_receipts_request_hash_length":         "octet_length(request_hash) = 32",
			"calendar_mutation_receipts_response_json_length":        "octet_length(response_json) <= 1048576",
			"calendar_mutation_receipts_response_status_range":       "response_status IS NULL OR (response_status >= 100 AND response_status <= 599)",
		}),
	}
}

// Edges описывает контекст актёра и необязательную ссылку на итоговое событие.
func (CalendarMutationReceipt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("actor", User.Type).
			Ref("calendar_mutation_receipts").
			Field("actor_user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("event", CalendarEvent.Type).
			Ref("mutation_receipts").
			Field("calendar_event_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

// Indexes фиксирует область идемпотентности аутентифицированного актора.
func (CalendarMutationReceipt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "actor_user_id", "operation", "idempotency_key_hash").
			Unique(),
	}
}
