# Phase-145 sealed typed-IR/Sieve-carrier contract

> **Status:** `CORRECTION_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED` / `IMPLEMENTATION_BLOCKED`. Commit `692b99d825c66ac5e7ece6ad1f40d278a18df8aa` is the prior committed parent containing old blob `2e764c58` and the `COMMITTED_UNACCEPTED` / correction-required history; it is not a SHA, status, review verdict, or authority of this current worktree text. This standalone docs-only correction candidate has no commit SHA and is not a Phase-145 status transition, implementation plan, source/test authorization, runtime claim, or acceptance/source authority. The latest historical review trail records Contract/Spec/Oracle `gpt-5.6-terra` high `FAIL` and Lifecycle/Dependency/Boundary `gpt-5.6-terra` high `FAIL`; Security/Adversarial `gpt-5.6-terra` xhigh `PASS` is trace only. The current candidate requires fresh independent Contract/Spec/Oracle, Security/Adversarial, and Lifecycle/Dependency/Boundary precommit reviews; no earlier review is active, an acceptance, or a verdict for it. All prior routing used `provenance_source: dispatcher-enforced`; fallback is forbidden and telemetry is `UNKNOWN` unless exposed.

> **Correction lifecycle:** the current working child is a one-file correction candidate only. It requires the three precommit reviews, one atomic child commit, then three exact-SHA reviews before any status sync. It creates no acceptance, source, implementation, test, runtime, or release authority.

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

`GetRuleScript` privately derives expected bytes only after its preconditions. It returns `absent` only when no managed or foreign identity conflict exists. It returns `drifted` for a definitive expected-byte mismatch, ambiguity, duplicate managed identity, **any foreign script at the fixed name/provider identity (active or inactive)**, or **any foreign active script**. All drift returns have zero destructive mutation. `indeterminate` is a lost/unprovable outcome. From persisted/deployed `unknown` or a prior `indeterminate`, it permits only expectation-aware `GetRuleScript` observations; it never reopens an automatic write path.

`UpsertRuleScript` may return only `applied`, `drifted`, or `indeterminate` on a nil error: `applied` requires exact post-write private refetch proof; `drifted` is a definitive mismatch with no mutation; `indeterminate` may-have-applied and forbids a blind retry. Automatic bounded repair is permitted only on a fresh `UpsertRuleScript` path with no prior indeterminate/unknown outcome, after that path's initial pre-write `GetRuleScript`, for exactly one canonical managed identity, no foreign active script, and all 04M fences. Duplicate or ambiguous identities are always `drifted`. The provider-private remote state CAS below is additional to, and never substitutes for, every 04M account/generation/hash/lease/mutation-epoch/desired-tuple/deployment-token/completion fence.

The initial mailbox-context `SieveScript/get` is the sole pre-write remote-state observation for an `UpsertRuleScript` attempt. Its response for the exact request account MUST contain exactly one `state` JSON string for that `SieveScript` collection: valid JSON-string UTF-8, nonempty without trimming or normalization, and at most 1024 encoded UTF-8 bytes. The adapter captures that exact byte sequence as an adapter-private, attempt-local token. A missing, null, non-string, duplicate/ambiguous, account-mismatched, malformed, empty, or oversized state is a pre-write `Unavailable` with zero state and **no upload, `SieveScript/set`, activation, or other mutation**. The token is neither exposed in the public ABI, errors, logs, traces, metrics, API, persistence, worker payload, completion data, nor retained across a return/restart.

Every later `SieveScript/set` in that attempt, including a `SieveScript/set` containing `onSuccessActivateScript`, MUST carry `ifInState` equal byte-for-byte to the captured initial Get state. There is no set/activation fallback when that state is missing, unsupported, omitted, changed, ignored, or cannot be proven honored. Success requires exactly one account-matched standard `SetResponse` for that dispatched `SieveScript/set`, containing exactly one valid bounded provider-private `oldState` byte-for-byte equal to the captured initial Get state and exactly one valid bounded provider-private `newState` distinct from `oldState`; both states use the same nonempty JSON-string UTF-8 and <=1024 encoded-byte rules as the initial state. Absent, duplicate, ambiguous, null, non-string, malformed, empty, oversized, wrong-account, old-state mismatch, equal old/new state, or otherwise unprovable old/new state means `indeterminate`, nil error, persists `MAIL_RULE_OUTCOME_UNKNOWN`, and enters read-only/manual recovery. This includes a provider that ignores `ifInState` but returns a canonical-looking state. `stateMismatch`, an equivalent conditional failure, or any inability to prove the condition honored has the same result. It performs no cleanup, re-upload, set retry, activation retry, or second write, even if upload already left an orphan blob. It is not `drifted` or `Unavailable`. A definitive policy rejection may be `RejectedPolicy` only if the provider proves the matching exact `ifInState` was honored and that no Sieve script mutation occurred; every otherwise ambiguous rejection is `indeterminate`, nil.

A successful conditional set is proved only by that single verified `SetResponse` new state. The immediate post-write account-matched `SieveScript/get` must return state byte-for-byte equal to that exact new state and prove the exact active canonical managed script identity and exact private rendered digest. Missing, malformed, oversized, mismatched, duplicate, stale, wrong-account, or otherwise unprovable response/refetch proof is `indeterminate`, nil, persists `MAIL_RULE_OUTCOME_UNKNOWN`, and allows only read-only/manual recovery. No remote state token is a durable correlation or recovery artifact.

`EngineErrorUnavailable` with zero state is permitted only when the adapter proves that no state-changing request was dispatched or accepted: an initial Get-only transport failure, local dial/connect failure before dispatch, definitive `401`/`403` or definitive provider failure proving no mutation, or a malformed pre-write response. A timeout, connection loss, cancellation, or malformed response after an upload, `SieveScript/set` with `onSuccessActivateScript`, or any state-changing request that may have been processed returns `EngineRuleScriptState{State: EngineRuleScriptStateIndeterminate}, nil`. This includes upload uncertainty even when activation may not have occurred: the upload may have orphaned or duplicated a private effect, so retrying it blindly is forbidden.

A failed, lost, or malformed post-write refetch/verification also returns `EngineRuleScriptState{State: EngineRuleScriptStateIndeterminate}, nil`, because application cannot be proven. Every post-dispatch uncertainty after upload, set/activation, or post-write refetch records `MAIL_RULE_OUTCOME_UNKNOWN` and returns `indeterminate`, nil. Never return both `indeterminate` and an error, and do not let a generic transport or malformed-provider mapping swallow that uncertainty.

From persistence/deployment `unknown` or a prior `indeterminate`, including across retry or restart, the worker scopes recovery to the exact prior attempted tuple key `(accountID, desiredGeneration, ruleHashVersion, ruleHash)`. For that same tuple it may issue only expectation-aware `GetRuleScript` read observations, manual reconciliation, or accepted 02F exhaustion; it MUST NOT call `UpsertRuleScript`, upload, `SieveScript/set`, activation, delete, cleanup, repair, or any other write automatically. A matching expected tuple from Get permits fenced completion only under the caller's CAS/completion fences. `absent` or `drifted` does not prove that an orphan or partial artifact is absent and remains unknown for manual reconciliation; `Unavailable` likewise remains unknown with no write. While that prior tuple is unknown at `p<10`, a desired update changes only the desired tuple and MUST NOT rewrite the prior attempted tuple, lease, or budget; recovery remains read-only for that prior attempted tuple. Only at `p=10` takeover, and only when the desired tuple has genuinely advanced, may the service under the account lock clear the old attempted tuple and create the new pending desired tuple at attempt zero. Only that distinct tuple may later take the normal fresh path, still under every 04M lock/fence/capability condition; it must not reuse, clean up, correlate, or infer any prior orphan/provider ID. At `p=10` without desired advance, exhaustion is internal/no-I/O/manual and creates no fresh write path. The same tuple can never be relabeled fresh. Accepted 02F exhaustion terminalizes only its exact old unknown tuple with no I/O and never retries a write.

The carrier stores, exposes, and persists no blob, script, or provider correlation ID. Therefore set-only retry and cleanup are impossible and forbidden. Any future controlled recovery needs a separately accepted durable artifact-ID correlation, idempotency, and cleanup contract; it is not a fallback for this carrier.

Add exactly one sentinel/kind:

```go
var ErrEngineCapabilityUnavailable = errors.New("mail engine capability unavailable")
const EngineErrorCapabilityUnavailable EngineErrorKind = "capability_unavailable"
```

Extend `EngineError.Error` and `EngineError.Unwrap` with the exact safe mapping. Invalid request and rejected provider-set/script policy map to existing `RejectedPolicy`; oversize request/render or proven pre-write provider response maps to `PayloadTooLarge`; only proven no-mutation pre-dispatch transport, definitive `401`/`403`, definitive provider failure, or malformed pre-write response maps to `Unavailable`; stale/same-generation mismatch maps to `Conflict`; account scope failure maps to `UnauthorizedAccount`; unavailable capability maps to `CapabilityUnavailable`. Definitive drift and every may-have-applied indeterminate outcome return their state with nil error. Errors always return zero state and contain no adapter-private detail, provider body, capability value, marker, digest, or script bytes.

## Capability, mailbox, and I/O ordering

Before **any** Sieve or blob I/O, the adapter validates the sealed request, gets account context, and proves an exact case-sensitive mailbox-session capability `urn:ietf:params:jmap:sieve`. Missing, malformed, or account-inapplicable capability returns zero state plus `EngineErrorCapabilityUnavailable`, with no upload, set, activation, get, or blob operation. It then gates `fileinto` for move and `imap4flags` for mark-read/label; any attachment predicate blocks the entire rule set as capability unavailable until a separately accepted MIME mapping exists.

Only then it obtains one current account-scoped mailbox snapshot and resolves every move destination by its exact opaque `EngineFolderID`, uniquely. Missing, foreign, ambiguous, stale, or malformed target is `EngineErrorConflict`, zero state, and no upload/set. A malformed provider mailbox response is `Unavailable`. There is no mailbox-name fallback.

On a fresh desired tuple with no prior indeterminate/unknown outcome for that exact tuple, after resolution: render privately; perform the one initial mailbox-context `SieveScript/get` and validate/capture its exact provider-private state before any upload; then, only with all 04M fences and that exact `ifInState`, upload and issue one conditional `SieveScript/set` with `onSuccessActivateScript`; require the one account-matched standard `SetResponse` proof binding valid `oldState` to the captured state and distinct valid `newState`; refetch with `SieveScript/get`; and privately prove that exact new state, the exact active canonical managed identity, and the full rendered digest before returning `applied`. A foreign script creation or activation between the initial Get and conditional Set must fail `ifInState`; that failure is `indeterminate`/unknown/manual recovery with zero second write, including after restart, notwithstanding a possible orphan upload. The fresh path may perform the canonical managed stale repair specified above, but an uncertainty permanently selects the read-only recovery path for that exact attempted tuple. Use only mailbox RFC 9661 `SieveScript/*` surface; `x:SieveSystemScript` and `x:SieveUserScript` are prohibited. No timeout receives a blind retry.

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
| State/identity | all foreign and duplicate/ambiguous cases drift/no mutation; canonical-only repair only on fresh no-prior-unknown path under fences | `NOT_RUN` |
| Capability/I-O | exact case-sensitive Sieve capability before all Sieve/blob I/O; action/attachment/cap errors are zero-state | `NOT_RUN` |
| Mailbox ordering | one scoped snapshot; exact opaque unique folders before render/get/set; conflict/unavailable/no fallback | `NOT_RUN` |
| Renderer | hostile vectors, byte-exact UTF-8, prohibited constructs, LF/no BOM/final LF | `NOT_RUN` |
| Marker/vector | canonical marker and exact 64-zero, 231-byte preimage/digest vector | `NOT_RUN` |
| Size | absent/default; lossless JMAP unsigned validation; min cap and limit-plus-one | `NOT_RUN` |
| Reconcile/fences | refetch proof, `MAIL_RULE_OUTCOME_UNKNOWN`, 04M attempt/completion behavior, remote-state CAS in addition to—not instead of—account/generation/hash/lease/epoch/desired/deployment/completion fences, and fenced completion only after matching expected Get | `NOT_RUN` |
| Initial remote state | initial mailbox-context `SieveScript/get` returns exactly one nonempty valid <=1024-byte private account state before upload; absent/malformed/oversized state is zero-state `Unavailable` and zero writes | `NOT_RUN` |
| Conditional set | every set/onSuccessActivateScript sends exact captured `ifInState`; success has exactly one account-matched standard SetResponse with valid bounded `oldState` byte-equal initial state and distinct valid bounded `newState`; missing/unsupported/ignored/unprovable condition is `indeterminate`/unknown/manual, never fallback | `NOT_RUN` |
| Conditional conflict | concurrent foreign create/activation after Get causes `ifInState` failure; any orphan upload yields `indeterminate`, persisted unknown, restart/Get-only recovery, and zero second write | `NOT_RUN` |
| Conditional response | provider ignores `ifInState` but returns canonical-looking state; old-state mismatch; missing/duplicate/malformed/oversized/wrong-account old/new; equal old/new; and `stateMismatch` each return `indeterminate`, nil, persist unknown, and cause zero further writes/cleanup | `NOT_RUN` |
| New-state verification | refetch state must byte-equal exact verified SetResponse `newState` and prove exact active identity/private digest; state mismatch or any missing/duplicate/malformed/oversized/wrong-account proof is `indeterminate`/manual with no write | `NOT_RUN` |
| Loss before dispatch | local dial/connect or initial Get-only loss proves no state-changing request; return `Unavailable` with zero state and no write | `NOT_RUN` |
| Loss after upload dispatch | upload timeout/loss/cancel/malformed response is `indeterminate`, nil error, even when activation was not attempted | `NOT_RUN` |
| Loss after set/activation dispatch | `SieveScript/set` or `onSuccessActivateScript` timeout/loss/cancel/malformed response is `indeterminate`, nil error | `NOT_RUN` |
| Post-write verification loss | failed, lost, or malformed refetch/verification is `indeterminate`, nil error | `NOT_RUN` |
| Initial Get failure | initial Get failure is `Unavailable`, zero state, and no write | `NOT_RUN` |
| Lost upload recovery | lost upload -> `unknown` -> restart/Get `absent` -> zero second upload and no cleanup/repair | `NOT_RUN` |
| Lost set recovery | lost set/activation -> Get `drifted` -> zero write and manual reconciliation | `NOT_RUN` |
| Failed refetch recovery | failed post-write refetch -> `indeterminate` -> later matching Get `applied` permits fenced completion only | `NOT_RUN` |
| Unavailable recovery Get | `GetRuleScript` unavailable after unknown -> no write, no cleanup/repair, remains unknown | `NOT_RUN` |
| Exact-tuple recovery | old unknown tuple across retry/restart permits only Get/manual/exhaustion; it cannot be relabeled fresh, cleanup, reuse, or correlate an orphan/provider ID | `NOT_RUN` |
| Desired advance | while a prior unknown tuple is at `p<10`, a desired update changes only desired and cannot rewrite its attempted tuple, lease, or budget; recovery stays read-only for that prior tuple. Only at `p=10` takeover with genuine desired advance may the account-locked service clear the old attempted tuple and create a new pending desired tuple at attempt zero, which may later take the fresh path without prior orphan/provider correlation. At `p=10` without desired advance, exhaustion is internal/no-I/O/manual with no fresh write | `NOT_RUN` |
| Redaction/provider surface | no private diagnostics; only `SieveScript/*`, prohibited x: calls fail | `NOT_RUN` |

## Future candidate boundary, not authorization

Future **candidate-only** ownership is: `backend/internal/mail/mailbox_types.go`, `backend/internal/mail/mailbox_engine.go`, `backend/internal/mail/mailbox_errors.go`, new `backend/internal/mail/rule_script_request.go`, `backend/tests/mail/mailbox_engine_contract_test.go` (existing external ABI oracle that must atomically replace/remove **all** old Sieve ABI coverage, not only consumer-stub methods: old `GetSieve`/`UpsertSieve` signatures; every `mail.SieveScript` construction, reflection, and exact-fields assertion; and all old `EngineSieve*` types and literals, replacing them with the exact sealed hash/request/state/error/removal oracle), and `backend/tests/mail/rule_script_request_test.go` (future sealed factory/hash/deep-copy oracle). Stalwart Sieve adapter/source and tests are later work. The rule compiler ABI is unchanged. This contract does not authorize any candidate file, implementation, test, plan, command, build, lint, validator, Docker/Compose action, Windows script, or `pro-qa-ai/**` access.

## Candidate lifecycle and review routing

Required precommit reviews for the current correction candidate are Contract/Spec/Oracle `gpt-5.6-terra` high, Security/Adversarial `gpt-5.6-terra` xhigh, and Lifecycle/Dependency/Boundary `gpt-5.6-terra` high. After its atomic child commit, the same three reviews must assess that exact child SHA before any status synchronization. This authoring packet requests `gpt-5.6-terra` medium with `provenance_source: dispatcher-enforced`; fallback is forbidden and telemetry is `UNKNOWN` unless exposed. Every review records actual routing or remains `UNKNOWN/BLOCKED`. `CANDIDATE READY` only means ready for these reviews, never accepted, implemented, tested, runtime-ready, or unblocked.
