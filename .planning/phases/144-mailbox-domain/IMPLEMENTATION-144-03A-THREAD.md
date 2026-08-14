# Task 144-03A-Thread — exact-commit static-acceptance evidence

> **Lifecycle status:** narrow `STATIC_ACCEPTED` / `EXACT_SHA_PASS` — static only.
> **Exact accepted commit:** `9903f4df3e5ca8034ea3f043b0386170e83a5101`; parent `73060a4b879334c5b1acee68746fa834e6666424`; tree `b73f23b37747dd4d17f7fceda001c340ed81fa32`; subject `feat(mail): add deterministic thread identity`.
> **Exact scope:** 19 paths only; no drift.
> **Reviewed source blob:** `c221d547a931e084265a558be7aa0b23658b487d`.
> **External oracle blob:** `b21bd6fd7f260fabbdfa1084af6521f3ea44c771`.
> **Evidence blob committed in the accepted boundary:** `489905ac4ef4108224984929194d3db726ccfdf2`.

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
C/I/M `0/0/0`), static only. Its exact 20-path docs candidate awaits three exact-SHA reviews;
after PASS root only freezes/adjudicates `144-03B-Rule`. 144-03C/144-03D eligibility is status
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
> **144-03B Folder post-source lifecycle (2026-08-14):** source/oracle commit `f8cb7643986fc591cae16764575337b72ef3a692` (parent `32e3dcb5b4f7ae65a88b79781646d25726ae16c6`, tree `14e7f6ce9a0f4b7912b955515d122c411159c579`) is narrowly `STATIC_ACCEPTED`, static-only, with exactly `backend/internal/mail/folder_tree.go` blob `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b` and `backend/tests/mail/folder_tree_test.go` blob `efe0db0071cb22ce4b026c635157f0fb08885d64`. Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Compile-shape Terra/high exact-SHA reviews are `EXACT_SHA_PASS`, C/I/M `0/0/0`. This exact 20-path docs candidate (the frozen prior 19-path lifecycle manifest plus its new evidence file, including the three stale-gate evidence documents) awaits its own Lifecycle/Evidence high, Security/Dependency xhigh and Exact/Boundary high exact-SHA reviews. Launcher `NOT_READY`; compile, unit, race, build, lint, validators, Docker and runtime `NOT_RUN`; 04B reparent, 144-01B, Phase 142 and Phase 148 remain blocked. Author provenance Terra/medium, `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. After candidate reviews PASS, root only freezes/adjudicates `144-03B-Rule`; no Rule TDD or production authorization.
