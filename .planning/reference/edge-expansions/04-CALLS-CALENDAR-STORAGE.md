# Edge Expansion 04 — Calls + Calendar + Storage

> **Тип:** только **Edge** (границы лимитов / TTL / размеров / дат / rate)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual`  
> **Не трогает** основной каталог `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`  
> **SFU:** LiveKit only (mediasoup / SEED-001 = REJECTED — строки N/A **не** добавлять)  
> **Связи:** NFR-03/04/08/12/27 · US-M06 · US-CAL-01…04 · US-V03 · TC-CALL-* · TC-STORAGE-* · JULY brief §3 CAL-SYNC · `backend/internal/{calls,calendar,storage}`

**Назначение:** foresight-бэклог граничных кейсов для будущих `TC-*`. Happy / Neg / Corner сюда не входят.

| Колонка | Значение |
|---------|----------|
| Gap | по умолчанию **TODO** (расширение ещё не в основном каталоге) |
| Слой | API / Unit / Manual / E2E — где проверять границу |
| Связь | NFR / US / TC / Brief / константа кода |

---

## 1. Calls + LiveKit — `ECC-CALL-100+`

> Источники границ: `defaultMaxCallParticipants=50` · member token `SetValidFor(24h)` · guest `4h` · `maxActiveRecordingsPerCall=2` · recording download TTL `1h` · RRULE scheduled · conflict check window · screen share alone (US-M06)

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-100 | Edge | 50-й участник join (ровно max) | 200, token; room accepts | API/Manual | P0 | NFR-03 | TODO |
| ECC-CALL-101 | Edge | 51-й участник join (max+1) | reject LiveKit / Hub warn; count≤50 | Manual | P0 | NFR-03, ECC-CALL-017 | TODO |
| ECC-CALL-102 | Edge | Одновременный join 49→50→51 (burst) | ровно один отказ на 51-м | Manual | P0 | NFR-03 | TODO |
| ECC-CALL-103 | Edge | Guest join при уже 50 members | guest rejected; members stay | API/Manual | P1 | NFR-03, TC-CALL-04 | TODO |
| ECC-CALL-104 | Edge | Длительность звонка ровно 24h (member token TTL) | до TTL media OK; после — reconnect/rejoin | Manual | P0 | token 24h | TODO |
| ECC-CALL-105 | Edge | Member LiveKit JWT на границе 24h−1s / 24h+1s | −1s valid; +1s invalid | Unit/API | P1 | SetValidFor(24h) | TODO |
| ECC-CALL-106 | Edge | Guest LiveKit JWT на границе 4h−1s / 4h+1s | −1s valid; +1s invalid | Unit/API | P1 | guest 4h, BUSINESS.md | TODO |
| ECC-CALL-107 | Edge | Guest token refresh/rejoin после 4h | новый guest token; старый мёртв | API/Manual | P1 | TC-CALL-04 | TODO |
| ECC-CALL-108 | Edge | Recording file size на лимите storage (25 MB BodyLimit path) | download via presign OK; metadata size exact | API | P1 | NFR-04, recording download | TODO |
| ECC-CALL-109 | Edge | Recording egress output > 25 MB (presigned path / NF-09) | файл доступен через S3 presign; не через Fiber buffer | Manual | P0 | NF-09, NFR-27 | TODO |
| ECC-CALL-110 | Edge | Recording download URL на границе TTL 1h−1s / 1h+1s | −1s 200; +1s S3 403 | Manual | P1 | downloadURLTTL=1h | TODO |
| ECC-CALL-111 | Edge | Max active recordings = 2; старт 3-й | 400/409 reject | API | P1 | maxActiveRecordingsPerCall=2 | TODO |
| ECC-CALL-112 | Edge | RRULE monthly: DTSTART 31-е → короткий месяц (Feb) | occurrence = last day of month | Unit | P1 | RRULE, US-CAL-02 analog | TODO |
| ECC-CALL-113 | Edge | RRULE `UNTIL` = DTSTART (одно вхождение) | ровно 1 occurrence; worker fires once | Unit/API | P1 | scheduled RRULE | TODO |
| ECC-CALL-114 | Edge | RRULE `COUNT=1` | одно вхождение; после — no next | Unit | P2 | recording_test RRULE | TODO |
| ECC-CALL-115 | Edge | RRULE leap-day `FREQ=YEARLY;BYMONTHDAY=29` в non-leap | skip или last-day policy documented | Unit | P1 | RRULE edge dates | TODO |
| ECC-CALL-116 | Edge | RRULE `UNTIL` в UTC vs local TZ boundary (23:59→00:00) | last occurrence correct in UTC store | Unit | P0 | NFR-12 | TODO |
| ECC-CALL-117 | Edge | Conflict window: ends_at == other.starts_at (смежные) | **нет** conflict (touching OK) | API | P1 | CheckConflicts | TODO |
| ECC-CALL-118 | Edge | Conflict window: overlap 1s | HasConflict=true | API | P1 | CheckConflicts | TODO |
| ECC-CALL-119 | Edge | Conflict window: identical [start,end) | HasConflict=true | API | P1 | CheckConflicts | TODO |
| ECC-CALL-120 | Edge | Conflict check: duration=0 / start==end | 400 invalid window | API | P2 | ECC-CALL-031 related | TODO |
| ECC-CALL-121 | Edge | Screen share alone (no mic publish, no webcam) | room OK; screen track visible to peers | Manual | P1 | US-M06, ECC-CALL-020 | TODO |
| ECC-CALL-122 | Edge | Screen share alone + recording start | egress captures screen; audio may be silent | Manual | P1 | TC-CALL-07 | TODO |
| ECC-CALL-123 | Edge | Только 1 participant + screen share (solo room) | share works; no peer required | Manual | P2 | US-M06 | TODO |
| ECC-CALL-124 | Edge | Call duration soft-cap / long-running >24h with token refresh | rejoin renews JWT; room persists until end | Manual | P1 | token TTL boundary | TODO |
| ECC-CALL-125 | Edge | Invite code join на границе call just-ended (status flip) | 400/404; no token leak | API | P1 | TC-CALL-03/06 | TODO |

**Calls Edge в этом файле: 26** (`ECC-CALL-100`…`125`)

---

## 2. Calendar as-is + Calendar v2 (future) — `ECC-CAL-100+`

> As-is: aggregator `maxRangeDays=366`, iCal export.  
> v2 future: ресурсы (US-CAL-01), recurring until/count (US-CAL-02), iCal feed size / LimitReader (US-CAL-03).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-100 | Edge | List range ровно 366 дней (to−from = 366d) | 200, events | API | P0 | maxRangeDays=366 | TODO |
| ECC-CAL-101 | Edge | List range 366d+1s / 367d | 400 `calendar range too wide` | API | P0 | ErrRangeTooWide, ECC-CAL-005 | TODO |
| ECC-CAL-102 | Edge | List range from==to (zero-width) | 400 invalid **или** 200 [] — зафиксировать контракт | API | P1 | validateRange | TODO |
| ECC-CAL-103 | Edge | List range 1ms (минимальный ненулевой) | 200 (возможно []) | API | P2 | validateRange | TODO |
| ECC-CAL-104 | Edge | Export.ics на диапазоне 366d с max событий | text/calendar; ответ завершается; без OOM | API/Manual | P1 | US-CAL-03, ExportICal | TODO |
| ECC-CAL-105 | Edge | Export.ics пустой диапазон (нет событий) | валидный VCALENDAR, 0 VEVENT | API | P2 | ECC-CAL-006 | TODO |
| ECC-CAL-106 | Edge | All-day event DTEND exclusive (RFC 5545) на границе дня | ICS end = next day; FE shows 1 day | Unit/API | P1 | ical.go all-day | TODO |
| ECC-CAL-107 | Edge | Event starts_at на DST spring-forward gap | stored UTC; local display no phantom hour | E2E/Manual | P0 | NFR-12 | TODO |
| ECC-CAL-108 | Edge | Event ends_at на DST fall-back overlap hour | single occurrence; no duplicate | E2E/Manual | P1 | NFR-12 | TODO |
| ECC-CAL-109 | Edge | v2: resource booking смежные интервалы (15:00 touch) | **не** 409 | API | P1 | US-CAL-01 edge | Future |
| ECC-CAL-110 | Edge | v2: resource booking overlap 1s | 409 conflict | API | P0 | US-CAL-01 | Future |
| ECC-CAL-111 | Edge | v2: recurring `COUNT=1` | ровно 1 instance | API | P1 | US-CAL-02 | Future |
| ECC-CAL-112 | Edge | v2: recurring `UNTIL` = start date | 1 instance или 400 — контракт | API | P1 | US-CAL-02 | Future |
| ECC-CAL-113 | Edge | v2: recurring `COUNT=0` / negative | 400 | API | P0 | US-CAL-02 AC | Future |
| ECC-CAL-114 | Edge | v2: recurring `UNTIL` в прошлом | 400 | API | P0 | US-CAL-02 AC | Future |
| ECC-CAL-115 | Edge | v2: monthly 31st → Feb (short month) | last day of month | Unit | P1 | US-CAL-02, ECC-CAL-013 | Future |
| ECC-CAL-116 | Edge | v2: series + resource — partial conflict dates | 409 + список конфликтных дат | API | P0 | US-CAL-02 edge | Future |
| ECC-CAL-117 | Edge | v2: iCal feed body ровно на LimitReader cap | parse OK / truncate policy documented | API | P0 | US-CAL-03, NF-07 | Future |
| ECC-CAL-118 | Edge | v2: iCal feed body cap+1 byte | reject / error state; no unbounded read | API | P0 | io.LimitReader | Future |
| ECC-CAL-119 | Edge | v2: max subscriptions per user (напр. 10) → 11-я | 422 | API | P1 | US-CAL-03 edge | Future |
| ECC-CAL-120 | Edge | v2: feed poll interval boundary (≥1h) | next poll ≥1h; no tighter storm | API | P1 | US-CAL-03 «не реже раза в час» | Future |
| ECC-CAL-121 | Edge | Expand RRULE as-is calls: COUNT large vs range 366d clip | instances clipped to range; no explode | Unit | P1 | ECC-CAL-003 | TODO |

**Calendar (as-is + v2) Edge в этом файле: 22** (`ECC-CAL-100`…`121`)

---

## 3. CAL-SYNC Edge (границы only) — `ECC-CAL-E100+`

> Только **Edge**-границы RSVP/CAL-SYNC (не Corner C1/C7–C10/C13/C15).  
> Фокус: max attendees, ICS attachment size, feed poll, range 366d в контексте invite/availability.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-E100 | Edge | Max attendees на событии (org size / hard cap) — ровно max | create OK; all invited | API | P0 | SEED-002, Brief | Future |
| ECC-CAL-E101 | Edge | Attendees = max+1 | 400/422; событие не создано | API | P0 | Brief C4 related | Future |
| ECC-CAL-E102 | Edge | Attendees = 0 (только organizer) | OK или 400 — контракт v1 | API | P1 | Brief C11 | Future |
| ECC-CAL-E103 | Edge | Attendees = 1 (organizer + one) | invite created; NATS fired | API | P1 | SEED-002 | Future |
| ECC-CAL-E104 | Edge | ICS attachment size ровно mail/provider limit | hub-mail sends; client opens | Manual | P1 | Brief L0 | Future |
| ECC-CAL-E105 | Edge | ICS attachment size limit+1 | truncate / link-fallback / 4xx — no silent drop | API/Manual | P1 | Brief L0, hub-mail | Future |
| ECC-CAL-E106 | Edge | ICS VTIMEZONE + event spanning TZ offset boundary | correct local in Outlook/Apple | Manual | P0 | Brief C6, NFR-12 | Future |
| ECC-CAL-E107 | Edge | Feed poll interval ровно 1h после last success | poll fires; no duplicate events | API | P1 | US-CAL-03 | Future |
| ECC-CAL-E108 | Edge | Feed poll interval <1h forced (misconfig) | clamped to ≥1h | API | P2 | US-CAL-03 | Future |
| ECC-CAL-E109 | Edge | CAL-SYNC list/availability range = 366d | 200 | API | P0 | maxRangeDays | Future |
| ECC-CAL-E110 | Edge | CAL-SYNC list/availability range > 366d | 400 too wide | API | P0 | ErrRangeTooWide | Future |
| ECC-CAL-E111 | Edge | Recurring series accept: COUNT на границе max expand | accept series; instances ≤ cap | API | P1 | Brief C2 Edge | Future |
| ECC-CAL-E112 | Edge | Two invites soft-conflict: overlap 1s vs touching | 1s → warning; touching → no warning | API | P1 | Brief C3 Edge | Future |
| ECC-CAL-E113 | Edge | needs_action не занимает hard-busy на границе slot | availability free/soft per policy | API/E2E | P1 | ECC-CAL-C20 | Future |
| ECC-CAL-E114 | Edge | Channel-call invite: attendees = all channel members на max channel size | all required; no silent truncate | API | P1 | Brief C11 | Future |

**CAL-SYNC Edge в этом файле: 15** (`ECC-CAL-E100`…`114`)

---

## 4. Storage / Files — `ECC-STOR-100+`

> Границы: Fiber BodyLimit **25 MB** · rate **10/min** · upload presign TTL **5m** · download **1h** · empty file · unicode filename · mime/Content-Type · 25MB+1

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-STOR-100 | Edge | Upload через backend buffer ровно 25 MB (26214400 bytes) | 200/201 OK | API | P0 | NFR-04, BodyLimit | TODO |
| ECC-STOR-101 | Edge | Upload через backend 25 MB + 1 byte | 413 Request Entity Too Large | API | P0 | NFR-04 | TODO |
| ECC-STOR-102 | Edge | Upload ровно 25 MB − 1 byte | 200 OK | API | P1 | NFR-04 boundary | TODO |
| ECC-STOR-103 | Edge | Presigned PUT ровно 25 MB (клиент→S3) | S3 accepts; Hub metadata OK | Manual | P1 | NFR-27 | TODO |
| ECC-STOR-104 | Edge | Presigned PUT 25 MB + 1 (NF-09 uncapped) | **может** пройти S3 — зафиксировать риск/мониторинг | Manual | P0 | NF-09, ECC-STOR-010 | TODO |
| ECC-STOR-105 | Edge | File rate: 10-й upload-url / мин | 200 | API | P0 | NFR-08 | TODO |
| ECC-STOR-106 | Edge | File rate: 11-й upload-url / мин | 429 | API | P0 | NFR-08, ECC-STOR-008 | TODO |
| ECC-STOR-107 | Edge | File rate: окно ровно 60s после 10-го | 11-й после окна → 200 | API | P1 | NFR-08 | TODO |
| ECC-STOR-108 | Edge | Upload presign TTL 5m−1s / 5m+1s | −1s PUT OK; +1s S3 403 | Manual | P1 | uploadURLTTL=5m | TODO |
| ECC-STOR-109 | Edge | Download presign TTL 1h−1s / 1h+1s | −1s GET OK; +1s S3 403 | Manual | P1 | downloadURLTTL=1h | TODO |
| ECC-STOR-110 | Edge | Empty file 0 bytes (presign + complete) | allow **или** 400 — единый контракт | API | P2 | ECC-STOR-013 | TODO |
| ECC-STOR-111 | Edge | Filename unicode NFC (`файл.pdf`) | sanitize keep readable; key safe | API | P1 | sanitizeFilename | TODO |
| ECC-STOR-112 | Edge | Filename unicode NFD / combining marks | normalized; no double-key | API | P2 | sanitizeFilename | TODO |
| ECC-STOR-113 | Edge | Filename emoji + CJK (`报告🚀.png`) | stored; download name correct | API | P1 | filename edge | TODO |
| ECC-STOR-114 | Edge | Filename max length boundary (ровно max) | 200 | API | P2 | ECC-STOR-015 | TODO |
| ECC-STOR-115 | Edge | Filename max length + 1 | 400 | API | P2 | ECC-STOR-015 | TODO |
| ECC-STOR-116 | Edge | Content-Type `application/octet-stream` + `.pdf` name | accept or reject per policy | API | P2 | ECC-STOR-014 | TODO |
| ECC-STOR-117 | Edge | Content-Type empty / missing on upload-url | 400 required fields | API | P1 | handler required | TODO |
| ECC-STOR-118 | Edge | Content-Type spoof (`image/png` for `.exe`) | store as declared **или** sniff reject | API | P1 | mime edge | TODO |
| ECC-STOR-119 | Edge | MIME charset boundary (`text/plain; charset=utf-8`) | accepted; stored metadata | API | P2 | contentType | TODO |
| ECC-STOR-120 | Edge | Multipart/form proxy path vs pure presign — size 25MB | BodyLimit applies only non-presign path | API | P1 | NFR-04 vs NFR-27 | TODO |

**Storage Edge в этом файле: 21** (`ECC-STOR-100`…`120`)

---

## Счётчик

| Домен | ID range | Кол-во Edge |
|-------|----------|-------------|
| Calls + LiveKit | `ECC-CALL-100`…`125` | **26** |
| Calendar as-is + v2 | `ECC-CAL-100`…`121` | **22** |
| CAL-SYNC (edge only) | `ECC-CAL-E100`…`114` | **15** |
| Storage / Files | `ECC-STOR-100`…`120` | **21** |
| **Итого** | | **84** |

| Проверки | Статус |
|----------|--------|
| Только тип **Edge** | ✅ |
| Минимум 45 строк | ✅ 84 ≥ 45 |
| mediasoup / SFU N/A | ✅ не добавлено |
| Основной `EDGE-CORNER-CATALOG.md` | ✅ не изменён |
| Git commit | ✅ не создавался |

**Следующий шаг:** `qa-manual-ver` → выборочно промоутить P0 в `docs/TESTING.md` как `TC-*` (после verify).
