# PLAN: Phase 133 — backend security и platform waves

**Статус:** PLANNED · **Оценка:** 2–3 дня

Канон: [`06-TEST-ARCHITECTURE-REBASE.md`](../../active/august-2026/06-TEST-ARCHITECTURE-REBASE.md).

## Цель (goal)

Перенести Go-native security/platform tests в `backend/tests/**`, а API/
integration candidates оставить исполняемыми в legacy до Python parity.

## Depends on

- Phase 132 PASS (native runner/support и auth pilot).

## Success criteria

- [ ] Auth remainder, rbac, audit, ws, calls, platform, apigateway, dev, ent и
  status residual имеют accepted disposition.
- [ ] Go-native replacements green с race и frozen coverage method.
- [ ] Python candidates не удалены и связаны с Phase 136 inventory rows.
- [ ] Hook removal отсутствует без `consumer = 0` + verifier evidence.

## Задачи (атомарные)

| Task ID | Outcome | Owner | Verifier | Files | Depends | Estimate |
|---|---|---|---|---|---|---:|
| TASK-210 | Auth remainder + rbac/audit/ws/calls Go-native migration и Python-candidate links | be1 | be-ver | `backend/tests/{auth,rbac,audit,ws,calls}/**`, `.planning/audits/test-rebase/inventory.json`, `.planning/evidence/test-rebase/133/security/**` | Phase 132 | 1–1.5д |
| TASK-211 | Platform/apigateway/dev/ent/status migration с nested mapping и DB/source-guard classification | be2 | be-ver | `backend/tests/{platform,apigateway,dev,ent,status}/**`, `.planning/audits/test-rebase/hooks.json`, `.planning/evidence/test-rebase/133/platform/**` | Phase 132 | 1–1.5д |

## Волны

- Wave 1: TASK-210 и TASK-211 параллельно по непересекающимся доменам.
- Wave 2: combined legacy/native regression и coverage reconciliation.

## Риски

| Риск | Mitigation |
|---|---|
| Security test пропущен из-за service skip | Security/release scope fail-closed; expected skip allowlist |
| Ent/SQLite принят за real DB integration | Engine-sensitive rows → Phase 136, исходник остаётся live |
| Calls/WS timing flake | Event/polling evidence; не переносить bare sleeps как норму |

## Verification

- `be-ver`: domain race, coverage, counts, source guards и hooks.
- `security-auditor`: auth/rbac/audit/ws critical negative paths не потеряны.
