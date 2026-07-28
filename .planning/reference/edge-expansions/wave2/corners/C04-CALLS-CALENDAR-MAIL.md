# Corner Expansion C04 wave2 — Calls + Calendar + Mail

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave2  
> **Не трогает** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`, parent `00-CORNER-PARENT.md`, wave1 `corners/C04-*`  
> **ID range:** `ECC-*-C400+` (не пересекает parent `C200–C215`, wave1 `C300+`, catalog `C01–C25`, edge `*100+`)  
> **SFU:** LiveKit stay — mediasoup / SEED-001 = **N/A** (строки mediasoup не добавлять)  
> **Фокус wave2:** RSVP×recording×LiveKit restart · ICS×reschedule×bounce · dual messenger+calendar create×conflict API · DST×series×accept instance · GDPR attendee×export iCal  
> **Источники:** JULY brief §3 · SEED-002 · US-M06 · US-CAL-* · NFR-12 · SEC-02 · NF-07 · GDPR soft/hard delete

**Назначение:** foresight-бэклог **новых** corner-комбинаций wave2 для будущих `TC-*`. Happy / Neg / Edge сюда не входят. **Не копировать** сценарии C200/C300 — только новые мультифакторные связки.

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (ещё не в каталоге) / **Future** (CAL-SYNC / mail L*) |
| Слой | API / Unit / Manual / E2E / WS |
| Связь | Brief C# · SEED-002 · TC · NFR · theme (расширение, не копия) |

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — RSVP × recording × SFU restart — `ECC-CALL-C400+`

> Тройные/четверные связки: calendar RSVP → LiveKit room → egress recording → SFU restart / webhook / end.  
> Wave1 C300–C328 = другие комбинации; здесь — **RSVP-driven** и **recording-during-restart** с calendar/mail.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-C400 | Corner | Attendee Accept → worker starts LiveKit room → egress recording ON → SFU restart mid-first-minute | media reconnect; egress resume/retry; RSVP stays accepted; call ≠ ended | Manual | P0 | RSVP×rec×LK, SEED-002 | Future |
| ECC-CALL-C401 | Corner | Tentative RSVP + host starts recording + SFU restart before tentative→accept | tentative user may rejoin; recording not tied to tentative-as-accepted; no PII leak in egress meta | Manual | P0 | RSVP tentative×rec×LK | Future |
| ECC-CALL-C402 | Corner | Decline after room already live + recording active + SFU restart | declined user cannot rejoin post-restart; egress continues for remaining; attendee list reflects decline | API/Manual | P0 | Decline×rec×LK | Future |
| ECC-CALL-C403 | Corner | Organizer reschedule (RSVP reset) **while** previous occurrence still recording + SFU restart on old room | old room ends/isolates; new occurrence no orphan egress; RSVP needs_action on new version | API/Manual | P0 | Reschedule×rec×LK | Future |
| ECC-CALL-C404 | Corner | Accept instance (exception) of series → recording on that instance only → SFU restart | series other instances untouched; recording keyed by occurrence/call id; reconnect stays on exception room | API/Manual | P0 | Series instance×rec×LK | Future |
| ECC-CALL-C405 | Corner | needs_action ignored + scheduled worker starts room + auto-recording policy + SFU restart | room for organizer/accepted only; needs_action not auto-joined; recording policy documented | API | P1 | needs_action×rec×LK | Future |
| ECC-CALL-C406 | Corner | Accept race (two tabs) + recording start + SFU restart overlapping | single accepted version; ≤1 egress start; reconnect count sane | API/Manual | P0 | Accept race×rec×LK | Future |
| ECC-CALL-C407 | Corner | Guest invite from calendar external + recording ON + SFU restart + guest JWT near 4h | guest reconnect or fresh token; recording not downloadable by guest; SEC-02 webhooks OK | Manual | P0 | Guest×rec×LK, TC-CALL-04 | Future |
| ECC-CALL-C408 | Corner | RSVP Accept → join → start recording A+B (max=2) → SFU restart → attempt recording C | after restart still ≤2 active; C rejected; no zombie egress | API | P0 | maxRec=2×LK restart | TODO |
| ECC-CALL-C409 | Corner | Recording `egress_complete` webhook arrives **during** SFU restart window (room briefly empty) | complete applied to correct call_id; not marked failed-empty; downloadable after | API | P0 | webhook×LK empty window | TODO |
| ECC-CALL-C410 | Corner | Recording `egress_failed` + delayed complete + SFU restart + host End Call triple | terminal consistent (failed **or** complete **or** cancelled); never ready+failed; room gone | API | P0 | egress SM×LK×end | TODO |
| ECC-CALL-C411 | Corner | Calendar cancel after Accept **while** LiveKit room + recording + SFU restart in progress | join blocked; egress stop/fail-closed; late webhooks cannot reopen call | API/Manual | P0 | Cancel×rec×LK | Future |
| ECC-CALL-C412 | Corner | Channel remove accepted attendee mid-call + recording + SFU restart | removed cannot rejoin; egress continues; participant count never includes removed after kick | API/Manual | P0 | ACL×rec×LK, NF-02 | Future |
| ECC-CALL-C413 | Corner | Soft-delete CalendarEvent (trash) while LiveKit active + recording + SFU restart | fail-closed: end room or keep until end then purge link; trash Accept deep-link 410 | API | P0 | Trash×rec×LK | Future |
| ECC-CALL-C414 | Corner | GDPR hard-delete attendee mid-call (accepted) + recording active + SFU restart | media kicked; PII stripped from participant meta; recording retained without attendee PII or redacted | API/Manual | P0 | GDPR×rec×LK | Future |
| ECC-CALL-C415 | Corner | Screen-share + RSVP late-joiner Accept + recording + SFU restart (tracks republish) | late joiner sees share after reconnect policy; recording continuity documented; no crash | Manual | P1 | US-M06×RSVP×rec×LK | Future |
| ECC-CALL-C416 | Corner | Raise-hand + recording metadata + SFU restart + RSVP Decline from second device | hand cleared; decline kicks; recording not containing declinee identity beyond policy | FE/Manual | P2 | hand×Decline×rec×LK | Future |
| ECC-CALL-C417 | Corner | Scheduled RRULE occurrence N recording still uploading + occurrence N+1 Accept-driven start + SFU restart on N+1 | recordings isolated by occurrence; N+1 reconnect OK; no cross-link Seaweed keys | API | P0 | RRULE×rec×LK | Future |
| ECC-CALL-C418 | Corner | Conflict API warned slot ignored → dual rooms race (messenger start + calendar Accept start) + recording on both + SFU restart one SFU | both rooms tracked; conflict badge persists; restart only affects one room_id; no webhook cross-end | API/Manual | P0 | Conflict×dual×rec×LK | Future |
| ECC-CALL-C419 | Corner | LiveKit webhook `participant_joined` for declined user after SFU restart (stale token) | reject/ignore; count unchanged; SEC-02 signature valid still required | API | P0 | Decline×webhook×LK, SEC-02 | Future |
| ECC-CALL-C420 | Corner | Host ends call from FE during SFU restart while egress stopping and RSVP inbox still shows «Join» | Join CTA disabled/410; single ended; no re-create from stale inbox | E2E | P0 | End×inbox×LK | Future |
| ECC-CALL-C421 | Corner | Member JWT 24h boundary during SFU restart **and** recording finalize | rejoin refresh; finalize not lost; no false ended toast | Manual | P1 | token 24h×rec×LK | TODO |
| ECC-CALL-C422 | Corner | Download recording URL issued → GDPR purge attendee who spoke → SFU room already gone | download fail-closed after purge policy; no cross-user; audit | API | P0 | GDPR×download×LK gone | Future |

**Calls Corner в этом файле: 23** (`ECC-CALL-C400`…`C422`)

---

## 2. Calendar — DST × series × accept instance · dual create × conflict · GDPR × iCal — `ECC-CAL-C400+`

> Wave2: instance-level Accept на DST, dual messenger+calendar × FindConflicts, GDPR attendee × ExportICal.  
> Не дублировать C300–C330 дословно.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-C400 | Corner | Weekly series spans DST spring-forward; Accept **series**; then Accept **single instance** on gap week (exception path) | series accepted; instance wall-time correct; no phantom hour; busy one block | API/E2E | P0 | DST×series×instance, NFR-12 | Future |
| ECC-CAL-C401 | Corner | Series Accept then Decline **only** fall-back overlap instance (ambiguous local hour) | exception declined; other instances stay accepted; UTC store unambiguous | API | P0 | DST fall-back×instance | Future |
| ECC-CAL-C402 | Corner | Create series DTSTART in spring gap via API (UTC ok) + FE shows resolved + Accept instance | stored valid UTC; ICS VTIMEZONE consistent; Accept binds correct RECURRENCE-ID | E2E | P0 | DST gap×Accept instance | Future |
| ECC-CAL-C403 | Corner | Reschedule **one instance** across DST boundary while attendee Accept in-flight on old RECURRENCE-ID | 409 or apply-to-new; no accepted-on-deleted-instance | API | P0 | Reschedule instance×Accept×DST | Future |
| ECC-CAL-C404 | Corner | Series Accept + organizer moves series DTSTART across DST + attendee second device still Accept old | version bump; second Accept 409/new version; inbox shows new wall times | API/E2E | P0 | Series move×DST×Accept | Future |
| ECC-CAL-C405 | Corner | Expand series 366d across both DST transitions + Accept instance at range edge (last occurrence) | expand capped; Accept OK; no OOM; edge instance not dropped | Unit/API | P0 | expand×DST×Accept edge | Future |
| ECC-CAL-C406 | Corner | Dual create: messenger scheduled call + calendar event same slot → both hit conflict API same ms | both get HasConflict/warn; neither silently wins; source_ref link if policy | API | P0 | Dual×conflict API | Future |
| ECC-CAL-C407 | Corner | Dual create linked (source_ref) then conflict API on third overlapping booking | conflict lists both linked sources once (dedupe); third warned | API | P0 | Dual linked×conflict | Future |
| ECC-CAL-C408 | Corner | Dual create unlinked → user Accepts calendar + messenger auto-starts LiveKit → conflict API re-check | soft/hard conflict UI; no double busy without override; rooms policy documented | API/E2E | P0 | Dual×Accept×conflict×LK | Future |
| ECC-CAL-C409 | Corner | Conflict check pass → dual create race inserts second → Accept first → conflict API on Accept path | Accept may warn/409 per policy; no silent double-accept hard-busy | API | P0 | TOCTOU conflict×Accept | Future |
| ECC-CAL-C410 | Corner | Messenger thread delete after dual create + calendar Accept + conflict badge still shown | thread gone ≠ cancel event; Accept OK; conflict recompute without messenger source | API | P0 | Dual×thread delete×Accept | Future |
| ECC-CAL-C411 | Corner | Dual create + channel ACL revoke mid conflict check | fail-closed 403; no partial CalendarEvent without membership | API | P0 | Dual×ACL×conflict | Future |
| ECC-CAL-C412 | Corner | GDPR hard-delete attendee **after** series Accept + ExportICal for organizer range | VEVENT ATTENDEE without PII / omitted; CN anonymized; UID stable; no email in ICS | API | P0 | GDPR×series×iCal export | Future |
| ECC-CAL-C413 | Corner | GDPR delete attendee mid ExportICal streaming response | export completes fail-closed or omits PII; no partial email leak mid-stream | API | P0 | GDPR×export race | Future |
| ECC-CAL-C414 | Corner | GDPR delete + attendee had Accept instance exception + export includes RECURRENCE-ID overrides | exception rows purged/anonymized; series export clean | API | P0 | GDPR×instance×iCal | Future |
| ECC-CAL-C415 | Corner | Soft-delete event (trash) + ExportICal range still open | trashed VEVENT absent; restore then export reappears with RSVP policy | API | P0 | Trash×iCal export | Future |
| ECC-CAL-C416 | Corner | ExportICal while organizer reschedule storm (version++) + attendee Accept | export SEQUENCE/version matches final Hub; Accept not on stale export | API | P1 | Export×reschedule×Accept | Future |
| ECC-CAL-C417 | Corner | As-is aggregator + future RSVP Accept instance + DST week + soft-deleted call same day | call hidden; accepted instance shown correct local; no ghost | API | P1 | Aggregator×DST×Accept | Future |
| ECC-CAL-C418 | Corner | Tentative on series instance + HR leave same DST week + conflict API | leave priority; tentative soft; conflict payload distinguishes | API | P1 | Tentative×leave×DST×conflict | Future |
| ECC-CAL-C419 | Corner | External L1 busy + Hub series Accept instance + conflict API merge | Hub SoT for Hub instance; external read-only; merge documented | API | P1 | L1×Accept instance×conflict | Future |
| ECC-CAL-C420 | Corner | Add attendee not in channel during dual-create conflict resolution | 403 first; no orphan ATTENDEE in export; retry after ACL | API | P1 | C4×dual×export | Future |
| ECC-CAL-C421 | Corner | Organizer self-decline blocked **during** dual-create cancel of messenger leg only | calendar event remains; conflict updates; organizer still organizer | API | P0 | Dual cancel one leg | Future |
| ECC-CAL-C422 | Corner | Accept instance + ExportICal + LiveKit room start for that instance only (series rest future) | export marks instance; room only for due instance; no premature rooms | API | P0 | Accept instance×iCal×LK | Future |
| ECC-CAL-C423 | Corner | Two overlapping series (user in both) Accept instance A + Decline instance B same local DST day | busy reflects A; B free; conflict API accurate | API/E2E | P1 | Multi-series×DST×RSVP | Future |
| ECC-CAL-C424 | Corner | GDPR attendee delete + dual-create cleanup job + conflict API cache | cache invalidated; conflict no longer lists deleted user email | API | P0 | GDPR×dual×conflict cache | Future |

**Calendar Corner в этом файле: 25** (`ECC-CAL-C400`…`C424`)

---

## 3. hub-mail — ICS × reschedule × bounce · GDPR × iCal — `ECC-MAIL-C400+`

> Тройки L0 METHOD + bounce + Hub RSVP/reschedule; export/feed vs GDPR.  
> Не копировать C300–C324.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MAIL-C400 | Corner | L0 REQUEST delivered → reschedule UPDATE queued → bounce 5xx on UPDATE → attendee Accepts in-app on new time | Hub SoT new time; Accept OK; bounce logged; retry UPDATE not second REQUEST | API/Manual | P0 | ICS×reschedule×bounce | Future |
| ECC-MAIL-C401 | Corner | L0 REQUEST bounce + reschedule before retry succeeds | retry sends UPDATE/CANCEL+REQUEST per policy for **final** time only; no zombie first slot in client | Manual | P0 | Bounce×reschedule order | Future |
| ECC-MAIL-C402 | Corner | L0 UPDATE bounce + CANCEL race (organizer cancels while UPDATE retrying) | mailbox ends cancelled; retries stop; Hub cancelled; Accept 410 | API/Manual | P0 | UPDATE bounce×CANCEL | Future |
| ECC-MAIL-C403 | Corner | Reschedule twice (UPDATE1, UPDATE2) + bounce only on UPDATE1 + client applies late UPDATE1 after UPDATE2 | SEQUENCE/UID rules: ignore stale; Hub version wins; no revert time | Manual | P0 | Stale UPDATE×bounce | Future |
| ECC-MAIL-C404 | Corner | ICS REQUEST bounce + in-app Accept + second reschedule | Accept stays/resets per C8; mail retry chain SEQUENCE monotonic | API/Manual | P0 | Bounce×Accept×reschedule | Future |
| ECC-MAIL-C405 | Corner | Soft bounce (4xx) vs hard bounce (5xx) on reschedule UPDATE | soft retry; hard mark undeliverable + in-app banner; Hub time unchanged | API | P1 | Bounce class×UPDATE | Future |
| ECC-MAIL-C406 | Corner | Dual create → two L0 REQUEST → reschedule only calendar leg → bounce on calendar UPDATE | messenger ICS unchanged; calendar SoT updated; conflict warn refresh | Manual | P0 | Dual×ICS×reschedule×bounce | Future |
| ECC-MAIL-C407 | Corner | L0 CANCEL bounce after Hub already cancelled + attendee Decline in-flight | Decline 410; CANCEL retry idempotent; no reopen | API/Manual | P0 | CANCEL bounce×Decline | Future |
| ECC-MAIL-C408 | Corner | Provider 429 during reschedule storm (channel-wide attendees) + one hard bounce | backoff; in-app invites OK; bounced identity flagged; others delivered | API | P1 | 429×bounce×reschedule | Future |
| ECC-MAIL-C409 | Corner | L0 UPDATE with VTIMEZONE DST change + bounce + Accept instance | Hub instance time correct; retry ICS RECURRENCE-ID matches; Accept binds instance | Manual | P0 | DST×ICS×bounce×Accept | Future |
| ECC-MAIL-C410 | Corner | GDPR delete attendee after REQUEST sent + before UPDATE reschedule | CANCEL/anonymized to old address once; subsequent UPDATE omit PII; Hub purged | API/Manual | P0 | GDPR×ICS×reschedule | Future |
| ECC-MAIL-C411 | Corner | GDPR delete during L0 bounce-retry job holding email in payload | job redacts/drops; no PII in logs/queue after purge; audit | API | P0 | GDPR×bounce job | Future |
| ECC-MAIL-C412 | Corner | GDPR attendee + L1 iCal feed export poll | next poll removes ATTENDEE email; feed token cannot resurrect PII | API | P0 | GDPR×L1 export | Future |
| ECC-MAIL-C413 | Corner | GDPR + organizer ExportICal (.ics download) overlapping L0 CANCEL send | export anonymized; CANCEL outbound no email body PII beyond necessary envelope policy | API/Manual | P0 | GDPR×export×CANCEL | Future |
| ECC-MAIL-C414 | Corner | L1 feed SSRF attempt **during** reschedule fanout + bounce handler | ValidateOutboundURL reject; Hub reschedule OK; NF-07 | API | P0 | NF-07×reschedule×L1 | Future |
| ECC-MAIL-C415 | Corner | L1 LimitReader cap mid-VEVENT **while** Hub reschedule UPDATE generated | import fail-closed; Hub outbound ICS intact; no corrupt merge | API | P0 | LimitReader×reschedule | Future |
| ECC-MAIL-C416 | Corner | L2 IMAP ingest stale REQUEST after Hub reschedule + bounce on outbound UPDATE | Hub SoT; inbound stale ignored/quarantined; no time revert | API | P0 | L2×stale×bounce | Future |
| ECC-MAIL-C417 | Corner | L3 OAuth revoke mid reschedule push + L0 bounce fallback | Hub local consistent; L0 retry path; no half remote event | API | P0 | L3×reschedule×bounce | Future |
| ECC-MAIL-C418 | Corner | Triple channel: L0 UPDATE bounce + L1 still shows old + in-app Accept new | in-app SoT; L1 eventually converges; Accept not blocked by bounce | Manual | P0 | L0×L1×Accept×bounce | Future |
| ECC-MAIL-C419 | Corner | Email bounce + Telegram/in-app success + reschedule + Accept | badge stays; bounce ≠ hide invite; Accept OK | API | P1 | Multi-channel×bounce×Accept | Future |
| ECC-MAIL-C420 | Corner | METHOD=COUNTER (if ever) / unsupported METHOD during bounce recovery | ignore/quarantine; Hub SoT; no crash | API | P2 | ICS METHOD×bounce | Future |
| ECC-MAIL-C421 | Corner | Attachment ICS + HTML body reschedule + bounce after SMTP data phase unknown | idempotent job key; no duplicate SEQUENCE | API | P0 | SMTP unknown×reschedule | Future |
| ECC-MAIL-C422 | Corner | Export iCal feed token rotate **while** GDPR delete + reschedule mail | old token 401; new feed clean; mail uses new attendee set | API | P1 | Token rotate×GDPR×reschedule | Future |

**Mail Corner в этом файле: 23** (`ECC-MAIL-C400`…`C422`)

---

## 4. Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Calls + LiveKit (RSVP×rec×restart) | `ECC-CALL-C400`…`C422` | **23** |
| Calendar (DST×series×instance · dual×conflict · GDPR×iCal) | `ECC-CAL-C400`…`C424` | **25** |
| hub-mail (ICS×reschedule×bounce · GDPR×export) | `ECC-MAIL-C400`…`C422` | **23** |
| **Итого Corner** | | **71** |

| Проверка | Статус |
|----------|--------|
| Минимум 60 Corner | ✅ 71 ≥ 60 |
| Только тип Corner | ✅ |
| ID `C400+` only | ✅ |
| mediasoup / SEED-001 | ✅ N/A (не включено) |
| Parent C200 / wave1 C300 дословно | ✅ не копированы |
| Каталог `EDGE-CORNER-CATALOG.md` | ✅ не тронут |
| Git commit | ✅ не создавался |

**Фокус-покрытие wave2:**

| Тема | Примеры ID |
|------|------------|
| RSVP × recording × LiveKit restart | `ECC-CALL-C400`…`C422` |
| ICS × reschedule × bounce | `ECC-MAIL-C400`…`C409`, `C416`–`C421` |
| Dual messenger+calendar × conflict API | `ECC-CAL-C406`…`C411`, `C421`, `C424`; `ECC-CALL-C418`; `ECC-MAIL-C406` |
| DST × series × accept instance | `ECC-CAL-C400`…`C405`, `C417`, `C423`; `ECC-MAIL-C409` |
| GDPR attendee × export iCal | `ECC-CAL-C412`…`C416`, `C424`; `ECC-MAIL-C410`…`C413`, `C422`; `ECC-CALL-C414`, `C422` |

**Handoff:** `qa-manual-ver` → после `verified: true` выборочно `qa-auto1` (API races/conflict/GDPR export) / Manual (SFU restart + ICS clients).  
**Merge later:** в `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` § Corner expansion (отдельной волной; каталог сейчас не трогать).
