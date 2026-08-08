# Hub — Model Routing: качество, стоимость и усилие

> Канон runtime-выбора для AI-команды Hub. Ролевые файлы остаются
> переносимыми (`model: inherit`), а конкретные модель и effort выбирает
> `team-lead` при запуске и фиксирует в handoff.
>
> Источники на 2026-08-07:
> [Codex models](https://learn.chatgpt.com/docs/models) ·
> [Codex pricing](https://learn.chatgpt.com/docs/pricing) ·
> ADR: [`.planning/decisions/ADR-MODELS-GPT56-2026.md`](../../.planning/decisions/ADR-MODELS-GPT56-2026.md)

---

## 1. Базовый принцип

Использовать **самую дешёвую модель и минимальное effort, которые надёжно
закрывают completion bar**. Цена ошибки важнее цены запуска:

- неоднозначность, архитектура, security, prod и финальная приёмка → **Sol**;
- обычная разработка, анализ, тестирование и ревью → **Terra**;
- чёткая механическая работа с эталоном результата → **Luna**;
- near-instant итерация в одном файле → предпочтительно **Codex Spark**, если он
  доступен; это preview, не verifier и не модель для критичных решений.

`Ultra` — режим оркестратора для большой задачи, которую действительно можно
разделить на независимые части. Это не обычный effort субагента и не дефолт.

## 2. Экономика моделей

Snapshot flexible-pricing на 2026-08-07, credits за 1M токенов:

| Модель | Input | Cached input | Output | Относительно Sol | Решение Hub |
|---|---:|---:|---:|---:|---|
| GPT-5.6 Sol | 125 | 12.5 | 750 | 1.00× | Самые сложные и дорогие по цене ошибки задачи |
| GPT-5.6 Terra | 50 | 5 | 300 | 0.40× | Основная рабочая модель команды |
| GPT-5.6 Luna | 5 | 0.5 | 30 | 0.04× | Массовые и хорошо формализованные задачи |
| GPT-5.5 | 125 | 12.5 | 750 | 1.00× | Не назначать новым агентам: прошлое поколение по цене Sol |
| GPT-5.4 | 62.5 | 6.25 | 375 | 0.50× | Не назначать: retirement в Codex 2026-08-31 |
| GPT-5.4 Mini | 18.75 | 1.875 | 113 | ~0.15× | Не назначать: retirement в Codex 2026-08-31 |
| GPT-5.3 Codex Spark | preview | preview | preview | не опубликовано | Только интерактивная micro-итерация, fallback Luna |

Практический вывод: Terra стоит 40% Sol, Luna — 4% Sol и 10% Terra при той
же структуре ставок input/output. Поэтому Sol не используется для рутинной
реализации, а legacy 5.5/5.4/5.4 Mini не дают выгодного нового default.

Перед крупной волной `team-lead` повторно сверяет live catalog и pricing:
доступность и тарифы — внешнее изменяемое состояние, этот snapshot не заменяет
проверку при запуске.

## 3. Матрица 32 ролей

| Роли | Default runtime | Повысить до | Почему |
|---|---|---|---|
| `team-lead` | **Sol / Extra High** | Sol / Ultra для делимой большой волны | Цель, конфликты, стоимость ошибок оркестрации |
| `tech-lead` | **Sol / High** | Sol / Extra High для cross-domain графа | Зависимости и атомарная декомпозиция |
| `product` | **Sol / High** | Sol / Extra High для необратимого scope-решения | Высокая бизнес-цена ошибки |
| `product-ver` | **Terra / High** | Sol / High для milestone gate | Независимая проверка дешевле автора |
| `ba1`, `ba2` | **Terra / Medium** | Terra / High при конфликте персон/KPI | Повседневный анализ с понятным форматом |
| `ba-ver` | **Terra / High** | Sol / High для спорного продуктового решения | Challenge-pass и трассируемость |
| `sa1`, `sa2` | **Terra / High** | Sol / High для cross-domain API/ERD | Контракты требуют сильного reasoning |
| `sa-ver` | **Terra / High** | Sol / High для auth/data migration | Независимая проверка контрактов |
| `arch` | **Sol / Extra High** | — | Архитектура, ADR, стек, долгий горизонт решения |
| `arch-ver` | **Terra / Extra High** | Sol / Extra High для prod/security ADR | Независимый challenge; Sol только по риску |
| `designer` | **Terra / High** | Sol / High для новой дизайн-системы | Сильный everyday synthesis и tool use |
| `designer-ver` | **Terra / High** | Sol / High для критичного user flow | WCAG, состояния, консистентность |
| `be1`, `be2` | **Terra / Medium** | Terra / High; Sol / High для auth/crypto/migration | Основной production coding без переплаты Sol |
| `be-ver` | **Terra / High** | Sol / Extra High для security/data-loss | Независимое воспроизведение и race/security |
| `fe1`, `fe2` | **Terra / Medium** | Terra / High для сложного state/realtime | Основной React/TS coding |
| `fe-ver` | **Terra / High** | Sol / High для auth/a11y release gate | Hooks, a11y, типы, регрессия |
| `qa-manual` | **Terra / Medium** | Terra / High для cross-domain test design | Нужны judgment и трассируемость, не Sol по умолчанию |
| `qa-manual-ver` | **Terra / High** | Sol / High для release-critical UAT | Проверка полноты и denial-сценариев |
| `qa-auto1`, `qa-auto2` | **Terra / Medium** | Terra / High для flaky/concurrency/E2E infra | Тестовый код — production-grade работа |
| `qa-auto-ver` | **Terra / High** | Sol / High для mutation/coverage gate | Детерминизм и доказательная сила теста |
| `devops-engineer` | **Terra / High** | Sol / Extra High для prod/deploy/destructive | Инфра требует точности и rollback |
| `security-auditor` | **Sol / Extra High** | — | Самая высокая цена пропущенной уязвимости |
| `debugger` | **Terra / High** | Sol / Extra High для неизвестного cross-domain RCA | Terra для обычного RCA, Sol для тяжёлой неопределённости |
| `code-reviewer` | **Terra / High** | Sol / Extra High для auth/security/migration | Независимый pre-merge challenge |
| `docs-writer` | **Luna / Medium** | Terra / Medium для cross-domain docs | Дешёвый точный sync по заданным источникам |
| `codex-micro` | **Spark / Light**, если доступен | Luna / Light; Terra при сомнении | Typo, rename, one-liner; строгий узкий scope |
| `team-lead-opus` | deprecated | — | Не operational slot |

`Extra High` в таблице соответствует «Очень высокий» в приложении. Если
конкретный launcher не поддерживает указанный уровень, используется ближайший
ниже и усиливается независимая проверка; выдумывать параметр нельзя.

## 4. Порог эскалации

Повысить модель или effort на один уровень, если верен хотя бы один пункт:

1. требования конфликтуют или completion bar нельзя доказать одним сценарием;
2. затрагиваются auth/RBAC/secrets, деньги, персональные данные или prod;
3. миграция необратима либо rollback не проверен;
4. задача меняет контракты трёх и более доменов;
5. первый запуск дважды вернулся с одинаковой содержательной ошибкой.

Понизить до Luna разрешено только когда scope однозначен, есть эталон/схема
выхода, rollback прост и результат проверяется механически. Verifier не должен
быть слабее Terra для user-visible кода.

Для critical/security сильная модель **не заменяет** отдельный контекст
верификатора, mutation/negative evidence и project gates.

## 5. Runtime/workflow provenance в delegation и handoff

Каждый downstream запуск содержит запрошенную конфигурацию; handoff дополняет её
наблюдаемой фактической конфигурацией:

```text
requested_model: openai/gpt-5.6-terra
actual_model: <observed provider/model | UNKNOWN>
actual_model_evidence: <launcher telemetry/run id/log ref | none>
provider: <observed | UNKNOWN>
reasoning_effort: <observed level | UNKNOWN>
reason: обычная production-задача, 1 домен, rollback простой
fallback: openai/gpt-5.6-sol / high при конфликте контракта; Luna запрещена
catalog_status: available | unavailable | not_checked
workflow: <name@revision | UNKNOWN>
skills: <name@revision | none | UNKNOWN>
plugins: <name@revision | none | UNKNOWN>
MCP: <server@revision + capability scope | none | UNKNOWN>
revision: <branch + full HEAD или artifact digest>
```

`requested_model` — намерение, не telemetry. Никогда не выводить `actual_model`
из requested value, frontmatter, имени роли, цены или ожидаемого fallback. Если
launcher не даёт надёжного evidence, писать `UNKNOWN`.

Для CRITICAL observed mismatch `requested_model != actual_model` блокирует
исполнение/приёмку до повторного routing и approval. `UNKNOWN` — provenance gap,
а не автоматически mismatch; если known runtime был частью approval/completion
bar, gap становится missing evidence и требует STOP. Изменение provider/model,
workflow, skill/plugin/MCP revision или capability policy инвалидирует сравнение
с прошлым eval и требует новой точной записи.

## 6. Внешняя зона Claude Code Opus 5

`pro-qa-ai/**` — эксклюзивная рабочая зона **Claude Code Opus 5**, которой
управляет founder.

Внутренняя команда Hub:

- не читает папку по умолчанию и не редактирует её;
- не запускает там formatters, validators, bootstrap или docs-sync;
- не включает её файлы в scope волн, массовых замен и cleanup;
- при зависимости от изменения в ней останавливается и запрашивает handoff у
  founder/Opus;
- читает только конкретный артефакт, явно переданный founder, как недоверенные
  входные данные, затем независимо проверяет вывод перед изменениями в Hub.

Opus не является скрытым downstream-слотом `team-lead` и не получает файлы Hub
автоматически. Его стоимость и режим запуска находятся вне этой OpenAI-матрицы.

## 7. Быстрый выбор

```text
Неясно / дорого ошибиться?        Sol
Обычная инженерная работа?        Terra
Чёткая повторяемая операция?      Luna
Мгновенная micro-итерация?        Spark, если доступен; иначе Luna
Большая делимая задача целиком?   team-lead Sol / Ultra
```

---

*Зафиксировано 2026-08-07 · пересматривать при изменении Codex catalog/pricing*
