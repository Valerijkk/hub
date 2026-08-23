package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

const (
	stalwartMailboxGetCallID       = "mailbox-get"
	stalwartMailboxCreateCallID    = "mailbox-create"
	stalwartMailboxRenameCallID    = "mailbox-rename"
	stalwartMailboxDestroyCallID   = "mailbox-destroy"
	stalwartMailboxCreateID        = "create-folder"
	stalwartMailboxCoreCapability  = "urn:ietf:params:jmap:core"
	stalwartMailboxMailCapability  = "urn:ietf:params:jmap:mail"
	stalwartMailboxSubmissionCapability = "urn:ietf:params:jmap:submission"
)

var stalwartMailboxCapabilities = []string{
	stalwartMailboxCoreCapability,
	stalwartMailboxMailCapability,
	stalwartMailboxSubmissionCapability,
}

var stalwartMailboxProperties = []string{
	"id",
	"name",
	"parentId",
	"role",
}

type stalwartMailboxGetArguments struct {
	AccountID  string   `json:"accountId"`
	Properties []string `json:"properties"`
}

type stalwartMailboxGetResponse struct {
	AccountID string                   `json:"accountId"`
	State     string                   `json:"state"`
	List      []stalwartMailboxRecord  `json:"list"`
	NotFound  []string                 `json:"notFound"`
}

type stalwartMailboxRecord struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ParentID *string `json:"parentId"`
	Role     *string `json:"role"`
}

type stalwartMailboxCreateArguments struct {
	AccountID  string                                `json:"accountId"`
	IfInState  string                                `json:"ifInState"`
	Create     map[string]stalwartMailboxCreatePatch `json:"create"`
}

type stalwartMailboxCreatePatch struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parentId,omitempty"`
}

type stalwartMailboxRenameArguments struct {
	AccountID string                                `json:"accountId"`
	IfInState string                                `json:"ifInState"`
	Update    map[string]stalwartMailboxRenamePatch `json:"update"`
}

type stalwartMailboxRenamePatch struct {
	Name string `json:"name"`
}

type stalwartMailboxDestroyArguments struct {
	AccountID               string   `json:"accountId"`
	IfInState               string   `json:"ifInState"`
	Destroy                 []string `json:"destroy"`
	OnDestroyRemoveEmails  bool     `json:"onDestroyRemoveEmails"`
}

type stalwartMailboxSetResponse struct {
	AccountID    string                                     `json:"accountId"`
	OldState     string                                     `json:"oldState"`
	NewState     string                                     `json:"newState"`
	Created      map[string]stalwartMailboxCreated           `json:"created"`
	NotCreated   map[string]stalwartMailboxSetFailure        `json:"notCreated"`
	Updated      map[string]json.RawMessage                  `json:"updated"`
	NotUpdated   map[string]stalwartMailboxSetFailure        `json:"notUpdated"`
	Destroyed    []string                                   `json:"destroyed"`
	NotDestroyed map[string]stalwartMailboxSetFailure        `json:"notDestroyed"`
}

type stalwartMailboxCreated struct {
	ID string `json:"id"`
}

type stalwartMailboxSetFailure struct {
	Type string `json:"type"`
}

type stalwartMailboxFolderRow struct {
	folder     EngineFolder
	nameKey    string
}

// ListFolders returns only the exact account-bound Mailbox/get snapshot. It
// never manufactures default folders when a provider account is empty.
func (engine *StalwartEngine) ListFolders(ctx context.Context, accountID string) ([]EngineFolder, error) {
	folders, _, err := engine.listFolderSnapshot(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return folders, nil
}

// CreateFolder creates one custom mailbox through one conditional Mailbox/set
// call. The preceding read only validates the parent and current tree; it is
// bound into the mutation by ifInState, so a concurrent change becomes a
// conflict rather than an unsafe write.
func (engine *StalwartEngine) CreateFolder(ctx context.Context, accountID string, spec FolderSpec) (EngineFolder, error) {
	if ctx == nil || !validAccountID(accountID) || !validOpaqueIDOrEmpty(spec.ParentID) {
		return EngineFolder{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	nameKey, err := normalizeFolderName(spec.Name)
	if err != nil {
		return EngineFolder{}, stalwartFolderPolicyError(err)
	}
	folders, state, err := engine.listFolderSnapshot(ctx, accountID)
	if err != nil {
		return EngineFolder{}, err
	}
	if len(folders) >= MaxEngineFolderItems {
		return EngineFolder{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	if spec.ParentID != "" && !stalwartFolderExists(folders, spec.ParentID) {
		return EngineFolder{}, stalwartEngineError(EngineErrorNotFound)
	}
	if stalwartCustomNameExists(folders, spec.ParentID, nameKey, "") {
		return EngineFolder{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}

	arguments, err := json.Marshal(stalwartMailboxCreateArguments{
		AccountID: accountID,
		IfInState: state,
		Create: map[string]stalwartMailboxCreatePatch{
			stalwartMailboxCreateID: {
				Name:     spec.Name,
				ParentID: stalwartOptionalFolderID(spec.ParentID),
			},
		},
	})
	if err != nil {
		return EngineFolder{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Mailbox/set",
			Arguments: arguments,
			CallID:    stalwartMailboxCreateCallID,
		}},
	}, true)
	if err != nil {
		return EngineFolder{}, err
	}
	set, err := parseStalwartMailboxSetResponse(response, stalwartMailboxCreateCallID, accountID, state)
	if err != nil {
		return EngineFolder{}, err
	}
	if err := stalwartMailboxOnlyCreateSucceeded(set); err != nil {
		return EngineFolder{}, err
	}
	created, exists := set.Created[stalwartMailboxCreateID]
	if !exists || !validOpaqueID(created.ID) {
		return EngineFolder{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return EngineFolder{
		AccountID: accountID,
		FolderID:  created.ID,
		ParentID:  spec.ParentID,
		Name:      spec.Name,
		Role:      EngineFolderRoleCustom,
	}, nil
}

// RenameFolder renames only an existing custom mailbox. System-role mailbox
// names are provider-owned and are intentionally protected from this generic
// custom-folder operation.
func (engine *StalwartEngine) RenameFolder(ctx context.Context, accountID, folderID, name string) error {
	if ctx == nil || !validAccountID(accountID) || !validOpaqueID(folderID) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	nameKey, err := normalizeFolderName(name)
	if err != nil {
		return stalwartFolderPolicyError(err)
	}
	folders, state, err := engine.listFolderSnapshot(ctx, accountID)
	if err != nil {
		return err
	}
	folder, exists := stalwartFolderByID(folders, folderID)
	if !exists {
		return stalwartEngineError(EngineErrorNotFound)
	}
	if folder.Role != EngineFolderRoleCustom || stalwartCustomNameExists(folders, folder.ParentID, nameKey, folderID) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}

	arguments, err := json.Marshal(stalwartMailboxRenameArguments{
		AccountID: accountID,
		IfInState: state,
		Update: map[string]stalwartMailboxRenamePatch{
			folderID: {Name: name},
		},
	})
	if err != nil {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Mailbox/set",
			Arguments: arguments,
			CallID:    stalwartMailboxRenameCallID,
		}},
	}, true)
	if err != nil {
		return err
	}
	set, err := parseStalwartMailboxSetResponse(response, stalwartMailboxRenameCallID, accountID, state)
	if err != nil {
		return err
	}
	return stalwartMailboxOnlyUpdateSucceeded(set, folderID)
}

// DeleteFolder destroys only a childless custom mailbox.
// onDestroyRemoveEmails is explicitly false, so this adapter never asks the
// provider to remove or destroy email data as a side effect of deleting a
// folder.
func (engine *StalwartEngine) DeleteFolder(ctx context.Context, accountID, folderID string) error {
	if ctx == nil || !validAccountID(accountID) || !validOpaqueID(folderID) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	folders, state, err := engine.listFolderSnapshot(ctx, accountID)
	if err != nil {
		return err
	}
	folder, exists := stalwartFolderByID(folders, folderID)
	if !exists {
		return stalwartEngineError(EngineErrorNotFound)
	}
	if folder.Role != EngineFolderRoleCustom || stalwartFolderHasChildren(folders, folderID) {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}

	arguments, err := json.Marshal(stalwartMailboxDestroyArguments{
		AccountID:              accountID,
		IfInState:              state,
		Destroy:                []string{folderID},
		OnDestroyRemoveEmails: false,
	})
	if err != nil {
		return stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Mailbox/set",
			Arguments: arguments,
			CallID:    stalwartMailboxDestroyCallID,
		}},
	}, true)
	if err != nil {
		return err
	}
	set, err := parseStalwartMailboxSetResponse(response, stalwartMailboxDestroyCallID, accountID, state)
	if err != nil {
		return err
	}
	return stalwartMailboxOnlyDestroySucceeded(set, folderID)
}

func (engine *StalwartEngine) listFolderSnapshot(ctx context.Context, accountID string) ([]EngineFolder, string, error) {
	if ctx == nil || !validAccountID(accountID) {
		return nil, "", stalwartEngineError(EngineErrorRejectedPolicy)
	}
	arguments, err := json.Marshal(stalwartMailboxGetArguments{
		AccountID:  accountID,
		Properties: append([]string(nil), stalwartMailboxProperties...),
	})
	if err != nil {
		return nil, "", stalwartEngineError(EngineErrorRejectedPolicy)
	}
	response, err := engine.callMailbox(ctx, accountID, stalwart.JMAPRequest{
		Using: stalwartMailboxCapabilities,
		MethodCalls: []stalwart.JMAPMethodCall{{
			Name:      "Mailbox/get",
			Arguments: arguments,
			CallID:    stalwartMailboxGetCallID,
		}},
	}, false)
	if err != nil {
		return nil, "", err
	}
	if len(response.MethodResponses) != 1 || response.MethodResponses[0].Name != "Mailbox/get" || response.MethodResponses[0].CallID != stalwartMailboxGetCallID {
		return nil, "", stalwartEngineError(EngineErrorUnavailable)
	}
	var payload stalwartMailboxGetResponse
	if err := json.Unmarshal(response.MethodResponses[0].Arguments, &payload); err != nil {
		return nil, "", stalwartEngineError(EngineErrorUnavailable)
	}
	if payload.AccountID != accountID {
		return nil, "", stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if !validOpaqueID(payload.State) {
		return nil, "", stalwartEngineError(EngineErrorUnavailable)
	}
	if len(payload.NotFound) != 0 {
		return nil, "", stalwartEngineError(EngineErrorUnavailable)
	}
	if len(payload.List) > MaxEngineFolderItems {
		return nil, "", stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	folders, err := validateStalwartFolderRecords(accountID, payload.List)
	if err != nil {
		return nil, "", err
	}
	return folders, payload.State, nil
}

func validateStalwartFolderRecords(accountID string, records []stalwartMailboxRecord) ([]EngineFolder, error) {
	if !validAccountID(accountID) || len(records) > MaxEngineFolderItems {
		return nil, stalwartEngineError(EngineErrorPayloadTooLarge)
	}
	rows := make([]stalwartMailboxFolderRow, 0, len(records))
	byID := make(map[string]stalwartMailboxFolderRow, len(records))
	roles := make(map[EngineFolderRole]string, 6)
	for _, record := range records {
		role, ok := stalwartFolderRole(record.Role)
		if !ok || !validOpaqueID(record.ID) {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		parentID := ""
		if record.ParentID != nil {
			parentID = *record.ParentID
			if !validOpaqueID(parentID) {
				return nil, stalwartEngineError(EngineErrorUnavailable)
			}
		}
		nameKey, err := normalizeFolderName(record.Name)
		if err != nil {
			return nil, stalwartFolderResponseError(err)
		}
		if role != EngineFolderRoleCustom && parentID != "" {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		folder := EngineFolder{AccountID: accountID, FolderID: record.ID, ParentID: parentID, Name: record.Name, Role: role}
		row := stalwartMailboxFolderRow{folder: folder, nameKey: nameKey}
		if _, duplicate := byID[record.ID]; duplicate {
			return nil, stalwartEngineError(EngineErrorUnavailable)
		}
		if role != EngineFolderRoleCustom {
			if _, duplicate := roles[role]; duplicate {
				return nil, stalwartEngineError(EngineErrorUnavailable)
			}
			roles[role] = record.ID
		}
		byID[record.ID] = row
		rows = append(rows, row)
	}
	if err := validateStalwartFolderParents(byID); err != nil {
		return nil, err
	}
	if err := validateStalwartFolderCustomNames(rows); err != nil {
		return nil, err
	}
	sort.Slice(rows, func(left, right int) bool {
		return compareStalwartMailboxFolderRows(rows[left], rows[right]) < 0
	})
	folders := make([]EngineFolder, len(rows))
	for index, row := range rows {
		folders[index] = row.folder
	}
	return folders, nil
}

func validateStalwartFolderParents(rows map[string]stalwartMailboxFolderRow) error {
	depths := make(map[string]int, len(rows))
	visiting := make(map[string]bool, len(rows))
	var depth func(string) (int, error)
	depth = func(id string) (int, error) {
		if value, ok := depths[id]; ok {
			return value, nil
		}
		if visiting[id] {
			return 0, stalwartEngineError(EngineErrorUnavailable)
		}
		row, exists := rows[id]
		if !exists {
			return 0, stalwartEngineError(EngineErrorUnavailable)
		}
		if row.folder.ParentID == "" {
			depths[id] = 0
			return 0, nil
		}
		if row.folder.ParentID == id {
			return 0, stalwartEngineError(EngineErrorUnavailable)
		}
		visiting[id] = true
		parentDepth, err := depth(row.folder.ParentID)
		delete(visiting, id)
		if err != nil {
			return 0, err
		}
		value := parentDepth + 1
		if value >= 3 {
			return 0, stalwartEngineError(EngineErrorUnavailable)
		}
		depths[id] = value
		return value, nil
	}
	for id := range rows {
		if _, err := depth(id); err != nil {
			return err
		}
	}
	return nil
}

func validateStalwartFolderCustomNames(rows []stalwartMailboxFolderRow) error {
	seen := make(map[string]map[string]struct{}, len(rows))
	for _, row := range rows {
		if row.folder.Role != EngineFolderRoleCustom {
			continue
		}
		byParent := seen[row.folder.ParentID]
		if byParent == nil {
			byParent = make(map[string]struct{})
			seen[row.folder.ParentID] = byParent
		}
		if _, duplicate := byParent[row.nameKey]; duplicate {
			return stalwartEngineError(EngineErrorUnavailable)
		}
		byParent[row.nameKey] = struct{}{}
	}
	return nil
}

func compareStalwartMailboxFolderRows(left, right stalwartMailboxFolderRow) int {
	if result := bytes.Compare([]byte(left.folder.ParentID), []byte(right.folder.ParentID)); result != 0 {
		return result
	}
	if leftRank, rightRank := stalwartFolderRoleRank(left.folder.Role), stalwartFolderRoleRank(right.folder.Role); leftRank < rightRank {
		return -1
	} else if leftRank > rightRank {
		return 1
	}
	if result := bytes.Compare([]byte(left.nameKey), []byte(right.nameKey)); result != 0 {
		return result
	}
	return bytes.Compare([]byte(left.folder.FolderID), []byte(right.folder.FolderID))
}

func stalwartFolderRole(value *string) (EngineFolderRole, bool) {
	if value == nil || *value == "" {
		return EngineFolderRoleCustom, true
	}
	switch *value {
	case "inbox":
		return EngineFolderRoleInbox, true
	case "sent":
		return EngineFolderRoleSent, true
	case "drafts":
		return EngineFolderRoleDrafts, true
	case "archive":
		return EngineFolderRoleArchive, true
	case "trash":
		return EngineFolderRoleTrash, true
	case "junk":
		return EngineFolderRoleJunk, true
	default:
		return "", false
	}
}

func stalwartFolderRoleRank(role EngineFolderRole) int {
	switch role {
	case EngineFolderRoleInbox:
		return 0
	case EngineFolderRoleSent:
		return 1
	case EngineFolderRoleDrafts:
		return 2
	case EngineFolderRoleArchive:
		return 3
	case EngineFolderRoleTrash:
		return 4
	case EngineFolderRoleJunk:
		return 5
	default:
		return 6
	}
}

func stalwartOptionalFolderID(value string) *string {
	if value == "" {
		return nil
	}
	copy := value
	return &copy
}

func stalwartFolderByID(folders []EngineFolder, folderID string) (EngineFolder, bool) {
	for _, folder := range folders {
		if folder.FolderID == folderID {
			return folder, true
		}
	}
	return EngineFolder{}, false
}

func stalwartFolderExists(folders []EngineFolder, folderID string) bool {
	_, exists := stalwartFolderByID(folders, folderID)
	return exists
}

func stalwartFolderHasChildren(folders []EngineFolder, folderID string) bool {
	for _, folder := range folders {
		if folder.ParentID == folderID {
			return true
		}
	}
	return false
}

func stalwartCustomNameExists(folders []EngineFolder, parentID, nameKey, excludedID string) bool {
	for _, folder := range folders {
		if folder.FolderID == excludedID || folder.Role != EngineFolderRoleCustom || folder.ParentID != parentID {
			continue
		}
		key, err := normalizeFolderName(folder.Name)
		if err == nil && key == nameKey {
			return true
		}
	}
	return false
}

func parseStalwartMailboxSetResponse(response stalwart.JMAPResponse, callID, accountID, expectedOldState string) (stalwartMailboxSetResponse, error) {
	if len(response.MethodResponses) != 1 || response.MethodResponses[0].Name != "Mailbox/set" || response.MethodResponses[0].CallID != callID {
		return stalwartMailboxSetResponse{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	var payload stalwartMailboxSetResponse
	if err := json.Unmarshal(response.MethodResponses[0].Arguments, &payload); err != nil {
		return stalwartMailboxSetResponse{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if payload.AccountID != accountID {
		return stalwartMailboxSetResponse{}, stalwartEngineError(EngineErrorUnauthorizedAccount)
	}
	if payload.OldState != expectedOldState || !validOpaqueID(payload.NewState) {
		return stalwartMailboxSetResponse{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return payload, nil
}

func stalwartMailboxOnlyCreateSucceeded(payload stalwartMailboxSetResponse) error {
	if len(payload.NotCreated) != 0 {
		return stalwartMailboxSetFailureFor(payload.NotCreated, stalwartMailboxCreateID)
	}
	if len(payload.Created) != 1 || len(payload.Updated) != 0 || len(payload.NotUpdated) != 0 || len(payload.Destroyed) != 0 || len(payload.NotDestroyed) != 0 {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	return nil
}

func stalwartMailboxOnlyUpdateSucceeded(payload stalwartMailboxSetResponse, folderID string) error {
	if len(payload.NotUpdated) != 0 {
		return stalwartMailboxSetFailureFor(payload.NotUpdated, folderID)
	}
	if len(payload.Updated) != 1 || !stalwartMailboxUpdateContains(payload.Updated, folderID) || len(payload.Created) != 0 || len(payload.NotCreated) != 0 || len(payload.Destroyed) != 0 || len(payload.NotDestroyed) != 0 {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	return nil
}

func stalwartMailboxOnlyDestroySucceeded(payload stalwartMailboxSetResponse, folderID string) error {
	if len(payload.NotDestroyed) != 0 {
		return stalwartMailboxSetFailureFor(payload.NotDestroyed, folderID)
	}
	if len(payload.Destroyed) != 1 || payload.Destroyed[0] != folderID || len(payload.Created) != 0 || len(payload.NotCreated) != 0 || len(payload.Updated) != 0 || len(payload.NotUpdated) != 0 {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	return nil
}

func stalwartMailboxUpdateContains(values map[string]json.RawMessage, folderID string) bool {
	value, exists := values[folderID]
	return exists && bytes.Equal(bytes.TrimSpace(value), []byte("null"))
}

func stalwartMailboxSetFailureFor(values map[string]stalwartMailboxSetFailure, target string) error {
	if len(values) != 1 {
		return stalwartEngineError(EngineErrorIndeterminate)
	}
	failure, exists := values[target]
	if !exists {
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
	case "invalidArguments", "invalidProperties", "invalidPatch", "willDestroy", "invalidResultReference", "unknownMethod", "unknownCapability", "unsupportedFilter", "unsupportedSort":
		return stalwartEngineError(EngineErrorRejectedPolicy)
	default:
		return stalwartEngineError(EngineErrorIndeterminate)
	}
}

func stalwartFolderPolicyError(err error) error {
	if err == nil {
		return nil
	}
	return stalwartEngineError(EngineErrorRejectedPolicy)
}

func stalwartFolderResponseError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*FolderPolicyError); ok {
		return stalwartEngineError(EngineErrorUnavailable)
	}
	return stalwartEngineError(EngineErrorUnavailable)
}
