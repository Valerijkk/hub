---
name: verify-goal
description: Goal-backward приёмка — доказать исход фазы по success criteria из PLAN с evidence (файл/команда) и статусом регрессии. Использовать после wave-orchestrate или завершения execute, до handoff или merge-ready.
disable-model-invocation: true
---

# Верификация цели (`verify-goal`)

Оценивать **достижение исхода**, а не закрытие чекбоксов задач. У каждого success criterion должны быть прикреплённый evidence либо явный NOT_RUN.

## Когда использовать

- Волна/фаза заявлена complete — до handoff или merge-ready
- После цикла исправлений после FAIL
- Перед выставлением `verified: true` в отчётах или skills

## Когда остановиться

- Нет PLAN или success criteria → остановиться; запросить PLAN у `phase-plan`
- Для критерия не определён тип evidence → FAIL по этому критерию; не выводить PASS по догадке
- Verifier — единственный автор всех артефактов без независимой проверки → отметить в отчёте; при high-impact эскалировать по [rules/40](../../../prompt-engineering/rules/40-verification-protocol.md)

## Принцип

```
Goal (from PLAN)
  → Success criteria (falsifiable)
    → Evidence (path | command+log | doc | artifact)
      → Verdict PASS | FAIL + gaps
```

Закрытые задачи без evidence по критерию = **FAIL**.

## Контекст (загрузить сначала)

| Нужно | Primary | Legacy fallback |
|------|---------|-----------------|
| PLAN + criteria | Путь к phase PLAN | Список criteria от пользователя |
| Commands | `context/projects/<profile>/COMMANDS.md` | PROJECT-CONTEXT §4 |
| CI/regression | `context/projects/<profile>/CI-PIPELINE.md` | TESTFRAMEWORK-MAP |
| Docs sync matrix | Project playbook / правило пользователя | — |

## Шаги

1. Прочитать **goal** из PLAN (дословно), а не краткое резюме списка задач.
2. Перечислить success criteria; для каждого определить требуемый тип evidence (путь к коду, лог теста, секция документа, команда с exit 0).
3. **Собрать evidence** — открыть файлы, выполнить команды из profile (или пометить NOT_RUN с причиной). Не заявлять о прогонах без логов.
4. **Регрессия** — применить scope регрессии из PLAN; integration gate волны, если задан. Работа только с docs → N/A с обоснованием.
5. **Docs/context** — если критерии включают docs sync, проверить обновление путей или FAIL.
6. **Блокеры ревью** — открытые CRITICAL/HIGH из review без mitigation → общий FAIL.
7. **Verdict** — PASS только если все критерии PASS или задокументированный N/A; иначе FAIL + таблица gaps (owner, следующее действие).
8. **Цикл FAIL** — fix → повторный `verify-goal` (полный проход по критериям, не частичный).

## Evidence по типу критерия

| Тип критерия | Минимальный evidence |
|----------------|------------------|
| Behavior | тест существует + лог команды PASS + путь к коду |
| API contract | выравнивание doc/route; handler зарегистрирован |
| UI | путь к component/test + наблюдаемая проверка |
| Security | denial-тест + ссылка на constraint |
| Docs-only | валидные ссылки + чек-лист приёмки |
| Infra | diff конфига + sync deploy-дока |

При расхождении приоритет у project profile над этой таблицей.

## Шаблон отчёта

```markdown
## Goal verification — {{phase / wave}} — {{date}}

Verifier: {{role}} | Profile: {{profile}}

### Goal
<from PLAN>

### Criteria
| ID | Criterion | Evidence | Status |
|----|-----------|----------|--------|
| C1 | … | `path:line` / `cmd` → exit 0 / log excerpt | PASS/FAIL/N/A |

### Regression
- Scope: {{from PLAN/profile}}
- Status: PASS | FAIL | N/A — evidence: …

### Docs / context sync
- Status: PASS | FAIL | N/A — paths: …

### Review / security
- CRITICAL open: 0 | list
- HIGH open: 0 | list + mitigation?

### Verdict
PASS | FAIL

### Gaps (if FAIL)
| Gap | Severity | Owner | Next |
|-----|----------|-------|------|

Verification:
- Claims checked: each criterion evidence inspected
- Evidence: commands run / files read (list)
- Not run / N/A: …
- Risks / assumptions: …
- Mode: self-check | independent(author:X, verifier:Y)
- Verdict: PASS | FAIL | PARTIAL | NOT_RUN
```

`verified: true` только когда общий Verdict PASS и блок evidence заполнен.

## Handoff

- **PASS** → merge-ready / handoff по `playbooks/handoff-state.md`
- **FAIL** → владельцы gaps; `wave-orchestrate` или inline fix; повторить verify-goal
- Не делать handoff с PARTIAL без явного принятия владельцем

## Анти-паттерны

- PASS по одному зелёному локальному прогону без маппинга на критерии
- PASS при том, что wiki обещает поведение, которого нет в коде (drift)
- Пропуск регрессии из‑за «маленького изменения»

## Ссылки

- Planning: `../phase-plan/SKILL.md`
- Execution: `../wave-orchestrate/SKILL.md`
- Lifecycle: `../../playbooks/gsd-lifecycle.md`
- Verification protocol: [rules/40](../../../prompt-engineering/rules/40-verification-protocol.md)
