# 15 — Test automation

## Objective

Тестовая архитектура оптимизирует feedback, confidence и диагностичность по failure model. Геометрическая форма suite и проценты по слоям не являются целью.

## Test card

| Principle | Evidence | Mode | BAD | GOOD | Sources |
|---|---|---|---|---|---|
| Один behavioral concept | `REC`; high | review | one-assert ritual или смешанные outcomes | несколько assertions, подтверждающих один observable outcome | [S25, S26](sources/registry.md) |
| DAMP narrative, DRY mechanics | `REC`; high | review | fixture DSL скрывает scenario | relevant values локальны, mechanical setup вынесен | [S26](sources/registry.md) |
| Hermetic feedback | `REC`; high | hybrid | shared mutable state/network по умолчанию | isolated resources + controlled boundary tests | [S25, S28](sources/registry.md) |
| Doubles по контракту | `REC`, `DISPUTED`; high/medium | review | mock private call order | state/output либо interaction, если interaction observable | [S23](sources/registry.md) |
| UI abstraction по повторению знания | `REC`; high | review | giant AppPage | semantic locator → helper → component/page object при реальном reuse | [S31, S32](sources/registry.md) |
| Risk-based suite mix | `REC`, `OPINION`; high/medium | review | 70/20/10, Trophy или Honeycomb как gate | выбрать scopes по failure model и latency | [S25, S28](sources/registry.md) |

## Size ≠ scope

- `size` описывает ресурсы и isolation;
- `scope` описывает объём системы;
- small test может покрывать несколько классов;
- broad scope не оправдывает uncontrolled dependencies.

Не принуждайте все suites к одной taxonomy только ради имени каталога.

## Doubles

| Double | Допустимое применение | Risk |
|---|---|---|
| Stub | Заданный ответ boundary | Overfit к shape запросов |
| Fake | Lightweight working semantics | Drift от production |
| Spy | Наблюдать meaningful collaboration | Interaction coupling |
| Mock | Проверить interaction как contract | Break при harmless refactor |

Мок не обязателен и не запрещён. Вопрос: является ли interaction наблюдаемой частью контракта?

## UI layers

1. Plain accessible locator для небольшого flow.
2. Tiny action helper при повторе meaningful action.
3. Component Object для повторяемого widget.
4. Page Object для стабильных page services/locators в крупной suite.

Specific oracle обычно остаётся в test. POM сам по себе не gate.

## Flakiness controls

- inject/freeze time, record seed;
- fresh isolated state;
- deterministic condition вместо sleep;
- bounded eventual assertion;
- explicit cleanup;
- visible ownership и временная quarantine с expiry.

Проценты причин flakiness из чужого проекта не переносятся в Hub без локальной telemetry.

## Scenario families

Выбирать по риску: happy, invalid, authn, authz/tenant, boundaries, empty/one/many, duplicate/reorder, timeout, unavailable dependency, partial completion, retry/recovery, cancellation, concurrency, replay/idempotency, resource exhaustion, version compatibility. Это не требование создать фиксированное число tests на endpoint.

См. [test examples](examples/tests.md), [warning signals](controls/warning-signals.md) и [negative sentinels](controls/negative-sentinels.md).
