# Phase 16 — Automate (v2.6)

> **Branch:** `feat/v2.6-automate` · **PO LOCK:** `DISCUSS-ROADMAP-LOCKED.md`  
> **Depends on:** Phase 15 Trust (org, hub-mail skeleton)

## Exit criteria

PROD automation max · flaky max · issue↔message · webhooks log UI · Gantt · webcam · email marketing · triage inbox · Meilisearch indexer

## Verification loop

```
be/fe → unit tests → qa-gate slice → *-ver → fix → re-verify → commit
```

---

## Wave 1 — PROD-01 Workflow automation lite

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 Ent `automation_rule` + migration | be | arch-ver | ✓ |
| W1-02 CRUD API + RBAC | be | be-ver | ✓ |
| W1-03 Automation engine + NATS consumer | be | be-ver | ✓ |
| W1-04 FE Automation tab on workflow settings | fe | fe-ver | ✓ |
| W1-05 Unit tests + E2E smoke | qa | qa-ver | ✓ |

## Wave 2 — WEBHOOK-01 delivery log UI

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W2-01 FE webhook deliveries viewer (API exists) | fe | fe-ver | ✓ |

## Wave 3 — PROD-02 Flaky detection max

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W3-01 Enhance flaky algorithm + FE report | be/fe | qa-ver | ✓ |

## Wave 4 — PROD-03/04 Issue↔message link

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W4-01 `issue_message_link` schema + sync | be | sa-ver | ✓ |
| W4-02 FE thread header links | fe | fe-ver | ✓ |

## Wave 5 — TRIAGE-01 Triage inbox

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W5-01 Triage view + saved filter preset | fe | fe-ver | ✓ |

## Wave 6 — GANTT-01

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W6-01 Gantt chart component + dependencies | fe | arch-ver | ✓ |

## Wave 7 — CALL-01 Webcam

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W7-01 LiveKit camera track toggle | fe | fe-ver | ✓ |

## Wave 8 — NOTIF-09 Email marketing lite

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W8-01 Admin broadcast + hub-mail bridge | be/fe | security-auditor | ✓ |

## Wave 9 — SEARCH-01 Meilisearch

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W9-01 Indexer worker + compose | devops | be-ver | ✓ |

---

*Spike: `.planning/PROD-01-SPIKE.md`*

**Phase 16: SHIPPED (code complete, uncommitted) — 2026-06-22**
