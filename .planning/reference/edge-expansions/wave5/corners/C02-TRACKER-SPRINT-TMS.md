# Corner Expansion C02 Wave5 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Corner** (multi-actor races, mid-flight, cross-domain partial fail, schema/index lag)  
> **ID ranges:** `ECC-TRCK-C700+`, `ECC-SPR-C700+`, `ECC-TMS-C700+` — **не пересекают** parent `C200…`, wave1 `C300…`, wave2 `C400…`, wave3 `C500…`, wave4 `C600…`, edge `*100+`, канон `001–040`  
> **Не заменяет** `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`, `00-CORNER-PARENT.md`, wave1–4 `corners/C02-*.md`, `docs/TESTING.md`  
> **Не копировать** сценарии C200–C600 1:1 — только новые комбинации / углубление  
> **Фокус wave5:** cross-project move∩rank∩search-lag · CF-schema migrate∩transition · GH/GL webhook∩trash-purge · timer∩worklog∩bulk · board-config∩multi-board · sprint scope-freeze∩carry-over-chain∩poker-commit · capacity∩deactivate · shared-steps∩data-driven∩env-matrix · AS_ID collision∩BDD convert · version-prune∩run-snapshot  
> **Связи:** US-M03/M11/M12, US-S01, US-M05/S13/C02/C04 · NFR-11/12/24 · NF-03 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*, TC-JQL-*

### Gap-легенда

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated corner-теста / TC |
| **Partial** | Happy/unit есть, race/mid-flight не закрыт |
| **Future** | Поведение ещё не зафиксировано в ADR/политике |

---

## 1. Tracker — `ECC-TRCK-C700+`

> Wave5: cross-project / rank / CF-schema / VCS webhook / trash day30 / timer+worklog / multi-board config

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-C700 | Corner | Cross-project move A→B: Actor1 move; Actor2 rank-reorder в A; Actor3 JQL/search ещё видит issue в A (index lag) | move atomic; rank A без ghost; search ≤bound без leak чужого project; NF-03 | API | P0 | US-M03, NF-03, NFR-24 | TODO |
| ECC-TRCK-C701 | Corner | Cross-project move mid-flight: target project membership revoked у mover | 403; issue остаётся в source; no half-key/prefix rewrite | API | P0 | RBAC, move | TODO |
| ECC-TRCK-C702 | Corner | Cross-project move + soft-delete в source + create same human-seq в target | keys unique per project; deleted source key не reused; target KEY-N корректный | API | P0 | NFR-11, TC-TRCK-12 | TODO |
| ECC-TRCK-C703 | Corner | Board rank: два клиента drag-reorder одной колонки (before/after ids) same tick | один порядок побеждает; нет cycle/duplicate rank; board GET = API | API/E2E | P0 | US-M11, rank | TODO |
| ECC-TRCK-C704 | Corner | Rank PATCH + filter hide card mid-reorder + WS broadcast stale order | UI reconcile; no invisible card with broken neighbors; rank contiguous | E2E/WS | P1 | WS, rank | TODO |
| ECC-TRCK-C705 | Corner | Admin CF-schema: rename/type-change field F while Actor PUT issue с old F payload | 409/400 schema mismatch; issue version не ++ на reject; no corrupt CF JSON | API | P0 | ECC-TRCK-030, NFR-24 | TODO |
| ECC-TRCK-C706 | Corner | CF required-on-transition: schema adds required F while Actor transition без F | transition 400; status unchanged; schema visible на retry | API | P0 | workflow, CF | TODO |
| ECC-TRCK-C707 | Corner | CF multi-select options prune (option removed) ∥ issue edit still sending removed option | reject or coerce documented; no silent keep of deleted option id | API | P1 | CF options | TODO |
| ECC-TRCK-C708 | Corner | Bulk fill CF на 40 issues + schema delete field mid-bulk | partial: OK rows до delete; rest 400; report per-key; no all-or-nothing silent | API | P0 | bulk, CF | TODO |
| ECC-TRCK-C709 | Corner | GitHub/GitLab webhook autolink KEY mid soft-delete issue | webhook SEC path; no resurrect; link skip/404; audit | API | P0 | US-S04/S06, SEC-02 | TODO |
| ECC-TRCK-C710 | Corner | Two VCS webhooks (GH+GL) same commit message two keys + one issue trash-purged day30 | live key linked; purged key no link/leak; no 500 | API | P0 | NFR-11, VCS | TODO |
| ECC-TRCK-C711 | Corner | Trash purge job day30 hard-delete ∥ admin restore ∥ PUT title same ms | hard XOR restore; PUT 404 if purged; restore keeps version chain if wins | API | P0 | NFR-11, trash | TODO |
| ECC-TRCK-C712 | Corner | Trash day29→30 boundary UTC: restore at T−ε, purge at T, comment POST | UTC cutoff deterministic; no comment on purged; NFR-12 | API | P0 | NFR-12, trash | TODO |
| ECC-TRCK-C713 | Corner | Timer start A + timer stop B (other user) + manual worklog C same issue | at most one open timer; worklogs sum consistent; no negative time_spent | API | P0 | TC-TRCK-07/08 | TODO |
| ECC-TRCK-C714 | Corner | Timer running + bulk time adjust + soft-delete issue | timer closed/cancelled; bulk skip deleted; no orphan timer row | API | P1 | timer, bulk | TODO |
| ECC-TRCK-C715 | Corner | Worklog edit (duration) ∥ another deletes that worklog ∥ issue versioned field PUT | worklog 404/409; issue PUT независим или documented; no double-count time | API | P1 | TC-TRCK-07 | TODO |
| ECC-TRCK-C716 | Corner | Multi-board: issue on Board1+Board2; column map diverge; drag on B1 while admin remaps status on B2 | status single source; boards reconcile; no dual status | API/E2E | P0 | US-M11, multi-board | TODO |
| ECC-TRCK-C717 | Corner | Board config: delete column with cards while two drags target that column | delete blocked or cards migrated atomically; drags 409; no lost cards | API/E2E | P0 | board config | TODO |
| ECC-TRCK-C718 | Corner | Board config rename status ∥ workflow transition using old status id | transition fail-closed or alias map; no 500; audit | API | P0 | workflow, board | TODO |
| ECC-TRCK-C719 | Corner | Clone issue mid-flight: source soft-delete + parent re-parent + attachment copy | clone 404/partial policy documented; no clone of deleted; attachments not orphaned in Seaweed without issue | API | P1 | clone, storage | TODO |
| ECC-TRCK-C720 | Corner | Convert type Story→Epic while child Sub-task create + link blocks→Story | type convert XOR child create consistent hierarchy; no illegal Story-parent-of-Epic | API | P0 | hierarchy, types | TODO |
| ECC-TRCK-C721 | Corner | @mention storm (N users) in comment + one mentioned user GDPR erase mid-fanout | notif bounded; erased user no PII in notif/search; comment body redacted/policy | API | P0 | TC-TRCK-05, US-S03, NF-03 | TODO |
| ECC-TRCK-C722 | Corner | Checklist item toggle race (A check, B uncheck, C delete item) same version window | final checklist consistent; no ghost items; version monotonic if versioned | API | P1 | checklist | TODO |
| ECC-TRCK-C723 | Corner | Presigned attachment complete callback after issue soft-delete + project archive kickoff | upload reject/GC; archive not blocked by orphan blob forever; no download of deleted | API | P0 | storage, NF-09, archive | TODO |
| ECC-TRCK-C724 | Corner | Saved filter owner edits JQL broader while shared viewer paginates + project leave | viewer pages fail-closed scoped; no newly-broad leak; NF-03 | API | P0 | TC-JQL, C205-depth≠copy | TODO |
| ECC-TRCK-C725 | Corner | Component/label rename global ∥ issue edit still sends old label id ∥ board filter by label | rename map or 400; filter results eventual consistent; no duplicate label rows | API | P1 | labels | TODO |
| ECC-TRCK-C726 | Corner | Due-date automation fires ∥ human clears due ∥ calendar aggregator read mid-race | final due null XOR set once; calendar no stale event after clear; US-V03 | API | P1 | automation, calendar | Partial |
| ECC-TRCK-C727 | Corner | Imex import issues into project ∥ live create KEY race ∥ soft-delete imported mid-import | unique keys; import resume/idempotent; deleted import rows not resurrected as live | API | P0 | imex, keys | TODO |
| ECC-TRCK-C728 | Corner | Search reindex lag: issue security-level/project membership tightened; stale index still has hit | query path re-checks ACL; zero leak; NF-03 | API | P0 | search, NF-03 | TODO |

**Tracker Corner в этом файле: 29**

---

## 2. Sprint — `ECC-SPR-C700+`

> Wave5: scope-freeze · carry-over chains · poker-commit · capacity∩deactivate · goal/retro · DST burndown

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-C700 | Corner | Active sprint scope-freeze ON: Actor1 add from backlog; Actor2 remove; Actor3 points edit — same window | freeze policy: reject adds/removes XOR allow with audit; points edit documented; scope snapshot for velocity intact | API | P0 | scope, TC-SPRINT-05 | Future |
| ECC-SPR-C701 | Corner | Scope-freeze flip OFF→ON mid bulk «add 20 issues» | partial committed before freeze + rest 409; inventory = committed set; no silent drop | API | P0 | bulk, scope | Future |
| ECC-SPR-C702 | Corner | Carry-over chain: S1 complete→S2; S2 complete→S3 while issue points edited on each hop | issue ровно в одном sprint; points history per sprint snapshot; no lost carry | API | P0 | TC-SPRINT-07, carry-over | TODO |
| ECC-SPR-C703 | Corner | Carry-over issue soft-deleted between S1 complete flush and S2 start | deleted не в S2; trash OK; S2 start inventory clean | API | P0 | NFR-11, TC-SPRINT-07 | TODO |
| ECC-SPR-C704 | Corner | Carry-over epic to S2 while children still flushing from S1 to backlog | hierarchy policy: family together XOR documented split; no orphan child in completed S1 active flag | API | P1 | epic, carry-over | TODO |
| ECC-SPR-C705 | Corner | Poker: three voters submit; reveal; fourth late vote; estimate commit; second reveal race | late vote rejected post-reveal; estimate one value; no double-reveal corruption | API/E2E | P1 | TC-SPRINT-06 | TODO |
| ECC-SPR-C706 | Corner | Poker estimate commit ∥ human PATCH story_points ∥ sprint start | one points source wins; start sees committed points; burndown baseline stable | API | P0 | TC-SPRINT-06/04 | TODO |
| ECC-SPR-C707 | Corner | Poker room WS: voter disconnect mid-vote + reconnect + reveal | vote counted once or clear re-vote; reveal waits quorum policy; no phantom voter | E2E/WS | P1 | WS, poker | TODO |
| ECC-SPR-C708 | Corner | Capacity: member deactivated mid-sprint; capacity recalc; complete with their issues | capacity shrink documented; issues reassigned or stay; velocity not NaN | API | P0 | capacity, RBAC | TODO |
| ECC-SPR-C709 | Corner | Capacity per-person edit ∥ two planners add issues over new cap | reject over-cap or warn+allow documented; no silent overcommit corruption | API | P1 | capacity | Future |
| ECC-SPR-C710 | Corner | Sprint goal PATCH ∥ complete sprint ∥ retro notes create | complete freezes goal snapshot; retro reads completed goal; late goal edit 400 | API | P1 | US-S01, goal | TODO |
| ECC-SPR-C711 | Corner | Retro notes edit ∥ reopen Done issue from completed sprint (velocity touch) | retro intact; velocity versioned; reopen policy не ломает completed snapshot | API | P1 | ECC-TRCK-036, retro | TODO |
| ECC-SPR-C712 | Corner | Burndown GET across DST spring-forward day while points burn mid-day UTC | buckets UTC half-open; no missing/duplicate day; NFR-12 | API | P0 | TC-SPRINT-04, NFR-12 | TODO |
| ECC-SPR-C713 | Corner | Burndown + scope add after freeze-off + complete same evening | series reflects policy (scope change annotation or ignore); complete terminal; no 500 | API | P1 | burndown, scope | TODO |
| ECC-SPR-C714 | Corner | Planning: filter «unestimated» + bulk set points + poker reveal overlapping set | points final consistent; filter refresh; no double-write points | API/E2E | P1 | poker, bulk | TODO |
| ECC-SPR-C715 | Corner | Split issue (clone+close) mid active sprint: points move race with burndown tick | total points conserved per policy; burndown continuous; no duplicate velocity credit | API | P1 | split, velocity | Future |
| ECC-SPR-C716 | Corner | Calendar aggregator shows sprint bar ∥ sprint dates PATCH ∥ complete early | calendar updates or drops bar; no stale active sprint event after complete | API | P1 | US-V03, ECC-SPR-021 | Partial |
| ECC-SPR-C717 | Corner | Two projects' planning boards (same user) drag issue between projects' sprints (if allowed) or reject | cross-project sprint_id fail-closed unless ADR; no FK to foreign sprint | API | P0 | multi-project | Future |
| ECC-SPR-C718 | Corner | Commitment snapshot at start ∥ late admin force-add after start ∥ complete inventory | commitment vs actual distinct; velocity uses documented basis; inventory complete | API | P1 | commitment | Future |

**Sprint Corner в этом файле: 19**

---

## 3. TMS — `ECC-TMS-C700+`

> Wave5: shared-steps · data-driven · env×matrix · AS_ID collision · BDD convert · version-prune · parallel runs

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-C700 | Corner | Shared step S edit body ∥ Case A execute using S ∥ Case B unlink S — same window | execute snapshot XOR live policy documented; unlink не 500 mid-execute; A/B consistent | API | P0 | shared steps | TODO |
| ECC-TMS-C701 | Corner | Soft-delete shared step S while N cases reference S + Allure import matching those cases | cases fail-closed or inline copy; import unmatched/skip; no broken step pointer | API | P0 | shared steps, Allure | TODO |
| ECC-TMS-C702 | Corner | Data-driven: two executors run different rows of same case in one run | row results isolated; stats = sum rows; no cross-row overwrite | API | P0 | data-driven, TC-TMS-06 | TODO |
| ECC-TMS-C703 | Corner | Data-driven row set replace (CSV) mid-run execute of old row ids | old row execute 404/409; new rows not auto-added to in-flight run unless policy; stats clean | API | P0 | data-driven | TODO |
| ECC-TMS-C704 | Corner | Env×case matrix: flip env staging→prod on case while run R bound to staging executes | execute uses run-bound env snapshot; flip не меняет in-flight results; matrix report eventual | API | P0 | env matrix | TODO |
| ECC-TMS-C705 | Corner | Env delete while active runs reference env + matrix rebuild | delete blocked or runs freeze env name; matrix no null env crash | API | P1 | env | TODO |
| ECC-TMS-C706 | Corner | AS_ID collision: two cases claim same Allure id; import batch + manual AS_ID PATCH | deterministic match one or unmatched both; no dual-link; audit conflict | API | P0 | TC-ALLURE-05, AS_ID | TODO |
| ECC-TMS-C707 | Corner | AS_ID clear on case X while Allure rematch storm for X | unmatched after clear; no result attach to wrong case; idempotent import | API | P0 | Allure | TODO |
| ECC-TMS-C708 | Corner | BDD↔classic convert mid: editor save BDD + executor classic steps + version++ | convert atomic; execute 409/reload if type changed; version chain linear | API | P0 | TC-TMS-02/03, NFR-24 | TODO |
| ECC-TMS-C709 | Corner | Convert BDD→classic ∥ shared-step link ∥ data-driven rows present | reject convert or drop incompatible links with explicit errors; no half-type body | API | P1 | convert, shared | TODO |
| ECC-TMS-C710 | Corner | Version prune job drops v1..vK while Actor rollback to v1 + run holds snapshot of vK+2 | rollback 404 if pruned; run snapshot intact; no fork history | API | P0 | TC-TMS-04, prune | Future |
| ECC-TMS-C711 | Corner | Version prune ∥ concurrent PUT new version ∥ export case PDF | prune skips in-use; PUT contiguous version; PDF = one frozen body | API | P1 | prune, export | Future |
| ECC-TMS-C712 | Corner | Parallel runs R1/R2 same suite: complete R1 ∥ execute in R2 ∥ suite reorder cases | R1 frozen; R2 independent; suite reorder не меняет completed R1 set | API | P0 | TC-TMS-05/06 | TODO |
| ECC-TMS-C713 | Corner | Suite folder move (ACL change) mid execute + add-case-to-run | 403 if ACL lost; run membership frozen or reject; no IDOR case body | API | P0 | RBAC, suite | TODO |
| ECC-TMS-C714 | Corner | Flaky quarantine flag ON ∥ CI webhook re-execute ∥ human override pass | quarantine policy: auto skip or count separate; stats not double; SEC-02 on webhook | API | P1 | TC-TMS-10, SEC-02 | TODO |
| ECC-TMS-C715 | Corner | Result attachment upload (presigned) ∥ complete run ∥ case soft-delete | attachment GC or bound to result snapshot; complete stats OK; no download via deleted case alone | API | P1 | storage, TC-TMS-06/07 | TODO |
| ECC-TMS-C716 | Corner | Bug-from-fail creates issue ∥ tracker soft-delete that issue ∥ re-fail same case creates second bug | policy: reopen XOR new bug; TMS link not 500; NF-03 on deleted bug | API | P1 | TC-TMS-09, TRCK | TODO |
| ECC-TMS-C717 | Corner | Traceability matrix rebuild job ∥ mass unlink case↔issue ∥ JQL on linked issues | matrix consistent snapshot; unlinked cells empty; JQL no stale link leak | API | P1 | TC-TMS-11, NF-03 | TODO |
| ECC-TMS-C718 | Corner | Case clone storm (10 parallel) + key allocator + AS_ID copy policy | unique KEY-TC-N; AS_ID copied or cleared per policy; no collision | API | P0 | ECC-TMS-015-depth≠copy | TODO |
| ECC-TMS-C719 | Corner | Mongo result write OK + PG run stats update fail (tx boundary) mid execute | fail-closed or outbox retry; stats eventually match results; no silent drift | API | P0 | NFR-24, consistency | TODO |
| ECC-TMS-C720 | Corner | Plan/milestone (if present): add run to plan ∥ plan archive ∥ execute | archive blocks execute or run detaches; documented; no 500 | API | P2 | plan | Future |

**TMS Corner в этом файле: 21**

---

## Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Tracker | ECC-TRCK-C700 … C728 | **29** |
| Sprint | ECC-SPR-C700 … C718 | **19** |
| TMS | ECC-TMS-C700 … C720 | **21** |
| **Итого Corner** | | **69** |

### Фокус-покрытие (wave5)

| Фокус | IDs (примеры) |
|-------|----------------|
| Cross-project move ∩ rank ∩ search-lag | ECC-TRCK-C700…C704, C728 |
| CF-schema migrate ∩ transition/bulk | ECC-TRCK-C705…C708 |
| GH/GL webhook ∩ trash day30 purge | ECC-TRCK-C709…C712, C723 |
| Timer ∩ worklog ∩ bulk | ECC-TRCK-C713…C715 |
| Multi-board / board-config mid-drag | ECC-TRCK-C716…C718 |
| Type convert / clone / mention GDPR | ECC-TRCK-C719…C722, C727 |
| Sprint scope-freeze ∩ carry-over chain | ECC-SPR-C700…C704 |
| Poker-commit ∩ capacity∩deactivate | ECC-SPR-C705…C709 |
| Goal/retro/DST burndown | ECC-SPR-C710…C718 |
| Shared-steps ∩ data-driven ∩ env-matrix | ECC-TMS-C700…C705 |
| AS_ID collision ∩ BDD convert | ECC-TMS-C706…C709 |
| Version-prune ∩ parallel runs ∩ consistency | ECC-TMS-C710…C720 |

> Overlap check: `C200–C206` / `C300–C339` / `C400–C430` (TRCK), `C200–C202` / `C300–C319` / `C400–C417` (SPR), `C200–C202` / `C300–C324` / `C400–C419` (TMS), wave3/4 `C500–C6xx` — **не использованы как копии**. Канон `EDGE-CORNER-CATALOG.md` — **не изменялся**. Коммит не создавался.
