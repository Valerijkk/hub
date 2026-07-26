# Verification P0/P1

**Дата:** 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Зона:** A17 — верификация всех P0/P1 из `01-auth-rbac.md` … `16-hygiene.md`  
**Режим:** READ-ONLY (кроме этого отчёта)

## Метод

1. Извлечены все строки findings с severity **P0** или **P1** из отчётов A01–A16 (включая A15 planning-таблицу; A09 частично дублирует A01–A06).
2. Для каждой находки открыт указанный `файл:строка` (и смежные handler/repo при необходимости).
3. Вердикт: **CONFIRMED** | **FALSE_POSITIVE** | **NEEDS_INFO** | **DOWNGRADED_TO_P2** | **UPGRADED_TO_P0**.
4. Дубликаты A9 ↔ A1/A2/A3/A5/A6 помечены `= Ax-yyy`; код проверен один раз, вердикт совпадает.

**Всего заявлено P0/P1:** 110 строк (106 табличных в A01–A14/A16 + 4 в A15).

## Результаты

| Source ID | Claimed | Verdict | Файл:строка | Пояснение |
|-----------|---------|---------|-------------|-----------|
| A1-001 | P0 | CONFIRMED | `auth/service_login.go:64` + `middleware.go:128` | `LoginWithMagicLink` всегда `CreateSession`; `userIsDeactivated` только `StatusDeactivated`. Pending (invite_only → `CreatePendingMember`) получает JWT. |
| A1-002 | P0 | CONFIRMED | `rbac/custom_roles.go:73-122` | `UpdateCustomRole`/`DeleteCustomRole` по `roleID` без `role.project_id == :projectId`. |
| A1-003 | P0 | CONFIRMED | `rbac/custom_roles.go:126-132` | `AssignCustomRole` → `SetMemberCustomRole` без сверки роли с проектом. |
| A1-004 | P0 | CONFIRMED | `audit/service.go:379-441` | `DeleteUserGDPR` без проверки `SystemRoleFounder`. |
| A1-005 | P1 | CONFIRMED | `rbac/service.go:47` | Bypass только `SystemRoleAdmin`; `IsSystemAdmin` = founder\|\|admin — founder без membership режется. |
| A1-006 | P1 | CONFIRMED | `rbac/middleware.go:113-114` | `RequireAdmin` → только `admin`, founder 403. |
| A1-007 | P1 | CONFIRMED | `auth/magiclink.go:91-93` + `handler.go:90-102` | Сбой Resend → `return code, nil` → HTTP 200; код в prod не в ответе и не в логе (комментарий врёт). |
| A1-008 | P1 | CONFIRMED | `audit/consumer.go:50-54,97-100` | `Ack` всегда после `processEvent`, даже при err `CreateAuditLog`. |
| A1-009 | P1 | CONFIRMED | `audit/service.go:283-309` | Ban/Deactivate без защиты founder (только self-ban). |
| A2-001 | P0 | CONFIRMED | `messenger/service_search.go:11-29` + `mongo_repository.go:282-288` | Поиск без `userID`/membership; без `channel_id` — все каналы. |
| A2-002 | P1 | CONFIRMED | `calls/service.go:49-65` | `CreateCall` без `IsChannelMember`. |
| A2-003 | P1 | CONFIRMED | `calls/service_chat.go:33-79` | Send/List chat без participant/channel authz. |
| A2-004 | P1 | CONFIRMED | `calls/recording.go:89-97,173-177` | Start без membership; Stop по `recordingID` без user/RBAC. |
| A2-005 | P1 | CONFIRMED | `notification/push.go:230-261` | Rebind endpoint → caller; unsubscribe по endpoint без `userID`. |
| A2-006 | P1 | CONFIRMED | `calls/handler_calls.go:268-294` | `RaiseHand` без participant check. |
| A2-007 | P1 | CONFIRMED | `wiring_ws.go:49-52` | Instant (`ChannelID==nil`) → всегда `false` → deny `call:{id}` WS. |
| A3-001 | P0 | CONFIRMED | `sprint/service.go:107-149` | `CompleteSprint` грузит sprint по ID; перенос issues в `projectID` из URL без ownership sprint. |
| A3-002 | P1 | CONFIRMED | `tracker` FindIssueByKey | Lookup по `issueKey` без `:projectId` → cross-project IDOR. |
| A3-003 | P1 | CONFIRMED | `tms/ent_repository.go:149-158` | `FindTestCaseByKey` без project scope. |
| A3-004 | P1 | CONFIRMED | `tms/ent_repository.go:247-305` | `UpdateOneID`+`AddVersion(1)` без `VersionEQ`. |
| A3-005 | P1 | CONFIRMED | `tracker/repo_time.go:22-45,66-83` | LogTime без started/ended; StopTimer `*StartedAt` → nil panic на manual log. |
| A3-006 | P1 | CONFIRMED | `tracker/automation_consumer.go:84-109` | Ошибки → всё равно `_ = msg.Ack()`. |
| A3-007 | P1 | CONFIRMED | `tms/service_test_cases.go:207-224` | Rollback version без принадлежности к test case. |
| A3-008 | P1 | CONFIRMED | `tms/service_runs.go:214-241` + bug-from-result | `resultID` без привязки к run/project. |
| A3-009 | P1 | CONFIRMED | `tms/service_test_cases.go:324-336` | `testCaseID=uuid.Nil` → `UpdateOneID(Nil)`. |
| A3-010 | P1 | CONFIRMED | `tms/allure_matcher.go:139` | `FindTestCaseByKey` без `projectID`. |
| A3-011 | P1 | CONFIRMED | `tracker/service_comments.go:74-97` | Update/Delete comment только по UUID, без author/issue. |
| A3-012 | P1 | CONFIRMED | `tms/service_import.go:140-148` | Run по `runID` без сверки с `:projectId`. |
| A4-001 | P1 | CONFIRMED | `search/jql.go:104-108` | Обе ветки `assignee` → `HasAssignee()`, UUID игнорируется. |
| A4-002 | P1 | CONFIRMED | `search/service.go:227+` | tsvector SQL без `deleted_at IS NULL`. |
| A4-003 | P1 | CONFIRMED | `search/service.go:628-645` | CreateSavedFilter без membership. |
| A4-004 | P1 | CONFIRMED | `search/service.go:648-659` | ListSavedFilters без membership. |
| A4-005 | P1 | CONFIRMED | `ai/fallback.go:51-61` + `adapter_openai.go:32` + `safedialer.go:36` | Offline/Ollama выбирает localhost; SafeDial блокирует `127.0.0.0/8`. |
| A4-006 | P1 | CONFIRMED | `storage/repository.go:62-67` | `ConfirmUpload` bare `UpdateOneID` без status/soft-delete guard. |
| A5-001 | P0 | CONFIRMED | `imex/handler.go:80` + `service.go:78-98` | `mappings` всегда `nil` → пустой fieldMap → все строки skip («нет title»); import сломан. |
| A5-002 | P1 | CONFIRMED | `imex/handler.go:26-35` | Только `authMw`, нет `RequirePermission`. |
| A5-003 | P1 | CONFIRMED | `github` DeleteIntegration | Delete по ID без привязки к `:projectId`. |
| A5-004 | P1 | CONFIRMED | `gitlab` DeleteIntegration | То же. |
| A5-005 | P1 | CONFIRMED | `gitlab/service.go:347-356` | `TriggerPipeline` не сверяет integration.project с URL. |
| A5-006 | P1 | CONFIRMED | `github/handler.go:51` | CreateBranch — только authMw. |
| A5-007 | P1 | CONFIRMED | `gitlab/handler.go:43` | То же. |
| A5-008 | P1 | CONFIRMED | `github/handler.go:119-126` | HMAC: любой секрет инстанса. |
| A5-009 | P1 | CONFIRMED | `gitlab/handler.go:99-102` | Token: любой секрет инстанса. |
| A5-010 | P1 | CONFIRMED | `imex/pdf.go:22-29,74-81` | `projectID` unused; load by run/sprint ID only. |
| A6-001 | P0 | CONFIRMED | `crm/service.go:80-123,172-214` | Update/Delete contact/deal без owner/admin. |
| A6-002 | P0 | CONFIRMED | `accounting/service_invoice.go:116-168` | MarkInvoicePaid вне одной tx; TOCTOU status. |
| A6-003 | P1 | CONFIRMED | `crm/service.go:88-90,180-181` + repo | App-level version check без `VersionEQ` в UPDATE. |
| A6-004 | P1 | CONFIRMED | `accounting` InvoiceExists+Create + schema | TOCTOU; Deal→invoices не Unique. |
| A6-005 | P1 | CONFIRMED | `calendar/service.go:77-83` | Все approved leave org любому JWT. |
| A7-001 | P1 | CONFIRMED | `dashboard/handler.go:71-81` | `userID` в `_`; stats без membership. |
| A7-002 | P1 | CONFIRMED | `dashboard/handler.go:111` | Только `"admin"`; founder 403. |
| A7-003 | P1 | CONFIRMED | `backup/service.go:206-214` | `setMaintenanceMode` только лог. |
| A7-004 | P1 | CONFIRMED | `backup/service.go:75-104` | Ошибка dump/zip/S3 → `return …, nil` (HTTP success). |
| A7-005 | P1 | CONFIRMED | `mail/service.go:32-39` | При ключе Resend — log `"queued"` + nil, без HTTP. |
| A7-006 | P1 | CONFIRMED | `intranet/coffee_repository.go:77` | `Offset(1)` **до** CreateRound → пропускает реальный previous. |
| A8-001 | P1 | DOWNGRADED_TO_P2 | `ent/schema/issue.go:117-127` | Нет индекса на project FK — perf/hygiene, не correctness bug. |
| A8-002 | P1 | DOWNGRADED_TO_P2 | `ent/schema/testcase.go:116-127` | То же для test cases. |
| A8-003 | P1 | CONFIRMED | `ent/schema/project.go:17-52` | Mutable Project без `version` (TZ optimistic lock). |
| A8-004 | P1 | CONFIRMED | `ent/schema/*` | **14**/~82 схем с `field.Int("version")` — системный gap vs TZ. |
| A8-005 | P1 | CONFIRMED | `channel.go` + `messenger/ent_repository.go:86-88` | Нет `deleted_at`; hard `DeleteOneID`. |
| A8-006 | P1 | CONFIRMED | `comment.go` + `repo_comments.go:91` | Hard delete comments. |
| A8-007 | P1 | CONFIRMED | `wiring_messenger.go:26-31` | Mongo connect без `Disconnect` cleanup. |
| A8-008 | P1 | CONFIRMED | `wiring_run.go:16-22` | `os.Exit(1)` на Listen err обходит deferred cleanups. |
| A9-001 | P0 | CONFIRMED | = A1-001 | Тот же pending→JWT. |
| A9-002 | P0 | CONFIRMED | = A1-002 | Role IDOR. |
| A9-003 | P0 | CONFIRMED | = A1-003 | Assign escalation. |
| A9-004 | P0 | CONFIRMED | = A6-001 | CRM IDOR. |
| A9-005 | P0 | CONFIRMED | = A3-001 | Sprint IDOR. |
| A9-006 | P0 | CONFIRMED | = A2-001 | Messenger search. |
| A9-007 | P1 | CONFIRMED | = A5-002 | imex без RBAC. |
| A9-008 | P1 | CONFIRMED | = A5-008/009 | Webhook any-secret. |
| A9-009 | P1 | CONFIRMED | = A3-002/003 + A5 | Key/UUID без project scope (агрегат). |
| A9-010 | P1 | CONFIRMED | = A1-005/006 | Founder bypass gap. |
| A9-011 | P1 | CONFIRMED | = A1-004/009 | GDPR/ban founder. |
| A9-012 | P1 | CONFIRMED | `apigateway/apikey.go:189-203` vs `auth/middleware.go:59-64` | APIKey пишет `userID`, RequireAuth всё равно требует Bearer. |
| A9-013 | P1 | CONFIRMED | = A2-002…006 | Calls authz gap (агрегат). |
| A10-001 | P1 | CONFIRMED | `hooks/use-channels.ts:51-74` | WS → zustand; UI читает React Query. |
| A10-002 | P1 | CONFIRMED | `issues/index.tsx:74-81` | Cursor loadMore кладёт текущую страницу, следующая «теряется» на один клик. |
| A10-003 | P1 | CONFIRMED | `stores/ws-store.ts:63-65` | `onclose` очищает socket; reconnect только при смене auth deps. |
| A10-004 | P1 | CONFIRMED | `hooks/use-presence.ts:24` | `userIds.sort()` мутирует входной массив. |
| A11-001 | P0 | CONFIRMED | FE `api/projects.ts:133-135` vs BE `project/handler.go:49` | FE `/projects/join`, BE `/invites/join`; onboarding вызывает FE. |
| A11-002 | P0 | CONFIRMED | FE `api/tms.ts:681-691` + `client.ts:44-67` | FormData → `JSON.stringify` + `application/json`; BE ждёт multipart. |
| A11-003 | P0 | DOWNGRADED_TO_P2 | `telegram-button.tsx:25-27` | Роута нет, но **TelegramButton нигде не импортирован** — мёртвый код, не live UX. |
| A11-004 | P1 | CONFIRMED | `api/auth.ts:94-96` | `loginTelegram` → `/auth/telegram` отсутствует; dead API helper. |
| A11-005 | P1 | CONFIRMED | `api/messenger.ts:44-50` vs handler | `PUT …/notify-mode` не зарегистрирован. |
| A11-006 | P1 | CONFIRMED | FE date `YYYY-MM-DD` vs BE `*time.Time` | RFC3339 mismatch → 400. |
| A11-007 | P1 | CONFIRMED | `tms.ts` export + `client.ts:99-103` | `api.get<Blob>` всегда `JSON.parse`. |
| A12-001 | P0 | CONFIRMED | `ent/runtime.go:1734-1737` + `schema/outboxevent.go:28-34` | После вставки `trace_context` fields[4]=`status` (string); cast to `DefaultAttempts int` → panic на `ent.init`. |
| A12-002 | P1 | CONFIRMED | `notification/coverage_wave79_test.go:45-48` | Имя про Ack, assert нет. |
| A12-003 | P1 | CONFIRMED | `github/service_mock_test.go:310-314` | «PublishesWhen…» без assert. |
| A12-004 | P1 | DOWNGRADED_TO_P2 | `modules/settings/*` | 0 Vitest — gap покрытия, не runtime defect. |
| A12-005 | P1 | DOWNGRADED_TO_P2 | `modules/calls/` | Нет Vitest-домена — test gap (security-critical UI, но не доказанный баг). |
| A12-006 | P1 | DOWNGRADED_TO_P2 | messenger Vitest = icon | Покрытие-гигиена. |
| A12-007 | P1 | DOWNGRADED_TO_P2 | `tracker/service_test.go:51-77` | mockRepo vs Ent — стратегия тестов, не баг продукта. |
| A13-001 | P1 | CONFIRMED | `tracker/repo_boards.go:56-62` | `WithIssues` без Limit на колонку. |
| A13-002 | P1 | CONFIRMED | `bases/repository.go:123-127` | GetTable грузит все records; write path вызывает полный GetTable. |
| A13-003 | P1 | CONFIRMED | `dashboard/repository.go:114-122` | `.All` без limit. |
| A13-004 | P1 | CONFIRMED | `gantt-chart.tsx:113-116` | N+1 `listIssueLinks` per row. |
| A13-005 | P1 | CONFIRMED | `notification/service.go:75-116` | Sync Dispatch × recipients с repeated SQL. |
| A13-006 | P1 | CONFIRMED | `sprint/repository.go:87-93` | `ListSprints` + `WithIssues()` unbounded. |
| A14-001 | P1 | CONFIRMED | `wiring_messenger.go:33-74` + compose healthcheck | `/api/health` = deps readiness, используется как liveness. |
| A14-002 | P1 | CONFIRMED | `seaweedfs-s3.json:7-8` + compose prod | Committed `hub`/`hubsecret`; config.validate запрещает в non-dev. |
| A14-003 | P1 | CONFIRMED | `docker-compose.prod.yml` NATS | JetStream без auth. |
| A14-004 | P1 | CONFIRMED | `docker-compose.monitoring.yml` | Порты 9090/9093/3100/3001 на хост. |
| A14-005 | P1 | CONFIRMED | `config.go:147` + prometheus | Пустой `METRICS_ALLOWED_CIDRS` → open `/metrics`. |
| A15-001 | P0 | CONFIRMED | `STATE.md` vs `ROADMAP.md` | Planning P0: STATE 125/125, корневой ROADMAP без 65–125 / Progress до 17. |
| A15-002 | P1 | CONFIRMED | `ROADMAP.md` 38→40 | Phase 39 пропущена в checklist при наличии dir. |
| A15-003 | P1 | CONFIRMED | `PROJECT.md` | ~103 unchecked при shipped модулях. |
| A15-004 | P1 | CONFIRMED | REQ LOGN-05 vs TZ | Telegram login в REQ/dead code vs Magic-Link-only TZ. |
| A16-004 | P1 | DOWNGRADED_TO_P2 | `backend/go.mod:14` | Fiber `v3.0.0-beta.3` — факт; осознанный стек Hub, не свежий regression. |

## Сводка

### Counts

| Метрика | Значение |
|---------|----------|
| Всего P0/P1 заявлено | **110** |
| CONFIRMED (тот же severity) | **101** |
| DOWNGRADED_TO_P2 | **9** |
| UPGRADED_TO_P0 | **0** |
| FALSE_POSITIVE | **0** |
| NEEDS_INFO | **0** |

### CONFIRMED P0 (уникальные + дубли A9)

**Уникальные product/runtime P0 (12):**  
A1-001, A1-002, A1-003, A1-004, A2-001, A3-001, A5-001, A6-001, A6-002, A11-001, A11-002, A12-001  

**Planning P0 (1):** A15-001  

**A9 дубли тех же P0 (6, тоже CONFIRMED):** A9-001…A9-006  

**P0 → DOWNGRADED:** A11-003 (мёртвый TelegramButton)

### CONFIRMED P1 (после downgrade)

**CONFIRMED P1 rows:** 101 − 19 P0-confirmed rows (13 unique+planning + 6 A9) = **82** P1-confirmed строк в таблице  
(часть — дубли A9; уникальных P1 дефектов ≈ **70+**)

### DOWNGRADED_TO_P2 list (9)

| ID | Было | Почему downgrade |
|----|------|------------------|
| A8-001 | P1 | Missing FK index — perf, не логический баг |
| A8-002 | P1 | То же |
| A11-003 | P0 | Компонент не смонтирован нигде |
| A12-004 | P1 | Test coverage gap |
| A12-005 | P1 | Test coverage gap |
| A12-006 | P1 | Test coverage gap |
| A12-007 | P1 | Test strategy (mocks), не product bug |
| A16-004 | P1 | Известный pin стека (Fiber v3 beta) |

### FALSE_POSITIVE list

*(пусто)*

### Замечания по качеству отчётов

- **Сильные стороны:** точные `file:line`, воспроизводимые security/IDOR claims (auth pending, RBAC roles, messenger search, CRM, sprint, FE/BE path mismatch, ent.runtime panic).
- **Слабые:** A09 почти полный rehash волны 1 без новой глубины; завышение severity на test-gaps (A12-004…007) и индексы (A8-001/002); A11-003 как P0 без проверки mount.
- **Критичный блокер регрессии:** A12-001 (`ent.init` panic) — ломает любой `go test`/сервер, импортирующий `ent`.

## Оценка качества волны 1–2: **8/10**

Высокая точность находок (0 FP при 110), хорошая привязка к коду и TZ; минус за дубли A9, inflated P1 на coverage/hygiene и один ложный P0 (мёртвый Telegram UI).
