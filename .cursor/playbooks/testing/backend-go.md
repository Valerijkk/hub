# Backend — genuine isolated Go units

> Target: `backend/tests/{domain}/**`, CI `backend-unit`. Legacy
> `backend/testframework/**` — delete-not-port inventory, не пример.

## Что считается unit

Unit проверяет поведение Go unit без PostgreSQL, MongoDB, Redis, NATS, storage,
HTTP server, containers или внешней сети. Если важен реальный boundary, это
greenfield Python `api`/`integration`/`contract`, а не Go unit с другим именем.

Допустимы pure functions, validators, domain policy, deterministic service
logic за узкими fakes/stubs и concurrency logic с controllable scheduler/time.
Handler/API semantics, Ent schema/index/cascade и broker delivery в этот job не
попадают.

## Структура

```text
backend/tests/<domain>/
  <behavior>_test.go
```

- имя отражает outcome, не метод implementation;
- один test/subtest — один behavioral concept; несколько связанных assertions
  допустимы;
- table-driven cases полезны для equivalence/boundary/state matrix, но не
  обязательны для единичного рассказа;
- `t.Helper` и factories скрывают механику, не scenario values/oracle;
- `t.Cleanup` владеет каждым созданным ресурсом;
- `t.Parallel` разрешён только при доказанной изоляции state и identifiers.

## Doubles

Stub/fake/mock — не цель. Сначала предпочитать state/output oracle. Проверять
interaction можно, когда сам факт collaboration наблюдаем: publish ровно один
domain event, не отправить notification при deny, передать cancellation или
redacted payload. Private call order и полный mock transcript — review smell
(`NEG-08`), а не contract.

Fake обязан сохранять semantics, важные для теста: not-found, conflict,
idempotency и cancellation. Универсальный `nil, nil` создаёт false confidence.

## Failure model

Выбирать только применимое: happy, invalid, authn/authz/tenant, boundary,
empty/one/many, duplicate/reorder, cancellation, concurrency, replay,
version. Timeout/retry/dependency failure чаще требует Python boundary test,
если unit-double уже не доказывает production integration.

## Docker evidence

```powershell
docker run --rm --name hub-backend-unit-<task> `
  --cpus 2.0 --memory 4g --memory-swap 4g `
  --mount type=bind,src="${PWD}/backend",dst=/work/backend,readonly `
  -w /work/backend golang:1.25-bookworm `
  sh -lc "go build -tags release ./... && go vet ./... && go test -race -count=1 -timeout=600s ./tests/..."
```

Host-run запрещён и не является evidence. `--rm` очищает task-owned container.
Записать image, source SHA, exact command, collect/pass/fail/skip, duration и
cleanup. Исторический 82% coverage не блокирует cutover; coverage не заменяет
TC/risk/oracle и failing sentinel.

## Review

- путь только `backend/tests/{domain}`;
- нет service/container/network dependency;
- DAMP scenario, DRY mechanics;
- double соответствует observable contract;
- race/cancellation ownership проверены где применимо;
- тест красный при удалении защищаемого behavior;
- TC/task/requirement traceability точна.

См. [основной playbook](../testing.md) и
[`15-test-automation.md`](../../knowledge/code-practices/15-test-automation.md).
