# TASK-099 — 📐 Storage NF-09: size==0 legacy bound

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта · MUTATION-2026-07-29-W3  
> Источник: W7-A4-001 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Legacy Confirm/presigned path с size==0 без ContentLength отвергается или ограничен по ADR NF-09.

## [ROLE]

- Role: product→be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/storage/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать валидные multipart без ADR
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: ADR Accepted в репо; owner 2026-07-28 — отдельный approve не требуется

## [REFERENCES]

- Finding IDs: W7-A4-001 NF-09 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-STORAGE-NF09-SIZE-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR/product boundary
  - reject или require ContentLength
  - тест
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

## [EVIDENCE]

| ID | Status | Evidence |
|----|--------|----------|
| W7-A4-001 | FIXED | `GenerateUploadURL` reject `size<=0`; always `ContentLength`; `TestGenerateUploadURL_RejectsZeroSize_W7_A4_001` |

Mutation: убрать `size <= 0` → `ErrInvalidFileSize` → тест красный.
