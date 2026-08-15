# 144-02A1 — typed DNS static acceptance

## Immutable receipt

- Commit: `5892a90b7582ff4ac35db218428400ac9608c4e3`
- Parent: `29a5c624c0bacc2de3cc4b5a93e3e9176320b5bd`
- Tree: `c78d82ef8ea55a4656d33b82530b2fbb72582201`
- Added source: `backend/internal/mail/domain_dns.go` — blob `e00586637f209f85c52626c1ec8d96116af934b3`
- Added test: `backend/tests/mail/domain_dns_test.go` — blob `8a43140964d12cca7cdc09e34675a074ee164d7b`
- Subject: `feat(mail): add typed DNS records`

## Static acceptance

Tests were authored first. Terra/high static precommit and immutable exact-SHA review PASS C/I/M `0/0/0`. Provenance is `dispatcher-enforced`; fallback is none; telemetry is `UNKNOWN` unless exposed.

RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime are `NOT_RUN`. The task-owned launcher is `NOT_READY`.

This is narrow static source acceptance for A1 only. It grants no Ent/schema generation, migration, service, API, UI, product, phase, or release authority. In the combined A1/A2/A3 sequence, A1 is accepted; A2 and A3 remain unimplemented. Founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker; every exact packet requires its own test-first/static review/commit. This sequencing neither authorizes Ent generation, migrations, or runtime nor accepts A2. Existing blockers remain unchanged.
