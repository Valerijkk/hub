package mail

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// MailboxOAuthBearerSnapshotSource resolves one authoritative active OAuth
// credential snapshot by its exact provider mailbox account identifier.
type MailboxOAuthBearerSnapshotSource interface {
	LoadMailboxOAuthBearerSnapshot(context.Context, string) (MailboxOAuthCredentialSnapshot, error)
}

type stalwartMailboxAuthorizerSource struct {
	snapshots    MailboxOAuthBearerSnapshotSource
	refresher    MailboxOAuthRefreshAuthorizer
	cipher       *cryptopkg.Cipher
	clock        MailboxRepositoryClock
	issuerOrigin string
}

type accountBoundStalwartMailboxAuthorizer struct {
	accountID    string
	snapshots    MailboxOAuthBearerSnapshotSource
	refresher    MailboxOAuthRefreshAuthorizer
	cipher       *cryptopkg.Cipher
	clock        MailboxRepositoryClock
	issuerOrigin string
}

type stalwartMailboxOAuthAccountProof struct {
	origin             string
	httpClient         *http.Client
	registryAuthorizer stalwart.RegistryAuthorizer
}

type fixedStalwartMailboxBearerAuthorizer struct {
	accessToken []byte
}

// NewStalwartMailboxAuthorizerSource constructs a dynamic account-bound OAuth
// bearer source without loading or retaining plaintext access tokens.
func NewStalwartMailboxAuthorizerSource(issuerOrigin string, snapshots MailboxOAuthBearerSnapshotSource, refresher MailboxOAuthRefreshAuthorizer, cipher *cryptopkg.Cipher, clock MailboxRepositoryClock) (MailboxAuthorizerSource, error) {
	if !validMailboxOAuthIssuerOrigin(issuerOrigin) || mailboxEntNil(snapshots) || mailboxEntNil(refresher) || cipher == nil || mailboxEntNil(clock) {
		return nil, stalwart.ErrInvalidConfiguration
	}
	return &stalwartMailboxAuthorizerSource{snapshots: snapshots, refresher: refresher, cipher: cipher, clock: clock, issuerOrigin: issuerOrigin}, nil
}

func (source *stalwartMailboxAuthorizerSource) OAuthBearerAuthorizer(ctx context.Context, accountID string) (MailboxAuthorizerBinding, error) {
	if !source.valid() || ctx == nil || !validAccountID(accountID) {
		return MailboxAuthorizerBinding{}, stalwart.ErrUnauthorized
	}
	if err := stalwartMailboxOAuthContextError(ctx); err != nil {
		return MailboxAuthorizerBinding{}, err
	}
	return MailboxAuthorizerBinding{
		AccountID: accountID,
		Authorizer: &accountBoundStalwartMailboxAuthorizer{
			accountID: accountID, snapshots: source.snapshots, refresher: source.refresher,
			cipher: source.cipher, clock: source.clock, issuerOrigin: source.issuerOrigin,
		},
	}, nil
}

func (source *stalwartMailboxAuthorizerSource) valid() bool {
	return source != nil && !mailboxEntNil(source.snapshots) && !mailboxEntNil(source.refresher) && source.cipher != nil && !mailboxEntNil(source.clock) && validMailboxOAuthIssuerOrigin(source.issuerOrigin)
}

func (authorizer *accountBoundStalwartMailboxAuthorizer) AuthorizeMailbox(ctx context.Context, request *http.Request) error {
	if !authorizer.valid() || ctx == nil || request == nil || request.Header == nil {
		return stalwart.ErrUnauthorized
	}
	if err := stalwartMailboxOAuthContextError(ctx); err != nil {
		return err
	}

	snapshot, err := authorizer.snapshots.LoadMailboxOAuthBearerSnapshot(ctx, authorizer.accountID)
	if err != nil {
		return safeStalwartMailboxOAuthAuthorizationError(ctx, err)
	}
	defer clearStalwartMailboxOAuthSnapshot(&snapshot)
	if !validStalwartMailboxOAuthBearerSnapshot(snapshot, authorizer.accountID, authorizer.issuerOrigin) {
		return stalwart.ErrUnavailable
	}

	now, ok := mailboxOAuthNow(authorizer.clock)
	if !ok {
		return stalwart.ErrUnavailable
	}
	refreshBoundary := now.Add(mailboxOAuthRefreshSafetySkew)
	if !snapshot.AccessTokenExpiresAt.After(refreshBoundary) {
		previousScope := MailboxOAuthScope{ActorID: snapshot.ActorID, OrganizationID: snapshot.OrganizationID, AccountID: snapshot.AccountID}
		previousGeneration := snapshot.CredentialGeneration
		previousVersion := snapshot.Version
		previousIssuer := snapshot.IssuerOrigin
		if err := authorizer.refresher.RefreshMailboxOAuthAuthorization(ctx, previousScope); err != nil {
			return safeStalwartMailboxOAuthAuthorizationError(ctx, err)
		}
		clearStalwartMailboxOAuthSnapshot(&snapshot)
		snapshot, err = authorizer.snapshots.LoadMailboxOAuthBearerSnapshot(ctx, authorizer.accountID)
		if err != nil {
			return safeStalwartMailboxOAuthAuthorizationError(ctx, err)
		}
		refreshedScope := MailboxOAuthScope{ActorID: snapshot.ActorID, OrganizationID: snapshot.OrganizationID, AccountID: snapshot.AccountID}
		if !validStalwartMailboxOAuthBearerSnapshot(snapshot, authorizer.accountID, authorizer.issuerOrigin) ||
			refreshedScope != previousScope ||
			snapshot.CredentialGeneration != previousGeneration || snapshot.Version <= previousVersion || snapshot.IssuerOrigin != previousIssuer {
			return stalwart.ErrUnavailable
		}
		now, ok = mailboxOAuthNow(authorizer.clock)
		if !ok || !snapshot.AccessTokenExpiresAt.After(now.Add(mailboxOAuthRefreshSafetySkew)) {
			return stalwart.ErrUnavailable
		}
	}
	if err := stalwartMailboxOAuthContextError(ctx); err != nil {
		return err
	}

	account := MailboxOAuthAccount{
		Scope: MailboxOAuthScope{ActorID: snapshot.ActorID, OrganizationID: snapshot.OrganizationID, AccountID: snapshot.AccountID},
		EngineAccountID: snapshot.EngineAccountID, CredentialGeneration: snapshot.CredentialGeneration,
	}
	aad := mailboxOAuthAAD(account, snapshot.IssuerOrigin, "access")
	defer clear(aad)
	token, err := authorizer.cipher.DecryptWithAAD(string(snapshot.AccessTokenCiphertext), aad)
	if err != nil {
		clear(token)
		return stalwart.ErrUnavailable
	}
	defer clear(token)
	if !validStalwartMailboxBearerToken(token) {
		return stalwart.ErrUnavailable
	}
	if err := stalwartMailboxOAuthContextError(ctx); err != nil {
		return err
	}

	value := "Bearer " + string(token)
	request.Header.Set("Authorization", value)
	value = ""
	return nil
}

func (authorizer *accountBoundStalwartMailboxAuthorizer) valid() bool {
	return authorizer != nil && validAccountID(authorizer.accountID) && !mailboxEntNil(authorizer.snapshots) && !mailboxEntNil(authorizer.refresher) && authorizer.cipher != nil && !mailboxEntNil(authorizer.clock) && validMailboxOAuthIssuerOrigin(authorizer.issuerOrigin)
}

// NewStalwartMailboxOAuthAccountProof constructs the read-only account proof
// boundary. It performs no discovery or credential access until verification.
func NewStalwartMailboxOAuthAccountProof(origin string, httpClient *http.Client, registryAuthorizer stalwart.RegistryAuthorizer) (MailboxOAuthAccountProof, error) {
	canonicalOrigin, ok := canonicalStalwartMailboxOrigin(origin)
	if !ok || httpClient == nil || registryAuthorizer == nil {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return &stalwartMailboxOAuthAccountProof{origin: canonicalOrigin, httpClient: httpClient, registryAuthorizer: registryAuthorizer}, nil
}

func (proof *stalwartMailboxOAuthAccountProof) VerifyMailboxOAuthAccount(ctx context.Context, request MailboxOAuthAccountProofRequest) (MailboxOAuthAccountProofKind, error) {
	if proof == nil || proof.httpClient == nil || proof.registryAuthorizer == nil || ctx == nil || request.OrganizationID == uuid.Nil || request.AccountID == uuid.Nil || !validMailboxOAuthOpaque(request.ExpectedEngineAccountID, MaxEngineOpaqueIDBytes) || len(request.AccessToken) < 1 || len(request.AccessToken) > mailboxOAuthTokenMaximum {
		return "", mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	if err := stalwartMailboxOAuthContextError(ctx); err != nil {
		return "", mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	token := []byte(request.AccessToken)
	request.AccessToken = ""
	defer clear(token)
	if !validStalwartMailboxBearerToken(token) {
		return "", mailboxOAuthError(MailboxOAuthErrorInvalid)
	}

	authorizer := &fixedStalwartMailboxBearerAuthorizer{accessToken: token}
	defer authorizer.clear()
	client, err := stalwart.NewClient(stalwart.Config{
		Origin: proof.origin, HTTPClient: proof.httpClient,
		RegistryAuthorizer: proof.registryAuthorizer, MailboxAuthorizer: authorizer,
	})
	if err != nil {
		return "", mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	session, err := client.DiscoverMailboxSession(ctx)
	if err != nil {
		if kind, classified := stalwart.KindOf(err); classified && kind == stalwart.ErrorUnauthorized {
			return MailboxOAuthAccountUnauthorized, nil
		}
		return "", mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if session == nil {
		return "", mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if !session.HasAccount(request.ExpectedEngineAccountID) {
		return MailboxOAuthAccountMismatch, nil
	}
	return MailboxOAuthAccountVerified, nil
}

func (authorizer *fixedStalwartMailboxBearerAuthorizer) AuthorizeMailbox(ctx context.Context, request *http.Request) error {
	if authorizer == nil || ctx == nil || request == nil || request.Header == nil || !validStalwartMailboxBearerToken(authorizer.accessToken) {
		return stalwart.ErrUnauthorized
	}
	if err := stalwartMailboxOAuthContextError(ctx); err != nil {
		return err
	}
	value := "Bearer " + string(authorizer.accessToken)
	request.Header.Set("Authorization", value)
	value = ""
	return nil
}

func (authorizer *fixedStalwartMailboxBearerAuthorizer) clear() {
	if authorizer == nil {
		return
	}
	clear(authorizer.accessToken)
	authorizer.accessToken = nil
}

func validStalwartMailboxOAuthBearerSnapshot(snapshot MailboxOAuthCredentialSnapshot, accountID, issuerOrigin string) bool {
	scope := MailboxOAuthScope{ActorID: snapshot.ActorID, OrganizationID: snapshot.OrganizationID, AccountID: snapshot.AccountID}
	return snapshot.EngineAccountID == accountID && snapshot.IssuerOrigin == issuerOrigin && validMailboxOAuthSnapshot(snapshot, scope)
}

func validStalwartMailboxBearerToken(token []byte) bool {
	if len(token) < 1 || len(token) > mailboxOAuthTokenMaximum {
		return false
	}
	padding := false
	content := false
	for _, character := range token {
		if character == '=' {
			padding = true
			continue
		}
		if padding || !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("-._~+/", rune(character))) {
			return false
		}
		content = true
	}
	return content
}

func clearStalwartMailboxOAuthSnapshot(snapshot *MailboxOAuthCredentialSnapshot) {
	if snapshot == nil {
		return
	}
	clear(snapshot.AccessTokenCiphertext)
	clear(snapshot.RefreshTokenCiphertext)
	snapshot.AccessTokenCiphertext = nil
	snapshot.RefreshTokenCiphertext = nil
}

func stalwartMailboxOAuthContextError(ctx context.Context) error {
	if ctx == nil {
		return stalwart.ErrUnauthorized
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return stalwart.ErrCancelled
		}
		return stalwart.ErrUnavailable
	}
	return nil
}

func safeStalwartMailboxOAuthAuthorizationError(ctx context.Context, err error) error {
	if contextError := stalwartMailboxOAuthContextError(ctx); contextError != nil {
		return contextError
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, stalwart.ErrCancelled) {
		return stalwart.ErrCancelled
	}
	if errors.Is(err, ErrMailboxOAuthNotFound) || errors.Is(err, ErrMailboxOAuthUnauthorized) || errors.Is(err, ErrMailboxOAuthReconnectRequired) || errors.Is(err, stalwart.ErrUnauthorized) {
		return stalwart.ErrUnauthorized
	}
	return stalwart.ErrUnavailable
}

var _ MailboxAuthorizerSource = (*stalwartMailboxAuthorizerSource)(nil)
var _ stalwart.MailboxAuthorizer = (*accountBoundStalwartMailboxAuthorizer)(nil)
var _ MailboxOAuthAccountProof = (*stalwartMailboxOAuthAccountProof)(nil)
var _ stalwart.MailboxAuthorizer = (*fixedStalwartMailboxBearerAuthorizer)(nil)
