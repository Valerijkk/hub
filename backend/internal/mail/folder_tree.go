package mail

import (
	"bytes"
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

type FolderTreeRole string

const (
	FolderTreeRoleInbox   FolderTreeRole = "inbox"
	FolderTreeRoleSent    FolderTreeRole = "sent"
	FolderTreeRoleDrafts  FolderTreeRole = "drafts"
	FolderTreeRoleArchive FolderTreeRole = "archive"
	FolderTreeRoleTrash   FolderTreeRole = "trash"
	FolderTreeRoleJunk    FolderTreeRole = "junk"
)

type FolderTreeFolder struct {
	AccountID uuid.UUID
	ID        uuid.UUID
	ParentID  uuid.UUID
	Role      FolderTreeRole
	Name      string
	SortOrder int32
}

type FolderTreeSnapshot struct {
	AccountID uuid.UUID
	Folders   []FolderTreeFolder
}

type FolderTreeNode struct {
	AccountID      uuid.UUID
	ID             uuid.UUID
	ParentID       uuid.UUID
	Role           FolderTreeRole
	Name           string
	NormalizedName string
	SortOrder      int32
}

type FolderTreeCreate struct {
	ID        uuid.UUID
	ParentID  uuid.UUID
	Name      string
	SortOrder int32
}

type FolderTreeRename struct {
	FolderID uuid.UUID
	Name     string
}

type FolderTreeReparent struct {
	FolderID uuid.UUID
	ParentID uuid.UUID
}

type FolderReferenceState string

const (
	FolderReferenceAbsent  FolderReferenceState = "absent"
	FolderReferencePresent FolderReferenceState = "present"
)

type FolderTreeDelete struct {
	FolderID                uuid.UUID
	ActiveMoveRuleReference FolderReferenceState
}

type FolderTreeChange struct {
	Create   *FolderTreeCreate
	Rename   *FolderTreeRename
	Reparent *FolderTreeReparent
	Delete   *FolderTreeDelete
}

type FolderTreeDecision struct {
	NormalizedName string
}

type FolderErrorKind string

const (
	FolderErrorInvalid           FolderErrorKind = "invalid"
	FolderErrorTooLarge          FolderErrorKind = "too_large"
	FolderErrorDuplicate         FolderErrorKind = "duplicate"
	FolderErrorNameConflict      FolderErrorKind = "name_conflict"
	FolderErrorAccountMismatch   FolderErrorKind = "account_mismatch"
	FolderErrorNotFound          FolderErrorKind = "not_found"
	FolderErrorParentUnavailable FolderErrorKind = "parent_unavailable"
	FolderErrorSelfParent        FolderErrorKind = "self_parent"
	FolderErrorCycle             FolderErrorKind = "cycle"
	FolderErrorDepthExceeded     FolderErrorKind = "depth_exceeded"
	FolderErrorSystemRole        FolderErrorKind = "system_role"
	FolderErrorInUse             FolderErrorKind = "in_use"
	FolderErrorLimitExceeded     FolderErrorKind = "limit_exceeded"
)

var (
	ErrFolderInvalid           = errors.New("mail folder invalid")
	ErrFolderTooLarge          = errors.New("mail folder exceeds bounds")
	ErrFolderDuplicate         = errors.New("mail folder duplicate")
	ErrFolderNameConflict      = errors.New("mail folder name conflict")
	ErrFolderAccountMismatch   = errors.New("mail folder account mismatch")
	ErrFolderNotFound          = errors.New("mail folder not found")
	ErrFolderParentUnavailable = errors.New("mail folder parent unavailable")
	ErrFolderSelfParent        = errors.New("mail folder self parent")
	ErrFolderCycle             = errors.New("mail folder cycle")
	ErrFolderDepthExceeded     = errors.New("mail folder depth exceeds bounds")
	ErrFolderSystemRole        = errors.New("mail folder system role protected")
	ErrFolderInUse             = errors.New("mail folder in use")
	ErrFolderLimitExceeded     = errors.New("mail folder limit exceeded")
)

type FolderPolicyError struct {
	Kind FolderErrorKind
}

func (e *FolderPolicyError) Error() string {
	if e == nil {
		return "mail folder policy error"
	}
	switch e.Kind {
	case FolderErrorInvalid:
		return "mail folder invalid"
	case FolderErrorTooLarge:
		return "mail folder exceeds bounds"
	case FolderErrorDuplicate:
		return "mail folder duplicate"
	case FolderErrorNameConflict:
		return "mail folder name conflict"
	case FolderErrorAccountMismatch:
		return "mail folder account mismatch"
	case FolderErrorNotFound:
		return "mail folder not found"
	case FolderErrorParentUnavailable:
		return "mail folder parent unavailable"
	case FolderErrorSelfParent:
		return "mail folder self parent"
	case FolderErrorCycle:
		return "mail folder cycle"
	case FolderErrorDepthExceeded:
		return "mail folder depth exceeds bounds"
	case FolderErrorSystemRole:
		return "mail folder system role protected"
	case FolderErrorInUse:
		return "mail folder in use"
	case FolderErrorLimitExceeded:
		return "mail folder limit exceeded"
	default:
		return "mail folder policy error"
	}
}

func (e *FolderPolicyError) Unwrap() error {
	if e == nil {
		return nil
	}
	switch e.Kind {
	case FolderErrorInvalid:
		return ErrFolderInvalid
	case FolderErrorTooLarge:
		return ErrFolderTooLarge
	case FolderErrorDuplicate:
		return ErrFolderDuplicate
	case FolderErrorNameConflict:
		return ErrFolderNameConflict
	case FolderErrorAccountMismatch:
		return ErrFolderAccountMismatch
	case FolderErrorNotFound:
		return ErrFolderNotFound
	case FolderErrorParentUnavailable:
		return ErrFolderParentUnavailable
	case FolderErrorSelfParent:
		return ErrFolderSelfParent
	case FolderErrorCycle:
		return ErrFolderCycle
	case FolderErrorDepthExceeded:
		return ErrFolderDepthExceeded
	case FolderErrorSystemRole:
		return ErrFolderSystemRole
	case FolderErrorInUse:
		return ErrFolderInUse
	case FolderErrorLimitExceeded:
		return ErrFolderLimitExceeded
	default:
		return nil
	}
}

type indexedFolder struct {
	folder FolderTreeFolder
	key    string
}

type snapshotIndex struct {
	rows       map[uuid.UUID]indexedFolder
	children   map[uuid.UUID][]uuid.UUID
	systemRole map[FolderTreeRole]uuid.UUID
}

func NormalizeFolderTree(snapshot FolderTreeSnapshot) ([]FolderTreeNode, error) {
	index, err := validateSnapshot(snapshot)
	if err != nil {
		return nil, err
	}
	out := make([]FolderTreeNode, 0, len(index.rows))
	for _, rootID := range orderedRoots(index) {
		appendPreorder(&out, index, rootID)
	}
	return out, nil
}

func EvaluateFolderTreeChange(snapshot FolderTreeSnapshot, change FolderTreeChange) (FolderTreeDecision, error) {
	if countChangeArms(change) != 1 {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorInvalid}
	}
	index, err := validateSnapshot(snapshot)
	if err != nil {
		return FolderTreeDecision{}, err
	}
	if change.Create != nil {
		return evaluateCreate(index, *change.Create)
	}
	target := changeTarget(change)
	if target == uuid.Nil {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorInvalid}
	}
	record, ok := index.rows[target]
	if !ok {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorNotFound}
	}
	if record.folder.Role != "" {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorSystemRole}
	}
	if change.Rename != nil {
		return evaluateRename(index, target, change.Rename.Name)
	}
	if change.Reparent != nil {
		return evaluateReparent(index, target, change.Reparent.ParentID)
	}
	return evaluateDelete(index, target, change.Delete.ActiveMoveRuleReference)
}

func normalizeFolderName(raw string) (string, error) {
	if !utf8.ValidString(raw) || strings.IndexByte(raw, 0) >= 0 {
		return "", &FolderPolicyError{Kind: FolderErrorInvalid}
	}
	runes := 0
	for _, r := range raw {
		if unicode.IsControl(r) {
			return "", &FolderPolicyError{Kind: FolderErrorInvalid}
		}
		runes++
	}
	if runes == 0 {
		return "", &FolderPolicyError{Kind: FolderErrorInvalid}
	}
	if runes > 120 || len(raw) > 480 {
		return "", &FolderPolicyError{Kind: FolderErrorTooLarge}
	}
	var b strings.Builder
	b.Grow(len(raw))
	wrote, gap := false, false
	for _, r := range raw {
		if unicode.IsSpace(r) {
			if wrote {
				gap = true
			}
			continue
		}
		if gap {
			b.WriteByte(' ')
			gap = false
		}
		b.WriteRune(r)
		wrote = true
	}
	key := norm.NFC.String(cases.Fold().String(norm.NFC.String(b.String())))
	if key == "" {
		return "", &FolderPolicyError{Kind: FolderErrorInvalid}
	}
	if len(key) > 480 {
		return "", &FolderPolicyError{Kind: FolderErrorTooLarge}
	}
	return key, nil
}

func validateSnapshot(snapshot FolderTreeSnapshot) (snapshotIndex, error) {
	if snapshot.AccountID == uuid.Nil {
		return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorInvalid}
	}
	if len(snapshot.Folders) > 100 {
		return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorTooLarge}
	}
	index := snapshotIndex{
		rows:       make(map[uuid.UUID]indexedFolder, len(snapshot.Folders)),
		children:   make(map[uuid.UUID][]uuid.UUID, len(snapshot.Folders)),
		systemRole: make(map[FolderTreeRole]uuid.UUID, 6),
	}
	validated := make([]indexedFolder, 0, len(snapshot.Folders))
	for _, folder := range snapshot.Folders {
		if folder.AccountID == uuid.Nil || folder.ID == uuid.Nil {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorInvalid}
		}
		if folder.AccountID != snapshot.AccountID {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorAccountMismatch}
		}
		key, err := normalizeFolderName(folder.Name)
		if err != nil {
			return snapshotIndex{}, err
		}
		if !isSupportedRole(folder.Role) {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorInvalid}
		}
		if folder.Role != "" && folder.ParentID != uuid.Nil {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorInvalid}
		}
		validated = append(validated, indexedFolder{folder: folder, key: key})
	}
	for _, record := range validated {
		if _, exists := index.rows[record.folder.ID]; exists {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorDuplicate}
		}
		index.rows[record.folder.ID] = record
	}
	for _, record := range validated {
		if record.folder.Role != "" {
			if _, exists := index.systemRole[record.folder.Role]; exists {
				return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorDuplicate}
			}
			index.systemRole[record.folder.Role] = record.folder.ID
		}
	}
	for _, folder := range snapshot.Folders {
		if folder.Role != "" || folder.ParentID == uuid.Nil {
			continue
		}
		if folder.ParentID == folder.ID {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorSelfParent}
		}
		parent, exists := index.rows[folder.ParentID]
		if !exists {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorParentUnavailable}
		}
		if parent.folder.AccountID != snapshot.AccountID {
			return snapshotIndex{}, &FolderPolicyError{Kind: FolderErrorAccountMismatch}
		}
		index.children[folder.ParentID] = append(index.children[folder.ParentID], folder.ID)
	}
	if err := validateCycles(index); err != nil {
		return snapshotIndex{}, err
	}
	if err := validateDepth(index); err != nil {
		return snapshotIndex{}, err
	}
	if err := validateCustomNameCollisions(index); err != nil {
		return snapshotIndex{}, err
	}
	sortChildren(index)
	return index, nil
}

func isSupportedRole(role FolderTreeRole) bool {
	if role == "" {
		return true
	}
	switch role {
	case FolderTreeRoleInbox, FolderTreeRoleSent, FolderTreeRoleDrafts, FolderTreeRoleArchive, FolderTreeRoleTrash, FolderTreeRoleJunk:
		return true
	default:
		return false
	}
}

func validateCycles(index snapshotIndex) error {
	state := make(map[uuid.UUID]uint8, len(index.rows))
	var visit func(uuid.UUID) error
	visit = func(id uuid.UUID) error {
		switch state[id] {
		case 1:
			return &FolderPolicyError{Kind: FolderErrorCycle}
		case 2:
			return nil
		}
		state[id] = 1
		for _, childID := range index.children[id] {
			if err := visit(childID); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id := range index.rows {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func validateDepth(index snapshotIndex) error {
	depths := make(map[uuid.UUID]int, len(index.rows))
	var depth func(uuid.UUID) (int, error)
	depth = func(id uuid.UUID) (int, error) {
		if value, ok := depths[id]; ok {
			return value, nil
		}
		record := index.rows[id]
		if record.folder.ParentID == uuid.Nil {
			depths[id] = 0
			return 0, nil
		}
		parentDepth, err := depth(record.folder.ParentID)
		if err != nil {
			return 0, err
		}
		value := parentDepth + 1
		if value >= 3 {
			return 0, &FolderPolicyError{Kind: FolderErrorDepthExceeded}
		}
		depths[id] = value
		return value, nil
	}
	for id := range index.rows {
		if _, err := depth(id); err != nil {
			return err
		}
	}
	return nil
}

func validateCustomNameCollisions(index snapshotIndex) error {
	keys := make(map[uuid.UUID]map[string]struct{}, len(index.rows))
	for _, record := range index.rows {
		if record.folder.Role != "" {
			continue
		}
		byName := keys[record.folder.ParentID]
		if byName == nil {
			byName = make(map[string]struct{})
			keys[record.folder.ParentID] = byName
		}
		if _, exists := byName[record.key]; exists {
			return &FolderPolicyError{Kind: FolderErrorNameConflict}
		}
		byName[record.key] = struct{}{}
	}
	return nil
}

func sortChildren(index snapshotIndex) {
	for parentID := range index.children {
		ids := index.children[parentID]
		sort.Slice(ids, func(i, j int) bool {
			return compareIndexedFolders(index, ids[i], ids[j]) < 0
		})
	}
}

func compareIndexedFolders(index snapshotIndex, leftID, rightID uuid.UUID) int {
	left := index.rows[leftID]
	right := index.rows[rightID]
	if left.folder.SortOrder < right.folder.SortOrder {
		return -1
	}
	if left.folder.SortOrder > right.folder.SortOrder {
		return 1
	}
	if result := bytes.Compare([]byte(left.key), []byte(right.key)); result != 0 {
		return result
	}
	return bytes.Compare(leftID[:], rightID[:])
}

func orderedRoots(index snapshotIndex) []uuid.UUID {
	roots := make([]uuid.UUID, 0, len(index.rows))
	for _, role := range [6]FolderTreeRole{
		FolderTreeRoleInbox,
		FolderTreeRoleSent,
		FolderTreeRoleDrafts,
		FolderTreeRoleArchive,
		FolderTreeRoleTrash,
		FolderTreeRoleJunk,
	} {
		if id, exists := index.systemRole[role]; exists {
			roots = append(roots, id)
		}
	}
	customRoots := make([]uuid.UUID, 0, len(index.rows))
	for id, record := range index.rows {
		if record.folder.Role == "" && record.folder.ParentID == uuid.Nil {
			customRoots = append(customRoots, id)
		}
	}
	sort.Slice(customRoots, func(i, j int) bool {
		return compareIndexedFolders(index, customRoots[i], customRoots[j]) < 0
	})
	return append(roots, customRoots...)
}

func appendPreorder(out *[]FolderTreeNode, index snapshotIndex, id uuid.UUID) {
	record := index.rows[id]
	folder := record.folder
	*out = append(*out, FolderTreeNode{
		AccountID:      folder.AccountID,
		ID:             folder.ID,
		ParentID:       folder.ParentID,
		Role:           folder.Role,
		Name:           folder.Name,
		NormalizedName: record.key,
		SortOrder:      folder.SortOrder,
	})
	for _, childID := range index.children[id] {
		appendPreorder(out, index, childID)
	}
}

func countChangeArms(change FolderTreeChange) int {
	count := 0
	if change.Create != nil {
		count++
	}
	if change.Rename != nil {
		count++
	}
	if change.Reparent != nil {
		count++
	}
	if change.Delete != nil {
		count++
	}
	return count
}

func changeTarget(change FolderTreeChange) uuid.UUID {
	if change.Rename != nil {
		return change.Rename.FolderID
	}
	if change.Reparent != nil {
		return change.Reparent.FolderID
	}
	if change.Delete != nil {
		return change.Delete.FolderID
	}
	return uuid.Nil
}

func evaluateCreate(index snapshotIndex, create FolderTreeCreate) (FolderTreeDecision, error) {
	if create.ID == uuid.Nil {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorInvalid}
	}
	if len(create.Name) > 480 {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorTooLarge}
	}
	if len(index.rows) == 100 {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorLimitExceeded}
	}
	key, err := normalizeFolderName(create.Name)
	if err != nil {
		return FolderTreeDecision{}, err
	}
	if _, exists := index.rows[create.ID]; exists {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorDuplicate}
	}
	if err := validateProspectiveParent(index, create.ID, create.ParentID); err != nil {
		return FolderTreeDecision{}, err
	}
	if err := validateProspectiveDepth(index, create.ParentID); err != nil {
		return FolderTreeDecision{}, err
	}
	if customSiblingNameConflict(index, create.ParentID, key, uuid.Nil) {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorNameConflict}
	}
	return FolderTreeDecision{NormalizedName: key}, nil
}

func evaluateRename(index snapshotIndex, target uuid.UUID, name string) (FolderTreeDecision, error) {
	key, err := normalizeFolderName(name)
	if err != nil {
		return FolderTreeDecision{}, err
	}
	if customSiblingNameConflict(index, index.rows[target].folder.ParentID, key, target) {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorNameConflict}
	}
	return FolderTreeDecision{NormalizedName: key}, nil
}

func evaluateReparent(index snapshotIndex, target, parentID uuid.UUID) (FolderTreeDecision, error) {
	if err := validateProspectiveParent(index, target, parentID); err != nil {
		return FolderTreeDecision{}, err
	}
	overlay := snapshotIndex{
		rows:     make(map[uuid.UUID]indexedFolder, len(index.rows)),
		children: make(map[uuid.UUID][]uuid.UUID, len(index.rows)),
	}
	for id, record := range index.rows {
		overlay.rows[id] = record
	}
	record := overlay.rows[target]
	record.folder.ParentID = parentID
	overlay.rows[target] = record
	rebuildChildren(&overlay)
	if err := validateCycles(overlay); err != nil {
		return FolderTreeDecision{}, err
	}
	if err := validateDepth(overlay); err != nil {
		return FolderTreeDecision{}, err
	}
	if err := validateCustomNameCollisions(overlay); err != nil {
		return FolderTreeDecision{}, err
	}
	return FolderTreeDecision{}, nil
}

func evaluateDelete(index snapshotIndex, target uuid.UUID, reference FolderReferenceState) (FolderTreeDecision, error) {
	if len(index.children[target]) > 0 {
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorInUse}
	}
	switch reference {
	case FolderReferenceAbsent:
		return FolderTreeDecision{}, nil
	case FolderReferencePresent:
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorInUse}
	default:
		return FolderTreeDecision{}, &FolderPolicyError{Kind: FolderErrorInvalid}
	}
}

func validateProspectiveParent(index snapshotIndex, id, parentID uuid.UUID) error {
	if parentID == uuid.Nil {
		return nil
	}
	if parentID == id {
		return &FolderPolicyError{Kind: FolderErrorSelfParent}
	}
	if _, exists := index.rows[parentID]; !exists {
		return &FolderPolicyError{Kind: FolderErrorParentUnavailable}
	}
	return nil
}

func validateProspectiveDepth(index snapshotIndex, parentID uuid.UUID) error {
	if parentID == uuid.Nil {
		return nil
	}
	depths := make(map[uuid.UUID]int, len(index.rows))
	depth, err := folderDepth(index, parentID, depths)
	if err != nil {
		return err
	}
	if depth+1 >= 3 {
		return &FolderPolicyError{Kind: FolderErrorDepthExceeded}
	}
	return nil
}

func folderDepth(index snapshotIndex, id uuid.UUID, depths map[uuid.UUID]int) (int, error) {
	if value, ok := depths[id]; ok {
		return value, nil
	}
	record, exists := index.rows[id]
	if !exists {
		return 0, &FolderPolicyError{Kind: FolderErrorParentUnavailable}
	}
	if record.folder.ParentID == uuid.Nil {
		depths[id] = 0
		return 0, nil
	}
	parentDepth, err := folderDepth(index, record.folder.ParentID, depths)
	if err != nil {
		return 0, err
	}
	depth := parentDepth + 1
	depths[id] = depth
	return depth, nil
}

func customSiblingNameConflict(index snapshotIndex, parentID uuid.UUID, key string, excludedID uuid.UUID) bool {
	for id, record := range index.rows {
		if id == excludedID || record.folder.Role != "" {
			continue
		}
		if record.folder.ParentID == parentID && record.key == key {
			return true
		}
	}
	return false
}

func rebuildChildren(index *snapshotIndex) {
	for id, record := range index.rows {
		if record.folder.Role == "" && record.folder.ParentID != uuid.Nil {
			index.children[record.folder.ParentID] = append(index.children[record.folder.ParentID], id)
		}
	}
}
