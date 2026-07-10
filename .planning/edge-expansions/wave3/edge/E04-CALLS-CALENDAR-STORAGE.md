# Wave3 Edge E04 — Calls + Calendar + Storage

> **Тип:** только **Edge** (границы лимитов / TTL / размеров / дат / counts / quality / pagination)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave3  
> **Не трогает** `.planning/EDGE-CORNER-CATALOG.md`, wave1 `edge-expansions/04-*.md`, wave2 `wave2/edge/E04-*.md`  
> **ID range:** `ECC-*-300+` — **не пересекает** wave1 `*100+` / wave2 `*200+` / catalog `*0xx` / Corner `*C*`  
> **SFU:** LiveKit only (mediasoup / SEED-001 = REJECTED — строки N/A **не** добавлять)  
> **Связи:** NFR-03/04/08/12/20/27 · NF-09 · US-M06 · US-CAL-01…04 · SEED-002 · JULY brief §3 · TC-CALL-* · TC-STORAGE-*

**Назначение:** foresight-бэклог **новых** граничных кейсов (углубление wave1+wave2). Happy / Neg / Corner сюда не входят.

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (as-is) / **Future** (CAL-SYNC / v2 / product-lock cap) |
| Слой | API / Unit / Manual / E2E |
| Связь | NFR / US / TC / Brief / константа кода / wave1–2 ECC (расширение, не копия) |

**Фокус wave3:** LiveKit egress quality/bitrate · empty-room / departure_timeout · recording pagination · webhook latency skew · RRULE `INTERVAL`/`BYSETPOS`/`EXDATE` · all-day multi-day exclusive · feed ETag/304 · RSVP/invite TTL · availability slot grain · soft-delete 30d · filename* / Content-Disposition · Seaweed read-after-write · project file count/quota edge.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — `ECC-CALL-300+`

> Фокус: **egress quality/bitrate** · **empty-room / departure_timeout** · **recording list pagination** · **webhook latency skew** · **track publish at capacity**.  
> Wave1 `ECC-CALL-100…125` = participants/token TTL/active recordings/RRULE scheduled/conflict — **не дублировать**.  
> Wave2 `ECC-CALL-200…215` = recording duration cap — **не дублировать**.  
> LiveKit stay: RoomComposite egress → Seaweed; webhooks SEC-02.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-300 | Edge | RoomComposite egress bitrate = product soft-cap (ровно max kbps) | egress starts; file playable; metadata bitrate ≤cap | Manual | P1 | LiveKit egress quality | Future |
| ECC-CALL-301 | Edge | Egress bitrate request = soft-cap + 1 | clamp to cap **или** 400 — контракт; no OOM encoder | API/Manual | P1 | egress bitrate boundary | Future |
| ECC-CALL-302 | Edge | Egress resolution boundary 720p vs 1080p (product max) | max accepted; above → clamp/reject | Manual | P2 | LiveKit stay, US-M06 | Future |
| ECC-CALL-303 | Edge | Empty room: last participant leave; departure_timeout −1s vs +1s | room still open −1s; closed/cleaned +1s per LiveKit cfg | Manual | P0 | INF-01 departure_timeout | TODO |
| ECC-CALL-304 | Edge | Empty room timeout while recording active | egress finalize **или** fail-closed; no orphan `recording` forever | Manual/API | P0 | maxActive × empty-room | Future |
| ECC-CALL-305 | Edge | List project recordings: `limit` = product max page size (ровно) | 200; len≤limit; total correct | API | P1 | ListProjectRecordings | TODO |
| ECC-CALL-306 | Edge | List recordings: `limit` = max + 1 | clamp to max **или** 400 | API | P1 | pagination edge | TODO |
| ECC-CALL-307 | Edge | List recordings: `offset` = total (empty page) | 200 []; total unchanged | API | P1 | offset/limit | TODO |
| ECC-CALL-308 | Edge | List recordings: `offset` = total − 1, `limit` = 1 | ровно 1 last item | API | P1 | pagination boundary | TODO |
| ECC-CALL-309 | Edge | List recordings: `offset` = total + 1 | 200 [] **или** 400 — контракт | API | P2 | offset overflow | TODO |
| ECC-CALL-310 | Edge | LiveKit webhook `OnEgressEnded` arrives latency = Hub processing SLA −1s | status→ready/failed timely; UI consistent | Manual | P1 | NFR-20, OnEgressEnded | TODO |
| ECC-CALL-311 | Edge | Webhook `OnEgressEnded` delayed beyond Hub soft-timeout | status not stuck `recording`; reconcile job **или** failed | API/Manual | P0 | egress reconcile | Future |
| ECC-CALL-312 | Edge | Webhook event timestamp skew ±N vs Hub clock (within allow window) | accept if sig OK; store Hub-now for audit | API | P1 | SEC-02 × clock skew | TODO |
| ECC-CALL-313 | Edge | Webhook timestamp skew beyond allow window | reject 401/400; no state change | API | P0 | NFR-20 skew | TODO |
| ECC-CALL-314 | Edge | At 50 participants: 50-й publishes screen share (others audio-only) | share OK; room stays ≤50; no silent kick | Manual | P0 | NFR-03 × US-M06 | TODO |
| ECC-CALL-315 | Edge | At 50 participants: attempt 51-й track publish from existing member (2nd camera) | reject/limit tracks per policy; count participants still 50 | Manual | P1 | track publish cap | Future |
| ECC-CALL-316 | Edge | Stop both active recordings in same second (maxActive=2 → 0) | both terminal; next start allowed; no double-count | API | P1 | maxActiveRecordingsPerCall=2 | TODO |
| ECC-CALL-317 | Edge | Recording list filter status boundary: only `ready` vs empty statuses | inclusive filter; unknown status → 400 | API | P2 | ListProjectRecordings | TODO |
| ECC-CALL-318 | Edge | Guest invite code TTL −1s / +1s relative to call end window | −1s join OK; +1s 400/404; no token leak | API | P1 | TC-CALL-03/04 | TODO |
| ECC-CALL-319 | Edge | Call chat message length = max − 1 / max / max + 1 | −1 & max OK; max+1 → 400 | API | P2 | TC-CALL-09, ECC-CALL-015 deepen | TODO |
| ECC-CALL-320 | Edge | Raise-hand count at participant=50 (all hands up) | all visible; no drop last hand; UI scroll OK | Manual | P2 | TC-CALL-10 | TODO |
| ECC-CALL-321 | Edge | Member JWT remaining TTL = 60s when publishing screen | media OK until expiry; rejoin renews; screen re-publish after rejoin | Manual | P1 | token 24h remaining | TODO |
| ECC-CALL-322 | Edge | Recording download: offset=0 byte Range request vs full GET | Range honored **или** full only — documented; no 500 | Manual | P2 | downloadURLTTL=1h | TODO |
| ECC-CALL-323 | Edge | LiveKit room name / egress ID length at string max used by Hub | create/start OK; DB no truncate mid-id | Unit/API | P2 | room_name, egress_id | TODO |

**Calls Edge в этом файле: 24** (`ECC-CALL-300`…`323`)

---

## 2. Calendar as-is + Calendar v2 — `ECC-CAL-300+`

> Фокус: **RRULE `INTERVAL` / `BYSETPOS` / `EXDATE`** · **all-day multi-day exclusive** · **feed ETag/304** · **resource capacity touch**.  
> Wave1 `ECC-CAL-100…121` = range 366d, DST, basic COUNT/UNTIL — **не дублировать**.  
> Wave2 `ECC-CAL-200…217` = COUNT/UNTIL max, iCal event-count caps — **не дублировать**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-300 | Edge | RRULE `INTERVAL=1` (min meaningful) | instances every period; expand OK | Unit | P1 | US-CAL-02 | Future |
| ECC-CAL-301 | Edge | RRULE `INTERVAL=0` / negative | 400 validate; never expand | API/Unit | P0 | parseRRule | Future |
| ECC-CAL-302 | Edge | RRULE `INTERVAL` = product max (ровно) | expand OK within range clip | Unit/API | P0 | INTERVAL cap | Future |
| ECC-CAL-303 | Edge | RRULE `INTERVAL` = max + 1 | 400/422 | API | P0 | INTERVAL boundary | Future |
| ECC-CAL-304 | Edge | RRULE `FREQ=MONTHLY;BYSETPOS=-1` (last weekday/month) | last matching day; UTC store correct | Unit | P0 | BYSETPOS, NFR-12 | Future |
| ECC-CAL-305 | Edge | RRULE `BYSETPOS=1` vs `BYSETPOS=5` in short month (no 5th) | 5th skipped; no phantom day | Unit | P1 | BYSETPOS edge | Future |
| ECC-CAL-306 | Edge | RRULE `BYSETPOS=0` (invalid RFC) | 400 | Unit/API | P1 | RFC 5545 | Future |
| ECC-CAL-307 | Edge | `EXDATE` = first DTSTART (exclude seed) | 0 or next-only per policy; no crash | Unit | P0 | EXDATE | Future |
| ECC-CAL-308 | Edge | `EXDATE` = last COUNT occurrence | COUNT−1 instances; last gone | Unit | P1 | EXDATE × COUNT | Future |
| ECC-CAL-309 | Edge | `EXDATE` outside series window | ignored; series unchanged | Unit | P2 | EXDATE noop | Future |
| ECC-CAL-310 | Edge | All-day multi-day: DTSTART=D1, DTEND=D1+3 (exclusive) | FE shows 3 days; ICS DTEND=D1+3 | Unit/API | P0 | RFC 5545 all-day | TODO |
| ECC-CAL-311 | Edge | All-day: DTEND = DTSTART (zero-length exclusive) | 400 **или** 1-day normalize — контракт | API | P0 | all-day exclusive | TODO |
| ECC-CAL-312 | Edge | All-day spanning DST spring-forward (local midnights) | UTC instants correct; no missing/extra day | E2E/Manual | P0 | NFR-12 × all-day | TODO |
| ECC-CAL-313 | Edge | iCal feed: `If-None-Match` / ETag match → 304 | 304 empty body; no re-parse storm | API | P1 | US-CAL-03 poll | Future |
| ECC-CAL-314 | Edge | iCal feed: ETag change by 1 byte content | 200 full body; new ETag | API | P1 | feed ETag | Future |
| ECC-CAL-315 | Edge | iCal feed: `If-Modified-Since` = Last-Modified exact | 304 **или** 200 per HTTP semantics — зафиксировать | API | P2 | feed IMS | Future |
| ECC-CAL-316 | Edge | Resource capacity=1: booking A ends_at == B starts_at (touch) | **не** 409 | API | P0 | US-CAL-01, ECC-CAL-109 deepen | Future |
| ECC-CAL-317 | Edge | Resource capacity=1: overlap 1ms | 409 | API | P0 | US-CAL-01 | Future |
| ECC-CAL-318 | Edge | Resource capacity=N: N concurrent OK; N+1 → 409 | exact capacity boundary | API | P0 | US-CAL-01 capacity | Future |
| ECC-CAL-319 | Edge | List range from/to with microsecond precision at 366d edge | inclusive/exclusive per validateRange; no off-by-1d | API | P1 | maxRangeDays=366 | TODO |
| ECC-CAL-320 | Edge | VEVENT `DTSTART` floating (no TZID) vs `Z` in same feed | parse policy; store UTC consistently | Unit | P1 | ical.go, NFR-12 | Future |
| ECC-CAL-321 | Edge | Subscriptions per user = max − 1 / max / max + 1 | −1 & max OK; max+1 → 422 | API | P1 | ECC-CAL-119 deepen | Future |
| ECC-CAL-322 | Edge | Feed poll success then next allowed at exactly 1h boundary | poll allowed; earlier clamped | API | P1 | US-CAL-03 ≥1h | Future |

**Calendar Edge в этом файле: 23** (`ECC-CAL-300`…`322`)

---

## 3. CAL-SYNC Edge — `ECC-CAL-E300+`

> Фокус: **RSVP/invite token TTL** · **availability slot grain** · **ICS METHOD size** · **series exception RSVP** · **soft-conflict threshold**.  
> Wave1 `ECC-CAL-E100…114` / wave2 `ECC-CAL-E200…213` = attendee max / ICS size / poll — здесь **другие** границы.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-E300 | Edge | RSVP action token TTL −1s / +1s | −1s accept OK; +1s 401/410; no state flip | API | P0 | SEED-002 invite TTL | Future |
| ECC-CAL-E301 | Edge | Invite link open at TTL exact second | inclusive **или** exclusive — documented; deterministic | API | P0 | invite TTL boundary | Future |
| ECC-CAL-E302 | Edge | Availability slot grain = 15m (product default): query aligned | slots on grid; no half-slot leak | API | P1 | US-CAL-04 | Future |
| ECC-CAL-E303 | Edge | Availability query with from/to not aligned to grain | snap/reject per policy; no silent wrong busy | API | P1 | slot grain | Future |
| ECC-CAL-E304 | Edge | Availability grain = 1m (min) vs max grain (e.g. 60m) | both honor bounds; invalid grain → 400 | API | P1 | US-CAL-04 edge | Future |
| ECC-CAL-E305 | Edge | Soft-conflict warning: overlap exactly 1s between two invites | warning fired; create still OK | API | P0 | Brief C3, ECC-CAL-E112 deepen | Future |
| ECC-CAL-E306 | Edge | Soft-conflict: events touch (end==start) | **no** warning | API | P0 | Brief C3 Edge | Future |
| ECC-CAL-E307 | Edge | Soft-conflict: overlap duration = product min threshold − 1ms | no warning if below threshold | API | P2 | soft-conflict threshold | Future |
| ECC-CAL-E308 | Edge | Series exception instance: RSVP only that instance (if v2 allows) | instance status changes; series root unchanged **или** v1 series-only 400 | API | P1 | Brief C2 | Future |
| ECC-CAL-E309 | Edge | ICS `METHOD=REQUEST` attachment size = mail limit − 1 | send OK | Manual | P1 | Brief L0 | Future |
| ECC-CAL-E310 | Edge | ICS `METHOD=CANCEL` after delete: size at limit | cancel mail sent; inbox cleared | Manual/API | P1 | Brief C1/C9 × L0 | Future |
| ECC-CAL-E311 | Edge | needs_action invite: slot start−1s / start / start+1s vs busy calc | only accepted hard-busy; needs_action soft/free per policy | API | P1 | ECC-CAL-C20, E113 | Future |
| ECC-CAL-E312 | Edge | Leave approved busy block: ends_at == meeting starts_at | touching OK (free after leave) | API | P1 | Brief C12, US-CAL-04 | Future |
| ECC-CAL-E313 | Edge | Leave busy overlap 1s into meeting | busy shown; organizer sees conflict hint | API | P1 | Brief C12 | Future |
| ECC-CAL-E314 | Edge | Organizer cancel at attendee RSVP mix (all states present) | all terminals; NATS × N; no stuck needs_action | API | P1 | Brief C5/C9 | Future |
| ECC-CAL-E315 | Edge | Re-invite same user after decline: unique attendees still ≤ max | OK; history/audit; count unique | API | P2 | attendee unique | Future |
| ECC-CAL-E316 | Edge | Availability range = 366d with max-attendee events dense | 200; response time bounded; no OOM | API | P1 | maxRangeDays × C11 | Future |

**CAL-SYNC Edge в этом файле: 17** (`ECC-CAL-E300`…`316`)

---

## 4. Storage / Files — `ECC-STOR-300+`

> Фокус: **soft-delete 30d** · **Content-Disposition / filename*** · **Seaweed read-after-write** · **checksum/ETag** · **concurrent complete boundary** · **project file count**.  
> Wave1 `ECC-STOR-100…120` / wave2 `ECC-STOR-200…220` = BodyLimit/0B/25MB/mime/skew — **не копировать** те же строки.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-STOR-300 | Edge | Soft-delete file: restore at trash day = 30 − 1s | restore OK; ready again | API | P0 | NFR-11, 30d trash | TODO |
| ECC-STOR-301 | Edge | Soft-delete: hard purge window = 30d + 1s | restore 404/gone; GDPR/admin hard only | API | P0 | soft delete 30d | TODO |
| ECC-STOR-302 | Edge | Soft-delete at exactly 30d boundary (cron tick) | deterministic purge **или** inclusive keep — documented | API | P0 | trash TTL | TODO |
| ECC-STOR-303 | Edge | Download `Content-Disposition` filename* UTF-8 (RFC 5987) | browser shows unicode name; no mojibake | Manual | P1 | filename*, sanitize | TODO |
| ECC-STOR-304 | Edge | Filename with `;` / `"` / CR/LF in Content-Disposition | sanitized; no header injection | API | P0 | header injection | TODO |
| ECC-STOR-305 | Edge | Filename path segments (`../a.pdf`, `a/b/c.pdf`) | flattened/safe key; no escape bucket prefix | API | P0 | sanitizeFilename | TODO |
| ECC-STOR-306 | Edge | Seaweed read-after-write: GET download URL issued 0ms after complete | 200 **или** retry-until-consistent; no Hub ready + S3 404 forever | Manual | P0 | Seaweed eventual consistency | TODO |
| ECC-STOR-307 | Edge | Read-after-write delay = product retry budget − 1 / + 1 | −1 eventually OK; +1 surface error/retry exhausted | Manual | P1 | complete→download | TODO |
| ECC-STOR-308 | Edge | Complete with client checksum/ETag match (exact) | ready; ETag stored if supported | API | P1 | integrity | Future |
| ECC-STOR-309 | Edge | Complete with checksum mismatch vs S3 object | reject; pending not → ready | API | P0 | size/checksum integrity | Future |
| ECC-STOR-310 | Edge | Complete called twice at same version/state (idempotent edge) | second 200 same **или** 409; no double-ready side effects | API | P1 | complete idempotency | TODO |
| ECC-STOR-311 | Edge | Two completes for different pending IDs same S3 key (collision) | second fail-closed; no overwrite silent | API | P0 | key uniqueness | TODO |
| ECC-STOR-312 | Edge | Project file count = product soft-cap (ровно) | upload-url OK | API | P2 | project quota | Future |
| ECC-STOR-313 | Edge | Project file count = soft-cap + 1 | 422/403; no new pending | API | P2 | project quota | Future |
| ECC-STOR-314 | Edge | Pending upload abandoned: TTL of pending row = uploadURLTTL + grace | pending GC; no ready orphan | API | P1 | uploadURLTTL=5m | TODO |
| ECC-STOR-315 | Edge | Pending GC at TTL − 1s vs + 1s | −1s still completable; +1s complete rejected | API | P1 | pending TTL | TODO |
| ECC-STOR-316 | Edge | Download after soft-delete (in trash, before purge) | 403/410; no presign leak | API | P0 | trash ACL | TODO |
| ECC-STOR-317 | Edge | Re-presign download for ready file at rate limit boundary (10th/11th if shared bucket) | file download rate policy documented; 429 if capped | API | P2 | NFR-08 interaction | TODO |
| ECC-STOR-318 | Edge | Multipart form field filename empty + binary body size 1 byte | 400 required filename **или** generated name — контракт | API | P2 | buffer path | TODO |
| ECC-STOR-319 | Edge | S3 object Content-Length header 25 MB exact vs Hub declared 25 MB − 1 | mismatch reject on complete | API | P0 | NF-09 integrity | TODO |
| ECC-STOR-320 | Edge | Trash list pagination: last page offset/limit boundary | last items; empty next page | API | P2 | soft-delete list | TODO |
| ECC-STOR-321 | Edge | Recording file via storage key: download TTL 1h with Seaweed clock ±skew | same policy as STOR skew wave2; recording path consistent | Manual | P1 | CALL download × STOR | TODO |

**Storage Edge в этом файле: 22** (`ECC-STOR-300`…`321`)

---

## Счётчик

| Домен | ID range | Кол-во Edge |
|-------|----------|-------------|
| Calls + LiveKit (egress/empty-room/pagination/webhook skew) | `ECC-CALL-300`…`323` | **24** |
| Calendar (INTERVAL/BYSETPOS/EXDATE/all-day/ETag/resource) | `ECC-CAL-300`…`322` | **23** |
| CAL-SYNC (RSVP TTL/slot grain/soft-conflict/ICS METHOD) | `ECC-CAL-E300`…`316` | **17** |
| Storage (30d trash/filename*/RAW/checksum/quota) | `ECC-STOR-300`…`321` | **22** |
| **Итого** | | **86** |

| Проверки | Статус |
|----------|--------|
| Только тип **Edge** | ✅ |
| Минимум 50 строк | ✅ 86 ≥ 50 |
| ID `*300+` без пересечения `*100+` / `*200+` | ✅ |
| LiveKit stay / mediasoup N/A | ✅ не добавлено |
| Основной `EDGE-CORNER-CATALOG.md` | ✅ не изменён |
| Git commit | ✅ не создавался |

**Фокус coverage (wave3):**

| Тема | ID (примеры) |
|------|----------------|
| Egress quality/bitrate/resolution | `ECC-CALL-300`…`302` |
| Empty-room / departure_timeout | `ECC-CALL-303`…`304` |
| Recording pagination | `ECC-CALL-305`…`309`, `317` |
| Webhook latency / clock skew | `ECC-CALL-310`…`313` |
| Capacity × screen/tracks | `ECC-CALL-314`…`315`, `320`–`321` |
| RRULE INTERVAL/BYSETPOS/EXDATE | `ECC-CAL-300`…`309` |
| All-day multi-day / DST | `ECC-CAL-310`…`312` |
| Feed ETag/304 / subscriptions | `ECC-CAL-313`…`315`, `321`–`322` |
| Resource capacity touch | `ECC-CAL-316`…`318` |
| RSVP/invite TTL | `ECC-CAL-E300`…`301` |
| Availability slot grain | `ECC-CAL-E302`…`304` |
| Soft-conflict threshold | `ECC-CAL-E305`…`307` |
| Soft-delete 30d | `ECC-STOR-300`…`302`, `316`, `320` |
| filename* / path injection | `ECC-STOR-303`…`305` |
| Seaweed RAW / checksum | `ECC-STOR-306`…`311`, `319` |
| Pending TTL / quota | `ECC-STOR-312`…`315` |

**Следующий шаг:** `qa-manual-ver` → выборочно промоутить P0 в `docs/TESTING.md` как `TC-*` (после verify). Не мержить в каталог без отдельной задачи.
