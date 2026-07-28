# ADR-WS-JWT-HANDSHAKE-2026 — JWT вне sessionStorage + WS ticket/subprotocol

> **Status:** Accepted · **Date:** 2026-07-28 · **Deciders:** owner lock TASK-008 / A2-DEDUP (отдельный approve не требуется)  
> **Task:** `.planning/backlog/TASK-103-2026-07-28-jwt-ws-handshake-adr.md`  
> **Findings:** A2W3-4-001, A2W11-2-001, A2W11-3-002

## Context

1. Access + refresh JWT лежали в `sessionStorage` (`hub-auth`) — XSS крадёт refresh (30d).
2. WS upgrade принимал access JWT в `GET /ws?token=` — токен утекает в access/proxy logs, history, Referer.
3. WebSocket browser API не даёт кастомный `Authorization` на handshake; допустимы `Sec-WebSocket-Protocol` и short-lived ticket.

## Decision

1. **FE storage:** access и refresh JWT — **только memory** (Zustand без persist токенов). `sessionStorage`/`localStorage` не содержат JWT. Reload вкладки → повторный login (до будущего ADR httpOnly refresh cookie).
2. **WS auth:** short-lived **одноразовый ticket** (Redis, TTL 30s):
   - `POST /api/auth/ws-ticket` (Bearer access JWT) → `{ "ticket", "expires_in" }`
   - `GET /ws` с `Sec-WebSocket-Protocol: hub.ws.ticket.<ticket>`
   - Сервер делает atomic redeem (`GETDEL`), затем upgrade; RBAC/`CanSubscribe` **не ослабляются**.
3. **Denial:** любой непустой `?token=` на `/ws` → **401** (query auth запрещён навсегда).

## Consequences

### Positive
- JWT не в query logs / sessionStorage
- Ticket короткоживущий и one-shot
- Subprotocol совместим с браузерным `WebSocket`

### Negative
- Reload = re-login (пока нет httpOnly cookie ADR)
- Лишний round-trip `POST /auth/ws-ticket` перед каждым WS connect/reconnect

### Out of scope
- httpOnly Secure cookie для refresh (отдельный ADR)
- First-message post-connect auth (не выбран)

## Compliance

Закрывает A2W3-4-001 / A2W11-2-001 / A2W11-3-002.  
NF-02 (room authz) без изменений.
