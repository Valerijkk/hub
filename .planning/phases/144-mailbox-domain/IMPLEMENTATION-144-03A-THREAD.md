# Task 144-03A-Thread — pre-commit static-conformance evidence

> **Lifecycle status:** `STATIC_CONFORMANT` atomic candidate — **not** `STATIC_ACCEPTED`.
> **Base HEAD:** `73060a4b879334c5b1acee68746fa834e6666424`.
> **Candidate scope:** exact 19-path atomic manifest below; no other path is in scope.
> **Reviewed source blob:** `c221d547a931e084265a558be7aa0b23658b487d`.
> **External oracle blob:** `b21bd6fd7f260fabbdfa1084af6521f3ea44c771`.

## Static evidence

The two uncommitted candidate files are statically conformant to the accepted callable,
encoding and oracle contract at `cbc12577a6444df9c660e43fedfd0c8133c6e461`. Three
independent static reviews — Spec/Quality, Security, and Exact/Oracle — report PASS with
C/I/M `0/0/0`. This records an atomic candidate boundary only; it is not an exact-commit
review and must not be represented as `STATIC_ACCEPTED` before an exact commit SHA exists.

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

Create one exact atomic commit containing exactly these 19 paths, then obtain independent
exact-SHA review. Do not dispatch or authorize `144-03B-Folder` yet.

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
