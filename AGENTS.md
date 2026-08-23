# Hub — AI Agent Workspace

> **FOUNDER MODEL ROUTING LOCK (amended 2026-08-22; highest precedence):** only the
> root `team-lead`/orchestrator may use `gpt-5.6-sol`, always with `max`
> reasoning. Every subagent dispatch MUST set both `model` and
> `reasoning_effort` explicitly; omitted values and `inherit` are invalid.
> All downstream work uses `gpt-5.6-terra` with
> task-routed effort: `low` for deterministic mechanics/hash/provenance,
> `medium` for ordinary analysis/authoring, `high` for independent review, and
> `xhigh` for unusually complex downstream audits and every atomic code
> implementation. Atomic implementation starts only after root/Terra freezes
> the contract and remains limited to one function, small module, or single
> behaviour with exact files, tests and acceptance criteria;
> architecture/security/product choices remain with root/Terra. If the task
> does not fit, the Terra implementer returns `BLOCKED` for
> further decomposition. Sol is forbidden for subagents. Unavailable runtime
> means `UNKNOWN/BLOCKED`; silent model or effort fallback is forbidden.

> **FOUNDER TASK-LOCAL ROUTING EXCEPTION (2026-08-13; newer and narrowly
> scoped):** only on branch `codex/mail-module-20260813`, for Mail workstream
> Phases 143–148, atomic production/test-code packets use
> `gpt-5.6-terra` with `xhigh` and no fallback. An explicit native
> collaboration dispatch that sets the exact model and effort, is accepted by
> the dispatcher, and runs without rejection is sufficient provenance here:
> record `provenance_source: dispatcher-enforced`. If actual telemetry is
> exposed, it must match; a mismatch, rejection, omitted model/effort or
> fallback remains `BLOCKED`. Independent review and every other acceptance
> gate remain required. Root remains `gpt-5.6-sol`/`max` as
> orchestrator/reviewer and does not own atomic code. This overrides only the
> Spark sentence above for this branch/workstream; every other routing, scope,
> protected-zone and verification rule remains unchanged. Evidence:
> `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md`.

> **Читай первым.** Регламент: `.cursor/REGULATION.md` | База знаний: `docs/README.md`

---

## Проект

**Hub** — self-hosted **corporate super-app** для команд 10–50 человек: Task Tracker + Messenger + TMS + Voice Calls + Dashboards **+** интранет, календарь, CRM-lite, HR, учёт, no-code базы, AI hub.

**Roadmap v4:** `.planning/milestones/v4-super-app/SUMMARY.md` · Vision: `.planning/decisions/specs/PLATFORM-VISION-v4.md`

Монорепо: `backend/` (Go, **87 Ent schemas**) + `frontend/` (React) + `cli/` + `deployments/` + **`testframework/`** (единый greenfield Python QA-проект в корне каталога)

### Целевое расположение тестов (founder greenfield cutover)

| CI Job | Стадия | Путь | Домены |
|--------|--------|------|--------|
| `backend-unit` | unit | `backend/tests/{domain}/` | только genuine isolated Go units |
| `frontend-unit` | unit | `frontend/tests/{domain}/`, `frontend/tests/support/` | только genuine Vitest/RTL units |
| `python-contract` | framework | `testframework/selftests/` | install/package/framework contracts; no SUT product behavior in the minimal slice |
| `python-api` | API | `testframework/tests/api/` | публичное поведение backend/frontend BFF |
| `python-integration` | integration | `testframework/tests/integration/` | DB/broker/storage boundaries |
| `python-e2e-smoke/full` | e2e | `testframework/tests/e2e/` | `smoke` is a pytest marker; `full` is a launcher profile |
| `python-load` | load | `testframework/load/` | Locust profiles |

**Founder greenfield cutover (LOCKED, 2026-08-07):** переносим только
доказанные genuine units в `backend/tests/**` и `frontend/tests/**`. Все legacy
Go component/API/integration, static-contract Vitest, TypeScript Playwright,
их support/config/entrypoints удаляются по exact manifest и **не портируются**.
Non-unit QA создаётся с нуля как единый Python-проект прямо в `testframework/`:
reusable package code принадлежит `testframework/src/hub_testkit/**`, а
исполняемые pytest suites — `testframework/tests/{contract,api,integration,e2e}/**`.
Отдельные roots `selftests/`, `cases/`, `schemas/`, `policies/`, `load/` и
`reports/` не смешиваются ни с package, ни с SUT suites. Будущие
`testframework/pyproject.toml`, `uv.lock` и `src/hub_testkit/py.typed` описаны
документационным `testframework/PACKAGE-CONTRACT.md`, но ещё не реализованы.
Row-level parity и
сохранение исторического 82% Go coverage не являются migration gate. Новый
набор принимается по traceability `task/requirement/TC`, risk-based design и
явным happy/negative/authz/boundary/edge/corner/failure-recovery сценариям. План:
`.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md`.

---

## Маршрутизация контекста (`docs/`)

Всегда начинай с `AGENTS.md`, `.planning/STATE.md` и `docs/README.md`.
Затем загружай **только документы, влияющие на текущую задачу**, по таблице
ниже. Полный набор нужен для cross-domain аудита/архитектурной приёмки, но
вреден для точечной правки: раздувает контекст и увеличивает drift.

| Файл | Когда нужен |
|------|-------------|
| `docs/vision_and_scope.md` | Бизнес-контекст, User Stories |
| `docs/system_specification.md` | API, модели данных |
| `docs/architecture.md` | Модули, потоки, deployment |
| `docs/test_plan.md` | Тесты, регрессия |
| `docs/stack_and_tools.md` | Версии, инструменты |
| `docs/code_style_guide.md` | Стиль Go/TS |
| `docs/security_guidelines.md` | OWASP, blockers |
| `docs/tz_constraints.md` | TZ mandatory/forbidden |
| `docs/competitive_analysis_2026.md` | Конкуренты + Mighty/Bitrix/Lark |
| `docs/competitive_audit.md` | Best practices + backlog |
| `docs/user_needs_analysis_2026.md` | Персоны, JTBD, MoSCoW |
| `docs/GLOSSARY.md` | Термины, ID-конвенции (TC-/FR-/GZ-/ECC-…), модули backend |
| `docs/TROUBLESHOOTING.md` | Типовые грабли разработки (сборка, тесты, окружение) |
| `docs/runbooks/README.md` | Эксплуатация прода: инцидент, алерты, деплой/откат, БД, бэкапы |

Планирование: [`.planning/README.md`](.planning/README.md) (навигация) · `.planning/STATE.md` · `.planning/COVERAGE.md` · `.planning/ROADMAP.md` · `.planning/archive/roadmaps/DISCUSS-ROADMAP-LOCKED.md`

**Protected external zone (2026-08-07):** `pro-qa-ai/**` — личный QA-AI toolkit
founder, эксклюзивный исполнитель Claude Code Opus 5. Внутренние агенты Hub
папку не читают по умолчанию и не меняют; конкретный результат принимается
только через явный founder handoff. Правило: `.cursor/rules/70-pro-qa-ai.mdc`.

Agent safety: `.cursor/rules/07-agent-safety.mdc` — authority/egress, secret
incident, dependency supply-chain, dirty-worktree handoff и provenance gate.

GSD + агенты: `.cursor/GSD-INTEGRATION.md`, `.cursor/REGULATION.md` §8 (GSD), §9 (test triage), §10 (docs sync)

---

## AI-команда (32 файла агентов)

> Реорганизована 2026-07-02 под параллельную разработку. Ростер: **`.cursor/playbooks/README.md`** · RACI: `.cursor/playbooks/team-workflow.md` · модели: **`.planning/decisions/ADR-MODELS-GPT56-2026.md`**.
> **Модели (founder lock, amended 2026-08-22):** `team-lead` — единственный **GPT-5.6 Sol / Max** слот (`model: gpt-5.6-sol`). Все downstream-задачи явно используют Terra (`gpt-5.6-terra`, `low|medium|high|xhigh`): анализ/декомпозиция/авторинг получают task-routed effort, независимый review — `high`, а атомарный код по замороженному контракту — всегда `xhigh`. Каждый downstream dispatch обязан явно задать model+effort; `inherit`, пропуск параметра, Sol-субагенты и silent fallback запрещены. `codex-micro` сохраняется только как стабильный legacy ID роли и также запускается на `gpt-5.6-terra` / `xhigh`. Полная матрица и dispatch guard: **`.cursor/playbooks/model-routing.md`**. `team-lead-opus` — deprecated redirect.
> **Live runtime inventory (2026-08-08):** `.cursor/agents/` = **32 файла: 31 operational + 1 deprecated redirect**; GSD runtime = **34 typed agents** (`~/.codex/agents/gsd-*.toml`) + **71 skills** (`~/.agents/skills/gsd-*`). Superpowers установлен как workflow plugin, но не является агентом/authority. Claude Code Opus 5 остаётся отдельным founder-operated worker только для `pro-qa-ai/**`.

### Лидерство / продукт
`team-lead` (оркестратор, Sol / Max), `tech-lead` (Terra / High–XHigh), `product` + `product-ver`, `codex-micro` (legacy role ID; Terra / XHigh, атомарный код)

### Аналитика / архитектура / дизайн (потоки + верификаторы)

| Роль | Агенты | Verifier |
|------|--------|----------|
| Business Analyst | `ba1`, `ba2` | `ba-ver` |
| System Analyst | `sa1`, `sa2` | `sa-ver` |
| Architect | `arch` | `arch-ver` |
| Designer | `designer` | `designer-ver` |

### Разработка / QA (потоки + кросс-ревью + верификаторы)

| Роль | Агенты | Verifier |
|------|--------|----------|
| Backend | `be1`, `be2` | `be-ver` |
| Frontend | `fe1`, `fe2` | `fe-ver` |
| Manual QA | `qa-manual` | `qa-manual-ver` |
| Auto QA | `qa-auto1`, `qa-auto2` | `qa-auto-ver` |

### Cross-cutting

`devops-engineer`, `security-auditor`, `debugger`, `code-reviewer`, `docs-writer`

### Practices playbooks (`.cursor/playbooks/`)

`backend.md` · `frontend.md` · `design.md` · `testing.md` (Allure + замер времени + MD↔автотесты) · `observability.md` · `code-comments.md` · `git-workflow.md` · `analysis.md` · `verification.md` · `team-workflow.md` · **`model-routing.md`** · **`prompt-contract.md`** (контракт промпта: приём задачи, формат, техники, температура, 10-этапная верификация, матрица ролей — обязателен всем)

---

## GSD Integration

Hub использует **GSD skills** (`~/.agents/skills/gsd-*`) + typed GSD agents (`~/.codex/agents/gsd-*.toml`) + Hub slash-skills (`.cursor/skills/`).

| Документ | Назначение |
|----------|------------|
| `.cursor/GSD-INTEGRATION.md` | Матрица агент → GSD skills |
| `.cursor/skills/test-triage/SKILL.md` | FAIL теста: функционал → код vs тест |

**Test triage (обязательно):** при падении теста сначала проверить корректность функционала по `docs/`, только потом чинить код или тест. REGULATION §9 (детальные сценарии — §14, §18).

**Docker-only execution (founder lock, 2026-08-07):** агентам запрещено
запускать на Windows host любые tests, validators, linters, build/vet, Vite,
Playwright, pytest/Locust или тестовые backend/frontend процессы. Все такие
команды выполняются только в task-owned Docker containers/Compose с CPU/RAM
limits. Host используется лишь для чтения/редактирования файлов и Git. Любое
старое host-run evidence неканонично и должно быть повторено в Docker.

**Docs sync (обязательно):** после любого изменения кода обновляй затронутую документацию в той же сессии. Правило: `.cursor/rules/55-docs-sync.mdc` · skill: `/handoff-sync`.

---

## Текущий статус (2026-08-22)

> July Rebuild planning · LiveKit канон · метрики ниже = baseline master (без регрессии).
>
> Текущая product-truth (founder, 2026-08-12): Hub остаётся незавершённым MVP. Ни один продуктовый модуль не принят как complete или stable. Метки SHIPPED/COMPLETE ниже фиксируют только историческое выполнение конкретных milestone, audit или technical work и не являются заявлением о готовности продукта, модулей либо production.

| Milestone | Статус |
|-----------|--------|
| v1.0 MVP | ✅ SHIPPED 2026-03-21 |
| v2.0 Hub Redesign | ✅ COMPLETE 2026-05-15 |
| v2.5 Trust | ✅ COMPLETE 2026-06-22 |
| v2.6 Automate + Arch + OBS | ✅ COMPLETE 2026-06-24 |
| v3.0 AI Native | ✅ COMPLETE 2026-06-22 |
| v4 Super-app | ✅ COMPLETE 2026-06-25 |
| Coverage wave (Ph 59–125) | ✅ COV **83.0%** (2026-06-22 Docker baseline) |
| Test framework migration | ✅ Wave B COMPLETE (0 internal tests) |
| Final audit (FA-01–FA-12) | ✅ Wave C COMPLETE (кроме FA-06 manual) |
| Lighthouse CI (FB-06) | ✅ artifact job PR #27 (non-blocking) |
| Волна 1 (roadmap 2026-07-03) | ✅ merged to `master` by user 2026-07-11 |
| July Rebuild / QA-HARDENING | 📋 planning · kickoff ~конец июля |
| AI quality layer / GSD / portable kit | **Touchstone** `0.4.0-beta.3` foundation-ready pre-release · workflow contracts `30f75a25` · `STRUCTURAL_PASS` · paired eval **NOT_RUN** |
| Test architecture rebase (Ph 128–141) | ✅ **MINIMAL MERGE CANDIDATE (2026-08-23):** backend/frontend units and 181 framework selftests pass in Docker; five SUT contract tests/cases were removed by founder direction; API/integration/e2e/load are outside this merge gate; hosted CI is off. Publication still requires final Git/remote verification. |

### Метрики качества

| Метрика | Значение |
|---------|----------|
| COV overall | **82.0373%** ✅ (2026-08-07 Docker Go 1.25 baseline) |
| Security packages ≥80% | **6/6** ✅ (2026-08-07 Docker Go 1.25 baseline) |
| auth / ws / rbac / audit / calls | 80.6110% / 92.0993% / 87.0801% / 85.6879% / 82.6698% (2026-08-07) |
| tracker / messenger / notification / tms / project | 82.4% / 82.3% / 82.5% / 82.0% / 82.0% (2026-06-22 Docker baseline) |
| Product score | **89/100** |
| REST API | **286+** endpoints |
| Ent schemas | **82** |
| Test-rebase ledger | **1,049** source rows · SHA-256 `83b6a311…` (source revision `b9f690da`, measured 2026-08-07, ротация 2026-08-13) |
| Backend tests | Historical ledger: **832** total / **3,674** declared; the Variant A cutover deleted the remaining **245** hook-dependent legacy `backend/testframework/**/*_test.go` paths. Current runnable units live only under `backend/tests/**` (runtime evidence is tracked separately). |
| Vitest | **180** files / **519** literal cases: **156** frontend unit + **24** static contract; **3** additional frontend support rows |
| Playwright E2E | **20** specs / **53** declared cases + **14** E2E support rows |
| FA-10 archived legacy gate | archival record: build ✅ · hubtest ✅ · race ✅ · Biome ✅ · e2e-smoke **3/3** ✅ |

### Следующие шаги

| Gap | Статус |
|-----|--------|
| FA-01 audit + remediation backlog | ✅ closed 2026-06-22 |
| Test framework migration TF-M04–M08 | ✅ Wave B COMPLETE |
| Wave C final audit FA-01–FA-12 | ✅ COMPLETE (кроме FA-06 manual) |
| FA-10 e2e-smoke | ✅ 3/3 PASS |
| PG-03 Secret rotation (FA-06) | ⏳ manual ops — **отдельный ops-блокер production cutover; не индикатор готовности MVP** |
| Test-rebase Phase 128/129 | ⏳ `PUBLICATION_PENDING`: local minimal Docker matrix green and deletion receipts accepted; API/integration are not gates for this slice; final fetch/fast-forward verification remains |
| Merge Волна 1 → master | ✅ user merged 2026-07-11 |
| July return / CAL-SYNC / QA-HARDENING / Redesign | 📋 planning → kickoff ~конец июля |
| AI-layer prompt research / paired pilot / portable release | ⏳ `0.4.0-beta.3` on Hub + Desktop, `STRUCTURAL_PASS`; paired effectiveness/ROI eval NOT_RUN |
| TF-CLEANUP-01 legacy e2e/specs | ✅ closed 2026-06-22 |
| RU-COMMENTS-01 top files | ✅ closed 2026-06-22 |
| DOC-02 historical SUMMARY banner | ✅ closed 2026-06-22 |
| SSH deploy + v2.5.0 tag | ⏳ после FA-06 + CTO sign-off |
| NFR-16 WCAG full sign-off | ⚠️ login axe PASS |
| NFR-17 Lighthouse blocking gate | ⚠️ artifact only (PR #27) |

Полный аудит: `.planning/audits/2026-06/AUDIT-2026-06-23.md` · матрица: `.planning/COVERAGE.md` · July index: `.planning/archive/july-2026/INDEX-JULY-2026.md`

---

## Workflow

```
1. Загрузить docs/*
2. ba → sa → arch (для новых фич)
3. be/fe → атомарная реализация + тесты
4. Цикл качества нового + регрессия старого (rule 56) — до green
5. Docs sync (STATE + domain docs)
6. *-ver → верификация (verified: true в skill)
7. qa → применимые exact Compose services (Docker-only)
8. evidence/triage → финальный чек
```

`team-lead` делегирует субагентам **только** по шаблону §5 (`team-lead.md` / `team-workflow.md` §5.3.1).

**Git / коммиты:** **запрещено** писать `Co-authored-by` / AI attribution (Cursor, Claude, Codex, Cloud, Copilot) — все модели и агенты. См. `.cursor/rules/51-no-ai-coauthor.mdc`, `.cursor/rules/50-git.mdc`. Хук: `.githooks/commit-msg` → `git config core.hooksPath .githooks`.
---

## Команды

- Dev stack: `docker compose -f docker-compose.local.yml up -d`.
- Единственный QA API: `docker compose -f docker-compose.local.yml run --rm <exact-service>`.
  Разрешены только `backend-unit`, `frontend-unit`, `python-contract`, `python-api`,
  `python-integration`, `python-e2e-smoke`, `python-e2e-full` и `python-load`.
  QA services profile-explicit и disposable; integration/e2e используют изолированную
  `qa-internal` инфраструктуру. Host не запускает test/build/lint/validator процессы.
- Historical legacy QA entrypoints are archived and are not a fallback. Use only
  `docker compose -f docker-compose.local.yml run --rm <service>` with an exact listed service.

---

*Обновлено: 2026-08-23 — минимальный unit/framework slice проверен в Docker; hosted CI выключен; публикация зависит только от финального Git/remote fast-forward audit.*
