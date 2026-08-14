# Mail engine boundary 0.1.8

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` remain accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; unrelated gates remain current. No source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Base/correction lifecycle `CONSUMED/HISTORICAL`; no plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness keep Rule Task 1/145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Current lifecycle (2026-08-14):** 02D retains docs-only acceptance. 02F correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and Rule-plan-next instructions are `CONSUMED/HISTORICAL`. This engine contract does not own persistence canonical hash v1; no engine/source/schema/runtime or Phase-145 carrier authority follows. Accepted Rule plan `669ef761` does not authorize Rule Task 1, which remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.

> **Status:** 144-01A remains `STATIC_ACCEPTED`; 144-03A Thread callable/encoding/oracle is
> `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
> `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` after independent
> Spec/API/Persistence, Security and Lifecycle/Evidence PASS reviews with C/I/M `0/0/0`;
> implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`), static only; Folder source/oracle `f8cb7643986fc591cae16764575337b72ef3a692` and docs `ca4e2854a319a020f7322ae0a9a98e3fb4a07374` are static-only accepted. The `f199546513b758ca7b95328b5ccd7680009b732e` / `6735ae746c7acde657359e8a709cd897c4dce61c` Rule lifecycle is `CONSUMED/HISTORICAL`; the standalone Rule plan is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4`, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`, after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`; no Rule source, schema, runtime, product or release authority follows.
> **Historical 144-02D dependency update:** docs-only boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after three exact PASS routes C/I/M `0/0/0`. The former sole-next 02F closure-review instruction is `CONSUMED/HISTORICAL`; the current exact action is stated in the lifecycle banner above. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.
> **Runtime:** NOT_RUN
> **Precedence:** this document replaces the 0.1.7 message/blob/draft portions of
> `ADR-MAIL-01-mail-engine-calendar-sync.md` and Phase 144-01A. Unmentioned methods and safe
> error behavior remain unchanged.
>
> **Selected-adapter gate (2026-08-13):** this amendment does not resolve Stalwart's absent
> manual-DNS verification result; G145-5C and founder questions 43-44 must bind that
> unmentioned method before packet 145-03C. Exact candidate source confirms stable identity
> ids exist through mailbox RFC 8621 `Identity/get|set`; packet 145-03A therefore uses a
> reviewed two-plane mapping and remains blocked by G145-5B mailbox authentication.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Hard limits

All limits count UTF-8 bytes for strings and raw bytes for byte slices. An adapter validates
before retaining a value. A provider-declared size is not trusted; it is only an opportunity
to reject earlier.

| Constant | Value | Applies to |
|---|---:|---|
| `MaxEngineOpaqueIDBytes` | 1,024 | account/folder/message/blob/submission/draft keys and versions |
| `MaxEngineCursorBytes` | 4,096 | cursor, checkpoint and opaque state |
| `MaxEnginePageItems` | 100 | one list/change/enumeration page |
| `MaxEngineProjectionItemBytes` | 1,048,576 | one retained projection item including strings/slices |
| `MaxEngineProjectionPageBytes` | 8,388,608 | all retained items plus cursor/state values in one page |
| `MaxEngineMessageBytes` | 52,428,800 | one authorized complete message/draft value or submitted MIME result |
| `MaxEngineAttachmentBytes` | 26,214,400 | one blob/uploaded attachment |
| `MaxEngineAddressItems` | 200 | each projection address collection; draft write uses OpenAPI 100 |
| `MaxEngineDraftRecipientItems` | 100 | each draft `to`, `cc` or `bcc` collection |
| `MaxEngineDraftTotalRecipientItems` | 300 | normalized unique recipients across draft `to` + `cc` + `bcc` |
| `MaxEngineReferenceItems` | 100 | adapter candidate references; accepted stored literals are strict byte-preserved case-sensitive ASCII Message-ID values |
| `MaxEngineFolderItems` | 100 | one account's non-deleted engine-backed folders |
| `MaxEngineKeywordItems` | 200 | one projection's reserved keyword markers |
| `MaxEngineAttachmentItems` | 50 | one message/draft attachment set |
| `MaxEngineAddressEmailBytes` | 320 | one normalized ASCII address |
| `MaxEngineAddressNameBytes` | 640 | one optional display name |
| `MaxEngineHeaderValueBytes` | 998 | subject, Message-ID, In-Reply-To or one References value |
| `MaxEngineSnippetBytes` | 4,096 | server-derived safe plain snippet |
| `MaxEngineMIMEFieldBytes` | 1,020 | filename, media type or Content-ID |
| `MaxEngineBodyPartBytes` | 8,388,608 | each normalized plain or rich body value |

Deployment policy may lower message, attachment and recipient ceilings but can never raise
them. `MailboxPolicy.max_recipients` is 1..300 and a draft must satisfy both the independent
100-item collection limits and `len(to)+len(cc)+len(bcc) <= max_recipients` after normalized
cross-collection deduplication. `max_attachment_bytes <= max_message_bytes` is mandatory.

## Address policy

`NormalizeAddress` is the only canonicalization rule for an `EngineAddress` at this boundary.
It validates the email as UTF-8 and rejects a raw email value longer than 320 bytes before
IDNA conversion. The raw value contains no Unicode whitespace, Unicode control character, or
header-injection character, and contains exactly one `@`. The local part is 1..64 ASCII bytes
and uses RFC 5322 `atext` dot-atom syntax only: a dot cannot be leading, trailing, or repeated.

The domain is converted with `idna.Lookup.ToASCII`, then ASCII-lowercased. Its canonical value
is 1..253 bytes; every label is 1..63 bytes of lowercase `a-z`, `0-9`, or interior hyphens only.
Empty labels and leading or trailing label hyphens reject. The canonical email lowercases the
ASCII local part and domain, has the form `local@domain`, and is at most 320 bytes. This policy
does not perform DNS, network, or deliverability checks: `.test`, `.invalid`, and synthetic
`u-*` local parts are valid deterministic fixtures when they otherwise satisfy this grammar.

`EngineAddress.Name` is optional valid UTF-8 with no Unicode control character, at most 160
runes and at most 640 bytes. Normalization preserves its exact input bytes/runes: no trimming,
NFC, whitespace collapsing, case folding, or other display-name transformation occurs.

`NormalizeAddressSet(addresses, maximum)` accepts only a configured `maximum` in 1..200,
rejects a collection above that maximum, normalizes each element in input order, and rejects a
canonical-email duplicate. It neither mutates the caller input nor returns a partial result on
any failure. A successful empty set is concrete and non-nil.

The only stable public address-error kinds are `invalid`, `too_large`, and `duplicate`.
Their errors are content-free: they do not expose a supplied address, display name, provider
detail, or other untrusted value. Address normalization changes neither OpenAPI nor persistence
in this packet.

## Port amendment

The following method signatures replace their 0.1.7 equivalents:

```go
ListMessages(context.Context, EngineMessageQuery) (EngineProjectionPage, error)
GetMessage(context.Context, EngineMessageReadRequest) (EngineMessage, error)
GetBlob(context.Context, EngineBlobReadRequest) (EngineBlob, error)
Changes(context.Context, EngineChangeRequest) (EngineChangePage, error)
EnumerateProjection(context.Context, EngineEnumerationRequest) (EngineEnumerationPage, error)
SaveDraft(context.Context, EngineDraft) (EngineMessage, error)
ReconcileDraft(context.Context, EngineDraftReconciliation) (EngineDraftResult, error)
Submit(context.Context, EngineSubmission) (EngineSubmissionResult, error)
```

### Closed request and page representation

The following field grammar is exact. An implementation may add methods, but it may not add
another exported data field, an untyped map or a provider-specific value to these DTOs:

```go
type EnginePageRequest struct {
    Limit int
}

type EngineMessageQuery struct {
    AccountID string
    FolderID  string // empty means every folder in the account
    Cursor    string // empty only for the first page
    Page      EnginePageRequest
}

type EngineMessageReadRequest struct {
    AccountID       string
    MessageID       string
    MaxContentBytes int64
}

type EngineBlobReadRequest struct {
    AccountID string
    BlobID    string
    MaxBytes  int64
}

type EngineChangeRequest struct {
    AccountID  string
    FromCursor string
    Page       EnginePageRequest
}

type EngineEnumerationRequest struct {
    AccountID  string
    BaseCursor string
    Checkpoint string
    Page       EnginePageRequest
}

type EngineProjectionPage struct {
    AccountID  string
    Cursor     string
    NextCursor string
    Bounds     EnginePageBounds
    Messages   []EngineMessageProjection
}

type EngineChangePage struct {
    AccountID        string
    FromCursor       string
    NextCursor       string
    Bounds           EnginePageBounds
    CreatedOrUpdated []EngineMessageProjection
    DestroyedIDs     []string
}

type EngineEnumerationPage struct {
    AccountID      string
    BaseCursor     string
    Checkpoint     string
    NextCheckpoint string
    Complete       bool
    Bounds         EnginePageBounds
    Messages       []EngineMessageProjection
}
```

The unchanged closed value shapes remain exactly `EnginePageBounds{Requested, Maximum,
Returned int}`, `EngineAddress{Email, Name string}` and
`EngineMessageFlags{Seen, Answered, Draft, Flagged, Junk bool}`. They do not acquire
provider ids, arbitrary headers or extra state.

`EnginePageRequest.Limit` is 1..100. `Bounds.Requested` echoes it,
`Bounds.Maximum == MaxEnginePageItems`, and `Bounds.Returned` equals the number of returned
messages, or for `EngineChangePage` the combined number of projections and destroyed ids.
The combined change collections are at most 100 and disjoint. Every slice above is concrete:
empty is a non-null zero-length slice. `NextCursor` is empty only on a terminal list/change
page. `EngineChangeRequest.FromCursor` is required and non-empty. Initial enumeration sends
both base cursor and checkpoint empty; the adapter captures
one composite base cursor. A resumed request must echo that base cursor and the immediately
preceding `NextCheckpoint`. A complete page has empty `NextCheckpoint`; any other empty or
mismatched resume value is invalid/restart-required, never an implicit fresh enumeration.

### Closed projection and content representation

```go
type EngineProjectionAttachment struct {
    BlobID      string
    ContentID   string
    Filename    string
    MediaType   string
    SizeBytes   int64
    Disposition EngineAttachmentDisposition
    SHA256      []byte // length 0 when unknown, otherwise exactly 32
}

type EngineMessageProjection struct {
    AccountID        string
    MessageID        string
    Version          string
    InternetMessageID string
    InReplyTo        string
    References       []string
    From             EngineAddress
    To               []EngineAddress
    CC               []EngineAddress
    BCC              []EngineAddress
    ReplyTo          []EngineAddress
    Subject          string
    Snippet          string
    SentAt           *time.Time
    ReceivedAt       *time.Time
    SizeBytes        int64
    Priority         EngineMessagePriority
    FolderIDs        []string
    Flags            EngineMessageFlags
    ReservedKeywords []EngineKeywordMarker
    Attachments      []EngineProjectionAttachment
    Origin           EngineProjectionOrigin
}

type EngineMessageBody struct {
    Format    EngineBodyFormat
    PlainBody string
    RichBody  string
}

type EngineMessage struct {
    Projection EngineMessageProjection
    Body       EngineMessageBody
}

type EngineDraftAttachment struct {
    AttachmentKey string
    ContentID     string
    Filename      string
    MediaType     string
    Disposition   EngineAttachmentDisposition
    SHA256        [32]byte
    Bytes         []byte
}

type EngineMessageContent struct {
    From        EngineAddress
    To          []EngineAddress
    CC          []EngineAddress
    BCC         []EngineAddress
    ReplyTo     []EngineAddress
    Subject     string
    InReplyTo   string
    References  []string
    Format      EngineBodyFormat
    PlainBody   string
    RichBody    string
    Priority    EngineMessagePriority
    Attachments []EngineDraftAttachment
}

type EngineBlob struct {
    AccountID string
    BlobID    string
    MediaType string
    SizeBytes int64
    Bytes     []byte
}

type EngineDraft struct {
    AccountID       string
    DraftKey        string
    MessageID       string
    PreviousVersion string
    MutationID      string
    MaxContentBytes int64
    Message         EngineMessageContent
}

type EngineDraftReconciliation struct {
    AccountID       string
    DraftKey        string
    MessageID       string
    PreviousVersion string
    MutationID      string
    MaxContentBytes int64
}

type EngineSubmission struct {
    AccountID       string
    DraftKey        string
    MessageID       string
    ExpectedVersion string
    SubmissionKey   string
    MaxMessageBytes int64
}
```

The enum sets are exact: body format `plain|rich`, priority `low|normal|high`, attachment
disposition `attachment|inline`, and projection origin
`inbound_or_unknown|confirmed_owner_submission|confirmed_owner_draft`. A projection carries
no body, raw MIME or attachment bytes. `EngineMessage.Body` contains only the authorized
decoded source body; all header/address/attachment metadata occurs exactly once in its
projection. On an engine read, `plain` requires empty `RichBody`; `rich` requires empty
`PlainBody`. A rich read is untrusted decoded provider HTML and must cross the Hub sanitizer
before API, search or persistence; Hub then derives plain text from the accepted sanitized
rich value. On an outbound `EngineMessageContent`, `plain` likewise requires empty
`RichBody`; `rich` carries the already server-sanitized `RichBody` plus the `PlainBody`
derived from that exact accepted value. No two independently supplied canonical bodies are
accepted. `EngineSubmission` is reference-only and submits the already confirmed engine
draft; it cannot contain or silently replace content. Its maximum is positive, no greater
than effective policy and no greater than `MaxEngineMessageBytes`, and applies to the final
submitted MIME result.

Every `EngineMessageProjection` has non-empty `AccountID`, `MessageID` and `Version`; the
opaque message id and version are each 1..`MaxEngineOpaqueIDBytes` and describe the same
accepted projection snapshot.

All address/reference/recipient/folder/keyword/attachment slices are concrete, unique after
normalization where identity has meaning, and obey their independent count ceilings.
Projection addresses use the 200-item ceiling per collection; draft `to`, `cc` and `bcc`
each use 100 and their normalized cross-collection combined count is at most
`min(MailboxPolicy.max_recipients, MaxEngineDraftTotalRecipientItems)`. `From` is one required
normalized address. BCC is non-empty only for confirmed-owner draft/submission origin.
`References` contains at most 100 strict byte-preserved case-sensitive ASCII Message-ID
literals. Projection folder ids contain
1..100 values. Attachment metadata contains 0..50 values. An inline attachment requires a
Content-ID. A projected MIME attachment may be empty; a draft upload has 1..effective
attachment maximum bytes, a required stable Hub attachment UUID-derived key and a confirmed
32-byte hash. That key is not a raw HTTP idempotency key.

Email is normalized ASCII and at most 320 bytes; display name is at most 640 UTF-8 bytes;
subject, Message-ID, In-Reply-To and each References value are at most 998 bytes; snippet is
at most 4,096 bytes; filename/media type/Content-ID are non-control UTF-8 values at most
1,020 bytes. Each body value is at most 8,388,608 UTF-8 bytes. `SizeBytes` is non-negative,
no greater than the hard message/attachment maximum applicable to its DTO. Optional header
strings use empty for absent. Optional timestamps are UTC when present. No arbitrary header
map exists.

Byte accounting is representation-independent and deterministic. `EngineVariableBytes(v)`
is the sum of `len` in bytes for every string and byte-slice value retained anywhere in `v`,
including values nested in addresses, headers, attachment metadata and collection elements.
Fixed-width scalars and struct/slice headers are not counted by this function because their
memory is independently bounded by the closed DTO shape and every collection cardinality.
An item satisfies `MaxEngineProjectionItemBytes` only when
`EngineVariableBytes(item) <= MaxEngineProjectionItemBytes`; a page satisfies the aggregate
limit only when the sum for all projections/destroyed ids plus every other page string,
including account/cursor/checkpoint/state values, is at most
`MaxEngineProjectionPageBytes`. JSON/protobuf encoding overhead, map iteration order,
allocator capacity and provider-declared sizes never participate. Projection DTOs contain
no unbounded map; every collection has an explicit count ceiling.

`EngineMessageReadRequest` contains `AccountID`, `MessageID` and `MaxContentBytes`.
`MaxContentBytes` is positive, no greater than the effective mailbox message policy and no
greater than `MaxEngineMessageBytes`. The returned `EngineMessage` contains exactly one
projection plus authorized content. Combined retained UTF-8 bytes of subject/address values,
plain/rich content and bounded attachment metadata, measured by `EngineVariableBytes`, cannot
exceed the supplied budget.

`EngineBlobReadRequest` contains `AccountID`, `BlobID` and `MaxBytes`. `MaxBytes` is positive,
no greater than the effective attachment policy and no greater than
`MaxEngineAttachmentBytes`. The adapter returns a complete byte value only. If the provider
size is unknown it reads through a limiting reader up to `MaxBytes + 1`; overflow discards
the partial value and returns `ErrEnginePayloadTooLarge` through
`EngineErrorPayloadTooLarge`. A successful value has `SizeBytes == int64(len(Bytes))`.
No partial-success result exists.

`EngineChangeRequest` and `EngineEnumerationRequest` carry account, exact cursor/checkpoint
and page request. Every page validates item count, each item byte count and aggregate page
bytes before returning. Overflow is a stable whole-page error: no cursor advances and no
partial page reaches the caller.

## Draft create identity

`EngineDraft` has required immutable `DraftKey` and required `MaxContentBytes`;
`EngineDraftReconciliation` uses required `AccountID`, `DraftKey`, `MutationID` and
`MaxContentBytes`, with empty `MessageID` and `PreviousVersion` only for the initial
create. The byte ceiling is positive, no greater than effective policy and no greater than
`MaxEngineMessageBytes`; both the outbound draft content and any `EngineMessage` returned by
`SaveDraft` or `EngineDraftApplied` fit it under `EngineVariableBytes`. `DraftKey` is derived
from the Hub draft UUID, is bounded by `MaxEngineOpaqueIDBytes`, and is never the raw HTTP
idempotency key.

The adapter must make initial create find-or-create idempotent on `(account, DraftKey)` and
must reconcile that same key before another create. Applied reconciliation returns the
opaque message id/version; definitive absence permits the domain to retry the same
`DraftKey`; indeterminate state permits no save. Later saves additionally require the last
confirmed message id/version plus a fresh stable mutation id.

## Submission result identity

The unchanged closed success value is
`EngineSubmissionSent{SubmissionID string, MessageID string}`. Both values are opaque engine
identifiers, are independently required and bounded by `MaxEngineOpaqueIDBytes`, and are
never Hub UUIDs. The persistence names are therefore `engine_submission_id` and
`engine_result_message_id`; an adapter or repository must not place either value in the
public `MailSubmission.message_id` field.

After `Submit` or `ReconcileSubmission` returns `EngineSubmissionSent`, the domain uses the
same account plus `MessageID` to obtain a bounded `EngineMessage`, atomically upserts its
projection (including the required projection `Version`) and only then records the resulting
Hub `MailMessage.id` snapshot and public `sent` receipt. A transient failure between engine
confirmation and that projection transaction remains reconciliation-required and never
permits a second submit.

The returned projection is accepted only when its `AccountID` exactly equals the requested
account, its `MessageID` exactly equals the confirmed opaque result message id and its
`Version` is non-empty and bounded. The domain derives the stored
`confirmed_owner_submission` origin from the persisted confirmed result pair; address,
folder membership and an adapter-provided origin are not independent proof. Any mismatch
retains reconciliation-required state and produces no Hub message binding, `sent` receipt or
outbox event.

Recovery has two disjoint branches. With no persisted opaque result pair, the domain calls
`ReconcileSubmission`; only definitive absence from that branch may permit the one controlled
retry. With a complete persisted pair, the domain does not call `ReconcileSubmission` and
repeats only the bounded `GetMessage` validation, projection upsert and finalization using
those exact ids. That branch can never requeue or call `Submit`.

## Validation and errors

- All opaque ids, versions, states, subject/header strings, addresses, media types and
  filenames reject NUL/control violations and over-limit UTF-8 bytes before retention.
- Slice/map limits and byte limits are both required; one cannot substitute for the other.
- `ErrEnginePayloadTooLarge` is the stable sentinel and
  `EngineErrorPayloadTooLarge` is the stable `payload_too_large` kind. Both expose only the
  safe string
  `mail engine payload too large`; it contains no id, address, body, size from the provider,
  URL, token or raw provider error.
- A list/change/enumeration overflow is retryable only with a smaller caller limit when the
  caller controls it. An individual over-limit provider object is quarantined/degraded and
  never silently truncated into canonical projection state.
- Blob/message reads, projection pages and reconciliation never perform implicit network
  retries after an indeterminate mutation.

## Required conformance oracles

1. Count, item and aggregate limits are independent: item 101, item bytes +1 and aggregate
   bytes +1 all fail without a next cursor. A 100-item page is valid only when its aggregate
   retained bytes also fit; a smaller caller limit is used when valid large items do not.
2. A page DTO is structurally unable to contain body or blob bytes.
3. Message/blob exact budget succeeds; +1 rejects before a successful value is visible.
4. An unknown-length blob is read through `max+1`, rejects as a whole and never returns the
   retained prefix.
5. Every message/blob/draft/submission request binds the server-resolved account and a
   positive bounded byte ceiling; page requests bind the account and a 1..100 limit.
6. A lost first draft-create response reconciles by stable draft key and creates one engine
   draft; changed raw HTTP keys cannot affect that identity.
7. Draft save/reconciliation rejects an expanded adapter result above its caller budget as a
   whole even when the original outbound draft fit.
8. Every over-limit error has the fixed safe text and preserves provider/body/id secrecy.
9. The external compile oracle proves the exact field grammar: page DTOs cannot carry body
   or attachment bytes, message metadata is not duplicated in the body, and submission is
   reference-only.
10. Combined normalized recipients at the configured maximum 300 are accepted when each
    collection is at most 100; recipient 301 rejects before any engine save or submit.
11. No direct Docker/host fallback is valid; executable evidence remains `NOT_RUN` until the
    accepted launcher exists.

## 144-03A Thread accepted-contract clarification

This engine contract remains count/item bounded only. A provider adapter may present up to 100
candidate References and one InReplyTo, but the Thread callable accepts only the strict
byte-preserved case-sensitive ASCII Message-ID literal defined in `MAIL-SPEC.md`; malformed
candidates are skipped by Thread, not normalized, and only after a definitive successful
ancestor lookup outcome. The adapter preflights raw candidate bytes before any
UTF-8/Message-ID/IDNA work; it does not alter the accepted 144-01A DTO/source or its structural
acceptance.

## 144-03B Folder lifecycle clarification

`MaxEngineOpaqueIDBytes` and every string engine account/folder identifier in this adapter
contract apply only to opaque engine DTO values. The Folder pure tree
policy receives only projection/persistence-mapped Hub `uuid.UUID` account, folder and parent
rows; it never validates, sorts or accepts engine IDs. Source/oracle for this narrow policy are
`STATIC_ACCEPTED`, static-only, at `f8cb7643986fc591cae16764575337b72ef3a692`; this does not
alter the adapter ABI or establish executable validation: compile, tests, build and runtime remain
`NOT_RUN`.

## 144-03B Rule compiler / adapter split

The docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` Rule contract compiles typed rules to provider-neutral IR,
not an engine port or Sieve bytes. Future Phase 145 separately resolves engine folder ID (not Sieve
mailbox name), renders/escapes, checks RFC 9661 and carries rendered digest. Attachment IR is
capability-blocked without an accepted MIME mapping.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan-next text is `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. Accepted Rule plan `669ef761` does not authorize Task 1, which remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`; Task4 is post-future-Task3 source acceptance only. 144-01B remains Docker-blocked, Phase 142/148 and release pending, launcher `NOT_READY`, executables `NOT_RUN`.
