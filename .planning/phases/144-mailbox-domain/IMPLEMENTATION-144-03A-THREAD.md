# Task 144-03A-Thread — exact-commit static-acceptance evidence

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` stay accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; no source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness block Rule Task 1/145-07B; launcher `NOT_READY`, executables `NOT_RUN`.

> **Lifecycle status:** narrow `STATIC_ACCEPTED` / `EXACT_SHA_PASS` — static only.
> **Exact accepted commit:** `9903f4df3e5ca8034ea3f043b0386170e83a5101`; parent `73060a4b879334c5b1acee68746fa834e6666424`; tree `b73f23b37747dd4d17f7fceda001c340ed81fa32`; subject `feat(mail): add deterministic thread identity`.
> **Exact scope:** 19 paths only; no drift.
> **Reviewed source blob:** `c221d547a931e084265a558be7aa0b23658b487d`.
> **External oracle blob:** `b21bd6fd7f260fabbdfa1084af6521f3ea44c771`.
> **Evidence blob committed in the accepted boundary:** `489905ac4ef4108224984929194d3db726ccfdf2`.

> **Historical mail lifecycle update:** correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract, Security and Lifecycle Terra routes PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and its Rule-plan acceptance next-step are `CONSUMED/HISTORICAL`; Rule/Phase145/runtime blockers remain. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept/review/commit instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. The Rule-plan acceptance receipt is the documentation sync; it is not a future Rule Task 4. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Static evidence

The exact 19-path commit is narrowly `STATIC_ACCEPTED` at
`9903f4df3e5ca8034ea3f043b0386170e83a5101`, preserving the accepted callable, encoding and
oracle contract at `cbc12577a6444df9c660e43fedfd0c8133c6e461`. Three independent exact-SHA
reviews — Spec/Quality/Fresh, Security/Adversarial and Exact/Oracle — report PASS with C/I/M
`0/0/0`. This is static acceptance only, not executable, product, module, phase or release
acceptance.

## Dispatcher provenance

- Authoring: requested `gpt-5.6-terra` / `xhigh`; `provenance_source: dispatcher-enforced`;
  fallback `none`; actual model/reasoning telemetry `UNKNOWN` unless exposed.
- Spec/Quality/Fresh review: requested `gpt-5.6-terra` / `high`;
  `provenance_source: dispatcher-enforced`; fallback `none`; actual telemetry `UNKNOWN`.
- Security/Adversarial review: requested `gpt-5.6-terra` / `xhigh`;
  `provenance_source: dispatcher-enforced`; fallback `none`; actual telemetry `UNKNOWN`.
- Exact/Oracle production review: requested `gpt-5.6-terra` / `high`;
  `provenance_source: dispatcher-enforced`; fallback `none`; actual telemetry `UNKNOWN`.

For every route, exposed telemetry must match the corresponding requested route. Any mismatch,
dispatcher rejection, omitted model/effort or fallback is `BLOCKED`.

## Scope and non-claims

No schemas, OpenAPI routes, migrations, services, generation, Docker execution, tests,
build, lint, validators, or runtime are authorized or evidenced by this candidate. The
launcher remains `NOT_READY`; every executable status remains `NOT_RUN`. This does not
accept a product, module, release, or Phase 144. Phase 142 broader work, 144-02, 144-01B
and Phase 148 remain pending; 144-01B still requires applicable supported current-SHA
Docker evidence.

## Exact next action

Folder source+oracle are narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692`
(source `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
`ca4e2854a319a020f7322ae0a9a98e3fb4a07374`. Historical Rule references to
`f1995465`/`6735ae74` are `CONSUMED/HISTORICAL` and superseded by the current `669ef761`
sole-plan receipt above. The 144-02F docs/readiness boundary is
separately `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `fcb5b84f` plus `2f4381a2`. 144-03C/144-03D eligibility is status
only, not the current action.

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/144-mailbox-domain/PLAN.md`
4. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
5. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
6. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-01A.md`
7. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-ADDRESS.md`
8. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`
9. `.planning/phases/143-calendar-rsvp-core/PREIMPLEMENTATION-READINESS.md`
10. `.planning/phases/143-calendar-rsvp-core/TERMINAL-RESUME-PROMPT-RU.md`
11. `docs/mail/QA-HANDOFF.md`
12. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
13. `docs/mail/ENGINE-CONTRACT-0.1.8.md`
14. `docs/mail/CONTRACT-CHANGELOG.md`
15. `docs/mail/openapi.yaml`
16. `.planning/phases/148-mail-release-gate/GATE-MATRIX.md`
17. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-03A-THREAD.md`
18. `backend/internal/mail/threading.go`
19. `backend/tests/mail/threading_test.go`

> **Historical Rule gate:** the former `f1995465`/`6735ae74` acceptance transition is `CONSUMED/HISTORICAL` and superseded by the current `669ef761` sole-plan receipt above. 144-02D and 144-02F remain docs-only accepted; Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.
> **144-02F sync (historical next-step):** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text and Rule-plan acceptance next-step are `CONSUMED/HISTORICAL`; no implementation authority. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
