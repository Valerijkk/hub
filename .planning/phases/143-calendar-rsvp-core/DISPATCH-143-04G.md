[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только admission внутренних
calendar subjects в существующий JetStream EVENTS stream.

- User outcome: canonical calendar outbox events могут быть durably published до запуска
  первого producer, но приватные event metadata не становятся внешними webhooks.
- Completion: failing/snapshot genuine unit сначала; затем одна минимальная subject-list
  правка; outbound webhook list строго не меняется.
- Non-goals: calendar service/event mapping, broker runtime, webhook feature, mail subject,
  HTTP, Ent, frontend, migration, any protected test suite.

[OWNERSHIP]

- Entry: root accepted 143-04A event subject contract and 143-04L envelope contract on a
  clean reviewed HEAD.
- Allowed mutations only:
  - `backend/internal/platform/nats/client.go`
  - one focused genuine unit under `backend/tests/calendar/**` that imports public/exported
    subject inventory if available; if the production list is not externally observable,
    use the repository's existing genuine unit location only when it is outside protected
    roots and explicitly approved by orchestrator; otherwise return `BLOCKED` for test seam
- Do not modify `WebhookFilterSubjects`, webhook packages/allowlists, event payload schema,
  mail package or any `backend/testframework/**` path.
- You are not alone; stop on dirty overlap and never revert another packet.

[FROZEN CHANGE]

- Append/add `calendar.>` exactly once to canonical `StreamSubjects` while preserving all
  prior subjects and stable ordering conventions.
- Preserve `WebhookFilterSubjects` byte-for-byte. Do not add `calendar.*`, wildcard or any
  individual calendar subject to external dispatch.
- Do not add `mail.>`; Phase 144 owns that later, after this packet is accepted.
- No change to stream name, retention, replicas, max age/storage or consumer behavior.

[REQUIRED ORACLES]

1. Canonical stream set is exactly prior set plus one `calendar.>` entry.
2. No duplicate subject and no broader `>`/`*.>` wildcard is introduced.
3. Exact outbound webhook filter snapshot is unchanged.
4. All five schema-listed calendar subjects (`calendar.invite.created|updated|cancelled.v1`,
   `calendar.rsvp.changed.v1`, `calendar.reminder.due.v1`) match `calendar.>` and no
   `mail.*` or linked-call subject is claimed by this packet.
5. No runtime broker interaction is needed for the pure inventory oracle.

[EXECUTION / PROVENANCE]

- TDD first; no dependency/generated code/network/broker/listener.
- Host tests/build/lint/validators and direct Docker/Compose forbidden. Accepted launcher
  unavailable => runtime `NOT_RUN`; static diff/oracle review only.
- No commit/push/PR. max_depth 0; fan_out 0; retries 2; wall clock 30m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none; actual mismatch means
  immediate `BLOCKED` before mutation.

[ОТВЕТ]

На русском: exact before/after sets; changed files; webhook immutability evidence;
verification; requested/actual model; unknowns.
