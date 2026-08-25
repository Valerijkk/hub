package mail_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxSubmissionStaleRecoveryRepository = (*mail.MailboxEntRepository)(nil)
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) = (*mail.MailboxEntRepository).ScanDueMailboxSubmissionRecovery
var _ func(*mail.MailboxEntRepository, context.Context, mail.MailboxSubmissionStaleRecoveryRequest) (mail.MailboxSubmissionStaleRecoveryResult, error) = (*mail.MailboxEntRepository).RecoverStaleMailboxSubmission

func TestMailboxSubmissionRecoveryPublicContract(t *testing.T) {
	for _, testCase := range []struct {
		name string
		got  string
		want string
	}{
		{name: "noop", got: string(mail.MailboxSubmissionStaleRecoveryOutcomeNoop), want: "noop"},
		{name: "requeued", got: string(mail.MailboxSubmissionStaleRecoveryOutcomeRequeued), want: "requeued"},
		{name: "failed", got: string(mail.MailboxSubmissionStaleRecoveryOutcomeFailed), want: "failed"},
		{name: "reconciliation required", got: string(mail.MailboxSubmissionStaleRecoveryOutcomeReconciliationRequired), want: "reconciliation_required"},
	} {
		if testCase.got != testCase.want {
			t.Fatalf("%s outcome = %q, want %q", testCase.name, testCase.got, testCase.want)
		}
	}

	requireMailboxSubmissionRecoveryFields(t, reflect.TypeOf(mail.MailboxSubmissionRecoveryScanRequest{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "Limit", type_: reflect.TypeOf(int(0))},
	})
	requireMailboxSubmissionRecoveryFields(t, reflect.TypeOf(mail.MailboxSubmissionRecoveryDue{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "SubmissionID", type_: reflect.TypeOf(uuid.UUID{})},
		{name: "Status", type_: reflect.TypeOf(mail.MailboxSubmissionStatus(""))},
		{name: "LeaseUntil", type_: reflect.TypeOf(time.Time{})},
	})
	requireMailboxSubmissionRecoveryFields(t, reflect.TypeOf(mail.MailboxSubmissionRecoveryScanResult{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "SelectedAt", type_: reflect.TypeOf(time.Time{})},
		{name: "Submissions", type_: reflect.TypeOf([]mail.MailboxSubmissionRecoveryDue(nil))},
	})
	requireMailboxSubmissionRecoveryFields(t, reflect.TypeOf(mail.MailboxSubmissionStaleRecoveryRequest{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "SubmissionID", type_: reflect.TypeOf(uuid.UUID{})},
	})
	requireMailboxSubmissionRecoveryFields(t, reflect.TypeOf(mail.MailboxSubmissionStaleRecoveryResult{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "Outcome", type_: reflect.TypeOf(mail.MailboxSubmissionStaleRecoveryOutcome(""))},
		{name: "RetryAt", type_: reflect.TypeOf((*time.Time)(nil))},
		{name: "FailureCode", type_: reflect.TypeOf((*mail.MailboxSubmissionErrorCode)(nil))},
	})
}

func TestMailboxSubmissionRecoveryRejectsMalformedRequestsBeforeDatabaseAccess(t *testing.T) {
	var repository *mail.MailboxEntRepository
	for _, testCase := range []struct {
		name    string
		request mail.MailboxSubmissionRecoveryScanRequest
	}{
		{name: "zero limit", request: mail.MailboxSubmissionRecoveryScanRequest{}},
		{name: "negative limit", request: mail.MailboxSubmissionRecoveryScanRequest{Limit: -1}},
		{name: "above maximum", request: mail.MailboxSubmissionRecoveryScanRequest{Limit: 51}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := repository.ScanDueMailboxSubmissionRecovery(context.Background(), testCase.request)
			if !result.SelectedAt.IsZero() || result.Submissions != nil || !mailboxSubmissionRecoveryRepositoryInvalid(err) {
				t.Fatalf("ScanDueMailboxSubmissionRecovery(%#v) = %#v, %v; want zero result and invalid error", testCase.request, result, err)
			}
		})
	}

	result, err := repository.RecoverStaleMailboxSubmission(context.Background(), mail.MailboxSubmissionStaleRecoveryRequest{})
	if result != (mail.MailboxSubmissionStaleRecoveryResult{}) || !mailboxSubmissionRecoveryRepositoryInvalid(err) {
		t.Fatalf("RecoverStaleMailboxSubmission(zero) = %#v, %v; want zero result and invalid error", result, err)
	}
}

func requireMailboxSubmissionRecoveryFields(t *testing.T, candidate reflect.Type, want []struct {
	name  string
	type_ reflect.Type
}) {
	t.Helper()

	if candidate.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", candidate.Name(), candidate.NumField(), len(want))
	}
	for index, expected := range want {
		field := candidate.Field(index)
		if field.Name != expected.name || field.Type != expected.type_ || field.PkgPath != "" {
			t.Fatalf("%s field %d = %s %s exported=%t, want %s %s exported", candidate.Name(), index, field.Name, field.Type, field.PkgPath == "", expected.name, expected.type_)
		}
	}
}

func mailboxSubmissionRecoveryRepositoryInvalid(err error) bool {
	var repositoryErr *mail.MailboxSubmissionRepositoryError
	return errors.As(err, &repositoryErr) && repositoryErr.Kind() == mail.MailboxSubmissionRepositoryErrorInvalid
}
