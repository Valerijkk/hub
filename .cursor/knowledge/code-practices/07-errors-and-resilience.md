# 07 — Ошибки и resilience

## Typed failure contract

- Принцип: caller должен различать failure classes только там, где отличается recovery.
- Evidence: `NORM`, `REC`; high; [S47, S49–S53](sources/registry.md).
- Режим: `HYBRID` — structured API/error contract проверяется тестом; смысл retryability проверяется review.
- BAD: `catch (Exception) { retry(); }` или `fmt.Errorf("failed")` без сохранения причины.
- GOOD: `PaymentDeclined` завершается сразу; `PaymentUnavailable` может быть retried в пределах deadline.
- Исключение: human-only diagnostic leaf, если ни один caller не ветвится по типу.

## Remote operation

- Принцип: remote call не является local function call; ожидание и amplification должны быть ограничены.
- Evidence: `CONTROL`, `REC`; high; [S49, S51, S53](sources/registry.md).
- Режим: `HARD/HYBRID` — timeout/deadline и finite retry budget обязательны; классификация transient failures проверяется review/tests.
- BAD: infinite wait, `while true`, одинаковый retry для validation и network timeout.
- GOOD: deadline → classified transient failure → exponential backoff with jitter → finite cap → observable terminal state.

## Idempotency and partial failure

- Принцип: повтор логической команды не должен непредсказуемо дублировать non-idempotent effect.
- Evidence: `CONTROL`, `REC`; high; [S50, S52](sources/registry.md).
- Режим: `HYBRID` — idempotency key/result retention, dedup/outbox/reconciliation и replay tests по documented boundary.
- BAD: повторить `POST /charge` после timeout без знания, прошёл ли первый charge.
- GOOD: stable command key, persisted result/state, bounded retry и explicit reconciliation.
- Исключение: naturally idempotent read либо операция, где duplicate доказанно harmless.

## Cancellation and ownership

- Принцип: request-scoped work прекращается вместе с owner, если оно не объявлено durable detached job.
- Evidence: `NORM`, `REC`; high; [S39, S40, S51](sources/registry.md).
- Режим: `HYBRID` — propagation tests + language review.
- BAD: goroutine/background promise без owner, cancellation и terminal state.
- GOOD: context/signal проходит через call graph; detached job сохраняет durable identity и status.

## Exactly once

Любой claim обязан ответить: exactly once **что**, между **какими boundaries**, при **какой модели restart/failure** и для **каких external side effects**. Иначе срабатывает `NEG-17`, а claim не принимается.

## Review checklist

- timeout/deadline;
- cancellation owner;
- transient/permanent classification;
- retry cap and jitter;
- idempotency/dedup/reconciliation;
- partial completion state;
- observable retry/terminal status;
- replay and version compatibility tests.

Примеры: [Python](examples/python.md), [Go](examples/go.md), [decision matrix](decisions/decision-matrix.md).
