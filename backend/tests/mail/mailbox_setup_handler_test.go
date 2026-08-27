package mail_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxSetupHandler_ConstructorAndAuthenticatedRoutes(t *testing.T) {
	fixture := newMailboxSetupHTTPFixture(t)

	request := httptest.NewRequest(http.MethodPost, "/api/mail/setup/password", bytes.NewReader([]byte("twelve-bytes")))
	request.Header.Set("Content-Type", "application/octet-stream")
	response := mailboxSetupHTTPResponse(t, fixture.app, request)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
	if fixture.password.calls != 0 || fixture.oauth.beginCalls != 0 || fixture.oauth.completeCalls != 0 {
		t.Fatalf("unauthenticated service calls password/begin/complete = %d/%d/%d, want 0/0/0", fixture.password.calls, fixture.oauth.beginCalls, fixture.oauth.completeCalls)
	}

	if _, err := mail.NewMailboxSetupHandler(nil, fixture.oauth, fixture.owners, fixture.config); err == nil {
		t.Fatal("NewMailboxSetupHandler(nil, ...) error = nil, want invalid dependency error")
	}
	invalidConfig := fixture.config
	invalidConfig.MaximumPasswordBytes--
	if _, err := mail.NewMailboxSetupHandler(fixture.password, fixture.oauth, fixture.owners, invalidConfig); err == nil {
		t.Fatal("NewMailboxSetupHandler(... invalid config) error = nil, want invalid configuration error")
	}
}

func TestMailboxSetupHandler_PasswordRejectsTransportAndBounds(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		path        string
		contentType string
		body        []byte
		wantStatus  int
		wantCode    string
	}{
		{name: "query", path: "/api/mail/setup/password?unexpected=1", contentType: "application/octet-stream", body: []byte("twelve-bytes"), wantStatus: http.StatusBadRequest, wantCode: "MAIL_SETUP_INVALID_QUERY"},
		{name: "large query", path: "/api/mail/setup/password?" + strings.Repeat("x", 2049), contentType: "application/octet-stream", body: []byte("twelve-bytes"), wantStatus: http.StatusRequestURITooLong, wantCode: "MAIL_SETUP_QUERY_TOO_LARGE"},
		{name: "json", path: "/api/mail/setup/password", contentType: "application/json", body: []byte(`{"password":"twelve-bytes"}`), wantStatus: http.StatusUnsupportedMediaType, wantCode: "MAIL_SETUP_INVALID_CONTENT_TYPE"},
		{name: "charset", path: "/api/mail/setup/password", contentType: "application/octet-stream; charset=utf-8", body: []byte("twelve-bytes"), wantStatus: http.StatusUnsupportedMediaType, wantCode: "MAIL_SETUP_INVALID_CONTENT_TYPE"},
		{name: "short", path: "/api/mail/setup/password", contentType: "application/octet-stream", body: bytes.Repeat([]byte{'a'}, 11), wantStatus: http.StatusBadRequest, wantCode: "MAIL_SETUP_INVALID_PASSWORD"},
		{name: "large", path: "/api/mail/setup/password", contentType: "application/octet-stream", body: bytes.Repeat([]byte{'a'}, 257), wantStatus: http.StatusRequestEntityTooLarge, wantCode: "MAIL_SETUP_PASSWORD_TOO_LARGE"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newMailboxSetupHTTPFixture(t)
			request := httptest.NewRequest(http.MethodPost, testCase.path, bytes.NewReader(testCase.body))
			request.Header.Set("Content-Type", testCase.contentType)
			request.Header.Set("X-Test-Auth", "yes")
			response := mailboxSetupHTTPResponse(t, fixture.app, request)
			if response.StatusCode != testCase.wantStatus || response.Header.Get("X-Hub-Error-Code") != testCase.wantCode {
				t.Fatalf("status/code = %d/%q, want %d/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), testCase.wantStatus, testCase.wantCode)
			}
			if fixture.password.calls != 0 {
				t.Fatalf("password calls = %d, want 0", fixture.password.calls)
			}
		})
	}
}

func TestMailboxSetupHandler_PasswordCopiesClearsAndReturnsClosedOutcome(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		result     mail.MailboxPasswordEnrollmentResult
		serviceErr error
		wantStatus int
		wantCode   string
		wantBody   string
	}{
		{name: "enrolled", result: mail.MailboxPasswordEnrollmentResultEnrolled, wantStatus: http.StatusOK, wantBody: `{"status":"enrolled"}`},
		{name: "already enrolled", result: mail.MailboxPasswordEnrollmentResultAlreadyEnrolled, wantStatus: http.StatusOK, wantBody: `{"status":"already_enrolled"}`},
		{name: "busy", result: mail.MailboxPasswordEnrollmentResultBusy, wantStatus: http.StatusConflict, wantCode: "MAIL_SETUP_BUSY"},
		{name: "invalid outcome", wantStatus: http.StatusBadRequest, wantCode: "MAIL_SETUP_INVALID_PASSWORD"},
		{name: "unknown failure", serviceErr: errors.New("registry detail must stay private"), wantStatus: http.StatusServiceUnavailable, wantCode: "MAIL_SETUP_UNAVAILABLE"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newMailboxSetupHTTPFixture(t)
			fixture.password.result = testCase.result
			fixture.password.err = testCase.serviceErr
			secret := []byte{'r', 'a', 'w', 0, 'p', 'a', 's', 's', 'w', 'o', 'r', 'd'}
			request := httptest.NewRequest(http.MethodPost, "/api/mail/setup/password", bytes.NewReader(secret))
			request.Header.Set("Content-Type", "application/octet-stream")
			request.Header.Set("X-Test-Auth", "yes")
			response := mailboxSetupHTTPResponse(t, fixture.app, request)
			body := mailboxSetupBody(t, response)
			if response.StatusCode != testCase.wantStatus || response.Header.Get("X-Hub-Error-Code") != testCase.wantCode {
				t.Fatalf("status/code = %d/%q, want %d/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), testCase.wantStatus, testCase.wantCode)
			}
			if testCase.wantBody != "" && string(body) != testCase.wantBody {
				t.Fatalf("response body = %q, want %q", body, testCase.wantBody)
			}
			if fixture.password.calls != 1 || !bytes.Equal(fixture.password.passwordAtCall, secret) || !mailboxSetupAllZero(fixture.password.retainedPassword) {
				t.Fatalf("password calls/value/cleared = %d/%q/%v, want 1/%q/true", fixture.password.calls, fixture.password.passwordAtCall, fixture.password.retainedPassword, secret)
			}
			if bytes.Contains(body, secret) || strings.Contains(string(body), "registry detail") {
				t.Fatalf("response leaked password or service detail: %q", body)
			}
		})
	}
}

func TestMailboxSetupHandler_PasswordRequiresBootstrapOwner(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mailboxSetupHTTPFixture)
	}{
		{name: "projection ready", mutate: func(fixture *mailboxSetupHTTPFixture) {
			fixture.ownerRecord.ProjectionState = mail.MailboxProjectionReady
		}},
		{name: "recovery code", mutate: func(fixture *mailboxSetupHTTPFixture) {
			code := mail.MailboxOwnerRecoveryCode("MAILBOX_PROVISIONING_RETRY")
			fixture.ownerRecord.RecoveryCode = &code
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newMailboxSetupHTTPFixture(t)
			testCase.mutate(fixture)
			fixture.owners = mailboxSetupOwnerService(t, fixture.ownerRecord)
			fixture.replaceHandler(t)
			request := httptest.NewRequest(http.MethodPost, "/api/mail/setup/password", bytes.NewReader([]byte("twelve-bytes")))
			request.Header.Set("Content-Type", "application/octet-stream")
			request.Header.Set("X-Test-Auth", "yes")
			response := mailboxSetupHTTPResponse(t, fixture.app, request)
			if response.StatusCode != http.StatusConflict || response.Header.Get("X-Hub-Error-Code") != "MAIL_SETUP_NOT_READY" || fixture.password.calls != 0 {
				t.Fatalf("password owner gate status/code/calls = %d/%q/%d, want 409/%q/0", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), fixture.password.calls, "MAIL_SETUP_NOT_READY")
			}
		})
	}
}

func TestMailboxSetupHandler_AuthorizeRequiresBootstrapOwnerAndSetsScopedCookie(t *testing.T) {
	fixture := newMailboxSetupHTTPFixture(t)
	request := httptest.NewRequest(http.MethodPost, "/api/mail/setup/oauth/authorize", nil)
	request.Header.Set("X-Test-Auth", "yes")
	response := mailboxSetupHTTPResponse(t, fixture.app, request)
	body := mailboxSetupBody(t, response)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("authorize status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	var payload struct {
		AuthorizationURL string    `json:"authorization_url"`
		ExpiresAt        time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.AuthorizationURL != fixture.oauth.begin.AuthorizationURL || !payload.ExpiresAt.Equal(fixture.oauth.begin.ExpiresAt) {
		t.Fatalf("authorize payload = %#v, %v, want bounded begin result", payload, err)
	}
	if fixture.oauth.beginCalls != 1 || fixture.oauth.beginScope != fixture.scope {
		t.Fatalf("begin calls/scope = %d/%#v, want 1/%#v", fixture.oauth.beginCalls, fixture.oauth.beginScope, fixture.scope)
	}
	cookie := response.Header.Get("Set-Cookie")
	for _, fragment := range []string{fixture.config.BindingCookieName + "=", "Path=/api/mail/setup/oauth", "HttpOnly", "Secure", "SameSite=Lax"} {
		if !strings.Contains(cookie, fragment) {
			t.Fatalf("Set-Cookie = %q, missing %q", cookie, fragment)
		}
	}

	fixture = newMailboxSetupHTTPFixture(t)
	fixture.ownerRecord.State = mail.MailboxOwnerActive
	fixture.owners = mailboxSetupOwnerService(t, fixture.ownerRecord)
	fixture.replaceHandler(t)
	request = httptest.NewRequest(http.MethodPost, "/api/mail/setup/oauth/authorize", nil)
	request.Header.Set("X-Test-Auth", "yes")
	response = mailboxSetupHTTPResponse(t, fixture.app, request)
	if response.StatusCode != http.StatusConflict || response.Header.Get("X-Hub-Error-Code") != "MAIL_SETUP_NOT_READY" || fixture.oauth.beginCalls != 0 {
		t.Fatalf("non-bootstrap owner status/code/calls = %d/%q/%d, want 409/%q/0", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), fixture.oauth.beginCalls, "MAIL_SETUP_NOT_READY")
	}
}

func TestMailboxSetupHandler_CallbackIsStrictClearsCookieAndPassesExactBinding(t *testing.T) {
	fixture := newMailboxSetupHTTPFixture(t)
	validState := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	validCode := "authorization-code-1"
	validCookie := fixture.oauth.begin.BrowserCookie
	request := httptest.NewRequest(http.MethodPost, "/api/mail/setup/oauth/callback", strings.NewReader(`{"state":"`+validState+`","code":"`+validCode+`"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Test-Auth", "yes")
	request.Header.Set("Cookie", fixture.config.BindingCookieName+"="+validCookie)
	response := mailboxSetupHTTPResponse(t, fixture.app, request)
	if response.StatusCode != http.StatusNoContent || fixture.oauth.completeCalls != 1 || fixture.oauth.completeScope != fixture.scope || fixture.oauth.completeState != validState || fixture.oauth.completeCode != validCode || fixture.oauth.completeCookie != validCookie {
		t.Fatalf("callback status/call = %d/%d scope=%#v state=%q code=%q cookie=%q, want 204/1 exact values", response.StatusCode, fixture.oauth.completeCalls, fixture.oauth.completeScope, fixture.oauth.completeState, fixture.oauth.completeCode, fixture.oauth.completeCookie)
	}
	mailboxSetupRequireClearingCookie(t, response.Header.Get("Set-Cookie"), fixture.config.BindingCookieName)

	for _, body := range []string{
		`{"state":"` + validState + `","state":"` + validState + `","code":"` + validCode + `"}`,
		`{"state":"` + validState + `","code":"` + validCode + `","extra":"no"}`,
		`{"state":"` + validState + `","code":"` + validCode + `"}{}`,
	} {
		fixture := newMailboxSetupHTTPFixture(t)
		request := httptest.NewRequest(http.MethodPost, "/api/mail/setup/oauth/callback", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Test-Auth", "yes")
		request.Header.Set("Cookie", fixture.config.BindingCookieName+"="+validCookie)
		response := mailboxSetupHTTPResponse(t, fixture.app, request)
		if response.StatusCode != http.StatusBadRequest || response.Header.Get("X-Hub-Error-Code") != "MAIL_SETUP_INVALID_REQUEST" || fixture.oauth.completeCalls != 0 {
			t.Fatalf("strict callback status/code/calls = %d/%q/%d, want 400/%q/0", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), fixture.oauth.completeCalls, "MAIL_SETUP_INVALID_REQUEST")
		}
		mailboxSetupRequireClearingCookie(t, response.Header.Get("Set-Cookie"), fixture.config.BindingCookieName)
	}
}

func TestMailboxSetupHandler_CallbackMapsContentFreeOAuthFailures(t *testing.T) {
	fixture := newMailboxSetupHTTPFixture(t)
	fixture.oauth.completeErr = errors.New("provider response must stay private")
	state := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	request := httptest.NewRequest(http.MethodPost, "/api/mail/setup/oauth/callback", strings.NewReader(`{"state":"`+state+`","code":"authorization-code-1"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Test-Auth", "yes")
	request.Header.Set("Cookie", fixture.config.BindingCookieName+"="+fixture.oauth.begin.BrowserCookie)
	response := mailboxSetupHTTPResponse(t, fixture.app, request)
	body := mailboxSetupBody(t, response)
	if response.StatusCode != http.StatusServiceUnavailable || response.Header.Get("X-Hub-Error-Code") != "MAIL_SETUP_UNAVAILABLE" || strings.Contains(string(body), "provider response") {
		t.Fatalf("callback failure status/code/body = %d/%q/%q, want redacted 503/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), body, "MAIL_SETUP_UNAVAILABLE")
	}
	mailboxSetupRequireClearingCookie(t, response.Header.Get("Set-Cookie"), fixture.config.BindingCookieName)
}

type mailboxSetupHTTPFixture struct {
	app         *fiber.App
	password    *mailboxSetupPasswordFake
	oauth       *mailboxSetupOAuthFake
	owners      *mail.MailboxService
	ownerRecord mail.MailboxOwnerRecord
	actorID     uuid.UUID
	scope       mail.MailboxOAuthScope
	config      mail.MailboxSetupHandlerConfig
}

func newMailboxSetupHTTPFixture(t *testing.T) *mailboxSetupHTTPFixture {
	t.Helper()
	actorID := uuid.New()
	scope := mail.MailboxOAuthScope{ActorID: actorID, OrganizationID: uuid.New(), AccountID: uuid.New()}
	engineAccountID := "engine-account-1"
	primaryAddress := "member@example.test"
	fixture := &mailboxSetupHTTPFixture{
		password: &mailboxSetupPasswordFake{result: mail.MailboxPasswordEnrollmentResultEnrolled},
		oauth: &mailboxSetupOAuthFake{begin: mail.MailboxOAuthBeginResult{
			AuthorizationURL: "https://mail.example.test/authorize?state=test",
			ExpiresAt:        time.Now().UTC().Add(time.Minute).Truncate(time.Second),
			BrowserCookie:    base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32)),
		}},
		ownerRecord: mail.MailboxOwnerRecord{ActorID: actorID, OrganizationID: scope.OrganizationID, AccountID: scope.AccountID, EngineAccountID: &engineAccountID, State: mail.MailboxOwnerProvisioning, ProjectionState: mail.MailboxProjectionBootstrapping, PrimaryAddress: &primaryAddress},
		actorID:     actorID,
		scope:       scope,
		config: mail.MailboxSetupHandlerConfig{
			BindingCookieName:      "mail_setup_binding",
			BindingCookieSecure:    true,
			BindingCookieMaxAgeSec: 300,
			MaximumJSONBytes:       9216,
			MaximumQueryBytes:      2048,
			MaximumPasswordBytes:   256,
		},
	}
	fixture.owners = mailboxSetupOwnerService(t, fixture.ownerRecord)
	fixture.replaceHandler(t)
	return fixture
}

func (fixture *mailboxSetupHTTPFixture) replaceHandler(t *testing.T) {
	t.Helper()
	handler, err := mail.NewMailboxSetupHandler(fixture.password, fixture.oauth, fixture.owners, fixture.config)
	if err != nil {
		t.Fatalf("NewMailboxSetupHandler() error = %v", err)
	}
	fixture.app = fiber.New()
	api := fixture.app.Group("/api")
	authMiddleware := func(ctx fiber.Ctx) error {
		if ctx.Get("X-Test-Auth") != "yes" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		ctx.Locals("userID", fixture.actorID.String())
		return ctx.Next()
	}
	if err := handler.RegisterRoutes(api, authMiddleware); err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}
}

func mailboxSetupOwnerService(t *testing.T, record mail.MailboxOwnerRecord) *mail.MailboxService {
	t.Helper()
	owners, err := mail.NewMailboxService(&mailboxSetupOwnerRepository{record: record})
	if err != nil {
		t.Fatalf("NewMailboxService() error = %v", err)
	}
	return owners
}

func mailboxSetupHTTPResponse(t *testing.T, app *fiber.App, request *http.Request) *http.Response {
	t.Helper()
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return response
}

func mailboxSetupBody(t *testing.T, response *http.Response) []byte {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return body
}

func mailboxSetupRequireClearingCookie(t *testing.T, cookie, name string) {
	t.Helper()
	for _, fragment := range []string{name + "=", "Path=/api/mail/setup/oauth", "HttpOnly", "Secure", "SameSite=Lax"} {
		if !strings.Contains(cookie, fragment) {
			t.Fatalf("clearing Set-Cookie = %q, missing %q", cookie, fragment)
		}
	}
	if !strings.Contains(cookie, "Max-Age=0") && !strings.Contains(cookie, "max-age=0") {
		t.Fatalf("clearing Set-Cookie = %q, missing Max-Age=0", cookie)
	}
}

func mailboxSetupAllZero(value []byte) bool {
	for _, byteValue := range value {
		if byteValue != 0 {
			return false
		}
	}
	return true
}

type mailboxSetupPasswordFake struct {
	result          mail.MailboxPasswordEnrollmentResult
	err             error
	calls           int
	passwordAtCall  []byte
	retainedPassword []byte
}

func (fake *mailboxSetupPasswordFake) EnrollPassword(_ context.Context, _ uuid.UUID, password []byte) (mail.MailboxPasswordEnrollmentResult, error) {
	fake.calls++
	fake.passwordAtCall = append([]byte(nil), password...)
	fake.retainedPassword = password
	return fake.result, fake.err
}

type mailboxSetupOAuthFake struct {
	begin          mail.MailboxOAuthBeginResult
	beginErr       error
	completeErr    error
	beginCalls     int
	completeCalls  int
	beginScope     mail.MailboxOAuthScope
	completeScope  mail.MailboxOAuthScope
	completeState  string
	completeCode   string
	completeCookie string
}

func (fake *mailboxSetupOAuthFake) BeginMailboxBootstrapOAuth(_ context.Context, scope mail.MailboxOAuthScope) (mail.MailboxOAuthBeginResult, error) {
	fake.beginCalls++
	fake.beginScope = scope
	return fake.begin, fake.beginErr
}

func (fake *mailboxSetupOAuthFake) CompleteMailboxBootstrapOAuth(_ context.Context, scope mail.MailboxOAuthScope, state, code, cookie string) (mail.MailboxOAuthConnectionResult, error) {
	fake.completeCalls++
	fake.completeScope = scope
	fake.completeState = state
	fake.completeCode = code
	fake.completeCookie = cookie
	return mail.MailboxOAuthConnectionResult{}, fake.completeErr
}

type mailboxSetupOwnerRepository struct {
	record mail.MailboxOwnerRecord
	err    error
}

func (repository *mailboxSetupOwnerRepository) ResolveMailboxOwner(_ context.Context, _ uuid.UUID) (mail.MailboxOwnerRecord, error) {
	return repository.record, repository.err
}
