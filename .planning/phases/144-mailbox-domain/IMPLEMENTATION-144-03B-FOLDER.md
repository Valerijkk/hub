# 144-03B Folder implementation evidence

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

This candidate remains pending its own Lifecycle/Evidence `gpt-5.6-terra/high`,
Security/Dependency `gpt-5.6-terra/xhigh` and Exact/Boundary `gpt-5.6-terra/high` exact-SHA
reviews. Author provenance is `gpt-5.6-terra/medium`, dispatcher-enforced, no fallback,
telemetry `UNKNOWN`.

## Preserved blockers and next action

`144-04B` engine-port reparent and service work remains blocked. `144-01B` remains blocked on
applicable supported current-SHA Docker evidence under Task3. Broader Phase 142 and Phase 148
remain pending or blocked.

Folder docs are accepted static-only at `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
`144-03B-Rule-Compiler` is separately an unaccepted docs-only `ROOT_FROZEN_CANDIDATE`: the
current next action is its atomic 27-path docs commit, then three Rule exact-SHA reviews.
144-02D/02F closure, a later plan and a separate Phase-145 engine-port/carrier remain blocked;
this is not Rule TDD or production authorization, and `rule_compiler.go` plus its test remain
unauthorized.
