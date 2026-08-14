# Phase 144-02A preimplementation review — schema contract readiness

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No carrier plan/source/test/runtime authority. Rule Task 1/145-07B await separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness; launcher `NOT_READY`, executables `NOT_RUN`.

> **Historical 02F lifecycle (2026-08-14):** Correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact 28-doc transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is narrowly docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Former unaccepted/transition-review lifecycle and its Rule-plan next step are `CONSUMED/HISTORICAL`. This changes neither 02A's narrow acceptance nor any source/schema/runtime authority.

> **Historical downstream lifecycle update:** 144-02D docs-only boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` is narrowly `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three exact routes PASS C/I/M `0/0/0`. The former sole-next 02F closure-review instruction is `CONSUMED/HISTORICAL`; the current exact action is stated in the lifecycle banner above. No source/schema/runtime authority follows.
> **Detected:** 2026-08-13 root deep static audit
> **Status:** STATIC_CONTRACT_READY · DOCS_ONLY_ACCEPTED_AT_REVIEWED_SNAPSHOT_b631866b2ba372101fd80eea30951eaa267c1931 · IMPLEMENTATION_BLOCKED · RUNTIME_NOT_RUN
> **Scope:** `MailAccount`, `MailIdentity`, `MailDomain`, `MailboxPolicy` only

> `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` now supplies the root candidate matrix; this
> historical discovery verdict is preserved below; its representation closure is now docs-only accepted.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept/review/commit instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. The Rule-plan acceptance receipt is the documentation sync; it is not a future Rule Task 4. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Current root closure

`docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` now freezes the exact scalar widths, normalized
ASCII validators, typed DNS JSON value, index order/predicates, named row CHECK predicates,
edge/delete behavior, service-only cross-row invariants and the sequential 144-02A1/A2/A3
packet boundary. The historical discovery findings below are closed in the root candidate,
and three independent reviews accepted the reviewed representation snapshot: Contract/Spec
`gpt-5.6-terra/high`, Security/Adversarial `gpt-5.6-terra/xhigh`, and Lifecycle/Dependency
`gpt-5.6-terra/high`, each PASS C/I/M `0/0/0`; provenance is dispatcher-enforced, fallback
none, telemetry `UNKNOWN` unless exposed.

This acceptance covers only the frozen A1/A2/A3 persistence representation. Product writes,
Ent source/TDD, generated output, migration, launcher execution, tests/validators/build and
runtime remain blocked/not authorized. Founder migration authority, supported task-owned launcher,
sequential explicit source dispatch and source-level review remain required. 144-02D's docs/readiness
boundary is now `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. The 144-02F docs/readiness boundary at
`fcb5b84f` plus `2f4381a2` is likewise `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; the former direct
02F-closure action and its former Rule-plan authoring/adjudication/acceptance instruction are
`CONSUMED/HISTORICAL`, not code.

## Historical discovery outcome (preserved)

The packet could not satisfy its own requirement to implement the exact section 3
fields, enums and bounds. The existing documents describe the lifecycle correctly, but
several persistence shapes and cardinality limits remain non-machine-readable or
representation-ambiguous. No schema implementation should guess them.

## P1 — operational persistence matrix is incomplete

The contract does not freeze enough detail for a deterministic Ent schema:

- `MailAccount` calls engine ids/cursors/checkpoints bounded but does not bind their stored
  columns to the accepted `MaxEngineOpaqueIDBytes=1024` and
  `MaxEngineCursorBytes=4096` limits;
- `provision_error_code` and `sync_error_code` require closed safe allowlists, but no exact
  persisted enums exist;
- account and domain deployed generation/hash values are described together without an
  explicit both-null-or-both-present constraint;
- `MailDomain` names no exact columns/nullability for its stable provisioning key, engine
  id, desired/deployed hashes, retry schedule/attempts, error, lease pair or deletion
  finalization;
- `EngineDNSRecord` contains `TTL`, while the public `MailDomain.required_dns_records`
  shape omits it; the persistence projection therefore has no exact canonical shape;
- `MailboxPolicy` still lacks hard maxima for message/attachment bytes and does not state
  whether `max_attachment_bytes <= max_message_bytes` is mandatory.

The minimum amendment must publish one field table with exact names, types, nullability,
byte/item limits, defaults, immutable fields, enums and relational checks. At minimum the
database must express all lease token/until pairs and all deployed generation/hash pairs as
both present or both absent, require exact SHA-256 length, and forbid a `ready` projection
without a committed freshness instant.

## P1 — account/domain/identity cardinality is not bounded

The API permits repeated domain and alias creation, but neither the spec nor OpenAPI defines
an organization domain cap or account identity/alias cap. Both admin list responses return an
unpaginated array without `maxItems`, and `AccountSpec.Aliases` therefore also has no
authoritative bound. An administrator or compromised admin session can grow storage,
response memory and one account-control engine request without a contract ceiling.

A decision must freeze either:

1. hard creation caps plus bounded single-page responses; or
2. hard creation caps plus cursor pagination, with the account-control adapter still
   receiving a separately bounded exact alias set.

The existing engine constants provide a coherent default basis: at most 100 domains per
organization, at most 200 total identities per account including the primary identity,
and list pagination default 50 / maximum 100. These values are recommendations only until
the contract is amended.

## Representation ambiguity — normalized identity uniqueness

`MailIdentity.address` is specified as unique within an organization, but the listed entity
has no `organization_id` column. The schema must choose between a denormalized organization
foreign key with consistency enforcement or a globally unique normalized address. Because
mail domains are already globally unique and every identity must belong to its account's
organization domain, global normalized-address uniqueness is the safer non-denormalized
default, but the contract must state that strengthening explicitly.

## Deterministic details already available

- normalized domain wire length: 253 characters;
- normalized address/display name: 320 / 160 characters;
- opaque engine id and cursor/checkpoint limits: 1024 / 4096 bytes;
- public DNS record count/name/value limits: 32 / 253 / 2048;
- generation/version values are positive; attempts/epochs are non-negative;
- no password, token, administrator secret, private DKIM material or provider text belongs
  in any of the four schemas.

## Historical gate impact

- 144-02A was `BLOCKED_CONTRACT`; its representation is now docs-only accepted, but no Ent
  schema or inverse edge is authorized until the remaining source gates close;
- 144-02B/C/D/F, generation 144-02E and services/workers depending on 02A remain blocked;
- the separate founder migration decision and accepted Docker launcher remain required even
  after this contract gap is closed;
- runtime, tests, validators, generators, migrations and Docker commands remain `NOT_RUN`.

No production/test source was changed. The review used the routed Mail contract, OpenAPI,
Phase 144 packet documents and the already accepted engine-port/schema conventions only.
> **144-02F sync (historical):** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text and Rule-plan acceptance instruction are `CONSUMED/HISTORICAL`; no implementation authority. Rule TDD remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated prerequisites; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
