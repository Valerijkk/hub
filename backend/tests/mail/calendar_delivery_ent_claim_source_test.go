package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestCalendarDeliveryEntClaimSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_calendar_delivery_claim.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, term := range []string{
		"repository.withMailboxEntTransaction",
		"mailboxEntDatabaseWallClock",
		"delivery.RecurrenceID != nil",
		"calendarattendee.RecurrenceIDIsNil()",
		"mailidentity.KindEQ(mailidentity.KindPrimary)",
		"mailidentity.SendAllowedEQ(true)",
		"ical.NormalizeITIPMessage",
		"ical.CanonicalITIPIdentity",
		"mailboxEntCheckedTimeAdd(databaseNow, calendarMailDeliveryLease)",
		"SetClaimGeneration(int64(generation))",
		"SetLeaseToken(uuid.UUID(token))",
		"SetLeaseUntil(leaseUntil)",
		"SetAttempts(delivery.Attempts + 1)",
	} {
		if !strings.Contains(source, term) {
			t.Fatalf("missing %q", term)
		}
	}
	for _, forbidden := range []string{
		"time.Now",
		"net/http",
		"http.",
		"github.com/nats-io/nats.go",
		"stalwart.",
		"Publish(",
		"BindCalendarMailDeliverySubmission",
		"ResolveCalendarMailDelivery",
		"CalendarMailDeliveryClaimOutcomeObserved",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("forbidden %q", forbidden)
		}
	}
}
