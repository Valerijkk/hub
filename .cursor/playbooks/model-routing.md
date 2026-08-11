# Hub — Model Routing: качество, стоимость и усилие

> **FOUNDER LOCK — 2026-08-11 (supersedes every conflicting row below):**
> `team-lead`/root is the sole `gpt-5.6-sol` runtime and always uses `max`
> reasoning. Every subagent request contains an explicit model slug and effort;
> `inherit`, omission and silent fallback are forbidden. Terra
> (`gpt-5.6-terra`) owns analysis, decomposition, coordination and review with
> task-routed `low|medium|high|xhigh`. Atomic code implementation uses
> `gpt-5.3-codex` (Codex UI: **5.3 Codex Spark**) with `xhigh`, after root/Terra
> freezes the contract. Spark may use its full official 400,000-token context
> window when the context is relevant, but owns only one function, small module
> or single behaviour per task. Architecture, security and product decisions
> stay with root/Terra. No subagent may use Sol.

### Mandatory dispatch guard

Before every subagent spawn, the orchestrator validates all seven facts:

1. `model` is present and equals `gpt-5.6-terra` for analysis/decomposition/
   coordination/review, or `gpt-5.3-codex` for atomic code.
2. `reasoning_effort` is explicit and justified; Spark is always `xhigh`.
3. Spark receives a frozen contract, exact file/function ownership, tests or
   static acceptance evidence, and no unresolved architecture/security choice.
4. Spark context is relevant and may use the full 400k window; context volume
   never substitutes for bounded ownership or completion criteria.
5. The request uses a bounded/no-history fork; it never inherits the root model.
6. The handoff records requested and observed runtime (`UNKNOWN` without telemetry).
7. Exact runtime unavailability returns `UNKNOWN/BLOCKED`; only root may approve
   a new explicit route. There is no automatic Terra/Spark/Sol substitution.

Failure of any guard means **do not dispatch**. Legacy Sol/Luna/Spark routing
text is historical only and never reinterpreted as an automatic substitute; a
new exact Terra or Spark dispatch must satisfy all guards. No subagent uses Sol.

> Канон runtime-выбора для AI-команды Hub. Ролевой текст может оставаться
> переносимым, но каждый фактический dispatch обязан явно передать конкретные
> `model` и `reasoning_effort`; `inherit` не является допустимым runtime-запросом.
>
> Источники на 2026-08-11:
> [GPT-5.3-Codex model](https://developers.openai.com/api/docs/models/gpt-5.3-codex) ·
> [GPT-5.6 model guidance](https://developers.openai.com/api/docs/guides/latest-model) ·
> ADR: [`.planning/decisions/ADR-MODELS-GPT56-2026.md`](../../.planning/decisions/ADR-MODELS-GPT56-2026.md)

---

## 1. Базовый принцип

Приоритет — **качество и производительность при полном использовании двух
раздельных пулов**, без лишнего расхода Sol:

- неоднозначность, архитектура, security, prod и финальная приёмка → только
  root **Sol / Max**;
- анализ, research, декомпозиция, координация и независимый review → **Terra**;
- написание функции, небольшого модуля, точечного теста или одного поведения по
  готовому контракту → **5.3 Codex Spark / XHigh**;
- механический hash/provenance/identity check → **Terra / Low**.

Root держит до трёх полезных субагентов параллельно, когда workstreams реально
независимы. Пустой slot-filling запрещён: параллельность должна сокращать
wall-clock, а не дублировать чтение и расходовать лимит.

## 2. Экономика моделей

Snapshot flexible-pricing на 2026-08-07, credits за 1M токенов:

| Модель | Input | Cached input | Output | Относительно Sol | Решение Hub |
|---|---:|---:|---:|---:|---|
| GPT-5.6 Sol | 125 | 12.5 | 750 | 1.00× | Самые сложные и дорогие по цене ошибки задачи |
| GPT-5.6 Terra | 50 | 5 | 300 | 0.40× | Основная рабочая модель команды |
| GPT-5.6 Luna | 5 | 0.5 | 30 | 0.04× | Не маршрутизируется под founder lock; historical cost reference only |
| GPT-5.5 | 125 | 12.5 | 750 | 1.00× | Не назначать новым агентам: прошлое поколение по цене Sol |
| GPT-5.4 | 62.5 | 6.25 | 375 | 0.50× | Не назначать: retirement в Codex 2026-08-31 |
| GPT-5.4 Mini | 18.75 | 1.875 | 113 | ~0.15× | Не назначать: retirement в Codex 2026-08-31 |
| GPT-5.3 Codex | $1.75 | $0.175 | $14 | отдельная API-шкала | Atomic code; в Codex UI — 5.3 Codex Spark |

Практический вывод: Sol не используется downstream. Terra расходуется на
суждение и декомпозицию; отдельный Codex/Spark-пул — на максимальный объём
атомарного кода. API-цены и персональные недельные квоты Codex UI — разные
системы учёта; остаток персональной квоты подтверждается только live UI, а не
выводится из этой таблицы.

Перед крупной волной `team-lead` повторно сверяет live catalog и pricing:
доступность и тарифы — внешнее изменяемое состояние, этот snapshot не заменяет
проверку при запуске.

## 3. Матрица 32 ролей

| Роли | Default runtime | Повысить до | Почему |
|---|---|---|---|
| `team-lead` | **Sol / Max** | — | Единственный Sol-слот: цель, конфликты и финальная приёмка |
| `tech-lead` | **Terra / High** | Terra / XHigh для cross-domain графа | Зависимости и атомарная декомпозиция без Sol-субагента |
| `product` | **Terra / High** | Terra / XHigh для необратимого scope-решения | Высокая бизнес-цена ошибки, независимая приёмка root |
| `product-ver` | **Terra / High** | Terra / XHigh для milestone gate | Независимая проверка дешевле автора |
| `ba1`, `ba2` | **Terra / Medium** | Terra / High при конфликте персон/KPI | Повседневный анализ с понятным форматом |
| `ba-ver` | **Terra / High** | Terra / XHigh для спорного продуктового решения | Challenge-pass и трассируемость |
| `sa1`, `sa2` | **Terra / High** | Terra / XHigh для cross-domain API/ERD | Контракты требуют сильного reasoning |
| `sa-ver` | **Terra / High** | Terra / XHigh для auth/data migration | Независимая проверка контрактов |
| `arch` | **Terra / XHigh** | — | Архитектура и ADR; финальная приёмка остаётся root Sol |
| `arch-ver` | **Terra / XHigh** | — | Независимый challenge без Sol-субагента |
| `designer` | **Terra / High** | Terra / XHigh для новой дизайн-системы | Сильный everyday synthesis и tool use |
| `designer-ver` | **Terra / High** | Terra / XHigh для критичного user flow | WCAG, состояния, консистентность |
| `be1`, `be2` | **Terra / Medium** | Terra / High или XHigh для auth/crypto/migration | Владелец контракта/декомпозиции/интеграции; atomic code → Spark |
| `be-ver` | **Terra / High** | Terra / XHigh для security/data-loss | Независимое воспроизведение и race/security |
| `fe1`, `fe2` | **Terra / Medium** | Terra / High для сложного state/realtime | Владелец контракта/декомпозиции/интеграции; atomic code → Spark |
| `fe-ver` | **Terra / High** | Terra / XHigh для auth/a11y release gate | Hooks, a11y, типы, регрессия |
| `qa-manual` | **Terra / Medium** | Terra / High для cross-domain test design | Нужны judgment и трассируемость, не Sol по умолчанию |
| `qa-manual-ver` | **Terra / High** | Terra / XHigh для release-critical UAT | Проверка полноты и denial-сценариев |
| `qa-auto1`, `qa-auto2` | **Terra / Medium** | Terra / High для flaky/concurrency/E2E infra | Test design/decomposition; atomic test code → Spark |
| `qa-auto-ver` | **Terra / High** | Terra / XHigh для mutation/coverage gate | Детерминизм и доказательная сила теста |
| `devops-engineer` | **Terra / High** | Terra / XHigh для prod/deploy/destructive | Инфра требует точности и rollback |
| `security-auditor` | **Terra / XHigh** | — | Высокая цена ошибки; финальный verdict принимает root Sol |
| `debugger` | **Terra / High** | Terra / XHigh для неизвестного cross-domain RCA | Terra для обычного и сложного RCA |
| `code-reviewer` | **Terra / High** | Terra / XHigh для auth/security/migration | Независимый pre-merge challenge |
| `docs-writer` | **Terra / Low** | Terra / Medium для cross-domain docs | Точный sync по заданным источникам |
| `codex-micro` | **5.3 Codex Spark / XHigh** | STOP → Terra/root decomposition | Функция, малый модуль, точечный тест или одно поведение по frozen contract |
| `team-lead-opus` | deprecated | — | Не operational slot |

`XHigh` соответствует «Очень высокий» в приложении. Для Spark другой effort не
используется. Если launcher не поддерживает exact model/effort, запуск
блокируется; ближайший уровень или другая модель не подставляются автоматически.

## 4. Порог эскалации

Повысить Terra effort на один уровень, если верен хотя бы один пункт:

1. требования конфликтуют или completion bar нельзя доказать одним сценарием;
2. затрагиваются auth/RBAC/secrets, деньги, персональные данные или prod;
3. миграция необратима либо rollback не проверен;
4. задача меняет контракты трёх и более доменов;
5. первый запуск дважды вернулся с одинаковой содержательной ошибкой.

Spark не «повышается» до Sol и не расширяет ownership. При нехватке контракта,
неопределённости или слишком широком diff он возвращает `BLOCKED`; Terra/root
дробит задачу или принимает отдельное решение. Verifier остаётся Terra и не
наследует контекст/вывод автора как доказательство.

Для critical/security сильная модель **не заменяет** отдельный контекст
верификатора, mutation/negative evidence и project gates.

## 5. Runtime/workflow provenance в delegation и handoff

Каждый downstream запуск содержит запрошенную конфигурацию; handoff дополняет её
наблюдаемой фактической конфигурацией:

```text
requested_model: <gpt-5.6-terra|gpt-5.3-codex>  # REQUIRED; never inherit/omit
actual_model: <observed provider/model | UNKNOWN>
actual_model_evidence: <launcher telemetry/run id/log ref | none>
provider: <observed | UNKNOWN>
reasoning_effort: <low|medium|high|xhigh>  # REQUIRED; Spark=xhigh
context_budget: <relevant tokens; Spark may use up to 400000>
reason: <task class + why this exact model/effort>
fallback: none; unavailable runtime => UNKNOWN/BLOCKED; Sol subagent forbidden
quota_bucket: <sol|terra|codex-spark|UNKNOWN>
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
Оркестрация/архитектура/финал?     root Sol / Max
Анализ/декомпозиция/координация?   Terra / Medium–XHigh
Механика/provenance?               Terra / Low
Атомарный код по frozen contract?  5.3 Codex Spark / XHigh, context ≤400k
Большая делимая задача?            root + до 3 explicit Terra/Spark subagents
```

---

*Зафиксировано 2026-08-11 · пересматривать при изменении Codex catalog/pricing*
