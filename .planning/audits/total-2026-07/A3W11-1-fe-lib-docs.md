# A3 W11 агент 1 — fe lib+docs part 1 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W11-1-fe-lib-docs.md](A2W11-1-fe-lib-docs.md)  
> Stage A: [A11-1-fe-lib-docs.md](A11-1-fe-lib-docs.md)  
> Периметр: 19 файлов (W11-ASSIGNMENTS A1) · Прочитано: 19 · Пропущено: 0  
> Фокус: api client auth headers · token storage XSS · docs drift vs code

## Сверка с A2 (A2W11-1-*)

| A2 / Stage A ID | Tag (A2) | Статус в HEAD | Доказательство |
|-----------------|----------|---------------|----------------|
| W11-A1-001 | FIXED-OK | без регресса | `calendar.ts:38-40` → `api.getBlob(\`/calendar/export.ics?…\`)` |
| W11-A1-002 | FIXED-OK | без регресса | `client.ts:101,116-125` `_retried`; `getBlob` `retried` → `expireSession` |
| W11-A1-003 | FIXED-OK | без регресса | `calls.ts:159-164` — только `code === 404` → null; иначе rethrow |
| W11-A1-004 | FIXED-OK | без регресса | `calls.ts:481-499` `Promise.allSettled`; throw если все rejected |
| W11-A1-005 | FIXED-OK | без регресса | `bases.ts:74,116` — `version: number` required |
| W11-A1-006 | FIXED-OK (**ложное**) | **FIXED** (TASK-129) | `mapWireToApiError` в `client.ts` (`request`/`getBlob`); wiki `API.md:47` = код; A3W11-1-001 closed fe1 |
| W11-A1-007 | FIXED-OK | без регресса | `client.ts:65-85` `revokeSessionBestEffort` → POST `/auth/logout` до `logout()` |
| W11-A1-008 | FIXED-OK | без регресса | calendar export через `getBlob` (общий error path) |
| W11-A1-009 | FIXED-OK | без регресса | `client.ts:152-160` JSON.parse try/catch → `{ code: 502, message: "invalid json" }` |
| W11-A1-010 | FIXED-OK | без регресса | `api/ai-types.ts` + alias в `admin-ai.ts` / `ai-my.ts` |

**REGRESS (closed):** 1 → FIXED TASK-129 · **RESIDUAL (A2 actives):** 0 · **NEW (этот слот):** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/api/accounting.ts` | 145 | 0 | чисто |
| `frontend/src/api/admin-ai.ts` | 154 | 0 | FIXED-OK AIProvider → ai-types |
| `frontend/src/api/admin-mail.ts` | 40 | 0 | чисто |
| `frontend/src/api/admin-onec.ts` | 35 | 0 | чисто |
| `frontend/src/api/admin-org.ts` | 49 | 0 | чисто |
| `frontend/src/api/admin.ts` | 151 | 0 | чисто |
| `frontend/src/api/ai-my.ts` | 31 | 0 | FIXED-OK MyAIProvider |
| `frontend/src/api/ai.ts` | 119 | 0 | чисто |
| `frontend/src/api/apikeys.ts` | 37 | 0 | чисто |
| `frontend/src/api/auth.ts` | 161 | 0 | чисто |
| `frontend/src/api/automation.ts` | 53 | 0 | чисто |
| `frontend/src/api/backup.ts` | 33 | 0 | чисто |
| `frontend/src/api/bases.ts` | 132 | 0 | FIXED-OK `version` required |
| `frontend/src/api/calendar.ts` | 47 | 0 | FIXED-OK `api.getBlob` |
| `frontend/src/api/calls.ts` | 500 | 0 | FIXED-OK 404→null / allSettled |
| `frontend/src/api/client.ts` | 261 | 0 | FIXED wire→ApiError map (TASK-129) |
| `frontend/src/api/crm.ts` | 133 | 0 | чисто |
| `frontend/src/api/dashboard.ts` | 160 | 0 | чисто |
| `frontend/src/api/files.ts` | 91 | 0 | чисто |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W11-1-001 | **FIXED** | P1 | `client.ts` `mapWireToApiError` (`request` + `getBlob`); docs claim `docs/wiki/API.md:47` | ~~После TZ-wire клиент не мапил `error`→`message`~~ → **FIXED** TASK-129: wire→`ApiError` map + Vitest (429 toast/request_id + 404 + getBlob); mutation «убрать map» краснеет | ~~toast unknown / no request_id / no 429 countdown~~ | fe1 2026-07-29 · статус задачи «На проверке» |

## FIXED-OK (сверка A2 + Stage A фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| Auth header | `client.ts:47-56` — Bearer из `useAuthStore`; FormData без ручного Content-Type |
| Refresh single-flight | `tryRefresh`/`doRefresh` FE-H1; POST `/auth/refresh` без Bearer, body `refresh_token` |
| One-retry 401 | `_retried` / getBlob `retried` → второй 401 = `expireSession` |
| Logout revoke | `revokeSessionBestEffort` сырой fetch (без цикла refresh) |
| Calendar ICS | только `api.getBlob` — нет сырого Bearer fetch |
| Files PUT | presigned S3 без JWT (`files.ts:77-81`) |
| Optimistic lock DTO | bases/crm update: `version: number` обязателен |
| AI types | `ai-types.ts` Admin/My; admin-ai/ai-my — deprecated aliases |
| Calls soft-null | `getActiveCall` только 404→null; `listScheduledCallsForChannels` allSettled |
| Token storage XSS (смежно A2) | `auth-store.ts` — JWT **только в памяти**, без persist sessionStorage (A2W11-2-001); BroadcastChannel mirror в localStorage несёт только `{type,userId}`, не токены. XSS всё ещё читает in-memory, пока вкладка жива — ожидаемо до httpOnly cookie ADR |

## Auth / tokens / docs (наблюдения)

| Тема | Факт | Оценка |
|------|------|--------|
| Bearer | все domain wrappers через `api.*` | OK |
| Сырой fetch + Bearer | только logout best-effort | OK |
| Persist JWT | убран (память) | FIXED-OK vs Stage A note |
| Docs wire vs FE | wiki описывает map; код не мапит | **A3W11-1-001** |
| `ApiError` type | `{code,message,request_id?}` — клиентский DTO | OK как тип; runtime object сейчас ближе к wire+`code` |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 0 |
| **Total** | **0** |

| Метрика | Значение |
|---------|----------|
| Файлов | **19 / 19 / 0** |
| Активных | **0** (NEW 0 · REGRESS closed 1 · RESIDUAL 0) |
| FIXED-OK (Stage A via A2 + TASK-129) | **10/10** |

### Топ-3

1. ~~**A3W11-1-001 (P1, REGRESS)**~~ → **FIXED** TASK-129 (`mapWireToApiError`).
2. _(нет P2+)_ — calendar/getBlob/retry/logout/version/calls — без регресса.
3. _(нет P0)_ — token persist XSS закрыт memory-only (слот A2).

---

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=0 · P3=0 · Total=0.  
Топ: _(пусто — A3W11-1-001 FIXED)_.
