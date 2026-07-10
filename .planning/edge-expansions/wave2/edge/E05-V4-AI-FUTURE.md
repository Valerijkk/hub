# Edge expansions wave2 — v4 + Future (AI / Wiki / SSO / MCP / MS / UX)

> **Тип:** только **Edge** (границы лимитов, TTL, empty/max, quota hard-stop).  
> **Wave:** 2 · ID namespace `*200+` — **не пересекает** wave1 `05-V4-AI-FUTURE.md` (`*100+`).  
> **Не трогает** основной каталог `.planning/EDGE-CORNER-CATALOG.md` — foresight backlog для будущих `TC-*`.  
> **Gap:** только `Future` | `TODO`.  
> **Фокус wave2:** bases **50k rows / 200 cols** · wiki **page depth** · SSO **session TTL** · MCP **tool count** · MS **contract version** · a11y **24px target** · AI **token budget boundary**.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` (US-INTR / US-WIKI / US-F0x) · PLAN-bases (GZ-V4-19/20) · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · WCAG 2.2 AA (2.5.8 Target Size ≥24×24px).  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** из этой сессии без явной команды.

Формат строк = каталог ECC: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 Bases — 50k rows / 200 cols (+ formula/cell) (`ECC-V4-200+`)

> Углубление границ PLAN-bases / GZ-V4-19/20 относительно `ECC-V4-123..127` (wave1): offset/limit у потолка, bulk insert, col type mix, formula fan-out, share scope на max таблице.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-200 | Edge | Bases: ровно **49_999** rows → insert одной → **50_000** | 201; row_count=50000 | API | P0 | GZ-V4-19/20 | Future |
| ECC-V4-201 | Edge | Bases: insert при **50_000** rows (ровно лимит) | 400/422; row_count остаётся 50000 | API | P0 | GZ-V4-19 | Future |
| ECC-V4-202 | Edge | Bases: bulk insert N строк, где current+N = 50000 | 201 all; current+N = 50001 → reject entire batch (no partial) | API | P0 | PLAN-bases | Future |
| ECC-V4-203 | Edge | Bases: list rows `offset=49990&limit=20` на таблице 50k | 200, ≤10 items; no wrap/leak past end | API | P1 | pagination | Future |
| ECC-V4-204 | Edge | Bases: list `offset=50000&limit=1` | 200 [] (не 404) | API | P2 | empty page | TODO |
| ECC-V4-205 | Edge | Bases: ровно **199** cols → add 200-я | 201; schema.cols=200 | API | P0 | GZ-V4-20 | Future |
| ECC-V4-206 | Edge | Bases: add col при **200** cols | 400; schema unchanged | API | P0 | PLAN-bases Q3 | Future |
| ECC-V4-207 | Edge | Bases: rename/reorder при 200 cols (без add) | 200 OK; count остаётся 200 | API | P1 | schema mutate | Future |
| ECC-V4-208 | Edge | Bases: delete 1 col at 200 → add again | delete OK; re-add до 200 OK; 201-я снова 400 | API | P1 | col lifecycle | Future |
| ECC-V4-209 | Edge | Bases: 200 cols × mixed types (text/number/date/select/relation) | create OK; invalid type on 201-й slot N/A | API | P1 | col types | Future |
| ECC-V4-210 | Edge | Bases: **30** formula cols на таблице с 50k rows | recalc completes or async job; no OOM; 31-я formula → 400 | API | P0 | PLAN-bases formula | Future |
| ECC-V4-211 | Edge | Bases: formula referencing col index 200 (last) | OK; ref to missing/201 → 400 | API | P1 | formula bound | Future |
| ECC-V4-212 | Edge | Bases: cell string length = max policy на row 50000 | OK at max; max+1 → 400; row persists | API | P1 | cell max | TODO |
| ECC-V4-213 | Edge | Bases: empty cell vs null vs omitted field на PATCH | normalize per schema; no phantom overwrite | API | P2 | cell empty | TODO |
| ECC-V4-214 | Edge | Bases: filter/sort на колонке 200 при 50k rows | 200 paged; p95 within NFR budget or documented degrade | API/Perf | P1 | GZ-V4-19 | Future |
| ECC-V4-215 | Edge | Bases: export CSV всей таблицы 50k×200 | stream/chunk OK; hard fail if over export quota | API | P1 | export limit | Future |
| ECC-V4-216 | Edge | Bases: duplicate table at 50k/200 | reject OR async copy with quota check; no silent truncate | API | P2 | clone | TODO |
| ECC-V4-217 | Edge | Bases: share link / project ACL на max table — list without membership | 403; no row count leak in error | API | P0 | GZ-SEC-08/12 | Future |

**Подсчёт §1:** 18

---

## 2. AI — token budget boundary (`ECC-AI-200+`)

> Углубление GZ-SEC-22 / US-F05 относительно `ECC-AI-101..111`: exact remaining tokens, multi-request race на границе, streaming last-chunk, model max vs org budget, reset window.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-200 | Edge | Org daily budget: remaining = **1** token, request needs 1 | OK; usage → limit; alert optional | API | P0 | GZ-SEC-22 | Future |
| ECC-AI-201 | Edge | Org daily budget: remaining = **1**, request needs **2** | **hard-stop** 429/403; provider not called; usage unchanged | API | P0 | hard-stop | Future |
| ECC-AI-202 | Edge | Org daily budget: remaining = **0** | hard-stop; no soft retry | API | P0 | US-F05 | Future |
| ECC-AI-203 | Edge | Estimated prompt tokens = remaining exactly | accept; actual overage after response → count + block next | API | P0 | estimate vs actual | Future |
| ECC-AI-204 | Edge | Completion tokens push usage from limit-10 to limit+5 | stop generation; bill only allowed slice OR full+block next per ADR; no silent overage without audit | API | P0 | streaming budget | Future |
| ECC-AI-205 | Edge | User budget limit < org; user remaining=0, org remaining>0 | user hard-stop; org counter untouched for blocked call | API | P0 | virtual keys | Future |
| ECC-AI-206 | Edge | Org remaining=0, user remaining>0 | org wins; hard-stop | API | P0 | min(budget) | Future |
| ECC-AI-207 | Edge | Budget window: request at T=UTC midnight−1ms vs +1ms | before: old day; after: counters reset; no double-count | Unit | P1 | NFR-12 | Future |
| ECC-AI-208 | Edge | Model context max tokens < org remaining | reject by model max (400); org budget not decremented | API | P1 | ECC-AI-011 | TODO |
| ECC-AI-209 | Edge | Model context max > org remaining, prompt fits remaining | OK until org boundary; not model max | API | P1 | budget vs context | Future |
| ECC-AI-210 | Edge | BYOK provider error after pre-reserve tokens | release reservation; usage not permanently inflated | API | P0 | reservation | Future |
| ECC-AI-211 | Edge | Concurrent N requests each needing remaining/N + 1 | at most floor(remaining/cost) succeed; rest hard-stop; no overspend | API | P0 | race at boundary | Future |
| ECC-AI-212 | Edge | Embedding/batch job: total tokens = daily limit | last batch OK or split; overage batch rejected whole | API | P1 | batch | Future |
| ECC-AI-213 | Edge | Ollama local: org cloud budget=0, local unlimited policy | local OK; any cloud model still hard-stop | Unit | P1 | ECC-AI-019 | TODO |
| ECC-AI-214 | Edge | Budget alert threshold 80%/100%: usage hits exactly 80% then 100% | notify at 80; hard-stop at 100; no duplicate storm | API | P2 | alerts | TODO |

**Подсчёт §2:** 15

---

## 3. Future Wiki — page depth (`ECC-WIKI-200+`)

> Углубление US-WIKI-01 depth≤10 относительно `ECC-WIKI-100..112`: create/move/copy/import на depth 9→10→11, root=0/1 semantics, soft-delete restore depth.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-200 | Edge | Create child under page at depth **9** → child depth **10** | 201; tree valid | API | P0 | US-WIKI-01 | Future |
| ECC-WIKI-201 | Edge | Create child under page at depth **10** | 422; parent unchanged | API | P0 | depth max | Future |
| ECC-WIKI-202 | Edge | Move page (subtree height 1) under parent at depth 9 → result depth 10 | 200 OK | API | P0 | move | Future |
| ECC-WIKI-203 | Edge | Move page under parent such that deepest descendant would be depth **11** | 422; no partial move | API | P0 | subtree depth | Future |
| ECC-WIKI-204 | Edge | Copy/duplicate page at depth 10 | OK as sibling at 10 OR reject per policy; never depth 11 | API | P1 | copy | Future |
| ECC-WIKI-205 | Edge | Root page depth semantics: space root = 0 or 1 | documented; max leaf still ≤10 absolute | Unit | P1 | depth model | TODO |
| ECC-WIKI-206 | Edge | Import markdown tree with depth 10 | import OK; depth 11 node skipped/fail whole per policy | API | P1 | import | Future |
| ECC-WIKI-207 | Edge | Soft-delete leaf at depth 10 → restore | restore OK at depth 10; if parent chain shortened, depth recalculated ≤10 | API | P1 | NFR-11 | Future |
| ECC-WIKI-208 | Edge | Reparent to root: page was depth 10 → depth 1/0 | OK; children depths shift down | API | P1 | reparent | Future |
| ECC-WIKI-209 | Edge | Breadcrumb/API tree expand at depth 10 | returns full path ≤10; no extra phantom level | API/E2E | P2 | UI tree | TODO |
| ECC-WIKI-210 | Edge | Title/slug empty vs max length at depth 10 page | empty 400; max OK; max+1 400 | API | P2 | page fields | TODO |

**Подсчёт §3:** 11

---

## 4. Future SSO — session TTL (`ECC-SSO-200+`)

> ADR-AUTH-SSO + существующие JWT 15m / refresh 30d: границы TTL для OIDC session, break-glass, idle, absolute timeout, clock skew на expiry.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-200 | Edge | Access JWT: now = iat+15m−1s | 200 API OK | API | P0 | JWT 15m | Future |
| ECC-SSO-201 | Edge | Access JWT: now = iat+15m+1s | 401; refresh required | API | P0 | JWT expiry | Future |
| ECC-SSO-202 | Edge | Refresh token: age = 30d−1s | refresh OK; new access issued | API | P0 | refresh 30d | Future |
| ECC-SSO-203 | Edge | Refresh token: age = 30d+1s | 401; session dead; re-login SSO/magic | API | P0 | refresh expiry | Future |
| ECC-SSO-204 | Edge | OIDC IdP session TTL shorter than Hub refresh | Hub refresh fails closed when IdP session ended if policy requires; or Hub-only until 30d per ADR | API | P0 | IdP TTL | Future |
| ECC-SSO-205 | Edge | SSO idle timeout = policy N min: last activity = N−1s | session alive | API | P1 | idle TTL | Future |
| ECC-SSO-206 | Edge | SSO idle timeout: last activity = N+1s | 401; refresh rejected if idle-bound | API | P1 | idle | Future |
| ECC-SSO-207 | Edge | Absolute session max (e.g. 12h) hit while refresh still <30d | force re-auth; refresh revoked | API | P1 | absolute TTL | Future |
| ECC-SSO-208 | Edge | Break-glass Magic Link session: TTL = access 15m / refresh policy | same Hub TTL as normal; no extended god-mode TTL | API | P0 | break-glass | Future |
| ECC-SSO-209 | Edge | Clock skew: token nbf/exp ± skew window | accept inside skew; reject outside | Unit | P1 | OIDC skew | Future |
| ECC-SSO-210 | Edge | Org toggle flip sso→magic: existing access until exp, refresh after flip | access OK until TTL; refresh per ADR (allow/deny) | API | P1 | org toggle | Future |
| ECC-SSO-211 | Edge | LDAP deactivate: refresh TTL remaining >0 | immediate revoke; remaining TTL irrelevant | API | P0 | LDAP-sync | Future |

**Подсчёт §4:** 12

---

## 5. Future MCP — tool count (`ECC-MCP-200+`)

> Curated toolset size / PAT scope count / autonomy steps — границы count относительно `ECC-MCP-100..109`.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-200 | Edge | Agent toolset: ровно **max curated tools** (N) | register OK; list length=N | API | P0 | curated toolset | Future |
| ECC-MCP-201 | Edge | Agent toolset: N+1 tools | 400; config not saved | API | P0 | tool count max | Future |
| ECC-MCP-202 | Edge | Toolset: **0** tools | 400 or allow empty read-only agent per ADR; documented | API | P1 | empty toolset | Future |
| ECC-MCP-203 | Edge | Toolset: **1** tool (min useful) | OK; only that tool invocable | API | P1 | min | Future |
| ECC-MCP-204 | Edge | Invoke tool index N (last allowed) | 200/OK side-effect within scope | API | P0 | allow-list | Future |
| ECC-MCP-205 | Edge | Invoke tool index N+1 / unknown name | 403; audit; no side-effect | API | P0 | GZ-SEC-23 | Future |
| ECC-MCP-206 | Edge | PAT scopes count = max | OK; max+1 scope → 400 | API | P1 | PAT scopes | Future |
| ECC-MCP-207 | Edge | Autonomy max steps = M: step M completes | OK; step M+1 requires human confirm | API | P0 | GZ-SEC-09/20 | Future |
| ECC-MCP-208 | Edge | Parallel tool calls = concurrency limit K | K OK; K+1 queue/429; no extra tool beyond set | API | P2 | concurrency | TODO |
| ECC-MCP-209 | Edge | Tool result payload size = max | OK; max+1 truncated/400 per policy | API | P1 | result size | TODO |

**Подсчёт §5:** 10

---

## 6. Future Microservices — contract version (`ECC-MS-200+`)

> NATS/event contract versioning + dual-run compatibility — углубление `ECC-MS-102..104`.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-200 | Edge | Event `schema_version=v1`, consumer supports v1 | process OK; ack | Contract | P0 | GZ-QA-08/09 | Future |
| ECC-MS-201 | Edge | Event `schema_version=v2`, consumer only v1 | DLQ/reject; metric; no silent drop | Contract | P0 | contracts/events | Future |
| ECC-MS-202 | Edge | Event missing `schema_version` field | reject at ingress; both sides schema-test fail | Unit | P0 | M0-02 | Future |
| ECC-MS-203 | Edge | Producer advertises v1+v2; consumer v1 during dual-run | v1 path OK; v2 not forced until flip | Contract | P0 | strangler | Future |
| ECC-MS-204 | Edge | Additive optional field in v1.1 (compatible) | v1 consumer ignores unknown; no 500 | Contract | P1 | expand-contract | Future |
| ECC-MS-205 | Edge | Removing required field in v2 without consumer bump | contract test FAIL; deploy blocked | Unit | P0 | breaking change | Future |
| ECC-MS-206 | Edge | NATS header `Nats-Msg-Id` duplicate with same version | idempotent consumer; once side-effect | API | P1 | dedup | Future |
| ECC-MS-207 | Edge | Route-flip 50%: mixed monolith v1 + service v1 consumers | both process; no double-publish without idempotency key | Infra | P0 | GZ-ARCH-16/17 | Future |
| ECC-MS-208 | Edge | Contract version matrix: hub-ai vN vs hub-core expects vN−1 | fail-closed startup or runtime reject | Infra | P0 | M3 split | Future |
| ECC-MS-209 | Edge | Empty payload / null body with valid version header | 400/reject; version alone insufficient | Contract | P2 | empty event | TODO |

**Подсчёт §6:** 10

---

## 7. Future Redesign a11y — 24px target (`ECC-UX-200+`)

> WCAG 2.2 **2.5.8 Target Size (Minimum)** ≥24×24 CSS px — углубление `ECC-UX-102..103` + spacing/adjacent targets.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-200 | Edge | Icon-only button: computed hit box **24×24** CSS px | PASS target-size | E2E | P0 | WCAG 2.5.8 | Future |
| ECC-UX-201 | Edge | Icon-only button: **23×23** | FAIL a11y gate | E2E | P0 | 24px | Future |
| ECC-UX-202 | Edge | Icon-only button: **24×23** (one axis short) | FAIL | E2E | P0 | both axes | Future |
| ECC-UX-203 | Edge | Adjacent icon buttons: centers ≥24px apart (spacing exception) | PASS if spacing exception met; else FAIL | E2E | P1 | 2.5.8 exceptions | Future |
| ECC-UX-204 | Edge | Adjacent icon buttons: gap such that targets overlap <24 | FAIL / redesign spacing | E2E | P1 | spacing | Future |
| ECC-UX-205 | Edge | Inline link in text: target height &lt;24 but exception «inline» | PASS under inline exception; documented | E2E | P2 | 2.5.8 inline | TODO |
| ECC-UX-206 | Edge | Zoom 200%: target still ≥24 CSS px | PASS (CSS px not device px) | E2E | P1 | zoom | Future |
| ECC-UX-207 | Edge | Touch target in Kanban card actions at 24px | PASS; keyboard alternative still required (2.5.7) | E2E | P0 | NFR-16 | Future |
| ECC-UX-208 | Edge | Mobile 375px width: toolbar icon cluster all ≥24 | PASS; no clipped hit area | E2E | P1 | mobile | Future |
| ECC-UX-209 | Edge | Focus ring outside 24px box but hit target ≥24 | PASS size; focus visible (2.4.7) separate check | E2E | P2 | focus vs target | TODO |

**Подсчёт §7:** 10

---

## Счётчик

| Домен | Prefix | Edge rows | Gap |
|-------|--------|-----------|-----|
| v4 Bases 50k/200 | `ECC-V4-200+` | **18** | Future/TODO |
| AI token budget boundary | `ECC-AI-200+` | **15** | Future/TODO |
| Wiki page depth | `ECC-WIKI-200+` | **11** | Future/TODO |
| SSO session TTL | `ECC-SSO-200+` | **12** | Future/TODO |
| MCP tool count | `ECC-MCP-200+` | **10** | Future/TODO |
| MS contract version | `ECC-MS-200+` | **10** | Future/TODO |
| Redesign a11y 24px | `ECC-UX-200+` | **10** | Future/TODO |
| **Итого Edge** | | **86** | только Edge; Happy/Neg/Corner не включались |

**Минимум 55:** выполнен (86 ≥ 55).  
**Пересечение `*100+`:** нет (все ID ≥200).  
**Основной каталог:** не изменялся.  
**Коммит:** не создавался.

### Handoff

- Далее: `qa-manual-ver` на полноту/техники/трассируемость.  
- После `verified: true` — выборочная автоматизация границ → `qa-auto1` (API/Unit) / `qa-auto2` (E2E a11y 24px).  
- Не мержить в `EDGE-CORNER-CATALOG.md` без отдельной команды на каталогизацию.
