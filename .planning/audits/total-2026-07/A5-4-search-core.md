# W5 агент A4 — search prod + half tests

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: **NF-03** user-scope filter; soft-delete exclude; JQL injection.  
> Сверка: TZ `docs/tz_constraints.md` / NFR-19; prior `grok-2026-07/04-ai-search-storage.md` (A4-001…004 FIXED в коде; A4-015/016/017/018/020 — residual).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/search/dto.go` | 56 | 0 | чисто |
| `backend/internal/search/export_hooks.go` | 133 | 0 | чисто |
| `backend/internal/search/handler.go` | 253 | 3 | замечания |
| `backend/internal/search/indexer.go` | 67 | 0 | чисто |
| `backend/internal/search/jql.go` | 406 | 1 | замечания |
| `backend/internal/search/meilisearch_indexer.go` | 103 | 1 | замечания |
| `backend/internal/search/repository.go` | 123 | 0 | чисто |
| `backend/internal/search/service.go` | 741 | 5 | замечания |
| `backend/testframework/search/handler_flow_test.go` | 128 | 0 | чисто |
| `backend/testframework/search/handler_wave100_test.go` | 200 | 0 | чисто |
| `backend/testframework/search/handler_wave109_test.go` | 63 | 1 | замечания |
| `backend/testframework/search/handler_wave112_test.go` | 94 | 0 | чисто |
| `backend/testframework/search/handler_wave116_test.go` | 33 | 0 | чисто |
| `backend/testframework/search/handler_wave93_test.go` | 62 | 0 | чисто |
| `backend/testframework/search/jql_assignee_a4001_test.go` | 71 | 0 | чисто |
| `backend/testframework/search/jql_predicates_extended_test.go` | 94 | 0 | чисто |
| `backend/testframework/search/jql_test.go` | 171 | 1 | замечания |
| `backend/testframework/search/jql_wave100_test.go` | 58 | 0 | чисто |
| `backend/testframework/search/jql_wave116_test.go` | 32 | 0 | чисто |
| `backend/testframework/search/meilisearch_indexer_test.go` | 105 | 0 | чисто |
| `backend/testframework/search/repository_wave100_test.go` | 125 | 0 | чисто |
| `backend/testframework/search/service_query_test.go` | 70 | 0 | чисто |

Файлов в списке: 22. Прочитано: 22. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W5-A4-001 | P1 | `service.go:519-534` | **`searchChannels` без `deleted_at IS NULL`** (оба пути: tsvector и ILIKE). После A8-005 у Channel есть soft-delete | Soft-deleted каналы остаются в Ctrl+K (`type=channel`); A4-002 закрыл issues/test_cases/projects/files, **channels пропущены** | `AND deleted_at IS NULL` в обоих SQL; runtime/source regression как A4-002 |
| W5-A4-002 | P1 | `service.go:522-523`; `service.go:531-532` | Шард каналов фильтрует только `project_id = ANY($2)`, **без membership в канале** и без различия `public`/`private` | NFR-19 / NF-03: поиск обязан scope’ить по **channel membership**. Любой участник проекта видит имена **private** каналов, в которых не состоит (discovery IDOR) | Для `private`/`group`: `id IN (SELECT channel_id FROM channel_members WHERE user_id=$caller)`; `public` — project scope OK; DM с `project_id IS NULL` сюда не попадают (OK) |
| W5-A4-003 | P1 | `service.go:587-590` | Mongo `$text` filter **без** `deleted_at: nil` | Messenger soft-delete ставит `deleted_at` (`mongo_repository.go` SoftDeleteMessage); Ctrl+K отдаёт удалённые сообщения | Добавить `"deleted_at": nil` (как List в messenger); тест на soft-deleted message → 0 hits |
| W5-A4-004 | P2 | `service.go:182-184` | `getUserChannelIDs` читает `channel_members` **без JOIN** на `channels.deleted_at IS NULL` | После soft-delete канала membership-строки остаются (A8-005: children keep) → `searchMessages` продолжает индексировать корзину | JOIN/`EXISTS` active channel; либо чистить members при delete |
| W5-A4-005 | P2 | `jql.go:47-57` | `ToEntPredicates` **игнорирует** `cond.Logic` (AND/OR) — все preds через последовательный `Where` (= AND). Residual prior **A4-016** | `priority = high OR type = bug` семантически = AND → ложные пустые выдачи JQL API (`tracker/handler_issues.go` потребляет ToEntPredicates) | Собирать дерево `issue.And`/`issue.Or` по `Logic`; mutation: OR-запрос должен вернуть union |
| W5-A4-006 | P2 | `handler.go:194-195`; `handler.go:231-233` | `deleteFilter` / `shareFilter`: ownership deny и not-found → **500** + `err.Error()`. Residual **A4-017** | Неверный HTTP; клиент видит внутренний текст («доступ запрещён» / «не найден») | Map: not found → 404; not owner → 403; sentinel `ErrFilterForbidden` / `ErrNotFound` |
| W5-A4-007 | P2 | `service.go:161-173` | Ошибки шардов только `slog.Warn`; `Search` всегда `return results, nil`. Residual **A4-015** | Частичный fail выглядит как «пустой поиск» — ложный UX / скрытый outage | Если все шарды failed → error; иначе `partial` flag / Warning header |
| W5-A4-008 | P2 | `handler.go:63-66` | `q` проверяется только на `== ""`; **нет** `max=500` из DTO `validate` | Гигантский `q` → нагрузка на `plainto_tsquery` / Mongo `$text` (DoS-класс) | `utf8.RuneCountInString(q) > 500` → 400; либо validator |
| W5-A4-009 | P2 | `meilisearch_indexer.go:33-36` | Meili client = `&http.Client{Timeout:30s}` без `ValidateOutboundURL` / SafeDial. Residual **A4-018** | `MEILI_HOST` из env без SSRF-паттерна TZ (риск при компромете env / SSRF-to-internal) | `ValidateOutboundURL(host)` + `NewSafeHTTPClient`; allowlist internal CIDR если нужен |
| W5-A4-010 | P3 | `service.go:621-622` | `title = title[:100]` по **байтам**. Residual **A4-020** | Mid-rune truncate UTF-8 в snippet сообщений | Truncate по рунам |
| W5-A4-011 | P3 | `handler.go:126`; `handler.go:166` | `createFilter` / `listFilters` при non-forbidden ошибке отдают `err.Error()` в 500 body | Утечка деталей БД/обёрток наружу (create/list membership-path уже 403 — OK) | Стабильное `apperr` message; детали только в slog |
| W5-A4-012 | P3 | `jql_test.go:143-148`; `jql_wave100_test.go:25-30` | Тесты OR только проверяют **парсинг** `Logic`, не поведение `ToEntPredicates` | Мутация «игнор Logic» (W5-A4-005) **не краснеет** — guard без ловца | Integration: OR JQL → count/keys union; AND control |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **NF-03 issues/test_cases/projects/files** | `getUserProjectIDs` → `project_* = ANY($2)`; empty projectIDs → early `nil` (fail-closed); `plainto_tsquery($1)` parameterized — SQL injection по `q` нет |
| **NF-03 users** | common-project JOIN + self; admin unscoped через `IsSystemAdmin` (fail-closed on repo err) |
| **NF-03 messages (membership)** | `channel_id $in` из `channel_members` (SEC-H2) — scope по членству есть; дыра только soft-delete (W5-A4-003/004) |
| **Soft-delete A4-002 (PG core)** | `issues`/`test_cases`/`projects`/`file_metadata`: `deleted_at IS NULL`; files + `status = 'confirmed'` — **закрыто** |
| **Saved filters authz A4-003/004** | `requireProjectMembership` → `ErrFilterForbidden`; handler create/list → 403 |
| **JQL injection** | Нет raw SQL: whitelist полей в `conditionToPredicate`; значения → Ent predicates / UUID parse для assignee; unknown field → nil pred (не инъекция). `ORDER BY` field whitelist на стороне tracker consumer |
| **JQL assignee A4-001** | `HasAssigneeWith(user.IDEQ)` + runtime `jql_assignee_a4001_test.go` |
| **Auth routes** | Group + `authMw` на `/search` и `/filters`; `getUserID` string/UUID/unknown |
| **Meili soft-delete index** | `IndexAllIssues`: `issue.DeletedAtIsNil()`; body `LimitReader` 4KiB на error path |
| **Indexer** | Идемпотентные `ADD COLUMN IF NOT EXISTS` / GIN; context Exec |
| **Repository layering** | SearchService без `*ent.Client`; Ent только в `EntRepository` / Meili indexer |
| **dto / export_hooks** | DTO контракты; hubtest re-exports без prod-логики |
| **handler_flow / wave93/100/112/116** | Missing q/401/CRUD/limit clamp/type filter/UUID locals — asserts есть |
| **jql_* tests** | Parse/IN/dates/unknown field/assignee mutation A4-001 |
| **service_query_test** | Empty membership → empty; type filter; closed DB → error |
| **meilisearch_indexer_test** | env unset/configured; HTTP upsert; error status; ent error |
| **repository_wave100** | IsSystemAdmin founder/not-found; shared filters list |

**JQL SQL/injection периметр:** безопасен (Ent + field whitelist). Главный дефект JQL — семантика OR (W5-A4-005), не инъекция.

**Residual вне списка A4 (A5 tests):** `soft_delete_a4002_*.go` / `service_scope_test.go` — не читались; статическая A4-002 регрессия не покрывает `channels`/`messages` (согласуется с W5-A4-001/003).

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **22 / 22 / 0** |
| Находок | **12** |
| P0 | 0 |
| P1 | 3 |
| P2 | 6 |
| P3 | 3 |

### Топ-3

1. **W5-A4-001 (P1)** — soft-deleted channels в Ctrl+K (нет `deleted_at IS NULL`).  
2. **W5-A4-002 (P1)** — NF-03/NFR-19: private channel discovery через project-scope без channel membership.  
3. **W5-A4-003 (P1)** — soft-deleted Mongo messages в поиске (нет `deleted_at: nil`).
