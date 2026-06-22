# IMPROVEMENT MASTER PLAN — Hub v2.0.2 → v2.6

> Закрывает все findings из multi-audit + AUDIT-SCORECARD.md  
> Принцип: **prod pipeline ∥ UX promises** (CTO recommendation)

---

## Waves overview

| Wave | Focus | Status | Agents / Skills |
|------|-------|--------|-----------------|
| **A** | Prod pipeline | ✅ DONE | devops, `gsd-secure-phase` |
| **B** | CI = qa-gate | ✅ DONE | `/qa-gate`, `gsd-validate-phase` |
| **UX-0** | Vision promises | ✅ DONE (2026-06-20) | fe, `gsd-ui-phase` |
| **UX-1** | Competitive catch-up | ✅ DONE (2026-06-20) | fe + qa |
| **C** | Test pyramid gaps | ✅ DONE | qa, test-triage |
| **PRODUCT** | Differentiation v2.6–v3 | 📋 PLANNED | `gsd-ai-integration-phase` |

---

## Wave UX-0 — COMPLETE ✅

| ID | Deliverable | Status |
|----|-------------|--------|
| UX-P0-01 | Global Ctrl+K → `/api/search` | ✅ |
| UX-P0-02 | UserPicker assignee | ✅ |
| UX-P0-03 | Inline validation (title) | ⚠️ partial |
| UX-P0-04 | Kanban «Move to status» + toast + color sanitize | ✅ |
| UX-P0-05 | Skip link + contrast `--t3` | ✅ |
| UX-P0-06 | `/help` route + sidebar | ✅ |
| UX-P0-07 | E2E-06 Ctrl+K spec | ✅ |
| UX-P0-08 | sanitizeHexColor + unit test | ✅ |

---

## Wave UX-1 — Competitive catch-up (2–4 weeks)

| ID | Feature | Effort | Agent |
|----|---------|--------|-------|
| UX-P1-01 | Shortcuts overlay `?` | ✅ DONE Phase 60 |
| UX-P1-02 | Command palette actions (Create issue, Go to…) | ✅ DONE | command-palette.tsx |
| UX-P1-03 | Onboarding step 4: create/join project | ✅ DONE | onboarding.lazy.tsx |
| UX-P1-04 | Contextual tooltips (JQL, WIP, Allure) | ✅ DONE Phase 61 |
| UX-P1-05 | Actionable API errors + rate limit timer | ✅ DONE Phase 65 |
| UX-P1-06 | E2E axe: messenger, board, calls, TMS | ✅ DONE Phase 60 |
| UX-P1-07 | Full RHF+Zod: issue-form, onboarding | ✅ issue-form + profile RHF + project Zod step 4 |
| UX-P1-08 | E2E-06 Ctrl+K Playwright | ✅ DONE (command-palette.spec.ts in CI) |

**Exit:** UX score ≥ 80, a11y ≥ 70, E2E-06 green.

---

## Wave C — Test pyramid

| ID | Target | Status |
|----|--------|--------|
| C-01 | E2E-06 Ctrl+K | ✅ DONE |
| C-02 | `platform/config` unit tests | ✅ DONE |
| C-03 | Component test UserPicker | ✅ DONE |
| C-04 | Expand axe E2E (3+ pages) | ✅ DONE (help, projects) |
| C-05 | PG-03 secret rotation checklist | ✅ DONE |

---

## Wave PRODUCT — v2.6 → v3.0

> **Детальный backlog:** [.planning/FEATURE-BACKLOG-2026.md](./FEATURE-BACKLOG-2026.md)  
> **Анализ:** [docs/competitive_analysis_2026.md](../docs/competitive_analysis_2026.md) · [docs/user_needs_analysis_2026.md](../docs/user_needs_analysis_2026.md)

| ID | Feature | Why |
|----|---------|-----|
| PROD-01 | Workflow automation lite | #1 vs Jira |
| PROD-02 | Flaky test detection TMS | vs TestOps |
| PROD-03 | Issue↔message bidirectional | vs Linear+Slack |
| PROD-04 | Webhook delivery dashboard | trust |
| PROD-05 | AI Chat Hub v3.0 | Slack/Notion AI |

Skill: `gsd-ai-integration-phase` when PROD-05 starts.

---

## Rules (enforcement)

| Rule | Path |
|------|------|
| Docs sync after change | `.cursor/rules/55-docs-sync.mdc` |
| Regression + verifiers | `.cursor/rules/56-regression-verify.mdc` |
| Artifact cleanup | `.cursor/rules/45-artifacts-cleanup.mdc` |

---

## Immediate next 5

1. ✅ Wave UX-0 batch (this session)
2. E2E-06 Ctrl+K Playwright
3. Onboarding step 4
4. Command palette actions
5. PG-03 secret rotation doc + ops

---

## Positioning (unchanged)

**Hub** — self-hosted workspace для 10–50 человек: tracker + TMS Allure + чат + звонки без $50/user SaaS.

**Не для:** mobile-first, Slack power users, enterprise automation.  
**Лучше:** traceability, TMS economics, data control, admin observability.  
**Хуже (closing):** keyboard UX, automation, mobile Kanban.

*Updated: 2026-06-20*
