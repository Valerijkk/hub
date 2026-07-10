# Wave4 Edge E04 — Calls + Calendar + Storage

> **Тип:** только **Edge** (границы лимитов / TTL / размеров / дат / counts / skew)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave4  
> **Не трогает** `.planning/EDGE-CORNER-CATALOG.md` и wave1/wave2 `edge-expansions/**`  
> **ID range:** `ECC-*-400+` — **не пересекает** wave1 `*100+` / wave2 `*200+` / catalog `*0xx` / Corner `*C*`  
> **SFU:** LiveKit only (mediasoup / SEED-001 = REJECTED — строки N/A **не** добавлять)  
> **Связи:** NFR-03/04/08/12/27 · NF-09 · SEC-02 · US-M06 · US-CAL-01…04 · SEED-002 · JULY brief §3 · TC-CALL-* · TC-STORAGE-*

**Назначение:** foresight-бэклог **новых** граничных кейсов (углубление wave1/wave2). Happy / Neg / Corner сюда не входят.  
**Не копии** `*100–*300` (participants/token TTL/duration-cap/RRULE COUNT-UNTIL max/attendee max/0B-25MB-mime-skew).

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (as-is) / **Future** (CAL-SYNC / v2 / product-lock cap) |
| Слой | API / Unit / Manual / E2E |
| Связь | NFR / US / TC / Brief / константа кода / wave1–2 ECC (расширение, не копия) |

**Фокус wave4:** LiveKit webhook body/JWT skew · room emptyTimeout/idle · invite-code/guest-name length · call-chat size · RRULE `INTERVAL`/`BYSETPOS`/`EXDATE` · aggregator density · RSVP/reminder/SEQUENCE TTL · soft-conflict minutes · trash day-30 · orphan pending cutoff · filename/key length · sliding rate window.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — `ECC-CALL-400+`

> Фокус: **webhook body/JWT skew** · **emptyTimeout / idle room** · **invite-code / guest-name / title length** · **call-chat size** · **recording list pagination** · **retention 30d**.  
> Wave1 `ECC-CALL-100…125` = participants/token TTL/active recordings/RRULE scheduled/conflict — **не дублировать**.  
> Wave2 `ECC-CALL-200…215` = recording duration cap — **не дублировать**.  
> LiveKit stay: RoomService + webhook JWT (SEC-02 sha256) + egress → Seaweed.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-400 | Edge | LiveKit webhook body ровно product/LimitReader cap (bytes) | verify OK; event processed; no truncate mid-JSON | API | P0 | SEC-02, webhook body | TODO |
| ECC-CALL-401 | Edge | LiveKit webhook body = cap + 1 byte | 400/413; signature path fail-closed; no partial apply | API | P0 | LimitReader, SEC-02 | TODO |
| ECC-CALL-402 | Edge | Webhook JWT `exp` = Hub-now − 1s (clock skew −) | 401; no state change | API/Unit | P0 | webhook JWT skew | TODO |
| ECC-CALL-403 | Edge | Webhook JWT `exp` = Hub-now + skew-window − 1s | accept if within LiveKit/Hub skew policy | API/Unit | P0 | webhook JWT skew | TODO |
| ECC-CALL-404 | Edge | Webhook JWT `nbf` = Hub-now + 1s (future) | 401 until nbf; no early apply | Unit | P1 | webhook nbf | TODO |
| ECC-CALL-405 | Edge | Webhook `sha256` claim present but body hash differs by 1 bit | 401; SEC-02 enforced | API | P0 | SEC-02 | TODO |
| ECC-CALL-406 | Edge | Room `emptyTimeout` = product default (LiveKit CreateRoom) — last participant leaves | room closes after exactly timeout; Hub `OnRoomFinished` once | Manual/API | P0 | emptyTimeout, LiveKit | TODO |
| ECC-CALL-407 | Edge | Last leave at emptyTimeout − 1s then rejoin | room stays; no premature finish | Manual | P0 | emptyTimeout boundary | TODO |
| ECC-CALL-408 | Edge | `departureTimeout` boundary: disconnect then reconnect within window | participant not dropped; media resumes | Manual | P1 | LiveKit departureTimeout | TODO |
| ECC-CALL-409 | Edge | Idle room (0 participants) created but never joined until emptyTimeout | room auto-delete; Hub call ended/cancelled; invite dead | API/Manual | P1 | emptyTimeout × invite | TODO |
| ECC-CALL-410 | Edge | Invite code length = 8 (as-is `uuid[:8]`) join | 200 token | API | P1 | inviteCode len=8 | TODO |
| ECC-CALL-411 | Edge | Invite code length = 7 / 9 (truncated/padded) | 400/404; no token leak | API | P1 | invite code boundary | TODO |
| ECC-CALL-412 | Edge | Invite code charset edge: uppercase vs lowercase same hex | case policy documented; one canonical match | API | P2 | invite normalize | TODO |
| ECC-CALL-413 | Edge | Guest `guest_name` length = product max (ровно max runes) | join OK; display truncated only if UI cap separate | API | P1 | GuestJoinCall | TODO |
| ECC-CALL-414 | Edge | Guest `guest_name` length = max + 1 | 400; no LiveKit identity created | API | P1 | guest_name max | TODO |
| ECC-CALL-415 | Edge | Guest `guest_name` length = 1 (min non-empty) vs empty/whitespace | 1 OK; empty 400 | API | P1 | guest_name min | TODO |
| ECC-CALL-416 | Edge | Instant call `title` length = max vs max+1 | max OK; max+1 400 | API | P2 | CreateInstantCall title | TODO |
| ECC-CALL-417 | Edge | Call-chat message body = product max bytes (ровно) | 200; stored in Mongo | API | P1 | call chat size | TODO |
| ECC-CALL-418 | Edge | Call-chat message body = max + 1 | 400; no partial write | API | P1 | call chat size | TODO |
| ECC-CALL-419 | Edge | Call-chat rate: N-th message at messages/min boundary (60/min NFR) | N OK; N+1 → 429 | API | P1 | NFR messages 60/min | TODO |
| ECC-CALL-420 | Edge | List project recordings: `limit` = product max page size | 200; ≤max items; total correct | API | P1 | ListProjectRecordings | TODO |
| ECC-CALL-421 | Edge | List recordings: `limit` = max + 1 | clamp to max **или** 400 — контракт | API | P1 | pagination clamp | TODO |
| ECC-CALL-422 | Edge | List recordings: `offset` = total (empty page) vs offset = total−1 (last 1) | [] vs 1 item; no 500 | API | P2 | offset boundary | TODO |
| ECC-CALL-423 | Edge | Recording soft-delete / trash: age = 30d − 1s | still restorable / download policy OK | API | P0 | soft-delete 30d | TODO |
| ECC-CALL-424 | Edge | Recording trash age = 30d + 1s | hard-purged or download 404; no leak | API | P0 | soft-delete 30d | TODO |
| ECC-CALL-425 | Edge | LiveKit identity string = `user:`+UUID length boundary / malformed short UUID | resolve skip or 400; webhook no panic | Unit/API | P1 | resolveParticipant | TODO |
| ECC-CALL-426 | Edge | Max video publishes: webcam + screen-share simultaneous (2 video tracks) at participant=50 room | both tracks OK; peers see both; egress includes per policy | Manual | P1 | US-M06 × NFR-03 | TODO |
| ECC-CALL-427 | Edge | Scheduled call `starts_at` = now + minLead (ровно) vs minLead − 1s | OK vs 400 too soon | API | P1 | scheduled lead time | Future |
| ECC-CALL-428 | Edge | Webhook flood: same `room_finished` delivered twice within 1s (idempotent edge) | second no-op; call stays ended once | API | P0 | webhook idempotency edge | TODO |

**Calls Edge в этом файле: 29** (`ECC-CALL-400`…`428`)

---

## 2. Calendar as-is + Calendar v2 — `ECC-CAL-400+`

> Фокус: **RRULE `INTERVAL` / `BYSETPOS` / `EXDATE` count** · **aggregator density** · **timestamp precision** · **title length** · **all-day multi-day span**.  
> Wave1 `ECC-CAL-100…121` = range 366d, DST, basic COUNT/UNTIL — **не дублировать**.  
> Wave2 `ECC-CAL-200…217` = COUNT/UNTIL max + iCal event caps — **не дублировать**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-400 | Edge | RRULE `INTERVAL=1` (min meaningful) daily | every day; expand OK | Unit | P1 | US-CAL-02 INTERVAL | Future |
| ECC-CAL-401 | Edge | RRULE `INTERVAL=0` | 400; never expand | Unit/API | P0 | INTERVAL=0 | Future |
| ECC-CAL-402 | Edge | RRULE `INTERVAL` = product max | expand OK; spacing correct | Unit | P1 | INTERVAL max | Future |
| ECC-CAL-403 | Edge | RRULE `INTERVAL` = max + 1 | 400 | API | P1 | INTERVAL max | Future |
| ECC-CAL-404 | Edge | RRULE `BYSETPOS=-1` (last in set) monthly | last weekday/instance correct | Unit | P0 | BYSETPOS | Future |
| ECC-CAL-405 | Edge | RRULE `BYSETPOS=0` (invalid) | 400 | Unit/API | P0 | BYSETPOS=0 | Future |
| ECC-CAL-406 | Edge | RRULE `BYSETPOS` = ±N beyond set size | 0 instances that month **или** skip — RFC policy | Unit | P1 | BYSETPOS OOB | Future |
| ECC-CAL-407 | Edge | `EXDATE` count = product max exclusions | series OK; excluded dates absent | Unit/API | P0 | EXDATE cap | Future |
| ECC-CAL-408 | Edge | `EXDATE` count = max + 1 | 400; series not saved | API | P0 | EXDATE cap | Future |
| ECC-CAL-409 | Edge | `EXDATE` exactly = one DTSTART occurrence | that instance removed; others remain | Unit | P1 | EXDATE = start | Future |
| ECC-CAL-410 | Edge | `RDATE` count = max vs max+1 | max OK; max+1 400 | API | P1 | RDATE cap | Future |
| ECC-CAL-411 | Edge | Aggregator list: dense mix (calls+sprints+deadlines+leave) count = product page/cap at 366d range | 200; complete; no OOM; order stable | API | P0 | maxRangeDays × density | TODO |
| ECC-CAL-412 | Edge | Aggregator density = cap + 1 in range | truncate+cursor **или** 400 — no unbounded payload | API | P0 | list density cap | Future |
| ECC-CAL-413 | Edge | Query `from`/`to` with millisecond vs second vs nanosecond precision | parse OK; range compare in UTC consistent | API | P1 | timestamp precision | TODO |
| ECC-CAL-414 | Edge | Query `from` with timezone offset `+14:00` / `-12:00` extreme IANA | converted to UTC store; list correct | API | P1 | NFR-12 extreme TZ | TODO |
| ECC-CAL-415 | Edge | All-day multi-day: DTEND exclusive = start+N days at N=product max span | OK; FE shows N days | Unit/API | P1 | all-day span max | Future |
| ECC-CAL-416 | Edge | All-day span = max + 1 day | 400 | API | P1 | all-day span | Future |
| ECC-CAL-417 | Edge | Event title length = max vs max+1 | max OK; max+1 400 | API | P2 | title max | Future |
| ECC-CAL-418 | Edge | Event description length = max bytes (UTF-8 multibyte at edge) | OK if ≤max bytes; max+1 rune that crosses byte cap → 400 | API | P1 | description byte cap | Future |
| ECC-CAL-419 | Edge | Leave item spanning year boundary inside 366d window (Dec 31 → Jan 1) | single item; dates correct UTC | API | P1 | leave × NFR-12 | TODO |
| ECC-CAL-420 | Edge | Sprint start/end touching calendar range edges (from inclusive / to exclusive policy) | included iff contract; documented | API | P1 | validateRange inclusivity | TODO |
| ECC-CAL-421 | Edge | `FREQ=WEEKLY;BYDAY=` empty / invalid token | 400 | Unit | P1 | BYDAY validate | Future |
| ECC-CAL-422 | Edge | `FREQ=MONTHLY;BYMONTHDAY=-1` (last day) across 28–31 months | last day each month | Unit | P0 | BYMONTHDAY=-1 | Future |
| ECC-CAL-423 | Edge | Floating local all-day vs `TZID=Etc/UTC` same civil date | one canonical UTC day; no duplicate | Unit | P1 | floating × TZID | Future |

**Calendar Edge в этом файле: 24** (`ECC-CAL-400`…`423`)

---

## 3. CAL-SYNC Edge — `ECC-CAL-E400+`

> Фокус: **reminder lead-time** · **RSVP token TTL** · **ICS SEQUENCE** · **soft-conflict minutes** · **availability slot granularity** · **reschedule delta**.  
> Wave1 `ECC-CAL-E100…114` / wave2 `ECC-CAL-E200…213` = attendee max / ICS size / poll — **не дублировать**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-E400 | Edge | Reminder lead = 0 (at start) | one notif at starts_at; no early fire | API | P1 | reminder 0 | Future |
| ECC-CAL-E401 | Edge | Reminder lead = product max (e.g. 7d) | fire at starts_at − max; OK | API | P1 | reminder max | Future |
| ECC-CAL-E402 | Edge | Reminder lead = max + 1 / negative | 400 | API | P0 | reminder bounds | Future |
| ECC-CAL-E403 | Edge | Reminder lead = 1s (min >0) | fires; no duplicate with start notif if distinct channels | API | P2 | reminder min | Future |
| ECC-CAL-E404 | Edge | RSVP / invite token TTL = product − 1s | accept OK | API | P0 | invite token TTL | Future |
| ECC-CAL-E405 | Edge | RSVP token TTL + 1s | 401/410; RSVP unchanged | API | P0 | invite token TTL | Future |
| ECC-CAL-E406 | Edge | ICS `SEQUENCE` = 0 (first) vs = maxInt32-adjacent | send OK; client replaces; no overflow | Unit/API | P1 | ICS SEQUENCE | Future |
| ECC-CAL-E407 | Edge | ICS `SEQUENCE` increment on reschedule: n → n+1 at boundary | mail update; old SEQUENCE ignored by clients | Manual | P1 | Brief reschedule | Future |
| ECC-CAL-E408 | Edge | Soft-conflict warning window = 0 min (touching only) | touching → no warning; 1s overlap → warning | API | P0 | Brief C3 minutes=0 | Future |
| ECC-CAL-E409 | Edge | Soft-conflict window = product max minutes | overlap within window warns; outside silent | API | P1 | soft-conflict max | Future |
| ECC-CAL-E410 | Edge | Availability slot granularity = 15m (product): query aligned vs +1m offset | aligned buckets; offset rounded per policy | API | P1 | §3.2 slot grain | Future |
| ECC-CAL-E411 | Edge | Availability range duration = 1 slot (min) vs 0 | 1 slot OK; 0 → 400 | API | P1 | availability min | Future |
| ECC-CAL-E412 | Edge | Reschedule delta = 0 (same start/end) | 400 no-op **или** 200 idempotent — контракт | API | P1 | reschedule 0 | Future |
| ECC-CAL-E413 | Edge | Reschedule delta = +1s / −1s | OK; SEQUENCE++; RSVP policy applied | API | P1 | reschedule 1s | Future |
| ECC-CAL-E414 | Edge | Decline then re-invite same user: cycle count = max vs max+1 | max OK; max+1 429/422 | API | P2 | reinvite cap | Future |
| ECC-CAL-E415 | Edge | needs_action TTL (auto-expire invite) = product days − 1s / +1s | −1s still actionable; +1s expired/needs_action cleared per policy | API | P1 | invite expire | Future |
| ECC-CAL-E416 | Edge | Free/busy query with leave block exactly touching event start | leave ends == event start → free (touching) | API | P1 | leave × busy touch | Future |
| ECC-CAL-E417 | Edge | Organizer-only event (0 attendees) + reminder lead max | reminder only to organizer; no fanout | API | P2 | reminder × 0 attendees | Future |

**CAL-SYNC Edge в этом файле: 18** (`ECC-CAL-E400`…`417`)

---

## 4. Storage / Files — `ECC-STOR-400+`

> Фокус: **trash day-30** · **orphan pending cutoff** · **filename/key length** · **sliding rate window** · **size mismatch ±1** · **Content-Disposition / path separators**.  
> Wave1 `ECC-STOR-100…120` / wave2 `ECC-STOR-200…220` = BodyLimit 25MB / mime / skew / 0-byte — **не дублировать** те же строки.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-STOR-400 | Edge | Soft-delete trash age = 30d − 1s | restore OK; download OK | API | P0 | soft-delete 30d | TODO |
| ECC-STOR-401 | Edge | Trash age = 30d exactly | policy inclusive/exclusive documented; one consistent | API | P0 | day-30 boundary | TODO |
| ECC-STOR-402 | Edge | Trash age = 30d + 1s | hard purge; GET 404; restore 404 | API | P0 | soft-delete 30d | TODO |
| ECC-STOR-403 | Edge | Orphan pending cutoff = product age − 1s | not cleaned yet; still pending | API | P1 | orphan_cleanup cutoff | TODO |
| ECC-STOR-404 | Edge | Orphan pending age = cutoff + 1s | cleaned in batch; S3 object deleted or marked | API | P1 | orphanBatchLimit | TODO |
| ECC-STOR-405 | Edge | Orphan cleanup batch = `orphanBatchLimit` (500) orphans | processes ≤500/run; remainder next run | Unit/API | P1 | orphanBatchLimit=500 | TODO |
| ECC-STOR-406 | Edge | Orphan batch = 500 + 1 pending | first run 500; second run 1; no skip forever | API | P1 | orphan batch edge | TODO |
| ECC-STOR-407 | Edge | Filename length = product max (bytes or runes — SoT) | 200; key safe | API | P1 | sanitizeFilename max | TODO |
| ECC-STOR-408 | Edge | Filename length = max + 1 | 400 | API | P1 | filename max | TODO |
| ECC-STOR-409 | Edge | Filename with path separators (`../`, `\`, null byte) at max-1 length | sanitized/rejected; no path escape | API | P0 | path traversal | TODO |
| ECC-STOR-410 | Edge | S3 object key length = Seaweed/S3 max − 1 / max / max+1 | −1/max OK; max+1 reject at Hub before PUT | API | P1 | key length | TODO |
| ECC-STOR-411 | Edge | Sliding rate window: 10th upload-url at T+0; 11th at T+59.999s | 11th → 429 | API | P0 | NFR-08 sliding | TODO |
| ECC-STOR-412 | Edge | Sliding rate: 11th at T+60.001s after 10th | 200; window rolled | API | P0 | NFR-08 window | TODO |
| ECC-STOR-413 | Edge | Complete: declared size = S3 size − 1 byte | reject mismatch; pending not ready | API | P0 | size ±1 integrity | TODO |
| ECC-STOR-414 | Edge | Complete: declared size = S3 size + 1 byte | reject; no ready | API | P0 | size ±1 | TODO |
| ECC-STOR-415 | Edge | Content-Disposition `filename*` RFC5987 length = max encoded | download header valid; client name OK | API | P2 | RFC5987 | TODO |
| ECC-STOR-416 | Edge | Content-Disposition `filename*` = max + 1 encoded bytes | truncate header **или** 400 — no corrupt header | API | P2 | RFC5987 max | TODO |
| ECC-STOR-417 | Edge | Pending→ready: ETag/checksum length = empty vs exact S3 ETag | empty policy; mismatch → reject | API | P1 | ETag edge | TODO |
| ECC-STOR-418 | Edge | Download of file in trash at day-30 − 1s via old bookmarked URL | 403/404 per trash policy; no silent serve | API | P0 | trash × download | TODO |
| ECC-STOR-419 | Edge | Re-upload same sanitized filename after trash purge day-30+ | new file id; no key collision with purged | API | P1 | key reuse after purge | TODO |
| ECC-STOR-420 | Edge | Multipart form field count / name length at Fiber parse limits | 400 on over-limit; no panic | API | P2 | multipart edge | TODO |
| ECC-STOR-421 | Edge | Presign upload Content-Length header = 0 with intent size>0 (client lie) | S3 may store 0; Hub complete must catch vs declared | Manual/API | P1 | NF-09 × declare | TODO |

**Storage Edge в этом файле: 22** (`ECC-STOR-400`…`421`)

---

## Счётчик

| Домен | ID range | Кол-во Edge |
|-------|----------|-------------|
| Calls + LiveKit (webhook/idle/invite/chat/retention) | `ECC-CALL-400`…`428` | **29** |
| Calendar (INTERVAL/BYSETPOS/EXDATE/density) | `ECC-CAL-400`…`423` | **24** |
| CAL-SYNC (reminder/RSVP/SEQUENCE/soft-conflict) | `ECC-CAL-E400`…`417` | **18** |
| Storage (trash30/orphan/key/rate±1) | `ECC-STOR-400`…`421` | **22** |
| **Итого** | | **93** |

| Проверки | Статус |
|----------|--------|
| Только тип **Edge** | ✅ |
| Минимум 50 строк | ✅ 93 ≥ 50 |
| ID `*400+` без пересечения `*100+` / `*200+` | ✅ |
| Не копии сценариев wave1/wave2 | ✅ новый фокус |
| LiveKit stay / mediasoup N/A | ✅ не добавлено |
| Основной `EDGE-CORNER-CATALOG.md` | ✅ не изменён |
| Git commit | ✅ не создавался |

**Фокус coverage (wave4):**

| Тема | ID (примеры) |
|------|----------------|
| Webhook body / JWT skew / SEC-02 | `ECC-CALL-400`…`405`, `428` |
| emptyTimeout / departureTimeout / idle | `ECC-CALL-406`…`409` |
| Invite / guest_name / title / chat size | `ECC-CALL-410`…`419` |
| Recording pagination / trash 30d | `ECC-CALL-420`…`424` |
| RRULE INTERVAL / BYSETPOS / EXDATE / RDATE | `ECC-CAL-400`…`410`, `421`–`423` |
| Aggregator density / TS precision / all-day span | `ECC-CAL-411`…`420` |
| Reminder / RSVP TTL / SEQUENCE / soft-conflict | `ECC-CAL-E400`…`417` |
| Trash 30d / orphan batch / key / sliding rate / ±1 size | `ECC-STOR-400`…`421` |

**Следующий шаг:** `qa-manual-ver` → выборочно промоутить P0 в `docs/TESTING.md` как `TC-*` (после verify). Не мержить в каталог без отдельной задачи.
