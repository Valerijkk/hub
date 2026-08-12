# AUDIT — Ent schemas + apigateway

> TASK-008 · File-auditor · READ-ONLY product · 2026-07-29  
> Ветка: `cursor-grok-4.5` · HEAD: `617af223`  
> Scope: `backend/ent/schema/*.go` (87) · `backend/internal/apigateway/*.go` production (6; `export_hooks.go` — hubtest)  
> Фокус: **OnDelete** · **soft-delete fields** · **webhook SEC-02** · **VersionEQ**  
> Выход: **только этот файл** · коммит **не** делать · **TASK-008 не закрывать**

Перепроверка после ADR/TASK-095/096/097 и remediation A3W4-5 / A3W8-*. Baseline: `A8-*-schemas.md`, `A3W8-*-schemas.md`, `A3W4-5-apigateway.md`.

---

## Охват

| Каталог | Найдено | Прочитано | Пропущено |
|---------|--------:|----------:|----------:|
| `backend/ent/schema/*.go` | 87 | 87 | 0 |
| `backend/internal/apigateway` production | 6 | 6 | 0 |
| `apigateway/export_hooks.go` (hubtest) | 1 | 1 (вне prod-фокуса) | 0 |
| **Σ product in scope** | **93** | **93** | **0** |

### Сводка по схемам (замер HEAD)

| Метрика | Count |
|---------|------:|
| Файлов `.go` в `schema/` | 87 |
| UUID PK (`field.UUID("id")`) на сущностях | 85/85 (excl. helpers/mixin) |
| С `field.Time("deleted_at")` | **35** |
| С `VersionMixin` / inline `version` | **45** (+ mixin-файл) |
| Файлов с ≥1 `entsql.OnDelete` | **26** |
| Без `version` (junction / append / ADR-exclusion) | **35** (ожидаемо) |

---

## Фокус-чеклист

| Тема | Вердикт |
|------|---------|
| **OnDelete** | Крупные parents закрыты ADR-ENT-ONDELETE-CASCADE: `Project`, `Channel`, `Call`, `Issue`, `Poll`, `HubBase`/`HubTable`, `Team`→members, `WebhookEndpoint`→deliveries, `TestRun`→results, `TestCase`→versions, `WorkflowStatus`→from_transitions, `Organization`→NoAction. Остаток: явные аннотации на части intermediate `To` (TestSuite) и user-owned GDPR edges — см. findings. Soft-delete **не** триггерит OnDelete (by design). |
| **soft-delete** | Корзина на ключевых user/project сущностях есть (Issue/Project/Channel/Board/Sprint/Label/…/WebhookEndpoint/APIKey/Recording/Kudo/…). `JournalLine` — join-filter через entry (ADR-ENT-SOFT-DELETE-GAPS §B), колонки нет. **Gap:** `AutomationRule` — mutable + hard `Delete()` без `deleted_at`. |
| **SEC-02 (outbound)** | **OK.** `deliverWithRetry` → всегда `signPayload` → всегда `X-Hub-Signature-256`. Skip/empty-signature path нет. Inbound CTC — github/gitlab, не этот модуль. |
| **VersionEQ** | Webhook **Update:** `VersionEQ` + `AddVersion(1)` + DTO `version` ≥1 → 409 (`A3W4-5-002` **FIXED**). Soft-**Delete** webhook **без** VersionEQ — P3. APIKey без version — ADR-ENT-VERSION-MIXIN unclear exclusion. |

### Закрыто относительно A3 (evidence на HEAD)

| ID / тема | Статус |
|-----------|--------|
| A3W4-5-001 ListActive + `project.DeletedAtIsNil` | FIXED-OK · `repository.go:173-178` |
| A3W4-5-002 Update VersionEQ | FIXED-OK · `repository.go:231-238`, `webhook.go:528-539` |
| A3W4-5-003 `sprint.completed` allowlist | FIXED-OK · убран; комментарий `webhook.go:451-453` |
| A3W4-5-004 API key name MaxLen runes | FIXED-OK · `handler.go:124-127` |
| A3W8-5 Team soft + Cascade members | FIXED-OK · `team.go:27-30`, `:53-54` |
| A3W8-5 WorkflowStatus soft + Cascade transitions | FIXED-OK · `workflowstatus.go` |
| A3W8-5 TestCase versions / TestRun results / WH deliveries Cascade | FIXED-OK |
| SEC-02 outbound HMAC no-skip | FIXED-OK (без регрессии) |

---

## Таблица файлов — `backend/ent/schema`

| file | lines | status | severity | note (фокус) |
|------|------:|--------|----------|--------------|
| `activitylog.go` | 61 | clean | — | append-only; Issue→activities Cascade |
| `aiprovider.go` | 83 | clean* | — | VersionMixin; trash нет — disable via `enabled`; credentials Sensitive |
| `aiusagedaily.go` | 50 | clean | — | counters; unique(user,date) |
| `aiuserpreference.go` | 67 | clean | — | Mixin; provider SetNull OnDelete |
| `apikey.go` | 69 | clean | — | `deleted_at`; version excluded (ADR unclear) |
| `auditlog.go` | 62 | clean | — | append-only |
| `automationrule.go` | 75 | finding | P2 | Mixin; **нет deleted_at**; hard Delete в tracker |
| `backuprecord.go` | 56 | clean | — | ops artifact |
| `board.go` | 71 | clean | — | soft+Mixin+indexes; Project Cascade |
| `bookmark.go` | 58 | clean | — | Channel Cascade→bookmarks |
| `call.go` | 91 | clean | — | Cascade participants/recordings; version excluded |
| `callparticipant.go` | 59 | clean | — | junction |
| `channel.go` | 107 | clean | — | soft+Mixin+Cascade/SetNull owned |
| `channelinvite.go` | 74 | clean | — | junction/invite; Channel Cascade |
| `channelmember.go` | 68 | clean | — | LWW ADR; no VersionMixin |
| `coffeeenrollment.go` | 47 | clean | — | O2O Unique на User |
| `coffeepair.go` | 48 | clean | — | junction |
| `coffeeround.go` | 48 | clean | — | Cascade pairs |
| `comment.go` | 77 | clean | — | soft+Mixin; Issue Cascade |
| `component.go` | 64 | clean | — | soft+Mixin; Project Cascade |
| `crmcontact.go` | 74 | clean | — | soft+Mixin |
| `crmdeal.go` | 81 | clean | — | soft+Mixin; Project Cascade |
| `customfielddefinition.go` | 73 | clean | — | soft+Mixin (TASK-097) |
| `dashboardlayout.go` | 66 | clean | — | Mixin; per-user layout |
| `environment.go` | 70 | clean | — | soft+Mixin |
| `filemetadata.go` | 93 | clean | — | soft; version excluded; Project Cascade |
| `githubintegration.go` | 83 | clean | — | soft+Mixin+Cascade links |
| `githublink.go` | 76 | clean | — | inline version; Issue Cascade |
| `gitlabintegration.go` | 80 | clean | — | soft+Mixin+Cascade links |
| `gitlablink.go` | 79 | clean | — | inline version |
| `helpers.go` | 14 | clean | — | `nowUTC` |
| `hubbase.go` | 64 | clean | — | soft+Mixin; Cascade tables |
| `hubcolumn.go` | 68 | clean | — | soft+Mixin; HubTable Cascade |
| `hubrecord.go` | 60 | clean | — | soft+Mixin; HubTable Cascade |
| `hubtable.go` | 65 | clean | — | soft+Mixin; Cascade columns/records |
| `invite.go` | 74 | clean | — | `revoked_at` + version (не deleted_at) |
| `invoice.go` | 89 | clean | — | soft+Mixin |
| `issue.go` | 140 | clean | — | soft+Mixin; Cascade/SetNull owned |
| `issuelink.go` | 66 | clean | — | Cascade source/target |
| `issuemessagelink.go` | 66 | clean | — | Channel/Issue Cascade peers |
| `journalentry.go` | 76 | clean | — | soft+Mixin; Cascade lines |
| `journalline.go` | 67 | clean | — | нет колонки deleted_at (ADR §B); join-filter |
| `kudo.go` | 55 | clean | — | soft; version N/A (append-ish) |
| `label.go` | 72 | clean | — | soft+Mixin |
| `labelgroup.go` | 65 | clean | — | soft+Mixin |
| `leaverequest.go` | 76 | clean | — | soft+Mixin |
| `ledgeraccount.go` | 61 | clean | — | soft+Mixin |
| `marketplacelisting.go` | 81 | clean | — | soft+Mixin |
| `mixin_version.go` | 22 | clean | — | VersionMixin source |
| `newspoll.go` | 65 | clean | — | inline version; From post Cascade; options/votes via child From Cascade |
| `newspolloption.go` | 58 | clean | — | From poll Cascade; votes via child From |
| `newspollvote.go` | 66 | clean | — | Cascade poll+option |
| `newspost.go` | 68 | clean | — | soft+Mixin; poll O2O через child Cascade |
| `notification.go` | 82 | clean | — | inline version |
| `notificationpreference.go` | 77 | clean | — | Mixin |
| `organization.go` | 69 | clean | — | Mixin; NoAction users/projects |
| `outboxevent.go` | 51 | clean | — | append/outbox |
| `pinnedmessage.go` | 51 | clean | — | Channel Cascade |
| `poll.go` | 83 | clean | — | Mixin; Cascade options/votes |
| `polloption.go` | 54 | clean | — | Cascade votes |
| `pollvote.go` | 62 | clean | — | junction; parent Cascade |
| `project.go` | 107 | clean | — | soft+Mixin; Cascade owned_child set |
| `projectmember.go` | 54 | clean | — | junction; Project Cascade |
| `projecttemplate.go` | 66 | clean | — | Mixin |
| `pushsubscription.go` | 64 | clean | — | version excluded |
| `reaction.go` | 52 | clean | — | junction |
| `recording.go` | 93 | clean | — | soft; version excluded; Call/Project Cascade |
| `role.go` | 63 | clean | — | Mixin; Project Cascade |
| `savedfilter.go` | 74 | clean | — | Mixin; Project Cascade |
| `scheduledcall.go` | 103 | clean | — | soft+Mixin |
| `session.go` | 52 | clean | — | auth sessions |
| `sprint.go` | 85 | clean | — | soft+Mixin; issues SetNull |
| `systemsetting.go` | 37 | clean | — | Mixin KV |
| `team.go` | 63 | clean | — | soft+Mixin; Cascade members |
| `teammember.go` | 48 | clean | — | junction |
| `testcase.go` | 136 | clean | — | soft+Mixin; Cascade versions |
| `testcaseversion.go` | 55 | clean | — | snapshot; unique(case,version_number) |
| `testresult.go` | 92 | clean | — | append; TestRun Cascade |
| `testrun.go` | 99 | clean | — | Mixin; Cascade results |
| `testsuite.go` | 70 | finding | P3 | soft+Mixin; **To test_cases/children без явного OnDelete** |
| `timelog.go` | 59 | clean | — | append; Issue Cascade |
| `user.go` | 161 | clean | — | Mixin; lifecycle `status`; GDPR OnDelete — app playbook |
| `userblock.go` | 56 | clean | — | Unique(blocker,blocked) |
| `webhookdelivery.go` | 58 | clean | — | append log; Endpoint Cascade |
| `webhookendpoint.go` | 70 | clean | — | soft+Mixin; Cascade deliveries |
| `workflowstatus.go` | 85 | clean | — | soft+Mixin; Cascade from_transitions |
| `workflowtransition.go` | 57 | clean | — | Mixin; Cascade from parent |

\* `aiprovider`: без trash-колонки, но delete-API не найден; disable через `enabled` — не эскалирую.

---

## Таблица файлов — `backend/internal/apigateway`

| file | lines | status | severity | note |
|------|------:|--------|----------|------|
| `webhook.go` | 556 | clean | — | SEC-02 sign always; SSRF; Version≥1; project scope; allowlist |
| `repository.go` | 283 | finding | P3 | Update VersionEQ OK; **Delete soft без VersionEQ** |
| `handler.go` | 340 | clean | — | RBAC IDOR; MaxLen runes; authorizeEndpoint |
| `apikey.go` | 181 | clean | — | SHA-256; soft revoke; deactivated user; touch pool |
| `dto.go` | 63 | clean | — | Update.Version обязателен в контракте |
| `errors.go` | 60 | clean | — | MapToFiber 409 VersionConflict |
| `export_hooks.go` | 120 | n/a | — | `//go:build hubtest` — не production |

---

## Findings (активные на HEAD)

| ID | Sev | file:line | Категория | Проблема |
|----|-----|-----------|-----------|----------|
| EA-001 | P2 | `automationrule.go:17-43` (+ runtime `tracker/repo_automation.go:85-91`) | soft-delete | Mutable project rule с `VersionMixin`, **без `deleted_at`**. `DeleteAutomationRule` = hard `AutomationRule.Delete()`. Нарушает TZ «корзина 30д / hard только Admin-GDPR» для user-managed сущности (класс ADR-ENT-SOFT-DELETE-GAPS). |
| EA-002 | P3 | `testsuite.go:56-59` | OnDelete | `To("test_cases")` / self `children` без явного `entsql.OnDelete`. Suite→case optional → Ent default SET NULL; children self-ref — то же. Работает, но **не соответствует** политике ADR «явный OnDelete per-edge» (асимметрия с Channel/Poll/Project). |
| EA-003 | P3 | `repository.go:276-282` | VersionEQ | Soft-delete webhook: `Where(ID, DeletedAtIsNil)` **без** `VersionEQ`, хотя Update требует version (NFR-24). Сценарий: клиент со stale `version` всё ещё может DELETE (wipe чужих незакоммиченных правок URL/events). Низкий риск; симметрия с Update желательна. |

### Не findings (явно проверено)

- **SEC-02:** нет ветки «отправить без подписи» / «пропустить если secret пуст» в send-path; secret `NotEmpty` + encrypt на create.
- **ListActive** фильтрует soft-deleted project (`HasProjectWith(DeletedAtIsNil)`).
- **JournalLine** без своей `deleted_at` — ADR §B, не дефект.
- **APIKey/Recording/Call/… без version** — ADR-ENT-VERSION-MIXIN unclear exclusion.
- **NewsPoll** Cascade на options/votes через `From` children — эквивалент Poll.To Cascade.
- **Organization NoAction** — ADR org_anchor.

---

## Сводка

| Severity | Count |
|----------|------:|
| P0 / P1 | 0 |
| P2 | 1 (EA-001) |
| P3 | 2 (EA-002, EA-003) |
| Schema finding files | 2 |
| Apigateway finding files | 1 |
| Clean (product) | 90 |

**Рекомендация remediation (не в scope этого слота):** soft-delete AutomationRule (+ list filter); явный OnDelete на TestSuite edges; optional `VersionEQ` на soft-delete webhook.

**Не коммитить. TASK-008 не закрывать.**
