# 144-03B Folder implementation evidence

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` stay accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; no source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness block Rule Task 1/145-07B; launcher `NOT_READY`, executables `NOT_RUN`.

> **Historical mail lifecycle update:** correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and its Rule-plan acceptance next-step are `CONSUMED/HISTORICAL`; implementation/runtime blockers remain. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept/review/commit instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. The Rule-plan acceptance receipt is the documentation sync; it is not a future Rule Task 4. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Accepted source/oracle boundary

| Fact | Exact value |
|---|---|
| Source/oracle commit | `f8cb7643986fc591cae16764575337b72ef3a692` |
| Parent | `32e3dcb5b4f7ae65a88b79781646d25726ae16c6` |
| Tree | `14e7f6ce9a0f4b7912b955515d122c411159c579` |
| Source | `backend/internal/mail/folder_tree.go`, blob `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, 708 physical lines |
| Oracle | `backend/tests/mail/folder_tree_test.go`, blob `efe0db0071cb22ce4b026c635157f0fb08885d64`, 1008 physical lines |

The boundary contains exactly those two added paths and is narrowly `STATIC_ACCEPTED`, static-only.
It does not accept a product, module, phase, release, schema, handler, migration, repository,
engine or runtime.

## Exact source reviews and provenance

| Review | Route | Result |
|---|---|---|
| Spec/Oracle | `gpt-5.6-terra/high` | `EXACT_SHA_PASS` |
| Security/Adversarial | `gpt-5.6-terra/xhigh` | `EXACT_SHA_PASS` |
| ABI/Compile-shape | `gpt-5.6-terra/high` | `EXACT_SHA_PASS` |

Every review is dispatcher-enforced with no fallback; telemetry is `UNKNOWN` unless exposed.
All three report C/I/M `0/0/0` for the exact two-path commit.

## Static behavior and non-claims

The pure policy validates a bounded UUID projection snapshot, normalizes folder names privately,
enforces deterministic tree ordering and returns typed policy errors. The external oracle covers
the public callable/error shape and static policy behavior. This is source-only evidence.

Runner status is `NOT_READY`. Compile, unit, race, build, lint, validators, Docker and runtime are
all `NOT_RUN`. No direct host or Docker/Compose execution is a substitute.

## Docs candidate boundary

Step 1 literal discovery found 16 existing paths. Adding the three existing stale-gate
implementation evidence documents yields the frozen 19 existing paths; adding this one new
evidence file yields this exact 20-path candidate. Synchronizing that bounded lifecycle inventory
is not a new product decision.

The literal exact 20-path manifest, in accepted contract order and spelling, is:

1. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03B-FOLDER.md` (new evidence file)
2. `.planning/STATE.md`
3. `.planning/decisions/specs/MAIL-SPEC.md`
4. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`
5. `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md`
6. `.planning/phases/143-calendar-rsvp-core/TERMINAL-RESUME-PROMPT-RU.md`
7. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
8. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
9. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-01A.md`
10. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-ADDRESS.md`
11. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-THREAD.md`
12. `.planning/phases/144-mailbox-domain/PLAN.md`
13. `.planning/phases/148-mail-release-gate/GATE-MATRIX.md`
14. `docs/mail/CONTRACT-CHANGELOG.md`
15. `docs/mail/ENGINE-CONTRACT-0.1.8.md`
16. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
17. `docs/mail/QA-HANDOFF.md`
18. `docs/mail/openapi.yaml`
19. `.planning/phases/144-mailbox-domain/CONTRACT-144-03B-FOLDER.md`
20. `.planning/phases/144-mailbox-domain/PLAN-144-03B-FOLDER.md`

**CONSUMED/HISTORICAL:** This 20-path docs candidate formerly required its own
Lifecycle/Evidence `gpt-5.6-terra/high`, Security/Dependency `gpt-5.6-terra/xhigh` and
Exact/Boundary `gpt-5.6-terra/high` exact-SHA reviews. The Folder docs lifecycle is now
accepted static-only at `ca4e2854a319a020f7322ae0a9a98e3fb4a07374` after those reviews;
author/review provenance remains dispatcher-enforced with no fallback and telemetry `UNKNOWN`
unless exposed. This historical acceptance grants no executable or runtime authority.

## Preserved blockers and next action

`144-04B` engine-port reparent and service work remains blocked. `144-01B` remains blocked on
applicable supported current-SHA Docker evidence under Task3. Broader Phase 142 and Phase 148
remain pending or blocked.

Folder docs are accepted static-only at `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
The former `f1995465`/`6735ae74` Rule acceptance transition is `CONSUMED/HISTORICAL` and superseded
by the current `669ef761` sole-plan receipt above. 144-02D and 144-02F remain docs-only accepted.
The 02F boundary at `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus
`2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; former
transition instructions are `CONSUMED/HISTORICAL`. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED`
pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher
readiness, without implied order; post-task runtime `NOT_RUN`; `rule_compiler.go` and its test
remain unauthorized.
No source, schema or runtime authority follows; launcher is `NOT_READY` and executable evidence is `NOT_RUN`.
> **144-02F sync (historical next-step):** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text and Rule-plan acceptance next-step are `CONSUMED/HISTORICAL`; no implementation authority. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
