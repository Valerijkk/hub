---
name: handoff
description: Передаёт результат между ролями или волнами через outcome, evidence и следующий handoff. Используй после goal-verify, при смене роли или паузе; docs/state синхронизирует отдельный /handoff-sync.
disable-model-invocation: true
---

# handoff

Передавай между ролями, волнами или сессиями самодостаточный снимок результата.
Skill не выдаёт полномочия, не заменяет docs-sync и не просит получателя
восстанавливать контекст из чата.

## Когда

- исполнитель передаёт работу верификатору;
- верификатор возвращает PASS/FAIL оркестратору;
- завершённая волна разблокирует следующую;
- `team-lead` передаёт merge-ready результат пользователю;
- работа ставится на паузу с назначенным продолжателем.

## Минимальный контракт: 14 полей

| # | Поле | Содержание |
|---|------|------------|
| 1 | Task | ID/фаза/волна |
| 2 | Objective | исходная цель |
| 3 | User-visible outcome | что наблюдает пользователь |
| 4 | Status | DONE / PARTIAL / BLOCKED |
| 5 | Source revision | branch, full commit, dirty worktree |
| 6 | Actual artifacts | изменённые/созданные/проверенные пути |
| 7 | Owned scope | фактические ownership-границы |
| 8 | Non-goals | `do not touch` и сознательно исключённое |
| 9 | Decisions/assumptions | только существенные решения и допущения |
| 10 | Acceptance/oracle | критерии и точный expected result |
| 11 | Evidence | check → PASS / FAIL / NOT_RUN + причина |
| 12 | Runtime provenance | requested model/effort; actual telemetry или `UNKNOWN`; env/image |
| 13 | Risks/blockers | severity, owner, approval/stop boundary |
| 14 | Receiver/next | одна роль → одно конкретное действие |

Эти 14 полей обязательны для role/wave/session handoff. Прямой ответ
пользователю по LOW-задаче handoff не является: не добавляй церемонию без
получателя.

## Адаптация по получателю

| Передача | Что подчеркнуть |
|----------|-----------------|
| implementer → `*-ver` | контракт/AC, diff scope, команды регрессии |
| `*-ver` → `team-lead` | verdict, findings, воспроизводимое evidence |
| wave N → wave N+1 | разблокированный контракт и незакрытые зависимости |
| `team-lead` → пользователь | ветка/PR, merge-ready status; rebase-merge в `master` делает только пользователь |

Краткий формат сохраняет все поля:

```text
Task: <ID> | Objective: <цель> | User outcome: <наблюдаемое>
Status: <DONE|PARTIAL|BLOCKED> | Revision: <branch/full SHA/dirty>
Artifacts: <paths> | Owned: <only> | Non-goals: <do not touch>
Decisions/assumptions: <существенное или none>
Acceptance/oracle: <criterion -> expected>
Evidence: <check -> PASS|FAIL|NOT_RUN>
Runtime: requested=<...>; actual=<telemetry|UNKNOWN>; env=<...>
Risks/blockers: <severity; owner; approval/stop>
Receiver/next: <role -> one action>
```

## Docs и state

Если работа требует обновления docs или `.planning/STATE.md`, отдельно запусти
[`/handoff-sync`](../../handoff-sync/SKILL.md) по
[`rule 55`](../../../rules/55-docs-sync.mdc). Не копируй сюда матрицу docs-sync
или session-note template; в evidence укажи только результат их выполнения.

## DoD

- [ ] Outcome описывает факт, а не процесс
- [ ] Все 14 полей присутствуют и привязаны к exact revision
- [ ] Evidence позволяет воспроизвести verdict без исходного чата
- [ ] Непроведённое обозначено `NOT_RUN`, actual runtime не угадан
- [ ] Handoff называет одну принимающую роль и следующее действие
- [ ] Gaps/blockers имеют владельца
- [ ] Для пользовательского merge явно сохранена policy rule 50

Связь: [goal-verify](../goal-verify/SKILL.md) ·
[GSD-RU README](../README.md) ·
[team-workflow §5.3](../../../playbooks/team-workflow.md)
