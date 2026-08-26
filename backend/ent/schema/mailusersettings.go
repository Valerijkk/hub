package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailUserSettings struct {
	ent.Schema
}

func (MailUserSettings) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).
			Immutable(),
		field.UUID("default_identity_id", uuid.UUID{}),
		field.UUID("default_signature_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Enum("remote_images").
			Values("blocked", "trusted_senders").
			Default("blocked"),
		field.Enum("compose_format").
			Values("rich", "plain").
			Default("rich"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

func (MailUserSettings) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

func (MailUserSettings) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).
			Ref("user_settings").
			Field("mail_account_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("default_identity", MailIdentity.Type).
			Ref("default_for_settings").
			Field("default_identity_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("default_signature", MailSignature.Type).
			Ref("default_for_settings").
			Field("default_signature_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.Restrict)),
	}
}

func (MailUserSettings) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id").
			Unique(),
	}
}
