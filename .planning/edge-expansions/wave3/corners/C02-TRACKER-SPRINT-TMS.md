# Corner Expansion C02 Wave3 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Corner** (multi-actor races, mid-flight, partial fail, clock skew, dual-path)  
> **ID ranges:** `ECC-TRCK-C500+`, `ECC-SPR-C500+`, `ECC-TMS-C500+` — **не пересекают** parent `C200…`, wave1 `C300…`, wave2 `C400…`, edge `*100+`/`*200+`/`*300+`, канон `001–040`  
> **Не заменяет** `.planning/EDGE-CORNER-CATALOG.md`, `00-CORNER-PARENT.md`, wave1/2 `corners/C02-*.md`, `docs/TESTING.md`  
> **Не копировать** сценарии C200/C300/C400 1:1 — только новые комбинации  
> **Фокус wave3:** clone+imex+webhook mid-flight · attachment/presigned∩soft-delete · workflow-schema edit mid-transition · cross-project move · poker∩start/complete · burndown UTC skew · BDD↔classic mid-run · suite-tree move∩Allure · shared-steps∩execute · dual JUnit+Allure  
> **Связи:** US-M03/M11/M12, US-S01/S04/S06, US-M05/S13/C02/C04 · NFR-11/12/24 · NF-03/09 · SEC-02 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*, TC-JQL-*

### Gap-легенда

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated corner-теста / TC |
| **Partial** | Happy/unit есть, race/mid-flight не закрыт |
| **Future** | Поведение ещё не зафиксировано в ADR/политике |
| **Manual** | Только ручной / exploratory E2E |

---

## 1. Tracker — `ECC-TRCK-C500+`

> Wave3: clone∥imex∥GH/GL webhook · attachment/presigned∩delete · workflow schema mid-transition · cross-project move · timer/time-log∩version · component/label rename mid-bulk

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-C500 | Corner | Clone issue I (deep: links+attachments meta) ∥ CSV imex import overlapping key prefix ∥ GitHub webhook autolink на исходный I | clone получает новый KEY; import не коллизит с clone key; webhook link только на живой I; нет dual-key / orphan attachment meta | API | P0 | US-S04, clone, imex | TODO |
| ECC-TRCK-C501 | Corner | Clone mid-flight: source soft-delete до commit clone | clone 404/410 или snapshot-then-OK documented; нет half-clone без source audit | API | P0 | TC-TRCK-12, clone | TODO |
| ECC-TRCK-C502 | Corner | Clone + GitLab MR webhook + human PUT title на source — same window | clone body = snapshot policy; source version lock независим; webhook не пишет в clone | API | P1 | US-S06, NFR-24 | TODO |
| ECC-TRCK-C503 | Corner | Imex export stream (N=500) ∥ bulk edit overlapping keys ∥ single soft-delete mid-export | export snapshot или 409 retry; файл не truncated; deleted keys absent/marked; bulk не ломает export cursor | API | P1 | imex, C206-depth≠ | TODO |
| ECC-TRCK-C504 | Corner | Imex import create + concurrent create same title/type (no idempotency key) | unique KEY-N; no duplicate rows; import report lists created keys | API | P0 | ECC-TRCK-022 depth | TODO |
| ECC-TRCK-C505 | Corner | GitHub push comment `fixes KEY-1` ∥ issue KEY-1 soft-delete ∥ second webhook retry | first link skip/404; retry idempotent; no resurrect; SEC audit | API | P0 | US-S04, TC-TRCK-12 | TODO |
| ECC-TRCK-C506 | Corner | GitLab + GitHub dual webhook same commit message mentioning KEY-1 | at most one link per provider or deduped; no duplicate blocks link rows | API | P1 | US-S04/S06 | TODO |
| ECC-TRCK-C507 | Corner | Presigned PUT attachment start → soft-delete issue → client completes PUT → register attachment API | register 404/410; object GC/orphan job; issue trash без attachment row | API | P0 | NF-09, NFR-11, storage | TODO |
| ECC-TRCK-C508 | Corner | Two clients: A uploads attachment, B deletes attachment meta, A confirms register | one wins; no dangling Seaweed key without meta XOR meta without object | API | P0 | storage | TODO |
| ECC-TRCK-C509 | Corner | Attachment download URL issued ∥ GDPR hard-delete issue mid-download | subsequent bytes fail-closed / 404; no content leak after hard-delete; audit | API | P0 | US-S03, NF-03 | TODO |
| ECC-TRCK-C510 | Corner | Comment with @mention + attachment upload + soft-delete issue mid both | comment/notif not delivered on deleted; upload orphan handled; no 500 | API | P1 | TC-TRCK-05/12 | TODO |
| ECC-TRCK-C511 | Corner | Admin edits workflow: removes status S while Actor transitions issue → S | transition 400/migrate; no issue stuck in removed status; board columns reconcile | API | P0 | ECC-TRCK-039 depth, TC-TRCK-10/11 | TODO |
| ECC-TRCK-C512 | Corner | Workflow add required custom field on transition ∥ Actor clears field ∥ second Actor transitions | gate blocks; status unchanged; version monotonic | API | P0 | workflow gate, ECC-TRCK-030 | TODO |
| ECC-TRCK-C513 | Corner | Workflow graph replace (import JSON) ∥ Kanban drag using old column map | drag reject or remap; no ghost column id; WIP counts on new graph | API/E2E | P0 | US-M11, workflow | TODO |
| ECC-TRCK-C514 | Corner | Two admins: A renames status, B deletes same status, C transitions into it | fail-closed; final graph consistent; issues not in deleted status | API | P0 | workflow race | TODO |
| ECC-TRCK-C515 | Corner | Cross-project move KEY-A → Project B ∥ watcher/comment/time-log mid-move ∥ board filter on A | move atomic; post-move A board empty of issue; B has issue; side-effects target B or reject mid-move | API | P0 | cross-project | Future |
| ECC-TRCK-C516 | Corner | Cross-project move ∥ epic parent in source project ∥ child link create | parent link policy: drop/block/re-home documented; no cross-project orphan FK | API | P1 | epic, move | Future |
| ECC-TRCK-C517 | Corner | Timer start on I ∥ soft-delete I ∥ timer stop | stop 404/410; no time_spent on deleted; no zombie timer Redis key | API | P1 | TC-TRCK-08/12 | TODO |
| ECC-TRCK-C518 | Corner | Timer running ∥ manual time-log POST ∥ PUT title version race | time_spent consistent; version lock on fields; timer not double-count | API | P1 | TC-TRCK-07/08, NFR-24 | TODO |
| ECC-TRCK-C519 | Corner | Component rename C→C2 ∥ bulk set component=C ∥ JQL `component = C` mid-rename | bulk rows resolve by id or fail; JQL no leak wrong set; NF-03 scope intact | API | P1 | JQL, bulk | TODO |
| ECC-TRCK-C520 | Corner | Label merge (L1+L2→L3) ∥ issue edit adds L1 ∥ saved JQL share uses L2 | post-merge: labels canonical; share results scoped; no phantom L1/L2 | API | P1 | labels, C205-depth≠ | TODO |
| ECC-TRCK-C521 | Corner | Checklist item toggle ∥ issue soft-delete ∥ restore then toggle again | delete clears in-flight; after restore version continues; no checklist ghost | API | P2 | NFR-11 | TODO |
| ECC-TRCK-C522 | Corner | Saved board filter + swimlane by component ∥ component deleted mid-render ∥ drag | UI empty lane or reject; no lost card; API board consistent | E2E | P1 | ECC-TRCK-023 depth | TODO |
| ECC-TRCK-C523 | Corner | Double-submit create with Idempotency-Key + network retry + webhook autolink on new key | one issue; webhook once; key unique | API | P0 | ECC-TRCK-025 depth | TODO |
| ECC-TRCK-C524 | Corner | Issue reopen from Done in **completed** sprint ∥ sprint inventory GET ∥ burndown refresh | reopen policy vs sprint lock documented; velocity/burndown not silently rewrite history XOR explicit recalc flag | API | P1 | ECC-TRCK-036, TC-SPRINT-04/05 | TODO |
| ECC-TRCK-C525 | Corner | Automation: on link create → set label; human removes label; second link create same tick | final labels deterministic; no thrash loop; audit ordered | API | P1 | automation, TC-TRCK-06 | Partial |
| ECC-TRCK-C526 | Corner | Project key rename / prefix change ∥ open issue tabs ∥ JQL by old key ∥ webhook with old key | redirects or 404 documented; no dual identity; webhooks fail-closed on unknown | API/E2E | P0 | NFR-25, US-S04 | Future |
| ECC-TRCK-C527 | Corner | Mentions storm: 20 @user in one comment ∥ one user deactivated mid-fanout ∥ issue delete | notif partial OK; deactivated skipped; delete stops further delivery | API | P2 | TC-TRCK-05 | TODO |
| ECC-TRCK-C528 | Corner | Sub-task create under Story ∥ Story move cross-board column WIP-full ∥ parent soft-delete | sub-task not orphaned under deleted; WIP reject atomic with parent move | API | P1 | hierarchy, WIP | TODO |
| ECC-TRCK-C529 | Corner | Export PDF issue ∥ concurrent PUT description + attachment add | PDF snapshot coherent (old or new gen); not mixed half-fields; no 500 | API | P2 | export | TODO |

**Tracker Corner в этом файле: 30**

---

## 2. Sprint — `ECC-SPR-C500+`

> Wave3: poker∩start/complete · date shift mid-active · burndown UTC day skew · capacity replan mid-add · velocity recalc mid-mutation · retro mid-edit

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-C500 | Corner | Planning poker: votes in-flight ∥ admin Start sprint S ∥ one voter still POST vote | start wins or waits; late vote 400; points on issues = revealed XOR cleared per policy; at most one active | API/E2E | P0 | TC-SPRINT-02/06 | TODO |
| ECC-SPR-C501 | Corner | Poker reveal ∥ two facilitators Reveal + Reset same second | one terminal state (revealed XOR reset); votes not half-visible; WS clients converge | API/WS | P0 | TC-SPRINT-06 | TODO |
| ECC-SPR-C502 | Corner | Poker vote on issue I ∥ I moved out of sprint to backlog mid-vote | vote rejected/ignored; backlog issue not scored into sprint capacity | API | P1 | TC-SPRINT-06 | TODO |
| ECC-SPR-C503 | Corner | Poker apply points to issue ∥ human PUT estimate different value ∥ version race | one writer wins; estimate monotonic with version; poker UI conflict | API | P0 | NFR-24, estimate | TODO |
| ECC-SPR-C504 | Corner | Start S ∥ poker session still open on S ∥ add issue from backlog | start closes poker or blocks add; no active sprint with open poker ghost room | API | P1 | TC-SPRINT-02/06 | TODO |
| ECC-SPR-C505 | Corner | Complete S ∥ poker on Next (planning) reveal applies points to issues being moved into Next | points land on Next issues only; velocity S snapshot untouched by Next poker | API | P1 | TC-SPRINT-05/06/07 | TODO |
| ECC-SPR-C506 | Corner | Active sprint: admin PATCH end_date −1 day crossing UTC midnight ∥ burndown GET from TZ+14 and TZ−10 clients | series keyed by UTC days; no duplicate/missing day bucket; NFR-12 | API | P0 | ECC-SPR-018 depth, NFR-12 | TODO |
| ECC-SPR-C507 | Corner | Clock skew: server date rolls UTC day while client still on previous local day; points burn + complete same window | burndown day assignment = server UTC; complete snapshot deterministic | API | P0 | NFR-12 | TODO |
| ECC-SPR-C508 | Corner | PATCH sprint dates while burndown cache warm ∥ second client GET stale cache | cache invalidate; no mixed old/new axis; 200 consistent | API | P1 | burndown cache | TODO |
| ECC-SPR-C509 | Corner | Capacity replan: lower capacity below planned points ∥ bulk add 10 issues ∥ start attempt | start warn/block per policy; bulk add doesn't corrupt capacity counter; documented | API | P1 | capacity | Future |
| ECC-SPR-C510 | Corner | Two PMs: A raises capacity, B removes high-point issues, C starts — same tick | start sees consistent (capacity, planned); no start with phantom over-capacity | API | P1 | TC-SPRINT-02 | Future |
| ECC-SPR-C511 | Corner | Velocity GET while complete flush still moving incompletes + points edits | velocity either waits or returns previous completed only; no partial sprint in series | API | P0 | TC-SPRINT-05/07 | TODO |
| ECC-SPR-C512 | Corner | Recalc velocity job ∥ admin deletes historical completed sprint mid-job | job fail-closed or skips deleted; series length consistent; no NaN | API | P1 | velocity | TODO |
| ECC-SPR-C513 | Corner | Retro notes edit ∥ complete already done ∥ second admin reopens sprint (if allowed) | retro bound to completed id; reopen policy blocks or forks retro; no silent wipe | API | P2 | ECC-SPR-017, US-S01 | Future |
| ECC-SPR-C514 | Corner | Retro export PDF ∥ concurrent retro PATCH + sprint rename | export coherent snapshot; rename doesn't 500 export | API | P2 | retro | Future |
| ECC-SPR-C515 | Corner | Add issue to planning sprint ∥ issue already in another planning sprint (swap) ∥ start both race | issue in ≤1 sprint; at most one active start success | API | P0 | TC-SPRINT-03, ECC-SPR-014 depth | TODO |
| ECC-SPR-C516 | Corner | Active sprint board: filter hide Done ∥ complete mid-filter ∥ card mid-drag to Done | complete/drag reconcile; no card in limbo; WIP Done column consistent | E2E | P1 | TC-SPRINT-07, US-M11 | TODO |
| ECC-SPR-C517 | Corner | Membership revoked mid poker vote + mid start | 403; vote not counted; start only if actor still admin; no partial start | API | P0 | RBAC | TODO |

**Sprint Corner в этом файле: 18**

---

## 3. TMS — `ECC-TMS-C500+`

> Wave3: BDD↔classic mid-run · suite-tree move∩Allure · shared-steps∩execute · dual JUnit+Allure · coverage/traceability mid-mutation · env matrix parallel

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-C500 | Corner | Convert case X classic→BDD while X in active run R execute in-flight | execute uses frozen step snapshot XOR 409; convert doesn't fork result mid-row; version chain linear | API | P0 | TC-TMS-01/02/03/06 | TODO |
| ECC-TMS-C501 | Corner | Convert BDD→classic ∥ Allure match on BDD scenario name ∥ human edit steps | match strategy documented post-convert; no execute into wrong body; stats OK | API | P0 | TC-ALLURE-05, TC-TMS-02 | TODO |
| ECC-TMS-C502 | Corner | Two editors: A converts type, B rollback to v1 classic, C execute | one type wins; rollback/convert ordered; execute fail-closed on conflict | API | P0 | TC-TMS-03/04/06 | TODO |
| ECC-TMS-C503 | Corner | Suite folder move (X from S1→S2) ∥ Allure batch resolving path S1/X ∥ execute X | match by ASID/id not stale path; execute OK; tree consistent; no duplicate case | API | P0 | suite, TC-ALLURE-05 | TODO |
| ECC-TMS-C504 | Corner | Suite delete S1 (non-empty) ∥ move cases out ∥ Allure import targeting S1 | delete blocked until empty XOR cascade policy; import unmatched for gone path; no orphan keys | API | P1 | suite | TODO |
| ECC-TMS-C505 | Corner | Drag suite tree reorder ∥ two Allure imports creating suites by path | no cycle; deterministic tree; import idempotent suite nodes | API/E2E | P1 | Allure, suite | TODO |
| ECC-TMS-C506 | Corner | Shared step SS edit body ∥ 5 cases using SS ∥ execute one case mid-edit | execute snapshot old or new documented; other cases see one coherent SS version; no mixed steps in one result | API | P0 | shared-steps | Future |
| ECC-TMS-C507 | Corner | Soft-delete shared step SS ∥ cases still reference ∥ Allure import ∥ execute | delete block or break refs with 409; import/execute fail-closed; no 500 | API | P0 | shared-steps, TC-TMS-07 | Future |
| ECC-TMS-C508 | Corner | Replace SS in case X ∥ run R already has result draft for X | draft invalidated or frozen; stats not double-count on re-execute | API | P1 | TC-TMS-06 | Future |
| ECC-TMS-C509 | Corner | JUnit webhook (signed) + Allure import + UI execute — three writers same case X in run R | SEC-02 always; one result winner; counts exact; no unsigned bypass | API | P0 | TC-TMS-08, SEC-02, NFR-24 | TODO |
| ECC-TMS-C510 | Corner | JUnit webhook replay (same delivery-id) ∥ Allure reimport ∥ delete result row | webhook idempotent; no stats drift; delete terminal for that result_id | API | P0 | SEC-02, idempotency | TODO |
| ECC-TMS-C511 | Corner | JUnit for project A signed with project B secret (cross) ∥ Allure for A | JUnit 401/403; Allure A OK; no cross-project write | API | P0 | SEC-02, NF-03 | TODO |
| ECC-TMS-C512 | Corner | Coverage report GET ∥ bulk link/unlink cases↔issues ∥ soft-delete case mid-report | report snapshot or 409; matrix no leak deleted case body; NF-03 | API | P1 | TC-TMS-10/11 | TODO |
| ECC-TMS-C513 | Corner | Traceability cell click-through ∥ issue soft-delete ∥ case execute | UI/API 404 on issue; execute still OK; matrix cell cleared | API/E2E | P1 | TC-TMS-11, TC-TRCK-12 | TODO |
| ECC-TMS-C514 | Corner | Env matrix: run R on env=staging + env=prod parallel execute same case X | results isolated per env; stats per env; no cross-env overwrite | API | P1 | env matrix | Future |
| ECC-TMS-C515 | Corner | Change run env label mid execute storm (10 results) | in-flight bind old env or reject; completed rows immutable env; counts split correct | API | P1 | env | Future |
| ECC-TMS-C516 | Corner | Clone test case X→Y ∥ add both to run ∥ Allure ASID still points X ∥ execute Y | Allure→X; Y independent; no ASID steal; stats per case | API | P1 | clone, TC-ALLURE-05 | TODO |
| ECC-TMS-C517 | Corner | Case key rename X→Z ∥ open run still lists X ∥ JUnit posts for X ∥ UI execute Z | remap or unmatched documented; no dual results for one logical case; run list consistent | API | P0 | TC-TMS-05/06/08 | TODO |
| ECC-TMS-C518 | Corner | Complete run R ∥ flaky re-execute API late ∥ coverage job | late execute 400; flaky/coverage use frozen completed stats | API | P0 | TC-TMS-06/10, ECC-TMS-009 depth | TODO |
| ECC-TMS-C519 | Corner | Export CSV run ∥ Allure still appending ∥ one case GDPR hard-delete | export coherent or 409; hard-delete redacts from live views; historical export policy documented | API | P1 | TC-TMS-12, US-S03 | TODO |
| ECC-TMS-C520 | Corner | Membership revoke mid suite move + Allure + execute | 403 writes; no partial suite move committed under revoked user; import/execute stop | API | P0 | RBAC, ECC-TMS-010 depth | TODO |
| ECC-TMS-C521 | Corner | BDD outline with Examples table edit ∥ execute parameterized row ∥ Allure match scenario outline | row identity stable; match doesn't attach to wrong example; stats per row policy | API | P1 | BDD, Allure | Future |
| ECC-TMS-C522 | Corner | Two runs R1/R2 share case set; bulk status update R1 ∥ complete R2 ∥ Allure targets both | R1/R2 isolated; complete R2 frozen; bulk doesn't touch R2 | API | P0 | TC-TMS-05/06 | TODO |
| ECC-TMS-C523 | Corner | Attach evidence file to result ∥ result status flip pass↔fail ∥ case soft-delete | evidence GC policy; status flip audited; delete doesn't 500 on evidence FK | API | P1 | storage, TC-TMS-06/07 | TODO |

**TMS Corner в этом файле: 24**

---

## Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Tracker | ECC-TRCK-C500 … C529 | **30** |
| Sprint | ECC-SPR-C500 … C517 | **18** |
| TMS | ECC-TMS-C500 … C523 | **24** |
| **Итого Corner** | | **72** |

### Фокус-покрытие (wave3)

| Фокус | IDs (примеры) |
|-------|----------------|
| Clone + imex + GH/GL webhook mid-flight | ECC-TRCK-C500…C506, C523, C526 |
| Attachment/presigned ∩ soft/hard-delete | ECC-TRCK-C507…C510, C529 |
| Workflow-schema edit mid-transition | ECC-TRCK-C511…C514, C528 |
| Cross-project move / reopen∩completed sprint | ECC-TRCK-C515…C516, C524 |
| Timer/time-log ∩ version; component/label | ECC-TRCK-C517…C520, C525, C527 |
| Poker ∩ start/complete | ECC-SPR-C500…C505 |
| Burndown UTC skew / date shift | ECC-SPR-C506…C508 |
| Capacity / velocity / retro mid-mutation | ECC-SPR-C509…C514 |
| BDD↔classic mid-run | ECC-TMS-C500…C502, C521 |
| Suite-tree move ∩ Allure | ECC-TMS-C503…C505 |
| Shared-steps ∩ execute | ECC-TMS-C506…C508 |
| Dual JUnit+Allure (+ SEC-02) | ECC-TMS-C509…C511 |
| Coverage/traceability/env matrix | ECC-TMS-C512…C515, C518…C519, C522…C523 |

> Overlap check: `C200–C206` / `C300–C339` / `C400–C430` (TRCK), `C200–C202` / `C300–C319` / `C400–C417` (SPR), `C200–C202` / `C300–C324` / `C400–C419` (TMS) — **не использованы**. Канон `EDGE-CORNER-CATALOG.md` — **не изменялся**. Коммит не создавался.
