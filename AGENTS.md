# Hub — AI Agent Workspace

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
| `python-contract` | contract | `testframework/contract/` | API/schema/event/source contracts |
| `python-api` | API | `testframework/api/` | публичное поведение backend/frontend BFF |
| `python-integration` | integration | `testframework/integration/` | DB/broker/storage boundaries |
| `python-e2e-smoke/full` | e2e | `testframework/e2e/` | markers `smoke`/`full` |
| `python-load` | load | `testframework/load/` | Locust profiles |

**Founder greenfield cutover (LOCKED, 2026-08-07):** переносим только
доказанные genuine units в `backend/tests/**` и `frontend/tests/**`. Все legacy
Go component/API/integration, static-contract Vitest, TypeScript Playwright,
их support/config/entrypoints удаляются по exact manifest и **не портируются**.
Non-unit QA создаётся с нуля как единый Python-проект прямо в `testframework/`:
`{core,api,integration,contract,e2e,load,docker,reports}/**`,
`testframework/pyproject.toml` и корневой pytest config. Row-level parity и
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
> **Модели (ADR-MODELS-GPT56, amended 2026-08-07):** `team-lead` — **GPT-5.6 Sol / Extra High** (`model: gpt-5.6-sol`; Ultra только для большой делимой задачи). Downstream frontmatter = `inherit`, но defaults явные: Terra — основная работа, Luna — механика, Sol — high-risk. Полная матрица 32 ролей, effort, цены и fallback: **`.cursor/playbooks/model-routing.md`**. `team-lead-opus` — deprecated redirect.
> **Live runtime inventory (2026-08-08):** `.cursor/agents/` = **32 файла: 31 operational + 1 deprecated redirect**; GSD runtime = **34 typed agents** (`~/.codex/agents/gsd-*.toml`) + **71 skills** (`~/.agents/skills/gsd-*`). Superpowers установлен как workflow plugin, но не является агентом/authority. Claude Code Opus 5 остаётся отдельным founder-operated worker только для `pro-qa-ai/**`.

### Лидерство / продукт
`team-lead` (оркестратор, Sol / Extra High), `tech-lead` (декомпозиция), `product` + `product-ver`, `codex-micro` (мелочь)

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

## Текущий статус (2026-08-08)

> July Rebuild planning · LiveKit канон · метрики ниже = baseline master (без регрессии).

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
| AI quality layer / GSD / portable kit | **Touchstone** Standard `0.4.0-beta.2` pre-release · hardening `65b6a60d` · structural Docker **63/63 PASS** · prompt-engineering research + final portable release + paired eval **NOT_RUN** |
| Test architecture rebase (Ph 128–141) | 🚧 Phase 128 IN PROGRESS · TASK-202 archived/verified · TASK-203 + TASK-224 baselines closed · TASK-204 independent verification pending · no tests moved |

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
| Test-rebase ledger | **1,049** source rows · SHA-256 `72de6a81…` (source revision `5708671f`, measured 2026-08-07) |
| Backend tests | **832** total / **3,674** declared: **831 / 3,672** в legacy `backend/testframework/` + **1 / 2** server test; taxonomy total = **595** genuine unit + **237** component (**31** API + **206** integration) |
| Vitest | **180** files / **519** literal cases: **156** frontend unit + **24** static contract; **3** additional frontend support rows |
| Playwright E2E | **20** specs / **53** declared cases + **14** E2E support rows |
| qa-gate (FA-10) | build ✅ · hubtest ✅ · race ✅ · Biome ✅ · e2e-smoke **3/3** ✅ |

### Следующие шаги

| Gap | Статус |
|-----|--------|
| FA-01 audit + remediation backlog | ✅ closed 2026-06-22 |
| Test framework migration TF-M04–M08 | ✅ Wave B COMPLETE |
| Wave C final audit FA-01–FA-12 | ✅ COMPLETE (кроме FA-06 manual) |
| FA-10 e2e-smoke | ✅ 3/3 PASS |
| PG-03 Secret rotation (FA-06) | ⏳ manual ops — **единственный prod-блокер** |
| Test-rebase Phase 128 | 🚧 TASK-202 archived/verified; TASK-203 historical baseline + TASK-224 E2E baseline closed (**48 pass / 5 expected skip / 0 fail**); TASK-204 verification pending |
| Merge Волна 1 → master | ✅ user merged 2026-07-11 |
| July return / CAL-SYNC / QA-HARDENING / Redesign | 📋 planning → kickoff ~конец июля |
| AI-layer prompt research / paired pilot / portable release | ⏳ `0.4.0-beta.2` pre-release; structural Docker 63/63 PASS; prompt research and paired eval NOT_RUN |
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
7. qa → применимые target jobs `/qa-gate` (Docker-only)
8. /qa-gate → финальный чек
```

`team-lead` делегирует субагентам **только** по шаблону §5 (`team-lead.md` / `team-workflow.md` §5.3.1).

**Git / коммиты:** **запрещено** писать `Co-authored-by` / AI attribution (Cursor, Claude, Codex, Cloud, Copilot) — все модели и агенты. См. `.cursor/rules/51-no-ai-coauthor.mdc`, `.cursor/rules/50-git.mdc`. Хук: `.githooks/commit-msg` → `git config core.hooksPath .githooks`.
---

## Команды

- Dev stack: `docker compose -f docker-compose.local.yml up -d`.
- Все test/build/lint/validator jobs запускаются только через `/qa-gate` и его
  единый versioned outer launcher: task-owned names, CPU/RAM/PID limits,
  wall-clock timeout, isolated `testframework/reports/<task-run>/` с owner marker
  и unconditional exact-resource cleanup в `finally`.
- Пока target launcher/scaffold не реализован и независимо не проверен, runtime
  status равен `NOT_READY`. Прямые `docker run`/`compose up` snippets и legacy
  `hubtest`/Vitest/TS Playwright runners не являются допустимым fallback.

---

*Обновлено: 2026-08-08 — executable cutover waves, current TASK-202 archive/counts, Touchstone beta.2 pre-release*
