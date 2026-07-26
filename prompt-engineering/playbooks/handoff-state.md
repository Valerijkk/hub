# Playbook: Handoff & State

## Цель

Сохранить контекст между сессиями/агентами без утечки «памяти чата».

## Шаги

1. Записать session note в `{{state_file}}` (по умолчанию `.planning/STATE.md` проекта).
2. Синхронизировать docs по матрице проекта.
3. Собрать handoff-контракт: outcome, роль, факты, context/pattern cards,
   files, DoD/stop и approval boundary.
4. Приложить model decision/fallback log и actual resolved model.
5. Перечислить blockers / next.

## Шаблон state note

```markdown
## Session note — {{date}} ({{topic}})
- Done: {{done}}
- Files: {{files}}
- Checks: {{checks}}
- Routing/fallback: {{routing_log}}
- Approval status: {{approval_status}}
- Next: {{next}}
```

## Чек-лист

- [ ] Note понятен с нуля
- [ ] Нет секретов
- [ ] Next actionable
- [ ] Actual model/fallback и approval status не скрыты

## Антипаттерны

- «Продолжи с того места» без state
- Handoff ссылается на «как обсуждали»
