package mail_test

import (
	"testing"

	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestMailSendRateBucketSchema_HasExactFieldsDefaultsAndBounds(t *testing.T) {
	bucket := entschema.MailSendRateBucket{}
	assertExactOrderedFieldNames(t, "MailSendRateBucket fields", bucket.Fields(), []string{
		"id", "organization_id", "scope", "user_id", "tokens", "last_refill_at", "refill_remainder_nanos", "version", "created_at", "updated_at",
	})
	fields := fieldDescriptors(bucket.Fields())
	assertMailDeliveryFieldMetadata(t, "MailSendRateBucket", bucket.Fields(), []mailDeliveryFieldMetadata{
		{"id", field.TypeUUID, false, false, true, deliveryUUIDDefault, false},
		{"organization_id", field.TypeUUID, true, true, true, deliveryNoDefault, false},
		{"scope", field.TypeEnum, false, false, true, deliveryNoDefault, false},
		{"user_id", field.TypeUUID, true, true, true, deliveryNoDefault, false},
		{"tokens", field.TypeInt32, false, false, false, deliveryNoDefault, false},
		{"last_refill_at", field.TypeTime, false, false, false, deliveryNoDefault, false},
		{"refill_remainder_nanos", field.TypeInt64, false, false, false, int64(0), false},
		{"version", field.TypeInt64, false, false, false, int64(1), false},
		{"created_at", field.TypeTime, false, false, true, deliveryNowDefault, false},
		{"updated_at", field.TypeTime, false, false, false, deliveryNowDefault, true},
	})
	assertExactEnumValues(t, "scope", fields["scope"], []string{"organization_instance", "user"})
	assertInt32Validators(t, "tokens", fields["tokens"], []struct {
		value int32
		want  bool
	}{{-1, false}, {0, true}, {1_000, true}, {1_001, false}})
	assertInt64Validators(t, "refill_remainder_nanos", fields["refill_remainder_nanos"], []struct {
		value int64
		want  bool
	}{{-1, false}, {0, true}, {59_999_999_999, true}, {60_000_000_000, false}})
	assertInt64Validators(t, "version", fields["version"], []struct {
		value int64
		want  bool
	}{{0, false}, {1, true}})
	assertTimestampsWithoutMixin(t, fields)

	for _, forbidden := range []string{"mail_account_id", "deployment_id", "instance_id"} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("MailSendRateBucket must not persist forbidden %q", forbidden)
		}
	}
}

func TestMailSendRateBucketSchema_HasExactChecksIndexesAndRelations(t *testing.T) {
	bucket := entschema.MailSendRateBucket{}
	assertExactChecks(t, bucket.Annotations(), map[string]string{
		"mail_send_rate_buckets_organization_nonzero": "COALESCE(organization_id IS NULL OR organization_id <> '00000000-0000-0000-0000-000000000000'::uuid, FALSE)",
		"mail_send_rate_buckets_user_nonzero":         "COALESCE(user_id IS NULL OR user_id <> '00000000-0000-0000-0000-000000000000'::uuid, FALSE)",
		"mail_send_rate_buckets_scope_owner":          "COALESCE((scope = 'organization_instance' AND organization_id IS NOT NULL AND user_id IS NULL) OR (scope = 'user' AND organization_id IS NULL AND user_id IS NOT NULL), FALSE)",
		"mail_send_rate_buckets_tokens":               "COALESCE(tokens BETWEEN 0 AND 1000, FALSE)",
		"mail_send_rate_buckets_refill_remainder":     "COALESCE(refill_remainder_nanos BETWEEN 0 AND 59999999999, FALSE)",
		"mail_send_rate_buckets_version":              "COALESCE(version >= 1, FALSE)",
	})
	assertExactIndexes(t, bucket.Indexes(), []wantedIndex{
		{fields: []string{"organization_id"}, unique: true, where: "scope = 'organization_instance' AND organization_id IS NOT NULL"},
		{fields: []string{"user_id"}, unique: true, where: "scope = 'user' AND user_id IS NOT NULL"},
	})
	for position, storageKey := range []string{"mail_send_rate_buckets_organization_instance_uidx", "mail_send_rate_buckets_user_uidx"} {
		if got := bucket.Indexes()[position].Descriptor().StorageKey; got != storageKey {
			t.Fatalf("MailSendRateBucket index[%d] storage key = %q, want %q", position, got, storageKey)
		}
	}

	bucketEdges := edgeDescriptors(bucket.Edges())
	assertExactNames(t, "MailSendRateBucket edges", edgeNames(bucketEdges), []string{"organization", "user"})
	organizationEdge := bucketEdges["organization"]
	if organizationEdge == nil || !organizationEdge.Inverse || !organizationEdge.Unique || organizationEdge.Required || organizationEdge.Field != "organization_id" || organizationEdge.RefName != "mail_send_rate_buckets" {
		t.Fatalf("MailSendRateBucket.organization = %#v, want inverse optional unique organization_id ref mail_send_rate_buckets", organizationEdge)
	}
	assertOnDelete(t, "MailSendRateBucket.organization", organizationEdge.Annotations, entsql.NoAction)
	userEdge := bucketEdges["user"]
	if userEdge == nil || !userEdge.Inverse || !userEdge.Unique || userEdge.Required || userEdge.Field != "user_id" || userEdge.RefName != "mail_send_rate_buckets" {
		t.Fatalf("MailSendRateBucket.user = %#v, want inverse optional unique user_id ref mail_send_rate_buckets", userEdge)
	}
	assertOnDelete(t, "MailSendRateBucket.user", userEdge.Annotations, entsql.NoAction)

	organizationEdges := edgeDescriptors((entschema.Organization{}).Edges())
	assertMailSendRateBucketInverse(t, "Organization.mail_send_rate_buckets", organizationEdges["mail_send_rate_buckets"])
	userEdges := edgeDescriptors((entschema.User{}).Edges())
	assertMailSendRateBucketInverse(t, "User.mail_send_rate_buckets", userEdges["mail_send_rate_buckets"])
}

func assertMailSendRateBucketInverse(t *testing.T, name string, descriptor *edge.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Inverse || descriptor.Unique || descriptor.Required || descriptor.RefName != "" {
		t.Fatalf("%s = %#v, want plain to-many inverse collection", name, descriptor)
	}
}
