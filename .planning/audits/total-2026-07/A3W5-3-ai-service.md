# A3 W5 агент 3 — ai service/repo/bot (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W5-3-ai-service.md](A2W5-3-ai-service.md) · Stage A: [A5-3-ai-service.md](A5-3-ai-service.md)  
> Фокус: policy/BYOK, soft-delete, optimistic lock, bot scopes, `*ent.Client` layering  
> Remediation: TASK-085 (review fail-closed) · TASK-113 (docs/bot/audit/tests) · TASK-044 (privacy default service)  
> Периметр: 29 файлов (15 prod + 14 tests) · Прочитано: **29/29** · Пропущено: **0**

## Сверка с A2 (A2W5-A3-* / W5-A3-*)

| A2 / A1 ID | Tag | Статус в HEAD | Комментарий |
|------------|-----|---------------|-------------|
| A2W5-A3-001 / W5-A3-010 | REGRESS | **открыт** | `composer_suggest.go:100` — `s[len(prefix):]` после case-fold; теста на `Тест`/UTF-8 нет |
| A2W5-A3-002 / W5-A3-009 | REGRESS | **открыт** | `IncrementUsage` check-then-act без upsert/tx (`repository_ent.go:248-273`) |
| A2W5-A3-003 / W5-A3-008 | REGRESS | **открыт** | `settings_store.Upsert` Update→Create race; plaintext value (`:42-59`) |
| A2W5-A3-004 | FIXED-OK | закрыт | `/ai docs` → fixed string + slog; `TestMessengerBot_HandleCommand_DocsSearchError` |
| A2W5-A3-005 / TASK-085 | FIXED-OK | закрыт | `authorizeReview`: `access==nil` → deny; `code_review_test` / `task085_regress_test` |
| A2W5-A3-006 | FIXED-OK | закрыт | `EnsureBotUser` conflict → re-query; parallel tests TASK-113/066 |
| A2W5-A3-007 | FIXED-OK | закрыт | foreign project → NotFound; `TestTrackerIssueLookup_ForeignProjectNotFound_TASK113` |
| A2W5-A3-008 | FIXED-OK | закрыт | audit details без prompt keys |
| A2W5-A3-009 | FIXED-OK | закрыт | `assert.Empty(APIKeyEnc)` на GetUserProvider |
| W5-A3-001 | FIXED-OK | закрыт | `shouldRespond`: bot∉members → false |
| W5-A3-002 | FIXED-OK | закрыт | `IssueContext(ctx, userID, key)` + CanViewProject (при wired checker) |
| W5-A3-003 | FIXED-OK | закрыт | FetchReview + authorizeReview + wiring `SetAccessChecker` |
| W5-A3-005 | FIXED-OK | закрыт | stream: `WithoutCancel`+5m timeout |
| W5-A3-006/007 | FIXED-OK | закрыт | encrypt/decrypt nil cipher → `ErrEncryptionRequired` |
| W5-A3-011 | FIXED-OK | закрыт | DM требует mention/`/ai` |
| W5-A3-012 | FIXED-OK* | частичный | service/privacy: nil record → false; **схема Ent Default(true)** + create без флага — см. NEW |
| W5-A3-013/014 | FIXED-OK | закрыт | GitLab nil explicit; nil lookup → error |
| W5-A3-015 layering | RESIDUAL | открыт | conflict retry есть; `*ent.Client` в `EnsureBotUser` остаётся |
| W5-A3-017 hollow flow | RESIDUAL | открыт | Docs assert есть; большинство flow — только `require.NoError` |

\* Service-level default закрыт; DB/create path — нет (A3W5-3-005).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/admin_settings.go` | 48 | 0 | чисто (prompt cap 8000) |
| `backend/internal/ai/audit.go` | 27 | 0 | FIXED-OK metadata-only |
| `backend/internal/ai/bot_commands.go` | 101 | 0 | чисто (SELECT-only SQL prompt) |
| `backend/internal/ai/bot_user.go` | 45 | 1 | FIXED-OK retry; RESIDUAL `*ent.Client` |
| `backend/internal/ai/code_review.go` | 122 | 0 | FIXED-OK fail-closed + GitLab guard |
| `backend/internal/ai/composer_suggest.go` | 92 | 1 | REGRESS byte trim |
| `backend/internal/ai/doc_search.go` | 52 | 0 | чисто (Search + userID) |
| `backend/internal/ai/issue_lookup.go` | 76 | 1 | FIXED-OK foreign deny; NEW nil-access fail-open |
| `backend/internal/ai/messenger_bot.go` | 319 | 0 | FIXED-OK membership/DM/docs sanitize |
| `backend/internal/ai/preference_input.go` | 8 | 0 | чисто |
| `backend/internal/ai/repository.go` | 70 | 0 | чисто (интерфейс) |
| `backend/internal/ai/repository_ent.go` | 460 | 2 | REGRESS IncrementUsage; NEW version ignore / consent create |
| `backend/internal/ai/review_context.go` | 40 | 0 | чисто |
| `backend/internal/ai/service.go` | 601 | 1 | FIXED-OK stream/privacy/BYOK; NEW WS err leak |
| `backend/internal/ai/settings_store.go` | 53 | 1 | REGRESS Upsert race |
| `backend/testframework/ai/admin_settings_test.go` | 65 | 0 | чисто |
| `backend/testframework/ai/audit_test.go` | 38 | 0 | FIXED-OK banned keys |
| `backend/testframework/ai/bot_commands_test.go` | 56 | 0 | чисто |
| `backend/testframework/ai/bot_user_test.go` | 26 | 0 | чисто (happy + nil) |
| `backend/testframework/ai/code_review_test.go` | 73 | 0 | FIXED-OK nil access deny |
| `backend/testframework/ai/code_review_unified_test.go` | 236 | 0 | чисто (GH/GL + access) |
| `backend/testframework/ai/composer_suggest_test.go` | 55 | 0 | gap: нет UTF-8 case-fold (см. REGRESS) |
| `backend/testframework/ai/doc_search_test.go` | 20 | 0 | чисто |
| `backend/testframework/ai/issue_lookup_test.go` | 43 | 0 | FIXED-OK foreign; нет nil-access deny |
| `backend/testframework/ai/messenger_bot_flow_test.go` | 393 | 1 | FIXED-OK docs; RESIDUAL hollow asserts |
| `backend/testframework/ai/messenger_bot_test.go` | 55 | 0 | FIXED-OK membership/DM |
| `backend/testframework/ai/repository_ent_test.go` | 201 | 0 | FIXED-OK redact; нет race Increment |
| `backend/testframework/ai/service_test.go` | 483 | 0 | чисто (мок-контур) |
| `backend/testframework/ai/settings_store_test.go` | 27 | 0 | нет concurrent Upsert |

Файлов в списке: **29**. Прочитано: **29**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W5-3-001 | REGRESS | P2 | `composer_suggest.go:97-101` | Case-insensitive trim: после `ToLower` match режет `s[len(prefix):]` | Байты `prefix` ≠ префикса в `s` (case/UTF-8) → panic / битая руна. A2W5-A3-001 не закрыт | Резать по rune/EqualFold-индексу; тест `Prefix:"Тест"` + mixed-case completion |
| A3W5-3-002 | REGRESS | P2 | `repository_ent.go:248-273` | `IncrementUsage` Query→Create/Update без upsert/tx | Parallel first-of-day → unique violation или потеря счётчика → квота неточна | `ON CONFLICT DO UPDATE` / Ent upsert по (user, usage_date); `-race` parallel test |
| A3W5-3-003 | REGRESS | P2 | `settings_store.go:42-59` | Upsert: Update n==0 → Create; unique race | Concurrent admin prompt update → один Create падает / нестабильность | Constraint retry / `ON CONFLICT`; документировать non-secret only |
| A3W5-3-004 | NEW | P2 | `issue_lookup.go:58-67` | `if l.access != nil` — **fail-open** без checker | Зеркало A2W5-A3-005 (review уже fail-closed). `NewTrackerIssueLookup` без `SetAccessChecker` → IDOR issue context в `/ai test-cases` | `access==nil` → error «not configured»; тест как у code_review |
| A3W5-3-005 | NEW | P2 | `repository_ent.go:355-372` (+ schema default) | Create preference без `AllowExternalAI` → Ent **Default(true)** | Service nil-record = false (FIXED), но Save только `DefaultModel` пишет row с `allow_external_ai=true` → silent cloud consent | Create: `SetAllowExternalAi(false)` если input nil; schema Default(false); IT: save model-only → allow=false |
| A3W5-3-006 | NEW | P2 | `repository_ent.go:111-149`, `:378-397`; `settings_store.go:43-46` | Update paths игнорируют `VersionMixin` | Схемы AIProvider/AIUserPreference/SystemSetting имеют version (TASK-095/ADR); lost update при concurrent admin/BYOK patch | `Where(VersionEQ).AddVersion(1)` → 409; прокинуть version в DTO |
| A3W5-3-007 | NEW | P3 | `service.go:529-532` | `publish("error", {message: err.Error()})` | Upstream/internal детали LLM уходят в WS клиенту (класс sanitize A2W5-A3-004) | Фиксированное «AI stream failed»; сырой err в slog |
| A3W5-3-008 | RESIDUAL | P3 | `bot_user.go:20-49` | `EnsureBotUser(ctx, *ent.Client)` | Layering: bootstrap держит Ent вне Repository (W5-A3-015 partial). Race retry FIXED-OK | Вынести в `BotUserRepository` / auth repo; wiring без raw client в ai package |
| A3W5-3-009 | RESIDUAL | P3 | `messenger_bot_flow_test.go` (большинство кейсов) | Flow: `require.NoError` без assert текста/числа bot-сообщений | Регресс reply/команд остаётся зелёным (W5-A3-017); Docs — исключение | Assert placeholder→final / error replies на HookTestRepo |

## Фокус-чеклист

| Тема | Вердикт |
|------|---------|
| **policy / BYOK** | Consent service-path OK (`AllowExternalAI` nil→false, `filterProvidersByConsent`). BYOK: `scope=user` + owner на List/Get; admin `requireSystemProvider` режет user. **Дыра:** schema/create default true (A3W5-3-005). Cipher fail-closed FIXED-OK |
| **soft-delete** | В слоте нет trash/deleted_at на AI entities; lifecycle = `enabled` / preference upsert. Hard-delete API отсутствует — **N/A**, регрессии soft-delete нет |
| **optimistic lock** | Mixin на схемах есть; **repo updates без VersionEQ** (A3W5-3-006) — rollout TASK-095 не дошёл до ai repo |
| **bot scopes** | Membership + DM mention FIXED-OK; review fail-closed FIXED-OK; issue foreign FIXED-OK; **issue nil-access fail-open** NEW (A3W5-3-004) |
| **`*ent.Client` layering** | `EntRepository` / `EntSettingsStore` — норма (repo). `EnsureBotUser(*ent.Client)` — RESIDUAL (A3W5-3-008). Service/MessengerBot — без Ent |

## Чисто

- Bot membership / DM opt-in / docs sanitize / review nil-access deny / encrypt required / stream ctx timeout / GetUserProvider redact / audit metadata / doc search userID / admin BYOK isolation — без регрессии FIXED-OK.
- Wiring `wiring_audit_ai.go`: cipher на repo, `SetAccessChecker` на issue+review, settings store, bot Ensure.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 6 |
| P3 | 3 |
| **Total open** | **9** |

| Tag | Count |
|-----|------:|
| NEW | 4 |
| REGRESS | 3 |
| RESIDUAL | 2 |
| FIXED-OK (сверка) | 15+ |

A2 open REGRESS 001–003 **не закрыты** remediation 081…123 · NEW consent/version/issue-nil/WS · RESIDUAL layering + hollow flow.
