package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartMailboxBootstrapIdentityGetCallID              = "bootstrap-identity-get"
	stalwartMailboxBootstrapIdentityMaximumResponseBytes int64 = 32 * 1024
)

// MailboxBootstrapIdentityProof proves one externally bootstrapped mailbox
// identity without exposing a provider transport, a bearer source, or mutable
// mailbox capability to its caller.
type MailboxBootstrapIdentityProof interface {
	VerifyMailboxBootstrapIdentity(context.Context, MailboxBootstrapIdentityProofRequest) (MailboxBootstrapIdentityProofResult, error)
}

// MailboxBootstrapIdentityProofRequest carries one already-normalized primary
// address and an ephemeral OAuth bearer. The proof consumes the bearer only
// for this read-only verification and never retains it after the call returns.
type MailboxBootstrapIdentityProofRequest struct {
	EngineAccountID      string
	ExpectedPrimaryEmail string
	AccessToken          string
}

// MailboxBootstrapIdentityProofResult intentionally returns only the opaque
// engine identity required by the next bootstrap step.
type MailboxBootstrapIdentityProofResult struct {
	EngineIdentityID string
}

type stalwartMailboxBootstrapIdentityProof struct {
	origin             string
	httpClient         *http.Client
	registryAuthorizer stalwart.RegistryAuthorizer
}

type stalwartMailboxBootstrapIdentityGetResponse struct {
	AccountID *string                                    `json:"accountId"`
	State     *string                                    `json:"state"`
	List      *[]stalwartMailboxBootstrapIdentityRecord `json:"list"`
	NotFound  *[]string                                  `json:"notFound"`
}

type stalwartMailboxBootstrapIdentityRecord struct {
	ID    *string `json:"id"`
	Email *string `json:"email"`
}

// NewStalwartMailboxBootstrapIdentityProof constructs a one-shot identity
// proof. The registry authorizer is required by the typed Stalwart client but
// this proof never calls registry methods; registry use therefore fails closed
// at the supplied capability boundary.
func NewStalwartMailboxBootstrapIdentityProof(origin string, httpClient *http.Client, registryAuthorizer stalwart.RegistryAuthorizer) (MailboxBootstrapIdentityProof, error) {
	canonicalOrigin, ok := canonicalStalwartMailboxOrigin(origin)
	if !ok || httpClient == nil || registryAuthorizer == nil {
		return nil, stalwart.ErrInvalidConfiguration
	}
	return &stalwartMailboxBootstrapIdentityProof{
		origin: canonicalOrigin, httpClient: httpClient, registryAuthorizer: registryAuthorizer,
	}, nil
}

// VerifyMailboxBootstrapIdentity discovers exactly one mailbox session, proves
// its exact account binding, and performs one bounded Identity/get request.
// It has no retry, mutation, database, or persisted-token path.
func (proof *stalwartMailboxBootstrapIdentityProof) VerifyMailboxBootstrapIdentity(ctx context.Context, request MailboxBootstrapIdentityProofRequest) (MailboxBootstrapIdentityProofResult, error) {
	if proof == nil || proof.httpClient == nil || proof.registryAuthorizer == nil || ctx == nil || !validAccountID(request.EngineAccountID) || !validMailboxBootstrapIdentityExpectedEmail(request.ExpectedPrimaryEmail) || len(request.AccessToken) < 1 || len(request.AccessToken) > mailboxOAuthTokenMaximum {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrInvalidRequest
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return MailboxBootstrapIdentityProofResult{}, err
	}

	token := []byte(request.AccessToken)
	request.AccessToken = ""
	defer clear(token)
	if !validStalwartMailboxBearerToken(token) {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrInvalidRequest
	}

	authorizer := &fixedStalwartMailboxBearerAuthorizer{accessToken: token}
	defer authorizer.clear()
	client, err := stalwart.NewClient(stalwart.Config{
		Origin: proof.origin, HTTPClient: proof.httpClient,
		RegistryAuthorizer: proof.registryAuthorizer, MailboxAuthorizer: authorizer,
	})
	if err != nil {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrInvalidConfiguration
	}

	session, err := client.DiscoverMailboxSession(ctx)
	if err != nil {
		return MailboxBootstrapIdentityProofResult{}, err
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return MailboxBootstrapIdentityProofResult{}, err
	}
	if session == nil || !session.HasAccount(request.EngineAccountID) {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrUnauthorized
	}

	arguments, err := json.Marshal(struct {
		AccountID  string   `json:"accountId"`
		Properties []string `json:"properties"`
	}{AccountID: request.EngineAccountID, Properties: []string{"id", "email"}})
	if err != nil {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrInvalidRequest
	}
	response, err := client.CallMailboxBounded(ctx, session, stalwart.JMAPRequest{
		Using: append([]string(nil), stalwartMailboxCapabilities...),
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name: "Identity/get", Arguments: arguments, CallID: stalwartMailboxBootstrapIdentityGetCallID,
		}},
	}, stalwartMailboxBootstrapIdentityMaximumResponseBytes)
	if err != nil {
		return MailboxBootstrapIdentityProofResult{}, err
	}
	if err := stalwartMailboxContextError(ctx); err != nil {
		return MailboxBootstrapIdentityProofResult{}, err
	}
	return parseStalwartMailboxBootstrapIdentityProofResponse(response, request.EngineAccountID, request.ExpectedPrimaryEmail)
}

func validMailboxBootstrapIdentityExpectedEmail(value string) bool {
	canonical, err := NormalizeAddress(EngineAddress{Email: value})
	return err == nil && canonical.Email == value
}

func parseStalwartMailboxBootstrapIdentityProofResponse(response stalwart.JMAPResponse, accountID, expectedEmail string) (MailboxBootstrapIdentityProofResult, error) {
	if len(response.MethodResponses) != 1 {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrProtocol
	}
	method := response.MethodResponses[0]
	if method.Name != "Identity/get" || method.CallID != stalwartMailboxBootstrapIdentityGetCallID || len(method.Arguments) == 0 {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrProtocol
	}

	decoder := json.NewDecoder(bytes.NewReader(method.Arguments))
	decoder.DisallowUnknownFields()
	var identities stalwartMailboxBootstrapIdentityGetResponse
	if err := decoder.Decode(&identities); err != nil {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrProtocol
	}
	if err := requireStalwartMailboxBootstrapIdentityEOF(decoder); err != nil {
		return MailboxBootstrapIdentityProofResult{}, err
	}
	if identities.AccountID == nil || *identities.AccountID != accountID {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrUnauthorized
	}
	if identities.State == nil || !validOpaqueID(*identities.State) || identities.List == nil || identities.NotFound == nil {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrProtocol
	}
	if len(*identities.NotFound) != 0 {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrProtocol
	}
	if len(*identities.List) != 1 {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrStateMismatch
	}

	identity := (*identities.List)[0]
	if identity.ID == nil || identity.Email == nil || !validOpaqueID(*identity.ID) {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrProtocol
	}
	canonical, err := NormalizeAddress(EngineAddress{Email: *identity.Email})
	if err != nil {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrProtocol
	}
	if canonical.Email != expectedEmail {
		return MailboxBootstrapIdentityProofResult{}, stalwart.ErrStateMismatch
	}
	return MailboxBootstrapIdentityProofResult{EngineIdentityID: *identity.ID}, nil
}

func requireStalwartMailboxBootstrapIdentityEOF(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return stalwart.ErrProtocol
	}
	return nil
}

var _ MailboxBootstrapIdentityProof = (*stalwartMailboxBootstrapIdentityProof)(nil)
