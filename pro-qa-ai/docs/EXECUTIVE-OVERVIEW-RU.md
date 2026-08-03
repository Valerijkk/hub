# AI Quality Layer для автоматизации тестирования

## Решение в одном абзаце

`pro-qa-ai/` — переносимая прослойка между автоматизатором и AI. Она
изучает существующий тестовый фреймворк по указанным папкам, создаёт
верифицированный project context, применяет approved patterns и вручную
добавленный опыт коллег, направляет специализированных агентов и измеряет
качество относительно bare-prompt baseline.

Это не генератор каркаса тестового фреймворка и не библиотека «магических
промптов». Ответственность за требования, архитектуру, security и merge остаётся
у людей.

## Какую проблему решает

- модель выдумывает пути, команды и ожидания;
- сгенерированный тест зелёный, но не доказывает требование;
- ревью AI-кода повторяет одни и те же замечания;
- flaky «чинится» sleep/retry вместо root cause;
- знания о хорошем коде остаются в комментариях отдельных коллег;
- разные модели получают разный, неполный контекст;
- экономия времени заявляется без измерений.

## Как работает

```text
пути проекта/тестов
  → bootstrap generated context
  → human review
  → optional manual before/comment/after examples
  → approved pattern cards
  → recipe + role agent + catalog-backed model routing
  → evidence-first verification
  → paired eval + KPI
```

Bootstrap не читает закрытые MR, историю review или внешние системы.
Автоматизатор сам решает, какие комментарии коллег вручную добавить в
подготовленный шаблон. Project examples по умолчанию не покидают проект.

## Что получает команда

- воспроизводимую карту framework/CI/test data/traceability;
- каталог проверенных patterns с provenance/freshness;
- 14 production recipes поверх 1000 prompt fragments;
- GSD-подобный lifecycle plan → execute → verify → handoff;
- Sol Pro orchestrator и динамический выбор доступных Cursor models;
- структурный валидатор и протокол парного замера;
- governance, approval и data-handling policy.

## Текущий статус

- Source и packaged validators: **PASS**, 0 errors / 0 warnings.
- Combined security/tooling tests: **42 PASS / 1 environmental SKIP** (43 total;
  mocked Windows reparse checks PASS).
- Portable package: **162 files**, Hub profile/smoke/raw inbox исключены.
- Eval effectiveness/ROI: **NOT_RUN**.
- Release: **pre-release**, adoption blocked до измеримого пилота и sign-off.

Structural PASS означает, что контракты и ссылки согласованы. Он не доказывает
рост качества или ROI.

Model routing настроен, но runtime provenance пока **PARTIAL**: для демо или
измеримого run нужно сохранить живой catalog snapshot, фактический model ID,
fallback и tool trace. Конфигурация не доказывает, что выбранная модель реально
исполняла задачу.

## Пилот

1. Выбрать 12–24 типовые задачи автоматизаторов.
2. Для каждой зафиксировать модель, права, репозиторий, вход и критерий оценки.
3. Сравнить bare-prompt и assisted режим в изолированных сессиях.
4. Отдать результаты слепому ревьюеру.
5. Измерить first-pass acceptance, review minutes, defects, hallucinations,
   scope escapes, traceability, flaky risk, false positives, tokens/tools/latency.
6. Принять решение: ADOPT, LIMITED_ADOPT, ITERATE или STOP.

## Критерии безопасного внедрения

- ноль scope escapes и утечек данных;
- critical/high defects не растут;
- каждый факт/команда имеет source или `UNKNOWN`/`NOT_RUN`;
- project examples прошли redaction/provenance/human approval;
- destructive/external/security действия остаются за человеком;
- хотя бы один целевой KPI улучшен без материальной деградации остальных.

Утверждение «ускоряет в N раз» допускается только после такого пилота.
