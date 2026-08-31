//go:build qa && !release

package qamailcalendar

import (
	"encoding/hex"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const (
	publishedRecoveryRoutePath             = "/_qa/mail-calendar/fixtures/:nonce/published-recovery"
	publishedRecoveryChildPIDHeader        = "X-Hub-QA-Child-PID"
	publishedRecoveryChildGenerationHeader = "X-Hub-QA-Child-Generation"
	publishedRecoveryChildBootSHA256Header = "X-Hub-QA-Child-Boot-SHA256"
)

type publishedRecoveryRouteEvidenceSession struct {
	evidence PublishedRecoveryChildEvidence
}

func (s publishedRecoveryRouteEvidenceSession) PublishedRecoveryChildEvidence() PublishedRecoveryChildEvidence {
	return s.evidence
}

// RegisterPublishedRecoveryRoute exposes the closed QA-only published-cleanup
// recovery surface on /api. Its session is supplied only by trusted server
// composition before Fiber begins serving requests.
func (c *Controller) RegisterPublishedRecoveryRoute(
	api fiber.Router,
	session PublishedRecoveryRouteSession,
	requestRestart func() error,
) error {
	if c == nil ||
		api == nil ||
		session == nil ||
		c.client == nil ||
		c.engine == nil ||
		!validToken(string(c.config.token)) {
		return errInboundRequestFixtureControllerInput
	}

	evidence := session.PublishedRecoveryChildEvidence()
	var zeroFingerprint [32]byte
	if evidence.ChildPID <= 1 ||
		evidence.BootSHA256 == zeroFingerprint {
		return errInboundRequestFixtureControllerInput
	}

	var handler fiber.Handler
	switch evidence.Generation {
	case 1:
		if requestRestart == nil || !c.installInboundRequestFixturePublishedRestartRoute() {
			return errInboundRequestFixtureLifecycle
		}
		handler = c.publishedRecoveryRestartHTTP(evidence, requestRestart)
	case 2:
		if requestRestart != nil {
			return errInboundRequestFixtureLifecycle
		}
		admission, installed := c.installInboundRequestFixturePublishedMaplessAdmission(
			publishedRecoveryRouteEvidenceSession{evidence: evidence},
		)
		if !installed {
			return errInboundRequestFixtureLifecycle
		}
		handler = c.publishedRecoveryMaplessHTTP(evidence, admission)
	default:
		return errInboundRequestFixtureControllerInput
	}

	api.Post(publishedRecoveryRoutePath, handler)
	return nil
}

func (c *Controller) publishedRecoveryRestartHTTP(
	evidence PublishedRecoveryChildEvidence,
	requestRestart func() error,
) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if !c.authorized(ctx) {
			return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
		}

		nonce, ok := qaFixtureMutationNonce(ctx)
		if !ok {
			return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
		}

		setPublishedRecoveryChildEvidenceHeaders(ctx, evidence)
		if !c.canRestartInboundRequestFixturePublishedRecovery(nonce) || requestRestart() != nil {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}

		return ctx.Status(fiber.StatusAccepted).Send(nil)
	}
}

func (c *Controller) publishedRecoveryMaplessHTTP(
	evidence PublishedRecoveryChildEvidence,
	admission *inboundRequestFixturePublishedMaplessAdmission,
) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if !c.authorized(ctx) {
			return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
		}

		nonce, ok := qaFixtureMutationNonce(ctx)
		if !ok {
			return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
		}

		setPublishedRecoveryChildEvidenceHeaders(ctx, evidence)
		if err := c.deleteInboundRequestFixturePublishedMapless(ctx.Context(), nonce, admission); err != nil {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}

		return ctx.Status(fiber.StatusNoContent).Send(nil)
	}
}

func setPublishedRecoveryChildEvidenceHeaders(ctx fiber.Ctx, evidence PublishedRecoveryChildEvidence) {
	ctx.Set(publishedRecoveryChildPIDHeader, strconv.Itoa(evidence.ChildPID))
	ctx.Set(publishedRecoveryChildGenerationHeader, strconv.FormatUint(evidence.Generation, 10))
	ctx.Set(publishedRecoveryChildBootSHA256Header, hex.EncodeToString(evidence.BootSHA256[:]))
}
