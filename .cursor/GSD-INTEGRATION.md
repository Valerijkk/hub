# Hub ↔ GSD Skills Integration

> GSD skills: `~/.claude/skills/gsd-*/SKILL.md` · Hub slash-skills: `.cursor/skills/*/SKILL.md`.
> Ростер команды: `.cursor/playbooks/README.md` · Регламент: `.cursor/REGULATION.md`.

*Обновлено: 2026-07-02 — новый ростер 30 агентов + playbooks.*

---

## Orchestrator (team-lead / пользователь)

| Intent | GSD Skill | Hub Slash |
|--------|-----------|-----------|
| Не знаю какую команду | `gsd-do` | — |
| Статус проекта | `gsd-progress` | — |
| Следующий шаг | `gsd-next` | — |
| Волны/исполнение | `gsd-execute-phase`, `gsd-workstreams` | — |
| Мелкая задача | `gsd-quick`, `gsd-fast` | — |
| Синхронизация STATE | `gsd-handoff` via | `/handoff-sync` |
| Полный quality gate | — | `/qa-gate` |
| Матрица покрытия | `gsd-coverage-audit` | `/coverage-audit` |

---

## Ростер 30 агентов → GSD + playbook

| Agent | Primary GSD | Secondary GSD | Playbook |
|-------|-------------|---------------|----------|
| `team-lead` | `gsd-do`, `gsd-execute-phase`, `gsd-autonomous` | `gsd-workstreams`, `gsd-progress` | `team-workflow.md` |
| `tech-lead` | `gsd-plan-phase`, `gsd-analyze-dependencies` | `gsd-insert-phase`, `gsd-execute-phase` | `team-workflow.md` |
| `product` | `gsd-new-milestone`, `gsd-review-backlog` | `gsd-explore`, `gsd-plant-seed` | `analysis.md`, `team-workflow.md` |
| `product-ver` | `gsd-verify-work` | `gsd-audit-milestone` | `verification.md` |
| `ba1`, `ba2` | `gsd-discuss-phase`, `gsd-spec-phase` | `gsd-review-backlog` | `analysis.md` |
| `ba-ver` | `gsd-verify-work` | — | `verification.md` |
| `sa1`, `sa2` | `gsd-coverage-audit`, `gsd-intel` | `gsd-audit-uat`, `gsd-map-codebase` | `analysis.md` |
| `sa-ver` | `gsd-verify-work` | `gsd-audit-uat` | `verification.md` |
| `arch` | `gsd-ai-integration-phase`, `gsd-ui-phase` | `gsd-research-phase`, `gsd-select-framework` | `analysis.md`, `backend.md`, `frontend.md` |
| `arch-ver` | `gsd-verify-work` | `gsd-secure-phase` | `verification.md` |
| `designer` | `gsd-sketch`, `gsd-ui-phase` | `gsd-ui-review`, `gsd-sketch-wrap-up` | `design.md` |
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
| `code-reviewer` | `gsd-code-review`, `gsd-code-review-fix` | — | `verification.md`, `code-comments.md`, `git-workflow.md` |
| `docs-writer` | `gsd-docs-update`, `gsd-milestone-summary` | — | `code-comments.md` |

---

## Планы Hub в GSD-терминах

| Hub Artifact | GSD Equivalent |
|--------------|----------------|
| `.planning/ROADMAP.md` | Milestone phases |
| `.planning/STATE.md` | `gsd-progress` state |
| `.planning/COVERAGE.md` | `gsd-coverage-audit` output |
| `.planning/EXPANSION-ANALYSIS-*.md` | backlog для `gsd-plan-phase` |
| `.planning/quick/` | `gsd-quick` tasks |
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

*Updated: 2026-07-02*
