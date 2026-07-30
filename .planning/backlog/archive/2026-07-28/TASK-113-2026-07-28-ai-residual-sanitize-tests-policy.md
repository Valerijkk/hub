# TASK-113 — 🔒 AI residual: sanitize/policy/hollow tests

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-29-W6  
> Источник: A2W5 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-29 · be1 · evidence: `go test -tags hubtest -race ./testframework/ai/ ./testframework/search/` PASS

## [OUTCOME]

W5 open вне TASK-085/092 закрыты: sanitize errors, content-policy, bot leak, hollow tests.

## [CHECKLIST — FIXED]

| ID | Evidence |
|----|----------|
| A2W5-A1-003 | `client_errors.go` mapProviderMutateError; admin/user handlers; `task113_sanitize_handlers_test.go` |
| A2W5-A1-004 | `sanitizeProviderTestError` в TestProvider; тест NoUpstreamLeak |
| A2W5-A1-005 | Create admin/user → NotContains raw api_key |
| A2W5-A2-006 | NFKC+ZW denylist RU/EN; `TestCheckContentPolicy_ZWAndNFKCBypass_TASK113` |
| A2W5-A2-007 | RPM fail-closed + ZRem; `TestRedisProviderRPMLimiter_FailClosedAndNoZSETInflate_TASK113` |
| A2W5-A2-008 | loopback/link-local SSRF; `TestOpenAIAdapter_ChatCompletion_LoopbackAndLinkLocal_TASK113` |
| A2W5-A3-004 | `/ai docs` fixed string + slog; messenger_bot_flow assert |
| A2W5-A3-006 | EnsureBotUser conflict retry; parallel test |
| A2W5-A3-007 | issue_lookup foreign NotFound |
| A2W5-A3-008 | audit details без prompt keys |
| A2W5-A3-009 | GetUserProvider assert.Empty(APIKeyEnc) |
| A2W5-A4-003 | ErrSearchUnavailable total outage; `TestSearch_TotalOutageHardError_TASK113` |
| A2W5-A5-001 | hollow → Postgres TASK037 delegate |
| A2W5-A5-002 | soft-delete runtime + test_case/project |
| A2W5-A5-003 | static Contains = secondary; primary = runtime |
| A2W5-A5-004 | канон EmptyProjectIDs; дубли wave убраны |
| A2W5-A5-005 | один DeleteSavedFilter NotFound (wave109) |
| A2W5-A5-006 | sqlite Empty ≠ NF-03; missing tables → ErrSearchUnavailable |

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/ai/**
  - search tests hollow
  - testframework/ai search
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - Ollama ADR (TASK-101)
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W5-A1-003..005, A2-006..008, A3-004/006..009, A4-003, A5-001..006 (18)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 18 ID checklist + tests
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
