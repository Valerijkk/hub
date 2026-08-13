package calendar_test

import (
	"crypto/sha256"
	"testing"

	"entgo.io/ent/dialect/entsql"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarAttendee_InboundReplyReceiptsCascadeOnDelete(t *testing.T) {
	for _, attendeeEdge := range (entschema.CalendarAttendee{}).Edges() {
		descriptor := attendeeEdge.Descriptor()
		if descriptor.Name != "inbound_reply_receipts" {
			continue
		}
		if descriptor.Inverse {
			t.Fatal("inbound_reply_receipts must be an owning edge.To")
		}
		for _, annotation := range descriptor.Annotations {
			sqlAnnotation, ok := annotation.(*entsql.Annotation)
			if ok && sqlAnnotation.OnDelete == entsql.Cascade {
				return
			}
		}
		t.Fatal("inbound_reply_receipts must cascade on delete")
	}

	t.Fatal("missing inbound_reply_receipts edge")
}

func TestCalendarInboundReplyReceipt_PayloadHashRequiresExactSHA256Length(t *testing.T) {
	for _, receiptField := range (entschema.CalendarInboundReplyReceipt{}).Fields() {
		descriptor := receiptField.Descriptor()
		if descriptor.Name != "payload_hash" {
			continue
		}
		if got, want := descriptor.Size, sha256.Size; got != want {
			t.Fatalf("payload_hash size = %d, want %d", got, want)
		}

		shortRejections := 0
		for _, rawValidator := range descriptor.Validators {
			validator, ok := rawValidator.(func([]byte) error)
			if !ok {
				t.Fatalf("payload_hash validator has type %T, want func([]byte) error", rawValidator)
			}
			if err := validator(make([]byte, sha256.Size-1)); err != nil {
				shortRejections++
			}
			if err := validator(make([]byte, sha256.Size)); err != nil {
				t.Fatalf("payload_hash validator rejects %d-byte hash: %v", sha256.Size, err)
			}
		}
		if shortRejections < 2 {
			t.Fatalf("payload_hash has %d validators rejecting a %d-byte hash, want at least 2", shortRejections, sha256.Size-1)
		}
		return
	}

	t.Fatal("missing payload_hash field")
}

func TestCalendarInboundReplyReceipt_PayloadHashHasExactLengthCheck(t *testing.T) {
	const (
		checkName = "calendar_inbound_reply_receipts_payload_hash_length"
		predicate = "octet_length(payload_hash) = 32"
	)

	for _, annotation := range (entschema.CalendarInboundReplyReceipt{}).Annotations() {
		sqlAnnotation, ok := annotation.(*entsql.Annotation)
		if !ok {
			continue
		}
		if got, found := sqlAnnotation.Checks[checkName]; found {
			if got != predicate {
				t.Fatalf("%s = %q, want %q", checkName, got, predicate)
			}
			return
		}
	}

	t.Fatalf("missing named CHECK %s", checkName)
}
