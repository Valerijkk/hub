# Mail workstream pre-implementation readiness

> **Snapshot:** 2026-08-13
> **Branch:** `codex/mail-module-20260813`
> **Product truth:** NOT_IMPLEMENTED · NOT_RUN · not accepted

## Ready

- Architecture, source-of-truth, security and compatibility decisions are frozen in Mail
  contract 0.1.2.
- OpenAPI covers mailbox, submission recovery, legacy calendar, canonical calendar and
  separate legacy/new admin resources.
- Event subjects and webhook privacy boundary are frozen.
- Phase 143 calendar, Phase 144 mailbox, Phase 145 gateway and Phase 146 UI work are split
  into exact-file atomic dispatch indexes; high-risk behaviors have full prompts.
- Independent contract/security and dispatch atomicity rereviews both returned static
  `PASS` with no remaining P0/P1/P2 after occurrence, schema-closure, inbound-REPLY-ledger,
  gateway ownership and mobile-navigation remediation.
- Phase 145 has a nine-gate legal/provenance/runtime/ops plan.
- Founder questions have safe unattended defaults and authorize no external action.

## Hard blockers

| Blocker | Evidence | Unblocks |
|---|---|---|
| Exact atomic-code runtime unavailable | `RUNTIME-BLOCKER.md` | all product/test code in Phases 143–146 |
| Accepted versioned Docker launcher not implemented | root `AGENTS.md`, `STATE.md` | Ent generation, build/test/validator/runtime evidence |
| Production migration mechanism unresolved | `MIGRATION-PLAN.md`, founder question 12 | schema generation/migration/backfill acceptance |
| Stalwart license not selected | founder question 1, Phase 145 `GATE-PLAN.md` | gateway/image adoption |
| Immutable image/config/API provenance not refreshed | Phase 145 `GATE-PLAN.md` | gateway wire implementation and Docker profile |
| Protected frontend test cutover not handed off/merged | Phase 146 `DISPATCH-INDEX.md` | executable genuine frontend units |
| Legacy broadcast governance unresolved | founder question 13 | Phase 148 mailbox acceptance |

## First authorized execution sequence

1. Founder exposes exact `gpt-5.3-codex` / `xhigh` runtime without fallback.
2. Dispatch 143-01 on the recorded clean HEAD; root reviews allowed-path diff.
3. Dispatch 143-02A then 143-02B; accept deterministic iTIP surface.
4. Resolve migration mechanism and accepted launcher before Ent generation.
5. Admit `calendar.>` before calendar producers.
6. Execute Phase 143 service/HTTP packets; retain legacy calendar envelope.
7. Execute Phase 144 port/fake/pure-policy packets, then schemas/generation and services.
8. Admit `mail.>` only after calendar admission; keep webhook filter unchanged.
9. Execute UI transport/routes/core only after backend contracts are accepted and the
   feature-test harness is available.
10. Begin Phase 145 only in the gate order; external SMTP remains separately disabled.

## Forbidden shortcuts

- Sol/Terra writing atomic product/test code;
- silent model fallback;
- host test/build/lint/validator/generator/process execution;
- direct Docker/Compose as accepted evidence;
- manual edits to generated Ent/router artifacts;
- reading or changing protected test-rebase/QA/pro-qa-ai zones;
- raw draft/message body or idempotency-key persistence in PostgreSQL;
- project membership as mailbox attachment authorization;
- blind retry after unknown draft save or submission;
- push, PR, merge, deployment, DNS, license purchase or secret mutation without authority.
