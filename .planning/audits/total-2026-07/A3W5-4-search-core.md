# A3 W5 агент 4 — search prod + half tests (FINAL re-audit)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · ветка `cursor-grok-4.5`  
> Mirror: [A2W5-4-search-core.md](A2W5-4-search-core.md) · Assignment: [W5-ASSIGNMENTS.md](W5-ASSIGNMENTS.md) § A4  
> Периметр: **22** файла (8 prod + 14 tests) · Прочитано: **22/22** · Пропущено: **0**  
> Plan: [A3-PLAN.md](A3-PLAN.md) · NF-03 user-scope · SafeDial Meili

> **Вне списка assignment (evidence FIXED-OK, не в таблице чтения):**  
> `filter_safedial_task092_test.go`, `meili_ssrf_task044_test.go`,  
> `partial_flag_task044_test.go`, `service_test.go` (ErrFilterForbidden / outage) — слот A5 / соседние.

## Сверка с A2 (A2W5-A4-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W5-A4-001 | FIXED-OK | закрыт | TASK-092: `ErrFilterForbidden`/`ErrFilterNotFound` → handler 403/404 (`handler.go:206–215`, `254–261`); mutation-ready `filter_safedial_task092_test.go` |
| A2W5-A4-002 | FIXED-OK* | SSRF-intent закрыт; ops residual → A3 | Client = `safehttp.NewSafeHTTPClient(30s)` (`meilisearch_indexer.go:50–51`); metadata dial blocked. *Operator loopback/RFC1918 validate≠dial — см. A3W5-4-001 (REGRESS ops) |
| A2W5-A4-003 | FIXED-OK | закрыт (ещё в A2) | `ErrSearchUnavailable` при `shardErr>0 && shardOK==0` (`service.go:185–190`); hubtest в слоте |
| A2W5-A4-004 | FIXED-OK | закрыт | Sentinel `ErrFilterForbidden` на delete/share non-owner (`service.go:735–736`, `750–751`) + handler `errors.Is` |

**REGRESS (функционал Meili ops):** 1 — A3W5-4-001 (после TASK-092 SafeDial).

### Cascade FIXED-OK (Stage A W5-A4-* / A2)

| Stage A / A2 | Evidence (файл:строка) |
|--------------|------------------------|
| W5-A4-001/002 | `searchChannels`: `deleted_at IS NULL` + public OR membership — `service.go:549–553` |
| W5-A4-003 | Mongo `"deleted_at": nil` — `service.go:624–628` |
| W5-A4-004 | `getUserChannelIDs` JOIN channels + soft-delete — `service.go:200–206` |
| W5-A4-005 | `ToEntPredicates` AND/OR — `jql.go:48–86` |
| W5-A4-007 | `Search` → `(results, partial, err)` — `service.go:61`, `192–193` |
| W5-A4-008 | rune-safe `q` max 500 — `handler.go:68–71` |
| W5-A4-009 | `validateMeiliHost` + metadata/link-local — `meilisearch_indexer.go:58–82` |
| W5-A4-010 | `truncateRunes` (TASK-060; вне слота `truncate.go`, export в hooks) |
| W5-A4-012 | OR union — hubtest вне/рядом |
| A2W5-A4-001/004 | TASK-092 filter HTTP status |
| A2W5-A4-002 | TASK-092 SafeDial client (SSRF mid-flight) |
| A2W5-A4-003 | TASK-113 total outage hard error |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/search/dto.go` | 58 | 0 | чисто |
| `backend/internal/search/export_hooks.go` | 152 | 0 | чисто — SafeDial/validate exports |
| `backend/internal/search/handler.go` | 283 | 0 | FIXED-OK 403/404 filters + rune q |
| `backend/internal/search/indexer.go` | 67 | 0 | чисто — idempotent tsvector DDL |
| `backend/internal/search/jql.go` | 435 | 1 | NEW — silent drop comparison ops |
| `backend/internal/search/meilisearch_indexer.go` | 178 | 2 | REGRESS dial + Limit 5000 |
| `backend/internal/search/repository.go` | 123 | 0 | чисто — membership + layering |
| `backend/internal/search/service.go` | 782 | 0 | FIXED-OK NF-03 shards + outage |
| `backend/testframework/search/handler_flow_test.go` | 128 | 0 | чисто |
| `backend/testframework/search/handler_wave100_test.go` | 206 | 0 | чисто — ErrFilterForbidden service |
| `backend/testframework/search/handler_wave109_test.go` | 64 | 0 | чисто — delete 404 |
| `backend/testframework/search/handler_wave112_test.go` | 98 | 0 | чисто — partial / outage |
| `backend/testframework/search/handler_wave116_test.go` | 33 | 0 | чисто |
| `backend/testframework/search/handler_wave93_test.go` | 63 | 0 | чисто |
| `backend/testframework/search/jql_assignee_a4001_test.go` | 71 | 0 | чисто — A4-001 mutation |
| `backend/testframework/search/jql_predicates_extended_test.go` | 94 | 0 | чисто |
| `backend/testframework/search/jql_test.go` | 171 | 0 | чисто (parse `>` не бьёт ToEnt) |
| `backend/testframework/search/jql_wave100_test.go` | 58 | 0 | чисто |
| `backend/testframework/search/jql_wave116_test.go` | 32 | 0 | чисто |
| `backend/testframework/search/meilisearch_indexer_test.go` | 106 | 0 | слабо dial/ops (сильнее TASK-092 вне слота) |
| `backend/testframework/search/repository_wave100_test.go` | 125 | 0 | чисто |
| `backend/testframework/search/service_query_test.go` | 72 | 0 | чисто — empty scope / outage |

Файлов в списке: **22**. Прочитано: **22**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A3W5-4-001 | REGRESS | P2 | `meilisearch_indexer.go:48–52` vs `85–112`; dial `safedialer.go:128–150` | `meiliAllowsOperatorPrivate` пропускает `127.0.0.1`/RFC1918 на validate (`meili_ssrf_task044_test.go:24`), но client = `NewSafeHTTPClient` → `SafeDialContext` **режет** loopback/private | До TASK-092 plain `http.Client` дозванивался до local/Docker Meili; после фикса `cmd/meilisearch-indexer` с `MEILI_HOST=http://127.0.0.1:7700` (или hostname→RFC1918) создаёт indexer и падает на Do. Комментарий в коде сам признаёт класс Ollama U01 | По аналогии с `NewOfflineAIHTTPClient`: operator Meili → dial allowlist (loopback и/или RFC1918 **только** после `meiliAllowsOperatorPrivate`); публичный host — SafeDial. Тест: FromEnv(`127.0.0.1:7700`) + DialContext loopback OK; metadata по-прежнему FAIL |
| A3W5-4-002 | NEW | P3 | `jql.go:108–133` vs parse `350–356`; тест `jql_test.go:100–117` | Парсер принимает `>`, `<`, `>=`, `<=` для любого поля; `conditionToPredicate` для `priority`/`type` (и status) их **молча игнорирует** (`return nil`) | `priority > medium` → пустой pred; `ToEntPredicates` может вернуть `nil` → вызывающий слой без фильтра (все issues) вместо ошибки | Либо поддержать ops, либо `ParseJQL`/ToEnt → error на unsupported field+op; тест mutation на silent-nil |
| A3W5-4-003 | NEW | P3 | `meilisearch_indexer.go:127–131` | `IndexAllIssues` `Limit(5000)` без пагинации/warn | Org >5000 issues → тихая усечённая индексация (ECC-SRCH-303 TODO) | Cursor/offset batches или slog.Warn + return truncated flag; тест 5001 |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| A2W5-A4-001 | TASK-092: delete/share → 403/404 не 500 |
| A2W5-A4-002 | TASK-092: `NewSafeHTTPClient` + metadata dial block (SSRF) |
| A2W5-A4-003 | `ErrSearchUnavailable` total outage |
| A2W5-A4-004 | `ErrFilterForbidden` sentinel + handler map |
| W5-A4-001…010, 012 | cascade table — NF-03/soft-delete/JQL OR/partial/q-runes/Meili validate |

## Чисто

- **NF-03:** issues/test_cases/projects/files — `projectIDs` + `deleted_at IS NULL`; channels — membership/public; messages — `getUserChannelIDs` + Mongo channel `$in` + `deleted_at` nil; users — shared-project scope (admin/founder unscoped через repo).
- **Auth:** `/search`, `/filters` через Group+authMw (Fiber v3).
- **JQL:** whitelist полей → Ent predicates (не raw SQL injection).
- **Layering:** Service без `*ent.Client`; Ent только в `EntRepository` / Meili IndexAllIssues (cmd indexer).
- **Meili SSRF metadata:** `169.254.169.254` / `metadata.google.internal` → nil indexer.
- **Тесты слота:** handler CRUD/authz edges; JQL parse+assignee A4-001; query empty-scope/outage; Meili upsert/env (public host).

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active → closed) | 4 (`A2W5-A4-001`…`004`) |
| FIXED-OK (Stage A cascade) | 10 |
| RESIDUAL | 0 |
| NEW | 2 (`A3W5-4-002`, `A3W5-4-003`) |
| REGRESS | 1 (`A3W5-4-001`) |

**Топ IDs:** `A3W5-4-001` (Meili operator SafeDial vs validate allowlist — P2 REGRESS).
