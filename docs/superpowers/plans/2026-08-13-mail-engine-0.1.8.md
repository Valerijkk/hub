# Mail Engine 0.1.8 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `superpowers:subagent-driven-development` to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the historical hydrated Mail engine port with the closed, projection-only,
byte-bounded 0.1.8 port without adding provider behavior or persistence.

**Architecture:** `backend/internal/mail` remains the provider-neutral public boundary.
List/change/enumeration calls carry only bounded projection metadata; one authorized read
returns one bounded body, one blob read returns one complete bounded byte value, draft create
has a stable engine key, and submission references the confirmed draft instead of copying
content. This packet defines types and compile-time contracts only; the deterministic fake
and Stalwart adapter own behavioral conformance in later packets.

**Tech Stack:** Go 1.25, standard library only, external-package Go unit contract.

## Global Constraints

- Exact contract commit before execution: the clean branch must contain
  `c0f204983fdc929bebeacf75ceb69a416b8b5eb3` and the latest docs-only plan commit.
- Entry gate: an independent `gpt-5.6-terra` / `high` review with accepted per-dispatch
  provenance must accept the exact contract SHA. Current status is
  `BLOCKED_INDEPENDENT_CONTRACT_REVIEW_PENDING`.
- Worker runtime: exactly `gpt-5.6-terra` / `xhigh`, fallback none. For this branch/Phases
  143–148, an accepted native dispatcher record without rejection is
  `provenance_source: dispatcher-enforced`; actual telemetry, when exposed, must match.
- Allowed implementation files only:
  `backend/internal/mail/mailbox_engine.go`,
  `backend/internal/mail/mailbox_types.go`,
  `backend/internal/mail/mailbox_errors.go`, and
  `backend/tests/mail/mailbox_engine_contract_test.go`.
- Do not read or change `testframework/**`, `backend/testframework/**`,
  `.planning/audits/test-rebase/**`, phases 128–141/TASK-193–234, `pro-qa-ai/**`,
  `.env.local`, or `.git/index.lock`.
- No dependency, provider/JMAP value, persistence, filesystem, network, DB, broker, listener,
  wall clock or global mutable state.
- Host tests/build/lint/validators are forbidden. Direct Docker/Compose is not a fallback.
  Until the task-owned launcher is accepted, runtime evidence is exactly `NOT_RUN`.
- Comments on exported Go identifiers are in Russian. No AI attribution in commits.
- The implementation commit may be created after static review, but does not become accepted
  runtime evidence until the supported Docker job is current-SHA green.

---

## File Structure

- `backend/internal/mail/mailbox_engine.go` — the provider-neutral interface; only seven read/
  draft signatures change, while all other historical methods remain byte-for-byte stable.
- `backend/internal/mail/mailbox_types.go` — all constants, closed requests, projections,
  bodies, attachments, draft/submission values and existing unchanged engine value types.
- `backend/internal/mail/mailbox_errors.go` — the existing safe taxonomy plus one bounded
  payload sentinel/kind.
- `backend/tests/mail/mailbox_engine_contract_test.go` — external consumer compile oracle,
  exact-field/constant/security assertions and no adapter implementation.

---

### Task 1: Freeze the external consumer oracle first

**Files:**

- Modify: `backend/tests/mail/mailbox_engine_contract_test.go`

**Interfaces:**

- Consumes: the frozen names and values in `docs/mail/ENGINE-CONTRACT-0.1.8.md`.
- Produces: a deliberately non-compiling external test until Task 2 adds the exact 0.1.8
  public surface.

- [ ] **Step 1: Confirm execution gates before any write**

Record all of the following in the worker response before opening the file:

```text
branch=codex/mail-module-20260813
contract_ancestor=c0f204983fdc929bebeacf75ceb69a416b8b5eb3
worktree=clean
requested_model=gpt-5.6-terra/xhigh
provenance_source=dispatcher-enforced
dispatcher_acceptance=accepted_without_rejection
actual_model=<matching telemetry|UNKNOWN>
actual_reasoning_effort=<matching telemetry|UNKNOWN>
independent_contract_review=PASS
```

If the branch, ancestor, clean state or independent review cannot be proven, or the route has
a rejection, omitted model/effort, fallback or telemetry mismatch, return `BLOCKED` and make
no write. Unavailable telemetry alone is recorded as `UNKNOWN` with dispatcher-enforced
provenance.

- [ ] **Step 2: Change only the consumer stub signatures**

Replace the seven historical methods with:

```go
func (consumerStub) ListMessages(context.Context, mail.EngineMessageQuery) (mail.EngineProjectionPage, error) {
	return mail.EngineProjectionPage{}, nil
}

func (consumerStub) GetMessage(context.Context, mail.EngineMessageReadRequest) (mail.EngineMessage, error) {
	return mail.EngineMessage{}, nil
}

func (consumerStub) GetBlob(context.Context, mail.EngineBlobReadRequest) (mail.EngineBlob, error) {
	return mail.EngineBlob{}, nil
}

func (consumerStub) Changes(context.Context, mail.EngineChangeRequest) (mail.EngineChangePage, error) {
	return mail.EngineChangePage{}, nil
}

func (consumerStub) EnumerateProjection(context.Context, mail.EngineEnumerationRequest) (mail.EngineEnumerationPage, error) {
	return mail.EngineEnumerationPage{}, nil
}

func (consumerStub) SaveDraft(context.Context, mail.EngineDraft) (mail.EngineMessage, error) {
	return mail.EngineMessage{}, nil
}

func (consumerStub) ReconcileDraft(context.Context, mail.EngineDraftReconciliation) (mail.EngineDraftResult, error) {
	return mail.EngineDraftAbsent{}, nil
}
```

Keep `Submit(context.Context, mail.EngineSubmission)` and every unrelated method name stable;
the changed `EngineSubmission` fields are asserted below.

- [ ] **Step 3: Add an exact exported-field helper**

Add this test-only helper:

```go
func assertExactFields(t *testing.T, valueType reflect.Type, expected ...string) {
	t.Helper()
	if valueType.NumField() != len(expected) {
		t.Fatalf("%s field count = %d, want %d", valueType.Name(), valueType.NumField(), len(expected))
	}
	for index, name := range expected {
		if field := valueType.Field(index); field.Name != name {
			t.Fatalf("%s field %d = %s, want %s", valueType.Name(), index, field.Name, name)
		}
	}
}
```

- [ ] **Step 4: Add exact hard-limit assertions**

```go
func TestMailEngine_HardLimitsMatchContract018(t *testing.T) {
	want := map[string]int{
		"opaque": 1024, "cursor": 4096, "page": 100,
		"projection_item": 1048576, "projection_page": 8388608,
		"message": 52428800, "attachment": 26214400,
		"projection_addresses": 200, "draft_recipients": 100,
		"references": 100, "folders": 100, "keywords": 200,
		"attachments": 50, "email": 320, "display_name": 640,
		"header": 998, "snippet": 4096, "mime_field": 1020,
		"body_part": 8388608,
	}
	got := map[string]int{
		"opaque": mail.MaxEngineOpaqueIDBytes,
		"cursor": mail.MaxEngineCursorBytes,
		"page": mail.MaxEnginePageItems,
		"projection_item": mail.MaxEngineProjectionItemBytes,
		"projection_page": mail.MaxEngineProjectionPageBytes,
		"message": mail.MaxEngineMessageBytes,
		"attachment": mail.MaxEngineAttachmentBytes,
		"projection_addresses": mail.MaxEngineAddressItems,
		"draft_recipients": mail.MaxEngineDraftRecipientItems,
		"references": mail.MaxEngineReferenceItems,
		"folders": mail.MaxEngineFolderItems,
		"keywords": mail.MaxEngineKeywordItems,
		"attachments": mail.MaxEngineAttachmentItems,
		"email": mail.MaxEngineAddressEmailBytes,
		"display_name": mail.MaxEngineAddressNameBytes,
		"header": mail.MaxEngineHeaderValueBytes,
		"snippet": mail.MaxEngineSnippetBytes,
		"mime_field": mail.MaxEngineMIMEFieldBytes,
		"body_part": mail.MaxEngineBodyPartBytes,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hard limits = %#v, want %#v", got, want)
	}
}
```

- [ ] **Step 5: Add closed request/page shape assertions**

```go
func TestMailEngine_RequestAndPageShapesAreClosed(t *testing.T) {
	assertExactFields(t, reflect.TypeOf(mail.EnginePageRequest{}), "Limit")
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageQuery{}), "AccountID", "FolderID", "Cursor", "Page")
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageReadRequest{}), "AccountID", "MessageID", "MaxContentBytes")
	assertExactFields(t, reflect.TypeOf(mail.EngineBlobReadRequest{}), "AccountID", "BlobID", "MaxBytes")
	assertExactFields(t, reflect.TypeOf(mail.EngineChangeRequest{}), "AccountID", "FromCursor", "Page")
	assertExactFields(t, reflect.TypeOf(mail.EngineEnumerationRequest{}), "AccountID", "BaseCursor", "Checkpoint", "Page")
	assertExactFields(t, reflect.TypeOf(mail.EngineProjectionPage{}), "AccountID", "Cursor", "NextCursor", "Bounds", "Messages")
	assertExactFields(t, reflect.TypeOf(mail.EngineChangePage{}), "AccountID", "FromCursor", "NextCursor", "Bounds", "CreatedOrUpdated", "DestroyedIDs")
	assertExactFields(t, reflect.TypeOf(mail.EngineEnumerationPage{}), "AccountID", "BaseCursor", "Checkpoint", "NextCheckpoint", "Complete", "Bounds", "Messages")
}
```

- [ ] **Step 6: Add closed projection/content/submission assertions**

```go
func TestMailEngine_ContentAndSubmissionShapesAreClosed(t *testing.T) {
	assertExactFields(t, reflect.TypeOf(mail.EngineProjectionAttachment{}),
		"BlobID", "ContentID", "Filename", "MediaType", "SizeBytes", "Disposition", "SHA256")
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageProjection{}),
		"AccountID", "MessageID", "Version", "InternetMessageID", "InReplyTo", "References",
		"From", "To", "CC", "BCC", "ReplyTo", "Subject", "Snippet", "SentAt", "ReceivedAt",
		"SizeBytes", "Priority", "FolderIDs", "Flags", "ReservedKeywords", "Attachments", "Origin")
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageBody{}), "Format", "PlainBody", "RichBody")
	assertExactFields(t, reflect.TypeOf(mail.EngineMessage{}), "Projection", "Body")
	assertExactFields(t, reflect.TypeOf(mail.EngineDraftAttachment{}),
		"AttachmentKey", "ContentID", "Filename", "MediaType", "Disposition", "SHA256", "Bytes")
	assertExactFields(t, reflect.TypeOf(mail.EngineMessageContent{}),
		"From", "To", "CC", "BCC", "ReplyTo", "Subject", "InReplyTo", "References", "Format",
		"PlainBody", "RichBody", "Priority", "Attachments")
	assertExactFields(t, reflect.TypeOf(mail.EngineDraft{}),
		"AccountID", "DraftKey", "MessageID", "PreviousVersion", "MutationID", "MaxContentBytes", "Message")
	assertExactFields(t, reflect.TypeOf(mail.EngineDraftReconciliation{}),
		"AccountID", "DraftKey", "MessageID", "PreviousVersion", "MutationID", "MaxContentBytes")
	assertExactFields(t, reflect.TypeOf(mail.EngineSubmission{}),
		"AccountID", "DraftKey", "MessageID", "ExpectedVersion", "SubmissionKey", "MaxMessageBytes")
	assertNoFields(t, reflect.TypeOf(mail.EngineMessageProjection{}), "Content", "Body", "Bytes", "RawMIME")
	assertNoFields(t, reflect.TypeOf(mail.EngineSubmission{}), "Message", "Content", "Attachments", "Bytes", "RawMIME")
}
```

- [ ] **Step 7: Add enum and safe-error assertions**

```go
func TestMailEngine_Contract018EnumsAndPayloadErrorAreStable(t *testing.T) {
	if mail.EngineBodyPlain != "plain" || mail.EngineBodyRich != "rich" ||
		mail.EnginePriorityLow != "low" || mail.EnginePriorityNormal != "normal" ||
		mail.EnginePriorityHigh != "high" ||
		mail.EngineAttachmentDispositionAttachment != "attachment" ||
		mail.EngineAttachmentDispositionInline != "inline" ||
		mail.EngineProjectionOriginConfirmedOwnerDraft != "confirmed_owner_draft" {
		t.Fatal("0.1.8 enum value drift")
	}
	err := &mail.EngineError{Kind: mail.EngineErrorPayloadTooLarge}
	if !errors.Is(err, mail.ErrEnginePayloadTooLarge) {
		t.Fatal("payload-too-large must match its stable sentinel")
	}
	if got := err.Error(); got != "mail engine payload too large" {
		t.Fatalf("payload error = %q", got)
	}
}
```

- [ ] **Step 8: Record the expected RED state without running on the host**

Expected compile failures before Task 2 include missing `EngineProjectionPage`, bounded
request types/constants/enums and `ErrEnginePayloadTooLarge`. Because the accepted launcher
is `NOT_READY`, record:

```text
TDD_RED=STATIC_EXPECTED
RUNTIME_RED=NOT_RUN
reason=task-owned Docker launcher NOT_READY; host/direct-Docker fallback forbidden
```

Do not weaken or remove assertions to make the historical port compile.

---

### Task 2: Implement the minimal closed 0.1.8 port

**Files:**

- Modify: `backend/internal/mail/mailbox_engine.go`
- Modify: `backend/internal/mail/mailbox_types.go`
- Modify: `backend/internal/mail/mailbox_errors.go`
- Test: `backend/tests/mail/mailbox_engine_contract_test.go`

**Interfaces:**

- Consumes: the failing external oracle from Task 1.
- Produces: the exact public types/signatures/constants/error kind in
  `docs/mail/ENGINE-CONTRACT-0.1.8.md`; no fake, validator, adapter or persistence behavior.

- [ ] **Step 1: Change only the seven interface signatures**

Use these exact replacements in `MailEngine`:

```go
ListMessages(context.Context, EngineMessageQuery) (EngineProjectionPage, error)
GetMessage(context.Context, EngineMessageReadRequest) (EngineMessage, error)
GetBlob(context.Context, EngineBlobReadRequest) (EngineBlob, error)
Changes(context.Context, EngineChangeRequest) (EngineChangePage, error)
EnumerateProjection(context.Context, EngineEnumerationRequest) (EngineEnumerationPage, error)
SaveDraft(context.Context, EngineDraft) (EngineMessage, error)
ReconcileDraft(context.Context, EngineDraftReconciliation) (EngineDraftResult, error)
```

Keep `Submit(context.Context, EngineSubmission)` and every other method name/order stable.

- [ ] **Step 2: Add the exact integer constants**

Add the 0.1.8 constants with the values asserted in Task 1. Preserve
`MaxEngineSpamTrainingItems=200` and `MaxEngineDNSRecordItems=32`; they are unchanged and
must not be removed.

- [ ] **Step 3: Replace the page/read request types**

Remove `EnginePageRequest.Cursor`. Add `Cursor` to `EngineMessageQuery`, add
`EngineMessageReadRequest`, `EngineBlobReadRequest`, `EngineChangeRequest`, and
`EngineEnumerationRequest`, replace `EngineMessagePage` with `EngineProjectionPage`, and
change the message element type in change/enumeration pages to `EngineMessageProjection`.
Use the exact field names and order asserted in Task 1.

- [ ] **Step 4: Add the exact projection and attachment values**

Import `time` alongside no other new package. Define the projection/attachment/message body
types and these enum names/values:

```go
type EngineBodyFormat string
const (
	EngineBodyPlain EngineBodyFormat = "plain"
	EngineBodyRich  EngineBodyFormat = "rich"
)

type EngineMessagePriority string
const (
	EnginePriorityLow    EngineMessagePriority = "low"
	EnginePriorityNormal EngineMessagePriority = "normal"
	EnginePriorityHigh   EngineMessagePriority = "high"
)

type EngineAttachmentDisposition string
const (
	EngineAttachmentDispositionAttachment EngineAttachmentDisposition = "attachment"
	EngineAttachmentDispositionInline     EngineAttachmentDisposition = "inline"
)
```

Retain all existing flag and keyword types. Change only the draft origin string from its
existing value to:

```go
EngineProjectionOriginConfirmedOwnerDraft EngineProjectionOrigin = "confirmed_owner_draft"
```

- [ ] **Step 5: Replace content, draft and submission values**

Define the exact fields and order asserted by Task 1. `EngineMessage` becomes
`Projection + Body`; `EngineMessageContent` gains reply-to/references/format/priority/
attachments; draft/reconciliation gain `DraftKey` and `MaxContentBytes`; submission becomes
the six-field reference-only value. Do not add compatibility aliases for the historical
hydrated types: callers must fail at compile time until migrated.

- [ ] **Step 6: Add the safe payload error**

In `mailbox_errors.go` add:

```go
ErrEnginePayloadTooLarge = errors.New("mail engine payload too large")
```

and:

```go
EngineErrorPayloadTooLarge EngineErrorKind = "payload_too_large"
```

Add the kind to both allowlist switches:

```go
case EngineErrorPayloadTooLarge:
	return "mail engine payload too large"
```

```go
case EngineErrorPayloadTooLarge:
	return ErrEnginePayloadTooLarge
```

Unknown kinds must still return `mail engine error` and unwrap to nil.

- [ ] **Step 7: Perform static GREEN inspection**

Without executing Go tooling, prove:

```powershell
rg -n "EngineMessagePage|EnginePageRequest\{[^}]*Cursor|EngineSubmission\{[^}]*Message" backend/internal/mail backend/tests/mail
rg -n "EngineProjectionPage|EngineMessageReadRequest|EngineBlobReadRequest|EngineErrorPayloadTooLarge|confirmed_owner_draft" backend/internal/mail backend/tests/mail
git diff --check -- backend/internal/mail/mailbox_engine.go backend/internal/mail/mailbox_types.go backend/internal/mail/mailbox_errors.go backend/tests/mail/mailbox_engine_contract_test.go
```

Expected: the first search returns no live historical declarations/usages; the second shows
interface, type/error definitions and external assertions; `diff --check` exits zero.

- [ ] **Step 8: Preserve runtime honesty**

If the accepted task-owned launcher is still absent, record exactly:

```text
compile=NOT_RUN
unit=NOT_RUN
race=NOT_RUN
reason=task-owned Docker launcher NOT_READY; host/direct-Docker fallback forbidden
```

Do not run `go test`, `go test -race`, `go vet`, a direct container, or Compose.

- [ ] **Step 9: Return the uncommitted worker handoff**

Return exact requested model/effort, provenance source, actual telemetry when exposed, starting
SHA, four changed paths, exact static searches, `git diff --check`, and every `NOT_RUN`. Do not
commit, push, open a PR or touch docs.

---

### Task 3: Root review, docs sync and atomic commit

**Files:**

- Review only: the four Task 2 paths.
- Modify: `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-01A.md`
- Modify: `.planning/phases/144-mailbox-domain/REVIEW-144-01A-ENGINE-BOUNDS.md`
- Modify: `.planning/STATE.md`

**Interfaces:**

- Consumes: the Terra/xhigh worker handoff and exact source diff.
- Produces: one atomic implementation commit whose evidence distinguishes static source
  conformance from executable readiness.

- [ ] **Step 1: Review exact ownership and field grammar**

Reject the handoff if any changed path is outside the four allowed code/test paths, any
historical hydrated page remains, any projection contains body/bytes, submission contains
  content, an enum/constant differs, or the route record has a rejection, omitted
  model/effort, fallback or telemetry mismatch.

- [ ] **Step 2: Review non-tautological test coverage**

Confirm the test compiles as an external consumer, asserts exact field order/count, asserts
absence of content fields, checks literal limit/enum values independently, and exercises
the new sentinel through `errors.Is`. An assertion that only reads back a value it just set
does not count.

- [ ] **Step 3: Sync evidence without a PASS overclaim**

Record source status as `STATIC_CONFORMANT` only if Steps 1–2 pass. Keep runtime, compile,
unit and race as `NOT_RUN` while the launcher is unavailable. Keep 144-01B and Phase 145
blocked until independent source review and applicable Docker evidence close their gates.

- [ ] **Step 4: Stage only the seven owned code/evidence paths**

Stage the four Task 2 paths plus the three docs above, inspect `git diff --cached --name-only`,
and run `git diff --cached --check`. No unrelated dirty or protected path may be staged.

- [ ] **Step 5: Create the atomic commit**

```text
feat(mail): close bounded engine port
```

- [ ] **Step 6: Verify the commit boundary**

Confirm the worktree is clean, record the exact commit SHA and `git show --stat`, and retain
all executable statuses as `NOT_RUN` unless a supported current-SHA launcher result exists.

---

## Plan Self-Review

- Spec coverage: every 0.1.8 interface, request/page/projection/body/blob/draft/submission
  representation, hard constant and safe error owned by 144-01A maps to Task 1 or Task 2.
- Deliberate later ownership: value/state/cardinality validation belongs to 144-01B fake
  conformance and Phase 145 adapter conformance; no validator or provider behavior is smuggled
  into this compile-contract packet.
- Type consistency: all test names, fields, signatures and enum constants match
  `ENGINE-CONTRACT-0.1.8.md` at `c0f20498` plus the docs-only plan amendment.
- Placeholder scan: no implementation placeholder or guessed launcher command exists.
- Current execution state: `BLOCKED_INDEPENDENT_CONTRACT_REVIEW_PENDING` until an independent
  Terra/high contract review accepts the exact SHA. Once that separate gate passes, an exact
  Terra/xhigh worker may use dispatcher-enforced provenance; runtime remains `NOT_RUN`.
