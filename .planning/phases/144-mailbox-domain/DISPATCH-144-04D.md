[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только сервисный state machine
создания, чтения, сохранения и удаления mailbox draft после принятых 144-01/02/03/04A.

- User outcome: autosave не теряет более новую редакцию и не дублирует тело письма в
  PostgreSQL; потерянный ответ движка восстанавливается безопасно.
- Completion: сначала failing external-package units; затем минимальный draft service,
  который проходит все ветви create replay, If-Match и unknown reconciliation.
- Non-goals: HTTP handlers, UI, send/submission, storage upload, Ent schema/generation,
  Stalwart/JMAP, NATS, search, migration, Resend code.

[OWNERSHIP]

- Entry: root independently accepted 144-01B, 144-02C/E, 144-03C and 144-04A on clean
  reviewed HEAD. Если их surface отличается от frozen contract — `BLOCKED`.
- Allowed mutations only:
  - `backend/internal/mail/draft_service.go`
  - `backend/tests/mail/draft_service_test.go`
  - узкое добавление draft repository method signatures в уже принятом
    `backend/internal/mail/mailbox_repository.go`; implementation туда не добавлять.
- Не изменяй engine port, schema/generated Ent, handler/wiring, submission service или
  существующий Resend subsystem.
- Package test: `mail_test`; deterministic fakes; no filesystem/network/DB/broker/clock.

[FROZEN STATE MACHINE]

```text
create key + input hash
  absent -> reserve saving -> engine SaveDraft(create) -> active revision 1
  same completed -> replay same draft
  same key/different hash -> idempotency conflict

active revision N + If-Match N
  -> persist saving + stable mutation id
  -> engine SaveDraft(expected engine version)
  -> persist returned version, active revision N+1

lost/indeterminate response
  -> persist unknown, keep mutation id and previous version
  -> next read/save calls ReconcileDraft for that mutation/version
  -> confirmed new engine version: finalize N+1 once
  -> confirmed absence/no change: return active N, then caller may issue a new mutation
  -> indeterminate: remain unknown; no second SaveDraft
```

Repository transaction/lease must serialize one draft mutation. Do not hold an ordinary
database transaction open across an unbounded engine call: use a persisted mutation state
and bounded ownership/lease primitive already frozen by 144-04A. Any missing primitive is
`BLOCKED`, not an improvised global mutex.

PostgreSQL draft rows may contain only account/engine ids, opaque versions, Hub revision,
state, idempotency/request hashes, reply/forward links, mutation id and timestamps.
Recipients, subject, rich/plain bodies and MIME appear only in the engine request/result
and the authorized HTTP response value, never in repository records/log/errors.

Before every engine-authoritative draft create/save/delete or unknown-result reconciliation
workflow, the service atomically acquires the owning account mutation lease and increments
`mutation_epoch` through `BeginEngineCommand`; it commits that first edge before engine I/O.
Completion/unknown handoff verifies the exact token, reconciles as required, clears the
lease and increments the epoch again through `FinishEngineCommand` in one transaction.
Hub-only metadata changes do not acquire it. A competing active lease returns the stable
retryable busy result; a reclaimed/expired token cannot finalize and must reconcile under a
new lease. The service never holds a database transaction open across engine I/O.

[REQUIRED ORACLES]

1. Client account/org fields do not exist; authenticated account resolution happens before
   repository or engine access.
2. Create same key + same canonical input returns the original draft and one engine call.
3. Create same key + changed recipient/body/subject/attachment input conflicts before
   engine call; raw key is absent from repository fake snapshots and returned JSON values.
4. Correct `If-Match` saves once and increments Hub revision exactly once.
5. Stale `If-Match` returns conflict and makes zero engine mutation.
6. Lost engine response persists `unknown`; an immediate retry calls reconciliation, not
   `SaveDraft`.
7. Confirmed reconciliation finalizes once; repeated reconciliation/retry replays it.
8. Indeterminate reconciliation remains unknown with stable safe error and zero new save.
9. Cross-account draft id returns not-found/forbidden policy with zero engine call.
10. Read hydrates content from engine only after owner resolution; repository body capture
    assertion remains empty.
11. Delete is idempotent and never deletes another account's engine draft.
12. Errors/loggable values do not contain synthetic body, recipient, raw key or engine
    response.
13. Each engine-authoritative draft command/reconciliation brackets I/O with exact account
    mutation-lease begin/end epoch edges; active-lease busy makes zero engine call, expiry/
    reclaim forbids late finalization, and no sync page can capture during the lease or
    overwrite the reconciled draft state.

[EXECUTION / PROVENANCE]

- TDD first; no new dependency, generated code or broad refactor.
- Host verification and direct Docker/Compose forbidden. Accepted launcher absent =>
  `NOT_RUN` plus exact static evidence.
- No commit/push/PR/network; max_depth 0; max_fan_out 0; retries 2; wall clock 60m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none; wrong actual runtime
  means immediate `BLOCKED` before mutation.

[ОТВЕТ]

На русском: state transitions implemented; files/symbols; oracle matrix; verification;
privacy evidence; requested/actual model; risks/unknowns.
