# Mail 144-02C delivery schemas - immutable source receipt

## Acceptance

`STATIC_ACCEPTED / EXACT_SHA_PASS` applies only to immutable source commit `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`, subject `feat(mail): add delivery schemas`) and the exact 13-path manifest below. Tests/oracles were authored before the production schemas. Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh, and ABI/Lifecycle/Dependency/Boundary Terra/high precommit and immutable exact-SHA reviews PASS C/I/M `0/0/0`.

Provenance: Terra/medium docs sync, dispatcher-enforced/no fallback; telemetry `UNKNOWN` unless exposed.

## Exact immutable manifest

| Path | Blob |
|---|---|
| `backend/ent/schema/mailaccount.go` | `e2aa018ee24f92a2fe483df3ccd4d9c4b79f6b69` |
| `backend/ent/schema/mailattachment.go` | `2a4dcc904ef2633846368be2c7f5b1a58e5019da` |
| `backend/ent/schema/maildraft.go` | `0d0c56d90c934c1c8a959b62f62c11e6958225ea` |
| `backend/ent/schema/mailmessage.go` | `2e7b6b16fd4f3a7f15532136fc6de5a7f3b5706c` |
| `backend/ent/schema/mailobject.go` | `076f20d29f0e10b7a8b390bbb35f30480334f1e1` |
| `backend/ent/schema/mailsubmission.go` | `182dff9f62957140bc3035be58b267e4269be2a7` |
| `backend/ent/schema/organization.go` | `63a8918472dd6e82ede856594a5e3cc79a2ca012` |
| `backend/ent/schema/user.go` | `2c81a0fc37e44fe9df8318119e98a3c70967a95c` |
| `backend/tests/calendar/calendar_mutation_receipt_schema_test.go` | `9c10c67c62c31f14af6fac247782a829694fc176` |
| `backend/tests/mail/mail_account_identity_schema_test.go` | `1a79d1be0fb6bf0fdc4d2cea1708e66aeb10de02` |
| `backend/tests/mail/mail_delivery_schema_test.go` | `ca94171e20aa989b6c24b73eaaa08835555adf8d` |
| `backend/tests/mail/mail_domain_policy_schema_test.go` | `c5bf0ffe4d24947f574e0012c435ad01e899e43c` |
| `backend/tests/mail/mail_projection_schema_test.go` | `1f465e60e6503517dbfee18e705f22cd684fbd52` |

## Lifecycle and remaining boundaries

Every prior 144-02C authority-candidate, source-unaccepted, source-lifecycle-next, acceptance-sync-next, or implied-next-packet statement in the owned 02C carriers is `CONSUMED/HISTORICAL`. No next Mail code packet is selected or authorized; root must select any later packet separately.

This receipt does not establish executable evidence. Ent generation, migration, direct SQL, services, API/UI, product, phase, release, engine deletion port, physical User/GDPR intent, 02F-D and 02E remain outside this receipt; launcher is `NOT_READY` and test/build/gofmt/lint/validator/Docker/runtime evidence is `NOT_RUN`.
