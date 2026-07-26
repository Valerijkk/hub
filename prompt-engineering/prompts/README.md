# prompts/ — библиотека prompt fragments и recipes

> ~1000 фрагментов по 12 категориям тестирования + 112 corporate-delivery. Это **не** готовые «лучшие промпты» сами по себе — однострочники-слоты `[ЧТО]`, которые собираются в production-recipe вместе с active profile и approved pattern cards. Слоты: `{{…}}` (или заполнение из `context/projects/<profile>/`).

## Production use

**Bare fragment недостаточен.** Перед запуском собери recipe:

1. **Context profile** — `../context/projects/<profile>/` (first setup через
   `bootstrap-test-context`; stale/reorganized profile через
   `refresh-test-context`; см. INSTALL).
2. **Pattern cards** — `../patterns/` (approved cards для роли и домена).
3. **Fragment** — выбери `#N` из категории ниже.
4. **Assembly** — оберни фрагмент по профилю категории (блок «Recipe assembly» в шапке файла) + [мастер-шаблон](../agents/templates/_MASTER-PROMPT-TEMPLATE.md).
5. **Eval gate** — smoke/regression из `../evals/` после structural PASS.

Production contracts: [`../recipes/`](../recipes/README.md). Эталоны assembly:
[`EXPANDED-EXAMPLES.md`](EXPANDED-EXAMPLES.md). GSD-порт:
[../gsd-port/README.md](../gsd-port/README.md).

## Категории

| # | Файл | Тема | Фрагментов |
|---|------|------|------------|
| 01 | [01-test-case-design.md](01-test-case-design.md) | Дизайн тест-кейсов, техники тест-анализа | 85 |
| 02 | [02-unit-testing.md](02-unit-testing.md) | Юнит-тесты: написание, структура, ассерты | 85 |
| 03 | [03-integration-testing.md](03-integration-testing.md) | Интеграционные: БД, очереди, транзакции, GraphQL/gRPC/WS, наблюдаемость | 85 |
| 04 | [04-api-testing.md](04-api-testing.md) | API/contract: REST, статусы, схемы, auth, кэш, версии | 85 |
| 05 | [05-e2e-ui-testing.md](05-e2e-ui-testing.md) | E2E и UI: page objects, a11y, мобильные, реалтайм, perf | 85 |
| 06 | [06-test-review-quality.md](06-test-review-quality.md) | Ревью тестов, нейминг, качество, anti-AI, самопроверка | 85 |
| 07 | [07-debugging-flaky.md](07-debugging-flaky.md) | Дебаг падений, flaky, triage, chaos-симптомы | 70 |
| 08 | [08-test-data-mocking.md](08-test-data-mocking.md) | Данные, фикстуры, фабрики, билдеры, моки, стабы | 85 |
| 09 | [09-coverage-analysis.md](09-coverage-analysis.md) | Покрытие (branch/mutation), gap-анализ по риску | 80 |
| 10 | [10-ci-automation.md](10-ci-automation.md) | CI-пайплайны, гейты, окружения, merge queue, метрики | 85 |
| 11 | [11-performance-security-testing.md](11-performance-security-testing.md) | Нагрузочное, chaos, resilience, OWASP security | 85 |
| 12 | [12-documentation-reporting.md](12-documentation-reporting.md) | Тест-документация, планы, отчёты, метрики, онбординг | 85 |
| 13 | [13-corporate-delivery.md](13-corporate-delivery.md) | Corporate delivery: plan/waves/review/strangler/PR/handoff/verify/kit/ops | 112 |

**Итого: 1000+ фрагментов** (тесты) **+ 112 corporate**.

## Как пользоваться

1. **Не напрямую:** копирование однострочника без context profile и pattern cards даёт слабый результат.
2. **Recipe assembly (рекомендуется):** фрагмент = слот `[ЧТО]` мастер-шаблона. Оберни по **профилю категории** (блок «Recipe assembly» в шапке: роль, контекст, формат, rationale, assumptions, evidence, верификация).
3. **Эталоны:** [`EXPANDED-EXAMPLES.md`](EXPANDED-EXAMPLES.md) — 12 полностью собранных recipes (разворот `#1` каждой категории) + сводная таблица профилей.
4. **Композиция:** фрагменты связываются через «Затем: <следующий фрагмент>» — типовые связки в конце категорий.
5. **С агентами:** для конвейерных задач (ревью, генерация TC, дебаг) предпочитай агентов из `agents/` — они несут процесс и формат; библиотека — для точечных задач и одноразовых операций.

## Почему не все 1000 развёрнуты

Полная форма мастер-шаблона × 1000 = ~40k строк одинакового boilerplate. Повторяющаяся часть вынесена **один раз на категорию** (профиль в шапке), уникальная часть — задача — осталась в однострочниках, результат assembly показан в 12 эталонах. Любой фрагмент собирается за ~30 секунд; библиотека остаётся навигируемой.

## Конвенции

- `{{слот}}` — обязательная подстановка; `[опция|опция]` — выбор.
- Фрагмент формулирует **и задачу, и формат выхода** — не убирай вторую часть при assembly.
- Нумерация сквозная внутри файла — ссылайся как `prompts/04 #23`.
- Параметры модели (temperature и т.п.) — зона model router/launcher, не prompt fragment.
