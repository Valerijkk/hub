package schema

import (
	"errors"
	"unicode"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type RecentContact struct {
	ent.Schema
}

func (RecentContact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("normalized_email").
			MinLen(1).
			MaxLen(320).
			Validate(validateNormalizedMailAddress),
		field.String("display_name").
			Optional().
			Nillable().
			MaxLen(640).
			Validate(validateRecentContactDisplayName),
		field.Time("last_used_at"),
		field.Int32("use_count").
			NonNegative(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

func (RecentContact) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"recent_contacts_display_name": "display_name IS NULL OR (char_length(display_name) <= 160 AND octet_length(display_name) <= 640)",
			"recent_contacts_email_bounds": "(char_length(normalized_email) BETWEEN 1 AND 320 AND octet_length(normalized_email) BETWEEN 1 AND 320)",
			"recent_contacts_use_count":     "use_count BETWEEN 0 AND 2147483647",
		}),
	}
}

func (RecentContact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).
			Ref("recent_contacts").
			Field("mail_account_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("user", User.Type).
			Ref("recent_contacts").
			Field("user_id").
			Unique().
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (RecentContact) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "normalized_email").
			Unique(),
		index.Fields("mail_account_id", "last_used_at", "id").
			Annotations(entsql.IndexWhere("user_id IS NULL")),
	}
}

func validateRecentContactDisplayName(value string) error {
	if !utf8.ValidString(value) || len(value) > 640 || utf8.RuneCountInString(value) > 160 {
		return errors.New("invalid recent contact display name")
	}
	for _, character := range value {
		if unicode.In(character, unicode.Cc) || character == '\u2028' || character == '\u2029' {
			return errors.New("invalid recent contact display name")
		}
	}
	return nil
}
