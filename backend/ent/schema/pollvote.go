package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PollVote содержит определение схемы для голоса пользователя в опросе.
type PollVote struct {
	ent.Schema
}

// Fields описывает поля сущности PollVote.
func (PollVote) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("poll_id", uuid.UUID{}).
			Comment("ID опроса"),
		field.UUID("option_id", uuid.UUID{}).
			Comment("ID выбранного варианта"),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID проголосовавшего пользователя"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges описывает связи сущности PollVote.
func (PollVote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("poll", Poll.Type).
			Ref("votes").
			Field("poll_id").
			Unique().
			Required(),
		edge.From("option", PollOption.Type).
			Ref("votes").
			Field("option_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("poll_votes").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes описывает индексы сущности PollVote.
func (PollVote) Indexes() []ent.Index {
	return []ent.Index{
		// Уникальное ограничение: один голос на вариант от одного пользователя.
		index.Fields("poll_id", "option_id", "user_id").Unique(),
		index.Fields("poll_id", "user_id"),
	}
}
