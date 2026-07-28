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

// NewsPollOption — вариант ответа news-опроса (FR-INTR-01).
type NewsPollOption struct {
	ent.Schema
}

func (NewsPollOption) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "news_poll_options"},
	}
}

func (NewsPollOption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("poll_id", uuid.UUID{}),
		field.String("text").
			NotEmpty().
			MaxLen(200),
		field.Int("position").
			Default(0),
		field.Int("votes_count").
			Default(0).
			NonNegative().
			Comment("Денормализованный счётчик голосов"),
	}
}

func (NewsPollOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("poll", NewsPoll.Type).
			Ref("options").
			Field("poll_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("votes", NewsPollVote.Type),
	}
}

func (NewsPollOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("poll_id", "position"),
	}
}
