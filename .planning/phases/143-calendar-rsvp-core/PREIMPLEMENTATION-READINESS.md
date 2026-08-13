# Mail workstream pre-implementation readiness

> **Snapshot:** 2026-08-13
> **Branch:** `codex/mail-module-20260813`
> **Product truth:** IMPLEMENTATION_IN_PROGRESS · 143-01/02A/02B_STATIC_ACCEPTED · 03A/03C_P1_REMEDIATION_BLOCKED · 03B/03F_CHAIN_BLOCKED · RUNTIME_NOT_RUN · not accepted

## Prepared artifacts (root and independently reviewed)

- Architecture, source-of-truth, security and compatibility decisions are amended in Mail
  contract 0.1.7. Its exact `b3fb6f48` review historically passed after closing three CVE-v1
  findings, but a later root deep audit reopened a distinct P1 in engine response/admin byte
  bounds. Mail 0.1.6 SHA `5a351d5b` remains historical `STALE`.
- OpenAPI covers mailbox, submission recovery, legacy calendar, canonical calendar and
  separate legacy/new admin resources.
- Event subjects and webhook privacy boundary are frozen.
- Phase 143 calendar, Phase 144 mailbox, Phase 145 gateway and Phase 146 UI work are split
  into exact-file atomic dispatch indexes; high-risk behaviors have full prompts.
- Historical independent contract/security and dispatch atomicity rereviews returned static
  `PASS` after occurrence, schema-closure, inbound-REPLY-ledger, gateway ownership and
  mobile-navigation remediation. They target a pre-0.1.4 ancestor and remain `STALE` as
  independent current-contract evidence; the new root review does not replace them.
- Phase 145 has a nine-gate legal/provenance/runtime/ops plan.
- Founder questions have safe unattended defaults and authorize no external action.
- Founder question 11 is answered: exact `gpt-5.6-terra` / `xhigh`, no fallback, is the
  task-local atomic-code route for this Mail branch; mandatory first-read `AGENTS.md` records
  the exact narrow exception and links its provenance record.

## Hard blockers

| Blocker | Evidence | Unblocks |
|---|---|---|
| Accepted versioned Docker launcher not implemented | root `AGENTS.md`, `STATE.md` | Ent generation, build/test/validator/runtime evidence |
| Actual subagent model telemetry unavailable for new packets | `RUNTIME-BLOCKER.md`, founder question 29 | 143-03A/C remediation, independent acceptance of 143-03C/F, 143-03D, 144-01B and later packets |
| Calendar event/exception canonical hash pair CHECK absent | `REMEDIATION-143-03-HASH-PAIR.md` | re-acceptance of 143-03A/C and their dependent schema chain |
| Engine message/blob responses and admin byte settings lack authoritative hard bounds | Phase 144 `REVIEW-144-01A-ENGINE-BOUNDS.md`, founder questions 28/30/31 | accepted 144-01A amendment, fake, sync/content mapping and runtime adapter |
| Account/domain schema field matrix and domain/identity cardinality are not exact | Phase 144 `REVIEW-144-02A-SCHEMA-READINESS.md`, founder questions 32–34 | 144-02A and every dependent mailbox schema/service packet |
| Production migration mechanism unresolved | `MIGRATION-PLAN.md`, founder question 12 | schema generation/migration/backfill acceptance |
| Recurring ScheduledCall semantics unresolved | `MIGRATION-PLAN.md`, founder question 18 | recurring backfill and all-or-nothing cutover |
| Legacy POST idempotency compatibility break unaccepted | OpenAPI, founder question 22 | enabling the ScheduledCall bridge cutover; both shipped clients must move atomically |
| Stalwart license not selected | founder question 1, Phase 145 `GATE-PLAN.md` | gateway/image adoption |
| Immutable image/config/API provenance not refreshed | Phase 145 `GATE-PLAN.md` | gateway wire implementation and Docker profile |
| Protected frontend test cutover not handed off/merged | Phase 146 `DISPATCH-INDEX.md` | executable genuine frontend units |
| Legacy broadcast governance unresolved | founder question 13 | Phase 148 mailbox acceptance |

## First authorized execution sequence

1. ~~Dispatch 143-01 as exact `gpt-5.6-terra` / `xhigh` on the recorded clean HEAD; root
   records actual provenance and reviews the allowed-path diff.~~ `STATIC_ACCEPTED` on
   `e7d5f450`; evidence: `IMPLEMENTATION-143-01.md`; runtime `NOT_RUN`.
2. ~~Accept or reject that packet before any dependent dispatch.~~ Independent Terra/high
   re-review PASS with no P0/P1/P2 after remediation.
3. ~~Dispatch 143-02A and 143-02B.~~ Both are `STATIC_ACCEPTED`; 143-02B accepted through
   `efa7814d` after parser hardening and host-local timezone rejection. Evidence:
   `IMPLEMENTATION-143-02A.md` and `IMPLEMENTATION-143-02B.md`; runtime `NOT_RUN`.
4. 143-03A/C need the frozen hash-pair CHECK remediation. 143-03B/F retain source-content
   passes but do not unlock the chain. Do not start 143-03D or generation until both the P1
   and provenance gates close.
5. Resolve migration mechanism and accepted launcher before Ent generation.
6. Admit `calendar.>` before calendar producers.
7. Execute Phase 143 service/HTTP plus ScheduledCall bridge packets; retain the legacy
   calendar envelope and count linked pairs once. Keep linked external guests fail-closed
   until founder question 23 has a separately reviewed join-credential contract.
8. Phase 144-01A is source-conformant but its byte-boundary contract is reopened. Freeze the
   amendment before 144-01B; new packets also wait for actual model telemetry. Then execute
   fake/pure-policy packets, schemas/generation and services.
9. Admit `mail.>` only after calendar admission; keep webhook filter unchanged.
10. Execute UI transport/routes/core only after backend contracts are accepted and the
   feature-test harness is available.
11. Begin Phase 145 only in the gate order; external SMTP remains separately disabled.

## Forbidden shortcuts

- Root Sol or any runtime other than exact `gpt-5.6-terra` / `xhigh` writing atomic
  Mail product/test code;
- silent model fallback;
- host test/build/lint/validator/generator/process execution;
- direct Docker/Compose as accepted evidence;
- manual edits to generated Ent/router artifacts;
- reading or changing protected test-rebase/QA/pro-qa-ai zones;
- raw draft/message body or idempotency-key persistence in PostgreSQL;
- project membership as mailbox attachment authorization;
- blind retry after unknown draft save or submission;
- push, PR, merge, deployment, DNS, license purchase or secret mutation without authority.
