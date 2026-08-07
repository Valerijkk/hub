# Negative sentinels

Sentinel — минимальная fixture, которая обязана сделать соответствующий
detector красным. Без неё «валидатор» может всегда возвращать PASS.

| ID | Сильный сигнал | Default |
|---|---|---|
| NEG-01 | broad exception catch с безусловным retry | hard candidate |
| NEG-02 | бесконечный цикл с remote call без deadline/budget | hard candidate |
| NEG-03 | `sleep` как UI/integration synchronization | hard candidate |
| NEG-04 | mutable global singleton из domain/service code | hybrid |
| NEG-05 | React Effect только копирует derived state | review warning |
| NEG-06 | test body ветвится и меняет oracle по role/state | hybrid |
| NEG-07 | shared account/record между parallel tests | hard candidate |
| NEG-08 | mock проверяет длинную последовательность private calls | review warning |
| NEG-09 | network client без видимой timeout policy | hard candidate |
| NEG-10 | goroutine/task без owner, cancellation и observable completion | hybrid |
| NEG-11 | unbounded queue/channel/fan-out | hard candidate |
| NEG-12 | SQL interpolation/concatenation с external value | hard candidate |
| NEG-13 | tenant ID принимается как достаточное доказательство authz | hard candidate |
| NEG-14 | third-party GitHub Action pinned к mutable tag/branch | hard candidate |
| NEG-15 | release/deploy использует mutable image tag `latest` | hard candidate |
| NEG-16 | retry create/POST без idempotency/reconciliation analysis | hybrid |
| NEG-17 | «exactly once» без границы гарантии и failure model | hybrid |
| NEG-18 | prompt-инструкция заменяет capability/tool restriction | hard candidate |
| NEG-19 | issue/web/email/retrieved text автоматически становится authority | hard candidate |
| NEG-20 | destructive/external agent tool доступен без approval/policy | hard candidate |

## Правила реализации

- Fixture не содержит реальный secret, credential, tenant data или опасный
  executable payload.
- Sentinel проверяет не только finding code, но и отсутствие секретного значения
  в diagnostic output.
- Hard detector fail-closed на unreadable/ambiguous input только в заявленном
  scope; вне scope возвращает `UNKNOWN`, а не выдуманный PASS.
- Изменение detector и его sentinel входит в один атомарный change.
- Стилистические smells не становятся blocker без локальной telemetry и
  подтверждённого сочетания более сильных условий.

## Пример

```python
def retry_forever(client):
    while True:  # NEG-02: no deadline, no finite attempt budget
        client.post("/orders")
```

Правильная проверка доказывает, что fixture обнаружена под точным ID, затем
отдельная good fixture с deadline и bounded retry остаётся зелёной.

