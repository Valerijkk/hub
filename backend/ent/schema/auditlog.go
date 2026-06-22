package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AuditLog С…СЂР°РЅРёС‚ Р·Р°РїРёСЃРё Р¶СѓСЂРЅР°Р»Р° Р°СѓРґРёС‚Р° РІСЃРµС… РґРµР№СЃС‚РІРёР№ РІ СЃРёСЃС‚РµРјРµ.
type AuditLog struct {
	ent.Schema
}

// Поля сущности AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("action").
			NotEmpty().
			Comment("Р”РµР№СЃС‚РІРёРµ: login, logout, create, update, delete"),
		field.String("entity_type").
			NotEmpty().
			Comment("РўРёРї СЃСѓС‰РЅРѕСЃС‚Рё: issue, project, test_case, channel, file, user"),
		field.String("entity_id").
			Optional().
			Comment("ID СЃСѓС‰РЅРѕСЃС‚Рё (UUID РёР»Рё РєР»СЋС‡)"),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, РІС‹РїРѕР»РЅРёРІС€РµРіРѕ РґРµР№СЃС‚РІРёРµ"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РїСЂРѕРµРєС‚Р° (РµСЃР»Рё РґРµР№СЃС‚РІРёРµ РїСЂРёРІСЏР·Р°РЅРѕ Рє РїСЂРѕРµРєС‚Сѓ)"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("Р”РѕРїРѕР»РЅРёС‚РµР»СЊРЅС‹Рµ РґР°РЅРЅС‹Рµ Рѕ РґРµР№СЃС‚РІРёРё РІ JSON"),
		field.String("ip_address").
			Optional().
			Comment("IP-Р°РґСЂРµСЃ РєР»РёРµРЅС‚Р°"),
		field.String("user_agent").
			Optional().
			Comment("User-Agent РєР»РёРµРЅС‚Р°"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Индексы сущности AuditLog.
func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action"),
		index.Fields("entity_type"),
		index.Fields("user_id"),
		index.Fields("project_id"),
		index.Fields("created_at"),
	}
}
