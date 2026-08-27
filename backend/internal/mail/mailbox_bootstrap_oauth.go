package mail

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"io"
	"time"

	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// MailboxBootstrapOAuthAccount is the complete pending password-enrollment
// snapshot required to begin and finalize one bootstrap-only OAuth flow.
type MailboxBootstrapOAuthAccount struct {
	Scope                        MailboxOAuthScope
	EngineAccountID              string
	PrimaryEmail                 string
	PasswordEnrollmentGeneration int64
}

// MailboxBootstrapOAuthFinalizeRequest carries the entire re-resolved
// provisioning snapshot and only encrypted OAuth material. Plaintext OAuth
// tokens never cross this durable boundary.
type MailboxBootstrapOAuthFinalizeRequest struct {
	Account                  MailboxBootstrapOAuthAccount
	IssuerOrigin             string
	EngineIdentityID         string
	AccessTokenCiphertext    []byte
	RefreshTokenCiphertext   []byte
	AccessTokenExpiresAt     time.Time
	RefreshTokenExpiresAt    time.Time
	ObservedAt               time.Time
}

// MailboxBootstrapOAuthRepository resolves and atomically finalizes only the
// password-enrollment bootstrap flow. It must not perform provider I/O.
type MailboxBootstrapOAuthRepository interface {
	ResolveMailboxBootstrapOAuthAccount(context.Context, MailboxOAuthScope) (MailboxBootstrapOAuthAccount, error)
	FinalizeMailboxBootstrapOAuth(context.Context, MailboxBootstrapOAuthFinalizeRequest) error
}

// MailboxBootstrapOAuthService performs the short-lived OAuth bootstrap after
// password enrollment. It deliberately has no normal OAuth state, refresh, or
// reconnect behavior.
type MailboxBootstrapOAuthService struct {
	repository    MailboxBootstrapOAuthRepository
	states        MailboxBootstrapOAuthStateStore
	protocol      MailboxOAuthProtocol
	identityProof MailboxBootstrapIdentityProof
	cipher        *cryptopkg.Cipher
	clock         MailboxRepositoryClock
	issuerOrigin  string
}

// NewMailboxBootstrapOAuthService constructs a bootstrap-only OAuth service
// without issuing durable or provider calls.
func NewMailboxBootstrapOAuthService(repository MailboxBootstrapOAuthRepository, states MailboxBootstrapOAuthStateStore, protocol MailboxOAuthProtocol, identityProof MailboxBootstrapIdentityProof, cipher *cryptopkg.Cipher, clock MailboxRepositoryClock, issuerOrigin string) (*MailboxBootstrapOAuthService, error) {
	if mailboxEntNil(repository) || mailboxEntNil(states) || mailboxEntNil(protocol) || mailboxEntNil(identityProof) || cipher == nil || mailboxEntNil(clock) || !validMailboxOAuthIssuerOrigin(issuerOrigin) {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return &MailboxBootstrapOAuthService{
		repository: repository, states: states, protocol: protocol, identityProof: identityProof, cipher: cipher, clock: clock, issuerOrigin: issuerOrigin,
	}, nil
}

// BeginMailboxBootstrapOAuth binds one PKCE authorization to the exact
// pending password-enrollment snapshot and a random browser cookie.
func (service *MailboxBootstrapOAuthService) BeginMailboxBootstrapOAuth(ctx context.Context, scope MailboxOAuthScope) (MailboxOAuthBeginResult, error) {
	if !service.valid() || ctx == nil || !validMailboxOAuthScope(scope) {
		return MailboxOAuthBeginResult{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	account, err := service.repository.ResolveMailboxBootstrapOAuthAccount(ctx, scope)
	if err != nil {
		return MailboxOAuthBeginResult{}, mailboxOAuthBoundaryError(err)
	}
	if !validMailboxBootstrapOAuthAccount(account) || account.Scope != scope {
		return MailboxOAuthBeginResult{}, mailboxOAuthError(MailboxOAuthErrorNotFound)
	}
	authorization, err := service.protocol.BeginAuthorization(ctx)
	if err != nil {
		return MailboxOAuthBeginResult{}, mailboxOAuthProtocolError(err)
	}
	if !validMailboxOAuthAuthorization(authorization) {
		return MailboxOAuthBeginResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	bindingRaw := make([]byte, mailboxOAuthBrowserBindingBytes)
	defer clear(bindingRaw)
	if _, err := io.ReadFull(rand.Reader, bindingRaw); err != nil {
		return MailboxOAuthBeginResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	bindingHash := sha256.Sum256(bindingRaw)
	browserCookie := base64.RawURLEncoding.EncodeToString(bindingRaw)
	now, ok := mailboxOAuthNow(service.clock)
	if !ok {
		return MailboxOAuthBeginResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	expiresAt := now.Add(mailboxBootstrapOAuthStateTTL)
	if !validMailboxOAuthTime(expiresAt) {
		return MailboxOAuthBeginResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	pending := MailboxBootstrapOAuthPending{
		Scope:                        scope,
		EngineAccountID:              account.EngineAccountID,
		PasswordEnrollmentGeneration: account.PasswordEnrollmentGeneration,
		CodeVerifier:                 authorization.CodeVerifier,
		BrowserBindingHash:           bindingHash,
		ExpiresAt:                    expiresAt,
	}
	if err := service.states.SaveMailboxBootstrapOAuthState(ctx, authorization.State, pending); err != nil {
		return MailboxOAuthBeginResult{}, mailboxOAuthBoundaryError(err)
	}
	return MailboxOAuthBeginResult{AuthorizationURL: authorization.URL, ExpiresAt: expiresAt, BrowserCookie: browserCookie}, nil
}

// CompleteMailboxBootstrapOAuth consumes its callback state before any token
// exchange. Every failure after consumption therefore requires a fresh Begin.
func (service *MailboxBootstrapOAuthService) CompleteMailboxBootstrapOAuth(ctx context.Context, scope MailboxOAuthScope, state, code, browserCookie string) (MailboxOAuthConnectionResult, error) {
	if !service.valid() || ctx == nil || !validMailboxOAuthScope(scope) || !validMailboxOAuthBase64URL(state, 43) || !validMailboxOAuthOpaque(code, mailboxOAuthCodeMaximum) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	if !validMailboxOAuthBase64URL(browserCookie, mailboxOAuthBrowserBindingEncoded) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	bindingRaw, err := base64.RawURLEncoding.DecodeString(browserCookie)
	if err != nil || len(bindingRaw) != mailboxOAuthBrowserBindingBytes {
		clear(bindingRaw)
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	bindingHash := sha256.Sum256(bindingRaw)
	clear(bindingRaw)
	pending, err := service.states.ConsumeMailboxBootstrapOAuthState(ctx, state, bindingHash)
	if err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthBoundaryError(err)
	}
	now, ok := mailboxOAuthNow(service.clock)
	if !ok || !validMailboxBootstrapOAuthPending(pending) || pending.Scope != scope || subtle.ConstantTimeCompare(pending.BrowserBindingHash[:], bindingHash[:]) != 1 || !pending.ExpiresAt.After(now) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	account, err := service.repository.ResolveMailboxBootstrapOAuthAccount(ctx, scope)
	if err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthBoundaryError(err)
	}
	if !validMailboxBootstrapOAuthAccount(account) || account.Scope != scope || account.EngineAccountID != pending.EngineAccountID || account.PasswordEnrollmentGeneration != pending.PasswordEnrollmentGeneration {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	tokens, err := service.protocol.ExchangeAuthorizationCode(ctx, code, pending.CodeVerifier)
	defer clearMailboxOAuthTokenSet(&tokens)
	if err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthProtocolError(err)
	}
	if !validMailboxOAuthTokenSet(tokens, true) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	identity, err := service.identityProof.VerifyMailboxBootstrapIdentity(ctx, MailboxBootstrapIdentityProofRequest{
		EngineAccountID: account.EngineAccountID, ExpectedPrimaryEmail: account.PrimaryEmail, AccessToken: tokens.AccessToken,
	})
	if err != nil || !validMailboxBootstrapOAuthIdentityID(identity.EngineIdentityID) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	observedAt, ok := mailboxOAuthNow(service.clock)
	if !ok {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	accessExpiresAt, refreshExpiresAt, ok := mailboxOAuthExpiries(observedAt, tokens.ExpiresInSeconds)
	if !ok {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	oauthAccount := MailboxOAuthAccount{Scope: scope, EngineAccountID: account.EngineAccountID, CredentialGeneration: 1}
	accessCiphertext, err := service.encryptToken(tokens.AccessToken, oauthAccount, "access")
	if err != nil {
		return MailboxOAuthConnectionResult{}, err
	}
	defer clear(accessCiphertext)
	refreshCiphertext, err := service.encryptToken(tokens.RefreshToken, oauthAccount, "refresh")
	if err != nil {
		return MailboxOAuthConnectionResult{}, err
	}
	defer clear(refreshCiphertext)
	request := MailboxBootstrapOAuthFinalizeRequest{
		Account:                account,
		IssuerOrigin:           service.issuerOrigin,
		EngineIdentityID:       identity.EngineIdentityID,
		AccessTokenCiphertext:  accessCiphertext,
		RefreshTokenCiphertext: refreshCiphertext,
		AccessTokenExpiresAt:   accessExpiresAt,
		RefreshTokenExpiresAt:  refreshExpiresAt,
		ObservedAt:             observedAt,
	}
	if !validMailboxBootstrapOAuthFinalizeRequest(request) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if err := service.repository.FinalizeMailboxBootstrapOAuth(ctx, request); err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthBoundaryError(err)
	}
	return MailboxOAuthConnectionResult{AccountID: scope.AccountID, CredentialGeneration: 1, AccessTokenExpiresAt: accessExpiresAt}, nil
}

func (service *MailboxBootstrapOAuthService) encryptToken(token string, account MailboxOAuthAccount, kind string) ([]byte, error) {
	ciphertext, err := service.cipher.EncryptStringWithAAD(token, mailboxOAuthAAD(account, service.issuerOrigin, kind))
	if err != nil || len(ciphertext) < mailboxOAuthCiphertextMinimum || len(ciphertext) > mailboxOAuthCiphertextMaximum {
		return nil, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	return []byte(ciphertext), nil
}

func (service *MailboxBootstrapOAuthService) valid() bool {
	return service != nil && !mailboxEntNil(service.repository) && !mailboxEntNil(service.states) && !mailboxEntNil(service.protocol) && !mailboxEntNil(service.identityProof) && service.cipher != nil && !mailboxEntNil(service.clock) && validMailboxOAuthIssuerOrigin(service.issuerOrigin)
}

func validMailboxBootstrapOAuthAccount(account MailboxBootstrapOAuthAccount) bool {
	return validMailboxOAuthScope(account.Scope) && validMailboxOAuthOpaque(account.EngineAccountID, MaxEngineOpaqueIDBytes) && validMailboxBootstrapIdentityExpectedEmail(account.PrimaryEmail) && account.PasswordEnrollmentGeneration >= 1
}

func validMailboxBootstrapOAuthIdentityID(value string) bool {
	return validMailboxOAuthOpaque(value, MaxEngineOpaqueIDBytes)
}

func validMailboxBootstrapOAuthFinalizeRequest(request MailboxBootstrapOAuthFinalizeRequest) bool {
	return validMailboxBootstrapOAuthAccount(request.Account) && validMailboxOAuthIssuerOrigin(request.IssuerOrigin) && validMailboxBootstrapOAuthIdentityID(request.EngineIdentityID) && len(request.AccessTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(request.AccessTokenCiphertext) <= mailboxOAuthCiphertextMaximum && len(request.RefreshTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(request.RefreshTokenCiphertext) <= mailboxOAuthCiphertextMaximum && validMailboxOAuthTime(request.AccessTokenExpiresAt) && validMailboxOAuthTime(request.RefreshTokenExpiresAt) && validMailboxOAuthTime(request.ObservedAt) && request.AccessTokenExpiresAt.After(request.ObservedAt) && request.RefreshTokenExpiresAt.After(request.AccessTokenExpiresAt)
}
