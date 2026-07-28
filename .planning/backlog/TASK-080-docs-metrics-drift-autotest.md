# TASK-080 — Docs metrics drift autotest

> Заведено: 2026-07-28 · Автор: W18-C5 (auditor) · Приоритет: **P2** · Статус: Исправлена  
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
