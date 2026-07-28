# W12 — реестр уникальных P2-классов (backend, волны 1–8)

> TASK-008 · Стадия B · W12 · 2026-07-28  
> Источники: `A{1-8}-*.md`, сводки `W1-SUMMARY`…`W8-SUMMARY`, [`00-INDEX.md`](00-INDEX.md)  
> Scope: **только backend** (W1–W8). Frontend W9–W11 — вне артефакта.  
> **TASK-\*.md не создаются** — только реестр + предложение группировки.  
> Формат: зеркало [`W12-REGISTRY-P1-BACKEND.md`](W12-REGISTRY-P1-BACKEND.md).

## Сводка

| Метрика | Значение |
|---------|---------:|
| Сырых P2 (W1–W8) | **257** |
| Уникальных классов | **27** |
| Из них с критерием мутации «да» | **22** |
| В секции «требует уточнения» | **5** |
| Предлагаемых задач бэклога | **11** |

Сырые по волнам: W1=28 · W2=36 · W3=32 · W4=24 · W5=31 · W6=28 · W7=38 · W8=40 → **257**  
(полный Stage A P2=422 = 257 backend + 165 frontend W9–W11).

P0/P1 W1–W8 **не** в таблице (отдельные реестры): [`W12-REGISTRY-P0.md`](W12-REGISTRY-P0.md), [`W12-REGISTRY-P1-BACKEND.md`](W12-REGISTRY-P1-BACKEND.md).

---

## Уникальные классы дефектов

| ClassID | Severity | Title | Source IDs | Primary file:line | Mutation criterion | Suggested TASK slug |
|---------|----------|-------|------------|-------------------|--------------------|---------------------|
| P2-BE-01 | P2 | DTO `validate` tags не исполняются (нет StructValidator) | W1-A1-007, W1-A1-012, W2-A1-005, W2-A1-009, W3-A1-006, W6-A1-004, W6-A3-007, W7-A3-005 | `backend/internal/tracker/dto.go:16` | **да** — после wiring `StructValidator` пустой `title`/`email`/`permissions` → 400; снять validator → снова 201 | `be-p2-dto-struct-validator` |
| P2-BE-02 | P2 | Тесты на mock зелёные при живом IDOR/authz (нет Ent-negative) | W1-A1-009, W1-A1-010, W1-A1-011, W1-A2-011, W1-A2-012, W1-A5-012, W2-A1-007, W2-A1-008, W2-A1-010, W2-A2-016, W2-A2-017, W2-A4-008, W2-A4-009, W3-A1-008, W3-A1-009, W3-A1-010, W3-A2-016, W3-A4-005, W4-A1-006, W4-A2-005, W5-A1-006, W5-A1-007, W5-A5-001, W5-A5-002, W5-A5-003, W5-A5-004, W6-A1-006, W6-A3-010, W6-A3-011, W6-A4-008, W6-A5-001, W6-A5-002, W6-A5-003, W6-A5-004, W6-A5-005, W7-A1-006, W7-A1-007, W7-A1-008, W7-A4-012 | `backend/testframework/tracker/handler_flow_test.go:20` | **да** — снять project-scope в Ent-repo → Ent-negative stranger→404 краснеет; mock-only suite остаётся зелёным | `be-p2-authz-negative-ent-tests` |
| P2-BE-03 | P2 | Optimistic lock: нет `VersionEQ` / TOCTOU check-then-act | W1-A2-007, W1-A3-004, W1-A4-008, W2-A5-004, W7-A2-003 | `backend/internal/tracker/repo_issues.go:416` | **да** — убрать `VersionEQ` → два concurrent PUT с одним version оба 200; с guard → второй 409 | `be-p2-optimistic-lock-version-eq` |
| P2-BE-04 | P2 | Update/Find без `DeletedAtIsNil` (мутация soft-deleted) | W1-A3-005, W2-A3-004, W2-A3-005, W3-A5-011, W7-A3-008 | `backend/internal/tracker/repo_issues.go:316` | **да** — Update soft-deleted issue/TC/scheduled / ZIP trash → 404/исключены; без фильтра → 200 / trash в ZIP | `be-p2-block-mutate-softdeleted` |
| P2-BE-05 | P2 | Multi-step запись без одной tx (+ bases cells replace) | W1-A3-006, W1-A5-008, W2-A2-009, W2-A2-011, W2-A4-006, W2-A5-001, W2-A5-002, W2-A5-006, W3-A2-009, W3-A2-011, W3-A3-007, W3-A5-009, W3-A5-012, W6-A3-008, W7-A2-004, W7-A5-004 | `backend/internal/tms/service_test_cases.go:127` | **да** — fail на 2-м шаге → orphan отсутствует (rollback); partial PUT cells → merge, не wipe | `be-p2-atomic-multi-step-tx` |
| P2-BE-06 | P2 | Unbounded List / N+1 / silent trunc без `has_more` | W2-A2-006, W2-A2-007, W2-A2-008, W2-A3-006, W2-A3-007, W2-A4-007, W2-A5-005, W3-A3-006, W7-A2-010, W7-A4-008, W7-A5-003 | `backend/internal/tms/service_reports.go:256` | **да** — List без Limit при N>cap → 400/`has_more`; снять Limit → silent truncate или timeout | `be-p2-list-caps-nplus1` |
| P2-BE-07 | P2 | Сырой `err.Error()` / неверный status + JSON ≠ TZ | W1-A2-010, W2-A2-010, W2-A2-013, W3-A1-005, W4-A1-005, W4-A4-005, W5-A1-001, W5-A1-004, W5-A1-005, W5-A3-004, W5-A4-006, W6-A3-006, W7-A1-002, W7-A3-006, W7-A4-007, W7-A4-011 | `backend/internal/platform/middleware.go:24` | **да** — Forbidden→403 без internals; ErrorHandler только `{"error","request_id"}`; вернуть `code`/`message` → контракт краснеет | `be-p2-safe-error-mapping` |
| P2-BE-08 | P2 | Hard delete вместо soft-delete / корзины 30д | W2-A3-008, W2-A5-003, W4-A5-007, W7-A2-006, W7-A3-010 | `backend/internal/tms/ent_repository.go:444` | **да** — Delete suite/webhook/integration/CRM → `deleted_at` + trash API; hard `Delete()` → тест корзины краснеет | `be-p2-soft-delete-consistency` |
| P2-BE-09 | P2 | Секреты plaintext / без `Sensitive()` / argv с паролем | W4-A5-008, W5-A3-006, W5-A3-007, W7-A3-004, W7-A3-007, W7-A4-013, W8-A2-012, W8-A2-013, W8-A4-006, W8-A4-007 | `backend/internal/ai/repository_ent.go:32` | **да** — `cipher==nil` → Create provider fail; `Sensitive()` скрывает String(); dump URI без password в argv | `be-p2-secrets-hardening` |
| P2-BE-10 | P2 | Fail-open Redis/DB/settings/authz=nil + rate-limit ZAdd | W1-A3-007, W1-A4-004, W3-A3-004, W3-A4-003, W4-A1-003, W4-A1-004, W4-A3-004, W4-A4-007, W5-A2-006, W5-A2-007, W6-A1-003, W6-A4-004, W6-A4-005, W7-A4-009 | `backend/internal/ws/handler.go:148` | **да** — Redis/settings fail → deny/invite_only; `authz==nil` → refuse start; 429 не раздувает ZSET | `be-p2-fail-closed-defaults` |
| P2-BE-11 | P2 | Sprint poker: CAS/TTL/RBAC + StartSprint race | W1-A4-002, W1-A4-003, W1-A4-005, W1-A4-006, W1-A4-007, W1-A4-009 | `backend/internal/sprint/poker.go:108` | **да** — vote после reveal → reject; stranger session → 403; два Start → один active; end&lt;start → 400 | `be-p2-sprint-poker-and-start` |
| P2-BE-12 | P2 | Слабый perm / Get без membership / avatar URL / cross-ref | W1-A1-006, W1-A2-009, W1-A5-006, W1-A5-007, W2-A1-004, W2-A2-015, W3-A1-004, W3-A1-007, W6-A1-005, W7-A2-005, W7-A2-009 | `backend/internal/tracker/handler.go:73` | **да** — CreateBoard без Edit → 403; Get deleted project → 404; `javascript:` avatar → 400; stranger contact_id → 403 | `be-p2-weak-perm-get-avatar` |
| P2-BE-13 | P2 | Messenger: invite token в NATS, leave owner, forward unread, mentions | W3-A2-010, W3-A2-012, W3-A2-013, W3-A2-014, W3-A2-015, W3-A3-003, W3-A3-005 | `backend/internal/messenger/service_channel_invites.go:191` | **да** — `member.joined` без token; last owner Leave → 400; Forward → unread++; mention non-member → drop | `be-p2-messenger-invite-leave-forward` |
| P2-BE-14 | P2 | WS: zombie subscribe + session revoke не рвёт комнаты | W3-A4-002, W3-A4-004 | `backend/internal/ws/hub.go:204` | **да** — Unregister до subscribe → нет утечки room; revoke refresh → клиент Unregister | `be-p2-ws-session-lifecycle` |
| P2-BE-15 | P2 | Calls: guest entropy, conflicts IDOR, ephemeral limit, DTO drop | W3-A5-005, W3-A5-006, W3-A5-007, W3-A5-008, W3-A5-010 | `backend/internal/calls/service.go:86` | **да** — guest code/RL; stranger CheckConflicts → 403; ephemeral &gt;2000 → 400; DurationMinutes персистится | `be-p2-calls-guest-conflicts-dto` |
| P2-BE-16 | P2 | Platform boot/config + outbox failed/claim + OTEL/Prom PII | W4-A1-002, W4-A2-002, W4-A2-003, W4-A2-004, W4-A3-001, W4-A3-002, W4-A3-003, W4-A4-003, W4-A4-004, W4-A4-006, W4-A4-008, W4-A4-009, W4-A5-005, W4-A5-006 | `backend/internal/platform/config.go:260` | **да** — prod+DeriveDevKey → validate fail; attempts≥10 → `failed`; SKIP LOCKED; route template без UUID | `be-p2-platform-boot-outbox-obs` |
| P2-BE-17 | P2 | AI: stream cancel/ctx, privacy default, DM→LLM, Assist mapping | W5-A1-002, W5-A1-003, W5-A2-004, W5-A2-005, W5-A2-008, W5-A3-005, W5-A3-008, W5-A3-009, W5-A3-010, W5-A3-011, W5-A3-012 | `backend/internal/ai/service.go:467` | **да** — cancel parent → stream stops; `AllowExternalAI` default false; DM без `/ai` → no LLM; quota→429 | `be-p2-ai-stream-privacy` |
| P2-BE-18 | P2 | Search: JQL OR игнор, shard swallow, Meili SSRF, soft-deleted channel | W5-A4-004, W5-A4-005, W5-A4-007, W5-A4-008, W5-A4-009 | `backend/internal/search/jql.go:47` | **да** — OR → OR hits; shard err → partial flag; `MEILI_HOST` SSRF blocked; soft-deleted channel → 0 hits | `be-p2-search-jql-shard-ssrf` |
| P2-BE-19 | P2 | RBAC Has(0)/names + GDPR actor/founder + layering/panic | W6-A3-004, W6-A3-005, W6-A3-009, W6-A4-003, W6-A4-006, W6-A4-007, W6-A4-009 | `backend/internal/rbac/permissions.go:126` | **да** — Bit=0 → Has false; ClearRole → cache miss; gdpr details с `actor_id`; founder self-erase → 403 | `be-p2-rbac-gdpr-hardening` |
| P2-BE-20 | P2 | Integrations/imex: CTC length, xlsx mislabel, Allure/JUnit edge | W2-A1-006, W2-A2-012, W2-A2-014, W2-A4-004, W2-A4-005, W7-A3-003, W7-A3-009 | `backend/internal/gitlab/webhook_verify.go:41` | **да** — unequal-length token → reject без length-oracle; excel → xlsx magic; Allure zip/swagger согласованы | `be-p2-integrations-imex-contract` |
| P2-BE-21 | P2 | Storage confirm/contentType + mail/backup/dashboard residuals | W1-A2-008, W7-A1-003, W7-A1-004, W7-A1-005, W7-A2-007, W7-A4-005, W7-A4-006, W7-A4-010, W7-A5-005, W7-A5-006 | `backend/internal/storage/service.go:104` | **да** — Confirm без object → 404; `text/html` → 400; DailyLimit enforced; invoice number unique; today UTC; HasMore корректный | `be-p2-storage-mail-dashboard` |
| P2-BE-22 | P2 | Ent: missing unique indexes / FK edges / CHECK invariants | W8-A1-005, W8-A1-006, W8-A1-007, W8-A2-006, W8-A2-007, W8-A2-008, W8-A2-009, W8-A2-010, W8-A2-014, W8-A2-015, W8-A2-016, W8-A2-017, W8-A2-018, W8-A3-006, W8-A3-008, W8-A3-013, W8-A4-005, W8-A5-004 | `backend/ent/schema/component.go:46` | **да** — duplicate `(project,name)` → unique violation; XOR debit/credit CHECK; TestCaseVersion unique; bookmark→channel FK | `be-p2-ent-unique-fk-check` |

**Классы с мутацией «да»: 22** (P2-BE-01…22).  
**Всего уникальных классов: 27** (= 22 + 5 в секции уточнения).

---

## Требует уточнения

Находки/классы, для которых **нельзя** завести TASK по правилу 57 (нет однозначного критерия мутации / нужен ADR или product-scope). После уточнения — вернуть в очередь задач.

| ClassID | Source IDs | Суть | Чего не хватает для мутации / задачи |
|---------|------------|------|--------------------------------------|
| P2-BE-U01 | W6-A1-001, W6-A1-002, W6-A2-002, W6-A2-003 | Legacy Telegram login API + suite (ADR-AUTH-01); рядом с P1-U02 | **FIXED TASK-102:** `telegram.go`/`telegram_bot.go`+suite удалены; bind в notification |
| P2-BE-U02 | W8-A1-008, W8-A2-005, W8-A2-011, W8-A3-004, W8-A3-005, W8-A3-007, W8-A3-009, W8-A3-010, W8-A3-011, W8-A3-012, W8-A3-014, W8-A3-015, W8-A4-004, W8-A4-008, W8-A4-009, W8-A5-001 | Системный gap `version`/`deleted_at` на mutable Ent (остаток P1-U06) | ADR VersionMixin + exclusion list. Unique/FK/CHECK — в BE-22; без списка scope задача бесконечна |
| P2-BE-U03 | W8-A1-004, W8-A4-010 | `OnDelete` policy на `edge.To` (остаток P1-U05) | ADR Cascade vs SetNull vs app purge. Без политики нет канона «правильного» cascade-теста |
| P2-BE-U04 | W7-A5-001, W7-A5-002 | Анонимный `/api/docs` + `/api/status` recon | Product: закрыть за auth/`!release` vs оставить V4-06. Без решения mutation неоднозначна |
| P2-BE-U05 | W7-A2-008 | FR-INTR-11 coffee auto-DM не реализован | Scope: greenfield DM vs только notification. Без scope mutation на отсутствующем API | 

Покрытие сырых ID: все **257** входят ровно в один класс из P2-BE-01…22 или P2-BE-U01…U05.

Проверка:

| Группа | Σ сырых |
|--------|--------:|
| BE-01…02 (DTO + tests) | 47 |
| BE-03…06 (OL / soft-mutate / tx / lists) | 37 |
| BE-07…10 (errors / soft-delete / secrets / fail-closed) | 45 |
| BE-11…15 (sprint / perm / messenger / ws / calls) | 31 |
| BE-16…18 (platform / AI / search) | 30 |
| BE-19…22 (rbac-gdpr / imex / storage / ent) | 42 |
| **Σ с мутацией** | **232** |
| U01…U05 | 25 |
| **Итого** | **257** |

---

## Предлагаемая группировка в задачи бэклога (~11)

Не создавать файлы сейчас — только карта ClassID → будущий TASK slug. Нумерация TASK после закрытия P1-волны (свободный слот — у team-lead).

| # | Suggested slug | Priority | ClassIDs | Сырых ID (оценка) | Заметка |
|---|---------------|----------|----------|------------------:|---------|
| 1 | `be-p2-dto-error-contract` | P2 | BE-01, BE-07 | 24 | StructValidator wiring + TZ `{"error","request_id"}` + safe mapping |
| 2 | `be-p2-authz-ent-negative-tests` | P2 | BE-02 | 39 | Не чинит prod IDOR (это P1), но закрывает «зелёный mock» — парно с P1-фиксами |
| 3 | `be-p2-ol-softdelete-tx-lists` | P2 | BE-03, BE-04, BE-05, BE-06 | 37 | VersionEQ + DeletedAtIsNil + WithTx + Limit/has_more |
| 4 | `be-p2-soft-delete-secrets` | P2 | BE-08, BE-09 | 15 | Trash path + Sensitive/encrypt/argv |
| 5 | `be-p2-fail-closed-ratelimit` | P2 | BE-10 | 14 | Fail-closed defaults + ZSET order |
| 6 | `be-p2-sprint-poker-start` | P2 | BE-11 | 6 | Poker CAS/RBAC + StartSprint atomic |
| 7 | `be-p2-perm-messenger-ws-calls` | P2 | BE-12, BE-13, BE-14, BE-15 | 25 | Weak perm/avatar + messenger lifecycle + WS revoke + calls guest |
| 8 | `be-p2-platform-outbox-obs` | P2 | BE-16 | 14 | Boot/config + outbox failed/claim + OTEL/Prom |
| 9 | `be-p2-ai-search` | P2 | BE-17, BE-18 | 16 | Stream/privacy + JQL OR / Meili SSRF |
| 10 | `be-p2-rbac-gdpr-integrations-storage` | P2 | BE-19, BE-20, BE-21 | 24 | RBAC/GDPR + CTC/imex + storage/mail/dashboard |
| 11 | `be-p2-ent-unique-fk-check` | P2 | BE-22 | 18 | Unique/FK/CHECK сейчас; **U02+U03** — после ADR VersionMixin/OnDelete |

**Не в задачах до уточнения:** U01, U02, U03, U04, U05 (5 классов / 25 сырых ID).

---

## Handoff

- Следующий шаг W12 (другие агенты): frontend P2 (W9–W11), P3-сводка — по назначениям оркестратора.
- Team-lead: завести TASK **только** по строкам таблицы группировки с мутацией «да»; U\* — в backlog README как «Blocked: needs ADR/scope» (часть уже пересекается с P1-U05/U06/U02).
- Исполнение: после green P0/P1; порядок — fail-closed/secrets/error-contract раньше schema mixin; Ent-тесты (BE-02) — парно с соответствующими P1 IDOR-фиксами.

---

*Артефакт W12 Stage B · read-only по product code · 2026-07-28*
