# 144-04F sync-page policy source-authority receipt

## 143-03B-T immutable source acceptance and conditional 143-04C gate (2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; 144-04G still requires separate docs authority because its repository port is absent. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## 144-04F immutable source acceptance (2026-08-20)

Mail 144-04F is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `4715f939f24e5b5c2d3f80a38d204babe68bdd34` (parent `7ac8714e99db2935ce290ad67836e1309cc05c78`, tree `6461894a90c4340bf96118ed1a242516c02b48c7`, subject `feat(mail): add sync page policy`). Its exact4 ordered test-first blobs are recorded in `IMPLEMENTATION-144-04F.md`: events test `50a41cefb539d9b840a43da9e1ff34cce6c8653f`, sync-policy test `b6410709e4fb8166c28a5c502cfba535eef08d06`, events source `a3017b0a6064d8fc1017e37e2d4f675ee8732945`, sync-policy source `acd642fcdb6290af79b7fa8e5aa36a5036576f47`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file asserting that Mail exact4 is dirty, `UNCOMMITTED`, `UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is `CONSUMED/HISTORICAL`; the immutable source receipt above is current. No automatic Mail next source follows: 144-04G remains `NOT_AUTHORIZED / BLOCKED` pending separate docs authority because its required repository port is absent.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted. 143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; 143-04A/B/G receipts remain accepted. Launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. 143-03E, 144-02E, Ent generation and migration stay blocked. No raw iTIP or service/API/UI/product/phase/release authority follows.

## Combined exact17 conditional acceptance sync (2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

## Status and authority

Candidate `4e608e890ec12aeed8de965f9649ae503260b94e` (parent
`c53a5fe60e1283cab58be5efaec7ffa131145b31`, tree
`ed7e2bdab6a44254a7c677603ad0afc298aafed8`, subject `docs(mail): propose sync policy authority`)
is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` for its immutable exact9 tree blobs. Contract,
Security and Lifecycle exact-SHA reviews PASS C/I/M `0/0/0`; provenance is dispatcher-enforced, no
fallback, telemetry `UNKNOWN`. Candidate/precommit/commit-review/separate-acceptance-next clauses
are `CONSUMED/HISTORICAL` for that receipt. The previously committed 144-04F authority acceptance
sync records that existing receipt; it is not a new acceptance claim. Existing uncommitted exact4
source/test candidates remain `UNACCEPTED`; executables/runtime are `NOT_RUN`.

**ROOT_CORRECTION_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN — cutoff sentences only:** this candidate corrects only the first-seen inbound/unknown receipt-cutoff rule below. The former immediate exact4 root-dispatch clause is `SUSPENDED / CONSUMED/HISTORICAL` for this route. The existing exact4 remains only the future manifest, and its uncommitted source/test candidates MUST NOT be committed until Contract/Security/Lifecycle precommit PASS, one atomic exact9 correction-candidate commit, all three immutable exact-SHA PASS, and a separate docs-only acceptance sync. This is not correction acceptance and changes no other 144-04F behavior.

The sole future source manifest is canonical and test-first:

1. `backend/tests/mail/mailbox_events_test.go`
2. `backend/tests/mail/mailbox_sync_policy_test.go`
3. `backend/internal/mail/mailbox_events.go`
4. `backend/internal/mail/mailbox_sync_policy.go`

No substitute, extra file or reordered path is permitted. The former statement permitting root to
dispatch these four immediately is `SUSPENDED / CONSUMED/HISTORICAL`; this list is only the future
manifest pending the correction lifecycle above.

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
first-seen message needs nonnil `ReceivedAt`. If it is strictly after `EventCutoff`, it needs a
nonzero `ReceivedEventID`, persists that marker and emits exactly one `mail.received.v1`. If it is
before or equal to `EventCutoff`, it accepts only a zero reservation, upserts with a nil marker and
emits no received event; a supplied historical/equal reservation is `invalid`. Stable matching
confirmed submission/draft never emits received; spoofed From/alias/mutable Sent never proves that
state.
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

Author self-review of this exact9 cutoff-correction candidate is C/I/M only and is not an acceptance
claim. Record exact-nine-path scope, `staged=0`, `git diff --check`, stale/authority review and an
exact blob map. The stale/authority review confirms only the original accepted receipt's
candidate/precommit/immediate-dispatch lifecycle is `CONSUMED/HISTORICAL`; the correction route
above is current and remains precommit-blocked. Unrelated historical 144-03B and Thread exact-two
records remain preserved and do not grant 144-04F authority.
