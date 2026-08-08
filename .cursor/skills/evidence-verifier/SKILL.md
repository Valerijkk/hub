---
name: evidence-verifier
description: Use when a non-trivial Hub change, completion claim, phase, or handoff needs an independent read-only verdict against a specification and executable evidence.
---

# evidence-verifier

Проведи conclusion-blind, spec-aware и read-only проверку. Сначала сформируй
собственный результат по контракту и evidence; вывод автора прочитай только
после фиксации первичного verdict. Не создавай патчи.

## Вход

- exact task contract: objective, user outcome, acceptance, non-goals,
  constraints, risk, owned paths, required evidence, approval и stop conditions;
- `BASE..HEAD` diff, full source revisions, worktree state и текущий код;
- trusted expected result/oracle и применимые спецификации;
- raw executable checks/logs с командой, средой, временем и outcome;
- runtime provenance: requested model/effort и observed actual telemetry.

Если actual model/effort не наблюдаемы, запиши `UNKNOWN`. Не выводи их из
названия роли, launcher config или самоотчёта агента.

## Проверка

1. Сверь scope и diff с task contract на точной revision.
2. Выбери сильнейший применимый oracle: trusted invariant/contract → focused
   executable check → integration/recovery exercise → independent expert review.
   Self-critique используй только для гипотез.
3. Воспроизведи требуемое evidence в разрешённой среде либо поставь `NOT_RUN`.
4. Проверь негативные, boundary, authz и recovery-сценарии по риску.
5. Назначь каждому criterion статус:
   - `PASS` — oracle подтверждён свежим evidence на указанной revision;
   - `FAIL` — criterion или инвариант опровергнут;
   - `NOT_RUN` — требуемая проверка не выполнялась или provenance недостаточен.
6. Итог: любой required `FAIL` → `FAIL`; без FAIL, но с required `NOT_RUN` →
   `NOT_RUN`; только все required PASS → `PASS`.
7. При remediation повторно проверяй только finding scope и затронутую
   регрессионную границу, не больше bounds из task contract. CRITICAL/HIGH
   нельзя park, принять по обещанию или вынести из PASS.

## Отчёт

```text
Verdict: PASS | FAIL | NOT_RUN
Task/revision: <ID> @ <full revision>; worktree: <state>
Oracle: <точный expected result и источник>
Runtime: requested=<...>; actual=<telemetry | UNKNOWN>; env/image=<digest | UNKNOWN>
Scope: <проверено> | Excluded: <не проверено>
Criteria:
- <ID> — PASS|FAIL|NOT_RUN — <evidence>
Findings: <severity, факт, влияние>
Remediation: <не более 3 действий в finding scope: owner -> action -> oracle>
Residual risk: <остаток или none>
```

Не ставь `verified: true` за автора и не заменяй missing oracle количеством
reviewers. Передай verdict в
[`goal-verify`](../gsd-ru/goal-verify/SKILL.md) или
[`handoff`](../gsd-ru/handoff/SKILL.md).
