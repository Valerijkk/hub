package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"reflect"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagelabel"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/schema"
)

const mailboxEntApplySyncEpochMaximum int64 = 2_147_483_647

type mailboxEntApplyMessageProjection struct {
	engineMessageID   string
	engineVersion     string
	internetMessageID *string
	inReplyTo         *string
	references        schema.MailMessageReferences
	from              schema.MailStoredAddress
	to                schema.MailStoredAddresses
	cc                schema.MailStoredAddresses
	bcc               schema.MailStoredAddresses
	replyTo           schema.MailStoredAddresses
	subject           string
	snippet           string
	sentAt            *time.Time
	receivedAt        *time.Time
	sizeBytes         int64
	priority          mailmessage.Priority
	flags             EngineMessageFlags
	origin            mailmessage.ProjectionOrigin
	calendarMethod    *mailmessage.CalendarMethod
	receivedEventID   *uuid.UUID
	attachments       []EngineProjectionAttachment
}

type mailboxEntApplyMessagePlan struct {
	intent     MailboxSyncPolicyMessageIntent
	projection mailboxEntApplyMessageProjection
	row        *ent.MailMessage
}

type mailboxEntApplyTombstonePlan struct {
	intent MailboxSyncPolicyTombstoneIntent
	row    *ent.MailMessage
}

type mailboxEntApplyThreadKey struct {
	version int32
	key     ThreadKey
}

type mailboxEntApplyThreadPlan struct {
	id              uuid.UUID
	key             mailboxEntApplyThreadKey
	fallback        *ThreadIdentity
	fallbackMessage *mailboxEntApplyMessageProjection
	requiresExisting bool
	row             *ent.MailThread
}

// mailboxEntApplySyncMessages persists the message, link, attachment and tombstone
// portion of one already canonicalized Mail sync apply. Account fencing, aggregate
// recomputation, clock access and transaction ownership belong to later atoms.
func (r *MailboxEntRepository) mailboxEntApplySyncMessages(ctx context.Context, tx *ent.Tx, canonical MailboxSyncCanonicalApply, transactionTime time.Time) error {
	if r == nil || ctx == nil || tx == nil || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return mailboxEntSafeError(err)
	}

	result, err := mailboxEntApplyCanonicalSyncResult(canonical)
	if err != nil {
		return err
	}
	accountID := canonical.Fence.AccountID
	upserts, tombstones, err := mailboxEntApplyMessagePlans(accountID, result)
	if err != nil {
		return err
	}
	rows, err := mailboxEntApplyLoadMessageRows(ctx, tx, accountID, upserts, tombstones)
	if err != nil {
		return err
	}
	threadPlans, err := mailboxEntApplyThreadPlans(accountID, upserts)
	if err != nil {
		return err
	}
	if err := mailboxEntApplyEnsureThreads(ctx, tx, accountID, result.Advance.SyncEpoch, transactionTime, threadPlans, rows); err != nil {
		return err
	}

	for _, plan := range upserts {
		if err := mailboxEntApplyUpsertMessage(ctx, tx, accountID, result.Advance.SyncEpoch, transactionTime, plan); err != nil {
			return err
		}
	}
	for _, plan := range tombstones {
		if err := mailboxEntApplyTombstoneMessage(ctx, tx, accountID, result.Advance.SyncEpoch, transactionTime, plan); err != nil {
			return err
		}
	}
	return nil
}

func mailboxEntApplyCanonicalSyncResult(canonical MailboxSyncCanonicalApply) (MailboxSyncPolicyResult, error) {
	if !validMailboxSyncFence(canonical.Fence) || canonical.Fence.SyncEpoch > mailboxEntApplySyncEpochMaximum || !syncValidTime(canonical.NextAttemptAt) {
		return MailboxSyncPolicyResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	result, err := canonicalizeMailboxSyncResult(canonical.Fence, canonical.Result)
	if err != nil || !reflect.DeepEqual(result, canonical.Result) {
		return MailboxSyncPolicyResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	expectedPhase, ok := mailboxSyncExpectedPhase(canonical.Fence, result.Advance)
	if !ok || expectedPhase != canonical.ExpectedPhase || result.Advance.SyncEpoch != canonical.Fence.SyncEpoch || result.Advance.SyncEpoch > mailboxEntApplySyncEpochMaximum {
		return MailboxSyncPolicyResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	stream, err := canonicalMailboxSyncApplyStream(canonical.Fence, canonical.NextAttemptAt, result)
	if err != nil || sha256.Sum256(stream) != canonical.ResultHash {
		return MailboxSyncPolicyResult{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return result, nil
}

func mailboxEntApplyMessagePlans(accountID uuid.UUID, result MailboxSyncPolicyResult) ([]*mailboxEntApplyMessagePlan, []*mailboxEntApplyTombstonePlan, error) {
	if accountID == uuid.Nil || result.Upserts == nil || result.Tombstones == nil || len(result.Upserts) > MaxEnginePageItems || len(result.Tombstones) > MaxEnginePageItems || len(result.Upserts)+len(result.Tombstones) > MaxEnginePageItems {
		return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	upserts := make([]*mailboxEntApplyMessagePlan, len(result.Upserts))
	tombstones := make([]*mailboxEntApplyTombstonePlan, len(result.Tombstones))
	messageIDs := make(map[uuid.UUID]struct{}, len(result.Upserts)+len(result.Tombstones))
	engineIDs := make(map[string]struct{}, len(result.Upserts)+len(result.Tombstones))
	for index, intent := range result.Upserts {
		if err := validateMailboxEventIDForAccount(accountID, intent.Message); err != nil || validateMailboxEventIDForAccount(accountID, intent.Thread) != nil || intent.FolderIDs == nil || intent.Labels == nil || intent.ChangedFields == nil {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if err := canonicalValidateThreadIdentity(&intent.ThreadIdentity, intent.Projection); err != nil || !canonicalEqualCalendarMethods(intent.CalendarMethod, intent.Projection.CalendarMethod) || canonicalValidateUpsertChangedFields(intent.ChangedFields) != nil {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		projection, err := mailboxEntApplyProjectionIntent(intent)
		if err != nil {
			return nil, nil, err
		}
		if _, duplicate := messageIDs[intent.Message.ID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := engineIDs[projection.engineMessageID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		messageIDs[intent.Message.ID] = struct{}{}
		engineIDs[projection.engineMessageID] = struct{}{}
		upserts[index] = &mailboxEntApplyMessagePlan{intent: intent, projection: projection}
	}
	for index, intent := range result.Tombstones {
		labels, err := syncValidateLabelStates(accountID, intent.Labels)
		if err != nil || !reflect.DeepEqual(labels, intent.Labels) || validateMailboxEventIDForAccount(accountID, intent.Message) != nil || !syncValidOpaque(intent.EngineMessageID, MaxEngineOpaqueIDBytes) {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := messageIDs[intent.Message.ID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := engineIDs[intent.EngineMessageID]; duplicate {
			return nil, nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		messageIDs[intent.Message.ID] = struct{}{}
		engineIDs[intent.EngineMessageID] = struct{}{}
		tombstones[index] = &mailboxEntApplyTombstonePlan{intent: intent}
	}
	return upserts, tombstones, nil
}

func mailboxEntApplyProjectionIntent(intent MailboxSyncPolicyMessageIntent) (mailboxEntApplyMessageProjection, error) {
	projection := intent.Projection
	if err := syncValidateProjection(projection); err != nil {
		return mailboxEntApplyMessageProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	references := make(schema.MailMessageReferences, len(projection.References))
	copy(references, projection.References)
	if err := references.Validate(); err != nil {
		return mailboxEntApplyMessageProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	from := schema.MailStoredAddress{Email: projection.From.Email, Name: projection.From.Name}
	if err := from.Validate(); err != nil {
		return mailboxEntApplyMessageProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	to, err := mailboxEntApplyStoredAddresses(projection.To)
	if err != nil {
		return mailboxEntApplyMessageProjection{}, err
	}
	cc, err := mailboxEntApplyStoredAddresses(projection.CC)
	if err != nil {
		return mailboxEntApplyMessageProjection{}, err
	}
	bcc, err := mailboxEntApplyStoredAddresses(projection.BCC)
	if err != nil {
		return mailboxEntApplyMessageProjection{}, err
	}
	replyTo, err := mailboxEntApplyStoredAddresses(projection.ReplyTo)
	if err != nil {
		return mailboxEntApplyMessageProjection{}, err
	}
	if intent.ReceivedEventID != nil && *intent.ReceivedEventID == uuid.Nil {
		return mailboxEntApplyMessageProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	attachments := make([]EngineProjectionAttachment, len(projection.Attachments))
	copy(attachments, projection.Attachments)
	calendarMethod := mailboxEntApplyCalendarMethodPointer(intent.CalendarMethod)
	return mailboxEntApplyMessageProjection{
		engineMessageID:   projection.MessageID,
		engineVersion:     projection.Version,
		internetMessageID: mailboxEntApplyOptionalString(projection.InternetMessageID),
		inReplyTo:         mailboxEntApplyOptionalString(projection.InReplyTo),
		references:        references,
		from:              from,
		to:                to,
		cc:                cc,
		bcc:               bcc,
		replyTo:           replyTo,
		subject:           projection.Subject,
		snippet:           projection.Snippet,
		sentAt:            mailboxEntApplyTimePointer(projection.SentAt),
		receivedAt:        mailboxEntApplyTimePointer(projection.ReceivedAt),
		sizeBytes:         projection.SizeBytes,
		priority:          mailmessage.Priority(projection.Priority),
		flags:             projection.Flags,
		origin:            mailmessage.ProjectionOrigin(projection.Origin),
		calendarMethod:    calendarMethod,
		receivedEventID:   syncCopyUUIDPointer(intent.ReceivedEventID),
		attachments:       attachments,
	}, nil
}

func mailboxEntApplyStoredAddresses(values []EngineAddress) (schema.MailStoredAddresses, error) {
	addresses := make(schema.MailStoredAddresses, len(values))
	for index, value := range values {
		addresses[index] = schema.MailStoredAddress{Email: value.Email, Name: value.Name}
	}
	if err := addresses.Validate(); err != nil {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return addresses, nil
}

func mailboxEntApplyLoadMessageRows(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, upserts []*mailboxEntApplyMessagePlan, tombstones []*mailboxEntApplyTombstonePlan) ([]*ent.MailMessage, error) {
	messageIDSet := make(map[uuid.UUID]struct{}, len(upserts)+len(tombstones))
	engineIDSet := make(map[string]struct{}, len(upserts)+len(tombstones))
	for _, plan := range upserts {
		messageIDSet[plan.intent.Message.ID] = struct{}{}
		engineIDSet[plan.projection.engineMessageID] = struct{}{}
	}
	for _, plan := range tombstones {
		messageIDSet[plan.intent.Message.ID] = struct{}{}
		engineIDSet[plan.intent.EngineMessageID] = struct{}{}
	}
	messageIDs := mailboxEntApplySortedUUIDs(messageIDSet)
	engineIDs := mailboxEntApplySortedStrings(engineIDSet)
	byID := make(map[uuid.UUID]*ent.MailMessage, len(messageIDs))
	allRows := make(map[uuid.UUID]*ent.MailMessage, len(messageIDs))
	if len(messageIDs) != 0 {
		rows, err := tx.MailMessage.Query().
			Where(mailmessage.IDIn(messageIDs...)).
			Order(ent.Asc(mailmessage.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return nil, mailboxEntSafeError(err)
		}
		for _, row := range rows {
			if !mailboxEntApplyMessageRowValid(row, accountID) {
				return nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if _, duplicate := byID[row.ID]; duplicate {
				return nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			byID[row.ID] = row
			allRows[row.ID] = row
		}
	}
	byEngineID := make(map[string]*ent.MailMessage, len(engineIDs))
	if len(engineIDs) != 0 {
		rows, err := tx.MailMessage.Query().
			Where(mailmessage.MailAccountIDEQ(accountID), mailmessage.EngineMessageIDIn(engineIDs...)).
			Order(ent.Asc(mailmessage.FieldID)).
			ForUpdate().
			All(ctx)
		if err != nil {
			return nil, mailboxEntSafeError(err)
		}
		for _, row := range rows {
			if !mailboxEntApplyMessageRowValid(row, accountID) {
				return nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if _, duplicate := byEngineID[row.EngineMessageID]; duplicate {
				return nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			byEngineID[row.EngineMessageID] = row
			if prior, present := allRows[row.ID]; present && prior != row {
				return nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			allRows[row.ID] = row
		}
	}
	for _, plan := range upserts {
		byMessage := byID[plan.intent.Message.ID]
		byEngine := byEngineID[plan.projection.engineMessageID]
		if byMessage != nil && byEngine != nil && byMessage.ID != byEngine.ID {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if byMessage != nil && byMessage.EngineMessageID != plan.projection.engineMessageID {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if byEngine != nil && byEngine.ID != plan.intent.Message.ID {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		plan.row = byMessage
	}
	for _, plan := range tombstones {
		byMessage := byID[plan.intent.Message.ID]
		byEngine := byEngineID[plan.intent.EngineMessageID]
		if byMessage == nil || byEngine == nil || byMessage.ID != byEngine.ID || byMessage.DeletedAt != nil || byMessage.EngineMessageID != plan.intent.EngineMessageID {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		plan.row = byMessage
	}
	rows := make([]*ent.MailMessage, 0, len(allRows))
	for _, row := range allRows {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(left, right int) bool { return syncLessUUID(rows[left].ID, rows[right].ID) })
	return rows, nil
}

func mailboxEntApplyThreadPlans(accountID uuid.UUID, upserts []*mailboxEntApplyMessagePlan) ([]*mailboxEntApplyThreadPlan, error) {
	byID := make(map[uuid.UUID]*mailboxEntApplyThreadPlan, len(upserts))
	plans := make([]*mailboxEntApplyThreadPlan, 0, len(upserts))
	for _, message := range upserts {
		if message == nil || validateMailboxEventIDForAccount(accountID, message.intent.Thread) != nil || canonicalValidateThreadIdentity(&message.intent.ThreadIdentity, message.intent.Projection) != nil {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		identity := message.intent.ThreadIdentity
		key := mailboxEntApplyThreadKey{version: int32(identity.KeyVersion), key: identity.Key}
		plan, found := byID[message.intent.Thread.ID]
		if !found {
			plan = &mailboxEntApplyThreadPlan{id: message.intent.Thread.ID, key: key}
			byID[plan.id] = plan
			plans = append(plans, plan)
		} else if plan.key != key {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		switch identity.Source {
		case ThreadIdentitySourceAncestor:
			plan.requiresExisting = true
		case ThreadIdentitySourceFallback:
			if plan.fallback != nil && !mailboxEntApplyFallbackIdentityEqual(*plan.fallback, identity) {
				return nil, mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if plan.fallback == nil {
				copied := mailboxEntApplyCopyThreadIdentity(identity)
				plan.fallback = &copied
				projection := message.projection
				plan.fallbackMessage = &projection
			}
		default:
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}
	sort.Slice(plans, func(left, right int) bool { return syncLessUUID(plans[left].id, plans[right].id) })
	return plans, nil
}

func mailboxEntApplyEnsureThreads(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, transactionTime time.Time, plans []*mailboxEntApplyThreadPlan, messageRows []*ent.MailMessage) error {
	if ctx == nil || tx == nil || accountID == uuid.Nil || syncEpoch < 0 || syncEpoch > mailboxEntApplySyncEpochMaximum || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if len(plans) == 0 {
		return nil
	}
	threadIDSet := make(map[uuid.UUID]struct{}, len(plans)+len(messageRows))
	existingThreadIDs := make(map[uuid.UUID]struct{}, len(messageRows))
	for _, plan := range plans {
		if plan == nil || plan.id == uuid.Nil || plan.key.version != int32(ThreadKeyVersionV1) || threadKeyIsZero(plan.key.key) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		threadIDSet[plan.id] = struct{}{}
	}
	for _, message := range messageRows {
		if message == nil || message.MailThreadID == uuid.Nil {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		threadIDSet[message.MailThreadID] = struct{}{}
		existingThreadIDs[message.MailThreadID] = struct{}{}
	}
	threadIDs := mailboxEntApplySortedUUIDs(threadIDSet)
	byID := make(map[uuid.UUID]*ent.MailThread, len(threadIDs))
	rows, err := tx.MailThread.Query().
		Where(mailthread.IDIn(threadIDs...)).
		Order(ent.Asc(mailthread.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	for _, row := range rows {
		if !mailboxEntApplyThreadRowValid(row, accountID) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := byID[row.ID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		byID[row.ID] = row
	}
	for threadID := range existingThreadIDs {
		if byID[threadID] == nil {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	keyPredicates := make([]predicate.MailThread, 0, len(plans))
	for _, plan := range plans {
		keyPredicates = append(keyPredicates, mailthread.And(
			mailthread.ThreadKeyVersionEQ(plan.key.version),
			mailthread.ThreadKeyEQ(plan.key.key[:]),
		))
	}
	keyRows, err := tx.MailThread.Query().
		Where(mailthread.MailAccountIDEQ(accountID), mailthread.Or(keyPredicates...)).
		Order(ent.Asc(mailthread.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	byKey := make(map[mailboxEntApplyThreadKey]*ent.MailThread, len(keyRows))
	for _, row := range keyRows {
		if !mailboxEntApplyThreadRowValid(row, accountID) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		key, ok := mailboxEntApplyThreadKeyFromRow(row)
		if !ok {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := byKey[key]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		byKey[key] = row
		if existing, found := byID[row.ID]; found && existing != row {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		byID[row.ID] = row
	}
	for _, plan := range plans {
		byTargetID := byID[plan.id]
		byTargetKey := byKey[plan.key]
		if byTargetID != nil {
			key, ok := mailboxEntApplyThreadKeyFromRow(byTargetID)
			if !ok || key != plan.key {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
		}
		if byTargetKey != nil && byTargetKey.ID != plan.id {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if byTargetID == nil {
			if plan.requiresExisting || plan.fallback == nil || plan.fallbackMessage == nil {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			created, err := mailboxEntApplyCreateThread(ctx, tx, accountID, syncEpoch, transactionTime, plan)
			if err != nil {
				return err
			}
			plan.row = created
			continue
		}
		if plan.fallback != nil && !mailboxEntApplyThreadFallbackMetadataMatches(byTargetID, *plan.fallback) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		plan.row = byTargetID
	}
	return nil
}

func mailboxEntApplyCreateThread(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, transactionTime time.Time, plan *mailboxEntApplyThreadPlan) (*ent.MailThread, error) {
	if tx == nil || plan == nil || plan.row != nil || plan.fallback == nil || plan.fallbackMessage == nil || plan.requiresExisting || plan.fallback.Source != ThreadIdentitySourceFallback || plan.key.version != int32(ThreadKeyVersionV1) || threadKeyIsZero(plan.key.key) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	participants := schema.MailThreadParticipants(append([]string(nil), plan.fallback.CanonicalParticipants...))
	if err := participants.Validate(); err != nil || !syncValidSafeText(plan.fallback.NormalizedSubject, MaxEngineHeaderValueBytes, true) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	latest := mailboxEntApplyEffectiveMessageTime(*plan.fallbackMessage, transactionTime)
	unreadCount := int32(0)
	if !plan.fallbackMessage.flags.Seen {
		unreadCount = 1
	}
	created, err := tx.MailThread.Create().
		SetID(plan.id).
		SetMailAccountID(accountID).
		SetThreadKeyVersion(plan.key.version).
		SetThreadKey(append([]byte(nil), plan.key.key[:]...)).
		SetNormalizedSubject(plan.fallback.NormalizedSubject).
		SetParticipantAddresses(participants).
		SetLastMessageAt(latest).
		SetMessageCount(1).
		SetUnreadCount(unreadCount).
		SetIsPinned(false).
		SetHasAttachments(len(plan.fallbackMessage.attachments) != 0).
		SetSnippet(plan.fallbackMessage.snippet).
		SetLastSeenSyncEpoch(syncEpoch).
		SetCreatedAt(transactionTime).
		SetUpdatedAt(transactionTime).
		Save(ctx)
	if err != nil {
		return nil, mailboxEntSafeError(err)
	}
	if !mailboxEntApplyThreadRowValid(created, accountID) || created.ID != plan.id {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return created, nil
}

func mailboxEntApplyThreadRowValid(row *ent.MailThread, accountID uuid.UUID) bool {
	if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || row.ThreadKeyVersion != int32(ThreadKeyVersionV1) || len(row.ThreadKey) != len(ThreadKey{}) || row.MessageCount < 1 || row.UnreadCount < 0 || row.UnreadCount > row.MessageCount || row.LastSeenSyncEpoch < 0 || row.LastSeenSyncEpoch > mailboxEntApplySyncEpochMaximum || !syncValidSafeText(row.NormalizedSubject, MaxEngineHeaderValueBytes, true) || !syncValidTime(row.LastMessageAt) || !syncValidTime(row.CreatedAt) || !syncValidTime(row.UpdatedAt) {
		return false
	}
	if row.DeletedAt != nil && !syncValidTime(*row.DeletedAt) {
		return false
	}
	key, ok := mailboxEntApplyThreadKeyFromRow(row)
	return ok && !threadKeyIsZero(key.key) && row.ParticipantAddresses.Validate() == nil
}

func mailboxEntApplyThreadKeyFromRow(row *ent.MailThread) (mailboxEntApplyThreadKey, bool) {
	if row == nil || row.ThreadKeyVersion != int32(ThreadKeyVersionV1) || len(row.ThreadKey) != len(ThreadKey{}) {
		return mailboxEntApplyThreadKey{}, false
	}
	var key ThreadKey
	copy(key[:], row.ThreadKey)
	if threadKeyIsZero(key) {
		return mailboxEntApplyThreadKey{}, false
	}
	return mailboxEntApplyThreadKey{version: row.ThreadKeyVersion, key: key}, true
}

func mailboxEntApplyThreadFallbackMetadataMatches(row *ent.MailThread, identity ThreadIdentity) bool {
	if row == nil || identity.Source != ThreadIdentitySourceFallback || row.NormalizedSubject != identity.NormalizedSubject || !bytes.Equal(row.ThreadKey, identity.Key[:]) {
		return false
	}
	participants := schema.MailThreadParticipants(identity.CanonicalParticipants)
	return reflect.DeepEqual(row.ParticipantAddresses, participants)
}

func mailboxEntApplyFallbackIdentityEqual(left, right ThreadIdentity) bool {
	return left.Source == ThreadIdentitySourceFallback && right.Source == ThreadIdentitySourceFallback && left.KeyVersion == right.KeyVersion && left.Key == right.Key && left.NormalizedSubject == right.NormalizedSubject && reflect.DeepEqual(left.CanonicalParticipants, right.CanonicalParticipants)
}

func mailboxEntApplyCopyThreadIdentity(value ThreadIdentity) ThreadIdentity {
	return ThreadIdentity{
		KeyVersion:           value.KeyVersion,
		Key:                  value.Key,
		Source:               value.Source,
		NormalizedSubject:    value.NormalizedSubject,
		CanonicalParticipants: append([]string(nil), value.CanonicalParticipants...),
	}
}

func mailboxEntApplyUpsertMessage(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, transactionTime time.Time, plan *mailboxEntApplyMessagePlan) error {
	if ctx == nil || tx == nil || plan == nil || accountID == uuid.Nil || syncEpoch < 0 || syncEpoch > mailboxEntApplySyncEpochMaximum || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return mailboxEntSafeError(err)
	}
	row := plan.row
	created := row == nil
	baseChanged := true
	if created {
		var err error
		row, err = mailboxEntApplyCreateMessage(ctx, tx, accountID, syncEpoch, transactionTime, plan)
		if err != nil {
			return err
		}
	} else {
		if !mailboxEntApplyMessageRowValid(row, accountID) || row.ID != plan.intent.Message.ID || row.EngineMessageID != plan.projection.engineMessageID {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		receivedEventChanged, err := mailboxEntApplyReceivedEventChange(row.ReceivedEventID, plan.projection.receivedEventID)
		if err != nil {
			return err
		}
		baseChanged = row.DeletedAt != nil || receivedEventChanged || !mailboxEntApplyMessageProjectionMatches(row, plan)
		update := row.Update()
		mailboxEntApplySetMessageUpdate(update, plan, syncEpoch, transactionTime, row.HasAttachments)
		if row.DeletedAt != nil {
			update.ClearDeletedAt()
		}
		if baseChanged {
			update.SetUpdatedAt(transactionTime)
		} else {
			update.SetUpdatedAt(row.UpdatedAt)
		}
		if err := update.Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
	}

	folderChanged, err := mailboxEntApplyMessageFolderLinks(ctx, tx, accountID, row.ID, plan.intent.FolderIDs, syncEpoch, transactionTime)
	if err != nil {
		return err
	}
	labelChanged, err := mailboxEntApplyMessageLabelLinks(ctx, tx, accountID, row.ID, plan.intent.Labels, transactionTime)
	if err != nil {
		return err
	}
	attachmentChanged, hasAttachments, err := mailboxEntApplyMessageAttachments(ctx, tx, accountID, row.ID, plan.projection.attachments, transactionTime)
	if err != nil {
		return err
	}
	if hasAttachments != row.HasAttachments || (!created && !baseChanged && (folderChanged || labelChanged || attachmentChanged)) {
		if err := row.Update().SetHasAttachments(hasAttachments).SetUpdatedAt(transactionTime).Exec(ctx); err != nil {
			return mailboxEntSafeError(err)
		}
	}
	return nil
}

func mailboxEntApplyCreateMessage(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, transactionTime time.Time, plan *mailboxEntApplyMessagePlan) (*ent.MailMessage, error) {
	if tx == nil || plan == nil || plan.row != nil || plan.intent.Message.ID == uuid.Nil || plan.intent.Thread.ID == uuid.Nil {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	create := tx.MailMessage.Create().
		SetID(plan.intent.Message.ID).
		SetMailAccountID(accountID).
		SetMailThreadID(plan.intent.Thread.ID).
		SetEngineMessageID(plan.projection.engineMessageID).
		SetEngineVersion(plan.projection.engineVersion).
		SetReferences(plan.projection.references).
		SetFrom(plan.projection.from).
		SetTo(plan.projection.to).
		SetCc(plan.projection.cc).
		SetBcc(plan.projection.bcc).
		SetReplyTo(plan.projection.replyTo).
		SetSubject(plan.projection.subject).
		SetSnippet(plan.projection.snippet).
		SetSizeBytes(plan.projection.sizeBytes).
		SetPriority(plan.projection.priority).
		SetHasAttachments(false).
		SetIsDraft(plan.projection.flags.Draft).
		SetIsSeen(plan.projection.flags.Seen).
		SetIsAnswered(plan.projection.flags.Answered).
		SetIsFlagged(plan.projection.flags.Flagged).
		SetIsJunk(plan.projection.flags.Junk).
		SetProjectionOrigin(plan.projection.origin).
		SetFlagsSyncedAt(transactionTime).
		SetLastSeenSyncEpoch(syncEpoch).
		SetCreatedAt(transactionTime).
		SetUpdatedAt(transactionTime)
	mailboxEntApplySetMessageOptionalCreate(create, plan.projection)
	created, err := create.Save(ctx)
	if err != nil {
		return nil, mailboxEntSafeError(err)
	}
	if !mailboxEntApplyMessageRowValid(created, accountID) || created.ID != plan.intent.Message.ID || created.EngineMessageID != plan.projection.engineMessageID || created.DeletedAt != nil {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return created, nil
}

func mailboxEntApplySetMessageOptionalCreate(create *ent.MailMessageCreate, projection mailboxEntApplyMessageProjection) {
	if projection.internetMessageID != nil {
		create.SetInternetMessageID(*projection.internetMessageID)
	}
	if projection.inReplyTo != nil {
		create.SetInReplyTo(*projection.inReplyTo)
	}
	if projection.sentAt != nil {
		create.SetSentAt(*projection.sentAt)
	}
	if projection.receivedAt != nil {
		create.SetReceivedAt(*projection.receivedAt)
	}
	if projection.calendarMethod != nil {
		create.SetCalendarMethod(*projection.calendarMethod)
	}
	if projection.receivedEventID != nil {
		create.SetReceivedEventID(*projection.receivedEventID)
	}
}

func mailboxEntApplySetMessageUpdate(update *ent.MailMessageUpdateOne, plan *mailboxEntApplyMessagePlan, syncEpoch int64, transactionTime time.Time, hasAttachments bool) {
	projection := plan.projection
	update.SetMailThreadID(plan.intent.Thread.ID).
		SetEngineVersion(projection.engineVersion).
		SetReferences(projection.references).
		SetFrom(projection.from).
		SetTo(projection.to).
		SetCc(projection.cc).
		SetBcc(projection.bcc).
		SetReplyTo(projection.replyTo).
		SetSubject(projection.subject).
		SetSnippet(projection.snippet).
		SetSizeBytes(projection.sizeBytes).
		SetPriority(projection.priority).
		SetHasAttachments(hasAttachments).
		SetIsDraft(projection.flags.Draft).
		SetIsSeen(projection.flags.Seen).
		SetIsAnswered(projection.flags.Answered).
		SetIsFlagged(projection.flags.Flagged).
		SetIsJunk(projection.flags.Junk).
		SetProjectionOrigin(projection.origin).
		SetFlagsSyncedAt(transactionTime).
		SetLastSeenSyncEpoch(syncEpoch)
	if projection.internetMessageID == nil {
		update.ClearInternetMessageID()
	} else {
		update.SetInternetMessageID(*projection.internetMessageID)
	}
	if projection.inReplyTo == nil {
		update.ClearInReplyTo()
	} else {
		update.SetInReplyTo(*projection.inReplyTo)
	}
	if projection.sentAt == nil {
		update.ClearSentAt()
	} else {
		update.SetSentAt(*projection.sentAt)
	}
	if projection.receivedAt == nil {
		update.ClearReceivedAt()
	} else {
		update.SetReceivedAt(*projection.receivedAt)
	}
	if projection.calendarMethod == nil {
		update.ClearCalendarMethod()
	} else {
		update.SetCalendarMethod(*projection.calendarMethod)
	}
	if projection.receivedEventID != nil {
		update.SetReceivedEventID(*projection.receivedEventID)
	}
}

func mailboxEntApplyMessageProjectionMatches(row *ent.MailMessage, plan *mailboxEntApplyMessagePlan) bool {
	if row == nil || plan == nil || row.MailThreadID != plan.intent.Thread.ID || row.EngineMessageID != plan.projection.engineMessageID || row.EngineVersion != plan.projection.engineVersion || !mailboxEntApplyOptionalStringEqual(row.InternetMessageID, plan.projection.internetMessageID) || !mailboxEntApplyOptionalStringEqual(row.InReplyTo, plan.projection.inReplyTo) || !reflect.DeepEqual(row.References, plan.projection.references) || row.From != plan.projection.from || !reflect.DeepEqual(row.To, plan.projection.to) || !reflect.DeepEqual(row.Cc, plan.projection.cc) || !reflect.DeepEqual(row.Bcc, plan.projection.bcc) || !reflect.DeepEqual(row.ReplyTo, plan.projection.replyTo) || row.Subject != plan.projection.subject || row.Snippet != plan.projection.snippet || !mailboxEntApplyOptionalTimeEqual(row.SentAt, plan.projection.sentAt) || !mailboxEntApplyOptionalTimeEqual(row.ReceivedAt, plan.projection.receivedAt) || row.SizeBytes != plan.projection.sizeBytes || row.Priority != plan.projection.priority || row.IsDraft != plan.projection.flags.Draft || row.IsSeen != plan.projection.flags.Seen || row.IsAnswered != plan.projection.flags.Answered || row.IsFlagged != plan.projection.flags.Flagged || row.IsJunk != plan.projection.flags.Junk || row.ProjectionOrigin != plan.projection.origin || !mailboxEntApplyCalendarMethodEqual(row.CalendarMethod, plan.projection.calendarMethod) {
		return false
	}
	return true
}

func mailboxEntApplyReceivedEventChange(current, desired *uuid.UUID) (bool, error) {
	if current == nil {
		return desired != nil, nil
	}
	if *current == uuid.Nil || desired == nil || *desired == uuid.Nil || *current != *desired {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return false, nil
}

func mailboxEntApplyMessageRowValid(row *ent.MailMessage, accountID uuid.UUID) bool {
	if row == nil || row.ID == uuid.Nil || row.MailAccountID != accountID || row.MailThreadID == uuid.Nil || !syncValidOpaque(row.EngineMessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(row.EngineVersion, MaxEngineOpaqueIDBytes) || (row.InternetMessageID != nil && !validThreadMessageID(*row.InternetMessageID)) || (row.InReplyTo != nil && !validThreadMessageID(*row.InReplyTo)) || row.SizeBytes < 0 || row.SizeBytes > MaxEngineMessageBytes || row.LastSeenSyncEpoch < 0 || row.LastSeenSyncEpoch > mailboxEntApplySyncEpochMaximum || !syncValidSafeText(row.Subject, MaxEngineHeaderValueBytes, true) || !syncValidSafeText(row.Snippet, MaxEngineSnippetBytes, true) || !syncValidTime(row.CreatedAt) || !syncValidTime(row.UpdatedAt) {
		return false
	}
	if row.SentAt != nil && !syncValidTime(*row.SentAt) || row.ReceivedAt != nil && !syncValidTime(*row.ReceivedAt) || row.FlagsSyncedAt != nil && !syncValidTime(*row.FlagsSyncedAt) || row.DeletedAt != nil && !syncValidTime(*row.DeletedAt) || row.CalendarEventID != nil && *row.CalendarEventID == uuid.Nil || row.ReceivedEventID != nil && *row.ReceivedEventID == uuid.Nil || row.EngineState != nil && !syncValidSafeText(*row.EngineState, 4_096, false) {
		return false
	}
	if row.Priority != mailmessage.Priority(EnginePriorityLow) && row.Priority != mailmessage.Priority(EnginePriorityNormal) && row.Priority != mailmessage.Priority(EnginePriorityHigh) {
		return false
	}
	if row.ProjectionOrigin != mailmessage.ProjectionOrigin(EngineProjectionOriginInboundOrUnknown) && row.ProjectionOrigin != mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerSubmission) && row.ProjectionOrigin != mailmessage.ProjectionOrigin(EngineProjectionOriginConfirmedOwnerDraft) {
		return false
	}
	if row.CalendarMethod != nil && *row.CalendarMethod != mailmessage.CalendarMethod(MailCalendarMethodRequest) && *row.CalendarMethod != mailmessage.CalendarMethod(MailCalendarMethodCancel) && *row.CalendarMethod != mailmessage.CalendarMethod(MailCalendarMethodReply) {
		return false
	}
	return row.References.Validate() == nil && row.From.Validate() == nil && row.To.Validate() == nil && row.Cc.Validate() == nil && row.Bcc.Validate() == nil && row.ReplyTo.Validate() == nil
}

func mailboxEntApplyTombstoneMessage(ctx context.Context, tx *ent.Tx, accountID uuid.UUID, syncEpoch int64, transactionTime time.Time, plan *mailboxEntApplyTombstonePlan) error {
	if ctx == nil || tx == nil || plan == nil || plan.row == nil || accountID == uuid.Nil || syncEpoch < 0 || syncEpoch > mailboxEntApplySyncEpochMaximum || !mailboxEntValidTime(transactionTime) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return mailboxEntSafeError(err)
	}
	row := plan.row
	if !mailboxEntApplyMessageRowValid(row, accountID) || row.ID != plan.intent.Message.ID || row.EngineMessageID != plan.intent.EngineMessageID || row.DeletedAt != nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := mailboxEntApplyLockMessageThread(ctx, tx, accountID, row.MailThreadID); err != nil {
		return err
	}
	if err := mailboxEntApplyTombstoneLabelsMatch(ctx, tx, accountID, row.ID, plan.intent.Labels); err != nil {
		return err
	}
	if _, err := mailboxEntApplyDeleteMessageFolderLinks(ctx, tx, accountID, row.ID); err != nil {
		return err
	}
	if err := row.Update().SetDeletedAt(transactionTime).SetLastSeenSyncEpoch(syncEpoch).SetUpdatedAt(transactionTime).Exec(ctx); err != nil {
		return mailboxEntSafeError(err)
	}
	return nil
}

func mailboxEntApplyLockMessageThread(ctx context.Context, tx *ent.Tx, accountID, threadID uuid.UUID) error {
	if ctx == nil || tx == nil || accountID == uuid.Nil || threadID == uuid.Nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	thread, err := tx.MailThread.Query().Where(mailthread.IDEQ(threadID), mailthread.MailAccountIDEQ(accountID)).ForUpdate().Only(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if !mailboxEntApplyThreadRowValid(thread, accountID) || thread.ID != threadID {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return nil
}

func mailboxEntApplyTombstoneLabelsMatch(ctx context.Context, tx *ent.Tx, accountID, messageID uuid.UUID, desired []MailboxSyncPolicyLabelState) error {
	if ctx == nil || tx == nil || accountID == uuid.Nil || messageID == uuid.Nil || desired == nil || len(desired) > MaxEnginePageItems {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	canonical, err := syncValidateLabelStates(accountID, desired)
	if err != nil || !reflect.DeepEqual(canonical, desired) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	rows, err := tx.MailMessageLabel.Query().Where(mailmessagelabel.MailMessageIDEQ(messageID)).Order(ent.Asc(mailmessagelabel.FieldID)).ForUpdate().All(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	if len(rows) != len(desired) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	byLabelID := make(map[uuid.UUID]*ent.MailMessageLabel, len(rows))
	labelIDSet := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || row.MailMessageID != messageID || row.MailLabelID == uuid.Nil || (!row.Manual && !row.EngineRule) || row.CreatedAt.IsZero() || row.UpdatedAt.IsZero() {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := byLabelID[row.MailLabelID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		byLabelID[row.MailLabelID] = row
		labelIDSet[row.MailLabelID] = struct{}{}
	}
	if len(labelIDSet) != 0 {
		labelIDs := mailboxEntApplySortedUUIDs(labelIDSet)
		labels, err := tx.MailLabel.Query().Where(maillabel.MailAccountIDEQ(accountID), maillabel.IDIn(labelIDs...), maillabel.DeletedAtIsNil()).Order(ent.Asc(maillabel.FieldID)).ForUpdate().All(ctx)
		if err != nil {
			return mailboxEntSafeError(err)
		}
		if len(labels) != len(labelIDs) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen := make(map[uuid.UUID]struct{}, len(labels))
		for _, label := range labels {
			if label == nil || label.ID == uuid.Nil || label.MailAccountID != accountID || label.DeletedAt != nil {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			if _, duplicate := seen[label.ID]; duplicate {
				return mailboxServiceError(MailboxServiceErrorInvalid)
			}
			seen[label.ID] = struct{}{}
		}
	}
	for _, state := range desired {
		link := byLabelID[state.Label.ID]
		if link == nil || link.Manual != state.Manual || link.EngineRule != state.EngineRule {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}
	return nil
}

func mailboxEntApplyEffectiveMessageTime(projection mailboxEntApplyMessageProjection, transactionTime time.Time) time.Time {
	if projection.sentAt == nil {
		if projection.receivedAt == nil {
			return transactionTime
		}
		return *projection.receivedAt
	}
	if projection.receivedAt == nil || projection.sentAt.After(*projection.receivedAt) {
		return *projection.sentAt
	}
	return *projection.receivedAt
}

func mailboxEntApplyOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	copied := value
	return &copied
}

func mailboxEntApplyTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxEntApplyCalendarMethodPointer(value *MailCalendarMethod) *mailmessage.CalendarMethod {
	if value == nil {
		return nil
	}
	copied := mailmessage.CalendarMethod(*value)
	return &copied
}

func mailboxEntApplyOptionalStringEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntApplyOptionalTimeEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func mailboxEntApplyCalendarMethodEqual(left, right *mailmessage.CalendarMethod) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntApplySortedUUIDs(values map[uuid.UUID]struct{}) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(left, right int) bool { return syncLessUUID(result[left], result[right]) })
	return result
}

func mailboxEntApplySortedStrings(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
