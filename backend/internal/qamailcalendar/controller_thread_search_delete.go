//go:build qa && !release

package qamailcalendar

import (
	"context"
	"errors"

	"entgo.io/ent/dialect/sql"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/maildomain"
	"github.com/valerii/hub/backend/ent/mailfolder"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailmessage"
	"github.com/valerii/hub/backend/ent/mailmessagefolder"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/user"
)

func (c *Controller) deleteFixture(ctx context.Context, nonce string) error {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return errThreadSearchFixtureControllerInput
	}

	scenario, err := c.storedFixtureScenario(nonce)
	if err != nil {
		return errThreadSearchFixtureLifecycle
	}

	switch scenario {
	case ScenarioInboundRequestV1:
		return c.deleteInboundRequestFixture(ctx, nonce)
	case ScenarioAccountMaterializationV1:
		return c.deleteAccountMaterializationFixture(ctx, nonce)
	case ScenarioThreadSearchV1:
		return c.deleteThreadSearchFixture(ctx, nonce)
	default:
		return errThreadSearchFixtureLifecycle
	}
}

func (c *Controller) deleteThreadSearchFixture(ctx context.Context, nonce string) error {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return errThreadSearchFixtureControllerInput
	}

	c.mu.Lock()
	record, found := c.threadSearchFixtures[nonce]
	if !found ||
		record.Nonce != nonce ||
		!threadSearchFixtureLifecycleRecordMatches(record, ThreadSearchFixtureLifecycleStateCreated) {
		c.mu.Unlock()
		return errThreadSearchFixtureLifecycle
	}
	deletingRecord := record
	deletingRecord.State = ThreadSearchFixtureLifecycleStateDeleting
	if current, stillFound := c.threadSearchFixtures[nonce]; !stillFound || current != record {
		c.mu.Unlock()
		return errThreadSearchFixtureLifecycle
	}
	c.threadSearchFixtures[nonce] = deletingRecord
	c.mu.Unlock()

	reconcileDelete := func(
		phase ThreadSearchCommitPhase,
		probe ThreadSearchFixtureFreshProbeSummary,
	) ThreadSearchFixtureCommitReconcileResult {
		c.mu.Lock()
		reconciled := ReconcileThreadSearchFixtureCommit(
			c.threadSearchFixtures,
			ThreadSearchFixtureCommitReconcileInput{
				Operation: ThreadSearchCommitOperationDelete,
				Phase:     phase,
				Probe:     probe,
				Expected:  deletingRecord,
			},
		)
		c.mu.Unlock()
		return reconciled
	}

	restoreCreatedAfterPreCommit := func() bool {
		reconciled := reconcileDelete(
			ThreadSearchCommitPhasePreCommit,
			ThreadSearchFixtureFreshProbeSummary{},
		)
		return reconciled.Status == 409 &&
			reconciled.Disposition == ThreadSearchFixtureLifecycleDispositionRestored &&
			reconciled.Applied
	}

	tx, err := c.client.Tx(ctx)
	if err != nil || tx == nil {
		if !restoreCreatedAfterPreCommit() {
			return errThreadSearchFixtureLifecycle
		}
		return errThreadSearchFixtureLifecycle
	}
	finished := false
	commitAttempted := false
	defer func() {
		if !finished && !commitAttempted {
			_ = tx.Rollback()
		}
	}()
	rollbackAndReconcile := func(allowGuardSentinel bool) error {
		rollbackErr := tx.Rollback()
		finished = true
		if rollbackErr != nil || !restoreCreatedAfterPreCommit() {
			return errThreadSearchFixtureLifecycle
		}
		if allowGuardSentinel {
			return errThreadSearchFixtureDeleteGuardDependent
		}
		return errThreadSearchFixtureLifecycle
	}

	messageFolders, err := tx.MailMessageFolder.Query().
		Where(
			mailmessagefolder.IDIn(deletingRecord.persisted.messageFolderIDs[:]...),
			mailmessagefolder.MailMessageIDIn(deletingRecord.persisted.messageIDs[:]...),
			mailmessagefolder.MailFolderIDEQ(deletingRecord.persisted.mailFolderID),
			threadSearchFixtureMailMessageFolderNoWait(),
		).
		All(ctx)
	if err != nil || len(messageFolders) != threadSearchFixtureRowCount {
		return rollbackAndReconcile(false)
	}

	messages, err := tx.MailMessage.Query().
		Where(
			mailmessage.IDIn(deletingRecord.persisted.messageIDs[:]...),
			mailmessage.MailAccountIDEQ(deletingRecord.persisted.mailAccountID),
			mailmessage.MailThreadIDIn(deletingRecord.persisted.threadIDs[:]...),
			threadSearchFixtureMailMessageNoWait(),
		).
		All(ctx)
	if err != nil || len(messages) != threadSearchFixtureRowCount {
		return rollbackAndReconcile(false)
	}

	threads, err := tx.MailThread.Query().
		Where(
			mailthread.IDIn(deletingRecord.persisted.threadIDs[:]...),
			mailthread.MailAccountIDEQ(deletingRecord.persisted.mailAccountID),
			threadSearchFixtureMailThreadNoWait(),
		).
		All(ctx)
	if err != nil || len(threads) != threadSearchFixtureRowCount {
		return rollbackAndReconcile(false)
	}

	folder, err := tx.MailFolder.Query().
		Where(
			mailfolder.IDEQ(deletingRecord.persisted.mailFolderID),
			mailfolder.MailAccountIDEQ(deletingRecord.persisted.mailAccountID),
			threadSearchFixtureMailFolderNoWait(),
		).
		Only(ctx)
	if err != nil {
		return rollbackAndReconcile(false)
	}

	account, err := tx.MailAccount.Query().
		Where(
			mailaccount.IDEQ(deletingRecord.persisted.mailAccountID),
			mailaccount.UserIDEQ(deletingRecord.persisted.ownerUserID),
			mailaccount.MailDomainIDEQ(deletingRecord.persisted.mailDomainID),
			mailaccount.OrganizationIDEQ(deletingRecord.persisted.organizationID),
			threadSearchFixtureMailAccountNoWait(),
		).
		Only(ctx)
	if err != nil {
		return rollbackAndReconcile(false)
	}

	identities, err := tx.MailIdentity.Query().
		Where(
			mailidentity.IDEQ(deletingRecord.persisted.mailIdentityID),
			mailidentity.MailAccountIDEQ(deletingRecord.persisted.mailAccountID),
			threadSearchFixtureMailIdentityNoWait(),
		).
		All(ctx)
	if err != nil || len(identities) != 1 {
		return rollbackAndReconcile(false)
	}
	identity := identities[0]
	extraIdentityExists, err := tx.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(deletingRecord.persisted.mailAccountID),
			mailidentity.IDNEQ(deletingRecord.persisted.mailIdentityID),
			threadSearchFixtureMailIdentityNoWait(),
		).
		Exist(ctx)
	if err != nil || extraIdentityExists {
		return rollbackAndReconcile(false)
	}

	domain, err := tx.MailDomain.Query().
		Where(
			maildomain.IDEQ(deletingRecord.persisted.mailDomainID),
			maildomain.OrganizationIDEQ(deletingRecord.persisted.organizationID),
			threadSearchFixtureMailDomainNoWait(),
		).
		Only(ctx)
	if err != nil {
		return rollbackAndReconcile(false)
	}

	owner, err := tx.User.Query().
		Where(
			user.IDEQ(deletingRecord.persisted.ownerUserID),
			threadSearchFixtureUserNoWait(),
		).
		Only(ctx)
	if err != nil {
		return rollbackAndReconcile(false)
	}

	organizationRow, err := tx.Organization.Query().
		Where(
			organization.IDEQ(deletingRecord.persisted.organizationID),
			organization.HasUsersWith(user.IDEQ(deletingRecord.persisted.ownerUserID)),
			threadSearchFixtureOrganizationNoWait(),
		).
		Only(ctx)
	if err != nil ||
		!threadSearchFixtureDeleteGraphValid(
			deletingRecord,
			organizationRow,
			owner,
			domain,
			account,
			identity,
			folder,
			threads,
			messages,
			messageFolders,
		) {
		return rollbackAndReconcile(false)
	}
	if err := deleteFixtureAuthSideEffectsLocked(ctx, tx, deletingRecord.persisted.ownerUserID); err != nil {
		return rollbackAndReconcile(false)
	}
	if guardErr := threadSearchFixtureGuardDeleteDependents(ctx, tx, deletingRecord); guardErr != nil {
		return rollbackAndReconcile(errors.Is(guardErr, errThreadSearchFixtureDeleteGuardDependent))
	}

	for _, id := range deletingRecord.persisted.messageFolderIDs {
		if err := tx.MailMessageFolder.DeleteOneID(id).Exec(ctx); err != nil {
			return rollbackAndReconcile(false)
		}
	}
	for _, id := range deletingRecord.persisted.messageIDs {
		if err := tx.MailMessage.DeleteOneID(id).Exec(ctx); err != nil {
			return rollbackAndReconcile(false)
		}
	}
	for _, id := range deletingRecord.persisted.threadIDs {
		if err := tx.MailThread.DeleteOneID(id).Exec(ctx); err != nil {
			return rollbackAndReconcile(false)
		}
	}
	if err := tx.MailFolder.DeleteOneID(deletingRecord.persisted.mailFolderID).Exec(ctx); err != nil {
		return rollbackAndReconcile(false)
	}
	if err := tx.MailIdentity.DeleteOneID(deletingRecord.persisted.mailIdentityID).Exec(ctx); err != nil {
		return rollbackAndReconcile(false)
	}
	if err := tx.MailAccount.DeleteOneID(deletingRecord.persisted.mailAccountID).Exec(ctx); err != nil {
		return rollbackAndReconcile(false)
	}
	if err := tx.MailDomain.DeleteOneID(deletingRecord.persisted.mailDomainID).Exec(ctx); err != nil {
		return rollbackAndReconcile(false)
	}
	if err := tx.User.DeleteOneID(deletingRecord.persisted.ownerUserID).Exec(ctx); err != nil {
		return rollbackAndReconcile(false)
	}
	if err := tx.Organization.DeleteOneID(deletingRecord.persisted.organizationID).Exec(ctx); err != nil {
		return rollbackAndReconcile(false)
	}

	commitAttempted = true
	if err := tx.Commit(); err != nil {
		finished = true
		reconciled := reconcileDelete(
			ThreadSearchCommitPhaseOutcomeUnknown,
			c.threadSearchFixtureProbeManifest(ctx, deletingRecord.plan, deletingRecord.persisted),
		)
		if reconciled.Status == 204 &&
			reconciled.Disposition == ThreadSearchFixtureLifecycleDispositionRemoved {
			return nil
		}
		return errThreadSearchFixtureLifecycle
	}

	finished = true
	reconciled := reconcileDelete(
		ThreadSearchCommitPhaseSucceeded,
		ThreadSearchFixtureFreshProbeSummary{},
	)
	if reconciled.Status == 204 &&
		reconciled.Disposition == ThreadSearchFixtureLifecycleDispositionRemoved {
		return nil
	}
	return errThreadSearchFixtureLifecycle
}

func threadSearchFixtureDeleteGraphValid(
	record threadSearchFixtureRecord,
	organizationRow *ent.Organization,
	owner *ent.User,
	domain *ent.MailDomain,
	account *ent.MailAccount,
	identity *ent.MailIdentity,
	folder *ent.MailFolder,
	threads []*ent.MailThread,
	messages []*ent.MailMessage,
	messageFolders []*ent.MailMessageFolder,
) bool {
	if record.Nonce != record.plan.nonce ||
		record.Scenario != ThreadSearchFixtureLifecycleScenarioThreadSearchV1 ||
		!threadSearchFixtureLifecycleRecordMatches(record, ThreadSearchFixtureLifecycleStateDeleting) ||
		!threadSearchFixtureOrganizationMatchesPlan(organizationRow, record.plan) ||
		!threadSearchFixtureOwnerMatchesPlan(owner, record.plan) ||
		!threadSearchFixtureDomainMatchesPlan(domain, record.plan) ||
		!threadSearchFixtureAccountMatchesPlan(account, record.plan) ||
		!threadSearchFixtureMailIdentityMatchesPlan(identity, record.plan) ||
		!threadSearchFixtureFolderMatchesPlan(folder, record.plan) ||
		len(threads) != threadSearchFixtureRowCount ||
		len(messages) != threadSearchFixtureRowCount ||
		len(messageFolders) != threadSearchFixtureRowCount {
		return false
	}

	threadByID := make(map[uuid.UUID]*ent.MailThread, threadSearchFixtureRowCount)
	for _, thread := range threads {
		if thread == nil {
			return false
		}
		if _, exists := threadByID[thread.ID]; exists {
			return false
		}
		threadByID[thread.ID] = thread
	}
	messageByID := make(map[uuid.UUID]*ent.MailMessage, threadSearchFixtureRowCount)
	for _, message := range messages {
		if message == nil {
			return false
		}
		if _, exists := messageByID[message.ID]; exists {
			return false
		}
		messageByID[message.ID] = message
	}
	messageFolderByID := make(map[uuid.UUID]*ent.MailMessageFolder, threadSearchFixtureRowCount)
	for _, messageFolder := range messageFolders {
		if messageFolder == nil {
			return false
		}
		if _, exists := messageFolderByID[messageFolder.ID]; exists {
			return false
		}
		messageFolderByID[messageFolder.ID] = messageFolder
	}

	for index, row := range record.plan.rows {
		if record.persisted.threadIDs[index] != row.threadID ||
			record.persisted.messageIDs[index] != row.messageID ||
			record.persisted.messageFolderIDs[index] != row.messageFolderID ||
			!threadSearchFixtureThreadMatchesPlan(threadByID[row.threadID], record.plan, row) ||
			!threadSearchFixtureMessageMatchesPlan(messageByID[row.messageID], record.plan, row) ||
			!threadSearchFixtureMessageFolderMatchesPlan(messageFolderByID[row.messageFolderID], record.plan, row) {
			return false
		}
	}
	return true
}

func threadSearchFixtureMailMessageFolderNoWait() predicate.MailMessageFolder {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureMailMessageNoWait() predicate.MailMessage {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureMailThreadNoWait() predicate.MailThread {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureMailFolderNoWait() predicate.MailFolder {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureMailAccountNoWait() predicate.MailAccount {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureMailIdentityNoWait() predicate.MailIdentity {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureMailDomainNoWait() predicate.MailDomain {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureUserNoWait() predicate.User {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func threadSearchFixtureOrganizationNoWait() predicate.Organization {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
