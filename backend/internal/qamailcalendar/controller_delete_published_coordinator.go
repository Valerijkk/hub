//go:build qa && !release

package qamailcalendar

import (
	"context"
	"database/sql"

	"github.com/valerii/hub/backend/ent"
)

func (c *Controller) captureInboundRequestFixturePublished(
	ctx context.Context,
	nonce string,
) (
	inboundRequestFixtureRecord,
	inboundRequestFixturePlan,
	inboundRequestFixturePersisted,
	inboundRequestFixturePublishedEngineAuthority,
	bool,
) {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.engine == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return inboundRequestFixtureRecord{}, inboundRequestFixturePlan{}, inboundRequestFixturePersisted{}, inboundRequestFixturePublishedEngineAuthority{}, false
	}

	canonicalPlan, err := newInboundRequestFixturePlan(nonce)
	if err != nil {
		return inboundRequestFixtureRecord{}, inboundRequestFixturePlan{}, inboundRequestFixturePersisted{}, inboundRequestFixturePublishedEngineAuthority{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fixtures == nil {
		return inboundRequestFixtureRecord{}, inboundRequestFixturePlan{}, inboundRequestFixturePersisted{}, inboundRequestFixturePublishedEngineAuthority{}, false
	}
	record, found := c.fixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.state != inboundRequestFixtureStatePublished ||
		record.cleanupProof != (cleanupDomainProof{}) ||
		!inboundRequestFixtureStateAcceptsCleanupDomainProof(record.state, record.cleanupProof) ||
		record.plan != canonicalPlan ||
		!inboundRequestFixturePersistedMatchesPlan(record.persisted, record.plan) {
		return inboundRequestFixtureRecord{}, inboundRequestFixturePlan{}, inboundRequestFixturePersisted{}, inboundRequestFixturePublishedEngineAuthority{}, false
	}

	authority, valid := newInboundRequestFixturePublishedMappedEngineAuthority(
		record.plan,
		record.engineFixture,
	)
	if !valid {
		return inboundRequestFixtureRecord{}, inboundRequestFixturePlan{}, inboundRequestFixturePersisted{}, inboundRequestFixturePublishedEngineAuthority{}, false
	}
	return record, record.plan, record.persisted, authority, true
}

func (c *Controller) renewInboundRequestFixturePublishedFullProofAfterRollback(
	ctx context.Context,
	initialTx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	engineAuthority inboundRequestFixturePublishedEngineAuthority,
	first inboundRequestFixturePublishedFullProof,
) bool {
	if initialTx == nil {
		return false
	}
	if err := initialTx.Rollback(); err != nil {
		return false
	}
	if c == nil || ctx == nil || c.client == nil {
		return false
	}

	renewalTx, err := c.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return false
	}
	if !prepareInboundRequestFixturePublishedAuthCleanup(ctx, renewalTx, plan, persisted, engineAuthority) {
		_ = renewalTx.Rollback()
		return false
	}
	renewed, _, observed := observeInboundRequestFixturePublishedFullProof(
		ctx,
		renewalTx,
		plan,
		persisted,
		engineAuthority,
	)
	releaseErr := renewalTx.Rollback()
	return observed && renewed == first && releaseErr == nil
}

func (c *Controller) deleteInboundRequestFixturePublished(
	ctx context.Context,
	nonce string,
) error {
	expected, plan, persisted, engineAuthority, captured := c.captureInboundRequestFixturePublished(ctx, nonce)
	if !captured {
		return errInboundRequestFixtureLifecycle
	}

	tx, err := c.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return errInboundRequestFixtureLifecycle
	}
	commitAttempted := false
	rollbackBeforeCommit := func() bool {
		return !commitAttempted && tx.Rollback() == nil
	}
	if !prepareInboundRequestFixturePublishedAuthCleanup(ctx, tx, plan, persisted, engineAuthority) {
		if !rollbackBeforeCommit() {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	first, _, observed := observeInboundRequestFixturePublishedFullProof(
		ctx,
		tx,
		plan,
		persisted,
		engineAuthority,
	)
	if !observed {
		if !rollbackBeforeCommit() {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}
	if !c.applyInboundRequestFixturePublishedProofFault(ctx, tx, plan, first) {
		if !rollbackBeforeCommit() {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	claimed, claimedOK := c.claimInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		first.cleanup,
	)
	if !claimedOK {
		if !rollbackBeforeCommit() {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	second, secondLockedRoots, observed := observeInboundRequestFixturePublishedFullProof(
		ctx,
		tx,
		plan,
		persisted,
		engineAuthority,
	)
	if !observed || first != second {
		if commitAttempted ||
			!c.renewInboundRequestFixturePublishedFullProofAfterRollback(
				ctx,
				tx,
				plan,
				persisted,
				engineAuthority,
				first,
			) {
			return errInboundRequestFixtureLifecycle
		}
		if _, restored := c.restoreInboundRequestFixturePublishedAfterRollback(nonce, claimed); !restored {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	if !deleteInboundRequestFixturePublishedSequenceExact(
		ctx,
		tx,
		plan,
		persisted,
		second.engineAuthority,
		secondLockedRoots,
		second.mailClosure,
		second.calendar,
		second.external,
		second.outbox,
	) {
		if commitAttempted ||
			!c.renewInboundRequestFixturePublishedFullProofAfterRollback(
				ctx,
				tx,
				plan,
				persisted,
				engineAuthority,
				first,
			) {
			return errInboundRequestFixtureLifecycle
		}
		if _, restored := c.restoreInboundRequestFixturePublishedAfterRollback(nonce, claimed); !restored {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	if c.takeInboundRequestFixturePublishedFault(
		nonce,
		inboundRequestFixturePublishedFaultCleanupSerializationRollback,
	) {
		if !c.renewInboundRequestFixturePublishedFullProofAfterRollback(
			ctx,
			tx,
			plan,
			persisted,
			engineAuthority,
			first,
		) {
			return errInboundRequestFixtureLifecycle
		}
		if _, restored := c.restoreInboundRequestFixturePublishedAfterRollback(nonce, claimed); !restored {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	if c.takeInboundRequestFixturePublishedFault(
		nonce,
		inboundRequestFixturePublishedFaultCleanupCommitUnknownPresent,
	) {
		if !rollbackBeforeCommit() {
			return errInboundRequestFixtureLifecycle
		}
		if _, marked := c.markInboundRequestFixturePublishedCleanupCommitUnknown(nonce, claimed); !marked {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	commitAttempted = true
	if err := tx.Commit(); err != nil {
		if _, marked := c.markInboundRequestFixturePublishedCleanupCommitUnknown(nonce, claimed); !marked {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	if c.takeInboundRequestFixturePublishedFault(
		nonce,
		inboundRequestFixturePublishedFaultCleanupCommitUnknownAbsent,
	) {
		if _, marked := c.markInboundRequestFixturePublishedCleanupCommitUnknown(nonce, claimed); !marked {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	_, err = c.finalizeInboundRequestFixturePublishedCleanupEngine(claimed)
	return err
}
