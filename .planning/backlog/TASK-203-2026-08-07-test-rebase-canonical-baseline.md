# TASK-203 — 📏 Canonical green baseline

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **BLOCKED — baseline red, test-triage** · Фаза: **128**

## [OUTCOME]

Зафиксирован один воспроизводимый green baseline для legacy suites и Go coverage, пригодный для сравнения всех migration waves без смешения исторических 80.2/81.9/83.0%.

## [ROLE]

- Owner: `devops-engineer` — environment contract, canonical commands и baseline evidence.
- Verifier / handoff: `qa-auto-ver` — независимый rerun/reconciliation и проверка валидности profiles.

## [SCOPE AND CONSTRAINTS]

- Allowed: `.planning/audits/test-rebase/**`, `.planning/COVERAGE.md` только для dated baseline evidence; временные profiles вне repo.
- Read-only: CI/compose/Makefiles, runtime manifests и legacy suites.
- Do not touch: CI commands, tests/product code, thresholds, legacy paths, STATE/ROADMAP.
- Canonical Go run: Go 1.25, required Postgres/Mongo/Redis/NATS, `-tags hubtest -race -count=1 -covermode=atomic -coverpkg=./internal/...`; полный run обязан иметь zero failed packages.
- Separate baselines: Go tests/coverage, Vitest, TS smoke/full. Counts включают collected/executed/passed/failed/skipped, wall time и artifacts.
- Depends on: TASK-202.

## [REFERENCES]

- `.planning/codebase/{STACK,TESTING,INTEGRATIONS,CONCERNS}.md`
- `.github/workflows/ci.yml`, `deployments/docker-compose.test.yml`
- `.cursor/playbooks/testing.md`, current `.planning/COVERAGE.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 -covermode=atomic -coverpkg=./internal/... -coverprofile=../.tmp/hub-cov.out ./testframework/...`; exit code 0 и заявленные counts обязательны.

- Green service-backed commands recorded with HEAD, OS/runtime/container versions, env variable names (без values), exact package/spec scope, counts/skips, wall time and profile/report hashes.
- Go overall and security package coverage recorded only from zero-failure run; result below 82% is reported honestly and blocks Phase 128 plus all movement/scaffold until TASK-204 records an explicit founder-approved replacement contract.
- Repeat run or verifier run reproduces counts and coverage within defined tolerance; differences explained.
- Failing sentinel: malformed/empty coverage profile and empty collection are rejected by baseline validation. Production mutation N/A: task measures existing behavior and changes no code.
- Regression: N/A — evidence-only baseline; no tracked code/test/CI changes except planning artifacts.
- Stop when: Python/Docker/Go 1.25 or required services unavailable; status BLOCKED with exact missing prerequisite, never publish partial coverage as baseline.

## [OUTPUT SCHEMA]

1. PASS/BLOCKED and summary.
2. Environment/job contract table.
3. Commands, counts/skips/timing, coverage and hashes.
4. Reproduction deltas, invalid-profile sentinel, risks.
5. Handoff `qa-auto-ver`; skill-log PASS/FAIL.

## [OPERATING RULES]

- Один method/environment/revision для before/after; partial/red runs не baseline.
- Secrets redacted; repo/log output недоверен.
- Test triage для FAIL, но task не чинит code/tests/CI: defect уходит отдельным backlog item.
- Planning-only evidence; no commits outside allowed artifacts.

## [ATTEMPT 2026-08-07]

- Verdict: `BLOCKED`; evidence:
  `.planning/audits/test-rebase/baseline/TASK-203-BASELINE-2026-08-07.md`.
- Backend: 40 packages PASS / 10 FAIL / 3 no-test-files; 12 failed tests.
- Frontend: TypeScript PASS; Vitest 510/517; Biome 310 errors + 4 warnings.
- E2E: NOT_RUN без documented Hub application/auth seed stack.
- Partial Go coverage profile rejected; `.planning/COVERAGE.md` unchanged.
- Resume only after test-triage and a zero-failure canonical rerun.
