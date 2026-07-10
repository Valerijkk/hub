# Edge expansions wave4 — v4 + Future (AI / Wiki / SSO / MCP / MS / UX)

> **Тип:** только **Edge** (границы лимитов, TTL, empty/max, quota hard-stop, retention).  
> **Wave:** 4 · ID namespace `*400+` — **не пересекает** wave1 (`*100+`), wave2 (`*200+`), wave3 (`*300+`).  
> **Не трогает** основной каталог `.planning/EDGE-CORNER-CATALOG.md` — foresight backlog для будущих `TC-*`.  
> **Gap:** только `Future` | `TODO`.  
> **Фокус wave4:** v4 **CRM pipeline/import · HR org-chart/accrual calendar · Accounting money/AR · Intranet coffee/onboarding M** · AI **embedding/RAG chunk · @hub-ai rate · provider timeout** · wiki **soft-delete 30d cascade · edit-lock TTL · slug/attachment** · SSO **JWKS cache TTL · claim bounds · break-glass daily cap · LDAP bind timeout** · MCP **autonomy levels · tool timeout · concurrent runs · revoke** · MS **NATS payload/ack/DLQ retries · JetStream retention** · a11y **zoom 200/400 · reflow · focus-visible · skip-link · text-spacing**.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` (US-INTR / US-WIKI / US-F0x / US-V0x) · PLAN-bases / GZ-V4-08..20 · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · WCAG 2.2 AA (1.4.4 Resize, 1.4.10 Reflow, 1.4.12 Text Spacing, 2.4.1 Bypass, 2.4.7 Focus Visible) · NFR-11 soft-delete 30d · NFR-05.  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** из этой сессии без явной команды.

Формат строк = каталог ECC: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 — CRM / HR / Accounting / Intranet bounds (`ECC-V4-400+`)

> Углубление GZ-V4-01..18 / US-INTR / US-V0x относительно wave1 (poll/leave/CSV/PDF/bases counts), wave2 (50k/200), wave3 (4 views/attachment/filter): **pipeline stages, deal money, org-chart depth, production calendar, AR aging buckets, coffee pool, onboarding M** — без повтора rows/cols/views.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-400 | Edge | CRM pipeline: stages count = policy max S | create OK; stages.length=S | API | P0 | GZ-V4-08/09 | Future |
| ECC-V4-401 | Edge | CRM pipeline: add stage S+1 | 400/422; pipeline unchanged | API | P0 | pipeline max | Future |
| ECC-V4-402 | Edge | CRM pipeline: stages = **1** (min useful) | OK; deal can sit in single stage | API | P1 | pipeline min | Future |
| ECC-V4-403 | Edge | CRM deal amount = **0** (zero-value deal) | 201 per policy OR 400; documented; no negative | API | P1 | US-V04 | Future |
| ECC-V4-404 | Edge | CRM deal amount = max money scale (e.g. 999…99.99) | OK; max+epsilon → 400; exact decimal | Unit | P0 | money scale | Future |
| ECC-V4-405 | Edge | CRM activity timeline: entries = max T per deal | OK at T; T+1 → 400 or oldest archived per policy | API | P1 | GZ-V4-08 activity | Future |
| ECC-V4-406 | Edge | CRM XLSX import: sheets = 1 used / extra sheets ignored | only first/mapped sheet; no silent multi-sheet merge | API | P1 | GZ-V4-10 | Future |
| ECC-V4-407 | Edge | CRM import: column map empty vs all required mapped | empty 400; full map OK | API | P1 | import map | Future |
| ECC-V4-408 | Edge | HR org-chart: manager chain depth = max D | OK; D+1 cycle/over-depth → 422 | API | P0 | GZ-V4-12; US-V02 | Future |
| ECC-V4-409 | Edge | HR org-chart: `manager_id` = self | 422; chart unchanged | API | P0 | self-manager | Future |
| ECC-V4-410 | Edge | HR grade field: value = last enum / unknown | OK / 400 | API | P1 | GZ-V4-15 grade | Future |
| ECC-V4-411 | Edge | HR production calendar RF: leave day = holiday | duration excludes holiday OR counts per policy; documented | API | P1 | GZ-V4-14 | Future |
| ECC-V4-412 | Edge | HR leave: half-day 0.5 on boundary of accrual | accept; balance −0.5 exact; no float drift | Unit | P0 | GZ-V4-15 | Future |
| ECC-V4-413 | Edge | Accounting journal: debit total = credit total at scale 0.01 | 201 balanced | API | P0 | GZ-V4-16 | Future |
| ECC-V4-414 | Edge | Accounting journal: imbalance = 0.01 | 400; entry not posted | API | P0 | balance check | Future |
| ECC-V4-415 | Edge | Accounting AR aging: bucket boundary day N vs N+1 | invoice lands in correct bucket; no off-by-one | Unit | P1 | GZ-V4-17 AR | Future |
| ECC-V4-416 | Edge | Accounting currency field present, multi-currency OFF | store code; convert ops 501/disabled | API | P1 | GZ-V4-17/18 | Future |
| ECC-V4-417 | Edge | Intranet onboarding checklist: items = **0** | 400 template OR empty progress 0/0 guarded (no div0) | API | P1 | US-INTR-02 | Future |
| ECC-V4-418 | Edge | Intranet onboarding: items = max M; complete M/M | progress 100%; M+1 item → 400 | API | P1 | checklist max | Future |
| ECC-V4-419 | Edge | Random Coffee: opt-in pool = **1** (odd singleton) | no pair / bye policy; no crash; no DM to self | API | P0 | US-INTR-05 | Future |
| ECC-V4-420 | Edge | Random Coffee: pool = **0** opt-in | no-op week; no phantom DM | API | P1 | empty pool | Future |
| ECC-V4-421 | Edge | Random Coffee: pool = max P participants | matching completes ≤ SLA; P+1 waitlist/reject | API | P1 | coffee max | Future |
| ECC-V4-422 | Edge | Intranet poll close: votes = 0 at close boundary | results empty OK; no 500 | API | P2 | US-INTR-01 | TODO |
| ECC-V4-423 | Edge | Soft-delete v4 entity: restore at day **30−1s** vs **30+1s** | −1 restore OK; +1 hard-gone / 404 | API | P0 | NFR-11 | Future |

**Подсчёт §1:** 24

---

## 2. AI — embedding/RAG · @hub-ai rate · provider timeout (`ECC-AI-400+`)

> Углубление US-F01/F05/F06 / GZ-SEC-22 относительно wave1–3 (budget tokens, VK TTL, content-policy length): **embedding batch, RAG chunk/overlap, mention rate, provider timeout, temperature, tool-schema depth** — без remaining=0/1 и без VK expiry ±1s.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-400 | Edge | Embedding batch size = max B texts | 200; vectors.length=B | API | P0 | embeddings | Future |
| ECC-AI-401 | Edge | Embedding batch size = B+1 | 400/413; no partial vectors persisted | API | P0 | batch max | Future |
| ECC-AI-402 | Edge | Embedding batch size = **0** / empty array | 400 | API | P2 | empty batch | TODO |
| ECC-AI-403 | Edge | RAG chunk size = max C tokens; overlap = max O | index OK; C+1 / O≥C → 400 | Unit | P1 | RAG bounds | Future |
| ECC-AI-404 | Edge | RAG chunk size = **1** (min) with overlap 0 | OK; retrieval returns chunk | Unit | P2 | chunk min | TODO |
| ECC-AI-405 | Edge | `@hub-ai` mentions in channel = rate R / window | R OK; R+1 muted/429; no provider storm | API | P0 | US-F01 | Future |
| ECC-AI-406 | Edge | `@hub-ai` in readonly channel at rate boundary | 403 before rate; no usage burn | API | P0 | channel ACL | Future |
| ECC-AI-407 | Edge | Provider HTTP timeout = T: response at T−1ms vs T+1ms | −1 OK; +1 abort; reservation released | API | P0 | provider timeout | Future |
| ECC-AI-408 | Edge | Provider timeout mid-stream after N chunks | abort; partial usage counted; client error | API | P0 | stream timeout | Future |
| ECC-AI-409 | Edge | Temperature = 0.0 (min) vs policy max | OK at bounds; max+ε → 400 | API | P1 | sampling | Future |
| ECC-AI-410 | Edge | Tool/function schema JSON depth = max D | accept; D+1 → 400 | Unit | P1 | tool schema | Future |
| ECC-AI-411 | Edge | Model id length = max / empty | OK / 400; catalog lookup fail-closed | API | P1 | US-F04 | Future |
| ECC-AI-412 | Edge | AES-GCM key rotation window: decrypt with previous key at TTL−1s | OK; after TTL+1s old key unusable | Unit | P0 | GZ-SEC-22 | Future |
| ECC-AI-413 | Edge | Audit log: response body redacted at max stored meta size | no raw completion in audit; meta ≤ max | API | P0 | US-F05 | Future |

**Подсчёт §2:** 14

---

## 3. Future Wiki — soft-delete 30d · edit-lock · slug/attachment (`ECC-WIKI-400+`)

> Углубление US-WIKI / GZ-SEC-07/11 / NFR-11 относительно wave1–3 (depth, retention R, body size, restrict U): **корзина 30д + cascade, edit-lock TTL, slug bounds, page attachment, handbook singleton** — без depth 9→11 и без R/R+1 version purge.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-400 | Edge | Soft-delete page: restore at trash age = 30d−1s | restore 200; tree intact | API | P0 | NFR-11 | Future |
| ECC-WIKI-401 | Edge | Soft-delete page: restore at trash age = 30d+1s | 404; hard-purged; no partial ghost | API | P0 | 30d purge | Future |
| ECC-WIKI-402 | Edge | Soft-delete parent with children: cascade at day 30 | children purged with parent; no orphan depth | API | P0 | GZ-SEC-17 wiki cascade | Future |
| ECC-WIKI-403 | Edge | Soft-delete space: pages count = 0 vs max | empty OK; max cascade within timeout/job | API | P1 | space delete | Future |
| ECC-WIKI-404 | Edge | Edit-lock TTL = L: second editor at L−1s | 409 locked; first keeps lock | API | P0 | concurrent edit | Future |
| ECC-WIKI-405 | Edge | Edit-lock: second editor at L+1s after expiry | lock acquired; first stale save → 409 version | API | P0 | lock expiry | Future |
| ECC-WIKI-406 | Edge | Slug length = max / empty / unicode NFC | OK at max; empty 400; unicode normalized unique | API | P1 | slug | Future |
| ECC-WIKI-407 | Edge | Slug duplicate in same space | 409/422; no overwrite | API | P1 | slug unique | Future |
| ECC-WIKI-408 | Edge | Page attachment size = policy max | OK; max+1 → 400/413 | API | P1 | page files | Future |
| ECC-WIKI-409 | Edge | Page attachments count = max A | OK; A+1 → 400 | API | P1 | attachment count | Future |
| ECC-WIKI-410 | Edge | Handbook space: create second org handbook | 409/422; singleton per GZ-V4-13 | API | P0 | handbook | Future |
| ECC-WIKI-411 | Edge | @mention users on page = max Mentions | OK; max+1 → 400; notify only allowed | API | P2 | mentions | TODO |

**Подсчёт §3:** 12

---

## 4. Future SSO — JWKS cache · claims · break-glass cap · LDAP bind (`ECC-SSO-400+`)

> Углубление ADR-AUTH-SSO / GZ-SEC-03..10 относительно wave1–3 (toggle, JWT/refresh TTL, rate 3/10m, JIT email length): **JWKS cache TTL, claim presence/size, break-glass daily cap, LDAP bind timeout, refresh-family rotation** — без iat±15m и без 3rd/4th attempt.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-400 | Edge | JWKS cache TTL = J: key rotate at J−1s still served from cache | verify OK with cached set | Unit | P0 | OIDC JWKS | Future |
| ECC-SSO-401 | Edge | JWKS after cache TTL+1s: fetch new; old kid rejected | 401 invalid kid; refresh JWKS | Unit | P0 | JWKS rotate | Future |
| ECC-SSO-402 | Edge | IdP JWKS endpoint timeout / empty keys | fail-closed login; no accept unsigned | API | P0 | JWKS fail-closed | Future |
| ECC-SSO-403 | Edge | OIDC claim `email` missing / null | 400/403; no JIT user | API | P0 | required claims | Future |
| ECC-SSO-404 | Edge | OIDC claim string length = max / max+1 | OK / reject; no DB truncate corruption | API | P1 | claim bounds | Future |
| ECC-SSO-405 | Edge | Break-glass Magic Link: founder sends = daily cap C | C OK; C+1 → 429; audit | API | P0 | break-glass cap | Future |
| ECC-SSO-406 | Edge | Break-glass: non-founder at cap boundary | always 403; cap counter not consumed | API | P0 | founder only | Future |
| ECC-SSO-407 | Edge | LDAP bind timeout = Tb: response Tb−1 vs Tb+1 | −1 auth path continues; +1 fail-closed | API | P0 | LDAP bind | Future |
| ECC-SSO-408 | Edge | LDAP sync deactivate batch = max U users / window | all revoked; U+1 split next tick; no partial silent skip without metric | API | P1 | LDAP-sync batch | Future |
| ECC-SSO-409 | Edge | Refresh-token family: reuse of rotated refresh | reuse detected → family revoke; 401 | API | P0 | refresh rotation | Future |
| ECC-SSO-410 | Edge | Org policy `both`: simultaneous Magic + OIDC sessions = max S devices | S OK; S+1 oldest revoked or 403 per ADR | API | P1 | session cap | Future |
| ECC-SSO-411 | Edge | Mapped system role claim = unknown value | default Member OR reject; never auto-Admin | API | P0 | GZ-SEC-10 | Future |

**Подсчёт §4:** 12

---

## 5. Future MCP — autonomy levels · tool timeout · concurrent runs · revoke (`ECC-MCP-400+`)

> Углубление GZ-SEC-09/20/21/23 относительно wave1–3 (tool count N, PAT TTL, confirm window W): **autonomy level enum, tool wall-clock timeout, max concurrent agent runs, immediate revoke, name length** — без N/N+1 tools и без PAT ±1s expiry.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-400 | Edge | Autonomy level = **0** (confirm every step) | each tool waits confirm; no auto chain | API | P0 | GZ-SEC-09/20 | Future |
| ECC-MCP-401 | Edge | Autonomy level = max L (policy) | L steps auto; L+1 needs confirm | API | P0 | autonomy max | Future |
| ECC-MCP-402 | Edge | Autonomy level = L+1 / negative on config | 400; config not saved | API | P1 | level validate | Future |
| ECC-MCP-403 | Edge | Tool wall-clock timeout = Tt: finishes at Tt−1ms | OK result | API | P0 | tool timeout | Future |
| ECC-MCP-404 | Edge | Tool still running at Tt+1ms | kill/cancel; 504/abort; no orphan side-effect (or compensating) | API | P0 | timeout kill | Future |
| ECC-MCP-405 | Edge | Concurrent agent runs per org = max R | R OK; R+1 → 429 | API | P0 | concurrency | Future |
| ECC-MCP-406 | Edge | Concurrent runs = **0** allowed (feature flag off) | all starts 403/503; documented | API | P2 | kill switch | TODO |
| ECC-MCP-407 | Edge | PAT revoke: invoke at revoke_at − 1ms vs +1ms | −1 OK; +1 401 immediate (no grace) | API | P0 | revoke | Future |
| ECC-MCP-408 | Edge | Tool name length = max / empty / unknown unicode | OK / 400 / 403 allow-list | API | P1 | tool name | Future |
| ECC-MCP-409 | Edge | Sandbox: tool attempts outbound URL not in allow-list | blocked; SSRF deny; audit | API | P0 | GZ-SEC-23 / NF-07 | Future |
| ECC-MCP-410 | Edge | Agent principal: user soft-deleted mid-run | subsequent tools 401/403; run halt | API | P0 | agent-principal | Future |
| ECC-MCP-411 | Edge | MCP coverage gate: package = 80.0% exactly | CI pass; 79.99 fail | Infra | P1 | GZ-QA-01/02 | TODO |

**Подсчёт §5:** 12

---

## 6. Future Microservices — NATS payload / ack / DLQ · JetStream retention (`ECC-MS-400+`)

> Углубление Track P / GZ-ARCH / GZ-QA-08/09 относительно wave1–3 (schema_version, route-flip %): **payload max, ack wait, max deliver/DLQ, JetStream retention, request-reply timeout** — без 0/1/50/99/100% flip и без v1/v2 version matrix.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-400 | Edge | NATS message payload size = max bytes | publish OK; consume OK | Contract | P0 | JetStream | Future |
| ECC-MS-401 | Edge | NATS payload size = max+1 | publish reject; no truncated poison | Contract | P0 | payload max | Future |
| ECC-MS-402 | Edge | NATS payload size = **0** (empty body) with required schema | reject; ack not success-path | Contract | P1 | empty payload | Future |
| ECC-MS-403 | Edge | Ack wait = Aw: consumer acks at Aw−1ms | message not redelivered | Contract | P0 | ack wait | Future |
| ECC-MS-404 | Edge | Ack at Aw+1ms / no ack | redelivery; max_deliver counted | Contract | P0 | redelivery | Future |
| ECC-MS-405 | Edge | max_deliver = D: fail D times | after D → DLQ; metric; no infinite loop | Contract | P0 | DLQ | Future |
| ECC-MS-406 | Edge | max_deliver = D+1 attempt after DLQ | not reprocessed from primary; DLQ inspect only | Contract | P1 | DLQ isolation | Future |
| ECC-MS-407 | Edge | JetStream retention = age R: message age R−1s vs R+1s | −1 available; +1 gone | Infra | P1 | retention TTL | Future |
| ECC-MS-408 | Edge | JetStream retention = max msgs M: publish M+1 | oldest dropped or publish backpressure per config | Infra | P1 | retention count | Future |
| ECC-MS-409 | Edge | Request-reply timeout = Tr: reply Tr−1 vs Tr+1 | −1 OK; +1 caller error; no double-apply | API | P0 | req-reply | Future |
| ECC-MS-410 | Edge | Consumer replicas / interest = 0 (misconfig) | startup fail-closed OR warn+block flip | Infra | P0 | GZ-ARCH-16 | Future |
| ECC-MS-411 | Edge | Header-only control msg (no body) with valid subject | reject if schema requires body; metric | Unit | P2 | header-only | TODO |

**Подсчёт §6:** 12

---

## 7. Future Redesign a11y — zoom / reflow / focus-visible / skip-link / text-spacing (`ECC-UX-400+`)

> Углубление WCAG 2.2 AA / NFR-16 относительно wave1–3 (focus trap, 24px target, contrast 4.5:1, reduced-motion, FNO, 320px): **zoom 200/400, reflow 320 without 2D scroll, focus-visible contrast, skip-link, text-spacing 1.4.12** — без 23/24px hit-box и без 4.49:1 contrast.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-400 | Edge | Browser zoom **200%**: primary UI readable, no clipped controls | PASS 1.4.4; targets usable | E2E | P0 | WCAG 1.4.4 | Future |
| ECC-UX-401 | Edge | Browser zoom **400%**: content reflows; essential tasks completable | PASS or documented exception; no JS crash | E2E | P0 | 400% zoom | Future |
| ECC-UX-402 | Edge | Reflow at 320 CSS px width: no horizontal scroll for vertical content | PASS 1.4.10 | E2E | P0 | reflow | Future |
| ECC-UX-403 | Edge | Reflow: data table with horizontal scroll exception only for complex table | exception documented; rest of page no 2D scroll | E2E | P1 | 1.4.10 exceptions | Future |
| ECC-UX-404 | Edge | `:focus-visible` ring contrast ≥ **3:1** against adjacent bg | PASS 1.4.11 / 2.4.7 | E2E | P0 | focus-visible | Future |
| ECC-UX-405 | Edge | Focus ring contrast = **2.9:1** | FAIL a11y gate | E2E | P0 | ring contrast | Future |
| ECC-UX-406 | Edge | Skip-link: first Tab reveals «Skip to content»; activates main | PASS 2.4.1; focus in main | E2E | P0 | bypass blocks | Future |
| ECC-UX-407 | Edge | Skip-link: hidden until focus; not visible on load clutter | visible on focus only | E2E | P1 | skip-link UX | Future |
| ECC-UX-408 | Edge | Text spacing 1.4.12: line-height 1.5× / paragraph 2× / letter 0.12 / word 0.16 applied via bookmarklet | no loss of content/controls | E2E | P1 | 1.4.12 | Future |
| ECC-UX-409 | Edge | Text spacing: overflow clipped after spacing increase | FAIL; layout must adapt | E2E | P1 | spacing clip | Future |
| ECC-UX-410 | Edge | Landmark regions: **0** main landmarks vs **>1** main | exactly one `main`; 0 or many → FAIL | E2E | P1 | landmarks | Future |
| ECC-UX-411 | Edge | Keyboard-only Kanban move: item at column edge (first/last) | move disabled or wrap per policy; no silent no-op without SR text | E2E | P0 | WCAG 2.5.7 | Future |

**Подсчёт §7:** 12

---

## Счётчик

| Домен | Prefix | Edge rows | Gap |
|-------|--------|-----------|-----|
| v4 CRM/HR/Accounting/Intranet | `ECC-V4-400+` | **24** | Future/TODO |
| AI embedding/RAG · @hub-ai · timeout | `ECC-AI-400+` | **14** | Future/TODO |
| Wiki soft-delete 30d · lock · slug | `ECC-WIKI-400+` | **12** | Future/TODO |
| SSO JWKS · claims · break-glass cap | `ECC-SSO-400+` | **12** | Future/TODO |
| MCP autonomy · timeout · revoke | `ECC-MCP-400+` | **12** | Future/TODO |
| MS NATS payload/ack/DLQ · retention | `ECC-MS-400+` | **12** | Future/TODO |
| Redesign zoom/reflow/focus/skip | `ECC-UX-400+` | **12** | Future/TODO |
| **Итого Edge** | | **98** | только Edge; Happy/Neg/Corner не включались |

**Минимум 55:** выполнен (98 ≥ 55).  
**Пересечение `*100+` / `*200+` / `*300+`:** нет (все ID ≥400; фокус-сценарии не копируют wave1–3).  
**Основной каталог:** не изменялся.  
**Коммит:** не создавался.

### Handoff

- Далее: `qa-manual-ver` на полноту/техники/трассируемость.  
- После `verified: true` — выборочная автоматизация границ → `qa-auto1` (API/Unit/Contract/Infra) / `qa-auto2` (E2E a11y zoom/reflow/skip).  
- Не мержить в `EDGE-CORNER-CATALOG.md` без отдельной команды на каталогизацию.
