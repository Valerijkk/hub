# TASK-121 — 🛒 Marketplace/status/docs exposure gates

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта  
> Источник: W7 marketplace · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Mutation W13 be2: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task121.md` — **Закрыта** (004 PROVEN mid-window; 001 N/A layer_mismatch)

## [OUTCOME]

/api/docs и status заgated; marketplace bounds/PII/price validation.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/{marketplace,status,dev}/**
  - swagger exposure
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - DEV_MODE в prod
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: W7-A5-001..006 (6)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 6 ID + tests
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

## [EVIDENCE — be1 2026-07-29]

| ID | Статус | Evidence |
|----|--------|----------|
| W7-A5-001 | FIXED | `/api/docs` + `/docs/doc.json` → `requireAuth` + `RequireAdmin` в `wiring_core.go`; тест `TestAPIDocs_RequiresAuthAdmin_W7A5001` |
| W7-A5-002 | FIXED | публичный `/api/status` — агрегат `hub`; детали `/api/status/details` за auth; тесты status handler/service |
| W7-A5-003 | FIXED | `ListMine`/`ListPending`/`ListApproved` — page/per_page cap 100 + `has_more`; `TestListMine_PaginationCap_W7A5003` |
| W7-A5-004 | FIXED | `UpdateApprovedListingToPending` — один UPDATE; `TestUpdateApproved_AtomicRemoderation_W7A5004` |
| W7-A5-005 | FIXED | публичный `author_name` без email; `TestAuthorName_NoEmailOnPublicCatalog_W7A5005` |
| W7-A5-006 | FIXED | `price_cents≥0` + currency allowlist; `TestValidatePriceCurrency_W7A5006` |

Команда: `go test -tags hubtest -race -count=1 ./testframework/marketplace/ ./testframework/status/ ./testframework/dev/` → PASS

## [MUTATION — qa-auto2 2026-07-30 W12]

| ID | Verdict | Note |
|----|---------|------|
| W7-A5-001 | N/A | layer_mismatch; source-assert wiring_core OK; stub не RED от prod BREAK |
| W7-A5-002 | PROVEN | GetPublicStatus→getFull + details без authMw |
| W7-A5-003 | PROVEN | drop Limit ListListingsByAuthor |
| W7-A5-004 | GAP_NOTED | split Update+SetStatus → final-state PASS (weak) |
| W7-A5-005 | PROVEN | email fallback public |
| W7-A5-006 | PROVEN | drop price/currency checks |

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W12-qa-auto2.md`  
Итог слота: **Открыта** — 004 atomicity не PROVEN.

## [MUTATION — be2 2026-07-30 W13 refill]

| ID | Verdict | Note |
|----|---------|------|
| W7-A5-001 | N/A | layer_mismatch; source-assert OK |
| W7-A5-002 | PROVEN | GetPublicStatus→getFull + details без authMw |
| W7-A5-003 | PROVEN | drop Limit ListListingsByAuthor |
| W7-A5-004 | PROVEN | Ent hook mid-window: split → approved+newTitle → RED |
| W7-A5-005 | PROVEN | email fallback public |
| W7-A5-006 | PROVEN | drop price/currency checks |

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task121.md`  
Итог слота: **Закрыта** — 5/6 PROVEN · 001 N/A.

## [MUTATION CRITERIA] — verifier-ready

> W7 slot 3 / qa-auto1 · criteria only (no mutate) · 2026-07-30

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| W7-A5-001 | `cmd/server/wiring_core.go:47-57` — снять `requireAuth`/`RequireAdmin` с `/docs` (+ `/docs/doc.json`) | `TestAPIDocs_RequiresAuthAdmin_W7A5001` | wiring (authz) |
| W7-A5-002 | `status/service.go:31-40` — `GetPublicStatus` вернуть `getFull` components; **и/или** `handler.go:22` — `/status/details` без `authMw` | `TestHandler_GetStatus_PublicAggregate_W7A5002`; `TestService_GetPublicStatus_NoComponents_W7A5002` | service+handler |
| W7-A5-003 | `marketplace/repository.go:39-46` / List* — убрать `Limit(perPage)` (оставить Offset) | `TestListMine_PaginationCap_W7A5003` (+ Pending/Approved HasMore) | repository |
| W7-A5-004 | `marketplace/repository.go:164-174` — split: Update fields + отдельный `SetListingStatus(approved→pending)` | `TestUpdateApproved_AtomicRemoderation_W7A5004` | repository |
| W7-A5-005 | `marketplace/dto.go:140-143` — public `displayName`: fallback на `u.Email` при пустых name | `TestAuthorName_NoEmailOnPublicCatalog_W7A5005` | dto |
| W7-A5-006 | `marketplace/service.go:383-388` — убрать `priceCents < 0` / `isValidCurrency` checks | `TestValidatePriceCurrency_W7A5006` | service |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 001 | **stub** Fiber middleware (не `wiring_core`) | **layer_mismatch** — мутация prod wiring ≠ RED этого теста |
| 002 | Handler RegisterRoutes + GetPublicStatus | Auth на details + aggregate — OK (не wiring) |
| 003 | Ent List* + normalizePage | Cap в repo — Ent OK |
| 004 | Ent UpdateApprovedListingToPending | Ent hook mid-window (approved+newTitle) — kills split |
| 005 | Ent catalog + displayName(allowEmail=false) | PII — OK |
| 006 | service validateListingInput | Allowlist — OK |

`layer_mismatch IDs: W7-A5-001`

### Gaps

- **W7-A5-001:** N/A layer_mismatch — source-assert на `wiring_core.go` OK; stub Fiber ≠ wireCore (не блокирует Закрыта).
- **W7-A5-004:** CLOSED W13 — Ent `UpdateOne` hook detects mid-window approved+newTitle on split.
