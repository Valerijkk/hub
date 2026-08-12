# TASKS — Закрытие всех пробелов Hub

> Orchestrator: 2026-06-17 | План: `.planning/archive/tasks/PLAN-CLOSE-GAPS.md`
> Верификация gate: **2026-06-20** · Wave 6: **2026-06-20**
> Правило: test-triage при FAIL — `.cursor/skills/test-triage/SKILL.md`

## Wave 2 — Backend Coverage (qa + be) ✅ COMPLETE

| ID | Задача | Агент | Критерий готовности | Статус |
|----|--------|-------|---------------------|--------|
| T2-01 | Тесты: calls webhook, messenger rate, storage presign | `qa` | +coverage, go test green | ✅ |
| T2-02 | Тесты: platform/nats, apigateway middleware order | `qa` | SEC-01 integration test | ✅ |
| T2-03 | `Timed()` на auth/audit/tracker hot paths | `be` | observability.Timed wrapped | ✅ |
| T2-04 | BE-H6/BE-H7 docs fix в security_guidelines | `docs-writer` | NOT FOUND → FIXED | ✅ |

## Wave 3 — Frontend Quality (fe + qa) ✅ COMPLETE

| ID | Задача | Агент | Критерий готовности | Статус |
|----|--------|-------|---------------------|--------|
| T3-01 | Vitest: ≥12 tests (auth, admin, hooks) | `fe` | vitest run green | ✅ 36 tests |
| T3-02 | ErrorBoundary: tracker, tms, calls, dashboard routes | `fe` | 4 route wrappers | ✅ |
| T3-03 | Playwright e2e: 5 critical paths | `qa` | `frontend/e2e/` scaffold | ✅ 6 specs |
| T3-04 | Biome: fix top 30 critical lint | `fe` | lint errors reduced | ⚠️ partial (not gated) |

## Wave 4 — Observability + Docs (be + fe + docs) ✅ COMPLETE

| ID | Задача | Агент | Критерий готовности | Статус |
|----|--------|-------|---------------------|--------|
| T4-01 | Audit: filter `slow=1` (duration_ms>500) | `be` | query param works | ✅ |
| T4-02 | Admin UI: slow requests filter | `fe` | checkbox `slow=1` | ✅ |
| T4-03 | `docs/wiki/API.md` sync +15 routes | `docs-writer` | count matches code | ✅ |
| T4-04 | `.planning/COVERAGE.md` refresh | `sa` | COV-01 updated % | ✅ 56.0% (2026-06-20) |

## Wave 5 — Gate (orchestrator) ✅ COMPLETE

| ID | Задача | Агент | Критерий | Статус |
|----|--------|-------|----------|--------|
| T5-01 | Full qa-gate | orchestrator | build + test -race + tsc + vitest | ✅ 2026-06-20 |
| T5-02 | verifier acceptance | `verifier` | PLAN success criteria | ✅ see matrix below |
| T5-03 | STATE.md handoff | `docs-writer` | STATE updated | ✅ |
| T5-04 | Coverage auth/rbac 50%+ | `qa` | auth 83.6%, rbac 85% | ✅ |
| T5-05 | Vitest 10 files / 36 tests | `fe` | ≥10 files | ✅ |
| T5-06 | Playwright a11y axe | `qa` | a11y.spec.ts | ✅ |
| T5-07 | audit/ws/calls security uplift | `qa` | audit 80.5%, ws 88%, calls 72.5% | ✅ partial (calls residual) |

## Exit Gate Matrix (verified 2026-06-20)

| Критерий | Target | Actual | Verdict |
|----------|--------|--------|---------|
| **COV-01 overall** | ≥50% stmt | **56.6%** | ✅ PASS |
| **COV-01 security** | ≥80% per pkg | auth 83.6%, ws 88%, rbac 85%, audit 80.5%, middleware 84%, calls **80.2%** | ✅ PASS |
| **FE-TEST** | ≥10 Vitest files | 10 files, 36 tests | ✅ PASS |
| **E2E-01** | 5 Playwright paths | 6 specs in `frontend/e2e/` | ✅ PASS |
| **REGRESSION** | build + test -race + tsc | all green | ✅ PASS |
| **NFR-16** | a11y baseline | login axe + a11y.spec.ts | ✅ PASS |
| **SEC-PROD** | 0 CRITICAL open | fixed in code (v2.0.1) | ✅ (formal re-audit optional) |
| **Biome lint** | reduced | **0 errors / 0 warnings** (Phase 37) | ✅ gated |

COV-01 security: **6/6** packages ≥80% ✅ (PG-01 closed 2026-06-20).

## Wave 6 — Prod cutover + UX-0 (2026-06-20) ✅ CODE COMPLETE

| ID | Задача | Агент | Критерий | Статус |
|----|--------|-------|----------|--------|
| T6-01 | Dockerfile `-tags release` | `devops-engineer` | prod build excludes `/dev/*` | ✅ |
| T6-02 | CI: race 600s, release build, E2E Mongo authSource | `devops-engineer` | `ci.yml` green | ✅ |
| T6-03 | Deploy: workflow_run gate, SHA tags, safe prune | `devops-engineer` | deploy only after CI | ✅ |
| T6-04 | Prod compose env: CORS, ENCRYPTION_KEY, ENVIRONMENT | `devops-engineer` | `.env.example` synced | ✅ |
| T6-05 | PG-01 calls → 80% | `qa` | calls **80.2%** | ✅ |
| T6-06 | US-M07 Ctrl+K command palette | `fe` | Ctrl/Cmd+K + topbar trigger | ✅ |
| T6-07 | Docs sync (rule 55 + planning/wiki) | `docs-writer` | STATE/COVERAGE/DEPLOY updated | ✅ |

## Post-gate backlog

| ID | Задача | Target |
|----|--------|--------|
| ~~PG-01~~ | ~~`calls/` → 80%~~ | ✅ **80.2%** (2026-06-20) |
| PG-02 | Biome lint → 0 warnings | ✅ Phase 37 complete |
| PG-03 | Secret rotation git history | manual ops |
| ~~PG-04~~ | ~~E2E-06 Ctrl+K Playwright~~ | ✅ `command-palette.spec.ts` (2026-06-22) |
| PG-05 | UserPicker + inline validation | ✅ issue-form RHF+Zod + Vitest (Phase 59) |
