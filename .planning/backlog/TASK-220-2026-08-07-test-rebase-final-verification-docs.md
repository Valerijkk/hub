# TASK-220 — ✅ Final verification, docs и coverage

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **141**

## [OUTCOME]

Test-rebase имеет независимый финальный PASS: target topology/CI/coverage/traceability/docs согласованы, legacy loss отсутствует и handoff готов к приёмке team-lead.

## [ROLE]

- Owner: `qa-auto-ver` — финальная независимая верификация и docs/coverage reconciliation.
- Verifier / handoff: `team-lead` — итоговая goal-backward приёмка фазы 141 и программы.

## [SCOPE AND CONSTRAINTS]

- Allowed: final verification report, `.planning/COVERAGE.md`, `.planning/STATE.md`, backlog statuses/index, `docs/{TESTING.md,test_plan.md,testframework_strategy.md,TROUBLESHOOTING.md}`, `testframework/{README.md,PIPELINE.md}` when evidence requires sync.
- Read-only: code/tests/CI and all task evidence. Do not fix implementation; findings become reopen/new TASK.
- Verify against TASK-202 ledger/hash and TASK-203 method, not stale AGENTS/docs counters.
- Depends on: TASK-219 and verifier/review evidence TASK-193–218.

## [REFERENCES]

- TASK-202–219 artifacts; `.planning/codebase/*.md`
- `.cursor/playbooks/{testing,verification,team-workflow}.md`, docs-sync/test-triage rules

## [ACCEPTANCE]

- Completion command: `pwsh -NoProfile -File scripts/test-entrypoints.ps1 target`; exit code 0 и заявленные counts обязательны.

- Recount and ledger reconciliation prove all source rows accounted for, target paths canonical, orphan/duplicate/unmapped/unexpected skip = 0.
- Canonical Go race/coverage, FE type/lint/test/coverage, Python API/integration/smoke/full and CI gates pass with recorded revision/counts/timing/artifacts; ≥82% Go rule evaluated by TASK-203 method.
- Sentinel evidence from TASK-218 remains valid; sample rerun confirms failure sensitivity. Production mutation N/A: independent verification/docs-only; no implementation mutation permitted here.
- Regression: independently rerun all canonical target lanes plus release build; any red/skip/count/coverage drift blocks final PASS.
- Docs/STATE/COVERAGE/README state measured current facts, commands and known limitations consistently; no protected-template dependency.
- Final report lists PASS/PARTIAL/FAIL per gate and rollback/residual risks; `team-lead` receives exact acceptance packet.
- Stop when: any required evidence missing/stale/red; do not mark tasks/program closed, return actionable findings.

## [OUTPUT SCHEMA]

1. Final verdict and 2–5 sentence executive summary.
2. Gate matrix with evidence paths/outcomes.
3. Counts/coverage/skip/traceability reconciliation.
4. Docs-sync diff and residual risks/rollback.
5. Handoff `team-lead`; skill-log PASS/FAIL.

## [OPERATING RULES]

- Independently reproduce critical evidence; do not accept self-reported green.
- No code/test/CI fixes in verifier task; reopen owner task on defect.
- Docs-only changes use evidence and same commands; no stale metric blending.
- No commit/merge/push without normal policy and owner approval; no AI attribution.
