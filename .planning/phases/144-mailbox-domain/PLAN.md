# Phase 144 — Mailbox Domain

> Task: TASK-237 · Status: IN_PROGRESS · FIVE_DOCUMENT_DEFAULT_VS_HARD_CAP_144_01A_ENTRY_AMENDMENT_PASS `5e314fb0` (authorizes 144-01A entry only; broader Phase 142/cross-contract pending) · 144-01A STATIC_ACCEPTED `e3e8c774` / EXACT_SHA_PASS · 144-03A-Address STATIC_ACCEPTED `41564d00` / EXACT_SHA_PASS · 144-03A-Thread STATIC_ACCEPTED `9903f4df` / EXACT_SHA_PASS, contract `cbc12577`, C/I/M `0/0/0`, executable NOT_RUN · Folder source+oracle STATIC_ACCEPTED `f8cb7643986fc591cae16764575337b72ef3a692`, Folder docs acceptance `ca4e2854` · 144-03B-Rule-Compiler ROOT_FROZEN_CANDIDATE unaccepted pending root commit/reviews, 02D/02F, later plan and engine-port/145 carrier · 144-01B BLOCKED_ON_APPLICABLE_SUPPORTED_CURRENT_SHA_DOCKER_EVIDENCE_BY_TASK3 · 144-02A1/A2/A3 EXACT_PLAN_BLOCKED_REVIEW · 144-02B/02C/02D/02F CONTRACT_AMENDED_NOT_ACCEPTED · RUNTIME_NOT_RUN · Depends on: Phase 142

## Goal

Implement the engine-independent domain and HTTP behavior against a deterministic fake.

`143-04L` and `144-01A` are satisfied for `144-03D`; Address is `STATIC_ACCEPTED` at
`41564d003946ddb78df18efca166ca86e4055405`. Thread contract remains accepted at `cbc12577` and
its exact 19-path implementation/oracle boundary is `STATIC_ACCEPTED` at `9903f4df` /
`EXACT_SHA_PASS`, C/I/M `0/0/0`; this remains static-only evidence. Folder source+oracle are
narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
C/I/M `0/0/0`), static only; Folder docs acceptance is `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
The distinct Rule contract is `ROOT_FROZEN_CANDIDATE`, unaccepted; its root commit/reviews,
02D/02F closure, later plan and engine-port/145 carrier precede TDD/production.

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
> **Current Rule lifecycle (2026-08-14):** Folder source/oracle `f8cb7643986fc591cae16764575337b72ef3a692` and Folder docs `ca4e2854a319a020f7322ae0a9a98e3fb4a07374` are static-only accepted. `144-03B-Rule-Compiler` is an unaccepted docs-only `ROOT_FROZEN_CANDIDATE`: it awaits its own root atomic 27-path commit and three exact-SHA reviews C/I/M `0/0/0`, explicit 144-02D/02F closure, a later separate plan and a separate Phase-145 engine-port/carrier. Rule TDD/production remain unauthorized; launcher `NOT_READY` and all executable evidence `NOT_RUN`; Phase 148 remains blocked.

> **Current 144-03B rule gate:** `144-03B-Rule-Compiler` contract is a docs-only
> `ROOT_FROZEN_CANDIDATE`, not accepted and not a TDD plan. Its exact two-file future boundary
> is `rule_compiler.go` plus `rule_compiler_test.go`; 144-02D/02F closure, an independently
> accepted later plan and a separate engine-port/145 carrier remain mandatory.
