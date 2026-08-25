package mail

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

const (
	mailboxSubmissionCleanAttachmentProofDomain = "hub.mail.submission.attachment.clean.proof.v1"
	mailboxSubmissionDispatchAuthorizationDomain = "hub.mail.submission.dispatch.authorization.v1"
	mailboxSubmissionAuthorizationVersion        = uint32(1)
	maxMailboxSubmissionCleanAttachments         = 50
)

var (
	// ErrMailboxSubmissionAuthorizationInvalid reports invalid content-free
	// authorization or attachment-proof input without exposing its values.
	ErrMailboxSubmissionAuthorizationInvalid = errors.New("mail submission authorization invalid")
)

// MailboxSubmissionCleanAttachment is one durably clean attachment used by a
// submission authorization. SHA256 is the scanner-locked blob digest, not a
// provider message-content digest.
type MailboxSubmissionCleanAttachment struct {
	ID     uuid.UUID
	SHA256 [sha256.Size]byte
}

// MailboxSubmissionDispatchAuthorizationHashInput is the complete
// content-free durable authorization binding for a single dispatch attempt.
type MailboxSubmissionDispatchAuthorizationHashInput struct {
	AccountID                  uuid.UUID
	SubmissionID               uuid.UUID
	LeaseID                    uuid.UUID
	CommandFence               int64
	LeaseUntil                 time.Time
	AuthorizedAt               time.Time
	EngineBindingHash          [sha256.Size]byte
	PreflightBindingHash       [sha256.Size]byte
	PreflightPolicyBindingHash [sha256.Size]byte
	DurablePolicyBindingHash   [sha256.Size]byte
	AttachmentCleanProofHash   [sha256.Size]byte
}

// MailboxSubmissionCleanAttachmentProofHash returns the canonical proof for
// zero through fifty clean attachments. It sorts a defensive copy by raw UUID
// bytes and is intentionally separate from the provider-content binding.
func MailboxSubmissionCleanAttachmentProofHash(attachments []MailboxSubmissionCleanAttachment) ([sha256.Size]byte, error) {
	if len(attachments) > maxMailboxSubmissionCleanAttachments {
		return [sha256.Size]byte{}, ErrMailboxSubmissionAuthorizationInvalid
	}

	canonical := append([]MailboxSubmissionCleanAttachment(nil), attachments...)
	for _, attachment := range canonical {
		if attachment.ID == uuid.Nil || mailboxSubmissionAuthorizationHashZero(attachment.SHA256) {
			return [sha256.Size]byte{}, ErrMailboxSubmissionAuthorizationInvalid
		}
	}
	sort.Slice(canonical, func(left, right int) bool {
		return bytes.Compare(canonical[left].ID[:], canonical[right].ID[:]) < 0
	})
	for index := 1; index < len(canonical); index++ {
		if canonical[index-1].ID == canonical[index].ID {
			return [sha256.Size]byte{}, ErrMailboxSubmissionAuthorizationInvalid
		}
	}

	hasher := sha256.New()
	_, _ = hasher.Write([]byte(mailboxSubmissionCleanAttachmentProofDomain))
	mailboxSubmissionAuthorizationWriteUint32(hasher, mailboxSubmissionAuthorizationVersion)
	mailboxSubmissionAuthorizationWriteUint32(hasher, uint32(len(canonical)))
	for _, attachment := range canonical {
		_, _ = hasher.Write(attachment.ID[:])
		_, _ = hasher.Write(attachment.SHA256[:])
	}

	return mailboxSubmissionAuthorizationHashResult(hasher), nil
}

// MailboxSubmissionDispatchAuthorizationHash returns the immutable binding for
// one authorization made while a specific dual lease and command fence are
// live. It never includes recipients, provider identifiers, or message bytes.
func MailboxSubmissionDispatchAuthorizationHash(input MailboxSubmissionDispatchAuthorizationHashInput) ([sha256.Size]byte, error) {
	if input.AccountID == uuid.Nil || input.SubmissionID == uuid.Nil || input.LeaseID == uuid.Nil || input.CommandFence < 1 ||
		!mailboxSubmissionAuthorizationTimeValid(input.LeaseUntil) || !mailboxSubmissionAuthorizationTimeValid(input.AuthorizedAt) ||
		!input.AuthorizedAt.Before(input.LeaseUntil) ||
		mailboxSubmissionAuthorizationHashZero(input.EngineBindingHash) ||
		mailboxSubmissionAuthorizationHashZero(input.PreflightBindingHash) ||
		mailboxSubmissionAuthorizationHashZero(input.PreflightPolicyBindingHash) ||
		mailboxSubmissionAuthorizationHashZero(input.DurablePolicyBindingHash) ||
		mailboxSubmissionAuthorizationHashZero(input.AttachmentCleanProofHash) {
		return [sha256.Size]byte{}, ErrMailboxSubmissionAuthorizationInvalid
	}

	hasher := sha256.New()
	_, _ = hasher.Write([]byte(mailboxSubmissionDispatchAuthorizationDomain))
	mailboxSubmissionAuthorizationWriteUint32(hasher, mailboxSubmissionAuthorizationVersion)
	_, _ = hasher.Write(input.AccountID[:])
	_, _ = hasher.Write(input.SubmissionID[:])
	_, _ = hasher.Write(input.LeaseID[:])
	mailboxSubmissionAuthorizationWriteUint64(hasher, uint64(input.CommandFence))
	mailboxSubmissionAuthorizationWriteUint64(hasher, uint64(input.LeaseUntil.Unix()))
	mailboxSubmissionAuthorizationWriteUint64(hasher, uint64(input.AuthorizedAt.Unix()))
	_, _ = hasher.Write(input.EngineBindingHash[:])
	_, _ = hasher.Write(input.PreflightBindingHash[:])
	_, _ = hasher.Write(input.PreflightPolicyBindingHash[:])
	_, _ = hasher.Write(input.DurablePolicyBindingHash[:])
	_, _ = hasher.Write(input.AttachmentCleanProofHash[:])

	return mailboxSubmissionAuthorizationHashResult(hasher), nil
}

func mailboxSubmissionAuthorizationTimeValid(value time.Time) bool {
	return !value.IsZero() && value == value.UTC() && value.Nanosecond() == 0 && value.Unix() >= 0
}

func mailboxSubmissionAuthorizationHashZero(value [sha256.Size]byte) bool {
	return value == [sha256.Size]byte{}
}

func mailboxSubmissionAuthorizationWriteUint32(writer interface{ Write([]byte) (int, error) }, value uint32) {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	_, _ = writer.Write(encoded[:])
}

func mailboxSubmissionAuthorizationWriteUint64(writer interface{ Write([]byte) (int, error) }, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = writer.Write(encoded[:])
}

func mailboxSubmissionAuthorizationHashResult(hasher interface{ Sum([]byte) []byte }) [sha256.Size]byte {
	var result [sha256.Size]byte
	copy(result[:], hasher.Sum(nil))
	return result
}
