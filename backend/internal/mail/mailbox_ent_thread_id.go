package mail

import "github.com/google/uuid"

func mailboxEntStableThreadID(accountID uuid.UUID, identity ThreadIdentity) (uuid.UUID, error) {
	if accountID == uuid.Nil || identity.KeyVersion != ThreadKeyVersionV1 || threadKeyIsZero(identity.Key) {
		return uuid.Nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var frame [1 + len(ThreadKey{})]byte
	frame[0] = byte(identity.KeyVersion)
	copy(frame[1:], identity.Key[:])
	return mailboxEntStableSyncID(accountID, "thread", frame[:])
}
