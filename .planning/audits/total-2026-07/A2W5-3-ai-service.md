# A2-W5 агент A3 — ai service/repo/bot (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc`  
> Зеркало: `W5-ASSIGNMENTS.md` §A3 · Stage A: `A5-3-ai-service.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/admin_settings.go` | 54 | 0 | чисто |
| `backend/internal/ai/audit.go` | 32 | 0 | чисто |
| `backend/internal/ai/bot_commands.go` | 115 | 0 | чисто |
| `backend/internal/ai/bot_user.go` | 49 | 1 | замечания |
| `backend/internal/ai/code_review.go` | 132 | 1 | замечания |
| `backend/internal/ai/composer_suggest.go` | 104 | 1 | замечания |
| `backend/internal/ai/doc_search.go` | 59 | 0 | чисто |
| `backend/internal/ai/issue_lookup.go` | 85 | 0 | чисто |
| `backend/internal/ai/messenger_bot.go` | 347 | 1 | замечания |
| `backend/internal/ai/preference_input.go` | 10 | 0 | чисто |
| `backend/internal/ai/repository.go` | 78 | 0 | чисто |
| `backend/internal/ai/repository_ent.go` | 486 | 1 | замечания |
| `backend/internal/ai/review_context.go` | 43 | 0 | чисто |
| `backend/internal/ai/service.go` | 640 | 0 | чисто (FIXED-OK stream/privacy) |
| `backend/internal/ai/settings_store.go` | 61 | 1 | замечания |
| `backend/testframework/ai/admin_settings_test.go` | 73 | 0 | чисто |
| `backend/testframework/ai/audit_test.go` | 30 | 1 | замечания |
| `backend/testframework/ai/bot_commands_test.go` | 71 | 0 | чисто |
| `backend/testframework/ai/bot_user_test.go` | 32 | 0 | чисто |
| `backend/testframework/ai/code_review_test.go` | 74 | 0 | чисто |
| `backend/testframework/ai/code_review_unified_test.go` | 237 | 0 | чисто |
| `backend/testframework/ai/composer_suggest_test.go` | 61 | 0 | чисто |
| `backend/testframework/ai/doc_search_test.go` | 25 | 0 | чисто |
| `backend/testframework/ai/issue_lookup_test.go` | 32 | 1 | замечания |
| `backend/testframework/ai/messenger_bot_flow_test.go` | 436 | 0 | чисто |
| `backend/testframework/ai/messenger_bot_test.go` | 69 | 0 | чисто |
| `backend/testframework/ai/repository_ent_test.go` | 234 | 1 | замечания |
| `backend/testframework/ai/service_test.go` | 527 | 0 | чисто |
| `backend/testframework/ai/settings_store_test.go` | 34 | 0 | чисто |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки

| ID | Sev | Tag | Файл:строка | Проблема | A1 / задача | Мутация |
|----|-----|-----|-------------|----------|-------------|---------|
| A2W5-A3-001 | P2 | REGRESS | `composer_suggest.go:98-100` | Trim prefix через **`s[len(prefix):]`** после case-fold — байты ≠ руны → panic/битая руна | W5-A3-010 · P2-BE-17 | Prefix `Тест` / mixed case → без panic, корректный хвост |
| A2W5-A3-002 | P2 | REGRESS | `repository_ent.go:248-273` | **IncrementUsage** check-then-act без upsert/tx | W5-A3-009 · P2-BE-17 | Parallel first-of-day → без duplicate/loss |
| A2W5-A3-003 | P2 | REGRESS | `settings_store.go:42-59` | Upsert Update→Create race; value plaintext (контракт non-secret не enforced) | W5-A3-008 · P2-BE-17 | Concurrent Upsert одного key → 1 row |
| A2W5-A3-004 | P2 | FIXED | `messenger_bot.go` | `/ai docs` → fixed string + slog | TASK-113 | `TestMessengerBot_HandleCommand_DocsSearchError` |
| A2W5-A3-005 | P2 | NEW | `code_review.go:71-72` | `authorizeReview`: **`access == nil` → allow** (fail-open). Wiring ставит checker, но API допускает nil | TASK-085 | `access==nil` → deny / error «not configured» |
| A2W5-A3-006 | P3 | FIXED | `bot_user.go` conflict retry | Parallel Ensure → 1 user | TASK-113 | `TestEnsureBotUser_Parallel_TASK113` |
| A2W5-A3-007 | P3 | FIXED | `issue_lookup_test.go` | foreign project → NotFound | TASK-113 | `TestTrackerIssueLookup_ForeignProjectNotFound_TASK113` |
| A2W5-A3-008 | P3 | FIXED | `audit_test.go` | details без prompt/content keys | TASK-113 | assert в `TestAuditRepoLogger_LogAIChat` |
| A2W5-A3-009 | P3 | FIXED | `repository_ent_test.go` | `assert.Empty(APIKeyEnc)` | TASK-113 | GetUserProvider redact |

## FIXED-OK

| A1 ID | Evidence |
|-------|----------|
| W5-A3-001 | `shouldRespond`: bot∉members → false; flow/test coverage |
| W5-A3-002 | `IssueContext(ctx, userID, key)` + `CanViewProject` |
| W5-A3-003 | `FetchReview` + `authorizeReview` + `SetAccessChecker` in wiring |
| W5-A3-005 | stream: `WithoutCancel`+timeout, не `Background` |
| W5-A3-006/007 | encrypt/decrypt: nil cipher → `ErrEncryptionRequired` |
| W5-A3-011 | DM требует mention/`/ai` |
| W5-A3-012 | `AllowExternalAI` default false |
| W5-A3-013/014 | GitLab nil explicit; nil lookup → error |

## Чисто

`admin_settings` prompt cap; `audit` metadata-only; `doc_search` передаёт userID; `repository` GetUserProvider обнуляет key; bot commands parse без SQL exec.

## Drift

`backend/internal/ai/project_access.go` — вне assignments (добор TASK-008); используется wiring. Не в таблице «прочитанные» слота A3.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **29 / 29 / 0** |
| Находок | **9** (NEW 1 · REGRESS 3 · RESIDUAL 5) |
| P0 | 0 |
| P1 | 0 |
| P2 | 5 |
| P3 | 4 |
