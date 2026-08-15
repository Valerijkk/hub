# 144-02D — user configuration schema receipt

## Static acceptance

`2242a4e7846944498eb5ed701ab8f47c986bd804` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`.

- Parent: `e9a1764b001e28074e4074a07ca8ac982e5345ba`
- Tree: `8129303399e627ca5cedac3e70648955c74cfed9`
- Subject: `feat(mail): add user configuration schemas`

## Exact accepted paths and blobs

| Path | Blob | Role |
|---|---|---|
| `backend/ent/schema/mailaccount.go` | `55b9ccfb4e19a190fc7c083a61350ce7ba8c7e5b` | Account inverse-edge update |
| `backend/ent/schema/mailidentity.go` | `6b8614c651ee70a70bdb46c00456362c8edad796` | Identity inverse-edge update |
| `backend/ent/schema/mailrule.go` | `a826916965ab4875a9c61da83fa0a68a8d5d3d4d` | Rule schema |
| `backend/ent/schema/mailsignature.go` | `fba258559fe6a8557b045fa10757330be29b5c08` | Signature schema |
| `backend/ent/schema/mailusersettings.go` | `7643892d0af8934892d23713e4d35b07b0bc455c` | User-settings schema |
| `backend/ent/schema/recentcontact.go` | `b2216313dfdf492315a377b3567fdc4f71d05a0` | Recent-contact schema |
| `backend/ent/schema/user.go` | `b595855cbe9cd39f522ef8263eb72ea4db4f9b55` | User inverse-edge update |
| `backend/tests/mail/mail_account_identity_schema_test.go` | `74599da63c180942e092a909ee53eb0ca6b7e51f` | A3 compatibility oracle |
| `backend/tests/mail/mail_user_configuration_schema_test.go` | `b94c686c35e3a5cca2d7838424f7e255012437d8` | Authored-first 02D oracle |

## Review, scope and remaining gates

Tests were authored first. Terra/xhigh authored the packet; Terra/high precommit and immutable exact-SHA reviews PASS with C/I/M `0/0/0`. Provenance is `dispatcher-enforced`, fallback is absent, and actual telemetry is `UNKNOWN` unless exposed.

This is static source acceptance only. Launcher is `NOT_READY`; RED/GREEN, tests, build, gofmt, lint, validators, Docker, Ent generation, migration and runtime are `NOT_RUN`. It grants no generation, migration, runtime, service, API, UI, product, phase or release authority.

The former 02D-next, 02D-unimplemented and 02D-source-blocked clauses are `CONSUMED/HISTORICAL` only. Earlier 02D docs/readiness receipts remain authoritative. 02B remains blocked on its exact 02F bridge; 02F source is not authorized. There is no concrete next Mail packet: root must select one separately.
