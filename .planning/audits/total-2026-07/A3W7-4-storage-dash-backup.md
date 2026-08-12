# A3 W7 агент A4 — storage+dashboard+backup (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W7-4-storage-dash-backup.md](A2W7-4-storage-dash-backup.md) · A1: [A7-4-storage-dash-backup.md](A7-4-storage-dash-backup.md)  
> Assignment: [W7-ASSIGNMENTS.md](W7-ASSIGNMENTS.md) § A4 · Периметр: **48** · Прочитано: **48/48** · Пропущено: **0**  
> Фокус: SeaweedFS aws-sdk-go-v2 presigned · BodyLimit/NF-09 · backup MaintenanceGate · dashboard RBAC  
> Remediation: TASK-084/099/100/119 (+ earlier 024/045/061/067)

## Сверка с A2 (W7-A4-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W7-A4-001 | FIXED-OK | закрыт | `GenerateUploadURL`: `size<=0` → `ErrInvalidFileSize`; always `ContentLength`; `nf09_size_zero_task099_test.go`; BodyLimit 25MiB в `wiring_fiber.go:45` = `maxUploadBytes` |
| W7-A4-002 | FIXED-OK | закрыт | `validateEntityScope` issue/avatar/channel/message/voice + membership; `confirm_entity_scope_task084_test.go` |
| W7-A4-003 | FIXED-OK | закрыт | Redis `hub:backup:maintenance` + `NewRedisMaintenanceGate`; IT `maintenance_distributed_task100_test.go`; wiring `wiring_fiber.go:68-69` |
| W7-A4-004 | FIXED-OK | закрыт | `s3Client == nil` → `markFailed`, не MarkCompleted (`service.go:105-109`) |
| W7-A4-005 | FIXED-OK | закрыт | `assertObjectWithinLimit` HEAD + `ErrObjectNotFound` |
| W7-A4-006 | FIXED-OK | закрыт | `validateUploadContentType` HTML/SVG/JS reject |
| W7-A4-007 | FIXED-OK | закрыт | `sanitizeBackupErrorMessage` + `markFailed` (`sanitize.go`, вне assignment; export в `export_hooks.go`) |
| W7-A4-008 | FIXED-OK | закрыт | `ListTopProjectsByIssueCount` GROUP BY; `task119_hygiene_test.go` |
| W7-A4-009 | FIXED-OK (partial) | Personal/Admin-top закрыты | Personal PG widgets fail-visible; TopProjects fail-closed. **Хвост:** ProjectStats soft-fail → A3W7-4-001 |
| W7-A4-010 | FIXED-OK | закрыт | UTC midnight для timeToday (`service.go:141-165`) |
| W7-A4-011 | FIXED-OK | закрыт | Restore/Delete → `ErrNotFound` → 404 |
| W7-A4-012 | FIXED-OK | закрыт | RequireAdmin denial: `task119_hygiene_test.go` (реальный `rbac.RequireAdmin`); `handler_test.go` stub Next остаётся happy-path only |
| W7-A4-013 | FIXED-OK | закрыт | `pgDumpArgsEnv` PGPASSWORD + mongo `--config` |
| W7-A4-014 | FIXED-OK | закрыт | SoftDelete `DeletedAtIsNil` + affected==0 → NotFound |
| W7-A4-015 | FIXED-OK | закрыт | orphan drain loop + stop-on-zero-deleted |
| W7-A4-016 | FIXED-OK | закрыт | SaveLayout/ApplyTemplate `IsProjectMember` |
| W7-A4-017 | FIXED-OK | закрыт | Admin через `RequireSystemAdmin` (не repo из handler) |
| W7-A4-018 | FIXED-OK | закрыт | `SetRestoredAt` / `restored_at` column |
| W7-A4-019 | FIXED-OK (partial) | storage slog OK | cleanup pending/presign — slog. **Хвост:** backup DeleteObject `_, _` → A3W7-4-002 |
| W7-A4-020 | FIXED-OK | закрыт | orphan wave117 assert meta remains on S3 fail |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/storage/export_hooks.go` | 80 | 0 | чисто |
| `backend/internal/storage/filename.go` | 35 | 0 | чисто |
| `backend/internal/storage/handler.go` | 283 | 1 | Delete SoftDelete→500 |
| `backend/internal/storage/orphan_cleanup.go` | 88 | 0 | drain FIXED-OK |
| `backend/internal/storage/repository.go` | 212 | 0 | scope/SoftDelete OK |
| `backend/internal/storage/service.go` | 357 | 0 | NF-09/presign/entity OK |
| `backend/testframework/storage/confirm_a4006_test.go` | 104 | 0 | чисто |
| `backend/testframework/storage/download_a4007_test.go` | 54 | 0 | чисто |
| `backend/testframework/storage/filename_test.go` | 58 | 0 | чисто |
| `backend/testframework/storage/handler_extended_test.go` | 100 | 0 | чисто |
| `backend/testframework/storage/handler_routes_test.go` | 195 | 0 | чисто |
| `backend/testframework/storage/handler_wave117_test.go` | 209 | 0 | чисто |
| `backend/testframework/storage/orphan_cleanup_test.go` | 61 | 0 | чисто |
| `backend/testframework/storage/repository_test.go` | 109 | 0 | чисто |
| `backend/testframework/storage/service_ent_test.go` | 294 | 0 | чисто |
| `backend/testframework/storage/service_test.go` | 25 | 0 | чисто |
| `backend/testframework/storage/service_wave117_test.go` | 42 | 0 | assert FIXED-OK |
| `backend/internal/dashboard/dto.go` | 126 | 0 | чисто |
| `backend/internal/dashboard/export_hooks.go` | 32 | 0 | чисто |
| `backend/internal/dashboard/handler.go` | 284 | 0 | Admin/Project RBAC OK |
| `backend/internal/dashboard/provisioner.go` | 23 | 0 | чисто |
| `backend/internal/dashboard/repository.go` | 413 | 1 | burndown All() без cap |
| `backend/internal/dashboard/service.go` | 497 | 2 | Project soft-fail; GetLayout ACL |
| `backend/testframework/dashboard/handler_test.go` | 345 | 0 | чисто |
| `backend/testframework/dashboard/repository_agg_a13003_test.go` | 84 | 0 | чисто |
| `backend/testframework/dashboard/repository_extra_test.go` | 93 | 0 | чисто |
| `backend/testframework/dashboard/repository_test.go` | 113 | 0 | чисто |
| `backend/testframework/dashboard/service_extended_test.go` | 299 | 0 | чисто |
| `backend/testframework/dashboard/service_test.go` | 95 | 0 | чисто |
| `backend/testframework/dashboard/service_wave107_test.go` | 173 | 0 | чисто |
| `backend/internal/backup/dto.go` | 19 | 0 | restored_at OK |
| `backend/internal/backup/export_hooks.go` | 133 | 0 | sanitize/pg hooks OK |
| `backend/internal/backup/handler.go` | 139 | 0 | 404 FIXED-OK |
| `backend/internal/backup/maintenance.go` | 111 | 0 | Redis gate FIXED-OK |
| `backend/internal/backup/repository.go` | 97 | 0 | SetRestoredAt OK |
| `backend/internal/backup/service.go` | 471 | 1 | DeleteObject silent |
| `backend/testframework/backup/coverage_wave80_test.go` | 82 | 0 | чисто |
| `backend/testframework/backup/dump_sanitize_a7007_test.go` | 40 | 0 | чисто |
| `backend/testframework/backup/handler_test.go` | 153 | 0 | stub admin — denial вне файла |
| `backend/testframework/backup/maintenance_a7003_test.go` | 75 | 0 | чисто |
| `backend/testframework/backup/prerestore_a7009_test.go` | 27 | 0 | чисто |
| `backend/testframework/backup/repository_test.go` | 88 | 0 | чисто |
| `backend/testframework/backup/restore_secrets_a7007_test.go` | 48 | 0 | чисто |
| `backend/testframework/backup/service_extended_test.go` | 188 | 0 | чисто |
| `backend/testframework/backup/service_test.go` | 61 | 0 | чисто |
| `backend/testframework/backup/service_wave100_test.go` | 151 | 0 | чисто |
| `backend/testframework/backup/service_wave98_test.go` | 60 | 0 | чисто |
| `backend/testframework/backup/service_wave99_test.go` | 124 | 0 | чисто |

Файлов в списке: **48**. Прочитано: **48**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W7-4-001 | RESIDUAL (хвост W7-A4-009) | P2 | `dashboard/service.go:207-216`, `221-224`, `255-258`, `274-277`; Admin counts `323-346` | `GetProjectStats`: ошибка `AggregateProjectIssueStats`/burndown/testRuns/activity → slog + пустые виджеты, `return stats, nil` (200). Admin `CountUsers`/DAU и т.п. — тот же soft-fail; только TopProjects fail-closed | TASK-119 закрыл Personal fail-visible; project/admin counters всё ещё silent empty → UI «ноль задач» при деградации PG | Critical PG widgets (TasksByStatus/BugCount/Overdue/counts) → return err → 500; soft только для необязательных графиков |
| A3W7-4-002 | RESIDUAL (хвост W7-A4-019) | P3 | `backup/service.go:347-351`, `433-437` | `DeleteBackup` / `cleanupOldBackups`: `_, _ = s3Client.DeleteObject` без slog | Storage cleanup уже логирует; backup молча оставляет orphan ZIP в SeaweedFS при S3 err, DB row уже удалена | `slog.WarnContext` при err; опционально fail-closed DeleteBackup если S3 недоступен |
| A3W7-4-003 | NEW | P3 | `dashboard/service.go:458-481` vs `486-495` | `GetLayout` с `project_id` **без** `IsProjectMember`; SaveLayout/ApplyTemplate проверяют | Асимметрия ACL: stranger читает свой (пустой) layout для чужого project_id — сейчас без утечки чужих данных, но нарушает инвариант «project_id → membership» | Membership check в GetLayout при non-nil projectID → ErrForbidden |
| A3W7-4-004 | NEW | P3 | `storage/handler.go:278-280`; `repository.go:185-200` | SoftDelete NotFound (race/повтор) → handler всегда 500 | SoftDelete уже возвращает `*ent.NotFoundError`; Confirm/Download мапят NotFound→404, Delete — нет | `ent.IsNotFound(err)` → 404 |
| A3W7-4-005 | NEW | P3 | `dashboard/repository.go:204-212` | `ListProjectCompletedIssuesSince` = `.All()` без Limit за 30 дней | A13-003 закрыл issue-list агрегатом; burndown всё ещё материализует все done-issues периода → memory на крупных проектах | GROUP BY day / Limit / streaming aggregate как у `AggregateProjectIssueStats` |

## Чисто (фокус-инварианты)

| Зона | Статус |
|------|--------|
| aws-sdk-go-v2 + path-style SeaweedFS | Держит (`service.go:73-76`) |
| Presign PUT + ContentLength bind + size 1…25MiB | Держит (NF-09 / ADR-STORAGE-NF09) |
| Fiber BodyLimit 25MiB ≡ maxUploadBytes | Держит (`wiring_fiber.go:45`) |
| CR-04 file project RBAC (upload/confirm/download/delete) | Держит |
| Filename path traversal sanitize | Держит |
| Confirm entity scope (issue/avatar/channel/message/voice) | Держит (TASK-084) |
| Backup admin wiring RequireAdmin | Держит (denial evidence вне assignment) |
| MaintenanceGate Redis cluster-wide | Держит (TASK-100 / ADR U04) |
| Restore pre-backup fail-closed + LimitReader ZIP | Держит |
| Dashboard Project membership before cache (A7-001) | Держит |
| Dashboard Admin RequireSystemAdmin | Держит |
| pg_dump/mongodump secrets not in argv | Держит |

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 4 |
| **Total open** | **5** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 closes) | 18 (+ 2 partial → residual) |
| RESIDUAL | 2 (`A3W7-4-001`, `A3W7-4-002`) |
| NEW | 3 (`A3W7-4-003`…`005`) |
| REGRESS | 0 |

A2 baseline: 20 findings (многие FIXED в A2) → **FIXED-OK ~18** · RESIDUAL 2 · NEW 3 · REGRESS 0 · open **5** (P2=1 · P3=4).

**Топ ID:** A3W7-4-001 (ProjectStats soft-fail) · A3W7-4-005 (burndown All) · A3W7-4-002 (silent S3 delete).
