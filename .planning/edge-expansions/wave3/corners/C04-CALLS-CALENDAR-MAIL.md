# Corner Expansion C04 wave3 — Calls + Calendar + Mail

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave3  
> **Не трогает** `.planning/EDGE-CORNER-CATALOG.md`, parent `00-CORNER-PARENT.md`, wave1 `corners/C04-*`, wave2 `wave2/corners/C04-*`  
> **ID range:** `ECC-*-C500+` (не пересекает parent `C200–C215`, wave1 `C300+`, wave2 `C400+`, catalog `C01–C25`, edge `*100+`)  
> **SFU:** LiveKit stay — mediasoup / SEED-001 = **N/A** (строки mediasoup не добавлять)  
> **Фокус wave3:** multi-device×ICE partition×Seaweed egress · all-day/EXDATE×LK worker×WS reorder · MIME/OOO/forwarded ICS×notif mute×rate fanout · organizer transfer×ACL mid-call · NATS invite lag×room start  
> **Источники:** JULY brief §3 · SEED-002 · US-M06 · US-CAL-* · NFR-07 · NFR-12 · SEC-02 · NF-02 · GDPR soft/hard delete

**Назначение:** foresight-бэклог **новых** corner-комбинаций wave3 для будущих `TC-*`. Happy / Neg / Edge сюда не входят. **Не копировать** сценарии C200/C300/C400 — только новые мультифакторные связки.

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (ещё не в каталоге) / **Future** (CAL-SYNC / mail L*) |
| Слой | API / Unit / Manual / E2E / WS |
| Связь | Brief C# · SEED-002 · TC · NFR · theme (расширение, не копия) |

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — multi-device × ICE × Seaweed egress × NATS lag — `ECC-CALL-C500+`

> Wave3: один identity на 2+ устройствах, ICE/network partition, SeaweedFS недоступен на egress, NATS `calendar.invite` lag vs room start, organizer transfer mid-call.  
> Wave2 C400 = RSVP×recording×SFU restart; здесь — **другие** оси (не restart-as-primary).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-C500 | Corner | Same user desktop+mobile both joined LiveKit; desktop ICE hard-fail; mobile stays; host End Call from desktop | single ended; mobile lands ended UI; no zombie room; participant identity counted once | Manual | P0 | multi-device×ICE×end | Future |
| ECC-CALL-C501 | Corner | Multi-device: Accept on mobile → join desktop only → mobile later joins mid-egress recording | one identity in attendee list; ≤1 participant slot per user policy; recording meta not duplicated | API/Manual | P0 | multi-device×Accept×egress | Future |
| ECC-CALL-C502 | Corner | ICE partition 30s (both peers) during active egress to Seaweed; Seaweed returns 503 on part upload | call stays active; egress retry/fail-closed; no false call-ended; Seaweed orphan parts cleaned | Manual | P0 | ICE×Seaweed 503×egress | TODO |
| ECC-CALL-C503 | Corner | SeaweedFS full/unavailable **before** egress start; host taps Record; LiveKit room healthy | recording start 503/fail; room continues; UI error; no half-ready download | API/Manual | P0 | Seaweed down×record start | TODO |
| ECC-CALL-C504 | Corner | Egress complete webhook OK but Seaweed GET for download fails (object missing) | recording state failed/unavailable; no 200 empty; audit; call history intact | API | P0 | egress OK×Seaweed miss | TODO |
| ECC-CALL-C505 | Corner | NATS JetStream lag: `calendar.invite.accepted` delayed 5m while scheduled worker already started LiveKit room | room exists; late Accept still binds same call_id; no second room; inbox Join works | API | P0 | NATS lag×LK worker | Future |
| ECC-CALL-C506 | Corner | NATS invite.created lost (ack timeout) + manual messenger Start Call + calendar Accept later | one room or explicit link policy; Accept does not spawn duplicate LiveKit; conflict if dual | API | P0 | NATS loss×manual×Accept | Future |
| ECC-CALL-C507 | Corner | Organizer transfer (new host) mid-call while old host device ICE-dead + recording active | new host can End/Record stop; old host token demoted; egress ownership clear | API/Manual | P0 | transfer×ICE×egress | Future |
| ECC-CALL-C508 | Corner | Organizer transfer + channel ACL revoke of **new** host mid-transfer | fail-closed: transfer aborted or rolled back; room ACL consistent; NF-02 | API | P0 | transfer×ACL×NF-02 | Future |
| ECC-CALL-C509 | Corner | Co-host End Call while primary host starts second recording (max=2) on multi-device | terminal ended; both egress stop; no third start; version race → one winner | API | P0 | co-host end×maxRec | TODO |
| ECC-CALL-C510 | Corner | Waiting-room / max-participants-1: late Accept + guest link + ICE reconnect burst | count ≤ max; excess 403; accepted member priority documented; no negative count | Manual | P0 | capacity×Accept×guest×ICE | Future |
| ECC-CALL-C511 | Corner | Mute-all storm (host) + unmute race on second device + egress metadata flags | mute state reconciled per identity; recording not crash; UI no flicker loop | FE/Manual | P1 | mute storm×multi-device×egress | Future |
| ECC-CALL-C512 | Corner | Camera permission revoke mid-call (browser) while screen-share + egress ON | audio continues; video track unpublished; egress continues without camera; no room drop | Manual | P1 | cam revoke×share×egress | Future |
| ECC-CALL-C513 | Corner | LiveKit data-channel chat message + messenger thread message same ms + Accept from third user | no cross-leak of data-channel into Mongo messenger; Accept independent; NF-02 rooms | API/WS | P1 | data-ch×msg×Accept | Future |
| ECC-CALL-C514 | Corner | Caddy/proxy idle timeout on signaling WS during long silent call + ICE keep-alive | reconnect without ending call; participant count stable; SEC-02 webhooks still valid | Manual | P0 | proxy timeout×ICE×LK | TODO |
| ECC-CALL-C515 | Corner | Webhook `track_published` / `track_unpublished` out-of-order after ICE recover | Hub state eventual-consistent; no crash; recording track list sane | API | P1 | webhook reorder×ICE | TODO |
| ECC-CALL-C516 | Corner | Soft-delete call row while LiveKit room still has participants (ops mistake) + Seaweed egress uploading | fail-closed: kick/end or block delete; no orphan room without DB; egress cancelled | API | P0 | soft-del×live room×egress | Future |
| ECC-CALL-C517 | Corner | GDPR hard-delete user who is **only** on mobile leg of multi-device join + desktop still connected | both legs kicked; PII stripped; desktop cannot republish as deleted user | API/Manual | P0 | GDPR×multi-device | Future |
| ECC-CALL-C518 | Corner | Reminder NATS fires Join CTA while user already in room on other device + notif prefs mute push | no duplicate room; CTA opens existing; push suppressed per prefs; in-app OK | API/E2E | P1 | reminder×multi-device×mute | Future |
| ECC-CALL-C519 | Corner | Scheduled all-day calendar event wrongly linked to LiveKit worker tick | worker must **not** create room for all-day; log/metric; no ghost call | API | P0 | all-day×LK worker | Future |
| ECC-CALL-C520 | Corner | EXDATE removed today's occurrence + worker still has stale due snapshot + Accept on series | no room for EXDATE day; Accept series does not resurrect removed instance room | API | P0 | EXDATE×worker×Accept | Future |
| ECC-CALL-C521 | Corner | LiveKit room_finished for call A arrives after user started call B in same channel (different room_id) | A webhook cannot end B; bind by room/call id; SEC-02 | API | P0 | room_id bind×channel reuse | TODO |
| ECC-CALL-C522 | Corner | Egress bitrate drop (simulcast layer loss) during ICE brownout + download later | recording playable or marked degraded; no Hub panic; call stayed active | Manual | P2 | simulcast×ICE×egress | TODO |
| ECC-CALL-C523 | Corner | Audit/export of call participants while GDPR purge + multi-device leave race | export omit purged PII; counts match final; no email in audit payload | API | P0 | audit×GDPR×multi-device | Future |

**Calls Corner в этом файле: 24** (`ECC-CALL-C500`…`C523`)

---

## 2. Calendar — all-day / EXDATE × LK worker × WS reorder × organizer transfer — `ECC-CAL-C500+`

> Wave3: all-day и EXDATE vs LiveKit scheduled worker; WS `calendar.updated` out-of-order; organizer transfer; reminder×mute; personal vs team calendar conflict.  
> Не дублировать C300–C330 / C400–C424 (DST×instance Accept, dual×conflict, GDPR×iCal как **единственная** ось).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-C500 | Corner | All-day series Accept + timed overlapping meeting same UTC day + conflict API | all-day busy policy documented; timed conflict warn; Accept both allowed or soft-block per product | API | P0 | all-day×timed×conflict | Future |
| ECC-CAL-C501 | Corner | All-day event spanning TZ date-line for attendee; Accept; FE day grid | single local day block; UTC store correct; NFR-12; no double-day ghost | E2E | P0 | all-day×TZ×Accept | Future |
| ECC-CAL-C502 | Corner | EXDATE tomorrow + Accept series today + organizer adds RDATE same EXDATE day | EXDATE wins or RDATE re-adds per RFC policy; Hub consistent; no phantom Accept on removed | API | P0 | EXDATE×RDATE×Accept | Future |
| ECC-CAL-C503 | Corner | EXDATE applied **while** attendee Accept in-flight on that RECURRENCE-ID | Accept 410/gone; inbox clear for instance; series RSVP unchanged | API | P0 | EXDATE×Accept race | Future |
| ECC-CAL-C504 | Corner | COUNT=1 series edge + expand + Accept + LiveKit worker for that single occurrence | one room at due; expand no extras; after end no further ticks | API | P0 | COUNT=1×LK worker | Future |
| ECC-CAL-C505 | Corner | UNTIL before DTSTART (bad RRULE) repaired by organizer mid Accept storm | reject or clamp; Accept not on invalid; no 500; FE error | API | P1 | bad RRULE×Accept | Future |
| ECC-CAL-C506 | Corner | WS `calendar.updated` v3 arrives before v2 on attendee FE during reschedule | FE applies in version order; grid shows final; no flicker to old time | WS/E2E | P0 | WS reorder×reschedule | Future |
| ECC-CAL-C507 | Corner | WS disconnect 2m; Accept on other device; reconnect backlog | inbox/grid converge; no duplicate event keys; Join CTA once | WS/E2E | P0 | WS backlog×Accept | Future |
| ECC-CAL-C508 | Corner | Organizer transfer to attendee B while B Tentative + LiveKit room not yet started | B becomes organizer; Tentative→accepted or needs_action policy; old organizer demoted | API | P0 | transfer×Tentative | Future |
| ECC-CAL-C509 | Corner | Organizer transfer **during** LiveKit active call + recording | room ACL updates; End Call rights move; calendar organizer field matches | API/Manual | P0 | transfer×LK live | Future |
| ECC-CAL-C510 | Corner | Organizer transfer + L0 ICS UPDATE fanout + B offline | Hub SoT transfer; mail queued; B sees new role on next sync; no dual organizers | API/Manual | P0 | transfer×ICS | Future |
| ECC-CAL-C511 | Corner | Personal calendar Accept + team channel event same slot + conflict API merge | both sources listed; soft warn; busy union; no silent drop | API | P0 | personal×team×conflict | Future |
| ECC-CAL-C512 | Corner | Reminder 5m NATS + user muted calendar push + LiveKit room already open | no push; in-app/WS optional; no second room create | API | P1 | reminder×mute×LK | Future |
| ECC-CAL-C513 | Corner | Reminder fires after EXDATE removed occurrence | no reminder; no Join; metrics no false positive | API | P0 | reminder×EXDATE | Future |
| ECC-CAL-C514 | Corner | Search index lag: Ctrl+K finds old title after reschedule + Accept | search eventually consistent; Accept uses event id not title; no wrong Join | API/E2E | P1 | search lag×Accept | Future |
| ECC-CAL-C515 | Corner | Inbox cursor pagination mid Accept storm (50 invites) | no skipped/dup rows; badge count matches; Accept idempotent | API/E2E | P1 | inbox page×Accept storm | Future |
| ECC-CAL-C516 | Corner | HR leave cancelled mid-flight while Tentative on overlapping call | conflict recompute; Tentative soft-busy returns; leave gone | API | P1 | leave cancel×Tentative | Future |
| ECC-CAL-C517 | Corner | Sprint ceremony linked CalendarEvent rescheduled + Accept reset + LK worker old due | old tick no-op; new due one room; sprint link intact | API | P1 | sprint link×reschedule×LK | Future |
| ECC-CAL-C518 | Corner | Private/busy-only visibility event + ExportICal for non-attendee | free/busy block without title/attendees; attendee export full; no PII leak | API | P0 | busy-vis×iCal | Future |
| ECC-CAL-C519 | Corner | Floating all-day (no TZ) imported then Accept + Hub UTC normalize | stored UTC midnight policy; Accept binds; ICS round-trip documented | API | P1 | floating×Accept | Future |
| ECC-CAL-C520 | Corner | Two organizers race transfer A→B and A→C same version | one winner 409 other; single organizer invariant | API | P0 | transfer race | Future |
| ECC-CAL-C521 | Corner | Soft-delete series + restore within 30d + EXDATE set + Accept instance | restore policy clear; EXDATE preserved; Accept only valid instances | API | P0 | trash restore×EXDATE | Future |
| ECC-CAL-C522 | Corner | GDPR delete organizer after transfer completed + export range | old organizer PII gone; new organizer present; UID stable | API | P0 | GDPR×post-transfer export | Future |
| ECC-CAL-C523 | Corner | Aggregator as-is (scheduled call) + future CalendarEvent same slot + Accept Event only | dual until migration; Accept Event does not auto-hide ScheduledCall without link | API | P1 | aggregator×Event×Accept | Future |

**Calendar Corner в этом файле: 24** (`ECC-CAL-C500`…`C523`)

---

## 3. hub-mail — MIME / OOO / forwarded ICS × mute × rate fanout — `ECC-MAIL-C500+`

> Wave3: corrupt MIME, OOO auto-reply as bounce-class, forwarded invite, notif mute, channel-wide rate limits, WebPush NF-07 during fanout, METHOD=REFRESH while room live.  
> Не копировать C300–C324 / C400–C422 (reschedule×bounce / GDPR×export как единственная ось).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MAIL-C500 | Corner | Multipart MIME boundary corrupt on L0 REQUEST; in-app invite OK; Accept in-app | Hub SoT Accept; mail retry or quarantine; no crash parse; no partial event from bad MIME | API/Manual | P0 | MIME corrupt×Accept | Future |
| ECC-MAIL-C501 | Corner | HTML body XSS payload in DESCRIPTION + ICS ATTACH + Accept | sanitized FE; ICS text safe; no script exec; audit | API/E2E | P0 | XSS DESCRIPTION×ICS | Future |
| ECC-MAIL-C502 | Corner | OOO auto-reply 5xx-looking + real Accept in-app same attendee | OOO ≠ Decline; Accept stands; bounce classifier separates auto-reply | API/Manual | P0 | OOO×Accept | Future |
| ECC-MAIL-C503 | Corner | OOO during reschedule UPDATE + Tentative then Accept | Hub times new; Tentative→Accept OK; OOO logged not undeliverable-hard | Manual | P1 | OOO×reschedule×RSVP | Future |
| ECC-MAIL-C504 | Corner | Forwarded ICS REQUEST opened by non-attendee Hub user → Accept attempt | 403/reject; no attendee row created; organizer notif optional | API | P0 | forwarded ICS×Accept | Future |
| ECC-MAIL-C505 | Corner | Forwarded CANCEL for still-active Hub event | ignore/quarantine; Hub SoT not cancelled; SEC | API | P0 | forwarded CANCEL | Future |
| ECC-MAIL-C506 | Corner | METHOD=REFRESH while LiveKit room already live + recording | Hub ignores or no-op refresh; room not restarted; RSVP unchanged | API/Manual | P1 | REFRESH×LK live | Future |
| ECC-MAIL-C507 | Corner | METHOD=ADD attendee via external client while Hub ACL denies | Hub reject; outbound error mail optional; no orphan ATTENDEE | API | P0 | METHOD=ADD×ACL | Future |
| ECC-MAIL-C508 | Corner | Channel-wide invite 40 attendees + mail rate 60/min + Accept storm first 10 | queue backoff; in-app invites immediate; Accept OK without waiting SMTP | API | P0 | rate fanout×Accept | Future |
| ECC-MAIL-C509 | Corner | Messages rate limit hit mid ICS fanout + NATS invite lag | Hub events created; mail delayed; Accept not blocked; metrics | API | P0 | 60/min×NATS lag | Future |
| ECC-MAIL-C510 | Corner | Calendar push muted + email L0 fail + Telegram unbound + LiveKit start | in-app/WS still notify policy; call startable; no silent drop of SoT | API/Manual | P0 | mute×mail fail×LK | Future |
| ECC-MAIL-C511 | Corner | WebPush endpoint SSRF during invite fanout (NF-07) + L0 ICS still sending | ValidateOutboundURL reject bad push; ICS continues; invite SoT OK | API | P0 | NF-07×fanout×ICS | Future |
| ECC-MAIL-C512 | Corner | List-Unsubscribe clicked + user later Accepts in-app | unsubscribe ≠ Decline; Accept OK; marketing vs transactional paths isolated | Manual | P1 | List-Unsub×Accept | Future |
| ECC-MAIL-C513 | Corner | Attachment ICS near Fiber BodyLimit + HTML + inline image | reject oversize or strip; event still from Hub API create; no OOM | API | P1 | BodyLimit×ICS | Future |
| ECC-MAIL-C514 | Corner | Poison queue message (bad JSON) in hub-mail + healthy REQUEST behind | poison isolated; healthy delivered; Accept path unaffected | API | P0 | poison queue×REQUEST | Future |
| ECC-MAIL-C515 | Corner | SPF/DKIM fail quarantine on inbound L2 reply-ICS + Hub Accept already done | Hub SoT; inbound quarantined; no RSVP revert | API | P0 | SPF fail×Accept | Future |
| ECC-MAIL-C516 | Corner | Conversation/thread-id break on reschedule chain (provider new thread) + Accept | SEQUENCE/UID bind; Accept on Hub id; no duplicate events from new thread | Manual | P1 | thread-id×UID | Future |
| ECC-MAIL-C517 | Corner | Transactional invite routed to marketing template by misconfig + Accept | content may wrong brand but RSVP/ICS UID correct; ops alert; no PII in marketing list | Manual | P1 | tx vs mkt mix | Future |
| ECC-MAIL-C518 | Corner | VALARM in outbound ICS + Hub reminder NATS + mute push | dual reminders possible; mute suppresses push only; no double LiveKit start | API | P2 | VALARM×NATS reminder | Future |
| ECC-MAIL-C519 | Corner | External COUNTER propose new time + Hub organizer ignores + attendee Accepts Hub time | Hub time wins; COUNTER quarantined; no silent reschedule | API/Manual | P1 | COUNTER×Accept | Future |
| ECC-MAIL-C520 | Corner | L1 feed poll during MIME-retry storm + organizer transfer | feed shows final organizer; no torn read; token auth OK | API | P1 | L1×MIME retry×transfer | Future |
| ECC-MAIL-C521 | Corner | GDPR delete mid MIME-retry holding raw email MIME on disk/queue | redact/delete blob; retry aborted; logs scrubbed; audit | API | P0 | GDPR×MIME retry | Future |
| ECC-MAIL-C522 | Corner | Soft bounce loop (4xx×N) on REQUEST while user Declines in-app | Decline terminal; mail retries stop; no Accept from late delivery | API/Manual | P0 | soft bounce×Decline | Future |
| ECC-MAIL-C523 | Corner | ICS UID collision across two Hub orgs (same generator bug) on one mailbox | Hub scopes by org/instance; no cross-org apply; quarantine | API | P0 | UID collision×org | Future |

**Mail Corner в этом файле: 24** (`ECC-MAIL-C500`…`C523`)

---

## 4. Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Calls + LiveKit (multi-device×ICE×Seaweed×NATS) | `ECC-CALL-C500`…`C523` | **24** |
| Calendar (all-day/EXDATE×LK worker×WS×transfer) | `ECC-CAL-C500`…`C523` | **24** |
| hub-mail (MIME/OOO/forward×mute×rate fanout) | `ECC-MAIL-C500`…`C523` | **24** |
| **Итого Corner** | | **72** |

| Проверка | Статус |
|----------|--------|
| Минимум 60 Corner | ✅ 72 ≥ 60 |
| Только тип Corner | ✅ |
| ID `C500+` only | ✅ |
| mediasoup / SEED-001 | ✅ N/A (не включено) |
| Parent C200 / wave1 C300 / wave2 C400 дословно | ✅ не копированы |
| Каталог `EDGE-CORNER-CATALOG.md` | ✅ не тронут |
| Git commit | ✅ не создавался |

**Фокус-покрытие wave3:**

| Тема | Примеры ID |
|------|------------|
| Multi-device × ICE × End/Record | `ECC-CALL-C500`…`C502`, `C511`, `C517`–`C518` |
| Seaweed egress outage / orphan | `ECC-CALL-C502`…`C504`, `C516` |
| NATS invite lag × room start | `ECC-CALL-C505`–`C506`; `ECC-MAIL-C509` |
| Organizer transfer × LK/ACL | `ECC-CALL-C507`–`C508`; `ECC-CAL-C508`–`C510`, `C520` |
| All-day / EXDATE × LK worker | `ECC-CALL-C519`–`C520`; `ECC-CAL-C500`–`C505`, `C513` |
| WS reorder / backlog × RSVP | `ECC-CAL-C506`–`C507` |
| MIME / OOO / forwarded ICS | `ECC-MAIL-C500`–`C505`, `C515` |
| Mute × rate fanout × NF-07 | `ECC-MAIL-C508`–`C511`; `ECC-CAL-C512` |
| METHOD=REFRESH / COUNTER × LK | `ECC-MAIL-C506`, `C519` |

**Handoff:** `qa-manual-ver` → после `verified: true` выборочно `qa-auto1` (NATS lag, EXDATE race, transfer, NF-07) / Manual (ICE partition, MIME clients, multi-device).  
**Merge later:** в `.planning/EDGE-CORNER-CATALOG.md` § Corner expansion (отдельной волной; каталог сейчас не трогать).
