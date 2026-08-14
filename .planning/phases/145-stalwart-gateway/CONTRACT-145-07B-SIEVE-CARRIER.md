# Phase-145 sealed typed-IR/Sieve-carrier contract

> **Status:** `CONTRACT_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED` / `IMPLEMENTATION_BLOCKED`. This standalone docs-only artifact is based on `ff427ef253b11138446bca1584a058e3881e1961`; it is not a Phase-145 status transition, implementation plan, source/test authorization, runtime claim, or commit receipt. No commit SHA is predeclared.

> **Prerequisite topology:** this non-marker carrier is prerequisite input to 145-07B; 145-07B cannot bootstrap it. Contract acceptance grants design authority only. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until a separately accepted carrier implementation plan, G145-5B, G145-5C, and launcher readiness exist. Runtime evidence remains `NOT_RUN`; launcher remains `NOT_READY`.

## Scope, lifecycle, and ownership

This carrier replaces the marker-only Sieve engine port with a sealed request that carries accepted Phase-144 `CompiledRuleSet` IR and its desired generation/hash. It does not create a renderer, Stalwart adapter, persistence schema, service, worker, migration, transport API, test, deployment, implementation plan, or runtime work.

Phase 144 owns compiler validation, canonical ordering, target semantics, and the deep-copy contract of `CompiledRuleSet`. Phase 145 may privately render and reconcile a validated sealed request. The future caller/service retains **04M** responsibility: it maps account-lock ownership, obtains the complete current account snapshot, persists an attempt before I/O, and applies account-lock, lease, mutation-epoch, desired-tuple, deployment-token, and completion fences. The factory cannot prove any of those facts; the adapter owns no DB lock, ownership proof, or completion/finalization. A zero-row completion drops the outcome and performs no compensating I/O.

## Frozen public Go ABI proposal

The future atomic carrier implementation replaces the old exports atomically, with no compatibility shim:

```go
type EngineRuleSetHash struct { version int32; sha256 [32]byte }
func NewEngineRuleSetHashV1([32]byte) (EngineRuleSetHash, error)
func (h EngineRuleSetHash) Version() int32
func (h EngineRuleSetHash) SHA256() [32]byte

type EngineRuleScriptRequest interface {
	engineRuleScriptRequest()
	AccountID() string
	DesiredGeneration() int64
	DesiredRuleSetHash() EngineRuleSetHash
	CompiledRuleSet() CompiledRuleSet
}
func NewEngineRuleScriptRequest(engineAccountID string, desiredGeneration int64, expectedHash EngineRuleSetHash, live RuleCompilerInput) (EngineRuleScriptRequest, error)
```

Only concrete `*engineRuleScriptRequest` implements this interface. Its fields are private: engine account ID, generation, v1 hash, and an IR deep copy. Its marker is unexported. `MailEngine` accepts this sealed interface in both methods:

```go
type EngineRuleScriptStateKind string
const (
	EngineRuleScriptStateAbsent EngineRuleScriptStateKind = "absent"
	EngineRuleScriptStateApplied EngineRuleScriptStateKind = "applied"
	EngineRuleScriptStateDrifted EngineRuleScriptStateKind = "drifted"
	EngineRuleScriptStateIndeterminate EngineRuleScriptStateKind = "indeterminate"
)
type EngineRuleScriptState struct { State EngineRuleScriptStateKind }
GetRuleScript(context.Context, EngineRuleScriptRequest) (EngineRuleScriptState, error)
UpsertRuleScript(context.Context, EngineRuleScriptRequest) (EngineRuleScriptState, error)
```

`CompiledRuleSet()` returns a new recursive deep copy on every call. Hash accessors return value copies. A nil interface, typed nil concrete request, zero request, or otherwise forged/unrecognized request is rejected before I/O. No adapter boundary may accept separate raw hash and IR fields.

`NewEngineRuleSetHashV1` creates only version `1` and rejects an all-zero sum; the zero value is invalid. The request factory is pure and exactly ordered:

1. Deep-copy `live`.
2. Validate the opaque engine account ID (`1..MaxEngineOpaqueIDBytes`, 1024), positive generation, and `expectedHash.Version() == 1`.
3. Run `CompileRuleSet` on that same copy.
4. Recompute the exact persistence ruleset hash v1 from that same copy's `live.Rules`; compare it with `expectedHash`.
5. Store only private engine ID, generation, v1 hash, and another deep copy of the compiled IR.

A hash mismatch returns a zero request and the existing `EngineErrorConflict`. Compiler failure or malformed compiler input returns a zero request and existing `EngineErrorRejectedPolicy`. It must not wrap or expose `RuleCompilerError`, address sentinels, or raw compiler detail.

The old exports are removed: `SieveScript`, `EngineSieveState`, `EngineSieveStateKind`, `EngineSieveAbsent`, `EngineSieveApplied`, `EngineSieveIndeterminate`, `GetSieve`, and `UpsertSieve`. There is no alias, deprecated wrapper, conversion helper, or dual method set. The public request/state exposes no script name/ID, bytes/body/blob, provider response, capability URL, marker, rendered digest, credential, mailbox name, or provider-specific field.

## State and error semantics

`GetRuleScript` privately derives expected bytes only after its preconditions. It returns `absent` only when no managed or foreign identity conflict exists. It returns `drifted` for a definitive expected-byte mismatch, ambiguity, duplicate managed identity, **any foreign script at the fixed name/provider identity (active or inactive)**, or **any foreign active script**. All drift returns have zero destructive mutation. `indeterminate` is a lost/unprovable outcome and mandates Get-before-retry.

`UpsertRuleScript` may return only `applied`, `drifted`, or `indeterminate` on a nil error: `applied` requires exact post-write private refetch proof; `drifted` is a definitive mismatch with no mutation; `indeterminate` may-have-applied and forbids a blind retry. Repair is permitted only for exactly one canonical managed identity, with no foreign active script and all 04M fences. Duplicate or ambiguous identities are always `drifted`.

Add exactly one sentinel/kind:

```go
var ErrEngineCapabilityUnavailable = errors.New("mail engine capability unavailable")
const EngineErrorCapabilityUnavailable EngineErrorKind = "capability_unavailable"
```

Extend `EngineError.Error` and `EngineError.Unwrap` with the exact safe mapping. Invalid request and rejected provider-set/script policy map to existing `RejectedPolicy`; oversize request/render/provider response maps to `PayloadTooLarge`; transport, 401, 403, and malformed provider response map to `Unavailable`; stale/same-generation mismatch maps to `Conflict`; account scope failure maps to `UnauthorizedAccount`; unavailable capability maps to `CapabilityUnavailable`. Definitive drift and indeterminate return their state with nil error. Errors always return zero state and contain no adapter-private detail, provider body, capability value, marker, digest, or script bytes.

## Capability, mailbox, and I/O ordering

Before **any** Sieve or blob I/O, the adapter validates the sealed request, gets account context, and proves an exact case-sensitive mailbox-session capability `urn:ietf:params:jmap:sieve`. Missing, malformed, or account-inapplicable capability returns zero state plus `EngineErrorCapabilityUnavailable`, with no upload, set, activation, get, or blob operation. It then gates `fileinto` for move and `imap4flags` for mark-read/label; any attachment predicate blocks the entire rule set as capability unavailable until a separately accepted MIME mapping exists.

Only then it obtains one current account-scoped mailbox snapshot and resolves every move destination by its exact opaque `EngineFolderID`, uniquely. Missing, foreign, ambiguous, stale, or malformed target is `EngineErrorConflict`, zero state, and no upload/set. A malformed provider mailbox response is `Unavailable`. There is no mailbox-name fallback.

After resolution: render privately; Sieve get; upload/set with `onSuccessActivateScript`; refetch script blob; privately compare full rendered digest; then return `applied`. Use only mailbox RFC 9661 `SieveScript/*` surface; `x:SieveSystemScript` and `x:SieveUserScript` are prohibited. No timeout receives a blind retry.

`maxSizeScript` absent means `65536`. When present it must be a lossless, positive, integral JMAP unsigned value within the implementation integer range. Null, fraction, string, overflow, zero, or negative is `CapabilityUnavailable` before I/O. The effective cap is `min(65536, value)`; preflight and refetch reject exactly limit-plus-one.

## Private identity, renderer, and digest freeze

Only the adapter holds script material. Fixed script name is ASCII `hub-managed`; callers cannot select it. The first line is exactly:

```text
# hub-managed format=1 generation=<decimal> rule-hash-v=<decimal> rule-hash=<64-lowercase-hex>\n
```

Both decimal fields are positive canonical decimal without sign/leading zero. Rendered bytes are valid UTF-8, LF-only, BOM-free, and have exactly one final LF. The identity parser is canonical and never infers identity from a prefix, name fallback, or unparsed provider metadata.

Every IR string and resolved folder destination is validated before I/O: reject NUL, Unicode controls, CR, LF, U+2028, and U+2029. Sieve strings use ASCII quote delimiters and only `\\` -> `\\\\` and `"` -> `\\"`; every other permitted UTF-8 byte is preserved byte-exactly with no normalization, case-folding, or re-encoding. Multiline/text literals, encoded-character/hex/Unicode escapes, concatenation, comments, interpolation, and raw insertion are forbidden. Hostile oracle vectors include quote, backslash, CR, LF, NUL, U+0001, U+2028, U+2029, `"; keep; #`, `text:\nkeep;\n.`, `${x}`, and non-ASCII permitted UTF-8 that must round-trip.

The private rendered digest is:

```text
SHA-256("hub.mail.sieve.rendered.v1\x00" || U32BE(len(accountID)) || accountID ||
U32BE(11) || "hub-managed" || U64BE(generation) || U32BE(ruleHashVersion) ||
ruleHash[32] || U32BE(len(renderedBytes)) || renderedBytes)
```

Literal vector: account `acct-1`; generation/version `1`; zero32 rule hash; rendered bytes marker with exactly 64 ASCII `0` bytes, LF, `keep;`, LF. Rendered length is `131`; preimage length is `231`; exact full preimage hex:

```text
6875622e6d61696c2e73696576652e72656e64657265642e76310000000006616363742d310000000b6875622d6d616e6167656400000000000000010000000100000000000000000000000000000000000000000000000000000000000000000000008323206875622d6d616e6167656420666f726d61743d312067656e65726174696f6e3d312072756c652d686173682d763d312072756c652d686173683d303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030300a6b6565703b0a
```

Expected digest: `537f83c108892cdc7fda9c51caed53c8539673917e1ae0351793282e5f6d0fd2`. The vector was independently matched by .NET SHA256 and Windows CryptoAPI/CertUtil; that note is not runtime evidence. Private bytes never enter persistence, logs, errors, traces, metrics, health, APIs, or test diagnostics; lifetime is bounded and clearing is best effort, without secure-erasure claim.

## Complete future oracle matrix (static/future until independently run)

| Oracle | Required future assertion | Evidence |
|---|---|---|
| External ABI | Exact sealed exports/methods/new sentinel; eight old exports absent | `NOT_RUN` |
| Factory/copy | exact factory order, v1/nonzero hash, zero/nil rejection, fresh recursive copies, no raw hash+IR adapter ABI | `NOT_RUN` |
| State/identity | all foreign and duplicate/ambiguous cases drift/no mutation; canonical-only repair under fences | `NOT_RUN` |
| Capability/I-O | exact case-sensitive Sieve capability before all Sieve/blob I/O; action/attachment/cap errors are zero-state | `NOT_RUN` |
| Mailbox ordering | one scoped snapshot; exact opaque unique folders before render/get/set; conflict/unavailable/no fallback | `NOT_RUN` |
| Renderer | hostile vectors, byte-exact UTF-8, prohibited constructs, LF/no BOM/final LF | `NOT_RUN` |
| Marker/vector | canonical marker and exact 64-zero, 231-byte preimage/digest vector | `NOT_RUN` |
| Size | absent/default; lossless JMAP unsigned validation; min cap and limit-plus-one | `NOT_RUN` |
| Reconcile/fences | refetch proof, lost-response indeterminate/Get-before-retry, 04M attempt/completion behavior | `NOT_RUN` |
| Redaction/provider surface | no private diagnostics; only `SieveScript/*`, prohibited x: calls fail | `NOT_RUN` |

## Future candidate boundary, not authorization

Future **candidate-only** ownership is: `backend/internal/mail/mailbox_types.go`, `backend/internal/mail/mailbox_engine.go`, `backend/internal/mail/mailbox_errors.go`, new `backend/internal/mail/rule_script_request.go`, `backend/tests/mail/mailbox_engine_contract_test.go` (existing external ABI oracle that must atomically replace/remove **all** old Sieve ABI coverage, not only consumer-stub methods: old `GetSieve`/`UpsertSieve` signatures; every `mail.SieveScript` construction, reflection, and exact-fields assertion; and all old `EngineSieve*` types and literals, replacing them with the exact sealed hash/request/state/error/removal oracle), and `backend/tests/mail/rule_script_request_test.go` (future sealed factory/hash/deep-copy oracle). Stalwart Sieve adapter/source and tests are later work. The rule compiler ABI is unchanged. This contract does not authorize any candidate file, implementation, test, plan, command, build, lint, validator, Docker/Compose action, Windows script, or `pro-qa-ai/**` access.

## Candidate lifecycle and review routing

Required precommit reviews: Contract/Spec/Oracle `gpt-5.6-terra` high, Security/Adversarial `gpt-5.6-terra` xhigh, Lifecycle/Dependency/Boundary `gpt-5.6-terra` high. This authoring packet requests `gpt-5.6-terra` medium with `provenance_source: dispatcher-enforced`; fallback is forbidden and telemetry is `UNKNOWN` unless exposed. Every review records actual routing or remains `UNKNOWN/BLOCKED`. `CANDIDATE READY` only means ready for these reviews, never accepted, implemented, tested, runtime-ready, or unblocked.
