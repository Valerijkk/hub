//go:build hubtest

package mail

import (
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent/schema"
)

// ExportInternalTokenHeader возвращает имя внутреннего auth-заголовка для hubtest.
func ExportInternalTokenHeader() string {
	return internalTokenHeader
}

// ExportGetEnv оборачивает getEnv для hubtest.
func ExportGetEnv(key, fallback string) string {
	return getEnv(key, fallback)
}

func ExportMailboxEntStableThreadID(accountID uuid.UUID, version ThreadKeyVersion, key ThreadKey) (uuid.UUID, error) {
	return mailboxEntStableThreadID(accountID, ThreadIdentity{KeyVersion: version, Key: key})
}

func ExportMailboxSubmissionLiveLeaseBusyUntil(databaseNow time.Time, mutationLeaseUntil, syncLeaseUntil *time.Time) (time.Time, bool) {
	return mailboxEntSubmissionLiveLeaseBusyUntil(databaseNow, mutationLeaseUntil, syncLeaseUntil)
}

func MailboxEntSubmissionFinalizationReferencesForTest(values []string) (schema.MailMessageReferences, error) {
	return mailboxEntSubmissionFinalizationReferences(values)
}
