package mail_test

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	mail "github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

func TestStalwartEnginePreflightSubmission_ReadOnlySourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/mail/stalwart_engine_submission_preflight.go")
	if err != nil {
		t.Fatalf("read preflight source: %v", err)
	}
	source := strings.ReplaceAll(string(sourceBytes), "\r\n", "\n")
	methodStart := strings.Index(source, "func (engine *StalwartEngine) PreflightSubmission")
	identityStart := strings.Index(source, "func (engine *StalwartEngine) stalwartPreflightSubmissionIdentity")
	if methodStart == -1 || identityStart == -1 || identityStart <= methodStart {
		t.Fatal("preflight and identity helper source markers are missing")
	}
	method := source[methodStart:identityStart]
	identity := source[identityStart:]
	for _, expected := range []string{
		"engine.GetDraft(ctx, EngineDraftReadRequest{",
		"preflightSubmissionForMessage(submission, message)",
		"stalwartPreflightSubmissionIdentity(ctx, submission.AccountID, preflight.FromEmail)",
	} {
		if !strings.Contains(method, expected) {
			t.Fatalf("preflight source is missing %q", expected)
		}
	}
	for _, expected := range []string{
		`Name: "Identity/get"`,
		`Properties: []string{"id", "email"}`,
		"canonical.Email != fromEmail",
	} {
		if !strings.Contains(identity, expected) {
			t.Fatalf("identity helper source is missing %q", expected)
		}
	}
	for _, forbidden := range []string{"EmailSubmission/get", "Email/set", "EmailSubmission/set", "UploadBlob", "DownloadBlob"} {
		if strings.Contains(method, forbidden) || strings.Contains(identity, forbidden) {
			t.Fatalf("read-only preflight source contains forbidden %q", forbidden)
		}
	}
}

func TestStalwartEnginePreflightSubmission_JMAPFixtureReadsExactIdentityAndBindsResult(t *testing.T) {
	submission := stalwartPreflightFixtureSubmission()
	client := &stalwartPreflightFixtureClient{identities: []stalwartPreflightFixtureIdentity{{ID: "identity-a", Email: "sender@example.test"}}}
	engine, err := mail.NewStalwartEngine(stalwartPreflightFixtureFactory{client: client})
	if err != nil {
		t.Fatalf("NewStalwartEngine: %v", err)
	}
	first, err := engine.PreflightSubmission(context.Background(), submission)
	if err != nil {
		t.Fatalf("PreflightSubmission: %v", err)
	}
	if first.FromEmail != "sender@example.test" || first.RecipientCount != 1 || first.MessageBytes < 1 || first.BindingHash == [sha256.Size]byte{} {
		t.Fatalf("PreflightSubmission = %#v, want bound exact local projection", first)
	}
	second, err := engine.PreflightSubmission(context.Background(), submission)
	if err != nil {
		t.Fatalf("PreflightSubmission repeat: %v", err)
	}
	if second != first {
		t.Fatalf("PreflightSubmission repeat = %#v, want deterministic %#v", second, first)
	}
	wantCalls := []string{"Mailbox/get", "Email/query", "Email/get", "Identity/get", "Mailbox/get", "Email/query", "Email/get", "Identity/get"}
	if strings.Join(client.calls, ",") != strings.Join(wantCalls, ",") {
		t.Fatalf("JMAP calls = %#v, want only exact read path %#v", client.calls, wantCalls)
	}
}

func TestStalwartEnginePreflightSubmission_JMAPFixtureRejectsUnavailableIdentity(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		identities []stalwartPreflightFixtureIdentity
	}{
		{name: "absent"},
		{name: "mismatch", identities: []stalwartPreflightFixtureIdentity{{ID: "identity-a", Email: "other@example.test"}}},
		{name: "duplicate", identities: []stalwartPreflightFixtureIdentity{{ID: "identity-a", Email: "sender@example.test"}, {ID: "identity-b", Email: "sender@example.test"}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			engine, err := mail.NewStalwartEngine(stalwartPreflightFixtureFactory{client: &stalwartPreflightFixtureClient{identities: testCase.identities}})
			if err != nil {
				t.Fatalf("NewStalwartEngine: %v", err)
			}
			_, err = engine.PreflightSubmission(context.Background(), stalwartPreflightFixtureSubmission())
			if !errors.Is(err, mail.ErrEngineIdentityUnavailable) {
				t.Fatalf("PreflightSubmission error = %v, want identity unavailable", err)
			}
		})
	}
}

type stalwartPreflightFixtureIdentity struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type stalwartPreflightFixtureClient struct {
	identities []stalwartPreflightFixtureIdentity
	calls      []string
}

func (client *stalwartPreflightFixtureClient) AccountID() string {
	return "account-a"
}

func (client *stalwartPreflightFixtureClient) CallMailbox(_ context.Context, request stalwart.JMAPRequest) (stalwart.JMAPResponse, error) {
	if len(request.MethodCalls) != 1 {
		return stalwart.JMAPResponse{}, errors.New("fixture requires one JMAP call")
	}
	method := request.MethodCalls[0]
	client.calls = append(client.calls, method.Name)
	var payload []byte
	switch method.Name {
	case "Mailbox/get":
		payload = []byte(`{"accountId":"account-a","state":"folders-a","list":[{"id":"drafts-a","name":"Drafts","role":"drafts"}],"notFound":[]}`)
	case "Email/query":
		payload = []byte(`{"accountId":"account-a","queryState":"query-a","ids":["message-a"],"position":0,"total":1}`)
	case "Email/get":
		payload = []byte(stalwartPreflightFixtureEmailJSON())
	case "Identity/get":
		var err error
		payload, err = json.Marshal(struct {
			AccountID string                             `json:"accountId"`
			State     string                             `json:"state"`
			List      []stalwartPreflightFixtureIdentity `json:"list"`
			NotFound  []string                           `json:"notFound"`
		}{AccountID: "account-a", State: "identities-a", List: client.identities, NotFound: []string{}})
		if err != nil {
			return stalwart.JMAPResponse{}, err
		}
	default:
		return stalwart.JMAPResponse{}, errors.New("unexpected JMAP method")
	}
	return stalwart.JMAPResponse{MethodResponses: []stalwart.JMAPMethodResponse{{Name: method.Name, Arguments: payload, CallID: method.CallID}}}, nil
}

func (*stalwartPreflightFixtureClient) UploadBlob(context.Context, mail.StalwartMailboxUploadRequest) (mail.StalwartMailboxUploadResult, error) {
	return mail.StalwartMailboxUploadResult{}, errors.New("upload must not be called by preflight")
}

func (*stalwartPreflightFixtureClient) DownloadBlob(context.Context, string, int64) (mail.StalwartMailboxBlob, error) {
	return mail.StalwartMailboxBlob{}, errors.New("download must not be called by preflight")
}

type stalwartPreflightFixtureFactory struct {
	client mail.StalwartMailboxClient
}

func (factory stalwartPreflightFixtureFactory) OpenMailboxClient(context.Context, string) (mail.StalwartMailboxClient, error) {
	return factory.client, nil
}

func stalwartPreflightFixtureSubmission() mail.EngineSubmission {
	return mail.EngineSubmission{AccountID: "account-a", DraftKey: "draft-a", MessageID: "message-a", ExpectedVersion: "message-state-a", SubmissionKey: "submission-a", MaxMessageBytes: 4096}
}

func stalwartPreflightFixtureEmailJSON() string {
	hash := sha256.New()
	hash.Write([]byte("hub-mail-marker-v1"))
	hash.Write([]byte{0})
	hash.Write([]byte("draft"))
	hash.Write([]byte{0})
	hash.Write([]byte("account-a"))
	hash.Write([]byte{0})
	hash.Write([]byte("draft-a"))
	marker := "hubk-" + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(nil)))
	return strings.ReplaceAll(`{"accountId":"account-a","state":"message-state-a","list":[{"id":"message-a","threadId":"thread-a","mailboxIds":{"drafts-a":true},"keywords":{"$draft":true,"__marker__":true},"size":4,"receivedAt":"2026-08-25T00:00:00Z","from":[{"email":"sender@example.test","name":"Sender"}],"to":[{"email":"recipient@example.test","name":"Recipient"}],"cc":[],"replyTo":[],"subject":"Subject","preview":"Body","attachments":[],"bodyStructure":{"partId":"text","type":"text/plain","size":4},"textBody":[{"partId":"text","type":"text/plain","size":4}],"htmlBody":[],"bodyValues":{"text":{"value":"Body"}}}],"notFound":[]}`, "__marker__", marker)
}
