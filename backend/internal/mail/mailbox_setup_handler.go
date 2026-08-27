package mail

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const (
	mailboxSetupHTTPBindingCookiePath  = "/api/mail/setup/oauth"
	mailboxSetupHTTPMaximumJSONBytes   = mailboxOAuthHTTPMaximumJSONBytes
	mailboxSetupHTTPMaximumQueryBytes  = mailboxOAuthHTTPMaximumQueryBytes
	mailboxSetupHTTPMaximumPasswordBytes = mailboxPasswordEnrollmentMaximumBytes
)

type mailboxSetupPasswordService interface {
	EnrollPassword(context.Context, uuid.UUID, []byte) (MailboxPasswordEnrollmentResult, error)
}

type mailboxSetupOAuthService interface {
	BeginMailboxBootstrapOAuth(context.Context, MailboxOAuthScope) (MailboxOAuthBeginResult, error)
	CompleteMailboxBootstrapOAuth(context.Context, MailboxOAuthScope, string, string, string) (MailboxOAuthConnectionResult, error)
}

// MailboxSetupHandlerConfig fixes setup-only browser binding-cookie and input limits.
type MailboxSetupHandlerConfig struct {
	BindingCookieName      string
	BindingCookieSecure    bool
	BindingCookieMaxAgeSec int
	MaximumJSONBytes       int
	MaximumQueryBytes      int
	MaximumPasswordBytes   int
}

// MailboxSetupHandler exposes the authenticated password and bootstrap OAuth flow.
type MailboxSetupHandler struct {
	passwords mailboxSetupPasswordService
	oauth     mailboxSetupOAuthService
	owners    mailboxOwnerResolver
	config    MailboxSetupHandlerConfig
}

type mailboxSetupPasswordResponse struct {
	Status MailboxPasswordEnrollmentResult `json:"status"`
}

// NewMailboxSetupHandler creates the setup-only HTTP boundary.
func NewMailboxSetupHandler(passwords mailboxSetupPasswordService, oauth mailboxSetupOAuthService, owners mailboxOwnerResolver, config MailboxSetupHandlerConfig) (*MailboxSetupHandler, error) {
	if mailboxHTTPDependencyNil(passwords) || mailboxHTTPDependencyNil(oauth) || mailboxHTTPDependencyNil(owners) || !validMailboxSetupHandlerConfig(config) {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return &MailboxSetupHandler{passwords: passwords, oauth: oauth, owners: owners, config: config}, nil
}

// RegisterRoutes registers authenticated bootstrap setup routes only.
func (handler *MailboxSetupHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.passwords) || mailboxHTTPDependencyNil(handler.oauth) || mailboxHTTPDependencyNil(handler.owners) || mailboxHTTPDependencyNil(api) || authMiddleware == nil || !validMailboxSetupHandlerConfig(handler.config) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	setup := api.Group("/mail/setup", authMiddleware)
	setup.Post("/password", handler.password)
	setup.Post("/oauth/authorize", handler.authorize)
	setup.Post("/oauth/callback", handler.callback)
	return nil
}

func (handler *MailboxSetupHandler) password(ctx fiber.Ctx) error {
	if err := handler.rejectQuery(ctx); err != nil {
		return err
	}
	if !strings.EqualFold(ctx.Get(fiber.HeaderContentType), "application/octet-stream") {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusUnsupportedMediaType, "MAIL_SETUP_INVALID_CONTENT_TYPE", "invalid mailbox setup content type")
	}
	body := ctx.BodyRaw()
	defer clear(body)
	if len(body) > handler.config.MaximumPasswordBytes {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusRequestEntityTooLarge, "MAIL_SETUP_PASSWORD_TOO_LARGE", "invalid mailbox setup password")
	}
	if len(body) < mailboxPasswordEnrollmentMinimumBytes {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_INVALID_PASSWORD", "invalid mailbox setup password")
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusUnauthorized, "MAIL_SETUP_UNAUTHORIZED", "unauthorized")
	}
	scope, err := handler.resolveBootstrapScope(ctx)
	if err != nil {
		return err
	}
	if scope.ActorID != actorID {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_SETUP_NOT_READY", "mailbox setup not ready")
	}
	password := append([]byte(nil), body...)
	defer clear(password)
	result, err := handler.passwords.EnrollPassword(ctx.Context(), actorID, password)
	if err != nil {
		return mailboxSetupHTTPPasswordError(ctx, err)
	}
	switch result {
	case MailboxPasswordEnrollmentResultEnrolled, MailboxPasswordEnrollmentResultAlreadyEnrolled:
		return ctx.JSON(mailboxSetupPasswordResponse{Status: result})
	case MailboxPasswordEnrollmentResultBusy:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_SETUP_BUSY", "mailbox setup busy")
	default:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_INVALID_PASSWORD", "invalid mailbox setup password")
	}
}

func (handler *MailboxSetupHandler) authorize(ctx fiber.Ctx) error {
	if err := handler.rejectQuery(ctx); err != nil {
		return err
	}
	scope, err := handler.resolveBootstrapScope(ctx)
	if err != nil {
		return err
	}
	result, err := handler.oauth.BeginMailboxBootstrapOAuth(ctx.Context(), scope)
	if err != nil {
		return mailboxSetupHTTPOAuthError(ctx, err)
	}
	if len(result.AuthorizationURL) == 0 || len(result.AuthorizationURL) > mailboxOAuthAuthorizationMaximum || !validMailboxOAuthTime(result.ExpiresAt) || !validMailboxOAuthBase64URL(result.BrowserCookie, mailboxOAuthBrowserBindingEncoded) {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	}
	handler.setBindingCookie(ctx, result.BrowserCookie)
	return ctx.JSON(mailboxOAuthAuthorizeResponse{AuthorizationURL: result.AuthorizationURL, ExpiresAt: result.ExpiresAt})
}

func (handler *MailboxSetupHandler) callback(ctx fiber.Ctx) error {
	defer handler.clearBindingCookie(ctx)
	if err := handler.rejectQuery(ctx); err != nil {
		return err
	}
	scope, err := handler.resolveBootstrapScope(ctx)
	if err != nil {
		return err
	}
	request, err := handler.decodeCallback(ctx)
	if err != nil {
		return err
	}
	bindingCookie := ctx.Cookies(handler.config.BindingCookieName)
	if !validMailboxOAuthBase64URL(bindingCookie, mailboxOAuthBrowserBindingEncoded) {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_STATE_INVALID", "invalid mailbox setup state")
	}
	if _, err := handler.oauth.CompleteMailboxBootstrapOAuth(ctx.Context(), scope, request.State, request.Code, bindingCookie); err != nil {
		return mailboxSetupHTTPOAuthError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (handler *MailboxSetupHandler) resolveBootstrapScope(ctx fiber.Ctx) (MailboxOAuthScope, error) {
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return MailboxOAuthScope{}, mailboxSetupHTTPFixedError(ctx, fiber.StatusUnauthorized, "MAIL_SETUP_UNAUTHORIZED", "unauthorized")
	}
	owner, err := handler.owners.ResolveOwner(ctx.Context(), actorID)
	if err != nil {
		return MailboxOAuthScope{}, mailboxSetupHTTPOwnerError(ctx, err)
	}
	scope, err := mailboxSetupScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxOAuthScope{}, mailboxSetupHTTPOwnerError(ctx, err)
	}
	return scope, nil
}

func mailboxSetupScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (MailboxOAuthScope, error) {
	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	primaryAddress, hasPrimaryAddress := owner.PrimaryAddress()
	if actorID == uuid.Nil || owner.OrganizationID() == uuid.Nil || !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
		return MailboxOAuthScope{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	if owner.State() == MailboxOwnerDeleting {
		return MailboxOAuthScope{}, mailboxServiceError(MailboxServiceErrorDeleting)
	}
	if owner.State() != MailboxOwnerProvisioning || projectionState != MailboxProjectionBootstrapping || !owner.hasEngineAccountID || !validMailboxOAuthOpaque(owner.engineAccountID, MaxEngineOpaqueIDBytes) || !hasPrimaryAddress || !validMailboxOwnerPrimaryAddress(primaryAddress) || owner.hasRecoveryCode {
		return MailboxOAuthScope{}, mailboxServiceError(MailboxServiceErrorNotFound)
	}
	return MailboxOAuthScope{ActorID: actorID, OrganizationID: owner.OrganizationID(), AccountID: accountID}, nil
}

func (handler *MailboxSetupHandler) setBindingCookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{Name: handler.config.BindingCookieName, Value: value, Path: mailboxSetupHTTPBindingCookiePath, MaxAge: handler.config.BindingCookieMaxAgeSec, Expires: time.Now().UTC().Add(time.Duration(handler.config.BindingCookieMaxAgeSec) * time.Second), Secure: handler.config.BindingCookieSecure, HTTPOnly: true, SameSite: "lax"})
}

func (handler *MailboxSetupHandler) clearBindingCookie(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{Name: handler.config.BindingCookieName, Value: "", Path: mailboxSetupHTTPBindingCookiePath, MaxAge: -1, Expires: time.Unix(1, 0).UTC(), Secure: handler.config.BindingCookieSecure, HTTPOnly: true, SameSite: "lax"})
}

func (handler *MailboxSetupHandler) rejectQuery(ctx fiber.Ctx) error {
	query := ctx.Request().URI().QueryString()
	if len(query) > handler.config.MaximumQueryBytes {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusRequestURITooLong, "MAIL_SETUP_QUERY_TOO_LARGE", "invalid mailbox setup query")
	}
	if len(query) != 0 {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_INVALID_QUERY", "invalid mailbox setup query")
	}
	return nil
}

func (handler *MailboxSetupHandler) decodeCallback(ctx fiber.Ctx) (mailboxOAuthCallbackRequest, error) {
	if !strings.EqualFold(ctx.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		return mailboxOAuthCallbackRequest{}, mailboxSetupHTTPFixedError(ctx, fiber.StatusUnsupportedMediaType, "MAIL_SETUP_INVALID_CONTENT_TYPE", "invalid mailbox setup content type")
	}
	body := ctx.BodyRaw()
	if len(body) > handler.config.MaximumJSONBytes {
		return mailboxOAuthCallbackRequest{}, mailboxSetupHTTPFixedError(ctx, fiber.StatusRequestEntityTooLarge, "MAIL_SETUP_BODY_TOO_LARGE", "invalid mailbox setup request")
	}
	request, valid := decodeMailboxOAuthCallbackJSON(body)
	if !valid {
		return mailboxOAuthCallbackRequest{}, mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_INVALID_REQUEST", "invalid mailbox setup request")
	}
	return request, nil
}

func validMailboxSetupHandlerConfig(config MailboxSetupHandlerConfig) bool {
	return validMailboxOAuthCookieName(config.BindingCookieName) && config.BindingCookieMaxAgeSec > 0 && config.BindingCookieMaxAgeSec <= int(mailboxOAuthStateTTL/time.Second) && config.MaximumJSONBytes > 0 && config.MaximumJSONBytes <= mailboxSetupHTTPMaximumJSONBytes && config.MaximumQueryBytes > 0 && config.MaximumQueryBytes <= mailboxSetupHTTPMaximumQueryBytes && config.MaximumPasswordBytes == mailboxSetupHTTPMaximumPasswordBytes
}

func mailboxSetupHTTPPasswordError(ctx fiber.Ctx, err error) error {
	var passwordError *MailboxPasswordEnrollmentError
	if !errors.As(err, &passwordError) || passwordError == nil {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	}
	switch passwordError.Kind() {
	case MailboxPasswordEnrollmentErrorInvalid:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_INVALID_PASSWORD", "invalid mailbox setup password")
	case MailboxPasswordEnrollmentErrorRepositoryUnavailable, MailboxPasswordEnrollmentErrorRetryable:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	default:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	}
}

func mailboxSetupHTTPOwnerError(ctx fiber.Ctx, err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	}
	switch serviceError.Kind {
	case MailboxServiceErrorDeleting:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_ACCOUNT_DELETING", "mailbox account deleting")
	case MailboxServiceErrorInvalid, MailboxServiceErrorNotFound:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_SETUP_NOT_READY", "mailbox setup not ready")
	default:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	}
}

func mailboxSetupHTTPOAuthError(ctx fiber.Ctx, err error) error {
	var oauthError *MailboxOAuthError
	if !errors.As(err, &oauthError) || oauthError == nil {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	}
	switch oauthError.Kind {
	case MailboxOAuthErrorInvalid:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_INVALID_REQUEST", "invalid mailbox setup request")
	case MailboxOAuthErrorStateInvalid:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_STATE_INVALID", "invalid mailbox setup state")
	case MailboxOAuthErrorNotFound:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_SETUP_NOT_READY", "mailbox setup not ready")
	case MailboxOAuthErrorUnauthorized:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusForbidden, "MAIL_SETUP_FORBIDDEN", "mailbox setup unavailable")
	case MailboxOAuthErrorConflict, MailboxOAuthErrorReconnectRequired:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_SETUP_CONFLICT", "mailbox setup conflict")
	default:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
	}
}

func mailboxSetupHTTPFixedError(ctx fiber.Ctx, status int, code, message string) error {
	ctx.Set("X-Hub-Error-Code", code)
	return fiber.NewError(status, message)
}
