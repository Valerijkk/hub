# Corner Expansion C02 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Corner** (races, mid-flight, multi-actor, partial fail, revoke-after-share)  
> **ID ranges:** `ECC-TRCK-C300+`, `ECC-SPR-C300+`, `ECC-TMS-C300+` — **не пересекают** parent `ECC-*-C200…C206` и edge `*100+` / канон `001–040`  
> **Не заменяет** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`, `00-CORNER-PARENT.md`, `docs/TESTING.md`  
> **Parent seeds (не дублировать 1:1):** TRCK-C200…C206 · SPR-C200…C202 · TMS-C200…C202  
> **Фокус:** optimistic lock races · workflow+assignee races · WIP mid-drag · soft-delete+key · bulk partial fail · sprint complete races · Allure import mid-delete · run complete vs execute · version rollback mid-edit · JQL share after revoke  
> **Связи:** US-M03/M11/M12, US-S01, US-M05/S13/C02/C04 · NFR-11/12/24 · NF-03 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*, TC-JQL-*

### Gap-легенда

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated corner-теста / TC |
| **Partial** | Happy/unit есть, race/mid-flight не закрыт |
| **Future** | Поведение ещё не зафиксировано в ADR/политике |

---

## 1. Tracker — `ECC-TRCK-C300+`

> Углубление parent C200–C206: optimistic lock, workflow∥assignee, WIP mid-drag, soft-delete+key, bulk partial, JQL revoke

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-C300 | Corner | A/B: A PATCH title v=1; B PATCH description v=1 одновременно | оба не silent-OK; один 200 version=2, другой 409 | API | P0 | NFR-24, TC-TRCK-03/04, C200 | TODO |
| ECC-TRCK-C301 | Corner | A saves; B retries с stale v без re-GET | 409; UI merge/reload, не overwrite | E2E | P0 | C200, ECC-TRCK-024 | TODO |
| ECC-TRCK-C302 | Corner | Comment POST + PUT issue same version window | comment OK; issue version независимо или 409 documented | API | P1 | TC-TRCK-05 | TODO |
| ECC-TRCK-C303 | Corner | Time-log + title edit race на одном version | оба не теряют данные; version monotonic | API | P1 | TC-TRCK-07, NFR-24 | TODO |
| ECC-TRCK-C304 | Corner | Link create + issue soft-delete mid-request | 404/410; no orphan link row | API | P0 | C202, TC-TRCK-06/12 | TODO |
| ECC-TRCK-C305 | Corner | Triple concurrent PUT (A/B/C) same v | ровно один 200; два 409; final version=v+1 | API | P0 | NFR-24 | TODO |
| ECC-TRCK-C306 | Corner | WS board refresh mid-conflict: B sees A's fields before retry | UI shows conflict banner; no silent clobber | E2E | P1 | WS, C200 | TODO |
| ECC-TRCK-C307 | Corner | Automation rule updates issue while human PUT in-flight | ordered apply или 409; no lost human field | API | P1 | ECC-TRCK-029 | Partial |
| ECC-TRCK-C308 | Corner | Workflow transition (status) + assignee change same PUT body | atomic: both applied or neither; version++ once | API | P0 | C201, TC-TRCK-10 | TODO |
| ECC-TRCK-C309 | Corner | Workflow transition PUT ∥ separate assignee PATCH race | consistent final (status∧assignee) или 409 loser | API | P0 | C201 | TODO |
| ECC-TRCK-C310 | Corner | Transition to Done + assignee=null race | Done without assignee OK if policy; else 400/409 | API | P1 | workflow | TODO |
| ECC-TRCK-C311 | Corner | Invalid transition + valid assignee in one request | 400; assignee unchanged | API | P0 | TC-TRCK-11 | TODO |
| ECC-TRCK-C312 | Corner | Assignee deactivated mid-transition | 400/warn; status not half-applied | API | P1 | ECC-TRCK-028 | TODO |
| ECC-TRCK-C313 | Corner | Required-field gate on transition + concurrent field clear | transition blocked; no status flip | API | P0 | workflow | TODO |
| ECC-TRCK-C314 | Corner | Kanban drag start → WIP fills to limit before drop | drop reject; card snap-back; no ghost column | E2E | P0 | C204, US-M11, ECC-TRCK-018 | TODO |
| ECC-TRCK-C315 | Corner | Two cards drag into WIP=1 column same tick | one OK; one 409/WIP; UI both consistent | E2E/API | P0 | C204, KANB-02 | TODO |
| ECC-TRCK-C316 | Corner | Drag + admin lowers WIP N→M (count>M) mid-flight | reject move; existing stay; clear error | API/E2E | P1 | WIP edit | TODO |
| ECC-TRCK-C317 | Corner | Drag cancel (Esc) after optimistic UI move, server still pending | UI rollback; no double-count WIP | E2E | P1 | FE board | TODO |
| ECC-TRCK-C318 | Corner | Drag across swimlane + filter hide target mid-drag | reject or re-show; no lost issue | E2E | P2 | ECC-TRCK-023 | TODO |
| ECC-TRCK-C319 | Corner | Soft-delete issue → create new with same human key | new KEY-N+1 only; never reuse deleted key | API | P0 | C203, TC-TRCK-12, NFR-11 | TODO |
| ECC-TRCK-C320 | Corner | Soft-delete → restore → create parallel (seq race) | unique keys; restore keeps original key | API | P0 | C203, ECC-TRCK-027 | TODO |
| ECC-TRCK-C321 | Corner | Soft-delete parent epic with open child links | cascade policy or block; no dangling FK | API | P1 | ECC-TRCK-033, C202 | TODO |
| ECC-TRCK-C322 | Corner | Soft-delete then JQL by key / GET by key | 404/empty; no leak of deleted payload | API | P0 | NF-03, TC-TRCK-12 | TODO |
| ECC-TRCK-C323 | Corner | Soft-delete mid-attachment upload (presigned) | upload orphan GC or reject; issue gone | API | P1 | storage, NF-09 | TODO |
| ECC-TRCK-C324 | Corner | Bulk edit 50: 3 stale version, 2 deleted, rest OK | partial report: success/fail per key; no all-or-nothing silent | API | P0 | C206 | TODO |
| ECC-TRCK-C325 | Corner | Bulk assignee change + one target non-member | fail that row; others commit; documented policy | API | P1 | C206, RBAC | TODO |
| ECC-TRCK-C326 | Corner | Bulk transition mixed valid/invalid paths | invalid 400 rows; valid transitioned; version++ only OK | API | P1 | TC-TRCK-10/11 | TODO |
| ECC-TRCK-C327 | Corner | Bulk + concurrent single PUT on overlapping issue | bulk row 409 or single 409; no double version jump | API | P0 | NFR-24, C206 | TODO |
| ECC-TRCK-C328 | Corner | Bulk abort/timeout after N successes | idempotent retry safe; no duplicate side-effects | API | P1 | C206 | TODO |
| ECC-TRCK-C329 | Corner | Saved JQL shared → membership revoked → open share link | 403/empty; zero foreign issues (NF-03) | API | P0 | C205, ECC-TRCK-031, US-C04 | TODO |
| ECC-TRCK-C330 | Corner | Shared JQL + project archived while viewer loads | empty/403; no cached leak | API/E2E | P0 | C205, NF-03 | TODO |
| ECC-TRCK-C331 | Corner | Share token revoke mid-poll refresh of results | subsequent pages 403; first page not re-expanded | API | P0 | C205 | TODO |
| ECC-TRCK-C332 | Corner | Owner edits shared JQL to broader scope after share | ACL re-check; non-members still scoped | API | P0 | NF-03 | TODO |
| ECC-TRCK-C333 | Corner | JQL share to role then role demoted | next run fail-closed | API | P0 | RBAC, C205 | TODO |
| ECC-TRCK-C334 | Corner | Move + soft-delete race (drag drop vs DELETE) | 404/409; board no zombie card | API/E2E | P1 | ECC-TRCK-026 | TODO |
| ECC-TRCK-C335 | Corner | Restore soft-deleted while bulk includes its key | restore wins or bulk skip; documented | API | P1 | NFR-11, C206 | TODO |
| ECC-TRCK-C336 | Corner | Issue in active sprint + soft-delete + sprint complete race | no lost orphan; sprint inventory consistent | API | P0 | C203, SPR complete | TODO |
| ECC-TRCK-C337 | Corner | Custom field edit + workflow transition race | both applied atomically or 409 | API | P1 | ECC-TRCK-030, C201 | TODO |
| ECC-TRCK-C338 | Corner | Subtask create while parent soft-deleting | 404 parent; no orphan subtask | API | P1 | C202 | TODO |
| ECC-TRCK-C339 | Corner | Optimistic UI WIP count + server reject → second drag | count restored before next drag allowed | E2E | P1 | C204 | TODO |

**Tracker Corner в этом файле: 40**

---

## 2. Sprint — `ECC-SPR-C300+`

> Углубление parent C200–C202: complete/start/move races, inventory integrity

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-C300 | Corner | Complete sprint ∥ issue moved to backlog concurrently | issue not lost; not duplicated; inventory = union | API | P0 | C200, TC-SPRINT-07 | TODO |
| ECC-SPR-C301 | Corner | Complete ∥ issue moved into sprint from backlog | policy: include in move-out or stay backlog once; no ghost | API | P0 | C200 | TODO |
| ECC-SPR-C302 | Corner | Complete ∥ workflow transition on incomplete issue | final status + sprint move consistent | API | P1 | ECC-SPR-015, C200 | TODO |
| ECC-SPR-C303 | Corner | Two admins complete same sprint (double-submit) | first 200; second 400/idempotent completed | API | P0 | C201, TC-SPRINT-07 | TODO |
| ECC-SPR-C304 | Corner | Complete + start next sprint same second | at most one active; incomplete routing correct | API | P0 | C201, TC-SPRINT-03 | TODO |
| ECC-SPR-C305 | Corner | Complete while burndown GET in-flight | GET 200 snapshot or completed series; no 500 | API | P2 | TC-SPRINT-04 | TODO |
| ECC-SPR-C306 | Corner | Start sprint A ∥ delete sprint A | terminal: active XOR deleted; no half-state | API | P1 | C202 | TODO |
| ECC-SPR-C307 | Corner | Start A ∥ start B (two planning) race | one active; other 400 | API | P0 | ECC-SPR-014, TC-SPRINT-03 | TODO |
| ECC-SPR-C308 | Corner | Delete planning sprint ∥ add issues to it | 404/410 on add; no orphan sprint_id on issues | API | P1 | C202 | TODO |
| ECC-SPR-C309 | Corner | Complete → move incomplete to next; next deleted mid-move | fallback backlog; no NULL sprint FK break | API | P0 | C200, TC-SPRINT-07 | TODO |
| ECC-SPR-C310 | Corner | Soft-delete issue mid-complete inventory scan | skipped cleanly; counts match remaining | API | P1 | TRCK soft-delete | TODO |
| ECC-SPR-C311 | Corner | Poker reveal ∥ sprint start | votes frozen or cleared per policy; sprint active | API/Manual | P2 | TC-SPRINT-06 | TODO |
| ECC-SPR-C312 | Corner | Capacity change ∥ complete (points recount) | velocity uses committed snapshot; no NaN | API | P1 | TC-SPRINT-05 | TODO |
| ECC-SPR-C313 | Corner | Project archive ∥ active sprint complete | block archive or force-complete first; documented | API | P1 | ECC-SPR-016 | TODO |
| ECC-SPR-C314 | Corner | Reopen issue from completed sprint while velocity recalc | velocity stable or versioned; no race corruption | API | P1 | ECC-TRCK-036 | TODO |
| ECC-SPR-C315 | Corner | Complete with 0 issues ∥ concurrent add-then-remove | completed empty; no stuck active | API | P2 | ECC-SPR-011 | TODO |
| ECC-SPR-C316 | Corner | WS board: complete event while card mid-drag in sprint board | drag cancel/reject; sprint UI → completed | E2E | P1 | C204, C200 | TODO |
| ECC-SPR-C317 | Corner | Complete ∥ bulk edit sprint field on subset | bulk rows fail or excluded; complete inventory OK | API | P1 | C206 bulk | TODO |
| ECC-SPR-C318 | Corner | Two completes with different «move to» targets raced | one target wins; all incompletes same destination | API | P0 | C201 | Future |
| ECC-SPR-C319 | Corner | Start after complete of previous still flushing moves | start waits or 409 until flush done | API | P1 | C200/C201 | TODO |

**Sprint Corner в этом файле: 20**

---

## 3. TMS — `ECC-TMS-C300+`

> Углубление parent C200–C202: Allure mid-delete, run complete∥execute, version rollback mid-edit

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-C300 | Corner | Allure import match ASID + case soft-deleted mid-batch | unmatched/skipped bucket; import completes; no 500 | API | P0 | C200, TC-TMS-07, TC-ALLURE-05 | TODO |
| ECC-TMS-C301 | Corner | Allure import + hard GDPR delete case mid-match | no resurrect; unmatched; audit | API | P0 | C200, US-S03 | TODO |
| ECC-TMS-C302 | Corner | Allure import + case key rename/move mid-match | deterministic rematch or unmatched; no wrong link | API | P1 | TC-ALLURE-05 | TODO |
| ECC-TMS-C303 | Corner | Two Allure imports same suite overlapping cases | idempotent results; stats not double-count | API | P0 | TC-TMS-07 | TODO |
| ECC-TMS-C304 | Corner | Import creates run + admin deletes run mid-parse | fail soft; no orphan results without run | API | P1 | C200 | TODO |
| ECC-TMS-C305 | Corner | Execute result PUT ∥ complete run | execute 400 if completed first OR counted then lock; stats consistent | API | P0 | C201, TC-TMS-06, ECC-TMS-009 | TODO |
| ECC-TMS-C306 | Corner | Two executors same result id (pass vs fail) race | one wins; version/lock; stats match winner | API | P0 | C201, NFR-24 | TODO |
| ECC-TMS-C307 | Corner | Complete run while last execute in-flight | wait or reject complete; final passed/failed counts exact | API | P0 | C201 | TODO |
| ECC-TMS-C308 | Corner | Execute after complete (retry client) | 400; stats unchanged | API | P0 | ECC-TMS-009 | TODO |
| ECC-TMS-C309 | Corner | Add cases to run ∥ complete run | add 400/409; completed set frozen | API | P1 | TC-TMS-05 | TODO |
| ECC-TMS-C310 | Corner | Bug-from-failed ∥ complete run | bug create OK or policy block; no half bug | API | P1 | TC-TMS-09 | TODO |
| ECC-TMS-C311 | Corner | Soft-delete case in active run mid-execute | block execute or skip; run stats exclude cleanly | API | P1 | ECC-TMS-018 | TODO |
| ECC-TMS-C312 | Corner | Rollback case to v1 while editor has v2 form open | save 409/reload; no silent overwrite of rollback | API/E2E | P0 | C202, TC-TMS-03/04 | TODO |
| ECC-TMS-C313 | Corner | Rollback ∥ concurrent PUT new version | one wins; version chain contiguous; no fork | API | P0 | C202, NFR-24, ECC-TMS-016 | TODO |
| ECC-TMS-C314 | Corner | Rollback to vN while Allure import matches current body | match uses post-rollback body or snapshot; documented | API | P1 | C200, C202 | TODO |
| ECC-TMS-C315 | Corner | Rollback while case linked in in-progress run | run keeps executed snapshot XOR warn; no corrupt steps | API | P1 | C202, TC-TMS-06 | TODO |
| ECC-TMS-C316 | Corner | Double rollback (A→v1, B→v2) race | linear history; final body = one version | API | P0 | C202 | TODO |
| ECC-TMS-C317 | Corner | Edit BDD steps + rollback classic history mid-save | type/steps consistent; 409 if conflict | API | P1 | TC-TMS-02/04 | TODO |
| ECC-TMS-C318 | Corner | Export PDF/CSV ∥ complete run mid-stream | export finishes snapshot or 409 retry; no truncated corrupt file | API | P2 | TC-TMS-12, ECC-TMS-019 | Partial |
| ECC-TMS-C319 | Corner | JUnit webhook + complete run race | signature OK path; late results rejected; SEC-02 always | API | P0 | TC-TMS-08, SEC-02 | TODO |
| ECC-TMS-C320 | Corner | Traceability link issue soft-deleted mid-report build | cell empty/404; no leak deleted issue fields | API | P1 | TC-TMS-11, NF-03 | TODO |
| ECC-TMS-C321 | Corner | Flaky report query ∥ mass re-execute same cases | stable aggregation or versioned; no crash | API | P2 | TC-TMS-10 | TODO |
| ECC-TMS-C322 | Corner | Case create key race under import+manual create | unique KEY-TC-N; no collision | API | P0 | ECC-TMS-015 | TODO |
| ECC-TMS-C323 | Corner | Membership revoked mid-run execute | 403; no result written | API | P0 | ECC-TMS-010, RBAC | TODO |
| ECC-TMS-C324 | Corner | Rollback UI open + another user deletes case | 404/gone; no resurrect on save | E2E | P1 | C202, soft-delete | TODO |

**TMS Corner в этом файле: 25**

---

## Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Tracker | ECC-TRCK-C300 … C339 | **40** |
| Sprint | ECC-SPR-C300 … C319 | **20** |
| TMS | ECC-TMS-C300 … C324 | **25** |
| **Итого Corner** | | **85** |

> Parent overlap check: C200–C206 (TRCK), C200–C202 (SPR/TMS) — **не использованы**. Канон `EDGE-CORNER-CATALOG.md` и `docs/TESTING.md` — **не изменялись**. Коммит не создавался.
