package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// BackupRecord holds the schema definition for the BackupRecord entity.
type BackupRecord struct {
	ent.Schema
}

// Fields of the BackupRecord.
func (BackupRecord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("s3_key").
			NotEmpty().
			Comment("РљР»СЋС‡ РѕР±СЉРµРєС‚Р° РІ S3 (backups/{timestamp}.zip)"),
		field.Int64("file_size").
			Default(0).
			Comment("Р Р°Р·РјРµСЂ Р°СЂС…РёРІР° РІ Р±Р°Р№С‚Р°С…"),
		field.Enum("status").
			Values("in_progress", "completed", "failed").
			Default("in_progress").
			Comment("РўРµРєСѓС‰РёР№ СЃС‚Р°С‚СѓСЃ СЂРµР·РµСЂРІРЅРѕРіРѕ РєРѕРїРёСЂРѕРІР°РЅРёСЏ"),
		field.Text("error_message").
			Optional().
			Nillable().
			Comment("РЎРѕРѕР±С‰РµРЅРёРµ РѕР± РѕС€РёР±РєРµ РїСЂРё РЅРµСѓРґР°С‡Рµ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the BackupRecord.
func (BackupRecord) Edges() []ent.Edge {
	return nil // РЎРёСЃС‚РµРјРЅР°СЏ СЃСѓС‰РЅРѕСЃС‚СЊ Р±РµР· СЃРІСЏР·РµР№
}

// Indexes of the BackupRecord.
func (BackupRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
