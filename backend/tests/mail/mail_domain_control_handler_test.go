package mail_test

import (
	"context"
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

var (
	mailDomainControlHTTPActorID        = uuid.MustParse("2ebf0c39-3d1b-4f23-9e24-0d570907fad1")
	mailDomainControlHTTPDomainID       = uuid.MustParse("b2c07faa-6bf1-4c2e-9c8b-7f21dcb6f8a2")
	mailDomainControlHTTPIdempotencyKey = uuid.MustParse("b8d7d95a-e1ef-4374-b018-889f95aeef30")
)

type mailDomainControlHTTPServiceSpy struct {
	listResult    []mail.MailDomainControlView
	createResult  mail.MailDomainControlCreateResult
	defaultResult mail.MailDomainControlView
	err           error
	listCalls     int
	createCalls   int
	defaultCalls  int
	listInputs    []mail.MailDomainControlListRequest
	createInputs  []mail.MailDomainControlCreateRequest
	defaultInputs []mail.MailDomainControlSetDefaultRequest
}

func (spy *mailDomainControlHTTPServiceSpy) ListMailDomains(_ context.Context, request mail.MailDomainControlListRequest) ([]mail.MailDomainControlView, error) {
	spy.listCalls++
	spy.listInputs = append(spy.listInputs, request)
	return append([]mail.MailDomainControlView(nil), spy.listResult...), spy.err
}

func (spy *mailDomainControlHTTPServiceSpy) CreateMailDomain(_ context.Context, request mail.MailDomainControlCreateRequest) (mail.MailDomainControlCreateResult, error) {
	spy.createCalls++
	spy.createInputs = append(spy.createInputs, request)
	return spy.createResult, spy.err
}

func (spy *mailDomainControlHTTPServiceSpy) SetDefaultMailDomain(_ context.Context, request mail.MailDomainControlSetDefaultRequest) (mail.MailDomainControlView, error) {
	spy.defaultCalls++
	spy.defaultInputs = append(spy.defaultInputs, request)
	return spy.defaultResult, spy.err
}

func (spy *mailDomainControlHTTPServiceSpy) calls() int {
	return spy.listCalls + spy.createCalls + spy.defaultCalls
}

func TestMailDomainControlHandler_AdminRoutesUseOnlySafeCoreInputsAndDTOs(t *testing.T) {
	pending := mailDomainControlHTTPPendingView()
	verified := mailDomainControlHTTPVerifiedView()
	service := &mailDomainControlHTTPServiceSpy{
		listResult:    []mail.MailDomainControlView{pending},
		createResult:  mail.MailDomainControlCreateResult{Domain: pending},
		defaultResult: verified,
	}
	app, adminCalls := mailDomainControlHTTPApp(t, service, mailDomainControlHTTPActorID.String(), true, nil)

	listResponse := mailDomainControlHTTPDo(t, app, httptest.NewRequest(http.MethodGet, "/api/admin/mail/domains", nil))
	listBody := mailDomainControlHTTPBody(t, listResponse)
	if listResponse.StatusCode != http.StatusOK {
		t.Fatalf("GET domains status = %d, want 200", listResponse.StatusCode)
	}
	var listed struct {
		Domains []mailDomainControlHTTPDomainResponse `json:"domains"`
	}
	if err := json.Unmarshal(listBody, &listed); err != nil {
		t.Fatalf("decode GET domains: %v", err)
	}
	if len(listed.Domains) != 1 || !mailDomainControlHTTPDomainResponseEqual(listed.Domains[0], pending) {
		t.Fatalf("GET domains payload = %#v, want one safe pending domain", listed)
	}
	if strings.Contains(string(listBody), "hub-mail-domain-v1") || strings.Contains(string(listBody), "lease") || strings.Contains(string(listBody), "engine_") {
		t.Fatalf("GET domains leaked control material: %s", listBody)
	}
	if service.listCalls != 1 || len(service.listInputs) != 1 || service.listInputs[0].ActorID != mailDomainControlHTTPActorID || *adminCalls != 1 {
		t.Fatalf("GET inputs/calls = list=%d inputs=%#v admin=%d, want exact actor call", service.listCalls, service.listInputs, *adminCalls)
	}

	createRequest := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"example.test"}`))
	createRequest.Header.Set(fiber.HeaderContentType, "application/json")
	createRequest.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
	createResponse := mailDomainControlHTTPDo(t, app, createRequest)
	createBody := mailDomainControlHTTPBody(t, createResponse)
	if createResponse.StatusCode != http.StatusAccepted {
		t.Fatalf("POST domain status = %d, want 202", createResponse.StatusCode)
	}
	var created struct {
		Domain mailDomainControlHTTPDomainResponse `json:"domain"`
	}
	if err := json.Unmarshal(createBody, &created); err != nil {
		t.Fatalf("decode POST domain: %v", err)
	}
	if !mailDomainControlHTTPDomainResponseEqual(created.Domain, pending) || service.createCalls != 1 || len(service.createInputs) != 1 {
		t.Fatalf("POST payload/calls = %#v/%#v, want one safe result", created, service.createInputs)
	}
	if input := service.createInputs[0]; input.ActorID != mailDomainControlHTTPActorID || input.IdempotencyKey != mailDomainControlHTTPIdempotencyKey || input.Domain != "example.test" {
		t.Fatalf("POST input = %#v, want exact actor/key/canonical domain", input)
	}

	defaultRequest := httptest.NewRequest(http.MethodPut, "/api/admin/mail/domains/"+mailDomainControlHTTPDomainID.String()+"/default", nil)
	defaultResponse := mailDomainControlHTTPDo(t, app, defaultRequest)
	defaultBody := mailDomainControlHTTPBody(t, defaultResponse)
	if defaultResponse.StatusCode != http.StatusOK {
		t.Fatalf("PUT default status = %d, want 200", defaultResponse.StatusCode)
	}
	var madeDefault struct {
		Domain mailDomainControlHTTPDomainResponse `json:"domain"`
	}
	if err := json.Unmarshal(defaultBody, &madeDefault); err != nil {
		t.Fatalf("decode PUT default: %v", err)
	}
	if !mailDomainControlHTTPDomainResponseEqual(madeDefault.Domain, verified) || service.defaultCalls != 1 || len(service.defaultInputs) != 1 {
		t.Fatalf("PUT payload/calls = %#v/%#v, want one safe verified result", madeDefault, service.defaultInputs)
	}
	if input := service.defaultInputs[0]; input.ActorID != mailDomainControlHTTPActorID || input.DomainID != mailDomainControlHTTPDomainID {
		t.Fatalf("PUT input = %#v, want exact actor/domain", input)
	}
}

func TestMailDomainControlHandler_RejectsTransportBeforeService(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		request func() *http.Request
	}{
		{name: "list query", request: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/admin/mail/domains?org_id=forbidden", nil) }},
		{name: "list body", request: func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/admin/mail/domains", strings.NewReader("{}")) }},
		{name: "post query", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains?unexpected=1", strings.NewReader(`{"domain":"example.test"}`))
			request.Header.Set(fiber.HeaderContentType, "application/json")
			request.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			return request
		}},
		{name: "post duplicate idempotency", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"example.test"}`))
			request.Header.Set(fiber.HeaderContentType, "application/json")
			request.Header.Add("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			request.Header.Add("Idempotency-Key", uuid.NewString())
			return request
		}},
		{name: "post duplicate content type", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"example.test"}`))
			request.Header.Add(fiber.HeaderContentType, "application/json")
			request.Header.Add(fiber.HeaderContentType, "application/json")
			request.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			return request
		}},
		{name: "post noncanonical domain", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"Example.Test"}`))
			request.Header.Set(fiber.HeaderContentType, "application/json")
			request.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			return request
		}},
		{name: "post unknown field", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"example.test","organization_id":"forbidden"}`))
			request.Header.Set(fiber.HeaderContentType, "application/json")
			request.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			return request
		}},
		{name: "post duplicate field", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"example.test","domain":"other.test"}`))
			request.Header.Set(fiber.HeaderContentType, "application/json")
			request.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			return request
		}},
		{name: "post trailing JSON", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"example.test"} {}`))
			request.Header.Set(fiber.HeaderContentType, "application/json")
			request.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			return request
		}},
		{name: "post oversized", request: func() *http.Request {
			request := httptest.NewRequest(http.MethodPost, "/api/admin/mail/domains", strings.NewReader(`{"domain":"`+strings.Repeat("a", 4096)+`.test"}`))
			request.Header.Set(fiber.HeaderContentType, "application/json")
			request.Header.Set("Idempotency-Key", mailDomainControlHTTPIdempotencyKey.String())
			return request
		}},
		{name: "default body", request: func() *http.Request { return httptest.NewRequest(http.MethodPut, "/api/admin/mail/domains/"+mailDomainControlHTTPDomainID.String()+"/default", strings.NewReader("{}")) }},
		{name: "default invalid id", request: func() *http.Request { return httptest.NewRequest(http.MethodPut, "/api/admin/mail/domains/not-a-uuid/default", nil) }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			service := &mailDomainControlHTTPServiceSpy{}
			app, _ := mailDomainControlHTTPApp(t, service, mailDomainControlHTTPActorID.String(), true, nil)
			response := mailDomainControlHTTPDo(t, app, testCase.request())
			body := mailDomainControlHTTPBody(t, response)
			if response.StatusCode != http.StatusBadRequest || response.Header.Get("X-Hub-Error-Code") != "MAIL_DOMAIN_INVALID" {
				t.Fatalf("status/code = %d/%q, want 400/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), "MAIL_DOMAIN_INVALID")
			}
			if service.calls() != 0 || strings.Contains(string(body), "forbidden") || strings.Contains(string(body), "other.test") {
				t.Fatalf("rejected transport service/body = %d/%q, want zero call and no echo", service.calls(), body)
			}
		})
	}
}

func TestMailDomainControlHandler_FailsClosedForActorAndAdminGate(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		actorValue string
		admin      bool
		adminErr   error
		wantStatus int
		wantCode   string
		wantAdmin  int
	}{
		{name: "missing actor", wantStatus: http.StatusUnauthorized, wantCode: "MAIL_DOMAIN_UNAUTHORIZED"},
		{name: "invalid actor", actorValue: "not-a-uuid", wantStatus: http.StatusUnauthorized, wantCode: "MAIL_DOMAIN_UNAUTHORIZED"},
		{name: "known non-admin", actorValue: mailDomainControlHTTPActorID.String(), wantStatus: http.StatusForbidden, wantCode: "MAIL_DOMAIN_FORBIDDEN", wantAdmin: 1},
		{name: "admin source failure", actorValue: mailDomainControlHTTPActorID.String(), admin: true, adminErr: errors.New("private audit source failure"), wantStatus: http.StatusServiceUnavailable, wantCode: "MAIL_DOMAIN_UNAVAILABLE", wantAdmin: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			service := &mailDomainControlHTTPServiceSpy{}
			app, adminCalls := mailDomainControlHTTPApp(t, service, testCase.actorValue, testCase.admin, testCase.adminErr)
			response := mailDomainControlHTTPDo(t, app, httptest.NewRequest(http.MethodGet, "/api/admin/mail/domains", nil))
			body := mailDomainControlHTTPBody(t, response)
			if response.StatusCode != testCase.wantStatus || response.Header.Get("X-Hub-Error-Code") != testCase.wantCode {
				t.Fatalf("status/code = %d/%q, want %d/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), testCase.wantStatus, testCase.wantCode)
			}
			if service.calls() != 0 || *adminCalls != testCase.wantAdmin || strings.Contains(string(body), "private audit source failure") {
				t.Fatalf("gate service/admin/body = %d/%d/%q, want 0/%d/no leak", service.calls(), *adminCalls, body, testCase.wantAdmin)
			}
		})
	}
}

func TestMailDomainControlHandler_MapsOnlyClosedCoreErrors(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid", err: &mail.MailDomainControlError{Kind: mail.MailDomainControlErrorInvalid}, wantStatus: http.StatusBadRequest, wantCode: "MAIL_DOMAIN_INVALID"},
		{name: "not found", err: &mail.MailDomainControlError{Kind: mail.MailDomainControlErrorNotFound}, wantStatus: http.StatusNotFound, wantCode: "MAIL_DOMAIN_NOT_FOUND"},
		{name: "conflict", err: &mail.MailDomainControlError{Kind: mail.MailDomainControlErrorConflict}, wantStatus: http.StatusConflict, wantCode: "MAIL_DOMAIN_CONFLICT"},
		{name: "unavailable", err: &mail.MailDomainControlError{Kind: mail.MailDomainControlErrorUnavailable}, wantStatus: http.StatusServiceUnavailable, wantCode: "MAIL_DOMAIN_UNAVAILABLE"},
		{name: "opaque", err: errors.New("provider database private cause"), wantStatus: http.StatusServiceUnavailable, wantCode: "MAIL_DOMAIN_UNAVAILABLE"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			service := &mailDomainControlHTTPServiceSpy{err: testCase.err}
			app, _ := mailDomainControlHTTPApp(t, service, mailDomainControlHTTPActorID.String(), true, nil)
			response := mailDomainControlHTTPDo(t, app, httptest.NewRequest(http.MethodGet, "/api/admin/mail/domains", nil))
			body := mailDomainControlHTTPBody(t, response)
			if response.StatusCode != testCase.wantStatus || response.Header.Get("X-Hub-Error-Code") != testCase.wantCode {
				t.Fatalf("status/code = %d/%q, want %d/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), testCase.wantStatus, testCase.wantCode)
			}
			if strings.Contains(string(body), "provider database private cause") {
				t.Fatalf("opaque service cause leaked: %q", body)
			}
		})
	}
}

type mailDomainControlHTTPDomainResponse struct {
	ID                 string                    `json:"id"`
	Domain             string                    `json:"domain"`
	State              string                    `json:"state"`
	VerificationStatus string                    `json:"verification_status"`
	DKIMStatus         string                    `json:"dkim_status"`
	RequiredDNSRecords []mail.RequiredDNSRecord `json:"required_dns_records"`
	IsDefault          bool                      `json:"is_default"`
}

func mailDomainControlHTTPDomainResponseEqual(response mailDomainControlHTTPDomainResponse, view mail.MailDomainControlView) bool {
	return response.ID == view.ID.String() && response.Domain == view.Domain && response.State == view.State && response.VerificationStatus == view.VerificationStatus && response.DKIMStatus == view.DKIMStatus && len(response.RequiredDNSRecords) == len(view.RequiredDNSRecords) && response.IsDefault == view.IsDefault
}

func mailDomainControlHTTPApp(t *testing.T, service *mailDomainControlHTTPServiceSpy, actorValue string, admin bool, adminErr error) (*fiber.App, *int) {
	t.Helper()
	adminCalls := new(int)
	handler, err := mail.NewMailDomainControlHandler(service, func(_ fiber.Ctx, actorID uuid.UUID) (bool, error) {
		*adminCalls++
		if actorValue != mailDomainControlHTTPActorID.String() || actorID != mailDomainControlHTTPActorID {
			return false, errors.New("unexpected actor")
		}
		return admin, adminErr
	})
	if err != nil {
		t.Fatalf("NewMailDomainControlHandler: %v", err)
	}
	app := fiber.New()
	api := app.Group("/api")
	auth := func(ctx fiber.Ctx) error {
		if actorValue != "" {
			ctx.Locals("userID", actorValue)
		}
		return ctx.Next()
	}
	if err := handler.RegisterRoutes(api, auth); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	return app, adminCalls
}

func mailDomainControlHTTPDo(t *testing.T, app *fiber.App, request *http.Request) *http.Response {
	t.Helper()
	response, err := app.Test(request, 5*time.Second)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return response
}

func mailDomainControlHTTPBody(t *testing.T, response *http.Response) []byte {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return body
}

func mailDomainControlHTTPPendingView() mail.MailDomainControlView {
	return mail.MailDomainControlView{
		ID:                 mailDomainControlHTTPDomainID,
		Domain:             "example.test",
		State:              "pending",
		VerificationStatus: "pending",
		DKIMStatus:         "unknown",
		RequiredDNSRecords: mail.RequiredDNSRecords{},
	}
}

func mailDomainControlHTTPVerifiedView() mail.MailDomainControlView {
	return mail.MailDomainControlView{
		ID:                 mailDomainControlHTTPDomainID,
		Domain:             "example.test",
		State:              "verified",
		VerificationStatus: "verified",
		DKIMStatus:         "valid",
		RequiredDNSRecords: mail.RequiredDNSRecords{},
		IsDefault:          true,
	}
}
