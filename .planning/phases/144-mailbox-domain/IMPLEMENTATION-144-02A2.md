# 144-02A2 — domain and policy schema receipt

## Static acceptance

`65c91b4824e8a1948fc97c149305aa7f009de061` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`.

- Parent: `5df6b47678364c438ade579252209c8fd238c43c`
- Tree: `08583554939f597443566093f30b6ec7aa9373db`
- Subject: `feat(mail): add domain and policy schemas`

## Exact accepted paths and blobs

| Path | Blob | Role |
|---|---|---|
| `backend/ent/schema/maildomain.go` | `4de2ec0882d94db595aaabf8d4117566118ee472` | Domain schema |
| `backend/ent/schema/mailboxpolicy.go` | `a70e05122ffa126e9754d029ca00493faae2aeef` | Policy schema |
| `backend/ent/schema/organization.go` | `ef6e2e65430a1397b0f563f1f361ecd89660a39c` | Required additive inverse edges |
| `backend/tests/mail/mail_domain_policy_schema_test.go` | `65de04719a60a38d7fc456553e092bb106456957` | Authored-first metadata oracle |

## Review and limits

Tests were authored first. Terra/xhigh authored the packet; Terra/high precommit and immutable exact-SHA reviews PASS with C/I/M `0/0/0`. Provenance is `dispatcher-enforced`, fallback is absent, and actual telemetry is `UNKNOWN` unless exposed.

This is static source acceptance only. Launcher is `NOT_READY`; RED/GREEN, tests, build, gofmt, lint, validators, Docker, generation, migration and runtime are `NOT_RUN`. It grants no generation, migration, runtime, service, API, UI, product, phase or release authority.

## Sequencing

A2-unimplemented clauses are `CONSUMED/HISTORICAL`. A3 account/identity remains unimplemented; its founder-sequenced exact source boundary is `backend/ent/schema/mailaccount.go`, `backend/ent/schema/mailidentity.go`, and only their required narrow `Organization` inverse edges. `144-02E` and every other blocker remain unchanged.
