# TASK-100 — 📐 Backup MaintenanceGate distributed (U04)

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
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
  - deployments docs
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - оставить process-local в prod K3s
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: требуется ADR/owner approve перед кодом/migrate

## [REFERENCES]

- Finding IDs: W7-A4-003 U04 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR U04
  - distributed gate + IT multi-instance или simulated
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
