package mail

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/json"
	"strings"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartEmailSetCallID         = "email-set"
	stalwartMutationMaxMessageIDs = 200
	stalwartHubKeywordPrefix      = "hubk-"
	stalwartHubKeywordLength      = 57
)

type stalwartEmailSetArguments struct {
	AccountID string                    `json:"accountId"`
	IfInState string                    `json:"ifInState"`
	Update    map[string]map[string]any `json:"update,omitempty"`
	Destroy   []string                  `json:"destroy,omitempty"`
}

type stalwartEmailSetResponse struct {
	AccountID       string                               `json:"accountId"`
	OldState        string                               `json:"oldState"`
	OldStatePresent bool                                 `json:"-"`
	OldStateKnown   bool                                 `json:"-"`
	NewState        string                               `json:"newState"`
	Created         map[string]json.RawMessage           `json:"created"`
	NotCreated      map[string]stalwartMailboxSetFailure `json:"notCreated"`
	Updated         map[string]json.RawMessage           `json:"updated"`
	NotUpdated      map[string]stalwartMailboxSetFailure `json:"notUpdated"`
	Destroyed       []string                             `json:"destroyed"`
	NotDestroyed    map[string]stalwartMailboxSetFailure `json:"notDestroyed"`
}

// MoveMessages replaces mailboxIds with one exact, prevalidated target. The
// Mailbox/get preflight and Email/set mutation remain separate JMAP calls.
func (engine *StalwartEngine) MoveMessages(ctx context.Context, accountID string, messageIDs []string, folderID string) error {
	if !validStalwartMutationInput(ctx, accountID, messageIDs) || !validOpaqueID(folderID) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	folders, _, err := engine.listFolderSnapshot(ctx, accountID)
	if err != nil {
		return err
	}
	if !stalwartFolderExists(folders, folderID) {
		return stalwartEngineError(EngineErrorNotFound)
	}
	patch := map[string]any{"mailboxIds": map[string]bool{folderID: true}}
	return engine.stalwartUpdateMessages(ctx, accountID, messageIDs, patch)
}

// SetFlags patches only the four parent-engine flags represented by standard
// JMAP keywords. Unspecified and unrelated keywords remain untouched.
func (engine *StalwartEngine) SetFlags(ctx context.Context, accountID string, messageIDs []string, patch FlagPatch) error {
	if !validStalwartMutationInput(ctx, accountID, messageIDs) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	update := make(map[string]any, 4)
	stalwartAddFlagPatch(update, "$seen", patch.Seen)
	stalwartAddFlagPatch(update, "$answered", patch.Answered)
	stalwartAddFlagPatch(update, "$flagged", patch.Flagged)
	stalwartAddFlagPatch(update, "$junk", patch.Junk)
	if len(update) == 0 {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return engine.stalwartUpdateMessages(ctx, accountID, messageIDs, update)
}

// SetKeywords accepts only canonical Hub label markers. Standard/system JMAP
// keywords are exclusively owned by SetFlags and cannot be confused with a
// label mutation.
func (engine *StalwartEngine) SetKeywords(ctx context.Context, accountID string, messageIDs []string, patch KeywordPatch) error {
	if !validStalwartMutationInput(ctx, accountID, messageIDs) || len(patch.Add)+len(patch.Remove) == 0 || len(patch.Add)+len(patch.Remove) > MaxEngineKeywordItems {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	update := make(map[string]any, len(patch.Add)+len(patch.Remove))
	seen := make(map[string]bool, len(patch.Add)+len(patch.Remove))
	for _, marker := range patch.Add {
		keyword := string(marker)
		if !validStalwartHubKeyword(keyword) || seen[keyword] {
			return stalwartEngineError(EngineErrorRejectedPolicy)
		}
		seen[keyword] = true
		update["keywords/"+keyword] = true
	}
	for _, marker := range patch.Remove {
		keyword := string(marker)
		if !validStalwartHubKeyword(keyword) || seen[keyword] {
			return stalwartEngineError(EngineErrorRejectedPolicy)
		}
		seen[keyword] = true
		update["keywords/"+keyword] = nil
	}
	return engine.stalwartUpdateMessages(ctx, accountID, messageIDs, update)
}

// DeleteMessages destroys exact messages permanently or moves them to the one
// authoritative trash-role mailbox resolved in a separate read.
func (engine *StalwartEngine) DeleteMessages(ctx context.Context, accountID string, messageIDs []string, permanent bool) error {
	if !validStalwartMutationInput(ctx, accountID, messageIDs) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if permanent {
		return engine.stalwartDestroyMessages(ctx, accountID, messageIDs)
	}
	folders, _, err := engine.listFolderSnapshot(ctx, accountID)
	if err != nil {
		return err
	}
	trashID := ""
	for _, folder := range folders {
		if folder.Role != EngineFolderRoleTrash {
			continue
		}
		if trashID != "" {
			return stalwartEngineError(EngineErrorUnavailable)
		}
		trashID = folder.FolderID
	}
	if trashID == "" {
		return stalwartEngineError(EngineErrorNotFound)
	}
	return engine.stalwartUpdateMessages(ctx, accountID, messageIDs, map[string]any{"mailboxIds": map[string]bool{trashID: true}})
}

func (engine *StalwartEngine) stalwartUpdateMessages(ctx context.Context, accountID string, messageIDs []string, patch map[string]any) error {
	if !validStalwartMutationInput(ctx, accountID, messageIDs) || len(patch) == 0 {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	state, err := engine.stalwartReadEmailState(ctx, accountID)
	if err != nil {
		return err
	}
	updates := make(map[string]map[string]any, len(messageIDs))
	for _, messageID := range messageIDs {
		messagePatch := make(map[string]any, len(patch))
		for path, value := range patch {
			messagePatch[path] = value
		}
		updates[messageID] = messagePatch
	}
	response, err := engine.stalwartEmailSet(ctx, accountID, stalwartEmailSetArguments{AccountID: accountID, IfInState: state, Update: updates})
	if err != nil {
		return err
	}
	set, err := parseStalwartEmailSetResponse(response, accountID, state)
	if err != nil {
		return err
	}
	return stalwartEmailOnlyUpdateSucceeded(set, messageIDs)
}

func (engine *StalwartEngine) stalwartDestroyMessages(ctx context.Context, accountID string, messageIDs []string) error {
	if !validStalwartMutationInput(ctx, accountID, messageIDs) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	state, err := engine.stalwartReadEmailState(ctx, accountID)
	if err != nil {
		return err
	}
	response, err := engine.stalwartEmailSet(ctx, accountID, stalwartEmailSetArguments{AccountID: accountID, IfInState: state, Destroy: append([]string(nil), messageIDs...)})
	if err != nil {
		return err
	}
	set, err := parseStalwartEmailSetResponse(response, accountID, state)
	if err != nil {
		return err
	}
	return stalwartEmailOnlyDestroySucceeded(set, messageIDs)
}

func (engine *StalwartEngine) stalwartEmailSet(ctx context.Context, accountID string, arguments stalwartEmailSetArguments) (stalwart.JMAPResponse, error) {
	payload, err := json.Marshal(arguments)
	if err != nil {
		return stalwart.JMAPResponse{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	return engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Email/set",
			Arguments: payload,
			CallID:    stalwartEmailSetCallID,
		}},
	}, true)
}

func parseStalwartEmailSetResponse(response stalwart.JMAPResponse, accountID, expectedOldState string) (stalwartEmailSetResponse, error) {
	if len(response.MethodResponses) != 1 || response.MethodResponses[0].Name != "Email/set" || response.MethodResponses[0].CallID != stalwartEmailSetCallID {
		return stalwartEmailSetResponse{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	var wire struct {
		AccountID    string                               `json:"accountId"`
		OldState     json.RawMessage                      `json:"oldState"`
		NewState     string                               `json:"newState"`
		Created      map[string]json.RawMessage           `json:"created"`
		NotCreated   map[string]stalwartMailboxSetFailure `json:"notCreated"`
		Updated      map[string]json.RawMessage           `json:"updated"`
		NotUpdated   map[string]stalwartMailboxSetFailure `json:"notUpdated"`
		Destroyed    []string                             `json:"destroyed"`
		NotDestroyed map[string]stalwartMailboxSetFailure `json:"notDestroyed"`
	}
	if err := json.Unmarshal(response.MethodResponses[0].Arguments, &wire); err != nil {
		return stalwartEmailSetResponse{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if wire.AccountID != accountID {
		return stalwartEmailSetResponse{}, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(wire.NewState) {
		return stalwartEmailSetResponse{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	oldState, oldStatePresent, oldStateKnown, err := stalwartDecodeSetOldState(wire.OldState, expectedOldState)
	if err != nil {
		return stalwartEmailSetResponse{}, err
	}
	payload := stalwartEmailSetResponse{
		AccountID:       wire.AccountID,
		OldState:        oldState,
		OldStatePresent: oldStatePresent,
		OldStateKnown:   oldStateKnown,
		NewState:        wire.NewState,
		Created:         wire.Created,
		NotCreated:      wire.NotCreated,
		Updated:         wire.Updated,
		NotUpdated:      wire.NotUpdated,
		Destroyed:       wire.Destroyed,
		NotDestroyed:    wire.NotDestroyed,
	}
	return payload, nil
}

func stalwartDecodeSetOldState(raw json.RawMessage, expected string) (string, bool, bool, error) {
	if len(raw) == 0 {
		return "", false, false, nil
	}
	trimmed := bytes.TrimSpace(raw)
	if bytes.Equal(trimmed, []byte("null")) {
		return "", true, false, nil
	}
	var value string
	if err := json.Unmarshal(trimmed, &value); err != nil || value != expected {
		return "", true, false, stalwartEngineError(EngineErrorIndeterminate)
	}
	return value, true, true, nil
}

func stalwartEmailOnlyUpdateSucceeded(payload stalwartEmailSetResponse, messageIDs []string) error {
	if len(payload.Created) != 0 || len(payload.NotCreated) != 0 || len(payload.Destroyed) != 0 || len(payload.NotDestroyed) != 0 {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	if len(payload.NotUpdated) != 0 {
		if len(payload.Updated) != 0 || payload.OldStateKnown && payload.NewState != payload.OldState {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		return stalwartEmailSetFailuresFor(payload.NotUpdated, messageIDs)
	}
	if !stalwartEmailUpdatedExactly(payload.Updated, messageIDs) {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	return nil
}

func stalwartEmailOnlyDestroySucceeded(payload stalwartEmailSetResponse, messageIDs []string) error {
	if len(payload.Created) != 0 || len(payload.NotCreated) != 0 || len(payload.Updated) != 0 || len(payload.NotUpdated) != 0 {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	if len(payload.NotDestroyed) != 0 {
		if len(payload.Destroyed) != 0 || payload.OldStateKnown && payload.NewState != payload.OldState {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		return stalwartEmailSetFailuresFor(payload.NotDestroyed, messageIDs)
	}
	if !payload.OldStateKnown || payload.NewState == payload.OldState || !stalwartEmailDestroyedExactly(payload.Destroyed, messageIDs) {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	return nil
}

func stalwartEmailUpdatedExactly(updated map[string]json.RawMessage, messageIDs []string) bool {
	if len(updated) != len(messageIDs) {
		return false
	}
	for _, messageID := range messageIDs {
		value, exists := updated[messageID]
		if !exists || !stalwartValidUpdatedValue(value) {
			return false
		}
	}
	return true
}

func stalwartValidUpdatedValue(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	if bytes.Equal(trimmed, []byte("null")) {
		return true
	}
	if len(trimmed) < 2 || trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return false
	}
	var object map[string]json.RawMessage
	return json.Unmarshal(trimmed, &object) == nil && object != nil
}

func stalwartEmailDestroyedExactly(destroyed, messageIDs []string) bool {
	if len(destroyed) != len(messageIDs) {
		return false
	}
	expected := make(map[string]struct{}, len(messageIDs))
	for _, messageID := range messageIDs {
		expected[messageID] = struct{}{}
	}
	for _, messageID := range destroyed {
		if _, exists := expected[messageID]; !exists {
			return false
		}
		delete(expected, messageID)
	}
	return len(expected) == 0
}

func stalwartEmailSetFailuresFor(failures map[string]stalwartMailboxSetFailure, messageIDs []string) error {
	if len(failures) != len(messageIDs) {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	failureType := ""
	for _, messageID := range messageIDs {
		failure, exists := failures[messageID]
		if !exists || failure.Type == "" || failureType != "" && failure.Type != failureType {
			return stalwartEngineError(EngineErrorIndeterminate)
		}
		failureType = failure.Type
	}
	switch failureType {
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

func validStalwartMutationInput(ctx context.Context, accountID string, messageIDs []string) bool {
	return ctx != nil && validAccountID(accountID) && len(messageIDs) > 0 && len(messageIDs) <= stalwartMutationMaxMessageIDs && validOpaqueIDs(messageIDs)
}

func stalwartAddFlagPatch(patch map[string]any, keyword string, value FlagValue) {
	if !value.Present {
		return
	}
	path := "keywords/" + keyword
	if value.Value {
		patch[path] = true
		return
	}
	patch[path] = nil
}

func validStalwartHubKeyword(value string) bool {
	if len(value) != stalwartHubKeywordLength || !strings.HasPrefix(value, stalwartHubKeywordPrefix) {
		return false
	}
	suffix := value[len(stalwartHubKeywordPrefix):]
	for _, character := range suffix {
		if (character < 'a' || character > 'z') && (character < '2' || character > '7') {
			return false
		}
	}
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(suffix))
	return err == nil && len(decoded) == 32 && strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(decoded)) == suffix
}
