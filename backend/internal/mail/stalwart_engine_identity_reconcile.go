package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sort"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartIdentityReconcileGetCallID                = "identity-reconcile-get"
	stalwartIdentityReconcileResponseMaximumBytes int64 = MaxCalendarPartBytes
)

// MailboxIdentityReconcileRequest names the one provider mailbox whose
// identities are being observed. The caller obtains its authoritative active
// and enabled account facts before crossing this provider boundary.
type MailboxIdentityReconcileRequest struct {
	EngineAccountID string
}

// MailboxIdentityReconcileIdentity is one canonical address and its stable
// provider-assigned opaque identity ID. It intentionally carries no display
// name, provider state, credential, or mutable identity settings.
type MailboxIdentityReconcileIdentity struct {
	EngineIdentityID string
	Email            string
}

// MailboxIdentityReconcileSnapshot is a bounded, deterministic read-only
// provider observation. Desired-address matching and durable state changes
// belong to the later reconciliation worker, not this adapter.
type MailboxIdentityReconcileSnapshot struct {
	EngineAccountID string
	Identities      []MailboxIdentityReconcileIdentity
}

// MailboxIdentityReconcileProvider is an internal-package execution seam for
// the dedicated identity reconciliation worker and its external-package unit
// tests. It is deliberately outside MailEngine and is not a product-facing
// public API, provider mutation capability, or activation authority.
type MailboxIdentityReconcileProvider interface {
	ObserveMailboxIdentityReconcile(context.Context, MailboxIdentityReconcileRequest) (MailboxIdentityReconcileSnapshot, error)
}

// stalwartIdentityReconcileBoundedMailboxClient keeps response-bound identity
// observation opt-in. Existing mailbox clients do not gain a transport
// requirement, and an unavailable bounded capability fails closed.
type stalwartIdentityReconcileBoundedMailboxClient interface {
	CallMailboxBounded(context.Context, stalwart.JMAPRequest, int64) (stalwart.JMAPResponse, error)
}

type stalwartIdentityReconcileGetResponse struct {
	AccountID *string                                    `json:"accountId"`
	State     *string                                    `json:"state"`
	List      *[]stalwartIdentityReconcileProviderRecord `json:"list"`
	NotFound  *[]string                                  `json:"notFound"`
}

type stalwartIdentityReconcileProviderRecord struct {
	ID    *string `json:"id"`
	Email *string `json:"email"`
}

// ObserveMailboxIdentityReconcile performs one bounded, account-bound
// Identity/get observation through the existing OAuth-backed mailbox client
// factory. It never uses a password, bootstrap bearer, registry endpoint, or
// identity mutation path.
func (engine *StalwartEngine) ObserveMailboxIdentityReconcile(ctx context.Context, request MailboxIdentityReconcileRequest) (MailboxIdentityReconcileSnapshot, error) {
	if ctx == nil || !validAccountID(request.EngineAccountID) {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	client, err := engine.mailboxClient(ctx, request.EngineAccountID)
	if err != nil {
		return MailboxIdentityReconcileSnapshot{}, err
	}
	boundedClient, ok := client.(stalwartIdentityReconcileBoundedMailboxClient)
	if !ok {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
	}
	arguments, err := json.Marshal(stalwartIdentityGetArguments{
		AccountID:  request.EngineAccountID,
		Properties: []string{"id", "email"},
	})
	if err != nil {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := boundedClient.CallMailboxBounded(ctx, stalwart.JMAPRequest{
		Using: append([]string(nil), stalwartMailboxCapabilities...),
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Identity/get",
			Arguments: arguments,
			CallID:    stalwartIdentityReconcileGetCallID,
		}},
	}, stalwartIdentityReconcileResponseMaximumBytes)
	if err != nil {
		return MailboxIdentityReconcileSnapshot{}, mapStalwartEngineError(err, false)
	}
	return parseStalwartIdentityReconcileSnapshot(response, request.EngineAccountID)
}

func parseStalwartIdentityReconcileSnapshot(response stalwart.JMAPResponse, accountID string) (MailboxIdentityReconcileSnapshot, error) {
	if !validAccountID(accountID) || len(response.MethodResponses) != 1 {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
	}
	method := response.MethodResponses[0]
	if method.Name != "Identity/get" || method.CallID != stalwartIdentityReconcileGetCallID || len(method.Arguments) == 0 || int64(len(method.Arguments)) > stalwartIdentityReconcileResponseMaximumBytes {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
	}

	decoder := json.NewDecoder(bytes.NewReader(method.Arguments))
	decoder.DisallowUnknownFields()
	var identities stalwartIdentityReconcileGetResponse
	if err := decoder.Decode(&identities); err != nil {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if err := requireStalwartIdentityReconcileEOF(decoder); err != nil {
		return MailboxIdentityReconcileSnapshot{}, err
	}
	if identities.AccountID == nil {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if *identities.AccountID != accountID {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if identities.State == nil || !validOpaqueID(*identities.State) || identities.List == nil || identities.NotFound == nil || len(*identities.List) > MaxEnginePageItems || len(*identities.NotFound) != 0 {
		return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
	}

	result := MailboxIdentityReconcileSnapshot{
		EngineAccountID: accountID,
		Identities:      make([]MailboxIdentityReconcileIdentity, 0, len(*identities.List)),
	}
	seenIDs := make(map[string]struct{}, len(*identities.List))
	seenEmails := make(map[string]struct{}, len(*identities.List))
	for _, identity := range *identities.List {
		if identity.ID == nil || identity.Email == nil || !validOpaqueID(*identity.ID) {
			return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
		}
		canonical, err := NormalizeAddress(EngineAddress{Email: *identity.Email})
		if err != nil {
			return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
		}
		if _, duplicate := seenIDs[*identity.ID]; duplicate {
			return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
		}
		if _, duplicate := seenEmails[canonical.Email]; duplicate {
			return MailboxIdentityReconcileSnapshot{}, stalwartEngineError(EngineErrorUnavailable)
		}
		seenIDs[*identity.ID] = struct{}{}
		seenEmails[canonical.Email] = struct{}{}
		result.Identities = append(result.Identities, MailboxIdentityReconcileIdentity{
			EngineIdentityID: *identity.ID,
			Email:            canonical.Email,
		})
	}
	sort.Slice(result.Identities, func(left, right int) bool {
		if result.Identities[left].Email != result.Identities[right].Email {
			return result.Identities[left].Email < result.Identities[right].Email
		}
		return result.Identities[left].EngineIdentityID < result.Identities[right].EngineIdentityID
	})
	return result, nil
}

func requireStalwartIdentityReconcileEOF(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return stalwartEngineError(EngineErrorUnavailable)
	}
	return nil
}

var _ MailboxIdentityReconcileProvider = (*StalwartEngine)(nil)
