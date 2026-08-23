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

type stalwartSubmissionSetResponse struct {
	AccountID    string                                      `json:"accountId"`
	OldState     string                                      `json:"oldState"`
	NewState     string                                      `json:"newState"`
	Created      map[string]json.RawMessage                  `json:"created"`
	NotCreated   map[string]stalwartMailboxSetFailure        `json:"notCreated"`
	Updated      map[string]json.RawMessage                  `json:"updated"`
	NotUpdated   map[string]stalwartMailboxSetFailure        `json:"notUpdated"`
	Destroyed    []string                                    `json:"destroyed"`
	NotDestroyed map[string]stalwartMailboxSetFailure        `json:"notDestroyed"`
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
	mutationMarker, err := stalwartOpaqueMarker("draft-mutation", draft.AccountID, draft.DraftKey, draft.MessageID, draft.PreviousVersion, draft.MutationID)
	if err != nil {
		return EngineMessage{}, err
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
	if found && stalwartExactDraftOperation(existing, draftsID, draftMarker, mutationMarker) {
		return existing, nil
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
		IfInState: draft.PreviousVersion,
		Create: map[string]stalwartDraftEmailCreate{stalwartDraftCreationID: created},
	}
	if draft.MessageID != "" {
		arguments.Destroy = []string{draft.MessageID}
	}
	response, err := engine.stalwartSaveDraftSet(ctx, draft.AccountID, arguments)
	if err != nil {
		return EngineMessage{}, err
	}
	createdID, _, err := stalwartValidateDraftSetResponse(response, draft)
	if err != nil {
		return EngineMessage{}, err
	}
	message, err := engine.stalwartReadDraftMessage(ctx, draft.AccountID, createdID, draft.MaxContentBytes)
	if err != nil {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if !stalwartExactDraftOperation(message, draftsID, draftMarker, mutationMarker) || draft.MessageID != "" && createdID == draft.MessageID {
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
	mutationMarker, err := stalwartOpaqueMarker("draft-mutation", request.AccountID, request.DraftKey, request.MessageID, request.PreviousVersion, request.MutationID)
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
	if !stalwartExactDraftOperation(message, draftsID, draftMarker, mutationMarker) || request.MessageID != "" && message.Projection.Version == request.PreviousVersion {
		return EngineDraftIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	return EngineDraftApplied{Message: message}, nil
}

// Submit first records an opaque submission marker with Email/set guarded by
// the caller's Email state, then creates one EmailSubmission. The marker makes
// a lost submission response reconcilable without resending the message.
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
	message, found, ambiguous, err := engine.stalwartLookupDraft(ctx, submission.AccountID, draftsID, draftMarker, submission.MaxMessageBytes)
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
	if stalwartProjectionHasKeyword(message.Projection, submissionMarker) {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}
	if message.Projection.SizeBytes > submission.MaxMessageBytes || int64(messageVariableBytes(message)) > submission.MaxMessageBytes {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	identityID, submissionState, err := engine.stalwartResolveSubmissionContext(ctx, submission.AccountID, message.Projection.From.Email)
	if err != nil {
		return nil, err
	}
	markerResponse, err := engine.stalwartEmailSet(ctx, submission.AccountID, stalwartEmailSetArguments{
		AccountID: submission.AccountID,
		IfInState: submission.ExpectedVersion,
		Update: map[string]map[string]any{
			submission.MessageID: {"keywords/" + submissionMarker: true},
		},
	})
	if err != nil {
		return nil, err
	}
	markerSet, err := parseStalwartEmailSetResponse(markerResponse, submission.AccountID, submission.ExpectedVersion)
	if err != nil {
		return nil, err
	}
	if err := stalwartEmailOnlyUpdateSucceeded(markerSet, []string{submission.MessageID}); err != nil {
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
	response, err := engine.stalwartSubmitSet(ctx, submission.AccountID, arguments)
	if err != nil {
		return nil, err
	}
	submissionID, err := stalwartValidateSubmissionSetResponse(response, submission, submissionState, markerSet.NewState)
	if err != nil {
		return nil, err
	}
	projections, _, notFound, err := engine.stalwartFetchProjections(ctx, submission.AccountID, []string{submission.MessageID})
	if err != nil || len(notFound) != 0 || len(projections) != 1 || projections[0].Flags.Draft || !stalwartExactFolder(projections[0], sentID) || !stalwartProjectionHasKeyword(projections[0], submissionMarker) {
		return nil, stalwartEngineError(EngineErrorIndeterminate)
	}
	return EngineSubmissionSent{SubmissionID: submissionID, MessageID: submission.MessageID}, nil
}

// ReconcileSubmission joins the opaque Email marker to a bounded
// EmailSubmission/query. It proves both provider objects and the sent-folder
// transition before reporting Sent.
func (engine *StalwartEngine) ReconcileSubmission(ctx context.Context, accountID, submissionKey string) (EngineSubmissionResult, error) {
	if ctx == nil || !validAccountID(accountID) || !validSafeText(submissionKey, MaxEngineOpaqueIDBytes, true) {
		return nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	marker, err := stalwartOpaqueMarker("submission", accountID, submissionKey)
	if err != nil {
		return nil, err
	}
	folders, _, err := engine.listFolderSnapshot(ctx, accountID)
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
	ids, ambiguous, err := engine.stalwartQueryEmailsByMarker(ctx, accountID, marker)
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	if len(ids) == 0 {
		return EngineSubmissionAbsent{}, nil
	}
	if ambiguous || len(ids) != 1 {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	projections, _, notFound, err := engine.stalwartFetchProjections(ctx, accountID, ids)
	if err != nil || len(notFound) != 0 || len(projections) != 1 || !stalwartProjectionHasKeyword(projections[0], marker) {
		if err != nil {
			return stalwartSubmissionReconciliationError(err)
		}
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	submissionIDs, submissionAmbiguous, err := engine.stalwartQuerySubmissionsByEmail(ctx, accountID, ids[0])
	if err != nil {
		return stalwartSubmissionReconciliationError(err)
	}
	projection := projections[0]
	if len(submissionIDs) == 0 {
		if projection.Flags.Draft && stalwartExactFolder(projection, draftsID) {
			return EngineSubmissionAbsent{}, nil
		}
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	if submissionAmbiguous || len(submissionIDs) != 1 || projection.Flags.Draft || !stalwartExactFolder(projection, sentID) || projectionHasFolder(projection, draftsID) {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	return EngineSubmissionSent{SubmissionID: submissionIDs[0], MessageID: ids[0]}, nil
}

func (engine *StalwartEngine) stalwartSaveDraftSet(ctx context.Context, accountID string, arguments stalwartDraftEmailSetArguments) (stalwart.JMAPResponse, error) {
	payload, err := json.Marshal(arguments)
	if err != nil {
		return stalwart.JMAPResponse{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "Email/set", Arguments: payload, CallID: stalwartDraftEmailSetCallID}},
	}, true)
}

func (engine *StalwartEngine) stalwartSubmitSet(ctx context.Context, accountID string, arguments stalwartSubmissionSetArguments) (stalwart.JMAPResponse, error) {
	payload, err := json.Marshal(arguments)
	if err != nil {
		return stalwart.JMAPResponse{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{Name: "EmailSubmission/set", Arguments: payload, CallID: stalwartSubmissionSetCallID}},
	}, true)
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

func stalwartValidateDraftSetResponse(response stalwart.JMAPResponse, draft EngineDraft) (string, string, error) {
	raw, err := stalwartDraftMethodPayload(response, "Email/set", stalwartDraftEmailSetCallID)
	if err != nil {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	var set stalwartEmailSetResponse
	if err := json.Unmarshal(raw, &set); err != nil {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	if set.AccountID != draft.AccountID {
		return "", "", stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(set.OldState) || !validOpaqueID(set.NewState) {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	if draft.PreviousVersion != "" && set.OldState != draft.PreviousVersion {
		return "", "", stalwartEngineError(EngineErrorIndeterminate)
	}
	if len(set.NotCreated) != 0 {
		if len(set.Created) != 0 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.Destroyed) != 0 || len(set.NotDestroyed) != 0 || set.NewState != set.OldState {
			return "", "", stalwartEngineError(EngineErrorIndeterminate)
		}
		return "", "", stalwartEmailSetFailuresFor(set.NotCreated, []string{stalwartDraftCreationID})
	}
	if len(set.Created) != 1 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.NotDestroyed) != 0 || set.NewState == set.OldState {
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
	var set stalwartSubmissionSetResponse
	if err := json.Unmarshal(first.Arguments, &set); err != nil {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	if set.AccountID != submission.AccountID {
		return "", stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if set.OldState != expectedState || !validOpaqueID(set.NewState) {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	if len(set.NotCreated) != 0 {
		if len(response.MethodResponses) != 1 || len(set.Created) != 0 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.Destroyed) != 0 || len(set.NotDestroyed) != 0 || set.NewState != set.OldState {
			return "", stalwartEngineError(EngineErrorIndeterminate)
		}
		return "", stalwartEmailSetFailuresFor(set.NotCreated, []string{stalwartSubmissionCreationID})
	}
	if len(response.MethodResponses) != 2 || len(set.Created) != 1 || len(set.Updated) != 0 || len(set.NotUpdated) != 0 || len(set.Destroyed) != 0 || len(set.NotDestroyed) != 0 || set.NewState == set.OldState {
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
	var emailSet stalwartEmailSetResponse
	if err := json.Unmarshal(implicit.Arguments, &emailSet); err != nil {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	if emailSet.AccountID != submission.AccountID {
		return "", stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if emailSet.OldState != expectedEmailState || !validOpaqueID(emailSet.NewState) || emailSet.NewState == emailSet.OldState || len(emailSet.Created) != 0 || len(emailSet.NotCreated) != 0 || len(emailSet.NotUpdated) != 0 || len(emailSet.Destroyed) != 0 || len(emailSet.NotDestroyed) != 0 || !stalwartEmailUpdatedExactly(emailSet.Updated, []string{submission.MessageID}) {
		return "", stalwartEngineError(EngineErrorIndeterminate)
	}
	return created.ID, nil
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

func stalwartSubmissionReconciliationError(err error) (EngineSubmissionResult, error) {
	if stalwartReconciliationIndeterminate(err) {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	return nil, err
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

func stalwartExactDraftOperation(message EngineMessage, draftsID, draftMarker, mutationMarker string) bool {
	if !message.Projection.Flags.Draft || !stalwartExactFolder(message.Projection, draftsID) || len(message.Projection.ReservedKeywords) != 3 {
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
