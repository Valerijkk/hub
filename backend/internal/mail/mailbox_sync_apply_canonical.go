package mail

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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

	result, err := canonicalizeMailboxSyncResult(request.Fence.AccountID, request.Result)
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

func canonicalizeMailboxSyncResult(accountID uuid.UUID, value MailboxSyncPolicyResult) (MailboxSyncPolicyResult, error) {
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
		canonicalizeProjectionSets(&upsert.Projection)
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

	events, err := canonicalizeEvents(result.Events)
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

func canonicalizeProjectionSets(value *EngineMessageProjection) {
	sort.Strings(value.FolderIDs)
	sort.Slice(value.ReservedKeywords, func(left, right int) bool { return value.ReservedKeywords[left] < value.ReservedKeywords[right] })
	sort.Slice(value.Attachments, func(left, right int) bool {
		if value.Attachments[left].BlobID != value.Attachments[right].BlobID {
			return value.Attachments[left].BlobID < value.Attachments[right].BlobID
		}
		return value.Attachments[left].ContentID < value.Attachments[right].ContentID
	})
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

func canonicalizeEvents(values []events.EnvelopeV1) ([]events.EnvelopeV1, error) {
	result := make([]events.EnvelopeV1, len(values))
	copy(result, values)
	records := make([]canonicalEventRecord, len(result))
	seen := make(map[uuid.UUID]struct{}, len(result))
	for index, value := range result {
		envelopeBytes, err := canonicalEventEnvelope(value)
		if err != nil {
			return nil, err
		}
		eventID := value.EventID()
		if _, duplicate := seen[eventID]; duplicate {
			return nil, syncPolicyError(MailboxSyncPolicyErrorDuplicate)
		}
		seen[eventID] = struct{}{}
		records[index] = canonicalEventRecord{envelopeBytes: envelopeBytes, eventID: eventID, index: index}
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

func canonicalEventEnvelope(value events.EnvelopeV1) ([]byte, error) {
	subject := value.Subject()
	eventID := value.EventID()
	if !canonicalValidEventSubject(subject) || eventID == uuid.Nil {
		return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	frozen, err := value.MarshalJSON()
	if err != nil {
		return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if len(frozen) == 0 || len(frozen) > MaxEngineProjectionItemBytes || !json.Valid(frozen) {
		if len(frozen) > MaxEngineProjectionItemBytes {
			return nil, syncPolicyError(MailboxSyncPolicyErrorTooLarge)
		}
		return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	var decoded canonicalEnvelopeJSON
	if err := json.Unmarshal(frozen, &decoded); err != nil || decoded.Subject == nil || decoded.SchemaVersion == nil || decoded.EventID == nil || decoded.OccurredAt == nil || decoded.OrganizationID == nil || decoded.CorrelationID == nil || decoded.ActorID == nil || decoded.Payload == nil || *decoded.Subject != subject || *decoded.SchemaVersion != 1 || *decoded.EventID != eventID || *decoded.OrganizationID == uuid.Nil || !syncValidEventOccurredAt(*decoded.OccurredAt) || !syncValidCorrelationID(*decoded.CorrelationID) {
		return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	if bytes.Equal(decoded.ActorID, []byte("null")) {
		// The event envelope permits an absent actor.
	} else {
		var actorID uuid.UUID
		if err := json.Unmarshal(decoded.ActorID, &actorID); err != nil || actorID == uuid.Nil {
			return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
		}
	}
	payload := bytes.TrimSpace(decoded.Payload)
	if len(payload) < 2 || payload[0] != '{' || payload[len(payload)-1] != '}' || !json.Valid(payload) {
		return nil, syncPolicyError(MailboxSyncPolicyErrorInvalid)
	}
	return canonicalRecord(
		canonicalTLV(1, []byte(subject)),
		canonicalTLV(2, canonicalUUID(eventID)),
		canonicalTLV(3, append([]byte(nil), frozen...)),
	), nil
}

func canonicalValidEventSubject(value string) bool {
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > 255 || !strings.HasSuffix(value, ".v1") {
		return false
	}
	for _, token := range strings.Split(value, ".") {
		if token == "" {
			return false
		}
	}
	for _, character := range value {
		if character == '*' || character == '>' || unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func canonicalMailboxSyncApplyStream(fence MailboxSyncFence, nextAttemptAt time.Time, result MailboxSyncPolicyResult) ([]byte, error) {
	encodedResult, err := canonicalResultRecord(result)
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

func canonicalResultRecord(value MailboxSyncPolicyResult) ([]byte, error) {
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
		encoded, err := canonicalEventEnvelope(event)
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
