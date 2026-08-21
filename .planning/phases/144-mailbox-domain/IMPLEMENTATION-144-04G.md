# 144-04G — mailbox sync coordinator immutable source receipt

> **Acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact12 working tree grants no authority. The proposed immutable receipt below activates only after one atomic commit of exactly the 12-document manifest stated here; until then, the prior 04G source-authority, source-unimplemented and current-next lifecycle remains current. Activation selects no next Mail packet and lifts no downstream blocker.

## Proposed static acceptance (2026-08-21)

After the acceptance-sync gate commits, Mail 144-04G is `STATIC_ACCEPTED / EXACT_SHA_PASS` for immutable source commit `3031c53d399accc84fe10cd0c5421a85c089ac7c`:

- parent: `085a81aed317be364811af9507912002a7dae457`
- tree: `b146acb45bf60f2f468a926caa3546c0b8f79893`
- subject: `feat(mail): add mailbox sync coordinator`
- tests first: `backend/tests/mail/mailbox_sync_service_test.go` — `4b8de26d16a749536b3fc3207999d06dc20ab379`
- repository: `backend/internal/mail/mailbox_repository.go` — `3b16c7a0dcf4fc89d3b773cfb8d3f75cee6219d4`
- service: `backend/internal/mail/mailbox_sync_service.go` — `0ed93f7247b428e8af8de2c6bd0f82bc5b9bb441`

Contract, Security and Lifecycle review each recorded precommit and immutable `EXACT_SHA_PASS`, with `C/I/M 0/0/0`. The author provenance is `gpt-5.6-terra` / `xhigh`, dispatcher-enforced with no fallback; telemetry is `UNKNOWN` because it is not exposed.

This receipt is static source-only evidence for the frozen coordinator ABI, state transitions and contract oracles. It is not runtime evidence and does not authorize an Ent or SQL schema, concrete adapter, generated binding, migration, public OpenAPI operation, or any execution. The launcher remains `NOT_READY`; executables and runtime remain `NOT_RUN`.

Only after the atomic exact12 commit may the prior 04G source-authority, source-unimplemented and current-next routes become `CONSUMED/HISTORICAL`. Mail 04F remains accepted. Mail 04H and 02E remain blocked; no next Mail packet is selected.

## Exact12 acceptance-sync manifest

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/144-mailbox-domain/PLAN.md`
4. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
5. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
6. `.planning/phases/144-mailbox-domain/REVIEW-144-04G-SYNC-SERVICE-READINESS.md`
7. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-04G.md`
8. `docs/mail/CONTRACT-CHANGELOG.md`
9. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
10. `docs/mail/QA-HANDOFF.md`
11. `docs/mail/events-v1.schema.json`
12. `docs/mail/openapi.yaml`
