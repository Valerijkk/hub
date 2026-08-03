# Playbook: Portable kit dry-run

> Процесс проверки переноса AI kit в **пустой** (или новый) monorepo без копирования product-кода.

## Когда

- Перед реальным переносом kit в другой репозиторий
- После обновления L1–L5 артефактов (регрессия переносимости)
- Закрытие gap «dry-run» из AI-expansion acceptance

## Локальный канон

Этот файл содержит переносимый smoke-контракт и не зависит от planning-путей
репозитория-источника. Дополнительный installation guide может быть добавлен
adapter-пакетом целевого репозитория.

## Конвейер (кратко)

```
1. Preflight: источник, чистая цель, allowed paths, no secrets
2. Copy portable core; выбрать ровно один runtime adapter
3. skills/bootstrap-test-context/SKILL.md — каждый факт имеет evidence или `UNKNOWN`
4. Render adapter wrappers; raw core agents не копировать как runtime config
5. Populate models registry from actual runtime catalog
6. Пробная docs-only волна с explicit files/approval/stop
7. Goal-verify + link/permission/model fallback checks
8. STATE/handoff note с kit revision и результатом
```

## Гейты

| Гейт | Владелец | Критерий |
|------|----------|----------|
| Sanitize | docs-writer | нет секретов / нет source-repo facts в target context |
| Bootstrap | любой + Verification | Profile report + `CONTEXT-MANIFEST.yaml` |
| Smoke wave | orchestrator | шаблон делегирования + Verification |
| Routing | orchestrator | actual catalog + decision/fallback log |
| Accept | orchestrator / docs-writer | Must DoD PASS |

## Участники

`docs-writer` (ведущий) · `orchestrator` (пробная волна) · опционально `devops` (rsync/sandbox).

## DoD

- [ ] Context cards contain only target-repository facts or literal `UNKNOWN`
- [ ] Runtime adapter is installed and permissions smoke-tested
- [ ] Current catalog snapshot resolves orchestrator default or logs fallback
- [ ] Downstream wrapper defaults to inherit unless a decision log overrides it
- [ ] Docs-only wave changes only delegated files
- [ ] Verification, routing log and handoff are self-contained

## Связь

- GSD-порт: [`../gsd-port/README.md`](../gsd-port/README.md)
- Handoff: [`handoff-state.md`](handoff-state.md)
- Wave: [`wave-execution.md`](wave-execution.md)
- Adapters: [`../adapters/`](../adapters/)
- Model routing: [`model-routing.md`](../playbooks/model-routing.md)
