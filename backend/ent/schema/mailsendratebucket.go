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

type MailSendRateBucket struct {
	ent.Schema
}

func (MailSendRateBucket) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("organization_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.Enum("scope").Values("organization_instance", "user").Immutable(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.Int32("tokens").Min(0).Max(1_000),
		field.Time("last_refill_at"),
		field.Int64("refill_remainder_nanos").Default(0).Min(0).Max(59_999_999_999),
		field.Int64("version").Default(1).Min(1),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailSendRateBucket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_send_rate_buckets_organization_nonzero": "COALESCE(organization_id IS NULL OR organization_id <> '00000000-0000-0000-0000-000000000000'::uuid, FALSE)",
			"mail_send_rate_buckets_user_nonzero":         "COALESCE(user_id IS NULL OR user_id <> '00000000-0000-0000-0000-000000000000'::uuid, FALSE)",
			"mail_send_rate_buckets_scope_owner":          "COALESCE((scope = 'organization_instance' AND organization_id IS NOT NULL AND user_id IS NULL) OR (scope = 'user' AND organization_id IS NULL AND user_id IS NOT NULL), FALSE)",
			"mail_send_rate_buckets_tokens":               "COALESCE(tokens BETWEEN 0 AND 1000, FALSE)",
			"mail_send_rate_buckets_refill_remainder":     "COALESCE(refill_remainder_nanos BETWEEN 0 AND 59999999999, FALSE)",
			"mail_send_rate_buckets_version":              "COALESCE(version >= 1, FALSE)",
		}),
	}
}

func (MailSendRateBucket) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).Ref("mail_send_rate_buckets").Field("organization_id").Unique().Immutable().Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("user", User.Type).Ref("mail_send_rate_buckets").Field("user_id").Unique().Immutable().Annotations(entsql.OnDelete(entsql.NoAction)),
	}
}

func (MailSendRateBucket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id").Unique().StorageKey("mail_send_rate_buckets_organization_instance_uidx").Annotations(entsql.IndexWhere("scope = 'organization_instance' AND organization_id IS NOT NULL")),
		index.Fields("user_id").Unique().StorageKey("mail_send_rate_buckets_user_uidx").Annotations(entsql.IndexWhere("scope = 'user' AND user_id IS NOT NULL")),
	}
}
