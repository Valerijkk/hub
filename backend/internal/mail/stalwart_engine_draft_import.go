package mail

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartDraftEmailImportCallID     = "draft-email-import"
	stalwartDraftEmailImportCreationID = "draft-import"
)

// StalwartDraftMIMEImportRequest binds one already uploaded MIME blob to the
// exact initial-draft mutation that owns it. ReceivedAt is durable application
// state and is the only import timestamp accepted by RFC 8621; createdAt is a
// provider-derived Email property and is therefore intentionally not supplied.
type StalwartDraftMIMEImportRequest struct {
	AccountID       string
	DraftsMailboxID string
	BlobID          string
	SizeBytes       int64
	IfInState       string
	DraftKey        string
	MutationID      string
	ReceivedAt      time.Time
}

// StalwartDraftMIMEImportResult contains only the opaque identifiers and
// state transition confirmed by one successful Email/import response.
type StalwartDraftMIMEImportResult struct {
	AccountID string
	MessageID string
	BlobID    string
	ThreadID  string
	SizeBytes int64
	OldState  string
	NewState  string
}

type stalwartEmailImportArguments struct {
	AccountID string                              `json:"accountId"`
	IfInState string                              `json:"ifInState"`
	Emails    map[string]stalwartEmailImportEmail `json:"emails"`
}

type stalwartEmailImportEmail struct {
	BlobID     string          `json:"blobId"`
	MailboxIDs map[string]bool `json:"mailboxIds"`
	Keywords   map[string]bool `json:"keywords"`
	ReceivedAt time.Time       `json:"receivedAt"`
}

type stalwartEmailImportResponse struct {
	AccountID  string                              `json:"accountId"`
	OldState   json.RawMessage                     `json:"oldState"`
	NewState   string                              `json:"newState"`
	Created    map[string]json.RawMessage          `json:"created"`
	NotCreated map[string]stalwartEmailImportError `json:"notCreated"`
}

type stalwartEmailImportError struct {
	Type       string   `json:"type"`
	NotFound   []string `json:"notFound"`
	ExistingID string   `json:"existingId"`
}

// ImportDraftMIME performs exactly one RFC 8621 Email/import for a MIME blob
// that was already uploaded for this account. It never retries after dispatch;
// callers must reconcile an indeterminate result with the exact durable
// draft/mutation markers before any subsequent mutation.
func (engine *StalwartEngine) ImportDraftMIME(ctx context.Context, request StalwartDraftMIMEImportRequest) (StalwartDraftMIMEImportResult, error) {
	result, err := engine.stalwartImportDraftMIME(ctx, request)
	if errors.Is(err, errStalwartDraftMIMEImportExactBlobNotFound) {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorRestartRequired)
	}
	return result, err
}

// stalwartImportDraftMIME keeps the response-proof path private so
// ApplyDraftMIMEImport can distinguish it from provider-wide restart errors.
func (engine *StalwartEngine) stalwartImportDraftMIME(ctx context.Context, request StalwartDraftMIMEImportRequest) (StalwartDraftMIMEImportResult, error) {
	if !stalwartValidDraftMIMEImportRequest(ctx, request) {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", request.AccountID, request.DraftKey)
	if err != nil {
		return StalwartDraftMIMEImportResult{}, err
	}
	mutationMarker, err := stalwartDraftMutationMarkerV2(request.AccountID, request.DraftKey, request.MutationID)
	if err != nil {
		return StalwartDraftMIMEImportResult{}, err
	}
	payload, err := json.Marshal(stalwartEmailImportArguments{
		AccountID: request.AccountID,
		IfInState: request.IfInState,
		Emails: map[string]stalwartEmailImportEmail{
			stalwartDraftEmailImportCreationID: {
				BlobID:     request.BlobID,
				MailboxIDs: map[string]bool{request.DraftsMailboxID: true},
				Keywords:   map[string]bool{"$draft": true, draftMarker: true, mutationMarker: true},
				ReceivedAt: request.ReceivedAt,
			},
		},
	})
	if err != nil {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.stalwartCallBoundedMutation(ctx, request.AccountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/import",
			Arguments: payload,
			CallID:    stalwartDraftEmailImportCallID,
		}},
	})
	if err != nil {
		return StalwartDraftMIMEImportResult{}, err
	}
	return stalwartValidateDraftMIMEImportResponse(response, request)
}

func stalwartValidDraftMIMEImportRequest(ctx context.Context, request StalwartDraftMIMEImportRequest) bool {
	return ctx != nil &&
		validAccountID(request.AccountID) &&
		validOpaqueID(request.DraftsMailboxID) &&
		validOpaqueID(request.BlobID) &&
		request.SizeBytes >= 1 &&
		request.SizeBytes <= MaxEngineMessageBytes &&
		validOpaqueID(request.IfInState) &&
		validSafeText(request.DraftKey, MaxEngineOpaqueIDBytes, true) &&
		validSafeText(request.MutationID, MaxEngineOpaqueIDBytes, true) &&
		stalwartValidImportReceivedAt(request.ReceivedAt)
}

func stalwartValidImportReceivedAt(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond() == 0
}

func stalwartValidateDraftMIMEImportResponse(response stalwart.JMAPResponse, request StalwartDraftMIMEImportRequest) (StalwartDraftMIMEImportResult, error) {
	if len(response.MethodResponses) != 1 {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	method := response.MethodResponses[0]
	if method.Name != "Email/import" || method.CallID != stalwartDraftEmailImportCallID {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	var payload stalwartEmailImportResponse
	if err := json.Unmarshal(method.Arguments, &payload); err != nil {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if payload.AccountID != request.AccountID {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(payload.NewState) {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	oldState, oldStatePresent, oldStateKnown, err := stalwartDecodeSetOldState(payload.OldState, request.IfInState)
	if err != nil || !oldStatePresent || !oldStateKnown {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if len(payload.NotCreated) != 0 {
		return StalwartDraftMIMEImportResult{}, stalwartDraftMIMEImportFailure(payload, oldState, request.BlobID)
	}
	if payload.NewState == oldState || len(payload.Created) != 1 {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	createdRaw, exists := payload.Created[stalwartDraftEmailImportCreationID]
	if !exists {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	var created stalwartCreatedEmail
	if err := json.Unmarshal(createdRaw, &created); err != nil ||
		!validOpaqueID(created.ID) ||
		created.BlobID != request.BlobID ||
		!validOpaqueID(created.ThreadID) ||
		created.Size != request.SizeBytes {
		return StalwartDraftMIMEImportResult{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return StalwartDraftMIMEImportResult{
		AccountID: payload.AccountID,
		MessageID: created.ID,
		BlobID:    created.BlobID,
		ThreadID:  created.ThreadID,
		SizeBytes: created.Size,
		OldState:  oldState,
		NewState:  payload.NewState,
	}, nil
}

func stalwartDraftMIMEImportFailure(payload stalwartEmailImportResponse, oldState, expectedBlobID string) error {
	if len(payload.Created) != 0 || len(payload.NotCreated) != 1 || payload.NewState != oldState {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	failure, exists := payload.NotCreated[stalwartDraftEmailImportCreationID]
	if !exists || failure.Type == "" {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	if failure.Type == "blobNotFound" {
		if len(failure.NotFound) != 1 || !validOpaqueID(failure.NotFound[0]) || failure.NotFound[0] != expectedBlobID {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		return errStalwartDraftMIMEImportExactBlobNotFound
	}
	// RFC 8621 does not make alreadyExists proof that this draft/mutation was
	// applied. Only a separate exact-marker reconciliation can establish that.
	if failure.Type == "alreadyExists" {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	switch failure.Type {
	case "stateMismatch":
		return stalwartEngineError(EngineErrorConflict)
	case "notFound":
		return stalwartEngineError(EngineErrorNotFound)
	case "accountNotFound", "forbidden", "accountNotSupported":
		return stalwartEngineError(EngineErrorUnauthorizedAccount)
	case "tooLarge", "requestTooLarge":
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	case "invalidArguments", "invalidProperties", "invalidPatch", "willDestroy", "invalidResultReference", "unknownMethod", "unknownCapability":
		return stalwartEngineError(EngineErrorRejectedPolicy)
	default:
		return stalwartEngineError(EngineErrorIndeterminate)
	}
}
