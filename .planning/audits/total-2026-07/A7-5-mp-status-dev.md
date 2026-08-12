# W7 агент A5 — marketplace + status + dev

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: DEV_MODE gates, marketplace authz, status endpoint exposure.  
> Сверка: grok A9-021 (`/api/docs`); ME-09 (`!release` + DEV_MODE); V4-06 status page; A6-003 optimistic lock.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/marketplace/dto.go` | 179 | 1 | замечания |
| `backend/internal/marketplace/export_hooks.go` | 47 | 0 | чисто |
| `backend/internal/marketplace/handler.go` | 245 | 1 | замечания |
| `backend/internal/marketplace/repository.go` | 197 | 2 | замечания |
| `backend/internal/marketplace/service.go` | 332 | 4 | замечания |
| `backend/testframework/marketplace/dto_wave99_test.go` | 27 | 0 | чисто |
| `backend/testframework/marketplace/fixtures_test.go` | 74 | 0 | чисто |
| `backend/testframework/marketplace/handler_test.go` | 205 | 0 | чисто |
| `backend/testframework/marketplace/handler_wave105_test.go` | 156 | 0 | чисто |
| `backend/testframework/marketplace/handler_wave97_test.go` | 79 | 0 | чисто |
| `backend/testframework/marketplace/optimistic_lock_a6003_test.go` | 46 | 0 | чисто |
| `backend/testframework/marketplace/service_test.go` | 180 | 0 | чисто |
| `backend/testframework/marketplace/service_wave105_test.go` | 85 | 0 | чисто |
| `backend/internal/status/dto.go` | 18 | 0 | чисто |
| `backend/internal/status/handler.go` | 35 | 1 | замечания |
| `backend/internal/status/service.go` | 65 | 1 | замечания |
| `backend/testframework/status/handler_test.go` | 45 | 0 | чисто |
| `backend/testframework/status/service_test.go` | 40 | 0 | чисто |
| `backend/internal/dev/apidocs.go` | 203 | 2 | замечания |
| `backend/internal/dev/export_hooks.go` | 159 | 0 | чисто |
| `backend/internal/dev/handler.go` | 124 | 1 | замечания |
| `backend/internal/dev/handler_release.go` | 39 | 0 | чисто |
| `backend/internal/dev/seed.go` | 768 | 1 | замечания |
| `backend/internal/dev/seed_data.go` | 271 | 1 | замечания |
| `backend/testframework/dev/apidocs_wave102_test.go` | 102 | 0 | чисто |
| `backend/testframework/dev/handler_wave102_test.go` | 139 | 0 | чисто |
| `backend/testframework/dev/seed_test.go` | 92 | 0 | чисто |

Файлов в списке: 27. Прочитано: 27. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W7-A5-001 | P2 | `apidocs.go:1-203`; wiring `wiring_fiber.go:77-85` | `BuildAPIDocsHTML`/`BuildAPIDocsJSON` в пакете `dev`, но **без** `//go:build !release` и без DEV_MODE; `/api/docs` + `/api/docs/doc.json` публичны в release | Полная карта method+path без auth (ранее A9-021 P3). Пакет `dev` вводит в заблуждение: ME-09 закрыл `/api/dev/*`, но docs остались | Gate: auth admin **или** только `!release`/`DEV_MODE`; либо вынести из `dev` и явно документировать public |
| W7-A5-002 | P2 | `status/handler.go:7-19`; `status/service.go:29-64` | `GET /api/status` без auth отдаёт per-component `postgres/redis/mongo/nats` → `operational/outage/unknown` (+ описания ролей) | V4-06 by design, но анонимный recon: видно, какой бэкенд упал, до `/ready`. Совпадает с health-топологией | Accepted risk **или** агрегировать в `operational/degraded` без id компонентов для anon; деталь — за auth/admin |
| W7-A5-003 | P2 | `marketplace/repository.go:94-114`; `service.go:58-83` | `ListListingsByAuthor` / `ListPendingListings` **без Limit** | Неограниченный SELECT при росте объявлений → memory/latency DoS на authz-protected эндпоинтах | Пагинация как у `ListApproved` (page/per_page cap 100); тест на Limit |
| W7-A5-004 | P2 | `marketplace/service.go:141-154` | Edit approved: `UpdateListing` затем отдельный `SetListingStatus(…Pending)` **без tx** | Между шагами листинг остаётся `approved` с новым контентом (или version+1 approved при fail reset) — обход re-moderation на гонке | Одна Ent-tx: update fields + status=pending + version; тест concurrent |
| W7-A5-005 | P2 | `marketplace/dto.go:117-124`; `toListingResponse:106-109` | `displayName` fallback на **email**; публичный `ListApproved` отдаёт `author_name` | Auth user видит email автора объявления, если нет display/name — PII на «витрине» | Fallback на «Пользователь» / маскированный id; email только в `ListMine`/admin |
| W7-A5-006 | P2 | `marketplace/service.go:321-331`; `repository.go:43-48` | `validateListingInput` не проверяет `price_cents` (≥0) и `currency` | Отрицательная цена / произвольная currency пишутся в БД | Валидация: price≥0; currency allowlist (ISO-4217 / RUB); тест |
| W7-A5-007 | P3 | `marketplace/service.go:245` | `DeleteListing`: `isAdmin, _ := IsSystemAdmin` — ошибка БД глотается | При DB blip админ без ownership получает `ErrForbidden` вместо 500 (fail-closed OK, но ops-маска) | Пропагировать err; 500 на repo fail |
| W7-A5-008 | P3 | `marketplace/handler.go:232-233` | `ErrForbidden` → `"admin access required"` и для stranger-delete (не-admin path) | Клиент/лог путает IDOR с отсутствием admin-роли | Разные сообщения / коды: `not listing owner` vs admin |
| W7-A5-009 | P3 | `marketplace/service.go:325-326` | Лимит title/description через `len` (байты), не руны | Non-ASCII обрежется/отклонится иначе, чем UI (класс auth rune-bug) | `utf8.RuneCountInString` |
| W7-A5-010 | P3 | `marketplace/repository.go:75-76` | `category` query → `CategoryEQ` без `isValidCategory` | Невалидная category → Ent error → 500 вместо пустого списка/400 | Валидировать в service/handler |
| W7-A5-011 | P3 | `marketplace/repository.go:162-168` | `SoftDeleteListing` без `VersionEQ` | Двойной concurrent delete / TOCTOU с update; слабее optimistic lock остальных мутаций | Where Version + body version **или** accepted soft-delete |
| W7-A5-012 | P3 | `dev/handler.go:93-116` | `/api/dev/login` → `CreateSession` **без** проверки `user.Status` (ban/pending) | В non-release+DEV_MODE забаненный получает JWT (обход magic-link status guards) | Проверять `StatusActive` перед сессией; тест |
| W7-A5-013 | P3 | `dev/seed.go:92-169` | Ошибки seed workflow/issues/… только slog; HTTP всё равно **200** с частичным `seeded` | «Успешный» seed при битой БД → flaky local/E2E | Fail-fast на critical steps или `partial: true` + non-2xx |
| W7-A5-014 | P3 | `dev/seed_data.go:168-173`; `:104-105` | Демо TC/issue про «Telegram авторизация / вход» | Расходится с ADR-AUTH-01 (Telegram ≠ login); путает онбординг | Переписать демо под notification bind |
| W7-A5-015 | P3 | `dev/apidocs.go:199-202` | `json.MarshalIndent` error `_` → пустая/битая JSON-строка | Редкий fail → 200 с пустым телом без сигнала | Проверять err, 500 |
| W7-A5-016 | P3 | `handler_test.go` vs `service_test.go:163-165` | HTTP DELETE stranger → 403 **не** покрыт (есть только service) | Регресс маппинга `ErrForbidden` на handler не ловится | `TestHandler_DeleteListing_Forbidden` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **DEV_MODE dual gate (ME-09)** | `handler.go` `//go:build !release` + runtime `devMode`→404; `handler_release.go` no-op RegisterRoutes; bootstrap WARN; тесты `TestHandler_DevModeGuard` |
| **Marketplace authz** | Все routes под `authMw`; pending/approve/reject → `requireAdmin` (founder\|admin); update/sold/withdraw → owner; delete → owner\|admin; pending 403 member в handler_test |
| **Optimistic lock A6-003** | `UpdateListing`/`SetListingStatus` всегда `VersionEQ`; test `version:0` + stale → conflict |
| **Soft delete / trash** | `SoftDeleteListing` + `DeletedAtIsNil` filters; hard delete нет |
| `export_hooks.go` (mp/dev) | hubtest re-exports only |
| `status/dto.go` | DTO only |
| `handler_release.go` | stub API идентичен, тело пустое |
| mp tests (wave97/105/a6003/dto/fixtures) | assert на 403/409/400/remoderation/owner; mutation-oriented a6003 |
| status tests | 200 operational / 503 degraded; nil pinger → unknown |
| `apidocs_wave102_test.go` | methodOrder/color/guessModule/collectRoutes |
| `seed_test.go` | demo fixtures non-empty / unique keys |
| `handler_wave102_test.go` | login validation, seed idempotent, status |

**Фокус-вердикт:** DEV_MODE/`!release` для `/api/dev/*` **закрыты корректно**. Главный остаточный риск зоны — **публичные `/api/docs` (вне ME-09)** и **status recon**; marketplace authz в целом solid, слабые места — unbounded lists, non-atomic re-moderation, email в `author_name`.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **27 / 27 / 0** |
| Находок | **16** |
| P0 | 0 |
| P1 | 0 |
| P2 | 6 |
| P3 | 10 |

### Топ-3

1. **W7-A5-001 (P2)** — `/api/docs` (+ `apidocs.go` без `!release`/DEV_MODE) публично отдаёт карту всех эндпоинтов.
2. **W7-A5-002 (P2)** — публичный `/api/status` раскрывает outage по postgres/redis/mongo/nats (recon поверх V4-06).
3. **W7-A5-004 (P2)** — edit approved listing без транзакции: окно «approved + новый контент» до re-moderation.
