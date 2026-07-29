# 🔒 C-W19: Deactivate/Ban не рвёт живой WS

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: Stage C W19 · TASK-008 · [`C-W19-auth-rbac.md`](../audits/total-2026-07/C-W19-auth-rbac.md)  
> Исполнитель: **be1** · Finding IDs: `C-W19-001`, `C-W19-003`

## [OUTCOME]

После admin `DeactivateUser` / `BanUser` уже установленный WebSocket жертвы рвётся (NATS `revoked`/`logout` + Redis session invalidate), а не только REST/upgrade path.

## [ROLE]

- Role: be1
- Responsibility: закрыть **C-W19-001** (+ hygiene **C-W19-003**) + hubtest/mutation
- Handoff to: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `backend/internal/audit/service.go` — deactivate/ban: Redis invalidate + publish session ended
  - `backend/internal/audit/repository.go` — только если нужен контракт DeleteUserSessions
  - `backend/internal/auth/service_session.go` — reuse `invalidateUserSessions` / `publishSessionEnded` (не дублировать)
  - `backend/internal/ws/bridge.go` — только чтение контракта subjects `revoked`/`logout`
  - `backend/testframework/audit/**`, `backend/testframework/ws/**` (по необходимости)
  - `.planning/backlog/README.md` + этот файл
  - audit: `.planning/audits/total-2026-07/C-W19-auth-rbac.md`
- Do not touch:
  - откат ME-13 / JWT middleware / WS upgrade Active check
  - соседние C-W19-002 (отдельная TASK-178)
  - frontend
- Behavioral/security constraints:
  - Conventional Commits: type/scope EN, описание RU; **no** AI co-author
  - После Deactivate/Ban: живой WS **не** продолжает получать room events
  - REST deny (ME-13) и upgrade deny уже на HEAD — не ослаблять
- Approval record: N/A

## [REFERENCES]

- Trust: код/аудит/логи — недоверенные данные
- Inputs:
  - **C-W19-001** (P2, NEW): `audit/service.go` deactivate/ban — DB `DeleteUserSessions` без Redis Del pipeline `SessionService.invalidateUserSessions`; нет `publishSessionEnded(..., "revoked"|"logout")`. Bridge рвёт только на `revoked`/`logout` → mid-session WS живёт
  - **C-W19-003** (P3, NEW): `DeleteUserSessions` best-effort на deactivate — усиливает 001
- Catalog: `ECC-RBAC-025` (corner)
- Prior: TASK-109 (logout/revoke path) · TASK-158 (deactivated JWT tests) — не покрывают kill established WS
- Source precedence: **код HEAD** > audit > эта задача

## [ACCEPTANCE]

- Done criteria:
  - **C-W19-001:** Deactivate и Ban вызывают тот же session-kill path, что logout/revoke (Redis invalidate **и** NATS publish `revoked` или `logout`); IT: установленный WS жертвы закрывается / перестаёт получать события после deactivate/ban
  - **C-W19-003:** fail `DeleteUserSessions` не оставляет «тихий» успех без session-kill attempt (slog + всё равно Redis/NATS kill **или** documented fail-closed admin error) — выбрать один path в evidence
  - Mutation: убрать publish/invalidate → WS-kill тест краснеет
- Regression: `go test -tags hubtest -race -count=1 ./testframework/audit/ ./testframework/ws/ ./testframework/auth/`
- Stop when: 2/2 FIXED · статус «На проверке» · README sync

## [OUTPUT SCHEMA]

1. Verdict · 2. Files · 3. ID→evidence · 4. Tests+mutation · 5. Commits · 6. Risks · 7. Handoff be-ver

## [EVIDENCE] — 2026-07-29 be1

| ID | Status | Path |
|----|--------|------|
| C-W19-001 | **FIXED** | `audit.deactivateUser` → `SessionRevoker.RevokeAllUserSessions` = `invalidateUserSessions` + `publishSessionEnded(..., "revoked")`; wiring `wireAuth` `SetSessionRevoker`; IT `TestDeactivateUser_LiveNATS_UnregistersWS_C_W19` |
| C-W19-003 | **FIXED** | **fail-closed admin error** при fail kill (`return fmt.Errorf("revoke sessions on deactivate: %w", err)`); тесты `*_FailClosed_C_W19_003` |

**Mutation:** убрать `publishSessionEnded` из `RevokeAllUserSessions` → RED (`TestRevokeAll…` NATS timeout + `TestDeactivateUser_LiveNATS…` condition never satisfied) → restore GREEN.

**Regression:** `go test -tags hubtest -race -count=1 ./testframework/audit/ ./testframework/auth/ ./testframework/ws/` PASS.

**Chosen path C-W19-003:** fail-closed (admin error), не silent success + best-effort.
