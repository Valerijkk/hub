[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный TypeScript/React hook-исполнитель Hub. Реализуй только send
intent, submission receipt polling and unknown recovery after accepted draft autosave.

- User outcome: двойной клик не отправляет второе письмо; timeout/unknown сходится через
  тот же idempotency intent без потери черновика.
- Completion: deterministic hook units first; implementation proves stable key, one POST,
  receipt polling and content-clear conditions.
- Non-goals: autosave implementation, visual composer, API client/types, attachments,
  backend, route, WebSocket, persistent offline outbox.

[OWNERSHIP]

- Entry: independently accepted 146-04C on clean reviewed HEAD plus 146-01A/B.
- Allowed mutations only:
  - `frontend/src/modules/mail/use-mail-submit.ts`
  - `frontend/tests/mail/use-mail-submit.test.tsx`
- Do not edit `use-mail-draft.ts`; consume only its accepted public state/actions.
- No local/session storage, Zustand, global mutable key or body-bearing mutation metadata.

[FROZEN SEND]

- One explicit user send intent creates one cryptographically strong key outside render
  and retains it for retries/polls of that intent. It is never persisted or logged.
- POST carries the exact last-confirmed quoted draft revision; a retry preserves both key and
  `If-Match`. Same-key receipt replay is accepted even when the draft has since become
  submitted; a new-key stale/submission conflict refetches and never blindly sends.
- Double click/Enter while active returns the same promise/state and makes one POST.
- Send waits for current autosave to confirm/reconcile; conflict, unknown or submitting
  draft blocks a new send with actionable state.
- If POST returns a receipt id in queued/submitting/unknown, poll
  `GET /mail/submissions/{id}` with bounded backoff and stop on unmount/cancel.
- If fetch fails before a receipt id is known, expose explicit unknown state; an explicit
  retry calls the same POST with the original key and frozen intent. Never create a new key
  merely because the network timed out.
- `sent` clears only the matching frozen composer snapshot after message confirmation.
  `failed` preserves content; later edits typed during send are never cleared.

[REQUIRED ORACLES]

1. Double send produces one POST with one key.
2. API client's one auth retry preserves the same key and exact draft `If-Match`
   (integration with 146-01A oracle).
3. Queued/submitting/unknown receipt polls GET only and never repeats POST.
4. Network failure before receipt id creates unknown UI; explicit retry reuses original key
   and exact frozen intent.
5. Repeated indeterminate status does not create a second POST and stops at bounded policy.
6. Definitive sent clears only matching snapshot; edits made during send remain dirty.
7. Failed preserves all content and exposes only safe error code/message.
8. New key can be created only by an explicit new-intent transition after prior terminal
   resolution, never on render/remount timer.
9. Unmount aborts polling/timers without later state updates.
10. Query/mutation/error metadata and logs contain no body, recipients or key.

[EXECUTION / PROVENANCE]

- TDD first, fake timers/client, no MSW/new dependency.
- Host test/build/lint forbidden; accepted Docker launcher only. Missing harness/launcher
  => `NOT_RUN`.
- No commit/push/PR/network/generated files. max_depth 0; fan_out 0; retries 2; 45m.
- requested_model `gpt-5.6-terra`; reasoning `xhigh`; fallback none; mismatch => `BLOCKED`.

[ОТВЕТ]

На русском: send state table; files/symbols; key lifecycle/oracles; verification;
requested/actual model; risks.
