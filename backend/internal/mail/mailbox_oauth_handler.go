package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const (
	mailboxOAuthHTTPBindingCookiePath = "/api/mail/oauth"
	mailboxOAuthHTTPMaximumJSONBytes = mailboxOAuthCodeMaximum + 1024
	mailboxOAuthHTTPMaximumQueryBytes = 2048
)

type mailboxOAuthHTTPService interface {
	BeginMailboxOAuth(context.Context, MailboxOAuthScope) (MailboxOAuthBeginResult, error)
	CompleteMailboxOAuth(context.Context, MailboxOAuthScope, string, string, string) (MailboxOAuthConnectionResult, error)
	DisconnectMailboxOAuth(context.Context, MailboxOAuthScope) error
}

// MailboxOAuthHandlerConfig fixes browser binding-cookie and input limits.
type MailboxOAuthHandlerConfig struct {
	BindingCookieName      string
	BindingCookieSecure    bool
	BindingCookieMaxAgeSec int
	MaximumJSONBytes       int
	MaximumQueryBytes      int
}

// MailboxOAuthHandler exposes authenticated browser OAuth actions only.
type MailboxOAuthHandler struct {
	service  mailboxOAuthHTTPService
	owners   mailboxOwnerResolver
	statuses MailboxOAuthConnectionStatusSource
	config   MailboxOAuthHandlerConfig
}

type mailboxOAuthAuthorizeResponse struct {
	AuthorizationURL string    `json:"authorization_url"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type mailboxOAuthStatusResponse struct {
	Status               MailboxOAuthConnectionStatus `json:"status"`
	AccessTokenExpiresAt *time.Time                    `json:"access_token_expires_at"`
}

// NewMailboxOAuthHandler creates the isolated OAuth HTTP boundary.
func NewMailboxOAuthHandler(service mailboxOAuthHTTPService, owners mailboxOwnerResolver, statuses MailboxOAuthConnectionStatusSource, config MailboxOAuthHandlerConfig) (*MailboxOAuthHandler, error) {
	if mailboxHTTPDependencyNil(service) || mailboxHTTPDependencyNil(owners) || mailboxHTTPDependencyNil(statuses) || !validMailboxOAuthHandlerConfig(config) {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return &MailboxOAuthHandler{service: service, owners: owners, statuses: statuses, config: config}, nil
}

// RegisterRoutes registers only authenticated routes on the supplied /api router.
func (handler *MailboxOAuthHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(handler.owners) || mailboxHTTPDependencyNil(handler.statuses) || mailboxHTTPDependencyNil(api) || authMiddleware == nil || !validMailboxOAuthHandlerConfig(handler.config) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	oauth := api.Group("/mail/oauth", authMiddleware)
	oauth.Post("/authorize", handler.authorize)
	oauth.Post("/callback", handler.callback)
	oauth.Delete("/connection", handler.disconnect)
	oauth.Get("/status", handler.status)
	return nil
}

func (handler *MailboxOAuthHandler) authorize(ctx fiber.Ctx) error {
	if err := handler.rejectQuery(ctx); err != nil { return err }
	scope, err := handler.resolveScope(ctx)
	if err != nil { return err }
	result, err := handler.service.BeginMailboxOAuth(ctx.Context(), scope)
	if err != nil { return mailboxOAuthHTTPServiceError(ctx, err) }
	if len(result.AuthorizationURL) == 0 || len(result.AuthorizationURL) > mailboxOAuthAuthorizationMaximum || !validMailboxOAuthTime(result.ExpiresAt) || !validMailboxOAuthBase64URL(result.BrowserCookie, mailboxOAuthBrowserBindingEncoded) {
		return mailboxOAuthHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_OAUTH_UNAVAILABLE", "mailbox OAuth unavailable")
	}
	handler.setBindingCookie(ctx, result.BrowserCookie)
	return ctx.JSON(mailboxOAuthAuthorizeResponse{AuthorizationURL: result.AuthorizationURL, ExpiresAt: result.ExpiresAt})
}

func (handler *MailboxOAuthHandler) callback(ctx fiber.Ctx) error {
	defer handler.clearBindingCookie(ctx)
	if err := handler.rejectQuery(ctx); err != nil { return err }
	scope, err := handler.resolveScope(ctx)
	if err != nil { return err }
	request, err := handler.decodeCallback(ctx)
	if err != nil { return err }
	bindingCookie := ctx.Cookies(handler.config.BindingCookieName)
	if !validMailboxOAuthBase64URL(bindingCookie, mailboxOAuthBrowserBindingEncoded) {
		return mailboxOAuthHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_OAUTH_STATE_INVALID", "invalid mailbox OAuth state")
	}
	if _, err := handler.service.CompleteMailboxOAuth(ctx.Context(), scope, request.State, request.Code, bindingCookie); err != nil {
		return mailboxOAuthHTTPServiceError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (handler *MailboxOAuthHandler) disconnect(ctx fiber.Ctx) error {
	defer handler.clearBindingCookie(ctx)
	if err := handler.rejectQuery(ctx); err != nil { return err }
	scope, err := handler.resolveScope(ctx)
	if err != nil { return err }
	if err := handler.service.DisconnectMailboxOAuth(ctx.Context(), scope); err != nil { return mailboxOAuthHTTPServiceError(ctx, err) }
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (handler *MailboxOAuthHandler) status(ctx fiber.Ctx) error {
	if err := handler.rejectQuery(ctx); err != nil { return err }
	scope, engineAccountID, err := handler.resolveStatusScope(ctx)
	if err != nil { return err }
	status, err := handler.statuses.ReadMailboxOAuthConnectionStatus(ctx.Context(), scope, engineAccountID)
	if err != nil { return mailboxOAuthHTTPServiceError(ctx, err) }
	if !validMailboxOAuthConnectionStatusSnapshot(status) {
		return mailboxOAuthHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_OAUTH_UNAVAILABLE", "mailbox OAuth unavailable")
	}
	return ctx.JSON(mailboxOAuthStatusResponse{Status: status.Status, AccessTokenExpiresAt: status.AccessTokenExpiresAt})
}

func (handler *MailboxOAuthHandler) resolveScope(ctx fiber.Ctx) (MailboxOAuthScope, error) {
	scope, _, err := handler.resolveStatusScope(ctx)
	return scope, err
}

func (handler *MailboxOAuthHandler) resolveStatusScope(ctx fiber.Ctx) (MailboxOAuthScope, string, error) {
	actorID, err := mailboxActorID(ctx)
	if err != nil { return MailboxOAuthScope{}, "", mailboxOAuthHTTPFixedError(ctx, fiber.StatusUnauthorized, "MAIL_OAUTH_UNAUTHORIZED", "unauthorized") }
	owner, err := handler.owners.ResolveOwner(ctx.Context(), actorID)
	if err != nil { return MailboxOAuthScope{}, "", mailboxOAuthHTTPOwnerError(ctx, err) }
	scope, engineAccountID, err := mailboxOAuthScopeFromOwner(actorID, owner)
	if err != nil { return MailboxOAuthScope{}, "", mailboxOAuthHTTPOwnerError(ctx, err) }
	return scope, engineAccountID, nil
}

func mailboxOAuthScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (MailboxOAuthScope, string, error) {
	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	if actorID == uuid.Nil || owner.OrganizationID() == uuid.Nil || !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
		return MailboxOAuthScope{}, "", mailboxServiceError(MailboxServiceErrorNotFound)
	}
	if owner.State() == MailboxOwnerDeleting { return MailboxOAuthScope{}, "", mailboxServiceError(MailboxServiceErrorDeleting) }
	if owner.State() != MailboxOwnerActive || projectionState != MailboxProjectionReady || !owner.hasEngineAccountID || !validMailboxOAuthOpaque(owner.engineAccountID, MaxEngineOpaqueIDBytes) || owner.hasRecoveryCode {
		return MailboxOAuthScope{}, "", mailboxServiceError(MailboxServiceErrorNotFound)
	}
	return MailboxOAuthScope{ActorID: actorID, OrganizationID: owner.OrganizationID(), AccountID: accountID}, owner.engineAccountID, nil
}

func (handler *MailboxOAuthHandler) setBindingCookie(ctx fiber.Ctx, value string) {
	ctx.Cookie(&fiber.Cookie{Name: handler.config.BindingCookieName, Value: value, Path: mailboxOAuthHTTPBindingCookiePath, MaxAge: handler.config.BindingCookieMaxAgeSec, Expires: time.Now().UTC().Add(time.Duration(handler.config.BindingCookieMaxAgeSec) * time.Second), Secure: handler.config.BindingCookieSecure, HTTPOnly: true, SameSite: "lax"})
}

func (handler *MailboxOAuthHandler) clearBindingCookie(ctx fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{Name: handler.config.BindingCookieName, Value: "", Path: mailboxOAuthHTTPBindingCookiePath, MaxAge: -1, Expires: time.Unix(1, 0).UTC(), Secure: handler.config.BindingCookieSecure, HTTPOnly: true, SameSite: "lax"})
}

func (handler *MailboxOAuthHandler) rejectQuery(ctx fiber.Ctx) error {
	query := ctx.Request().URI().QueryString()
	if len(query) > handler.config.MaximumQueryBytes { return mailboxOAuthHTTPFixedError(ctx, fiber.StatusRequestURITooLong, "MAIL_OAUTH_QUERY_TOO_LARGE", "invalid mailbox OAuth query") }
	if len(query) != 0 { return mailboxOAuthHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_OAUTH_INVALID_QUERY", "invalid mailbox OAuth query") }
	return nil
}

type mailboxOAuthCallbackRequest struct { State string; Code string }

func (handler *MailboxOAuthHandler) decodeCallback(ctx fiber.Ctx) (mailboxOAuthCallbackRequest, error) {
	contentType, _, err := mime.ParseMediaType(strings.TrimSpace(ctx.Get(fiber.HeaderContentType)))
	if err != nil || !strings.EqualFold(contentType, fiber.MIMEApplicationJSON) { return mailboxOAuthCallbackRequest{}, mailboxOAuthHTTPFixedError(ctx, fiber.StatusUnsupportedMediaType, "MAIL_OAUTH_INVALID_CONTENT_TYPE", "invalid mailbox OAuth content type") }
	body := ctx.BodyRaw()
	if len(body) > handler.config.MaximumJSONBytes { return mailboxOAuthCallbackRequest{}, mailboxOAuthHTTPFixedError(ctx, fiber.StatusRequestEntityTooLarge, "MAIL_OAUTH_BODY_TOO_LARGE", "invalid mailbox OAuth request") }
	request, valid := decodeMailboxOAuthCallbackJSON(body)
	if !valid { return mailboxOAuthCallbackRequest{}, mailboxOAuthHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_OAUTH_INVALID_REQUEST", "invalid mailbox OAuth request") }
	return request, nil
}

func decodeMailboxOAuthCallbackJSON(body []byte) (mailboxOAuthCallbackRequest, bool) {
	if len(body) == 0 || len(body) > mailboxOAuthHTTPMaximumJSONBytes || !utf8.Valid(body) { return mailboxOAuthCallbackRequest{}, false }
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') { return mailboxOAuthCallbackRequest{}, false }
	var request mailboxOAuthCallbackRequest
	seen := make(map[string]struct{}, 2)
	for decoder.More() {
		keyToken, err := decoder.Token()
		key, ok := keyToken.(string)
		if err != nil || !ok { return mailboxOAuthCallbackRequest{}, false }
		if _, duplicate := seen[key]; duplicate { return mailboxOAuthCallbackRequest{}, false }
		seen[key] = struct{}{}
		var value string
		if err := decoder.Decode(&value); err != nil { return mailboxOAuthCallbackRequest{}, false }
		switch key {
		case "state": request.State = value
		case "code": request.Code = value
		default: return mailboxOAuthCallbackRequest{}, false
		}
	}
	token, err = decoder.Token()
	if err != nil || token != json.Delim('}') || decoder.Decode(&struct{}{}) != io.EOF || len(seen) != 2 || !validMailboxOAuthBase64URL(request.State, mailboxOAuthBrowserBindingEncoded) || !validMailboxOAuthOpaque(request.Code, mailboxOAuthCodeMaximum) { return mailboxOAuthCallbackRequest{}, false }
	return request, true
}

func validMailboxOAuthHandlerConfig(config MailboxOAuthHandlerConfig) bool {
	return validMailboxOAuthCookieName(config.BindingCookieName) && config.BindingCookieMaxAgeSec > 0 && config.BindingCookieMaxAgeSec <= int(mailboxOAuthStateTTL/time.Second) && config.MaximumJSONBytes > 0 && config.MaximumJSONBytes <= mailboxOAuthHTTPMaximumJSONBytes && config.MaximumQueryBytes > 0 && config.MaximumQueryBytes <= mailboxOAuthHTTPMaximumQueryBytes
}

func validMailboxOAuthCookieName(value string) bool {
	if len(value) == 0 || len(value) > 64 || !utf8.ValidString(value) { return false }
	for _, character := range []byte(value) {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || character == '_' { continue }
		return false
	}
	return true
}

func validMailboxOAuthConnectionStatusSnapshot(snapshot MailboxOAuthConnectionStatusSnapshot) bool {
	switch snapshot.Status {
	case MailboxOAuthConnectionNotConnected: return snapshot.CredentialGeneration == 0 && snapshot.AccessTokenExpiresAt == nil
	case MailboxOAuthConnectionActive: return snapshot.CredentialGeneration >= 1 && snapshot.AccessTokenExpiresAt != nil && validMailboxOAuthTime(*snapshot.AccessTokenExpiresAt)
	case MailboxOAuthConnectionReconnectRequired, MailboxOAuthConnectionDisconnected: return snapshot.CredentialGeneration >= 1 && snapshot.AccessTokenExpiresAt == nil
	default: return false
	}
}

func mailboxOAuthHTTPOwnerError(ctx fiber.Ctx, err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil { return mailboxOAuthHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_OAUTH_UNAVAILABLE", "mailbox OAuth unavailable") }
	switch serviceError.Kind {
	case MailboxServiceErrorDeleting: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_ACCOUNT_DELETING", "mailbox account deleting")
	case MailboxServiceErrorNotFound: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusNotFound, "MAIL_OAUTH_NOT_FOUND", "mailbox OAuth account not found")
	default: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_OAUTH_UNAVAILABLE", "mailbox OAuth unavailable")
	}
}

func mailboxOAuthHTTPServiceError(ctx fiber.Ctx, err error) error {
	var oauthError *MailboxOAuthError
	if !errors.As(err, &oauthError) || oauthError == nil { return mailboxOAuthHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_OAUTH_UNAVAILABLE", "mailbox OAuth unavailable") }
	switch oauthError.Kind {
	case MailboxOAuthErrorNotFound: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusNotFound, "MAIL_OAUTH_NOT_FOUND", "mailbox OAuth account not found")
	case MailboxOAuthErrorUnauthorized: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusForbidden, "MAIL_OAUTH_FORBIDDEN", "mailbox OAuth unavailable")
	case MailboxOAuthErrorStateInvalid: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_OAUTH_STATE_INVALID", "invalid mailbox OAuth state")
	case MailboxOAuthErrorConflict, MailboxOAuthErrorReconnectRequired: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_OAUTH_CONFLICT", "mailbox OAuth unavailable")
	default: return mailboxOAuthHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_OAUTH_UNAVAILABLE", "mailbox OAuth unavailable")
	}
}

func mailboxOAuthHTTPFixedError(ctx fiber.Ctx, status int, code, message string) error {
	ctx.Set("X-Hub-Error-Code", code)
	return fiber.NewError(status, message)
}