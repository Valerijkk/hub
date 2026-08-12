package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NewsPost — корпоративная новость (MOD-INTRA, Hub One v4).
type NewsPost struct {
	ent.Schema
}

func (NewsPost) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(200),
		field.Text("body").
			NotEmpty().
			MaxLen(50000).
			Comment("News body plain/markdown; MaxLen app-level DoS guard (TEXT column, no varchar migrate)"),
		field.Bool("pinned").
			Default(false),
		field.Bool("published").
			Default(true),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (NewsPost) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


func (NewsPost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("news_posts").
			Unique().
			Required(),
		// FR-INTR-01: 0..1 опрос; soft-hide через deleted_at поста (ADR-INTRANET-U03).
		edge.To("poll", NewsPoll.Type).
			Unique(),
	}
}

func (NewsPost) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("published", "pinned", "created_at"),
		index.Fields("deleted_at"),
	}
}
