# Mail workstream pre-implementation readiness

> **Snapshot:** 2026-08-13
> **Branch:** `codex/mail-module-20260813`
> **Product truth:** NOT_IMPLEMENTED · NOT_RUN · not accepted

## Prepared artifacts (root-reviewed; independent pending)

- Architecture, source-of-truth, security and compatibility decisions are amended in Mail
  contract 0.1.6. Its root static review is pending; the 0.1.5 docs snapshot on
  `eae6f5e5` is historical `STALE`, and independent current-provenance review remains pending.
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
  task-local atomic-code route for this Mail branch.

## Hard blockers

| Blocker | Evidence | Unblocks |
|---|---|---|
| Accepted versioned Docker launcher not implemented | root `AGENTS.md`, `STATE.md` | Ent generation, build/test/validator/runtime evidence |
| Production migration mechanism unresolved | `MIGRATION-PLAN.md`, founder question 12 | schema generation/migration/backfill acceptance |
| Recurring ScheduledCall semantics unresolved | `MIGRATION-PLAN.md`, founder question 18 | recurring backfill and all-or-nothing cutover |
| Legacy POST idempotency compatibility break unaccepted | OpenAPI, founder question 22 | enabling the ScheduledCall bridge cutover; both shipped clients must move atomically |
| Stalwart license not selected | founder question 1, Phase 145 `GATE-PLAN.md` | gateway/image adoption |
| Immutable image/config/API provenance not refreshed | Phase 145 `GATE-PLAN.md` | gateway wire implementation and Docker profile |
| Protected frontend test cutover not handed off/merged | Phase 146 `DISPATCH-INDEX.md` | executable genuine frontend units |
| Legacy broadcast governance unresolved | founder question 13 | Phase 148 mailbox acceptance |

## First authorized execution sequence

1. Dispatch 143-01 as exact `gpt-5.6-terra` / `xhigh` on the recorded clean HEAD; root
   records actual provenance and reviews the allowed-path diff.
2. Accept or reject that packet before any dependent dispatch.
3. Dispatch 143-02A then 143-02B; accept deterministic iTIP surface.
4. Resolve migration mechanism and accepted launcher before Ent generation.
5. Admit `calendar.>` before calendar producers.
6. Execute Phase 143 service/HTTP plus ScheduledCall bridge packets; retain the legacy
   calendar envelope and count linked pairs once. Keep linked external guests fail-closed
   until founder question 23 has a separately reviewed join-credential contract.
7. Execute Phase 144 port/fake/pure-policy packets, then schemas/generation and services.
8. Admit `mail.>` only after calendar admission; keep webhook filter unchanged.
9. Execute UI transport/routes/core only after backend contracts are accepted and the
   feature-test harness is available.
10. Begin Phase 145 only in the gate order; external SMTP remains separately disabled.

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
