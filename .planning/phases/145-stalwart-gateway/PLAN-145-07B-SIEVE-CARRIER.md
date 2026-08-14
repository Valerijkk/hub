# Phase-145 sealed typed-IR/Sieve-carrier implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the marker-only Mail engine Sieve public surface with a sealed typed-IR request and safe state/error ABI, without creating a provider adapter, renderer, persistence workflow, or runtime behavior.

**Architecture:** `rule_script_request.go` owns only the sealed request, v1 value hash, pure factory, recursive copies, and a private persistence-grammar hash helper. `mailbox_types.go`, `mailbox_engine.go`, and `mailbox_errors.go` make the public replacement atomic; external tests are the only consumers and prevent the legacy marker ABI or an accidental provider surface from returning.

**Tech Stack:** Go 1.25, existing `backend/internal/mail` package, external `backend/tests/mail` Go unit oracle. Execution is only through the accepted task-owned Docker launcher when it exists; no host execution.

## Status and authority

- This is `PLAN_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED` / `IMPLEMENTATION_BLOCKED`. It has no self-SHA, current acceptance, source, test, runtime, release, schema, migration, service, worker, renderer, adapter, JMAP, CAS, persistence, or deployment authority.
- Its immutable design input is accepted carrier contract blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa` in commit `1daef78768048d7672c38a65b78ef0f119f3c207`. Plan acceptance documents the Rule prerequisite only; it does not authorize this source packet.
- Before any future RED oracle or source edit, all five conditions must be independently true: this plan is separately `STATIC_ACCEPTED` after exact-SHA reviews; the Rule compiler source/ABI (`RuleCompilerInput`, `CompiledRuleSet`, `CompileRuleSet`) is separately accepted; G145-5B is accepted; G145-5C is accepted; and the versioned task-owned launcher is ready for that dispatch. Do not ask or restate founder questions Q42-Q45.
- Launcher status is `NOT_READY`. Every Go test, build, lint, validator, Docker/Compose operation, and Windows script is `NOT_AUTHORIZED` / `NOT_RUN` until that dispatch is authorized. Static file/Git inspection never substitutes for executable evidence.
- Future production/test author route: native `gpt-5.6-terra` / `xhigh`, explicit dispatcher-enforced model and effort, no fallback; actual telemetry remains `UNKNOWN` unless exposed. Reviews use Contract/Purity `gpt-5.6-terra` / `high`, Security `gpt-5.6-terra` / `xhigh`, Lifecycle/Boundary `gpt-5.6-terra` / `high`, also no fallback.

## Global constraints

- The sole future source/test boundary is exactly six paths: modify `backend/internal/mail/mailbox_types.go`, `backend/internal/mail/mailbox_engine.go`, `backend/internal/mail/mailbox_errors.go`; create `backend/internal/mail/rule_script_request.go`; modify `backend/tests/mail/mailbox_engine_contract_test.go`; create `backend/tests/mail/rule_script_request_test.go`.
- The packet excludes every Stalwart adapter, renderer, JMAP transport, remote state CAS, blob/script I/O, persistence/schema/migration, service, worker, handler, API, UI, config, generated file, deployment, docs, and `pro-qa-ai/**` path.
- One post-gate source commit contains exactly those six paths and no AI attribution trailer. No compatibility alias, deprecated wrapper, conversion helper, dual method set, raw hash/IR adapter arguments, or old Sieve export survives.
- `EngineRuleSetHash` stays value-only with private fields; `EngineRuleScriptRequest` is sealed by an unexported marker and implemented solely by private `*engineRuleScriptRequest`. No public field or method exposes script/provider/body/blob/digest/capability/credential/mailbox data.
- The factory is pure: it copies `live`, validates the engine ID and generation/hash, calls `CompileRuleSet` on that same copy, recomputes the persistence-owned v1 hash from that same copied `live.Rules`, compares it, then retains a second private deep copy. It never performs I/O or proves 04M locking, ownership, lease, epoch, desired-tuple, deployment-token, or completion fencing.
- The private helper must encode exactly the accepted persistence grammar: SHA-256 of `hub.mail.ruleset` ASCII (16 bytes), `0x00`, `0x01`, `U32BE(active_count)`, then rules sorted by priority ascending and raw UUID16 unsigned lexicographic. The caller-owned `RuleCompilerInput` is already the complete all-and-only active (`enabled=true AND deleted_at IS NULL`) same-account ruleset, so `active_count` is exactly `len(live.Rules)` and this packet neither adds unavailable enabled/deleted fields nor refilters it. Each `U32BE(rule_frame_len)||rule_frame` is `UUID16||U32BE(priority)||U32BE(len(name UTF-8))||name||U32BE(condition_count)||stored-order condition frames||U32BE(action_count)||stored-order action frames`. A nested condition/action frame is `U32BE(frame_len)||frame`; condition is `U8(field)||U8(operator)||U8(value_tag)||value` with field from/to/subject/has_attachment = 1/2/3/4, operator equals/contains/matches_domain/is = 1/2/3/4, and tag string/bool = 1/2, string `U32BE(len UTF-8)||bytes`, bool `00|01`. Action is `U8(action)||U8(target_present)||target?`, with move/add_label/mark_read = 1/2/3; move/add_label require `01||UUID16(nonzero)`, mark_read requires `00` and no following byte. Every count, priority, frame length, and UTF-8 byte length is U32BE; every stated frame length must consume exactly its bytes with no trailing byte. No JSON, locale, time, provider value, rendered-script digest, or alternate grammar participates. The fixed vectors are empty `04886ba512485ccfa2a66d5769ff8772ade648dec3c216293bc226f5c51b4e68`, one-rule `5d08cfaba9849d4886666cdc3752588426fb96603c82a49ffee483594907c871`, and two-rule `80cd4ae7cff1967c8a25853e4786392cfbc562d86bec90fbd046e96877baf49a` with full preimages in `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` section 144-02F.

## Frozen public ABI and removal set

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
func NewEngineRuleScriptRequest(string, int64, EngineRuleSetHash, RuleCompilerInput) (EngineRuleScriptRequest, error)

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

var ErrEngineCapabilityUnavailable = errors.New("mail engine capability unavailable")
const EngineErrorCapabilityUnavailable EngineErrorKind = "capability_unavailable"
```

`NewEngineRuleSetHashV1` accepts only a nonzero `[32]byte` and returns version `1`; the zero value is invalid. Opaque engine account ID validation is byte length only: `1..MaxEngineOpaqueIDBytes` (`1024`), with no UTF-8 or control-character policy. Empty/oversize ID, nonpositive generation, zero/invalid hash, and malformed or compiler-rejected input return a zero request plus existing `EngineErrorRejectedPolicy`; only expected-vs-recomputed v1 hash mismatch returns zero request plus existing `EngineErrorConflict`. Every failure is content-free and must neither wrap nor leak `RuleCompilerError`, address sentinels, input, raw hash grammar bytes, or compiler detail. `CompiledRuleSet()` returns a new recursive deep copy for every call; hash accessors return values.

Remove exactly `SieveScript`, `EngineSieveState`, `EngineSieveStateKind`, `EngineSieveAbsent`, `EngineSieveApplied`, `EngineSieveIndeterminate`, `GetSieve`, and `UpsertSieve`. Extend only `EngineError.Error` and `EngineError.Unwrap` for the new capability kind/sentinel using its fixed safe text. Unknown kinds remain generic and do not unwrap.

## Future external oracle matrix

All items are `NOT_AUTHORIZED` / `NOT_RUN` until the authority table above is closed.

| Oracle family | Exact external assertions |
|---|---|
| Export closure | AST/export allowlist contains only the frozen new hash/request/state declarations and factory; old Sieve types, literals, methods, constructors, aliases, and wrappers are absent from every one of the six paths. |
| Sealed ABI | Only a factory result can satisfy the unexported request marker outside package `mail`; methods have exact signatures; no concrete fields, raw script, provider, name, blob, digest, capability, credential, or mailbox accessor leaks. |
| Engine interface | A consumer stub implements only `GetRuleScript`/`UpsertRuleScript` with the sealed request; old consumer compile assertions are atomically replaced. |
| Hash and factory | Constructor-produced nonzero v1 hash round-trips version/sum and can produce a valid factory request; zero-value hash fails. Empty/1025-byte ID, nonpositive generation, zero hash, and compiler rejection return zero request plus content-free `RejectedPolicy`; a 1024-byte opaque ID is accepted regardless of UTF-8/control content. Expected-vs-recomputed v1 mismatch alone returns zero request plus content-free `Conflict`. |
| Copy and purity | Mutating every nested `RuleCompilerInput` slice before/after factory and every nested `CompiledRuleSet()` result never changes another result or stored request; no source/import may add I/O. |
| Persistence v1 | Private helper is exercised only through factory against accepted canonical 02F literal vectors and one-byte/order/length mutations; stored-order frames and active count affect the digest exactly. Exact accepted maximum and maximum-plus-one rule/condition/action counts, every applicable UTF-8 rune/byte maximum, zero target UUID, and the maximum legal framing path prove validation before unsafe allocation, U32 length conversion, or arithmetic wrap. |
| Error safety | Capability sentinel/kind uses exact `errors.Is`/`errors.As`, fixed text, and safe unwrap; unknown kind remains generic/no unwrap; factory errors never contain IDs, addresses, IR, hash bytes, scripts, or provider data. |
| State grammar | Exactly four literals and one-field state shape; absence/applied/drifted/indeterminate remain distinct values; no I/O behavior or Stalwart implementation is tested in this packet. |
| Version/sealing source proof | External AST/source assertions prove the factory directly checks `expectedHash.Version() == 1`; only `NewEngineRuleSetHashV1` constructs a nonzero hash; no alternate version constructor, exported field initialization, `unsafe`, `//go:linkname`, reflection bypass, or provider/I/O import exists. External runtime is limited to zero-value invalid hash and constructor-produced v1 values; all other version and sealing guarantees are source assertions. |
| Bounded property oracle | A bounded no-panic fuzz/property corpus asserts deterministic factory output, recursive deep-copy isolation, stable content-free error text/unwrap, and bounded rejection across arbitrary opaque ID bytes and accepted compiler inputs. A literal AST deny scan over the four production paths rejects `unsafe`, `//go:linkname`, `reflect.NewAt`, network/filesystem/database/provider imports, and provider-specific identifiers. |

## Exact plan-acceptance status-sync manifest

After this plan's three exact-plan-SHA reviews pass, its separate docs-only acceptance sync is frozen to exactly the following 28 paths, equal by literal name-set to the accepted `ff427` status-sync boundary. No path may be added, removed, renamed, substituted, globbed, or implicitly expanded:

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`
4. `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md`
5. `.planning/phases/144-mailbox-domain/CONTRACT-144-03B-FOLDER.md`
6. `.planning/phases/144-mailbox-domain/CONTRACT-144-03B-RULE.md`
7. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
8. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
9. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-01A.md`
10. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-ADDRESS.md`
11. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-THREAD.md`
12. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03B-FOLDER.md`
13. `.planning/phases/144-mailbox-domain/PLAN-144-03B-FOLDER.md`
14. `.planning/phases/144-mailbox-domain/PLAN.md`
15. `.planning/phases/144-mailbox-domain/REVIEW-144-02A-SCHEMA-READINESS.md`
16. `.planning/phases/144-mailbox-domain/REVIEW-144-02D-SCHEMA-READINESS.md`
17. `.planning/phases/144-mailbox-domain/REVIEW-144-02F-SCHEMA-READINESS.md`
18. `.planning/phases/145-stalwart-gateway/DISPATCH-INDEX.md`
19. `.planning/phases/145-stalwart-gateway/DISPATCH-PACKETS.md`
20. `.planning/phases/145-stalwart-gateway/GATE-PLAN.md`
21. `.planning/phases/145-stalwart-gateway/PLAN.md`
22. `.planning/phases/148-mail-release-gate/GATE-MATRIX.md`
23. `docs/mail/CONTRACT-CHANGELOG.md`
24. `docs/mail/ENGINE-CONTRACT-0.1.8.md`
25. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
26. `docs/mail/QA-HANDOFF.md`
27. `docs/mail/openapi.yaml`
28. `docs/superpowers/plans/2026-08-13-mail-control-plane-schemas.md`

## Tasks

### Task 1: Future RED external ABI and factory oracles

**Files:**
- Modify: `backend/tests/mail/mailbox_engine_contract_test.go`
- Create: `backend/tests/mail/rule_script_request_test.go`
- Read: accepted carrier contract blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`, accepted Rule compiler source/ABI and `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` section 144-02F

**Consumes:** accepted `RuleCompilerInput`, `CompiledRuleSet`, `CompileRuleSet`, canonical persistence hash vectors, and the frozen ABI above.

**Produces:** external `package mail_test` RED tests that reject legacy Sieve symbols, ABI drift, aliasing, incorrect hash/factory ordering, and unsafe errors.

- [ ] Confirm all five authority conditions, exact six-path boundary, clean worktree, accepted Rule compiler ABI/vector source, dispatcher route, and task-owned launcher. If any is absent, return `BLOCKED` without creating tests.
- [ ] In `mailbox_engine_contract_test.go`, replace every old Sieve construction, reflection, exact-fields check, consumer-stub method, and compile assertion with AST/reflection checks for the frozen hash/request/state/engine/error ABI and an explicit no-legacy-export scan.
- [ ] Add `rule_script_request_test.go` as external `mail_test`: construct valid full `RuleCompilerInput` fixtures, obtain its expected v1 sum from the accepted canonical vectors, and assert factory values, recursive deep copies, exact content-free errors, hash accessors, state literals, capability error mapping, and no sensitive field name.
- [ ] Add mutation cases that alter rule order, condition/action order, active count, frame length, and one frame byte before factory; expected hashes must diverge according to the accepted 02F grammar. Add all-zero/zero-value hash, empty/1025-byte opaque ID, accepted 1024-byte arbitrary opaque ID, zero/negative generation, compiler-invalid input, and hash mismatch cases. Runtime tests use only constructor-produced v1 hashes; every other version assertion is static source inspection.
- [ ] Add exact accepted compiler-boundary cases using the accepted Rule constants and grammar: maximum and maximum-plus-one rule, condition, and action counts; all applicable UTF-8 rune/byte maxima; zero target UUID; and maximum legal framing. Every invalid case must return zero request plus content-free `RejectedPolicy` before unsafe allocation, U32 conversion, length addition, or wrap.
- [ ] Add a bounded no-panic fuzz/property target for deterministic request values, recursive deep-copy isolation, and error redaction. Add literal AST/source assertions that the factory directly requires `Version() == 1`, that no alternate version constructor or field initializer exists, and that the four production paths contain no `unsafe`, `//go:linkname`, `reflect.NewAt`, provider-specific identifier, or network/filesystem/database/provider import.
- [ ] Run the exact authorized launcher task for these two tests only; expected initial result is RED because the new ABI does not yet exist. Record its task-owned report path and result only after the launcher has actually produced it.
- [ ] Obtain independent Contract/Purity, Security, and Lifecycle/Boundary reviews of the RED oracle before source work. A finding blocks Task 2.

### Task 2: Future sealed carrier implementation

**Files:**
- Modify: `backend/internal/mail/mailbox_types.go`
- Modify: `backend/internal/mail/mailbox_engine.go`
- Modify: `backend/internal/mail/mailbox_errors.go`
- Create: `backend/internal/mail/rule_script_request.go`
- Test: `backend/tests/mail/mailbox_engine_contract_test.go`
- Test: `backend/tests/mail/rule_script_request_test.go`

**Consumes:** accepted RED oracle and the frozen ABI, plus accepted Rule compiler API and canonical 02F hash grammar/vectors.

**Produces:** the exact sealed carrier ABI, pure request factory, private v1 persistence-hash helper, deep-copy logic, atomic legacy removal, and capability error mapping - nothing provider-facing.

- [ ] In `mailbox_types.go`, delete the eight old Sieve declarations and add only `EngineRuleSetHash`, `EngineRuleScriptStateKind`, four literals, and one-field `EngineRuleScriptState` as frozen above.
- [ ] In `mailbox_engine.go`, replace both old Sieve methods atomically with `GetRuleScript` and `UpsertRuleScript` exact signatures; update every local compile assertion in the six-path boundary so no old method remains.
- [ ] In `mailbox_errors.go`, add only the capability sentinel/kind and the precise safe `Error`/`Unwrap` branches; retain generic behavior for unknown kinds and all pre-existing mappings.
- [ ] In `rule_script_request.go`, define the unexported concrete request/marker, nonzero v1 constructor, value accessors, recursive IR copy helpers, factory ordering, and unexported canonical persistence-v1 hash helper. The helper copies the accepted 02F frame grammar exactly, checks every accepted compiler bound before allocation/conversion/arithmetic, and has no renderer, Sieve, JMAP, adapter, network, storage, logging, or persistence call.
- [ ] Make factory validation and translation exact: opaque engine account ID is valid solely by byte length `1..1024`; do not reject non-UTF-8 or controls. Empty/oversize ID, nonpositive generation, zero/invalid hash, malformed compiler input, and compiler rejection yield zero request plus content-free `RejectedPolicy`; recomputed-hash mismatch alone yields zero request plus content-free `Conflict`. Never return a nonzero request with an error or include compiler/provider/input detail.
- [ ] Re-run the exact authorized launcher task for both external test files. Expected result is green only if every matrix row passes; then request static Contract/Purity, Security, and Lifecycle/Boundary review of the exact six-path diff.

### Task 3: One atomic future source acceptance

**Files:** exactly the six paths in the global boundary.

**Consumes:** green external oracle, three PASS static source reviews, and all authority evidence.

**Produces:** one reviewable source commit with no broadened scope.

- [ ] Recheck `git status --short`, `git diff --check`, literal six-path diff, legacy-symbol absence, placeholder scan, and no generated/docs/protected path before staging.
- [ ] Stage exactly the six literal paths and create one commit with no AI attribution. Do not combine a status sync, plan, adapter, renderer, provider, persistence, or runtime artifact.
- [ ] Capture full source SHA, parent, tree, subject, exact changed-path set, blobs, and clean-tree result. Dispatch three fresh exact-SHA reviews: Contract/Purity high, Security xhigh, Lifecycle/Boundary high. Any finding leaves the source commit unaccepted and blocks all follow-up work.

### Task 4: Later source-status acceptance sync

**Files:** an exact docs manifest discovered and frozen only after the Task 3 source commit and its three exact-source-SHA reviews; never the six source paths and never inherited from the plan-acceptance or carrier-contract manifests.

**Consumes:** source exact SHA and three PASS exact-SHA reviews.

**Produces:** docs/readiness receipt only.

- [ ] After Task 3 exact-SHA PASS only, discover every then-active status carrier that semantically records this source boundary. Freeze the reviewed exact literal docs name-set before editing; do not preselect 28 paths, reuse this plan's acceptance manifest, or inherit the carrier-contract set. No glob, substitution, omission, or implicit expansion is allowed.
- [ ] Have Contract/Purity, Security, and Lifecycle/Boundary reviewers independently confirm that the newly discovered source-status manifest is complete, minimal, source-SHA-bound, and distinct from the earlier plan/status lifecycles before any status edit.
- [ ] State only evidence actually available: static source acceptance versus executable launcher results. Preserve `NOT_RUN` for any absent Docker evidence and never claim adapter/JMAP/CAS/persistence/service/worker/runtime/release completion.
- [ ] Run fresh precommit reviews, make one docs-only commit, then fresh exact-SHA reviews. This status sync does not authorize further source or external behavior.

## Plan lifecycle

1. Fresh precommit reviews of this one file: Contract `gpt-5.6-terra/high`, Security/Purity `gpt-5.6-terra/xhigh`, Lifecycle/Boundary `gpt-5.6-terra/high`; all require PASS C/I/M `0/0/0`.
2. Root creates one one-file plan commit only after those PASS results.
3. Three fresh exact-plan-SHA reviews on the committed file require the same PASS C/I/M `0/0/0`.
4. A separate docs-only exact 28-document plan-acceptance status sync follows only then, uses exactly the frozen 28-path manifest above, binds the immutable carrier-contract blob/commit and this exact accepted plan SHA, and receives its own precommit, commit, and exact-SHA reviews. This is separate from future source Task 4, whose active docs manifest is discovered only after the future source commit.

## Self-review

- Contract coverage: ABI, legacy removals, sealed request, factory order/copies/hash mismatch, safe errors, capability error, external tests, source boundary, and no provider work each map to Tasks 1-3.
- Boundary coverage: Task 4 is isolated from source and requires a freshly discovered exact literal docs manifest whose cardinality is not preselected and which is distinct from the plan-acceptance and carrier-contract manifests; no task edits a seventh source/test path.
- Authority coverage: source remains blocked by accepted plan, accepted Rule compiler source/ABI, G145-5B, G145-5C, and launcher readiness; tests and all executables remain `NOT_AUTHORIZED` / `NOT_RUN` now.
- Placeholder scan: this plan contains no implementation placeholders, no self-SHA, and no claim of current test/runtime success.
