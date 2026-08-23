package mail

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const stalwartEmailChangesCallID = "email-changes"

// Changes obtains one bounded JMAP Email/changes transition and hydrates the
// changed objects in one additional bounded Email/get request. A stale or
// unhydrated transition never becomes a partial projection page.
func (engine *StalwartEngine) Changes(ctx context.Context, request EngineChangeRequest) (EngineChangePage, error) {
	if ctx == nil || !validAccountID(request.AccountID) {
		return EngineChangePage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if err := validatePageRequest(request.Page); err != nil {
		return EngineChangePage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	cursor, err := stalwartParseCursor(request.FromCursor, request.AccountID, stalwartCursorChanges, "")
	if err != nil {
		return EngineChangePage{}, err
	}
	arguments, err := json.Marshal(stalwartEmailChangesArguments{AccountID: request.AccountID, MaxChanges: request.Page.Limit, SinceState: cursor.State})
	if err != nil {
		return EngineChangePage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, request.AccountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/changes",
			Arguments: arguments,
			CallID:    stalwartEmailChangesCallID,
		}},
	}, false)
	if err != nil {
		return EngineChangePage{}, err
	}
	changes, err := stalwartParseEmailChanges(response, stalwartEmailChangesCallID, request.AccountID)
	if err != nil {
		return EngineChangePage{}, err
	}
	if changes.OldState != cursor.State {
		return EngineChangePage{}, stalwartEngineError(EngineErrorRestartRequired)
	}
	updatedIDs, destroyedIDs, err := stalwartNormalizeChangeIDs(changes)
	if err != nil {
		return EngineChangePage{}, err
	}
	if len(updatedIDs)+len(destroyedIDs) > request.Page.Limit {
		return EngineChangePage{}, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	changed := len(updatedIDs)+len(destroyedIDs) != 0
	if !changed {
		if *changes.HasMoreChanges || changes.NewState != changes.OldState {
			return EngineChangePage{}, stalwartEngineError(EngineErrorUnavailable)
		}
	} else if changes.NewState == changes.OldState {
		return EngineChangePage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	projections, getState, notFound, err := engine.stalwartFetchProjections(ctx, request.AccountID, updatedIDs)
	if err != nil {
		return EngineChangePage{}, err
	}
	if len(notFound) != 0 || getState != "" && getState != changes.NewState {
		return EngineChangePage{}, stalwartEngineError(EngineErrorRestartRequired)
	}
	nextCursor := ""
	if changed {
		nextCursor, err = stalwartPackCursor(stalwartMessageCursor{AccountID: request.AccountID, Kind: stalwartCursorChanges, State: changes.NewState})
		if err != nil {
			return EngineChangePage{}, err
		}
	}
	page := EngineChangePage{AccountID: request.AccountID, FromCursor: request.FromCursor, NextCursor: nextCursor, Bounds: EnginePageBounds{Requested: request.Page.Limit, Maximum: MaxEnginePageItems, Returned: len(projections) + len(destroyedIDs)}, CreatedOrUpdated: projections, DestroyedIDs: destroyedIDs}
	if err := validateChangePage(page, request); err != nil {
		return EngineChangePage{}, stalwartMappingResponseError(err)
	}
	return page, nil
}

// EnumerateProjection reads a stable base snapshot by detecting a JMAP
// queryState change between pages. A changed query state requires restart
// instead of silently mixing two mailbox snapshots.
func (engine *StalwartEngine) EnumerateProjection(ctx context.Context, request EngineEnumerationRequest) (EngineEnumerationPage, error) {
	if ctx == nil || !validAccountID(request.AccountID) {
		return EngineEnumerationPage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if err := validatePageRequest(request.Page); err != nil {
		return EngineEnumerationPage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	baseState := ""
	queryState := ""
	if request.BaseCursor == "" {
		if request.Checkpoint != "" {
			return EngineEnumerationPage{}, stalwartEngineError(EngineErrorInvalidCursor)
		}
		state, stateErr := engine.stalwartReadEmailState(ctx, request.AccountID)
		if stateErr != nil {
			return EngineEnumerationPage{}, stateErr
		}
		baseState = state
	} else {
		base, err := stalwartParseCursor(request.BaseCursor, request.AccountID, stalwartCursorChanges, "")
		if err != nil {
			return EngineEnumerationPage{}, err
		}
		baseState = base.State
	}
	position := 0
	if request.Checkpoint != "" {
		checkpoint, err := stalwartParseCursor(request.Checkpoint, request.AccountID, stalwartCursorEnumerationPage, "")
		if err != nil {
			return EngineEnumerationPage{}, err
		}
		if baseState == "" || checkpoint.BaseState != baseState {
			return EngineEnumerationPage{}, stalwartEngineError(EngineErrorInvalidCursor)
		}
		position = checkpoint.Position
		queryState = checkpoint.State
	}
	arguments, err := json.Marshal(stalwartEmailQueryArguments{
		AccountID:      request.AccountID,
		CalculateTotal: true,
		Limit:          request.Page.Limit,
		Position:       position,
		Sort: []stalwartEmailComparator{
			{Property: "receivedAt", IsAscending: false},
		},
	})
	if err != nil {
		return EngineEnumerationPage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, request.AccountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/query",
			Arguments: arguments,
			CallID:    stalwartEmailQueryCallID,
		}},
	}, false)
	if err != nil {
		return EngineEnumerationPage{}, err
	}
	query, err := stalwartParseEmailQuery(response, stalwartEmailQueryCallID, request.AccountID)
	if err != nil {
		return EngineEnumerationPage{}, err
	}
	if query.Position != position || len(query.IDs) > request.Page.Limit || queryState != "" && query.QueryState != queryState {
		return EngineEnumerationPage{}, stalwartEngineError(EngineErrorRestartRequired)
	}
	if len(query.IDs) > *query.Total-query.Position || len(query.IDs) == 0 && query.Position < *query.Total {
		return EngineEnumerationPage{}, stalwartEngineError(EngineErrorUnavailable)
	}
	baseCursor, err := stalwartPackCursor(stalwartMessageCursor{AccountID: request.AccountID, Kind: stalwartCursorChanges, State: baseState})
	if err != nil {
		return EngineEnumerationPage{}, err
	}
	projections, _, notFound, err := engine.stalwartFetchProjections(ctx, request.AccountID, query.IDs)
	if err != nil {
		return EngineEnumerationPage{}, err
	}
	if len(notFound) != 0 {
		return EngineEnumerationPage{}, stalwartEngineError(EngineErrorRestartRequired)
	}
	nextPosition := position + len(query.IDs)
	complete := nextPosition >= *query.Total
	nextCheckpoint := ""
	if !complete {
		nextCheckpoint, err = stalwartPackCursor(stalwartMessageCursor{AccountID: request.AccountID, BaseState: baseState, Kind: stalwartCursorEnumerationPage, Position: nextPosition, State: query.QueryState})
		if err != nil {
			return EngineEnumerationPage{}, err
		}
	}
	page := EngineEnumerationPage{AccountID: request.AccountID, BaseCursor: baseCursor, Checkpoint: request.Checkpoint, NextCheckpoint: nextCheckpoint, Complete: complete, Bounds: EnginePageBounds{Requested: request.Page.Limit, Maximum: MaxEnginePageItems, Returned: len(projections)}, Messages: projections}
	if err := validateEnumerationPage(page, request); err != nil {
		return EngineEnumerationPage{}, stalwartMappingResponseError(err)
	}
	return page, nil
}

func stalwartNormalizeChangeIDs(changes stalwartEmailChangesResponse) ([]string, []string, error) {
	if len(changes.Created)+len(changes.Updated)+len(changes.Destroyed) > MaxEnginePageItems {
		return nil, nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	updated := make(map[string]struct{}, len(changes.Created)+len(changes.Updated))
	destroyed := make(map[string]struct{}, len(changes.Destroyed))
	for _, id := range changes.Created {
		updated[id] = struct{}{}
	}
	for _, id := range changes.Updated {
		if _, duplicate := updated[id]; duplicate {
			continue
		}
		updated[id] = struct{}{}
	}
	for _, id := range changes.Destroyed {
		if _, collision := updated[id]; collision {
			return nil, nil, stalwartEngineError(EngineErrorUnavailable)
		}
		if _, duplicate := destroyed[id]; duplicate {
			continue
		}
		destroyed[id] = struct{}{}
	}
	updatedIDs := make([]string, 0, len(updated))
	for id := range updated {
		updatedIDs = append(updatedIDs, id)
	}
	destroyedIDs := make([]string, 0, len(destroyed))
	for id := range destroyed {
		destroyedIDs = append(destroyedIDs, id)
	}
	sort.Strings(updatedIDs)
	sort.Strings(destroyedIDs)
	return updatedIDs, destroyedIDs, nil
}
