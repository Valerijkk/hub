# Frontend — genuine Vitest/RTL units

> Target: `frontend/tests/{domain}/**` и `frontend/tests/support/**`, CI
> `frontend-unit`. Legacy static-contract Vitest и
> `testframework/stages/unit/**` удаляются и не портируются.

## Scope

Vitest/RTL unit проверяет observable component, hook, store, reducer или
client-mapping behavior в service-free окружении. Source-text/AST/grep guard,
browser journey, live backend и cross-service contract сюда не входят.

## Queries и interactions

Приоритет: accessible role/name → label → visible text → semantic scoped query.
`data-testid` допустим, когда доступной семантики действительно нет. Raw CSS,
DOM shape и implementation state требуют rationale.

Использовать `userEvent` и ожидать наблюдаемый результат: accessible state,
rendered content, navigation intent или public boundary call. Не утверждать
private hook order.

## Doubles и support

- локальный stub/spy — для одной простой boundary;
- MSW/fake — когда важна HTTP semantics нескольких requests;
- unhandled request должен падать;
- interaction assert допустим только для observable collaboration;
- `frontend/tests/support/**` хранит повторяемую механику, а relevant values и
  specific oracle остаются в test;
- store, clock, timers, mocks и storage сбрасываются между tests.

Не требовать mock, MSW, factory или общий renderer ради единообразия. Выбор
обосновывается failure mode. Snapshot не заменяет behavioral oracle.

## Scenario families

Для UI выбирать применимое: happy, invalid, loading/error/empty/one/many,
authn/authz, boundary/Unicode/timezone, duplicate/reorder, timeout/unavailable,
partial/retry/recovery/cancel, stale/version и accessibility. Это не quota.

## Docker evidence

Frontend unit job запускается только versioned outer launcher из `/qa-gate`:
outer wall-clock timeout, exit-code preservation, task-owned container/volumes
и unconditional cleanup в `finally`. Host Node/Vitest/tsc/Biome/Vite и прямой
`docker run` запрещены. Пока launcher не реализован и независимо не проверен,
verdict — `NOT_READY`; cleanup result входит в handoff.

## Review

- только unit behavior, не static contract;
- semantic query first и observable oracle;
- DAMP scenario, DRY support;
- нет arbitrary sleep/shared state/real network;
- test чувствителен к удалению защищаемого behavior;
- TC/task/requirement mapping точен.

См. [основной playbook](../testing.md) и [Python E2E](e2e.md).
