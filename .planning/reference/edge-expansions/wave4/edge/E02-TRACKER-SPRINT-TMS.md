# Edge Expansion Wave4 E02 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Edge** (границы лимитов, fan-out интеграций, matrix size, rolling windows, export/import caps)  
> **ID ranges:** `ECC-TRCK-400+`, `ECC-SPR-400+`, `ECC-TMS-400+` — **не пересекают** wave1 `*100+`, wave2 `*200+`, wave3 `*300+`, канон `001–040`, corner `*C*`  
> **Не заменяет** каталог и `docs/TESTING.md` — углубление edge для July Rebuild / QA-HARDENING  
> **Wave1 (не дублировать 1:1):** `.planning/reference/edge-expansions/02-TRACKER-SPRINT-TMS.md` (`ECC-*-100…`)  
> **Wave2 (не дублировать 1:1):** `.planning/reference/edge-expansions/wave2/edge/E02-TRACKER-SPRINT-TMS.md` (`ECC-*-200…`)  
> **Wave3 (не дублировать 1:1):** `.planning/reference/edge-expansions/wave3/edge/E02-TRACKER-SPRINT-TMS.md` (`ECC-*-300…` — links/watchers/comments/worklogs, trash day, rank, saved filters, scope/carry-over/poker, native suite/run, shared steps, data-driven, version depth, env×case, flaky N, AS_ID)  
> **Фокус wave4:** GitHub/GitLab autolink+branch/PR · workflow/board columns+transitions · swimlanes · due-date/estimate · clone/move · JQL operator depth · multi-board · velocity rolling window · multi-planning · burndown ideal/zero · person-capacity · sprint archive · hub-cli/JUnit rate · export huge · traceability/coverage matrix · compare-runs · result attachments · parallel runs · TestRail import  
> **Связи:** US-M03/M11/M12, US-S01/S04/S06/S08, US-M05/S13/C02/C06 · NFR-08/12/14 · NF-03 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-JQL-*, TC-ALLURE-*  
> **Канон каталога:** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — **не изменять** этим файлом

### Gap-легенда (для этого файла)

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated edge-теста / TC |
| **Partial** | Happy/unit есть, граница не закрыта |
| **Future** | Лимит/фича ещё не зафиксированы в коде/ADR |

### Отличие от wave1 / wave2 / wave3

| Тема | Wave1 (`*100+`) | Wave2 (`*200+`) | Wave3 (`*300+`) | Wave4 (`*400+`) — новое |
|------|-----------------|-----------------|-----------------|-------------------------|
| Tracker | WIP, JQL len, labels, attach, pagination | CF types, epic depth, attach combo | links/watchers/comments/worklogs, trash, rank, saved filters | **GH/GL autolink**, workflow/columns/transitions, swimlanes, due/estimate, clone/move, JQL ops depth, multi-board |
| Sprint | capacity SP, same-day, empty backlog, velocity zeros | TZ/DST/leap, burndown buckets | scope mid-sprint, carry-over, poker voters, goal length | **velocity window N**, multi-planning, ideal/zero burndown, person-capacity, archive/rename, pre-Done start |
| TMS | Allure batch/BodyLimit, suite depth, BDD 0/max | Allure huge tree, BDD convert | native suite/run, shared/DD, version depth, env matrix, flaky N, AS_ID | **hub-cli/JUnit rate**, export huge, traceability/coverage matrix, compare-runs, result attach, parallel runs, TestRail |

---

## 1. Tracker — `ECC-TRCK-400+`

> Фокус: **GitHub/GitLab autolink**, **workflow/board columns+transitions**, **swimlanes**, **due-date/estimate**, **clone/move**, **JQL operator depth**, **multi-board**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-400 | Edge | GitHub webhook: commit message with exactly max issue keys (HUB-1…HUB-N) | all linked ≤max; max+1 ignored or 400 documented | API | P0 | US-S04, ECC-TRCK-034 | TODO |
| ECC-TRCK-401 | Edge | GitHub webhook: key at high seq `PROJ-999999` in PR title | autolink OK; no truncate | API | P1 | US-S04 | TODO |
| ECC-TRCK-402 | Edge | GitHub: create branch from issue when branch name length = max / max+1 | OK / 400 | API | P1 | US-S04 | TODO |
| ECC-TRCK-403 | Edge | GitHub: max open PR links per issue → add one more | 400 or soft-cap+warn | API | P1 | — | Future |
| ECC-TRCK-404 | Edge | GitLab MR autolink: same key appears max times in description | one link row or dedupe; no N duplicates | API | P1 | US-S06 | TODO |
| ECC-TRCK-405 | Edge | GitLab pipeline status fan-out: max pipelines linked to one issue | list paginated; no OOM | API | P2 | US-C06 | TODO |
| ECC-TRCK-406 | Edge | Autolink when issue soft-deleted | no new link or 404; no resurrect | API | P0 | NFR-11, TC-TRCK-12 | TODO |
| ECC-TRCK-407 | Edge | Workflow: statuses count = max−1 → add one | 201, count=max | API | P1 | US-M11 | TODO |
| ECC-TRCK-408 | Edge | Workflow: statuses count = max → add one | 400 «status limit» | API | P0 | ECC-TRCK-039 | TODO |
| ECC-TRCK-409 | Edge | Workflow: transitions count = max (dense matrix) | all editable; GET size bounded | API | P1 | TC-TRCK-10/11 | TODO |
| ECC-TRCK-410 | Edge | Workflow: transitions = max+1 | 400 | API | P1 | — | TODO |
| ECC-TRCK-411 | Edge | Board columns = statuses at max; WIP set on each | board loads; WIP independent per column | API/E2E | P1 | US-M11, ECC-TRCK-018 | TODO |
| ECC-TRCK-412 | Edge | Board columns count = max+1 (orphan status without column) | reject or hide orphan documented | API | P1 | — | TODO |
| ECC-TRCK-413 | Edge | Swimlanes count = 0 (disabled) | board flat; 200 | API/E2E | P2 | US-M11, ECC-TRCK-023 | TODO |
| ECC-TRCK-414 | Edge | Swimlanes count = max (e.g. by assignee) | all lanes; empty lanes shown or collapsed per policy | API/E2E | P1 | US-M11 | TODO |
| ECC-TRCK-415 | Edge | Swimlanes count = max+1 | 400 or clamp | API | P1 | — | Future |
| ECC-TRCK-416 | Edge | Swimlane × filter → exactly 1 issue in 1 lane | subset correct; other lanes empty | E2E | P2 | ECC-TRCK-023 | TODO |
| ECC-TRCK-417 | Edge | Due date = null vs omitted on create | consistent stored null; no epoch 1970 | API | P1 | NFR-12 | TODO |
| ECC-TRCK-418 | Edge | Due date at UTC midnight boundary; user TZ+14 display | DB UTC; local label next/prev day correct | API/Manual | P1 | NFR-12 | TODO |
| ECC-TRCK-419 | Edge | Due date far future (e.g. year 9999) / year 0 | 400 or clamp documented | API | P2 | — | TODO |
| ECC-TRCK-420 | Edge | Estimate = 0 with due date set | both OK; board shows 0 | API | P2 | ECC-TRCK-037/139 | TODO |
| ECC-TRCK-421 | Edge | Estimate fractional boundary (0.5 vs 0.01 vs max decimals) | accept per schema; overflow 400 | API | P2 | — | Future |
| ECC-TRCK-422 | Edge | Clone issue: fields at max (title/desc/labels) without attachments | 201; new key; version=1 | API | P1 | — | TODO |
| ECC-TRCK-423 | Edge | Clone when project issue count at soft-cap | 400 or allow | API | P2 | — | Future |
| ECC-TRCK-424 | Edge | Move issue to another project: key renumber at target high seq | unique target key; links policy documented | API | P0 | US-M03 | TODO |
| ECC-TRCK-425 | Edge | Move issue when target workflow lacks current status | map or block 400; no silent status loss | API | P0 | TC-TRCK-10 | TODO |
| ECC-TRCK-426 | Edge | JQL: nested parentheses depth = max / max+1 | OK / 400 complexity | API | P1 | TC-JQL-01, ECC-TRCK-109 | TODO |
| ECC-TRCK-427 | Edge | JQL: ORDER BY with max sort keys | 200 stable order; max+1 → 400 | API | P1 | TC-JQL-02 | TODO |
| ECC-TRCK-428 | Edge | JQL: `WAS`/`CHANGED` (if supported) on empty history | 200 empty; no 500 | API | P2 | Phase 17 JQL | Future |
| ECC-TRCK-429 | Edge | JQL: result set exactly page size = max limit | one full page; next offset empty or has more flag | API | P1 | NFR-14, ECC-TRCK-132 | Partial |
| ECC-TRCK-430 | Edge | Multi-board: boards per project = max−1 → create | 201, count=max | API | P1 | ECC-TRCK-038 | TODO |
| ECC-TRCK-431 | Edge | Multi-board: boards = max → create one | 400 | API | P1 | — | Future |
| ECC-TRCK-432 | Edge | Same issue visible on max boards (shared statuses) | move on one board reflects all; no desync | API/E2E | P1 | US-M11 | TODO |
| ECC-TRCK-433 | Edge | Board filter JQL length at max on secondary board | same limits as primary search | API | P2 | ECC-TRCK-106 | TODO |
| ECC-TRCK-434 | Edge | Sub-tasks under story: count = max / max+1 (not epic depth) | OK / 400 child-cap | API | P1 | ECC-TRCK-033 | TODO |
| ECC-TRCK-435 | Edge | Components/versions catalog: count = max / max+1 (if enabled) | OK / 400 | API | P2 | — | Future |
| ECC-TRCK-436 | Edge | Bulk transition: exactly max issues in one request | all moved or atomic fail; max+1 → 400 | API | P1 | bulk, ECC-TRCK-213 | Future |
| ECC-TRCK-437 | Edge | Assignee = self when user is only project member | 200; board swimlane self | API | P2 | RBAC | TODO |
| ECC-TRCK-438 | Edge | Reporter deactivated; GET issue still returns reporter stub | 200; no PII leak beyond policy | API | P1 | ECC-TRCK-028 | TODO |

**Tracker Edge в этом файле: 39**

---

## 2. Sprint — `ECC-SPR-400+`

> Фокус: **velocity rolling window**, **multi-planning**, **burndown ideal/zero**, **person-capacity**, **archive/rename**, **pre-Done start**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-400 | Edge | Velocity: rolling window = N completed (default) | exactly N points; oldest dropped | API | P0 | TC-SPRINT-05, ECC-SPR-116 | TODO |
| ECC-SPR-401 | Edge | Velocity: completed count = N−1 (< window) | series length = N−1; no pad NaN | API | P1 | ECC-SPR-115 | TODO |
| ECC-SPR-402 | Edge | Velocity: completed count = N+1 | still N points (drop oldest) | API | P1 | — | TODO |
| ECC-SPR-403 | Edge | Velocity: window param = 0 / negative / max+1 | 400 or clamp to default | API | P1 | — | Future |
| ECC-SPR-404 | Edge | Velocity: all N sprints planned=0 done=0 | zeros; chart renders | API | P1 | ECC-SPR-117 | TODO |
| ECC-SPR-405 | Edge | Planning sprints open count = max−1 → create | 201 | API | P1 | ECC-SPR-001 | TODO |
| ECC-SPR-406 | Edge | Planning sprints open count = max → create | 400 «too many planning» | API | P1 | — | Future |
| ECC-SPR-407 | Edge | One active + max planning coexist | start blocked only by active rule; planning OK | API | P0 | TC-SPRINT-03 | TODO |
| ECC-SPR-408 | Edge | Burndown: ideal line when remaining = 0 on day 0 | flat zero or single point; no div/0 | API | P1 | TC-SPRINT-04 | TODO |
| ECC-SPR-409 | Edge | Burndown: all issues estimate null/0 | ideal+actual zero series; 200 | API | P1 | ECC-SPR-121 | TODO |
| ECC-SPR-410 | Edge | Burndown: actual > ideal on every day (scope creep) | series honest; no clamp to ideal | API | P2 | ECC-SPR-119 | TODO |
| ECC-SPR-411 | Edge | Burndown snapshot cadence: query twice same UTC day | same day bucket; idempotent points | API | P1 | NFR-12 | TODO |
| ECC-SPR-412 | Edge | Person-capacity: sum member capacities = 0 → start | warn/allow or 400 per policy | API | P2 | US-S01 | Future |
| ECC-SPR-413 | Edge | Person-capacity: assigned load = capacity exactly | start OK | API | P2 | — | Future |
| ECC-SPR-414 | Edge | Person-capacity: load = capacity+1 | warn/block documented | API | P2 | — | Future |
| ECC-SPR-415 | Edge | Sprint name length = max / max+1 | OK / 400 | API | P2 | TC-SPRINT-01 | TODO |
| ECC-SPR-416 | Edge | Rename active sprint at max name length | 200; burndown title updates | API | P2 | — | TODO |
| ECC-SPR-417 | Edge | Archive completed sprint when archived count = max | OK at max; max+1 → 400 or GC oldest | API | P2 | — | Future |
| ECC-SPR-418 | Edge | List sprints: offset/limit at archive-heavy project | pages stable; completed not lost | API | P1 | NFR-14 | TODO |
| ECC-SPR-419 | Edge | Start sprint where all issues already Done | active; burndown flat done; complete → nothing to move | API | P1 | TC-SPRINT-02/07 | TODO |
| ECC-SPR-420 | Edge | Start sprint: mix Done + ToDo at capacity boundary by count | start OK; WIP on board still enforced | API | P1 | ECC-SPR-013/122 | TODO |
| ECC-SPR-421 | Edge | Calendar aggregator: sprint event at start/end UTC midnight | single event span; no duplicate day events | API | P2 | US-V03, ECC-SPR-021 | Partial |
| ECC-SPR-422 | Edge | Complete → move incomplete when destination backlog already at max issues | 400 or overflow policy; sprint not half-completed | API | P0 | TC-SPRINT-07 | TODO |
| ECC-SPR-423 | Edge | Velocity after archived sprints only in window | archived still count or excluded — documented | API | P2 | — | Future |

**Sprint Edge в этом файле: 24**

---

## 3. TMS — `ECC-TMS-400+`

> Фокус: **hub-cli / JUnit rate**, **export huge**, **traceability/coverage matrix**, **compare-runs**, **result attachments**, **parallel runs**, **TestRail import**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-400 | Edge | hub-cli `push-results`: batch size = max−1 → +1 | 200, count=max | API/CLI | P0 | US-S08 | TODO |
| ECC-TMS-401 | Edge | hub-cli `push-results`: batch = max → +1 | 400/413 | API/CLI | P0 | — | TODO |
| ECC-TMS-402 | Edge | hub-cli push at API rate limit boundary (100/sec) | 429 with request_id; no partial corrupt run | API | P0 | NFR rate | TODO |
| ECC-TMS-403 | Edge | JUnit webhook: testsuites count = max / max+1 | OK / 400 | API | P1 | TC-TMS-08, ECC-TMS-130 | TODO |
| ECC-TMS-404 | Edge | JUnit webhook: single testcase name length = max / max+1 | OK / 400 | API | P2 | — | TODO |
| ECC-TMS-405 | Edge | JUnit webhook: identical payload replay within TTL | idempotent; no double stats | API | P0 | ECC-TMS-026 | Partial |
| ECC-TMS-406 | Edge | Export CSV: run with results count = max page | streams; Content-Length or chunked OK | API | P1 | TC-TMS-12, ECC-TMS-019 | Partial |
| ECC-TMS-407 | Edge | Export PDF: run with 0 cases vs 1 case vs max cases | empty policy / OK / bounded time | API | P1 | ECC-TMS-128 | Partial |
| ECC-TMS-408 | Edge | Export when run still in_progress | 400 or snapshot consistent | API | P1 | — | TODO |
| ECC-TMS-409 | Edge | Traceability matrix: cases = max, each linked to 1 issue | 200; cells complete; no silent truncate | API | P0 | TC-TMS-11, US-C02 | TODO |
| ECC-TMS-410 | Edge | Traceability: cases = max+1 | 400 or paginated matrix | API | P1 | NFR-14 | TODO |
| ECC-TMS-411 | Edge | Traceability: one case linked to max issues | row fan-out OK; max+1 → 400 | API | P1 | ECC-TMS-024 | TODO |
| ECC-TMS-412 | Edge | Coverage report: 0 linked cases in project | 200 empty / 0% | API | P2 | TC-TMS-10 | TODO |
| ECC-TMS-413 | Edge | Coverage report: 100% cases linked + all passed | 100%; no float 99.999 | API | P1 | — | TODO |
| ECC-TMS-414 | Edge | Compare runs: left=0 results, right=max | diff = all added; no crash | API | P1 | US-C02, ECC-TMS-029 | TODO |
| ECC-TMS-415 | Edge | Compare runs: identical max-size runs | empty diff; 200 | API | P1 | — | TODO |
| ECC-TMS-416 | Edge | Compare runs: status-only changes at scale | counted; order stable | API | P2 | — | TODO |
| ECC-TMS-417 | Edge | Result attachment count = max−1 → add one | 201, count=max | API | P1 | storage | TODO |
| ECC-TMS-418 | Edge | Result attachment count = max → add one | 400 | API | P1 | — | TODO |
| ECC-TMS-419 | Edge | Result attachment size = BodyLimit 25 MB / +1 byte | OK / 413 | API | P0 | NFR-08, NF-09 | Partial |
| ECC-TMS-420 | Edge | Parallel runs active per project = max−1 → start one | 201 | API | P1 | TC-TMS-05 | TODO |
| ECC-TMS-421 | Edge | Parallel runs active = max → start one | 400 «run limit» | API | P1 | — | Future |
| ECC-TMS-422 | Edge | Same case in max parallel runs | results isolated per run; no cross-write | API | P0 | NFR-24 | TODO |
| ECC-TMS-423 | Edge | TestRail import: cases = max / max+1 | OK / 400 or partial report | API | P1 | ECC-TMS-023, ECC-TMS-122 | Partial |
| ECC-TMS-424 | Edge | TestRail import: suite path depth = max (native tree) | mapped; max+1 reject | API | P1 | ECC-TMS-123 | TODO |
| ECC-TMS-425 | Edge | TestRail import: empty file / 0 cases | 400 or empty report | API | P2 | — | TODO |
| ECC-TMS-426 | Edge | Match strategy priority at scale: annotation vs AS_ID vs title (N=max) | deterministic winner per ECC-TMS-017; duration OK | API | P1 | TC-ALLURE-05 | Partial |
| ECC-TMS-427 | Edge | Create bug from failed: title length = max / max+1 | OK / 400; link case↔bug set | API | P1 | TC-TMS-09 | TODO |
| ECC-TMS-428 | Edge | Soft-deleted case: retention day 29 vs 30 vs 31 | restore OK / OK / hard-gone per NFR-11 | API | P1 | NFR-11 | TODO |
| ECC-TMS-429 | Edge | Mongo result document near 16 MB BSON limit | accept or 413 before write; no corrupt partial | API | P0 | Mongo | TODO |
| ECC-TMS-430 | Edge | Run stats: passed+failed+skipped+blocked = total at max cases | sum exact; no drift | API | P1 | TC-TMS-06 | Partial |
| ECC-TMS-431 | Edge | Environments filter: select 0 of max envs vs all max | empty / full result sets | API | P2 | ECC-TMS-022/127 | Partial |

**TMS Edge в этом файле: 32**

---

## Сводка счётчиков

| Домен | ID range | Edge count | Gap TODO | Gap Partial | Gap Future |
|-------|----------|------------|----------|-------------|------------|
| Tracker | ECC-TRCK-400…438 | **39** | 28 | 1 | 10 |
| Sprint | ECC-SPR-400…423 | **24** | 14 | 1 | 9 |
| TMS | ECC-TMS-400…431 | **32** | 20 | 8 | 4 |
| **ИТОГО** | | **95** | **62** | **10** | **23** |

| Метрика | Значение |
|---------|----------|
| Тип сценариев | **100% Edge** (Happy/Neg/Corner = 0) |
| Минимум задачи | ≥50 |
| Факт | **95** |
| Пересечение с wave1 `*100+` | **нет** (не 1:1) |
| Пересечение с wave2 `*200+` | **нет** (не 1:1) |
| Пересечение с wave3 `*300+` | **нет** (другие фокусы) |
| Пересечение с каталогом 001–040 | **нет** |
| Изменения `EDGE-CORNER-CATALOG.md` | **нет** |
| Коммит | **не создавался** |

### Фокус-покрытие (wave4)

| Фокус | ID (примеры) | Count ≈ |
|-------|--------------|--------:|
| GitHub/GitLab autolink+branch/PR | ECC-TRCK-400…406 | 7 |
| Workflow/board columns+transitions | ECC-TRCK-407…412 | 6 |
| Swimlanes | ECC-TRCK-413…416 | 4 |
| Due-date/estimate + clone/move | ECC-TRCK-417…425, 434–438 | 14 |
| JQL ops depth + multi-board | ECC-TRCK-426…433 | 8 |
| Velocity window + multi-planning | ECC-SPR-400…407 | 8 |
| Burndown ideal/zero + person-cap + archive | ECC-SPR-408…423 | 16 |
| hub-cli/JUnit rate | ECC-TMS-400…405 | 6 |
| Export + traceability/coverage + compare | ECC-TMS-406…416 | 11 |
| Result attach + parallel runs + TestRail + misc | ECC-TMS-417…431 | 15 |

### Приоритеты (P)

| P | Count |
|---|------:|
| P0 | 16 |
| P1 | 52 |
| P2 | 27 |

### Рекомендуемый next step

1. `qa-manual-ver` — верификация wave4 expansion (нет дублей *100–300, границы, трассируемость).  
2. P0: GH/GL autolink soft-delete, move cross-project status map, velocity window, complete→backlog cap, hub-cli batch/rate, traceability max, Mongo 16 MB → полные `TC-*` в `docs/TESTING.md`.  
3. После verified → `qa-auto1` (API/CLI/webhook) + `qa-auto2` (swimlane/multi-board E2E).  
4. Каталог не трогать; импорт в канон — отдельным PR позже.

---

*Создано: 2026-07-10 · qa-manual · wave4 · только Edge · domains Tracker + Sprint + TMS*
