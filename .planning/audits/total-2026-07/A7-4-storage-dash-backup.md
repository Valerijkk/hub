# W7 агент A4 — storage+dashboard+backup

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: SeaweedFS `aws-sdk-go-v2` · presigned URLs · admin-only backup · path traversal.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/storage/export_hooks.go` | 42 | 0 | чисто |
| `backend/internal/storage/filename.go` | 32 | 0 | чисто |
| `backend/internal/storage/handler.go` | 233 | 2 | замечания |
| `backend/internal/storage/orphan_cleanup.go` | 57 | 1 | замечания |
| `backend/internal/storage/repository.go` | 123 | 1 | замечания |
| `backend/internal/storage/service.go` | 140 | 4 | замечания |
| `backend/testframework/storage/confirm_a4006_test.go` | 77 | 0 | чисто |
| `backend/testframework/storage/download_a4007_test.go` | 46 | 0 | чисто |
| `backend/testframework/storage/filename_test.go` | 55 | 0 | чисто |
| `backend/testframework/storage/handler_extended_test.go` | 86 | 0 | чисто |
| `backend/testframework/storage/handler_routes_test.go` | 149 | 0 | чисто |
| `backend/testframework/storage/handler_wave117_test.go` | 182 | 0 | чисто |
| `backend/testframework/storage/orphan_cleanup_test.go` | 52 | 0 | чисто |
| `backend/testframework/storage/repository_test.go` | 95 | 0 | чисто |
| `backend/testframework/storage/service_ent_test.go` | 237 | 0 | чисто |
| `backend/testframework/storage/service_test.go` | 20 | 0 | чисто |
| `backend/testframework/storage/service_wave117_test.go` | 35 | 1 | замечания |
| `backend/internal/dashboard/dto.go` | 109 | 0 | чисто |
| `backend/internal/dashboard/export_hooks.go` | 25 | 0 | чисто |
| `backend/internal/dashboard/handler.go` | 243 | 2 | замечания |
| `backend/internal/dashboard/provisioner.go` | 18 | 0 | чисто |
| `backend/internal/dashboard/repository.go` | 315 | 1 | замечания |
| `backend/internal/dashboard/service.go` | 441 | 4 | замечания |
| `backend/testframework/dashboard/handler_test.go` | 286 | 0 | чисто |
| `backend/testframework/dashboard/repository_agg_a13003_test.go` | 71 | 0 | чисто |
| `backend/testframework/dashboard/repository_extra_test.go` | 79 | 0 | чисто |
| `backend/testframework/dashboard/repository_test.go` | 94 | 0 | чисто |
| `backend/testframework/dashboard/service_extended_test.go` | 225 | 0 | чисто |
| `backend/testframework/dashboard/service_test.go` | 82 | 0 | чисто |
| `backend/testframework/dashboard/service_wave107_test.go` | 149 | 0 | чисто |
| `backend/internal/backup/dto.go` | 15 | 0 | чисто |
| `backend/internal/backup/export_hooks.go` | 90 | 0 | чисто |
| `backend/internal/backup/handler.go` | 117 | 1 | замечания |
| `backend/internal/backup/maintenance.go` | 55 | 1 | замечания |
| `backend/internal/backup/repository.go` | 77 | 1 | замечания |
| `backend/internal/backup/service.go` | 376 | 4 | замечания |
| `backend/testframework/backup/coverage_wave80_test.go` | 70 | 0 | чисто |
| `backend/testframework/backup/dump_sanitize_a7007_test.go` | 35 | 0 | чисто |
| `backend/testframework/backup/handler_test.go` | 132 | 1 | замечания |
| `backend/testframework/backup/maintenance_a7003_test.go` | 63 | 0 | чисто |
| `backend/testframework/backup/prerestore_a7009_test.go` | 22 | 0 | чисто |
| `backend/testframework/backup/repository_test.go` | 72 | 0 | чисто |
| `backend/testframework/backup/restore_secrets_a7007_test.go` | 41 | 0 | чисто |
| `backend/testframework/backup/service_extended_test.go` | 160 | 0 | чисто |
| `backend/testframework/backup/service_test.go` | 50 | 0 | чисто |
| `backend/testframework/backup/service_wave100_test.go` | 130 | 0 | чисто |
| `backend/testframework/backup/service_wave98_test.go` | 53 | 0 | чисто |
| `backend/testframework/backup/service_wave99_test.go` | 106 | 0 | чисто |

Файлов в списке: 48. Прочитано: 48. Пропущено: 0.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 4 |
| P2 | 9 |
| P3 | 7 |
| **Σ** | **20** |

### Топ-3

1. **W7-A4-001 (P1)** — `PresignPutObject` без `ContentLength`/лимита размера → NF-09: клиент может залить произвольный объём напрямую в SeaweedFS.
2. **W7-A4-002 (P1)** — `ConfirmUpload` принимает любой `entityType`/`entityId` без проверки существования/принадлежности сущности → ложные привязки вложений.
3. **W7-A4-003 (P1)** — `MaintenanceGate` — только `atomic.Bool` in-process; при >1 реплике K3s остальные поды продолжают писать во время `pg_restore --clean` / `mongorestore --drop`.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W7-A4-001 | P1 | `storage/service.go:84-88` | `PresignPutObject` задаёт только `Bucket`/`Key`/`ContentType`; нет `ContentLength` и server-side лимита | TZ NF-09: Fiber BodyLimit 25 MB не действует на direct PUT; диск SeaweedFS можно исчерпать | Presign с `ContentLength` (клиент сообщает size + max 25 MB) или post-upload HEAD + orphan при превышении; тест: oversized PUT → reject/cleanup |
| W7-A4-002 | P1 | `storage/handler.go:171-176`; `storage/service.go:104-108`; `storage/repository.go:72-82` | Confirm пишет `entityType`/`entityId` как есть: нет allowlist типов, нет UUID/проверки, что issue/message принадлежит тому же `projectID` файла | Член проекта может «привязать» файл к чужой/несуществующей сущности; метаданные вложений врут; нет regression на entity-scope | Валидировать type∈{issue,message,…}; парсить UUID; проверять entity.project_id == file.project_id; denial-тест чужой issue |
| W7-A4-003 | P1 | `backup/maintenance.go`; wiring `wiring_fiber.go` | **FIXED TASK-100 (2026-07-28):** Redis `hub:backup:maintenance` + multi-instance IT; ADR-BACKUP-MAINTENANCE-GATE-2026. Было: process-local atomic.Bool | Restore на A не блокировал B | Redis shared gate (выбрано) |
| W7-A4-004 | P1 | `backup/service.go:105-118`; закреплён `service_wave99_test.go:58-72` | При `s3Client == nil` ZIP не загружается, но `MarkCompleted` вызывается | Админ видит `completed` без объекта в SeaweedFS → restore невозможен / ложное чувство безопасности | Fail-closed: без S3 → `MarkFailed` + error; тест CompletesWithoutS3 заменить на fail |
| W7-A4-005 | P2 | `storage/service.go:104-108` | Confirm не делает `HeadObject`/`GetObjectAttributes` | Клиент может confirm без PUT → confirmed-призрак; download URL 404 у S3 | Перед Confirm — HEAD по `S3Key`; иначе 400/409; тест без объекта → fail |
| W7-A4-006 | P2 | `storage/service.go:66-88`; `handler.go:105-107` | `contentType` без allowlist; уходит в Presign и метаданные | `text/html`/`image/svg+xml` через presigned GET → XSS при открытии URL в браузере | Allowlist MIME; `ResponseContentDisposition: attachment`; тест html → 400 |
| W7-A4-007 | P2 | `backup/service.go:415-418,341-345`; `toBackupResponse:422-434` | `markFailed` пишет `err.Error()` (обёртка `pg_dump`/`mongodump`) в `ErrorMessage`; List отдаёт поле админу | Restore-путь санитизирован (A7-007), dump-fail путь — нет; в сообщении могут быть host/DSN фрагменты | Санитизировать/урезать ErrorMessage; не включать `%w` exec; тест как dump_sanitize |
| W7-A4-008 | P2 | `dashboard/repository.go:269-272`; `service.go:372-401` | `ListProjectsWithIssues` = `Project.Query().WithIssues().All()` для top-5 | Память/latency растут со всеми issues всех проектов; A13-003 закрыл hot path задач, admin top — нет | SQL `GROUP BY project ORDER BY count LIMIT 5`; тест на агрегат без `.All` issues |
| W7-A4-009 | P2 | `dashboard/service.go:65-67,103-104,123-124` | Ошибки repo/Redis логируются, метод всё равно `return data, nil` | Клиент получает 200 с пустыми виджетами при падении БД — silent degradation | На critical-ошибках возвращать err → 500; частичный degrade только с флагом |
| W7-A4-010 | P2 | `dashboard/service.go:164-165` | `today` через `now.Location()` (локаль процесса), не UTC | TZ: UTC в БД; граница суток time_logs съезжает относительно UTC `created_at` | `time.Now().UTC()` + UTC midnight; тест около полуночи |
| W7-A4-011 | P2 | `backup/handler.go:98-100,125-127` | NotFound restore/delete → generic 500 `"restore/delete failed"` | Клиент не отличает 404 от сбоя; усложняет ops | `errors.Is`/IsNotFound → 404; тест missing id → 404 |
| W7-A4-012 | P2 | `backup/handler_test.go:24`; wiring `wiring_integrations.go:108` | Handler-тесты stubs `adminMw = Next()`; нет denial non-admin | Регресс снятия `RequireAdmin` с `/admin/backups` зелёный навсегда | Тест с реальным `rbac.RequireAdmin`: member → 403, admin/founder → pass |
| W7-A4-013 | P2 | `backup/service.go:341,351,270-274,289-292` | `pg_dump`/`mongodump`/`pg_restore`/`mongorestore` получают URI с паролем в argv | Утечка секретов через `/proc`, auditd, crash dumps | `.pgpass`/env `PGPASSWORD` + URI без пароля; аналог для mongo |
| W7-A4-014 | P3 | `storage/repository.go:123-127` | `SoftDelete` = `UpdateOneID` без `DeletedAtIsNil`/status | Повторный delete «успешен»; нет идемпотентного 404 | Where active; 0 rows → NotFound |
| W7-A4-015 | P3 | `storage/orphan_cleanup.go:12,32` | Один батч `Limit(500)`, без цикла | При >500 stale pending хвост живёт ещё сутки+ | `for { list; if empty break }` до исчерпания |
| W7-A4-016 | P3 | `dashboard/service.go:448-459,489-490`; `handler.go:202,258` | `ApplyTemplate`/`SaveLayout` без `IsProjectMember` | Не IDOR данных (layout per-user), но мусорные layout’ы чужих project_id | Membership check как в GetProjectStats |
| W7-A4-017 | P3 | `dashboard/handler.go:115` | `Admin` читает `h.service.repo.GetUser` напрямую | Нарушение layering Handler→Service | Метод `service.RequireSystemAdmin` / `GetUserRole` |
| W7-A4-018 | P3 | `backup/repository.go:60-64`; `service.go:208` | `SetRestoredNote` пишет в `ErrorMessage` | Путает ошибку dump и факт restore в List API | Отдельное поле `restored_at`/`note` в схеме |
| W7-A4-019 | P3 | `storage/service.go:91`; `backup/service.go:208` | `_ = DeleteByID` / `_ = SetRestoredNote` | Тихий fail cleanup/аудита | Логировать; при критичном — escalate |
| W7-A4-020 | P3 | `storage/service_wave117_test.go:38-41` | `_ = err` после Get orphan — assert фактически отсутствует | Тест зелёный при любом исходе ветки S3-error | `require.NoError` на Get (meta осталась) **или** assert deleted |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `storage/service.go:9-12,50-54` | SeaweedFS через **aws-sdk-go-v2** (`s3.NewFromConfig`, `UsePathStyle=true`), не minio-go |
| `storage/service.go:83-88,130-133` | Presigned PUT (5 min) / GET (1 h); upload без буфера backend |
| `storage/filename.go` + `filename_test.go` | Path traversal: `filepath.Base`, strip `..` `/` `\`, unicode; payloads `../../etc/passwd` → `passwd` |
| `storage/handler.go:119-122,162-168,212-218,253-259` | CR-04: RBAC по `FileProjectID` файла, не по client projectId на confirm/download/delete; nil RBAC fail-closed 500 |
| `storage/repository.go:112-120` | A4-007: download только `confirmed` + not deleted |
| `storage/repository.go:70-102` | A4-006: re-confirm → `ErrAlreadyConfirmed`, entity не перепривязывается |
| `storage/repository.go:123-127` | Soft delete (корзина), не hard delete user path |
| `storage/export_hooks.go` | Только hubtest exports |
| `confirm_a4006` / `download_a4007` / `handler_wave117` | Asserts есть; viewer/non-member → 403 |
| `dashboard/service.go:178-186` | A7-001: membership **до** Redis cache/DB stats |
| `dashboard/handler.go:114-120` | A7-002: admin dashboard — `auth.IsSystemAdmin` (founder\|\|admin) |
| `dashboard/repository.go:130-189` | A13-003: агрегаты Count, не `.All` всех issues на hot path |
| `dashboard/repository.go:62-117` | Personal widgets scoped по `userID` |
| `dashboard/provisioner.go` | Только ApplyTemplate `"dev"` для owner |
| `dashboard/export_hooks.go` | hubtest-only |
| dashboard tests | Membership denial, founder admin, agg, cache hit — asserts есть |
| `backup/handler.go:22-27` | Routes `/admin/backups` + `authMw` + `adminMw`; wiring: `rbac.RequireAdmin` |
| `backup/service.go:156-161` | A7-009 fail-closed pre-restore backup |
| `backup/service.go:178,251` | `io.LimitReader` 512 MiB (не голый `ReadAll`) |
| `backup/service.go:257-262` | extractZIP: только exact `postgres.dump`/`mongo.archive` — zip-slip на FS отсутствует (in-memory) |
| `backup/service.go:278-282` | A7-007: restore errors без CombinedOutput/URI в client error |
| `backup/service.go:83-87` | A7-004: failed dump → error, не HTTP success |
| `backup/maintenance.go` | A7-003: 503 + health bypass; `request_id` в JSON |
| backup tests A7-003/007/009 | Asserts есть; secrets не в HTTP body |

## Фокус-сводка

| Тема | Вердикт |
|------|---------|
| aws-sdk-go-v2 / SeaweedFS | ✅ path-style S3 client |
| Presigned URLs | ⚠️ работают; **нет size bind (NF-09)** |
| Path traversal (filename/S3 key) | ✅ sanitize + тесты |
| Admin-only backup | ✅ wiring `RequireAdmin`; ⚠️ нет denial-теста в handler_test |
| Restore safety | ✅ pre-backup + maintenance; ⚠️ gate не cluster-wide |
