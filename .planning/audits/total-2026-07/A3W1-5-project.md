# A3 W1 агент 5 — project (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W1-5-project.md](A2W1-5-project.md) · A1: [A1-5-project.md](A1-5-project.md)  
> Периметр: 38 файлов (10 prod + 28 tests) · Прочитано: **38/38** · Пропущено: **0**

## Сверка с A2 (A2W1-A5-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W1-A5-001 | FIXED-OK | закрыт | TASK-122: `SoftDeleteWithCrossLinkCleanup` Ent-tx; fail-closed unit `TestDeleteProject_Cleanup*Error_DoesNotSoftDelete` |
| A2W1-A5-002 | FIXED-OK | закрыт | TASK-088: `AddTeamToProject` AddTeamIDs+AddMember в одной Ent-tx + mid-fail hook/rollback |
| A2W1-A5-003 | RESIDUAL | живо | `RemoveTeamMember` без защиты последнего участника; `TestService_Wave122_RemoveTeamMember` даже фиксирует empty team как OK |

**REGRESS:** 0

### Сверка с A1 (кратко)

| A1 ID | Tag на A3 |
|-------|-----------|
| W1-A5-001…007, 009…012 | FIXED-OK (без регрессии) |
| W1-A5-008 | FIXED-OK через A2W1-A5-001 / TASK-122 (больше не residual) |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/project/dto.go | 93 | 0 | чисто |
| backend/internal/project/events.go | 19 | 0 | чисто |
| backend/internal/project/export_hooks.go | 39 | 0 | чисто |
| backend/internal/project/handler.go | 653 | 0 | чисто |
| backend/internal/project/invite.go | 97 | 1 | замечания |
| backend/internal/project/provisioner.go | 65 | 0 | чисто |
| backend/internal/project/repository.go | 671 | 1 | замечания |
| backend/internal/project/service.go | 561 | 3 | замечания |
| backend/internal/project/template_export.go | 134 | 1 | замечания |
| backend/internal/project/templates.go | 283 | 1 | замечания |
| backend/testframework/project/handler_flow_test.go | 389 | 0 | чисто |
| backend/testframework/project/handler_wave110_test.go | 102 | 0 | чисто |
| backend/testframework/project/handler_wave113_test.go | 65 | 0 | чисто |
| backend/testframework/project/handler_wave120_test.go | 60 | 0 | чисто |
| backend/testframework/project/handler_wave122_routes_test.go | 194 | 0 | чисто |
| backend/testframework/project/handler_wave122_test.go | 43 | 0 | чисто |
| backend/testframework/project/provisioner_test.go | 80 | 0 | чисто |
| backend/testframework/project/repository_ent_extended_test.go | 135 | 0 | чисто |
| backend/testframework/project/repository_ent_test.go | 217 | 0 | чисто |
| backend/testframework/project/repository_wave124_extra_test.go | 23 | 0 | чисто |
| backend/testframework/project/repository_wave124_test.go | 48 | 0 | чисто |
| backend/testframework/project/service_extended_test.go | 332 | 0 | чисто |
| backend/testframework/project/service_fa11_test.go | 106 | 0 | чисто |
| backend/testframework/project/service_test.go | 642 | 0 | чисто |
| backend/testframework/project/service_wave101_test.go | 211 | 0 | чисто |
| backend/testframework/project/service_wave105_test.go | 105 | 0 | чисто |
| backend/testframework/project/service_wave113_test.go | 115 | 0 | чисто |
| backend/testframework/project/service_wave120_test.go | 50 | 0 | чисто |
| backend/testframework/project/service_wave122_extra_test.go | 201 | 1 | замечания |
| backend/testframework/project/service_wave122_test.go | 58 | 0 | чисто |
| backend/testframework/project/service_wave124_test.go | 84 | 0 | чисто |
| backend/testframework/project/template_export_ent_test.go | 45 | 0 | чисто |
| backend/testframework/project/template_handler_fa11_test.go | 25 | 0 | чисто |
| backend/testframework/project/template_handler_routes_test.go | 96 | 0 | чисто |
| backend/testframework/project/template_handler_wave101_test.go | 73 | 0 | чисто |
| backend/testframework/project/template_multipart_test.go | 17 | 1 | замечания |
| backend/testframework/project/templates_test.go | 151 | 0 | чисто |
| backend/testframework/project/templates_wave105_test.go | 56 | 0 | чисто |

Файлов в списке: 38. Прочитано: 38. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A2W1-A5-003 | RESIDUAL | P3 | `service.go:565-573` | `RemoveTeamMember`: любой member удаляет любого; нет запрета «последний участник» | Пустая команда → никто не проходит `requireTeamMember` → orphan team навсегда; тест `service_wave122_extra_test.go:103-107` закрепляет empty как success | Перед remove: `TeamMemberCount`; если ≤1 → 400; либо admin override |
| A3W1-A5-001 | FIXED | P2 | `repository.go` / `service.go` Archive/Unarchive/Restore | ~~ClearDeletedAt на любом UpdateStatus(false)~~ | Soft-deleted resurrect через archive | TASK-137: guard deleted + `RestoreFromTrash` |
| A3W1-A5-002 | FIXED | P2 | `service.go` PermanentDelete / `handler.go` | ~~hard-delete active без admin~~ | TZ Admin/GDPR | TASK-137: `status=deleted` + `IsSystemAdmin` |
| A3W1-A5-003 | FIXED | P3 | `repository.go` CreateTeamWithCreator | ~~CreateTeam+AddMember вне tx~~ | orphan team | TASK-137: Ent-tx + mid-fail hook |
| A3W1-A5-004 | FIXED | P3 | `invite.go` JoinByInvite | ~~MemberCount err → 0~~ | ложный UI | TASK-137: slog + NewInternal |
| A3W1-A5-005 | FIXED | P3 | `templates.go` / `template_export.go` | ~~err.Error() в Fiber~~ | leak internals | TASK-137: `apperr.NewInternal` |
| A3W1-A5-006 | FIXED | P3 | `a3w1_a5_task137_test.go` | ~~нет Test* >25MB~~ | регрессия 413 | TASK-137: oversize → 413 |

## Чисто (подтверждено после remediation)

- P0 RBAC project CRUD/members/invites — `handler.go:24-58` + `rbacSvc`; Fiber v3 order OK.
- RevokeInvite scoped `DeleteInvite(projectID, inviteID)` + soft revoke (`revoked_at`).
- Last-owner demote/remove; Get deleted/non-member → 404; avatar `validateProjectAvatarURL` http(s).
- Teams list by membership; AddTeamToProject — `requireTeamMember` + `PermTeamManage`.
- Template DeleteTemplate — author/admin; SaveAsTemplate — membership; import `LimitReader` + size check.
- HI-06 `TryConsumeInvite` атомарный; statuses/labels fail-closed; invite Link + `resolveFrontendURL`.
- A2W1-A5-001/002 evidence: tx cleanup + AddTeam tx + unit/Ent tests PASS-пути в wave101 / TASK-088.
- **TASK-137 (2026-07-29):** A3W1-A5-001…006 FIXED — hubtest `./testframework/project/` PASS; evidence в `.planning/backlog/TASK-137-2026-07-29-a3-1-5-project.md`.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 1 (A2 residual) |
| **Total active** | **1** |
| FIXED (TASK-137) | **6** |

A2 baseline: 3 активных → FIXED-OK 2 · RESIDUAL 1 · NEW 6 → FIXED 6 (TASK-137) · REGRESS 0.

**Топ ID (осталось):** A2W1-A5-003 (orphan team last member) — вне TASK-137.