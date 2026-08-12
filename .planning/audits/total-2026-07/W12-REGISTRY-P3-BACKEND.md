# W12 — реестр уникальных P3-классов (backend, волны 1–8)

> TASK-008 · Стадия B · W12 · 2026-07-28  
> Источники: `A{1-8}-*.md`, сводки `W1-SUMMARY`…`W8-SUMMARY`, [`00-INDEX.md`](00-INDEX.md)  
> Scope: **только backend** (W1–W8). Frontend W9–W11 — вне артефакта.  
> **TASK-\*.md не создаются** — только реестр + предложение группировки.  
> Формат: зеркало [`W12-REGISTRY-P2-BACKEND.md`](W12-REGISTRY-P2-BACKEND.md).

## Сводка

| Метрика | Значение |
|---------|---------:|
| Сырых P3 (W1–W8) | **164** |
| Уникальных классов | **23** |
| Из них с критерием мутации «да» | **18** |
| В секции «требует уточнения» | **5** |
| Предлагаемых задач бэклога | **12** |

Сырые по волнам: W1=18 · W2=16 · W3=18 · W4=23 · W5=28 · W6=21 · W7=28 · W8=12 → **164**  
(полный Stage A P3=245 = 164 backend + 81 frontend W9–W11).

P0/P1/P2 W1–W8 **не** в таблице (отдельные реестры): [`W12-REGISTRY-P0.md`](W12-REGISTRY-P0.md), [`W12-REGISTRY-P1-BACKEND.md`](W12-REGISTRY-P1-BACKEND.md), [`W12-REGISTRY-P2-BACKEND.md`](W12-REGISTRY-P2-BACKEND.md).

---

## Уникальные классы дефектов

| ClassID | Severity | Title | Source IDs | Primary file:line | Mutation criterion | Suggested TASK slug |
|---------|----------|-------|------------|-------------------|--------------------|---------------------|
| P3-BE-01 | P3 | Silent swallow: `_=` Publish/Count/Del/notifier → zero/`continue` без slog | W1-A2-013, W1-A3-008, W1-A4-010, W1-A4-011, W1-A4-012, W1-A5-009, W1-A5-010, W2-A2-020, W2-A4-013, W3-A3-010, W5-A2-010, W6-A4-013, W6-A4-015, W7-A1-010, W7-A2-015, W7-A3-011, W7-A4-019, W7-A5-007, W7-A5-015 | `backend/internal/tracker/repo_issues.go:376` | **да** — inject Publish/Count err → slog/metric или 500; silent 0/continue → тест краснеет | `be-p3-silent-error-swallow` |
| P3-BE-02 | P3 | Assert-free / слабый статус / дубли empty / misleading Test* names | W1-A3-011, W1-A4-016, W2-A5-010, W3-A1-012, W3-A3-011, W4-A2-008, W4-A4-011, W5-A5-005, W5-A5-006, W6-A2-004, W6-A2-005, W6-A3-015, W6-A4-014, W6-A5-007, W6-A5-008, W6-A5-009, W7-A1-009, W7-A4-020 | `backend/testframework/sprint/handler_wave113_test.go:30` | **да** — сломать ветку (Ack/status/branch) → точный assert краснеет; assert-free suite остаётся зелёным | `be-p3-assert-free-tests` |
| P3-BE-03 | P3 | Нет denial / mutation-proof тестов для известных P1/P2 guards | W1-A4-014, W1-A4-015, W2-A5-009, W3-A3-012, W3-A5-015, W3-A5-016, W4-A1-010, W4-A2-007, W4-A3-009, W4-A3-010, W4-A5-010, W4-A5-011, W5-A1-011, W5-A2-013, W5-A2-014, W5-A2-015, W5-A2-016, W5-A2-017, W5-A3-016, W5-A3-017, W5-A3-018, W5-A3-019, W5-A3-020, W5-A3-021, W5-A4-012, W5-A5-007, W6-A5-006, W6-A5-010, W7-A5-016 | `backend/testframework/calls/handler_routes_test.go:42` | **да** — снять guard (poker RBAC / soft-delete / OR JQL / api_key redact) → новый Ent/HTTP-negative краснеет | `be-p3-mutation-proof-gaps` |
| P3-BE-04 | P3 | Truncate / limit по байтам (mid-rune UTF-8) | W2-A2-019, W3-A2-017, W5-A4-010, W7-A3-012, W7-A5-009 | `backend/internal/search/service.go:621` | **да** — строка с emoji/кириллицей → truncate по рунам; `s[:N]` байты → invalid UTF-8 / красный тест | `be-p3-utf8-rune-truncate` |
| P3-BE-05 | P3 | Repo defense: mutate/list без ProjectIDEQ / DeletedAtIsNil | W1-A3-009, W1-A4-013, W2-A3-009, W2-A3-010, W3-A3-008, W3-A5-014, W7-A4-014, W7-A5-011 | `backend/internal/tracker/repo_automation.go:55` | **да** — Update/Delete soft-deleted / чужой project UUID → 0 rows/404; без Where → 200 | `be-p3-repo-project-softdelete` |
| P3-BE-06 | P3 | Unbounded List residual (activity/comments/bases/CRM) | W1-A3-010, W2-A5-012, W7-A2-013 | `backend/internal/tracker/repo_activity.go:42` | **да** — N>cap → Limit/`has_more`; снять Limit → unbounded / timeout | `be-p3-list-caps-residual` |
| P3-BE-07 | P3 | Layering / dead code / unused params / CreateTable без tx | W1-A1-008, W1-A2-014, W1-A2-015, W2-A2-018, W2-A4-010, W2-A4-011, W2-A5-007, W3-A2-018, W3-A3-009, W4-A3-007, W4-A4-010, W4-A5-009, W6-A1-009, W7-A4-017 | `backend/internal/tracker/handler_issues.go:285` | **да** — JQL только через service; CreateTable fail column → rollback; мёртвый reader/param удалён (compile/test) | `be-p3-dead-code-layering` |
| P3-BE-08 | P3 | Сырой `err.Error()` / misleading Forbidden text / fragile Contains | W2-A1-011, W5-A1-008, W5-A4-011, W6-A4-011, W7-A2-011, W7-A5-008 | `backend/internal/tms/handler_webhook.go:109` | **да** — parser/DB err → стабильный `apperr` без internals; Forbidden≠`"admin access required"` для non-admin path | `be-p3-safe-error-messages` |
| P3-BE-09 | P3 | Input hygiene: negative limit, ColType, version=0, nil UUID, category, invite link, format case | W1-A5-011, W2-A4-012, W2-A5-008, W2-A5-011, W3-A1-011, W5-A1-009, W7-A5-010 | `backend/internal/messenger/handler.go:27` | **да** — `limit=-5` → default; `uuid.Nil`→400; `version=0`→400; `.CSV`→detect; invite Link с frontend base | `be-p3-input-hygiene` |
| P3-BE-10 | P3 | WS/calls residual: JWT `?token=`, empty rooms, bridge Stop, deny frame, nil-RBAC recordings | W3-A4-006, W3-A4-007, W3-A4-008, W3-A4-009, W3-A4-010, W3-A5-013 | `backend/internal/ws/handler.go:153` | **да** — nil rbac → 500; empty room deleted from map; Start fail → Stop; ExportRooms под mu | `be-p3-ws-calls-lifecycle` |
| P3-BE-11 | P3 | Platform HTTP residual: OG image SSRF, ValidateOutbound IP-literal, CORS origin shape | W4-A1-008, W4-A1-009, W4-A1-011 | `backend/internal/platform/http/opengraph.go:103` | **да** — `og:image` private/`file:` → cleared; `http://169.254…` на Validate → ErrBlocked; bad CORS origin → reject | `be-p3-platform-http-hardening` |
| P3-BE-12 | P3 | Platform shutdown/obs: NATS Drain, outbox ctx, baggage, Timed Info, metrics cardinality, probe ctx | W4-A2-006, W4-A2-009, W4-A3-005, W4-A3-006, W4-A3-008, W4-A4-012 | `backend/internal/platform/nats/client.go:70` | **да** — Close→Drain; cancel mid-batch stops; propagator без Baggage; unknown operation→`"other"`; Ready uses `c.Context()` | `be-p3-platform-shutdown-obs` |
| P3-BE-13 | P3 | AI residual: admin BYOK cross-scope, provider routing, nil lookup, EnsureBotUser race | W5-A1-010, W5-A2-012, W5-A3-013, W5-A3-014, W5-A3-015 | `backend/internal/ai/admin_handler.go:74` | **да** — admin Update user-scope → 404; `IsYandex` только host allowlist; nil tracker → error; parallel EnsureBot → 1 user | `be-p3-ai-scope-routing` |
| P3-BE-14 | P3 | Auth/RBAC/audit residual: onboarding tx, UnassignRole, Invalidate race, CTC length, Upsert race | W6-A1-010, W6-A3-012, W6-A3-014, W6-A4-010, W6-A4-012 | `backend/internal/rbac/handler.go:212` | **да** — fail mark-complete → profile rollback; wrong `:roleId`→404; concurrent Upsert → 1 row; CTC fixed-size digest | `be-p3-auth-rbac-audit` |
| P3-BE-15 | P3 | V4/storage/dashboard/imex residual: flaky DSN, coffee email PII, news audit, orphan loop, layout ACL | W7-A2-012, W7-A2-014, W7-A2-016, W7-A3-013, W7-A4-015, W7-A4-016, W7-A4-018 | `backend/internal/intranet/coffee_dto.go:20` | **да** — partner DTO без email; soft-delete news → audit row; orphan >500 drained; ApplyTemplate stranger→403 | `be-p3-v4-storage-dashboard` |
| P3-BE-16 | P3 | Dev: login без StatusActive; seed partial 200; Telegram-login demo copy | W7-A5-012, W7-A5-013, W7-A5-014 | `backend/internal/dev/handler.go:93` | **да** — banned user `/api/dev/login` → 403; seed fail → non-2xx/`partial`; demo без «Telegram вход» | `be-p3-dev-login-seed` |
| P3-BE-17 | P3 | Ent hygiene: duplicate Unique index, mentions `[]string`, MaxLen code/body | W8-A1-009, W8-A2-019, W8-A3-017, W8-A3-018 | `backend/ent/schema/comment.go:29` | **да** — `mentions` JSON `[]uuid.UUID`; oversize code/body → validation; один unique source на field | `be-p3-ent-hygiene-maxlen` |
| P3-BE-18 | P3 | Ent mojibake / «ибра» godoc comments (TASK-007 residual) | W8-A1-010, W8-A2-020, W8-A3-016, W8-A4-011, W8-A5-003, W8-A5-007, W8-A5-011 | `backend/ent/schema/apikey.go:48` | **да** — grep `ибра`/`Р` в `ent/schema` = 0; «Рёбра (edges)» | `be-p3-ent-mojibake-comments` |

**Классы с мутацией «да»: 18** (P3-BE-01…18).  
**Всего уникальных классов: 23** (= 18 + 5 в секции уточнения).

---

## Требует уточнения

Находки/классы, для которых **нельзя** завести TASK по правилу 57 (нет однозначного критерия мутации / нужен ADR или product-scope). После уточнения — вернуть в очередь задач.

| ClassID | Source IDs | Суть | Чего не хватает для мутации / задачи |
|---------|------------|------|--------------------------------------|
| P3-BE-U01 | W4-A1-007, W4-A5-012 | HSTS в app vs делегирование Caddy; docs CTC drift outbound webhook | Product: кто ставит HSTS. Docs-only правка vs код. Без решения «правильный» тест неоднозначен |
| P3-BE-U02 | W4-A4-013 | AES-GCM `additionalData=nil`; `MaybeDecrypt` legacy soft-passthrough | ADR migration: AAD=`entityID` ломает существующие ciphertext. Нужен rollout план |
| P3-BE-U03 | W5-A2-009, W5-A2-011 | Слабый content-policy denylist; `pref==nil` → `allowExternal=true` | Пересекается с P2-BE-17 (privacy default). Policy strength = product/SEC scope, не точечный bugfix |
| P3-BE-U04 | W6-A1-007, W6-A1-008, W6-A3-013 | Telegram login leftovers в auth comments/repo; RBAC `RequireAuth` misnomer | Рядом с P2-BE-U01 / P1-BE-U02. Решение: delete vs quarantine. Misnomer rename — без break callers |
| P3-BE-U05 | W8-A4-012 | `ScheduledCall` soft-delete без `version` | Остаток P1-BE-U06 VersionMixin. Без exclusion list задача бесконечна |

Покрытие сырых ID: все **164** входят ровно в один класс из P3-BE-01…18 или P3-BE-U01…U05.

Проверка:

| Группа | Σ сырых |
|--------|--------:|
| BE-01…03 (swallow / assert-free / mutation-proof) | 66 |
| BE-04…07 (rune / repo / lists / dead-code) | 30 |
| BE-08…10 (errors / input / ws-calls) | 19 |
| BE-11…14 (platform / AI / auth-rbac-audit) | 19 |
| BE-15…18 (v4-storage / dev / ent) | 21 |
| **Σ с мутацией** | **155** |
| U01…U05 | 9 |
| **Итого** | **164** |

---

## Предлагаемая группировка в задачи бэклога (~12)

Не создавать файлы сейчас — только карта ClassID → будущий TASK slug. Нумерация TASK после закрытия P2-волны (свободный слот — у team-lead).

| # | Suggested slug | Priority | ClassIDs | Сырых ID (оценка) | Заметка |
|---|---------------|----------|----------|------------------:|---------|
| 1 | `be-p3-silent-error-swallow` | P3 | BE-01 | 19 | `_=` / Count→0 / Del без err — slog+metric или fail-closed |
| 2 | `be-p3-assert-free-tests` | P3 | BE-02 | 18 | Точные status/Ack asserts; убрать assert-free / дубли |
| 3 | `be-p3-mutation-proof-gaps` | P3 | BE-03 | 29 | Парно с P1/P2 фиксами: denial + redact + OR JQL + restore DeletedAt |
| 4 | `be-p3-utf8-rune-truncate` | P3 | BE-04 | 5 | Единый helper truncate-by-runes (search/messenger/imex/marketplace) |
| 5 | `be-p3-repo-lists-defense` | P3 | BE-05, BE-06 | 11 | ProjectIDEQ/DeletedAtIsNil + Limit на residual lists |
| 6 | `be-p3-dead-code-layering` | P3 | BE-07 | 14 | JQL service API; OpenRawDB/stopCh/dead reader; CreateTable tx |
| 7 | `be-p3-errors-input-hygiene` | P3 | BE-08, BE-09 | 13 | Safe apperr + clamp/allowlist/nil UUID/invite base URL |
| 8 | `be-p3-ws-calls-lifecycle` | P3 | BE-10 | 6 | Rooms cleanup / bridge Stop / nil-RBAC fail-closed; JWT ticket — после product |
| 9 | `be-p3-platform-http-obs` | P3 | BE-11, BE-12 | 9 | OG/Validate/CORS + Drain/baggage/Timed/metrics/probes |
| 10 | `be-p3-ai-auth-rbac-audit` | P3 | BE-13, BE-14 | 10 | AI scope/routing + onboarding tx / UnassignRole / CTC / Upsert |
| 11 | `be-p3-v4-dev-storage` | P3 | BE-15, BE-16 | 10 | Coffee PII / news audit / orphan / layout ACL + dev login/seed |
| 12 | `be-p3-ent-hygiene-mojibake` | P3 | BE-17, BE-18 | 11 | MaxLen/mentions/dup index + «Рёбра»/UTF-8 comments; **U05** после VersionMixin ADR |

**Не в задачах до уточнения:** U01, U02, U03, U04, U05 (5 классов / 9 сырых ID).

---

## Handoff

- Следующий шаг W12 (другие агенты): frontend P3 (W9–W11) — по назначениям оркестратора.
- Team-lead: завести TASK **только** по строкам таблицы группировки с мутацией «да»; U\* — в backlog README как «Blocked: needs ADR/scope» (пересечения с P1-U06 / P2-U01 / P2-BE-17).
- Исполнение: после green P0/P1/P2; порядок — silent-errors + mutation-proof раньше schema cosmetics; BE-03 парно с соответствующими P1/P2 фичами.

---

*Артефакт W12 Stage B · read-only по product code · 2026-07-28*
