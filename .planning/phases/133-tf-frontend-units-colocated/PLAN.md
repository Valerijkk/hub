# PLAN: Phase 133 — backend security и platform waves

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Перенести только genuine Go units security/platform в `backend/tests/**` и в
той же фазе дать API/integration/contract rows Python replacements в корневом
`testframework/**`.

## Depends on

- Phase 132 PASS (native runner/support и auth pilot).

## Success criteria

- [ ] Auth remainder, rbac, audit, ws, calls, platform, apigateway, dev, ent и
  status residual имеют accepted disposition.
- [ ] Только service-free Go units находятся в `backend/tests/**`; component
  rows там отсутствуют.
- [ ] Python candidates имеют выполненные replacements, а не только ссылки на
  будущую Phase 136.
- [ ] Hook removal отсутствует без `consumer = 0` + verifier evidence.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-210 | Auth remainder + rbac/audit/ws/calls Go-native migration и Python-candidate links | be1 | be-ver | `backend/tests/{auth,rbac,audit,ws,calls}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/133/security/**` | Phase 132 | 1–1.5д |
| TASK-211 | Platform/apigateway/dev/ent/status migration с nested mapping и DB/source-guard classification | be2 | be-ver | `backend/tests/{platform,apigateway,dev,ent,status}/**`, `.planning/audits/test-rebase/hooks.json`, `.planning/evidence/test-rebase/133/platform/**` | Phase 132 | 1–1.5д |
| TASK-227 | Security/platform API/integration/contract component parity в Python | qa-auto1 | qa-auto-ver + security-auditor | `testframework/{api,integration,contract}/{auth,rbac,audit,ws,calls,platform,apigateway,dev,ent,status}/**` | TASK-193/205/210/211 | 1–2д |

## Волны

- Wave 1: TASK-210 и TASK-211 параллельно по непересекающимся unit-доменам.
- Wave 2: TASK-227 получает frozen classification/assertion handoff.
- Wave 3: combined legacy/native/Python regression и coverage reconciliation.

## Риски

| Риск | Mitigation |
|---|---|
| Security test пропущен из-за service skip | Security/release scope fail-closed; expected skip allowlist |
| Ent/SQLite принят за unit | Любая DB fixture → TASK-227 Python integration; в backend unit lane не попадает |
| Calls/WS timing flake | Event/polling evidence; не переносить bare sleeps как норму |

## Verification

- `be-ver`: domain race, coverage, counts, source guards и hooks.
- `security-auditor`: auth/rbac/audit/ws critical negative paths не потеряны.
