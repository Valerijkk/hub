package mail

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const mailboxSyncCanonicalApplyVersion int32 = 1
const mailboxSyncCanonicalApplyDomain = "hub.mail.sync.apply.result.v1"

// MailboxSyncCanonicalApply is the pure, canonicalized apply handoff for one sync lease.
type MailboxSyncCanonicalApply struct {
	Fence         MailboxSyncFence
	Result        MailboxSyncPolicyResult
	ExpectedPhase MailboxSyncPhase
	NextAttemptAt time.Time
	ResultHash    [sha256.Size]byte
}

// CanonicalizeMailboxSyncApplyRequest validates, deep-copies, canonicalizes, and hashes one apply request.
func CanonicalizeMailboxSyncApplyRequest(request MailboxSyncApplyRequest) (MailboxSyncCanonicalApply, error) {
	if !syncValidTime(request.Now) || !syncValidTime(request.NextAttemptAt) || !validMailboxSyncFence(request.Fence) {
		return MailboxSyncCanonicalApply{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}

	result, err := canonicalizeMailboxSyncResult(request.Fence, request.Result)
	if err != nil {
		return MailboxSyncCanonicalApply{}, err
	}
	if err := canonicalValidateAdvance(result.Advance); err != nil {
		return MailboxSyncCanonicalApply{}, err
	}
	expectedPhase, ok := mailboxSyncExpectedPhase(request.Fence, result.Advance)
	if !ok {
		return MailboxSyncCanonicalApply{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}

	stream, err := canonicalMailboxSyncApplyStream(request.Fence, request.NextAttemptAt, result)
	if err != nil {
		return MailboxSyncCanonicalApply{}, err
	}
	return MailboxSyncCanonicalApply{
		Fence:         request.Fence,
		Result:        result,
		ExpectedPhase: expectedPhase,
		NextAttemptAt: request.NextAttemptAt,
		ResultHash:    sha256.Sum256(stream),
	}, nil
}

func canonicalizeMailboxSyncResult(fence MailboxSyncFence, value MailboxSyncPolicyResult) (MailboxSyncPolicyResult, error) {
	accountID := fence.AccountID
	if value.Folders == nil || value.Upserts == nil || value.Tombstones == nil || value.Counts.Folders == nil || value.Counts.Threads == nil || value.Events == nil {
		return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if len(value.Folders) > MaxEngineFolderItems || len(value.Upserts) > MaxEnginePageItems || len(value.Tombstones) > MaxEnginePageItems || len(value.Upserts)+len(value.Tombstones) > MaxEnginePageItems || len(value.Events) > MaxEnginePageItems {
		return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}

	result := cloneMailboxSyncPolicyResult(value)
	if result.Folders == nil {
		result.Folders = make([]MailboxSyncPolicyFolderIntent, 0)
	}
	if result.Events == nil {
		result.Events = make([]events.EnvelopeV1, 0)
	}
	for index := range result.Upserts {
		canonicalRestoreProjectionCollections(&result.Upserts[index].Projection, value.Upserts[index].Projection)
		if value.Upserts[index].ThreadIdentity.CanonicalParticipants != nil && result.Upserts[index].ThreadIdentity.CanonicalParticipants == nil {
			result.Upserts[index].ThreadIdentity.CanonicalParticipants = make([]string, 0)
		}
		if value.Upserts[index].FolderIDs != nil && result.Upserts[index].FolderIDs == nil {
			result.Upserts[index].FolderIDs = make([]MailboxEventID, 0)
		}
		if value.Upserts[index].Labels != nil && result.Upserts[index].Labels == nil {
			result.Upserts[index].Labels = make([]MailboxSyncPolicyLabelState, 0)
		}
	}
	for index := range result.Tombstones {
		if value.Tombstones[index].Labels != nil && result.Tombstones[index].Labels == nil {
			result.Tombstones[index].Labels = make([]MailboxSyncPolicyLabelState, 0)
		}
	}
	if result.Counts.Folders == nil {
		result.Counts.Folders = make([]MailboxEventID, 0)
	}
	if result.Counts.Threads == nil {
		result.Counts.Threads = make([]MailboxEventID, 0)
	}
	if err := canonicalValidateAdvance(result.Advance); err != nil {
		return MailboxSyncPolicyResult{}, err
	}

	engineAccountID := ""
	setEngineAccountID := func(candidate string) error {
		if !syncValidOpaque(candidate, MaxEngineOpaqueIDBytes) {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if engineAccountID == "" {
			engineAccountID = candidate
			return nil
		}
		if engineAccountID != candidate {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		return nil
	}

	folderIDs := make(map[uuid.UUID]struct{}, len(result.Folders))
	engineFolderIDs := make(map[string]struct{}, len(result.Folders))
	for index := range result.Folders {
		folder := &result.Folders[index]
		if _, err := syncValidateMailboxID(folder.Folder, accountID); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if err := setEngineAccountID(folder.Snapshot.AccountID); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if !syncValidOpaque(folder.Snapshot.FolderID, MaxEngineOpaqueIDBytes) || !syncValidOpaqueOrEmpty(folder.Snapshot.ParentID, MaxEngineOpaqueIDBytes) || !syncValidSafeText(folder.Snapshot.Name, MaxEngineHeaderValueBytes, false) || !syncValidFolderRole(folder.Snapshot.Role) {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := folderIDs[folder.Folder.ID]; duplicate {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		if _, duplicate := engineFolderIDs[folder.Snapshot.FolderID]; duplicate {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		folderIDs[folder.Folder.ID] = struct{}{}
		engineFolderIDs[folder.Snapshot.FolderID] = struct{}{}
	}
	sort.Slice(result.Folders, func(left, right int) bool {
		comparison := bytes.Compare(result.Folders[left].Folder.ID[:], result.Folders[right].Folder.ID[:])
		if comparison != 0 {
			return comparison < 0
		}
		return result.Folders[left].Snapshot.FolderID < result.Folders[right].Snapshot.FolderID
	})

	messageIDs := make(map[uuid.UUID]struct{}, len(result.Upserts)+len(result.Tombstones))
	engineMessageIDs := make(map[string]struct{}, len(result.Upserts)+len(result.Tombstones))
	for index := range result.Upserts {
		upsert := &result.Upserts[index]
		if _, err := syncValidateMailboxID(upsert.Message, accountID); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if _, err := syncValidateMailboxID(upsert.Thread, accountID); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if err := syncValidateProjection(upsert.Projection); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if err := canonicalValidateThreadIdentity(&upsert.ThreadIdentity, upsert.Projection); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if err := setEngineAccountID(upsert.Projection.AccountID); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if upsert.FolderIDs == nil || len(upsert.FolderIDs) == 0 || upsert.Labels == nil {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		folders, err := canonicalMailboxIDSet(accountID, upsert.FolderIDs)
		if err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		labels, err := syncValidateLabelStates(accountID, upsert.Labels)
		if err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if upsert.ReceivedEventID != nil && *upsert.ReceivedEventID == uuid.Nil {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, duplicate := messageIDs[upsert.Message.ID]; duplicate {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		if _, duplicate := engineMessageIDs[upsert.Projection.MessageID]; duplicate {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		messageIDs[upsert.Message.ID] = struct{}{}
		engineMessageIDs[upsert.Projection.MessageID] = struct{}{}
		upsert.FolderIDs = folders
		upsert.Labels = labels
		if err := canonicalizeProjectionSets(&upsert.Projection); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
	}
	sort.Slice(result.Upserts, func(left, right int) bool {
		comparison := bytes.Compare(result.Upserts[left].Message.ID[:], result.Upserts[right].Message.ID[:])
		if comparison != 0 {
			return comparison < 0
		}
		return result.Upserts[left].Projection.MessageID < result.Upserts[right].Projection.MessageID
	})

	for index := range result.Tombstones {
		tombstone := &result.Tombstones[index]
		if !syncValidOpaque(tombstone.EngineMessageID, MaxEngineOpaqueIDBytes) || tombstone.Labels == nil {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if _, err := syncValidateMailboxID(tombstone.Message, accountID); err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		labels, err := syncValidateLabelStates(accountID, tombstone.Labels)
		if err != nil {
			return MailboxSyncPolicyResult{}, err
		}
		if _, duplicate := messageIDs[tombstone.Message.ID]; duplicate {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		if _, duplicate := engineMessageIDs[tombstone.EngineMessageID]; duplicate {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		messageIDs[tombstone.Message.ID] = struct{}{}
		engineMessageIDs[tombstone.EngineMessageID] = struct{}{}
		tombstone.Labels = labels
	}
	sort.Slice(result.Tombstones, func(left, right int) bool {
		comparison := bytes.Compare(result.Tombstones[left].Message.ID[:], result.Tombstones[right].Message.ID[:])
		if comparison != 0 {
			return comparison < 0
		}
		return result.Tombstones[left].EngineMessageID < result.Tombstones[right].EngineMessageID
	})

	counts, err := canonicalizeCountIntent(accountID, result.Counts)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	result.Counts = counts
	folderCounts := make(map[uuid.UUID]struct{}, len(result.Counts.Folders))
	for _, value := range result.Counts.Folders {
		folderCounts[value.ID] = struct{}{}
	}
	threadCounts := make(map[uuid.UUID]struct{}, len(result.Counts.Threads))
	for _, value := range result.Counts.Threads {
		threadCounts[value.ID] = struct{}{}
	}
	for _, folder := range result.Folders {
		if _, found := folderCounts[folder.Folder.ID]; !found {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	}
	for _, upsert := range result.Upserts {
		for _, folder := range upsert.FolderIDs {
			if _, found := folderCounts[folder.ID]; !found {
				return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
			}
		}
		if _, found := threadCounts[upsert.Thread.ID]; !found {
			return MailboxSyncPolicyResult{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	}

	events, err := canonicalizeEvents(fence, result.Upserts, result.Tombstones, result.Events)
	if err != nil {
		return MailboxSyncPolicyResult{}, err
	}
	result.Events = events
	return result, nil
}

func canonicalValidateAdvance(value MailboxSyncPolicyAdvance) error {
	if value.SyncEpoch < 0 || (value.Change == nil) == (value.Enumeration == nil) {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if value.Change != nil {
		if !syncValidCursor(value.Change.FromCursor, false) || !syncValidCursor(value.Change.NextCursor, true) {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		return nil
	}
	if !syncValidCursor(value.Enumeration.BaseCursor, false) || !syncValidCursor(value.Enumeration.Checkpoint, true) || !syncValidCursor(value.Enumeration.NextCheckpoint, true) || (!value.Enumeration.Complete && value.Enumeration.NextCheckpoint == "") || (value.Enumeration.Complete && value.Enumeration.NextCheckpoint != "") {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return nil
}

func canonicalValidateThreadIdentity(value *ThreadIdentity, projection EngineMessageProjection) error {
	if value == nil || value.KeyVersion != ThreadKeyVersionV1 || threadKeyIsZero(value.Key) || value.CanonicalParticipants == nil {
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	switch value.Source {
	case ThreadIdentitySourceAncestor:
		if value.NormalizedSubject != "" || len(value.CanonicalParticipants) != 0 {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	case ThreadIdentitySourceFallback:
		if !syncValidSafeText(value.NormalizedSubject, MaxEngineHeaderValueBytes, true) || len(value.CanonicalParticipants) == 0 || len(value.CanonicalParticipants) > MaxEngineAddressItems+1 {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		seen := make(map[string]struct{}, len(value.CanonicalParticipants))
		for _, participant := range value.CanonicalParticipants {
			normalized, err := NormalizeAddress(EngineAddress{Email: participant})
			if err != nil || normalized.Email != participant || normalized.Name != "" {
				return syncPolicyError(MailboxSyncPolicyErrorInvalid)
			}
			if _, duplicate := seen[participant]; duplicate {
				return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
			}
			seen[participant] = struct{}{}
		}
		sort.Strings(value.CanonicalParticipants)
		derived, err := DeriveThreadIdentity(ThreadingInput{
			References: projection.References, InReplyTo: projection.InReplyTo, Subject: projection.Subject,
			From: projection.From, To: projection.To, CC: projection.CC,
			Ancestors: []ThreadAncestor{}, AncestorLookupComplete: true,
		})
		if err != nil || derived.KeyVersion != value.KeyVersion || derived.Key != value.Key || derived.Source != value.Source || derived.NormalizedSubject != value.NormalizedSubject || !canonicalEqualStrings(derived.CanonicalParticipants, value.CanonicalParticipants) {
			return syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	default:
		return syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return nil
}

func canonicalEqualStrings(left, right []string) bool {
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

func canonicalMailboxIDSet(accountID uuid.UUID, values []MailboxEventID) ([]MailboxEventID, error) {
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
	sort.Slice(result, func(left, right int) bool { return bytes.Compare(result[left].ID[:], result[right].ID[:]) < 0 })
	return result, nil
}

func canonicalizeCountIntent(accountID uuid.UUID, value MailboxSyncPolicyCountIntent) (MailboxSyncPolicyCountIntent, error) {
	if value.Folders == nil || value.Threads == nil || len(value.Folders) > MaxEngineFolderItems || len(value.Threads) > MaxEnginePageItems*2 {
		return MailboxSyncPolicyCountIntent{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
	}
	folders, err := canonicalMailboxIDSet(accountID, value.Folders)
	if err != nil {
		return MailboxSyncPolicyCountIntent{}, err
	}
	threads, err := canonicalMailboxIDSet(accountID, value.Threads)
	if err != nil {
		return MailboxSyncPolicyCountIntent{}, err
	}
	return MailboxSyncPolicyCountIntent{Folders: folders, Threads: threads}, nil
}

func canonicalizeProjectionSets(value *EngineMessageProjection) error {
	sort.Strings(value.FolderIDs)
	sort.Slice(value.ReservedKeywords, func(left, right int) bool { return value.ReservedKeywords[left] < value.ReservedKeywords[right] })
	sort.Slice(value.Attachments, func(left, right int) bool {
		if value.Attachments[left].BlobID != value.Attachments[right].BlobID {
			return value.Attachments[left].BlobID < value.Attachments[right].BlobID
		}
		if value.Attachments[left].ContentID != value.Attachments[right].ContentID {
			return value.Attachments[left].ContentID < value.Attachments[right].ContentID
		}
		return bytes.Compare(canonicalAttachmentRecord(value.Attachments[left]), canonicalAttachmentRecord(value.Attachments[right])) < 0
	})
	for index := 1; index < len(value.Attachments); index++ {
		if bytes.Equal(canonicalAttachmentRecord(value.Attachments[index-1]), canonicalAttachmentRecord(value.Attachments[index])) {
			return syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
	}
	return nil
}

func canonicalRestoreProjectionCollections(destination *EngineMessageProjection, source EngineMessageProjection) {
	if source.References != nil && destination.References == nil {
		destination.References = make([]string, 0)
	}
	if source.To != nil && destination.To == nil {
		destination.To = make([]EngineAddress, 0)
	}
	if source.CC != nil && destination.CC == nil {
		destination.CC = make([]EngineAddress, 0)
	}
	if source.BCC != nil && destination.BCC == nil {
		destination.BCC = make([]EngineAddress, 0)
	}
	if source.ReplyTo != nil && destination.ReplyTo == nil {
		destination.ReplyTo = make([]EngineAddress, 0)
	}
	if source.FolderIDs != nil && destination.FolderIDs == nil {
		destination.FolderIDs = make([]string, 0)
	}
	if source.ReservedKeywords != nil && destination.ReservedKeywords == nil {
		destination.ReservedKeywords = make([]EngineKeywordMarker, 0)
	}
	if source.Attachments != nil && destination.Attachments == nil {
		destination.Attachments = make([]EngineProjectionAttachment, 0)
	}
}

type canonicalEventRecord struct {
	envelopeBytes []byte
	eventID       uuid.UUID
	index         int
}

type canonicalDecodedEvent struct {
	envelopeBytes []byte
	eventID       uuid.UUID
	subject       string
	received      *mailReceivedEventJSON
	changed       *mailMessageChangedEventJSON
}

func canonicalizeEvents(fence MailboxSyncFence, upserts []MailboxSyncPolicyMessageIntent, tombstones []MailboxSyncPolicyTombstoneIntent, values []events.EnvelopeV1) ([]events.EnvelopeV1, error) {
	result := make([]events.EnvelopeV1, len(values))
	copy(result, values)
	records := make([]canonicalEventRecord, len(result))
	seen := make(map[uuid.UUID]struct{}, len(result))
	emittedMessages := make(map[uuid.UUID]struct{}, len(result))
	for index, value := range result {
		decoded, err := canonicalDecodeEventEnvelope(fence, value)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[decoded.eventID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[decoded.eventID] = struct{}{}

		var messageID uuid.UUID
		switch decoded.subject {
		case "mail.received.v1":
			messageID, err = canonicalAssociateReceivedEvent(decoded.eventID, *decoded.received, upserts)
		case "mail.message.changed.v1":
			messageID, err = canonicalAssociateChangedEvent(*decoded.changed, upserts, tombstones)
		default:
			err = syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if err != nil {
			return nil, err
		}
		if _, duplicate := emittedMessages[messageID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		emittedMessages[messageID] = struct{}{}
		records[index] = canonicalEventRecord{envelopeBytes: decoded.envelopeBytes, eventID: decoded.eventID, index: index}
	}
	sort.Slice(records, func(left, right int) bool {
		comparison := bytes.Compare(records[left].eventID[:], records[right].eventID[:])
		if comparison != 0 {
			return comparison < 0
		}
		return bytes.Compare(records[left].envelopeBytes, records[right].envelopeBytes) < 0
	})
	for index, record := range records {
		result[index] = values[record.index]
	}
	return result, nil
}

type canonicalEnvelopeJSON struct {
	Subject        *string         `json:"subject"`
	SchemaVersion  *int            `json:"schema_version"`
	EventID        *uuid.UUID      `json:"event_id"`
	OccurredAt     *time.Time      `json:"occurred_at"`
	OrganizationID *uuid.UUID      `json:"organization_id"`
	ActorID        json.RawMessage `json:"actor_id"`
	CorrelationID  *string         `json:"correlation_id"`
	Payload        json.RawMessage `json:"payload"`
}

func canonicalEventEnvelope(fence MailboxSyncFence, value events.EnvelopeV1) ([]byte, error) {
	decoded, err := canonicalDecodeEventEnvelope(fence, value)
	if err != nil {
		return nil, err
	}
	return decoded.envelopeBytes, nil
}

func canonicalDecodeEventEnvelope(fence MailboxSyncFence, value events.EnvelopeV1) (canonicalDecodedEvent, error) {
	subject := value.Subject()
	eventID := value.EventID()
	if (subject != "mail.received.v1" && subject != "mail.message.changed.v1") || eventID == uuid.Nil {
		return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	frozen, err := value.MarshalJSON()
	if err != nil {
		return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if len(frozen) == 0 || len(frozen) > MaxEngineProjectionItemBytes || !json.Valid(frozen) {
		if len(frozen) > MaxEngineProjectionItemBytes {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
		}
		return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	var decoded canonicalEnvelopeJSON
	if err := json.Unmarshal(frozen, &decoded); err != nil || decoded.Subject == nil || decoded.SchemaVersion == nil || decoded.EventID == nil || decoded.OccurredAt == nil || decoded.OrganizationID == nil || decoded.CorrelationID == nil || decoded.ActorID == nil || decoded.Payload == nil || *decoded.Subject != subject || *decoded.SchemaVersion != 1 || *decoded.EventID != eventID || *decoded.OrganizationID == uuid.Nil || !syncValidEventOccurredAt(*decoded.OccurredAt) || !syncValidCorrelationID(*decoded.CorrelationID) {
		return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if *decoded.CorrelationID != fence.LeaseID.String() {
		return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	actorJSON := bytes.TrimSpace(decoded.ActorID)
	var actorID *uuid.UUID
	if !bytes.Equal(actorJSON, []byte("null")) {
		var parsedActorID uuid.UUID
		if err := json.Unmarshal(actorJSON, &parsedActorID); err != nil || parsedActorID == uuid.Nil {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		actorID = &parsedActorID
	}
	payload := bytes.TrimSpace(decoded.Payload)
	if len(payload) < 2 || payload[0] != '{' || payload[len(payload)-1] != '}' || !json.Valid(payload) {
		return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}

	result := canonicalDecodedEvent{
		eventID: eventID,
		subject: subject,
		envelopeBytes: canonicalRecord(
			canonicalTLV(1, []byte(subject)),
			canonicalTLV(2, canonicalUUID(eventID)),
			canonicalTLV(3, append([]byte(nil), frozen...)),
		),
	}
	switch subject {
	case "mail.received.v1":
		var received mailReceivedEventJSON
		if err := json.Unmarshal(payload, &received); err != nil {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if received.MailboxID == uuid.Nil {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if received.MailboxID != fence.AccountID {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorAccountMismatch)
		}
		rebuilt, err := NewMailReceivedEnvelopeV1(eventID, *decoded.OccurredAt, *decoded.OrganizationID, actorID, *decoded.CorrelationID, canonicalReceivedPayload(fence.AccountID, received))
		if err != nil {
			return canonicalDecodedEvent{}, syncMapMailEventError(err)
		}
		if !canonicalFrozenEnvelopeEqual(frozen, rebuilt) {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		result.received = &received
	case "mail.message.changed.v1":
		var changed mailMessageChangedEventJSON
		if err := json.Unmarshal(payload, &changed); err != nil {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if changed.MailboxID == uuid.Nil {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		if changed.MailboxID != fence.AccountID {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorAccountMismatch)
		}
		rebuilt, err := NewMailMessageChangedEnvelopeV1(eventID, *decoded.OccurredAt, *decoded.OrganizationID, actorID, *decoded.CorrelationID, canonicalChangedPayload(fence.AccountID, changed))
		if err != nil {
			return canonicalDecodedEvent{}, syncMapMailEventError(err)
		}
		if !canonicalFrozenEnvelopeEqual(frozen, rebuilt) {
			return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		result.changed = &changed
	default:
		return canonicalDecodedEvent{}, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return result, nil
}

func canonicalFrozenEnvelopeEqual(frozen []byte, value events.EnvelopeV1) bool {
	rebuilt, err := value.MarshalJSON()
	return err == nil && bytes.Equal(frozen, rebuilt)
}

func canonicalReceivedPayload(accountID uuid.UUID, value mailReceivedEventJSON) MailReceivedEventPayload {
	return MailReceivedEventPayload{
		Mailbox:        MailboxEventID{AccountID: accountID, ID: value.MailboxID},
		Message:        MailboxEventID{AccountID: accountID, ID: value.MessageID},
		EngineMessage:  MailboxEventOpaqueID{AccountID: accountID, ID: value.EngineMessageID},
		Thread:         MailboxEventID{AccountID: accountID, ID: value.ThreadID},
		From:           canonicalMailEventAddress(value.From),
		To:             canonicalMailEventAddresses(value.To),
		ReceivedAt:     value.ReceivedAt,
		FolderIDs:      canonicalMailboxEventIDs(accountID, value.FolderIDs),
		HasAttachments: value.HasAttachments,
		CalendarMethod:  copyMailCalendarMethod(value.CalendarMethod),
	}
}

func canonicalChangedPayload(accountID uuid.UUID, value mailMessageChangedEventJSON) MailMessageChangedEventPayload {
	return MailMessageChangedEventPayload{
		Mailbox:       MailboxEventID{AccountID: accountID, ID: value.MailboxID},
		Message:       MailboxEventID{AccountID: accountID, ID: value.MessageID},
		ChangedFields: append([]MailChangedField(nil), value.ChangedFields...),
		FolderIDs:     canonicalMailboxEventIDs(accountID, value.FolderIDs),
		Flags:         append([]MailMessageFlag(nil), value.Flags...),
		LabelIDs:      canonicalMailboxEventIDs(accountID, value.LabelIDs),
	}
}

func canonicalMailEventAddress(value mailEventAddressJSON) MailEventAddress {
	return MailEventAddress{Email: value.Email, Name: copyMailEventName(value.Name)}
}

func canonicalMailEventAddresses(values []mailEventAddressJSON) []MailEventAddress {
	result := make([]MailEventAddress, len(values))
	for index, value := range values {
		result[index] = canonicalMailEventAddress(value)
	}
	return result
}

func canonicalMailboxEventIDs(accountID uuid.UUID, values []uuid.UUID) []MailboxEventID {
	result := make([]MailboxEventID, len(values))
	for index, value := range values {
		result[index] = MailboxEventID{AccountID: accountID, ID: value}
	}
	return result
}

func canonicalAssociateReceivedEvent(eventID uuid.UUID, value mailReceivedEventJSON, upserts []MailboxSyncPolicyMessageIntent) (uuid.UUID, error) {
	match := -1
	for index := range upserts {
		if upserts[index].ReceivedEventID == nil || *upserts[index].ReceivedEventID != eventID {
			continue
		}
		if match != -1 {
			return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		match = index
	}
	if match == -1 {
		return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	upsert := upserts[match]
	if value.MessageID != upsert.Message.ID || value.ThreadID != upsert.Thread.ID || value.EngineMessageID != upsert.Projection.MessageID || upsert.Projection.ReceivedAt == nil || !value.ReceivedAt.Equal(*upsert.Projection.ReceivedAt) || !canonicalEqualUUIDs(value.FolderIDs, canonicalMailboxIDs(upsert.FolderIDs)) || !canonicalEventAddressMatches(value.From, upsert.Projection.From) || !canonicalEventAddressesMatch(value.To, upsert.Projection.To) || value.HasAttachments != (len(upsert.Projection.Attachments) != 0) {
		return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return upsert.Message.ID, nil
}

func canonicalAssociateChangedEvent(value mailMessageChangedEventJSON, upserts []MailboxSyncPolicyMessageIntent, tombstones []MailboxSyncPolicyTombstoneIntent) (uuid.UUID, error) {
	if hasMailChangedField(value.ChangedFields, MailChangedFieldDeleted) {
		match := -1
		for index := range tombstones {
			if tombstones[index].Message.ID != value.MessageID {
				continue
			}
			if match != -1 {
				return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
			}
			match = index
		}
		if match == -1 || !canonicalEqualUUIDs(value.LabelIDs, canonicalLabelIDs(tombstones[match].Labels)) {
			return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
		return tombstones[match].Message.ID, nil
	}

	match := -1
	for index := range upserts {
		if upserts[index].Message.ID != value.MessageID {
			continue
		}
		if match != -1 {
			return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		match = index
	}
	if match == -1 || !canonicalEqualUUIDs(value.FolderIDs, canonicalMailboxIDs(upserts[match].FolderIDs)) || !canonicalEqualFlags(value.Flags, syncEventFlags(upserts[match].Projection.Flags)) || !canonicalEqualUUIDs(value.LabelIDs, canonicalLabelIDs(upserts[match].Labels)) {
		return uuid.Nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return upserts[match].Message.ID, nil
}

func canonicalMailboxIDs(values []MailboxEventID) []uuid.UUID {
	result := make([]uuid.UUID, len(values))
	for index, value := range values {
		result[index] = value.ID
	}
	return result
}

func canonicalLabelIDs(values []MailboxSyncPolicyLabelState) []uuid.UUID {
	result := make([]uuid.UUID, len(values))
	for index, value := range values {
		result[index] = value.Label.ID
	}
	return result
}

func canonicalEqualUUIDs(left, right []uuid.UUID) bool {
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

func canonicalEqualFlags(left, right []MailMessageFlag) bool {
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

func canonicalEventAddressMatches(value mailEventAddressJSON, projection EngineAddress) bool {
	if value.Email != projection.Email {
		return false
	}
	if projection.Name == "" {
		return value.Name == nil
	}
	return value.Name != nil && *value.Name == projection.Name
}

func canonicalEventAddressesMatch(values []mailEventAddressJSON, projections []EngineAddress) bool {
	if len(values) != len(projections) {
		return false
	}
	for index := range values {
		if !canonicalEventAddressMatches(values[index], projections[index]) {
			return false
		}
	}
	return true
}

func canonicalMailboxSyncApplyStream(fence MailboxSyncFence, nextAttemptAt time.Time, result MailboxSyncPolicyResult) ([]byte, error) {
	encodedResult, err := canonicalResultRecord(fence, result)
	if err != nil {
		return nil, err
	}
	return canonicalRecord(
		canonicalTLV(1, []byte(mailboxSyncCanonicalApplyDomain)),
		canonicalTLV(2, canonicalU16(uint16(mailboxSyncCanonicalApplyVersion))),
		canonicalTLV(3, canonicalFenceRecord(fence)),
		canonicalTLV(4, canonicalTime(nextAttemptAt)),
		canonicalTLV(5, encodedResult),
	), nil
}

func canonicalFenceRecord(value MailboxSyncFence) []byte {
	return canonicalRecord(
		canonicalTLV(1, canonicalUUID(value.AccountID)),
		canonicalTLV(2, canonicalUUID(value.LeaseID)),
		canonicalTLV(3, canonicalI64(value.LeaseEpoch)),
		canonicalTLV(4, canonicalI64(value.SyncEpoch)),
		canonicalTLV(5, canonicalI64(value.MutationEpoch)),
		canonicalTLV(6, []byte{byte(value.Phase)}),
		canonicalTLV(7, []byte(value.Cursor)),
		canonicalTLV(8, []byte(value.BaseCursor)),
		canonicalTLV(9, []byte(value.Checkpoint)),
	)
}

func canonicalResultRecord(fence MailboxSyncFence, value MailboxSyncPolicyResult) ([]byte, error) {
	folders := make([][]byte, len(value.Folders))
	for index, folder := range value.Folders {
		folders[index] = canonicalFolderIntentRecord(folder)
	}
	upserts := make([][]byte, len(value.Upserts))
	for index, upsert := range value.Upserts {
		upserts[index] = canonicalUpsertRecord(upsert)
	}
	tombstones := make([][]byte, len(value.Tombstones))
	for index, tombstone := range value.Tombstones {
		tombstones[index] = canonicalTombstoneRecord(tombstone)
	}
	events := make([][]byte, len(value.Events))
	for index, event := range value.Events {
		encoded, err := canonicalEventEnvelope(fence, event)
		if err != nil {
			return nil, err
		}
		events[index] = encoded
	}
	return canonicalRecord(
		canonicalTLV(1, canonicalAdvanceRecord(value.Advance)),
		canonicalTLV(2, canonicalCollection(folders)),
		canonicalTLV(3, canonicalCollection(upserts)),
		canonicalTLV(4, canonicalCollection(tombstones)),
		canonicalTLV(5, canonicalCountRecord(value.Counts)),
		canonicalTLV(6, canonicalCollection(events)),
	), nil
}

func canonicalAdvanceRecord(value MailboxSyncPolicyAdvance) []byte {
	change := []byte{0}
	if value.Change != nil {
		change = append([]byte{1}, canonicalRecord(canonicalTLV(1, []byte(value.Change.FromCursor)), canonicalTLV(2, []byte(value.Change.NextCursor)))...)
	}
	enumeration := []byte{0}
	if value.Enumeration != nil {
		enumeration = append([]byte{1}, canonicalRecord(canonicalTLV(1, []byte(value.Enumeration.BaseCursor)), canonicalTLV(2, []byte(value.Enumeration.Checkpoint)), canonicalTLV(3, []byte(value.Enumeration.NextCheckpoint)), canonicalTLV(4, canonicalBool(value.Enumeration.Complete)))...)
	}
	return canonicalRecord(canonicalTLV(1, canonicalI64(value.SyncEpoch)), canonicalTLV(2, change), canonicalTLV(3, enumeration))
}

func canonicalFolderIntentRecord(value MailboxSyncPolicyFolderIntent) []byte {
	return canonicalRecord(
		canonicalTLV(1, canonicalMailboxEventIDRecord(value.Folder)),
		canonicalTLV(2, canonicalRecord(canonicalTLV(1, []byte(value.Snapshot.AccountID)), canonicalTLV(2, []byte(value.Snapshot.FolderID)), canonicalTLV(3, []byte(value.Snapshot.ParentID)), canonicalTLV(4, []byte(value.Snapshot.Name)), canonicalTLV(5, []byte(value.Snapshot.Role)))),
	)
}

func canonicalUpsertRecord(value MailboxSyncPolicyMessageIntent) []byte {
	participants := make([][]byte, len(value.ThreadIdentity.CanonicalParticipants))
	for index, participant := range value.ThreadIdentity.CanonicalParticipants {
		participants[index] = []byte(participant)
	}
	folders := make([][]byte, len(value.FolderIDs))
	for index, folder := range value.FolderIDs {
		folders[index] = canonicalMailboxEventIDRecord(folder)
	}
	labels := make([][]byte, len(value.Labels))
	for index, label := range value.Labels {
		labels[index] = canonicalLabelStateRecord(label)
	}
	received := []byte{0}
	if value.ReceivedEventID != nil {
		received = append([]byte{1}, canonicalUUID(*value.ReceivedEventID)...)
	}
	return canonicalRecord(
		canonicalTLV(1, canonicalMailboxEventIDRecord(value.Message)),
		canonicalTLV(2, canonicalMailboxEventIDRecord(value.Thread)),
		canonicalTLV(3, canonicalRecord(canonicalTLV(1, []byte{byte(value.ThreadIdentity.KeyVersion)}), canonicalTLV(2, value.ThreadIdentity.Key[:]), canonicalTLV(3, []byte(value.ThreadIdentity.Source)), canonicalTLV(4, []byte(value.ThreadIdentity.NormalizedSubject)), canonicalTLV(5, canonicalCollection(participants)))),
		canonicalTLV(4, canonicalProjectionRecord(value.Projection)),
		canonicalTLV(5, canonicalCollection(folders)),
		canonicalTLV(6, canonicalCollection(labels)),
		canonicalTLV(7, received),
	)
}

func canonicalProjectionRecord(value EngineMessageProjection) []byte {
	references := canonicalStringCollection(value.References)
	to := canonicalAddressCollection(value.To)
	cc := canonicalAddressCollection(value.CC)
	bcc := canonicalAddressCollection(value.BCC)
	replyTo := canonicalAddressCollection(value.ReplyTo)
	folderIDs := canonicalStringCollection(value.FolderIDs)
	keywords := make([][]byte, len(value.ReservedKeywords))
	for index, keyword := range value.ReservedKeywords {
		keywords[index] = []byte(keyword)
	}
	attachments := make([][]byte, len(value.Attachments))
	for index, attachment := range value.Attachments {
		attachments[index] = canonicalAttachmentRecord(attachment)
	}
	return canonicalRecord(
		canonicalTLV(1, []byte(value.AccountID)), canonicalTLV(2, []byte(value.MessageID)), canonicalTLV(3, []byte(value.Version)), canonicalTLV(4, []byte(value.InternetMessageID)), canonicalTLV(5, []byte(value.InReplyTo)), canonicalTLV(6, references),
		canonicalTLV(7, canonicalAddressRecord(value.From)), canonicalTLV(8, to), canonicalTLV(9, cc), canonicalTLV(10, bcc), canonicalTLV(11, replyTo),
		canonicalTLV(12, []byte(value.Subject)), canonicalTLV(13, []byte(value.Snippet)), canonicalTLV(14, canonicalOptionalTime(value.SentAt)), canonicalTLV(15, canonicalOptionalTime(value.ReceivedAt)), canonicalTLV(16, canonicalI64(value.SizeBytes)), canonicalTLV(17, []byte(value.Priority)), canonicalTLV(18, folderIDs),
		canonicalTLV(19, canonicalRecord(canonicalTLV(1, canonicalBool(value.Flags.Seen)), canonicalTLV(2, canonicalBool(value.Flags.Answered)), canonicalTLV(3, canonicalBool(value.Flags.Draft)), canonicalTLV(4, canonicalBool(value.Flags.Flagged)), canonicalTLV(5, canonicalBool(value.Flags.Junk)))),
		canonicalTLV(20, canonicalCollection(keywords)), canonicalTLV(21, canonicalCollection(attachments)), canonicalTLV(22, []byte(value.Origin)),
	)
}

func canonicalAddressCollection(values []EngineAddress) []byte {
	result := make([][]byte, len(values))
	for index, value := range values {
		result[index] = canonicalAddressRecord(value)
	}
	return canonicalCollection(result)
}

func canonicalAddressRecord(value EngineAddress) []byte {
	return canonicalRecord(canonicalTLV(1, []byte(value.Email)), canonicalTLV(2, []byte(value.Name)))
}

func canonicalStringCollection(values []string) []byte {
	result := make([][]byte, len(values))
	for index, value := range values {
		result[index] = []byte(value)
	}
	return canonicalCollection(result)
}

func canonicalAttachmentRecord(value EngineProjectionAttachment) []byte {
	return canonicalRecord(canonicalTLV(1, []byte(value.BlobID)), canonicalTLV(2, []byte(value.ContentID)), canonicalTLV(3, []byte(value.Filename)), canonicalTLV(4, []byte(value.MediaType)), canonicalTLV(5, canonicalI64(value.SizeBytes)), canonicalTLV(6, []byte(value.Disposition)), canonicalTLV(7, value.SHA256))
}

func canonicalLabelStateRecord(value MailboxSyncPolicyLabelState) []byte {
	return canonicalRecord(canonicalTLV(1, canonicalMailboxEventIDRecord(value.Label)), canonicalTLV(2, canonicalBool(value.Manual)), canonicalTLV(3, canonicalBool(value.EngineRule)))
}

func canonicalTombstoneRecord(value MailboxSyncPolicyTombstoneIntent) []byte {
	labels := make([][]byte, len(value.Labels))
	for index, label := range value.Labels {
		labels[index] = canonicalLabelStateRecord(label)
	}
	return canonicalRecord(canonicalTLV(1, []byte(value.EngineMessageID)), canonicalTLV(2, canonicalMailboxEventIDRecord(value.Message)), canonicalTLV(3, canonicalCollection(labels)))
}

func canonicalCountRecord(value MailboxSyncPolicyCountIntent) []byte {
	folders := make([][]byte, len(value.Folders))
	for index, folder := range value.Folders {
		folders[index] = canonicalMailboxEventIDRecord(folder)
	}
	threads := make([][]byte, len(value.Threads))
	for index, thread := range value.Threads {
		threads[index] = canonicalMailboxEventIDRecord(thread)
	}
	return canonicalRecord(canonicalTLV(1, canonicalCollection(folders)), canonicalTLV(2, canonicalCollection(threads)))
}

func canonicalMailboxEventIDRecord(value MailboxEventID) []byte {
	return canonicalRecord(canonicalTLV(1, canonicalUUID(value.AccountID)), canonicalTLV(2, canonicalUUID(value.ID)))
}

func canonicalTLV(tag uint16, payload []byte) []byte {
	result := make([]byte, 10+len(payload))
	binary.BigEndian.PutUint16(result, tag)
	binary.BigEndian.PutUint64(result[2:], uint64(len(payload)))
	copy(result[10:], payload)
	return result
}

func canonicalRecord(fields ...[]byte) []byte { return bytes.Join(fields, nil) }

func canonicalCollection(values [][]byte) []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, uint32(len(values)))
	for _, value := range values {
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(value)))
		result = append(result, length...)
		result = append(result, value...)
	}
	return result
}

func canonicalU16(value uint16) []byte {
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, value)
	return result
}

func canonicalI64(value int64) []byte { return canonicalU64(uint64(value)) }

func canonicalU64(value uint64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, value)
	return result
}

func canonicalBool(value bool) []byte {
	if value {
		return []byte{1}
	}
	return []byte{0}
}

func canonicalUUID(value uuid.UUID) []byte { return append([]byte(nil), value[:]...) }

func canonicalTime(value time.Time) []byte {
	return append(canonicalI64(value.UTC().Unix()), canonicalU32(uint32(value.UTC().Nanosecond()))...)
}

func canonicalU32(value uint32) []byte {
	result := make([]byte, 4)
	binary.BigEndian.PutUint32(result, value)
	return result
}

func canonicalOptionalTime(value *time.Time) []byte {
	if value == nil {
		return []byte{0}
	}
	return append([]byte{1}, canonicalTime(*value)...)
}
