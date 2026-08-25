package mail

import (
	"context"
	"errors"
	"time"
)

var errMailboxDraftCalendarCreatePersistedBlobNotFound = errors.New("mailbox draft calendar create persisted blob not found")

type mailboxDraftCalendarCreateEngine interface {
	uploadCalendarDraftCreateMIME(context.Context, EngineDraft, *calendarMIMEPart) (mailboxDraftCalendarCreateUpload, error)
	importCalendarDraftCreateMIME(context.Context, mailboxDraftCalendarCreateImport) (EngineMessage, error)
}

type mailboxDraftCalendarCreateUpload struct {
	blobID    string
	sizeBytes int64
}

type mailboxDraftCalendarCreateImport struct {
	accountID  string
	draftKey   string
	mutationID string
	blobID     string
	sizeBytes  int64
	receivedAt time.Time
}

func validMailboxDraftCalendarCreateUpload(value mailboxDraftCalendarCreateUpload) bool {
	return validOpaqueID(value.blobID) &&
		value.sizeBytes >= 1 &&
		value.sizeBytes <= MaxEngineMessageBytes
}

func validMailboxDraftCalendarCreateImport(value mailboxDraftCalendarCreateImport) bool {
	return validAccountID(value.accountID) &&
		validSafeText(value.draftKey, MaxEngineOpaqueIDBytes, true) &&
		validSafeText(value.mutationID, MaxEngineOpaqueIDBytes, true) &&
		validOpaqueID(value.blobID) &&
		value.sizeBytes >= 1 &&
		value.sizeBytes <= MaxEngineMessageBytes &&
		mailboxDraftCalendarCreateReceivedAtValid(value.receivedAt)
}

func mailboxDraftCalendarCreateReceivedAtValid(value time.Time) bool {
	return !value.IsZero() &&
		value.Location() == time.UTC &&
		value.Nanosecond() == 0
}
