# Phase 149 - Verification

**Verdict: PARTIAL (2026-08-25).** The frozen unit and static slice is accepted; external integration and browser evidence were intentionally not fabricated.

| Area | Evidence | Verdict |
|---|---|---|
| Static independent review | Full Phase 149 diff reviewed after implementation | `PASS` - C0/I0/M0 |
| Backend units | `backend-unit`: `first_exit:0`; build, vet, unit, race all `PASS` | `PASS` |
| Frontend units | `frontend-unit`: `first_exit:0`; TypeScript and Vitest `PASS` | `PASS` |
| MongoDB | Unique-index/upsert behavior and Ent direct-DB current-audience behavior were not run against Mongo | `NOT_RUN` |
| NATS | No real broker publication/consumption execution | `NOT_RUN` |
| WebSocket/browser | No external live WS or browser E2E execution | `NOT_RUN` |
| Python integration | `python-integration` was not run | `NOT_RUN` |

## Accepted contract

- Own-room delivery is exactly `user:<uuid>`. NATS user events are `messenger.user.<uuid>.channel_created`, `.channel_updated`, `.channel_removed`, and `.new_message_notification`; payloads are exactly `{channel_id}` or `{channel_id,message_id}`.
- The bridge strictly validates the user subject and payload, then fails closed. Messenger and channel-scoped Calls resolve current channel membership immediately before delivery; call chat uses `CallChatAudienceResolver`, with instant calls targeting active participants plus creator. Each path filters connections by authenticated user ID.
- Lifecycle delivery is post-success and recipient-specific. All five canonical message producers notify current members except the canonical sender; idempotent nonce retries do not add notifications.
- `useChannels` itself does not subscribe. `useMessengerChannelsRealtime` in `MessengerLayout` owns the subscription and uses subscribe-then-exact-invalidate on reconnect. A session boundary resets WebSocket, shared QueryClient, and Messenger state before local auth mutation/broadcast; token refresh preserves state.

## Residual boundary

- NATS publication is best-effort; no transactional outbox exists.
- Current-audience lookup followed by fanout retains a bounded lookup-to-send TOCTOU window.
- Delete recipient snapshot followed by concurrent membership insertion can race.
- Durable unread/read-state projection recovery belongs to Phase 151 and remains `NOT_READY`.