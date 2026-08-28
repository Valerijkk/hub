package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const mailDomainControlHTTPMaximumJSONBytes = 2048

const mailDomainControlHTTPRouteInputLocalKey = "mail_domain_control_http_route_input_v1"

type mailDomainControlHTTPService interface {
	CreateMailDomain(context.Context, MailDomainControlCreateRequest) (MailDomainControlCreateResult, error)
	ListMailDomains(context.Context, MailDomainControlListRequest) ([]MailDomainControlView, error)
	SetDefaultMailDomain(context.Context, MailDomainControlSetDefaultRequest) (MailDomainControlView, error)
}

type mailDomainControlHTTPAdminSource func(fiber.Ctx, uuid.UUID) (bool, error)

// MailDomainControlHandler exposes only the authenticated MailDomain control plane.
type MailDomainControlHandler struct {
	service mailDomainControlHTTPService
	isAdmin mailDomainControlHTTPAdminSource
}

// NewMailDomainControlHandler creates the narrow administrator-only HTTP boundary.
func NewMailDomainControlHandler(service mailDomainControlHTTPService, isAdmin mailDomainControlHTTPAdminSource) (*MailDomainControlHandler, error) {
	if mailboxHTTPDependencyNil(service) || isAdmin == nil {
		return nil, mailDomainControlError(MailDomainControlErrorInvalid)
	}
	return &MailDomainControlHandler{service: service, isAdmin: isAdmin}, nil
}

// RegisterRoutes registers the three safe control-plane routes under auth and admin gates.
func (handler *MailDomainControlHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || handler.isAdmin == nil || mailboxHTTPDependencyNil(api) || authMiddleware == nil {
		return mailDomainControlError(MailDomainControlErrorInvalid)
	}
	routes := api.Group("/admin/mail", authMiddleware)
	routes.Get("/domains", handler.list, handler.preflightList, handler.requireAdmin())
	routes.Post("/domains", handler.create, handler.preflightCreate, handler.requireAdmin())
	routes.Put("/domains/:id/default", handler.setDefault, handler.preflightSetDefault, handler.requireAdmin())
	return nil
}

type mailDomainControlHTTPRoute uint8

const (
	mailDomainControlHTTPRouteList mailDomainControlHTTPRoute = iota + 1
	mailDomainControlHTTPRouteCreate
	mailDomainControlHTTPRouteSetDefault
)

type mailDomainControlHTTPRouteInput struct {
	route          mailDomainControlHTTPRoute
	idempotencyKey uuid.UUID
	domain         string
	domainID       uuid.UUID
}

func (handler *MailDomainControlHandler) preflightList(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	if err := mailDomainControlHTTPRejectEmptyTransport(ctx); err != nil {
		return err
	}
	if mailDomainControlHTTPInvalidResponseWritten(ctx) {
		return nil
	}
	ctx.Locals(mailDomainControlHTTPRouteInputLocalKey, mailDomainControlHTTPRouteInput{route: mailDomainControlHTTPRouteList})
	return ctx.Next()
}

func (handler *MailDomainControlHandler) preflightCreate(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	if err := mailDomainControlHTTPRejectQuery(ctx); err != nil {
		return err
	}
	if mailDomainControlHTTPInvalidResponseWritten(ctx) {
		return nil
	}
	idempotencyKey, err := mailDomainControlHTTPIdempotencyKey(ctx)
	if err != nil {
		return err
	}
	if mailDomainControlHTTPInvalidResponseWritten(ctx) {
		return nil
	}
	request, err := mailDomainControlHTTPDecodeCreate(ctx)
	if err != nil {
		return err
	}
	if mailDomainControlHTTPInvalidResponseWritten(ctx) {
		return nil
	}
	canonicalDomain, ok := canonicalMailDomainControlDomain(request.Domain)
	if !ok || request.Domain != canonicalDomain {
		return mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	ctx.Locals(mailDomainControlHTTPRouteInputLocalKey, mailDomainControlHTTPRouteInput{
		route:          mailDomainControlHTTPRouteCreate,
		idempotencyKey: idempotencyKey,
		domain:         canonicalDomain,
	})
	return ctx.Next()
}

func (handler *MailDomainControlHandler) preflightSetDefault(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	if err := mailDomainControlHTTPRejectEmptyTransport(ctx); err != nil {
		return err
	}
	if mailDomainControlHTTPInvalidResponseWritten(ctx) {
		return nil
	}
	domainID, err := mailDomainControlHTTPCanonicalUUID(ctx.Params("id"))
	if err != nil {
		return mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	ctx.Locals(mailDomainControlHTTPRouteInputLocalKey, mailDomainControlHTTPRouteInput{
		route:    mailDomainControlHTTPRouteSetDefault,
		domainID: domainID,
	})
	return ctx.Next()
}

func (handler *MailDomainControlHandler) requireAdmin() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if handler == nil || handler.isAdmin == nil {
			return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
		}
		actorID, err := mailboxActorID(ctx)
		if err != nil {
			return mailDomainControlHTTPError(ctx, http.StatusUnauthorized, "MAIL_DOMAIN_UNAUTHORIZED")
		}
		isAdmin, err := handler.isAdmin(ctx, actorID)
		if err != nil {
			return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
		}
		if !isAdmin {
			return mailDomainControlHTTPError(ctx, http.StatusForbidden, "MAIL_DOMAIN_FORBIDDEN")
		}
		return ctx.Next()
	}
}

func (handler *MailDomainControlHandler) list(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	if _, ok := mailDomainControlHTTPRouteInputFromContext(ctx, mailDomainControlHTTPRouteList); !ok {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return mailDomainControlHTTPError(ctx, http.StatusUnauthorized, "MAIL_DOMAIN_UNAUTHORIZED")
	}
	views, err := handler.service.ListMailDomains(ctx.Context(), MailDomainControlListRequest{ActorID: actorID})
	if err != nil {
		return mailDomainControlHTTPServiceError(ctx, err)
	}
	if len(views) > mailDomainControlMaximumList {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	domains := make([]mailDomainControlHTTPDomainResponse, 0, len(views))
	for _, view := range views {
		response, ok := mailDomainControlHTTPDomainFromView(view)
		if !ok {
			return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
		}
		domains = append(domains, response)
	}
	return ctx.Status(http.StatusOK).JSON(mailDomainControlHTTPListResponse{Domains: domains})
}

func (handler *MailDomainControlHandler) create(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	input, ok := mailDomainControlHTTPRouteInputFromContext(ctx, mailDomainControlHTTPRouteCreate)
	if !ok {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return mailDomainControlHTTPError(ctx, http.StatusUnauthorized, "MAIL_DOMAIN_UNAUTHORIZED")
	}
	result, err := handler.service.CreateMailDomain(ctx.Context(), MailDomainControlCreateRequest{
		ActorID:        actorID,
		IdempotencyKey: input.idempotencyKey,
		Domain:         input.domain,
	})
	if err != nil {
		return mailDomainControlHTTPServiceError(ctx, err)
	}
	response, ok := mailDomainControlHTTPDomainFromView(result.Domain)
	if !ok {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	return ctx.Status(http.StatusAccepted).JSON(mailDomainControlHTTPSingleResponse{Domain: response})
}

func (handler *MailDomainControlHandler) setDefault(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	input, ok := mailDomainControlHTTPRouteInputFromContext(ctx, mailDomainControlHTTPRouteSetDefault)
	if !ok {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return mailDomainControlHTTPError(ctx, http.StatusUnauthorized, "MAIL_DOMAIN_UNAUTHORIZED")
	}
	view, err := handler.service.SetDefaultMailDomain(ctx.Context(), MailDomainControlSetDefaultRequest{ActorID: actorID, DomainID: input.domainID})
	if err != nil {
		return mailDomainControlHTTPServiceError(ctx, err)
	}
	if !view.IsDefault || !validMailDomainControlVerifiedView(view) {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	response, ok := mailDomainControlHTTPDomainFromView(view)
	if !ok {
		return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
	}
	return ctx.Status(http.StatusOK).JSON(mailDomainControlHTTPSingleResponse{Domain: response})
}

type mailDomainControlHTTPCreateRequest struct {
	Domain string
}

func mailDomainControlHTTPRouteInputFromContext(ctx fiber.Ctx, expectedRoute mailDomainControlHTTPRoute) (mailDomainControlHTTPRouteInput, bool) {
	if ctx == nil {
		return mailDomainControlHTTPRouteInput{}, false
	}
	input, ok := ctx.Locals(mailDomainControlHTTPRouteInputLocalKey).(mailDomainControlHTTPRouteInput)
	if !ok || input.route != expectedRoute {
		return mailDomainControlHTTPRouteInput{}, false
	}
	return input, true
}

type mailDomainControlHTTPDomainResponse struct {
	ID                 string             `json:"id"`
	Domain             string             `json:"domain"`
	State              string             `json:"state"`
	VerificationStatus string             `json:"verification_status"`
	DKIMStatus         string             `json:"dkim_status"`
	RequiredDNSRecords RequiredDNSRecords `json:"required_dns_records"`
	IsDefault          bool               `json:"is_default"`
}

type mailDomainControlHTTPListResponse struct {
	Domains []mailDomainControlHTTPDomainResponse `json:"domains"`
}

type mailDomainControlHTTPSingleResponse struct {
	Domain mailDomainControlHTTPDomainResponse `json:"domain"`
}

type mailDomainControlHTTPErrorResponse struct {
	Error string `json:"error"`
}

func mailDomainControlHTTPDomainFromView(view MailDomainControlView) (mailDomainControlHTTPDomainResponse, bool) {
	if !validMailDomainControlView(view) {
		return mailDomainControlHTTPDomainResponse{}, false
	}
	return mailDomainControlHTTPDomainResponse{
		ID:                 view.ID.String(),
		Domain:             view.Domain,
		State:              view.State,
		VerificationStatus: view.VerificationStatus,
		DKIMStatus:         view.DKIMStatus,
		RequiredDNSRecords: append(RequiredDNSRecords(nil), view.RequiredDNSRecords...),
		IsDefault:          view.IsDefault,
	}, true
}

func mailDomainControlHTTPRejectEmptyTransport(ctx fiber.Ctx) error {
	if err := mailDomainControlHTTPRejectQuery(ctx); err != nil {
		return err
	}
	if ctx == nil || len(ctx.BodyRaw()) != 0 || len(ctx.Request().Header.Peek(fiber.HeaderContentType)) != 0 {
		return mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	return nil
}

func mailDomainControlHTTPInvalidResponseWritten(ctx fiber.Ctx) bool {
	return ctx != nil &&
		ctx.Response().StatusCode() == http.StatusBadRequest &&
		bytes.Equal(ctx.Response().Header.Peek("X-Hub-Error-Code"), []byte("MAIL_DOMAIN_INVALID"))
}

func mailDomainControlHTTPRejectQuery(ctx fiber.Ctx) error {
	if ctx == nil || len(ctx.Request().URI().QueryString()) != 0 {
		return mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	return nil
}

func mailDomainControlHTTPIdempotencyKey(ctx fiber.Ctx) (uuid.UUID, error) {
	value, one := mailboxSubmissionHTTPExactlyOneHeaderValue(ctx, "Idempotency-Key")
	if !one {
		return uuid.Nil, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	idempotencyKey, err := mailDomainControlHTTPCanonicalUUID(string(value))
	if err != nil || idempotencyKey.Version() != 4 && idempotencyKey.Version() != 7 {
		return uuid.Nil, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	return idempotencyKey, nil
}

func mailDomainControlHTTPCanonicalUUID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil || value == uuid.Nil || raw != value.String() {
		return uuid.Nil, mailDomainControlHTTPError(nil, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	return value, nil
}

func mailDomainControlHTTPDecodeCreate(ctx fiber.Ctx) (mailDomainControlHTTPCreateRequest, error) {
	if ctx == nil {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(nil, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	contentType, one := mailboxSubmissionHTTPExactlyOneHeaderValue(ctx, fiber.HeaderContentType)
	if !one {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(string(contentType)))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	rawBody := ctx.BodyRaw()
	if len(rawBody) > mailDomainControlHTTPMaximumJSONBytes {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	decoder := json.NewDecoder(bytes.NewReader(bytes.TrimSpace(rawBody)))
	first, err := decoder.Token()
	if err != nil {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	delimiter, ok := first.(json.Delim)
	if !ok || delimiter != '{' {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	var request mailDomainControlHTTPCreateRequest
	domainSeen := false
	for decoder.More() {
		keyToken, err := decoder.Token()
		key, ok := keyToken.(string)
		if err != nil || !ok || key != "domain" || domainSeen {
			return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
		}
		if err := decoder.Decode(&request.Domain); err != nil {
			return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
		}
		domainSeen = true
	}
	last, err := decoder.Token()
	lastDelimiter, ok := last.(json.Delim)
	if err != nil || !ok || lastDelimiter != '}' || !domainSeen {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return mailDomainControlHTTPCreateRequest{}, mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
	}
	return request, nil
}

func mailDomainControlHTTPServiceError(ctx fiber.Ctx, err error) error {
	var controlError *MailDomainControlError
	if errors.As(err, &controlError) && controlError != nil {
		switch controlError.Kind {
		case MailDomainControlErrorInvalid:
			return mailDomainControlHTTPError(ctx, http.StatusBadRequest, "MAIL_DOMAIN_INVALID")
		case MailDomainControlErrorNotFound:
			return mailDomainControlHTTPError(ctx, http.StatusNotFound, "MAIL_DOMAIN_NOT_FOUND")
		case MailDomainControlErrorConflict:
			return mailDomainControlHTTPError(ctx, http.StatusConflict, "MAIL_DOMAIN_CONFLICT")
		case MailDomainControlErrorUnavailable:
			return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
		}
	}
	return mailDomainControlHTTPError(ctx, http.StatusServiceUnavailable, "MAIL_DOMAIN_UNAVAILABLE")
}

func mailDomainControlHTTPError(ctx fiber.Ctx, status int, code string) error {
	if ctx == nil {
		return fiber.NewError(status, code)
	}
	ctx.Set("X-Hub-Error-Code", code)
	return ctx.Status(status).JSON(mailDomainControlHTTPErrorResponse{Error: code})
}
