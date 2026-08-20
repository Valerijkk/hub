# Implementation 144-02B — projection schemas

> **Status:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS`
> **Source commit:** `947e2a27650ff1c6db60af9ce13bb33b6bf19cb0`
> **Parent:** `856edb25cf40ad436b114cc928923203c195d15c`
> **Tree:** `10c0dcac9953ec53fd1631a095a44052e8cec9c9`
> **Subject:** `feat(mail): add projection schemas`

| Path | Blob |
|---|---|
| `backend/ent/schema/calendarevent.go` | `821e175a91a30b8485b8fed0379911d2c8efe245` |
| `backend/ent/schema/mailaccount.go` | `b08171899d21d35a7b08fa6e1f06bd6a703b83c0` |
| `backend/ent/schema/mailfolder.go` | `97788cedbb6786752b5b52f2bf486da08bb2ed7f` |
| `backend/ent/schema/maillabel.go` | `9d7789b55c7d7958c257084b6c8f4185a217f35a` |
| `backend/ent/schema/mailmessage.go` | `aa6da68b2f6b5258c88ec4b4ea7f2295a2e05c63` |
| `backend/ent/schema/mailmessagefolder.go` | `ba0999642d4d37d48ae0f9994871263cb58b0ae0` |
| `backend/ent/schema/mailmessagelabel.go` | `352b097bd9496d8020d544c0629a1d3b0c8acb28` |
| `backend/ent/schema/mailmutationreceipt.go` | `d4315d40539b339b34a71fe1eabbcc86778b63ef` |
| `backend/ent/schema/mailthread.go` | `83f9299e4228646a7541bb297360af15bf63d2d5` |
| `backend/tests/calendar/calendar_event_schema_test.go` | `02e746fa5b42b2dd15bb1b4338d9c1a8e4d5cf26` |
| `backend/tests/mail/mail_account_identity_schema_test.go` | `737934a18e30d4c6842e932c69814e4f4554d72e` |
| `backend/tests/mail/mail_mutation_receipt_schema_test.go` | `21f233cb4288d4589b3e4a33a9fdaa6ad056a1b9` |
| `backend/tests/mail/mail_projection_schema_test.go` | `72c5bd3c8e40040f50257de79844be085d1f61c7` |

The projection-schema oracle was authored before the new schemas. It records the accepted
PERSISTENCE 02B tables, checks, indexes, canonical JSON preflight and only the permitted
receipt inverse: `MailMutationReceipt.folders` for `MailFolder.control_receipt_id -> receipt
RESTRICT`. No actor/message receipt edge or rule-target database relation was added.

Contract/high, Security/xhigh and Lifecycle/high precommit plus immutable exact-SHA reviews
all PASS C/I/M `0/0/0`. Authoring was `gpt-5.6-terra`/`xhigh`, dispatcher-enforced with no
fallback; actual telemetry is `UNKNOWN` unless exposed. This is static source acceptance only:
launcher is `NOT_READY`; tests, build, generation, migrations and runtime are `NOT_RUN`.

All preceding 02B `SOURCE_AUTHORITY_ACCEPTED`, source-unaccepted and 02C-waits-02B-source
instructions are `CONSUMED/HISTORICAL`. Root may separately author and adjudicate a narrow
02C docs/source-authority bridge; 02C source is not automatically authorized. 02F-D/02E,
Ent generation, migrations, services, API, UI, product, phase and release work remain blocked.
