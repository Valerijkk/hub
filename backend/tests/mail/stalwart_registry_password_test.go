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
	"sync"
	"testing"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	registryPasswordAccountID = "account-user-1"
	registryPasswordSecret    = "twølve visible"
)

type registryPasswordAuthorizer struct{}

func (registryPasswordAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	return nil
}

func (registryPasswordAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error {
	return nil
}

func TestRegistryPasswordClient_SetUserPassword(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		response      string
		cancelled     bool
		wantSentinel  error
		wantRequests  int
		assertRequest bool
	}{
		{
			name:          "success sends one minimal password update",
			response:      registryPasswordUpdatedResponse(registryPasswordAccountID),
			wantRequests:  1,
			assertRequest: true,
		},
		{
			name:         "wrong updated id is rejected",
			response:     registryPasswordUpdatedResponse("account-other"),
			wantSentinel: stalwart.ErrProtocol,
			wantRequests: 1,
		},
		{
			name: "not updated is rejected",
			response: `{"methodResponses":[["x:Account/set",{"notUpdated":{"account-user-1":{"type":"forbidden"}}},"set-user-password"]]}`,
			wantSentinel: stalwart.ErrProtocol,
			wantRequests: 1,
		},
		{
			name: "malformed updated response is rejected",
			response: `{"methodResponses":[["x:Account/set",{"updated":{"account-user-1":{}}},"set-user-password"]]}`,
			wantSentinel: stalwart.ErrProtocol,
			wantRequests: 1,
		},
		{
			name: "provider error is detail free",
			response: `{"methodResponses":[["error",{"type":"serverFail","description":"provider detail must not leak"},"set-user-password"]]}`,
			wantSentinel: stalwart.ErrIndeterminate,
			wantRequests: 1,
		},
		{
			name:         "cancelled mutation is indeterminate without provider request",
			cancelled:     true,
			wantSentinel: stalwart.ErrIndeterminate,
			wantRequests: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var requests int
			var requestLock sync.Mutex
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				requestLock.Lock()
				requests++
				requestLock.Unlock()
				if testCase.assertRequest {
					assertRegistryPasswordRequest(t, request)
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(writer, testCase.response)
			}))
			defer server.Close()

			client := newRegistryPasswordClient(t, server)
			secret := []byte(registryPasswordSecret)
			before := append([]byte(nil), secret...)
			ctx := context.Background()
			if testCase.cancelled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := client.SetUserPassword(ctx, registryPasswordAccountID, secret)
			if testCase.wantSentinel == nil && err != nil {
				t.Fatalf("SetUserPassword() error = %v", err)
			}
			if testCase.wantSentinel != nil && !errors.Is(err, testCase.wantSentinel) {
				t.Fatalf("SetUserPassword() error = %v; want %v", err, testCase.wantSentinel)
			}
			if err != nil && strings.Contains(err.Error(), registryPasswordSecret) {
				t.Fatalf("SetUserPassword() leaked secret in error %q", err)
			}
			if !bytes.Equal(secret, before) {
				t.Fatalf("SetUserPassword() modified caller password buffer: got %q, want %q", secret, before)
			}
			requestLock.Lock()
			gotRequests := requests
			requestLock.Unlock()
			if gotRequests != testCase.wantRequests {
				t.Fatalf("provider requests = %d; want %d", gotRequests, testCase.wantRequests)
			}
		})
	}
}

func TestRegistryPasswordClient_SetUserPassword_RejectsInvalidSecretWithoutCallingProvider(t *testing.T) {
	for _, secret := range [][]byte{
		[]byte("too-short"),
		[]byte("twelve\x00byte"),
		[]byte("twelve\x1fbyte"),
		append([]byte("eleven-byte"), 0xff),
	} {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			requests++
		}))
		client := newRegistryPasswordClient(t, server)
		before := append([]byte(nil), secret...)
		err := client.SetUserPassword(context.Background(), registryPasswordAccountID, secret)
		server.Close()
		if !errors.Is(err, stalwart.ErrInvalidRequest) {
			t.Fatalf("SetUserPassword(%q) error = %v; want invalid request", secret, err)
		}
		if requests != 0 {
			t.Fatalf("SetUserPassword(%q) provider requests = %d; want 0", secret, requests)
		}
		if !bytes.Equal(secret, before) {
			t.Fatalf("SetUserPassword(%q) modified caller password buffer", secret)
		}
	}
}

func TestRegistryPasswordClient_SetUserPassword_RejectsInvalidAccountIDWithoutCallingProvider(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	defer server.Close()
	client := newRegistryPasswordClient(t, server)
	secret := []byte(registryPasswordSecret)
	before := append([]byte(nil), secret...)
	err := client.SetUserPassword(context.Background(), "account\ninvalid", secret)
	if !errors.Is(err, stalwart.ErrInvalidRequest) || requests != 0 || !bytes.Equal(secret, before) {
		t.Fatalf("SetUserPassword() = %v, requests=%d, secret preserved=%t; want invalid request, no request, preserved secret", err, requests, bytes.Equal(secret, before))
	}
}

func newRegistryPasswordClient(t *testing.T, server *httptest.Server) *stalwart.RegistryPasswordClient {
	t.Helper()
	transport, err := stalwart.NewClient(stalwart.Config{
		Origin:             server.URL,
		HTTPClient:         server.Client(),
		RegistryAuthorizer: registryPasswordAuthorizer{},
		MailboxAuthorizer:  registryPasswordAuthorizer{},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	client, err := stalwart.NewRegistryPasswordClient(transport)
	if err != nil {
		t.Fatalf("NewRegistryPasswordClient() error = %v", err)
	}
	return client
}

func assertRegistryPasswordRequest(t *testing.T, request *http.Request) {
	t.Helper()
	if request.Method != http.MethodPost || request.URL.Path != "/jmap" || request.URL.RawQuery != "" || request.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("request = %s %s content-type=%q; want POST /jmap application/json", request.Method, request.URL.String(), request.Header.Get("Content-Type"))
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	var envelope struct {
		Using       []string            `json:"using"`
		MethodCalls [][]json.RawMessage `json:"methodCalls"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if !equalStrings(envelope.Using, []string{"urn:ietf:params:jmap:core", "urn:stalwart:jmap"}) || len(envelope.MethodCalls) != 1 || len(envelope.MethodCalls[0]) != 3 {
		t.Fatalf("request envelope = %s; want one exact Registry method", body)
	}
	var method, callID string
	if err := json.Unmarshal(envelope.MethodCalls[0][0], &method); err != nil {
		t.Fatalf("decode method: %v", err)
	}
	if err := json.Unmarshal(envelope.MethodCalls[0][2], &callID); err != nil {
		t.Fatalf("decode call id: %v", err)
	}
	if method != "x:Account/set" || callID != "set-user-password" {
		t.Fatalf("method=%q callID=%q; want x:Account/set set-user-password", method, callID)
	}
	var arguments map[string]json.RawMessage
	if err := json.Unmarshal(envelope.MethodCalls[0][1], &arguments); err != nil || len(arguments) != 1 {
		t.Fatalf("arguments = %s; want only update", envelope.MethodCalls[0][1])
	}
	update, ok := arguments["update"]
	if !ok {
		t.Fatal("request has no update")
	}
	var updates map[string]json.RawMessage
	if err := json.Unmarshal(update, &updates); err != nil || len(updates) != 1 {
		t.Fatalf("update = %s; want one account", update)
	}
	account, ok := updates[registryPasswordAccountID]
	if !ok {
		t.Fatalf("update lacks exact account id %q", registryPasswordAccountID)
	}
	var change map[string]json.RawMessage
	if err := json.Unmarshal(account, &change); err != nil || len(change) != 1 {
		t.Fatalf("account update = %s; want only credentials", account)
	}
	credentials, ok := change["credentials"]
	if !ok {
		t.Fatal("account update has no credentials")
	}
	var passwordSlots map[string]json.RawMessage
	if err := json.Unmarshal(credentials, &passwordSlots); err != nil || len(passwordSlots) != 1 {
		t.Fatalf("credentials = %s; want one Password credential", credentials)
	}
	password, ok := passwordSlots["0"]
	if !ok {
		t.Fatal("credentials has no exact Password slot")
	}
	var credential map[string]json.RawMessage
	if err := json.Unmarshal(password, &credential); err != nil || len(credential) != 2 {
		t.Fatalf("credential = %s; want only @type and secret", password)
	}
	var kind, secret string
	if err := json.Unmarshal(credential["@type"], &kind); err != nil {
		t.Fatalf("decode credential type: %v", err)
	}
	if err := json.Unmarshal(credential["secret"], &secret); err != nil {
		t.Fatalf("decode credential secret: %v", err)
	}
	if kind != "Password" || secret != registryPasswordSecret {
		t.Fatalf("credential type=%q secret=%q; want Password and exact caller secret", kind, secret)
	}
}

func registryPasswordUpdatedResponse(accountID string) string {
	return `{"methodResponses":[["x:Account/set",{"updated":{` + strconvQuote(accountID) + `:null}},"set-user-password"]]}`
}

func strconvQuote(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func equalStrings(left, right []string) bool {
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
