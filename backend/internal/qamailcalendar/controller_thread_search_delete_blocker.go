//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
	"github.com/valerii/hub/backend/ent/calendarmutationreceipt"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

var (
	errThreadSearchDeleteBlockerInput     = errors.New("qa mail thread search delete blocker input is invalid")
	errThreadSearchDeleteBlockerLifecycle = errors.New("qa mail thread search delete blocker lifecycle failed")
)

type threadSearchDeleteBlockerKind string

const (
	threadSearchDeleteBlockerKindAIProvider threadSearchDeleteBlockerKind = "atlas-user-set-null-v1"
	threadSearchDeleteBlockerKindReceipt    threadSearchDeleteBlockerKind = "ent-user-cascade-v1"
	threadSearchDeleteBlockerKindByMessage  threadSearchDeleteBlockerKind = "mail-message-folder-by-message-extra-v1"
	threadSearchDeleteBlockerKindByFolder   threadSearchDeleteBlockerKind = "mail-message-folder-by-folder-extra-v1"
)

type threadSearchDeleteBlockerAIProviderPlan struct {
	id      uuid.UUID
	ownerID uuid.UUID
	nonce   string
	fixedAt time.Time
}

type threadSearchDeleteBlockerReceiptPlan struct {
	id                 uuid.UUID
	organizationID     uuid.UUID
	ownerID            uuid.UUID
	idempotencyKeyHash [sha256.Size]byte
	requestHash        [sha256.Size]byte
	fixedAt            time.Time
	expiresAt          time.Time
}

type threadSearchDeleteBlockerPeerPlan struct {
	id          uuid.UUID
	kind        threadSearchDeleteBlockerKind
	messageID   uuid.UUID
	folderID    uuid.UUID
	projectedAt time.Time
	fixedAt     time.Time
}

type threadSearchDeleteBlockerRootKind uint8

const (
	threadSearchDeleteBlockerRootOrganization threadSearchDeleteBlockerRootKind = iota
	threadSearchDeleteBlockerRootOwner
	threadSearchDeleteBlockerRootAccount
)

type threadSearchDeleteBlockerRoot struct {
	id   uuid.UUID
	kind threadSearchDeleteBlockerRootKind
}

type threadSearchDeleteBlockerPeerParentKind uint8

const (
	threadSearchDeleteBlockerPeerParentMessage threadSearchDeleteBlockerPeerParentKind = iota
	threadSearchDeleteBlockerPeerParentFolder
)

type threadSearchDeleteBlockerPeerParent struct {
	id     uuid.UUID
	kind   threadSearchDeleteBlockerPeerParentKind
	plan   threadSearchFixturePlan
	row    threadSearchFixtureRow
}

func (c *Controller) armThreadSearchDeleteBlocker(ctx context.Context, nonce, kind string) error {
	if !c.threadSearchDeleteBlockerControllerValid(ctx) || !ValidNonce(nonce) {
		return errThreadSearchDeleteBlockerInput
	}
	blockerKind, ok := threadSearchDeleteBlockerSingleKind(kind)
	if !ok {
		return errThreadSearchDeleteBlockerInput
	}

	c.mu.Lock()
	record, ok := c.threadSearchDeleteBlockerCreatedRecordLocked(nonce)
	c.mu.Unlock()
	if !ok {
		return errThreadSearchDeleteBlockerLifecycle
	}

	switch blockerKind {
	case threadSearchDeleteBlockerKindAIProvider:
		return c.armThreadSearchDeleteBlockerAIProvider(ctx, record.plan)
	case threadSearchDeleteBlockerKindReceipt:
		return c.armThreadSearchDeleteBlockerReceipt(ctx, record.plan)
	default:
		return errThreadSearchDeleteBlockerLifecycle
	}
}

func (c *Controller) disarmThreadSearchDeleteBlocker(ctx context.Context, nonce, kind string) error {
	if !c.threadSearchDeleteBlockerControllerValid(ctx) || !ValidNonce(nonce) {
		return errThreadSearchDeleteBlockerInput
	}
	blockerKind, ok := threadSearchDeleteBlockerSingleKind(kind)
	if !ok {
		return errThreadSearchDeleteBlockerInput
	}

	c.mu.Lock()
	record, ok := c.threadSearchDeleteBlockerCreatedRecordLocked(nonce)
	if !ok {
		c.mu.Unlock()
		return errThreadSearchDeleteBlockerLifecycle
	}

	switch blockerKind {
	case threadSearchDeleteBlockerKindAIProvider:
		plan, valid := newThreadSearchDeleteBlockerAIProviderPlan(record.plan)
		c.mu.Unlock()
		if !valid {
			return errThreadSearchDeleteBlockerLifecycle
		}
		return c.disarmThreadSearchDeleteBlockerAIProvider(ctx, plan)
	case threadSearchDeleteBlockerKindReceipt:
		plan, valid := newThreadSearchDeleteBlockerReceiptPlan(record.plan)
		c.mu.Unlock()
		if !valid {
			return errThreadSearchDeleteBlockerLifecycle
		}
		return c.disarmThreadSearchDeleteBlockerReceipt(ctx, plan)
	default:
		c.mu.Unlock()
		return errThreadSearchDeleteBlockerLifecycle
	}
}

func (c *Controller) armThreadSearchDeletePeerBlocker(ctx context.Context, nonce, peerNonce, kind string) error {
	if !c.threadSearchDeleteBlockerControllerValid(ctx) ||
		!ValidNonce(nonce) ||
		!ValidNonce(peerNonce) ||
		nonce == peerNonce {
		return errThreadSearchDeleteBlockerInput
	}
	blockerKind, ok := threadSearchDeleteBlockerPeerKind(kind)
	if !ok {
		return errThreadSearchDeleteBlockerInput
	}

	c.mu.Lock()
	targetRecord, ok := c.threadSearchDeleteBlockerCreatedRecordLocked(nonce)
	if !ok {
		c.mu.Unlock()
		return errThreadSearchDeleteBlockerLifecycle
	}
	peerRecord, ok := c.threadSearchDeleteBlockerCreatedRecordLocked(peerNonce)
	if !ok {
		c.mu.Unlock()
		return errThreadSearchDeleteBlockerLifecycle
	}
	c.mu.Unlock()
	plan, ok := newThreadSearchDeleteBlockerPeerPlan(blockerKind, targetRecord.plan, peerRecord.plan)
	if !ok {
		return errThreadSearchDeleteBlockerLifecycle
	}
	return c.armThreadSearchDeleteBlockerPeer(ctx, plan, targetRecord.plan, peerRecord.plan)
}

func (c *Controller) disarmThreadSearchDeletePeerBlocker(ctx context.Context, nonce, peerNonce, kind string) error {
	if !c.threadSearchDeleteBlockerControllerValid(ctx) ||
		!ValidNonce(nonce) ||
		!ValidNonce(peerNonce) ||
		nonce == peerNonce {
		return errThreadSearchDeleteBlockerInput
	}
	blockerKind, ok := threadSearchDeleteBlockerPeerKind(kind)
	if !ok {
		return errThreadSearchDeleteBlockerInput
	}

	c.mu.Lock()
	targetRecord, ok := c.threadSearchDeleteBlockerCreatedRecordLocked(nonce)
	if !ok {
		c.mu.Unlock()
		return errThreadSearchDeleteBlockerLifecycle
	}
	peerRecord, ok := c.threadSearchDeleteBlockerCreatedRecordLocked(peerNonce)
	if !ok {
		c.mu.Unlock()
		return errThreadSearchDeleteBlockerLifecycle
	}
	plan, ok := newThreadSearchDeleteBlockerPeerPlan(blockerKind, targetRecord.plan, peerRecord.plan)
	c.mu.Unlock()
	if !ok {
		return errThreadSearchDeleteBlockerLifecycle
	}

	return c.disarmThreadSearchDeleteBlockerPeer(ctx, plan)
}

func (c *Controller) threadSearchDeleteBlockerControllerValid(ctx context.Context) bool {
	return c != nil &&
		ctx != nil &&
		c.client != nil &&
		c.fixtures != nil &&
		c.accountMaterializationFixtures != nil &&
		c.threadSearchFixtures != nil &&
		validToken(string(c.config.token))
}

func (c *Controller) threadSearchDeleteBlockerCreatedRecordLocked(nonce string) (threadSearchFixtureRecord, bool) {
	record, found := c.threadSearchFixtures[nonce]
	if !found ||
		record.Nonce != nonce ||
		record.Scenario != ThreadSearchFixtureLifecycleScenarioThreadSearchV1 ||
		!threadSearchFixtureLifecycleRecordMatches(record, ThreadSearchFixtureLifecycleStateCreated) {
		return threadSearchFixtureRecord{}, false
	}
	return record, true
}

func threadSearchDeleteBlockerSingleKind(value string) (threadSearchDeleteBlockerKind, bool) {
	switch threadSearchDeleteBlockerKind(value) {
	case threadSearchDeleteBlockerKindAIProvider:
		return threadSearchDeleteBlockerKindAIProvider, true
	case threadSearchDeleteBlockerKindReceipt:
		return threadSearchDeleteBlockerKindReceipt, true
	default:
		return "", false
	}
}

func threadSearchDeleteBlockerPeerKind(value string) (threadSearchDeleteBlockerKind, bool) {
	switch threadSearchDeleteBlockerKind(value) {
	case threadSearchDeleteBlockerKindByMessage:
		return threadSearchDeleteBlockerKindByMessage, true
	case threadSearchDeleteBlockerKindByFolder:
		return threadSearchDeleteBlockerKindByFolder, true
	default:
		return "", false
	}
}

func newThreadSearchDeleteBlockerAIProviderPlan(plan threadSearchFixturePlan) (threadSearchDeleteBlockerAIProviderPlan, bool) {
	if !threadSearchFixturePlanMatchesCanonical(plan) {
		return threadSearchDeleteBlockerAIProviderPlan{}, false
	}
	return threadSearchDeleteBlockerAIProviderPlan{
		id:      threadSearchFixtureUUID("delete-blocker-ai-provider", plan.nonce),
		ownerID: plan.ownerUserID,
		nonce:   plan.nonce,
		fixedAt: plan.fixedAt,
	}, true
}

func newThreadSearchDeleteBlockerReceiptPlan(plan threadSearchFixturePlan) (threadSearchDeleteBlockerReceiptPlan, bool) {
	if !threadSearchFixturePlanMatchesCanonical(plan) {
		return threadSearchDeleteBlockerReceiptPlan{}, false
	}
	return threadSearchDeleteBlockerReceiptPlan{
		id:                 threadSearchFixtureUUID("delete-blocker-calendar-receipt", plan.nonce),
		organizationID:     plan.organizationID,
		ownerID:            plan.ownerUserID,
		idempotencyKeyHash: threadSearchFixtureHash("delete-blocker-calendar-receipt-idempotency", plan.nonce),
		requestHash:        threadSearchFixtureHash("delete-blocker-calendar-receipt-request", plan.nonce),
		fixedAt:            plan.fixedAt,
		expiresAt:          plan.fixedAt.Add(24 * time.Hour),
	}, true
}

func newThreadSearchDeleteBlockerPeerPlan(
	kind threadSearchDeleteBlockerKind,
	targetPlan, peerPlan threadSearchFixturePlan,
) (threadSearchDeleteBlockerPeerPlan, bool) {
	if !threadSearchFixturePlanMatchesCanonical(targetPlan) ||
		!threadSearchFixturePlanMatchesCanonical(peerPlan) ||
		targetPlan.nonce == peerPlan.nonce {
		return threadSearchDeleteBlockerPeerPlan{}, false
	}

	pairMaterial := string(kind) + "\x00" + targetPlan.nonce + "\x00" + peerPlan.nonce
	plan := threadSearchDeleteBlockerPeerPlan{
		id:          threadSearchFixtureUUID("delete-blocker-mail-message-folder-"+string(kind), pairMaterial),
		kind:        kind,
		projectedAt: threadSearchDeleteBlockerPeerTime("delete-blocker-mail-message-folder-projected-at", pairMaterial, targetPlan.fixedAt),
		fixedAt:     threadSearchDeleteBlockerPeerTime("delete-blocker-mail-message-folder-fixed-at", pairMaterial, targetPlan.fixedAt),
	}
	switch kind {
	case threadSearchDeleteBlockerKindByMessage:
		plan.messageID = targetPlan.rows[0].messageID
		plan.folderID = peerPlan.mailFolderID
	case threadSearchDeleteBlockerKindByFolder:
		plan.messageID = peerPlan.rows[0].messageID
		plan.folderID = targetPlan.mailFolderID
	default:
		return threadSearchDeleteBlockerPeerPlan{}, false
	}
	return plan, plan.messageID != uuid.Nil && plan.folderID != uuid.Nil
}

func threadSearchDeleteBlockerPeerTime(label, pairMaterial string, base time.Time) time.Time {
	digest := threadSearchFixtureHash(label, pairMaterial)
	seconds := binary.BigEndian.Uint32(digest[:4]) % uint32((24 * time.Hour) / time.Second)
	return base.Add(time.Duration(seconds) * time.Second)
}

func (c *Controller) armThreadSearchDeleteBlockerAIProvider(ctx context.Context, fixturePlan threadSearchFixturePlan) error {
	plan, ok := newThreadSearchDeleteBlockerAIProviderPlan(fixturePlan)
	if !ok {
		return errThreadSearchDeleteBlockerLifecycle
	}
	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted := false
	defer func() {
		if !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	if err := threadSearchDeleteBlockerLockRoots(ctx, tx, fixturePlan); err != nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	exists, err := tx.AIProvider.Query().Where(aiprovider.IDEQ(plan.id)).Exist(ctx)
	if err != nil || exists {
		return errThreadSearchDeleteBlockerLifecycle
	}
	provider, err := tx.AIProvider.Create().
		SetID(plan.id).
		SetName("QA Thread Search Delete Blocker " + plan.nonce).
		SetProviderType(aiprovider.ProviderType("ollama")).
		SetBaseURL("https://qa-" + plan.nonce + ".invalid").
		SetModels([]string{"qa-thread-search-delete-blocker-" + plan.nonce}).
		SetEnabled(true).
		SetPriority(0).
		SetMaxRpm(1).
		SetRuResident(true).
		SetScope(aiprovider.Scope("user")).
		SetOwnerID(plan.ownerID).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchDeleteBlockerAIProviderMatches(ctx, provider, plan, false) {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.Commit(); err != nil {
		commitAttempted = true
		present, probeErr := threadSearchDeleteBlockerAIProviderPresent(ctx, c.client, plan, false)
		if probeErr == nil && present {
			return nil
		}
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted = true
	return nil
}

func (c *Controller) disarmThreadSearchDeleteBlockerAIProvider(
	ctx context.Context,
	plan threadSearchDeleteBlockerAIProviderPlan,
) error {
	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted := false
	defer func() {
		if !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	provider, err := tx.AIProvider.Query().Where(
		aiprovider.IDEQ(plan.id),
		threadSearchDeleteBlockerAIProviderNoWait(),
	).Only(ctx)
	if err != nil || !threadSearchDeleteBlockerAIProviderMatches(ctx, provider, plan, true) {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.AIProvider.DeleteOneID(plan.id).Exec(ctx); err != nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.Commit(); err != nil {
		commitAttempted = true
		present, probeErr := threadSearchDeleteBlockerAIProviderPresent(ctx, c.client, plan, true)
		if probeErr == nil && !present {
			return nil
		}
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted = true
	return nil
}

func (c *Controller) armThreadSearchDeleteBlockerReceipt(ctx context.Context, fixturePlan threadSearchFixturePlan) error {
	plan, ok := newThreadSearchDeleteBlockerReceiptPlan(fixturePlan)
	if !ok {
		return errThreadSearchDeleteBlockerLifecycle
	}
	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted := false
	defer func() {
		if !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	if err := threadSearchDeleteBlockerLockRoots(ctx, tx, fixturePlan); err != nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	exists, err := tx.CalendarMutationReceipt.Query().Where(calendarmutationreceipt.IDEQ(plan.id)).Exist(ctx)
	if err != nil || exists {
		return errThreadSearchDeleteBlockerLifecycle
	}
	receipt, err := tx.CalendarMutationReceipt.Create().
		SetID(plan.id).
		SetOrganizationID(plan.organizationID).
		SetActorUserID(plan.ownerID).
		SetOperation(calendarmutationreceipt.Operation("event_create")).
		SetIdempotencyKeyHash(append([]byte(nil), plan.idempotencyKeyHash[:]...)).
		SetRequestHash(append([]byte(nil), plan.requestHash[:]...)).
		SetState(calendarmutationreceipt.State("completed")).
		SetExpiresAt(plan.expiresAt).
		SetResponseStatus(204).
		SetResponseJSON([]byte("{} ")).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchDeleteBlockerReceiptMatches(receipt, plan) {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.Commit(); err != nil {
		commitAttempted = true
		present, probeErr := threadSearchDeleteBlockerReceiptPresent(ctx, c.client, plan)
		if probeErr == nil && present {
			return nil
		}
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted = true
	return nil
}

func (c *Controller) disarmThreadSearchDeleteBlockerReceipt(
	ctx context.Context,
	plan threadSearchDeleteBlockerReceiptPlan,
) error {
	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted := false
	defer func() {
		if !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	receipt, err := tx.CalendarMutationReceipt.Query().Where(
		calendarmutationreceipt.IDEQ(plan.id),
		threadSearchDeleteBlockerCalendarMutationReceiptNoWait(),
	).Only(ctx)
	if err != nil || !threadSearchDeleteBlockerReceiptMatches(receipt, plan) {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.CalendarMutationReceipt.DeleteOneID(plan.id).Exec(ctx); err != nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.Commit(); err != nil {
		commitAttempted = true
		present, probeErr := threadSearchDeleteBlockerReceiptPresent(ctx, c.client, plan)
		if probeErr == nil && !present {
			return nil
		}
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted = true
	return nil
}

func (c *Controller) armThreadSearchDeleteBlockerPeer(
	ctx context.Context,
	plan threadSearchDeleteBlockerPeerPlan,
	targetPlan, peerPlan threadSearchFixturePlan,
) error {
	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted := false
	defer func() {
		if !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	if err := threadSearchDeleteBlockerLockPeerParents(ctx, tx, plan, targetPlan, peerPlan); err != nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	exists, err := tx.MailMessageFolder.Query().Where(mailmessagefolder.IDEQ(plan.id)).Exist(ctx)
	if err != nil || exists {
		return errThreadSearchDeleteBlockerLifecycle
	}
	relation, err := tx.MailMessageFolder.Create().
		SetID(plan.id).
		SetMailMessageID(plan.messageID).
		SetMailFolderID(plan.folderID).
		SetLastSeenSyncEpoch(0).
		SetProjectedAt(plan.projectedAt).
		SetCreatedAt(plan.fixedAt).
		SetUpdatedAt(plan.fixedAt).
		Save(ctx)
	if err != nil || !threadSearchDeleteBlockerPeerMatches(relation, plan) {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.Commit(); err != nil {
		commitAttempted = true
		present, probeErr := threadSearchDeleteBlockerPeerPresent(ctx, c.client, plan)
		if probeErr == nil && present {
			return nil
		}
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted = true
	return nil
}

func (c *Controller) disarmThreadSearchDeleteBlockerPeer(ctx context.Context, plan threadSearchDeleteBlockerPeerPlan) error {
	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted := false
	defer func() {
		if !commitAttempted {
			_ = tx.Rollback()
		}
	}()

	relation, err := tx.MailMessageFolder.Query().Where(
		mailmessagefolder.IDEQ(plan.id),
		threadSearchFixtureMailMessageFolderNoWait(),
	).Only(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	if err != nil || !threadSearchDeleteBlockerPeerMatches(relation, plan) {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.MailMessageFolder.DeleteOneID(plan.id).Exec(ctx); err != nil {
		return errThreadSearchDeleteBlockerLifecycle
	}
	if err := tx.Commit(); err != nil {
		commitAttempted = true
		present, probeErr := threadSearchDeleteBlockerPeerPresent(ctx, c.client, plan)
		if probeErr == nil && !present {
			return nil
		}
		return errThreadSearchDeleteBlockerLifecycle
	}
	commitAttempted = true
	return nil
}

func threadSearchDeleteBlockerLockRoots(ctx context.Context, tx *ent.Tx, plan threadSearchFixturePlan) error {
	if ctx == nil || tx == nil || !threadSearchFixturePlanMatchesCanonical(plan) {
		return errThreadSearchDeleteBlockerLifecycle
	}
	roots := []threadSearchDeleteBlockerRoot{
		{id: plan.organizationID, kind: threadSearchDeleteBlockerRootOrganization},
		{id: plan.ownerUserID, kind: threadSearchDeleteBlockerRootOwner},
		{id: plan.mailAccountID, kind: threadSearchDeleteBlockerRootAccount},
	}
	sort.Slice(roots, func(left, right int) bool {
		return bytes.Compare(roots[left].id[:], roots[right].id[:]) < 0
	})
	for index, root := range roots {
		if root.id == uuid.Nil || (index > 0 && root.id == roots[index-1].id) {
			return errThreadSearchDeleteBlockerLifecycle
		}
		switch root.kind {
		case threadSearchDeleteBlockerRootOrganization:
			value, err := tx.Organization.Query().Where(
				organization.IDEQ(plan.organizationID),
				organization.HasUsersWith(user.IDEQ(plan.ownerUserID)),
				threadSearchFixtureOrganizationNoWait(),
			).Only(ctx)
			if err != nil || !threadSearchFixtureOrganizationMatchesPlan(value, plan) {
				return errThreadSearchDeleteBlockerLifecycle
			}
		case threadSearchDeleteBlockerRootOwner:
			value, err := tx.User.Query().Where(
				user.IDEQ(plan.ownerUserID),
				threadSearchFixtureUserNoWait(),
			).Only(ctx)
			if err != nil || !threadSearchFixtureOwnerMatchesPlan(value, plan) {
				return errThreadSearchDeleteBlockerLifecycle
			}
		case threadSearchDeleteBlockerRootAccount:
			value, err := tx.MailAccount.Query().Where(
				mailaccount.IDEQ(plan.mailAccountID),
				mailaccount.UserIDEQ(plan.ownerUserID),
				mailaccount.OrganizationIDEQ(plan.organizationID),
				mailaccount.MailDomainIDEQ(plan.mailDomainID),
				threadSearchFixtureMailAccountNoWait(),
			).Only(ctx)
			if err != nil || !threadSearchFixtureAccountMatchesPlan(value, plan) {
				return errThreadSearchDeleteBlockerLifecycle
			}
		default:
			return errThreadSearchDeleteBlockerLifecycle
		}
	}
	return nil
}

func threadSearchDeleteBlockerLockPeerParents(
	ctx context.Context,
	tx *ent.Tx,
	blockerPlan threadSearchDeleteBlockerPeerPlan,
	targetPlan, peerPlan threadSearchFixturePlan,
) error {
	if ctx == nil || tx == nil ||
		!threadSearchFixturePlanMatchesCanonical(targetPlan) ||
		!threadSearchFixturePlanMatchesCanonical(peerPlan) ||
		targetPlan.nonce == peerPlan.nonce {
		return errThreadSearchDeleteBlockerLifecycle
	}

	var parents []threadSearchDeleteBlockerPeerParent
	switch blockerPlan.kind {
	case threadSearchDeleteBlockerKindByMessage:
		parents = []threadSearchDeleteBlockerPeerParent{
			{id: targetPlan.rows[0].messageID, kind: threadSearchDeleteBlockerPeerParentMessage, plan: targetPlan, row: targetPlan.rows[0]},
			{id: peerPlan.mailFolderID, kind: threadSearchDeleteBlockerPeerParentFolder, plan: peerPlan},
		}
	case threadSearchDeleteBlockerKindByFolder:
		parents = []threadSearchDeleteBlockerPeerParent{
			{id: peerPlan.rows[0].messageID, kind: threadSearchDeleteBlockerPeerParentMessage, plan: peerPlan, row: peerPlan.rows[0]},
			{id: targetPlan.mailFolderID, kind: threadSearchDeleteBlockerPeerParentFolder, plan: targetPlan},
		}
	default:
		return errThreadSearchDeleteBlockerLifecycle
	}
	sort.Slice(parents, func(left, right int) bool {
		return bytes.Compare(parents[left].id[:], parents[right].id[:]) < 0
	})
	for index, parent := range parents {
		if parent.id == uuid.Nil || (index > 0 && parent.id == parents[index-1].id) {
			return errThreadSearchDeleteBlockerLifecycle
		}
		switch parent.kind {
		case threadSearchDeleteBlockerPeerParentMessage:
			value, err := tx.MailMessage.Query().Where(
				mailmessage.IDEQ(parent.row.messageID),
				mailmessage.MailAccountIDEQ(parent.plan.mailAccountID),
				mailmessage.MailThreadIDEQ(parent.row.threadID),
				threadSearchFixtureMailMessageNoWait(),
			).Only(ctx)
			if err != nil || !threadSearchFixtureMessageMatchesPlan(value, parent.plan, parent.row) {
				return errThreadSearchDeleteBlockerLifecycle
			}
		case threadSearchDeleteBlockerPeerParentFolder:
			value, err := tx.MailFolder.Query().Where(
				mailfolder.IDEQ(parent.plan.mailFolderID),
				mailfolder.MailAccountIDEQ(parent.plan.mailAccountID),
				threadSearchFixtureMailFolderNoWait(),
			).Only(ctx)
			if err != nil || !threadSearchFixtureFolderMatchesPlan(value, parent.plan) {
				return errThreadSearchDeleteBlockerLifecycle
			}
		default:
			return errThreadSearchDeleteBlockerLifecycle
		}
	}
	return nil
}

func threadSearchDeleteBlockerAIProviderMatches(
	ctx context.Context,
	value *ent.AIProvider,
	plan threadSearchDeleteBlockerAIProviderPlan,
	allowMissingOwner bool,
) bool {
	if ctx == nil || value == nil ||
		value.ID != plan.id ||
		value.Name != "QA Thread Search Delete Blocker "+plan.nonce ||
		value.ProviderType != aiprovider.ProviderType("ollama") ||
		value.BaseURL != "https://qa-"+plan.nonce+".invalid" ||
		len(value.Models) != 1 || value.Models[0] != "qa-thread-search-delete-blocker-"+plan.nonce ||
		!value.Enabled ||
		value.Priority != 0 ||
		value.MaxRpm != 1 ||
		!value.RuResident ||
		value.Scope != aiprovider.Scope("user") ||
		!value.CreatedAt.Equal(plan.fixedAt) ||
		!value.UpdatedAt.Equal(plan.fixedAt) {
		return false
	}
	owner, err := value.QueryOwner().Only(ctx)
	if ent.IsNotFound(err) {
		return allowMissingOwner
	}
	return err == nil && owner != nil && owner.ID == plan.ownerID
}

func threadSearchDeleteBlockerReceiptMatches(
	value *ent.CalendarMutationReceipt,
	plan threadSearchDeleteBlockerReceiptPlan,
) bool {
	return value != nil &&
		value.ID == plan.id &&
		value.OrganizationID == plan.organizationID &&
		value.ActorUserID == plan.ownerID &&
		value.Operation == calendarmutationreceipt.Operation("event_create") &&
		bytes.Equal(value.IdempotencyKeyHash, plan.idempotencyKeyHash[:]) &&
		bytes.Equal(value.RequestHash, plan.requestHash[:]) &&
		value.State == calendarmutationreceipt.State("completed") &&
		value.ResponseStatus != nil && *value.ResponseStatus == 204 &&
		value.ResponseJSON != nil && bytes.Equal(*value.ResponseJSON, []byte("{} ")) &&
		value.CalendarEventID == nil &&
		value.ExpiresAt.Equal(plan.expiresAt) &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchDeleteBlockerPeerMatches(
	value *ent.MailMessageFolder,
	plan threadSearchDeleteBlockerPeerPlan,
) bool {
	return value != nil &&
		value.ID == plan.id &&
		value.MailMessageID == plan.messageID &&
		value.MailFolderID == plan.folderID &&
		value.EngineState == nil &&
		value.LastSeenSyncEpoch == 0 &&
		value.ProjectedAt.Equal(plan.projectedAt) &&
		value.CreatedAt.Equal(plan.fixedAt) &&
		value.UpdatedAt.Equal(plan.fixedAt)
}

func threadSearchDeleteBlockerAIProviderPresent(
	ctx context.Context,
	client *ent.Client,
	plan threadSearchDeleteBlockerAIProviderPlan,
	allowMissingOwner bool,
) (bool, error) {
	if ctx == nil || client == nil {
		return false, errThreadSearchDeleteBlockerLifecycle
	}
	value, err := client.AIProvider.Query().Where(aiprovider.IDEQ(plan.id)).Only(ctx)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil || !threadSearchDeleteBlockerAIProviderMatches(ctx, value, plan, allowMissingOwner) {
		return false, errThreadSearchDeleteBlockerLifecycle
	}
	return true, nil
}

func threadSearchDeleteBlockerReceiptPresent(
	ctx context.Context,
	client *ent.Client,
	plan threadSearchDeleteBlockerReceiptPlan,
) (bool, error) {
	if ctx == nil || client == nil {
		return false, errThreadSearchDeleteBlockerLifecycle
	}
	value, err := client.CalendarMutationReceipt.Query().Where(calendarmutationreceipt.IDEQ(plan.id)).Only(ctx)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil || !threadSearchDeleteBlockerReceiptMatches(value, plan) {
		return false, errThreadSearchDeleteBlockerLifecycle
	}
	return true, nil
}

func threadSearchDeleteBlockerPeerPresent(
	ctx context.Context,
	client *ent.Client,
	plan threadSearchDeleteBlockerPeerPlan,
) (bool, error) {
	if ctx == nil || client == nil {
		return false, errThreadSearchDeleteBlockerLifecycle
	}
	value, err := client.MailMessageFolder.Query().Where(mailmessagefolder.IDEQ(plan.id)).Only(ctx)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil || !threadSearchDeleteBlockerPeerMatches(value, plan) {
		return false, errThreadSearchDeleteBlockerLifecycle
	}
	return true, nil
}

func threadSearchDeleteBlockerAIProviderNoWait() predicate.AIProvider {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchDeleteBlockerCalendarMutationReceiptNoWait() predicate.CalendarMutationReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
