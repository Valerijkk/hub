# Edge expansions wave3 — v4 + Future (AI / Wiki / SSO / MCP / MS / UX)

> **Тип:** только **Edge** (границы лимитов, TTL, empty/max, quota hard-stop, % flip).  
> **Wave:** 3 · ID namespace `*300+` — **не пересекает** wave1 `05-V4-AI-FUTURE.md` (`*100+`) и wave2 `E05-V4-AI-FUTURE.md` (`*200+`).  
> **Не трогает** основной каталог `.planning/EDGE-CORNER-CATALOG.md` — foresight backlog для будущих `TC-*`.  
> **Gap:** только `Future` | `TODO`.  
> **Фокус wave3:** bases **4 views / attachment / filter predicates** · AI **virtual-key TTL + content-policy max** · wiki **version retention + body size + restrict cardinality** · SSO **auth rate-limit 3/10m + JIT email bounds** · MCP **PAT TTL + confirm window** · MS **route-flip % (0→100)** · a11y **contrast 4.5:1/3:1 + reduced-motion + Focus Not Obscured**.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` (US-INTR / US-WIKI / US-F0x) · PLAN-bases (GZ-V4-19/20: 4 views) · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · WCAG 2.2 AA (1.4.3 Contrast, 2.3.3 / prefers-reduced-motion, 2.4.11 Focus Not Obscured) · NFR-05 (3/10m).  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** из этой сессии без явной команды.

Формат строк = каталог ECC: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 Bases — 4 views / attachment / filter predicates (`ECC-V4-300+`)

> Углубление PLAN-bases / GZ-V4-19/20 относительно `ECC-V4-123..127` (wave1) и `ECC-V4-200..217` (wave2 rows/cols): **ровно 4 view-типа**, attachment byte/count, filter predicate count, view share inherit — без повтора 50k/200 insert.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-300 | Edge | Bases: создать ровно **4** named views (grid/kanban/calendar/gallery per PLAN) | 201; views.length=4 | API | P0 | GZ-V4-19/20 | Future |
| ECC-V4-301 | Edge | Bases: add **5-й** view того же или нового типа | 400/422; views остаются 4 | API | P0 | PLAN-bases 4 views | Future |
| ECC-V4-302 | Edge | Bases: delete 1 of 4 → recreate to 4 | delete OK; re-add OK; 5-я снова 400 | API | P1 | view lifecycle | Future |
| ECC-V4-303 | Edge | Bases: rename/reorder views at count=4 (без add) | 200; count остаётся 4 | API | P1 | view mutate | Future |
| ECC-V4-304 | Edge | Bases: attachment field — file size = policy max bytes | 201/presign OK; stored | API | P0 | PLAN-bases attachment | Future |
| ECC-V4-305 | Edge | Bases: attachment size = max+1 | 400/413; cell без файла | API | P0 | NF-09 / policy | Future |
| ECC-V4-306 | Edge | Bases: attachment count per cell = max M | OK at M; M+1 → 400 | API | P1 | attachment count | Future |
| ECC-V4-307 | Edge | Bases: attachment MIME empty / unknown at allow-list boundary | reject unknown; empty 400 | API | P2 | MIME allow-list | TODO |
| ECC-V4-308 | Edge | Bases: filter view — ровно **max predicates** K | save OK; query runs | API | P0 | filter limit | Future |
| ECC-V4-309 | Edge | Bases: filter predicates = K+1 | 400; view config not saved | API | P0 | filter max | Future |
| ECC-V4-310 | Edge | Bases: filter predicates = **0** (empty filter) | OK = show all; not 400 | API | P2 | empty filter | TODO |
| ECC-V4-311 | Edge | Bases: sort keys = max S columns | OK; S+1 → 400 | API | P1 | sort limit | Future |
| ECC-V4-312 | Edge | Bases: view inherits table ACL — member with table read, view manage denied | read OK; manage 403 | API | P0 | GZ-SEC-08/12 | Future |
| ECC-V4-313 | Edge | Bases: share-link backlog flag OFF — create public link | 404/403; no token issued | API | P1 | share-link backlog | Future |
| ECC-V4-314 | Edge | Bases: formula functions catalog — invoke function index 30 (last of 30) | OK; index 31 / unknown → 400 | API | P1 | PLAN-bases 30 fn | Future |
| ECC-V4-315 | Edge | Bases: realtime BS-5 — subscribe at max concurrent watchers W | W OK; W+1 reject/queue | API/WS | P2 | BS-5 realtime | TODO |

**Подсчёт §1:** 16

---

## 2. AI — virtual-key TTL + content-policy max (`ECC-AI-300+`)

> Углубление US-F05 / GZ-SEC-22 относительно `ECC-AI-100..112` и `ECC-AI-200..214` (token budget): **TTL virtual key**, content-policy length, streaming chunk size, catalog empty→1, fallback chain length — без повтора remaining=0/1 token.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-300 | Edge | Virtual key: now = expires_at − 1s | request OK | API | P0 | virtual keys | Future |
| ECC-AI-301 | Edge | Virtual key: now = expires_at + 1s | 401; no provider call; usage unchanged | API | P0 | VK TTL | Future |
| ECC-AI-302 | Edge | Virtual key TTL = **0** / negative on create | 400; key not stored | API | P1 | TTL validate | Future |
| ECC-AI-303 | Edge | Virtual key TTL = policy max days | create OK; max+1d → 400 | API | P1 | TTL max | Future |
| ECC-AI-304 | Edge | Content policy: prompt length = max runes | accept; audit without raw dump | API | P0 | US-F05 | Future |
| ECC-AI-305 | Edge | Content policy: prompt length = max+1 | 400; provider not called | API | P0 | content max | Future |
| ECC-AI-306 | Edge | Content policy: empty prompt / whitespace-only | 400 | API | P2 | empty prompt | TODO |
| ECC-AI-307 | Edge | Streaming: chunk size = max bytes per event | OK; max+1 truncated or split per ADR | API | P1 | stream chunk | Future |
| ECC-AI-308 | Edge | Model catalog: **0** enabled → enable exactly **1** | empty UI then single model selectable | API/E2E | P1 | US-F04 | Future |
| ECC-AI-309 | Edge | Fallback chain length = max F providers | OK; F+1 → 400 config | API | P1 | LiteLLM pattern | Future |
| ECC-AI-310 | Edge | Fallback: all F providers fail at boundary | 502/503; usage not over-billed; audit | API | P0 | fallback exhaust | Future |
| ECC-AI-311 | Edge | Concurrent AI slots = limit C: request C and C+1 | C OK; C+1 429/queue; no silent drop | API | P1 | ECC-AI-022 | TODO |
| ECC-AI-312 | Edge | BYOK ciphertext length = AES-GCM envelope max | store OK; max+1 reject | Unit | P1 | GZ-SEC-22 | Future |
| ECC-AI-313 | Edge | OpenAI-compat path US-F06: max body = gateway limit | OK at limit; +1 → 413 | API | P2 | US-F06 backlog | TODO |

**Подсчёт §2:** 14

---

## 3. Future Wiki — version retention + body + restrict cardinality (`ECC-WIKI-300+`)

> Углубление US-WIKI-01..05 относительно depth (`*100`/`*200`): **version retention R**, body max bytes, restrict allow-list size, search query length — без create depth 9→10→11.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-300 | Edge | Version history: saves = retention **R** | all R kept; list length=R | API | P0 | US-WIKI-03 | Future |
| ECC-WIKI-301 | Edge | Version history: save R+1 | oldest purged OR 422; length ≤R | API | P0 | retention | Future |
| ECC-WIKI-302 | Edge | Restore version at index R (oldest kept) | 200; body matches that version | API | P0 | restore | Future |
| ECC-WIKI-303 | Edge | Restore version index R+1 / purged | 404; current unchanged | API | P1 | purged | Future |
| ECC-WIKI-304 | Edge | Page body size = max bytes | 200/201 OK | API | P0 | US-WIKI-01 | Future |
| ECC-WIKI-305 | Edge | Page body size = max+1 | 400; previous version intact | API | P0 | body max | Future |
| ECC-WIKI-306 | Edge | Page body empty string vs null | policy: 400 or allow empty stub; documented | API | P2 | empty body | TODO |
| ECC-WIKI-307 | Edge | Restrict allow-list size = max U users | OK; U+1 → 400 | API | P0 | US-WIKI-02 | Future |
| ECC-WIKI-308 | Edge | Restrict allow-list size = **1** (min useful) | OK; only that user + admin | API | P1 | restrict min | Future |
| ECC-WIKI-309 | Edge | Search query length = max Q chars | 200 scoped results | API | P1 | US-WIKI-05 | Future |
| ECC-WIKI-310 | Edge | Search query length = max+1 | 400; no full dump | API | P1 | NF-03 | Future |
| ECC-WIKI-311 | Edge | Space members = 0 (empty space) list pages | 200 []; no cross-space leak | API | P2 | empty space | TODO |

**Подсчёт §3:** 12

---

## 4. Future SSO — auth rate-limit 3/10m + JIT email bounds (`ECC-SSO-300+`)

> ADR-AUTH-SSO + NFR-05: границы **3 попытки / 10 минут** на OIDC/Magic, JIT email length, LDAP batch page size — без повтора JWT 15m / refresh 30d exact TTL (`*200`).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-300 | Edge | OIDC login attempts = **3** within 10m window | 3-я ещё обрабатывается (success or IdP error) | API | P0 | NFR-05 | Future |
| ECC-SSO-301 | Edge | OIDC login attempt = **4-я** within same 10m | 429; IdP not called / short-circuit | API | P0 | rate-limit | Future |
| ECC-SSO-302 | Edge | Attempt at window start T and T+10m−1s vs T+10m+1s | inside: counted; after: counter reset | Unit | P0 | 10m window | Future |
| ECC-SSO-303 | Edge | Magic Link send = 3/10m then 4th (break-glass founder) | 4th 429 same as ordinary; no god-mode bypass | API | P0 | break-glass RL | Future |
| ECC-SSO-304 | Edge | Org `both`: Magic + OIDC share vs split counters at limit | per ADR; documented; neither exceeds 3/10m effective abuse | API | P1 | GZ-SEC-04/08 | Future |
| ECC-SSO-305 | Edge | JIT email length = max RFC/policy | user created | API | P1 | JIT | Future |
| ECC-SSO-306 | Edge | JIT email length = max+1 / empty | 400; no user row | API | P1 | email bound | Future |
| ECC-SSO-307 | Edge | LDAP sync page size = max P entries | page OK; P+1 rejected or split | API | P1 | LDAP-sync | Future |
| ECC-SSO-308 | Edge | LDAP sync: **0** entries in page (empty AD OU) | no-op OK; no mass deactivate false-positive | API | P0 | empty sync | Future |
| ECC-SSO-309 | Edge | IdP `state`/`nonce` TTL = N: callback at N−1s vs N+1s | −1 OK; +1 reject CSRF | API | P0 | OIDC state | Future |
| ECC-SSO-310 | Edge | Group→role map entries = max G | OK; G+1 → 400 | API | P1 | GZ-SEC-10 | Future |
| ECC-SSO-311 | Edge | SAML path (out of v1): enable attempt | 404/501; OIDC-only until ADR | API | P2 | GZ-SEC-03 | TODO |

**Подсчёт §4:** 12

---

## 5. Future MCP — PAT TTL + confirm window (`ECC-MCP-300+`)

> GZ-SEC-09/20/23: границы **PAT TTL**, human-confirm timeout, tool arg size, agent∩user empty intersection — без повтора toolset count N / autonomy step M (`*200`).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-300 | Edge | PAT: now = expires_at − 1s | tool invoke OK | API | P0 | PAT TTL | Future |
| ECC-MCP-301 | Edge | PAT: now = expires_at + 1s | 401; no side-effect; audit | API | P0 | PAT expiry | Future |
| ECC-MCP-302 | Edge | PAT TTL on create = 0 / negative | 400; token not issued | API | P1 | TTL validate | Future |
| ECC-MCP-303 | Edge | PAT TTL = policy max | OK; max+1 → 400 | API | P1 | TTL max | Future |
| ECC-MCP-304 | Edge | Autonomy confirm window = W: human confirms at W−1s | step continues | API | P0 | GZ-SEC-09/20 | Future |
| ECC-MCP-305 | Edge | Confirm at W+1s (timeout) | step cancelled; no tool side-effect | API | P0 | confirm TTL | Future |
| ECC-MCP-306 | Edge | Tool args JSON size = max | OK; max+1 → 400 | API | P1 | arg size | Future |
| ECC-MCP-307 | Edge | Tool args empty object `{}` for required-arg tool | 400; schema fail | API | P1 | empty args | Future |
| ECC-MCP-308 | Edge | agent ∩ user rights = **empty** (user lost all projects) | all tools 403; run stops | API | P0 | agent-principal | Future |
| ECC-MCP-309 | Edge | Audit fields: user+agent+task-id all present at max UUID length | write OK; truncate forbidden | Unit | P1 | GZ-SEC-10/21 | Future |
| ECC-MCP-310 | Edge | Coverage gate MCP package: line% = 79.9 vs 80.0 | &lt;80 fail CI; ≥80 pass | Infra | P1 | GZ-QA-01/02 | TODO |
| ECC-MCP-311 | Edge | Read-heavy tool: result rows = max page | 200 paged; max+1 not returned without cursor | API | P2 | MCP read | TODO |

**Подсчёт §5:** 12

---

## 6. Future Microservices — route-flip % (`ECC-MS-300+`)

> Track P / GZ-ARCH-16/17: границы **Caddy traffic %** 0/1/50/99/100 + dual-run lag SLA — без повтора schema_version v1/v2 matrix (`*200`).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-300 | Edge | Route-flip weight = **0%** new service | 100% monolith; new idle healthy | Infra | P0 | Caddy flip | Future |
| ECC-MS-301 | Edge | Route-flip weight = **1%** (min canary) | ~1% to new; p95 gate; errors fail-closed | Infra/Perf | P0 | canary | Future |
| ECC-MS-302 | Edge | Route-flip weight = **50%** | split; idempotent consumers; no double side-effect | Infra | P0 | GZ-ARCH-16/17 | Future |
| ECC-MS-303 | Edge | Route-flip weight = **99%** | almost all new; monolith still serves 1% | Infra | P1 | near-cutover | Future |
| ECC-MS-304 | Edge | Route-flip weight = **100%** | monolith path dark; new only; rollback to 0 documented | Infra | P0 | cutover | Future |
| ECC-MS-305 | Edge | Route-flip weight = **101%** / negative | config reject; previous weight kept | Infra | P1 | invalid % | Future |
| ECC-MS-306 | Edge | Dual-run lag: consumer lag = SLA max ms | OK; metric green | Infra | P1 | lag SLA | Future |
| ECC-MS-307 | Edge | Dual-run lag = SLA max+1 | alert; flip blocked or auto-rollback per runbook | Infra | P0 | lag breach | Future |
| ECC-MS-308 | Edge | Schema expand: add nullable col on empty service schema | migrate OK; app starts | API/Infra | P1 | GZ-ARCH-08 | Future |
| ECC-MS-309 | Edge | NetworkPolicy: hub-ai pod read hub-core secret at boundary | deny; BYOK isolated | Infra | P0 | M3 BYOK | Future |

**Подсчёт §6:** 10

---

## 7. Future Redesign a11y — contrast / reduced-motion / Focus Not Obscured (`ECC-UX-300+`)

> WCAG 2.2: **1.4.3 Contrast** 4.5:1 text / 3:1 UI, **prefers-reduced-motion**, **2.4.11 Focus Not Obscured**, viewport **320px** — без повтора 24×24 target size (`*200`).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-300 | Edge | Body text contrast = **4.5:1** exactly (light theme) | PASS 1.4.3 | E2E | P0 | WCAG 1.4.3 | Future |
| ECC-UX-301 | Edge | Body text contrast = **4.49:1** | FAIL a11y gate | E2E | P0 | contrast floor | Future |
| ECC-UX-302 | Edge | Non-text UI (icons/borders) contrast = **3:1** exactly | PASS 1.4.11 | E2E | P0 | 3:1 UI | Future |
| ECC-UX-303 | Edge | Non-text UI contrast = **2.9:1** | FAIL | E2E | P0 | UI contrast | Future |
| ECC-UX-304 | Edge | Dark theme: same 4.5:1 / 3:1 floors | PASS both themes | E2E | P0 | GZ-QA-06/07 | Future |
| ECC-UX-305 | Edge | `prefers-reduced-motion: reduce` — decorative animation duration | 0 / instant; essential state change remains | E2E | P1 | 2.3.3 / redesign | Future |
| ECC-UX-306 | Edge | `prefers-reduced-motion: no-preference` — motion within policy max | OK; no unbounded loop | E2E | P2 | motion budget | TODO |
| ECC-UX-307 | Edge | Sticky header height H: focused control scrolled under header by 1px | FAIL 2.4.11 OR auto-scroll into view | E2E | P0 | Focus Not Obscured | Future |
| ECC-UX-308 | Edge | Focused control fully below sticky (clearance ≥1px) | PASS 2.4.11 | E2E | P0 | 2.4.11 | Future |
| ECC-UX-309 | Edge | Viewport width = **320px**: primary controls usable, no horizontal clip of focus ring | PASS; NFR-16 | E2E | P1 | 320px | Future |
| ECC-UX-310 | Edge | Viewport width = **319px** (below support floor if documented) | graceful degrade or unsupported banner; no JS crash | E2E | P2 | min viewport | TODO |

**Подсчёт §7:** 11

---

## Счётчик

| Домен | Prefix | Edge rows | Gap |
|-------|--------|-----------|-----|
| v4 Bases 4 views / attachment / filter | `ECC-V4-300+` | **16** | Future/TODO |
| AI virtual-key TTL + content-policy | `ECC-AI-300+` | **14** | Future/TODO |
| Wiki version retention + body + restrict | `ECC-WIKI-300+` | **12** | Future/TODO |
| SSO rate-limit 3/10m + JIT bounds | `ECC-SSO-300+` | **12** | Future/TODO |
| MCP PAT TTL + confirm window | `ECC-MCP-300+` | **12** | Future/TODO |
| MS route-flip % | `ECC-MS-300+` | **10** | Future/TODO |
| Redesign contrast / motion / FNO | `ECC-UX-300+` | **11** | Future/TODO |
| **Итого Edge** | | **87** | только Edge; Happy/Neg/Corner не включались |

**Минимум 55:** выполнен (87 ≥ 55).  
**Пересечение `*100+` / `*200+`:** нет (все ID ≥300; фокус-сценарии не копируют wave1/wave2).  
**Основной каталог:** не изменялся.  
**Коммит:** не создавался.

### Handoff

- Далее: `qa-manual-ver` на полноту/техники/трассируемость.  
- После `verified: true` — выборочная автоматизация границ → `qa-auto1` (API/Unit/Infra) / `qa-auto2` (E2E a11y contrast/motion/FNO).  
- Не мержить в `EDGE-CORNER-CATALOG.md` без отдельной команды на каталогизацию.
