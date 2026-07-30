# File-auditor — search / ai / ws

> TASK-008 · File-auditor · READ-ONLY · 2026-07-29  
> Модель: `cursor-grok-4.5` · ветка `cursor-grok-4.5`  
> HEAD: `617af22372be61b98e7ea63d825b35075f495588` (`617af223`)  
> Периметр: **все** production `.go` (не `*_test.go`) в  
> `backend/internal/search/` · `backend/internal/ai/` · `backend/internal/ws/`  
> Фокус: **NF-03** user-scope · **SSRF** · **consent** · **WS membership NF-02** · **sanitize**  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 остаётся открытым**

Связанные A3 (не дублировать закрытие TASK-008):  
[A3W5-4-search-core.md](A3W5-4-search-core.md) · [A3W5-1-ai-handlers.md](A3W5-1-ai-handlers.md) · [A3W5-2-ai-adapters.md](A3W5-2-ai-adapters.md) · [A3W5-3-ai-service.md](A3W5-3-ai-service.md) · [A3W3-4-ws.md](A3W3-4-ws.md)  
Stage C: [C-W19-auth-rbac.md](C-W19-auth-rbac.md) (NF-02/NF-03 + mid-session WS)

## Вердикт

На HEAD фокус-инварианты **держатся**: Ctrl+K scoped по membership + soft-delete; Meili/AI outbound через SafeDial (+ operator/offline allowlist); external AI opt-in consent; WS `CanSubscribe` fail-closed + ticket handshake. Remediation A3 (TASK-146/153–156 и соседние) в коде **FIXED-OK**. **Открытые находки этого прохода:** 2×P3 sanitize (Meili error-body swallow; AI handler `err.Error()` на AllProvidersFailed / catch-all AssistJQL·SuggestComposer). P0/P1/P2 по фокусу — **0**.

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 2 |
| **Total open** | **2** |

## Прочитанные файлы (49/49)

### search (9)

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/search/dto.go` | 50 | 0 | чисто |
| `backend/internal/search/export_hooks.go` | 136 | 0 | hubtest-only (`//go:build hubtest`) |
| `backend/internal/search/handler.go` | 253 | 0 | auth Group; rune-safe `q`≤500; filters 403/404 |
| `backend/internal/search/indexer.go` | 59 | 0 | idempotent tsvector DDL |
| `backend/internal/search/jql.go` | 427 | 0 | FIXED-OK A3W5-4-002 — `ErrUnsupportedJQLOperator` |
| `backend/internal/search/meilisearch_indexer.go` | 196 | 1 | FIXED-OK dial/pagination; **SAW-001** ReadAll `_` |
| `backend/internal/search/repository.go` | 108 | 0 | membership + admin/founder; `*ent.Client` только repo |
| `backend/internal/search/service.go` | 715 | 0 | FIXED-OK NF-03 shards + soft-delete + outage |
| `backend/internal/search/truncate.go` | 14 | 0 | rune-safe truncate |

### ai (33)

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/ai/adapter_anthropic.go` | 138 | 0 | ValidateOutboundURL + SafeHTTPClient + LimitReader |
| `backend/internal/ai/adapter_gigachat.go` | 150 | 0 | OAuth+base Validate; ReadAll err checked |
| `backend/internal/ai/adapter_openai.go` | 131 | 0 | FIXED-OK offlineStreamClient (A3W5-2-001) |
| `backend/internal/ai/adapter_stream.go` | 98 | 0 | FIXED-OK ReadAll err (A3W5-2-002); ctx cancel |
| `backend/internal/ai/adapter_yandex.go` | 166 | 0 | hard-coded URL Validate + host allowlist |
| `backend/internal/ai/admin_handler.go` | 144 | 0 | requireAdmin fail-closed; mapProviderMutateError |
| `backend/internal/ai/admin_settings.go` | 48 | 0 | prompt cap |
| `backend/internal/ai/audit.go` | 27 | 0 | metadata-only (без prompt) |
| `backend/internal/ai/bot_commands.go` | 101 | 0 | SELECT-only SQL prompt |
| `backend/internal/ai/bot_user.go` | 65 | 0 | FIXED-OK A3W5-3-008 — `BotUserRepository` |
| `backend/internal/ai/client_errors.go` | 51 | 0 | sanitize mutate/test errors |
| `backend/internal/ai/code_review.go` | 122 | 0 | access==nil fail-closed |
| `backend/internal/ai/composer_suggest.go` | 107 | 0 | FIXED-OK A3W5-3-001 EqualFold runes |
| `backend/internal/ai/content_policy.go` | 140 | 0 | NFKC+ZW+homoglyph+compact denylist |
| `backend/internal/ai/doc_search.go` | 52 | 0 | Search + userID (NF-03 через search) |
| `backend/internal/ai/dto.go` | 155 | 0 | `has_api_key`, без plaintext |
| `backend/internal/ai/export_hooks.go` | 232 | 0 | hubtest-only |
| `backend/internal/ai/fallback.go` | 199 | 0 | consent filter; offline host allowlist |
| `backend/internal/ai/handler.go` | 284 | 1 | FIXED stream NoProvider; **SAW-002** err.Error leak |
| `backend/internal/ai/http_stream_client.go` | 30 | 0 | SafeDial / AllowLoopback stream clients |
| `backend/internal/ai/issue_lookup.go` | 78 | 0 | FIXED-OK A3W5-3-004 fail-closed |
| `backend/internal/ai/messenger_bot.go` | 319 | 0 | membership + DM mention; sanitize replies |
| `backend/internal/ai/preference_input.go` | 11 | 0 | чисто |
| `backend/internal/ai/privacy.go` | 49 | 0 | nil pref → allowExternal=false |
| `backend/internal/ai/project_access.go` | 29 | 0 | PermProjectView; nil svc → false |
| `backend/internal/ai/provider_rpm.go` | 56 | 0 | Redis fail-closed + ZRem |
| `backend/internal/ai/repository.go` | 72 | 0 | интерфейс |
| `backend/internal/ai/repository_ent.go` | 497 | 0 | SSRF BaseURL; consent create false; VersionEQ; IncrementUsage conflict-retry |
| `backend/internal/ai/review_context.go` | 40 | 0 | чисто |
| `backend/internal/ai/service.go` | 612 | 0 | consent/RPM/stream sanitize WS (`msgAIStreamFailed`) |
| `backend/internal/ai/settings_store.go` | 145 | 0 | FIXED Upsert conflict+VersionEQ |
| `backend/internal/ai/stream.go` | 32 | 0 | SendToUser; payload без secrets |
| `backend/internal/ai/user_handler.go` | 82 | 0 | BYOK owner-scope + mapProviderMutateError |

### ws (7)

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/ws/authorizer.go` | 81 | 0 | NF-02 fail-closed (nil checker / unknown prefix) |
| `backend/internal/ws/bridge.go` | 240 | 0 | revoke/logout → UnregisterBySession |
| `backend/internal/ws/export_hooks.go` | 90 | 0 | hubtest-only |
| `backend/internal/ws/handler.go` | 302 | 0 | ticket; deny `?token=`; nil authz; CanSubscribe; ME-13 upgrade |
| `backend/internal/ws/hub.go` | 398 | 0 | multi-tab; late subscribe; SEC-05 send |
| `backend/internal/ws/protocol.go` | 32 | 0 | SubscribeDenied/Error events |
| `backend/internal/ws/ticket_protocol.go` | 44 | 0 | `hub.ws.ticket.` parse |

Файлов в периметре: **49**. Прочитано: **49**. Пропущено: **0**.

## Находки (активные)

| ID | Severity | Файл:строка | Фокус | Проблема | Почему дефект | Как чинить |
|----|----------|-------------|-------|----------|---------------|------------|
| SAW-001 | P3 | `search/meilisearch_indexer.go:206–208` | sanitize | `b, _ := io.ReadAll(LimitReader(...))` — ошибка чтения error-body проглочена | Тот же класс, что уже чинили в GigaChat/SSE: truncated/read err → misleading `meilisearch %d: ` без тела / скрытый I/O fail | `if err != nil { return fmt.Errorf("meilisearch read error body: %w", err) }` |
| SAW-002 | P3 | `ai/handler.go:149`, `:158`, `:193`, `:202`, `:270` (+ fallback wrap) | sanitize | `ErrAllProvidersFailed` и catch-all AssistJQL/SuggestComposer отдают **`err.Error()`**; при этом `fallback.go` wrap’ит `%w: %v` с upstream lastErr | ChatCompletions catch-all уже `"provider error"`; stream WS — `msgAIStreamFailed`. Здесь клиент видит `provider HTTP 4xx: …` / internals. Асимметрия sanitize | Для `ErrAllProvidersFailed` → стабильный `"all AI providers failed"`; AssistJQL/SuggestComposer catch-all → `"provider error"`; detail только slog |

**REGRESS относительно A3W5/A3W3 фокус-инвариантов:** 0  
**REF (вне новых ID, уже Stage C):** mid-session WS после admin deactivate — [C-W19-001](C-W19-auth-rbac.md) / TASK-177; в `ws/` на upgrade ME-13 OK, gap в auth→NATS revoke path.

## Сверка A3 → HEAD (фокус)

| A3 ID | Tag | Статус в HEAD | Evidence |
|-------|-----|---------------|----------|
| A3W5-4-001 | FIXED-OK | закрыт | `NewOperatorMeiliHTTPClient` после `meiliAllowsOperatorPrivate` (`meilisearch_indexer.go:53–56`) |
| A3W5-4-002 | FIXED-OK | закрыт | `validateJQLCondition` в ParseJQL + never-match в ToEnt (`jql.go:105–113`, `:328–330`) |
| A3W5-4-003 | FIXED-OK | закрыт | offset-пагинация `IndexAllIssues` (`meilisearch_indexer.go:132–183`) |
| A3W5-1-001 | FIXED-OK | закрыт | stream `ErrNoProvider` → 503 (`handler.go:236–238`) |
| A3W5-1-002 | FIXED-OK | закрыт | `ErrQueryRequired` → 400 (`handler.go:154–156`, `service.go:563–564`) |
| A3W5-2-001 | FIXED-OK | закрыт | `offlineStreamClient` (`adapter_openai.go:48–49`, `:86–88`) |
| A3W5-2-002 | FIXED-OK | закрыт | ReadAll err в SSE error path (`adapter_stream.go:52–56`) |
| A3W5-2-003 | FIXED-OK* | закрыт* | denylist + NFKC/ZW/homoglyph/compact (`content_policy.go`); *denylist никогда не полный — residual by design, не open bug |
| A3W5-3-001 | FIXED-OK | закрыт | `equalFoldPrefixLen` (`composer_suggest.go:99–119`) |
| A3W5-3-002 | FIXED-OK | закрыт | IncrementUsage create-first + conflict Update (`repository_ent.go:263–294`) |
| A3W5-3-003 | FIXED-OK | закрыт | Upsert Create→conflict→Update + VersionEQ (`settings_store.go:56–155`) |
| A3W5-3-004 | FIXED-OK | закрыт | issue `access==nil` deny (`issue_lookup.go:58–60`) |
| A3W5-3-005 | FIXED-OK | закрыт | create `SetAllowExternalAi(false)` + schema `Default(false)` |
| A3W5-3-006 | FIXED-OK | закрыт | VersionEQ на provider/pref/settings |
| A3W5-3-007 | FIXED-OK | закрыт | WS stream error = `msgAIStreamFailed` (`service.go:539–542`) |
| A3W5-3-008 | FIXED-OK | закрыт | `EntBotUserRepository` (`bot_user.go`) |
| A3W3-4-001 | FIXED-OK | закрыт | query+ticket deny (`handler.go:157–160`); TASK-146 |

## Фокус-чеклист (сводка)

| Инвариант | search | ai | ws |
|-----------|--------|----|----|
| **NF-03 scope** | **OK** — `projectIDs` / channel membership / Mongo `$in` + empty→nil; soft-delete `deleted_at IS NULL` / Mongo `nil` | **OK** — doc search через Search+userID; issue/review `CanViewProject`; bot∈members | N/A (не search) |
| **SSRF** | **OK** — Meili Validate + SafeDial; operator private только после allowlist; metadata/link-local reject | **OK** — ValidateOutboundURL / ValidateOfflineAIURL; SafeDial; offline AllowLoopback; нет `DefaultClient` | N/A |
| **Consent** | N/A | **OK** — nil pref / schema default false; `filterProvidersByConsent`; offline host allowlist (не substring ollama) | N/A |
| **NF-02 membership** | filters: `IsProjectMember`; channels membership | bot membership; review/issue access fail-closed | **OK** — `CanSubscribe` до Subscribe; nil authz → 503; unknown prefix deny |
| **Sanitize** | handler search → generic 500; filters 403/404 | TestProvider/mutate sanitized; stream WS fixed; **gap SAW-002** HTTP AllProvidersFailed/catch-all | deny/error frames без internals |

## Чисто (держат TZ)

- **NF-03 Ctrl+K:** issues/TC/projects/files — `ANY(projectIDs)` + soft-delete; channels — public OR membership; messages — `getUserChannelIDs` JOIN soft-delete + Mongo scope; users — shared-project (admin/founder unscoped).
- **JQL:** whitelist fields; unsupported comparison → error (не silent unfiltered).
- **Meili:** публичный host SafeDial; operator loopback/RFC1918 — отдельный dialer; 169.254/metadata → nil indexer.
- **AI consent:** opt-in; cloud BYOK BaseURL Validate; Ollama только loopback host после ValidateOfflineAIURL.
- **AI secrets:** API key encrypt; GetUserProvider redact; DTO `has_api_key`.
- **WS NF-02:** ticket subprotocol; `?token=` forbidden; Origin CORS allowlist; deactivated block на upgrade.
- **Layering:** `*ent.Client` только в EntRepository / SettingsStore / BotUserRepository / Meili IndexAllIssues(cmd); Service/Handler без Ent.
- Нет `http.DefaultClient`, нет `panic()` в prod-paths периметра.

## Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 2 |
| **Total open** | **2** |

| Tag | n |
|-----|--:|
| NEW | 2 (SAW-001…002) |
| REGRESS | 0 |
| FIXED-OK (A3 в периметре) | A3W5-4-001…003 · A3W5-1-001/002 · A3W5-2-001/002 · A3W5-3-001…008 · A3W3-4-001 (+ content policy layers) |

## Рекомендации оркестратору

1. **Не закрывать TASK-008** по этому отчёту — file-auditor только evidence.
2. SAW-001/002 — мелкий sanitize hygiene; можно прицепить к ближайшему AI/search remediation без отдельного P2 TASK.
3. TASK-146/153–156 по коду выглядят merge-ready с точки зрения prod evidence; статус README «Открыта/На проверке» — процессный drift для be-ver, не orphan finding.
4. C-W19-001 (WS mid-session после deactivate) — уже вынесено (TASK-177); в этом слоте не дублировать.

## Статус TASK-008

- Артефакт: `.planning/audits/total-2026-07/AUDIT-BE-search-ai-ws-2026-07-29.md`
- **Коммит не создавался**
- **TASK-008 не закрывать** (программа тотального аудита; нет mutation/COV gate)

---
reason: file-auditor READ-ONLY; 49/49 go read; 2 findings (P3 sanitize); A3 focus FIXED-OK; TASK-008 stays open; no commit
