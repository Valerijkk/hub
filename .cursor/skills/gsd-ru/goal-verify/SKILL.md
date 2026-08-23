---
name: goal-verify
description: Goal-backward приёмка фазы — проверяет достижение цели по evidence, rule 56 и rule 55. Используй после wave-run, до handoff и verified:true.
disable-model-invocation: true
---

# goal-verify

Goal-backward приёмка: проверяй **достижение цели**, а не закрытие чекбоксов задач.
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
      → Verdict PASS | FAIL | NOT_RUN + gaps
```

Закрытые таски без выполненного evidence по критериям = **NOT_RUN**; criterion,
опровергнутый oracle, = **FAIL**. Ни одно из них не допускает PASS.

## Иерархия oracle

Выбирай сверху вниз сильнейший применимый источник:

1. trusted invariant/contract с точным expected result;
2. focused executable test, type/static/security check или browser assertion;
3. integration/recovery exercise на затронутой границе;
4. независимая экспертная read-only проверка для неисполняемой семантики;
5. self-critique только как источник гипотез, никогда как completion proof.

Количество reviewers не компенсирует отсутствующий oracle. Для независимой
проверки используй [`$evidence-verifier`](../../evidence-verifier/SKILL.md).

## Шаги

1. Прочитать exact task contract (`$task-packet`/PLAN) и **goal** (не список задач).
2. Выписать success criteria; для каждого — требуемый тип evidence.
3. Зафиксировать full source revision, worktree, `BASE..HEAD` diff и runtime
   provenance: requested model/effort, actual telemetry или `UNKNOWN`, env/image digest.
4. Собрать evidence: путь к текущему коду, команда + exit code, лог, ссылка на doc.
5. **Rule 56:** новое green + регрессия домена; на интеграции — все применимые
   fixed services, каждый через
   `docker compose -f docker-compose.local.yml run --rm <exact-service>`.
   Невыполненная обязательная проверка = `NOT_RUN`, не PASS.
6. **Rule 55:** STATE session note + domain docs по матрице.
7. Открытый CRITICAL/HIGH = `FAIL`; его нельзя park или понизить ради PASS.
8. Вердикт PASS/FAIL/NOT_RUN. `verified: true` только независимым verifier при PASS.
9. Remediation ограничить finding scope и bounds PLAN; затем повторить проверку.

Итог: любой required `FAIL` → `FAIL`; без FAIL, но с required `NOT_RUN` →
`NOT_RUN`; только все required `PASS` → `PASS`.

## Полный шаблон отчёта

```markdown
## Goal-verify: <фаза / волна>
Дата: YYYY-MM-DD
Проверяющий: team-lead | be-ver | …
Revision: <full SHA>; diff: <BASE>..<HEAD>; worktree: <state>
Runtime: requested=<model/effort>; actual=<telemetry|UNKNOWN>; env=<image digest|UNKNOWN>

### Goal
<дословно из PLAN>

### Criteria
| ID | Criterion | Evidence | Status |
|----|-----------|----------|--------|
| C1 | … | `path:line` / `cmd` → exit 0 / лог | PASS/FAIL/NOT_RUN |
| C2 | … | … | PASS/FAIL/NOT_RUN |

### Rule 56 — качество
- Новый контур: PASS | FAIL | NOT_RUN — evidence: …
- Регрессия домена: PASS | FAIL | NOT_RUN — evidence: …
- fixed Compose service (интеграция волны): PASS | FAIL | NOT_RUN — evidence: …

### Rule 55 — docs-sync
- STATE.md: PASS | FAIL | NOT_RUN — секция/дата
- Domain docs: PASS | FAIL | NOT_RUN — список путей
- Drift check (wiki ≰ код): PASS | FAIL | NOT_RUN

### Security / review
- CRITICAL open: 0 | список
- HIGH open: 0 | список

### Verdict
PASS | FAIL | NOT_RUN

### Gaps (если FAIL)
| Gap | Severity | Owner | Next |
|-----|----------|-------|------|
| … | CRITICAL/HIGH/MED | be1 | wave-run #… |

### Verification (self)
- [ ] Evidence проверен по факту (файл открыт / команда в логе)
- [ ] Не ставил PASS по «выглядит ок»
- [ ] Revision/oracle/runtime provenance указаны; actual model не угадан
```

## Worked example (Hub): bases security — current acceptance

**Goal:** чужая Hub Base недоступна; владелец без регрессии.

| Criterion | Evidence | Status |
|-----------|----------|--------|
| authorizeBase на путях | current code inspection на exact revision | NOT_RUN — exact revision не приложена |
| ListBases по владельцу | current code + target tests | NOT_RUN — current target evidence отсутствует |
| Denial suite 403/404 | archived pre-cutover evidence at commit `8ff10d7`; не target-path proof | HISTORICAL_CONTEXT — не acceptance evidence |
| Current target regression | Docker-only `backend/tests/bases/**` + applicable `testframework/tests/api/bases/**` | NOT_RUN until greenfield cutover |
| Docs | `docs/wiki/API.md` owner-scoped; STATE session note | NOT_RUN — exact revision не приложена |

**Current overall verdict:** `NOT_RUN`. Ни один исторический лог или merge-факт не
может сделать текущую ревизию PASS.

**Исторический контекст, не текущая приёмка:** ветка
`feat/bases/security-membership-fix` была признана merge-ready и смержена
пользователем 2026-07-11; denial evidence хранится на commit `8ff10d7`.

## Worked example (отрицательный)

| Criterion | Evidence | Verdict impact |
|-----------|----------|----------------|
| «Тесты написаны» | файлы есть, прогон не приложен | FAIL — нет лога |
| «API обновлён» | wiki обещает endpoint, в RegisterRoutes нет | FAIL — drift |
| Security HIGH open | code-reviewer finding без фикса | FAIL — блок PASS |

## Матрица evidence по типу критерия

| Тип критерия | Минимальный evidence |
|--------------|----------------------|
| Backend поведение | applicable `backend-unit`/Python target job + путь кода + exact Docker evidence |
| Frontend UX | `frontend-unit` или Python E2E target job + путь компонента + exact Docker evidence |
| API контракт | wiki/API.md + handler route + (желательно) swagger |
| Security invariant | denial-тест + ссылка на guideline/TZ |
| Docs-only | link-check + acceptance checklist |
| Infra | diff compose/ci + DEPLOY.md sync |

## DoD

- [ ] Goal процитирован, не подменён списком задач
- [ ] Каждый criterion имеет evidence (не мнение)
- [ ] Exact revision, `BASE..HEAD` diff и oracle указаны
- [ ] Requested/actual runtime записаны; missing telemetry = UNKNOWN
- [ ] Rule 56 закрыт; непроведённое явно `NOT_RUN`
- [ ] Rule 55 закрыт
- [ ] CRITICAL/HIGH gaps = 0 для PASS; park запрещён
- [ ] `verified: true` только при PASS + лог

## Anti-patterns

| Антипаттерн | Почему | Что вместо |
|-------------|--------|------------|
| «Выглядит ок» | нет evidence | команда/путь/лог |
| PASS при открытом HIGH security | прод-риск | FAIL до закрытия finding |
| Считать таски = цель | goal drift | goal-backward |
| One-shot зелёный → verified | rule 56 запрещает | цикл + регрессия |
| Wiki обещает несуществующее | drift | сверка с кодом |
| Скрыть непрогнанную регрессию | ложный PASS | `NOT_RUN` с причиной и владельцем |

## Связь

- Upstream: `gsd-verify-work`, `gsd-audit-uat`
<!-- hub.qa-routing-declaration/v1
{"schema_version":"hub.qa-routing-declaration/v1","scenario_id":"selftest-markers","assigned_skill":".cursor/skills/gsd-ru/goal-verify/SKILL.md","declared_routes":["testframework/selftests/pytest_plugin/test_markers.py","testframework/src/hub_testkit/pytest_plugin/markers.py"],"expected_routes":["testframework/selftests/pytest_plugin/test_markers.py","testframework/src/hub_testkit/pytest_plugin/markers.py"],"prohibited_legacy_tokens":["testframework/conftest.py","testframework/core/**"]}
-->
- Hub: `playbooks/verification.md`, prompt-contract §6, rule 56, rule 55
- Предыдущий: [wave-run](../wave-run/SKILL.md) · Следующий: [handoff](../handoff/SKILL.md)
- Индекс: [gsd-ru/README.md](../README.md)
