# Corner Expansion C02 Wave4 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Corner** (multi-actor races, mid-flight, triple-path, partial fail, taxonomy/VCS/rank)  
> **ID ranges:** `ECC-TRCK-C600+`, `ECC-SPR-C600+`, `ECC-TMS-C600+` — **не пересекают** parent `C200…`, wave1 `C300…`, wave2 `C400…`, wave3 `C500…`, edge `*100+`, канон `001–040`  
> **Не заменяет** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`, `00-CORNER-PARENT.md`, wave1/2 `corners/C02-*.md`, `docs/TESTING.md`  
> **Не копировать** сценарии C200–C500 1:1 — только новые комбинации / углубление  
> **Фокус wave4:** LexoRank/backlog reorder races · cross-project move+links+WIP · clone/IMEX mid-flight · VCS webhook∥human∥delete · component/label rename mid-bulk · issue-type change mid-gate · Gantt/blocks cascade · sprint poker+capacity+start · goal/dates mid-active · shared-steps mid-run · suite reparent mid-execute · data-driven params · evidence mid-delete · multi-CI webhook  
> **Связи:** US-M03/M11/M12, US-S01, US-M05/S13/C02/C04 · NFR-11/12/24 · NF-03 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*, TC-JQL-*

### Gap-легенда

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated corner-теста / TC |
| **Partial** | Happy/unit есть, race/mid-flight не закрыт |
| **Future** | Поведение ещё не зафиксировано в ADR/политике |

---

## 1. Tracker — `ECC-TRCK-C600+`

> Wave4: LexoRank/backlog · cross-project move · clone/IMEX · VCS webhook · taxonomy rename · type-change mid-gate · blocks/Gantt cascade · timer∥sprint

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-C600 | Corner | Три актора одновременно PATCH rank (LexoRank) одной issue между двумя соседями с разными before/after | ровно один порядок побеждает; нет duplicate rank; список backlog детерминирован после re-GET | API | P0 | backlog rank, NFR-24 | TODO |
| ECC-TRCK-C601 | Corner | Actor1 rank-up issue A; Actor2 soft-delete соседа B (якорь before); Actor3 filter hide C mid-reorder | rank API fail-closed или re-anchor; нет «дыры»/цикла; deleted не в порядке | API/E2E | P0 | NFR-11, rank | TODO |
| ECC-TRCK-C602 | Corner | Bulk rank 40 issues + single drag overlapping key + JQL saved-filter refresh | bulk и drag не дают split order; filter видит тот же порядок, что API | API/E2E | P1 | bulk, TC-JQL-01 | TODO |
| ECC-TRCK-C603 | Corner | Rank PATCH + cross-board swimlane filter меняет visible set mid-drag | drop reject или re-target; карточка не исчезает из обоих списков | E2E | P1 | ECC-TRCK-023-depth | TODO |
| ECC-TRCK-C604 | Corner | Cross-project move issue P1→P2: Actor1 move; Actor2 создаёт link blocks→issue в P1; Actor3 WIP-drag на board P1 | move atomic XOR reject; link не orphan на старый project; WIP P1 без ghost | API/E2E | P0 | cross-project, TC-TRCK-06, US-M11 | TODO |
| ECC-TRCK-C605 | Corner | Cross-project move mid-flight: issue в active sprint P1; P2 не имеет sprint; параллельно complete sprint P1 | sprint_id cleared/fallback documented; move не оставляет FK на completed sprint чужого project | API | P0 | TC-SPRINT-07 ∩ move | TODO |
| ECC-TRCK-C606 | Corner | Cross-project move + component/label IDs из P1 + watcher из P1-only member | taxonomy remap или strip; watcher без membership P2 снят; 403 не на половину полей | API | P0 | RBAC, taxonomy | TODO |
| ECC-TRCK-C607 | Corner | Cross-project move + attachment presigned PUT ещё uploading в bucket P1 path | upload GC/reject; move не 500; нет файла «между» projects без ACL | API | P1 | storage, NF-09 | TODO |
| ECC-TRCK-C608 | Corner | Clone issue A→A' while Actor2 PUT versioned fields A and Actor3 soft-deletes A | clone snapshot = documented generation; delete не orphan'ит A'; A' не получает deleted key | API | P0 | clone, NFR-11/24 | TODO |
| ECC-TRCK-C609 | Corner | IMEX import CSV 200 rows ∥ human create same titles ∥ bulk soft-delete overlapping keys mid-import | unique keys; partial report; no resurrect deleted; import не 500 | API | P0 | imex, C206-depth≠copy | TODO |
| ECC-TRCK-C610 | Corner | IMEX export stream mid: rank reorder + cross-project move subset | export snapshot или 409 retry; CSV не truncated; keys consistent with frozen view | API | P2 | imex export | Partial |
| ECC-TRCK-C611 | Corner | GitHub webhook issue_comment + human PUT description + soft-delete same issue — same window | SEC signature always; delete wins terminal; webhook не resurrect; version monotonic если ещё alive | API | P0 | github, NFR-24, SEC | TODO |
| ECC-TRCK-C612 | Corner | GitLab pipeline failed → auto-transition + human WIP-drag opposite + assignee deactivate | automation и human не дают ghost column; assignee gate fail-closed; WIP count exact | API | P0 | gitlab, WIP, workflow | TODO |
| ECC-TRCK-C613 | Corner | Два VCS webhooks (GH+GL) на один commit-link + human unlink mid-flight | at most one link row per SHA policy; unlink не 500; audit | API | P1 | github/gitlab links | TODO |
| ECC-TRCK-C614 | Corner | Admin renames component CompA→CompB while bulk edit sets component=CompA on 30 issues + JQL `component = CompA` runs | bulk remap или fail rows; JQL не leak старый id; search NF-03 scoped | API | P0 | taxonomy, NF-03, TC-JQL-01 | TODO |
| ECC-TRCK-C615 | Corner | Label merge (bug+defect→bug) ∥ triple-tab edit labels на issue ∥ saved filter by old label | final labels set без дублей; filter empty или remapped — documented; no 500 | API/E2E | P1 | labels, JQL | TODO |
| ECC-TRCK-C616 | Corner | Issue type Story→Bug mid: required custom-field gate + workflow transition + rank on backlog | type change atomic с fields или 400; transition не half-applies; rank валиден для нового type board | API | P0 | type change, workflow | TODO |
| ECC-TRCK-C617 | Corner | Type change Epic→Story while children still linked as epic-children + third actor adds sub-task | hierarchy policy reject или cascade retype; no illegal epic-child after; FK intact | API | P0 | epic hierarchy | TODO |
| ECC-TRCK-C618 | Corner | Blocks-link cascade: A blocks B blocks C; Actor1 transition A→Done; Actor2 soft-delete B; Actor3 Gantt drag C dates | cascade stops at deleted B; C dates consistent; no NPE/500; board/Gantt reconcile | API/E2E | P1 | TC-TRCK-06, Gantt Ph16 | Future |
| ECC-TRCK-C619 | Corner | Circular blocks attempt (A↔B) while both issues cross-project moving opposite directions | both reject cycle; moves fail-closed or serialize; no cycle row across projects | API | P0 | TC-TRCK-06 | TODO |
| ECC-TRCK-C620 | Corner | Timer running on issue + soft-delete + sprint complete move-out same tick | timer stop/discard documented; time_spent not double; issue not in completed sprint inventory if deleted | API | P0 | TC-TRCK-08/12, TC-SPRINT-07 | TODO |
| ECC-TRCK-C621 | Corner | Manual time-log POST + estimate PATCH + clone mid-log | clone time policy documented (copy/zero); estimate version lock; no negative remaining | API | P1 | TC-TRCK-07, clone | TODO |
| ECC-TRCK-C622 | Corner | @mention in comment + mentioned user membership revoked + issue soft-delete before notif flush | notif not delivered or fail-closed; no leak issue body to revoked; NF-03 | API | P0 | TC-TRCK-05, NF-03 | TODO |
| ECC-TRCK-C623 | Corner | Board column reorder (admin) ∥ two WIP-drags into old column ids ∥ workflow status map remap | drops reject or remap; WIP counts on new columns; no card in deleted column id | API/E2E | P0 | US-M11, board config | TODO |
| ECC-TRCK-C624 | Corner | Saved JQL uses `rank > X` + LexoRank compaction job mid-page cursor | cursor stable или 409 restart; no skip/dup issues across pages | API | P1 | JQL, rank, pagination | TODO |
| ECC-TRCK-C625 | Corner | Watcher add/remove storm (10 actors) ∥ versioned title PUT ∥ WS board subscribers | watcher side-effect не ломает version monotonicity title; WS не clobber form | API/WS | P1 | watchers, NFR-24 | TODO |

**Tracker Corner в этом файле: 26**

---

## 2. Sprint — `ECC-SPR-C600+`

> Wave4: poker+capacity+start triple · goal/dates mid-active · release/version link · multi-team capacity · reopen/velocity · rank-into-sprint

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-C600 | Corner | Planning poker: 5 voters reveal; admin меняет capacity; третий start sprint — same window | votes frozen/cleared per policy; capacity snapshot for start documented; at most one active | API | P0 | TC-SPRINT-06/03/05 | TODO |
| ECC-SPR-C601 | Corner | Two poker sessions on same issue (tabs) + points PATCH from reveal + move issue out of sprint | points не пишутся в issue вне sprint; session closed; no ghost votes | API/E2E | P1 | TC-SPRINT-06 | TODO |
| ECC-SPR-C602 | Corner | Capacity edit (team A) ∥ capacity edit (team B multi-team) ∥ start — overlapping | start видит consistent capacity sum; reject если overcommit policy; no NaN velocity base | API | P1 | capacity, Future | Future |
| ECC-SPR-C603 | Corner | Sprint goal PATCH + date range shrink ∥ burndown GET ∥ issue due-date outside new range | goal/dates atomic; burndown series rebuild или 409; issues не silently dropped | API | P0 | TC-SPRINT-04, goal | TODO |
| ECC-SPR-C604 | Corner | Active sprint: end_date moved to past mid-day ∥ complete race ∥ add issue | add 400; complete deterministic; burndown last point stable | API | P1 | TC-SPRINT-04/07 | TODO |
| ECC-SPR-C605 | Corner | Link sprint to release/version V; V archived mid; complete sprint move incompletes | incompletes routing OK; release link nullify/block documented; velocity не ссылает archived V leak | API | P1 | release, TC-SPRINT-05 | Future |
| ECC-SPR-C606 | Corner | Rank backlog → drag into active sprint board (rank+sprint_id) ∥ complete inventory scan | issue counted once; rank in sprint list valid; no lost between backlog and sprint | API/E2E | P0 | rank ∩ TC-SPRINT-07 | TODO |
| ECC-SPR-C607 | Corner | Three planning sprints: bulk move 20 issues S1→S2→S3 cascading while S2 deleted | no NULL sprint_id; survivors in S1/S3/backlog; deleted S2 empty refs | API | P0 | TC-SPRINT-07-depth | TODO |
| ECC-SPR-C608 | Corner | Reopen completed sprint (if allowed) ∥ velocity recalc ∥ new complete of later sprint | velocity history versioned; no double-count points; reopen 403 if Future-disabled | API | P1 | TC-SPRINT-05, ECC-TRCK-036-depth | Future |
| ECC-SPR-C609 | Corner | Start sprint while IMEX imports issues with sprint_id=S and poker reveal sets points | import rows + points + active flag consistent; partial import report; no second active | API | P1 | imex ∩ sprint | TODO |
| ECC-SPR-C610 | Corner | Cross-project move of issue out of active sprint P1 into P2 ∥ sprint complete P1 | inventory excludes moved; P2 backlog/sprint policy; no duplicate velocity credit | API | P0 | TRCK-C605∩SPR | TODO |
| ECC-SPR-C611 | Corner | WS: capacity toast + poker reveal event + start event — three clients different order | UI converge to one active; no double start button success; cards match API | E2E/WS | P1 | WS, TC-SPRINT-03/06 | TODO |
| ECC-SPR-C612 | Corner | Soft-delete 10% of sprint issues mid-poker + mid-capacity recalc + start | deleted excluded from commitment; start counts match live issues; trash OK | API | P0 | NFR-11, TC-SPRINT-03 | TODO |
| ECC-SPR-C613 | Corner | Sprint rename/key slug change ∥ saved JQL `sprint = OldName` ∥ burndown bookmark | JQL remaps or empty fail-closed; burndown URL 301/409; NF-03 | API | P1 | TC-JQL-01, NF-03 | TODO |
| ECC-SPR-C614 | Corner | Timer still running on incomplete ∥ complete→Next ∥ points edit on Next | timer bound to issue not sprint; Next points OK; velocity S snapshot ignores post-complete edits | API | P1 | TC-TRCK-08, TC-SPRINT-05/07 | TODO |
| ECC-SPR-C615 | Corner | Two admins: A sets destination=Next on complete; B deletes Next; C creates Next2 mid-flush | incompletes → Next2 XOR backlog fallback; no FK break; one complete terminal | API | P0 | TC-SPRINT-07 | TODO |
| ECC-SPR-C616 | Corner | Commitment chart GET ∥ live add/remove 50 issues ∥ capacity change — cursor pagination | pages without dup/skip; commitment number matches final inventory or frozen snapshot header | API | P2 | capacity, pagination | TODO |
| ECC-SPR-C617 | Corner | Sub-tasks in sprint, parent in backlog: complete moves parent policy ∥ rank children ∥ WIP parent board | family sprint fields per hierarchy policy; no half-family undocumented; WIP parent OK | API | P1 | hierarchy ∩ sprint | TODO |

**Sprint Corner в этом файле: 18**

---

## 3. TMS — `ECC-TMS-C600+`

> Wave4: shared-steps mid-run · suite reparent mid-execute · data-driven params · evidence mid-delete · clone-run mid-complete · env switch · review gate · multi-CI webhook

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-C600 | Corner | Shared-steps library S edited (steps reorder) while 3 runs execute cases embedding S | in-flight results use snapshot XOR 409 reload; library edit не corruption чужих results; stats exact | API | P0 | shared-steps, TC-TMS-06 | Future |
| ECC-TMS-C601 | Corner | Soft-delete shared-step S ∥ case PUT still references S ∥ execute in run | execute reject or inline snapshot; case save 409; no FK 500 | API | P0 | shared-steps, NFR-11 | Future |
| ECC-TMS-C602 | Corner | Suite tree: reparent folder F1→F2 while Allure match by suite path + execute case in F1 + move case to F3 | match deterministic path; execute не пишет в moved-away slot; tree consistent | API | P0 | suite, TC-ALLURE-05 | TODO |
| ECC-TMS-C603 | Corner | Suite delete (non-empty) ∥ bulk add cases from suite to run ∥ rank cases in run | delete block or cascade policy; run add partial report; no orphan suite_id | API | P1 | suite, TC-TMS-05 | TODO |
| ECC-TMS-C604 | Corner | Data-driven case: Actor1 edits param set; Actor2 executes row#3; Actor3 deletes row#3 mid-PUT result | result bound to param generation; deleted row → 404; stats no phantom row | API | P0 | data-driven, NFR-24 | Future |
| ECC-TMS-C605 | Corner | Two executors different param rows same case + Allure import flat result without params | mapping policy documented; no cross-row overwrite; counts exact | API | P1 | Allure ∩ params | Future |
| ECC-TMS-C606 | Corner | Evidence attachment presigned upload mid: soft-delete case + complete run + execute retry | upload GC; complete freezes without evidence XOR attach-to-snapshot; retry 400 after complete | API | P0 | storage, TC-TMS-06/07 | TODO |
| ECC-TMS-C607 | Corner | Evidence upload 25MB boundary + run delete mid-PUT + case still live | BodyLimit/presigned policy; no orphan ACL leak; run gone → 404 evidence API | API | P1 | NF-09, storage | TODO |
| ECC-TMS-C608 | Corner | Clone run R→R' while executes still in-flight on R + admin completes R | R' copies frozen case set without in-flight dirty results OR waits; R stats final; no cross-run bleed | API | P0 | clone-run, TC-TMS-05/06 | TODO |
| ECC-TMS-C609 | Corner | Clone run + soft-delete half source cases mid-clone + Allure targets R' | clone skips deleted; Allure unmatched for gone; R' without zombie slots | API | P0 | clone-run, TC-TMS-07 | TODO |
| ECC-TMS-C610 | Corner | Environment switch (env A→B) on active run ∥ execute ∥ JUnit webhook labeled env A | results tagged with env at write-time; late A rejected or quarantined; stats per env consistent | API | P1 | env, TC-TMS-08 | Future |
| ECC-TMS-C611 | Corner | Review/approval gate on case: approve∥reject race + execute before approve + rollback version | execute blocked until approved XOR policy; rollback invalidates approval; audit | API | P1 | review gate | Future |
| ECC-TMS-C612 | Corner | Multi-CI: GitHub Actions JUnit + GitLab Allure + manual execute — same case X same second | SEC-02 all webhooks; one result winner or per-source rows documented; stats not triple-count | API | P0 | TC-TMS-07/08, SEC-02 | TODO |
| ECC-TMS-C613 | Corner | Webhook signature valid but timestamp skew + case rename ASID mid + run complete | skew reject policy; rename unmatched; complete frozen; no unsigned bypass | API | P0 | SEC-02, TC-ALLURE-05 | TODO |
| ECC-TMS-C614 | Corner | Traceability: link case↔issue; issue cross-project move; case suite reparent; matrix rebuild mid | matrix cells follow live ACL; NF-03 no foreign project leak; 404 cells clean | API | P0 | TC-TMS-11, NF-03 | TODO |
| ECC-TMS-C615 | Corner | Flaky detector job ∥ mass re-execute ∥ delete case ∥ export flaky CSV | export snapshot; deleted excluded; no crash/NaN flaky score | API | P2 | TC-TMS-10/12 | TODO |
| ECC-TMS-C616 | Corner | BDD↔classic type convert mid: rollback to other type + execute + shared-step attach | type/steps consistent; execute sees one type; 409 on conflict; no hybrid body | API | P1 | TC-TMS-02/04 | TODO |
| ECC-TMS-C617 | Corner | Membership revoked mid shared-step edit + suite admin move + execute | 403 on writes; no partial suite move; execute not attributed to revoked user | API | P0 | RBAC, ECC-TMS-010-depth | TODO |
| ECC-TMS-C618 | Corner | Run export PDF stream ∥ clone run ∥ env switch ∥ last execute | export intact snapshot; clone/export isolation; no truncated PDF | API | P2 | TC-TMS-12 | Partial |
| ECC-TMS-C619 | Corner | Bug-from-fail create while issue-type scheme blocks Bug in project + Allure reimport fail + case delete | bug create 400 policy; case delete OK; import stats; no orphan bug draft | API | P1 | TC-TMS-09/07 | TODO |

**TMS Corner в этом файле: 20**

---

## Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Tracker | ECC-TRCK-C600 … C625 | **26** |
| Sprint | ECC-SPR-C600 … C617 | **18** |
| TMS | ECC-TMS-C600 … C619 | **20** |
| **Итого Corner** | | **64** |

### Фокус-покрытие (wave4)

| Фокус | IDs (примеры) |
|-------|----------------|
| LexoRank / backlog reorder | ECC-TRCK-C600…C603, C624; ECC-SPR-C606 |
| Cross-project move + links + WIP | ECC-TRCK-C604…C607, C619; ECC-SPR-C610 |
| Clone / IMEX mid-flight | ECC-TRCK-C608…C610; ECC-SPR-C609 |
| VCS webhook ∥ human ∥ delete | ECC-TRCK-C611…C613 |
| Component/label rename mid-bulk | ECC-TRCK-C614…C615 |
| Issue-type change mid-gate | ECC-TRCK-C616…C617 |
| Blocks / Gantt cascade | ECC-TRCK-C618…C619 |
| Timer ∥ sprint ∥ mention ACL | ECC-TRCK-C620…C622; ECC-SPR-C614 |
| Board column remap | ECC-TRCK-C623 |
| Poker + capacity + start | ECC-SPR-C600…C602, C611–C612 |
| Goal/dates mid-active | ECC-SPR-C603…C604 |
| Release / multi-sprint cascade | ECC-SPR-C605, C607–C608, C615 |
| Shared-steps mid-run | ECC-TMS-C600…C601, C616 |
| Suite reparent mid-execute | ECC-TMS-C602…C603 |
| Data-driven params | ECC-TMS-C604…C605 |
| Evidence mid-delete | ECC-TMS-C606…C607 |
| Clone-run mid-complete | ECC-TMS-C608…C609 |
| Env / review / multi-CI | ECC-TMS-C610…C613, C617 |
| Traceability / flaky / bug gate | ECC-TMS-C614…C615, C619 |

> Overlap check: `C200–C206` / `C300–C339` / `C400–C430` (TRCK), `C200–C202` / `C300–C319` / `C400–C417` (SPR), `C200–C202` / `C300–C324` / `C400–C419` (TMS), wave3 `C500…` — **не использованы как копии**. Канон `EDGE-CORNER-CATALOG.md` — **не изменялся**. Коммит не создавался.
