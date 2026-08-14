# Task 144-03A-Address — static acceptance evidence

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` stay accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; no source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness block Rule Task 1/145-07B; launcher `NOT_READY`, executables `NOT_RUN`.

> **Status:** `STATIC_ACCEPTED` · `EXACT_SHA_PASS` · `RUNTIME_NOT_RUN`
> **Atomic commit:** `41564d003946ddb78df18efca166ca86e4055405`; parent `d659c67ce657fb33da391819edbdb954454be4b9`; tree `13a97333f08df6946e65766d3223c1ede85e1f50`
> **Subject and scope:** `feat(mail): add bounded address policy`; exactly five paths.

> **Historical mail lifecycle update:** correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract, Security and Lifecycle Terra routes PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and its Rule-plan acceptance next-step are `CONSUMED/HISTORICAL`; Rule/Phase145/runtime blockers remain. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept/review/commit instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. The Rule-plan acceptance receipt is the documentation sync; it is not a future Rule Task 4. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Scope and provenance

- Authoring route: requested exact `gpt-5.6-terra` / `xhigh`; accepted native dispatch,
  `provenance_source: dispatcher-enforced`, fallback `none`; actual model and reasoning
  telemetry: `UNKNOWN` unless exposed.
- Independent static reviews: requested exact `gpt-5.6-terra` / `high` in three distinct
  contexts; accepted native dispatch, `provenance_source: dispatcher-enforced`, fallback
  `none`; actual model and reasoning telemetry: `UNKNOWN` unless exposed.
- Documentation sync route: requested `gpt-5.6-terra` / `medium`,
  `provenance_source: dispatcher-enforced`; actual telemetry: `UNKNOWN` unless exposed.
- Owned committed code/test paths, outside this evidence document, are
  `backend/internal/mail/address_policy.go` and
  `backend/tests/mail/address_policy_test.go`.

## Exact candidate blobs

| Path | Committed Git blob |
|---|---|
| `backend/internal/mail/address_policy.go` | `fe7472d6a5bf9baa140fef95352cbbf206f49612` |
| `backend/tests/mail/address_policy_test.go` | `67f35c39dfbbfdd92573f52103798e15de922b37` |

The parent and tree identify the immutable accepted boundary. The code/test blobs above are the
committed blobs reviewed for that exact commit.

## Static-conformant behavior

The candidate validates UTF-8 email input, rejects raw values above 320 bytes before IDNA,
forbids Unicode whitespace/control and header injection, requires exactly one `@`, and permits
only ASCII RFC 5322 `atext` dot-atom local parts of 1..64 bytes. It converts the domain through
`idna.Lookup.ToASCII`, ASCII-lowercases it, enforces a 1..253-byte domain with valid 1..63-byte
lowercase ASCII labels, then returns a lowercase canonical email no longer than 320 bytes.

Display names are UTF-8, control-free, bounded to 160 runes and 640 bytes, and preserved exactly
without trim, NFC, collapsing, or case folding. Set normalization admits a configured maximum of
1..200, preserves caller input and result order, produces a concrete non-nil empty success,
rejects overcount and canonical duplicates, and yields no partial result on any failure. It
accepts `.test`, `.invalid`, and synthetic `u-*` fixtures without DNS or network access.

Only `invalid`, `too_large`, and `duplicate` are public address error kinds; their typed
sentinels and text are stable and content-free. This packet does not change OpenAPI or
persistence.

## Independent static reviews

Three independent exact-commit Terra/high reviews reported `EXACT_SHA_PASS`:

| Review | Result | C/I/M |
|---|---|---|
| Oracle conformance | `EXACT_SHA_PASS` | `0/0/0` |
| Specification conformance | `EXACT_SHA_PASS` | `0/0/0` |
| Security review | `EXACT_SHA_PASS` | `0/0/0` |

All report Spec/Quality/Security/Evidence PASS. These are exact-commit static reviews only; they
do not establish executable, phase, product, or runtime acceptance.

## Remaining gates and next action

- This packet is `STATIC_ACCEPTED`; it is not executable, phase, product, or runtime `PASS`.
- Thread is narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`; C/I/M `0/0/0`), static only.
- Folder source+oracle are narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692`
  (source `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
  C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
  `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`. Historical Rule references to
  `f1995465`/`6735ae74` are `CONSUMED/HISTORICAL` and superseded by the current `669ef761`
  sole-plan receipt above. The 144-02F docs/readiness
  boundary is separately `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `fcb5b84f` plus `2f4381a2`.
  No execution follows. Its dependency on accepted
  Address is satisfied; its callable/encoding/oracle contract is
  `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
  `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` with C/I/M `0/0/0`.
  No Folder production/test work is authorized before the gates stated above.
- The task-owned launcher is `NOT_READY`; compile, unit, race, build, lint, validators,
  Docker, and runtime are all `NOT_RUN`. No direct host or Docker substitute is valid evidence.
- `144-01B` remains `BLOCKED`; broader Phase 142/cross-contract acceptance and Phase 148 are
  pending. This evidence does not accept the phase, product, persistence, OpenAPI, or runtime.

> **Historical Rule gate:** the former `f1995465`/`6735ae74` acceptance transition is `CONSUMED/HISTORICAL` and superseded by the current `669ef761` sole-plan receipt above. 144-02D and 144-02F remain docs-only accepted; Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.
> **144-02F sync (historical next-step):** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text and Rule-plan acceptance next-step are `CONSUMED/HISTORICAL`; no implementation authority. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
