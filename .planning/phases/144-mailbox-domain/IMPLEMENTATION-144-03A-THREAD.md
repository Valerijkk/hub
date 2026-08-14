# Task 144-03A-Thread — exact-commit static-acceptance evidence

> **Lifecycle status:** narrow `STATIC_ACCEPTED` / `EXACT_SHA_PASS` — static only.
> **Exact accepted commit:** `9903f4df3e5ca8034ea3f043b0386170e83a5101`; parent `73060a4b879334c5b1acee68746fa834e6666424`; tree `b73f23b37747dd4d17f7fceda001c340ed81fa32`; subject `feat(mail): add deterministic thread identity`.
> **Exact scope:** 19 paths only; no drift.
> **Reviewed source blob:** `c221d547a931e084265a558be7aa0b23658b487d`.
> **External oracle blob:** `b21bd6fd7f260fabbdfa1084af6521f3ea44c771`.
> **Evidence blob committed in the accepted boundary:** `489905ac4ef4108224984929194d3db726ccfdf2`.

> **Mail lifecycle update:** correction commit `fcb5b84fed4a417e736ddb4053256e494a0c25f6` (parent `55e36b7a1a3516d07b2baf390a5742a24ec8fbe5`, tree `5a0637f2306ff42295adea5560793e2a8f4a294a`, subject `docs(mail): correct receipt and rule invariants`, exact 28 docs) is `COMMITTED_UNACCEPTED_AT_FCB5B84F` / `EXACT_SHA_TRANSITION_REQUIRED` / `IMPLEMENTATION_BLOCKED`. Contract Terra/high and Security Terra/xhigh PASS `0/0/0`; Lifecycle Terra/high FAIL `0/1/0` solely stale committed precommit lifecycle; dispatcher-enforced, no fallback, telemetry `UNKNOWN`. Former correction/precommit/root-commit-next instructions are `CONSUMED/HISTORICAL`. Next: transition review, exact 28-doc transition commit without self-SHA, Lifecycle exact-SHA re-review of fcb plus transition and trace Contract/Security, then separate 02F docs-only sync only after all PASS; Rule/Phase145/runtime blockers remain.

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
`ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately docs-only `STATIC_ACCEPTED` /
`EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; acceptance transition/review target
is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. Before any later plan or separate
Phase-145 carrier, the active 28-doc correction requires transition review, one exact transition
commit without self-SHA, fresh Lifecycle exact-SHA review of fcb plus transition with trace Contract/Security as needed, all exact reviews PASS, then separate 02F docs-only acceptance sync. 144-03C/144-03D eligibility is status
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

> **Current Rule gate:** distinct `144-03B-Rule-Compiler` is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; the acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. 144-02D is docs-only accepted; `55e36b7a` 02F closure is historical and the committed correction requires transition review, one exact transition commit without self-SHA, fresh Lifecycle exact-SHA review of fcb plus transition with trace Contract/Security as needed, all exact reviews PASS, then separate 02F docs-only acceptance sync before accepted 02F, a later plan and the separate engine-port/145 carrier precede Rule TDD/production.
