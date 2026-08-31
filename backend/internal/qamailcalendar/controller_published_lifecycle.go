//go:build qa && !release

package qamailcalendar

type inboundRequestFixturePublishedCleanupTransition uint8

const (
	inboundRequestFixturePublishedCleanupTransitionInvalid inboundRequestFixturePublishedCleanupTransition = iota
	inboundRequestFixturePublishedCleanupTransitionClaim
	inboundRequestFixturePublishedCleanupTransitionRollbackRestore
	inboundRequestFixturePublishedCleanupTransitionCommitUnknown
	inboundRequestFixturePublishedCleanupTransitionClaimedEngineFailed
	inboundRequestFixturePublishedCleanupTransitionClaimedCleaned
	inboundRequestFixturePublishedCleanupTransitionCommitUnknownPresent
	inboundRequestFixturePublishedCleanupTransitionCommitUnknownClaimed
)

func (c *Controller) transitionInboundRequestFixturePublishedCleanup(
	nonce string,
	expected inboundRequestFixtureRecord,
	transition inboundRequestFixturePublishedCleanupTransition,
	claimedProof cleanupDomainProof,
) (inboundRequestFixtureRecord, bool) {
	if c == nil {
		return inboundRequestFixtureRecord{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	canonicalPlan, err := newInboundRequestFixturePlan(nonce)
	if c.fixtures == nil ||
		!ValidNonce(nonce) ||
		err != nil ||
		expected.nonce != nonce ||
		expected.plan != canonicalPlan ||
		!inboundRequestFixturePersistedMatchesPlan(expected.persisted, expected.plan) ||
		!inboundRequestFixtureStateAcceptsCleanupDomainProof(expected.state, expected.cleanupProof) {
		return inboundRequestFixtureRecord{}, false
	}

	current, found := c.fixtures[nonce]
	if !found || current != expected {
		return inboundRequestFixtureRecord{}, false
	}

	next := expected
	switch transition {
	case inboundRequestFixturePublishedCleanupTransitionClaim:
		if expected.state != inboundRequestFixtureStatePublished ||
			!claimedProof.present {
			return inboundRequestFixtureRecord{}, false
		}
		next.state = inboundRequestFixtureStateCleanupClaimed
		next.cleanupProof = claimedProof
	case inboundRequestFixturePublishedCleanupTransitionRollbackRestore:
		if expected.state != inboundRequestFixtureStateCleanupClaimed ||
			claimedProof != (cleanupDomainProof{}) {
			return inboundRequestFixtureRecord{}, false
		}
		next.state = inboundRequestFixtureStatePublished
		next.cleanupProof = cleanupDomainProof{}
	case inboundRequestFixturePublishedCleanupTransitionCommitUnknown:
		if expected.state != inboundRequestFixtureStateCleanupClaimed ||
			claimedProof != (cleanupDomainProof{}) {
			return inboundRequestFixtureRecord{}, false
		}
		next.state = inboundRequestFixtureStateCleanupCommitUnknown
	case inboundRequestFixturePublishedCleanupTransitionClaimedEngineFailed:
		if expected.state != inboundRequestFixtureStateCleanupClaimed ||
			claimedProof != (cleanupDomainProof{}) {
			return inboundRequestFixtureRecord{}, false
		}
		next.state = inboundRequestFixtureStateCleanupEngineFailed
	case inboundRequestFixturePublishedCleanupTransitionClaimedCleaned:
		if expected.state != inboundRequestFixtureStateCleanupClaimed ||
			claimedProof != (cleanupDomainProof{}) {
			return inboundRequestFixtureRecord{}, false
		}
		next.state = inboundRequestFixtureStateCleaned
	case inboundRequestFixturePublishedCleanupTransitionCommitUnknownPresent:
		if expected.state != inboundRequestFixtureStateCleanupCommitUnknown ||
			claimedProof != (cleanupDomainProof{}) {
			return inboundRequestFixtureRecord{}, false
		}
		next.state = inboundRequestFixtureStatePublished
		next.cleanupProof = cleanupDomainProof{}
	case inboundRequestFixturePublishedCleanupTransitionCommitUnknownClaimed:
		if expected.state != inboundRequestFixtureStateCleanupCommitUnknown ||
			claimedProof != (cleanupDomainProof{}) {
			return inboundRequestFixtureRecord{}, false
		}
		next.state = inboundRequestFixtureStateCleanupClaimed
	default:
		return inboundRequestFixtureRecord{}, false
	}

	if !inboundRequestFixtureStateAcceptsCleanupDomainProof(next.state, next.cleanupProof) {
		return inboundRequestFixtureRecord{}, false
	}

	c.fixtures[nonce] = next
	return next, true
}

func (c *Controller) claimInboundRequestFixturePublishedCleanup(
	nonce string,
	expected inboundRequestFixtureRecord,
	proof cleanupDomainProof,
) (inboundRequestFixtureRecord, bool) {
	return c.transitionInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		inboundRequestFixturePublishedCleanupTransitionClaim,
		proof,
	)
}

func (c *Controller) restoreInboundRequestFixturePublishedAfterRollback(
	nonce string,
	expected inboundRequestFixtureRecord,
) (inboundRequestFixtureRecord, bool) {
	return c.transitionInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		inboundRequestFixturePublishedCleanupTransitionRollbackRestore,
		cleanupDomainProof{},
	)
}

func (c *Controller) markInboundRequestFixturePublishedCleanupCommitUnknown(
	nonce string,
	expected inboundRequestFixtureRecord,
) (inboundRequestFixtureRecord, bool) {
	return c.transitionInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		inboundRequestFixturePublishedCleanupTransitionCommitUnknown,
		cleanupDomainProof{},
	)
}

func (c *Controller) markInboundRequestFixturePublishedCleanupEngineFailedFromClaim(
	nonce string,
	expected inboundRequestFixtureRecord,
) (inboundRequestFixtureRecord, bool) {
	return c.transitionInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		inboundRequestFixturePublishedCleanupTransitionClaimedEngineFailed,
		cleanupDomainProof{},
	)
}

func (c *Controller) markInboundRequestFixturePublishedCleanupCleanedFromClaim(
	nonce string,
	expected inboundRequestFixtureRecord,
) (inboundRequestFixtureRecord, bool) {
	return c.transitionInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		inboundRequestFixturePublishedCleanupTransitionClaimedCleaned,
		cleanupDomainProof{},
	)
}

func (c *Controller) restoreInboundRequestFixturePublishedAfterCommitUnknownPresent(
	nonce string,
	expected inboundRequestFixtureRecord,
) (inboundRequestFixtureRecord, bool) {
	return c.transitionInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		inboundRequestFixturePublishedCleanupTransitionCommitUnknownPresent,
		cleanupDomainProof{},
	)
}

func (c *Controller) claimInboundRequestFixturePublishedCleanupEngineFromCommitUnknown(
	nonce string,
	expected inboundRequestFixtureRecord,
) (inboundRequestFixtureRecord, bool) {
	return c.transitionInboundRequestFixturePublishedCleanup(
		nonce,
		expected,
		inboundRequestFixturePublishedCleanupTransitionCommitUnknownClaimed,
		cleanupDomainProof{},
	)
}
