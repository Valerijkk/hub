# 144-03B Folder Tree Policy Implementation Plan

> **Current Task 4 authority (2026-08-14):** Task 1 through Task 3 instructions below are
> `CONSUMED/HISTORICAL`. Source/oracle commit `f8cb7643986fc591cae16764575337b72ef3a692`
> (parent `32e3dcb5b4f7ae65a88b79781646d25726ae16c6`, tree
> `14e7f6ce9a0f4b7912b955515d122c411159c579`) is narrowly `STATIC_ACCEPTED`, static-only,
> after three exact-SHA PASS reviews C/I/M `0/0/0`. The Task 4 candidate is exactly 20 paths:
> 16 existing paths from literal Step 1 discovery, plus three existing stale-gate implementation
> evidence documents, plus `IMPLEMENTATION-144-03B-FOLDER.md` as one new evidence file. This
> yields the exact 20-path candidate and creates no product decision. Candidate reviews precede only root
> freeze/adjudication of `144-03B-Rule`; no Rule TDD or production authorization.

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## Goal

After all prerequisite acceptance gates, deliver only the external Folder-tree oracle and the pure two-callable policy implementation, then separately synchronize lifecycle evidence without claiming executable validation.

## Architecture

The first packet is an external `mail_test` oracle that imports only the public `internal/mail` surface and independently inspects the later source AST. The second packet is one pure policy file that validates a bounded Hub UUID snapshot, builds private indexes and overlays, and produces fresh deterministic output. The third packet commits exactly the source and oracle; a fourth, immediate docs-only packet records evidence and lifecycle after that source boundary has three exact-SHA reviews.

## Tech Stack

Go 1.25; `github.com/google/uuid v1.6.0`; `golang.org/x/text v0.35.0` using `cases` and `unicode/norm`; production imports limited to `bytes`, `errors`, `sort`, `strings`, `unicode`, `unicode/utf8`, `github.com/google/uuid`, `golang.org/x/text/cases`, and `golang.org/x/text/unicode/norm`.

## Global Constraints

- Contract commit `a565af2632bc5f76a7d2fe26ebf3065fb3cefba9` (blob `0f9ed0adb0cec6ffa83619cb4c7c4cb4d3f3263c`) is narrowly `STATIC_ACCEPTED` after three independent exact-SHA reviews PASS with C/I/M `0/0/0`.
- Plan commit `a98c7235521852dee1ebc0d09558c7bcbefc4534` (blob `c79658d61e496489be697bc971d0797ed30fb61f`) is narrowly `EXACT_SHA_PASS` / `STATIC_ACCEPTED` after exact, security and boundary reviews PASS with C/I/M `0/0/0`. Folder source+oracle are now statically accepted at `f8cb7643986fc591cae16764575337b72ef3a692`; this exact 20-path docs candidate is pending its own three exact-SHA reviews, after which root only freezes/adjudicates `144-03B-Rule`.
- The production boundary contains exactly two paths: `backend/tests/mail/folder_tree_test.go` first, then `backend/internal/mail/folder_tree.go`. No dependency, schema, handler, repository, engine, migration, generated, configuration, or additional source change is allowed.
- Hub identifiers are `uuid.UUID`. Snapshot AccountID, row AccountID/ID, all action IDs, and target IDs reject `uuid.Nil`; only ParentID `uuid.Nil` means root. Engine opaque IDs are mapped to Hub UUID projection rows before either public callable.
- Snapshot input includes every live non-deleted control row, including system and reserved creating/unknown/error/deleting rows. It is bounded to 100 rows. Deleted tombstones are excluded by caller. The separate engine enumeration bound does not alter this pure policy.
- Raw name validity is UTF-8, no Unicode control/NUL, 1..120 runes, <=480 bytes. The private key collapses every `unicode.IsSpace` run to one ASCII space, trims, NFC-normalizes, folds with `cases.Fold` from x/text v0.35.0, final-NFC normalizes, remains nonempty and <=480 bytes, and never enters OpenAPI.
- Snapshot order is exact: fixed system root rank Inbox, Sent, Drafts, Archive, Trash, Junk; then custom roots; every sibling uses SortOrder, normalized-key UTF-8 byte order, then raw 16-byte UUID order; traversal is depth-first and returned slice is fresh/non-nil on successful empty input.
- Create precedence is exact: union shape; complete snapshot; nonzero Create.ID; raw candidate name byte bound; capacity; remaining candidate name semantics, parent, depth and collision. Zero ID at 100 is invalid; nonzero ID and raw name >480 bytes at 100 is too_large; nonzero ID and malformed <=480-byte name at 100 is limit_exceeded.
- Rename/Reparent/Delete resolve exact target after full snapshot validation. System target returns system_role before name, parent, rule-reference or children semantics; Delete system_role precedes in_use. Create is custom/no target. Delete never removes a row, frees a slot, or changes lifecycle.
- Every executable test, build, formatter, linter, validator and Docker command is **NOT AUTHORIZED / NOT_RUN** until an accepted versioned task-owned runner exists. Host `go test`, `go build`, `gofmt`, lint and direct Docker/Compose are prohibited.
- Allowed read-only static evidence commands are `git status`, `git diff --check`, `git diff`, `git diff-tree`, `git hash-object`, `git show`, `git rev-parse` and scoped `rg`. Root alone stages and commits literal reviewed paths, with no AI trailers.

---

## Prerequisite acceptance gate

No Task 1 author is dispatched until every row is true.

| Gate | Required evidence | Owner | Current planning status |
|---|---|---|---|
| Contract lifecycle acceptance | `a565af2632bc5f76a7d2fe26ebf3065fb3cefba9` / blob `0f9ed0adb0cec6ffa83619cb4c7c4cb4d3f3263c` is narrowly `STATIC_ACCEPTED` after three independent exact-SHA PASS reviews, C/I/M `0/0/0`; this sync still needs its own three exact-SHA PASS reviews | root | contract accepted; sync pending |
| Plan acceptance | `a98c7235521852dee1ebc0d09558c7bcbefc4534` / blob `c79658d61e496489be697bc971d0797ed30fb61f` is narrowly `EXACT_SHA_PASS` / `STATIC_ACCEPTED` after exact, security and boundary PASS reviews, C/I/M `0/0/0` | root | accepted |
| Runner truth | Versioned task-owned runner status remains `NOT_READY`; compile/runtime stay `NOT_RUN` | root | fixed blocker |
| Dispatch provenance | Every later dispatch is native `gpt-5.6-terra`, explicit effort, no fallback, `dispatcher-enforced`; actual telemetry `UNKNOWN` unless exposed | root | required |

The first two rows are authorization gates, not documentation wishes. Failure of either leaves Task 1 forbidden.

## Frozen public ABI

```go
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

func (e *FolderPolicyError) Error() string
func (e *FolderPolicyError) Unwrap() error
func NormalizeFolderTree(FolderTreeSnapshot) ([]FolderTreeNode, error)
func EvaluateFolderTreeChange(FolderTreeSnapshot, FolderTreeChange) (FolderTreeDecision, error)
```

## Fixed deterministic fixtures

Every test uses only these `uuid.MustParse` literals. It never uses `uuid.New`, time, random values, or generated identifiers.

| Fixture | UUID literal | Use |
|---|---|---|
| account | `00000000-0000-4000-8000-000000000001` | every same-account snapshot |
| foreign account | `00000000-0000-4000-8000-000000000002` | account mismatch |
| unavailable parent | `00000000-0000-4000-8000-000000000003` | nonzero parent absent from snapshot |
| unavailable mutation target | `00000000-0000-4000-8000-000000000004` | nonzero Rename/Delete target absent from snapshot |
| root parent | `uuid.Nil` | sole root marker |
| system inbox | `00000000-0000-4000-8000-000000000101` | fixed rank root |
| system sent | `00000000-0000-4000-8000-000000000102` | fixed rank root |
| system drafts | `00000000-0000-4000-8000-000000000103` | fixed rank root |
| system archive | `00000000-0000-4000-8000-000000000104` | fixed rank root |
| system trash | `00000000-0000-4000-8000-000000000105` | fixed rank root |
| system junk | `00000000-0000-4000-8000-000000000106` | fixed rank root |
| custom roots/children | `00000000-0000-4000-8000-000000000107` through `00000000-0000-4000-8000-000000000201` | fixed capacity, graph, collision and ordering cases |

The capacity fixture is this literal array. The first 100 values form the complete 100-slot snapshot; the 101st value forms the 101-row snapshot.

```go
var (
    fixtureAccountID        = uuid.MustParse("00000000-0000-4000-8000-000000000001")
    fixtureForeignAccountID = uuid.MustParse("00000000-0000-4000-8000-000000000002")
    fixtureMissingParentID  = uuid.MustParse("00000000-0000-4000-8000-000000000003")
    fixtureMissingTargetID  = uuid.MustParse("00000000-0000-4000-8000-000000000004")
)

var fixedFolderIDs = []uuid.UUID{
    uuid.MustParse("00000000-0000-4000-8000-000000000101"),
    uuid.MustParse("00000000-0000-4000-8000-000000000102"),
    uuid.MustParse("00000000-0000-4000-8000-000000000103"),
    uuid.MustParse("00000000-0000-4000-8000-000000000104"),
    uuid.MustParse("00000000-0000-4000-8000-000000000105"),
    uuid.MustParse("00000000-0000-4000-8000-000000000106"),
    uuid.MustParse("00000000-0000-4000-8000-000000000107"),
    uuid.MustParse("00000000-0000-4000-8000-000000000108"),
    uuid.MustParse("00000000-0000-4000-8000-000000000109"),
    uuid.MustParse("00000000-0000-4000-8000-000000000110"),
    uuid.MustParse("00000000-0000-4000-8000-000000000111"),
    uuid.MustParse("00000000-0000-4000-8000-000000000112"),
    uuid.MustParse("00000000-0000-4000-8000-000000000113"),
    uuid.MustParse("00000000-0000-4000-8000-000000000114"),
    uuid.MustParse("00000000-0000-4000-8000-000000000115"),
    uuid.MustParse("00000000-0000-4000-8000-000000000116"),
    uuid.MustParse("00000000-0000-4000-8000-000000000117"),
    uuid.MustParse("00000000-0000-4000-8000-000000000118"),
    uuid.MustParse("00000000-0000-4000-8000-000000000119"),
    uuid.MustParse("00000000-0000-4000-8000-000000000120"),
    uuid.MustParse("00000000-0000-4000-8000-000000000121"),
    uuid.MustParse("00000000-0000-4000-8000-000000000122"),
    uuid.MustParse("00000000-0000-4000-8000-000000000123"),
    uuid.MustParse("00000000-0000-4000-8000-000000000124"),
    uuid.MustParse("00000000-0000-4000-8000-000000000125"),
    uuid.MustParse("00000000-0000-4000-8000-000000000126"),
    uuid.MustParse("00000000-0000-4000-8000-000000000127"),
    uuid.MustParse("00000000-0000-4000-8000-000000000128"),
    uuid.MustParse("00000000-0000-4000-8000-000000000129"),
    uuid.MustParse("00000000-0000-4000-8000-000000000130"),
    uuid.MustParse("00000000-0000-4000-8000-000000000131"),
    uuid.MustParse("00000000-0000-4000-8000-000000000132"),
    uuid.MustParse("00000000-0000-4000-8000-000000000133"),
    uuid.MustParse("00000000-0000-4000-8000-000000000134"),
    uuid.MustParse("00000000-0000-4000-8000-000000000135"),
    uuid.MustParse("00000000-0000-4000-8000-000000000136"),
    uuid.MustParse("00000000-0000-4000-8000-000000000137"),
    uuid.MustParse("00000000-0000-4000-8000-000000000138"),
    uuid.MustParse("00000000-0000-4000-8000-000000000139"),
    uuid.MustParse("00000000-0000-4000-8000-000000000140"),
    uuid.MustParse("00000000-0000-4000-8000-000000000141"),
    uuid.MustParse("00000000-0000-4000-8000-000000000142"),
    uuid.MustParse("00000000-0000-4000-8000-000000000143"),
    uuid.MustParse("00000000-0000-4000-8000-000000000144"),
    uuid.MustParse("00000000-0000-4000-8000-000000000145"),
    uuid.MustParse("00000000-0000-4000-8000-000000000146"),
    uuid.MustParse("00000000-0000-4000-8000-000000000147"),
    uuid.MustParse("00000000-0000-4000-8000-000000000148"),
    uuid.MustParse("00000000-0000-4000-8000-000000000149"),
    uuid.MustParse("00000000-0000-4000-8000-000000000150"),
    uuid.MustParse("00000000-0000-4000-8000-000000000151"),
    uuid.MustParse("00000000-0000-4000-8000-000000000152"),
    uuid.MustParse("00000000-0000-4000-8000-000000000153"),
    uuid.MustParse("00000000-0000-4000-8000-000000000154"),
    uuid.MustParse("00000000-0000-4000-8000-000000000155"),
    uuid.MustParse("00000000-0000-4000-8000-000000000156"),
    uuid.MustParse("00000000-0000-4000-8000-000000000157"),
    uuid.MustParse("00000000-0000-4000-8000-000000000158"),
    uuid.MustParse("00000000-0000-4000-8000-000000000159"),
    uuid.MustParse("00000000-0000-4000-8000-000000000160"),
    uuid.MustParse("00000000-0000-4000-8000-000000000161"),
    uuid.MustParse("00000000-0000-4000-8000-000000000162"),
    uuid.MustParse("00000000-0000-4000-8000-000000000163"),
    uuid.MustParse("00000000-0000-4000-8000-000000000164"),
    uuid.MustParse("00000000-0000-4000-8000-000000000165"),
    uuid.MustParse("00000000-0000-4000-8000-000000000166"),
    uuid.MustParse("00000000-0000-4000-8000-000000000167"),
    uuid.MustParse("00000000-0000-4000-8000-000000000168"),
    uuid.MustParse("00000000-0000-4000-8000-000000000169"),
    uuid.MustParse("00000000-0000-4000-8000-000000000170"),
    uuid.MustParse("00000000-0000-4000-8000-000000000171"),
    uuid.MustParse("00000000-0000-4000-8000-000000000172"),
    uuid.MustParse("00000000-0000-4000-8000-000000000173"),
    uuid.MustParse("00000000-0000-4000-8000-000000000174"),
    uuid.MustParse("00000000-0000-4000-8000-000000000175"),
    uuid.MustParse("00000000-0000-4000-8000-000000000176"),
    uuid.MustParse("00000000-0000-4000-8000-000000000177"),
    uuid.MustParse("00000000-0000-4000-8000-000000000178"),
    uuid.MustParse("00000000-0000-4000-8000-000000000179"),
    uuid.MustParse("00000000-0000-4000-8000-000000000180"),
    uuid.MustParse("00000000-0000-4000-8000-000000000181"),
    uuid.MustParse("00000000-0000-4000-8000-000000000182"),
    uuid.MustParse("00000000-0000-4000-8000-000000000183"),
    uuid.MustParse("00000000-0000-4000-8000-000000000184"),
    uuid.MustParse("00000000-0000-4000-8000-000000000185"),
    uuid.MustParse("00000000-0000-4000-8000-000000000186"),
    uuid.MustParse("00000000-0000-4000-8000-000000000187"),
    uuid.MustParse("00000000-0000-4000-8000-000000000188"),
    uuid.MustParse("00000000-0000-4000-8000-000000000189"),
    uuid.MustParse("00000000-0000-4000-8000-000000000190"),
    uuid.MustParse("00000000-0000-4000-8000-000000000191"),
    uuid.MustParse("00000000-0000-4000-8000-000000000192"),
    uuid.MustParse("00000000-0000-4000-8000-000000000193"),
    uuid.MustParse("00000000-0000-4000-8000-000000000194"),
    uuid.MustParse("00000000-0000-4000-8000-000000000195"),
    uuid.MustParse("00000000-0000-4000-8000-000000000196"),
    uuid.MustParse("00000000-0000-4000-8000-000000000197"),
    uuid.MustParse("00000000-0000-4000-8000-000000000198"),
    uuid.MustParse("00000000-0000-4000-8000-000000000199"),
    uuid.MustParse("00000000-0000-4000-8000-000000000200"),
    uuid.MustParse("00000000-0000-4000-8000-000000000201"),
}
```

## Task 1: External RED oracle only

**Files:**
- Create: `backend/tests/mail/folder_tree_test.go`
- Do not modify: `backend/internal/mail/folder_tree.go` or any document in this task.

**Author route:** fresh `gpt-5.6-terra` at `xhigh`, explicit no fallback, `dispatcher-enforced`; actual telemetry `UNKNOWN` unless exposed.

**Consumes:** the frozen ABI and fixed UUID fixtures above.

**Produces:** only an external `package mail_test` RED oracle. Compile and runtime remain `NOT_RUN`; no commit occurs here.

- [ ] **Step 1: Declare external imports, fixed fixture array, exact callable assignments and reflection contract.**

```go
package mail_test

import (
    "bytes"
    "errors"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "reflect"
    "runtime"
    "strings"
    "testing"
    "unicode/utf8"

    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    "github.com/valerii/hub/backend/internal/mail"
)

var (
    _ func(mail.FolderTreeSnapshot) ([]mail.FolderTreeNode, error) = mail.NormalizeFolderTree
    _ func(mail.FolderTreeSnapshot, mail.FolderTreeChange) (mail.FolderTreeDecision, error) = mail.EvaluateFolderTreeChange
)

func fields(t *testing.T, value any, names []string, types []reflect.Type) {
    t.Helper()
    typ := reflect.TypeOf(value)
    require.Equal(t, len(names), typ.NumField())
    for i := range names {
        f := typ.Field(i)
        require.Equal(t, names[i], f.Name)
        require.Equal(t, types[i], f.Type)
    }
}

func TestFolderTreeABIFields(t *testing.T) {
    uuidType := reflect.TypeOf(uuid.UUID{})
    roleType := reflect.TypeOf(mail.FolderTreeRole(""))
    refType := reflect.TypeOf(mail.FolderReferenceState(""))
    fields(t, mail.FolderTreeFolder{}, []string{"AccountID", "ID", "ParentID", "Role", "Name", "SortOrder"}, []reflect.Type{uuidType, uuidType, uuidType, roleType, reflect.TypeOf(""), reflect.TypeOf(int32(0))})
    fields(t, mail.FolderTreeSnapshot{}, []string{"AccountID", "Folders"}, []reflect.Type{uuidType, reflect.TypeOf([]mail.FolderTreeFolder{})})
    fields(t, mail.FolderTreeNode{}, []string{"AccountID", "ID", "ParentID", "Role", "Name", "NormalizedName", "SortOrder"}, []reflect.Type{uuidType, uuidType, uuidType, roleType, reflect.TypeOf(""), reflect.TypeOf(""), reflect.TypeOf(int32(0))})
    fields(t, mail.FolderTreeCreate{}, []string{"ID", "ParentID", "Name", "SortOrder"}, []reflect.Type{uuidType, uuidType, reflect.TypeOf(""), reflect.TypeOf(int32(0))})
    fields(t, mail.FolderTreeRename{}, []string{"FolderID", "Name"}, []reflect.Type{uuidType, reflect.TypeOf("")})
    fields(t, mail.FolderTreeReparent{}, []string{"FolderID", "ParentID"}, []reflect.Type{uuidType, uuidType})
    fields(t, mail.FolderTreeDelete{}, []string{"FolderID", "ActiveMoveRuleReference"}, []reflect.Type{uuidType, refType})
    fields(t, mail.FolderTreeChange{}, []string{"Create", "Rename", "Reparent", "Delete"}, []reflect.Type{reflect.TypeOf((*mail.FolderTreeCreate)(nil)), reflect.TypeOf((*mail.FolderTreeRename)(nil)), reflect.TypeOf((*mail.FolderTreeReparent)(nil)), reflect.TypeOf((*mail.FolderTreeDelete)(nil))})
    fields(t, mail.FolderTreeDecision{}, []string{"NormalizedName"}, []reflect.Type{reflect.TypeOf("")})
    fields(t, mail.FolderPolicyError{}, []string{"Kind"}, []reflect.Type{reflect.TypeOf(mail.FolderErrorKind(""))})
}
```

Expected static RED state: all references demand the exact exported ABI and the source path is still absent before Task 2. No executable command is authorized.

- [ ] **Step 2: Encode direct literal, sentinel, nil, `errors.Is` and `errors.As` assertions.**

| Kind | Literal | Sentinel text |
|---|---|---|
| invalid | `invalid` | `mail folder invalid` |
| too_large | `too_large` | `mail folder exceeds bounds` |
| duplicate | `duplicate` | `mail folder duplicate` |
| name_conflict | `name_conflict` | `mail folder name conflict` |
| account_mismatch | `account_mismatch` | `mail folder account mismatch` |
| not_found | `not_found` | `mail folder not found` |
| parent_unavailable | `parent_unavailable` | `mail folder parent unavailable` |
| self_parent | `self_parent` | `mail folder self parent` |
| cycle | `cycle` | `mail folder cycle` |
| depth_exceeded | `depth_exceeded` | `mail folder depth exceeds bounds` |
| system_role | `system_role` | `mail folder system role protected` |
| in_use | `in_use` | `mail folder in use` |
| limit_exceeded | `limit_exceeded` | `mail folder limit exceeded` |

```go
func TestFolderPolicyErrors(t *testing.T) {
    cases := []struct {
        kind mail.FolderErrorKind
        literal string
        text string
        sentinel error
    }{
        {mail.FolderErrorInvalid, "invalid", "mail folder invalid", mail.ErrFolderInvalid},
        {mail.FolderErrorTooLarge, "too_large", "mail folder exceeds bounds", mail.ErrFolderTooLarge},
        {mail.FolderErrorDuplicate, "duplicate", "mail folder duplicate", mail.ErrFolderDuplicate},
        {mail.FolderErrorNameConflict, "name_conflict", "mail folder name conflict", mail.ErrFolderNameConflict},
        {mail.FolderErrorAccountMismatch, "account_mismatch", "mail folder account mismatch", mail.ErrFolderAccountMismatch},
        {mail.FolderErrorNotFound, "not_found", "mail folder not found", mail.ErrFolderNotFound},
        {mail.FolderErrorParentUnavailable, "parent_unavailable", "mail folder parent unavailable", mail.ErrFolderParentUnavailable},
        {mail.FolderErrorSelfParent, "self_parent", "mail folder self parent", mail.ErrFolderSelfParent},
        {mail.FolderErrorCycle, "cycle", "mail folder cycle", mail.ErrFolderCycle},
        {mail.FolderErrorDepthExceeded, "depth_exceeded", "mail folder depth exceeds bounds", mail.ErrFolderDepthExceeded},
        {mail.FolderErrorSystemRole, "system_role", "mail folder system role protected", mail.ErrFolderSystemRole},
        {mail.FolderErrorInUse, "in_use", "mail folder in use", mail.ErrFolderInUse},
        {mail.FolderErrorLimitExceeded, "limit_exceeded", "mail folder limit exceeded", mail.ErrFolderLimitExceeded},
    }
    require.Equal(t, "absent", string(mail.FolderReferenceAbsent))
    require.Equal(t, "present", string(mail.FolderReferencePresent))
    require.Equal(t, "", string(mail.FolderTreeRole("")))
    require.Equal(t, "inbox", string(mail.FolderTreeRoleInbox))
    require.Equal(t, "sent", string(mail.FolderTreeRoleSent))
    require.Equal(t, "drafts", string(mail.FolderTreeRoleDrafts))
    require.Equal(t, "archive", string(mail.FolderTreeRoleArchive))
    require.Equal(t, "trash", string(mail.FolderTreeRoleTrash))
    require.Equal(t, "junk", string(mail.FolderTreeRoleJunk))
    for _, tc := range cases {
        require.Equal(t, tc.literal, string(tc.kind))
        err := &mail.FolderPolicyError{Kind: tc.kind}
        require.Equal(t, tc.text, err.Error())
        require.True(t, errors.Is(err, tc.sentinel))
        var typed *mail.FolderPolicyError
        require.True(t, errors.As(err, &typed))
        require.Equal(t, tc.kind, typed.Kind)
        for _, other := range cases {
            if other.sentinel != tc.sentinel {
                require.False(t, errors.Is(err, other.sentinel))
            }
        }
    }
    var nilPolicy *mail.FolderPolicyError
    require.Equal(t, "mail folder policy error", nilPolicy.Error())
    require.Nil(t, nilPolicy.Unwrap())
    unknown := &mail.FolderPolicyError{Kind: mail.FolderErrorKind("unknown")}
    require.Equal(t, "mail folder policy error", unknown.Error())
    require.Nil(t, unknown.Unwrap())
}
```

- [ ] **Step 3: Encode fixed snapshot, normalizer, capacity, graph, collision, ordering and four-overlay vectors.**

| Vector | Input | Expected |
|---|---|---|
| valid empty snapshot | nonzero fixture account; empty Folders | concrete nonnil `[]FolderTreeNode` length 0; nil error |
| zero snapshot account | Snapshot AccountID `uuid.Nil` | invalid |
| zero row account | nonzero snapshot account; row AccountID `uuid.Nil` | invalid |
| foreign row account | row AccountID `00000000-0000-4000-8000-000000000002` | account_mismatch |
| zero row ID | row ID `uuid.Nil` | invalid |
| root | nonzero ID and ParentID `uuid.Nil` | valid root depth 0 |
| 101 rows | 101 fixed IDs | too_large |
| invalid names | `"bad\x00name"`; `"bad\u0085name"`; `string([]byte{0xff})` | invalid each |
| raw rune bound | `strings.Repeat("a", 121)` | too_large |
| post-fold key bound | `strings.Repeat("\u1F88", 120)`; raw 120 runes/360 UTF-8 bytes and independently expected fold key 600 bytes | too_large |
| duplicate ID | two fixed rows with same nonzero ID | duplicate |
| duplicate system role | distinct fixed IDs, both Inbox | duplicate |
| missing parent | custom row ParentID `00000000-0000-4000-8000-000000000003`, absent from snapshot | parent_unavailable |
| whitespace key | `\u00a0\t` | invalid |
| NFC | `A\u030a` | `å` |
| collapse/trim | `\tCafe\u0301\u00a0` | `café` |
| fold expansion | `Straße` | `strasse` |
| sigma | `Σ` | `σ` |
| dotted-I expansion | `İ` | `i\u0307` |
| self/cycle/depth | self; two-node cycle; root→1→2→3 | self_parent; cycle; depth_exceeded |
| custom collision | same ParentID and `\u00a0Straße`/`STRASSE\t` | name_conflict |
| system/custom coexistence | Inbox `Inbox` and custom `inbox` | valid |
| ordering | role roots plus custom SortOrder/key/UUID ties | fixed rank then byte-order then raw UUID bytes |
| aliases | mutate source after call; mutate returned slice | later output/input unchanged |
| Create at 100 | `uuid.Nil`, valid name | invalid |
| Create at 100 | `00000000-0000-4000-8000-000000000201`, 480-byte valid raw name | limit_exceeded |
| Create at 100 | same ID, 481-byte raw name | too_large |
| Create at 100 | same ID, `\xff` | limit_exceeded |
| Create below capacity duplicate ID | valid base snapshot contains supplied fixed Create.ID | duplicate; zero decision |
| missing action target | nonzero `00000000-0000-4000-8000-000000000004` Rename/Delete FolderID absent from snapshot | not_found; zero decision |
| Delete at 100 | delete custom then create on unchanged snapshot | limit_exceeded |
| system action precedence | system rename `\xff`; system reparent missing parent; system delete unknown ref | system_role each |
| Delete rules | direct child; present; absent; zero reference | in_use; in_use; success empty decision; invalid |

```go
func id(n int) uuid.UUID { return fixedFolderIDs[n] }

func snapshot100() mail.FolderTreeSnapshot {
    account := fixtureAccountID
    rows := make([]mail.FolderTreeFolder, 0, 100)
    roles := []mail.FolderTreeRole{mail.FolderTreeRoleInbox, mail.FolderTreeRoleSent, mail.FolderTreeRoleDrafts, mail.FolderTreeRoleArchive, mail.FolderTreeRoleTrash, mail.FolderTreeRoleJunk}
    for i := 0; i < 100; i++ {
        role := mail.FolderTreeRole("")
        name := "custom-" + string(rune('a'+(i%26))) + strings.Repeat("x", i/26)
        if i < len(roles) {
            role = roles[i]
            name = []string{"Inbox", "Sent", "Drafts", "Archive", "Trash", "Junk"}[i]
        }
        rows = append(rows, mail.FolderTreeFolder{AccountID: account, ID: id(i), Role: role, Name: name, SortOrder: int32(i)})
    }
    return mail.FolderTreeSnapshot{AccountID: account, Folders: rows}
}

func folderRow(folderID, parentID uuid.UUID, role mail.FolderTreeRole, name string) mail.FolderTreeFolder {
    return mail.FolderTreeFolder{
        AccountID: fixtureAccountID,
        ID:        folderID,
        ParentID:  parentID,
        Role:      role,
        Name:      name,
    }
}

func snapshot(rows ...mail.FolderTreeFolder) mail.FolderTreeSnapshot {
    return mail.FolderTreeSnapshot{AccountID: fixtureAccountID, Folders: rows}
}

func requireNormalizeFailure(t *testing.T, input mail.FolderTreeSnapshot, sentinel error, text string) {
    t.Helper()
    nodes, err := mail.NormalizeFolderTree(input)
    require.Nil(t, nodes)
    require.ErrorIs(t, err, sentinel)
    require.Equal(t, text, err.Error())
}

func requireChangeFailure(t *testing.T, input mail.FolderTreeSnapshot, change mail.FolderTreeChange, sentinel error, text string) {
    t.Helper()
    decision, err := mail.EvaluateFolderTreeChange(input, change)
    require.Equal(t, mail.FolderTreeDecision{}, decision)
    require.ErrorIs(t, err, sentinel)
    require.Equal(t, text, err.Error())
}

func TestNormalizeFolderTreeNamedBoundaryRows(t *testing.T) {
    postFoldRaw := strings.Repeat("\u1F88", 120)
    // This independent expected key is the literal full-fold expansion U+1F00+U+03B9, not a production helper.
    postFoldExpected := strings.Repeat("\u1F00\u03B9", 120)
    require.Equal(t, 120, utf8.RuneCountInString(postFoldRaw))
    require.Equal(t, 360, len(postFoldRaw))
    require.Equal(t, 600, len(postFoldExpected))

    cases := []struct {
        name     string
        input    mail.FolderTreeSnapshot
        sentinel error
        text     string
        emptyOK  bool
    }{
        {name: "valid-empty-snapshot", input: mail.FolderTreeSnapshot{AccountID: fixtureAccountID, Folders: []mail.FolderTreeFolder{}}, emptyOK: true},
        {name: "nul-name", input: snapshot(folderRow(id(0), uuid.Nil, "", "bad\x00name")), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
        {name: "c1-control-name", input: snapshot(folderRow(id(0), uuid.Nil, "", "bad\u0085name")), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
        {name: "invalid-utf8-name", input: snapshot(folderRow(id(0), uuid.Nil, "", string([]byte{0xff}))), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
        {name: "121-rune-name", input: snapshot(folderRow(id(0), uuid.Nil, "", strings.Repeat("a", 121))), sentinel: mail.ErrFolderTooLarge, text: "mail folder exceeds bounds"},
        {name: "post-fold-key-overflow", input: snapshot(folderRow(id(0), uuid.Nil, "", postFoldRaw)), sentinel: mail.ErrFolderTooLarge, text: "mail folder exceeds bounds"},
        {name: "duplicate-snapshot-id", input: snapshot(folderRow(id(0), uuid.Nil, "", "first"), folderRow(id(0), uuid.Nil, "", "second")), sentinel: mail.ErrFolderDuplicate, text: "mail folder duplicate"},
        {name: "duplicate-inbox-role", input: snapshot(folderRow(id(0), uuid.Nil, mail.FolderTreeRoleInbox, "Inbox"), folderRow(id(1), uuid.Nil, mail.FolderTreeRoleInbox, "Inbox second")), sentinel: mail.ErrFolderDuplicate, text: "mail folder duplicate"},
        {name: "missing-parent", input: snapshot(folderRow(id(0), fixtureMissingParentID, "", "child")), sentinel: mail.ErrFolderParentUnavailable, text: "mail folder parent unavailable"},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            if tc.emptyOK {
                nodes, err := mail.NormalizeFolderTree(tc.input)
                require.NoError(t, err)
                require.NotNil(t, nodes)
                require.Len(t, nodes, 0)
                return
            }
            requireNormalizeFailure(t, tc.input, tc.sentinel, tc.text)
        })
    }
}

func TestFolderTreeNamedMutationFailureRows(t *testing.T) {
    base := snapshot(folderRow(id(0), uuid.Nil, "", "existing"))
    cases := []struct {
        name     string
        input    mail.FolderTreeSnapshot
        change   mail.FolderTreeChange
        sentinel error
        text     string
    }{
        {name: "rename-missing-target", input: snapshot(), change: mail.FolderTreeChange{Rename: &mail.FolderTreeRename{FolderID: fixtureMissingTargetID, Name: "renamed"}}, sentinel: mail.ErrFolderNotFound, text: "mail folder not found"},
        {name: "delete-missing-target", input: snapshot(), change: mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: fixtureMissingTargetID, ActiveMoveRuleReference: mail.FolderReferenceAbsent}}, sentinel: mail.ErrFolderNotFound, text: "mail folder not found"},
        {name: "create-duplicate-id-below-capacity", input: base, change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(0), Name: "new custom"}}, sentinel: mail.ErrFolderDuplicate, text: "mail folder duplicate"},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            requireChangeFailure(t, tc.input, tc.change, tc.sentinel, tc.text)
        })
    }
}

func TestCreateCapacityPrecedence(t *testing.T) {
    full := snapshot100()
    cases := []struct {
        name     string
        change   mail.FolderTreeChange
        sentinel error
        text     string
    }{
        {name: "zero-id-before-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: uuid.Nil, Name: "valid"}}, sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
        {name: "valid-raw-name-at-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(100), Name: strings.Repeat("x", 480)}}, sentinel: mail.ErrFolderLimitExceeded, text: "mail folder limit exceeded"},
        {name: "raw-byte-overflow-before-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(100), Name: strings.Repeat("x", 481)}}, sentinel: mail.ErrFolderTooLarge, text: "mail folder exceeds bounds"},
        {name: "malformed-within-raw-bound-at-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(100), Name: "\xff"}}, sentinel: mail.ErrFolderLimitExceeded, text: "mail folder limit exceeded"},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            requireChangeFailure(t, full, tc.change, tc.sentinel, tc.text)
        })
    }
}
```

For each table row, copy the input snapshot before the callable and compare it byte-for-byte after the call. Every Normalize error uses `requireNormalizeFailure` and therefore has a nil result slice, exact text and `errors.Is` sentinel; every Evaluate error uses `requireChangeFailure` and therefore has a zero decision, exact text and `errors.Is` sentinel. Only `valid-empty-snapshot` asserts a nonnil empty result. The union matrix includes zero arms, two arms, unknown reference state, zero rename/reparent/delete targets and zero Create ID. Rename asserts same-name success and a returned key; Reparent asserts empty decision and subtree graph revalidation; Delete asserts empty decision and unchanged snapshot.

- [ ] **Step 4: Encode AST public-surface and purity oracle against only `folder_tree.go`.**

```go
func sourcePath(t *testing.T) string {
    t.Helper()
    _, testFile, _, ok := runtime.Caller(0)
    require.True(t, ok)
    return filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", "folder_tree.go"))
}

func TestFolderTreeSourceASTContract(t *testing.T) {
    path := sourcePath(t)
    data, err := os.ReadFile(path)
    require.NoError(t, err)
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
    require.NoError(t, err)

    allowedImports := map[string]bool{
        "bytes": true, "errors": true, "sort": true, "strings": true,
        "unicode": true, "unicode/utf8": true, "github.com/google/uuid": true,
        "golang.org/x/text/cases": true, "golang.org/x/text/unicode/norm": true,
    }
    for _, imp := range file.Imports {
        require.True(t, allowedImports[strings.Trim(imp.Path.Value, "\"")])
    }

    allowedTypes := map[string]bool{"FolderTreeRole": true, "FolderTreeFolder": true, "FolderTreeSnapshot": true, "FolderTreeNode": true, "FolderTreeCreate": true, "FolderTreeRename": true, "FolderTreeReparent": true, "FolderReferenceState": true, "FolderTreeDelete": true, "FolderTreeChange": true, "FolderTreeDecision": true, "FolderErrorKind": true, "FolderPolicyError": true}
    allowedFunctions := map[string]bool{"NormalizeFolderTree": true, "EvaluateFolderTreeChange": true}
    allowedMethods := map[string]bool{"Error": true, "Unwrap": true}
    allowedConstants := map[string]bool{"FolderTreeRoleInbox": true, "FolderTreeRoleSent": true, "FolderTreeRoleDrafts": true, "FolderTreeRoleArchive": true, "FolderTreeRoleTrash": true, "FolderTreeRoleJunk": true, "FolderReferenceAbsent": true, "FolderReferencePresent": true, "FolderErrorInvalid": true, "FolderErrorTooLarge": true, "FolderErrorDuplicate": true, "FolderErrorNameConflict": true, "FolderErrorAccountMismatch": true, "FolderErrorNotFound": true, "FolderErrorParentUnavailable": true, "FolderErrorSelfParent": true, "FolderErrorCycle": true, "FolderErrorDepthExceeded": true, "FolderErrorSystemRole": true, "FolderErrorInUse": true, "FolderErrorLimitExceeded": true}
    allowedSentinels := map[string]string{"ErrFolderInvalid": "mail folder invalid", "ErrFolderTooLarge": "mail folder exceeds bounds", "ErrFolderDuplicate": "mail folder duplicate", "ErrFolderNameConflict": "mail folder name conflict", "ErrFolderAccountMismatch": "mail folder account mismatch", "ErrFolderNotFound": "mail folder not found", "ErrFolderParentUnavailable": "mail folder parent unavailable", "ErrFolderSelfParent": "mail folder self parent", "ErrFolderCycle": "mail folder cycle", "ErrFolderDepthExceeded": "mail folder depth exceeds bounds", "ErrFolderSystemRole": "mail folder system role protected", "ErrFolderInUse": "mail folder in use", "ErrFolderLimitExceeded": "mail folder limit exceeded"}
    verifyPublicDeclarations(t, file, allowedTypes, allowedFunctions, allowedMethods, allowedConstants, allowedSentinels)
    verifyPurityAST(t, file, allowedSentinels)
}
```

`verifyPublicDeclarations` walks every top-level declaration and records each encountered exported name: exported `type` names must equal `allowedTypes`; exported functions without receivers must equal `allowedFunctions`; exported methods must have receiver `*FolderPolicyError` and, as a set, equal `Error` and `Unwrap`; exported constants must equal `allowedConstants`; exported vars must equal the 13 sentinel names. It fails if `FolderTreeRoleCustom` or any other exported declaration appears. It checks every sentinel value is an `errors.New` call with its mapped literal and rejects any non-sentinel package variable.

`verifyPurityAST` uses `ast.Inspect`: an `*ast.GoStmt` fails; a top-level `init` fails; any identifier named `context`, `MailEngine`, `Repository`, `repository` or `callback` fails; every `*ast.Field` whose type is `*ast.FuncType` fails; every function-typed struct field fails; any package variable beyond allowed sentinels fails. A human reviewer additionally reads the file for indirect I/O, global mutation, logging, clock, random and callback behavior that AST spelling checks cannot prove.

- [ ] **Step 5: Collect static RED evidence and route the three reviews; do not run or commit.**

```powershell
git status --short
rg -n 'FolderTree(Role|Folder|Snapshot|Node|Create|Rename|Reparent|Delete|Change|Decision|ErrorKind|PolicyError)|ErrFolder|NormalizeFolderTree|EvaluateFolderTreeChange' backend/tests/mail/folder_tree_test.go backend/internal/mail
git diff --check -- backend/tests/mail/folder_tree_test.go
git hash-object backend/tests/mail/folder_tree_test.go
git diff -- backend/tests/mail/folder_tree_test.go
```

Expected: Task 1 source is a static RED oracle; no executable invocation is authorized, so compile and runtime are `NOT_RUN`. Route these exact review packets: Spec/Oracle reviewer `gpt-5.6-terra/high`, Security reviewer `gpt-5.6-terra/xhigh`, ABI/Compile reviewer `gpt-5.6-terra/high`. Each native dispatch explicitly says no fallback, `dispatcher-enforced` and telemetry `UNKNOWN` unless exposed. No commit occurs.

## Task 2: Production Folder Tree Policy only

**Files:**
- Create: `backend/internal/mail/folder_tree.go`
- Do not modify: `backend/tests/mail/folder_tree_test.go` or documentation in this task.

**Author route:** fresh `gpt-5.6-terra` at `xhigh`, explicit no fallback, `dispatcher-enforced`; actual telemetry `UNKNOWN` unless exposed.

**Consumes:** Task 1 reviewed exact oracle blob and the ABI/precedence above.

**Produces:** one self-contained pure source file with the full public ABI and no additional public declarations.

- [ ] **Step 1: Declare the frozen ABI, fixed sentinels and total error mapping.**

Write the types/constants/13 `errors.New` sentinels exactly as the Frozen public ABI section. Implement `Error` and `Unwrap` as these total mappings:

| Kind | Error text | Unwrap |
|---|---|---|
| invalid | `mail folder invalid` | ErrFolderInvalid |
| too_large | `mail folder exceeds bounds` | ErrFolderTooLarge |
| duplicate | `mail folder duplicate` | ErrFolderDuplicate |
| name_conflict | `mail folder name conflict` | ErrFolderNameConflict |
| account_mismatch | `mail folder account mismatch` | ErrFolderAccountMismatch |
| not_found | `mail folder not found` | ErrFolderNotFound |
| parent_unavailable | `mail folder parent unavailable` | ErrFolderParentUnavailable |
| self_parent | `mail folder self parent` | ErrFolderSelfParent |
| cycle | `mail folder cycle` | ErrFolderCycle |
| depth_exceeded | `mail folder depth exceeds bounds` | ErrFolderDepthExceeded |
| system_role | `mail folder system role protected` | ErrFolderSystemRole |
| in_use | `mail folder in use` | ErrFolderInUse |
| limit_exceeded | `mail folder limit exceeded` | ErrFolderLimitExceeded |
| nil or unknown | `mail folder policy error` | nil |

The only package variables are those sentinel declarations; all indexes and fold state are per-call local allocations.

- [ ] **Step 2: Implement raw-name validation and private normalized key.**

```go
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
```

This function returns only the private key. Callers preserve `FolderTreeFolder.Name` exact input bytes/runes in `FolderTreeNode.Name`.

- [ ] **Step 3: Implement complete snapshot index, graph validation and deterministic Normalize.**

Use per-call internal records:

```go
type indexedFolder struct {
    folder FolderTreeFolder
    key    string
}
type snapshotIndex struct {
    rows       map[uuid.UUID]indexedFolder
    children   map[uuid.UUID][]uuid.UUID
    systemRole map[FolderTreeRole]uuid.UUID
}
```

`validateSnapshot` follows this exact sequence:

1. Reject zero Snapshot.AccountID as invalid; reject len(Folders)>100 as too_large.
2. For each row: reject zero row AccountID or zero row ID as invalid; reject nonzero unequal AccountID as account_mismatch; normalize raw Name; reject unsupported nonzero Role as invalid; reject system Role with nonzero ParentID as invalid.
3. Insert row ID once; duplicate ID is duplicate. Insert nonzero Role once; duplicate system role is duplicate.
4. For every custom row with nonzero ParentID, reject ParentID==ID as self_parent; reject absent parent as parent_unavailable; reject present parent whose AccountID differs as account_mismatch; append child ID to a local children map.
5. DFS each root/row with white-gray-black state. A gray revisit is cycle. After every child result, fail if root depth plus child edges reaches 3; root is depth 0 and maximum accepted depth is 2.
6. For every custom row, form collision key ParentID plus private key. A repeated pair is name_conflict. System/custom key equality is ignored.
7. Sort each children slice with `SortOrder` ascending, `bytes.Compare([]byte(keyA), []byte(keyB))`, then `bytes.Compare(idA[:], idB[:])`. Build roots as fixed system rank followed by sorted custom roots. Append copied scalar fields depth-first to `make([]FolderTreeNode, 0, len(Folders))`. Return that allocation even when len is zero.

```go
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
```

`appendPreorder` copies AccountID, ID, ParentID, Role, Name, key as NormalizedName and SortOrder into a new node, then visits each already-sorted child ID. It reads neither caller backing storage nor package state.

- [ ] **Step 4: Implement exact union/overlay Evaluate behavior.**

```go
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
```

`evaluateCreate` rejects zero ID, then raw `len(Name)>480`, then `len(index.rows)==100`, then invokes full name normalizer; it rejects duplicate supplied ID, validates optional parent, applies prospective depth and sibling custom collision, and returns `FolderTreeDecision{NormalizedName:key}` without mutating index. `evaluateRename` normalizes after custom target gate, excludes target UUID from collision comparison and returns the key. `evaluateReparent` validates optional parent, copies only target ParentID into a fresh bounded overlay, rebuilds local children and runs cycle before depth across the target's entire existing subtree; it returns `FolderTreeDecision{}`. `evaluateDelete` checks direct children first, then reference: absent returns `FolderTreeDecision{}`, present returns in_use, zero/unknown returns invalid; it never removes or changes any row.

- [ ] **Step 5: Collect source static evidence and route the three production reviews; do not run or commit.**

```powershell
git status --short
git diff --check -- backend/internal/mail/folder_tree.go
rg -n '^import|context|MailEngine|Repository|repository|callback|go func|func init|time\.|rand\.|slog\.|log\.' backend/internal/mail/folder_tree.go
rg -n 'type FolderTree(Role|Folder|Snapshot|Node|Create|Rename|Reparent|Delete|Change|Decision|ErrorKind|PolicyError)|ErrFolder|func \(e \*FolderPolicyError\) (Error|Unwrap)|func (NormalizeFolderTree|EvaluateFolderTreeChange)' backend/internal/mail/folder_tree.go
git hash-object backend/internal/mail/folder_tree.go
git diff -- backend/internal/mail/folder_tree.go
```

Expected: diff check is clean; import and exported-declaration scans match the ABI; reviewer AST inspection enforces purity; compile/runtime remain `NOT_RUN`. Route Spec/Quality `gpt-5.6-terra/high`, Security `gpt-5.6-terra/xhigh`, and Oracle `gpt-5.6-terra/high`; every dispatch explicitly records no fallback, `dispatcher-enforced` and telemetry `UNKNOWN` unless exposed. No commit occurs.

## Task 3: Root exact two-file atomic source/oracle commit and exact-SHA reviews

**Read-only evidence allowance for Tasks 3 and 4:** `git status --short`, `git diff --check`, `git diff`, `git diff-tree`, `git hash-object`, `git show`, `git rev-parse` and scoped `rg` only. These commands inspect Git/file state; none executes a project target.

**Files:**
- Stage and commit exactly: `backend/tests/mail/folder_tree_test.go`
- Stage and commit exactly: `backend/internal/mail/folder_tree.go`
- Do not stage: any lifecycle, evidence, contract, plan, generated, schema or handler file.

**Root-only route:** root stages and commits; no worker runs a commit. The three reviews route to `gpt-5.6-terra/high`, `gpt-5.6-terra/xhigh` and `gpt-5.6-terra/high`, respectively; every native dispatch states explicit model/effort/no fallback/`dispatcher-enforced`/telemetry `UNKNOWN`.

- [ ] **Step 1: Verify exact reviewed two-file boundary before stage.**

```powershell
git status --short
git diff --check -- backend/tests/mail/folder_tree_test.go backend/internal/mail/folder_tree.go
git diff --name-only -- backend/tests/mail/folder_tree_test.go backend/internal/mail/folder_tree.go
git hash-object backend/tests/mail/folder_tree_test.go
git hash-object backend/internal/mail/folder_tree.go
```

Expected: both reviewed blobs are present, no third production or documentation path is selected, and execution state remains `NOT_RUN`.

- [ ] **Step 2: Root stages exactly two literal paths and creates one atomic commit.**

```powershell
git add -- backend/tests/mail/folder_tree_test.go backend/internal/mail/folder_tree.go
git diff --cached --name-only
git commit -m "feat(mail): add folder tree policy"
git rev-parse HEAD
```

Expected: cached name-only output has exactly those two paths. The commit has no AI trailer. Root records the resulting full SHA before any review.

- [ ] **Step 3: Complete three independent reviews of that exact commit SHA.**

```powershell
git show --stat --oneline HEAD
git diff HEAD^ HEAD -- backend/tests/mail/folder_tree_test.go backend/internal/mail/folder_tree.go
git rev-parse HEAD
```

Expected: Spec/Oracle review `gpt-5.6-terra/high`, Security review `gpt-5.6-terra/xhigh`, and ABI/Compile review `gpt-5.6-terra/high` each name the same SHA and check the two-file boundary. Any non-PASS review leaves lifecycle unelevated and blocks Task 4 commit.

## Task 4: Separate docs-only lifecycle and evidence sync

**Files:**
- Create: `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03B-FOLDER.md`
- Modify: only active lifecycle/docs paths discovered by Step 1.
- Do not modify: `backend/tests/mail/folder_tree_test.go` or `backend/internal/mail/folder_tree.go` in this task.

**Author route:** fresh `gpt-5.6-terra` at `medium`, explicit no fallback, `dispatcher-enforced`; actual telemetry `UNKNOWN` unless exposed.

**Produces:** a docs-only commit in the same session after Task 3's three exact-SHA reviews. Those passed reviews make the exact two-file source/oracle boundary `STATIC_ACCEPTED` static-only before Task 4 begins. Task 4 records that accepted boundary, while its new docs-only commit is a candidate pending its own exact-SHA reviews. Neither state creates a runtime PASS or unblocks 04B engine reparent, 144-01B, Phase 142 or Phase 148.

- [ ] **Step 1: Discover the exact active docs inventory and source/oracle blobs.**

```powershell
git rev-parse HEAD
git show --stat --oneline HEAD
git hash-object backend/tests/mail/folder_tree_test.go
git hash-object backend/internal/mail/folder_tree.go
rg -l '144-03B|FOLDER_TREE_CALLABLE_ERROR_NORMALIZATION_ORACLE_STATIC_ACCEPTED|STATIC_ACCEPTED' .planning/STATE.md .planning/decisions/specs/MAIL-SPEC.md .planning/phases/143-calendar-rsvp-core/.continue-here.md .planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md .planning/phases/143-calendar-rsvp-core/TERMINAL-RESUME-PROMPT-RU.md .planning/phases/144-mailbox-domain/DISPATCH-INDEX.md .planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md .planning/phases/144-mailbox-domain/PLAN.md .planning/phases/148-mail-release-gate/GATE-MATRIX.md docs/mail/CONTRACT-CHANGELOG.md docs/mail/ENGINE-CONTRACT-0.1.8.md docs/mail/PERSISTENCE-CONTRACT-0.1.8.md docs/mail/QA-HANDOFF.md docs/mail/openapi.yaml .planning/phases/144-mailbox-domain/CONTRACT-144-03B-FOLDER.md .planning/phases/144-mailbox-domain/PLAN-144-03B-FOLDER.md
git status --short
```

Expected: root selects only current lifecycle references returned by this scoped inventory. Historical-only docs and all nonmatching files remain outside this commit.

The discovery command is historical inventory evidence only. The authoritative Task 4 manifest is
the following literal exact 20-path list, in the accepted contract order and spelling:

1. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03B-FOLDER.md` (new evidence file)
2. `.planning/STATE.md`
3. `.planning/decisions/specs/MAIL-SPEC.md`
4. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`
5. `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md`
6. `.planning/phases/143-calendar-rsvp-core/TERMINAL-RESUME-PROMPT-RU.md`
7. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
8. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
9. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-01A.md`
10. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-ADDRESS.md`
11. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-THREAD.md`
12. `.planning/phases/144-mailbox-domain/PLAN.md`
13. `.planning/phases/148-mail-release-gate/GATE-MATRIX.md`
14. `docs/mail/CONTRACT-CHANGELOG.md`
15. `docs/mail/ENGINE-CONTRACT-0.1.8.md`
16. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
17. `docs/mail/QA-HANDOFF.md`
18. `docs/mail/openapi.yaml`
19. `.planning/phases/144-mailbox-domain/CONTRACT-144-03B-FOLDER.md`
20. `.planning/phases/144-mailbox-domain/PLAN-144-03B-FOLDER.md`

- [ ] **Step 2: Write evidence with exact static facts and preserved blockers.**

```markdown
# 144-03B Folder implementation evidence

- Contract base: a565af2632bc5f76a7d2fe26ebf3065fb3cefba9
- Source: backend/internal/mail/folder_tree.go; exact blob copied from Step 1
- Oracle: backend/tests/mail/folder_tree_test.go; exact blob copied from Step 1
- Source/oracle commit: exact Task 3 SHA
- Source/oracle boundary: STATIC_ACCEPTED, static-only, after three Task 3 exact-SHA PASS reviews
- This docs-only sync commit: candidate pending its own Lifecycle/Evidence, Security/Dependency and Exact/Boundary exact-SHA reviews
- Reviews: exact Task 3 Spec/Oracle, Security and ABI/Compile PASS outcomes
- Runner: NOT_READY
- Compile, unit, race, build, lint, Docker and runtime: NOT_RUN
- Preserved blockers: 04B engine-port reparent, 144-01B, Phase 142 and Phase 148
```

The selected lifecycle docs repeat only that evidence. They do not state product acceptance, schema/migration/handler completion, executable pass, or engine reparent authorization.

- [ ] **Step 3: Root stages only docs/evidence literal paths and makes a separate docs-only commit.**

```powershell
git diff --check -- .planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03B-FOLDER.md
git add -- .planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03B-FOLDER.md
git diff --cached --name-only
```

Root then adds each Step 1 selected lifecycle path through one separate `git add -- exact/literal/path` command, with no glob, wildcard, variable expansion or substitution.

```powershell
git commit -m "docs(mail): record folder tree static evidence"
git rev-parse HEAD
```

Expected: the docs-only cached boundary contains the evidence file and only selected active lifecycle docs; the commit has no AI trailer.

- [ ] **Step 4: Complete separate exact-SHA docs reviews and retain honest runtime state.**

```powershell
git show --stat --oneline HEAD
git diff HEAD^ HEAD
git diff-tree --no-commit-id --name-only -r HEAD
git rev-parse HEAD
```

Route Lifecycle/Evidence `gpt-5.6-terra/high`, Security/Dependency `gpt-5.6-terra/xhigh`, and Exact/Boundary `gpt-5.6-terra/high`. Each reviewer compares the full `git diff-tree` name-only output with the literal exact 20-path manifest in Task 4 Step 1, then inspects the full `git diff HEAD^ HEAD` content; no reviewer filters to the evidence file alone. Every dispatch explicitly uses model `gpt-5.6-terra`, its stated effort, no fallback, `dispatcher-enforced` and telemetry `UNKNOWN` unless exposed. Until those three exact-SHA reviews complete, the docs commit stays a candidate; the already `STATIC_ACCEPTED` source/oracle boundary and `NOT_READY`/`NOT_RUN` runtime truth do not change.

## Self-review checklist

- [x] **Spec coverage:** The prerequisite gate blocks Task 1 until both lifecycle acceptance-sync and plan acceptance. Task 1 contains external ABI/literal/AST/purity/fixture oracle requirements, named valid-empty/error result rows, the independent post-fold overflow key, fixed duplicate/missing rows, and zero decision checks. Task 2 contains every validation, graph, ordering and overlay branch. Task 3 commits exactly two source/oracle paths and requires three same-SHA reviews. Task 4 records the already `STATIC_ACCEPTED` static-only two-file boundary, then separately reviews its still-candidate docs commit.
- [x] **Placeholder scan:** Before plan acceptance, run `rg -n -i '\\x54\\x42\\x44|\\x54\\x4f\\x44\\x4f|\\x73\\x69\\x6d\\x69\\x6c\\x61\\x72|<\\x5b\\x41-\\x5a\\x61-\\x7a\\x5d[^>]*>' .planning/phases/144-mailbox-domain/PLAN-144-03B-FOLDER.md`. Expected: no matches. The plan uses explicit tables, path names, source statements and static command sources instead of unspecified actions.
- [x] **Type consistency:** Task 1 reflection/call assignments, Task 2 declarations and Task 3's two staged paths use identical `FolderTree*` names, `uuid.UUID` fields, error kinds, sentinels and public callables. Task 4 refers to those two exact paths only.
- [x] **Scope:** Plan authoring changes only this file. Task 1 writes only the external test; Task 2 writes only the policy source; Task 3 commits only those two; Task 4 commits docs/evidence separately.
- [x] **Fixed fixtures:** The account, foreign account, root marker and all 101 capacity IDs are fixed `uuid.MustParse` values. No random/time-derived identifier appears.

## Execution handoff

Founder `--auto` preselects `superpowers:subagent-driven-development`; no user choice is requested. Root orchestrates gates, dispatches, reviews and commits. Task 1 author is fresh Terra/xhigh; Task 2 author is fresh Terra/xhigh; Task 3 review routing is Terra/high, Terra/xhigh, Terra/high; Task 4 docs author is Terra/medium and reviews are Terra/high, Terra/xhigh, Terra/high. Every dispatch is explicit `gpt-5.6-terra` with stated effort, no fallback, `dispatcher-enforced` and telemetry `UNKNOWN` unless exposed. No task author stages, commits or executes tests.
