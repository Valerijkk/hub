# TASK-120 — 🔌 Git*/1C/imex: secrets RAM + trash filter

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: W7 integrations · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Закрыто: 2026-07-29 · be2 · ветка `fix/integrations/task-120-secrets-imex`
> Mutation PROVEN: 2026-07-30 · qa-auto2 · `.planning/audits/total-2026-07/MUTATION-2026-07-30-W10-qa-auto2.md` (4/4)

## [OUTCOME]

Нет ListAllWebhookSecrets RAM blast; GitLab body redact; 1C token encrypt; imex trash filter.

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/{github,gitlab,imex,onec}/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - plaintext tokens в settings
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: W7-A3-004/006/007/008 (4)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 4 ID + tests
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record

## [MUTATION CRITERIA] — verifier-ready

> W7 slot 3 / qa-auto1 · criteria only (no mutate) · 2026-07-30

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| W7-A3-004 | `github/webhook_verify.go:28-46` — VerifyWebhookSignature any-of по `ListAllWebhookSecrets` вместо `FindIntegrationByRepo` (+ gitlab token analog) | **Канон:** `TestHandler_HandleWebhook_RejectsForeignSecret_A5_008` (+ gitlab `TestHandler_*_RejectsForeign*_A5_009`). `TestEntRepository_WebhookSecretBoundToRepo_W7A3004` — **weak** (только FindByRepo) | service (verify) |
| W7-A3-006 | `gitlab/service.go:417-418` / `:520` — вернуть `respBody` в `apperrors.NewInternal(...)` клиенту | `TestTriggerPipeline_UpstreamBodyNotLeaked_W7A3006`; `TestCreateBranch_UpstreamBodyNotLeaked_W7A3006` | service |
| W7-A3-007 | `integrations/onec/service.go:74-78` — `encryptToken` return plaintext (`cipher==nil` path / skip EncryptString) | `TestService_BearerTokenEncryptedAtRest_W7A3007` | service |
| W7-A3-008 | `imex/repository.go:152-158` — убрать `issue.DeletedAtIsNil()` из `ListProjectIssuesPlain` | `TestListProjectIssuesPlain_ExcludesTrash_W7A3008` | repository |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 004 | mockRepo + handler HMAC | Guard в Verify* — mock OK для binding; Ent W7A3004 **не** ловит restore ListAll |
| 006 | httptest upstream + service | Body redact — OK |
| 007 | Ent system_setting + cipher | At-rest encrypt — Ent OK |
| 008 | Ent ListProjectIssuesPlain | Soft-delete filter — **Ent OK** |

`layer_mismatch IDs: (none)` — канон мутации 004 = A5-008/A5-009, не W7A3004

### Gaps

- **W7-A3-004:** `TestEntRepository_WebhookSecretBoundToRepo_W7A3004` остаётся зелёным при restore any-of — не использовать как PROVEN gate.
- Evidence в теле TASK-120 минимален (закрыто в README/ветке) — W8 сверить hubtest green.
