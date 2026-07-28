# ADR-BACKUP-MAINTENANCE-GATE-2026 — Distributed MaintenanceGate (U04)

- Статус: **Accepted** (owner 2026-07-28: отдельный approve ADR не требуется)
- Дата: 2026-07-28
- Автор: be1 · Верифицирует: be-ver / arch-ver
- Связанные: A7-003, W7-A4-003, P1-BE-U04, TASK-100,
  `.planning/audits/total-2026-07/A7-4-storage-dash-backup.md`

## Контекст

`MaintenanceGate` (A7-003) блокировал non-health API 503 на время
`pg_restore --clean` / `mongorestore --drop`, но хранил флаг в
process-local `atomic.Bool`. При >1 реплике K3s restore на поде A
не останавливал запись на поде B → риск corruption.

## Рассмотренные варианты

1. **Redis shared flag / lease** — выбрано
2. Single-replica affinity + drain перед restore — отклонено как
   ops-only workaround; не закрывает multi-replica без ручного drain
3. DB-таблица флага — отклонено: во время `pg_restore --clean` сама
   таблица может быть недоступна/пересоздана
4. Оставить process-local — запрещено outcome TASK-100 / U04

## Решение

1. Канон: ключ Redis `hub:backup:maintenance` = `"1"` пока restore идёт.
2. `NewRedisMaintenanceGate(rdb)` в prod wiring (`wireFiber` после
   `wireInfra`); `NewMaintenanceGate()` без Redis — только unit-тесты
   без кластера.
3. `Set(ctx, true)` → `SET` с TTL **2h** (safety lease: под умер без
   `defer` clear → ключ истечёт). `Set(ctx, false)` → `DEL`.
4. `Enabled(ctx)` читает Redis (source of truth). При ошибке Redis —
   fallback на локальный write-through кэш (реплика-restorer всё ещё
   блокирует себя); residual: другие реплики при полном outage Redis
   не увидят флаг — Redis обязателен в prod (как sessions/rate-limit).
5. Middleware без изменений контракта: 503 + health/live/ready/metrics
   bypass.
6. **Не** single-replica affinity в K3s — multi-replica остаётся нормой.

## Finding IDs (TASK-100)

| ID | Статус | Evidence |
|----|--------|----------|
| W7-A4-003 (U04 / P1-BE-U04) | **FIXED** | Redis gate + IT multi-instance (`maintenance_distributed_task100_test.go`) |

## Последствия

- (+) Cluster-wide 503 во время restore
- (+) TTL защищает от «залипшего» maintenance после crash
- (−) Hot-path `GET` Redis на каждый API-запрос в режиме проверки
  (приемлемо: тот же класс, что rate-limit Redis)
- (−) При Redis down без локального Set — fail-open на чужих репликах

## Approval record

- Owner: **апрув ADR не требуется** (2026-07-28, делегация TASK-100)
- Migrate: N/A (без Ent)
