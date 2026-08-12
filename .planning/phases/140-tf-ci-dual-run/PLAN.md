# PLAN: Phase 140 — CI dual-run и execution gates

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Выровнять GitHub/GitLab/Makefile/manifests на одном job contract и добавить
legacy/new dual-run с fail-closed collection/skip/sentinel gates без cutover.

## Depends on

- Phase 131 PASS (frontend native complete).
- Phase 135 PASS (backend native disposition complete).
- Phase 136 PASS (Python API/integration).
- Phase 138 PASS (Python E2E 20/53 parity).
- Phase 139 PASS (load command/baseline; job manual/scheduled).

## Success criteria

- [ ] GitHub, GitLab, Makefile и job manifests используют согласованные paths,
  tags, services, timeouts и artifacts или зеркало явно retired.
- [ ] Legacy/new backend, frontend, API/integration, E2E smoke/full запускаются
  раздельно на одном commit; Locust только manual/scheduled.
- [ ] Backend ledger имеет `PARITY_REQUIRED = 0`, а Go coverage остаётся ≥82%
  по принятому TASK-203/204 contract; Python evidence не входит в Go profile.
- [ ] Каждый новый job проверяет non-zero collection, expected counts,
  unexpected skip=0 для required scope и working failing sentinel.
- [ ] Ни один required check/path ещё не удалён и не переключён на new-only.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-201 | Общий job contract и alignment GitHub/GitLab/Makefiles/package scripts/pipeline manifests | devops-engineer | qa-auto-ver | `.github/workflows/ci.yml`, `.gitlab-ci.yml`, `Makefile`, `backend/Makefile`, `frontend/package.json`, `testframework/pipeline/**` | Phases 131,135,136,138,139 | 1–1.5д |
| TASK-218 | Non-zero collection, exact-count, unexpected-skip и failing-sentinel gates + separate artifacts | qa-auto2 | qa-auto-ver | `scripts/ci/**`, `.planning/audits/test-rebase/baseline/**`, `.planning/evidence/test-rebase/140/**` | TASK-201 | 1д |

## Волны

- Wave 1: TASK-201.
- Wave 2: TASK-218.
- Wave 3: real CI dual-run observation; no cutover in this phase.

## Риски

| Риск | Mitigation |
|---|---|
| CI minutes удваиваются | Cached dependencies, bounded PR smoke, full parity on main/manual; gates не ослаблять |
| GitLab false mirror | Same executable scripts или явный retirement decision |
| Sentinel остаётся в suite | Temporary generated/fixture sentinel with cleanup assertion; no committed permanent fail |

## Verification

- `qa-auto-ver`: real GitHub run artifacts, sentinel red then clean green, exact
  counts/skips и legacy/new comparison.
- `devops-engineer`: GitLab/Makefile/manifests command parity report.
