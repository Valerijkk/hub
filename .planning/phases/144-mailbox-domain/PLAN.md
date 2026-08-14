# Phase 144 — Mailbox Domain

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No carrier plan/source/test/runtime authority. Rule Task 1/145-07B await separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness; launcher `NOT_READY`, executables `NOT_RUN`.

> Task: TASK-237 · Status: IN_PROGRESS · 144-02A A1/A2/A3 representation `STATIC_CONTRACT_READY` / docs-only accepted at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931` (three independent Terra reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry UNKNOWN unless exposed) · 144-01A STATIC_ACCEPTED `e3e8c774` / EXACT_SHA_PASS · 144-03A-Address STATIC_ACCEPTED `41564d00` / EXACT_SHA_PASS · 144-03A-Thread STATIC_ACCEPTED `9903f4df` / EXACT_SHA_PASS, contract `cbc12577`, C/I/M `0/0/0`, executable NOT_RUN · 144-02D docs-only STATIC_ACCEPTED / EXACT_SHA_PASS at `7d29d08c` plus `e563a3d3`; 144-02F docs/readiness-only STATIC_ACCEPTED / EXACT_SHA_PASS at `fcb5b84f` plus `2f4381a2`; 144-02B/02C remain unaccepted · 144-01B BLOCKED_ON_APPLICABLE_SUPPORTED_CURRENT_SHA_DOCKER_EVIDENCE_BY_TASK3 · RUNTIME_NOT_RUN · Depends on: Phase 142

> **Current 144-02D/02F acceptance:** 02D remains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus `e563a3d3be8898371e5c8a9861c58f11b72167bc`. The exact-28 correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high all PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. The former 02F lifecycle and Rule-plan author/adjudication are `CONSUMED/HISTORICAL`. No source/Ent/generation/migration/services/runtime authority follows. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase-145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; Phase 145 remains blocked.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Goal

Implement the engine-independent domain and HTTP behavior against a deterministic fake.

`143-04L` and `144-01A` are satisfied for `144-03D`; Address is `STATIC_ACCEPTED` at
`41564d003946ddb78df18efca166ca86e4055405`. Thread contract remains accepted at `cbc12577` and
its exact 19-path implementation/oracle boundary is `STATIC_ACCEPTED` at `9903f4df` /
`EXACT_SHA_PASS`, C/I/M `0/0/0`; this remains static-only evidence. Folder source+oracle are
narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
C/I/M `0/0/0`), static only; Folder docs acceptance is `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
The distinct Rule contract is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; its acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. The separately accepted Rule plan is at full `669ef761` (parent `77eb`, tree `1b554d`, added plan blob `b86c0423`) and authorizes neither source nor TDD. The future implementation boundary remains exactly `rule_compiler.go` plus `rule_compiler_test.go`, but only after all accepted non-marker Phase-145 engine-port/typed-IR/script carrier, G145-5B, G145-5C and launcher readiness.

## Ordered tasks

1. `144-01` — public `MailEngine` port, DTOs, error taxonomy and fake; write adapter
   contract units under `backend/tests/mail/**`.
2. `144-02` — Ent schemas for account, identity, domain/policy, folder, thread, message,
   message-folder membership, label, mail object, attachment, draft, user settings,
   signature, rule/deployment, mutation receipt/target/member, contact and submission;
   Docker-only generation.
3. `144-03` — pure address normalization, threading, reply/forward quoting, folder tree,
   typed rule compiler, rate/loop and attachment policies plus deterministic mapping of
   all three mail events through the shared 143-04L envelope, with failing units first.
4. `144-04` — repositories/services for folders, labels, list/read/bulk, draft revisions,
   unknown-save reconciliation, send/generic mutation idempotency/reconciliation, owned
   provisioning, settings/signatures and desired rule generation; then the
   separately packetized incremental/enumeration sync policy, epoch-fenced service, atomic Ent
   apply plus domain-control/account-control/projection-with-mutation-recovery/rule bounded worker lifecycles
   against fake ports.
5. `144-05` — dedicated mail-storage presign/confirm adapter, scanner interface,
   server sanitizer/remote-image proxy, versioned event projection consumption/subject
   admission, bounded body/filename search hydration, recent contacts, notifications/ws/
   audit seams; project file ACLs are not reused.
6. `144-06` — Fiber handlers/routes matching `docs/mail/openapi.yaml`, owner-visible
   account/projection state, admin policy/health/alias DTOs, error mapping and genuine units.
7. `144-07` — after Phase 143 admits `calendar.>`, add `mail.>` to EVENTS without changing
   the webhook filter, then system/architecture/security docs sync and independent review.

Each item is one or more atomic commits split further when generated schema, service,
handler, or docs cannot be reviewed independently. Existing Resend alert behavior stays
separate and green.

> **02A source gate:** the accepted b631 snapshot is representation-only and is not an
> implementation commit. It does not authorize Ent/product source or TDD, generated output,
> migrations, launcher execution, tests/validators/build or runtime. Founder migration authority,
> supported task-owned launcher, sequential explicit source dispatch and source-level review remain
> required. 144-02D is docs-only accepted; before any source work, 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 02F and Rule-plan authoring/review lifecycle is `CONSUMED/HISTORICAL`.
> The Rule plan is already `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; its former author/adjudication instruction is `CONSUMED/HISTORICAL`. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase-145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, with no implied order; post-task runtime is `NOT_RUN`.
> **Current Rule lifecycle (2026-08-14):** standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole added plan blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); exact Contract high, Security xhigh and Lifecycle high reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 02F and Rule-plan authoring/review is `CONSUMED/HISTORICAL`. Rule Task 1 RED remains blocked until all accepted non-marker Phase-145 engine-port/typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, unordered. No Rule source/test/TDD/schema/runtime/product/release authority; launcher `NOT_READY`, executable evidence `NOT_RUN`, Phase 148 blocked.

> **Current 144-03B rule gate:** Rule-plan acceptance is complete and authorizes neither source nor TDD. 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, with Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; former lifecycle is `CONSUMED/HISTORICAL`. The future implementation boundary remains exactly `rule_compiler.go` plus `rule_compiler_test.go`; Rule Task 1 is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase-145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order. All other blockers remain in force.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text and Rule-plan author/adjudication are `CONSUMED/HISTORICAL`; no implementation authority. Rule Task 1 remains blocked on all Phase-145 carrier/G145-5B/G145-5C gates and launcher readiness; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
