# TECH PLAN — AI Expansion (22.07 → 15.08.2026)

## Цель

Порт паттернов GSD + расширение AI-команды/PE + portable kit, без продового кода.

**Уточнение 2026-07-26:** `pro-qa-ai/` — portable **AI quality
layer для test automation**: bootstrap создаёт `context/projects/<profile>/`
из явных roots; humans вручную добавляют good/bad code + colleague comments
после bootstrap; **без** closed MR scanning.

## Timeline

| Неделя | Даты | Фокус |
|--------|------|-------|
| W1 | 22–28.07 | Inventory + scope + agents GSD-секции + GSD-RU skeleton |
| W2 | 29.07–04.08 | PE corporate agents + GSD playbooks/skills |
| W3 | 05–11.08 | Portable kit + reduced-GitHub blueprint + microservices checklist |
| W4 | 12–15.08 | Eval pilot, portable dry-run, final acceptance; runtime routing уже applied 2026-07-26 |

## Задачи (атомарные)

| # | Задача | Исполнитель | Файлы | Depends | Оценка |
|---|--------|-------------|-------|---------|--------|
| 1 | Product scope MoSCoW | product | `ai-expansion/01-PRODUCT-SCOPE.md` | — | 0.5д |
| 2 | Tech plan (этот файл) | tech-lead | `02-TECH-PLAN.md` | 1 | 0.5д |
| 3 | GSD inventory P0/P1/P2 | research | `03-GSD-PORT-INVENTORY.md` | — | 1д |
| 4 | Portable hub arch | arch | `04-PORTABLE-HUB-ARCH.md` | 1 | 1д |
| 5 | GSD-секции leadership agents | docs | `team-lead,tech-lead,product,product-ver,docs-writer` | 3 | 0.5д |
| 6 | GSD-секции builders/QA | docs | `be*,fe*,qa*,debugger,code-reviewer` | 3 | 0.5д |
| 7 | GSD-секции analysis/ver/support | docs | `ba*,sa*,arch*,designer*,*-ver,devops,security` | 3 | 0.5д |
| 8 | GSD-RU skills pack | docs | `.cursor/skills/gsd-ru/**` | 3 | 1д |
| 9 | Update gsd-index | docs | `gsd-index/SKILL.md` | 8 | 0.2д |
| 10 | PE playbook gsd-lifecycle | docs | `pro-qa-ai/workflows/playbooks/gsd-lifecycle.md` | 3 | 0.5д |
| 11 | PE playbook wave-execution | docs | `playbooks/wave-execution.md` | 3 | 0.5д |
| 12 | PE playbook handoff-state | docs | `playbooks/handoff-state.md` | 3 | 0.5д |
| 13 | PE skills phase-plan/wave/verify | docs | `skills/{phase-plan,wave-orchestrate,verify-goal}/` | 10–12 | 1д |
| 14 | PE templates phase/wave | docs | `agents/templates/{phase-kickoff,wave-delegate}.md` | 13 | 0.3д |
| 15 | PE corporate agents (7) | docs | `agents/{orchestrator,backend-dev,…}.md` | 1 | 1д |
| 16 | PE prompts/13-corporate-delivery | docs | `prompts/13-corporate-delivery.md` | 15 | 0.5д |
| 17 | PE gsd-port README + index updates | docs | `gsd-port/README.md`, PE READMEs | 10–16 | 0.3д |
| 18 | Portable kit spec | docs | `05-PORTABLE-KIT.md`, `PORTABLE-HUB-KIT.md` | 4,8,15 | 0.5д |
| 19 | Reduced-GitHub blueprint | product+arch | `06-REDUCED-GITHUB-MODULE.md` | 4 | 0.5д |
| 20 | GSD-INTEGRATION + STATE + INDEX | docs | `.cursor/GSD-INTEGRATION.md`, `STATE.md`, `july-2026/INDEX` | 1–19 | 0.5д |
| 21 | Dynamic model routing | team-lead | agents/frontmatter + ADR + playbooks | ADR | **done 26.07** |
| 22 | Приёмка product-ver + team-lead | product-ver, team-lead | — | all | 0.5д |
| 23 | PE test-automation quality layer | docs | context/profile/patterns/recipes/validator/evals | 18 | **done 26.07** |
| 24 | Structural remediation | docs | links, skill names, routes, frontmatter | 22 | **done 26.07** |
| 25 | Paired eval pilot | automation-quality + human reviewer | `evals/` run records | 23,24 | W4 pending |
| 26 | Portable packager + package validation | automation-quality | `scripts/package_*`; temp package | 23,24 | **done 26.07** |

## Волны

```
Wave A (∥): #1, #3, #4
Wave B (∥): #5, #6, #7, #8
Wave C (∥): #10–#16
Wave D: #9, #17–#21 → #22
Wave E: #23, #24, #26 — structural/package layer complete
Wave F (W4): #25 — paired effectiveness/ROI pilot
```

## Success criteria

- [ ] Все Must из `01-PRODUCT-SCOPE.md` (**qualified** — см. `07-ACCEPTANCE.md`)
- [ ] Нет правок `backend/` / `frontend/src` в рамках этой фазы
- [x] Structural validator: links/frontmatter/skills/routes — PASS
- [ ] Portable kit чек-лист проходим «на бумаге» без Hub-хардкода в PE agents
- [x] Context bootstrap implemented; Hub profile complete; closed-MR access = 0
- [x] GSD skill names/routes соответствуют реальным installed/Hub skills

## Риски

| R | Mitigation |
|---|------------|
| Конфликт правок agents | Разделение ownership по файлам в волнах |
| PE потеряет обезличенность | Review: grep Hub-путей в agents/skills PE |
| Слишком много текста в agents | **Concise GSD sections + progressive disclosure** — ссылки на playbooks/skills вместо дублирования; **не** гнаться за 40–80 строк; цель — минимум для handoff, детали в `.cursor/skills/` |
| Skill name / route drift | Inventory `03` ↔ `GSD-INTEGRATION` ↔ agent frontmatter — task #24 |
| False portability claims | Frontmatter `model:` Hub-specific — document as non-portable unless stripped (#24) |
| Effectiveness not measured | Harness exists; task #25 keeps ROI `NOT_RUN` until paired pilot |

## Verification

| Артефакт | Verifier |
|----------|----------|
| Scope | product-ver |
| Arch | arch-ver |
| Agents/PE/docs | docs-writer self-check + team-lead |
| Финал | team-lead (**qualified** until task #25 paired pilot) |

## Model routing status

1. `[done]` ADR amended: Sol Pro orchestrator + dynamic downstream.
2. `[done]` `team-lead model: gpt-5.6-sol`; 31 downstream `model: inherit`.
3. `[done]` `team-lead-opus` deprecated redirect.
4. `[done]` AGENTS/REGULATION/playbooks synchronized.
5. `[pending]` Hub-profile model-routing/eval pilot; structural PASS его не заменяет.

*Обновлено: 2026-07-26 — PE test-automation intent; GSD section sizing → concise/progressive disclosure.*
