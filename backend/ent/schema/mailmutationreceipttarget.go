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

type MailMutationReceiptTarget struct {
	ent.Schema
}

func (MailMutationReceiptTarget) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("receipt_id", uuid.UUID{}).Immutable(),
		field.Int32("request_index").Min(0).Max(199).Immutable(),
		field.Enum("target_type").Values("message", "thread").Immutable(),
		field.UUID("target_id", uuid.UUID{}).Immutable(),
		field.Enum("status").Values("succeeded", "partial", "failed", "unknown"),
		field.Int32("affected_count").Min(0).Max(200),
		field.Int32("succeeded_count").Min(0).Max(200),
		field.Int32("failed_count").Min(0).Max(200),
		field.Enum("training_status").Values("not_requested", "accepted", "rejected", "unknown"),
		field.Enum("safe_code").Values("MAIL_ACTION_REJECTED", "MAIL_CLASSIFIER_REJECTED", "MAIL_TARGET_UNAVAILABLE", "MAIL_FOLDER_UNAVAILABLE", "MAIL_LABEL_UNAVAILABLE", "MAIL_ENGINE_OUTCOME_UNKNOWN", "MAIL_CLASSIFIER_OUTCOME_UNKNOWN").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailMutationReceiptTarget) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_mutation_receipt_targets_request_index_bounds": "(request_index BETWEEN 0 AND 199)",
			"mail_mutation_receipt_targets_counts_bounds": "(affected_count BETWEEN 0 AND 200 AND succeeded_count BETWEEN 0 AND 200 AND failed_count BETWEEN 0 AND 200)",
			"mail_mutation_receipt_targets_counts_sum": "(succeeded_count + failed_count <= affected_count)",
			"mail_mutation_receipt_targets_outcome_training": "COALESCE(CASE status WHEN 'succeeded' THEN affected_count >= 1 AND succeeded_count = affected_count AND failed_count = 0 AND safe_code IS NULL AND training_status IN ('not_requested','accepted') WHEN 'partial' THEN affected_count >= 2 AND succeeded_count >= 1 AND failed_count >= 1 AND succeeded_count + failed_count = affected_count AND ((safe_code = 'MAIL_ACTION_REJECTED' AND training_status = 'not_requested') OR (safe_code = 'MAIL_CLASSIFIER_REJECTED' AND training_status = 'rejected')) WHEN 'failed' THEN ((affected_count = 0 AND succeeded_count = 0 AND failed_count = 0 AND safe_code IN ('MAIL_TARGET_UNAVAILABLE','MAIL_FOLDER_UNAVAILABLE','MAIL_LABEL_UNAVAILABLE') AND training_status = 'not_requested') OR (affected_count >= 1 AND succeeded_count = 0 AND failed_count = affected_count AND ((safe_code = 'MAIL_ACTION_REJECTED' AND training_status = 'not_requested') OR (safe_code = 'MAIL_CLASSIFIER_REJECTED' AND training_status = 'rejected')))) WHEN 'unknown' THEN affected_count >= 1 AND succeeded_count + failed_count < affected_count AND ((safe_code = 'MAIL_ENGINE_OUTCOME_UNKNOWN' AND training_status = 'not_requested') OR (safe_code = 'MAIL_CLASSIFIER_OUTCOME_UNKNOWN' AND training_status = 'unknown')) ELSE FALSE END, FALSE)",
		}),
	}
}

func (MailMutationReceiptTarget) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("receipt", MailMutationReceipt.Type).Ref("targets").Field("receipt_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("members", MailMutationReceiptMember.Type),
	}
}

func (MailMutationReceiptTarget) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("receipt_id", "request_index").Unique(),
		index.Fields("receipt_id", "target_type", "target_id").Unique(),
	}
}
