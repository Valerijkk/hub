# A2W7-4 — storage+dashboard+backup (Stage A2)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Baseline: `A7-4-storage-dash-backup.md` · Assignments: `W7-ASSIGNMENTS.md` §A4

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/storage/export_hooks.go` | 80 | 0 | чисто |
| `backend/internal/storage/filename.go` | 35 | 0 | чисто |
| `backend/internal/storage/handler.go` | 283 | 1 | замечания |
| `backend/internal/storage/orphan_cleanup.go` | 88 | 1 | замечания |
| `backend/internal/storage/repository.go` | 175 | 1 | замечания |
| `backend/internal/storage/service.go` | 315 | 1 | замечания |
| `backend/testframework/storage/confirm_a4006_test.go` | 90 | 0 | чисто |
| `backend/testframework/storage/download_a4007_test.go` | 54 | 0 | чисто |
| `backend/testframework/storage/filename_test.go` | 58 | 0 | чисто |
| `backend/testframework/storage/handler_extended_test.go` | 100 | 0 | чисто |
| `backend/testframework/storage/handler_routes_test.go` | 181 | 0 | чисто |
| `backend/testframework/storage/handler_wave117_test.go` | 209 | 0 | чисто |
| `backend/testframework/storage/orphan_cleanup_test.go` | 61 | 0 | чисто |
| `backend/testframework/storage/repository_test.go` | 109 | 0 | чисто |
| `backend/testframework/storage/service_ent_test.go` | 280 | 0 | чисто |
| `backend/testframework/storage/service_test.go` | 25 | 0 | чисто |
| `backend/testframework/storage/service_wave117_test.go` | 42 | 1 | замечания |
| `backend/internal/dashboard/dto.go` | 126 | 0 | чисто |
| `backend/internal/dashboard/export_hooks.go` | 32 | 0 | чисто |
| `backend/internal/dashboard/handler.go` | 284 | 1 | замечания |
| `backend/internal/dashboard/provisioner.go` | 23 | 0 | чисто |
| `backend/internal/dashboard/repository.go` | 358 | 1 | замечания |
| `backend/internal/dashboard/service.go` | 527 | 1 | замечания |
| `backend/testframework/dashboard/handler_test.go` | 345 | 0 | чисто |
| `backend/testframework/dashboard/repository_agg_a13003_test.go` | 84 | 0 | чисто |
| `backend/testframework/dashboard/repository_extra_test.go` | 93 | 0 | чисто |
| `backend/testframework/dashboard/repository_test.go` | 113 | 0 | чисто |
| `backend/testframework/dashboard/service_extended_test.go` | 299 | 0 | чисто |
| `backend/testframework/dashboard/service_test.go` | 95 | 0 | чисто |
| `backend/testframework/dashboard/service_wave107_test.go` | 173 | 0 | чисто |
| `backend/internal/backup/dto.go` | 18 | 0 | чисто |
| `backend/internal/backup/export_hooks.go` | 113 | 0 | чисто |
| `backend/internal/backup/handler.go` | 130 | 1 | замечания |
| `backend/internal/backup/maintenance.go` | 63 | 1 | замечания |
| `backend/internal/backup/repository.go` | 92 | 1 | замечания |
| `backend/internal/backup/service.go` | 441 | 1 | замечания |
| `backend/testframework/backup/coverage_wave80_test.go` | 82 | 0 | чисто |
| `backend/testframework/backup/dump_sanitize_a7007_test.go` | 40 | 0 | чисто |
| `backend/testframework/backup/handler_test.go` | 153 | 1 | замечания |
| `backend/testframework/backup/maintenance_a7003_test.go` | 75 | 0 | чисто |
| `backend/testframework/backup/prerestore_a7009_test.go` | 27 | 0 | чисто |
| `backend/testframework/backup/repository_test.go` | 85 | 0 | чисто |
| `backend/testframework/backup/restore_secrets_a7007_test.go` | 48 | 0 | чисто |
| `backend/testframework/backup/service_extended_test.go` | 188 | 0 | чисто |
| `backend/testframework/backup/service_test.go` | 61 | 0 | чисто |
| `backend/testframework/backup/service_wave100_test.go` | 151 | 0 | чисто |
| `backend/testframework/backup/service_wave98_test.go` | 60 | 0 | чисто |
| `backend/testframework/backup/service_wave99_test.go` | 124 | 0 | чисто |

Файлов в списке: **48**. Прочитано: **48**. Пропущено: **0**.

## Находки

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект / статус |
|----|-----|-----|-------------|---------|------------------------|
| W7-A4-001 | FIXED | P1 | `storage/service.go` GenerateUploadURL | NF-09: `size==0` legacy **без** ContentLength bind | **FIXED TASK-099**: reject size≤0 + always ContentLength; ADR-STORAGE-NF09-SIZE-2026; тест `nf09_size_zero_task099_test.go` |
| W7-A4-002 | RESIDUAL | P1 | `storage/service.go:232-262` | Confirm entity scope: `issue` проверен; `message`/`avatar`/… без project match | Частичный фикс TASK-024/045; чужая привязка non-issue типов |
| W7-A4-003 | RESIDUAL | P1 | `backup/maintenance.go:13-15` | Gate = process-local `atomic.Bool` | W12 U04; multi-replica K3s не блокируется |
| W7-A4-004 | FIXED-OK | P1 | `backup/service.go:104-110` | MarkCompleted без S3 | Закрыто TASK-024: fail-closed MarkFailed |
| W7-A4-005 | FIXED-OK | P2 | `storage/service.go:159-162` | Confirm без HEAD | Закрыто TASK-045: `assertObjectWithinLimit` / ErrObjectNotFound |
| W7-A4-006 | FIXED-OK | P2 | `storage/service.go:207-217` | contentType без allowlist | Закрыто TASK-045: HTML/SVG/JS reject |
| W7-A4-007 | RESIDUAL | P2 | `backup/service.go:420-422` | `markFailed` пишет `err.Error()` (pg_dump wrap) в ErrorMessage | Host/DSN фрагменты админу |
| W7-A4-008 | RESIDUAL | P2 | `dashboard/repository.go:269-272` | `ListProjectsWithIssues` = WithIssues().All() | Unbounded memory на admin top |
| W7-A4-009 | RESIDUAL | P2 | `dashboard/service.go:178` | Repo/Redis err → всё равно `return data, nil` | Silent 200 с пустыми виджетами |
| W7-A4-010 | FIXED-OK | P2 | `dashboard/service.go:143-169` | today через process TZ | Закрыто TASK-045: UTC midnight |
| W7-A4-011 | RESIDUAL | P2 | `backup/handler.go` restore/delete | NotFound → generic 500 | Клиент не отличает 404 |
| W7-A4-012 | RESIDUAL | P2 | `backup/handler_test.go` | adminMw stubbed Next(); нет denial | Регресс снятия RequireAdmin зелёный |
| W7-A4-013 | RESIDUAL | P2 | `backup/service.go:347` | pg_dump/mongodump URI с паролем в argv | `/proc` / auditd leak |
| W7-A4-014 | FIXED-OK | P3 | storage SoftDelete | Без DeletedAtIsNil | Закрыто TASK-061 path |
| W7-A4-015 | FIXED-OK | P3 | orphan_cleanup | Один батч 500 | Закрыто TASK-067: drain loop |
| W7-A4-016 | FIXED-OK | P3 | dashboard ApplyTemplate/SaveLayout | Без membership | Закрыто TASK-067: IsProjectMember |
| W7-A4-017 | RESIDUAL | P3 | `dashboard/handler.go` Admin | Handler читает repo.GetUser | Layering |
| W7-A4-018 | RESIDUAL | P3 | backup SetRestoredNote | Пишет в ErrorMessage | Путает fail и restore note |
| W7-A4-019 | FIXED-OK | P3 | storage/backup `_ =` cleanup | Silent | Закрыто TASK-057: slog |
| W7-A4-020 | FIXED-OK | P3 | service_wave117 orphan assert | `_ = err` | Закрыто TASK-058 |

## Чисто (инварианты держат)

| Зона | Статус |
|------|--------|
| aws-sdk-go-v2 / path-style SeaweedFS | Держит |
| Presign + CR-04 file project RBAC | Держит |
| Filename path traversal sanitize | Держит |
| Backup admin wiring RequireAdmin | Держит (тест denial — gap) |
| Restore pre-backup + LimitReader ZIP | Держит |
| Dashboard membership before cache (A7-001) | Держит |

## Counts

| Метрика | n |
|---------|--:|
| Файлов | 48 |
| Open (RESIDUAL) | **11** (вкл. partial P1 001/002 + U04 003) |
| FIXED-OK | 9 |
| NEW | 0 |
| REGRESS | 0 |
| Open by sev | P0:0 · P1:3 · P2:6 · P3:2 |
