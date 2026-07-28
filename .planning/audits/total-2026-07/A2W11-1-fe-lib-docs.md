# A2 W11 агент 1 — fe lib+docs part 1 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W11-ASSIGNMENTS.md` A1 · Stage A: `A11-1-fe-lib-docs.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/api/accounting.ts` | 145 | 0 | чисто |
| `frontend/src/api/admin-ai.ts` | 136 | 0 | FIXED-OK AIProvider → ai-types |
| `frontend/src/api/admin-mail.ts` | 40 | 0 | чисто |
| `frontend/src/api/admin-onec.ts` | 35 | 0 | чисто |
| `frontend/src/api/admin-org.ts` | 49 | 0 | чисто |
| `frontend/src/api/admin.ts` | 151 | 0 | чисто |
| `frontend/src/api/ai-my.ts` | 25 | 0 | FIXED-OK MyAIProvider |
| `frontend/src/api/ai.ts` | 119 | 0 | чисто |
| `frontend/src/api/apikeys.ts` | 31 | 0 | чисто |
| `frontend/src/api/auth.ts` | 150 | 0 | чисто |
| `frontend/src/api/automation.ts` | 53 | 0 | чисто |
| `frontend/src/api/backup.ts` | 28 | 0 | чисто |
| `frontend/src/api/bases.ts` | 114 | 0 | FIXED-OK `version: number` required |
| `frontend/src/api/calendar.ts` | 47 | 0 | FIXED-OK `api.getBlob` |
| `frontend/src/api/calls.ts` | 455 | 0 | FIXED-OK 404→null / allSettled |
| `frontend/src/api/client.ts` | 244 | 0 | FIXED-OK `_retried` + JSON try + logout |
| `frontend/src/api/crm.ts` | 118 | 0 | чисто |
| `frontend/src/api/dashboard.ts` | 142 | 0 | чисто |
| `frontend/src/api/files.ts` | 88 | 0 | чисто |

Файлов в списке: **19**. Прочитано: **19**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| — | — | — | — | Открытых дефектов в периметре A1 нет | — | — |

## FIXED-OK (Stage A → закрыто после B/C)

| Stage A ID | Evidence |
|------------|----------|
| W11-A1-001 | `calendar.ts` → `api.getBlob(\`/calendar/export.ics?…\`)` |
| W11-A1-002 | `client.ts` `_retried` / `getBlob` `retried` — второй 401 → `expireSession` |
| W11-A1-003 | `getActiveCall`: только `code === 404` → null; иначе rethrow |
| W11-A1-004 | `listScheduledCallsForChannels` → `Promise.allSettled` |
| W11-A1-005 | `bases.ts` update DTO: `version: number` обязателен |
| W11-A1-006 | Wire `{"error","request_id"}` vs FE `ApiError` разъяснён в `docs/wiki/API.md` (клиентский слой) |
| W11-A1-007 | `revokeSessionBestEffort` → `POST /auth/logout` до local clear |
| W11-A1-008 | calendar export через getBlob (ApiError path) |
| W11-A1-009 | `JSON.parse` в try/catch → `ApiError` code 502 |
| W11-A1-010 | `api/ai-types.ts` + alias в admin-ai / ai-my |

## Чисто

Проверено: все модули через `api` (calendar ICS мигрирован); нет сырого fetch с Bearer кроме logout best-effort; FormData/S3 PUT без JWT; нет `any`/`@ts-ignore`; crm/accounting version на мутациях.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 19/19/0 |
| Открытых находок | **0** |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| FIXED-OK | 10 |
