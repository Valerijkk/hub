# TASK-098 — 📐 Intranet FR-INTR-07/08 author-delete + polls

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: **Исправлена**  
> Источник: W7-A2-001 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-28 · be1 · ветка `fix/intranet/task-098-author-delete`

## [OUTCOME]

Автор может удалять свою новость; polls/news_poll* восстанавливаются/чистятся по FR-INTR-07/08 (после ADR U03).

## [ROLE]

- Role: product→be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/intranet/**
  - testframework intranet
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - обход soft-delete 30d
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **апрув ADR/migrate не требуется** (owner 2026-07-28) — ADR+код+тесты в одном цикле

## [REFERENCES]

- Finding IDs: W7-A2-001 U03 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-INTRANET-U03-NEWS-POLL-AUTHOR-DELETE-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR U03 ✅
  - author-delete + poll path ✅
  - mutation IT ✅
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record — **снято owner 2026-07-28**

## Evidence (be1 2026-07-28)

| ID | Вердикт | Evidence |
|----|---------|----------|
| W7-A2-001 / P1-BE-U03 | **FIXED** | ADR-INTRANET-U03; `DeleteNews` author\|\|admin; `news_poll*`; restore counters; `task098_author_delete_poll_test.go` |

**Тесты:**
```
go test -tags hubtest -race -count=1 -timeout=180s ./testframework/intranet/
→ ok (2.491s)
```

**Мутация:** убрать author-ветку в `DeleteNews` / `HasNewsPostWith(DeletedAtIsNil)` в `GetNewsPollByPostID` → краснеют `TestDeleteNews_AuthorCanDeleteOwn_Task098` / `TestNewsPoll_SoftDeleteRestoreCounters_Task098`.
