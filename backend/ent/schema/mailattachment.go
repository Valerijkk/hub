package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailAttachment struct {
	ent.Schema
}

func (MailAttachment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_message_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("mail_draft_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("pending_mail_draft_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("draft_client_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.UUID("storage_object_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.String("engine_blob_id").Optional().Nillable().Validate(validateMailMessageOpaqueID).Immutable(),
		field.String("content_id").Optional().Nillable().Validate(validateMailDeliveryContentID).Immutable(),
		field.String("filename").Validate(validateMailDeliveryFilename).Immutable(),
		field.String("media_type").Validate(validateMailDeliveryMediaType).Immutable(),
		field.Int64("size_bytes").Min(0).Max(26_214_400).Immutable(),
		field.Enum("disposition").Values("attachment", "inline").Immutable(),
		field.Enum("scan_status").Values("pending", "clean", "rejected", "error"),
		field.Bytes("sha256").Optional().Nillable().MinLen(32).MaxLen(32),
		field.UUID("scan_lease_token", uuid.UUID{}).Optional().Nillable(),
		field.Int64("scan_lease_generation").Default(0).NonNegative(),
		field.Time("scan_lease_until").Optional().Nillable(),
		field.Int32("scan_attempts").Default(0).Min(0).Max(10),
		field.Time("next_scan_at").Optional().Nillable(),
		field.Time("scan_terminal_at").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailAttachment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_attachments_text":                  "COALESCE((engine_blob_id IS NULL OR octet_length(engine_blob_id) BETWEEN 1 AND 1024) AND (content_id IS NULL OR (char_length(content_id) BETWEEN 1 AND 255 AND octet_length(content_id) BETWEEN 1 AND 1020)) AND char_length(filename) BETWEEN 1 AND 255 AND octet_length(filename) BETWEEN 1 AND 1020 AND char_length(media_type) BETWEEN 1 AND 255 AND octet_length(media_type) BETWEEN 1 AND 1020 AND size_bytes BETWEEN 0 AND 26214400 AND (sha256 IS NULL OR octet_length(sha256) = 32), FALSE)",
			"mail_attachments_backing_xor":           "COALESCE((storage_object_id IS NOT NULL) <> (engine_blob_id IS NOT NULL), FALSE)",
			"mail_attachments_owner_exclusive":       "COALESCE(NOT (mail_message_id IS NOT NULL AND (mail_draft_id IS NOT NULL OR pending_mail_draft_id IS NOT NULL)) AND (mail_draft_id IS NULL OR pending_mail_draft_id IS NULL OR mail_draft_id = pending_mail_draft_id), FALSE)",
			"mail_attachments_storage_shape":         "COALESCE(storage_object_id IS NULL OR (draft_client_id IS NOT NULL AND size_bytes >= 1), FALSE)",
			"mail_attachments_pending_storage_shape": "COALESCE(pending_mail_draft_id IS NULL OR (storage_object_id IS NOT NULL AND draft_client_id IS NOT NULL AND size_bytes >= 1), FALSE)",
			"mail_attachments_engine_shape":          "COALESCE(engine_blob_id IS NULL OR (mail_message_id IS NOT NULL AND mail_draft_id IS NULL AND pending_mail_draft_id IS NULL AND draft_client_id IS NULL), FALSE)",
			"mail_attachments_inline_content":        "COALESCE(disposition <> 'inline' OR content_id IS NOT NULL, FALSE)",
			"mail_attachments_disposition":           "COALESCE(disposition IN ('attachment','inline'), FALSE)",
			"mail_attachments_scan_enum":             "COALESCE(scan_status IN ('pending','clean','rejected','error'), FALSE)",
			"mail_attachments_hash_scan":             "COALESCE(scan_status NOT IN ('clean','rejected') OR sha256 IS NOT NULL, FALSE)",
			"mail_attachments_scan_lease_pair":       "COALESCE((scan_lease_token IS NULL) = (scan_lease_until IS NULL), FALSE)",
			"mail_attachments_scan_retry_bounds":     "COALESCE(scan_lease_generation >= 0 AND scan_attempts BETWEEN 0 AND 10, FALSE)",
			"mail_attachments_scan_schedule":         "COALESCE((storage_object_id IS NULL AND scan_lease_token IS NULL AND scan_lease_generation = 0 AND scan_lease_until IS NULL AND scan_attempts = 0 AND next_scan_at IS NULL AND scan_terminal_at IS NULL) OR (storage_object_id IS NOT NULL AND ((scan_status = 'pending' AND sha256 IS NULL AND scan_lease_token IS NULL AND scan_lease_generation = 0 AND scan_lease_until IS NULL AND scan_attempts = 0 AND next_scan_at IS NULL AND scan_terminal_at IS NULL) OR (scan_status = 'pending' AND sha256 IS NOT NULL AND scan_terminal_at IS NULL AND next_scan_at IS NOT NULL AND ((scan_attempts BETWEEN 0 AND 9 AND (scan_lease_token IS NULL OR scan_lease_generation >= 1)) OR (scan_attempts = 10 AND scan_lease_token IS NOT NULL AND scan_lease_generation >= 1))) OR (scan_status IN ('clean','rejected') AND sha256 IS NOT NULL AND scan_lease_token IS NULL AND scan_lease_until IS NULL AND next_scan_at IS NULL AND scan_terminal_at IS NOT NULL AND scan_attempts BETWEEN 1 AND 10 AND scan_lease_generation >= 1) OR (scan_status = 'error' AND sha256 IS NOT NULL AND scan_lease_token IS NULL AND scan_lease_until IS NULL AND next_scan_at IS NULL AND scan_terminal_at IS NOT NULL AND scan_attempts = 10 AND scan_lease_generation >= 1))), FALSE)",
		}),
	}
}

func (MailAttachment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("storage_object", MailObject.Type).Ref("attachment").Field("storage_object_id").Unique().Immutable().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("message", MailMessage.Type).Ref("attachments").Field("mail_message_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("draft", MailDraft.Type).Ref("attachments").Field("mail_draft_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("pending_draft", MailDraft.Type).Ref("pending_attachments").Field("pending_mail_draft_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
	}
}

func (MailAttachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("storage_object_id").Unique().StorageKey("mail_attachments_storage_object_uidx").Annotations(entsql.IndexWhere("storage_object_id IS NOT NULL")),
		index.Fields("mail_message_id").StorageKey("mail_attachments_message_idx"),
		index.Fields("mail_draft_id").StorageKey("mail_attachments_draft_idx"),
		index.Fields("pending_mail_draft_id").StorageKey("mail_attachments_pending_draft_idx"),
		index.Fields("next_scan_at", "id").StorageKey("mail_attachments_scan_due_idx").Annotations(entsql.IndexWhere("storage_object_id IS NOT NULL AND scan_status = 'pending' AND scan_terminal_at IS NULL AND next_scan_at IS NOT NULL")),
	}
}

func validateMailDeliveryContentID(value string) error {
	return validateMailDeliveryText(value, 1, 255, 1020)
}
