# W1 агент A4 — sprint

> Дата: 2026-07-28 · TASK-008 · READ-ONLY по коду · только этот отчёт  
> Перепроверка A3 (`grok-2026-07/03-tracking.md`): A3-001 / A3-015 / A3-016 / A3-019

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/sprint/dto.go` | 86 | 0 | чисто |
| `backend/internal/sprint/export_hooks.go` | 19 | 0 | чисто |
| `backend/internal/sprint/handler.go` | 515 | 2 | замечания |
| `backend/internal/sprint/poker.go` | 168 | 4 | замечания |
| `backend/internal/sprint/repository.go` | 376 | 2 | замечания |
| `backend/internal/sprint/service.go` | 362 | 4 | замечания |
| `backend/testframework/sprint/fail_closed_a3015_test.go` | 46 | 0 | чисто |
| `backend/testframework/sprint/fail_closed_a3016_test.go` | 51 | 0 | чисто |
| `backend/testframework/sprint/handler_flow_test.go` | 213 | 0 | чисто |
| `backend/testframework/sprint/handler_rbac_test.go` | 145 | 1 | замечания |
| `backend/testframework/sprint/handler_routes_test.go` | 204 | 0 | чисто |
| `backend/testframework/sprint/handler_wave113_test.go` | 106 | 1 | замечания |
| `backend/testframework/sprint/list_batch_counts_a13006_test.go` | 154 | 0 | чисто |
| `backend/testframework/sprint/repository_ent_test.go` | 231 | 0 | чисто |
| `backend/testframework/sprint/repository_get_in_project_a3001_test.go` | 98 | 0 | чисто |
| `backend/testframework/sprint/repository_list_no_issues_a13006_test.go` | 63 | 0 | чисто |
| `backend/testframework/sprint/service_extended_test.go` | 206 | 1 | замечания |
| `backend/testframework/sprint/service_test.go` | 657 | 0 | чисто |
| `backend/testframework/sprint/service_wave114_test.go` | 18 | 0 | чисто |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Перепроверка A3 (sprint)

| ID | Было | Сейчас | Вердикт |
|----|------|--------|---------|
| **A3-001** P0 | `GetSprint`/`Start`/`Complete`/`Update` без сверки `sprint.project_id` ↔ URL `projectID` | Все мутации/чтения через `GetSprintInProject` (`HasProjectWith`); Ent-тест `repository_get_in_project_a3001_test.go`; service IDOR-сьюты в `service_test.go` | **ИСПРАВЛЕНО** |
| **A3-015** P2 | `CheckWIPLimit` fail-open (`return nil` на err) | NotFound / Internal на ошибки status/count; тест `fail_closed_a3015_test.go` | **ИСПРАВЛЕНО** |
| **A3-016** P2 | `StartSprint` игнорировал DB error `FindActiveSprint` | `err != nil && !ent.IsNotFound` → Internal; тест `fail_closed_a3016_test.go` | **ИСПРАВЛЕНО** |
| **A3-019** P2 | Votes без TTL до первого голоса; Reveal без CAS `revealed` | Session key: pipeline HSet+Expire OK. Votes: Expire после HSet (частично). Reveal: обычный `HSet` без CAS — **живёт** | **ЧАСТИЧНО / ЖИВО** (CAS) |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W1-A4-001 | P1 | `backend/internal/sprint/service.go:156-181` | `CompleteSprint`: `FindOrCreateNextPlanningSprint` + `MoveIssuesToSprint` + `UpdateSprintStatus` **не в одной tx** | При ошибке `UpdateSprintStatus` после успешного move задачи уже в следующем спринте, текущий остаётся `active` — рассинхрон данных / повторный complete | Одна Ent-tx: create-next (если нужно) + move + status=completed; rollback целиком |
| W1-A4-002 | P2 | `backend/internal/sprint/poker.go:108-109` | A3-019: `RevealVotes` делает `HSet(revealed=true)` без атомарного CAS | Гонка vote↔reveal: голос проходит check `revealed!=true`, затем reveal, затем HSet голоса — голос после reveal или потеря порядка | Lua / `HSETNX`/`WATCH`+MULTI: CAS `revealed` false→true; CastVote только если CAS-поле false |
| W1-A4-003 | P2 | `backend/internal/sprint/poker.go:82-88` | A3-019 residual: `votes` key создаётся только на первом голосе; `HSet` и `Expire` не атомарны; ошибки `Expire` игнорируются | Crash/сеть между HSet и Expire → голоса без TTL; ключ не инициализируется при Start | В `StartPokerSession` pipeline: пустой votes hash + Expire; в CastVote — один pipeline HSet+Expire обоих ключей, проверять err |
| W1-A4-004 | P2 | `backend/internal/sprint/poker.go:77` | `revealed, _ := p.redis.HGet(...)` — ошибка Redis глотается | Fail-open: при ошибке HGet голос принимается как будто сессия не раскрыта | Проверять err HGet → Internal; empty/redis.Nil отдельно от true |
| W1-A4-005 | P2 | `backend/internal/sprint/handler.go:43-47` | Poker `vote`/`reveal`/`GET` только под `authMw`, без project RBAC и без membership по `sprint_id` сессии | Любой аутентифицированный с `sessionId` (из NATS `poker.*` / утечки) голосует/раскрывает без `PermProjectView` | Вложить poker под `/projects/:projectId/...` или читать `sprint_id` из Redis и `RequirePermission` + membership |
| W1-A4-006 | P2 | `backend/internal/sprint/handler.go:418-423` | `StartPokerSession`: после scope спринта `issue_id` не проверяется на принадлежность проекту/спринту | Poker-сессия с чужим/несуществующим `issue_id` (метаданные Redis + NATS) | `FindIssue` с `HasProjectWith(projectID)` (и опционально sprint); mismatch → 400/404 |
| W1-A4-007 | P2 | `backend/internal/sprint/service.go:85-101` | `UpdateSprint` не валидирует `EndDate.Before(StartDate)` (в отличие от Create) | Через PUT можно выставить end &lt; start → сломанный burndown (`days` clamp) и некорректные даты | После merge полей: та же проверка, что в `CreateSprint` |
| W1-A4-008 | P2 | `backend/internal/sprint/repository.go:125-139` (+ `:143-147`, `:151-155`) | `UpdateSprint` / `UpdateSprintStatus` / `UpdateRetroNotes` — `UpdateOneID` без optimistic lock; в схеме Sprint нет `version` (TZ) | Concurrent PUT/retro/start → last-write-wins, тихая потеря правок | Добавить `version` в schema (W8) + `VersionEQ`/`AddVersion`; 409 на conflict |
| W1-A4-009 | P2 | `backend/internal/sprint/service.go:116-124` | `StartSprint`: check `FindActiveSprint` и `UpdateSprintStatus` не атомарны | Два параллельных Start разных planning-спринтов → оба active в одном проекте | Unique partial index `(project_id) WHERE status=active` + обработка unique violation, либо tx+`SELECT … FOR UPDATE` |
| W1-A4-010 | P3 | `backend/internal/sprint/repository.go:334-345` | `GetVelocityData`: `committed, _ :=` / `completed, _ :=` — ошибки Count глотаются | При сбое Count velocity отдаёт 0/0 как «факт» | Пропагировать err; не маскировать нулями |
| W1-A4-011 | P3 | `backend/internal/sprint/poker.go:109` | `RevealVotes`: результат `HSet` не проверяется | Клиент получает Revealed=true в JSON, Redis мог не обновиться | `if err := …HSet….Err(); err != nil { return Internal }` |
| W1-A4-012 | P3 | `backend/internal/sprint/service.go:73-75` | `ListSprints`: ошибка `GetSprintIssueCounts` → пустая map (нули) | Список 200 OK с ложным `issue_count=0` при DB error | Fail closed (500) или slog+metric; не silent zero без сигнала |
| W1-A4-013 | P3 | `backend/internal/sprint/service.go:299-301` | `RemoveIssuesFromSprint` чистит sprint по UUID задач без `projectID` | HTTP-маршрута нет, но API опасен при будущем expose / внутреннем вызове | Скопировать паттерн `SetIssuesSprint(projectID, …)` + Clear scoped |
| W1-A4-014 | P3 | `backend/testframework/sprint/handler_rbac_test.go:121-144` | RBAC-сьют покрывает create/update/start outsider; **poker vote/reveal** не проверяются | Регрессия «добавили RBAC на poker» или наоборот не ловится; дыра W1-A4-005 без теста | Кейсы outsider на `/sprints/poker/:id/{vote,reveal}` + member OK |
| W1-A4-015 | P3 | `backend/testframework/sprint/service_extended_test.go:115-117` | `AddIssuesToSprint` с `issueID` без записи в `issueProject` — mock `SetIssuesSprint` пропускает «отсутствующий» issue | Тест зелёный и на слабом mock-guard; реальный Ent guard (`n != len`) здесь не бьётся (есть в routes/ent) | В mock: отсутствие `issueProject[id]` → `ErrIssueNotInProject`; в тесте задать project |
| W1-A4-016 | P3 | `backend/testframework/sprint/handler_wave113_test.go:30-38` | `StartSprint_NotFound`: `assert.NotEqual(OK)` вместо точного 404 | Проходит и на 500/400 — слабый сигнал регрессии IDOR/NotFound | `assert.Equal(StatusNotFound, …)` (+ тело `error`) |

## Чисто

Файлы / зоны без находок — что проверялось:

- **`dto.go`**: validate-теги на create; RetroNotes/Vote/Poker/Move DTO без логики; unbounded slice в retro — принимается как контракт JSON (Ent JSON), не дефект кода.
- **`export_hooks.go`**: только hubtest-экспорт `toSprintResponse` / `nilIfEmpty`; нет prod-путей.
- **`fail_closed_a3015_test.go`**: WIP fail-closed на слое service (где живёт guard); assert на NotFound и count-error; mutation-комментарий есть.
- **`fail_closed_a3016_test.go`**: FindActiveSprint DB error → Internal, статус planning сохранён; assert-ы есть.
- **`handler_flow_test.go`**: CRUD lifecycle, burndown, velocity, poker happy-path, invalid project UUID; asserts присутствуют (не assert-free).
- **`handler_routes_test.go`**: Update + AddIssues; foreign issue UUID → 400 (RT-005) на **реальном Ent**; слой guard верный.
- **`list_batch_counts_a13006_test.go`**: batch vs N× Count (mock counter) + Ent GROUP BY; mutation-инструкции.
- **`repository_ent_test.go`**: CRUD, FindOrCreateNext, SetIssuesSprint, burndown, velocity, WIP status count — против SQLite Ent.
- **`repository_get_in_project_a3001_test.go`**: Ent Postgres `HasProjectWith` — правильный слой для A3-001 guard.
- **`repository_list_no_issues_a13006_test.go`**: List без `WithIssues`; GetSprint capped.
- **`service_test.go`**: Create date validation; Start/Complete/WIP; полный IDOR-сьют RT-005/A3-001 на mock `GetSprintInProject` (дополняет Ent-тест); asserts есть.
- **`service_wave114_test.go`**: UpdateSprint NotFound — тонкий, но с assert.
- **IDOR A3-001 (prod)**: `GetSprint`/`Update`/`Start`/`Complete`/`Retro`/`Burndown`/`AddIssues` → `GetSprintInProject`; `SetIssuesSprint` с count-guard.
- **WIP A3-015 / Start A3-016 (prod)**: fail-closed подтверждён чтением `service.go:115-119`, `307-323`.
- **Fiber v3 RBAC order**: RegisterRoutes — handler затем `RequirePermission` (канон FA-02-02).
- **N+1 list**: `GetSprintIssueCounts` + List без eager issues.

## Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 1 |
| P2 | 8 |
| P3 | 7 |
| **Total** | **16** |

**Топ-3:** W1-A4-001 (CompleteSprint без tx) · W1-A4-002 (poker Reveal без CAS, A3-019) · W1-A4-005 (poker HTTP без project RBAC).
