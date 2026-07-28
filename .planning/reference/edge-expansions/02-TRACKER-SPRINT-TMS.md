# Edge Expansion 02 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Edge** (границы лимитов, пустые/полные коллекции, max size, pagination, date/capacity)  
> **ID ranges:** `ECC-TRCK-100+`, `ECC-SPR-100+`, `ECC-TMS-100+` — **не пересекают** канон `EDGE-CORNER-CATALOG.md` (001–040)  
> **Не заменяет** каталог и `docs/TESTING.md` — дополняет edge-глубину для July Rebuild / QA-HARDENING  
> **Связи:** US-M03/M11/M12, US-S01, US-M05/S13/C02 · NFR-12/14/24 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*, TC-JQL-*  
> **Канон каталога:** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — **не изменять** этим файлом

### Gap-легенда (для этого файла)

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated edge-теста / TC |
| **Partial** | Happy/unit есть, граница не закрыта |
| **Future** | Лимит/фича ещё не зафиксированы в коде/ADR |

---

## 1. Tracker — `ECC-TRCK-100+`

> Фокус: WIP limits, JQL length, custom fields max, labels, attachments count, board empty/full, key overflow, pagination

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-100 | Edge | WIP=1, в колонке уже 1 issue → move второй | 400/409 «WIP-лимит достигнут» | API | P1 | US-M11, ECC-TRCK-018, TC-TRCK-10 | Partial |
| ECC-TRCK-101 | Edge | WIP=N, в колонке ровно N−1 → move OK | 200, count=N | API | P1 | KANB-02, TestWIPLimit_* | Partial |
| ECC-TRCK-102 | Edge | WIP=N, в колонке ровно N → move blocked | 400/409 | API | P0 | ECC-TRCK-018 | Partial |
| ECC-TRCK-103 | Edge | WIP=0 (без лимита) → неограниченный move | 200 всегда | API | P1 | schema WipLimit | Partial |
| ECC-TRCK-104 | Edge | WIP=1, create issue сразу в колонке с 1 issue | create reject или warn | API | P1 | US-M11 | TODO |
| ECC-TRCK-105 | Edge | Смена WIP колонки с N→M при count>M | block или warn+allow existing | API | P1 | workflow edit | TODO |
| ECC-TRCK-106 | Edge | JQL query length = max−1 rune | 200, parse OK | API | P1 | TC-JQL-01, NFR-14 | TODO |
| ECC-TRCK-107 | Edge | JQL query length = max | 200 или 400 по политике | API | P1 | TC-TRCK-09 | TODO |
| ECC-TRCK-108 | Edge | JQL query length = max+1 | 400 clear error | API | P1 | ECC-TRCK-017 | TODO |
| ECC-TRCK-109 | Edge | JQL с 50+ AND/OR условиями | 200 или 400 complexity limit | API | P2 | TC-JQL-01 | TODO |
| ECC-TRCK-110 | Edge | JQL IN-list на границе max IDs | OK at max; reject max+1 | API | P2 | TestToEntPredicatesIN | Partial |
| ECC-TRCK-111 | Edge | Custom fields: ровно max полей на project | 201 last; next 400 | API | P1 | ECC-TRCK-030 | TODO |
| ECC-TRCK-112 | Edge | Custom field value string length = max | 200 | API | P2 | custom-fields API | TODO |
| ECC-TRCK-113 | Edge | Custom field value string length = max+1 | 400 | API | P2 | — | TODO |
| ECC-TRCK-114 | Edge | Custom field select: options count = max | OK; max+1 reject | API | P2 | — | Future |
| ECC-TRCK-115 | Edge | Issue с заполненными всеми custom fields | GET/PUT OK, payload intact | API | P1 | ECC-TRCK-030 | Partial |
| ECC-TRCK-116 | Edge | Labels: 0 labels на issue | 200, empty array | API | P2 | ECC-TRCK-021 | TODO |
| ECC-TRCK-117 | Edge | Labels: ровно max labels | 200 | API | P2 | ECC-TRCK-021 | TODO |
| ECC-TRCK-118 | Edge | Labels: max+1 | 400 | API | P2 | ECC-TRCK-021 | TODO |
| ECC-TRCK-119 | Edge | Label name length = max / max+1 | OK / 400 | API | P2 | — | TODO |
| ECC-TRCK-120 | Edge | Label с unicode/emoji (rune length) | count by runes; reject overflow | API | P2 | — | TODO |
| ECC-TRCK-121 | Edge | Attachments: 0 файлов на issue | 200 empty list | API | P2 | storage | TODO |
| ECC-TRCK-122 | Edge | Attachments: count = max per issue | OK; max+1 reject | API | P1 | NFR-08 | TODO |
| ECC-TRCK-123 | Edge | Attachment ровно 25 MB (BodyLimit) | 200/presigned OK | API | P0 | NFR BodyLimit, NF-09 | Partial |
| ECC-TRCK-124 | Edge | Attachment 25 MB + 1 byte | 413/400 | API | P0 | Fiber BodyLimit | Partial |
| ECC-TRCK-125 | Edge | Board: 0 issues, 0 columns populated | 200 empty columns | API/E2E | P2 | ECC-TRCK-019 | TODO |
| ECC-TRCK-126 | Edge | Board: одна колонка full (WIP=N, count=N) | UI/API показывает limit reached | E2E | P1 | US-M11 | TODO |
| ECC-TRCK-127 | Edge | Board: 100+ issues (full board) | load OK; scroll/virtualize | Manual/Perf | P2 | ECC-TRCK-032 | TODO |
| ECC-TRCK-128 | Edge | Board filter → 0 matches | 200 empty subset | API/E2E | P2 | ECC-TRCK-023 | TODO |
| ECC-TRCK-129 | Edge | Issue key на границе большого N (KEY-999999) | create KEY-1000000 unique | API | P1 | ECC-TRCK-022 | TODO |
| ECC-TRCK-130 | Edge | Project key prefix max length + high seq | key string fits DB/UI | API | P2 | NFR-25 | TODO |
| ECC-TRCK-131 | Edge | Pagination: offset=0, limit=1 | ровно 1 item | API | P1 | NFR-14, ECC-TRCK-020 | Partial |
| ECC-TRCK-132 | Edge | Pagination: limit=max allowed | 200, ≤max items | API | P1 | NFR-14 | Partial |
| ECC-TRCK-133 | Edge | Pagination: limit=max+1 | 400 или clamp to max | API | P1 | NFR-14 | TODO |
| ECC-TRCK-134 | Edge | Pagination: offset ≥ total | 200, empty items | API | P1 | ECC-TRCK-020 | Partial |
| ECC-TRCK-135 | Edge | Pagination: offset+limit straddles last page | partial last page | API | P1 | NFR-14 | Partial |
| ECC-TRCK-136 | Edge | Pagination: limit=0 | 400 или empty policy | API | P2 | — | TODO |
| ECC-TRCK-137 | Edge | Checklist items: max / max+1 | OK / 400 | API | P2 | ECC-TRCK-021 | TODO |
| ECC-TRCK-138 | Edge | Description / comment body at max length | 200; max+1 → 400 | API | P2 | — | TODO |
| ECC-TRCK-139 | Edge | Estimate = 0 (boundary) | 200 accepted | API | P2 | ECC-TRCK-037 | TODO |
| ECC-TRCK-140 | Edge | Estimate = MaxInt / huge | 400 | API | P2 | ECC-TRCK-037 | TODO |

**Tracker Edge в этом файле: 41**

---

## 2. Sprint — `ECC-SPR-100+`

> Фокус: capacity, date boundaries, empty backlog, velocity edge

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-100 | Edge | Capacity=0 story points → start sprint | allow + warn или 400 | API | P2 | US-S01 | Future |
| ECC-SPR-101 | Edge | Planned points = capacity exactly | start OK | API | P1 | capacity | Future |
| ECC-SPR-102 | Edge | Planned points = capacity+1 | warn/block per policy | API | P1 | — | Future |
| ECC-SPR-103 | Edge | Capacity unset (null) → no enforcement | start OK | API | P2 | — | Future |
| ECC-SPR-104 | Edge | Start date = end date (same calendar day UTC) | 400 или 1-day sprint OK | API | P1 | ECC-SPR-012, TC-SPRINT-01 | TODO |
| ECC-SPR-105 | Edge | Start date = end date − 1s (sub-day) | 400 or normalize to days | API | P2 | NFR-12 | TODO |
| ECC-SPR-106 | Edge | Sprint length = 1 day (min) | 201 | API | P2 | ECC-SPR-012 | TODO |
| ECC-SPR-107 | Edge | Sprint length = max allowed weeks | 201; max+1 → 400 | API | P2 | — | TODO |
| ECC-SPR-108 | Edge | Start at UTC midnight / end at 23:59:59 UTC | burndown day buckets correct | API | P1 | ECC-SPR-018, NFR-12 | TODO |
| ECC-SPR-109 | Edge | Dates crossing DST / TZ boundary (user TZ ≠ UTC) | storage UTC; UI local OK | API/Manual | P1 | NFR-12 | TODO |
| ECC-SPR-110 | Edge | Empty backlog (0 issues) → create planning sprint | 201 | API | P2 | ECC-SPR-011 | TODO |
| ECC-SPR-111 | Edge | Empty backlog → start sprint | allow/warn | API | P2 | ECC-SPR-011, TC-SPRINT-02 | TODO |
| ECC-SPR-112 | Edge | Backlog empty after remove-all in planning | sprint stays planning, 0 issues | API | P2 | ECC-SPR-019 | Partial |
| ECC-SPR-113 | Edge | Complete sprint with 0 incomplete (all done) | completed, nothing to move | API | P1 | TC-SPRINT-07 | Partial |
| ECC-SPR-114 | Edge | Complete sprint with all incomplete | all moved to backlog/next | API | P1 | TC-SPRINT-07 | Partial |
| ECC-SPR-115 | Edge | Velocity: 0 completed sprints | 200 empty series / zeros | API | P1 | TC-SPRINT-05 | TODO |
| ECC-SPR-116 | Edge | Velocity: ровно 1 completed sprint | 200, single point | API | P1 | TestGetVelocity | Partial |
| ECC-SPR-117 | Edge | Velocity: sprint with planned=0, done=0 | zeros, no NaN/Inf | API | P1 | — | TODO |
| ECC-SPR-118 | Edge | Velocity: planned>0, done=0 (failed sprint) | ratio 0, chart OK | API | P1 | US-S01 | TODO |
| ECC-SPR-119 | Edge | Velocity: done > planned (scope added mid-sprint) | values honest, no crash | API | P2 | — | TODO |
| ECC-SPR-120 | Edge | Burndown day 0 (start) vs last day (end) | points at boundaries present | API | P1 | TC-SPRINT-04, ECC-SPR-018 | Partial |
| ECC-SPR-121 | Edge | Burndown with 0 estimate issues only | flat/zero line, 200 | API | P2 | — | TODO |
| ECC-SPR-122 | Edge | Max issues in one sprint (capacity by count) | OK at max; reject over | API | P2 | — | Future |
| ECC-SPR-123 | Edge | WIP on sprint board at column limit | enforce same as Kanban | API | P1 | ECC-SPR-013 | Partial |
| ECC-SPR-124 | Edge | Planning Poker: 0 voters / 1 voter reveal | reveal OK or policy | Manual/E2E | P2 | TC-SPRINT-06 | TODO |

**Sprint Edge в этом файле: 25**

---

## 3. TMS — `ECC-TMS-100+`

> Фокус: Allure batch size, suite depth, run with 0 cases, BDD step limits, import huge JSON

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-100 | Edge | Allure import batch = 0 results | 400 or empty run policy | API | P2 | TC-ALLURE-02 | Partial |
| ECC-TMS-101 | Edge | Allure import batch = 1 result | 200, matched/unmatched | API | P1 | TC-TMS-07 | Partial |
| ECC-TMS-102 | Edge | Allure import batch = 1000 results | <5s, 200 | Perf/API | P2 | ECC-TMS-014, perf | TODO |
| ECC-TMS-103 | Edge | Allure import batch = max+1 | 400/413 batch limit | API | P1 | — | TODO |
| ECC-TMS-104 | Edge | Allure multipart part size = BodyLimit 25 MB | OK | API | P0 | BodyLimit | Partial |
| ECC-TMS-105 | Edge | Allure JSON file 25 MB + 1 byte | 413/400 | API | P0 | Fiber BodyLimit | Partial |
| ECC-TMS-106 | Edge | Allure huge JSON: 10k results streaming | no OOM; 200 or 413 | Perf | P1 | ECC-TMS-014 | TODO |
| ECC-TMS-107 | Edge | Allure empty JSON array `[]` | 400 or empty import report | API | P2 | TC-ALLURE-02 | Partial |
| ECC-TMS-108 | Edge | Suite hierarchy depth = 1 (root only) | CRUD OK | API | P2 | ECC-TMS-020 | Partial |
| ECC-TMS-109 | Edge | Suite hierarchy depth = max | 201 | API | P1 | US-S13 | TODO |
| ECC-TMS-110 | Edge | Suite hierarchy depth = max+1 | 400 | API | P1 | — | TODO |
| ECC-TMS-111 | Edge | Suite with 0 child cases | list empty, delete OK | API | P2 | ECC-TMS-020 | TODO |
| ECC-TMS-112 | Edge | Move case into deepest suite | 200, path intact | API | P2 | — | TODO |
| ECC-TMS-113 | Edge | Create run with 0 cases | 201 empty run or 400 | API | P1 | TC-TMS-05, ECC-TMS-013 | TODO |
| ECC-TMS-114 | Edge | Start/execute run with 0 cases | 400 or completed 0/0 | API | P1 | TC-TMS-06 | TODO |
| ECC-TMS-115 | Edge | Add cases then remove all → 0 left mid-run | stats 0; no panic | API | P1 | — | TODO |
| ECC-TMS-116 | Edge | BDD steps count = 0 | 400 or allow empty | API | P2 | ECC-TMS-013, TC-TMS-02 | TODO |
| ECC-TMS-117 | Edge | BDD steps count = 1 (min useful) | 201 | API | P2 | TC-TMS-02 | Partial |
| ECC-TMS-118 | Edge | BDD steps count = max | 201 | API | P1 | US-S13 | TODO |
| ECC-TMS-119 | Edge | BDD steps count = max+1 | 400 | API | P1 | — | TODO |
| ECC-TMS-120 | Edge | BDD step text length = max / max+1 | OK / 400 | API | P2 | — | TODO |
| ECC-TMS-121 | Edge | Classic steps count = max / max+1 | OK / 400 | API | P2 | ECC-TMS-013 | TODO |
| ECC-TMS-122 | Edge | Import huge TestRail CSV/XML = 25 MB | OK; +1 byte reject | API | P1 | imex BodyLimit | Partial |
| ECC-TMS-123 | Edge | Import JSON with nested suites depth=max | tree built; max+1 reject | API | P1 | ECC-TMS-020 | TODO |
| ECC-TMS-124 | Edge | Flaky boundary: exactly 1 fail + 1 pass | listed as flaky | API | P2 | ECC-TMS-025, US-C02 | TODO |
| ECC-TMS-125 | Edge | Pagination runs: offset/limit boundaries | same as NFR-14 | API | P2 | ECC-TMS-030 | Partial |
| ECC-TMS-126 | Edge | Case key high sequence (PROJ-TC-999999) | next unique | API | P1 | ECC-TMS-015 | TODO |
| ECC-TMS-127 | Edge | Environments matrix: 0 envs vs max envs | filter empty / full OK | API | P2 | ECC-TMS-022, US-S13 | Partial |
| ECC-TMS-128 | Edge | Export PDF/CSV on run with 0 results | empty file or 400 | API | P2 | TC-TMS-12, ECC-TMS-019 | Partial |
| ECC-TMS-129 | Edge | Rollback when case has exactly 1 version | 400/noop policy | API | P2 | TC-TMS-04 | Partial |
| ECC-TMS-130 | Edge | Webhook JUnit body size at BodyLimit | OK; over → 413 | API | P0 | TC-TMS-08 | TODO |

**TMS Edge в этом файле: 31**

---

## Сводка счётчиков

| Домен | ID range | Edge count | Gap TODO | Gap Partial | Gap Future |
|-------|----------|------------|----------|-------------|------------|
| Tracker | ECC-TRCK-100…140 | **41** | 28 | 12 | 1 |
| Sprint | ECC-SPR-100…124 | **25** | 14 | 6 | 5 |
| TMS | ECC-TMS-100…130 | **31** | 19 | 12 | 0 |
| **ИТОГО** | | **97** | **61** | **30** | **6** |

| Метрика | Значение |
|---------|----------|
| Тип сценариев | **100% Edge** (Happy/Neg/Corner = 0) |
| Минимум задачи | ≥45 |
| Факт | **97** |
| Пересечение с каталогом 001–040 | **нет** |
| Изменения `EDGE-CORNER-CATALOG.md` | **нет** |
| Коммит | **не создавался** |

### Приоритеты (P)

| P | Count |
|---|------:|
| P0 | 6 |
| P1 | 44 |
| P2 | 47 |

### Рекомендуемый next step

1. `qa-manual-ver` — верификация этого expansion (техники границ, трассируемость US/FR).  
2. Выбрать P0/P1 → оформить полные `TC-*` в `docs/TESTING.md`.  
3. После verified → `qa-auto1` (API: WIP/JQL/pagination/Allure batch) + `qa-auto2` (board empty/full E2E).  
4. По закрытии Gap — опционально импорт строк в канон-каталог отдельным PR (не в этой задаче).

---

*Создано: 2026-07-10 · qa-manual · только Edge · domains Tracker + Sprint + TMS*
