# TASK-110 — 📞 Calls: recording/room rollback/guest entropy

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-29-W5
> Источник: A2W3 calls · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Stop orphan egress; rollback room on AddParticipant fail; invite code entropy увеличен.

## [EVIDENCE — be2 2026-07-28]

| ID | Status | Evidence |
|----|--------|----------|
| A2W3-5-001 | **FIXED** | `compensateOrphanEgress` / `StopEgress` после fail `CreateRecording`; IT `TestStartRecording_CreateRecordingFail_StopsEgress_A2W3_5_001` |
| A2W3-5-002 | **FIXED** | `finalizeCreate` возвращает err + `DeleteRoom` + `HardDeleteCall`; IT `TestFinalizeCreate_AddParticipantFail_RollsBackRoomAndCall_A2W3_5_002` |
| A2W3-5-003 | **FIXED** | crypto/rand base62×12 (~71 bit) within MaxLen(12); IT `TestCreateInstantCall_InviteCodeBase62Entropy_A2W3_5_003` |

**Mutation:** убрать `compensateOrphanEgress` → StopEgress=0 RED; slog-only AddParticipant → call active + DeleteRoom=0 RED; uuid-hex[:12] → non-hex letter assert RED.

**Команды (PASS):**
```
go test -tags hubtest -race -count=1 -timeout=300s ./testframework/calls/
```

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/calls/**
  - testframework/calls/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - mediasoup migration
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W3-5-001..003 (3)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - три фикса + IT
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
