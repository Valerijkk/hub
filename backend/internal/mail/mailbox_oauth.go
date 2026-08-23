package mail

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

const (
	mailboxOAuthProvider              = "stalwart"
	mailboxOAuthCanonicalScope        = "mail offline_access"
	mailboxOAuthStateTTL              = 10 * time.Minute
	mailboxOAuthRefreshLifetime       = 7 * 24 * time.Hour
	mailboxOAuthRefreshSafetySkew     = 60 * time.Second
	mailboxOAuthBrowserBindingBytes   = 32
	mailboxOAuthBrowserBindingEncoded = 43
	mailboxOAuthCiphertextMinimum     = 32
	mailboxOAuthCiphertextMaximum     = 16384
	mailboxOAuthIssuerMaximum         = 2048
	mailboxOAuthAuthorizationMaximum  = 8192
	mailboxOAuthCodeMaximum           = 8192
	// 16,384 base64 bytes decode to at most 12,288 bytes; AES-GCM consumes
	// another 12-byte nonce and 16-byte authentication tag.
	mailboxOAuthTokenMaximum          = 12260
	mailboxOAuthEncryptionFormat      = int32(1)
)

var (
	ErrMailboxOAuthInvalid           = errors.New("mailbox OAuth request invalid")
	ErrMailboxOAuthNotFound          = errors.New("mailbox OAuth account not found")
	ErrMailboxOAuthUnauthorized      = errors.New("mailbox OAuth account unauthorized")
	ErrMailboxOAuthStateInvalid      = errors.New("mailbox OAuth state invalid")
	ErrMailboxOAuthReconnectRequired = errors.New("mailbox OAuth reconnect required")
	ErrMailboxOAuthConflict          = errors.New("mailbox OAuth state conflict")
	ErrMailboxOAuthUnavailable       = errors.New("mailbox OAuth unavailable")
)

// MailboxOAuthErrorKind is deliberately content-free: provider responses,
// codes, verifiers, cookies and tokens never cross this error boundary.
type MailboxOAuthErrorKind string

const (
	MailboxOAuthErrorInvalid           MailboxOAuthErrorKind = "invalid"
	MailboxOAuthErrorNotFound          MailboxOAuthErrorKind = "not_found"
	MailboxOAuthErrorUnauthorized      MailboxOAuthErrorKind = "unauthorized"
	MailboxOAuthErrorStateInvalid      MailboxOAuthErrorKind = "state_invalid"
	MailboxOAuthErrorReconnectRequired MailboxOAuthErrorKind = "reconnect_required"
	MailboxOAuthErrorConflict          MailboxOAuthErrorKind = "conflict"
	MailboxOAuthErrorUnavailable       MailboxOAuthErrorKind = "unavailable"
)

type MailboxOAuthError struct{ Kind MailboxOAuthErrorKind }

func (value *MailboxOAuthError) Error() string {
	if value == nil {
		return "mailbox OAuth error"
	}
	return mailboxOAuthSentinel(value.Kind).Error()
}

func (value *MailboxOAuthError) Unwrap() error {
	if value == nil {
		return nil
	}
	return mailboxOAuthSentinel(value.Kind)
}

func mailboxOAuthError(kind MailboxOAuthErrorKind) error { return &MailboxOAuthError{Kind: kind} }

func mailboxOAuthSentinel(kind MailboxOAuthErrorKind) error {
	switch kind {
	case MailboxOAuthErrorInvalid:
		return ErrMailboxOAuthInvalid
	case MailboxOAuthErrorNotFound:
		return ErrMailboxOAuthNotFound
	case MailboxOAuthErrorUnauthorized:
		return ErrMailboxOAuthUnauthorized
	case MailboxOAuthErrorStateInvalid:
		return ErrMailboxOAuthStateInvalid
	case MailboxOAuthErrorReconnectRequired:
		return ErrMailboxOAuthReconnectRequired
	case MailboxOAuthErrorConflict:
		return ErrMailboxOAuthConflict
	case MailboxOAuthErrorUnavailable:
		return ErrMailboxOAuthUnavailable
	default:
		return ErrMailboxOAuthUnavailable
	}
}

type MailboxOAuthScope struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
}

type MailboxOAuthAccount struct {
	Scope                MailboxOAuthScope
	EngineAccountID      string
	CredentialGeneration int64
}

type MailboxOAuthPendingState struct {
	Scope                MailboxOAuthScope
	CredentialGeneration int64
	CodeVerifier         string
	BrowserBindingHash   [sha256.Size]byte
	ExpiresAt            time.Time
}

type MailboxOAuthCredentialSnapshot struct {
	AccountID              uuid.UUID
	OrganizationID         uuid.UUID
	ActorID                uuid.UUID
	EngineAccountID        string
	IssuerOrigin           string
	CredentialGeneration   int64
	Version                int
	AccessTokenCiphertext  []byte
	RefreshTokenCiphertext []byte
	AccessTokenExpiresAt   time.Time
	RefreshTokenExpiresAt  *time.Time
}

type MailboxOAuthStoreActiveRequest struct {
	Account                   MailboxOAuthAccount
	IssuerOrigin              string
	AccessTokenCiphertext     []byte
	RefreshTokenCiphertext    []byte
	AccessTokenExpiresAt      time.Time
	RefreshTokenExpiresAt     *time.Time
	ObservedAt                time.Time
}

type MailboxOAuthRefreshRequest struct {
	Scope                     MailboxOAuthScope
	ExpectedGeneration        int64
	ExpectedVersion           int
	AccessTokenCiphertext     []byte
	RefreshTokenCiphertext    []byte
	AccessTokenExpiresAt      time.Time
	RefreshTokenExpiresAt     *time.Time
	ObservedAt                time.Time
}

type MailboxOAuthReconnectReason string

const (
	MailboxOAuthReconnectInvalidGrant       MailboxOAuthReconnectReason = "invalid_grant"
	MailboxOAuthReconnectUnauthorizedAccount MailboxOAuthReconnectReason = "unauthorized_account"
	MailboxOAuthReconnectAccountMismatch     MailboxOAuthReconnectReason = "account_mismatch"
)

type MailboxOAuthRepository interface {
	ResolveMailboxOAuthAccount(context.Context, MailboxOAuthScope) (MailboxOAuthAccount, error)
	StoreMailboxOAuthActive(context.Context, MailboxOAuthStoreActiveRequest) error
	LoadMailboxOAuthActive(context.Context, MailboxOAuthScope) (MailboxOAuthCredentialSnapshot, error)
	RefreshMailboxOAuthActive(context.Context, MailboxOAuthRefreshRequest) error
	RequireMailboxOAuthReconnect(context.Context, MailboxOAuthScope, int64, int, MailboxOAuthReconnectReason, time.Time) error
	DisconnectMailboxOAuth(context.Context, MailboxOAuthScope, string, time.Time) error
}

type MailboxOAuthStateStore interface {
	SaveMailboxOAuthState(context.Context, string, MailboxOAuthPendingState) error
	ConsumeMailboxOAuthState(context.Context, string, [sha256.Size]byte) (MailboxOAuthPendingState, error)
}

type MailboxOAuthProtocol interface {
	BeginAuthorization(context.Context) (stalwart.OAuthAuthorization, error)
	ExchangeAuthorizationCode(context.Context, string, string) (stalwart.OAuthTokenSet, error)
	RefreshAccessToken(context.Context, string) (stalwart.OAuthTokenSet, error)
}

type MailboxOAuthAccountProofKind string

const (
	MailboxOAuthAccountVerified     MailboxOAuthAccountProofKind = "verified"
	MailboxOAuthAccountUnauthorized MailboxOAuthAccountProofKind = "unauthorized_account"
	MailboxOAuthAccountMismatch     MailboxOAuthAccountProofKind = "account_mismatch"
)

type MailboxOAuthAccountProofRequest struct {
	OrganizationID          uuid.UUID
	AccountID               uuid.UUID
	ExpectedEngineAccountID string
	AccessToken             string
}

type MailboxOAuthAccountProof interface {
	VerifyMailboxOAuthAccount(context.Context, MailboxOAuthAccountProofRequest) (MailboxOAuthAccountProofKind, error)
}

// MailboxOAuthRefreshAuthorizer is the narrow hook future HTTP bearer sourcing
// can use without coupling request handlers to refresh storage mechanics.
type MailboxOAuthRefreshAuthorizer interface {
	RefreshMailboxOAuthAuthorization(context.Context, MailboxOAuthScope) error
}

type MailboxOAuthBeginResult struct {
	AuthorizationURL string
	ExpiresAt        time.Time
	BrowserCookie    string
}

type MailboxOAuthConnectionResult struct {
	AccountID             uuid.UUID
	CredentialGeneration  int64
	AccessTokenExpiresAt  time.Time
}

type MailboxOAuthService struct {
	repository   MailboxOAuthRepository
	states       MailboxOAuthStateStore
	protocol     MailboxOAuthProtocol
	accountProof MailboxOAuthAccountProof
	cipher       *cryptopkg.Cipher
	clock        MailboxRepositoryClock
	issuerOrigin string
}

func NewMailboxOAuthService(repository MailboxOAuthRepository, states MailboxOAuthStateStore, protocol MailboxOAuthProtocol, accountProof MailboxOAuthAccountProof, cipher *cryptopkg.Cipher, clock MailboxRepositoryClock, issuerOrigin string) (*MailboxOAuthService, error) {
	if mailboxEntNil(repository) || mailboxEntNil(states) || mailboxEntNil(protocol) || mailboxEntNil(accountProof) || cipher == nil || mailboxEntNil(clock) || !validMailboxOAuthIssuerOrigin(issuerOrigin) {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return &MailboxOAuthService{repository: repository, states: states, protocol: protocol, accountProof: accountProof, cipher: cipher, clock: clock, issuerOrigin: issuerOrigin}, nil
}

func (service *MailboxOAuthService) BeginMailboxOAuth(ctx context.Context, scope MailboxOAuthScope) (MailboxOAuthBeginResult, error) {
	if !service.valid() || ctx == nil || !validMailboxOAuthScope(scope) {
		return MailboxOAuthBeginResult{}, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	account, err := service.repository.ResolveMailboxOAuthAccount(ctx, scope)
	if err != nil {
		return MailboxOAuthBeginResult{}, mailboxOAuthBoundaryError(err)
	}
	if !validMailboxOAuthAccount(account) || account.Scope != scope {
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
	expiresAt := now.Add(mailboxOAuthStateTTL)
	pending := MailboxOAuthPendingState{Scope: scope, CredentialGeneration: account.CredentialGeneration, CodeVerifier: authorization.CodeVerifier, BrowserBindingHash: bindingHash, ExpiresAt: expiresAt}
	if err := service.states.SaveMailboxOAuthState(ctx, authorization.State, pending); err != nil {
		return MailboxOAuthBeginResult{}, mailboxOAuthBoundaryError(err)
	}
	return MailboxOAuthBeginResult{AuthorizationURL: authorization.URL, ExpiresAt: expiresAt, BrowserCookie: browserCookie}, nil
}

func (service *MailboxOAuthService) CompleteMailboxOAuth(ctx context.Context, scope MailboxOAuthScope, state, code, browserCookie string) (MailboxOAuthConnectionResult, error) {
	if !service.valid() || ctx == nil || !validMailboxOAuthScope(scope) || !validMailboxOAuthOpaque(code, mailboxOAuthCodeMaximum) {
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
	pending, err := service.states.ConsumeMailboxOAuthState(ctx, state, bindingHash)
	if err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthBoundaryError(err)
	}
	now, ok := mailboxOAuthNow(service.clock)
	if !ok || !validMailboxOAuthPendingState(pending) || pending.Scope != scope || subtle.ConstantTimeCompare(pending.BrowserBindingHash[:], bindingHash[:]) != 1 || !pending.ExpiresAt.After(now) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	account, err := service.repository.ResolveMailboxOAuthAccount(ctx, scope)
	if err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthBoundaryError(err)
	}
	if !validMailboxOAuthAccount(account) || account.Scope != scope || account.CredentialGeneration != pending.CredentialGeneration {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	tokens, err := service.protocol.ExchangeAuthorizationCode(ctx, code, pending.CodeVerifier)
	if err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthProtocolError(err)
	}
	defer clearMailboxOAuthTokenSet(&tokens)
	if !validMailboxOAuthTokenSet(tokens, true) {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	proof, err := service.accountProof.VerifyMailboxOAuthAccount(ctx, MailboxOAuthAccountProofRequest{OrganizationID: scope.OrganizationID, AccountID: scope.AccountID, ExpectedEngineAccountID: account.EngineAccountID, AccessToken: tokens.AccessToken})
	if err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if proof != MailboxOAuthAccountVerified {
		return MailboxOAuthConnectionResult{}, mailboxOAuthProofError(proof)
	}
	observedAt, ok := mailboxOAuthNow(service.clock)
	if !ok {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	accessExpiresAt, refreshExpiresAt, ok := mailboxOAuthExpiries(observedAt, tokens.ExpiresInSeconds)
	if !ok {
		return MailboxOAuthConnectionResult{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	accessCiphertext, err := service.encryptMailboxOAuthToken(tokens.AccessToken, account, "access")
	if err != nil {
		return MailboxOAuthConnectionResult{}, err
	}
	refreshCiphertext, err := service.encryptMailboxOAuthToken(tokens.RefreshToken, account, "refresh")
	if err != nil {
		return MailboxOAuthConnectionResult{}, err
	}
	request := MailboxOAuthStoreActiveRequest{Account: account, IssuerOrigin: service.issuerOrigin, AccessTokenCiphertext: accessCiphertext, RefreshTokenCiphertext: refreshCiphertext, AccessTokenExpiresAt: accessExpiresAt, RefreshTokenExpiresAt: &refreshExpiresAt, ObservedAt: observedAt}
	if err := service.repository.StoreMailboxOAuthActive(ctx, request); err != nil {
		return MailboxOAuthConnectionResult{}, mailboxOAuthBoundaryError(err)
	}
	return MailboxOAuthConnectionResult{AccountID: scope.AccountID, CredentialGeneration: account.CredentialGeneration, AccessTokenExpiresAt: accessExpiresAt}, nil
}

func (service *MailboxOAuthService) RefreshMailboxOAuthAuthorization(ctx context.Context, scope MailboxOAuthScope) error {
	if !service.valid() || ctx == nil || !validMailboxOAuthScope(scope) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	snapshot, err := service.repository.LoadMailboxOAuthActive(ctx, scope)
	if err != nil {
		return mailboxOAuthBoundaryError(err)
	}
	if !validMailboxOAuthSnapshot(snapshot, scope) || snapshot.IssuerOrigin != service.issuerOrigin {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	now, ok := mailboxOAuthNow(service.clock)
	if !ok {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if snapshot.RefreshTokenExpiresAt == nil || !snapshot.RefreshTokenExpiresAt.After(now) {
		return service.requireReconnect(ctx, snapshot, MailboxOAuthReconnectInvalidGrant, now)
	}
	account := MailboxOAuthAccount{Scope: scope, EngineAccountID: snapshot.EngineAccountID, CredentialGeneration: snapshot.CredentialGeneration}
	refreshToken, err := service.cipher.DecryptStringWithAAD(string(snapshot.RefreshTokenCiphertext), mailboxOAuthAAD(account, snapshot.IssuerOrigin, "refresh"))
	if err != nil || !validMailboxOAuthOpaque(refreshToken, mailboxOAuthTokenMaximum) {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	defer func() { refreshToken = "" }()
	tokens, err := service.protocol.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		if kind, classified := stalwart.OAuthKindOf(err); classified && kind == stalwart.OAuthErrorInvalidGrant {
			return service.requireReconnect(ctx, snapshot, MailboxOAuthReconnectInvalidGrant, now)
		}
		return mailboxOAuthProtocolError(err)
	}
	defer clearMailboxOAuthTokenSet(&tokens)
	if !validMailboxOAuthTokenSet(tokens, true) {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	proof, err := service.accountProof.VerifyMailboxOAuthAccount(ctx, MailboxOAuthAccountProofRequest{OrganizationID: scope.OrganizationID, AccountID: scope.AccountID, ExpectedEngineAccountID: snapshot.EngineAccountID, AccessToken: tokens.AccessToken})
	if err != nil {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if proof == MailboxOAuthAccountUnauthorized {
		return service.requireReconnect(ctx, snapshot, MailboxOAuthReconnectUnauthorizedAccount, now)
	}
	if proof == MailboxOAuthAccountMismatch {
		return service.requireReconnect(ctx, snapshot, MailboxOAuthReconnectAccountMismatch, now)
	}
	if proof != MailboxOAuthAccountVerified {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	observedAt, ok := mailboxOAuthNow(service.clock)
	if !ok {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	accessExpiresAt, rotatedRefreshExpiresAt, ok := mailboxOAuthExpiries(observedAt, tokens.ExpiresInSeconds)
	if !ok {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	accessCiphertext, err := service.encryptMailboxOAuthToken(tokens.AccessToken, account, "access")
	if err != nil {
		return err
	}
	request := MailboxOAuthRefreshRequest{Scope: scope, ExpectedGeneration: snapshot.CredentialGeneration, ExpectedVersion: snapshot.Version, AccessTokenCiphertext: accessCiphertext, AccessTokenExpiresAt: accessExpiresAt, ObservedAt: observedAt}
	if subtle.ConstantTimeCompare([]byte(tokens.RefreshToken), []byte(refreshToken)) == 1 {
		if snapshot.RefreshTokenExpiresAt == nil || !snapshot.RefreshTokenExpiresAt.After(accessExpiresAt) {
			return mailboxOAuthError(MailboxOAuthErrorUnavailable)
		}
	} else {
		refreshCiphertext, err := service.encryptMailboxOAuthToken(tokens.RefreshToken, account, "refresh")
		if err != nil {
			return err
		}
		request.RefreshTokenCiphertext = refreshCiphertext
		request.RefreshTokenExpiresAt = &rotatedRefreshExpiresAt
	}
	if err := service.repository.RefreshMailboxOAuthActive(ctx, request); err != nil {
		return mailboxOAuthBoundaryError(err)
	}
	return nil
}

func (service *MailboxOAuthService) DisconnectMailboxOAuth(ctx context.Context, scope MailboxOAuthScope) error {
	if !service.valid() || ctx == nil || !validMailboxOAuthScope(scope) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	now, ok := mailboxOAuthNow(service.clock)
	if !ok {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	return mailboxOAuthBoundaryError(service.repository.DisconnectMailboxOAuth(ctx, scope, service.issuerOrigin, now))
}

func (service *MailboxOAuthService) requireReconnect(ctx context.Context, snapshot MailboxOAuthCredentialSnapshot, reason MailboxOAuthReconnectReason, now time.Time) error {
	err := service.repository.RequireMailboxOAuthReconnect(ctx, MailboxOAuthScope{ActorID: snapshot.ActorID, OrganizationID: snapshot.OrganizationID, AccountID: snapshot.AccountID}, snapshot.CredentialGeneration, snapshot.Version, reason, now)
	if err != nil {
		return mailboxOAuthBoundaryError(err)
	}
	return mailboxOAuthError(MailboxOAuthErrorReconnectRequired)
}

func (service *MailboxOAuthService) encryptMailboxOAuthToken(token string, account MailboxOAuthAccount, kind string) ([]byte, error) {
	ciphertext, err := service.cipher.EncryptStringWithAAD(token, mailboxOAuthAAD(account, service.issuerOrigin, kind))
	if err != nil || len(ciphertext) < mailboxOAuthCiphertextMinimum || len(ciphertext) > mailboxOAuthCiphertextMaximum {
		return nil, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	return []byte(ciphertext), nil
}

func (service *MailboxOAuthService) valid() bool {
	return service != nil && !mailboxEntNil(service.repository) && !mailboxEntNil(service.states) && !mailboxEntNil(service.protocol) && !mailboxEntNil(service.accountProof) && service.cipher != nil && !mailboxEntNil(service.clock) && validMailboxOAuthIssuerOrigin(service.issuerOrigin)
}

func mailboxOAuthAAD(account MailboxOAuthAccount, issuerOrigin, kind string) []byte {
	return []byte("hub.mail.oauth/v1\x00" + account.Scope.OrganizationID.String() + "\x00" + account.Scope.AccountID.String() + "\x00" + mailboxOAuthProvider + "\x00" + issuerOrigin + "\x00" + strconv.FormatInt(account.CredentialGeneration, 10) + "\x00" + kind)
}

func mailboxOAuthExpiries(observedAt time.Time, expiresInSeconds int64) (time.Time, time.Time, bool) {
	maximum := int64((mailboxOAuthRefreshLifetime - mailboxOAuthRefreshSafetySkew) / time.Second)
	if !validMailboxOAuthTime(observedAt) || expiresInSeconds < 1 || expiresInSeconds >= maximum {
		return time.Time{}, time.Time{}, false
	}
	accessExpiresAt := observedAt.Add(time.Duration(expiresInSeconds) * time.Second)
	refreshExpiresAt := observedAt.Add(mailboxOAuthRefreshLifetime - mailboxOAuthRefreshSafetySkew)
	return accessExpiresAt, refreshExpiresAt, validMailboxOAuthTime(accessExpiresAt) && validMailboxOAuthTime(refreshExpiresAt) && refreshExpiresAt.After(accessExpiresAt)
}

func mailboxOAuthNow(clock MailboxRepositoryClock) (time.Time, bool) {
	if mailboxEntNil(clock) {
		return time.Time{}, false
	}
	now := clock.Now()
	return now, validMailboxOAuthTime(now)
}

func validMailboxOAuthScope(scope MailboxOAuthScope) bool {
	return scope.ActorID != uuid.Nil && scope.OrganizationID != uuid.Nil && scope.AccountID != uuid.Nil
}

func validMailboxOAuthAccount(account MailboxOAuthAccount) bool {
	return validMailboxOAuthScope(account.Scope) && validMailboxOAuthOpaque(account.EngineAccountID, MaxEngineOpaqueIDBytes) && account.CredentialGeneration >= 1
}

func validMailboxOAuthPendingState(pending MailboxOAuthPendingState) bool {
	return validMailboxOAuthScope(pending.Scope) && pending.CredentialGeneration >= 1 && validMailboxOAuthBase64URL(pending.CodeVerifier, 86) && pending.BrowserBindingHash != [sha256.Size]byte{} && validMailboxOAuthTime(pending.ExpiresAt)
}

func validMailboxOAuthAuthorization(value stalwart.OAuthAuthorization) bool {
	return len(value.URL) >= 1 && len(value.URL) <= mailboxOAuthAuthorizationMaximum && validMailboxOAuthBase64URL(value.State, 43) && validMailboxOAuthBase64URL(value.CodeVerifier, 86)
}

func validMailboxOAuthTokenSet(tokens stalwart.OAuthTokenSet, requireRefresh bool) bool {
	return tokens.TokenType == "Bearer" && tokens.Scope == "offline_access mail" && validMailboxOAuthOpaque(tokens.AccessToken, mailboxOAuthTokenMaximum) && (!requireRefresh || validMailboxOAuthOpaque(tokens.RefreshToken, mailboxOAuthTokenMaximum)) && tokens.ExpiresInSeconds >= 1
}

func validMailboxOAuthSnapshot(snapshot MailboxOAuthCredentialSnapshot, scope MailboxOAuthScope) bool {
	return snapshot.AccountID == scope.AccountID && snapshot.OrganizationID == scope.OrganizationID && snapshot.ActorID == scope.ActorID && validMailboxOAuthOpaque(snapshot.EngineAccountID, MaxEngineOpaqueIDBytes) && validMailboxOAuthIssuerOrigin(snapshot.IssuerOrigin) && snapshot.CredentialGeneration >= 1 && snapshot.Version >= 1 && len(snapshot.AccessTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(snapshot.AccessTokenCiphertext) <= mailboxOAuthCiphertextMaximum && len(snapshot.RefreshTokenCiphertext) >= mailboxOAuthCiphertextMinimum && len(snapshot.RefreshTokenCiphertext) <= mailboxOAuthCiphertextMaximum && validMailboxOAuthTime(snapshot.AccessTokenExpiresAt) && snapshot.RefreshTokenExpiresAt != nil && validMailboxOAuthTime(*snapshot.RefreshTokenExpiresAt) && snapshot.RefreshTokenExpiresAt.After(snapshot.AccessTokenExpiresAt)
}

func validMailboxOAuthOpaque(value string, maximum int) bool {
	if value == "" || len(value) > maximum || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character <= 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func validMailboxOAuthBase64URL(value string, exact int) bool {
	if len(value) != exact {
		return false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func validMailboxOAuthTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}

func validMailboxOAuthIssuerOrigin(value string) bool {
	if value == "" || len(value) > mailboxOAuthIssuerMaximum || !utf8.ValidString(value) {
		return false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Opaque != "" || parsed.User != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawPath != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return false
	}
	hostname := parsed.Hostname()
	if hostname == "" || hostname != strings.ToLower(hostname) || strings.HasSuffix(hostname, ".") {
		return false
	}
	canonicalHost := hostname
	if address := net.ParseIP(hostname); address != nil {
		if address.String() != hostname {
			return false
		}
		if strings.Contains(hostname, ":") {
			canonicalHost = "[" + hostname + "]"
		}
	} else if validateDomain(hostname) != nil {
		return false
	}
	if port := parsed.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 || strconv.Itoa(number) != port || number == 443 {
			return false
		}
		canonicalHost += ":" + port
	}
	return value == "https://"+canonicalHost
}

func mailboxOAuthProofError(kind MailboxOAuthAccountProofKind) error {
	if kind == MailboxOAuthAccountUnauthorized || kind == MailboxOAuthAccountMismatch {
		return mailboxOAuthError(MailboxOAuthErrorUnauthorized)
	}
	return mailboxOAuthError(MailboxOAuthErrorUnavailable)
}

func mailboxOAuthProtocolError(err error) error {
	kind, ok := stalwart.OAuthKindOf(err)
	if !ok {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	switch kind {
	case stalwart.OAuthErrorInvalidRequest, stalwart.OAuthErrorInvalidGrant:
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	case stalwart.OAuthErrorCancelled, stalwart.OAuthErrorUnavailable, stalwart.OAuthErrorInvalidClient, stalwart.OAuthErrorInvalidConfiguration, stalwart.OAuthErrorProtocol, stalwart.OAuthErrorPayloadTooLarge:
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	default:
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
}

func mailboxOAuthBoundaryError(err error) error {
	if err == nil {
		return nil
	}
	var oauthError *MailboxOAuthError
	if errors.As(err, &oauthError) && oauthError != nil {
		return mailboxOAuthError(oauthError.Kind)
	}
	return mailboxOAuthError(MailboxOAuthErrorUnavailable)
}

func clearMailboxOAuthTokenSet(tokens *stalwart.OAuthTokenSet) {
	if tokens == nil {
		return
	}
	tokens.AccessToken = ""
	tokens.RefreshToken = ""
	tokens.TokenType = ""
	tokens.Scope = ""
}

var _ MailboxOAuthRefreshAuthorizer = (*MailboxOAuthService)(nil)
