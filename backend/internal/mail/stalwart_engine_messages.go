package mail

import (
	"context"
	"encoding/json"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartEmailQueryCallID = "email-query"
	stalwartEmailGetCallID   = "email-get"
)

// ListMessages reads one stable, account-bound projection page. It first asks
// Email/query for a bounded ID set, then obtains metadata for that one set in
// one Email/get request; it never asks the provider for bodies or blob bytes.
func (engine *StalwartEngine) ListMessages(ctx context.Context, query EngineMessageQuery) (EngineProjectionPage, error) {
	if ctx == nil || !validAccountID(query.AccountID) || !validOpaqueID(query.FolderID) {
		return EngineProjectionPage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if err := validatePageRequest(query.Page); err != nil {
		return EngineProjectionPage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	position := 0
	previousState := ""
	if query.Cursor != "" {
		cursor, err := stalwartParseCursor(query.Cursor, query.AccountID, stalwartCursorList, query.FolderID)
		if err != nil {
			return EngineProjectionPage{}, err
		}
		position = cursor.Position
		previousState = cursor.State
	}
	arguments, err := json.Marshal(stalwartEmailQueryArguments{
		AccountID:      query.AccountID,
		CalculateTotal: true,
		Filter:         stalwartEmailQueryFilter{InMailbox: query.FolderID},
		Limit:          query.Page.Limit,
		Position:       position,
		Sort: []stalwartEmailComparator{
			{Property: "receivedAt", IsAscending: false},
			{Property: "id", IsAscending: false},
		},
	})
	if err != nil {
		return EngineProjectionPage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, query.AccountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/query",
			Arguments: arguments,
			CallID:    stalwartEmailQueryCallID,
		}},
	}, false)
	if err != nil {
		return EngineProjectionPage{}, err
	}
	pageResult, err := stalwartParseEmailQuery(response, stalwartEmailQueryCallID, query.AccountID)
	if err != nil {
		return EngineProjectionPage{}, err
	}
	if pageResult.Position != position || len(pageResult.IDs) > query.Page.Limit || previousState != "" && pageResult.QueryState != previousState {
		return EngineProjectionPage{}, stalwartEngineError(EngineErrorRestartRequired)
	}
	currentCursor, err := stalwartPackCursor(stalwartMessageCursor{AccountID: query.AccountID, FolderID: query.FolderID, Kind: stalwartCursorList, Position: position, State: pageResult.QueryState})
	if err != nil {
		return EngineProjectionPage{}, err
	}
	projections, _, notFound, err := engine.stalwartFetchProjections(ctx, query.AccountID, pageResult.IDs)
	if err != nil {
		return EngineProjectionPage{}, err
	}
	if len(notFound) != 0 {
		return EngineProjectionPage{}, stalwartEngineError(EngineErrorRestartRequired)
	}
	for _, projection := range projections {
		if !projectionHasFolder(projection, query.FolderID) {
			return EngineProjectionPage{}, stalwartEngineError(EngineErrorUnavailable)
		}
	}
	nextPosition := position + len(pageResult.IDs)
	nextCursor := ""
	if nextPosition < *pageResult.Total {
		nextCursor, err = stalwartPackCursor(stalwartMessageCursor{AccountID: query.AccountID, FolderID: query.FolderID, Kind: stalwartCursorList, Position: nextPosition, State: pageResult.QueryState})
		if err != nil {
			return EngineProjectionPage{}, err
		}
	}
	page := EngineProjectionPage{AccountID: query.AccountID, Cursor: currentCursor, NextCursor: nextCursor, Bounds: EnginePageBounds{Requested: query.Page.Limit, Maximum: MaxEnginePageItems, Returned: len(projections)}, Messages: projections}
	if err := validateProjectionPage(page); err != nil {
		return EngineProjectionPage{}, stalwartMappingResponseError(err)
	}
	return page, nil
}

// GetMessage returns provider source body bytes represented as bounded UTF-8
// text. It deliberately does not sanitize HTML, load remote images, or expose
// a provider URL; those are public-handler responsibilities.
func (engine *StalwartEngine) GetMessage(ctx context.Context, request EngineMessageReadRequest) (EngineMessage, error) {
	if ctx == nil || !validAccountID(request.AccountID) || !validOpaqueID(request.MessageID) || request.MaxContentBytes < 1 || request.MaxContentBytes > MaxEngineMessageBytes {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	bodyLimit := request.MaxContentBytes
	if bodyLimit > MaxEngineBodyPartBytes {
		bodyLimit = MaxEngineBodyPartBytes
	}
	arguments, err := json.Marshal(stalwartEmailGetArguments{
		AccountID:           request.AccountID,
		FetchHTMLBodyValues: true,
		FetchTextBodyValues: true,
		IDs:                 []string{request.MessageID},
		MaxBodyValueBytes:   bodyLimit,
		Properties:          append([]string(nil), stalwartEmailDetailProperties...),
	})
	if err != nil {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, request.AccountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/get",
			Arguments: arguments,
			CallID:    stalwartEmailGetCallID,
		}},
	}, false)
	if err != nil {
		return EngineMessage{}, err
	}
	getResult, err := stalwartParseEmailGet(response, stalwartEmailGetCallID, request.AccountID)
	if err != nil {
		return EngineMessage{}, err
	}
	if len(getResult.NotFound) != 0 {
		if len(getResult.NotFound) == 1 && getResult.NotFound[0] == request.MessageID && len(getResult.List) == 0 {
			return EngineMessage{}, stalwartEngineError(EngineErrorNotFound)
		}
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	if len(getResult.List) != 1 {
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	record, err := stalwartDecodeEmailRecord(getResult.List[0], MaxEngineProjectionPageBytes)
	if err != nil {
		return EngineMessage{}, err
	}
	if record.ID != request.MessageID {
		return EngineMessage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	projection, err := stalwartMapEmailProjection(request.AccountID, getResult.State, record)
	if err != nil {
		return EngineMessage{}, err
	}
	body, err := stalwartMapEmailBody(record, bodyLimit)
	if err != nil {
		return EngineMessage{}, err
	}
	message := EngineMessage{Projection: projection, Body: body}
	if int64(messageVariableBytes(message)) > request.MaxContentBytes {
		return EngineMessage{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	return message, nil
}

// GetBlob is intentionally closed until the accepted account-bound client
// exposes the validated Stalwart download-session boundary. Calling raw URLs
// or rediscovering a session here would bypass the frozen credential scope.
func (engine *StalwartEngine) GetBlob(ctx context.Context, request EngineBlobReadRequest) (EngineBlob, error) {
	if ctx == nil || !validAccountID(request.AccountID) || !validOpaqueID(request.BlobID) || request.MaxBytes < 1 || request.MaxBytes > MaxEngineAttachmentBytes {
		return EngineBlob{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return EngineBlob{}, stalwartEngineError(EngineErrorRejectedPolicy)
}

func (engine *StalwartEngine) stalwartFetchProjections(ctx context.Context, accountID string, ids []string) ([]EngineMessageProjection, string, []string, error) {
	if ctx == nil || !validAccountID(accountID) || len(ids) > MaxEnginePageItems || !validOpaqueIDs(ids) {
		return nil, "", nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if len(ids) == 0 {
		return []EngineMessageProjection{}, "", []string{}, nil
	}
	arguments, err := json.Marshal(stalwartEmailGetArguments{AccountID: accountID, IDs: append([]string(nil), ids...), Properties: append([]string(nil), stalwartEmailProjectionProperties...)})
	if err != nil {
		return nil, "", nil, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/get",
			Arguments: arguments,
			CallID:    stalwartEmailGetCallID,
		}},
	}, false)
	if err != nil {
		return nil, "", nil, err
	}
	getResult, err := stalwartParseEmailGet(response, stalwartEmailGetCallID, accountID)
	if err != nil {
		return nil, "", nil, err
	}
	if len(getResult.List)+len(getResult.NotFound) != len(ids) {
		return nil, "", nil, stalwartEngineError(EngineErrorUnavailable)
	}
	byID := make(map[string]EngineMessageProjection, len(getResult.List))
	for _, raw := range getResult.List {
		record, err := stalwartDecodeEmailRecord(raw, MaxEngineProjectionItemBytes)
		if err != nil {
			return nil, "", nil, err
		}
		projection, err := stalwartMapEmailProjection(accountID, getResult.State, record)
		if err != nil {
			return nil, "", nil, err
		}
		if _, duplicate := byID[projection.MessageID]; duplicate {
			return nil, "", nil, stalwartEngineError(EngineErrorUnavailable)
		}
		byID[projection.MessageID] = projection
	}
	notFound := append([]string(nil), getResult.NotFound...)
	notFoundSet := make(map[string]struct{}, len(notFound))
	for _, id := range notFound {
		if _, duplicate := notFoundSet[id]; duplicate {
			return nil, "", nil, stalwartEngineError(EngineErrorUnavailable)
		}
		notFoundSet[id] = struct{}{}
	}
	projections := make([]EngineMessageProjection, 0, len(ids))
	for _, id := range ids {
		if _, missing := notFoundSet[id]; missing {
			continue
		}
		projection, exists := byID[id]
		if !exists {
			return nil, "", nil, stalwartEngineError(EngineErrorUnavailable)
		}
		projections = append(projections, projection)
	}
	return projections, getResult.State, notFound, nil
}
