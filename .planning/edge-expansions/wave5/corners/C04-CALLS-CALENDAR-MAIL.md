# Corner Expansion C04 wave5 — Calls + Calendar + Mail

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave5  
> **Не трогает** `.planning/EDGE-CORNER-CATALOG.md`, parent `00-CORNER-PARENT.md`, wave1–4 `corners/C04-*`  
> **ID range:** `ECC-*-C700+` (не пересекает parent `C200–C215`, wave1 `C300+`, wave2 `C400+`, wave3 `C500+`, wave4 `C600+`, catalog `C01–C25`, edge `*100+`)  
> **SFU:** LiveKit stay — mediasoup / SEED-001 = **N/A** (строки mediasoup не добавлять)  
> **Фокус wave5:** webcam Ph16 × adaptive bitrate × egress layout · resource booking (US-CAL-01) × FindFreeSlots TOCTOU × LK · in-call chat × GDPR × calendar · codec/simulcast renegotiation · PWA/background lifecycle · hub-mail DKIM/TLS rotate × METHOD=REPLY · tzdb bump × series  
> **Источники:** JULY brief §3 · SEED-002 · US-M06 · US-CAL-01…04 · NFR-03/12/20 · SEC-02 · NF-02 · NF-07 · GDPR · TC-CALL-07/09 · Phase 16 webcam

**Назначение:** foresight-бэклог **новых** corner-комбинаций wave5 для будущих `TC-*`. Happy / Neg / Edge сюда не входят. **Не копировать** сценарии C200–C600 — только новые мультифакторные связки.

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (ещё не в каталоге) / **Future** (CAL-SYNC / mail L* / Ph16) |
| Слой | API / Unit / Manual / E2E / WS |
| Связь | Brief C# · SEED-002 · TC · NFR · theme (расширение, не копия) |

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — webcam Ph16 × adaptive bitrate × egress layout × in-call chat — `ECC-CALL-C700+`

> Wave5: webcam (Phase 16) + adaptive/simulcast + RoomComposite layout switch; in-call chat (TC-CALL-09) × GDPR; PWA/background tab freeze; codec renegotiation mid-egress.  
> Не повторять wave1 restart-primary, wave2 RSVP×rec×restart, wave3 multi-device×ICE×Seaweed как **единственную** ось.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-C700 | Corner | Webcam ON (Ph16) + screen-share + RoomComposite egress; host switches layout speaker→grid mid-recording | egress continues; layout change applied or marked degraded; call ≠ ended; file playable | Manual | P0 | webcam×share×egress layout, US-M06 Ph16 | Future |
| ECC-CALL-C701 | Corner | Webcam publish denied (browser) after Accept → join audio-only → host starts recording | recording audio(+screen if any); no crash waiting for camera; UI shows cam-off; RSVP stays accepted | Manual/E2E | P0 | cam deny×Accept×egress | Future |
| ECC-CALL-C702 | Corner | Adaptive bitrate / simulcast layer drop under bandwidth cap **while** webcam+share+egress ON | call stays; egress degraded-or-ok; no Hub panic; participant not kicked | Manual | P0 | ABR×webcam×egress | Future |
| ECC-CALL-C703 | Corner | Codec renegotiation VP8→H264 (or product pair) mid-call with active egress + late Accept joiner | renegotiation completes or fail-closed per track; late joiner gets current codec; egress not orphaned | Manual | P1 | codec reneg×egress×Accept | Future |
| ECC-CALL-C704 | Corner | Spotlight/pin participant A → egress layout follows → pin B during SFU brief brownout (not full restart) | pin state reconciled; egress layout eventual; no double-pin crash | Manual | P1 | pin×egress×brownout | Future |
| ECC-CALL-C705 | Corner | In-call chat (TC-CALL-09) burst + GDPR hard-delete of sender mid-flight + calendar inbox still open | chat messages redact/purge sender PII; NATS fanout safe; calendar invite rows without deleted email | API | P0 | chat×GDPR×calendar, TC-CALL-09 | Future |
| ECC-CALL-C706 | Corner | In-call chat text at max length + concurrent End Call + egress stop | chat 400 or last msg persisted; call ended once; egress finalize; no zombie room | API | P0 | chat max×end×egress | TODO |
| ECC-CALL-C707 | Corner | PWA / browser Page Lifecycle freeze 5m (tab discarded) mid-webcam call + egress; user returns | reconnect/rejoin policy; webcam re-prompt; egress continuity or clean gap; call not false-ended | Manual | P0 | PWA lifecycle×webcam×egress | Future |
| ECC-CALL-C708 | Corner | Background tab throttled JS timers + LiveKit keepalive + scheduled End from calendar cancel | cancel ends room; throttled client lands ended UI; no media to dead room | Manual | P0 | bg throttle×calendar cancel×LK | Future |
| ECC-CALL-C709 | Corner | Virtual background / blur ON (if shipped) + egress composite + GDPR face-retention policy | recording respects policy (blur baked or flagged); no raw face leak beyond policy; call OK | Manual | P2 | blur×egress×GDPR | Future |
| ECC-CALL-C710 | Corner | All participants JWT near 15m access / LiveKit grant expiry storm + webcam republish + egress | coordinated refresh; ≤1 rejoin storm; egress not cancelled by token refresh alone | Manual | P0 | token storm×webcam×egress | TODO |
| ECC-CALL-C711 | Corner | Host mute-all + webcam track still published + egress + attendee Decline from calendar | declined kicked; mute state cleared for remaining; egress continues without declinee | API/Manual | P0 | mute-all×webcam×Decline | Future |
| ECC-CALL-C712 | Corner | Raise-hand + webcam spotlight auto-follow (if policy) + recording metadata + second device Decline | hand/spotlight cleared for declinee; recording meta without declinee beyond policy | FE/Manual | P2 | hand×spotlight×Decline | Future |
| ECC-CALL-C713 | Corner | Resource-booked meeting room (US-CAL-01) linked call: LiveKit starts while resource double-book race loses | call may proceed or soft-block per product; resource conflict surfaced; no silent double-book | API/E2E | P0 | resource×LK start×conflict, US-CAL-01 | Future |
| ECC-CALL-C714 | Corner | FindFreeSlots suggested slot → Accept → LiveKit worker; concurrent book fills slot after suggestion | TOCTOU: warn/409 on start or Accept; no silent overlap hard-busy | API | P0 | FindFreeSlots×Accept×LK, US-CAL-04 | Future |
| ECC-CALL-C715 | Corner | Webcam resolution switch 720p→1080p mid-egress (product max) | clamp/reject over max; egress stable; no OOM encoder path | Manual | P1 | webcam res×egress cap | Future |
| ECC-CALL-C716 | Corner | Screen-share ends while webcam remains + egress layout was share-primary | layout falls back to speaker/grid; recording continuous; no black forever without mark | Manual | P1 | share end×webcam×egress | Future |
| ECC-CALL-C717 | Corner | Guest webcam publish attempt (policy deny) + member recording ON + guest chat try | guest cam/chat rejected per policy; recording without guest video; SEC no PII leak | Manual | P0 | guest cam×rec×chat, TC-CALL-04 | Future |
| ECC-CALL-C718 | Corner | LiveKit `track_published` webcam webhook delayed past End Call + calendar CANCEL already sent | late webhook ignored; call stays ended; CANCEL not reopened; SEC-02 still required | API | P0 | late webcam webhook×end×CANCEL | Future |
| ECC-CALL-C719 | Corner | Audio ducking / AGC spike during share+webcam+egress under CPU throttle | media continues; egress may degrade; Hub process no panic; UI recoverable | Manual | P2 | AGC×CPU×egress | Future |
| ECC-CALL-C720 | Corner | Concurrent host End + calendar series EXDATE today + webcam still publishing | room ended; EXDATE no resurrect; webcam unpublished; single terminal | API/Manual | P0 | End×EXDATE×webcam | Future |
| ECC-CALL-C721 | Corner | Download recording after webcam-heavy egress while attendee GDPR purge mid-download | download fail-closed after purge policy; partial bytes not cross-user; audit | API | P0 | download×webcam egress×GDPR | Future |
| ECC-CALL-C722 | Corner | In-call chat @mention calendar organizer + organizer already left LiveKit but event still active | mention notif path; no re-admit without join; NF-02 | API/WS | P1 | chat mention×left×event | Future |
| ECC-CALL-C723 | Corner | Adaptive subscribe (unsubscribe remote webcam) to save bandwidth + host pins that user + egress | pin may force subscribe or UI warn; egress still has pinned if SFU composite; documented | Manual | P1 | adaptive sub×pin×egress | Future |

**Calls Corner в этом файле: 24** (`ECC-CALL-C700`…`C723`)

---

## 2. Calendar — resource booking × FindFreeSlots × tzdb × series × LK — `ECC-CAL-C700+`

> Wave5: US-CAL-01 resources, US-CAL-04 FindFreeSlots TOCTOU, tzdb/zoneinfo bump mid-series, working-hours quiet window × LK start, CRM/intranet cross-link.  
> Не дублировать C300–C600 (DST×instance Accept, dual×conflict, all-day/EXDATE/transfer как единственная ось).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-C700 | Corner | Book resource R + Accept attendees + LiveKit due; second event books R same slot mid-Accept storm | resource exclusive; second 409/conflict; first Accept OK; one LK room for winner | API | P0 | resource×Accept×LK, US-CAL-01 | Future |
| ECC-CAL-C701 | Corner | Resource ACL revoke (user loses book right) **while** Accept in-flight on event holding R | Accept fail-closed or keep if already booked per policy; no orphan resource hold without event | API | P0 | resource ACL×Accept | Future |
| ECC-CAL-C702 | Corner | Resource soft-deleted (trash) while series future occurrences still reference R + next LK tick | fail-closed: block tick or clear resource; no ghost room for deleted resource | API | P0 | resource trash×series×LK | Future |
| ECC-CAL-C703 | Corner | FindFreeSlots returns slot S; organizer creates; attendee Accept; HR leave approved into S before LK start | conflict recompute; leave priority; Accept may warn; LK start policy documented | API | P0 | FindFreeSlots×leave×LK, US-CAL-04 | Future |
| ECC-CAL-C704 | Corner | FindFreeSlots TOCTOU: two organizers pick same S same ms → both create → both Accept paths | one/both conflict; busy correct; ≤1 hard-busy without override; rooms policy | API | P0 | FindFreeSlots race×Accept | Future |
| ECC-CAL-C705 | Corner | Team availability busy/free hides titles; Accept then ExportICal for non-member | export free/busy only for outsider; attendee sees title; no PII leak | API | P0 | avail privacy×Accept×iCal | Future |
| ECC-CAL-C706 | Corner | Working-hours / quiet-hours policy blocks LK auto-start; Accept already done; reminder fires in quiet window | no room until window opens **or** explicit override; Accept retained; reminder suppressed/queued | API | P1 | quiet-hours×Accept×LK | Future |
| ECC-CAL-C707 | Corner | tzdb / zoneinfo update (IANA bump) mid-series; wall times shift for attendee TZ; Accept instance after bump | Hub UTC stable; FE/ICS re-resolve; Accept binds correct RECURRENCE-ID; no phantom hour | API/E2E | P0 | tzdb×series×Accept, NFR-12 | Future |
| ECC-CAL-C708 | Corner | tzdb bump **during** L0 UPDATE fanout + in-app Accept | Hub version SoT; ICS VTIMEZONE matches post-bump; Accept not on stale wall | Manual | P0 | tzdb×ICS×Accept | Future |
| ECC-CAL-C709 | Corner | BYSETPOS / INTERVAL change on series while occurrence N LiveKit live + Accept on N+1 | live N isolated; N+1 uses new rule; no cross-link rooms | API | P0 | BYSETPOS×live LK×Accept | Future |
| ECC-CAL-C710 | Corner | Resource capacity=1; recurring weekly books R; EXDATE one week; another team books R that week + Accept | EXDATE frees R; second book OK; first series Accept unchanged | API | P0 | resource×EXDATE×Accept | Future |
| ECC-CAL-C711 | Corner | Intranet / CRM meeting deep-link creates CalendarEvent + messenger dual + FindConflicts | conflict lists sources; single SoT link; Accept one path; no triple busy | API | P1 | CRM/intranet×dual×conflict | Future |
| ECC-CAL-C712 | Corner | Ctrl+K search opens stale FindFreeSlots result card → Accept after slot filled | Accept/create 409/warn; search card invalidated; no silent overlap | E2E | P1 | search stale×FindFreeSlots×Accept | Future |
| ECC-CAL-C713 | Corner | Attendee Tentative + resource held + organizer upgrades to Accept-required policy mid-flight | Tentative may reset needs_action; resource hold retained; version bump | API | P1 | Tentative×resource×policy | Future |
| ECC-CAL-C714 | Corner | Soft-delete resource booking only (event remains) + Accept + LK start | event OK without R or block start per policy; Accept not 410 | API | P1 | unbook resource×Accept×LK | Future |
| ECC-CAL-C715 | Corner | GDPR delete resource contact/owner metadata on booked event + ExportICal + Accept others | resource id may remain; PII owner stripped; attendees OK | API | P0 | GDPR×resource meta×iCal | Future |
| ECC-CAL-C716 | Corner | Availability cache TTL stale shows free; Accept; conflict API on join path before LK | join/start re-checks; warn; cache invalidate | API | P0 | avail cache×Accept×LK | Future |
| ECC-CAL-C717 | Corner | Multi-resource event (room+equipment) partial book fail on equipment mid Accept | atomic: all-or-nothing holds; Accept not half-booked; retry clean | API | P0 | multi-resource atomic×Accept | Future |
| ECC-CAL-C718 | Corner | Series COUNT exhausted + organizer extends UNTIL while last occurrence LK recording uploads | extension new instances only; last recording isolated; Accept on new needs invite | API | P1 | COUNT→UNTIL×LK rec | Future |
| ECC-CAL-C719 | Corner | Floating TZ event + resource in building TZ + Accept + LK worker | normalize UTC; resource local hours respected; worker due correct | API | P1 | floating×resource×LK | Future |
| ECC-CAL-C720 | Corner | Organizer demoted from channel ACL while resource booking active + attendees Accepted | fail-closed transfer or cancel; Accepted get CANCEL/notif; LK not start for orphan | API | P0 | ACL demote×resource×Accept | Future |
| ECC-CAL-C721 | Corner | FindFreeSlots across DST spring week + resource closed Sundays + Accept Monday instance | slots skip closed; Accept Monday OK; no Sunday ghost book | API | P0 | FindFreeSlots×DST×resource | Future |
| ECC-CAL-C722 | Corner | Private event + resource name visible in conflict API to non-attendee | conflict shows busy/resource id policy without private title; no attendee emails | API | P0 | private×resource conflict privacy | Future |
| ECC-CAL-C723 | Corner | Restore event from trash with resource R now held by other + Accept retry | restore 409 or drop R; Accept path clear; no double-hold | API | P0 | trash restore×resource×Accept | Future |

**Calendar Corner в этом файле: 24** (`ECC-CAL-C700`…`C723`)

---

## 3. hub-mail — DKIM/TLS rotate × METHOD=REPLY × tzdb × resource ICS — `ECC-MAIL-C700+`

> Wave5: SMTP/DKIM/TLS certificate rotate mid-fanout; external METHOD=REPLY vs Hub RSVP; resource LOCATION in ICS; tzdb bump on outbound; complaint/FBL × GDPR; winmail.dat TNEF.  
> Не копировать C300–C600 (bounce×reschedule, MIME/OOO/forward, GDPR×export как единственная ось).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MAIL-C700 | Corner | DKIM key rotate mid L0 REQUEST fanout (half signed old) + Accept in-app | both signatures valid during grace **or** retry unsigned half; Accept SoT Hub; no drop invites | API/Manual | P0 | DKIM rotate×Accept | Future |
| ECC-MAIL-C701 | Corner | SMTP STARTTLS cert rotate mid-batch; connection reset; LiveKit room already open for early accepters | mail retry idempotent; room stays; late mail not second REQUEST; SEQUENCE ok | API/Manual | P0 | TLS rotate×LK live×ICS | Future |
| ECC-MAIL-C702 | Corner | External client METHOD=REPLY ACCEPTED while Hub still needs_action + concurrent in-app Decline | Hub version wins per policy; one terminal RSVP; no Accept+Decline split; LK join matches terminal | API | P0 | REPLY×Decline race×LK | Future |
| ECC-MAIL-C703 | Corner | METHOD=REPLY TENTATIVE from Outlook + Hub Accept on other device + resource held | single attendee state; resource hold consistent; L0 UPDATE optional | API/Manual | P0 | REPLY Tentative×Accept×resource | Future |
| ECC-MAIL-C704 | Corner | METHOD=REPLY after Hub cancel + LiveKit ended | ignore/410; no resurrect; mailbox CANCEL stands | API | P0 | REPLY after cancel×LK end | Future |
| ECC-MAIL-C705 | Corner | Outbound ICS LOCATION=resource name + resource renamed mid-fanout + Accept | Hub LOCATION final; retry UPDATE; Accept sees new name in-app | Manual | P1 | LOCATION resource×rename×Accept | Future |
| ECC-MAIL-C706 | Corner | tzdb bump between REQUEST and UPDATE; external client wall-time shifts; Hub Accept | Hub UTC SoT; UPDATE SEQUENCE++; Accept binds Hub; no silent hour shift in DB | Manual | P0 | tzdb×ICS×Accept | Future |
| ECC-MAIL-C707 | Corner | FBL / spam complaint on invite domain during series fanout + GDPR delete complainant | stop mail to address; Hub attendee purge/flag; series others continue; audit | API/Manual | P0 | FBL×GDPR×series | Future |
| ECC-MAIL-C708 | Corner | winmail.dat / TNEF L2 ingest with embedded ICS + Hub already has event from L0 | idempotent; TNEF parsed or quarantined; no duplicate; Accept path Hub | API | P1 | TNEF×L0×Accept | Future |
| ECC-MAIL-C709 | Corner | SMTPUTF8 / internationalized attendee email + GDPR purge + L0 CANCEL | CANCEL delivers or fail-closed logged; PII purged from queue; no ASCII corruption leak | API | P1 | SMTPUTF8×GDPR×CANCEL | Future |
| ECC-MAIL-C710 | Corner | Shared mailbox / alias: REPLY from alias not matching Hub user email + Accept attempt | map alias policy or reject; no wrong-user Accept; NF-02 | API | P0 | alias REPLY×Accept | Future |
| ECC-MAIL-C711 | Corner | Calendar ALARM + Hub quiet-hours + L0 delivered + LK auto-start suppressed | ALARM client-side only; Hub quiet holds start; Accept retained | Manual | P2 | ALARM×quiet×LK | Future |
| ECC-MAIL-C712 | Corner | Provider greylisting (4xx) on UPDATE after resource conflict resolved (new LOCATION) | retry UPDATE final LOCATION; Accept not blocked; SEQUENCE monotonic | API | P1 | greylist×resource UPDATE | Future |
| ECC-MAIL-C713 | Corner | L1 feed ETag/304 race: Accept changes projection mid Conditional-GET | 200 with new body **or** consistent 304 then next poll; no torn ATTENDEE list | API | P0 | ETag×Accept race, US-CAL-03 | Future |
| ECC-MAIL-C714 | Corner | L1 subscribe URL token rotate **while** METHOD=REPLY inbound + FindFreeSlots UI open | old feed 401; REPLY still maps Hub id; slots UI refresh; NF-07 on fetch | API | P1 | token×REPLY×FindFreeSlots | Future |
| ECC-MAIL-C715 | Corner | hub-mail worker crash after SMTP DATA accepted unknown + webcam call already recording | idempotent job; no duplicate ICS; recording unaffected; Accept OK | API | P0 | SMTP unknown×webcam rec×ICS | Future |
| ECC-MAIL-C716 | Corner | Marketing suppression list accidentally includes transactional invite recipient + Accept | transactional must still deliver **or** in-app SoT + ops alert; Accept OK | Manual | P0 | suppression×tx invite×Accept | Future |
| ECC-MAIL-C717 | Corner | ICS ATTENDEE CUTYPE=RESOURCE for Hub resource R + human Accept storm | resource row not human RSVP; humans Accept OK; no resource «Decline» | API | P1 | CUTYPE=RESOURCE×Accept | Future |
| ECC-MAIL-C718 | Corner | External REPLY proposes COUNTER new time overlapping resource booked elsewhere | Hub ignores COUNTER or queues for organizer; resource not auto-moved; Accept Hub time only | API/Manual | P1 | COUNTER×resource×Accept | Future |
| ECC-MAIL-C719 | Corner | DKIM fail quarantine on inbound REPLY + user Accepts in-app + LK join | Hub Accept SoT; inbound quarantined; join allowed; no RSVP revert | API | P0 | DKIM fail REPLY×Accept×LK | Future |
| ECC-MAIL-C720 | Corner | Rate 60/min + DKIM rotate + channel 40 invites + first 5 already in LiveKit lobby/join | early joiners OK; mail backlog; no second rooms from retries | API/Manual | P0 | rate×DKIM×LK early join | Future |
| ECC-MAIL-C721 | Corner | GDPR delete during TLS-retry holding MIME with webcam meeting DESCRIPTION link | redact blob; retry abort; description link dead/410; audit | API | P0 | GDPR×TLS retry×webcam link | Future |
| ECC-MAIL-C722 | Corner | L3 OAuth (if ADR) push fails mid resource rename + L0 fallback UPDATE | Hub local SoT; L0 UPDATE sent; remote not half-named; Accept Hub | API | P1 | L3 fail×resource×L0 | Future |
| ECC-MAIL-C723 | Corner | Triple: external REPLY Accept + in-app Tentative + calendar cancel same version window | cancel wins; REPLY/Tentative 410; LK not start; mail CANCEL | API | P0 | REPLY×Tentative×cancel | Future |

**Mail Corner в этом файле: 24** (`ECC-MAIL-C700`…`C723`)

---

## 4. Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Calls + LiveKit (webcam×ABR×egress×chat×PWA) | `ECC-CALL-C700`…`C723` | **24** |
| Calendar (resource×FindFreeSlots×tzdb×series×LK) | `ECC-CAL-C700`…`C723` | **24** |
| hub-mail (DKIM/TLS×REPLY×tzdb×resource ICS) | `ECC-MAIL-C700`…`C723` | **24** |
| **Итого Corner** | | **72** |

| Проверка | Статус |
|----------|--------|
| Минимум 60 Corner | ✅ 72 ≥ 60 |
| Только тип Corner | ✅ |
| ID `C700+` only | ✅ |
| mediasoup / SEED-001 | ✅ N/A (не включено) |
| Parent C200 / wave1–4 C300–C600 дословно | ✅ не копированы |
| Каталог `EDGE-CORNER-CATALOG.md` | ✅ не тронут |
| Git commit | ✅ не создавался |

**Фокус-покрытие wave5:**

| Тема | Примеры ID |
|------|------------|
| Webcam Ph16 × egress layout / ABR / codec | `ECC-CALL-C700`…`C704`, `C715`–`C717`, `C723` |
| In-call chat × GDPR × calendar | `ECC-CALL-C705`–`C706`, `C722` |
| PWA / background lifecycle × LK | `ECC-CALL-C707`–`C708` |
| Resource booking × Accept × LK | `ECC-CAL-C700`…`C702`, `C710`, `C714`, `C717`; `ECC-CALL-C713` |
| FindFreeSlots TOCTOU × leave/conflict | `ECC-CAL-C703`–`C704`, `C712`, `C716`, `C721`; `ECC-CALL-C714` |
| tzdb bump × series / ICS | `ECC-CAL-C707`–`C708`; `ECC-MAIL-C706` |
| DKIM/TLS rotate × fanout × LK | `ECC-MAIL-C700`–`C701`, `C715`, `C720`–`C721` |
| METHOD=REPLY × Hub RSVP × LK | `ECC-MAIL-C702`–`C704`, `C710`, `C719`, `C723` |
| Resource LOCATION / CUTYPE in ICS | `ECC-MAIL-C705`, `C712`, `C717`–`C718` |

**Handoff:** `qa-manual-ver` → после `verified: true` выборочно `qa-auto1` (resource/FindFreeSlots races, REPLY, ETag) / Manual (webcam Ph16, ABR, DKIM/TLS, PWA).  
**Merge later:** в `.planning/EDGE-CORNER-CATALOG.md` § Corner expansion (отдельной волной; каталог сейчас не трогать).
