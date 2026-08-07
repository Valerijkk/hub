# 00 — Принципы

## Определение

Код считается поддерживаемым, когда инженер может локально понять его контракт, предсказать радиус изменения и быстро получить доказательство, что изменение не нарушило значимые свойства системы.

## Пять целей

| Цель | Принцип | Evidence | Control или review | BAD → GOOD | Sources |
|---|---|---|---|---|---|
| Cognitive load | Держать поведение, assumptions и effects локально обозримыми | `EMP`, `REC`; medium | warning по размеру/сложности + review cohesion; не line gate | mixed-level procedure → ясная orchestration с meaningful units | [S02, S12–S14](sources/registry.md) |
| Coupling | Зависеть от минимального необходимого capability | `NORM`, `REC`; high | contract tests и boundary review | service locator/широкий interface → explicit consumer contract | [S03, S20, S21](sources/registry.md) |
| Change radius | Скрывать реально изменчивое решение за стабильной границей | `EMP/FOUNDATIONAL`; high | architecture review; hard только для конкретного invariant | слои без скрытого решения → adapter вокруг vendor/schema | [S02](sources/registry.md) |
| Feedback latency | Проверять риск максимально быстрым hermetic способом, не теряя boundary coverage | `REC`; high | CI jobs по scope/resources; risk-based mix | один большой E2E gate → быстрые units + boundary/API/E2E evidence | [S25, S28](sources/registry.md) |
| Failure/security blast radius | Явно ограничивать trust, retry, concurrency и side effects | `CONTROL`, `NORM`; high | hard/hybrid controls | infinite retry/global privilege → deadline, idempotency, least privilege | [S49–S53, S56–S61](sources/registry.md) |

## Базовые design questions

Для изменяемой unit должно быть возможно ответить:

- Что она обещает наблюдаемому клиенту?
- Какие входы и состояния допустимы?
- Какие side effects и dependencies она затрагивает?
- Какова область lifetime/ownership?
- Как ведёт себя при timeout, cancellation, duplicate и partial failure?
- Каким самым быстрым evidence подтверждается изменение?

Если ответ требует чтения половины репозитория, проблема обычно в границе или скрытом состоянии, а не в числе строк.

## Иерархия решений

1. Сначала корректность, security, data integrity и failure model.
2. Затем ясные contracts, ownership и change boundaries.
3. Затем feedback и observability.
4. После этого readability и локальные эвристики.
5. Style не может оправдать ослабление уровней 1–3.

## Anti-cargo-cult

- SOLID — вопросы о change/dependency/behavior, не квота интерфейсов.
- DRY — один authority для знания, которое обязано меняться вместе; похожие строки могут описывать разные concepts.
- Малые функции полезны, пока extraction создаёт имя и cohesion, а не навигационный шум.
- Composition — default при независимой вариативности; стабильный behavioral subtype остаётся допустим.
- Test Pyramid, Trophy и Honeycomb — контекстные модели, не обязательные пропорции.

См. [cargo-cults](decisions/cargo-cults.md) и [contradictions](decisions/contradictions.md).
