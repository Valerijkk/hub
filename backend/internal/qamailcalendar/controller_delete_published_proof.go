//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
)

type inboundRequestFixturePublishedFullProof struct {
	roots           inboundRequestFixtureRootProof
	mailClosure     inboundRequestFixturePublishedMailClosure
	calendar        inboundRequestFixtureCalendarProof
	cleanup         cleanupDomainProof
	external        inboundRequestFixturePublishedExternalProof
	outbox          inboundRequestFixturePublishedProjectOutboxProof
	engineAuthority inboundRequestFixturePublishedEngineAuthority
}

func (proof inboundRequestFixturePublishedFullProof) valid(
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
) bool {
	if !inboundRequestFixturePlanIsCanonical(plan) ||
		!inboundRequestFixturePersistedMatchesPlan(persisted, plan) ||
		!inboundRequestFixtureRootProofMatchesPlan(proof.roots, plan, persisted) ||
		!proof.engineAuthority.valid() ||
		proof.engineAuthority.nonce != plan.nonce ||
		!proof.mailClosure.valid() ||
		proof.mailClosure.roots != proof.roots ||
		!proof.calendar.valid() {
		return false
	}

	if _, valid := newInboundRequestFixtureCompleteGraphProof(
		proof.roots,
		proof.mailClosure.falseSent,
		proof.calendar,
	); !valid {
		return false
	}

	if !proof.external.valid(proof.roots, proof.calendar) ||
		!proof.outbox.valid(proof.roots, proof.calendar, proof.engineAuthority) {
		return false
	}

	cleanup, valid := newCleanupDomainProof(proof.calendar)
	return valid && proof.cleanup.present && proof.cleanup == cleanup
}

func observeInboundRequestFixturePublishedFullProof(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	engineAuthority inboundRequestFixturePublishedEngineAuthority,
) (
	inboundRequestFixturePublishedFullProof,
	inboundRequestFixtureLockedPublishedRoots,
	bool,
) {
	if ctx == nil ||
		tx == nil ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!inboundRequestFixturePersistedMatchesPlan(persisted, plan) ||
		!engineAuthority.valid() ||
		engineAuthority.nonce != plan.nonce {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}

	lockedRoots, locked := lockInboundRequestFixturePublishedRoots(
		ctx,
		tx,
		plan,
		persisted,
		engineAuthority,
	)
	if !locked {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}
	roots := lockedRoots.roots

	mailClosure, observed := observeInboundRequestFixturePublishedMailClosure(
		ctx,
		tx,
		plan,
		persisted,
		engineAuthority,
		lockedRoots,
	)
	if !observed {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}

	calendar, observed := observeInboundRequestFixturePublishedCalendarEffects(
		ctx,
		tx,
		plan,
		persisted,
		engineAuthority,
		lockedRoots,
		mailClosure,
	)
	if !observed {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}

	cleanup, valid := newCleanupDomainProof(calendar)
	if !valid {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}

	external, observed := observeInboundRequestFixturePublishedExternal(
		ctx,
		tx,
		roots,
		calendar,
	)
	if !observed {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}

	outbox, observed := observeInboundRequestFixturePublishedProjectOutbox(
		ctx,
		tx,
		plan,
		persisted,
		roots,
		calendar,
		engineAuthority,
	)
	if !observed {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}

	proof := inboundRequestFixturePublishedFullProof{
		roots:           roots,
		mailClosure:     mailClosure,
		calendar:        calendar,
		cleanup:         cleanup,
		external:        external,
		outbox:          outbox,
		engineAuthority: engineAuthority,
	}
	if !proof.valid(plan, persisted) ||
		!inboundRequestFixtureLockedPublishedRootsValid(
			proof.roots,
			plan,
			persisted,
			proof.engineAuthority,
			lockedRoots.mail,
			lockedRoots.account,
			lockedRoots.owner,
		) {
		return inboundRequestFixturePublishedFullProof{}, inboundRequestFixtureLockedPublishedRoots{}, false
	}

	return proof, lockedRoots, true
}
