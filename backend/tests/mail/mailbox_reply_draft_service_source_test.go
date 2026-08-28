package mail_test

import (
	"os"
	"strings"
	"testing"
)

func TestMailboxReplyDraftSourceEligibleSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/mailbox_reply_draft_service.go")
	if err != nil {
		t.Fatalf("read reply-draft service: %v", err)
	}

	eligible := mailboxReplyDraftSourceEligibleSlice(t, string(sourceBytes))
	for _, gate := range []string{
		"!snapshot.summary.IsDraft",
		"snapshot.projectionOrigin == EngineProjectionOriginInboundOrUnknown",
		"snapshot.summary.CalendarMethod == nil",
		"len(snapshot.calendarPartLocatorHash) == 0",
	} {
		if !strings.Contains(eligible, gate) {
			t.Fatalf("mailboxReplyDraftSourceEligible must retain %q", gate)
		}
	}

	for _, forbidden := range []string{"CalendarEventID", "calendar_event_id"} {
		if strings.Contains(eligible, forbidden) {
			t.Fatalf("mailboxReplyDraftSourceEligible must not reference %q", forbidden)
		}
	}
}

func mailboxReplyDraftSourceEligibleSlice(t *testing.T, source string) string {
	t.Helper()

	const (
		start = "func mailboxReplyDraftSourceEligible"
		end   = "func mailboxReplyDraftRecipients"
	)

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("%s not found", start)
	}

	endOffset := strings.Index(source[startIndex:], end)
	if endOffset < 0 {
		t.Fatalf("end of %s not found", start)
	}

	return source[startIndex : startIndex+endOffset]
}
