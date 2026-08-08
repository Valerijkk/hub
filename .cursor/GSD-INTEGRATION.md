# Hub ↔ GSD Skills Integration

> GSD skills: `~/.agents/skills/gsd-*/SKILL.md` · typed agents: `~/.codex/agents/gsd-*.toml` · Hub slash-skills: `.cursor/skills/*/SKILL.md`.
> Ростер команды: `.cursor/playbooks/README.md` · Регламент: `.cursor/REGULATION.md`.
>
> **Этот файл отвечает «какая ситуация → какой скил» (маршрутизация).** Вопрос «где в промпте объявить скил и что агент с ним делает» — [`playbooks/prompt-contract.md`](playbooks/prompt-contract.md) §13 (в т.ч. `/task-packet`, `/context-selector`, `/evidence-verifier`, `/workflow-eval` и обязательные скилы-ворота `/test-triage`, `/qa-gate`, `/handoff-sync`, `/style-check`, `/coverage-audit`).

*Обновлено: 2026-08-08 — live install: 71 GSD skills + 34 typed agents; Superpowers installed as a workflow source, not a role/authority · historical port inventory remains dated 2026-07-22.*

---

## Installed GSD vs Hub-native (source of truth)

| Слой | Где лежит | Имя skill | Примеры |
|------|-----------|-----------|---------|
| **Installed GSD skills** | `~/.agents/skills/gsd-<name>/` | = имя папки | `gsd-plan-phase`, `gsd-execute-phase`, `gsd-ai-integration-phase` (**71 live skills**) |
| **Typed GSD agents** | `~/.codex/agents/gsd-<role>.toml` | = `agent_type` | `gsd-planner`, `gsd-plan-checker`, `gsd-executor`, `gsd-security-auditor` (**34 live profiles**) |
| **Hub slash** | `.cursor/skills/<name>/SKILL.md` | = имя папки | `task-packet`, `context-selector`, `evidence-verifier`, `workflow-eval`, `qa-gate`, `handoff-sync`, `coverage-audit`, `test-triage` |
| **GSD-RU lite** | `.cursor/skills/gsd-ru/<name>/SKILL.md` | = имя папки | `phase-plan`, `wave-run`, `goal-verify`, `handoff` |
| **Superpowers** | Codex plugin runtime | plugin-contributed workflow practices | Дополняет discovery/spec/planning/review; не заменяет GSD и не создаёт agent slot |

Этот файл — каноническая Hub-матрица маршрутизации. Полный каталог installed
skills показывает `gsd-help`, а неясный intent принимает `gsd-next` или `gsd-progress`.
Маршрутизируй только существующий installed skill или Hub-путь из таблицы выше.
Coverage — Hub `/coverage-audit`; role/wave handoff — GSD-RU skill
`.cursor/skills/gsd-ru/handoff/SKILL.md`; state/docs sync — `/handoff-sync`;
framework selection входит в `gsd-ai-integration-phase`.

Live-инвентарь 2026-08-08: **71 skills + 34 typed agents**. Файл
`ai-expansion/03-GSD-PORT-INVENTORY.md` фиксирует исторический срез **79 skills**
от 2026-07-22 и не является текущим install count.

---

## GSD-RU (Hub, русский lite)

Локальный pack без обязательного `~/.claude` на каждый шаг: [`.cursor/skills/gsd-ru/README.md`](skills/gsd-ru/README.md)

| Intent | GSD-RU invocation |
|--------|-------------------|
| План фазы | `/phase-plan` |
| Волна агентов | `/wave-run` |
| Goal-backward | `/goal-verify` |
| Role/wave handoff | GSD-RU `handoff` (`.cursor/skills/gsd-ru/handoff/SKILL.md`) |
| Перенос AI kit | `/portable-kit` |
| Strangler checklist | `/microservices-strangler` |

Планы порта и portable kit (горизонт → 15.08.2026): [`.planning/archive/july-2026/ai-expansion/`](../.planning/archive/july-2026/ai-expansion/README.md)
Слой разработки продукта живёт в `.cursor/` + `.planning/`; отдельного пакета под него нет — он и есть эти папки.

Инвентарь upstream skills: `ai-expansion/03-GSD-PORT-INVENTORY.md`.

---

## Единый контур Hub + GSD + Superpowers

Эти слои не конкурируют за authority:

1. **Hub policy/controls** задают разрешения, founder locks, ownership и
   обязательные project gates.
2. **GSD** ведёт жизненный цикл substantial work и persistent artifacts:
   discuss/spec → plan → execute → verify → ship.
3. **Superpowers 6.2.0** используется как versioned pattern source и отдельная
   eval-arm: task/diff-oriented review, bounded remediation, read-only reviewer,
   scoped state и context isolation. Его текст не выдаёт permissions и не
   доказывает superiority.
4. **Hub prompt skills** нормализуют конкретную границу работы:
   `/task-packet` → `/context-selector` → GSD/direct execution →
   `/evidence-verifier`; `/workflow-eval` измеряет, полезна ли добавленная
   ceremony на реальных Hub-задачах.

Приоритет конфликтов: founder locks → Hub safety/authority → GSD lifecycle →
совместимые Superpowers practices → локальные эвристики. Trivial reversible
правка не обязана проходить полный phase lifecycle: используется smallest
workflow, который способен дать требуемое evidence.

| Risk/shape | Маршрут |
|---|---|
| LOW: локально, обратимо, oracle очевиден | direct / `gsd-fast`; короткий task packet при делегировании |
| MEDIUM: несколько зависимых изменений | `/task-packet` → `gsd-quick` или короткий plan → focused verification |
| HIGH: cross-module/неясный design | `/task-packet` → `gsd-discuss-phase`/`gsd-plan-phase` → bounded waves → `/evidence-verifier` |
| CRITICAL: security/migration/prod/необратимость | design approval + explicit plan/threat model + executable oracle + independent verifier |

---

## P0/P1 port status mid-Aug

> Источник: [`03-GSD-PORT-INVENTORY.md`](../.planning/archive/july-2026/ai-expansion/03-GSD-PORT-INVENTORY.md) (79 skills · 2026-07-22).
> Σ: **P0=21 · P1=35 · P2=15 · SKIP=8**.
> Политика: паттерны в RU/PE/agents — не vendor-copy upstream `SKILL.md`.

### Сводка

| Pri | Count | Где закрываем |
|-----|------:|---------------|
| **P0** | **21** | `gsd-ru` (ядро цикла) + agents/PE (качество, discuss, ship) |
| **P1** | **35** | в основном **route-only** здесь + усиление agents/PE до 15.08 |
| **P2** | **15** | later |
| **SKIP** | **8** | не портировать (`autonomous`, `cleanup`, `from-gsd2`, `inbox`, `join-discord`, `reapply-patches`, `undo`, `update`) |

### P0 = 21 — полный список

`gsd-add-tests`, `gsd-audit-uat`, `gsd-capture`, `gsd-code-review`, `gsd-debug`, `gsd-discuss-phase`, `gsd-docs-update`, `gsd-execute-phase`, `gsd-fast`, `gsd-graphify`, `gsd-phase`, `gsd-plan-phase`, `gsd-progress`, `gsd-quick`, `gsd-resume-work`, `gsd-secure-phase`, `gsd-ship`, `gsd-spec-phase`, `gsd-validate-phase`, `gsd-verify-work`, `gsd-workstreams`.

### Уже в gsd-ru (done) — 6 upstream P0 → 4 skills

| Upstream P0 | GSD-RU invocation |
|-------------|-------------------|
| `gsd-plan-phase`, `gsd-graphify` | `/phase-plan` |
| `gsd-execute-phase` | `/wave-run` |
| `gsd-verify-work` | `/goal-verify` |
| `gsd-progress`, `gsd-resume-work` (+ pause как паттерн) | GSD-RU `handoff` |

Hub-native (не 1:1 из P0-списка, но в pack): `portable-kit`, `microservices-strangler`.

### P0 ещё не как gsd-ru skill — agents / practice / route

| Upstream P0 | Статус mid-Aug | Куда добить до 15.08 |
|-------------|----------------|----------------------|
| `gsd-discuss-phase`, `gsd-spec-phase` | practice · agents `ba1`/`ba2` | секции GSD + PE |
| `gsd-capture` | installed · routes seeds/tasks/notes to the right destination | agents `ba2`/`product` |
| `gsd-workstreams` | practice · team-workflow | PE `wave-execution` + `team-lead` |
| `gsd-quick`, `gsd-fast` | partial · route | agents `be*`/`fe*` / `codex-micro` |
| `gsd-add-tests`, `gsd-validate-phase` | partial · route | agents `qa-auto*` |
| `gsd-audit-uat` | partial · route | agents `qa-manual` |
| `gsd-code-review`, `gsd-code-review --fix` | installed review + fix mode | agents `code-reviewer` |
| `gsd-secure-phase` | practice · `security-auditor` | agents + при ADR — strangler |
| `gsd-debug` | practice · `debugger` + `/test-triage` | agents `debugger` |
| `gsd-docs-update` | practice · `docs-writer` + rule 55 | agents `docs-writer` |
| `gsd-ship` | practice · devops | agents `devops-engineer` |

**Итого P0:** 6/21 закрыты локальным `gsd-ru` skill; **15/21** — паттерн в agents/practice или route (не отдельный RU skill). Не смешивать с Hub slash: `/qa-gate` ≠ `gsd-verify-work`.

### P1 — route-only и PE (ориентир до 15.08)

Большинство из **35** P1 остаются **partial · route** в этой матрице (таблицы ниже) + точечное усиление:

| Кластер P1 | Примеры | Port target |
|------------|---------|-------------|
| Roadmap / orient | `add-phase`, `insert-phase`, `remove-phase`, `complete-milestone`, `do`, `next`, `plan-milestone-gaps` | GSD-INTEGRATION (route) |
| Intel / map | `intel`, `map-codebase`, `scan` | GSD-INTEGRATION · sa |
| Backlog / notes | `add-backlog`, `add-todo`, `check-todos`, `note`, `review-backlog` | PE / product |
| Design / UI | `sketch`, `sketch-wrap-up`, `ui-phase`, `ui-review` | agents designer* |
| Research / AI | `research-phase`, `ai-integration-phase`, `explore`, `list-phase-assumptions` | PE / arch / product |
| Ship / docs extras | `pr-branch`, `milestone-summary`, `help` | playbooks / docs-writer / `gsd-help` |
| Models | `set-profile`, `settings` | ADR-MODELS cutover checklist (W4) |

Полная построчная таблица 79 — только в `03-GSD-PORT-INVENTORY.md` (здесь не дублируем).

---

## Orchestrator (team-lead / пользователь)

| Intent | GSD Skill | Hub Slash |
|--------|-----------|-----------|
| Не знаю какую команду | `gsd-progress` | `gsd-next` для выбора следующего lifecycle-шага |
| Статус проекта | `gsd-progress` | — |
| Следующий шаг | `gsd-next` | — |
| Волны/исполнение | `gsd-execute-phase`, `gsd-workstreams` | — |
| Мелкая задача | `gsd-quick`, `gsd-fast` | — |
| Role/wave handoff; синхронизация STATE | GSD-RU `handoff` | `/handoff-sync` |
| Полный quality gate | — | `/qa-gate` |
| Матрица покрытия | — | `/coverage-audit` (Hub) |
| Нормализовать задачу и authority boundary | lifecycle выбирается после risk | `/task-packet` |
| Выбрать минимальный релевантный контекст | route к domain docs/plan | `/context-selector` |
| Независимо проверить claim/evidence | `gsd-verify-work` для phase/UAT | `/evidence-verifier` для artifact/diff claim |
| Сравнить plain/GSD/Superpowers workflow | — | `/workflow-eval` |

---

## Ростер 32 агента → GSD + playbook

| Agent | Primary GSD | Secondary GSD | Playbook |
|-------|-------------|---------------|----------|
| `team-lead` | `gsd-progress`, `gsd-execute-phase`, `gsd-autonomous` | `gsd-workstreams`, `gsd-next` | `team-workflow.md` |
| `tech-lead` | `gsd-plan-phase`, `gsd-graphify` | `gsd-phase`, `gsd-execute-phase` | `team-workflow.md` |
| `product` | `gsd-new-milestone`, `gsd-review-backlog` | `gsd-explore`, `gsd-capture` | `analysis.md`, `team-workflow.md` |
| `product-ver` | `gsd-verify-work` | `gsd-audit-milestone` | `verification.md` |
| `ba1`, `ba2` | `gsd-discuss-phase`, `gsd-spec-phase` | `gsd-review-backlog` | `analysis.md` |
| `ba-ver` | `gsd-verify-work` | — | `verification.md` |
| `sa1`, `sa2` | Hub `coverage-audit`, `gsd-ns-context` | `gsd-audit-uat`, `gsd-map-codebase` | `analysis.md` |
| `sa-ver` | `gsd-verify-work` | `gsd-audit-uat` | `verification.md` |
| `arch` | `gsd-ai-integration-phase`, `gsd-ui-phase` | `gsd-plan-phase --research-phase`, `gsd-eval-review` | `analysis.md`, `backend.md`, `frontend.md` |
| `arch-ver` | `gsd-verify-work` | `gsd-secure-phase` | `verification.md` |
| `designer` | `gsd-sketch`, `gsd-ui-phase` | `gsd-ui-review` | `design.md` |
| `designer-ver` | `gsd-ui-review` | `gsd-verify-work` | `verification.md`, `design.md` |
| `be1`, `be2` | `gsd-quick`, `gsd-debug` | `gsd-fast`, `gsd-add-tests` | `backend.md`, `observability.md` |
| `be-ver` | `gsd-verify-work`, `gsd-code-review` | `gsd-secure-phase` | `verification.md`, `backend.md` |
| `fe1`, `fe2` | `gsd-quick`, `gsd-ui-review` | `gsd-sketch`, `gsd-add-tests` | `frontend.md`, `design.md` |
| `fe-ver` | `gsd-verify-work`, `gsd-ui-review` | — | `verification.md`, `frontend.md` |
| `qa-manual` | `gsd-audit-uat`, `gsd-verify-work` | — | `testing.md` |
| `qa-manual-ver` | `gsd-verify-work`, `gsd-audit-uat` | — | `verification.md` |
| `qa-auto1`, `qa-auto2` | `gsd-add-tests`, `gsd-validate-phase` | `gsd-audit-fix` | `testing.md` |
| `qa-auto-ver` | `gsd-verify-work`, `gsd-validate-phase` | — | `verification.md`, `testing.md` |
| `devops-engineer` | `gsd-ship` | — | `observability.md`, `git-workflow.md` |
| `security-auditor` | `gsd-secure-phase`, `gsd-audit-fix` | — | `verification.md`, `backend.md` |
| `debugger` | `gsd-debug` | — | `observability.md`, `testing.md` |
| `code-reviewer` | `gsd-code-review`, `gsd-code-review --fix` | — | `verification.md`, `code-comments.md`, `git-workflow.md` |
| `docs-writer` | `gsd-docs-update`, `gsd-milestone-summary` | — | `code-comments.md` |

---

## Планы Hub в GSD-терминах

| Hub Artifact | GSD Equivalent |
|--------------|----------------|
| `.planning/ROADMAP.md` | Milestone phases |
| `.planning/STATE.md` | `gsd-progress` state |
| `.planning/COVERAGE.md` | Hub `coverage-audit` output |
| `.planning/EXPANSION-ANALYSIS-*.md` | backlog для `gsd-plan-phase` |
| `.planning/archive/quick/` | `gsd-quick` tasks |
| `.planning/debug/` | `gsd-debug` sessions |

---

## Правило: Test Triage (обязательно)

Перед фиксом упавшего теста — `.cursor/skills/test-triage/SKILL.md`:
1. Воспроизвести функционал (не тест). 2. Сверить с `docs/`. 3. Код верен → чинить тест; код неверен → чинить код. 4. Запрет «зелёных» без шагов 1–2.

---

## Wave execution (gsd-execute-phase)

```
Волна N: team-lead запускает ≤4 параллельных агента (разные модули/файлы)
  → каждый: реализация → регрессия → test-triage на FAIL → персональный *-ver
  → кросс-ревью парного агента (be1↔be2 / fe1↔fe2 / qa-auto1↔qa-auto2)
  → team-lead: сборка + /handoff-sync
Волна N+1: зависит от артефактов Волны N
```

Детали и RACI: `.cursor/playbooks/team-workflow.md`

---

## Когда какую GSD-команду — расширенная матрица по фазам жизненного цикла

> Это каноническая подробная Hub-маршрутизация с привязкой к реальным
> ситуациям. Полный список installed-команд показывает `gsd-help`; здесь —
> только команды, используемые в Hub-практике.

### 1. Старт / ориентация в сессии

| Ситуация | Команда | Комментарий |
|----------|---------|-------------|
| Новая сессия, не знаю с чего начать | `gsd-progress` | Роутит freeform-текст в нужную команду |
| «Где мы сейчас в проекте?» | `gsd-progress` | Сверяет `.planning/STATE.md` + `ROADMAP.md`, есть флаг `--forensic` для integrity-аудита |
| «Что делать дальше?» | `gsd-next` | Логический шаг после завершения фазы/плана |
| Хочу понять открытые задачи | `gsd-progress` | Перед стартом новой волны; новые идеи/notes фиксирует `gsd-capture` |
| Резюме предыдущей сессии | `gsd-resume-work` | После `gsd-pause-work` в прошлой сессии |

### 2. Планирование фичи/фазы

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Крупная новая фича, нужна декомпозиция | `gsd-discuss-phase` → `gsd-plan-phase` | CAL-SYNC (email+calendar+calls sync) |
| Нужно понять зависимости фаз | `gsd-graphify` | Перед вставкой новой фазы в roadmap |
| Срочная задача между фазами | `gsd-phase` insert (напр. `72.1`) | Хотфикс безопасности вне очереди |
| Фича на конец roadmap | `gsd-phase` add | Добавление v4 AI hub фазы |
| Нужен research перед планом | `gsd-plan-phase --research-phase N` | Выбор AI framework для gateway BYOK |
| AI-система (провайдер/агент/RAG) | `gsd-ai-integration-phase` | AI Provider Hub Spec (`.planning/decisions/specs/AI-PROVIDER-HUB-SPEC.md`) |
| UI-контракт перед реализацией | `gsd-ui-phase` → `gsd-ui-review` | Bases inline-edit экран |
| Независимое peer review плана | `gsd-review` | team-lead перед запуском рискованной волны |

### 3. Исполнение

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Параллельное исполнение планов волны | `gsd-execute-phase` | Волна be1/be2/fe1/fe2 на 4 независимых модуля |
| Несколько долгоживущих параллельных потоков | `gsd-workstreams` | Track F (frontend) и Track P (platform) одновременно |
| Мелкая точечная задача с трекингом | `gsd-quick` | Добавить один RBAC-guard в `bases/service.go` |
| Тривиальный inline-фикс без оверхеда | `gsd-fast` | Опечатка в error message, переименование поля DTO |
| Всё подряд автономно | `gsd-autonomous` | Только `team-lead`, финальные волны |

### 4. Качество / тесты / баги

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Нужны тесты после фичи | `gsd-add-tests` | После WIP-лимитов в tracker — `TestBoards_WipLimit*` |
| Nyquist-gaps после фазы | `gsd-validate-phase` | Retroactive проверка фазы 121 |
| UAT / acceptance проверка | `gsd-verify-work` | Приёмка Wave C аудита |
| Сложный баг, нужна persistent-сессия | `gsd-debug` | Flaky `TestSessionMiddleware_ExpiredJWT`, race в `ws/hub.go` |
| Массовые findings → фикс | `gsd-audit-fix` | После `AUDIT-2026-06-23-full.md` (сначала `--dry-run`) |
| Cross-phase UAT-долги | `gsd-audit-uat` | Ревизия открытых UAT перед релизом |
| Root-cause post-mortem неудачного workflow | `gsd-forensics` | Волна упала на интеграции — разбор что пошло не так |
| **Hub:** упавший тест — код vs тест | `.cursor/skills/test-triage/SKILL.md` | См. `REGULATION.md` §14, §18 |
| **Hub:** полный gate | `.cursor/skills/qa-gate/SKILL.md` | См. `REGULATION.md` §15 |

### 5. Документация / аудит покрытия

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Bulk-обновление docs с верификацией по коду | `gsd-docs-update` | После `.planning/` reorg — проверка ссылок в `docs/README.md` |
| Матрица требование×код×тест | Hub `coverage-audit` | Обновление `.planning/COVERAGE.md` |
| Codebase intelligence refresh | `gsd-ns-context` или `gsd-map-codebase` | После крупного рефакторинга tracker automation |
| Глубокий анализ кодовой базы | `gsd-map-codebase` | Перед v4 super-app декомпозицией |
| Milestone-резюме для онбординга | `gsd-milestone-summary` | `.planning/milestones/v4-super-app/SUMMARY.md` |
| Peer review плана/кода | `gsd-code-review` → `gsd-code-review --fix` | Pre-merge diff review |
| Threat-mitigation verification | `gsd-secure-phase` | После изменений auth/RBAC |
| Аудит eval-покрытия AI-фазы | `gsd-eval-review` | AI gateway BYOK evaluation strategy |

### 6. Завершение / handoff

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Конец значимой сессии | `/handoff-sync` (Hub) + `gsd-progress` | Всегда перед закрытием сессии с кодом |
| Пауза посреди фазы | `gsd-pause-work` | Context handoff на следующую сессию |
| Завершение milestone | `gsd-complete-milestone`, `gsd-audit-milestone` | v4 super-app → архивация |
| PR готов к ревью | `gsd-ship` | После green `/qa-gate` |
| Разбить на несколько PR | `gsd-pr-branch` | Волна с 4 модулями → 4 отдельных PR |

---

## Ростер 32 агента → GSD + playbook (детализация по use-case)

> Таблица выше (§«Ростер 32 агента») даёт primary/secondary skill. Ниже — конкретные ситуации, когда агент реально вызывает каждый skill.

| Agent | Skill | Реальная ситуация Hub |
|-------|-------|------------------------|
| `team-lead` | `gsd-execute-phase` | Запуск волны 4 параллельных be/fe агентов на разные модули (auth+tracker+bases+tms) |
| `team-lead` | `gsd-workstreams` | Одновременное ведение Track F (frontend redesign) и Track P (platform/otel) |
| `tech-lead` | `gsd-plan-phase` | Декомпозиция CAL-SYNC на фазы email L0-L3 + calendar + calls RSVP |
| `ba1`/`ba2` | `gsd-discuss-phase` | Уточнение серых зон перед US для intranet v2 (poll-in-post, kudos) |
| `sa1`/`sa2` | Hub `coverage-audit` | Трассировка FR-CAL-01…08 → `calendar_event` ERD → тесты |
| `arch` | `gsd-ai-integration-phase` | Framework selection для AI gateway (BYOK vs Ollama) — внутри skill |
| `designer` | `gsd-sketch` | Throwaway HTML-мокапы bases inline-edit перед `gsd-ui-phase` |
| `be1`/`be2` | `gsd-debug` | Race condition в `ws/hub_shutdown_test.go` — persistent debug session |
| `fe1`/`fe2` | `gsd-ui-review` | Retroactive 6-pillar аудит v4 bases-page.tsx после реализации |
| `qa-manual` | `gsd-audit-uat` | Свежая ревизия TC для tms после wave96 testrail import |
| `qa-auto1`/`qa-auto2` | `gsd-add-tests` | Новые тесты `service_wave12x_test.go` после каждой волны tracker |
| `devops-engineer` | `gsd-ship` | PR для CI job `coverage-gate` после изменения threshold |
| `security-auditor` | `gsd-secure-phase` | Retroactive проверка threat model после RBAC-фикса в bases |
| `debugger` | `gsd-debug` | Root-cause паники в `automation_consumer.go` под нагрузкой |
| `code-reviewer` | `gsd-code-review` | Pre-merge review волны из 4 be/fe веток перед интеграцией |
| `docs-writer` | `gsd-docs-update` | Синхронизация `docs/system_specification.md` после Волны 2 (Intranet/Calendar/Wiki спеки) |

---

## Планы Hub в GSD-терминах (актуализировано после reorg `.planning/` 2026-07-18)

| Hub Artifact | GSD Equivalent | Путь |
|--------------|-----------------|------|
| `.planning/ROADMAP.md` | Milestone phases (исторический, 125 фаз) | корень |
| `.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md` | Актуальный roadmap (волны Track F/P, TF-PY, DOCS-EXCELLENCE) | `july-2026/` |
| `.planning/STATE.md` | `gsd-progress` state | корень |
| `.planning/COVERAGE.md` | Hub `coverage-audit` output | корень |
| `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md` | backlog для `gsd-plan-phase` | `july-2026/` |
| `.planning/archive/quick/` | `gsd-quick` tasks | корень (если создаётся) |
| `.planning/debug/` | `gsd-debug` sessions | корень |
| `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` | Зафиксированные gray-area решения (аналог assumptions-mode `gsd-discuss-phase`) | `gray-zones/` |
| `.planning/audits/2026-06/AUDIT-SCORECARD.md` | `gsd-audit-uat` / `gsd-validate-phase` output | `audits/2026-06/` |
| `.planning/decisions/ADR-*.md` | ADR (аналог decisions log GSD) | `decisions/` |
| `.planning/reference/seeds/SEED-*.md` | `gsd-capture` seed output | `seeds/` |
| `.planning/milestones/v4-super-app/SUMMARY.md` | `gsd-milestone-summary` output | `milestones/` |
| `.planning/ops/production/PRODUCTION-HANDOFF.md` | `gsd-ship` / prod handoff | `production/` |

---

## Пример: полный GSD-маршрут для новой крупной фичи (CAL-SYNC)

```
1. gsd-discuss-phase --power   → ba1/ba2 собирают серые зоны email L0-L3, calendar RSVP
2. gsd-plan-phase --research-phase N → arch исследует iCal RFC 5545 constraints
3. gsd-plan-phase               → tech-lead: PLAN.md с task breakdown + Depends on
4. gsd-review                    → независимое peer review плана до исполнения
5. gsd-execute-phase             → be1 (calendar_event ERD) + be2 (ical sync) + fe1 (UI) параллельно
   → каждый: /test-triage на FAIL → персональный *-ver
6. gsd-add-tests                 → qa-auto1/2 добивают coverage gaps
7. gsd-secure-phase               → security-auditor: ValidateOutboundURL для ical_subscription (SSRF)
8. gsd-code-review                → code-reviewer pre-merge
9. gsd-docs-update                → docs-writer: system_specification.md + wiki/API.md
10. /qa-gate (Hub)                → team-lead финальный чек
11. gsd-milestone-summary          → архив после релиза
```

---

## Пример: полный GSD-маршрут для мелкого бага (auth rate-limit)

```
1. gsd-progress "падает тест rate limit в auth" → роутит в gsd-debug
2. gsd-debug                                → воспроизвести, найти root cause в service_login.go
3. .cursor/skills/test-triage/SKILL.md      → классификация ROOT=code (см. REGULATION.md §14, A1)
4. gsd-fast (или gsd-quick, если >1 файл)   → минимальный фикс
5. /qa-gate --domain auth (точечный прогон) → testframework/auth/...
6. gsd-docs-update (если менялся контракт)  → security_guidelines.md
7. be-ver                                    → verified: true
```

---

## Hub slash-skills vs GSD skills — когда что

> Частая ошибка маршрутизации: пытаться использовать GSD `gsd-verify-work` там, где нужен доменный Hub-гейт, или наоборот.

| Нужно | Используй | НЕ используй |
|-------|-----------|---------------|
| Полный build+test+lint+security гейт перед PR | `.cursor/skills/qa-gate/SKILL.md` | `gsd-verify-work` (это UAT, не CI-гейт) |
| Упавший тест — код или тест виноват | `.cursor/skills/test-triage/SKILL.md` | `gsd-debug` напрямую (сначала триаж, `gsd-debug` — только если триаж показал сложный root cause) |
| STATE/COVERAGE/ROADMAP синхронизация в конце сессии | `.cursor/skills/handoff-sync/SKILL.md` | `gsd-milestone-summary` (это для завершённого milestone, не рутинный handoff) |
| Требование×код×тест матрица | `.cursor/skills/coverage-audit/SKILL.md` | `gsd-map-codebase` (это структурный анализ, не coverage) |
| Anti-AI паттерны + стиль | `.cursor/skills/style-check/SKILL.md` + `references/anti-ai.md` | `gsd-code-review` (более общий peer review, не построчный grep) |
| Новая крупная фаза с research+planning | `gsd-discuss-phase` → `gsd-plan-phase` | Hub-skills (их нет для планирования — это зона GSD) |
| Retroactive verify реализованной фазы | `gsd-verify-work`, `gsd-validate-phase` | Hub-skills (Hub-гейты — pre-merge, GSD verify — post-hoc) |
| Сформировать bounded task contract перед делегированием | `.cursor/skills/task-packet/SKILL.md` | GSD plan как замену authority/acceptance |
| Выбрать smallest sufficient context | `.cursor/skills/context-selector/SKILL.md` | загрузку всех `docs/`/истории фаз |
| Проверить конкретный diff/артефакт без generator anchoring | `.cursor/skills/evidence-verifier/SKILL.md` | self-review или второй идентичный prompt |
| Доказать пользу нового workflow/plugin | `.cursor/skills/workflow-eval/SKILL.md` | одну demo-задачу или upstream benchmark |

## Частые ошибки маршрутизации (наблюдались в практике)

1. **Запуск `gsd-audit-fix` без `--dry-run`** на первом проходе — приводит к неконтролируемым правкам сразу в 10+ файлах. Правило: всегда dry-run → ревью списка → apply.
2. **`gsd-quick` для многофайловой фичи** — превышает scope skill, нужен `gsd-plan-phase`. Признак: если задача трогает >3 модуля или требует ADR — это `plan-phase`, не `quick`.
3. **Пропуск `test-triage` перед `gsd-debug`** — `gsd-debug` дороже (persistent session), для простых кейсов (см. REGULATION §14 таблицы A1-A4/T1-T4/B1-B4) достаточно триажа за 2 минуты.
4. **`/handoff-sync` вместо `gsd-docs-update`** для bulk-верификации всех doc против кода — `handoff-sync` обновляет память сессии, не верифицирует факты по коду построчно.
5. **Забытый `gsd-secure-phase`** после auth/RBAC/webhook изменений — security-инварианты TZ (webhook signature, WS RBAC, search user-scope, WebPush SSRF) требуют явной retroactive-проверки, не покрываются обычным `be-ver`.

## Каталог GSD skills, релевантных Hub (полный, по категориям)

> `~/.agents/skills/gsd-*/SKILL.md` — **71 live installed skill**; ниже — те, что реально применимы в контексте Hub (монорепо Go+React с 32 файлами Hub-ролей). Typed GSD agents считаются отдельно: **34** профиля в `~/.codex/agents/`. Остальные skills (например, `gsd-new-project`) не нужны для уже существующего проекта без явного lifecycle-запроса.

| Категория | Skills |
|-----------|--------|
| Оркестрация сессии | `gsd-manager`, `gsd-progress`, `gsd-next`, `gsd-help`, `gsd-thread`, `gsd-resume-work`, `gsd-pause-work` |
| Планирование | `gsd-discuss-phase`, `gsd-spec-phase`, `gsd-plan-phase`, `gsd-review`, `gsd-plan-review-convergence`, `gsd-phase`, `gsd-graphify` |
| AI-фичи | `gsd-ai-integration-phase` (вкл. framework selection), `gsd-eval-review`, `gsd-plan-phase --research-phase N` |
| UI/дизайн | `gsd-ui-phase`, `gsd-ui-review`, `gsd-sketch` |
| Исполнение | `gsd-execute-phase`, `gsd-workstreams`, `gsd-quick`, `gsd-fast`, `gsd-autonomous` |
| Качество/тесты | `gsd-add-tests`, `gsd-validate-phase`, `gsd-verify-work`, `gsd-audit-fix`, `gsd-audit-uat`, `gsd-debug`, `gsd-forensics` |
| Документация/анализ | `gsd-docs-update`, Hub `coverage-audit`, `gsd-ns-context`, `gsd-map-codebase`, `gsd-code-review`, `gsd-code-review --fix`, `gsd-milestone-summary`, `gsd-extract-learnings` |
| Безопасность | `gsd-secure-phase` |
| Завершение | `gsd-complete-milestone`, `gsd-audit-milestone`, `gsd-ship`, `gsd-pr-branch`, `gsd-undo` |
| Идеи/backlog | `gsd-explore`, `gsd-capture`, `gsd-review-backlog`, `gsd-inbox` |
| Служебное | `gsd-health`, `gsd-config`, `gsd-settings`, `gsd-surface`, `gsd-stats`, `gsd-graphify` |

---

## Типовые многошаговые сценарии Hub (skill-цепочки)

### Сценарий: security finding из аудита → закрытие

```
1. security-auditor читает .planning/audits/2026-06/FA-02-SECURITY-REAUDIT.md
2. gsd-audit-fix --dry-run           → список кандидатов на фикс с файлами
3. Ревью списка (team-lead/security-auditor)
4. gsd-audit-fix (apply)              → be1/be2 фиксят по одному finding = 1 коммит
5. test-triage на каждый FAIL после фикса
6. gsd-secure-phase                   → retroactive verify threat model
7. docs-writer: docs/security_guidelines.md + AUDIT-SCORECARD.md статус
8. /qa-gate (Hub)                      → team-lead финальная приёмка
```

### Сценарий: coverage gap найден для домена `tms`

```
1. Hub `/coverage-audit` → `.planning/COVERAGE.md` обновлён, виден risk/requirements gap TMS
2. qa-auto1/qa-auto2 назначаются на gap
3. gsd-add-tests                        → genuine Go unit в `backend/tests/tms/**` либо greenfield Python non-unit в `testframework/{api,integration,contract,e2e}/**` по taxonomy
4. Цикл: новое поведение → регрессия домена TMS → Docker-only точечный job соответствующего целевого пути
5. qa-auto-ver                           → verified: true
6. docs-writer: .planning/COVERAGE.md, docs/test_plan.md
```

### Сценарий: усиленная проверка высокорискового решения

```
Условие: декомпозиция затрагивает >4 модуля одновременно,
  ИЛИ архитектурное решение без готового ADR,
  ИЛИ финальная приёмка milestone (не фазы).

1. team-lead запускает `/task-packet`; `/context-selector` выбирает только релевантные ADR/PLAN/docs sections
2. Выбирает downstream runtime по capability/cost/risk; при высокой цене ошибки —
   conclusion-blind `/evidence-verifier` или профильный `*-ver`/`gsd-audit-milestone`
3. Результат фиксируется в ADR/PLAN/SUMMARY; дальнейшие волны — штатная оркестрация team-lead
```

### Сценарий: UI-фича с нуля (bases inline-edit)

```
1. gsd-sketch                    → designer: throwaway HTML варианты (2-3)
2. Выбор варианта пользователем/PM
3. gsd-ui-phase                  → designer: UI-SPEC.md (токены --hub-*, состояния, a11y)
4. designer-ver                   → BLOCK/FLAG/PASS для UI-SPEC до реализации
5. fe1/fe2 implement по UI-SPEC.md
6. gsd-ui-review                  → fe-ver / designer-ver: 6-pillar аудит после реализации
7. Genuine Vitest/RTL unit: `frontend/tests/bases/**`
8. Greenfield Python E2E: `testframework/e2e/bases/**`
9. docs-writer: wiki/SHORTCUTS.md если новый хоткей
```

### Сценарий: паника в production path (ws hub)

```
1. Симптом: паника в логах ws/hub.go под нагрузкой
2. .cursor/skills/test-triage/SKILL.md — воспроизвести НЕ через "почему тест не падает", а через docs (Fiber WS — не менять на net/http, канон)
3. gsd-debug                       → debugger: систематический root-cause (race, nil pointer, closed channel)
4. Фикс минимальный (rule 10-style.mdc: no panic() in production paths)
5. Regression: genuine Go units в `backend/tests/ws/**` и Python boundary scenarios в `testframework/integration/ws/**`; все команды Docker-only, race только для Go concurrency-sensitive units
6. be-ver + security-auditor (если паника эксплуатируема — DoS)
7. docs-writer: STATE.md + при необходимости observability.md (алерты/метрики на этот паттерн)
```

---

## Глоссарий: термины GSD ↔ термины Hub

| GSD термин | Эквивалент в Hub | Комментарий |
|------------|------------------|-------------|
| Phase | Фаза (`.planning/phases/NN-*/PLAN.md`) | В Hub фазы нумеруются сквозно (1-125+), не сбрасываются на milestone |
| Plan | План внутри фазы, атомарные коммиты | PR = фаза, коммиты внутри = планы (rule `50-git.mdc`) |
| Milestone | v1.0 / v2.0 / v2.5 / v2.6 / v3.0 / v4 | Архивируется в `.planning/milestones/` |
| Gray area | Серая зона | `.planning/decisions/gray-zones/AREA-*.md` + `GRAY-ZONES-DECISIONS-*.md` |
| Verification loop | Цикл верификации | REGULATION.md §3, §16 (цикличное качество) |
| UAT | Приёмка пользователем | `qa-manual` + `gsd-verify-work` |
| Seed | Отложенная идея с триггером | `.planning/reference/seeds/SEED-*.md` |
| Workstream | Параллельный поток разработки | Track F / Track P в `july-2026/ROADMAP-FINAL-2026-07-03.md` |
| Backlog parking lot | Бэклог 999.x | `.planning/archive/roadmaps/FEATURE-BACKLOG-2026.md` |

---

*Updated: 2026-08-08 — live GSD/agent roster, bounded Hub prompt skills, Superpowers-as-pattern-source и founder greenfield test paths; historical P0/P1 inventory remains a dated 2026-07-22 snapshot.*
