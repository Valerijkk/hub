package mail

import (
	"context"
	"crypto/sha256"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// FakeMailEngineConfig supplies deterministic dependencies for FakeMailEngine.
type FakeMailEngineConfig struct {
	// NextID is serialized and invoked without the fake state mutex. It MUST NOT
	// re-enter source-consuming operations (EnsureDomain, EnsureAccount,
	// CreateFolder, SaveDraft, or Submit), which would recursively need the
	// source mutex. Source-free methods such as Health may be re-entered.
	NextID func() string
	// Now follows the same serialized callback and re-entry contract as NextID.
	Now func() time.Time
}

// FakeMailOperation identifies one dispatch point which can be scripted to fail.
type FakeMailOperation string

const (
	FakeMailOperationCapabilities         FakeMailOperation = "capabilities"
	FakeMailOperationGetDomain            FakeMailOperation = "get_domain"
	FakeMailOperationEnsureDomain         FakeMailOperation = "ensure_domain"
	FakeMailOperationEnsureAccount        FakeMailOperation = "ensure_account"
	FakeMailOperationListFolders          FakeMailOperation = "list_folders"
	FakeMailOperationCreateFolder         FakeMailOperation = "create_folder"
	FakeMailOperationRenameFolder         FakeMailOperation = "rename_folder"
	FakeMailOperationDeleteFolder         FakeMailOperation = "delete_folder"
	FakeMailOperationListMessages         FakeMailOperation = "list_messages"
	FakeMailOperationGetMessage           FakeMailOperation = "get_message"
	FakeMailOperationGetDraft             FakeMailOperation = "get_draft"
	FakeMailOperationGetBlob              FakeMailOperation = "get_blob"
	FakeMailOperationReadCalendarPart     FakeMailOperation = "read_calendar_part"
	FakeMailOperationChanges              FakeMailOperation = "changes"
	FakeMailOperationEnumerateProjection  FakeMailOperation = "enumerate_projection"
	FakeMailOperationSaveDraft            FakeMailOperation = "save_draft"
	FakeMailOperationReconcileDraft       FakeMailOperation = "reconcile_draft"
	FakeMailOperationDeleteDraft          FakeMailOperation = "delete_draft"
	FakeMailOperationReconcileDraftDeletion FakeMailOperation = "reconcile_draft_deletion"
	FakeMailOperationSubmit               FakeMailOperation = "submit"
	FakeMailOperationReconcileSubmission  FakeMailOperation = "reconcile_submission"
	FakeMailOperationMoveMessages         FakeMailOperation = "move_messages"
	FakeMailOperationSetFlags             FakeMailOperation = "set_flags"
	FakeMailOperationSetKeywords          FakeMailOperation = "set_keywords"
	FakeMailOperationDeleteMessages       FakeMailOperation = "delete_messages"
	FakeMailOperationGetSieve             FakeMailOperation = "get_sieve"
	FakeMailOperationUpsertSieve          FakeMailOperation = "upsert_sieve"
	FakeMailOperationClassifyMessages     FakeMailOperation = "classify_messages"
	FakeMailOperationHealth               FakeMailOperation = "health"
)

// FakeMailFault is a safe typed dispatch failure. AfterDispatch models a lost response.
type FakeMailFault struct {
	Operation     FakeMailOperation
	Kind          EngineErrorKind
	AfterDispatch bool
}

// FakeMailChangeScript supplies one deterministic incremental response for a cursor.
type FakeMailChangeScript struct {
	AccountID  string
	FromCursor string
	Page       EngineChangePage
	Fault      EngineErrorKind
}

// FakeMailEnumerationScript supplies one deterministic base enumeration response.
type FakeMailEnumerationScript struct {
	AccountID  string
	BaseCursor string
	Checkpoint string
	Page       EngineEnumerationPage
	Fault      EngineErrorKind
}

type fakeDraft struct {
	message  EngineMessage
	identity fakeDraftIdentity
}

// FakeCalendarPart seeds one deterministic private calendar-part read.
type FakeCalendarPart struct {
	AccountID   string
	MessageID   string
	Version     string
	LocatorHash [sha256.Size]byte
	Method      MailCalendarMethod
	ITIP        []byte
}

type fakeCalendarPartKey struct {
	accountID   string
	messageID   string
	locatorHash [sha256.Size]byte
}

type fakeCalendarPart struct {
	version string
	method  MailCalendarMethod
	itip    []byte
}

type fakeDraftIdentity struct {
	messageID       string
	previousVersion string
	mutationID      string
}

type fakeSubmission struct {
	accepted        EngineSubmissionAccepted
	draftKey        string
	messageID       string
	expectedVersion string
}

type fakeAccountCursorKey struct {
	accountID string
	cursor    string
}

type fakeEnumerationKey struct {
	accountID  string
	baseCursor string
	checkpoint string
}

type fakeChange struct {
	projection *EngineMessageProjection
	destroyed  string
}

type fakeEnumerationSnapshot struct {
	accountID string
	messages  []EngineMessageProjection
}

type fakeCheckpoint struct {
	accountID string
	base      string
	offset    int
}

// FakeMailEngine is an in-memory deterministic MailEngine implementation for unit tests.
type FakeMailEngine struct {
	mu       sync.Mutex
	sourceMu sync.Mutex

	nextID func() string
	now    func() time.Time
	serial int

	domains      map[string]EngineDomainState
	domainKeys   map[string]string
	accounts     map[string]EngineAccount
	accountKeys  map[string]string
	folders      map[string]map[string]EngineFolder
	messages     map[string]map[string]EngineMessage
	blobs        map[string]map[string]EngineBlob
	calendarParts map[fakeCalendarPartKey]fakeCalendarPart
	drafts       map[string]map[string]fakeDraft
	submissions  map[string]map[string]fakeSubmission
	sieves       map[string]EngineSieveState
	spamResults  map[fakeAccountCursorKey]EngineSpamTrainingResult
	faults       []FakeMailFault
	changes      map[string][]fakeChange
	changeScript map[fakeAccountCursorKey]FakeMailChangeScript
	enumScript   map[fakeEnumerationKey]FakeMailEnumerationScript
	overrides    map[fakeAccountCursorKey]EngineMessageProjection
	snapshots    map[fakeAccountCursorKey]fakeEnumerationSnapshot
	checkpoints  map[fakeEnumerationKey]fakeCheckpoint
	issuedIDs    map[string]struct{}
}

// NewFakeMailEngine constructs a deterministic fake without global clocks or randomness.
func NewFakeMailEngine(config FakeMailEngineConfig) *FakeMailEngine {
	nextID := config.NextID
	if nextID == nil {
		var sequence uint64
		nextID = func() string {
			return "fake-" + strconv.FormatUint(atomic.AddUint64(&sequence, 1), 10)
		}
	}
	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Unix(0, 0).UTC() }
	}
	return &FakeMailEngine{
		nextID:       nextID,
		now:          now,
		domains:      make(map[string]EngineDomainState),
		domainKeys:   make(map[string]string),
		accounts:     make(map[string]EngineAccount),
		accountKeys:  make(map[string]string),
		folders:      make(map[string]map[string]EngineFolder),
		messages:     make(map[string]map[string]EngineMessage),
		blobs:        make(map[string]map[string]EngineBlob),
		calendarParts: make(map[fakeCalendarPartKey]fakeCalendarPart),
		drafts:       make(map[string]map[string]fakeDraft),
		submissions:  make(map[string]map[string]fakeSubmission),
		sieves:       make(map[string]EngineSieveState),
		spamResults:  make(map[fakeAccountCursorKey]EngineSpamTrainingResult),
		changeScript: make(map[fakeAccountCursorKey]FakeMailChangeScript),
		enumScript:   make(map[fakeEnumerationKey]FakeMailEnumerationScript),
		overrides:    make(map[fakeAccountCursorKey]EngineMessageProjection),
		snapshots:    make(map[fakeAccountCursorKey]fakeEnumerationSnapshot),
		checkpoints:  make(map[fakeEnumerationKey]fakeCheckpoint),
		issuedIDs:    make(map[string]struct{}),
	}
}

// EnqueueFault appends one typed failure consumed by the next matching dispatch.
func (f *FakeMailEngine) EnqueueFault(fault FakeMailFault) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.faults = append(f.faults, fault)
}

// PutMessage seeds an account-scoped message by value.
func (f *FakeMailEngine) PutMessage(message EngineMessage) {
	if validateEngineMessage(message) != nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, used := f.issuedIDs[message.Projection.MessageID]; used {
		return
	}
	if _, used := f.issuedIDs[message.Projection.Version]; used {
		return
	}
	f.issuedIDs[message.Projection.MessageID] = struct{}{}
	f.issuedIDs[message.Projection.Version] = struct{}{}
	f.putMessageLocked(message, true)
}

// PutBlob seeds an account-scoped blob by value.
func (f *FakeMailEngine) PutBlob(blob EngineBlob) {
	if !validAccountID(blob.AccountID) || !validOpaqueID(blob.BlobID) || !validSafeText(blob.MediaType, MaxEngineMIMEFieldBytes, false) || blob.SizeBytes != int64(len(blob.Bytes)) || blob.SizeBytes < 0 || blob.SizeBytes > MaxEngineAttachmentBytes {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, used := f.issuedIDs[blob.BlobID]; used {
		return
	}
	f.issuedIDs[blob.BlobID] = struct{}{}
	if f.blobs[blob.AccountID] == nil {
		f.blobs[blob.AccountID] = make(map[string]EngineBlob)
	}
	f.blobs[blob.AccountID][blob.BlobID] = cloneBlob(blob)
}

// PutCalendarPart seeds one hash-addressed decoded iTIP part by value.
func (f *FakeMailEngine) PutCalendarPart(part FakeCalendarPart) {
	if !validFakeCalendarPart(part) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calendarParts[fakeCalendarPartKey{
		accountID:   part.AccountID,
		messageID:   part.MessageID,
		locatorHash: part.LocatorHash,
	}] = fakeCalendarPart{
		version: part.Version,
		method:  part.Method,
		itip:    append([]byte(nil), part.ITIP...),
	}
}

// ScriptChanges installs a deterministic response for one account and from cursor.
func (f *FakeMailEngine) ScriptChanges(script FakeMailChangeScript) {
	if !validAccountID(script.AccountID) || !validCursor(script.FromCursor, false) {
		return
	}
	if script.Fault == "" || len(script.Page.CreatedOrUpdated) != 0 || len(script.Page.DestroyedIDs) != 0 {
		if validateChangePage(script.Page, EngineChangeRequest{AccountID: script.AccountID, FromCursor: script.FromCursor, Page: EnginePageRequest{Limit: script.Page.Bounds.Requested}}) != nil {
			return
		}
	}
	for _, projection := range script.Page.CreatedOrUpdated {
		if projection.AccountID != script.AccountID {
			return
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	script.Page = cloneChangePage(script.Page)
	f.changeScript[fakeAccountCursorKey{accountID: script.AccountID, cursor: script.FromCursor}] = script
}

// ScriptEnumeration installs a deterministic response for one account/base/checkpoint tuple.
func (f *FakeMailEngine) ScriptEnumeration(script FakeMailEnumerationScript) {
	if !validAccountID(script.AccountID) || !validCursor(script.BaseCursor, true) || !validCursor(script.Checkpoint, true) {
		return
	}
	if script.Fault == "" || len(script.Page.Messages) != 0 {
		if validateEnumerationPage(script.Page, EngineEnumerationRequest{AccountID: script.AccountID, BaseCursor: script.BaseCursor, Checkpoint: script.Checkpoint, Page: EnginePageRequest{Limit: script.Page.Bounds.Requested}}) != nil {
			return
		}
	}
	for _, projection := range script.Page.Messages {
		if projection.AccountID != script.AccountID {
			return
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	script.Page = cloneEnumerationPage(script.Page)
	f.enumScript[fakeEnumerationKey{accountID: script.AccountID, baseCursor: script.BaseCursor, checkpoint: script.Checkpoint}] = script
}

// ScriptMessageProjection replaces a matching read projection while retaining seeded content.
func (f *FakeMailEngine) ScriptMessageProjection(accountID, messageID string, projection EngineMessageProjection) {
	if !validAccountID(accountID) || !validOpaqueID(messageID) || validateProjection(projection) != nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.overrides[fakeAccountCursorKey{accountID: accountID, cursor: messageID}] = cloneProjection(projection)
}

func (f *FakeMailEngine) Capabilities(_ context.Context) (EngineCapabilities, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationCapabilities); ok {
		return EngineCapabilities{}, fakeEngineError(fault.Kind)
	}
	return EngineCapabilities{ExclusiveCustomFolderWriter: true, StableProjectionEnumeration: true, SubmissionReconciliation: true, DraftReconciliation: true}, nil
}

func (f *FakeMailEngine) GetDomain(_ context.Context, domain string) (EngineDomainState, error) {
	if !validSafeText(domain, MaxEngineHeaderValueBytes, true) {
		return EngineDomainState{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationGetDomain); ok {
		return EngineDomainState{}, fakeEngineError(fault.Kind)
	}
	if id, ok := f.domainKeys[domain]; ok {
		return cloneDomain(f.domains[id]), nil
	}
	return EngineDomainState{State: EngineDomainAbsent, DNSRecords: []EngineDNSRecord{}}, nil
}

func (f *FakeMailEngine) EnsureDomain(_ context.Context, spec DomainSpec) (EngineDomainState, error) {
	if !validSafeText(spec.ProvisioningKey, MaxEngineOpaqueIDBytes, true) || !validSafeText(spec.Domain, MaxEngineHeaderValueBytes, true) || !validSafeText(spec.DesiredHash, MaxEngineOpaqueIDBytes, false) {
		return EngineDomainState{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationEnsureDomain)
	if faulted && !fault.AfterDispatch {
		f.mu.Unlock()
		return EngineDomainState{}, fakeEngineError(fault.Kind)
	}
	if existing := f.domainKeys[spec.Domain]; existing != "" {
		state := domainStateForSpec(existing, spec)
		f.domains[existing] = state
		f.mu.Unlock()
		if faulted && fault.AfterDispatch {
			return EngineDomainState{}, fakeEngineError(fault.Kind)
		}
		return cloneDomain(state), nil
	}
	f.mu.Unlock()

	id, err := f.nextOpaqueID()
	if err != nil {
		return EngineDomainState{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if existing := f.domainKeys[spec.Domain]; existing != "" {
		state := domainStateForSpec(existing, spec)
		f.domains[existing] = state
		if faulted && fault.AfterDispatch {
			return EngineDomainState{}, fakeEngineError(fault.Kind)
		}
		return cloneDomain(state), nil
	}
	if err := f.reserveOpaqueIDsLocked(id); err != nil {
		return EngineDomainState{}, err
	}
	f.domainKeys[spec.Domain] = id
	state := domainStateForSpec(id, spec)
	f.domains[id] = state
	if faulted && fault.AfterDispatch {
		return EngineDomainState{}, fakeEngineError(fault.Kind)
	}
	return cloneDomain(state), nil
}

func (f *FakeMailEngine) EnsureAccount(_ context.Context, spec AccountSpec) (EngineAccount, error) {
	if !validSafeText(spec.ProvisioningKey, MaxEngineOpaqueIDBytes, true) || !validSafeText(spec.LocalPart, MaxEngineAddressEmailBytes, true) || !validSafeText(spec.Domain, MaxEngineHeaderValueBytes, true) || !validSafeText(spec.DesiredHash, MaxEngineOpaqueIDBytes, false) || len(spec.Aliases) > MaxEngineAddressItems {
		return EngineAccount{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	if spec.ProvisioningKey == "" || spec.LocalPart == "" || spec.Domain == "" {
		return EngineAccount{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationEnsureAccount)
	if faulted && !fault.AfterDispatch {
		f.mu.Unlock()
		return EngineAccount{}, fakeEngineError(fault.Kind)
	}
	if accountID := f.accountKeys[spec.ProvisioningKey]; accountID != "" {
		account := cloneAccount(f.accounts[accountID])
		f.mu.Unlock()
		if faulted && fault.AfterDispatch {
			return EngineAccount{}, fakeEngineError(fault.Kind)
		}
		return account, nil
	}
	f.mu.Unlock()

	accountID, err := f.nextOpaqueID()
	if err != nil {
		return EngineAccount{}, err
	}
	primaryID, aliases, err := accountIdentityIDs(accountID, len(spec.Aliases))
	if err != nil {
		return EngineAccount{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if existingID := f.accountKeys[spec.ProvisioningKey]; existingID != "" {
		if faulted && fault.AfterDispatch {
			return EngineAccount{}, fakeEngineError(fault.Kind)
		}
		return cloneAccount(f.accounts[existingID]), nil
	}
	reservedIDs := append([]string{accountID, primaryID}, aliases...)
	if err := f.reserveOpaqueIDsLocked(reservedIDs...); err != nil {
		return EngineAccount{}, err
	}
	state := EngineAccount{AccountID: accountID, State: EngineAccountActive, AppliedGeneration: spec.DesiredGeneration, AppliedHash: spec.DesiredHash, PrimaryIdentityID: primaryID, AliasIdentityIDs: aliases}
	if !spec.Enabled {
		state.State = EngineAccountSuspended
	}
	f.accounts[accountID] = state
	f.accountKeys[spec.ProvisioningKey] = accountID
	if faulted && fault.AfterDispatch {
		return EngineAccount{}, fakeEngineError(fault.Kind)
	}
	return cloneAccount(state), nil
}

func (f *FakeMailEngine) ListFolders(_ context.Context, accountID string) ([]EngineFolder, error) {
	if !validAccountID(accountID) {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationListFolders); ok {
		return nil, fakeEngineError(fault.Kind)
	}
	folders := f.folders[accountID]
	result := make([]EngineFolder, 0, len(folders))
	for _, folder := range folders {
		result = append(result, folder)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].FolderID < result[right].FolderID })
	return result, nil
}

func (f *FakeMailEngine) CreateFolder(_ context.Context, accountID string, spec FolderSpec) (EngineFolder, error) {
	if !validAccountID(accountID) || !validSafeText(spec.Name, MaxEngineMIMEFieldBytes, true) || !validOpaqueIDOrEmpty(spec.ParentID) {
		return EngineFolder{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationCreateFolder)
	if faulted && !fault.AfterDispatch {
		f.mu.Unlock()
		return EngineFolder{}, fakeEngineError(fault.Kind)
	}
	if spec.ParentID != "" {
		if _, ok := f.folders[accountID][spec.ParentID]; !ok {
			f.mu.Unlock()
			return EngineFolder{}, fakeEngineError(EngineErrorNotFound)
		}
	}
	f.mu.Unlock()

	folderID, err := f.nextOpaqueID()
	if err != nil {
		return EngineFolder{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if spec.ParentID != "" {
		if _, ok := f.folders[accountID][spec.ParentID]; !ok {
			return EngineFolder{}, fakeEngineError(EngineErrorNotFound)
		}
	}
	if err := f.reserveOpaqueIDsLocked(folderID); err != nil {
		return EngineFolder{}, err
	}
	folder := EngineFolder{AccountID: accountID, FolderID: folderID, ParentID: spec.ParentID, Name: spec.Name, Role: EngineFolderRoleCustom}
	if f.folders[accountID] == nil {
		f.folders[accountID] = make(map[string]EngineFolder)
	}
	f.folders[accountID][folderID] = folder
	if faulted && fault.AfterDispatch {
		return EngineFolder{}, fakeEngineError(fault.Kind)
	}
	return folder, nil
}

func (f *FakeMailEngine) RenameFolder(_ context.Context, accountID, folderID, name string) error {
	if !validAccountID(accountID) || !validOpaqueID(folderID) || !validSafeText(name, MaxEngineMIMEFieldBytes, true) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationRenameFolder)
	if faulted && !fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	folder, ok := f.folders[accountID][folderID]
	if !ok {
		return fakeEngineError(EngineErrorNotFound)
	}
	folder.Name = name
	f.folders[accountID][folderID] = folder
	if faulted && fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	return nil
}

func (f *FakeMailEngine) DeleteFolder(_ context.Context, accountID, folderID string) error {
	if !validAccountID(accountID) || !validOpaqueID(folderID) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationDeleteFolder)
	if faulted && !fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	if _, ok := f.folders[accountID][folderID]; !ok {
		return fakeEngineError(EngineErrorNotFound)
	}
	delete(f.folders[accountID], folderID)
	if faulted && fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	return nil
}

func (f *FakeMailEngine) ListMessages(_ context.Context, query EngineMessageQuery) (EngineProjectionPage, error) {
	if !validAccountID(query.AccountID) || !validOpaqueIDOrEmpty(query.FolderID) || !validCursor(query.Cursor, true) {
		return EngineProjectionPage{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationListMessages); ok {
		return EngineProjectionPage{}, fakeEngineError(fault.Kind)
	}
	if err := validatePageRequest(query.Page); err != nil {
		return EngineProjectionPage{}, err
	}
	start, err := parseListCursor(query.Cursor)
	if err != nil {
		return EngineProjectionPage{}, err
	}
	projections := f.sortedProjectionsLocked(query.AccountID, query.FolderID)
	if start > len(projections) {
		return EngineProjectionPage{}, fakeEngineError(EngineErrorInvalidCursor)
	}
	end := start + query.Page.Limit
	if end > len(projections) {
		end = len(projections)
	}
	next := ""
	if end < len(projections) {
		next = "list:" + strconv.Itoa(end)
	}
	page := EngineProjectionPage{AccountID: query.AccountID, Cursor: "list:" + strconv.Itoa(start), NextCursor: next, Bounds: EnginePageBounds{Requested: query.Page.Limit, Maximum: MaxEnginePageItems, Returned: end - start}, Messages: cloneProjections(projections[start:end])}
	if err := validateProjectionPage(page); err != nil {
		return EngineProjectionPage{}, err
	}
	return page, nil
}

func (f *FakeMailEngine) GetMessage(_ context.Context, request EngineMessageReadRequest) (EngineMessage, error) {
	if !validAccountID(request.AccountID) || !validOpaqueID(request.MessageID) || request.MaxContentBytes <= 0 || request.MaxContentBytes > MaxEngineMessageBytes {
		return EngineMessage{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationGetMessage); ok {
		return EngineMessage{}, fakeEngineError(fault.Kind)
	}
	message, ok := f.messages[request.AccountID][request.MessageID]
	if !ok {
		return EngineMessage{}, fakeEngineError(EngineErrorNotFound)
	}
	result := cloneMessage(message)
	if override, ok := f.overrides[fakeAccountCursorKey{accountID: request.AccountID, cursor: request.MessageID}]; ok {
		result.Projection = cloneProjection(override)
	}
	if int64(messageVariableBytes(result)) > request.MaxContentBytes {
		return EngineMessage{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return result, nil
}

func (f *FakeMailEngine) GetDraft(_ context.Context, request EngineDraftReadRequest) (EngineMessage, error) {
	if !validAccountID(request.AccountID) || !validSafeText(request.DraftKey, MaxEngineOpaqueIDBytes, true) || !validOpaqueID(request.MessageID) || !validOpaqueID(request.ExpectedVersion) || request.MaxContentBytes <= 0 || request.MaxContentBytes > MaxEngineMessageBytes {
		return EngineMessage{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationGetDraft); ok {
		return EngineMessage{}, fakeEngineError(fault.Kind)
	}
	draft, ok := f.drafts[request.AccountID][request.DraftKey]
	if !ok {
		return EngineMessage{}, fakeEngineError(EngineErrorNotFound)
	}
	message := draft.message
	if message.Projection.AccountID != request.AccountID || !message.Projection.Flags.Draft || message.Projection.Origin != EngineProjectionOriginConfirmedOwnerDraft {
		return EngineMessage{}, fakeEngineError(EngineErrorUnavailable)
	}
	if message.Projection.MessageID != request.MessageID || message.Projection.Version != request.ExpectedVersion {
		return EngineMessage{}, fakeEngineError(EngineErrorConflict)
	}
	if err := validateEngineMessage(message); err != nil {
		return EngineMessage{}, fakeEngineError(EngineErrorUnavailable)
	}
	if int64(messageVariableBytes(message)) > request.MaxContentBytes {
		return EngineMessage{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return cloneMessage(message), nil
}

func (f *FakeMailEngine) GetBlob(_ context.Context, request EngineBlobReadRequest) (EngineBlob, error) {
	if !validAccountID(request.AccountID) || !validOpaqueID(request.BlobID) || request.MaxBytes <= 0 || request.MaxBytes > MaxEngineAttachmentBytes {
		return EngineBlob{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationGetBlob); ok {
		return EngineBlob{}, fakeEngineError(fault.Kind)
	}
	blob, ok := f.blobs[request.AccountID][request.BlobID]
	if !ok {
		return EngineBlob{}, fakeEngineError(EngineErrorNotFound)
	}
	if blob.SizeBytes != int64(len(blob.Bytes)) || blob.SizeBytes > request.MaxBytes {
		return EngineBlob{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return cloneBlob(blob), nil
}

func (f *FakeMailEngine) ReadCalendarPart(_ context.Context, request EngineCalendarPartReadRequest) (EngineCalendarPartReadResult, error) {
	if !validFakeCalendarPartReadRequest(request) {
		return EngineCalendarPartReadResult{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationReadCalendarPart); ok {
		return EngineCalendarPartReadResult{}, fakeEngineError(fault.Kind)
	}
	part, ok := f.calendarParts[fakeCalendarPartKey{
		accountID:   request.AccountID,
		messageID:   request.MessageID,
		locatorHash: request.LocatorHash,
	}]
	if !ok {
		return EngineCalendarPartReadResult{}, fakeEngineError(EngineErrorNotFound)
	}
	if part.version != request.Version || part.method != request.Method {
		return EngineCalendarPartReadResult{}, fakeEngineError(EngineErrorConflict)
	}
	if len(part.itip) > request.MaxBytes {
		return EngineCalendarPartReadResult{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return EngineCalendarPartReadResult{
		LocatorHash: request.LocatorHash,
		ITIP:        append([]byte(nil), part.itip...),
	}, nil
}

func (f *FakeMailEngine) Changes(_ context.Context, request EngineChangeRequest) (EngineChangePage, error) {
	if !validAccountID(request.AccountID) {
		return EngineChangePage{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationChanges); ok {
		return EngineChangePage{}, fakeEngineError(fault.Kind)
	}
	if err := validatePageRequest(request.Page); err != nil {
		return EngineChangePage{}, err
	}
	if !validCursor(request.FromCursor, false) {
		return EngineChangePage{}, fakeEngineError(EngineErrorInvalidCursor)
	}
	if script, ok := f.changeScript[fakeAccountCursorKey{accountID: request.AccountID, cursor: request.FromCursor}]; ok {
		if script.Fault != "" {
			return EngineChangePage{}, fakeEngineError(script.Fault)
		}
		page := cloneChangePage(script.Page)
		if err := validateChangePage(page, request); err != nil {
			return EngineChangePage{}, err
		}
		return page, nil
	}
	start, err := parseChangesCursor(request.FromCursor)
	if err != nil || start > len(f.changes[request.AccountID]) {
		return EngineChangePage{}, fakeEngineError(EngineErrorInvalidCursor)
	}
	end := start + request.Page.Limit
	if end > len(f.changes[request.AccountID]) {
		end = len(f.changes[request.AccountID])
	}
	created := make([]EngineMessageProjection, 0, end-start)
	destroyed := make([]string, 0, end-start)
	for _, change := range f.changes[request.AccountID][start:end] {
		if change.projection != nil {
			created = append(created, cloneProjection(*change.projection))
		} else {
			destroyed = append(destroyed, change.destroyed)
		}
	}
	next := ""
	if end < len(f.changes[request.AccountID]) {
		next = syncCursor(end)
	}
	page := EngineChangePage{AccountID: request.AccountID, FromCursor: request.FromCursor, NextCursor: next, Bounds: EnginePageBounds{Requested: request.Page.Limit, Maximum: MaxEnginePageItems, Returned: end - start}, CreatedOrUpdated: created, DestroyedIDs: destroyed}
	if err := validateChangePage(page, request); err != nil {
		return EngineChangePage{}, err
	}
	return page, nil
}

func (f *FakeMailEngine) EnumerateProjection(_ context.Context, request EngineEnumerationRequest) (EngineEnumerationPage, error) {
	if !validAccountID(request.AccountID) || !validCursor(request.BaseCursor, true) || !validCursor(request.Checkpoint, true) {
		return EngineEnumerationPage{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	if (request.BaseCursor == "") != (request.Checkpoint == "") {
		return EngineEnumerationPage{}, fakeEngineError(EngineErrorRestartRequired)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationEnumerateProjection); ok {
		return EngineEnumerationPage{}, fakeEngineError(fault.Kind)
	}
	if err := validatePageRequest(request.Page); err != nil {
		return EngineEnumerationPage{}, err
	}
	if script, ok := f.enumScript[fakeEnumerationKey{accountID: request.AccountID, baseCursor: request.BaseCursor, checkpoint: request.Checkpoint}]; ok {
		if script.Fault != "" {
			return EngineEnumerationPage{}, fakeEngineError(script.Fault)
		}
		page := cloneEnumerationPage(script.Page)
		if err := validateEnumerationPage(page, request); err != nil {
			return EngineEnumerationPage{}, err
		}
		return page, nil
	}
	if request.BaseCursor == "" && request.Checkpoint == "" {
		base := syncCursor(len(f.changes[request.AccountID]))
		f.snapshots[fakeAccountCursorKey{accountID: request.AccountID, cursor: base}] = fakeEnumerationSnapshot{accountID: request.AccountID, messages: f.sortedProjectionsLocked(request.AccountID, "")}
		return f.enumerationPageLocked(request, base, 0)
	}
	checkpoint, ok := f.checkpoints[fakeEnumerationKey{accountID: request.AccountID, baseCursor: request.BaseCursor, checkpoint: request.Checkpoint}]
	if !ok || checkpoint.accountID != request.AccountID || checkpoint.base != request.BaseCursor {
		return EngineEnumerationPage{}, fakeEngineError(EngineErrorRestartRequired)
	}
	return f.enumerationPageLocked(request, checkpoint.base, checkpoint.offset)
}

func (f *FakeMailEngine) SaveDraft(ctx context.Context, draft EngineDraft) (EngineMessage, error) {
	capability, sourceMode, capabilityErr := mailboxDraftAttachmentCapabilityFromContext(ctx, draft)
	if capabilityErr != nil {
		return EngineMessage{}, capabilityErr
	}
	if sourceMode {
		base := draft
		base.Message.Attachments = nil
		if err := validateDraftRequest(base); err != nil {
			return EngineMessage{}, err
		}
		plan, err := mailboxDraftMIMEPlanForDraft(ctx, draft, capability)
		if err != nil {
			return EngineMessage{}, err
		}
		materialized, err := mailboxDraftMIMEMaterializeAttachments(ctx, plan)
		if err != nil {
			return EngineMessage{}, err
		}
		draft = materialized
	}
	if err := validateDraftRequest(draft); err != nil {
		return EngineMessage{}, err
	}
	if draftMinimumBytes(draft) > draft.MaxContentBytes {
		return EngineMessage{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	f.mu.Lock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationSaveDraft)
	if faulted && !fault.AfterDispatch {
		f.mu.Unlock()
		return EngineMessage{}, fakeEngineError(fault.Kind)
	}
	existing, exists := f.drafts[draft.AccountID][draft.DraftKey]
	if draft.MessageID == "" {
		if exists {
			message := cloneMessage(existing.message)
			f.mu.Unlock()
			if faulted && fault.AfterDispatch {
				return EngineMessage{}, fakeEngineError(fault.Kind)
			}
			return message, nil
		}
	} else {
		if !exists || existing.message.Projection.MessageID != draft.MessageID {
			f.mu.Unlock()
			return EngineMessage{}, fakeEngineError(EngineErrorNotFound)
		}
		if existing.message.Projection.Version != draft.PreviousVersion {
			f.mu.Unlock()
			return EngineMessage{}, fakeEngineError(EngineErrorConflict)
		}
	}
	f.mu.Unlock()

	var candidateMessageID string
	var err error
	if draft.MessageID == "" {
		candidateMessageID, err = f.nextOpaqueID()
		if err != nil {
			return EngineMessage{}, err
		}
	}
	candidateVersion, err := f.nextOpaqueID()
	if err != nil {
		return EngineMessage{}, err
	}
	now := f.currentTime()
	messageID := draft.MessageID
	if messageID == "" {
		messageID = candidateMessageID
	}
	message := messageFromDraft(draft, messageID, candidateVersion, now)
	if int64(messageVariableBytes(message)) > draft.MaxContentBytes || validateEngineMessage(message) != nil {
		return EngineMessage{}, fakeEngineError(EngineErrorPayloadTooLarge)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	existing, exists = f.drafts[draft.AccountID][draft.DraftKey]
	if draft.MessageID == "" {
		if exists {
			if faulted && fault.AfterDispatch {
				return EngineMessage{}, fakeEngineError(fault.Kind)
			}
			return cloneMessage(existing.message), nil
		}
		if err := f.reserveOpaqueIDsLocked(candidateMessageID, candidateVersion); err != nil {
			return EngineMessage{}, err
		}
	} else {
		if !exists || existing.message.Projection.MessageID != draft.MessageID {
			return EngineMessage{}, fakeEngineError(EngineErrorNotFound)
		}
		if existing.message.Projection.Version != draft.PreviousVersion {
			return EngineMessage{}, fakeEngineError(EngineErrorConflict)
		}
		if err := f.reserveOpaqueIDsLocked(candidateVersion); err != nil {
			return EngineMessage{}, err
		}
	}
	if f.drafts[draft.AccountID] == nil {
		f.drafts[draft.AccountID] = make(map[string]fakeDraft)
	}
	f.retainDraftBlobsLocked(draft)
	f.drafts[draft.AccountID][draft.DraftKey] = fakeDraft{message: message, identity: fakeDraftIdentity{messageID: draft.MessageID, previousVersion: draft.PreviousVersion, mutationID: draft.MutationID}}
	f.putMessageLocked(message, true)
	if faulted && fault.AfterDispatch {
		return EngineMessage{}, fakeEngineError(fault.Kind)
	}
	return cloneMessage(message), nil
}

func (f *FakeMailEngine) ReconcileDraft(_ context.Context, request EngineDraftReconciliation) (EngineDraftResult, error) {
	if err := validateDraftReconciliation(request); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationReconcileDraft); ok {
		if fault.Kind == EngineErrorIndeterminate {
			return EngineDraftIndeterminate{Code: EngineIndeterminateTimeout}, nil
		}
		return nil, fakeEngineError(fault.Kind)
	}
	draft, ok := f.drafts[request.AccountID][request.DraftKey]
	if !ok {
		if request.MessageID == "" {
			return EngineDraftAbsent{}, nil
		}
		return EngineDraftIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	if draft.identity.mutationID == request.MutationID &&
		(request.MessageID == "" ||
			draft.message.Projection.MessageID != request.MessageID ||
			draft.message.Projection.Version != request.PreviousVersion) {
		if int64(messageVariableBytes(draft.message)) > request.MaxContentBytes {
			return nil, fakeEngineError(EngineErrorPayloadTooLarge)
		}
		return EngineDraftApplied{Message: cloneMessage(draft.message)}, nil
	}
	if request.MessageID == "" {
		return EngineDraftIndeterminate{Code: EngineIndeterminateTimeout}, nil
	}
	message, exists := f.messages[request.AccountID][request.MessageID]
	if !exists || !fakeExactRetainedDraft(request, draft, message) {
		return EngineDraftIndeterminate{Code: EngineIndeterminateTimeout}, nil
	}
	if int64(messageVariableBytes(message)) > request.MaxContentBytes {
		return nil, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return EngineDraftRetained{Message: cloneMessage(message)}, nil
}

func (f *FakeMailEngine) DeleteDraft(_ context.Context, request EngineDraftDeletion) error {
	if err := validateDraftDeletion(request); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationDeleteDraft)
	if faulted && !fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	draft, exists := f.drafts[request.AccountID][request.DraftKey]
	if !exists {
		return fakeEngineError(EngineErrorNotFound)
	}
	if draft.message.Projection.MessageID != request.MessageID || draft.message.Projection.Version != request.ExpectedVersion {
		return fakeEngineError(EngineErrorConflict)
	}
	message, exists := f.messages[request.AccountID][request.MessageID]
	if !exists || !fakeExactDraftDeletion(request, draft, message) {
		return fakeEngineError(EngineErrorIndeterminate)
	}
	delete(f.drafts[request.AccountID], request.DraftKey)
	delete(f.messages[request.AccountID], request.MessageID)
	f.changes[request.AccountID] = append(f.changes[request.AccountID], fakeChange{destroyed: request.MessageID})
	if faulted && fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	return nil
}

func (f *FakeMailEngine) ReconcileDraftDeletion(_ context.Context, request EngineDraftDeletion) (EngineDraftDeletionResult, error) {
	if err := validateDraftDeletion(request); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationReconcileDraftDeletion); ok {
		switch fault.Kind {
		case EngineErrorUnavailable, EngineErrorIndeterminate, EngineErrorRestartRequired:
			return EngineDraftDeletionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
		default:
			return nil, fakeEngineError(fault.Kind)
		}
	}
	draft, exists := f.drafts[request.AccountID][request.DraftKey]
	if !exists {
		return EngineDraftDeleted{}, nil
	}
	message, exists := f.messages[request.AccountID][request.MessageID]
	if !exists || !fakeExactDraftDeletion(request, draft, message) {
		return EngineDraftDeletionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	return EngineDraftRetainedUnchanged{}, nil
}

func (f *FakeMailEngine) Submit(_ context.Context, submission EngineSubmission) (EngineSubmissionResult, error) {
	if err := validateSubmission(submission); err != nil {
		return nil, err
	}
	f.mu.Lock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationSubmit)
	if faulted && !fault.AfterDispatch {
		f.mu.Unlock()
		return nil, fakeEngineError(fault.Kind)
	}
	if stored, ok := f.submissions[submission.AccountID][submission.SubmissionKey]; ok {
		if !fakeSubmissionBindingMatches(stored, submission) {
			f.mu.Unlock()
			return nil, fakeEngineError(EngineErrorIndeterminate)
		}
		f.mu.Unlock()
		if faulted && fault.AfterDispatch {
			return nil, fakeEngineError(fault.Kind)
		}
		return stored.accepted, nil
	}
	draft, ok := f.drafts[submission.AccountID][submission.DraftKey]
	if !ok || draft.message.Projection.MessageID != submission.MessageID {
		f.mu.Unlock()
		return nil, fakeEngineError(EngineErrorNotFound)
	}
	if draft.message.Projection.Version != submission.ExpectedVersion {
		f.mu.Unlock()
		return nil, fakeEngineError(EngineErrorConflict)
	}
	message := cloneMessage(draft.message)
	f.mu.Unlock()
	message.Projection.MessageID = ""
	if int64(messageVariableBytes(message)) > submission.MaxMessageBytes {
		return nil, fakeEngineError(EngineErrorPayloadTooLarge)
	}

	candidateSubmissionID, err := f.nextOpaqueID()
	if err != nil {
		return nil, err
	}
	candidateMessageID, err := f.nextOpaqueID()
	if err != nil {
		return nil, err
	}
	message.Projection.MessageID = candidateMessageID
	message.Projection.Flags.Draft = false
	message.Projection.Origin = EngineProjectionOriginConfirmedOwnerSubmission
	message.Projection.FolderIDs = []string{"sent"}
	if int64(messageVariableBytes(message)) > submission.MaxMessageBytes {
		return nil, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	if err := validateEngineMessage(message); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if stored, ok := f.submissions[submission.AccountID][submission.SubmissionKey]; ok {
		if !fakeSubmissionBindingMatches(stored, submission) {
			return nil, fakeEngineError(EngineErrorIndeterminate)
		}
		if faulted && fault.AfterDispatch {
			return nil, fakeEngineError(fault.Kind)
		}
		return stored.accepted, nil
	}
	draft, ok = f.drafts[submission.AccountID][submission.DraftKey]
	if !ok || draft.message.Projection.MessageID != submission.MessageID {
		return nil, fakeEngineError(EngineErrorNotFound)
	}
	if draft.message.Projection.Version != submission.ExpectedVersion {
		return nil, fakeEngineError(EngineErrorConflict)
	}
	if err := f.reserveOpaqueIDsLocked(candidateSubmissionID, candidateMessageID); err != nil {
		return nil, err
	}
	f.putMessageLocked(message, true)
	if f.submissions[submission.AccountID] == nil {
		f.submissions[submission.AccountID] = make(map[string]fakeSubmission)
	}
	accepted := EngineSubmissionAccepted{SubmissionID: candidateSubmissionID, MessageID: candidateMessageID}
	f.submissions[submission.AccountID][submission.SubmissionKey] = fakeSubmission{accepted: accepted, draftKey: submission.DraftKey, messageID: submission.MessageID, expectedVersion: submission.ExpectedVersion}
	if faulted && fault.AfterDispatch {
		return nil, fakeEngineError(fault.Kind)
	}
	return accepted, nil
}

func (f *FakeMailEngine) ReconcileSubmission(ctx context.Context, submission EngineSubmission) (EngineSubmissionResult, error) {
	if ctx == nil {
		return nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	if err := validateSubmission(submission); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationReconcileSubmission); ok {
		if fault.Kind == EngineErrorIndeterminate {
			return EngineSubmissionIndeterminate{Code: EngineIndeterminateTimeout}, nil
		}
		return nil, fakeEngineError(fault.Kind)
	}
	stored, ok := f.submissions[submission.AccountID][submission.SubmissionKey]
	if !ok {
		return EngineSubmissionAbsent{}, nil
	}
	if !fakeSubmissionBindingMatches(stored, submission) {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	accepted := stored.accepted
	message, exists := f.messages[submission.AccountID][accepted.MessageID]
	if !exists || !validOpaqueID(accepted.SubmissionID) || !validOpaqueID(accepted.MessageID) ||
		validateEngineMessage(message) != nil || message.Projection.AccountID != submission.AccountID ||
		message.Projection.MessageID != accepted.MessageID || message.Projection.Flags.Draft ||
		message.Projection.Origin != EngineProjectionOriginConfirmedOwnerSubmission {
		return EngineSubmissionIndeterminate{Code: EngineIndeterminateUnavailable}, nil
	}
	if int64(messageVariableBytes(message)) > submission.MaxMessageBytes {
		return nil, fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return accepted, nil
}

func fakeSubmissionBindingMatches(stored fakeSubmission, submission EngineSubmission) bool {
	return stored.draftKey == submission.DraftKey && stored.messageID == submission.MessageID && stored.expectedVersion == submission.ExpectedVersion
}

func (f *FakeMailEngine) MoveMessages(_ context.Context, accountID string, messageIDs []string, folderID string) error {
	if !validAccountID(accountID) || !validOpaqueID(folderID) || !validMessageIDs(messageIDs) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationMoveMessages)
	if faulted && !fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	if _, ok := f.folders[accountID][folderID]; !ok {
		return fakeEngineError(EngineErrorNotFound)
	}
	for _, messageID := range messageIDs {
		if _, ok := f.messages[accountID][messageID]; !ok {
			return fakeEngineError(EngineErrorNotFound)
		}
	}
	for _, messageID := range messageIDs {
		message := f.messages[accountID][messageID]
		message.Projection.FolderIDs = []string{folderID}
		f.putMessageLocked(message, true)
	}
	if faulted && fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	return nil
}

func (f *FakeMailEngine) SetFlags(_ context.Context, accountID string, messageIDs []string, patch FlagPatch) error {
	if !validAccountID(accountID) || !validMessageIDs(messageIDs) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationSetFlags)
	if faulted && !fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	for _, messageID := range messageIDs {
		if _, ok := f.messages[accountID][messageID]; !ok {
			return fakeEngineError(EngineErrorNotFound)
		}
	}
	for _, messageID := range messageIDs {
		message := f.messages[accountID][messageID]
		if patch.Seen.Present {
			message.Projection.Flags.Seen = patch.Seen.Value
		}
		if patch.Answered.Present {
			message.Projection.Flags.Answered = patch.Answered.Value
		}
		if patch.Flagged.Present {
			message.Projection.Flags.Flagged = patch.Flagged.Value
		}
		if patch.Junk.Present {
			message.Projection.Flags.Junk = patch.Junk.Value
		}
		f.putMessageLocked(message, true)
	}
	if faulted && fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	return nil
}

func (f *FakeMailEngine) SetKeywords(_ context.Context, accountID string, messageIDs []string, patch KeywordPatch) error {
	if !validAccountID(accountID) || !validMessageIDs(messageIDs) || len(patch.Add) > MaxEngineKeywordItems || len(patch.Remove) > MaxEngineKeywordItems || !validKeywords(patch.Add) || !validKeywords(patch.Remove) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationSetKeywords)
	if faulted && !fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	for _, messageID := range messageIDs {
		if _, ok := f.messages[accountID][messageID]; !ok {
			return fakeEngineError(EngineErrorNotFound)
		}
	}
	for _, messageID := range messageIDs {
		message := f.messages[accountID][messageID]
		keywords := make(map[EngineKeywordMarker]bool, len(message.Projection.ReservedKeywords)+len(patch.Add))
		for _, keyword := range message.Projection.ReservedKeywords {
			keywords[keyword] = true
		}
		for _, keyword := range patch.Add {
			keywords[keyword] = true
		}
		for _, keyword := range patch.Remove {
			delete(keywords, keyword)
		}
		if len(keywords) > MaxEngineKeywordItems {
			return fakeEngineError(EngineErrorRejectedPolicy)
		}
		message.Projection.ReservedKeywords = make([]EngineKeywordMarker, 0, len(keywords))
		for keyword := range keywords {
			message.Projection.ReservedKeywords = append(message.Projection.ReservedKeywords, keyword)
		}
		sort.Slice(message.Projection.ReservedKeywords, func(left, right int) bool { return message.Projection.ReservedKeywords[left] < message.Projection.ReservedKeywords[right] })
		f.putMessageLocked(message, true)
	}
	if faulted && fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	return nil
}

func (f *FakeMailEngine) DeleteMessages(_ context.Context, accountID string, messageIDs []string, permanent bool) error {
	if !validAccountID(accountID) || !validMessageIDs(messageIDs) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationDeleteMessages)
	if faulted && !fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	for _, messageID := range messageIDs {
		if _, ok := f.messages[accountID][messageID]; !ok {
			return fakeEngineError(EngineErrorNotFound)
		}
	}
	for _, messageID := range messageIDs {
		if permanent {
			delete(f.messages[accountID], messageID)
			f.changes[accountID] = append(f.changes[accountID], fakeChange{destroyed: messageID})
			continue
		}
		message := f.messages[accountID][messageID]
		message.Projection.FolderIDs = []string{"trash"}
		f.putMessageLocked(message, true)
	}
	if faulted && fault.AfterDispatch {
		return fakeEngineError(fault.Kind)
	}
	return nil
}

func (f *FakeMailEngine) GetSieve(_ context.Context, accountID, name string) (EngineSieveState, error) {
	if !validAccountID(accountID) || !validSafeText(name, MaxEngineMIMEFieldBytes, true) {
		return EngineSieveState{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationGetSieve); ok {
		return EngineSieveState{}, fakeEngineError(fault.Kind)
	}
	state, ok := f.sieves[accountID]
	if !ok || state.Name != name {
		return EngineSieveState{Name: name, State: EngineSieveAbsent}, nil
	}
	return state, nil
}

func (f *FakeMailEngine) UpsertSieve(_ context.Context, accountID string, script SieveScript) (EngineSieveState, error) {
	if !validAccountID(accountID) || !validSafeText(script.Name, MaxEngineMIMEFieldBytes, true) || !validSafeText(script.Hash, MaxEngineOpaqueIDBytes, true) {
		return EngineSieveState{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationUpsertSieve)
	if faulted && !fault.AfterDispatch {
		return EngineSieveState{}, fakeEngineError(fault.Kind)
	}
	state := EngineSieveState{Name: script.Name, Active: true, Generation: script.Generation, Hash: script.Hash, State: EngineSieveApplied}
	f.sieves[accountID] = state
	if faulted && fault.AfterDispatch {
		return EngineSieveState{}, fakeEngineError(fault.Kind)
	}
	return state, nil
}

func (f *FakeMailEngine) ClassifyMessages(_ context.Context, training EngineSpamTraining) (EngineSpamTrainingResult, error) {
	if !validAccountID(training.AccountID) || !validSafeText(training.ReceiptID, MaxEngineOpaqueIDBytes, true) || !validSafeText(training.CommandID, MaxEngineOpaqueIDBytes, true) || len(training.Samples) > MaxEngineSpamTrainingItems || !validSpamSamples(training.Samples) {
		return EngineSpamTrainingResult{}, fakeEngineError(EngineErrorRejectedPolicy)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	fault, faulted := f.takeFaultLocked(FakeMailOperationClassifyMessages)
	if faulted && !fault.AfterDispatch {
		return EngineSpamTrainingResult{}, fakeEngineError(fault.Kind)
	}
	key := fakeAccountCursorKey{accountID: training.AccountID, cursor: training.ReceiptID}
	if result, ok := f.spamResults[key]; ok {
		if faulted && fault.AfterDispatch {
			return EngineSpamTrainingResult{}, fakeEngineError(fault.Kind)
		}
		return cloneSpamResult(result), nil
	}
	result := EngineSpamTrainingResult{Outcomes: make([]EngineSpamTrainingOutcome, 0, len(training.Samples))}
	for _, sample := range training.Samples {
		outcome := EngineSpamTrainingOutcome{MessageID: sample.MessageID, Status: EngineSpamTrainingAccepted}
		if _, ok := f.messages[training.AccountID][sample.MessageID]; !ok {
			outcome.Status = EngineSpamTrainingRejected
		}
		result.Outcomes = append(result.Outcomes, outcome)
	}
	f.spamResults[key] = cloneSpamResult(result)
	if faulted && fault.AfterDispatch {
		return EngineSpamTrainingResult{}, fakeEngineError(fault.Kind)
	}
	return cloneSpamResult(result), nil
}

func (f *FakeMailEngine) Health(_ context.Context) (EngineHealth, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if fault, ok := f.takeFaultLocked(FakeMailOperationHealth); ok {
		return EngineHealth{}, fakeEngineError(fault.Kind)
	}
	return EngineHealth{Available: true, State: EngineHealthReady}, nil
}

func (f *FakeMailEngine) putMessageLocked(message EngineMessage, recordChange bool) {
	message = cloneMessage(message)
	if f.messages[message.Projection.AccountID] == nil {
		f.messages[message.Projection.AccountID] = make(map[string]EngineMessage)
	}
	f.messages[message.Projection.AccountID][message.Projection.MessageID] = message
	if recordChange {
		projection := cloneProjection(message.Projection)
		f.changes[message.Projection.AccountID] = append(f.changes[message.Projection.AccountID], fakeChange{projection: &projection})
	}
}

func (f *FakeMailEngine) sortedProjectionsLocked(accountID, folderID string) []EngineMessageProjection {
	result := make([]EngineMessageProjection, 0, len(f.messages[accountID]))
	for _, message := range f.messages[accountID] {
		projection := cloneProjection(message.Projection)
		if folderID != "" && !projectionHasFolder(projection, folderID) {
			continue
		}
		result = append(result, projection)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].MessageID < result[right].MessageID })
	return result
}

func (f *FakeMailEngine) enumerationPageLocked(request EngineEnumerationRequest, base string, start int) (EngineEnumerationPage, error) {
	snapshot, ok := f.snapshots[fakeAccountCursorKey{accountID: request.AccountID, cursor: base}]
	if !ok || snapshot.accountID != request.AccountID {
		return EngineEnumerationPage{}, fakeEngineError(EngineErrorRestartRequired)
	}
	end := start + request.Page.Limit
	if end > len(snapshot.messages) {
		end = len(snapshot.messages)
	}
	next := ""
	complete := end == len(snapshot.messages)
	if !complete {
		f.serial++
		next = "checkpoint:" + strconv.Itoa(start) + ":" + strconv.Itoa(end) + ":" + strconv.Itoa(f.serial)
		f.checkpoints[fakeEnumerationKey{accountID: request.AccountID, baseCursor: base, checkpoint: next}] = fakeCheckpoint{accountID: request.AccountID, base: base, offset: end}
	}
	page := EngineEnumerationPage{AccountID: request.AccountID, BaseCursor: base, Checkpoint: request.Checkpoint, NextCheckpoint: next, Complete: complete, Bounds: EnginePageBounds{Requested: request.Page.Limit, Maximum: MaxEnginePageItems, Returned: end - start}, Messages: cloneProjections(snapshot.messages[start:end])}
	if err := validateEnumerationPage(page, request); err != nil {
		return EngineEnumerationPage{}, err
	}
	return page, nil
}

func messageFromDraft(draft EngineDraft, messageID, version string, now time.Time) EngineMessage {
	attachments := make([]EngineProjectionAttachment, 0, len(draft.Message.Attachments))
	for _, attachment := range draft.Message.Attachments {
		blobID := attachment.AttachmentKey
		attachments = append(attachments, EngineProjectionAttachment{BlobID: blobID, AttachmentKey: attachment.AttachmentKey, ContentID: attachment.ContentID, Filename: attachment.Filename, MediaType: attachment.MediaType, SizeBytes: int64(len(attachment.Bytes)), Disposition: attachment.Disposition, SHA256: append([]byte(nil), attachment.SHA256[:]...)})
	}
	now = now.UTC()
	return EngineMessage{Projection: EngineMessageProjection{AccountID: draft.AccountID, MessageID: messageID, Version: version, From: draft.Message.From, To: cloneAddresses(draft.Message.To), CC: cloneAddresses(draft.Message.CC), BCC: cloneAddresses(draft.Message.BCC), ReplyTo: cloneAddresses(draft.Message.ReplyTo), Subject: draft.Message.Subject, InReplyTo: draft.Message.InReplyTo, References: cloneStrings(draft.Message.References), ReceivedAt: &now, AuthenticationClass: MailAuthenticationClassUntrusted, SizeBytes: int64(len(draft.Message.PlainBody) + len(draft.Message.RichBody)), Priority: draft.Message.Priority, FolderIDs: []string{"drafts"}, Flags: EngineMessageFlags{Draft: true}, Attachments: attachments, Origin: EngineProjectionOriginConfirmedOwnerDraft}, Body: EngineMessageBody{Format: draft.Message.Format, PlainBody: draft.Message.PlainBody, RichBody: draft.Message.RichBody}}
}

func (f *FakeMailEngine) retainDraftBlobsLocked(draft EngineDraft) {
	if len(draft.Message.Attachments) == 0 {
		return
	}
	if f.blobs[draft.AccountID] == nil {
		f.blobs[draft.AccountID] = make(map[string]EngineBlob)
	}
	for _, attachment := range draft.Message.Attachments {
		f.blobs[draft.AccountID][attachment.AttachmentKey] = EngineBlob{AccountID: draft.AccountID, BlobID: attachment.AttachmentKey, MediaType: attachment.MediaType, SizeBytes: int64(len(attachment.Bytes)), Bytes: append([]byte(nil), attachment.Bytes...)}
	}
}

func (f *FakeMailEngine) nextOpaqueID() (string, error) {
	f.sourceMu.Lock()
	id := f.nextID()
	f.sourceMu.Unlock()
	if !validOpaqueID(id) {
		return "", fakeEngineError(EngineErrorRejectedPolicy)
	}
	return id, nil
}

func (f *FakeMailEngine) currentTime() time.Time {
	f.sourceMu.Lock()
	now := f.now()
	f.sourceMu.Unlock()
	return now.UTC()
}

func domainStateForSpec(id string, spec DomainSpec) EngineDomainState {
	state := EngineDomainState{DomainID: id, DNSRecords: []EngineDNSRecord{}, Verification: EngineVerificationUnknown, DKIM: EngineVerificationUnknown}
	if !spec.Present {
		state.State = EngineDomainAbsent
	} else if spec.Enabled {
		state.State = EngineDomainVerified
		state.Applied = true
		state.Verification = EngineVerificationVerified
		state.DKIM = EngineVerificationVerified
	} else {
		state.State = EngineDomainDNSRequired
		state.Verification = EngineVerificationPending
		state.DKIM = EngineVerificationPending
	}
	return state
}

func accountIdentityIDs(accountID string, aliasCount int) (string, []string, error) {
	primaryID := accountID + "-primary"
	if !validOpaqueID(primaryID) {
		return "", nil, fakeEngineError(EngineErrorRejectedPolicy)
	}
	aliases := make([]string, aliasCount)
	for index := range aliases {
		aliases[index] = accountID + "-alias-" + strconv.Itoa(index+1)
		if !validOpaqueID(aliases[index]) {
			return "", nil, fakeEngineError(EngineErrorRejectedPolicy)
		}
	}
	return primaryID, aliases, nil
}

func draftMinimumBytes(draft EngineDraft) int64 {
	messageID := draft.MessageID
	minimum := messageFromDraft(draft, messageID, "", time.Time{})
	return int64(messageVariableBytes(minimum))
}

func (f *FakeMailEngine) reserveOpaqueIDsLocked(ids ...string) error {
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			return fakeEngineError(EngineErrorConflict)
		}
		if _, duplicate := f.issuedIDs[id]; duplicate {
			return fakeEngineError(EngineErrorConflict)
		}
		seen[id] = struct{}{}
	}
	for id := range seen {
		f.issuedIDs[id] = struct{}{}
	}
	return nil
}

func (f *FakeMailEngine) takeFaultLocked(operation FakeMailOperation) (FakeMailFault, bool) {
	if len(f.faults) == 0 || f.faults[0].Operation != operation {
		return FakeMailFault{}, false
	}
	fault := f.faults[0]
	f.faults = f.faults[1:]
	return fault, true
}

func fakeEngineError(kind EngineErrorKind) error {
	if kind == "" {
		kind = EngineErrorIndeterminate
	}
	return &EngineError{Kind: kind}
}


func validatePageRequest(page EnginePageRequest) error {
	if page.Limit < 1 || page.Limit > MaxEnginePageItems {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	return nil
}

func validateDraftRequest(draft EngineDraft) error {
	if !validAccountID(draft.AccountID) || !validSafeText(draft.DraftKey, MaxEngineOpaqueIDBytes, true) || !validSafeText(draft.MutationID, MaxEngineOpaqueIDBytes, true) || draft.MaxContentBytes <= 0 || draft.MaxContentBytes > MaxEngineMessageBytes {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	if (draft.MessageID == "") != (draft.PreviousVersion == "") || (draft.MessageID != "" && (!validOpaqueID(draft.MessageID) || !validOpaqueID(draft.PreviousVersion))) || validateMessageContent(draft.Message) != nil {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	return nil
}

func validateDraftReconciliation(request EngineDraftReconciliation) error {
	pairPresent := request.MessageID != ""
	if !validAccountID(request.AccountID) || !validSafeText(request.DraftKey, MaxEngineOpaqueIDBytes, true) || !validSafeText(request.MutationID, MaxEngineOpaqueIDBytes, true) || request.MaxContentBytes <= 0 || request.MaxContentBytes > MaxEngineMessageBytes || pairPresent != (request.PreviousVersion != "") || (pairPresent && (!validOpaqueID(request.MessageID) || !validOpaqueID(request.PreviousVersion))) || (!pairPresent && request.ConfirmedMutationID != "") || (pairPresent && request.ConfirmedMutationID != "" && (!validSafeText(request.ConfirmedMutationID, MaxEngineOpaqueIDBytes, true) || request.ConfirmedMutationID == request.MutationID)) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	return nil
}

func validateDraftDeletion(request EngineDraftDeletion) error {
	if !validAccountID(request.AccountID) || !validSafeText(request.DraftKey, MaxEngineOpaqueIDBytes, true) || !validOpaqueID(request.MessageID) || !validOpaqueID(request.ExpectedVersion) || !validSafeText(request.MutationID, MaxEngineOpaqueIDBytes, true) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	return nil
}

func fakeExactDraftDeletion(request EngineDraftDeletion, draft fakeDraft, message EngineMessage) bool {
	draftProjection := draft.message.Projection
	projection := message.Projection
	return validateEngineMessage(draft.message) == nil &&
		validateEngineMessage(message) == nil &&
		reflect.DeepEqual(draftProjection, projection) &&
		draftProjection.AccountID == request.AccountID &&
		draftProjection.MessageID == request.MessageID &&
		draftProjection.Version == request.ExpectedVersion &&
		draftProjection.Origin == EngineProjectionOriginConfirmedOwnerDraft &&
		draftProjection.Flags.Draft &&
		reflect.DeepEqual(draftProjection.FolderIDs, []string{"drafts"})
}

func fakeExactRetainedDraft(request EngineDraftReconciliation, draft fakeDraft, message EngineMessage) bool {
	draftProjection := draft.message.Projection
	return request.MessageID != "" &&
		request.PreviousVersion != "" &&
		request.ConfirmedMutationID != "" &&
		request.ConfirmedMutationID != request.MutationID &&
		draft.identity.mutationID == request.ConfirmedMutationID &&
		validateEngineMessage(draft.message) == nil &&
		validateEngineMessage(message) == nil &&
		reflect.DeepEqual(cloneMessage(draft.message), cloneMessage(message)) &&
		draftProjection.AccountID == request.AccountID &&
		draftProjection.MessageID == request.MessageID &&
		draftProjection.Version == request.PreviousVersion &&
		draftProjection.Origin == EngineProjectionOriginConfirmedOwnerDraft &&
		draftProjection.Flags.Draft &&
		reflect.DeepEqual(draftProjection.FolderIDs, []string{"drafts"})
}

func validateSubmission(submission EngineSubmission) error {
	if !validAccountID(submission.AccountID) || !validSafeText(submission.DraftKey, MaxEngineOpaqueIDBytes, true) || !validOpaqueID(submission.MessageID) || !validOpaqueID(submission.ExpectedVersion) || !validSafeText(submission.SubmissionKey, MaxEngineOpaqueIDBytes, true) || submission.MaxMessageBytes <= 0 || submission.MaxMessageBytes > MaxEngineMessageBytes {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	return nil
}

func validateMessageContent(content EngineMessageContent) error {
	if !validAddress(content.From, true) || !validAddresses(content.To, MaxEngineDraftRecipientItems) || !validAddresses(content.CC, MaxEngineDraftRecipientItems) || !validAddresses(content.BCC, MaxEngineDraftRecipientItems) || !validAddresses(content.ReplyTo, MaxEngineDraftRecipientItems) || !uniqueRecipients(content.To, content.CC, content.BCC) || len(content.To)+len(content.CC)+len(content.BCC) > MaxEngineDraftTotalRecipientItems || !validSafeText(content.Subject, MaxEngineHeaderValueBytes, false) || !validSafeText(content.InReplyTo, MaxEngineHeaderValueBytes, false) || !validStrings(content.References, MaxEngineReferenceItems, MaxEngineHeaderValueBytes) || !validDraftAttachments(content.Attachments) || (content.Priority != EnginePriorityLow && content.Priority != EnginePriorityNormal && content.Priority != EnginePriorityHigh) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	if content.Format == EngineBodyPlain {
		if content.RichBody != "" {
			return fakeEngineError(EngineErrorRejectedPolicy)
		}
	} else if content.Format != EngineBodyRich {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	if !validSafeText(content.PlainBody, MaxEngineBodyPartBytes, false) || !validSafeText(content.RichBody, MaxEngineBodyPartBytes, false) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	return nil
}

func validateEngineMessage(message EngineMessage) error {
	if err := validateProjection(message.Projection); err != nil {
		return err
	}
	if (message.Body.Format == EngineBodyPlain && message.Body.RichBody != "") || (message.Body.Format != EngineBodyPlain && message.Body.Format != EngineBodyRich) || !validSafeText(message.Body.PlainBody, MaxEngineBodyPartBytes, false) || !validSafeText(message.Body.RichBody, MaxEngineBodyPartBytes, false) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	return nil
}

func validFakeCalendarPart(part FakeCalendarPart) bool {
	return validAccountID(part.AccountID) && validOpaqueID(part.MessageID) && validOpaqueID(part.Version) && !calendarPartLocatorHashZero(part.LocatorHash) && validCalendarPartMethod(part.Method) && len(part.ITIP) > 0 && len(part.ITIP) <= MaxCalendarPartBytes
}

func validFakeCalendarPartReadRequest(request EngineCalendarPartReadRequest) bool {
	return validAccountID(request.AccountID) && validOpaqueID(request.MessageID) && validOpaqueID(request.Version) && !calendarPartLocatorHashZero(request.LocatorHash) && validCalendarPartMethod(request.Method) && request.MaxBytes >= 1 && request.MaxBytes <= MaxCalendarPartBytes
}

func validateProjection(projection EngineMessageProjection) error {
	if !validAccountID(projection.AccountID) || !validOpaqueID(projection.MessageID) || !validOpaqueID(projection.Version) || !validSafeText(projection.InternetMessageID, MaxEngineHeaderValueBytes, false) || !validSafeText(projection.InReplyTo, MaxEngineHeaderValueBytes, false) || !validStrings(projection.References, MaxEngineReferenceItems, MaxEngineHeaderValueBytes) || !validAddress(projection.From, true) || !validAddresses(projection.To, MaxEngineAddressItems) || !validAddresses(projection.CC, MaxEngineAddressItems) || !validAddresses(projection.BCC, MaxEngineAddressItems) || !validAddresses(projection.ReplyTo, MaxEngineAddressItems) || !uniqueRecipients(projection.To, projection.CC, projection.BCC) || !validSafeText(projection.Subject, MaxEngineHeaderValueBytes, false) || !validSafeText(projection.Snippet, MaxEngineSnippetBytes, false) || len(projection.FolderIDs) < 1 || len(projection.FolderIDs) > MaxEngineFolderItems || !validOpaqueIDs(projection.FolderIDs) || len(projection.ReservedKeywords) > MaxEngineKeywordItems || !validKeywords(projection.ReservedKeywords) || len(projection.Attachments) > MaxEngineAttachmentItems || !validProjectionAttachments(projection.Attachments) || projection.SizeBytes < 0 || projection.SizeBytes > MaxEngineMessageBytes || (projection.Priority != EnginePriorityLow && projection.Priority != EnginePriorityNormal && projection.Priority != EnginePriorityHigh) || (projection.Origin != EngineProjectionOriginInboundOrUnknown && projection.Origin != EngineProjectionOriginConfirmedOwnerDraft && projection.Origin != EngineProjectionOriginConfirmedOwnerSubmission) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	if !validMailCalendarMethod(projection.CalendarMethod) {
		return fakeEngineError(EngineErrorRejectedPolicy)
	}
	if projection.Origin == EngineProjectionOriginConfirmedOwnerDraft {
		attachmentKeys := make(map[string]struct{}, len(projection.Attachments))
		for _, attachment := range projection.Attachments {
			if !validOpaqueID(attachment.AttachmentKey) {
				return fakeEngineError(EngineErrorRejectedPolicy)
			}
			if _, duplicate := attachmentKeys[attachment.AttachmentKey]; duplicate {
				return fakeEngineError(EngineErrorRejectedPolicy)
			}
			attachmentKeys[attachment.AttachmentKey] = struct{}{}
		}
	} else {
		for _, attachment := range projection.Attachments {
			if attachment.AttachmentKey != "" {
				return fakeEngineError(EngineErrorRejectedPolicy)
			}
		}
	}
	if projectionVariableBytes(projection) > MaxEngineProjectionItemBytes {
		return fakeEngineError(EngineErrorPayloadTooLarge)
	}
	return nil
}

func validDraftAttachments(attachments []EngineDraftAttachment) bool {
	if len(attachments) > MaxEngineAttachmentItems {
		return false
	}
	seen := make(map[string]struct{}, len(attachments))
	for _, attachment := range attachments {
		if !validOpaqueID(attachment.AttachmentKey) || !validSafeText(attachment.ContentID, MaxEngineMIMEFieldBytes, false) || !validSafeText(attachment.Filename, MaxEngineMIMEFieldBytes, true) || !validSafeText(attachment.MediaType, MaxEngineMIMEFieldBytes, true) || len(attachment.Bytes) == 0 || len(attachment.Bytes) > MaxEngineAttachmentBytes || (attachment.Disposition != EngineAttachmentDispositionAttachment && attachment.Disposition != EngineAttachmentDispositionInline) || (attachment.Disposition == EngineAttachmentDispositionInline && attachment.ContentID == "") {
			return false
		}
		if _, duplicate := seen[attachment.AttachmentKey]; duplicate {
			return false
		}
		seen[attachment.AttachmentKey] = struct{}{}
	}
	return true
}

func validProjectionAttachments(attachments []EngineProjectionAttachment) bool {
	for _, attachment := range attachments {
		if !validOpaqueID(attachment.BlobID) || !validSafeText(attachment.ContentID, MaxEngineMIMEFieldBytes, false) || !validSafeText(attachment.Filename, MaxEngineMIMEFieldBytes, true) || !validSafeText(attachment.MediaType, MaxEngineMIMEFieldBytes, true) || attachment.SizeBytes < 0 || attachment.SizeBytes > MaxEngineAttachmentBytes || (len(attachment.SHA256) != 0 && len(attachment.SHA256) != 32) || (attachment.Disposition != EngineAttachmentDispositionAttachment && attachment.Disposition != EngineAttachmentDispositionInline) || (attachment.Disposition == EngineAttachmentDispositionInline && attachment.ContentID == "") {
			return false
		}
	}
	return true
}

func validAddress(address EngineAddress, required bool) bool {
	if address.Email == "" {
		return !required && validSafeText(address.Name, MaxEngineAddressNameBytes, false)
	}
	canonical, err := NormalizeAddress(address)
	return err == nil && canonical.Email == address.Email && canonical.Name == address.Name
}

func validAddresses(addresses []EngineAddress, maximum int) bool {
	if len(addresses) > maximum {
		return false
	}
	seen := make(map[string]struct{}, len(addresses))
	for _, address := range addresses {
		if !validAddress(address, true) {
			return false
		}
		if _, duplicate := seen[address.Email]; duplicate {
			return false
		}
		seen[address.Email] = struct{}{}
	}
	return true
}

func uniqueRecipients(collections ...[]EngineAddress) bool {
	seen := make(map[string]struct{})
	for _, collection := range collections {
		for _, address := range collection {
			if _, duplicate := seen[address.Email]; duplicate {
				return false
			}
			seen[address.Email] = struct{}{}
		}
	}
	return true
}

func validStrings(values []string, maximumCount, maximumBytes int) bool {
	if len(values) > maximumCount {
		return false
	}
	for _, value := range values {
		if !validSafeText(value, maximumBytes, true) {
			return false
		}
	}
	return true
}

func validOpaqueIDs(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validOpaqueID(value) {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validKeywords(keywords []EngineKeywordMarker) bool {
	seen := make(map[EngineKeywordMarker]struct{}, len(keywords))
	for _, keyword := range keywords {
		if !validSafeText(string(keyword), MaxEngineMIMEFieldBytes, true) {
			return false
		}
		if _, duplicate := seen[keyword]; duplicate {
			return false
		}
		seen[keyword] = struct{}{}
	}
	return true
}

func validMessageIDs(messageIDs []string) bool {
	if len(messageIDs) == 0 || len(messageIDs) > MaxEnginePageItems {
		return false
	}
	return validOpaqueIDs(messageIDs)
}

func validSpamSamples(samples []EngineSpamTrainingSample) bool {
	seen := make(map[string]struct{}, len(samples))
	for _, sample := range samples {
		if !validOpaqueID(sample.MessageID) {
			return false
		}
		if _, duplicate := seen[sample.MessageID]; duplicate {
			return false
		}
		seen[sample.MessageID] = struct{}{}
	}
	return true
}

func validateProjectionPage(page EngineProjectionPage) error {
	if !validAccountID(page.AccountID) || !validCursor(page.Cursor, false) || !validCursor(page.NextCursor, true) || len(page.Messages) > MaxEnginePageItems || page.Bounds.Returned != len(page.Messages) || page.Bounds.Maximum != MaxEnginePageItems || page.Bounds.Requested < 1 || page.Bounds.Requested > MaxEnginePageItems || projectionPageBytes(page.Messages, page.AccountID, page.Cursor, page.NextCursor) > MaxEngineProjectionPageBytes {
		return fakeEngineError(EngineErrorPayloadTooLarge)
	}
	for _, projection := range page.Messages {
		if err := validateProjection(projection); err != nil { return err }
	}
	return nil
}

func validateChangePage(page EngineChangePage, request EngineChangeRequest) error {
	if page.AccountID != request.AccountID || page.FromCursor != request.FromCursor || !validCursor(page.FromCursor, false) || !validCursor(page.NextCursor, true) || page.Bounds.Requested != request.Page.Limit || page.Bounds.Maximum != MaxEnginePageItems || page.Bounds.Returned != len(page.CreatedOrUpdated)+len(page.DestroyedIDs) || page.Bounds.Returned > MaxEnginePageItems || projectionPageBytes(page.CreatedOrUpdated, page.AccountID, page.FromCursor, page.NextCursor)+stringBytes(page.DestroyedIDs) > MaxEngineProjectionPageBytes {
		return fakeEngineError(EngineErrorPayloadTooLarge)
	}
	destroyed := make(map[string]struct{}, len(page.DestroyedIDs))
	for _, messageID := range page.DestroyedIDs {
		if !validOpaqueID(messageID) { return fakeEngineError(EngineErrorPayloadTooLarge) }
		if _, duplicate := destroyed[messageID]; duplicate { return fakeEngineError(EngineErrorPayloadTooLarge) }
		destroyed[messageID] = struct{}{}
	}
	for _, projection := range page.CreatedOrUpdated {
		if projection.AccountID != request.AccountID {
			return fakeEngineError(EngineErrorPayloadTooLarge)
		}
		if _, collision := destroyed[projection.MessageID]; collision { return fakeEngineError(EngineErrorPayloadTooLarge) }
		if err := validateProjection(projection); err != nil { return err }
	}
	return nil
}

func validateEnumerationPage(page EngineEnumerationPage, request EngineEnumerationRequest) error {
	if page.AccountID != request.AccountID || page.Checkpoint != request.Checkpoint || page.BaseCursor == "" || (request.BaseCursor != "" && page.BaseCursor != request.BaseCursor) || !validCursor(page.BaseCursor, false) || !validCursor(page.Checkpoint, true) || !validCursor(page.NextCheckpoint, true) || page.Bounds.Requested != request.Page.Limit || page.Bounds.Maximum != MaxEnginePageItems || page.Bounds.Returned != len(page.Messages) || page.Bounds.Returned > MaxEnginePageItems || (!page.Complete && page.NextCheckpoint == "") || (page.Complete && page.NextCheckpoint != "") || projectionPageBytes(page.Messages, page.AccountID, page.BaseCursor, page.Checkpoint, page.NextCheckpoint) > MaxEngineProjectionPageBytes {
		return fakeEngineError(EngineErrorPayloadTooLarge)
	}
	for _, projection := range page.Messages {
		if projection.AccountID != request.AccountID {
			return fakeEngineError(EngineErrorPayloadTooLarge)
		}
		if err := validateProjection(projection); err != nil { return err }
	}
	return nil
}

func parseListCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	if len(cursor) < len("list:") || cursor[:len("list:")] != "list:" {
		return 0, fakeEngineError(EngineErrorInvalidCursor)
	}
	value, err := strconv.Atoi(cursor[len("list:"):])
	if err != nil || value < 0 {
		return 0, fakeEngineError(EngineErrorInvalidCursor)
	}
	return value, nil
}

func parseChangesCursor(cursor string) (int, error) {
	if len(cursor) < len("sync:") || cursor[:len("sync:")] != "sync:" {
		return 0, fakeEngineError(EngineErrorInvalidCursor)
	}
	value, err := strconv.Atoi(cursor[len("sync:"):])
	if err != nil || value < 0 {
		return 0, fakeEngineError(EngineErrorInvalidCursor)
	}
	return value, nil
}

func syncCursor(offset int) string {
	return "sync:" + strconv.Itoa(offset)
}

func validSafeText(value string, maximum int, required bool) bool {
	if (required && value == "") || len(value) > maximum || !utf8.ValidString(value) {
		return false
	}
	for _, runeValue := range value {
		if runeValue < 0x20 || (runeValue >= 0x7f && runeValue <= 0x9f) {
			return false
		}
	}
	return true
}

func validOpaqueID(value string) bool {
	return validSafeText(value, MaxEngineOpaqueIDBytes, true) && !looksLikeUUID(value)
}

func validOpaqueIDOrEmpty(value string) bool {
	return value == "" || validOpaqueID(value)
}

func validAccountID(value string) bool {
	return validOpaqueID(value)
}

func validCursor(value string, allowEmpty bool) bool {
	return (allowEmpty && value == "") || validSafeText(value, MaxEngineCursorBytes, true)
}

func looksLikeUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for index := 0; index < len(value); index++ {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if (value[index] < '0' || value[index] > '9') && (value[index] < 'a' || value[index] > 'f') && (value[index] < 'A' || value[index] > 'F') {
			return false
		}
	}
	return true
}

func projectionHasFolder(projection EngineMessageProjection, folderID string) bool {
	for _, candidate := range projection.FolderIDs {
		if candidate == folderID {
			return true
		}
	}
	return false
}

func cloneDomain(state EngineDomainState) EngineDomainState {
	state.DNSRecords = append([]EngineDNSRecord{}, state.DNSRecords...)
	return state
}

func cloneAccount(account EngineAccount) EngineAccount {
	account.AliasIdentityIDs = cloneStrings(account.AliasIdentityIDs)
	return account
}

func cloneMessage(message EngineMessage) EngineMessage {
	message.Projection = cloneProjection(message.Projection)
	return message
}

func cloneProjection(projection EngineMessageProjection) EngineMessageProjection {
	projection.References = cloneStrings(projection.References)
	projection.To = cloneAddresses(projection.To)
	projection.CC = cloneAddresses(projection.CC)
	projection.BCC = cloneAddresses(projection.BCC)
	projection.ReplyTo = cloneAddresses(projection.ReplyTo)
	projection.FolderIDs = cloneStrings(projection.FolderIDs)
	projection.ReservedKeywords = append([]EngineKeywordMarker{}, projection.ReservedKeywords...)
	projection.Attachments = make([]EngineProjectionAttachment, len(projection.Attachments))
	for index, attachment := range projection.Attachments {
		attachment.SHA256 = append([]byte{}, attachment.SHA256...)
		projection.Attachments[index] = attachment
	}
	if projection.SentAt != nil {
		instant := *projection.SentAt
		projection.SentAt = &instant
	}
	if projection.ReceivedAt != nil {
		instant := *projection.ReceivedAt
		projection.ReceivedAt = &instant
	}
	projection.CalendarMethod = copyMailCalendarMethod(projection.CalendarMethod)
	if projection.CalendarPartLocatorHash != nil {
		projection.CalendarPartLocatorHash = append([]byte{}, projection.CalendarPartLocatorHash...)
	}
	return projection
}

func cloneBlob(blob EngineBlob) EngineBlob {
	blob.Bytes = append([]byte{}, blob.Bytes...)
	return blob
}

func cloneChangePage(page EngineChangePage) EngineChangePage {
	page.CreatedOrUpdated = cloneProjections(page.CreatedOrUpdated)
	page.DestroyedIDs = cloneStrings(page.DestroyedIDs)
	return page
}

func cloneEnumerationPage(page EngineEnumerationPage) EngineEnumerationPage {
	page.Messages = cloneProjections(page.Messages)
	return page
}

func cloneProjections(projections []EngineMessageProjection) []EngineMessageProjection {
	result := make([]EngineMessageProjection, len(projections))
	for index, projection := range projections {
		result[index] = cloneProjection(projection)
	}
	return result
}

func cloneAddresses(addresses []EngineAddress) []EngineAddress {
	return append([]EngineAddress{}, addresses...)
}

func cloneStrings(values []string) []string {
	return append([]string{}, values...)
}

func cloneSpamResult(result EngineSpamTrainingResult) EngineSpamTrainingResult {
	result.Outcomes = append([]EngineSpamTrainingOutcome{}, result.Outcomes...)
	return result
}

func messageVariableBytes(message EngineMessage) int {
	return projectionVariableBytes(message.Projection) + len(message.Body.PlainBody) + len(message.Body.RichBody)
}

func projectionVariableBytes(projection EngineMessageProjection) int {
	total := len(projection.AccountID) + len(projection.MessageID) + len(projection.Version) + len(projection.InternetMessageID) + len(projection.InReplyTo) + len(projection.Subject) + len(projection.Snippet) + addressBytes(projection.From) + stringBytes(projection.References) + addressSliceBytes(projection.To) + addressSliceBytes(projection.CC) + addressSliceBytes(projection.BCC) + addressSliceBytes(projection.ReplyTo) + stringBytes(projection.FolderIDs)
	if projection.CalendarMethod != nil {
		total += len(*projection.CalendarMethod)
	}
	for _, keyword := range projection.ReservedKeywords {
		total += len(keyword)
	}
	for _, attachment := range projection.Attachments {
		total += len(attachment.BlobID) + len(attachment.AttachmentKey) + len(attachment.ContentID) + len(attachment.Filename) + len(attachment.MediaType) + len(attachment.SHA256)
	}
	return total
}

func projectionPageBytes(projections []EngineMessageProjection, values ...string) int {
	total := stringBytes(values)
	for _, projection := range projections {
		total += projectionVariableBytes(projection)
	}
	return total
}

func addressBytes(address EngineAddress) int {
	return len(address.Email) + len(address.Name)
}

func addressSliceBytes(addresses []EngineAddress) int {
	total := 0
	for _, address := range addresses {
		total += addressBytes(address)
	}
	return total
}

func stringBytes(values []string) int {
	total := 0
	for _, value := range values {
		total += len(value)
	}
	return total
}
