# Phase Status Matrix - live scan (A3)

> **Auditor:** Grok 4.5 / branch `cursor-grok-4.5`
> **Scanned:** 2026-07-27 00:34 +0300
> **Source:** `.planning/phases/*/` - counts from filesystem NOW (not copied from prior audits)
> **Scope:** documentation only; no production code touched
> **Note:** User expected ~113 dirs; live count at scan = **124** (includes backlog `999-test-framework` and any phases added after 125).

## Summary counts

| Metric | Count |
|--------|------:|
| Phase directories (`.planning/phases/*`) | **124** |
| `PLAN.md` files (recursive under phases) | 122 |
| `VERIFICATION.md` files | 70 |
| `VERIFICATION-SKIPPED.md` files | 0 |
| Phases with `PLAN.md` (dir root) | 122 |
| Phases with `VERIFICATION.md` (dir root) | 70 |
| Phases **without** verification artifact | **54** |
| Phases without `PLAN.md` | 2 |

### Classification

| classification | count |
|----------------|------:|
| VERIFIED_PASS | 63 |
| VERIFIED_FAIL | 0 |
| VERIFIED_PARTIAL | 5 |
| CODE_GATE_DEPLOY_BLOCKED | 2 |
| NO_VERIFICATION | 54 |
| SKIPPED | 0 |
| **Total** | **124** |

### Verdict token (first Verdict / Result)

| verdict | count |
|---------|------:|
| PASS | 63 |
| FAIL | 0 |
| PARTIAL | 5 |
| PASS_CODE_ONLY | 2 |
| ABSENT (has VERIFICATION.md, no extractable verdict) | 0 |
| SKIPPED | 0 |
| - (no verification file) | 54 |

### Goal section (from PLAN.md)

| goal_measurable | count |
|-----------------|------:|
| yes | 63 |
| partial | 13 |
| no | 2 |
| unknown (no Goal section / unclear) | 46 |

| goal_met | count |
|----------|------:|
| yes | 34 |
| no | 10 |
| partial | 0 |
| unknown | 80 |

## Method

1. Enumerate every immediate subdirectory of `.planning/phases/` (Python `Path.iterdir()`).
2. `has_plan` / `has_verification` = presence of `PLAN.md` / `VERIFICATION.md` at phase dir root.
3. `SKIPPED` only if `VERIFICATION-SKIPPED.md` exists (found: 0).
4. First verdict: inline `Verdict:`, else first substantive line under `## Verdict`, else `GATE PASSED/FAILED`, else header `Result:` when it states ACHIEVED/missed.
5. `PASS (code gate)` + deploy blocked -> `PASS_CODE_ONLY` / `CODE_GATE_DEPLOY_BLOCKED`.
6. `PASS (partial ...)` / explicit `PARTIAL` / stretch goal missed -> `VERIFIED_PARTIAL`.
7. Goal: parse `## Goal` (or `goal:` frontmatter); measurable = numeric/%/target language; met inferred from Goal text + VERIFICATION outcome when present.

## Phases without VERIFICATION.md

Count: **54** (PW2 2026-07-27 closed 14–21)

- `22-arch-storage`
- `23-arch-apigateway`
- `24-arch-imex`
- `25-arch-dashboard`
- `26-arch-rbac`
- `27-arch-auth`
- `28-v26-arch-gate`
- `29-obs-01`
- `30-obs-01-fe`
- `31-obs-prometheus`
- `32-obs-otel`
- `33-monitoring-stack`
- `34-loki-alerts`
- `35-alertmanager`
- `36-alert-webhook`
- `37-biome-gate`
- `38-v26-milestone-gate`
- `39-intranet-news`
- `40-org-directory`
- `41-unified-calendar`
- `42-crm-lite`
- `43-hr-lite`
- `44-kudos`
- `45-marketplace`
- `46-status-page`
- `47-random-coffee`
- `48-onec-connector`
- `49-lark-base`
- `51-accounting-mvp`
- `52-accounting-v45`
- `59-coverage-ux-closeout`
- `60-ux1-gate`
- `61-coverage-tooltips`
- `62-onboarding-profile-rhf`
- `64-auth-coverage`
- `65-ux-rate-limit-messenger-cov`
- `66-notification-coverage`
- `68-messenger-fanout-cov`
- `122-tms-project-82`
- `123-messenger-storage-82`
- `124-floor-82`
- `125-regression-gate`
- `126-intranet-newspost-trash`
- `127-calendar-series-trash`
- `128-tf-python-scaffold`
- `129-tf-e2e-shared`
- `130-tf-e2e-mechanical`
- `131-tf-e2e-rewrite`
- `132-tf-backend-units-colocated`
- `133-tf-frontend-units-colocated`
- `134-tf-integration-python`
- `135-tf-locust-load`
- `136-tf-ci-cleanup`
- `999-test-framework`

## Phases without PLAN.md

- `14-architectural-hardening`
- `999-test-framework`

## Full matrix

| phase | has_plan | has_verification | verdict | goal_measurable | goal_met | classification | verdict_line |
|-------|----------|------------------|---------|-----------------|----------|----------------|--------------|
| 14-architectural-hardening | no (PHASE+5 plans) | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 15-v25-trust | yes | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 16-v26-automate | yes | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 17-v30-ai-native | yes | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 18-v35-ai-privacy | yes | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 19-v30-ai-polish | yes | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 20-milestone-gate | yes | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 21-arch-completion | yes | yes | PASS | yes | yes | VERIFIED_PASS | 2026-07-27 PW2 PASS |
| 22-arch-storage | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 23-arch-apigateway | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 24-arch-imex | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 25-arch-dashboard | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 26-arch-rbac | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 27-arch-auth | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 28-v26-arch-gate | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 29-obs-01 | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 30-obs-01-fe | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 31-obs-prometheus | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 32-obs-otel | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 33-monitoring-stack | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 34-loki-alerts | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 35-alertmanager | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 36-alert-webhook | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 37-biome-gate | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 38-v26-milestone-gate | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 39-intranet-news | yes | no | - | unknown | unknown | NO_VERIFICATION | - |
| 40-org-directory | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 41-unified-calendar | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 42-crm-lite | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 43-hr-lite | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 44-kudos | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 45-marketplace | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 46-status-page | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 47-random-coffee | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 48-onec-connector | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 49-lark-base | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 50-v4-milestone-gate | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 51-accounting-mvp | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 52-accounting-v45 | yes | no | - | partial | unknown | NO_VERIFICATION | - |
| 53-v4-module-gate | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 54-v4-coverage-uplift | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 55-onec-intranet-coverage | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 56-v4-final-gate | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS ? v4 Super-app milestone gate |
| 57-post-v4-ci-handoff | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 58-v4-milestone-archive | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** ? PASS |
| 59-coverage-ux-closeout | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 60-ux1-gate | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 61-coverage-tooltips | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 62-onboarding-profile-rhf | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 63-production-gate | yes | yes | PASS_CODE_ONLY | yes | no | CODE_GATE_DEPLOY_BLOCKED | **Verdict:** ? PASS (code gate) ? ?? deploy blocked on commit + PG-03 ops |
| 64-auth-coverage | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 65-ux-rate-limit-messenger-cov | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 66-notification-coverage | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 67-production-gate | yes | yes | PASS_CODE_ONLY | yes | no | CODE_GATE_DEPLOY_BLOCKED | **Verdict:** ? PASS (code gate) ? ?? deploy blocked on commit + PG-03 ops |
| 68-messenger-fanout-cov | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 69-apigateway-cov-baseline | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** ? PASS |
| 70-gitlab-github-cov | yes | yes | PASS | yes | no | VERIFIED_PASS | **PASS** ? SEC-04 test gap closed for repository + handler routing layers. |
| 71-mail-dashboard-cov | yes | yes | PASS | yes | yes | VERIFIED_PASS | **PASS** ? MAIL-01 skeleton fully test-covered; dashboard project stats path unblocked. |
| 72-production-qa-gate | yes | yes | PASS | yes | yes | VERIFIED_PASS | **GATE PASSED** ? ready for commit (pending user approval). |
| 73-platform-http-outbox-cov | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 74-ai-coverage | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 75-post-ai-qa-gate | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS ? branch ready for user-approved commit + PR |
| 76-sprint-coverage | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 77-storage-routes-cov | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 78-tracker-tms-coverage | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 79-notification-project-cov | yes | yes | PASS | yes | no | VERIFIED_PASS | **Verdict:** PASS |
| 80-backup-hr-imex-cov | yes | yes | PASS | yes | no | VERIFIED_PASS | **Verdict:** PASS |
| 81-ai-coverage-qa-gate | yes | yes | PASS | yes | no | VERIFIED_PASS | **Verdict:** PASS |
| 82-ai-bot-docsearch | yes | yes | PASS | yes | no | VERIFIED_PASS | **Verdict:** PASS |
| 83-ai-commands-search | yes | yes | PASS | yes | no | VERIFIED_PASS | **Verdict:** PASS |
| 84-vcs-bases-intranet | yes | yes | PASS | yes | no | VERIFIED_PASS | **Verdict:** PASS |
| 85-github-pr-review | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 86-coverage-uplift | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 87-tms-notif-nats | yes | yes | PASS | partial | yes | VERIFIED_PASS | **Verdict:** PASS |
| 88-tms-notif-65 | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 89-e2e-wave3 | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 90-imex-jira | yes | yes | PASS | partial | yes | VERIFIED_PASS | **Verdict:** PASS |
| 91-intranet-observability | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 92-messenger-search | yes | yes | PARTIAL | yes | no | VERIFIED_PARTIAL | **Verdict:** PASS (partial target) |
| 93-autochannels-handlers | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 94-search-meili-bot | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 95-notification-tracker-tms | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 96-tms-testrail-mongo | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 97-mongo-ci-webhook-marketplace | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 98-accounting-calendar | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 99-backup-package-75 | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 100-restore-search-tracker | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 101-sub75-package-push | yes | yes | PASS | yes | yes | VERIFIED_PASS | **Verdict:** PASS |
| 102-dev-notification-push | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 103-github-mongo-integration | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS (mongo paths skip without Docker) |
| 104-flaky-github-mongotest | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 105-boundary-package-push | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 106-triple-boundary-push | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 107-crm-dashboard-bases | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 108-cov-79-milestone | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 109-triple-boundary-push-2 | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 110-tms-ai-project | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 111-notification-hr-crm | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 112-search-messenger-tracker | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS (package targets met; COV 79.7% ? 0.3pp below 80% stretch) |
| 113-tms-project-sprint | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Verdict:** PASS |
| 114-search-messenger-tms | yes | yes | PARTIAL | unknown | unknown | VERIFIED_PARTIAL | **Verdict:** PARTIAL (80.2% vs 81% stretch goal) |
| 115-github-auth-messenger | yes | yes | PARTIAL | unknown | unknown | VERIFIED_PARTIAL | **Result:** **80.7%** ? stretch goal missed by **0.3pp** |
| 116-search-messenger-push | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Result:** **81.6%** ? goal **ACHIEVED** (+0.9pp from 80.7%) |
| 117-storage-notification-tracker | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Result:** **81.8%** overall ? **ACHIEVED** (+0.2pp from 81.6%) |
| 118-notification-tracker-push | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **Result:** **81.9%** overall ? **ACHIEVED**; notification **80.2%**, tracker **80.1%** |
| 119-notification-tracker-82 | yes | yes | PARTIAL | unknown | unknown | VERIFIED_PARTIAL | **PASS with partial package targets** ? overall COV crossed **82%**; notification essentially at target; tracker needs Phase 120 for rema... |
| 120-tracker-tms-project-82 | yes | yes | PASS | unknown | unknown | VERIFIED_PASS | **PASS overall** ? COV **82.1%**; floor packages improved (+0.5?1.6pp) but individual 82% targets need Phase 121. |
| 121-notification-tracker-82 | yes | yes | PARTIAL | unknown | unknown | VERIFIED_PARTIAL | IMPLIED: PARTIAL (goals table gaps) |
| 122-tms-project-82 | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 123-messenger-storage-82 | yes | no | - | yes | yes | NO_VERIFICATION | - |
| 124-floor-82 | yes | no | - | yes | yes | NO_VERIFICATION | - |
| 125-regression-gate | yes | no | - | yes | yes | NO_VERIFICATION | - |
| 126-intranet-newspost-trash | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 127-calendar-series-trash | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 128-tf-python-scaffold | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 129-tf-e2e-shared | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 130-tf-e2e-mechanical | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 131-tf-e2e-rewrite | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 132-tf-backend-units-colocated | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 133-tf-frontend-units-colocated | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 134-tf-integration-python | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 135-tf-locust-load | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 136-tf-ci-cleanup | yes | no | - | yes | unknown | NO_VERIFICATION | - |
| 999-test-framework | no | no | - | no | unknown | NO_VERIFICATION | - |

## Notable non-PASS verifications

- **63-production-gate** - `PASS_CODE_ONLY` / `CODE_GATE_DEPLOY_BLOCKED`: **Verdict:** ? PASS (code gate) ? ?? deploy blocked on commit + PG-03 ops
- **67-production-gate** - `PASS_CODE_ONLY` / `CODE_GATE_DEPLOY_BLOCKED`: **Verdict:** ? PASS (code gate) ? ?? deploy blocked on commit + PG-03 ops
- **92-messenger-search** - `PARTIAL` / `VERIFIED_PARTIAL`: **Verdict:** PASS (partial target)
- **114-search-messenger-tms** - `PARTIAL` / `VERIFIED_PARTIAL`: **Verdict:** PARTIAL (80.2% vs 81% stretch goal)
- **115-github-auth-messenger** - `PARTIAL` / `VERIFIED_PARTIAL`: **Result:** **80.7%** ? stretch goal missed by **0.3pp**
- **119-notification-tracker-82** - `PARTIAL` / `VERIFIED_PARTIAL`: **PASS with partial package targets** ? overall COV crossed **82%**; notification essentially at target; tracker needs Phase 120 for rema...
- **121-notification-tracker-82** - `PARTIAL` / `VERIFIED_PARTIAL`: IMPLIED: PARTIAL (goals table gaps)

---

*Generated by planning audit task A3 - live filesystem scan.*
