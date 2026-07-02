# Hub — Анализ расширения и доведения функционала

> **Дата:** 2026-07-02
> **Автор:** глубокий аудит кода + планов (read-only, без изменений кода/тестов)
> **Цель:** подготовить разговор о РАСШИРЕНИИ функционала и ДОДЕЛЫВАНИИ планов.
> **Гипотеза пользователя:** «реализовано всё минимально, а нужно всё» — многие фичи есть в roadmap/vision, но реализованы поверхностно/частично.
> **Источники:** `.planning/{ROADMAP,STATE,PROJECT,COVERAGE,PLATFORM-VISION-v4,MILESTONE-3}.md`, `.planning/milestones/v4-super-app/SUMMARY.md`, `.planning/AUDIT-2026-06-23.md`, `.planning/TECH-DEBT-AUDIT-2026-06-23.md`, `docs/{vision_and_scope,competitive_audit}.md`, `.cursor/rules/60-tz-constraints.mdc`, фактический код `backend/internal/**` и `frontend/src/**`.

---

## 1. Резюме

**Гипотеза подтверждается частично — и её нужно разбить на две разные истории:**

1. **Ядро (dev-suite: auth, tracker, sprint, messenger, tms, calls, dashboard, search, notification, rbac, ws, github/gitlab, ai)** — реализовано **глубоко и полноценно**. Это НЕ минимум. Крупные модули: `tms` 6834 строк Go, `messenger` 5681, `tracker` 5462, `ai` 3634, `calls` 2848, `notification` 2681, `auth` 2452. Покрытие тестами 80–92%, 286+ REST endpoints, реальные фичи (JQL-парсер `search/jql.go` 348 строк, webcam в звонках `components/calls/CallControls.tsx`, Planning Poker, burndown/velocity, Gantt по зависимостям). Гипотеза «минимально» к ядру **не применима**.

2. **Модули v4 super-app (intranet, calendar, crm, hr, accounting, bases, marketplace, status) — это MVP-скелеты**, и здесь гипотеза **подтверждается полностью**. Признаки:
   - Фронтенд каждого v4-модуля — это **один-два файла** (calendar 430 строк одним файлом, accounting 362, marketplace 331, bases 341, crm 358, hr 244, status 79). Для сравнения: tms фронтенд — 19 файлов/4672 строки, tracker — 17 файлов/3352.
   - Бэкенд v4-модулей — классический `handler/service/repository` CRUD на 600–1500 строк без глубокой бизнес-логики (accounting 1099, marketplace 915, bases 855, crm 757, hr 642, calendar 517).
   - В PLAN-файлах явно перечислен большой «Out of scope»: у accounting (`.planning/phases/52-accounting-v45/PLAN.md`) — нет PDF, НДФЛ, мультивалюты, AR aging, баланса, cash flow; у bases — нет формул; и т.д.

3. **Отдельная категория — «polish/depth-долг» даже в ядре**: сам `docs/competitive_audit.md` §3 честно помечает как ⚠️ partial: optimistic UI (только tracker), skeleton loaders, empty states, RHF+Zod формы (только onboarding), flaky detection (partial), issue↔message bidirectional (partial), notification digest (backlog).

4. **Важное расхождение планов и статуса (doc drift):** в `.planning/ROADMAP.md` таблица Progress показывает **Phase 8 «In Progress» (1/2), Phase 9 «Not started», а Phase 12 — 4 из 5 планов не отмечены** (`[ ] 12-01/02/03/05`), тогда как `STATE.md` и `AGENTS.md` заявляют v2.0 Redesign «COMPLETE». То есть даже «завершённый» редизайн формально имеет неотмеченные планы. Это надо сверить с фактом до планирования (см. §6).

**Вывод одной фразой:** ядро — «полноценное», v4-надстройка — «широкая, но неглубокая» (по фиче на модуль), плюс сквозной UX/polish-долг. Разговор о «доведении до полной реализации» — это в первую очередь **углубление v4-модулей** и **сквозной UX/observability polish**, а не переписывание ядра.

---

## 2. Таблица по модулям/доменам

Легенда глубины: **полное** = соответствует vision, есть нетривиальная бизнес-логика и тесты · **частичное** = базовый happy-path есть, заметные пробелы vs vision · **скелет** = CRUD/1 экран, «MVP-заглушка масштаба» · **нет** = не реализовано.
Приоритет доведения: **P0** = критично для «полноты»/доверия · **P1** = высокая ценность · **P2** = средняя · **P3** = nice-to-have.

| Модуль | Заявлено (roadmap/vision) | Реализовано (факт, глубина) | Основные пробелы | Приоритет |
|--------|---------------------------|------------------------------|------------------|-----------|
| **auth** | Magic Link + Telegram bind, JWT 15м + refresh 30д, сессии, rate limit, онбординг | **полное** (2452 стр, cov 82.4%). Split-сервисы session/login/profile. | 2FA — только placeholder (`routes/.../settings/security.tsx:35`); «уведомление о новом входе» — проверить. | P2 |
| **tracker** | 9 типов, кастомные поля, связи, workflow, комментарии, таймтрекинг, метки | **полное** (5462 стр, cov 82.4%). | Кастомные поля: схема есть, **UI-билдер отсутствует** (competitive_audit §4 Could Have); метки на карточке — заглушка (`issue-card.tsx:223`). | P1 |
| **kanban/board** | колонки=статусы, WIP, swimlanes, несколько досок, фильтры | **частичное/полное**. WIP/swimlanes есть. | «Несколько досок в проекте», сохранённые/расшариваемые фильтры — проверить глубину. | P2 |
| **sprint/scrum** | спринты, burndown, velocity, retro, автоперенос, Planning Poker | **полное** (1179 стр backend, `burndown-chart.tsx`, `velocity-chart.tsx`, `planning-poker.tsx` 191 стр). | Retrospective и автоперенос — проверить полноту. | P2 |
| **messenger** | DM/группы/каналы, треды, реакции, голосовые, превью ссылок, поиск, статусы | **полное** (backend 5681, FE-компоненты 6308 стр). | `message-composer.tsx` >500 стр (god-component, tech-debt); optimistic reactions — partial. | P1 |
| **tms** | кейсы classic/BDD, папки, версии, прогоны, Allure, окружения, отчёты Allure-уровня, импорт/экспорт | **полное** (backend 6834, FE 4672; cov 82%). Flaky, autotest, traceability, environments — есть отдельные экраны. | Отчёты «уровня Allure TestOps» (тренды, run diff, категории дефектов) — **частично**; flaky detection — partial (competitive_audit §2). | P1 |
| **calls** | голос+screen share (без вебкамер по vision), запись, запланированные, чат звонка | **полное** (2848 стр; cov 80.2%). **Webcam ФАКТИЧЕСКИ реализован** (`CallControls.tsx`, `ParticipantList.tsx`) — превышает исходный vision (Phase 16 IN). | Breakout rooms, noise cancel (competitive_audit — Zoom-gap) — не в scope. | P3 |
| **dashboard** | личный/проектный/админский, drag&drop виджеты, шаблоны Dev/QA/Manager | **частичное** (1071 backend / 1455 FE). Виджеты есть. | Виджеты «статичные» (competitive_audit §2); шаблоны дашбордов Dev/QA/Manager — проверить; кастомные панели — нет. | P2 |
| **search** | Ctrl+K глобальный, JQL, локальный поиск, Meilisearch | **полное**. `jql.go` 348 стр — реальный парсер; Meilisearch (Phase 94); cov 92.2%. | Сохранённые/расшариваемые фильтры-запросы — частично. | P2 |
| **notification** | триггеры, in-app+Telegram+WebPush, настройки по типам/каналам/проектам, quiet hours, digest | **частичное/полное** (2681 стр; cov 82.5%). Quiet hours ✅. | **Digest — backlog** (competitive_audit §3); smart notifications нет. | P2 |
| **rbac** | системные + проектные роли (Owner/Dev/Tester/Viewer) + кастомные с полным набором прав | **полное** (881 стр; cov 84.7%). Repository split (Phase 26). | Permission representation расширен; UI редактора кастомных ролей — проверить глубину. | P2 |
| **github/gitlab** | PR/коммиты, автопривязка HUB-123, статус PR, автосмена статуса, создание ветки | **полное** (github 1601 / gitlab 1276; cov 80.8%/68.5%). SafeHTTP. | gitlab cov 68.5% (ниже gate); «создание ветки из задачи» — проверить. | P2 |
| **ai** | 5+ провайдеров (GigaChat/YandexGPT/DeepSeek/Mistral/Ollama), @hub-ai, стриминг, /ai команды, RAG, BYOK, model picker, лимиты | **полное** (3634 стр; cov 78.9%). Все M3-001..030 отмечены done. | cov 78.9% (<80); RAG «по докам проекта» — глубину проверить; Tab-suggest в composer. | P2 |
| **intranet** | news feed, org directory, kudos, polls, marketplace, onboarding paths, Random Coffee | **скелет** (backend 1516; FE 5 файлов/754 стр). News/kudos/directory/coffee — по одному экрану. | Polls company-wide, onboarding paths (не-dev), лента: pin/reactions глубина, модерация. | **P1** |
| **calendar** | unified (спринты/созвоны/отпуска/дедлайны), iCal export, room booking, team availability | **скелет** (backend 517; FE **1 файл** 430 стр). Unified view + iCal есть. | Room booking, team availability, двусторонний Google Calendar sync, recurring. | **P1** |
| **crm** | контакты/компании, deals pipeline (kanban), link deal→project, won→invoice | **скелет** (backend 757; FE 2 файла/358 стр). Won→draft invoice hook есть. | Компании как сущность, активности/задачи по сделке, воронка-аналитика, импорт. | P2 |
| **hr** | leave requests + approver workflow, org chart, handbook, 1C sync | **скелет** (backend 642; FE 1 файл/244 стр). Leave + 1C webhook (Phase 48). | Org chart визуальный, employee handbook (wiki), баланс отпусков, календарная интеграция глубина. | P2 |
| **accounting** | ledger + trial balance + invoices + P&L (MVP; расширенное — 1C connector/v4.5) | **скелет/MVP** (backend 1099; FE 1 файл/362 стр). Double-entry, invoice draft/issue/pay, income statement. | Явный out-of-scope (PLAN 52): **нет PDF, НДФЛ, мультивалюты, AR aging, баланса, cash flow**. | P2 |
| **bases** (Lark Base) | no-code таблицы, grid-редактор | **скелет** (backend 855; FE 1 файл/341 стр). Base→table→column→record CRUD, inline-edit. | **Нет формул** (известное ограничение), нет relations/lookup, views/filters/sort, kanban/calendar-view, sharing. | P2 |
| **marketplace** | доска объявлений (moderated) | **скелет** (backend 915; FE 1 файл/331 стр). Listings + mod queue. | Категории/поиск/фото-галерея, сделки, глубина модерации. | P3 |
| **status** | публичная страница статуса | **скелет** (backend 101; FE 79 стр). Health probes. | История инцидентов, uptime %, подписка. | P3 |
| **infra/CI** | Docker Compose dev, K3s prod, Caddy, GitHub Actions (~3000 мин) | **полное**. CI green (coverage-gate ≥82%, e2e-full 20 specs). | GZIP/Brotli compression отсутствует (`wiring_fiber.go`, tech-debt); Lighthouse gate non-blocking. | P2 |
| **observability** | slog + Loki + Prometheus + Grafana + OTel + Alertmanager | **полное** (Phases 29–36). | Дашборды «статичные»; алерты — базовый набор. | P3 |
| **security** | 0 CRITICAL, secret rotation, WCAG, headers | **полное по коду** (12/12 blockers fixed, 6/6 security ≥80%). | **FA-06/PG-03 ротация секретов — manual, единственный prod-блокер**; NFR-16 WCAG full sign-off; FA-02-03/04 backlog. | **P0** |

---

## 3. Планы, не доведённые до конца

### 3.1 Формально «IN SCOPE by phase» (из `.cursor/rules/60-tz-constraints.mdc` + PO-lock) — статус по факту

| Фича | Заявленная фаза | Факт |
|------|-----------------|------|
| **Gantt** | 16 | ✅ есть, но **минимальный** — `gantt-chart.tsx` 135 стр, только сроки + `blocked_by`. Нет критического пути, drag-resize, зависимостей FS/SS/FF. → **углубить (P2)**. |
| **Webcam in calls** | 16 | ✅ реализован (`CallControls.tsx`) — фактически done. |
| **hub-mail: транзакционные письма** | 15.2 | ✅ модуль `mail` (516 стр) + FA-05 prod compose. Глубину шаблонов проверить. |
| **hub-mail: маркетинговые письма/кампании** | 16 | ⚠️ есть `/admin/mail`; полноценные кампании/сегменты/аналитика — **проверить/доработать (P2)**. |
| **Full JQL automation** | 17 | ✅ JQL-парсер `search/jql.go`; automation Phase 16 (PROD-01). Триаж-инбокс — **polish**; полнота JQL-функций и сохранённых автоматизаций — проверить. |
| **AI gateway BYOK + Ollama** | 17 | ✅ реализовано (M3-001..030, `ai/byok-settings.tsx`, `admin/ai-settings.tsx`). cov 78.9%. |
| **No-code DB (Lark Base)** | v4 | ⚠️ **скелет** — без формул/relations/views. → **углубить (P2)**. |
| **Accounting** | v4 | ⚠️ **MVP** — без PDF/НДФЛ/мультивалюты/баланса. → **углубить или ограничить 1C-коннектором (P2)**. |
| **Intranet / Calendar / CRM / HR / Marketplace** | v4 | ⚠️ **скелеты** (см. §2). → **основной фронт работ по «полноте» (P1–P2)**. |

### 3.2 Явно отложенное в backlog (competitive_audit §4 / vision)

- Custom fields builder **UI** (схема в БД есть, редактора нет) — Could Have.
- Timeline roadmap / initiatives view — Could Have.
- Triage inbox UX polish — Should Have.
- Notification **digest** — backlog.
- Flaky test detection (ML/статистика) — partial.
- Feature flags (Redis-backed) — backlog.
- Bases formula engine — v4.5+ backlog.
- Accounting v4.5 (расширенный P&L, AR aging) — backlog.
- Mobile PWA offline depth — Could Have.

### 3.3 Незакрытые/дрейфующие планы в ROADMAP

- **Phase 8 Design System** — Progress-таблица: `1/2`, «In Progress» (план `08-02` !important-cleanup помечен `[x]`, но статус строки не обновлён).
- **Phase 9 Layout Shell** — «Not started» (0/2) в таблице, при этом планы `09-01/09-02` помечены `[x]` выше. Drift.
- **Phase 12** — планы `12-01` (Issue Detail+Calls), `12-02` (TMS restyle), `12-03` (Settings+Admin), `12-05` (Unmocked pages) помечены `[ ]`, при этом строка Progress = «Complete 5/5». **Расхождение: либо restyle не завершён, либо чекбоксы не проставлены.** → сверить визуально (P1 для «полноты редизайна»).
- **Phase 999.1** — единый тест-фреймворк: помечен BACKLOG в ROADMAP, но STATE заявляет TF-миграцию COMPLETE. Формально закрыт по STATE; ROADMAP-backlog не вычищен (drift).

### 3.4 v5 / за рамками (нужно решение ADR прежде чем «доделывать»)

Из tz_constraints FORBIDDEN до ADR: multi-org, SSO/LDAP/OAuth, SaaS billing, нативные приложения. Их «доведение» — не про полноту текущего vision, а про новый milestone.

---

## 4. Технический долг и блокеры (из AUDIT / TECH-DEBT / STATE)

### 4.1 Prod-блокер (P0)
- **FA-06 / PG-03 — ротация секретов** (`RESEND_API_KEY`, `TELEGRAM_BOT_TOKEN`, `livekit-prod.yaml` devkey утекали в git history до `859ecef`). Manual ops, скрипты готовы (`scripts/fa06.sh`). **Единственный блокер тега v2.5.0/деплоя.** Также git history purge (BFG).

### 4.2 Из `TECH-DEBT-AUDIT-2026-06-23.md`
- **CRITICAL perf:** N+1 в `audit/service.go:122` (`GetResourceLimits`, 3 запроса → батч).
- **CRITICAL hygiene:** hardcoded GIPHY API key `frontend/src/components/messenger/gif-picker.tsx:47` → `import.meta.env`. (Файл сейчас в `git status` как изменённый — вероятно уже правится.)
- **Bug:** `auth/handler.go` VerifyMagicLink проверяет длину кода по байтам (`len(req.Code)!=6`) → `utf8.RuneCountInString`.
- **God-components:** `message-composer.tsx` >500 стр; `onboarding.lazy.tsx` и `cases_.$caseKey.lazy.tsx` ~30KB — декомпозиция.
- **Perf/UX:** `MessageItem` без `React.memo`; `messenger-store.updateChannel` O(n) `.map()`; нет GZIP/Brotli в `wiring_fiber.go`; отсутствуют composite-индексы в `issue.go/user.go/project.go`.
- **INFO:** много «naked errors» (`return err`) без `%w`-обёртки в handlers.

### 4.3 Качество/compliance (P1, не блокер)
- **NFR-16 WCAG 2.1 AA** — только login axe PASS, полного human-аудита (dark+light) нет.
- **NFR-17 Lighthouse ≥90** — job non-blocking (artifact only), нет blocking-gate.
- **Покрытие ниже цели:** gitlab 68.5%, ai 78.9%, bases 78.6% (при overall gate 82%).
- **FA-02 backlog:** LiveKit JWT `iss` constant-time compare (FA-02-03); RequireAuth fail-open edge (FA-02-04).
- **Biome:** ~477 pre-existing warnings (tech debt).

### 4.4 Что мешает именно «полной» реализации v4
- v4-модули писались под floor **≥50% coverage** и «MVP-scope» (см. SUMMARY «All v4 backend modules meet ≥50%»); нет глубины domain-логики и почти нет FE-декомпозиции (1 файл на модуль). Доведение = новая волна ba→sa→arch→be/fe, а не патч.

---

## 5. Приоритизированный бэклог расширения (по эпикам)

> Оценка размера: **S** ≈ 1–2 дня · **M** ≈ 3–5 дней · **L** ≈ 1–2 недели · **XL** ≈ >2 недель. Готово к декомпозиции `tech-lead`.

### ЭПИК A — Prod cutover & trust (P0, разблокировка релиза)
- A1. FA-06/PG-03 ротация секретов + BFG purge + CTO sign-off — **M** (ops). Зависимости: нет.
- A2. SSH deploy (GitHub Secrets, firewall :22) + `git tag v2.5.0` — **S**. Зависит от A1.
- A3. Быстрые фиксы tech-debt: GIPHY env, magic-link `utf8.RuneCount`, N+1 в audit — **S**. Зависимости: нет.

### ЭПИК B — Доведение v4 «широких, но неглубоких» модулей (P1) — **главный ответ на гипотезу**
- B1. **Intranet до «Mighty-lite»**: polls company-wide, onboarding paths (не-dev), лента (pin/reactions/комментарии), модерация — **L**. Зависит от messenger polls.
- B2. **Calendar first-class**: room booking, team availability (presence+blocks), recurring events, двусторонний iCal/Google — **L**. Зависит от calls/sprint данных.
- B3. **CRM углубление**: companies-сущность, активности/задачи по сделке, воронка-аналитика, импорт контактов — **L**. Зависит от tracker (link deal→project) + accounting (invoice).
- B4. **HR углубление**: визуальный org chart, employee handbook (wiki-страницы), баланс отпусков, календарная интеграция — **M/L**. Зависит от B2 + intranet wiki.
- B5. **Bases до Lark**: formula engine, relations/lookup, views (grid/kanban/calendar), фильтры/сортировки, sharing — **XL**. Самостоятельный.
- B6. **Marketplace/Status polish**: категории/поиск/галерея; история инцидентов/uptime — **M**. Низкий приоритет внутри эпика.
- B7. **FE-декомпозиция v4**: каждый модуль из 1 файла → components/hooks/stores по правилу `.cursor/rules/30-frontend.mdc` — **M** (сквозное, параллелит fe1/fe2).

### ЭПИК C — Depth ядра / фичи vision, отмеченные partial (P1–P2)
- C1. **Custom fields builder UI** (backend-схема есть) — **M**.
- C2. **Gantt pro**: критический путь, drag-resize, типы зависимостей — **M**. Зависит от tracker links.
- C3. **TMS reporting уровня TestOps**: тренды, run diff, категории дефектов, flaky-статистика — **L**. Зависит от tms.
- C4. **Notification digest + smart delivery** — **M**. Зависит от notification.
- C5. **Dashboard кастомные панели + шаблоны Dev/QA/Manager** — **M/L**.
- C6. **Issue↔message полноценная двусторонняя связь**, triage inbox polish — **M**.
- C7. **Email marketing кампании** (сегменты, аналитика) — **M**. Зависит от hub-mail.

### ЭПИК D — UX polish (сквозной, P2)
- D1. Skeleton loaders вместо «Загрузка...», empty states с CTA — **M** (сквозной FE).
- D2. Optimistic UI для messenger reactions и остальных мутаций — **S/M**.
- D3. RHF+Zod на все формы (issue-form и др.) — **M**.
- D4. God-components рефактор (`message-composer`, `onboarding.lazy`, `cases_.$caseKey.lazy`) — **M**.

### ЭПИК E — Quality/compliance gates (P1–P2)
- E1. **WCAG 2.1 AA full sign-off** (dark+light, все страницы) — **M**. Разблокирует NFR-16.
- E2. **Lighthouse blocking gate ≥90** после baseline — **S**. Зависит от baseline scores.
- E3. **Coverage uplift**: gitlab 68.5%→80, ai→80, bases→80 — **M**.
- E4. **Perf**: React.memo (MessageItem), нормализация messenger-store, GZIP/Brotli, composite-индексы — **M**.
- E5. FA-02-03/04 security backlog — **S**.

### ЭПИК F — Платформа/масштаб (P3, требует ADR)
- F1. Feature flags (Redis) — **M**.
- F2. Микросервисный split (hub-ai-gateway, hub-intranet) при >500 concurrent WS — **XL** (v4.5, только по нагрузке).
- F3. v5-кандидаты (multi-org, SSO/LDAP) — **XL**, только после продуктового ADR.

**Рекомендованный порядок волн:** A (разблокировка) → E1/E4 + C (быстрая ценность ядра) параллельно B7 → B (глубина v4) → D (сквозной polish) → F (по решению).

---

## 6. Открытые вопросы к продукту (уточнить до старта)

1. **Что значит «нужно всё»?** Довести v4-модули до «полноценных» (Bitrix/Lark-паритет) — это XL-объём. Приоритет: (а) вглубь v4 или (б) вширь ядра (custom fields UI, TMS reports, dashboards)?
2. **Bitrix/Lark-паритет или «достаточно для 10–50 человек»?** Для bases: реально нужен formula engine и relations, или хватит таблиц+представлений?
3. **Accounting: строить свою бухгалтерию** (баланс, cash flow, НДФЛ, мультивалюта) **или ограничиться 1C-коннектором** (как декларирует PLATFORM-VISION-v4 §3.6 и §7)? Это меняет размер с XL на M.
4. **Релиз сейчас или после расширения?** FA-06 (P0) блокирует v2.5.0 tag. Тегаем production-ready ядро сначала, потом расширяем — или всё одной волной?
5. **Redesign Phase 12 (Settings/Admin/TMS/Unmocked)**: чекбоксы в ROADMAP не проставлены — это реальный незавершённый restyle или только doc-drift? Нужен визуальный аудит.
6. **v5-фичи (multi-org, SSO/LDAP)** — входят в горизонт этого разговора о «доведении» или это отдельный milestone за ADR-границей tz_constraints?
7. **Целевое покрытие для новых v4-фич** — держим ≥82% как для ядра, или для «широких» модулей допускаем ≥50% на первой итерации?
8. **Команда/параллелизм**: сколько параллельных потоков (fe1/fe2, be1/be2) закладывать — от этого зависит нарезка волн у tech-lead.

---

## Приложение — метрики глубины (факт, 2026-07-02)

**Backend `internal/` (строки Go без тестов, топ и v4):**
tms 6834 · messenger 5681 · tracker 5462 · ai 3634 · calls 2848 · notification 2681 · auth 2452 · project 2297 · audit 1939 · search 1607 · github 1601 · **intranet 1516** · apigateway 1286 · gitlab 1276 · imex 1189 · sprint 1179 · **accounting 1099** · dashboard 1071 · ws 1004 · **marketplace 915** · rbac 881 · **bases 855** · **crm 757** · backup 643 · **hr 642** · storage 579 · **calendar 517** · mail 516 · **status 101**.

**Frontend `src/modules/` (файлы/строки):**
tms 19/4672 · tracker 17/3352 · calls 10/2234 · settings 5/815 · **intranet 5/754** · **calendar 1/430** · messenger 2/406 · **accounting 1/362** · **crm 2/358** · **bases 1/341** · **marketplace 1/331** · **hr 1/244** · admin 1/204 · **status 1/79**.
(Основной UI messenger/admin/dashboard живёт в `src/components/`: messenger 28/6308, admin 9/2363, dashboard 13/1455.)

**Контекст качества:** overall coverage 83.0%, security 6/6 ≥80%, 286+ REST, 83 Ent-схемы, 495 backend test-файлов, Vitest 99/99, 20 Playwright specs. Литеральных «not implemented/заглушка» стубов в коде почти нет — «минимальность» проявляется в **узости фич v4 и глубине polish**, а не в TODO-заглушках.
