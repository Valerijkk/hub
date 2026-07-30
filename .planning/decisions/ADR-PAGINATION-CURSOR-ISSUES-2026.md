# ADR-PAGINATION-CURSOR-ISSUES-2026 — List-issues: cursor, не offset

> **Status:** Accepted · **Date:** 2026-07-28 · **Deciders:** Owner (TASK-104 lock: апрув не требуется; канон = код)  
> **Findings:** A2W11-2-002, A2W11-5-002 · **Task:** `.planning/backlog/archive/2026-07-28/TASK-104-2026-07-28-docs-cursor-pagination-adr.md`

## Context

TZ (`docs/tz_constraints.md`) и rule `60-tz-constraints.mdc` утверждали: «задачи — offset/limit».  
Фактический API и клиент давно на **cursor**:

| Слой | Факт |
|------|------|
| BE DTO | `ListIssuesRequest.Cursor` + `Limit`; ответ `PaginatedResponse` с `next_cursor` / `has_more` |
| BE repo | `ListIssues` — курсор по `issue_key` (+ tie-break `order_by`) |
| FE | `ListIssuesParams.cursor`; wiki `API.md` уже «cursor pagination» |
| Chat | cursor (NFR-15) — без изменений |
| Intranet / wiki lists | offset/limit — **другие** эндпоинты, не list-issues |

Source precedence TASK-104: **код**. Мигрировать list-issues на offset запрещено (ломает FE infinite scroll / `issues-cursor.ts`). Chat cursor не трогать.

## Decision

1. **Канон list-issues:** `GET /projects/:projectId/issues/?cursor=&limit=` → `{ items, next_cursor, has_more, total }`.
2. **NFR-14** = cursor (+ limit, max 100) для задач. Offset/limit остаётся для admin/intranet/wiki списков (без привязки к NFR-14).
3. **NFR-15** (чат) без изменений.
4. Документы и rule-60 выравниваются под код; dual `PaginatedResponse` на FE — общий тип (cursor **или** offset по домену).

## Consequences

- Обновить: `docs/tz_constraints.md`, `.cursor/rules/60-tz-constraints.mdc`, `docs/system_specification.md` (NFR-14), wiki pagination note.
- Regression: тест курсорной пагинации list-issues (TASK-104).
- Новые list-API задач не вводить через `offset`/`page` без нового ADR.

## Compliance

Код list-issues не меняется (уже cursor). Messenger cursor не затрагивается.
