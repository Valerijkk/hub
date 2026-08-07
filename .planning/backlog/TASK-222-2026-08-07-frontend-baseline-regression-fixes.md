# TASK-222 — 🧪 Frontend baseline regression fixes

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: **На проверке** · Фаза: **128**

## [OUTCOME]

Семь Vitest regressions TASK-203 исправлены по актуальным runtime/type/i18n и
design-token contracts: шесть stale fixtures и один реальный semantic-token
дефект компонента.

## [ROLE]

- Owner: `fe1` — fixtures и pipeline semantic tokens.
- Verifier / handoff: `fe-ver` — focused/full Vitest, TypeScript и visual token
  contract review.

## [SCOPE AND CONSTRAINTS]

- Allowed: три test files и `frontend/src/modules/tracker/pipeline-status.tsx`
  из frontend diagnosis handoff.
- Do not touch: global Biome debt/config, другие production/tests, CI/Compose.
- Depends on: TASK-202 и красная попытка TASK-203 от 2026-08-07.

## [ACCEPTANCE]

- Calls fixtures являются typed `PersonalData`; сценарий без `channel_id`
  сохраняет остальные обязательные поля и доказывает `joinCall(call.id)`.
- Health fixtures являются typed `AdminStats`; меняется только
  `service_health`.
- ErrorBoundary resetKey проверяется locale-independent assertion.
- Pipeline status не содержит raw green/red/yellow/gray palette и сохраняет
  URL sanitization.
- Четыре focused suites, `tsc --noEmit` и все 517+ Vitest tests green.

## [OUTPUT]

1. Атомарные commits отдельно для fixtures и semantic tokens.
2. Focused/full test counts и TypeScript evidence.
3. Handoff `fe-ver`; global lint честно остаётся TASK-223.
