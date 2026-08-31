//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
)

// inboundRequestFixturePublishedMailClosure is the private comparable result
// of the complete frozen Mail closure. It retains only the deterministic root
// proof and the derived false-sent proof; the caller-owned transaction retains
// every Ent row used while observing the closure.
type inboundRequestFixturePublishedMailClosure struct {
	roots     inboundRequestFixtureRootProof
	falseSent inboundRequestFixtureFalseSentProof
}

func (closure inboundRequestFixturePublishedMailClosure) valid() bool {
	return closure.roots.valid() && closure.falseSent.matchesRoots(closure.roots)
}

// observeInboundRequestFixturePublishedMailClosure observes C2a, C2b, then
// C2c in the frozen order and joins their comparable observations into the
// false-sent proof. The caller owns transaction isolation, commit, rollback,
// and every later Calendar, external, Outbox, lifecycle, or delete step.
func observeInboundRequestFixturePublishedMailClosure(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
	lockedRoots inboundRequestFixtureLockedPublishedRoots,
) (inboundRequestFixturePublishedMailClosure, bool) {
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
		return inboundRequestFixturePublishedMailClosure{}, false
	}

	messageObservation, observed := lockInboundRequestFixturePublishedMailClosureMessage(
		ctx,
		tx,
		plan,
		persisted,
		authority,
		lockedRoots,
	)
	if !observed || !messageObservation.valid() {
		return inboundRequestFixturePublishedMailClosure{}, false
	}

	folderIdentityObservation, observed := observeInboundRequestFixturePublishedMailFolderIdentityClosure(
		ctx,
		tx,
		plan,
		persisted,
		lockedRoots,
	)
	if !observed ||
		!folderIdentityObservation.valid() ||
		folderIdentityObservation.roots != roots {
		return inboundRequestFixturePublishedMailClosure{}, false
	}

	accountObservation, observed := observeInboundRequestFixturePublishedMailAccountClosure(
		ctx,
		tx,
		plan,
		persisted,
		authority,
		lockedRoots,
	)
	if !observed ||
		!accountObservation.valid() ||
		accountObservation.roots != roots {
		return inboundRequestFixturePublishedMailClosure{}, false
	}

	falseSent, valid := newInboundRequestFixtureFalseSentProof(
		roots,
		roots.mailAccountID,
		roots.mailMessageID,
		messageObservation.originalInboundMessageCount,
		accountObservation.mailDraftCount,
		accountObservation.mailSubmissionCount,
	)
	if !valid || !falseSent.matchesRoots(roots) ||
		!inboundRequestFixtureLockedPublishedRootsValid(
			roots,
			plan,
			persisted,
			authority,
			lockedRoots.mail,
			lockedRoots.account,
			lockedRoots.owner,
		) {
		return inboundRequestFixturePublishedMailClosure{}, false
	}

	closure := inboundRequestFixturePublishedMailClosure{
		roots:     roots,
		falseSent: falseSent,
	}
	if !closure.valid() {
		return inboundRequestFixturePublishedMailClosure{}, false
	}
	return closure, true
}
