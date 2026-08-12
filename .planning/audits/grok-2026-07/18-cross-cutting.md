# Cross-cutting Combinations

**Аудит:** A18 · READ-ONLY (кроме этого отчёта) · 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Источники:** `.planning/audits/grok-2026-07/01-auth-rbac.md` … `16-hygiene.md`  
**Метод:** стык находок X+Y → составной баг хуже суммы частей; сверка `file:LINE` в коде.

## Что проверено

| Стык | Зоны | Фокус |
|------|------|--------|
| auth × любой домен | A1 × A2/A3/A5/A6/A7/A9 | pending→JWT усиливает все IDOR/authz-дыры |
| RBAC × tracker/TMS | A1 × A3 × A9 | role IDOR + resource-by-key без project scope |
| Ent schema × handlers | A8 × A3/A5/A6 | `version` в схеме vs отсутствие `VersionEQ` / gap ~80% схем |
| CRM × accounting | A6 × A9 | write-IDOR → won→invoice → race pay / double journal |
| API-контракт × FE state | A11 × A10 × A15 | missing routes, multipart/JSON, Telegram UI, pagination |
| JetStream Ack pattern | A1 × A2 × A3 | Ack-on-error в audit / notification / automation |
| Founder admin-bypass | A1 × A7 × A16 | RBAC/dashboard cut-off + duplicated `IsSystemAdmin` |
| Tests broken × claimed coverage | A12 × AGENTS/COVERAGE | ent.init panic vs «83% / gate PASS» |
| Perf × indexes × IDOR | A8 × A13 × A7 | unbounded board/dashboard + missing FK index + dashboard IDOR |
| Storage pipeline | A4 × A9 | confirm/download/presign NF-09 compound |
| Soft-delete policy | A8 × A4 × A2 | hard-delete Channel/Comment × search deleted_at × msg search |
| Ops shutdown / health | A8 × A14 | liveness=deps × consumer fail-open × `os.Exit` |
| Auth delivery × mail stub | A1 × A7 × A14 | Resend fail→200 × mail stub × OTP in logs |
| Integrations surface | A5 × A9 | imex auth-only × PDF unscoping × webhook any-secret |
| Realtime FE×BE | A10 × A2 | no reconnect × instant-call WS deny × channel store/query drift |

Повторы уже зафиксированные как «= A1-00x» внутри A9 **не** считаются новыми составными находками сами по себе — в таблицу попали только комбинации, где **два+ независимых дефекта** усиливают impact.

## Находки

| ID | Severity | Связанные ID | Файл:строка | Комбинированная проблема | Почему хуже вместе | Как чинить |
|----|----------|--------------|-------------|--------------------------|--------------------|------------|
| A18-001 | P0 | A1-001, A9-001 × A2-001/A9-006 × A6-001/A9-004 × A5-002/A9-007 | `auth/service_login.go:64`; `auth/middleware.go:128`; `messenger/service_search.go:11-29`; `crm/service.go:80-123,172-214`; `imex/handler.go:26-35` | Pending-user после Magic Link получает JWT (`CreateSession` без status gate; `userIsDeactivated` игнорирует pending) и сразу эксплуатирует org-wide дыры: поиск всех сообщений, CRM write-IDOR, imex export без RBAC | По отдельности IDOR требуют «уже authenticated member». С pending→JWT любой, кто прошёл verify до approve (ORG-06 bypass), читает весь мессенджер / правит CRM / выгружает чужие проекты **до** админского approve | Сначала закрыть A1-001 (нет сессии при `status≠active`); затем scoped search / `authorizeDeal` / RBAC на imex. Regression: pending verify → 403 + denial на search/CRM/export |
| A18-002 | P0 | A1-002, A1-003 × A3-002, A3-003 | `rbac/custom_roles.go:73-84,126-132`; `tracker/handler_issues.go:100-105`; `tms/handler_test_cases.go:94-128` | Cross-project role Update/Assign + issue/TC lookup по key без `:projectId` | Украденная/назначенная чужая роль даёт bitmask в проекте A; параллельно key-IDOR позволяет читать/править `B-42` из membership A → **escalation + lateral movement** по org | Scope role по `projectId`; `FindIssueByKey(ctx, projectID, key)` / то же для TC; deny system roles |
| A18-003 | P0 | A3-001 × A8-004 × A1-001 | `sprint/service.go:107-149`; `ent/schema/sprint.go:17-43` (нет `version`); `auth/service_login.go:64` | `CompleteSprint` переносит issues чужого спринта в sprint вызывающего `projectID`; у Sprint нет optimistic lock; pending JWT расширяет круг атакующих | Один вызов = **theft задач между проектами**. Без `version` гонка Complete/Start не детектится. Pending user с membership в любом проекте = готовый вектор data-exfil | Scoped `GetSprint`+project; VersionMixin на Sprint; закрыть pending→JWT |
| A18-004 | P0 | A6-001 × A6-002 × A6-004 | `crm/service.go:200-201` (`tryCreateInvoice` на won); `crm/service.go:172-214` (нет owner check); `accounting/service_invoice.go:116-168`; schema `invoice.go` FK deal не unique | CRM write-IDOR → `stage=won` на чужой сделке создаёт invoice; concurrent `MarkInvoicePaid` вне tx → orphan posted journals; TOCTOU duplicate invoices per deal | Authz-дыра превращается в **финансовую порчу книг** (двойной revenue/банк), не «просто PII edit» | `authorizeDeal` до Update; atomic pay tx + unique `crm_deal_id`; deny non-owner won |
| A18-005 | P0 | A12-001 × COVERAGE/AGENTS «83% / COV-01 PASS» × A12-007 | `ent/runtime.go:1735-1737` (+ `schema/outboxevent.go:28-34`); `.planning/COVERAGE.md:41,119,158`; `testframework/tracker/service_test.go:51-77` | `ent.init` panic ломает почти все `go test -tags hubtest` пакеты; COVERAGE/AGENTS заявляют 83% и gate PASS; tracker критичный слой на mock без Ent | **Claimed coverage ≠ runnable fact.** Регрессии A18-001…004 не ловятся suite’ом на этой ветке; mock-tracker ещё и не ловил бы VersionEQ/FK даже после codegen | `go generate` Ent; перегнать coverage; критичные CRUD на enttest; пометить COVERAGE stale до зелёного gate |
| A18-006 | P1 | A8-003, A8-004 × A3-004 × A6-003 × A5-011 | `ent/schema/project.go:17-52` (нет version); `tms/ent_repository.go:247-303` (`AddVersion` без `VersionEQ`); `crm/repository.go:81-84` (то же); `marketplace/service.go:121` (`version>0` optional) | TZ «version везде» нарушен на ~80% схем; где поле есть (CRM/TMS/marketplace) — handlers не делают WHERE VersionEQ / позволяют `version:0` bypass | Схема и runtime **врут в разные стороны**: поле есть → клиент шлёт version → lost update всё равно; поля нет → гонки необнаружимы. Системный отказ optimistic lock, не один баг модуля | VersionMixin rollout; всегда `Where(VersionEQ)` + require `version≥1`; 409 на conflict |
| A18-007 | P1 | A1-005, A1-006 × A7-002 × A16-005 | `rbac/service.go:47-48` (только `admin`); `rbac/middleware.go:113-114`; `dashboard/handler.go:111-112`; `auth/service_organization.go:29` (+ копии accounting/audit/hr/…) | Founder отрезан от RBAC admin-bypass и `/dashboard/admin`; критерий `IsSystemAdmin` скопирован ≥6 раз и расходится с RBAC/dashboard | ORG-04 (единственный system owner) ломается **фрагментарно**: audit/auth admin OK, project RBAC и dashboard — нет. Дубликаты гарантируют повтор drift | Единый `auth.IsSystemAdmin` везде; founder\|\|admin в RBAC + dashboard + RequireAdmin |
| A18-008 | P1 | A1-008 × A2-010 × A3-006 | `audit/consumer.go:50-54,97-100`; `notification/consumer.go:132-170` (комментарий врёт); `tracker/automation_consumer.go:84-109` | Три JetStream consumer’а ACK’ают после/независимо от ошибки обработки | Одновременная потеря **audit trail + notifications + automation** при DB blip. MaxDeliver бесполезен. Event mesh деградирует молча | Ack только на success; иначе Nak/NakWithDelay; единый consumer helper + тест |
| A18-009 | P1 | A1-007 × A7-005 × A14-006 | `auth/magiclink.go:91-96`; `mail/service.go:32-39`; `auth/magiclink.go:92-96` (+ Loki path A14) | Magic Link: Resend fail → HTTP 200 «sent»; mail SendAlert — stub при живом `RESEND_API_KEY`; OTP/email в логах | Пользователь не входит; ops думает, что mail/Resend работают (`ResendReady`/`Sent++`); OTP может уехать в Loki. **Auth delivery + ops alert + secrecy** ломаются вместе | Fail closed на send error; реальный Resend client или `ResendReady=false`; никогда не логировать OTP; маскировать email |
| A18-010 | P1 | A11-003 × A9-019 × A15-004 × A11-004 | FE `telegram-button.tsx:25-27`; BE `auth/handler.go:31-55` (нет route); `auth/service_login.go:94-115`; `api/auth.ts:94-96`; `REQUIREMENTS.md` LOGN-05 vs TZ | UI зовёт `/auth/telegram-code`; API-клиент — `/auth/telegram`; service login Telegram жив; TZ/ADR запрещают Telegram login; REQ обещает | Пользователь видит forbidden login UX (вечный 404); dead attack surface ждёт accidental re-wire; planning↔TZ↔UI — тройной drift | Удалить UI+DTO+LoginWithTelegram*; REQ/wiki → Magic Link only; тест «no telegram auth routes» |
| A18-011 | P1 | A11-002 × A11-007 × A5-001 × A12-004 | FE `api/tms.ts:681-691` + `client.ts:44-67` (JSON.stringify FormData); `api/tms.ts:695-707` (`JSON.parse` blob); `imex/handler.go:80` (mappings `nil`); settings modules без Vitest | TMS import через JSON client мёртв; export blob ломается parse; CSV imex HTTP mappings всегда nil; FE settings/calls без тестов | Три независимых «import/export работает» пути **все мертвы**, а suite их не ловит → ложная уверенность в imex/TMS | Multipart/binary helpers (как tracker import-export); auto-map CSV; Vitest на import-dialog |
| A18-012 | P1 | A10-003 × A2-007 × A10-001 | FE `stores/ws-store.ts:63-65`; BE `wiring_ws.go:49-52`; FE `hooks/use-channels.ts:51-74` | Нет WS reconnect; instant-call room всегда deny; WS channel events пишут в zustand, UI читает Query | После сетевого обрыва real-time мёртв до remount; instant-звонок **никогда** не получает chat events; список каналов stale даже при живом WS | Backoff reconnect; `IsCallParticipant` для instant; invalidate/setQueryData на channel_* |
| A18-013 | P1 | A10-002 × A3-018 | FE `routes/.../issues/index.tsx:74-81`; BE `tracker/service_issues.go:169-182` | FE «Загрузить ещё» не мержит новую страницу в display; BE `HasMore: len==limit` ложно на последней полной странице | Двойная ложь пагинации: пользователь не видит уже загруженную page2 **и** видит «есть ещё» когда нет (или наоборот после лишних кликов) | `useInfiniteQuery` / append effect; BE fetch `limit+1` |
| A18-014 | P1 | A5-002 × A5-010 × A5-008/A5-009 | `imex/handler.go:26-35`; `imex/pdf.go:22-29,74-81`; `github/handler.go:119-126` + `service.go:152`; `gitlab/handler.go:99-102` | Imex только `authMw`; PDF игнорирует path `projectID`; webhook принимается любым secret инстанса, repo/path из payload | Authenticated user экспортирует PDF чужого run/sprint; параллельно утечка webhook secret A → подпись событий в интеграцию B. **Data exfil + supply-chain status spoof** на одной интеграции | RBAC imex; scoped PDF query; verify secret **найденной** интеграции |
| A18-015 | P1 | A8-001 × A13-001 × A7-001 | `ent/schema/issue.go:117-127`; `tracker/repo_boards.go:56-62`; `dashboard/handler.go:71-81` | Нет индекса `(project, deleted_at)` на issues; Kanban `WithIssues` без Limit; dashboard project stats без membership | IDOR на dashboard/stats + unbounded board load **без** FK-индекса = дешёвый CPU/IO DoS любым JWT (усиливается A18-001) | Index + per-column limit; membership на Project dashboard |
| A18-016 | P1 | A4-006 × A4-007 × A4-008/A9-017 | `storage/repository.go:62-67,77-80`; `storage/service.go:84-88` | Confirm без `status=pending`; download для pending; presigned PUT без max size (NF-09) | Залить объект ≫25MB → confirm перепривязать entity → скачать/раздать незавершённый или огромный объект. Полный обход BodyLimit + entity hijack | Confirm WHERE pending; download только confirmed; presign size constraint / lifecycle |
| A18-017 | P2 | A8-005, A8-006 × A4-002 × A2-001 | `ent/schema/channel.go` / `messenger/ent_repository.go:86-88`; `ent/schema/comment.go` + `tracker/repo_comments.go:91`; `search/service.go:227-234`; `messenger/service_search.go:11-29` | Channel/Comment hard-delete vs TZ soft-delete; Ctrl+K без `deleted_at`; messenger search без membership | Hard-delete ломает FK/members; soft-deleted issues всё ещё в search; msg search читает всё. Политика «корзина 30д» + NF-03 **системно** нарушены | Soft-delete Channel/Comment; `deleted_at IS NULL` в search; membership filter в msg search |
| A18-018 | P2 | A14-001 × A8-016 × A8-008 | `wiring_messenger.go:33-74`; `wiring_notification.go:21-23,53-55`; `wiring_run.go:16-22` | `/api/health` = liveness с 503 на deps; consumer Start fail → process жив; `Listen` err → `os.Exit` минуя defer cleanup | Restart storm при кратком Redis/PG; при этом NATS consumers могут быть мертвы (silent lag); аварийный Listen exit не закрывает PG/NATS/Mongo | Split `/live`/`/ready`; fatal если consumers не стартовали; return+cleanup вместо `os.Exit` |
| A18-019 | P2 | A11-001 × A1-001 × A10-006 | FE `api/projects.ts:133-135` + `onboarding.lazy.tsx:312`; BE `project/handler.go:49` (`/invites/join`); `components/admin/user-management.tsx:186-202` | Onboarding join бьёт в `/projects/join` (404); pending всё равно может получить JWT иным путём; admin UI глотает ошибки как «пользователей нет» | Регистрация/онбординг UX сломан; security hole pending открыт; админ не видит pending queue при сетевой ошибке → **approve не случится, JWT уже есть** | FE path `/invites/join`; закрыть pending session; QueryStateView на user-management |
| A18-020 | P2 | A9-012 × A9-021 × A12-005 | `apigateway/apikey.go:189-203` vs `auth/middleware.go:59-98`; `wiring_fiber.go:63-68`; FE calls без Vitest | X-API-Key выставляет `userID`, но RequireAuth требует JWT; `/api/docs` публичен; calls UI без functional tests | CI/operators вынуждены класть JWT в скрипты (обход «API keys»); swagger отдаёт карту атаки; calls authz-дыры (A2) без FE regression | APIKey path в RequireAuth; docs auth/off в prod; Vitest/e2e calls smoke |

## Чисто / не подтвердилось

- **A9 «= A1/A2/…» как отдельные cross-findings:** не дублировались — A9 уже агрегатор волны 1; в A18 только усиливающие пары.
- **SSRF baseline × SafeHTTP:** AI adapters / WebPush / GitHub client / apigateway outbound согласованы; A4-005 (Ollama vs SafeDial) — одиночный конфликт зоны A4, не новый стык с A9 (тот же SSRF guard).
- **Fiber v3 RBAC order:** tracker/tms/github порядок middleware после handler — в A9 чисто; не комбинируется с IDOR (IDOR = resource scope, не order).
- **Magic Link constant-time / refresh rotation / CORS fail-start:** зрелые контроли; не усиливают IDOR-класс сами по себе.
- **Bases owner-scope (после фикса) × storage CR-04:** оба scoped; стык не даёт нового IDOR (в отличие от CRM).
- **A15 planning drift (ROADMAP 65–125) × product bugs:** документальный drift; не повышает severity runtime IDOR → оставлено в A15, не A18.
- **A16 unused VelocityChart / jsdom:** hygiene-only; нет security amplify.

## Оценка: 3.5/10

Системный класс «authenticated ⇒ почти org-admin на чужих ресурсах» усиливается **pending→JWT** и непроверяемым test/coverage слоем. Отдельные зоны ~4.5–6.5/10; на стыках появляются P0 data-theft / financial corruption / untestable gate. После закрытия A18-001…005 и Ack/VersionEQ/Founder унификации оценка cross-cutting могла бы вырасти к ~7/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 5 |
| P1 | 11 |
| P2 | 4 |
| P3 | 0 |
| **Total** | **20** |

**Score: 3.5/10**
