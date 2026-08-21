# 143-04J — ScheduledCall projection implementation receipt

> **Status:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT14_COMMIT_REQUIRED /
> RUNTIME_NOT_RUN`. This document and the proposed receipt facts below grant no authority before
> one atomic commit of the complete exact14 documentation manifest.

## Proposed immutable source receipt

After that exact14 commit, and only then, Calendar 143-04J is `STATIC_ACCEPTED /
EXACT_SHA_PASS` for immutable source commit
`e74efa827b52f35e3d3fbe8346adf0c0121ba48d`:

- sole parent: `06a0d6e506a31fb5546ef168a5dba72d8956a984`;
- tree: `073e887a957bb65c6c06a9759e239323ab882aa7`;
- subject: `feat(calendar): add call projection policy`;
- ordered test-first exact2:
  1. `backend/tests/calendar/event_call_policy_test.go` — blob
     `28273191ddd1babec2cbe61c217cd561322a7ac0`;
  2. `backend/internal/calendar/event_call_policy.go` — blob
     `cc0218bddce25936cbfb8e987a0bb6681f979d5c`.

The external oracle was authored before production. The author route was
`gpt-5.6-terra/xhigh`, dispatcher-enforced, with no fallback; actual telemetry is `UNKNOWN` unless
exposed. Contract/Spec/Oracle, Security/Adversarial and Lifecycle/Dependency/Boundary precommit
reviews and immutable exact-SHA reviews PASS C/I/M `0/0/0`. The immutable commit is exact2-only,
and `git show --check` is clean.

## Accepted boundary after activation

Acceptance remains static and pure: the policy validates caller-owned canonical snapshots and
returns a deep-copied projection. It owns no repository, Ent schema, generated output, database,
transaction, outbox, NATS, clock, random source, filesystem, network, adapter, handler, API/UI or
raw iTIP. AQ-02 remains a migration/backfill question only.

After exact14 activation, all lower 04J authority-candidate, precommit, implementation-blocked,
source-unimplemented/uncommitted/unaccepted and exact2-next clauses are
`CONSUMED/HISTORICAL` only for 04J. No next Calendar packet is selected. 04K requires its own
docs authority and root dispatch; 04F-J/04F-K, Ent generation, migration, adapters and runtime
remain blocked. Launcher is `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime are
`NOT_RUN`.
