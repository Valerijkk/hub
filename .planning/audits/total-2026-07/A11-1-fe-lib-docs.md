# W11 агент A1 — frontend api part 1 (lib+docs wave)

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус: API client auth headers · token storage / refresh · zustand persist secrets (смежно) · docs drift vs code  
> ID: `W11-A1-001`…

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/api/accounting.ts | 145 | 0 | чисто |
| frontend/src/api/admin-ai.ts | 164 | 1 | замечания |
| frontend/src/api/admin-mail.ts | 40 | 0 | чисто |
| frontend/src/api/admin-onec.ts | 35 | 0 | чисто |
| frontend/src/api/admin-org.ts | 49 | 0 | чисто |
| frontend/src/api/admin.ts | 151 | 0 | чисто |
| frontend/src/api/ai-my.ts | 37 | 1 | замечания |
| frontend/src/api/ai.ts | 119 | 0 | чисто |
| frontend/src/api/apikeys.ts | 37 | 0 | чисто |
| frontend/src/api/auth.ts | 150 | 0 | чисто |
| frontend/src/api/automation.ts | 53 | 0 | чисто |
| frontend/src/api/backup.ts | 33 | 0 | чисто |
| frontend/src/api/bases.ts | 132 | 1 | замечания |
| frontend/src/api/calendar.ts | 56 | 2 | замечания |
| frontend/src/api/calls.ts | 469 | 2 | замечания |
| frontend/src/api/client.ts | 209 | 3 | замечания |
| frontend/src/api/crm.ts | 133 | 0 | чисто |
| frontend/src/api/dashboard.ts | 160 | 0 | чисто |
| frontend/src/api/files.ts | 88 | 0 | чисто |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W11-A1-001 | P1 | frontend/src/api/calendar.ts:40 | `downloadCalendarExport` делает сырой `fetch` с ручным `Authorization`, минуя `api` / `getBlob` | Нет single-flight refresh (FE-H1), нет logout при протухшей сессии, нет разбора `ApiError`. TMS уже на `api.getBlob` (`client.ts:162`), calendar — регрессия того же класса | Перевести на `api.getBlob(\`/calendar/export.ics?…\`)` + тот же download UX; убрать дубль чтения токена |
| W11-A1-002 | P1 | frontend/src/api/client.ts:79 | После успешного refresh `request` рекурсивно вызывает себя без флага «уже ретраили» (то же в `getBlob`/`doFetch` ~169) | Если эндпоинт снова отдаёт 401 (RBAC/запрет, а не expiry) — бесконечная рекурсия → hang / stack overflow; refresh крутит ротацию токенов | Параметр `_retried?: boolean` (или один retry): при втором 401 — logout + throw, без нового refresh |
| W11-A1-003 | P2 | frontend/src/api/calls.ts:144 | `getActiveCall`: любой non-404 (и даже 404-ветка) в `catch` → `return null` | Сеть/500/403 неотличимы от «нет звонка»; UI показывает пусто вместо ошибки | В `catch`: только `code === 404` → null; иначе rethrow |
| W11-A1-004 | P2 | frontend/src/api/calls.ts:463 | `listScheduledCallsForChannels`: `.catch(() => [])` на каждый канал | Падение одного канала тихо выкидывает его звонки из календаря | Пробрасывать/агрегировать ошибки; или partial success + toast |
| W11-A1-005 | P2 | frontend/src/api/bases.ts:74 | `updateBase` / `updateRecord` (стр. 116): `version?: number` | BE (`bases/service.go`) проверяет lock только при `Version > 0`; опциональный тип провоцирует вызовы без version → silent skip optimistic lock | Сделать `version: number` обязательным в update DTO (как `crm.ts` updateContact/Deal) |
| W11-A1-006 | P2 | frontend/src/api/client.ts:10 | `ApiError` = `{ code, message, request_id? }` — совпадает с `platform/errors.AppError` | Docs drift: `docs/tz_constraints.md`, `docs/system_specification.md` NFR-13, `docs/wiki/API.md:40`, `docs/code_style_guide.md` обещают `{"error","request_id"}` без `code`/`message`. FE/BE канон другой → путаница контракта | Синхронизировать docs под фактический JSON (`code`+`message`+`request_id`) **или** менять BE+FE (дорого). Передать A4/A5 |
| W11-A1-007 | P2 | frontend/src/api/client.ts:84 | При failed refresh — только `useAuthStore.logout()` (локальная очистка), без `POST /auth/logout` | При сетевом сбое refresh валидный refresh в Redis живёт до 30д, локально юзер уже «вышел» — orphaned session | Best-effort `api.post("/auth/logout")` до clear **или** отдельный unauth revoke-by-refresh body; fire-and-forget |
| W11-A1-008 | P3 | frontend/src/api/calendar.ts:47 | `throw new Error("export failed")` без status/`ApiError` | `toastApiError`/UX теряют HTTP-код и `request_id` | Парсить JSON ошибки как в `getBlob`; бросать `ApiError` |
| W11-A1-009 | P3 | frontend/src/api/client.ts:112 | `JSON.parse(text)` без try/catch на 2xx | Невалидный JSON → сырой SyntaxError вместо `ApiError` | try/catch → `{ code: 502, message: "invalid json" }` |
| W11-A1-010 | P3 | frontend/src/api/admin-ai.ts:3 | Дублирующий `AIProvider` (+ create DTO) vs `ai-my.ts:3` с разной формой | Коллизия имён при совместном импорте; drift полей (`priority`/`max_rpm`/`scope` только в admin) | Общий тип в `api/ai-types.ts` или `Pick`/extends |

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 2 |
| P2 | 5 |
| P3 | 3 |
| **Σ** | **10** |

### Топ-3
1. **W11-A1-001** — calendar ICS export обходит `ApiClient`/`getBlob` → нет refresh при 401, сессия «ломается» иначе, чем у остальных API.
2. **W11-A1-002** — рекурсивный retry 401 без лимита → потенциальный hang/DoS вкладки при стабильном 401 после refresh.
3. **W11-A1-006** — docs (NFR-13 / wiki / TZ) vs реальный `ApiError`/`AppError` (`code`+`message`) — системный drift контракта ошибок.

## Auth / tokens / persist (наблюдения по зоне A1)

| Тема | Факт в коде A1 | Оценка |
|------|----------------|--------|
| Auth header | `client.ts:52–55` — `Authorization: Bearer ${accessToken}` из `useAuthStore` | OK; FormData без ручного Content-Type |
| Refresh | `doRefresh` POST `/auth/refresh` **без** Bearer, body `refresh_token`; single-flight mutex FE-H1 | OK паттерн; дыра — W11-A1-002 |
| getBlob | Добавлен после A11-007; TMS использует | calendar **не** мигрирован (W11-A1-001) |
| files PUT | `files.ts:74` — presigned без JWT (правильно) | OK |
| Секреты в DTO | `admin-ai`/`ai-my` `api_key?` write-only; `admin-onec` `bearer_token` + `bearer_token_set`; `apikeys` полный `key` только в create response | OK для API-слоя |
| Zustand persist | **вне списка A1**: `auth-store.ts` persist `accessToken`+`refreshToken` в `sessionStorage` (`hub-auth`) | **Передать A2** — XSS → кража пары токенов; `partialize` / memory-only refresh — тема A2 |

Контекст (read для сверки, не в scope находок A1): `frontend/src/stores/auth-store.ts` — persist обеих JWT в sessionStorage.

## Чисто
Проверено: все модули ходят через `api` (кроме calendar export и S3 PUT); нет `any`/`@ts-ignore`; admin-секреты не возвращаются полным ключом в list DTO; auth.ts нормализует birthday RFC3339; crm требует `version` на update; accounting/dashboard/admin*/automation/backup/apikeys — тонкие typed wrappers без обхода клиента.

| Файл | Что проверено |
|------|----------------|
| accounting.ts | только `api.*`; version на ledger entities в типах; нет сырого fetch |
| admin-mail.ts | settings/broadcast через api; без секретов в response |
| admin-onec.ts | `bearer_token_set` + write-only `bearer_token` |
| admin-org.ts | pending/invite/registration-mode; invite_token только в ответе GET |
| admin.ts | audit/trash/GDPR/ban — query params encode; через api |
| ai.ts | chat/jql/composer/preferences — через api |
| apikeys.ts | list без полного key; create возвращает `key` один раз (ожидаемо) |
| auth.ts | magic-link/verify/me/sessions/logout через api; типы TokenPair |
| automation.ts | CRUD rules через api |
| backup.ts | admin backups + confirm RESTORE |
| crm.ts | version обязателен на update contact/deal |
| dashboard.ts | personal/project/admin/layout/templates |
| files.ts | camelCase совпадает с BE storage handler; S3 PUT без JWT |

**Хорошие паттерны:** single-flight refresh (`client.ts:33–34,123–128`); `getBlob` для binary; `has_api_key` / `bearer_token_set` без эха секрета; crm `version: number` required.
