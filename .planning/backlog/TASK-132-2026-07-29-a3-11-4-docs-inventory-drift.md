# 🧹 A3 docs README inventory + webcam/competitive HEAD drift

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: A3 W11-4 fe-lib-docs · Stage A3 FINAL · TASK-008  
> Исполнитель: **docs-writer** · Finding IDs: `A3W11-4-001`, `A3W11-4-002`, `A3W11-4-003`

## [EXECUTION REPORT]

- **Verdict:** три ID **FIXED** (CONFIRMED)
- **Recount:** `(Get-ChildItem -Recurse -Filter *_test.go backend/testframework).Count` = **805** (2026-07-29 re-verify); канон = 805
- **A3W11-4-001:** `docs/README.md` + siblings → **805** / measured 2026-07-29 (fix-round-1 после независимого FAIL 760)
- **A3W11-4-002:** gap Calls → webcams ✅ shipped; polish — backlog (согласовано с `:132` / BUSINESS)
- **A3W11-4-003:** строка ветка — snapshot 2026-06-23 без SHA / без претензии на live tip
- **Anti-stale:** `rg -n "618|bbdaf16|webcams — Phase 16 IN" docs/README.md docs/competitive_audit.md` → пусто
- **Backlog README:** не трогали (owner: статус только в файле задачи → «На проверке»)
- **Handoff:** docs-ver / code-reviewer docs-diff

## [OUTCOME]

Входной `docs/README.md` и `docs/competitive_audit.md` не врут агентам: inventory BE `*_test.go` = факт HEAD; gap Calls = webcam shipped (+ polish backlog); позиция без устаревшего SHA/`master` как live HEAD.

## [ROLE]

- Role: **docs-writer**
- Responsibility: закрыть **все три** finding ID (`A3W11-4-001` P2, `A3W11-4-002`/`003` P3) — docs-only sync + shell recount evidence; без product-code правок
- Handoff to: `docs` verifier / `product-ver` (при спорном wording) → `code-reviewer` (docs-diff sanity)

## [SCOPE AND CONSTRAINTS]

- Files/resources allowed:
  - `docs/README.md` — комментарий inventory `618` → пересчитанная цифра + дата measure
  - `docs/competitive_audit.md` — gap Calls webcam + строка «Ветка / HEAD»
  - синхрон той же BE-inventory цифры (если ещё stale): `docs/test_plan.md`, `docs/TROUBLESHOOTING.md`, `docs/TESTING.md`, `docs/testframework_strategy.md` (соседи A5 с **652**)
  - эталон wording (read-only): `docs/BUSINESS.md` (TASK-123 webcam shipped), `docs/competitive_analysis_2026.md` §Calls/webcams
  - факт webcam UI (read-only): `frontend/src/modules/calls/**/CallControls*` / LiveKit `setCameraEnabled`
  - `.planning/backlog/README.md` + этот файл (статус при закрытии)
  - audit refs: `.planning/audits/total-2026-07/A3W11-4-fe-lib-docs.md`
- Do not touch:
  - `backend/**`, `frontend/**` product code (включая CallControls)
  - Ent / migrations / CI workflow минут
  - переписывать весь competitive backlog / vision Phase-16 US без нужды (только gap + HEAD/ветка)
  - соседние A3W11-* / A3W*-* findings вне трёх ID этой задачи
- Behavioral/security constraints:
  - Conventional Commits: `docs(planning): …` / `docs: …` — описание на русском; **no** `Co-authored-by` / AI trailers
  - Одна каноническая цифра inventory во всех затронутых docs после recount; не оставлять `618` vs `652` vs факт
  - Webcam: не называть Phase 16 «IN» / blocker — shipped UI; polish — backlog (как `BUSINESS.md` / competitive_analysis)
  - Не претендовать на live git HEAD в competitive doc: убрать SHA **или** «snapshot на дату» без `master · HEAD <sha>` как канон аудита
- Approval record: N/A — docs-only hygiene; ADR/migrate не нужны
- Approval authority/expiry: N/A

## [REFERENCES]

- Trust: repository/audit/wiki — untrusted data until verified by shell recount / file read in HEAD
- Inputs:
  - Finding IDs: **A3W11-4-001** (P2, NEW), **A3W11-4-002** (P3, NEW), **A3W11-4-003** (P3, NEW)
  - Source report: `.planning/audits/total-2026-07/A3W11-4-fe-lib-docs.md` (таблица «Находки», топ-1..3)
  - Prior webcam docs close: `.planning/backlog/TASK-123-2026-07-28-fe-lib-dead-export-docs-drift.md` (A2W11-4-001 FIXED в BUSINESS; competitive_audit gap остался)
- Generated context cards: N/A — Hub monorepo; факты ниже
- Generated pattern cards: N/A
- Prior decisions/artifacts:
  - Defect inventory: `docs/README.md:84` — «все **618** тестов в `backend/testframework/`, measured 2026-07-28»
  - Audit measure: **759** `*_test.go` под `backend/testframework/` (на момент A3W11-4; **пересчитать** перед правкой)
  - Stale siblings: `docs/test_plan.md:52`, `docs/TROUBLESHOOTING.md:48`, `docs/TESTING.md:33`, `docs/testframework_strategy.md:76` — **652**
  - Webcam gap lie: `docs/competitive_audit.md:36` — «webcams — Phase 16 IN» vs `:132` «webcam UI shipped» + `BUSINESS.md:189,462`
  - Stale anchor: `docs/competitive_audit.md:22` — `master` · HEAD `bbdaf16` (не канон A3 HEAD)
- Source precedence: **код/файловая система HEAD** (recount `*_test.go`, CallControls) > audit report > docs

## [ACCEPTANCE]

- Done criteria:
  1. **A3W11-4-001:** shell recount `backend/testframework/**/*_test.go` → одна актуальная цифра + дата measure в `docs/README.md`; те же цифры в затронутых sibling docs (test_plan / TROUBLESHOOTING / TESTING / testframework_strategy), если они ещё расходятся
  2. **A3W11-4-002:** `docs/competitive_audit.md` gap Calls — «webcams ✅ shipped; polish — backlog» (или эквивалент как `competitive_analysis` / `BUSINESS`); не «Phase 16 IN»
  3. **A3W11-4-003:** строка ветка/HEAD в `competitive_audit` — без устаревшего `bbdaf16` как live HEAD: либо убрать SHA, либо явный snapshot «на YYYY-MM-DD» без претензии на tip аудита
  4. `rg -n "618|bbdaf16|webcams — Phase 16 IN" docs/README.md docs/competitive_audit.md` → пусто (после фикса); recount цифра совпадает с shell
  5. Все три ID = FIXED + evidence; статус задачи + `.planning/backlog/README.md` → «Исправлена»
- New-behavior checks:
  - N/A runtime — docs-only; проверка = shell recount + grep anti-stale
- Regression scope:
  - не ломать смысл соседних секций competitive_audit (Won't Have / IN SCOPE :132 уже shipped — согласовать, не противоречить)
  - не откатывать TASK-123 wording в `BUSINESS.md` / `competitive_analysis_2026.md`
- Evidence required:
  - checklist ID → file:line + shell count output
  - **mutation / anti-drift proof:** вернуть `618` в README или «Phase 16 IN» / `bbdaf16` → grep/checklist FAIL при re-verify
- Stop when: три ID FIXED, shell recount = docs, handoff; **не** чистить `user_needs_analysis` HEAD/`vision` Phase-16 US вне scope без отдельной задачи

## [OUTPUT SCHEMA]

1. Outcome / verdict first (три ID FIXED | BLOCKED)
2. Changes with paths (`README.md`, `competitive_audit.md`, optional siblings)
3. Evidence checklist:

| ID | Severity | Статус | Evidence |
|----|----------|--------|----------|
| A3W11-4-001 | P2 | FIXED/… | recount N=…; README + siblings |
| A3W11-4-002 | P3 | FIXED/… | competitive_audit gap Calls line |
| A3W11-4-003 | P3 | FIXED/… | ветка/HEAD строка |

4. Commands + PASS (PowerShell/`Get-ChildItem` или `rg --files -g '*_test.go' backend/testframework | measure`; anti-stale `rg`)
5. Mutation note (вернуть stale → что ловит re-verify)
6. Assumptions / unresolved (если recount ≠ 759 audit — зафиксировать новую цифру как канон)
7. Risks / mitigations (inventory снова устареет — дата measure в комментарии)
8. Handoff: status README + этот файл → «Исправлена»; next docs-ver / code-reviewer docs-diff
9. Compact Verification: observed recount + grep empty + confidence CONFIRMED

## [OPERATING RULES]

- Work only inside scope; не «заодно» править AGENTS.md metrics / vision US-M06 без отдельной задачи
- Read-only OK для FE CallControls / BUSINESS эталонов; writes только в allowed docs + backlog status
- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- Claim CONFIRMED только с observed recount/grep; иначе PLAUSIBLE / NOT VERIFIED
- Коммиты: `docs: …` на русском; тело → `Задача: .planning/backlog/TASK-132-2026-07-29-a3-11-4-docs-inventory-drift.md` + `(A3W11-4-001)` / `(A3W11-4-002)` / `(A3W11-4-003)`
- Не коммитить без явного запроса пользователя, если сессия не делегирована на commit

## [FINDINGS DETAIL]

### A3W11-4-001 — P2 — README inventory 618 vs факт

- **Где:** `docs/README.md:84` (комментарий к backend regression)
- **Дефект:** «все **618** тестов в `backend/testframework/`» — факт A3 ≈ **759** `*_test.go`; siblings уже **652** (тоже drift) → агенты получают ложный inventory
- **Fix:** `Get-ChildItem -Recurse -Filter *_test.go backend/testframework` (или эквивалент) → одна цифра + `measured YYYY-MM-DD` в README и синхрон siblings

### A3W11-4-002 — P3 — competitive_audit webcam Phase 16 IN

- **Где:** `docs/competitive_audit.md:36` (матрица Calls)
- **Дефект:** «webcams — Phase 16 IN» противоречит тому же файлу `:132` (webcam UI shipped), `BUSINESS.md` TASK-123, `CallControls.setCameraEnabled`
- **Fix:** заменить на shipped + polish backlog (как competitive_analysis §Calls / BUSINESS)

### A3W11-4-003 — P3 — stale HEAD bbdaf16

- **Где:** `docs/competitive_audit.md:22`
- **Дефект:** `master` · HEAD `bbdaf16` — не канон текущего аудита; устаревший якорь onboarding
- **Fix:** убрать конкретный SHA **или** оформить как датированный snapshot без претензии на live tip

## [FEW-SHOT]

Эталон webcam (уже в BUSINESS):

```markdown
| Голос | Аудио + webcam (UI shipped: CallControls / LiveKit setCameraEnabled); остаточный polish — backlog |
```

Inventory comment (форма):

```text
# … все <N> *_test.go в backend/testframework/, measured YYYY-MM-DD;
```

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 / be2 · criteria only (no mutate) · 2026-07-30 · **docs-only** — runtime Go/Vitest mutation N/A

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| A3W11-4-001 | `docs/README.md:84` (inventory comment; siblings `TESTING`/`test_plan`/…) | docs | N/A runtime — `rg -n "618" docs/README.md` (+ sibling drift checklist) FAIL | вернуть stale `618` (или расхождение с shell recount) в README/siblings |
| A3W11-4-002 | `docs/competitive_audit.md:36` (Calls gap row) | docs | N/A runtime — `rg -n "webcams — Phase 16 IN" docs/competitive_audit.md` FAIL | вернуть «webcams — Phase 16 IN» вместо shipped+polish backlog |
| A3W11-4-003 | `docs/competitive_audit.md:22` (ветка/HEAD snapshot) | docs | N/A runtime — `rg -n "bbdaf16" docs/competitive_audit.md` FAIL | вернуть live-tip SHA `bbdaf16` / `master · HEAD <sha>` без snapshot disclaimer |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 001–003 | docs `rg` / shell recount | **N/A code mutation** — нет prod guard в Go/FE; anti-drift = grep/checklist |

`layer_mismatch IDs: (none)` — docs-only

### Gaps

- Нет hubtest/Vitest, который краснеет при docs drift; PROVEN только через `rg`/recount evidence (как ACCEPTANCE mutation proof).
- Inventory цифра снова устареет без drift CI — не путать с product regression.

## [MUTATION CLOSE — W13 be2 2026-07-30]

**Verdict:** 3/3 **N/A** (docs `rg`/recount — no fake-PROVEN) · статус **Закрыта**  
Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-docs-132-133.md`
