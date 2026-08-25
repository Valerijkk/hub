package mail

import (
	"context"
	"encoding/json"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

// PreflightSubmission reads an exact draft and confirms that its canonical
// sender remains available as one exact provider identity. It never creates a
// provider submission or changes provider state.
func (engine *StalwartEngine) PreflightSubmission(ctx context.Context, request EngineSubmissionPreflightRequest) (EngineSubmissionPreflight, error) {
	submission := request.Submission
	if ctx == nil || validateSubmission(submission) != nil {
		return EngineSubmissionPreflight{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	_, effectiveMaxMessageBytes, err := preflightSubmissionPolicy(request)
	if err != nil {
		return EngineSubmissionPreflight{}, err
	}
	message, err := engine.GetDraft(ctx, EngineDraftReadRequest{
		AccountID: submission.AccountID, DraftKey: submission.DraftKey,
		MessageID: submission.MessageID, ExpectedVersion: submission.ExpectedVersion,
		MaxContentBytes: effectiveMaxMessageBytes,
	})
	if err != nil {
		return EngineSubmissionPreflight{}, err
	}
	preflight, err := preflightSubmissionForMessage(request, message)
	if err != nil {
		return EngineSubmissionPreflight{}, err
	}
	if err := engine.stalwartPreflightSubmissionIdentity(ctx, submission.AccountID, preflight.FromEmail); err != nil {
		return EngineSubmissionPreflight{}, err
	}
	return preflight, nil
}

func (engine *StalwartEngine) stalwartPreflightSubmissionIdentity(ctx context.Context, accountID, fromEmail string) error {
	identityPayload, err := json.Marshal(stalwartIdentityGetArguments{AccountID: accountID, Properties: []string{"id", "email"}})
	if err != nil {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using:       stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "Identity/get", Arguments: identityPayload, CallID: stalwartIdentityGetCallID}},
	}, false)
	if err != nil {
		return err
	}
	raw, err := stalwartDraftMethodPayload(response, "Identity/get", stalwartIdentityGetCallID)
	if err != nil {
		return err
	}
	var identities stalwartIdentityGetResponse
	if err := json.Unmarshal(raw, &identities); err != nil {
		return stalwartEngineError(EngineErrorUnavailable)
	}
	if identities.AccountID != accountID {
		return stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(identities.State) || len(identities.List) > stalwartMaximumIdentityItems || len(identities.NotFound) != 0 {
		return stalwartEngineError(EngineErrorUnavailable)
	}
	matchingIdentities := 0
	for _, identity := range identities.List {
		if !validOpaqueID(identity.ID) || !validSafeText(identity.Email, MaxEngineAddressEmailBytes, true) {
			return stalwartEngineError(EngineErrorUnavailable)
		}
		canonical, normalizeErr := NormalizeAddress(EngineAddress{Email: identity.Email})
		if normalizeErr != nil {
			return stalwartEngineError(EngineErrorUnavailable)
		}
		if canonical.Email != fromEmail {
			continue
		}
		matchingIdentities++
		if matchingIdentities > 1 {
			return stalwartEngineError(EngineErrorIdentityUnavailable)
		}
	}
	if matchingIdentities != 1 {
		return stalwartEngineError(EngineErrorIdentityUnavailable)
	}
	return nil
}
