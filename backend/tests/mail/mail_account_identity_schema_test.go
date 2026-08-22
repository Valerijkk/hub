package mail_test

import (
	"crypto/sha256"
	"reflect"
	"strings"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestMailAccountSchema_HasExactFieldsDefaultsAndValidators(t *testing.T) {
	account := entschema.MailAccount{}
	assertNoMixins(t, account.Mixin())
	assertExactOrderedFieldNames(t, "MailAccount fields", account.Fields(), []string{
		"id", "organization_id", "user_id", "mail_domain_id", "engine_provisioning_key", "local_part",
		"engine_account_id", "desired_enabled", "status", "control_generation", "control_hash_version",
		"control_hash", "attempted_control_generation", "attempted_control_hash_version",
		"attempted_control_hash", "deployed_generation", "deployed_hash_version", "deployed_hash",
		"projection_state", "sync_cursor", "rebuild_checkpoint", "rebuild_base_cursor", "sync_epoch",
		"sync_lease_epoch", "mutation_epoch", "provision_attempts", "sync_attempts", "provision_lease_token",
		"sync_lease_token", "sync_lease_sync_epoch", "sync_lease_mutation_epoch", "sync_lease_phase",
		"sync_lease_cursor", "sync_lease_base_cursor", "sync_lease_checkpoint", "sync_lease_result_hash_version",
		"sync_lease_result_hash", "sync_lease_applied_phase", "sync_lease_applied_next_at", "sync_lease_applied_at",
		"mutation_lease_token", "provision_lease_until", "sync_lease_until",
		"mutation_lease_until", "next_provision_at", "next_sync_at", "sync_schedule_updated_at", "last_synced_at",
		"provision_error_code", "sync_error_code", "event_cutoff_at", "created_at", "updated_at",
	})
	fields := fieldDescriptors(account.Fields())

	for _, name := range []string{"id", "organization_id", "user_id", "mail_domain_id", "provision_lease_token", "sync_lease_token", "mutation_lease_token"} {
		assertFieldType(t, name, fields[name], field.TypeUUID)
	}
	assertFieldType(t, "desired_enabled", fields["desired_enabled"], field.TypeBool)
	for _, name := range []string{"status", "projection_state", "sync_lease_phase", "sync_lease_applied_phase", "provision_error_code", "sync_error_code"} {
		assertFieldType(t, name, fields[name], field.TypeEnum)
	}
	for _, name := range []string{"id", "organization_id", "user_id", "mail_domain_id", "engine_provisioning_key", "local_part", "event_cutoff_at"} {
		if !fields[name].Immutable {
			t.Fatalf("%s must be immutable", name)
		}
	}
	for _, name := range []string{"id", "organization_id", "user_id", "mail_domain_id", "engine_provisioning_key", "local_part", "desired_enabled", "status", "control_generation", "control_hash_version", "control_hash", "projection_state", "sync_epoch", "sync_lease_epoch", "mutation_epoch", "provision_attempts", "sync_attempts", "sync_schedule_updated_at", "event_cutoff_at", "created_at", "updated_at"} {
		assertRequired(t, name, fields[name])
	}
	if fields["id"].Default == nil {
		t.Fatal("id must have a UUID default")
	}

	for _, name := range []string{"engine_provisioning_key", "local_part", "engine_account_id", "sync_cursor", "rebuild_checkpoint", "rebuild_base_cursor", "sync_lease_cursor", "sync_lease_base_cursor", "sync_lease_checkpoint"} {
		assertFieldType(t, name, fields[name], field.TypeString)
	}
	assertStringValidators(t, "engine_provisioning_key", fields["engine_provisioning_key"], []struct {
		value string
		want  bool
	}{
		{"key", true},
		{"", false},
		{strings.Repeat("k", 1025), false},
	})
	assertStringValidators(t, "local_part", fields["local_part"], []struct {
		value string
		want  bool
	}{
		{"mailbox", true},
		{"a.b", true},
		{strings.Repeat("a", 64), true},
		{"", false},
		{".mailbox", false},
		{"mailbox.", false},
		{"a..b", false},
		{"Mailbox", false},
		{"mail box", false},
		{"m\u00e4ilbox", false},
		{strings.Repeat("a", 65), false},
	})
	for _, name := range []string{"engine_account_id", "sync_cursor", "rebuild_checkpoint", "rebuild_base_cursor"} {
		assertOptionalNillable(t, name, fields[name])
		assertStringValidators(t, name, fields[name], []struct {
			value string
			want  bool
		}{
			{"opaque", true},
			{"", false},
			{strings.Repeat("x", 4097), false},
		})
	}
	for _, name := range []string{"sync_lease_cursor", "sync_lease_base_cursor", "sync_lease_checkpoint"} {
		assertOptionalNillable(t, name, fields[name])
		assertStringValidators(t, name, fields[name], []struct {
			value string
			want  bool
		}{
			{"", false},
			{"x", true},
			{strings.Repeat("x", 4096), true},
			{strings.Repeat("x", 4097), false},
		})
	}
	assertStringValidators(t, "engine_account_id", fields["engine_account_id"], []struct {
		value string
		want  bool
	}{
		{strings.Repeat("x", 1024), true},
		{strings.Repeat("x", 1025), false},
	})

	assertExactDefaults(t, fields, map[string]any{
		"desired_enabled":      true,
		"status":               "provisioning",
		"control_generation":   int64(1),
		"control_hash_version": int32(1),
		"projection_state":     "bootstrapping",
		"sync_epoch":           int64(0),
		"sync_lease_epoch":     int64(1),
		"mutation_epoch":       int64(0),
		"provision_attempts":   int32(0),
		"sync_attempts":        int32(0),
	})
	assertOnlyDefaults(t, fields, map[string]bool{
		"id": true, "desired_enabled": true, "status": true, "control_generation": true,
		"control_hash_version": true, "projection_state": true, "sync_epoch": true,
		"sync_lease_epoch": true,
		"mutation_epoch": true, "provision_attempts": true, "sync_attempts": true, "sync_schedule_updated_at": true,
		"created_at": true, "updated_at": true,
	})
	descriptor := fields["sync_schedule_updated_at"]
	if descriptor == nil {
		t.Fatal("sync_schedule_updated_at descriptor is nil")
	}
	defaultNow, ok := descriptor.Default.(func() time.Time)
	if !ok {
		t.Fatalf("sync_schedule_updated_at default = %T, want func() time.Time", descriptor.Default)
	}
	scheduledAt := defaultNow()
	if scheduledAt.IsZero() {
		t.Fatal("sync_schedule_updated_at default returned zero time")
	}
	if scheduledAt.Location() != time.UTC {
		t.Fatalf("sync_schedule_updated_at location = %v, want UTC", scheduledAt.Location())
	}
	if scheduledAt.Nanosecond() != 0 {
		t.Fatalf("sync_schedule_updated_at nanosecond = %d, want 0", scheduledAt.Nanosecond())
	}
	if descriptor.UpdateDefault != nil {
		t.Fatalf("sync_schedule_updated_at update default = %#v, want nil", descriptor.UpdateDefault)
	}
	assertExactEnumValues(t, "status", fields["status"], []string{"provisioning", "active", "suspended", "error", "deleting"})
	assertExactEnumValues(t, "projection_state", fields["projection_state"], []string{"bootstrapping", "ready", "degraded", "rebuilding"})
	assertExactEnumValues(t, "sync_lease_phase", fields["sync_lease_phase"], []string{"incremental", "enumeration", "catch_up"})
	assertExactEnumValues(t, "sync_lease_applied_phase", fields["sync_lease_applied_phase"], []string{"incremental", "enumeration", "catch_up"})
	assertExactEnumValues(t, "provision_error_code", fields["provision_error_code"], []string{
		"MAILBOX_ENGINE_DISABLED", "MAILBOX_DOMAIN_REQUIRED", "MAILBOX_PROVISIONING_RETRY", "MAILBOX_SUSPENDED", "MAILBOX_UNAVAILABLE",
	})
	assertExactEnumValues(t, "sync_error_code", fields["sync_error_code"], []string{
		"MAIL_SYNC_ENGINE_UNAVAILABLE", "MAIL_SYNC_CURSOR_INVALID", "MAIL_SYNC_RESTART_REQUIRED", "MAIL_SYNC_PAGE_REJECTED", "MAIL_SYNC_PROJECTION_LIMIT", "MAIL_SYNC_CONTROL_CONFLICT",
	})
	for _, name := range []string{"provision_error_code", "sync_error_code"} {
		assertOptionalNillable(t, name, fields[name])
	}

	for _, name := range []string{"control_generation", "attempted_control_generation", "deployed_generation", "sync_epoch", "sync_lease_epoch", "sync_lease_sync_epoch", "sync_lease_mutation_epoch", "mutation_epoch"} {
		assertFieldType(t, name, fields[name], field.TypeInt64)
	}
	for _, name := range []string{"control_hash_version", "attempted_control_hash_version", "deployed_hash_version", "sync_lease_result_hash_version", "provision_attempts", "sync_attempts"} {
		assertFieldType(t, name, fields[name], field.TypeInt32)
	}
	for _, name := range []string{"control_generation", "attempted_control_generation", "deployed_generation", "sync_lease_epoch"} {
		assertInt64Validators(t, name, fields[name], []struct {
			value int64
			want  bool
		}{{0, false}, {1, true}})
	}
	for _, name := range []string{"control_hash_version", "attempted_control_hash_version", "deployed_hash_version", "sync_lease_result_hash_version"} {
		assertInt32Validators(t, name, fields[name], []struct {
			value int32
			want  bool
		}{{0, false}, {1, true}})
	}
	for _, name := range []string{"sync_epoch", "sync_lease_sync_epoch", "sync_lease_mutation_epoch", "mutation_epoch"} {
		assertInt64Validators(t, name, fields[name], []struct {
			value int64
			want  bool
		}{{-1, false}, {0, true}})
	}
	for _, name := range []string{"provision_attempts", "sync_attempts"} {
		assertInt32Validators(t, name, fields[name], []struct {
			value int32
			want  bool
		}{{-1, false}, {0, true}})
	}
	for _, name := range []string{"control_hash", "attempted_control_hash", "deployed_hash", "sync_lease_result_hash"} {
		assertFieldType(t, name, fields[name], field.TypeBytes)
		assertBytesValidators(t, name, fields[name], []struct {
			value []byte
			want  bool
		}{
			{make([]byte, sha256.Size-1), false},
			{make([]byte, sha256.Size), true},
			{make([]byte, sha256.Size+1), false},
		})
	}
	for _, name := range []string{
		"attempted_control_generation", "attempted_control_hash_version", "attempted_control_hash",
		"deployed_generation", "deployed_hash_version", "deployed_hash", "provision_lease_token",
		"sync_lease_token", "sync_lease_sync_epoch", "sync_lease_mutation_epoch", "sync_lease_phase",
		"sync_lease_cursor", "sync_lease_base_cursor", "sync_lease_checkpoint", "sync_lease_result_hash_version",
		"sync_lease_result_hash", "sync_lease_applied_phase", "sync_lease_applied_next_at", "sync_lease_applied_at",
		"mutation_lease_token", "provision_lease_until", "sync_lease_until",
		"mutation_lease_until", "next_provision_at", "next_sync_at", "last_synced_at",
		"provision_error_code", "sync_error_code",
	} {
		assertOptionalNillable(t, name, fields[name])
	}
	for _, name := range []string{"event_cutoff_at", "created_at", "updated_at", "provision_lease_until", "sync_lease_until", "sync_lease_applied_next_at", "sync_lease_applied_at", "mutation_lease_until", "next_provision_at", "next_sync_at", "sync_schedule_updated_at", "last_synced_at"} {
		assertFieldType(t, name, fields[name], field.TypeTime)
	}
	assertTimestampsWithoutMixin(t, fields)
}

func TestMailIdentitySchema_HasExactFieldsDefaultsAndValidators(t *testing.T) {
	identity := entschema.MailIdentity{}
	assertNoMixins(t, identity.Mixin())
	assertExactOrderedFieldNames(t, "MailIdentity fields", identity.Fields(), []string{
		"id", "mail_account_id", "normalized_address", "display_name", "kind", "send_allowed", "status", "engine_identity_id", "created_at", "updated_at",
	})
	fields := fieldDescriptors(identity.Fields())
	for _, name := range []string{"id", "mail_account_id"} {
		assertFieldType(t, name, fields[name], field.TypeUUID)
		assertRequired(t, name, fields[name])
		if !fields[name].Immutable {
			t.Fatalf("%s must be immutable", name)
		}
	}
	assertFieldType(t, "kind", fields["kind"], field.TypeEnum)
	assertFieldType(t, "send_allowed", fields["send_allowed"], field.TypeBool)
	assertFieldType(t, "status", fields["status"], field.TypeEnum)
	if fields["id"].Default == nil {
		t.Fatal("id must have a UUID default")
	}
	for _, name := range []string{"normalized_address", "display_name", "engine_identity_id"} {
		assertFieldType(t, name, fields[name], field.TypeString)
	}
	for _, name := range []string{"normalized_address", "kind"} {
		if !fields[name].Immutable {
			t.Fatalf("%s must be immutable", name)
		}
	}
	assertRequired(t, "normalized_address", fields["normalized_address"])
	assertStringValidators(t, "normalized_address", fields["normalized_address"], []struct {
		value string
		want  bool
	}{
		{"mailbox@example.test", true},
		{maxNormalizedMailAddress(), true},
		{"mailbox@Example.test", false},
		{"Mailbox@example.test", false},
		{"mailbox@example..test", false},
		{"mailbox@@example.test", false},
		{"mailbox@", false},
		{string([]byte{'m', 'a', 'i', 'l', 0xff, '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 't', 'e', 's', 't'}), false},
		{strings.Repeat("a", 65) + "@example.test", false},
	})
	assertOptionalNillable(t, "display_name", fields["display_name"])
	assertStringValidators(t, "display_name", fields["display_name"], []struct {
		value string
		want  bool
	}{
		{"Mailbox", true},
		{strings.Repeat("\U0001f600", 160), true},
		{strings.Repeat("\U0001f600", 161), false},
		{strings.Repeat("a", 641), false},
		{string([]byte{0xff}), false},
	})
	assertOptionalNillable(t, "engine_identity_id", fields["engine_identity_id"])
	assertStringValidators(t, "engine_identity_id", fields["engine_identity_id"], []struct {
		value string
		want  bool
	}{
		{"opaque-identity-id", true},
		{"", false},
		{strings.Repeat("i", 1024), true},
		{strings.Repeat("i", 1025), false},
	})
	for _, name := range []string{"kind", "send_allowed", "status"} {
		assertRequired(t, name, fields[name])
	}
	assertExactDefaults(t, fields, map[string]any{
		"send_allowed": false,
		"status":       "pending",
	})
	assertOnlyDefaults(t, fields, map[string]bool{
		"id": true, "send_allowed": true, "status": true, "created_at": true, "updated_at": true,
	})
	assertExactEnumValues(t, "kind", fields["kind"], []string{"primary", "alias", "group"})
	assertExactEnumValues(t, "status", fields["status"], []string{"pending", "active", "error", "deleting"})
	for _, name := range []string{"created_at", "updated_at"} {
		assertFieldType(t, name, fields[name], field.TypeTime)
		assertRequired(t, name, fields[name])
	}
	assertTimestampsWithoutMixin(t, fields)
}

func TestMailAccountAndIdentitySchemas_HaveExactChecksAndIndexes(t *testing.T) {
	assertExactChecks(t, (entschema.MailAccount{}).Annotations(), map[string]string{
		"mail_accounts_control_hash":                   "control_generation >= 1 AND control_hash_version >= 1 AND length(control_hash) = 32",
		"mail_accounts_attempted_control_tuple":        "((attempted_control_generation IS NULL) = (attempted_control_hash_version IS NULL)) AND ((attempted_control_generation IS NULL) = (attempted_control_hash IS NULL)) AND (attempted_control_generation IS NULL OR (attempted_control_generation >= 1 AND attempted_control_hash_version >= 1 AND length(attempted_control_hash) = 32 AND attempted_control_generation <= control_generation))",
		"mail_accounts_deployed_control_tuple":         "((deployed_generation IS NULL) = (deployed_hash_version IS NULL)) AND ((deployed_generation IS NULL) = (deployed_hash IS NULL)) AND (deployed_generation IS NULL OR (deployed_generation >= 1 AND deployed_hash_version >= 1 AND length(deployed_hash) = 32 AND deployed_generation <= control_generation))",
		"mail_accounts_provision_lease_pair":           "(provision_lease_token IS NULL) = (provision_lease_until IS NULL)",
		"mail_accounts_sync_lease_pair":                "(sync_lease_token IS NULL) = (sync_lease_until IS NULL)",
		"mail_accounts_sync_lease_snapshot":            "COALESCE((sync_lease_token IS NULL AND sync_lease_sync_epoch IS NULL AND sync_lease_mutation_epoch IS NULL AND sync_lease_phase IS NULL AND sync_lease_cursor IS NULL AND sync_lease_base_cursor IS NULL AND sync_lease_checkpoint IS NULL) OR (sync_lease_token IS NOT NULL AND sync_lease_until IS NOT NULL AND sync_lease_sync_epoch >= 0 AND sync_lease_mutation_epoch >= 0 AND ((sync_lease_phase IN ('incremental', 'catch_up') AND sync_lease_cursor IS NOT NULL AND sync_lease_base_cursor IS NULL AND sync_lease_checkpoint IS NULL) OR (sync_lease_phase = 'enumeration' AND sync_lease_cursor IS NULL AND ((sync_lease_base_cursor IS NULL AND sync_lease_checkpoint IS NULL) OR (sync_lease_base_cursor IS NOT NULL AND sync_lease_checkpoint IS NOT NULL))))), FALSE)",
		"mail_accounts_sync_lease_apply_proof":         "COALESCE((sync_lease_result_hash_version IS NULL AND sync_lease_result_hash IS NULL AND sync_lease_applied_phase IS NULL AND sync_lease_applied_next_at IS NULL AND sync_lease_applied_at IS NULL) OR (sync_lease_token IS NOT NULL AND sync_lease_until IS NOT NULL AND sync_lease_result_hash_version IS NOT NULL AND sync_lease_result_hash IS NOT NULL AND sync_lease_applied_phase IS NOT NULL AND sync_lease_applied_next_at IS NOT NULL AND sync_lease_applied_at IS NOT NULL AND sync_lease_result_hash_version >= 1 AND length(sync_lease_result_hash) = 32 AND sync_lease_applied_at < sync_lease_until AND sync_lease_applied_next_at >= sync_lease_applied_at AND ((sync_lease_phase = 'incremental' AND sync_lease_applied_phase = 'incremental') OR (sync_lease_phase = 'catch_up' AND sync_lease_applied_phase IN ('catch_up', 'incremental')) OR (sync_lease_phase = 'enumeration' AND sync_lease_applied_phase IN ('enumeration', 'catch_up')))), FALSE)",
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
	})
	assertExactChecks(t, (entschema.MailIdentity{}).Annotations(), map[string]string{
		"mail_identities_send_allowed":      "send_allowed = FALSE OR status = 'active'",
		"mail_identities_group_not_sendable": "kind <> 'group' OR send_allowed = FALSE",
		"mail_identities_deleting_alias_only": "status <> 'deleting' OR kind = 'alias'",
	})

	accountIndexes := (entschema.MailAccount{}).Indexes()
	assertExactIndexes(t, accountIndexes, []wantedIndex{
		{fields: []string{"user_id"}, unique: true},
		{fields: []string{"engine_provisioning_key"}, unique: true},
		{fields: []string{"engine_account_id"}, unique: true, where: "engine_account_id IS NOT NULL"},
		{fields: []string{"mail_domain_id", "local_part"}, unique: true},
		{fields: []string{"status", "next_provision_at"}},
		{fields: []string{"projection_state", "next_sync_at"}},
		{fields: []string{"next_provision_at", "id"}, where: "status = 'deleting'"},
		{fields: []string{"next_sync_at", "id"}, where: "status = 'active' AND next_sync_at IS NOT NULL"},
	})
	if got := accountIndexes[6].Descriptor().StorageKey; got != "mail_accounts_deleting_blocked_idx" {
		t.Fatalf("MailAccount index[6] storage key = %q, want mail_accounts_deleting_blocked_idx", got)
	}
	if got := accountIndexes[7].Descriptor().StorageKey; got != "mail_accounts_sync_due_idx" {
		t.Fatalf("MailAccount index[7] storage key = %q, want mail_accounts_sync_due_idx", got)
	}
	assertExactIndexes(t, (entschema.MailIdentity{}).Indexes(), []wantedIndex{
		{fields: []string{"normalized_address"}, unique: true},
		{fields: []string{"mail_account_id", "engine_identity_id"}, unique: true, where: "engine_identity_id IS NOT NULL"},
		{fields: []string{"mail_account_id"}, unique: true, where: "kind = 'primary'"},
		{fields: []string{"mail_account_id", "status"}},
	})
}

func TestMailAccountAndIdentitySchemas_HaveOnlyFrozenEdges(t *testing.T) {
	accountEdges := (entschema.MailAccount{}).Edges()
	assertExactOrderedEdgeNames(t, "MailAccount edges", accountEdges, []string{"organization", "user", "domain", "identities", "user_settings", "signatures", "rules", "recent_contacts", "mutation_receipts", "folders", "threads", "messages", "labels", "objects", "drafts", "submissions"})
	accountEdgeDescriptors := edgeDescriptors(accountEdges)
	assertRequiredInverseEdge(t, "MailAccount.organization", accountEdgeDescriptors["organization"], "organization_id", "mail_accounts", entsql.NoAction)
	assertRequiredInverseEdge(t, "MailAccount.user", accountEdgeDescriptors["user"], "user_id", "mail_account", entsql.NoAction)
	assertRequiredInverseEdge(t, "MailAccount.domain", accountEdgeDescriptors["domain"], "mail_domain_id", "accounts", entsql.NoAction)
	assertPlainToManyEdge(t, "MailAccount.identities", accountEdgeDescriptors["identities"])
	if userSettings := accountEdgeDescriptors["user_settings"]; userSettings == nil || userSettings.Inverse || !userSettings.Unique || userSettings.Required || userSettings.RefName != "" {
		t.Fatalf("MailAccount.user_settings = %#v, want plain unique inverse target", userSettings)
	}
	for _, name := range []string{"signatures", "rules", "recent_contacts", "mutation_receipts", "folders", "threads", "messages", "labels", "objects", "drafts", "submissions"} {
		assertPlainToManyEdge(t, "MailAccount."+name, accountEdgeDescriptors[name])
	}

	identityEdges := (entschema.MailIdentity{}).Edges()
	assertExactOrderedEdgeNames(t, "MailIdentity edges", identityEdges, []string{"account", "default_for_settings"})
	identityEdgeDescriptors := edgeDescriptors(identityEdges)
	assertRequiredInverseEdge(t, "MailIdentity.account", identityEdgeDescriptors["account"], "mail_account_id", "identities", entsql.Cascade)
	if defaultForSettings := identityEdgeDescriptors["default_for_settings"]; defaultForSettings == nil || defaultForSettings.Inverse || !defaultForSettings.Unique || defaultForSettings.Required || defaultForSettings.RefName != "" {
		t.Fatalf("MailIdentity.default_for_settings = %#v, want plain unique inverse target", defaultForSettings)
	}

	domainEdges := (entschema.MailDomain{}).Edges()
	assertExactOrderedEdgeNames(t, "MailDomain edges", domainEdges, []string{"organization", "accounts"})
	assertPlainToManyEdge(t, "MailDomain.accounts", edgeDescriptors(domainEdges)["accounts"])

	organizationEdges := (entschema.Organization{}).Edges()
	assertExactOrderedEdgeNames(t, "Organization edges", organizationEdges, []string{"users", "projects", "mail_domains", "mailbox_policy", "mail_accounts", "mail_mutation_receipts", "mail_objects"})
	assertPlainToManyEdge(t, "Organization.mail_accounts", edgeDescriptors(organizationEdges)["mail_accounts"])
	assertPlainToManyEdge(t, "Organization.mail_mutation_receipts", edgeDescriptors(organizationEdges)["mail_mutation_receipts"])
	assertPlainToManyEdge(t, "Organization.mail_objects", edgeDescriptors(organizationEdges)["mail_objects"])

	userEdges := (entschema.User{}).Edges()
	if len(userEdges) < 3 || userEdges[len(userEdges)-3].Descriptor().Name != "mail_account" || userEdges[len(userEdges)-2].Descriptor().Name != "recent_contacts" || userEdges[len(userEdges)-1].Descriptor().Name != "mail_objects" {
		t.Fatalf("User must append mail_account, recent_contacts, then mail_objects as the A3/02D/02C inverse edges; trailing edges = %#v", userEdges)
	}
	mailAccount := edgeDescriptors(userEdges)["mail_account"]
	if mailAccount == nil || mailAccount.Inverse || !mailAccount.Unique || mailAccount.Required || mailAccount.RefName != "" {
		t.Fatalf("User.mail_account = %#v, want plain unique inverse target", mailAccount)
	}
	assertPlainToManyEdge(t, "User.recent_contacts", edgeDescriptors(userEdges)["recent_contacts"])
	assertPlainToManyEdge(t, "User.mail_objects", edgeDescriptors(userEdges)["mail_objects"])
}

func assertNoMixins(t *testing.T, mixins []ent.Mixin) {
	t.Helper()
	if len(mixins) != 0 {
		t.Fatalf("mixins = %#v, want none", mixins)
	}
}

func assertExactOrderedFieldNames(t *testing.T, label string, fields []ent.Field, want []string) {
	t.Helper()
	got := make([]string, 0, len(fields))
	for _, schemaField := range fields {
		got = append(got, schemaField.Descriptor().Name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func assertExactOrderedEdgeNames(t *testing.T, label string, edges []ent.Edge, want []string) {
	t.Helper()
	got := make([]string, 0, len(edges))
	for _, schemaEdge := range edges {
		got = append(got, schemaEdge.Descriptor().Name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func assertRequiredInverseEdge(t *testing.T, name string, descriptor *edge.Descriptor, fieldName, ref string, onDelete entsql.ReferenceOption) {
	t.Helper()
	if descriptor == nil || !descriptor.Inverse || !descriptor.Unique || !descriptor.Required || descriptor.Field != fieldName || descriptor.RefName != ref {
		t.Fatalf("%s = %#v, want inverse required unique field %q ref %q", name, descriptor, fieldName, ref)
	}
	assertOnDelete(t, name, descriptor.Annotations, onDelete)
}

func assertPlainToManyEdge(t *testing.T, name string, descriptor *edge.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Inverse || descriptor.Unique || descriptor.Required || descriptor.RefName != "" {
		t.Fatalf("%s = %#v, want plain to-many edge", name, descriptor)
	}
}

func assertTimestampsWithoutMixin(t *testing.T, fields map[string]*field.Descriptor) {
	t.Helper()
	if fields["created_at"].Default == nil || !fields["created_at"].Immutable {
		t.Fatalf("created_at = %#v, want immutable creation default", fields["created_at"])
	}
	if fields["updated_at"].Default == nil || fields["updated_at"].UpdateDefault == nil {
		t.Fatalf("updated_at = %#v, want creation and update defaults", fields["updated_at"])
	}
}

func assertOnlyDefaults(t *testing.T, fields map[string]*field.Descriptor, allowed map[string]bool) {
	t.Helper()
	for name, descriptor := range fields {
		if descriptor.Default != nil && !allowed[name] {
			t.Fatalf("%s default = %#v, want no schema default", name, descriptor.Default)
		}
	}
}

func maxNormalizedMailAddress() string {
	return strings.Repeat("a", 64) + "@" + strings.Repeat("a", 63) + "." + strings.Repeat("a", 63) + "." + strings.Repeat("a", 63) + "." + strings.Repeat("a", 61)
}
