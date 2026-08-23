package mail

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

// MailboxSyncPolicyInput is the caller-resolved bounded state for one Mail sync page.
type MailboxSyncPolicyInput struct {
	Page        MailboxSyncPolicyPage
	Folders     []MailboxSyncPolicyFolder
	Messages    []MailboxSyncPolicyMessage
	Destroyed   []MailboxSyncPolicyDestroyed
	Labels      []MailboxSyncPolicyLabelDefinition
	SyncEpoch   int64
	EventCutoff time.Time
	Event       MailboxSyncPolicyEventContext
}

// MailboxSyncPolicyPage is an exact-one engine page union.
type MailboxSyncPolicyPage struct {
	Change      *EngineChangePage
	Enumeration *EngineEnumerationPage
}

// MailboxSyncPolicyFolder binds one engine folder snapshot to its Hub projection identifier.
type MailboxSyncPolicyFolder struct {
	Snapshot EngineFolder
	Folder   MailboxEventID
}

// MailboxSyncPolicyMessage is caller-resolved state for one live engine projection.
type MailboxSyncPolicyMessage struct {
	EngineMessageID string
	Message         MailboxEventID
	Thread          MailboxSyncPolicyThreadResolution
	Previous        *MailboxSyncPolicyPreviousMessage
	Provenance      MailboxSyncPolicyProvenance
	ReceivedEventID uuid.UUID
	ChangedEvent    MailboxSyncPolicyEventReservation
}

// MailboxSyncPolicyThreadResolution contains the complete caller-resolved thread state.
type MailboxSyncPolicyThreadResolution struct {
	Thread                 MailboxEventID
	KeyVersion             ThreadKeyVersion
	Key                    ThreadKey
	Ancestors              []ThreadAncestor
	AncestorLookupComplete bool
}

// MailboxSyncPolicyPreviousMessage is the persisted state needed for a safe projection delta.
type MailboxSyncPolicyPreviousMessage struct {
	Thread          MailboxEventID
	FolderIDs       []MailboxEventID
	Flags           EngineMessageFlags
	Labels          []MailboxSyncPolicyLabelState
	ReceivedEventID *uuid.UUID
}

// MailboxSyncPolicyProvenance has zero or one confirmed owner proof.
type MailboxSyncPolicyProvenance struct {
	ConfirmedSubmission *MailboxSyncPolicyProvenanceProof
	ConfirmedDraft      *MailboxSyncPolicyProvenanceProof
}

// MailboxSyncPolicyProvenanceProof binds a Hub relation to one exact engine message identifier.
type MailboxSyncPolicyProvenanceProof struct {
	Relation        MailboxEventID
	EngineMessageID string
}

// MailboxSyncPolicyEventReservation is a caller-owned outbox reservation.
type MailboxSyncPolicyEventReservation struct {
	ID        uuid.UUID
	Committed bool
}

// MailboxSyncPolicyDestroyed is caller-resolved state for one destroyed engine message.
type MailboxSyncPolicyDestroyed struct {
	EngineMessageID string
	Projected       bool
	Message         MailboxEventID
	Thread          MailboxEventID
	Labels          []MailboxSyncPolicyLabelState
	ChangedEvent    MailboxSyncPolicyEventReservation
}

// MailboxSyncPolicyLabelDefinition maps one known engine keyword to one Hub label.
type MailboxSyncPolicyLabelDefinition struct {
	Label         MailboxEventID
	EngineKeyword EngineKeywordMarker
}

// MailboxSyncPolicyLabelState is the source-separated membership state for one Hub label.
type MailboxSyncPolicyLabelState struct {
	Label      MailboxEventID
	Manual     bool
	EngineRule bool
}

// MailboxSyncPolicyEventContext contains the public envelope metadata shared by page events.
type MailboxSyncPolicyEventContext struct {
	Mailbox       MailboxEventID
	OrganizationID uuid.UUID
	ActorID       *uuid.UUID
	OccurredAt    time.Time
	CorrelationID string
}

// MailboxSyncPolicyResult is the deterministic projection and outbox intent for one page.
type MailboxSyncPolicyResult struct {
	Advance    MailboxSyncPolicyAdvance
	Folders    []MailboxSyncPolicyFolderIntent
	Upserts    []MailboxSyncPolicyMessageIntent
	Tombstones []MailboxSyncPolicyTombstoneIntent
	Counts     MailboxSyncPolicyCountIntent
	Events     []events.EnvelopeV1
}

// MailboxSyncPolicyAdvance is an exact-one cursor/checkpoint advance union.
type MailboxSyncPolicyAdvance struct {
	SyncEpoch   int64
	Change      *MailboxSyncPolicyChangeAdvance
	Enumeration *MailboxSyncPolicyEnumerationAdvance
}

// MailboxSyncPolicyChangeAdvance is a safe opaque change cursor advance.
type MailboxSyncPolicyChangeAdvance struct {
	FromCursor string
	NextCursor string
}

// MailboxSyncPolicyEnumerationAdvance is a safe resumable enumeration advance.
type MailboxSyncPolicyEnumerationAdvance struct {
	BaseCursor     string
	Checkpoint     string
	NextCheckpoint string
	Complete       bool
}

// MailboxSyncPolicyFolderIntent carries a validated engine folder snapshot.
type MailboxSyncPolicyFolderIntent struct {
	Folder   MailboxEventID
	Snapshot EngineFolder
}

// MailboxSyncPolicyMessageIntent carries a validated, content-free message projection.
type MailboxSyncPolicyMessageIntent struct {
	Message         MailboxEventID
	Thread          MailboxEventID
	ThreadIdentity  ThreadIdentity
	Projection      EngineMessageProjection
	FolderIDs       []MailboxEventID
	Labels          []MailboxSyncPolicyLabelState
	ReceivedEventID *uuid.UUID
	CalendarMethod  *MailCalendarMethod
	AuthenticationClass      MailAuthenticationClass
	CalendarPartLocatorHash  []byte
	ChangedFields   []MailChangedField
}

// MailboxSyncPolicyTombstoneIntent carries deletion state without live folders or flags.
type MailboxSyncPolicyTombstoneIntent struct {
	EngineMessageID string
	Message         MailboxEventID
	Labels          []MailboxSyncPolicyLabelState
}

// MailboxSyncPolicyCountIntent identifies projections needing count recomputation.
type MailboxSyncPolicyCountIntent struct {
	Folders []MailboxEventID
	Threads []MailboxEventID
}

// MailboxSyncPolicyErrorKind classifies a content-free sync policy failure.
type MailboxSyncPolicyErrorKind string

const (
	MailboxSyncPolicyErrorInvalid         MailboxSyncPolicyErrorKind = "invalid"
	MailboxSyncPolicyErrorTooLarge        MailboxSyncPolicyErrorKind = "too_large"
	MailboxSyncPolicyErrorDuplicate       MailboxSyncPolicyErrorKind = "duplicate"
	MailboxSyncPolicyErrorAccountMismatch MailboxSyncPolicyErrorKind = "account_mismatch"
)

var (
	ErrMailboxSyncPolicyInvalid         = errors.New("mailbox sync policy invalid")
	ErrMailboxSyncPolicyTooLarge        = errors.New("mailbox sync policy exceeds bounds")
	ErrMailboxSyncPolicyDuplicate       = errors.New("mailbox sync policy duplicate")
	ErrMailboxSyncPolicyAccountMismatch = errors.New("mailbox sync policy account mismatch")
)

// MailboxSyncPolicyError is a stable typed sync policy error with no input detail.
type MailboxSyncPolicyError struct {
	Kind MailboxSyncPolicyErrorKind
}

func (errorValue *MailboxSyncPolicyError) Error() string {
	if errorValue == nil {
		return "mailbox sync policy error"
	}
	switch errorValue.Kind {
	case MailboxSyncPolicyErrorInvalid:
		return ErrMailboxSyncPolicyInvalid.Error()
	case MailboxSyncPolicyErrorTooLarge:
		return ErrMailboxSyncPolicyTooLarge.Error()
	case MailboxSyncPolicyErrorDuplicate:
		return ErrMailboxSyncPolicyDuplicate.Error()
	case MailboxSyncPolicyErrorAccountMismatch:
		return ErrMailboxSyncPolicyAccountMismatch.Error()
	default:
		return "mailbox sync policy error"
	}
}

func (errorValue *MailboxSyncPolicyError) Unwrap() error {
	if errorValue == nil {
		return nil
	}
	switch errorValue.Kind {
	case MailboxSyncPolicyErrorInvalid:
		return ErrMailboxSyncPolicyInvalid
	case MailboxSyncPolicyErrorTooLarge:
		return ErrMailboxSyncPolicyTooLarge
	case MailboxSyncPolicyErrorDuplicate:
		return ErrMailboxSyncPolicyDuplicate
	case MailboxSyncPolicyErrorAccountMismatch:
		return ErrMailboxSyncPolicyAccountMismatch
	default:
		return nil
	}
}

// ClassifyMailboxSyncPage validates and deterministically classifies one caller-resolved engine page.
func ClassifyMailboxSyncPage(input MailboxSyncPolicyInput) (MailboxSyncPolicyResult, error) {
	accountID, err := syncValidateEventContext(input.Event)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	if input.SyncEpoch < 0 || !syncValidTime(input.EventCutoff) {
		return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}

	page, err := syncValidatePage(input.Page)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	folderByEngineID, folderIntents, err := syncIndexFolders(accountID, page.accountID, input.Folders)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	labelsByKeyword, err := syncIndexLabels(accountID, input.Labels)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}

	messageByEngineID, err := syncIndexMessages(accountID, input.Messages)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	destroyedByEngineID, err := syncIndexDestroyed(accountID, input.Destroyed)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	for _, message := range messageByEngineID {
		for _, destroyed := range destroyedByEngineID {
			if !destroyed.Projected {
				continue
			}
			if message.Message.ID == destroyed.Message.ID {
				return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
			}
		}
	}
	if err := syncRequireExactPageState(page, messageByEngineID, destroyedByEngineID); err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	if err := syncValidatePageEventIDUniqueness(input.Messages, input.Destroyed); err != nil {
		return MailboxSyncPolicyResult{}, err
	}

	result := MailboxSyncPolicyResult{
		Folders:    folderIntents,
		Upserts:    make([]MailboxSyncPolicyMessageIntent, 0, len(messageByEngineID)),
		Tombstones: make([]MailboxSyncPolicyTombstoneIntent, 0, len(destroyedByEngineID)),
		Events:     make([]events.EnvelopeV1, 0, len(messageByEngineID)+len(destroyedByEngineID)),
	}
	if page.change != nil {
		result.Advance = MailboxSyncPolicyAdvance{SyncEpoch: input.SyncEpoch, Change: &MailboxSyncPolicyChangeAdvance{FromCursor: page.change.FromCursor, NextCursor: page.change.NextCursor}}
	} else {
		result.Advance = MailboxSyncPolicyAdvance{SyncEpoch: input.SyncEpoch, Enumeration: &MailboxSyncPolicyEnumerationAdvance{BaseCursor: page.enumeration.BaseCursor, Checkpoint: page.enumeration.Checkpoint, NextCheckpoint: page.enumeration.NextCheckpoint, Complete: page.enumeration.Complete}}
	}

	folderCountIDs := make(map[uuid.UUID]MailboxEventID, len(folderIntents))
	threadCountIDs := make(map[uuid.UUID]MailboxEventID, len(messageByEngineID)+len(destroyedByEngineID))
	for _, folder := range folderIntents {
		folderCountIDs[folder.Folder.ID] = folder.Folder
	}
	eventRecords := make([]syncEventRecord, 0, len(messageByEngineID)+len(destroyedByEngineID))
	for _, projection := range page.projections {
		message := messageByEngineID[projection.MessageID]
		intent, eventRecord, currentFolders, previousFolders, err := syncClassifyMessage(input, accountID, page.accountID, projection, message, folderByEngineID, labelsByKeyword)
		if err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		result.Upserts = append(result.Upserts, intent)
		for _, folder := range currentFolders {
			folderCountIDs[folder.ID] = folder
		}
		for _, folder := range previousFolders {
			folderCountIDs[folder.ID] = folder
		}
		threadCountIDs[intent.Thread.ID] = intent.Thread
		if message.Previous != nil {
			threadCountIDs[message.Previous.Thread.ID] = message.Previous.Thread
		}
		if eventRecord.present {
			eventRecords = append(eventRecords, eventRecord)
		}
	}
	for _, engineMessageID := range page.destroyedIDs {
		destroyed := destroyedByEngineID[engineMessageID]
		if !destroyed.Projected {
			continue
		}
		intent, eventRecord, err := syncClassifyDestroyed(input, accountID, destroyed)
		if err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		result.Tombstones = append(result.Tombstones, intent)
		threadCountIDs[destroyed.Thread.ID] = destroyed.Thread
		if eventRecord.present {
			eventRecords = append(eventRecords, eventRecord)
		}
	}

	sort.Slice(result.Folders, func(left, right int) bool { return syncLessUUID(result.Folders[left].Folder.ID, result.Folders[right].Folder.ID) })
	sort.Slice(result.Upserts, func(left, right int) bool { return result.Upserts[left].Projection.MessageID < result.Upserts[right].Projection.MessageID })
	sort.Slice(result.Tombstones, func(left, right int) bool { return result.Tombstones[left].EngineMessageID < result.Tombstones[right].EngineMessageID })
	sort.Slice(eventRecords, func(left, right int) bool { return eventRecords[left].engineMessageID < eventRecords[right].engineMessageID })
	result.Counts = MailboxSyncPolicyCountIntent{Folders: syncSortedIDs(folderCountIDs), Threads: syncSortedIDs(threadCountIDs)}
	result.Events = make([]events.EnvelopeV1, len(eventRecords))
	for index, record := range eventRecords {
		result.Events[index] = record.envelope
	}
	return result, nil
}

type syncValidatedPage struct {
	accountID    string
	change       *EngineChangePage
	enumeration  *EngineEnumerationPage
	projections  []EngineMessageProjection
	destroyedIDs []string
}

type syncEventRecord struct {
	engineMessageID string
	envelope        events.EnvelopeV1
	present         bool
}

func syncValidateEventContext(value MailboxSyncPolicyEventContext) (uuid.UUID, error) {
	accountID, err := syncValidateMailboxID(value.Mailbox, uuid.Nil)
	if err != nil {
		return uuid.Nil, err
	}
	if value.OrganizationID == uuid.Nil || (value.ActorID != nil && *value.ActorID == uuid.Nil) || !syncValidEventOccurredAt(value.OccurredAt) || !syncValidCorrelationID(value.CorrelationID) {
		return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return accountID, nil
}

func syncValidatePage(value MailboxSyncPolicyPage) (syncValidatedPage, error) {
	if (value.Change == nil) == (value.Enumeration == nil) {
		return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if value.Change != nil {
		page := value.Change
		if page.CreatedOrUpdated == nil || page.DestroyedIDs == nil || !syncValidOpaque(page.AccountID, MaxEngineOpaqueIDBytes) || !syncValidCursor(page.FromCursor, false) || !syncValidCursor(page.NextCursor, true) {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if page.Bounds.Requested > MaxEnginePageItems || page.Bounds.Returned > MaxEnginePageItems || len(page.CreatedOrUpdated)+len(page.DestroyedIDs) > MaxEnginePageItems {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
		}
		if err := syncPreflightProjectionCollections(page.CreatedOrUpdated); err != nil {
			return syncValidatedPage{}, err
		}
		if syncPageBytes(page.CreatedOrUpdated, page.AccountID, page.FromCursor, page.NextCursor, page.DestroyedIDs) > MaxEngineProjectionPageBytes {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
		}
		if !syncValidBounds(page.Bounds, len(page.CreatedOrUpdated)+len(page.DestroyedIDs)) {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		return syncValidatePageIDs(syncValidatedPage{accountID: page.AccountID, change: page, projections: page.CreatedOrUpdated, destroyedIDs: page.DestroyedIDs})
	}

	page := value.Enumeration
	if page.Messages == nil || !syncValidOpaque(page.AccountID, MaxEngineOpaqueIDBytes) || !syncValidCursor(page.BaseCursor, false) || !syncValidCursor(page.Checkpoint, true) || !syncValidCursor(page.NextCheckpoint, true) || (!page.Complete && page.NextCheckpoint == "") || (page.Complete && page.NextCheckpoint != "") {
		return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if page.Bounds.Requested > MaxEnginePageItems || page.Bounds.Returned > MaxEnginePageItems || len(page.Messages) > MaxEnginePageItems {
		return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	if err := syncPreflightProjectionCollections(page.Messages); err != nil {
		return syncValidatedPage{}, err
	}
	if syncPageBytes(page.Messages, page.AccountID, page.BaseCursor, page.Checkpoint, []string{page.NextCheckpoint}) > MaxEngineProjectionPageBytes {
		return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	if !syncValidBounds(page.Bounds, len(page.Messages)) {
		return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return syncValidatePageIDs(syncValidatedPage{accountID: page.AccountID, enumeration: page, projections: page.Messages, destroyedIDs: []string{}})
}

func syncPreflightProjectionCollections(values []EngineMessageProjection) error {
	for _, value := range values {
		if len(value.To) > MaxEngineAddressItems || len(value.CC) > MaxEngineAddressItems || len(value.BCC) > MaxEngineAddressItems || len(value.ReplyTo) > MaxEngineAddressItems || len(value.References) > MaxEngineReferenceItems || len(value.FolderIDs) > MaxEngineFolderItems || len(value.ReservedKeywords) > MaxEngineKeywordItems || len(value.Attachments) > MaxEngineAttachmentItems {
			return syncPolicyError(MailboxSyncPolicyErrorTooLarge)
		}
	}
	return nil
}

func syncValidatePageIDs(page syncValidatedPage) (syncValidatedPage, error) {
	seen := make(map[string]struct{}, len(page.projections)+len(page.destroyedIDs))
	for _, projection := range page.projections {
		if !syncValidOpaque(projection.MessageID, MaxEngineOpaqueIDBytes) {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seen[projection.MessageID]; duplicate {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[projection.MessageID] = struct{}{}
	}
	for _, id := range page.destroyedIDs {
		if !syncValidOpaque(id, MaxEngineOpaqueIDBytes) {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seen[id]; duplicate {
			return syncValidatedPage{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[id] = struct{}{}
	}
	return page, nil
}

func syncIndexFolders(accountID uuid.UUID, engineAccountID string, values []MailboxSyncPolicyFolder) (map[string]MailboxSyncPolicyFolder, []MailboxSyncPolicyFolderIntent, error) {
	if len(values) > MaxEngineFolderItems {
		return nil, nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	indexed := make(map[string]MailboxSyncPolicyFolder, len(values))
	intents := make([]MailboxSyncPolicyFolderIntent, 0, len(values))
	seenHub := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if _, err := syncValidateMailboxID(value.Folder, accountID); err != nil {
			return nil, nil, err
		}
		if value.Snapshot.AccountID != engineAccountID || !syncValidOpaque(value.Snapshot.FolderID, MaxEngineOpaqueIDBytes) || !syncValidOpaqueOrEmpty(value.Snapshot.ParentID, MaxEngineOpaqueIDBytes) || !syncValidSafeText(value.Snapshot.Name, MaxEngineHeaderValueBytes, false) || !syncValidFolderRole(value.Snapshot.Role) {
			return nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := indexed[value.Snapshot.FolderID]; duplicate {
			return nil, nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		if _, duplicate := seenHub[value.Folder.ID]; duplicate {
			return nil, nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seenHub[value.Folder.ID] = struct{}{}
		indexed[value.Snapshot.FolderID] = value
		intents = append(intents, MailboxSyncPolicyFolderIntent{Folder: value.Folder, Snapshot: value.Snapshot})
	}
	return indexed, intents, nil
}

func syncIndexLabels(accountID uuid.UUID, values []MailboxSyncPolicyLabelDefinition) (map[EngineKeywordMarker]MailboxSyncPolicyLabelDefinition, error) {
	if len(values) > MaxEnginePageItems {
		return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	indexed := make(map[EngineKeywordMarker]MailboxSyncPolicyLabelDefinition, len(values))
	seenLabels := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if _, err := syncValidateMailboxID(value.Label, accountID); err != nil {
			return nil, err
		}
		if !syncValidOpaque(string(value.EngineKeyword), MaxEngineOpaqueIDBytes) {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := indexed[value.EngineKeyword]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		if _, duplicate := seenLabels[value.Label.ID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seenLabels[value.Label.ID] = struct{}{}
		indexed[value.EngineKeyword] = value
	}
	return indexed, nil
}

func syncIndexMessages(accountID uuid.UUID, values []MailboxSyncPolicyMessage) (map[string]MailboxSyncPolicyMessage, error) {
	if len(values) > MaxEnginePageItems {
		return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	indexed := make(map[string]MailboxSyncPolicyMessage, len(values))
	seenMessages := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if !syncValidOpaque(value.EngineMessageID, MaxEngineOpaqueIDBytes) {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, err := syncValidateMailboxID(value.Message, accountID); err != nil {
			return nil, err
		}
		if _, duplicate := indexed[value.EngineMessageID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		if _, duplicate := seenMessages[value.Message.ID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seenMessages[value.Message.ID] = struct{}{}
		indexed[value.EngineMessageID] = value
	}
	return indexed, nil
}

func syncIndexDestroyed(accountID uuid.UUID, values []MailboxSyncPolicyDestroyed) (map[string]MailboxSyncPolicyDestroyed, error) {
	if len(values) > MaxEnginePageItems {
		return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	indexed := make(map[string]MailboxSyncPolicyDestroyed, len(values))
	seenMessages := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if !syncValidOpaque(value.EngineMessageID, MaxEngineOpaqueIDBytes) {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := indexed[value.EngineMessageID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		if !value.Projected {
			if value.Message != (MailboxEventID{}) || value.Thread != (MailboxEventID{}) || len(value.Labels) != 0 || value.ChangedEvent != (MailboxSyncPolicyEventReservation{}) {
				return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
			}
			indexed[value.EngineMessageID] = value
			continue
		}
		for _, id := range []MailboxEventID{value.Message, value.Thread} {
			if _, err := syncValidateMailboxID(id, accountID); err != nil {
				return nil, err
			}
		}
		if _, duplicate := seenMessages[value.Message.ID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seenMessages[value.Message.ID] = struct{}{}
		indexed[value.EngineMessageID] = value
	}
	return indexed, nil
}

func syncRequireExactPageState(page syncValidatedPage, messages map[string]MailboxSyncPolicyMessage, destroyed map[string]MailboxSyncPolicyDestroyed) error {
	if len(messages) != len(page.projections) || len(destroyed) != len(page.destroyedIDs) {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	for _, projection := range page.projections {
		if _, found := messages[projection.MessageID]; !found {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	}
	for _, engineMessageID := range page.destroyedIDs {
		if _, found := destroyed[engineMessageID]; !found {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	}
	return nil
}

func syncValidatePageEventIDUniqueness(messages []MailboxSyncPolicyMessage, destroyed []MailboxSyncPolicyDestroyed) error {
	seen := make(map[uuid.UUID]struct{}, 3*len(messages)+len(destroyed))
	validate := func(eventID uuid.UUID) error {
		if eventID == uuid.Nil {
			return nil
		}
		if _, duplicate := seen[eventID]; duplicate {
			return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[eventID] = struct{}{}
		return nil
	}
	for _, message := range messages {
		if message.Previous == nil || message.Previous.ReceivedEventID == nil {
			continue
		}
		if *message.Previous.ReceivedEventID == uuid.Nil {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if err := validate(*message.Previous.ReceivedEventID); err != nil {
			return err
		}
	}
	for _, message := range messages {
		if err := validate(message.ReceivedEventID); err != nil {
			return err
		}
		if err := validate(message.ChangedEvent.ID); err != nil {
			return err
		}
	}
	for _, value := range destroyed {
		if err := validate(value.ChangedEvent.ID); err != nil {
			return err
		}
	}
	return nil
}

func syncClassifyMessage(input MailboxSyncPolicyInput, accountID uuid.UUID, engineAccountID string, projection EngineMessageProjection, message MailboxSyncPolicyMessage, folders map[string]MailboxSyncPolicyFolder, labels map[EngineKeywordMarker]MailboxSyncPolicyLabelDefinition) (MailboxSyncPolicyMessageIntent, syncEventRecord, []MailboxEventID, []MailboxEventID, error) {
	if projection.MessageID != message.EngineMessageID || projection.AccountID != engineAccountID {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if err := syncValidateProjection(projection); err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}
	currentFolders, err := syncResolveFolders(accountID, projection.FolderIDs, folders)
	if err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}
	identity, err := syncResolveThread(accountID, projection, message.Thread)
	if err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}
	if err := syncValidateProvenance(accountID, projection, message.Provenance); err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}
	previousFolders, previousLabels, previousReceived, err := syncValidatePrevious(accountID, message.Previous)
	if err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}
	desiredLabels, err := syncDesiredLabels(accountID, projection.ReservedKeywords, previousLabels, labels)
	if err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}

	firstSeen := message.Previous == nil
	inbound := projection.Origin == EngineProjectionOriginInboundOrUnknown
	if firstSeen && inbound {
		if projection.ReceivedAt == nil {
			return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	}
	emitReceived := firstSeen && inbound && projection.ReceivedAt.After(input.EventCutoff)
	if emitReceived && message.ReceivedEventID == uuid.Nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if !emitReceived && message.ReceivedEventID != uuid.Nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}

	resultReceived := syncCopyUUIDPointer(previousReceived)
	if emitReceived {
		resultReceived = syncUUIDPointer(message.ReceivedEventID)
	}
	changedFields := syncChangedFields(currentFolders, previousFolders, projection.Flags, message.Previous, desiredLabels, previousLabels)
	intent := MailboxSyncPolicyMessageIntent{
		Message: message.Message, Thread: message.Thread.Thread, ThreadIdentity: identity,
		Projection: syncCloneSafeProjection(projection), FolderIDs: syncCopyMailboxIDs(currentFolders),
		Labels: syncCopyLabelStates(desiredLabels), ReceivedEventID: resultReceived,
		CalendarMethod: copyMailCalendarMethod(projection.CalendarMethod), AuthenticationClass: projection.AuthenticationClass,
		CalendarPartLocatorHash: syncCloneBytes(projection.CalendarPartLocatorHash), ChangedFields: syncCopyChangedFields(changedFields),
	}

	if emitReceived {
		if message.ChangedEvent.ID != uuid.Nil || message.ChangedEvent.Committed {
			return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		event, err := syncNewReceivedEvent(input.Event, message.ReceivedEventID, intent)
		if err != nil {
			return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
		}
		return intent, syncEventRecord{engineMessageID: projection.MessageID, envelope: event, present: true}, currentFolders, previousFolders, nil
	}

	if len(changedFields) == 0 {
		if message.ChangedEvent.ID != uuid.Nil || message.ChangedEvent.Committed {
			return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		return intent, syncEventRecord{}, currentFolders, previousFolders, nil
	}
	if err := syncValidateReservation(message.ChangedEvent, true); err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}
	if message.ChangedEvent.Committed {
		return intent, syncEventRecord{}, currentFolders, previousFolders, nil
	}
	event, err := syncNewChangedEvent(input.Event, message.ChangedEvent.ID, message.Message, changedFields, currentFolders, projection.Flags, desiredLabels)
	if err != nil {
		return MailboxSyncPolicyMessageIntent{}, syncEventRecord{}, nil, nil, err
	}
	return intent, syncEventRecord{engineMessageID: projection.MessageID, envelope: event, present: true}, currentFolders, previousFolders, nil
}

func syncClassifyDestroyed(input MailboxSyncPolicyInput, accountID uuid.UUID, destroyed MailboxSyncPolicyDestroyed) (MailboxSyncPolicyTombstoneIntent, syncEventRecord, error) {
	if !destroyed.Projected {
		return MailboxSyncPolicyTombstoneIntent{}, syncEventRecord{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	labels, err := syncValidateLabelStates(accountID, destroyed.Labels)
	if err != nil {
		return MailboxSyncPolicyTombstoneIntent{}, syncEventRecord{}, err
	}
	if err := syncValidateReservation(destroyed.ChangedEvent, true); err != nil {
		return MailboxSyncPolicyTombstoneIntent{}, syncEventRecord{}, err
	}
	intent := MailboxSyncPolicyTombstoneIntent{EngineMessageID: destroyed.EngineMessageID, Message: destroyed.Message, Labels: syncCopyLabelStates(labels)}
	if destroyed.ChangedEvent.Committed {
		return intent, syncEventRecord{}, nil
	}
	event, err := syncNewChangedEvent(input.Event, destroyed.ChangedEvent.ID, destroyed.Message, []MailChangedField{MailChangedFieldDeleted}, []MailboxEventID{}, EngineMessageFlags{}, labels)
	if err != nil {
		return MailboxSyncPolicyTombstoneIntent{}, syncEventRecord{}, err
	}
	return intent, syncEventRecord{engineMessageID: destroyed.EngineMessageID, envelope: event, present: true}, nil
}

func syncResolveThread(accountID uuid.UUID, projection EngineMessageProjection, value MailboxSyncPolicyThreadResolution) (ThreadIdentity, error) {
	if _, err := syncValidateMailboxID(value.Thread, accountID); err != nil {
		return ThreadIdentity{}, err
	}
	if len(value.Ancestors) > MaxEngineReferenceItems+MaxEngineAddressItems+2 {
		return ThreadIdentity{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	identity, err := DeriveThreadIdentity(ThreadingInput{References: projection.References, InReplyTo: projection.InReplyTo, Subject: projection.Subject, From: projection.From, To: projection.To, CC: projection.CC, Ancestors: value.Ancestors, AncestorLookupComplete: value.AncestorLookupComplete})
	if err != nil {
		return ThreadIdentity{}, syncMapThreadError(err)
	}
	if identity.KeyVersion != value.KeyVersion || identity.Key != value.Key {
		return ThreadIdentity{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return identity, nil
}

func syncValidateProvenance(accountID uuid.UUID, projection EngineMessageProjection, value MailboxSyncPolicyProvenance) error {
	if value.ConfirmedSubmission != nil && value.ConfirmedDraft != nil {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	proof := value.ConfirmedSubmission
	wantOrigin := EngineProjectionOriginConfirmedOwnerSubmission
	if value.ConfirmedDraft != nil {
		proof = value.ConfirmedDraft
		wantOrigin = EngineProjectionOriginConfirmedOwnerDraft
	}
	if projection.Origin == EngineProjectionOriginInboundOrUnknown {
		if proof != nil || len(projection.BCC) != 0 {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		return nil
	}
	if proof == nil || projection.Origin != wantOrigin {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if _, err := syncValidateMailboxID(proof.Relation, accountID); err != nil {
		return err
	}
	if proof.EngineMessageID != projection.MessageID || !syncValidOpaque(proof.EngineMessageID, MaxEngineOpaqueIDBytes) {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return nil
}

func syncValidatePrevious(accountID uuid.UUID, previous *MailboxSyncPolicyPreviousMessage) ([]MailboxEventID, []MailboxSyncPolicyLabelState, *uuid.UUID, error) {
	if previous == nil {
		return []MailboxEventID{}, []MailboxSyncPolicyLabelState{}, nil, nil
	}
	if previous.ReceivedEventID != nil && *previous.ReceivedEventID == uuid.Nil {
		return nil, nil, nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if _, err := syncValidateMailboxID(previous.Thread, accountID); err != nil {
		return nil, nil, nil, err
	}
	if len(previous.FolderIDs) > MaxEngineFolderItems {
		return nil, nil, nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	folders, err := syncValidateMailboxIDSet(accountID, previous.FolderIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	labels, err := syncValidateLabelStates(accountID, previous.Labels)
	if err != nil {
		return nil, nil, nil, err
	}
	return folders, labels, syncCopyUUIDPointer(previous.ReceivedEventID), nil
}

func syncDesiredLabels(accountID uuid.UUID, markers []EngineKeywordMarker, previous []MailboxSyncPolicyLabelState, definitions map[EngineKeywordMarker]MailboxSyncPolicyLabelDefinition) ([]MailboxSyncPolicyLabelState, error) {
	if markers == nil {
		return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if len(markers) > MaxEngineKeywordItems {
		return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	states := make(map[uuid.UUID]MailboxSyncPolicyLabelState, len(previous)+len(markers))
	for _, previousState := range previous {
		if previousState.Manual {
			states[previousState.Label.ID] = MailboxSyncPolicyLabelState{Label: previousState.Label, Manual: true}
		}
	}
	seenMarkers := make(map[EngineKeywordMarker]struct{}, len(markers))
	for _, marker := range markers {
		if !syncValidOpaque(string(marker), MaxEngineOpaqueIDBytes) {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seenMarkers[marker]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seenMarkers[marker] = struct{}{}
		definition, found := definitions[marker]
		if !found {
			continue
		}
		state := states[definition.Label.ID]
		state.Label = definition.Label
		state.EngineRule = true
		states[definition.Label.ID] = state
	}
	if len(states) > MaxEnginePageItems {
		return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	result := make([]MailboxSyncPolicyLabelState, 0, len(states))
	for _, state := range states {
		if _, err := syncValidateMailboxID(state.Label, accountID); err != nil {
			return nil, err
		}
		result = append(result, state)
	}
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left].Label.ID, result[right].Label.ID) })
	return result, nil
}

func syncValidateLabelStates(accountID uuid.UUID, values []MailboxSyncPolicyLabelState) ([]MailboxSyncPolicyLabelState, error) {
	if len(values) > MaxEnginePageItems {
		return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	result := make([]MailboxSyncPolicyLabelState, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for index, value := range values {
		if _, err := syncValidateMailboxID(value.Label, accountID); err != nil {
			return nil, err
		}
		if !value.Manual && !value.EngineRule {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seen[value.Label.ID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[value.Label.ID] = struct{}{}
		result[index] = value
	}
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left].Label.ID, result[right].Label.ID) })
	return result, nil
}

func syncChangedFields(currentFolders, previousFolders []MailboxEventID, currentFlags EngineMessageFlags, previous *MailboxSyncPolicyPreviousMessage, currentLabels, previousLabels []MailboxSyncPolicyLabelState) []MailChangedField {
	if previous == nil {
		return []MailChangedField{}
	}
	result := make([]MailChangedField, 0, 6)
	if !syncEqualMailboxIDSets(currentFolders, previousFolders) {
		result = append(result, MailChangedFieldFolderIDs)
	}
	if currentFlags.Seen != previous.Flags.Seen {
		result = append(result, MailChangedFieldSeen)
	}
	if currentFlags.Answered != previous.Flags.Answered {
		result = append(result, MailChangedFieldAnswered)
	}
	if currentFlags.Flagged != previous.Flags.Flagged {
		result = append(result, MailChangedFieldFlagged)
	}
	if currentFlags.Junk != previous.Flags.Junk {
		result = append(result, MailChangedFieldJunk)
	}
	if !syncEqualLabelStates(currentLabels, previousLabels) {
		result = append(result, MailChangedFieldLabels)
	}
	return result
}

func syncNewReceivedEvent(context MailboxSyncPolicyEventContext, eventID uuid.UUID, intent MailboxSyncPolicyMessageIntent) (events.EnvelopeV1, error) {
	projection := intent.Projection
	from, err := syncEventAddress(projection.From)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	to, err := syncEventAddresses(projection.To)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	event, err := NewMailReceivedEnvelopeV1(eventID, context.OccurredAt, context.OrganizationID, syncCopyUUIDPointer(context.ActorID), context.CorrelationID, MailReceivedEventPayload{Mailbox: context.Mailbox, Message: intent.Message, EngineMessage: MailboxEventOpaqueID{AccountID: context.Mailbox.AccountID, ID: projection.MessageID}, Thread: intent.Thread, From: from, To: to, ReceivedAt: *projection.ReceivedAt, FolderIDs: syncCopyMailboxIDs(intent.FolderIDs), HasAttachments: len(projection.Attachments) != 0, CalendarMethod: copyMailCalendarMethod(intent.CalendarMethod), AuthenticationClass: intent.AuthenticationClass, CalendarPartLocatorHash: syncCloneBytes(intent.CalendarPartLocatorHash)})
	if err != nil {
		return events.EnvelopeV1{}, syncMapMailEventError(err)
	}
	return event, nil
}

func syncNewChangedEvent(context MailboxSyncPolicyEventContext, eventID uuid.UUID, message MailboxEventID, fields []MailChangedField, folders []MailboxEventID, flags EngineMessageFlags, labels []MailboxSyncPolicyLabelState) (events.EnvelopeV1, error) {
	labelIDs := make([]MailboxEventID, len(labels))
	for index, label := range labels {
		labelIDs[index] = label.Label
	}
	event, err := NewMailMessageChangedEnvelopeV1(eventID, context.OccurredAt, context.OrganizationID, syncCopyUUIDPointer(context.ActorID), context.CorrelationID, MailMessageChangedEventPayload{Mailbox: context.Mailbox, Message: message, ChangedFields: append([]MailChangedField(nil), fields...), FolderIDs: syncCopyMailboxIDs(folders), Flags: syncEventFlags(flags), LabelIDs: labelIDs})
	if err != nil {
		return events.EnvelopeV1{}, syncMapMailEventError(err)
	}
	return event, nil
}

func syncEventAddress(value EngineAddress) (MailEventAddress, error) {
	normalized, err := NormalizeAddress(value)
	if err != nil || normalized != value {
		return MailEventAddress{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	result := MailEventAddress{Email: normalized.Email}
	if normalized.Name != "" {
		result.Name = &normalized.Name
	}
	return result, nil
}

func syncEventAddresses(values []EngineAddress) ([]MailEventAddress, error) {
	result := make([]MailEventAddress, len(values))
	for index, value := range values {
		address, err := syncEventAddress(value)
		if err != nil {
			return nil, err
		}
		result[index] = address
	}
	return result, nil
}

func syncResolveFolders(accountID uuid.UUID, values []string, folders map[string]MailboxSyncPolicyFolder) ([]MailboxEventID, error) {
	if values == nil || len(values) == 0 {
		return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if len(values) > MaxEngineFolderItems {
		return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	result := make([]MailboxEventID, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		if !syncValidOpaque(value, MaxEngineOpaqueIDBytes) {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seen[value]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[value] = struct{}{}
		folder, found := folders[value]
		if !found {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, err := syncValidateMailboxID(folder.Folder, accountID); err != nil {
			return nil, err
		}
		result[index] = folder.Folder
	}
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left].ID, result[right].ID) })
	return result, nil
}

func syncValidateProjection(value EngineMessageProjection) error {
	if value.References == nil || value.To == nil || value.CC == nil || value.BCC == nil || value.ReplyTo == nil || value.FolderIDs == nil || value.ReservedKeywords == nil || value.Attachments == nil {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if !syncValidOpaque(value.AccountID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(value.MessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(value.Version, MaxEngineOpaqueIDBytes) || (value.InternetMessageID != "" && !validThreadMessageID(value.InternetMessageID)) || (value.InReplyTo != "" && !validThreadMessageID(value.InReplyTo)) || !syncValidSafeText(value.Subject, MaxEngineHeaderValueBytes, true) || !syncValidSafeText(value.Snippet, MaxEngineSnippetBytes, true) || !validMailCalendarMethod(value.CalendarMethod) || value.SizeBytes < 0 || value.SizeBytes > MaxEngineMessageBytes || !syncValidPriority(value.Priority) || !syncValidOrigin(value.Origin) {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if value.AuthenticationClass != MailAuthenticationClassTrustedInternal && value.AuthenticationClass != MailAuthenticationClassDMARCAligned && value.AuthenticationClass != MailAuthenticationClassUntrusted {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if (value.CalendarMethod == nil) != (value.CalendarPartLocatorHash == nil) || (value.CalendarPartLocatorHash != nil && len(value.CalendarPartLocatorHash) != sha256.Size) {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if value.ReceivedAt != nil && !syncValidTime(*value.ReceivedAt) {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if value.SentAt != nil && !syncValidTime(*value.SentAt) {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if err := syncValidateProjectionAddresses(value); err != nil {
		return err
	}
	if err := syncValidateProjectionCollections(value); err != nil {
		return err
	}
	if syncProjectionBytes(value) > MaxEngineProjectionItemBytes {
		return syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	return nil
}

func syncValidateProjectionAddresses(value EngineMessageProjection) error {
	if len(value.To) > MaxEngineAddressItems || len(value.CC) > MaxEngineAddressItems || len(value.BCC) > MaxEngineAddressItems || len(value.ReplyTo) > MaxEngineAddressItems {
		return syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	if err := syncValidateEngineAddress(value.From, true); err != nil {
		return err
	}
	all := make(map[string]struct{}, len(value.To)+len(value.CC)+len(value.BCC))
	for collectionIndex, collection := range [][]EngineAddress{value.To, value.CC, value.BCC, value.ReplyTo} {
		seenCollection := make(map[string]struct{}, len(collection))
		for _, address := range collection {
			if err := syncValidateEngineAddress(address, true); err != nil {
				return err
			}
			if _, duplicate := seenCollection[address.Email]; duplicate {
				return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
			}
			seenCollection[address.Email] = struct{}{}
			if collectionIndex != 3 {
				if _, duplicate := all[address.Email]; duplicate {
					return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
				}
				all[address.Email] = struct{}{}
			}
		}
	}
	return nil
}

func syncValidateProjectionCollections(value EngineMessageProjection) error {
	if len(value.FolderIDs) == 0 {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if len(value.References) > MaxEngineReferenceItems || len(value.FolderIDs) > MaxEngineFolderItems || len(value.ReservedKeywords) > MaxEngineKeywordItems || len(value.Attachments) > MaxEngineAttachmentItems {
		return syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	if err := syncValidateThreadMessageIDs(value.References); err != nil {
		return err
	}
	if err := syncValidateOpaqueStrings(value.FolderIDs); err != nil {
		return err
	}
	seenKeywords := make(map[EngineKeywordMarker]struct{}, len(value.ReservedKeywords))
	for _, keyword := range value.ReservedKeywords {
		if !syncValidOpaque(string(keyword), MaxEngineOpaqueIDBytes) {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seenKeywords[keyword]; duplicate {
			return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seenKeywords[keyword] = struct{}{}
	}
	for _, attachment := range value.Attachments {
		if !syncValidOpaque(attachment.BlobID, MaxEngineOpaqueIDBytes) || !syncValidSafeText(attachment.ContentID, MaxEngineMIMEFieldBytes, true) || !syncValidSafeText(attachment.Filename, MaxEngineMIMEFieldBytes, true) || !syncValidSafeText(attachment.MediaType, MaxEngineMIMEFieldBytes, true) || attachment.SizeBytes < 0 || attachment.SizeBytes > MaxEngineAttachmentBytes || (len(attachment.SHA256) != 0 && len(attachment.SHA256) != 32) || !syncValidAttachmentDisposition(attachment.Disposition) || (attachment.Disposition == EngineAttachmentDispositionInline && attachment.ContentID == "") {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	}
	return nil
}

func syncValidateEngineAddress(value EngineAddress, required bool) error {
	if value.Email == "" {
		if !required && syncValidSafeText(value.Name, MaxEngineAddressNameBytes, true) {
			return nil
		}
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	normalized, err := NormalizeAddress(value)
	if err != nil {
		return syncMapAddressError(err)
	}
	if normalized != value {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return nil
}

func syncValidateReservation(value MailboxSyncPolicyEventReservation, required bool) error {
	if value.Committed && value.ID == uuid.Nil {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if required && value.ID == uuid.Nil {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return nil
}

func syncValidateMailboxID(value MailboxEventID, accountID uuid.UUID) (uuid.UUID, error) {
	if value.AccountID == uuid.Nil || value.ID == uuid.Nil {
		return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if accountID != uuid.Nil && value.AccountID != accountID {
		return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorAccountMismatch)
	}
	return value.AccountID, nil
}

func syncValidateMailboxIDSet(accountID uuid.UUID, values []MailboxEventID) ([]MailboxEventID, error) {
	result := make([]MailboxEventID, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for index, value := range values {
		if _, err := syncValidateMailboxID(value, accountID); err != nil {
			return nil, err
		}
		if _, duplicate := seen[value.ID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[value.ID] = struct{}{}
		result[index] = value
	}
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left].ID, result[right].ID) })
	return result, nil
}

func syncValidBounds(value EnginePageBounds, returned int) bool {
	return value.Requested >= 1 && value.Requested <= MaxEnginePageItems && value.Maximum == MaxEnginePageItems && value.Returned == returned && value.Returned <= MaxEnginePageItems
}

func syncValidCursor(value string, allowEmpty bool) bool {
	return (allowEmpty && value == "") || syncValidOpaque(value, MaxEngineCursorBytes)
}

func syncValidOpaque(value string, maximum int) bool {
	return value != "" && syncValidSafeText(value, maximum, false)
}

func syncValidOpaqueOrEmpty(value string, maximum int) bool {
	return value == "" || syncValidOpaque(value, maximum)
}

func syncValidSafeText(value string, maximum int, allowEmpty bool) bool {
	if (!allowEmpty && value == "") || len(value) > maximum || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func syncValidTime(value time.Time) bool {
	if value.IsZero() || value.Location() != time.UTC {
		return false
	}
	_, err := value.MarshalJSON()
	return err == nil
}

func syncValidEventOccurredAt(value time.Time) bool {
	if value.IsZero() {
		return false
	}
	_, offset := value.Zone()
	if offset != 0 {
		return false
	}
	_, err := value.UTC().MarshalJSON()
	return err == nil
}

func syncValidCorrelationID(value string) bool {
	return value != "" && len(value) <= 128 && utf8.ValidString(value) && value == strings.TrimSpace(value)
}

func syncValidFolderRole(value EngineFolderRole) bool {
	switch value {
	case EngineFolderRoleCustom, EngineFolderRoleInbox, EngineFolderRoleSent, EngineFolderRoleDrafts, EngineFolderRoleArchive, EngineFolderRoleTrash, EngineFolderRoleJunk:
		return true
	default:
		return false
	}
}

func syncValidPriority(value EngineMessagePriority) bool {
	return value == EnginePriorityLow || value == EnginePriorityNormal || value == EnginePriorityHigh
}

func syncValidOrigin(value EngineProjectionOrigin) bool {
	return value == EngineProjectionOriginInboundOrUnknown || value == EngineProjectionOriginConfirmedOwnerSubmission || value == EngineProjectionOriginConfirmedOwnerDraft
}

func syncValidAttachmentDisposition(value EngineAttachmentDisposition) bool {
	return value == EngineAttachmentDispositionAttachment || value == EngineAttachmentDispositionInline
}

func syncValidateThreadMessageIDs(values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validThreadMessageID(value) {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seen[value]; duplicate {
			return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func syncValidateOpaqueStrings(values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !syncValidOpaque(value, MaxEngineOpaqueIDBytes) {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := seen[value]; duplicate {
			return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func syncPageBytes(projections []EngineMessageProjection, accountID string, first, second string, destroyed []string) int {
	total := len(accountID) + len(first) + len(second)
	for _, projection := range projections {
		total += syncProjectionBytes(projection)
	}
	for _, id := range destroyed {
		total += len(id)
	}
	return total
}

func syncProjectionBytes(value EngineMessageProjection) int {
	total := len(value.AccountID) + len(value.MessageID) + len(value.Version) + len(value.InternetMessageID) + len(value.InReplyTo) + len(value.Subject) + len(value.Snippet) + len(value.From.Email) + len(value.From.Name) + syncAddressBytes(value.To) + syncAddressBytes(value.CC) + syncAddressBytes(value.BCC) + syncAddressBytes(value.ReplyTo) + syncStringBytes(value.References) + syncStringBytes(value.FolderIDs) + len(value.AuthenticationClass) + len(value.CalendarPartLocatorHash)
	if value.CalendarMethod != nil {
		total += len(*value.CalendarMethod)
	}
	for _, keyword := range value.ReservedKeywords {
		total += len(keyword)
	}
	for _, attachment := range value.Attachments {
		total += len(attachment.BlobID) + len(attachment.ContentID) + len(attachment.Filename) + len(attachment.MediaType) + len(attachment.SHA256)
	}
	return total
}

func syncAddressBytes(values []EngineAddress) int {
	total := 0
	for _, value := range values {
		total += len(value.Email) + len(value.Name)
	}
	return total
}

func syncStringBytes(values []string) int {
	total := 0
	for _, value := range values {
		total += len(value)
	}
	return total
}

func syncCloneSafeProjection(value EngineMessageProjection) EngineMessageProjection {
	result := value
	result.References = syncCloneStrings(value.References)
	result.To = syncCloneAddresses(value.To)
	result.CC = syncCloneAddresses(value.CC)
	result.BCC = []EngineAddress{}
	result.ReplyTo = syncCloneAddresses(value.ReplyTo)
	result.FolderIDs = syncCloneStrings(value.FolderIDs)
	result.ReservedKeywords = []EngineKeywordMarker{}
	result.Attachments = make([]EngineProjectionAttachment, len(value.Attachments))
	for index, attachment := range value.Attachments {
		result.Attachments[index] = attachment
		result.Attachments[index].SHA256 = append([]byte(nil), attachment.SHA256...)
	}
	if value.SentAt != nil {
		timestamp := *value.SentAt
		result.SentAt = &timestamp
	}
	if value.ReceivedAt != nil {
		timestamp := *value.ReceivedAt
		result.ReceivedAt = &timestamp
	}
	result.CalendarMethod = copyMailCalendarMethod(value.CalendarMethod)
	if value.CalendarPartLocatorHash != nil {
		result.CalendarPartLocatorHash = append([]byte{}, value.CalendarPartLocatorHash...)
	}
	return result
}

func syncCloneAddresses(values []EngineAddress) []EngineAddress {
	result := make([]EngineAddress, len(values))
	copy(result, values)
	return result
}

func syncCopyMailboxIDs(values []MailboxEventID) []MailboxEventID {
	result := make([]MailboxEventID, len(values))
	copy(result, values)
	return result
}

func syncCopyLabelStates(values []MailboxSyncPolicyLabelState) []MailboxSyncPolicyLabelState {
	result := make([]MailboxSyncPolicyLabelState, len(values))
	copy(result, values)
	return result
}

func syncCopyChangedFields(values []MailChangedField) []MailChangedField {
	result := make([]MailChangedField, len(values))
	copy(result, values)
	return result
}

func syncCloneStrings(values []string) []string {
	result := make([]string, len(values))
	copy(result, values)
	return result
}

func syncCloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte{}, value...)
}

func syncCopyUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func syncUUIDPointer(value uuid.UUID) *uuid.UUID {
	copyValue := value
	return &copyValue
}

func syncEventFlags(value EngineMessageFlags) []MailMessageFlag {
	result := make([]MailMessageFlag, 0, 5)
	if value.Seen {
		result = append(result, MailMessageFlagSeen)
	}
	if value.Answered {
		result = append(result, MailMessageFlagAnswered)
	}
	if value.Flagged {
		result = append(result, MailMessageFlagFlagged)
	}
	if value.Junk {
		result = append(result, MailMessageFlagJunk)
	}
	if value.Draft {
		result = append(result, MailMessageFlagDraft)
	}
	return result
}

func syncEqualMailboxIDSets(left, right []MailboxEventID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].ID != right[index].ID || left[index].AccountID != right[index].AccountID {
			return false
		}
	}
	return true
}

func syncEqualLabelStates(left, right []MailboxSyncPolicyLabelState) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func syncSortedIDs(values map[uuid.UUID]MailboxEventID) []MailboxEventID {
	result := make([]MailboxEventID, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left].ID, result[right].ID) })
	return result
}

func syncLessUUID(left, right uuid.UUID) bool {
	return bytes.Compare(left[:], right[:]) < 0
}

func syncMapThreadError(err error) error {
	if errors.Is(err, ErrThreadTooLarge) || errors.Is(err, ErrAddressTooLarge) {
		return syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	if errors.Is(err, ErrAddressDuplicate) {
		return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
	}
	return syncPolicyError(MailboxSyncPolicyErrorInvalid)
}

func syncMapAddressError(err error) error {
	if errors.Is(err, ErrAddressTooLarge) {
		return syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	if errors.Is(err, ErrAddressDuplicate) {
		return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
	}
	return syncPolicyError(MailboxSyncPolicyErrorInvalid)
}

func syncMapMailEventError(err error) error {
	switch {
	case errors.Is(err, ErrMailEventTooLarge):
		return syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	case errors.Is(err, ErrMailEventDuplicate):
		return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
	case errors.Is(err, ErrMailEventAccountMismatch):
		return syncPolicyError(MailboxSyncPolicyErrorAccountMismatch)
	default:
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
}

func syncPolicyError(kind MailboxSyncPolicyErrorKind) error {
	return &MailboxSyncPolicyError{Kind: kind}
}
