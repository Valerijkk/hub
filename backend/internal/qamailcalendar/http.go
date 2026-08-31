//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v3"
)

const (
	qaFixtureHTTPErrorUnauthorized = "unauthorized"
	qaFixtureHTTPErrorBadRequest   = "invalid request"
	qaFixtureHTTPErrorInternal     = "internal server error"
	qaFixtureHTTPErrorConflict     = "fixture lifecycle conflict"
	qaFixtureHTTPDeleteGuardHeader = "X-Hub-QA-Fixture-Delete-Guard"
)

type qaFixtureHTTPErrorResponse struct {
	Error string `json:"error"`
}

type qaFixtureHTTPResponse struct {
	Nonce           string `json:"nonce"`
	Scenario        string `json:"scenario"`
	State           string `json:"state"`
	OwnerEmail      string `json:"owner_email"`
	OrganizationID  string `json:"organization_id"`
	OwnerUserID     string `json:"owner_user_id"`
	MailAccountID   string `json:"mail_account_id"`
	MailFolderID    string `json:"mail_folder_id"`
	MailThreadID    string `json:"mail_thread_id"`
	MailMessageID   string `json:"mail_message_id"`
	EnvelopeEventID string `json:"envelope_event_id"`
}

type qaFixtureHTTPThreadSearchResponse struct {
	Nonce             string    `json:"nonce"`
	Scenario          string    `json:"scenario"`
	State             string    `json:"state"`
	OwnerEmail        string    `json:"owner_email"`
	OrganizationID    string    `json:"organization_id"`
	OwnerUserID       string    `json:"owner_user_id"`
	MailAccountID     string    `json:"mail_account_id"`
	MailFolderID      string    `json:"mail_folder_id"`
	MatchingThreadIDs [4]string `json:"matching_thread_ids"`
	DecoyThreadIDs    [2]string `json:"decoy_thread_ids"`
}

type qaFixtureHTTPAccountMaterializationResponse struct {
	Nonce                  string  `json:"nonce"`
	Scenario               string  `json:"scenario"`
	State                  string  `json:"state"`
	OwnerEmail             string  `json:"owner_email"`
	OrganizationID         string  `json:"organization_id"`
	OwnerUserID            string  `json:"owner_user_id"`
	MailDomainID           string  `json:"mail_domain_id"`
	ExpectedMailAccountID  string  `json:"expected_mail_account_id"`
	MaterializationState   string  `json:"materialization_state"`
	MailAccountID          *string `json:"mail_account_id"`
	PrimaryIdentityID      *string `json:"primary_identity_id"`
	PrimaryIdentityStatus  *string `json:"primary_identity_status"`
}

type qaFixtureHTTPAccountMaterializationRecoveryResponse struct {
	Outcome string `json:"outcome"`
}

// RegisterRoutes exposes the closed QA-only Mail/Calendar fixture surface on /api.
func (c *Controller) RegisterRoutes(api fiber.Router) error {
	if c == nil || api == nil {
		return errInboundRequestFixtureControllerInput
	}

	api.Post("/_qa/mail-calendar/fixtures", c.create)
	api.Post("/_qa/mail-calendar/fixtures/:nonce/publish", c.publish)
	api.Post("/_qa/mail-calendar/fixtures/:nonce/thread-search-delete-blocker/:kind", c.armThreadSearchDeleteBlockerHTTP)
	api.Delete("/_qa/mail-calendar/fixtures/:nonce/thread-search-delete-blocker/:kind", c.disarmThreadSearchDeleteBlockerHTTP)
	api.Post("/_qa/mail-calendar/fixtures/:nonce/thread-search-delete-blocker/:kind/peer/:peer_nonce", c.armThreadSearchDeletePeerBlockerHTTP)
	api.Delete("/_qa/mail-calendar/fixtures/:nonce/thread-search-delete-blocker/:kind/peer/:peer_nonce", c.disarmThreadSearchDeletePeerBlockerHTTP)
	api.Delete("/_qa/mail-calendar/fixtures/:nonce", c.delete)
	api.Get("/_qa/mail-calendar/fixtures/:nonce/account-materialization", c.observeAccountMaterialization)
	api.Post("/_qa/mail-calendar/fixtures/:nonce/account-materialization/quiesce", c.quiesceAccountMaterialization)
	api.Delete("/_qa/mail-calendar/fixtures/:nonce/account-materialization", c.deleteAccountMaterialization)
	api.Post("/_qa/mail-calendar/fixtures/:nonce/account-materialization/recover", c.recoverAccountMaterialization)
	return nil
}

func (c *Controller) observeAccountMaterialization(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	snapshot, err := c.observeAccountMaterializationFixture(ctx.Context(), nonce)
	if err != nil {
		if err == errAccountMaterializationFixtureLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}

		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}

	return ctx.Status(fiber.StatusOK).JSON(qaFixtureHTTPAccountMaterializationSnapshot(snapshot))
}

func (c *Controller) quiesceAccountMaterialization(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}
	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}
	if err := c.quiesceAccountMaterializationFixture(ctx.Context(), nonce); err != nil {
		if err == errAccountMaterializationFixtureLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}
		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controller) deleteAccountMaterialization(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	if err := c.deleteAccountMaterializationFixture(ctx.Context(), nonce); err != nil {
		if err == errAccountMaterializationFixtureLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}

		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func (c *Controller) recoverAccountMaterialization(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	outcome, err := c.recoverAccountMaterializationFixture(ctx.Context(), nonce)
	if err != nil {
		if err == errAccountMaterializationFixtureLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}

		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}

	return ctx.Status(fiber.StatusOK).JSON(qaFixtureHTTPAccountMaterializationRecoveryOutcome(outcome))
}

func (c *Controller) create(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	scenario, ok := qaFixtureCreateScenario(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	switch scenario {
	case ScenarioInboundRequestV1:
		snapshot, err := c.createInboundRequestFixture(ctx.Context())
		if err != nil {
			return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
		}
		return ctx.Status(fiber.StatusCreated).JSON(qaFixtureHTTPSnapshot(snapshot))
	case ScenarioAccountMaterializationV1:
		snapshot, err := c.createAccountMaterializationFixture(ctx.Context())
		if err != nil {
			return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
		}
		return ctx.Status(fiber.StatusCreated).JSON(qaFixtureHTTPAccountMaterializationSnapshot(snapshot))
	case ScenarioThreadSearchV1:
		snapshot, err := c.createThreadSearchFixture(ctx.Context())
		if err != nil {
			return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
		}
		return ctx.Status(fiber.StatusCreated).JSON(qaFixtureHTTPThreadSearchSnapshot(snapshot))
	default:
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}
}

func (c *Controller) publish(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	snapshot, err := c.publishInboundRequestFixture(ctx.Context(), nonce)
	if err != nil {
		return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
	}
	return ctx.Status(fiber.StatusOK).JSON(qaFixtureHTTPSnapshot(snapshot))
}

func (c *Controller) delete(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	if err := c.deleteFixture(ctx.Context(), nonce); err != nil {
		if err == errThreadSearchFixtureDeleteGuardDependent {
			ctx.Set(qaFixtureHTTPDeleteGuardHeader, "predelete")
		}
		return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
	}
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

func (c *Controller) armThreadSearchDeleteBlockerHTTP(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	kind := ctx.Params("kind")
	if _, ok := threadSearchDeleteBlockerSingleKind(kind); !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	if err := c.armThreadSearchDeleteBlocker(ctx.Context(), nonce, kind); err != nil {
		if err == errThreadSearchDeleteBlockerInput {
			return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
		}
		if err == errThreadSearchDeleteBlockerLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}
		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

func (c *Controller) disarmThreadSearchDeleteBlockerHTTP(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	kind := ctx.Params("kind")
	if _, ok := threadSearchDeleteBlockerSingleKind(kind); !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	if err := c.disarmThreadSearchDeleteBlocker(ctx.Context(), nonce, kind); err != nil {
		if err == errThreadSearchDeleteBlockerInput {
			return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
		}
		if err == errThreadSearchDeleteBlockerLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}
		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

func (c *Controller) armThreadSearchDeletePeerBlockerHTTP(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	kind := ctx.Params("kind")
	if _, ok := threadSearchDeleteBlockerPeerKind(kind); !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	peerNonce := ctx.Params("peer_nonce")
	if !ValidNonce(peerNonce) || peerNonce == nonce {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	if err := c.armThreadSearchDeletePeerBlocker(ctx.Context(), nonce, peerNonce, kind); err != nil {
		if err == errThreadSearchDeleteBlockerInput {
			return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
		}
		if err == errThreadSearchDeleteBlockerLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}
		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

func (c *Controller) disarmThreadSearchDeletePeerBlockerHTTP(ctx fiber.Ctx) error {
	if !c.authorized(ctx) {
		return qaFixtureHTTPError(ctx, fiber.StatusUnauthorized, qaFixtureHTTPErrorUnauthorized)
	}

	nonce, ok := qaFixtureMutationNonce(ctx)
	if !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	kind := ctx.Params("kind")
	if _, ok := threadSearchDeleteBlockerPeerKind(kind); !ok {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	peerNonce := ctx.Params("peer_nonce")
	if !ValidNonce(peerNonce) || peerNonce == nonce {
		return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
	}

	if err := c.disarmThreadSearchDeletePeerBlocker(ctx.Context(), nonce, peerNonce, kind); err != nil {
		if err == errThreadSearchDeleteBlockerInput {
			return qaFixtureHTTPError(ctx, fiber.StatusBadRequest, qaFixtureHTTPErrorBadRequest)
		}
		if err == errThreadSearchDeleteBlockerLifecycle {
			return qaFixtureHTTPError(ctx, fiber.StatusConflict, qaFixtureHTTPErrorConflict)
		}
		return qaFixtureHTTPError(ctx, fiber.StatusInternalServerError, qaFixtureHTTPErrorInternal)
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

func (c *Controller) authorized(ctx fiber.Ctx) bool {
	return c.config.Authorized(string(ctx.Request().Header.Peek(HeaderHubQAToken)))
}

func qaFixtureCreateScenario(ctx fiber.Ctx) (Scenario, bool) {
	if len(ctx.Request().URI().QueryString()) != 0 ||
		!bytes.Equal(ctx.Request().Header.Peek(fiber.HeaderContentType), []byte(fiber.MIMEApplicationJSON)) {
		return "", false
	}

	body := ctx.BodyRaw()
	if len(body) == 0 || len(body) > 128 {
		return "", false
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	start, err := decoder.Token()
	if err != nil {
		return "", false
	}
	startObject, ok := start.(json.Delim)
	if !ok || startObject != '{' {
		return "", false
	}

	field, err := decoder.Token()
	if err != nil {
		return "", false
	}
	name, ok := field.(string)
	if !ok || name != "scenario" {
		return "", false
	}

	var value string
	if err := decoder.Decode(&value); err != nil {
		return "", false
	}

	end, err := decoder.Token()
	if err != nil {
		return "", false
	}
	endObject, ok := end.(json.Delim)
	if !ok || endObject != '}' {
		return "", false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return "", false
	}

	scenario, err := ParseScenario(value)
	if err != nil {
		return "", false
	}
	return scenario, true
}

func qaFixtureMutationNonce(ctx fiber.Ctx) (string, bool) {
	if len(ctx.Request().URI().QueryString()) != 0 ||
		len(ctx.BodyRaw()) != 0 ||
		len(ctx.Request().Header.Peek(fiber.HeaderContentType)) != 0 {
		return "", false
	}

	nonce := ctx.Params("nonce")
	if !ValidNonce(nonce) {
		return "", false
	}
	return nonce, true
}

func qaFixtureHTTPSnapshot(snapshot inboundRequestFixtureSnapshot) qaFixtureHTTPResponse {
	return qaFixtureHTTPResponse{
		Nonce:           snapshot.nonce,
		Scenario:        string(snapshot.scenario),
		State:           string(snapshot.state),
		OwnerEmail:      snapshot.ownerEmail,
		OrganizationID:  snapshot.organizationID.String(),
		OwnerUserID:     snapshot.ownerUserID.String(),
		MailAccountID:   snapshot.mailAccountID.String(),
		MailFolderID:    snapshot.mailFolderID.String(),
		MailThreadID:    snapshot.mailThreadID.String(),
		MailMessageID:   snapshot.mailMessageID.String(),
		EnvelopeEventID: snapshot.envelopeEventID.String(),
	}
}

func qaFixtureHTTPThreadSearchSnapshot(snapshot threadSearchFixtureSnapshot) qaFixtureHTTPThreadSearchResponse {
	response := qaFixtureHTTPThreadSearchResponse{
		Nonce:          snapshot.nonce,
		Scenario:       string(snapshot.scenario),
		State:          string(snapshot.state),
		OwnerEmail:     snapshot.ownerEmail,
		OrganizationID: snapshot.organizationID.String(),
		OwnerUserID:    snapshot.ownerUserID.String(),
		MailAccountID:  snapshot.mailAccountID.String(),
		MailFolderID:   snapshot.mailFolderID.String(),
	}
	for index, id := range snapshot.matchingThreadIDs {
		response.MatchingThreadIDs[index] = id.String()
	}
	for index, id := range snapshot.decoyThreadIDs {
		response.DecoyThreadIDs[index] = id.String()
	}
	return response
}

func qaFixtureHTTPAccountMaterializationSnapshot(snapshot accountMaterializationFixtureSnapshot) qaFixtureHTTPAccountMaterializationResponse {
	var mailAccountID *string
	if snapshot.mailAccountID != nil {
		value := snapshot.mailAccountID.String()
		mailAccountID = &value
	}

	var primaryIdentityID *string
	if snapshot.primaryIdentityID != nil {
		value := snapshot.primaryIdentityID.String()
		primaryIdentityID = &value
	}

	var primaryIdentityStatus *string
	if snapshot.primaryIdentityStatus != nil {
		value := *snapshot.primaryIdentityStatus
		primaryIdentityStatus = &value
	}

	return qaFixtureHTTPAccountMaterializationResponse{
		Nonce:                 snapshot.nonce,
		Scenario:              string(snapshot.scenario),
		State:                 string(snapshot.state),
		OwnerEmail:            snapshot.ownerEmail,
		OrganizationID:        snapshot.organizationID.String(),
		OwnerUserID:           snapshot.ownerUserID.String(),
		MailDomainID:          snapshot.mailDomainID.String(),
		ExpectedMailAccountID: snapshot.expectedMailAccountID.String(),
		MaterializationState:  string(snapshot.materializationState),
		MailAccountID:         mailAccountID,
		PrimaryIdentityID:     primaryIdentityID,
		PrimaryIdentityStatus: primaryIdentityStatus,
	}
}

func qaFixtureHTTPAccountMaterializationRecoveryOutcome(outcome accountMaterializationFixtureRecoveryOutcome) qaFixtureHTTPAccountMaterializationRecoveryResponse {
	return qaFixtureHTTPAccountMaterializationRecoveryResponse{Outcome: string(outcome)}
}

func qaFixtureHTTPError(ctx fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(qaFixtureHTTPErrorResponse{Error: message})
}
