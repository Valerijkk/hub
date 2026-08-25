package mail_test

import (
	"crypto/sha256"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxSubmissionAuthorizationPublicContract(t *testing.T) {
	requireSubmissionAuthorizationFields(t, reflect.TypeOf(mail.MailboxSubmissionCleanAttachment{}), []submissionAuthorizationField{
		{name: "ID", typeValue: reflect.TypeOf(uuid.UUID{})},
		{name: "SHA256", typeValue: reflect.TypeOf([sha256.Size]byte{})},
	})
	requireSubmissionAuthorizationFields(t, reflect.TypeOf(mail.MailboxSubmissionDispatchAuthorizationHashInput{}), []submissionAuthorizationField{
		{name: "AccountID", typeValue: reflect.TypeOf(uuid.UUID{})},
		{name: "SubmissionID", typeValue: reflect.TypeOf(uuid.UUID{})},
		{name: "LeaseID", typeValue: reflect.TypeOf(uuid.UUID{})},
		{name: "CommandFence", typeValue: reflect.TypeOf(int64(0))},
		{name: "LeaseUntil", typeValue: reflect.TypeOf(time.Time{})},
		{name: "AuthorizedAt", typeValue: reflect.TypeOf(time.Time{})},
		{name: "EngineBindingHash", typeValue: reflect.TypeOf([sha256.Size]byte{})},
		{name: "PreflightBindingHash", typeValue: reflect.TypeOf([sha256.Size]byte{})},
		{name: "PreflightPolicyBindingHash", typeValue: reflect.TypeOf([sha256.Size]byte{})},
		{name: "DurablePolicyBindingHash", typeValue: reflect.TypeOf([sha256.Size]byte{})},
		{name: "AttachmentCleanProofHash", typeValue: reflect.TypeOf([sha256.Size]byte{})},
	})

	proofType := reflect.TypeOf(mail.MailboxSubmissionCleanAttachmentProofHash)
	if want := reflect.TypeOf(func([]mail.MailboxSubmissionCleanAttachment) ([sha256.Size]byte, error) { return [sha256.Size]byte{}, nil }); proofType != want {
		t.Fatalf("MailboxSubmissionCleanAttachmentProofHash type = %v, want %v", proofType, want)
	}
	authorizationType := reflect.TypeOf(mail.MailboxSubmissionDispatchAuthorizationHash)
	if want := reflect.TypeOf(func(mail.MailboxSubmissionDispatchAuthorizationHashInput) ([sha256.Size]byte, error) { return [sha256.Size]byte{}, nil }); authorizationType != want {
		t.Fatalf("MailboxSubmissionDispatchAuthorizationHash type = %v, want %v", authorizationType, want)
	}
	if got := mail.ErrMailboxSubmissionAuthorizationInvalid.Error(); got != "mail submission authorization invalid" {
		t.Fatalf("ErrMailboxSubmissionAuthorizationInvalid = %q, want fixed content-free literal", got)
	}

	inputType := reflect.TypeOf(mail.MailboxSubmissionDispatchAuthorizationHashInput{})
	for _, forbidden := range []string{"To", "CC", "BCC", "Recipients", "RecipientAddresses", "ProviderID", "MessageID", "ExpectedVersion", "MessageBytes"} {
		if _, exists := inputType.FieldByName(forbidden); exists {
			t.Fatalf("MailboxSubmissionDispatchAuthorizationHashInput must not expose %q", forbidden)
		}
	}
}

func TestMailboxSubmissionCleanAttachmentProofHash_CanonicalAndSensitive(t *testing.T) {
	attachments := submissionAuthorizationAttachments()
	original := append([]mail.MailboxSubmissionCleanAttachment(nil), attachments...)

	got, err := mail.MailboxSubmissionCleanAttachmentProofHash(attachments)
	if err != nil {
		t.Fatalf("MailboxSubmissionCleanAttachmentProofHash() error = %v", err)
	}
	if got == ([sha256.Size]byte{}) {
		t.Fatal("nonempty attachment proof is zero")
	}
	if !reflect.DeepEqual(attachments, original) {
		t.Fatalf("proof mutated input: got %#v, want %#v", attachments, original)
	}

	reversed := []mail.MailboxSubmissionCleanAttachment{attachments[1], attachments[0]}
	reversedHash, err := mail.MailboxSubmissionCleanAttachmentProofHash(reversed)
	if err != nil {
		t.Fatalf("MailboxSubmissionCleanAttachmentProofHash(reversed) error = %v", err)
	}
	if reversedHash != got {
		t.Fatalf("proof must be order-independent: reversed = %x, original = %x", reversedHash, got)
	}

	emptyFirst, err := mail.MailboxSubmissionCleanAttachmentProofHash(nil)
	if err != nil {
		t.Fatalf("MailboxSubmissionCleanAttachmentProofHash(nil) error = %v", err)
	}
	emptySecond, err := mail.MailboxSubmissionCleanAttachmentProofHash([]mail.MailboxSubmissionCleanAttachment{})
	if err != nil {
		t.Fatalf("MailboxSubmissionCleanAttachmentProofHash(empty) error = %v", err)
	}
	if emptyFirst == ([sha256.Size]byte{}) || emptyFirst != emptySecond {
		t.Fatalf("empty proof must be deterministic and nonzero: first=%x second=%x", emptyFirst, emptySecond)
	}
	if emptyFirst == got {
		t.Fatal("empty proof must differ from nonempty proof")
	}

	for _, testCase := range []struct {
		name        string
		attachments []mail.MailboxSubmissionCleanAttachment
	}{
		{name: "attachment id", attachments: []mail.MailboxSubmissionCleanAttachment{{ID: submissionAuthorizationUUID(3), SHA256: attachments[0].SHA256}, attachments[1]}},
		{name: "attachment hash", attachments: []mail.MailboxSubmissionCleanAttachment{{ID: attachments[0].ID, SHA256: submissionAuthorizationHash(9)}, attachments[1]}},
		{name: "attachment count", attachments: append(append([]mail.MailboxSubmissionCleanAttachment(nil), attachments...), mail.MailboxSubmissionCleanAttachment{ID: submissionAuthorizationUUID(3), SHA256: submissionAuthorizationHash(3)})},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			changed, err := mail.MailboxSubmissionCleanAttachmentProofHash(testCase.attachments)
			if err != nil {
				t.Fatalf("MailboxSubmissionCleanAttachmentProofHash() error = %v", err)
			}
			if changed == got {
				t.Fatal("proof did not change after a bound attachment value changed")
			}
		})
	}
}

func TestMailboxSubmissionCleanAttachmentProofHash_RejectsInvalidContentFreeInput(t *testing.T) {
	valid := submissionAuthorizationAttachments()
	tooMany := make([]mail.MailboxSubmissionCleanAttachment, 51)
	for index := range tooMany {
		tooMany[index] = mail.MailboxSubmissionCleanAttachment{ID: uuid.NewSHA1(uuid.Nil, []byte{byte(index + 1)}), SHA256: submissionAuthorizationHash(byte(index + 1))}
	}

	for _, testCase := range []struct {
		name        string
		attachments []mail.MailboxSubmissionCleanAttachment
	}{
		{name: "nil id", attachments: []mail.MailboxSubmissionCleanAttachment{{ID: uuid.Nil, SHA256: submissionAuthorizationHash(1)}}},
		{name: "zero hash", attachments: []mail.MailboxSubmissionCleanAttachment{{ID: submissionAuthorizationUUID(1)}}},
		{name: "duplicate id", attachments: []mail.MailboxSubmissionCleanAttachment{valid[0], valid[0]}},
		{name: "too many", attachments: tooMany},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := mail.MailboxSubmissionCleanAttachmentProofHash(testCase.attachments)
			if got != ([sha256.Size]byte{}) {
				t.Fatalf("invalid proof = %x, want zero", got)
			}
			requireSubmissionAuthorizationInvalid(t, err)
		})
	}
}

func TestMailboxSubmissionDispatchAuthorizationHash_DeterministicSensitiveAndPure(t *testing.T) {
	input := submissionAuthorizationInput()
	original := input

	got, err := mail.MailboxSubmissionDispatchAuthorizationHash(input)
	if err != nil {
		t.Fatalf("MailboxSubmissionDispatchAuthorizationHash() error = %v", err)
	}
	if got == ([sha256.Size]byte{}) {
		t.Fatal("dispatch authorization hash is zero")
	}
	second, err := mail.MailboxSubmissionDispatchAuthorizationHash(input)
	if err != nil {
		t.Fatalf("second MailboxSubmissionDispatchAuthorizationHash() error = %v", err)
	}
	if second != got {
		t.Fatalf("dispatch authorization hash is not deterministic: first=%x second=%x", got, second)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatalf("authorization hash mutated input: got %#v, want %#v", input, original)
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSubmissionDispatchAuthorizationHashInput)
	}{
		{name: "account", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AccountID = submissionAuthorizationUUID(4) }},
		{name: "submission", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.SubmissionID = submissionAuthorizationUUID(5) }},
		{name: "lease", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.LeaseID = submissionAuthorizationUUID(6) }},
		{name: "fence", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.CommandFence++ }},
		{name: "lease until", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.LeaseUntil = value.LeaseUntil.Add(time.Second) }},
		{name: "authorized at", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AuthorizedAt = value.AuthorizedAt.Add(time.Second) }},
		{name: "engine binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.EngineBindingHash = submissionAuthorizationHash(10) }},
		{name: "preflight binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.PreflightBindingHash = submissionAuthorizationHash(11) }},
		{name: "preflight policy binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.PreflightPolicyBindingHash = submissionAuthorizationHash(12) }},
		{name: "durable policy binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.DurablePolicyBindingHash = submissionAuthorizationHash(13) }},
		{name: "clean attachment proof", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AttachmentCleanProofHash = submissionAuthorizationHash(14) }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			changed := input
			testCase.mutate(&changed)
			changedHash, err := mail.MailboxSubmissionDispatchAuthorizationHash(changed)
			if err != nil {
				t.Fatalf("MailboxSubmissionDispatchAuthorizationHash() error = %v", err)
			}
			if changedHash == got {
				t.Fatal("authorization hash did not change after a bound value changed")
			}
		})
	}
}

func TestMailboxSubmissionDispatchAuthorizationHash_RejectsInvalidContentFreeInput(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSubmissionDispatchAuthorizationHashInput)
	}{
		{name: "nil account", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AccountID = uuid.Nil }},
		{name: "nil submission", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.SubmissionID = uuid.Nil }},
		{name: "nil lease", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.LeaseID = uuid.Nil }},
		{name: "zero fence", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.CommandFence = 0 }},
		{name: "negative fence", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.CommandFence = -1 }},
		{name: "zero engine binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.EngineBindingHash = [sha256.Size]byte{} }},
		{name: "zero preflight binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.PreflightBindingHash = [sha256.Size]byte{} }},
		{name: "zero preflight policy binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.PreflightPolicyBindingHash = [sha256.Size]byte{} }},
		{name: "zero durable policy binding", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.DurablePolicyBindingHash = [sha256.Size]byte{} }},
		{name: "zero clean attachment proof", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AttachmentCleanProofHash = [sha256.Size]byte{} }},
		{name: "zero lease until", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.LeaseUntil = time.Time{} }},
		{name: "zero authorized at", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AuthorizedAt = time.Time{} }},
		{name: "negative lease until", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.LeaseUntil = time.Unix(-1, 0).UTC() }},
		{name: "negative authorized at", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AuthorizedAt = time.Unix(-1, 0).UTC() }},
		{name: "noncanonical lease until zone", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.LeaseUntil = time.Date(2026, time.August, 25, 12, 1, 30, 0, time.FixedZone("UTC-alias", 0)) }},
		{name: "noncanonical authorized at zone", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AuthorizedAt = time.Date(2026, time.August, 25, 12, 0, 0, 0, time.FixedZone("UTC-alias", 0)) }},
		{name: "nanosecond lease until", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.LeaseUntil = value.LeaseUntil.Add(time.Nanosecond) }},
		{name: "nanosecond authorized at", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AuthorizedAt = value.AuthorizedAt.Add(time.Nanosecond) }},
		{name: "authorization equals expiry", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AuthorizedAt = value.LeaseUntil }},
		{name: "authorization after expiry", mutate: func(value *mail.MailboxSubmissionDispatchAuthorizationHashInput) { value.AuthorizedAt = value.LeaseUntil.Add(time.Second) }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := submissionAuthorizationInput()
			testCase.mutate(&input)
			got, err := mail.MailboxSubmissionDispatchAuthorizationHash(input)
			if got != ([sha256.Size]byte{}) {
				t.Fatalf("invalid authorization hash = %x, want zero", got)
			}
			requireSubmissionAuthorizationInvalid(t, err)
		})
	}
}

type submissionAuthorizationField struct {
	name      string
	typeValue reflect.Type
}

func requireSubmissionAuthorizationFields(t *testing.T, got reflect.Type, want []submissionAuthorizationField) {
	t.Helper()
	if got.Kind() != reflect.Struct || got.NumField() != len(want) {
		t.Fatalf("type %v has kind=%v fields=%d, want struct with %d fields", got, got.Kind(), got.NumField(), len(want))
	}
	for index, field := range want {
		gotField := got.Field(index)
		if gotField.Name != field.name || gotField.Type != field.typeValue || gotField.PkgPath != "" {
			t.Fatalf("type %v field[%d] = (%s, %v, %q), want (%s, %v, exported)", got, index, gotField.Name, gotField.Type, gotField.PkgPath, field.name, field.typeValue)
		}
	}
}

func requireSubmissionAuthorizationInvalid(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, mail.ErrMailboxSubmissionAuthorizationInvalid) {
		t.Fatalf("error = %v, want errors.Is(..., ErrMailboxSubmissionAuthorizationInvalid)", err)
	}
	if err == nil || err.Error() != "mail submission authorization invalid" {
		t.Fatalf("error text = %v, want fixed content-free text", err)
	}
	for _, forbidden := range []string{"recipient", "provider", "attachment", "hash"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("content-free error %q exposes %q", err, forbidden)
		}
	}
}

func submissionAuthorizationAttachments() []mail.MailboxSubmissionCleanAttachment {
	return []mail.MailboxSubmissionCleanAttachment{
		{ID: submissionAuthorizationUUID(1), SHA256: submissionAuthorizationHash(1)},
		{ID: submissionAuthorizationUUID(2), SHA256: submissionAuthorizationHash(2)},
	}
}

func submissionAuthorizationInput() mail.MailboxSubmissionDispatchAuthorizationHashInput {
	return mail.MailboxSubmissionDispatchAuthorizationHashInput{
		AccountID:                  submissionAuthorizationUUID(10),
		SubmissionID:               submissionAuthorizationUUID(11),
		LeaseID:                    submissionAuthorizationUUID(12),
		CommandFence:               7,
		LeaseUntil:                 time.Date(2026, time.August, 25, 12, 1, 30, 0, time.UTC),
		AuthorizedAt:               time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC),
		EngineBindingHash:          submissionAuthorizationHash(4),
		PreflightBindingHash:       submissionAuthorizationHash(5),
		PreflightPolicyBindingHash: submissionAuthorizationHash(6),
		DurablePolicyBindingHash:   submissionAuthorizationHash(7),
		AttachmentCleanProofHash:   submissionAuthorizationHash(8),
	}
}

func submissionAuthorizationUUID(value byte) uuid.UUID {
	return uuid.NewSHA1(uuid.Nil, []byte{value})
}

func submissionAuthorizationHash(value byte) [sha256.Size]byte {
	var result [sha256.Size]byte
	for index := range result {
		result[index] = value
	}
	return result
}
