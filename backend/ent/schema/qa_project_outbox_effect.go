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

type QAProjectOutboxEffect struct {
	ent.Schema
}

func (QAProjectOutboxEffect) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("manifest_id", uuid.UUID{}).Immutable(),
		field.Int("ordinal").Immutable(),
		field.UUID("organization_id", uuid.UUID{}).Immutable(),
		field.UUID("owner_user_id", uuid.UUID{}).Immutable(),
		field.UUID("project_id", uuid.UUID{}).Immutable(),
		field.UUID("project_member_id", uuid.UUID{}).Immutable(),
		field.UUID("outbox_event_id", uuid.UUID{}).Immutable(),
		field.Time("registered_at").Default(nowUTC).Immutable(),
		field.Time("outbox_event_deleted_at").Optional().Nillable(),
		field.Time("project_member_deleted_at").Optional().Nillable(),
		field.Time("project_deleted_at").Optional().Nillable(),
	}
}

func (QAProjectOutboxEffect) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"qa_project_outbox_effects_uuid_authority_nonzero": "COALESCE(id <> '00000000-0000-0000-0000-000000000000'::uuid AND manifest_id <> '00000000-0000-0000-0000-000000000000'::uuid AND organization_id <> '00000000-0000-0000-0000-000000000000'::uuid AND owner_user_id <> '00000000-0000-0000-0000-000000000000'::uuid AND project_id <> '00000000-0000-0000-0000-000000000000'::uuid AND project_member_id <> '00000000-0000-0000-0000-000000000000'::uuid AND outbox_event_id <> '00000000-0000-0000-0000-000000000000'::uuid, FALSE)",
		}),
	}
}

func (QAProjectOutboxEffect) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("manifest", QAProjectOutboxManifest.Type).Ref("effects").Field("manifest_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.NoAction)),
	}
}

func (QAProjectOutboxEffect) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("manifest_id", "ordinal").Unique().StorageKey("qa_project_outbox_effects_manifest_ordinal_key"),
		index.Fields("manifest_id", "project_id").Unique().StorageKey("qa_project_outbox_effects_manifest_project_key"),
		index.Fields("manifest_id", "project_member_id").Unique().StorageKey("qa_project_outbox_effects_manifest_project_member_key"),
		index.Fields("manifest_id", "outbox_event_id").Unique().StorageKey("qa_project_outbox_effects_manifest_outbox_event_key"),
	}
}
