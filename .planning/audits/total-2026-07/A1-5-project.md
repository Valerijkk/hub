# W1 агент A5 — project

> TASK-008 · READ-ONLY · 2026-07-28  
> Сверка с `grok-2026-07/03-tracking.md`: A3-023 **подтверждён** на текущем коде; invite consume (HI-06) и `LimitReader` на import — **чисты**.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/project/dto.go` | 106 | 1 | замечания |
| `backend/internal/project/events.go` | 22 | 0 | чисто |
| `backend/internal/project/export_hooks.go` | 45 | 0 | чисто |
| `backend/internal/project/handler.go` | 681 | 3 | замечания |
| `backend/internal/project/invite.go` | 112 | 0 | чисто |
| `backend/internal/project/provisioner.go` | 78 | 0 | чисто |
| `backend/internal/project/repository.go` | 583 | 2 | замечания |
| `backend/internal/project/service.go` | 527 | 3 | замечания |
| `backend/internal/project/template_export.go` | 161 | 0 | чисто |
| `backend/internal/project/templates.go` | 304 | 1 | замечания |
| `backend/testframework/project/handler_flow_test.go` | 460 | 1 | замечания |
| `backend/testframework/project/handler_wave110_test.go` | 122 | 0 | чисто |
| `backend/testframework/project/handler_wave113_test.go` | 77 | 0 | чисто |
| `backend/testframework/project/handler_wave120_test.go` | 72 | 0 | чисто |
| `backend/testframework/project/handler_wave122_routes_test.go` | 230 | 0 | чисто |
| `backend/testframework/project/handler_wave122_test.go` | 53 | 0 | чисто |
| `backend/testframework/project/provisioner_test.go` | 98 | 0 | чисто |
| `backend/testframework/project/repository_ent_extended_test.go` | 164 | 0 | чисто |
| `backend/testframework/project/repository_ent_test.go` | 262 | 0 | чисто |
| `backend/testframework/project/repository_wave124_extra_test.go` | 32 | 0 | чисто |
| `backend/testframework/project/repository_wave124_test.go` | 60 | 0 | чисто |
| `backend/testframework/project/service_extended_test.go` | 377 | 0 | чисто |
| `backend/testframework/project/service_fa11_test.go` | 119 | 0 | чисто |
| `backend/testframework/project/service_test.go` | 606 | 1 | замечания |
| `backend/testframework/project/service_wave101_test.go` | 153 | 0 | чисто |
| `backend/testframework/project/service_wave105_test.go` | 128 | 0 | чисто |
| `backend/testframework/project/service_wave113_test.go` | 123 | 0 | чисто |
| `backend/testframework/project/service_wave120_test.go` | 63 | 0 | чисто |
| `backend/testframework/project/service_wave122_extra_test.go` | 251 | 0 | чисто |
| `backend/testframework/project/service_wave122_test.go` | 74 | 0 | чисто |
| `backend/testframework/project/service_wave124_test.go` | 108 | 0 | чисто |
| `backend/testframework/project/template_export_ent_test.go` | 58 | 0 | чисто |
| `backend/testframework/project/template_handler_fa11_test.go` | 31 | 0 | чисто |
| `backend/testframework/project/template_handler_routes_test.go` | 117 | 0 | чисто |
| `backend/testframework/project/template_handler_wave101_test.go` | 83 | 0 | чисто |
| `backend/testframework/project/template_multipart_test.go` | 20 | 1 | замечания |
| `backend/testframework/project/templates_test.go` | 174 | 0 | чисто |
| `backend/testframework/project/templates_wave105_test.go` | 66 | 0 | чисто |

Файлов в списке: 38. Прочитано: 38. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W1-A5-001 | P0 | `handler.go:23-57` (wiring: `wiring_core.go:28`) | `RegisterRoutes` принимает только `authMw`; на CRUD/members/invites/teams **нет** `rbac.RequirePermission`. Комментарий упоминает `requirePerm`, параметра нет. В prod: `projectHandler.RegisterRoutes(api, requireAuth)` — без `rbacSvc` | В `rbac` есть `PermProjectView/Edit/Delete/ManageMembers/...`, tracker/tms их используют. Здесь любой аутентифицированный пользователь с UUID проекта может soft-delete, archive, менять members/roles (в т.ч. на `owner`), создавать invites. Потеря/захват проекта | Расширить `RegisterRoutes(api, authMw, rbacSvc)`; навесить `PermProject*` по маршрутам (Fiber v3: middleware после handler). Service: defense-in-depth membership/perm. Тест: stranger → 403 |
| W1-A5-002 | P1 | `handler.go:505-516`; `service.go:524-525`; `repository.go:402-404` | `RevokeInvite` парсит только `inviteId`; `:projectId` из URL **не используется**. `DeleteInvite` — `DeleteOneID` без `HasProjectWith` | Классический IDOR (как A3-001/002): при любом доступе к своему проекту можно отозвать чужой invite по UUID. Даже после фикса RBAC на `:projectId` дыра останется | `DeleteInvite(ctx, projectID, inviteID)` с `Where(ID, HasProjectWith(projectID))`; mismatch → 404. Handler передаёт оба ID |
| W1-A5-003 | P1 | `service.go:364-383` vs `service.go:340-347` | `UpdateMemberRole` меняет роль без проверки «последний owner». `RemoveMember` такую проверку делает | Демоут единственного owner → проект без управляющего; либо чужой (при W1-A5-001) ставит себя owner | Перед сменой с `owner` → не-owner: `CountOwners`; если ≤1 → 400. Запретить повышение в owner без `PermProjectManageRoles` |
| W1-A5-004 | P1 | `handler.go:550-680`; `service.go:444-497` | Teams: create/list/add/remove member/add-to-project — только auth, без владельца команды и без perm на целевой project | Любой auth-юзер видит все команды, добавляет людей, bulk-вносит команду в чужой проект (`AddTeamToProject`) | Модель ownership/ACL для Team; `AddTeamToProject` требует `PermProjectManageMembers` на `projectId`; list — scope по membership/org-policy |
| W1-A5-005 | P1 | `templates.go:101-116`; `templates.go:207-214` | `DeleteTemplate` блокирует только `IsSystem`; нет проверки `created_by` / admin. Routes — только `authMw` | Любой auth-юзер удаляет чужой custom-шаблон | Разрешить delete автору или admin; иначе 403. Опционально RBAC на export/import |
| W1-A5-006 | P2 | `service.go:85-103`; `repository.go:106-108` | `GetProject`/`FindByID` не фильтруют `status!=deleted` и не требуют membership (myRole пустой ок) | Утечка метаданных soft-deleted/чужого проекта по UUID (усиливает W1-A5-001) | 404 если не member / status deleted (кроме restore). `FindByID` для мутаций — явный pre-check |
| W1-A5-007 | P2 | `dto.go:16`; `repository.go:202-214`; `service.go:129-138` | `AvatarURL` без `validate` URL; Update пишет строку as-is | Stored XSS / опасный `javascript:`/`data:` в UI при рендере `avatar_url` без санитизации | `validate:"omitempty,url"` + allowlist http(s); или `ValidateOutboundURL` при set |
| W1-A5-008 | P2 | `service.go:176-185`; `service.go:192-225` | `cleanupCrossProjectLinks` best-effort (log+continue); затем soft-delete всё равно выполняется | Часть cross-project links остаётся после «удаления» проекта | Tx: delete links + UpdateStatus; или fail DeleteProject при ошибке cleanup |
| W1-A5-009 | P3 | `repository.go:574-580` | **A3-023 (открыт):** `_, _ = r.AddMember(...)` в `AddTeamToProject` | Не-constraint ошибки БД глотаются → частичный membership без ошибки клиенту | Игнорировать только `ent.IsConstraintError`; иначе `return err`. Не в одной tx с `AddTeamIDs` — обернуть в tx |
| W1-A5-010 | P3 | `repository.go:516-522` | `GetProjectWithWorkflowStatusesAndLabels`: `statuses, _ :=` / `labels, _ :=` | Ошибка загрузки статусов/меток → тихий пустой шаблон | Логировать warn; при err ≠ NotFound — return err |
| W1-A5-011 | P3 | `service.go:438`; `invite.go:40` | `toInviteResponse(..., "")` — Link = `/invite/{code}` без frontend base | Клиенты получают битую invite-ссылку | Прокинуть `frontendURL` из конфига (как в export_hooks) |
| W1-A5-012 | P2 | `handler_flow_test.go:22-36`; `service_test.go:343-345`; `template_multipart_test.go:12-19` | Нет тестов: stranger→403 на project mutate; DB-error в `AddTeamToProject`; oversize import (>25MB). `mockRepo.AddTeamToProject` всегда `nil` — A3-023 не ловится. multipart — только helper, без `Test*` | Регрессия authz/partial-add/LimitReader не поймается; guard в repo, тест на моке зелёный на сломанном коде | Ent-тесты: foreign user 403; `AddMember` inject non-constraint err → fail; multipart body > `maxTemplateImportBytes` → 413 |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `invite.go` + `repository.go:368-393` | HI-06 `TryConsumeInvite` атомарный (`UseCountLT`); expire/maxUses/already-member; documented lost slot при fail `AddMember` — осознанный tradeoff |
| `template_export.go:147-153` | `io.LimitReader` + проверка размера; filename в `Content-Disposition` захардкожен (`template.json`) — path traversal нет; FormFile читается из stream |
| `templates.go:71-80` | `SaveProjectAsTemplate` проверяет membership (закрытый IDOR на статусы/метки) |
| `service.go:165-186` | Soft delete: `status=deleted` + `deleted_at`; hard `PermanentDelete` есть, в handler не экспонирован |
| `service.go:112-120`, `289-297` | `MemberCountBatch` — N+1 закрыт |
| `repository.go:146-199` | Create: tx + owner member + outbox; rollback на ошибках |
| `provisioner.go` | Best-effort provision, без panic; nil-safe |
| `events.go`, `export_hooks.go` | DTO событий / hubtest hooks — без логики безопасности |
| `service_wave122_extra_test.go` TryConsume / Join Ent | Лимит MaxUses на реальном repo покрыт |
| `repository_wave124_*` | Outbox publish + rollback on outbox fail |
| `templates_test.go` / wave105 | System delete forbidden; seed идемпотентен |
| Handler wave* invalid JSON/UUID | Парсинг ошибок покрыт; **не** заменяет authz-тесты |

## Сверка с grok-2026-07 / A3

| ID | Статус на 2026-07-28 |
|----|----------------------|
| A3-023 `AddTeamToProject` `_, _ = AddMember` | **OPEN** → W1-A5-009 |
| Invite atomic consume (HI-06) | **OK** |
| Template `LimitReader` | **OK** |
| `MemberCountBatch` | **OK** |
| Missing project-route RBAC | **не был в A3** — новая P0 W1-A5-001 |

## Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 1 |
| P1 | 4 |
| P2 | 4 |
| P3 | 3 |
| **Total** | **12** |

**Топ-3:** W1-A5-001 (нет RBAC на project routes) · W1-A5-002 (RevokeInvite IDOR) · W1-A5-003 (демоут последнего owner).
