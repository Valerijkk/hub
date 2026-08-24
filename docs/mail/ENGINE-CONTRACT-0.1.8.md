# Mail engine boundary 0.1.8
> **Draft-request canonicalization V1 source seam (2026-08-24):** `CanonicalizeMailboxDraftRequest` is the one pure pre-dispatch boundary for a complete draft request. It validates nonzero account/client/identity and optional message UUIDs; normalizes the sender and each ordered address set; caps each set at 100 and the combined To/CC/BCC recipients at 300; rejects canonical recipient duplicates across those three sets; retains ReplyTo as its own unique ordered set; permits at most 100 nonempty valid thread references and at most 50 unique content-addressed attachment IDs. Header values are UTF-8, at most 998 bytes, and reject C0/C1 controls; `low|normal|high` is the complete priority set.
>
> Its body must already be the exact V1 `MailboxContentResult` with remote images blocked. The seam re-runs content normalization from canonical plain or sanitized rich content and requires complete equality; an empty canonical rich result remains valid. It deep-copies every returned slice, pointer, address, header, attachment, and body result, then computes `sha256` over the ordered TLV V1 frame with literal domain `hub-mail-draft-request-hash-v1`: identity tuple, normalized addresses, headers/threading, priority, optional-message presence values, ordered attachment UUID+digest records, body source-format/sanitizer-version, and SHA-256 digests of canonical plain and sanitized rich strings. The frame contains no raw body bytes.
>
> Later `DraftService` dispatch must consume only this returned canonical request, never a separately normalized hash input with raw dispatch values, and must server-generate mutation IDs after canonicalization. This source seam adds no runtime wiring, provider call, repository body persistence, API/UI, migration, or release authority. The repository remains content-free; stale content tests remain deferred and build, tests, lint, gofmt, Docker, and runtime are `NOT_RUN`.
>
> **144-05E rich-content static source (2026-08-24):** `backend/internal/mail/mailbox_content.go`, `mailbox_message_service.go`, and `mailbox_message_repository.go` now implement source-only deterministic receive sanitization. `MailboxContentResult` reports the selected `plain|rich` source format, sanitizer version V1, always-derived canonical plain text, rich-only bounded `SanitizedHTML`, and `RemoteImagesBlocked=true`. The V1 allowlist is exactly `p`, `br`, `hr`, `strong`, `b`, `em`, `i`, `u`, `s`, `del`, `code`, `pre`, `blockquote`, `ul`, `ol`, `li`, `h1` through `h6`, `div`, and `span`; all attributes are stripped, `br`/`hr` serialize as void tags, and nonvoid self-closing input serializes as an open/close pair. Active/foreign roots and descendants, comments, declarations and processing instructions emit nothing; `a` and other safe unknown tags unwrap to retained escaped text. Rich reads expose a copied `sanitized_html` pointer, including a nonnil pointer for an empty safe result; plain reads expose none. No raw attribute URL, `href`, `src`, `style`, event handler, class, id, data, or cid reaches output. The 8,388,608-byte sanitized-HTML ceiling is independent of input/plain limits and includes escaping. This is source implemented static-only: test fixtures/oracles, build, tests, lint, gofmt, Docker, and runtime remain `DEFERRED` / `NOT_RUN`; it creates no remote-image retrieval, proxy, policy, persistence, API, UI, or release authority.
> **Current narrow deletion-source authority (2026-08-23; highest precedence):** accepted plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` is `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` only for one future atomic no-test exact2 production packet: `backend/internal/mail/mailbox_deletion.go` and `backend/internal/mail/mailbox_deletion_fake.go`. Plan Contract/high, Security/xhigh and Lifecycle/high reviews are C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own commit and immutable C/S/L triple `0/0/0`. Adapter capability proof is not required only for this provider-neutral ABI plus two standalone fakes exception, but remains mandatory for every concrete provider adapter. Persistence/Ent/schema/generation/migration/SQL/worker/storage drain/account finalization/User-GDPR/HTTP/NATS/Docker/runtime stay `BLOCKED` / `NOT_RUN`. Old blanket deletion-source-blocked or plan-pending claims are `CONSUMED/HISTORICAL` only within this exact2 scope; unrelated gates are unchanged.
> **Engine deletion-port adjudication (2026-08-23):** `STATIC_ACCEPTED / DOCS_DESIGN_ONLY`; all former deletion-port-next clauses are `CONSUMED/HISTORICAL`, never source authority. Freeze separate future Go ports without widening normal `MailEngine` or combining account removal with object drain: `type EngineAccountDeletionPort interface { DeleteAccount(ctx context.Context, request EngineAccountDeleteRequest) (EngineAccountDeleteResult, error); ReconcileAccountAbsence(ctx context.Context, request EngineAccountAbsenceRequest) (EngineAccountAbsenceResult, error) }` and `type MailObjectDeletionPort interface { DeleteObject(ctx context.Context, request MailObjectDeleteRequest) (MailObjectDeleteResult, error); ReconcileObjectAbsence(ctx context.Context, request MailObjectAbsenceRequest) (MailObjectAbsenceResult, error) }`. `EngineAccountDeleteRequest` has exactly exported `ProvisioningKey string`, `EngineAccountID string`, and `DeletionID uuid.UUID`; `MailObjectDeleteRequest` has exactly exported `Bucket string`, `ObjectKey string`, and `DeletionID uuid.UUID`. The aliases are explicit: `type EngineAccountAbsenceRequest = EngineAccountDeleteRequest` and `type MailObjectAbsenceRequest = MailObjectDeleteRequest`, so an absence request carries the identical corresponding scope and nonzero `DeletionID`. Account and object delete results are distinct structs with the closed state `deleted|authoritatively_absent`; account and object absence results are distinct structs with the closed observation `present|authoritatively_absent`. Account failures use the existing `*EngineError`, constrained here to `not_found|conflict|unavailable|indeterminate|rejected_policy|unauthorized_account`; bare `not_found` never finalizes intent. Object failures use the dedicated content-free `MailObjectDeletionErrorKind` / `MailObjectDeletionError` surface only: exact kinds `unavailable|indeterminate|conflict|rejected_policy|unauthorized_scope`, `NewMailObjectDeletionError`, `Kind`, `KindOf`, and an `Error()` method with a constant result; it exposes no `Unwrap`, cause, detail, provider/raw identifier, or loggable content. `ProvisioningKey` and `EngineAccountID` are UTF-8 1..`MaxEngineOpaqueIDBytes` bytes; `Bucket` is UTF-8 1..63 bytes; `ObjectKey` is UTF-8 1..512 bytes; every `DeletionID` is nonzero. Private intent/control is not owner projection; short transactions keep I/O outside locks, reclaim only expired leases, and every retry/finalization uses exact token+generation+captured-scope CAS so stale token or zero rows perform no I/O/compensation. Object drain starts only after `upload_drain_until` and attachment/reference terminality; account removal is last after the exact child/object/receipt/control/activity manifest; hard delete follows only closed evidence. No provider, Stalwart route, OAuth, DNS, OpenAPI, source, schema, migration, SQL, test, Docker, runtime, or physical User/GDPR authority follows; future source remains separately planned and `BLOCKED` / `NOT_RUN`.

```go
type EngineAccountDeletionPort interface {
	DeleteAccount(ctx context.Context, request EngineAccountDeleteRequest) (EngineAccountDeleteResult, error)
	ReconcileAccountAbsence(ctx context.Context, request EngineAccountAbsenceRequest) (EngineAccountAbsenceResult, error)
}

type MailObjectDeletionPort interface {
	DeleteObject(ctx context.Context, request MailObjectDeleteRequest) (MailObjectDeleteResult, error)
	ReconcileObjectAbsence(ctx context.Context, request MailObjectAbsenceRequest) (MailObjectAbsenceResult, error)
}

type EngineAccountDeleteRequest struct {
	ProvisioningKey string
	EngineAccountID string
	DeletionID      uuid.UUID
}

type MailObjectDeleteRequest struct {
	Bucket     string
	ObjectKey  string
	DeletionID uuid.UUID
}

type EngineAccountAbsenceRequest = EngineAccountDeleteRequest
type MailObjectAbsenceRequest = MailObjectDeleteRequest

type EngineAccountDeletionState string

const (
	EngineAccountDeletionStateDeleted               EngineAccountDeletionState = "deleted"
	EngineAccountDeletionStateAuthoritativelyAbsent EngineAccountDeletionState = "authoritatively_absent"
)

type EngineAccountAbsenceState string

const (
	EngineAccountAbsenceStatePresent                EngineAccountAbsenceState = "present"
	EngineAccountAbsenceStateAuthoritativelyAbsent  EngineAccountAbsenceState = "authoritatively_absent"
)

type MailObjectDeletionState string

const (
	MailObjectDeletionStateDeleted                  MailObjectDeletionState = "deleted"
	MailObjectDeletionStateAuthoritativelyAbsent    MailObjectDeletionState = "authoritatively_absent"
)

type MailObjectAbsenceState string

const (
	MailObjectAbsenceStatePresent                   MailObjectAbsenceState = "present"
	MailObjectAbsenceStateAuthoritativelyAbsent     MailObjectAbsenceState = "authoritatively_absent"
)

type EngineAccountDeleteResult struct {
	State EngineAccountDeletionState
}

type EngineAccountAbsenceResult struct {
	State EngineAccountAbsenceState
}

type MailObjectDeleteResult struct {
	State MailObjectDeletionState
}

type MailObjectAbsenceResult struct {
	State MailObjectAbsenceState
}

type MailObjectDeletionErrorKind string

const (
	MailObjectDeletionErrorUnavailable       MailObjectDeletionErrorKind = "unavailable"
	MailObjectDeletionErrorIndeterminate     MailObjectDeletionErrorKind = "indeterminate"
	MailObjectDeletionErrorConflict          MailObjectDeletionErrorKind = "conflict"
	MailObjectDeletionErrorRejectedPolicy    MailObjectDeletionErrorKind = "rejected_policy"
	MailObjectDeletionErrorUnauthorizedScope MailObjectDeletionErrorKind = "unauthorized_scope"
)

type MailObjectDeletionError struct {
	kind MailObjectDeletionErrorKind
}

func NewMailObjectDeletionError(kind MailObjectDeletionErrorKind) (*MailObjectDeletionError, error) {
	switch kind {
	case MailObjectDeletionErrorUnavailable, MailObjectDeletionErrorIndeterminate, MailObjectDeletionErrorConflict, MailObjectDeletionErrorRejectedPolicy, MailObjectDeletionErrorUnauthorizedScope:
		return &MailObjectDeletionError{kind: kind}, nil
	default:
		return nil, errors.New("invalid mail object deletion error kind")
	}
}

func (*MailObjectDeletionError) Error() string { return "mail object deletion failed" }
func (e *MailObjectDeletionError) Kind() MailObjectDeletionErrorKind {
	if e == nil {
		return ""
	}
	switch e.kind {
	case MailObjectDeletionErrorUnavailable, MailObjectDeletionErrorIndeterminate, MailObjectDeletionErrorConflict, MailObjectDeletionErrorRejectedPolicy, MailObjectDeletionErrorUnauthorizedScope:
		return e.kind
	default:
		return ""
	}
}
func MailObjectDeletionErrorKindOf(err error) (MailObjectDeletionErrorKind, bool) {
	var target *MailObjectDeletionError
	if errors.As(err, &target) {
		kind := target.Kind()
		if kind != "" {
			return kind, true
		}
	}
	return "", false
}
```

Zero/unknown result or observation states are invalid. The adapter boundary rejects them and preserves durable intent as indeterminate; it never finalizes. `MailObjectDeletionError.Kind` is nil-safe and validates the closed kind set, returning `""` for a nil receiver or zero/unknown stored value. `MailObjectDeletionErrorKindOf` uses `errors.As` but returns `("", false)` when its target is typed-nil or has a zero/unknown kind; an invalid constructor kind returns a generic safe error. `MailObjectDeletionError` exposes no `Unwrap`, cause, or detail.
> **P3-D0 static acceptance (2026-08-23):** immutable source `47faf3a705e5641f4ceb7ee365c3dfeeeace725d` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/ent/schema/calendarinboundmessagereceipt.go`. It adds required immutable `organization_id`/`mailbox_id` UUID scalars and database nonzero checks, with no Ent edges or raw data; existing uniqueness, state-shape, and due behavior are unchanged. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. The D0-next route is `CONSUMED/HISTORICAL`. P3-C remains blocked on Mail Ent generation; 147-02B/147-02C and D1/D2 remain blocked on active Calendar generation/dependencies; P3-E is conjunction-blocked. No next unblocked P3 source is selected. Generation, migration, SQL, build, tests, Docker, and runtime are `NOT_RUN`; no readiness, merge, or product claim.

> **Phase-147 P3-B static acceptance (2026-08-23):** `43a6c42af88524be2951966bdbfc83b422ecb7f9` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/internal/calendar/event_repository.go`, `backend/internal/calendar/event_service_inbound_reply.go`, and `backend/internal/calendar/event_service_inbound_message.go`. It accepts the Calendar transaction ABI only: receipt-first single `EventTransaction`, opaque claim/scope/completion, and closed observed/error-code constructor ABI, preserving REPLY receipt/outbox behavior without nested transactions or duplicate logic. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. CONSUMED/HISTORICAL: P3-C was the sole next pending atom, blocked on Mail Ent generation, for exactly `backend/internal/mail/mailbox_ent_calendar_part.go` and `backend/internal/mail/mailbox_ent_repository.go`; P3-D/E remain pending/blocked. Generation, migration, SQL, build, tests, lint/gofmt, parser/validator, Docker, and runtime are `NOT_RUN`; no merge, product, or Docker-readiness claim follows. Earlier P3-B-next clauses are `CONSUMED/HISTORICAL`.

> **144-02A/B/C/D status sync (2026-08-15):** receipt `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for the separate exact 14 content documents after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Engine deletion-port docs adjudication is accepted by the docs chain including this correction, following `16ff90b0` and `950b22ca`; the former nearest-work selection is `CONSUMED/HISTORICAL`, never source authority. Deletion source/adapter/persistence/worker and User deletion intent remain `BLOCKED` / `NOT_DISPATCHABLE` / `NOT_RUN`; launcher `NOT_READY`, executables/runtime `NOT_RUN`.

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

> **144-01B static acceptance (2026-08-15):** immutable `0c064363b5ccb657a7c444802e94075c4ca743e5` (parent `48a7273d100c07788def88e2f1b8d69c7173580f`, tree `9c568a0c58c6bf7b8c3171cfd934760b839db623`, subject `feat(mail): add deterministic engine fake`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for fake `12bcd921668e406f07a98f7f3b0618ad8d44e81d` and authored-first test `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f`. Terra/xhigh author; Spec Terra/high, Security Terra/xhigh, ABI Terra/high precommit and exact-SHA Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 01B Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executables `NOT_RUN`; no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority.

> **144-02A2 static source receipt (2026-08-15):** `65c91b4824e8a1948fc97c149305aa7f009de061` (parent `5df6b47678364c438ade579252209c8fd238c43c`, tree `08583554939f597443566093f30b6ec7aa9373db`) accepts blobs `4de2ec0882d94db595aaabf8d4117566118ee472`, `a70e05122ffa126e9754d029ca00493faae2aeef`, `ef6e2e65430a1397b0f563f1f361ecd89660a39c`, authored-first oracle `65de04719a60a38d7fc456553e092bb106456957` as `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Terra/xhigh author, Terra/high precommit+exact-SHA PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. No engine/generation/migration/runtime/service/API/UI/product/phase/release authority; launcher `NOT_READY`, executables `NOT_RUN`; A3/02E blockers unchanged.

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

### P3 Calendar-part private read amendment

For the future P3 Mail-to-Calendar source only, the owner is `backend/internal/mail` (`package mail`):
Calendar may import its exported sealed port/types but Mail never imports Calendar. Add
`MailEngine.ReadCalendarPart` rather than widening Calendar's view of `GetMessage`. P3-A is
`STATIC_ACCEPTED / STATIC_SOURCE_ONLY` at `05f87f87f6dc30bd57648030026faf5912add3e7`
(`feat(mail): add sealed calendar part port`) for exactly `calendar_part_source.go`, `mailbox_engine.go`,
`mailbox_types.go`, `mailbox_errors.go`, and `mailbox_fake.go`; independent Contract/high and
Security/xhigh reviews are APPROVED C/I/M `0/0/0`. That acceptance covers the sealed ABI/port only, not
a concrete source, engine/provider I/O, MIME/locator handling, Ent, NATS, wiring, or runtime. The P3-A
static accepted compile shape in `package mail` is:

```go
type CalendarPartReadRequest struct { /* unexported fields */ }
func NewCalendarPartReadRequest(organizationID, mailboxID, messageID uuid.UUID, locatorHash [sha256.Size]byte, maxBytes int) (CalendarPartReadRequest, error)
type CalendarPartSource interface { ReadCalendarPart(context.Context, CalendarPartReadRequest) (CalendarPartReadResult, error) }
type CalendarPartReadResult struct { /* unexported fields */ }
func (r CalendarPartReadResult) OwnerID() uuid.UUID
func (r CalendarPartReadResult) Sender() string
func (r CalendarPartReadResult) Method() MailCalendarMethod
func (r CalendarPartReadResult) AuthenticationClass() MailAuthenticationClass
func (r CalendarPartReadResult) ITIP() []byte

type EngineCalendarPartReadRequest struct { AccountID, MessageID, Version string; LocatorHash [sha256.Size]byte; Method MailCalendarMethod; MaxBytes int }
type EngineCalendarPartReadResult struct { LocatorHash [sha256.Size]byte; ITIP []byte }
// method added to MailEngine:
ReadCalendarPart(context.Context, EngineCalendarPartReadRequest) (EngineCalendarPartReadResult, error)

type CalendarInboundDelivery struct { /* unexported fields */ }
func (d CalendarInboundDelivery) EnvelopeID() uuid.UUID
func (d CalendarInboundDelivery) OrganizationID() uuid.UUID
func (d CalendarInboundDelivery) MailboxID() uuid.UUID
func (d CalendarInboundDelivery) MessageID() uuid.UUID
func (d CalendarInboundDelivery) Sender() string
func (d CalendarInboundDelivery) Method() MailCalendarMethod
func (d CalendarInboundDelivery) AuthenticationClass() MailAuthenticationClass
func (d CalendarInboundDelivery) LocatorHash() [32]byte
type CalendarInboundSink interface { HandleCalendarInbound(context.Context, CalendarInboundDelivery) error }
```

`CalendarPartReadRequest`/Result remain opaque; only the stated constructor/result accessors are
Calendar-facing. The closed `CalendarPartSourceErrorKind` names/values are exactly
`invalid|untrusted|message_unavailable|deleting|part_unavailable|payload_too_large|
repository_unavailable|engine_unavailable`; no other kind exists. Its exact future ABI in `package mail`
is:

```go
type CalendarPartSourceErrorKind string
// exact constants already listed
type CalendarPartSourceError struct { /* unexported kind only */ }
func (e *CalendarPartSourceError) Error() string // constant content-free text
func (e *CalendarPartSourceError) Kind() CalendarPartSourceErrorKind
func CalendarPartSourceErrorKindOf(error) (CalendarPartSourceErrorKind, bool)
```

P3-A accepts this ABI only. A concrete P3-C `CalendarPartSource.ReadCalendarPart` remains pending and must
return only this family; its constructor is unexported to `package mail`, and it maps engine/provider
errors inside the source. `KindOf` uses `errors.As`.
The error has no `Unwrap`, retains no provider/SQL/body/cause value, and returns the same constant
content-free `Error()` text for every kind. No engine/provider ID, raw locator, raw MIME, `MailEngine`,
or Mail Ent value crosses the source boundary.

P3-A accepts the `MailEngine.ReadCalendarPart` signature only; it is exported for Go interface
implementation but engine-private: only a future Mail source may call it, never Calendar/server. P3-C is
the pending concrete source atom after Mail generation. It owns private account/message IDs, expected
`Version`/hash/method/max bytes, one copied iTIP result, exact-one canonical private MIME locator match,
and `maxBytes=1..1,048,576` with a five-second full-operation cap and whole-result
`payload_too_large`. V1 locator hash is exactly
`SHA-256("hub.mail.calendar-part-locator" || 0x00 || 0x01 || u32be(len(engine_account_id)) || engine_account_id || u32be(len(engine_message_id)) || engine_message_id || u32be(len(canonical_private_part_locator)) || canonical_private_part_locator || u32be(32) || SHA-256(decoded_itip_bytes))`.
`Version` is excluded from hash identity. P3-C must recompute from returned exact decoded iTIP bytes and
private canonical locator and compare to requested hash; body substitution is an integrity failure.
Changing the formula requires a versioned schema/event amendment. Prior 147-02D
Calendar-ingestion use of `GetMessage`, full-body/MIME parsing, or legacy `engine_message_id` is
`CONSUMED/HISTORICAL`; those are forbidden Calendar fallbacks, not a removal of unrelated Mail reads.
P3-C snapshot/recheck requires the same private `engine_version` before/after engine I/O for TOCTOU, but
an engine-version-only flag/folder change neither changes the hash nor loses the existing delivery.
Provider/projection must recompute the hash only when canonical locator or decoded iTIP bytes change.
Mail content is immutable: a P3-C mismatch maps to `invalid`, durable `quarantined/internal_error`, and a
content-free integrity-degraded signal, never processing/relabeling an old event.

P3-B transaction port below is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` at
`43a6c42af88524be2951966bdbfc83b422ecb7f9`; CONSUMED/HISTORICAL: P3-C was the sole next
`UNIMPLEMENTED / PENDING / BLOCKED` atom on Mail Ent generation, and P3-D/E remain pending. Its accepted
`package calendar` transaction port is:

```go
type CalendarInboundMessageClaim struct { ReceiptID, LeaseToken uuid.UUID; LeaseGeneration int64 }
type CalendarInboundMessageScope struct { /* opaque transaction-bound */ }
type CalendarInboundObservedState string // only ignored_stale|duplicate
type CalendarInboundMessageErrorCode string // exact Phase147 allowlist
type CalendarInboundMessageCompletion struct { /* opaque */ }
func NewCalendarInboundAppliedCompletion(calendarEventID uuid.UUID) (CalendarInboundMessageCompletion, error)
func NewCalendarInboundObservedCompletion(state CalendarInboundObservedState, calendarEventID *uuid.UUID) (CalendarInboundMessageCompletion, error)
func NewCalendarInboundQuarantinedCompletion(code CalendarInboundMessageErrorCode) (CalendarInboundMessageCompletion, error)
func NewCalendarInboundDeadLetterCompletion(code CalendarInboundMessageErrorCode) (CalendarInboundMessageCompletion, error)
func NewCalendarInboundRetryCompletion(code CalendarInboundMessageErrorCode) (CalendarInboundMessageCompletion, error)
func NewCalendarInboundDeferredCompletion(calendarEventID *uuid.UUID) (CalendarInboundMessageCompletion, error)
// added to EventTransaction:
LockCalendarInboundMessageReceipt(context.Context, CalendarInboundMessageClaim) (CalendarInboundMessageScope, bool, error)
CompleteCalendarInboundMessageReceipt(context.Context, CalendarInboundMessageScope, CalendarInboundMessageCompletion) error
```

Completion is constructor-only, carries no raw MIME, iTIP, provider text, or unvalidated outcome, and
defensively copies optional UUID pointers; every nonnil UUID is nonzero and applied requires a nonzero
event UUID. Observed is only `ignored_stale|duplicate` with nil error. Error code is exactly the Phase 147
receipt allowlist: `message_unavailable|part_unavailable|malformed_itip|unsupported_method|
unsupported_version|untrusted_sender|organizer_mismatch|equal_sequence_conflict|missing_master_expired|
resource_limit|retry_exhausted|engine_unavailable|database_unavailable|internal_error`.
Quarantine allows only `malformed_itip|unsupported_method|unsupported_version|untrusted_sender|
organizer_mismatch|equal_sequence_conflict|internal_error`; dead letter only
`message_unavailable|part_unavailable|missing_master_expired|resource_limit|retry_exhausted|
internal_error`; retry only `engine_unavailable|database_unavailable|internal_error`; deferred has nil
error. Constructors take no timestamp, lease token, generation, or raw content. The repository derives due
time from receipt attempts and its DB clock; `CompleteCalendarInboundMessageReceipt` alone clears/sets
lease and due by validated kind. `bool=false, nil` (that is, `valid=false, err=nil`) is a lease-lost/
terminal no-op with no Calendar mutation.
`ApplyInboundITIPReply` stays a thin `WithEventTransaction` wrapper around
`applyInboundITIPReplyInTransaction(ctx, tx, input)`.

P3-E remains pending after concrete engine/source. Only its future Mail strict v1 decoder constructs opaque `CalendarInboundDelivery`; there is no public
arbitrary-field constructor. The Mail-owned decoder validates the full compatibility envelope, drops
`engine_message_id`, and constructs the delivery without it before invoking the injected
Calendar-implemented `CalendarInboundSink`; Calendar never receives that field or provider data. A nil
sink result means the durable Calendar disposition committed and is ACKed. A content-free/provider-free
sink error means no durable disposition and uses `NakWithDelay(65s)`/MaxDeliver20; no Calendar package
subscribes to v1.
P3-B is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` at `43a6c42af88524be2951966bdbfc83b422ecb7f9`; CONSUMED/HISTORICAL: P3-C was the sole next
`UNIMPLEMENTED / PENDING / BLOCKED` atom on Mail Ent generation, P3-D/E remain pending, and the concrete Stalwart adapter remains
blocked on Phase 145 authority. Generated Ent, migration, SQL, build, tests, lint/gofmt, parser/validator,
Docker, and runtime remain `NOT_RUN`.

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
    ConfirmedMutationID string
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

### Two-phase draft save representation

For a save, persistence owns the two durable graphs: confirmed O and reserved desired D. The
repository writes D only through `Reserve` under its account/draft/attachment lock order, then
commits before calling `SaveDraft`; no provider I/O, adapter callback, or reconciliation read is
permitted while that transaction or lock set is held. The engine receives the reserved requested
value but never owns or mutates durable O/D links.

The repository source statically persists `confirmed_mutation_id` as the binding for the
confirmed pair and exposes it with O/D work. `EngineDraftReconciliation` carries authoritative
`ConfirmedMutationID`; the marker formula remains mutation-only
`draft-mutation-v2(account,draftKey,mutationID)` and is not tuple-bound. The current exact v2
marker returns `Applied` for initial create. For an update it returns `Applied` only when the
result `Projection (MessageID,Version)` is not exactly the old pair; a same-pair current proof is
`Indeterminate` unless the separate exact `ConfirmedMutationID` branch proves `Retained`.
Legacy v1, nil, or generic marker observations are `Indeterminate` only for retained proof; an
exact current v2 marker may be `Applied` with a nil legacy binding after a new operation.
Deployment requires all marker-v1 writers to be quiesced before marker-v2 writes are enabled.
Fresh server-side `MutationID` generation remains a future `DraftService` obligation.

`Applied` permits the repository to promote D to confirmed O only after the exact mutation and
fence are revalidated. The provider-neutral sealed `EngineDraftRetained{Message EngineMessage}`
alternative, the Fake engine's authoritative current-store proof, and the Stalwart marker-v2
proof source are implemented. The exact current marker is `Applied` for initial create and, for
updates, only when its result projection differs from the old pair; a same-pair current proof is
`Indeterminate` unless the separate exact `ConfirmedMutationID` branch proves `Retained`.
Durable repository/service `SaveRetained` remains `SOURCE_NOT_IMPLEMENTED` and reject-only until
a joint service+repository packet. Every
non-`Applied` outcome retains O and D in `unknown`; no engine observation can clear D, finalize a
request, retry, or clean up anything by itself. Pending attachment ownership is live persistence
state; the adapter must not interpret O∩D, O\D, or D\O as duplicate upload/delete work.

`EngineDraftRetained` is a sealed result alternative containing exactly `Message EngineMessage`.
It may be returned only for the same authorized save/reconciliation request whose `AccountID`,
`DraftKey`, `MutationID`, `MessageID`, and `PreviousVersion` match the durably reserved mutation
and confirmed O pair, and whose authoritative `ConfirmedMutationID` binds the exact prior marker.
Its `Message.Projection.AccountID`, `Message.Projection.MessageID`, and
`Message.Projection.Version` must respectively equal that account, old message id, and old
version. The Fake establishes this from its mutex-protected authoritative current draft/message
stores with the existing message/projection validation and bounds. A bare or unsealed old-pair
observation remains insufficient. A sealed live reconcile result plus the exact persisted O pair
and binding can be evidence only in a later separately authorized provider-specific proof packet;
it requires neither reverse mapping of provider attachment ids to Hub UUIDs nor a
`confirmed_request_hash`.

That later repository packet must reacquire the established lock order and prove the exact current
account/draft/mutation/fence, unchanged O pair, binding, and still-reserved D before it clears D;
otherwise it retains O and D as `unknown`. No direct `SaveDraft` success, address, folder
membership, mutable provider metadata, or bare old-pair observation is proof. This narrow source
adds no `MailEngine` method and authorizes neither generated Ent output (`NOT_GENERATED`), a
migration (`NOT_MIGRATED`), nor execution, build, tests, Docker, or runtime (`NOT_RUN`).

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

> **Local 01B clause reconciliation:** older 144-01B Docker-blocked/do-not-start/dependency wording in this Rule split record is `CONSUMED/HISTORICAL` only; no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority follows.

## 144-03B Rule compiler / adapter split

The docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` Rule contract compiles typed rules to provider-neutral IR,
not an engine port or Sieve bytes. Future Phase 145 separately resolves engine folder ID (not Sieve
mailbox name), renders/escapes, checks RFC 9661 and carries rendered digest. Attachment IR is
capability-blocked without an accepted MIME mapping.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan-next text is `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. Accepted Rule plan `669ef761` does not authorize Task 1, which remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`; Task4 is post-future-Task3 source acceptance only. 144-01B remains Docker-blocked, Phase 142/148 and release pending, launcher `NOT_READY`, executables `NOT_RUN`.
> **Current deletion-port source receipt (2026-08-23; highest precedence):** immutable source `e4f81bf3d8b932f5ff2b9a8069d41bfa1d1a0e26` (parent `bd76d8469782e9bb7926b440b6b45495bc865e84`, tree `93bac3c6a5798cb75b0a6647fa938416c5e82ede`) is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/internal/mail/mailbox_deletion.go` blob `001b03c879aab36031feac07f0d3894f76e23abd` and `backend/internal/mail/mailbox_deletion_fake.go` blob `7d43b4dee20ad762d251e3c14e428f2f0ccda497`. `provenance_source: dispatcher-enforced`; Terra/xhigh implementation; Contract/high, Security/xhigh, and Lifecycle/high immutable reviews are C/I/M `0/0/0`. This accepts only the provider-neutral deletion ABI and two standalone deterministic fakes. Tests, build, gofmt, lint, parser, validator, generation, migration, SQL, Docker, and runtime are `NOT_RUN`. Within this exact2 packet, former `UNIMPLEMENTED`/`UNCOMMITTED`/`UNACCEPTED`/exact2-next/source-review-pending clauses are `CONSUMED/HISTORICAL`. Concrete provider adapter capability, persistence/Ent/schema, migrations, workers/drains/finalization/User-GDPR/HTTP/NATS remain `BLOCKED`; no next source is selected. Testframework merge facts are unchanged.
