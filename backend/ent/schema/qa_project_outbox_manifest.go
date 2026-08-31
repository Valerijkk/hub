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

type QAProjectOutboxManifest struct {
	ent.Schema
}

func (QAProjectOutboxManifest) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.Enum("scenario").Values("project-outbox-v1").Immutable(),
		field.Bytes("nonce_hash").MaxLen(sha256.Size).Validate(validateQAProjectOutboxSHA256).Immutable(),
		field.Bytes("control_capability_hash").MaxLen(sha256.Size).Validate(validateQAProjectOutboxSHA256).Immutable(),
		field.Int64("creation_generation").Immutable(),
		field.UUID("organization_id", uuid.UUID{}).Immutable(),
		field.UUID("owner_user_id", uuid.UUID{}).Immutable(),
		field.UUID("session_id", uuid.UUID{}).Immutable(),
		field.String("consumer_stream").NotEmpty().Immutable(),
		field.String("consumer_durable_name").NotEmpty().Immutable(),
		field.String("consumer_filter_subject").NotEmpty().Immutable(),
		field.Bytes("consumer_config_fingerprint").MaxLen(sha256.Size).Validate(validateQAProjectOutboxSHA256).Immutable(),
		field.Enum("state").Values("creating", "ready", "project_registered", "observer_disconnected", "receipt_cached", "quiescing", "deleting", "deleted", "cleanup_failed").Default("creating"),
		field.UUID("project_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("created_by_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("cached_at").Optional().Nillable(),
		field.Time("consumer_deleted_at").Optional().Nillable(),
		field.Time("session_ended_at").Optional().Nillable(),
		field.Time("owner_user_deleted_at").Optional().Nillable(),
		field.Time("organization_deleted_at").Optional().Nillable(),
		field.String("cleanup_error_class").Optional().Nillable().MaxLen(128),
		field.Time("tombstoned_at").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (QAProjectOutboxManifest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"qa_project_outbox_manifests_hash_lengths": "COALESCE(octet_length(nonce_hash) = 32 AND octet_length(control_capability_hash) = 32 AND octet_length(consumer_config_fingerprint) = 32, FALSE)",
			"qa_project_outbox_manifests_uuid_authority_nonzero": "COALESCE(id <> '00000000-0000-0000-0000-000000000000'::uuid AND organization_id <> '00000000-0000-0000-0000-000000000000'::uuid AND owner_user_id <> '00000000-0000-0000-0000-000000000000'::uuid AND session_id <> '00000000-0000-0000-0000-000000000000'::uuid AND (project_id IS NULL OR project_id <> '00000000-0000-0000-0000-000000000000'::uuid) AND (created_by_id IS NULL OR created_by_id <> '00000000-0000-0000-0000-000000000000'::uuid), FALSE)",
			"qa_project_outbox_manifests_state_tombstone": "COALESCE((((state)::text = ANY (ARRAY[('creating'::character varying)::text, ('ready'::character varying)::text, ('project_registered'::character varying)::text, ('observer_disconnected'::character varying)::text, ('receipt_cached'::character varying)::text, ('quiescing'::character varying)::text, ('deleting'::character varying)::text, ('deleted'::character varying)::text, ('cleanup_failed'::character varying)::text])) AND (((state)::text = 'deleted'::text) = (tombstoned_at IS NOT NULL))), false)",
			"qa_project_outbox_manifests_cached_receipt_triple": "COALESCE((project_id IS NULL AND created_by_id IS NULL AND cached_at IS NULL) OR (project_id IS NOT NULL AND created_by_id IS NOT NULL AND cached_at IS NOT NULL), FALSE)",
		}),
	}
}

func (QAProjectOutboxManifest) Edges() []ent.Edge {
	return []ent.Edge{edge.To("effects", QAProjectOutboxEffect.Type)}
}

func (QAProjectOutboxManifest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("nonce_hash").Unique().StorageKey("qa_project_outbox_manifests_nonce_hash_key"),
		index.Fields("organization_id").Unique().StorageKey("qa_project_outbox_manifests_organization_id_key"),
		index.Fields("owner_user_id").Unique().StorageKey("qa_project_outbox_manifests_owner_user_id_key"),
		index.Fields("session_id").Unique().StorageKey("qa_project_outbox_manifests_session_id_key"),
	}
}

func validateQAProjectOutboxSHA256(value []byte) error {
	if len(value) != sha256.Size {
		return errors.New("SHA-256 value must contain exactly 32 bytes")
	}
	return nil
}
