//go:build !qa && !release

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"time"

	mailmod "github.com/valerii/hub/backend/internal/mail"
)

const (
	devMailSourcePMEmail   = "pm@hub.local"
	devMailSourceDev1Email = "dev1@hub.local"

	devMailSourceReleaseFirstKey  = "release-checklist-first"
	devMailSourceReleaseReplyKey  = "release-checklist-reply"
	devMailSourceBrandKitKey      = "brand-kit"
	devMailSourceRetentionNoteKey = "retention-note"
)

var (
	errDevMailSourceInvalid = errors.New("dev mail source is invalid")
	errDevMailSourceLoad    = errors.New("dev mail source load failed")
)

// devMailSourceAccountHandle exposes only provider-safe handles needed by the
// future server adapter to synchronize this immutable demo source.
type devMailSourceAccountHandle struct {
	engineAccountID         string
	enginePrimaryIdentityID string
	folderIDs               map[mailmod.EngineFolderRole]string
}

// devMailSourceMessageHandle exposes no message content; it identifies one
// deterministic provider message after the source has been loaded.
type devMailSourceMessageHandle struct {
	accountEmail string
	messageID    string
	version      string
	folderID     string
	blobID       string
}

// devMailSourceResult reports the fixed source shape and narrow provider
// handles without retaining caller-owned control-plane maps or message data.
type devMailSourceResult struct {
	accounts map[string]devMailSourceAccountHandle
	messages map[string]devMailSourceMessageHandle

	accountsTotal    int
	foldersTotal     int
	threadsTotal     int
	messagesTotal    int
	attachmentsTotal int
}

type devMailSourcePlan struct {
	folders  []mailmod.EngineFolder
	messages []mailmod.EngineMessage
	blobs    []mailmod.EngineBlob
	result   devMailSourceResult
}

type devMailSourceBlobKey struct {
	accountID string
	blobID    string
}

// seedDevMailSource loads the fixed inbound-only Mail dataset into a
// server-owned fake provider. The caller owns the first-load guard; this
// function never creates Ent Mail projection rows or Calendar correlation.
func seedDevMailSource(
	ctx context.Context,
	engine *mailmod.FakeMailEngine,
	control devMailControlPlaneResult,
) (devMailSourceResult, error) {
	if !validDevMailControlContext(ctx) || engine == nil || !validDevMailControlResult(control) {
		return devMailSourceResult{}, errDevMailSourceInvalid
	}
	if err := ctx.Err(); err != nil {
		return devMailSourceResult{}, errDevMailSourceLoad
	}
	plan, err := newDevMailSourcePlan(control)
	if err != nil || !validDevMailSourcePlan(plan) {
		return devMailSourceResult{}, errDevMailSourceInvalid
	}

	for _, folder := range plan.folders {
		if !engine.PutFolder(folder) {
			return devMailSourceResult{}, errDevMailSourceLoad
		}
	}
	for _, blob := range plan.blobs {
		engine.PutBlob(blob)
	}
	for _, message := range plan.messages {
		engine.PutMessage(message)
	}
	if !readBackDevMailSource(ctx, engine, plan) {
		return devMailSourceResult{}, errDevMailSourceLoad
	}
	return cloneDevMailSourceResult(plan.result), nil
}

func newDevMailSourcePlan(control devMailControlPlaneResult) (devMailSourcePlan, error) {
	if !validDevMailControlResult(control) {
		return devMailSourcePlan{}, errDevMailSourceInvalid
	}
	pmAccount, pmOK := control.accounts[devMailSourcePMEmail]
	dev1Account, dev1OK := control.accounts[devMailSourceDev1Email]
	if !pmOK || !dev1OK || pmAccount.engineAccountID == dev1Account.engineAccountID {
		return devMailSourcePlan{}, errDevMailSourceInvalid
	}

	pmFolders := newDevMailSourceFolders("pm")
	dev1Folders := newDevMailSourceFolders("dev1")
	if len(pmFolders) != 6 || len(dev1Folders) != 6 {
		return devMailSourcePlan{}, errDevMailSourceInvalid
	}
	dev1InboxID, inboxOK := dev1Folders[mailmod.EngineFolderRoleInbox]
	dev1ArchiveID, archiveOK := dev1Folders[mailmod.EngineFolderRoleArchive]
	if !inboxOK || !archiveOK {
		return devMailSourcePlan{}, errDevMailSourceInvalid
	}

	releaseFirstAt := time.Date(2026, time.August, 31, 9, 0, 0, 0, time.UTC)
	releaseReplyAt := time.Date(2026, time.August, 31, 10, 0, 0, 0, time.UTC)
	brandKitAt := time.Date(2026, time.September, 1, 9, 0, 0, 0, time.UTC)
	retentionAt := time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC)

	const (
		releaseFirstInternetID = "<devmail-release-01@hub.local>"
		releaseReplyInternetID = "<devmail-release-02@hub.local>"
	)
	brandBrief := []byte("Use the shared release palette and include the Calendar handoff in the final review.\n")
	brandBriefHash := sha256.Sum256(brandBrief)
	brandBriefBlob := mailmod.EngineBlob{
		AccountID: dev1Account.engineAccountID,
		BlobID:    "devmail-blob-brand-brief-v1",
		MediaType: "text/plain",
		SizeBytes: int64(len(brandBrief)),
		Bytes:     append([]byte(nil), brandBrief...),
	}

	firstReleaseBody := "Please verify the Calendar mail release checklist before the handoff."
	replyReleaseBody := "Re: the Calendar mail checklist now includes the final handoff review."
	brandKitBody := "Attached is the brief for the mail demo brand kit."
	retentionBody := "This note is archived for the mail retention review."
	plan := devMailSourcePlan{
		folders: append(devMailSourceFolderRows(pmAccount.engineAccountID, pmFolders), devMailSourceFolderRows(dev1Account.engineAccountID, dev1Folders)...),
		blobs:   []mailmod.EngineBlob{brandBriefBlob},
		messages: []mailmod.EngineMessage{
			newDevMailSourceMessage(
				dev1Account.engineAccountID,
				"devmail-message-release-01-v1",
				"devmail-version-release-01-v1",
				releaseFirstInternetID,
				"",
				[]string{},
				mailmod.EngineAddress{Email: devMailSourcePMEmail},
				[]mailmod.EngineAddress{{Email: devMailSourceDev1Email}},
				[]mailmod.EngineAddress{{Email: "dev2@hub.local"}},
				"Release checklist: Calendar mail",
				"Calendar mail release checklist",
				firstReleaseBody,
				releaseFirstAt,
				dev1InboxID,
				mailmod.EngineMessageFlags{Seen: true},
				[]mailmod.EngineProjectionAttachment{},
			),
			newDevMailSourceMessage(
				dev1Account.engineAccountID,
				"devmail-message-release-02-v1",
				"devmail-version-release-02-v1",
				releaseReplyInternetID,
				releaseFirstInternetID,
				[]string{releaseFirstInternetID},
				mailmod.EngineAddress{Email: devMailSourcePMEmail},
				[]mailmod.EngineAddress{{Email: devMailSourceDev1Email}},
				[]mailmod.EngineAddress{{Email: "dev2@hub.local"}},
				"Re: Release checklist: Calendar mail",
				"Calendar mail checklist update",
				replyReleaseBody,
				releaseReplyAt,
				dev1InboxID,
				mailmod.EngineMessageFlags{Flagged: true},
				[]mailmod.EngineProjectionAttachment{},
			),
			newDevMailSourceMessage(
				dev1Account.engineAccountID,
				"devmail-message-brand-kit-v1",
				"devmail-version-brand-kit-v1",
				"<devmail-brand-kit@hub.local>",
				"",
				[]string{},
				mailmod.EngineAddress{Email: "design@hub.local"},
				[]mailmod.EngineAddress{{Email: devMailSourceDev1Email}},
				[]mailmod.EngineAddress{},
				"Brand kit for mail demo",
				"Brand kit brief attached",
				brandKitBody,
				brandKitAt,
				dev1InboxID,
				mailmod.EngineMessageFlags{Seen: true},
				[]mailmod.EngineProjectionAttachment{{
					BlobID:      brandBriefBlob.BlobID,
					Filename:    "mail-demo-brief.txt",
					MediaType:   "text/plain",
					SizeBytes:   brandBriefBlob.SizeBytes,
					Disposition: mailmod.EngineAttachmentDispositionAttachment,
					SHA256:      append([]byte(nil), brandBriefHash[:]...),
				}},
			),
			newDevMailSourceMessage(
				dev1Account.engineAccountID,
				"devmail-message-retention-note-v1",
				"devmail-version-retention-note-v1",
				"<devmail-retention-note@hub.local>",
				"",
				[]string{},
				mailmod.EngineAddress{Email: "records@hub.local"},
				[]mailmod.EngineAddress{{Email: devMailSourceDev1Email}},
				[]mailmod.EngineAddress{},
				"Archived: mail retention note",
				"Archived mail retention note",
				retentionBody,
				retentionAt,
				dev1ArchiveID,
				mailmod.EngineMessageFlags{Seen: true},
				[]mailmod.EngineProjectionAttachment{},
			),
		},
		result: devMailSourceResult{
			accounts: map[string]devMailSourceAccountHandle{
				devMailSourcePMEmail: {
					engineAccountID:         pmAccount.engineAccountID,
					enginePrimaryIdentityID: pmAccount.enginePrimaryIdentityID,
					folderIDs:               cloneDevMailSourceFolders(pmFolders),
				},
				devMailSourceDev1Email: {
					engineAccountID:         dev1Account.engineAccountID,
					enginePrimaryIdentityID: dev1Account.enginePrimaryIdentityID,
					folderIDs:               cloneDevMailSourceFolders(dev1Folders),
				},
			},
			messages: map[string]devMailSourceMessageHandle{
				devMailSourceReleaseFirstKey: {
					accountEmail: devMailSourceDev1Email,
					messageID:    "devmail-message-release-01-v1",
					version:      "devmail-version-release-01-v1",
					folderID:     dev1InboxID,
				},
				devMailSourceReleaseReplyKey: {
					accountEmail: devMailSourceDev1Email,
					messageID:    "devmail-message-release-02-v1",
					version:      "devmail-version-release-02-v1",
					folderID:     dev1InboxID,
				},
				devMailSourceBrandKitKey: {
					accountEmail: devMailSourceDev1Email,
					messageID:    "devmail-message-brand-kit-v1",
					version:      "devmail-version-brand-kit-v1",
					folderID:     dev1InboxID,
					blobID:       brandBriefBlob.BlobID,
				},
				devMailSourceRetentionNoteKey: {
					accountEmail: devMailSourceDev1Email,
					messageID:    "devmail-message-retention-note-v1",
					version:      "devmail-version-retention-note-v1",
					folderID:     dev1ArchiveID,
				},
			},
			accountsTotal:    2,
			foldersTotal:     12,
			threadsTotal:     3,
			messagesTotal:    4,
			attachmentsTotal: 1,
		},
	}
	if !validDevMailSourcePlan(plan) {
		return devMailSourcePlan{}, errDevMailSourceInvalid
	}
	return plan, nil
}

func newDevMailSourceFolders(owner string) map[mailmod.EngineFolderRole]string {
	return map[mailmod.EngineFolderRole]string{
		mailmod.EngineFolderRoleInbox:   "devmail-folder-" + owner + "-inbox-v1",
		mailmod.EngineFolderRoleSent:    "devmail-folder-" + owner + "-sent-v1",
		mailmod.EngineFolderRoleDrafts:  "devmail-folder-" + owner + "-drafts-v1",
		mailmod.EngineFolderRoleArchive: "devmail-folder-" + owner + "-archive-v1",
		mailmod.EngineFolderRoleTrash:   "devmail-folder-" + owner + "-trash-v1",
		mailmod.EngineFolderRoleJunk:    "devmail-folder-" + owner + "-junk-v1",
	}
}

func devMailSourceFolderRows(accountID string, folderIDs map[mailmod.EngineFolderRole]string) []mailmod.EngineFolder {
	return []mailmod.EngineFolder{
		{AccountID: accountID, FolderID: folderIDs[mailmod.EngineFolderRoleInbox], Name: "Inbox", Role: mailmod.EngineFolderRoleInbox},
		{AccountID: accountID, FolderID: folderIDs[mailmod.EngineFolderRoleSent], Name: "Sent", Role: mailmod.EngineFolderRoleSent},
		{AccountID: accountID, FolderID: folderIDs[mailmod.EngineFolderRoleDrafts], Name: "Drafts", Role: mailmod.EngineFolderRoleDrafts},
		{AccountID: accountID, FolderID: folderIDs[mailmod.EngineFolderRoleArchive], Name: "Archive", Role: mailmod.EngineFolderRoleArchive},
		{AccountID: accountID, FolderID: folderIDs[mailmod.EngineFolderRoleTrash], Name: "Trash", Role: mailmod.EngineFolderRoleTrash},
		{AccountID: accountID, FolderID: folderIDs[mailmod.EngineFolderRoleJunk], Name: "Junk", Role: mailmod.EngineFolderRoleJunk},
	}
}

func newDevMailSourceMessage(
	accountID, messageID, version, internetMessageID, inReplyTo string,
	references []string,
	from mailmod.EngineAddress,
	to, cc []mailmod.EngineAddress,
	subject, snippet, body string,
	receivedAt time.Time,
	folderID string,
	flags mailmod.EngineMessageFlags,
	attachments []mailmod.EngineProjectionAttachment,
) mailmod.EngineMessage {
	return mailmod.EngineMessage{
		Projection: mailmod.EngineMessageProjection{
			AccountID:             accountID,
			MessageID:             messageID,
			Version:               version,
			InternetMessageID:     internetMessageID,
			InReplyTo:             inReplyTo,
			References:            append([]string{}, references...),
			From:                  from,
			To:                    append([]mailmod.EngineAddress{}, to...),
			CC:                    append([]mailmod.EngineAddress{}, cc...),
			BCC:                   []mailmod.EngineAddress{},
			ReplyTo:               []mailmod.EngineAddress{},
			Subject:               subject,
			Snippet:               snippet,
			ReceivedAt:            &receivedAt,
			AuthenticationClass:   mailmod.MailAuthenticationClassUntrusted,
			SizeBytes:             int64(len(body)),
			Priority:              mailmod.EnginePriorityNormal,
			FolderIDs:             []string{folderID},
			Flags:                 flags,
			ReservedKeywords:      []mailmod.EngineKeywordMarker{},
			Attachments:           append([]mailmod.EngineProjectionAttachment{}, attachments...),
			Origin:                mailmod.EngineProjectionOriginInboundOrUnknown,
		},
		Body: mailmod.EngineMessageBody{Format: mailmod.EngineBodyPlain, PlainBody: body},
	}
}

func readBackDevMailSource(ctx context.Context, engine *mailmod.FakeMailEngine, plan devMailSourcePlan) bool {
	for _, expected := range plan.messages {
		actual, err := engine.GetMessage(ctx, mailmod.EngineMessageReadRequest{
			AccountID:       expected.Projection.AccountID,
			MessageID:       expected.Projection.MessageID,
			MaxContentBytes: mailmod.MaxEngineMessageBytes,
		})
		if err != nil || !reflect.DeepEqual(actual, expected) {
			return false
		}
	}
	for _, expected := range plan.blobs {
		actual, err := engine.GetBlob(ctx, mailmod.EngineBlobReadRequest{
			AccountID: expected.AccountID,
			BlobID:    expected.BlobID,
			MaxBytes:  mailmod.MaxEngineAttachmentBytes,
		})
		if err != nil || actual.AccountID != expected.AccountID || actual.BlobID != expected.BlobID || actual.MediaType != expected.MediaType || actual.SizeBytes != expected.SizeBytes || !bytes.Equal(actual.Bytes, expected.Bytes) {
			return false
		}
	}
	return true
}

func validDevMailSourcePlan(plan devMailSourcePlan) bool {
	if !validDevMailSourceResult(plan.result) || len(plan.folders) != plan.result.foldersTotal || len(plan.messages) != plan.result.messagesTotal || len(plan.blobs) != plan.result.attachmentsTotal {
		return false
	}
	seenOpaqueIDs := make(map[string]struct{}, len(plan.folders)+len(plan.messages)*2+len(plan.blobs))
	seenInternetMessageIDs := make(map[string]struct{}, len(plan.messages))
	scopedBlobs := make(map[devMailSourceBlobKey]mailmod.EngineBlob, len(plan.blobs))
	for _, folder := range plan.folders {
		if !validDevMailControlOpaqueID(folder.AccountID) || !validDevMailControlOpaqueID(folder.FolderID) || folder.ParentID != "" || !validDevMailSourceFolder(folder) || !addDevMailSourceOpaqueID(seenOpaqueIDs, folder.FolderID) {
			return false
		}
	}
	for _, blob := range plan.blobs {
		key := devMailSourceBlobKey{accountID: blob.AccountID, blobID: blob.BlobID}
		if !validDevMailControlOpaqueID(blob.AccountID) || !addDevMailSourceOpaqueID(seenOpaqueIDs, blob.BlobID) || blob.MediaType != "text/plain" || blob.SizeBytes != int64(len(blob.Bytes)) || blob.SizeBytes <= 0 || blob.SizeBytes > mailmod.MaxEngineAttachmentBytes {
			return false
		}
		if _, exists := scopedBlobs[key]; exists {
			return false
		}
		scopedBlobs[key] = blob
	}
	attachmentReferences := make(map[devMailSourceBlobKey]int, len(scopedBlobs))
	attachmentTotal := 0
	for _, message := range plan.messages {
		projection := message.Projection
		if !validDevMailControlOpaqueID(projection.AccountID) || !addDevMailSourceOpaqueID(seenOpaqueIDs, projection.MessageID) || !addDevMailSourceOpaqueID(seenOpaqueIDs, projection.Version) || !addDevMailSourceInternetMessageID(seenInternetMessageIDs, projection.InternetMessageID) || projection.SentAt != nil || projection.ReceivedAt == nil || !validDevMailControlNow(*projection.ReceivedAt) || projection.CalendarMethod != nil || len(projection.CalendarPartLocatorHash) != 0 || projection.Origin != mailmod.EngineProjectionOriginInboundOrUnknown || projection.AuthenticationClass != mailmod.MailAuthenticationClassUntrusted || len(projection.FolderIDs) != 1 || !validDevMailControlOpaqueID(projection.FolderIDs[0]) || projection.Attachments == nil || message.Body.Format != mailmod.EngineBodyPlain || message.Body.RichBody != "" || message.Body.PlainBody == "" {
			return false
		}
		for _, attachment := range projection.Attachments {
			key := devMailSourceBlobKey{accountID: projection.AccountID, blobID: attachment.BlobID}
			blob, exists := scopedBlobs[key]
			blobHash := sha256.Sum256(blob.Bytes)
			if !exists || attachment.AttachmentKey != "" || attachment.ContentID != "" || attachment.Filename != "mail-demo-brief.txt" || attachment.MediaType != blob.MediaType || blob.MediaType != "text/plain" || attachment.Disposition != mailmod.EngineAttachmentDispositionAttachment || attachment.SizeBytes != blob.SizeBytes || !bytes.Equal(attachment.SHA256, blobHash[:]) {
				return false
			}
			attachmentReferences[key]++
			attachmentTotal++
		}
	}
	if attachmentTotal != plan.result.attachmentsTotal || len(attachmentReferences) != len(scopedBlobs) {
		return false
	}
	for key := range scopedBlobs {
		if attachmentReferences[key] != 1 {
			return false
		}
	}
	return true
}

func validDevMailSourceResult(result devMailSourceResult) bool {
	if len(result.accounts) != result.accountsTotal || len(result.accounts) != 2 || len(result.messages) != result.messagesTotal || len(result.messages) != 4 || result.accountsTotal != 2 || result.foldersTotal != 12 || result.threadsTotal != 3 || result.messagesTotal != 4 || result.attachmentsTotal != 1 {
		return false
	}
	for _, email := range []string{devMailSourcePMEmail, devMailSourceDev1Email} {
		account, ok := result.accounts[email]
		if !ok || !validDevMailControlOpaqueID(account.engineAccountID) || !validDevMailControlOpaqueID(account.enginePrimaryIdentityID) || !validDevMailSourceFolders(account.folderIDs) {
			return false
		}
	}
	for _, key := range []string{devMailSourceReleaseFirstKey, devMailSourceReleaseReplyKey, devMailSourceBrandKitKey, devMailSourceRetentionNoteKey} {
		message, ok := result.messages[key]
		if !ok || message.accountEmail != devMailSourceDev1Email || !validDevMailControlOpaqueID(message.messageID) || !validDevMailControlOpaqueID(message.version) || !validDevMailControlOpaqueID(message.folderID) {
			return false
		}
	}
	return result.messages[devMailSourceBrandKitKey].blobID != "" &&
		validDevMailControlOpaqueID(result.messages[devMailSourceBrandKitKey].blobID) &&
		result.messages[devMailSourceReleaseFirstKey].blobID == "" &&
		result.messages[devMailSourceReleaseReplyKey].blobID == "" &&
		result.messages[devMailSourceRetentionNoteKey].blobID == ""
}

func validDevMailSourceFolder(folder mailmod.EngineFolder) bool {
	switch folder.Role {
	case mailmod.EngineFolderRoleInbox:
		return folder.Name == "Inbox"
	case mailmod.EngineFolderRoleSent:
		return folder.Name == "Sent"
	case mailmod.EngineFolderRoleDrafts:
		return folder.Name == "Drafts"
	case mailmod.EngineFolderRoleArchive:
		return folder.Name == "Archive"
	case mailmod.EngineFolderRoleTrash:
		return folder.Name == "Trash"
	case mailmod.EngineFolderRoleJunk:
		return folder.Name == "Junk"
	default:
		return false
	}
}

func validDevMailSourceFolders(folderIDs map[mailmod.EngineFolderRole]string) bool {
	if len(folderIDs) != 6 {
		return false
	}
	seen := make(map[string]struct{}, len(folderIDs))
	for _, role := range []mailmod.EngineFolderRole{
		mailmod.EngineFolderRoleInbox,
		mailmod.EngineFolderRoleSent,
		mailmod.EngineFolderRoleDrafts,
		mailmod.EngineFolderRoleArchive,
		mailmod.EngineFolderRoleTrash,
		mailmod.EngineFolderRoleJunk,
	} {
		folderID, ok := folderIDs[role]
		if !ok || !validDevMailControlOpaqueID(folderID) || !addDevMailSourceOpaqueID(seen, folderID) {
			return false
		}
	}
	return true
}

func addDevMailSourceOpaqueID(seen map[string]struct{}, value string) bool {
	if !validDevMailControlOpaqueID(value) {
		return false
	}
	if _, exists := seen[value]; exists {
		return false
	}
	seen[value] = struct{}{}
	return true
}

func addDevMailSourceInternetMessageID(seen map[string]struct{}, value string) bool {
	if value == "" || len(value) > 998 {
		return false
	}
	if _, exists := seen[value]; exists {
		return false
	}
	seen[value] = struct{}{}
	return true
}

func cloneDevMailSourceResult(result devMailSourceResult) devMailSourceResult {
	cloned := result
	cloned.accounts = make(map[string]devMailSourceAccountHandle, len(result.accounts))
	for email, account := range result.accounts {
		account.folderIDs = cloneDevMailSourceFolders(account.folderIDs)
		cloned.accounts[email] = account
	}
	cloned.messages = make(map[string]devMailSourceMessageHandle, len(result.messages))
	for key, message := range result.messages {
		cloned.messages[key] = message
	}
	return cloned
}

func cloneDevMailSourceFolders(source map[mailmod.EngineFolderRole]string) map[mailmod.EngineFolderRole]string {
	cloned := make(map[mailmod.EngineFolderRole]string, len(source))
	for role, folderID := range source {
		cloned[role] = folderID
	}
	return cloned
}
