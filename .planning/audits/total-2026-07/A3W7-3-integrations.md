# A3 W7 агент A3 — github+gitlab+imex+onec (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W7-3-integrations.md](A2W7-3-integrations.md) · A1: [A7-3-integrations.md](A7-3-integrations.md)  
> Assignment: [W7-ASSIGNMENTS.md](W7-ASSIGNMENTS.md) § A3 · Периметр: **71** · Прочитано: **71/71** · Пропущено: **0**  
> Фокус: webhook HMAC / ConstantTimeCompare · soft-delete · SSRF · secrets  
> Remediation: TASK-024/045/057/060/067/097/120 · P2-BE-01 StructValidator

## Сверка с A2 (W7-A3-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| W7-A3-001 | FIXED-OK | закрыт | `redactPipelineVariables` → NATS `variables` = keys+`***` (`gitlab/service.go:431-458`) |
| W7-A3-002 | FIXED-OK | закрыт | `ListGitHubLinks` → `IsProjectMember` (`github/service.go:472-489`; handler передаёт `userID`) |
| W7-A3-003 | FIXED-OK | закрыт | `constantTimeEqualDigest` sha256+CTC (`gitlab/webhook_verify.go:42-55`) |
| W7-A3-004 | FIXED-OK | закрыт | `ListAllWebhookSecrets` отсутствует; verify = FindByRepo / FindByProjectPath |
| W7-A3-005 | FIXED-OK | закрыт | `validate:"required"` + `StructValidator` (`wiring_fiber.go:32-33`; `github/dto.go:11-14`) |
| W7-A3-006 | FIXED-OK | закрыт | upstream body → slog; клиенту `HTTP %d` (`gitlab/service.go:410-418`) |
| W7-A3-007 | FIXED-OK | закрыт | BearerToken AES-GCM via `integrationCipher` (`onec/service.go:34-86`; wiring `wiring_modules.go:48`) |
| W7-A3-008 | FIXED-OK | закрыт | `ListProjectIssuesPlain` / `ForExport` + `DeletedAtIsNil` (`imex/repository.go:142-158`) |
| W7-A3-009 | FIXED-OK | закрыт | excelize XLSX (`imex/service.go:247-295`) |
| W7-A3-010 | FIXED-OK | закрыт (был A2 residual) | soft-delete integrations: `SetDeletedAt` + `DeletedAtIsNil` на list/get/find/delete (`github/repository.go:168-251`; `gitlab/repository.go:120-191`); IT `soft_delete_task097_test.go` (вне слота) |
| W7-A3-011 | FIXED-OK | закрыт | `LogActivity` err → `slog.WarnContext` (`github/service.go:282-287`) |
| W7-A3-012 | FIXED-OK | закрыт | `truncateCell` на рунах (`imex/truncate.go`; вызов `pdf.go:205`) |
| W7-A3-013 | FIXED-OK | закрыт | default status: `IsDefault` → fallback `Order SortOrder` (`imex/repository.go:69-84`) |

**REGRESS:** 0 · **A2 open residual:** 1 (`W7-A3-010`) → **FIXED-OK** · **NEW:** 3

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/github/converters.go` | 33 | 0 | чисто · DTO без token/secret |
| `backend/internal/github/dto.go` | 105 | 0 | FIXED-OK validate required |
| `backend/internal/github/export_hooks.go` | 95 | 0 | hubtest encrypt helpers · чисто |
| `backend/internal/github/handler.go` | 352 | 0 | FIXED-OK webhook bind + ListGitHubLinks userID |
| `backend/internal/github/pr_review.go` | 105 | 0 | чисто · SafeHTTP client |
| `backend/internal/github/preview.go` | 151 | 1 | SafeHTTP OG OK; NEW issue preview authz |
| `backend/internal/github/repository.go` | 428 | 1 | FIXED-OK soft-delete/encrypt; NEW FindIssue trash |
| `backend/internal/github/service.go` | 576 | 0 | FIXED-OK membership + SafeHTTP + slog activity |
| `backend/internal/github/webhook_verify.go` | 74 | 0 | FIXED-OK HMAC CTC + repo bind · no-skip |
| `backend/testframework/github/converters_test.go` | 66 | 0 | чисто |
| `backend/testframework/github/ent_repository_test.go` | 286 | 0 | чисто |
| `backend/testframework/github/handler_extended_test.go` | 141 | 0 | чисто |
| `backend/testframework/github/handler_validation_test.go` | 158 | 0 | чисто |
| `backend/testframework/github/handler_wave101_test.go` | 58 | 0 | чисто |
| `backend/testframework/github/handler_wave103_test.go` | 126 | 0 | чисто |
| `backend/testframework/github/pr_review_test.go` | 107 | 0 | чисто |
| `backend/testframework/github/preview_test.go` | 128 | 0 | чисто (authz gap в prod) |
| `backend/testframework/github/preview_wave103_test.go` | 57 | 0 | чисто |
| `backend/testframework/github/repository_helpers_test.go` | 54 | 0 | чисто |
| `backend/testframework/github/repository_wave101_test.go` | 59 | 0 | чисто |
| `backend/testframework/github/repository_wave103_test.go` | 137 | 0 | FIXED-OK ListAllWebhookSecrets removed note |
| `backend/testframework/github/service_http_test.go` | 19 | 0 | чисто |
| `backend/testframework/github/service_mock_test.go` | 583 | 0 | чисто |
| `backend/testframework/github/service_test.go` | 154 | 0 | HMAC helper · чисто |
| `backend/testframework/github/service_wave104_test.go` | 69 | 0 | чисто |
| `backend/testframework/github/service_wave115_test.go` | 38 | 0 | чисто |
| `backend/testframework/github/webhook_secret_binding_a5008_test.go` | 89 | 0 | FIXED-OK binding · mutation any-of |
| `backend/internal/gitlab/dto.go` | 133 | 0 | чисто · validate tags |
| `backend/internal/gitlab/export_hooks.go` | 49 | 0 | redact export · чисто |
| `backend/internal/gitlab/handler.go` | 298 | 0 | FIXED-OK LimitReader + token verify |
| `backend/internal/gitlab/mr_review.go` | 177 | 0 | LimitReader · чисто |
| `backend/internal/gitlab/repository.go` | 224 | 1 | FIXED-OK soft-delete; NEW FindIssue trash |
| `backend/internal/gitlab/service.go` | 628 | 0 | FIXED-OK redact/SSRF/body/slog |
| `backend/internal/gitlab/webhook_verify.go` | 67 | 0 | FIXED-OK digest CTC · path bind |
| `backend/testframework/gitlab/converters_test.go` | 28 | 0 | чисто |
| `backend/testframework/gitlab/ent_repository_test.go` | 220 | 0 | чисто |
| `backend/testframework/gitlab/handler_extended_test.go` | 151 | 0 | чисто |
| `backend/testframework/gitlab/handler_test.go` | 158 | 0 | чисто |
| `backend/testframework/gitlab/handler_wave115_test.go` | 37 | 0 | чисто |
| `backend/testframework/gitlab/mr_review_test.go` | 78 | 0 | чисто |
| `backend/testframework/gitlab/service_http_test.go` | 301 | 0 | SSRF smoke · чисто |
| `backend/testframework/gitlab/service_mock_test.go` | 276 | 0 | чисто |
| `backend/testframework/gitlab/service_test.go` | 74 | 0 | чисто |
| `backend/testframework/gitlab/webhook_token_binding_a5009_test.go` | 84 | 0 | FIXED-OK binding · mutation any-of |
| `backend/internal/imex/dto.go` | 47 | 0 | чисто |
| `backend/internal/imex/export_hooks.go` | 52 | 0 | hubtest · чисто |
| `backend/internal/imex/handler.go` | 247 | 0 | FIXED-OK RBAC + LimitReader + xlsx MIME |
| `backend/internal/imex/jira_import.go` | 357 | 0 | LimitReader · чисто |
| `backend/internal/imex/pdf.go` | 213 | 1 | FIXED-OK truncate; NEW sprint trash via ListSprintIssues |
| `backend/internal/imex/repository.go` | 199 | 1 | FIXED-OK ZIP/CSV trash; NEW ListSprintIssues |
| `backend/internal/imex/service.go` | 368 | 0 | FIXED-OK excelize |
| `backend/testframework/imex/handler_flow_test.go` | 225 | 0 | чисто |
| `backend/testframework/imex/jira_import_flow_test.go` | 97 | 0 | чисто |
| `backend/testframework/imex/jira_import_test.go` | 102 | 0 | чисто |
| `backend/testframework/imex/pdf_export_test.go` | 182 | 0 | чисто |
| `backend/testframework/imex/pdf_scope_a5010_ent_test.go` | 121 | 0 | FIXED-OK project scope |
| `backend/testframework/imex/rbac_a5002_test.go` | 139 | 0 | FIXED-OK RBAC |
| `backend/testframework/imex/repository_pdf_test.go` | 78 | 0 | чисто (нет trash-кейса) |
| `backend/testframework/imex/repository_test.go` | 141 | 0 | чисто |
| `backend/testframework/imex/service_extended_test.go` | 172 | 0 | чисто |
| `backend/testframework/imex/service_test.go` | 102 | 0 | чисто |
| `backend/internal/integrations/onec/dto.go` | 130 | 0 | FIXED-OK mask `bearer_token_set` |
| `backend/internal/integrations/onec/export_hooks.go` | 52 | 0 | hubtest · чисто |
| `backend/internal/integrations/onec/handler.go` | 77 | 0 | admin-only · чисто |
| `backend/internal/integrations/onec/service.go` | 250 | 0 | FIXED-OK cipher + ValidateOutboundURL |
| `backend/internal/integrations/onec/settings_store.go` | 59 | 0 | store ciphertext blob · чисто |
| `backend/internal/integrations/onec/user_lookup.go` | 31 | 0 | чисто |
| `backend/testframework/onec/fixtures_test.go` | 26 | 0 | чисто |
| `backend/testframework/onec/handler_test.go` | 82 | 0 | чисто |
| `backend/testframework/onec/service_extended_test.go` | 237 | 0 | FIXED-OK encrypt/SSRF paths |
| `backend/testframework/onec/settings_store_wave101_test.go` | 72 | 0 | чисто |

Файлов в списке: **71**. Прочитано: **71**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W7-3-001 | NEW | P2 | `imex/repository.go:189-193`; call `pdf.go:85` | `ListSprintIssues` **без** `issue.DeletedAtIsNil()` | CSV/ZIP/Excel paths закрыты TASK-120 (W7-A3-008); sprint PDF всё ещё включает задачи из корзины → leakage trash в отчёт / неверные метрики `Задач в спринте` | Добавить `.Where(issue.DeletedAtIsNil())` (как `ListProjectIssuesPlain`); IT: soft-deleted issue не в PDF-списке |
| A3W7-3-002 | NEW | P2 | `github/repository.go:303-323`; `gitlab/repository.go:202-214` | `FindIssueByKey` / `FindIssueByKeyInProject` без `DeletedAtIsNil` | Webhook auto-merge / CreateBranch / link create работают по trashed issue; интеграции soft-delete закрыты (W7-A3-010), issue trash — нет | `Where(issue.DeletedAtIsNil())` на обоих Find*; webhook/branch → NotFound на trash; regression IT |
| A3W7-3-003 | NEW | P2 | `github/preview.go:94-117`; route `handler.go:47-50` | `resolveIssuePreview` по ключу без membership (только `authMw`) | Класс W7-A3-002: любой auth-юзер с угаданным `HUB-N` читает title/status/assignee чужого проекта через `/link-preview` | После FindIssue: `IsProjectMember` (или RBAC project view); denial → generic external/empty; IT foreign key |

## Фокус-чеклист (подтверждено на HEAD)

| Тема | Статус | Evidence |
|------|--------|----------|
| GH webhook HMAC + CTC | **OK** | `webhook_verify.go:54-58` `subtle.ConstantTimeCompare` на MAC digests; empty/bad hex/no secret → 403; no skip |
| GH bind A5-008 | **OK** | FindByRepo only; A5-008 test + mutation note against any-of |
| GL token CTC (digest) | **OK** | `constantTimeEqualDigest` — equal-length sha256; no raw-string length oracle |
| GL bind A5-009 | **OK** | FindByProjectPath; A5-009 test |
| Soft-delete integrations | **OK** | SetDeletedAt + DeletedAtIsNil list/get/find; TASK-097 IT |
| Soft-delete issues (export) | **PARTIAL** | ZIP/CSV/XLSX OK; **sprint PDF gap** → A3W7-3-001 |
| Soft-delete issues (GH/GL find) | **GAP** | → A3W7-3-002 |
| SSRF outbound | **OK** | SafeHTTPClient GH/GL/1С; ValidateOutboundURL gitlab_url + onec webhook + OG preview |
| Secrets at-rest | **OK** | AES-GCM PAT/webhook (GH/GL) + BearerToken (1С); responses mask |
| Secrets in NATS/logs | **OK** | redactPipelineVariables; no AccessToken/Bearer in slog fields; upstream body server-only |
| Body DoS | **OK** | LimitReader webhook 10MB; imex import 25MB; GL API reads capped |
| imex RBAC / PDF project scope | **OK** | A5-002 / A5-010 tests |

## Чисто (подтверждено после remediation)

- **Layering:** Handler→Service→Repository; `*ent.Client` только в Ent adapters / onec settings+lookup.
- **DTO secrets:** `IntegrationResponse` / `SettingsResponse` не отдают plaintext tokens.
- **Dead path ListAllWebhookSecrets:** удалён; verify не перебирает секреты инстанса.
- **excelize:** реальный XLSX, не CSV под xlsx MIME.
- **StructValidator:** github CreateIntegration `validate:"required"` enforced globally.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 0 |
| **Total** | **3** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 W7-A3-001…013) | **13/13** |
| RESIDUAL | 0 |
| NEW | 3 (`A3W7-3-001`…`003`) |
| REGRESS | 0 |

A2 baseline: **1** open residual (`W7-A3-010`) → FIXED-OK · RESIDUAL 0 · NEW 3 · REGRESS 0 · open **3** (P2=3).

**Топ ID:** A3W7-3-001 (sprint PDF trash) · A3W7-3-002 (FindIssue trash → webhook/branch) · A3W7-3-003 (link-preview membership).
