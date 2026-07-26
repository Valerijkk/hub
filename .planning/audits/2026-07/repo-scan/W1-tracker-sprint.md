# W1 — tracker / sprint (repo-scan)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> Дата: 2026-07-23 · Агент: be-ver (Grok 4.5) · Режим: readonly · Scope: `backend/internal/tracker/`, `backend/internal/sprint/`
> Фокус: layering (`*ent.Client`), optimistic lock `version`, error format, pagination
> Код не менялся. Регрессия не запускалась (scan-only).

## Layering verdict

| Модуль | Вердикт | Доказательство |
|--------|---------|----------------|
| **tracker** | **FAIL** | `handler_issues.go:269` — handler вызывает `h.service.repo.GetDB()` → `*ent.Client` и строит Ent-query в HTTP-слое. Интерфейс `Repository.GetDB() *ent.Client` (`repository.go:175`) легализует утечку клиента. В `TrackerService` поля `*ent.Client` нет — ок. |
| **sprint** | **PASS** | `*ent.Client` только в `EntRepository` (`repository.go:57`). `SprintService` / `Handler` / `PokerService` клиента не держат и не вызывают Ent напрямую. |

**Итог layering: FAIL (из-за tracker JQL).**

## Findings

| Sev | Файл:строка | Evidence | Fix |
|-----|-------------|----------|-----|
| **HIGH** | `tracker/handler_issues.go:269-295` | `SearchIssuesJQL`: `h.service.repo.GetDB().Issue.Query()...All(c.Context())` — handler обходит service, держит Ent-query, нарушает Handler→Service→Repository. | Вынести JQL в `TrackerService`/`Repository` (метод `SearchByJQL`); убрать `GetDB()` из публичного `Repository` (или оставить только для EntRepository, не в interface). |
| **HIGH** | `tracker/repository.go:174-175`, `:277` | `GetDB() *ent.Client` в интерфейсе Repository — канон layering запрещает `*ent.Client` вне EntRepository. | Удалить `GetDB` из interface; JQL-предикаты применять внутри `EntRepository`. |
| **HIGH** | `tracker/repo_issues.go:373-387` | `MoveIssue` делает `UpdateOneID(issueID).SetStatusID/SetRank` **без** `VersionEQ` / `AddVersion(1)`. Обходит optimistic lock `UpdateIssue` — lost update статуса/ранга при конкурентном PUT + drag-and-drop. | Принимать `version` в `MoveIssueRequest`; `Where(issue.VersionEQ(...)).AddVersion(1)`; 409 при `affected==0`. |
| **HIGH** | `ent/schema/sprint.go:17-43` + `sprint/repository.go:97-119` | В схеме Sprint **нет** `version`; `UpdateSprint` / `UpdateSprintStatus` / `UpdateRetroNotes` — `UpdateOneID` без optimistic lock. TZ: `version` в изменяемых Ent-схемах. | Добавить `field.Int("version").Default(1)`; миграция; на Update — `VersionEQ` + `AddVersion(1)` + conflict. |
| **HIGH** | `tracker/dto.go:55-65`, `repo_issues.go:211-255` | TZ (`tz_constraints`): задачи — **offset/limit**; факт — **cursor + limit** (`ListIssuesRequest.Cursor`, курсор = `issue_key`). Offset не поддерживается. | Либо привести API к offset/limit (+ `total` уже есть), либо ADR/правка TZ на cursor для issues (сейчас drift). |
| **HIGH** | `tracker/handler_issues.go:99-110` + `service_issues.go:123-131` | `GetIssue`/`UpdateIssue`/`DeleteIssue`/`MoveIssue` **не** сверяют `projectId` из URL с проектом задачи (в отличие от message-links: `service_issue_message_link.go:132-134`). RBAC проверяет только URL-project → IDOR между проектами при известном `issueKey`. | Прокидывать `projectID` в service; при mismatch → 404. |
| **HIGH** | `sprint/handler.go:124-136`, `:153-169`, `:249-265` + `service.go:47-56` | `GetSprint`/`UpdateSprint`/`UpdateRetroNotes`/`GetBurndown` принимают только `sprintId`, **не** проверяют принадлежность к `projectId` из URL. Аналогичный IDOR. | Передавать `projectID`; в repo/service фильтр `HasProjectWith` / сравнение edge; mismatch → 404. |
| **MEDIUM** | `platform/errors/errors.go:7-12` + `middleware/middleware.go:24` (потребители tracker/sprint) | Ответ ошибок: `{"code","message","request_id"}`. TZ: `{"error","request_id"}`. Tracker/sprint корректно возвращают `*AppError`/`fiber.Error` в ErrorHandler, но **формат поля** ≠ TZ. | Платформенный фикс ErrorHandler/AppError JSON-тега `error` (вне scope модулей, но блокирует TZ-соответствие API). |
| **MEDIUM** | `tracker/handler_issues.go:291-340` | JQL: хардкод `Limit(100)`; `"total": len(result)` = размер страницы, не count; нет offset/cursor. | Пагинация + реальный `Count`; метод в service/repo. |
| **MEDIUM** | `sprint/repository.go:87-93` + `service.go:59-70` | `ListSprints` без пагинации, `WithIssues()` на все спринты — риск раздувания ответа. | Offset/limit (или cursor) + опционально не грузить issues в list. |
| **MEDIUM** | `sprint/service.go:257-270` | `CheckWIPLimit`: при ошибке Get/Count — `return nil` (fail-open), WIP не enforced. | Fail-closed или явный wrap ошибки; не глотать infra-err. |
| **MEDIUM** | `tracker/service_issue_message_link.go:111-114` | `return err` / `fmt.Errorf("issue not in project")` — не `*AppError`, уйдёт в generic 500 с `"internal server error"`. | `apperrors.NewNotFound` / `NewBadRequest`. |
| **MEDIUM** | `tracker/repo_time.go:38-40`, `:98-100` | `AddTimeSpentMinutes` через `UpdateOneID` без version — гонка с `UpdateIssue` (не атомарно с version-bump). | Обновлять spent в той же политике version/Tx или отдельный atomic SQL `ADD`. |
| **LOW** | `tracker` lists: comments/activity/time-logs | `ListComments`/`ListActivities`/`ListTimeLogs` — без пагинации (обычно малые коллекции). | Cursor/limit при росте объёма. |
| **LOW** | `tracker/service_*.go` (много) | `NewInternal("...")` глотает underlying err (нет `%w`) — хуже диагностика; на API-формат не влияет. | Логировать `err` + wrap внутри, наружу — AppError. |

## Чек-лист (кратко)

| Проверка | tracker | sprint |
|----------|---------|--------|
| `*ent.Client` не в service | OK | OK |
| `*ent.Client` не в handler | **FAIL** (JQL) | OK |
| Optimistic lock Issue Update | OK (`repo_issues.go:284-333`) | n/a |
| Optimistic lock Move / Sprint Update | **FAIL** MoveIssue | **FAIL** нет version в schema |
| Error → центральный ErrorHandler | OK (через return err) | OK |
| JSON error shape TZ `error` | **FAIL** (platform `message`) | **FAIL** (то же) |
| Пагинация задач TZ offset/limit | **FAIL** (cursor) | n/a (list sprints без пагинации) |

## Вердикт

```
## Верификация (BE): tracker + sprint (repo-scan W1)
- Воспроизведено: статический разбор Handler→Service→Repository + schema version + List*/JQL
- Layering: FAIL — tracker/handler_issues.go:269 GetDB(); sprint OK
- Security-инварианты: FAIL — IDOR projectId vs issueKey/sprintId (HIGH)
- Optimistic lock: PARTIAL — UpdateIssue OK; MoveIssue + Sprint без version
- Error format: PARTIAL — AppError path OK; поле message ≠ TZ error
- Pagination: FAIL vs TZ offset/limit для issues; JQL/sprints без нормальной пагинации
- Регрессия: не запускалась (readonly scan)
- Вердикт: verified: false — 2026-07-23 00:56 UTC+3
```

## Verification

- Воспроизведено по коду (не по PR-описанию): GetDB/JQL, MoveIssue без version, sprint schema, ListIssues cursor, IDOR project scope.
- CRITICAL: 0 · HIGH: 7 · MEDIUM: 6 · LOW: 2
- Регрессия: N/A (scan-only)
- Next: be1/be2 — фикс layering JQL + version на Move/Sprint + project-scope checks; platform — error JSON; ADR или смена пагинации issues.
