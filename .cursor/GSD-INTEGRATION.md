# Hub ↔ GSD Skills Integration

> GSD skills live at `~/.claude/skills/gsd-*/SKILL.md` (Claude Code).
> Hub slash-skills live at `.cursor/skills/*/SKILL.md` (Cursor).
> Регламент: `.cursor/REGULATION.md` §9–§10.

---

## Orchestrator (любой агент / пользователь)

| Intent | GSD Skill | Hub Slash |
|--------|-----------|-----------|
| Не знаю какую команду | `gsd-do` | — |
| Статус проекта | `gsd-progress` | — |
| Следующий шаг | `gsd-next` | — |
| Мелкая задача | `gsd-quick` | — |
| Синхронизация STATE | `gsd-handoff` via | `/handoff-sync` |
| Полный quality gate | — | `/qa-gate` |
| Матрица покрытия | `gsd-coverage-audit` | `/coverage-audit` |

---

## 12 агентов регламента → GSD

| Agent | Primary GSD | Secondary GSD | Hub Skill |
|-------|-------------|---------------|-----------|
| `ba` | `gsd-discuss-phase`, `gsd-spec-phase` | `gsd-new-milestone`, `gsd-review-backlog` | `BA_skill.md` |
| `sa` | `gsd-coverage-audit`, `gsd-intel` | `gsd-audit-uat`, `gsd-map-codebase` | `SA_skill.md` |
| `arch` | `gsd-ai-integration-phase`, `gsd-ui-phase` | `gsd-research-phase` | `ARCH_skill.md` |
| `be` | `gsd-quick`, `gsd-debug` | `gsd-fast`, `gsd-add-tests` | `BE_skill.md` |
| `fe` | `gsd-quick`, `gsd-ui-review` | `gsd-sketch`, `gsd-ui-phase` | `FE_skill.md` |
| `qa` | `gsd-add-tests`, `gsd-validate-phase` | `gsd-audit-fix`, `gsd-audit-uat` | `QA_skill.md` |
| `ba-ver` | `gsd-verify-work` | — | `BA_VER_skill.md` |
| `sa-ver` | `gsd-verify-work` | `gsd-audit-uat` | `SA_VER_skill.md` |
| `arch-ver` | `gsd-verify-work` | `gsd-secure-phase` | `ARCH_VER_skill.md` |
| `be-ver` | `gsd-verify-work` | `gsd-code-review` | `BE_VER_skill.md` |
| `fe-ver` | `gsd-verify-work` | `gsd-ui-review` | `FE_VER_skill.md` |
| `qa-ver` | `gsd-verify-work` | `gsd-validate-phase` | `QA_VER_skill.md` |

---

## Вспомогательные агенты → GSD

| Agent | GSD Skills |
|-------|------------|
| `tech-lead` | `gsd-plan-phase`, `gsd-execute-phase`, `gsd-workstreams`, `gsd-insert-phase` |
| `debugger` | `gsd-debug` |
| `security-auditor` | `gsd-secure-phase`, `gsd-audit-fix` |
| `code-reviewer` | `gsd-code-review`, `gsd-code-review-fix` |
| `verifier` | `gsd-verify-work` |
| `docs-writer` | `gsd-docs-update`, `gsd-milestone-summary` |
| `devops-engineer` | `gsd-ship` (deploy часть) |
| `refactor-specialist` | `gsd-quick` (batch refactor) |

---

## Планы Hub в GSD-терминах

| Hub Artifact | GSD Equivalent |
|--------------|------------------|
| `.planning/PLAN-CLOSE-GAPS.md` | Multi-phase plan (manual) |
| `.planning/ROADMAP.md` | Milestone phases |
| `.planning/STATE.md` | `gsd-progress` state |
| `.planning/COVERAGE.md` | `gsd-coverage-audit` output |
| `.planning/quick/` | `gsd-quick` tasks (create if needed) |
| `.planning/debug/` | `gsd-debug` sessions |

---

## Правило: Test Triage (обязательно)

**Перед любым фиксом упавшего теста** — прочитать `.cursor/skills/test-triage/SKILL.md`.

Кратко:
1. Воспроизвести поведение функционала (не теста)
2. Сверить с `docs/system_specification.md` / требованиями
3. Код верен → чинить тест; код неверен → чинить код
4. Запрет: «зелёные тесты» без проверки шагов 1–2

---

## Wave execution (gsd-execute-phase pattern)

```
Wave N: spawn parallel agents (max 4)
  → each runs test-triage on FAIL
  → each runs /qa-gate subset on touched packages
  → orchestrator merges + gsd-handoff-sync
Wave N+1: depends on Wave N artifacts
```

Текущий план: `.planning/PLAN-CLOSE-GAPS.md`

---

*Updated: 2026-06-17*
