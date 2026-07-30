# TASK-095 — 📐 Ent VersionMixin: ADR + rollout version

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта · MUTATION-2026-07-29-W5  
> Источник: A2W8 VersionMixin · ADR-first · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-28 · be1 · ветка `fix/ent/task-095-version-mixin`

## [OUTCOME]

Все mutable Ent-схемы из класса A2W8 VersionMixin имеют field version + optimistic lock в update-путях (после принятого ADR P1-BE-U06).

## [ROLE]

- Role: arch→be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/ent/schema/**
  - Atlas migrate (только после owner approve)
  - `.planning/decisions/` ADR draft
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - генерированный ent/* кроме schema; прод без approve migrate
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: ✅ owner 2026-07-28 — ADR Accepted, migrate apply разрешён, unclear-12 exclusion

## [REFERENCES]

- Finding IDs: A2W8-1-001/002/006, A2W8-2-001..003/006, A2W8-3-007, A2W8-4-001/003, A2W8-5-001..007, A2W1-A4-001, W7-A2-003 (19 rows)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-ENT-VERSION-MIXIN-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ✅ ADR Accepted владельцем
  - ✅ version на схемах класса + VersionEQ в сервисах (sprint, intranet news, project/channel/comment pilot)
  - ✅ migrate SQL `backend/migrations/20260728_task095_version_mixin.sql` (additive IF NOT EXISTS)
  - ✅ мутация: schema guard — duplicate `field.Int("version")` / отсутствие Mixin → FAIL
  - ✅ Unclear-12 — exclusion (не version)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

### Checklist ID → evidence

| ID | Evidence |
|----|----------|
| A2W8-1-001 | VersionMixin + 45 mutable schemas |
| A2W8-1-002/006 | docs+PO: unclear-12 exclusion in ADR |
| A2W8-2-* / A2W8-3-007 / A2W8-4-* / A2W8-5-* / A2W1-A4-001 | schema Mixin + migrate SQL |
| W7-A2-003 | intranet `VersionEQ` in `UpdateNewsPost` + test |

### Test commands + PASS

```
go test -tags hubtest -count=1 -timeout=180s \
  ./testframework/ent/ ./testframework/intranet/ ./testframework/sprint/ \
  ./testframework/project/ ./testframework/messenger/ ./testframework/tracker/
```

PASS (2026-07-28): ent, intranet, sprint, project, messenger, tracker.

### Mutation note

- Убрать `VersionMixin{}` / вернуть inline `field.Int("version")` → `TestSchema_TASK095_VersionMixinOnMutable` FAIL
- Убрать `VersionEQ` из `UpdateNewsPost` / `UpdateSprint` → stale version проходит → OL tests FAIL

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record — ✅ выполнено
