//go:build qa && !release

package qamailcalendar

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/maildraft"
	"github.com/valerii/hub/backend/ent/maillabel"
	"github.com/valerii/hub/backend/ent/mailidentity"
	"github.com/valerii/hub/backend/ent/mailmutationreceipt"
	"github.com/valerii/hub/backend/ent/mailobject"
	"github.com/valerii/hub/backend/ent/mailoauthcredential"
	"github.com/valerii/hub/backend/ent/mailrule"
	"github.com/valerii/hub/backend/ent/mailsignature"
	"github.com/valerii/hub/backend/ent/mailsubmission"
	"github.com/valerii/hub/backend/ent/mailthread"
	"github.com/valerii/hub/backend/ent/mailusersettings"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/recentcontact"
)

// inboundRequestFixturePublishedMailAccountClosureObservation is the private,
// comparable result for Mail closure slots 16-27. It retains no transaction-
// bound Ent rows, and leaves false-sent proof construction to its caller.
type inboundRequestFixturePublishedMailAccountClosureObservation struct {
	roots inboundRequestFixtureRootProof

	mailDraftCount           int
	mailSubmissionCount      int
	mailUserSettingsCount    int
	mailOAuthCredentialCount int
	mailSignatureCount       int
	mailRuleCount            int
	mailLabelCount           int
	mailObjectCount          int
	recentContactCount       int
	mailMutationReceiptCount int

	mailThreadCount         int
	mailThreadID            uuid.UUID
	mailThreadMailAccountID uuid.UUID

	mailIdentityCount         int
	mailIdentityID            uuid.UUID
	mailIdentityMailAccountID uuid.UUID
}

func (observation inboundRequestFixturePublishedMailAccountClosureObservation) valid() bool {
	return observation.roots.valid() &&
		observation.mailDraftCount == 0 &&
		observation.mailSubmissionCount == 0 &&
		observation.mailUserSettingsCount == 0 &&
		observation.mailOAuthCredentialCount == 0 &&
		observation.mailSignatureCount == 0 &&
		observation.mailRuleCount == 0 &&
		observation.mailLabelCount == 0 &&
		observation.mailObjectCount == 0 &&
		observation.recentContactCount == 0 &&
		observation.mailMutationReceiptCount == 0 &&
		observation.mailThreadCount == 1 &&
		observation.mailThreadID == observation.roots.mailThreadID &&
		observation.mailThreadMailAccountID == observation.roots.mailAccountID &&
		observation.mailIdentityCount == 1 &&
		observation.mailIdentityID == observation.roots.mailIdentityID &&
		observation.mailIdentityMailAccountID == observation.roots.mailAccountID
}

// observeInboundRequestFixturePublishedMailAccountClosure locks and proves the
// account half of the Mail closure in its caller-owned transaction. The ten
// dependent sets are fail-closed empty, while the complete account thread and
// identity sets must each contain only their C1 root.
func observeInboundRequestFixturePublishedMailAccountClosure(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
) (inboundRequestFixturePublishedMailAccountClosureObservation, bool) {
	roots, valid := newInboundRequestFixtureRootProof(plan, persisted)
	if ctx == nil ||
		tx == nil ||
		!valid ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!authority.valid() ||
		authority.nonce != plan.nonce ||
		lockedRoots.roots != roots ||
		!inboundRequestFixtureLockedPublishedRootsValid(
			roots,
			plan,
			persisted,
			authority,
			lockedRoots.mail,
			lockedRoots.account,
			lockedRoots.owner,
		) {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	drafts, err := tx.MailDraft.Query().
		Where(
			maildraft.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailDraftClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(drafts) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	submissions, err := tx.MailSubmission.Query().
		Where(
			mailsubmission.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailSubmissionAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(submissions) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	userSettings, err := tx.MailUserSettings.Query().
		Where(
			mailusersettings.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailUserSettingsClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(userSettings) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	oauthCredentials, err := tx.MailOAuthCredential.Query().
		Where(
			mailoauthcredential.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailOAuthCredentialAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(oauthCredentials) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	signatures, err := tx.MailSignature.Query().
		Where(
			mailsignature.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailSignatureAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(signatures) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	rules, err := tx.MailRule.Query().
		Where(
			mailrule.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailRuleAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(rules) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	labels, err := tx.MailLabel.Query().
		Where(
			maillabel.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailLabelAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(labels) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	objects, err := tx.MailObject.Query().
		Where(
			mailobject.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailObjectAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(objects) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	contacts, err := tx.RecentContact.Query().
		Where(
			recentcontact.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedRecentContactAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(contacts) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	mutationReceipts, err := tx.MailMutationReceipt.Query().
		Where(
			mailmutationreceipt.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailMutationReceiptAccountClosureNoWait(),
		).
		Limit(1).
		All(ctx)
	if err != nil || len(mutationReceipts) != 0 {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	threads, err := tx.MailThread.Query().
		Where(
			mailthread.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailThreadAccountClosureNoWait(),
		).
		Order(ent.Asc(mailthread.FieldID)).
		All(ctx)
	if err != nil ||
		len(threads) != 1 ||
		!inboundRequestFixturePublishedMailAccountClosureThreadMatchesLocked(
			threads[0],
			lockedRoots.mail.thread,
			roots,
		) {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	identities, err := tx.MailIdentity.Query().
		Where(
			mailidentity.MailAccountIDEQ(roots.mailAccountID),
			inboundRequestFixturePublishedMailIdentityNoWait(),
		).
		Order(ent.Asc(mailidentity.FieldID)).
		All(ctx)
	if err != nil ||
		len(identities) != 1 ||
		!inboundRequestFixturePublishedMailAccountClosureIdentityMatchesLocked(
			identities[0],
			lockedRoots.account.primaryIdentity,
			roots,
		) {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}

	observation := inboundRequestFixturePublishedMailAccountClosureObservation{
		roots:                     roots,
		mailDraftCount:            len(drafts),
		mailSubmissionCount:       len(submissions),
		mailUserSettingsCount:     len(userSettings),
		mailOAuthCredentialCount:  len(oauthCredentials),
		mailSignatureCount:        len(signatures),
		mailRuleCount:             len(rules),
		mailLabelCount:            len(labels),
		mailObjectCount:           len(objects),
		recentContactCount:        len(contacts),
		mailMutationReceiptCount:  len(mutationReceipts),
		mailThreadCount:           len(threads),
		mailThreadID:              threads[0].ID,
		mailThreadMailAccountID:   threads[0].MailAccountID,
		mailIdentityCount:         len(identities),
		mailIdentityID:            identities[0].ID,
		mailIdentityMailAccountID: identities[0].MailAccountID,
	}
	if !observation.valid() {
		return inboundRequestFixturePublishedMailAccountClosureObservation{}, false
	}
	return observation, true
}

func inboundRequestFixturePublishedMailAccountClosureThreadMatchesLocked(
	value *ent.MailThread,
	locked *ent.MailThread,
	roots inboundRequestFixtureRootProof,
) bool {
	return value != nil &&
		locked != nil &&
		value.ID == roots.mailThreadID &&
		value.ID == locked.ID &&
		value.MailAccountID == roots.mailAccountID &&
		value.MailAccountID == locked.MailAccountID
}

func inboundRequestFixturePublishedMailAccountClosureIdentityMatchesLocked(
	value *ent.MailIdentity,
	locked *ent.MailIdentity,
	roots inboundRequestFixtureRootProof,
) bool {
	return value != nil &&
		locked != nil &&
		value.ID == roots.mailIdentityID &&
		value.ID == locked.ID &&
		value.MailAccountID == roots.mailAccountID &&
		value.MailAccountID == locked.MailAccountID
}

func inboundRequestFixturePublishedMailSubmissionAccountClosureNoWait() predicate.MailSubmission {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailOAuthCredentialAccountClosureNoWait() predicate.MailOAuthCredential {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailSignatureAccountClosureNoWait() predicate.MailSignature {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailRuleAccountClosureNoWait() predicate.MailRule {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailLabelAccountClosureNoWait() predicate.MailLabel {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailObjectAccountClosureNoWait() predicate.MailObject {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedRecentContactAccountClosureNoWait() predicate.RecentContact {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailMutationReceiptAccountClosureNoWait() predicate.MailMutationReceipt {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}

func inboundRequestFixturePublishedMailThreadAccountClosureNoWait() predicate.MailThread {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.NoWait))
	}
}
