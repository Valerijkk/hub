# Wave5 Edge E04 — Calls + Calendar + Storage

> **Тип:** только **Edge** (границы лимитов / TTL / размеров / дат / counts / media / sync)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave5  
> **Не трогает** `.planning/EDGE-CORNER-CATALOG.md`, wave1–4 `edge-expansions/**`  
> **ID range:** `ECC-*-500+` — **не пересекает** wave1 `*100+` / wave2 `*200+` / wave3 `*300+` / wave4 `*400+` / catalog `*0xx` / Corner `*C*`  
> **SFU:** LiveKit only (mediasoup / SEED-001 = REJECTED — строки N/A **не** добавлять)  
> **Связи:** NFR-03/04/08/12/20/27 · NF-09 · SEC-02 · US-M06 · US-CAL-01…04 · SEED-002 · JULY brief §3 · TC-CALL-* · TC-STORAGE-*

**Назначение:** foresight-бэклог **новых** граничных кейсов (углубление wave1–4). Happy / Neg / Corner сюда не входят.  
**Не копии** `*100–*400` (participants/token/duration-cap/bitrate/emptyTimeout/webhook JWT/invite length/INTERVAL-EXDATE/attendee max/0B-25MB/mime/skew/trash-30d/orphan batch/…).

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (as-is) / **Future** (CAL-SYNC / v2 / product-lock cap) |
| Слой | API / Unit / Manual / E2E |
| Связь | NFR / US / TC / Brief / константа кода / wave1–4 ECC (расширение, не копия) |

**Фокус wave5:** LiveKit ICE/TURN reconnect · simulcast/dynacast layers · Opus floor · Hub↔LK participant drift · egress fileSize/audio-missing · scheduled worker lookback · raise-hand/chat list caps · RRULE `BYDAY`/`BYHOUR`/`WKST` · DURATION↔DTEND · feed redirect hops · sub failure streak · METHOD=REPLY · inbox badge · NATS fanout batch · multipart parts · **bytes** quota · Cache-Control · file `version` · restore count.

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — `ECC-CALL-500+`

> Фокус: **ICE/TURN reconnect budget** · **simulcast/dynacast** · **Opus bitrate floor** · **Hub↔LiveKit participant_count drift** · **egress fileSize / audio-missing** · **scheduled worker lookback** · **raise-hand rate** · **chat list pagination**.  
> Wave1–4: participants/token/duration/bitrate/emptyTimeout/webhook body-JWT/invite-guest length/chat body size/pagination/retention — **не дублировать**.  
> LiveKit stay: Room + Egress RoomComposite → Seaweed; SEC-02.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-500 | Edge | ICE reconnect budget: disconnect duration = product reconnect window − 1s | media resumes same participant; Hub count unchanged | Manual | P0 | LiveKit ICE/TURN | Future |
| ECC-CALL-501 | Edge | ICE reconnect: disconnect = window + 1s | treated as leave; participant_left; rejoin needs new token | Manual | P0 | ICE reconnect window | Future |
| ECC-CALL-502 | Edge | TURN-only path (UDP blocked): join at 50−1 participants | media OK via TURN; 50th still joinable | Manual | P1 | NFR-03 × TURN | Future |
| ECC-CALL-503 | Edge | Simulcast layers = product max (e.g. 3) published | peers subscribe; egress picks configured layer; no encoder OOM | Manual | P1 | LiveKit simulcast | Future |
| ECC-CALL-504 | Edge | Dynacast: subscriber count = 1 vs = max−1 at room=50 | layer downshift/upshift within budget; no freeze >SLA | Manual | P2 | dynacast | Future |
| ECC-CALL-505 | Edge | Opus audio bitrate = product floor (min kbps) | audible; egress audio present | Manual | P1 | Opus floor | Future |
| ECC-CALL-506 | Edge | Opus bitrate request below floor | clamp to floor **или** reject publish — контракт | Manual | P1 | Opus floor boundary | Future |
| ECC-CALL-507 | Edge | Hub `participant_count` vs LiveKit room count differ by 1 (webhook lag) | reconcile within SLA; UI shows Hub SoT **или** LK — documented | API/Manual | P0 | participant drift | TODO |
| ECC-CALL-508 | Edge | Drift = product tolerance + 1 without reconcile | alert/fail-soft; no silent permanent skew | API | P0 | drift tolerance | Future |
| ECC-CALL-509 | Edge | `OnEgressEnded` `fileSize` = 0 with duration > 0 | status failed/quarantine; download not ready | API | P0 | fileSize=0, OnEgressEnded | TODO |
| ECC-CALL-510 | Edge | `fileSize` = MaxInt64-adjacent spoofed webhook | reject/clamp; SEC-02 still required; no DB overflow | API | P0 | SEC-02 × fileSize | TODO |
| ECC-CALL-511 | Edge | Egress completes with video-only (all mics muted / no audio track) | file ready; playable; UI marks silent OK | Manual | P1 | US-M06 egress | TODO |
| ECC-CALL-512 | Edge | Scheduled worker lookback = product window − 1s past due | still picks due call; room created | API | P0 | ProcessDueScheduled | TODO |
| ECC-CALL-513 | Edge | Due call older than lookback + 1s | skipped **или** catch-up job — no infinite backlog storm | API | P1 | worker lookback | Future |
| ECC-CALL-514 | Edge | Raise-hand rate: N-th raise at product/min boundary | N OK; N+1 → 429/ignore; no stuck hand flood | API | P2 | TC-CALL-10 rate | Future |
| ECC-CALL-515 | Edge | Raise-hand at participant=1 (solo) then leave | hand cleared; no orphan raised state | API | P2 | raise-hand cleanup | TODO |
| ECC-CALL-516 | Edge | Call-chat list `limit` = product max page | 200; ≤max msgs; oldest/newest per contract | API | P1 | ListMessages limit | TODO |
| ECC-CALL-517 | Edge | Call-chat list `limit` = max + 1 | clamp **или** 400 | API | P1 | chat pagination | TODO |
| ECC-CALL-518 | Edge | Call-chat list when messages = 0 | 200 []; no 500 | API | P2 | empty chat | TODO |
| ECC-CALL-519 | Edge | Room metadata / participant attributes payload = product max bytes | join OK; over max → truncate/reject at LiveKit/Hub | Manual/Unit | P2 | room metadata size | Future |
| ECC-CALL-520 | Edge | Screen-share resolution = product max px (width×height) | publish OK; above → LiveKit reject/clamp | Manual | P1 | US-M06 resolution | Future |
| ECC-CALL-521 | Edge | Concurrent stop-recording + leave last participant (same second) | egress terminal; room finish once; no stuck `recording` | Manual/API | P0 | maxActive × leave | TODO |

**Calls Edge в этом файле: 22** (`ECC-CALL-500`…`521`)

---

## 2. Calendar as-is + Calendar v2 — `ECC-CAL-500+`

> Фокус: **RRULE `BYDAY` / `BYHOUR` / `BYMINUTE` / `WKST`** · **DURATION vs DTEND** · **feed redirect hop limit** · **subscription failure streak** · **VTIMEZONE transition count** · **UID length**.  
> Wave1–4: range 366d / COUNT-UNTIL / INTERVAL-BYSETPOS-EXDATE / ETag / aggregator density / title length — **не дублировать**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-500 | Edge | RRULE `FREQ=WEEKLY;BYDAY=MO,WE,FR` (3 days) | instances only those weekdays; UTC store | Unit | P0 | BYDAY multi, US-CAL-02 | Future |
| ECC-CAL-501 | Edge | `BYDAY` count = product max weekdays (7) | expand OK; denser than daily INTERVAL | Unit | P1 | BYDAY max | Future |
| ECC-CAL-502 | Edge | `BYDAY` with invalid token (`XX`) / empty after strip | 400 | Unit/API | P0 | BYDAY validate | Future |
| ECC-CAL-503 | Edge | `FREQ=DAILY;BYHOUR=9,17` (2 hours) | 2 instances/day within range clip | Unit | P1 | BYHOUR | Future |
| ECC-CAL-504 | Edge | `BYHOUR=24` / negative | 400 | Unit | P0 | BYHOUR bounds | Future |
| ECC-CAL-505 | Edge | `BYMINUTE=0` vs `BYMINUTE=60` | 0 OK; 60 → 400 | Unit | P1 | BYMINUTE | Future |
| ECC-CAL-506 | Edge | `WKST=SU` vs `WKST=MO` with weekly BYDAY | week boundaries differ; documented | Unit | P1 | WKST | Future |
| ECC-CAL-507 | Edge | `WKST` invalid (`XX`) | 400 | Unit/API | P1 | WKST validate | Future |
| ECC-CAL-508 | Edge | VEVENT with `DURATION=PT1H` and no DTEND | end = start+1h; list OK | Unit/API | P0 | DURATION | Future |
| ECC-CAL-509 | Edge | Both `DURATION` and `DTEND` set | 400 prefer one **или** RFC one-wins — зафиксировать | Unit | P0 | DURATION×DTEND | Future |
| ECC-CAL-510 | Edge | `DURATION=PT0S` / negative | 400 | Unit/API | P0 | DURATION=0 | Future |
| ECC-CAL-511 | Edge | `DURATION` = product max event length (ровно) | OK; max+1 → 400 | API | P1 | duration max | Future |
| ECC-CAL-512 | Edge | iCal feed URL: redirect hops = product max (e.g. 3) | follow ≤max; parse body | API | P0 | US-CAL-03, NFR-21 | Future |
| ECC-CAL-513 | Edge | Redirect hops = max + 1 | abort; no SSRF open redirect chain | API | P0 | ValidateOutboundURL × hops | Future |
| ECC-CAL-514 | Edge | Subscription consecutive failures = streak max − 1 | still active; next poll scheduled | API | P1 | feed failure streak | Future |
| ECC-CAL-515 | Edge | Failures = streak max | auto-disable/pause; user notified; no poll storm | API | P0 | failure streak cap | Future |
| ECC-CAL-516 | Edge | VTIMEZONE transition count = product max in one TZ | export OK; over max → simplify/reject | Unit | P2 | VTIMEZONE size | Future |
| ECC-CAL-517 | Edge | VEVENT `UID` length = max vs max+1 | max OK; max+1 400; no truncate mid-UID | API | P1 | UID length | Future |
| ECC-CAL-518 | Edge | `FREQ=YEARLY;BYMONTH=2;BYMONTHDAY=29` across leap→non-leap window in 366d list | leap days only; no Feb 30 | Unit | P0 | leap BYMONTHDAY | Future |
| ECC-CAL-519 | Edge | List aggregator: single event `starts_at` exactly = `to` boundary | included/excluded per validateRange exclusivity — documented | API | P1 | range edge inclusivity | TODO |
| ECC-CAL-520 | Edge | Resource capacity=N: N bookings touch end-to-start chain (no overlap) | all OK; N+1 overlapping → 409 | API | P0 | US-CAL-01 chain | Future |

**Calendar Edge в этом файле: 21** (`ECC-CAL-500`…`520`)

---

## 3. CAL-SYNC Edge — `ECC-CAL-E500+`

> Фокус: **ICS METHOD=REPLY/COUNTER** · **inbox badge cap** · **tentative soft-busy hold** · **NATS invite fanout batch** · **RSVP note length** · **exception instance count** · **sync lag budget**.  
> Wave1–4: attendee max / RSVP TTL / reminder lead / SEQUENCE / soft-conflict minutes / slot grain — **не дублировать**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-E500 | Edge | ICS `METHOD=REPLY` attachment size = mail limit − 1 | hub-mail sends; organizer inbox updates | Manual | P1 | Brief L0 REPLY | Future |
| ECC-CAL-E501 | Edge | `METHOD=REPLY` size = limit + 1 | truncate/link-fallback/4xx — no silent drop | Manual/API | P1 | L0 size | Future |
| ECC-CAL-E502 | Edge | `METHOD=COUNTER` (propose new time) at v1 scope | 400 unsupported **или** Future accept — ADR | API | P2 | COUNTER scope | Future |
| ECC-CAL-E503 | Edge | Inbox badge count = product soft-cap (ровно) | badge shows cap; all invites still listable | API/E2E | P1 | inbox badge | Future |
| ECC-CAL-E504 | Edge | Badge count = soft-cap + 1 pending invites | badge `cap+` **или** exact; no overflow UI | E2E | P1 | badge overflow | Future |
| ECC-CAL-E505 | Edge | Tentative soft-busy hold duration = product max | slot soft-busy until TTL; then needs_action policy | API | P1 | tentative hold | Future |
| ECC-CAL-E506 | Edge | Tentative hold = max + 1s past | hold expired; availability free/soft per policy | API | P1 | tentative TTL | Future |
| ECC-CAL-E507 | Edge | NATS invite fanout batch = product batch size (ровно N attendees) | one batch; all delivered | API | P0 | NATS fanout, C11 | Future |
| ECC-CAL-E508 | Edge | Fanout attendees = batch + 1 | ≥2 batches; no silent loss of last | API | P0 | fanout batch edge | Future |
| ECC-CAL-E509 | Edge | RSVP response note/comment length = max vs max+1 | max OK; max+1 400 | API | P2 | RSVP note | Future |
| ECC-CAL-E510 | Edge | Series exceptions (EXDATE/RDATE instances) with per-instance RSVP count = max | each exception RSVP independent **или** v1 series-only | API | P1 | Brief C2 × exceptions | Future |
| ECC-CAL-E511 | Edge | Exception count = max + 1 on series | 400; series unchanged | API | P1 | exception cap | Future |
| ECC-CAL-E512 | Edge | Dual-calendar sync lag budget: accept visible on peer calendar within SLA − 1s | both grids show; badge clear | E2E | P0 | SEED-002 sync lag | Future |
| ECC-CAL-E513 | Edge | Sync lag > SLA + 1s | retry/reconcile; no permanent split-brain | API/E2E | P0 | sync lag SLA | Future |
| ECC-CAL-E514 | Edge | Organizer transfer at attendee count = max | new organizer counted per policy; invites ≤max | API | P1 | Brief C5 transfer | Future |
| ECC-CAL-E515 | Edge | Accept at event starts_at − 1s vs + 1s | −1s OK; +1s allow late RSVP **или** 410 — контракт | API | P1 | late RSVP boundary | Future |
| ECC-CAL-E516 | Edge | Availability: max concurrent soft-conflicts in one slot = product warn cap | warnings bounded; create still OK | API | P2 | soft-conflict storm | Future |

**CAL-SYNC Edge в этом файле: 17** (`ECC-CAL-E500`…`516`)

---

## 4. Storage / Files — `ECC-STOR-500+`

> Фокус: **multipart part size/count** · **project bytes quota** · **Cache-Control / Content-Encoding** · **file optimistic `version`** · **restore count** · **concurrent download URLs** · **SSE metadata**.  
> Wave1–4: BodyLimit/0B/25MB/mime/skew/trash-30d/orphan batch/filename length/sliding rate/size±1 — **не копировать**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-STOR-500 | Edge | Multipart upload: part size = product min (e.g. 5 MB) exactly | parts accepted; complete OK | Manual/API | P1 | multipart min part | Future |
| ECC-STOR-501 | Edge | Part size = min − 1 (except last) | S3/Hub reject; no ready | Manual | P1 | multipart part min | Future |
| ECC-STOR-502 | Edge | Part count = product max parts | complete OK; object size = sum | API | P1 | multipart max parts | Future |
| ECC-STOR-503 | Edge | Part count = max + 1 | 400; abort upload; no orphan ready | API | P0 | multipart max | Future |
| ECC-STOR-504 | Edge | Project **bytes** used = soft-cap − 1 | upload-url OK | API | P0 | bytes quota | Future |
| ECC-STOR-505 | Edge | Bytes used = soft-cap; next file size 1 byte | 422/403; no new pending | API | P0 | bytes quota edge | Future |
| ECC-STOR-506 | Edge | Bytes used = soft-cap − 10; upload declared 25 MB | reject if would exceed **или** allow then complete-check — documented | API | P1 | quota × NFR-27 | Future |
| ECC-STOR-507 | Edge | Download `Cache-Control` max-age = downloadURLTTL (1h) | header ≤TTL; no cache beyond presign life | Manual | P2 | downloadURLTTL=1h | TODO |
| ECC-STOR-508 | Edge | `Content-Encoding: gzip` on upload metadata | accept+store **или** 400 — контракт; download consistent | API | P2 | Content-Encoding | Future |
| ECC-STOR-509 | Edge | Complete with stale `version` (optimistic lock) = n vs n−1 | n OK; n−1 → 409 conflict | API | P0 | version field | TODO |
| ECC-STOR-510 | Edge | Complete `version` = MaxInt32-adjacent spoof | reject/clamp; no overflow | Unit/API | P1 | version bounds | TODO |
| ECC-STOR-511 | Edge | Restore from trash count = product max restores per file | last restore OK | API | P2 | restore count | Future |
| ECC-STOR-512 | Edge | Restore count = max + 1 | 422; stays in trash **или** purged policy | API | P2 | restore cap | Future |
| ECC-STOR-513 | Edge | Concurrent download-url issuances = product max parallel | all URLs valid ≤TTL; max+1 → 429 if capped | API | P1 | download URL concurrency | Future |
| ECC-STOR-514 | Edge | SSE-S3 / encryption metadata present on Seaweed object | Hub download still works; metadata not leaked to unauthorized | Manual | P2 | SSE metadata | Future |
| ECC-STOR-515 | Edge | Attach same file id to issue + message (multi-parent) at refcount = max | OK; max+1 link → 400 | API | P1 | file refcount | Future |
| ECC-STOR-516 | Edge | Abort multipart after K parts uploaded; orphan age = cutoff + 1s | parts GC; no billable leak forever | API | P1 | multipart abort × orphan | Future |
| ECC-STOR-517 | Edge | Presign PUT Content-MD5 header mismatch vs body | S3 400; Hub complete never ready | Manual | P1 | Content-MD5 | TODO |
| ECC-STOR-518 | Edge | Filename only whitespace / ZWSP after sanitize → empty | 400; no empty key | API | P0 | sanitizeFilename empty | TODO |
| ECC-STOR-519 | Edge | Download Range `bytes=0-0` (1 byte) vs `bytes=0-` full | 206/200 per policy; size integrity | Manual | P2 | Range requests | TODO |
| ECC-STOR-520 | Edge | Project quota bytes = 0 (admin lock) | all upload-url 403/422; downloads OK | API | P1 | quota zero | Future |

**Storage Edge в этом файле: 21** (`ECC-STOR-500`…`520`)

---

## Счётчик

| Домен | ID range | Кол-во Edge |
|-------|----------|-------------|
| Calls + LiveKit (ICE/simulcast/Opus/drift/egress/worker/chat list) | `ECC-CALL-500`…`521` | **22** |
| Calendar (BYDAY/BYHOUR/WKST/DURATION/redirect/streak/UID) | `ECC-CAL-500`…`520` | **21** |
| CAL-SYNC (REPLY/badge/tentative/fanout/exceptions/sync lag) | `ECC-CAL-E500`…`516` | **17** |
| Storage (multipart/bytes quota/Cache-Control/version/restore) | `ECC-STOR-500`…`520` | **21** |
| **Итого** | | **81** |

| Проверки | Статус |
|----------|--------|
| Только тип **Edge** | ✅ |
| Минимум 50 строк | ✅ 81 ≥ 50 |
| ID `*500+` без пересечения `*100+`…`*400+` | ✅ |
| LiveKit stay / mediasoup N/A | ✅ не добавлено |
| Основной `EDGE-CORNER-CATALOG.md` | ✅ не изменён |
| Git commit | ✅ не создавался |

**Фокус coverage (wave5):**

| Тема | ID (примеры) |
|------|----------------|
| ICE/TURN reconnect | `ECC-CALL-500`…`502` |
| Simulcast / dynacast / Opus | `ECC-CALL-503`…`506`, `520` |
| Hub↔LK participant drift | `ECC-CALL-507`…`508` |
| Egress fileSize / silent | `ECC-CALL-509`…`511`, `521` |
| Worker lookback / raise-hand / chat list | `ECC-CALL-512`…`518` |
| BYDAY / BYHOUR / WKST | `ECC-CAL-500`…`507` |
| DURATION ↔ DTEND | `ECC-CAL-508`…`511` |
| Feed redirect / failure streak | `ECC-CAL-512`…`515` |
| METHOD=REPLY / badge / tentative | `ECC-CAL-E500`…`506` |
| NATS fanout / exceptions / sync lag | `ECC-CAL-E507`…`513` |
| Multipart parts | `ECC-STOR-500`…`503`, `516` |
| Bytes quota | `ECC-STOR-504`…`506`, `520` |
| version / restore / Range | `ECC-STOR-509`…`512`, `519` |

**Следующий шаг:** `qa-manual-ver` → выборочно промоутить P0 в `docs/TESTING.md` как `TC-*` (после verify). Не мержить в каталог без отдельной задачи.
