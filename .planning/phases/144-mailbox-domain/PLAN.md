# Phase 144 — Mailbox Domain

> Task: TASK-237 · Status: IN_PROGRESS · FIVE_DOCUMENT_DEFAULT_VS_HARD_CAP_144_01A_ENTRY_AMENDMENT_PASS `5e314fb0` (authorizes 144-01A entry only; broader Phase 142/cross-contract pending) · 144-01A STATIC_ACCEPTED `e3e8c774` / EXACT_SHA_PASS · 144-03A-Address STATIC_ACCEPTED `41564d00` / EXACT_SHA_PASS · 144-03A-Thread STATIC_ACCEPTED `9903f4df` / EXACT_SHA_PASS, contract `cbc12577`, C/I/M `0/0/0`, executable NOT_RUN · 144-03B source+oracle STATIC_ACCEPTED `f8cb7643986fc591cae16764575337b72ef3a692` / EXACT_SHA_PASS, C/I/M `0/0/0`, static only; exact 20-path docs candidate reviews pending, then root-only Rule contract freeze · 144-01B BLOCKED_ON_APPLICABLE_SUPPORTED_CURRENT_SHA_DOCKER_EVIDENCE_BY_TASK3 · 144-02A1/A2/A3 EXACT_PLAN_BLOCKED_REVIEW · 144-02B/02C/02D/02F CONTRACT_AMENDED_NOT_ACCEPTED · RUNTIME_NOT_RUN · Depends on: Phase 142

## Goal

Implement the engine-independent domain and HTTP behavior against a deterministic fake.

`143-04L` and `144-01A` are satisfied for `144-03D`; Address is `STATIC_ACCEPTED` at
`41564d003946ddb78df18efca166ca86e4055405`. Thread contract remains accepted at `cbc12577` and
its exact 19-path implementation/oracle boundary is `STATIC_ACCEPTED` at `9903f4df` /
`EXACT_SHA_PASS`, C/I/M `0/0/0`; this remains static-only evidence. Folder source+oracle are
narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
C/I/M `0/0/0`), static only. Its exact 20-path docs candidate awaits three exact-SHA reviews;
after PASS root only freezes/adjudicates `144-03B-Rule`; other packet eligibility is status only.

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
> **144-03B Folder post-source lifecycle (2026-08-14):** source/oracle commit `f8cb7643986fc591cae16764575337b72ef3a692` (parent `32e3dcb5b4f7ae65a88b79781646d25726ae16c6`, tree `14e7f6ce9a0f4b7912b955515d122c411159c579`) is narrowly `STATIC_ACCEPTED`, static-only, with exactly `backend/internal/mail/folder_tree.go` blob `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b` and `backend/tests/mail/folder_tree_test.go` blob `efe0db0071cb22ce4b026c635157f0fb08885d64`. Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Compile-shape Terra/high exact-SHA reviews are `EXACT_SHA_PASS`, C/I/M `0/0/0`. This exact 20-path docs candidate (the frozen prior 19-path lifecycle manifest plus its new evidence file, including the three stale-gate evidence documents) awaits its own Lifecycle/Evidence high, Security/Dependency xhigh and Exact/Boundary high exact-SHA reviews. Launcher `NOT_READY`; compile, unit, race, build, lint, validators, Docker and runtime `NOT_RUN`; 04B reparent, 144-01B, Phase 142 and Phase 148 remain blocked. Author provenance Terra/medium, `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. After candidate reviews PASS, root only freezes/adjudicates `144-03B-Rule`; no Rule TDD or production authorization.
