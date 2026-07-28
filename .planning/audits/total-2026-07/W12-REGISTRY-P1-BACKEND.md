# W12 — реестр уникальных P1-классов (backend, волны 1–8)

> TASK-008 · Стадия B · W12 · 2026-07-28  
> Источники: `A{1-8}-*.md`, сводки `W1-SUMMARY`…`W8-SUMMARY`, [`00-INDEX.md`](00-INDEX.md)  
> Scope: **только backend** (W1–W8). Frontend W9–W11 — вне артефакта.  
> **TASK-\*.md не создаются** — только реестр + предложение группировки.

## Сводка

| Метрика | Значение |
|---------|---------:|
| Сырых P1 (W1–W8) | **99** |
| Уникальных классов | **49** |
| Из них с критерием мутации «да» | **43** |
| В секции «требует уточнения» | **6** |
| Предлагаемых задач бэклога | **13** |

Сырые по волнам: W1=19 · W2=14 · W3=17 · W4=6 · W5=9 · W6=6 · W7=9 · W8=19 → **99**  
(полный Stage A P1=141 = 99 backend + 42 frontend W9–W11).

P0 W1–W8 **не** в таблице (отдельный трек): W1-A5-001 · messenger P0 в A3-2 · W4-A5 P0×2.

---

## Уникальные классы дефектов

| ClassID | Severity | Title | Source IDs | Primary file:line | Mutation criterion | Suggested TASK slug |
|---------|----------|-------|------------|-------------------|--------------------|---------------------|
| P1-BE-01 | P1 | Tracker workflow CRUD IDOR (status/transition без project scope) | W1-A1-002, W1-A1-003, W1-A1-004, W1-A1-005, W1-A2-002, W1-A2-003, W1-A2-004, W1-A3-002, W1-A3-003 | `backend/internal/tracker/handler_workflow.go:86` | **да** — убрать `HasProjectWith` → member A мутирует status/transition проекта B | `tracker-workflow-project-scope-idor` |
| P1-BE-02 | P1 | Tracker `DeleteCustomField` IDOR | W1-A1-001, W1-A2-001, W1-A3-001 | `backend/internal/tracker/service_custom_fields.go:48` | **да** — delete только по `fieldID` → чужое поле удаляется | `tracker-custom-field-delete-idor` |
| P1-BE-03 | P1 | CreateIssue: StatusID/ParentID/Labels без ∈ project | W1-A2-005 | `backend/internal/tracker/service_issues.go:58` | **да** — `status_id` чужого проекта → 400/404; без guard → 201 | `tracker-create-issue-cross-project-refs` |
| P1-BE-04 | P1 | Issue↔message link: channel без project/membership | W1-A2-006 | `backend/internal/tracker/service_issue_message_link.go:79` | **да** — link на чужой channel → 404 | `tracker-issue-message-link-channel-scope` |
| P1-BE-05 | P1 | `CompleteSprint` не в одной tx | W1-A4-001 | `backend/internal/sprint/service.go:156` | **да** — fail после MoveIssues → issues в новом sprint, status старого не completed | `sprint-complete-atomic-tx` |
| P1-BE-06 | P1 | Project `RevokeInvite` IDOR | W1-A5-002 | `backend/internal/project/handler.go:505` | **да** — revoke invite чужого project по UUID → 404 | `project-invite-revoke-idor` |
| P1-BE-07 | P1 | `UpdateMemberRole` без защиты последнего owner | W1-A5-003 | `backend/internal/project/service.go:364` | **да** — demote единственного owner → 400 | `project-last-owner-demote-guard` |
| P1-BE-08 | P1 | Teams: только auth, нет ownership / project ACL | W1-A5-004 | `backend/internal/project/handler.go:550` | **да** — stranger `AddTeamToProject` → 403 | `project-teams-acl` |
| P1-BE-09 | P1 | `DeleteTemplate` без created_by / admin | W1-A5-005 | `backend/internal/project/templates.go:101` | **да** — B удаляет шаблон A → 403 | `project-template-delete-ownership` |
| P1-BE-10 | P1 | TMS Suite/Component Update/Delete IDOR | W2-A1-001, W2-A1-002, W2-A1-003, W2-A2-001, W2-A2-002, W2-A3-002, W2-A3-003 | `backend/internal/tms/handler_suites.go:72` | **да** — mutate suite/component чужого проекта → 404 | `tms-suite-component-project-scope-idor` |
| P1-BE-11 | P1 | TMS Environment Update/Delete + CompareEnvironments IDOR | W2-A3-001, W2-A2-003 | `backend/internal/tms/environment.go:181` | **да** — Compare/Update чужого env → 404 / без имён | `tms-environment-project-scope-idor` |
| P1-BE-12 | P1 | TMS cross-project FK (suite/label/component/TC→run) | W2-A2-004, W2-A2-005 | `backend/internal/tms/service_test_cases.go:35` | **да** — CreateTC с чужим suite_id / AddCases чужого TC → 404 | `tms-cross-project-fk-refs` |
| P1-BE-13 | P1 | TMS import: нет LimitReader / cap глубины | W2-A4-001, W2-A4-003 | `backend/internal/tms/importer.go:268` | **да** — oversized/nested → 413/400; снять лимит → accept | `tms-import-body-depth-limits` |
| P1-BE-14 | P1 | TMS import: steps молча теряются | W2-A4-002 | `backend/internal/tms/importer.go:108` | **да** — колонка Steps → непустой `CreateTestCaseInput.Steps` | `tms-import-steps-mapping` |
| P1-BE-15 | P1 | Messenger `RemoveBookmark` IDOR (без userID) | W3-A1-001, W3-A2-006, W3-A3-001 | `backend/internal/messenger/service_bookmarks.go:19` | **да** — B удаляет bookmark A → 404 | `messenger-bookmark-owner-scope` |
| P1-BE-16 | P1 | Reaction / Pin / AddBookmark без channel membership | W3-A1-002, W3-A2-003, W3-A2-004, W3-A2-007 | `backend/internal/messenger/service_reactions.go:13` | **да** — non-member mutate → 403 | `messenger-channel-membership-mutates` |
| P1-BE-17 | P1 | DeleteMessage: роль по URL-каналу ≠ msg.ChannelID | W3-A1-003, W3-A2-002 | `backend/internal/messenger/service_messages.go:161` | **да** — owner A + msg B → 404 | `messenger-delete-message-channel-match` |
| P1-BE-18 | P1 | GetPollResults без membership (утечка VoterIDs) | W3-A2-005 | `backend/internal/messenger/service_polls.go:202` | **да** — non-member → 403 | `messenger-poll-results-membership` |
| P1-BE-19 | P1 | Channel invite revoke без ChannelIDEQ | W3-A2-008 | `backend/internal/messenger/service_channel_invites.go:207` | **да** — admin A revoke invite B → 404 | `messenger-channel-invite-revoke-scope` |
| P1-BE-20 | P1 | AddReaction: дубликат emoji при ModifiedCount==0 | W3-A3-002 | `backend/internal/messenger/mongo_repository.go:218` | **да** — double AddReaction → 1 entry (не 2) | `messenger-reaction-emoji-dedup` |
| P1-BE-21 | P1 | WS Register/Unregister reorder → zombie clients | W3-A4-001 | `backend/internal/ws/hub.go:114` | **да** — Unregister до Register → клиент не остаётся в map после done | `ws-register-unregister-race` |
| P1-BE-22 | P1 | Scheduled calls CRUD/list без creator/membership | W3-A5-001, W3-A5-002, W3-A5-003, W3-A5-004 | `backend/internal/calls/scheduled.go:81` | **да** — stranger Update/Delete/List/Create → 403 | `calls-scheduled-authz` |
| P1-BE-23 | P1 | SSRF: isBlockedIP не нормализует IPv4-mapped IPv6 | W4-A1-001 | `backend/internal/platform/http/safedialer.go:65` | **да** — `::ffff:169.254.169.254` → ErrBlockedAddress | `platform-ssrf-ipv4-mapped-block` |
| P1-BE-24 | P1 | Outbox: status=published до Flush/JetStream PubAck | W4-A2-001 | `backend/internal/platform/outbox/worker.go:93` | **да** — drop до Flush → status остаётся pending | `platform-outbox-jetstream-puback` |
| P1-BE-25 | P1 | Postgres: нет SetMaxOpenConns / pool limits | W4-A4-001 | `backend/internal/platform/postgres.go:21` | **да** — OpenConnections ≤ MaxOpen после N запросов | `platform-postgres-conn-pool-limits` |
| P1-BE-26 | P1 | Prod validate пропускает weak DB/Redis/Mongo defaults | W4-A4-002 | `backend/internal/platform/config.go:223` | **да** — DEV_MODE=false + hub:hub@ → validate fail | `platform-prod-reject-default-db-creds` |
| P1-BE-27 | P1 | Webhook Create: auto-secret не в ответе | W4-A5-003 | `backend/internal/apigateway/webhook.go:355` | **да** — Create без secret → one-time secret в JSON | `apigateway-webhook-secret-once` |
| P1-BE-28 | P1 | Webhook Events без allowlist / empty OK | W4-A5-004 | `backend/internal/apigateway/handler.go:213` | **да** — events:[] → 400 | `apigateway-webhook-events-allowlist` |
| P1-BE-29 | P1 | AI content policy только role==user | W5-A2-001 | `backend/internal/ai/content_policy.go:22` | **да** — jailbreak role:system → reject | `ai-content-policy-all-roles` |
| P1-BE-30 | P1 | isOfflineProvider: substring bypass AllowExternalAI | W5-A2-002 | `backend/internal/ai/fallback.go:51` | **да** — host с `/ollama` на публичном URL → filtered out | `ai-offline-provider-host-allowlist` |
| P1-BE-31 | P1 | AI bot: mention / issue lookup / review без authz | W5-A3-001, W5-A3-002, W5-A3-003 | `backend/internal/ai/messenger_bot.go:242` | **да** — mention без bot∈members / foreign key / review вне project → denial | `ai-bot-channel-issue-review-authz` |
| P1-BE-32 | P1 | Search: soft-delete channels/messages + private discovery | W5-A4-001, W5-A4-002, W5-A4-003 | `backend/internal/search/service.go:519` | **да** — soft-deleted → 0 hits; private non-member → не в выдаче | `search-channel-membership-softdelete` |
| P1-BE-33 | P1 | RBAC: swallow invalidate + stale admin mask после demote | W6-A3-001, W6-A3-002, W6-A3-003 | `backend/internal/rbac/custom_roles.go:103` | **да** — сузить perms/demote → RequirePermission сразу 403 | `rbac-cache-invalidate-on-role-change` |
| P1-BE-34 | P1 | GDPR erase неполный + без транзакции | W6-A4-001, W6-A4-002 | `backend/internal/audit/service.go:408` | **да** — после erase нет Mongo/files PII; cascade fail → HardDelete не вызывается | `audit-gdpr-cascade-atomic` |
| P1-BE-35 | P1 | Notification consumer: repo err → nil → Ack | W7-A1-001 | `backend/internal/notification/consumer.go:253` | **да** — transient repo err → Nak; soft-nil → событие потеряно | `notification-consumer-nak-on-repo-err` |
| P1-BE-36 | P1 | GitLab pipeline: plaintext variables в NATS | W7-A3-001 | `backend/internal/gitlab/service.go:424` | **да** — secret var → payload без plaintext values | `gitlab-pipeline-variables-redact` |
| P1-BE-37 | P1 | ListGitHubLinks без project membership | W7-A3-002 | `backend/internal/github/handler.go:47` | **да** — user вне project → 403 | `github-links-project-rbac` |
| P1-BE-38 | P1 | Storage NF-09: presign без size + Confirm entity scope | W7-A4-001, W7-A4-002 | `backend/internal/storage/service.go:84` | **да** — oversized PUT / confirm чужой issue → reject | `storage-presign-limit-entity-scope` |
| P1-BE-39 | P1 | Backup MarkCompleted при s3Client==nil | W7-A4-004 | `backend/internal/backup/service.go:105` | **да** — без S3 → MarkFailed + error | `backup-fail-closed-without-s3` |

*Продолжение — schema / coffee / классы с уточнением мутации:*

| ClassID | Severity | Title | Source IDs | Primary file:line | Mutation criterion | Suggested TASK slug |
|---------|----------|-------|------------|-------------------|--------------------|---------------------|
| P1-BE-40 | P1 | CoffeeEnrollment: нет unique на user | W8-A1-001 | `backend/ent/schema/coffeeenrollment.go:32` | **да** — двойной enroll → unique violation | `ent-coffee-enrollment-unique-user` |
| P1-BE-41 | P1 | NotificationPreference UNIQUE с nullable project_id | W8-A3-001 | `backend/ent/schema/notificationpreference.go:60` | **да** — два Create с NULL project → одна строка (partial unique) | `ent-notif-pref-partial-unique` |
| P1-BE-42 | P1 | JournalEntry→Invoice без .Unique() (ломает 1:1) | W8-A3-003 | `backend/ent/schema/journalentry.go:56` | **да** — вторая Invoice на тот же entry → constraint | `ent-journal-invoice-o2o` |
| P1-BE-43 | P1 | FR-INTR-10: нечётный N → unpaired, не тройка | W7-A2-002 | `backend/internal/intranet/coffee_service.go:201` | **да** — N=3 → 0 unpaired / 1 triple | `intranet-coffee-odd-n-triple` |

**Классы с мутацией «да»: 43** (P1-BE-01…43).  
**Всего уникальных классов: 49** (= 43 + 6 в секции уточнения).

---

## Требует уточнения

Находки/классы, для которых **нельзя** завести TASK по правилу 57 (нет однозначного критерия мутации / нужен ADR или product-scope). После уточнения — вернуть в очередь задач.

| ClassID | Source IDs | Суть | Чего не хватает для мутации / задачи |
|---------|------------|------|--------------------------------------|
| P1-BE-U01 | W5-A2-003 | Offline/Ollama помечен OK, `NewSafeHTTPClient` блокирует loopback — consent и SSRF противоречат | Product/ADR: trusted offline dialer vs sidecar vs запрет loopback BYOK. Пока нет выбранного поведения — нельзя написать «сломать X → красный тест» без ломания SSRF-инварианта |
| P1-BE-U02 | W6-A2-001 | Позитивный тест `HandleLogin_StoresCode` держит мёртвый Telegram-login путь (ADR-AUTH-01) | **FIXED TASK-102:** login API+тесты удалены; denial `TestAuth_NoTelegramLoginSurface_Task102` |
| P1-BE-U03 | W7-A2-001 | FR-INTR-07/08: author-delete + `news_poll*` + restore с опросом — модуль не реализован | Scope: полный greenfield polls vs только author-or-admin delete. Без scope нельзя зафиксировать mutation на несуществующем API |
| P1-BE-U04 | W7-A4-003 | Backup maintenance gate — process-local; multi-replica не блокируется | Нужен тестовый harness multi-instance (или явный single-replica affinity ADR). Unit-мутация process flag не доказывает distributed gate |
| P1-BE-U05 | W8-A4-002 | Project `PermanentDelete` / 25+ edges без OnDelete policy | ADR: staged soft→30d→tx cleanup vs explicit OnDelete per-edge. Без политики тест «правильного» cascade неоднозначен |
| P1-BE-U06 | W8-A1-002, W8-A1-003, W8-A2-001, W8-A2-002, W8-A2-003, W8-A2-004, W8-A3-002, W8-A4-001, W8-A4-003, W8-A5-002, W8-A5-005, W8-A5-006, W8-A5-008, W8-A5-009, W8-A5-010 | Системный TZ-gap: `version` / `deleted_at` на mutable-схемах (Channel, Project, Comment, Team, TestSuite, User, WebhookEndpoint, …) | Нужен ADR VersionMixin + список исключений (junction/append-only) и политика soft-delete per-entity. Мутация на *одной* схеме возможна, но без списка scope задача бесконечна |

Покрытие сырых ID: все **99** входят ровно в один класс из P1-BE-01…43 или P1-BE-U01…U06  
(U06 агрегирует 15 schema-ID; U01–U05 — по одному; 43 класса с мутацией покрывают **79** сырых; U\* — **20**).

Проверка:

| Группа | Σ сырых |
|--------|--------:|
| BE-01…09 (W1) | 19 |
| BE-10…14 (W2) | 14 |
| BE-15…22 (W3) | 17 |
| BE-23…28 (W4) | 6 |
| BE-29…32 (W5 без A2-003) | 8 |
| BE-33…34 (W6 без A2-001) | 5 |
| BE-35…39 + BE-43 (W7 без A2-001,A4-003) | 7 |
| BE-40…42 (W8 unique/edge) | 3 |
| **Σ с мутацией** | **79** |
| U01…U05 | 5 |
| U06 | 15 |
| **Итого** | **99** |

---

## Предлагаемая группировка в задачи бэклога (~13)

Не создавать файлы сейчас — только карта ClassID → будущий TASK slug. Нумерация TASK после TASK-008 (свободный слот — у team-lead).

| # | Suggested slug | Priority | ClassIDs | Сырых ID (оценка) | Заметка |
|---|---------------|----------|----------|------------------:|---------|
| 1 | `tracker-project-scope-idor` | P1 | BE-01, BE-02, BE-03, BE-04 | 14 | Одна цепочка handler→service→repo; тесты на Ent, не mock |
| 2 | `project-acl-sprint-tx` | P1 | BE-05, BE-06, BE-07, BE-08, BE-09 | 5 | Invite/teams/templates/owner + CompleteSprint atomic |
| 3 | `tms-project-scope-idor-refs` | P1 | BE-10, BE-11, BE-12 | 11 | Suite/component/env + cross-FK |
| 4 | `tms-import-limits-steps` | P1 | BE-13, BE-14 | 3 | LimitReader/depth + steps mapping |
| 5 | `messenger-membership-idor` | P1 | BE-15, BE-16, BE-17, BE-18, BE-19, BE-20 | 12 | Bookmarks/reactions/pins/polls/invites/delete + emoji dedup |
| 6 | `ws-calls-authz-race` | P1 | BE-21, BE-22 | 5 | WS zombie + scheduled calls ACL |
| 7 | `platform-ssrf-outbox-pool-config` | P1 | BE-23, BE-24, BE-25, BE-26 | 4 | SafeDialer + outbox PubAck + pool + prod defaults |
| 8 | `apigateway-webhook-create-contract` | P1 | BE-27, BE-28 | 2 | Secret once + events allowlist |
| 9 | `ai-policy-bot-authz` | P1 | BE-29, BE-30, BE-31 | 5 | (+ U01 отдельно после ADR) |
| 10 | `search-nf03-softdelete` | P1 | BE-32 | 3 | Channel membership + deleted_at |
| 11 | `rbac-cache-gdpr` | P1 | BE-33, BE-34 | 5 | (+ U02 telegram-login cleanup после решения) |
| 12 | `notif-integrations-storage-backup` | P1 | BE-35, BE-36, BE-37, BE-38, BE-39, BE-43 | 8 | Consumer Nak + GitHub/GitLab + storage + backup + coffee triple |
| 13 | `ent-schema-constraints-tz` | P1 | BE-40, BE-41, BE-42 | 3 | Unique/O2O сейчас; **U05+U06** — после ADR VersionMixin/soft-delete/cascade |

**Не в задачах до уточнения:** U01, U02, U03, U04, U05, U06 (6 классов / 20 сырых ID).

---

## Handoff

- Следующий шаг W12 (другие агенты): frontend P1 (W9–W11), P0-реестр, P2-сводка — по назначениям оркестратора.
- Team-lead: завести TASK-009+ **только** по строкам таблицы группировки с мутацией «да»; U\* — в backlog README как «Blocked: needs ADR/scope».
- Исполнение (W13+): порядок P0 → эти P1 по числу пользователей (IDOR/authz раньше schema mixin).

---

*Артефакт W12 Stage B · read-only по product code · 2026-07-28*
