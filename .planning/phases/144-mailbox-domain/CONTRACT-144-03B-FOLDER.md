# 144-03B-Folder — static-accepted callable, error and oracle contract

> **Lifecycle:** `FOLDER_TREE_CALLABLE_ERROR_NORMALIZATION_ORACLE_STATIC_ACCEPTED`.
> **Contract commit:** `a565af2632bc5f76a7d2fe26ebf3065fb3cefba9` · **blob:** `0f9ed0adb0cec6ffa83619cb4c7c4cb4d3f3263c`.
> **Status:** narrowly `STATIC_ACCEPTED` after three independent exact-SHA reviews PASS, C/I/M `0/0/0`; executable evidence is `NOT_RUN`.
> **Author provenance:** native dispatcher-enforced `gpt-5.6-terra` / `medium`, no fallback; actual telemetry `UNKNOWN` unless exposed.

> **Mail lifecycle update:** 02D retains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; 02F correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. Contract, Security and Lifecycle Terra routes PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle is `CONSUMED/HISTORICAL`; Rule/Phase145/runtime blockers remain. Root next accepts separate Rule plan only; Rule TDD awaits Phase-145/G145-5B/G145-5C.

## Scope and next gate

This is an accepted docs-only ABI contract. Its separately accepted source+oracle boundary is
`f8cb7643986fc591cae16764575337b72ef3a692` (source `backend/internal/mail/folder_tree.go` blob
`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `backend/tests/mail/folder_tree_test.go` blob
`efe0db0071cb22ce4b026c635157f0fb08885d64`), narrowly `STATIC_ACCEPTED` / `EXACT_SHA_PASS`,
C/I/M `0/0/0`, static only. It makes no product, schema, handler, migration, repository, engine,
or runtime claim. Thread remains `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS` and is not changed by this acceptance.

The accepted plan is `a98c7235521852dee1ebc0d09558c7bcbefc4534` / blob `c79658d61e496489be697bc971d0797ed30fb61f`, narrowly `EXACT_SHA_PASS` / `STATIC_ACCEPTED` after exact, security and boundary reviews PASS, C/I/M `0/0/0`. Folder source/oracle and docs are accepted static-only at `f8cb7643986fc591cae16764575337b72ef3a692` and `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`. Rule contract is separately docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; its acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. The former 02F transition route is `CONSUMED/HISTORICAL`; correction `fcb5b84f` plus transition `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`. The sole next action is separate Rule-plan author/adjudication; Rule TDD/production also awaits that plan and Phase-145 carrier/G145-5B/G145-5C. The runner is `NOT_READY`; Docker, tests, build, lint and validators remain `NOT_RUN`.

## Frozen Go ABI

The accepted source exposes only the following frozen Folder-tree ABI; there is no custom role literal and no additional public API:

```go
import (
    "errors"

    "github.com/google/uuid"
)

type FolderTreeRole string
const (
    FolderTreeRoleInbox FolderTreeRole = "inbox"
    FolderTreeRoleSent FolderTreeRole = "sent"
    FolderTreeRoleDrafts FolderTreeRole = "drafts"
    FolderTreeRoleArchive FolderTreeRole = "archive"
    FolderTreeRoleTrash FolderTreeRole = "trash"
    FolderTreeRoleJunk FolderTreeRole = "junk"
)
type FolderTreeFolder struct { AccountID uuid.UUID; ID uuid.UUID; ParentID uuid.UUID; Role FolderTreeRole; Name string; SortOrder int32 }
type FolderTreeSnapshot struct { AccountID uuid.UUID; Folders []FolderTreeFolder }
type FolderTreeNode struct { AccountID uuid.UUID; ID uuid.UUID; ParentID uuid.UUID; Role FolderTreeRole; Name string; NormalizedName string; SortOrder int32 }
type FolderTreeCreate struct { ID uuid.UUID; ParentID uuid.UUID; Name string; SortOrder int32 }
type FolderTreeRename struct { FolderID uuid.UUID; Name string }
type FolderTreeReparent struct { FolderID uuid.UUID; ParentID uuid.UUID }
type FolderReferenceState string
const ( FolderReferenceAbsent FolderReferenceState = "absent"; FolderReferencePresent FolderReferenceState = "present" )
type FolderTreeDelete struct { FolderID uuid.UUID; ActiveMoveRuleReference FolderReferenceState }
type FolderTreeChange struct { Create *FolderTreeCreate; Rename *FolderTreeRename; Reparent *FolderTreeReparent; Delete *FolderTreeDelete }
type FolderTreeDecision struct { NormalizedName string }

type FolderErrorKind string
const (
    FolderErrorInvalid FolderErrorKind = "invalid"
    FolderErrorTooLarge FolderErrorKind = "too_large"
    FolderErrorDuplicate FolderErrorKind = "duplicate"
    FolderErrorNameConflict FolderErrorKind = "name_conflict"
    FolderErrorAccountMismatch FolderErrorKind = "account_mismatch"
    FolderErrorNotFound FolderErrorKind = "not_found"
    FolderErrorParentUnavailable FolderErrorKind = "parent_unavailable"
    FolderErrorSelfParent FolderErrorKind = "self_parent"
    FolderErrorCycle FolderErrorKind = "cycle"
    FolderErrorDepthExceeded FolderErrorKind = "depth_exceeded"
    FolderErrorSystemRole FolderErrorKind = "system_role"
    FolderErrorInUse FolderErrorKind = "in_use"
    FolderErrorLimitExceeded FolderErrorKind = "limit_exceeded"
)
var (
    ErrFolderInvalid = errors.New("mail folder invalid")
    ErrFolderTooLarge = errors.New("mail folder exceeds bounds")
    ErrFolderDuplicate = errors.New("mail folder duplicate")
    ErrFolderNameConflict = errors.New("mail folder name conflict")
    ErrFolderAccountMismatch = errors.New("mail folder account mismatch")
    ErrFolderNotFound = errors.New("mail folder not found")
    ErrFolderParentUnavailable = errors.New("mail folder parent unavailable")
    ErrFolderSelfParent = errors.New("mail folder self parent")
    ErrFolderCycle = errors.New("mail folder cycle")
    ErrFolderDepthExceeded = errors.New("mail folder depth exceeds bounds")
    ErrFolderSystemRole = errors.New("mail folder system role protected")
    ErrFolderInUse = errors.New("mail folder in use")
    ErrFolderLimitExceeded = errors.New("mail folder limit exceeded")
)
type FolderPolicyError struct { Kind FolderErrorKind }
func (e *FolderPolicyError) Error() string {
    if e == nil { return "mail folder policy error" }
    switch e.Kind {
    case FolderErrorInvalid: return "mail folder invalid"
    case FolderErrorTooLarge: return "mail folder exceeds bounds"
    case FolderErrorDuplicate: return "mail folder duplicate"
    case FolderErrorNameConflict: return "mail folder name conflict"
    case FolderErrorAccountMismatch: return "mail folder account mismatch"
    case FolderErrorNotFound: return "mail folder not found"
    case FolderErrorParentUnavailable: return "mail folder parent unavailable"
    case FolderErrorSelfParent: return "mail folder self parent"
    case FolderErrorCycle: return "mail folder cycle"
    case FolderErrorDepthExceeded: return "mail folder depth exceeds bounds"
    case FolderErrorSystemRole: return "mail folder system role protected"
    case FolderErrorInUse: return "mail folder in use"
    case FolderErrorLimitExceeded: return "mail folder limit exceeded"
    default: return "mail folder policy error"
    }
}
func (e *FolderPolicyError) Unwrap() error {
    if e == nil { return nil }
    switch e.Kind {
    case FolderErrorInvalid: return ErrFolderInvalid
    case FolderErrorTooLarge: return ErrFolderTooLarge
    case FolderErrorDuplicate: return ErrFolderDuplicate
    case FolderErrorNameConflict: return ErrFolderNameConflict
    case FolderErrorAccountMismatch: return ErrFolderAccountMismatch
    case FolderErrorNotFound: return ErrFolderNotFound
    case FolderErrorParentUnavailable: return ErrFolderParentUnavailable
    case FolderErrorSelfParent: return ErrFolderSelfParent
    case FolderErrorCycle: return ErrFolderCycle
    case FolderErrorDepthExceeded: return ErrFolderDepthExceeded
    case FolderErrorSystemRole: return ErrFolderSystemRole
    case FolderErrorInUse: return ErrFolderInUse
    case FolderErrorLimitExceeded: return ErrFolderLimitExceeded
    default: return nil
    }
}

func NormalizeFolderTree(FolderTreeSnapshot) ([]FolderTreeNode, error)
func EvaluateFolderTreeChange(FolderTreeSnapshot, FolderTreeChange) (FolderTreeDecision, error)
```

Zero `FolderTreeRole` means custom; empty/unknown `FolderReferenceState` is invalid. Zero AccountID, folder ID or change target is invalid; only zero ParentID means root. Exactly one `FolderTreeChange` pointer must be non-nil. The success result of Normalize is a fresh concrete non-nil empty slice for an empty snapshot. No result may alias input/backing storage or mutate input.

## Errors and normalization

`FolderErrorKind` has exactly: `invalid`, `too_large`, `duplicate`, `name_conflict`, `account_mismatch`, `not_found`, `parent_unavailable`, `self_parent`, `cycle`, `depth_exceeded`, `system_role`, `in_use`, `limit_exceeded`; matching sentinels are exactly `ErrFolderInvalid`, `ErrFolderTooLarge`, `ErrFolderDuplicate`, `ErrFolderNameConflict`, `ErrFolderAccountMismatch`, `ErrFolderNotFound`, `ErrFolderParentUnavailable`, `ErrFolderSelfParent`, `ErrFolderCycle`, `ErrFolderDepthExceeded`, `ErrFolderSystemRole`, `ErrFolderInUse`, `ErrFolderLimitExceeded`. `*FolderPolicyError{Kind}` maps to the exact input-free texts respectively: `mail folder invalid`, `mail folder exceeds bounds`, `mail folder duplicate`, `mail folder name conflict`, `mail folder account mismatch`, `mail folder not found`, `mail folder parent unavailable`, `mail folder self parent`, `mail folder cycle`, `mail folder depth exceeds bounds`, `mail folder system role protected`, `mail folder in use`, `mail folder limit exceeded`. Nil/unknown yields `mail folder policy error` with nil unwrap. `errors.Is` is exclusive to its matching sentinel.

Folder IDs are caller-authorized Hub `uuid.UUID` values, never engine IDs, and compare byte-exactly over their 16 raw bytes. Every row AccountID must exactly match Snapshot.AccountID; mismatch is a safe internal error. The projection/persistence layer maps engine opaque IDs to Hub UUID rows before this pure policy: custom role maps to zero, the six system values map identically, unknown/empty engine role degrades, and empty engine parent maps to zero/null.

Raw names are valid UTF-8, contain no Unicode control/NUL, have 1..120 runes and <=480 bytes. Display `Name` retains the exact raw bytes/runes. The private key collapses each `unicode.IsSpace` run to one ASCII space, trims, NFC-normalizes, `cases.Fold`s using `x/text v0.35.0`, then NFC-normalizes again; it must be nonempty and <=480 bytes. It never appears in OpenAPI. Any x/text/Unicode change requires an accepted migration/collision plan.

## Tree policy and deterministic evaluation

A complete snapshot includes every live non-deleted control row, including system and `creating`/unknown/error/deleting; deleted tombstones are excluded by caller. The 0..100 policy count is separate from the observed engine 100-folder bound. Zero role is custom; the six roles are unique roots. Custom folders may be roots or children. Parents must exist exactly; self parent rejects; cycles precede depth; root depth is 0, maximum is 2 and depth 3 rejects. Custom name collision is only same ParentID plus private key. System/custom same name/key is allowed. Create/rename/reparent/delete target custom only; create is always custom.

Normalize validates and returns flat preorder: system roots first in Inbox, Sent, Drafts, Archive, Trash, Junk rank, then custom roots; custom roots and all children sort by SortOrder ascending, normalized-key UTF-8 byte order, then UUID ID's 16 raw-byte order. Traversal is depth-first and must not depend on map iteration. It is pure and bounded O(n log n), n<=100: no I/O, context, MailEngine, repository, callback, HTTP, clock, random, goroutine, logging or global mutable state.

Evaluate validates O(1) union shape, then the complete current snapshot (including its raw count/length before Unicode/maps). Create is custom and has no target: after the valid base snapshot, it first requires Create.ID != uuid.Nil, then preflights only its raw candidate byte bound (a name >480 bytes is too_large), then at current count 100 returns limit_exceeded before candidate Unicode/name semantics, parent, depth or collision. It uses its supplied UUID ID and never generates one. Thus zero ID at 100 is invalid, while nonzero ID at 100 with a malformed but <=480-byte name returns limit_exceeded; remaining name semantics run only below capacity. Rename/reparent/delete first look up their exact nonzero UUID target; a system target returns system_role before candidate name, parent, reference or children semantics. Only after the custom gate do they validate action fields and build the bounded private overlay. Rename excludes its exact target, so same-name rename succeeds. Reparent changes only the private target parent and recomputes cycles/depth for the full existing subtree. Delete never removes a row, frees a slot or changes lifecycle; after the system gate a direct child or `ActiveMoveRuleReference=present` is in use. `absent` means an owner-scoped, account-locked transaction has positively proved no active move rule; zero/unknown fails closed and this value never derives from HTTP. Decision returns the normalized key only for create/rename; reparent/delete return an empty key. Error returns zero decision with no partial mutation.

Precedence is fixed: union shape; complete current snapshot (scope/count/raw bounds, per-row semantic+name, duplicate ID, duplicate system role, self, missing/foreign unavailable, cycles, depth, custom collision); for Create: nonzero ID, raw candidate byte bound, capacity, then remaining name semantics/parent/depth/collision; for rename/reparent/delete: exact nonzero target lookup, system target before action name/parent/reference/children semantics, then custom action-field validation and prospective overlay. Create has no target and is custom. Delete system_role precedes in_use.

## Wire amendment and accepted oracle scope

OpenAPI remains flat ordered items. `MailFolder` requires nullable `parent_id` and `role`; its response name has minLength 1. Wording distinguishes live-slot quota. POST/PATCH/DELETE expose 400 through the existing generic `Error` response and 404 through `NotFound`; POST parent-unavailable is 404, PATCH/DELETE target-unavailable is 404. After nondisclosing target resolution and system-role guard, only a resolved custom PATCH target's supplied `parent_id` is temporarily rejected with 409 TREE_CONFLICT before receipt or engine I/O until port acceptance; a missing target is 404 and a system target is SYSTEM_ROLE. Create and mutation 409 vocab gain `MAIL_FOLDER_TREE_CONFLICT` for self/cycle/depth: existing name conflict remains 409, system/in-use remain existing 409, limit is 422, invalid/too-large are 400, and duplicate/account mismatch projection is degraded/internal. Evaluate may cover reparent, but 04B remains blocked. Color/sort are outside this structural packet.

The accepted external-package oracle uses reflection for exact UUID fields/types, literals and callables and proves no extra public API or custom role. It covers literal errors/sentinels/nil, zero account/ID/target versus zero-root-parent, raw and post-fold names (whitespace-only, NFC, Straße/STRASSE, sigma, fold expansion, empty-or-over-480-byte postfold key), 0/100/101 slots and delete-not-freeing, Create zero ID at 100 => invalid, nonzero ID at 100 with >480 raw bytes => too_large, nonzero ID at 100 with malformed <=480-byte name => limit_exceeded, mixed account/UUIDs, duplicate ID/role, root depth 0..2/depth 3, self, cycles/subtree reparent, custom collisions/system-name coexistence, all system mutations (system before name/parent/reference/children and delete before in_use), invalid union, same-name rename, exact UUID raw-byte ordering, no mutation/alias/partial, and static source purity. Fixtures use only reserved synthetic UUIDs/names.

## Exact 20-path post-source docs candidate manifest

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

Step 1 literal discovery found 16 existing paths. Adding the three existing stale-gate implementation evidence documents yields the frozen 19 existing paths; adding this one new evidence file yields this exact 20-path candidate. This bounded lifecycle inventory does not create a product decision. Facts in this file are current lifecycle/commit/provenance facts. The ABI, policy, wire and oracle paragraphs are narrowly accepted static contract, not product or executable acceptance.

**Current subordinate gate:** Folder docs acceptance is `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
Rule contract is separately docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; its acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. The `55e36b7a` closure and former 02F transition route are `CONSUMED/HISTORICAL`; correction `fcb5b84f` plus transition `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` with all three Terra routes PASS C/I/M `0/0/0`. The sole next action is separate Rule-plan author/adjudication; the accepted Rule plan and Phase-145 carrier/G145-5B/G145-5C still precede Rule TDD/production.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. Root next accepts separate Rule plan only; Rule TDD also awaits Phase-145/G145-5B/G145-5C; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
