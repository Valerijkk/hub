//go:build qa && !release

package qamailcalendar

import "sync/atomic"

// PublishedRecoveryChildEvidence carries data from the trusted server-composition
// TCB. It is not a cryptographic or unforgeable capability; HTTP, config, and
// raw tokens do not issue it.
type PublishedRecoveryChildEvidence struct {
	ChildPID   int
	Generation uint64
	BootSHA256 [32]byte
}

// PublishedRecoveryRouteSession supplies trusted server-composition evidence.
// It is not an HTTP/config/raw-token issuer or an unforgeable capability.
type PublishedRecoveryRouteSession interface {
	PublishedRecoveryChildEvidence() PublishedRecoveryChildEvidence
}

type inboundRequestFixturePublishedMaplessAdmission struct {
	evidence PublishedRecoveryChildEvidence
	claimed  atomic.Bool
}

func (a *inboundRequestFixturePublishedMaplessAdmission) claim() bool {
	if a == nil {
		return false
	}
	return a.claimed.CompareAndSwap(false, true)
}

func (c *Controller) installInboundRequestFixturePublishedMaplessAdmission(
	session PublishedRecoveryRouteSession,
) (*inboundRequestFixturePublishedMaplessAdmission, bool) {
	if c == nil ||
		session == nil ||
		c.client == nil ||
		c.engine == nil ||
		!validToken(string(c.config.token)) {
		return nil, false
	}

	evidence := session.PublishedRecoveryChildEvidence()
	var zeroFingerprint [32]byte
	if evidence.ChildPID <= 1 ||
		evidence.Generation != 2 ||
		evidence.BootSHA256 == zeroFingerprint {
		return nil, false
	}

	// Worker suppression outside this method is required: engine and map
	// observations are snapshots, while this critical section reserves the route.
	if !c.engine.IsEmpty() {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fixtures == nil ||
		c.accountMaterializationFixtures == nil ||
		c.threadSearchFixtures == nil ||
		len(c.fixtures) != 0 ||
		len(c.accountMaterializationFixtures) != 0 ||
		len(c.threadSearchFixtures) != 0 ||
		c.publishedRecoveryRouteInstalled {
		return nil, false
	}

	c.publishedRecoveryRouteInstalled = true
	return &inboundRequestFixturePublishedMaplessAdmission{evidence: evidence}, true
}

func (c *Controller) installInboundRequestFixturePublishedRestartRoute() bool {
	if c == nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.publishedRecoveryRouteInstalled {
		return false
	}

	c.publishedRecoveryRouteInstalled = true
	return true
}

func (c *Controller) canRestartInboundRequestFixturePublishedRecovery(nonce string) bool {
	if c == nil ||
		c.client == nil ||
		c.engine == nil ||
		!validToken(string(c.config.token)) {
		return false
	}

	canonicalPlan, err := newInboundRequestFixturePlan(nonce)
	if err != nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fixtures == nil {
		return false
	}
	record, found := c.fixtures[nonce]
	if !found ||
		record.nonce != nonce ||
		record.plan != canonicalPlan ||
		!inboundRequestFixturePersistedMatchesPlan(record.persisted, record.plan) ||
		record.state != inboundRequestFixtureStateCleanupCommitUnknown ||
		!inboundRequestFixtureStateAcceptsCleanupDomainProof(record.state, record.cleanupProof) ||
		!record.cleanupProof.present {
		return false
	}

	authority, valid := newInboundRequestFixturePublishedMappedEngineAuthority(
		record.plan,
		record.engineFixture,
	)
	return valid && authority.mappedMode()
}
