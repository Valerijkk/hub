package mail_test

import (
	"reflect"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxEntSubmissionFinalizationReferencesForTest_PreservesEmptyNonNilSlice(t *testing.T) {
	references, err := mail.MailboxEntSubmissionFinalizationReferencesForTest([]string{})
	if err != nil {
		t.Fatalf("MailboxEntSubmissionFinalizationReferencesForTest() error = %v", err)
	}
	if references == nil {
		t.Fatal("MailboxEntSubmissionFinalizationReferencesForTest() returned nil references for non-nil empty input")
	}
	if len(references) != 0 {
		t.Fatalf("MailboxEntSubmissionFinalizationReferencesForTest() length = %d, want 0", len(references))
	}
	if err := references.Validate(); err != nil {
		t.Fatalf("empty references Validate() error = %v", err)
	}
}

func TestMailboxEntSubmissionFinalizationReferencesForTest_PreservesValuesAndCopiesInput(t *testing.T) {
	input := []string{"<first@example.test>", "<second@example.test>"}
	references, err := mail.MailboxEntSubmissionFinalizationReferencesForTest(input)
	if err != nil {
		t.Fatalf("MailboxEntSubmissionFinalizationReferencesForTest() error = %v", err)
	}
	if !reflect.DeepEqual([]string(references), input) {
		t.Fatalf("MailboxEntSubmissionFinalizationReferencesForTest() = %#v, want %#v", references, input)
	}
	input[0] = "<changed@example.test>"
	if references[0] != "<first@example.test>" {
		t.Fatalf("references changed with input mutation: got %q, want %q", references[0], "<first@example.test>")
	}
	if err := references.Validate(); err != nil {
		t.Fatalf("references Validate() error = %v", err)
	}
}

func TestMailboxEntSubmissionFinalizationReferencesForTest_RejectsNilInput(t *testing.T) {
	_, err := mail.MailboxEntSubmissionFinalizationReferencesForTest(nil)
	if err == nil {
		t.Fatal("MailboxEntSubmissionFinalizationReferencesForTest() accepted nil input")
	}
}

func TestMailboxEntSubmissionFinalizationReferencesForTest_RejectsInvalidMessageID(t *testing.T) {
	_, err := mail.MailboxEntSubmissionFinalizationReferencesForTest([]string{"not-a-message-id"})
	if err == nil {
		t.Fatal("MailboxEntSubmissionFinalizationReferencesForTest() accepted an invalid message ID")
	}
}
