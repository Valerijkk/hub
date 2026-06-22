package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CoffeePair — пара участников Random Coffee в раунде.
type CoffeePair struct {
	ent.Schema
}

func (CoffeePair) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

func (CoffeePair) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("round", CoffeeRound.Type).
			Ref("pairs").
			Unique().
			Required(),
		edge.From("participant_a", User.Type).
			Ref("coffee_pairs_a").
			Unique().
			Required(),
		edge.From("participant_b", User.Type).
			Ref("coffee_pairs_b").
			Unique().
			Required(),
	}
}

func (CoffeePair) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("round", "participant_a", "participant_b").Unique(),
	}
}
