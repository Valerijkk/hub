# TASK-121 — 🛒 Marketplace/status/docs exposure gates

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: На проверке  
> Источник: W7 marketplace · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

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
