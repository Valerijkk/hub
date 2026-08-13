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
| `MaxEngineFolderItems` | 100 | one account's non-deleted engine-backed folders |
| `MaxEngineKeywordItems` | 200 | one projection's reserved keyword markers |
| `MaxEngineAttachmentItems` | 50 | one message/draft attachment set |

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
```

`EngineProjectionPage`, `EngineChangePage` and `EngineEnumerationPage` contain only
`EngineMessageProjection` values. A projection contains bounded opaque ids/version/state,
normalized address/header metadata, flags, folder ids, reserved keywords, size/timestamps,
attachment metadata and confirmed origin. It structurally contains no plain body, rich body,
raw MIME or attachment bytes.

Byte accounting is representation-independent and deterministic. `EngineVariableBytes(v)`
is the sum of `len` in bytes for every string and byte-slice value retained anywhere in `v`,
including values nested in addresses, headers, attachment metadata and collection elements.
Fixed-width scalars and struct/slice headers are not counted by this function because their
memory is independently bounded by the closed DTO shape and every collection cardinality.
An item satisfies `MaxEngineProjectionItemBytes` only when
`EngineVariableBytes(item) <= MaxEngineProjectionItemBytes`; a page satisfies the aggregate
limit only when the sum for all items plus every returned cursor/checkpoint/state string is
at most `MaxEngineProjectionPageBytes`. JSON/protobuf encoding overhead, map iteration order,
allocator capacity and provider-declared sizes never participate. Projection DTOs contain no
unbounded map; every collection has an explicit count ceiling.

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
the partial value and returns `EnginePayloadTooLarge`. No partial-success result exists.

`EngineChangeRequest` and `EngineEnumerationRequest` carry account, exact cursor/checkpoint
and page request. Every page validates item count, each item byte count and aggregate page
bytes before returning. Overflow is a stable whole-page error: no cursor advances and no
partial page reaches the caller.

## Draft create identity

`EngineDraft` adds required immutable `DraftKey` and required `MaxContentBytes`;
`EngineDraftReconciliation` uses required `AccountID`, `DraftKey`, `MutationID` and
`MaxContentBytes`, with nullable/empty `MessageID` and `PreviousVersion` only for the initial
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
- `EnginePayloadTooLarge` is a stable sentinel/error kind with the safe string
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
5. Every request binds the server-resolved account and a positive bounded byte ceiling.
6. A lost first draft-create response reconciles by stable draft key and creates one engine
   draft; changed raw HTTP keys cannot affect that identity.
7. Draft save/reconciliation rejects an expanded adapter result above its caller budget as a
   whole even when the original outbound draft fit.
8. Every over-limit error has the fixed safe text and preserves provider/body/id secrecy.
9. No direct Docker/host fallback is valid; executable evidence remains `NOT_RUN` until the
   accepted launcher exists.
