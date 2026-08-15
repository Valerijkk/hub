package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

var slugRegexp = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// Organization — единственная организация на инстанс (v2.5).
type Organization struct {
	ent.Schema
}

func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.String("slug").
			Unique().
			NotEmpty().
			Match(slugRegexp),
		field.String("invite_token").
			Unique().
			NotEmpty().
			Sensitive().
			Comment("Token for org-wide invite links"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Edges — ADR-ENT-ONDELETE-CASCADE-2026: Org→users/projects = RESTRICT + app purge (A2W8-4-002).
func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type).
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("projects", Project.Type).
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("mail_domains", MailDomain.Type),
		edge.To("mailbox_policy", MailboxPolicy.Type).
			Unique(),
		edge.To("mail_accounts", MailAccount.Type),
	}
}

func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug"),
		index.Fields("invite_token"),
	}
}
