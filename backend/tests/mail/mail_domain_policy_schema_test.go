package mail_test

import (
	"crypto/sha256"
	"reflect"
	"sort"
	"strings"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailDomainSchema_HasExactFieldsDefaultsAndEnums(t *testing.T) {
	fields := fieldDescriptors((entschema.MailDomain{}).Fields())
	assertExactNames(t, "MailDomain fields", fieldNames(fields), []string{
		"attempt_count", "attempted_generation", "attempted_hash", "attempted_hash_version",
		"created_at", "deleted_at", "deployed_generation", "deployed_hash", "deployed_hash_version",
		"desired_generation", "desired_hash", "desired_hash_version", "desired_present", "dkim_status",
		"domain", "enabled", "engine_domain_id", "engine_provisioning_key", "id", "is_default",
		"last_error_code", "lease_token", "lease_until", "next_retry_at", "organization_id",
		"required_dns_records", "state", "updated_at", "verification_status",
	})

	for _, name := range []string{"id", "organization_id"} {
		assertFieldType(t, name, fields[name], field.TypeUUID)
		assertRequired(t, name, fields[name])
	}
	for _, name := range []string{"domain", "engine_provisioning_key"} {
		assertFieldType(t, name, fields[name], field.TypeString)
		assertRequired(t, name, fields[name])
		if !fields[name].Immutable {
			t.Fatalf("%s must be immutable", name)
		}
	}
	if !fields["organization_id"].Immutable {
		t.Fatal("organization_id must be immutable")
	}
	if !fields["id"].Immutable {
		t.Fatal("id must be immutable")
	}

	assertStringValidators(t, "domain", fields["domain"], []struct {
		value string
		want  bool
	}{
		{"mail.example.test", true},
		{"a", true},
		{"a-1.example-2", true},
		{"", false},
		{"Mail.example.test", false},
		{"mail.example.test.", false},
		{"-mail.example.test", false},
		{"mail-.example.test", false},
		{"mail..example.test", false},
		{"mail_example.test", false},
		{strings.Repeat("a", 64) + ".example.test", false},
	})
	assertStringValidators(t, "engine_provisioning_key", fields["engine_provisioning_key"], []struct {
		value string
		want  bool
	}{
		{"key", true},
		{"", false},
		{strings.Repeat("k", 1025), false},
	})
	engineDomainID := fields["engine_domain_id"]
	assertFieldType(t, "engine_domain_id", engineDomainID, field.TypeString)
	assertOptionalNillable(t, "engine_domain_id", engineDomainID)
	assertStringValidators(t, "engine_domain_id", engineDomainID, []struct {
		value string
		want  bool
	}{
		{"opaque-domain-id", true},
		{"", false},
		{strings.Repeat("i", 1025), false},
	})

	assertExactDefaults(t, fields, map[string]any{
		"desired_present":      true,
		"enabled":              true,
		"is_default":           false,
		"state":                "pending",
		"verification_status":  "pending",
		"dkim_status":          "unknown",
		"desired_generation":   int64(1),
		"desired_hash_version": int32(1),
		"attempt_count":        int32(0),
	})
	assertExactEnumValues(t, "state", fields["state"], []string{"pending", "provisioning", "dns_required", "verified", "error", "deleting"})
	assertExactEnumValues(t, "verification_status", fields["verification_status"], []string{"pending", "verified", "failed"})
	assertExactEnumValues(t, "dkim_status", fields["dkim_status"], []string{"unknown", "pending", "valid", "invalid"})
	assertExactEnumValues(t, "last_error_code", fields["last_error_code"], []string{
		"MAIL_DOMAIN_ENGINE_UNAVAILABLE", "MAIL_DOMAIN_APPLY_REJECTED", "MAIL_DOMAIN_OUTCOME_UNKNOWN", "MAIL_DOMAIN_DNS_INVALID",
	})
	assertOptionalNillable(t, "last_error_code", fields["last_error_code"])

	dnsRecords := fields["required_dns_records"]
	assertFieldType(t, "required_dns_records", dnsRecords, field.TypeJSON)
	assertRequired(t, "required_dns_records", dnsRecords)
	if got, want := dnsRecords.Default, mail.RequiredDNSRecords{}; !reflect.DeepEqual(got, want) {
		t.Fatalf("required_dns_records default = %#v, want concrete empty %#v", got, want)
	}

	assertFieldType(t, "desired_generation", fields["desired_generation"], field.TypeInt64)
	assertFieldType(t, "desired_hash_version", fields["desired_hash_version"], field.TypeInt32)
	assertInt64Validators(t, "desired_generation", fields["desired_generation"], []struct {
		value int64
		want  bool
	}{{0, false}, {1, true}})
	assertInt32Validators(t, "desired_hash_version", fields["desired_hash_version"], []struct {
		value int32
		want  bool
	}{{0, false}, {1, true}})
	assertInt32Validators(t, "attempt_count", fields["attempt_count"], []struct {
		value int32
		want  bool
	}{{-1, false}, {0, true}})

	for _, name := range []string{"desired_hash", "attempted_hash", "deployed_hash"} {
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
	for _, name := range []string{"attempted_generation", "attempted_hash_version", "attempted_hash", "deployed_generation", "deployed_hash_version", "deployed_hash", "last_error_code", "next_retry_at", "lease_token", "lease_until", "deleted_at"} {
		assertOptionalNillable(t, name, fields[name])
	}
	for _, name := range []string{"attempted_generation", "deployed_generation"} {
		assertFieldType(t, name, fields[name], field.TypeInt64)
		assertInt64Validators(t, name, fields[name], []struct {
			value int64
			want  bool
		}{{0, false}, {1, true}})
	}
	for _, name := range []string{"attempted_hash_version", "deployed_hash_version"} {
		assertFieldType(t, name, fields[name], field.TypeInt32)
		assertInt32Validators(t, name, fields[name], []struct {
			value int32
			want  bool
		}{{0, false}, {1, true}})
	}
	for _, name := range []string{"next_retry_at", "lease_until", "deleted_at"} {
		assertFieldType(t, name, fields[name], field.TypeTime)
	}
	assertFieldType(t, "lease_token", fields["lease_token"], field.TypeUUID)
	assertVersionAndTimestamps(t, entschema.MailDomain{}, fields)

	for _, forbidden := range []string{"dkim_private_key", "dkim_private_material", "password", "admin_token", "dns_zone_file"} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("MailDomain must not persist %q", forbidden)
		}
	}
}

func TestMailboxPolicySchema_HasExactFieldsRequiredValuesAndBounds(t *testing.T) {
	fields := fieldDescriptors((entschema.MailboxPolicy{}).Fields())
	assertExactNames(t, "MailboxPolicy fields", fieldNames(fields), []string{
		"created_at", "external_mail_enabled", "id", "instance_send_rate_per_minute", "junk_retention_days", "max_attachment_bytes",
		"max_message_bytes", "max_recipients", "organization_id", "send_rate_per_minute",
		"trash_retention_days", "updated_at",
	})

	for _, name := range []string{"id", "organization_id"} {
		assertFieldType(t, name, fields[name], field.TypeUUID)
		assertRequired(t, name, fields[name])
	}
	if !fields["id"].Immutable || !fields["organization_id"].Immutable {
		t.Fatal("MailboxPolicy id and organization_id must be immutable")
	}
	assertFieldType(t, "external_mail_enabled", fields["external_mail_enabled"], field.TypeBool)
	assertRequired(t, "external_mail_enabled", fields["external_mail_enabled"])

	for _, name := range []string{"external_mail_enabled", "max_message_bytes", "max_attachment_bytes", "max_recipients", "send_rate_per_minute", "trash_retention_days", "junk_retention_days"} {
		assertRequired(t, name, fields[name])
		if fields[name].Default != nil {
			t.Fatalf("%s default = %#v, want no schema default", name, fields[name].Default)
		}
	}
	assertRequired(t, "instance_send_rate_per_minute", fields["instance_send_rate_per_minute"])
	if got, want := fields["instance_send_rate_per_minute"].Default, int32(1_000); !reflect.DeepEqual(got, want) {
		t.Fatalf("instance_send_rate_per_minute default = %#v, want %#v", got, want)
	}
	for _, name := range []string{"max_message_bytes", "max_attachment_bytes"} {
		assertFieldType(t, name, fields[name], field.TypeInt64)
	}
	for _, name := range []string{"max_recipients", "send_rate_per_minute", "instance_send_rate_per_minute", "trash_retention_days", "junk_retention_days"} {
		assertFieldType(t, name, fields[name], field.TypeInt32)
	}
	assertInt64Validators(t, "max_message_bytes", fields["max_message_bytes"], []struct {
		value int64
		want  bool
	}{{1_048_575, false}, {1_048_576, true}, {52_428_800, true}, {52_428_801, false}})
	assertInt64Validators(t, "max_attachment_bytes", fields["max_attachment_bytes"], []struct {
		value int64
		want  bool
	}{{0, false}, {1, true}, {26_214_400, true}, {26_214_401, false}})
	assertInt32Validators(t, "max_recipients", fields["max_recipients"], []struct {
		value int32
		want  bool
	}{{0, false}, {1, true}, {300, true}, {301, false}})
	assertInt32Validators(t, "send_rate_per_minute", fields["send_rate_per_minute"], []struct {
		value int32
		want  bool
	}{{0, false}, {1, true}, {1_000, true}, {1_001, false}})
	assertInt32Validators(t, "instance_send_rate_per_minute", fields["instance_send_rate_per_minute"], []struct {
		value int32
		want  bool
	}{{0, false}, {1, true}, {1_000, true}, {1_001, false}})
	for _, name := range []string{"trash_retention_days", "junk_retention_days"} {
		assertInt32Validators(t, name, fields[name], []struct {
			value int32
			want  bool
		}{{0, false}, {1, true}, {3_650, true}, {3_651, false}})
	}
	assertVersionAndTimestamps(t, entschema.MailboxPolicy{}, fields)
}

func TestMailDomainAndMailboxPolicySchemas_HaveExactNamedChecks(t *testing.T) {
	assertExactChecks(t, (entschema.MailDomain{}).Annotations(), map[string]string{
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
	})
	assertExactChecks(t, (entschema.MailboxPolicy{}).Annotations(), map[string]string{
		"mailbox_policies_bounds": "max_message_bytes BETWEEN 1048576 AND 52428800 AND max_attachment_bytes BETWEEN 1 AND 26214400 AND max_attachment_bytes <= max_message_bytes AND max_recipients BETWEEN 1 AND 300 AND send_rate_per_minute BETWEEN 1 AND 1000 AND instance_send_rate_per_minute BETWEEN 1 AND 1000 AND trash_retention_days BETWEEN 1 AND 3650 AND junk_retention_days BETWEEN 1 AND 3650",
	})
}

func TestMailDomainAndMailboxPolicySchemas_HaveExactIndexesAndOrganizationEdges(t *testing.T) {
	assertExactIndexes(t, (entschema.MailDomain{}).Indexes(), []wantedIndex{
		{fields: []string{"domain"}, unique: true, where: "deleted_at IS NULL"},
		{fields: []string{"engine_provisioning_key"}, unique: true},
		{fields: []string{"engine_domain_id"}, unique: true, where: "engine_domain_id IS NOT NULL"},
		{fields: []string{"organization_id"}, unique: true, where: "deleted_at IS NULL AND is_default = TRUE"},
		{fields: []string{"state", "next_retry_at"}},
	})
	assertExactIndexes(t, (entschema.MailboxPolicy{}).Indexes(), []wantedIndex{{fields: []string{"organization_id"}, unique: true}})

	domainEdges := edgeDescriptors((entschema.MailDomain{}).Edges())
	assertExactNames(t, "MailDomain edges", edgeNames(domainEdges), []string{"accounts", "organization"})
	assertInverseRequiredOrganizationEdge(t, "MailDomain.organization", domainEdges["organization"], "mail_domains")
	accounts := domainEdges["accounts"]
	if accounts.Inverse || accounts.Unique || accounts.Required || accounts.RefName != "" {
		t.Fatalf("MailDomain.accounts = inverse:%t unique:%t required:%t ref:%q, want false/false/false/empty", accounts.Inverse, accounts.Unique, accounts.Required, accounts.RefName)
	}

	policyEdges := edgeDescriptors((entschema.MailboxPolicy{}).Edges())
	assertExactNames(t, "MailboxPolicy edges", edgeNames(policyEdges), []string{"organization"})
	assertInverseRequiredOrganizationEdge(t, "MailboxPolicy.organization", policyEdges["organization"], "mailbox_policy")

	organizationEdges := edgeDescriptors((entschema.Organization{}).Edges())
	assertExactNames(t, "Organization edges", edgeNames(organizationEdges), []string{"mail_accounts", "mail_domains", "mail_mutation_receipts", "mail_objects", "mail_send_rate_buckets", "mailbox_policy", "projects", "users"})
	for _, name := range []string{"projects", "users"} {
		assertOnDelete(t, "Organization."+name, organizationEdges[name].Annotations, entsql.NoAction)
	}
	mailDomains := organizationEdges["mail_domains"]
	if mailDomains.Inverse || mailDomains.Unique || mailDomains.Required || mailDomains.RefName != "" {
		t.Fatalf("Organization.mail_domains = inverse:%t unique:%t required:%t ref:%q, want false/false/false/empty", mailDomains.Inverse, mailDomains.Unique, mailDomains.Required, mailDomains.RefName)
	}
	mailboxPolicy := organizationEdges["mailbox_policy"]
	if mailboxPolicy.Inverse || !mailboxPolicy.Unique || mailboxPolicy.Required || mailboxPolicy.RefName != "" {
		t.Fatalf("Organization.mailbox_policy = inverse:%t unique:%t required:%t ref:%q, want false/true/false/empty", mailboxPolicy.Inverse, mailboxPolicy.Unique, mailboxPolicy.Required, mailboxPolicy.RefName)
	}
	mailAccounts := organizationEdges["mail_accounts"]
	if mailAccounts.Inverse || mailAccounts.Unique || mailAccounts.Required || mailAccounts.RefName != "" {
		t.Fatalf("Organization.mail_accounts = inverse:%t unique:%t required:%t ref:%q, want false/false/false/empty", mailAccounts.Inverse, mailAccounts.Unique, mailAccounts.Required, mailAccounts.RefName)
	}
	mailMutationReceipts := organizationEdges["mail_mutation_receipts"]
	if mailMutationReceipts.Inverse || mailMutationReceipts.Unique || mailMutationReceipts.Required || mailMutationReceipts.RefName != "" {
		t.Fatalf("Organization.mail_mutation_receipts = inverse:%t unique:%t required:%t ref:%q, want false/false/false/empty", mailMutationReceipts.Inverse, mailMutationReceipts.Unique, mailMutationReceipts.Required, mailMutationReceipts.RefName)
	}
	mailObjects := organizationEdges["mail_objects"]
	if mailObjects.Inverse || mailObjects.Unique || mailObjects.Required || mailObjects.RefName != "" {
		t.Fatalf("Organization.mail_objects = inverse:%t unique:%t required:%t ref:%q, want false/false/false/empty", mailObjects.Inverse, mailObjects.Unique, mailObjects.Required, mailObjects.RefName)
	}
}

func assertVersionAndTimestamps(t *testing.T, subject interface{ Mixin() []ent.Mixin }, fields map[string]*field.Descriptor) {
	t.Helper()
	if len(subject.Mixin()) != 1 {
		t.Fatalf("mixins = %d, want exactly VersionMixin", len(subject.Mixin()))
	}
	mixinFields := subject.Mixin()[0].Fields()
	if len(mixinFields) != 1 || mixinFields[0].Descriptor().Name != "version" {
		t.Fatalf("mixin fields = %#v, want only version", mixinFields)
	}
	version := mixinFields[0].Descriptor()
	assertFieldType(t, "version", version, field.TypeInt)
	if got, want := version.Default, 1; !reflect.DeepEqual(got, want) {
		t.Fatalf("version default = %#v, want %#v", got, want)
	}
	for _, name := range []string{"created_at", "updated_at"} {
		assertFieldType(t, name, fields[name], field.TypeTime)
		assertRequired(t, name, fields[name])
		if fields[name].Default == nil {
			t.Fatalf("%s must have a creation default", name)
		}
	}
	if !fields["created_at"].Immutable {
		t.Fatal("created_at must be immutable")
	}
	if fields["updated_at"].UpdateDefault == nil {
		t.Fatal("updated_at must have an update default")
	}
}

func assertExactChecks(t *testing.T, annotations []schema.Annotation, want map[string]string) {
	t.Helper()
	got := map[string]string{}
	for _, annotation := range annotations {
		sqlAnnotation, ok := annotation.(*entsql.Annotation)
		if !ok {
			continue
		}
		for name, predicate := range sqlAnnotation.Checks {
			got[name] = predicate
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("named checks = %#v, want %#v", got, want)
	}
}

type wantedIndex struct {
	fields []string
	unique bool
	where  string
}

func assertExactIndexes(t *testing.T, indexes []ent.Index, want []wantedIndex) {
	t.Helper()
	if len(indexes) != len(want) {
		t.Fatalf("index count = %d, want %d", len(indexes), len(want))
	}
	for index, wanted := range want {
		descriptor := indexes[index].Descriptor()
		if !reflect.DeepEqual(descriptor.Fields, wanted.fields) || descriptor.Unique != wanted.unique || indexWhere(descriptor.Annotations) != wanted.where {
			t.Fatalf("index[%d] = fields:%v unique:%t where:%q, want fields:%v unique:%t where:%q", index, descriptor.Fields, descriptor.Unique, indexWhere(descriptor.Annotations), wanted.fields, wanted.unique, wanted.where)
		}
	}
}

func indexWhere(annotations []schema.Annotation) string {
	for _, annotation := range annotations {
		if sqlAnnotation, ok := annotation.(*entsql.IndexAnnotation); ok {
			return sqlAnnotation.Where
		}
	}
	return ""
}

func assertInverseRequiredOrganizationEdge(t *testing.T, name string, descriptor *edge.Descriptor, ref string) {
	t.Helper()
	if descriptor == nil || !descriptor.Inverse || !descriptor.Unique || !descriptor.Required || descriptor.Field != "organization_id" || descriptor.RefName != ref {
		t.Fatalf("%s = %#v, want inverse required unique organization_id ref %q", name, descriptor, ref)
	}
	for _, annotation := range descriptor.Annotations {
		if sqlAnnotation, ok := annotation.(*entsql.Annotation); ok && sqlAnnotation.OnDelete == entsql.Cascade {
			return
		}
	}
	t.Fatalf("%s must cascade on delete", name)
}

func assertOnDelete(t *testing.T, name string, annotations []schema.Annotation, want any) {
	t.Helper()
	for _, annotation := range annotations {
		if sqlAnnotation, ok := annotation.(*entsql.Annotation); ok && reflect.DeepEqual(sqlAnnotation.OnDelete, want) {
			return
		}
	}
	t.Fatalf("%s OnDelete = missing, want %q", name, want)
}

func assertExactDefaults(t *testing.T, fields map[string]*field.Descriptor, want map[string]any) {
	t.Helper()
	for name, wanted := range want {
		descriptor := fields[name]
		if descriptor == nil {
			t.Fatalf("missing %s", name)
		}
		if !reflect.DeepEqual(descriptor.Default, wanted) {
			t.Fatalf("%s default = %#v, want %#v", name, descriptor.Default, wanted)
		}
	}
}

func assertFieldType(t *testing.T, name string, descriptor *field.Descriptor, want field.Type) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != want {
		t.Fatalf("%s type = %#v, want %v", name, descriptor, want)
	}
}

func assertRequired(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Optional || descriptor.Nillable {
		t.Fatalf("%s optional/nillable = %t/%t, want false/false", name, descriptor.Optional, descriptor.Nillable)
	}
}

func assertOptionalNillable(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil || !descriptor.Optional || !descriptor.Nillable {
		t.Fatalf("%s optional/nillable = %t/%t, want true/true", name, descriptor.Optional, descriptor.Nillable)
	}
}

func assertExactEnumValues(t *testing.T, name string, descriptor *field.Descriptor, want []string) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s", name)
	}
	got := make([]string, 0, len(descriptor.Enums))
	for _, value := range descriptor.Enums {
		got = append(got, value.V)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s enum values = %v, want %v", name, got, want)
	}
}

func assertStringValidators(t *testing.T, name string, descriptor *field.Descriptor, cases []struct {
	value string
	want  bool
}) {
	t.Helper()
	for _, testCase := range cases {
		if got := stringValidatorsAccept(descriptor.Validators, testCase.value); got != testCase.want {
			t.Fatalf("%s accepts %q = %t, want %t", name, testCase.value, got, testCase.want)
		}
	}
}

func assertBytesValidators(t *testing.T, name string, descriptor *field.Descriptor, cases []struct {
	value []byte
	want  bool
}) {
	t.Helper()
	for _, testCase := range cases {
		if got := bytesValidatorsAccept(descriptor.Validators, testCase.value); got != testCase.want {
			t.Fatalf("%s accepts %d bytes = %t, want %t", name, len(testCase.value), got, testCase.want)
		}
	}
}

func assertInt64Validators(t *testing.T, name string, descriptor *field.Descriptor, cases []struct {
	value int64
	want  bool
}) {
	t.Helper()
	for _, testCase := range cases {
		if got := int64ValidatorsAccept(descriptor.Validators, testCase.value); got != testCase.want {
			t.Fatalf("%s accepts %d = %t, want %t", name, testCase.value, got, testCase.want)
		}
	}
}

func assertInt32Validators(t *testing.T, name string, descriptor *field.Descriptor, cases []struct {
	value int32
	want  bool
}) {
	t.Helper()
	for _, testCase := range cases {
		if got := int32ValidatorsAccept(descriptor.Validators, testCase.value); got != testCase.want {
			t.Fatalf("%s accepts %d = %t, want %t", name, testCase.value, got, testCase.want)
		}
	}
}

func stringValidatorsAccept(validators []any, value string) bool {
	for _, rawValidator := range validators {
		if validator, ok := rawValidator.(func(string) error); ok && validator(value) != nil {
			return false
		}
	}
	return true
}

func bytesValidatorsAccept(validators []any, value []byte) bool {
	for _, rawValidator := range validators {
		if validator, ok := rawValidator.(func([]byte) error); ok && validator(value) != nil {
			return false
		}
	}
	return true
}

func int64ValidatorsAccept(validators []any, value int64) bool {
	for _, rawValidator := range validators {
		if validator, ok := rawValidator.(func(int64) error); ok && validator(value) != nil {
			return false
		}
	}
	return true
}

func int32ValidatorsAccept(validators []any, value int32) bool {
	for _, rawValidator := range validators {
		if validator, ok := rawValidator.(func(int32) error); ok && validator(value) != nil {
			return false
		}
	}
	return true
}

func fieldDescriptors(fields []ent.Field) map[string]*field.Descriptor {
	descriptors := make(map[string]*field.Descriptor, len(fields))
	for _, schemaField := range fields {
		descriptor := schemaField.Descriptor()
		descriptors[descriptor.Name] = descriptor
	}
	return descriptors
}

func edgeDescriptors(edges []ent.Edge) map[string]*edge.Descriptor {
	descriptors := make(map[string]*edge.Descriptor, len(edges))
	for _, schemaEdge := range edges {
		descriptor := schemaEdge.Descriptor()
		descriptors[descriptor.Name] = descriptor
	}
	return descriptors
}

func fieldNames(fields map[string]*field.Descriptor) []string {
	return mapKeys(fields)
}

func edgeNames(edges map[string]*edge.Descriptor) []string {
	return mapKeys(edges)
}

func mapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func assertExactNames(t *testing.T, label string, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}
