# Corner Expansion C02 Wave2 — Tracker / Sprint / TMS

> **Статус:** foresight backlog · 2026-07-10 · владелец: `qa-manual`  
> **Тип:** **только Corner** (multi-actor races, mid-flight, triple-path, partial fail)  
> **ID ranges:** `ECC-TRCK-C400+`, `ECC-SPR-C400+`, `ECC-TMS-C400+` — **не пересекают** parent `C200…`, wave1 corners `C300…`, edge `*100+`, канон `001–040`  
> **Не заменяет** `.planning/july-2026/EDGE-CORNER-CATALOG.md`, `00-CORNER-PARENT.md`, wave1 `corners/C02-*.md`, `docs/TESTING.md`  
> **Не копировать** сценарии C200/C300 1:1 — только углубление / новые комбинации  
> **Фокус wave2:** triple-tab version conflict · workflow+WIP+assignee simultaneous · sprint complete+move race · Allure+run execute+delete case · epic delete with children mid-link  
> **Связи:** US-M03/M11/M12, US-S01, US-M05/S13/C02/C04 · NFR-11/12/24 · NF-03 · TC-TRCK-*, TC-SPRINT-*, TC-TMS-*, TC-ALLURE-*, TC-JQL-*

### Gap-легенда

| Gap | Значение |
|-----|----------|
| **TODO** | Нет dedicated corner-теста / TC |
| **Partial** | Happy/unit есть, race/mid-flight не закрыт |
| **Future** | Поведение ещё не зафиксировано в ADR/политике |

---

## 1. Tracker — `ECC-TRCK-C400+`

> Wave2: triple-tab optimistic lock · workflow∥WIP∥assignee в одном окне · epic soft-delete mid child-link

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TRCK-C400 | Corner | Три вкладки A/B/C, issue v=5: A PUT title, B PUT description, C PUT priority — все с version=5 в одном окне ~50ms | ровно один 200 (version=6); два 409; финальное тело = поля победителя + неизменённые поля БД; нет silent merge трёх тел | API/E2E | P0 | NFR-24, TC-TRCK-03/04 | TODO |
| ECC-TRCK-C401 | Corner | Triple-tab cascade: A wins → B re-GET merge → B saves → C всё ещё шлёт исходный v=5 | C всегда 409; после B version=7; C обязан re-GET, не «угадывать» v | E2E | P0 | NFR-24 | TODO |
| ECC-TRCK-C402 | Corner | Triple-tab: A и B меняют **одно** поле (title) разными строками; C меняет labels | один из A/B 200; другой 409; C 409 или 200 только если выиграл первым; title не «склеивается» | API | P0 | TC-TRCK-03 | TODO |
| ECC-TRCK-C403 | Corner | Triple-tab offline queue: A/B/C теряют сеть, правят локально, reconnect flush в разном порядке | сервер сериализует по version; UI каждой вкладки показывает conflict/reload; нет тройного overwrite | E2E | P1 | NFR-24, FE | TODO |
| ECC-TRCK-C404 | Corner | Triple-tab + WS: каждая вкладка подписана на issue room; порядок conflict-событий разный | баннер конфликта на проигравших; победитель без ложного 409; нет clobber из stale WS payload | E2E/WS | P1 | WS, NFR-24 | TODO |
| ECC-TRCK-C405 | Corner | Tab A: autosave debounce PUT; Tab B: explicit Save; Tab C: blur-save — три клиента, один version | debounce не обходит lock; максимум один success на поколение version | E2E | P1 | FE autosave | TODO |
| ECC-TRCK-C406 | Corner | Triple-tab: A custom-field, B watcher add, C title — watcher side-effect vs versioned PUT | versioned fields подчиняются 409; watcher add либо вне version, либо documented atomic с version++ | API | P1 | TC-TRCK-03, ECC-TRCK-030 | TODO |
| ECC-TRCK-C407 | Corner | Triple-tab после того как D soft-delete issue между открытием и save | все PUT → 404/410; UI gone; никто не воскрешает issue | API/E2E | P0 | TC-TRCK-12, NFR-11 | TODO |
| ECC-TRCK-C408 | Corner | Triple-tab: A откатывает форму к старому snapshot (Undo), B/C уже сохранили два поколения | A save со stale v → 409; Undo не форсирует downgrade version | E2E | P1 | NFR-24 | TODO |
| ECC-TRCK-C409 | Corner | Одновременно: workflow transition в колонку с WIP=full, assignee PATCH, Kanban drag той же issue | не более одного успешного мутатора статуса/колонки; assignee не «полуприменён» при WIP-reject; version monotonic | API/E2E | P0 | US-M11, TC-TRCK-10, NFR-24 | TODO |
| ECC-TRCK-C410 | Corner | Один PUT body: `{status_id→WIP-full column, assignee_id, version}` | 400/409 WIP; assignee **и** status не применены (atomic); version без ++ | API | P0 | C201-depth, KANB-02 | TODO |
| ECC-TRCK-C411 | Corner | Actor1 transition In Progress; Actor2 drag в Done (WIP=1 уже занят); Actor3 assignee=null — same tick | финальное (status, column, assignee) consistent; проигравшие 409/WIP; board = API | API/E2E | P0 | US-M11, TC-TRCK-10 | TODO |
| ECC-TRCK-C412 | Corner | Transition требует assignee; параллельно WIP-drag в целевую колонку и PATCH assignee=user | либо transition ждёт assignee затем проверяет WIP, либо единый reject; нет Done без assignee при policy | API | P0 | workflow gate | TODO |
| ECC-TRCK-C413 | Corner | Automation: on assignee → auto-transition; human одновременно drag против WIP | automation и human не дают ghost column + broken WIP count | API | P1 | automation, WIP | Partial |
| ECC-TRCK-C414 | Corner | Bulk assignee на 10 issues + single-issue workflow + WIP drag overlapping один key | bulk row и single не double-apply; WIP count после всех = факт колонки | API | P1 | bulk, US-M11 | TODO |
| ECC-TRCK-C415 | Corner | Swimlane «by assignee» + drag (WIP) + assignee change mid-drag (карточка должна сменить lane) | drop reject или lane re-target; нет исчезнувшей карточки; WIP source/target корректны | E2E | P1 | ECC-TRCK-023, WIP | TODO |
| ECC-TRCK-C416 | Corner | Invalid transition + WIP-ok column + valid assignee в одном запросе | 400 transition; WIP/assignee unchanged | API | P0 | TC-TRCK-11 | TODO |
| ECC-TRCK-C417 | Corner | Admin снижает WIP N→0(unlimited) в тот же миг, что transition+assignee+drag | либо все видят новый лимит, либо reject по старому snapshot — documented; нет split-brain count | API/E2E | P1 | WIP edit | TODO |
| ECC-TRCK-C418 | Corner | Soft-delete **epic** E пока child Story S создаёт link `parent=E` / `is_child_of` | link 404/410 или block delete; нет orphan link row на deleted epic | API | P0 | TC-TRCK-06/12, epic | TODO |
| ECC-TRCK-C419 | Corner | Soft-delete epic E; два children одновременно POST links E←S1, E←S2 | оба fail-closed или один OK до delete commit; после delete — zero links to E; FK intact | API | P0 | TC-TRCK-06 | TODO |
| ECC-TRCK-C420 | Corner | Soft-delete epic mid-flight: child меняет link type blocks→relates на существующей связи с E | 404/409; нет link на deleted; audit | API | P1 | TC-TRCK-06 | TODO |
| ECC-TRCK-C421 | Corner | Epic delete + child «move under other epic» (re-parent) + third actor creates link to old epic | re-parent OK XOR old-link fail; нет dual-parent; нет link к deleted | API | P0 | epic hierarchy | TODO |
| ECC-TRCK-C422 | Corner | Epic soft-delete + create Sub-task under Story, у которой parent=epic, пока delete flush | sub-task не создаётся с цепочкой на deleted epic; 404 parent chain | API | P1 | hierarchy | TODO |
| ECC-TRCK-C423 | Corner | Epic delete (50 children) ∥ bulk create links children→epic | partial: успевшие до delete / остальные 404; post-state: no links to deleted epic | API | P1 | bulk, C206-depth | TODO |
| ECC-TRCK-C424 | Corner | Soft-delete epic → restore race с child link create, который получил 404 и ретраит | после restore: link либо создаётся на живой epic, либо явный re-GET; нет link к «призраку» pre-restore id mismatch | API | P1 | NFR-11 | TODO |
| ECC-TRCK-C425 | Corner | GDPR hard-delete epic ∥ child link + JQL `parent = EKEY` mid-query | hard-delete wins; JQL/search без leak; link create 404; NF-03 | API | P0 | US-S03, NF-03 | TODO |
| ECC-TRCK-C426 | Corner | Epic delete + board WS: children ещё показывают parent chip пока link API уже 404 | UI reconcile; chip cleared; нет клика в 404 loop без ошибки | E2E/WS | P1 | WS, epic | TODO |
| ECC-TRCK-C427 | Corner | Circular link attempt (S→E и E→S) пока epic soft-deleting | оба reject; delete завершается; нет cycle row | API | P1 | TC-TRCK-06 | TODO |
| ECC-TRCK-C428 | Corner | Triple-tab edit **epic** fields + child link create + WIP drag child на board | epic version lock независим от child move; link/WIP не ломают epic version monotonicity | API/E2E | P1 | NFR-24, WIP | TODO |
| ECC-TRCK-C429 | Corner | Workflow+WIP+assignee на child, пока epic parent soft-delete | child mutation OK если policy decoupled; parent chip cleared; нет FK error 500 | API | P1 | epic, WIP | TODO |
| ECC-TRCK-C430 | Corner | Triple-tab: Tab A меняет parent=epic1→epic2; B soft-deletes epic2; C links child to epic2 | A/C fail-closed на deleted target; parent field consistent; no orphan | API/E2E | P0 | epic, NFR-24 | TODO |

**Tracker Corner в этом файле: 31**

---

## 2. Sprint — `ECC-SPR-C400+`

> Wave2: complete ∥ multi-move (backlog/next/other) · inventory integrity под тройным давлением

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SPR-C400 | Corner | Complete sprint S: Actor1 complete(move→next); Actor2 PATCH issue→backlog; Actor3 PATCH same issue→S2 — same window | issue ровно в одном месте (next XOR backlog XOR S2); не lost, не duplicated; complete inventory сходится | API | P0 | TC-SPRINT-07, NFR-24 | TODO |
| ECC-SPR-C401 | Corner | Complete S + Kanban drag issue out of sprint board + API move into S from backlog (opposite directions) | drag и API не создают ghost в S после completed; final sprint_id consistent | API/E2E | P0 | TC-SPRINT-07, US-M11 | TODO |
| ECC-SPR-C402 | Corner | Complete с destination=Next; Next одновременно start + принимает чужие adds | incompletes либо все в Next, либо fallback backlog если Next не ready; at most one active | API | P0 | TC-SPRINT-03/07 | TODO |
| ECC-SPR-C403 | Corner | Два incomplete issues swap sprint_id (A↔B) во время complete inventory scan | оба учтены ровно раз; swap не даёт «оба вне» или «оба в S» после complete | API | P0 | TC-SPRINT-07 | TODO |
| ECC-SPR-C404 | Corner | Complete: Done-issues остаются в S; Incomplete move race — Actor меняет status→Done mid-complete | policy: считается Done до cutoff или всё же move; counts/velocity snapshot deterministic | API | P1 | TC-SPRINT-05/07 | TODO |
| ECC-SPR-C405 | Corner | Три админа: A complete→Next, B complete→Backlog (разные body), C move issue | один complete побеждает; destination единый для всех incompletes; второй complete 400/idempotent | API | P0 | TC-SPRINT-07 | Future |
| ECC-SPR-C406 | Corner | Complete flush (N=200 issues) + points edit на issue mid-move + burndown GET | points на completed snapshot стабильны; burndown не NaN/500; move не теряет issue | API | P1 | TC-SPRINT-04/05 | TODO |
| ECC-SPR-C407 | Corner | Complete + soft-delete issue mid-routing to Next | deleted не появляется в Next; trash OK; inventory counts exclude | API | P0 | TC-TRCK-12, TC-SPRINT-07 | TODO |
| ECC-SPR-C408 | Corner | Complete + clone issue mid-complete (clone inherits sprint=S) | clone не остаётся в completed S как active orphan; policy: backlog/Next/block | API | P1 | clone, sprint | TODO |
| ECC-SPR-C409 | Corner | Complete S + move epic out + children still in S mid-flight | epic/children sprint fields consistent per hierarchy policy; no half-family split undocumented | API | P1 | epic, TC-SPRINT-07 | TODO |
| ECC-SPR-C410 | Corner | WS sprint board: complete event while card mid-drag to backlog column | drag cancel/reject; UI → completed; card не в двух местах | E2E/WS | P1 | WS, TC-SPRINT-07 | TODO |
| ECC-SPR-C411 | Corner | Complete S1 while issue moved to S2 (other planning sprint), S2 deleted mid-move | issue → backlog fallback; no NULL/broken FK; S1 completed clean | API | P0 | TC-SPRINT-07 | TODO |
| ECC-SPR-C412 | Corner | Complete + bulk «set sprint=null» на subset overlapping complete move-to-Next | bulk rows 409/skip; incompletes destination единый; no double write | API | P1 | bulk | TODO |
| ECC-SPR-C413 | Corner | Complete still flushing moves; second client starts Next early | start 409/wait until flush; never two actives; no issues stuck with sprint=S1 active flag | API | P0 | TC-SPRINT-03 | TODO |
| ECC-SPR-C414 | Corner | Capacity: Next already at capacity; complete dumps 30 incompletes into Next | reject dump / overflow warn / spill backlog — documented; no silent over-capacity corruption | API | P1 | capacity, Future | Future |
| ECC-SPR-C415 | Corner | Complete + poker reveal на planning Next + move incompletes into Next | votes/points не портят velocity S; Next planning state valid | API | P2 | TC-SPRINT-06/07 | TODO |
| ECC-SPR-C416 | Corner | Project archive kickoff ∥ complete+move race | archive blocked until complete terminal OR complete blocked; no archived project with active sprint_id refs | API | P1 | archive | TODO |
| ECC-SPR-C417 | Corner | Complete S + issue workflow transition + assignee change + move out (tracker triple ∩ sprint) | sprint_id и status/assignee финал согласованы; один source of truth | API | P0 | TRCK C409∩SPR | TODO |

**Sprint Corner в этом файле: 18**

---

## 3. TMS — `ECC-TMS-C400+`

> Wave2: Allure import ∥ run execute ∥ case delete — тройной mid-flight; stats/match integrity

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-TMS-C400 | Corner | Одновременно: Allure import match case X; PUT execute result X in run R; soft-delete case X | delete не 500; execute 404/409 или counted-then-lock; import → unmatched/skipped; run stats = только валидные results; no resurrect X | API | P0 | TC-TMS-06/07, TC-ALLURE-05, TC-TRCK-12-analog | TODO |
| ECC-TMS-C401 | Corner | Soft-delete X стартует первым; Allure mid-batch ещё резолвит ASID→X; execute in-flight | match skip; execute reject; import completes; audit «case gone» | API | P0 | TC-ALLURE-05, ECC-TMS-018 | TODO |
| ECC-TMS-C402 | Corner | Human execute pass + Allure import fail для того же result row + delete case | один writer wins на result; stats = winner; delete не оставляет result→case FK break (nullify/snapshot) | API | P0 | TC-TMS-06/07, NFR-24 | TODO |
| ECC-TMS-C403 | Corner | Allure создаёт/апдейтит run R; admin complete R; execute late; delete matched case | late execute 400; complete stats frozen; delete не меняет completed snapshot | API | P0 | TC-TMS-06, ECC-TMS-009 | TODO |
| ECC-TMS-C404 | Corner | Delete case X → Allure rematch пытается relink → execute по старому result_id | no execute on deleted; rematch unmatched; result_id dead → 404 | API | P0 | TC-ALLURE-05 | TODO |
| ECC-TMS-C405 | Corner | Run R1 execute X; Run R2 Allure import X; soft-delete X — parallel | R1/R2 independently consistent; оба не падают 500; X gone; historical results snapshot OK | API | P1 | TC-TMS-05/06/07 | TODO |
| ECC-TMS-C406 | Corner | GDPR hard-delete case X ∥ Allure match ∥ execute | hard-delete wins; no resurrect via import; execute 404; SEC/audit | API | P0 | US-S03, TC-TMS-07 | TODO |
| ECC-TMS-C407 | Corner | Delete suite/folder containing X mid Allure batch + execute X | import skip X; execute reject; suite tree consistent; no orphan case key | API | P1 | suite, Allure | TODO |
| ECC-TMS-C408 | Corner | Rename case key / ASID label mid: Allure match + execute + delete | deterministic: unmatched or match new key; execute не пишет в чужой case; delete terminal | API | P1 | TC-ALLURE-05 | TODO |
| ECC-TMS-C409 | Corner | Failed execute → bug-from-fail create ∥ Allure reimport ∥ delete case | bug issue OK или policy block; case delete не orphan’ит bug link incorrectly; import stats OK | API | P1 | TC-TMS-09/07 | TODO |
| ECC-TMS-C410 | Corner | Rollback case X to v1 ∥ Allure match body ∥ execute ∥ soft-delete | version chain linear; match uses documented body revision; execute/delete fail-closed без fork history | API | P0 | TC-TMS-03/04/06/07 | TODO |
| ECC-TMS-C411 | Corner | Add X to run R ∥ soft-delete X ∥ Allure posts result for X | add 404/409; Allure unmatched; R без zombie case slot | API | P0 | TC-TMS-05/07 | TODO |
| ECC-TMS-C412 | Corner | Flaky re-execute loop on X + Allure batch + delete X mid-loop | loop stops clean; flaky report без crash; no stats drift after delete | API | P2 | TC-TMS-10 | TODO |
| ECC-TMS-C413 | Corner | Traceability link X↔issue ∥ Allure ∥ execute ∥ delete X | matrix cell empty/404 after delete; issue не leak case body; NF-03 | API | P1 | TC-TMS-11, NF-03 | TODO |
| ECC-TMS-C414 | Corner | Membership revoked mid triple (import/execute/delete) | 403 на write; import/execute не коммитят под revoked user; delete только если ещё был permission | API | P0 | RBAC, ECC-TMS-010 | TODO |
| ECC-TMS-C415 | Corner | Export PDF/CSV run R mid Allure+execute+delete X | export snapshot или 409 retry; файл не truncated/corrupt; stats внутри export = frozen view | API | P2 | TC-TMS-12 | Partial |
| ECC-TMS-C416 | Corner | Two Allure imports overlapping X + two executors + one delete | idempotent import; one execute winner; delete terminal; passed/failed counts exact | API | P0 | TC-TMS-07, NFR-24 | TODO |
| ECC-TMS-C417 | Corner | JUnit webhook result for X ∥ Allure ∥ execute UI ∥ delete X (SEC-02 path) | signature always checked; late writers reject after delete; no unsigned bypass | API | P0 | TC-TMS-08, SEC-02 | TODO |
| ECC-TMS-C418 | Corner | Case X in two active runs; delete X; Allure targets R1; execute on R2 | both runs clean skip/reject; no cross-run stats bleed | API | P1 | TC-TMS-05/06/07 | TODO |
| ECC-TMS-C419 | Corner | Soft-delete X + restore X race while Allure matched old generation + execute | restore не принимает stale execute/import на pre-delete generation без re-match; documented | API | P1 | NFR-11, Allure | TODO |

**TMS Corner в этом файле: 20**

---

## Счётчик

| Домен | ID range | Count |
|-------|----------|------:|
| Tracker | ECC-TRCK-C400 … C430 | **31** |
| Sprint | ECC-SPR-C400 … C417 | **18** |
| TMS | ECC-TMS-C400 … C419 | **20** |
| **Итого Corner** | | **69** |

### Фокус-покрытие (wave2)

| Фокус | IDs (примеры) |
|-------|----------------|
| Triple-tab version conflict | ECC-TRCK-C400…C408, C428, C430 |
| Workflow+WIP+assignee simultaneous | ECC-TRCK-C409…C417, C429; ECC-SPR-C417 |
| Sprint complete+move race | ECC-SPR-C400…C417 |
| Allure+run execute+delete case | ECC-TMS-C400…C419 |
| Epic delete with children mid-link | ECC-TRCK-C418…C427, C430 |

> Parent/wave1 overlap check: `C200–C206` / `C300–C339` (TRCK), `C200–C202` / `C300–C319` (SPR), `C200–C202` / `C300–C324` (TMS) — **не использованы**. Канон `EDGE-CORNER-CATALOG.md` и каталог — **не изменялись**. Коммит не создавался.
