# Security (cross-cutting)

**Дата:** 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Зона A9:** authn/authz · CORS/CSRF/cookies · JWT · rate limit · Magic Link / no passwords · IDOR · SSRF · upload/path traversal · secrets  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Контекст волны 1:** `.planning/audits/grok-2026-07/01-*.md` … `08-*.md`

## Что проверено

| Область | Источник в коде |
|---------|-----------------|
| Middleware chain | `wiring_fiber.go`, `middleware/{cors,security_headers,ratelimit,logging,requestid}.go` |
| Auth wiring | `wiring_auth.go`, `auth/{handler,middleware,magiclink,service_login,service_session}.go` |
| JWT / refresh | `service_session.go` (HS256, hash-at-rest, rotation), `config.go` TTL + prod secret validate |
| RBAC Fiber v3 order | `tracker/handler.go`, `tms/handler.go`, `github/handler.go`, `rbac/handler.go`, `sprint/handler.go` |
| Webhooks | LiveKit `calls/handler.go`, GitHub/GitLab handlers, Alertmanager `audit/handler_alerts.go`, TMS webhook routes |
| WS | `ws/handler.go` (CheckOrigin, JWT query), `wiring_ws.go` Authorizer |
| SSRF | `platform/http/safedialer.go` + grep `ValidateOutboundURL` / `http.DefaultClient` |
| Files | `storage/{handler,filename,service}.go` |
| Secrets / env | `deployments/.env.example`, `livekit.yaml`, `docker-compose.prod.yml`, `deploy.yml` (no `DEV_MODE`) |
| Password / Telegram login | grep routes + `auth/handler.RegisterRoutes` (нет `/telegram`, нет password) |
| Cookies / CSRF | grep `Cookie`/`SetCookie`/`CSRF` — отсутствуют |
| Волна 1 P0 security | A1 pending/JWT+role IDOR; A2 search; A3 sprint/issue; A5 imex/GH; A6 CRM |

Grep-скан §security-auditor: `http.DefaultClient` в prod — нет; Fiber RBAC «middleware перед handler» — не найдено; `DEV_MODE` в `deploy.yml` — нет.

## Таблица роутов (authz) — ключевые / рискованные

| Метод / путь | Требуемая роль / auth | Фактическая проверка | Вердикт |
|--------------|----------------------|----------------------|---------|
| `POST /api/auth/magic-link` | public + rate limit | нет JWT; Redis RL 3/10m (prod); email validate | OK (public) |
| `POST /api/auth/magic-link/verify` | public + tighter RL | verify + **CreateSession без status gate** | **FAIL** — pending→JWT |
| `POST /api/auth/refresh` | refresh token | Redis hash + session validate + rotate | OK |
| `GET/PUT /api/auth/me`, sessions | authenticated | `auth.Group("", authMw)` | OK |
| `GET/PUT /api/admin/registration-mode`, pending users, approve/reject | system admin/founder | `authMw` + `requireAdmin` → `IsSystemAdmin` | OK |
| `GET /api/admin/organization/invite` | admin | `requireAdmin` в handler | OK |
| `POST /api/calls/webhook` | LiveKit JWT+sha256 | `verifyWebhookSignature` mandatory sha256 | OK |
| `POST /api/github/webhook` | HMAC any-of secrets | signature required; **secret≠repo binding** | **WEAK** (A5-008) |
| `POST /api/gitlab/webhook` | X-Gitlab-Token any-of | token required; **token≠path binding** | **WEAK** (A5-009) |
| `POST /api/webhooks/alertmanager` | Bearer secret | `ConstantTimeCompare`; route only if secret set | OK |
| `POST /api/projects/:id/tms/webhook*` | JWT + `PermTestRunCreate` | group `authMw` + RBAC после handler | OK (док про X-Hub-Token устарела) |
| `WS /ws?token=` | JWT access | ParseAccessToken + Authorizer.CanSubscribe | OK authz; token-in-query risk |
| `POST/GET/DELETE /api/files/*` | JWT + project perms | `authMw` + `authorizeProject` via `FileProjectID` | OK |
| `POST …/upload-url` | `PermFileUpload` | RBAC on body `projectId` | OK |
| Tracker `…/issues` CRUD | project perms | Fiber v3: handler, then `RequirePermission` | order OK; **issueKey vs projectId IDOR** |
| `…/roles/:roleId` Update/Delete/Assign | `PermProjectManageRoles` | RBAC on path project; **roleId not scoped** | **FAIL** |
| `…/import|export/*` (imex) | должен быть project RBAC | только `authMw` | **FAIL** |
| `…/github|gitlab/integrations` | manage_workflow / view | RBAC order OK; delete/trigger IDOR | **WEAK** |
| CRM contacts/deals mutate | owner/admin (ожидание) | JWT only; **нет owner check** | **FAIL** |
| `/api/admin/backups`, `/admin/ai`, `/admin/mail`, audit admin | admin | `authMw` + admin MW / `requireAdmin` | OK |
| `/api/docs`, `/api/docs/doc.json` | — | **без auth** | info disclosure |
| `GET /metrics` | CIDR allowlist (опц.) | пустой `METRICS_ALLOWED_CIDRS` → **open** | WEAK default |
| `/api/dev/*` | DEV_MODE | runtime guard; **stripped в `-tags release`** | OK prod |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A9-001 | P0 | `auth/service_login.go:64` (+ `auth/middleware.go:128`) | После Magic Link verify всегда `CreateSession`; `userIsDeactivated` считает только `deactivated`, не `pending` | Pending-пользователь получает JWT и полный API до approve (ORG-06 bypass). = A1-001 | Не выдавать сессию если `status != active`; RequireAuth/Refresh fail-closed на pending |
| A9-002 | P0 | `rbac/custom_roles.go:73-84` (+ `handler.go:99-114`) | Update/Delete роли по `roleId` без `role.project_id == :projectId` | Member с manage_roles на проекте A мутирует роль проекта B / system role. = A1-002 | Scope Get/Update/Delete по project; Update — отказ для `IsSystem` |
| A9-003 | P0 | `rbac/custom_roles.go:126-132` | Assign: `roleID` не сверяется с проектом участника | Privilege escalation: чужая роль → другой проект. = A1-003 | Перед assign: роль ∈ project, `!IsSystem` |
| A9-004 | P0 | `crm/service.go:80-123,172-214` (+ handlers) | Update/Delete contact/deal без owner/admin | Любой JWT правит/удаляет PII и deals (+ won→invoice). = A6-001 | `authorizeContact`/`authorizeDeal` (owner\|\|admin) |
| A9-005 | P0 | `sprint/service.go:107-149` (+ repo) | Sprint ops без сверки `sprint.project_id` с URL | Complete чужого спринта + перенос issues в свой проект. = A3-001 | Get/mutate sprint `Where(projectID)` |
| A9-006 | P0 | `messenger/service_search.go:11-29` (+ `mongo_repository.go:282-288`) | Поиск сообщений без membership / user-scope | Чтение текстов всех каналов (NF-03 класс). = A2-001 | Фильтр по `IsMember` / channel allowlist |
| A9-007 | P1 | `imex/handler.go:26-35` | Import/export только `authMw`, без `RequirePermission` | Auth user → CSV/ZIP/PDF чужого проекта. = A5-002 | `PermProjectView` / `PermTaskCreate` как tracker |
| A9-008 | P1 | `github/handler.go:119-126` + `service` verify; `gitlab/handler.go:99-102` | Webhook: совпадение **любого** секрета инстанса, репозиторий из payload | Утечка secret A → подпись событий в интеграцию B. = A5-008/009 | Verify секретом **найденной** интеграции |
| A9-009 | P1 | `tracker`/`tms`/`github`/`gitlab` handlers+services (см. A3/A5) | Resource lookup по key/UUID без project scope | Классический IDOR внутри org при membership в «чужом» path-проекте | Scoped queries `projectID + key/id`; 404 mismatch |
| A9-010 | P1 | `rbac/service.go:47`; `rbac/middleware.go:113-114` | Founder не в admin-bypass / `RequireAdmin` режет founder | ORG-04: founder без membership → 403; backup и др. недоступны. = A1-005/006 | Тот же критерий, что `auth.IsSystemAdmin` |
| A9-011 | P1 | `audit/service.go:379-441,283-309` | GDPR hard-delete / ban без защиты founder | Admin удаляет/деактивирует единственного founder. = A1-004/009 | Deny target `SystemRoleFounder` |
| A9-012 | P1 | `apigateway/apikey.go:189-203` vs `auth/middleware.go:59-98` | `APIKeyAuth` пишет `userID`, но `RequireAuth` всё равно требует Bearer JWT | X-API-Key alone не аутентифицирует защищённые роуты; док TMS про token вводит в заблуждение → риск обходных JWT в CI | RequireAuth: если `Locals("userID")` уже set API key — skip JWT; или отдельный CI middleware |
| A9-013 | P1 | `calls/service.go` / recording / chat (см. A2-002…006) | Calls/recording/chat без participant/channel authz | Authenticated user управляет чужими звонками/записями/чатом | Membership/participant checks до LiveKit/DB |
| A9-014 | P2 | `wiring_fiber.go:57-59` + `ratelimit.go:78-95` | `RateLimitByUser` на `/files` и `/messenger` **до** `authMw` | `userID` ещё пуст → фолбэк на IP; TZ «10/min user / 60/min user» не enforced per-user | Вешать user RL **после** auth group, либо читать JWT claims в limiter |
| A9-015 | P2 | `observability/prometheus.go:72-88` + `wiring_fiber.go:46` | `GET /metrics` без CIDR при пустом allowlist | Утечка latency/RPS/labels наружу (reconnaissance) | В prod требовать `METRICS_ALLOWED_CIDRS` (fail-start как CORS) |
| A9-016 | P2 | `middleware/security_headers.go:7-14` | Нет `Strict-Transport-Security` | Зависит только от Caddy; прямой hit backend без HSTS | Добавить HSTS когда TLS terminated / конфиг flag |
| A9-017 | P2 | `storage/service.go:84-88` | Presigned PUT без ContentLength/max (NF-09) | Обход Fiber BodyLimit 25MB → SeaweedFS DoS/quota | Presign constraint / lifecycle / accepted-risk doc owner |
| A9-018 | P2 | `notification/push.go:230-261` | Push subscribe hijack / unsubscribe без user scope | DoS уведомлений жертвы. = A2-005 | Update/Delete `Endpoint AND UserID` |
| A9-019 | P3 | `auth/service_login.go:94-115` (+ dto Telegram) | `LoginWithTelegram*` жив в service, роут снят | Dead attack surface: случайный re-wire нарушит ADR-AUTH-01 | Удалить login-path или `# deprecated` + test «no route» |
| A9-020 | P3 | `ws/handler.go:140-156` | Access JWT в query `?token=` | Утечка в access/proxy logs. = A2-013 | One-time WS ticket |
| A9-021 | P3 | `wiring_fiber.go:63-68` | `/api/docs` + `doc.json` публичны | Полная карта эндпоинтов без auth | Auth или отключать в prod |
| A9-022 | P3 | `deployments/livekit.yaml:10-12` | Local LiveKit keys в git (dev) | Ок для local; риск copy-paste в prod | Уже есть `.example` + PG-03; не деплоить этот файл |
| A9-023 | P3 | `docker-compose.prod.yml:172,237,261` | Default image tag `:latest` | Unpinned pull при пустом IMAGE_* | Pin digest/tag в prod defaults |

## Чисто

- **Magic Link only:** `RegisterRoutes` — только magic-link / refresh / me / admin; password routes отсутствуют; Telegram login **не** зарегистрирован.
- **JWT:** `SigningMethodHS256` issue; parse требует `*jwt.SigningMethodHMAC` (alg confusion none/RS256 закрыт); TTL 15m / refresh 30d в `config.go:171-174`; prod запрещает dev JWT secret (`config.go:223-230`).
- **Refresh:** SHA-256 at rest; rotation store-new→update→delete-old (`service_session.go:229-261`); revoke sessions + Redis Del.
- **Constant-time:** magic-link (`magiclink.go:123`), LiveKit iss (`calls/handler.go`), Alertmanager Bearer, webhook HMAC paths.
- **CORS:** allowlist, reject `*`, fail-start без `CORS_ORIGINS` в non-dev (`cors.go:30-41`); `AllowCredentials: true`.
- **CSRF / cookies:** Bearer-only, Set-Cookie не используется → CSRF токены moot (NF-14 accepted).
- **Security headers:** nosniff, `X-Frame-Options: DENY`, CSP `default-src 'self'`, Referrer-Policy.
- **Rate limits (настроены):** global 100/s IP; auth Redis 3/10m + verify email 10/10m (prod); AI 20/min; BodyLimit 25MB.
- **SSRF:** `ValidateOutboundURL` + `NewSafeHTTPClient` на GitLab/GitHub client, WebPush, OpenGraph, AI adapters, apigateway outbound, 1C; `http.DefaultClient` в prod internal — нет.
- **LiveKit webhook:** sha256 claim mandatory (SEC-02).
- **Storage path traversal:** `sanitizeFilename` (`filename.go:11-34`); S3 key `{project}/{fileID}/{sanitized}`; download/delete RBAC via DB projectId (CR-04).
- **Fiber v3 RBAC order:** tracker/tms/github/gitlab/sprint/roles — middleware **после** handler в аргументах.
- **Dev routes:** `!release` build tag; release no-op (`dev/handler_release.go`).
- **Secrets in `.env.example`:** placeholders `CHANGE_ME_*` / `re_xxx…` / sample telegram format — не live secrets; `deploy.yml` без `DEV_MODE=true`.
- **Trusted proxies:** CIDR private ranges (`wiring_fiber.go:22-28`).

## Оценка зоны: 5/10

**Обоснование.** Cross-cutting **контролы** (Magic Link, JWT/HMAC, refresh rotation, CORS fail-start, SafeHTTP, LiveKit sha256, storage sanitize+RBAC, release-stripped `/dev`, Fiber RBAC order) выглядят на **7–8/10**. Оценка зоны тянется вниз системным **Broken Access Control**: pending→JWT (P0), custom-role IDOR (P0), и повторяющийся класс «authz на path-проект / auth-only, resource по UUID/key без scope» (sprint, CRM, messenger search, imex, GH/GL, calls) — это A01, блокеры релиза. После закрытия A9-001…A9-013 зона поднялась бы к ~7.5–8/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 6 |
| P1 | 7 |
| P2 | 5 |
| P3 | 5 |
| **Всего** | **23** |

**Score: 5/10**

`verified: true` — 2026-07-26 (readonly audit; 0 CRITICAL/HIGH **закрытых** в этой сессии — открытые P0/P1 = release blockers).
