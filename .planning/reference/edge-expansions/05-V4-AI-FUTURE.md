# Edge expansions — v4 + Future (AI / Wiki / SSO / MCP / MS / UX)

> **Тип:** только **Edge** (границы лимитов, TTL, пустые/max, quota hard-stop).  
> **Не трогает** основной каталог `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` — foresight backlog для будущих `TC-*`.  
> **Gap:** только `Future` | `TODO`.  
> **Источники:** `ROADMAP-FINAL-2026-07-03.md` · `GRAY-ZONES-DECISIONS-2026-07-03.md` §2–§4 · `docs/vision_and_scope.md` (US-INTR / US-CAL / US-WIKI) · PLAN-bases (50k/200) · ADR-AUTH-SSO / ADR-AI-agents (планируемые) · WCAG 2.2 AA (24px, reduced motion).  
> **Владелец:** `qa-manual` · 2026-07-10 · **не коммитить** из этой сессии без явной команды.

Формат строк = каталог ECC: `| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |`

---

## 1. v4 — Intranet / CRM / HR / Accounting / Bases (`ECC-V4-100+`)

> Лимиты: poll 2–10 options (US-INTR-01); bases **50k rows / 200 cols / 30 formula** (GZ-V4-19/20); leave accrual (GZ-V4-12/15); CRM CSV/XLSX import (GZ-V4-08/10); PDF-инвойсы (GZ-V4-17); kudos leaderboard empty/period (US-INTR-04).

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-V4-100 | Edge | Poll: ровно **2** варианта ответа (min) | 201, опрос публикуется | API | P1 | US-INTR-01 | Future |
| ECC-V4-101 | Edge | Poll: ровно **10** вариантов (max) | 201, все 10 видны | API | P1 | US-INTR-01 | Future |
| ECC-V4-102 | Edge | Poll: **11** вариантов | 400, минимум/максимум в error | API | P1 | US-INTR-01 | Future |
| ECC-V4-103 | Edge | Poll: **1** вариант | 400, min=2 | API | P1 | US-INTR-01 | Future |
| ECC-V4-104 | Edge | Poll: вариант с пустой строкой / только whitespace | 400 | API | P2 | US-INTR-01 | Future |
| ECC-V4-105 | Edge | Poll: длина текста варианта на max rune-границе | OK на max; +1 → 400 | API | P2 | US-INTR-01 | Future |
| ECC-V4-106 | Edge | Kudos: отправка себе | 400/422, счётчик не растёт | API | P1 | US-INTR-04 | TODO |
| ECC-V4-107 | Edge | Kudos: сообщение пустое vs max length | empty policy; oversize 400 | API | P2 | kudos v4 | TODO |
| ECC-V4-108 | Edge | Kudos leaderboard: период без голосов | 200, empty state (не 404) | API/E2E | P2 | US-INTR-04 | Future |
| ECC-V4-109 | Edge | Kudos leaderboard: граница месяца (UTC midnight) | kudos в окне/вне окна корректно | API | P1 | NFR-12, US-INTR-04 | Future |
| ECC-V4-110 | Edge | Leave accrual: начисление в день policy boundary (ровно N дней стажа) | баланс += accrual_rate | API | P1 | GZ-V4-12/15 | Future |
| ECC-V4-111 | Edge | Leave accrual: запрос отпуска при балансе **0** | 400/422, баланс не уходит в минус | API | P0 | US-V05 | Future |
| ECC-V4-112 | Edge | Leave accrual: запрос ровно на весь доступный баланс | 201 pending; баланс reserved/0 | API | P1 | accrual policy | Future |
| ECC-V4-113 | Edge | Leave accrual: дробный день (0.5) на границе политики | accept или 400 по policy; без float drift | Unit | P1 | GZ-V4-15 | Future |
| ECC-V4-114 | Edge | Leave: end = start (1 календарный день) | 201, duration=1 | API | P1 | TC-HR leave | TODO |
| ECC-V4-115 | Edge | CRM CSV import: **0** data rows (только header) | 400 или 200 с 0 imported + warning | API | P2 | GZ-V4-08/10 | Future |
| ECC-V4-116 | Edge | CRM CSV import: ровно max rows (лимит импорта) | 200, all imported | API | P1 | GZ-V4-10 | Future |
| ECC-V4-117 | Edge | CRM CSV import: max+1 rows | 400/413; частичный импорт запрещён или documented truncate | API | P1 | GZ-V4-10 | Future |
| ECC-V4-118 | Edge | CRM CSV: строка с unicode/emoji в имени | импорт OK, поиск находит | API | P2 | CRM | Future |
| ECC-V4-119 | Edge | CRM CSV: дубликат email на границе unique | skip/merge policy; no silent overwrite | API | P1 | GZ-V4-10 | Future |
| ECC-V4-120 | Edge | Invoice PDF: размер ровно BodyLimit / policy max | 200, PDF отдаётся | API | P1 | GZ-V4-17, NF-09 | Future |
| ECC-V4-121 | Edge | Invoice PDF: размер max+1 байт | 413/400; файл не сохраняется | API | P0 | PDF invoices | Future |
| ECC-V4-122 | Edge | Invoice PDF: 0 line items | 400 или empty PDF policy | API | P2 | accounting | Future |
| ECC-V4-123 | Edge | Bases: таблица с **50_000** rows (лимит) | create/list OK; insert 50_001 → 400 | API | P0 | GZ-V4-19/20 | Future |
| ECC-V4-124 | Edge | Bases: **200** колонок (лимит) | OK; 201-я → 400 | API | P0 | PLAN-bases Q3 | Future |
| ECC-V4-125 | Edge | Bases: **30** formula-колонок; 31-я | OK / 400 | API | P1 | PLAN-bases | Future |
| ECC-V4-126 | Edge | Bases: пустая таблица list rows offset/limit | 200 [] | API | P2 | ECC-V4B-005 | TODO |
| ECC-V4-127 | Edge | Bases: cell value на max JSON size / string length | 400 over; OK at max | API | P1 | bases | TODO |
| ECC-V4-128 | Edge | Accounting: decimal money на границе scale (0.01 / 0.001) | exact; no float | Unit | P0 | ECC-V4A-008 | TODO |
| ECC-V4-129 | Edge | Accounting: journal entry с нулевой суммой линии | 400 | API | P2 | ECC-V4A-005 | TODO |
| ECC-V4-130 | Edge | Intranet onboarding: чеклист 0/M и M/M прогресс | 0% / 100%; без div-by-zero | API | P2 | US-INTR-02 | Future |

**Подсчёт §1:** 31

---

## 2. AI — BYOK / budgets / model catalog (`ECC-AI-100+`)

> GZ-SEC-22: AES-GCM + org-дневные лимиты токенов с **hard-stop**; ROADMAP W7 LiteLLM-паттерн, virtual keys, catalog + fallback.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-AI-100 | Edge | BYOK key: длина/формат на min/max | reject invalid; accept valid boundary | API | P1 | GZ-SEC-22 | Future |
| ECC-AI-101 | Edge | Org daily token budget: usage == limit | последний запрос OK или soft-last; alert | API | P0 | US-F05, GZ-SEC-22 | Future |
| ECC-AI-102 | Edge | Org daily token budget: usage = limit+1 | **hard-stop** 429/403; no provider call | API | P0 | hard-stop | Future |
| ECC-AI-103 | Edge | User daily budget vs org budget: меньший выигрывает | stop на первом достигнутом | API | P0 | virtual keys | Future |
| ECC-AI-104 | Edge | Budget reset ровно в UTC midnight | до/после границы корректные counters | Unit | P1 | NFR-12 | Future |
| ECC-AI-105 | Edge | Prompt tokens ровно max context модели | accept; max+1 → 400 | API | P1 | ECC-AI-011 | TODO |
| ECC-AI-106 | Edge | Model catalog: пустой список провайдеров | 200 []; UI empty state | API/E2E | P2 | US-F04 | TODO |
| ECC-AI-107 | Edge | Model catalog: fallback когда primary disabled | выбирается next; audit | API | P1 | LiteLLM pattern | Future |
| ECC-AI-108 | Edge | Virtual key: TTL ровно на expiry | до OK; после 401 | API | P0 | virtual keys | Future |
| ECC-AI-109 | Edge | Ollama (local) при cloud budget=0 | local разрешён; cloud blocked | Unit | P1 | ECC-AI-019 | TODO |
| ECC-AI-110 | Edge | Concurrent AI requests = concurrency limit | N OK; N+1 queue/429 | API | P2 | ECC-AI-022 | TODO |
| ECC-AI-111 | Edge | Streaming chunk: mid-stream budget exhaustion | abort + partial usage counted; hard-stop | API | P0 | GZ-SEC-22 | Future |
| ECC-AI-112 | Edge | Provider base URL length / trailing slash normalize | stored canonical; SSRF still blocked | API | P0 | ECC-AI-018 | TODO |

**Подсчёт §2:** 13

---

## 3. Future Wiki (`ECC-WIKI-100+`)

> US-WIKI-01..05 · GZ-SEC-07/11: space RBAC + restrict · depth ≤10 · version history · NF-03 search scope.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-WIKI-100 | Edge | Page depth = **10** (max) | 201; дерево валидно | API | P1 | US-WIKI-01 | Future |
| ECC-WIKI-101 | Edge | Page depth = **11** | 422 | API | P1 | US-WIKI-01 | Future |
| ECC-WIKI-102 | Edge | Space с **0** pages: list tree | 200 [] | API | P2 | US-WIKI-01 | Future |
| ECC-WIKI-103 | Edge | Version history: ровно retention/max versions | oldest kept/purged per policy | API | P1 | US-WIKI-03 | Future |
| ECC-WIKI-104 | Edge | Version history: max+1 save | purge oldest OR 422; история не бесконечна | API | P1 | US-WIKI-03 | Future |
| ECC-WIKI-105 | Edge | Restrict-флаг: пустой allow-list | 422 или эквивалент «никто кроме admin» | API | P1 | US-WIKI-02 | Future |
| ECC-WIKI-106 | Edge | Restrict: allow-list = все space members | эквивалент снятию restrict / OK | API | P2 | US-WIKI-02 | Future |
| ECC-WIKI-107 | Edge | Restrict: попытка добавить user вне space | 422; расширить нельзя | API | P0 | GZ-SEC-07/11 | Future |
| ECC-WIKI-108 | Edge | Space RBAC: последний admin снят | 422; ≥1 admin остаётся | API | P0 | US-WIKI-02 | Future |
| ECC-WIKI-109 | Edge | Search Ctrl+K: пустой query / 1 char | 400 или empty; no full dump | API | P1 | US-WIKI-05, NF-03 | Future |
| ECC-WIKI-110 | Edge | Search: hit только в restrict-странице вне allow-list | 0 results; no title leak | API | P0 | NF-03 | Future |
| ECC-WIKI-111 | Edge | Markdown page body на max size | OK at max; +1 → 400 | API | P1 | US-WIKI-01 | Future |
| ECC-WIKI-112 | Edge | Move page: depth после move = max | OK; move делающий depth>max → 422 | API | P1 | US-WIKI-01 | Future |

**Подсчёт §3:** 13

---

## 4. Future SSO OIDC + LDAP (`ECC-SSO-100+`)

> ADR-AUTH-SSO (W2/W10): OIDC+LDAP · org toggle magic/sso/both · break-glass Magic Link для founder · JIT + LDAP-sync deactivate · без MFA · system-role mapping only.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-SSO-100 | Edge | Org toggle = `magic-link` only: OIDC login attempt | 403/404 SSO path; magic OK | API | P0 | GZ-SEC-04/08 | Future |
| ECC-SSO-101 | Edge | Org toggle = `sso` only: ordinary Magic Link | reject; **founder break-glass** OK | API | P0 | break-glass | Future |
| ECC-SSO-102 | Edge | Org toggle = `both`: оба пути на границе rate-limit | независимые/общие лимиты per ADR | API | P1 | NFR-05 | Future |
| ECC-SSO-103 | Edge | JIT: первый OIDC login неизвестного email | user created + mapped system role; invite не нужен | API | P0 | GZ-SEC-05/09 | Future |
| ECC-SSO-104 | Edge | JIT: email уже pending invite | merge/activate без дубля UUID | API | P1 | JIT | Future |
| ECC-SSO-105 | Edge | LDAP sync: user исчез из AD → deactivate | sessions revoked; API 401 | API | P0 | LDAP-sync | Future |
| ECC-SSO-106 | Edge | LDAP sync: user вернулся в AD | reactivate policy; no auto-admin escalate | API | P1 | LDAP-sync | Future |
| ECC-SSO-107 | Edge | AD group → system role: на границе «только Admin/Member» | project roles не из AD | API | P0 | GZ-SEC-10 | Future |
| ECC-SSO-108 | Edge | Break-glass Magic Link: non-founder при sso-only | 403 | API | P0 | founder only | Future |
| ECC-SSO-109 | Edge | Org toggle flip mid-session (sso→magic) | existing JWT до expiry; refresh policy per ADR | API | P1 | org toggle | Future |
| ECC-SSO-110 | Edge | OIDC token clock skew ±N sec | accept within skew; reject outside | Unit | P1 | OIDC | Future |

**Подсчёт §4:** 11

---

## 5. Future MCP / agents (`ECC-MCP-100+`)

> GZ-SEC-09/20/21/23: ступенчатая автономия · agent-principal · PAT + curated read-heavy toolset · BYOK hard-stop · coverage ≥80% security.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MCP-100 | Edge | PAT: пустой scopes list | 400; токен не создаётся | API | P0 | GZ-SEC-23 | Future |
| ECC-MCP-101 | Edge | PAT: ровно max scopes / unknown scope | OK / 400 | API | P1 | PAT scopes | Future |
| ECC-MCP-102 | Edge | Toolset size = curated max tools | OK; +1 custom tool → reject | API | P1 | curated toolset | Future |
| ECC-MCP-103 | Edge | Toolset: вызов tool вне allow-list | 403; audit | API | P0 | MCP v1 | Future |
| ECC-MCP-104 | Edge | Autonomy: step count = max autonomous steps | last auto OK; next requires confirm | API | P0 | GZ-SEC-09/20 | Future |
| ECC-MCP-105 | Edge | Autonomy: step = max+1 без confirm | blocked; no side-effect | API | P0 | autonomy | Future |
| ECC-MCP-106 | Edge | Agent token budget hard-stop mid-tool-loop | stop; partial audit; no further tools | API | P0 | GZ-SEC-22 | Future |
| ECC-MCP-107 | Edge | PAT TTL на границе expiry | до OK; после 401 | API | P0 | PAT | Future |
| ECC-MCP-108 | Edge | Agent rights = agent ∩ user: user lost project mid-run | subsequent tools 403 | API | P0 | agent-principal | Future |
| ECC-MCP-109 | Edge | Read-heavy tool: empty result set | 200 []; не ошибка | API | P2 | MCP | Future |

**Подсчёт §5:** 10

---

## 6. Future Microservices split (`ECC-MS-100+`)

> Track P: schema-per-service · NATS contract version · Caddy route-flip dual-run (notification/ai/ws) · GZ-ARCH-16/17 · ADR-ARCH-MICROSERVICES.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-MS-100 | Edge | Schema-per-service: migrate expand на пустой schema | init OK; app starts | API/Infra | P1 | GZ-ARCH-08 | Future |
| ECC-MS-101 | Edge | Schema-per-service: contract phase — old monolith read new cols | backward-compatible; no 500 | API | P0 | Atlas expand-contract | Future |
| ECC-MS-102 | Edge | NATS event: producer vN, consumer expects vN | OK | Contract | P0 | GZ-QA-08/09 | Future |
| ECC-MS-103 | Edge | NATS event: producer vN+1, consumer vN only | reject/DLQ; no silent drop of required fields | Contract | P0 | contracts/events | Future |
| ECC-MS-104 | Edge | NATS event: missing required field at schema boundary | schema-test fail both sides | Unit | P0 | M0-02 | Future |
| ECC-MS-105 | Edge | Route-flip dual-run: 0% traffic on new service | monolith serves; new idle healthy | Infra | P1 | Caddy flip | Future |
| ECC-MS-106 | Edge | Route-flip dual-run: 100% flip `/api/notifications*` | new service only; p95 ≤ +10% gate | Infra/Perf | P0 | GZ-ARCH-05/17 | Future |
| ECC-MS-107 | Edge | Route-flip mid-request in-flight | no duplicate side-effect / idempotent consumer | API | P0 | strangler | Future |
| ECC-MS-108 | Edge | hub-ai split: BYOK secret не читается из hub-core pod | NetworkPolicy/Secret isolation | Infra | P0 | M3 BYOK | Future |
| ECC-MS-109 | Edge | hub-ws split: membership read-model empty Redis | subscribe deny (fail-closed) | API | P0 | NF-02 | Future |

**Подсчёт §6:** 10

---

## 7. Future Redesign a11y edges (`ECC-UX-100+`)

> Только **UI Edge-границы**: WCAG 2.2 AA — Target Size ≥24×24px · focus trap · `prefers-reduced-motion` · Focus Not Obscured. Не Happy-флоу экранов.

| ID | Тип | Сценарий | Ожидание | Слой | P | Связь | Gap |
|----|-----|----------|----------|------|---|-------|-----|
| ECC-UX-100 | Edge | Modal/dialog: Tab cycle focus trap (first↔last) | focus не уходит на background | E2E | P0 | WCAG 2.4.3/2.1.2 | Future |
| ECC-UX-101 | Edge | Modal: Escape закрывает; focus возвращается на trigger | restore focus | E2E | P1 | a11y redesign | Future |
| ECC-UX-102 | Edge | Icon-only control: hit target **24×24 CSS px** ровно | PASS axe/target-size | E2E | P0 | WCAG 2.5.8 | Future |
| ECC-UX-103 | Edge | Icon-only control: target **23px** | FAIL gate / redesign required | E2E | P0 | 24px | Future |
| ECC-UX-104 | Edge | `prefers-reduced-motion: reduce` | decorative motion off; essential state changes remain | E2E | P1 | redesign | Future |
| ECC-UX-105 | Edge | Sticky header + focused control: Focus Not Obscured | focused element не перекрыт sticky | E2E | P1 | WCAG 2.4.11 | Future |
| ECC-UX-106 | Edge | Kanban DnD: keyboard alternative на границе «без drag» | move via buttons/menu; 2.5.7 | E2E | P0 | WCAG 2.5.7 | Future |
| ECC-UX-107 | Edge | Viewport 375px: primary CTA ≥24px, no clipped focus ring | usable; ring visible | E2E | P1 | NFR-16 | Future |

**Подсчёт §7:** 8

---

## Счётчик

| Домен | Prefix | Edge rows | Gap |
|-------|--------|-----------|-----|
| v4 Intranet/CRM/HR/Accounting/Bases | `ECC-V4-100+` | **31** | Future/TODO |
| AI BYOK/budgets/catalog | `ECC-AI-100+` | **13** | Future/TODO |
| Wiki (future) | `ECC-WIKI-100+` | **13** | Future/TODO |
| SSO OIDC+LDAP (future) | `ECC-SSO-100+` | **11** | Future/TODO |
| MCP/agents (future) | `ECC-MCP-100+` | **10** | Future/TODO |
| Microservices split (future) | `ECC-MS-100+` | **10** | Future/TODO |
| Redesign a11y edges (future) | `ECC-UX-100+` | **8** | Future/TODO |
| **Итого Edge** | | **96** | только Edge; Happy/Neg/Corner не включались |

**Минимум 50:** выполнен (96 ≥ 50).  
**Основной каталог:** не изменялся.  
**Коммит:** не создавался.
