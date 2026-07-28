# ADR-AUTH-SSO-2026: Корпоративный SSO — OIDC + LDAP/AD bind поверх Magic Link

- **Status:** accepted (2026-07-03, решения пользователя — `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` §2, батч 2)
- **Дата:** 2026-07-03
- **Автор:** arch · **Верифицируют:** arch-ver + security-auditor · **Личное ревью пользователя обязательно** (security/auth, правило 50)
- **Связанные:** GZ-SEC-03/04/05/06/08/09/10, `docs/tz_constraints.md` (Authentication), `ADR-AUTH-01` (Magic Link only — уточняется этим ADR), `ADR-ARCH-MICROSERVICES-2026` §5 (JWKS), `.planning/archive/roadmaps/MICROSERVICES-ROADMAP.md` §3
- **Уточняет:** MANDATORY-раздел «Аутентификация» ТЗ — SSO переводится из FORBIDDEN в дополнительный корпоративный метод

---

## Контекст

Hub — self-hosted платформа для команд 10–50 человек; ключевой сегмент — RU-корпораты on-prem, для которых вход через корпоративный каталог (Active Directory / корпоративный IdP) — барьер закупки (у Bitrix24 On-Premise LDAP есть). Текущее ТЗ: **Magic Link (email) — единственный метод входа**, SSO/LDAP — FORBIDDEN до отдельного ADR.

Пользователь одобрил SSO в scope (2026-07-03) и явно зафиксировал все ключевые решения в gray-zones батче 2. Этот ADR оформляет их в реализуемый контракт. Инварианты, которые **не меняются**: JWT access 15 мин, refresh в Redis 30 дней с instant revoke, rate limit auth 3/10 мин, 1 org на инстанс, Telegram — только notify bind.

## Рассмотренные варианты

1. **SAML 2.0 first** — стандарт энтерпрайза, но XML-подпись/каноникализация — классический источник CVE, тяжёлые библиотеки, избыточен для сегмента 10–50 человек. Отклонено для v1 (вернётся при спросе).
2. **Только OIDC** — покрывает Keycloak/Authentik/Google/Azure AD, но не покрывает «голый» on-prem AD без федерации — а это как раз RU-корпорат. Недостаточно.
3. **Только LDAP bind** — покрывает AD, но LDAP-пароль проходит через Hub (мы становимся верификатором паролей), нет SSO-опыта «один вход на все системы». Недостаточно.
4. **OIDC + LDAP/AD bind (принято)** — OIDC для организаций с IdP, LDAP bind как fallback для организаций с одним только AD. Два протокола закрывают весь целевой сегмент при умеренной сложности.

## Решение

### 1. Протоколы и режимы (зафиксировано пользователем)

| Решение | Содержание |
|---------|-----------|
| Протоколы v1 | **OIDC (Authorization Code + PKCE)** + **LDAP/AD bind** (форма логин/пароль → bind к AD). SAML — позже, отдельным дополнением |
| Org-политика | Поле `auth_policy` организации: `magic_link` (дефолт) / `sso` / `both` |
| Break-glass | При `auth_policy=sso` Magic Link остаётся доступен **только founder** — защита от локаута при падении IdP |
| Провижининг | **JIT при первом SSO-входе** (создание user по `email`+`sub`) + **периодический LDAP-sync деактиваций каждые 15 мин**; invite-модель сохраняется для non-SSO |
| MFA | **Собственного MFA нет** — фактор email (Magic Link) либо MFA на стороне IdP |
| Маппинг групп | AD/IdP-группы → **только `system_role`** (`admin`/`member`); `founder` НЕ назначается через SSO; project-роли — исключительно внутри Hub |

### 2. Изменения ERD

**User** (дополнение, `backend/ent/schema/user.go`):

```
sso_subject   string, Optional, Nillable         # `sub` claim OIDC или objectGUID AD
idp_id        UUID,   Optional, Nillable, FK → IdentityProvider
auth_source   enum { magic_link, oidc, ldap }, Default magic_link
Unique-индекс (idp_id, sso_subject)              # один subject — один user в рамках IdP
```

Матчинг при JIT: сначала по `(idp_id, sso_subject)`, затем по verified `email` (линковка существующего аккаунта с записью в audit); email-коллизия при несовпадении статуса → отказ с manual resolution админом.

**Organization** (дополнение, `backend/ent/schema/organization.go`):

```
auth_policy   enum { magic_link, sso, both }, Default magic_link
```

**IdentityProvider** (новая схема, admin-managed, максимум несколько записей):

```
id            UUID v4 PK
type          enum { oidc, ldap }
name          string                      # отображаемое имя кнопки входа
enabled       bool, Default false
config        JSON                        # OIDC: issuer_url, client_id, scopes, group_claim, redirect_uri
                                          # LDAP: host, port, base_dn, bind_dn, user_filter, group_map, starttls/ldaps
client_secret / bind_password             # AES-GCM at rest (platform/crypto), НЕ в JSON config
group_role_map JSON                       # {"CN=HubAdmins,...": "admin"} → system_role
version       int                         # optimistic lock
created_at / updated_at                   # UTC
```

Инварианты ТЗ соблюдены: UUID v4, `version`, UTC; секреты — AES-GCM (как BYOK-ключи в `ai`).

### 3. API (новые эндпоинты)

| Endpoint | Назначение |
|----------|-----------|
| `GET  /api/auth/providers` | Публичный: список включённых методов входа для рендера login-страницы (учитывает `auth_policy`) |
| `GET  /api/auth/oidc/login?provider_id=` | Редирект на IdP: генерирует `state`+`nonce`+PKCE `code_verifier`, сохраняет в Redis (TTL 10 мин, one-time) |
| `GET  /api/auth/oidc/callback` | Обмен code→token, валидация `state`/`nonce`/`iss`/`aud`/подписи, JIT-провижининг, выдача стандартной пары JWT+refresh |
| `POST /api/auth/ldap/login` | `{login, password}` → LDAP bind → JIT → пара токенов; rate limit 3/10 мин по IP (общий auth-лимитер) |
| `GET/POST/PUT/DELETE /api/admin/identity-providers` | CRUD конфигурации IdP (RBAC: только admin/founder); `POST /:id/test` — проверка соединения без сохранения сессии |

Формат ошибок `{"error","request_id"}`, коды 400/401/403/409/422/429 — по NFR-13. После callback сессия и refresh-цикл **идентичны** Magic Link-потоку (единый `service_session.go`) — SSO меняет только способ первичной аутентификации.

### 4. Библиотеки

| Библиотека | Задача | Обоснование |
|------------|--------|-------------|
| `github.com/coreos/go-oidc/v3` | OIDC discovery, верификация ID-token (JWKS, iss/aud/exp/nonce) | Де-факто стандарт Go-экосистемы, CNCF-происхождение, активная поддержка, минимальные зависимости; альтернатива `zitadel/oidc` тяжелее (полный фреймворк RP+OP — нам нужен только RP) |
| `golang.org/x/oauth2` | Authorization Code + PKCE flow | Каноничная реализация, зависимость go-oidc |
| `github.com/go-ldap/ldap/v3` | LDAP bind, поиск, чтение групп | Единственная зрелая maintained Go-LDAP библиотека; поддерживает LDAPS/StartTLS, paged search для sync-джобы |

Обе библиотеки — pure Go, без CGO, совместимы с Go 1.25; HTTP-клиент OIDC — через `safehttp` c таймаутом (issuer URL задаёт админ → SSRF-контроль обязателен).

### 5. Security-требования (вход в реализацию, проверяет security-auditor)

1. **OIDC:** `state` (CSRF) + `nonce` (replay) + **PKCE S256** — все три обязательны; one-time значения в Redis с TTL 10 мин; строгая валидация `iss`/`aud`/`exp`/подписи через go-oidc; redirect_uri — только зарегистрированный (никаких open-redirect по `?next=`).
2. **LDAP:** только **LDAPS или StartTLS** (plaintext 389 без TLS запрещён в prod); экранирование фильтров (`ldap.EscapeFilter`) от LDAP-injection; пустой пароль → немедленный отказ (анонимный bind AD «успешен» с пустым паролем — известная ловушка); bind-ошибки не различаются в ответе (нет user enumeration).
3. **Session fixation:** после SSO-входа — только новая пара токенов (никакого повторного использования сессионного материала); refresh-цикл общий с Magic Link.
4. **Rate limiting:** `/auth/oidc/*` и `/auth/ldap/login` под общим auth-лимитером 3/10 мин по IP; `POST /admin/identity-providers/:id/test` — 10/мин.
5. **Секреты:** `client_secret`/`bind_password` — AES-GCM at rest, маскирование в API-ответах (`***`), запрет логирования; сравнение секретов — `subtle.ConstantTimeCompare` где применимо.
6. **Деактивация:** LDAP-sync (15 мин) при обнаружении disabled/deleted аккаунта — деактивация user + **instant revoke** всех refresh-токенов из Redis; событие в audit.
7. **Break-glass:** проверка «founder может Magic Link при enforce-SSO» — на уровне service (не UI); попытка Magic Link не-founder'ом при `auth_policy=sso` → 403 + audit-событие.

### 6. Тест-требования

- Пакет auth (включая новый SSO-код) — **security-critical ≥80%** (текущий auth 82.4% — не уронить).
- Тесты по образцу security-пакетов SEC-06/07: negative-first — подделка `state`, повтор `nonce`, истёкший code, невалидная подпись ID-token, LDAP-injection в login, пустой пароль, Magic Link при enforce-SSO (member → 403, founder → OK), JIT-коллизия email, revoke после sync-деактивации.
- Contract-тесты OIDC — через мок-IdP (httptest + подписанные JWT), LDAP — контейнер openldap в Docker-регрессии (`-tags hubtest`).
- E2E: happy path OIDC-входа (мок-IdP) + переключение `auth_policy` админом.

### 7. Влияние на tz_constraints (MANDATORY-раздел)

`docs/tz_constraints.md` и `.cursor/rules/60-tz-constraints.mdc` уже помечены (2026-07-03) — этот ADR фиксирует финальную формулировку:

- **Аутентификация / Единственный вход** → «**Magic Link (email) — дефолтный метод**; SSO (OIDC + LDAP/AD bind) — дополнительный корпоративный метод по org-политике `magic_link/sso/both`; при `sso` Magic Link = break-glass только для founder (ADR-AUTH-SSO-2026)».
- **FORBIDDEN:** строка «SSO/LDAP» удаляется; остаётся «OAuth-провайдеры соцсетей — вне scope».
- Неизменны: JWT 15 мин, refresh Redis 30 дней + instant revoke, rate limit 3/10 мин, Telegram только notify, invite-модель для non-SSO.

### 8. Риски и митигации

| Риск | Митигация |
|------|-----------|
| Локаут org при падении/misconfig IdP | Break-glass Magic Link для founder; `POST /:id/test` до включения; `enabled=false` не ломает активные сессии |
| Дубли/угон аккаунта при JIT email-матчинге | Матч по `(idp_id, sso_subject)` первичен; линковка по email — только verified + audit; конфликт → manual resolve |
| LDAP-канал как вектор (injection, plaintext) | LDAPS/StartTLS обязателен, EscapeFilter, отказ на пустой пароль, bind под сервисным аккаунтом с минимальными правами |
| Приведение к admin через group-mapping | Маппинг только на admin/member (не founder); изменение `group_role_map` — audit-событие; пересчёт роли при каждом входе |
| SSRF через issuer_url админа | `safehttp` + `ValidateOutboundURL` для discovery/JWKS-запросов |
| Дрейф деактиваций (уволенный сотрудник) | Sync 15 мин + instant revoke; событие `user.deactivated` в NATS для каскадов |
| Рост поверхности auth перед выносом сервисов | SSO живёт в hub-core (auth остаётся в ядре по ADR-MICROSERVICES); JWKS-миграция HS256→asymmetric не блокируется |

### 9. План фаз реализации

| Фаза | Содержание | Выход |
|------|-----------|-------|
| SSO-0 Spec | sa: FR/ERD/контракты по этому ADR; qa-manual: TC-таблица (полная, GZ-QA-14); designer: login-страница с провайдерами | verified спека |
| SSO-1 Core OIDC | Ent-миграции (User/Organization/IdentityProvider), `auth/oidc.go` (login/callback, state/nonce/PKCE), JIT, admin CRUD IdP | OIDC-вход за feature-флагом |
| SSO-2 LDAP | `auth/ldap.go` (bind-вход), sync-джоба деактиваций 15 мин (`platform/cron`), group→system_role маппинг | LDAP-вход + sync |
| SSO-3 Policy + break-glass | `auth_policy` enforcement, break-glass founder, login UI (кнопки провайдеров по `GET /auth/providers`) | Полный флоу 3 политик |
| SSO-4 Hardening | security-auditor прогон по §5, negative-тесты, E2E, coverage-gate ≥80%, docs sync (system_specification, wiki/API, security_guidelines, tz_constraints финал) | Приёмка + личное ревью пользователя |

Зависимости: не блокирует и не блокируется Track P (auth в hub-core); желательно до JWKS-миграции (§5 ADR-MICROSERVICES) либо совместно с ней.

## Последствия

- **+** Снимается барьер закупки RU-корпоратов (AD-вход); JIT убирает ручной invite-цикл для SSO-org; централизованная деактивация из каталога ≤15 мин; MFA бесплатно делегируется IdP; ядро auth-флоу (JWT/refresh/revoke) не трогается.
- **−** Поверхность атаки auth растёт (два новых протокола) → митигируется §5 + обязательным security-audit и личным ревью; появляется зависимость доступности от внешнего IdP → break-glass; +2 библиотеки в go.mod; LDAP bind означает, что пароль проходит через Hub (только TLS, не хранится).
- **Долг:** SAML отложен (возврат при спросе — отдельным дополнением к этому ADR); SCIM-провижининг не делаем (JIT+sync достаточно для 10–50 человек).

## Проверка соответствия (для arch-ver / security-auditor)

`ent/schema/user.go` содержит `sso_subject/idp_id/auth_source` + unique-индекс; `organization.go` — `auth_policy`; `internal/auth/` содержит oidc/ldap-модули без утечки `*ent.Client` в service (repository-паттерн); `handler.go` регистрирует `/auth/oidc/*`, `/auth/ldap/login`, `/auth/providers`; rate-limiter покрывает новые роуты; тесты в `backend/testframework/auth/` покрывают negative-набор §6; в `tz_constraints.md` FORBIDDEN не содержит SSO.
