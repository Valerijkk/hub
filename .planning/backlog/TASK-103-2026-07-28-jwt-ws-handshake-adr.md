# TASK-103 — 📐 JWT: no sessionStorage / no WS ?token=

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-29-W6  
> Источник: A2W3/W11 JWT · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Access JWT не в sessionStorage; WS не принимает ?token= — ticket/subprotocol по ADR.

## [ROLE]

- Role: arch→be1+fe1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - frontend/src/** auth/ws
  - backend/internal/ws/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ослабить WS RBAC
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **не требуется** (A2-DEDUP 2026-07-28 — ADR+код в одном цикле)

## [REFERENCES]

- Finding IDs: A2W3-4-001, A2W11-2-001, A2W11-3-002 (3)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-WS-JWT-HANDSHAKE-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR ✅
  - FE+BE handshake ✅
  - denial на ?token= ✅
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record (для TASK-103 — owner lock: approve не требуется)

## [EVIDENCE] 2026-07-28

| ID | Status | Evidence |
|----|--------|----------|
| A2W3-4-001 | FIXED | `TestWSHandler_QueryToken_Denied` — `?token=` → 401 |
| A2W11-2-001 | FIXED | auth-store без persist; тест `does not persist JWT into sessionStorage` |
| A2W11-3-002 | FIXED | ws-store ticket subprotocol; тест `connects via ticket subprotocol without ?token=` |

**Mutation:** вернуть `c.Query("token")` accept → `TestWSHandler_QueryToken_Denied` краснеет; вернуть sessionStorage persist JWT → FE unit краснеет.

**ADR:** `.planning/decisions/ADR-WS-JWT-HANDSHAKE-2026.md` (Accepted).
