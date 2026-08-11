# TASK-233 — 🚦 Greenfield test CI gates

> Created: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **141**

> Runtime/traceability gate: every target job and launcher are `NOT_READY`; runtime
> is `NOT_RUN`. A future versioned launcher API may be separately implemented
> only under `testframework/runner/**` and used only after independent acceptance;
> `testframework/docker/**` supplies immutable execution assets only. No supported
> command, Docker-directory entrypoint, or direct Docker/Compose, host
> test/validator, mutable-image, or ad-hoc-cleanup fallback exists. Executable
> evidence must resolve TASK/requirement/TC/risk after acceptance.

## [OUTCOME]

GitHub CI is the fail-closed authority for new unit+Python architecture with
Docker provenance, traceability, sentinel, artifact and cleanup gates.

## [ROLE]

- Owner: `devops-engineer`; verifier: `qa-auto-ver`.
- Depends on: Phases 132–140.

## [OWNED TARGET PATHS]

`.github/workflows/**`, `.gitlab-ci.yml`, `scripts/**`; future CI consumers may
call only the separately implemented and independently accepted API under
`testframework/runner/**`. They consume immutable execution assets from
`testframework/docker/**` but never execute an entrypoint from that directory.
Until that acceptance, every target job is `NOT_READY` and CI has no direct
fallback. This card authorizes no path beyond its live PLAN/manifest and routes
no legacy non-unit row to Python.

## [ACCEPTANCE]

- Required jobs use the separately accepted runner API with pinned immutable
  execution assets and resource limits.
- Zero collection, unexpected skip, invalid/orphan TC metadata, missing artifact,
  secret canary and lingering task resource fail the build.
- Clean checkout real CI RED sentinel→GREEN; mirrors share scripts or are retired.
- No removed legacy path/reference; no push/merge without founder request.
