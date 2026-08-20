# 144-04F sync-page policy source-authority receipt

## Status and authority

Candidate `4e608e890ec12aeed8de965f9649ae503260b94e` (parent
`c53a5fe60e1283cab58be5efaec7ffa131145b31`, tree
`ed7e2bdab6a44254a7c677603ad0afc298aafed8`, subject `docs(mail): propose sync policy authority`)
is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` for its immutable exact9 tree blobs. Contract,
Security and Lifecycle exact-SHA reviews PASS C/I/M `0/0/0`; provenance is dispatcher-enforced, no
fallback, telemetry `UNKNOWN`. Candidate/precommit/commit-review/separate-acceptance-next clauses
are `CONSUMED/HISTORICAL`. The previously committed 144-04F authority acceptance sync records that existing receipt; it is
not a new acceptance claim. The four implementation files remain `UNIMPLEMENTED/UNACCEPTED` and
executables/runtime `NOT_RUN`.

The sole future source manifest is canonical and test-first:

1. `backend/tests/mail/mailbox_events_test.go`
2. `backend/tests/mail/mailbox_sync_policy_test.go`
3. `backend/internal/mail/mailbox_events.go`
4. `backend/internal/mail/mailbox_sync_policy.go`

No substitute, extra file or reordered path is permitted. Root may now dispatch exactly these four
paths and no others under the accepted authority receipt.

The narrow event-builder edit changes only `mail.received.v1.folder_ids`,
`mail.message.changed.v1.folder_ids`, and `mail.message.changed.v1.label_ids` from max 50 to
100: 100 passes and 101 rejects; field names, subjects and versions stay unchanged. All new policy
types live only in `mailbox_sync_policy.go`; `mailbox_types.go` is not edited.

## Frozen ABI

The only new callable is
`ClassifyMailboxSyncPage(MailboxSyncPolicyInput) (MailboxSyncPolicyResult, error)`. The exact
new types are:

```go
type MailboxSyncPolicyInput struct {
    Page MailboxSyncPolicyPage; Folders []MailboxSyncPolicyFolder; Messages []MailboxSyncPolicyMessage
    Destroyed []MailboxSyncPolicyDestroyed; Labels []MailboxSyncPolicyLabelDefinition
    SyncEpoch int64; EventCutoff time.Time; Event MailboxSyncPolicyEventContext
}
type MailboxSyncPolicyPage struct { Change *EngineChangePage; Enumeration *EngineEnumerationPage }
type MailboxSyncPolicyFolder struct { Snapshot EngineFolder; Folder MailboxEventID }
type MailboxSyncPolicyMessage struct {
    EngineMessageID string; Message MailboxEventID; Thread MailboxSyncPolicyThreadResolution
    Previous *MailboxSyncPolicyPreviousMessage; Provenance MailboxSyncPolicyProvenance
    ReceivedEventID uuid.UUID; ChangedEvent MailboxSyncPolicyEventReservation; CalendarMethod *MailCalendarMethod
}
type MailboxSyncPolicyThreadResolution struct {
    Thread MailboxEventID; KeyVersion ThreadKeyVersion; Key ThreadKey
    Ancestors []ThreadAncestor; AncestorLookupComplete bool
}
type MailboxSyncPolicyPreviousMessage struct {
    Thread MailboxEventID; FolderIDs []MailboxEventID; Flags EngineMessageFlags
    Labels []MailboxSyncPolicyLabelState; ReceivedEventID *uuid.UUID
}
type MailboxSyncPolicyProvenance struct {
    ConfirmedSubmission *MailboxSyncPolicyProvenanceProof; ConfirmedDraft *MailboxSyncPolicyProvenanceProof
}
type MailboxSyncPolicyProvenanceProof struct { Relation MailboxEventID; EngineMessageID string }
type MailboxSyncPolicyEventReservation struct { ID uuid.UUID; Committed bool }
type MailboxSyncPolicyDestroyed struct {
    EngineMessageID string; Message MailboxEventID; Thread MailboxEventID
    Labels []MailboxSyncPolicyLabelState; ChangedEvent MailboxSyncPolicyEventReservation
}
type MailboxSyncPolicyLabelDefinition struct { Label MailboxEventID; EngineKeyword EngineKeywordMarker }
type MailboxSyncPolicyLabelState struct { Label MailboxEventID; Manual bool; EngineRule bool }
type MailboxSyncPolicyEventContext struct {
    Mailbox MailboxEventID; OrganizationID uuid.UUID; ActorID *uuid.UUID
    OccurredAt time.Time; CorrelationID string
}
type MailboxSyncPolicyResult struct {
    Advance MailboxSyncPolicyAdvance; Folders []MailboxSyncPolicyFolderIntent
    Upserts []MailboxSyncPolicyMessageIntent; Tombstones []MailboxSyncPolicyTombstoneIntent
    Counts MailboxSyncPolicyCountIntent; Events []events.EnvelopeV1
}
type MailboxSyncPolicyAdvance struct {
    SyncEpoch int64; Change *MailboxSyncPolicyChangeAdvance; Enumeration *MailboxSyncPolicyEnumerationAdvance
}
type MailboxSyncPolicyChangeAdvance struct { FromCursor, NextCursor string }
type MailboxSyncPolicyEnumerationAdvance struct { BaseCursor, Checkpoint, NextCheckpoint string; Complete bool }
type MailboxSyncPolicyFolderIntent struct { Folder MailboxEventID; Snapshot EngineFolder }
type MailboxSyncPolicyMessageIntent struct {
    Message, Thread MailboxEventID; ThreadIdentity ThreadIdentity; Projection EngineMessageProjection
    FolderIDs []MailboxEventID; Labels []MailboxSyncPolicyLabelState; ReceivedEventID *uuid.UUID
}
type MailboxSyncPolicyTombstoneIntent struct {
    EngineMessageID string; Message MailboxEventID; Labels []MailboxSyncPolicyLabelState
}
type MailboxSyncPolicyCountIntent struct { Folders, Threads []MailboxEventID }
```

`MailboxSyncPolicyPage` and `MailboxSyncPolicyAdvance` are exact-one unions. Error ABI is exact:

```go
type MailboxSyncPolicyErrorKind string
const (
    MailboxSyncPolicyErrorInvalid = "invalid"
    MailboxSyncPolicyErrorTooLarge = "too_large"
    MailboxSyncPolicyErrorDuplicate = "duplicate"
    MailboxSyncPolicyErrorAccountMismatch = "account_mismatch"
)
var (
    ErrMailboxSyncPolicyInvalid = errors.New("mailbox sync policy invalid")
    ErrMailboxSyncPolicyTooLarge = errors.New("mailbox sync policy exceeds bounds")
    ErrMailboxSyncPolicyDuplicate = errors.New("mailbox sync policy duplicate")
    ErrMailboxSyncPolicyAccountMismatch = errors.New("mailbox sync policy account mismatch")
)
type MailboxSyncPolicyError struct{Kind MailboxSyncPolicyErrorKind}
```

`Error()` is content-free and nil-safe: each known kind returns its matching sentinel text;
unknown kind or nil receiver returns stable `mailbox sync policy error`. `Unwrap()` returns only
the matching sentinel or nil, so `errors.Is` works without exposing input/detail.

## Exact classification constraints

The caller supplies all-and-only `Messages`/`Destroyed` state records, every Hub ID/prior state,
proof, ancestor and reservation; the policy generates/queries nothing and validates same-account
identity exactly. Every normalized engine page requires `Bounds.Requested` in `1..100`,
`Bounds.Maximum == 100`, exact `Bounds.Returned`, and concrete nonnil engine page slices even when
empty; every accepted aggregate byte budget from `ENGINE-CONTRACT` is enforced and overflow is
`too_large`. Change pages require disjoint
`len(CreatedOrUpdated)+len(DestroyedIDs) <= 100`, with `Bounds.Returned` equal to that combined
count. Enumeration requires `len(Messages) <= 100` and exact `Returned`; folders independently
are `<=100`; label catalog, folder/previous/desired sets and event reservations are `<=100`.
Duplicates fail. Output is deep-copied: folders raw FolderID bytes, upserts/tombstones/events raw
engine message-ID bytes, all Hub-ID collections raw 16 UUID bytes. `DeriveThreadIdentity` matches
the supplied key/version and supplied Thread ID; `NormalizeAddress` is used.

Exact same-account `engine_keyword` mapping preserves `Manual`, replaces `EngineRule` from current
known keywords, ignores unknown markers and clears raw markers. Provenance has zero-or-one proof
total (submission or draft, never both); its Relation is same-account, its EngineMessageID is
byte-equal to the current projection ID, and it agrees exactly with Origin (submission versus
draft; inbound has neither). A nonempty BCC is
allowed only with matching confirmed proof/origin; inbound/unknown BCC is empty. Missing, dual,
engine-ID, cross-account, origin or BCC mismatches are `invalid` or `account_mismatch` as
appropriate. Existing markers/committed reservations suppress replay. An inbound-or-unknown
first-seen message needs nonnil `ReceivedAt` strictly after cutoff and nonzero `ReceivedEventID`;
stable matching confirmed submission/draft suppresses, spoofed From/alias/mutable Sent never does.
Changed field order is
`folder_ids,seen,answered,flagged,junk,labels`; flag order is `seen,answered,flagged,junk,draft`;
only an exact allowlisted projection delta emits changed; deleted is sole. Destroy uses empty live
folders/flags and preserves labels. Event-required records
need nonzero reservations; illegal zero/unused combinations are invalid. Only existing
`NewMailReceivedEnvelopeV1`/`NewMailMessageChangedEnvelopeV1` build events; cursors/checkpoints
live only in Advance, never events/errors; no BCC/raw body/MIME/provider text.

## Frozen policy boundary

The future unit is pure, deterministic and bounded: one account, no I/O, clock, ID generation,
DB/Ent/network/global state or raw body/MIME/cursor/provider text. It returns projection/event
intents only; no callbacks. The exact external oracle also covers union invalidity, missing/extra
state, 99+1/100+0 pass, 100+1/overlap/Returned mismatch fail, each 100/101 boundary, byte-identical
shuffle output, threading/provenance/cutoff/replay/noop/delta/destroy/counts/deep-copy/no-panic and
no secret leakage. It includes Bounds fields/nil slices/aggregate byte budget and proof
dual/missing/engine-ID/cross-account/origin/BCC rows.

## Boundaries and self-review

The accepted candidate changed exactly nine documentation paths: `.planning/STATE.md`,
`.planning/decisions/specs/MAIL-SPEC.md`, `.planning/phases/144-mailbox-domain/PLAN.md`,
`.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`,
`.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`, this review,
`docs/mail/CONTRACT-CHANGELOG.md`, `docs/mail/QA-HANDOFF.md`, and
`docs/mail/events-v1.schema.json`. It does not alter source, OpenAPI or persistence. Preserve accepted 02C source `e565523f589fbd7536f6f8e4df7effd39469c1af`
and receipt `c53a5fe6`; this receipt selects only the exact4 manifest above. 02E, Ent generation/migration,
services/API/UI/product/phase/release, 02F-D, engine deletion/User-GDPR remain outside scope.
Launcher is `NOT_READY`; executables/runtime are `NOT_RUN`.

Author self-review of the previously committed 144-04F authority acceptance sync is C/I/M only and is not a new acceptance claim. Record
exact-nine-path scope, `staged=0`, `git diff --check`, stale/authority review and an exact blob map.
The stale/authority review confirms candidate/precommit/no-dispatch/acceptance-sync-next wording is
historical, while unrelated historical 144-03B and Thread exact-two records remain preserved and do
not grant 144-04F authority.
