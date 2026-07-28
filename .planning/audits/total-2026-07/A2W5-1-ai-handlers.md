# A2-W5 агент A1 — ai handlers (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc`  
> Зеркало: `W5-ASSIGNMENTS.md` §A1 · Stage A: `A5-1-ai-handlers.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | (summary) `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/admin_handler.go` | 154 | 3 | замечания |
| `backend/internal/ai/dto.go` | 171 | 0 | чисто |
| `backend/internal/ai/export_hooks.go` | 228 | 0 | чисто |
| `backend/internal/ai/handler.go` | 256 | 1 | замечания |
| `backend/internal/ai/user_handler.go` | 82 | 2 | замечания |
| `backend/testframework/ai/admin_handler_test.go` | 113 | 0 | чисто |
| `backend/testframework/ai/admin_handler_wave110_test.go` | 122 | 0 | чисто |
| `backend/testframework/ai/coverage_wave81_test.go` | 223 | 0 | чисто |
| `backend/testframework/ai/handler_extended_test.go` | 182 | 0 | чисто |
| `backend/testframework/ai/handler_wave115_test.go` | 178 | 0 | чисто |
| `backend/testframework/ai/user_handler_test.go` | 79 | 0 | чисто |

Файлов в списке: 11. Прочитано: 11. Пропущено: 0.

## Находки

| ID | Sev | Tag | Файл:строка | Проблема | A1 / задача | Мутация |
|----|-----|-----|-------------|----------|-------------|---------|
| A2W5-A1-001 | P2 | REGRESS | `handler.go:126-130` | **AssistJQL** по-прежнему: только `ErrNoProvider`→503; остальное → **502 + `err.Error()`** (quota/consent/policy → Bad Gateway) | W5-A1-001 · P2-BE-07 «Исправлена» | Добавить ветки как SuggestComposer; снять → 502 на quota |
| A2W5-A1-002 | P2 | REGRESS | `admin_handler.go:23`; `user_handler.go:20`; `wiring_audit_ai.go:65-68` | **AIRateLimit** всё ещё только на user `Handler`; admin `TestProvider` и `/ai/my` без AI RL | W5-A1-003 · P2-BE-17 | Прокинуть `rateLimitMw`; без MW TestProvider без 429 при burst |
| A2W5-A1-003 | P2 | RESIDUAL | `admin_handler.go:69`; `:89`; `user_handler.go:55,79` | Create/Update отдают **`err.Error()`** клиенту (encrypt/base_url/repo) | W5-A1-004 · P2-BE-07 | Стабильные коды; inject cipher err → body без internals |
| A2W5-A1-004 | P2 | RESIDUAL | `service.go:257-259` → `admin_handler.go:100-104` | `TestProvider` при fail кладёт **сырой `err.Error()`** в JSON `error` | W5-A1-005 | Санитизировать; assert body ≠ upstream URL/auth |
| A2W5-A1-005 | P3 | RESIDUAL | `user_handler_test.go` / `admin_handler_test.go` | Create без assert «нет `api_key` в JSON» / только `has_api_key` | W5-A1-011 · P3-BE-03 | Create с `sk-test` → `NotContains` raw key |

## FIXED-OK (выборочно vs A1)

| A1 ID | Статус | Evidence |
|-------|--------|----------|
| W5-A1-002 | FIXED-OK | `StartChatStream` preflight `resolveProviders` до 202 (`service.go:489-491`) |
| W5-A1-006 | FIXED-OK | `TestUserHandler_UpdateProvider_ForeignOwner_TASK037` |
| W5-A1-008 | FIXED-OK | SuggestComposer: `errors.Is` sentinels (`handler.go:167+`) |
| W5-A1-009 | FIXED-OK | `uuid.Nil` → 400 в admin/user Update |
| W5-A1-010 | FIXED-OK | `requireSystemProvider` на admin Update/Test |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `dto.go` | Response: `has_api_key`, без plaintext key |
| `export_hooks.go` | hubtest re-export only |
| auth MW | admin `requireAdmin` fail-closed; user routes под `authMw` |
| Chat/composer | quota→429, consent→403, policy→400 на non-stream |
| handler tests | asserts на forbidden/invalid; stream gap закрыт preflight на service |

## Drift

В списке A1 нет `project_access.go` (добор TASK-008) — вне слота; wiring использует `SetAccessChecker`.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **11 / 11 / 0** |
| Находок | **5** (NEW 0 · REGRESS 2 · RESIDUAL 3) |
| P0 | 0 |
| P1 | 0 |
| P2 | 4 |
| P3 | 1 |
