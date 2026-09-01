//go:build qa && !release

package qamailcalendar

import "github.com/gofiber/fiber/v3"

const inboundRequestFixturePublishedFaultRoutePath = "/_qa/mail-calendar/fixtures/:nonce/published-cleanup-fault/:kind"

func (c *Controller) registerInboundRequestFixturePublishedFaultRoutes(api fiber.Router) error {
	if c == nil || api == nil || c.publishedFaults == nil {
		return errInboundRequestFixtureControllerInput
	}

	api.Post(inboundRequestFixturePublishedFaultRoutePath, c.armInboundRequestFixturePublishedFaultHTTP)
	api.Delete(inboundRequestFixturePublishedFaultRoutePath, c.disarmInboundRequestFixturePublishedFaultHTTP)
	return nil
}

func (c *Controller) armInboundRequestFixturePublishedFaultHTTP(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}
	kind, ok := parseInboundRequestFixturePublishedFaultKind(ctx.Params("kind"))
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}
	if !c.armInboundRequestFixturePublishedFaultForHTTP(nonce, kind) {
		return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
	}
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

func (c *Controller) disarmInboundRequestFixturePublishedFaultHTTP(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}
	kind, ok := parseInboundRequestFixturePublishedFaultKind(ctx.Params("kind"))
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}
	if !c.disarmInboundRequestFixturePublishedFault(nonce, kind) {
		return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
	}
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

func (c *Controller) armInboundRequestFixturePublishedFaultForHTTP(
	nonce string,
	kind inboundRequestFixturePublishedFaultKind,
) bool {
	if c == nil || c.publishedFaults == nil || !ValidNonce(nonce) ||
		!validInboundRequestFixturePublishedFaultKind(kind) {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.publishedFaults[nonce]; exists ||
		!c.inboundRequestFixturePublishedFaultTargetStateLocked(nonce, kind) {
		return false
	}
	c.publishedFaults[nonce] = kind
	return true
}

func (c *Controller) inboundRequestFixturePublishedFaultTargetStateLocked(
	nonce string,
	kind inboundRequestFixturePublishedFaultKind,
) bool {
	record, found := c.fixtures[nonce]
	switch kind {
	case inboundRequestFixturePublishedFaultOriginP3Duplicate,
		inboundRequestFixturePublishedFaultOriginP4MalformedEnvelope,
		inboundRequestFixturePublishedFaultExternalUserAIProvider,
		inboundRequestFixturePublishedFaultCleanupCommitUnknownPresent,
		inboundRequestFixturePublishedFaultCleanupCommitUnknownAbsent,
		inboundRequestFixturePublishedFaultCleanupSerializationRollback,
		inboundRequestFixturePublishedFaultCleanupEngineAfterDBDeleteFailure,
		inboundRequestFixturePublishedFaultCleanupFinalCASLost:
		return found &&
			record.nonce == nonce &&
			record.state == inboundRequestFixtureStatePublished
	case inboundRequestFixturePublishedFaultCleanupMaplessPartial,
		inboundRequestFixturePublishedFaultCleanupMaplessAmbiguous:
		return !found &&
			c.publishedRecoveryRouteInstalled &&
			len(c.fixtures) == 0 &&
			len(c.accountMaterializationFixtures) == 0 &&
			len(c.threadSearchFixtures) == 0
	default:
		return false
	}
}
