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

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	bootstrapIdentityAccountID = "account-a"
	bootstrapIdentityEmail     = "ada@example.test"
	bootstrapIdentityToken     = "ephemeral-token_123"
)

func TestStalwartMailboxBootstrapIdentityProof_VerifyMailboxBootstrapIdentity(t *testing.T) {
	overlong := `{"methodResponses":[["Identity/get",{"accountId":"account-a","state":"state-a","list":[{"id":"identity-a","email":"ada@example.test"}],"notFound":[],"providerDetail":"` + strings.Repeat("x", 33*1024) + `"},"bootstrap-identity-get"]]}`

	for _, testCase := range []struct {
		name             string
		sessionStatus    int
		sessionAccountID string
		identityResponse string
		cancelled        bool
		wantErr          error
		wantRequests     int
		wantIdentityID   string
	}{
		{
			name:             "returns the exact sole normalized identity",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: bootstrapIdentityResponse(bootstrapIdentityAccountID, "state-a", `[{"id":"identity-a","email":"Ada@EXAMPLE.test"}]`, "[]"),
			wantRequests:     3,
			wantIdentityID:   "identity-a",
		},
		{
			name:          "rejects unauthorized mailbox discovery",
			sessionStatus: http.StatusUnauthorized,
			wantErr:       stalwart.ErrUnauthorized,
			wantRequests:  2,
		},
		{
			name:             "rejects a session without the exact account",
			sessionAccountID: "account-other",
			wantErr:          stalwart.ErrUnauthorized,
			wantRequests:     2,
		},
		{
			name:             "rejects a missing identity",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: bootstrapIdentityResponse(bootstrapIdentityAccountID, "state-a", "[]", "[]"),
			wantErr:          stalwart.ErrStateMismatch,
			wantRequests:     3,
		},
		{
			name:             "rejects duplicate expected identities",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: bootstrapIdentityResponse(bootstrapIdentityAccountID, "state-a", `[{"id":"identity-a","email":"ada@example.test"},{"id":"identity-b","email":"ada@example.test"}]`, "[]"),
			wantErr:          stalwart.ErrStateMismatch,
			wantRequests:     3,
		},
		{
			name:             "rejects an extra foreign identity",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: bootstrapIdentityResponse(bootstrapIdentityAccountID, "state-a", `[{"id":"identity-a","email":"ada@example.test"},{"id":"identity-b","email":"other@example.test"}]`, "[]"),
			wantErr:          stalwart.ErrStateMismatch,
			wantRequests:     3,
		},
		{
			name:             "rejects a malformed response",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: `{"methodResponses":not-json}`,
			wantErr:          stalwart.ErrProtocol,
			wantRequests:     3,
		},
		{
			name:             "rejects an oversized response",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: overlong,
			wantErr:          stalwart.ErrPayloadTooLarge,
			wantRequests:     3,
		},
		{
			name:             "rejects a foreign response account",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: bootstrapIdentityResponse("account-other", "state-a", `[{"id":"identity-a","email":"ada@example.test"}]`, "[]"),
			wantErr:          stalwart.ErrUnauthorized,
			wantRequests:     3,
		},
		{
			name:             "redacts a provider error",
			sessionAccountID: bootstrapIdentityAccountID,
			identityResponse: `{"methodResponses":[["error",{"type":"serverFail","description":"provider-secret-` + bootstrapIdentityToken + `"},"bootstrap-identity-get"]]}`,
			wantErr:          stalwart.ErrUnavailable,
			wantRequests:     3,
		},
		{
			name:          "does not issue requests after cancellation",
			cancelled:     true,
			wantErr:       stalwart.ErrCancelled,
			wantRequests:  0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newBootstrapIdentityHTTPFixture(t, testCase.sessionStatus, testCase.sessionAccountID, testCase.identityResponse)
			proof, err := mail.NewStalwartMailboxBootstrapIdentityProof(fixture.server.URL, fixture.server.Client(), bootstrapIdentityRegistryAuthorizer{calls: &fixture.registryCalls})
			if err != nil {
				t.Fatalf("NewStalwartMailboxBootstrapIdentityProof() error = %v", err)
			}

			ctx := context.Background()
			if testCase.cancelled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			result, err := proof.VerifyMailboxBootstrapIdentity(ctx, mail.MailboxBootstrapIdentityProofRequest{
				EngineAccountID:      bootstrapIdentityAccountID,
				ExpectedPrimaryEmail: bootstrapIdentityEmail,
				AccessToken:          bootstrapIdentityToken,
			})
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("VerifyMailboxBootstrapIdentity() error = %v; want %v", err, testCase.wantErr)
			}
			if result.EngineIdentityID != testCase.wantIdentityID {
				t.Fatalf("VerifyMailboxBootstrapIdentity() result = %#v; want identity %q", result, testCase.wantIdentityID)
			}
			if err != nil {
				assertBootstrapIdentityProofRedacted(t, err, fixture.server.URL)
			}
			fixture.require(t, testCase.wantRequests, testCase.wantRequests == 3)
		})
	}
}

type bootstrapIdentityHTTPFixture struct {
	server          *httptest.Server
	sessionStatus   int
	sessionAccount  string
	identityResponse string
	requests        []bootstrapIdentityHTTPRequest
	registryCalls   int
}

type bootstrapIdentityHTTPRequest struct {
	path          string
	authorization string
	body          string
}

func newBootstrapIdentityHTTPFixture(t *testing.T, sessionStatus int, sessionAccount, identityResponse string) *bootstrapIdentityHTTPFixture {
	t.Helper()
	fixture := &bootstrapIdentityHTTPFixture{sessionStatus: sessionStatus, sessionAccount: sessionAccount, identityResponse: identityResponse}
	fixture.server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		fixture.requests = append(fixture.requests, bootstrapIdentityHTTPRequest{
			path: request.URL.Path, authorization: request.Header.Get("Authorization"), body: string(body),
		})

		switch request.URL.Path {
		case "/.well-known/jmap":
			writer.Header().Set("Location", "/jmap/session")
			writer.WriteHeader(http.StatusTemporaryRedirect)
		case "/jmap/session":
			status := fixture.sessionStatus
			if status == 0 {
				status = http.StatusOK
			}
			if status != http.StatusOK {
				writer.WriteHeader(status)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(bootstrapIdentitySessionResponse(fixture.server.URL, fixture.sessionAccount)))
		case "/jmap/":
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(fixture.identityResponse))
		default:
			writer.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func (fixture *bootstrapIdentityHTTPFixture) require(t *testing.T, wantRequests int, wantIdentityCall bool) {
	t.Helper()
	if fixture.registryCalls != 0 {
		t.Fatalf("Registry authorizer calls = %d; want 0", fixture.registryCalls)
	}
	if len(fixture.requests) != wantRequests {
		t.Fatalf("HTTP requests = %#v; want %d", fixture.requests, wantRequests)
	}
	if wantRequests == 0 {
		return
	}
	if fixture.requests[0].path != "/.well-known/jmap" || fixture.requests[0].authorization != "" {
		t.Fatalf("well-known request = %#v; want unauthenticated exact discovery", fixture.requests[0])
	}
	if len(fixture.requests) > 1 && (fixture.requests[1].path != "/jmap/session" || fixture.requests[1].authorization != "Bearer "+bootstrapIdentityToken) {
		t.Fatalf("session request = %#v; want exact fixed bearer", fixture.requests[1])
	}
	if !wantIdentityCall {
		return
	}
	if fixture.requests[2].path != "/jmap/" || fixture.requests[2].authorization != "Bearer "+bootstrapIdentityToken {
		t.Fatalf("Identity/get request = %#v; want exact mailbox bearer", fixture.requests[2])
	}
	var envelope struct {
		Using       []string            `json:"using"`
		MethodCalls [][]json.RawMessage `json:"methodCalls"`
	}
	if err := json.Unmarshal([]byte(fixture.requests[2].body), &envelope); err != nil {
		t.Fatalf("Identity/get request JSON = %v", err)
	}
	if !equalBootstrapIdentityStrings(envelope.Using, []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail", "urn:ietf:params:jmap:submission"}) || len(envelope.MethodCalls) != 1 || len(envelope.MethodCalls[0]) != 3 {
		t.Fatalf("Identity/get envelope = %s; want one bounded mailbox call", fixture.requests[2].body)
	}
	var method, callID string
	if err := json.Unmarshal(envelope.MethodCalls[0][0], &method); err != nil {
		t.Fatalf("Identity/get method = %v", err)
	}
	if err := json.Unmarshal(envelope.MethodCalls[0][2], &callID); err != nil {
		t.Fatalf("Identity/get call id = %v", err)
	}
	var arguments struct {
		AccountID  string   `json:"accountId"`
		Properties []string `json:"properties"`
	}
	if err := json.Unmarshal(envelope.MethodCalls[0][1], &arguments); err != nil {
		t.Fatalf("Identity/get arguments = %v", err)
	}
	if method != "Identity/get" || callID != "bootstrap-identity-get" || arguments.AccountID != bootstrapIdentityAccountID || !equalBootstrapIdentityStrings(arguments.Properties, []string{"id", "email"}) {
		t.Fatalf("Identity/get = method=%q callID=%q arguments=%#v; want exact identity proof", method, callID, arguments)
	}
}

func bootstrapIdentitySessionResponse(origin, accountID string) string {
	encodedAccountID, _ := json.Marshal(accountID)
	encodedAPIURL, _ := json.Marshal(origin + "/jmap/")
	encodedUploadURL, _ := json.Marshal(origin + "/jmap/upload/{accountId}/")
	encodedDownloadURL, _ := json.Marshal(origin + "/jmap/download/{accountId}/{blobId}/{name}?type={type}")
	return `{"capabilities":{"urn:ietf:params:jmap:core":{},"urn:ietf:params:jmap:mail":{},"urn:ietf:params:jmap:submission":{}},"accounts":{` + string(encodedAccountID) + `:{}},"apiUrl":` + string(encodedAPIURL) + `,"uploadUrl":` + string(encodedUploadURL) + `,"downloadUrl":` + string(encodedDownloadURL) + `}`
}

func bootstrapIdentityResponse(accountID, state, list, notFound string) string {
	return `{"methodResponses":[["Identity/get",{"accountId":"` + accountID + `","state":"` + state + `","list":` + list + `,"notFound":` + notFound + `},"bootstrap-identity-get"]]}`
}

func assertBootstrapIdentityProofRedacted(t *testing.T, err error, origin string) {
	t.Helper()
	for _, forbidden := range []string{bootstrapIdentityToken, origin, "provider-secret", "other@example.test"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("VerifyMailboxBootstrapIdentity() error leaked %q in %q", forbidden, err)
		}
	}
}

func equalBootstrapIdentityStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

type bootstrapIdentityRegistryAuthorizer struct {
	calls *int
}

func (authorizer bootstrapIdentityRegistryAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	if authorizer.calls != nil {
		*authorizer.calls++
	}
	return stalwart.ErrUnauthorized
}
