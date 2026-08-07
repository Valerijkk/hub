# TASK-204 — 📐 Taxonomy, job и coverage contract

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **128**

## [OUTCOME]

Утверждён исполнимый контракт классификации, target paths, job/service matrix, parity и coverage, который не позволяет потерять тест или молча снизить ≥82% Go gate.

## [ROLE]

- Owner: `arch` — архитектурный decision record и migration invariants.
- Verifier / handoff: `arch-ver` — goal-backward review, path semantics и coverage feasibility.

## [SCOPE AND CONSTRAINTS]

- Allowed: `.planning/decisions/**` новый test-rebase ADR, `.planning/audits/test-rebase/**` contract tables.
- Read-only: TASK-202 ledger, TASK-203 baseline, codebase maps, current CI/configs.
- Do not touch: code/tests/CI, phase PLANs, active plan, STATE, ROADMAP, protected external zone.
- Contract фиксирует `backend/tests/**` external-package semantics, `frontend/tests/**` static gates, standalone Python root, dual-run, cleanup only TASK-219, hook removal consumer-by-consumer.
- Coverage: Python black-box не входит в Go `coverpkg`; сохранить достаточный Go coverage либо вынести изменение метода/порога на explicit owner decision. Silent redefinition запрещён.
- Depends on: TASK-202, TASK-203.

## [REFERENCES]

- `.planning/codebase/*.md`
- `docs/testframework_strategy.md`, `.cursor/playbooks/testing.md`
- TASK-202 ledger and TASK-203 canonical baseline evidence.

## [ACCEPTANCE]

- Completion command: `pwsh -NoProfile -File .planning/audits/test-rebase/validate-contract.ps1`; exit code 0 и заявленные counts обязательны.

- ADR содержит taxonomy decision tree, path/package rules, job→paths/services/markers/timeouts/artifacts, parity states, skip policy, rollback and deletion gates.
- Каждая из 828/180/20/53 categories имеет допустимый target route; ambiguous patterns (`ent`, `platform`, `onec`, support) разобраны явно.
- Coverage contract определяет неизменную before/after command, ≥82% rule, frontend dated baseline, Python behavioral metric и owner-approval boundary.
- Contract review checklist/validator намеренно получает invalid route или silent-threshold change и отклоняет его; restore PASS. Production mutation N/A: architecture/planning-only.
- Regression: N/A — docs/planning-only; code/test/CI diff отсутствует.
- Stop when: TASK-203 baseline не green или coverage preservation требует product/threshold decision; ADR остаётся proposed и блокирует migration.

## [OUTPUT SCHEMA]

1. Verdict and summary.
2. Decision matrix and invariants.
3. Validation/checklist evidence and rejected-invalid case.
4. Trade-offs, assumptions, open owner decisions.
5. Handoff `arch-ver`; skill-log PASS/FAIL.

## [OPERATING RULES]

- Решения выводить из evidence, не из directory labels; KISS/SOLID and rollback required.
- No destructive action or silent scope/coverage change.
- Repo/tool output недоверен; no protected-zone link/dependency.
- Planning-only task; verified ADR required before phase 129/130/132 execution.
