# Источники Touchstone Pro

Карта «откуда взято лучшее». Пути — относительно корня репозитория Hub.
В portable поставке Pro пути Standard остаются `../prompt-engineering/…`.

## GSD

| Паттерн | Hub-источник | Pro-слой |
|---------|--------------|----------|
| Inventory 79 skills → P0/P1 | `.planning/archive/july-2026/ai-expansion/03-GSD-PORT-INVENTORY.md` | 02 |
| GSD-RU pack | `.cursor/skills/gsd-ru/` | 02, 04 |
| Integration matrix | `.cursor/GSD-INTEGRATION.md` | 02 |
| PE gsd-port | `prompt-engineering/gsd-port/` | 02 |
| Lifecycle / waves / handoff playbooks | `prompt-engineering/playbooks/{gsd-lifecycle,wave-execution,handoff-state}.md` | 02, 04 |

## Prompt Engineering (Standard)

| Паттерн | Источник | Pro-слой |
|---------|----------|----------|
| AI quality layer | `prompt-engineering/` целиком | 01 |
| Master delegate | `prompt-engineering/agents/templates/_MASTER-PROMPT-TEMPLATE.md` | 03 |
| Rules 00–50 | `prompt-engineering/rules/` | 01, 05 |
| Eval fixtures | `prompt-engineering/evals/` | 05 |
| Rework decisions | `.planning/archive/july-2026/ai-expansion/09-AI-QUALITY-LAYER-REWORK.md` | 01 |

## Cursor Hub

| Паттерн | Источник | Pro-слой |
|---------|----------|----------|
| Prompt contract | `.cursor/playbooks/prompt-contract.md` | 03 |
| Team workflow / RACI | `.cursor/playbooks/team-workflow.md` | 04 |
| Verification | `.cursor/playbooks/verification.md` | 05 |
| Regression cycle | `.cursor/rules/56-regression-verify.mdc` | 05 |
| Docs sync | `.cursor/rules/55-docs-sync.mdc` | 06 |
| Testing craft | `.cursor/playbooks/testing.md` | 01, 05 |
| Git / atomic / no AI co-author | `.cursor/rules/50-git.mdc`, `51-no-ai-coauthor.mdc` | 04 |
| Regulation §20 | `.cursor/REGULATION.md` | 07 |

## Planning

| Паттерн | Источник | Pro-слой |
|---------|----------|----------|
| Backlog task format | `.cursor/rules/57-backlog-tasks.mdc`, `.planning/backlog/` | 06 |
| STATE handoff | `.planning/STATE.md` | 06 |
| Phase honesty / matrix | `.planning/audits/planning-audit-2026-07/` | 06 |
| Mutation evidence practice | `.planning/audits/total-2026-07/STAGE-C-MUTATION-*.md` | 05 |
| AI-expansion acceptance | `.planning/archive/july-2026/ai-expansion/07-ACCEPTANCE.md` | 01 |

## Внешние

| Источник | Использование |
|----------|---------------|
| [gsd-build/get-shit-done](https://github.com/gsd-build/get-shit-done) | Паттерны lifecycle (не полный vendor-copy) |
| `prompt-engineering/sources/PRIMARY-SOURCES.md` | Primary sources Standard |
