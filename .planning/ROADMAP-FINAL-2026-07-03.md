# Hub — Финальный Roadmap (2026-07-03)

> **Статус:** на проверку пользователю → после апрува коммит → старт Волны 1 в фича-ветках.
> **Входы:** `GRAY-ZONES-DECISIONS-2026-07-03.md` (все решения), `ADR-ARCH-MICROSERVICES-2026`, `MICROSERVICES-ROADMAP.md`, `EXPANSION-ANALYSIS`, `MARKET-AND-NEEDS`, A/B-планы (`ab-tests/`, `ab-microservices/`).
> **Принцип:** rolling-wave — ближние волны декомпозированы атомарно (≤200–300 строк/план), дальние укрупнённо; `tech-lead` дробит их до предела перед стартом (принцип максимальной декомпозиции, team-workflow §6.1).
> **Процесс:** агенты коммитят в фича-ветки свободно (и пушат ветки на remote); **push/мерж в master — только пользователь лично**; PR=фича, rebase-merge; QA-конвейер: TC → qa-manual-ver → автотесты → qa-auto-ver; юниты пишут be/fe.

---

## Общая картина: два трека + новые модули

```
Track F (функционал):  W1 quick-wins → W2-3 аналитика+wiki-спека → W4-6 intranet+calendar
                       → W7-8 crm+hr → W9-10 accounting → bases-эпик (BS-0..5, параллельно на be-потоке)
                       → wiki-реализация → AI-hub (LiteLLM-паттерн) + MCP → SSO
Track P (платформа):   W2 M0 (ADR-детали+K3s+OTel+path-CI) → W3-4 M1 hardening
                       → W5-6 M2 пилот hub-notification → GATE Go/No-Go
                       → M3 hub-ai → M4 realtime-трио (hub-ws/hub-messenger/hub-calls)
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

## Сквозные ворота каждой волны

1. Максимальная декомпозиция планов до старта (tech-lead → приёмка team-lead).
2. Агент: регрессия своего домена; полный `/qa-gate` — на интеграции волны.
3. QA-конвейер: TC → qa-manual-ver → автотесты → qa-auto-ver; юниты пишут be/fe.
4. Coverage нового кода ≥82% (MCP/agents/SSO ≥80% security-critical).
5. Docs-sync в той же волне; TC-таблицы полные для новых модулей.
6. Мерж в master — только по команде пользователя (личное ревью: security/миграции/CI/ADR).

## Что осталось ДО коммита (текущий статус)

| # | Шаг | Статус |
|---|-----|--------|
| 1 | Best practices 2026 (2 фоновых исследования: backend, frontend/testing) | ⏳ в работе → интеграция дельт в playbooks |
| 2 | Актуализация всей доки (docs/*, wiki, STATE, ROADMAP-дрейф) | ⏳ после п.1 |
| 3 | Этот roadmap — проверка пользователем | ⏳ **ждёт тебя** |
| 4 | Коммит всего пакета (30 агентов, 10 playbooks, ADR, планы, правила) | после твоего апрува |
| 5 | Старт Волны 1 в фича-ветках | после коммита |
