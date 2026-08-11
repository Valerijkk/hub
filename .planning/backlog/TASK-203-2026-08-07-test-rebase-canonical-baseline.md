# TASK-203 — 📏 Canonical green baseline

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **CLOSED — discovery recorded; migration gate superseded** · Фаза: **128**

## [OUTCOME]

Зафиксирован discovery baseline legacy suites и Go coverage. Founder cutover
отменил сравнение migration waves и row parity: данные остаются только
историческим снимком до удаления non-unit legacy.

## [ROLE]

- Owner: `devops-engineer` — historical environment contract and baseline evidence.
- Verifier / handoff: `qa-auto-ver` — независимый rerun/reconciliation и проверка валидности profiles.

## [SCOPE AND CONSTRAINTS]

- This closed card is historical, non-executable provenance. Nothing below is
  a current command, launcher acceptance, implementation path or runtime PASS.
  Any future versioned launcher API belongs only to `testframework/runner/**`;
  `testframework/docker/**` contains immutable execution assets only. Launcher
  status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate
  implementation and independent acceptance.
- Allowed: `.planning/audits/test-rebase/**`, `.planning/COVERAGE.md` только для dated baseline evidence; временные profiles вне repo.
- Read-only: CI/compose/Makefiles, runtime manifests и legacy suites.
- Do not touch: CI commands, tests/product code, thresholds, legacy paths, STATE/ROADMAP.
- Historical Go provenance: Go 1.25, required Postgres/Mongo/Redis/NATS and the
  recorded race/coverage flags; this is not a current runnable entrypoint.
- Separate baselines: Go tests/coverage, Vitest, TS smoke/full. Counts включают collected/executed/passed/failed/skipped, wall time и artifacts.
- Depends on: TASK-202.

## [REFERENCES]

- `.planning/codebase/{STACK,TESTING,INTEGRATIONS,CONCERNS}.md`
- `.github/workflows/ci.yml`, `deployments/docker-compose.test.yml`
- `.cursor/playbooks/testing.md`, current `.planning/COVERAGE.md`

## [ACCEPTANCE]

- Historical inner command provenance is recorded in the dated evidence only;
  it is non-executable here and must not be run on the host or used as fallback.

- Historical service-backed invocation provenance is recorded with HEAD, OS/runtime/container versions, env variable names (без values), exact package/spec scope, counts/skips, wall time and profile/report hashes.
- Go overall and security package coverage recorded only from zero-failure run.
  Former `>=82%` movement blocker is superseded by the founder-approved
  greenfield cutover and remains historical evidence only.
- Repeat run or verifier run reproduces counts and coverage within defined tolerance; differences explained.
- Failing sentinel: malformed/empty coverage profile and empty collection are rejected by baseline validation. Production mutation N/A: task measures existing behavior and changes no code.
- Regression: N/A — evidence-only baseline; no tracked code/test/CI changes except planning artifacts.
- Stop when: Python/Docker/Go 1.25 or required services unavailable; status BLOCKED with exact missing prerequisite, never publish partial coverage as baseline.

## [OUTPUT SCHEMA]

1. Historical PASS/BLOCKED verdict and summary.
2. Environment/job contract table.
3. Historical invocation provenance, counts/skips/timing, coverage and hashes.
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

## [REMEDIATION RESULT 2026-08-07]

- Backend full Docker/race suite: **50 PASS / 0 FAIL / 3 no-test-files**.
- Exact merged coverage: **82.0373%** (21,269 / 25,926), overall floor PASS.
- Security floors: **6/6 PASS**; `audit` = **85.6879%**.
- Frontend: TypeScript PASS, Biome 370 files / 0 diagnostics; Vitest inventory: **180 files / 519 literal cases**.
- Docker E2E smoke: **3/3 PASS**. Full with retries disabled: **48 PASS / 5
  expected skip / 0 fail**.
- Final resource-limited Docker legacy Vitest diagnostic: **510/519 PASS, 9
  FAIL**. The result is frozen evidence, not a remediation queue.
- Founder decision supersedes the former baseline/parity gate: only genuine
  native units are retained; legacy non-unit suites are deleted by exact
  manifest; new non-unit coverage is greenfield Python from requirements and
  risks. The historical 82% target does not block migration.
