package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MarketplaceListing — объявление на внутренней доске (MOD-08).
type MarketplaceListing struct {
	ent.Schema
}

func (MarketplaceListing) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(200),
		field.Text("description").
			NotEmpty().
			MaxLen(2000),
		field.Enum("category").
			Values("equipment", "books", "services", "furniture", "other").
			Default("other"),
		field.Enum("status").
			Values("pending", "approved", "rejected", "sold", "withdrawn").
			Default("pending"),
		field.Int64("price_cents").
			Optional().
			Nillable().
			Comment("Цена в минорных единицах; nil = договорная/бесплатно"),
		field.String("currency").
			Default("RUB").
			MaxLen(3),
		field.Text("review_note").
			Optional(),
		field.Int("version").
			Default(1),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("reviewed_at").
			Optional().
			Nillable(),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (MarketplaceListing) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("marketplace_listings").
			Unique().
			Required(),
		edge.From("reviewer", User.Type).
			Ref("marketplace_reviews").
			Unique(),
	}
}

func (MarketplaceListing) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "category", "created_at"),
		index.Fields("deleted_at"),
	}
}
