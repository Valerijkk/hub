package mail

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"errors"
	"strings"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartDraftEmailSetCallID       = "draft-email-set"
	stalwartMarkerEmailQueryCallID    = "marker-email-query"
	stalwartDraftEmailGetCallID       = "draft-email-get"
	stalwartIdentityGetCallID         = "identity-get"
	stalwartSubmissionStateCallID     = "submission-state"
	stalwartSubmissionSetCallID       = "submission-set"
	stalwartSubmissionQueryCallID     = "submission-query"
	stalwartDraftCreationID           = "draft"
	stalwartSubmissionCreationID      = "submission"
	stalwartMaximumIdentityItems      = 100
	stalwartMarkerQueryLimit           = 2
	stalwartDraftTextPartID            = "text"
	stalwartDraftHTMLPartID            = "html"
	stalwartSubmissionBindingVersion   = "v1"
)

var stalwartDraftDetailProperties = append(append([]string(nil), stalwartEmailDetailProperties...), "bcc", "header:Importance")

type stalwartDraftBodyPart struct {
	PartID string `json:"partId"`
	Type   string `json:"type"`
}

type stalwartDraftBodyValue struct {
	Value string `json:"value"`
}

type stalwartDraftEmailCreate struct {
	MailboxIDs  map[string]bool                  `json:"mailboxIds"`
	Keywords    map[string]bool                  `json:"keywords"`
	From        []stalwartEmailAddress           `json:"from"`
	To          []stalwartEmailAddress           `json:"to,omitempty"`
	CC          []stalwartEmailAddress           `json:"cc,omitempty"`
	BCC         []stalwartEmailAddress           `json:"bcc,omitempty"`
	ReplyTo     []stalwartEmailAddress           `json:"replyTo,omitempty"`
	Subject     string                            `json:"subject"`
	InReplyTo   []string                          `json:"inReplyTo,omitempty"`
	References  []string                          `json:"references,omitempty"`
	Importance  string                            `json:"header:Importance,omitempty"`
	TextBody    []stalwartDraftBodyPart           `json:"textBody,omitempty"`
	HTMLBody    []stalwartDraftBodyPart           `json:"htmlBody,omitempty"`
	BodyValues  map[string]stalwartDraftBodyValue `json:"bodyValues"`
}

type stalwartDraftEmailSetArguments struct {
	AccountID string                                `json:"accountId"`
	IfInState string                                `json:"ifInState,omitempty"`
	Create    map[string]stalwartDraftEmailCreate   `json:"create"`
	Destroy   []string                              `json:"destroy,omitempty"`
}

type stalwartCreatedEmail struct {
	ID       string `json:"id"`
	BlobID   string `json:"blobId"`
	ThreadID string `json:"threadId"`
	Size     int64  `json:"size"`
}

type stalwartMarkerEmailQueryArguments struct {
	AccountID      string                         `json:"accountId"`
	CalculateTotal bool                           `json:"calculateTotal"`
	Filter         stalwartMarkerEmailQueryFilter `json:"filter"`
	Limit          int                            `json:"limit"`
	Position       int                            `json:"position"`
}

type stalwartMarkerEmailQueryFilter struct {
	HasKeyword string `json:"hasKeyword"`
}

type stalwartIdentityGetArguments struct {
	AccountID string   `json:"accountId"`
	Properties []string `json:"properties"`
}

type stalwartIdentityRecord struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type stalwartIdentityGetResponse struct {
	AccountID string                    `json:"accountId"`
	State     string                    `json:"state"`
	List      []stalwartIdentityRecord  `json:"list"`
	NotFound  []string                  `json:"notFound"`
}

type stalwartSubmissionGetArguments struct {
	AccountID  string   `json:"accountId"`
	IDs        []string `json:"ids"`
	Properties []string `json:"properties"`
}

type stalwartSubmissionGetResponse struct {
	AccountID string            `json:"accountId"`
	State     string            `json:"state"`
	List      []json.RawMessage `json:"list"`
	NotFound  []string          `json:"notFound"`
}

type stalwartSubmissionAddress struct {
	Email      string            `json:"email"`
	Parameters map[string]string `json:"parameters"`
}

type stalwartSubmissionEnvelope struct {
	MailFrom stalwartSubmissionAddress   `json:"mailFrom"`
	RcptTo   []stalwartSubmissionAddress `json:"rcptTo"`
}

type stalwartSubmissionCreate struct {
	IdentityID string                      `json:"identityId"`
	EmailID    string                      `json:"emailId"`
	Envelope   *stalwartSubmissionEnvelope `json:"envelope"`
}

type stalwartSubmissionSetArguments struct {
	AccountID             string                               `json:"accountId"`
	IfInState             string                               `json:"ifInState"`
	Create                map[string]stalwartSubmissionCreate  `json:"create"`
	OnSuccessUpdateEmail  map[string]map[string]any            `json:"onSuccessUpdateEmail"`
}

type stalwartCreatedSubmission struct {
	ID string `json:"id"`
}

type stalwartSubmissionQueryArguments struct {
	AccountID      string                         `json:"accountId"`
	CalculateTotal bool                           `json:"calculateTotal"`
	Filter         stalwartSubmissionQueryFilter  `json:"filter"`
	Limit          int                            `json:"limit"`
	Position       int                            `json:"position"`
}

type stalwartSubmissionQueryFilter struct {
	EmailIDs []string `json:"emailIds"`
}

type stalwartSubmissionQueryResponse struct {
	AccountID  string   `json:"accountId"`
	IDs        []string `json:"ids"`
	Position   int      `json:"position"`
	QueryState string   `json:"queryState"`
	Total      *int     `json:"total"`
}

// stalwartSubmissionBinding is the versioned, canonical preimage for the
// opaque binding marker. It excludes mutable provider state such as the
// Email/set version and folder transition, while retaining the exact request
// identity, immutable content, and send-relevant metadata.
type stalwartSubmissionBinding struct {
	Version         string                                `json:"version"`
	AccountID       string                                `json:"account_id"`
	DraftKey        string                                `json:"draft_key"`
	SubmissionKey   string                                `json:"submission_key"`
	MessageID       string                                `json:"message_id"`
	ExpectedVersion string                                `json:"expected_version"`
	InternetMessageID string                              `json:"internet_message_id"`
	From            stalwartSubmissionBindingAddress     `json:"from"`
	To              []stalwartSubmissionBindingAddress   `json:"to"`
	CC              []stalwartSubmissionBindingAddress   `json:"cc"`
	BCC             []stalwartSubmissionBindingAddress   `json:"bcc"`
	ReplyTo         []stalwartSubmissionBindingAddress   `json:"reply_to"`
	Subject         string                                `json:"subject"`
	InReplyTo       string                                `json:"in_reply_to"`
	References      []string                              `json:"references"`
	Priority        EngineMessagePriority                 `json:"priority"`
	Body            stalwartSubmissionBindingBody        `json:"body"`
	Attachments     []stalwartSubmissionBindingAttachment `json:"attachments"`
}

type stalwartSubmissionBindingAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type stalwartSubmissionBindingBody struct {
	Format    EngineBodyFormat `json:"format"`
	PlainBody string           `json:"plain_body"`
	RichBody  string           `json:"rich_body"`
}

type stalwartSubmissionBindingAttachment struct {
	BlobID      string                      `json:"blob_id"`
	ContentID   string                      `json:"content_id"`
	Filename    string                      `json:"filename"`
	MediaType   string                      `json:"media_type"`
	SizeBytes   int64                       `json:"size_bytes"`
	Disposition EngineAttachmentDisposition `json:"disposition"`
	SHA256      string                      `json:"sha256"`
}

// SaveDraft creates an RFC 8621 draft or atomically replaces the previous
// immutable Email with a new draft and destroys the old Email in one set call.
func (engine *StalwartEngine) SaveDraft(ctx context.Context, draft EngineDraft) (EngineMessage, error) {
	if ctx == nil || validateDraftRequest(draft) != nil || !stalwartValidDraftThreading(draft.Message) {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if len(draft.Message.Attachments) != 0 {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if draftMinimumBytes(draft) > draft.MaxContentBytes {
		return EngineMessage{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", draft.AccountID, draft.DraftKey)
	if err != nil {
		return EngineMessage{}, err
	}
	mutationMarker, err := stalwartDraftMutationMarkerV2(draft.AccountID, draft.DraftKey, draft.MutationID)
	if err != nil {
		return EngineMessage{}, err
	}
	expectedState := draft.PreviousVersion
	if draft.MessageID == "" {
		expectedState, err = engine.stalwartReadEmailState(ctx, draft.AccountID)
		if err != nil {
			return EngineMessage{}, err
		}
	}
	folders, _, err := engine.listFolderSnapshot(ctx, draft.AccountID)
	if err != nil {
		return EngineMessage{}, err
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return EngineMessage{}, err
	}
	existing, found, ambiguous, err := engine.stalwartLookupDraft(ctx, draft.AccountID, draftsID, draftMarker, draft.MaxContentBytes)
	if err != nil {
		return EngineMessage{}, err
	}
	if ambiguous {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if found && stalwartExactDraftOperation(existing, draft.AccountID, draftsID, draftMarker, mutationMarker) {
		return existing, nil
	}
	// A prepared or fenced submission adds reserved markers to the immutable
	// draft. Do not let a concurrent SaveDraft replace that Email between the
	// content binding check and dispatch.
	if found && len(existing.Projection.ReservedKeywords) != 3 {
		return EngineMessage{}, stalwartEngineError(EngineErrorConflict)
	}
	if draft.MessageID == "" {
		if found {
			return EngineMessage{}, stalwartEngineError(EngineErrorConflict)
		}
	} else {
		if !found || existing.Projection.MessageID != draft.MessageID {
			return EngineMessage{}, stalwartEngineError(EngineErrorNotFound)
		}
		if existing.Projection.Version != draft.PreviousVersion {
			return EngineMessage{}, stalwartEngineError(EngineErrorConflict)
		}
	}
	created := stalwartDraftCreate(draft.Message, draftsID, draftMarker, mutationMarker)
	arguments := stalwartDraftEmailSetArguments{
		AccountID: draft.AccountID,
		IfInState: expectedState,
		Create: map[string]stalwartDraftEmailCreate{stalwartDraftCreationID: created},
	}
	if draft.MessageID != "" {
		arguments.Destroy = []string{draft.MessageID}
	}
	response, err := engine.stalwartSaveDraftSet(ctx, draft.AccountID, arguments)
	if err != nil {
		if draft.MessageID == "" && stalwartErrorHasKind(err, EngineErrorConflict) {
			return engine.stalwartResolveInitialDraftConflict(ctx, draft, err)
		}
		return EngineMessage{}, err
	}
	createdID, _, err := stalwartValidateDraftSetResponse(response, draft, expectedState)
	if err != nil {
		if draft.MessageID == "" && stalwartErrorHasKind(err, EngineErrorConflict) {
			return engine.stalwartResolveInitialDraftConflict(ctx, draft, err)
		}
		return EngineMessage{}, err
	}
	message, err := engine.stalwartReadDraftMessage(ctx, draft.AccountID, createdID, draft.MaxContentBytes)
	if err != nil {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if !stalwartExactDraftOperation(message, draft.AccountID, draftsID, draftMarker, mutationMarker) || draft.MessageID != "" && createdID == draft.MessageID {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return message, nil
}

// ReconcileDraft locates the exact opaque operation marker. A different
// operation for the same DraftKey, duplicate marker, or malformed projection
// is an indeterminate outcome and never authorises a blind retry.
func (engine *StalwartEngine) ReconcileDraft(ctx context.Context, request EngineDraftReconciliation) (EngineDraftResult, error) {
	if ctx == nil || validateDraftReconciliation(request) != nil {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", request.AccountID, request.DraftKey)
	if err != nil {
		return nil, err
	}
	mutationMarker, err := stalwartDraftMutationMarkerV2(request.AccountID, request.DraftKey, request.MutationID)
	if err != nil {
		return nil, err
	}
	folders, _, err := engine.listFolderSnapshot(ctx, request.AccountID)
	if err != nil {
		return stalwartDraftReconciliationError(err)
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return stalwartDraftReconciliationError(err)
	}
	message, found, ambiguous, err := engine.stalwartLookupDraft(ctx, request.AccountID, draftsID, draftMarker, request.MaxContentBytes)
	if err != nil {
		return stalwartDraftReconciliationError(err)
	}
	if ambiguous {
		return EngineDraftIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	if !found && request.MessageID == "" {
		return EngineDraftAbsent{}, nil
	}
	if !found {
		return EngineDraftIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	if stalwartExactDraftOperation(message, request.AccountID, draftsID, draftMarker, mutationMarker) &&
		(request.MessageID == "" ||
			message.Projection.MessageID != request.MessageID ||
			message.Projection.Version != request.PreviousVersion) {
		return EngineDraftApplied{Message: message}, nil
	}
	if request.MessageID == "" || request.ConfirmedMutationID == "" {
		return EngineDraftIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	confirmedMarker, err := stalwartDraftMutationMarkerV2(request.AccountID, request.DraftKey, request.ConfirmedMutationID)
	if err != nil {
		return nil, err
	}
	if stalwartExactDraftOperation(message, request.AccountID, draftsID, draftMarker, confirmedMarker) && message.Projection.MessageID == request.MessageID && message.Projection.Version == request.PreviousVersion {
		return EngineDraftRetained{Message: message}, nil
	}
	return EngineDraftIndeterminate{Code: EngineIndeterminateUnavailable}, nil
}

// DeleteDraft destroys exactly one versioned owner draft without adding a
// deletion marker. A transport result after dispatch is never retried.
func (engine *StalwartEngine) DeleteDraft(ctx context.Context, request EngineDraftDeletion) error {
	if ctx == nil || validateDraftDeletion(request) != nil {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", request.AccountID, request.DraftKey)
	if err != nil {
		return err
	}
	folders, _, err := engine.listFolderSnapshot(ctx, request.AccountID)
	if err != nil {
		return err
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return err
	}
	message, found, ambiguous, err := engine.stalwartLookupDraft(ctx, request.AccountID, draftsID, draftMarker, MaxEngineMessageBytes)
	if err != nil {
		return err
	}
	if ambiguous {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	if !found {
		return stalwartEngineError(EngineErrorNotFound)
	}
	if message.Projection.MessageID != request.MessageID || message.Projection.Version != request.ExpectedVersion {
		return stalwartEngineError(EngineErrorConflict)
	}
	if !stalwartExactDraftDeletion(message, request, draftsID, draftMarker) {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	response, err := engine.stalwartBoundedEmailSet(ctx, request.AccountID, stalwartEmailSetArguments{
		AccountID: request.AccountID,
		IfInState: request.ExpectedVersion,
		Destroy:   []string{request.MessageID},
	})
	if err != nil {
		return err
	}
	set, err := parseStalwartEmailSetResponse(response, request.AccountID, request.ExpectedVersion)
	if err != nil {
		return err
	}
	return stalwartEmailOnlyDestroySucceeded(set, []string{request.MessageID})
}

// ReconcileDraftDeletion observes only the exact stable draft marker. Provider
// absence proves the desired draft absence, not the actor that caused it.
func (engine *StalwartEngine) ReconcileDraftDeletion(ctx context.Context, request EngineDraftDeletion) (EngineDraftDeletionResult, error) {
	if ctx == nil || validateDraftDeletion(request) != nil {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", request.AccountID, request.DraftKey)
	if err != nil {
		return nil, err
	}
	folders, _, err := engine.listFolderSnapshot(ctx, request.AccountID)
	if err != nil {
		return stalwartDraftDeletionReconciliationError(err)
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return stalwartDraftDeletionReconciliationError(err)
	}
	message, found, ambiguous, err := engine.stalwartLookupDraft(ctx, request.AccountID, draftsID, draftMarker, MaxEngineMessageBytes)
	if err != nil {
		return stalwartDraftDeletionReconciliationError(err)
	}
	if ambiguous || found && !stalwartExactDraftDeletion(message, request, draftsID, draftMarker) {
		return EngineDraftDeletionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	if !found {
		return EngineDraftDeleted{}, nil
	}
	return EngineDraftRetainedUnchanged{}, nil
}

// GetDraft reads only the exact owner draft selected by its opaque marker.
// It never accepts actor or organization scope because the service has already
// resolved the engine account before calling this provider adapter.
func (engine *StalwartEngine) GetDraft(ctx context.Context, request EngineDraftReadRequest) (EngineMessage, error) {
	if ctx == nil || !validAccountID(request.AccountID) || !validSafeText(request.DraftKey, MaxEngineOpaqueIDBytes, true) || !validOpaqueID(request.MessageID) || !validOpaqueID(request.ExpectedVersion) || request.MaxContentBytes < 1 || request.MaxContentBytes > MaxEngineMessageBytes {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", request.AccountID, request.DraftKey)
	if err != nil {
		return EngineMessage{}, err
	}
	folders, _, err := engine.listFolderSnapshot(ctx, request.AccountID)
	if err != nil {
		return EngineMessage{}, err
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return EngineMessage{}, err
	}
	message, found, ambiguous, err := engine.stalwartLookupDraft(ctx, request.AccountID, draftsID, draftMarker, request.MaxContentBytes)
	if err != nil {
		return EngineMessage{}, err
	}
	if ambiguous {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if !found {
		return EngineMessage{}, stalwartEngineError(EngineErrorNotFound)
	}
	if message.Projection.AccountID != request.AccountID || !message.Projection.Flags.Draft || !stalwartExactFolder(message.Projection, draftsID) || !stalwartProjectionHasKeyword(message.Projection, draftMarker) {
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if message.Projection.MessageID != request.MessageID || message.Projection.Version != request.ExpectedVersion {
		return EngineMessage{}, stalwartEngineError(EngineErrorConflict)
	}
	if int64(messageVariableBytes(message)) > request.MaxContentBytes {
		return EngineMessage{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	if err := validateEngineMessage(message); err != nil {
		return EngineMessage{}, stalwartMappingResponseError(err)
	}
	return message, nil
}

func (engine *StalwartEngine) stalwartResolveInitialDraftConflict(ctx context.Context, draft EngineDraft, conflict error) (EngineMessage, error) {
	result, err := engine.ReconcileDraft(ctx, EngineDraftReconciliation{
		AccountID: draft.AccountID, DraftKey: draft.DraftKey, MutationID: draft.MutationID,
		MaxContentBytes: draft.MaxContentBytes,
	})
	if err != nil {
		return EngineMessage{}, err
	}
	switch reconciled := result.(type) {
	case EngineDraftApplied:
		return reconciled.Message, nil
	case EngineDraftAbsent:
		return EngineMessage{}, conflict
	case EngineDraftIndeterminate:
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	default:
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
}

func stalwartErrorHasKind(err error, kind EngineErrorKind) bool {
	var engineErr *EngineError
	return errors.As(err, &engineErr) && engineErr != nil && engineErr.Kind == kind
}

// Submit prepares an exact locator/binding pair, then records a separate
// dispatch fence before creating one EmailSubmission. Only a confirmed fence
// response authorises dispatch; an uncertain fence is never retried.
func (engine *StalwartEngine) Submit(ctx context.Context, submission EngineSubmission) (EngineSubmissionResult, error) {
	if ctx == nil || validateSubmission(submission) != nil {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", submission.AccountID, submission.DraftKey)
	if err != nil {
		return nil, err
	}
	submissionMarker, err := stalwartOpaqueMarker("submission", submission.AccountID, submission.SubmissionKey)
	if err != nil {
		return nil, err
	}
	dispatchMarker, err := stalwartOpaqueMarker("submission-dispatch", submission.AccountID, submission.SubmissionKey)
	if err != nil {
		return nil, err
	}
	folders, _, err := engine.listFolderSnapshot(ctx, submission.AccountID)
	if err != nil {
		return nil, err
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return nil, err
	}
	sentID, err := stalwartExactFolderRole(folders, EngineFolderRoleSent)
	if err != nil {
		return nil, err
	}

	locatorIDs, locatorAmbiguous, err := engine.stalwartQueryEmailsByMarker(ctx, submission.AccountID, submissionMarker)
	if err != nil {
		return nil, err
	}
	if locatorAmbiguous || len(locatorIDs) > 1 {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}
	if len(locatorIDs) == 1 && locatorIDs[0] != submission.MessageID {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}
	dispatchIDs, dispatchAmbiguous, err := engine.stalwartQueryEmailsByMarker(ctx, submission.AccountID, dispatchMarker)
	if err != nil {
		return nil, err
	}
	if dispatchAmbiguous || len(dispatchIDs) > 1 || len(dispatchIDs) == 1 && (len(locatorIDs) != 1 || dispatchIDs[0] != locatorIDs[0]) {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}

	var message EngineMessage
	preparedState := ""
	needsPrepare := false
	alreadyFenced := false
	if len(locatorIDs) == 1 {
		message, err = engine.stalwartReadDraftMessage(ctx, submission.AccountID, locatorIDs[0], submission.MaxMessageBytes)
		if err != nil {
			return nil, err
		}
		preparedState = message.Projection.Version
		alreadyFenced = len(dispatchIDs) == 1
	} else {
		if len(dispatchIDs) != 0 {
			return nil, stalwartEngineError(EngineErrorIndeterminate)
		}
		var found, ambiguous bool
		message, found, ambiguous, err = engine.stalwartLookupDraft(ctx, submission.AccountID, draftsID, draftMarker, submission.MaxMessageBytes)
		if err != nil {
			return nil, err
		}
		if ambiguous {
			return nil, stalwartEngineError(EngineErrorIndeterminate)
		}
		if !found || message.Projection.MessageID != submission.MessageID {
			return nil, stalwartEngineError(EngineErrorNotFound)
		}
		if message.Projection.Version != submission.ExpectedVersion {
			return nil, stalwartEngineError(EngineErrorConflict)
		}
		needsPrepare = true
	}
	if err := stalwartValidateSubmissionMessage(message, submission); err != nil {
		return nil, err
	}
	bindingMarker, err := stalwartSubmissionBindingMarker(submission, message)
	if err != nil {
		return nil, err
	}
	if alreadyFenced {
		reconciled, reconcileErr := engine.stalwartReconcileBoundSubmission(ctx, submission, message, draftsID, sentID, draftMarker, submissionMarker, bindingMarker, dispatchMarker)
		if reconcileErr != nil {
			return nil, reconcileErr
		}
		if accepted, ok := reconciled.(EngineSubmissionAccepted); ok {
			return accepted, nil
		}
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}
	if needsPrepare {
		if len(message.Projection.ReservedKeywords) != 3 ||
			stalwartProjectionHasKeyword(message.Projection, submissionMarker) ||
			stalwartProjectionHasKeyword(message.Projection, bindingMarker) ||
			stalwartProjectionHasKeyword(message.Projection, dispatchMarker) {
			return nil, stalwartEngineError(EngineErrorIndeterminate)
		}
	} else if !stalwartExactPreparedSubmission(message, submission, draftsID, draftMarker, submissionMarker, bindingMarker, dispatchMarker) {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}

	identityID, submissionState, err := engine.stalwartResolveSubmissionContext(ctx, submission.AccountID, message.Projection.From.Email)
	if err != nil {
		return nil, err
	}
	envelope := stalwartSubmissionEnvelope{
		MailFrom: stalwartSubmissionAddress{Email: message.Projection.From.Email},
		RcptTo:   stalwartSubmissionRecipients(message.Projection),
	}
	arguments := stalwartSubmissionSetArguments{
		AccountID: submission.AccountID,
		IfInState: submissionState,
		Create: map[string]stalwartSubmissionCreate{
			stalwartSubmissionCreationID: {IdentityID: identityID, EmailID: submission.MessageID, Envelope: &envelope},
		},
		OnSuccessUpdateEmail: map[string]map[string]any{
			"#" + stalwartSubmissionCreationID: {
				"mailboxIds/" + draftsID: nil,
				"mailboxIds/" + sentID:   true,
				"keywords/$draft":        nil,
			},
		},
	}
	request, err := stalwartSubmissionSetRequest(arguments)
	if err != nil {
		return nil, err
	}
	if err := stalwartEnsureRequestFitsTransport(request); err != nil {
		return nil, err
	}

	if needsPrepare {
		prepareResponse, prepareErr := engine.stalwartBoundedEmailSet(ctx, submission.AccountID, stalwartEmailSetArguments{
			AccountID: submission.AccountID,
			IfInState: submission.ExpectedVersion,
			Update: map[string]map[string]any{
				submission.MessageID: {
					"keywords/" + submissionMarker: true,
					"keywords/" + bindingMarker:    true,
				},
			},
		})
		if prepareErr != nil {
			return nil, prepareErr
		}
		prepareSet, prepareErr := parseStalwartEmailSetResponse(prepareResponse, submission.AccountID, submission.ExpectedVersion)
		if prepareErr != nil {
			return nil, prepareErr
		}
		if prepareErr = stalwartSubmissionMarkerUpdateSucceeded(prepareSet, submission.MessageID); prepareErr != nil {
			return nil, prepareErr
		}
		preparedState = prepareSet.NewState
	}
	fenceResponse, err := engine.stalwartBoundedEmailSet(ctx, submission.AccountID, stalwartEmailSetArguments{
		AccountID: submission.AccountID,
		IfInState: preparedState,
		Update: map[string]map[string]any{
			submission.MessageID: {"keywords/" + dispatchMarker: true},
		},
	})
	if err != nil {
		return nil, err
	}
	fenceSet, err := parseStalwartEmailSetResponse(fenceResponse, submission.AccountID, preparedState)
	if err != nil {
		return nil, err
	}
	if err := stalwartSubmissionMarkerUpdateSucceeded(fenceSet, submission.MessageID); err != nil {
		return nil, err
	}
	response, err := engine.stalwartSubmitSet(ctx, submission.AccountID, request)
	if err != nil {
		return nil, err
	}
	submissionID, err := stalwartValidateSubmissionSetResponse(response, submission, submissionState, fenceSet.NewState)
	if err != nil {
		return nil, err
	}
	verified, verificationErr := engine.stalwartReadDraftMessage(ctx, submission.AccountID, submission.MessageID, submission.MaxMessageBytes)
	if verificationErr != nil {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}
	verifiedBinding, verificationErr := stalwartSubmissionBindingMarker(submission, verified)
	if verificationErr != nil || verifiedBinding != bindingMarker ||
		!stalwartExactDispatchedSubmission(verified, submission, draftsID, sentID, draftMarker, submissionMarker, bindingMarker, dispatchMarker) {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}
	return EngineSubmissionAccepted{SubmissionID: submissionID, MessageID: submission.MessageID}, nil
}

// ReconcileSubmission reads the exact locator, content binding, dispatch fence,
// sent message and EmailSubmission identity for one fully bound request. It
// never mutates provider state.
func (engine *StalwartEngine) ReconcileSubmission(ctx context.Context, submission EngineSubmission) (EngineSubmissionResult, error) {
	if ctx == nil || validateSubmission(submission) != nil {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", submission.AccountID, submission.DraftKey)
	if err != nil {
		return nil, err
	}
	locatorMarker, err := stalwartOpaqueMarker("submission", submission.AccountID, submission.SubmissionKey)
	if err != nil {
		return nil, err
	}
	dispatchMarker, err := stalwartOpaqueMarker("submission-dispatch", submission.AccountID, submission.SubmissionKey)
	if err != nil {
		return nil, err
	}
	locatorIDs, locatorAmbiguous, err := engine.stalwartQueryEmailsByMarker(ctx, submission.AccountID, locatorMarker)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	dispatchIDs, dispatchAmbiguous, err := engine.stalwartQueryEmailsByMarker(ctx, submission.AccountID, dispatchMarker)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	if !locatorAmbiguous && len(locatorIDs) == 0 && !dispatchAmbiguous && len(dispatchIDs) == 0 {
		return EngineSubmissionAbsent{}, nil
	}
	if locatorAmbiguous || dispatchAmbiguous || len(locatorIDs) != 1 || locatorIDs[0] != submission.MessageID || len(dispatchIDs) > 1 || len(dispatchIDs) == 1 && dispatchIDs[0] != submission.MessageID {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	folders, _, err := engine.listFolderSnapshot(ctx, submission.AccountID)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	sentID, err := stalwartExactFolderRole(folders, EngineFolderRoleSent)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	message, err := engine.stalwartReadDraftMessage(ctx, submission.AccountID, submission.MessageID, submission.MaxMessageBytes)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	if err := stalwartValidateSubmissionMessage(message, submission); err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	bindingMarker, err := stalwartSubmissionBindingMarker(submission, message)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	if len(dispatchIDs) == 0 {
		if stalwartExactPreparedSubmission(message, submission, draftsID, draftMarker, locatorMarker, bindingMarker, dispatchMarker) {
			return EngineSubmissionAbsent{}, nil
		}
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	return engine.stalwartReconcileBoundSubmission(ctx, submission, message, draftsID, sentID, draftMarker, locatorMarker, bindingMarker, dispatchMarker)
}

func stalwartValidateSubmissionMessage(message EngineMessage, submission EngineSubmission) error {
	projection := message.Projection
	recipientCount := len(projection.To) + len(projection.CC) + len(projection.BCC)
	if projection.AccountID != submission.AccountID || projection.MessageID != submission.MessageID ||
		!validAddress(projection.From, true) ||
		!validAddresses(projection.To, MaxEngineDraftRecipientItems) ||
		!validAddresses(projection.CC, MaxEngineDraftRecipientItems) ||
		!validAddresses(projection.BCC, MaxEngineDraftRecipientItems) ||
		recipientCount < 1 || recipientCount > MaxEngineDraftTotalRecipientItems ||
		!uniqueRecipients(projection.To, projection.CC, projection.BCC) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if projection.SizeBytes > submission.MaxMessageBytes || int64(messageVariableBytes(message)) > submission.MaxMessageBytes {
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return nil
}

func stalwartExactPreparedSubmission(message EngineMessage, submission EngineSubmission, draftsID, draftMarker, locatorMarker, bindingMarker, dispatchMarker string) bool {
	projection := message.Projection
	return projection.AccountID == submission.AccountID &&
		projection.MessageID == submission.MessageID &&
		projection.Flags.Draft &&
		stalwartExactFolder(projection, draftsID) &&
		len(projection.ReservedKeywords) == 5 &&
		stalwartProjectionHasKeyword(projection, "$draft") &&
		stalwartProjectionHasKeyword(projection, draftMarker) &&
		stalwartProjectionHasKeyword(projection, locatorMarker) &&
		stalwartProjectionHasKeyword(projection, bindingMarker) &&
		!stalwartProjectionHasKeyword(projection, dispatchMarker)
}

func stalwartExactDispatchedSubmission(message EngineMessage, submission EngineSubmission, draftsID, sentID, draftMarker, locatorMarker, bindingMarker, dispatchMarker string) bool {
	projection := message.Projection
	return projection.AccountID == submission.AccountID &&
		projection.MessageID == submission.MessageID &&
		!projection.Flags.Draft &&
		stalwartExactFolder(projection, sentID) &&
		!projectionHasFolder(projection, draftsID) &&
		len(projection.ReservedKeywords) == 5 &&
		stalwartProjectionHasKeyword(projection, draftMarker) &&
		stalwartProjectionHasKeyword(projection, locatorMarker) &&
		stalwartProjectionHasKeyword(projection, bindingMarker) &&
		stalwartProjectionHasKeyword(projection, dispatchMarker)
}

func (engine *StalwartEngine) stalwartReconcileBoundSubmission(ctx context.Context, submission EngineSubmission, message EngineMessage, draftsID, sentID, draftMarker, locatorMarker, bindingMarker, dispatchMarker string) (EngineSubmissionResult, error) {
	if !stalwartExactDispatchedSubmission(message, submission, draftsID, sentID, draftMarker, locatorMarker, bindingMarker, dispatchMarker) {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	submissionIDs, submissionAmbiguous, err := engine.stalwartQuerySubmissionsByEmail(ctx, submission.AccountID, submission.MessageID)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	if submissionAmbiguous || len(submissionIDs) != 1 {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	return EngineSubmissionAccepted{SubmissionID: submissionIDs[0], MessageID: submission.MessageID}, nil
}

func stalwartSubmissionBindingMarker(submission EngineSubmission, message EngineMessage) (string, error) {
	projection := message.Projection
	binding := stalwartSubmissionBinding{
		Version:           stalwartSubmissionBindingVersion,
		AccountID:         submission.AccountID,
		DraftKey:          submission.DraftKey,
		SubmissionKey:     submission.SubmissionKey,
		MessageID:         submission.MessageID,
		ExpectedVersion:   submission.ExpectedVersion,
		InternetMessageID: projection.InternetMessageID,
		From:              stalwartSubmissionBindingAddress{Email: projection.From.Email, Name: projection.From.Name},
		To:                stalwartSubmissionBindingAddresses(projection.To),
		CC:                stalwartSubmissionBindingAddresses(projection.CC),
		BCC:               stalwartSubmissionBindingAddresses(projection.BCC),
		ReplyTo:           stalwartSubmissionBindingAddresses(projection.ReplyTo),
		Subject:           projection.Subject,
		InReplyTo:         projection.InReplyTo,
		References:        append([]string(nil), projection.References...),
		Priority:          projection.Priority,
		Body: stalwartSubmissionBindingBody{
			Format: message.Body.Format, PlainBody: message.Body.PlainBody, RichBody: message.Body.RichBody,
		},
		Attachments: stalwartSubmissionBindingAttachments(projection.Attachments),
	}
	payload, err := json.Marshal(binding)
	if err != nil {
		return "", stalwartEngineError(EngineErrorRejectedPolicy)
	}
	digest := sha256.Sum256(payload)
	return stalwartOpaqueMarker("submission-binding", string(digest[:]))
}

func stalwartSubmissionBindingAddresses(addresses []EngineAddress) []stalwartSubmissionBindingAddress {
	result := make([]stalwartSubmissionBindingAddress, len(addresses))
	for index, address := range addresses {
		result[index] = stalwartSubmissionBindingAddress{Email: address.Email, Name: address.Name}
	}
	return result
}

func stalwartSubmissionBindingAttachments(attachments []EngineProjectionAttachment) []stalwartSubmissionBindingAttachment {
	result := make([]stalwartSubmissionBindingAttachment, len(attachments))
	for index, attachment := range attachments {
		sha256Value := ""
		if len(attachment.SHA256) != 0 {
			sha256Value = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(attachment.SHA256)
		}
		result[index] = stalwartSubmissionBindingAttachment{
			BlobID: attachment.BlobID, ContentID: attachment.ContentID, Filename: attachment.Filename,
			MediaType: attachment.MediaType, SizeBytes: attachment.SizeBytes, Disposition: attachment.Disposition,
			SHA256: sha256Value,
		}
	}
	return result
}

func stalwartSubmissionSetRequest(arguments stalwartSubmissionSetArguments) (stalwart.JMAPRequest, error) {
	payload, err := json.Marshal(arguments)
	if err != nil {
		return stalwart.JMAPRequest{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name: "EmailSubmission/set", Arguments: payload, CallID: stalwartSubmissionSetCallID,
		}},
	}, nil
}

func stalwartEnsureRequestFitsTransport(request stalwart.JMAPRequest) error {
	fits, err := stalwart.RequestFitsTransport(request)
	if err != nil {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if !fits {
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return nil
}

func (engine *StalwartEngine) stalwartSaveDraftSet(ctx context.Context, accountID string, arguments stalwartDraftEmailSetArguments) (stalwart.JMAPResponse, error) {
	payload, err := json.Marshal(arguments)
	if err != nil {
		return stalwart.JMAPResponse{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return engine.stalwartCallBoundedMutation(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "Email/set", Arguments: payload, CallID: stalwartDraftEmailSetCallID}},
	})
}

func (engine *StalwartEngine) stalwartSubmitSet(ctx context.Context, accountID string, request stalwart.JMAPRequest) (stalwart.JMAPResponse, error) {
	return engine.stalwartCallBoundedMutation(ctx, accountID, request)
}

func (engine *StalwartEngine) stalwartBoundedEmailSet(ctx context.Context, accountID string, arguments stalwartEmailSetArguments) (stalwart.JMAPResponse, error) {
	payload, err := json.Marshal(arguments)
	if err != nil {
		return stalwart.JMAPResponse{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return engine.stalwartCallBoundedMutation(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "Email/set", Arguments: payload, CallID: stalwartEmailSetCallID}},
	})
}

func (engine *StalwartEngine) stalwartCallBoundedMutation(ctx context.Context, accountID string, request stalwart.JMAPRequest) (stalwart.JMAPResponse, error) {
	if err := stalwartEnsureRequestFitsTransport(request); err != nil {
		return stalwart.JMAPResponse{}, err
	}
	return engine.callMailbox(ctx, accountID, request, true)
}

func (engine *StalwartEngine) stalwartLookupDraft(ctx context.Context, accountID, draftsID, marker string, maximum int64) (EngineMessage, bool, bool, error) {
	ids, ambiguous, err := engine.stalwartQueryEmailsByMarker(ctx, accountID, marker)
	if err != nil || ambiguous || len(ids) == 0 {
		return EngineMessage{}, false, ambiguous, err
	}
	message, err := engine.stalwartReadDraftMessage(ctx, accountID, ids[0], maximum)
	if err != nil {
		return EngineMessage{}, false, false, err
	}
	if !message.Projection.Flags.Draft || !stalwartExactFolder(message.Projection, draftsID) || !stalwartProjectionHasKeyword(message.Projection, marker) {
		return EngineMessage{}, false, true, nil
	}
	return message, true, false, nil
}

func (engine *StalwartEngine) stalwartQueryEmailsByMarker(ctx context.Context, accountID, marker string) ([]string, bool, error) {
	if ctx == nil || !validAccountID(accountID) || !validStalwartHubKeyword(marker) {
		return nil, false, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	payload, err := json.Marshal(stalwartMarkerEmailQueryArguments{
		AccountID: accountID, CalculateTotal: true,
		Filter: stalwartMarkerEmailQueryFilter{HasKeyword: marker},
		Limit: stalwartMarkerQueryLimit, Position: 0,
	})
	if err != nil {
		return nil, false, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "Email/query", Arguments: payload, CallID: stalwartMarkerEmailQueryCallID}},
	}, false)
	if err != nil {
		return nil, false, err
	}
	query, err := stalwartParseEmailQuery(response, stalwartMarkerEmailQueryCallID, accountID)
	if err != nil {
		return nil, false, err
	}
	if query.Position != 0 || len(query.IDs) > stalwartMarkerQueryLimit || len(query.IDs) > *query.Total || len(query.IDs) == 0 && *query.Total != 0 {
		return nil, false, stalwartEngineError(EngineErrorUnavailable)
	}
	if *query.Total == 0 {
		return []string{}, false, nil
	}
	if *query.Total != 1 || len(query.IDs) != 1 {
		return append([]string(nil), query.IDs...), true, nil
	}
	return append([]string(nil), query.IDs...), false, nil
}

func (engine *StalwartEngine) stalwartReadDraftMessage(ctx context.Context, accountID, messageID string, maximum int64) (EngineMessage, error) {
	if ctx == nil || !validAccountID(accountID) || !validOpaqueID(messageID) || maximum < 1 || maximum > MaxEngineMessageBytes {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	bodyLimit := maximum
	if bodyLimit > MaxEngineBodyPartBytes {
		bodyLimit = MaxEngineBodyPartBytes
	}
	payload, err := json.Marshal(stalwartEmailGetArguments{
		AccountID: accountID, FetchHTMLBodyValues: true, FetchTextBodyValues: true,
		IDs: []string{messageID}, MaxBodyValueBytes: bodyLimit,
		Properties: append([]string(nil), stalwartDraftDetailProperties...),
	})
	if err != nil {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "Email/get", Arguments: payload, CallID: stalwartDraftEmailGetCallID}},
	}, false)
	if err != nil {
		return EngineMessage{}, err
	}
	get, err := stalwartParseEmailGet(response, stalwartDraftEmailGetCallID, accountID)
	if err != nil {
		return EngineMessage{}, err
	}
	if len(get.NotFound) != 0 {
		if len(get.NotFound) == 1 && get.NotFound[0] == messageID && len(get.List) == 0 {
			return EngineMessage{}, stalwartEngineError(EngineErrorNotFound)
		}
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if len(get.List) != 1 {
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	record, err := stalwartDecodeEmailRecord(get.List[0], MaxEngineProjectionPageBytes)
	if err != nil {
		return EngineMessage{}, stalwartMappingResponseError(err)
	}
	if record.ID != messageID {
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	var auxiliary struct {
		BCC        []stalwartEmailAddress `json:"bcc"`
		Importance *string                `json:"header:Importance"`
	}
	if err := json.Unmarshal(get.List[0], &auxiliary); err != nil {
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	projection, err := stalwartMapEmailProjection(accountID, get.State, record)
	if err != nil {
		return EngineMessage{}, err
	}
	bcc, err := stalwartMapAddresses(auxiliary.BCC, MaxEngineAddressItems)
	if err != nil || !uniqueRecipients(projection.To, projection.CC, bcc) {
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	priority, err := stalwartDraftPriority(auxiliary.Importance)
	if err != nil {
		return EngineMessage{}, err
	}
	projection.BCC = bcc
	projection.Priority = priority
	projection.Origin = EngineProjectionOriginConfirmedOwnerDraft
	attachments, err := stalwartMapDraftAttachments(record.Attachments, projection.Attachments)
	if err != nil {
		return EngineMessage{}, err
	}
	projection.Attachments = attachments
	body, err := stalwartMapEmailBody(record, bodyLimit)
	if err != nil {
		return EngineMessage{}, err
	}
	message := EngineMessage{Projection: projection, Body: body}
	if int64(messageVariableBytes(message)) > maximum {
		return EngineMessage{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	if err := validateEngineMessage(message); err != nil {
		return EngineMessage{}, stalwartMappingResponseError(err)
	}
	return message, nil
}

func (engine *StalwartEngine) stalwartResolveSubmissionContext(ctx context.Context, accountID, fromEmail string) (string, string, error) {
	identityPayload, err := json.Marshal(stalwartIdentityGetArguments{AccountID: accountID, Properties: []string{"id", "email"}})
	if err != nil {
		return "", "", stalwartEngineError(EngineErrorRejectedPolicy)
	}
	statePayload, err := json.Marshal(stalwartSubmissionGetArguments{AccountID: accountID, IDs: []string{}, Properties: []string{"id"}})
	if err != nil {
		return "", "", stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{
			{Name: "Identity/get", Arguments: identityPayload, CallID: stalwartIdentityGetCallID},
			{Name: "EmailSubmission/get", Arguments: statePayload, CallID: stalwartSubmissionStateCallID},
		},
	}, false)
	if err != nil {
		return "", "", err
	}
	identityRaw, err := stalwartDraftMethodPayload(response, "Identity/get", stalwartIdentityGetCallID)
	if err != nil {
		return "", "", err
	}
	stateRaw, err := stalwartDraftMethodPayload(response, "EmailSubmission/get", stalwartSubmissionStateCallID)
	if err != nil {
		return "", "", err
	}
	var identities stalwartIdentityGetResponse
	if err := json.Unmarshal(identityRaw, &identities); err != nil {
		return "", "", stalwartEngineError(EngineErrorUnavailable)
	}
	if identities.AccountID != accountID {
		return "", "", stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(identities.State) || len(identities.List) > stalwartMaximumIdentityItems || len(identities.NotFound) != 0 {
		return "", "", stalwartEngineError(EngineErrorUnavailable)
	}
	identityID := ""
	for _, identity := range identities.List {
		if !validOpaqueID(identity.ID) || !validSafeText(identity.Email, MaxEngineAddressEmailBytes, true) {
			return "", "", stalwartEngineError(EngineErrorUnavailable)
		}
		canonical, normalizeErr := NormalizeAddress(EngineAddress{Email: identity.Email})
		if normalizeErr != nil || canonical.Email != fromEmail {
			continue
		}
		if identityID != "" {
			return "", "", stalwartEngineError(EngineErrorUnavailable)
		}
		identityID = identity.ID
	}
	if identityID == "" {
		return "", "", stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	var state stalwartSubmissionGetResponse
	if err := json.Unmarshal(stateRaw, &state); err != nil {
		return "", "", stalwartEngineError(EngineErrorUnavailable)
	}
	if state.AccountID != accountID {
		return "", "", stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(state.State) || len(state.List) != 0 || len(state.NotFound) != 0 {
		return "", "", stalwartEngineError(EngineErrorUnavailable)
	}
	return identityID, state.State, nil
}

func (engine *StalwartEngine) stalwartQuerySubmissionsByEmail(ctx context.Context, accountID, messageID string) ([]string, bool, error) {
	if ctx == nil || !validAccountID(accountID) || !validOpaqueID(messageID) {
		return nil, false, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	payload, err := json.Marshal(stalwartSubmissionQueryArguments{
		AccountID: accountID, CalculateTotal: true,
		Filter: stalwartSubmissionQueryFilter{EmailIDs: []string{messageID}},
		Limit: stalwartMarkerQueryLimit, Position: 0,
	})
	if err != nil {
		return nil, false, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "EmailSubmission/query", Arguments: payload, CallID: stalwartSubmissionQueryCallID}},
	}, false)
	if err != nil {
		return nil, false, err
	}
	raw, err := stalwartDraftMethodPayload(response, "EmailSubmission/query", stalwartSubmissionQueryCallID)
	if err != nil {
		return nil, false, err
	}
	var query stalwartSubmissionQueryResponse
	if err := json.Unmarshal(raw, &query); err != nil || query.AccountID != accountID || !validOpaqueID(query.QueryState) || query.Position != 0 || query.Total == nil || *query.Total < 0 || *query.Total > stalwartMaximumQueryPosition || len(query.IDs) > stalwartMarkerQueryLimit || !validOpaqueIDs(query.IDs) || len(query.IDs) > *query.Total || len(query.IDs) == 0 && *query.Total != 0 {
		return nil, false, stalwartEngineError(EngineErrorUnavailable)
	}
	if *query.Total == 0 {
		return []string{}, false, nil
	}
	if *query.Total != 1 || len(query.IDs) != 1 {
		return append([]string(nil), query.IDs...), true, nil
	}
	return append([]string(nil), query.IDs...), false, nil
}

func stalwartValidateDraftSetResponse(response stalwart.JMAPResponse, draft EngineDraft, expectedState string) (string, string, error) {
	raw, err := stalwartDraftMethodPayload(response, "Email/set", stalwartDraftEmailSetCallID)
	if err != nil {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	set, err := stalwartParseEmailSetPayload(raw, draft.AccountID, expectedState)
	if err != nil {
		return "", "", err
	}
	if len(set.NotCreated) != 0 {
		if len(set.Created) != 0 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.Destroyed) != 0 || len(set.NotDestroyed) != 0 || set.OldStateKnown && set.NewState != set.OldState {
			return "", "", stalwartEngineError(EngineErrorIndeterminate)
		}
		return "", "", stalwartDraftSetFailuresFor(set.NotCreated)
	}
	if len(set.Created) != 1 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.NotDestroyed) != 0 || set.OldStateKnown && set.NewState == set.OldState {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	createdRaw, exists := set.Created[stalwartDraftCreationID]
	if !exists {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	var created stalwartCreatedEmail
	if err := json.Unmarshal(createdRaw, &created); err != nil || !validOpaqueID(created.ID) || !validOpaqueID(created.BlobID) || !validOpaqueID(created.ThreadID) || created.Size < 0 || created.Size > MaxEngineMessageBytes {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	if draft.MessageID == "" {
		if len(set.Destroyed) != 0 {
			return "", "", stalwartEngineError(EngineErrorIndeterminate)
		}
	} else if !stalwartEmailDestroyedExactly(set.Destroyed, []string{draft.MessageID}) {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	return created.ID, set.NewState, nil
}

func stalwartValidateSubmissionSetResponse(response stalwart.JMAPResponse, submission EngineSubmission, expectedState, expectedEmailState string) (string, error) {
	if len(response.MethodResponses) < 1 || len(response.MethodResponses) > 2 {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	first := response.MethodResponses[0]
	if first.Name != "EmailSubmission/set" || first.CallID != stalwartSubmissionSetCallID {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	set, err := stalwartParseEmailSetPayload(first.Arguments, submission.AccountID, expectedState)
	if err != nil {
		return "", err
	}
	if len(set.NotCreated) != 0 {
		if len(response.MethodResponses) != 1 || len(set.Created) != 0 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.Destroyed) != 0 || len(set.NotDestroyed) != 0 || set.OldStateKnown && set.NewState != set.OldState {
			return "", stalwartEngineError(EngineErrorIndeterminate)
		}
		return "", stalwartSubmissionSetFailuresFor(set.NotCreated)
	}
	if len(response.MethodResponses) != 2 || len(set.Created) != 1 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.Destroyed) != 0 || len(set.NotDestroyed) != 0 || set.OldStateKnown && set.NewState == set.OldState {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	createdRaw, exists := set.Created[stalwartSubmissionCreationID]
	if !exists {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	var created stalwartCreatedSubmission
	if err := json.Unmarshal(createdRaw, &created); err != nil || !validOpaqueID(created.ID) {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	implicit := response.MethodResponses[1]
	if implicit.Name != "Email/set" || implicit.CallID != stalwartSubmissionSetCallID {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	emailSet, err := stalwartParseEmailSetPayload(implicit.Arguments, submission.AccountID, expectedEmailState)
	if err != nil {
		return "", err
	}
	if emailSet.OldStateKnown && emailSet.NewState == emailSet.OldState || len(emailSet.Created) != 0 || len(emailSet.NotCreated) != 0 || len(emailSet.NotUpdated) != 0 || len(emailSet.Destroyed) != 0 || len(emailSet.NotDestroyed) != 0 || !stalwartEmailUpdatedExactly(emailSet.Updated, []string{submission.MessageID}) {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	return created.ID, nil
}

func stalwartParseEmailSetPayload(raw json.RawMessage, accountID, expectedState string) (stalwartEmailSetResponse, error) {
	return parseStalwartEmailSetResponse(stalwart.JMAPResponse{
		MethodResponses: []stalwart.JMAPMethodResponse{{
			Name: "Email/set", Arguments: raw, CallID: stalwartEmailSetCallID,
		}},
	}, accountID, expectedState)
}

func stalwartDraftSetFailuresFor(failures map[string]stalwartMailboxSetFailure) error {
	failure, exists := failures[stalwartDraftCreationID]
	if len(failures) != 1 || !exists || failure.Type == "" {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	switch failure.Type {
	case "overQuota", "tooManyKeywords", "tooManyMailboxes":
		return stalwartEngineError(EngineErrorRejectedPolicy)
	default:
		return stalwartEmailSetFailuresFor(failures, []string{stalwartDraftCreationID})
	}
}

func stalwartSubmissionSetFailuresFor(failures map[string]stalwartMailboxSetFailure) error {
	failure, exists := failures[stalwartSubmissionCreationID]
	if len(failures) != 1 || !exists || failure.Type == "" {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	switch failure.Type {
	case "tooManyRecipients":
		return stalwartEngineError(EngineErrorPayloadTooLarge)
	case "overQuota", "invalidEmail", "noRecipients", "invalidRecipients", "forbiddenMailFrom", "forbiddenFrom", "forbiddenToSend":
		return stalwartEngineError(EngineErrorRejectedPolicy)
	default:
		return stalwartEmailSetFailuresFor(failures, []string{stalwartSubmissionCreationID})
	}
}

func stalwartSubmissionMarkerUpdateSucceeded(set stalwartEmailSetResponse, messageID string) error {
	if len(set.Created) != 0 || len(set.NotCreated) != 0 || len(set.Destroyed) != 0 || len(set.NotDestroyed) != 0 {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	if len(set.NotUpdated) != 0 {
		if len(set.Updated) != 0 || set.OldStateKnown && set.NewState != set.OldState {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		failure, exists := set.NotUpdated[messageID]
		if len(set.NotUpdated) != 1 || !exists || failure.Type == "" {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		switch failure.Type {
		case "overQuota", "tooManyKeywords", "tooManyMailboxes":
			return stalwartEngineError(EngineErrorRejectedPolicy)
		default:
			return stalwartEmailSetFailuresFor(set.NotUpdated, []string{messageID})
		}
	}
	if !stalwartEmailUpdatedExactly(set.Updated, []string{messageID}) {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	return nil
}

func stalwartDraftCreate(content EngineMessageContent, draftsID, draftMarker, mutationMarker string) stalwartDraftEmailCreate {
	created := stalwartDraftEmailCreate{
		MailboxIDs: map[string]bool{draftsID: true},
		Keywords: map[string]bool{"$draft": true, draftMarker: true, mutationMarker: true},
		From: stalwartDraftAddresses([]EngineAddress{content.From}),
		To: stalwartDraftAddresses(content.To),
		CC: stalwartDraftAddresses(content.CC),
		BCC: stalwartDraftAddresses(content.BCC),
		ReplyTo: stalwartDraftAddresses(content.ReplyTo),
		Subject: content.Subject,
		References: append([]string(nil), content.References...),
		Importance: stalwartDraftImportance(content.Priority),
		BodyValues: make(map[string]stalwartDraftBodyValue, 2),
	}
	if content.InReplyTo != "" {
		created.InReplyTo = []string{content.InReplyTo}
	}
	created.TextBody = []stalwartDraftBodyPart{{PartID: stalwartDraftTextPartID, Type: "text/plain"}}
	created.BodyValues[stalwartDraftTextPartID] = stalwartDraftBodyValue{Value: content.PlainBody}
	if content.Format == EngineBodyRich {
		created.HTMLBody = []stalwartDraftBodyPart{{PartID: stalwartDraftHTMLPartID, Type: "text/html"}}
		created.BodyValues[stalwartDraftHTMLPartID] = stalwartDraftBodyValue{Value: content.RichBody}
	}
	return created
}

func stalwartDraftAddresses(addresses []EngineAddress) []stalwartEmailAddress {
	result := make([]stalwartEmailAddress, len(addresses))
	for index, address := range addresses {
		result[index] = stalwartEmailAddress{Email: address.Email, Name: address.Name}
	}
	return result
}

func stalwartSubmissionRecipients(projection EngineMessageProjection) []stalwartSubmissionAddress {
	result := make([]stalwartSubmissionAddress, 0, len(projection.To)+len(projection.CC)+len(projection.BCC))
	for _, collection := range [][]EngineAddress{projection.To, projection.CC, projection.BCC} {
		for _, address := range collection {
			result = append(result, stalwartSubmissionAddress{Email: address.Email})
		}
	}
	return result
}

func stalwartDraftMethodPayload(response stalwart.JMAPResponse, name, callID string) (json.RawMessage, error) {
	for _, method := range response.MethodResponses {
		if method.CallID == callID && method.Name == name && len(method.Arguments) != 0 {
			return append(json.RawMessage(nil), method.Arguments...), nil
		}
	}
	return nil, stalwartEngineError(EngineErrorUnavailable)
}

func stalwartDraftReconciliationError(err error) (EngineDraftResult, error) {
	if stalwartReconciliationIndeterminate(err) {
		return EngineDraftIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	return nil, err
}

func stalwartDraftDeletionReconciliationError(err error) (EngineDraftDeletionResult, error) {
	var engineErr *EngineError
	if errors.As(err, &engineErr) && engineErr != nil {
		switch engineErr.Kind {
		case EngineErrorRejectedPolicy, EngineErrorUnauthorizedAccount, EngineErrorPayloadTooLarge:
			return nil, err
		}
	}
	return EngineDraftDeletionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
}

func stalwartSubmissionReconciliationError(err error) (EngineSubmissionResult, error) {
	var engineErr *EngineError
	if errors.As(err, &engineErr) && engineErr != nil {
		switch engineErr.Kind {
		case EngineErrorRejectedPolicy, EngineErrorUnauthorizedAccount, EngineErrorPayloadTooLarge:
			return nil, err
		}
	}
	return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
}

func stalwartReconciliationIndeterminate(err error) bool {
	var engineErr *EngineError
	if !errors.As(err, &engineErr) {
		return false
	}
	return engineErr.Kind == EngineErrorUnavailable || engineErr.Kind == EngineErrorIndeterminate || engineErr.Kind == EngineErrorRestartRequired
}

func stalwartOpaqueMarker(kind string, values ...string) (string, error) {
	if kind == "" || len(values) == 0 {
		return "", stalwartEngineError(EngineErrorRejectedPolicy)
	}
	hash := sha256.New()
	hash.Write([]byte("hub-mail-marker-v1"))
	hash.Write([]byte{0})
	hash.Write([]byte(kind))
	for _, value := range values {
		hash.Write([]byte{0})
		hash.Write([]byte(value))
	}
	encoded := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(nil)))
	marker := stalwartHubKeywordPrefix + encoded
	if !validStalwartHubKeyword(marker) {
		return "", stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return marker, nil
}

func stalwartDraftMutationMarkerV2(accountID, draftKey, mutationID string) (string, error) {
	return stalwartOpaqueMarker("draft-mutation-v2", accountID, draftKey, mutationID)
}

func stalwartValidDraftThreading(content EngineMessageContent) bool {
	if content.InReplyTo != "" && !validThreadMessageID(content.InReplyTo) {
		return false
	}
	for _, reference := range content.References {
		if !validThreadMessageID(reference) {
			return false
		}
	}
	return true
}

func stalwartExactFolderRole(folders []EngineFolder, role EngineFolderRole) (string, error) {
	folderID := ""
	for _, folder := range folders {
		if folder.Role != role {
			continue
		}
		if folderID != "" {
			return "", stalwartEngineError(EngineErrorUnavailable)
		}
		folderID = folder.FolderID
	}
	if folderID == "" {
		return "", stalwartEngineError(EngineErrorNotFound)
	}
	return folderID, nil
}

func stalwartExactDraftOperation(message EngineMessage, accountID, draftsID, draftMarker, mutationMarker string) bool {
	if message.Projection.AccountID != accountID || !message.Projection.Flags.Draft || !stalwartExactFolder(message.Projection, draftsID) || len(message.Projection.ReservedKeywords) != 3 {
		return false
	}
	expected := map[EngineKeywordMarker]struct{}{
		EngineKeywordMarker("$draft"): {},
		EngineKeywordMarker(draftMarker): {},
		EngineKeywordMarker(mutationMarker): {},
	}
	for _, keyword := range message.Projection.ReservedKeywords {
		if _, exists := expected[keyword]; !exists {
			return false
		}
		delete(expected, keyword)
	}
	return len(expected) == 0
}

func stalwartExactDraftDeletion(message EngineMessage, request EngineDraftDeletion, draftsID, draftMarker string) bool {
	projection := message.Projection
	return projection.AccountID == request.AccountID &&
		projection.MessageID == request.MessageID &&
		projection.Version == request.ExpectedVersion &&
		projection.Flags.Draft &&
		stalwartExactFolder(projection, draftsID) &&
		stalwartProjectionHasKeyword(projection, "$draft") &&
		stalwartProjectionHasKeyword(projection, draftMarker)
}

func stalwartExactFolder(projection EngineMessageProjection, folderID string) bool {
	return len(projection.FolderIDs) == 1 && projection.FolderIDs[0] == folderID
}

func stalwartProjectionHasKeyword(projection EngineMessageProjection, marker string) bool {
	for _, keyword := range projection.ReservedKeywords {
		if string(keyword) == marker {
			return true
		}
	}
	return false
}

func stalwartDraftImportance(priority EngineMessagePriority) string {
	switch priority {
	case EnginePriorityHigh:
		return "high"
	case EnginePriorityLow:
		return "low"
	default:
		return ""
	}
}

func stalwartDraftPriority(value *string) (EngineMessagePriority, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return EnginePriorityNormal, nil
	}
	if !validSafeText(*value, MaxEngineHeaderValueBytes, false) {
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
	switch strings.ToLower(strings.TrimSpace(*value)) {
	case "high":
		return EnginePriorityHigh, nil
	case "normal":
		return EnginePriorityNormal, nil
	case "low":
		return EnginePriorityLow, nil
	default:
		return "", stalwartEngineError(EngineErrorUnavailable)
	}
}
