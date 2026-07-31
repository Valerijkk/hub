# Плейбуки поставки — Touchstone Pro

Процессы разработки продукта. Плейбуки тестового фреймворка
(onboarding, test-generation, review-cycle, debugging, documentation)
остались в Standard и отсюда не дублируются.

| Плейбук | Поток | Роли |
|---------|-------|------|
| [corporate-delivery.md](corporate-delivery.md) | scope → design → build → test → review → verify → handoff | orchestrator, весь ростер |
| [gsd-lifecycle.md](gsd-lifecycle.md) | discuss → plan → execute → verify → handoff (порт GSD) | orchestrator, все роли |
| [wave-execution.md](wave-execution.md) | Маршрутизация моделей в волне и evidence выбора | orchestrator |
| [handoff-state.md](handoff-state.md) | STATE, session notes, передача между агентами | orchestrator, любой |
| [portable-kit-dry-run.md](portable-kit-dry-run.md) | Dry-run переноса kit (L1–L5 в пустой monorepo) | docs-writer, orchestrator |
| [discuss-and-research.md](discuss-and-research.md) | Что делать ДО плана: серые зоны, допущения, ресёрч, спайк | orchestrator, architect, product-owner |
| [debug-forensics.md](debug-forensics.md) | Отладка дефекта и разбор провала прогона агента | debugger, orchestrator |

Выбор модели — [`model-routing.md`](../../prompt-engineering/playbooks/model-routing.md) в Standard:
маршрутизация одна на оба издания, дублировать её нельзя.

## Правила конвейеров

1. **Волна не стартует без непересекающегося scope.** Два агента на одном
   файле — не параллелизм, а гонка. Разрез по файлам делает orchestrator
   до запуска, а не по ходу.
2. **Verify ≠ зелёный CI.** CI необходим, но не достаточен: приёмка идёт
   goal-backward от критерия фазы — [`../layers/02-gsd-max.md`](../layers/02-gsd-max.md).
3. **Handoff самодостаточен.** Следующий агент не видит чат. Всё, что нужно,
   лежит в артефакте, а не «как договорились выше».
4. **Фикс с тестом закрывается мутацией** — [`../extracts/mutation-evidence.md`](../extracts/mutation-evidence.md).
