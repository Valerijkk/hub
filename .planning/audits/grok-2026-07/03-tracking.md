# TMS / Tracker / Project / Sprint

**Дата:** 2026-07-26  
**Зона:** A3 — `backend/internal/{tms,tracker,project,sprint}`  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Ветка:** `cursor-grok-4.5`  
**Объём:** ~94 `.go` файла; просмотрены handler/service/repository, automation, poker, time, optimistic lock, IDOR/authz, pagination, tx, goroutines.

## Что проверено

| Модуль | Фокус |
|--------|--------|
| `tracker/` | CRUD issues, MoveIssue, optimistic lock (`VersionEQ`), comments/time, JQL, automation consumer/engine, links, boards/workflow |
| `tms/` | TestCase CRUD/version/rollback, runs/results, Allure import/matcher, CreateBugIssue, reports (flaky N+1), environments |
| `project/` | Create/Get/Update, invite join, AddTeamToProject, provision |
| `sprint/` | Start/Complete sprint scoping, WIP, poker Redis, burndown |
| Паттерны | `_ = err`, nil deref, IDOR (URL project vs resource), tx rollback, context/goroutine, pagination HasMore |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A3-001 | P0 | `sprint/service.go:107-149` (+ `repository.go:78-83`, `115-119`) | `CompleteSprint` / `StartSprint` / `GetSprint` / `UpdateSprint` не сверяют `sprint.project_id` с `projectID` из URL | RBAC проверяет только membership в `:projectId`. Зная UUID спринта чужого проекта, можно завершить его и **перенести задачи в спринт вызывающего проекта** (`FindOrCreateNextPlanningSprint(ctx, projectID, sp)`), либо активировать чужой спринт | Везде: `GetSprint` с `Where(ID, HasProjectWith(projectID))`; при mismatch → 404. Complete/Start только после проверки ownership |
| A3-002 | P1 | `tracker/handler_issues.go:100-105`, `128-143`; `service_issues.go:123-124`, `191`, `288`, `300` | CRUD/move issue идут по `issueKey` без сверки с `:projectId` | Member проекта A с ключом `B-42` читает/правит задачу проекта B (IDOR). Тот же паттерн у comments/time/links через `FindIssueByKey` | `FindIssueByKey(ctx, projectID, key)` или post-check `issue.Edges.Project.ID == projectID` → 404 |
| A3-003 | P1 | `tms/handler_test_cases.go:94-128`; `service_test_cases.go:65-66`, `117-118`, `186-194` | Test case GET/PUT/DELETE/rollback/links по `caseKey` без `:projectId` | Аналог A3-002 для TMS: доступ к чужим кейсам при знании ключа | Scope `FindTestCaseByKey` по `projectID` из URL |
| A3-004 | P1 | `tms/ent_repository.go:247-303`; `service_test_cases.go:148`; `dto.go:52` | `UpdateTestCaseRequest.Version` required, но repo делает `UpdateOneID` + `AddVersion(1)` **без** `VersionEQ` | Optimistic lock из TZ не работает: last-write-wins, тихая потеря правок; `req.Version` мёртв | Как в tracker: `Update().Where(ID, VersionEQ(input.Version)).AddVersion(1)`; `affected==0` → 409 |
| A3-005 | P1 | `tracker/repo_time.go:22-45`, `66-83` | `LogTime` создаёт запись с `ended_at=nil` и без `started_at`; `StopTimer` ищет `EndedAtIsNil` и делает `*openTimer.StartedAt` | Ручной лог + StopTimer → **nil panic**. Несколько open-записей → `Only` error / неоднозначность | Фильтр StopTimer: `StartedAtNotNil()`; LogTime сразу `SetEndedAt(now)`; StartTimer отклонять при уже открытом таймере |
| A3-006 | P1 | `tracker/automation_consumer.go:84-109` | После unmarshal/engine ошибок всегда `_ = msg.Ack()` | Сломанные/необработанные события **теряются** навсегда (нет Nak/retry). MaxDeliver бесполезен при Ack | Ack только после успешной обработки; иначе `Nak` / `NakWithDelay`; poison → DLQ + Ack |
| A3-007 | P1 | `tms/service_test_cases.go:207-224`; `ent_repository.go:349-354` | Rollback: `GetTestCaseVersion(versionID)` без проверки, что версия принадлежит `existing` | Можно применить snapshot чужого кейса к текущему (data corruption / IDOR по UUID версии) | `Where(ID, HasTestCaseWith(IDEQ(existing.ID)))`; mismatch → 404 |
| A3-008 | P1 | `tms/service_runs.go:214-235`; `service_test_cases.go:313-342` | `ExecuteTestResult` обновляет `resultID` без привязки к `runID`; `CreateBugFromFailedResult` не проверяет, что result ∈ run/project URL | Можно менять результат другого прогона / создавать баг в своём проекте по чужому `resultID` | `Where(result.ID, HasRunWith(IDEQ(runID), HasProject…))`; bug — после verify run+project |
| A3-009 | P1 | `tms/service_test_cases.go:324-342`; `ent_repository.go:834-840` | При отсутствии `Edges.TestCase` `testCaseID=uuid.Nil`, затем всегда `UpdateOneID(testCaseID).AddLinkedIssueIDs` | Баг из unmapped failed result всегда fail/rollback; UUID nil недопустим | Если `testCaseID==uuid.Nil` — создать issue без link; иначе link только после проверки TC ∈ project |
| A3-010 | P1 | `tms/allure_matcher.go:139` | Match по case_key: `FindTestCaseByKey` **без** `projectID` | Allure import может привязать результат к кейсу **другого** проекта | После find: `tc.ProjectID == projectID` или scoped query |
| A3-011 | P1 | `tracker/handler_comments.go:92-131`; `service_comments.go:74-97` | Update/Delete comment только по `commentID`; нет автора, нет привязки к `issueKey` | Любой с `PermTaskComment` в проекте A правит/удаляет чужой комментарий (в т.ч. другого issue) по UUID | Load comment → verify Author + IssueKey/project; иначе 403/404 |
| A3-012 | P1 | `tms/service_import.go:140-148` | `ImportAllureToExistingRun` грузит run по `runID`, не сверяет project | IDOR: импорт Allure в чужой run под RBAC своего `:projectId` | `run` + `HasProjectWith(projectID)` или compare edges |
| A3-013 | P2 | `tracker/repo_issues.go:373-387`; `service_issues.go:329` | `MoveIssue` без `version` / optimistic lock | Гонка двух drag-and-drop: last write wins, WIP check TOCTOU | `VersionEQ` + AddVersion; WIP в той же tx |
| A3-014 | P2 | `tms/service_test_cases.go:126-177` | Snapshot `CreateTestCaseVersion` до `UpdateTestCase`; не одна tx; lock не проверяется | При ошибке update — orphan version; при гонке — дубликаты снимков и lost update | Одна tx: version snapshot + `VersionEQ` update; rollback всей tx при 409 |
| A3-015 | P2 | `sprint/service.go:257-270` | `CheckWIPLimit`: статус не найден / ошибка count → `return nil` | WIP fail-open: лимит обходится при сбоях БД | На ошибке query — fail closed (500) или явный slog+метрика; не silent skip |
| A3-016 | P2 | `sprint/service.go:88-95` | `StartSprint`: конфликт только если `err == nil`; иные ошибки `FindActiveSprint` игнорируются | При DB error можно запустить второй active sprint | `if err != nil && !ent.IsNotFound(err) { return Internal }`; иначе conflict-check |
| A3-017 | P2 | `tms/service_runs.go:214-235` | `userID` аргумент `ExecuteTestResult` не пишется в `ExecutedByID` | Аудит «кто прогнал» пустой | `input.ExecutedByID = &userID` в UpdateTestResult |
| A3-018 | P2 | `tracker/service_issues.go:169-182` | `HasMore: len(items) == limit` | На последней странице ровно `limit` элементов → ложный `HasMore` (off-by-one) | Fetch `limit+1` или сравнивать с `total`/курсором |
| A3-019 | P2 | `sprint/poker.go:82-88`, `108-109` | Votes key без TTL до первого голоса; Reveal без атомарного CAS `revealed` | Голоса могут жить без expire; race vote+reveal | `SETNX`/`HSET NX` revealed; TTL votes при StartSession; Lua/tx |
| A3-020 | P2 | `tms/service_reports.go:256-257` | Flaky: `ListTestResults` на каждый run (N+1) | Лишняя нагрузка на PG при отчётах | Batch join / `Where(HasRunWith(IDIn(...)))` одним запросом |
| A3-021 | P3 | `tracker/repo_issues.go:344-352`, `400`; `repo_comments.go:46`; `repo_links.go:57` | `_ = r.publisher.Publish(...)` глотает ошибки outbox | Клиенты/automation не получают события без сигнала | Log+metric; критичные — fail tx/`PublishInTx` |
| A3-022 | P3 | `tracker/handler_issues.go:269` | JQL handler дергает `h.service.repo.GetDB()` | Нарушение layering; handler знает Ent | Метод service `SearchJQL(ctx, projectID, …)` |
| A3-023 | P3 | `project/repository.go:579` | `_, _ = r.AddMember(...)` в `AddTeamToProject` | Не-constraint ошибки (DB) тихо теряются → частичный membership | Игнорировать только constraint/duplicate; остальное return err |

## Чисто

- Tracker `UpdateIssue` использует `VersionEQ` + `AddVersion(1)` и 409 на conflict (`repo_issues.go:284-334`) — образец для TMS.
- CreateIssue / CreateTestCase / CreateBugIssue: tx + Rollback на ошибках.
- Allure/JUnit/TestRail/template import: `io.LimitReader` (не голый `ReadAll`).
- Project `ListProjects`: `MemberCountBatch` (N+1 закрыт).
- Fiber v3 RBAC порядок в `tracker/handler.go` / `tms/handler.go` задокументирован (FA-02-02).
- Invite join: атомарный consume (HI-06) в `project/invite.go`.
- Webhook TMS под `PermTestRunCreate` (комментарий FA-02-02 в `tms/handler.go:109-117`).
- Sprint date validation `EndDate.Before(StartDate)` на create.

## Оценка зоны: 5.5/10

**Обоснование:** Базовый CRUD, Ent-транзакции создания, LimitReader и tracker optimistic lock выглядят зрело. Но системный класс багов — **resource lookup без project scope** (sprint P0 data-move, issue/TC/run/result/comment P1 IDOR) — снижает оценку сильно. Плюс сломанный optimistic lock в TMS, panic-путь StopTimer, Ack-on-error в automation. После фикса scoping + VersionEQ + timer/automation зона могла бы быть ~8/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 1 |
| P1 | 11 |
| P2 | 8 |
| P3 | 3 |
| **Total** | **23** |

**Score: 5.5/10**
