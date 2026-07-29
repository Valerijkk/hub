package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CoffeeRound — раунд подбора пар Random Coffee (V4-05).
type CoffeeRound struct {
	ent.Schema
}

func (CoffeeRound) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("pair_count").
			Default(0),
		field.Int("unpaired_count").
			Default(0),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

func (CoffeeRound) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("triggered_by", User.Type).
			Ref("coffee_rounds_triggered").
			Unique().
			Required(),
		// A3W8-2-002: owned_child Cascade (ADR-ENT-ONDELETE-CASCADE).
		edge.To("pairs", CoffeePair.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (CoffeeRound) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("created_at"),
	}
}
