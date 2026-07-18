# Edge Expansion Wave2 E02 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Edge** (границы лимитов, depth, TZ/date buckets, huge collections, count caps)  
> **ID ranges:** `ECC-TRCK-200+`, `ECC-SPR-200+`, `ECC-TMS-200+` — **не пересекают** wave1 `*100+`, канон `001–040`, corner `*C200+` / `*C300+`  
> **Не заменяет** каталог и `docs/TESTING.md` — углубление edge для July Rebuild / QA-HARDENING  
> **Wave1 (не дублировать 1:1):** `.planning/edge-expansions/02-TRACKER-SPRINT-TMS.md` (`ECC-*-100…140/124/130`)  
> **Фокус wave2:** custom field limits · epic hierarchy depth · sprint date TZ · Allure huge suite · BDD step count · attachment count per issue  
> **Связи:** US-M03/M11/M12, US-S01, US-M05/S13/C02 · NFR-08/12/14 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*  
> **Канон каталога:** `.planning/july-2026/EDGE-CORNER-CATALOG.md` — **не изменять** этим файлом

### Gap-легенда (для этого файла)

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated edge-теста / TC |
| **Partial** | Happy/unit есть, граница не закрыта |
| **Future** | Лимит/фича ещё не зафиксированы в коде/ADR |

### Отличие от wave1

| Тема | Wave1 (`*100+`) | Wave2 (`*200+`) — углубление |
|------|-----------------|------------------------------|
| Custom fields | max полей / string length / select options | типы, multi-value, required-on-transition, bulk fill, unicode/null |
| Epic hierarchy | почти не покрыто | depth min/max, orphan, cycle, move across levels |
| Sprint dates/TZ | same-day, DST one-liner | half-open intervals, TZ±14, leap day, burndown bucket edges |
| Attachments | 0 / max / 25 MB size | count+size combo, MIME mix, soft-delete restore, concurrent at cap |
| Allure | batch size / BodyLimit | suite tree huge, nested depth, attachment fan-out, match storm |
| BDD steps | 0 / 1 / max count | keyword mix, nested tables, reorder at max, classic↔BDD convert |

---

## 1. Tracker — `ECC-TRCK-200+`

> Фокус: **custom field limits**, **epic hierarchy depth**, **attachment count per issue**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-200 | Edge | Custom field definitions: count = max−1 → add one | 201, count=max | API | P1 | US-M03, ECC-TRCK-111 | TODO |
| ECC-TRCK-201 | Edge | Custom field definitions: count = max → add one | 400 «limit reached» | API | P0 | ECC-TRCK-030 | TODO |
| ECC-TRCK-202 | Edge | Custom field type=number: value = MaxInt64 | 200 or 400 per schema | API | P2 | custom-fields | Future |
| ECC-TRCK-203 | Edge | Custom field type=number: value = MaxInt64+1 / NaN / Inf | 400 | API | P1 | — | TODO |
| ECC-TRCK-204 | Edge | Custom field type=date: value at UTC midnight boundary | stored UTC; GET round-trip OK | API | P1 | NFR-12 | TODO |
| ECC-TRCK-205 | Edge | Custom field type=user: assignee not in project | 400 | API | P1 | RBAC | TODO |
| ECC-TRCK-206 | Edge | Custom field multi-select: selected = max options | 200; max+1 → 400 | API | P1 | ECC-TRCK-114 | Future |
| ECC-TRCK-207 | Edge | Custom field multi-select: duplicate option IDs in payload | 400 or dedupe documented | API | P2 | — | TODO |
| ECC-TRCK-208 | Edge | Custom field string: empty `""` vs null vs omitted | policy consistent; no silent coerce | API | P1 | — | TODO |
| ECC-TRCK-209 | Edge | Custom field string: unicode combining marks at max runes | count by runes; reject overflow | API | P2 | ECC-TRCK-120 | TODO |
| ECC-TRCK-210 | Edge | Required custom field empty on create | 400 with field key | API | P0 | workflow | TODO |
| ECC-TRCK-211 | Edge | Required custom field empty on transition to Done | 400; status unchanged | API | P0 | TC-TRCK-10 | TODO |
| ECC-TRCK-212 | Edge | Issue with all custom fields at max string length each | PUT/GET OK; payload size within BodyLimit | API | P1 | ECC-TRCK-115, NFR-08 | TODO |
| ECC-TRCK-213 | Edge | Bulk edit: set one CF on max issues in one request | OK at batch max; max+1 → 400 | API | P2 | bulk | Future |
| ECC-TRCK-214 | Edge | Delete CF definition while issues hold values | cascade null / block / soft-orphan policy | API | P1 | ECC-TRCK-030 | TODO |
| ECC-TRCK-215 | Edge | Rename CF key at max name length / max+1 | OK / 400 | API | P2 | — | TODO |
| ECC-TRCK-216 | Edge | Epic hierarchy depth = 1 (epic only, no children) | CRUD OK; board shows epic | API | P2 | US-M03 | TODO |
| ECC-TRCK-217 | Edge | Epic → Story → Sub-task depth = max | 201 last level | API | P0 | ECC-TRCK-033 | TODO |
| ECC-TRCK-218 | Edge | Epic hierarchy depth = max+1 (extra nest) | 400 «max depth» | API | P0 | — | TODO |
| ECC-TRCK-219 | Edge | Create child under epic at depth=max−1 | 201, depth=max | API | P1 | — | TODO |
| ECC-TRCK-220 | Edge | Move story from epic A to epic B when B already at max children | OK or 400 per child-count cap | API | P1 | — | Future |
| ECC-TRCK-221 | Edge | Reparent that would create cycle (epic under own descendant) | 400 | API | P0 | ECC-TRCK-033 | TODO |
| ECC-TRCK-222 | Edge | Orphan story: parent epic soft-deleted | block delete or reparent; no dangling parent_id | API | P0 | TC-TRCK-12, NFR-11 | TODO |
| ECC-TRCK-223 | Edge | Epic with 0 children → convert type epic→story | 200; hierarchy links cleared | API | P2 | issue types | TODO |
| ECC-TRCK-224 | Edge | Epic with children → convert type epic→task | 400 or cascade policy | API | P1 | — | TODO |
| ECC-TRCK-225 | Edge | Max epics per project (if capped) | OK at max; max+1 → 400 | API | P2 | — | Future |
| ECC-TRCK-226 | Edge | JQL `parent = EPIC` returns all descendants at max depth | complete set; no truncation silent | API | P1 | TC-JQL-01, NF-03 | TODO |
| ECC-TRCK-227 | Edge | Attachments count = max−1 → upload one | 201, count=max | API | P0 | NFR-08, ECC-TRCK-122 | TODO |
| ECC-TRCK-228 | Edge | Attachments count = max → upload one more | 400/413 «attachment limit» | API | P0 | ECC-TRCK-122 | TODO |
| ECC-TRCK-229 | Edge | Attachments: max count × each file = 1 byte | all accepted; list length=max | API | P1 | storage | TODO |
| ECC-TRCK-230 | Edge | Attachments: max−1 files + one at BodyLimit 25 MB | OK if under count; size gate separate | API | P1 | NF-09, ECC-TRCK-123 | Partial |
| ECC-TRCK-231 | Edge | Attachments: delete one at count=max → re-upload | 201; count back to max | API | P1 | — | TODO |
| ECC-TRCK-232 | Edge | Soft-delete issue with max attachments → restore | attachments restored; count intact | API | P1 | NFR-11, TC-TRCK-12 | TODO |
| ECC-TRCK-233 | Edge | Same filename × max attachments (collision) | unique storage keys; UI disambiguates | API | P2 | SeaweedFS | TODO |
| ECC-TRCK-234 | Edge | Mixed MIME at max count (image/pdf/zip) | all listed; download each OK | API | P2 | — | TODO |
| ECC-TRCK-235 | Edge | Presigned PUT succeeds but register attachment when count already max | register 400; orphan GC | API | P0 | NF-09 | TODO |
| ECC-TRCK-236 | Edge | Attachment metadata name length = max / max+1 | OK / 400 | API | P2 | — | TODO |
| ECC-TRCK-237 | Edge | Copy/clone issue with max attachments | clone OK or attachment-cap policy | API | P2 | — | Future |
| ECC-TRCK-238 | Edge | Epic at max depth + max CF filled + max attachments | create/update OK; no OOM | Perf/API | P2 | NFR-14 | TODO |

**Tracker Edge в этом файле: 39**

---

## 2. Sprint — `ECC-SPR-200+`

> Фокус: **sprint date TZ** (UTC storage, local display, burndown day buckets)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-200 | Edge | Create sprint: start/end as RFC3339 with offset +03:00 | DB stores UTC; response UTC or documented | API | P0 | NFR-12, TC-SPRINT-01 | TODO |
| ECC-SPR-201 | Edge | Create sprint: start/end with offset −12:00 / +14:00 | accepted; UTC normalized | API | P1 | NFR-12 | TODO |
| ECC-SPR-202 | Edge | Start date local midnight (user TZ) maps to previous UTC day | burndown day-0 = UTC date of start | API/Manual | P0 | ECC-SPR-108/109 | TODO |
| ECC-SPR-203 | Edge | End date local 23:59:59.999 in TZ+14 | inclusive/exclusive policy documented; no missing last day | API | P0 | ECC-SPR-018 | TODO |
| ECC-SPR-204 | Edge | Half-open interval [start, end): end == start+1day UTC | exactly 1 burndown bucket | API | P1 | TC-SPRINT-04 | TODO |
| ECC-SPR-205 | Edge | Sprint spanning DST spring-forward (lose 1h) | day count by calendar dates UTC; no NaN velocity | API | P0 | NFR-12 | TODO |
| ECC-SPR-206 | Edge | Sprint spanning DST fall-back (gain 1h) | no duplicate burndown day key | API | P0 | — | TODO |
| ECC-SPR-207 | Edge | Leap day: start 2028-02-28 → end 2028-03-01 | Feb 29 bucket present if in range | API | P2 | — | TODO |
| ECC-SPR-208 | Edge | Year boundary: start Dec 31 22:00 UTC → end Jan 2 | buckets in correct years | API | P1 | — | TODO |
| ECC-SPR-209 | Edge | User TZ change mid-active sprint (profile PUT timezone) | charts re-render local; DB dates unchanged | Manual/E2E | P1 | NFR-12, /auth/me | TODO |
| ECC-SPR-210 | Edge | Two users different TZ view same sprint dates | same UTC truth; different local labels | Manual | P1 | US-S01 | TODO |
| ECC-SPR-211 | Edge | start_date with milliseconds / nanoseconds | truncate or reject consistently | API | P2 | — | TODO |
| ECC-SPR-212 | Edge | end_date before start after TZ normalize (offset trick) | 400 (not silent swap) | API | P0 | TestCreateSprint_EndDateBeforeStartDate | Partial |
| ECC-SPR-213 | Edge | Update dates of planning sprint to cross TZ midnight | 200; burndown regen on start | API | P1 | — | TODO |
| ECC-SPR-214 | Edge | Update dates of active sprint (shorten/lengthen) | allow+regen or 400 policy | API | P1 | — | Future |
| ECC-SPR-215 | Edge | Burndown: query at exact start instant vs start−1s | first point only at/after start | API | P1 | TC-SPRINT-04 | TODO |
| ECC-SPR-216 | Edge | Burndown: query at exact end instant vs end+1s | last point inclusive per policy | API | P1 | ECC-SPR-120 | TODO |
| ECC-SPR-217 | Edge | Complete sprint at local midnight ≠ UTC midnight | completed_at UTC; incomplete move OK | API | P1 | TC-SPRINT-07 | Partial |
| ECC-SPR-218 | Edge | Velocity chart: sprints with different TZ creators | series aligned by UTC end dates | API | P2 | TC-SPRINT-05 | TODO |
| ECC-SPR-219 | Edge | Max sprint length crossing many TZ midnights (e.g. 12 weeks) | bucket count = calendar days; no O(n²) blowup | Perf/API | P2 | ECC-SPR-107 | TODO |
| ECC-SPR-220 | Edge | start/end as date-only `YYYY-MM-DD` (no time) | interpret as UTC midnight or user TZ midnight — documented | API | P1 | — | Future |
| ECC-SPR-221 | Edge | Invalid TZ in client Accept-Language / profile while loading board | fallback UTC; 200 no crash | API/E2E | P2 | — | TODO |

**Sprint Edge в этом файле: 22**

---

## 3. TMS — `ECC-TMS-200+`

> Фокус: **Allure huge suite**, **BDD step count**

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-200 | Edge | Allure import: suite tree with depth = max | tree built; paths intact | API | P0 | US-S13, ECC-TMS-109 | TODO |
| ECC-TMS-201 | Edge | Allure import: suite tree depth = max+1 | 400 or truncate+warn documented | API | P0 | ECC-TMS-110 | TODO |
| ECC-TMS-202 | Edge | Allure: single suite with test count = max | 200; all matched/unmatched counted | API | P1 | ECC-TMS-102 | TODO |
| ECC-TMS-203 | Edge | Allure: single suite with test count = max+1 | 400/413 or partial+report | API | P1 | — | TODO |
| ECC-TMS-204 | Edge | Allure: total suites in file = max flat siblings | import OK; UI list paginated | API | P1 | NFR-14 | TODO |
| ECC-TMS-205 | Edge | Allure: total suites = max+1 | reject or soft-cap | API | P2 | — | Future |
| ECC-TMS-206 | Edge | Allure huge: 5k suites × 2 tests (wide tree) | no OOM; <timeout; 200 or 413 | Perf | P0 | ECC-TMS-106 | TODO |
| ECC-TMS-207 | Edge | Allure huge: 1 suite × 10k tests (deep list) | streaming parse; memory bounded | Perf | P0 | ECC-TMS-014 | TODO |
| ECC-TMS-208 | Edge | Allure: each result has max labels | import OK; match still works | API | P2 | TC-ALLURE-05 | TODO |
| ECC-TMS-209 | Edge | Allure: each result has max+1 labels | 400 or ignore extras | API | P2 | — | TODO |
| ECC-TMS-210 | Edge | Allure: attachment count per result = max | stored/linked; download OK | API | P1 | allure_reports | TODO |
| ECC-TMS-211 | Edge | Allure: attachment count per result = max+1 | reject result or drop extras | API | P1 | — | TODO |
| ECC-TMS-212 | Edge | Allure: nested suites with duplicate names different paths | distinct nodes; no merge collision | API | P1 | ECC-TMS-020 | TODO |
| ECC-TMS-213 | Edge | Allure: empty suite nodes (0 tests) in huge tree | kept or pruned per policy; no crash | API | P2 | ECC-TMS-111 | TODO |
| ECC-TMS-214 | Edge | Allure: status mix at scale (pass/fail/broken/skipped) | stats sum = total; no float drift | API | P1 | TC-ALLURE-04 | Partial |
| ECC-TMS-215 | Edge | Allure: match storm — all results have AS_ID | 100% matched; duration OK | API | P1 | TC-TMS-07, TC-ALLURE-05 | Partial |
| ECC-TMS-216 | Edge | Allure: match storm — 0 AS_ID, fuzzy only at huge N | unmatched report complete; no timeout | Perf | P2 | — | TODO |
| ECC-TMS-217 | Edge | Allure re-import same huge suite (idempotent) | no duplicate cases/runs or version++ policy | API | P1 | TC-ALLURE-01 | TODO |
| ECC-TMS-218 | Edge | BDD steps count = max−1 → append one | 200, count=max | API | P0 | TC-TMS-02, ECC-TMS-118 | TODO |
| ECC-TMS-219 | Edge | BDD steps count = max → append one | 400 | API | P0 | ECC-TMS-119 | TODO |
| ECC-TMS-220 | Edge | BDD at max steps: reorder only (no add) | 200, order persisted | API | P1 | — | TODO |
| ECC-TMS-221 | Edge | BDD at max steps: replace one step text | 200; version++ | API | P1 | TC-TMS-03 | Partial |
| ECC-TMS-222 | Edge | BDD: all steps keyword=Given (no When/Then) | 201 or 400 style policy | API | P2 | TC-TMS-02 | TODO |
| ECC-TMS-223 | Edge | BDD: alternating Given/When/Then × max/3 | 201 | API | P2 | US-S13 | TODO |
| ECC-TMS-224 | Edge | BDD step with data-table rows = max / max+1 | OK / 400 | API | P1 | — | Future |
| ECC-TMS-225 | Edge | BDD step with docstring length = max / max+1 | OK / 400 | API | P2 | ECC-TMS-120 | TODO |
| ECC-TMS-226 | Edge | BDD → classic convert at max BDD steps | mapped or 400 if classic max lower | API | P1 | ECC-TMS-121 | TODO |
| ECC-TMS-227 | Edge | Classic → BDD convert at max classic steps | 201 BDD or 400 | API | P1 | — | TODO |
| ECC-TMS-228 | Edge | BDD case at max steps in run with max cases | execute OK; result payload bounded | API/Perf | P2 | TC-TMS-05/06 | TODO |
| ECC-TMS-229 | Edge | Update BDD: delete steps down to 0 from max | 400 or allow empty per ECC-TMS-116 policy | API | P1 | ECC-TMS-116 | TODO |
| ECC-TMS-230 | Edge | BDD step keyword unknown / empty at position max | 400; no partial save | API | P1 | — | TODO |
| ECC-TMS-231 | Edge | Import TestRail/Allure that expands to BDD steps > max | reject file or per-case error report | API | P1 | ECC-TMS-122/123 | TODO |
| ECC-TMS-232 | Edge | Allure huge suite + create bugs from all failed (fan-out) | rate-limit/batch; no issue storm crash | API | P2 | TC-TMS-09 | TODO |

**TMS Edge в этом файле: 33**

---

## Сводка счётчиков

| Домен | ID range | Edge count | Gap TODO | Gap Partial | Gap Future |
|-------|----------|------------|----------|-------------|------------|
| Tracker | ECC-TRCK-200…238 | **39** | 32 | 1 | 6 |
| Sprint | ECC-SPR-200…221 | **22** | 18 | 2 | 2 |
| TMS | ECC-TMS-200…232 | **33** | 28 | 3 | 2 |
| **ИТОГО** | | **94** | **78** | **6** | **10** |

| Метрика | Значение |
|---------|----------|
| Тип сценариев | **100% Edge** (Happy/Neg/Corner = 0) |
| Минимум задачи | ≥50 |
| Факт | **94** |
| Пересечение с wave1 `*100+` | **нет** (углубление, не 1:1) |
| Пересечение с каталогом 001–040 | **нет** |
| Изменения `EDGE-CORNER-CATALOG.md` | **нет** |
| Коммит | **не создавался** |

### Фокус-покрытие (wave2)

| Фокус | ID (примеры) | Count ≈ |
|-------|--------------|--------:|
| Custom field limits | ECC-TRCK-200…215 | 16 |
| Epic hierarchy depth | ECC-TRCK-216…226, 238 | 12 |
| Attachment count per issue | ECC-TRCK-227…237 | 11 |
| Sprint date TZ | ECC-SPR-200…221 | 22 |
| Allure huge suite | ECC-TMS-200…217, 232 | 19 |
| BDD step count | ECC-TMS-218…231 | 14 |

### Приоритеты (P)

| P | Count |
|---|------:|
| P0 | 22 |
| P1 | 44 |
| P2 | 28 |

### Рекомендуемый next step

1. `qa-manual-ver` — верификация wave2 expansion (границы, нет дублей wave1, трассируемость).  
2. P0: TRCK attachment/epic depth, SPR TZ normalize, TMS Allure depth/BDD max → полные `TC-*` в `docs/TESTING.md`.  
3. После verified → `qa-auto1` (API limits/TZ/Allure) + `qa-auto2` (TZ display E2E).  
4. Каталог не трогать; импорт в канон — отдельным PR позже.

---

*Создано: 2026-07-10 · qa-manual · wave2 · только Edge · domains Tracker + Sprint + TMS*
