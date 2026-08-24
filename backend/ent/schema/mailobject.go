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

type MailObject struct {
	ent.Schema
}

func (MailObject) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("deletion_id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("organization_id", uuid.UUID{}).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Immutable(),
		field.UUID("uploader_user_id", uuid.UUID{}).Immutable(),
		field.UUID("draft_client_id", uuid.UUID{}).Immutable(),
		field.String("bucket").Validate(validateMailObjectBucket).Immutable(),
		field.String("object_key").Validate(validateMailObjectKey).Immutable(),
		field.Int64("expected_size_bytes").Min(1).Max(26_214_400).Immutable(),
		field.Int64("actual_size_bytes").Optional().Nillable().Min(1).Max(26_214_400),
		field.String("filename").Validate(validateMailDeliveryFilename).Immutable(),
		field.String("media_type").Validate(validateMailDeliveryMediaType).Immutable(),
		field.Bytes("sha256").Optional().Nillable().MinLen(32).MaxLen(32),
		field.Enum("upload_state").Values("pending", "confirmed", "orphaned", "deleted").Default("pending"),
		field.Time("expires_at").Immutable(),
		field.Time("upload_drain_until").Immutable(),
		field.UUID("delete_lease_token", uuid.UUID{}).Optional().Nillable(),
		field.Int64("delete_lease_generation").Default(0).NonNegative(),
		field.Time("delete_lease_until").Optional().Nillable(),
		field.Int32("delete_attempts").Default(0).Min(0).Max(10),
		field.Time("next_delete_at").Optional().Nillable(),
		field.Time("delete_terminal_at").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailObject) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_objects_text":                  "COALESCE(char_length(bucket) BETWEEN 1 AND 63 AND octet_length(bucket) BETWEEN 1 AND 63 AND char_length(object_key) BETWEEN 1 AND 512 AND octet_length(object_key) BETWEEN 1 AND 512 AND char_length(filename) BETWEEN 1 AND 255 AND octet_length(filename) BETWEEN 1 AND 1020 AND char_length(media_type) BETWEEN 1 AND 255 AND octet_length(media_type) BETWEEN 1 AND 1020, FALSE)",
			"mail_objects_sizes":                 "COALESCE(expected_size_bytes BETWEEN 1 AND 26214400 AND (actual_size_bytes IS NULL OR actual_size_bytes BETWEEN 1 AND 26214400), FALSE)",
			"mail_objects_actual_sha_pair":       "COALESCE((actual_size_bytes IS NULL) = (sha256 IS NULL) AND (sha256 IS NULL OR octet_length(sha256) = 32), FALSE)",
			"mail_objects_upload_enum":           "COALESCE(upload_state IN ('pending','confirmed','orphaned','deleted'), FALSE)",
			"mail_objects_delete_lease_pair":     "COALESCE((delete_lease_token IS NULL) = (delete_lease_until IS NULL), FALSE)",
			"mail_objects_delete_retry_bounds":   "COALESCE(delete_lease_generation >= 0 AND delete_attempts BETWEEN 0 AND 10, FALSE)",
			"mail_objects_upload_state":          "COALESCE((upload_state = 'pending' AND actual_size_bytes IS NULL AND sha256 IS NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL AND delete_lease_generation = 0 AND delete_terminal_at IS NULL AND delete_attempts = 0) OR (upload_state = 'confirmed' AND actual_size_bytes = expected_size_bytes AND sha256 IS NOT NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL AND delete_lease_generation = 0 AND delete_terminal_at IS NULL AND delete_attempts = 0) OR (upload_state IN ('orphaned','deleted') AND ((delete_terminal_at IS NOT NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL AND delete_attempts = 10) OR (delete_terminal_at IS NULL AND next_delete_at IS NOT NULL AND (delete_lease_token IS NULL OR delete_lease_generation >= 1)))), FALSE)",
			"mail_objects_upload_drain_boundary": "COALESCE(upload_drain_until = expires_at + INTERVAL '11 minutes' AND (next_delete_at IS NULL OR next_delete_at >= upload_drain_until), FALSE)",
		}),
	}
}

func (MailObject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).Ref("mail_objects").Field("organization_id").Unique().Required().Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("account", MailAccount.Type).Ref("objects").Field("mail_account_id").Unique().Required().Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("uploader", User.Type).Ref("mail_objects").Field("uploader_user_id").Unique().Required().Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("attachment", MailAttachment.Type).Unique(),
	}
}

func (MailObject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("bucket", "object_key").Unique().StorageKey("mail_objects_bucket_object_key_uidx"),
		index.Fields("next_delete_at", "id").StorageKey("mail_objects_delete_due_idx").Annotations(entsql.IndexWhere("upload_state IN ('orphaned','deleted') AND delete_terminal_at IS NULL AND next_delete_at IS NOT NULL")),
	}
}

func validateMailObjectBucket(value string) error {
	return validateMailDeliveryText(value, 1, 63, 63)
}

func validateMailObjectKey(value string) error {
	return validateMailDeliveryText(value, 1, 512, 512)
}

func validateMailDeliveryFilename(value string) error {
	return validateMailDeliveryText(value, 1, 255, 1020)
}

func validateMailDeliveryMediaType(value string) error {
	return validateMailDeliveryText(value, 1, 255, 1020)
}

func validateMailDeliveryText(value string, minimumRunes, maximumRunes, maximumBytes int) error {
	if !utf8.ValidString(value) || len(value) > maximumBytes || utf8.RuneCountInString(value) < minimumRunes || utf8.RuneCountInString(value) > maximumRunes {
		return errors.New("invalid mail delivery text")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return errors.New("invalid mail delivery text")
		}
	}
	return nil
}
