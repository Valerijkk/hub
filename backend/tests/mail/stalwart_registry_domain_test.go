package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const localRegistryDomain = "example.test"

type registryDomainTestAuthorizer struct{}

func (registryDomainTestAuthorizer) AuthorizeRegistry(_ context.Context, request *http.Request) error {
	request.Header.Set("Authorization", "Bearer registry-domain-test-key")
	return nil
}

func (registryDomainTestAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error {
	return errors.New("mailbox authorization must not be used")
}

type registryDomainReply struct {
	want     stalwart.JMAPRequest
	response string
}

func TestRegistryDomainController_EnsureLocalDomain(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		domain    string
		replies   []registryDomainReply
		cancel    bool
		wantID    string
		wantErr   error
		wantCalls int
	}{
		{
			name:   "present local domain",
			domain: " Example.TEST ",
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReply("domain-1", localRegistryDomain, true, false),
			},
			wantID:    "domain-1",
			wantCalls: 2,
		},
		{
			name:   "present local domain accepts enriched automatic DKIM management",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": map[string]any{
						"@type":            "Automatic",
						"algorithms":       map[string]bool{"Dkim1Ed25519Sha256": true, "Dkim1RsaSha256": true},
						"selectorTemplate": "mail-{n}",
						"rotateAfter":      int64(2_592_000_000),
						"retireAfter":      int64(86_400_000),
						"deleteAfter":      int64(172_800_000),
					},
					"dnsManagement": registryDomainManualDNSManagement(),
				}),
			},
			wantID:    "domain-1",
			wantCalls: 2,
		},
		{
			name:   "absent local domain is created then verified",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{}),
				registryDomainCreateReply("domain-new"),
				registryDomainGetReply("domain-new", localRegistryDomain, true, false),
			},
			wantID:    "domain-new",
			wantCalls: 3,
		},
		{
			name:   "non local existing domain is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReply("domain-1", localRegistryDomain, true, true),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain without DKIM management is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dnsManagement": registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain with null DKIM management is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": nil,
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain with manual DKIM management is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": map[string]any{"@type": "Manual"},
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain with automatic DNS management is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": registryDomainAutomaticDKIMManagement(),
					"dnsManagement":  map[string]any{"@type": "Automatic"},
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain with unknown DKIM management property is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": map[string]any{"@type": "Automatic", "unexpected": true},
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain with unknown DKIM algorithm is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": map[string]any{"@type": "Automatic", "algorithms": map[string]bool{"DkimUnknown": true}},
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain with disabled DKIM algorithm is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": map[string]any{"@type": "Automatic", "algorithms": map[string]bool{"Dkim1Ed25519Sha256": false}},
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "existing domain with malformed optional DKIM selector is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainGetReplyWithManagement("domain-1", localRegistryDomain, true, false, map[string]any{
					"dkimManagement": map[string]any{"@type": "Automatic", "selectorTemplate": true},
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "malformed query response is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{{
				want:     registryDomainQueryRequest(localRegistryDomain),
				response: `{"methodResponses":[["x:Domain/query",{"queryState":"query-state","canCalculateChanges":false,"position":0,"ids":["domain-1"]},"wrong-call"]]}`,
			}},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 1,
		},
		{
			name:   "empty registry response account id is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{{
				want:     registryDomainQueryRequest(localRegistryDomain),
				response: `{"methodResponses":[["x:Domain/query",{"accountId":"","queryState":"query-state","canCalculateChanges":false,"position":0,"ids":[]},"domain-query"]]}`,
			}},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 1,
		},
		{
			name:   "oversize registry response account id is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{{
				want:     registryDomainQueryRequest(localRegistryDomain),
				response: registryDomainOversizeAccountIDQueryResponse(),
			}},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 1,
		},
		{
			name:   "duplicate domain query is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1", "domain-2"}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 1,
		},
		{
			name:   "not created is indeterminate",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{}),
				registryDomainNotCreatedReply(),
			},
			wantErr:   stalwart.ErrIndeterminate,
			wantCalls: 2,
		},
		{
			name:   "oversize method arguments are rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{{
				want:     registryDomainQueryRequest(localRegistryDomain),
				response: registryDomainOversizeQueryResponse(),
			}},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 1,
		},
		{
			name:   "provider method error remains generic",
			domain: localRegistryDomain,
			replies: []registryDomainReply{{
				want:     registryDomainQueryRequest(localRegistryDomain),
				response: `{"methodResponses":[["error",{"type":"serverFail","detail":"must never surface"},"domain-query"]]}`,
			}},
			wantErr:   stalwart.ErrUnavailable,
			wantCalls: 1,
		},
		{
			name:      "cancelled context performs no request",
			domain:    localRegistryDomain,
			cancel:    true,
			wantErr:   stalwart.ErrCancelled,
			wantCalls: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				mu    sync.Mutex
				calls int
			)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				calls++
				if request.Method != http.MethodPost || request.URL.Path != "/jmap" || request.URL.RawQuery != "" || request.Header.Get("Authorization") != "Bearer registry-domain-test-key" {
					writer.WriteHeader(http.StatusForbidden)
					return
				}
				if calls > len(testCase.replies) {
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}

				var got stalwart.JMAPRequest
				if err := json.NewDecoder(request.Body).Decode(&got); err != nil || !registryDomainRequestEqual(got, testCase.replies[calls-1].want) {
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(testCase.replies[calls-1].response))
			}))
			defer server.Close()

			client, err := stalwart.NewClient(stalwart.Config{
				Origin:             server.URL,
				HTTPClient:         server.Client(),
				RegistryAuthorizer: registryDomainTestAuthorizer{},
				MailboxAuthorizer:  registryDomainTestAuthorizer{},
			})
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			controller, err := stalwart.NewRegistryDomainController(client)
			if err != nil {
				t.Fatalf("NewRegistryDomainController() error = %v", err)
			}

			ctx := context.Background()
			if testCase.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			gotID, err := controller.EnsureLocalDomain(ctx, testCase.domain)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("EnsureLocalDomain() error = %v; want %v", err, testCase.wantErr)
			}
			if err != nil && strings.Contains(err.Error(), "must never surface") {
				t.Fatalf("EnsureLocalDomain() exposed provider detail in %q", err)
			}
			if err == nil && gotID != testCase.wantID {
				t.Fatalf("EnsureLocalDomain() id = %q; want %q", gotID, testCase.wantID)
			}
			mu.Lock()
			defer mu.Unlock()
			if calls != testCase.wantCalls {
				t.Fatalf("requests = %d; want %d", calls, testCase.wantCalls)
			}
		})
	}
}

func TestRegistryDomainController_ReadLocalDomainDNSZoneFile(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		domain    string
		replies   []registryDomainReply
		wantID    string
		wantName  string
		wantErr   error
		wantCalls int
	}{
		{
			name:   "reads exact enabled non-relaying domain zone",
			domain: " Example.TEST ",
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReply("domain-1", localRegistryDomain, true, false, "$ORIGIN example.test.\n@ IN MX 10 mx.example.test.\n"),
			},
			wantID:    "domain-1",
			wantName:  localRegistryDomain,
			wantCalls: 2,
		},
		{
			name:   "zone read accepts enriched automatic DKIM management",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReplyWithManagement("domain-1", localRegistryDomain, true, false, "example.test. IN MX 10 mail.example.test.\n", map[string]any{
					"dkimManagement": map[string]any{
						"@type":            "Automatic",
						"algorithms":       map[string]bool{"Dkim1Ed25519Sha256": true, "Dkim1RsaSha256": true},
						"selectorTemplate": "mail-{n}",
						"rotateAfter":      int64(2_592_000_000),
						"retireAfter":      int64(86_400_000),
						"deleteAfter":      int64(172_800_000),
					},
					"dnsManagement": registryDomainManualDNSManagement(),
				}),
			},
			wantID:    "domain-1",
			wantName:  localRegistryDomain,
			wantCalls: 2,
		},
		{
			name:   "zone read rejects drifted DKIM management",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReplyWithManagement("domain-1", localRegistryDomain, true, false, "example.test.", map[string]any{
					"dkimManagement": map[string]any{"@type": "Manual"},
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "missing zone is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileMissingReply("domain-1", localRegistryDomain, true, false),
			},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 2,
		},
		{
			name:   "empty zone is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReply("domain-1", localRegistryDomain, true, false, ""),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "non string zone is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReply("domain-1", localRegistryDomain, true, false, true),
			},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 2,
		},
		{
			name:   "control zone is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReply("domain-1", localRegistryDomain, true, false, "example.test.\x00"),
			},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 2,
		},
		{
			name:   "oversize zone response is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReply("domain-1", localRegistryDomain, true, false, strings.Repeat("z", 64<<10)),
			},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 2,
		},
		{
			name:   "mismatched domain read-back is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReplyForRequest("domain-1", "domain-other", localRegistryDomain, true, false, "example.test."),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				mu    sync.Mutex
				calls int
			)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				calls++
				if request.Method != http.MethodPost || request.URL.Path != "/jmap" || request.URL.RawQuery != "" || request.Header.Get("Authorization") != "Bearer registry-domain-test-key" || calls > len(testCase.replies) {
					writer.WriteHeader(http.StatusForbidden)
					return
				}
				var got stalwart.JMAPRequest
				if err := json.NewDecoder(request.Body).Decode(&got); err != nil || !registryDomainRequestEqual(got, testCase.replies[calls-1].want) {
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(testCase.replies[calls-1].response))
			}))
			defer server.Close()

			client, err := stalwart.NewClient(stalwart.Config{Origin: server.URL, HTTPClient: server.Client(), RegistryAuthorizer: registryDomainTestAuthorizer{}, MailboxAuthorizer: registryDomainTestAuthorizer{}})
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			controller, err := stalwart.NewRegistryDomainController(client)
			if err != nil {
				t.Fatalf("NewRegistryDomainController() error = %v", err)
			}

			result, err := controller.ReadLocalDomainDNSZoneFile(context.Background(), testCase.domain)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("ReadLocalDomainDNSZoneFile() error = %v; want %v", err, testCase.wantErr)
			}
			if err == nil && (result.DomainID() != testCase.wantID || result.DomainName() != testCase.wantName) {
				t.Fatalf("ReadLocalDomainDNSZoneFile() identity = %q/%q; want %q/%q", result.DomainID(), result.DomainName(), testCase.wantID, testCase.wantName)
			}
			mu.Lock()
			defer mu.Unlock()
			if calls != testCase.wantCalls {
				t.Fatalf("requests = %d; want %d", calls, testCase.wantCalls)
			}
		})
	}
}

func TestRegistryDomainController_ObserveLocalDomainDNS(t *testing.T) {
	const projectionCapableZone = "example.test. IN MX 10 mail.example.test.\n" +
		"example.test. IN TXT \"v=spf1 -all\"\n" +
		"_dmarc.example.test. IN TXT \"v=DMARC1; p=none\"\n" +
		"selector._domainkey.example.test. IN TXT \"v=DKIM1; p=key\"\n"

	for _, testCase := range []struct {
		name           string
		domain         string
		replies        []registryDomainReply
		cancel         bool
		wantPresent    bool
		wantID         string
		wantName       string
		wantProjection bool
		wantErr        error
		wantCalls      int
	}{
		{
			name:      "confirmed absence requires one query",
			domain:    " Example.TEST ",
			replies:   []registryDomainReply{registryDomainQueryReply([]string{})},
			wantCalls: 1,
		},
		{
			name:   "present authoritative zone remains private and projection capable",
			domain: " Example.TEST ",
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReply("domain-1", localRegistryDomain, true, false, projectionCapableZone),
			},
			wantPresent:    true,
			wantID:         "domain-1",
			wantName:       localRegistryDomain,
			wantProjection: true,
			wantCalls:      2,
		},
		{
			name:   "duplicate query identifiers are rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1", "domain-1"}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 1,
		},
		{
			name:   "multiple query identifiers are rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1", "domain-2"}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 1,
		},
		{
			name:   "management drift is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{
				registryDomainQueryReply([]string{"domain-1"}),
				registryDomainDNSZoneFileReplyWithManagement("domain-1", localRegistryDomain, true, false, projectionCapableZone, map[string]any{
					"dkimManagement": map[string]any{"@type": "Manual"},
					"dnsManagement":  registryDomainManualDNSManagement(),
				}),
			},
			wantErr:   stalwart.ErrStateMismatch,
			wantCalls: 2,
		},
		{
			name:   "malformed query response is rejected",
			domain: localRegistryDomain,
			replies: []registryDomainReply{{
				want:     registryDomainQueryRequest(localRegistryDomain),
				response: `{"methodResponses":[["x:Domain/query",{"queryState":"query-state","canCalculateChanges":false,"position":0,"ids":["domain-1"]},"wrong-call"]]}`,
			}},
			wantErr:   stalwart.ErrProtocol,
			wantCalls: 1,
		},
		{
			name:      "cancelled context performs no request",
			domain:    localRegistryDomain,
			cancel:    true,
			wantErr:   stalwart.ErrCancelled,
			wantCalls: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				mu    sync.Mutex
				calls int
			)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				calls++
				if request.Method != http.MethodPost || request.URL.Path != "/jmap" || request.URL.RawQuery != "" || request.Header.Get("Authorization") != "Bearer registry-domain-test-key" || calls > len(testCase.replies) {
					writer.WriteHeader(http.StatusForbidden)
					return
				}
				var got stalwart.JMAPRequest
				if err := json.NewDecoder(request.Body).Decode(&got); err != nil || !registryDomainRequestEqual(got, testCase.replies[calls-1].want) {
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(testCase.replies[calls-1].response))
			}))
			defer server.Close()

			client, err := stalwart.NewClient(stalwart.Config{Origin: server.URL, HTTPClient: server.Client(), RegistryAuthorizer: registryDomainTestAuthorizer{}, MailboxAuthorizer: registryDomainTestAuthorizer{}})
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			controller, err := stalwart.NewRegistryDomainController(client)
			if err != nil {
				t.Fatalf("NewRegistryDomainController() error = %v", err)
			}

			ctx := context.Background()
			if testCase.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			observation, err := controller.ObserveLocalDomainDNS(ctx, testCase.domain)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("ObserveLocalDomainDNS() error = %v; want %v", err, testCase.wantErr)
			}
			if err == nil && observation.IsPresent() != testCase.wantPresent {
				t.Fatalf("IsPresent() = %t; want %t", observation.IsPresent(), testCase.wantPresent)
			}
			if err != nil && observation.IsPresent() {
				t.Fatal("failed observation must not be present")
			}

			zoneFile, present := observation.DNSZoneFile()
			if present != (err == nil && testCase.wantPresent) {
				t.Fatalf("DNSZoneFile() present = %t; want %t", present, err == nil && testCase.wantPresent)
			}
			if present && (zoneFile.DomainID() != testCase.wantID || zoneFile.DomainName() != testCase.wantName) {
				t.Fatalf("DNSZoneFile() identity = %q/%q; want %q/%q", zoneFile.DomainID(), zoneFile.DomainName(), testCase.wantID, testCase.wantName)
			}
			if testCase.wantProjection {
				projection, projectionErr := zoneFile.ProjectMailCriticalDNS()
				if projectionErr != nil {
					t.Fatalf("ProjectMailCriticalDNS() error = %v", projectionErr)
				}
				if projection.DomainID() != testCase.wantID || projection.DomainName() != testCase.wantName {
					t.Fatalf("ProjectMailCriticalDNS() identity = %q/%q; want %q/%q", projection.DomainID(), projection.DomainName(), testCase.wantID, testCase.wantName)
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if calls != testCase.wantCalls {
				t.Fatalf("requests = %d; want %d", calls, testCase.wantCalls)
			}
		})
	}

	var zero stalwart.LocalDomainDNSObservation
	if zero.IsPresent() {
		t.Fatal("zero observation must not be present")
	}
	if zoneFile, present := zero.DNSZoneFile(); present || zoneFile.DomainID() != "" || zoneFile.DomainName() != "" {
		t.Fatalf("zero DNSZoneFile() = %#v, %t; want zero, false", zoneFile, present)
	}
}

func registryDomainQueryReply(ids []string) registryDomainReply {
	payload, _ := json.Marshal(map[string]any{
		"methodResponses": []any{[]any{"x:Domain/query", map[string]any{
			"accountId":           "0",
			"queryState":          "query-state",
			"canCalculateChanges": false,
			"position":            0,
			"ids":                 ids,
		}, "domain-query"}},
	})
	return registryDomainReply{want: registryDomainQueryRequest(localRegistryDomain), response: string(payload)}
}

func registryDomainGetReply(id, name string, isEnabled, allowRelaying bool) registryDomainReply {
	return registryDomainGetReplyWithManagement(id, name, isEnabled, allowRelaying, map[string]any{
		"dkimManagement": registryDomainAutomaticDKIMManagement(),
		"dnsManagement":  registryDomainManualDNSManagement(),
	})
}

func registryDomainGetReplyWithManagement(id, name string, isEnabled, allowRelaying bool, management map[string]any) registryDomainReply {
	domain := map[string]any{
		"id":            id,
		"name":          name,
		"isEnabled":     isEnabled,
		"allowRelaying": allowRelaying,
	}
	for key, value := range management {
		domain[key] = value
	}
	payload, _ := json.Marshal(map[string]any{
		"methodResponses": []any{[]any{"x:Domain/get", map[string]any{
			"accountId": "0",
			"state":     "get-state",
			"list":      []any{domain},
			"notFound":  []any{},
		}, "domain-get"}},
	})
	return registryDomainReply{want: registryDomainGetRequest(id), response: string(payload)}
}

func registryDomainAutomaticDKIMManagement() map[string]any {
	return map[string]any{"@type": "Automatic"}
}

func registryDomainManualDNSManagement() map[string]any {
	return map[string]any{"@type": "Manual"}
}

func registryDomainDNSZoneFileReply(id, name string, isEnabled, allowRelaying bool, zone any) registryDomainReply {
	return registryDomainDNSZoneFileReplyWithManagement(id, name, isEnabled, allowRelaying, zone, map[string]any{
		"dkimManagement": registryDomainAutomaticDKIMManagement(),
		"dnsManagement":  registryDomainManualDNSManagement(),
	})
}

func registryDomainDNSZoneFileReplyWithManagement(id, name string, isEnabled, allowRelaying bool, zone any, management map[string]any) registryDomainReply {
	domain := map[string]any{
		"id":            id,
		"name":          name,
		"isEnabled":     isEnabled,
		"allowRelaying": allowRelaying,
		"dnsZoneFile":   zone,
	}
	for key, value := range management {
		domain[key] = value
	}
	payload, _ := json.Marshal(map[string]any{
		"methodResponses": []any{[]any{"x:Domain/get", map[string]any{
			"accountId": "0",
			"state":     "domain-zone-get-state",
			"list":      []any{domain},
			"notFound":  []any{},
		}, "domain-zone-get"}},
	})
	return registryDomainReply{want: registryDomainDNSZoneFileGetRequest(id), response: string(payload)}
}

func registryDomainDNSZoneFileReplyForRequest(requestID, responseID, name string, isEnabled, allowRelaying bool, zone any) registryDomainReply {
	reply := registryDomainDNSZoneFileReply(responseID, name, isEnabled, allowRelaying, zone)
	reply.want = registryDomainDNSZoneFileGetRequest(requestID)
	return reply
}

func registryDomainDNSZoneFileMissingReply(id, name string, isEnabled, allowRelaying bool) registryDomainReply {
	return registryDomainDNSZoneFileReplyWithManagement(id, name, isEnabled, allowRelaying, nil, map[string]any{
		"dkimManagement": registryDomainAutomaticDKIMManagement(),
		"dnsManagement":  registryDomainManualDNSManagement(),
	})
}

func registryDomainCreateReply(id string) registryDomainReply {
	payload, _ := json.Marshal(map[string]any{
		"methodResponses": []any{[]any{"x:Domain/set", map[string]any{
			"accountId": "0",
			"oldState":  "set-state-before",
			"newState":  "set-state-after",
			"created":   map[string]any{"create-local-domain": map[string]any{"id": id}},
		}, "domain-set"}},
	})
	return registryDomainReply{want: registryDomainCreateRequest(localRegistryDomain), response: string(payload)}
}

func registryDomainNotCreatedReply() registryDomainReply {
	return registryDomainReply{
		want:     registryDomainCreateRequest(localRegistryDomain),
		response: `{"methodResponses":[["x:Domain/set",{"accountId":"0","oldState":"set-state-before","newState":"set-state-after","notCreated":{"create-local-domain":{"type":"invalidProperties","detail":"must never surface"}}},"domain-set"]]}`,
	}
}

func registryDomainOversizeQueryResponse() string {
	payload, _ := json.Marshal(map[string]any{
		"methodResponses": []any{[]any{"x:Domain/query", map[string]any{
			"accountId":           "0",
			"queryState":          strings.Repeat("q", 17<<10),
			"canCalculateChanges": false,
			"position":            0,
			"ids":                 []string{},
		}, "domain-query"}},
	})
	return string(payload)
}

func registryDomainOversizeAccountIDQueryResponse() string {
	payload, _ := json.Marshal(map[string]any{
		"methodResponses": []any{[]any{"x:Domain/query", map[string]any{
			"accountId":           strings.Repeat("a", 1025),
			"queryState":          "query-state",
			"canCalculateChanges": false,
			"position":            0,
			"ids":                 []string{},
		}, "domain-query"}},
	})
	return string(payload)
}

func registryDomainQueryRequest(name string) stalwart.JMAPRequest {
	return registryDomainRequest("x:Domain/query", json.RawMessage(`{"filter":{"name":"`+name+`"},"position":0,"limit":2}`), "domain-query")
}

func registryDomainGetRequest(id string) stalwart.JMAPRequest {
	return registryDomainRequest("x:Domain/get", json.RawMessage(`{"ids":["`+id+`"],"properties":["name","isEnabled","allowRelaying","dkimManagement","dnsManagement"]}`), "domain-get")
}

func registryDomainDNSZoneFileGetRequest(id string) stalwart.JMAPRequest {
	return registryDomainRequest("x:Domain/get", json.RawMessage(`{"ids":["`+id+`"],"properties":["name","isEnabled","allowRelaying","dkimManagement","dnsManagement","dnsZoneFile"]}`), "domain-zone-get")
}

func registryDomainCreateRequest(name string) stalwart.JMAPRequest {
	return registryDomainRequest("x:Domain/set", json.RawMessage(`{"create":{"create-local-domain":{"name":"`+name+`","aliases":{},"isEnabled":true,"allowRelaying":false,"certificateManagement":{"@type":"Manual"},"dkimManagement":{"@type":"Automatic"},"dnsManagement":{"@type":"Manual"},"subAddressing":{"@type":"Enabled"}}}}`), "domain-set")
}

func registryDomainRequest(method string, arguments json.RawMessage, callID string) stalwart.JMAPRequest {
	return stalwart.JMAPRequest{
		Using: []string{"urn:ietf:params:jmap:core", "urn:stalwart:jmap"},
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      method,
			Arguments: arguments,
			CallID:    callID,
		}},
	}
}

func registryDomainRequestEqual(left, right stalwart.JMAPRequest) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}
