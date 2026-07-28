# A2 W1 агент 5 — project (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc` (`0d6a659f`)  
> Baseline: [A1-5-project.md](A1-5-project.md)  
> Периметр: 38 файлов (10 prod + 28 tests) · Прочитано: 38 · Пропущено: 0

## Сверка с A1 (W1-A5-*)

| A1 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| W1-A5-001 | FIXED-OK | закрыт | `RegisterRoutes(..., rbacSvc)` + PermProject* (TASK-009) |
| W1-A5-002 | FIXED-OK | закрыт | `DeleteInvite(ctx, projectID, inviteID)` |
| W1-A5-003 | FIXED-OK | закрыт | last-owner demote guard |
| W1-A5-004 | FIXED-OK | закрыт | ListTeamsByUserID + requireTeamMember + PermTeamManage на AddTeamToProject |
| W1-A5-005 | FIXED-OK | закрыт | DeleteTemplate author/admin |
| W1-A5-006 | FIXED-OK | закрыт | Get deleted/non-member → 404 |
| W1-A5-007 | FIXED-OK | закрыт | `validateProjectAvatarURL` http(s) |
| W1-A5-008 | RESIDUAL | живо | `cleanupCrossProjectLinks` best-effort; soft-delete всё равно | TASK-038 out-of-scope |
| W1-A5-009 | FIXED-OK | частично | non-constraint AddMember → return err; **нет** tx с AddTeamIDs → NEW A2W1-A5-002 |
| W1-A5-010 | FIXED-OK | закрыт | statuses/labels fail-closed |
| W1-A5-011 | FIXED-OK | закрыт | invite Link + frontendURL |
| W1-A5-012 | FIXED-OK | закрыт | authz Ent-negative / avatar / invite tests |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/project/dto.go | 91 | 0 | чисто |
| backend/internal/project/events.go | 19 | 0 | чисто |
| backend/internal/project/export_hooks.go | 35 | 0 | чисто |
| backend/internal/project/handler.go | 630 | 0 | чисто |
| backend/internal/project/invite.go | 97 | 0 | чисто |
| backend/internal/project/provisioner.go | 65 | 0 | чисто |
| backend/internal/project/repository.go | 570 | 1 | замечания |
| backend/internal/project/service.go | 574 | 2 | замечания |
| backend/internal/project/template_export.go | 134 | 0 | чисто |
| backend/internal/project/templates.go | 283 | 0 | чисто |
| backend/testframework/project/handler_flow_test.go | 389 | 0 | чисто |
| backend/testframework/project/handler_wave110_test.go | 102 | 0 | чисто |
| backend/testframework/project/handler_wave113_test.go | 65 | 0 | чисто |
| backend/testframework/project/handler_wave120_test.go | 60 | 0 | чисто |
| backend/testframework/project/handler_wave122_routes_test.go | 194 | 0 | чисто |
| backend/testframework/project/handler_wave122_test.go | 43 | 0 | чисто |
| backend/testframework/project/provisioner_test.go | 80 | 0 | чисто |
| backend/testframework/project/repository_ent_extended_test.go | 135 | 0 | чисто |
| backend/testframework/project/repository_ent_test.go | 206 | 0 | чисто |
| backend/testframework/project/repository_wave124_extra_test.go | 23 | 0 | чисто |
| backend/testframework/project/repository_wave124_test.go | 48 | 0 | чисто |
| backend/testframework/project/service_extended_test.go | 331 | 0 | чисто |
| backend/testframework/project/service_fa11_test.go | 106 | 0 | чисто |
| backend/testframework/project/service_test.go | 609 | 0 | чисто |
| backend/testframework/project/service_wave101_test.go | 128 | 0 | чисто |
| backend/testframework/project/service_wave105_test.go | 105 | 0 | чисто |
| backend/testframework/project/service_wave113_test.go | 115 | 0 | чисто |
| backend/testframework/project/service_wave120_test.go | 50 | 0 | чисто |
| backend/testframework/project/service_wave122_extra_test.go | 201 | 0 | чисто |
| backend/testframework/project/service_wave122_test.go | 58 | 0 | чисто |
| backend/testframework/project/service_wave124_test.go | 84 | 0 | чисто |
| backend/testframework/project/template_export_ent_test.go | 45 | 0 | чисто |
| backend/testframework/project/template_handler_fa11_test.go | 25 | 0 | чисто |
| backend/testframework/project/template_handler_routes_test.go | 96 | 0 | чисто |
| backend/testframework/project/template_handler_wave101_test.go | 73 | 0 | чисто |
| backend/testframework/project/template_multipart_test.go | 17 | 0 | чисто |
| backend/testframework/project/templates_test.go | 151 | 0 | чисто |
| backend/testframework/project/templates_wave105_test.go | 56 | 0 | чисто |

Файлов в списке: 38. Прочитано: 38. Пропущено: 0.

## Находки (активные A2)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A2W1-A5-001 | RESIDUAL | P2 | `service.go:240-241`, `256-278` | `cleanupCrossProjectLinks` log+continue; затем soft-delete проекта всегда | Часть cross-project links остаётся после «удаления» (W1-A5-008; TASK-038 out-of-scope) | Tx: delete links + UpdateStatus; или fail DeleteProject при ошибке cleanup |
| A2W1-A5-002 | NEW | P2 | `repository.go:607-637` | `AddTeamIDs` выполняется **до** цикла `AddMember`; при mid-fail не-constraint — return err, но команда уже привязана | Частичный membership + team link без rollback (улучшение W1-A5-009 неполное) | Одна Ent-tx: AddTeamIDs + AddMembers; Rollback целиком |
| A2W1-A5-003 | NEW | P3 | `service.go:582-590` | `RemoveTeamMember`: любой member может удалить любого; нет защиты «последний участник» | Команда без участников → никто не может `AddTeamMember`/`AddTeamToProject` (requireTeamMember) — orphan team | Запрет remove последнего; или system-admin override |

## Чисто

- P0 RBAC на project CRUD/members/invites — подтверждён (`handler.go:24-58`).
- RevokeInvite + last-owner + template ownership + avatar + Get deleted — закрыты.
- HI-06 TryConsumeInvite атомарный; template LimitReader.
- Teams list scoped by membership; AddTeamToProject требует team member + PermTeamManage.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 1 |
| **Total** | **3** |

A1 baseline: 12 → FIXED-OK 11 · RESIDUAL 1 · NEW 2 · REGRESS 0 · (W1-A5-009 → FIXED + NEW follow-up).
