# Mail workstream pre-implementation readiness

> **Snapshot:** 2026-08-14
> **Branch:** `codex/mail-module-20260813`
> **Product truth:** IMPLEMENTATION_IN_PROGRESS · 143-01/02A/02B/03A/03B/03C/03D/03F/04L_STATIC_ACCEPTED · RUNTIME_NOT_RUN · Phase/product/migration/runtime not accepted
> **Execution truth:** launcher `NOT_READY`; generation, migration, tests and runtime `NOT_RUN`.

## Prepared artifacts (root and independently reviewed)

- Architecture, source-of-truth, security and compatibility decisions include the exact
  five-document default-vs-hard-cap/144-01A entry amendment, which passed at
  `5e314fb080d420696642058ae47397e251c23207`. This is not a full Mail 0.1.8, Phase 142 or
  global cross-contract PASS. The bounded 144-01A engine DTO/error boundary is `STATIC_ACCEPTED` at
  `e3e8c774d734f926f77f8e600b58dc68dabbd993` after `EXACT_SHA_PASS`. The 0.1.7
  `b3fb6f48` and 0.1.6 `5a351d5b` evidence remain historical.
- `144-03A-Thread` callable/encoding/oracle is separately
  `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
  `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` after independent
  Spec/API/Persistence, Security and Lifecycle/Evidence PASS reviews with C/I/M `0/0/0`.
  Implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`
  (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`), static only. Folder source+oracle are
  narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
  `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
  C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
  `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately an unaccepted docs-only
  `ROOT_FROZEN_CANDIDATE` pending its own commit/reviews, 02D/02F closure, later plan and
  Phase-145 carrier. No execution.
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
| Accepted versioned Docker launcher not implemented; separate founder-approved production migration mechanism/gate still required | root `AGENTS.md`, `STATE.md`, `MIGRATION-PLAN.md`, all-A non-acceptance boundary | 143-03E generation/inspection evidence; all 03A/B/C/D/F schema packets are already static accepted |
| 144-01B requires applicable supported current-SHA Docker evidence under stricter Task3 authority | Phase 144 `IMPLEMENTATION-144-01A.md`, `REVIEW-144-01A-ENGINE-BOUNDS.md`, `DISPATCH-REMAINING.md` | 144-01B deterministic fake only; Folder source/oracle and docs are static-only accepted, while Rule remains separately unaccepted pending its own commit/reviews and gates |
| 144-02A source/schema packet is unimplemented and not independently source-accepted; generation remains launcher-gated | five-document 144-01A entry amendment PASS `5e314fb080d420696642058ae47397e251c23207`, Phase 144 `REVIEW-144-02A-SCHEMA-READINESS.md`, `DISPATCH-INDEX.md` | 144-02A and every dependent mailbox schema/service packet |
| Calendar DTO/handler implementation is unimplemented; accepted 0.1.8 OpenAPI page/expansion bounds still lack executable evidence | five-document 144-01A entry amendment PASS `5e314fb080d420696642058ae47397e251c23207`; accepted 0.1.8 OpenAPI page/expansion bounds are a contract fact, Phase 142 `REVIEW-OPENAPI-RESPONSE-BOUNDS.md` | calendar list/legacy aggregate DTO, handler and UI transport packets |
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
4. 143-03A is `STATIC_ACCEPTED` on `31bfff8f58ba83d12b59600135503f98a6c92fbd`;
   143-03C is `STATIC_ACCEPTED` on `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`;
   143-03B is `STATIC_ACCEPTED` on clean `dd346e31`. 143-03F is `STATIC_ACCEPTED` on
   `846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final
   docs/whole-boundary review C/I/M `0/0/0`; the intermediate C/I/M `0/0/1` docs-sync finding
   is superseded/closed. 143-03D is `STATIC_ACCEPTED` on
   `10b4ceead119cb74a16b325b33f9121f60d7b1b5` after final independent review C/I/M `0/0/0`.
   Do not start generation: 143-03E is `BLOCKED/NOT_RUN` on the accepted launcher `NOT_READY`
   and the separate founder-approved migration mechanism/gate; 03F is satisfied.
5. 143-04L is `STATIC_ACCEPTED` at `4525c559da22729aa8bc51ff5336c29dad4f27e9` after final
   independent Terra/high PASS C/I/M `0/0/0`; requested Terra/xhigh implementation and Terra/high
   review provenance are dispatcher-enforced, actual telemetry `UNKNOWN`. Its static RED only
   does not make the launcher ready.
6. Resolve the separate migration mechanism/gate and accepted launcher before Ent generation.
7. Admit `calendar.>` before calendar producers.
8. Execute Phase 143 service/HTTP plus ScheduledCall bridge packets; retain the legacy
   calendar envelope and count linked pairs once. Keep linked external guests fail-closed
   until founder question 23 has a separately reviewed join-credential contract.
9. Phase 144-01A is `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993` after
   exact-commit Terra/high `EXACT_SHA_PASS`, Spec/Quality/Security/Evidence PASS and C/I/M
   `0/0/0`. Exact native dispatch provenance is dispatcher-enforced; actual telemetry `UNKNOWN`
   is not a telemetry-only blocker. `144-03A-Address` is `STATIC_ACCEPTED` at
   `41564d003946ddb78df18efca166ca86e4055405` after exact-SHA PASS. `144-03A-Thread`
   callable/encoding/oracle is `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
   `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` with independent
   Spec/API/Persistence, Security and Lifecycle/Evidence PASS reviews and C/I/M `0/0/0`.
    Its implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`
    (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`), static only. Folder source+oracle are
    narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
    `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
    C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
    `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately an unaccepted docs-only
    `ROOT_FROZEN_CANDIDATE` pending its own commit/reviews, 02D/02F closure, later plan and
    Phase-145 carrier. No execution.
   `144-01B` remains blocked by applicable
   supported current-SHA Docker evidence and executable status `NOT_RUN`.
10. Admit `mail.>` only after calendar admission; keep webhook filter unchanged.
11. Execute UI transport/routes/core only after backend contracts are accepted and the
   feature-test harness is available.
12. Begin Phase 145 only in the gate order; external SMTP remains separately disabled.

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

> **Current Rule candidate:** `144-03B-Rule-Compiler` is docs-only `ROOT_FROZEN_CANDIDATE`,
> unaccepted. Root commit and three exact-SHA reviews, 144-02D/02F, later plan and engine-port
> carrier remain blockers; no Rule TDD/production.
