# TASK-080 — Docs metrics drift autotest

> Заведено: 2026-07-28 · Автор: W18-C5 (auditor) · Приоритет: **P2** · Статус: **Закрыта** (N/A docs/script · W13-be2)
> Источник: TASK-008 W18-C5 mutation FAIL TASK-054 · Class: P2-FE-12  
> Отчёт: `.planning/audits/total-2026-07/W18-C5-MUTATION.md`

## [OUTCOME]

Расхождение канонических счётчиков (`backend/testframework` `_test.go`, `export_hooks.go`) с цифрами в `docs/*` ловится автоматической проверкой (script и/или CI), мутация docs краснеет.

## [ROLE]

- Role: docs-writer (+ devops при CI)
- Responsibility: скрипт сверки counts + подключение в CI/PR-guard (экономно по минутам)
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: `docs/**`; `scripts/**`; `.github/workflows/pr-guard.yml` (минимально); `.planning/backlog/README.md`; этот файл
- Do not touch: product runtime; secrets; тяжёлые матрицы CI
- Constraints: студенческий пакет GH Actions — один лёгкий job/step
- Approval record: N/A

## [REFERENCES]

- TASK-054 ACCEPTANCE; W18-C5-MUTATION FAIL «нет автотеста на drift»
- Канон counts: `docs/TROUBLESHOOTING.md`, `docs/test_plan.md`, `docs/TESTING.md`, `docs/testframework_strategy.md`
- Source precedence: `go list` / `Get-ChildItem` факт > docs

## [ACCEPTANCE]

- Done criteria: команда/script сравнивает measured counts с заявленными в docs (или с golden file) и exit≠0 при drift
- **Мутация (C):** заменить в docs `618`→`515` (или актуальный count−N) → script/CI RED; restore → GREEN
- Regression: локальный запуск script; при CI — pr-guard step
- Evidence: вывод script, SHA
- Stop when: green + TASK-054 можно закрыть повторной мутацией C

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README; связь с TASK-054

## [OPERATING RULES]

- Не дублировать полный qa-gate; только inventory assert
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only · 2026-07-30 · **docs/script-only** — runtime Go/Vitest/prod mutation N/A · Finding class **P2-FE-12** / W18-C5 docs drift gate

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W18-C5-DRIFT | `scripts/check-docs-inventory.mjs` vs claimed `**N**` in `docs/TROUBLESHOOTING.md`, `docs/test_plan.md`, `docs/TESTING.md`, `docs/testframework_strategy.md` | docs + node script | N/A hubtest/Vitest — `node scripts/check-docs-inventory.mjs` exit≠0 | заменить claimed count на stale (напр. `812`→`515` или count−N) → script FAIL; restore → OK when synced |
| W18-C5-CI | `.github/workflows` — step на inventory script | CI (optional) | **GAP** — нет step в pr-guard/ci | добавить/убрать CI step → не Stage-C product RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W18-C5-DRIFT | docs inventory script | **N/A** docs/script (не prod Go/FE) |
| W18-C5-CI | — | **GAP** CI wiring |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: W18-C5-DRIFT` · `GAP IDs: W18-C5-CI`

### Gaps

- Нет hubtest/Vitest Exact RED на docs metrics — Stage C product PROVEN **не применим** (no fake-PROVEN).
- Script уже ловит drift (baseline 2026-07-30: claimed **812** ≠ measured **824** → exit 1). Sync counts + optional CI step — follow-up docs-writer/devops, не reopen как code mutate.
- TASK-054 уже **Закрыта** (N/A docs-only · W13-fe1).

## [MUTATION CLOSE — W13 be2 2026-07-30]

**Verdict:** 1/1 primary **N/A** (docs/script) · CI **GAP** noted · статус **Закрыта** (не fake-PROVEN)  
Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task079-080.md`
