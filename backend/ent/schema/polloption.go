package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PollOption содержит определение схемы для варианта ответа в опросе.
type PollOption struct {
	ent.Schema
}

// Fields описывает поля сущности PollOption.
func (PollOption) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("poll_id", uuid.UUID{}).
			Comment("ID опроса"),
		field.String("text").
			NotEmpty().
			MaxLen(200).
			Comment("Текст варианта ответа"),
		field.Int("position").
			Default(0).
			Comment("Порядковый номер варианта"),
	}
}

// Edges описывает связи сущности PollOption.
// OnDelete: ADR-ENT-ONDELETE-CASCADE-2026 — votes CASCADE (A2W8-4-006).
func (PollOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("poll", Poll.Type).
			Ref("options").
			Field("poll_id").
			Unique().
			Required(),
		edge.To("votes", PollVote.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes описывает индексы сущности PollOption.
func (PollOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("poll_id", "position"),
	}
}
