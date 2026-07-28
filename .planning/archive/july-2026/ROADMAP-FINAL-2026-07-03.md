# Hub — Финальный Roadmap (2026-07-03)

> **Статус:** Волна 1 исполнена в фича-ветках (2026-07-03); мерж master — пользователь; код W2+ / July Rebuild — freeze до align ~конец июля 2026.
> **Входы:** `../gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`, `ADR-ARCH-MICROSERVICES-2026`, `../roadmaps/MICROSERVICES-ROADMAP.md`, `EXPANSION-ANALYSIS`, `MARKET-AND-NEEDS`, A/B-планы.
> **Принцип:** rolling-wave — ближние волны атомарно (≤200–300 строк/план), дальние укрупнённо; `tech-lead` дробит перед стартом.
> **Процесс:** агенты коммитят/пушат фича-ветки; **push/мерж в master — только пользователь**; PR=фича, rebase-merge; QA-конвейер: TC → qa-manual-ver → автотесты → qa-auto-ver; юниты пишут be/fe.

## Статус на 2026-07-10

> Подробный снимок: **`.planning/archive/july-2026/PROJECT-STATUS-2026-07-10.md`**.
> Операционный план возврата: **`.planning/archive/july-2026/ROADMAP-DETAILED-JULY-RETURN.md`**.
> Brief: `.planning/archive/july-2026/JULY-2026-REBUILD-BRIEF.md` · Индекс: `.planning/archive/july-2026/INDEX-JULY-2026.md`.

| Факт | Статус |
|------|--------|
| Волна 1 | ✅ Исполнена в фича-ветках · `integration/wave1` merge-ready · **мерж master = пользователь** |
| Волны 2–12 / Track P | ⏸ Код freeze до July align (~конец июля) |
| SFU / M4 hub-calls | **LiveKit** · SEED-001 mediasoup **REJECTED** |
| Active эпики (planning) | EPIC-CAL-SYNC · EPIC-REDESIGN · EPIC-QA-HARDENING |
| FA-06 | ⏳ Manual ops (не гейтит feature-dev) |
| Код July Rebuild | Не стартовал |

---

## Общая картина: два трека + новые модули

```
Track F (функционал):  W1 quick-wins → W2-3 аналитика+wiki-спека → W4-6 intranet+calendar
                       → W7-8 crm+hr → W9-10 accounting → bases-эпик (BS-0..5, параллельно на be-потоке)
                       → wiki-реализация → AI-hub (LiteLLM-паттерн) + MCP → SSO
Track P (платформа):   W2 M0 (ADR-детали+K3s+OTel+path-CI) → W3-4 M1 hardening
                       → W5-6 M2 пилот hub-notification → GATE Go/No-Go
                       → M3 hub-ai → M4 realtime-трио (hub-ws/hub-messenger/hub-calls **на LiveKit**)
Ops-лейн (фоновый):    FA-06 ротация (пользователь по runbook) → tag v2.5.0 → SSH deploy
```

Горизонт: ~12–14 волн ≈ 2.5–3 месяца. Волны ≤4 агента (мягкий лимит, до 6–8 по решению team-lead при явной карте файлов).

---

## ВОЛНА 1 — Quick-wins + разблокировка (атомарная декомпозиция)

> Цель: снять срочные риски (security-gap bases, tech-debt), запустить ops-лейн, дать аналитике фору. Ветка: `fix/wave1-quickwins` + отдельные.

| # | Задача (атомарный план) | Исполнитель | Верификатор | Файлы/границы | DoD |
|---|--------------------------|-------------|-------------|----------------|-----|
| 1.1 | **Bases security-fix:** `ListBases`/`GetBase` фильтр по membership | be1 | be-ver + security-auditor | `backend/internal/bases/service.go`, `ent_repository.go` | Чужие базы невидимы; unit-тест на отказ; регрессия green |
| 1.2 | Bases security-fix: тесты на RBAC-отказы (обновление/удаление чужой базы) | be1 | qa-auto-ver | `backend/testframework/bases/` | 403 на все мутации не-участника |
| 1.3 | GIPHY-ключ → `import.meta.env` + `.env.example` | fe1 | fe-ver | `frontend/src/components/messenger/gif-picker.tsx`, `.env.example` | Ключа нет в коде; env документирован |
| 1.4 | Magic-link: `len(req.Code)` → `utf8.RuneCountInString` | be2 | be-ver | `backend/internal/auth/handler.go` | Юнит с multibyte-кодом |
| 1.5 | N+1 в audit: `GetResourceLimits` 3 запроса → батч | be2 | be-ver | `backend/internal/audit/service.go`, `repository.go` | 1 запрос; тест на результат |
| 1.6 | FA-02-03: LiveKit JWT `iss` → constant-time compare | security-auditor (спека) + be2 (фикс) | be-ver | `backend/internal/calls/` | subtle.ConstantTimeCompare; тест |
| 1.7 | FA-02-04: RequireAuth fail-open edge → fail-closed | be2 | security-auditor | `backend/internal/platform/middleware/` | Тест на отказ без токена |
| 1.8 | NF-09: presigned PUT — S3-policy лимит размера | devops-engineer | security-auditor | SeaweedFS config / storage | PUT >25MB отклоняется |
| 1.9 | **Ops-лейн:** runbook ротации FA-06 для пользователя (чек-лист) | devops-engineer | security-auditor | `scripts/fa06.sh`, runbook.md | Пользователь выполняет по чек-листу; env/Secrets обновлены агентом |
| 1.10 | US/AC для Intranet (опрос-в-посте, onboarding-шаблоны, постмодерация, kudos-лидерборд, coffee-матчинг) | ba2 | ba-ver | `docs/vision_and_scope.md` (раздел intranet) | AC Gherkin, MoSCoW |
| 1.11 | US/AC для Calendar (ресурсы, recurring d/w/m, iCal-подписки, availability busy/free) | ba2 | ba-ver | `docs/vision_and_scope.md` (раздел calendar) | AC Gherkin |
| 1.12 | US/AC для Wiki (space-модель, редактор, версии, связь с задачами) — новый модуль | ba1 | ba-ver | `docs/vision_and_scope.md` (новый раздел wiki) | AC Gherkin; RBAC по GZ-решению |
| 1.13 | Docs-drift: сверить ROADMAP чекбоксы Phase 8/9/12 с фактом (визуальный аудит) | docs-writer | product-ver | `.planning/ROADMAP.md` | Дрейф устранён или заведены задачи |

Ворота волны: регрессия доменов у каждого агента → полный `/qa-gate` на интеграции (team-lead) → мерж по команде пользователя.

**СТАТУС ВОЛНЫ 1 (2026-07-03): ✅ ИСПОЛНЕНА И ВЕРИФИЦИРОВАНА.** Ветки: `fix/platform/wave1-quickfixes` (be-ver: verified true), `feat/bases/security-membership-fix` (be-ver+security: verified true, 11/11 эндпоинтов покрыты), `docs/wave1-analysis` (ba-ver: verified true, 14 US). Интеграция: `integration/wave1` — Docker full backend suite PASS (~187с, 0 fail), tsc PASS, Vitest 99/99; Biome-ошибки = локальные CRLF-артефакты Windows. FA-02-03/04 оказались закрыты ранее (a14aa20). NF-09 перенесён в Волну 2 (devops). **Мерж в master — за пользователем** (рекомендация: мержить `integration/wave1`; security-изменение bases — в списке личного ревью).

**Follow-ups из верификации (в Волну 2):**
- [MEDIUM] `VITE_GIPHY_API_KEY` не проброшен в prod-сборку фронтенда + обновить `wiki/DEPLOY.md` (→ devops, задача 2.15).
- [MEDIUM] bases: owner-only модель vs GZ-решение о project-membership sharing — дизайн-gap зафиксировать в эпике BS-1 (проект-шаринг = фича, не хотфикс).

## ВОЛНА 2 — ADR-пакет + Track P M0 старт + спеки (атомарная)

| # | Задача | Исполнитель | Верификатор | DoD |
|---|--------|-------------|-------------|-----|
| 2.1 | **ADR-AUTH-SSO** (OIDC+LDAP, org-toggle, break-glass, JIT+LDAP-sync, без MFA, маппинг только system-role) | arch | arch-ver + security-auditor | Принят; tz_constraints обновлён |
| 2.2 | **ADR-AI-agents/MCP** (ступенчатая автономия, agent-principal, BYOK hard-stop, PAT+curated toolset; hub-ai = LiteLLM-паттерн: универсальный OpenAI-адаптер, каталог моделей, virtual keys) | arch | arch-ver + security-auditor | Принят |
| 2.3 | Целевая C4 (уточнённый каталог: core + ai + ws/messenger/calls + notification) | arch | arch-ver | `docs/architecture.md` |
| 2.4 | M0-02: реестр NATS-subjects → `contracts/events/*.json` (спека) | sa1 | sa-ver | Все текущие subjects задокументированы версионно |
| 2.5 | M0-05: path-filtered CI (dorny/paths-filter + параметризованный Dockerfile) | devops-engineer | qa-auto-ver | PR в 1 домен = только его jobs; замер минут |
| 2.6 | M0-06: OTel trace-context через NATS-headers (outbox Worker + consumers) — **жёсткий gate пилота** | be1 | be-ver | Trace REST→outbox→NATS→consumer непрерывен |
| 2.7 | M0-04: K3s фундамент — namespaces, infra StatefulSets, Caddy ingress, NetworkPolicy (закрывает TD-04) | devops-engineer | security-auditor | Монолит поднимается в k3d/kind |
| 2.8 | FR/API-спека Intranet (по US из 1.10) | sa2 | sa-ver | Контракты полные, коды ошибок |
| 2.9 | FR/API-спека Calendar (по US из 1.11) | sa2 | sa-ver | — |
| 2.10 | FR/API/ERD-спека Wiki (по US из 1.12; space→page→version, restrict-флаг) | sa1 | sa-ver | ERD с UUID/version/soft-delete |
| 2.11 | DESIGN-SPEC: intranet-лента + calendar-страница + wiki-каркас | designer | designer-ver | Токены, состояния, a11y |
| 2.12 | **(BP-2026)** CI supply-chain: govulncheck + `go mod verify`/`tidy -diff` + gitleaks secrets-скан | devops-engineer | security-auditor | Jobs в CI, укладываемся в бюджет минут |
| 2.13 | **(BP-2026)** Versioned Atlas миграции + `atlas migrate lint` в CI (сейчас Schema.Create — drift; **предусловие schema-per-service M2**) | be1 | be-ver | `backend/ent/migrate/` каталог versioned; lint в CI |
| 2.14 | **(BP-2026)** Playwright CI-механика: шардинг + blob/merge-reports, `storageState` (логин 1 раз), `test.step()` | qa-auto2 | qa-auto-ver | e2e-full ускорен; отчёты мержатся |

**Кандидаты Волны 3+ из BP-2026** (tech-lead включает при нарезке): апгрейд Fiber v3 beta.3 → stable 3.x (переименования TrustProxy/ListenConfig — отдельный атомарный план + полная регрессия); `testing/synctest` для outbox-worker и WS-hub тестов; SKIP LOCKED в outbox-worker (до 2+ реплик); Nats-Msg-Id publish-дедуп; React Compiler пилот (за ADR); WCAG 2.2 AA как цель a11y (Dragging Movements для Kanban, Target Size 24px); INP ≤200ms бюджет в Lighthouse-gate.

## ВОЛНА 3 — Hardening M1 (Track P) + первые фичи (укрупнённо → tech-lead дробит перед стартом)

- 3.a M1-01/02: `*ent.Client` за repository-интерфейс: auth+rbac (be1), dashboard/backup/imex/storage/apigateway (be2) — закрывает TD-02.
- 3.b M1-03: `GetDB()` escape-hatch tracker + TD-08/09.
- 3.c M1-04: depguard-lint запрет cross-domain импортов + белый список → выжигание.
- 3.d M1-05/06: каталог cross-domain вызовов (sa1) + манифест владения данными `contracts/data-ownership.md` (sa2).
- 3.e FE: intranet-модуль — декомпозиция скелета + лента/опрос-в-посте (fe1); calendar — ресурсы+recurring UI (fe2).
- 3.f BE: intranet API (be по волне), calendar API (recurring/ресурсы/iCal-подписки).
- 3.g QA-конвейер: TC intranet+calendar (qa-manual → qa-manual-ver) — старт до реализации по спекам.

## ВОЛНА 4 — Intranet+Calendar завершение + M2 пилот подготовка

- 4.a Intranet: onboarding-шаблоны, kudos-лидерборд, coffee-матчинг (be+fe).
- 4.b Calendar: availability busy/free + отпуска, iCal export/подписки.
- 4.c Автотесты intranet+calendar по верифицированным TC (qa-auto1 API, qa-auto2 компонентные+E2E).
- 4.d M2-01/02: PG-схема `notification` + DB-user + Atlas expand-contract; отдельный Ent-подмодуль + `cmd/hub-notification`.
- 4.e Contract-схемы событий notification (JSON) + schema-тесты обеих сторон.

## ВОЛНА 5 — M2 пилот hub-notification (Track P) + Wiki старт (Track F)

- 5.a M2-04: K3s deploy + Caddy route-flip `/api/notifications*` + compose `--profile split`.
- 5.b M2-05: read-model user display-данных по событиям.
- 5.c M2-06/07: e2e поверх split-топологии + runbook + замер p95/CI-минут.
- 5.d **GATE: Go/No-Go** (композит: lead-time ↓, p95 ≤ +10%, 0 инцидентов 2 недели) — team-lead + product + пользователь.
- 5.e Wiki: Ent-схемы + space/page CRUD + RBAC space-level (be1); редактор+дерево страниц (fe1); TC wiki (qa-manual).

## ВОЛНЫ 6–8 — CRM+HR (Track F) + M3 hub-ai (Track P, после Go)

- 6.a CRM: companies + activity-timeline + импорт CSV/XLSX с пресетами + email-логирование через hub-mail.
- 6.b HR: user.manager_id + org chart; авто-начисление отпусков (accrual policy) + поле грейда + производственный календарь РФ; handbook = wiki-space (зависит от 5.e).
- 7.a M3: hub-ai вынос + **LiteLLM-паттерн**: универсальный OpenAI-совместимый адаптер (провайдер конфигом), каталог моделей + fallback, virtual keys/бюджеты hard-stop, внутренний OpenAI-endpoint.
- 7.b MCP-сервер: PAT со скоупами + curated read-heavy toolset (по ADR-AI-agents).
- 7.c Agent-native: agent-principal + ступенчатая автономия + audit on-behalf-of.
- 8.a Wiki завершение: версии/история, поиск (user-scope NF-03), связь с задачами, миграция handbook.
- 8.b Chat-UX AI: ветвление диалогов, сравнение моделей side-by-side (LibreChat-паттерн).

## ВОЛНЫ 9–11 — Accounting (своя, управленка) + SSO + realtime-трио (M4)

- 9.a Accounting волна 1: баланс, cash flow, PDF-инвойсы, AR aging; поле валюты в схеме; зарплата — проводки факта; legal-hold.
- 9.b 1C-sync расширение как экспорт-мост.
- 10.a SSO: OIDC + LDAP/AD bind, org-toggle, JIT + LDAP-sync деактиваций, break-glass (по ADR-AUTH-SSO); security-пакет тестов ≥80%.
- 11.a M4: realtime-трио — hub-ws (membership read-model в Redis), hub-messenger (Mongo split), hub-calls (LiveKit-домен); **Locust-профиль ДО старта** (500 WS baseline, см. TF-PY-LOAD).

## ЭПИК DOCS-EXCELLENCE — документация мирового уровня (решение пользователя 2026-07-03)

> Часть задач исполняется **сразу** (в шаге актуализации доки до коммита, помечены ⚡), остальное — docs-лейн параллельно волнам (владелец: docs-writer, верификация: product-ver/security-auditor по профилю).

| # | Задача (атомарная) | Содержание | Когда |
|---|--------------------|------------|-------|
| DOC-01 ⚡ | **LICENSE — проприетарная, «никому нельзя»** | All Rights Reserved: запрет использования/копирования/распространения/модификации без письменного разрешения владельца; + раздел Licensing в README. Согласовано с GZ-REL-14 (внутренний/проприетарный) | Сейчас (до коммита) |
| DOC-02 ⚡ | **README расширить максимально** | Что за продукт (super-app: модули с описанием), скриншоты-плейсхолдеры, архитектура (C4-обзор + микросервисный план), стек, quick start (dev), deployment (K3s), AI-команда (30 агентов, playbooks), тестовая пирамида, статус/roadmap, лицензия | Сейчас |
| DOC-03 ⚡ | **SECURITY.md + security-доки** | Политика уязвимостей (как сообщать), поддерживаемые версии, security-инварианты Hub (SSRF/timing/webhook/RBAC/rate-limits), hardening-guide для self-hosted (секреты, TLS, бэкапы), связка с `docs/security_guidelines.md` | Сейчас |
| DOC-04 ⚡ | **Тестовая стратегия тестфреймворка** | `docs/testframework_strategy.md`: пирамида по слоям (Go-юниты у be, TS-юниты у fe, Python pytest+allure для integration/API/smoke/regression/E2E, Locust для нагрузки), QA-конвейер (TC→verify→автотесты→verify), TC↔автотест трассируемость, Allure→Hub TMS, флейки/quarantine, coverage-gates per-service, тайминги/CI-бюджет | Сейчас (стратегия), реализация — фаза TF-PY |
| DOC-05 ⚡ | **Онбординг-листы по ролям (реальные люди)** | `docs/onboarding/`: чек-листы первого дня/недели для BE-разработчика, FE-разработчика, ручного QA, QA-автоматизатора, PM/тимлида, админа/ops — что прочитать (docs/playbooks), что поднять локально, первые задачи, к кому (какому агенту) обращаться | Сейчас |
| DOC-06 | **Wiki проекта — «пиздатое» состояние** | `docs/wiki/`: структура/навигация/индекс, актуализация API.md (286+ endpoints), QUICKSTART/SHORTCUTS/DEPLOY сверены с фактом, удаление drift, единый стиль | Docs-лейн, W1–W3 |
| DOC-07 | Продуктовая wiki-дока новых модулей | По мере волн: intranet/calendar/wiki-модуль/CRM/HR/accounting/bases — пользовательские гайды | По волнам |
| DOC-08 | CHANGELOG.md автогенерация + release notes на milestone | Согласовано GZ-REL-16 | С первого тега |

## ФАЗА TF-PY — Обновление тестового фреймворка (решение пользователя 2026-07-03; **после дробления на микросервисы**, Depends on: Track P M2–M4 завершены)

Целевое состояние тестовых слоёв:

| Слой | Язык/стек | Где живёт | Кто пишет |
|------|-----------|-----------|-----------|
| Unit backend | **Go** (как сейчас) — вместе с кодом | `backend/testframework/{domain}/` (или рядом с сервисом после выноса) | be1/be2 |
| Unit frontend | **TypeScript** (Vitest+RTL, как сейчас) — вместе с кодом | `testframework/stages/unit/frontend-test/` | fe1/fe2 |
| **Всё остальное** (integration/API/contract/smoke/regression/E2E) | **Python**: pytest + allure-pytest + httpx (API) + playwright-python (E2E) + pytest-xdist (параллелизм) | `testframework/` — раскидано по доменам/сервисам, по best practices тестирования | qa-auto1/qa-auto2 по верифицированным TC |
| **Нагрузочные** | **Python: Locust** (те же фикстуры/фабрики данных, что и pytest-слой) | `testframework/load/` — сценарии по доменам (WS connect+msg, auth, board API, Allure-import) | qa-auto1/2 |

Декомпозиция фазы (tech-lead дробит атомарно перед стартом):
- TF-PY-01: спека структуры Python-фреймворка (sa+qa-auto: раскладка по доменам/сервисам, фикстуры, conftest-иерархия, фабрики, POM для playwright-python, Allure-аннотации с TC-ID) — по best practices pytest.
- TF-PY-02: каркас: pyproject, pytest.ini, базовые фикстуры (auth/env/данные), allure-pytest конфиг, CI-job.
- TF-PY-03..NN: миграция по доменам — API-тесты каждого сервиса/домена (notification, ai, ws/messenger/calls, core-домены) — по одному атомарному плану на домен.
- TF-PY-E2E: перенос Playwright-спеков TS → playwright-python (POM/steps), сохранение smoke/full-разбиения.
- TF-PY-LOAD: **Locust**-сценарии в `testframework/load/` (WS connect+msg, auth-flow, board API, messenger throughput, Allure-import 1000 результатов) с целями из SLO (p95<200ms API, WS<100ms, 500+ concurrent WS); nightly/pre-release прогоны, тайминги в COVERAGE.md. **Примечание:** baseline-профиль Locust заводится РАНЬШЕ фазы TF-PY — до старта выноса realtime (заменяет k6 из GZ-QA-10; Locust самостоятелен и не ждёт остального фреймворка).
- TF-PY-ALLURE: сквозная привязка TC↔автотест через allure-pytest (tms-link/label ASID) + заливка в Hub TMS (webhook) — фреймворк сам отчитывается в наш TMS.
- Ворота: паритет покрытия сценариев со старым слоем до удаления TS-E2E/Go-integration дублей; тайминги прогонов зафиксированы; CI-бюджет соблюдён.

Обоснование: Python — стандарт QA-автоматизации (нанимаемость ручников/автоматизаторов), allure-pytest — самая зрелая Allure-интеграция (стыкуется с TMS Hub напрямую), отделение QA-слоя от языков продакшена соответствует модели «юниты пишут разработчики, остальное — QA-конвейер».

## ВОЛНА 12+ — Bases-эпик финал + polish

- Bases BS-0..5 идёт **параллельно на выделенном be-потоке** с Волны 3 (30 функций, 4 views, лимиты 50k/200, realtime BS-5, attachments): финальная интеграция и E2E.
- UX-polish эпик D (skeleton/empty/optimistic/RHF+Zod), Lighthouse ≥90 gate, WCAG критичные пути, notification digest, custom fields UI, Gantt pro, TMS-отчёты.
- BI-builder — спека (зависимость CRM-аналитики, решение GZ-V4-11).
- Заморожено: marketplace/status, tracker-вынос, SAML, multi-org.

---

## Отложено: July Rebuild (~конец июля 2026) — НЕ стартовать код до align

> Brief: `.planning/archive/july-2026/JULY-2026-REBUILD-BRIEF.md` · Status: `PROJECT-STATUS-2026-07-10.md` · Ops plan: `ROADMAP-DETAILED-JULY-RETURN.md`
> Seeds: SEED-001 **REJECTED** · SEED-002 **active**
> Edge catalog / Testing Guide: `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`, `testframework/TESTING_GUIDE.md`

### EPIC-SFU — REJECTED

| Поле | Значение |
|------|----------|
| Решение | **REJECTED 2026-07-10** — остаёмся на **LiveKit** |
| Seed | `seeds/SEED-001-mediasoup-migration.md` |
| M4 | `hub-calls` = **LiveKit-домен** (без смены SFU) |

### EPIC-CAL-SYNC — Unified calendar + RSVP + mail (L0/L1)

| # | Подзадача | Who | When | Depends on | Success criteria |
|---|-----------|-----|------|------------|------------------|
| CAL-0 | Align: mail L0–L3, reschedule-RSVP, приоритет vs Redesign | product + user | W0 D1–D2 | Brief §5 | Locked answers |
| CAL-1 | US/AC RSVP + create from messenger/calendar | ba2 → ba-ver | W1–W2 | CAL-0 | US в vision; AC тестируемы |
| CAL-2 | ERD CalendarEvent/Attendee + REST + NATS + ICS | sa2 → sa-ver | W2 | CAL-1 | Контракты + коды ошибок |
| CAL-3 | ADR-CAL-UNIFIED (+ mail ADR если L2+) | arch → arch-ver | W2 | CAL-2 draft | ADR accepted |
| CAL-4 | DESIGN-SPEC: invite inbox, accept/decline, conflict | designer → designer-ver | W2–W3 | CAL-0 + mockups | Токены, a11y |
| CAL-5 | BE: schema + RSVP API | be2 → be-ver | W3+ | CAL-3 | Unit/API green |
| CAL-6 | BE: messenger/calendar → event + NATS + ICS L0 | be1 → be-ver | W3+ | CAL-5 | ICS в письме |
| CAL-7 | FE: inbox + RSVP + calendar projection | fe2 → fe-ver | W3+ | CAL-4, CAL-5 | Accept → в «моём» календаре |
| CAL-8 | FE: create from messenger | fe1 → fe-ver | W3+ | CAL-6 | Happy path E2E |
| CAL-9 | iCal L1 export/subscribe (GZ-V4-05) | be+fe | W4 | CAL-5 | Feed URL; no OAuth |
| CAL-10 | TC C1–C15 → auto | qa-manual → qa-auto | W2–W3+ | EDGE catalog | TC↔тест |

**Entry:** July Align · Wave1 merged (рекоменд.) · CAL-0 locked. **Не делать до ADR:** OAuth L3, guest email-only, mediasoup.

### EPIC-REDESIGN — Cloud Code Designer → Hub UI

| # | Подзадача | Who | When | Depends on | Success criteria |
|---|-----------|-----|------|------------|------------------|
| RD-0 | Scope: full shell vs calendar/calls/messenger | product + user | W0 | Align | MoSCoW |
| RD-1 | Inventory макетов vs UI (gap list) | designer | W1 | RD-0 + mockups | Gap MD |
| RD-2 | DESIGN-SPEC токены/состояния/a11y | designer → designer-ver | W1–W2 | RD-1 | Spec для fe |
| RD-3..N | FE waves по экранам | fe1/fe2 → fe-ver | W3+ | RD-2 | Visual + a11y |
| RD-Q | TC + visual smoke E2E | qa-manual + qa-auto2 | ∥ RD-3+ | RD-2 | Smoke green |

### EPIC-QA-HARDENING — каталог → TC → автотесты + Testing Guide

| Фаза | Планы | Who | When | Success criteria |
|------|-------|-----|------|------------------|
| **QH-P0** | Каталог + Testing Guide + Charter | qa-manual / docs / tech-lead | 2026-07-10 → W0 | Артефакты в INDEX |
| **QH-P1** | Adoption Guide в PR/playbooks | docs-writer + team-lead | W0–W1 | Ссылка в checklist |
| **QH-P2** | TC из catalog (auth/ws/calls/calendar/messenger) | qa-manual → qa-manual-ver | W1–W2 | TC-ID + трассировка |
| **QH-P3** | Auto backend API/integration | qa-auto1 → qa-auto-ver | W2+ | TC link; race green |
| **QH-P4** | Auto FE Vitest + Playwright | qa-auto2 → qa-auto-ver | W2+ | Deterministic |
| **QH-P5** | Gate + flaky quarantine | team-lead + qa-auto-ver | Каждая волна | `/qa-gate` PASS |

### When / Entry — волны 2–12 (ориентир после July Align)

| Волна | When | Entry criteria |
|-------|------|----------------|
| **2** | После Align + merge Wave1; не раньше возврата ~конец июля | wave1 в master · Align Qs · tech-lead нарезка · NF-09 + GIPHY follow-ups |
| **3** | После W2 ADR + спеки 2.8–2.11 | ADR drafts · Atlas plan · DESIGN-SPEC 2.11 · TC start |
| **4** | После каркаса intranet/calendar из W3 | US-CAL/INTR · M2 schema plan · iCal без OAuth |
| **5** | После W4 slice + M2 prep | K3s 2.7 · OTel 2.6 · **Go/No-Go** с user |
| **6–8** | После Go на M2 (для M3); handbook ← Wiki 5.e | Go/No-Go PASS · CRM/HR US · Wiki RBAC |
| **9–11** | Accounting → SSO → M4; **Locust baseline до M4** | ADR-AUTH-SSO · Locust WS · **hub-calls = LiveKit** |
| **12+** | Bases ∥ с W3; polish после F-трека | BS-1 sharing design · Lighthouse/WCAG цели |

## ВОЛНА / ЭПИК QA-HARDENING (параллельный лейн с возврата июля)

> Не путать с TF-PY (после микросервисов). Ужесточение текущего Go/TS/Playwright контура.
> Фазы QH-P0…P5 выше · артефакты в `INDEX-JULY-2026.md`.
> Параллелизм с Track F: QH на доменах текущей фича-волны; полный `/qa-gate` — на интеграции (team-lead).

---

## Сквозные ворота каждой волны

1. Максимальная декомпозиция планов до старта (tech-lead → приёмка team-lead).
2. Агент: регрессия своего домена; полный `/qa-gate` — на интеграции волны.
3. QA-конвейер: TC → qa-manual-ver → автотесты → qa-auto-ver; юниты пишут be/fe.
4. Coverage нового кода ≥82% (MCP/agents/SSO ≥80% security-critical).
5. Docs-sync в той же волне; TC-таблицы полные для новых модулей.
6. CODE-QUALITY-CHARTER + Testing Guide в DoD be/fe/qa.
7. Мерж в master — только по команде пользователя (личное ревью: security/миграции/CI/ADR).

## Что осталось (статус 2026-07-10)

| # | Шаг | Статус |
|---|-----|--------|
| 1 | Волна 1 в фича-ветках | ✅ merge-ready (`integration/wave1`) |
| 2 | Мерж Wave1 → master | ⏳ **пользователь** (ревью bases) |
| 3 | FA-06 secret rotation | ⏳ manual ops |
| 4 | July Align (Brief §5) + kickoff | ⏳ ~конец июля |
| 5 | QA-HARDENING P0 (catalog + guide) | 🟡 planning 2026-07-10 |
| 6 | CAL-SYNC / Redesign / W2+ execute | ⏸ после Align |
| 7 | Tag v2.5.0 + SSH deploy | ⏸ после FA-06 + CTO |
