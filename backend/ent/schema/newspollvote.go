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

// NewsPollVote — один голос пользователя в news-опросе (NFR-31).
type NewsPollVote struct {
	ent.Schema
}

func (NewsPollVote) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "news_poll_votes"},
	}
}

func (NewsPollVote) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("poll_id", uuid.UUID{}),
		field.UUID("option_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

func (NewsPollVote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("poll", NewsPoll.Type).
			Ref("votes").
			Field("poll_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("option", NewsPollOption.Type).
			Ref("votes").
			Field("option_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("user", User.Type).
			Ref("news_poll_votes").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (NewsPollVote) Indexes() []ent.Index {
	return []ent.Index{
		// NFR-31: ровно один голос на пользователя в опросе.
		index.Fields("poll_id", "user_id").Unique(),
		index.Fields("option_id"),
	}
}
