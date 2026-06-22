package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// FileMetadata — определение Ent-схемы сущности FileMetadata.
type FileMetadata struct {
	ent.Schema
}

// Поля сущности FileMetadata.
func (FileMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("filename").
			NotEmpty(),
		field.String("content_type").
			NotEmpty(),
		field.Int64("size"),
		field.String("s3_key").
			NotEmpty(),
		field.String("s3_bucket").
			NotEmpty(),
		field.String("entity_type").
			Optional().
			Comment("task, message, testcase, avatar, etc."),
		field.String("entity_id").
			Optional(),
		field.Enum("status").
			Values("pending", "confirmed", "deleted").
			Default("pending"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ РјСЏРіРєРѕРіРѕ СѓРґР°Р»РµРЅРёСЏ (РєРѕСЂР·РёРЅР° 30 РґРЅРµР№)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности FileMetadata.
func (FileMetadata) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("uploader", User.Type).
			Ref("files").
			Unique(),
		edge.From("project", Project.Type).
			Ref("files").
			Unique(),
	}
}

// Индексы сущности FileMetadata.
func (FileMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entity_type", "entity_id"),
		index.Fields("status"),
		index.Fields("s3_key"),
		index.Fields("deleted_at"),
	}
}
