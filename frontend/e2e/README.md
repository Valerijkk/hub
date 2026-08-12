## Legacy TypeScript E2E quarantine

This root is legacy TypeScript E2E inventory. It provides no supported command,
runtime authority, or fallback, and it is not the future Python E2E root.

It is deleted only by the exact Phase 129 manifest. The planned Python E2E
location is `testframework/tests/e2e/**`, created only by its first accepted
consumer. Until the separate future `testframework/runner/**` implementation
and independent acceptance, launcher status is `NOT_READY` and runtime is
`NOT_RUN`; `testframework/docker/**` remains immutable-assets-only.
