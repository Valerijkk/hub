package mail_test

import (
	"os"
	"strings"
	"testing"
)

func mailboxEntSubmissionReservationReconciliationSourceSlice(t *testing.T, source, start, end string) string {
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

func TestMailboxEntSubmissionReservationReconciliationCompatibilitySourceContract(t *testing.T) {
	repositoryBytes, err := os.ReadFile("../../internal/mail/mailbox_submission_repository.go")
	if err != nil {
		t.Fatalf("read submission repository source: %v", err)
	}
	reservationBytes, err := os.ReadFile("../../internal/mail/mailbox_ent_submission_reservation.go")
	if err != nil {
		t.Fatalf("read submission reservation source: %v", err)
	}
	repositorySource := string(repositoryBytes)
	reservationSource := string(reservationBytes)

	for _, required := range []string{
		"MailboxSubmissionErrorCodeReconciliationExhausted MailboxSubmissionErrorCode =",
		"MAIL_SUBMISSION_RECONCILIATION_EXHAUSTED",
		"case MailboxSubmissionErrorCodeOutcomeUnknown,",
		"MailboxSubmissionErrorCodeReconciliationRequired,",
		"MailboxSubmissionErrorCodeReconciliationExhausted,",
		"MailboxSubmissionErrorCodeMailboxUnavailable,",
	} {
		if !strings.Contains(repositorySource, required) {
			t.Fatalf("submission repository source is missing %q", required)
		}
	}

	reservation := mailboxEntSubmissionReservationReconciliationSourceSlice(t, reservationSource,
		"func (repository *MailboxEntRepository) ReserveMailboxSubmission",
		"func mailboxEntSubmissionRequestKind",
	)
	for _, required := range []string{
		"SetReconcileAttempts(0)",
		"created.ReconcileAttempts != 0 || created.NextReconcileAt != nil",
	} {
		if !strings.Contains(reservation, required) {
			t.Fatalf("submission reservation source is missing %q", required)
		}
	}

	rowValidation := mailboxEntSubmissionReservationReconciliationSourceSlice(t, reservationSource,
		"func mailboxEntSubmissionRowValid",
		"func mailboxEntSubmissionBindingFromRow",
	)
	for _, required := range []string{
		"row.ReconcileAttempts < 0 || row.ReconcileAttempts > 10",
		"row.NextReconcileAt != nil && !mailboxEntValidTime(*row.NextReconcileAt)",
		"func mailboxEntSubmissionReconciliationPendingCode",
		"row.ReconcileAttempts == 0 && row.NextReconcileAt == nil",
		"row.ReconcileAttempts >= 0 && row.ReconcileAttempts <= 9 && row.NextReconcileAt != nil",
		"row.ReconcileAttempts >= 1 && row.ReconcileAttempts <= 10 && row.NextReconcileAt == nil",
		"row.ReconcileAttempts == 10 && row.NextReconcileAt == nil && *errorCode == MailboxSubmissionErrorCodeReconciliationExhausted",
		"case MailboxSubmissionStatusAccepted:",
		"case MailboxSubmissionStatusSent:",
		"case MailboxSubmissionStatusFailed:",
	} {
		if !strings.Contains(rowValidation, required) {
			t.Fatalf("submission row validation is missing %q", required)
		}
	}

	unknown := mailboxEntSubmissionReservationReconciliationSourceSlice(t, rowValidation,
		"case MailboxSubmissionStatusUnknown:",
		"case MailboxSubmissionStatusFailed:",
	)
	if !strings.Contains(unknown, "if row.CommandLeaseToken == nil {") {
		t.Fatal("unknown validation must distinguish waiting and owned command-lease shapes")
	}
	failed := mailboxEntSubmissionReservationReconciliationSourceSlice(t, rowValidation,
		"case MailboxSubmissionStatusFailed:",
		"default:",
	)
	if strings.Contains(failed, "MailboxSubmissionErrorCodeReconciliationExhausted") {
		t.Fatal("reconciliation exhaustion must remain the exact unknown exhausted shape, not failed")
	}
}
