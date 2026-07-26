---
name: goal-verify
description: Goal-backward приёмка фазы — проверяет достижение цели по evidence, rule 56 и rule 55. Используй после wave-run, до handoff и verified:true.
disable-model-invocation: true
---

# goal-verify

Goal-backward приёмка: проверяем **достижение цели**, а не закрытие чекбоксов задач.
Исполнитель: **`team-lead`** (финальная), профильные **`*-ver`**, при продуктовом исходе — **`product-ver`**.

## Когда

- Фаза/волна «вроде готова», до handoff пользователю / merge-ready
- После серии фиксов triage — повтор до PASS
- Перед `verified: true` в skill/отчёте

**Не путать** с unit-прогоном исполнителя: goal-verify смотрит на goal + evidence + docs + регрессию целиком.

## Принцип

```
Goal (outcome)
  → Success criteria (проверяемые)
    → Evidence (путь / команда / лог / скрин)
      → Verdict PASS | FAIL + gaps
```

Закрытые таски без evidence по критериям = **FAIL**, даже если «всё зелёное локально один раз».

## Шаги

1. Прочитать **goal** из PLAN / ROADMAP (не список задач).
2. Выписать success criteria; для каждого — требуемый тип evidence.
3. Собрать evidence: путь к коду, команда + exit code, фрагмент лога, ссылка на doc.
4. **Rule 56:** новое green (цикл) + регрессия домена; на интеграции волны — статус `/qa-gate`. Docs-only → N/A с обоснованием.
5. **Rule 55:** STATE session note + domain docs по матрице.
6. Security/HIGH из ревью: открытый CRITICAL/HIGH без mitigation → блок PASS.
7. Вердикт PASS/FAIL + gaps. `verified: true` **только** при PASS + приложенный лог.
8. FAIL → triage → точечный `wave-run` / fix → повтор goal-verify (цикл).

## Полный шаблон отчёта

```markdown
## Goal-verify: <фаза / волна>
Дата: YYYY-MM-DD
Проверяющий: team-lead | be-ver | …

### Goal
<дословно из PLAN>

### Criteria
| ID | Criterion | Evidence | Status |
|----|-----------|----------|--------|
| C1 | … | `path:line` / `cmd` → exit 0 / лог | PASS/FAIL |
| C2 | … | … | PASS/FAIL |

### Rule 56 — качество
- Новый контур: PASS | FAIL — evidence: …
- Регрессия домена: PASS | FAIL | N/A — evidence: …
- /qa-gate (интеграция волны): PASS | FAIL | N/A — evidence: …

### Rule 55 — docs-sync
- STATE.md: PASS | FAIL — секция/дата
- Domain docs: PASS | FAIL — список путей
- Drift check (wiki ≰ код): PASS | FAIL

### Security / review
- CRITICAL open: 0 | список
- HIGH open: 0 | список + mitigation?

### Verdict
PASS | FAIL

### Gaps (если FAIL)
| Gap | Severity | Owner | Next |
|-----|----------|-------|------|
| … | CRITICAL/HIGH/MED | be1 | wave-run #… |

### Verification (self)
- [ ] Evidence проверен по факту (файл открыт / команда в логе)
- [ ] Не ставил PASS по «выглядит ок»
```

## Worked example (Hub): bases security

**Goal:** чужая Hub Base недоступна; владелец без регрессии.

| Criterion | Evidence | Status |
|-----------|----------|--------|
| authorizeBase на путях | `backend/internal/bases/service.go` — `authorizeBase()` | PASS |
| ListBases по владельцу | тот же service + тесты list | PASS |
| Denial suite 403/404 | `backend/testframework/bases/rbac_test.go` (`8ff10d7`) | PASS |
| Регрессия bases | `go test -tags hubtest -race ./testframework/bases/...` PASS | PASS |
| Docs | `docs/wiki/API.md` owner-scoped; STATE session note | PASS |

**Исторический verdict:** PASS → merge-ready ветка
`feat/bases/security-membership-fix`; merged пользователем 2026-07-11.
Эталон цикла: `.cursor/rules/56-regression-verify.mdc` кейс 1; промпт — `team-lead.md` §5.3.

## Worked example (отрицательный)

| Criterion | Evidence | Verdict impact |
|-----------|----------|----------------|
| «Тесты написаны» | файлы есть, прогон не приложен | FAIL — нет лога |
| «API обновлён» | wiki обещает endpoint, в RegisterRoutes нет | FAIL — drift |
| Security HIGH open | code-reviewer finding без фикса | FAIL — блок PASS |

## Матрица evidence по типу критерия

| Тип критерия | Минимальный evidence |
|--------------|----------------------|
| Backend поведение | тест + команда hubtest + путь кода |
| Frontend UX | vitest/RTL или TC + путь компонента |
| API контракт | wiki/API.md + handler route + (желательно) swagger |
| Security invariant | denial-тест + ссылка на guideline/TZ |
| Docs-only | link-check + acceptance checklist |
| Infra | diff compose/ci + DEPLOY.md sync |

## DoD

- [ ] Goal процитирован, не подменён списком задач
- [ ] Каждый criterion имеет evidence (не мнение)
- [ ] Rule 56 закрыт или явно N/A
- [ ] Rule 55 закрыт
- [ ] CRITICAL gaps = 0 для PASS
- [ ] HIGH: 0 или явный accepted risk с владельцем
- [ ] `verified: true` только при PASS + лог

## Anti-patterns

| Антипаттерн | Почему | Что вместо |
|-------------|--------|------------|
| «Выглядит ок» | нет evidence | команда/путь/лог |
| PASS при открытом HIGH security | прод-риск | FAIL или accepted risk в ADR |
| Считать таски = цель | goal drift | goal-backward |
| One-shot зелёный → verified | rule 56 запрещает | цикл + регрессия |
| Wiki обещает несуществующее | drift | сверка с кодом |
| N/A регрессии «потому что лень» | скрытый регресс | либо прогон, либо обоснованный N/A |

## Связь

- Upstream: `gsd-verify-work`, `gsd-audit-uat`
- Hub: `playbooks/verification.md`, prompt-contract §6, rule 56, rule 55
- Предыдущий: [wave-run](../wave-run/SKILL.md) · Следующий: [handoff](../handoff/SKILL.md)
- Индекс: [gsd-ru/README.md](../README.md)
