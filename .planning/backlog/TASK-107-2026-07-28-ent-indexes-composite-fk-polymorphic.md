# TASK-107 — 🔒 Ent indexes/FK/polymorphic IDs

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта  
> Источник: A2W8 indexes · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-28 · be1 · ветка `fix/ent/task-107-indexes-fk`  
> Mutation prove: 2026-07-30 · qa-auto1 · 8/8 PROVEN · `.planning/audits/total-2026-07/MUTATION-2026-07-30-W11-qa-auto1.md`

## [OUTCOME]

Composite unique/indexes и polymorphic ID-типы из A2W8-1/2/3 исправлены (migrate + тесты).

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/ent/schema/**
  - migrate owner approve
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - без approve SQL
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **Owner 2026-07-28 — ADR/migrate approve НЕ требуется; migrate для тестов OK; довести до «Исправлена».**

## [REFERENCES]

- Finding IDs: A2W8-1-004/005, A2W8-2-005/007/008/009, A2W8-3-002/006 (8)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код
- Migrate: `backend/migrations/20260728_task107_ent_indexes_fk_polymorphic.sql`
- Evidence tests: `backend/testframework/ent/schema_task107_indexes_fk_test.go`

## [ACCEPTANCE]

- Done criteria:
  - indexes/FK/UUID fields
  - enttest/IT
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

### Checklist ID → evidence

| ID | Fix | Evidence |
|----|-----|----------|
| A2W8-1-004 | Board `updated_at`/`deleted_at` + `index.Edges("project").Fields("deleted_at")`; List/Get filter soft-delete | schema + `TestSchema_TASK107_*` |
| A2W8-1-005 | `AIUserPreference.default_provider` FK edge → `AIProvider` (ON DELETE SET NULL) | schema + runtime orphan FK fail |
| A2W8-2-005 | `crmcontact` email index; `crmdeal` owner+`deleted_at` | schema markers |
| A2W8-2-007 | GitLab `webhook_secret` Required/NotEmpty + service/DTO | schema + runtime + gitlab tests |
| A2W8-2-008 | GitHub/GitLab link `version` + unique `(integration, external_id, link_type)` | schema + runtime unique fail |
| A2W8-2-009 | HubBase owner+deleted_at; HubColumn unique `(table,name)` + table+deleted_at; HubRecord table+deleted_at | schema + runtime unique fail |
| A2W8-3-002 | Invite `version` + `revoked_at`; soft revoke in project repo | schema + runtime + List/Find filter |
| A2W8-3-006 | Notification `entity_id` UUID + `version`; MarkRead `AddVersion` | schema + runtime + notification tests |

## [MUTATION CRITERIA] — verifier-ready

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| A2W8-1-004 | `backend/ent/schema/board.go:37-43,67-69` — убрать `updated_at`/`deleted_at` и/или `index.Edges("project").Fields("deleted_at")` | `TestSchema_TASK107_Markers` (+ Runtime: soft-delete Board) | schema (+ Ent runtime) |
| A2W8-1-005 | `backend/ent/schema/aiuserpreference.go:22-25,52-55` — убрать FK edge `default_provider` / Field | `TestSchema_TASK107_Markers` + `TestSchema_TASK107_RuntimeConstraints` (orphan UUID) | schema + Ent FK |
| A2W8-2-005 | `crmcontact.go:70` `index.Fields("email")`; `crmdeal.go` `index.Edges("owner").Fields("deleted_at")` — убрать | `TestSchema_TASK107_Markers` | schema (source) |
| A2W8-2-007 | `gitlabintegration.go` `webhook_secret` — снова Optional / без NotEmpty | `TestSchema_TASK107_Markers` + Runtime empty secret fail | schema + Ent |
| A2W8-2-008 | `githublink.go` / `gitlablink.go` — убрать unique `(integration, external_id, link_type)` | Markers + Runtime duplicate link | schema + Ent unique |
| A2W8-2-009 | `hubcolumn.go` — убрать `index.Edges("table").Fields("name").Unique()` (и/или owner/table+deleted_at indexes) | Markers + Runtime duplicate column name | schema + Ent unique |
| A2W8-3-002 | `invite.go` — убрать `version` / `revoked_at` | Markers + Runtime invite version/revoked | schema + Ent |
| A2W8-3-006 | `notification.go:37-40,56-57` — `entity_id` снова string / убрать `version` | Markers + Runtime UUID entity_id + AddVersion | schema + Ent |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 1-004 / 2-005 | schema source (`os.ReadFile`) | Index-only без runtime assert — **schema OK**; 2-005 **нет** Runtime unique |
| 1-005 / 2-007 / 2-008 / 2-009 / 3-002 / 3-006 | Markers + `enttest` RuntimeConstraints | Guard в Ent schema — **Ent OK** |

`layer_mismatch IDs: (none)`

### Gaps

- **A2W8-2-005:** только Markers (index); Runtime не ловит снятие email/owner+deleted_at index на sqlite.
- Markers = source-guard: ломает при удалении строки из schema; не доказывает migrate SQL drift без `20260728_task107_*.sql` assert (он есть в Markers).
- После schema-мутации нужен `go generate` / rebuild — иначе `MUTATION_FAILED_TO_BUILD` (как TASK-170 class).

## [OUTPUT SCHEMA]

- Checklist ID → evidence — выше
- Test commands + PASS — `go test -tags hubtest -race -count=1 ./testframework/{ent,notification,gitlab,project,ai,tracker,bases,crm,github}/` PASS
- Mutation note: убрать unique/NotEmpty/FK/UUID из schema → `TestSchema_TASK107_*` краснеет
- README status update — Исправлена

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record — **owner снял требование 2026-07-28**
