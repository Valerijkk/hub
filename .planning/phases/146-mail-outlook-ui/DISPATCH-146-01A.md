[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный TypeScript-исполнитель Hub. Расширь только shared API client
безопасными опциями для mail/calendar mutation headers и recovery metadata.

- User outcome: все create/bulk, draft/send/RSVP используют стабильные
  `Idempotency-Key`/`If-Match`, а
  клиент безопасно получает error code, ETag и same-origin refetch location.
- Completion: failing genuine units сначала; затем минимальная backward-compatible client
  change, включая сохранение опций через единственный auth refresh retry.
- Non-goals: mail API module, React/TanStack hooks/UI, backend, route, dependency change,
  arbitrary header API, storage, tests outside feature root.

[OWNERSHIP]

- Allowed mutations only:
  - `frontend/src/api/client.ts`
  - `frontend/tests/mail/api-client-mail-headers.test.ts`
- Do not touch auth store, other API modules, vite config, package files, protected
  testframework path or generated route tree.
- Entry requires founder-approved frontend feature-test harness. If unavailable, file
  implementation is also `BLOCKED`; do not create an unexecutable test silently.

[FROZEN PUBLIC SHAPE]

Add a narrow options type equivalent in meaning to:

```ts
export interface ApiRequestOptions {
  idempotencyKey?: string;
  ifMatch?: string;
  signal?: AbortSignal;
}
```

`request`, `post`, `put`, `patch`, and only methods that need it accept options without
breaking existing call sites. The client creates headers itself:

- `Idempotency-Key` exactly from a validated non-empty caller value;
- `If-Match` exactly from a validated quoted positive revision;
- caller cannot override `Authorization`, `Cookie`, `Host`, `Content-Type` or pass an
  arbitrary header map.

Extend `ApiError` with optional `error_code`, `etag`, `location`, and
`retry_after_seconds`. Read `X-Hub-Error-Code` and ETag as bounded strings. Reuse the
existing Retry-After policy, accepting only an integer 1..5 for `MAIL_ACCOUNT_BUSY`.
Accept Location only if it is a relative same-origin `/api/...` path; reject scheme-
relative, absolute, backslash/control and non-API values. The calendar wrapper will
enforce its stricter canonical-event regex.

[REQUIRED ORACLES]

1. Existing no-options GET/POST/PATCH behavior and body encoding remain unchanged.
2. Idempotency/If-Match appear exactly once with correct casing/value.
3. A 401 then successful refresh retries once with identical body, key, If-Match and
   AbortSignal; no new key is generated inside client.
3A. Folder/label/signature/rule/presign/admin-domain/alias create and message bulk callers
    pass a fresh cryptographic UUID v4/v7 for a new intent and the same narrow key through
    every retry; the client never scopes, downgrades or regenerates it mid-intent.
3B. Label/settings/signature/rule/mailbox-policy/domain update and versioned delete callers
    preserve the last confirmed quoted If-Match through the same single auth retry.
4. Retry 401 still expires session as before.
5. Absolute, `//host`, backslash/control and non-`/api/` Location are discarded.
6. Valid relative `/api/calendar/v2/events/<uuid>` location, error code and ETag are
   available on thrown ApiError; 423 busy exposes only a validated 1..5 second retry value.
7. Malformed error JSON keeps statusText fallback and never throws during mapping.
8. Existing 429 retry-after behavior remains intact; missing/fractional/out-of-range 423
   Retry-After is discarded, and busy is never mapped from an unknown engine outcome.
9. Compile-time negative cases prove arbitrary headers cannot be passed.
10. No token, key or request body is copied into an error object/message.

[EXECUTION / PROVENANCE]

- TDD first; mock global fetch/auth store only with repository-native Vitest facilities;
  no MSW/new dependency.
- Host tests/build/lint forbidden; accepted Docker launcher only. Unavailable => `NOT_RUN`.
- No commit/push/PR/network/generated files. max_depth 0; fan_out 0; retries 2; 45m.
- requested_model `gpt-5.6-terra`; reasoning `xhigh`; fallback none; mismatch => `BLOCKED`.

[ОТВЕТ]

На русском: API delta; changed files; retry/security oracle; verification; model
provenance; remaining risks.
