# TASK-201 — 🧹 CI entrypoints alignment и dual-run

> Заведено: 2026-07-30 · Revision: **2026-08-07** · Приоритет: **P1** · Статус: Открыта · Фаза: **140**

## [OUTCOME]

GitHub, GitLab, Make/package scripts и pipeline manifests вызывают единые canonical entrypoints для legacy и replacement suites в явном dual-run, без удаления старых путей.

## [ROLE]

- Owner: `devops-engineer` — executable entrypoints, CI wiring, services, caches и artifacts.
- Verifier / handoff: `qa-auto-ver` — job collection/failure sensitivity и parity evidence.

## [SCOPE AND CONSTRAINTS]

- Allowed: `.github/workflows/ci.yml`, `.gitlab-ci.yml`, `Makefile`, `backend/Makefile`, `frontend/package.json`, launch scripts, `testframework/pipeline/**`, `testframework/PIPELINE.md`.
- Do not touch: test behavior, product code, legacy or target file deletion, coverage threshold/redefinition.
- GitHub remains initial executable authority; mirrors must call the same scripts/arguments. Backend `hubtest`, Go 1.25, Node 22, locked Python and required service matrix are explicit.
- Dual-run jobs use distinct names/databases/artifacts; load is manual/scheduled only.
- Depends on: TASK-193–200, TASK-202–217 (all foundations and replacement suites through phase 139) and their verifier evidence.

## [REFERENCES]

- `.planning/codebase/{STACK,INTEGRATIONS,TESTING,CONCERNS}.md`
- `.github/workflows/ci.yml`, `.gitlab-ci.yml`, existing Makefiles/manifests
- TASK-204 job/coverage contract; TASK-218 owns final sentinel/skip gates.

## [ACCEPTANCE]

- Completion command: `pwsh -NoProfile -File scripts/test-entrypoints.ps1 dual-run`; exit code 0 и заявленные counts обязательны.

- Canonical commands exist for legacy Go/Vitest/TS-E2E and target Go/FE/Python smoke/full/integration; each CI/mirror points to them without fail-open `|| true`.
- Required services/timeouts/caches/artifacts and report paths match job contract; PR vs nightly/manual policy explicit; load never runs on normal PR.
- Config validation/static checks pass; a test PR runs old and new lanes and publishes distinct collected/executed/passed/failed/skipped artifacts.
- Failing sentinel sampling: one temporary sentinel per runtime lane causes only expected lane non-zero; full enforcement is completed in TASK-218; restore GREEN and diff-clean.
- Regression: release build and current legacy commands remain available; no legacy path/linker/hook deleted.
- Evidence: workflow validation, real PR URLs/job IDs if available, command matrix, counts/artifacts, sentinel RED→GREEN.
- Stop when: parity evidence missing or CI change needs owner approval; keep dual-run and request review, never cut over silently.

## [OUTPUT SCHEMA]

1. Verdict/summary.
2. Matrix entrypoint → callers → services → artifacts → trigger.
3. Validation/PR outcomes and sentinel evidence.
4. Regression/rollback and approvals.
5. Handoff `qa-auto-ver`, risks, skill-log.

## [OPERATING RULES]

- Align scripts first, then CI callers; validate after each edit and triage failures.
- No deletion/cutover: TASK-219 owns cleanup after TASK-218 gates.
- CI changes require owner review; no push/merge by agent; docs-sync pipeline docs.
- Atomic commits with TASK link, no AI attribution.
