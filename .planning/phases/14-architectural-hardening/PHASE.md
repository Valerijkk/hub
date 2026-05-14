---
phase: 14-architectural-hardening
title: Architectural Hardening
status: planned
depends_on: [13-polish-qa]
plans:
  - id: 14-01
    file: 14-01-PLAN.md
    title: Split AuthService into Session/Login/Profile/RegistrationMode
    depends_on: []
    wave: 1
    autonomous: true
    files_modified:
      - backend/internal/auth/service.go
      - backend/internal/auth/service_session.go
      - backend/internal/auth/service_login.go
      - backend/internal/auth/service_profile.go
      - backend/internal/auth/service_registration.go
      - backend/internal/auth/handler.go
      - backend/internal/auth/service_test.go
    success_criteria_addressed: [1]
  - id: 14-02
    file: 14-02-PLAN.md
    title: Split CallService into CallRoomService + CallEventPublisher + CallChatService
    depends_on: []
    wave: 1
    autonomous: true
    files_modified:
      - backend/internal/calls/service.go
      - backend/internal/calls/service_room.go
      - backend/internal/calls/service_events.go
      - backend/internal/calls/service_chat.go
      - backend/internal/calls/repository.go
      - backend/internal/calls/handler.go
      - backend/internal/calls/service_test.go
    success_criteria_addressed: [2]
  - id: 14-03
    file: 14-03-PLAN.md
    title: Remove *ent.Client leak from Messenger/Audit/Search services
    depends_on: []
    wave: 1
    autonomous: true
    files_modified:
      - backend/internal/messenger/service.go
      - backend/internal/messenger/service_channel_invites.go
      - backend/internal/messenger/service_polls.go
      - backend/internal/messenger/ent_repository.go
      - backend/internal/messenger/repository.go
      - backend/internal/audit/service.go
      - backend/internal/audit/repository.go
      - backend/internal/search/service.go
      - backend/cmd/server/main.go
    success_criteria_addressed: [3]
  - id: 14-04
    file: 14-04-PLAN.md
    title: Test coverage uplift (>=15 new test functions)
    depends_on: []
    wave: 1
    autonomous: true
    files_modified:
      - backend/internal/apigateway/webhook_test.go
      - backend/internal/apigateway/webhook_signing_test.go
      - backend/internal/rbac/middleware_integration_test.go
      - backend/internal/auth/magiclink_test.go
      - backend/internal/platform/outbox/outbox_test.go
      - backend/internal/calls/recording_webhook_test.go
    success_criteria_addressed: [5, 6]
  - id: 14-05
    file: 14-05-PLAN.md
    title: Permission representation upgrade (additive, non-breaking)
    depends_on: []
    wave: 1
    autonomous: true
    files_modified:
      - backend/internal/rbac/permissions.go
      - backend/internal/rbac/custom_roles.go
      - backend/ent/schema/role.go
      - backend/internal/rbac/permissions_test.go
    success_criteria_addressed: [4]
---

# Phase 14: Architectural Hardening

## Goal
Расщепить god-объекты бэкенда (AuthService 667 строк, CallService 479 строк) согласно SRP,
вывести `*ent.Client` из MessengerService/AuditService/SearchService в Repository интерфейсы,
расширить Permission представление за пределы uint64, закрыть пробелы в покрытии тестами
критичных путей, обнаруженных в REVIEW.md (webhook SSRF/retry/signing, RBAC route integration,
magic-link brute-force lockout, outbox panic recovery, LiveKit webhook signature verification).

## Scope
- Pure backend refactor + tests. **Никаких новых фич, никаких UI изменений.**
- Все хендлеры (HTTP API) сохраняют существующие сигнатуры и поведение — back-compat обязателен.
- Каждый план — атомарный, обратимый через `git revert <commit>` без каскада.

## Wave structure
Все 5 планов — **wave 1, независимы** (нет общих изменяемых файлов кроме `cmd/server/main.go`,
который трогает только 14-03; этот файл — последовательный bottleneck, выполняется последним
в составе 14-03). Можно запускать в параллельных контекстах executor.

## Files-modified overlap matrix
| Plan  | auth/service.go | calls/service.go | messenger/* | audit/* | search/* | webhook_test.go | rbac/permissions.go | ent/schema/role.go | main.go |
|-------|-----------------|------------------|-------------|---------|----------|-----------------|---------------------|--------------------|---------|
| 14-01 | yes             |                  |             |         |          |                 |                     |                    |         |
| 14-02 |                 | yes              |             |         |          |                 |                     |                    |         |
| 14-03 |                 |                  | yes         | yes     | yes      |                 |                     |                    | yes     |
| 14-04 |                 |                  |             |         |          | yes             |                     |                    |         |
| 14-05 |                 |                  |             |         |          |                 | yes                 | yes                |         |

**Zero overlap** — все 5 планов могут выполняться полностью параллельно.

## Acceptance for phase as a whole
1. `cd backend && go build ./...` зелёный.
2. `cd backend && go test -count=1 ./...` зелёный.
3. `cd backend && go test -race ./...` без data race warnings.
4. Все 6 success criteria из ROADMAP отмечены ✓ в финальном SUMMARY.
5. Test count (`grep -rc "^func Test" backend/internal/ | awk -F: '{s+=$2} END{print s}'`) вырос на >=15.

## Rollback strategy
Каждый план — 1-3 атомарных коммита. Откат любого плана: `git revert -m 1 <merge-commit>`
или revert отдельных коммитов плана. Никаких schema-breaking миграций (14-05 — additive only).
