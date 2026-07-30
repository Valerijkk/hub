# TASK-065 — Platform HTTP + obs residuals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P3** · Статус: Закрыта
> Источник: TASK-008 W12 · Classes: P3-BE-11,P3-BE-12 · Реестр: `.planning/audits/total-2026-07/W12-REGISTRY-P3-BACKEND.md`
> Mutation C (W13 slot 4 refill / qa-auto1): `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-qa1-task065.md` — **7 PROVEN · 2 N/A · 1 GAP_NOTED · 0 NOT_PROVEN**

## [OUTCOME]

Дефекты P3-BE-11/12 закрыты.

## [ROLE]

- Role: be1
- Responsibility: фикс + тесты на слое guard
- Handoff to: стадия C (мутация независимым агентом)

## [SCOPE AND CONSTRAINTS]

- Files allowed: backend/internal/platform/**; `.planning/backlog/README.md`; этот файл
- Do not touch: несвязанные модули; secrets; DEV_MODE в CI
- Constraints: Docker + `-tags hubtest` для BE; Vitest для FE; один логический коммит на подкласс при необходимости
- Approval record: миграции Ent — только с явным согласованием владельца

## [REFERENCES]

- Реестр W12 (P3-BE-11,P3-BE-12)
- Сырые находки: отчёты `A*-*.md` стадии A
- Source precedence: код > отчёт

## [ACCEPTANCE]

- Done criteria: все ClassID из P3-BE-11,P3-BE-12 с мутацией «да» закрыты по реестру
- **Мутация (C):** OG/Validate/CORS + shutdown/metrics probes
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

> W13 slot 4 refill / qa-auto1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified · Classes **P3-BE-11**, **P3-BE-12**

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| W4-A1-008 | `backend/internal/platform/http/opengraph.go:103-108` — og:image только после `ValidateOutboundURL==nil` | HTTP parse (OG) | `TestParseOpenGraphHTML_ClearsUnsafeImage_P3BE11` · `backend/testframework/platform/http/opengraph_test.go` | assign content без Validate → private/file ImageURL → RED |
| W4-A1-009 | `backend/internal/platform/http/safedialer.go:117-119` — IP-literal + isBlockedIP → ErrBlockedAddress | HTTP ValidateOutboundURL | `TestValidateOutboundURL` / blocked_link-local_IP_literal (+ loopback) · `backend/testframework/platform/http/safedialer_test.go` | убрать IP-literal block → `http://169.254.169.254/` Validate OK → ErrorIs RED |
| W4-A1-011 | `backend/internal/platform/middleware/cors.go:71-73` — `!isAbsoluteCORSOrigin` → reject | middleware CORS | `TestParseCORSOrigins_TrimsAndRejectsWildcardAndEmpty` (null/path/query/ftp) · `backend/testframework/platform/middleware/middleware_test.go` | ослабить isAbsoluteCORSOrigin → bad origin в list → wantOrigins RED |
| W4-A2-006 | `backend/internal/platform/nats/client.go:93-104` — Close через Conn.Drain()+timeout | NATS client | `TestClient_Close_Drains_P3BE12` · `backend/testframework/platform/nats/client_integration_test.go` | голый Conn.Close без Drain — **weak**: тест может остаться GREEN (IsClosed/timeout only) |
| W4-A2-009 | `backend/internal/platform/outbox/worker.go:87-91` — `select ctx.Done()` mid-batch | outbox worker | `TestProcessBatch_CancelMidBatch_Stops_P3BE12` · `backend/testframework/platform/outbox/outbox_test.go` | убрать ctx.Done check → processed==5 после cancel → RED (ожидается 1) |
| W4-A3-005 | `backend/internal/platform/observability/otel.go:83-84` — TraceContext{} без Baggage | observability | `TestInitTracing_PropagatorWithoutBaggage_P3BE12` · `backend/testframework/platform/observability/baggage_task065_test.go` | добавить Baggage{} в InitTracing — **weak/layer_mismatch**: тест вручную ставит TraceContext, не бьёт InitTracing+OTLP |
| W4-A3-006 | `backend/internal/platform/observability/timing.go:34` — success → `slog.DebugContext("operation completed")` | observability Timed | **GAP** — нет Exact RED на log level | mutate Debug→Info → **GAP** until hubtest slog capture |
| W4-A3-008 | `backend/internal/platform/observability/operation_label.go:27-51` + `metrics.go:34` — unknown→`other` | observability metrics | `TestSanitizeOperationLabel_UnknownToOther_P3BE12`; `TestRecordOperation_UnknownBucketsToOther_P3BE12` · `backend/testframework/platform/observability/operation_label_task065_test.go` | ослабить allowlist → bare UUID/`user-controlled` не `"other"` → RED |
| W4-A4-012 | `backend/internal/platform/health/probes.go:35,80-81` — Ready `WithTimeout(c.Context(),…)` | health Ready | `TestReady_ProbeParentIsRequestContext_P3BE12` · `backend/testframework/platform/health/probes_test.go` | probeParentContext → Background → Export ≠ c.Context() RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| W4-A1-008, W4-A1-009, W4-A1-011 | pure Validate/parse/CORS | **OK** |
| W4-A2-006 | real NATS Close | **layer_mismatch** (weak Drain kill) |
| W4-A2-009 | Ent outbox + cancel | **OK** |
| W4-A3-005 | manual SetTextMapPropagator | **layer_mismatch** |
| W4-A3-006 | — | **GAP** |
| W4-A3-008, W4-A4-012 | export sanitize / Fiber probe | **OK** |

`layer_mismatch IDs: W4-A2-006, W4-A3-005` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)` · `GAP IDs: W4-A3-006`

### Gaps

- Docs-only: **N/A**.
- Strong runnable: **7/9**; layer_mismatch weak: **2**; GAP: **W4-A3-006**.
