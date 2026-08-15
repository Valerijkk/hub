# 144-01B — Deterministic engine fake static acceptance

## Immutable receipt

- Commit: `0c064363b5ccb657a7c444802e94075c4ca743e5`
- Parent: `48a7273d100c07788def88e2f1b8d69c7173580f`
- Tree: `9c568a0c58c6bf7b8c3171cfd934760b839db623`
- Subject: `feat(mail): add deterministic engine fake`
- Status: `STATIC_ACCEPTED` / `EXACT_SHA_PASS`

## Exact accepted paths

| Path | Blob | Evidence |
|---|---|---|
| `backend/tests/mail/mailbox_fake_test.go` | `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f` | authored first; static oracle only |
| `backend/internal/mail/mailbox_fake.go` | `12bcd921668e406f07a98f7f3b0618ad8d44e81d` | deterministic fake only |

## Review and retained boundaries

Terra/xhigh authored after the test. Independent Spec Terra/high, Security Terra/xhigh and ABI Terra/high precommit reviews, plus Terra/high immutable exact-SHA review, PASS C/I/M `0/0/0`. Provenance is `dispatcher-enforced`, fallback none, telemetry `UNKNOWN` unless exposed. Former 01B Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`. Launcher `NOT_READY`; RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. No 144-02E, Ent/schema, generation, migration, service, API, UI, product, phase or release authority follows; Phase 142/148, Rule/Sieve/Phase145 and all other blockers remain.
