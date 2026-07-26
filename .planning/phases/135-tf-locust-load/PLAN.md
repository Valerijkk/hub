# Phase 135 — Locust load layer

**Status:** PLANNED  
**Migration stage:** Э6  
**Estimate:** 3–4 person-days  
**Branch (suggested):** `feat/qa/tf-locust-load`

## Goal

Реализовать Locust hot-path сценарии v1 по каталогу
[`TF-MIGRATE-E5-LOCUST.md`](../../planning-audit-2026-07/TF-MIGRATE-E5-LOCUST.md):
login, feed, search, issues + A13-001…006 baselines (board/bases/dashboard/sprint).

## Depends on

- **128** scaffold (hard)
- Soft: **134** httpx/auth helpers reuse

## Deliverables

- [ ] `tests/load/locustfile.py` + weighted `HubApiUser`
- [ ] Scenarios: `L-AUTH-01`, `L-FEED-01`, `L-SEARCH-01`, `L-ISSUE-01`, `L-BOARD-01`, `L-SPRINT-01`, `L-BASES-01`, `L-DASH-01`
- [ ] `tests/load/data/users.csv` template + seed notes
- [ ] README: smoke (10 users / 2m) vs soak
- [ ] Stubs/docs for v1.1: `L-GANTT-01`, `L-NOTIF-01`

## Out of scope

- PR-blocking Locust в CI (по умолчанию schedule/manual)
- WebRTC / WS soak

## Success criteria

- [ ] Smoke load run documented with p95 baselines
- [ ] A13 scenario tags present for board/bases/dash/sprint
- [ ] No secrets committed

## Risks

| Risk | Mitigation |
|------|------------|
| Destroying shared env | Dedicated load project + user pool |
| Dev-login only | Document load-env requirements |
