# Corner Expansion C04 wave4 — Calls + Calendar + Mail

> **Тип:** только **Corner** (редкие комбинации, races, mid-flight, multi-step, cross-domain)  
> **Дата:** 2026-07-10 · **Владелец:** `qa-manual` · **Волна:** wave4  
> **Не трогает** `.planning/EDGE-CORNER-CATALOG.md`, parent `00-CORNER-PARENT.md`, wave1–3 `corners/C04-*`  
> **ID range:** `ECC-*-C600+` (не пересекает parent `C200–C215`, wave1 `C300+`, wave2 `C400+`, wave3 `C500+`, catalog `C01–C25`, edge `*100+`)  
> **SFU:** LiveKit stay — mediasoup / SEED-001 = **N/A** (строки mediasoup не добавлять)  
> **Фокус wave4:** webcam×ICE/TURN×egress · multi-device×host-drain×RBAC mid-call · Hub-WS desync×LiveKit media · user-TZ flip×all-day×EXDATE · quiet-hours×due-worker · METHOD=REPLY×OOO×delayed ICS · charset/MIME×alias GDPR  
> **Источники:** JULY brief §3 · SEED-002 · US-M06 · Phase 16 webcam · US-CAL-* · NFR-03/12 · SEC-02 · NF-02/07 · GDPR soft/hard delete

**Назначение:** foresight-бэклог **новых** corner-комбинаций wave4 для будущих `TC-*`. Happy / Neg / Edge сюда не входят. **Не копировать** сценарии C200–C500 — только новые мультифакторные связки.

| Колонка | Значение |
|---------|----------|
| Gap | **TODO** (ещё не в каталоге) / **Future** (CAL-SYNC / mail L* / Phase 16) |
| Слой | API / Unit / Manual / E2E / WS |
| Связь | Brief C# · SEED-002 · TC · NFR · theme (расширение, не копия) |

Формат: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. Calls + LiveKit — webcam × ICE/TURN × multi-device × Hub-WS desync — `ECC-CALL-C600+`

> Тройки/четверки wave4: Phase-16 webcam + ICE/TURN failover + egress; multi-device same identity + host drain; Hub WS down while LiveKit media up; RBAC `canPublish` flip mid-call.  
> Wave1 C300 / wave2 C400 / wave3 C500 = другие комбинации; здесь — **media-path / identity / control-plane desync**.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CALL-C600 | Corner | Webcam ON (Phase 16) + ICE fail → TURN relay + egress recording active | video republish via TURN; egress continues or clean gap; call ≠ ended; no orphan Seaweed | Manual | P0 | webcam×TURN×egress, US-M06 | Future |
| ECC-CALL-C601 | Corner | Webcam HD + simulcast layer drop (bandwidth) + egress composite + SFU node drain | lower layer selected; recording quality policy documented; drain reconnect keeps call_id | Manual | P0 | webcam×simulcast×drain | Future |
| ECC-CALL-C602 | Corner | Webcam permission revoke mid-call (OS/browser) + mute state + egress | track unpublished; UI shows camera-off; egress audio-continues; no crash | Manual | P1 | webcam revoke×egress | Future |
| ECC-CALL-C603 | Corner | Webcam + screen-share both published + TURN-only path + late joiner Accept | late joiner receives both after subscribe; TURN OK; recording layout policy | Manual | P1 | webcam×share×TURN×Accept | Future |
| ECC-CALL-C604 | Corner | Multi-device same user: desktop in-call + mobile join same identity + host drain | policy: kick oldest **or** dual-presence documented; drain reconnect one session; count sane | API/Manual | P0 | multi-device×drain, NFR-03 | Future |
| ECC-CALL-C605 | Corner | Multi-device: mobile holds host + desktop End Call + LiveKit auto-empty race | single ended; both devices show ended; no zombie room | API/E2E | P0 | multi-device×end | Future |
| ECC-CALL-C606 | Corner | Multi-device mute-all from device A while device B unmutes + SFU drain | final mute state reconciled; no stuck unmuted after drain | Manual | P1 | multi-device×mute×drain | Future |
| ECC-CALL-C607 | Corner | Host transfer (A→B) mid-drain while A disconnects | B becomes host; room survives; recording ownership Hub-side unchanged | API/Manual | P0 | host-transfer×drain | Future |
| ECC-CALL-C608 | Corner | Hub WS disconnect (NATS lag) while LiveKit media healthy + participant_joined webhook | media stays; Hub UI may stale then reconcile; webhook SEC-02 applied; count matches LK | API/WS | P0 | Hub-WS×LK media desync, SEC-02 | TODO |
| ECC-CALL-C609 | Corner | Hub WS reconnect storm + LiveKit room briefly empty (all ICE restart) + auto-end timer | auto-end must **not** fire if reconnect window policy; or documented end; no false toast | Manual | P0 | empty-window×auto-end×WS | TODO |
| ECC-CALL-C610 | Corner | Hub marks call ended (API) while LiveKit tracks still publishing 2s | FE force-ended; LK room delete; late webhooks ignored by call_id; media cut | API | P0 | control-plane lead×LK | TODO |
| ECC-CALL-C611 | Corner | LiveKit `room_finished` arrives while Hub still `active` (ops killed SFU) | Hub transitions ended/failed; UI; no rejoin; audit | API | P0 | SFU-kill×Hub lag | TODO |
| ECC-CALL-C612 | Corner | Channel RBAC demote publisher→subscriber mid-call (`canPublish` false) + webcam + egress | Hub kicks publish rights; tracks unpublished; egress continues for remaining publishers; NF-02 | API/Manual | P0 | RBAC mid-call×webcam×egress, NF-02 | Future |
| ECC-CALL-C613 | Corner | RBAC promote subscriber→publisher during SFU drain + token refresh | new publish allowed only after fresh token; no ghost publish on old JWT | API/Manual | P0 | RBAC×token×drain | Future |
| ECC-CALL-C614 | Corner | Member removed from channel while Hub-WS down and LiveKit still connected | fail-closed: kick on next Hub reconcile or LK grant revoke; cannot stay forever | API/Manual | P0 | ACL×WS-down×LK, NF-02 | Future |
| ECC-CALL-C615 | Corner | Audio-only switch (unpublish video) during egress + TURN failover | egress layout adapts; audio continuous; no failed-complete race | Manual | P1 | audio-only×egress×TURN | Future |
| ECC-CALL-C616 | Corner | Codec mismatch one client (H264 fail) + others VP8 OK + reconnect after drain | failing client fallback or clear error; room stable; recording from OK publishers | Manual | P1 | codec×drain | Future |
| ECC-CALL-C617 | Corner | Browser background tab throttle 5m + burst rejoin + max participants-1 | count never > max; throttle rejoin queued; no negative count | Manual | P0 | tab-throttle×max, NFR-03 | TODO |
| ECC-CALL-C618 | Corner | Participant disconnect timeout (LK) vs Hub heartbeat still green race | one source of truth for «in call»; UI consistent within SLA; no double-leave audit | API | P1 | LK timeout×Hub heartbeat | TODO |
| ECC-CALL-C619 | Corner | Egress Seaweed partial PUT (network blip) + call continues + second egress slot | partial object not downloadable; retry/fail-closed; ≤2 active; no corrupt ready | API | P0 | Seaweed partial×egress max=2 | TODO |
| ECC-CALL-C620 | Corner | Egress finalize + member JWT refresh + host leaves last overlapping | finalize completes or cancelled cleanly; room gone; download ACL for remaining members | API | P0 | finalize×JWT×last-leave | TODO |
| ECC-CALL-C621 | Corner | Data-channel / in-call chat message in-flight + SFU drain + Hub WS reconnect | message at-most-once or reconnect replay policy; no crash; chat ≠ media path | WS/Manual | P2 | datachannel×drain×WS | Future |
| ECC-CALL-C622 | Corner | Raise-hand + mute-all + webcam republish during drain window (triple UI) | no crash; hands cleared or reconciled; mute-all wins until unmute | FE/Manual | P2 | hand×mute×webcam×drain | Future |
| ECC-CALL-C623 | Corner | Calendar-linked call: trash CalendarEvent mid-call + webcam + egress | fail-closed end **or** detach link keep room; Join CTA from calendar 410; egress stop policy | API/Manual | P0 | trash×LK mid-call, SEED-002 | Future |
| ECC-CALL-C624 | Corner | Scheduled due-worker starts LK room while quiet-hours policy blocks notify + Accept already | room may start for accepted; notifications suppressed; Join available in-app | API | P1 | quiet-hours×due×Accept | Future |
| ECC-CALL-C625 | Corner | Guest webcam attempt (if UI exposes) + recording ON + host drain | guest camera policy enforced; recording no unexpected guest PII beyond policy; SEC | Manual | P0 | guest×webcam×rec×drain, TC-CALL-04 | Future |
| ECC-CALL-C626 | Corner | Webhook storm (joins/leaves) during TURN failover + Hub clock skew clamp | SEC-02 required; idempotent counts; skew clamp; Hub stays up | API | P0 | webhook storm×TURN×skew, SEC-02 | TODO |
| ECC-CALL-C627 | Corner | LiveKit grant token refresh mid-screen-share + webcam + egress | share/camera stay or re-prompt once; egress continuous; no end | Manual | P1 | token refresh×share×webcam | Future |
| ECC-CALL-C628 | Corner | Region/node migration mid-call (LK) + multi-device + recording | both devices reconnect to new node; recording continuity/gap documented; call_id stable | Manual | P0 | region migrate×multi-device×rec | Future |
| ECC-CALL-C629 | Corner | Early lobby join before scheduled start + organizer cancel + guest token | lobby cleared; tokens revoked; no media room leak; calendar shows cancelled | API/Manual | P0 | lobby×cancel×guest | Future |

**Calls Corner в этом файле: 30** (`ECC-CALL-C600`…`C629`)

---

## 2. Calendar — user-TZ × all-day × EXDATE × quiet-hours × due-worker — `ECC-CAL-C600+`

> Wave4: смена TZ пользователя mid-series, all-day/overnight, EXDATE / this-and-following, quiet-hours × LiveKit due-worker, search-index lag, ACL diverge.  
> Не дублировать C300–C500 (DST instance Accept, dual-create conflict, GDPR×iCal export — другие связки).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-CAL-C600 | Corner | User changes profile TZ mid-series after series Accept; next instance due-worker fires LiveKit | Hub UTC SoT; wall times follow new TZ; room starts correct UTC; inbox times refresh | API/E2E | P0 | user-TZ×series×LK due, NFR-12 | Future |
| ECC-CAL-C601 | Corner | TZ flip during Accept in-flight on instance (old TZ client) | Accept binds UTC/RECURRENCE-ID; FE shows new TZ after; no wrong-day busy | API | P0 | TZ flip×Accept race | Future |
| ECC-CAL-C602 | Corner | All-day event Accept + overlapping timed LiveKit call same local day + conflict API | all-day busy policy vs timed; conflict distinguishes; both SoT clear | API | P0 | all-day×timed×conflict×LK | Future |
| ECC-CAL-C603 | Corner | Overnight all-day spanning TZ boundary + Accept + expand 366d | single logical day(s); no dup/missing; Accept OK | Unit/API | P0 | overnight×TZ×expand | Future |
| ECC-CAL-C604 | Corner | Floating TIME (no TZ) imported → Hub stores UTC policy + Accept + ICS out | conversion documented; Accept stable; outbound VTIMEZONE consistent | API/Manual | P1 | floating×Accept×ICS | Future |
| ECC-CAL-C605 | Corner | EXDATE removes next occurrence while attendee Accepts that RECURRENCE-ID in-flight | Accept 410/gone; EXDATE wins; no room for excluded instance | API | P0 | EXDATE×Accept race | Future |
| ECC-CAL-C606 | Corner | «This and following» delete mid-series + Accept on later remaining instance + due-worker | deleted range gone; remaining Accept OK; LK only for kept instances | API | P0 | this+following×Accept×LK | Future |
| ECC-CAL-C607 | Corner | EXDATE + RRULE BYDAY change same transaction + Tentative on old BYDAY | version bump; Tentative re-eval or reset; no ghost BYDAY | API | P0 | EXDATE×BYDAY×Tentative | Future |
| ECC-CAL-C608 | Corner | Quiet-hours suppress push while Accept + due-worker creates LiveKit room | in-app Join visible; push delayed/suppressed; room exists for accepted | API | P1 | quiet-hours×Accept×LK | Future |
| ECC-CAL-C609 | Corner | Quiet-hours end exactly at DTSTART + burst notify + SFU start | no duplicate storm beyond policy; single room; Accept still valid | API | P1 | quiet-hours edge×LK | Future |
| ECC-CAL-C610 | Corner | Working-hours policy blocks scheduling outside hours **while** override Accept on existing | existing event OK; new create blocked; Accept not used to bypass create ACL | API | P1 | working-hours×Accept | Future |
| ECC-CAL-C611 | Corner | Calendar search index lag: stale hit → Accept after organizer cancel | Accept 410; search eventually consistent; no resurrect | API | P0 | search lag×Accept×cancel | Future |
| ECC-CAL-C612 | Corner | Search hit for series instance already EXDATE'd + deep-link Join | 410; no LK token issued | API | P0 | search×EXDATE×Join | Future |
| ECC-CAL-C613 | Corner | Channel ACL allows write; calendar ACL denies attendee add mid dual-path create | fail-closed 403; no orphan CalendarAttendee; no LK schedule | API | P0 | ACL diverge×create | Future |
| ECC-CAL-C614 | Corner | Calendar ACL revoke during Accept + LiveKit token prefetch | Accept 403; prefetch token revoked/unused; NF-02 | API | P0 | ACL×Accept×token | Future |
| ECC-CAL-C615 | Corner | Linked issue due-date move + calendar event reschedule race + conflict API | both versioned; conflict recompute; no silent overwrite | API | P1 | issue-link×reschedule×conflict | Future |
| ECC-CAL-C616 | Corner | Sprint deadline overlay + Accept timed call overlapping + leave half-day | availability merge: leave > accept hard/soft per policy; UI badges distinct | API | P1 | sprint×leave×Accept | Future |
| ECC-CAL-C617 | Corner | Reminder/VALARM fire after cancel + late Accept from stale push | Accept 410; reminder suppressed after cancel; no room | API | P0 | reminder×cancel×Accept | Future |
| ECC-CAL-C618 | Corner | Hub reminder + L0 VALARM dual-fire + due-worker LK start | at-most-one Join CTA storm policy; single room | Manual | P1 | dual reminder×LK | Future |
| ECC-CAL-C619 | Corner | Propose-new-time / COUNTER (if enabled) + organizer Accept counter + old instance due-worker | old slot cancelled; new slot scheduled; worker no room on old | API | P1 | COUNTER×due-worker | Future |
| ECC-CAL-C620 | Corner | Restore from trash during due-worker already creating LK for pre-trash Accept | restore policy: no auto-room **or** attach existing; no dual rooms | API | P0 | restore×due-worker×LK | Future |
| ECC-CAL-C621 | Corner | Soft-delete → restore within 30d while user TZ changed + series expand | times correct in new TZ; RSVP policy on restore; no dup keys | API/E2E | P1 | restore×TZ×series | Future |
| ECC-CAL-C622 | Corner | Attendee cap storm: add 50th while Accept races on 49th + mail fanout | cap enforced; partial set consistent; no Accept for rejected add | API | P1 | attendee-cap×Accept | Future |
| ECC-CAL-C623 | Corner | Import ICS overlapping UID with Hub event + Accept Hub + L1 feed poll | Hub SoT; import quarantine/merge policy; Accept unaffected | API | P0 | import UID×Accept×L1 | Future |
| ECC-CAL-C624 | Corner | Corrupt RRULE COUNT=0 / UNTIL before DTSTART via API + Accept attempt | 400 on create/patch; Accept never on invalid series | API | P0 | corrupt RRULE×Accept | Future |
| ECC-CAL-C625 | Corner | Organizer role transfer (founder demote path) mid-series + Accept + due-worker | new organizer owns; Accept valid; LK host mapping updated | API | P0 | organizer transfer×Accept×LK | Future |
| ECC-CAL-C626 | Corner | Multi-overlay personal+channel calendars: Accept channel event hidden on personal filter | SoT accepted; filter UX; Join still from inbox | E2E | P2 | overlay filter×Accept | Future |
| ECC-CAL-C627 | Corner | CRM/intranet meeting dual-book same slot (v4) + Hub Accept + conflict API | Future: conflict lists both sources; Hub event SoT for Hub RSVP | API | P2 | v4 dual×Accept×conflict | Future |
| ECC-CAL-C628 | Corner | Notification digest batch reorders Accept before invite visible + second device Decline | one terminal RSVP; digest not create second attendee row | API/WS | P1 | digest order×Accept×Decline | Future |
| ECC-CAL-C629 | Corner | All-day Accept + EXDATE next day + quiet-hours + search lag Join | Join only for accepted non-EXDATE; quiet suppress OK; search 410 on EXDATE | API | P0 | all-day×EXDATE×quiet×search | Future |

**Calendar Corner в этом файле: 30** (`ECC-CAL-C600`…`C629`)

---

## 3. hub-mail — METHOD=REPLY × OOO × delayed ICS × charset/alias — `ECC-MAIL-C600+`

> Wave4: клиентский METHOD=REPLY vs in-app RSVP; OOO/autoresponder loops; delayed delivery after LK ended; MIME/charset; plus-alias GDPR; IMAP UIDVALIDITY; feed ETag.  
> Не копировать C300–C500 (reschedule×bounce, GDPR×export iCal — другие оси).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MAIL-C600 | Corner | Outlook METHOD=REPLY PARTSTAT=ACCEPTED arrives while in-app still needs_action + due-worker about to start LK | Hub applies REPLY or maps to Accept; single SoT; ≤1 room; no double fanout | API/Manual | P0 | REPLY×Accept×LK due | Future |
| ECC-MAIL-C601 | Corner | Apple REPLY TENTATIVE + in-app Decline same ms | one terminal; ICS SEQUENCE/PARTSTAT consistent outbound; no LK for declined | API | P0 | REPLY×Decline race | Future |
| ECC-MAIL-C602 | Corner | Client REPLY ACCEPTED on **old** SEQUENCE after Hub reschedule | ignore/stale; needs_action or keep per C8; no revert time; no premature LK on old slot | API/Manual | P0 | stale REPLY×reschedule | Future |
| ECC-MAIL-C603 | Corner | OOO autoresponder loop on REQUEST (mailer-daemon storm) + in-app Accept | Hub rate-limits/suppresses loop; Accept OK; no SEQUENCE explosion | API/Manual | P0 | OOO loop×Accept | Future |
| ECC-MAIL-C604 | Corner | OOO reply mis-parsed as METHOD=REPLY ACCEPTED | quarantine; attendee stays needs_action; no false Accept; no LK | API | P0 | OOO false Accept | Future |
| ECC-MAIL-C605 | Corner | Delayed ICS REQUEST delivered 48h later; event already started; LK room ended; user clicks Accept in mail | Accept 410/gone or post-hoc policy; cannot reopen room; Hub SoT ended | Manual | P0 | delayed ICS×LK ended | Future |
| ECC-MAIL-C606 | Corner | Delayed UPDATE after CANCEL already applied + in-app Decline in-flight | CANCEL wins; Decline 410; delayed UPDATE ignored | API/Manual | P0 | delayed UPDATE×CANCEL | Future |
| ECC-MAIL-C607 | Corner | Delayed REPLY after GDPR purge of attendee email | drop/ignore; no PII re-create; audit | API | P0 | delayed REPLY×GDPR | Future |
| ECC-MAIL-C608 | Corner | Charset ISO-8859-5 / broken UTF-8 in SUMMARY + Accept in-app + L0 UPDATE | Hub stores sanitized Unicode; no XSS; UPDATE readable; Accept OK | API/Manual | P0 | charset×Accept×ICS | Future |
| ECC-MAIL-C609 | Corner | MIME multipart: text/calendar vs HTML body disagree on DTSTART + Accept | Hub SoT from structured event not HTML; Accept binds Hub time | Manual | P0 | MIME mismatch×Accept | Future |
| ECC-MAIL-C610 | Corner | ICS DESCRIPTION HTML injection + FE inbox render + Accept | sanitized render; Accept OK; no script exec | E2E | P0 | XSS ICS×Accept | Future |
| ECC-MAIL-C611 | Corner | Plus-alias attendee `user+tag@` Accept in-app + GDPR delete primary mapping | alias+primary purged; mail CANCEL/anonymize both; no LK token for purged | API | P0 | alias×GDPR×Accept | Future |
| ECC-MAIL-C612 | Corner | Secondary email REPLY while primary Decline in-app | identity merge policy; one terminal; no dual attendee rows | API | P0 | secondary REPLY×Decline | Future |
| ECC-MAIL-C613 | Corner | Distribution-list expand on SMTP vs Hub attendee snapshot diverge + Accept one member | only Hub attendees can Accept; DL extras ignored/no ACL grant | API/Manual | P1 | DL expand×Accept | Future |
| ECC-MAIL-C614 | Corner | Message-ID / In-Reply-To break on UPDATE → client creates duplicate + Accept both UIDs | Hub single UID SoT; duplicate client event not create second Hub event | Manual | P1 | threading×Accept | Future |
| ECC-MAIL-C615 | Corner | Attachment size near provider limit: ICS + inline image + reschedule UPDATE | fail/backoff; Hub time updated; in-app SoT; retry without image or split policy | API | P1 | size limit×UPDATE | Future |
| ECC-MAIL-C616 | Corner | DKIM/SPF fail → quarantine folder; user Accepts in-app; later ICS appears | Accept OK; late ICS not duplicate create; SEQUENCE ok | Manual | P1 | SPF fail×Accept | Future |
| ECC-MAIL-C617 | Corner | Spam complaint (FBL) mid-series + next REQUEST + Accept | suppress further mail to complainant per policy; in-app invites continue; Accept OK | API | P1 | FBL×series×Accept | Future |
| ECC-MAIL-C618 | Corner | hub-mail rate limit during channel-wide quiet-hours release + LK due burst | backoff; in-app Join OK; no lost Accept; rooms for accepted only | API | P0 | rate×quiet×LK due | Future |
| ECC-MAIL-C619 | Corner | L2 IMAP UIDVALIDITY reset mid-ingest of REPLY + Hub Accept concurrent | Hub SoT wins; ingest re-sync no revert RSVP; no dup | API | P0 | UIDVALIDITY×REPLY×Accept | Future |
| ECC-MAIL-C620 | Corner | L2 poison message (malformed VEVENT) + retry queue + Accept path | poison quarantined; Accept unaffected; worker not stuck | API | P0 | poison×Accept | Future |
| ECC-MAIL-C621 | Corner | L1 feed ETag race: Hub cancel + client poll old ETag + Accept deep-link | Accept 410; next poll removes VEVENT; NF-07 on feed URL | API | P0 | ETag×cancel×Accept, NF-07 | Future |
| ECC-MAIL-C622 | Corner | L1 LimitReader mid-ATTENDEE while plus-alias GDPR purge | import fail-closed; purge complete; no email leak in feed | API | P0 | LimitReader×alias GDPR | Future |
| ECC-MAIL-C623 | Corner | L3 OAuth token refresh fail mid METHOD=REPLY push + in-app Accept | Hub local Accept OK; remote eventually consistent or flagged; no half state | API | P1 | L3 OAuth×REPLY×Accept | Future |
| ECC-MAIL-C624 | Corner | Relay failover mid SMTP DATA (unknown accept) + SEQUENCE on retry + LK due | idempotent job; no dup SEQUENCE; due-worker single room | API | P0 | SMTP unknown×SEQUENCE×LK | Future |
| ECC-MAIL-C625 | Corner | List-Unsubscribe / bulk headers strip calendar client support + Accept only in-app | in-app SoT; mail may lack UI buttons; Accept OK | Manual | P2 | bulk headers×Accept | Future |
| ECC-MAIL-C626 | Corner | Encrypted/forwarded ICS stripped of METHOD + deep-link Accept + guest policy | deep-link uses Hub auth; stripped ICS ignored; no guest elevation | Manual | P1 | stripped ICS×Accept | Future |
| ECC-MAIL-C627 | Corner | Soft-delete mail template mid fanout + GDPR attendee + reschedule | jobs use safe fallback template; no PII; Hub reschedule OK | API | P1 | template delete×GDPR×reschedule | Future |
| ECC-MAIL-C628 | Corner | Triple: delayed REPLY ACCEPTED + OOO storm + LK room already live for organizer-only | late REPLY may join policy or ignore; OOO suppressed; participant ACL enforced | Manual | P0 | delayed REPLY×OOO×LK live | Future |
| ECC-MAIL-C629 | Corner | Charset-corrupt REPLY + alias GDPR + L1 poll same second | REPLY dropped/sanitized; purge wins; feed clean; no 500 | API | P0 | charset×alias×L1 | Future |

**Mail Corner в этом файле: 30** (`ECC-MAIL-C600`…`C629`)

---

## 4. Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Calls + LiveKit (webcam×TURN×multi-device×WS desync) | `ECC-CALL-C600`…`C629` | **30** |
| Calendar (user-TZ×all-day×EXDATE×quiet-hours×due) | `ECC-CAL-C600`…`C629` | **30** |
| hub-mail (REPLY×OOO×delayed×charset×alias) | `ECC-MAIL-C600`…`C629` | **30** |
| **Итого Corner** | | **90** |

| Проверка | Статус |
|----------|--------|
| Минимум 60 Corner | ✅ 90 ≥ 60 |
| Только тип Corner | ✅ |
| ID `C600+` only | ✅ |
| LiveKit stay (не mediasoup) | ✅ |
| Parent C200 / wave1–3 C300–C500 дословно | ✅ не копированы |
| Каталог `EDGE-CORNER-CATALOG.md` | ✅ не тронут |
| Git commit | ✅ не создавался |

**Фокус-покрытие wave4:**

| Тема | Примеры ID |
|------|------------|
| Webcam × ICE/TURN × egress / drain | `ECC-CALL-C600`…`C603`, `C615`–`C616`, `C625`, `C627` |
| Multi-device × host-transfer × drain | `ECC-CALL-C604`…`C607`, `C628` |
| Hub-WS desync × LiveKit media / auto-end | `ECC-CALL-C608`…`C611`, `C614`, `C618` |
| RBAC mid-call × canPublish × ACL | `ECC-CALL-C612`…`C614` |
| User-TZ flip × series × LK due | `ECC-CAL-C600`…`C601`, `C621` |
| All-day / overnight × timed conflict | `ECC-CAL-C602`…`C604`, `C629` |
| EXDATE / this-and-following × Accept × LK | `ECC-CAL-C605`…`C607`, `C612`, `C629` |
| Quiet-hours × due-worker × notify | `ECC-CAL-C608`…`C610`; `ECC-MAIL-C618` |
| METHOD=REPLY × OOO × delayed ICS | `ECC-MAIL-C600`…`C607`, `C628` |
| Charset/MIME/XSS × alias GDPR × L1/L2 | `ECC-MAIL-C608`…`C612`, `C619`–`C622`, `C629` |

**Handoff:** `qa-manual-ver` → после `verified: true` выборочно `qa-auto1` (ACL/REPLY/EXDATE/UIDVALIDITY) / Manual (webcam×TURN×drain, delayed ICS clients).  
**Merge later:** в `.planning/EDGE-CORNER-CATALOG.md` § Corner expansion (отдельной волной; каталог сейчас не трогать).
