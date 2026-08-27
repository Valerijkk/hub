package mail_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/mail/stalwart"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

const bootstrapOAuthIssuer = "https://oauth.example.test"

func TestMailboxBootstrapOAuthService_ConstructorRejectsInvalidDependenciesAndIssuer(t *testing.T) {
	scope := bootstrapOAuthScope()
	account := bootstrapOAuthAccount(scope)
	repository := &bootstrapOAuthRepositoryFake{accounts: []mail.MailboxBootstrapOAuthAccount{account}}
	states := &bootstrapOAuthStateFake{pending: map[string]mail.MailboxBootstrapOAuthPending{}}
	protocol := &bootstrapOAuthProtocolFake{authorization: bootstrapOAuthAuthorization()}
	proof := &bootstrapOAuthIdentityProofFake{result: mail.MailboxBootstrapIdentityProofResult{EngineIdentityID: "engine-identity-1"}}
	clock := bootstrapOAuthClock{now: bootstrapOAuthNow()}
	cipher := bootstrapOAuthCipher(t)

	for _, testCase := range []struct {
		name   string
		build  func() (*mail.MailboxBootstrapOAuthService, error)
	}{
		{name: "nil repository", build: func() (*mail.MailboxBootstrapOAuthService, error) {
			return mail.NewMailboxBootstrapOAuthService(nil, states, protocol, proof, cipher, clock, bootstrapOAuthIssuer)
		}},
		{name: "nil states", build: func() (*mail.MailboxBootstrapOAuthService, error) {
			return mail.NewMailboxBootstrapOAuthService(repository, nil, protocol, proof, cipher, clock, bootstrapOAuthIssuer)
		}},
		{name: "nil protocol", build: func() (*mail.MailboxBootstrapOAuthService, error) {
			return mail.NewMailboxBootstrapOAuthService(repository, states, nil, proof, cipher, clock, bootstrapOAuthIssuer)
		}},
		{name: "nil proof", build: func() (*mail.MailboxBootstrapOAuthService, error) {
			return mail.NewMailboxBootstrapOAuthService(repository, states, protocol, nil, cipher, clock, bootstrapOAuthIssuer)
		}},
		{name: "nil cipher", build: func() (*mail.MailboxBootstrapOAuthService, error) {
			return mail.NewMailboxBootstrapOAuthService(repository, states, protocol, proof, nil, clock, bootstrapOAuthIssuer)
		}},
		{name: "nil clock", build: func() (*mail.MailboxBootstrapOAuthService, error) {
			return mail.NewMailboxBootstrapOAuthService(repository, states, protocol, proof, cipher, nil, bootstrapOAuthIssuer)
		}},
		{name: "noncanonical issuer", build: func() (*mail.MailboxBootstrapOAuthService, error) {
			return mail.NewMailboxBootstrapOAuthService(repository, states, protocol, proof, cipher, clock, "https://oauth.example.test/")
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			service, err := testCase.build()
			if service != nil || !errors.Is(err, mail.ErrMailboxOAuthInvalid) {
				t.Fatalf("NewMailboxBootstrapOAuthService() = (%v, %v), want nil invalid", service, err)
			}
		})
	}

	var typedNil *bootstrapOAuthRepositoryFake
	if service, err := mail.NewMailboxBootstrapOAuthService(typedNil, states, protocol, proof, cipher, clock, bootstrapOAuthIssuer); service != nil || !errors.Is(err, mail.ErrMailboxOAuthInvalid) {
		t.Fatalf("NewMailboxBootstrapOAuthService(typed nil repository) = (%v, %v), want nil invalid", service, err)
	}
}

func TestMailboxBootstrapOAuthService_BeginSavesExactProvisioningSnapshot(t *testing.T) {
	scope := bootstrapOAuthScope()
	account := bootstrapOAuthAccount(scope)
	fixture := newBootstrapOAuthFixture(t, account)

	result, err := fixture.service.BeginMailboxBootstrapOAuth(context.Background(), scope)
	if err != nil {
		t.Fatalf("BeginMailboxBootstrapOAuth() error = %v", err)
	}
	if result.AuthorizationURL != fixture.protocol.authorization.URL || !result.ExpiresAt.Equal(fixture.clock.now.Add(10*time.Minute)) {
		t.Fatalf("BeginMailboxBootstrapOAuth() result = %#v, want URL and exact expiry", result)
	}
	binding, err := base64.RawURLEncoding.DecodeString(result.BrowserCookie)
	if err != nil || len(binding) != 32 {
		t.Fatalf("BeginMailboxBootstrapOAuth() browser cookie = %q, decode=%v len=%d; want 32 bytes", result.BrowserCookie, err, len(binding))
	}
	defer clearBootstrapOAuthBytes(binding)

	pending, exists := fixture.states.pending[fixture.protocol.authorization.State]
	if !exists {
		t.Fatal("BeginMailboxBootstrapOAuth() did not save bootstrap callback state")
	}
	if pending.Scope != scope || pending.EngineAccountID != account.EngineAccountID || pending.PasswordEnrollmentGeneration != account.PasswordEnrollmentGeneration || pending.CodeVerifier != fixture.protocol.authorization.CodeVerifier || pending.ExpiresAt != result.ExpiresAt || pending.BrowserBindingHash != sha256.Sum256(binding) {
		t.Fatalf("saved pending state = %#v, want exact provisioning snapshot", pending)
	}
	if fixture.repository.resolveCalls != 1 || fixture.protocol.beginCalls != 1 || fixture.states.saveCalls != 1 {
		t.Fatalf("resolve/begin/save = %d/%d/%d, want 1/1/1", fixture.repository.resolveCalls, fixture.protocol.beginCalls, fixture.states.saveCalls)
	}
}

func TestMailboxBootstrapOAuthService_ExpiryConsumesStateBeforeProviderIO(t *testing.T) {
	scope := bootstrapOAuthScope()
	fixture := newBootstrapOAuthFixture(t, bootstrapOAuthAccount(scope))
	begin := beginBootstrapOAuth(t, fixture, scope)
	fixture.clock.now = begin.ExpiresAt

	result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, `authorization-code-1`, begin.BrowserCookie)
	if result != (mail.MailboxOAuthConnectionResult{}) || !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`CompleteMailboxBootstrapOAuth(expired) = (%#v, %v), want state invalid`, result, err)
	}
	if fixture.states.consumeCalls != 1 || fixture.protocol.exchangeCalls != 0 || fixture.identity.calls != 0 || fixture.repository.finalizeCalls != 0 {
		t.Fatalf(`consume/exchange/proof/finalize = %d/%d/%d/%d, want 1/0/0/0`, fixture.states.consumeCalls, fixture.protocol.exchangeCalls, fixture.identity.calls, fixture.repository.finalizeCalls)
	}
}

func TestMailboxBootstrapOAuthService_RejectsMalformedProvisioningSnapshotBeforeProviderIO(t *testing.T) {
	scope := bootstrapOAuthScope()
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxBootstrapOAuthAccount)
	}{
		{name: `noncanonical primary email`, mutate: func(account *mail.MailboxBootstrapOAuthAccount) { account.PrimaryEmail = `OWNER@example.test` }},
		{name: `nonopaque engine account`, mutate: func(account *mail.MailboxBootstrapOAuthAccount) { account.EngineAccountID = `engine account` }},
		{name: `zero enrollment generation`, mutate: func(account *mail.MailboxBootstrapOAuthAccount) { account.PasswordEnrollmentGeneration = 0 }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			account := bootstrapOAuthAccount(scope)
			testCase.mutate(&account)
			fixture := newBootstrapOAuthFixture(t, account)

			result, err := fixture.service.BeginMailboxBootstrapOAuth(context.Background(), scope)
			if result != (mail.MailboxOAuthBeginResult{}) || !errors.Is(err, mail.ErrMailboxOAuthNotFound) {
				t.Fatalf(`BeginMailboxBootstrapOAuth() = (%#v, %v), want not found`, result, err)
			}
			if fixture.protocol.beginCalls != 0 || fixture.states.saveCalls != 0 {
				t.Fatalf(`begin/save = %d/%d, want 0/0`, fixture.protocol.beginCalls, fixture.states.saveCalls)
			}
		})
	}
}

func TestMailboxBootstrapOAuthService_CallbackBindingAndReplay(t *testing.T) {
	scope := bootstrapOAuthScope()
	fixture := newBootstrapOAuthFixture(t, bootstrapOAuthAccount(scope))
	begin := beginBootstrapOAuth(t, fixture, scope)

	wrongCookie := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x02}, 32))
	if result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", wrongCookie); result != (mail.MailboxOAuthConnectionResult{}) || !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf("CompleteMailboxBootstrapOAuth(wrong binding) = (%#v, %v), want state invalid", result, err)
	}
	if fixture.protocol.exchangeCalls != 0 || fixture.states.consumeCalls != 1 {
		t.Fatalf("wrong binding exchange/consume = %d/%d, want 0/1", fixture.protocol.exchangeCalls, fixture.states.consumeCalls)
	}

	result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie)
	if err != nil || result.AccountID != scope.AccountID || result.CredentialGeneration != 1 {
		t.Fatalf("CompleteMailboxBootstrapOAuth() = (%#v, %v), want account with generation 1", result, err)
	}
	if fixture.repository.finalizeCalls != 1 {
		t.Fatalf("finalizer calls = %d, want 1", fixture.repository.finalizeCalls)
	}

	if result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie); result != (mail.MailboxOAuthConnectionResult{}) || !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf("CompleteMailboxBootstrapOAuth(replay) = (%#v, %v), want state invalid", result, err)
	}
	if fixture.protocol.exchangeCalls != 1 || fixture.repository.finalizeCalls != 1 {
		t.Fatalf("replay exchange/finalize = %d/%d, want 1/1", fixture.protocol.exchangeCalls, fixture.repository.finalizeCalls)
	}
}

func TestMailboxBootstrapOAuthService_ConsumesBeforeScopeEngineAndGenerationDrift(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		callback    func(mail.MailboxOAuthScope) mail.MailboxOAuthScope
		second      func(mail.MailboxBootstrapOAuthAccount) mail.MailboxBootstrapOAuthAccount
	}{
		{name: "scope", callback: func(scope mail.MailboxOAuthScope) mail.MailboxOAuthScope {
			scope.AccountID = uuid.New()
			return scope
		}, second: func(account mail.MailboxBootstrapOAuthAccount) mail.MailboxBootstrapOAuthAccount { return account }},
		{name: "engine account", callback: func(scope mail.MailboxOAuthScope) mail.MailboxOAuthScope { return scope }, second: func(account mail.MailboxBootstrapOAuthAccount) mail.MailboxBootstrapOAuthAccount {
			account.EngineAccountID = "engine-account-2"
			return account
		}},
		{name: "enrollment generation", callback: func(scope mail.MailboxOAuthScope) mail.MailboxOAuthScope { return scope }, second: func(account mail.MailboxBootstrapOAuthAccount) mail.MailboxBootstrapOAuthAccount {
			account.PasswordEnrollmentGeneration++
			return account
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			scope := bootstrapOAuthScope()
			first := bootstrapOAuthAccount(scope)
			fixture := newBootstrapOAuthFixture(t, first, testCase.second(first))
			begin := beginBootstrapOAuth(t, fixture, scope)

			result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), testCase.callback(scope), fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie)
			if result != (mail.MailboxOAuthConnectionResult{}) || !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
				t.Fatalf("CompleteMailboxBootstrapOAuth() = (%#v, %v), want consumed state invalid", result, err)
			}
			if fixture.states.consumeCalls != 1 || fixture.protocol.exchangeCalls != 0 || fixture.identity.calls != 0 || fixture.repository.finalizeCalls != 0 {
				t.Fatalf("consume/exchange/proof/finalize = %d/%d/%d/%d, want 1/0/0/0", fixture.states.consumeCalls, fixture.protocol.exchangeCalls, fixture.identity.calls, fixture.repository.finalizeCalls)
			}
		})
	}
}

func TestMailboxBootstrapOAuthService_ProtocolInvalidAndFailureConsumeState(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		tokens stalwart.OAuthTokenSet
		err    error
		want   error
	}{
		{name: "invalid token shape", tokens: stalwart.OAuthTokenSet{AccessToken: "access-token-1", RefreshToken: "refresh-token-1", TokenType: "Bearer", Scope: "mail", ExpiresInSeconds: 120}, want: mail.ErrMailboxOAuthUnavailable},
		{name: "provider failure", err: errors.New("provider body must not escape"), want: mail.ErrMailboxOAuthUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			scope := bootstrapOAuthScope()
			fixture := newBootstrapOAuthFixture(t, bootstrapOAuthAccount(scope))
			fixture.protocol.tokens = testCase.tokens
			fixture.protocol.exchangeErr = testCase.err
			begin := beginBootstrapOAuth(t, fixture, scope)

			result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie)
			if result != (mail.MailboxOAuthConnectionResult{}) || !errors.Is(err, testCase.want) || (testCase.err != nil && strings.Contains(err.Error(), testCase.err.Error())) {
				t.Fatalf("CompleteMailboxBootstrapOAuth() = (%#v, %v), want redacted %v", result, err, testCase.want)
			}
			if fixture.states.consumeCalls != 1 || fixture.protocol.exchangeCalls != 1 || fixture.identity.calls != 0 || fixture.repository.finalizeCalls != 0 {
				t.Fatalf("consume/exchange/proof/finalize = %d/%d/%d/%d, want 1/1/0/0", fixture.states.consumeCalls, fixture.protocol.exchangeCalls, fixture.identity.calls, fixture.repository.finalizeCalls)
			}
			if _, replayErr := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie); !errors.Is(replayErr, mail.ErrMailboxOAuthStateInvalid) {
				t.Fatalf("callback retry after protocol failure error = %v, want consumed state invalid", replayErr)
			}
		})
	}
}

func TestMailboxBootstrapOAuthService_IdentityMismatchAndFailureConsumeState(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		result mail.MailboxBootstrapIdentityProofResult
		err    error
	}{
		{name: "mismatch", result: mail.MailboxBootstrapIdentityProofResult{}},
		{name: "failure", err: errors.New("provider identity body must not escape")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			scope := bootstrapOAuthScope()
			fixture := newBootstrapOAuthFixture(t, bootstrapOAuthAccount(scope))
			fixture.identity.result = testCase.result
			fixture.identity.err = testCase.err
			begin := beginBootstrapOAuth(t, fixture, scope)

			result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie)
			if result != (mail.MailboxOAuthConnectionResult{}) || !errors.Is(err, mail.ErrMailboxOAuthUnavailable) || (testCase.err != nil && strings.Contains(err.Error(), testCase.err.Error())) {
				t.Fatalf("CompleteMailboxBootstrapOAuth() = (%#v, %v), want redacted unavailable", result, err)
			}
			if fixture.protocol.exchangeCalls != 1 || fixture.identity.calls != 1 || fixture.repository.finalizeCalls != 0 {
				t.Fatalf("exchange/proof/finalize = %d/%d/%d, want 1/1/0", fixture.protocol.exchangeCalls, fixture.identity.calls, fixture.repository.finalizeCalls)
			}
			if _, replayErr := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie); !errors.Is(replayErr, mail.ErrMailboxOAuthStateInvalid) {
				t.Fatalf("callback retry after identity failure error = %v, want consumed state invalid", replayErr)
			}
		})
	}
}

func TestMailboxBootstrapOAuthService_FinalizesExactEncryptedGenerationOneRequest(t *testing.T) {
	scope := bootstrapOAuthScope()
	account := bootstrapOAuthAccount(scope)
	fixture := newBootstrapOAuthFixture(t, account)
	begin := beginBootstrapOAuth(t, fixture, scope)

	result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie)
	if err != nil {
		t.Fatalf("CompleteMailboxBootstrapOAuth() error = %v", err)
	}
	wantAccessExpiry := fixture.clock.now.Add(120 * time.Second)
	wantRefreshExpiry := fixture.clock.now.Add(7*24*time.Hour - time.Minute)
	if result.AccountID != scope.AccountID || result.CredentialGeneration != 1 || !result.AccessTokenExpiresAt.Equal(wantAccessExpiry) {
		t.Fatalf("CompleteMailboxBootstrapOAuth() = %#v, want account/generation=1/access expiry", result)
	}
	if fixture.repository.finalizeCalls != 1 || fixture.repository.request.Account != account || fixture.repository.request.IssuerOrigin != bootstrapOAuthIssuer || fixture.repository.request.EngineIdentityID != "engine-identity-1" || !fixture.repository.request.AccessTokenExpiresAt.Equal(wantAccessExpiry) || !fixture.repository.request.RefreshTokenExpiresAt.Equal(wantRefreshExpiry) || !fixture.repository.request.ObservedAt.Equal(fixture.clock.now) {
		t.Fatalf("finalize request = %#v, want exact account snapshot and timestamps", fixture.repository.request)
	}

	if fixture.identity.calls != 1 || fixture.identity.request.EngineAccountID != account.EngineAccountID || fixture.identity.request.ExpectedPrimaryEmail != account.PrimaryEmail || fixture.identity.request.AccessToken != fixture.protocol.tokens.AccessToken {
		t.Fatalf(`identity request = %#v, want exactly one ephemeral access-bearer proof for the re-resolved account/email`, fixture.identity.request)
	}
	oauthAccount := mail.MailboxOAuthAccount{Scope: scope, EngineAccountID: account.EngineAccountID, CredentialGeneration: 1}
	access, accessErr := fixture.cipher.DecryptStringWithAAD(string(fixture.repository.accessAtCall), bootstrapOAuthAAD(oauthAccount, bootstrapOAuthIssuer, "access"))
	refresh, refreshErr := fixture.cipher.DecryptStringWithAAD(string(fixture.repository.refreshAtCall), bootstrapOAuthAAD(oauthAccount, bootstrapOAuthIssuer, "refresh"))
	if accessErr != nil || refreshErr != nil || access != fixture.protocol.tokens.AccessToken || refresh != fixture.protocol.tokens.RefreshToken {
		t.Fatalf("finalize ciphertext decrypt = (%q, %v) / (%q, %v), want exact tokens", access, accessErr, refresh, refreshErr)
	}
	if !bootstrapOAuthAllZero(fixture.repository.request.AccessTokenCiphertext) || !bootstrapOAuthAllZero(fixture.repository.request.RefreshTokenCiphertext) {
		t.Fatal("service did not clear its owned ciphertext buffers after finalizer returned")
	}
	wantEvents := []string{"resolve", "begin", "save", "consume", "resolve", "exchange", "proof", "finalize"}
	if !reflect.DeepEqual(fixture.events, wantEvents) {
		t.Fatalf("I/O ordering = %#v, want %#v", fixture.events, wantEvents)
	}
}

func TestMailboxBootstrapOAuthService_FinalizerFailureIsRedactedAndStateStaysConsumed(t *testing.T) {
	scope := bootstrapOAuthScope()
	fixture := newBootstrapOAuthFixture(t, bootstrapOAuthAccount(scope))
	fixture.repository.finalizeErr = errors.New("durable provider body must not escape")
	begin := beginBootstrapOAuth(t, fixture, scope)

	result, err := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie)
	if result != (mail.MailboxOAuthConnectionResult{}) || !errors.Is(err, mail.ErrMailboxOAuthUnavailable) || strings.Contains(err.Error(), "durable provider body") {
		t.Fatalf("CompleteMailboxBootstrapOAuth() = (%#v, %v), want redacted unavailable", result, err)
	}
	if fixture.repository.finalizeCalls != 1 || !bootstrapOAuthAllZero(fixture.repository.request.AccessTokenCiphertext) || !bootstrapOAuthAllZero(fixture.repository.request.RefreshTokenCiphertext) {
		t.Fatalf("finalizer/clear = %d/%t/%t, want 1/true/true", fixture.repository.finalizeCalls, bootstrapOAuthAllZero(fixture.repository.request.AccessTokenCiphertext), bootstrapOAuthAllZero(fixture.repository.request.RefreshTokenCiphertext))
	}
	if _, replayErr := fixture.service.CompleteMailboxBootstrapOAuth(context.Background(), scope, fixture.protocol.authorization.State, "authorization-code-1", begin.BrowserCookie); !errors.Is(replayErr, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf("callback retry after finalizer failure error = %v, want consumed state invalid", replayErr)
	}
}

func beginBootstrapOAuth(t *testing.T, fixture *bootstrapOAuthFixture, scope mail.MailboxOAuthScope) mail.MailboxOAuthBeginResult {
	t.Helper()
	result, err := fixture.service.BeginMailboxBootstrapOAuth(context.Background(), scope)
	if err != nil {
		t.Fatalf("BeginMailboxBootstrapOAuth() error = %v", err)
	}
	return result
}

type bootstrapOAuthFixture struct {
	service    *mail.MailboxBootstrapOAuthService
	repository *bootstrapOAuthRepositoryFake
	states     *bootstrapOAuthStateFake
	protocol   *bootstrapOAuthProtocolFake
	identity   *bootstrapOAuthIdentityProofFake
	clock      *bootstrapOAuthClock
	cipher     *cryptopkg.Cipher
	events     []string
}

func newBootstrapOAuthFixture(t *testing.T, accounts ...mail.MailboxBootstrapOAuthAccount) *bootstrapOAuthFixture {
	t.Helper()
	fixture := &bootstrapOAuthFixture{clock: &bootstrapOAuthClock{now: bootstrapOAuthNow()}}
	fixture.repository = &bootstrapOAuthRepositoryFake{accounts: accounts, events: &fixture.events}
	fixture.states = &bootstrapOAuthStateFake{pending: map[string]mail.MailboxBootstrapOAuthPending{}, events: &fixture.events}
	fixture.protocol = &bootstrapOAuthProtocolFake{authorization: bootstrapOAuthAuthorization(), tokens: bootstrapOAuthTokens(), events: &fixture.events}
	fixture.identity = &bootstrapOAuthIdentityProofFake{result: mail.MailboxBootstrapIdentityProofResult{EngineIdentityID: "engine-identity-1"}, events: &fixture.events}
	fixture.cipher = bootstrapOAuthCipher(t)
	service, err := mail.NewMailboxBootstrapOAuthService(fixture.repository, fixture.states, fixture.protocol, fixture.identity, fixture.cipher, fixture.clock, bootstrapOAuthIssuer)
	if err != nil {
		t.Fatalf("NewMailboxBootstrapOAuthService() error = %v", err)
	}
	fixture.service = service
	return fixture
}

func bootstrapOAuthScope() mail.MailboxOAuthScope {
	return mail.MailboxOAuthScope{ActorID: uuid.New(), OrganizationID: uuid.New(), AccountID: uuid.New()}
}

func bootstrapOAuthAccount(scope mail.MailboxOAuthScope) mail.MailboxBootstrapOAuthAccount {
	return mail.MailboxBootstrapOAuthAccount{Scope: scope, EngineAccountID: "engine-account-1", PrimaryEmail: "owner@example.test", PasswordEnrollmentGeneration: 7}
}

func bootstrapOAuthAuthorization() stalwart.OAuthAuthorization {
	return stalwart.OAuthAuthorization{
		URL:          "https://oauth.example.test/authorize?state=bootstrap",
		State:        base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x12}, 32)),
		CodeVerifier: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x34}, 64)),
	}
}

func bootstrapOAuthTokens() stalwart.OAuthTokenSet {
	return stalwart.OAuthTokenSet{AccessToken: "access-token-1", RefreshToken: "refresh-token-1", TokenType: "Bearer", Scope: "offline_access mail", ExpiresInSeconds: 120}
}

func bootstrapOAuthNow() time.Time {
	return time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC)
}

func bootstrapOAuthCipher(t *testing.T) *cryptopkg.Cipher {
	t.Helper()
	cipher, err := cryptopkg.NewCipherFromKey(bytes.Repeat([]byte{0x6c}, 32))
	if err != nil {
		t.Fatalf("NewCipherFromKey() error = %v", err)
	}
	return cipher
}

func bootstrapOAuthAAD(account mail.MailboxOAuthAccount, issuerOrigin, kind string) []byte {
	return []byte("hub.mail.oauth/v1\x00" + account.Scope.OrganizationID.String() + "\x00" + account.Scope.AccountID.String() + "\x00stalwart\x00" + issuerOrigin + "\x00" + "1\x00" + kind)
}

type bootstrapOAuthRepositoryFake struct {
	accounts      []mail.MailboxBootstrapOAuthAccount
	resolveErr    error
	finalizeErr   error
	resolveCalls  int
	finalizeCalls int
	request       mail.MailboxBootstrapOAuthFinalizeRequest
	accessAtCall  []byte
	refreshAtCall []byte
	events        *[]string
}

func (fake *bootstrapOAuthRepositoryFake) ResolveMailboxBootstrapOAuthAccount(_ context.Context, scope mail.MailboxOAuthScope) (mail.MailboxBootstrapOAuthAccount, error) {
	fake.resolveCalls++
	bootstrapOAuthEvent(fake.events, "resolve")
	if fake.resolveErr != nil || len(fake.accounts) == 0 {
		return mail.MailboxBootstrapOAuthAccount{}, fake.resolveErr
	}
	index := fake.resolveCalls - 1
	if index >= len(fake.accounts) {
		index = len(fake.accounts) - 1
	}
	account := fake.accounts[index]
	if account.Scope != scope {
		return account, nil
	}
	return account, nil
}

func (fake *bootstrapOAuthRepositoryFake) FinalizeMailboxBootstrapOAuth(_ context.Context, request mail.MailboxBootstrapOAuthFinalizeRequest) error {
	fake.finalizeCalls++
	bootstrapOAuthEvent(fake.events, "finalize")
	fake.request = request
	fake.accessAtCall = append([]byte(nil), request.AccessTokenCiphertext...)
	fake.refreshAtCall = append([]byte(nil), request.RefreshTokenCiphertext...)
	return fake.finalizeErr
}

type bootstrapOAuthStateFake struct {
	pending      map[string]mail.MailboxBootstrapOAuthPending
	saveErr      error
	consumeErr   error
	saveCalls    int
	consumeCalls int
	events       *[]string
}

func (fake *bootstrapOAuthStateFake) SaveMailboxBootstrapOAuthState(_ context.Context, state string, pending mail.MailboxBootstrapOAuthPending) error {
	fake.saveCalls++
	bootstrapOAuthEvent(fake.events, "save")
	if fake.saveErr != nil {
		return fake.saveErr
	}
	fake.pending[state] = pending
	return nil
}

func (fake *bootstrapOAuthStateFake) ConsumeMailboxBootstrapOAuthState(_ context.Context, state string, binding [sha256.Size]byte) (mail.MailboxBootstrapOAuthPending, error) {
	fake.consumeCalls++
	bootstrapOAuthEvent(fake.events, "consume")
	if fake.consumeErr != nil {
		return mail.MailboxBootstrapOAuthPending{}, fake.consumeErr
	}
	pending, exists := fake.pending[state]
	if !exists || pending.BrowserBindingHash != binding {
		return mail.MailboxBootstrapOAuthPending{}, mail.ErrMailboxOAuthStateInvalid
	}
	delete(fake.pending, state)
	return pending, nil
}

type bootstrapOAuthProtocolFake struct {
	authorization stalwart.OAuthAuthorization
	tokens        stalwart.OAuthTokenSet
	beginErr      error
	exchangeErr   error
	beginCalls    int
	exchangeCalls int
	events        *[]string
}

func (fake *bootstrapOAuthProtocolFake) BeginAuthorization(context.Context) (stalwart.OAuthAuthorization, error) {
	fake.beginCalls++
	bootstrapOAuthEvent(fake.events, "begin")
	return fake.authorization, fake.beginErr
}

func (fake *bootstrapOAuthProtocolFake) ExchangeAuthorizationCode(_ context.Context, code, verifier string) (stalwart.OAuthTokenSet, error) {
	fake.exchangeCalls++
	bootstrapOAuthEvent(fake.events, "exchange")
	if code != "authorization-code-1" || verifier != fake.authorization.CodeVerifier {
		return stalwart.OAuthTokenSet{}, errors.New("unexpected callback credentials")
	}
	return fake.tokens, fake.exchangeErr
}

func (fake *bootstrapOAuthProtocolFake) RefreshAccessToken(context.Context, string) (stalwart.OAuthTokenSet, error) {
	return stalwart.OAuthTokenSet{}, errors.New("unexpected refresh")
}

type bootstrapOAuthIdentityProofFake struct {
	result   mail.MailboxBootstrapIdentityProofResult
	err      error
	calls    int
	request  mail.MailboxBootstrapIdentityProofRequest
	events   *[]string
}

func (fake *bootstrapOAuthIdentityProofFake) VerifyMailboxBootstrapIdentity(_ context.Context, request mail.MailboxBootstrapIdentityProofRequest) (mail.MailboxBootstrapIdentityProofResult, error) {
	fake.calls++
	bootstrapOAuthEvent(fake.events, "proof")
	fake.request = request
	return fake.result, fake.err
}

type bootstrapOAuthClock struct{ now time.Time }

func (clock bootstrapOAuthClock) Now() time.Time { return clock.now }

func bootstrapOAuthEvent(events *[]string, event string) {
	if events != nil {
		*events = append(*events, event)
	}
}

func clearBootstrapOAuthBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}

func bootstrapOAuthAllZero(value []byte) bool {
	for _, item := range value {
		if item != 0 {
			return false
		}
	}
	return true
}

var _ mail.MailboxBootstrapOAuthRepository = (*bootstrapOAuthRepositoryFake)(nil)
var _ mail.MailboxBootstrapOAuthStateStore = (*bootstrapOAuthStateFake)(nil)
var _ mail.MailboxOAuthProtocol = (*bootstrapOAuthProtocolFake)(nil)
var _ mail.MailboxBootstrapIdentityProof = (*bootstrapOAuthIdentityProofFake)(nil)
var _ mail.MailboxRepositoryClock = bootstrapOAuthClock{}
