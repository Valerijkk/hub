# Edge Expansion Wave5 E02 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Edge** (границы коллекций, %/precision, webhook/IMEX caps, board/automation cardinality)  
> **ID ranges:** `ECC-TRCK-500+`, `ECC-SPR-500+`, `ECC-TMS-500+` — **не пересекают** wave1 `*100+`, wave2 `*200+`, wave3 `*300+`, канон `001–040`, corner `*C*`  
> **Не заменяет** каталог и `docs/TESTING.md` — углубление edge для July Rebuild / QA-HARDENING  
> **Wave1 (не дублировать 1:1):** `.planning/reference/edge-expansions/02-TRACKER-SPRINT-TMS.md` (`ECC-*-100…`)  
> **Wave2 (не дублировать 1:1):** `.planning/reference/edge-expansions/wave2/edge/E02-TRACKER-SPRINT-TMS.md` (`ECC-*-200…`)  
> **Wave3 (не дублировать 1:1):** `.planning/reference/edge-expansions/wave3/edge/E02-TRACKER-SPRINT-TMS.md` (`ECC-*-300…`)  
> **Фокус wave5:** components/fix-versions · swimlanes/boards/columns · automation rule caps · bulk transition/IMEX · git autolink · due-date buckets · story-point precision · sprint commitment/% · capacity mid-edit · calendar/retro bounds · JUnit webhook caps · coverage % · compare-runs · hub-cli batch · native result attachments · case tags/title · bug-from-fail · TestRail rows · webhook clock skew  
> **Связи:** US-M03/M11/M12, US-S01/S04/S06/S08/S09, US-M05/S13/C02 · NFR-08/11/12/14/24 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*, TC-JQL-*  
> **Канон каталога:** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — **не изменять** этим файлом

### Gap-легенда (для этого файла)

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated edge-теста / TC |
| **Partial** | Happy/unit есть, граница не закрыта |
| **Future** | Лимит/фича ещё не зафиксированы в коде/ADR |

### Отличие от wave1 / wave2 / wave3

| Тема | Wave1 (`*100+`) | Wave2 (`*200+`) | Wave3 (`*300+`) | Wave5 (`*500+`) — новое |
|------|-----------------|-----------------|-----------------|-------------------------|
| Tracker structure | WIP, JQL len, labels, attach size, pagination | CF types, epic depth, attach count | links/watchers/comments/worklogs, trash, rank, filters | **components/versions, swimlanes/boards/columns, automation caps, bulk transition, IMEX rows, git links, due-date** |
| Sprint metrics | capacity 0/exact, empty backlog, velocity zeros | TZ/DST/leap buckets | scope/carry-over/poker voters | **commitment %, unestimated ratio, capacity mid-edit, SP precision, calendar/retro, forecast@vel=0** |
| TMS ingress/reports | Allure batch/BodyLimit, suite depth, BDD 0/max | Allure huge tree, BDD convert | native suite/run, shared/DD, version/flaky/AS_ID | **JUnit suite/case caps, coverage %, compare-runs, hub-cli batch, result attach, tags, bug-from-fail, TestRail rows, skew** |

---

## 1. Tracker — `ECC-TRCK-500+`

> Фокус: **components / fix-versions**, **swimlanes / boards / columns**, **automation rule caps**, **bulk transition / IMEX**, **git autolink**, **due-date buckets**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-500 | Edge | Project components count = max−1 → create one | 201, count=max | API | P1 | US-M03 | Future |
| ECC-TRCK-501 | Edge | Project components count = max → create one | 400 «component limit» | API | P1 | — | Future |
| ECC-TRCK-502 | Edge | Issue: assign components = max / max+1 | 200 / 400 | API | P1 | — | Future |
| ECC-TRCK-503 | Edge | Component name length = max / max+1 | 201 / 400 | API | P2 | — | Future |
| ECC-TRCK-504 | Edge | Soft-delete component while max issues reference it | block or null-out; no dangling FK | API | P0 | NFR-11 | Future |
| ECC-TRCK-505 | Edge | Fix-versions (releases) count = max−1 → create | 201, count=max | API | P1 | US-M03 | Future |
| ECC-TRCK-506 | Edge | Fix-versions count = max → create | 400 | API | P1 | — | Future |
| ECC-TRCK-507 | Edge | Issue: fixVersions selected = max / max+1 | 200 / 400 | API | P1 | — | Future |
| ECC-TRCK-508 | Edge | Release date = past / far-future (UTC year+100) | accept or 400 policy; stored UTC | API | P2 | NFR-12 | Future |
| ECC-TRCK-509 | Edge | Boards per project = max−1 → create board | 201 | API | P1 | US-M11, ECC-TRCK-038 | Future |
| ECC-TRCK-510 | Edge | Boards per project = max → create | 400 | API | P1 | — | Future |
| ECC-TRCK-511 | Edge | Columns per board = max−1 → add column | 201; workflow map intact | API | P0 | US-M11 | Future |
| ECC-TRCK-512 | Edge | Columns per board = max → add | 400 | API | P0 | — | Future |
| ECC-TRCK-513 | Edge | Swimlanes per board = max−1 → add | 201 | API | P1 | ECC-TRCK-023 | Future |
| ECC-TRCK-514 | Edge | Swimlanes per board = max → add | 400 | API | P1 | — | Future |
| ECC-TRCK-515 | Edge | Swimlane group-by cardinality = max distinct values | board renders; no silent drop | API/E2E | P1 | US-M11 | TODO |
| ECC-TRCK-516 | Edge | Swimlane group-by cardinality = max+1 | overflow lane «Other» or 400 — documented | API/E2E | P2 | — | Future |
| ECC-TRCK-517 | Edge | Automation rules per project = max−1 → create | 201 | API | P0 | ECC-TRCK-029 | Future |
| ECC-TRCK-518 | Edge | Automation rules = max → create | 400 | API | P0 | automation | Future |
| ECC-TRCK-519 | Edge | One rule: conditions count = max / max+1 | OK / 400 | API | P1 | — | Future |
| ECC-TRCK-520 | Edge | One rule: actions count = max / max+1 | OK / 400 | API | P1 | — | Future |
| ECC-TRCK-521 | Edge | Automation fire rate: max matches in 1s burst | throttle/queue; no handler storm | API/Perf | P0 | NFR-14 | Future |
| ECC-TRCK-522 | Edge | Bulk transition: issues = max−1 → apply | 200; all status updated; version++ each | API | P0 | TC-TRCK-10, bulk | Future |
| ECC-TRCK-523 | Edge | Bulk transition: issues = max → apply | 400 or partial+report | API | P0 | — | Future |
| ECC-TRCK-524 | Edge | Bulk transition: mix valid+invalid transitions in one batch | per-item errors; no half-applied silent | API | P1 | TC-TRCK-11 | Future |
| ECC-TRCK-525 | Edge | IMEX issue CSV: rows = max−1 → import | 200; created=max−1 | API | P1 | US-S09 | Partial |
| ECC-TRCK-526 | Edge | IMEX issue CSV: rows = max → import | 200 or soft-cap | API | P1 | imex | Future |
| ECC-TRCK-527 | Edge | IMEX issue CSV: rows = max+1 | 400/413; 0 partial writes or documented resume | API | P0 | BodyLimit/NFR-08 | TODO |
| ECC-TRCK-528 | Edge | GitHub/GitLab autolinks per issue = max−1 → add | 201 | API | P2 | US-S04/S06, ECC-TRCK-034 | Partial |
| ECC-TRCK-529 | Edge | Autolinks per issue = max → add one | 400 | API | P2 | — | Future |
| ECC-TRCK-530 | Edge | Due date = UTC midnight boundary (create/update) | stored UTC; GET round-trip; board due-bucket correct | API | P1 | NFR-12 | TODO |
| ECC-TRCK-531 | Edge | Due date far past (year 1970) / far future (year+50) | accept or 400; overdue filter consistent | API | P2 | — | TODO |
| ECC-TRCK-532 | Edge | Sub-tasks under one parent story = max−1 → create | 201, count=max | API | P1 | US-M03 (≠ epic depth *200) | Future |
| ECC-TRCK-533 | Edge | Sub-tasks under parent = max → create | 400 | API | P1 | — | Future |
| ECC-TRCK-534 | Edge | Workflow statuses in scheme = max−1 → add | 201 | API | P1 | ECC-TRCK-039 sibling | Future |
| ECC-TRCK-535 | Edge | Workflow statuses = max → add | 400 | API | P1 | — | Future |
| ECC-TRCK-536 | Edge | Transitions from one status = max / max+1 | OK / 400 | API | P2 | TC-TRCK-10/11 | Future |
| ECC-TRCK-537 | Edge | Create-from-template: template fields all at max length | 201; no BodyLimit truncate silent | API | P2 | templates | Future |
| ECC-TRCK-538 | Edge | Story points field: 0.5 / 0.25 / unsupported 0.1 | accept Fibonacci-like set; else 400 | API | P1 | ECC-TRCK-037 deepen ≠ *139/140 | Future |
| ECC-TRCK-539 | Edge | Story points = max card value / max+1 | OK / 400 | API | P1 | poker cards | Future |

**Tracker Edge в этом файле: 40**

---

## 2. Sprint — `ECC-SPR-500+`

> Фокус: **commitment %**, **unestimated ratio**, **capacity mid-edit**, **SP precision**, **calendar/retro bounds**, **forecast @ velocity=0**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-500 | Edge | Commitment % = 0 (planned=0, capacity>0) at start | start OK/warn; chart shows 0% not NaN | API | P1 | US-S01, ECC-SPR-100 sibling | Future |
| ECC-SPR-501 | Edge | Commitment % = 100 (planned == capacity) | start OK; UI 100% | API | P1 | ECC-SPR-101 | Future |
| ECC-SPR-502 | Edge | Commitment % > 100 (planned = capacity+ε) | warn/block per policy; % capped display or honest >100 | API | P0 | ECC-SPR-102 | Future |
| ECC-SPR-503 | Edge | Unestimated issues ratio = 100% (all SP null) at start | allow/warn; burndown flat policy | API | P1 | ECC-SPR-121 deepen | TODO |
| ECC-SPR-504 | Edge | Unestimated ratio = 0% (all estimated) | burndown ideal line from sum SP | API | P1 | TC-SPRINT-04 | Partial |
| ECC-SPR-505 | Edge | Mix: exactly 1 unestimated + rest estimated | remaining includes/excludes null per policy | API | P1 | — | TODO |
| ECC-SPR-506 | Edge | Capacity edit mid-active sprint: capacity ↓ below planned | warn; commitment % recalc; burndown ideal regen or freeze — documented | API | P0 | — | Future |
| ECC-SPR-507 | Edge | Capacity edit mid-active: capacity ↑ | % drops; no crash; history audit | API | P1 | — | Future |
| ECC-SPR-508 | Edge | Capacity = fractional (e.g. 12.5) if SP half-points allowed | accept or 400 integer-only | API | P2 | ECC-TRCK-538 | Future |
| ECC-SPR-509 | Edge | Issue SP change after sprint start (0.5→8) | scope delta logged; burndown scope line jumps honestly | API | P0 | ECC-SPR-303 sibling ≠ copy | TODO |
| ECC-SPR-510 | Edge | Remaining working days = 0 (query on end calendar day) | forecast/remaining=0; no div-by-zero | API | P0 | TC-SPRINT-04 | TODO |
| ECC-SPR-511 | Edge | Forecast completion when velocity window avg = 0 | null/∞ policy; 200 no NaN | API | P1 | TC-SPRINT-05 | TODO |
| ECC-SPR-512 | Edge | Sprint ↔ calendar event: duration matches start/end UTC | single event; no duplicate day spill | API | P1 | US-V03, ECC-SPR-021 | Partial |
| ECC-SPR-513 | Edge | Calendar event for sprint length = max weeks | event accepted; feed size OK | API | P2 | ECC-SPR-107 sibling | Future |
| ECC-SPR-514 | Edge | Retro notes body length = max / max+1 | 200 / 400 | API | P2 | ECC-SPR-017 | Future |
| ECC-SPR-515 | Edge | Retro action items count = max / max+1 | OK / 400 | API | P2 | — | Future |
| ECC-SPR-516 | Edge | Completed sprints list: page at limit=max, offset last | partial page; stable order by end_date | API | P1 | NFR-14, ECC-SPR-321 sibling | TODO |
| ECC-SPR-517 | Edge | Active sprint board: all columns empty (issues filtered out) | 200 empty columns; membership unchanged | API/E2E | P2 | ECC-SPR-319 deepen | TODO |
| ECC-SPR-518 | Edge | Scope churn: same issue remove→add N=max times in active | membership final=in; audit bounded; burndown not corrupt | API | P1 | — | TODO |
| ECC-SPR-519 | Edge | Sprint name with unicode/emoji at max runes | 201; max+1 → 400 (rune count) | API | P2 | ECC-SPR-312 deepen ≠ length-only copy | TODO |
| ECC-SPR-520 | Edge | Filter sprint backlog by assignee=empty (unassigned only) | only null-assignee; count correct at max backlog | API | P2 | — | TODO |
| ECC-SPR-521 | Edge | Ideal burndown day-index = 0 and = last inclusive | both points present; length = day_count | API | P1 | ECC-SPR-120 deepen | Partial |
| ECC-SPR-522 | Edge | Done SP = capacity exactly on complete | velocity point = capacity; ratio 1.0 | API | P1 | TC-SPRINT-05/07 | Partial |
| ECC-SPR-523 | Edge | Done SP = 0 on complete with capacity>0 | velocity 0; no Inf in charts | API | P1 | ECC-SPR-118 | TODO |

**Sprint Edge в этом файле: 24**

---

## 3. TMS — `ECC-TMS-500+`

> Фокус: **JUnit webhook caps**, **coverage %**, **compare-runs**, **hub-cli batch**, **native result attachments**, **tags/title**, **bug-from-fail**, **TestRail rows**, **webhook clock skew**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-500 | Edge | JUnit webhook: testsuites count = max−1 | 200; all ingested | API | P0 | TC-TMS-08, US-S08 | TODO |
| ECC-TMS-501 | Edge | JUnit: testsuites = max → +1 suite | 400/413 or truncate+report | API | P0 | ECC-TMS-130 sibling ≠ BodyLimit-only | Future |
| ECC-TMS-502 | Edge | JUnit: testcases in one suite = max / max+1 | OK / reject | API | P1 | — | Future |
| ECC-TMS-503 | Edge | JUnit: testcase time=0.0 / negative time | 0 OK; negative → 400 or clamp | API | P1 | — | TODO |
| ECC-TMS-504 | Edge | JUnit: total failures = all cases (0 passes) | stats fail=N; coverage/run OK | API | P1 | TC-TMS-06 | TODO |
| ECC-TMS-505 | Edge | Coverage report: overall = 0% (0 linked cases executed) | 200, 0.0; no NaN | API | P1 | TC-TMS-10, US-C02 | Partial |
| ECC-TMS-506 | Edge | Coverage report: overall = 100% | 200, 1.0 or 100 | API | P1 | — | Partial |
| ECC-TMS-507 | Edge | Coverage: suite with 0 cases in denominator | skip suite or 0%; no div-by-zero | API | P0 | — | TODO |
| ECC-TMS-508 | Edge | Compare runs: case-count = max in both runs | diff size bounded; <timeout | Perf/API | P1 | ECC-TMS-029 | Future |
| ECC-TMS-509 | Edge | Compare runs: left empty (0 results) vs right full | full added set; 200 | API | P2 | — | Future |
| ECC-TMS-510 | Edge | Compare runs: identical results at max cases | empty diff; 200 | API | P2 | — | Future |
| ECC-TMS-511 | Edge | hub-cli push-results batch = max−1 | 200; all mapped | API/CLI | P1 | US-S08 | Future |
| ECC-TMS-512 | Edge | hub-cli push-results batch = max → +1 | 400/413 | API/CLI | P1 | — | Future |
| ECC-TMS-513 | Edge | Native result attachment count = max−1 → add | 201 | API | P1 | TC-TMS-06 (≠ Allure attach *210) | Future |
| ECC-TMS-514 | Edge | Native result attachments = max → add | 400 | API | P1 | — | Future |
| ECC-TMS-515 | Edge | Native result attachment size = 25 MB / 25 MB+1 | OK / 413 | API | P0 | NFR-08, NF-09 | Partial |
| ECC-TMS-516 | Edge | Case tags count = max−1 → add tag | 200, count=max | API | P2 | US-S13 | Future |
| ECC-TMS-517 | Edge | Case tags = max → add | 400 | API | P2 | — | Future |
| ECC-TMS-518 | Edge | Case title length = max / max+1 | 201 / 400 | API | P1 | TC-TMS-01 | TODO |
| ECC-TMS-519 | Edge | Case key prefix + seq at DB string limit | create OK; UI truncates display only | API | P2 | ECC-TMS-126 deepen | TODO |
| ECC-TMS-520 | Edge | Bug-from-fail: title auto length at max from long case name | 201 Bug; truncated per policy | API | P1 | TC-TMS-09 | Partial |
| ECC-TMS-521 | Edge | Bug-from-fail: N=max failed results → create bugs batch | rate-limit/batch; no issue storm | API/Perf | P1 | ECC-TMS-232 sibling ≠ Allure-only | Future |
| ECC-TMS-522 | Edge | Bug-from-fail twice on same result | 400 or return existing bug id (idempotent) | API | P0 | TC-TMS-09 | TODO |
| ECC-TMS-523 | Edge | TestRail import: rows = max−1 | 200; cases created | API | P1 | US-S09, ECC-TMS-023 | Partial |
| ECC-TMS-524 | Edge | TestRail import: rows = max+1 | 400/413; error report complete | API | P0 | ECC-TMS-122 deepen ≠ size-only | TODO |
| ECC-TMS-525 | Edge | TestRail import: partial fail rows = max errors listed | report truncated with total_errors count | API | P1 | ECC-TMS-023 | Partial |
| ECC-TMS-526 | Edge | Webhook signature clock skew = ±max allowed minutes | accept within window | API | P0 | TC-TMS-08, SEC | TODO |
| ECC-TMS-527 | Edge | Webhook clock skew = ±(max+1) minutes | 401/400 reject | API | P0 | — | TODO |
| ECC-TMS-528 | Edge | Result comment/note length = max / max+1 | 200 / 400 | API | P2 | — | Future |
| ECC-TMS-529 | Edge | Suite full path string length = max / max+1 | OK / 400 | API | P2 | ECC-TMS-020 | Future |
| ECC-TMS-530 | Edge | Traceability matrix: page limit=max, offset last link | partial page; no foreign cases (NF-03) | API | P1 | TC-TMS-11, NFR-14 | Partial |
| ECC-TMS-531 | Edge | Traceability: 0 links in project | 200 empty matrix | API | P2 | — | TODO |
| ECC-TMS-532 | Edge | Run description length = max / max+1 | 201 / 400 | API | P2 | TC-TMS-05 | Future |
| ECC-TMS-533 | Edge | Priority/severity enum at last valid + unknown | last OK; unknown → 400 | API | P1 | ECC-TMS-027 | Partial |
| ECC-TMS-534 | Edge | AS_ID match across projects: same AS_ID string | match scoped to project; no cross-leak | API | P0 | NF-03, ECC-TMS-320 sibling | TODO |

**TMS Edge в этом файле: 35**

---

## Сводка счётчиков

| Домен | ID range | Edge count | Gap TODO | Gap Partial | Gap Future |
|-------|----------|------------|----------|-------------|------------|
| Tracker | ECC-TRCK-500…539 | **40** | 4 | 2 | 34 |
| Sprint | ECC-SPR-500…523 | **24** | 9 | 4 | 11 |
| TMS | ECC-TMS-500…534 | **35** | 10 | 7 | 18 |
| **ИТОГО** | | **99** | **23** | **13** | **63** |

| Метрика | Значение |
|---------|----------|
| Тип сценариев | **100% Edge** (Happy/Neg/Corner = 0) |
| Минимум задачи | ≥50 |
| Факт | **99** |
| Пересечение с wave1 `*100+` | **нет** (новые темы, не 1:1) |
| Пересечение с wave2 `*200+` | **нет** |
| Пересечение с wave3 `*300+` | **нет** |
| Пересечение с каталогом 001–040 | **нет** (только ссылки-связи) |
| Изменения `EDGE-CORNER-CATALOG.md` | **нет** |
| Коммит | **не создавался** |

### Фокус-покрытие (wave5)

| Фокус | ID (примеры) | Count ≈ |
|-------|--------------|--------:|
| Components / fix-versions | ECC-TRCK-500…508 | 9 |
| Boards / columns / swimlanes | ECC-TRCK-509…516 | 8 |
| Automation caps / fire rate | ECC-TRCK-517…521 | 5 |
| Bulk transition / IMEX / git / due / sub-tasks / workflow / SP | ECC-TRCK-522…539 | 18 |
| Commitment % / unestimated / capacity mid-edit | ECC-SPR-500…509 | 10 |
| Forecast / calendar / retro / churn / precision display | ECC-SPR-510…523 | 14 |
| JUnit webhook caps | ECC-TMS-500…504 | 5 |
| Coverage % / compare-runs / hub-cli | ECC-TMS-505…512 | 8 |
| Result attach / tags / title / bug-from-fail | ECC-TMS-513…522 | 10 |
| TestRail / skew / traceability / enums / AS_ID scope | ECC-TMS-523…534 | 12 |

### Приоритеты (P)

| P | Count |
|---|------:|
| P0 | 22 |
| P1 | 52 |
| P2 | 25 |

### Рекомендуемый next step

1. `qa-manual-ver` — верификация wave5 expansion (границы, нет дублей *100–300, трассируемость).  
2. P0: TRCK automation/bulk/IMEX, SPR commitment/capacity mid-edit/remaining=0, TMS JUnit/coverage-div0/skew/bug-idempotent → полные `TC-*` в `docs/TESTING.md`.  
3. После verified → `qa-auto1` (API caps/JUnit/IMEX) + `qa-auto2` (swimlane/board E2E).  
4. Каталог не трогать; импорт в канон — отдельным PR позже.

---

*Создано: 2026-07-10 · qa-manual · wave5 · только Edge · domains Tracker + Sprint + TMS*
