//go:build qa && !release

package qamailcalendar

type inboundRequestFixturePublishedFaultKind string

const (
	inboundRequestFixturePublishedFaultOriginP3Duplicate                 inboundRequestFixturePublishedFaultKind = "origin-p3-duplicate-v1"
	inboundRequestFixturePublishedFaultOriginP4MalformedEnvelope         inboundRequestFixturePublishedFaultKind = "origin-p4-malformed-envelope-v1"
	inboundRequestFixturePublishedFaultExternalUserAIProvider            inboundRequestFixturePublishedFaultKind = "external-user-ai-provider-v1"
	inboundRequestFixturePublishedFaultCleanupCommitUnknownPresent       inboundRequestFixturePublishedFaultKind = "cleanup-commit-unknown-present-v1"
	inboundRequestFixturePublishedFaultCleanupCommitUnknownAbsent        inboundRequestFixturePublishedFaultKind = "cleanup-commit-unknown-absent-v1"
	inboundRequestFixturePublishedFaultCleanupSerializationRollback      inboundRequestFixturePublishedFaultKind = "cleanup-serialization-rollback-v1"
	inboundRequestFixturePublishedFaultCleanupEngineAfterDBDeleteFailure inboundRequestFixturePublishedFaultKind = "cleanup-engine-after-db-delete-failure-v1"
	inboundRequestFixturePublishedFaultCleanupFinalCASLost               inboundRequestFixturePublishedFaultKind = "cleanup-final-cas-lost-v1"
	inboundRequestFixturePublishedFaultCleanupMaplessPartial             inboundRequestFixturePublishedFaultKind = "cleanup-mapless-partial-v1"
	inboundRequestFixturePublishedFaultCleanupMaplessAmbiguous           inboundRequestFixturePublishedFaultKind = "cleanup-mapless-ambiguous-v1"
)

func parseInboundRequestFixturePublishedFaultKind(value string) (inboundRequestFixturePublishedFaultKind, bool) {
	kind := inboundRequestFixturePublishedFaultKind(value)
	return kind, validInboundRequestFixturePublishedFaultKind(kind)
}

func validInboundRequestFixturePublishedFaultKind(kind inboundRequestFixturePublishedFaultKind) bool {
	switch kind {
	case inboundRequestFixturePublishedFaultOriginP3Duplicate,
		inboundRequestFixturePublishedFaultOriginP4MalformedEnvelope,
		inboundRequestFixturePublishedFaultExternalUserAIProvider,
		inboundRequestFixturePublishedFaultCleanupCommitUnknownPresent,
		inboundRequestFixturePublishedFaultCleanupCommitUnknownAbsent,
		inboundRequestFixturePublishedFaultCleanupSerializationRollback,
		inboundRequestFixturePublishedFaultCleanupEngineAfterDBDeleteFailure,
		inboundRequestFixturePublishedFaultCleanupFinalCASLost,
		inboundRequestFixturePublishedFaultCleanupMaplessPartial,
		inboundRequestFixturePublishedFaultCleanupMaplessAmbiguous:
		return true
	default:
		return false
	}
}

func (c *Controller) armInboundRequestFixturePublishedFault(
	nonce string,
	kind inboundRequestFixturePublishedFaultKind,
) bool {
	if c == nil || c.publishedFaults == nil || !ValidNonce(nonce) ||
		!validInboundRequestFixturePublishedFaultKind(kind) {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.publishedFaults[nonce]; exists {
		return false
	}
	c.publishedFaults[nonce] = kind
	return true
}

func (c *Controller) takeInboundRequestFixturePublishedFault(
	nonce string,
	expected inboundRequestFixturePublishedFaultKind,
) bool {
	if c == nil || c.publishedFaults == nil || !ValidNonce(nonce) ||
		!validInboundRequestFixturePublishedFaultKind(expected) {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	armed, exists := c.publishedFaults[nonce]
	if !exists || armed != expected {
		return false
	}
	delete(c.publishedFaults, nonce)
	return true
}

func (c *Controller) disarmInboundRequestFixturePublishedFault(
	nonce string,
	expected inboundRequestFixturePublishedFaultKind,
) bool {
	return c.takeInboundRequestFixturePublishedFault(nonce, expected)
}
