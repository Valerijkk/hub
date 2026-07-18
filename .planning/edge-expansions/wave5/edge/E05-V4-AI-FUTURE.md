# Edge expansions wave5 — v4 + Future (AI / Wiki / SSO / MCP / MS / UX)

> **Тип:** только **Edge** (границы лимитов, TTL, empty/max, quota hard-stop, cardinality).  
> **Wave:** 5 · ID namespace `*500+` — **не пересекает** wave1–4 (`*100+`…`*400+`).  
> **Не трогает** основной каталог `.planning/july-2026/EDGE-CORNER-CATALOG.md` — foresight backlog для будущих `TC-*`.  
> **Gap:** только `Future` | `TODO`.  
> **Фокус wave5:** bases **tables/base · relation/lookup/rollup depth · select options · cross-base forbid** · AI **max_tokens/top_p/stop · system prompt · OpenAI-compat body · TTFB** · wiki **issue-link max · optimistic version · base64 reject · search page** · SSO **PKCE/nonce/redirect_uri · LDAP DN/nest · issuer mismatch** · MCP **run wall-clock · write-deny · scope∩tool · confirm double** · MS **outbox batch · shadow % · upstream health · expand-contract null** · a11y **aria-live · focus order · 4.1.2 name/role · forced-colors · 3.3.1 errors**.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` (US-WIKI-04 / US-F06 / US-INTR-01) · PLAN-bases (Q3: 100 tables/base, dep depth ≤10, relations) · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · WCAG 2.2 AA (1.3.1, 2.4.3, 4.1.2, 3.3.1, forced-colors) · NF-07 SSRF.  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** из этой сессии без явной команды.

Формат строк = каталог ECC: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 Bases — tables/base · relations/lookup/rollup · select options (`ECC-V4-500+`)

> Углубление PLAN-bases / GZ-V4-19/20 / BS-3 относительно wave1–4 (50k/200, views/attachment/filter, CRM/HR/ACC/intranet): **tables per base=100**, **dep depth ≤10**, link fan-in, select options, cross-base forbid, formula cycle, bases CSV import — без повтора row/col insert и pipeline/org-chart.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-500 | Edge | Bases: tables in one base = **100** (лимит) | 201; tables.length=100 | API | P0 | PLAN-bases Q3 | Future |
| ECC-V4-501 | Edge | Bases: add table **101** in same base | 400/422; count остаётся 100 | API | P0 | tables/base | Future |
| ECC-V4-502 | Edge | Bases: delete 1 of 100 → recreate to 100 | delete OK; re-add OK; 101-я снова 400 | API | P1 | table lifecycle | Future |
| ECC-V4-503 | Edge | Bases per instance = policy max B | OK at B; B+1 → 400 | API | P1 | bases/instance | Future |
| ECC-V4-504 | Edge | Relation links per cell = max L | OK at L; L+1 → 400; HubRecordLink count exact | API | P0 | BS-3 relations | Future |
| ECC-V4-505 | Edge | Relation links per cell = **0** (clear all) | OK; reverse links cleared; no orphan | API | P1 | unlink all | Future |
| ECC-V4-506 | Edge | Lookup/rollup/formula dependency chain depth = **10** | recalc OK | API | P0 | PLAN-bases R2 depth≤10 | Future |
| ECC-V4-507 | Edge | Dependency chain depth = **11** | 400 on save formula/lookup; graph unchanged | API | P0 | dep max | Future |
| ECC-V4-508 | Edge | Formula/lookup creates cycle A→B→A | 400; cycle detected; no partial graph | API | P0 | HubFormulaDep | Future |
| ECC-V4-509 | Edge | Cross-**base** relation create (table in base X → base Y) | 400/422; MVP only intra-base | API | P0 | PLAN-bases cross-base forbid | Future |
| ECC-V4-510 | Edge | Select / multi-select: options = max O | OK; O+1 → 400 | API | P1 | col options | Future |
| ECC-V4-511 | Edge | Select options = **0** / empty label | 400; column not created | API | P2 | empty options | TODO |
| ECC-V4-512 | Edge | Rollup on empty linked set (0 links) | returns 0/null per agg; no 500 | API | P1 | rollup empty | Future |
| ECC-V4-513 | Edge | Soft-delete linked target row → lookup | `#REF!` / null per policy; source row intact | API | P0 | BS-3 integrity | Future |
| ECC-V4-514 | Edge | Bases CSV import: rows = max I | all imported; I+1 → reject whole batch | API | P1 | BS-5 imex | Future |
| ECC-V4-515 | Edge | Bases CSV import: **0** data rows (header only) | 400 or 0 imported + warning; no phantom row | API | P2 | empty import | TODO |
| ECC-V4-516 | Edge | Intranet poll: change vote after first ballot | 422; aggregates unchanged (US-INTR-01) | API | P1 | poll immutable opts | Future |
| ECC-V4-517 | Edge | Marketplace module: enable while GZ-V4-21 freeze ON | 403/501; no install side-effect | API | P2 | marketplace freeze | TODO |

**Подсчёт §1:** 18

---

## 2. AI — max_tokens / top_p / stop · system prompt · OpenAI-compat · TTFB (`ECC-AI-500+`)

> Углубление US-F04/F05/F06 / GZ-SEC-22 относительно wave1–4 (budget, VK TTL, content length, embedding/RAG, temperature, provider timeout): **max_tokens, top_p, stop[], system prompt size, OpenAI-compat body, TTFB first-token, penalties** — без remaining=0/1, без embedding batch B, без temperature 0.0 duplicate as sole focus.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-500 | Edge | `max_tokens` = policy max M | accept; completion ≤ M | API | P0 | gen params | Future |
| ECC-AI-501 | Edge | `max_tokens` = M+1 / **0** / negative | 400; provider not called | API | P0 | max_tokens bound | Future |
| ECC-AI-502 | Edge | `top_p` = 0.0 vs 1.0 (bounds) | OK at both; &lt;0 or &gt;1 → 400 | API | P1 | nucleus | Future |
| ECC-AI-503 | Edge | `stop` sequences count = max S | OK; S+1 → 400 | API | P1 | stop[] | Future |
| ECC-AI-504 | Edge | Single stop string length = max / max+1 | OK / 400 | API | P2 | stop length | TODO |
| ECC-AI-505 | Edge | `frequency_penalty` / `presence_penalty` = policy min/max | OK at bounds; outside → 400 | API | P2 | penalties | TODO |
| ECC-AI-506 | Edge | System prompt size = max runes | OK; max+1 → 400; no provider call | API | P0 | system prompt | Future |
| ECC-AI-507 | Edge | System prompt empty + user prompt present | OK or 400 per ADR; documented | API | P2 | empty system | TODO |
| ECC-AI-508 | Edge | OpenAI-compat `POST /v1/chat/completions`: body = gateway max | 200; +1 → 413 | API | P1 | US-F06 | Future |
| ECC-AI-509 | Edge | OpenAI-compat: `stream=true` with max_tokens=1 | one chunk then done; usage counted | API | P1 | stream min | Future |
| ECC-AI-510 | Edge | TTFB first-token timeout = Tf: token at Tf−1ms vs Tf+1ms | −1 OK; +1 abort; reservation released | API | P0 | TTFB ≠ total timeout | Future |
| ECC-AI-511 | Edge | Seed / deterministic flag at boundary (supported model) | same seed → same hash prefix OR 501 if unsupported | API | P2 | seed | TODO |
| ECC-AI-512 | Edge | Multimodal image input bytes = max / max+1 | OK / 413; SSRF on image URL still blocked | API | P1 | vision input | Future |
| ECC-AI-513 | Edge | Content-filter category score = threshold exactly | block or allow per ADR; audit reason code; no raw prompt in log | API | P0 | US-F05 | Future |

**Подсчёт §2:** 14

---

## 3. Future Wiki — issue-link · optimistic version · base64 · search page (`ECC-WIKI-500+`)

> Углубление US-WIKI-03/04/05 относительно wave1–4 (depth, retention, body, trash 30d, lock, slug): **issue-link cardinality**, **optimistic version**, **base64 image reject**, **search page size**, draft autosave non-version — без 30d±1s trash и без depth 9→11.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-500 | Edge | Issue-links on page = max K | 201; links.length=K; reverse on issue OK | API | P0 | US-WIKI-04 | Future |
| ECC-WIKI-501 | Edge | Issue-link K+1 on same page | 400; no partial link | API | P0 | link max | Future |
| ECC-WIKI-502 | Edge | Issue-link to non-existent / inaccessible issue | 404/403; no dangling | API | P0 | link integrity | Future |
| ECC-WIKI-503 | Edge | Issue-links = **0** (unlink all) | OK; issue side cleared | API | P1 | unlink | Future |
| ECC-WIKI-504 | Edge | Optimistic lock: PUT with `version=N` (current) | 200; version=N+1 | API | P0 | NFR-24 | Future |
| ECC-WIKI-505 | Edge | Optimistic lock: PUT with `version=N−1` / stale | 409; body unchanged | API | P0 | version conflict | Future |
| ECC-WIKI-506 | Edge | Optimistic lock: omitted version field | 400; not silent overwrite | API | P1 | version required | Future |
| ECC-WIKI-507 | Edge | Markdown image as base64 data-URI | 400; must use presigned storage | API | P0 | US-WIKI-01 | Future |
| ECC-WIKI-508 | Edge | Autosave draft keystrokes: version history count | history not incremented; only explicit save versions | API | P1 | US-WIKI-03 edge | Future |
| ECC-WIKI-509 | Edge | Search Ctrl+K: page size = max P hits | ≤P; cursor for next; no full dump | API | P1 | US-WIKI-05 | Future |
| ECC-WIKI-510 | Edge | Search page size = P+1 requested | clamped to P or 400; NF-03 intact | API | P1 | page clamp | Future |
| ECC-WIKI-511 | Edge | Move page across spaces: target space write denied | 403; source tree unchanged | API | P0 | GZ-SEC-07/11 | Future |

**Подсчёт §3:** 12

---

## 4. Future SSO — PKCE / nonce / redirect_uri · LDAP DN · issuer (`ECC-SSO-500+`)

> Углубление ADR-AUTH-SSO относительно wave1–4 (toggle, JWT TTL, rate 3/10m, JWKS cache, claims length, break-glass cap): **PKCE verifier length**, **nonce replay**, **redirect_uri exact**, **issuer mismatch**, **LDAP DN/nest depth** — без JWKS TTL±1s и без refresh-family reuse (wave4).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-500 | Edge | PKCE `code_verifier` length = 43 (min RFC) | token exchange OK | API | P0 | OIDC PKCE | Future |
| ECC-SSO-501 | Edge | PKCE `code_verifier` length = 42 / 129 | 400; no tokens | API | P0 | PKCE bounds | Future |
| ECC-SSO-502 | Edge | PKCE `code_verifier` length = 128 (max RFC) | OK | API | P1 | PKCE max | Future |
| ECC-SSO-503 | Edge | `code_challenge` method = plain when S256 required | 400/403; reject | API | P0 | S256 only | Future |
| ECC-SSO-504 | Edge | OIDC `nonce` replay (same nonce twice) | second callback reject; no session | API | P0 | nonce | Future |
| ECC-SSO-505 | Edge | `redirect_uri` exact match vs trailing-slash variant | only registered exact OK; variant 400 | API | P0 | redirect_uri | Future |
| ECC-SSO-506 | Edge | `redirect_uri` open-redirect to external host | reject; never 302 off-origin | API | P0 | open redirect | Future |
| ECC-SSO-507 | Edge | IdP `iss` claim ≠ configured issuer | 401; no JIT | API | P0 | issuer mismatch | Future |
| ECC-SSO-508 | Edge | LDAP group DN length = max / max+1 | map OK / reject; no truncate | API | P1 | LDAP DN | Future |
| ECC-SSO-509 | Edge | Nested AD group depth = max Nd | flatten/map OK; Nd+1 ignored or 400 per ADR | API | P1 | nested groups | Future |
| ECC-SSO-510 | Edge | Auth `state` parameter empty / reused | reject CSRF; no login | API | P0 | state | Future |
| ECC-SSO-511 | Edge | MFA / TOTP enrollment attempt (out of v1 scope) | 404/501; OIDC+LDAP only | API | P2 | GZ-SEC no MFA | TODO |

**Подсчёт §4:** 12

---

## 5. Future MCP — run wall-clock · write-deny · scope∩tool · confirm double (`ECC-MCP-500+`)

> Углубление GZ-SEC-09/20/23 относительно wave1–4 (tool count, PAT TTL, confirm window, autonomy level, tool timeout Tt, concurrent R): **agent run wall-clock**, **read-heavy write deny**, **PAT scope ∩ tool**, **confirm double-submit**, **result page** — без Tt±1ms tool kill и без autonomy L/L+1.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-500 | Edge | Agent run wall-clock = max Wr: finishes Wr−1s | OK complete | API | P0 | run TTL | Future |
| ECC-MCP-501 | Edge | Agent run still active at Wr+1s | force-stop; no further tools; audit | API | P0 | run kill | Future |
| ECC-MCP-502 | Edge | Read-heavy v1 toolset: invoke write/mutating tool | 403; no side-effect; audit | API | P0 | GZ-SEC-23 curated | Future |
| ECC-MCP-503 | Edge | PAT scopes ∩ required tool scope = empty | 403 before invoke | API | P0 | scope intersection | Future |
| ECC-MCP-504 | Edge | PAT scopes ∩ tool = exactly one required scope | invoke OK | API | P1 | min scope | Future |
| ECC-MCP-505 | Edge | Human confirm: double-submit same confirm_id | first applies; second 409/idempotent no double tool | API | P0 | confirm double | Future |
| ECC-MCP-506 | Edge | Human confirm: confirm_id for other user's step | 403; no escalate | API | P0 | confirm ACL | Future |
| ECC-MCP-507 | Edge | Tool result rows = max page Pr | 200 paged; cursor for rest | API | P1 | result page | Future |
| ECC-MCP-508 | Edge | Tool result request Pr+1 | clamp or 400; no unbounded dump | API | P1 | page clamp | Future |
| ECC-MCP-509 | Edge | Tool args: required field null vs omitted | 400 schema; distinguish in error | API | P1 | args schema | Future |
| ECC-MCP-510 | Edge | Agent audit record size = max meta bytes | store OK; raw tool I/O not dumped beyond redaction | Unit | P0 | GZ-SEC-21 | Future |
| ECC-MCP-511 | Edge | Webhook/callback URL from tool config = private IP | ValidateOutboundURL deny; NF-07 | API | P0 | SSRF | Future |

**Подсчёт §5:** 12

---

## 6. Future Microservices — outbox batch · shadow % · upstream health · expand-contract (`ECC-MS-500+`)

> Углубление Track P / GZ-ARCH-16/17 относительно wave1–4 (schema_version, route-flip %, payload/ack/DLQ/retention): **outbox batch size**, **shadow compare %**, **upstream health fail**, **expand-contract NOT NULL**, **pending ack limit**, **poison quarantine TTL** — без max_deliver D и без 0/100% flip.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-500 | Edge | Outbox flush batch size = max Ob | all Ob published; ack OK | Infra | P0 | outbox | Future |
| ECC-MS-501 | Edge | Outbox batch = Ob+1 pending | split ticks; no lost rows; order preserved per key | Infra | P0 | batch split | Future |
| ECC-MS-502 | Edge | Outbox batch = **0** pending | no-op tick; no error storm | Infra | P2 | empty outbox | TODO |
| ECC-MS-503 | Edge | Shadow dual-run compare sample = **0%** | no compare overhead; primary only | Infra | P1 | shadow off | Future |
| ECC-MS-504 | Edge | Shadow sample = **100%** | every req compared; mismatch metric; primary wins | Infra | P0 | shadow full | Future |
| ECC-MS-505 | Edge | Shadow sample = 101% / negative | config reject; previous kept | Infra | P1 | invalid % | Future |
| ECC-MS-506 | Edge | Caddy upstream new service: health = unhealthy at flip&gt;0 | fail-closed to monolith OR auto weight 0 | Infra | P0 | upstream health | Future |
| ECC-MS-507 | Edge | Expand-contract: add NOT NULL col without default on live dual-run | migrate blocked / expand phase requires nullable first | Infra | P0 | Atlas expand | Future |
| ECC-MS-508 | Edge | Expand-contract: drop column while old consumer still reads it | contract phase only after consumer bump; else fail | Infra | P0 | contract | Future |
| ECC-MS-509 | Edge | Consumer pending acks = max Pa | Pa OK; Pa+1 backpressure/nak | Contract | P1 | pending limit | Future |
| ECC-MS-510 | Edge | Poison message quarantine TTL = Q: replay at Q−1s vs Q+1s | −1 inspectable; +1 purged | Infra | P1 | poison TTL | Future |
| ECC-MS-511 | Edge | NetworkPolicy: egress DNS to unexpected host from hub-ai | deny; BYOK/SSRF path intact | Infra | P0 | M3 isolation | Future |

**Подсчёт §6:** 12

---

## 7. Future Redesign a11y — aria-live · focus order · name/role · forced-colors · errors (`ECC-UX-500+`)

> Углубление WCAG 2.2 / NFR-16 относительно wave1–4 (trap, 24px, contrast, motion, FNO, zoom/reflow/skip/text-spacing): **aria-live storm**, **focus order 2.4.3**, **4.1.2 name/role/value**, **forced-colors**, **3.3.1 error ID**, **lang** — без 200/400% zoom и без 1.4.12 bookmarklet duplicate.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-500 | Edge | `aria-live="polite"`: N updates within 1s | SR coalesces / no unreadable storm; UI still correct | E2E | P1 | live regions | Future |
| ECC-UX-501 | Edge | `aria-live="assertive"` on non-critical toast | FAIL review / demote to polite | E2E | P1 | assertive misuse | Future |
| ECC-UX-502 | Edge | Focus order: Tab through modal then page — DOM order vs visual | matches meaningful sequence (2.4.3); no focus sink | E2E | P0 | WCAG 2.4.3 | Future |
| ECC-UX-503 | Edge | Focus order: skip hidden `display:none` controls | not focused; no dead Tab stops | E2E | P0 | hidden focus | Future |
| ECC-UX-504 | Edge | Icon-only button: accessible name empty | FAIL 4.1.2; needs aria-label/sr-only | E2E | P0 | name/role/value | Future |
| ECC-UX-505 | Edge | Icon-only button: aria-label length = max policy | OK; empty label FAIL | E2E | P1 | label bound | Future |
| ECC-UX-506 | Edge | Windows `forced-colors: active` / high-contrast | borders/focus visible; no information by color alone | E2E | P1 | forced-colors | Future |
| ECC-UX-507 | Edge | Form error: submit with empty required | error text associated (aria-describedby); focus moves to first error (3.3.1) | E2E | P0 | WCAG 3.3.1 | Future |
| ECC-UX-508 | Edge | Form error: 0 errors after fix + resubmit | error region cleared; no stale assertive announce | E2E | P1 | error clear | Future |
| ECC-UX-509 | Edge | `html lang` / page lang empty vs valid BCP47 | empty FAIL axe; valid PASS | E2E | P1 | 3.1.1 language | Future |
| ECC-UX-510 | Edge | Status message without role=status/aria-live | FAIL 4.1.3 Status Messages | E2E | P2 | 4.1.3 | TODO |
| ECC-UX-511 | Edge | Touch target AAA 44×44 optional vs AA 24×24 | AA gate still 24; 44 tracked as Should backlog | E2E | P2 | 2.5.5 vs 2.5.8 | TODO |

**Подсчёт §7:** 12

---

## Счётчик

| Домен | Prefix | Edge rows | Gap |
|-------|--------|-----------|-----|
| v4 Bases tables/relations/select | `ECC-V4-500+` | **18** | Future/TODO |
| AI max_tokens/top_p/OpenAI-compat/TTFB | `ECC-AI-500+` | **14** | Future/TODO |
| Wiki issue-link · version · base64 · search | `ECC-WIKI-500+` | **12** | Future/TODO |
| SSO PKCE/nonce/redirect/issuer/LDAP DN | `ECC-SSO-500+` | **12** | Future/TODO |
| MCP run wall-clock · write-deny · scope∩ | `ECC-MCP-500+` | **12** | Future/TODO |
| MS outbox · shadow % · health · expand | `ECC-MS-500+` | **12** | Future/TODO |
| Redesign aria-live · focus · 4.1.2 · errors | `ECC-UX-500+` | **12** | Future/TODO |
| **Итого Edge** | | **92** | только Edge; Happy/Neg/Corner не включались |

**Минимум 55:** выполнен (92 ≥ 55).  
**Пересечение `*100+`…`*400+`:** нет (все ID ≥500; фокусы не копируют wave1–4).  
**Основной каталог:** не изменялся.  
**Коммит:** не создавался.

### Handoff

- Далее: `qa-manual-ver` на полноту/техники/трассируемость.  
- После `verified: true` — выборочная автоматизация границ → `qa-auto1` (API/Unit/Infra) / `qa-auto2` (E2E a11y aria-live/focus/4.1.2).  
- Не мержить в `EDGE-CORNER-CATALOG.md` без отдельной команды на каталогизацию.
