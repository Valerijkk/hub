# TASK-066 — AI scope + auth/rbac/audit residuals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-13,P3-BE-14 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Мутация C (W13 be1 refill retry 2026-07-30): **10/10 PROVEN** · W6-A4-010 closed → `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task066-retry.md`

## [OUTCOME]

Дефекты P3-BE-13/14 закрыты.

## [ROLE]

- Role: be2
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/{ai,auth,rbac,audit}/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-13,P3-BE-14)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-13,P3-BE-14 с мутацией «да» закрыты по реестру
- **Мутация (C):** AI routing scope; UnassignRole/CTC/Upsert guards
- Regression: соответствующий `testframework` / Vitest domain package
- Evidence: вывод тестов, SHA коммитов
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Коммит(ы) Conventional (тип/scope EN, описание RU) + путь задачи в теле
- Статус README

## [OPERATING RULES]

- Guard тестировать на реальном слое, не только mock
- Co-authored-by запрещён

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-BE-13**, **P3-BE-14**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W5-A1-010 | `backend/internal/ai/service.go:277-287` — `requireSystemProvider` (admin Update/Test → user BYOK 404); call sites `:229`, `:250` | service | `TestAdminUpdate_UserScopeProvider_404_TASK066` · `backend/testframework/ai/scope_routing_task066_test.go` | убрать `requireSystemProvider` / принимать `ScopeUser` → PATCH/test user BYOK = 200 → StatusNotFound assert RED |
| W5-A2-012 | `backend/internal/ai/adapter_yandex.go:166+` — `IsYandexProvider` host allowlist (не substring) | helper | `TestIsYandexProvider` · `backend/testframework/ai/adapter_yandex_test.go` | вернуть `strings.Contains(…, "yandex")` → `evil.example/yandex` / `yandexgpt` True → False asserts RED |
| W5-A3-013 | `backend/internal/ai/code_review.go:50-52` (+ `:80-82`) — GitLab URL + `gitlab==nil` → `"GitLab not configured"` | service / code_review | `TestFetchReview_GitLabURL_NilGitLab_TASK066` · `backend/testframework/ai/scope_routing_task066_test.go` | убрать nil-check → GitHub path / иной err → Contains `"GitLab not configured"` RED |
| W5-A3-014 | `backend/internal/ai/issue_lookup.go:45-48` — nil `TrackerIssueLookup`/deps → `"issue lookup not configured"` | service | `TestTrackerIssueLookup_NilSafe` · `backend/testframework/ai/issue_lookup_test.go` | вернуть `("", "", nil)` при nil → Error/Contains assert RED |
| W5-A3-015 | `backend/internal/ai/bot_user.go:36-64` — parallel Create conflict → re-query | repository (Ent) | `TestEnsureBotUser_Parallel_OneUser_TASK066` · `backend/testframework/ai/scope_routing_task066_test.go` | убрать conflict re-query → count>1 / err / разные IDs → Equal(1)/Equal(ids[0]) RED |
| W6-A1-010 | `backend/internal/auth/service_profile.go:45-53` — `CompleteOnboarding` в `WithEntTx` | service + repository (Ent) | `TestCompleteOnboarding_TxRollback_TASK066` · `backend/testframework/auth/onboarding_tx_task066_test.go` | UpdateProfile + Mark вне tx → fail mark оставляет DisplayName → Equal("before") RED |
| W6-A3-012 | `backend/internal/rbac/handler.go:223-232` + `custom_roles.go:187-204` — parse `:roleId` + match current | handler + service (Ent) | `TestUnassignRole_WrongRoleID_404_TASK066` · `backend/testframework/rbac/unassign_invalidate_task066_test.go` | Clear без match roleId → 200 + роль снята → StatusNotFound / Has(PermTaskCreate) RED |
| W6-A3-014 | `backend/internal/rbac/cache.go:12-27`, `:70-85` — Lua SCAN+DEL `InvalidateProject` | cache (Redis) | `TestInvalidateProject_LuaClearsKeys_TASK066` · `backend/testframework/rbac/unassign_invalidate_task066_test.go` | non-atomic SCAN+DEL / skip DEL → key остаётся → `ok==false` RED |
| W6-A4-010 | `backend/internal/audit/handler_alerts.go` — `constantTimeDigestCompare` SHA-256 + CTC; operandLen=`sha256.Size` | handler / helper | `TestAlertWebhook_CTCFixedDigest_TASK066` · `backend/testframework/audit/ctc_upsert_task066_test.go` | raw CTC `([]byte(a),[]byte(b))` + `operandLen=len(a)` → `ExportConstantTimeEqualDigestOperandLen` ≠ `sha256.Size` RED |
| W6-A4-012 | `backend/internal/audit/repository.go:286-314` — Upsert Create+constraint → retry Update | repository (Ent) | `TestUpsertSystemSetting_Concurrent_OneRow_TASK066` · `backend/testframework/audit/ctc_upsert_task066_test.go` | Update→Create без retry → concurrent err / count>1 → NoError/Equal(1) RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W5-A1-010 | Fiber AdminHandler + **mockRepo** (не Ent GetProvider) | **mock≠Ent** |
| W5-A2-012 | pure `IsYandexProvider` | **OK** |
| W5-A3-013 | pure nil gitlab provider | **OK** |
| W5-A3-014 | nil receiver / nil deps | **OK** |
| W5-A3-015 | EntBotUserRepository + sqlite enttest | **OK** |
| W6-A1-010 | EntRepository `WithEntTx` | **OK** |
| W6-A3-012 | Fiber + Ent RBAC fixture | **OK** |
| W6-A3-014 | PermissionCache + miniredis | **OK** (не Ent) |
| W6-A4-010 | pure digest + OperandLen export + handler (mockAuditRepo body) | **OK** — CTC pure + digest-size assert; mock только для side-effect empty |
| W6-A4-012 | Ent UpsertSystemSetting | **OK** |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: W5-A1-010` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A**.
- Class source IDs = **10/10** с dedicated RED (W6-A4-010 closed on retry via OperandLen assert).
- W5-A1-010: primary RED на mockRepo — verifier: желателен Ent/integration twin (flag mock≠Ent).
