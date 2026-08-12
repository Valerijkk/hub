# Acceptance — AI-expansion (Must + maximal expand)

> Дата: 2026-07-22 (Must) · re-verify **2026-07-23** (волна maximal expand)
> **Provenance correction:** 2026-07-26 — qualified verdict; historical PASS **не стирается**
> Criteria: `01-PRODUCT-SCOPE.md` · Горизонт → 15.08.2026
> Inventory: `03-GSD-PORT-INVENTORY.md` (P0=21 / P1=35 / P2=15 / SKIP=8)

## Verdict

| Layer | 2026-07-23 (historical) | 2026-07-26 (current) |
|-------|-------------------------|----------------------|
| Must G1–G6 | **PASS** | **QUALIFIED PASS** — structural/package gates pass; runtime/effectiveness gates remain |
| Maximal expand M1–M7 | PARTIAL | PARTIAL (unchanged) |
| Runtime model routing | open | **APPLIED 2026-07-26**; eval pilot pending |

**Qualified PASS** = Must artifacts and structural/package gates pass.
Unqualified sign-off waits for runtime provenance, paired pilot and approvals.

Repo-scan note: W8-ai-expansion echo этого PASS — **не independent verification**; см. [`.planning/audits/2026-07/repo-scan/PROVENANCE.md`](../../../audits/2026-07/repo-scan/PROVENANCE.md).

### Must (G1–G6)

| # | Criterion (Must) | Evidence | 2026-07-23 | 2026-07-26 |
|---|------------------|----------|------------|------------|
| G1 | Инвентарь GSD + план порта | `03-GSD-PORT-INVENTORY.md` (79 skills) | PASS | PASS |
| G2 | GSD-секции в agent prompts | 32/32 `.cursor/agents/*` (opus — pointer на team-lead) | PASS | **QUALIFIED** — sizing drift risk; target concise/progressive disclosure |
| G3 | PE corporate + GSD playbooks/skills | PE corporate agents; playbooks `gsd-*`; skills phase-plan/wave/verify | PASS | **QUALIFIED** — см. structural gaps |
| G4 | GSD-RU pack | `.cursor/skills/gsd-ru/` (**6** skills + README) | PASS | **PASS** — name=folder; routes validated |
| G5 | Portable kit + reduced-GitHub | `05`, `06`, `09`, packager | PASS | **QUALIFIED** — 172-file package validates (incl. task-packet builder) and excludes local evidence; paired pilot pending |
| G6 | Dynamic model routing | ADR + agents/playbooks | PASS (historical checklist) | **PASS** — Sol Pro + 31 inherit |

### Maximal expand (2026-07-23)

| # | Criterion | Evidence | Status |
|---|-----------|----------|--------|
| M1 | Corporate prompts ≥100 | `prompts/13-corporate-delivery.md` **#1–112** | PASS |
| M2 | PE +QA/review/debug agents | `qa-manual`, `qa-auto`, `code-reviewer`, `debugger` (+7 corporate +4 test) | PASS |
| M3 | Dry-run portable kit (docs) | `08-PORTABLE-KIT-DRY-RUN.md` | PASS |
| M4 | P0/P1 status в маршрутизации | `.cursor/GSD-INTEGRATION.md` § «P0/P1 port status mid-Aug» | PASS |
| M5 | P0 в gsd-ru (ядро цикла) | 6/21 upstream P0 → `phase-plan` / `wave-run` / `goal-verify` / `handoff` | PASS (ядро) |
| M6 | Остальные P0 → agents (не RU skill) | 15/21 — practice/route (ba, qa, code-reviewer, security, …) | PARTIAL — углубление до 15.08 |
| M7 | P1 routes | 35 skills — преимущественно route-only | open (W2–W4 polish) |

## Structural defects (closed) and remaining release gates

> Historical structural defects are retained below. Remaining blockers are
> runtime/effectiveness/approval gates, not hidden structural PASS claims.

| Defect | Impact | Remediation | Status |
|--------|--------|-------------|--------|
| Broken `corporate-delivery` link / reference | PE navigation drift | Playbook created; links validated | **closed** |
| GSD skill name mismatch | Skills undiscoverable/ambiguous | 6 names = folders | **closed** |
| Phantom GSD routes | False completeness | Active docs search = 0 | **closed** |
| Unfilled context architecture | Kit not operational | Hub profile complete with 49 file hashes + 8 output hashes | **closed** |
| No eval harness | Cannot claim automation-ready | Paired eval suite created | **closed structurally**; results `NOT_RUN` |
| False frontmatter portability | Runtime incompatibility | Portable core model-agnostic + adapters; Hub 1+31 routing applied | **closed** |
| Closed MR scanning (removed from intent) | Was never required for Must | Explicit **no MR mining** in README/`02` | **closed** (policy) |

## Updated product intent (2026-07-26)

| Was implied | Now |
|-------------|-----|
| PE = corporate delivery only | PE = **portable AI quality layer for test automation** |
| Auto context from git history | **Path bootstrap** → humans add good/bad code + colleague comments |
| — | **No** closed MR scanning |

## Дополнительно

| Item | Evidence | Status |
|------|----------|--------|
| Product scope MoSCoW | `01-PRODUCT-SCOPE.md` | PASS |
| Tech plan atomic | `02-TECH-PLAN.md` (26 tasks incl. structural/package/eval gates) | PASS |
| Portable arch | `04-PORTABLE-HUB-ARCH.md` | PASS |
| GSD-INTEGRATION / gsd-index | gsd-ru + ai-expansion + P0/P1 mid-Aug | **QUALIFIED** |
| STATE / INDEX july | session notes | PASS |
| Prod code untouched | no backend/frontend in wave | PASS |

## Gaps

| Gap | Priority | Status |
|-----|----------|--------|
| Полная таблица 79 upstream skills | P1 | **CLOSED** — `03` |
| ≥100 corporate prompts | P1 | **CLOSED** — prompts/13 = **112** |
| Dry-run kit (docs) | P2 | **CLOSED** — `08-PORTABLE-KIT-DRY-RUN.md` |
| PE QA/review/debug corporate | P1 | **CLOSED** |
| P0/P1 статус в GSD-INTEGRATION | P0 | **CLOSED** — § mid-Aug |
| Expanded 04/06 blueprints (≥150 LOC, FR/API/C4) | P1 | **CLOSED** — applied from Blueprints |
| 15 P0 паттернов только в agents (не gsd-ru skill) | P0 | open polish → 15.08 |
| P1 route polish | P1 | open W2–W4 |
| Runtime frontmatter/model routing | W4 | **closed 2026-07-26** |
| Structural defects table above | P0 | **closed structurally** |
| PE eval harness | P1 | **implemented; effectiveness NOT_RUN** |

## Next

1. Physical empty-repo install dry-run (package → validate → inventory → bootstrap).
2. Прогнать representative baseline-vs-assisted eval pilot; до него ROI `NOT_RUN`.
3. Улучшить assisted smoke provenance: content grounding PASS, runtime trace PARTIAL.
4. Закрыть оставшийся P0/P1 route/content polish по фактическому inventory.
5. Коммит только по запросу пользователя.

UX P0 (inventory omission, hand-fill vs builder, package missing builder) —
closed 2026-07-26 evening. Structural/package gates remain PASS; effectiveness
gates unchanged (`NOT_RUN`).

*Обновлено: 2026-07-26 — qualified acceptance; UX P0 closed; historical 2026-07-23 PASS preserved.*
