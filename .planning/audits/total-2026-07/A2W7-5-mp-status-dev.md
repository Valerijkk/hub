# A2W7-5 — marketplace+status+dev (Stage A2)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Baseline: `A7-5-mp-status-dev.md` · Assignments: `W7-ASSIGNMENTS.md` §A5

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/marketplace/dto.go` | 179 | 1 | замечания |
| `backend/internal/marketplace/export_hooks.go` | 47 | 0 | чисто |
| `backend/internal/marketplace/handler.go` | 245 | 1 | замечания |
| `backend/internal/marketplace/repository.go` | 203 | 1 | замечания |
| `backend/internal/marketplace/service.go` | 343 | 1 | замечания |
| `backend/testframework/marketplace/dto_wave99_test.go` | 27 | 0 | чисто |
| `backend/testframework/marketplace/fixtures_test.go` | 74 | 0 | чисто |
| `backend/testframework/marketplace/handler_test.go` | 205 | 0 | чисто |
| `backend/testframework/marketplace/handler_wave105_test.go` | 156 | 0 | чисто |
| `backend/testframework/marketplace/handler_wave97_test.go` | 79 | 0 | чисто |
| `backend/testframework/marketplace/optimistic_lock_a6003_test.go` | 46 | 0 | чисто |
| `backend/testframework/marketplace/service_test.go` | 209 | 0 | чисто |
| `backend/testframework/marketplace/service_wave105_test.go` | 85 | 0 | чисто |
| `backend/internal/status/dto.go` | 18 | 0 | чисто |
| `backend/internal/status/handler.go` | 35 | 1 | замечания |
| `backend/internal/status/service.go` | 65 | 1 | замечания |
| `backend/testframework/status/handler_test.go` | 45 | 0 | чисто |
| `backend/testframework/status/service_test.go` | 40 | 0 | чисто |
| `backend/internal/dev/apidocs.go` | 207 | 1 | замечания |
| `backend/internal/dev/export_hooks.go` | 159 | 0 | чисто |
| `backend/internal/dev/handler.go` | 129 | 1 | замечания |
| `backend/internal/dev/handler_release.go` | 39 | 0 | чисто |
| `backend/internal/dev/seed.go` | 788 | 1 | замечания |
| `backend/internal/dev/seed_data.go` | 271 | 1 | замечания |
| `backend/testframework/dev/apidocs_wave102_test.go` | 103 | 0 | чисто |
| `backend/testframework/dev/handler_wave102_test.go` | 139 | 0 | чисто |
| `backend/testframework/dev/seed_test.go` | 92 | 0 | чисто |

Файлов в списке: **27**. Прочитано: **27**. Пропущено: **0**.

## Находки

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект / статус |
|----|-----|-----|-------------|---------|------------------------|
| W7-A5-001 | RESIDUAL | P2 | `dev/apidocs.go`; `wiring_fiber.go:77-86` | `/api/docs` публичен в release без auth/`!release` | W12 P2-BE-U04 product decision |
| W7-A5-002 | RESIDUAL | P2 | `status/handler.go` + service | Публичный `/api/status` раскрывает outage deps | Recon поверх V4-06; U04 |
| W7-A5-003 | RESIDUAL | P2 | `marketplace` ListByAuthor / admin lists | Unbounded author/admin lists; catalog имеет page но без `has_more` | Рост marketplace → large JSON |
| W7-A5-004 | RESIDUAL | P2 | `marketplace/service.go` edit approved | Update approved → pending без одной tx | Окно «approved + новый контент» |
| W7-A5-005 | RESIDUAL | P2 | `marketplace/dto.go:124` | `author_name` fallback = Email | PII email в публичном каталоге |
| W7-A5-006 | RESIDUAL | P2 | `marketplace/service.go:331-342` | `validateListingInput` без `price_cents`/`currency` | Отрицательная цена / произвольная currency |
| W7-A5-007 | FIXED-OK | P3 | DeleteListing IsSystemAdmin err | Swallow → Forbidden | Закрыто TASK-057 |
| W7-A5-008 | FIXED-OK | P3 | ErrForbidden message | «admin access required» для stranger | Закрыто TASK-063 path |
| W7-A5-009 | FIXED-OK | P3 | title/desc `len` bytes | Mid-rune | Закрыто TASK-060: RuneCount |
| W7-A5-010 | FIXED-OK | P3 | category filter | Invalid → 500 | Закрыто TASK-063 hygiene |
| W7-A5-011 | FIXED-OK | P3 | SoftDeleteListing VersionEQ | TOCTOU | Закрыто TASK-061 |
| W7-A5-012 | FIXED-OK | P3 | `dev/handler.go:110-111` | Dev login без StatusActive | Закрыто TASK-067 |
| W7-A5-013 | FIXED-OK | P3 | `dev/seed.go` | Partial seed → 200 | Закрыто TASK-067 fail-fast/partial |
| W7-A5-014 | FIXED-OK | P3 | `dev/seed_data.go:168-172` | Demo «Telegram login» | Закрыто TASK-067: notification bind copy |
| W7-A5-015 | FIXED-OK | P3 | apidocs MarshalIndent `_` | Empty JSON 200 | Закрыто TASK-057 |
| W7-A5-016 | FIXED-OK | P3 | DELETE stranger HTTP | Только service-level | Закрыто TASK-059 denial |

## Чисто (инварианты держат)

| Зона | Статус |
|------|--------|
| DEV_MODE + `!release` для `/api/dev/*` | Держит (ME-09) |
| Marketplace authz owner/admin + OL | Держит |
| Soft-delete listings | Держит |
| Status pinger degraded→503 | Держит (public recon — residual) |

## Counts

| Метрика | n |
|---------|--:|
| Файлов | 27 |
| Open (RESIDUAL) | **6** |
| FIXED-OK | 10 |
| NEW | 0 |
| REGRESS | 0 |
| Open by sev | P0:0 · P1:0 · P2:6 · P3:0 |
