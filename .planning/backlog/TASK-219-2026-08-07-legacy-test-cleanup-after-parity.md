# TASK-219 — 🧹 Legacy cleanup after parity

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **141**

## [OUTCOME]

Только после proven parity legacy test paths, obsolete linker/manifests и orphan hooks удалены по rollback manifest, а canonical replacement pipeline остаётся green.

## [ROLE]

- Owner: `devops-engineer` — guarded cleanup, entrypoint cutover and rollback artifact.
- Verifier / handoff: `code-reviewer` — deletion audit, consumer search, CI/rollback review.

## [SCOPE AND CONSTRAINTS]

- Allowed: deletion only for ledger rows marked cleanup-approved; CI/scripts/config/docs directly tied to removed paths; rollback manifest/report.
- Do not touch: product behavior, unapproved rows, coverage threshold/method, protected zone.
- Preconditions: TASK-218 PASS, real PR dual-run green, all rows parity-proven, unexpected skips/unmapped = 0, TASK-203 coverage contract satisfied, owner approval for destructive cleanup.
- Hooks removed one symbol/file at a time only after zero consumers across legacy/target; release build verified.
- Depends on: TASK-193–218 and TASK-221–230 all verified.

## [REFERENCES]

- TASK-202 ledger/hash; TASK-203/204; TASK-218 gate evidence
- `.planning/codebase/{STRUCTURE,ARCHITECTURE,CONCERNS}.md`, git/CI rules

## [ACCEPTANCE]

- Completion command: `pwsh -NoProfile -File scripts/test-entrypoints.ps1 target`; exit code 0 и заявленные counts обязательны.

- Rollback manifest lists every deleted path/hash/replacement/evidence; deletion set equals approved ledger set, extra/missing = 0.
- Removed: approved legacy Go/FE/TS-E2E assets, obsolete junction/linker/manifests only when zero consumers; retained: any unresolved row/support.
- Canonical target Go/FE/Python smoke/full/integration, coverage, static gates and release build green after each cleanup batch.
- Failing sentinel: final target lanes still RED on temporary sentinel and GREEN after restore; orphan/stale-path searches return zero.
- Regression: real PR target pipeline green; rollback rehearsal restores a sampled batch from manifest without ambiguity.
- Stop when: any parity/coverage/approval/consumer proof missing; perform no deletion and return BLOCKED.

## [OUTPUT SCHEMA]

1. Verdict. 2. Deleted/retained/rollback manifest summary. 3. Consumer/static/test/coverage/sentinel evidence. 4. Approval/rollback/risks. 5. Handoff `code-reviewer`; skill-log.

## [OPERATING RULES]

- Destructive action only after exact target resolution and explicit owner approval; batch small and recoverable.
- No broad glob deletion or hook cleanup; verify after each batch.
- Docs-sync all removed entrypoints; TASK-linked atomic commits, no AI attribution.
