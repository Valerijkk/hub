# Mail engine boundary 0.1.8

> **Status:** root-frozen contract; independent Terra/high review pending
> **Runtime:** NOT_RUN
> **Precedence:** this document replaces the 0.1.7 message/blob/draft portions of
> `ADR-MAIL-01-mail-engine-calendar-sync.md` and Phase 144-01A. Unmentioned methods and safe
> error behavior remain unchanged.

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
| `MaxEngineReferenceItems` | 100 | normalized RFC Message-ID references |
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
them. `max_attachment_bytes <= max_message_bytes` is mandatory.

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
normalized body pair; all header/address/attachment metadata occurs exactly once in its
projection. `EngineSubmission` is reference-only and submits the already confirmed engine
draft; it cannot contain or silently replace content. Its maximum is positive, no greater
than effective policy and no greater than `MaxEngineMessageBytes`, and applies to the final
submitted MIME result.

All address/reference/recipient/folder/keyword/attachment slices are concrete, unique after
normalization where identity has meaning, and obey their independent count ceilings.
Projection addresses use the 200-item ceiling per collection; draft `to`, `cc` and `bcc`
each use 100 and their combined count also obeys effective recipient policy. `From` is one
required normalized address. BCC is non-empty only for confirmed-owner draft/submission
origin. `References` contains at most 100 normalized values. Projection folder ids contain
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
10. No direct Docker/host fallback is valid; executable evidence remains `NOT_RUN` until the
   accepted launcher exists.
