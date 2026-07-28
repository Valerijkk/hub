# Corner Expansion C04 — Calls + Calendar + Mail

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual`  
> **Не трогает** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` и parent `00-CORNER-PARENT.md`  
> **ID range:** `ECC-*-C300+` (не пересекает parent `C200–C215`, catalog `C01–C25`, edge `*100+`)  
> **SFU:** LiveKit stay — mediasoup / SEED-001 = **N/A** (строки mediasoup не добавлять)  
> **Источники:** JULY brief §3 (C1–C15, L0–L3) · SEED-002 · parent CALL/CAL C200–C215 (не дублировать дословно) · US-M06 · US-CAL-* · NFR-12 · SEC-02 · NF-07

**Назначение:** foresight-бэклог **новых** corner-комбинаций для будущих `TC-*`. Happy / Neg / Edge сюда не входят.

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (ещё не в каталоге) / **Future** (CAL-SYNC / mail L*) |
| Слой | API / Unit / Manual / E2E / WS |
| Связь | Brief C# · SEED-002 · TC · NFR · parent theme (расширение, не копия) |

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — `ECC-CALL-C300+`

> Фокус: SFU restart × recording × webhook × key rotate × scheduled/manual × guest-after-end.  
> Parent C200–C209 = базовые темы; здесь — **мультифакторные** комбинации.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-C300 | Corner | LiveKit SFU restart mid-call **+** active egress recording | reconnect media; egress resumes or clean fail+retry; call status ≠ ended; no orphan Seaweed object | Manual | P0 | parent C200×C202, LiveKit | TODO |
| ECC-CALL-C301 | Corner | SFU restart during **guest** join handshake (token issued, room not yet joined) | guest completes join or gets fresh token; no ghost participant count | Manual | P0 | C200×C203, TC-CALL-04 | TODO |
| ECC-CALL-C302 | Corner | SFU restart + member JWT near 24h expiry | reconnect with rejoin/refresh; no false «call ended» toast | Manual | P1 | C200×C204, token 24h | TODO |
| ECC-CALL-C303 | Corner | Webhook `participant_joined` replay **after** key rotate (old signature) | 401; participant count unchanged; SEC-02 never skipped | API | P0 | C201×C209, SEC-02, FA-06 | TODO |
| ECC-CALL-C304 | Corner | Webhook replay **same event_id** twice with **new** valid key | idempotent; count +1 once; audit one logical join | API | P0 | C201, SEC-02 | TODO |
| ECC-CALL-C305 | Corner | Key rotate mid-flight: egress `recording_finished` signed old key arrives late | reject; recording state not flipped to ready; ops alert | API | P0 | C209×C202, FA-06 | TODO |
| ECC-CALL-C306 | Corner | Rotate webhook keys while SFU restart storm (burst joins/leaves) | only new-key events applied; no double-count; Hub stays up | Manual | P0 | C200×C201×C209 | TODO |
| ECC-CALL-C307 | Corner | End call + stop recording + SFU restart overlapping 3-way race | terminal: call ended; egress stopped or marked failed; no zombie room | API/Manual | P0 | C202×C200 | TODO |
| ECC-CALL-C308 | Corner | Start recording A, start recording B (max=2), end call before both finish | both egress finalize/cancel; ≤2 active; download only completed | API | P0 | C202, maxActiveRecordings=2 | TODO |
| ECC-CALL-C309 | Corner | Recording start + participant leave last member + auto-end race | no orphan egress; recording linked to ended call or purged | API | P0 | C202, auto-end | TODO |
| ECC-CALL-C310 | Corner | Recording webhook `egress_failed` then delayed `egress_complete` for same id | final state failed **or** complete per last-valid; never both ready+failed | API | P0 | recording state machine | TODO |
| ECC-CALL-C311 | Corner | Scheduled due-worker fires **while** manual start **and** conflict check in parallel | exactly one LiveKit room; second path 409/idempotent; one ScheduledCall→active | API | P0 | C205×C206 | TODO |
| ECC-CALL-C312 | Corner | Manual start scheduled call 1s before worker tick; worker uses stale «due» snapshot | worker no-op / idempotent; no second room | API | P0 | C205, scheduled worker | TODO |
| ECC-CALL-C313 | Corner | Two admins: A manual-starts, B cancels scheduled same ms | one terminal (active **or** cancelled); no room without DB row | API | P0 | C205×C208 | TODO |
| ECC-CALL-C314 | Corner | Guest invite link used **after** end, while recording still processing | join 410/404; recording metadata not leaked to guest | API | P0 | C203×C202, TC-CALL-04 | TODO |
| ECC-CALL-C315 | Corner | Guest joins, call ends, guest retries same invite code + old LiveKit JWT | both rejected; no re-open room | API | P0 | C203×C204 | TODO |
| ECC-CALL-C316 | Corner | Guest in room during SFU restart; host ends call mid-reconnect | guest lands on ended UI; no media to empty room | Manual | P1 | C200×C203 | TODO |
| ECC-CALL-C317 | Corner | Channel deleted **during** scheduled+manual start race | no new room; scheduled deactivated; 404 on join | API | P0 | C205×C208 | TODO |
| ECC-CALL-C318 | Corner | Screen-share + recording + SFU restart (tracks republish) | after reconnect share optional re-prompt; recording continuity policy documented | Manual | P1 | C200, US-M06, TC-CALL-07 | TODO |
| ECC-CALL-C319 | Corner | Raise-hand metadata update during SFU reconnect window | no crash; stale hand cleared or reconciled; UI safe | FE/Manual | P2 | C207×C200 | TODO |
| ECC-CALL-C320 | Corner | Webhook `room_finished` replay after user already re-created call same channel | old webhook must not end **new** call (room/call id bind) | API | P0 | C201, room id binding | TODO |
| ECC-CALL-C321 | Corner | Conflict check pass → concurrent book → then scheduled worker starts loser | loser not started; winner room only; conflict warn persisted | API | P1 | C206×C205 | TODO |
| ECC-CALL-C322 | Corner | Member removed from channel mid-call + SFU restart | removed user cannot rejoin after reconnect; media kicked | API/Manual | P0 | NF-02, C200 | TODO |
| ECC-CALL-C323 | Corner | Download recording while soft-delete call + GDPR purge job queued | download fails closed after purge; no cross-user leak | API | P0 | GDPR, recording TTL | TODO |
| ECC-CALL-C324 | Corner | Invite code rotate (new code) while guest holds old code + call still active | old code invalid; new works; no dual valid codes | API | P1 | TC-CALL-03/04 | TODO |
| ECC-CALL-C325 | Corner | LiveKit webhook clock skew: event timestamp in future vs Hub clock | reject or clamp; no state corruption; SEC-02 still enforced | API | P1 | SEC-02, clock skew | TODO |
| ECC-CALL-C326 | Corner | Max participants-1 + guest join + SFU restart + burst rejoin | count never > max; excess rejected; no negative count | Manual | P0 | NFR-03, C200 | TODO |
| ECC-CALL-C327 | Corner | Scheduled RRULE occurrence starts while previous occurrence recording still uploading | new room OK; recordings isolated by call/occurrence id | API | P1 | RRULE, C202 | TODO |
| ECC-CALL-C328 | Corner | Host ends call from FE; API end; LiveKit auto-empty race (triple) | single ended transition; webhooks idempotent | API | P0 | C202, end paths | TODO |

**Calls Corner в этом файле: 29** (`ECC-CALL-C300`…`C328`)

---

## 2. Calendar as-is + CAL-SYNC RSVP — `ECC-CAL-C300+`

> As-is = read-агрегатор (calls/sprints/deadlines/leave). Future = SEED-002 RSVP / Brief C1–C15.  
> Parent C200–C215 и catalog C01–C25 — не копировать; только **новые** связки.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-C300 | Corner | As-is aggregator: soft-deleted scheduled call **+** overlapping sprint deadline same day | call hidden; sprint/deadline still listed; no ghost call slot | API | P1 | parent C200×aggregator | TODO |
| ECC-CAL-C301 | Corner | As-is: call soft-deleted then restored within 30d while list range open in UI | appears after invalidate; no duplicate keys | API/E2E | P1 | soft-delete 30d | TODO |
| ECC-CAL-C302 | Corner | Expand RRULE infinite across DST spring-forward inside 366d range | instances wall-correct; no missing/dup day; no OOM | Unit | P0 | C201×C213, NFR-12 | TODO |
| ECC-CAL-C303 | Corner | Accept (A) + Decline (B) + Organizer delete — three-way same attendee version | one terminal: gone **or** declined; inbox clear; no 500 | API | P0 | C202×C203, Brief C1/C10 | Future |
| ECC-CAL-C304 | Corner | Attendee Accept while organizer Reschedule mid-flight (version bump) | 409 or apply-to-new-time per C8 policy; no accepted-on-old-slot ghost | API | P0 | C202×C204, Brief C8/C10 | Future |
| ECC-CAL-C305 | Corner | Reschedule → RSVP reset needs_action **while** second device still shows Accept UI and posts Accept | Accept targets new version or 409; no accepted without seeing new time | API/E2E | P0 | Brief C8×C10 | Future |
| ECC-CAL-C306 | Corner | Organizer cancel + L0 ICS CANCEL in-flight + attendee Accept | Accept 410; ICS CANCEL wins mailbox; in-app inbox cleared | API/Manual | P0 | C203×C204, Brief C1 | Future |
| ECC-CAL-C307 | Corner | Title-only patch (C7) **during** Accept race | rsvp stays accepted; title updated; no accidental needs_action | API | P1 | Brief C7×C10, C202 | Future |
| ECC-CAL-C308 | Corner | Recurring series Accept + DST fall-back overlap week | series accepted; local times correct; single busy block per occurrence | E2E | P0 | C205×C213, Brief C2/C6 | Future |
| ECC-CAL-C309 | Corner | Series Accept then future exception reschedule (backlog path if enabled) | v1: reject exception **or** documented backlog; series integrity | API | P1 | Brief C2, C205 | Future |
| ECC-CAL-C310 | Corner | Channel member **added** after invite snapshot created | policy: not auto-attendee **or** late-invite; no silent ACL grant to call | API | P1 | C206, Brief C11/C25 | Future |
| ECC-CAL-C311 | Corner | Channel member **removed** after Accept; call still upcoming | revoke calendar write / kick from attendees per policy; organizer notified | API | P0 | C206, Brief C25 | Future |
| ECC-CAL-C312 | Corner | Member removed while Accept in-flight | Accept fail-closed 403/410; no accepted orphan for non-member | API | P0 | C206×C202 | Future |
| ECC-CAL-C313 | Corner | Dual create: messenger scheduled call + calendar event same slot **+** both send invites | conflict warn both sides; dedupe by source_ref if linked; no double busy | API | P0 | C211, SEED-002 | Future |
| ECC-CAL-C314 | Corner | Dual create race without link: two CalendarEvents, user Accepts both overlapping | soft conflict warning; both can stay accepted only with explicit override | API/E2E | P1 | Brief C3×C211 | Future |
| ECC-CAL-C315 | Corner | Messenger thread delete **while** calendar Accept happens | thread gone ≠ event cancel; Accept OK; Brief C15 holds | API | P0 | Brief C15, C211 | Future |
| ECC-CAL-C316 | Corner | GDPR hard-delete attendee **after** Accept **+** series still recurring | attendee row anonymized/removed; future instances no PII; organizer sees «removed user» | API | P0 | C212, Brief C9, GDPR | Future |
| ECC-CAL-C317 | Corner | GDPR delete attendee mid Accept/Decline race | request fails closed; purge completes; no PII in audit payload | API | P0 | C212×C202 | Future |
| ECC-CAL-C318 | Corner | Tentative + HR leave approve same slot **+** reschedule leave | availability: leave wins; tentative soft; after leave move — tentative re-eval | API | P1 | C207, Brief C12, US-CAL-04 | Future |
| ECC-CAL-C319 | Corner | External iCal busy (L1) + Hub RSVP Accept same slot + leave | Hub SoT for Hub event; external read-only; availability merge policy documented | API | P1 | C208, Brief L1 | Future |
| ECC-CAL-C320 | Corner | needs_action ignored 30d **during** organizer reschedule storm | expire/reminder policy; after reschedule clock resets or not — product lock | API | P2 | C214×C204 | Future |
| ECC-CAL-C321 | Corner | Partial hub-mail outage: invite created in-app, ICS fail, then Accept before mail retry | Accept OK (in-app SoT); mail retry sends UPDATE/REQUEST not duplicate create | API/Manual | P0 | C215, Brief C13 | Future |
| ECC-CAL-C322 | Corner | Organizer self-decline attempt **during** cancel API (C5×cancel) | decline rejected; cancel succeeds; attendees get CANCEL not «organizer declined» | API | P0 | Brief C5 | Future |
| ECC-CAL-C323 | Corner | Add attendee not in channel (C4) **while** channel ACL changes to allow | second retry may succeed; first 403; no partial attendee without notif | API | P1 | Brief C4 | Future |
| ECC-CAL-C324 | Corner | Two invites same time (C3) + user Accepts first + Tentative second | conflict badge; busy/free reflects accepted hard + tentative soft | API/E2E | P1 | Brief C3×C19 | Future |
| ECC-CAL-C325 | Corner | Soft-delete event (trash) + attendee Accept from stale inbox deep-link | 410/gone; trash not resurrected by Accept | API | P0 | Brief C1×C9 | Future |
| ECC-CAL-C326 | Corner | Restore from trash + attendee had declined before delete | RSVP policy on restore documented; no auto-accepted | API | P1 | Brief C9 | Future |
| ECC-CAL-C327 | Corner | As-is list + future RSVP projection: needs_action must not appear as hard-busy in aggregator | inbox only; grid free/soft per §3.2 | API | P1 | Brief §3.2, C20 | Future |
| ECC-CAL-C328 | Corner | Worker creates LiveKit room from accepted CalendarEvent **while** organizer cancels | no room **or** immediate end; attendees not stuck in lobby | API | P0 | SEED-002, CALL C205 | Future |
| ECC-CAL-C329 | Corner | Accept idempotent double-submit **across** WS reconnect (two tabs) | single accepted; one NATS fanout; version +1 once | API/WS | P1 | C23, C10 | Future |
| ECC-CAL-C330 | Corner | DST spring gap: create event at non-existent local time via FE, store UTC, RSVP Accept | stored valid UTC; FE shows resolved local; ICS VTIMEZONE consistent | E2E | P0 | NFR-12, Brief C6 | Future |

**Calendar Corner в этом файле: 31** (`ECC-CAL-C300`…`C330`)

---

## 3. hub-mail L0–L3 — `ECC-MAIL-C300+`

> L0 ICS · L1 iCal feed · L2 IMAP/JMAP · L3 OAuth (ADR required).  
> Связь с bounce/SoT (Brief C13), UPDATE vs CANCEL, dual-create, GDPR.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MAIL-C300 | Corner | L0: send REQUEST, then UPDATE (reschedule), client still processing first mail | SEQUENCE/UID bump; client shows new time; no duplicate events | Manual | P0 | Brief L0, C8 | Future |
| ECC-MAIL-C301 | Corner | L0: UPDATE vs CANCEL race (cancel queued before update delivered) | mailbox ends cancelled; Hub event cancelled; no zombie accepted | Manual/API | P0 | L0 METHOD, Brief C1 | Future |
| ECC-MAIL-C302 | Corner | L0: CANCEL after attendee already Accept in-app (email lag) | in-app cancelled; late CANCEL ICS harmless; Accept not re-open | API/Manual | P0 | Brief C13×C1 | Future |
| ECC-MAIL-C303 | Corner | L0 ICS REQUEST bounce (5xx) + in-app invite exists; user Accepts | Accept OK; bounce logged; retry/backoff; SoT = Hub | API/Manual | P0 | Brief C13, C215 | Future |
| ECC-MAIL-C304 | Corner | L0 bounce on UPDATE only (REQUEST delivered earlier) | Hub keeps new time; attendee sees in-app banner; email retry UPDATE | Manual | P1 | Brief C13×C8 | Future |
| ECC-MAIL-C305 | Corner | L0: organizer GDPR-deletes attendee after ICS REQUEST sent | CANCEL or anonymized UPDATE; no PII in subsequent mail; Hub row purged | API/Manual | P0 | GDPR, C212 | Future |
| ECC-MAIL-C306 | Corner | L0 attachment + hub-mail provider timeout after SMTP accept unknown | no duplicate SEQUENCE on retry; idempotent mail job key | API | P0 | hub-mail jobs | Future |
| ECC-MAIL-C307 | Corner | L0 METHOD=REQUEST with stale SEQUENCE after concurrent title+time edits | final SEQUENCE matches Hub version; clients converge | Manual | P1 | Brief C7×C8 | Future |
| ECC-MAIL-C308 | Corner | Channel member change after invite: L0 mail list vs Hub attendees diverge | Hub attendees SoT; mail only to current set; removed get CANCEL | API/Manual | P1 | C206, Brief C25 | Future |
| ECC-MAIL-C309 | Corner | Dual create messenger+calendar → two L0 REQUEST for same slot | two UIDs **or** linked single UID; conflict warn; no silent merge wrong | Manual | P0 | C211, SEED-002 | Future |
| ECC-MAIL-C310 | Corner | L1 iCal feed poll during Hub cancel of exported event | next poll removes/cancels VEVENT; no stale busy forever | API | P0 | L1, GZ-V4-05 | Future |
| ECC-MAIL-C311 | Corner | L1 subscribe URL fetch SSRF (internal IP) | ValidateOutboundURL reject; NF-07 | API | P0 | NF-07, US-CAL-03 | Future |
| ECC-MAIL-C312 | Corner | L1 feed + Hub RSVP Accept: external consumer sees only accepted/organizer projection | needs_action not exported as busy; privacy OK | API | P1 | L1×§3.2 | Future |
| ECC-MAIL-C313 | Corner | L1 LimitReader cap hit mid-VEVENT during import | reject/error state; no partial corrupt event | API | P0 | io.LimitReader, US-CAL-03 | Future |
| ECC-MAIL-C314 | Corner | L2 IMAP: duplicate UID + Hub already has event from L0 | idempotent; no second CalendarEvent | API | P0 | C209, L2 ADR | Future |
| ECC-MAIL-C315 | Corner | L2 IMAP ingest CANCEL for event user already declined in Hub | stay declined/cancelled; no resurrect | API | P1 | L2, Brief C1 | Future |
| ECC-MAIL-C316 | Corner | L2 parse malformed ICS + parallel L0 send for same meeting | Hub L0 wins SoT; malformed quarantined; no crash | API | P1 | L2, C13 | Future |
| ECC-MAIL-C317 | Corner | L3 OAuth token revoke mid two-way sync **during** Hub reschedule | stop sync; Hub local state consistent; no half-pushed Google event | API | P0 | C210, L3 ADR | Future |
| ECC-MAIL-C318 | Corner | L3 without ADR override (GZ-V4-05) enabled in config | feature blocked; no token store; audit | API | P0 | C22, GZ-V4-05 | Future |
| ECC-MAIL-C319 | Corner | L3 sync + GDPR attendee delete | remote attendee removed/anonymized; Hub purge; no re-sync PII | API | P0 | L3×GDPR | Future |
| ECC-MAIL-C320 | Corner | L0 CANCEL + L1 feed + in-app Accept race (triple channel) | all channels show cancelled/gone; Accept 410 | Manual | P0 | L0×L1×C1 | Future |
| ECC-MAIL-C321 | Corner | Email bounce + Telegram/in-app notif success (multi-channel) | in-app+WS SoT; bounce doesn't hide invite badge | API | P1 | Brief C13, notif | Future |
| ECC-MAIL-C322 | Corner | L0 ICS VTIMEZONE across DST + recurring series UPDATE | clients show correct wall; SEQUENCE++ per instance policy | Manual | P0 | NFR-12, L0 | Future |
| ECC-MAIL-C323 | Corner | hub-mail rate-limit / provider 429 during invite storm (channel=all members) | queue backoff; all members eventually invited in-app; no drop without retry | API | P1 | Brief C11, L0 | Future |
| ECC-MAIL-C324 | Corner | External organizer email-only (C14) attempts L0 path | rejected v1; no outbound ICS to unmanaged identity | API | P2 | Brief C14 | Future |

**Mail Corner в этом файле: 25** (`ECC-MAIL-C300`…`C324`)

---

## 4. Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Calls + LiveKit | `ECC-CALL-C300`…`C328` | **29** |
| Calendar as-is + CAL-SYNC | `ECC-CAL-C300`…`C330` | **31** |
| hub-mail L0–L3 | `ECC-MAIL-C300`…`C324` | **25** |
| **Итого Corner** | | **85** |

| Проверка | Статус |
|----------|--------|
| Минимум 65 Corner | ✅ 85 ≥ 65 |
| Только тип Corner | ✅ |
| mediasoup / SEED-001 | ✅ N/A (не включено) |
| Parent C200–C215 дословно | ✅ не копированы; только расширения × |
| Каталог `EDGE-CORNER-CATALOG.md` | ✅ не тронут |
| Git commit | ✅ не создавался |

**Handoff:** `qa-manual-ver` → затем (после verified) выборочная автоматизация `qa-auto1` (API races) / Manual для SFU+ICS clients.  
**Merge later:** в `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` § Corner expansion (отдельной волной).
