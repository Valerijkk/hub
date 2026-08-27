package mail_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	mailboxDraftCreateHandlerActorID        = uuid.MustParse("2e1c0ee5-5467-4b2e-ab9e-16bc5f42b205")
	mailboxDraftCreateHandlerDraftClientID  = uuid.MustParse("a5c6428e-627e-401a-8e72-702d68eeb34f")
	mailboxDraftCreateHandlerFromIdentityID = uuid.MustParse("b2b0f72b-f4bb-42c1-9ed0-36645c5c4a5c")
	mailboxDraftCreateHandlerDraftID        = uuid.MustParse("c8e85fa4-62ba-43ce-b1fd-448bd0ff56d2")
)

func TestMailboxDraftCreateHandler_CreateValidatesBodyFormatAndBodyPresence(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		bodyFormat  string
		bodyFields  string
		wantStatus  int
		wantCreate  bool
		wantContent mail.MailboxContentInput
		viewContent mail.MailboxContentResult
	}{
		{
			name:       "valid plain body",
			bodyFormat: "plain",
			bodyFields: `"plain_body":"Plain body"`,
			wantStatus: http.StatusCreated,
			wantCreate: true,
			wantContent: mail.MailboxContentInput{
				SourceFormat: mail.MailboxContentFormatPlain,
				Body:         "Plain body",
			},
			viewContent: mail.MailboxContentResult{
				SourceFormat:        mail.MailboxContentFormatPlain,
				SanitizerVersion:    mail.MailboxContentSanitizerVersionV1,
				PlainBody:           "Plain body",
				RemoteImagesBlocked: true,
			},
		},
		{
			name:       "valid rich body",
			bodyFormat: "rich",
			bodyFields: `"rich_body":"<p>Rich body</p>"`,
			wantStatus: http.StatusCreated,
			wantCreate: true,
			wantContent: mail.MailboxContentInput{
				SourceFormat: mail.MailboxContentFormatRich,
				Body:         "<p>Rich body</p>",
			},
			viewContent: mail.MailboxContentResult{
				SourceFormat:        mail.MailboxContentFormatRich,
				SanitizerVersion:    mail.MailboxContentSanitizerVersionV1,
				PlainBody:           "Rich body",
				SanitizedHTML:       "<p>Rich body</p>",
				RemoteImagesBlocked: true,
			},
		},
		{name: "unknown markdown with plain body", bodyFormat: "markdown", bodyFields: `"plain_body":"Plain body"`, wantStatus: http.StatusBadRequest},
		{name: "plain with rich body only", bodyFormat: "plain", bodyFields: `"rich_body":"<p>Rich body</p>"`, wantStatus: http.StatusBadRequest},
		{name: "rich with plain body only", bodyFormat: "rich", bodyFields: `"plain_body":"Plain body"`, wantStatus: http.StatusBadRequest},
		{name: "plain with both bodies", bodyFormat: "plain", bodyFields: `"plain_body":"Plain body","rich_body":"<p>Rich body</p>"`, wantStatus: http.StatusBadRequest},
		{name: "rich with both bodies", bodyFormat: "rich", bodyFields: `"plain_body":"Plain body","rich_body":"<p>Rich body</p>"`, wantStatus: http.StatusBadRequest},
		{name: "plain with null body", bodyFormat: "plain", bodyFields: `"plain_body":null`, wantStatus: http.StatusBadRequest},
		{name: "rich with null body", bodyFormat: "rich", bodyFields: `"rich_body":null`, wantStatus: http.StatusBadRequest},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			spy := &mailboxDraftCreateHandlerServiceSpy{view: mailboxDraftCreateHandlerView(testCase.viewContent)}
			app := mailboxDraftCreateHandlerApp(t, spy)
			response := mailboxDraftCreateHandlerDo(t, app, mailboxDraftCreateHandlerRequest(t, mailboxDraftCreateHandlerBody(testCase.bodyFormat, testCase.bodyFields)))
			defer response.Body.Close()

			if response.StatusCode != testCase.wantStatus {
				t.Fatalf("POST /api/mail/drafts status = %d, want %d", response.StatusCode, testCase.wantStatus)
			}

			createCalls, getCalls, saveCalls, inputs := spy.snapshot()
			if !testCase.wantCreate {
				if createCalls != 0 || getCalls != 0 || saveCalls != 0 || len(inputs) != 0 {
					t.Fatalf("rejected request calls = create:%d get:%d save:%d inputs:%#v, want no service calls", createCalls, getCalls, saveCalls, inputs)
				}
				return
			}
			if createCalls != 1 || getCalls != 0 || saveCalls != 0 || len(inputs) != 1 {
				t.Fatalf("accepted request calls = create:%d get:%d save:%d inputs:%#v, want one Create only", createCalls, getCalls, saveCalls, inputs)
			}
			if inputs[0].Content != testCase.wantContent {
				t.Fatalf("CreateMailboxDraft content = %#v, want %#v", inputs[0].Content, testCase.wantContent)
			}
		})
	}
}

type mailboxDraftCreateHandlerServiceSpy struct {
	mu          sync.Mutex
	view        mail.MailboxDraftView
	createCalls int
	getCalls    int
	saveCalls   int
	inputs      []mail.MailboxDraftCreateInput
}

func (spy *mailboxDraftCreateHandlerServiceSpy) CreateMailboxDraft(_ context.Context, _ uuid.UUID, input mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.createCalls++
	spy.inputs = append(spy.inputs, input)
	return spy.view, nil
}

func (spy *mailboxDraftCreateHandlerServiceSpy) GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (mail.MailboxDraftView, error) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.getCalls++
	return spy.view, nil
}

func (spy *mailboxDraftCreateHandlerServiceSpy) SaveMailboxDraft(context.Context, uuid.UUID, uuid.UUID, mail.MailboxDraftSaveInput) (mail.MailboxDraftView, error) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.saveCalls++
	return spy.view, nil
}

func (spy *mailboxDraftCreateHandlerServiceSpy) snapshot() (int, int, int, []mail.MailboxDraftCreateInput) {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return spy.createCalls, spy.getCalls, spy.saveCalls, append([]mail.MailboxDraftCreateInput(nil), spy.inputs...)
}

func mailboxDraftCreateHandlerApp(t *testing.T, service interface {
	CreateMailboxDraft(context.Context, uuid.UUID, mail.MailboxDraftCreateInput) (mail.MailboxDraftView, error)
	GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (mail.MailboxDraftView, error)
	SaveMailboxDraft(context.Context, uuid.UUID, uuid.UUID, mail.MailboxDraftSaveInput) (mail.MailboxDraftView, error)
}) *fiber.App {
	t.Helper()
	handler, err := mail.NewMailboxDraftCreateHandler(service)
	if err != nil {
		t.Fatalf("NewMailboxDraftCreateHandler: %v", err)
	}
	app := fiber.New()
	api := app.Group("/api")
	auth := func(ctx fiber.Ctx) error {
		ctx.Locals("userID", mailboxDraftCreateHandlerActorID.String())
		return ctx.Next()
	}
	if err := handler.RegisterRoutes(api, auth); err != nil {
		t.Fatalf("RegisterRoutes: %v", err)
	}
	return app
}

func mailboxDraftCreateHandlerRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/mail/drafts", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "draft-create-key")
	return request
}

func mailboxDraftCreateHandlerBody(bodyFormat, bodyFields string) string {
	return fmt.Sprintf(`{"draft_client_id":%q,"from_identity_id":%q,"to":[{"email":"recipient@example.test"}],"cc":[],"bcc":[],"subject":"Subject","body_format":%q,%s,"priority":"normal","attachment_ids":[]}`,
		mailboxDraftCreateHandlerDraftClientID.String(), mailboxDraftCreateHandlerFromIdentityID.String(), bodyFormat, bodyFields)
}

func mailboxDraftCreateHandlerDo(t *testing.T, app *fiber.App, request *http.Request) *http.Response {
	t.Helper()
	response, err := app.Test(request, 5*time.Second)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	return response
}

func mailboxDraftCreateHandlerView(content mail.MailboxContentResult) mail.MailboxDraftView {
	return mail.MailboxDraftView{
		ID:            mailboxDraftCreateHandlerDraftID,
		DraftClientID: mailboxDraftCreateHandlerDraftClientID,
		Revision:      1,
		Status:        mail.MailboxDraftStatusActive,
		Effective: mail.MailboxDraftEffectiveState{
			FromIdentityID: mailboxDraftCreateHandlerFromIdentityID,
			AttachmentIDs:  []uuid.UUID{},
		},
		Content: mail.MailboxDraftContentView{
			To:       []mail.EngineAddress{{Email: "recipient@example.test"}},
			CC:       []mail.EngineAddress{},
			BCC:      []mail.EngineAddress{},
			Subject:  "Subject",
			Priority: mail.EnginePriorityNormal,
			Content:  content,
		},
		UpdatedAt: time.Date(2026, time.August, 28, 9, 10, 11, 0, time.UTC),
	}
}
