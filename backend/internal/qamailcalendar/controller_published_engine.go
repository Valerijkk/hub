//go:build qa && !release

package qamailcalendar

import "errors"

func (c *Controller) finalizeInboundRequestFixturePublishedCleanupEngine(
	expected inboundRequestFixtureRecord,
) (inboundRequestFixtureRecord, error) {
	if c == nil ||
		c.engine == nil ||
		!ValidNonce(expected.nonce) {
		return inboundRequestFixtureRecord{}, errInboundRequestFixtureLifecycle
	}

	canonicalPlan, err := newInboundRequestFixturePlan(expected.nonce)
	if err != nil ||
		expected.plan != canonicalPlan ||
		!inboundRequestFixturePersistedMatchesPlan(expected.persisted, expected.plan) ||
		expected.state != inboundRequestFixtureStateCleanupClaimed ||
		!inboundRequestFixtureStateAcceptsCleanupDomainProof(expected.state, expected.cleanupProof) ||
		!expected.cleanupProof.present ||
		expected.engineFixture.accountID == "" ||
		expected.engineFixture.primaryIdentityID == "" ||
		expected.engineFixture.folderID == "" ||
		expected.engineFixture.messageID == "" ||
		expected.engineFixture.version == "" {
		return inboundRequestFixtureRecord{}, errInboundRequestFixtureLifecycle
	}

	if err := destroyInboundRequestFixtureEngine(c.engine, expected.engineFixture); err != nil {
		failed, marked := c.markInboundRequestFixturePublishedCleanupEngineFailedFromClaim(
			expected.nonce,
			expected,
		)
		if !marked {
			return inboundRequestFixtureRecord{}, errInboundRequestFixtureLifecycle
		}
		return failed, errors.Join(
			errInboundRequestFixtureLifecycle,
			errInboundRequestFixtureCleanup,
		)
	}

	cleaned, marked := c.markInboundRequestFixturePublishedCleanupCleanedFromClaim(
		expected.nonce,
		expected,
	)
	if !marked {
		return inboundRequestFixtureRecord{}, errInboundRequestFixtureLifecycle
	}
	return cleaned, nil
}
