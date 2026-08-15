# 144-02A3 — account and identity schema receipt

## Static acceptance

`e9a1764b001e28074e4074a07ca8ac982e5345ba` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`.

- Parent: `47f8987b269e1625022071e0de95a76da0d2f00a`
- Tree: `c0b59b2ba1168076a4cc787cacb99fe5081cd301`
- Subject: `feat(mail): add account and identity schemas`

## Exact accepted paths and blobs

| Path | Blob | Role |
|---|---|---|
| `backend/ent/schema/mailaccount.go` | `30e68ff52171e643e1361f3d6acfd2310289d8e1` | Account schema |
| `backend/ent/schema/maildomain.go` | `16f5d20600d1359ba737417428d594ac73468360` | Required account inverse edge |
| `backend/ent/schema/mailidentity.go` | `1b06e04c9235e5c891a601261a2e400068155065` | Identity schema |
| `backend/ent/schema/organization.go` | `df85d2e2f61bcc8b62235f0557f774fb06eca1e3` | Required account inverse edge |
| `backend/ent/schema/user.go` | `5594ca1c68b8f3c726ab90a77dcc26b7e7713e0c` | Required account inverse edge |
| `backend/tests/mail/mail_account_identity_schema_test.go` | `0939618d620c7f9b503df50e4aff04fbf0b3a85d` | Authored-first A3 metadata oracle |
| `backend/tests/mail/mail_domain_policy_schema_test.go` | `70b0c56f20043aeb8b4b719ca4ae6f3405cb5221` | A2 compatibility oracle |

## Review and limits

Tests were authored first. Terra/xhigh authored the packet; Terra/high precommit and immutable exact-SHA reviews PASS with C/I/M `0/0/0`. Provenance is `dispatcher-enforced`, fallback is absent, and actual telemetry is `UNKNOWN` unless exposed.

This is static source acceptance only. Launcher is `NOT_READY`; RED/GREEN, tests, build, gofmt, lint, validators, Docker, Ent generation, migration and runtime are `NOT_RUN`. It grants no generation, migration, runtime, service, API, UI, product, phase or release authority.

## Sequencing

A3-unimplemented and A3-next clauses are `CONSUMED/HISTORICAL`. 02B remains blocked pending its exact 02F bridge/dispatch. The former founder root-override selected independent 02D exact nine paths; that next-route clause is `CONSUMED/HISTORICAL` because 02D is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `2242a4e7846944498eb5ed701ab8f47c986bd804`. 02F source remains unauthorized; 02E/generation/migration/runtime remain blocked; root separately selects the next Mail packet. A1/A2 facts and all other blockers remain unchanged.
