# Test design и traceability

## От чего проектируем

Источник scenario — current requirement, task, API/ADR, threat model и
observable runtime contract. Legacy row или существующий test может подсказать
риск, но не определяет новый case и не создаёт parity obligation.

Минимальная цепочка:

```text
task -> requirement -> risk/failure mode -> TC-* -> suite -> executable case -> evidence
```

Каждая ссылка разрешается в конкретный артефакт. «Покрыто», имя файла, status
emoji или coverage percentage без oracle/result не являются traceability.

## Карточка TC

| Поле | Требование |
|---|---|
| ID | стабильный `TC-<DOMAIN>-<NN>` |
| Task | задача, которая ввела/изменила behavior |
| Requirement | FR/NFR/GZ/ADR/API/spec section |
| Risk | конкретный failure и impact |
| Preconditions/data | воспроизводимые actor/tenant/state/input |
| Action | точное действие на наблюдаемой границе |
| Oracle | output/state/effect и запрещённые side effects |
| Layer | unit/contract/api/integration/e2e/load |
| Cleanup | owner и observable result |
| Automation | exact case reference либо честный manual/backlog status |

## Scenario families

Для каждого risk выбрать только применимые families:

- happy; invalid;
- authn; authz/tenant;
- min/max/boundary; empty/one/many; null/Unicode/time/timezone;
- duplicate/reorder; stale/version;
- timeout; unavailable dependency; partial completion;
- retry/recovery; cancellation; concurrency;
- replay/idempotency; resource exhaustion;
- security/privacy/redaction; accessibility; performance.

Нельзя требовать «по одному test каждого типа» или фиксированную quota.
Вместо этого reviewer спрашивает, какой failure model остался непроверенным и
почему `N/A` безопасно.

## Техники

- Equivalence classes — различающиеся правилами классы input/actor/state.
- Boundary values — `below / at / above`, если limit нормативно подтверждён.
- Decision table — независимые conditions меняют outcome (RBAC, workflow).
- State transition — допустимые/запрещённые переходы и повтор операции.
- Pairwise — когда полная комбинация велика, а interaction order ограничен.
- Model/property/metamorphic — когда invariant сильнее списка examples.

Техника не применяется ради отметки. Например, Magic Link требует expiry,
rate limit, replay и tenant/security risks; tracker optimistic locking — stale
version и concurrent update; messenger — membership, ordering/replay; calls —
webhook authenticity, timeout и recovery; search — user/tenant scope.

## Oracle и false confidence

Хороший oracle наблюдает contract: status/body/state/event/audit/no-side-effect,
а не private calls. Несколько assertions допустимы для одного outcome.

Для high-risk behavior временный sentinel/mutation должен сделать case
красным: убрать authz guard, принять stale version, не опубликовать event,
проглотить dependency failure. Sentinel возвращается; evidence фиксирует
expected red и restored green. Нельзя ослаблять production security ради
mutation.

## Manual cases

Manual status честен и содержит environment, data, steps, oracle и cleanup.
«Работает корректно» и «проверить UI» не воспроизводимы. Manual case не обязан
автоматизироваться; automation выбирается по repeatability, risk, cost и
observable boundary.

## Review checklist

- current requirement и task найдены;
- risk сформулирован как failure, не как feature name;
- suite выбран по feedback/boundary;
- применимые families рассмотрены без quota;
- test data и cleanup изолированы;
- oracle наблюдаем и чувствителен;
- automation link и evidence точны;
- row parity и historical coverage не использованы как acceptance.

См. [основной playbook](../testing.md),
[`15-test-automation.md`](../../knowledge/code-practices/15-test-automation.md) и
[`negative-sentinels.md`](../../knowledge/code-practices/controls/negative-sentinels.md).
