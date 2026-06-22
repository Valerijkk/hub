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
			NotEmpty(),
		field.Bool("pinned").
			Default(false),
		field.Bool("published").
			Default(true),
		field.Int("version").
			Default(1),
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

func (NewsPost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("news_posts").
			Unique().
			Required(),
	}
}

func (NewsPost) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("published", "pinned", "created_at"),
		index.Fields("deleted_at"),
	}
}
