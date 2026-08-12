# 🐛 A3 Postgres Ping + cron timeout test

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта** · MUTATION-2026-07-30-W13-qa1-task151  
> Источник: A3W4-4-001..002 · Stage A3 FINAL · TASK-008  
> Исполнитель: **be1** (fix) · Verifier: **qa-auto1** (W13 slot 4 refill)

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
# Isolated worktree (HEAD + SCOPE) — shared tree был бит concurrent ent regen
go test -tags hubtest -race -count=1 -timeout=120s \
  ./testframework/platform/database/ ./testframework/platform/cron/
→ ok database 6.866s · ok cron 3.308s · exit 0
```

### Mutation

| ID | Сломать | Результат |
|----|---------|-----------|
| A3W4-4-001 | убрать `PingContext` (Open-only + slog connected) | **RED** `An error is expected but got nil` + slog `connected to PostgreSQL` · MUT1_EXIT=1 |
| A3W4-4-002 | убрать `WithTimeout` из `runTick` (`fn(ctx)`) | **RED** `tick should have been cancelled by timeout` · MUT2_EXIT=1 |

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

Закрыто be1: 2/2 FIXED · Docker hubtest database+cron green · mutation qa-auto1 W13 **2/2 PROVEN** · статус **Закрыта**.

## [OPERATING RULES]

- Scope only; коммиты ссылают задачу в теле.

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 4 / qa-auto1 · criteria only (no mutate) · 2026-07-30

| ID | prod file:line | слой | RED test (path + name) | kill recipe |
|----|----------------|------|------------------------|-------------|
| A3W4-4-001 | `backend/internal/platform/database/postgres.go:51-55` (`PingContext` + `drv.Close` до `Schema.Create` / slog) | database connect (real `sql.Open`+Ping) | `backend/testframework/platform/database/postgres_ping_a3w44_test.go` · `TestNewPostgresClient_UnreachableDSN_FailClosed_A3W44001` | Убрать блок `PingContext` (оставить Open→`ent.NewClient`→slog `connected`) → expect RED: `An error is expected but got nil` (+ slog connected) |
| A3W4-4-002 | `backend/internal/platform/cron/cron.go:96` (`tickCtx, cancel := context.WithTimeout(ctx, tickTimeout)`) | cron `runTick` | `backend/testframework/platform/cron/cron_test.go` · `TestRunTick_TimeoutCancelsHang_A3W44002` | Заменить `WithTimeout` на `fn(ctx)` без бюджета → RED: `tick should have been cancelled by timeout` / `fn must observe context deadline` |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Real unreachable DSN (`127.0.0.1:1`) через `NewPostgresClient` | **Ent OK** (не mock; guard на connect path) |
| 002 | `ExportRunTick` hang > budget | **OK** (unit cron; не Ent) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)`

### Gaps

- LiveDSN sibling `TestNewPostgresClient_LiveDSN_PingOK_A3W44001` — Skip без compose; мутация 001 канон = UnreachableDSN.
- Не мутировать оба ID одним diff — ambiguous RED mapping.

## Progress (qa-auto1 · 2026-07-30 · W13 slot 4 refill)

Независимый Break→RED→restore→GREEN по `[MUTATION CRITERIA]`.  
Отчёт: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task151.md`.

| ID | Verdict | RED |
|----|---------|-----|
| A3W4-4-001 | **PROVEN** | `An error is expected but got nil` (+ slog connected) |
| A3W4-4-002 | **PROVEN** | `tick should have been cancelled by timeout` |

Leftover MUTATED: нет. Статус → **Закрыта**.
