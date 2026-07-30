# TASK-118 — 💼 v4 CRM/HR/calendar/coffee round residuals

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта
> Источник: W7 v4 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Mutation: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W10-be1.md` · 5/5 PROVEN (be1 W10 slot 2)

## [OUTCOME]

Coffee round tx; CRM trash types; HR leave scope; calendar unbounded lists — закрыты.

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/{crm,hr,calendar,intranet,accounting}/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - U03/U05 ADR tasks
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: W7-A2-004/006/009/010/013 (5)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 5 ID + tests
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [MUTATION CRITERIA] — verifier-ready

> Сверено по HEAD 2026-07-30 · be2 W7 slot 2 · W7-A2-004/006/009/010/013 · статус **На проверке**

| ID | что ломать (path:line + фрагмент) | какой тест | слой |
|----|-----------------------------------|------------|------|
| W7-A2-004 | `intranet/coffee_repository.go:73-103` — создать round **вне** tx / `Commit` сразу после `CoffeeRound.Create` до pairs (mid-fail → orphan round) | `TestCreateCoffeeRoundWithPairs_RollbackOnPairFail_W7A2004` (+ AtomicOK) | repository (Ent tx) |
| W7-A2-006 | `audit/repository.go:613` / `:641` — убрать `crmcontact/crmdeal.DeletedAtGT(cutoff)` из ListDeletedCRM*; **или** убрать SoftDelete/Restore/Purge CRM methods из wiring service switch | `TestRepository_SoftDeleteAndRestoreCRMContact_W7A2006` (+ Deal / `PurgeExpiredCRM_W7A2006` / RetentionHides CRM) | audit repository (Ent) |
| W7-A2-009 | `hr/repository.go:82-95` — убрать `HasRequesterWith(requesterPred)` (org-wide approved leave) | `TestEntRepository_ListApprovedLeaveInRange_ViewerScope_W7A2009` | hr repository (Ent) |
| W7-A2-010 | `calendar/repository.go:70` / `:87` / `:104` / `:128` — убрать `.Limit(MaxCalendarList)` (и/или `WithProject` на scheduled) | `TestListScheduledCalls_CapAndWithProject_W7A2010` (+ `TestListIssuesDueInRange_Cap_W7A2010`) | calendar repository (Ent) |
| W7-A2-013 | `crm/repository.go:78` / `:167` — убрать `.Limit(MaxCRMList)` с ListContacts/ListDeals | `TestEntRepository_ListContactsAndDeals_Caps_P3BE06` | crm repository (Ent) |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Комментарий |
|----|-----------|-------------|
| 004 | Ent sqlite tx + FK fail mid-pair | **Ent OK** |
| 006 | Ent audit trash CRM | **Ent OK** (service switch покрыт SoftDelete/Restore CRM IT) |
| 009 | Ent HR leave scope | **Ent OK** |
| 010 | Ent calendar Limit(+WithProject) | **Ent OK** |
| 013 | Ent CRM Limit | **Ent OK** |

`layer_mismatch IDs: (none)`

### Gaps

- TASK-118 **без** Evidence checklist в файле — W8 verifier подтверждает green + mutation.
- W7-A2-006 RetentionHides CRM (`TestRepository_CRMContact_RetentionHidesOlderThan30Days`) — доп. якорь на `DeletedAtGT`; канон soft/restore/purge — SoftDeleteAndRestore* / PurgeExpiredCRM.

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
