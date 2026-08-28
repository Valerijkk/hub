package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

func TestMailboxRegistryProvisioningEngineProjectsRegistryResponses(t *testing.T) {
	fixture := newRegistryProvisioningFixture(t, registryProvisioningHappyReplies(), nil)
	account, err := fixture.engine.EnsureAccount(context.Background(), registryAccountSpec())
	if err != nil {
		t.Fatalf(`EnsureAccount() error = %v`, err)
	}
	if account.AccountID != `account-1` || account.State != mail.EngineAccountProvisioning || account.AppliedGeneration != 9 || account.AppliedHash != `account-hash` {
		t.Fatalf(`EnsureAccount() = %#v, want confirmed provisioning account`, account)
	}
	fixture.requireMethods(t, `x:Domain/query`, `x:Domain/get`, `x:Account/query`, `x:Account/get`)
	fixture.requireAccountDomainID(t, `domain-1`)
}

func TestMailboxRegistryProvisioningEngineRejectsZeroGenerationBeforeIO(t *testing.T) {
	fixture := newRegistryProvisioningFixture(t, nil, nil)
	domain := registryDomainSpec()
	domain.DesiredGeneration = 0
	if _, err := fixture.engine.EnsureDomain(context.Background(), domain); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf(`zero-generation domain error = %v`, err)
	}
	if _, err := fixture.engine.GetDomain(context.Background(), ` Example.TEST `); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf(`noncanonical domain read error = %v`, err)
	}
	account := registryAccountSpec()
	account.DesiredGeneration = 0
	if _, err := fixture.engine.EnsureAccount(context.Background(), account); !errors.Is(err, mail.ErrEngineRejectedPolicy) {
		t.Fatalf(`zero-generation account error = %v`, err)
	}
	fixture.requireMethods(t)
}

func TestMailboxRegistryProvisioningEngineReturnsUnavailableBeforeIO(t *testing.T) {
	fixture := newRegistryProvisioningFixture(t, nil, nil)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fixture.engine.EnsureDomain(cancelled, registryDomainSpec()); !errors.Is(err, mail.ErrEngineUnavailable) || err.Error() != mail.ErrEngineUnavailable.Error() {
		t.Fatalf(`cancelled domain error = %v, want content-free unavailable`, err)
	}
	deadline, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer deadlineCancel()
	if _, err := fixture.engine.EnsureAccount(deadline, registryAccountSpec()); !errors.Is(err, mail.ErrEngineUnavailable) || err.Error() != mail.ErrEngineUnavailable.Error() {
		t.Fatalf(`expired account error = %v, want content-free unavailable`, err)
	}
	fixture.requireMethods(t)
}

func TestMailboxRegistryProvisioningEngineStopsBeforeAccountAfterDomainCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fixture := newRegistryProvisioningFixture(t, registryProvisioningHappyReplies()[:2], func(client *http.Client) {
		base := client.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		client.Transport = &registryProvisioningCancelAfterTransport{base: base, cancel: cancel}
	})
	if _, err := fixture.engine.EnsureAccount(ctx, registryAccountSpec()); !errors.Is(err, mail.ErrEngineUnavailable) || errors.Is(err, mail.ErrEngineIndeterminate) {
		t.Fatalf(`post-domain cancellation error = %v, want unavailable`, err)
	}
	fixture.requireMethods(t, `x:Domain/query`, `x:Domain/get`)
}

func TestMailboxRegistryProvisioningEngineKeepsStartedProviderFailureIndeterminateAndRedacted(t *testing.T) {
	replies := append(registryProvisioningHappyReplies()[:2], registryProviderFailureResponse())
	fixture := newRegistryProvisioningFixture(t, replies, nil)
	if _, err := fixture.engine.EnsureAccount(context.Background(), registryAccountSpec()); !errors.Is(err, mail.ErrEngineIndeterminate) || strings.Contains(err.Error(), `provider-secret`) {
		t.Fatalf(`provider failure error = %v, want redacted indeterminate`, err)
	}
	fixture.requireMethods(t, `x:Domain/query`, `x:Domain/get`, `x:Account/query`)
}

func TestMailboxRegistryProvisioningEngineGetDomainReturnsAuthoritativeRequiredDNS(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		responses   []string
		wantState   mail.EngineDomainStateKind
		wantRecords []mail.EngineDNSRecord
		wantCalls   []string
	}{
		{
			name:      `confirmed absence`,
			responses: []string{registryProvisioningDomainQueryReply([]string{})},
			wantState: mail.EngineDomainAbsent,
			wantCalls: []string{`x:Domain/query`},
		},
		{
			name:        `present projected zone`,
			responses:   registryProvisioningObservedDomainReplies(`domain-1`, `domain-1`, registryProvisioningMailCriticalZone()),
			wantState:   mail.EngineDomainDNSRequired,
			wantRecords: registryProvisioningExpectedDNSRecords(),
			wantCalls:   []string{`x:Domain/query`, `x:Domain/get`},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newRegistryProvisioningFixture(t, testCase.responses, nil)
			state, err := fixture.engine.GetDomain(context.Background(), `example.test`)
			if err != nil {
				t.Fatalf(`GetDomain() error = %v`, err)
			}
			if state.State != testCase.wantState || !slices.Equal(state.DNSRecords, testCase.wantRecords) {
				t.Fatalf(`GetDomain() = %#v, want state %q with records %#v`, state, testCase.wantState, testCase.wantRecords)
			}
			if testCase.wantState == mail.EngineDomainAbsent {
				if state.DomainID != `` || state.Applied || state.DNSRecords == nil || state.Verification != mail.EngineVerificationUnknown || state.DKIM != mail.EngineVerificationUnknown {
					t.Fatalf(`absent state = %#v, want exact confirmed absence`, state)
				}
			} else if state.DomainID != `domain-1` || !state.Applied || state.DNSRecords == nil || state.Verification != mail.EngineVerificationPending || state.DKIM != mail.EngineVerificationPending {
				t.Fatalf(`DNS-required state = %#v, want authoritative pending projection`, state)
			}
			fixture.requireMethods(t, testCase.wantCalls...)
		})
	}
}

func TestMailboxRegistryProvisioningEngineGetDomainReturnsIndependentDNSRecords(t *testing.T) {
	responses := append(registryProvisioningObservedDomainReplies(`domain-1`, `domain-1`, registryProvisioningMailCriticalZone()), registryProvisioningObservedDomainReplies(`domain-1`, `domain-1`, registryProvisioningMailCriticalZone())...)
	fixture := newRegistryProvisioningFixture(t, responses, nil)
	first, err := fixture.engine.GetDomain(context.Background(), `example.test`)
	if err != nil {
		t.Fatalf(`first GetDomain() error = %v`, err)
	}
	first.DNSRecords[0].Value = `mutated`
	second, err := fixture.engine.GetDomain(context.Background(), `example.test`)
	if err != nil {
		t.Fatalf(`second GetDomain() error = %v`, err)
	}
	if !slices.Equal(second.DNSRecords, registryProvisioningExpectedDNSRecords()) {
		t.Fatalf(`second GetDomain() records = %#v, want independent authoritative projection`, second.DNSRecords)
	}
	fixture.requireMethods(t, `x:Domain/query`, `x:Domain/get`, `x:Domain/query`, `x:Domain/get`)
}

func TestMailboxRegistryProvisioningEngineEnsureDomainReadsBackAuthoritativeDNS(t *testing.T) {
	fixture := newRegistryProvisioningFixture(t, registryProvisioningEnsureObservedDomainReplies(`domain-1`, `domain-1`, registryProvisioningMailCriticalZone()), nil)
	state, err := fixture.engine.EnsureDomain(context.Background(), registryDomainSpec())
	if err != nil {
		t.Fatalf(`EnsureDomain() error = %v`, err)
	}
	if state.DomainID != `domain-1` || state.State != mail.EngineDomainDNSRequired || !state.Applied || state.DNSRecords == nil || !slices.Equal(state.DNSRecords, registryProvisioningExpectedDNSRecords()) || state.Verification != mail.EngineVerificationPending || state.DKIM != mail.EngineVerificationPending {
		t.Fatalf(`EnsureDomain() = %#v, want authoritative DNS-required state`, state)
	}
	fixture.requireMethods(t, `x:Domain/query`, `x:Domain/get`, `x:Domain/query`, `x:Domain/get`)
}

func TestMailboxRegistryProvisioningEngineEnsureDomainDoesNotGuessFailedReadback(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		responses []string
		cancel    bool
		wantCalls []string
	}{
		{
			name:      `absent after ensure`,
			responses: append(registryProvisioningEnsureDomainReplies(`domain-1`), registryProvisioningDomainQueryReply([]string{})),
			wantCalls: []string{`x:Domain/query`, `x:Domain/get`, `x:Domain/query`},
		},
		{
			name:      `readback identifier drift`,
			responses: registryProvisioningEnsureObservedDomainReplies(`domain-1`, `domain-other`, registryProvisioningMailCriticalZone()),
			wantCalls: []string{`x:Domain/query`, `x:Domain/get`, `x:Domain/query`, `x:Domain/get`},
		},
		{
			name:      `incomplete asynchronous DKIM zone`,
			responses: registryProvisioningEnsureObservedDomainReplies(`domain-1`, `domain-1`, registryProvisioningIncompleteDKIMZone()),
			wantCalls: []string{`x:Domain/query`, `x:Domain/get`, `x:Domain/query`, `x:Domain/get`},
		},
		{
			name:      `post dispatch provider read failure remains redacted`,
			responses: append(registryProvisioningEnsureDomainReplies(`domain-1`), registryProvisioningProviderFailureResponse(`domain-query`)),
			wantCalls: []string{`x:Domain/query`, `x:Domain/get`, `x:Domain/query`},
		},
		{
			name:      `context loss after ensure`,
			responses: registryProvisioningEnsureDomainReplies(`domain-1`),
			cancel:    true,
			wantCalls: []string{`x:Domain/query`, `x:Domain/get`},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			configure := func(*http.Client) {}
			if testCase.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				defer cancel()
				configure = func(client *http.Client) {
					base := client.Transport
					if base == nil {
						base = http.DefaultTransport
					}
					client.Transport = &registryProvisioningCancelAfterTransport{base: base, cancel: cancel}
				}
			}
			fixture := newRegistryProvisioningFixture(t, testCase.responses, configure)
			state, err := fixture.engine.EnsureDomain(ctx, registryDomainSpec())
			if !errors.Is(err, mail.ErrEngineIndeterminate) || strings.Contains(err.Error(), `provider-secret`) {
				t.Fatalf(`EnsureDomain() error = %v, want redacted indeterminate`, err)
			}
			if state.State == mail.EngineDomainDNSRequired || state.State == mail.EngineDomainVerified || len(state.DNSRecords) != 0 || state.Verification == mail.EngineVerificationVerified || state.DKIM == mail.EngineVerificationVerified {
				t.Fatalf(`EnsureDomain() state = %#v, must not guess DNS-required or verified`, state)
			}
			fixture.requireMethods(t, testCase.wantCalls...)
		})
	}
}

func TestMailboxRegistryProvisioningEngineEnsureDomainRecoversProviderPanicsAsIndeterminate(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		responses []string
		panicAt   int
		wantCalls []string
	}{
		{
			name: `mutation dispatch`,
			responses: []string{
				registryProvisioningDomainQueryReply([]string{}),
				registryProvisioningDomainCreateReply(`domain-1`),
			},
			panicAt:   2,
			wantCalls: []string{`x:Domain/query`, `x:Domain/set`},
		},
		{
			name:      `authoritative readback`,
			responses: append(registryProvisioningEnsureDomainReplies(`domain-1`), registryProvisioningDomainQueryReply([]string{`domain-1`})),
			panicAt:   3,
			wantCalls: []string{`x:Domain/query`, `x:Domain/get`, `x:Domain/query`},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newRegistryProvisioningFixture(t, testCase.responses, func(client *http.Client) {
				base := client.Transport
				if base == nil {
					base = http.DefaultTransport
				}
				client.Transport = &registryProvisioningPanicAfterTransport{base: base, panicAt: testCase.panicAt}
			})

			state, err := fixture.engine.EnsureDomain(context.Background(), registryDomainSpec())
			if !errors.Is(err, mail.ErrEngineIndeterminate) || strings.Contains(err.Error(), `provider-secret`) {
				t.Fatalf(`EnsureDomain() error = %v, want redacted indeterminate`, err)
			}
			if state.DomainID != `` || state.State != `` || state.Applied || state.DNSRecords != nil || state.Verification != `` || state.DKIM != `` {
				t.Fatalf(`EnsureDomain() state = %#v, want exact zero state after provider panic`, state)
			}
			fixture.requireMethods(t, testCase.wantCalls...)
		})
	}
}

type registryProvisioningFixture struct {
	engine        *mail.MailboxRegistryProvisioningEngine
	responses     []string
	requests      []stalwart.JMAPRequest
	registryCalls int
}

func newRegistryProvisioningFixture(t *testing.T, responses []string, configure func(*http.Client)) *registryProvisioningFixture {
	t.Helper()
	fixture := &registryProvisioningFixture{responses: append([]string(nil), responses...)}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != `/jmap` || request.Header.Get(`Authorization`) != `Bearer registry-test` {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var payload stalwart.JMAPRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		fixture.requests = append(fixture.requests, payload)
		index := len(fixture.requests) - 1
		if index >= len(fixture.responses) {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set(`Content-Type`, `application/json`)
		_, _ = io.WriteString(writer, fixture.responses[index])
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	if configure != nil {
		configure(client)
	}
	authorizer := registryProvisioningAuthorizer{calls: &fixture.registryCalls}
	transport, err := stalwart.NewClient(stalwart.Config{Origin: server.URL, HTTPClient: client, RegistryAuthorizer: authorizer, MailboxAuthorizer: authorizer})
	if err != nil {
		t.Fatalf(`NewClient() error = %v`, err)
	}
	domains, err := stalwart.NewRegistryDomainController(transport)
	if err != nil {
		t.Fatalf(`NewRegistryDomainController() error = %v`, err)
	}
	accounts, err := stalwart.NewRegistryAccountClient(transport)
	if err != nil {
		t.Fatalf(`NewRegistryAccountClient() error = %v`, err)
	}
	fixture.engine, err = mail.NewMailboxRegistryProvisioningEngine(domains, accounts)
	if err != nil {
		t.Fatalf(`NewMailboxRegistryProvisioningEngine() error = %v`, err)
	}
	return fixture
}

func (fixture *registryProvisioningFixture) requireMethods(t *testing.T, want ...string) {
	t.Helper()
	got := make([]string, 0, len(fixture.requests))
	for _, request := range fixture.requests {
		if len(request.MethodCalls) != 1 {
			t.Fatalf(`JMAP method calls = %#v, want one call per request`, request.MethodCalls)
		}
		got = append(got, request.MethodCalls[0].Name)
	}
	if !slices.Equal(got, want) || fixture.registryCalls != len(want) {
		t.Fatalf(`Registry methods = %#v with %d authorizations, want %#v`, got, fixture.registryCalls, want)
	}
}

func (fixture *registryProvisioningFixture) requireAccountDomainID(t *testing.T, want string) {
	t.Helper()
	if len(fixture.requests) < 3 || len(fixture.requests[2].MethodCalls) != 1 {
		t.Fatalf(`account request is absent`)
	}
	var arguments map[string]any
	if err := json.Unmarshal(fixture.requests[2].MethodCalls[0].Arguments, &arguments); err != nil {
		t.Fatalf(`account query arguments error = %v`, err)
	}
	filter, validFilter := arguments[`filter`].(map[string]any)
	conditions, validConditions := filter[`conditions`].([]any)
	if !validFilter || !validConditions || len(conditions) != 2 {
		t.Fatalf(`account query arguments = %#v`, arguments)
	}
	domain, validDomain := conditions[1].(map[string]any)[`domainId`].(string)
	if !validDomain || domain != want {
		t.Fatalf(`account query domain = %q, want %q`, domain, want)
	}
}

type registryProvisioningAuthorizer struct {
	calls *int
}

func (authorizer registryProvisioningAuthorizer) AuthorizeRegistry(_ context.Context, request *http.Request) error {
	(*authorizer.calls)++
	request.Header.Set(`Authorization`, `Bearer registry-test`)
	return nil
}

func (registryProvisioningAuthorizer) AuthorizeMailbox(context.Context, *http.Request) error {
	return errors.New(`mailbox authorization must not be used`)
}

type registryProvisioningCancelAfterTransport struct {
	base   http.RoundTripper
	calls  int
	cancel context.CancelFunc
}

func (transport *registryProvisioningCancelAfterTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := transport.base.RoundTrip(request)
	if err != nil || response == nil || response.Body == nil {
		return response, err
	}
	transport.calls++
	if transport.calls == 2 {
		response.Body = &registryProvisioningCancelOnEOF{ReadCloser: response.Body, cancel: transport.cancel}
	}
	return response, nil
}

type registryProvisioningPanicAfterTransport struct {
	base    http.RoundTripper
	calls   int
	panicAt int
}

func (transport *registryProvisioningPanicAfterTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := transport.base.RoundTrip(request)
	transport.calls++
	if transport.calls != transport.panicAt {
		return response, err
	}
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
	panic(`provider-secret`)
}

type registryProvisioningCancelOnEOF struct {
	io.ReadCloser
	once   sync.Once
	cancel context.CancelFunc
}

func (reader *registryProvisioningCancelOnEOF) Read(buffer []byte) (int, error) {
	count, err := reader.ReadCloser.Read(buffer)
	if errors.Is(err, io.EOF) {
		reader.once.Do(reader.cancel)
	}
	return count, err
}

func registryProvisioningHappyReplies() []string {
	return []string{
		registryResponse(`x:Domain/query`, map[string]any{`accountId`: `0`, `queryState`: `domain-query-state`, `canCalculateChanges`: false, `position`: 0, `ids`: []string{`domain-1`}}, `domain-query`),
		registryResponse(`x:Domain/get`, map[string]any{`accountId`: `0`, `state`: `domain-get-state`, `list`: []any{map[string]any{`id`: `domain-1`, `name`: `example.test`, `isEnabled`: true, `allowRelaying`: false, `dkimManagement`: map[string]any{`@type`: `Automatic`}, `dnsManagement`: map[string]any{`@type`: `Manual`}}}, `notFound`: []any{}}, `domain-get`),
		registryResponse(`x:Account/query`, map[string]any{`accountId`: `0`, `queryState`: `account-query-state`, `ids`: []string{`account-1`}}, `registry-account-query`),
		registryResponse(`x:Account/get`, map[string]any{`accountId`: `0`, `list`: []any{map[string]any{`id`: `account-1`, `@type`: `User`, `name`: `ada`, `domainId`: `domain-1`, `emailAddress`: `ada@example.test`}}, `notFound`: []any{}}, `registry-account-get`),
	}
}

func registryProviderFailureResponse() string {
	return registryResponse(`error`, map[string]any{`type`: `serverFail`, `description`: `provider-secret`}, `registry-account-query`)
}

func registryResponse(method string, arguments any, callID string) string {
	payload, _ := json.Marshal(map[string]any{`methodResponses`: []any{[]any{method, arguments, callID}}})
	return string(payload)
}

func registryDomainSpec() mail.DomainSpec {
	return mail.DomainSpec{ProvisioningKey: `domain-key`, Domain: `example.test`, DesiredGeneration: 3, DesiredHash: `domain-hash`, Present: true, Enabled: true}
}

func registryAccountSpec() mail.AccountSpec {
	return mail.AccountSpec{
		ProvisioningKey: `account-key`, LocalPart: `ada`, Domain: `example.test`, DesiredGeneration: 9, DesiredHash: `account-hash`, Enabled: true,
		Primary: mail.AccountIdentitySpec{LocalPart: `ada`, Domain: `example.test`, Enabled: true}, Aliases: []mail.AccountIdentitySpec{},
	}
}

func registryProvisioningEnsureDomainReplies(id string) []string {
	return []string{
		registryProvisioningDomainQueryReply([]string{id}),
		registryProvisioningDomainGetReply(id),
	}
}

func registryProvisioningObservedDomainReplies(queryID, responseID, zone string) []string {
	return []string{
		registryProvisioningDomainQueryReply([]string{queryID}),
		registryProvisioningDomainZoneReply(responseID, zone),
	}
}

func registryProvisioningEnsureObservedDomainReplies(ensureID, observedID, zone string) []string {
	return append(registryProvisioningEnsureDomainReplies(ensureID), registryProvisioningObservedDomainReplies(ensureID, observedID, zone)...)
}

func registryProvisioningDomainQueryReply(ids []string) string {
	return registryResponse(`x:Domain/query`, map[string]any{`accountId`: `0`, `queryState`: `domain-query-state`, `canCalculateChanges`: false, `position`: 0, `ids`: ids}, `domain-query`)
}

func registryProvisioningDomainGetReply(id string) string {
	return registryResponse(`x:Domain/get`, map[string]any{`accountId`: `0`, `state`: `domain-get-state`, `list`: []any{map[string]any{`id`: id, `name`: `example.test`, `isEnabled`: true, `allowRelaying`: false, `dkimManagement`: map[string]any{`@type`: `Automatic`}, `dnsManagement`: map[string]any{`@type`: `Manual`}}}, `notFound`: []any{}}, `domain-get`)
}

func registryProvisioningDomainCreateReply(id string) string {
	return registryResponse(`x:Domain/set`, map[string]any{
		`accountId`: `0`,
		`oldState`:  `domain-set-before`,
		`newState`:  `domain-set-after`,
		`created`:   map[string]any{`create-local-domain`: map[string]any{`id`: id}},
	}, `domain-set`)
}

func registryProvisioningDomainZoneReply(id, zone string) string {
	return registryResponse(`x:Domain/get`, map[string]any{`accountId`: `0`, `state`: `domain-zone-get-state`, `list`: []any{map[string]any{`id`: id, `name`: `example.test`, `isEnabled`: true, `allowRelaying`: false, `dkimManagement`: map[string]any{`@type`: `Automatic`}, `dnsManagement`: map[string]any{`@type`: `Manual`}, `dnsZoneFile`: zone}}, `notFound`: []any{}}, `domain-zone-get`)
}

func registryProvisioningMailCriticalZone() string {
	return "autoconfig.example.test. IN CNAME mail.example.test.\n" +
		"example.test. 600 IN MX 10 mail.example.test.\n" +
		"example.test. 3600 IN TXT \"v=spf1 -all\"\n" +
		"_dmarc.example.test. IN TXT \"v=DMARC1; p=none\"\n" +
		"selector._domainkey.example.test. 60 IN TXT \"v=DKIM1; p=key\"\n"
}

func registryProvisioningIncompleteDKIMZone() string {
	return "example.test. IN MX 10 mail.example.test.\n" +
		"example.test. IN TXT \"v=spf1 -all\"\n" +
		"_dmarc.example.test. IN TXT \"v=DMARC1; p=none\"\n"
}

func registryProvisioningExpectedDNSRecords() []mail.EngineDNSRecord {
	return []mail.EngineDNSRecord{
		{Type: mail.EngineDNSRecordCNAME, Name: `autoconfig.example.test`, Value: `mail.example.test`, TTL: 300},
		{Type: mail.EngineDNSRecordMX, Name: `example.test`, Value: `10 mail.example.test`, TTL: 600},
		{Type: mail.EngineDNSRecordTXT, Name: `_dmarc.example.test`, Value: `v=DMARC1; p=none`, TTL: 300},
		{Type: mail.EngineDNSRecordTXT, Name: `example.test`, Value: `v=spf1 -all`, TTL: 3600},
		{Type: mail.EngineDNSRecordTXT, Name: `selector._domainkey.example.test`, Value: `v=DKIM1; p=key`, TTL: 60},
	}
}

func registryProvisioningProviderFailureResponse(callID string) string {
	return registryResponse(`error`, map[string]any{`type`: `serverFail`, `description`: `provider-secret`}, callID)
}
