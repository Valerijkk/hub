package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ScheduledCall С…СЂР°РЅРёС‚ РјРµС‚Р°РґР°РЅРЅС‹Рµ Р·Р°РїР»Р°РЅРёСЂРѕРІР°РЅРЅРѕРіРѕ РёР»Рё РїРѕРІС‚РѕСЂСЏСЋС‰РµРіРѕСЃСЏ Р·РІРѕРЅРєР°.
// РџРѕРґРґРµСЂР¶РёРІР°РµС‚ RFC 5545 RRULE РґР»СЏ РїРѕРІС‚РѕСЂСЏСЋС‰РёС…СЃСЏ Р·РІРѕРЅРєРѕРІ (РµР¶РµРґРЅРµРІРЅС‹Р№ СЃС‚РµРЅРґР°Рї Рё С‚.Рґ.).
type ScheduledCall struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё ScheduledCall.
func (ScheduledCall) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(255).
			Comment("РќР°Р·РІР°РЅРёРµ Р·Р°РїР»Р°РЅРёСЂРѕРІР°РЅРЅРѕРіРѕ Р·РІРѕРЅРєР°"),
		field.String("description").
			Optional().
			MaxLen(2000).
			Comment("РћРїРёСЃР°РЅРёРµ Р·РІРѕРЅРєР°"),
		field.Time("scheduled_at").
			Comment("РЎР»РµРґСѓСЋС‰РµРµ Р·Р°РїР»Р°РЅРёСЂРѕРІР°РЅРЅРѕРµ РІСЂРµРјСЏ Р·РІРѕРЅРєР°"),
		field.Int("duration_minutes").
			Optional().
			Default(60).
			Comment("Р”Р»РёС‚РµР»СЊРЅРѕСЃС‚СЊ Р·РІРѕРЅРєР° РІ РјРёРЅСѓС‚Р°С…"),
		field.Int("reminder_minutes").
			Optional().
			Nillable().
			Comment("Р—Р° СЃРєРѕР»СЊРєРѕ РјРёРЅСѓС‚ РґРѕ Р·РІРѕРЅРєР° РѕС‚РїСЂР°РІРёС‚СЊ РЅР°РїРѕРјРёРЅР°РЅРёРµ (null = Р±РµР· РЅР°РїРѕРјРёРЅР°РЅРёСЏ)"),
		field.String("rrule").
			Optional().
			Comment("RFC 5545 RRULE РґР»СЏ РїРѕРІС‚РѕСЂСЏСЋС‰РёС…СЃСЏ Р·РІРѕРЅРєРѕРІ"),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID РєР°РЅР°Р»Р°, РІ РєРѕС‚РѕСЂРѕРј Р±СѓРґРµС‚ Р·РІРѕРЅРѕРє"),
		field.UUID("project_id", uuid.UUID{}).
			Comment("ID РїСЂРѕРµРєС‚Р°"),
		field.UUID("created_by", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, СЃРѕР·РґР°РІС€РµРіРѕ Р·Р°РїР»Р°РЅРёСЂРѕРІР°РЅРЅС‹Р№ Р·РІРѕРЅРѕРє"),
		field.Bool("is_active").
			Default(true).
			Comment("Активен ли запланированный звонок"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Время мягкого удаления (корзина 30 дней, FR-CAL-05 / NFR-11)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё ScheduledCall.
func (ScheduledCall) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("scheduled_calls").
			Field("project_id").
			Unique().
			Required(),
		edge.From("channel", Channel.Type).
			Ref("scheduled_calls").
			Field("channel_id").
			Unique().
			Required(),
		edge.From("creator", User.Type).
			Ref("scheduled_calls").
			Field("created_by").
			Unique().
			Required(),
		edge.To("participants", User.Type).
			Comment("РџСЂРёРіР»Р°С€С‘РЅРЅС‹Рµ СѓС‡Р°СЃС‚РЅРёРєРё Р·РІРѕРЅРєР°"),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё ScheduledCall.
func (ScheduledCall) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "scheduled_at"),
		index.Fields("is_active", "scheduled_at"),
		index.Fields("deleted_at"),
	}
}
