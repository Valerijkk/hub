package mail_test

import (
	`context`
	`encoding/json`
	`errors`
	`io`
	`net/http`
	`net/http/httptest`
	`slices`
	`strings`
	`sync`
	`testing`
	`time`

	`github.com/valerii/hub/backend/internal/mail`
	`github.com/valerii/hub/backend/internal/mail/stalwart`
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
		registryResponse(`x:Domain/get`, map[string]any{`accountId`: `0`, `state`: `domain-get-state`, `list`: []any{map[string]any{`id`: `domain-1`, `name`: `example.test`, `isEnabled`: true, `allowRelaying`: false}}, `notFound`: []any{}}, `domain-get`),
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
