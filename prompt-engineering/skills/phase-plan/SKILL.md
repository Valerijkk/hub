---
name: phase-plan
description: Сформировать outcome-first PLAN фазы с атомарными задачами, границами файлов, dependency waves и фальсифицируемыми success criteria. Использовать до первой волны execute или при replan после verify-goal FAIL или конфликтов файлов.
disable-model-invocation: true
---

# Планирование фазы (`phase-plan`)

Превратить verified goal в исполнимый PLAN: атомарные задачи, параллельные волны, evidence-backed success criteria и явный handoff к execution.

## Когда использовать

- Новая фича/фаза до первого `wave-orchestrate`
- После discuss/spec, когда scope зафиксирован
- Replan после `verify-goal` FAIL или пересекающейся file map в волне

## Когда не использовать

- Микро-фикс одного файла (один артефакт, один слой) — execute inline
- Goal всё ещё неоднозначен — сначала уточнить spec

## Когда остановиться

- Goal нельзя выразить одним фальсифицируемым исходом → остановиться; запросить spec
- Success criteria не могут назвать тип evidence (command, path, metric) → остановиться; уточнить criteria
- Нет profile/context для commands или layout → отметить в PLAN как UNKNOWN; не изобретать paths

## Контекст (загрузить сначала)

| Нужно | Primary | Legacy fallback |
|------|---------|-----------------|
| Ограничения проекта | `context/projects/<profile>/PROJECT.md` | `context/PROJECT-CONTEXT.md` |
| Commands для criteria | `context/projects/<profile>/COMMANDS.md` | PROJECT-CONTEXT §4 |
| Scope тестов/регрессии | `context/projects/<profile>/CI-PIPELINE.md`, `TEST-ARCHITECTURE.md` | TESTFRAMEWORK-MAP |
| Шаблон делегирования | `agents/templates/phase-kickoff.md`, `wave-delegate.md` | — |
| Lifecycle | `playbooks/gsd-lifecycle.md` | — |

От пользователя: `{{goal}}`, constraints, список Won't, опциональные ссылки STATE/roadmap.

## Outcome-first шаги

1. **Goal (одно предложение)** — проверяемый исход, не активность («добавить тесты» ❌ → «deny foreign resource with 403 and regression green» ✅).
2. **Success criteria** — каждый пункт мапится на evidence: command + expected exit, file path, doc section или measurable status. Нет критерия без типа evidence.
3. **Атомарные задачи** — один артефакт / один слой / цель ≤200–300 LOC diff. Дробить, если задача соединяет «и» через слои (BE+FE, schema+handler, feature+security).
4. **Таблица задач** — # | task | owner role | files only | do not touch | depends | estimate | verifier.
5. **Waves** — параллелить только при **нулевом пересечении файлов** внутри волны; sync point между волнами.
6. **Risks** — impact + mitigation; изменения security/auth помечать границей human approval.
7. **Stopping** — PLAN complete, когда записан в согласованный path; execution — `wave-orchestrate`, не этот skill.
8. **Handoff** — далее: `wave-orchestrate` → per-task execution → `verify-goal`.

## Шаблон PLAN (выход)

```markdown
# PLAN: {{phase_name}}

## Goal
<one verifiable outcome sentence>

## Won't (out of scope)
- …

## Depends on
- [ ] {{artifact / ADR / prior wave}}

## Success criteria
- [ ] C1 — evidence: `{{command}}` or `{{path}}`
- [ ] C2 — …
- [ ] Docs/context sync if behavior or commands changed

## Tasks
| # | Task | Owner | Files only | Do not touch | Depends | Est | Verifier |
|---|------|-------|------------|--------------|---------|-----|----------|
| 1 | … | … | … | … | — | S | … |

## Waves
- Wave 1 (parallel): #1, #2 — file map: no overlap
- Sync: merge artifacts, conflict check
- Wave 2: #3 (depends Wave 1)

## Quality gates
- New behavior: implement → test → fix loop until green
- Regression scope: {{from COMMANDS.md / CI-PIPELINE.md}}
- Integration gate: {{project gate command or N/A with reason}}

## Risks
| Risk | Impact | Mitigation |
|------|--------|------------|

## Handoff
- Execute: `wave-orchestrate`
- Accept: `verify-goal` against criteria above
```

Записать PLAN в согласованное место (напр. `.planning/phases/<id>/PLAN.md` или эквивалент проекта).

## Evidence до handoff

Добавить блок [rules/40](../../rules/40-verification-protocol.md) — фокус на полноте плана, не на выполнении кода:

```markdown
Verification:
- Claims checked: goal falsifiable; criteria have evidence types; waves file-disjoint
- Evidence: PLAN path; profile commands cited where used
- Not run / N/A: code tests (pre-execution)
- Risks / assumptions: UNKNOWN slots, approval boundaries
- Mode: self-check
- Verdict: PASS | FAIL | PARTIAL
```

## Анти-паттерны

- Задачи без границ файлов
- Success criteria, зеркалящие названия задач без evidence
- Смешанные слои в одной задаче
- Пропуск Won't / Depends

## Ссылки

- Wave execution: `../wave-orchestrate/SKILL.md`
- Goal acceptance: `../verify-goal/SKILL.md`
- Lifecycle: `../../playbooks/gsd-lifecycle.md`
- Verification: [rules/40](../../rules/40-verification-protocol.md)
