# 144-03C-Send implementation evidence receipt

## Git boundary

Source receipt is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
`f8f87b09c4fe65be612fcf6d15ba1f4bd5f8aa19` (parent
`de6253963c3f787405b3aa63245c0acb0a8cd159`, tree
`2a4282ef0cf9d24cc722238cc5c28e792a4a8f2a`; `feat(mail): add deterministic send policy`).
The exact two-path boundary is `backend/internal/mail/send_policy.go`
(`07f2641f6efa846110e82240a72bc67b4a4f4509`) and
`backend/tests/mail/send_policy_test.go`
(`ab47f81812f0aecb05ed0c1d17e7b0f7ee340e78`).

## Behavior and oracle matrix

| Behavior | Static oracle result |
|---|---|
| Send/rate/loop policy | Pure deterministic allow/reject result with bounded recipients, token and loop handling |
| Persistence-backed policy floor | Root-retained floor of `1,048,576` bytes |
| Safety remediation | Fail-closed token and HTAB handling |

## Provenance and limits

Tests were authored first, but RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime are
`NOT_RUN` because the launcher is `NOT_READY`. Authoring was `gpt-5.6-terra` / `xhigh`,
dispatcher-enforced with no fallback; telemetry is `UNKNOWN` unless exposed. Contract Terra/high,
Security Terra/xhigh and ABI/Lifecycle Terra/high precommit plus exact-SHA reviews PASS C/I/M
`0/0/0`.

This is static-only and grants no schema, migration, service, API, UI, product, phase or release
authority; it does not claim tenant authorization. Existing 144-01B/02/Rule/Sieve/Phase145
blockers remain unchanged. Next mainline permitted packet is the separate `144-03C-Attachment`.
Calendar 143-03E RSVP/service remains blocked on launcher and migration authority.
