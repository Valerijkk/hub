# Wave2 Edge E04 — Calls + Calendar + Storage

> **Тип:** только **Edge** (границы лимитов / TTL / размеров / дат / counts / skew)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave2  
> **Не трогает** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` и wave1 `edge-expansions/04-*.md`  
> **ID range:** `ECC-*-200+` — **не пересекает** wave1 `*100+` / catalog `*0xx` / Corner `*C*`  
> **SFU:** LiveKit only (mediasoup / SEED-001 = REJECTED — строки N/A **не** добавлять)  
> **Связи:** NFR-03/04/08/12/27 · NF-09 · US-M06 · US-CAL-01…04 · SEED-002 · JULY brief §3 · TC-CALL-* · TC-STORAGE-*

**Назначение:** foresight-бэклог **новых** граничных кейсов (углубление wave1). Happy / Neg / Corner сюда не входят.

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (as-is) / **Future** (CAL-SYNC / v2 / product-lock cap) |
| Слой | API / Unit / Manual / E2E |
| Связь | NFR / US / TC / Brief / константа кода / wave1 ECC (расширение, не копия) |

**Фокус wave2:** recording duration cap · RRULE `UNTIL`/`COUNT` boundary · iCal feed max events · attendee max · 0-byte / 25 MB / 25 MB+1 · mime edge · presign clock skew.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — `ECC-CALL-200+`

> Фокус: **recording duration cap** (+ смежные границы egress/metadata).  
> Wave1 `ECC-CALL-100…125` = participants/token TTL/active recordings/RRULE scheduled/conflict — **не дублировать**.  
> LiveKit stay: egress RoomComposite → Seaweed; duration из webhook `OnEgressEnded`.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-200 | Edge | Recording duration ровно product/LiveKit soft-cap (напр. 4h / documented max) | egress auto-stop или Hub stop; status=completed; durationSec = cap | Manual/API | P0 | duration cap, OnEgressEnded | Future |
| ECC-CALL-201 | Edge | Recording duration = cap − 1s | recording continues; no premature stop | Manual | P0 | duration cap boundary | Future |
| ECC-CALL-202 | Edge | Recording duration = cap + 1s (late webhook) | Hub clamps/stores ≤cap **или** accepts actual; no negative/overflow | API/Unit | P0 | durationSec int | TODO |
| ECC-CALL-203 | Edge | Recording duration = 0s (egress failed instantly) | status=failed/empty; download 404; no 500 | API | P1 | OnEgressEnded duration=0 | TODO |
| ECC-CALL-204 | Edge | Recording duration fractional webhook (1.9s → int) | durationSec truncated/rounded per contract; metadata consistent | Unit | P2 | duration float→int | TODO |
| ECC-CALL-205 | Edge | Recording duration = MaxInt32-adjacent huge value from spoofed webhook | reject/clamp; SEC-02 still required; no DB overflow | API | P0 | SEC-02, durationSec | TODO |
| ECC-CALL-206 | Edge | Two active recordings both hit duration cap same second | both finalize; activeCount→0; 3-й start allowed after | API | P1 | maxActive=2 × duration cap | Future |
| ECC-CALL-207 | Edge | Duration cap stop while participant count = max (50) | egress stops; call stays active; media OK | Manual | P1 | NFR-03 × duration cap | Future |
| ECC-CALL-208 | Edge | Duration cap reached during screen-share-only egress | file finalized; screen track in recording; call continues | Manual | P1 | US-M06, TC-CALL-07 | Future |
| ECC-CALL-209 | Edge | Recording file size 0 bytes after duration>0 (corrupt egress) | failed/quarantine; no download of empty as «ready» | API | P1 | 0-byte recording | TODO |
| ECC-CALL-210 | Edge | Recording download presign issued at TTL−1s vs Hub clock skew ±30s | within skew window: GET OK; beyond: 403 — policy documented | Manual | P1 | downloadURLTTL=1h, skew | TODO |
| ECC-CALL-211 | Edge | Recording metadata durationSec vs media probe length differ by 1s | UI shows Hub SoT **или** probe; no crash | Manual | P2 | duration display | TODO |
| ECC-CALL-212 | Edge | Start recording at call age = member JWT 24h−60s; duration cap > remaining JWT | rejoin renews JWT; egress not orphaned by token expiry alone | Manual | P1 | token 24h × duration cap | Future |
| ECC-CALL-213 | Edge | Guest (4h JWT) in room when recording hits duration cap | guest stays until guest TTL; recording download guest-denied | API/Manual | P1 | guest 4h, TC-CALL-04 | Future |
| ECC-CALL-214 | Edge | List project recordings: duration filter boundary (min=0 / max=cap) | inclusive bounds; cap+1 excluded | API | P2 | ListProjectRecordings | TODO |
| ECC-CALL-215 | Edge | LiveKit egress auto-stop at platform max vs Hub soft-cap mismatch | Hub state matches final webhook; no «recording» forever | Manual | P0 | LiveKit stay, duration cap | Future |

**Calls Edge в этом файле: 16** (`ECC-CALL-200`…`215`)

---

## 2. Calendar as-is + Calendar v2 — `ECC-CAL-200+`

> Фокус: **RRULE `UNTIL`/`COUNT` boundary** · **iCal feed / export max events**.  
> Wave1 `ECC-CAL-100…121` = range 366d, DST, basic COUNT/UNTIL, LimitReader body — здесь **углубление границ**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-200 | Edge | RRULE `COUNT` = product max expand (ровно max) | max instances; create/expand OK | Unit/API | P0 | US-CAL-02, COUNT cap | Future |
| ECC-CAL-201 | Edge | RRULE `COUNT` = max + 1 | 400/422; series not saved | API | P0 | US-CAL-02 AC | Future |
| ECC-CAL-202 | Edge | RRULE `COUNT` = 2 (min meaningful >1) | ровно 2 occurrences | Unit | P1 | parseRRule | TODO |
| ECC-CAL-203 | Edge | RRULE `UNTIL` = last allowed occurrence timestamp (inclusive RFC) | last instance included; next not | Unit | P0 | RFC 5545 UNTIL | TODO |
| ECC-CAL-204 | Edge | RRULE `UNTIL` = one tick before first DTSTART | 0 instances **или** 400 — контракт | Unit/API | P0 | UNTIL < DTSTART | Future |
| ECC-CAL-205 | Edge | RRULE `UNTIL` + `COUNT` вместе (оба заданы) | 400 prefer one **или** RFC «more restrictive wins» — зафиксировать | Unit | P1 | RRULE conflict params | Future |
| ECC-CAL-206 | Edge | RRULE `FREQ=DAILY;COUNT=N` где N×interval пересекает maxRangeDays=366 clip | expand clipped to list range; no OOM | Unit | P0 | maxRangeDays=366 | TODO |
| ECC-CAL-207 | Edge | RRULE `UNTIL` exactly at UTC midnight vs local TZ day boundary | last occurrence correct in UTC store | Unit | P0 | NFR-12 | TODO |
| ECC-CAL-208 | Edge | RRULE `COUNT` huge (1e9) without UNTIL | reject at validate; never expand | API/Unit | P0 | DoS expand | Future |
| ECC-CAL-209 | Edge | iCal **export** max events = product cap (ровно max VEVENT) | 200; complete VCALENDAR; no truncate mid-VEVENT | API | P0 | ExportICal max events | Future |
| ECC-CAL-210 | Edge | iCal export events = max + 1 in range | 400 **или** truncate+header warning — no OOM | API | P0 | US-CAL-03 edge | Future |
| ECC-CAL-211 | Edge | iCal export events = 0 (empty) vs max | empty: 0 VEVENT valid; max: max VEVENT | API | P1 | ECC-CAL-105 deepen | TODO |
| ECC-CAL-212 | Edge | iCal **feed import** max events parsed = cap | accept ≤cap; state OK | API | P0 | US-CAL-03, L1 | Future |
| ECC-CAL-213 | Edge | iCal feed import events = cap + 1 | reject/error; no partial commit of cap+1 | API | P0 | feed max events | Future |
| ECC-CAL-214 | Edge | iCal feed body size OK but VEVENT count = cap (dense small events) | count cap enforced independent of LimitReader bytes | API | P0 | count vs byte cap | Future |
| ECC-CAL-215 | Edge | iCal feed: LimitReader byte cap hit with VEVENT count < max | byte-cap error wins; no half-event | API | P0 | ECC-CAL-117/118 deepen | Future |
| ECC-CAL-216 | Edge | Export.ics: recurring series COUNT=max expanded into range | instances ≤ min(COUNT, range, export-cap) | Unit/API | P1 | COUNT × export cap | Future |
| ECC-CAL-217 | Edge | `UNTIL` in floating local vs UTC form (`UNTIL=YYYYMMDD` vs `...Z`) | parse OK; last day inclusive per RFC | Unit | P1 | rrule-go, NFR-12 | TODO |

**Calendar Edge в этом файле: 18** (`ECC-CAL-200`…`217`)

---

## 3. CAL-SYNC Edge — `ECC-CAL-E200+`

> Фокус: **attendee max** (+ смежные границы invite list).  
> Wave1 `ECC-CAL-E100…114` = max/max+1/0/1 attendees, ICS size, poll — здесь **граничные вариации** без копирования.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-E200 | Edge | Attendees = hard cap − 1 | create OK; all invited; NATS × (cap−1) | API | P0 | SEED-002, Brief | Future |
| ECC-CAL-E201 | Edge | Attendees = hard cap (ровно max) | create OK; no silent drop | API | P0 | Brief C4/C11 | Future |
| ECC-CAL-E202 | Edge | Attendees = hard cap + 1 (explicit list) | 400/422; event not created | API | P0 | attendee max | Future |
| ECC-CAL-E203 | Edge | Channel-default attendees: channel size = attendee hard cap | all members required; OK | API | P0 | Brief C11 | Future |
| ECC-CAL-E204 | Edge | Channel-default: channel size = hard cap + 1 | 422 **или** require explicit subset — no truncate | API | P0 | Brief C11 edge | Future |
| ECC-CAL-E205 | Edge | Add attendee when already at max (patch) | 422; existing set unchanged | API | P0 | attendee max | Future |
| ECC-CAL-E206 | Edge | Remove one then add one at max (swap) | OK; count stays ≤ max | API | P1 | attendee max | Future |
| ECC-CAL-E207 | Edge | Duplicate attendee IDs in request totaling unique ≤ max but raw len > max | dedupe then validate; OK if unique≤max | API | P1 | validation order | Future |
| ECC-CAL-E208 | Edge | Duplicate IDs → unique = max + 1 after dedupe? (raw duplicates inflate) | validate on unique set; reject if unique>max | API | P1 | attendee max | Future |
| ECC-CAL-E209 | Edge | Organizer counted / not counted toward attendee max | policy documented; boundary uses same rule | API | P0 | Brief C5 | Future |
| ECC-CAL-E210 | Edge | External email-only attendees toward max (C14 rejected v1) | 0 external; only Hub users count | API | P1 | Brief C14 | Future |
| ECC-CAL-E211 | Edge | Max attendees + ICS attachment size at mail limit | send OK; all recipients; no drop last N | Manual | P1 | L0 × attendee max | Future |
| ECC-CAL-E212 | Edge | Max attendees storm: NATS/notif fanout rate boundary | all invites eventually; backoff; no silent loss | API | P1 | NFR notif, C11 | Future |
| ECC-CAL-E213 | Edge | Availability query with max-attendee event (busy blocks) | all accepted busy visible; needs_action not hard-busy | API | P1 | §3.2, ECC-CAL-C20 | Future |

**CAL-SYNC Edge в этом файле: 14** (`ECC-CAL-E200`…`213`)

---

## 4. Storage / Files — `ECC-STOR-200+`

> Фокус: **0-byte / 25 MB / 25 MB+1** · **mime edge** · **presign clock skew**.  
> Wave1 `ECC-STOR-100…120` = базовые BodyLimit/rate/TTL/unicode — здесь **углубление** (не копия тех же строк).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-STOR-200 | Edge | Presign complete: declared size = 0 bytes, S3 object 0 bytes | allow **или** 400 единый контракт; pending→terminal | API | P0 | 0-byte, ECC-STOR-110 deepen | TODO |
| ECC-STOR-201 | Edge | Presign complete: declared size = 0, S3 object > 0 (lie) | reject/mismatch; no ready with wrong size | API | P0 | size integrity | TODO |
| ECC-STOR-202 | Edge | Presign complete: declared = 25 MB (26214400), S3 exact 25 MB | ready; metadata size exact | API | P0 | NFR-04/27 | TODO |
| ECC-STOR-203 | Edge | Presign complete: declared = 25 MB, S3 = 25 MB + 1 (NF-09) | reject complete **или** quarantine; risk logged | API | P0 | NF-09 | TODO |
| ECC-STOR-204 | Edge | Presign PUT body = 0 bytes then complete | same policy as ECC-STOR-200; no orphan ready | Manual/API | P1 | 0-byte PUT | TODO |
| ECC-STOR-205 | Edge | Presign PUT body = 25 MB − 1 | S3 OK; complete OK | Manual | P1 | NFR-27 boundary | TODO |
| ECC-STOR-206 | Edge | Presign PUT body = 25 MB + 1 without S3 bucket policy | **may** accept (NF-09); Hub complete must not trust blindly | Manual | P0 | NF-09, ECC-STOR-104 deepen | TODO |
| ECC-STOR-207 | Edge | Backend buffer path: Content-Length = 25 MB + 1, body truncated mid-read | 413; no partial file row ready | API | P0 | BodyLimit | TODO |
| ECC-STOR-208 | Edge | Backend buffer: Content-Length = 25 MB, actual body 25 MB + 1 | 413/400; fail-closed | API | P0 | BodyLimit | TODO |
| ECC-STOR-209 | Edge | Mime: `Content-Type` with parameters only (`text/plain; charset=utf-8; x=y`) | accept primary type **или** 400 — контракт | API | P1 | mime edge | TODO |
| ECC-STOR-210 | Edge | Mime: uppercase / mixed (`Image/PNG`, `APPLICATION/PDF`) | normalized store; download CT correct | API | P1 | mime normalize | TODO |
| ECC-STOR-211 | Edge | Mime: unknown extension + valid mime (`file.bin` + `application/pdf`) | policy accept-as-declared **или** reject | API | P1 | mime×ext | TODO |
| ECC-STOR-212 | Edge | Mime: double extension (`report.pdf.exe`) + `application/pdf` | reject **или** sanitize; no executable serve as pdf | API | P0 | mime edge security | TODO |
| ECC-STOR-213 | Edge | Mime: `Content-Type: */*` or `invalid` token | 400 | API | P1 | mime validation | TODO |
| ECC-STOR-214 | Edge | Mime sniff vs declared: magic bytes PNG, CT `text/plain` | store declared **или** sniff override — documented | API | P1 | ECC-STOR-118 deepen | TODO |
| ECC-STOR-215 | Edge | Presign skew: client clock −5m vs Hub; upload URL used at Hub-now within 5m TTL | PUT OK if SigV4 skew window allows | Manual | P1 | uploadURLTTL=5m, skew | TODO |
| ECC-STOR-216 | Edge | Presign skew: client clock +10m; URL issued Hub-now, used immediately on skewed client | SigV4 may 403; document AWS skew ±15m interaction | Manual | P1 | presign skew | TODO |
| ECC-STOR-217 | Edge | Presign skew: Hub clock vs Seaweed ±N; URL at TTL−1s Hub-time | GET/PUT success iff within S3 skew+TTL; no Hub 200 with dead URL | Manual | P0 | upload 5m / download 1h | TODO |
| ECC-STOR-218 | Edge | Re-presign download after skew-induced 403 at TTL edge | new URL works; old stays dead | API/Manual | P1 | downloadURLTTL=1h | TODO |
| ECC-STOR-219 | Edge | Complete called with size=25MB+1 while object missing | 400/404; pending not → ready | API | P1 | complete validation | TODO |
| ECC-STOR-220 | Edge | Rate 10/min + 0-byte upload-url spam | 10th OK; 11th 429; empty intent still counts | API | P1 | NFR-08 × 0-byte | TODO |

**Storage Edge в этом файле: 21** (`ECC-STOR-200`…`220`)

---

## Счётчик

| Домен | ID range | Кол-во Edge |
|-------|----------|-------------|
| Calls + LiveKit (duration cap) | `ECC-CALL-200`…`215` | **16** |
| Calendar (RRULE / iCal max events) | `ECC-CAL-200`…`217` | **18** |
| CAL-SYNC (attendee max) | `ECC-CAL-E200`…`213` | **14** |
| Storage (0B / 25MB / mime / skew) | `ECC-STOR-200`…`220` | **21** |
| **Итого** | | **69** |

| Проверки | Статус |
|----------|--------|
| Только тип **Edge** | ✅ |
| Минимум 50 строк | ✅ 69 ≥ 50 |
| ID `*200+` без пересечения `*100+` | ✅ |
| LiveKit stay / mediasoup N/A | ✅ не добавлено |
| Основной `EDGE-CORNER-CATALOG.md` | ✅ не изменён |
| Git commit | ✅ не создавался |

**Фокус coverage (wave2):**

| Тема | ID (примеры) |
|------|----------------|
| Recording duration cap | `ECC-CALL-200`…`208`, `212`–`215` |
| RRULE UNTIL/COUNT boundary | `ECC-CAL-200`…`208`, `216`–`217` |
| iCal feed/export max events | `ECC-CAL-209`…`215` |
| Attendee max | `ECC-CAL-E200`…`213` |
| 0-byte / 25MB / 25MB+1 | `ECC-STOR-200`…`208`, `219`–`220`; CALL-209 |
| Mime edge | `ECC-STOR-209`…`214` |
| Presign skew | `ECC-CALL-210`, `ECC-STOR-215`…`218` |

**Следующий шаг:** `qa-manual-ver` → выборочно промоутить P0 в `docs/TESTING.md` как `TC-*` (после verify). Не мержить в каталог без отдельной задачи.
