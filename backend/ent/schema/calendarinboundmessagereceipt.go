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

// CalendarInboundMessageReceipt хранит лизинговую обработку входящей calendar-части.
type CalendarInboundMessageReceipt struct {
	ent.Schema
}

// Fields описывает безопасную идентичность, состояние и повторные попытки receipt.
func (CalendarInboundMessageReceipt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}).
			Immutable(),
		field.UUID("mailbox_id", uuid.UUID{}).
			Immutable(),
		field.UUID("mail_message_id", uuid.UUID{}).
			Immutable(),
		field.Bytes("part_locator_hash").
			MinLen(sha256.Size).
			MaxLen(sha256.Size).
			Validate(func(hash []byte) error {
				if len(hash) != sha256.Size {
					return errors.New("хеш locator календарной части должен иметь длину SHA-256")
				}
				return nil
			}).
			Immutable(),
		field.UUID("event_envelope_id", uuid.UUID{}).
			Immutable(),
		field.Enum("state").
			Values("pending", "processing", "applied", "ignored_stale", "duplicate", "deferred_missing_master", "quarantined", "dead_letter").
			Default("pending"),
		field.Int("attempts").
			Default(0).
			NonNegative().
			Max(8),
		field.UUID("lease_token", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int64("lease_generation").
			Default(0).
			NonNegative(),
		field.Time("lease_until").
			Optional().
			Nillable(),
		field.Time("next_attempt_at").
			Optional().
			Nillable(),
		field.Enum("error_code").
			Values("message_unavailable", "part_unavailable", "malformed_itip", "unsupported_method", "unsupported_version", "untrusted_sender", "organizer_mismatch", "equal_sequence_conflict", "missing_master_expired", "resource_limit", "retry_exhausted", "engine_unavailable", "database_unavailable", "internal_error").
			Optional().
			Nillable(),
		field.UUID("calendar_event_id", uuid.UUID{}).
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

// Annotations задаёт точную длину хеша locator на уровне PostgreSQL.
func (CalendarInboundMessageReceipt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"calendar_inbound_message_receipts_part_locator_hash_length": "octet_length(part_locator_hash) = 32",
			"calendar_inbound_message_receipts_attempts_bounds":          "attempts >= 0 AND attempts <= 8",
			"calendar_inbound_message_receipts_organization_id_nonzero":  "organization_id <> '00000000-0000-0000-0000-000000000000'::uuid",
			"calendar_inbound_message_receipts_mailbox_id_nonzero":       "mailbox_id <> '00000000-0000-0000-0000-000000000000'::uuid",
			"calendar_inbound_message_receipts_state_shape":              "COALESCE((state = 'pending' AND lease_token IS NULL AND lease_until IS NULL AND ((attempts = 0 AND next_attempt_at IS NULL) OR (attempts > 0 AND next_attempt_at IS NOT NULL))) OR (state = 'processing' AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND lease_generation >= 1 AND attempts >= 1 AND next_attempt_at IS NULL AND error_code IS NULL) OR (state IN ('applied','ignored_stale','duplicate','quarantined','dead_letter') AND lease_token IS NULL AND lease_until IS NULL AND next_attempt_at IS NULL) OR (state = 'deferred_missing_master' AND lease_token IS NULL AND lease_until IS NULL AND next_attempt_at IS NOT NULL), FALSE)",
		}),
	}
}

// Edges хранит только необязательную Calendar-ссылку без Mail Ent-связи.
func (CalendarInboundMessageReceipt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event", CalendarEvent.Type).
			Field("calendar_event_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

// Indexes задаёт дедупликацию сообщения и устойчивого locator-хеша.
func (CalendarInboundMessageReceipt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_message_id", "part_locator_hash").
			Unique(),
		index.Fields("event_envelope_id").
			Unique(),
		index.Fields("state", "next_attempt_at", "id").
			StorageKey("calendar_inbound_message_receipts_due_idx").
			Annotations(entsql.IndexWhere("state IN ('pending','deferred_missing_master') AND next_attempt_at IS NOT NULL")),
	}
}
