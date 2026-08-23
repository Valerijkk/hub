---
name: wave-run
description: Запуск волны Hub-агентов строго по шаблону team-lead §5.2 с картой файлов и rule 56. Используй после phase-plan, когда PLAN готов и нужно делегировать параллельных исполнителей.
disable-model-invocation: true
---

# wave-run

Запуск волны агентов Hub: планы готовы → промпты **только** по шаблону §5.2 → параллель с картой файлов → сбор → fixed Compose services / verifiers.
Исполнитель: **`team-lead`**.

## Когда

- После `phase-plan` (или эквивалентного PLAN с атомарными задачами)
- Повтор волны после triage FAIL (новый промпт §5.2 с уточнённым контекстом)

**Запрещено:** свободный текст «сделай X» субагенту (REGULATION / prompt-contract §2).

## Предусловия

- [ ] PLAN с целью, задачами, depends, картой файлов
- [ ] Для MEDIUM+ есть [`$task-packet`](../../task-packet/SKILL.md) или эквивалентный контракт
- [ ] Пересечение файлов внутри волны = **0**
- [ ] Один active writer на ownership boundary; reviewers/verifiers read-only
- [ ] Для каждого артефакта назначен `*-ver`
- [ ] Известно: кто гоняет доменные тесты, кто fixed Compose service на интеграции
- [ ] Явно заданы fan-out, delegation depth, retry/remediation limits и timebox

## Шаги

1. Взять task packet и задачи текущей волны из PLAN; проверить authority,
   stop conditions, карту файлов и execution bounds. LOW-задачу без независимых
   boundaries выполнить напрямую, не создавать волну.
2. Для **каждого** агента заполнить полный шаблон §5.2 (`team-lead.md` / `team-workflow.md` §5.3.1).
3. В `[ДОПОЛНИТЕЛЬНЫЕ УСЛОВИЯ]` **обязательно** оба контура rule 56 (новое + регрессия).
4. Запустить не больше PLAN `fan_out`; не превышать `delegation_depth` и
   `task_timebox`. Один writer владеет boundary до sync point.
5. Собрать артефакты; конфликт путей / противоречивые контракты / исчерпанный
   bound → **стоп / переплан**, а не скрытый дополнительный агент или retry.
6. Код: агент — доменные тесты; **team-lead** — все применимые fixed Compose services на интеграции волны. Docs-only: link-check + acceptance.
7. Handoff каждому `*-ver` **тоже по §5.2** (принимающий не видит чат оркестратора).
8. Remediation вернуть исходному writer только в finding scope. Hard bounds:
   максимум 3 fix attempts на один RCA и максимум 5 re-review rounds; раунды
   4–5 требуют нового подтверждённого RCA и нового/усиленного writer-а.
   CRITICAL/HIGH нельзя park для PASS; исчерпан cap → BLOCKED.
9. После verifiers → `goal-verify` → GSD-RU `handoff`.

## Канонический TaskContract — не дублировать

Единственный копируемый шаблон находится в `.cursor/agents/team-lead.md` §5.2;
нормативное зеркало — `.cursor/playbooks/team-workflow.md` §5.3.1. `wave-run`
не хранит сокращённую копию, потому что она неизбежно дрейфует.

Wave-specific overlay поверх канона обязан заполнить: PLAN task ID и dependencies,
точную ownership map, `fan_out`/`delegation_depth`/timebox, оба quality-контура
rule 56, writer→readonly verifier handoff, а также оба независимых remediation
counter. Ни одно из этих полей не заменяет остальные поля TaskContract.

## Worked example (Hub): делегирование be1 — bases

Полный заполненный промпт — `team-lead.md` §5.3. Краткая выжимка:

| Поле | Значение |
|------|----------|
| Роль | be1, backend Go/Fiber/Ent |
| Задача | закрыть gap: чужие bases видны/редактируемы (NF-03-аналог) |
| Файлы only | `backend/internal/bases/*.go`, genuine units `backend/tests/bases/**` |
| Не трогать | `ent/schema/base.go` |
| Ветка | `feat/bases/security-membership-fix` |
| Rule 56 new | тесты 403/404/list filter до green |
| Rule 56 regression | Docker `backend-unit` для `backend/tests/bases/**`; Python API authz job по применимым TC |
| Docs | `docs/wiki/API.md` |
| Verifier | be-ver |

**Исторический результат:** `authorizeBase` на всех путях; `98fb3df` +
`8ff10d7`; ветка стала merge-ready и была merged пользователем 2026-07-11.

## Параллель и лимиты

Каждая волна переносит из PLAN точные значения:

```text
fan_out=<N>; delegation_depth=<N>; retry_limit=<N/task>;
max_fix_attempts_per_root_cause=<N<=3>; max_re_review_rounds=<N<=5>;
task_timebox=<duration>;
on_limit=STOP|REPLAN|ESCALATE
```

Наличие свободных agent slots не повышает лимит и не расширяет authority.

**GZ-QA-13:** применимая fixed Compose service sequence запускается **один раз** на интеграции волны (team-lead), не у каждого агента.

## Чек-лист сборки волны

```markdown
## Wave sync
- [ ] Все агенты вернули формат §1–5 + Verification
- [ ] Нет пересечения изменённых файлов
- [ ] Один writer на ownership boundary; bounds не превышены
- [ ] Доменные тесты green (логи приложены)
- [ ] Docs-sync сделан или явно в next
- [ ] Промпты *-ver подготовлены по §5.2
- [ ] fixed Compose service (если код) — PASS или gaps зафиксированы
```

## DoD

- [ ] Нет свободного «сделай X» — только §5.2
- [ ] У каждого промпта: границы файлов + факты ≥2 + playbooks
- [ ] Rule 56 (новое + регрессия) в условиях
- [ ] Verifier назначен и получит отдельный §5.2
- [ ] Карта файлов волны без пересечений
- [ ] Fan-out/depth/retry/remediation/time заданы и соблюдены
- [ ] Известен владелец fixed Compose service
- [ ] Конфликты остановлены до merge-ready

## Anti-patterns

| Антипаттерн | Почему | Что вместо |
|-------------|--------|------------|
| be1 и be2 на один файл | merge hell | один владелец файла в волне |
| Волна без sync point | потеря конфликтов | явный sync после parallel |
| Приёмка без Verification блоков | «вроде ок» | требовать блок из prompt-contract §6 |
| Полный fixed-service прогон у каждого агента | расточительно (CI мин) | домен у агента + один service на интеграции |
| Handoff ver-у ссылкой на «чат выше» | ver не видит контекст | полный §5.2 |
| Смешать security-фикс с фичей | нельзя срочно смержить | отдельная ветка/PR |

## Связь

- Upstream: `gsd-execute-phase`, `gsd-workstreams`
- Hub: `team-lead.md` §5–§8, `prompt-contract.md` §2/§6/§13, rule 56
- Предыдущий: [phase-plan](../phase-plan/SKILL.md) · Следующий: [goal-verify](../goal-verify/SKILL.md)
- Индекс: [gsd-ru/README.md](../README.md)
