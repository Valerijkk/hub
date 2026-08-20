# Implementation 144-02F-R — mutation receipt schemas

> **Status:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS`
> **Source commit:** `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`
> **Parent:** `1792cdc381eeb4db0ab1ede5b512ba1222eb90b2`
> **Tree:** `2cadf80c9ae0af99b5c8027c1d78c3ec55af7071`

| Path | Blob |
|---|---|
| `backend/ent/schema/mailaccount.go` | `2c8135980c7314b2eb458b17a53564803dc91d5d` |
| `backend/ent/schema/mailmutationreceipt.go` | `d772e453539f30c61de90fc7d7abc7f25dbc95a5` |
| `backend/ent/schema/mailmutationreceiptmember.go` | `5ee6d467470c488f0c3c4def24e57fce56cd5cae` |
| `backend/ent/schema/mailmutationreceipttarget.go` | `e7d06a468a3a1916a5f99f69512e14bb20c5ae44` |
| `backend/ent/schema/organization.go` | `c7b30442179e54d42b4ec07e2a6230e6ca08d1f6` |
| `backend/tests/mail/mail_account_identity_schema_test.go` | `a7743e3e1c001f95174ca28197f6c490269716ad` |
| `backend/tests/mail/mail_domain_policy_schema_test.go` | `93a8fcff4e81c6123545f02acee3ac88448094c0` |
| `backend/tests/mail/mail_mutation_receipt_schema_test.go` | `9f12b7f9ce73604ce891b0d7ae0b37fd63a5b03e` |

Tests first; Contract, Security and Lifecycle exact reviews PASS C/I/M `0/0/0`; `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Authority-only/root-dispatch/source-unaccepted and 02B-separate-docs-authority/02C-predecessor-wait routes are `CONSUMED/HISTORICAL`. Immutable 02C candidate `e28603513f29f814fbcdd851e7b2c4816dd6c8af` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS`; root may dispatch only the frozen test-first exact13 source/test packet while its source remains unaccepted. Static only: launcher `NOT_READY`, executable/generation/migration/runtime `NOT_RUN`; 02F-D remains `BLOCKED` / `NOT_AUTHORIZED`.
