package schema

import (
	"errors"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type MailSignature struct {
	ent.Schema
}

func (MailSignature) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).
			Immutable(),
		field.String("name").
			MinLen(1).
			MaxLen(320).
			Validate(validateMailSignatureName),
		field.Enum("body_format").
			Values("rich", "plain"),
		field.String("rich_body").
			Optional().
			Nillable().
			MaxLen(400_000).
			Validate(validateMailSignatureBody),
		field.String("plain_body").
			MaxLen(400_000).
			Validate(validateMailSignatureBody),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (MailSignature) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

func (MailSignature) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_signatures_body_format":       "((body_format = 'rich' AND rich_body IS NOT NULL) OR (body_format = 'plain' AND rich_body IS NULL))",
			"mail_signatures_name_bounds":       "(char_length(name) BETWEEN 1 AND 80 AND octet_length(name) BETWEEN 1 AND 320)",
			"mail_signatures_plain_body_length": "(char_length(plain_body) <= 100000 AND octet_length(plain_body) <= 400000)",
			"mail_signatures_rich_body_length":  "(rich_body IS NULL OR (char_length(rich_body) <= 100000 AND octet_length(rich_body) <= 400000))",
		}),
	}
}

func (MailSignature) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).
			Ref("signatures").
			Field("mail_account_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("default_for_settings", MailUserSettings.Type).
			Unique(),
	}
}

func validateMailSignatureName(value string) error {
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > 320 || utf8.RuneCountInString(value) > 80 {
		return errors.New("invalid mail signature name")
	}
	return nil
}

func validateMailSignatureBody(value string) error {
	if !utf8.ValidString(value) || len(value) > 400_000 || utf8.RuneCountInString(value) > 100_000 {
		return errors.New("invalid mail signature body")
	}
	return nil
}
