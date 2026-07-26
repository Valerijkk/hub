---
name: phase-plan
description: Планирование фазы Hub — goal, atomic plans, depends, success criteria и волны. Используй до первого wave-run для новой фичи или переплана после goal-verify FAIL.
disable-model-invocation: true
---

# phase-plan

Планирование фазы Hub: проверяемая цель → атомарные планы → волны → риски → verification owners.
Исполнитель: **`tech-lead`**. Docs-only / оркестрация без кода: **`team-lead`**.

## Когда

- Новая фича/фаза до первого `wave-run`
- После `gsd-discuss-phase` / SPEC, когда scope зафиксирован
- Переплан после FAIL goal-verify или конфликта файлов в волне

**Не использовать**, если задача уровня `codex-micro` / `gsd-fast` (одна правка, один файл).

## Входы

| Вход | Источник |
|------|----------|
| Goal / outcome | ROADMAP, запрос пользователя, SPEC.md |
| Границы Won't | product / TZ / ADR |
| Контекст кода | `.planning/STATE.md`, docs/* |
| Playbooks | `team-workflow.md` §6, `prompt-contract.md` |
| Канон делегирования | `.cursor/agents/team-lead.md` §5.2 |

## Шаги

1. **Goal одним предложением** — проверяемый outcome (не «улучшить X»).
2. **Success criteria** — каждый пункт = команда, путь файла или измеримый статус.
3. **Нарезать атомарные планы:** один артефакт / один слой / ориентир ≤200–300 строк diff.
4. **Признак «слишком крупно»:** союз «и» между артефактами; BE+FE в одном плане; схема Ent + handler вместе; security-фикс + фича в одной ветке.
5. **Таблица задач:** # | задача | исполнитель | файлы (only/не трогать) | Depends | оценка | `*-ver`.
6. **Волны:** параллелить только независимое; карта файлов внутри волны без пересечений.
7. **Риски** + mitigation; security/auth → отдельная ветка (см. team-lead §11.1).
8. **Self-check** prompt-contract / rules/40 этапы 1–4 на полноту плана (scope, depends, критерии, owners).
9. Записать PLAN в `.planning/phases/<id>/PLAN.md` или эквивалент волны.

## Полный шаблон PLAN

```markdown
# PLAN: <фаза / волна>

## Цель
<одно предложение, проверяемый outcome>

## Won't (вне scope)
- …

## Depends on
- [ ] ADR / предыдущая фаза / артефакт: `path`

## Success criteria
- [ ] C1 — evidence: команда `…` или файл `…`
- [ ] C2 — …
- [ ] Docs-sync: STATE + domain doc (rule 55)

## Задачи
| # | Задача | Исполнитель | Файлы only | Не трогать | Depends | Оценка | Verifier |
|---|--------|-------------|------------|------------|---------|--------|----------|
| 1 | … | be1 | `backend/internal/X/` | `ent/schema/` | — | S | be-ver |

## Волны
- Wave 1 (parallel): #1, #2 — карта файлов: …
- Sync point: сбор артефактов + conflict check
- Wave 2: #3 (depends Wave1)

## Quality gates (rule 56)
- Новый контур: implement → тест → triage loop до green
- Регрессия домена: …
- `/qa-gate`: на интеграции волны — team-lead; одиночная работа — исполнитель

## Риски
| Риск | Impact | Mitigation |
|------|--------|------------|
| … | H/M/L | … |

## Verification
- owners: be-ver / fe-ver / …
- `verified: true` только при PASS + лог

## Handoff
- Следующий skill: `wave-run` → `goal-verify` → `handoff`
```

## Worked example (Hub): bases security fix

**Исторический контекст:** Волна 1, 2026-07-03 — аудит: любой auth-юзер
видит/правит все базы (аналог NF-03); результат merged пользователем
2026-07-11. Канон промпта: `team-lead.md` §5.3.

```markdown
# PLAN: bases owner-scope security

## Цель
Чужая Hub Base недоступна (list скрывает, GET/мутации → 403; missing → 404); владелец без регрессии.

## Won't
- Не менять ent/schema base
- Не смешивать с остальным bases-эпиком (отдельная ветка)

## Success criteria
- [ ] authorizeBase на всех read/mutate путях bases
- [ ] go test -tags hubtest -race ./testframework/bases/... PASS
- [ ] docs/wiki/API.md — owner-scoped доступ
- [ ] ветка feat/bases/security-membership-fix merge-ready

## Задачи
| # | Задача | Исп. | Файлы | Depends | Verifier |
|---|--------|------|-------|---------|----------|
| 1 | authorizeBase + ListBases filter | be1 | internal/bases/*, testframework/bases/* | — | be-ver |
| 2 | Docs API.md | docs-writer или be1 | docs/wiki/API.md, STATE.md | #1 | — |

## Волны
- Wave 1: #1 → sync → #2 (docs в той же сессии, rule 55)
```

**Факт:** коммиты `98fb3df` (fix) + `8ff10d7` (RBAC denial suite); эталон цикла — rule 56 кейс 1.

## Карта слоёв → роли

| Слой | Исполнитель | Verifier |
|------|-------------|----------|
| BE handler/service | be1 / be2 | be-ver |
| FE UI | fe1 / fe2 | fe-ver |
| TC ручные | qa-manual | qa-manual-ver |
| Автотесты | qa-auto1 / qa-auto2 | qa-auto-ver |
| Security boundary | security-auditor | (independent) |
| Arch / ADR | arch | arch-ver |
| Pre-merge diff | — | code-reviewer |

## DoD

- [ ] Goal — одно предложение, проверяемый outcome
- [ ] Нет задачи с двумя слоями (BE+FE, schema+код)
- [ ] У каждой задачи — success criteria + verifier
- [ ] Карта файлов внутри волны без пересечений
- [ ] Won't scope зафиксирован
- [ ] Rule 56 и `/qa-gate` ownership прописаны
- [ ] Нет depends-циклов
- [ ] Security/auth в отдельной ветке при необходимости

## Anti-patterns

| Антипаттерн | Почему плохо | Что вместо |
|-------------|--------------|------------|
| «Реализовать модуль целиком» | Неатомарно, непараллелится | Нарезать по слоям/артефактам |
| Success criteria «стало лучше» | Непроверяемо | Команда / путь / код ответа |
| Depends cycles | Волна зависает | Перестроить порядок |
| BE и FE в одном плане | Конфликт ownership | Два плана, sync point |
| План без Won't | Scope creep | Явный список вне scope |
| Свободный «сделай X» вместо волны | Нет границ для §5.2 | Сначала PLAN, потом wave-run |

## Связь

- Upstream: `gsd-plan-phase`, `gsd-analyze-dependencies`
- Hub: `tech-lead.md`, `team-workflow.md` §6, `team-lead.md` §5.2
- Следующий skill: [wave-run](../wave-run/SKILL.md)
- Индекс: [gsd-ru/README.md](../README.md)
