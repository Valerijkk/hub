# Phase 144 — Mailbox Domain

> Task: TASK-237 · Status: IN_PROGRESS · 144-02A A1/A2/A3 representation `STATIC_CONTRACT_READY` / docs-only accepted at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931` (three independent Terra reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry UNKNOWN unless exposed) · 144-01A STATIC_ACCEPTED `e3e8c774` / EXACT_SHA_PASS · 144-03A-Address STATIC_ACCEPTED `41564d00` / EXACT_SHA_PASS · 144-03A-Thread STATIC_ACCEPTED `9903f4df` / EXACT_SHA_PASS, contract `cbc12577`, C/I/M `0/0/0`, executable NOT_RUN · 144-02D docs-only STATIC_ACCEPTED / EXACT_SHA_PASS at `7d29d08c` plus `e563a3d3`; 144-02B/02C/02F remain unaccepted · 144-01B BLOCKED_ON_APPLICABLE_SUPPORTED_CURRENT_SHA_DOCKER_EVIDENCE_BY_TASK3 · RUNTIME_NOT_RUN · Depends on: Phase 142

> **Current 144-02D acceptance:** docs-only exact boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` and docs-only acceptance sync `55e36b7a` is narrowly `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three post-transition Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. The 55e 02F closure is `CONSUMED/HISTORICAL`: Contract/Spec/Oracle and Security/Adversarial FAIL, Lifecycle PASS. Active sole next action is the exact 28-doc `ROOT_CORRECTION_CANDIDATE` precommit sequence through root commit, independent exact-SHA review/transition, fresh 02F closure and separate docs-only sync. 02F, source/Ent/generation/migration/services/runtime, Rule plan/carrier and Phase 145 remain blocked.

## Goal

Implement the engine-independent domain and HTTP behavior against a deterministic fake.

`143-04L` and `144-01A` are satisfied for `144-03D`; Address is `STATIC_ACCEPTED` at
`41564d003946ddb78df18efca166ca86e4055405`. Thread contract remains accepted at `cbc12577` and
its exact 19-path implementation/oracle boundary is `STATIC_ACCEPTED` at `9903f4df` /
`EXACT_SHA_PASS`, C/I/M `0/0/0`; this remains static-only evidence. Folder source+oracle are
narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
C/I/M `0/0/0`), static only; Folder docs acceptance is `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
The distinct Rule contract is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; its acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. Rule docs acceptance authorizes neither source nor TDD. The future implementation boundary remains exactly `rule_compiler.go` plus `rule_compiler_test.go`, but only after accepted 02D and 02F, a separate accepted Rule plan, and the Phase-145 engine-port carrier.

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
> required. 144-02D is docs-only accepted; before any source work, the active sole next action is the
> 28-doc correction precommit-review -> root-commit -> exact-SHA -> fresh-closure -> separate-sync sequence.
> Rule plan/TDD and the Phase-145 carrier remain later.
> **Current Rule lifecycle (2026-08-14):** `144-03B-Rule-Compiler` is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. Contract/Spec Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced provenance, no fallback, telemetry `UNKNOWN` unless exposed. 144-02D is docs-only accepted; 144-02F remains unaccepted and the former direct-closure action is `CONSUMED/HISTORICAL`. Active sole next action is the 28-doc correction precommit-review -> root-commit -> exact-SHA -> fresh-closure -> separate-sync sequence. Only after accepted 02F may a separate Rule plan and Phase-145 engine-port/carrier advance. No Rule source/test/TDD/schema/runtime/product/release authority; launcher `NOT_READY`, executable evidence `NOT_RUN`, Phase 148 blocked.

> **Current 144-03B rule gate:** Rule docs acceptance authorizes neither source nor TDD. The `55e36b7a` 02F closure attempt is consumed/historical; the live exact 28-carrier candidate is precommit-review required, then root commit, exact-SHA review/transition, fresh closure reviews and separate sync. The future implementation boundary remains exactly `rule_compiler.go` plus `rule_compiler_test.go`, but only after accepted 02D and 02F, a separate accepted Rule plan, and the Phase-145 engine-port carrier. All other blockers remain in force.
