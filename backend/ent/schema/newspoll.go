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

// NewsPoll — опрос внутри новости интранета (FR-INTR-01; не messenger Poll).
type NewsPoll struct {
	ent.Schema
}

func (NewsPoll) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "news_polls"},
	}
}

func (NewsPoll) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("news_post_id", uuid.UUID{}).
			Comment("FK на NewsPost; 0..1 опрос на пост"),
		field.String("question").
			NotEmpty().
			MaxLen(500),
		field.Bool("closed").
			Default(false),
		field.Int("version").
			Default(1).
			Comment("Optimistic locking (NFR-24)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

func (NewsPoll) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("news_post", NewsPost.Type).
			Ref("poll").
			Field("news_post_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("options", NewsPollOption.Type),
		edge.To("votes", NewsPollVote.Type),
	}
}

func (NewsPoll) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("news_post_id").Unique(),
	}
}
