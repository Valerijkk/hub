# A2-W5 агент A4 — search prod + half tests (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc`  
> Зеркало: `W5-ASSIGNMENTS.md` §A4 · Stage A: `A5-4-search-core.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/search/dto.go` | 58 | 0 | чисто |
| `backend/internal/search/export_hooks.go` | 144 | 0 | чисто |
| `backend/internal/search/handler.go` | 265 | 2 | замечания |
| `backend/internal/search/indexer.go` | 67 | 0 | чисто |
| `backend/internal/search/jql.go` | 435 | 0 | чисто |
| `backend/internal/search/meilisearch_indexer.go` | 140 | 1 | замечания |
| `backend/internal/search/repository.go` | 123 | 0 | чисто |
| `backend/internal/search/service.go` | 767 | 1 | замечания |
| `backend/testframework/search/handler_flow_test.go` | 128 | 0 | чисто |
| `backend/testframework/search/handler_wave100_test.go` | 206 | 0 | чисто |
| `backend/testframework/search/handler_wave109_test.go` | 63 | 0 | чисто |
| `backend/testframework/search/handler_wave112_test.go` | 94 | 0 | чисто |
| `backend/testframework/search/handler_wave116_test.go` | 33 | 0 | чисто |
| `backend/testframework/search/handler_wave93_test.go` | 62 | 0 | чисто |
| `backend/testframework/search/jql_assignee_a4001_test.go` | 71 | 0 | чисто |
| `backend/testframework/search/jql_predicates_extended_test.go` | 94 | 0 | чисто |
| `backend/testframework/search/jql_test.go` | 171 | 0 | чисто |
| `backend/testframework/search/jql_wave100_test.go` | 58 | 0 | чисто |
| `backend/testframework/search/jql_wave116_test.go` | 32 | 0 | чисто |
| `backend/testframework/search/meilisearch_indexer_test.go` | 105 | 0 | чисто |
| `backend/testframework/search/repository_wave100_test.go` | 125 | 0 | чисто |
| `backend/testframework/search/service_query_test.go` | 70 | 0 | чисто |

Файлов в списке: 22. Прочитано: 22. Пропущено: 0.

## Находки

| ID | Sev | Tag | Файл:строка | Проблема | A1 / задача | Мутация |
|----|-----|-----|-------------|----------|-------------|---------|
| A2W5-A4-001 | P2 | RESIDUAL | `handler.go:204-206`; share аналогично | delete/share: ownership/not-found → **500** generic (не 403/404). `err.Error()` убран, **статус неверный** | W5-A4-006 · P2-BE-07 | Non-owner → 403; missing → 404 |
| A2W5-A4-002 | P2 | RESIDUAL | `meilisearch_indexer.go:47` | После `validateMeiliHost`: client = **`&http.Client{Timeout:30s}`** без SafeDial | W5-A4-009 partial · P2-BE-18 | Redirect/DNS rebinding mid-flight; `NewSafeHTTPClient` |
| A2W5-A4-003 | P3 | RESIDUAL | `service.go:178-180` | Все шарды failed → **partial+пусто**, не hard error (задокументированный follow-up) | W5-A4-007 residual | Total outage → error или explicit flag UX |
| A2W5-A4-004 | P3 | NEW | `handler.go:204-206` | Handler **не** `errors.Is(ErrFilterForbidden)` на delete/share — service возвращает plain `fmt.Errorf("доступ запрещён…")` | усиливает A2W5-A4-001 | Sentinel + map 403 |

## FIXED-OK

| A1 ID | Evidence |
|-------|----------|
| W5-A4-001/002 | `searchChannels`: `deleted_at IS NULL` + membership OR public |
| W5-A4-003 | Mongo `"deleted_at": nil` |
| W5-A4-004 | `getUserChannelIDs` JOIN channels + soft-delete |
| W5-A4-005 | `ToEntPredicates` AND/OR + `jql_or_task044_test.go` |
| W5-A4-007 | `Search` → `(results, partial, err)` + `partial_flag_task044_test.go` |
| W5-A4-008 | `utf8.RuneCountInString(q) > 500` → 400 |
| W5-A4-009 | `validateMeiliHost` + metadata/link-local block (client dial — residual) |
| W5-A4-010 | `truncateRunes` (TASK-060) |
| W5-A4-012 | OR union hubtest |

## Чисто

NF-03 issues/test_cases/projects/files/users; JQL whitelist (не SQL injection); auth на `/search`/`/filters`; Meili IndexAllIssues `DeletedAtIsNil`; layering без `*ent.Client` в service.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **22 / 22 / 0** |
| Находок | **4** (NEW 1 · REGRESS 0 · RESIDUAL 3) |
| P0 | 0 |
| P1 | 0 |
| P2 | 2 |
| P3 | 2 |
