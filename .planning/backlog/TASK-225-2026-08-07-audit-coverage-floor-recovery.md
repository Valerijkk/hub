# TASK-225 — 📈 Audit coverage floor recovery

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **VERIFIED** · Фаза: **128**

## [OUTCOME]

Канонический zero-failure Go profile проходит overall floor ≥82% и
`internal/audit` floor ≥80% за счёт двух недублирующихся behavior-тестов, без
изменения denominator, threshold, production code или export hooks.

## [ROLE]

- Owner: `be2` — real-store retention и JetStream lifecycle tests.
- Verifier / handoff: `be-ver` — focused race run и полный coverage rerun.

## [SCOPE AND CONSTRAINTS]

- Allowed: один новый файл
  `backend/testframework/audit/coverage_gate_task225_test.go`.
- Baseline: 21,206/25,924 = 81.8%; для 82% нужно минимум +52 покрытых
  statements. Audit: 79.5%, для package floor достаточно +6.
- Do not touch: production, existing tests, hooks, coverage denominator/floor,
  CI/Compose, `pro-qa-ai/**`.
- Depends on: TASK-221 deterministic backend fixes.

## [ACCEPTANCE]

- Real Mongo/PG/S3 retention test доказывает old-delete/fresh-retain и S3
  500 fail-soft; никаких внешних credentials/network.
- Real JetStream retry test доказывает Start → Enqueue → consume → Stop через
  channel/select без `Sleep` и `t.Parallel`.
- Focused audit Docker Go 1.25 `-tags hubtest -race` green.
- Полный zero-failure profile показывает overall ≥82% и audit ≥80% с hash,
  counts и wall time; ожидаемый gain около 67 statements, запас ≥15.

## [OUTPUT]

1. Один atomic test commit.
2. Focused + full profile evidence.
3. Handoff `be-ver`; если реальный gain ниже ожидания — честный BLOCKED без
   threshold/denominator правок.

## [VERIFICATION 2026-08-07]

- Focused real-store/lifecycle tests: PASS under Docker Go 1.25 with race.
- Full canonical suite: 50 PASS / 0 FAIL / 3 no-test-files.
- Overall exact coverage: **82.0373%** (21,269 / 25,926), floor PASS.
- `internal/audit`: **85.6879%** (928 / 1,083), floor PASS.
- Profile SHA-256:
  `7C20BF7452AD9387D96F548368830282FD2A50361A0D0106E7B8069FC4C99856`.
