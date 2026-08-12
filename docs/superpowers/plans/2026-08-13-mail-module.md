# Hub Mail + Calendar Sync implementation plan

> **Implementation agent:** follow `superpowers:executing-plans` task by task. Code-writing
> tasks require the founder-routed `gpt-5.3-codex` / `xhigh` worker after this contract is
> frozen. Root/Terra retain architecture, review, and acceptance authority.

**Goal:** Build the self-hosted Outlook-class Hub mail module and canonical RFC 5545/5546
calendar RSVP flows without changing the protected test-rebase program.

**Architecture:** Hub owns calendar, policy, projections, search/realtime, audit and UX;
an adapter-backed engine owns raw MIME, mailboxes, flags, protocols and deliverability.
Internal and external delivery both use the engine. Pure domain seams and a fake adapter
allow implementation before the Stalwart runtime gates are available.

**Stack:** Go 1.25, Fiber v3, Ent/PostgreSQL, NATS/outbox, Meilisearch, SeaweedFS,
React/Vite, TanStack Query, Plate, Vitest/RTL, optional Stalwart Docker profile.

## Non-negotiable execution rules

- Do not read or change `testframework/**`, `.planning/audits/test-rebase/**`, phases
  128–141, TASK-193–234, `pro-qa-ai/**`, `.env.local`, or `.git/index.lock`.
- Do not add tests under legacy `backend/testframework/**`.
- Tests, generators, build, lint, validators, Vite, backend processes, and runtime checks
  run only through the accepted task-owned Docker launcher. If it is unavailable, record
  `NOT_RUN`; do not use host or direct Docker/Compose fallback.
- Every code task begins with failing genuine units in the target test layout, then the
  minimal implementation, then Docker verification when available.
- Every commit is atomic, Conventional Commits, no AI attribution, Russian body, and
  same-session docs sync for changed behavior.
- No push, PR, merge, deployment, DNS change, or license action without founder authority.

## Dependency graph

```text
142 contracts
  ├─ 143 calendar domain/iTIP/API ─┐
  └─ 144 mailbox domain/fake API ──┼─ 146 frontend ─ 147 reliability ─ 148 gate
                    └─ 145 engine/runtime (external gates) ───────────┘
```

## Phase execution index

1. Phase 142 — `.planning/phases/142-mail-contract-foundation/PLAN.md`
2. Phase 143 — `.planning/phases/143-calendar-rsvp-core/PLAN.md`
3. Phase 144 — `.planning/phases/144-mailbox-domain/PLAN.md`
4. Phase 145 — `.planning/phases/145-stalwart-gateway/PLAN.md`
5. Phase 146 — `.planning/phases/146-mail-outlook-ui/PLAN.md`
6. Phase 147 — `.planning/phases/147-mail-calendar-reliability/PLAN.md`
7. Phase 148 — `.planning/phases/148-mail-release-gate/PLAN.md`

## Checkpoint policy

The unattended run proceeds through local reversible work. It pauses a particular slice
only when it hits an explicit external gate and continues independent slices. Founder
questions accumulate in `docs/mail/FOUNDER-QUESTIONS.md` for the next check-in.

Phase 145 may remain blocked while Phases 143, 144 and engine-independent parts of 146
continue. Phase 148 cannot close before founder notice that the protected test branch has
been merged and an accepted launcher can produce current-SHA evidence.
