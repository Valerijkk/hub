# 144-03B Rule Compiler Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After every listed gate is independently accepted, add exactly one external RED oracle and one pure compiler which turn a complete selected Hub rule set into deterministic provider-neutral typed IR without emitting or handling Sieve/script bytes.

**Architecture:** `backend/tests/mail/rule_compiler_test.go` is an external `package mail_test` specification and AST/purity oracle. `backend/internal/mail/rule_compiler.go` is the only future production file; it validates the whole bounded input in the frozen global precedence, resolves only complete same-account snapshots, sorts copied rules deterministically, and returns only deep-copied typed IR. Phase 145 alone may later render and terminate a script, through a separately accepted typed-IR carrier; this plan creates neither that carrier nor any engine integration.

**Tech Stack:** Go 1.25; `github.com/google/uuid v1.6.0`; existing `NormalizeAddress` address policy; existing `golang.org/x/net/idna` Lookup profile; production imports are the literal finite allowlist below and no new dependency is added.

## Global Constraints

- This is a `PLAN_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED` document, not implementation authority. Its base is `77eb4356de06a5304521b60c57af4e73c5740ce1`; it must contain no self-SHA. The Rule contract is docs-only `STATIC_ACCEPTED` at `f199546513b758ca7b95328b5ccd7680009b732e`; 144-02D and 144-02F are docs-only accepted, and 02F exact-SHA PASS is at `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260`.
- No Task below is authorized now. Root must not dispatch the future Task 1 RED oracle until this plan receives separate exact-SHA acceptance, the separate Phase-145 typed-IR/script carrier, G145-5B, and G145-5C are explicitly accepted, and the launcher readiness gate is closed for that dispatch. Runtime/executable evidence is not a Task 1 prerequisite and remains post-task/post-source `NOT_RUN` until separately authorized.
- The future product commit has exactly two paths, in this order: `backend/tests/mail/rule_compiler_test.go`, then `backend/internal/mail/rule_compiler.go`. It excludes every schema, generated, migration, service, worker, handler, port, adapter, Sieve, provider, config, runtime, documentation, persistence-hash, OpenAPI, engine, transport, repository and command path.
- Future code author routing is native `gpt-5.6-terra` / `xhigh` under the branch exception, explicit model and effort, no fallback, `provenance_source: dispatcher-enforced`; actual telemetry is `UNKNOWN` unless exposed. Review routes are Contract/Spec/Oracle `gpt-5.6-terra` / `high`, Security/Adversarial `gpt-5.6-terra` / `xhigh`, and ABI/Purity/Lifecycle `gpt-5.6-terra` / `high`, each with no fallback and dispatcher-enforced provenance. No AI co-author trailer is permitted.
- During plan authoring and later static reviews, only read-only `git status`, `git diff --check`, `git diff`, `git diff-tree`, `git hash-object`, `git show`, `git rev-parse`, and scoped `rg` are allowed. Executable commands are not authorization substitutes.
- Input is a complete same-account enabled/non-deleted ruleset and complete all-and-only live non-deleted folder/label inventory while the caller holds the account lock. 04M owns account-scoped lookup and must mask foreign and missing targets identically; compiler `account_mismatch` and `target_ambiguous` never become an HTTP existence oracle.
- The compiler does not select rules, soft-delete, lock, persist, hash, render, escape, upload, activate, reconcile, log, schedule, access a provider, or decide Sieve termination. First-match proof in this packet is ordered typed IR only; actual termination is a Phase-145 obligation.

## Frozen production source import allowlist

`backend/internal/mail/rule_compiler.go` may import exactly these eight paths, each once and only by its declared package name (no alias, dot, or blank imports):

```go
"encoding/base32"
"errors"
"sort"
"strings"
"unicode"
"unicode/utf8"
"github.com/google/uuid"
"golang.org/x/net/idna"
```

The future implementation needs `base32` for exact keyword decode/re-encode, `errors` for the frozen sentinels, `sort` for copied rule ordering, `strings` for exact keyword/domain operations and ASCII lower-casing, `unicode` plus `utf8` for frozen text rejection, `uuid` for the public ABI and bounded lookup keys, and `idna` for the frozen Lookup domain canonicalization. Raw UUID ordering is a private fixed 16-byte comparison, so no `bytes` import is permitted. The external AST oracle must reject every other import path and every alias, dot, or blank import. In particular it must reject ambient authority and all I/O/network/filesystem/time/randomness/environment/logging/provider/JMAP/Sieve/script imports or vocabulary.

---

## Prerequisite authorization gate

| Task 1 authorization prerequisite | Required evidence | Status before acceptance |
|---|---|---|
| Plan acceptance | This exact plan committed by root then independently exact-SHA reviewed and accepted | required; this file is only a candidate |
| Typed-IR/script carrier | Separate Phase-145 design and implementation authority proves a non-marker carrier; no `SieveScript` smuggling | open |
| Mail auth | G145-5B accepted | open |
| Domain verification | G145-5C accepted | open |
| Launcher readiness for future Task 1 RED dispatch | Accepted versioned task-owned launcher with limits/report/cleanup | `NOT_READY` |

Failure of any prerequisite row leaves Task 1 RED dispatch forbidden. Static plan acceptance alone does not close any row. Runtime/executable evidence is post-task/post-source evidence only; it is currently `NOT_RUN` and is not a prerequisite for Task 1.

## Frozen future public ABI

The following is copied verbatim from the accepted Rule contract. The later two-file packet exposes exactly this ABI and no provider/script carrier.

```go
const (
    MaxRuleCompilerRules = 100; MaxRuleCompilerConditions = 20; MaxRuleCompilerActions = 10
    MaxRuleCompilerPriority = 10_000; MaxRuleCompilerNameRunes = 120; MaxRuleCompilerNameBytes = 480
    MaxRuleCompilerValueRunes = 500; MaxRuleCompilerValueBytes = 2_000; MaxRuleCompilerTargets = 100
    MaxRuleCompilerEngineFolderIDBytes = 1_024
)
type RuleConditionField string
type RuleConditionOperator string
type RuleConditionValueKind string
type RuleActionKind string
const (
    RuleConditionFrom RuleConditionField = "from"; RuleConditionTo RuleConditionField = "to"
    RuleConditionSubject RuleConditionField = "subject"; RuleConditionHasAttachment RuleConditionField = "has_attachment"
    RuleOperatorEquals RuleConditionOperator = "equals"; RuleOperatorContains RuleConditionOperator = "contains"
    RuleOperatorMatchesDomain RuleConditionOperator = "matches_domain"; RuleOperatorIs RuleConditionOperator = "is"
    RuleValueString RuleConditionValueKind = "string"; RuleValueBoolean RuleConditionValueKind = "boolean"
    RuleActionMove RuleActionKind = "move"; RuleActionAddLabel RuleActionKind = "add_label"; RuleActionMarkRead RuleActionKind = "mark_read"
)
type RuleCompilerCondition struct { Field RuleConditionField; Operator RuleConditionOperator; ValueKind RuleConditionValueKind; StringValue string; BooleanValue bool }
type RuleCompilerAction struct { Kind RuleActionKind; TargetID uuid.UUID }
type RuleCompilerRule struct { ID uuid.UUID; AccountID uuid.UUID; Name string; Priority uint32; Conditions []RuleCompilerCondition; Actions []RuleCompilerAction }
type RuleCompilerFolderTarget struct { ID uuid.UUID; AccountID uuid.UUID; EngineFolderID string }
type RuleCompilerLabelTarget struct { ID uuid.UUID; AccountID uuid.UUID; EngineKeyword EngineKeywordMarker }
type RuleCompilerInput struct { AccountID uuid.UUID; Rules []RuleCompilerRule; Folders []RuleCompilerFolderTarget; Labels []RuleCompilerLabelTarget }
type CompiledRule struct { ID uuid.UUID; Priority uint32; Conditions []RuleCompilerCondition; MarkRead bool; EngineKeywords []EngineKeywordMarker; EngineFolderID string }
type CompiledRuleSet struct { Rules []CompiledRule }
func CompileRuleSet(input RuleCompilerInput) (CompiledRuleSet, error)
```

There are exactly ten bounds: rules `0..100`, conditions/rule `1..20`, actions/rule `1..10`, priority `0..10000`, name `1..120` Unicode code points and `1..480` UTF-8 bytes, condition string `1..500` code points and `1..2000` UTF-8 bytes, each resolver inventory `0..100`, and engine folder ID valid UTF-8/control-free `1..1024` bytes. The four alias types, thirteen union literals, and eight primary data structs above are literal ABI; `RuleCompilerError` is the separately counted error ABI; no exported declaration is added.

```go
type RuleCompilerErrorKind string
const (
    RuleCompilerErrorInvalid RuleCompilerErrorKind = "invalid"; RuleCompilerErrorTooLarge RuleCompilerErrorKind = "too_large"
    RuleCompilerErrorDuplicate RuleCompilerErrorKind = "duplicate"; RuleCompilerErrorAccountMismatch RuleCompilerErrorKind = "account_mismatch"
    RuleCompilerErrorTargetUnavailable RuleCompilerErrorKind = "target_unavailable"; RuleCompilerErrorTargetAmbiguous RuleCompilerErrorKind = "target_ambiguous"
)
var (
    ErrRuleCompilerInvalid = errors.New("mail rule invalid"); ErrRuleCompilerTooLarge = errors.New("mail rule exceeds bounds")
    ErrRuleCompilerDuplicate = errors.New("mail rule duplicate"); ErrRuleCompilerAccountMismatch = errors.New("mail rule account mismatch")
    ErrRuleCompilerTargetUnavailable = errors.New("mail rule target unavailable"); ErrRuleCompilerTargetAmbiguous = errors.New("mail rule target ambiguous")
)
type RuleCompilerError struct { Kind RuleCompilerErrorKind }
func (e *RuleCompilerError) Error() string
func (e *RuleCompilerError) Unwrap() error
```

The pointer receiver is frozen: `func (e *RuleCompilerError) Error() string` and `func (e *RuleCompilerError) Unwrap() error`. A value `RuleCompilerError` does not implement `error`. Nil/unknown returns exactly `mail rule compiler error` and unwraps nil; known kinds unwrap only their matching sentinel. Every compiler failure is a fresh `*RuleCompilerError`, never wraps or exposes Address sentinels, and returns zero `CompiledRuleSet` with zero partial slices.

## Fixed behavior and error adjudication

- Global complete-input precedence is `too_large` → `invalid` → `account_mismatch` → `duplicate` → `target_ambiguous` → `target_unavailable`; adjudicate over the complete input, not discovery order. Duplicate live resolver identity is `target_ambiguous`; only a duplicate rule ID, repeated `move`, repeated `mark_read`, repeated `add_label` `TargetID`, or duplicate private label keyword is `duplicate`. Distinct `add_label` targets are valid and retain stored action order.
- Each rule/account/target UUID is nonzero and each resolver row uses the input account. Rules are copied and sorted by ascending `(Priority, unsigned lexicographic raw 16 RFC-4122 UUID bytes)`. Condition order is retained; `CompiledRule` contains optional move, optional mark-read, then unique label keywords in first stored action order.
- The closed condition tuples are: `from|to + equals + string` through `NormalizeAddress` with empty display name and lower-case ASCII email; `from|to + contains + string` exact accepted bytes; `from|to + matches_domain + string` accepted IDNA Lookup lower-case ASCII domain (`1..253`, labels `1..63` bytes); `subject + equals|contains + string` exact accepted bytes; `has_attachment + is + boolean` with its boolean preserved. All irrelevant union fields are zero and every other tuple is invalid. `to` is only visible structured To; domain matching is equality, never suffix/wildcard.
- Names and string values must be valid UTF-8/nonempty/in bounds and reject Unicode control, format, surrogate, private-use, unassigned, CR, LF, NUL, U+2028, U+2029. Preserve accepted bytes: no trim, normalization, case fold, or whitespace collapse. Quotes, backslashes, semicolons, braces, comments, and `${...}` remain opaque IR data.
- `NormalizeAddress` error translation is exactly Address `too_large` → Rule `too_large`, Address `invalid` → Rule `invalid`, defensive unreachable Address `duplicate` → Rule `duplicate`; create a fresh RuleCompilerError and expose neither Address sentinel nor text.
- `move` needs exactly one same-account folder and emits its opaque engine ID; `add_label` needs exactly one label and emits only its private keyword; `mark_read` requires `uuid.Nil`, move/add_label require nonzero target. Missing is `target_unavailable`, multiple live target rows is `target_ambiguous`, cross-account is `account_mismatch`, malformed/wrong-kind resolver is `invalid`.
- A keyword is exactly 57 ASCII bytes: `hubk-` plus canonical lower-case unpadded RFC-4648 base32 of exactly 32 bytes. Validation decodes exactly 32 bytes then re-encodes and compares identical suffix; `[a-z2-7]{52}` alone fails. Output has no Hub target UUID/name, rule name, public label value, raw target name, script, provider data, hash, or secret.
- Successful zero-rule compilation returns `CompiledRuleSet{Rules: []CompiledRule{}}` with a concrete nonnil empty slice. Output and all nested slices are fresh deep copies, do not alias either input or a previous output, and input remains byte-for-byte unchanged.

## Complete future oracle matrix

| Oracle family | Exact assertions for `package mail_test` |
|---|---|
| Literal ABI and AST | Exactly the ten bounds; four alias types (`RuleConditionField`, `RuleConditionOperator`, `RuleConditionValueKind`, `RuleActionKind`); thirteen union literals (`RuleConditionFrom`, `RuleConditionTo`, `RuleConditionSubject`, `RuleConditionHasAttachment`, `RuleOperatorEquals`, `RuleOperatorContains`, `RuleOperatorMatchesDomain`, `RuleOperatorIs`, `RuleValueString`, `RuleValueBoolean`, `RuleActionMove`, `RuleActionAddLabel`, `RuleActionMarkRead`); eight primary data structs (`RuleCompilerCondition`, `RuleCompilerAction`, `RuleCompilerRule`, `RuleCompilerFolderTarget`, `RuleCompilerLabelTarget`, `RuleCompilerInput`, `CompiledRule`, `CompiledRuleSet`); separate error ABI types (`RuleCompilerErrorKind`, `RuleCompilerError`); six error-kind constants (`RuleCompilerErrorInvalid`, `RuleCompilerErrorTooLarge`, `RuleCompilerErrorDuplicate`, `RuleCompilerErrorAccountMismatch`, `RuleCompilerErrorTargetUnavailable`, `RuleCompilerErrorTargetAmbiguous`); six sentinel variables (`ErrRuleCompilerInvalid`, `ErrRuleCompilerTooLarge`, `ErrRuleCompilerDuplicate`, `ErrRuleCompilerAccountMismatch`, `ErrRuleCompilerTargetUnavailable`, `ErrRuleCompilerTargetAmbiguous`); one function (`CompileRuleSet`); and only `(*RuleCompilerError).Error` and `(*RuleCompilerError).Unwrap`. The AST/export allowlist rejects every other exported declaration, value-receiver method, and function; a value `RuleCompilerError` must not implement `error`. Signature/result fields match the frozen block exactly. |
| Error safety | nil/unknown Error/Unwrap; `errors.Is` only exact sentinel; `errors.As` fresh typed pointer; safe fixed text and no input/UUID/address/keyword/engine-ID/provider/script/hash leakage; Address translation never leaks Address sentinels. |
| Bounds | 0/1/100/101 rules; 0/1/20/21 conditions; 0/1/10/11 actions; priority 0/10000/10001; all name/value rune and byte minima/maxima/+1; raw byte overflow before Unicode/IDNA; inventories 0/100/101; folder ID 1/1024/1025. |
| Closed tuples | Every allowed field/operator/value-kind tuple, including false `has_attachment`; each forbidden field/operator/kind combination; every irrelevant union field set/nonzero; condition order remains exact. |
| Unicode and injection | invalid UTF-8; each prohibited Unicode class; CR/LF/NUL/U+2028/U+2029; exact preservation for permitted Unicode and whitespace; quotes, slashes, comments, semicolon, braces and interpolation-like bytes retained as opaque data. |
| Address and IDNA | `NormalizeAddress` canonical address acceptance, invalid/too-large translation, no display name, reserved synthetic domains, raw preflight; domain IDNA Lookup/lowercase, 1/253 domain and 1/63 label boundaries, equality not suffix/wildcard. |
| Keywords | valid 32-byte decode/re-encode canonical lower-case keyword; 57-byte exact prefix/length; upper-case, padding, malformed alphabet, valid-shape noncanonical encoding, decoded length !=32 and duplicate private keyword rejection. |
| Identity and resolver | zero account/rule/target UUID; each foreign account row; duplicate rule ID, repeated move, repeated mark-read, repeated add-label TargetID, duplicate private label keyword, and distinct add-label targets retained in stored order; duplicate live resolver ID as ambiguous; missing target; cross-account/malformed/wrong-kind; resolver inventory and repeat/target availability cases; 04M masking is documented as outside compiler HTTP behavior. |
| Canonical output | equal-priority raw UUID byte ordering, priority ordering, original condition sequence, move/mark-read/label projection, label first-stored order, concrete empty result, repeatability, zero partials on every error. |
| Ownership | bidirectional deep-copy/nonalias after mutating input and after mutating returned nested values; compilation does not mutate input; two calls have independent allocations. |
| Purity | affirmative import allowlist; no `init`, callbacks/function types, goroutines, mutable globals, context, I/O, network, filesystem, time, randomness, environment, logs, provider/JMAP/Sieve/script names/types/imports, repository/port/adapter/handler/worker/service references. |

## Task 1: Future external RED oracle — forbidden until its authorization prerequisites close

**Files:**
- Create: `backend/tests/mail/rule_compiler_test.go`
- Read only: `backend/internal/mail/address_policy.go`, `backend/internal/mail/threading.go`, `backend/internal/mail/folder_tree.go`, `backend/internal/mail/mailbox_types.go`, `backend/tests/mail/address_policy_test.go`, `backend/tests/mail/threading_test.go`, `backend/tests/mail/folder_tree_test.go`, `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`, `docs/mail/ENGINE-CONTRACT-0.1.8.md`, `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: the frozen ABI above, `mail.NormalizeAddress`, `mail.EngineKeywordMarker`, and no unaccepted engine port.
- Produces: an external compile-time/literal/behavior/AST oracle for `mail.CompileRuleSet`; it must not declare any production compatibility type or helper.

<read_first>
Read the accepted Rule contract, this separately exact-SHA accepted plan, the existing external mail policy tests, and these exact Phase-145 paths: `.planning/phases/145-stalwart-gateway/DISPATCH-INDEX.md`, `.planning/phases/145-stalwart-gateway/DISPATCH-PACKETS.md`, `.planning/phases/145-stalwart-gateway/GATE-PLAN.md`, `.planning/phases/145-stalwart-gateway/PLAN.md`. Before writing a byte, confirm only the Task 1 authorization prerequisites: separate plan exact-SHA acceptance, accepted Phase-145 typed-IR/script carrier, accepted G145-5B, accepted G145-5C, and launcher readiness for this future RED dispatch. Runtime/executable evidence is not a prerequisite; it remains post-task/post-source `NOT_RUN`. Otherwise return `BLOCKED` with no edit.
</read_first>

<action>
- [ ] Create an external `package mail_test` RED oracle with deterministic `uuid.MustParse` fixtures only; it imports public `internal/mail` API and AST packages solely to inspect `rule_compiler.go` after it exists.
- [ ] Encode every row of the complete oracle matrix above as named tests/helpers, including all six error kind/sentinel/text triples and zero-result assertions.
- [ ] Assert the global full-input precedence with inputs containing simultaneous faults: `too_large`, then `invalid`, then `account_mismatch`, then `duplicate`, then `target_ambiguous`, then `target_unavailable`.
- [ ] Assert first-match only as the sorted typed-IR sequence and never as rendered Sieve/termination behavior.
- [ ] Write AST checks that enumerate the complete frozen export allowlist from the Literal ABI and AST row (every permitted type, constant, sentinel variable, function, and only the two permitted pointer methods), prove value `RuleCompilerError` does not implement `error`, and accept only the exact literal production import allowlist above. Reject every other import, alias, dot, or blank import; forbidden mutable/effectful mechanisms; and provider/JMAP/Sieve/script carrier vocabulary.
</action>

<acceptance_criteria>
The file is the first product path, external package only, supplies no implementation, covers every matrix row, and intentionally fails before `rule_compiler.go` exists. Its source never authorizes test execution, runtime, Sieve, provider, persistence, or carrier work.
</acceptance_criteria>

<verify>
Static-only: root checks the exact one-path diff, `git diff --check`, scoped placeholder scan, and three independent static reviews: Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh, ABI/Purity/Lifecycle Terra/high. No Go/Docker/launcher command is authorized by this task.
</verify>

## Task 2: Future pure compiler source — forbidden until Task 1 is accepted

**Files:**
- Create: `backend/internal/mail/rule_compiler.go`
- Read only: Task 1 oracle and the same policy/source documents named in Task 1

**Interfaces:**
- Consumes: only `RuleCompilerInput` and existing address normalization.
- Produces: exactly `CompileRuleSet(RuleCompilerInput) (CompiledRuleSet, error)` and the ABI above; its non-exported helpers return only private validation/index/projection data.

<read_first>
Read the accepted exact Task-1 oracle and each of its three PASS reviews; re-check all prerequisite gates and confirm that the allowed file boundary remains exactly two paths. If any check is absent, return `BLOCKED` and do not write source.
</read_first>

<action>
- [ ] Define the frozen aliases, constants, structs, sentinels, and `RuleCompilerError` pointer methods verbatim; unknown/nil is generic/nil unwrap and known errors unwrap exactly one Rule sentinel.
- [ ] Copy inputs before sorting or canonicalizing. Make bounded private indexes for lookup only, validate all complete inventories/rules/actions/conditions under the frozen global precedence, and never derive output order by map iteration.
- [ ] Apply exact condition validation and address/IDNA conversion. Translate address errors to a freshly allocated RuleCompilerError, preserving no underlying sentinel or unsafe detail.
- [ ] Resolve actions to opaque folder ID/private keyword, reject only duplicate rule IDs, repeated move, repeated mark-read, repeated add-label TargetIDs, and duplicate private label keywords at duplicate precedence; permit distinct add-label TargetIDs and preserve their stored action order; reject live resolver identity ambiguity at its adjudicated precedence; and construct only copied `CompiledRuleSet` values in priority/raw-UUID/condition/label order.
- [ ] Return `CompiledRuleSet{Rules: []CompiledRule{}}` for valid zero rules and `CompiledRuleSet{}` for every failure; do not expose partial slices.
- [ ] Keep source pure: no init, callback, goroutine, mutable state, context, I/O, network, filesystem, clock, random, env, logger, provider/JMAP/Sieve/script, port, adapter, service, worker, handler, config or runtime reference.
</action>

<acceptance_criteria>
Only this source path plus Task 1 is changed. The complete oracle matrix is satisfied by construction; compiler output contains typed IR only and Phase 145 retains every rendering/termination/capability/upload/reconciliation responsibility.
</acceptance_criteria>

<verify>
Static-only: root repeats exact two-path diff/path boundary and `git diff --check`, verifies the public ABI/signature/errors/purity with scoped `rg` and AST-review reading, then obtains the same three independent static review routes. Executable validation remains `NOT_AUTHORIZED / NOT_RUN`.
</verify>

## Task 3: Root-only two-path atomic commit and exact-SHA reviews — forbidden until Tasks 1–2 pass static review

**Files:**
- Commit only: `backend/tests/mail/rule_compiler_test.go`, `backend/internal/mail/rule_compiler.go`

**Interfaces:**
- Consumes: reviewed two-file source/oracle packet.
- Produces: one root-created exact commit for independent review; no implementation acceptance is implied by commit creation.

<read_first>
Root reads all six prior static reviews, re-checks clean intended paths, confirms no concurrent/unrelated changes are staged, and confirms execution gates remain unclosed.
</read_first>

<action>
- [ ] Root alone stages exactly the two named paths and makes one atomic commit with no AI attribution/co-author trailer.
- [ ] Record parent, exact commit SHA, tree, two-path diff and blob hashes as review input, never in a self-accepting source claim.
- [ ] Dispatch three fresh exact-SHA reviews: Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh, ABI/Purity/Lifecycle Terra/high; require C/I/M `0/0/0`, dispatcher-enforced provenance, no fallback, telemetry `UNKNOWN` unless exposed.
</action>

<acceptance_criteria>
The commit changes exactly two product paths, each exact-SHA review independently targets that commit, and no review claim upgrades runtime or Phase-145 authority.
</acceptance_criteria>

<verify>
Read-only Git evidence only: exact parent/SHA/tree/path list/diff/blob checks plus three review records. Tests/build/lint/validator/Docker remain `NOT_RUN`.
</verify>

## Task 4: Separately discovered and frozen docs-only evidence sync — forbidden until Task 3 source acceptance

**Files:**
- Create/modify only after discovery: the minimum explicitly identified lifecycle/QA evidence documents outside the product packet
- Never modify: either Task 1/Task 2 product path in this task

**Interfaces:**
- Consumes: accepted Task 3 exact-SHA review records and newly discovered documentation destinations.
- Produces: a separately reviewed docs-only evidence packet; it is not a source amendment.

<read_first>
Root first discovers exact documentation destinations with scoped read-only `rg`, freezes the list in the dispatch contract, and verifies Task 3 is accepted. If discovery finds no required authoritative destination, do not invent a documentation change.
</read_first>

<action>
- [ ] Record only factual static evidence: exact source commit/review provenance, two-file boundary, `NOT_READY` launcher, `NOT_RUN` executables, still-open Phase-145 carrier/G145-5B/G145-5C, and Phase 145 ownership of script rendering/termination.
- [ ] Submit this docs-only packet for its own exact-commit review; do not amend or restage product code and do not claim a runtime/product/release PASS.
</action>

<acceptance_criteria>
The evidence sync is discovered/frozen separately, contains no self-SHA acceptance, changes no product source, and preserves the 04M masking and Phase-145 boundary.
</acceptance_criteria>

<verify>
Root uses only exact-path Git/rg/diff/hash evidence and independent exact-commit review. No executable command becomes permitted.
</verify>

## Threat model and reversibility

| Threat | Planned control |
|---|---|
| Injection into a future script | Compiler retains opaque typed data only; no Sieve/script/import/escaping carrier, while Phase 145 owns a separate quoted-string proof. |
| Target existence disclosure | Compiler receives complete internal snapshots; 04M alone resolves HTTP targets and masks foreign/missing identically. |
| Cross-account or ambiguous routing | Nonzero same-account IDs, complete inventories, account mismatch and duplicate-live identity ambiguity are rejected before output. |
| Nondeterminism/hash drift | Raw UUID ordering, preserved condition/label order, map lookup-only policy, deep copies and repeatability oracle prevent ordering/allocation leakage. |
| Unsafe error disclosure | Six fixed sentinel mappings; fresh typed wrapper; nil/unknown generic; translation never exposes Address errors or input values. |
| Side effects or ambient authority | AST purity oracle rejects state, callbacks, concurrency, clock/random/env/I/O/network/fs/log/provider/JMAP/Sieve/script and runtime layers. |
| Premature execution authority | Plan, oracle, source, exact-SHA reviews and docs sync are distinct gates; launcher and Phase-145 gates remain explicit blockers. |

Every future product change is one reversible two-path atomic commit. Revert only that identified commit with root approval after preserving review evidence; do not reset, rewrite, or alter 02D/02F/Phase-145 documents as part of a product revert. A docs sync is independently reversible because it is a later separate commit.

## Requirements and decision coverage

| Contract decision | Plan location |
|---|---|
| Exact ABI, bounds, closed union, error ABI | Frozen ABI; Task 1 literal/behavior oracle; Task 2 implementation actions |
| Complete validation precedence and no partials | Fixed adjudication; Task 1 precedence matrix; Task 2 complete-input validation |
| Deterministic typed IR and deep copy | Fixed behavior; Task 1 canonical/ownership matrix; Task 2 projection action |
| Address/IDNA/Unicode/keyword safety | Fixed behavior; oracle rows; Task 2 normalization action |
| Purity/no provider or script | Global constraints; Task 1 AST oracle; Task 2 purity action |
| 04M masking and Phase-145 termination boundary | Global constraints; Task 1 first-match assertion; Task 4 evidence sync |
| Separate acceptance and blocked runtime | Prerequisite gate; Task 3 exact-SHA reviews; Task 4 docs-only sync |

## Plan self-review checklist

- [x] Exactly two future product paths are named in order; all prohibited path classes are excluded.
- [x] The ABI block contains all ten bounds, four aliases, thirteen union literals, eight primary data structs, separately counted `RuleCompilerError` error ABI, exact function signature, six error kinds/sentinels/texts, and pointer-only method rule.
- [x] The oracle matrix covers ABI/AST, nil/unknown/errors.Is/errors.As/nonleakage, bounds, tuple closure, Unicode/injection, NormalizeAddress/IDNA, canonical keyword, UUID/resolver/account/repeat, deterministic ordering, concrete empty/zero partials, deep copy and purity.
- [x] Duplicate and global precedence adjudications, Address translation, 04M masking, and Phase-145 first-match ownership are explicit.
- [x] Every task has concrete `<read_first>`, `<action>`, `<acceptance_criteria>`, and `<verify>` content; no task authorizes current execution.
- [x] Review routing/provenance, no-fallback/telemetry truth, no-AI-coauthor rule, threat model, reversibility, artifacts, and requirements coverage are stated.

## Artifacts

| Artifact | Owner/status | Meaning |
|---|---|---|
| `CONTRACT-144-03B-RULE.md` | accepted docs-only | Frozen compiler boundary; not source authority |
| This plan | candidate | Requires independent exact-commit plan review before any dispatch |
| Future RED oracle | Task 1 blocked | External static specification only |
| Future pure compiler | Task 2 blocked | Typed-IR-only source |
| Future two-path commit + reviews | Task 3 blocked | Static evidence only, never runtime PASS |
| Future docs sync | Task 4 blocked | Separately discovered/frozen docs-only evidence |

## Execution handoff

Root must not dispatch Task 1 until this plan has separate exact-SHA acceptance, the Phase-145 typed-IR/script carrier plus G145-5B and G145-5C are closed, and launcher readiness is accepted for that future RED dispatch. Once these Task 1 authorization prerequisites close, use the stated Terra/xhigh author route and three Terra review routes with dispatcher-enforced provenance/no fallback; stop and return `BLOCKED` on any missing prerequisite or telemetry mismatch. Runtime/executable evidence is post-task/post-source `NOT_RUN`, not a Task 1 prerequisite; no plan acceptance, static review, commit, or docs sync authorizes an executable command until the approved task-owned launcher exists.

## PLANNING COMPLETE
