//go:build qa && !release

package qamailcalendar

import (
	"context"
	"database/sql"
)

func (c *Controller) deleteInboundRequestFixturePublishedMapless(
	ctx context.Context,
	nonce string,
	admission *inboundRequestFixturePublishedMaplessAdmission,
) error {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		c.engine == nil ||
		admission == nil ||
		!validToken(string(c.config.token)) ||
		!ValidNonce(nonce) {
		return errInboundRequestFixtureLifecycle
	}

	plan, err := newInboundRequestFixturePlan(nonce)
	if err != nil || !inboundRequestFixturePlanIsCanonical(plan) {
		return errInboundRequestFixtureLifecycle
	}
	persisted, valid := newInboundRequestFixturePersistedFromPlan(plan)
	if !valid {
		return errInboundRequestFixtureLifecycle
	}
	engineAuthority, valid := newInboundRequestFixturePublishedMaplessEngineAuthority(plan)
	if !valid || !engineAuthority.maplessMode() {
		return errInboundRequestFixtureLifecycle
	}

	if !admission.claim() {
		return errInboundRequestFixtureLifecycle
	}

	tx, err := c.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil || tx == nil {
		if tx != nil {
			_ = tx.Rollback()
		}
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
	faultInjected, faultApplied := c.applyInboundRequestFixturePublishedMaplessProofFault(
		ctx,
		tx,
		plan,
		first,
	)
	if !faultApplied {
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
	if faultInjected || !observed || first != second || second.engineAuthority != engineAuthority {
		if !rollbackBeforeCommit() {
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
		if !rollbackBeforeCommit() {
			return errInboundRequestFixtureLifecycle
		}
		return errInboundRequestFixtureLifecycle
	}

	commitAttempted = true
	if err := tx.Commit(); err != nil {
		return errInboundRequestFixtureLifecycle
	}
	return nil
}
