package mail_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartBlobURLAccountID = "account/a"
	stalwartBlobURLBlobID    = "blob/b"
	stalwartBlobURLToken     = "mailbox-token"
)

func TestStalwartBlobURLs_EscapeOpaqueSegments(t *testing.T) {
	fixture := newStalwartBlobURLFixture(t)
	registryAuthorizer := &stalwartBlobURLRegistryAuthorizer{}
	mailboxAuthorizer := &stalwartBlobURLMailboxAuthorizer{token: stalwartBlobURLToken}
	client, err := stalwart.NewClient(stalwart.Config{
		Origin: fixture.server.URL, HTTPClient: fixture.server.Client(),
		RegistryAuthorizer: registryAuthorizer, MailboxAuthorizer: mailboxAuthorizer,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	session, err := client.DiscoverMailboxSession(context.Background())
	if err != nil {
		t.Fatalf("DiscoverMailboxSession() error = %v", err)
	}
	if !session.HasAccount(stalwartBlobURLAccountID) {
		t.Fatalf("DiscoverMailboxSession() did not retain account %q", stalwartBlobURLAccountID)
	}

	content := newStalwartBlobURLContextReadCloser("blob")
	uploaded, err := client.UploadMailboxBlob(context.Background(), stalwart.UploadRequest{
		Session: session, AccountID: stalwartBlobURLAccountID, ContentType: "application/octet-stream",
		Content: content, ContentBytes: 4,
	})
	if err != nil {
		t.Fatalf("UploadMailboxBlob() error = %v", err)
	}
	if uploaded != (stalwart.UploadResult{AccountID: stalwartBlobURLAccountID, BlobID: stalwartBlobURLBlobID, Type: "application/octet-stream", Size: 4}) {
		t.Fatalf("UploadMailboxBlob() result = %#v", uploaded)
	}
	if !content.WasClosed() {
		t.Fatal("UploadMailboxBlob() did not close its ContextReadCloser")
	}

	downloaded, err := client.DownloadMailboxBlob(context.Background(), stalwart.DownloadRequest{
		Session: session, AccountID: stalwartBlobURLAccountID, BlobID: stalwartBlobURLBlobID,
		Name: "file %.txt", Type: "text/plain; charset=utf-8", MaximumBytes: 16,
	})
	if err != nil {
		t.Fatalf("DownloadMailboxBlob() error = %v", err)
	}
	if string(downloaded.Content) != "blob" || downloaded.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("DownloadMailboxBlob() result = %#v", downloaded)
	}

	fixture.Require(t, registryAuthorizer.calls, mailboxAuthorizer.calls)
}

type stalwartBlobURLFixture struct {
	server   *httptest.Server
	requests []stalwartBlobURLRequest
}

type stalwartBlobURLRequest struct {
	method        string
	requestURI    string
	authorization string
	contentType   string
	body          string
}

func newStalwartBlobURLFixture(t *testing.T) *stalwartBlobURLFixture {
	t.Helper()
	fixture := &stalwartBlobURLFixture{}
	fixture.server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		fixture.requests = append(fixture.requests, stalwartBlobURLRequest{
			method: request.Method, requestURI: request.RequestURI,
			authorization: request.Header.Get("Authorization"), contentType: request.Header.Get("Content-Type"), body: string(body),
		})

		switch request.RequestURI {
		case "/.well-known/jmap":
			writer.Header().Set("Location", "/jmap/session")
			writer.WriteHeader(http.StatusTemporaryRedirect)
		case "/jmap/session":
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(stalwartBlobURLSessionResponse(fixture.server.URL)))
		case "/jmap/upload/account%2Fa/":
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"accountId":"account/a","blobId":"blob/b","type":"application/octet-stream","size":4}`))
		case "/jmap/download/account%2Fa/blob%2Fb/file%20%25.txt?type=text%2Fplain%3B+charset%3Dutf-8":
			writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = writer.Write([]byte("blob"))
		default:
			writer.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func (fixture *stalwartBlobURLFixture) Require(t *testing.T, registryCalls, mailboxCalls int) {
	t.Helper()
	if registryCalls != 0 {
		t.Fatalf("Registry authorizer calls = %d; want 0", registryCalls)
	}
	if mailboxCalls != 3 {
		t.Fatalf("Mailbox authorizer calls = %d; want 3", mailboxCalls)
	}
	if len(fixture.requests) != 4 {
		t.Fatalf("HTTP requests = %#v; want 4", fixture.requests)
	}

	if fixture.requests[0] != (stalwartBlobURLRequest{method: http.MethodGet, requestURI: "/.well-known/jmap"}) {
		t.Fatalf("well-known request = %#v; want unauthenticated discovery", fixture.requests[0])
	}
	if fixture.requests[1] != (stalwartBlobURLRequest{method: http.MethodGet, requestURI: "/jmap/session", authorization: "Bearer " + stalwartBlobURLToken}) {
		t.Fatalf("session request = %#v; want exact mailbox authorization", fixture.requests[1])
	}
	if fixture.requests[2] != (stalwartBlobURLRequest{
		method: http.MethodPost, requestURI: "/jmap/upload/account%2Fa/", authorization: "Bearer " + stalwartBlobURLToken,
		contentType: "application/octet-stream", body: "blob",
	}) {
		t.Fatalf("upload request = %#v; want exact escaped upload target and body", fixture.requests[2])
	}
	if fixture.requests[3] != (stalwartBlobURLRequest{
		method: http.MethodGet, requestURI: "/jmap/download/account%2Fa/blob%2Fb/file%20%25.txt?type=text%2Fplain%3B+charset%3Dutf-8",
		authorization: "Bearer " + stalwartBlobURLToken,
	}) {
		t.Fatalf("download request = %#v; want exact escaped download target", fixture.requests[3])
	}
}

func stalwartBlobURLSessionResponse(origin string) string {
	encodedAPIURL, _ := json.Marshal(origin + "/jmap/")
	encodedUploadURL, _ := json.Marshal(origin + "/jmap/upload/{accountId}/")
	encodedDownloadURL, _ := json.Marshal(origin + "/jmap/download/{accountId}/{blobId}/{name}?type={type}")
	return `{"capabilities":{"urn:ietf:params:jmap:core":{},"urn:ietf:params:jmap:mail":{},"urn:ietf:params:jmap:submission":{}},"accounts":{"account/a":{}},"apiUrl":` + string(encodedAPIURL) + `,"uploadUrl":` + string(encodedUploadURL) + `,"downloadUrl":` + string(encodedDownloadURL) + `}`
}

type stalwartBlobURLRegistryAuthorizer struct {
	calls int
}

func (authorizer *stalwartBlobURLRegistryAuthorizer) AuthorizeRegistry(context.Context, *http.Request) error {
	authorizer.calls++
	return stalwart.ErrUnauthorized
}

type stalwartBlobURLMailboxAuthorizer struct {
	token string
	calls int
}

func (authorizer *stalwartBlobURLMailboxAuthorizer) AuthorizeMailbox(ctx context.Context, request *http.Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	authorizer.calls++
	request.Header.Set("Authorization", "Bearer "+authorizer.token)
	return nil
}

type stalwartBlobURLContextReadCloser struct {
	mu     sync.Mutex
	reader *strings.Reader
	closed bool
}

func newStalwartBlobURLContextReadCloser(content string) *stalwartBlobURLContextReadCloser {
	return &stalwartBlobURLContextReadCloser{reader: strings.NewReader(content)}
}

func (source *stalwartBlobURLContextReadCloser) Read(buffer []byte) (int, error) {
	return source.ReadContext(context.Background(), buffer)
}

func (source *stalwartBlobURLContextReadCloser) ReadContext(ctx context.Context, buffer []byte) (int, error) {
	if ctx == nil {
		return 0, context.Canceled
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	source.mu.Lock()
	defer source.mu.Unlock()
	if source.closed {
		return 0, io.ErrClosedPipe
	}
	return source.reader.Read(buffer)
}

func (source *stalwartBlobURLContextReadCloser) Close() error {
	source.mu.Lock()
	defer source.mu.Unlock()
	source.closed = true
	return nil
}

func (source *stalwartBlobURLContextReadCloser) WasClosed() bool {
	source.mu.Lock()
	defer source.mu.Unlock()
	return source.closed
}
