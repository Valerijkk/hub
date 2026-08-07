# PLAN: Phase 141 — cutover, legacy cleanup и final verification

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Переключить required CI на новую архитектуру и удалить только allowlisted
legacy entries после полного parity, сохранив проверяемый rollback и итоговый
coverage/docs evidence.

## Depends on

- Phase 140 PASS.
- Не менее двух последовательных comparable real CI dual-runs green на
  неизменённом contract.

## Success criteria

- [ ] Inventory: `unmapped = 0`, `PARITY_REQUIRED = 0`, unexpected duplicates=0.
- [ ] Legacy/new counts, behaviors, expected skips и artifacts приняты verifiers.
- [ ] Текущее решение по Go coverage явно принято; green profile получен тем же
  method и выполняет ≥82% либо заранее founder-approved новый contract.
- [ ] Real CI green после переключения; rollback проверен до cleanup.
- [ ] Корневой `testframework/` не удалён: это один Python-проект
  с `testframework/pyproject.toml`, единым pytest config и каталогами
  `{core,api,integration,contract,e2e,load,docker,reports}`.
- [ ] Все manifest-allowlisted legacy Go/Vitest/TS E2E tests, helpers, configs и
  entrypoints удалены; в `testframework/**` исполняемый test/load код только Python.
- [ ] `backend/tests/**` содержит только Go unit, `frontend/tests/**` — только
  Vitest/RTL unit; component/integration rows не замаскированы как unit.
- [ ] Hook удаляется только при consumer=0; нужные `hubtest` seams сохраняются.
- [ ] Финальная docs/state/coverage синхронизация ссылается на live evidence.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-219 | Cutover + mandatory manifest-authorized legacy/hook cleanup после preflight; rollback commit/commands проверены | devops-engineer | code-reviewer | `.github/workflows/ci.yml`, `.gitlab-ci.yml`, `Makefile`, `backend/testframework/**`, `testframework/{stages,guides,pipeline}/**`, `frontend/playwright.config.ts`, `frontend/package*.json`, `frontend/scripts/link-testframework-node-modules.mjs`, `.planning/audits/test-rebase/cutover.json` | Phase 140, cutover gates | 1–1.5д |
| TASK-220 | Независимая final verification: full runners, coverage, inventory, docs/state/commands и post-test handoff | qa-auto-ver | team-lead | `.planning/evidence/test-rebase/141/**`, `.planning/COVERAGE.md`, `.planning/STATE.md`, `AGENTS.md`, `docs/{test_plan,TESTING,testframework_strategy,TROUBLESHOOTING}.md`, `testframework/README.md` | TASK-219 | 1д |

## Волны

- Wave 0: read-only cutover preflight и rollback rehearsal.
- Wave 1: TASK-219 required-check switch; real CI green до deletion commit.
- Wave 2: allowlisted cleanup; rerun real CI.
- Wave 3: TASK-220 independent final audit/docs handoff.

## Риски

| Риск | Mitigation |
|---|---|
| Early destructive delete | Preflight gates + explicit allowlist/checksums; cleanup отдельным revertable commit |
| Root framework удалён вместе с legacy | Delete entries, не `testframework/`; корневой Python project и канонические каталоги защищены allowlist |
| Legacy оставлен «на всякий случай» | Cutover FAIL, пока Go/Vitest/TS E2E tests/configs/entrypoints не удалены по allowlist |
| Go component оказался в `backend/tests` | Ledger gate: только genuine unit; иначе Python replacement или refactor-before-move |
| Go hook удалён механически | Search + ledger consumer=0 + domain regression + verifier |
| Локальный green не повторяется в CI | Cutover требует real CI до и после cleanup; локальный run не заменяет его |
| Rollback формален | Rehearsal восстанавливает old jobs/paths одним revert и подтверждает collection |

## Verification

- `qa-auto-ver`: полный backend race/coverage, frontend type/lint/unit/coverage,
  Python API/integration, E2E smoke/full 20/53, load safety, inventory/parity gates.
- `tech-lead`: cleanup manifest, rollback evidence и coverage decision.
- Handoff: re-audit B1–B6 → productize calendar/intranet/accounting/crm/hr/bases
  → modular-monolith boundaries/schema ownership → production acceptance →
  Wiki/OIDC SSO/AI MCP. Detailed feature plans создаются только после re-audit.
