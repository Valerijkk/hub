[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только mail-owned object
storage boundary для вложений; не переиспользуй project file authorization.

- User outcome: пользователь может загрузить/подтвердить/прочитать/удалить вложение
  только своего mailbox/draft, а storage key не выбирается клиентом.
- Completion: failing units сначала; затем узкий port + adapter/service policy с полной
  owner/account reauthorization и lifecycle cleanup semantics.
- Non-goals: HTTP handler, message send, scanner implementation, SeaweedFS network client
  rewrite, Ent schema/generation, UI, NATS, Stalwart, project file behavior.

[OWNERSHIP]

- Entry: accepted 144-02C/E, 144-03C and 144-04A.
- Allowed mutations only:
  - `backend/internal/mail/mailbox_storage.go`
  - `backend/internal/mail/mailbox_storage_adapter.go`
  - `backend/tests/mail/mailbox_storage_test.go`
- Existing `backend/internal/storage/**` may be imported only through a storage-byte
  primitive that does not impose `project_id`; do not modify it in this packet. If no safe
  primitive exists, define a mail-owned port and return adapter implementation `BLOCKED`
  rather than abusing `FileMetadata`.
- No project membership or client-provided organization/account/object path.

[FROZEN CONTRACT]

- Server generates `mail/<organization>/<account>/<object-uuid>` after resolving the
  authenticated account. Reject separator/control/traversal in normalized filename but do
  not derive object key from filename.
- Presign persists a pending `MailObject` lease with expected size/type/name, uploader,
  draft client id and expiry before returning a bounded upload contract.
- Confirm re-resolves owner/account, checks pending state/expiry, reads authoritative
  object metadata, requires exact size and supplied SHA-256, transitions to scan pending,
  and is idempotent for an identical confirmation.
- Read/download requires confirmed clean attachment plus owner/account. Delete is
  idempotent; pending/orphaned cleanup is exact-object only and never prefix-wide.
- Upload URL may target the configured storage origin; API never accepts an arbitrary
  download/proxy URL. Required headers are an explicit allowlist.
- Scanner result is fail-closed: only `clean` is attachable/sendable.

[REQUIRED ORACLES]

1. Two accounts using the same filename receive unrelated server keys.
2. Client org/account/key fields cannot influence the generated key.
3. Cross-account presign/confirm/read/delete makes zero storage call after owner check.
4. Confirm size/hash mismatch or expiry fails and remains non-clean.
5. Duplicate identical confirm is replayed; changed hash/size conflicts.
6. Pending/rejected/error scan state cannot be attached or downloaded as clean.
7. Cleanup deletes only an expired object selected by immutable id + owner marker; a
   neighboring object survives.
8. Adapter never constructs/queries project-scoped `FileMetadata` or project membership.
9. Safe errors/log values omit upload URL query, object credentials, filename controls and
   synthetic hash secrets.
10. No storage call receives an unresolved or empty account id.

[EXECUTION / PROVENANCE]

- TDD first with in-memory repository/storage fake; no filesystem/network/DB/global clock.
- Host validation/direct Docker forbidden; accepted launcher absent => `NOT_RUN`.
- No new dependency/generated code/commit/push/PR. max_depth 0; fan_out 0; retries 2;
  wall clock 60m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none; mismatch => `BLOCKED`.

[ОТВЕТ]

На русском: ownership model; exact files/symbols; lifecycle oracle matrix; verification;
proof that project ACLs were not reused; model provenance; risks.
