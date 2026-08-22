package schema

import (
	"crypto/sha256"
	"errors"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailAccount struct {
	ent.Schema
}

func (MailAccount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Immutable(),
		field.UUID("mail_domain_id", uuid.UUID{}).
			Immutable(),
		field.String("engine_provisioning_key").
			MinLen(1).
			MaxLen(1024).
			Immutable(),
		field.String("local_part").
			MinLen(1).
			MaxLen(64).
			Validate(validateNormalizedMailLocalPart).
			Immutable(),
		field.String("engine_account_id").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(1024),
		field.Bool("desired_enabled").
			Default(true),
		field.Enum("status").
			Values("provisioning", "active", "suspended", "error", "deleting").
			Default("provisioning"),
		field.Int64("control_generation").
			Default(1).
			Min(1),
		field.Int32("control_hash_version").
			Default(1).
			Min(1),
		field.Bytes("control_hash").
			MinLen(sha256.Size).
			MaxLen(sha256.Size),
		field.Int64("attempted_control_generation").
			Optional().
			Nillable().
			Min(1),
		field.Int32("attempted_control_hash_version").
			Optional().
			Nillable().
			Min(1),
		field.Bytes("attempted_control_hash").
			Optional().
			Nillable().
			MinLen(sha256.Size).
			MaxLen(sha256.Size),
		field.Int64("deployed_generation").
			Optional().
			Nillable().
			Min(1),
		field.Int32("deployed_hash_version").
			Optional().
			Nillable().
			Min(1),
		field.Bytes("deployed_hash").
			Optional().
			Nillable().
			MinLen(sha256.Size).
			MaxLen(sha256.Size),
		field.Enum("projection_state").
			Values("bootstrapping", "ready", "degraded", "rebuilding").
			Default("bootstrapping"),
		field.String("sync_cursor").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(4096),
		field.String("rebuild_checkpoint").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(4096),
		field.String("rebuild_base_cursor").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(4096),
		field.Int64("sync_epoch").
			Default(0).
			NonNegative(),
		field.Int64("sync_lease_epoch").
			Default(1).
			Min(1),
		field.Int64("mutation_epoch").
			Default(0).
			NonNegative(),
		field.Int32("provision_attempts").
			Default(0).
			NonNegative(),
		field.Int32("sync_attempts").
			Default(0).
			NonNegative(),
		field.UUID("provision_lease_token", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("sync_lease_token", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int64("sync_lease_sync_epoch").
			Optional().
			Nillable().
			NonNegative(),
		field.Int64("sync_lease_mutation_epoch").
			Optional().
			Nillable().
			NonNegative(),
		field.Enum("sync_lease_phase").
			Values("incremental", "enumeration", "catch_up").
			Optional().
			Nillable(),
		field.String("sync_lease_cursor").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(4096),
		field.String("sync_lease_base_cursor").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(4096),
		field.String("sync_lease_checkpoint").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(4096),
		field.Int32("sync_lease_result_hash_version").
			Optional().
			Nillable().
			Min(1),
		field.Bytes("sync_lease_result_hash").
			Optional().
			Nillable().
			MinLen(sha256.Size).
			MaxLen(sha256.Size),
		field.Enum("sync_lease_applied_phase").
			Values("incremental", "enumeration", "catch_up").
			Optional().
			Nillable(),
		field.Time("sync_lease_applied_next_at").
			Optional().
			Nillable(),
		field.Time("sync_lease_applied_at").
			Optional().
			Nillable(),
		field.UUID("mutation_lease_token", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("provision_lease_until").
			Optional().
			Nillable(),
		field.Time("sync_lease_until").
			Optional().
			Nillable(),
		field.Time("mutation_lease_until").
			Optional().
			Nillable(),
		field.Time("next_provision_at").
			Optional().
			Nillable(),
		field.Time("next_sync_at").
			Optional().
			Nillable(),
		field.Time("sync_schedule_updated_at").
			Default(mailSyncScheduleNowUTC),
		field.Time("last_synced_at").
			Optional().
			Nillable(),
		field.Enum("provision_error_code").
			Values("MAILBOX_ENGINE_DISABLED", "MAILBOX_DOMAIN_REQUIRED", "MAILBOX_PROVISIONING_RETRY", "MAILBOX_SUSPENDED", "MAILBOX_UNAVAILABLE").
			Optional().
			Nillable(),
		field.Enum("sync_error_code").
			Values("MAIL_SYNC_ENGINE_UNAVAILABLE", "MAIL_SYNC_CURSOR_INVALID", "MAIL_SYNC_RESTART_REQUIRED", "MAIL_SYNC_PAGE_REJECTED", "MAIL_SYNC_PROJECTION_LIMIT", "MAIL_SYNC_CONTROL_CONFLICT").
			Optional().
			Nillable(),
		field.Time("event_cutoff_at").
			Immutable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

func (MailAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_accounts_control_hash":                   "control_generation >= 1 AND control_hash_version >= 1 AND length(control_hash) = 32",
			"mail_accounts_attempted_control_tuple":        "((attempted_control_generation IS NULL) = (attempted_control_hash_version IS NULL)) AND ((attempted_control_generation IS NULL) = (attempted_control_hash IS NULL)) AND (attempted_control_generation IS NULL OR (attempted_control_generation >= 1 AND attempted_control_hash_version >= 1 AND length(attempted_control_hash) = 32 AND attempted_control_generation <= control_generation))",
			"mail_accounts_deployed_control_tuple":         "((deployed_generation IS NULL) = (deployed_hash_version IS NULL)) AND ((deployed_generation IS NULL) = (deployed_hash IS NULL)) AND (deployed_generation IS NULL OR (deployed_generation >= 1 AND deployed_hash_version >= 1 AND length(deployed_hash) = 32 AND deployed_generation <= control_generation))",
			"mail_accounts_provision_lease_pair":           "(provision_lease_token IS NULL) = (provision_lease_until IS NULL)",
			"mail_accounts_sync_lease_pair":                "(sync_lease_token IS NULL) = (sync_lease_until IS NULL)",
			"mail_accounts_sync_lease_snapshot":            "COALESCE((sync_lease_token IS NULL AND sync_lease_sync_epoch IS NULL AND sync_lease_mutation_epoch IS NULL AND sync_lease_phase IS NULL AND sync_lease_cursor IS NULL AND sync_lease_base_cursor IS NULL AND sync_lease_checkpoint IS NULL) OR (sync_lease_token IS NOT NULL AND sync_lease_until IS NOT NULL AND sync_lease_sync_epoch >= 0 AND sync_lease_mutation_epoch >= 0 AND ((sync_lease_phase IN ('incremental', 'catch_up') AND sync_lease_cursor IS NOT NULL AND sync_lease_base_cursor IS NULL AND sync_lease_checkpoint IS NULL) OR (sync_lease_phase = 'enumeration' AND sync_lease_cursor IS NULL AND ((sync_lease_base_cursor IS NULL AND sync_lease_checkpoint IS NULL) OR (sync_lease_base_cursor IS NOT NULL AND sync_lease_checkpoint IS NOT NULL))))), FALSE)",
			"mail_accounts_sync_lease_apply_proof":         "COALESCE((sync_lease_result_hash_version IS NULL AND sync_lease_result_hash IS NULL AND sync_lease_applied_phase IS NULL AND sync_lease_applied_next_at IS NULL AND sync_lease_applied_at IS NULL) OR (sync_lease_token IS NOT NULL AND sync_lease_until IS NOT NULL AND sync_lease_result_hash_version IS NOT NULL AND sync_lease_result_hash IS NOT NULL AND sync_lease_applied_phase IS NOT NULL AND sync_lease_applied_next_at IS NOT NULL AND sync_lease_applied_at IS NOT NULL AND sync_lease_result_hash_version >= 1 AND length(sync_lease_result_hash) = 32 AND sync_lease_applied_at < sync_lease_until AND sync_lease_applied_next_at >= sync_lease_applied_at AND sync_lease_applied_next_at < sync_lease_until AND ((sync_lease_phase = 'incremental' AND sync_lease_applied_phase = 'incremental') OR (sync_lease_phase = 'catch_up' AND sync_lease_applied_phase IN ('catch_up', 'incremental')) OR (sync_lease_phase = 'enumeration' AND sync_lease_applied_phase IN ('enumeration', 'catch_up')))), FALSE)",
			"mail_accounts_mutation_lease_pair":            "(mutation_lease_token IS NULL) = (mutation_lease_until IS NULL)",
			"mail_accounts_sync_schedule_updated_at_second_precision": "COALESCE(date_trunc('second', sync_schedule_updated_at) = sync_schedule_updated_at, FALSE)",
			"mail_accounts_non_negative":                   "sync_epoch >= 0 AND sync_lease_epoch >= 1 AND mutation_epoch >= 0 AND provision_attempts >= 0 AND sync_attempts >= 0",
			"mail_accounts_status_state":                   "status = 'provisioning' OR (status = 'error' AND provision_error_code IS NOT NULL) OR (status = 'active' AND desired_enabled = TRUE AND engine_account_id IS NOT NULL AND deployed_generation IS NOT NULL) OR (status = 'suspended' AND desired_enabled = FALSE AND engine_account_id IS NOT NULL AND deployed_generation IS NOT NULL) OR status = 'deleting'",
			"mail_accounts_deleting_blocked_pending_shape": "status <> 'deleting' OR (desired_enabled = FALSE AND sync_lease_token IS NULL AND sync_lease_until IS NULL AND mutation_lease_token IS NULL AND mutation_lease_until IS NULL AND provision_lease_token IS NULL AND provision_lease_until IS NULL AND attempted_control_generation IS NULL AND attempted_control_hash_version IS NULL AND attempted_control_hash IS NULL AND provision_error_code IS NULL AND next_provision_at IS NOT NULL)",
			"mail_accounts_ready_projection":               "projection_state <> 'ready' OR (sync_cursor IS NOT NULL AND last_synced_at IS NOT NULL AND rebuild_checkpoint IS NULL AND rebuild_base_cursor IS NULL AND sync_error_code IS NULL)",
			"mail_accounts_rebuilding_projection":          "projection_state <> 'rebuilding' OR rebuild_base_cursor IS NOT NULL",
			"mail_accounts_live_provision_attempt":         "provision_lease_token IS NULL OR attempted_control_generation IS NOT NULL",
			"mail_accounts_retained_provision_attempt":     "attempted_control_generation IS NULL OR provision_lease_token IS NOT NULL OR (provision_error_code IS NOT NULL AND provision_error_code IN ('MAILBOX_PROVISIONING_RETRY','MAILBOX_UNAVAILABLE') AND next_provision_at IS NOT NULL)",
			"mail_accounts_unconverged_schedule":           "provision_lease_token IS NOT NULL OR attempted_control_generation IS NOT NULL OR next_provision_at IS NOT NULL OR (deployed_generation IS NOT NULL AND deployed_generation = control_generation AND deployed_hash_version = control_hash_version AND deployed_hash = control_hash)",
		}),
	}
}

func (MailAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("mail_accounts").
			Field("organization_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("user", User.Type).
			Ref("mail_account").
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("domain", MailDomain.Type).
			Ref("accounts").
			Field("mail_domain_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("identities", MailIdentity.Type),
		edge.To("user_settings", MailUserSettings.Type).
			Unique(),
		edge.To("signatures", MailSignature.Type),
		edge.To("rules", MailRule.Type),
		edge.To("recent_contacts", RecentContact.Type),
		edge.To("mutation_receipts", MailMutationReceipt.Type),
		edge.To("folders", MailFolder.Type),
		edge.To("threads", MailThread.Type),
		edge.To("messages", MailMessage.Type),
		edge.To("labels", MailLabel.Type),
		edge.To("objects", MailObject.Type),
		edge.To("drafts", MailDraft.Type),
		edge.To("submissions", MailSubmission.Type),
	}
}

func (MailAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id").
			Unique(),
		index.Fields("engine_provisioning_key").
			Unique(),
		index.Fields("engine_account_id").
			Unique().
			Annotations(entsql.IndexWhere("engine_account_id IS NOT NULL")),
		index.Fields("mail_domain_id", "local_part").
			Unique(),
		index.Fields("status", "next_provision_at"),
		index.Fields("projection_state", "next_sync_at"),
		index.Fields("next_provision_at", "id").
			StorageKey("mail_accounts_deleting_blocked_idx").
			Annotations(entsql.IndexWhere("status = 'deleting'")),
		index.Fields("next_sync_at", "id").
			StorageKey("mail_accounts_sync_due_idx").
			Annotations(entsql.IndexWhere("status = 'active' AND next_sync_at IS NOT NULL")),
	}
}

func mailSyncScheduleNowUTC() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func validateNormalizedMailLocalPart(value string) error {
	if len(value) == 0 || len(value) > 64 || value[0] == '.' || value[len(value)-1] == '.' {
		return errors.New("invalid normalized mail local part")
	}
	previousDot := false
	for _, character := range []byte(value) {
		if character == '.' {
			if previousDot {
				return errors.New("invalid normalized mail local part")
			}
			previousDot = true
			continue
		}
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || isMailAText(character) {
			previousDot = false
			continue
		}
		return errors.New("invalid normalized mail local part")
	}
	return nil
}

func isMailAText(character byte) bool {
	switch character {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '/', '=', '?', '^', '_', '`', '{', '|', '}', '~':
		return true
	default:
		return false
	}
}
