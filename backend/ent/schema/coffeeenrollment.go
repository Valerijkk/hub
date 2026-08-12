package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CoffeeEnrollment — участие в Random Coffee (V4-05).
type CoffeeEnrollment struct {
	ent.Schema
}

func (CoffeeEnrollment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Bool("is_active").
			Default(true),
		field.Time("enrolled_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

func (CoffeeEnrollment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("coffee_enrollment").
			Unique().
			Required(),
	}
}

func (CoffeeEnrollment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
		// P1-BE-40 / W8-A1-001: один enrollment на user (O2O на уровне БД).
		index.Edges("user").Unique(),
	}
}
