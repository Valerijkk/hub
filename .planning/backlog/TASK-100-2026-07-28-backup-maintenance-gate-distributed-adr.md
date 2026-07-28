# TASK-100 — 📐 Backup MaintenanceGate distributed (U04)

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: **Исправлена**  
> Источник: W7-A4-003 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

MaintenanceGate работает на multi-replica (Redis/lease), не process-local atomic.Bool.

## [ROLE]

- Role: devops→be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/backup/**
  - backend/cmd/server/wiring_fiber.go (Redis DI)
  - deployments docs / docs/wiki/DEPLOY.md
  - `.planning/decisions/ADR-BACKUP-MAINTENANCE-GATE-2026.md`
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - single-replica affinity в prod K3s (не замена Redis-gate)
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **апрув ADR не требуется** (owner 2026-07-28)

## [REFERENCES]

- Finding IDs: W7-A4-003 U04 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-BACKUP-MAINTENANCE-GATE-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR U04 ✅
  - distributed gate + IT multi-instance ✅
  - W7-A4-003 FIXED с evidence ✅
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код после approve в Approval record (здесь: approve не требуется)

## [EVIDENCE]

| ID | Status | Evidence |
|----|--------|----------|
| W7-A4-003 (U04) | FIXED | Redis `hub:backup:maintenance`; `NewRedisMaintenanceGate`; IT `maintenance_distributed_task100_test.go` |

**Mutation:** вернуть `Enabled` только на `local atomic.Bool` → `TestMaintenanceGate_DistributedRedis_Task100` падает на `gateB.Enabled`.

**Команды:**
```
cd backend && go test -tags hubtest -race -count=1 -timeout=120s ./testframework/backup/
```
