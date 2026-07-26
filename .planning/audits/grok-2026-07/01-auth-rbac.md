# Auth / RBAC / Audit / Platform

## Что проверено

- **auth** (16 `.go`): `handler.go`, `middleware.go`, `magiclink.go`, `service_login.go`, `service_session.go`, `service_organization.go`, `service_registration.go`, `repository.go`, `resend.go`, `telegram.go`, `telegram_bot.go`, `service.go`, `dto.go`, `events.go` — полное чтение критичных путей + grep `_ =`, `go func`, `ConstantTimeCompare`, `CreateSession`, `StatusPending`
- **rbac** (8 `.go`): `handler.go`, `middleware.go`, `service.go`, `custom_roles.go`, `repository.go`, `cache.go`, `permissions.go` — чтение + grep IDOR/`projectId`/`IsSystem`/`RequireAdmin`
- **audit** (9 `.go`): `handler.go`, `handler_alerts.go`, `service.go`, `repository.go`, `middleware.go`, `consumer.go`, `retention.go`, `dto.go` — чтение + grep Ack/Ban/GDPR/pagination/`_ =`
- **platform** (33 `.go`): `outbox/{worker,publisher}.go`, `middleware/{ratelimit,logging,cors,security_headers,requestid,middleware}.go`, `http/{safedialer,opengraph,helpers}.go`, `crypto/aesgcm.go`, `cron/cron.go`, `nats/{client,trace}.go`, `config/config.go`, `database/*`, `redis/*`, `observability/*`, `errors/*`, `telegram/*` — чтение критичных + grep `go func`/`DefaultClient`/`io.ReadAll`/Tx/Rollback
- Паттерны grep: `_ =`, `go func`, `context.Background`, `http.DefaultClient`, `io.ReadAll`, `panic(`, `ConstantTimeCompare`, `Offset(`/`Page`/`PerPage`, `SystemRoleAdmin`/`Founder`, `Ack(`

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A1-001 | P0 | `backend/internal/auth/service_login.go:64` (+ `middleware.go:128`) | ORG-06 bypass: после Magic Link verify создаётся сессия для `StatusPending` | `LoginWithMagicLink` всегда вызывает `CreateSession`; `userIsDeactivated` считает pending «не deactivated» → JWT и доступ к API до approve | После verify: если `u.Status != active` → 403 без сессии; в `RequireAuth`/`RefreshToken` fail-closed на pending; regression-тест |
| A1-002 | P0 | `backend/internal/rbac/custom_roles.go:73-84` (+ `handler.go:99-114`, `repository.go:85-91`) | Cross-project IDOR: Update/Delete роли по `roleId` без проверки принадлежности к `:projectId` | Middleware проверяет права только на `projectId` из URL; `UpdateCustomRole`/`DeleteCustomRole` мутируют любой `roleID` (в т.ч. чужой проект / system role на Update) | Перед Update/Delete: `GetRoleProjectID` == path `projectId`; Update — отказ для `IsSystem`; 404/403 иначе |
| A1-003 | P0 | `backend/internal/rbac/custom_roles.go:126-132` (+ `repository.go:153-156`) | Assign роли: `roleID` не сверяется с проектом участника | `SetMemberCustomRole` принимает любой UUID роли → права чужой роли применяются в другом проекте (escalation) | Перед assign: роль существует, `!IsSystem`, `role.ProjectID == projectID` |
| A1-004 | P0 | `backend/internal/audit/service.go:379-441` | GDPR hard-delete без защиты `founder` | Единственный founder на org можно удалить обычным admin → необратимая потеря ORG-04 роли | Запретить GDPR/ban/deactivate для `SystemRoleFounder` (кроме явного break-glass); тест denial |
| A1-005 | P1 | `backend/internal/rbac/service.go:47` | Founder не получает admin-bypass в RBAC | Bypass только для `SystemRoleAdmin`; founder без membership → 403 на project routes (ломает ORG-04) | Сравнивать через тот же критерий, что `auth.IsSystemAdmin` (founder \|\| admin) |
| A1-006 | P1 | `backend/internal/rbac/middleware.go:113-114` | `RequireAdmin` режет founder | `RequireSystemRole(admin)` используется backup и др.; founder получает 403 | Проверять founder\|\|admin (как audit/`IsSystemAdmin`) |
| A1-007 | P1 | `backend/internal/auth/magiclink.go:91-93` | Сбой Resend → HTTP 200 «Magic link sent», код не в ответе (prod) | Комментарий обещает код в логах, но логируется только email/error; пользователь и ops не получают код | Возвращать ошибку отправителю; либо логировать код в secure ops-канал; не маскировать fail успехом |
| A1-008 | P1 | `backend/internal/audit/consumer.go:50-54` (+ `97-100`) | JetStream Ack после failed `CreateAuditLog` | Событие снимается с очереди при ошибке БД → потеря audit trail, MaxDeliver не помогает | Ack только при успехе; иначе Nak/не Ack; пробрасывать err в span |
| A1-009 | P1 | `backend/internal/audit/service.go:283-309` | Ban/Deactivate без защиты founder | Любой admin может деактивировать founder (self-ban только) | Отклонять target с `SystemRoleFounder` |
| A1-010 | P2 | `backend/internal/auth/magiclink.go:128-129` | `redis.Del` после успешного verify без проверки ошибки | При сбое Del код остаётся → повторное использование одноразового кода | Проверять err Del; при fail — не считать login успешным / retry Del |
| A1-011 | P2 | `backend/internal/rbac/custom_roles.go:112-114` | N+1 clear members + игнор ошибок перед DeleteRole | `_ = ClearMemberCustomRoleByID` в цикле; при fail FK/остатки membership → частичный delete / битый state | Bulk `ClearCustomRole` WHERE role_id; fail → abort delete |
| A1-012 | P2 | `backend/internal/platform/outbox/worker.go:52-59` | Выборка pending без `FOR UPDATE SKIP LOCKED` | При ≥2 репликах один батч публикуется дважды (at-least-once без дедупа consumers) | Транзакция + SkipLocked; или claim status `processing` атомарно |
| A1-013 | P2 | `backend/internal/audit/service.go:140-141` (+ `handler.go:169-180`) | `UpdateResourceLimit` без allowlist ключей | Admin PUT любого `key` → Upsert в SystemSetting (в т.ч. `retention.*`) → неожиданная purge-политика | Allowlist известных limit/retention keys; 400 иначе |
| A1-014 | P3 | `backend/internal/rbac/custom_roles.go:89-90` | Ошибки invalidate-кэша после Update игнорируются (`_ =`) | Участники до TTL видят старые permissions | Логировать/ретраить Invalidate; при критичности — fail update |

## Чисто (проверено, проблем нет)

- Magic Link: `subtle.ConstantTimeCompare`, rate SET NX, rune-length в handler, Resend `LimitReader`
- Session refresh: hash compare constant-time; bulk invalidate sessions (не N+1)
- Auth admin handlers: `requireAdmin` через `IsSystemAdmin` (founder+admin)
- Audit HTTP admin group: `requireAdmin` + founder; Alertmanager webhook — Bearer + `ConstantTimeCompare`
- Platform SSRF guard (`safedialer`), AES-GCM crypto, cron recover+ctx cancel, ratelimit cleanup на ctx
- Outbox publish path: local recover, no `SaveX` panic; TraceContext inject/extract
- Pagination audit/trash: clamp page/per_page в repository/service (не off-by-one в offset)

## Оценка зоны: 5/10

Критичные security-дыры (pending→JWT, cross-project role IDOR/assign, GDPR founder) рядом с уже зрелыми паттернами (constant-time, rate limits, SSRF, admin на audit) — зона не «развал», но не production-safe без закрытия P0.
