# 144-03C-Attachment implementation evidence receipt

## Git boundary

Source receipt is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
`7b2ea5721684e43559d6f6b4b5fc462d8daafb8b` (parent
`4e7c609713983a8f6db6fbfeb7474abd73451e8c`, tree
`4c3fe4d27e170c5d3f42b116c39887036ce936c6`; subject
`feat(mail): add deterministic attachment policy`). The exact two-path boundary is
`backend/internal/mail/attachment_policy.go`
(`3f67ce1b9f2f056b70bfd3070f1c34f98b5c71c1`) and
`backend/tests/mail/attachment_policy_test.go`
(`bf7eda125d5aba1d0cc2594746d14e7bcfd06aa7`).

## Behavior and oracle matrix

| Behavior | Static oracle result |
|---|---|
| Attachment bounds | Deterministic per-attachment policy with at most 50 items and bounded size, filename, media type and inline CID fields |
| Filename and type defense | NFC output with NFKC structural/extension defense and closed unsafe MIME/type deny tables |
| Scan state | Validate accepts structurally valid metadata; send/read require clean scan state and valid SHA gate |
| Error and ownership safety | Typed content-free errors and deep-copy result; no I/O, authz, clock or global-state dependency |

## Provenance and limits

Tests were authored first, but RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime are
`NOT_RUN` because the launcher is `NOT_READY`. Authoring was `gpt-5.6-terra` / `xhigh`,
dispatcher-enforced with no fallback; telemetry is `UNKNOWN` unless exposed. Contract Terra/high,
Security Terra/xhigh and ABI/Lifecycle Terra/high precommit plus exact-SHA reviews PASS C/I/M
`0/0/0`.

This is static-only and grants no scanner, storage, network, project or tenant authorization,
schema, migration, service, API, UI, product, phase or release authority. Existing
144-01B/02/Rule/Sieve/Phase145 blockers and the Calendar 143-03E launcher+migration blocker
remain unchanged. No further source packet is presently authorized: root-only dependency/adjudication
must select or unblock 144-01B/02E for 144-04A, or the Calendar 143-03E launcher+migration gate.
