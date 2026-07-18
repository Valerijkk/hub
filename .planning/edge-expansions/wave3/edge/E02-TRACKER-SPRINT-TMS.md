# Edge Expansion Wave3 E02 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Edge** (границы коллекций, day-buckets trash, fan-out links/worklogs, scope/carry-over, native suite/run caps)  
> **ID ranges:** `ECC-TRCK-300+`, `ECC-SPR-300+`, `ECC-TMS-300+` — **не пересекают** wave1 `*100+`, wave2 `*200+`, канон `001–040`, corner `*C*`  
> **Не заменяет** каталог и `docs/TESTING.md` — углубление edge для July Rebuild / QA-HARDENING  
> **Wave1 (не дублировать 1:1):** `.planning/edge-expansions/02-TRACKER-SPRINT-TMS.md` (`ECC-*-100…`)  
> **Wave2 (не дублировать 1:1):** `.planning/edge-expansions/wave2/edge/E02-TRACKER-SPRINT-TMS.md` (`ECC-*-200…`)  
> **Фокус wave3:** issue links/watchers/comments/worklogs · trash day 29/30/31 · board rank · saved filters · sprint scope/carry-over/poker · native suite/run caps · shared steps · data-driven · version depth · env×case matrix · flaky window · AS_ID collision  
> **Связи:** US-M03/M11/M12, US-S01, US-M05/S13/C02 · NFR-08/11/12/14/24 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*  
> **Канон каталога:** `.planning/july-2026/EDGE-CORNER-CATALOG.md` — **не изменять** этим файлом

### Gap-легенда (для этого файла)

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated edge-теста / TC |
| **Partial** | Happy/unit есть, граница не закрыта |
| **Future** | Лимит/фича ещё не зафиксированы в коде/ADR |

### Отличие от wave1 / wave2

| Тема | Wave1 (`*100+`) | Wave2 (`*200+`) | Wave3 (`*300+`) — новое углубление |
|------|-----------------|-----------------|-------------------------------------|
| Tracker caps | WIP, JQL len, labels, attach size, pagination | CF types, epic depth, attach count combo | **links/watchers/comments/worklogs**, trash day, rank, saved filters |
| Sprint | capacity, same-day, empty backlog, velocity zeros | TZ/DST/leap, burndown buckets | **scope mid-sprint**, carry-over chains, poker voters, goal length |
| TMS | Allure batch/BodyLimit, suite depth, BDD 0/max | Allure huge tree, BDD reorder/convert | **native suite/run size**, shared steps, data-driven, version depth, env matrix, flaky N, AS_ID |

---

## 1. Tracker — `ECC-TRCK-300+`

> Фокус: **issue links / watchers / comments / worklogs**, **trash retention day**, **board rank**, **saved filters**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-300 | Edge | Issue links outbound count = max−1 → add one | 201, count=max | API | P0 | TC-TRCK-06, ECC-TRCK-004 | TODO |
| ECC-TRCK-301 | Edge | Issue links outbound count = max → add one | 400 «link limit» | API | P0 | ECC-TRCK-016 | TODO |
| ECC-TRCK-302 | Edge | Issue links: both directions at max (A↔B dense graph) | each side ≤max; no silent truncate | API | P1 | — | TODO |
| ECC-TRCK-303 | Edge | Link type mix at max (blocks+relates+duplicates) | all types counted toward same cap or per-type — documented | API | P1 | TC-TRCK-06 | Future |
| ECC-TRCK-304 | Edge | Soft-delete target of max links → source list | dangling filtered or tombstone; GET source OK | API | P0 | NFR-11, TC-TRCK-12 | TODO |
| ECC-TRCK-305 | Edge | Watchers count = max−1 → add watcher | 200, count=max | API | P1 | US-M03 | Future |
| ECC-TRCK-306 | Edge | Watchers count = max → add one | 400 | API | P1 | — | Future |
| ECC-TRCK-307 | Edge | Watcher = assignee + reporter + self (dedupe) | unique set; no double notif fan-out | API | P2 | notif | TODO |
| ECC-TRCK-308 | Edge | Comments count = max−1 → post one | 201, count=max | API | P1 | TC-TRCK-05 | TODO |
| ECC-TRCK-309 | Edge | Comments count = max → post one | 400 | API | P1 | — | Future |
| ECC-TRCK-310 | Edge | Comment body length = max / max+1 | 201 / 400 | API | P1 | ECC-TRCK-138 sibling deepen | TODO |
| ECC-TRCK-311 | Edge | @mentions in one comment = max / max+1 | OK / 400; notif fan-out bounded | API | P0 | TC-TRCK-05, ECC-TRCK-005 | TODO |
| ECC-TRCK-312 | Edge | Worklogs on issue: count = max−1 → add | 201, time_spent sum correct | API | P1 | TC-TRCK-07, ECC-TRCK-008 | TODO |
| ECC-TRCK-313 | Edge | Worklogs count = max → add one | 400 | API | P1 | — | Future |
| ECC-TRCK-314 | Edge | Single worklog duration = 0 / 1s / max allowed | 0 reject or OK; max OK; max+1 → 400 | API | P0 | TC-TRCK-07 | TODO |
| ECC-TRCK-315 | Edge | Timer start when another timer already running (same user) | 400 or auto-stop previous — documented | API | P1 | TC-TRCK-08 | TODO |
| ECC-TRCK-316 | Edge | Timer running across midnight UTC | duration continuous; no day-split bug | API | P1 | NFR-12 | TODO |
| ECC-TRCK-317 | Edge | Trash: soft-delete day 0 → restore | 200; issue live | API | P1 | NFR-11, TC-TRCK-12 | Partial |
| ECC-TRCK-318 | Edge | Trash: day = 29 (retention 30) → restore | 200 | API | P0 | NFR-11 | TODO |
| ECC-TRCK-319 | Edge | Trash: day = 30 exact boundary → restore or purge | policy documented; no silent 500 | API | P0 | NFR-11 | TODO |
| ECC-TRCK-320 | Edge | Trash: day = 31 → restore | 404/410; hard-purged | API | P0 | NFR-11 | TODO |
| ECC-TRCK-321 | Edge | Board rank: move issue to position 0 (top) at full column | 200; order stable | API | P1 | US-M11, KANB | TODO |
| ECC-TRCK-322 | Edge | Board rank: move to last index = count−1 at max column size | 200; no off-by-one drop | API | P1 | ECC-TRCK-032 | TODO |
| ECC-TRCK-323 | Edge | Board rank: fractional/lexorank at dense board (100+ reorders) | no rank collision; reorder still OK | API/Perf | P1 | — | TODO |
| ECC-TRCK-324 | Edge | Saved filters per user = max−1 → create | 201, count=max | API | P2 | US-C04, TC-TRCK-09 | Future |
| ECC-TRCK-325 | Edge | Saved filters = max → create one | 400 | API | P2 | — | Future |
| ECC-TRCK-326 | Edge | Saved filter JQL length at project max while filter-count at max | both gates independent | API | P2 | ECC-TRCK-106–108 | TODO |
| ECC-TRCK-327 | Edge | Changelog/history page: limit=max, offset at last event | partial page; no duplicate events | API | P2 | NFR-14 | TODO |
| ECC-TRCK-328 | Edge | Clone issue with max links + max watchers | clone policy: copy subset or 400 | API | P2 | — | Future |
| ECC-TRCK-329 | Edge | Move issue across projects when source has max worklogs | worklogs migrate or strip — documented; no orphan rows | API | P1 | — | Future |

**Tracker Edge в этом файле: 30**

---

## 2. Sprint — `ECC-SPR-300+`

> Фокус: **scope change mid-sprint**, **carry-over chains**, **planning poker voters**, **goal/name length**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-300 | Edge | Active sprint: add issue when issue-count = max−1 | 200, count=max | API | P0 | ECC-SPR-122 deepen | Future |
| ECC-SPR-301 | Edge | Active sprint: add issue when issue-count = max | 400 | API | P0 | — | Future |
| ECC-SPR-302 | Edge | Active sprint: remove last issue → count=0 mid-flight | sprint stays active; burndown flat OK | API | P1 | ECC-SPR-112 deepen | TODO |
| ECC-SPR-303 | Edge | Scope add mid-sprint: planned points jump over capacity | warn/allow; velocity later honest | API | P1 | US-S01, ECC-SPR-102 | Future |
| ECC-SPR-304 | Edge | Scope remove mid-sprint: done points > remaining planned | burndown no negative crash; values honest | API | P1 | ECC-SPR-119 deepen | TODO |
| ECC-SPR-305 | Edge | Carry-over: complete → all incomplete to next planning (chain length=1) | next has exact set | API | P0 | TC-SPRINT-07 | Partial |
| ECC-SPR-306 | Edge | Carry-over chain: N consecutive completes (N=max history window) | no duplicate membership; backlog consistent | API | P1 | — | TODO |
| ECC-SPR-307 | Edge | Carry-over when next sprint already at max issues | overflow → backlog or 400 — documented | API | P0 | ECC-SPR-122 | Future |
| ECC-SPR-308 | Edge | Planning sprints open count = max−1 → create | 201 | API | P2 | — | Future |
| ECC-SPR-309 | Edge | Planning sprints open count = max → create | 400 | API | P2 | — | Future |
| ECC-SPR-310 | Edge | One active + max planning coexist | list OK; start second active still 400 | API | P1 | TC-SPRINT-03, ECC-SPR-008 | Partial |
| ECC-SPR-311 | Edge | Sprint goal length = max / max+1 | 201 / 400 | API | P2 | TC-SPRINT-01 | TODO |
| ECC-SPR-312 | Edge | Sprint name length = max / max+1 | 201 / 400 | API | P2 | — | TODO |
| ECC-SPR-313 | Edge | Planning Poker: voters = 0 at reveal | 400 or empty reveal policy | Manual/E2E | P2 | TC-SPRINT-06, ECC-SPR-124 | TODO |
| ECC-SPR-314 | Edge | Planning Poker: voters = max−1 → join one | reveal allowed at max | Manual/E2E | P2 | — | Future |
| ECC-SPR-315 | Edge | Planning Poker: voters = max → join one | 400 | Manual/E2E | P2 | — | Future |
| ECC-SPR-316 | Edge | Poker vote value at Fibonacci edge (0 / ? / max card) | accepted set only; unknown → 400 | API | P1 | TC-SPRINT-06 | TODO |
| ECC-SPR-317 | Edge | Complete sprint with exactly 1 incomplete + max done | move only incomplete; stats sum = total | API | P1 | TC-SPRINT-07 | Partial |
| ECC-SPR-318 | Edge | Reopen completed sprint (if allowed) at day boundary after complete | 400 or restore policy; dates immutable if locked | API | P1 | ECC-SPR-020 | Future |
| ECC-SPR-319 | Edge | Sprint board: filter → 0 issues while sprint has max issues | 200 empty view; sprint membership unchanged | API/E2E | P2 | ECC-TRCK-128 sibling | TODO |
| ECC-SPR-320 | Edge | Add same issue to two planning sprints (if multi-sprint backlog) | 400 exclusive or multi-membership policy | API | P1 | — | Future |
| ECC-SPR-321 | Edge | Velocity window = max completed sprints returned | series length=max; max+1 oldest dropped or paginated | API | P1 | TC-SPRINT-05, ECC-SPR-116 | TODO |

**Sprint Edge в этом файле: 22**

---

## 3. TMS — `ECC-TMS-300+`

> Фокус: **native suite/run size**, **shared steps**, **data-driven rows**, **version depth**, **env×case matrix**, **flaky window**, **AS_ID collision**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-300 | Edge | Native suite: direct child cases = max−1 → add | 201, count=max | API | P0 | ECC-TMS-111 deepen | TODO |
| ECC-TMS-301 | Edge | Native suite: child cases = max → add | 400 | API | P0 | — | Future |
| ECC-TMS-302 | Edge | Move case into suite already at max children | 400; source unchanged | API | P1 | ECC-TMS-112 | TODO |
| ECC-TMS-303 | Edge | Test run: cases = max−1 → add one | 200, count=max | API | P0 | TC-TMS-05, ECC-TMS-113 | TODO |
| ECC-TMS-304 | Edge | Test run: cases = max → add one | 400 | API | P0 | — | Future |
| ECC-TMS-305 | Edge | Execute all results in run at case-count=max | stats sum=max; duration bounded | API/Perf | P1 | TC-TMS-06 | TODO |
| ECC-TMS-306 | Edge | Shared step library: refs from cases = max−1 → link | 200 | API | P1 | US-S13 | Future |
| ECC-TMS-307 | Edge | Shared step refs = max → link one more | 400 | API | P1 | — | Future |
| ECC-TMS-308 | Edge | Delete shared step while max cases reference it | block or expand-inline policy | API | P0 | — | Future |
| ECC-TMS-309 | Edge | Data-driven: parameter rows = max−1 → add | 200, count=max | API | P1 | — | Future |
| ECC-TMS-310 | Edge | Data-driven: rows = max → add | 400 | API | P1 | — | Future |
| ECC-TMS-311 | Edge | Data-driven: columns = max / max+1 | OK / 400 | API | P2 | — | Future |
| ECC-TMS-312 | Edge | Case version history depth = max−1 → update | version=max; history retained | API | P0 | TC-TMS-03, NFR-24 | TODO |
| ECC-TMS-313 | Edge | Case version depth = max → update | 400 or drop oldest — documented | API | P0 | TC-TMS-04 | Future |
| ECC-TMS-314 | Edge | Rollback to version at depth=max (oldest kept) | 200 restore; version++ | API | P1 | TC-TMS-04, ECC-TMS-129 | Partial |
| ECC-TMS-315 | Edge | Env×case matrix: envs=max, cases=max in one run | create OK or 400 product; no OOM | Perf/API | P1 | ECC-TMS-127, US-S13 | TODO |
| ECC-TMS-316 | Edge | Env×case: envs=max+1 on run | 400 | API | P2 | — | Future |
| ECC-TMS-317 | Edge | Flaky window: exactly N fails in last W runs (W=window) | flagged flaky at boundary | API | P1 | ECC-TMS-124, US-C02 | TODO |
| ECC-TMS-318 | Edge | Flaky window: N−1 fails in W (just under) | not flaky | API | P1 | ECC-TMS-025 | TODO |
| ECC-TMS-319 | Edge | Flaky window: W=1 (min) single fail | policy: flaky or not — documented | API | P2 | — | TODO |
| ECC-TMS-320 | Edge | AS_ID collision: two cases same AS_ID in project | 400 on second or match-ambiguity report | API | P0 | TC-ALLURE-05, TC-TMS-07 | TODO |
| ECC-TMS-321 | Edge | AS_ID empty string vs null vs whitespace | normalize reject; no false match | API | P1 | — | TODO |
| ECC-TMS-322 | Edge | Traceability: case↔issue links = max / max+1 | OK / 400 | API | P2 | TC-TMS-10 | Future |
| ECC-TMS-323 | Edge | Export CSV/PDF of run with case-count=max | streams; no truncation silent; <timeout | Perf | P1 | TC-TMS-12, ECC-TMS-019 | Partial |
| ECC-TMS-324 | Edge | Soft-delete case day=29/30/31 (retention 30) | restore OK / boundary / 410 — same NFR-11 | API | P1 | NFR-11, ECC-TMS-018 | TODO |
| ECC-TMS-325 | Edge | Concurrent result posts = max workers on same run case | last-write or 409; stats consistent | API | P0 | NFR-24, ECC-TMS-016 | TODO |

**TMS Edge в этом файле: 26**

---

## Сводка счётчиков

| Домен | ID range | Edge count | Gap TODO | Gap Partial | Gap Future |
|-------|----------|------------|----------|-------------|------------|
| Tracker | ECC-TRCK-300…329 | **30** | 18 | 1 | 11 |
| Sprint | ECC-SPR-300…321 | **22** | 8 | 3 | 11 |
| TMS | ECC-TMS-300…325 | **26** | 12 | 2 | 12 |
| **ИТОГО** | | **78** | **38** | **6** | **34** |

| Метрика | Значение |
|---------|----------|
| Тип сценариев | **100% Edge** (Happy/Neg/Corner = 0) |
| Минимум задачи | ≥50 |
| Факт | **78** |
| Пересечение с wave1 `*100+` | **нет** (новые темы, не 1:1) |
| Пересечение с wave2 `*200+` | **нет** (новые темы, не 1:1) |
| Пересечение с каталогом 001–040 | **нет** |
| Изменения `EDGE-CORNER-CATALOG.md` | **нет** |
| Коммит | **не создавался** |

### Фокус-покрытие (wave3)

| Фокус | ID (примеры) | Count ≈ |
|-------|--------------|--------:|
| Issue links / watchers / comments / worklogs | ECC-TRCK-300…316, 328–329 | 19 |
| Trash day 29/30/31 | ECC-TRCK-317…320; ECC-TMS-324 | 5 |
| Board rank / saved filters / changelog | ECC-TRCK-321…327 | 7 |
| Sprint scope / carry-over / multi-planning | ECC-SPR-300…310, 317–321 | 15 |
| Poker / goal·name length | ECC-SPR-311…316 | 6 |
| Native suite / run size | ECC-TMS-300…305 | 6 |
| Shared steps / data-driven | ECC-TMS-306…311 | 6 |
| Version depth / env matrix / flaky / AS_ID | ECC-TMS-312…325 | 14 |

### Приоритеты (P)

| P | Count |
|---|------:|
| P0 | 22 |
| P1 | 35 |
| P2 | 21 |

### Рекомендуемый next step

1. `qa-manual-ver` — верификация wave3 expansion (границы, нет дублей wave1/wave2, трассируемость).  
2. P0: TRCK links/mentions/trash-30, SPR scope/carry-over max, TMS suite/run/version/AS_ID → полные `TC-*` в `docs/TESTING.md`.  
3. После verified → `qa-auto1` (API caps/trash/AS_ID) + `qa-auto2` (poker/board rank E2E).  
4. Каталог не трогать; импорт в канон — отдельным PR позже.

---

*Создано: 2026-07-10 · qa-manual · wave3 · только Edge · domains Tracker + Sprint + TMS*
