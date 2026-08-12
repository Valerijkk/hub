# A2 W4 агент 4 — platform config/db/crypto/core

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W4-ASSIGNMENTS.md` A4 · Stage A: `A4-4-platform-core.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/config/config.go` | 378 | 0 | чисто |
| `backend/internal/platform/config/export_hooks.go` | 14 | 0 | чисто |
| `backend/internal/platform/cron/cron.go` | 102 | 0 | чисто |
| `backend/internal/platform/crypto/aesgcm.go` | 173 | 1 | замечания |
| `backend/internal/platform/database/mongodb.go` | 34 | 0 | чисто |
| `backend/internal/platform/database/postgres.go` | 60 | 0 | чисто |
| `backend/internal/platform/errors/errors.go` | 53 | 0 | чисто |
| `backend/internal/platform/errors/messages.go` | 16 | 0 | чисто |
| `backend/internal/platform/health/probes.go` | 71 | 0 | чисто |
| `backend/internal/platform/lifecycle/listen.go` | 14 | 0 | чисто |
| `backend/internal/platform/telegram/markdown.go` | 25 | 0 | чисто |
| `backend/testframework/platform/config/config_test.go` | 324 | 0 | чисто |
| `backend/testframework/platform/cron/cron_test.go` | 54 | 0 | чисто |
| `backend/testframework/platform/crypto/aesgcm_test.go` | 91 | 0 | чисто |
| `backend/testframework/platform/database/mongo_disconnect_a8007_test.go` | 37 | 0 | чисто |
| `backend/testframework/platform/errors/errors_test.go` | 130 | 0 | чисто |
| `backend/testframework/platform/errors/errors_wave101_test.go` | 18 | 0 | чисто |
| `backend/testframework/platform/health/probes_test.go` | 92 | 0 | чисто |
| `backend/testframework/platform/lifecycle/listen_a8008_test.go` | 25 | 0 | чисто |
| `backend/testframework/platform/telegram/markdown_test.go` | 28 | 0 | чисто |

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.

## Находки (активные)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A2W4-4-001 | FIXED | P3 | `aesgcm.go` EncryptWithAAD/DecryptWithAAD | AAD API + тест mismatch; legacy Encrypt/MaybeDecrypt без AAD для migration | Evidence: TASK-112 | ADR rollout: AAD=`entityID`; сузить MaybeDecrypt после миграции |

## FIXED-OK

| Stage A ID | Sev | Статус | Доказательство |
|------------|-----|--------|----------------|
| W4-A4-001 | P1 | FIXED-OK | `SetMaxOpenConns(25)` + idle/lifetime |
| W4-A4-002 | P1 | FIXED-OK | prod reject `hub:hub@` DB/Mongo/Redis |
| W4-A4-003 | P2 | FIXED-OK | `Schema.Create` только `DEV_MODE` |
| W4-A4-004 | P2 | FIXED-OK | prod sslmode require\|verify-* |
| W4-A4-005 | P2 | FIXED-OK | AppError / ErrorHandler TZ shape |
| W4-A4-006 | P2 | FIXED-OK | `\` первым в markdownV2Special |
| W4-A4-007 | P2 | FIXED-OK | Ready без ping → 503 |
| W4-A4-008 | P2 | FIXED-OK | tick `WithTimeout` |
| W4-A4-009 | P2 | FIXED-OK | `IsDevDerivedKey` reject в validate |
| W4-A4-010 | P3 | FIXED-OK | `OpenRawDB` удалён |
| W4-A4-011 | P3 | FIXED-OK | assert-free тест убран/заменён |
| W4-A4-012 | P3 | FIXED-OK | Ready parent = `c.Context()` |

## Чисто

| Зона | Проверено |
|------|-----------|
| AES-GCM core | random nonce 12B; tamper reject; no panic on nil cipher |
| DB pool | limits applied; no OpenRawDB |
| SO_REUSE | listen helper cancel-only; reuseport не включён |
| Config A14 | JWT/S3/LiveKit/NATS/metrics/encryption |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **20 / 20 / 0** |
| Активных | **1** (RESIDUAL P3) |
| FIXED-OK | **12** |
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 1 |
