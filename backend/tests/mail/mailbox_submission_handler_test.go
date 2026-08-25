package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	mailboxSubmissionHandlerActorID = uuid.MustParse("2e1c0ee5-5467-4b2e-ab9e-16bc5f42b205")
	mailboxSubmissionHandlerDraftID = uuid.MustParse("a5c6428e-627e-401a-8e72-702d68eeb34f")
	mailboxSubmissionHandlerReceiptID = uuid.MustParse("b2b0f72b-f4bb-42c1-9ed0-36645c5c4a5c")
)

type mailboxSubmissionHandlerServiceSpy struct {
	mu       sync.Mutex
	result   mail.MailboxSubmissionView
	err      error
	calls    int
	actors   []uuid.UUID
	drafts   []uuid.UUID
	inputs   []mail.MailboxSubmissionCreateInput
}

func (spy *mailboxSubmissionHandlerServiceSpy) CreateMailboxSubmission(_ context.Context, actorID, draftID uuid.UUID, input mail.MailboxSubmissionCreateInput) (mail.MailboxSubmissionView, error) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.calls++
	spy.actors = append(spy.actors, actorID)
	spy.drafts = append(spy.drafts, draftID)
	spy.inputs = append(spy.inputs, input)
	return spy.result, spy.err
}

func (spy *mailboxSubmissionHandlerServiceSpy) snapshot() (int, []uuid.UUID, []uuid.UUID, []mail.MailboxSubmissionCreateInput) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return spy.calls, append([]uuid.UUID(nil), spy.actors...), append([]uuid.UUID(nil), spy.drafts...), append([]mail.MailboxSubmissionCreateInput(nil), spy.inputs...)
}

func TestMailboxSubmissionHandler_SendCreatesDurableReceiptWithExactInput(t *testing.T) {
	createdAt := time.Date(2026, time.August, 25, 9, 10, 11, 0, time.UTC)
	spy := &mailboxSubmissionHandlerServiceSpy{result: mail.MailboxSubmissionView{
		ID:        mailboxSubmissionHandlerReceiptID,
		Status:    mail.MailboxSubmissionStatusQueued,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}}
	app := mailboxSubmissionHandlerApp(t, spy, true)

	request := mailboxSubmissionHandlerRequest(t, mailboxSubmissionHandlerDraftID.String(), nil)
	request.Header.Set("Idempotency-Key", "draft-e\u0301")
	request.Header.Set("If-Match", "\"7\"")
	response := mailboxSubmissionHandlerDo(t, app, request)
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("POST send status = %d, want %d", response.StatusCode, http.StatusAccepted)
	}
	var payload mailboxSubmissionHandlerResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode receipt: %v", err)
	}
	if payload.ID != mailboxSubmissionHandlerReceiptID.String() || payload.Status != string(mail.MailboxSubmissionStatusQueued) || payload.MessageID != nil || payload.ErrorCode != nil || !payload.CreatedAt.Equal(createdAt) || !payload.UpdatedAt.Equal(createdAt) {
		t.Fatalf("receipt = %#v, want queued durable receipt", payload)
	}

	calls, actors, drafts, inputs := spy.snapshot()
	if calls != 1 || len(actors) != 1 || actors[0] != mailboxSubmissionHandlerActorID || len(drafts) != 1 || drafts[0] != mailboxSubmissionHandlerDraftID || len(inputs) != 1 {
		t.Fatalf("service calls = %d actor=%v draft=%v input=%#v, want one exact call", calls, actors, drafts, inputs)
	}
	if inputs[0].ExpectedRevision != 7 || inputs[0].IdempotencyKey != "draft-é" {
		t.Fatalf("service input = %#v, want normalized key and exact revision", inputs[0])
	}
}

func TestMailboxSubmissionHandler_SendReplayStillReturnsAcceptedReceipt(t *testing.T) {
	messageID := uuid.MustParse("7ddba321-5b7d-42ce-98c8-8f92e650d40e")
	createdAt := time.Date(2026, time.August, 25, 9, 10, 11, 0, time.UTC)
	spy := &mailboxSubmissionHandlerServiceSpy{result: mail.MailboxSubmissionView{
		ID:        mailboxSubmissionHandlerReceiptID,
		Status:    mail.MailboxSubmissionStatusSent,
		MessageID: &messageID,
		CreatedAt: createdAt,
		UpdatedAt: createdAt.Add(time.Minute),
	}}
	app := mailboxSubmissionHandlerApp(t, spy, true)

	for attempt := 0; attempt < 2; attempt++ {
		response := mailboxSubmissionHandlerDo(t, app, mailboxSubmissionHandlerRequest(t, mailboxSubmissionHandlerDraftID.String(), nil))
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatalf("read replay response: %v", err)
		}
		if response.StatusCode != http.StatusAccepted || !strings.Contains(string(body), mailboxSubmissionHandlerReceiptID.String()) {
			t.Fatalf("replay attempt %d = status=%d body=%s, want accepted durable receipt", attempt, response.StatusCode, body)
		}
	}
	if calls, _, _, _ := spy.snapshot(); calls != 2 {
		t.Fatalf("replay service calls = %d, want exactly two transport requests", calls)
	}
}

func TestMailboxSubmissionHandler_SendRejectsInvalidTransportBeforeService(t *testing.T) {
	spy := &mailboxSubmissionHandlerServiceSpy{result: mailboxSubmissionHandlerQueuedView()}
	for _, testCase := range []struct {
		name         string
		draftID      string
		configure    func(*http.Request)
		authenticated bool
		wantStatus   int
	}{
		{name: "query", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(request *http.Request) { request.URL.RawQuery = "ignored=1" }, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "body", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(request *http.Request) { request.Body = io.NopCloser(strings.NewReader("{}")); request.ContentLength = 2 }, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "content type", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(request *http.Request) { request.Header.Set("Content-Type", "application/json") }, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "missing idempotency key", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(request *http.Request) { request.Header.Del("Idempotency-Key") }, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "control idempotency key", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(request *http.Request) { request.Header.Set("Idempotency-Key", "draft\x01key") }, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "invalid if match", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(request *http.Request) { request.Header.Set("If-Match", "7") }, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "missing if match", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(request *http.Request) { request.Header.Del("If-Match") }, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "noncanonical draft", draftID: strings.ToUpper(mailboxSubmissionHandlerDraftID.String()), configure: func(*http.Request) {}, authenticated: true, wantStatus: http.StatusBadRequest},
		{name: "missing actor", draftID: mailboxSubmissionHandlerDraftID.String(), configure: func(*http.Request) {}, authenticated: false, wantStatus: http.StatusUnauthorized},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			app := mailboxSubmissionHandlerApp(t, spy, testCase.authenticated)
			request := mailboxSubmissionHandlerRequest(t, testCase.draftID, nil)
			testCase.configure(request)
			response := mailboxSubmissionHandlerDo(t, app, request)
			response.Body.Close()
			if response.StatusCode != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", response.StatusCode, testCase.wantStatus)
			}
		})
	}
	if calls, _, _, _ := spy.snapshot(); calls != 0 {
		t.Fatalf("invalid transport invoked service %d times, want zero", calls)
	}
}

func TestMailboxSubmissionHandler_SendRejectsRepeatedRequiredHeadersBeforeService(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		configure func(*http.Request)
	}{
		{
			name: "idempotency key",
			configure: func(request *http.Request) {
				request.Header.Add("Idempotency-Key", "duplicate-key")
			},
		},
		{
			name: "if match",
			configure: func(request *http.Request) {
				request.Header.Add("If-Match", "\"8\"")
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			spy := &mailboxSubmissionHandlerServiceSpy{result: mailboxSubmissionHandlerQueuedView()}
			app := mailboxSubmissionHandlerApp(t, spy, true)
			request := mailboxSubmissionHandlerRequest(t, mailboxSubmissionHandlerDraftID.String(), nil)
			testCase.configure(request)

			response := mailboxSubmissionHandlerDo(t, app, request)
			response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
			}
			if calls, _, _, _ := spy.snapshot(); calls != 0 {
				t.Fatalf("repeated required header invoked service %d times, want zero", calls)
			}
		})
	}
}

func TestMailboxSubmissionHandler_SendMapsSafeServiceErrors(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		repository mail.MailboxSubmissionRepositoryErrorKind
		wantStatus int
		wantCode   string
	}{
		{name: "not found", repository: mail.MailboxSubmissionRepositoryErrorNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid", repository: mail.MailboxSubmissionRepositoryErrorInvalid, wantStatus: http.StatusBadRequest},
		{name: "unavailable", repository: mail.MailboxSubmissionRepositoryErrorUnavailable, wantStatus: http.StatusServiceUnavailable},
		{name: "account deleting", repository: mail.MailboxSubmissionRepositoryErrorAccountDeleting, wantStatus: http.StatusConflict, wantCode: "MAIL_ACCOUNT_DELETING"},
		{name: "version conflict", repository: mail.MailboxSubmissionRepositoryErrorVersionConflict, wantStatus: http.StatusConflict, wantCode: "MAIL_DRAFT_VERSION_CONFLICT"},
		{name: "submission in progress", repository: mail.MailboxSubmissionRepositoryErrorSubmissionInProgress, wantStatus: http.StatusConflict, wantCode: "MAIL_DRAFT_SUBMISSION_IN_PROGRESS"},
		{name: "draft not sendable", repository: mail.MailboxSubmissionRepositoryErrorDraftNotSendable, wantStatus: http.StatusConflict, wantCode: "MAIL_DRAFT_NOT_SENDABLE"},
		{name: "idempotency conflict", repository: mail.MailboxSubmissionRepositoryErrorIdempotencyConflict, wantStatus: http.StatusConflict, wantCode: "IDEMPOTENCY_KEY_REUSED"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &mailboxSubmissionHandlerRepositoryFake{err: mail.NewMailboxSubmissionRepositoryError(testCase.repository)}
			service := mailboxSubmissionHandlerRealService(t, repository)
			app := mailboxSubmissionHandlerApp(t, service, true)
			response := mailboxSubmissionHandlerDo(t, app, mailboxSubmissionHandlerRequest(t, mailboxSubmissionHandlerDraftID.String(), nil))
			body, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if response.StatusCode != testCase.wantStatus || response.Header.Get("X-Hub-Error-Code") != testCase.wantCode {
				t.Fatalf("status/code = %d/%q, want %d/%q", response.StatusCode, response.Header.Get("X-Hub-Error-Code"), testCase.wantStatus, testCase.wantCode)
			}
			if testCase.wantStatus == http.StatusNotFound && (strings.Contains(string(body), mailboxSubmissionHandlerDraftID.String()) || strings.Contains(strings.ToLower(string(body)), "owner")) {
				t.Fatalf("not found response discloses target: %s", body)
			}
			if response.StatusCode == http.StatusTooManyRequests || response.StatusCode == http.StatusLocked {
				t.Fatalf("send error status = %d, provider or lease status must not cross HTTP boundary", response.StatusCode)
			}
		})
	}
}

func TestMailboxSubmissionHandler_SendMapsOpaqueServiceErrorWithoutLeakage(t *testing.T) {
	spy := &mailboxSubmissionHandlerServiceSpy{err: errors.New("opaque")}
	app := mailboxSubmissionHandlerApp(t, spy, true)

	response := mailboxSubmissionHandlerDo(t, app, mailboxSubmissionHandlerRequest(t, mailboxSubmissionHandlerDraftID.String(), nil))
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if code := response.Header.Get("X-Hub-Error-Code"); code != "" {
		t.Fatalf("X-Hub-Error-Code = %q, want empty", code)
	}
	if strings.Contains(string(body), "opaque") {
		t.Fatalf("opaque service error leaked in response: %s", body)
	}
	if calls, _, _, _ := spy.snapshot(); calls != 1 {
		t.Fatalf("service calls = %d, want exactly one", calls)
	}
}

func TestMailboxSubmissionHandler_SourceContractHasNoProviderOrAsyncCapabilities(t *testing.T) {
	source, err := os.ReadFile("../../internal/mail/mailbox_submission_handler.go")
	if err != nil {
		t.Fatalf("read submission handler source: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"type mailboxSubmissionCreator interface",
		"func NewMailboxSubmissionHandler",
		"func (handler *MailboxSubmissionHandler) RegisterRoutes",
		"/drafts/:draftId/send",
		"mailboxActorID(ctx)",
		"mailboxSubmissionNormalizeIdempotencyKey",
		"mailboxDraftSaveHTTPExpectedRevision",
		".CreateMailboxSubmission(",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("submission handler source missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"MailEngine", "Provider", "Dispatcher", "Preflight", "NATS", "Calendar", "Storage", "GetMessage(", "GetBlob(", ".Submit(", ".Reconcile(", "go func",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("submission handler source exposes forbidden capability %q", forbidden)
		}
	}
	if calls := strings.Count(text, ".CreateMailboxSubmission("); calls != 1 {
		t.Fatalf("CreateMailboxSubmission call sites = %d, want exactly one", calls)
	}
}

type mailboxSubmissionHandlerResponse struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	MessageID *string    `json:"message_id"`
	ErrorCode *string    `json:"error_code"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type mailboxSubmissionHandlerRepositoryFake struct {
	err error
}

func (fake *mailboxSubmissionHandlerRepositoryFake) ReserveMailboxSubmission(context.Context, mail.MailboxSubmissionCreateRequest) (mail.MailboxSubmissionReservation, error) {
	return mail.MailboxSubmissionReservation{}, fake.err
}

var _ mail.MailboxSubmissionRepository = (*mailboxSubmissionHandlerRepositoryFake)(nil)

type mailboxSubmissionHandlerOwnerRepositoryFake struct {
	record mail.MailboxOwnerRecord
	err    error
}

func (fake *mailboxSubmissionHandlerOwnerRepositoryFake) ResolveMailboxOwner(context.Context, uuid.UUID) (mail.MailboxOwnerRecord, error) {
	return fake.record, fake.err
}

var _ mail.MailboxOwnerRepository = (*mailboxSubmissionHandlerOwnerRepositoryFake)(nil)

func mailboxSubmissionHandlerRealService(t *testing.T, repository mail.MailboxSubmissionRepository) *mail.MailboxSubmissionService {
	t.Helper()
	engineAccountID := "engine-account"
	primaryAddress := "sender@example.test"
	ownerRepository := &mailboxSubmissionHandlerOwnerRepositoryFake{record: mail.MailboxOwnerRecord{
		ActorID:         mailboxSubmissionHandlerActorID,
		OrganizationID:  uuid.MustParse("abf13d90-3230-470a-9fd8-745ecfe7d1f8"),
		AccountID:       uuid.MustParse("4aa6f9cb-63b1-4c1e-a41e-9139be205fd2"),
		EngineAccountID: &engineAccountID,
		State:           mail.MailboxOwnerActive,
		ProjectionState: mail.MailboxProjectionReady,
		PrimaryAddress:  &primaryAddress,
	}}
	owners, err := mail.NewMailboxService(ownerRepository)
	if err != nil {
		t.Fatalf("NewMailboxService: %v", err)
	}
	service, err := mail.NewMailboxSubmissionService(owners, repository)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionService: %v", err)
	}
	return service
}

func mailboxSubmissionHandlerApp(t *testing.T, service interface {
	CreateMailboxSubmission(context.Context, uuid.UUID, uuid.UUID, mail.MailboxSubmissionCreateInput) (mail.MailboxSubmissionView, error)
}, authenticated bool) *fiber.App {
	t.Helper()
	handler, err := mail.NewMailboxSubmissionHandler(service)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionHandler: %v", err)
	}
	app := fiber.New()
	api := app.Group("/api")
	auth := func(ctx fiber.Ctx) error {
		if authenticated {
			ctx.Locals("userID", mailboxSubmissionHandlerActorID.String())
		}
		return ctx.Next()
	}
	if err := handler.RegisterRoutes(api, auth); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	return app
}

func mailboxSubmissionHandlerRequest(t *testing.T, draftID string, body io.Reader) *http.Request {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/mail/drafts/"+draftID+"/send", body)
	request.Header.Set("Idempotency-Key", "send-key")
	request.Header.Set("If-Match", "\"7\"")
	return request
}

func mailboxSubmissionHandlerDo(t *testing.T, app *fiber.App, request *http.Request) *http.Response {
	t.Helper()
	response, err := app.Test(request, 5*time.Second)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return response
}

func mailboxSubmissionHandlerQueuedView() mail.MailboxSubmissionView {
	createdAt := time.Date(2026, time.August, 25, 9, 10, 11, 0, time.UTC)
	return mail.MailboxSubmissionView{ID: mailboxSubmissionHandlerReceiptID, Status: mail.MailboxSubmissionStatusQueued, CreatedAt: createdAt, UpdatedAt: createdAt}
}
