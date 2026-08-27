package mail_test

import (
	"os"
	"strings"
	"testing"
)

func calendarDeliveryEntScanSourceSlice(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing start marker %q", start)
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset < 0 {
		t.Fatalf("source is missing end marker %q after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}

func requireCalendarDeliveryEntScanSourceContains(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(source, fragment) {
			t.Fatalf("source is missing %q", fragment)
		}
	}
}

func TestCalendarDeliveryEntDueScanSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_scan.go")
	if err != nil {
		t.Fatalf("read calendar delivery Ent scan source: %v", err)
	}
	source := string(sourceBytes)
	scan := calendarDeliveryEntScanSourceSlice(t, source,
		"func (repository *MailboxEntRepository) ScanDueCalendarMailDeliveries",
		"type mailboxEntCalendarDueDelivery",
	)

	requireCalendarDeliveryEntScanSourceContains(t, scan,
		"repository == nil || repository.client == nil || ctx == nil || ctx.Err() != nil || !validCalendarMailDeliveryDueScanRequest(request)",
		"repository.withMailboxEntReadSnapshot",
		"mailboxEntDatabaseWallClock",
		"calendarinvitedelivery.StateIn(calendarinvitedelivery.StatePending, calendarinvitedelivery.StateUnknown)",
		"calendarinvitedelivery.NextAttemptAtNotNil()",
		"calendarinvitedelivery.NextAttemptAtLTE(databaseWallClock)",
		"calendarinvitedelivery.StateEQ(calendarinvitedelivery.StateSubmitting)",
		"calendarinvitedelivery.LeaseUntilNotNil()",
		"calendarinvitedelivery.LeaseUntilLTE(databaseWallClock)",
		"calendarreplydelivery.StateIn(calendarreplydelivery.StatePending, calendarreplydelivery.StateUnknown)",
		"calendarreplydelivery.NextAttemptAtNotNil()",
		"calendarreplydelivery.NextAttemptAtLTE(databaseWallClock)",
		"calendarreplydelivery.StateEQ(calendarreplydelivery.StateSubmitting)",
		"calendarreplydelivery.LeaseUntilNotNil()",
		"calendarreplydelivery.LeaseUntilLTE(databaseWallClock)",
		"Order(ent.Asc(calendarinvitedelivery.FieldNextAttemptAt), ent.Asc(calendarinvitedelivery.FieldID))",
		"Order(ent.Asc(calendarinvitedelivery.FieldLeaseUntil), ent.Asc(calendarinvitedelivery.FieldID))",
		"Order(ent.Asc(calendarreplydelivery.FieldNextAttemptAt), ent.Asc(calendarreplydelivery.FieldID))",
		"Order(ent.Asc(calendarreplydelivery.FieldLeaseUntil), ent.Asc(calendarreplydelivery.FieldID))",
		"NewCalendarMailDeliveryID",
		"CalendarMailDeliveryKindInvite",
		"CalendarMailDeliveryKindReply",
		"sort.SliceStable",
		"left.dueAt.Before(right.dueAt)",
		"leftUUID < rightUUID",
		"left.deliveryID.Kind() < right.deliveryID.Kind()",
		"make([]CalendarMailDeliveryID, 0, request.Limit)",
		"mailboxEntCalendarDueScanResultCopy(result)",
	)

	if got := strings.Count(scan, "transaction.CalendarInviteDelivery.Query()"); got != 2 {
		t.Fatalf("CalendarInviteDelivery due query count = %d, want 2", got)
	}
	if got := strings.Count(scan, "transaction.CalendarReplyDelivery.Query()"); got != 2 {
		t.Fatalf("CalendarReplyDelivery due query count = %d, want 2", got)
	}
	if got := strings.Count(scan, ".Limit(request.Limit)"); got != 4 {
		t.Fatalf("bounded stream limit count = %d, want 4", got)
	}
	for _, forbidden := range []string{
		"ForUpdate()",
		".Create()",
		".Update()",
		".Delete()",
		"time.Now",
		"net/http",
		"http.",
		"nats.",
		"Publish(",
		"stalwart.",
	} {
		if strings.Contains(scan, forbidden) {
			t.Fatalf("due scan contains forbidden lock, write, local clock, or external I/O %q", forbidden)
		}
	}
}

func TestCalendarDeliveryEntDueScanResultCopyStaysNonNil(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_scan.go")
	if err != nil {
		t.Fatalf("read calendar delivery Ent scan source: %v", err)
	}
	copy := calendarDeliveryEntScanSourceSlice(t, string(sourceBytes),
		"func mailboxEntCalendarDueScanResultCopy",
		"}",
	)
	if !strings.Contains(copy, "append(make([]CalendarMailDeliveryID, 0, len(result.DeliveryIDs)), result.DeliveryIDs...)") {
		t.Fatal("due scan result copy must preserve a non-nil empty list")
	}
}
