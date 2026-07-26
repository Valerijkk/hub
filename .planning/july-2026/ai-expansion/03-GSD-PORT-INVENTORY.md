# GSD Port Inventory — Hub (русский, лучше оригинала)

> **Источник фактов:** `C:\Users\Valerii\.claude\skills\gsd-*/SKILL.md` — **79** skills (filesystem 2026-07-22).
> Прочитано: YAML frontmatter (`name`/`description`) + первые ~40 строк каждого `SKILL.md`.
> Политика: **паттерны и контракты**, не vendor-copy полных upstream текстов.
> Цель порта: `.cursor/skills/gsd-ru/`, `prompt-engineering/`, усиление `.cursor/agents/`, маршрутизация в `.cursor/GSD-INTEGRATION.md`.
> Горизонт: **22.07.2026 → 15.08.2026** (см. `02-TECH-PLAN.md`).

## Статус GSD-RU (факт)

**Уже создан** локальный pack `.cursor/skills/gsd-ru/` (6 skills + README) — не нужно создавать skeleton заново:

| gsd-ru skill | Путь | Закрывает upstream-паттерн |
|--------------|------|----------------------------|
| `phase-plan` | `.cursor/skills/gsd-ru/phase-plan/` | `gsd-plan-phase`, `gsd-analyze-dependencies` |
| `wave-run` | `.cursor/skills/gsd-ru/wave-run/` | `gsd-execute-phase` (+ шаблон team-lead §5.2) |
| `goal-verify` | `.cursor/skills/gsd-ru/goal-verify/` | `gsd-verify-work` (goal-backward, не CI-gate) |
| `handoff` | `.cursor/skills/gsd-ru/handoff/` | `gsd-progress` / `gsd-pause-work` / `gsd-resume-work` + Hub `/handoff-sync` |
| `portable-kit` | `.cursor/skills/gsd-ru/portable-kit/` | Hub-native (нет 1:1 upstream skill) |
| `microservices-strangler` | `.cursor/skills/gsd-ru/microservices-strangler/` | Hub-native checklist поверх ADR-ARCH-MICROSERVICES |

PE-зеркало (project-agnostic): `prompt-engineering/gsd-port/README.md` — playbooks `gsd-lifecycle`, `wave-execution`, `handoff-state` + skills `phase-plan`, `wave-orchestrate`, `verify-goal`.

---

## 1. Методология

1. Полный список имён с диска (`Get-ChildItem …\gsd-*\SKILL.md`) → **79**.
2. Категория по lifecycle (не по алфавиту в итогах; в полной таблице — алфавит).
3. **in Hub?** — факт наличия локального аналога / практики:
   - `да · gsd-ru/…` — локальный RU skill
   - `да · slash/…` — Hub slash-skill
   - `да · practice` — есть практика (agents / playbooks / `.planning/seeds` / STATE), без 1:1 skill
   - `partial · route` — только маршрутизация в `GSD-INTEGRATION.md`
   - `нет` — только upstream
4. Приоритет порта паттерна (не «скопировать SKILL.md»):
   - **P0** — ежедневный цикл Hub; паттерн обязателен в RU/PE/agents
   - **P1** — усилить route + agents/PE до 15.08
   - **P2** — опционально после горизонта / later
   - **SKIP** — не портировать
5. **Port target:** куда класть паттерн (`gsd-ru` / PE / agents / GSD-INTEGRATION / —).

### Легенда port target

| Target | Смысл |
|--------|--------|
| `gsd-ru/… (done)` | Уже в `.cursor/skills/gsd-ru/` |
| `PE …` | `prompt-engineering/` (playbook/skill) |
| `agents …` | секция GSD-паттерны в `.cursor/agents/*.md` |
| `GSD-INTEGRATION` | только маршрут upstream → когда звать |
| `playbooks …` | `.cursor/playbooks/` |
| `—` | не портировать / later without local skill |

---

## 2. Полная таблица (79)

| # | skill | category | in Hub? | Pri | port target |
|---|-------|----------|---------|-----|-------------|
| 1 | `gsd-add-backlog` | backlog | partial · route | P1 | PE backlog notes / agents product |
| 2 | `gsd-add-phase` | roadmap | partial · route | P1 | GSD-INTEGRATION |
| 3 | `gsd-add-tests` | verify | partial · route (qa-auto) | P0 | agents qa-auto1/qa-auto2 |
| 4 | `gsd-add-todo` | backlog | partial · route | P1 | PE |
| 5 | `gsd-ai-integration-phase` | plan-ai | partial · route (arch) | P1 | PE corporate / agents arch |
| 6 | `gsd-analyze-dependencies` | plan | да · gsd-ru/phase-plan | P0 | gsd-ru/phase-plan (done) |
| 7 | `gsd-audit-fix` | verify | partial · route | P1 | agents security-auditor / qa-auto |
| 8 | `gsd-audit-milestone` | verify | partial · route (product-ver) | P1 | agents product-ver |
| 9 | `gsd-audit-uat` | verify | partial · route (qa-manual) | P0 | agents qa-manual |
| 10 | `gsd-autonomous` | execute | partial · route (team-lead) | SKIP | не портировать (полный авто-режим) |
| 11 | `gsd-check-todos` | backlog | partial · route | P1 | PE |
| 12 | `gsd-cleanup` | meta | нет | SKIP | — |
| 13 | `gsd-code-review` | verify | да · practice (code-reviewer) | P0 | agents code-reviewer |
| 14 | `gsd-code-review-fix` | verify | да · practice (code-reviewer) | P0 | agents code-reviewer |
| 15 | `gsd-complete-milestone` | roadmap | partial · route | P1 | GSD-INTEGRATION |
| 16 | `gsd-debug` | debug | да · practice (debugger) + slash/test-triage | P0 | agents debugger |
| 17 | `gsd-discuss-phase` | plan | да · practice (ba1/ba2) | P0 | agents ba1/ba2 |
| 18 | `gsd-do` | orient | partial · route | P1 | GSD-INTEGRATION |
| 19 | `gsd-docs-update` | docs | да · practice (docs-writer) + rule 55 | P0 | agents docs-writer |
| 20 | `gsd-eval-review` | verify-ai | partial · route | P2 | later (AI phases) |
| 21 | `gsd-execute-phase` | execute | да · gsd-ru/wave-run + PE wave-* | P0 | gsd-ru/wave-run (done) |
| 22 | `gsd-explore` | plan | partial · route (product) | P1 | agents product |
| 23 | `gsd-extract_learnings` | meta | нет | P2 | — |
| 24 | `gsd-fast` | execute | partial · route (be/fe) | P0 | agents be*/fe* / codex-micro |
| 25 | `gsd-forensics` | debug | partial · route | P2 | — |
| 26 | `gsd-from-gsd2` | meta | нет | SKIP | — |
| 27 | `gsd-graphify` | intel | нет | P2 | — |
| 28 | `gsd-health` | orient | нет | P2 | — |
| 29 | `gsd-help` | orient | partial · slash/gsd-index | P1 | gsd-index + gsd-ru/README |
| 30 | `gsd-import` | meta | нет | P2 | — |
| 31 | `gsd-inbox` | meta | нет | SKIP | — |
| 32 | `gsd-insert-phase` | roadmap | partial · route | P1 | GSD-INTEGRATION |
| 33 | `gsd-intel` | intel | partial · route (sa) | P1 | GSD-INTEGRATION |
| 34 | `gsd-join-discord` | meta | нет | SKIP | — |
| 35 | `gsd-list-phase-assumptions` | plan | да · practice (gray-zones/) | P1 | PE phase-plan |
| 36 | `gsd-list-workspaces` | meta | нет | P2 | portable-kit note only |
| 37 | `gsd-manager` | execute | нет | P2 | — |
| 38 | `gsd-map-codebase` | intel | partial · route (sa) | P1 | GSD-INTEGRATION |
| 39 | `gsd-milestone-summary` | docs | да · practice (milestones/) | P1 | agents docs-writer / PE |
| 40 | `gsd-new-milestone` | roadmap | partial · route (product) | P1 | agents product / portable-kit |
| 41 | `gsd-new-project` | roadmap | нет (Hub уже есть) | P1 | portable-kit |
| 42 | `gsd-new-workspace` | meta | нет | P2 | portable-kit note |
| 43 | `gsd-next` | orient | partial · route | P1 | GSD-INTEGRATION |
| 44 | `gsd-note` | backlog | нет | P1 | PE |
| 45 | `gsd-pause-work` | orient | да · gsd-ru/handoff + PE handoff-state | P1 | PE handoff-state (pattern); gsd-ru/handoff (done) |
| 46 | `gsd-plan-milestone-gaps` | roadmap | partial · route | P1 | GSD-INTEGRATION |
| 47 | `gsd-plan-phase` | plan | да · gsd-ru/phase-plan + PE phase-plan | P0 | gsd-ru/phase-plan (done) |
| 48 | `gsd-plant-seed` | backlog | да · practice (`.planning/seeds/`) | P0 | agents ba2/product |
| 49 | `gsd-pr-branch` | ship | да · practice (git-workflow) | P1 | playbooks/git-workflow |
| 50 | `gsd-profile-user` | orient | нет | P2 | — |
| 51 | `gsd-progress` | orient | да · practice (STATE.md) + gsd-ru/handoff | P0 | gsd-ru/handoff (done) |
| 52 | `gsd-quick` | execute | partial · route | P0 | agents be*/fe* |
| 53 | `gsd-reapply-patches` | meta | нет | SKIP | — |
| 54 | `gsd-remove-phase` | roadmap | partial · route | P1 | GSD-INTEGRATION |
| 55 | `gsd-remove-workspace` | meta | нет | P2 | — |
| 56 | `gsd-research-phase` | plan | partial · route | P1 | PE |
| 57 | `gsd-resume-work` | orient | да · slash/handoff-sync + gsd-ru/handoff | P0 | gsd-ru/handoff (done) |
| 58 | `gsd-review` | meta | нет | P2 | — |
| 59 | `gsd-review-backlog` | backlog | partial · route (product) | P1 | agents product / PE |
| 60 | `gsd-scan` | intel | partial · route | P1 | GSD-INTEGRATION |
| 61 | `gsd-secure-phase` | verify | да · practice (security-auditor) | P0 | agents security-auditor (+ gsd-ru/microservices-strangler для ADR) |
| 62 | `gsd-session-report` | orient | нет | P2 | optional PE |
| 63 | `gsd-set-profile` | orient | да · practice (ADR-MODELS-GPT56) | P1 | cutover checklist (`02-TECH-PLAN`) |
| 64 | `gsd-settings` | orient | да · practice (ADR-MODELS-GPT56) | P1 | cutover checklist |
| 65 | `gsd-ship` | ship | да · practice (devops) | P0 | agents devops-engineer |
| 66 | `gsd-sketch` | design | partial · route (designer) | P1 | agents designer |
| 67 | `gsd-sketch-wrap-up` | design | partial · route | P1 | agents designer |
| 68 | `gsd-spec-phase` | plan | да · practice (ba) | P0 | agents ba1/ba2 |
| 69 | `gsd-spike` | execute | нет | P1 | PE |
| 70 | `gsd-spike-wrap-up` | execute | нет | P1 | PE |
| 71 | `gsd-stats` | orient | нет | P2 | — |
| 72 | `gsd-thread` | orient | нет | P2 | — |
| 73 | `gsd-ui-phase` | design | partial · route (designer) | P1 | agents designer |
| 74 | `gsd-ui-review` | verify | partial · route (designer-ver/fe-ver) | P1 | agents designer-ver / fe-ver |
| 75 | `gsd-undo` | meta | нет | SKIP | — |
| 76 | `gsd-update` | meta | нет | SKIP | — |
| 77 | `gsd-validate-phase` | verify | partial · route (qa-auto; Nyquist gaps) | P0 | agents qa-auto* / PE |
| 78 | `gsd-verify-work` | verify | да · gsd-ru/goal-verify + PE verify-goal | P0 | gsd-ru/goal-verify (done) |
| 79 | `gsd-workstreams` | execute | да · practice (team-workflow) | P0 | PE wave-execution + agents team-lead |

### Descriptions (frontmatter, сжато)

| skill | description (из YAML) |
|-------|------------------------|
| `gsd-add-backlog` | Add an idea to the backlog parking lot (999.x numbering) |
| `gsd-add-phase` | Add phase to end of current milestone in roadmap |
| `gsd-add-tests` | Generate tests for a completed phase based on UAT criteria and implementation |
| `gsd-add-todo` | Capture idea or task as todo from current conversation context |
| `gsd-ai-integration-phase` | Generate AI design contract (AI-SPEC.md) for AI system phases |
| `gsd-analyze-dependencies` | Analyze phase dependencies and suggest Depends on entries for ROADMAP.md |
| `gsd-audit-fix` | Autonomous audit-to-fix pipeline — find, classify, fix, test, commit |
| `gsd-audit-milestone` | Audit milestone completion against original intent before archiving |
| `gsd-audit-uat` | Cross-phase audit of all outstanding UAT and verification items |
| `gsd-autonomous` | Run all remaining phases autonomously — discuss→plan→execute per phase |
| `gsd-check-todos` | List pending todos and select one to work on |
| `gsd-cleanup` | Archive accumulated phase directories from completed milestones |
| `gsd-code-review` | Review source files changed during a phase for bugs/security/quality |
| `gsd-code-review-fix` | Auto-fix issues found by code review in REVIEW.md |
| `gsd-complete-milestone` | Archive completed milestone and prepare for next version |
| `gsd-debug` | Systematic debugging with persistent state across context resets |
| `gsd-discuss-phase` | Gather phase context through adaptive questioning before planning |
| `gsd-do` | Route freeform text to the right GSD command automatically |
| `gsd-docs-update` | Generate or update project documentation verified against the codebase |
| `gsd-eval-review` | Retroactive audit of AI phase evaluation coverage |
| `gsd-execute-phase` | Execute all plans in a phase with wave-based parallelization |
| `gsd-explore` | Socratic ideation and idea routing |
| `gsd-extract_learnings` | Extract decisions, lessons, patterns from completed phase artifacts |
| `gsd-fast` | Execute a trivial task inline — no subagents, no planning overhead |
| `gsd-forensics` | Post-mortem for failed GSD workflows |
| `gsd-from-gsd2` | Import a GSD-2 (.gsd/) project back to GSD v1 (.planning/) |
| `gsd-graphify` | Build/query project knowledge graph in .planning/graphs/ |
| `gsd-health` | Diagnose planning directory health and optionally repair |
| `gsd-help` | Show available GSD commands and usage guide |
| `gsd-import` | Ingest external plans with conflict detection |
| `gsd-inbox` | Triage open GitHub issues/PRs against templates |
| `gsd-insert-phase` | Insert urgent work as decimal phase (e.g. 72.1) |
| `gsd-intel` | Query/inspect/refresh codebase intelligence in .planning/intel/ |
| `gsd-join-discord` | Join the GSD Discord community |
| `gsd-list-phase-assumptions` | Surface assumptions about a phase approach before planning |
| `gsd-list-workspaces` | List active GSD workspaces and status |
| `gsd-manager` | Interactive command center for managing multiple phases |
| `gsd-map-codebase` | Analyze codebase with parallel mapper agents → .planning/codebase/ |
| `gsd-milestone-summary` | Project summary from milestone artifacts |
| `gsd-new-milestone` | Start a new milestone cycle — update PROJECT.md |
| `gsd-new-project` | Initialize a new project with deep context gathering |
| `gsd-new-workspace` | Isolated workspace with repo copies and independent .planning/ |
| `gsd-next` | Automatically advance to the next logical GSD step |
| `gsd-note` | Zero-friction idea capture — append/list/promote notes |
| `gsd-pause-work` | Create context handoff when pausing mid-phase |
| `gsd-plan-milestone-gaps` | Create phases to close gaps from milestone audit |
| `gsd-plan-phase` | Create detailed phase plan (PLAN.md) with verification loop |
| `gsd-plant-seed` | Forward-looking idea with trigger conditions → `.planning/seeds/` |
| `gsd-pr-branch` | Clean PR branch filtering out .planning/ commits |
| `gsd-profile-user` | Generate developer behavioral profile |
| `gsd-progress` | Check progress, show context, route next action (`--forensic`) |
| `gsd-quick` | Quick task with GSD guarantees, skip optional agents |
| `gsd-reapply-patches` | Reapply local modifications after a GSD update |
| `gsd-remove-phase` | Remove a future phase and renumber |
| `gsd-remove-workspace` | Remove a GSD workspace and clean worktrees |
| `gsd-research-phase` | Research how to implement a phase (standalone) |
| `gsd-resume-work` | Resume work from previous session with full context |
| `gsd-review` | Cross-AI peer review of phase plans via external CLIs |
| `gsd-review-backlog` | Review and promote backlog items to active milestone |
| `gsd-scan` | Rapid codebase assessment (lite vs map-codebase) |
| `gsd-secure-phase` | Retroactively verify threat mitigations for a completed phase |
| `gsd-session-report` | Session report with token usage estimates |
| `gsd-set-profile` | Switch model profile (quality/balanced/budget/inherit) |
| `gsd-settings` | Configure GSD workflow toggles and model profile |
| `gsd-ship` | Create PR, run review, prepare merge after verification |
| `gsd-sketch` | Throwaway HTML UI mockups, multi-variant |
| `gsd-sketch-wrap-up` | Package sketch findings into a persistent project skill |
| `gsd-spec-phase` | Socratic WHAT-spec with ambiguity scoring → SPEC.md |
| `gsd-spike` | Throwaway experiments to validate feasibility |
| `gsd-spike-wrap-up` | Package spike findings into a persistent project skill |
| `gsd-stats` | Project statistics — phases, plans, requirements, git, timeline |
| `gsd-thread` | Manage persistent context threads across sessions |
| `gsd-ui-phase` | Generate UI design contract (UI-SPEC.md) |
| `gsd-ui-review` | Retroactive 6-pillar visual audit of frontend |
| `gsd-undo` | Safe git revert of phase/plan commits via phase manifest |
| `gsd-update` | Update GSD to latest version with changelog |
| `gsd-validate-phase` | Retroactively audit/fill Nyquist validation gaps |
| `gsd-verify-work` | Validate built features through conversational UAT |
| `gsd-workstreams` | Manage parallel workstreams (list/create/switch/…) |

---

## 3. Итоги приоритетов

| Pri | Count | Смысл |
|-----|------:|-------|
| **P0** | **21** | Ежедневный цикл: plan → wave → verify → handoff + quality/security |
| **P1** | **35** | Route + agents/PE до 15.08 |
| **P2** | **15** | Later / optional |
| **SKIP** | **8** | Не портировать |
| **Σ** | **79** | = filesystem count |

### P0 список (21)

`gsd-add-tests`, `gsd-analyze-dependencies`, `gsd-audit-uat`, `gsd-code-review`, `gsd-code-review-fix`, `gsd-debug`, `gsd-discuss-phase`, `gsd-docs-update`, `gsd-execute-phase`, `gsd-fast`, `gsd-plan-phase`, `gsd-plant-seed`, `gsd-progress`, `gsd-quick`, `gsd-resume-work`, `gsd-secure-phase`, `gsd-ship`, `gsd-spec-phase`, `gsd-validate-phase`, `gsd-verify-work`, `gsd-workstreams`.

Из них **паттерн уже в gsd-ru (done):**
`plan-phase` / `analyze-dependencies` → `phase-plan`; `execute-phase` → `wave-run`; `verify-work` → `goal-verify`; `progress` / `resume-work` (+ pause) → `handoff`.

### SKIP список (8)

`gsd-autonomous` (полный авто-режим), `gsd-cleanup`, `gsd-from-gsd2`, `gsd-inbox`, `gsd-join-discord`, `gsd-reapply-patches`, `gsd-undo`, `gsd-update`.

### Hub slash-skills (не из upstream GSD, но закрывают gaps)

| Hub skill | Связь с GSD |
|-----------|-------------|
| `gsd-index` | аналог `gsd-help` (индекс) |
| `handoff-sync` | handoff/STATE (рядом с `gsd-progress`/`pause`/`resume`) |
| `qa-gate` | CI/quality gate — **не** замена `gsd-verify-work` (UAT) |
| `test-triage` | ворота перед `gsd-debug` |
| `coverage-audit` | матрица требование×код×тест (в GSD-INTEGRATION иногда назван `gsd-coverage-audit`, **такого skill на диске ~/.claude нет**) |
| `style-check` / `anti-ai-check` | качество кода, не GSD-lifecycle |

### Уточнение: имена в GSD-INTEGRATION, которых нет как `gsd-*/SKILL.md`

На диске **нет** skills: `gsd-coverage-audit`, `gsd-select-framework`, `gsd-plan-checker`, `gsd-phase-researcher`, `gsd-domain-researcher`, `gsd-eval-planner` — это агенты/сабагенты или Hub-slash, не отдельные upstream skills. В таблицу 79 не входят.

---

## 4. Обязательные паттерны (портировать в первую очередь)

Из frontmatter/objective первых ~40 строк ключевых skills + Hub-практика:

1. **Goal-backward** — фаза принята, если goal верен (`gsd-verify-work` / `gsd-ru/goal-verify`), не «все таски done».
2. **Atomic plans** — один артефакт / один слой / ≤200–300 LOC; verification loop в `gsd-plan-phase`.
3. **Wave execution** — оркестратор лёгкий (~15%), subagent на план с полным контекстом (`gsd-execute-phase`); в Hub — только шаблон team-lead §5.2 (`gsd-ru/wave-run`).
4. **Depends / waves** — явные зависимости между планами (`gsd-analyze-dependencies`).
5. **Handoff + STATE** — pause/resume/progress без утечки чата (`gsd-pause-work` / `resume-work` / `progress` → `gsd-ru/handoff` + `/handoff-sync`).
6. **Discuss → Spec → Plan → Execute → Verify** lifecycle (`discuss-phase` → `spec-phase` → `plan-phase` → `execute-phase` → `verify-work`).
7. **Verification loops** — plan-checker / Nyquist gaps (`validate-phase`) + Hub rule 56 (новое + регрессия).
8. **Secure-phase** — threats из плана закрыты в коде → SECURITY.md (`gsd-secure-phase`).
9. **Workstreams** — параллельные линии с явными границами (`gsd-workstreams` ↔ team-workflow).
10. **Plant-seed** — идея с триггером → `.planning/seeds/` (`gsd-plant-seed`), не вечный Deferred-шум.
11. **Debug scientific method** — persistent session across resets (`gsd-debug`); в Hub сначала `/test-triage`.
12. **Quick vs plan** — `gsd-fast`/`gsd-quick` для мелкого; крупное → `plan-phase` (не раздувать quick).

---

## 5. НЕ портировать

| skill / класс | Почему (факт) |
|---------------|----------------|
| `gsd-update` | Самообновление upstream GSD CLI |
| `gsd-reapply-patches` | Патчи после update vendor tooling |
| `gsd-undo` | Revert через GSD phase manifest — meta CLI |
| `gsd-cleanup` | Архивация phase dirs upstream layout |
| `gsd-from-gsd2` | Миграция формата `.gsd/` → `.planning/` |
| `gsd-join-discord` | Внешнее community |
| `gsd-inbox` | GitHub inbox triage vendor |
| `gsd-autonomous` (полный режим) | Hub оркестрация = `team-lead` + §5.2, не solo auto-run всех фаз |
| Полные тексты upstream `SKILL.md` + `$HOME/.claude/get-shit-done/workflows/*` | Дрейф/vendor; берём структуру и gates, пишем RU/Hub |

---

## 6. Gap analysis (факт на 2026-07-22)

| Gap | Статус |
|-----|--------|
| Нет локального RU lite без `~/.claude` | **закрыт** — `.cursor/skills/gsd-ru/` (6 skills) |
| PE без GSD-lifecycle | **закрыт** — `prompt-engineering/gsd-port/` + 3 playbooks + 3 skills |
| Agents: GSD по имени, мало «как думать» | **в работе** (W1–W2 tech plan: секции GSD-паттерны) |
| Portable kit | **есть артефакты** — `05-PORTABLE-KIT.md`, `PORTABLE-HUB-KIT.md`, `gsd-ru/portable-kit` |
| Microservices только ADR | **закрыт checklist** — `gsd-ru/microservices-strangler` |
| `gsd-add-tests` не был в старом inventory | **добавлен** в полную таблицу (P0) |

---

## 7. План недель до 15.08.2026

Даты из `02-TECH-PLAN.md` / `ai-expansion/README.md`.

| Week | Даты | Port focus | Definition of done |
|------|------|------------|--------------------|
| **W1** | 22–28.07 | Inventory (этот файл); GSD-секции в agents; **gsd-ru уже создан** — не skeleton, а sync README/index | 03 готов; agents leadership/builders/analysis получили GSD-секции; `gsd-index` ссылается на gsd-ru |
| **W2** | 29.07–04.08 | PE corporate agents + playbooks/skills/templates (уже частично есть — довести consistency) | `gsd-port/README` ↔ agents PE; prompts/13; нет Hub-хардкода в PE agents |
| **W3** | 05–11.08 | Portable kit polish + reduced-GitHub blueprint + microservices checklist usage notes | `05`/`06`/`PORTABLE-HUB-KIT` согласованы; portable-kit skill usable |
| **W4** | 12–15.08 | GSD-INTEGRATION/P1 polish, eval pilot, приёмка; dynamic routing applied 26.07 | P0 patterns + structural PASS; representative eval evidence; product-ver + team-lead accept |

### Оставшийся P0-порт после gsd-ru (не skills, а agents)

| Upstream P0 | Куда добить до 15.08 |
|-------------|----------------------|
| `discuss-phase` / `spec-phase` | agents `ba1`/`ba2` |
| `plant-seed` | agents `ba2`/`product` + practice seeds/ |
| `workstreams` / `quick` / `fast` | agents `team-lead` / `be*` / `fe*` |
| `add-tests` / `validate-phase` / `audit-uat` | agents `qa-auto*` / `qa-manual` |
| `code-review` / `code-review-fix` | agents `code-reviewer` |
| `secure-phase` | agents `security-auditor` |
| `debug` | agents `debugger` (+ test-triage) |
| `docs-update` | agents `docs-writer` |
| `ship` | agents `devops-engineer` |

---

## 8. Сверка с filesystem

```
ON_DISK gsd-*/SKILL.md = 79
P0=21  P1=35  P2=15  SKIP=8  (21+35+15+8=79)
gsd-ru skills = 6 (phase-plan, wave-run, goal-verify, handoff, portable-kit, microservices-strangler)
Отличие от черновика inventory: добавлен gsd-add-tests; SKIP явно отделён от P2; gsd-ru отмечен как созданный.
```

*Факты сверены с filesystem и frontmatter 2026-07-22. Выдуманных skills нет.*
