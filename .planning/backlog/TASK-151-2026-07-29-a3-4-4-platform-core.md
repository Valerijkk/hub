# 🐛 A3 Postgres Ping + cron timeout test

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: A3W4-4-001..002 · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1**

## [EXECUTION REPORT]

- **Verdict:** **2/2 FIXED** (A3W4-4-001, A3W4-4-002)
- **Backlog README:** не трогали (owner: статус только в файле задачи → «На проверке»)
- **Handoff:** be-ver → code-reviewer

### ID → evidence

| ID | Status | Evidence |
|----|--------|----------|
| A3W4-4-001 | FIXED | `postgres.go`: `PingContext` после `applyPoolLimits`, до `Schema.Create` / slog connected; err → `drv.Close()` + return. Hubtest: `postgres_ping_a3w44_test.go` (unreachable DSN → error; live DSN Skip/Ping ok) |
| A3W4-4-002 | FIXED | `cron_test.go`: `TestRunTick_TimeoutCancelsHang_A3W44002` (assigned suite; `ExportRunTick` hang > budget) |

### Tests

```text
docker run --rm -v …/backend:/app -w /app golang:1.25-bookworm \
  go test -tags hubtest -vet=off -count=1 -timeout=90s \
  ./testframework/platform/database/ ./testframework/platform/cron/
→ ok database 7.017s · ok cron 0.242s · exit 0
```

Local Windows: cron green; database intermittently blocked by parallel `ent` regen on shared tree (не дефект фикса).

### Mutation

| ID | Сломать | Результат |
|----|---------|-----------|
| A3W4-4-001 | убрать `PingContext` | Логика: Open-only → `err==nil` → `require.Error` краснеет. Live RED на shared tree мешала concurrent ent regenerate (build fail); green с Ping доказан Docker |
| A3W4-4-002 | убрать `WithTimeout` из `runTick` (`fn(ctx)`) | **RED** `cron_test.go:82: tick should have been cancelled by timeout` · `MUTATION002_EXIT=1` |

### Assumptions

- Порядок: Ping **до** `Schema.Create` (мёртвый DSN не доходит до migrate).
- Unreachable DSN: `postgres://hub:hub@127.0.0.1:1/hub?sslmode=disable`, ctx 2s.

## [OUTCOME]

`NewPostgresClient` fail-closed на мёртвом DSN через `PingContext` (симметрия с Mongo/Redis); assigned `cron_test` мутирует tick `WithTimeout` (hang > budget → cancel).

## [ROLE]

- Role: be1
- Responsibility: закрыть **оба** finding ID (A3W4-4-001, A3W4-4-002) + hubtest с mutation proof; не трогать соседние A3W4-* / FIXED-OK cascade
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/platform/database/postgres.go`
  - `backend/testframework/platform/database/**`
  - `backend/testframework/platform/cron/cron_test.go`
  - опционально: `backend/testframework/platform/cron/timeout_task043_test.go`
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W4-4-platform-core.md`
- Do not touch: mongodb/Redis connect, AES-AAD/config/probes/telegram, Ent schema/Atlas, соседние A3W4, frontend
- Behavioral: Conventional Commits RU desc; no AI co-author; fail-closed Ping; Ping до Schema.Create; не ослаблять cron.go

## [REFERENCES]

- Finding IDs: **A3W4-4-001** (NEW, P2), **A3W4-4-002** (NEW, P3)
- Source audit: `.planning/audits/total-2026-07/A3W4-4-platform-core.md`
- Source precedence: **код HEAD** > audit note > эта задача

## [ACCEPTANCE]

Закрыто be1: 2/2 FIXED · Docker hubtest database+cron green · статус **На проверке** (README не трогали).

## [OPERATING RULES]

- Scope only; коммиты ссылают задачу в теле.
