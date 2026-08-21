# 144-04G — mailbox sync coordinator immutable source receipt

> **Coherence gate:** `COHERENCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-self-activating dirty exact12 docs-only coherence candidate grants no current static, source, runtime or downstream authority. It records the immutable source and prior immutable acceptance receipt below. Only an atomic manifest-equal exact12 coherence commit plus immutable coherence Contract/Security/Lifecycle triple PASS can make 144-04G `STATIC_ACCEPTED / EXACT_SHA_PASS`; then former 04G authority/acceptance/source-unimplemented/current-next routes become `CONSUMED/HISTORICAL`. Historical coherence selects no next Mail packet and lifts no downstream blocker.

## Immutable source receipt

Immutable source commit `3031c53d399accc84fe10cd0c5421a85c089ac7c` is not current static acceptance until the coherence gate activates:

- parent: `085a81aed317be364811af9507912002a7dae457`
- tree: `b146acb45bf60f2f468a926caa3546c0b8f79893`
- subject: `feat(mail): add mailbox sync coordinator`
- tests first: `backend/tests/mail/mailbox_sync_service_test.go` — `4b8de26d16a749536b3fc3207999d06dc20ab379`
- repository: `backend/internal/mail/mailbox_repository.go` — `3b16c7a0dcf4fc89d3b773cfb8d3f75cee6219d4`
- service: `backend/internal/mail/mailbox_sync_service.go` — `0ed93f7247b428e8af8de2c6bd0f82bc5b9bb441`

Contract, Security and Lifecycle review each recorded precommit and immutable `EXACT_SHA_PASS`, with `C/I/M 0/0/0`. The author provenance is `gpt-5.6-terra` / `xhigh`, dispatcher-enforced with no fallback; telemetry is `UNKNOWN` because it is not exposed.

This receipt is static source-only evidence for the frozen coordinator ABI, state transitions and contract oracles. It is not runtime evidence and does not authorize an Ent or SQL schema, concrete adapter, generated binding, migration, public OpenAPI operation, or any execution. The launcher remains `NOT_READY`; executables and runtime remain `NOT_RUN`.

## Immutable prior acceptance receipt

The prior acceptance receipt is immutable commit `2a2975056c3fb8001d88863a041310271282989d`:

- parent: `6c92046b374a7be0b6964fea335016ed6021e91a`
- tree: `3445df1500adda42d43aa557c370c24308857512`
- subject: `docs(mail): record sync coordinator acceptance`

Its exact12 manifest is:

| Order | Path | Blob |
|---:|---|---|
| 1 | `.planning/STATE.md` | `c28a5df6f27f6dc71a300dac2ac913854beccdba` |
| 2 | `.planning/decisions/specs/MAIL-SPEC.md` | `5d9679c4aafba13fe303bc9c149f71e43c44b165` |
| 3 | `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md` | `c6a7d7e4943d8338c85108daf3fc5e9dfc43f279` |
| 4 | `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md` | `a531fb36f584612e30d7a5fd8b2a9a682147d41a` |
| 5 | `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-04G.md` | `f257e6186b4dfc54f087cf7077ca3fd3bc3b50b8` |
| 6 | `.planning/phases/144-mailbox-domain/PLAN.md` | `36d44e1821804ad40bfcefd250850112f935ece9` |
| 7 | `.planning/phases/144-mailbox-domain/REVIEW-144-04G-SYNC-SERVICE-READINESS.md` | `a2571427a59f56a0cb431775791b9a06d4c89eb8` |
| 8 | `docs/mail/CONTRACT-CHANGELOG.md` | `c306dcd9b0e135edb38163858b6236aad11d311d` |
| 9 | `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` | `41415ed55cb3c745d09abeceea3ddb55facd78b3` |
| 10 | `docs/mail/QA-HANDOFF.md` | `c82c770f8b7843ab16ab34676ed06ad787412e34` |
| 11 | `docs/mail/events-v1.schema.json` | `89a5868341b0603e27082cd68272bb4f05efda4c` |
| 12 | `docs/mail/openapi.yaml` | `55cf67e6008a3b36e98270b29f39442fe4c1a1aa` |

## Current coherence exact12 manifest

The required atomic coherence commit diff is set-equal to this ordered exact12 manifest:

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
4. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
5. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-04G.md`
6. `.planning/phases/144-mailbox-domain/PLAN.md`
7. `.planning/phases/144-mailbox-domain/REVIEW-144-04G-SYNC-SERVICE-READINESS.md`
8. `docs/mail/CONTRACT-CHANGELOG.md`
9. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
10. `docs/mail/QA-HANDOFF.md`
11. `docs/mail/events-v1.schema.json`
12. `docs/mail/openapi.yaml`

The JSON Schema and OpenAPI entries in this manifest may change only status metadata,
comments or `x-*` extensions; no schema or operation semantics change.

Mail 04F remains accepted. Mail 04H and 02E, Ent/SQL/concrete adapter,
generation/migration, API/UI, launcher/testframework and runtime remain blocked or
`NOT_READY`/`NOT_RUN`; no next Mail packet is selected or authorized.
