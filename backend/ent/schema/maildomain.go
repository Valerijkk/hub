package schema

import (
	"crypto/sha256"
	"errors"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/mail"
)

type MailDomain struct {
	ent.Schema
}

func (MailDomain) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}).
			Immutable(),
		field.String("domain").
			MinLen(1).
			MaxLen(253).
			Validate(validateNormalizedMailDomain).
			Immutable(),
		field.String("engine_provisioning_key").
			MinLen(1).
			MaxLen(1024).
			Immutable(),
		field.String("engine_domain_id").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(1024),
		field.Bool("desired_present").
			Default(true),
		field.Bool("enabled").
			Default(true),
		field.Bool("is_default").
			Default(false),
		field.Enum("state").
			Values("pending", "provisioning", "dns_required", "verified", "error", "deleting").
			Default("pending"),
		field.Enum("verification_status").
			Values("pending", "verified", "failed").
			Default("pending"),
		field.Enum("dkim_status").
			Values("unknown", "pending", "valid", "invalid").
			Default("unknown"),
		field.JSON("required_dns_records", mail.RequiredDNSRecords{}).
			Default(mail.RequiredDNSRecords{}),
		field.Int64("desired_generation").
			Default(1).
			Min(1),
		field.Int32("desired_hash_version").
			Default(1).
			Min(1),
		field.Bytes("desired_hash").
			MinLen(sha256.Size).
			MaxLen(sha256.Size),
		field.Int64("attempted_generation").
			Optional().
			Nillable().
			Min(1),
		field.Int32("attempted_hash_version").
			Optional().
			Nillable().
			Min(1),
		field.Bytes("attempted_hash").
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
		field.Enum("last_error_code").
			Values("MAIL_DOMAIN_ENGINE_UNAVAILABLE", "MAIL_DOMAIN_APPLY_REJECTED", "MAIL_DOMAIN_OUTCOME_UNKNOWN", "MAIL_DOMAIN_DNS_INVALID").
			Optional().
			Nillable(),
		field.Time("next_retry_at").
			Optional().
			Nillable(),
		field.Int32("attempt_count").
			Default(0).
			NonNegative(),
		field.UUID("lease_token", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("lease_until").
			Optional().
			Nillable(),
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

func (MailDomain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

func (MailDomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_domains_desired_hash":       "desired_generation >= 1 AND desired_hash_version >= 1 AND length(desired_hash) = 32",
			"mail_domains_attempted_tuple":    "((attempted_generation IS NULL) = (attempted_hash_version IS NULL)) AND ((attempted_generation IS NULL) = (attempted_hash IS NULL)) AND (attempted_generation IS NULL OR (attempted_generation >= 1 AND attempted_hash_version >= 1 AND length(attempted_hash) = 32 AND attempted_generation <= desired_generation))",
			"mail_domains_deployed_tuple":     "((deployed_generation IS NULL) = (deployed_hash_version IS NULL)) AND ((deployed_generation IS NULL) = (deployed_hash IS NULL)) AND (deployed_generation IS NULL OR (deployed_generation >= 1 AND deployed_hash_version >= 1 AND length(deployed_hash) = 32 AND deployed_generation <= desired_generation))",
			"mail_domains_lease_pair":         "(lease_token IS NULL) = (lease_until IS NULL)",
			"mail_domains_attempt_count":      "attempt_count >= 0",
			"mail_domains_unknown_attempt":    "last_error_code IS NULL OR last_error_code <> 'MAIL_DOMAIN_OUTCOME_UNKNOWN' OR attempted_generation IS NOT NULL",
			"mail_domains_pending_state":      "state <> 'pending' OR (desired_present = TRUE AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL)",
			"mail_domains_provisioning_state": "state <> 'provisioning' OR (desired_present = TRUE AND attempted_generation IS NOT NULL AND lease_token IS NOT NULL AND last_error_code IS NULL AND next_retry_at IS NULL)",
			"mail_domains_dns_required_state": "state <> 'dns_required' OR (desired_present = TRUE AND engine_domain_id IS NOT NULL AND deployed_generation IS NOT NULL AND deployed_generation = desired_generation AND deployed_hash_version = desired_hash_version AND deployed_hash = desired_hash AND jsonb_array_length(required_dns_records) > 0 AND verification_status <> 'verified' AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL AND next_retry_at IS NULL)",
			"mail_domains_verified_state":     "state <> 'verified' OR (desired_present = TRUE AND engine_domain_id IS NOT NULL AND deployed_generation IS NOT NULL AND deployed_generation = desired_generation AND deployed_hash_version = desired_hash_version AND deployed_hash = desired_hash AND verification_status = 'verified' AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL AND next_retry_at IS NULL)",
			"mail_domains_error_state":        "state <> 'error' OR (desired_present = TRUE AND last_error_code IS NOT NULL AND lease_token IS NULL AND ((last_error_code = 'MAIL_DOMAIN_OUTCOME_UNKNOWN' AND attempted_generation IS NOT NULL AND next_retry_at IS NOT NULL) OR (last_error_code <> 'MAIL_DOMAIN_OUTCOME_UNKNOWN' AND attempted_generation IS NULL)))",
			"mail_domains_deleting_state":     "state <> 'deleting' OR (desired_present = FALSE AND enabled = FALSE AND is_default = FALSE AND ((lease_token IS NOT NULL AND attempted_generation IS NOT NULL AND last_error_code IS NULL AND next_retry_at IS NULL) OR (lease_token IS NULL AND attempted_generation IS NOT NULL AND last_error_code = 'MAIL_DOMAIN_OUTCOME_UNKNOWN' AND next_retry_at IS NOT NULL) OR (lease_token IS NULL AND attempted_generation IS NULL)))",
			"mail_domains_deleted_state":      "deleted_at IS NULL OR (state = 'deleting' AND desired_present = FALSE AND enabled = FALSE AND is_default = FALSE AND engine_domain_id IS NULL AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL AND next_retry_at IS NULL)",
			"mail_domains_default_state":      "is_default = FALSE OR (deleted_at IS NULL AND desired_present = TRUE AND enabled = TRUE AND state = 'verified' AND verification_status = 'verified')",
		}),
	}
}

func (MailDomain) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("mail_domains").
			Field("organization_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("accounts", MailAccount.Type),
	}
}

func (MailDomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("domain").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("engine_provisioning_key").
			Unique(),
		index.Fields("engine_domain_id").
			Unique().
			Annotations(entsql.IndexWhere("engine_domain_id IS NOT NULL")),
		index.Fields("organization_id").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL AND is_default = TRUE")),
		index.Fields("state", "next_retry_at"),
	}
}

func validateNormalizedMailDomain(value string) error {
	if len(value) == 0 || len(value) > 253 {
		return errors.New("invalid normalized mail domain")
	}
	for _, label := range strings.Split(value, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return errors.New("invalid normalized mail domain")
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
				return errors.New("invalid normalized mail domain")
			}
		}
	}
	return nil
}
