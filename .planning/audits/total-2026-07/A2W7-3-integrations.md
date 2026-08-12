# A2W7-3 — github+gitlab+imex+onec (Stage A2)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Baseline: `A7-3-integrations.md` · Assignments: `W7-ASSIGNMENTS.md` §A3

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/github/converters.go` | 33 | 0 | чисто |
| `backend/internal/github/dto.go` | 105 | 0 | чисто |
| `backend/internal/github/export_hooks.go` | 95 | 0 | чисто |
| `backend/internal/github/handler.go` | 352 | 1 | замечания |
| `backend/internal/github/pr_review.go` | 105 | 0 | чисто |
| `backend/internal/github/preview.go` | 151 | 0 | чисто |
| `backend/internal/github/repository.go` | 509 | 1 | замечания |
| `backend/internal/github/service.go` | 576 | 1 | замечания |
| `backend/internal/github/webhook_verify.go` | 74 | 0 | чисто |
| `backend/testframework/github/converters_test.go` | 66 | 0 | чисто |
| `backend/testframework/github/ent_repository_test.go` | 290 | 0 | чисто |
| `backend/testframework/github/handler_extended_test.go` | 141 | 0 | чисто |
| `backend/testframework/github/handler_validation_test.go` | 158 | 0 | чисто |
| `backend/testframework/github/handler_wave101_test.go` | 58 | 0 | чисто |
| `backend/testframework/github/handler_wave103_test.go` | 126 | 0 | чисто |
| `backend/testframework/github/pr_review_test.go` | 107 | 0 | чисто |
| `backend/testframework/github/preview_test.go` | 131 | 0 | чисто |
| `backend/testframework/github/preview_wave103_test.go` | 57 | 0 | чисто |
| `backend/testframework/github/repository_helpers_test.go` | 54 | 0 | чисто |
| `backend/testframework/github/repository_wave101_test.go` | 59 | 0 | чисто |
| `backend/testframework/github/repository_wave103_test.go` | 134 | 0 | чисто |
| `backend/testframework/github/service_http_test.go` | 19 | 0 | чисто |
| `backend/testframework/github/service_mock_test.go` | 587 | 0 | чисто |
| `backend/testframework/github/service_test.go` | 154 | 0 | чисто |
| `backend/testframework/github/service_wave104_test.go` | 69 | 0 | чисто |
| `backend/testframework/github/service_wave115_test.go` | 38 | 0 | чисто |
| `backend/testframework/github/webhook_secret_binding_a5008_test.go` | 89 | 0 | чисто |
| `backend/internal/gitlab/dto.go` | 133 | 0 | чисто |
| `backend/internal/gitlab/export_hooks.go` | 49 | 0 | чисто |
| `backend/internal/gitlab/handler.go` | 298 | 0 | чисто |
| `backend/internal/gitlab/mr_review.go` | 177 | 0 | чисто |
| `backend/internal/gitlab/repository.go` | 286 | 1 | замечания |
| `backend/internal/gitlab/service.go` | 616 | 1 | замечания |
| `backend/internal/gitlab/webhook_verify.go` | 67 | 1 | замечания |
| `backend/testframework/gitlab/converters_test.go` | 28 | 0 | чисто |
| `backend/testframework/gitlab/ent_repository_test.go` | 229 | 0 | чисто |
| `backend/testframework/gitlab/handler_extended_test.go` | 152 | 0 | чисто |
| `backend/testframework/gitlab/handler_test.go` | 158 | 0 | чисто |
| `backend/testframework/gitlab/handler_wave115_test.go` | 37 | 0 | чисто |
| `backend/testframework/gitlab/mr_review_test.go` | 78 | 0 | чисто |
| `backend/testframework/gitlab/service_http_test.go` | 301 | 0 | чисто |
| `backend/testframework/gitlab/service_mock_test.go` | 276 | 0 | чисто |
| `backend/testframework/gitlab/service_test.go` | 74 | 0 | чисто |
| `backend/testframework/gitlab/webhook_token_binding_a5009_test.go` | 84 | 0 | чисто |
| `backend/internal/imex/dto.go` | 47 | 0 | чисто |
| `backend/internal/imex/export_hooks.go` | 52 | 0 | чисто |
| `backend/internal/imex/handler.go` | 247 | 1 | замечания |
| `backend/internal/imex/jira_import.go` | 357 | 0 | чисто |
| `backend/internal/imex/pdf.go` | 213 | 1 | замечания |
| `backend/internal/imex/repository.go` | 197 | 1 | замечания |
| `backend/internal/imex/service.go` | 368 | 1 | замечания |
| `backend/testframework/imex/handler_flow_test.go` | 225 | 0 | чисто |
| `backend/testframework/imex/jira_import_flow_test.go` | 97 | 0 | чисто |
| `backend/testframework/imex/jira_import_test.go` | 102 | 0 | чисто |
| `backend/testframework/imex/pdf_export_test.go` | 182 | 0 | чисто |
| `backend/testframework/imex/pdf_scope_a5010_ent_test.go` | 121 | 0 | чисто |
| `backend/testframework/imex/rbac_a5002_test.go` | 139 | 0 | чисто |
| `backend/testframework/imex/repository_pdf_test.go` | 78 | 0 | чисто |
| `backend/testframework/imex/repository_test.go` | 141 | 0 | чисто |
| `backend/testframework/imex/service_extended_test.go` | 172 | 0 | чисто |
| `backend/testframework/imex/service_test.go` | 102 | 0 | чисто |
| `backend/internal/integrations/onec/dto.go` | 130 | 1 | замечания |
| `backend/internal/integrations/onec/export_hooks.go` | 51 | 0 | чисто |
| `backend/internal/integrations/onec/handler.go` | 77 | 0 | чисто |
| `backend/internal/integrations/onec/service.go` | 217 | 0 | чисто |
| `backend/internal/integrations/onec/settings_store.go` | 59 | 1 | замечания |
| `backend/internal/integrations/onec/user_lookup.go` | 31 | 0 | чисто |
| `backend/testframework/onec/fixtures_test.go` | 26 | 0 | чисто |
| `backend/testframework/onec/handler_test.go` | 82 | 0 | чисто |
| `backend/testframework/onec/service_extended_test.go` | 237 | 0 | чисто |
| `backend/testframework/onec/settings_store_wave101_test.go` | 72 | 0 | чисто |

Файлов в списке: **71**. Прочитано: **71**. Пропущено: **0**.

## Находки

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект / статус |
|----|-----|-----|-------------|---------|------------------------|
| W7-A3-001 | FIXED-OK | P1 | `gitlab/service.go:423-442` | Pipeline variables plaintext в NATS | Закрыто TASK-024: `redactPipelineVariables` |
| W7-A3-002 | FIXED-OK | P1 | `github/service.go:472-489` | ListGitHubLinks без membership | Закрыто TASK-024: `IsProjectMember` |
| W7-A3-003 | FIXED-OK | P2 | `gitlab/webhook_verify.go:44-54` | CTC length-oracle на сырых строках | Закрыто TASK-045: sha256 digests + CTC |
| W7-A3-004 | FIXED-OK | P2 | `github/repository.go` / `gitlab/repository.go` ListAllWebhookSecrets | Decrypt всех секретов в RAM; dead path | Закрыто TASK-120: API+cache удалены; verify = FindByRepo/Path |
| W7-A3-005 | FIXED-OK | P2 | `github/dto.go` validate tags + Bind | Пустые repo/token/secret не rejected | Закрыто P2-BE-01: `StructValidator` в `wiring_fiber.go` исполняет `validate:"required"` |
| W7-A3-006 | FIXED-OK | P2 | `gitlab/service.go` | Upstream `respBody` в client error | Закрыто TASK-120: body только в slog; клиенту HTTP status |
| W7-A3-007 | FIXED-OK | P2 | `onec/service.go` | BearerToken plaintext в `system_setting` | Закрыто TASK-120: AES-GCM via `integrationCipher` |
| W7-A3-008 | FIXED-OK | P2 | `imex/repository.go` ListProjectIssuesPlain | Без `DeletedAtIsNil` (CSV path фильтрует) | Закрыто TASK-120: trash filter на ZIP path |
| W7-A3-009 | FIXED-OK | P2 | `imex/service.go:247+` | Excel = CSV под xlsx MIME | Закрыто TASK-045: excelize XLSX |
| W7-A3-010 | **FIXED** (TASK-097) | P2 | `github/repository.go`; `gitlab/repository.go` | Soft-delete integrations | github/gitlab soft_delete_task097 |
| W7-A3-011 | FIXED-OK | P3 | github service `_ = LogActivity` / Publish | Silent swallow | Закрыто TASK-057 (slog path) |
| W7-A3-012 | FIXED-OK | P3 | `imex/pdf.go` byte truncate | Mid-rune UTF-8 | Закрыто TASK-060 |
| W7-A3-013 | FIXED-OK | P3 | `imex/repository.go:83` First() без Order | Нестабильный default status | Закрыто TASK-067: Order SortOrder |

## Чисто (инварианты держат)

| Зона | Статус |
|------|--------|
| GH/GL webhook bind + no-skip | Держит (A5-008/009) |
| SSRF SafeHTTP / ValidateOutboundURL | Держит |
| Body LimitReader | Держит |
| GH/GL token at-rest AES-GCM | Держит (+ 1С BearerToken — TASK-120) |
| imex RBAC / PDF project scope | Держит |

## Counts

| Метрика | n |
|---------|--:|
| Файлов | 71 |
| Open (RESIDUAL) | **1** |
| FIXED-OK | 12 |
| NEW | 0 |
| REGRESS | 0 |
| Open by sev | P0:0 · P1:0 · P2:1 · P3:0 |
