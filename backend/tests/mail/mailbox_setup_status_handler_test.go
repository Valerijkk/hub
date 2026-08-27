package mail_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxSetupStatusHandler_ConstructorAndAuthenticatedRoute(t *testing.T) {
	fixture := newMailboxSetupStatusHTTPFixture(t)

	request := httptest.NewRequest(http.MethodGet, "/api/mail/setup/status", nil)
	response := mailboxSetupStatusHTTPResponse(t, fixture.app, request)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
	if fixture.source.calls != 0 {
		t.Fatalf("unauthenticated source calls = %d, want 0", fixture.source.calls)
	}

	if _, err := mail.NewMailboxSetupStatusHandler(nil); err == nil {
		t.Fatal("NewMailboxSetupStatusHandler(nil) error = nil, want invalid dependency error")
	}
}

func TestMailboxSetupStatusHandler_ReturnsOnlyClosedPublicStatuses(t *testing.T) {
	primaryAddress := "member@example.test"
	for _, testCase := range []struct {
		name       string
		status     mail.MailboxSetupStatus
		wantBody   string
		wantAddress *string
	}{
		{name: "preparing without address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStagePreparing}, wantBody: `{"stage":"preparing","primary_address":null}`},
		{name: "password required", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStagePasswordRequired, PrimaryAddress: &primaryAddress}, wantBody: `{"stage":"password_required","primary_address":"member@example.test"}`, wantAddress: &primaryAddress},
		{name: "oauth required", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageOAuthRequired, PrimaryAddress: &primaryAddress}, wantBody: `{"stage":"oauth_required","primary_address":"member@example.test"}`, wantAddress: &primaryAddress},
		{name: "syncing", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageSyncing, PrimaryAddress: &primaryAddress}, wantBody: `{"stage":"syncing","primary_address":"member@example.test"}`, wantAddress: &primaryAddress},
		{name: "ready", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageReady, PrimaryAddress: &primaryAddress}, wantBody: `{"stage":"ready","primary_address":"member@example.test"}`, wantAddress: &primaryAddress},
		{name: "action required without address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageActionRequired}, wantBody: `{"stage":"action_required","primary_address":null}`},
		{name: "action required with address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageActionRequired, PrimaryAddress: &primaryAddress}, wantBody: `{"stage":"action_required","primary_address":"member@example.test"}`, wantAddress: &primaryAddress},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newMailboxSetupStatusHTTPFixture(t)
			fixture.source.status = testCase.status
			request := httptest.NewRequest(http.MethodGet, "/api/mail/setup/status", bytes.NewReader([]byte(`{"ignored":"body"}`)))
			request.Header.Set("X-Test-Auth", "yes")
			response := mailboxSetupStatusHTTPResponse(t, fixture.app, request)
			body := mailboxSetupStatusHTTPBody(t, response)

			if response.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
			if string(body) != testCase.wantBody {
				t.Fatalf("response body = %q, want %q", body, testCase.wantBody)
			}
			if fixture.source.calls != 1 || fixture.source.actorID != fixture.actorID {
				t.Fatalf("source calls/actor = %d/%s, want 1/%s", fixture.source.calls, fixture.source.actorID, fixture.actorID)
			}
			if strings.Contains(string(body), fixture.internalAccountID.String()) || strings.Contains(string(body), fixture.internalOrganizationID.String()) {
				t.Fatalf("response leaked internal identifier: %q", body)
			}
		})
	}
}

func TestMailboxSetupStatusHandler_RejectsQueriesBeforeSource(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		path       string
		wantStatus int
		wantCode   string
	}{
		{name: "non-empty", path: "/api/mail/setup/status?unexpected=1", wantStatus: http.StatusBadRequest, wantCode: "MAIL_SETUP_INVALID_QUERY"},
		{name: "too large", path: "/api/mail/setup/status?" + strings.Repeat("x", 2049), wantStatus: http.StatusRequestURITooLong, wantCode: "MAIL_SETUP_QUERY_TOO_LARGE"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newMailboxSetupStatusHTTPFixture(t)
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			request.Header.Set("X-Test-Auth", "yes")
			response := mailboxSetupStatusHTTPResponse(t, fixture.app, request)
			if response.StatusCode != testCase.wantStatus || response.Header.Get("X-Hub-Error-Code") != testCase.wantCode {
				t.Fatalf("status/code = %d/%q, want %d/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), testCase.wantStatus, testCase.wantCode)
			}
			if fixture.source.calls != 0 {
				t.Fatalf("source calls = %d, want 0", fixture.source.calls)
			}
		})
	}
}

func TestMailboxSetupStatusHandler_RejectsInvalidSourceOutput(t *testing.T) {
	malformedAddress := "not an address"
	for _, testCase := range []struct {
		name   string
		status mail.MailboxSetupStatus
	}{
		{name: "unknown stage", status: mail.MailboxSetupStatus{Stage: "unknown"}},
		{name: "password required without address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStagePasswordRequired}},
		{name: "oauth required without address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageOAuthRequired}},
		{name: "syncing without address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageSyncing}},
		{name: "ready without address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageReady}},
		{name: "invalid optional address", status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStageActionRequired, PrimaryAddress: &malformedAddress}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newMailboxSetupStatusHTTPFixture(t)
			fixture.source.status = testCase.status
			request := httptest.NewRequest(http.MethodGet, "/api/mail/setup/status", nil)
			request.Header.Set("X-Test-Auth", "yes")
			response := mailboxSetupStatusHTTPResponse(t, fixture.app, request)
			body := mailboxSetupStatusHTTPBody(t, response)
			if response.StatusCode != http.StatusServiceUnavailable || response.Header.Get("X-Hub-Error-Code") != "MAIL_SETUP_UNAVAILABLE" {
				t.Fatalf("status/code = %d/%q, want 503/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), "MAIL_SETUP_UNAVAILABLE")
			}
			if strings.Contains(string(body), malformedAddress) || strings.Contains(string(body), "unknown") {
				t.Fatalf("invalid source output leaked in response: %q", body)
			}
		})
	}
}

func TestMailboxSetupStatusHandler_MapsSourceFailuresWithoutCauseLeak(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "not found", err: &mail.MailboxServiceError{Kind: mail.MailboxServiceErrorNotFound}, wantStatus: http.StatusConflict, wantCode: "MAIL_SETUP_NOT_READY"},
		{name: "deleting", err: &mail.MailboxServiceError{Kind: mail.MailboxServiceErrorDeleting}, wantStatus: http.StatusConflict, wantCode: "MAIL_SETUP_NOT_READY"},
		{name: "invalid", err: &mail.MailboxServiceError{Kind: mail.MailboxServiceErrorInvalid}, wantStatus: http.StatusConflict, wantCode: "MAIL_SETUP_NOT_READY"},
		{name: "repository unavailable", err: &mail.MailboxServiceError{Kind: mail.MailboxServiceErrorUnavailable}, wantStatus: http.StatusServiceUnavailable, wantCode: "MAIL_SETUP_UNAVAILABLE"},
		{name: "unknown", err: errors.New("private repository cause"), wantStatus: http.StatusServiceUnavailable, wantCode: "MAIL_SETUP_UNAVAILABLE"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newMailboxSetupStatusHTTPFixture(t)
			fixture.source.err = testCase.err
			request := httptest.NewRequest(http.MethodGet, "/api/mail/setup/status", nil)
			request.Header.Set("X-Test-Auth", "yes")
			response := mailboxSetupStatusHTTPResponse(t, fixture.app, request)
			body := mailboxSetupStatusHTTPBody(t, response)
			if response.StatusCode != testCase.wantStatus || response.Header.Get("X-Hub-Error-Code") != testCase.wantCode {
				t.Fatalf("status/code = %d/%q, want %d/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), testCase.wantStatus, testCase.wantCode)
			}
			if strings.Contains(string(body), "private repository cause") {
				t.Fatalf("source error leaked in response: %q", body)
			}
		})
	}
}

type mailboxSetupStatusHTTPFixture struct {
	app                    *fiber.App
	source                 *mailboxSetupStatusFake
	actorID                uuid.UUID
	internalAccountID      uuid.UUID
	internalOrganizationID uuid.UUID
}

func newMailboxSetupStatusHTTPFixture(t *testing.T) *mailboxSetupStatusHTTPFixture {
	t.Helper()
	fixture := &mailboxSetupStatusHTTPFixture{
		actorID:                uuid.New(),
		internalAccountID:      uuid.New(),
		internalOrganizationID: uuid.New(),
	}
	fixture.source = &mailboxSetupStatusFake{status: mail.MailboxSetupStatus{Stage: mail.MailboxSetupStagePreparing}}
	handler, err := mail.NewMailboxSetupStatusHandler(fixture.source)
	if err != nil {
		t.Fatalf("NewMailboxSetupStatusHandler() error = %v", err)
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
	return fixture
}

func mailboxSetupStatusHTTPResponse(t *testing.T, app *fiber.App, request *http.Request) *http.Response {
	t.Helper()
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return response
}

func mailboxSetupStatusHTTPBody(t *testing.T, response *http.Response) []byte {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return body
}

type mailboxSetupStatusFake struct {
	status  mail.MailboxSetupStatus
	err     error
	calls   int
	actorID uuid.UUID
}

func (fake *mailboxSetupStatusFake) ResolveMailboxSetupStatus(_ context.Context, actorID uuid.UUID) (mail.MailboxSetupStatus, error) {
	fake.calls++
	fake.actorID = actorID
	return fake.status, fake.err
}

func TestMailboxSetupStatusHandler_ResponseSchemaHasNoUndocumentedFields(t *testing.T) {
	fixture := newMailboxSetupStatusHTTPFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/api/mail/setup/status", nil)
	request.Header.Set("X-Test-Auth", "yes")
	response := mailboxSetupStatusHTTPResponse(t, fixture.app, request)
	body := mailboxSetupStatusHTTPBody(t, response)
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload) != 2 || payload["stage"] == nil || payload["primary_address"] == nil {
		t.Fatalf("response fields = %#v, want only stage and primary_address", payload)
	}
}
