# TASK-199 — 🐍 Python framework integration conventions

> Revision: **2026-08-07** · Priority: **P1** · Status: Superseded-by-TASK-205/210/211 · Phase: **131–133**

> Historical/superseded card: non-executable provenance only; it grants no
> current command, launcher acceptance, implementation path or runtime PASS.
> Any future versioned launcher API belongs only to `testframework/runner/**`;
> `testframework/docker/**` contains immutable execution assets only. Launcher
> status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate
> implementation and independent acceptance.

## [OUTCOME]

Historical task is retired. Greenfield client/isolation work is TASK-205;
database and broker integration are TASK-210/TASK-211. Legacy file/case
comparison is not an acceptance input.

## [ACCEPTANCE]

- No executor starts TASK-199 as a separate implementation task.
- Links route to TASK-205, TASK-210 and TASK-211.
- Legacy rows are not acceptance inputs for new Python test design.
