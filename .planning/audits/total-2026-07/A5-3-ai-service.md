# W5 агент A3 — ai service/repo/bot

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: privacy, messenger bot authz, settings store secrets, code review scope.  
> Список: `W5-ASSIGNMENTS.md` §A3 (29 файлов). Код не менялся.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/admin_settings.go` | 55 | 0 | чисто |
| `backend/internal/ai/audit.go` | 33 | 0 | чисто |
| `backend/internal/ai/bot_commands.go` | 116 | 0 | чисто |
| `backend/internal/ai/bot_user.go` | 46 | 1 | замечания |
| `backend/internal/ai/code_review.go` | 83 | 2 | замечания |
| `backend/internal/ai/composer_suggest.go` | 99 | 1 | замечания |
| `backend/internal/ai/doc_search.go` | 60 | 0 | чисто |
| `backend/internal/ai/issue_lookup.go` | 63 | 2 | замечания |
| `backend/internal/ai/messenger_bot.go` | 334 | 4 | замечания |
| `backend/internal/ai/preference_input.go` | 11 | 0 | чисто |
| `backend/internal/ai/repository.go` | 79 | 0 | чисто |
| `backend/internal/ai/repository_ent.go` | 462 | 3 | замечания |
| `backend/internal/ai/review_context.go` | 44 | 0 | чисто |
| `backend/internal/ai/service.go` | 603 | 2 | замечания |
| `backend/internal/ai/settings_store.go` | 62 | 1 | замечания |
| `backend/testframework/ai/admin_settings_test.go` | 74 | 0 | чисто |
| `backend/testframework/ai/audit_test.go` | 31 | 1 | замечания |
| `backend/testframework/ai/bot_commands_test.go` | 72 | 0 | чисто |
| `backend/testframework/ai/bot_user_test.go` | 33 | 0 | чисто |
| `backend/testframework/ai/code_review_test.go` | 74 | 0 | чисто |
| `backend/testframework/ai/code_review_unified_test.go` | 235 | 0 | чисто |
| `backend/testframework/ai/composer_suggest_test.go` | 63 | 0 | чисто |
| `backend/testframework/ai/doc_search_test.go` | 26 | 0 | чисто |
| `backend/testframework/ai/issue_lookup_test.go` | 30 | 1 | замечания |
| `backend/testframework/ai/messenger_bot_flow_test.go` | 368 | 1 | замечания |
| `backend/testframework/ai/messenger_bot_test.go` | 49 | 1 | замечания |
| `backend/testframework/ai/repository_ent_test.go` | 232 | 1 | замечания |
| `backend/testframework/ai/service_test.go` | 523 | 1 | замечания |
| `backend/testframework/ai/settings_store_test.go` | 35 | 0 | чисто |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W5-A3-001 | P1 | `messenger_bot.go:242-244` | `@hub-ai` / mention в non-DM канале включает ответ **без** проверки, что `botID ∈ memberIDs` | `ListMessagesForBot` / `SendBotMessage` обходят membership. Участник private-канала, куда бот не добавлен, через mention читает историю (до `botHistoryLimit`/`botSummarizeLimit`) во внешний LLM и пишет ответ в канал — privacy + authz | В `shouldRespond` для group/private требовать `botID` в `memberIDs` (как для DM); иначе `return false`. Тест: public/private без bot member + mention → no reply / no history load |
| W5-A3-002 | P1 | `issue_lookup.go:45`; `messenger_bot.go:137` | `IssueContext` → `tracker.GetIssue` **без** `userID` / project membership | Любой, кто может дернуть бота, делает `/ai test-cases OTHER-42` и получает title+description чужой задачи в LLM и в тред чата (IDOR через bot, класс RT-007 / A3-002) | Прокинуть `evt.SenderID`; lookup только если user имеет доступ к проекту issue (RBAC/membership). Отказ → «не найдена» без утечки. Тест: foreign project key → denial |
| W5-A3-003 | P1 | `messenger_bot.go:157`; `code_review.go:31-49` | `/ai review` тянет PR/MR через **org-wide** GitHub/GitLab integration token без проверки, что `SenderID` состоит в проекте, к которому привязана интеграция | Любой участник канала с доступом к боту выгружает diff/body любого подключённого repo в чат и во внешний LLM (code review scope / data exfil) | Перед `FetchReview`: резолв integration → `projectID` → membership/RBAC для `evt.SenderID`. Иначе 403/«недоступно». Тест: user вне project → fetch denied |
| W5-A3-004 | P2 | `messenger_bot.go:159`; `messenger_bot.go:172` | В канал уходит `err.Error()` (`Не удалось загрузить PR/MR: …`, `Ошибка поиска: …`) | Внутренние детали интеграций/search (хосты, статусы API) попадают всем читателям треда | Маппить на фиксированные user-facing строки; сырой err только в `slog` |
| W5-A3-005 | P2 | `service.go:467` | `StartChatStream` запускает `runChatStream(context.Background(), …)` | Теряются cancel/deadline/trace родительского запроса; стрим живёт до своего конца после disconnect клиента | `context.WithoutCancel(ctx)` или явный timeout + propagate trace; отмена по WS/disconnect |
| W5-A3-006 | P2 | `repository_ent.go:42-43` | `decryptField`: `pt, _ := MaybeDecryptString` — ошибка decrypt глотается | Битый/чужой ciphertext уходит в HTTP Authorization как «ключ»; трудно диагностировать, риск странных outbound-запросов | При ошибке decrypt → return/log error, не использовать ciphertext; fail chat с явной ошибкой конфига |
| W5-A3-007 | P2 | `repository_ent.go:32-35` | `encryptField` при `cipher == nil` пишет API key **plaintext** в `api_key_enc` | Fail-open: любой `NewEntRepository(db, nil)` (тесты уже так) = секреты в БД открытым текстом. Wiring обычно передаёт cipher, но инвариант не enforced | `cipher == nil && v != ""` → error «encryption required»; тест: nil cipher + CreateProvider → err, в БД нет plaintext |
| W5-A3-008 | P2 | `settings_store.go:42-59` | `Upsert`: Update → если `n==0` Create; нет обработки unique race; value всегда plaintext | Два concurrent Upsert одного key → один Create падает / нестабильность. Store общий SystemSetting: секреты сюда класть нельзя (сейчас prompt — ок, контракт не запрещает) | `ON CONFLICT` / Get+Create с `IsConstraintError` retry; документировать «только non-secret»; секреты — только `AIProvider`+cipher |
| W5-A3-009 | P2 | `repository_ent.go:223-248` | `IncrementUsage` check-then-act без tx/unique upsert | Параллельные первые запросы дня → duplicate key или потерянный счётчик → квота `enforceQuota` неточна | `INSERT … ON CONFLICT DO UPDATE` / Ent upsert по (user, usage_date); тест race `-race` на parallel Increment |
| W5-A3-010 | P2 | `composer_suggest.go:93-94` | Case-insensitive trim: `s[len(prefix):]` после `ToLower` match | Длина в байтах `prefix` ≠ префикса в `s` (регистр/UTF-8) → panic slice bounds или битая руна | Сравнивать/резать по runes или `strings.EqualFold` + корректный индекс; тест: `Prefix:"Тест"` / completion с иным case |
| W5-A3-011 | P2 | `messenger_bot.go:235-240` | В DM с ботом **любое** сообщение (без mention/`/ai`) → stream + history во внешний LLM | Privacy default: случайный текст в DM уходит провайдеру (consent `AllowExternalAI` есть на resolve path, но UX-opt-in на каждое сообщение нет) | Требовать mention или `/ai` и в DM; либо явный «AI mode» флаг канала |
| W5-A3-012 | P2 | `service.go:302-306` | `GetUserPreference` default `AllowExternalAI: true` при отсутствии записи | Privacy opt-out, не opt-in: новый user сразу cloud-eligible (согласовано с `privacy.externalAIAllowed` nil→true, вне списка A3, но сервис закрепляет контракт) | Default `false` + явный opt-in в Settings; миграция существующих |
| W5-A3-013 | P3 | `code_review.go:35-49` | GitLab-like URL при `gitlab==nil` падает в GitHub path | Пользователь видит ошибку про формат GitHub вместо «GitLab не настроен» | Если `gitlabMRURLRe` match и gitlab nil → явная ошибка; не вызывать github |
| W5-A3-014 | P3 | `issue_lookup.go:34-36` | Nil `TrackerIssueLookup` / deps → `("", "", nil)` | `/ai test-cases` получает пустые title/desc вместо «недоступен» — тихий деград | Возвращать error «issue lookup not configured» |
| W5-A3-015 | P3 | `bot_user.go:24-44` | `EnsureBotUser` create без unique-conflict retry; прямой `*ent.Client` | Два параллельных bootstrap → один err; layering (bootstrap-исключение) | `ON CONFLICT` / IsConstraintError → re-query |
| W5-A3-016 | P3 | `messenger_bot_test.go:26-32` | `shouldRespond` тест не покрывает «mention без bot в memberIDs → false» | Регресс W5-A3-001 не краснеет | Assert: group + mention + members без botID → false |
| W5-A3-017 | P3 | `messenger_bot_flow_test.go:36-47` (и аналоги) | Большинство flow: только `require.NoError`, без assert текста/числа bot-сообщений | Сломанный reply/stream/команда остаётся зелёной | Assert через HookTestRepo: placeholder→final content, error replies |
| W5-A3-018 | P3 | `repository_ent_test.go:109-111` | `GetUserProvider` не `assert.Empty(APIKeyEnc)` | Мутация снятия `p.APIKeyEnc = ""` (`repository_ent.go:219`) не ловится | `assert.Empty(t, gotUser.APIKeyEnc)` |
| W5-A3-019 | P3 | `audit_test.go:29` | `LogAIChat` success без проверки details | Не доказывает privacy-инвариант «нет prompt/response в audit» | Read audit row: keys только provider/model/tokens/latency; нет content |
| W5-A3-020 | P3 | `issue_lookup_test.go:11-22` | Только nil-safe; нет RBAC/foreign-key denial | W5-A3-002 не закрыт тестом на слое lookup | Integration: foreign project issue key + user without membership → error |
| W5-A3-021 | P3 | `service_test.go` (нет Test*Consent в файле) | Нет теста default/preference `AllowExternalAI` на service API в этом наборе | Смена default true→false / Save preference не закреплена здесь (частично A2 privacy_test) | `GetUserPreference` без записи → ожидаемый default; Save false → resolveProviders фильтрует cloud |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `admin_settings.go` | Prompt trim/8000 cap; empty→default; nil settings→`ErrSettingsNotConfigured`; OfflineOnly read-only с env |
| `audit.go` | Metadata only (provider/model/tokens/latency); prompt/response не пишутся |
| `bot_commands.go` | Парсинг slash; SQL prompt только SELECT-инструкция (exec нет); issue key / PR URL extract |
| `doc_search.go` | `Search(ctx, query, userID, …)` — user scope на search (NF-03 делегирован); limit 8 |
| `preference_input.go` | DTO-only, секретов нет |
| `repository.go` | Интерфейс; AuditLogger контракт privacy |
| `review_context.go` | Сборка prompt; объём diff ограничивается github/gitlab (`maxPRPatchBytes` 12k) вне файла |
| `settings_store.go` Get | NotFound→`""`; key `ai.bot_system_prompt` — не API key path |
| `admin_settings_test` / `bot_commands_test` / `bot_user_test` / `composer_suggest_test` / `doc_search_test` / `code_review*_test` / `settings_store_test` | Asserts есть; CRUD settings, parse commands, unified GH/GL fetch на httptest |
| API keys в repo (позитив) | `CreateProvider`/`UpdateProvider` через `encryptField` при cipher; `GetUserProvider` обнуляет key; `ValidateOutboundURL` на BaseURL; `toProviderResponse` только `HasAPIKey` |
| Consent path (смежно A2) | `resolveProviders` в `fallback.go` зовёт `externalAIAllowed` + `filterProvidersByConsent` — chat/bot stream проходят через него |

**Privacy (сводка):** audit без prompt — ок; doc search с userID — ок; defaults AllowExternalAI=true и DM auto-stream — риск; bot mention без membership и issue/review без user scope — основные дыры.

**Messenger bot authz (сводка):** sender уже проверен messenger send-path; **бот** и **данные issue/PR** — нет.

**Settings store secrets (сводка):** сейчас хранит system prompt (не секрет); plaintext+upsert race; API keys живут в `AIProvider`+cipher (fail-open при nil cipher).

**Code review scope (сводка):** routing GH/GL ок; scope = org integration token без user↔project authz.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **29 / 29 / 0** |
| Находок | **21** |
| P0 | 0 |
| P1 | 3 |
| P2 | 9 |
| P3 | 9 |

### Топ-3

1. **W5-A3-001 (P1)** — `@hub-ai` без membership бота в канале → чтение истории private-канала во внешний LLM и запись ответа.
2. **W5-A3-002 (P1)** — `/ai test-cases KEY` без RBAC пользователя на issue/project (IDOR через бота).
3. **W5-A3-003 (P1)** — `/ai review` на org GitHub/GitLab token без membership пользователя в проекте интеграции (утечка diff в чат/LLM).
